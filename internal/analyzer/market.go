package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"orangefeed/internal/prompts"

	"github.com/sashabaranov/go-openai"
)

type Analysis struct {
	Summary            string   `json:"summary"`
	MarketImpact       string   `json:"market_impact"` // "bullish", "bearish", "neutral"
	Confidence         float64  `json:"confidence"`    // 0.0-1.0
	KeyPoints          []string `json:"key_points"`
	AffectedSectors    []string `json:"affected_sectors"`
	SpecificStocks     []string `json:"specific_stocks"`     // Ticker symbols mentioned or implied
	TradingSignal      string   `json:"trading_signal"`      // "buy", "sell", "hold", "watch"
	TimeHorizon        string   `json:"time_horizon"`        // "immediate", "short-term", "medium-term", "long-term"
	RiskLevel          string   `json:"risk_level"`          // "low", "medium", "high"
	ExpectedMagnitude  string   `json:"expected_magnitude"`  // "minimal", "moderate", "significant", "major"
	ActionableInsights []string `json:"actionable_insights"` // Specific trading recommendations
}

type MarketAnalyzer struct {
	openaiClient *openai.Client
}

func NewMarketAnalyzer(openaiKey string) *MarketAnalyzer {
	return &MarketAnalyzer{
		openaiClient: openai.NewClient(openaiKey),
	}
}

func (ma *MarketAnalyzer) AnalyzePost(content string) (*Analysis, error) {
	resp, err := ma.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompts.SystemPrompt(),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompts.MarketAnalysisPrompt(content),
				},
			},
			Temperature: 0.2, // Lower temperature for more consistent analysis
			MaxTokens:   800, // Reduced for more concise responses
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	responseContent := resp.Choices[0].Message.Content

	// Try to extract JSON from the response
	jsonStart := strings.Index(responseContent, "{")
	jsonEnd := strings.LastIndex(responseContent, "}") + 1

	if jsonStart == -1 || jsonEnd == 0 {
		return nil, fmt.Errorf("no JSON found in response: %s", responseContent)
	}

	jsonContent := responseContent[jsonStart:jsonEnd]

	var analysis Analysis
	if err := json.Unmarshal([]byte(jsonContent), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w", err)
	}

	return &analysis, nil
}

// AnalyzeBatch analyzes multiple posts and returns aggregated insights
func (ma *MarketAnalyzer) AnalyzeBatch(contents []string) ([]*Analysis, error) {
	var analyses []*Analysis

	for _, content := range contents {
		if len(strings.TrimSpace(content)) < 10 {
			continue // Skip very short content
		}

		analysis, err := ma.AnalyzePost(content)
		if err != nil {
			// Log error but continue with other posts
			continue
		}

		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// GetMarketSentiment provides an overall market sentiment based on recent analyses
func (ma *MarketAnalyzer) GetMarketSentiment(analyses []*Analysis) string {
	if len(analyses) == 0 {
		return "neutral"
	}

	bullishCount := 0
	bearishCount := 0

	for _, analysis := range analyses {
		switch strings.ToLower(analysis.MarketImpact) {
		case "bullish":
			bullishCount++
		case "bearish":
			bearishCount++
		}
	}

	if bullishCount > bearishCount {
		return "bullish"
	} else if bearishCount > bullishCount {
		return "bearish"
	}

	return "neutral"
}
