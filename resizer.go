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

func switchTables(sm1 *ShardManagerKeeperTemp, sm2 *ShardManagerKeeperTemp) {
	// essentially give some time for the newTable to get all the SETs settled in, the below thing is NOT reliable at all
	time.Sleep(500 * time.Microsecond) // too long ? // TODO: remove this shit for something more trusty you idiot, what the fuck is even this, are you a fucking clown ?

	// do stuff
	sm2.mutex.Lock()

	// make sm1 point to sm2's memory
	sm1.ShardManagers = sm2.ShardManagers
	sm1.totalCapacity = sm2.totalCapacity
	sm1.usedCapacity = sm2.usedCapacity
	// dereference sm2 and make it point to an empty SMkeeper object

	// sm2 = &ShardManagerKeeperTemp{
	// 	ShardManagers: nil,
	// 	totalCapacity: 0,
	// 	usedCapacity:  0,
	// 	isResizing:    0,
	// }

	sm2.mutex.Unlock()

	// sm2 = getNewShardManagerKeeper(1) // BUG: NOT THREAD SAFE, but we assume no ops are happening to sm2

	// pull the darn cork out
	atomic.AddInt32(&HaltSets, -1)
	fmt.Println("switched to main sm-----------------------")
	atomic.AddInt32(&sm1.isResizing, -1)
}

// migrates all keys from sm1 to sm2
func migrateKeys(sm1 *ShardManagerKeeperTemp, sm2 *ShardManagerKeeperTemp) {
	if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 > atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) {
		return
	}
	/*

		ShardManagerKeeper
			ShardManager..1.2.3..
				Shard..1.2.3..
					ValueData

	*/

	sm1.mutex.Lock()
	fmt.Println("Migrating Keys start---------------")
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
		// time.Sleep(1 * time.Microsecond) // TODO: please find a better way to do this thing ffs
	}

	// sm2.mutex.Lock()

	// sm1.ShardManagers = sm2.ShardManagers // make it point to that pointer
	// sm1.totalCapacity = int64(sm2.totalCapacity)
	// BUG: start here, note that we can't just make the usedCapacity as sm2.usedCapacity

	// sm1.isResizing = 0 // make sure this is the last thing you do!

	// sm2.mutex.Unlock()
	sm1.mutex.Unlock()
	fmt.Println("Migrating Keys end-----------------")

	atomic.AddInt32(&HaltSets, 1)

	go switchTables(sm1, sm2)
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
