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

// I know abstractions bad, but let it be for now, already too much complexity in life
type Shard struct {
	data sync.Map
	size int32
}

type ShardManager struct {
	Shards []*Shard // pointers to shards
	mutex  sync.RWMutex
}

var ShardSize int32 = 1

// INFO: this is the outermost layer!!
type ShardManagerKeeperTemp struct {
	ShardManagers []*ShardManager
	mutex         sync.RWMutex
	totalCapacity int64
	usedCapacity  int64
	isResizing    int32
}
