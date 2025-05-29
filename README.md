# Truth Social Market Impact Bot

A Go-based bot that monitors Donald Trump's Truth Social posts, analyzes their potential stock market impact using AI, and posts findings to a Telegram group.

## Features

- 🔍 **Truth Social Monitoring**: Fetches posts from specified Truth Social accounts
- 🤖 **AI Analysis**: Uses OpenAI GPT-4 to analyze market impact potential
- 📊 **Market Impact Assessment**: Categorizes posts as positive, negative, or neutral for markets
- 📱 **Telegram Integration**: Posts analysis results to Telegram groups/channels
- ⏰ **Automated Scheduling**: Configurable check intervals using cron jobs
- 🎯 **Sector Analysis**: Identifies potentially affected market sectors

## Prerequisites

- Go 1.21 or higher
- OpenAI API key
- Telegram Bot Token
- Telegram Chat/Group ID

## Setup

### 1. Clone and Install Dependencies

```bash
git clone <repository-url>
cd OrangeFeed
go mod tidy
```

### 2. Environment Configuration

Copy the example environment file and configure your settings:

```bash
cp config.env.example .env
```

Edit `.env` with your credentials:

```env
# Truth Social API Configuration
TRUTH_SOCIAL_USERNAME=your_username
TRUTH_SOCIAL_PASSWORD=your_password

# OpenAI API Configuration
OPENAI_API_KEY=your_openai_api_key

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
TELEGRAM_CHAT_ID=your_telegram_chat_id

# Bot Configuration
CHECK_INTERVAL_MINUTES=30
TARGET_USERNAME=realDonaldTrump
```

### 3. Getting Required Credentials

