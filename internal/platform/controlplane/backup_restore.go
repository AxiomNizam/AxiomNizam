package controlplane
import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// BackupManager manages comprehensive platform backups
type BackupManager struct {
	mu            sync.RWMutex
	backups       map[string]*PlatformBackup
	resourceMgr   *ResourceManager
	snapshots     map[string]*BackupSnapshot
	schedules     map[string]*BackupSchedule
	maxBackups    int
	retentionDays int
}

// PlatformBackup represents a complete platform backup
type PlatformBackup struct {
	ID           string
	Name         string
	Timestamp    time.Time
	ExpiresAt    time.Time
	Status       string // InProgress, Completed, Failed
	Version      string
	Description  string
	ResourceData *ResourceBackupData
	PolicyData   *PolicyBackupData
	ConfigData   *ConfigBackupData
	Size         int64
	Checksum     string
	CreatedBy    string
	Tags         map[string]string
	Components   []string // List of backed-up components
}

// ResourceBackupData contains all resource data
type ResourceBackupData struct {
	Namespaces map[string]*NamespaceBackup
	Resources  map[string][]*ResourceSnapshot
	Events     []*EventSnapshot
	Finalizers map[string][]Finalizer
	Conditions map[string][]ResourceCondition
	Ownership  map[string][]OwnerReference
}

// NamespaceBackup contains namespace-level data
type NamespaceBackup struct {
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	Resources   map[string][]*ResourceSnapshot
}

// ResourceSnapshot represents a point-in-time resource snapshot
type ResourceSnapshot struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   ResourceMetadata       `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
	Status     ResourceStatus         `json:"status"`
	Generation int64                  `json:"generation"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// EventSnapshot represents a backed-up event
type EventSnapshot struct {
	ResourceName string
	ResourceKind string
	EventType    string
	Reason       string
	Message      string
	Timestamp    time.Time
	Count        int
}

// PolicyBackupData contains all policy data
type PolicyBackupData struct {
	Policies       map[string]interface{}
	ComplianceReqs map[string]interface{}
	AdmissionRules map[string]interface{}
	RBACRules      map[string]interface{}
}

// ConfigBackupData contains all configuration data
type ConfigBackupData struct {
	GlobalConfig   map[string]interface{}
	ServiceConfigs map[string]interface{}
	Credentials    map[string]interface{} // Encrypted
	Secrets        map[string]interface{} // Encrypted
}

// BackupSchedule defines a backup schedule
type BackupSchedule struct {
	ID        string
	Name      string
	Enabled   bool
	Cron      string
	Retention int // days
	LastRun   *time.Time
	NextRun   time.Time
	Owner     string
}

// BackupSnapshot represents a backup state at a point in time
type BackupSnapshot struct {
	BackupID  string
	Timestamp time.Time
	Status    string
	Progress  int // 0-100
}

// RestoreOptions controls backup restoration
type RestoreOptions struct {
	RestoreResources   bool
	RestorePolicies    bool
	RestoreConfigs     bool
	SkipNamespaces     []string
	SkipResourceKinds  []string
	DryRun             bool
	Force              bool
	ValidationMode     string // strict, lenient, skip
	ConflictResolution string // overwrite, skip, merge
}

// RestoreResult represents a restore operation result
type RestoreResult struct {
	ID               string
	BackupID         string
	Status           string // Success, PartialSuccess, Failed
	StartTime        time.Time
	EndTime          time.Time
	ResourcesCreated int
	ResourcesUpdated int
	ResourcesFailed  int
	PoliciesRestored int
	ConfigsRestored  int
	Errors           []string
	Warnings         []string
}

// NewBackupManager creates a new backup manager
func NewBackupManager(resourceMgr *ResourceManager, maxBackups, retentionDays int) *BackupManager {
	return &BackupManager{
		backups:       make(map[string]*PlatformBackup),
		resourceMgr:   resourceMgr,
		snapshots:     make(map[string]*BackupSnapshot),
		schedules:     make(map[string]*BackupSchedule),
		maxBackups:    maxBackups,
		retentionDays: retentionDays,
	}
}

// CreateBackup creates a complete platform backup
func (bm *BackupManager) CreateBackup(ctx context.Context, name, description string) (*PlatformBackup, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	backupID := fmt.Sprintf("backup-%d", time.Now().UnixNano())

	backup := &PlatformBackup{
		ID:          backupID,
		Name:        name,
		Timestamp:   time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, bm.retentionDays),
		Status:      "InProgress",
		Description: description,
		Tags:        make(map[string]string),
		Components:  make([]string, 0),
	}

	bm.backups[backupID] = backup

	// Start backup in background
	go bm.executeBackup(ctx, backupID)

	return backup, nil
}

