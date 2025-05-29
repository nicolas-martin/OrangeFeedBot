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
	Summary         string   `json:"summary"`
	MarketImpact    string   `json:"market_impact"`
	Confidence      float64  `json:"confidence"`
	KeyPoints       []string `json:"key_points"`
	AffectedSectors []string `json:"affected_sectors"`
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

		analysis, err := analyzePost(client, content)
		if err != nil {
			fmt.Printf("❌ Analysis failed: %v\n", err)
			continue
		}

		fmt.Printf("📊 Market Impact: %s (Confidence: %.2f)\n", analysis.MarketImpact, analysis.Confidence)
		fmt.Printf("📝 Summary: %s\n", analysis.Summary)
		if len(analysis.KeyPoints) > 0 {
			fmt.Printf("🔑 Key Points: %v\n", analysis.KeyPoints)
		}
		if len(analysis.AffectedSectors) > 0 {
			fmt.Printf("🏭 Affected Sectors: %v\n", analysis.AffectedSectors)
		}
	}
}

func analyzePost(client *openai.Client, content string) (*Analysis, error) {
	prompt := fmt.Sprintf(`
Analyze the following social media post for its potential impact on the stock market:

Post: "%s"

Please provide a JSON response with the following structure:
{
  "summary": "Brief summary of the post content",
  "market_impact": "positive/negative/neutral",
  "confidence": 0.0-1.0,
  "key_points": ["key point 1", "key point 2"],
  "affected_sectors": ["sector 1", "sector 2"]
}

Consider factors like:
- Economic policy mentions
- Trade relations
- Regulatory changes
- Company-specific mentions
- Market sentiment language
- Historical impact of similar statements

Be objective and focus on potential market reactions rather than political opinions.
Respond ONLY with valid JSON, no additional text or explanations.
`, content)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a financial analyst specializing in market sentiment analysis. Provide objective, data-driven analysis of social media posts and their potential market impact. Respond ONLY with valid JSON format.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.3,
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
