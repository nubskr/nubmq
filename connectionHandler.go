package main

import (
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func WriteStuffToConn(conn net.Conn, stuff string, encoder *gob.Encoder) {
	log.Print("XXXXXXXXX, tryna write to conn: ", stuff)
	err := encoder.Encode(stuff)

	if err != nil {
		panic("we done for sire")
		log.Print("failed to reply message:", err)
	} else {
		log.Print("YYYYYYYYYYY, write done sire")
	}
}

func GenerateUUID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic("failed to generate UUID")
	}

	// Set version (4) and variant (RFC 4122)
	b[6] = (b[6] & 0x0F) | 0x40 // UUID version 4
	b[8] = (b[8] & 0x3F) | 0x80 // Variant is 10xx xxxx

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func LogRequest(start, end time.Time) {
	logChannel <- [3]string{GenerateUUID(), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano)}
}

func handleConnection(conn net.Conn) {
	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	fmt.Println("Client connected")
	buffer := make([]byte, 1024)

	writeChanPrimary := make(chan string, 10)
	writeChanSecondary := make(chan string, 100)

	// Writer
	go func(conn net.Conn) {
		// we have a hierarchy, like, SET and GET replies get higher priority than say Event Notifications
		for {
			select {
			case val := <-writeChanPrimary:

				WriteStuffToConn(conn, val, encoder)
			case val := <-writeChanSecondary:
				WriteStuffToConn(conn, val, encoder)
			}
		}
	}(conn)

	for {

		var message string
		err := decoder.Decode(&message)
		if err != nil {
			fmt.Println("an error occurred while decoding message:", err)
			return
		}
		length := len(message)
		copy(buffer[:length], message)

		if err != nil {
			fmt.Println("an error occured while reading message:", err)
			return
		}

		data := string(buffer[:length])
		log.Print("Received from client: ", data)

		stringData := strings.Fields(data)

		/*
			SET key value EX time_in_seconds
			 0	 1    2	    3	  4
		*/
		// logging_reqId := GenerateUUID()
		// logging_curTime := time.Now()

		if stringData[0] == "SET" {
			curReq := SetRequest{
				key:       stringData[1],
				value:     stringData[2],
				canExpire: false,
				TTL:       time.Now().Unix(), // just let it be for now
				status:    make(chan struct{}),
				InTime:    time.Now(),
			}

			entry := Entry{
				key:   stringData[1],
				value: stringData[2],
			}

			EventQueue <- entry

			if len(stringData) == 5 {

				parsedTime, err := strconv.ParseInt(stringData[4], 10, 64)

				log.Print("========Parsed time is: ", parsedTime)
				if err != nil {
					log.Fatal("Error parsing time:", err)
				}

				// canExpire     bool
				// TTL           int64
				// isExpiryEvent bool
				entry.canExpire = true
				entry.TTL = parsedTime

				SetContainer.queue <- entry // handles TTL notifications

				curReq = SetRequest{
					key:       stringData[1],
					value:     stringData[2],
					canExpire: true,
					TTL:       parsedTime,
					status:    make(chan struct{}),
				}
			}

			// connContext := ConnContextMap[conn]
			connContextemp, ok := ConnContextMap.Load(conn)
			if !ok {
				panic("wtf bro")
			}
			connContext := connContextemp.(Connection)

			curBatchSize := connContext.batchSize

			curBatchSize -= 1
			connContext.batchSize = curBatchSize
			connContext.Batch = append(connContext.Batch, curReq)
			ConnContextMap.Store(conn, connContext)
			// update it here too
			if curBatchSize == 0 {
				// process stuff in Batch
				for _, req := range connContext.Batch {
					allowSets.Lock()
					SetWG.Add(1)
					allowSets.Unlock()
					setQueue <- req
					select {
					case <-req.status:
					case <-time.After(2 * time.Second):
						log.Fatal("BAD WORKER, SET REQUEST TIMED OUT FOR KEY: ", req.key)
					}

					writeChanPrimary <- "SET done"
				}
				connContext.Batch = make([]SetRequest, 0)
				curBatchSize = 1
				connContext.batchSize = curBatchSize

				ConnContextMap.Store(conn, connContext)
			}

			// allowSets.Lock()
			// SetWG.Add(1)
			// allowSets.Unlock()
			// // start logging for SETs here
			// setQueue <- curReq

			// select {
			// case <-curReq.status:
			// 	// LogRequest(logging_reqId, logging_curTime, time.Now())
			// 	// end logging for SETs here
			// case <-time.After(2 * time.Second): // Timeout in case of delay
			// 	log.Fatal("BAD WORKER, SET REQUEST TIMED OUT FOR KEY: ", curReq.key)
			// }

			// writeChanPrimary <- "SET done"

		} else if stringData[0] == "GET" {
			res, exists := _getKey(stringData[1])

			if exists {
				writeChanPrimary <- res
			} else {
				log.Fatal("tf just happened here")
			}
		} else if stringData[0] == "SUBSCRIBE" {
			key := stringData[1]

			SubscribersMutex.Lock()
			Subscribers[key] = append(Subscribers[key], &writeChanSecondary)
			SubscribersMutex.Unlock()
			writeChanSecondary <- "SUBSCRIBED TO CHANNEL"
		} else if stringData[0] == "BATCH" {
			if len(stringData) != 2 {
				writeChanPrimary <- "ERROR: BATCH command requires exactly one argument"
				continue
			}
			batchSize, err := strconv.Atoi(stringData[1])
			if err != nil {
				writeChanPrimary <- "ERROR: BATCH size must be a valid integer"
				continue
			}
			// ConnContextMap.Store(conn)
			// fetch it, change its value, then store it again
			connContextemp, ok := ConnContextMap.Load(conn)
			if !ok {
				panic("wtf bro")
			}
			connContext := connContextemp.(Connection)
			connContext.batchSize = uint32(batchSize)

			ConnContextMap.Store(conn, connContext)
			writeChanPrimary <- "OK"
		} else {
			writeChanPrimary <- "ERROR: Unknown command"
		}
	}
}
