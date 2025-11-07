package chatbot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dvictor357/pipeline-plugin-system/core"
)

// IntentClassifierPlugin analyzes message text to determine user intent using keyword-based classification
type IntentClassifierPlugin struct {
	keywords map[string][]string
}

// NewIntentClassifierPlugin creates a new intent classifier with predefined keyword patterns
func NewIntentClassifierPlugin() *IntentClassifierPlugin {
	return &IntentClassifierPlugin{
		keywords: map[string][]string{
			"greeting": {"hello", "hi", "hey", "good morning", "good afternoon", "good evening", "greetings"},
			"farewell": {"bye", "goodbye", "see you", "farewell", "take care", "later"},
			"question": {"what", "when", "where", "who", "why", "how", "can you", "could you", "would you", "?"},
			"command":  {"do", "make", "create", "show", "tell", "give", "send", "help"},
		},
	}
}

// Execute analyzes the message text and stores the detected intent in Context metadata
func (p *IntentClassifierPlugin) Execute(ctx *core.Context) error {
	// Extract message from context
	msg, ok := ctx.GetData().(Message)
	if !ok {
		return fmt.Errorf("expected Message type in context data")
	}

	text := strings.ToLower(msg.Text)

	// Classify intent based on keyword matching
	intent := Intent{
		Type:       "unknown",
		Confidence: 0.0,
	}

	maxMatches := 0
	for intentType, keywords := range p.keywords {
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				matches++
			}
		}

		if matches > maxMatches {
			maxMatches = matches
			intent.Type = intentType
			// Calculate confidence based on number of matches
			intent.Confidence = float64(matches) / float64(len(keywords))
			if intent.Confidence > 1.0 {
				intent.Confidence = 1.0
			}
		}
	}

	// If no keywords matched, check for question mark
	if intent.Type == "unknown" && strings.Contains(text, "?") {
		intent.Type = "question"
		intent.Confidence = 0.5
	}

	// Store intent in context metadata
	ctx.Set("intent", intent)

	return nil
}

// EntityExtractorPlugin identifies and extracts entities from message text using regex patterns
type EntityExtractorPlugin struct {
	patterns map[string]*regexp.Regexp
}

// NewEntityExtractorPlugin creates a new entity extractor with predefined regex patterns
func NewEntityExtractorPlugin() *EntityExtractorPlugin {
	return &EntityExtractorPlugin{
		patterns: map[string]*regexp.Regexp{
			"date":   regexp.MustCompile(`\b(\d{1,2}[/-]\d{1,2}[/-]\d{2,4}|(?:jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]* \d{1,2}(?:st|nd|rd|th)?(?:,? \d{4})?|today|tomorrow|yesterday)\b`),
			"number": regexp.MustCompile(`\b\d+(?:\.\d+)?\b`),
			"email":  regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
			"phone":  regexp.MustCompile(`\b(?:\+\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`),
			"name":   regexp.MustCompile(`\b[A-Z][a-z]+ [A-Z][a-z]+\b`),
		},
	}
}

// Execute identifies entities in the message text and stores them in Context metadata
func (p *EntityExtractorPlugin) Execute(ctx *core.Context) error {
	// Extract message from context
	msg, ok := ctx.GetData().(Message)
	if !ok {
		return fmt.Errorf("expected Message type in context data")
	}

	entities := make([]Entity, 0)

	// Extract entities using regex patterns
	for entityType, pattern := range p.patterns {
		matches := pattern.FindAllStringIndex(msg.Text, -1)
		for _, match := range matches {
			entity := Entity{
				Type:  entityType,
				Value: msg.Text[match[0]:match[1]],
				Start: match[0],
				End:   match[1],
			}
			entities = append(entities, entity)
		}
	}

	// Store entities in context metadata
	ctx.Set("entities", entities)

	return nil
}

// ContextManagerPlugin maintains conversation state across multiple message exchanges
type ContextManagerPlugin struct {
	maxHistorySize int
}

// NewContextManagerPlugin creates a new context manager with a maximum history size
func NewContextManagerPlugin(maxHistorySize int) *ContextManagerPlugin {
	if maxHistorySize <= 0 {
		maxHistorySize = 10
	}
	return &ContextManagerPlugin{
		maxHistorySize: maxHistorySize,
	}
}

// Execute retrieves and updates conversation history, limiting it to the last N messages
func (p *ContextManagerPlugin) Execute(ctx *core.Context) error {
	// Extract message from context
	msg, ok := ctx.GetData().(Message)
	if !ok {
		return fmt.Errorf("expected Message type in context data")
	}

	// Retrieve or initialize conversation state
	var convState ConversationState
	stateKey := fmt.Sprintf("conversation:%s", msg.SessionID)

	if stateData, exists := ctx.GetState(stateKey); exists {
		if state, ok := stateData.(ConversationState); ok {
			convState = state
		} else {
			// Initialize if type assertion fails
			convState = ConversationState{
				History:   make([]Message, 0),
				UserPrefs: make(map[string]any),
			}
		}
	} else {
		// Initialize new conversation state
		convState = ConversationState{
			History:   make([]Message, 0),
			UserPrefs: make(map[string]any),
		}
	}

	// Append current message to history
	convState.History = append(convState.History, msg)

	// Limit history to last N messages
	if len(convState.History) > p.maxHistorySize {
		convState.History = convState.History[len(convState.History)-p.maxHistorySize:]
	}

	// Update last intent if available
	if intentData, exists := ctx.Get("intent"); exists {
		if intent, ok := intentData.(Intent); ok {
			convState.LastIntent = intent
		}
	}

	// Store updated conversation state
	ctx.SetState(stateKey, convState)
	ctx.Set("conversation_state", convState)

	return nil
}

