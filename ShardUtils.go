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

func getNewShardManager() *ShardManager {
	return &shardManager{
		// TODO: complete this shit
		// data: make()
	}
}
