package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dvictor357/pipeline-plugin-system/core"
	"github.com/dvictor357/pipeline-plugin-system/moderation"
)

// ModerationRequest represents the incoming HTTP request payload
type ModerationRequest struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	AuthorID string `json:"author_id"`
}

// ModerationResponse represents the HTTP response payload
type ModerationResponse struct {
	ContentID string                     `json:"content_id"`
	Action    string                     `json:"action"`
	Flagged   bool                       `json:"flagged"`
	Reason    string                     `json:"reason"`
	Score     moderation.ModerationScore `json:"score"`
	Timestamp time.Time                  `json:"timestamp"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// ModerationServer wraps the pipeline and provides HTTP endpoints
type ModerationServer struct {
	pipeline *core.Pipeline
}

// NewModerationServer creates a new moderation server with the configured pipeline
func NewModerationServer() *ModerationServer {
	pipeline := core.NewPipeline(core.AbortOnError).
		Use(moderation.NewProfanityFilterPlugin()).
		Use(moderation.NewSpamDetectorPlugin()).
		Use(moderation.NewSentimentAnalyzerPlugin()).
		Use(moderation.NewScoringPlugin()).
		Use(moderation.NewDecisionRouterPlugin()).
		Use(moderation.NewActionHandlerPlugin())

	return &ModerationServer{
		pipeline: pipeline,
	}
}

// HandleModerate processes incoming content for moderation
func (s *ModerationServer) HandleModerate(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Parse request body
	var req ModerationRequest
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
	if req.ID == "" {
		req.ID = fmt.Sprintf("content-%d", time.Now().Unix())
	}
	if req.AuthorID == "" {
		req.AuthorID = "anonymous"
	}

	// Create content
	content := moderation.Content{
		ID:        req.ID,
		Text:      req.Text,
		AuthorID:  req.AuthorID,
		Timestamp: time.Now(),
	}

	// Create context and execute pipeline
	ctx := core.NewContext(&content)
	if err := s.pipeline.Execute(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Pipeline error: %v", err)})
		return
	}

	// Extract result
	result, ok := ctx.GetData().(*moderation.ModerationResult)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unexpected result type"})
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ModerationResponse{
		ContentID: result.Content.ID,
		Action:    result.Decision.Action,
		Flagged:   result.Decision.Flagged,
		Reason:    result.Decision.Reason,
		Score:     result.Decision.Score,
		Timestamp: result.Content.Timestamp,
	})
}

// HandleHealth provides a health check endpoint
func (s *ModerationServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func main() {
	server := NewModerationServer()

	// Register handlers
	http.HandleFunc("/moderate", server.HandleModerate)
	http.HandleFunc("/health", server.HandleHealth)

	// Start server
	port := ":8081"
	fmt.Printf("Content Moderation Server starting on %s\n", port)
	fmt.Println("\nExample curl commands:")
	fmt.Println("\n# Health check:")
	fmt.Println("curl http://localhost:8081/health")
	fmt.Println("\n# Moderate clean content (should approve):")
	fmt.Println(`curl -X POST http://localhost:8081/moderate \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"text":"This is a great product! I love it.","author_id":"user123","id":"content-001"}'`)
	fmt.Println("\n# Moderate negative content (should review):")
	fmt.Println(`curl -X POST http://localhost:8081/moderate \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"text":"This is terrible and awful. I hate this.","author_id":"user456","id":"content-002"}'`)
	fmt.Println("\n# Moderate profane content (should reject):")
	fmt.Println(`curl -X POST http://localhost:8081/moderate \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"text":"This is badword1 and offensive inappropriate vulgar content.","author_id":"user789","id":"content-003"}'`)
	fmt.Println("\n# Moderate spam content (should reject):")
	fmt.Println(`curl -X POST http://localhost:8081/moderate \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"text":"BUY NOW!!! https://spam.com https://scam.com https://fake.com https://bad.com","author_id":"spammer","id":"content-004"}'`)
	fmt.Println()

	log.Fatal(http.ListenAndServe(port, nil))
}

/*
Example curl commands for testing:

# Health check
curl http://localhost:8081/health

# Clean content (approve)
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{"text":"This is a great product! I love it and would recommend it.","author_id":"user123","id":"content-001"}'

# Negative sentiment (review)
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{"text":"This is terrible and awful. I hate this product so much.","author_id":"user456","id":"content-002"}'

# Profanity (reject)
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{"text":"This is badword1 and offensive content that is inappropriate and vulgar.","author_id":"user789","id":"content-003"}'

# Spam patterns (reject)
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{"text":"BUY NOW!!! CLICK HERE https://spam.com https://scam.com https://fake.com https://bad.com HELLOOOOOOO!!!","author_id":"spammer","id":"content-004"}'

# Borderline content (review)
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{"text":"Not the best experience. Could be better. Somewhat disappointing.","author_id":"critic","id":"content-005"}'

# Positive content (approve)
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{"text":"Excellent service! Amazing quality and fantastic support. Very happy!","author_id":"happy-user","id":"content-006"}'

# Test without optional fields (will use defaults)
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{"text":"This is a test message."}'

# Test with minimal profanity
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{"text":"This product is obscene and profanity filled with badword2 content.","author_id":"test-user"}'
*/
