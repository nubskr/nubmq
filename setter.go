package main

func setAtIndex(idx int, key string, val string) {
	SMidx, localIdx := getShardNumberAndIndexPair(idx)
	target := ShardManagerKeeper.ShardManagers[SMidx].Shards[localIdx]
	target.data.Store(key, val)
}

func _setKey(key string, value string) {
	/*
		get the key hash and ShardNumber from there

		from that ShardNumber update that value for that Shard

		TODO: check for capacity exceeding desired upper bound and then trigger resizing state
	*/
	ShardManagerKeeper.mutex.Lock()

	setAtIndex(getKeyHash(key), key, value)

	ShardManagerKeeper.mutex.Unlock()
}
