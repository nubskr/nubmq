package main

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"golang.org/x/sys/unix"
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

// Helper to get file descriptor from net.Listener or net.Conn
func socketFD(v interface{}) int {
	switch t := v.(type) {
	case *net.TCPListener:
		file, err := t.File()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get file for listener: %v\n", err)
			return -1
		}
		// Set the file descriptor to non-blocking mode
		err = unix.SetNonblock(int(file.Fd()), true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set non-blocking mode: %v\n", err)
			file.Close()
			return -1
		}
		return int(file.Fd())
	case *net.TCPConn:
		file, err := t.File()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get file for connection: %v\n", err)
			return -1
		}
		// Set non-blocking mode for the connection
		err = unix.SetNonblock(int(file.Fd()), true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set non-blocking mode: %v\n", err)
			file.Close()
			return -1
		}
		return int(file.Fd())
	default:
		fmt.Fprintf(os.Stderr, "Unsupported type for socketFD: %T\n", v)
		return -1
	}
}

// Accept a new connection in non-blocking mode
func acceptNonBlocking(listener net.Listener) (net.Conn, error) {
	ln, ok := listener.(*net.TCPListener)
	if !ok {
		return nil, fmt.Errorf("listener is not a TCPListener")
	}

	return ln.Accept()
}

func main() {
	fasttttt := true

	// fasttttt = false

	if fasttttt {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// init for 2 now
	ShardManagerKeeper = *getNewShardManagerKeeper(2)
	newShardManagerKeeper = *getNewShardManagerKeeper(1)

	for i := 1; i <= MaxConcurrentClients; i++ {
		go handleSetWorker()
	}

	//-----------
	// Step 1: Create a kqueue instance
	kq, err := unix.Kqueue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create kqueue: %v\n", err)
		os.Exit(1)
	}
	defer unix.Close(kq)

	// Step 2: Create a TCP server
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create listener: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	// Get the listener's file descriptor and set non-blocking mode
	listenerFd := socketFD(listener)
	if listenerFd == -1 {
		fmt.Fprintln(os.Stderr, "Failed to get listener file descriptor")
		os.Exit(1)
	}

	fmt.Println("Listening on :8080...")

	// Step 3: Register the listener's file descriptor with kqueue
	event := unix.Kevent_t{
		Ident:  uint64(listenerFd),
		Filter: unix.EVFILT_READ, // Trigger on readability
		Flags:  unix.EV_ADD | unix.EV_ENABLE,
	}
	if _, err := unix.Kevent(kq, []unix.Kevent_t{event}, nil, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register listener with kqueue: %v\n", err)
		os.Exit(1)
	}

	// Track active connections
	connections := make(map[int]net.Conn)

	// Step 4: Event loop
	events := make([]unix.Kevent_t, 100)
	for {
		n, err := unix.Kevent(kq, nil, events, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Kevent error: %v\n", err)
			os.Exit(1)
		}

		// Process each triggered event
		for i := 0; i < n; i++ {
			fd := int(events[i].Ident)

			if fd == listenerFd {
				// New incoming connection
				conn, err := acceptNonBlocking(listener)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Accept error: %v\n", err)
					continue
				}

				// fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())

				connFd := socketFD(conn)
				connections[connFd] = conn

				// Add new connection to kqueue
				connEvent := unix.Kevent_t{
					Ident:  uint64(connFd),
					Filter: unix.EVFILT_READ, // Monitor for readability
					Flags:  unix.EV_ADD | unix.EV_ENABLE,
				}
				if _, err := unix.Kevent(kq, []unix.Kevent_t{connEvent}, nil, nil); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to register connection with kqueue: %v\n", err)
					conn.Close()
					delete(connections, connFd)
				}
			} else {
				// Handle readable connection
				conn := connections[fd]
				if conn == nil {
					continue
				}

				// data := make([]byte, 1024)
				// n, err := conn.Read(data)
				// if err != nil {
				// 	fmt.Printf("Error reading from connection %v: %v\n", conn.RemoteAddr(), err)
				// 	conn.Close()
				// 	delete(connections, fd)
				// 	continue
				// }

				// _p, err := conn.Write([]byte(fmt.Sprint("SET done\n")))

				// if err != nil {
				// 	log.Println("failed to reply message:", err, _p)
				// } else {
				// }
				// fmt.Printf("Received data from %v: %s\n", conn.RemoteAddr(), string(data[:n]))

				running, ok := activeConns.Load(conn.RemoteAddr())

				if !ok || running == false {
					activeConns.Store(conn.RemoteAddr(), true)
					go handleConnection(conn)
				} else {
				}
			}
		}
	}
	//-----------

	// for {
	// 	// Accept connection
	// 	conn, err := ln.Accept()
	// 	if err != nil {
	// 		log.Println(err)
	// 		continue
	// 	}

	// go handleConnection(conn)
	// }
}
