package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"orangefeed/internal/truthsocial"

	"github.com/joho/godotenv"
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

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	fmt.Println("🎯 OrangeFeed - Truth Social Real Data Scraper with uTLS")
	fmt.Println(strings.Repeat("=", 70))

	// Get credentials
	username := os.Getenv("TRUTHSOCIAL_USERNAME")
	password := os.Getenv("TRUTHSOCIAL_PASSWORD")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if username == "" || password == "" {
		log.Fatal("❌ Missing TRUTHSOCIAL_USERNAME or TRUTHSOCIAL_PASSWORD environment variables")
	}

	fmt.Printf("✅ Credentials loaded for user: %s\n", username)
	fmt.Println("🔐 Using uTLS Chrome fingerprint spoofing to bypass Cloudflare")

	// Create Truth Social client with uTLS
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("\n🌐 Creating Truth Social client with uTLS...")
	client, err := truthsocial.NewClient(ctx, username, password)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	fmt.Println("✅ Authentication successful!")

	// Test user lookup
	fmt.Println("\n👤 Looking up @realDonaldTrump...")
	account, err := client.Lookup(ctx, "realDonaldTrump")
	if err != nil {
		log.Fatalf("❌ Failed to lookup user: %v", err)
	}

	fmt.Printf("✅ Found user: %s (@%s)\n", account.DisplayName, account.Username)
	fmt.Printf("   📊 Followers: %d\n", account.FollowersCount)
	fmt.Printf("   📝 Posts: %d\n", account.StatusesCount)
	fmt.Printf("   ✓ Verified: %t\n", account.Verified)

	// Fetch recent posts using PullStatuses (same as Truthbrush)
	fmt.Println("\n📄 Fetching recent posts using PullStatuses method...")
	statuses, err := client.PullStatuses(ctx, "realDonaldTrump", true, 20) // exclude replies, limit 20
	if err != nil {
		log.Fatalf("❌ Failed to fetch posts: %v", err)
	}

	fmt.Printf("✅ Successfully fetched %d posts!\n", len(statuses))

	if len(statuses) == 0 {
		fmt.Println("⚠️ No posts found")
		return
	}

	// Display first few posts
	fmt.Println("\n📋 Recent Posts:")
	for i, status := range statuses {
		if i >= 5 { // Show first 5 posts
			break
		}

		// Clean content (remove HTML tags)
		content := cleanContent(status.Content)
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		fmt.Printf("\n%d. Post ID: %s\n", i+1, status.ID)
		fmt.Printf("   📅 Created: %s\n", status.CreatedAt)
		fmt.Printf("   📝 Content: %s\n", content)
		fmt.Printf("   🔗 URL: %s\n", status.URL)
		fmt.Printf("   👍 Likes: %d | 🔄 Reblogs: %d\n", status.FavouritesCount, status.ReblogsCount)
	}

	// Analyze posts with AI if OpenAI key is available
	if openaiKey != "" {
		fmt.Println("\n🤖 Analyzing posts with AI...")
		analyzePostsWithAI(statuses[:min(3, len(statuses))], openaiKey) // Analyze first 3 posts
	} else {
		fmt.Println("\n⚠️ No OPENAI_API_KEY found - skipping AI analysis")
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎉 Success! Real data extraction working with uTLS fingerprint spoofing!")
	fmt.Printf("📊 Total posts fetched: %d\n", len(statuses))
	fmt.Println("🔧 This proves our implementation bypasses Cloudflare protection")
}

func cleanContent(content string) string {
	// Simple HTML tag removal
	content = strings.ReplaceAll(content, "<p>", "")
	content = strings.ReplaceAll(content, "</p>", " ")
	content = strings.ReplaceAll(content, "<br>", " ")
	content = strings.ReplaceAll(content, "<br/>", " ")
	content = strings.ReplaceAll(content, "<br />", " ")

	// Remove other common HTML tags
	for _, tag := range []string{"<a", "</a>", "<span", "</span>", "<div", "</div>"} {
		if strings.Contains(content, tag) {
			// Simple tag removal - find opening and closing
			for strings.Contains(content, "<") && strings.Contains(content, ">") {
				start := strings.Index(content, "<")
				end := strings.Index(content[start:], ">")
				if end == -1 {
					break
				}
				content = content[:start] + content[start+end+1:]
			}
		}
	}

	return strings.TrimSpace(content)
}

func analyzePostsWithAI(statuses []truthsocial.Status, openaiKey string) {
	client := openai.NewClient(openaiKey)

	for i, status := range statuses {
		content := cleanContent(status.Content)
		if len(content) < 10 { // Skip very short posts
			continue
		}

		fmt.Printf("\n🔍 Analyzing Post %d...\n", i+1)
		fmt.Printf("📝 Content: %s\n", content)

		analysis, err := analyzePost(client, content)
		if err != nil {
			fmt.Printf("❌ Analysis failed: %v\n", err)
			continue
		}

		// Display comprehensive analysis
		fmt.Printf("\n📊 MARKET ANALYSIS RESULTS:\n")
		fmt.Printf("   🎯 Market Impact: %s (Confidence: %.1f%%)\n", strings.ToUpper(analysis.MarketImpact), analysis.Confidence*100)
		fmt.Printf("   📈 Trading Signal: %s\n", strings.ToUpper(analysis.TradingSignal))
		fmt.Printf("   ⏰ Time Horizon: %s\n", analysis.TimeHorizon)
		fmt.Printf("   ⚠️  Risk Level: %s\n", strings.ToUpper(analysis.RiskLevel))
		fmt.Printf("   📏 Expected Magnitude: %s\n", analysis.ExpectedMagnitude)

		fmt.Printf("\n📝 Summary: %s\n", analysis.Summary)

		if len(analysis.KeyPoints) > 0 {
			fmt.Printf("\n🔑 Key Market-Moving Points:\n")
			for _, point := range analysis.KeyPoints {
				fmt.Printf("   • %s\n", point)
			}
		}

		if len(analysis.AffectedSectors) > 0 {
			fmt.Printf("\n🏭 Affected Sectors: %s\n", strings.Join(analysis.AffectedSectors, ", "))
		}

		if len(analysis.SpecificStocks) > 0 {
			fmt.Printf("\n📈 Specific Stocks to Watch: %s\n", strings.Join(analysis.SpecificStocks, ", "))
		}

		if len(analysis.ActionableInsights) > 0 {
			fmt.Printf("\n💡 ACTIONABLE TRADING INSIGHTS:\n")
			for j, insight := range analysis.ActionableInsights {
				fmt.Printf("   %d. %s\n", j+1, insight)
			}
		}

		fmt.Printf("\n" + strings.Repeat("-", 60))
	}
}

func analyzePost(client *openai.Client, content string) (*Analysis, error) {
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

	resp, err := client.CreateChatCompletion(
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
