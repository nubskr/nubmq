package main

import "math"

func getNewShard(sz int32) *Shard {
	return &Shard{
		data: make([]*ValueData, sz, ShardSize),
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
	newSM := getNewShardManagerTemplate(sz)

	for i := 0; i < sz; i++ {
		newSM.Shards[i] = getNewShard(ShardSize)
	}

	return newSM
}

func getShardManagerKeeperIndex(pos int) int {
	// the below thing assuming that there is already a lock placed on SMKeeper
	sz := len(ShardManagerKeeper.ShardManagers)
	travTillNow := 0
	for i := 0; i < sz; i++ {

		// do we need a lock on the below ?
		curSMsize := len(ShardManagerKeeper.ShardManagers[i].Shards)

		travTillNow += curSMsize

		if travTillNow > pos {
			return i
		}
	}
	return -1
}

func getEstimatedCapacityFromShardNumber(shardNumber int) int64 {
	return int64(math.Pow(2, float64(shardNumber+1))) - 1
}

// TODO: make it faster with binary search
func getShardNumberAndIndexPair(rawValue int64) (int, int) {
	// (ShardManagerNumber ,)
	i := 0
	for getEstimatedCapacityFromShardNumber(i) < int64(rawValue+1) {
		i++
	}
	return i, i - 1
}
