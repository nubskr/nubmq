package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	numGoroutines = 100 // Number of concurrent connections
	uuidLength    = 36  // Length of UUID-like string
)

var totalActiveTime int64    // in nanoseconds (atomic)
var totalKeysProcessed int64 // atomic

func generateUUID() string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, uuidLength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomKeyValue() string {
	return fmt.Sprintf("SET %s %s\n", generateUUID(), generateUUID())
}

// spamConnection sends 10^4 keys per connection in batches of 100 and measures the active processing time.
func spamConnection() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	rand.Seed(time.Now().UnixNano())

	// Use gob encoder/decoder for communication
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	n := 100000       // Total keys per connection
	batchSize := 1000 // Keys per batch
	batches := n / batchSize

	startTime := time.Now() // start active processing timer

	for b := 0; b < batches; b++ {
		// Send batch command; here we use "BATCH 100" to notify the server.
		err = enc.Encode("BATCH 1000")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write error (BATCH cmd): %v\n", err)
			return
		}

		// Set a small read timeout to avoid getting stuck forever.
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		var response string
		err = dec.Decode(&response)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Read error (waiting for OK): %v\n", err)
			return
		}
		if !strings.Contains(response, "OK") {
			fmt.Fprintf(os.Stderr, "Unexpected server response: %s\n", response)
			return
		}

		// Send batchSize SET requests
		for j := 0; j < batchSize; j++ {
			msg := randomKeyValue()
			err = enc.Encode(msg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Write error (SET): %v\n", err)
				return
			}
		}

		// Read exactly batchSize responses
		setOKCount := 0
		for setOKCount < batchSize {
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			var resp string
			err = dec.Decode(&resp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Read error (SET done): %v\n", err)
				return
			}
			if strings.Contains(resp, "SET done") {
				setOKCount++
			}
		}
	}
	activeDuration := time.Since(startTime)
	atomic.AddInt64(&totalActiveTime, int64(activeDuration))
	atomic.AddInt64(&totalKeysProcessed, int64(n))
}

func main() {
	log.SetOutput(io.Discard)
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start the global timer before launching goroutines.
	globalStart := time.Now()

	// Launch numGoroutines concurrent connections.
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			spamConnection()
		}()
		time.Sleep(500 * time.Microsecond)
	}

	wg.Wait()

	// Measure the global end time.
	globalEnd := time.Now()
	elapsed := globalEnd.Sub(globalStart).Seconds()

	// Compute throughput based on wall-clock time.
	throughput := float64(totalKeysProcessed) / elapsed

	fmt.Printf("Processed %d keys in %.3f seconds.\n", totalKeysProcessed, elapsed)
	fmt.Printf("Overall throughput: %.2f keys/second\n", throughput)
}
