package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
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

var logChannel = make(chan [3]string, 1000000) // Buffer for async logging

func init() {
	go func() {
		file, _ := os.OpenFile("LOG.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()

		for entry := range logChannel {
			writer.Write(entry[:])
			writer.Flush()
		}
	}()
}

func main() {
	log.SetOutput(io.Discard)
	runtime.GOMAXPROCS(runtime.NumCPU())

	// defer func() {
	// 	f, err := os.Create("profile.prof")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer f.Close()
	// 	pprof.WriteHeapProfile(f) // Change this to StartCPUProfile if you want CPU usage instead
	// }()
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	Subscribers = make(map[string][]*chan string)
	ConnContextMap = sync.Map{}

	// go actuallyDoStuff()
	for i := 1; i <= MaxConcurrentCoreWorkers; i++ {
		go handleSetWorker()
	}

	go HandleKeyTTLInsertion(&SetContainer, &UpdateChan)
	go HandleKeyTTLEviction(&SetContainer, &UpdateChan, &EventQueue)
	go eventNotificationHandler()

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
		newConnection := Connection{
			conn:      conn,
			batchSize: 1,
			Batch:     make([]SetRequest, 0),
		}
		// this is not a critical async area right ?
		// ConnContextMap[conn] = &newConnection
		// sync.Map
		ConnContextMap.Store(conn, newConnection)
		go handleConnection(conn)
	}
}
