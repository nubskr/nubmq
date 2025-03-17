package main

import (
	"sync"
)

var MaxConcurrentCoreWorkers int = 200

// var EVENT_NOTIFICATION_BUFFER int = 1000 // WARN: magic number lmao, need it to avoid blocking connection reads in the core engine

var setQueue chan SetRequest = make(chan SetRequest, MaxConcurrentCoreWorkers)
var SetWG sync.WaitGroup
var allowSets sync.Mutex

var Subscribers map[string][]*chan string // key -> SubscriberWriteSecondaryChannels
var SubscribersMutex sync.Mutex

// var EventQueue chan Entry = make(chan Entry, EVENT_NOTIFICATION_BUFFER)

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

// net.Conn(*net.TCPConn) *{conn: net.conn {fd: *(*net.netFD)(0x14000142100)}}
// type Connection struct {
// }

/*
map conn to the WriteChannel

map topic to Subs slice for pub sub
*/

type Entry struct {
	key           string
	value         string
	canExpire     bool
	TTL           int64
	isExpiryEvent bool
}
