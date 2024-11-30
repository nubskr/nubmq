package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
)

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

	// TODO: remove this shit
	for i := 0; i < 1000; i++ {
		shardManager.Shards[i] = getNewShard(ShardSize)
	}

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
