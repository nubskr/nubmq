package main

import (
	"fmt"
	"log"
	"sync/atomic"
)

func setAtIndex(idx int, key string, val string, keeper *ShardManagerKeeperTemp, request SetRequest) {
	defer SetWG.Done()
	defer log.Print("Insertion done for key: ", key)
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
	request.status <- struct{}{}
}

func setAtIndexLazy(idx int, key string, val string, keeper *ShardManagerKeeperTemp) {
	defer SetWG.Done()
	defer log.Print("Insertion done for key: ", key)
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
	// sm.mutex.RLock()
	log.Print("+")
	setAtIndexLazy(getKeyHash(key, sm), key, value, sm)
	log.Print("-")
	// sm.mutex.RUnlock()
}

func _setKey(request SetRequest) {
	key := request.key
	value := request.value
	if atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 0 {
		fmt.Println("inserting in old table key: ", key)
		ShardManagerKeeper.mutex.RLock()
		// log.Print("Getting read lock in old table")

		setAtIndex(getKeyHash(key, &ShardManagerKeeper), key, value, &ShardManagerKeeper, request)

		ShardManagerKeeper.mutex.RUnlock()
		// log.Print("releasing read lock in old table")

		if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 <= atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) { // very hit and miss, will NOT work

			// log.Print("sm2 full lock acquired")
			ShardManagerKeeper.mutex.Lock()
			migrateOrNot := UpgradeShardManagerKeeper(ShardManagerKeeper.totalCapacity)
			ShardManagerKeeper.mutex.Unlock()

			// log.Print("sm2 full lock released")

			if migrateOrNot {
				fmt.Println("triggering resizing")
				go migrateKeys(&ShardManagerKeeper, &newShardManagerKeeper)
			}
		}
	} else {
		fmt.Println("inserting in new table key: ", key)

		newShardManagerKeeper.mutex.RLock()
		log.Print("Getting read lock in new table")

		setAtIndex(getKeyHash(key, &newShardManagerKeeper), key, value, &newShardManagerKeeper, request)

		newShardManagerKeeper.mutex.RUnlock()
		log.Print("releasing read lock in new table")
	}
}

func handleSetWorker() {
	log.Print("Worker started")

	for {
		setReq := <-setQueue
		_setKey(setReq)
	}
}
