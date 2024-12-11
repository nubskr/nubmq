package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"unsafe"
)

/*
1,2,4,6,8,16,32,64

Changes:

whenever we try to set something, we lazily traverse the ShardManagerKeeper for now, if we find the index in there,cool

if not, we initiate the addition of the next ShardManager with the current size and just wait until we have it

*/

func _setKey(key string, value string) {
	idx := int32(696969696)

	if value, ok := keyManager.Keys.Load(key); ok {
		if intValue, ok := value.(int32); ok {
			idx = int32(intValue)
		} else {
			fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO set-x-x-x-x-x-x-x-x-x-x-xx-x-x-x-x-x-x--x", value, "-->")
			os.Exit(1)
		}
	} else {
		val := atomic.AddInt32(&nextIdx, 1)
		keyManager.Keys.Store(key, val)
		idx = val
	}

	if idx == 696969696 {
		fmt.Println("trying to set non existing shit")
		os.Exit(1)
	}

	/*
		we need a few things:
		- SMkeeper index
		once we find the index, there will me multiple SMs in there, then we need to find the SM index
		from there we need to find the index of shard where we have that shit, from there, we need to find
		the index of the value inside that darn shard, how to do this all super duper fucking fast
	*/

	shardNumber := idx / ShardSize
	localShardIndex := idx % ShardSize

	fmt.Println("setting key", key, "at", idx, "at shard number", shardNumber, "at local index", localShardIndex)

	newVal := getNewValueData(value)
	fmt.Println("trying to acquire lock to set key")

	ShardManagerKeeper.mutex.Lock()

	fmt.Println("lock acquired to set key")
	// TODO: fix the below shit, it should not be this way
	// fmt.Println("set worker locked acquired")
	expectedCapacity := getEstimatedCapacityFromShardNumber(int(shardNumber))

	if expectedCapacity > int64(ShardManagerKeeper.capacity) {
		// do soemthing about it, lmao
		fmt.Println("++++++++++++++---[Need to Upgrate the SMKeeper to accomodate]---++++++++++++++")

		ShardManagerKeeper.mutex.Unlock()

		UpgradeShardManagerKeeper(shardNumber)

		fmt.Println(ShardManagerKeeper.capacity, expectedCapacity)
		for atomic.LoadInt32(&ShardManagerKeeper.capacity) < int32(expectedCapacity) {
			// wait it out
			//fmt.Println("staring into your soul")
		}

		ShardManagerKeeper.mutex.Lock()
	}

	SMidx := getShardManagerKeeperIndex(int(shardNumber))

	if SMidx == -1 {
		fmt.Println("we fucked up in resizing sire")
		os.Exit(1)
	}

	shard := ShardManagerKeeper.ShardManagers[SMidx].Shards[shardNumber]

	ShardManagerKeeper.mutex.Unlock()

	fmt.Println("set worker locked released")

	// value is a darn string
	atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.data[localShardIndex])), unsafe.Pointer(newVal))
}
