package distributedstate

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type EtcdStateStore struct {
	client       *clientv3.Client
	mu           sync.RWMutex
	watchers     map[string]context.CancelFunc
	leases       map[int64]*clientv3.LeaseGrantResponse
	sessionPool  map[string]*concurrency.Session
}

func NewEtcdStateStore(endpoints []string) (*EtcdStateStore, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	_, err = client.Get(ctx, "health")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	return &EtcdStateStore{
		client:      client,
		watchers:    make(map[string]context.CancelFunc),
		leases:      make(map[int64]*clientv3.LeaseGrantResponse),
		sessionPool: make(map[string]*concurrency.Session),
	}, nil
}

func (es *EtcdStateStore) Put(ctx context.Context, key string, value string) error {
	_, err := es.client.Put(ctx, key, value)
	return err
}

func (es *EtcdStateStore) Get(ctx context.Context, key string) (string, error) {
	resp, err := es.client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	
	return string(resp.Kvs[0].Value), nil
}

func (es *EtcdStateStore) Delete(ctx context.Context, key string) error {
	_, err := es.client.Delete(ctx, key)
	return err
}

func (es *EtcdStateStore) CompareAndSwap(ctx context.Context, key string, oldValue, newValue string) (bool, error) {
	resp, err := es.client.Get(ctx, key)
	if err != nil {
		return false, err
	}

	if len(resp.Kvs) == 0 {
		if oldValue == "" {
			_, err := es.client.Put(ctx, key, newValue)
			return err == nil, err
		}
		return false, nil
	}

	currentValue := string(resp.Kvs[0].Value)
	currentVersion := resp.Kvs[0].Version

	if currentValue != oldValue {
		return false, nil
	}

	txn := es.client.Txn(ctx).
		If(clientv3.Compare(clientv3.Version(key), "=", currentVersion)).
		Then(clientv3.OpPut(key, newValue)).
		Else(clientv3.OpGet(key))

	txnResp, err := txn.Commit()
	if err != nil {
		return false, err
	}

	return txnResp.Succeeded, nil
}

func (es *EtcdStateStore) GetWithRevision(ctx context.Context, key string) (string, int64, error) {
	resp, err := es.client.Get(ctx, key)
	if err != nil {
		return "", 0, err
	}

	if len(resp.Kvs) == 0 {
		return "", 0, nil
	}

	return string(resp.Kvs[0].Value), resp.Kvs[0].Version, nil
}

func (es *EtcdStateStore) PutWithLease(ctx context.Context, key string, value string, ttl time.Duration) (int64, error) {
	lease, err := es.client.Grant(ctx, int64(ttl.Seconds()))
	if err != nil {
		return 0, err
	}

	_, err = es.client.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		return 0, err
	}

	es.mu.Lock()
	es.leases[lease.ID] = lease
	es.mu.Unlock()

	return lease.ID, nil
}

func (es *EtcdStateStore) ReleaseLeaseID(ctx context.Context, leaseID int64) error {
	_, err := es.client.Revoke(ctx, clientv3.LeaseID(leaseID))
	if err == nil {
		es.mu.Lock()
		delete(es.leases, leaseID)
		es.mu.Unlock()
	}
	return err
}

func (es *EtcdStateStore) Watch(ctx context.Context, key string, callback func(Event)) error {
	watchCtx, cancel := context.WithCancel(ctx)
	
	es.mu.Lock()
	es.watchers[key] = cancel
	es.mu.Unlock()

	watchCh := es.client.Watch(watchCtx, key)

	go func() {
		for wresp := range watchCh {
			for _, event := range wresp.Events {
				eventType := EventTypePut
				if event.Type.String() == "DELETE" {
					eventType = EventTypeDelete
				}

				callback(Event{
					Type:      eventType,
					Key:       string(event.Kv.Key),
					Value:     string(event.Kv.Value),
					Revision:  event.Kv.Version,
					Timestamp: time.Now(),
				})
			}
		}
	}()

	return nil
}

func (es *EtcdStateStore) List(ctx context.Context, prefix string) (map[string]string, error) {
	resp, err := es.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}

	return result, nil
}

func (es *EtcdStateStore) Transaction(ctx context.Context, ops []Operation) (bool, error) {
	conds := []clientv3.Cmp{}
	thenOps := []clientv3.Op{}
	elseOps := []clientv3.Op{}

	for _, op := range ops {
		switch op.Type {
		case OperationCompare:
			conds = append(conds, clientv3.Compare(clientv3.Value(op.Key), "=", op.OldValue))
		case OperationPut:
			thenOps = append(thenOps, clientv3.OpPut(op.Key, op.Value))
		case OperationDelete:
			elseOps = append(elseOps, clientv3.OpDelete(op.Key))
		}
	}

	txn := es.client.Txn(ctx).If(conds...).Then(thenOps...).Else(elseOps...)
	resp, err := txn.Commit()
	if err != nil {
		return false, err
	}

	return resp.Succeeded, nil
}

func (es *EtcdStateStore) GetMutex(ctx context.Context, lockName string) (*concurrency.Mutex, error) {
	sessionKey := fmt.Sprintf("session-%s", lockName)
	
	es.mu.Lock()
	session, exists := es.sessionPool[sessionKey]
	es.mu.Unlock()

	if !exists {
		var err error
		session, err = concurrency.NewSession(es.client)
		if err != nil {
			return nil, err
		}
		
		es.mu.Lock()
		es.sessionPool[sessionKey] = session
		es.mu.Unlock()
	}

	return concurrency.NewMutex(session, lockName), nil
}

func (es *EtcdStateStore) Close() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	for _, cancel := range es.watchers {
		cancel()
	}
	es.watchers = make(map[string]context.CancelFunc)

	for _, session := range es.sessionPool {
		session.Close()
	}
	es.sessionPool = make(map[string]*concurrency.Session)

	return es.client.Close()
}
