package main

import "sync"

type nextShardManagerTemplate struct {
	SM    chan *[]ShardManager
	mutex sync.RWMutex
}

var nextShardManager = nextShardManagerTemplate{
	SM: make(chan *[]ShardManager),
}

var noOfShards int32 = 1000

func nextShardManagerWatcher() {
	curSz := 1

	for {
		nextShardManager.mutex.Lock()

		now := make([]ShardManager, curSz)
		for i := 0; i < curSz; i++ {
			now[i] = *getNewShardManager(noOfShards)
		}

		nextShardManager.SM <- &now
		curSz *= 2

		nextShardManager.mutex.Unlock()
	}
}

func UpgradeShardManagerKeeper(curSz int32) {
	// get a lock and check if we even need to resize at all,

	// how to resize

}
