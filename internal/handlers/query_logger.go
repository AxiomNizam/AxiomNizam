package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// QueryLog represents a logged query with enterprise-level information
type QueryLog struct {
	ID           string    `json:"id"`
	Query        string    `json:"query"`
	Params       []string  `json:"params"`
	Database     string    `json:"database"`
	User         string    `json:"user,omitempty"`
	Role         string    `json:"role,omitempty"`
	Status       string    `json:"status"` // "success" or "error"
	Error        string    `json:"error,omitempty"`
	Duration     int64     `json:"duration_ms"`
	RowsReturned int64     `json:"rows_returned,omitempty"`
	RowsAffected int64     `json:"rows_affected,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address,omitempty"`
	QueryType    string    `json:"query_type,omitempty"` // SELECT, INSERT, UPDATE, DELETE, etc.
}

// QueryMetrics tracks aggregated query metrics
type QueryMetrics struct {
	TotalQueries  int64
	SuccessCount  int64
	ErrorCount    int64
	TotalDuration int64
	AvgDuration   float64
	TotalRows     int64
	ByDatabase    map[string]*DatabaseMetrics
	ByUser        map[string]*UserMetrics
	ByQueryType   map[string]*QueryTypeMetrics
}

// DatabaseMetrics tracks metrics per database
type DatabaseMetrics struct {
	Database     string
	TotalQueries int64
	SuccessCount int64
	ErrorCount   int64
	AvgDuration  float64
	TotalRows    int64
}

// UserMetrics tracks metrics per user
type UserMetrics struct {
	User          string
	Role          string
	TotalQueries  int64
	SuccessCount  int64
	ErrorCount    int64
	AvgDuration   float64
	LastQueryTime time.Time
}

// QueryTypeMetrics tracks metrics by query type
type QueryTypeMetrics struct {
	QueryType    string
	TotalQueries int64
	SuccessCount int64
	ErrorCount   int64
	AvgDuration  float64
	TotalRows    int64
}

// QueryLogger handles persistent logging of queries with enterprise features
type QueryLogger struct {
	redisClient *redis.Client
	storageDir  string
	logFile     string
	metrics     *QueryMetrics
}

// NewQueryLogger creates a new query logger with metrics tracking
func NewQueryLogger(redisClient *redis.Client, storageDir string) *QueryLogger {
	if storageDir == "" {
		storageDir = "/data/logs"
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		fmt.Printf("⚠️  Warning: Could not create storage directory: %v\n", err)
	}

	logFile := filepath.Join(storageDir, "query_logs.jsonl")

	return &QueryLogger{
		redisClient: redisClient,
		storageDir:  storageDir,
		logFile:     logFile,
		metrics: &QueryMetrics{
			ByDatabase:  make(map[string]*DatabaseMetrics),
			ByUser:      make(map[string]*UserMetrics),
			ByQueryType: make(map[string]*QueryTypeMetrics),
		},
	}
}

// LogQuery logs a query to both disk and Redis with metrics tracking
func (ql *QueryLogger) LogQuery(log QueryLog) error {
	// Get hostname
	hostname, _ := os.Hostname()
	log.Hostname = hostname

	// Set ID if not set
	if log.ID == "" {
		log.ID = fmt.Sprintf("%s-%d-%d", hostname, time.Now().Unix(), time.Now().Nanosecond())
	}

	// Set timestamp if not set
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// Log to disk (append to file)
	if err := ql.logToDisk(log); err != nil {
		fmt.Printf("⚠️  Warning: Failed to log to disk: %v\n", err)
	}

	// Log to Redis (for distributed access and quick retrieval)
	if ql.redisClient != nil {
		if err := ql.logToRedis(log); err != nil {
			fmt.Printf("⚠️  Warning: Failed to log to Redis: %v\n", err)
		}
	}

	// Update metrics
	ql.UpdateMetrics(log)

	return nil
}

