# Pipeline Plugin System

A lightweight, flexible pipeline plugin system written in Go that enables chaining processing steps together to handle data transformations. The framework provides a minimal skeleton that supports diverse use cases through composable plugins.

## Features

- **Minimal Dependencies**: Uses only Go's standard library
- **Flexible Architecture**: Chain plugins together to create custom processing pipelines
- **Error Handling**: Configurable error strategies (abort or continue on error)
- **Stateful Processing**: Built-in support for maintaining state across pipeline executions
- **HTTP Integration**: Easy integration with HTTP handlers using `net/http`
- **Thread-Safe**: Registry supports concurrent access with RWMutex
- **Extensible**: Simple Plugin interface makes it easy to create custom plugins

## Architecture

The system consists of four core components:

```
┌───────────────────────────────────────────────────────────┐
│                     Application Layer                     │
│  ┌──────────────────────┐    ┌──────────────────────┐     │
│  │  Chat Bot Pipeline   │    │ Content Moderation   │     │
│  │                      │    │     Pipeline         │     │
│  └──────────┬───────────┘    └──────────┬───────────┘     │
└─────────────┼──────────────────────────┼──────────────────┘
              │                          │
┌─────────────┼──────────────────────────┼─────────────────┐
│             │    Core Framework        │                 │
│             ▼                          ▼                 │
│  ┌──────────────────┐      ┌──────────────────┐          │
│  │     Pipeline     │◄─────┤     Registry     │          │
│  └────────┬─────────┘      └──────────────────┘          │
│           │                                              │
│           │ executes                                     │
│           ▼                                              │
│  ┌──────────────────┐      ┌──────────────────┐          │
│  │  Plugin Interface│      │     Context      │          │
│  └──────────────────┘      └──────────────────┘          │
└──────────────────────────────────────────────────────────┘
```

### Core Components

1. **Plugin Interface**: Defines the contract for all processing units
2. **Pipeline**: Orchestrates plugin execution and manages data flow
3. **Context**: Carries request-scoped data and metadata through the pipeline
4. **Registry**: Manages plugin registration and retrieval

## Installation

```bash
go get github.com/dvictor357/pipeline-plugin-system
```

## Quick Start

### Basic Pipeline

```go
package main

import (
    "fmt"
    "github.com/dvictor357/pipeline-plugin-system/core"
)

// Create a simple plugin
type GreetingPlugin struct{}

func (p *GreetingPlugin) Execute(ctx *core.Context) error {
    name := ctx.GetData().(string)
    greeting := fmt.Sprintf("Hello, %s!", name)
    ctx.SetData(greeting)
    return nil
}

func main() {
    // Create pipeline with abort-on-error strategy
    pipeline := core.NewPipeline(core.AbortOnError).
        Use(&GreetingPlugin{})

    // Create context with input data
    ctx := core.NewContext("World")

    // Execute pipeline
    err := pipeline.Execute(ctx)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    // Get result
    result := ctx.GetData().(string)
    fmt.Println(result) // Output: Hello, World!
}
```

## Core API Reference

### Plugin Interface

All plugins must implement the `Plugin` interface:

```go
type Plugin interface {
    Execute(ctx *Context) error
}
```

**Example Plugin:**

```go
type MyPlugin struct {
    config MyConfig
}

func (p *MyPlugin) Execute(ctx *core.Context) error {
    // Get input data
    data := ctx.GetData()

    // Process data
    result := processData(data)

    // Store result
    ctx.SetData(result)

    // Store metadata
    ctx.Set("processing_time", time.Since(start))

    return nil
}
```

### Context

The `Context` carries data and metadata through the pipeline.

**Constructor:**

```go
func NewContext(data any) *Context
```

**Data Methods:**

```go
// Set/get primary data
func (c *Context) SetData(data any)
func (c *Context) GetData() any

// Set/get metadata
func (c *Context) Set(key string, value any)
func (c *Context) Get(key string) (any, bool)

// Set/get internal state (for stateful pipelines)
func (c *Context) SetState(key string, value any)
func (c *Context) GetState(key string) (any, bool)

// Error collection (for continue-on-error mode)
func (c *Context) AddError(err error)
```

