package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dvictor357/pipeline-plugin-system/chatbot"
	"github.com/dvictor357/pipeline-plugin-system/core"
)

// ChatRequest represents the incoming HTTP request payload
type ChatRequest struct {
	Text      string `json:"text"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

// ChatResponse represents the HTTP response payload
type ChatResponse struct {
	Text      string           `json:"text"`
	Intent    chatbot.Intent   `json:"intent"`
	Entities  []chatbot.Entity `json:"entities"`
	Timestamp time.Time        `json:"timestamp"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// ChatBotServer wraps the pipeline and provides HTTP endpoints
type ChatBotServer struct {
	pipeline *core.Pipeline
}

// NewChatBotServer creates a new chat bot server with the configured pipeline
func NewChatBotServer() *ChatBotServer {
	pipeline := core.NewPipeline(core.AbortOnError).
		Use(chatbot.NewIntentClassifierPlugin()).
		Use(chatbot.NewEntityExtractorPlugin()).
		Use(chatbot.NewContextManagerPlugin(10)).
		Use(chatbot.NewResponseGeneratorPlugin()).
		Use(chatbot.NewPersonalityFilterPlugin(chatbot.PersonalityConfig{
			Name:         "Friendly Bot",
			Emojis:       true,
			Casual:       true,
			Enthusiastic: false,
		}))

	return &ChatBotServer{
		pipeline: pipeline,
	}
}

// HandleChat processes incoming chat messages
func (s *ChatBotServer) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Parse request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Text field is required"})
		return
	}
	if req.UserID == "" {
		req.UserID = "anonymous"
	}
	if req.SessionID == "" {
		req.SessionID = "default-session"
	}

	// Create message
	msg := chatbot.Message{
		Text:      req.Text,
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Timestamp: time.Now(),
	}

	// Create context and execute pipeline
	ctx := core.NewContext(msg)
	if err := s.pipeline.Execute(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Pipeline error: %v", err)})
		return
	}

	// Extract response
	response, ok := ctx.GetData().(chatbot.Response)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unexpected response type"})
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ChatResponse{
		Text:      response.Text,
		Intent:    response.Intent,
		Entities:  response.Entities,
		Timestamp: response.Timestamp,
	})
}

// HandleHealth provides a health check endpoint
func (s *ChatBotServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func main() {
	server := NewChatBotServer()

	// Register handlers
	http.HandleFunc("/chat", server.HandleChat)
	http.HandleFunc("/health", server.HandleHealth)

	// Start server
	port := ":8080"
	fmt.Printf("Chat Bot Server starting on %s\n", port)
	fmt.Println("\nExample curl commands:")
	fmt.Println("\n# Health check:")
	fmt.Println("curl http://localhost:8080/health")
	fmt.Println("\n# Send a greeting:")
	fmt.Println(`curl -X POST http://localhost:8080/chat \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"text":"Hello! How are you?","user_id":"user123","session_id":"session456"}'`)
	fmt.Println("\n# Ask a question with entities:")
	fmt.Println(`curl -X POST http://localhost:8080/chat \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"text":"What is the weather on December 25th?","user_id":"user123","session_id":"session456"}'`)
	fmt.Println("\n# Send contact information:")
	fmt.Println(`curl -X POST http://localhost:8080/chat \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"text":"My email is john@example.com and phone is 555-1234","user_id":"user123","session_id":"session456"}'`)
	fmt.Println("\n# Say goodbye:")
	fmt.Println(`curl -X POST http://localhost:8080/chat \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"text":"Thanks! Goodbye!","user_id":"user123","session_id":"session456"}'`)
	fmt.Println()

	log.Fatal(http.ListenAndServe(port, nil))
}

/*
Example curl commands for testing:

# Health check
curl http://localhost:8080/health

# Basic greeting
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello! How are you?","user_id":"user123","session_id":"session456"}'

# Question with date entity
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"text":"What is the weather on December 25th?","user_id":"user123","session_id":"session456"}'

# Command with multiple entities
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"text":"My email is john@example.com and phone is 555-1234","user_id":"user123","session_id":"session456"}'

# Farewell
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"text":"Thanks! Goodbye!","user_id":"user123","session_id":"session456"}'

# Test conversation state (send multiple messages with same session_id)
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"text":"Hi there!","user_id":"user123","session_id":"conv789"}'

curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"text":"Can you help me?","user_id":"user123","session_id":"conv789"}'

curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"text":"Thank you!","user_id":"user123","session_id":"conv789"}'
*/
