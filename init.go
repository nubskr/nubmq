package main

import "sync"

type Message struct {
	data      string
	timestamp int64
}

type ValueData struct {
	data  string
	mutex sync.RWMutex
}

type Shard struct {
	data []*ValueData
	size int32
}

type KeyManager struct {
	Keys sync.Map
	// mutex sync.Mutex // for adding new keys
}

type ShardManager struct {
	Shards []*Shard // pointers to shards
	mutex  sync.RWMutex
}

var ShardSize int32 = 500

// Global variables
var keyManager = KeyManager{
	Keys: sync.Map{},
}

// INFO: this is the outermost layer!!
type ShardManagerKeeper struct {
	data []*ShardManager
}

// init a thousand shards
var shardManager = ShardManager{
	Shards: make([]*Shard, 1000),
}

var nextIdx int32 = -1
var curSetCnt int32 = 0
