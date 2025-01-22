package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
)

/*

ShardManagerKeeper
	ShardManager..1.2.3..
		Shard..1.2.3..
			ValueData

*/

// init an empty SMkeeper
var ShardManagerKeeper = ShardManagerKeeperTemp{
	ShardManagers: make([]*ShardManager, 0),
	totalCapacity: 0,
	usedCapacity:  0,
	isResizing:    0,
}

var newShardManagerKeeper = ShardManagerKeeperTemp{
	ShardManagers: make([]*ShardManager, 0),
	totalCapacity: 0,
	usedCapacity:  0,
	isResizing:    0,
}

func main() {
	log.SetOutput(io.Discard)

	runtime.GOMAXPROCS(runtime.NumCPU())

	Subscribers = make(map[string][]*chan string)
	for i := 1; i <= MaxConcurrentCoreWorkers; i++ {
		go handleSetWorker()
	}

	go evenNotificationHandler()

	ln, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server listening on :8080")

	ShardManagerKeeper = *getNewShardManagerKeeper(1)
	newShardManagerKeeper = *getNewShardManagerKeeper(1)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}