**Example:**

```go
ctx := core.NewContext(inputData)

// Store metadata
ctx.Set("user_id", "user-123")
ctx.Set("timestamp", time.Now())

// Retrieve metadata
if userID, exists := ctx.Get("user_id"); exists {
    fmt.Println("User ID:", userID)
}

// Manage state for stateful pipelines
ctx.SetState("conversation_history", []Message{})
if history, exists := ctx.GetState("conversation_history"); exists {
    messages := history.([]Message)
    // Use conversation history
}
```

### Pipeline

The `Pipeline` orchestrates plugin execution.

**Constructor:**

```go
func NewPipeline(strategy ErrorStrategy) *Pipeline
```

**Error Strategies:**

```go
const (
    AbortOnError    ErrorStrategy = iota  // Stop at first error
    ContinueOnError                       // Continue and collect errors
)
```

**Methods:**

```go
// Add plugin to pipeline (fluent interface)
func (p *Pipeline) Use(plugin Plugin) *Pipeline

// Execute all plugins sequentially
func (p *Pipeline) Execute(ctx *Context) error
```

**Example:**

```go
// Create pipeline with fluent interface
pipeline := core.NewPipeline(core.AbortOnError).
    Use(&Plugin1{}).
    Use(&Plugin2{}).
    Use(&Plugin3{})

// Execute pipeline
ctx := core.NewContext(data)
err := pipeline.Execute(ctx)
if err != nil {
    // Handle error
}
```

### Registry

The `Registry` manages plugin registration and retrieval.

**Constructor:**

```go
func NewRegistry() *Registry
```

**Methods:**

```go
// Register a plugin by name
func (r *Registry) Register(name string, plugin Plugin) error

// Retrieve a plugin by name
func (r *Registry) Get(name string) (Plugin, error)

// Build pipeline from plugin names
func (r *Registry) BuildPipeline(names []string, strategy ErrorStrategy) (*Pipeline, error)
```

**Example:**

```go
registry := core.NewRegistry()

// Register plugins
registry.Register("validator", &ValidatorPlugin{})
registry.Register("transformer", &TransformerPlugin{})
registry.Register("enricher", &EnricherPlugin{})

// Build pipeline from names
pipeline, err := registry.BuildPipeline(
    []string{"validator", "transformer", "enricher"},
    core.AbortOnError,
)
if err != nil {
    // Handle error
}

// Execute pipeline
ctx := core.NewContext(data)
pipeline.Execute(ctx)
```

## Creating Custom Plugins

### Step 1: Define Your Plugin Struct

```go
type MyCustomPlugin struct {
    // Configuration fields
    threshold float64
    enabled   bool
}
```

### Step 2: Implement the Execute Method

```go
func (p *MyCustomPlugin) Execute(ctx *core.Context) error {
    // 1. Get input data
    data := ctx.GetData()

    // 2. Type assert to expected type
    input, ok := data.(MyInputType)
    if !ok {
        return fmt.Errorf("expected MyInputType, got %T", data)
    }

    // 3. Perform processing
    result := p.process(input)

    // 4. Store result in context
    ctx.SetData(result)

    // 5. Optionally store metadata
    ctx.Set("my_plugin_metadata", someValue)

    return nil
}
```

### Step 3: Create Constructor (Optional)

```go
func NewMyCustomPlugin(threshold float64) *MyCustomPlugin {
    return &MyCustomPlugin{
        threshold: threshold,
        enabled:   true,
    }
}
```

### Step 4: Use Your Plugin

```go
pipeline := core.NewPipeline(core.AbortOnError).
    Use(NewMyCustomPlugin(0.75))

ctx := core.NewContext(myData)
err := pipeline.Execute(ctx)
```

## Example Applications

The repository includes two complete example applications demonstrating the framework's versatility.

### Chat Bot Pipeline

A conversational AI pipeline that processes user messages through multiple stages:

1. **Intent Classification**: Categorizes user messages (greeting, question, command, etc.)
2. **Entity Extraction**: Identifies key information (names, dates, emails, phone numbers)
3. **Context Management**: Maintains conversation state and history
4. **Response Generation**: Creates appropriate responses based on intent and entities
5. **Personality Filter**: Applies tone and style transformations

