package main

import (
	"fmt"
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
	ShardManagers:   make([]*ShardManager, 0),
	totalCapacity:   0,
	usedCapacity:    0,
	isResizing:      0,
	pendingRequests: 0,
}

var newShardManagerKeeper = ShardManagerKeeperTemp{
	ShardManagers:   make([]*ShardManager, 0),
	totalCapacity:   0,
	usedCapacity:    0,
	isResizing:      0,
	pendingRequests: 0,
}

func main() {
	fasttttt := true

	// fasttttt = false

	if fasttttt {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	ln, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err)
	}

	// go nextShardManagerWatcher()

	fmt.Println("Server listening on :8080")

	// init for 2 now
	ShardManagerKeeper = *getNewShardManagerKeeper(2)
	newShardManagerKeeper = *getNewShardManagerKeeper(1)

	for i := 1; i <= MaxConcurrentClients; i++ {
		go handleSetWorker()
	}

	for {
		// Accept connection
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}
