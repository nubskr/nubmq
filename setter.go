package main

import (
	"fmt"
	"log"
	"sync/atomic"
)

func setAtIndex(idx int, keeper *ShardManagerKeeperTemp, request SetRequest) {
	defer SetWG.Done()
	key := request.key
	val := request.value
	canExpire := request.canExpire
	TTL := request.TTL

	SMidx, localIdx := getShardNumberAndIndexPair(idx)

	keeper.ShardManagers[SMidx].mutex.RLock()
	targetSM := keeper.ShardManagers[SMidx]
	keeper.ShardManagers[SMidx].mutex.RUnlock()

	targetSM.mutex.RLock()
	target := targetSM.Shards[localIdx]
	targetSM.mutex.RUnlock()
	value, ok := target.data.Load(key)

	if !ok {
		atomic.AddInt64(&keeper.usedCapacity, 1)
	} else {
		fmt.Println("Ignore this log", value)
	}
	entry := Entry{
		value:     val,
		canExpire: canExpire,
		TTL:       TTL,
	}

	target.data.Store(key, entry)
	request.status <- struct{}{}
}

func setAtIndexLazy(idx int, keeper *ShardManagerKeeperTemp, request SetRequest) {
	defer SetWG.Done()
	key := request.key
	val := request.value
	canExpire := request.canExpire
	TTL := request.TTL

	SMidx, localIdx := getShardNumberAndIndexPair(idx)
	keeper.ShardManagers[SMidx].mutex.RLock()
	targetSM := keeper.ShardManagers[SMidx]
	keeper.ShardManagers[SMidx].mutex.RUnlock()

	targetSM.mutex.RLock()
	target := targetSM.Shards[localIdx]
	targetSM.mutex.RUnlock()
	value, ok := target.data.Load(key)

	if !ok {
		atomic.AddInt64(&keeper.usedCapacity, 1)
	} else {
		fmt.Println("Ignore this log", value)
	}
	entry := Entry{
		value:     val,
		canExpire: canExpire,
		TTL:       TTL,
	}

	target.data.Store(key, entry)
}

// force inserts the key in sm without any checks, use with caution
func forceSetKey(request SetRequest, sm *ShardManagerKeeperTemp) {
	setAtIndexLazy(getKeyHash(request.key, sm), sm, request)
}

func _setKey(request SetRequest) {
	defer log.Print("Used capacity changed to: ", atomic.LoadInt64(&ShardManagerKeeper.usedCapacity))
	defer log.Print("Total capacity changed to: ", atomic.LoadInt64(&ShardManagerKeeper.totalCapacity))
	key := request.key
	if atomic.LoadInt32(&ShardManagerKeeper.isResizing) == 0 {
		log.Print("inserting in old table key: ", key)
		ShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &ShardManagerKeeper), &ShardManagerKeeper, request)

		ShardManagerKeeper.mutex.RUnlock()

		// check for upgrades
		if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity)*2 <= atomic.LoadInt64(&ShardManagerKeeper.usedCapacity) { // very hit and miss, will NOT work

			ShardManagerKeeper.mutex.Lock()
			migrateOrNot := UpgradeShardManagerKeeper(ShardManagerKeeper.totalCapacity)
			ShardManagerKeeper.mutex.Unlock()

			if migrateOrNot {
				// fmt.Println("triggering resizing")
				go migrateKeys(&ShardManagerKeeper, &newShardManagerKeeper)
			}
		} else {
			// check for downgrades
			twichinessFactor := int64(2)
			if atomic.LoadInt64(&ShardManagerKeeper.totalCapacity) >= atomic.LoadInt64(&ShardManagerKeeper.usedCapacity)*twichinessFactor {
				ShardManagerKeeper.mutex.Lock()
				log.Print("Downgrade requested")
				migrateOrNot := DowngradeShardManagerKeeper(ShardManagerKeeper.totalCapacity, twichinessFactor)
				ShardManagerKeeper.mutex.Unlock()

				if migrateOrNot {
					// log.Fatal("downsizing is getting triggered")
					log.Print("triggering resizing")
					go migrateKeys(&ShardManagerKeeper, &newShardManagerKeeper)
				}
			}
		}
	} else {
		log.Print("inserting in new table key: ", key)

		newShardManagerKeeper.mutex.RLock()

		setAtIndex(getKeyHash(key, &newShardManagerKeeper), &newShardManagerKeeper, request)

		newShardManagerKeeper.mutex.RUnlock()
	}
}

func handleSetWorker() {
	log.Print("Worker started")

	for {
		setReq := <-setQueue
		_setKey(setReq)
	}
}
