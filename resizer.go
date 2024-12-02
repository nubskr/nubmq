package main

import "sync"

type nextShardManagerTemplate struct {
	data  chan *ShardManager
	mutex sync.RWMutex
}

var nextShardManager = nextShardManagerTemplate{
	data: make(chan *ShardManager),
}

var noOfShards int = 1000 // these are the number of shards in each ShardManager

func nextShardManagerWatcher() {
	curSz := 1

	for {
		nextShardManager.mutex.Lock()

		now := getNewShardManager(curSz)

		nextShardManager.data <- now
		curSz *= 2

		nextShardManager.mutex.Unlock()
	}
}

// Adds one more layer of SM to SMkeeper
func UpgradeShardManagerKeeper(newSz int32) {
	// get a lock and check if we even need to resize at all,

	// how to resize
	ShardManagerKeeper.mutex.Lock()

	if ShardManagerKeeper.capacity > newSz {
		ShardManagerKeeper.mutex.Unlock()
		return
	}

	nextShardManager.mutex.Lock()
	// append this SM to SMkeeper
	toBeAddedSM := <-nextShardManager.data

	ShardManagerKeeper.ShardManagers = append(ShardManagerKeeper.ShardManagers, toBeAddedSM)
	ShardManagerKeeper.capacity += int32(len(toBeAddedSM.Shards)) // do we need atomic here ? I don't think so, since this thing is only being updated one at a time due to locks

	if ShardManagerKeeper.capacity <= newSz {
		go UpgradeShardManagerKeeper(newSz)
	}

	nextShardManager.mutex.Unlock()

	ShardManagerKeeper.mutex.Unlock()

}
