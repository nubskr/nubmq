package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var numConnections = 100 // Number of concurrent connections
var numKeys = 10000000   // Total number of read and write requests each
var mxVals = 100000      // limit the number of unique keys or the system will just start evicting them

func whatever(shit string) string {
	if shit == "NaN" {
		fmt.Println("your server is sending bullshit, check it dumbass")
		// os.Exit(1)
	}

	return shit
}

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
		return whatever(response), nil
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
		return whatever(response), nil
	}
}

// func writeDurationsToCSV(filename string, durations []time.Duration) error {
// 	file, err := os.Create(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()

// 	// Write header
// 	writer.Write([]string{"Rank", "Duration_ms"})

// 	// Write data
// 	for i, duration := range durations {
// 		record := []string{
// 			fmt.Sprintf("%d", i+1),
// 			fmt.Sprintf("%.3f", duration.Seconds()*1000), // Convert to milliseconds
// 		}
// 		if err := writer.Write(record); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func writeDurationsToCSV(filename string, records []struct {
	Timestamp int64
	Duration  time.Duration
}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Rank", "Timestamp_ms", "Duration_ms"})

	// Write data
	for i, record := range records {
		durationInMs := float64(record.Duration.Nanoseconds()) / 1e6 // Convert to milliseconds
		row := []string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%d", record.Timestamp), // Unix timestamp in milliseconds
			fmt.Sprintf("%.3f", durationInMs),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// func writeDurationsToCSV(filename string, durations []time.Duration) error {
// 	file, err := os.Create(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()

// 	// Write header
// 	writer.Write([]string{"Rank", "Duration_ms"})

// 	// Write data

// 	for i, duration := range durations {
// 		durationInMs := float64(duration.Nanoseconds()) / 1e6 // Convert nanoseconds to milliseconds
// 		record := []string{
// 			fmt.Sprintf("%d", i+1),
// 			fmt.Sprintf("%.3f", durationInMs),
// 		}
// 		if err := writer.Write(record); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func writeDurationsWithTimestampsToCSV(filename string, durations []time.Duration, timestamps []int64) error {
	if len(durations) != len(timestamps) {
		return fmt.Errorf("mismatched lengths: durations=%d, timestamps=%d", len(durations), len(timestamps))
	}

	// Open file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Rank", "Timestamp_ms", "Duration_ms"})

	// Write data
	for i := range durations {
		durationInMs := float64(durations[i].Nanoseconds()) / 1e6
		record := []string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%d", timestamps[i]),
			fmt.Sprintf("%.3f", durationInMs),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func Test_gogo(t *testing.T) {
	// Configuration
	serverAddress := "localhost:8080" // Replace with your server's address and port

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
		keys[i] = fmt.Sprintf("key%d", i%mxVals)
		values[i] = fmt.Sprintf("value%d", i%mxVals)
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
		// time.Sleep(1 * time.Second)
		// time.Sleep(2 * time.Second)
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
					log.Fatalf("bad")
					// t.FailNow()
				}
				return
			}
			// Note: Do not defer conn.Close() to keep connections open
			reader := bufio.NewReader(conn)

			// Handle assigned keys
			startKey := connIdx * keysPerConn
			endKey := startKey + keysPerConn

			// Phase 1: SET operations
			// for i := startKey; i < endKey; i++ {
			// 	key := keys[i]
			// 	value := values[i]
			// 	setCommand := fmt.Sprintf("SET %s %s", key, value)

			// 	// Record start time
			// 	startTime := time.Now()

			// 	// Send SET command
			// 	_, err := fmt.Fprintf(conn, "%s\n", setCommand)
			// 	if err != nil {
			// 		fmt.Printf("[SET][Connection %d] Error sending SET command for '%s': %v\n", connIdx, key, err)
			// 		setResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
			// 		continue
			// 	}

			// 	// Wait for SET response
			// 	response, err := readValidResponse(reader)
			// 	if err != nil {
			// 		fmt.Printf("[SET][Connection %d] Error reading SET response for '%s': %v\n", connIdx, key, err)
			// 		setResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
			// 		continue
			// 	}
			// 	duration := time.Since(startTime)

			// 	// Store SET response and duration
			// 	setResponses[connIdx][i-startKey] = fmt.Sprintf("Response: %s (Time: %v)", response, duration)
			// 	// time.Sleep(100 * time.Microsecond)
			// }
			for i := startKey; i < endKey; i++ {
				key := keys[i]
				value := values[i]
				setCommand := fmt.Sprintf("SET %s %s", key, value)

				// Record timestamp at request capture (in milliseconds)
				timestamp := time.Now().UnixMilli()

				// Record start time
				startTime := time.Now()

				// Send SET command
				_, err := fmt.Fprintf(conn, "%s\n", setCommand)
				if err != nil {
					fmt.Printf("[SET][Connection %d] Error sending SET command for '%s' at %d: %v\n", connIdx, key, timestamp, err)
					setResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
					continue
				}

				// Wait for SET response
				response, err := readValidResponse(reader)
				if err != nil {
					fmt.Printf("[SET][Connection %d] Error reading SET response for '%s' at %d: %v\n", connIdx, key, timestamp, err)
					setResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
					continue
				}
				duration := time.Since(startTime)

				// Store SET response, duration, and timestamp
				setResponses[connIdx][i-startKey] = fmt.Sprintf("Response: %s (Time: %v) [Timestamp: %d]", response, duration, timestamp)
			}

			// time.Sleep(1 * time.Second) // this should be good enough for the most part
			// wg.Wait()

			// Phase 2: GET operations
			// for i := startKey; i < endKey; i++ {
			// 	key := keys[i]
			// 	getCommand := fmt.Sprintf("GET %s", key)

			// 	// Record start time
			// 	startTime := time.Now()

			// 	// Send GET command
			// 	_, err := fmt.Fprintf(conn, "%s\n", getCommand)
			// 	if err != nil {
			// 		fmt.Printf("[GET][Connection %d] Error sending GET command for '%s': %v\n", connIdx, key, err)
			// 		getResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
			// 		continue
			// 	}

			// 	// Wait for GET response
			// 	response, err := _readValidResponse(reader)
			// 	if err != nil {
			// 		fmt.Printf("[GET][Connection %d] Error reading GET response for '%s': %v\n", connIdx, key, err)
			// 		getResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
			// 		continue
			// 	}
			// 	duration := time.Since(startTime)

			// 	// Store GET response and duration
			// 	getResponses[connIdx][i-startKey] = fmt.Sprintf("Response: %s (Time: %v)", response, duration)
			// 	// time.Sleep(1 * time.Millisecond)
			// 	// time.Sleep(200 * time.Microsecond)
			// }
			for i := startKey; i < endKey; i++ {
				key := keys[i]
				getCommand := fmt.Sprintf("GET %s", key)

				// Record timestamp at request capture (in milliseconds)
				timestamp := time.Now().UnixMilli()

				// Record start time
				startTime := time.Now()

				// Send GET command
				_, err := fmt.Fprintf(conn, "%s\n", getCommand)
				if err != nil {
					fmt.Printf("[GET][Connection %d] Error sending GET command for '%s' at %d: %v\n", connIdx, key, timestamp, err)
					getResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
					continue
				}

				// Wait for GET response
				response, err := _readValidResponse(reader)
				if err != nil {
					fmt.Printf("[GET][Connection %d] Error reading GET response for '%s' at %d: %v\n", connIdx, key, timestamp, err)
					getResponses[connIdx][i-startKey] = fmt.Sprintf("Error: %v", err)
					continue
				}
				duration := time.Since(startTime)

				// Store GET response, duration, and timestamp
				getResponses[connIdx][i-startKey] = fmt.Sprintf("Response: %s (Time: %v) [Timestamp: %d]", response, duration, timestamp)
			}

			// Connection remains open here; do not close it
		}(connIdx)
		// time.Sleep(1 * time.Millisecond)
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

	// // Calculate average SET and GET response times
	// var totalSetTime time.Duration
	// var totalGetTime time.Duration

	// for connIdx := 0; connIdx < numConnections; connIdx++ {
	// 	for keyIdx := 0; keyIdx < keysPerConn; keyIdx++ {
	// 		// Parse durations from SET responses
	// 		setResp := setResponses[connIdx][keyIdx]
	// 		if strings.Contains(setResp, "(Time: ") {
	// 			parts := strings.Split(setResp, "(Time: ")
	// 			if len(parts) == 2 {
	// 				timeStr := strings.TrimSuffix(parts[1], ")")
	// 				duration, err := time.ParseDuration(timeStr)
	// 				if err == nil {
	// 					totalSetTime += duration
	// 				}
	// 			}
	// 		}

	// 		// Parse durations from GET responses
	// 		getResp := getResponses[connIdx][keyIdx]
	// 		if strings.Contains(getResp, "(Time: ") {
	// 			parts := strings.Split(getResp, "(Time: ")
	// 			if len(parts) == 2 {
	// 				timeStr := strings.TrimSuffix(parts[1], ")")
	// 				duration, err := time.ParseDuration(timeStr)
	// 				if err == nil {
	// 					totalGetTime += duration
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	// avgSetTime := totalSetTime / time.Duration(numKeys)
	// avgGetTime := totalGetTime / time.Duration(numKeys)

	// fmt.Printf("Average SET Response Time: %v\n", avgSetTime)
	// fmt.Printf("Average GET Response Time: %v\n", avgGetTime)

	var totalSetTime time.Duration
	var totalGetTime time.Duration

	var setDurations []time.Duration
	var getDurations []time.Duration
	var setTimestamps []int64
	var getTimestamps []int64

	for connIdx := 0; connIdx < numConnections; connIdx++ {
		for keyIdx := 0; keyIdx < keysPerConn; keyIdx++ {
			// Parse durations & timestamps from SET responses
			setResp := setResponses[connIdx][keyIdx]
			if strings.Contains(setResp, "(Time: ") && strings.Contains(setResp, "[Timestamp: ") {
				parts := strings.Split(setResp, "(Time: ")
				timePart := strings.Split(parts[1], ")")[0]
				duration, err := time.ParseDuration(timePart)
				if err == nil {
					totalSetTime += duration
					setDurations = append(setDurations, duration)
				}

				// Extract timestamp
				tsParts := strings.Split(setResp, "[Timestamp: ")
				if len(tsParts) > 1 {
					tsStr := strings.TrimSuffix(tsParts[1], "]")
					timestamp, err := strconv.ParseInt(tsStr, 10, 64)
					if err == nil {
						setTimestamps = append(setTimestamps, timestamp)
					}
				}
			}

			// Parse durations & timestamps from GET responses
			getResp := getResponses[connIdx][keyIdx]
			if strings.Contains(getResp, "(Time: ") && strings.Contains(getResp, "[Timestamp: ") {
				parts := strings.Split(getResp, "(Time: ")
				timePart := strings.Split(parts[1], ")")[0]
				duration, err := time.ParseDuration(timePart)
				if err == nil {
					totalGetTime += duration
					getDurations = append(getDurations, duration)
				}

				// Extract timestamp
				tsParts := strings.Split(getResp, "[Timestamp: ")
				if len(tsParts) > 1 {
					tsStr := strings.TrimSuffix(tsParts[1], "]")
					timestamp, err := strconv.ParseInt(tsStr, 10, 64)
					if err == nil {
						getTimestamps = append(getTimestamps, timestamp)
					}
				}
			}
		}
	}

	// Sort durations in ascending order (so we can get top N)
	sort.Slice(setDurations, func(i, j int) bool {
		return setDurations[i] < setDurations[j]
	})
	sort.Slice(getDurations, func(i, j int) bool {
		return getDurations[i] < getDurations[j]
	})

	topN := 1000000000
	if len(setDurations) < topN {
		topN = len(setDurations)
	}
	if len(getDurations) < topN {
		topN = len(getDurations)
	}

	// Extract the top N durations & timestamps for BOTH SET and GET
	topSetDurations := setDurations[len(setDurations)-topN:]
	topSetTimestamps := setTimestamps[len(setTimestamps)-topN:]
	topGetDurations := getDurations[len(getDurations)-topN:]
	topGetTimestamps := getTimestamps[len(getTimestamps)-topN:]

	// Write SET and GET durations to CSV with timestamps
	writeDurationsWithTimestampsToCSV("./analysing-stuff/top_set_durations.csv", topSetDurations, topSetTimestamps)
	writeDurationsWithTimestampsToCSV("./analysing-stuff/top_get_durations.csv", topGetDurations, topGetTimestamps)

	// Extract the top N durations & timestamps
	// topSetDurations := setDurations[len(setDurations)-topN:]
	// topGetDurations := getDurations[len(getDurations)-topN:]
	// topGetTimestamps := getTimestamps[len(getTimestamps)-topN:]

	// Write to CSV (without changing function signature)
	// writeDurationsWithTimestampsToCSV("./analysing-stuff/top_get_durations.csv", topGetDurations, topGetTimestamps)
	// Get the top 10 max SET durations

	// Get the top 10 max GET durations
	// topGetDurations := getDurations[len(getDurations)-topN:]
	avgSetTime := totalSetTime / time.Duration(numKeys)
	avgGetTime := totalGetTime / time.Duration(numKeys)
	avgThroughput := (totalGetTime + totalSetTime) / (time.Duration(numKeys) * 2)

	// avgThroughput
	fmt.Printf("Average Response Time: %v\n", avgThroughput)
	fmt.Printf("Average SET Response Time: %v\n", avgSetTime)
	fmt.Printf("Top %d Max SET Response Times:\n", len(topSetDurations))
	fmt.Printf("Average GET Response Time: %v\n", avgGetTime)
	fmt.Printf("Top %d Max GET Response Times:\n", len(topGetDurations))

}
