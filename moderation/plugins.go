package moderation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dvictor357/pipeline-plugin-system/core"
)

// ProfanityFilterPlugin detects inappropriate language in content
type ProfanityFilterPlugin struct {
	profanityWords []string
}

// NewProfanityFilterPlugin creates a new profanity filter with a default word list
func NewProfanityFilterPlugin() *ProfanityFilterPlugin {
	return &ProfanityFilterPlugin{
		profanityWords: []string{
			"badword1", "badword2", "offensive", "inappropriate",
			"profanity", "vulgar", "obscene", "explicit",
		},
	}
}

// Execute checks content for profanity and calculates a score
func (p *ProfanityFilterPlugin) Execute(ctx *core.Context) error {
	content, ok := ctx.GetData().(*Content)
	if !ok {
		return fmt.Errorf("expected *Content, got %T", ctx.GetData())
	}

	text := strings.ToLower(content.Text)
	matchCount := 0

	for _, word := range p.profanityWords {
		if strings.Contains(text, strings.ToLower(word)) {
			matchCount++
		}
	}

	// Calculate score: 0.0 (clean) to 1.0 (highly profane)
	// Each match adds 0.2, capped at 1.0
	score := float64(matchCount) * 0.2
	if score > 1.0 {
		score = 1.0
	}

	ctx.Set("profanity_score", score)
	return nil
}

// SpamDetectorPlugin identifies spam patterns in content
type SpamDetectorPlugin struct {
	linkPattern *regexp.Regexp
}

// NewSpamDetectorPlugin creates a new spam detector
func NewSpamDetectorPlugin() *SpamDetectorPlugin {
	return &SpamDetectorPlugin{
		linkPattern: regexp.MustCompile(`https?://[^\s]+`),
	}
}

