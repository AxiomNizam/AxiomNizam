package distributedstate

import (
	"context"
	"time"
)

type StateStore interface {
	Put(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	CompareAndSwap(ctx context.Context, key string, oldValue, newValue string) (bool, error)
	Watch(ctx context.Context, key string, callback func(Event)) error
	GetWithRevision(ctx context.Context, key string) (string, int64, error)
	PutWithLease(ctx context.Context, key string, value string, ttl time.Duration) (int64, error)
	ReleaseLeaseID(ctx context.Context, leaseID int64) error
	List(ctx context.Context, prefix string) (map[string]string, error)
	Transaction(ctx context.Context, ops []Operation) (bool, error)
	Close() error
}

type Event struct {
	Type   EventType
	Key    string
	Value  string
	OldValue string
	Revision int64
	Timestamp time.Time
}

type EventType string

const (
	EventTypePut    EventType = "PUT"
	EventTypeDelete EventType = "DELETE"
)

type Operation struct {
	Type  OperationType
	Key   string
	Value string
	OldValue string
}

type OperationType string

const (
	OperationPut     OperationType = "PUT"
	OperationDelete  OperationType = "DELETE"
	OperationCompare OperationType = "COMPARE"
)

type LeaseOption struct {
	TTL int64
}
