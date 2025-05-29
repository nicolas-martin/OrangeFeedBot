package scraper

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type TruthSocialPost struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	URL       string    `json:"url"`
}

type TruthSocialScraper struct {
	client    *http.Client
	baseURL   string
	userAgent string
}

func NewTruthSocialScraper() *TruthSocialScraper {
	return &TruthSocialScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://truthsocial.com",
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

func (ts *TruthSocialScraper) FetchUserPosts(username string, limit int) ([]TruthSocialPost, error) {
	// Remove @ if present
	username = strings.TrimPrefix(username, "@")

	// Only try to fetch real posts - no fallbacks
	posts, err := ts.fetchRealPosts(username, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch real posts: %w", err)
	}

	if len(posts) == 0 {
		return nil, fmt.Errorf("no posts found for user @%s", username)
	}

	return posts, nil
}

func (ts *TruthSocialScraper) fetchRealPosts(username string, limit int) ([]TruthSocialPost, error) {
	// Truth Social uses a profile URL format like: https://truthsocial.com/@username
	profileURL := fmt.Sprintf("%s/@%s", ts.baseURL, username)

	fmt.Printf("Attempting to fetch from: %s\n", profileURL)

	resp, err := ts.makeRequest(profileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile page: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("Content-Encoding: %s\n", resp.Header.Get("Content-Encoding"))
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	if resp.StatusCode != http.StatusOK {
		// Log response headers for debugging
		fmt.Printf("Response headers: %v\n", resp.Header)
		return nil, fmt.Errorf("received status code %d for profile page", resp.StatusCode)
	}

	// Handle compressed responses
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		fmt.Println("Decompressing gzip response...")
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	html := string(body)
	fmt.Printf("Response body length: %d bytes\n", len(html))

	// Save a snippet of the HTML for debugging (first 500 chars)
	if len(html) > 500 {
		fmt.Printf("HTML snippet: %s...\n", html[:500])
	} else {
		fmt.Printf("Full HTML: %s\n", html)
	}

	// Try to extract posts from the HTML
	posts, err := ts.parsePostsFromHTML(html, username)
	if err != nil {
		return nil, fmt.Errorf("failed to parse posts from HTML: %w", err)
	}

	fmt.Printf("Extracted %d posts from HTML\n", len(posts))

	// Limit the number of posts returned
	if len(posts) > limit {
		posts = posts[:limit]
	}

	return posts, nil
}

func (ts *TruthSocialScraper) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", ts.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	return ts.client.Do(req)
}

func (ts *TruthSocialScraper) parsePostsFromHTML(html, username string) ([]TruthSocialPost, error) {
	// Truth Social appears to use a React-based frontend, so we need to look for:
	// 1. JSON data embedded in script tags
	// 2. HTML structure for posts

	// First, try to find JSON data in script tags
	jsonPosts := ts.extractJSONPosts(html, username)
	if len(jsonPosts) > 0 {
		return jsonPosts, nil
	}

	// If no JSON found, try to parse HTML structure
	htmlPosts := ts.extractHTMLPosts(html, username)
	return htmlPosts, nil
}

func (ts *TruthSocialScraper) extractJSONPosts(html, username string) []TruthSocialPost {
	var posts []TruthSocialPost

	// Look for JSON data in script tags that might contain post data
	// Truth Social likely uses patterns like these:
	scriptPatterns := []string{
		`<script[^>]*>([^<]*window\.__INITIAL_STATE__[^<]*)</script>`,
		`<script[^>]*>([^<]*window\.__PRELOADED_STATE__[^<]*)</script>`,
		`<script[^>]*id="__NEXT_DATA__"[^>]*>([^<]*)</script>`,
		`<script[^>]*>([^<]*"statuses"[^<]*)</script>`,
	}

	for _, pattern := range scriptPatterns {
		scriptRegex := regexp.MustCompile(pattern)
		matches := scriptRegex.FindAllStringSubmatch(html, -1)

		for _, match := range matches {
			if len(match) >= 2 {
				scriptContent := match[1]
				fmt.Printf("Found script content with length: %d\n", len(scriptContent))

				// Try to extract post-like data from the JSON
				postPatterns := []string{
					`"content":"([^"]+)"[^}]*"created_at":"([^"]+)"[^}]*"id":"([^"]+)"`,
					`"text":"([^"]+)"[^}]*"created_at":"([^"]+)"[^}]*"id":"([^"]+)"`,
					`"status":"([^"]+)"[^}]*"created_at":"([^"]+)"[^}]*"id":"([^"]+)"`,
					`"id":"([^"]+)"[^}]*"content":"([^"]+)"[^}]*"created_at":"([^"]+)"`,
				}

				for _, postPattern := range postPatterns {
					postDataRegex := regexp.MustCompile(postPattern)
					postMatches := postDataRegex.FindAllStringSubmatch(scriptContent, -1)

					for _, postMatch := range postMatches {
						if len(postMatch) >= 4 {
							var content, createdAtStr, id string

							// Handle different match orders
							if strings.Contains(postPattern, `"content":"([^"]+)"`) {
								content = ts.unescapeJSON(postMatch[1])
								createdAtStr = postMatch[2]
								id = postMatch[3]
							} else {
								id = postMatch[1]
								content = ts.unescapeJSON(postMatch[2])
								createdAtStr = postMatch[3]
							}

							// Skip if content is too short or looks like metadata
							if len(content) < 10 || ts.isMetadataContent(content) {
								continue
							}

							createdAt, err := time.Parse(time.RFC3339, createdAtStr)
							if err != nil {
								// Try alternative time formats
								timeFormats := []string{
									"2006-01-02T15:04:05.000Z",
									"2006-01-02T15:04:05Z",
									"2006-01-02 15:04:05",
								}

								for _, format := range timeFormats {
									if t, err := time.Parse(format, createdAtStr); err == nil {
										createdAt = t
										break
									}
								}

								if createdAt.IsZero() {
									createdAt = time.Now() // fallback
								}
							}

							post := TruthSocialPost{
								ID:        id,
								Content:   content,
								CreatedAt: createdAt,
								Username:  username,
								URL:       fmt.Sprintf("%s/@%s/posts/%s", ts.baseURL, username, id),
							}
							posts = append(posts, post)

							fmt.Printf("Extracted post: ID=%s, Content=%.50s...\n", id, content)
						}
					}
				}
			}
		}
	}

	return posts
}

func (ts *TruthSocialScraper) extractHTMLPosts(html, username string) []TruthSocialPost {
	var posts []TruthSocialPost

	// Look for post-like structures in the HTML
	// Truth Social might use various div structures for posts
	postPatterns := []string{
		`(?s)<div[^>]*class="[^"]*status[^"]*"[^>]*data-id="([^"]*)"[^>]*>(.*?)</div>`,
		`(?s)<article[^>]*data-id="([^"]*)"[^>]*>(.*?)</article>`,
		`(?s)<div[^>]*class="[^"]*post[^"]*"[^>]*data-id="([^"]*)"[^>]*>(.*?)</div>`,
		`(?s)<div[^>]*id="status-([^"]*)"[^>]*>(.*?)</div>`,
	}

	for _, pattern := range postPatterns {
		postRegex := regexp.MustCompile(pattern)
		matches := postRegex.FindAllStringSubmatch(html, -1)

		for _, match := range matches {
			if len(match) >= 3 {
				id := match[1]
				postHTML := match[2]

				// Extract content from the post HTML
				content := ts.extractContentFromPostHTML(postHTML)
				if content == "" || len(content) < 10 {
					continue
				}

				// Extract timestamp if available
				createdAt := ts.extractTimestampFromPostHTML(postHTML)

				post := TruthSocialPost{
					ID:        id,
					Content:   content,
					CreatedAt: createdAt,
					Username:  username,
					URL:       fmt.Sprintf("%s/@%s/posts/%s", ts.baseURL, username, id),
				}
				posts = append(posts, post)

				fmt.Printf("Extracted HTML post: ID=%s, Content=%.50s...\n", id, content)
			}
		}
	}

	// If no posts found with the above patterns, try alternative patterns
	if len(posts) == 0 {
		posts = ts.extractPostsAlternativePattern(html, username)
	}

	return posts
}

func (ts *TruthSocialScraper) extractContentFromPostHTML(postHTML string) string {
	// Look for content in various possible containers
	contentPatterns := []string{
		`<div[^>]*class="[^"]*status__content[^"]*"[^>]*>(.*?)</div>`,
		`<div[^>]*class="[^"]*post-content[^"]*"[^>]*>(.*?)</div>`,
		`<p[^>]*class="[^"]*status-content[^"]*"[^>]*>(.*?)</p>`,
		`<span[^>]*class="[^"]*text[^"]*"[^>]*>(.*?)</span>`,
	}

	for _, pattern := range contentPatterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindStringSubmatch(postHTML)
		if len(matches) >= 2 {
			content := ts.cleanHTML(matches[1])
			if content != "" {
				return content
			}
		}
	}

	return ""
}

