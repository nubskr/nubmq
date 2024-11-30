ShardManagerController {
	ShardManager[1]
	ShardManager[2]
	ShardManager[4]
	ShardManager[8]
	ShardManager[16]
	...
}

instead of doubling the shardmanager size every time it gets full,

we will make a new instance of shardmanager which has double the number of shards as the last shardmanager instance

this way, there will not be as many shardManager instances, is this a good idea ? lets see for ourselves