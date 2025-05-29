package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"orangefeed/internal/prompts"
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

	fmt.Println("🎯 OrangeFeed - Truth Social Real Data Scraper with Enhanced AI")
	fmt.Println(strings.Repeat("=", 70))

	// Get credentials
	username := os.Getenv("TRUTHSOCIAL_USERNAME")
	password := os.Getenv("TRUTHSOCIAL_PASSWORD")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if username == "" || password == "" {
		log.Fatal("❌ Missing TRUTHSOCIAL_USERNAME or TRUTHSOCIAL_PASSWORD environment variables")
	}

	fmt.Printf("✅ Credentials loaded for user: %s\n", username)
	fmt.Println("🔐 Using CycleTLS Chrome fingerprint spoofing to bypass Cloudflare")

	// Create Truth Social client with CycleTLS
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("\n🌐 Creating Truth Social client with CycleTLS...")
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
		fmt.Println("\n🤖 Analyzing posts with Enhanced Concise AI...")
		analyzePostsWithAI(statuses[:min(3, len(statuses))], openaiKey) // Analyze first 3 posts
	} else {
		fmt.Println("\n⚠️ No OPENAI_API_KEY found - skipping AI analysis")
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎉 Success! Real data extraction working with CycleTLS fingerprint spoofing!")
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

		// Display ultra-concise analysis (chat format)
		fmt.Printf("\n🚨 %s (%.0f%%) | %s %s | %s risk\n",
			strings.ToUpper(analysis.MarketImpact),
			analysis.Confidence*100,
			getSignalEmoji(analysis.TradingSignal),
			strings.ToUpper(analysis.TradingSignal),
			strings.ToUpper(analysis.RiskLevel))

		fmt.Printf("🏭 %s | 📈 %s\n",
			formatList(analysis.AffectedSectors, 2),
			formatList(analysis.SpecificStocks, 3))

		fmt.Printf("💡 %s\n", analysis.Summary)

		if len(analysis.ActionableInsights) > 0 && len(analysis.ActionableInsights[0]) > 0 {
			fmt.Printf("⚡ %s\n", analysis.ActionableInsights[0])
		}

		fmt.Printf("👍 %d | 🔄 %d\n", status.FavouritesCount, status.ReblogsCount)
		fmt.Printf("\n" + strings.Repeat("-", 40))
	}
}

// Helper function to get emoji for trading signal
func getSignalEmoji(signal string) string {
	switch strings.ToLower(signal) {
	case "buy":
		return "🟢"
	case "sell":
		return "🔴"
	case "hold":
		return "🟡"
	case "watch":
		return "👀"
	default:
		return "⚪"
	}
}

// Helper function to format lists concisely
func formatList(items []string, maxItems int) string {
	if len(items) == 0 {
		return "None"
	}

	if len(items) <= maxItems {
		return strings.Join(items, ", ")
	}

	return strings.Join(items[:maxItems], ", ") + fmt.Sprintf(" +%d", len(items)-maxItems)
}

func analyzePost(client *openai.Client, content string) (*Analysis, error) {
	resp, err := client.CreateChatCompletion(
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
