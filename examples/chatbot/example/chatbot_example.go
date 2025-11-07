package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dvictor357/pipeline-plugin-system/chatbot"
	"github.com/dvictor357/pipeline-plugin-system/core"
)

func main() {
	fmt.Println("=== Chat Bot Pipeline Example ===")

	// Create the chat bot pipeline with all plugins
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

	// Simulate a conversation with multiple messages
	sessionID := "session-123"
	userID := "user-456"

	messages := []string{
		"Hello! How are you?",
		"What's the weather like on December 25th?",
		"Can you help me with my order?",
		"My email is john.doe@example.com and phone is 555-123-4567",
		"Thanks! Goodbye!",
	}

	fmt.Println("Starting conversation...")

	for i, text := range messages {
		fmt.Printf("--- Message %d ---\n", i+1)
		fmt.Printf("User: %s\n", text)

		// Create message
		msg := chatbot.Message{
			Text:      text,
			UserID:    userID,
			SessionID: sessionID,
			Timestamp: time.Now(),
		}

		// Create context with the message
		ctx := core.NewContext(msg)

		// Execute pipeline
		err := pipeline.Execute(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		// Extract response
		response, ok := ctx.GetData().(chatbot.Response)
		if !ok {
			fmt.Println("Error: unexpected response type")
			continue
		}

		// Display response
		fmt.Printf("Bot: %s\n", response.Text)

		// Display intent
		fmt.Printf("Intent: %s (confidence: %.2f)\n", response.Intent.Type, response.Intent.Confidence)

		// Display entities if any
		if len(response.Entities) > 0 {
			fmt.Println("Entities detected:")
			for _, entity := range response.Entities {
				fmt.Printf("  - %s: %s\n", entity.Type, entity.Value)
			}
		}

		// Display conversation state
		if convStateData, exists := ctx.Get("conversation_state"); exists {
			if convState, ok := convStateData.(chatbot.ConversationState); ok {
				fmt.Printf("Conversation history: %d messages\n", len(convState.History))
			}
		}

		fmt.Println()
	}

	fmt.Println("=== Conversation Complete ===")

	// Demonstrate JSON serialization
	fmt.Println("=== JSON Response Example ===")
	msg := chatbot.Message{
		Text:      "Show me John Smith's account details",
		UserID:    userID,
		SessionID: sessionID,
		Timestamp: time.Now(),
	}

	ctx := core.NewContext(msg)
	err := pipeline.Execute(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	response, ok := ctx.GetData().(chatbot.Response)
	if ok {
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
	}
}
