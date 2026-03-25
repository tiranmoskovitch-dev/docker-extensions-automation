package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Agent struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Status    string            `json:"status"`
	Health    string            `json:"health"`
	Labels    map[string]string `json:"labels"`
	CPU       float64           `json:"cpu"`
	Memory    int64             `json:"memory"`
	CreatedAt int64             `json:"created_at"`
}

var dockerClient *client.Client

func main() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/agents", handleListAgents)
	mux.HandleFunc("/api/agents/start", handleStartAgent)
	mux.HandleFunc("/api/agents/stop", handleStopAgent)
	mux.HandleFunc("/api/agents/logs", handleAgentLogs)
	mux.HandleFunc("/api/agents/stats", handleAgentStats)
	mux.HandleFunc("/api/health", handleHealth)

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

	log.Printf("Agent Orchestrator backend listening on %s", socketPath)
	server.Serve(listener)
}

func handleListAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	agents := []Agent{}
	for _, c := range containers {
		// Detect agent containers by label or naming convention
		isAgent := false
		if _, ok := c.Labels["automation.type"]; ok {
			isAgent = true
		}
		for _, name := range c.Names {
			if strings.Contains(name, "agent") || strings.Contains(name, "worker") {
				isAgent = true
			}
		}
		if !isAgent {
			continue
		}

		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		agents = append(agents, Agent{
			ID:        c.ID[:12],
			Name:      name,
			Image:     c.Image,
			Status:    c.State,
			Health:    c.Status,
			Labels:    c.Labels,
			CreatedAt: c.Created,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func handleStartAgent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "starting"})
}

func handleStopAgent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopping"})
}

func handleAgentLogs(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"logs": "streaming endpoint"})
}

func handleAgentStats(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"stats": "resource monitoring"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