**Run the example:**

```bash
# Console example
go run examples/chatbot/example/chatbot_example.go

# HTTP server
go run examples/chatbot/server/chatbot_server.go
```

**HTTP API:**

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Hello! What is the weather on December 25th?",
    "user_id": "user-123",
    "session_id": "session-456"
  }'
```

### Content Moderation Pipeline

A content filtering pipeline that analyzes user-generated content:

1. **Profanity Filter**: Detects inappropriate language
2. **Spam Detector**: Identifies spam patterns (repeated characters, excessive links)
3. **Sentiment Analyzer**: Determines content sentiment and toxicity
4. **Scoring Plugin**: Aggregates weighted scores from all checks
5. **Decision Router**: Makes moderation decision based on thresholds
6. **Action Handler**: Executes the moderation action (approve/review/reject)

**Run the example:**

```bash
# Console example
go run examples/moderation/example/moderation_example.go

# HTTP server
go run examples/moderation/server/moderation_server.go
```

**HTTP API:**

```bash
curl -X POST http://localhost:8081/moderate \
  -H "Content-Type: application/json" \
  -d '{
    "id": "content-001",
    "text": "This is a great product! I love it.",
    "author_id": "user-123"
  }'
```

## HTTP Integration

The framework includes an HTTP handler adapter for easy web integration.

```go
package main

import (
    "net/http"
    "github.com/dvictor357/pipeline-plugin-system/core"
    "github.com/dvictor357/pipeline-plugin-system/http"
)

func main() {
    // Create pipeline
    pipeline := core.NewPipeline(core.AbortOnError).
        Use(&MyPlugin1{}).
        Use(&MyPlugin2{})

    // Create HTTP handler
    handler := httphandler.NewHTTPHandler(pipeline)

    // Register with HTTP server
    http.Handle("/process", handler)
    http.ListenAndServe(":8080", nil)
}
```

The HTTP handler automatically:

- Extracts request body, headers, and query parameters into Context
- Executes the pipeline
- Writes JSON response on success
- Returns appropriate HTTP error codes on failure

## Error Handling

### Abort on Error

Stops pipeline execution at the first error:

```go
pipeline := core.NewPipeline(core.AbortOnError).
    Use(&Plugin1{}).
    Use(&Plugin2{}).
    Use(&Plugin3{})

err := pipeline.Execute(ctx)
if err != nil {
    // Pipeline stopped at first error
    fmt.Printf("Pipeline failed: %v\n", err)
}
```

### Continue on Error

Continues executing all plugins and collects errors:

```go
pipeline := core.NewPipeline(core.ContinueOnError).
    Use(&Plugin1{}).
    Use(&Plugin2{}).
    Use(&Plugin3{})

err := pipeline.Execute(ctx)
// err is always nil in ContinueOnError mode

// Check collected errors
if len(ctx.Errors) > 0 {
    fmt.Println("Errors occurred:")
    for _, e := range ctx.Errors {
        fmt.Printf("  - %v\n", e)
    }
}
```

### Error Wrapping

Plugin errors are automatically wrapped with context:

```go
// Error format: "plugin <index> failed: <original error>"
// Example: "plugin 2 failed: validation failed: missing required field"
```

## Advanced Patterns

### Stateful Pipelines

Use the Context state map for stateful processing:

```go
type ConversationPlugin struct{}

func (p *ConversationPlugin) Execute(ctx *core.Context) error {
    // Retrieve conversation history
    var history []Message
    if h, exists := ctx.GetState("history"); exists {
        history = h.([]Message)
    }

    // Process current message
    msg := ctx.GetData().(Message)
    history = append(history, msg)

    // Limit history size
    if len(history) > 10 {
        history = history[len(history)-10:]
    }

    // Store updated history
    ctx.SetState("history", history)

    return nil
}
```

### Conditional Plugin Execution

Plugins can check metadata to conditionally execute:

```go
type ConditionalPlugin struct{}

