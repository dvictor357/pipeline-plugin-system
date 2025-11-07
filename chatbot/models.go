package chatbot

import "time"

// Message represents an input message from a user
type Message struct {
	Text      string    `json:"text"`
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

// Intent represents the classification result of a user's message
type Intent struct {
	Type       string  `json:"type"`       // greeting, question, command, farewell, etc.
	Confidence float64 `json:"confidence"` // confidence score between 0.0 and 1.0
}

// Entity represents an extracted piece of information from a message
type Entity struct {
	Type  string `json:"type"`  // person, date, location, number, etc.
	Value string `json:"value"` // the extracted value
	Start int    `json:"start"` // start position in the text
	End   int    `json:"end"`   // end position in the text
}

// Response represents the bot's response to a user message
type Response struct {
	Text      string    `json:"text"`
	Intent    Intent    `json:"intent"`
	Entities  []Entity  `json:"entities"`
	Timestamp time.Time `json:"timestamp"`
}

// ConversationState maintains state across multiple message exchanges
type ConversationState struct {
	History    []Message      `json:"history"`
	UserPrefs  map[string]any `json:"user_prefs"`
	LastIntent Intent         `json:"last_intent"`
}