// ResponseGeneratorPlugin creates appropriate responses based on intent and entities
type ResponseGeneratorPlugin struct {
	templates map[string][]string
}

// NewResponseGeneratorPlugin creates a new response generator with predefined templates
func NewResponseGeneratorPlugin() *ResponseGeneratorPlugin {
	return &ResponseGeneratorPlugin{
		templates: map[string][]string{
			"greeting": {
				"Hello! How can I help you today?",
				"Hi there! What can I do for you?",
				"Hey! Nice to see you. What's on your mind?",
			},
			"farewell": {
				"Goodbye! Have a great day!",
				"See you later! Take care!",
				"Bye! Feel free to come back anytime!",
			},
			"question": {
				"That's a great question. Let me help you with that.",
				"I understand you're asking about something. Here's what I know.",
				"Good question! Let me provide you with some information.",
			},
			"command": {
				"I'll help you with that right away.",
				"Sure, I can do that for you.",
				"Consider it done!",
			},
			"unknown": {
				"I'm not sure I understand. Could you rephrase that?",
				"Hmm, I didn't quite get that. Can you tell me more?",
				"I'm still learning. Could you explain that differently?",
			},
		},
	}
}

// Execute selects a response template based on intent and fills it with entities and context
func (p *ResponseGeneratorPlugin) Execute(ctx *core.Context) error {
	// Extract intent from context
	var intent Intent
	if intentData, exists := ctx.Get("intent"); exists {
		if i, ok := intentData.(Intent); ok {
			intent = i
		} else {
			intent = Intent{Type: "unknown", Confidence: 0.0}
		}
	} else {
		intent = Intent{Type: "unknown", Confidence: 0.0}
	}

	// Extract entities from context
	var entities []Entity
	if entitiesData, exists := ctx.Get("entities"); exists {
		if e, ok := entitiesData.([]Entity); ok {
			entities = e
		}
	}

	// Select template based on intent
	templates, exists := p.templates[intent.Type]
	if !exists {
		templates = p.templates["unknown"]
	}

	// Select a template (simple: use first one, could be randomized)
	responseText := templates[0]

	// Enhance response with entity information
	if len(entities) > 0 {
		entityInfo := " I noticed you mentioned: "
		for i, entity := range entities {
			if i > 0 {
				entityInfo += ", "
			}
			entityInfo += fmt.Sprintf("%s (%s)", entity.Value, entity.Type)
		}
		responseText += entityInfo
	}

	// Check conversation history for context-aware responses
	if convStateData, exists := ctx.Get("conversation_state"); exists {
		if convState, ok := convStateData.(ConversationState); ok {
			if len(convState.History) > 1 {
				responseText += fmt.Sprintf(" (This is message #%d in our conversation)", len(convState.History))
			}
		}
	}

	// Create response
	response := Response{
		Text:      responseText,
		Intent:    intent,
		Entities:  entities,
		Timestamp: time.Now(),
	}

	// Store response in context
	ctx.SetData(response)

	return nil
}

// PersonalityConfig defines the tone and style for response transformations
type PersonalityConfig struct {
	Name         string
	Emojis       bool
	Casual       bool
	Enthusiastic bool
	Prefix       string
	Suffix       string
}

// PersonalityFilterPlugin applies tone and style transformations to responses
type PersonalityFilterPlugin struct {
	config PersonalityConfig
}

// NewPersonalityFilterPlugin creates a new personality filter with the given configuration
func NewPersonalityFilterPlugin(config PersonalityConfig) *PersonalityFilterPlugin {
	return &PersonalityFilterPlugin{
		config: config,
	}
}

// Execute applies personality transformations to the response text
func (p *PersonalityFilterPlugin) Execute(ctx *core.Context) error {
	// Extract response from context
	response, ok := ctx.GetData().(Response)
	if !ok {
		return fmt.Errorf("expected Response type in context data")
	}

	// Apply personality transformations
	text := response.Text

	// Add prefix if configured
	if p.config.Prefix != "" {
		text = p.config.Prefix + " " + text
	}

	// Make casual if configured
	if p.config.Casual {
		text = strings.ReplaceAll(text, "Hello!", "Hey!")
		text = strings.ReplaceAll(text, "Goodbye!", "Bye!")
		text = strings.ReplaceAll(text, "I will", "I'll")
		text = strings.ReplaceAll(text, "I am", "I'm")
	}

	// Add enthusiasm if configured
	if p.config.Enthusiastic {
		// Add exclamation marks for emphasis
		text = strings.ReplaceAll(text, ".", "!")
		// But don't double up
		text = strings.ReplaceAll(text, "!!", "!")
	}

	// Add emojis if configured
	if p.config.Emojis {
		// Add emojis based on intent
		switch response.Intent.Type {
		case "greeting":
			text += " 👋"
		case "farewell":
			text += " 👋"
		case "question":
			text += " 🤔"
		case "command":
			text += " ✅"
		}
	}

	// Add suffix if configured
	if p.config.Suffix != "" {
		text = text + " " + p.config.Suffix
	}

	// Update response with transformed text
	response.Text = text
	ctx.SetData(response)

	return nil
}
