package main

func getAtIndex(idx int, key string) (string, bool) {
	SMidx, localIdx := getShardNumberAndIndexPair(idx)
	target := ShardManagerKeeper.ShardManagers[SMidx].Shards[localIdx]
	// target.data.Store(key, val)
	value, ok := target.data.Load(key)
	if ok {
		return value.(string), true
	} else {
		return "NaN", false
	}
}

func _getKey(key string) (string, bool) {
	// ShardManagerKeeper.mutex.RLock()

	// ret,found := getAtIndex(getKeyHash(key), key)

	// ShardManagerKeeper.mutex.RUnlock()

	return getAtIndex(getKeyHash(key), key)
}
