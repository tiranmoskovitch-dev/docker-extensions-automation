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
)

type MCPServer struct {
	Name      string   `json:"name"`
	Transport string   `json:"transport"` // sse, stdio, streamable-http
	URL       string   `json:"url"`
	Status    string   `json:"status"`
	Tools     []string `json:"tools"`
}

type GatewayConfig struct {
	Transport string      `json:"transport"`
	Servers   []MCPServer `json:"servers"`
	Status    string      `json:"status"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/gateway/status", handleGatewayStatus)
	mux.HandleFunc("/api/gateway/deploy", handleDeployGateway)
	mux.HandleFunc("/api/servers", handleListServers)
	mux.HandleFunc("/api/servers/add", handleAddServer)
	mux.HandleFunc("/api/servers/remove", handleRemoveServer)
	mux.HandleFunc("/api/servers/tools", handleListTools)
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

	log.Printf("MCP Gateway Manager backend listening on %s", socketPath)
	server.Serve(listener)
}

func handleGatewayStatus(w http.ResponseWriter, r *http.Request) {
	// Check if docker/mcp-gateway container is running
	config := GatewayConfig{
		Transport: "sse",
		Status:    "not_deployed",
		Servers:   []MCPServer{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func handleDeployGateway(w http.ResponseWriter, r *http.Request) {
	// Deploy docker/mcp-gateway:latest with provided config
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "deploying",
		"image":   "docker/mcp-gateway:latest",
		"message": "Pulling and starting MCP Gateway",
	})
}

func handleListServers(w http.ResponseWriter, r *http.Request) {
	servers := []MCPServer{
		{
			Name:      "postgres",
			Transport: "sse",
			URL:       "http://mcp-gateway:8811/sse",
			Status:    "connected",
			Tools:     []string{"query", "list_tables", "describe_table"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

func handleAddServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}

func handleRemoveServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

func handleListTools(w http.ResponseWriter, r *http.Request) {
	tools := map[string][]string{
		"postgres":   {"query", "list_tables", "describe_table"},
		"filesystem": {"read_file", "write_file", "list_directory"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
