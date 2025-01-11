package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var myMap sync.Map

func __setKey(key string, value interface{}) {
	myMap.Store(key, value)
}

func __getKey(key string) (interface{}, bool) {
	return myMap.Load(key)
}

func __handleConnection(conn net.Conn) {
	fmt.Println("Client connected")
	buffer := make([]byte, 1024)
	for {
		err := conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			// return 0, fmt.Errorf("failed to set read deadline: %w", err)
			log.Fatal("Failed to set read deadline")
		}

		length, err := conn.Read(buffer)

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// exit the goroutine, connection has been dead for a while
			// log.Fatal("go routine terminated")
			return
		}
		if err != nil {
			fmt.Println("An error occurred while reading message:", err)
			return
		}

		data := string(buffer[:length])
		stringData := strings.Fields(data)

		if stringData[0] == "SET" {
			__setKey(stringData[1], stringData[2])
			_, err := conn.Write([]byte("SET done\n"))
			if err != nil {
				log.Println("Failed to reply message:", err)
			}
		} else {
			output, exists := __getKey(stringData[1])
			if !exists {
				output = "Key not found"
			}
			_, err := conn.Write([]byte(fmt.Sprint(output, "\n")))
			if err != nil {
				log.Println("Failed to send message:", err)
			}
		}
	}
}

// func main() {
// 	fasttttt := true

// 	// fasttttt = false

// 	if fasttttt {
// 		runtime.GOMAXPROCS(runtime.NumCPU())
// 	}

// 	// 	fasttttt := true
// 	// 	// fasttttt = false

// 	// 	if fasttttt {
// 	// 		runtime.GOMAXPROCS(runtime.NumCPU())
// 	// 	}

// 	// ln, err := net.Listen("tcp", ":8080")
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// fmt.Println("Server listening on :8080")

// 	// for {
// 	// 	conn, err := ln.Accept()
// 	// 	if err != nil {
// 	// 		log.Println(err)
// 	// 		continue
// 	// 	}
// 	// 	go __handleConnection(conn)
// 	// }
// 	// 	// init for 2 now
// 	// 	// ShardManagerKeeper = *getNewShardManagerKeeper(2)
// 	// 	// newShardManagerKeeper = *getNewShardManagerKeeper(1)

// 	// 	// for i := 1; i <= MaxConcurrentClients; i++ {
// 	// 	// 	go handleSetWorker()
// 	// 	// }

// 	// 	//-----------
// 	// 	// Step 1: Create a kqueue instance
// 	kq, err := unix.Kqueue()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Failed to create kqueue: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer unix.Close(kq)

// 	// Step 2: Create a TCP server
// 	listener, err := net.Listen("tcp", ":8080")
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Failed to create listener: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer listener.Close()

// 	// Get the listener's file descriptor and set non-blocking mode
// 	listenerFd := socketFD(listener)
// 	if listenerFd == -1 {
// 		fmt.Fprintln(os.Stderr, "Failed to get listener file descriptor")
// 		os.Exit(1)
// 	}

// 	fmt.Println("Listening on :8080...")

// 	// Step 3: Register the listener's file descriptor with kqueue
// 	event := unix.Kevent_t{
// 		Ident:  uint64(listenerFd),
// 		Filter: unix.EVFILT_READ, // Trigger on readability
// 		Flags:  unix.EV_ADD | unix.EV_ENABLE,
// 	}
// 	if _, err := unix.Kevent(kq, []unix.Kevent_t{event}, nil, nil); err != nil {
// 		fmt.Fprintf(os.Stderr, "Failed to register listener with kqueue: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Track active connections
// 	connections := make(map[int]net.Conn)

// 	// Step 4: Event loop
// 	events := make([]unix.Kevent_t, 100)
// 	for {
// 		n, err := unix.Kevent(kq, nil, events, nil)
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "Kevent error: %v\n", err)
// 			os.Exit(1)
// 		}

// 		// Process each triggered event
// 		for i := 0; i < n; i++ {
// 			fd := int(events[i].Ident)

// 			if fd == listenerFd {
// 				// New incoming connection
// 				conn, err := acceptNonBlocking(listener)
// 				if err != nil {
// 					fmt.Fprintf(os.Stderr, "Accept error: %v\n", err)
// 					continue
// 				}

// 				fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())

// 				connFd := socketFD(conn)
// 				connections[connFd] = conn

// 				// Add new connection to kqueue
// 				connEvent := unix.Kevent_t{
// 					Ident:  uint64(connFd),
// 					Filter: unix.EVFILT_READ, // Monitor for readability
// 					Flags:  unix.EV_ADD | unix.EV_ENABLE,
// 				}
// 				if _, err := unix.Kevent(kq, []unix.Kevent_t{connEvent}, nil, nil); err != nil {
// 					fmt.Fprintf(os.Stderr, "Failed to register connection with kqueue: %v\n", err)
// 					conn.Close()
// 					delete(connections, connFd)
// 				}
// 			} else {
// 				// Handle readable connection
// 				conn := connections[fd]
// 				if conn == nil {
// 					continue
// 				}

// 				// data := make([]byte, 1024)
// 				// n, err := conn.Read(data)
// 				// if err != nil {
// 				// 	fmt.Printf("Error reading from connection %v: %v\n", conn.RemoteAddr(), err)
// 				// 	conn.Close()
// 				// 	delete(connections, fd)
// 				// 	continue
// 				// }

// 				// _p, err := conn.Write([]byte(fmt.Sprint("SET done\n")))

// 				// if err != nil {
// 				// 	log.Println("failed to reply message:", err, _p)
// 				// } else {
// 				// }
// 				// fmt.Printf("Received data from %v: %s\n", conn.RemoteAddr(), string(data[:n]))

// 				go __handleConnection(conn)
// 			}
// 		}
// 	}
// }
