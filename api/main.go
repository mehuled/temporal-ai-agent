package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"temporal-ai-agent/workflows"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
)

// ChatRequest represents the request body for the /start-workflow endpoint
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatResponse represents the response from the /start-workflow endpoint
type ChatResponse struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
	Result     string `json:"result,omitempty"`
	Error      string `json:"error,omitempty"`
}

// SignalRequest represents the request body for signal endpoints
type SignalRequest struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id,omitempty"`
	Message    string `json:"message"`
}

// SignalResponse represents the response from signal endpoints
type SignalResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Server holds the HTTP server dependencies
type Server struct {
	temporalClient client.Client
	taskQueue      string
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get configuration from environment variables
	hostPort := getEnv("TEMPORAL_HOST_PORT", "localhost:7233")
	namespace := getEnv("TEMPORAL_NAMESPACE", "default")
	apiKey := getEnv("TEMPORAL_API_KEY", "")
	taskQueue := getEnv("TEMPORAL_TASK_QUEUE", "my-task-queue")
	tlsEnabled := getEnvBool("TEMPORAL_TLS_ENABLED", false)
	serverPort := getEnv("SERVER_PORT", "3000")

	// Validate required environment variables
	if apiKey == "" {
		log.Fatal("TEMPORAL_API_KEY environment variable is required")
	}

	// Configure client options
	clientOptions := client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	}

	// Configure TLS if enabled
	if tlsEnabled {
		clientOptions.ConnectionOptions = client.ConnectionOptions{TLS: &tls.Config{}}
	}

	// Configure credentials
	clientOptions.Credentials = client.NewAPIKeyStaticCredentials(apiKey)

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// Create server instance
	server := &Server{
		temporalClient: c,
		taskQueue:      taskQueue,
	}

	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/start-workflow", server.handleStartWorkflow).Methods("POST")
	r.HandleFunc("/signal/user-prompt", server.handleUserPromptSignal).Methods("POST")
	r.HandleFunc("/signal/confirm", server.handleConfirmSignal).Methods("POST")
	r.HandleFunc("/signal/end-chat", server.handleEndChatSignal).Methods("POST")
	r.HandleFunc("/health", server.handleHealth).Methods("GET")

	// Start HTTP server
	log.Printf("Starting API server on port %s", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, r))
}

// handleStartWorkflow handles POST /start-workflow requests
func (s *Server) handleStartWorkflow(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Start workflow
	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("chat-workflow-%d", time.Now().UnixNano()),
		TaskQueue: s.taskQueue,
	}

	we, err := s.temporalClient.ExecuteWorkflow(context.Background(), options, workflows.SayHelloWorkflow, req.Message)
	if err != nil {
		log.Printf("Unable to execute workflow: %v", err)
		response := ChatResponse{
			Error: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Started workflow: WorkflowID=%s, RunID=%s", we.GetID(), we.GetRunID())

	// Get workflow result
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Printf("Unable to get workflow result: %v", err)
		response := ChatResponse{
			WorkflowID: we.GetID(),
			RunID:      we.GetRunID(),
			Error:      err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return successful response
	response := ChatResponse{
		WorkflowID: we.GetID(),
		RunID:      we.GetRunID(),
		Result:     result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUserPromptSignal handles POST /signal/user-prompt requests
func (s *Server) handleUserPromptSignal(w http.ResponseWriter, r *http.Request) {
	var req SignalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.WorkflowID == "" {
		http.Error(w, "WorkflowID is required", http.StatusBadRequest)
		return
	}

	err := s.temporalClient.SignalWorkflow(context.Background(), req.WorkflowID, req.RunID, "user_prompt", req.Message)
	if err != nil {
		log.Printf("Error sending user_prompt signal: %v", err)
		response := SignalResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := SignalResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleConfirmSignal handles POST /signal/confirm requests
func (s *Server) handleConfirmSignal(w http.ResponseWriter, r *http.Request) {
	var req SignalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.WorkflowID == "" {
		http.Error(w, "WorkflowID is required", http.StatusBadRequest)
		return
	}

	err := s.temporalClient.SignalWorkflow(context.Background(), req.WorkflowID, req.RunID, "confirm", req.Message)
	if err != nil {
		log.Printf("Error sending confirm signal: %v", err)
		response := SignalResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := SignalResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleEndChatSignal handles POST /signal/end-chat requests
func (s *Server) handleEndChatSignal(w http.ResponseWriter, r *http.Request) {
	var req SignalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.WorkflowID == "" {
		http.Error(w, "WorkflowID is required", http.StatusBadRequest)
		return
	}

	err := s.temporalClient.SignalWorkflow(context.Background(), req.WorkflowID, req.RunID, "end_chat", req.Message)
	if err != nil {
		log.Printf("Error sending end_chat signal: %v", err)
		response := SignalResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := SignalResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles GET /health requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable with a fallback default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
