package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
    "sync"
    "time"
)

func main() {
    // Number of keys to test
    numKeys := 10

    // WaitGroup to wait for all goroutines to finish
    var wg sync.WaitGroup

    // Slice to store response times
    var responseTimes []time.Duration
    var responseTimesMutex sync.Mutex

    // Function to perform SET and GET operations
    performTest := func(key, value string) {
        defer wg.Done()

        // Connect to the server
        conn, err := net.Dial("tcp", "localhost:8080")
        if err != nil {
            fmt.Println("Error connecting to server:", err)
            return
        }
        defer conn.Close()

        reader := bufio.NewReader(conn)

        // Send SET command and wait for response
        setCommand := fmt.Sprintf("SET %s %s\n", key, value)
        startTime := time.Now()
        _, err = conn.Write([]byte(setCommand))
        if err != nil {
            fmt.Println("Error sending SET command:", err)
            return
        }

        // Wait for server response for SET
        setResponse, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Error reading SET response:", err)
            return
        }
        setDuration := time.Since(startTime)

        // Record response time
        responseTimesMutex.Lock()
        responseTimes = append(responseTimes, setDuration)
        responseTimesMutex.Unlock()

        // Send GET command and wait for response
        getCommand := fmt.Sprintf("GET %s\n", key)
        startTime = time.Now()
        _, err = conn.Write([]byte(getCommand))
        if err != nil {
            fmt.Println("Error sending GET command:", err)
            return
        }

        // Wait for server response for GET
        getResponse, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Error reading GET response:", err)
            return
        }
        getDuration := time.Since(startTime)

        // Record response time
        responseTimesMutex.Lock()
        responseTimes = append(responseTimes, getDuration)
        responseTimesMutex.Unlock()

        // Trim newline characters from responses
        setResponse = strings.TrimSpace(setResponse)
        getResponse = strings.TrimSpace(getResponse)

        // Print responses and times
        fmt.Printf("SET Response for key '%s': %s (Time: %v)\n", key, setResponse, setDuration)
        fmt.Printf("GET Response for key '%s': %s (Time: %v)\n", key, getResponse, getDuration)
    }

    // Start multiple goroutines to perform tests concurrently
    for i := 0; i < numKeys; i++ {
        wg.Add(1)
        key := fmt.Sprintf("key%d", i)
        value := fmt.Sprintf("value%d", i)
        go performTest(key, value)
    }

    // Wait for all goroutines to finish
    wg.Wait()

    // Calculate average response time
    var totalDuration time.Duration
    for _, duration := range responseTimes {
		fmt.Println("time taken",duration)
        totalDuration += duration
    }
    averageDuration := totalDuration / time.Duration(len(responseTimes))

    fmt.Printf("\nTotal Requests: %d\n", len(responseTimes))
    fmt.Printf("Average Response Time: %v\n", averageDuration)
}
