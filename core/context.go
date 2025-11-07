package core

// Context carries data and metadata through the pipeline.
// It supports both stateless transformations and stateful processing.
type Context struct {
	Data     any            // Primary data being processed
	Metadata map[string]any // Additional metadata
	Errors   []error        // Collected errors (for continue-on-error mode)
	state    map[string]any // Internal state for stateful pipelines
}

// NewContext creates a new Context with the given data.
func NewContext(data any) *Context {
	return &Context{
		Data:     data,
		Metadata: make(map[string]any),
		Errors:   make([]error, 0),
		state:    make(map[string]any),
	}
}

// Set stores a value in the metadata with the given key.
func (c *Context) Set(key string, value any) {
	c.Metadata[key] = value
}

// Get retrieves a value from the metadata by key.
// Returns the value and a boolean indicating whether the key exists.
func (c *Context) Get(key string) (any, bool) {
	value, exists := c.Metadata[key]
	return value, exists
}

// SetData updates the primary data being processed.
func (c *Context) SetData(data any) {
	c.Data = data
}

// GetData returns the primary data being processed.
func (c *Context) GetData() any {
	return c.Data
}

// SetState stores a value in the internal state map.
// This is used for stateful pipelines like conversation management.
func (c *Context) SetState(key string, value any) {
	c.state[key] = value
}

// GetState retrieves a value from the internal state map.
// Returns the value and a boolean indicating whether the key exists.
func (c *Context) GetState(key string) (any, bool) {
	value, exists := c.state[key]
	return value, exists
}

// AddError appends an error to the error collection.
// This is used in continue-on-error mode to track failures without stopping execution.
func (c *Context) AddError(err error) {
	c.Errors = append(c.Errors, err)
}
