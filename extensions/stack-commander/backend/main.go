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

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Stack struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	Services []Service `json:"services"`
	Path     string    `json:"path"`
}

type Service struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Ports  []int  `json:"ports"`
	Health string `json:"health"`
}

var dockerClient *client.Client

func main() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/stacks", handleListStacks)
	mux.HandleFunc("/api/stacks/start", handleStartStack)
	mux.HandleFunc("/api/stacks/stop", handleStopStack)
	mux.HandleFunc("/api/stacks/restart", handleRestartStack)
	mux.HandleFunc("/api/stacks/logs", handleStackLogs)
	mux.HandleFunc("/api/health", handleHealth)

	// Listen on Unix socket for Docker Desktop SDK communication
	socketPath := "/run/guest-services/backend.sock"
	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on socket: %v", err)
	}
	defer listener.Close()

	server := &http.Server{Handler: mux}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		server.Shutdown(context.Background())
	}()

	log.Printf("Stack Commander backend listening on %s", socketPath)
	if err := server.Serve(listener); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func handleListStacks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Group containers by compose project label
	stackMap := make(map[string]*Stack)
	for _, c := range containers {
		project := c.Labels["com.docker.compose.project"]
		if project == "" {
			continue
		}
		if _, exists := stackMap[project]; !exists {
			stackMap[project] = &Stack{
				Name:     project,
				Status:   "running",
				Services: []Service{},
				Path:     c.Labels["com.docker.compose.project.working_dir"],
			}
		}
		svc := Service{
			Name:   c.Labels["com.docker.compose.service"],
			Status: c.State,
			Health: c.Status,
		}
		for _, p := range c.Ports {
			if p.PublicPort > 0 {
				svc.Ports = append(svc.Ports, int(p.PublicPort))
			}
		}
		stackMap[project].Services = append(stackMap[project].Services, svc)
	}

	// Determine stack-level status
	stacks := make([]Stack, 0, len(stackMap))
	for _, stack := range stackMap {
		allRunning := true
		anyStopped := false
		for _, svc := range stack.Services {
			if svc.Status != "running" {
				allRunning = false
			}
			if svc.Status == "exited" || svc.Status == "dead" {
				anyStopped = true
			}
		}
		if allRunning {
			stack.Status = "running"
		} else if anyStopped {
			stack.Status = "partial"
		} else {
			stack.Status = "stopped"
		}
		stacks = append(stacks, *stack)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stacks)
}

func handleStartStack(w http.ResponseWriter, r *http.Request) {
	// Placeholder: execute `docker compose -p <name> up -d`
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "starting"})
}

func handleStopStack(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopping"})
}

func handleRestartStack(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "restarting"})
}

func handleStackLogs(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"logs": "streaming not yet implemented"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
