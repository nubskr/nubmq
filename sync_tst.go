package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Helper function to read the next valid response, ignoring lines starting with "GET"
func readValidResponse(reader *bufio.Reader) (string, error) {
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		response = strings.TrimSpace(response)
		if strings.HasPrefix(response, "GET") {
			// Unexpected echo of the GET command, ignore and continue reading
			continue
		}
		return response, nil
	}
}

func _readValidResponse(reader *bufio.Reader) (string, error) {
	for {
		response, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		response = strings.TrimSpace(response)
		if strings.HasPrefix(response, "GET") {
			// Unexpected echo of the GET command, ignore and continue reading
			continue
		}
        if strings.HasPrefix(response, "SET") {
			// Unexpected echo of the GET command, ignore and continue reading
			continue
		}
		return response, nil
	}
}

func main() {
	// Configuration
	serverAddress := "localhost:8080" // Replace with your server's address and port

	numConnections := 100             // Number of concurrent connections
	numKeys := 500      	  // Total number of unique keys

	// Validate that numKeys is divisible by numConnections for even distribution
	if numKeys%numConnections != 0 {
		fmt.Println("numKeys should be divisible by numConnections for even distribution.")
		return
	}

	// Calculate number of keys per connection
	keysPerConn := numKeys / numConnections

	// Generate all keys and values
	keys := make([]string, numKeys)
	values := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
		values[i] = fmt.Sprintf("value%d", i)
	}

	// Slice to collect SET and GET responses
	setResponses := make([][]string, numConnections)
	getResponses := make([][]string, numConnections)
	for i := 0; i < numConnections; i++ {
		setResponses[i] = make([]string, keysPerConn)
		getResponses[i] = make([]string, keysPerConn)
	}

	// WaitGroups to wait for all SET and GET operations
	var wg sync.WaitGroup
	wg.Add(numConnections)

	// Start all connections and perform SET and GET operations sequentially
	for connIdx := 0; connIdx < numConnections; connIdx++ {
		go func(connIdx int) {
			defer wg.Done()

			// Establish connection
			conn, err := net.Dial("tcp", serverAddress)
			if err != nil {
				fmt.Printf("[Connection %d] Error connecting to server: %v\n", connIdx, err)
				// Fill in error messages for all keys in this connection
				for k := 0; k < keysPerConn; k++ {
					setResponses[connIdx][k] = fmt.Sprintf("Error: %v", err)
					getResponses[connIdx][k] = fmt.Sprintf("Error: %v", err)
				}
				return
			}
			// Note: Do not defer conn.Close() to keep connections open
			reader := bufio.NewReader(conn)

			// Handle assigned keys
			startKey := connIdx * keysPerConn
			endKey := startKey + keysPerConn

			// Phase 1: SET operations
			for i := startKey; i < endKey; i++ {
				key := keys[i]
				value := values[i]
				setCommand := fmt.Sprintf("SET %s %s", key, value)

				// Record start time
				startTime := time.Now()

				// Send SET command
				_, err := fmt.Fprintf(conn, "%s\n", setCommand)
				if err != nil {
					fmt.Printf("[SET][Connection %d] Error sending SET command for '%s': %v\n", connIdx, key, err)
					setResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
					continue
				}

				// Wait for SET response
				response, err := readValidResponse(reader)
				if err != nil {
					fmt.Printf("[SET][Connection %d] Error reading SET response for '%s': %v\n", connIdx, key, err)
					setResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
					continue
				}
				duration := time.Since(startTime)

				// Store SET response and duration
				setResponses[connIdx][i-startKey] = fmt.Sprintf("Response: %s (Time: %v)", response, duration)
			}

			// Phase 2: GET operations
			for i := startKey; i < endKey; i++ {
				key := keys[i]
				getCommand := fmt.Sprintf("GET %s", key)

				// Record start time
				startTime := time.Now()

				// Send GET command
				_, err := fmt.Fprintf(conn, "%s\n", getCommand)
				if err != nil {
					fmt.Printf("[GET][Connection %d] Error sending GET command for '%s': %v\n", connIdx, key, err)
					getResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
					continue
				}

				// Wait for GET response
				response, err := _readValidResponse(reader)
				if err != nil {
					fmt.Printf("[GET][Connection %d] Error reading GET response for '%s': %v\n", connIdx, key, err)
					getResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
					continue
				}
				duration := time.Since(startTime)

				// Store GET response and duration
				getResponses[connIdx][i-startKey] = fmt.Sprintf("Response: %s (Time: %v)", response, duration)
			}

			// Connection remains open here; do not close it
		}(connIdx)
	}

	// Wait for all connections to complete their SET and GET operations
	wg.Wait()

	// Print all SET responses
	fmt.Println("=== SET Responses ===")
	for connIdx := 0; connIdx < numConnections; connIdx++ {
		for keyIdx := 0; keyIdx < keysPerConn; keyIdx++ {
			globalKeyIdx := connIdx*keysPerConn + keyIdx
			key := keys[globalKeyIdx]
			fmt.Printf("[SET][Connection %d] Key: '%s' - %s\n", connIdx, key, setResponses[connIdx][keyIdx])
		}
	}

	// Print all GET responses
	fmt.Println("\n=== GET Responses ===")
	for connIdx := 0; connIdx < numConnections; connIdx++ {
		for keyIdx := 0; keyIdx < keysPerConn; keyIdx++ {
			globalKeyIdx := connIdx*keysPerConn + keyIdx
			key := keys[globalKeyIdx]
			fmt.Printf("[GET][Connection %d] Key: '%s' - %s\n", connIdx, key, getResponses[connIdx][keyIdx])
		}
	}

	// Summary
	fmt.Println("\n--- Test Summary ---")
	fmt.Printf("Total Connections: %d\n", numConnections)
	fmt.Printf("Total SET Commands: %d\n", numKeys)
	fmt.Printf("Total GET Commands: %d\n", numKeys)

	// Calculate average SET and GET response times
	var totalSetTime time.Duration
	var totalGetTime time.Duration
	for connIdx := 0; connIdx < numConnections; connIdx++ {
		for keyIdx := 0; keyIdx < keysPerConn; keyIdx++ {
			// Parse durations from SET responses
			setResp := setResponses[connIdx][keyIdx]
			if strings.Contains(setResp, "(Time: ") {
				parts := strings.Split(setResp, "(Time: ")
				if len(parts) == 2 {
					timeStr := strings.TrimSuffix(parts[1], ")")
					duration, err := time.ParseDuration(timeStr)
					if err == nil {
						totalSetTime += duration
					}
				}
			}

			// Parse durations from GET responses
			getResp := getResponses[connIdx][keyIdx]
			if strings.Contains(getResp, "(Time: ") {
				parts := strings.Split(getResp, "(Time: ")
				if len(parts) == 2 {
					timeStr := strings.TrimSuffix(parts[1], ")")
					duration, err := time.ParseDuration(timeStr)
					if err == nil {
						totalGetTime += duration
					}
				}
			}
		}
	}

	avgSetTime := totalSetTime / time.Duration(numKeys)
	avgGetTime := totalGetTime / time.Duration(numKeys)

	fmt.Printf("Average SET Response Time: %v\n", avgSetTime)
	fmt.Printf("Average GET Response Time: %v\n", avgGetTime)

	// Note: Connections are kept open and will be closed when the program exits
}