// logToDisk writes query log to disk file
func (ql *QueryLogger) logToDisk(log QueryLog) error {
	file, err := os.OpenFile(ql.logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	// Write as JSONL (JSON Lines format - one JSON per line)
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}

// logToRedis stores query log in Redis for distributed access
func (ql *QueryLogger) logToRedis(log QueryLog) error {
	if ql.redisClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	// Store in Redis with key pattern: query_logs:{database}:{timestamp}:{id}
	key := fmt.Sprintf("query_logs:%s:%d:%s", log.Database, log.Timestamp.Unix(), log.ID)

	// Set with expiration (keep for 30 days)
	if err := ql.redisClient.Set(ctx, key, string(data), 30*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to write to Redis: %w", err)
	}

	// Also add to a sorted set for easy retrieval by time
	// Key format: query_logs_by_time:{database}
	sortedSetKey := fmt.Sprintf("query_logs_by_time:%s", log.Database)
	if err := ql.redisClient.ZAdd(ctx, sortedSetKey, redis.Z{
		Score:  float64(log.Timestamp.Unix()),
		Member: log.ID,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add to sorted set: %w", err)
	}

	// Keep the sorted set to 100k entries max (trim old ones)
	if err := ql.redisClient.ZRemRangeByRank(ctx, sortedSetKey, 0, -100001).Err(); err != nil {
		fmt.Printf("⚠️  Warning: Failed to trim sorted set: %v\n", err)
	}

	return nil
}

// GetQueryLogs retrieves logs from Redis or disk
func (ql *QueryLogger) GetQueryLogs(database string, limit int) ([]QueryLog, error) {
	var logs []QueryLog

	if ql.redisClient == nil {
		// Fall back to reading from disk
		return ql.getLogsFromDisk(limit)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get latest logs from Redis sorted set
	sortedSetKey := fmt.Sprintf("query_logs_by_time:%s", database)
	ids, err := ql.redisClient.ZRevRange(ctx, sortedSetKey, 0, int64(limit-1)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Get the full logs from Redis
	for _, id := range ids {
		// We need to use pattern matching, but we'll retrieve the most recent one
		pattern := fmt.Sprintf("query_logs:%s:*:%s", database, id)
		keys, err := ql.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		if len(keys) > 0 {
			val, err := ql.redisClient.Get(ctx, keys[len(keys)-1]).Result()
			if err == nil {
				var log QueryLog
				if err := json.Unmarshal([]byte(val), &log); err == nil {
					logs = append(logs, log)
				}
			}
		}
	}

	return logs, nil
}

// getLogsFromDisk reads logs from disk file
func (ql *QueryLogger) getLogsFromDisk(limit int) ([]QueryLog, error) {
	var logs []QueryLog

	_, err := os.ReadFile(ql.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return logs, nil // File doesn't exist yet
		}
		return nil, err
	}

	// Parse JSONL format (one JSON per line, in reverse order for latest first)
	// TODO: Implement parsing of JSONL format from disk

	return logs, nil
}

// GetQueryStats returns statistics about queries
func (ql *QueryLogger) GetQueryStats(database string) (map[string]interface{}, error) {
	if ql.redisClient == nil {
		return map[string]interface{}{"error": "Redis not available"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats := map[string]interface{}{}

	// Get total count
	sortedSetKey := fmt.Sprintf("query_logs_by_time:%s", database)
	count, err := ql.redisClient.ZCard(ctx, sortedSetKey).Result()
	if err == nil {
		stats["total_queries"] = count
	}

	// Get latest queries
	latestIDs, err := ql.redisClient.ZRevRange(ctx, sortedSetKey, 0, 9).Result()
	if err == nil && len(latestIDs) > 0 {
		stats["latest_query_count"] = len(latestIDs)
	}

	return stats, nil
}

// UpdateMetrics updates aggregated metrics based on query log
func (ql *QueryLogger) UpdateMetrics(log QueryLog) {
	if ql.metrics == nil {
		return
	}

	// Update total metrics
	ql.metrics.TotalQueries++
	ql.metrics.TotalDuration += log.Duration
	ql.metrics.AvgDuration = float64(ql.metrics.TotalDuration) / float64(ql.metrics.TotalQueries)

	if log.Status == "success" {
		ql.metrics.SuccessCount++
		ql.metrics.TotalRows += log.RowsReturned + log.RowsAffected
	} else {
		ql.metrics.ErrorCount++
	}

	// Update database metrics
	if log.Database != "" {
		dbMetrics, exists := ql.metrics.ByDatabase[log.Database]
		if !exists {
			dbMetrics = &DatabaseMetrics{Database: log.Database}
			ql.metrics.ByDatabase[log.Database] = dbMetrics
		}

		dbMetrics.TotalQueries++
		dbMetrics.TotalRows += log.RowsReturned + log.RowsAffected
		dbMetrics.AvgDuration = (dbMetrics.AvgDuration*float64(dbMetrics.TotalQueries-1) + float64(log.Duration)) / float64(dbMetrics.TotalQueries)

		if log.Status == "success" {
			dbMetrics.SuccessCount++
		} else {
			dbMetrics.ErrorCount++
		}
	}

	// Update user metrics
	if log.User != "" {
		userMetrics, exists := ql.metrics.ByUser[log.User]
		if !exists {
			userMetrics = &UserMetrics{
				User:          log.User,
				Role:          log.Role,
				LastQueryTime: log.Timestamp,
			}
			ql.metrics.ByUser[log.User] = userMetrics
		}

		userMetrics.TotalQueries++
		userMetrics.Role = log.Role // Update role in case it changed
		userMetrics.LastQueryTime = log.Timestamp
		userMetrics.AvgDuration = (userMetrics.AvgDuration*float64(userMetrics.TotalQueries-1) + float64(log.Duration)) / float64(userMetrics.TotalQueries)

		if log.Status == "success" {
			userMetrics.SuccessCount++
		} else {
			userMetrics.ErrorCount++
		}
	}

	// Update query type metrics
	queryType := log.QueryType
	if queryType == "" {
		queryType = "UNKNOWN"
	}

	typeMetrics, exists := ql.metrics.ByQueryType[queryType]
	if !exists {
		typeMetrics = &QueryTypeMetrics{QueryType: queryType}
		ql.metrics.ByQueryType[queryType] = typeMetrics
	}

	typeMetrics.TotalQueries++
	typeMetrics.TotalRows += log.RowsReturned + log.RowsAffected
	typeMetrics.AvgDuration = (typeMetrics.AvgDuration*float64(typeMetrics.TotalQueries-1) + float64(log.Duration)) / float64(typeMetrics.TotalQueries)

	if log.Status == "success" {
		typeMetrics.SuccessCount++
	} else {
		typeMetrics.ErrorCount++
	}
}

// GetMetrics returns current aggregated metrics
func (ql *QueryLogger) GetMetrics() *QueryMetrics {
	if ql.metrics == nil {
		return &QueryMetrics{}
	}
	return ql.metrics
}

// GetUserMetrics returns metrics for specific user
func (ql *QueryLogger) GetUserMetrics(user string) *UserMetrics {
	if ql.metrics == nil || ql.metrics.ByUser == nil {
		return nil
	}
	return ql.metrics.ByUser[user]
}

// GetDatabaseMetrics returns metrics for specific database
func (ql *QueryLogger) GetDatabaseMetrics(database string) *DatabaseMetrics {
	if ql.metrics == nil || ql.metrics.ByDatabase == nil {
		return nil
	}
	return ql.metrics.ByDatabase[database]
}

// GetQueryTypeMetrics returns metrics for specific query type
func (ql *QueryLogger) GetQueryTypeMetrics(queryType string) *QueryTypeMetrics {
	if ql.metrics == nil || ql.metrics.ByQueryType == nil {
		return nil
	}
	return ql.metrics.ByQueryType[queryType]
}

// GetSlowQueries returns slow queries (queries exceeding threshold)
func (ql *QueryLogger) GetSlowQueries(database string, thresholdMs int64, limit int) ([]QueryLog, error) {
	if ql.redisClient == nil {
		return ql.getSlowQueriesFromDisk(thresholdMs, limit)
	}

	var logs []QueryLog

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sortedSetKey := fmt.Sprintf("query_logs_by_time:%s", database)
	ids, err := ql.redisClient.ZRevRange(ctx, sortedSetKey, 0, int64(limit*10)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	for _, id := range ids {
		if len(logs) >= limit {
			break
		}

		pattern := fmt.Sprintf("query_logs:%s:*:%s", database, id)
		keys, err := ql.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		if len(keys) > 0 {
			val, err := ql.redisClient.Get(ctx, keys[len(keys)-1]).Result()
			if err == nil {
				var log QueryLog
				if err := json.Unmarshal([]byte(val), &log); err == nil {
					if log.Duration >= thresholdMs {
						logs = append(logs, log)
					}
				}
			}
		}
	}

	return logs, nil
}

// GetErroredQueries returns queries that resulted in errors
func (ql *QueryLogger) GetErroredQueries(database string, limit int) ([]QueryLog, error) {
	if ql.redisClient == nil {
		return nil, fmt.Errorf("Redis not available")
	}

	var logs []QueryLog

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sortedSetKey := fmt.Sprintf("query_logs_by_time:%s", database)
	ids, err := ql.redisClient.ZRevRange(ctx, sortedSetKey, 0, int64(limit*10)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	for _, id := range ids {
		if len(logs) >= limit {
			break
		}

		pattern := fmt.Sprintf("query_logs:%s:*:%s", database, id)
		keys, err := ql.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		if len(keys) > 0 {
			val, err := ql.redisClient.Get(ctx, keys[len(keys)-1]).Result()
			if err == nil {
				var log QueryLog
				if err := json.Unmarshal([]byte(val), &log); err == nil {
					if log.Status == "error" && log.Error != "" {
						logs = append(logs, log)
					}
				}
			}
		}
	}

	return logs, nil
}

// GetQuerysByUser returns all queries for a specific user
func (ql *QueryLogger) GetQuerysByUser(user string, limit int) ([]QueryLog, error) {
	if ql.redisClient == nil {
		return nil, fmt.Errorf("Redis not available")
	}

	var logs []QueryLog

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Scan all query logs for matching user
	pattern := fmt.Sprintf("query_logs:*:*:*")
	keys, err := ql.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if len(logs) >= limit {
			break
		}

		val, err := ql.redisClient.Get(ctx, key).Result()
		if err == nil {
			var log QueryLog
			if err := json.Unmarshal([]byte(val), &log); err == nil {
				if log.User == user {
					logs = append(logs, log)
				}
			}
		}
	}

	return logs, nil
}

// GetSlowQueriesFromDisk reads slow queries from disk file
func (ql *QueryLogger) getSlowQueriesFromDisk(thresholdMs int64, limit int) ([]QueryLog, error) {
	var logs []QueryLog
	// TODO: Implement parsing of JSONL format from disk for slow queries
	return logs, nil
}

// GetMetricsReport returns comprehensive metrics report
func (ql *QueryLogger) GetMetricsReport() map[string]interface{} {
	if ql.metrics == nil {
		return map[string]interface{}{}
	}

	// Top users by query count
	var topUsers []map[string]interface{}
	for _, user := range ql.metrics.ByUser {
		topUsers = append(topUsers, map[string]interface{}{
			"user":          user.User,
			"role":          user.Role,
			"total_queries": user.TotalQueries,
			"success_count": user.SuccessCount,
			"error_count":   user.ErrorCount,
			"avg_duration":  user.AvgDuration,
			"last_query":    user.LastQueryTime,
		})
	}

	// Database statistics
	var dbStats []map[string]interface{}
	for _, db := range ql.metrics.ByDatabase {
		dbStats = append(dbStats, map[string]interface{}{
			"database":      db.Database,
			"total_queries": db.TotalQueries,
			"success_count": db.SuccessCount,
			"error_count":   db.ErrorCount,
			"avg_duration":  db.AvgDuration,
			"total_rows":    db.TotalRows,
		})
	}

	// Query type statistics
	var typeStats []map[string]interface{}
	for _, qt := range ql.metrics.ByQueryType {
		typeStats = append(typeStats, map[string]interface{}{
			"query_type":    qt.QueryType,
			"total_queries": qt.TotalQueries,
			"success_count": qt.SuccessCount,
			"error_count":   qt.ErrorCount,
			"avg_duration":  qt.AvgDuration,
			"total_rows":    qt.TotalRows,
		})
	}

	return map[string]interface{}{
		"total_queries":  ql.metrics.TotalQueries,
		"success_count":  ql.metrics.SuccessCount,
		"error_count":    ql.metrics.ErrorCount,
		"total_duration": ql.metrics.TotalDuration,
		"avg_duration":   ql.metrics.AvgDuration,
		"total_rows":     ql.metrics.TotalRows,
		"success_rate":   float64(ql.metrics.SuccessCount) / float64(ql.metrics.TotalQueries) * 100,
		"by_user":        topUsers,
		"by_database":    dbStats,
		"by_query_type":  typeStats,
	}
}

// DeleteOldLogs deletes query logs older than specified days
// Parameters:
//   - database: Database name or "all" for all databases
//   - days: Delete logs older than N days
//
// Returns: error if operation fails
func (ql *QueryLogger) DeleteOldLogs(database string, days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	logFile := ql.logFile

	if logFile == "" {
		return fmt.Errorf("log file path not configured")
	}

	// Read existing logs from file
	logs := []QueryLog{}
	if _, err := os.Stat(logFile); err == nil {
		content, err := os.ReadFile(logFile)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}

		// Parse JSONL format
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		for _, line := range lines {
			if line != "" {
				var log QueryLog
				if err := json.Unmarshal([]byte(line), &log); err == nil {
					logs = append(logs, log)
				}
			}
		}
	}

	// Filter logs to keep only those newer than cutoff
	keptLogs := []QueryLog{}
	deletedCount := 0

	for _, log := range logs {
		// Skip logs that match database filter and are older than cutoff
		if (database == "all" || log.Database == database) && log.Timestamp.Before(cutoffTime) {
			deletedCount++
			continue
		}
		keptLogs = append(keptLogs, log)
	}

	// Write filtered logs back to file
	if err := os.Remove(logFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old log file: %w", err)
	}

	// Create new log file with kept logs
	for _, log := range keptLogs {
		data, err := json.Marshal(log)
		if err != nil {
			continue
		}

		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		_, _ = f.WriteString(string(data) + "\n")
		_ = f.Close()
	}

	// Also clear Redis entries for deleted logs
	ctx := context.Background()
	pattern := "query_logs:*"
	if database != "all" {
		pattern = fmt.Sprintf("query_logs:%s:*", database)
	}

	iter := ql.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		logStr, err := ql.redisClient.Get(ctx, key).Result()
		if err == nil {
			var log QueryLog
			if err := json.Unmarshal([]byte(logStr), &log); err == nil {
				if log.Timestamp.Before(cutoffTime) {
					_ = ql.redisClient.Del(ctx, key)
				}
			}
		}
	}

	if iter.Err() != nil {
		return fmt.Errorf("failed to scan Redis keys: %w", iter.Err())
	}

	return nil
}
