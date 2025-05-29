package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	prompt := fmt.Sprintf(`
You are a senior quantitative analyst at a top-tier investment bank. Analyze the following social media post from Donald Trump for its concrete impact on the stock market and provide specific trading recommendations.

Post: "%s"

Provide a detailed JSON response with the following structure:
{
  "summary": "Brief summary of the post content and its market implications",
  "market_impact": "bullish/bearish/neutral",
  "confidence": 0.0-1.0,
  "key_points": ["specific market-moving elements"],
  "affected_sectors": ["Technology", "Healthcare", "Energy", etc.],
  "specific_stocks": ["AAPL", "TSLA", "JPM", etc. - actual ticker symbols"],
  "trading_signal": "buy/sell/hold/watch",
  "time_horizon": "immediate/short-term/medium-term/long-term",
  "risk_level": "low/medium/high",
  "expected_magnitude": "minimal/moderate/significant/major",
  "actionable_insights": ["specific trading recommendations with reasoning"]
}

Analysis Guidelines:
1. **Specific Stocks**: Identify actual ticker symbols that would be directly affected
2. **Trading Signal**: Provide clear buy/sell/hold/watch recommendations
3. **Time Horizon**: 
   - immediate (0-24 hours)
   - short-term (1-7 days)
   - medium-term (1-4 weeks)
   - long-term (1+ months)
4. **Expected Magnitude**: Quantify the expected market movement
5. **Actionable Insights**: Provide specific trading strategies, entry/exit points, risk management

Consider these factors:
- Direct company mentions or implications
- Policy changes affecting specific industries
- Trade relations and tariff impacts
- Regulatory changes and their sector effects
- Economic policy shifts
- Geopolitical implications
- Historical market reactions to similar statements
- Current market conditions and sentiment
- Sector rotation opportunities
- Options strategies if appropriate

Be specific and actionable. If the post has minimal market impact, state that clearly.
Respond ONLY with valid JSON, no additional text.
`, content)

	resp, err := ma.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a senior quantitative analyst at Goldman Sachs with 15+ years of experience in political risk analysis and market impact assessment. You specialize in translating political events and statements into actionable trading strategies. Your analysis has consistently generated alpha for institutional clients. Provide concrete, specific, and actionable market analysis.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.2,  // Lower temperature for more consistent analysis
			MaxTokens:   1500, // Allow for more detailed responses
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content = resp.Choices[0].Message.Content

	// Try to extract JSON from the response
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}") + 1

	if jsonStart == -1 || jsonEnd == 0 {
		return nil, fmt.Errorf("no JSON found in response: %s", content)
	}

	jsonContent := content[jsonStart:jsonEnd]

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