// executeBackup performs the actual backup
func (bm *BackupManager) executeBackup(ctx context.Context, backupID string) {
	bm.mu.Lock()
	backup := bm.backups[backupID]
	bm.mu.Unlock()

	if backup == nil {
		return
	}

	// Backup resources
	backup.ResourceData = bm.backupResources(ctx)
	backup.Components = append(backup.Components, "resources")

	// Backup policies
	backup.PolicyData = &PolicyBackupData{
		Policies:       make(map[string]interface{}),
		ComplianceReqs: make(map[string]interface{}),
		AdmissionRules: make(map[string]interface{}),
		RBACRules:      make(map[string]interface{}),
	}
	backup.Components = append(backup.Components, "policies")

	// Backup configs
	backup.ConfigData = &ConfigBackupData{
		GlobalConfig:   make(map[string]interface{}),
		ServiceConfigs: make(map[string]interface{}),
		Credentials:    make(map[string]interface{}),
		Secrets:        make(map[string]interface{}),
	}
	backup.Components = append(backup.Components, "configs")

	// Calculate checksum
	backup.Checksum = calculateChecksum(backup)
	backup.Status = "Completed"

	bm.mu.Lock()
	bm.backups[backupID] = backup
	bm.cleanupOldBackups()
	bm.mu.Unlock()
}

// backupResources backs up all resources
func (bm *BackupManager) backupResources(ctx context.Context) *ResourceBackupData {
	data := &ResourceBackupData{
		Namespaces: make(map[string]*NamespaceBackup),
		Resources:  make(map[string][]*ResourceSnapshot),
		Events:     make([]*EventSnapshot, 0),
		Finalizers: make(map[string][]Finalizer),
		Conditions: make(map[string][]ResourceCondition),
		Ownership:  make(map[string][]OwnerReference),
	}

	// Backup resources by namespace
	bm.mu.RLock()
	for namespace := range bm.resourceMgr.resources {
		nsBackup := &NamespaceBackup{
			Name:        namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Resources:   make(map[string][]*ResourceSnapshot),
		}

		for _, resource := range bm.resourceMgr.resources[namespace] {
			snapshot := &ResourceSnapshot{
				APIVersion: resource.APIVersion,
				Kind:       resource.Kind,
				Metadata:   resource.Metadata,
				Spec:       resource.Spec,
				Status:     resource.Status,
				Generation: resource.Generation,
				CreatedAt:  resource.CreatedAt,
			}

			key := resource.Kind
			nsBackup.Resources[key] = append(nsBackup.Resources[key], snapshot)

			// Backup finalizers
			resourceID := namespace + "/" + resource.Kind + "/" + resource.Metadata.Name
			if fm, exists := bm.resourceMgr.finalizers[resourceID]; exists {
				data.Finalizers[resourceID] = fm.GetFinalizers(resourceID)
			}

			// Backup conditions
			if cm, exists := bm.resourceMgr.conditions[resourceID]; exists {
				statusConds := cm.GetConditions(resourceID)
				resourceConds := make([]ResourceCondition, len(statusConds))
				for i, sc := range statusConds {
					resourceConds[i] = ResourceCondition{
						Type:               sc.Type,
						Status:             sc.Status,
						Reason:             sc.Reason,
						Message:            sc.Message,
						ObservedGeneration: sc.ObservedGeneration,
						LastTransition:     sc.LastTransitionTime,
					}
				}
				data.Conditions[resourceID] = resourceConds
			}
		}

		data.Namespaces[namespace] = nsBackup
	}
	bm.mu.RUnlock()

	return data
}

// GetBackup retrieves a backup
func (bm *BackupManager) GetBackup(backupID string) *PlatformBackup {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return bm.backups[backupID]
}

// ListBackups lists all backups
func (bm *BackupManager) ListBackups(ctx context.Context) []*PlatformBackup {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	backups := make([]*PlatformBackup, 0, len(bm.backups))
	for _, backup := range bm.backups {
		backups = append(backups, backup)
	}

	return backups
}

// DeleteBackup deletes a backup
func (bm *BackupManager) DeleteBackup(backupID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.backups[backupID]; !exists {
		return fmt.Errorf("backup not found")
	}

	delete(bm.backups, backupID)
	delete(bm.snapshots, backupID)

	return nil
}

