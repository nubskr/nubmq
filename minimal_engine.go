package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
)

// init a thousand shards
// var shardManager = ShardManager{
// 	Shards: make([]*Shard, 1000),
// }

/*

ShardManagerKeeper
	ShardManager..1.2.3..
		Shard..1.2.3..
			ValueData

*/

// init an empty SMkeeper
var ShardManagerKeeper = ShardManagerKeeperTemp{
	ShardManagers: make([]*ShardManager, 0),
	capacity:      0,
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

	go nextShardManagerWatcher()
	// TODO: remove this shit

	// for i := 0; i < 1000; i++ {
	// 	ShardManagerKeeper.data[0].Shards[i] = getNewShard(ShardSize)
	// }

	fmt.Println("Server listening on :8080")

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