#### OpenAI API Key
1. Visit [OpenAI API](https://platform.openai.com/api-keys)
2. Create a new API key
3. Add it to your `.env` file

#### Telegram Bot Token
1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Create a new bot with `/newbot`
3. Copy the bot token to your `.env` file

#### Telegram Chat ID
1. Add your bot to the target group/channel
2. Send a message to the group
3. Visit `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
4. Find the chat ID in the response
5. Add it to your `.env` file

### 4. Build and Run

```bash
# Build the application
go build -o truthsocial-bot cmd/main.go

# Run the bot
./truthsocial-bot
```

Or run directly:

```bash
go run cmd/main.go
```

## Configuration Options

| Variable | Description | Default |
|----------|-------------|---------|
| `TARGET_USERNAME` | Truth Social username to monitor | `realDonaldTrump` |
| `CHECK_INTERVAL_MINUTES` | How often to check for new posts | `30` |
| `OPENAI_API_KEY` | Your OpenAI API key | Required |
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token | Required |
| `TELEGRAM_CHAT_ID` | Target Telegram chat/group ID | Required |

## Sample Output

The bot will send formatted messages to your Telegram group like this:

```
🚨 New Truth Social Post Analysis

👤 User: @realDonaldTrump
🕐 Time: 2024-01-15 14:30:00 UTC
🔗 Link: View Post

📝 Post Content:
The American economy is BOOMING! Record stock market highs, unemployment at historic lows. Our trade deals are working! 🇺🇸📈

📈 Market Impact: Positive
🎯 Confidence: 85.0%

📊 Summary:
Post expresses strong optimism about economic performance and trade policies, likely to boost market sentiment.

🔍 Key Points:
• Mentions record stock market highs
• References low unemployment
• Positive trade deal sentiment

🏭 Affected Sectors:
• General Market
• Trade-sensitive stocks
• Employment-related sectors

---
Analysis powered by AI • Not financial advice
```

## Architecture

```
cmd/
├── main.go              # Main application entry point

internal/
├── scraper/
│   └── truthsocial.go   # Truth Social scraping logic

config.env.example       # Environment configuration template
go.mod                   # Go module dependencies
README.md               # This file
```

## Important Notes

### Truth Social API Limitations

Truth Social doesn't provide a public API, so this implementation uses:
- Mock data for demonstration purposes
- Web scraping techniques (commented framework provided)
- You may need to implement actual scraping logic based on current site structure

### Legal and Ethical Considerations

- Respect Truth Social's Terms of Service
- Implement appropriate rate limiting
- Consider using official APIs when available
- This tool is for educational/research purposes

### Disclaimer

- This bot provides AI-generated analysis for informational purposes only
- Not financial advice
- Market impact predictions are speculative
- Always do your own research before making investment decisions

## Development

### Adding New Features

1. **Custom Analysis Prompts**: Modify the prompt in `analyzePost()` function
2. **Additional Data Sources**: Extend the scraper to support multiple platforms
3. **Enhanced Formatting**: Customize the Telegram message format in `sendAnalysis()`
4. **Database Storage**: Add persistence for historical analysis

### Testing

```bash
# Run tests
go test ./...

# Run with verbose output
go test -v ./...
```

### Building for Production

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o truthsocial-bot-linux cmd/main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o truthsocial-bot-macos cmd/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o truthsocial-bot.exe cmd/main.go
```

## Deployment

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o truthsocial-bot cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/truthsocial-bot .
CMD ["./truthsocial-bot"]
```

### Systemd Service

Create `/etc/systemd/system/truthsocial-bot.service`:

```ini
[Unit]
Description=Truth Social Market Impact Bot
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/your/bot
ExecStart=/path/to/your/bot/truthsocial-bot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
1. Check the existing issues
2. Create a new issue with detailed information
3. Include logs and configuration (without sensitive data)

## Current Status

✅ **Working Features:**
- Complete bot architecture with Truth Social scraping, AI analysis, and Telegram posting
- HTTP request handling with gzip decompression
- Robust error handling (bot will not post if no real data is available)
- Comprehensive test suite
- Docker deployment ready

⚠️ **Truth Social Scraping Limitations:**
Truth Social uses a React-based single-page application (SPA) where content is loaded dynamically via JavaScript. The current implementation successfully:
- Makes authenticated HTTP requests to Truth Social
- Handles compressed responses (gzip)
- Parses the initial HTML structure
- **Only works with real data - no mock/fake posts will ever be generated**

**Important:** The bot will only post to Telegram when it successfully extracts real posts from Truth Social. If scraping fails, the bot will log the error and wait for the next check interval.

## Next Steps for Production

To implement real Truth Social scraping, you would need to:

### 1. **JavaScript Rendering**
```bash
# Add headless browser support
go get github.com/chromedp/chromedp
```

### 2. **Authentication**
- Implement Truth Social login flow
- Handle session management and cookies
- Manage rate limiting and request throttling

### 3. **Dynamic Content Loading**
- Wait for JavaScript to load posts
- Handle infinite scroll pagination
- Parse dynamically generated DOM elements

### 4. **Advanced Parsing**
- Extract posts from React component state
- Handle Truth Social's specific data structures
- Parse timestamps, user information, and post metadata

## Sample Implementation for Dynamic Content

```go
// Example using chromedp for JavaScript rendering
func (ts *TruthSocialScraper) fetchWithBrowser(username string) ([]TruthSocialPost, error) {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()
    
    var html string
    err := chromedp.Run(ctx,
        chromedp.Navigate(fmt.Sprintf("https://truthsocial.com/@%s", username)),
        chromedp.WaitVisible(`[data-testid="post"]`, chromedp.ByQuery),
        chromedp.Sleep(2*time.Second), // Wait for posts to load
        chromedp.OuterHTML("html", &html),
    )
    
    if err != nil {
        return nil, err
    }
    
    return ts.parsePostsFromHTML(html, username)
}
```

---

**⚠️ Disclaimer**: This tool is for educational and research purposes only. Always comply with platform terms of service and applicable laws. The analysis provided is AI-generated and should not be considered financial advice. 