package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"orangefeed/internal/truthsocial"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type TruthSocialPost struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	URL       string    `json:"url"`
}

type MarketAnalysis struct {
	Summary         string   `json:"summary"`
	MarketImpact    string   `json:"market_impact"` // "positive", "negative", "neutral"
	Confidence      float64  `json:"confidence"`
	KeyPoints       []string `json:"key_points"`
	AffectedSectors []string `json:"affected_sectors"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get credentials
	truthUsername := os.Getenv("TRUTHSOCIAL_USERNAME")
	truthPassword := os.Getenv("TRUTHSOCIAL_PASSWORD")

	if truthUsername == "" || truthPassword == "" {
		log.Fatal("TRUTHSOCIAL_USERNAME and TRUTHSOCIAL_PASSWORD environment variables required")
	}

	// Try to fetch real posts
	fmt.Println("🔍 Attempting to fetch real Truth Social posts from @realDonaldTrump...")
	posts, err := fetchUserPosts("realDonaldTrump", 1, truthUsername, truthPassword)

	if err != nil {
		fmt.Printf("❌ Error fetching posts: %v\n", err)
		return
	}

	if len(posts) == 0 {
		fmt.Println("❌ No posts found")
		return
	}

	// If we got posts, analyze them
	fmt.Printf("✅ Found %d posts! Analyzing the latest one...\n\n", len(posts))

	latestPost := posts[0]
	fmt.Printf("📄 Latest post details:\n")
	fmt.Printf("   ID: %s\n", latestPost.ID)
	fmt.Printf("   Content: %s\n", latestPost.Content)
	fmt.Printf("   Timestamp: %s\n", latestPost.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   URL: %s\n\n", latestPost.URL)

	// Initialize OpenAI client
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	openaiClient := openai.NewClient(openaiKey)

	// Analyze the post
	fmt.Println("🤖 Analyzing post with AI...")
	analysis, err := analyzePost(openaiClient, latestPost)
	if err != nil {
		log.Fatalf("Error analyzing post: %v", err)
	}

	// Display analysis results
	fmt.Printf("\n🎯 AI Analysis Results:\n")
	fmt.Printf("📊 Market Impact: %s\n", analysis.MarketImpact)
	fmt.Printf("🎯 Confidence: %.2f\n", analysis.Confidence)
	fmt.Printf("🔑 Key Points:\n")
	for _, point := range analysis.KeyPoints {
		fmt.Printf("   • %s\n", point)
	}
	fmt.Printf("🏭 Affected Sectors:\n")
	for _, sector := range analysis.AffectedSectors {
		fmt.Printf("   • %s\n", sector)
	}
	fmt.Printf("📝 Summary: %s\n", analysis.Summary)
}

func fetchUserPosts(username string, limit int, truthUsername, truthPassword string) ([]TruthSocialPost, error) {
	// Remove @ if present
	username = strings.TrimPrefix(username, "@")

	fmt.Printf("🔍 Fetching Truth Social posts from @%s using Go client...\n", username)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Truth Social client
	fmt.Println("🔐 Authenticating with Truth Social...")
	client, err := truthsocial.NewClient(ctx, truthUsername, truthPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Look up the account to get the account ID
	fmt.Printf("👤 Looking up account @%s...\n", username)
	account, err := client.GetAccount(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find account @%s: %w", username, err)
	}

	fmt.Printf("✅ Found account: %s (ID: %s)\n", account.Username, account.ID)

	// Fetch statuses
	fmt.Printf("📄 Fetching up to %d posts...\n", limit)
	statuses, err := client.GetStatuses(ctx, account.ID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch statuses: %w", err)
	}

	// Convert to our format
	var posts []TruthSocialPost
	for _, status := range statuses {
		post := convertStatus(status, username)
		if post != nil {
			posts = append(posts, *post)
		}
	}

	fmt.Printf("✅ Successfully fetched %d posts!\n", len(posts))
	return posts, nil
}

func convertStatus(status truthsocial.Status, username string) *TruthSocialPost {
	// Clean HTML content
	content := cleanHTMLContent(status.Content)
	if content == "" {
		return nil
	}

	// Parse timestamp
	createdAt, err := time.Parse(time.RFC3339, status.CreatedAt)
	if err != nil {
		// Try alternative format
		createdAt, err = time.Parse("2006-01-02T15:04:05.000Z", status.CreatedAt)
		if err != nil {
			createdAt = time.Now() // Fallback
		}
	}

	// Create URL if not provided
	url := status.URL
	if url == "" {
		url = fmt.Sprintf("https://truthsocial.com/@%s/%s", username, status.ID)
	}

	return &TruthSocialPost{
		ID:        status.ID,
		Content:   content,
		CreatedAt: createdAt,
		Username:  username,
		URL:       url,
	}
}

func cleanHTMLContent(content string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(content, "")

	// Decode HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	return strings.TrimSpace(text)
}

func analyzePost(client *openai.Client, post TruthSocialPost) (*MarketAnalysis, error) {
	prompt := fmt.Sprintf(`
Analyze the following social media post from Donald Trump for its potential impact on the stock market:

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
`, post.Content)

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

	content := resp.Choices[0].Message.Content
	fmt.Printf("Raw OpenAI response: %s\n\n", content)

	// Try to extract JSON from the response if it contains extra text
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}") + 1

	if jsonStart == -1 || jsonEnd == 0 {
		return nil, fmt.Errorf("no JSON found in response: %s", content)
	}

	jsonContent := content[jsonStart:jsonEnd]
	fmt.Printf("Extracted JSON: %s\n\n", jsonContent)

	var analysis MarketAnalysis
	if err := json.Unmarshal([]byte(jsonContent), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w\nContent: %s", err, jsonContent)
	}

	return &analysis, nil
}
