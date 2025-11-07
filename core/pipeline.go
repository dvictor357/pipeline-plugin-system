package core

import (
	"fmt"
)

// ErrorStrategy defines how the pipeline handles plugin errors.
type ErrorStrategy int

const (
	// AbortOnError stops pipeline execution at the first error.
	AbortOnError ErrorStrategy = iota
	// ContinueOnError continues executing remaining plugins and collects errors.
	ContinueOnError
)

// Pipeline orchestrates the execution of plugins in sequential order.
type Pipeline struct {
	plugins       []Plugin
	errorStrategy ErrorStrategy
}

// NewPipeline creates a new Pipeline with the specified error handling strategy.
func NewPipeline(strategy ErrorStrategy) *Pipeline {
	return &Pipeline{
		plugins:       make([]Plugin, 0),
		errorStrategy: strategy,
	}
}

// Use adds a plugin to the pipeline and returns the pipeline for method chaining.
// This enables fluent interface for pipeline construction.
func (p *Pipeline) Use(plugin Plugin) *Pipeline {
	p.plugins = append(p.plugins, plugin)
	return p
}

// Execute runs all plugins in the pipeline sequentially.
// The behavior depends on the error strategy:
// - AbortOnError: stops at the first error and returns it wrapped with context
// - ContinueOnError: continues executing all plugins and collects errors in Context
func (p *Pipeline) Execute(ctx *Context) error {
	for i, plugin := range p.plugins {
		err := plugin.Execute(ctx)
		if err != nil {
			if p.errorStrategy == AbortOnError {
				// Wrap error with plugin context and return immediately
				return &PipelineError{
					PluginIndex: i,
					Err:         err,
				}
			}
			// ContinueOnError: collect error and continue
			ctx.AddError(&PipelineError{
				PluginIndex: i,
				Err:         err,
			})
		}
	}
	return nil
}

// PipelineError wraps plugin errors with context about which plugin failed.
type PipelineError struct {
	PluginIndex int
	Err         error
}

// Error implements the error interface.
func (e *PipelineError) Error() string {
	return fmt.Sprintf("plugin %d failed: %v", e.PluginIndex, e.Err)
}

// Unwrap returns the underlying error for error chain support.
func (e *PipelineError) Unwrap() error {
	return e.Err
}
