package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/axiomnizam/internal/distributedstate"
)

func ExampleEtcdStore() {
	store, err := distributedstate.NewEtcdStateStore([]string{"localhost:2379"})
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store.Put(ctx, "config/app", "active")

	value, err := store.Get(ctx, "config/app")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Config:", value)

	success, err := store.CompareAndSwap(ctx, "config/app", "active", "inactive")
	fmt.Println("CAS succeeded:", success)
}

func ExampleDistributedManager() {
	store := distributedstate.NewInMemoryStateStore()
	manager := distributedstate.NewDistributedManager(store, "myapp")

	ctx := context.Background()

	manager.PutState(ctx, "user/1/name", "John")
	manager.PutState(ctx, "user/1/email", "john@example.com")

	name, _ := manager.GetState(ctx, "user/1/name")
	fmt.Println("Name:", name)

	states, _ := manager.ListStates(ctx)
	fmt.Println("All states:", states)

	manager.UpdateStateIfUnchanged(ctx, "user/1/name", "John", "Jane")
	updated, _ := manager.GetState(ctx, "user/1/name")
	fmt.Println("Updated:", updated)
}

func ExampleDistributedLock() {
	store := distributedstate.NewInMemoryStateStore()
	ctx := context.Background()

	lock := distributedstate.NewDistributedLock(store, "critical-section", "node1", 10*time.Second)

	acquired, err := lock.Acquire(ctx)
	if err != nil || !acquired {
		log.Fatal("Failed to acquire lock")
	}
	fmt.Println("Lock acquired")

	defer lock.Release(ctx)

	fmt.Println("Critical section execution")
}

func ExampleLeaderElection() {
	store := distributedstate.NewInMemoryStateStore()
	ctx, cancel := context.WithCancel(context.Background())

	election := distributedstate.NewDistributedLeaderElection(store, "cluster", "node1", 2*time.Second)

	election.OnLeadershipChange(func(isLeader bool) {
		if isLeader {
			fmt.Println("Node1 became leader")
		} else {
			fmt.Println("Node1 lost leadership")
		}
	})

	go func() {
		if err := election.Start(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(5 * time.Second)
	cancel()
}

func ExampleDistributedCounter() {
	store := distributedstate.NewInMemoryStateStore()
	ctx := context.Background()

	counter := distributedstate.NewDistributedCounter(store, "visits")

	for i := 0; i < 5; i++ {
		val, _ := counter.Increment(ctx)
		fmt.Println("Counter:", val)
	}
}

func ExampleDistributedSet() {
	store := distributedstate.NewInMemoryStateStore()
	ctx := context.Background()

	set := distributedstate.NewDistributedSet(store, "users/active")

	set.Add(ctx, "user1")
	set.Add(ctx, "user2")
	set.Add(ctx, "user3")

	contains, _ := set.Contains(ctx, "user1")
	fmt.Println("Contains user1:", contains)

	size, _ := set.Size(ctx)
	fmt.Println("Set size:", size)

	members, _ := set.Members(ctx)
	fmt.Println("Members:", members)
}

func ExampleDistributedQueue() {
	store := distributedstate.NewInMemoryStateStore()
	ctx := context.Background()

	queue := distributedstate.NewDistributedQueue(store, "tasks")

	queue.Enqueue(ctx, "task1")
	queue.Enqueue(ctx, "task2")
	queue.Enqueue(ctx, "task3")

	task, _ := queue.Dequeue(ctx)
	fmt.Println("Dequeued:", task)

	size, _ := queue.Size(ctx)
	fmt.Println("Queue size:", size)
}

func ExampleCachedStore() {
	underlying := distributedstate.NewInMemoryStateStore()
	cached := distributedstate.NewCachedStateStore(underlying, 1*time.Second)

	ctx := context.Background()

	cached.Put(ctx, "key1", "value1")
	v1, _ := cached.Get(ctx, "key1")
	fmt.Println("First get (cached):", v1)

	v2, _ := cached.Get(ctx, "key1")
	fmt.Println("Second get (cached):", v2)

	time.Sleep(1100 * time.Millisecond)

	cached.Put(ctx, "key1", "value2")
	v3, _ := cached.Get(ctx, "key1")
	fmt.Println("After invalidation:", v3)

	cached.Close()
}

func ExampleBatchUpdate() {
	store := distributedstate.NewInMemoryStateStore()
	manager := distributedstate.NewDistributedManager(store, "app")
	ctx := context.Background()

	ops := []distributedstate.BatchOperation{
		{Key: "config/db", Value: "postgresql", Type: distributedstate.OperationPut},
		{Key: "config/cache", Value: "redis", Type: distributedstate.OperationPut},
		{Key: "config/queue", Value: "nats", Type: distributedstate.OperationPut},
	}

	success, _ := manager.BatchUpdate(ctx, ops)
	fmt.Println("Batch update succeeded:", success)

	states, _ := manager.ListStates(ctx)
	for k, v := range states {
		fmt.Printf("%s=%s\n", k, v)
	}
}
