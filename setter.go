package main

import (
	"fmt"
	"log"
	"sync/atomic"
)

func setAtIndex(idx int, key string, val string, keeper *ShardManagerKeeperTemp) {
	defer SetWG.Done()
	SMidx, localIdx := getShardNumberAndIndexPair(idx)

	keeper.ShardManagers[SMidx].mutex.RLock()
	targetSM := keeper.ShardManagers[SMidx]
	keeper.ShardManagers[SMidx].mutex.RUnlock()

	targetSM.mutex.RLock()
	target := targetSM.Shards[localIdx]
	targetSM.mutex.RUnlock()
	value, ok := target.data.Load(key)

	if !ok {
		atomic.AddInt64(&ShardManagerKeeper.usedCapacity, 1)
	} else {
		fmt.Println("Ignore this log", value)
	}
	target.data.Store(key, val)
}

// force inserts the key in sm without any checks, use with caution
func forceSetKey(key string, value string, sm *ShardManagerKeeperTemp) {
	sm.mutex.RLock()
	log.Print("+")
	setAtIndex(getKeyHash(key, sm), key, value, sm)
	log.Print("-")
	sm.mutex.RUnlock()
}

func _setKey(key string, value string) {

	if atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 0 {
		fmt.Println("inserting in old table")
		ShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &ShardManagerKeeper), key, value, &ShardManagerKeeper)

		ShardManagerKeeper.mutex.RUnlock()

		if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 <= atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) { // very hit and miss, will NOT work
			newShardManagerKeeper.mutex.Lock()
			ShardManagerKeeper.mutex.Lock()
			migrateOrNot := UpgradeShardManagerKeeper(ShardManagerKeeper.totalCapacity)
			ShardManagerKeeper.mutex.Unlock()
			newShardManagerKeeper.mutex.Unlock()
			if migrateOrNot {
				// UpgradeProcessWG.Wait()
				UpgradeProcessWG.Add(1)
				fmt.Println("triggering resizing")
				go migrateKeys(&ShardManagerKeeper, &newShardManagerKeeper)
			}
		}
	} else {
		fmt.Println("inserting in new table")

		newShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &newShardManagerKeeper), key, value, &newShardManagerKeeper)

		newShardManagerKeeper.mutex.RUnlock()
	}
}
