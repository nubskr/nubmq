package main

import (
	"fmt"
	"os"
)

func _getKey(key string) (string, bool) {
	idx := int32(696969696) // TODO: remove this shit

	if value, ok := keyManager.Keys.Load(key); ok {
		if intValue, ok := value.(int32); ok {
			idx = int32(intValue)
		} else {

			fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO get-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x", value, "-->")
			os.Exit(1)
		}
	} else {
		return "NaN", false
	}

	if idx == 696969696 {
		fmt.Println("trying to get non existing shit")
		os.Exit(1)
	}

	shardNumber := idx / ShardSize

	shardManager.mutex.Lock()

	if shardNumber >= int32(len(shardManager.Shards)) {
		shardManager.mutex.Unlock()
		return "", false
	}

	shard := shardManager.Shards[shardNumber]
	localShardIndex := idx % ShardSize

	if localShardIndex < int32(len(shard.data)) {
		shardManager.mutex.Unlock()
		return (shard.data[localShardIndex]).data, true
	}

	shardManager.mutex.Unlock()
	return "NaN", false
}
