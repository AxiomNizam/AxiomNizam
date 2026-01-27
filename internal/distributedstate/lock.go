package distributedstate

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type DistributedLock struct {
	store    StateStore
	lockName string
	holder   string
	timeout  time.Duration
	mu       sync.Mutex
	leaseID  int64
}

func NewDistributedLock(store StateStore, lockName string, holder string, timeout time.Duration) *DistributedLock {
	return &DistributedLock{
		store:    store,
		lockName: lockName,
		holder:   holder,
		timeout:  timeout,
	}
}

func (dl *DistributedLock) Acquire(ctx context.Context) (bool, error) {
	key := fmt.Sprintf("lock/%s", dl.lockName)

	leaseID, err := dl.store.PutWithLease(ctx, key, dl.holder, dl.timeout)
	if err != nil {
		return false, err
	}

	success, err := dl.store.CompareAndSwap(ctx, key, "", dl.holder)
	if !success {
		dl.store.ReleaseLeaseID(ctx, leaseID)
		return false, err
	}

	dl.mu.Lock()
	dl.leaseID = leaseID
	dl.mu.Unlock()

	return true, nil
}

func (dl *DistributedLock) Release(ctx context.Context) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.leaseID == 0 {
		return fmt.Errorf("lock not held")
	}

	err := dl.store.ReleaseLeaseID(ctx, dl.leaseID)
	dl.leaseID = 0
	return err
}

func (dl *DistributedLock) IsHeld(ctx context.Context) (bool, string, error) {
	key := fmt.Sprintf("lock/%s", dl.lockName)
	value, err := dl.store.Get(ctx, key)
	if err != nil {
		return false, "", err
	}

	if value == "" {
		return false, "", nil
	}

	return true, value, nil
}

func (dl *DistributedLock) WaitForRelease(ctx context.Context, checkInterval time.Duration) error {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			held, _, err := dl.IsHeld(ctx)
			if err != nil {
				return err
			}
			if !held {
				return nil
			}
		}
	}
}

type DistributedLeaderElection struct {
	store       StateStore
	electionKey string
	node        string
	ttl         time.Duration
	callback    func(bool)
	mu          sync.Mutex
	leaseID     int64
	isLeader    bool
}

func NewDistributedLeaderElection(store StateStore, electionKey string, node string, ttl time.Duration) *DistributedLeaderElection {
	return &DistributedLeaderElection{
		store:       store,
		electionKey: electionKey,
		node:        node,
		ttl:         ttl,
		isLeader:    false,
	}
}

func (dle *DistributedLeaderElection) Start(ctx context.Context) error {
	ticker := time.NewTicker(dle.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := dle.tryElection(ctx); err != nil {
				return err
			}
		}
	}
}

func (dle *DistributedLeaderElection) tryElection(ctx context.Context) error {
	key := fmt.Sprintf("election/%s", dle.electionKey)

	currentLeader, err := dle.store.Get(ctx, key)
	if err != nil {
		return err
	}

	dle.mu.Lock()
	defer dle.mu.Unlock()

	if currentLeader == dle.node {
		leaseID, err := dle.store.PutWithLease(ctx, key, dle.node, dle.ttl)
		if err != nil {
			dle.isLeader = false
			if dle.callback != nil {
				dle.callback(false)
			}
			return err
		}

		if !dle.isLeader {
			dle.isLeader = true
			if dle.callback != nil {
				dle.callback(true)
			}
		}

		dle.leaseID = leaseID
		return nil
	}

	if currentLeader == "" {
		leaseID, err := dle.store.PutWithLease(ctx, key, dle.node, dle.ttl)
		if err != nil {
			return err
		}

		success, err := dle.store.CompareAndSwap(ctx, key, "", dle.node)
		if success {
			if !dle.isLeader {
				dle.isLeader = true
				if dle.callback != nil {
					dle.callback(true)
				}
			}
			dle.leaseID = leaseID
		} else {
			dle.store.ReleaseLeaseID(ctx, leaseID)
			if dle.isLeader {
				dle.isLeader = false
				if dle.callback != nil {
					dle.callback(false)
				}
			}
		}
		return nil
	}

	if dle.isLeader {
		dle.isLeader = false
		if dle.callback != nil {
			dle.callback(false)
		}
	}

	return nil
}

func (dle *DistributedLeaderElection) IsLeader() bool {
	dle.mu.Lock()
	defer dle.mu.Unlock()
	return dle.isLeader
}

func (dle *DistributedLeaderElection) GetLeader(ctx context.Context) (string, error) {
	key := fmt.Sprintf("election/%s", dle.electionKey)
	return dle.store.Get(ctx, key)
}

func (dle *DistributedLeaderElection) OnLeadershipChange(callback func(bool)) {
	dle.mu.Lock()
	defer dle.mu.Unlock()
	dle.callback = callback
}
