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

type Model struct {
	Name         string  `json:"name"`
	Size         string  `json:"size"`
	Quantization string  `json:"quantization"`
	Status       string  `json:"status"` // available, running, downloading
	Endpoint     string  `json:"endpoint"`
	GPUMemory    float64 `json:"gpu_memory_mb"`
}

type InferenceStats struct {
	TotalRequests   int64   `json:"total_requests"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	TokensPerSecond float64 `json:"tokens_per_second"`
	ActiveModel     string  `json:"active_model"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/models", handleListModels)
	mux.HandleFunc("/api/models/pull", handlePullModel)
	mux.HandleFunc("/api/models/remove", handleRemoveModel)
	mux.HandleFunc("/api/models/stats", handleInferenceStats)
	mux.HandleFunc("/api/models/compare", handleCompareOllama)
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

	log.Printf("Model Runner Dashboard backend listening on %s", socketPath)
	server.Serve(listener)
}

func handleListModels(w http.ResponseWriter, r *http.Request) {
	// Query Docker Model Runner for available models
	// Uses the `models:` top-level Compose key
	models := []Model{
		{
			Name:         "ai/qwen3:14B-Q6_K",
			Size:         "11.2 GB",
			Quantization: "Q6_K",
			Status:       "available",
			Endpoint:     "model-runner/v1",
		},
		{
			Name:         "ai/llama3.1:8B-Q4_K_M",
			Size:         "4.9 GB",
			Quantization: "Q4_K_M",
			Status:       "available",
			Endpoint:     "model-runner/v1",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func handlePullModel(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "pulling"})
}

func handleRemoveModel(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

func handleInferenceStats(w http.ResponseWriter, r *http.Request) {
	stats := InferenceStats{
		ActiveModel: "ai/qwen3:14B-Q6_K",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleCompareOllama(w http.ResponseWriter, r *http.Request) {
	// Compare Model Runner setup vs current Ollama setup
	comparison := map[string]interface{}{
		"model_runner": map[string]string{
			"setup":       "Native Docker Compose `models:` key",
			"gpu_support": "Linux + Docker Desktop 4.43+",
			"api":         "OpenAI-compatible",
			"management":  "Declarative via compose.yaml",
		},
		"ollama": map[string]string{
			"setup":       "Separate service on port 11434",
			"gpu_support": "All platforms",
			"api":         "Ollama API + OpenAI-compatible",
			"management":  "CLI + REST API",
		},
		"recommendation": "Keep Ollama for now (better Windows GPU support). Monitor Model Runner for Windows parity.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
