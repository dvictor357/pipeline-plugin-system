package moderation

import "time"

// Threshold constants for moderation decisions
const (
	ApproveThreshold = 0.3 // Below this: auto-approve
	ReviewThreshold  = 0.7 // Between approve and review: flag for review
	// Above ReviewThreshold: auto-reject
)

// Content represents user-generated content to be moderated
type Content struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	AuthorID  string    `json:"author_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ModerationScore contains scores from various moderation checks
type ModerationScore struct {
	ProfanityScore float64 `json:"profanity_score"`
	SpamScore      float64 `json:"spam_score"`
	ToxicityScore  float64 `json:"toxicity_score"`
	OverallScore   float64 `json:"overall_score"`
}

// ModerationDecision represents the moderation decision for content
type ModerationDecision struct {
	Action  string          `json:"action"` // approve, review, reject
	Score   ModerationScore `json:"score"`
	Reason  string          `json:"reason"`
	Flagged bool            `json:"flagged"`
}

// ModerationResult is the final result of the moderation pipeline
type ModerationResult struct {
	Content  Content            `json:"content"`
	Decision ModerationDecision `json:"decision"`
}
