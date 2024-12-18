package main

import (
	"fmt"
	"sync/atomic"
)

func setAtIndex(idx int, key string, val string, keeper *ShardManagerKeeperTemp) {
	SMidx, localIdx := getShardNumberAndIndexPair(idx)
	fmt.Println("ping", SMidx, localIdx)
	target := keeper.ShardManagers[SMidx].Shards[localIdx]
	// is there any existing value for this key ? if not we increase the used capacity to keep track of active keys
	value, ok := target.data.Load(key)
	fmt.Println("trying to set at global SM index", SMidx, "at local index", localIdx)
	if !ok {
		// atomically increase by one
		atomic.AddInt64(&ShardManagerKeeper.usedCapacity, 1)
	} else {
		fmt.Println("Ignore this log", value)
	}
	target.data.Store(key, val)
}

func _setKey(key string, value string) {
	/*
		get the key hash and ShardNumber from there

		from that ShardNumber update that value for that Shard

		TODO: check for capacity exceeding desired upper bound and then trigger resizing state
	*/
	if atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 0 {
		fmt.Println("in")
		ShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &ShardManagerKeeper), key, value, &ShardManagerKeeper)

		ShardManagerKeeper.mutex.RUnlock()

		if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 <= atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) {
			// theoretically it should always trigger this before trying to set something which it can't set it
			// just be mindful of that
			go UpgradeShardManagerKeeper()
		}
		fmt.Println("out")
	} else {
		// put stuff in the new table
		newShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &newShardManagerKeeper), key, value, &newShardManagerKeeper)

		newShardManagerKeeper.mutex.RUnlock()
	}
}
