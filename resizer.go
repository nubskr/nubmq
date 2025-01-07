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
	defer allowSets.Unlock()
	SetWG.Wait()
	for atomic.LoadInt32(&sm2.pendingRequests) != 0 {
		log.Fatal("lol, get rekt idiot")
		fmt.Println("--------------waiting for all requests to be processed----------------")
	}

	sm2.mutex.RLock()
	sm1.mutex.Lock()

	sm1.ShardManagers = sm2.ShardManagers
	sm1.totalCapacity = sm2.totalCapacity
	sm1.usedCapacity = sm2.usedCapacity // this won't be accurate btw, since sets are still in pipeline

	sm2.mutex.RUnlock()
	sm1.mutex.Unlock()

	// pull the darn cork out
	atomic.AddInt32(&sm1.isResizing, -1)
	// fmt.Println("switched to main sm-----------------------")
	// sm1 and sm2 point to the same memory location now

}

// migrates all keys from sm1 to sm2
func migrateKeys(sm1 *ShardManagerKeeperTemp, sm2 *ShardManagerKeeperTemp) {
	// log.Print("Starting to migrate keys")
	if (atomic.LoadInt64(&sm1.totalCapacity))*int64(2) > atomic.LoadInt64(&sm1.usedCapacity) {
		sm1.mutex.Unlock()
		log.Fatal("IDIOTTTTTT")
		return
	}

	// log.Print("Migrating keys in")

	allowSets.Lock()

	// log.Print("Sets halted,waiting for existing sets to complete")

	SetWG.Wait()

	// log.Print("pending sets completed")
	sm1.mutex.RLock()
	// fmt.Println("Migrating Keys start---------------")
	for _, SM := range sm1.ShardManagers {
		for _, Shard := range SM.Shards {
			pairs := make(map[interface{}]interface{})

			Shard.data.Range(func(key, value interface{}) bool {
				pairs[key] = value
				return true
			})

			// log.Print("Trying to get inside read lock on new table")
			for k, v := range pairs {
				// TODO: should we use SetQueue here too ?
				// fmt.Println("Migrating key", k)
				sm2.mutex.RLock()
				SetWG.Add(1)
				forceSetKey(k.(string), v.(string), sm2)
				sm2.mutex.RUnlock()
			}
		}
	}

	sm1.mutex.RUnlock()
	// fmt.Println("Migrating Keys end-----------------")

	switchTables(sm1, sm2)
}

func UpgradeShardManagerKeeper(currentSize int64) bool {
	// fmt.Println("UpgradeShardManagerKeeper triggered")
	if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 > atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) || atomic.LoadInt32(&ShardManagerKeeper.isResizing) != 0 || atomic.LoadInt64(&ShardManagerKeeper.totalCapacity) > currentSize {
		// fmt.Println("False alarm, skipping upgrade")
		return false
	}

	tempNewSM := getNewShardManagerKeeper(ShardManagerKeeper.totalCapacity * 2)

	newShardManagerKeeper.mutex.Lock()
	newShardManagerKeeper.ShardManagers = tempNewSM.ShardManagers
	newShardManagerKeeper.totalCapacity = tempNewSM.totalCapacity
	newShardManagerKeeper.usedCapacity = 0
	newShardManagerKeeper.mutex.Unlock()

	atomic.AddInt32(&ShardManagerKeeper.isResizing, 1)

	return true
}
