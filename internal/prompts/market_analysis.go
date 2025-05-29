package prompts

import "fmt"

// MarketAnalysisPrompt generates a concise but effective prompt for market analysis
func MarketAnalysisPrompt(content string) string {
	return fmt.Sprintf(`Analyze this Trump post for market impact. Respond with ONLY valid JSON:

Post: "%s"

Required JSON format:
{
  "summary": "1-sentence market impact summary",
  "market_impact": "bullish/bearish/neutral", 
  "confidence": 0.0-1.0,
  "key_points": ["max 3 key market drivers"],
  "affected_sectors": ["max 3 sectors"],
  "specific_stocks": ["max 5 ticker symbols"],
  "trading_signal": "buy/sell/hold/watch",
  "time_horizon": "immediate/short-term/medium-term/long-term",
  "risk_level": "low/medium/high",
  "expected_magnitude": "minimal/moderate/significant/major",
  "actionable_insights": ["max 2 specific trading actions"]
}

Focus on:
- Direct company/sector mentions
- Policy implications (trade, regulation, rates)
- Historical market reactions to similar statements
- Specific actionable trades

Be concise but precise. If minimal impact, state clearly.`, content)
}

// SystemPrompt returns the system prompt for the AI analyst
func SystemPrompt() string {
	return "You are a senior quantitative analyst at Goldman Sachs. Provide concise, actionable market analysis with specific trading recommendations. Focus on immediate market impact and concrete trading opportunities."
}