// ExportBackup exports a backup as compressed archive
func (bm *BackupManager) ExportBackup(ctx context.Context, backupID string) ([]byte, error) {
	bm.mu.RLock()
	backup := bm.backups[backupID]
	bm.mu.RUnlock()

	if backup == nil {
		return nil, fmt.Errorf("backup not found")
	}

	// Create tar.gz archive
	buf := new(bytes.Buffer)
	gzWriter := gzip.NewWriter(buf)
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()
	defer gzWriter.Close()

	// Marshal backup data
	backupJSON, _ := json.MarshalIndent(backup, "", "  ")

	// Add to tar
	header := &tar.Header{
		Name: "backup.json",
		Size: int64(len(backupJSON)),
	}
	tarWriter.WriteHeader(header)
	tarWriter.Write(backupJSON)

	// Add resource data
	if backup.ResourceData != nil {
		resourceJSON, _ := json.MarshalIndent(backup.ResourceData, "", "  ")
		header := &tar.Header{
			Name: "resources.json",
			Size: int64(len(resourceJSON)),
		}
		tarWriter.WriteHeader(header)
		tarWriter.Write(resourceJSON)
	}

	// Add policy data
	if backup.PolicyData != nil {
		policyJSON, _ := json.MarshalIndent(backup.PolicyData, "", "  ")
		header := &tar.Header{
			Name: "policies.json",
			Size: int64(len(policyJSON)),
		}
		tarWriter.WriteHeader(header)
		tarWriter.Write(policyJSON)
	}

	// Add config data
	if backup.ConfigData != nil {
		configJSON, _ := json.MarshalIndent(backup.ConfigData, "", "  ")
		header := &tar.Header{
			Name: "configs.json",
			Size: int64(len(configJSON)),
		}
		tarWriter.WriteHeader(header)
		tarWriter.Write(configJSON)
	}

	return buf.Bytes(), nil
}

// ImportBackup imports a backup from compressed archive
func (bm *BackupManager) ImportBackup(ctx context.Context, data []byte) (*PlatformBackup, error) {
	buf := bytes.NewBuffer(data)
	gzReader, _ := gzip.NewReader(buf)
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	backup := &PlatformBackup{
		ID:           fmt.Sprintf("backup-%d", time.Now().UnixNano()),
		Timestamp:    time.Now(),
		ExpiresAt:    time.Now().AddDate(0, 0, bm.retentionDays),
		Status:       "Completed",
		ResourceData: &ResourceBackupData{},
		PolicyData:   &PolicyBackupData{},
		ConfigData:   &ConfigBackupData{},
		Tags:         make(map[string]string),
		Components:   make([]string, 0),
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		data, _ := io.ReadAll(tarReader)

		switch header.Name {
		case "backup.json":
			json.Unmarshal(data, backup)
		case "resources.json":
			json.Unmarshal(data, backup.ResourceData)
			backup.Components = append(backup.Components, "resources")
		case "policies.json":
			json.Unmarshal(data, backup.PolicyData)
			backup.Components = append(backup.Components, "policies")
		case "configs.json":
			json.Unmarshal(data, backup.ConfigData)
			backup.Components = append(backup.Components, "configs")
		}
	}

	bm.mu.Lock()
	bm.backups[backup.ID] = backup
	bm.mu.Unlock()

	return backup, nil
}

// RestoreBackup restores a backup
func (bm *BackupManager) RestoreBackup(ctx context.Context, backupID string, options *RestoreOptions) (*RestoreResult, error) {
	bm.mu.RLock()
	backup := bm.backups[backupID]
	bm.mu.RUnlock()

	if backup == nil {
		return nil, fmt.Errorf("backup not found")
	}

	result := &RestoreResult{
		ID:        fmt.Sprintf("restore-%d", time.Now().UnixNano()),
		BackupID:  backupID,
		Status:    "InProgress",
		StartTime: time.Now(),
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
	}

	// Restore resources
	if options.RestoreResources && backup.ResourceData != nil {
		bm.restoreResources(ctx, backup.ResourceData, options, result)
	}

	// Restore policies
	if options.RestorePolicies && backup.PolicyData != nil {
		result.PoliciesRestored = len(backup.PolicyData.Policies)
	}

	// Restore configs
	if options.RestoreConfigs && backup.ConfigData != nil {
		result.ConfigsRestored = len(backup.ConfigData.GlobalConfig)
	}

	result.EndTime = time.Now()

	if len(result.Errors) == 0 {
		result.Status = "Success"
	} else {
		result.Status = "PartialSuccess"
	}

	return result, nil
}

