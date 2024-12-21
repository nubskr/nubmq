package main

import (
	"fmt"
	"math"
	"sync"
)

func getNewShard() *Shard {
	return &Shard{
		// data: make([]*ValueData, sz, ShardSize),
		data: sync.Map{},
	}
}

func getNewValueData(value string) *ValueData {
	return &ValueData{
		data: value,
	}
}

func getNewShardManagerTemplate(sz int) *ShardManager {
	return &ShardManager{
		Shards: make([]*Shard, sz),
	}
}

// return a new ShardManager with `sz` Shards
func getNewShardManager(sz int) *ShardManager {
	// TODO: we can make this faster by parallelising all getNewShard stuffs
	newSM := getNewShardManagerTemplate(sz)

	for i := 0; i < sz; i++ {
		newSM.Shards[i] = getNewShard()
	}

	return newSM
}

func getEstimatedCapacityFromShardNumber(shardNumber int) int64 {
	return int64(math.Pow(2, float64(shardNumber+1))) - 1
}

// TODO: make it faster with binary search
// BUG: something wrong with localIdx calculation shit, fix it idiot
func getShardNumberAndIndexPair(rawidx int) (int, int) {
	// (ShardManagerNumber ,)
	fmt.Println("lmao, look at me", rawidx)
	// rawidx -= 1
	i := 0
	for getEstimatedCapacityFromShardNumber(i) <= int64(rawidx) {
		i++
	}
	// now that we have i, find the index it is in
	var localIdx int = int(rawidx)

	if i > 0 {
		localIdx = int(int64(rawidx) - getEstimatedCapacityFromShardNumber(i-1))
	}

	return i, localIdx
}
