package main

import (
	"fmt"
	"sync/atomic"
)

func setAtIndex(idx int, key string, val string, keeper *ShardManagerKeeperTemp) {
	SMidx, localIdx := getShardNumberAndIndexPair(idx)
	target := keeper.ShardManagers[SMidx].Shards[localIdx]
	value, ok := target.data.Load(key)
	// fmt.Println("trying to set at global SM index", SMidx, "at local index", localIdx)
	if !ok {
		atomic.AddInt64(&ShardManagerKeeper.usedCapacity, 1)
	} else {
		fmt.Println("Ignore this log", value)
	}
	target.data.Store(key, val)
	atomic.AddInt32(&keeper.pendingRequests, -1)
}

// force inserts the key in sm without any checks, use with caution
func forceSetKey(key string, value string, sm *ShardManagerKeeperTemp) {
	atomic.AddInt32(&sm.pendingRequests, 1)
	setAtIndex(getKeyHash(key, sm), key, value, sm)
}

func _setKey(key string, value string) {
	/*
		get the key hash and ShardNumber from there

		from that ShardNumber update that value for that Shard

		TODO: check for capacity exceeding desired upper bound and then trigger resizing state
	*/

	// halt to switch tables
	for atomic.LoadInt32(&HaltSets) == 1 {
		fmt.Println("Sets-----x------Halted----------------------------------")
	}

	if atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 0 {
		atomic.AddInt32(&ShardManagerKeeper.pendingRequests, 1)
		fmt.Println("inserting in old table")
		ShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &ShardManagerKeeper), key, value, &ShardManagerKeeper)

		ShardManagerKeeper.mutex.RUnlock()

		if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 <= atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) {
			fmt.Println("triggering resizing")
			newShardManagerKeeper.mutex.Lock()
			UpgradeShardManagerKeeper()
			newShardManagerKeeper.mutex.Unlock()
			// BUG: this might not be necessary, given that this might be called unnecessarily, note that upgrades are not always needed, look into it, possibly add a condition where we even need to migrate keys
			go migrateKeys(&ShardManagerKeeper, &newShardManagerKeeper)
		}
	} else {
		atomic.AddInt32(&newShardManagerKeeper.pendingRequests, 1)
		fmt.Println("inserting in new table")

		// WARN: the newSMKeeper might not be fully resized at this exact piece of time, stupid concurrency
		newShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &newShardManagerKeeper), key, value, &newShardManagerKeeper)

		newShardManagerKeeper.mutex.RUnlock()
	}
}