func (p *ConditionalPlugin) Execute(ctx *core.Context) error {
    // Check if we should execute
    if skip, exists := ctx.Get("skip_validation"); exists && skip.(bool) {
        return nil // Skip this plugin
    }

    // Normal processing
    return p.process(ctx)
}
```

### Plugin Composition

Create higher-level plugins by composing simpler ones:

```go
type CompositePlugin struct {
    subPipeline *core.Pipeline
}

func NewCompositePlugin() *CompositePlugin {
    subPipeline := core.NewPipeline(core.AbortOnError).
        Use(&SubPlugin1{}).
        Use(&SubPlugin2{})

    return &CompositePlugin{
        subPipeline: subPipeline,
    }
}

func (p *CompositePlugin) Execute(ctx *core.Context) error {
    return p.subPipeline.Execute(ctx)
}
```

## Project Structure

```
pipeline-plugin-system/
├── core/
│   ├── plugin.go       # Plugin interface
│   ├── context.go      # Context implementation
│   ├── pipeline.go     # Pipeline orchestration
│   └── registry.go     # Plugin registry
├── http/
│   └── handler.go      # HTTP handler adapter
├── chatbot/
│   ├── models.go       # Chat bot data models
│   └── plugins.go      # Chat bot plugin implementations
├── moderation/
│   ├── models.go       # Moderation data models
│   └── plugins.go      # Moderation plugin implementations
├── examples/
│   ├── chatbot/
│   │   ├── example/
│   │   │   └── chatbot_example.go
│   │   └── server/
│   │       └── chatbot_server.go
│   └── moderation/
│       ├── example/
│       │   └── moderation_example.go
│       └── server/
│           └── moderation_server.go
├── go.mod
└── README.md
```

## Dependencies

The framework uses **only Go's standard library**:

- `net/http`: HTTP server and client functionality
- `encoding/json`: JSON encoding/decoding
- `sync`: Thread-safe registry with RWMutex
- `fmt`: Error formatting and string operations
- `errors`: Error handling and wrapping
- `strings`: String manipulation
- `regexp`: Pattern matching
- `time`: Timestamps and time-based operations

No third-party dependencies required!

## Best Practices

### Plugin Design

1. **Single Responsibility**: Each plugin should do one thing well
2. **Stateless Plugins**: Keep plugins stateless for concurrent use
3. **Type Safety**: Always type assert Context data with error checking
4. **Error Messages**: Return descriptive errors with context
5. **Metadata Usage**: Use metadata for cross-plugin communication

### Pipeline Construction

1. **Order Matters**: Plugins execute sequentially, so order is important
2. **Error Strategy**: Choose appropriate strategy for your use case
3. **Fluent Interface**: Use method chaining for readable pipeline construction
4. **Registry for Dynamic**: Use Registry when pipeline composition is dynamic

### Performance

1. **Reuse Plugins**: Plugins are stateless and can be reused across requests
2. **Context Per Request**: Create a new Context for each pipeline execution
3. **Avoid Deep Copies**: Use pointers for large data structures
4. **Concurrent Pipelines**: Pipelines are not thread-safe; create one per goroutine

## Testing

### Unit Testing Plugins

```go
func TestMyPlugin(t *testing.T) {
    plugin := NewMyPlugin()
    ctx := core.NewContext(testData)

    err := plugin.Execute(ctx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    result := ctx.GetData()
    // Assert result
}
```

### Integration Testing Pipelines

```go
func TestPipeline(t *testing.T) {
    pipeline := core.NewPipeline(core.AbortOnError).
        Use(&Plugin1{}).
        Use(&Plugin2{})

    ctx := core.NewContext(testData)
    err := pipeline.Execute(ctx)

    if err != nil {
        t.Fatalf("pipeline failed: %v", err)
    }

    // Assert final result
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.

## Examples and Demos

For complete working examples, see:

- **Chat Bot**: `examples/chatbot/example/chatbot_example.go`
- **Content Moderation**: `examples/moderation/example/moderation_example.go`
- **HTTP Servers**: `examples/chatbot/server/` and `examples/moderation/server/`

Run the examples to see the framework in action!
