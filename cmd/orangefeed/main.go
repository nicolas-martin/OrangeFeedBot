package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"orangefeed/internal/analyzer"
	"orangefeed/internal/truthsocial"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

type OrangeFeedBot struct {
	telegramBot    *tgbotapi.BotAPI
	truthClient    *truthsocial.Client
	analyzer       *analyzer.MarketAnalyzer
	chatID         int64
	targetUsername string
	lastPostID     string
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	fmt.Println("🎯 Starting OrangeFeed - Truth Social Market Intelligence Bot")
	fmt.Println(strings.Repeat("=", 70))

	bot, err := NewOrangeFeedBot()
	if err != nil {
		log.Fatal("Failed to initialize OrangeFeed bot:", err)
	}

	// Start the monitoring system
	bot.Start()
}

func NewOrangeFeedBot() (*OrangeFeedBot, error) {
	// Initialize Telegram bot
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	telegramBot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	// Get chat ID
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TELEGRAM_CHAT_ID: %w", err)
	}

	// Initialize Truth Social client
	truthUsername := os.Getenv("TRUTHSOCIAL_USERNAME")
	truthPassword := os.Getenv("TRUTHSOCIAL_PASSWORD")
	if truthUsername == "" || truthPassword == "" {
		return nil, fmt.Errorf("TRUTHSOCIAL_USERNAME and TRUTHSOCIAL_PASSWORD are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	truthClient, err := truthsocial.NewClient(ctx, truthUsername, truthPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create Truth Social client: %w", err)
	}

	// Initialize market analyzer
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	analyzer := analyzer.NewMarketAnalyzer(openaiKey)

	targetUsername := os.Getenv("TARGET_USERNAME")
	if targetUsername == "" {
		targetUsername = "realDonaldTrump"
	}

	return &OrangeFeedBot{
		telegramBot:    telegramBot,
		truthClient:    truthClient,
		analyzer:       analyzer,
		chatID:         chatID,
		targetUsername: targetUsername,
	}, nil
}

func (b *OrangeFeedBot) Start() {
	log.Printf("🚀 Starting OrangeFeed monitoring for @%s", b.targetUsername)

	// Send startup message
	b.sendMessage(fmt.Sprintf(`🤖 *OrangeFeed Market Intelligence Bot Started!*

📊 Monitoring: @%s
🎯 Features:
• Real-time Truth Social monitoring
• Advanced AI market analysis
• Specific stock recommendations
• Trading signals & risk assessment
• Sector impact analysis

🔄 Bot is now active and monitoring for new posts...`, b.targetUsername))

	// Set up cron job for monitoring
	c := cron.New()

	intervalStr := os.Getenv("CHECK_INTERVAL_MINUTES")
	if intervalStr == "" {
		intervalStr = "15" // Default to 15 minutes
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 15
	}

	cronExpr := fmt.Sprintf("*/%d * * * *", interval)
	log.Printf("⏰ Setting up monitoring every %d minutes", interval)

	c.AddFunc(cronExpr, func() {
		log.Println("🔍 Checking for new posts...")
		b.checkForNewPosts()
	})

	c.Start()

	// Keep the program running
	log.Println("✅ OrangeFeed is running. Press Ctrl+C to stop.")
	select {}
}

func (b *OrangeFeedBot) checkForNewPosts() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Fetch recent posts
	statuses, err := b.truthClient.PullStatuses(ctx, b.targetUsername, true, 10)
	if err != nil {
		log.Printf("❌ Error fetching posts: %v", err)
		b.sendMessage(fmt.Sprintf("⚠️ Error fetching posts from @%s: %v", b.targetUsername, err))
		return
	}

	if len(statuses) == 0 {
		log.Printf("📭 No posts found for @%s", b.targetUsername)
		return
	}

	log.Printf("📄 Found %d posts to process", len(statuses))

	// Process new posts (stop when we reach a post we've already processed)
	newPostsCount := 0
	for _, status := range statuses {
		if status.ID == b.lastPostID {
			break // We've reached posts we've already processed
		}

		// Clean and validate content
		content := b.cleanContent(status.Content)
		if len(content) < 10 {
			continue // Skip very short posts
		}

		log.Printf("🔍 Analyzing new post: %s", status.ID)

		// Analyze the post
		analysis, err := b.analyzer.AnalyzePost(content)
		if err != nil {
			log.Printf("❌ Error analyzing post %s: %v", status.ID, err)
			continue
		}

		// Send analysis to Telegram
		b.sendAnalysis(status, analysis)
		newPostsCount++
	}

	if newPostsCount > 0 {
		b.lastPostID = statuses[0].ID
		log.Printf("✅ Processed %d new posts", newPostsCount)
	} else {
		log.Println("📭 No new posts to process")
	}
}

func (b *OrangeFeedBot) sendAnalysis(status truthsocial.Status, analysis *analyzer.Analysis) {
	content := b.cleanContent(status.Content)

	// Create comprehensive analysis message
	message := fmt.Sprintf(`🚨 *NEW POST ANALYSIS*

📝 *Post Content:*
%s

📊 *MARKET ANALYSIS:*
🎯 Impact: *%s* (%.0f%% confidence)
📈 Signal: *%s*
⏰ Horizon: %s
⚠️ Risk: %s
📏 Magnitude: %s

📝 *Summary:*
%s`,
		b.escapeMarkdown(content),
		strings.ToUpper(analysis.MarketImpact),
		analysis.Confidence*100,
		strings.ToUpper(analysis.TradingSignal),
		analysis.TimeHorizon,
		strings.ToUpper(analysis.RiskLevel),
		analysis.ExpectedMagnitude,
		b.escapeMarkdown(analysis.Summary))

	// Add key points
	if len(analysis.KeyPoints) > 0 {
		message += "\n\n🔑 *Key Points:*\n"
		for _, point := range analysis.KeyPoints {
			message += fmt.Sprintf("• %s\n", b.escapeMarkdown(point))
		}
	}

	// Add affected sectors
	if len(analysis.AffectedSectors) > 0 {
		message += fmt.Sprintf("\n🏭 *Sectors:* %s", strings.Join(analysis.AffectedSectors, ", "))
	}

	// Add specific stocks
	if len(analysis.SpecificStocks) > 0 {
		message += fmt.Sprintf("\n📈 *Stocks:* %s", strings.Join(analysis.SpecificStocks, ", "))
	}

	// Add actionable insights
	if len(analysis.ActionableInsights) > 0 {
		message += "\n\n💡 *TRADING INSIGHTS:*\n"
		for i, insight := range analysis.ActionableInsights {
			message += fmt.Sprintf("%d. %s\n", i+1, b.escapeMarkdown(insight))
		}
	}

	// Add post metadata
	message += fmt.Sprintf("\n\n🔗 [View Post](%s)\n📅 %s\n👍 %d likes | 🔄 %d reblogs",
		status.URL,
		status.CreatedAt,
		status.FavouritesCount,
		status.ReblogsCount)

	b.sendMessage(message)
}

func (b *OrangeFeedBot) sendMessage(text string) {
	msg := tgbotapi.NewMessage(b.chatID, text)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	_, err := b.telegramBot.Send(msg)
	if err != nil {
		log.Printf("❌ Error sending message: %v", err)
	}
}

func (b *OrangeFeedBot) cleanContent(content string) string {
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

func (b *OrangeFeedBot) escapeMarkdown(text string) string {
	// Escape special Markdown characters
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"`", "\\`",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
