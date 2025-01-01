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
	switchTablesWG.Add(1)
	defer switchTablesWG.Done()

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
	atomic.AddInt32(&HaltSets, -1)
	UpgradeProcessWG.Done()
	fmt.Println("switched to main sm-----------------------")
}

// migrates all keys from sm1 to sm2
func migrateKeys(sm1 *ShardManagerKeeperTemp, sm2 *ShardManagerKeeperTemp) {

	// migrateKeysWG.Add(1)
	// defer migrateKeysWG.Done()

	if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 > atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) {
		return
	}

	log.Print("Migrating keys in")
	/*

		ShardManagerKeeper
			ShardManager..1.2.3..
				Shard..1.2.3..
					ValueData

	*/

	atomic.AddInt32(&HaltSets, 1)

	SetWG.Wait() // wait for sets in sm1 to settle down

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
				fmt.Println("Migrating key", k, "value", v)
				SetWG.Add(1)
				go forceSetKey(k.(string), v.(string), sm2)
				// sm2.mutex.RUnlock()
			}
		}
		// time.Sleep(1 * time.Microsecond) // TODO: please find a better way to do this thing ffs
	}

	sm1.mutex.Unlock()

	SetWG.Wait() // maybe let all migrations end first ?

	fmt.Println("Migrating Keys end-----------------")

	// TODO: UNCOMMENT BELOW STUFF, YOU NEED IT, debugging rn

	// atomic.AddInt32(&HaltSets, 1)

	switchTablesWG.Wait()
	go switchTables(sm1, sm2)
}

// Adds one more layer of SM to newSMkeeper
// BUG: this shit is returning true more often than it needs to

func UpgradeShardManagerKeeper(currentSize int64) bool {
	// A LOT OF THIS ARE BEING TRIGGERED PARALLELY, ITS NOT FUCKING THREAD SAFE, ATOMIC IS FUCKING RISKY DUMBO
	SMUpgradeMutex.Lock()
	defer SMUpgradeMutex.Unlock()

	fmt.Println("UpgradeShardManagerKeeper triggered")
	if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 > atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) || atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 1 || atomic.LoadInt64(&ShardManagerKeeper.totalCapacity) > currentSize {
		return false
	}

	tempNewSM := getNewShardManagerKeeper(ShardManagerKeeper.totalCapacity * 2)

	newShardManagerKeeper.ShardManagers = tempNewSM.ShardManagers
	newShardManagerKeeper.totalCapacity = tempNewSM.totalCapacity
	newShardManagerKeeper.usedCapacity = 0

	atomic.AddInt32(&ShardManagerKeeper.isResizing, 1)

	return true
}
