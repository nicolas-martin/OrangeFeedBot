package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/sashabaranov/go-openai"

	"orangefeed/internal/scraper"
)

type MarketAnalysis struct {
	Summary         string   `json:"summary"`
	MarketImpact    string   `json:"market_impact"` // "positive", "negative", "neutral"
	Confidence      float64  `json:"confidence"`
	KeyPoints       []string `json:"key_points"`
	AffectedSectors []string `json:"affected_sectors"`
}

type Bot struct {
	telegramBot    *tgbotapi.BotAPI
	openaiClient   *openai.Client
	scraper        *scraper.TruthSocialScraper
	chatID         int64
	targetUsername string
	lastPostID     string
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	bot, err := NewBot()
	if err != nil {
		log.Fatal("Failed to initialize bot:", err)
	}

	// Start the bot
	bot.Start()
}

func NewBot() (*Bot, error) {
	// Initialize Telegram bot
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	telegramBot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	// Initialize OpenAI client
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	openaiClient := openai.NewClient(openaiKey)

	// Get chat ID
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TELEGRAM_CHAT_ID: %w", err)
	}

	targetUsername := os.Getenv("TARGET_USERNAME")
	if targetUsername == "" {
		targetUsername = "realDonaldTrump"
	}

	// Initialize scraper
	truthSocialScraper := scraper.NewTruthSocialScraper()

	return &Bot{
		telegramBot:    telegramBot,
		openaiClient:   openaiClient,
		scraper:        truthSocialScraper,
		chatID:         chatID,
		targetUsername: targetUsername,
	}, nil
}

func (b *Bot) Start() {
	log.Println("Starting Truth Social Monitor Bot...")

	// Send startup message
	b.sendMessage("🤖 Truth Social Monitor Bot started!\nMonitoring @" + b.targetUsername + " for new posts...")

	// Set up cron job
	c := cron.New()

	intervalStr := os.Getenv("CHECK_INTERVAL_MINUTES")
	if intervalStr == "" {
		intervalStr = "30"
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 30
	}

	cronExpr := fmt.Sprintf("*/%d * * * *", interval)

	c.AddFunc(cronExpr, func() {
		log.Println("Checking for new posts...")
		b.checkForNewPosts()
	})

	c.Start()

	// Keep the program running
	select {}
}

func (b *Bot) checkForNewPosts() {
	posts, err := b.scraper.FetchUserPosts(b.targetUsername, 5)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		return
	}

	if len(posts) == 0 {
		log.Printf("No posts found for @%s", b.targetUsername)
		return
	}

	log.Printf("Found %d posts to process", len(posts))

	for _, post := range posts {
		if post.ID == b.lastPostID {
			break // We've reached posts we've already processed
		}

		analysis, err := b.analyzePost(post)
		if err != nil {
			log.Printf("Error analyzing post %s: %v", post.ID, err)
			continue
		}

		b.sendAnalysis(post, analysis)
	}

	if len(posts) > 0 {
		b.lastPostID = posts[0].ID
	}
}

func (b *Bot) analyzePost(post scraper.TruthSocialPost) (*MarketAnalysis, error) {
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

	resp, err := b.openaiClient.CreateChatCompletion(
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

	var analysis MarketAnalysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %w", err)
	}

	return &analysis, nil
}

func (b *Bot) sendAnalysis(post scraper.TruthSocialPost, analysis *MarketAnalysis) {
	var impactEmoji string
	switch analysis.MarketImpact {
	case "positive":
		impactEmoji = "📈"
	case "negative":
		impactEmoji = "📉"
	default:
		impactEmoji = "➡️"
	}

	message := fmt.Sprintf(`
🚨 *New Truth Social Post Analysis*

👤 *User:* @%s
🕐 *Time:* %s
🔗 *Link:* [View Post](%s)

📝 *Post Content:*
_%s_

%s *Market Impact:* %s
🎯 *Confidence:* %.1f%%

📊 *Summary:*
%s

🔍 *Key Points:*
%s

🏭 *Affected Sectors:*
%s

---
_Analysis powered by AI • Not financial advice_
`,
		post.Username,
		post.CreatedAt.Format("2006-01-02 15:04:05 UTC"),
		post.URL,
		post.Content,
		impactEmoji,
		strings.Title(analysis.MarketImpact),
		analysis.Confidence*100,
		analysis.Summary,
		formatList(analysis.KeyPoints),
		formatList(analysis.AffectedSectors),
	)

	b.sendMessage(message)
}

func (b *Bot) sendMessage(text string) {
	msg := tgbotapi.NewMessage(b.chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true

	if _, err := b.telegramBot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "_None specified_"
	}

	var formatted []string
	for _, item := range items {
		formatted = append(formatted, "• "+item)
	}
	return strings.Join(formatted, "\n")
}
