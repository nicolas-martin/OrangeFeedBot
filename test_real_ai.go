package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"orangefeed/internal/scraper"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

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

	// Initialize scraper
	s := scraper.NewTruthSocialScraper()

	// Fetch real posts
	fmt.Println("Fetching real Truth Social posts...")
	posts, err := s.FetchUserPosts("realDonaldTrump", 1)
	if err != nil {
		log.Fatalf("Error fetching posts: %v", err)
	}

	if len(posts) == 0 {
		log.Fatal("No posts found")
	}

	// Get the latest post
	latestPost := posts[0]
	fmt.Printf("Latest post found:\n")
	fmt.Printf("ID: %s\n", latestPost.ID)
	fmt.Printf("Content: %s\n", latestPost.Content)
	fmt.Printf("Timestamp: %s\n", latestPost.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("URL: %s\n\n", latestPost.URL)

	// Initialize OpenAI client
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	openaiClient := openai.NewClient(openaiKey)

	// Analyze the post
	fmt.Println("Analyzing post with AI...")
	analysis, err := analyzePost(openaiClient, latestPost)
	if err != nil {
		log.Fatalf("Error analyzing post: %v", err)
	}

	// Display analysis results
	fmt.Printf("AI Analysis Results:\n")
	fmt.Printf("Market Impact: %s\n", analysis.MarketImpact)
	fmt.Printf("Confidence: %.2f\n", analysis.Confidence)
	fmt.Printf("Key Points: %v\n", analysis.KeyPoints)
	fmt.Printf("Affected Sectors: %v\n", analysis.AffectedSectors)
	fmt.Printf("Summary: %s\n", analysis.Summary)
}

func analyzePost(client *openai.Client, post scraper.TruthSocialPost) (*MarketAnalysis, error) {
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
`, post.Content)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a financial analyst specializing in market sentiment analysis. Provide objective, data-driven analysis of social media posts and their potential market impact.",
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

	var analysis MarketAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w", err)
	}

	return &analysis, nil
}
