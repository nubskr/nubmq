package main

func getAtIndex(idx int, key string, keeper *ShardManagerKeeperTemp) (string, bool) {
	SMidx, localIdx := getShardNumberAndIndexPair(idx)
	if SMidx >= len(keeper.ShardManagers) {
		return "NaN", false
	}
	target := keeper.ShardManagers[SMidx].Shards[localIdx]
	value, ok := target.data.Load(key)
	if ok {
		return value.(string), true
	} else {
		return "NaN", false
	}
}

func _getKey(key string) (string, bool) {
	// check the new table first then the old one
	// TODO: can we do something better ? having to check two tables to fulfil one request is slow

	if true {
		newShardManagerKeeper.mutex.RLock()

		ret, found := getAtIndex(getKeyHash(key, &newShardManagerKeeper), key, &newShardManagerKeeper)

		newShardManagerKeeper.mutex.RUnlock()

		if found {
			return ret, found
		}
	}

	ShardManagerKeeper.mutex.RLock()

	ret, found := getAtIndex(getKeyHash(key, &ShardManagerKeeper), key, &ShardManagerKeeper)

	ShardManagerKeeper.mutex.RUnlock()

	return ret, found
}
