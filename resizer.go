package main

import (
	"fmt"
	"sync"
)

type nextShardManagerTemplate struct {
	data  chan *ShardManager
	mutex sync.RWMutex
}

var nextShardManager = nextShardManagerTemplate{
	data: make(chan *ShardManager),
}

// var noOfShards int = 1000 // these are the number of shards in each ShardManager

func nextShardManagerWatcher() {
	curSz := 1

	for {
		// fmt.Println("<=======Next creation worker triggered for size=======>", curSz)

		// nextShardManager.mutex.Lock()

		now := getNewShardManager(curSz)

		nextShardManager.data <- now
		fmt.Println("<=======NextSM of size=======>", curSz, "X-digested-X")

		curSz *= 2

		// nextShardManager.mutex.Unlock()
	}
}

// Adds one more layer of SM to SMkeeper
func UpgradeShardManagerKeeper(newSz int32) {
	// get a lock and check if we even need to resize at all,

	// how to resize
	fmt.Println("SMKeeper upgrade triggered")

	fmt.Println("trying to acquire lock")

	ShardManagerKeeper.mutex.Lock()

	fmt.Println("lock acquired")

	if int32(ShardManagerKeeper.totalCapacity) > newSz {
		fmt.Println("trash, no need to upgrade, already big enough")

		ShardManagerKeeper.mutex.Unlock()
		return
	}

	fmt.Println("acquiring nextSM lock")

	// nextShardManager.mutex.Lock()

	fmt.Println("next smkeeper lock acquired")

	// append this SM to SMkeeper
	toBeAddedSM := <-nextShardManager.data

	fmt.Println("We good ?")
	ShardManagerKeeper.ShardManagers = append(ShardManagerKeeper.ShardManagers, toBeAddedSM)
	ShardManagerKeeper.totalCapacity += int64(len(toBeAddedSM.Shards)) // do we need atomic here ? I don't think so, since this thing is only being updated one at a time due to locks

	fmt.Println("capacity upgraded to", ShardManagerKeeper.totalCapacity)

	if int32(ShardManagerKeeper.totalCapacity) <= newSz {
		go UpgradeShardManagerKeeper(newSz)
	}

	// nextShardManager.mutex.Unlock()

	ShardManagerKeeper.mutex.Unlock()

}
