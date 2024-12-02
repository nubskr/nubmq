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
