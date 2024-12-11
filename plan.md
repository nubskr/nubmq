ShardManagerController {
	ShardManager[1]
	ShardManager[2]
	ShardManager[4]
	ShardManager[8]
	ShardManager[16]
	ShardManager[x] -> Shard1, Shard2,..., Shardx
	
	...
}

Shard.data -> [shit1,shit2,shit3...shardSize]


instead of doubling the shardmanager size every time it gets full,

we will make a new instance of shardmanager which has double the number of shards as the last shardmanager instance

this way, there will not be as many shardManager instances, is this a good idea ? lets see for ourselves

each ShardManager will also have their own hash function, also, can we darn replace Shard.data with a map ? would be fucking easier, oh, wait a fucking second, if we replace it with a darn map, how the hell are we gonna keep track of the capacity of the Shard ? we would need to do it manually, bro....


Is there any way we can also umm, auto scale down when a lot of umm, Keys expire ?

How to handle key evictions, have a present constant Size and if number of keys gets bigger than that,
start evicting shit ?

TLDR:

man, there are a darn lot of changes, ughhhhhhhhhhhhhh

forget whatever happens to reads and writes during the resizing mode as of now

- we need something like a global hash function
	- it should be a function which we can call like an goroutine every time some request comes
	- what should the hash function even be (in terms of implementation)
	- How the fuck do we upgrade it reliably, we can't just do that shit atomically, SETs and GETs would be left hanging
	-  

- How does SETs and GETs change here, let's assume the hash function as a black box for now
	- 

ughhhh, what is the smallest shit I can do to get it started

[+] Get rid of key manager
[+] Change SETs and GETs to not use the darn key manager and use KeyHash instead
[+] 


## Non Negotiables:

* to resize, need to change engine and hash function at the exact same time, can't allow any delay in between