func (ts *TruthSocialScraper) extractTimestampFromPostHTML(postHTML string) time.Time {
	// Look for timestamp in various formats
	timePatterns := []string{
		`<time[^>]*datetime="([^"]+)"`,
		`<span[^>]*class="[^"]*time[^"]*"[^>]*title="([^"]+)"`,
		`data-timestamp="([^"]+)"`,
	}

	for _, pattern := range timePatterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindStringSubmatch(postHTML)
		if len(matches) >= 2 {
			timeStr := matches[1]

			// Try different time formats
			timeFormats := []string{
				time.RFC3339,
				"2006-01-02T15:04:05.000Z",
				"2006-01-02 15:04:05",
				"Jan 2, 2006 3:04 PM",
			}

			for _, format := range timeFormats {
				if t, err := time.Parse(format, timeStr); err == nil {
					return t
				}
			}
		}
	}

	return time.Now() // fallback to current time
}

func (ts *TruthSocialScraper) extractPostsAlternativePattern(html, username string) []TruthSocialPost {
	var posts []TruthSocialPost

	// Try to find any text that looks like social media posts
	// This is a more aggressive approach for when structured data isn't available
	textBlocks := regexp.MustCompile(`(?s)<p[^>]*>(.*?)</p>`).FindAllStringSubmatch(html, -1)

	for i, match := range textBlocks {
		if len(match) >= 2 {
			content := ts.cleanHTML(match[1])

			// Filter out content that's likely not a post (too short, navigation text, etc.)
			if len(content) < 10 || ts.isNavigationText(content) {
				continue
			}

			post := TruthSocialPost{
				ID:        fmt.Sprintf("extracted_%d_%d", time.Now().Unix(), i),
				Content:   content,
				CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour), // Estimate timestamps
				Username:  username,
				URL:       fmt.Sprintf("%s/@%s", ts.baseURL, username),
			}
			posts = append(posts, post)

			// Limit to reasonable number
			if len(posts) >= 10 {
				break
			}
		}
	}

	return posts
}

