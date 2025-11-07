package core

// Plugin defines the interface that all processing units must implement.
// Each plugin receives a Context, performs its processing, and returns an error if something goes wrong.
type Plugin interface {
	Execute(ctx *Context) error
}
