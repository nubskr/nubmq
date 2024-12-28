package main

import (
	"fmt"
	"log"
	"sync/atomic"
)

// makes and returns a presized SMkeeper pointer
func getNewShardManagerKeeper(sz int64) *ShardManagerKeeperTemp {
	curSz := 1

	var newSMkeeper = ShardManagerKeeperTemp{
		ShardManagers:   make([]*ShardManager, 0),
		totalCapacity:   0,
		usedCapacity:    0,
		isResizing:      0,
		pendingRequests: 0,
	}

	for newSMkeeper.totalCapacity < sz {
		newSMkeeper.ShardManagers = append(newSMkeeper.ShardManagers, getNewShardManager(curSz))
		newSMkeeper.totalCapacity += int64(curSz)
		curSz *= 2 // WARN: this might overflow
	}

	return &newSMkeeper
}

func switchTables(sm1 *ShardManagerKeeperTemp, sm2 *ShardManagerKeeperTemp) {
	// time.Sleep(1 * time.Second)
	for atomic.LoadInt32(&sm2.pendingRequests) != 0 {
		// fmt.Println(atomic.LoadInt32(&sm2.pendingRequests))
		fmt.Println("--------------waiting for all requests to be processed----------------")
	}

	// os.Exit(1)

	if atomic.LoadInt32(&sm2.pendingRequests) != 0 {
		log.Fatal("WE FUCKED UP SIRE")
	}
	// do stuff
	sm2.mutex.Lock()
	sm1.mutex.Lock()

	// make sm1 point to sm2's memory
	sm1.ShardManagers = sm2.ShardManagers
	sm1.totalCapacity = sm2.totalCapacity
	sm1.usedCapacity = sm2.usedCapacity

	// TODO: dereference sm2 and make it point to an empty SMkeeper object

	// sm2 = &ShardManagerKeeperTemp{
	// 	ShardManagers: nil,
	// 	totalCapacity: 0,
	// 	usedCapacity:  0,
	// 	isResizing:    0,
	// }

	sm2.mutex.Unlock()
	sm1.mutex.Unlock()

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
				// sm2.mutex.RLock()
				go forceSetKey(k.(string), v.(string), sm2)
				// sm2.mutex.RUnlock()
			}
		}
		// time.Sleep(1 * time.Microsecond) // TODO: please find a better way to do this thing ffs
	}

	sm1.mutex.Unlock()
	fmt.Println("Migrating Keys end-----------------")

	atomic.AddInt32(&HaltSets, 1)

	go switchTables(sm1, sm2)
}

// Adds one more layer of SM to SMkeeper
func UpgradeShardManagerKeeper() {
	fmt.Println("UpgradeShardManagerKeeper triggered")
	if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 > atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) || atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 1 {
		return
	}

	tempNewSM := getNewShardManagerKeeper(ShardManagerKeeper.totalCapacity)

	newShardManagerKeeper.ShardManagers = tempNewSM.ShardManagers
	newShardManagerKeeper.totalCapacity = tempNewSM.totalCapacity
	newShardManagerKeeper.usedCapacity = 0

	atomic.AddInt32(&ShardManagerKeeper.isResizing, 1)

}
