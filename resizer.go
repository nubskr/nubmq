package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
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
	defer UpgradeProcessWG.Done()
	SetWG.Wait()
	for atomic.LoadInt32(&sm2.pendingRequests) != 0 {
		log.Fatal("lol, get rekt idiot")
		fmt.Println("--------------waiting for all requests to be processed----------------")
	}

	// do stuff
	sm2.mutex.Lock()
	sm1.mutex.Lock()

	// make sm1 point to sm2's memory
	sm1.ShardManagers = sm2.ShardManagers
	sm1.totalCapacity = sm2.totalCapacity
	sm1.usedCapacity = sm2.usedCapacity

	// TODO: dereference sm2 and make it point to an empty SMkeeper object

	sm2.mutex.Unlock()
	sm1.mutex.Unlock()

	// sm2 = getNewShardManagerKeeper(0)
	// pull the darn cork out
	atomic.AddInt32(&sm1.isResizing, -1)
	// atomic.AddInt32(&HaltSets, -1)
	HaltSetsMutex.Lock()
	HaltSets = 0
	HaltSetsMutex.Unlock()

	// UpgradeProcessWG.Done()
	fmt.Println("switched to main sm-----------------------")
}

// migrates all keys from sm1 to sm2
func migrateKeys(sm1 *ShardManagerKeeperTemp, sm2 *ShardManagerKeeperTemp) {
	sm1.mutex.Lock()
	if sm1.totalCapacity*2 > sm1.usedCapacity {
		sm1.mutex.Unlock()
		log.Fatal("IDIOTTTTTT")
		return
	}
	sm1.mutex.Unlock()

	log.Print("Migrating keys in")

	HaltSetsMutex.Lock()
	HaltSets = 1
	HaltSetsMutex.Unlock()
	log.Print("Sets halted,waiting for existing sets to complete")

	// this sleep helps avoid those corner scenarios where a set is added right after wait() is over
	time.Sleep(3 * time.Millisecond) // TODO: please find a better way to do this thing ffs, something event based maybe ?

	SetWG.Wait() // wait for sets in sm1 to settle down

	log.Print("pending sets completed")
	sm1.mutex.RLock()
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
				fmt.Println("Migrating key", k)
				SetWG.Add(1)
				forceSetKey(k.(string), v.(string), sm2)
				// sm2.mutex.RUnlock()
			}
		}
		// time.Sleep(1 * time.Microsecond) // TODO: please find a better way to do this thing ffs
	}

	sm1.mutex.RUnlock()
	fmt.Println("Migrating Keys end-----------------")

	switchTables(sm1, sm2)
}

func UpgradeShardManagerKeeper(currentSize int64) bool {
	SMUpgradeMutex.Lock()
	defer SMUpgradeMutex.Unlock()

	fmt.Println("UpgradeShardManagerKeeper triggered")
	if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 > atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) || atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 1 || atomic.LoadInt64(&ShardManagerKeeper.totalCapacity) > currentSize {
		fmt.Println("False alarm, skipping upgrade")
		return false
	}

	tempNewSM := getNewShardManagerKeeper(ShardManagerKeeper.totalCapacity * 2)

	newShardManagerKeeper.ShardManagers = tempNewSM.ShardManagers
	newShardManagerKeeper.totalCapacity = tempNewSM.totalCapacity
	newShardManagerKeeper.usedCapacity = 0

	atomic.AddInt32(&ShardManagerKeeper.isResizing, 1)

	return true
}
