package main

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

func getNewShardManager(sz int32) *ShardManager {
	return &ShardManager{
		Shards: make([]*Shard, sz),
	}
}
