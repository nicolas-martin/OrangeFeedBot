package prompts

import "fmt"

// MarketAnalysisPrompt generates a concise but effective prompt for market analysis
func MarketAnalysisPrompt(content string) string {
	return fmt.Sprintf(`Analyze this Trump post for market impact. Respond with ONLY valid JSON:

Post: "%s"

Required JSON format:
{
  "summary": "1 concise sentence (max 80 chars)",
  "market_impact": "bullish/bearish/neutral", 
  "confidence": 0.0-1.0,
  "key_points": ["max 2 brief points"],
  "affected_sectors": ["max 2 sectors"],
  "specific_stocks": ["max 3 ticker symbols"],
  "trading_signal": "buy/sell/hold/watch",
  "time_horizon": "immediate/short-term/medium-term/long-term",
  "risk_level": "low/medium/high",
  "expected_magnitude": "minimal/moderate/significant/major",
  "actionable_insights": ["1 brief trading action (max 60 chars)"]
}

Focus on:
- Direct company/sector mentions
- Policy implications (trade, regulation, rates)
- Specific actionable trades

Be extremely concise. Chat format requires brevity.`, content)
}

// SystemPrompt returns the system prompt for the AI analyst
func SystemPrompt() string {
	return "You are a senior quantitative analyst. Provide ultra-concise market analysis for chat format. Keep all responses brief and actionable. Focus on immediate impact and specific trades."
}
