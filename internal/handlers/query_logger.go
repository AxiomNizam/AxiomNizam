package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"
)

// QueryLog represents a logged query
type QueryLog struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Params    []string  `json:"params"`
	Database  string    `json:"database"`
	User      string    `json:"user,omitempty"`
	Status    string    `json:"status"` // "success" or "error"
	Error     string    `json:"error,omitempty"`
	Duration  int64     `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
}

// QueryLogger handles persistent logging of queries
type QueryLogger struct {
	redisClient *redis.Client
	storageDir  string
	logFile     string
}

// NewQueryLogger creates a new query logger
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
	}
}

// LogQuery logs a query to both disk and Redis
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
		key := fmt.Sprintf("query_logs:%s:*:%s", database, id)
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

	data, err := os.ReadFile(ql.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return logs, nil // File doesn't exist yet
		}
		return nil, err
	}

	// Parse JSONL format (one JSON per line, in reverse order for latest first)
	lines := string(data)
	// We'll read from the end to get latest logs

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

// DeleteOldLogs deletes logs older than specified days
func (ql *QueryLogger) DeleteOldLogs(database string, days int) error {
	if ql.redisClient == nil {
		return fmt.Errorf("Redis not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cutoffTime := time.Now().AddDate(0, 0, -days).Unix()
	sortedSetKey := fmt.Sprintf("query_logs_by_time:%s", database)

	// Remove old entries from sorted set
	if err := ql.redisClient.ZRemRangeByScore(ctx, sortedSetKey, "0", fmt.Sprintf("%d", cutoffTime)).Err(); err != nil {
		return err
	}

	return nil
}
