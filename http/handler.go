package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dvictor357/pipeline-plugin-system/core"
)

// HTTPHandler adapts a Pipeline to work as an http.Handler.
// It converts HTTP requests into pipeline Context and writes responses.
type HTTPHandler struct {
	pipeline *core.Pipeline
}

// NewHTTPHandler creates a new HTTPHandler with the given pipeline.
func NewHTTPHandler(pipeline *core.Pipeline) *HTTPHandler {
	return &HTTPHandler{
		pipeline: pipeline,
	}
}

// ServeHTTP implements the http.Handler interface.
// It extracts request data into a Context, executes the pipeline, and writes the response.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON body into a map
	var data map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &data); err != nil {
			http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
			return
		}
	} else {
		data = make(map[string]any)
	}

	// Create Context with request data
	ctx := core.NewContext(data)

	// Extract headers into metadata
	headers := make(map[string]any)
	for key, values := range r.Header {
		if len(values) == 1 {
			headers[key] = values[0]
		} else {
			headers[key] = values
		}
	}
	ctx.Set("headers", headers)

	// Extract query parameters into metadata
	queryParams := make(map[string]any)
	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			queryParams[key] = values[0]
		} else {
			queryParams[key] = values
		}
	}
	ctx.Set("query", queryParams)

	// Store HTTP method and path
	ctx.Set("method", r.Method)
	ctx.Set("path", r.URL.Path)

	// Execute pipeline
	if err := h.pipeline.Execute(ctx); err != nil {
		// Pipeline execution failed
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if there were collected errors (ContinueOnError mode)
	if len(ctx.Errors) > 0 {
		// Return 422 Unprocessable Entity with error details
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": formatErrors(ctx.Errors),
		})
		return
	}

	// Write successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ctx.GetData()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// formatErrors converts a slice of errors into a slice of error messages.
func formatErrors(errors []error) []string {
	messages := make([]string, len(errors))
	for i, err := range errors {
		messages[i] = err.Error()
	}
	return messages
}
