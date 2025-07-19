package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Metrics struct {
	ServerID  string  `json:"server_id"`
	Timestamp string  `json:"timestamp"`
	CPUUsage  float64 `json:"cpu_usage"`
	MemUsage  float64 `json:"mem_usage"`
	DiskUsage float64 `json:"disk_usage"`
}

func randomMetrics(serverID string) *Metrics {
	return &Metrics{
		ServerID:  serverID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		CPUUsage:  10 + rand.Float64()*80, // 10% - 90%
		MemUsage:  20 + rand.Float64()*70, // 20% - 90%
		DiskUsage: 30 + rand.Float64()*50, // 30% - 80%
	}
}

func pushMetrics(m *Metrics) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}

	resp, err := http.Post("http://localhost:8080/metrics", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Push error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Pushed: %+v\n", m)
}

func startMockAgent(serverID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		m := randomMetrics(serverID)
		pushMetrics(m)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// List of fake server IDs
	serverIDs := []string{"srv-001", "srv-002", "srv-003", "srv-004"}

	for _, id := range serverIDs {
		go startMockAgent(id, 10*time.Second) // Send every 10s
	}

	select {} // Keep main alive
}
