package distributedstate

import (
	"context"
	"testing"
	"time"
)

func TestBasicPutGet(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	if err := store.Put(ctx, "key1", "value1"); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	value, err := store.Get(ctx, "key1")
	if err != nil || value != "value1" {
		t.Fatalf("Get failed: expected 'value1', got '%s'", value)
	}
}

func TestCompareAndSwap(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	store.Put(ctx, "key", "old")

	success, err := store.CompareAndSwap(ctx, "key", "old", "new")
	if err != nil || !success {
		t.Fatalf("CAS failed: %v", err)
	}

	success, err = store.CompareAndSwap(ctx, "key", "old", "should_fail")
	if err != nil || success {
		t.Fatalf("CAS should have failed")
	}

	value, _ := store.Get(ctx, "key")
	if value != "new" {
		t.Fatalf("Expected 'new', got '%s'", value)
	}
}

func TestDistributedCounter(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	counter := NewDistributedCounter(store, "counter")

	for i := 1; i <= 10; i++ {
		val, err := counter.Increment(ctx)
		if err != nil {
			t.Fatalf("Increment failed: %v", err)
		}
		if val != int64(i) {
			t.Fatalf("Expected %d, got %d", i, val)
		}
	}

	current, _ := counter.Get(ctx)
	if current != 10 {
		t.Fatalf("Expected 10, got %d", current)
	}
}

func TestDistributedSet(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	set := NewDistributedSet(store, "set")

	set.Add(ctx, "member1")
	set.Add(ctx, "member2")

	contains, _ := set.Contains(ctx, "member1")
	if !contains {
		t.Fatalf("Set should contain member1")
	}

	size, _ := set.Size(ctx)
	if size != 2 {
		t.Fatalf("Expected size 2, got %d", size)
	}

	set.Remove(ctx, "member1")
	size, _ = set.Size(ctx)
	if size != 1 {
		t.Fatalf("Expected size 1 after remove, got %d", size)
	}
}

func TestDistributedQueue(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	queue := NewDistributedQueue(store, "queue")

	queue.Enqueue(ctx, "item1")
	queue.Enqueue(ctx, "item2")
	queue.Enqueue(ctx, "item3")

	size, _ := queue.Size(ctx)
	if size != 3 {
		t.Fatalf("Expected size 3, got %d", size)
	}

	item, _ := queue.Dequeue(ctx)
	if item == "" {
		t.Fatalf("Dequeue returned empty")
	}

	size, _ = queue.Size(ctx)
	if size != 2 {
		t.Fatalf("Expected size 2 after dequeue, got %d", size)
	}
}

func TestDistributedLock(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	lock1 := NewDistributedLock(store, "lock", "holder1", 5*time.Second)

	acquired, err := lock1.Acquire(ctx)
	if err != nil || !acquired {
		t.Fatalf("Acquire failed: %v", err)
	}

	held, holder, _ := lock1.IsHeld(ctx)
	if !held || holder != "holder1" {
		t.Fatalf("Lock state incorrect")
	}

	lock1.Release(ctx)

	held, _, _ = lock1.IsHeld(ctx)
	if held {
		t.Fatalf("Lock should be released")
	}
}

func TestLeaderElection(t *testing.T) {
	store := NewInMemoryStateStore()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	node1 := NewDistributedLeaderElection(store, "election", "node1", 500*time.Millisecond)

	go func() {
		node1.Start(ctx)
	}()

	time.Sleep(600 * time.Millisecond)

	if !node1.IsLeader() {
		t.Fatalf("Node1 should be leader")
	}

	leader, _ := node1.GetLeader(ctx)
	if leader != "node1" {
		t.Fatalf("Expected leader to be node1, got %s", leader)
	}
}

func TestCachedStore(t *testing.T) {
	underlying := NewInMemoryStateStore()
	cached := NewCachedStateStore(underlying, 1*time.Second)
	ctx := context.Background()

	cached.Put(ctx, "key", "value1")

	v1, _ := cached.Get(ctx, "key")
	if v1 != "value1" {
		t.Fatalf("Expected 'value1', got '%s'", v1)
	}

	cached.Invalidate("key")

	cached.Put(ctx, "key", "value2")
	v2, _ := cached.Get(ctx, "key")
	if v2 != "value2" {
		t.Fatalf("Expected 'value2', got '%s'", v2)
	}
}

func TestDistributedManager(t *testing.T) {
	store := NewInMemoryStateStore()
	manager := NewDistributedManager(store, "app")
	ctx := context.Background()

	manager.PutState(ctx, "config/db", "postgres")

	value, _ := manager.GetState(ctx, "config/db")
	if value != "postgres" {
		t.Fatalf("Expected 'postgres', got '%s'", value)
	}

	success, _ := manager.UpdateStateIfUnchanged(ctx, "config/db", "postgres", "mysql")
	if !success {
		t.Fatalf("CAS should succeed")
	}

	value, _ = manager.GetState(ctx, "config/db")
	if value != "mysql" {
		t.Fatalf("Expected 'mysql', got '%s'", value)
	}
}

func TestBatchOperations(t *testing.T) {
	store := NewInMemoryStateStore()
	manager := NewDistributedManager(store, "batch")
	ctx := context.Background()

	ops := []BatchOperation{
		{Key: "key1", Value: "value1", Type: OperationPut},
		{Key: "key2", Value: "value2", Type: OperationPut},
		{Key: "key3", Value: "value3", Type: OperationPut},
	}

	success, _ := manager.BatchUpdate(ctx, ops)
	if !success {
		t.Fatalf("Batch update failed")
	}

	v1, _ := manager.GetState(ctx, "key1")
	if v1 != "value1" {
		t.Fatalf("Expected 'value1', got '%s'", v1)
	}
}

func BenchmarkPutGet(b *testing.B) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Put(ctx, "key", "value")
		store.Get(ctx, "key")
	}
}

func BenchmarkCompareAndSwap(b *testing.B) {
	store := NewInMemoryStateStore()
	ctx := context.Background()

	store.Put(ctx, "key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.CompareAndSwap(ctx, "key", "value", "new_value")
		store.CompareAndSwap(ctx, "key", "new_value", "value")
	}
}

func BenchmarkCounter(b *testing.B) {
	store := NewInMemoryStateStore()
	ctx := context.Background()
	counter := NewDistributedCounter(store, "counter")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		counter.Increment(ctx)
	}
}
