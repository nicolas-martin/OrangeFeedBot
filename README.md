# 🎯 OrangeFeed - Truth Social Market Intelligence Bot

A sophisticated Go-based application that monitors Truth Social posts from @realDonaldTrump and provides real-time AI-powered market analysis with actionable trading insights.

## 🚀 Features

### 📊 Advanced Market Analysis
- **Real-time monitoring** of Truth Social posts
- **AI-powered analysis** using GPT-4 for market impact assessment
- **Specific stock recommendations** with ticker symbols
- **Trading signals** (Buy/Sell/Hold/Watch)
- **Risk assessment** and confidence scoring
- **Sector impact analysis**
- **Time horizon predictions** (immediate to long-term)

### 🔧 Technical Capabilities
- **Cloudflare bypass** using CycleTLS with Chrome fingerprint spoofing
- **OAuth authentication** with Truth Social API
- **Telegram integration** for real-time notifications
- **Automated monitoring** with configurable intervals
- **Docker support** for easy deployment
- **Production-ready** error handling and logging

## 📋 Prerequisites

- Go 1.21 or higher
- Truth Social account credentials
- OpenAI API key
- Telegram bot token and chat ID

## 🛠️ Quick Start

### 1. Clone and Setup
```bash
git clone <repository-url>
cd OrangeFeed
make setup
```

### 2. Configure Environment
Edit `.env` file with your credentials:
```env
# Truth Social Credentials
TRUTHSOCIAL_USERNAME=your_username
TRUTHSOCIAL_PASSWORD=your_password

# OpenAI API Key for Market Analysis
OPENAI_API_KEY=your_openai_api_key

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
TELEGRAM_CHAT_ID=your_chat_id

# Monitoring Configuration
TARGET_USERNAME=realDonaldTrump
CHECK_INTERVAL_MINUTES=15
```

### 3. Install Dependencies
```bash
make deps
```

### 4. Test the System
```bash
make test
```

### 5. Run the Application
```bash
make run
```

## 🐳 Docker Deployment

### Build and Run with Docker
```bash
make docker-run
```

### Manual Docker Commands
```bash
docker build -t orangefeed .
docker-compose up
```

## 📊 Sample Analysis Output

When a new post is detected, you'll receive a comprehensive analysis via Telegram:

```
🚨 NEW POST ANALYSIS

📝 Post Content:
"FHFA Director Pulte calls on Powell to lower interest rates"

📊 MARKET ANALYSIS:
🎯 Impact: BULLISH (70% confidence)
📈 Signal: BUY
⏰ Horizon: medium-term
⚠️ Risk: MEDIUM
📏 Magnitude: moderate

📝 Summary:
The post indicates potential Federal Reserve policy changes that could stimulate economic growth through lower borrowing costs.

🔑 Key Points:
• Potential lowering of interest rates
• Increased borrowing and investment activity
• Stimulated economic growth

🏭 Sectors: Financials, Real Estate, Consumer Discretionary
📈 Stocks: JPM, BAC, GS, WFC, XLF, VNQ, AMZN, HD

💡 TRADING INSIGHTS:
1. Buy financial sector ETFs such as XLF as lower rates could increase lending activity
2. Invest in real estate ETFs like VNQ as lower rates stimulate housing market
3. Consider consumer discretionary stocks as lower rates boost spending

🔗 View Post | 📅 2025-05-29T03:16:15.923Z
👍 1598 likes | 🔄 503 reblogs
```

## 🏗️ Architecture

### Project Structure
```
OrangeFeed/
├── cmd/orangefeed/          # Main application entry point
├── internal/
│   ├── truthsocial/         # Truth Social API client
│   └── analyzer/            # Market analysis engine
├── test_real_ai.go          # Test application
├── docker-compose.yml       # Docker configuration
├── Dockerfile              # Container definition
├── Makefile                # Build automation
└── README.md               # Documentation
```

### Key Components

#### 1. Truth Social Client (`internal/truthsocial/`)
- **CycleTLS integration** for Cloudflare bypass
- **Chrome fingerprint spoofing** for stealth operation
- **OAuth 2.0 authentication** using extracted client credentials
- **Pagination support** for fetching multiple posts
- **Rate limiting** and error handling

#### 2. Market Analyzer (`internal/analyzer/`)
- **GPT-4 integration** for sophisticated analysis
- **Structured output** with specific trading recommendations
- **Batch processing** for multiple posts
- **Sentiment aggregation** across multiple analyses

