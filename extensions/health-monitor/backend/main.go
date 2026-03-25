package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type HealthRecord struct {
	ContainerID   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	Status        string    `json:"status"`
	Health        string    `json:"health"`
	Uptime        string    `json:"uptime"`
	RestartCount  int       `json:"restart_count"`
	LastChecked   time.Time `json:"last_checked"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryMB      float64   `json:"memory_mb"`
}

type AlertRule struct {
	Name      string  `json:"name"`
	Metric    string  `json:"metric"` // cpu, memory, health, restart_count
	Threshold float64 `json:"threshold"`
	Action    string  `json:"action"` // notify, restart, webhook
	Enabled   bool    `json:"enabled"`
}

var dockerClient *client.Client

func main() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health/containers", handleContainerHealth)
	mux.HandleFunc("/api/health/history", handleHealthHistory)
	mux.HandleFunc("/api/alerts", handleListAlerts)
	mux.HandleFunc("/api/alerts/create", handleCreateAlert)
	mux.HandleFunc("/api/alerts/delete", handleDeleteAlert)
	mux.HandleFunc("/api/restart", handleRestartContainer)
	mux.HandleFunc("/api/health", handleSelfHealth)

	socketPath := "/run/guest-services/backend.sock"
	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	server := &http.Server{Handler: mux}
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		server.Shutdown(context.Background())
	}()

	// Start background health checker
	go healthCheckLoop()

	log.Printf("Health Monitor backend listening on %s", socketPath)
	server.Serve(listener)
}

func healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		ctx := context.Background()
		containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
		if err != nil {
			log.Printf("Health check error: %v", err)
			continue
		}
		for _, c := range containers {
			if c.State != "running" {
				log.Printf("Container %s is %s", c.ID[:12], c.State)
			}
		}
	}
}

func handleContainerHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	records := []HealthRecord{}
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0][1:] // strip leading /
		}
		records = append(records, HealthRecord{
			ContainerID:   c.ID[:12],
			ContainerName: name,
			Status:        c.State,
			Health:        c.Status,
			RestartCount:  0,
			LastChecked:   time.Now(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func handleHealthHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func handleListAlerts(w http.ResponseWriter, r *http.Request) {
	defaults := []AlertRule{
		{Name: "High CPU", Metric: "cpu", Threshold: 90, Action: "notify", Enabled: true},
		{Name: "High Memory", Metric: "memory", Threshold: 80, Action: "notify", Enabled: true},
		{Name: "Unhealthy", Metric: "health", Threshold: 1, Action: "restart", Enabled: false},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(defaults)
}

func handleCreateAlert(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func handleDeleteAlert(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func handleRestartContainer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "restarting"})
}

func handleSelfHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