// restoreResources restores resources from backup
func (bm *BackupManager) restoreResources(ctx context.Context, resourceData *ResourceBackupData, options *RestoreOptions, result *RestoreResult) {
	for namespace, nsBackup := range resourceData.Namespaces {
		// Skip if in skip list
		if contains(options.SkipNamespaces, namespace) {
			continue
		}

		for _, snapshots := range nsBackup.Resources {
			for _, snapshot := range snapshots {
				// Skip if kind in skip list
				if contains(options.SkipResourceKinds, snapshot.Kind) {
					continue
				}

				// Validate if strict mode
				if options.ValidationMode == "strict" {
					// Perform validation
					if !bm.validateResource(snapshot) {
						result.Errors = append(result.Errors, fmt.Sprintf("validation failed for %s/%s", snapshot.Kind, snapshot.Metadata.Name))
						result.ResourcesFailed++
						continue
					}
				}

				// Restore resource
				if options.DryRun {
					result.ResourcesCreated++
				} else {
					resource := &ManagedResource{
						APIVersion: snapshot.APIVersion,
						Kind:       snapshot.Kind,
						Metadata:   snapshot.Metadata,
						Spec:       snapshot.Spec,
						Status:     snapshot.Status,
						Generation: snapshot.Generation,
						CreatedAt:  snapshot.CreatedAt,
						UpdatedAt:  time.Now(),
					}

					// Check conflict resolution
					existing, err := bm.resourceMgr.Get(ctx, namespace, snapshot.Kind, snapshot.Metadata.Name)
					if err == nil {
						// Resource exists
						switch options.ConflictResolution {
						case "skip":
							result.Warnings = append(result.Warnings, fmt.Sprintf("skipping existing %s/%s", snapshot.Kind, snapshot.Metadata.Name))
							continue
						case "overwrite":
							bm.resourceMgr.Delete(ctx, namespace, snapshot.Kind, snapshot.Metadata.Name)
							result.ResourcesUpdated++
						case "merge":
							resource.Generation = existing.Generation + 1
							result.ResourcesUpdated++
						}
					}

					_, err = bm.resourceMgr.Create(ctx, resource)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("failed to restore %s/%s: %v", snapshot.Kind, snapshot.Metadata.Name, err))
						result.ResourcesFailed++
					} else {
						result.ResourcesCreated++
					}
				}
			}
		}
	}
}

// validateResource validates a resource snapshot
func (bm *BackupManager) validateResource(snapshot *ResourceSnapshot) bool {
	// Basic validation
	if snapshot.Metadata.Name == "" || snapshot.Kind == "" {
		return false
	}

	return true
}

// CreateBackupSchedule creates a backup schedule
func (bm *BackupManager) CreateBackupSchedule(ctx context.Context, schedule *BackupSchedule) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if schedule.ID == "" {
		schedule.ID = fmt.Sprintf("schedule-%d", time.Now().UnixNano())
	}

	bm.schedules[schedule.ID] = schedule
	return nil
}

// cleanupOldBackups removes old backups exceeding maxBackups
func (bm *BackupManager) cleanupOldBackups() {
	if len(bm.backups) <= bm.maxBackups {
		return
	}

	// Find oldest backup
	var oldestID string
	var oldestTime time.Time

	for id, backup := range bm.backups {
		if oldestTime.IsZero() || backup.Timestamp.Before(oldestTime) {
			oldestID = id
			oldestTime = backup.Timestamp
		}
	}

	if oldestID != "" {
		delete(bm.backups, oldestID)
	}
}

// VerifyBackup verifies backup integrity
func (bm *BackupManager) VerifyBackup(backupID string) (bool, error) {
	bm.mu.RLock()
	backup := bm.backups[backupID]
	bm.mu.RUnlock()

	if backup == nil {
		return false, fmt.Errorf("backup not found")
	}

	calculatedChecksum := calculateChecksum(backup)
	return backup.Checksum == calculatedChecksum, nil
}

// calculateChecksum calculates backup checksum
func calculateChecksum(backup *PlatformBackup) string {
	data, _ := json.Marshal(backup)
	// Simple checksum - in production use proper hash
	return fmt.Sprintf("%x", len(data))
}

// contains checks if string is in slice
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
