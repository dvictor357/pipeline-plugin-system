package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dvictor357/pipeline-plugin-system/core"
	"github.com/dvictor357/pipeline-plugin-system/moderation"
)

func main() {
	fmt.Println("=== Content Moderation Pipeline Example ===")

	// Create the content moderation pipeline with all plugins
	pipeline := core.NewPipeline(core.AbortOnError).
		Use(moderation.NewProfanityFilterPlugin()).
		Use(moderation.NewSpamDetectorPlugin()).
		Use(moderation.NewSentimentAnalyzerPlugin()).
		Use(moderation.NewScoringPlugin()).
		Use(moderation.NewDecisionRouterPlugin()).
		Use(moderation.NewActionHandlerPlugin())

	// Test various content samples demonstrating different scenarios
	testCases := []struct {
		name           string
		content        moderation.Content
		expectedAction string
	}{
		{
			name: "Clean Content (Approve)",
			content: moderation.Content{
				ID:        "content-001",
				Text:      "This is a great product! I love it and would recommend it to everyone.",
				AuthorID:  "user-123",
				Timestamp: time.Now(),
			},
			expectedAction: "approve",
		},
		{
			name: "Mild Issues (Review)",
			content: moderation.Content{
				ID:        "content-002",
				Text:      "This is terrible and awful. I hate this product so much.",
				AuthorID:  "user-456",
				Timestamp: time.Now(),
			},
			expectedAction: "review",
		},
		{
			name: "Profanity (Reject)",
			content: moderation.Content{
				ID:        "content-003",
				Text:      "This is badword1 and offensive content that is inappropriate and vulgar.",
				AuthorID:  "user-789",
				Timestamp: time.Now(),
			},
			expectedAction: "reject",
		},
		{
			name: "Spam Pattern (Reject)",
			content: moderation.Content{
				ID:        "content-004",
				Text:      "BUY NOW!!! CLICK HERE https://spam.com https://scam.com https://fake.com https://bad.com HELLOOOOOOO!!!",
				AuthorID:  "user-spam",
				Timestamp: time.Now(),
			},
			expectedAction: "reject",
		},
		{
			name: "Borderline Negative (Review)",
			content: moderation.Content{
				ID:        "content-005",
				Text:      "Not the best experience. Could be better. Somewhat disappointing.",
				AuthorID:  "user-critic",
				Timestamp: time.Now(),
			},
			expectedAction: "review",
		},
		{
			name: "Positive Content (Approve)",
			content: moderation.Content{
				ID:        "content-006",
				Text:      "Excellent service! Amazing quality and fantastic support. Very happy with my purchase.",
				AuthorID:  "user-happy",
				Timestamp: time.Now(),
			},
			expectedAction: "approve",
		},
	}

	fmt.Println("\nProcessing content samples...")

	for i, tc := range testCases {
		fmt.Printf("--- Test Case %d: %s ---\n", i+1, tc.name)
		fmt.Printf("Content ID: %s\n", tc.content.ID)
		fmt.Printf("Author: %s\n", tc.content.AuthorID)
		fmt.Printf("Text: %s\n", tc.content.Text)

		// Create context with the content
		ctx := core.NewContext(&tc.content)

		// Execute pipeline
		err := pipeline.Execute(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		// Extract result
		result, ok := ctx.GetData().(*moderation.ModerationResult)
		if !ok {
			fmt.Println("Error: unexpected result type")
			continue
		}

		// Display decision
		fmt.Printf("\nDecision: %s\n", result.Decision.Action)
		fmt.Printf("Flagged: %v\n", result.Decision.Flagged)
		fmt.Printf("Reason: %s\n", result.Decision.Reason)

		// Display scores
		fmt.Println("\nScores:")
		fmt.Printf("  Profanity: %.2f\n", result.Decision.Score.ProfanityScore)
		fmt.Printf("  Spam: %.2f\n", result.Decision.Score.SpamScore)
		fmt.Printf("  Toxicity: %.2f\n", result.Decision.Score.ToxicityScore)
		fmt.Printf("  Overall: %.2f\n", result.Decision.Score.OverallScore)

		// Verify expected action
		if result.Decision.Action == tc.expectedAction {
			fmt.Printf("\n✓ Expected action: %s\n", tc.expectedAction)
		} else {
			fmt.Printf("\n✗ Expected %s, got %s\n", tc.expectedAction, result.Decision.Action)
		}

		fmt.Println()
	}

	fmt.Println("=== Moderation Complete ===")

	// Demonstrate JSON serialization
	fmt.Println("\n=== JSON Result Example ===")
	content := moderation.Content{
		ID:        "content-json",
		Text:      "This product is obscene and profanity filled with badword2 content.",
		AuthorID:  "user-json",
		Timestamp: time.Now(),
	}

	ctx := core.NewContext(&content)
	err := pipeline.Execute(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	result, ok := ctx.GetData().(*moderation.ModerationResult)
	if ok {
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
	}

	// Demonstrate scoring and decision-making process
	fmt.Println("\n=== Scoring and Decision-Making Process ===")
	fmt.Println("\nThresholds:")
	fmt.Printf("  Approve: < %.2f\n", moderation.ApproveThreshold)
	fmt.Printf("  Review: %.2f - %.2f\n", moderation.ApproveThreshold, moderation.ReviewThreshold)
	fmt.Printf("  Reject: > %.2f\n", moderation.ReviewThreshold)
	fmt.Println("\nThe pipeline processes content through multiple stages:")
	fmt.Println("  1. Profanity Filter - Detects inappropriate language")
	fmt.Println("  2. Spam Detector - Identifies spam patterns")
	fmt.Println("  3. Sentiment Analyzer - Determines content sentiment and toxicity")
	fmt.Println("  4. Scoring Plugin - Aggregates weighted scores")
	fmt.Println("  5. Decision Router - Makes moderation decision based on thresholds")
	fmt.Println("  6. Action Handler - Executes the moderation action")
}
