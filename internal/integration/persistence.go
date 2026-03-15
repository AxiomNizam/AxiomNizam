package integration

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	persistenceMu         sync.RWMutex
	globalPersistenceEtcd *clientv3.Client
)

func integrationEtcdClient() *clientv3.Client {
	persistenceMu.RLock()
	defer persistenceMu.RUnlock()
	return globalPersistenceEtcd
}

func ConfigureGlobalPersistence(etcd *clientv3.Client) {
	persistenceMu.Lock()
	globalPersistenceEtcd = etcd
	persistenceMu.Unlock()

	if GlobalComplianceAuditor != nil {
		GlobalComplianceAuditor.ConfigurePersistence(etcd)
	}
	if GlobalHealthMonitor != nil {
		GlobalHealthMonitor.ConfigurePersistence(etcd)
	}
	if GlobalAlertManager != nil {
		GlobalAlertManager.ConfigurePersistence(etcd)
	}
}

func loadStateFromEtcd(etcd *clientv3.Client, stateKey string, out interface{}) bool {
	if etcd == nil || stateKey == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := etcd.Get(ctx, stateKey)
	if err != nil {
		log.Printf("integration: failed to load persisted state from etcd (%s): %v", stateKey, err)
		return false
	}
	if len(resp.Kvs) == 0 {
		return false
	}

	if err := json.Unmarshal(resp.Kvs[0].Value, out); err != nil {
		log.Printf("integration: failed to decode persisted state (%s): %v", stateKey, err)
		return false
	}
	return true
}

func saveStateToEtcd(etcd *clientv3.Client, stateKey string, payload interface{}) {
	if etcd == nil || stateKey == "" {
		return
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		log.Printf("integration: failed to encode state (%s): %v", stateKey, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := etcd.Put(ctx, stateKey, string(encoded)); err != nil {
		log.Printf("integration: failed to persist state to etcd (%s): %v", stateKey, err)
	}
}
