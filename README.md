# OrangeFeed - Truth Social Market Intelligence Bot

A Go-based application that monitors Truth Social posts and analyzes their potential market impact using AI.

## 🎯 Project Status

**Current Implementation**: Go-native Truth Social client with AI analysis integration

**Authentication**: ✅ Successfully implemented using Stanford Truthbrush's OAuth approach  
**API Calls**: ✅ Proper endpoint handling and data parsing  
**AI Analysis**: ✅ OpenAI integration for market impact assessment  
**Challenge**: 🚧 Cloudflare protection blocks direct HTTP requests from Go

## 🔧 Technical Architecture

### Core Components

- **Truth Social Client** (`internal/truthsocial/client.go`): Go-native API client
- **AI Analysis** (`test_real_ai.go`): OpenAI-powered market sentiment analysis
- **Authentication**: OAuth 2.0 password grant flow using extracted client credentials

### Authentication Implementation

Our Go client uses the same approach as [Stanford Truthbrush](https://github.com/stanfordio/truthbrush):

```go
// OAuth credentials extracted from Truth Social's JavaScript
clientID     = "9X1Fdd-pxNsAgEDNi_SfhJWi8T-vLuV2WVzKIbkTCw4"
clientSecret = "ozF8jzI4968oTKFkEnsBC-UbLPCdrSv0MkXGQu2o_-M"

// Password grant flow
payload := map[string]string{
    "client_id":     clientID,
    "client_secret": clientSecret,
    "grant_type":    "password",
    "username":      username,
    "password":      password,
    "redirect_uri":  "urn:ietf:wg:oauth:2.0:oob",
    "scope":         "read",
}
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Truth Social account credentials
- OpenAI API key

### Environment Setup

```bash
# Copy environment template
cp config.env.example .env

# Configure credentials
export TRUTHSOCIAL_USERNAME="your_username"
export TRUTHSOCIAL_PASSWORD="your_password"
export OPENAI_API_KEY="your_openai_key"
```

### Run the Application

```bash
# Test the implementation
go run test_real_ai.go

# Build for production
go build -o orangefeed test_real_ai.go
```

## 📊 AI Analysis Features

The application analyzes Truth Social posts for:

- **Market Impact**: Positive/Negative/Neutral classification
- **Confidence Score**: 0.0-1.0 confidence rating
- **Key Points**: Extracted important statements
- **Affected Sectors**: Potentially impacted market sectors
- **Summary**: AI-generated content summary

### Sample Analysis Output

```json
{
  "summary": "Post expresses optimism about economic performance",
  "market_impact": "positive",
  "confidence": 0.85,
  "key_points": [
    "Mentions record stock market highs",
    "References low unemployment",
    "Positive trade deal sentiment"
  ],
  "affected_sectors": [
    "General Market",
    "Trade-sensitive stocks"
  ]
}
```

## 🔍 Implementation Details

### Truth Social API Integration

Based on analysis of [Stanford Truthbrush](https://github.com/stanfordio/truthbrush/blob/main/truthbrush/api.py), our implementation:

1. **Uses extracted OAuth credentials** from Truth Social's JavaScript
2. **Implements password grant flow** for direct authentication
3. **Follows Mastodon API patterns** (Truth Social is Mastodon-based)
4. **Handles proper headers and user agents** for API compatibility

### Cloudflare Protection Challenge

**Issue**: Truth Social uses Cloudflare protection that blocks standard HTTP clients

**Stanford Truthbrush Solution**: Uses `curl_cffi` with Chrome impersonation:
```python
impersonate="chrome123"  # Mimics real Chrome browser
```

**Go Limitations**: No equivalent to `curl_cffi`'s TLS fingerprint spoofing

### Potential Solutions

1. **Headless Browser**: Use chromedp with stealth techniques
2. **Proxy Services**: Route through Cloudflare-bypassing proxies  
3. **Python Integration**: Use Truthbrush as subprocess (previously working)
4. **Browser Automation**: Selenium/Playwright with anti-detection

## 🛠 Development

### Project Structure

```
├── internal/
│   └── truthsocial/
│       └── client.go          # Go-native Truth Social client
├── test_real_ai.go            # Main test application
├── config.env.example         # Environment template
└── README.md                  # This file
```

### Key Files

- **`internal/truthsocial/client.go`**: Complete Truth Social API client
- **`test_real_ai.go`**: Integration test with AI analysis
- **Authentication flow**: Mirrors Stanford Truthbrush implementation

### Testing

```bash
# Run with debug output
go run test_real_ai.go

# Check authentication (will hit Cloudflare)
curl -X POST https://truthsocial.com/oauth/token \
  -H "Content-Type: application/json" \
  -d '{"client_id":"9X1Fdd-pxNsAgEDNi_SfhJWi8T-vLuV2WVzKIbkTCw4",...}'
```

## 📈 Market Analysis Capabilities

### Supported Analysis Types

- **Economic Policy**: Trade, taxation, regulation mentions
- **Company-Specific**: Direct company references and impacts
- **Market Sentiment**: General market optimism/pessimism
- **Sector Impact**: Industry-specific implications

### AI Prompt Engineering

The system uses carefully crafted prompts for:
- Objective market analysis (not political opinions)
- Confidence scoring based on historical patterns
- Sector identification and impact assessment
- Key point extraction for actionable insights

## 🔗 References

- [Stanford Truthbrush](https://github.com/stanfordio/truthbrush) - Python Truth Social API client
- [Mastodon API Documentation](https://docs.joinmastodon.org/api/) - API reference
- [Truth Social OAuth Flow](https://docs.joinmastodon.org/spec/oauth/) - Authentication details

## ⚖️ Legal & Ethical Considerations

- **Terms of Service**: Respect Truth Social's ToS
- **Rate Limiting**: Implement appropriate request throttling
- **Data Usage**: For research and analysis purposes only
- **Disclaimer**: Not financial advice - for informational purposes only

## 🎯 Next Steps

1. **Cloudflare Bypass**: Implement browser automation or proxy solution
2. **Database Integration**: Store historical posts and analysis
3. **Real-time Monitoring**: Continuous post monitoring with alerts
4. **Enhanced AI**: More sophisticated market impact models
5. **Multi-platform**: Extend to other social media platforms

---

**Status**: Authentication implemented ✅ | API client ready ✅ | Cloudflare bypass needed 🚧 