package main

import (
	"fmt"
	"log"
	"sync/atomic"
)

func setAtIndex(idx int, key string, val string, keeper *ShardManagerKeeperTemp) {
	SMidx, localIdx := getShardNumberAndIndexPair(idx)

	keeper.ShardManagers[SMidx].mutex.RLock()
	targetSM := keeper.ShardManagers[SMidx]
	keeper.ShardManagers[SMidx].mutex.RUnlock()

	targetSM.mutex.RLock()
	target := targetSM.Shards[localIdx]
	targetSM.mutex.RUnlock()
	value, ok := target.data.Load(key)

	// fmt.Println("trying to set at global SM index", SMidx, "at local index", localIdx)
	if !ok {
		atomic.AddInt64(&ShardManagerKeeper.usedCapacity, 1)
	} else {
		fmt.Println("Ignore this log", value)
	}
	target.data.Store(key, val)
	atomic.AddInt32(&keeper.pendingRequests, -1)
	SetWG.Done()
}

// force inserts the key in sm without any checks, use with caution
func forceSetKey(key string, value string, sm *ShardManagerKeeperTemp) {
	// if atomic.LoadInt32(&HaltSets) == 1 {
	// 	log.Fatal("The world is ending sire forcedddd", atomic.LoadInt32(&HaltSets))
	// }
	// for atomic.LoadInt32(&HaltSets) == 1 {
	// 	fmt.Println("ForcedSets-----x------Halted----------------------------------")
	// }

	atomic.AddInt32(&sm.pendingRequests, 1)
	sm.mutex.RLock()
	setAtIndex(getKeyHash(key, sm), key, value, sm)
	sm.mutex.RUnlock()
}

func _setKey(key string, value string) {
	if atomic.LoadInt32(&ShardManagerKeeper.isResizing) != 0 && atomic.LoadInt32(&ShardManagerKeeper.isResizing) != 1 {
		log.Fatal("GGWP")
	}
	/*
		get the key hash and ShardNumber from there

		from that ShardNumber update that value for that Shard

		TODO: check for capacity exceeding desired upper bound and then trigger resizing state
	*/

	// halt to switch tables
	val1 := atomic.LoadInt32(&HaltSets)
	if val1 != 1 && val1 != 0 {
		log.Fatal("The world is ending sire ", val1)
	}
	for atomic.LoadInt32(&HaltSets) == 1 {
		// fmt.Println("Sets-----x------Halted----------------------------------")
	}

	// SetWG.Add(1)

	if atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 0 {
		atomic.AddInt32(&ShardManagerKeeper.pendingRequests, 1)
		fmt.Println("inserting in old table")
		ShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &ShardManagerKeeper), key, value, &ShardManagerKeeper)

		ShardManagerKeeper.mutex.RUnlock()

		if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 <= atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) { // very hit and miss, will NOT work
			newShardManagerKeeper.mutex.Lock()
			migrateOrNot := UpgradeShardManagerKeeper(atomic.LoadInt64(&ShardManagerKeeper.totalCapacity))
			newShardManagerKeeper.mutex.Unlock()
			// BUG: this might not be necessary, given that this might be called unnecessarily, note that upgrades are not always needed, look into it, possibly add a condition where we even need to migrate keys
			if migrateOrNot {
				UpgradeProcessWG.Wait()
				UpgradeProcessWG.Add(1)
				fmt.Println("triggering resizing")
				go migrateKeys(&ShardManagerKeeper, &newShardManagerKeeper)
			}
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