#### 3. Main Application (`cmd/orangefeed/`)
- **Telegram bot integration** for notifications
- **Cron-based monitoring** with configurable intervals
- **Markdown formatting** for rich message display
- **Error recovery** and logging

## 🔧 Configuration Options

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TRUTHSOCIAL_USERNAME` | Truth Social username | Required |
| `TRUTHSOCIAL_PASSWORD` | Truth Social password | Required |
| `OPENAI_API_KEY` | OpenAI API key | Required |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token | Required |
| `TELEGRAM_CHAT_ID` | Telegram chat ID | Required |
| `TARGET_USERNAME` | Username to monitor | `realDonaldTrump` |
| `CHECK_INTERVAL_MINUTES` | Monitoring interval | `15` |

### Monitoring Intervals
- **Immediate**: Real-time monitoring (not recommended due to rate limits)
- **15 minutes**: Balanced approach (recommended)
- **30 minutes**: Conservative monitoring
- **60 minutes**: Low-frequency monitoring

## 🛡️ Security Features

### Cloudflare Bypass
- **CycleTLS library** with Chrome fingerprint spoofing
- **JA3 fingerprint**: `771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0`
- **User-Agent spoofing**: Chrome 123 on Windows 10
- **HTTP/2 support** for modern protocol compliance

### Authentication
- **OAuth 2.0 password grant** flow
- **Client credentials** extracted from Truth Social's JavaScript
- **Token management** with automatic refresh
- **Secure credential storage** via environment variables

## 📈 Market Analysis Features

### Analysis Dimensions
- **Market Impact**: Bullish/Bearish/Neutral
- **Confidence Score**: 0-100% confidence level
- **Trading Signal**: Buy/Sell/Hold/Watch recommendations
- **Time Horizon**: Immediate/Short-term/Medium-term/Long-term
- **Risk Level**: Low/Medium/High risk assessment
- **Expected Magnitude**: Minimal/Moderate/Significant/Major impact

### Specific Outputs
- **Stock Tickers**: Actual ticker symbols (e.g., AAPL, TSLA, JPM)
- **Sector Analysis**: Affected market sectors
- **Trading Strategies**: Specific entry/exit recommendations
- **Risk Management**: Position sizing and stop-loss guidance

## 🚨 Troubleshooting

### Common Issues

#### Authentication Failures
```bash
# Check credentials
echo $TRUTHSOCIAL_USERNAME
echo $TRUTHSOCIAL_PASSWORD

# Test authentication
make test
```

#### Cloudflare Blocking
- The application uses CycleTLS to bypass Cloudflare
- If blocked, try adjusting request intervals
- Monitor logs for specific error messages

#### OpenAI API Errors
```bash
# Check API key
echo $OPENAI_API_KEY

# Verify API quota and billing
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models
```

#### Telegram Integration Issues
```bash
# Test bot token
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe"

# Verify chat ID
curl "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getUpdates"
```

## 📊 Performance Metrics

### Typical Performance
- **Authentication**: ~2-3 seconds
- **Post Fetching**: ~5-10 seconds for 20 posts
- **AI Analysis**: ~10-15 seconds per post
- **Telegram Delivery**: ~1-2 seconds

### Resource Usage
- **Memory**: ~50-100MB during operation
- **CPU**: Low usage except during AI analysis
- **Network**: ~1-5MB per monitoring cycle

## 🔄 Development

### Available Commands
```bash
make help          # Show all available commands
make build         # Build the application
make run           # Build and run
make test          # Run test application
make clean         # Clean build artifacts
make deps          # Install dependencies
make lint          # Run code quality checks
make docker-build  # Build Docker image
make docker-run    # Run with Docker
make setup         # Setup development environment
```

### Testing
```bash
# Test Truth Social connection
make test

# Test specific components
go test ./internal/truthsocial/
go test ./internal/analyzer/
```

## 📜 License

This project is for educational and research purposes. Please ensure compliance with:
- Truth Social Terms of Service
- OpenAI API Terms of Use
- Telegram Bot API Terms
- Local regulations regarding automated trading advice

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📞 Support

For issues and questions:
1. Check the troubleshooting section
2. Review logs for error messages
3. Test individual components
4. Create an issue with detailed information

---

**⚠️ Disclaimer**: This tool provides automated analysis for informational purposes only. It is not financial advice. Always conduct your own research and consult with financial professionals before making investment decisions. 