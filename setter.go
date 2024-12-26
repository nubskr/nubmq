package main

import (
	"fmt"
	"sync/atomic"
)

func setAtIndex(idx int, key string, val string, keeper *ShardManagerKeeperTemp) {
	SMidx, localIdx := getShardNumberAndIndexPair(idx)
	// fmt.Println("ping", SMidx, localIdx)
	target := keeper.ShardManagers[SMidx].Shards[localIdx]
	// is there any existing value for this key ? if not we increase the used capacity to keep track of active keys
	value, ok := target.data.Load(key)
	// fmt.Println("trying to set at global SM index", SMidx, "at local index", localIdx)
	if !ok {
		// atomically increase by one
		atomic.AddInt64(&ShardManagerKeeper.usedCapacity, 1)
	} else {
		fmt.Println("Ignore this log", value)
	}
	target.data.Store(key, val)
}

// force inserts the key in sm without any checks, use with caution
func forceSetKey(key string, value string, sm *ShardManagerKeeperTemp) {
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
		fmt.Println("inserting in old table")
		fmt.Println("in")
		ShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &ShardManagerKeeper), key, value, &ShardManagerKeeper)

		ShardManagerKeeper.mutex.RUnlock()

		if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 <= atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) {
			// theoretically it should always trigger this before trying to set something which it can't set it
			// just be mindful of that
			fmt.Println("triggering resizing")
			newShardManagerKeeper.mutex.Lock()
			UpgradeShardManagerKeeper()
			newShardManagerKeeper.mutex.Unlock()
			// BUG: this might not be necessary, given that this might be called unneceaarily, note that upgrades are not always needed, look into it, possibly add a condition where we even need to migrate keys
			go migrateKeys(&ShardManagerKeeper, &newShardManagerKeeper)
		}
		fmt.Println("out")
	} else {
		fmt.Println("inserting in new table")
		// put stuff in the new table

		// WARN: the newSMKeeper might not be fully resized at this exact piece of time, stupid concurrency
		newShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &newShardManagerKeeper), key, value, &newShardManagerKeeper)

		newShardManagerKeeper.mutex.RUnlock()
	}
}
