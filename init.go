package main

import (
	"sync"
)

var MaxConcurrentClients int = 25

var setQueue chan SetRequest = make(chan SetRequest, MaxConcurrentClients)
var SetWG sync.WaitGroup
var allowSets sync.Mutex

// TODO: sync map too hardcode and abstracted for my taste
type Shard struct {
	data sync.Map
}

type ShardManager struct {
	Shards []*Shard
	mutex  sync.RWMutex
}

type ShardManagerKeeperTemp struct {
	ShardManagers []*ShardManager
	mutex         sync.RWMutex
	totalCapacity int64
	usedCapacity  int64
	isResizing    int32
}

type SetRequest struct {
	key       string
	value     string
	canExpire bool
	TTL       int64
	status    chan struct{}
}

type Entry struct {
	value     string
	canExpire bool
	TTL       int64
}
