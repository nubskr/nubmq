package main

import (
	"math"
	"sync"
)

func getNewShard() *Shard {
	return &Shard{
		data: sync.Map{},
	}
}

func getNewShardManagerTemplate(sz int) *ShardManager {
	return &ShardManager{
		Shards: make([]*Shard, sz),
	}
}

// return a new ShardManager with `sz` Shards
func getNewShardManager(sz int) *ShardManager {
	// TODO: can make this faster by parallelising all getNewShard stuffs
	newSM := getNewShardManagerTemplate(sz)

	for i := 0; i < sz; i++ {
		newSM.Shards[i] = getNewShard()
	}

	return newSM
}

func getEstimatedCapacityFromShardNumber(shardNumber int) int64 {
	return int64(math.Pow(2, float64(shardNumber+1))) - 1
}

func getShardNumberAndIndexPair(rawidx int) (int, int) {
	low, high := 0, rawidx
	for low < high {
		mid := low + (high-low)/2
		if getEstimatedCapacityFromShardNumber(mid) > int64(rawidx) {
			high = mid
		} else {
			low = mid + 1
		}
	}

	localIdx := rawidx
	if low > 0 {
		localIdx = int(int64(rawidx) - getEstimatedCapacityFromShardNumber(low-1))
	}

	return low, localIdx
}
