package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

// makes and returns a presized SMkeeper pointer
func getNewShardManagerKeeper(sz int64) *ShardManagerKeeperTemp {
	curSz := 1

	var newSMkeeper = ShardManagerKeeperTemp{
		ShardManagers: make([]*ShardManager, 0),
		totalCapacity: 0,
		usedCapacity:  0,
		isResizing:    0,
	}

	for newSMkeeper.totalCapacity < sz {
		// append shit in that
		newSMkeeper.ShardManagers = append(newSMkeeper.ShardManagers, getNewShardManager(curSz))
		newSMkeeper.totalCapacity += int64(curSz)
		curSz *= 2 // WARN: this might overflow
	}

	return &newSMkeeper
}

// migrates all keys from sm1 to sm2
func migrateKeys(sm1 *ShardManagerKeeperTemp, sm2 *ShardManagerKeeperTemp) {
	/*

		ShardManagerKeeper
			ShardManager..1.2.3..
				Shard..1.2.3..
					ValueData

	*/

	sm1.mutex.Lock()
	for _, SM := range sm1.ShardManagers {
		for _, Shard := range SM.Shards {
			pairs := make(map[interface{}]interface{})

			// Fetch key-value pairs using Range
			Shard.data.Range(func(key, value interface{}) bool {
				pairs[key] = value
				return true
			})

			for k, v := range pairs {
				go forceSetKey(k.(string), v.(string), sm2)
			}
		}
		time.Sleep(10 * time.Microsecond) // TODO: please find a better way to do this thing ffs
	}

	sm2.mutex.Lock()
	sm1.ShardManagers = sm2.ShardManagers // does this copies the darn pointers ?
	sm1.isResizing = 0
	sm2.mutex.Unlock()
	sm1.mutex.Unlock()
	// now we need to somehow swap SMkeeper with newSMkeeper "safely"
	// make SMkeeper point to newSMkeeper and , no, essentially swap the pointers of S

}

// Adds one more layer of SM to SMkeeper
func UpgradeShardManagerKeeper() {
	fmt.Println("UpgradeShardManagerKeeper triggered")
	/*
		first check if even need to resize

		this thing, once triggered will change the system mode to resizing

		then we make a newSM with double the size of current SM

		once made, then take each key, rehash it, and then insert it into newer SMkeeper, do it slowly as to now overwhelm the new table and affect new incoming sets

		once this all is done, we swap the SMkeeper pointer to the newer SMkeeper

		and change the system mode back to normal
	*/

	if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 > atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) || atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 1 {
		return
	}

	atomic.AddInt32(&ShardManagerKeeper.isResizing, 1)

	// newShardManagerKeeper.mutex.Lock()

	tempNewSM := getNewShardManagerKeeper(ShardManagerKeeper.totalCapacity)

	newShardManagerKeeper.ShardManagers = tempNewSM.ShardManagers
	newShardManagerKeeper.totalCapacity = tempNewSM.totalCapacity
	newShardManagerKeeper.usedCapacity = 0

	// newShardManagerKeeper.mutex.Unlock()
	/*
		TODO: start migrating those keys
			now we can start migrating those keys somehow, make another go routine for that, which keeps running in the backgroud and does stuff
	*/

}