// Execute checks content for spam patterns and calculates a score
func (p *SpamDetectorPlugin) Execute(ctx *core.Context) error {
	content, ok := ctx.GetData().(*Content)
	if !ok {
		return fmt.Errorf("expected *Content, got %T", ctx.GetData())
	}

	score := 0.0

	// Check for excessive links
	links := p.linkPattern.FindAllString(content.Text, -1)
	if len(links) > 3 {
		score += 0.5
	} else if len(links) > 1 {
		score += 0.2
	}

	// Check for repeated characters (e.g., "hellooooo")
	// Go's regexp doesn't support backreferences, so check manually
	hasRepeated := false
	runes := []rune(content.Text)
	for i := 0; i < len(runes)-4; i++ {
		if runes[i] == runes[i+1] && runes[i] == runes[i+2] &&
			runes[i] == runes[i+3] && runes[i] == runes[i+4] {
			hasRepeated = true
			break
		}
	}
	if hasRepeated {
		score += 0.3
	}

	// Check for excessive capitalization
	upperCount := 0
	for _, r := range content.Text {
		if r >= 'A' && r <= 'Z' {
			upperCount++
		}
	}
	if len(content.Text) > 0 {
		upperRatio := float64(upperCount) / float64(len(content.Text))
		if upperRatio > 0.5 {
			score += 0.3
		}
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	ctx.Set("spam_score", score)
	return nil
}

// SentimentAnalyzerPlugin performs lexicon-based sentiment analysis
type SentimentAnalyzerPlugin struct {
	positiveWords []string
	negativeWords []string
}

// NewSentimentAnalyzerPlugin creates a new sentiment analyzer
func NewSentimentAnalyzerPlugin() *SentimentAnalyzerPlugin {
	return &SentimentAnalyzerPlugin{
		positiveWords: []string{
			"good", "great", "excellent", "amazing", "wonderful",
			"love", "happy", "fantastic", "awesome", "perfect",
		},
		negativeWords: []string{
			"bad", "terrible", "awful", "horrible", "hate",
			"angry", "sad", "disgusting", "worst", "pathetic",
		},
	}
}

// Execute analyzes sentiment and stores the score
func (p *SentimentAnalyzerPlugin) Execute(ctx *core.Context) error {
	content, ok := ctx.GetData().(*Content)
	if !ok {
		return fmt.Errorf("expected *Content, got %T", ctx.GetData())
	}

	text := strings.ToLower(content.Text)
	words := strings.Fields(text)

	positiveCount := 0
	negativeCount := 0

	for _, word := range words {
		// Remove punctuation for matching
		cleanWord := strings.Trim(word, ".,!?;:")

		for _, posWord := range p.positiveWords {
			if cleanWord == posWord {
				positiveCount++
				break
			}
		}

		for _, negWord := range p.negativeWords {
			if cleanWord == negWord {
				negativeCount++
				break
			}
		}
	}

	// Calculate sentiment: -1.0 (very negative) to 1.0 (very positive)
	totalSentiment := positiveCount - negativeCount
	var sentimentScore float64
	if len(words) > 0 {
		sentimentScore = float64(totalSentiment) / float64(len(words)) * 10
		if sentimentScore > 1.0 {
			sentimentScore = 1.0
		} else if sentimentScore < -1.0 {
			sentimentScore = -1.0
		}
	}

	// Convert extreme negativity to toxicity score (0.0 to 1.0)
	toxicityScore := 0.0
	if sentimentScore < -0.3 {
		toxicityScore = -sentimentScore // More negative = higher toxicity
		if toxicityScore > 1.0 {
			toxicityScore = 1.0
		}
	}

	ctx.Set("sentiment_score", sentimentScore)
	ctx.Set("toxicity_score", toxicityScore)
	return nil
}

// ScoringPlugin aggregates scores from previous plugins
type ScoringPlugin struct {
	profanityWeight float64
	spamWeight      float64
	toxicityWeight  float64
}

// NewScoringPlugin creates a new scoring plugin with default weights
func NewScoringPlugin() *ScoringPlugin {
	return &ScoringPlugin{
		profanityWeight: 0.4,
		spamWeight:      0.3,
		toxicityWeight:  0.3,
	}
}

// Execute calculates the weighted overall moderation score
func (p *ScoringPlugin) Execute(ctx *core.Context) error {
	// Retrieve individual scores
	profanityScore := 0.0
	if val, ok := ctx.Get("profanity_score"); ok {
		if score, ok := val.(float64); ok {
			profanityScore = score
		}
	}

	spamScore := 0.0
	if val, ok := ctx.Get("spam_score"); ok {
		if score, ok := val.(float64); ok {
			spamScore = score
		}
	}

	toxicityScore := 0.0
	if val, ok := ctx.Get("toxicity_score"); ok {
		if score, ok := val.(float64); ok {
			toxicityScore = score
		}
	}

	// Calculate weighted overall score
	overallScore := (profanityScore * p.profanityWeight) +
		(spamScore * p.spamWeight) +
		(toxicityScore * p.toxicityWeight)

	// Create ModerationScore struct
	moderationScore := ModerationScore{
		ProfanityScore: profanityScore,
		SpamScore:      spamScore,
		ToxicityScore:  toxicityScore,
		OverallScore:   overallScore,
	}

	ctx.Set("moderation_score", moderationScore)
	return nil
}

// DecisionRouterPlugin makes moderation decisions based on score thresholds
type DecisionRouterPlugin struct {
	approveThreshold float64
	reviewThreshold  float64
}

// NewDecisionRouterPlugin creates a new decision router with default thresholds
func NewDecisionRouterPlugin() *DecisionRouterPlugin {
	return &DecisionRouterPlugin{
		approveThreshold: ApproveThreshold,
		reviewThreshold:  ReviewThreshold,
	}
}

// Execute determines the moderation action based on the overall score
func (p *DecisionRouterPlugin) Execute(ctx *core.Context) error {
	// Retrieve moderation score
	scoreVal, ok := ctx.Get("moderation_score")
	if !ok {
		return fmt.Errorf("moderation_score not found in context")
	}

	moderationScore, ok := scoreVal.(ModerationScore)
	if !ok {
		return fmt.Errorf("expected ModerationScore, got %T", scoreVal)
	}

	// Determine action based on thresholds
	var action string
	var reason string
	var flagged bool

	if moderationScore.OverallScore < p.approveThreshold {
		action = "approve"
		reason = "Content meets quality standards"
		flagged = false
	} else if moderationScore.OverallScore < p.reviewThreshold {
		action = "review"
		reason = "Content requires manual review"
		flagged = true
	} else {
		action = "reject"
		reason = "Content violates community guidelines"
		flagged = true
	}

	// Create decision
	decision := ModerationDecision{
		Action:  action,
		Score:   moderationScore,
		Reason:  reason,
		Flagged: flagged,
	}

	ctx.Set("moderation_decision", decision)
	return nil
}

// ActionHandlerPlugin executes the moderation decision
type ActionHandlerPlugin struct{}

// NewActionHandlerPlugin creates a new action handler
func NewActionHandlerPlugin() *ActionHandlerPlugin {
	return &ActionHandlerPlugin{}
}

// Execute logs the decision and updates the final result
func (p *ActionHandlerPlugin) Execute(ctx *core.Context) error {
	// Retrieve content
	content, ok := ctx.GetData().(*Content)
	if !ok {
		return fmt.Errorf("expected *Content, got %T", ctx.GetData())
	}

	// Retrieve decision
	decisionVal, ok := ctx.Get("moderation_decision")
	if !ok {
		return fmt.Errorf("moderation_decision not found in context")
	}

	decision, ok := decisionVal.(ModerationDecision)
	if !ok {
		return fmt.Errorf("expected ModerationDecision, got %T", decisionVal)
	}

	// Create final result
	result := ModerationResult{
		Content:  *content,
		Decision: decision,
	}

	// Update context with final result
	ctx.SetData(&result)

	return nil
}