func (ts *TruthSocialScraper) isNavigationText(text string) bool {
	navigationKeywords := []string{
		"home", "profile", "settings", "logout", "login", "sign up",
		"about", "privacy", "terms", "contact", "help", "support",
		"follow", "followers", "following", "notifications",
	}

	lowerText := strings.ToLower(text)
	for _, keyword := range navigationKeywords {
		if strings.Contains(lowerText, keyword) && len(text) < 50 {
			return true
		}
	}

	return false
}

func (ts *TruthSocialScraper) unescapeJSON(s string) string {
	// Unescape JSON string content
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

func (ts *TruthSocialScraper) cleanHTML(html string) string {
	// Remove HTML tags and decode entities
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	return strings.TrimSpace(text)
}

func (ts *TruthSocialScraper) isMetadataContent(content string) bool {
	metadataKeywords := []string{
		"application/json", "text/javascript", "window.", "document.",
		"function", "var ", "const ", "let ", "return", "null", "undefined",
		"true", "false", "{}", "[]", "http://", "https://",
	}

	lowerContent := strings.ToLower(content)
	for _, keyword := range metadataKeywords {
		if strings.Contains(lowerContent, keyword) {
			return true
		}
	}

	// Check if it's mostly special characters or numbers
	alphaCount := 0
	for _, r := range content {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			alphaCount++
		}
	}

	// If less than 30% alphabetic characters, likely metadata
	return float64(alphaCount)/float64(len(content)) < 0.3
}
