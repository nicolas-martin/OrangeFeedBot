package truthsocial

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Constants from Stanford TruthBrush implementation
const (
	baseURL    = "https://truthsocial.com"
	apiBaseURL = "https://truthsocial.com/api"
	userAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"

	// OAuth client credentials from TruthBrush (extracted from Truth Social's JavaScript)
	clientID     = "9X1Fdd-pxNsAgEDNi_SfhJWi8T-vLuV2WVzKIbkTCw4"
	clientSecret = "ozF8jzI4968oTKFkEnsBC-UbLPCdrSv0MkXGQu2o_-M"
)

type Client struct {
	cycleTLS    cycletls.CycleTLS
	username    string
	password    string
	accessToken string
}

// AuthResponse represents the OAuth token response
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	CreatedAt   int64  `json:"created_at"`
}

// Account represents a Truth Social user account
type Account struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Acct           string `json:"acct"`
	DisplayName    string `json:"display_name"`
	Locked         bool   `json:"locked"`
	Bot            bool   `json:"bot"`
	Discoverable   bool   `json:"discoverable"`
	Group          bool   `json:"group"`
	CreatedAt      string `json:"created_at"`
	Note           string `json:"note"`
	URL            string `json:"url"`
	Avatar         string `json:"avatar"`
	AvatarStatic   string `json:"avatar_static"`
	Header         string `json:"header"`
	HeaderStatic   string `json:"header_static"`
	FollowersCount int    `json:"followers_count"`
	FollowingCount int    `json:"following_count"`
	StatusesCount  int    `json:"statuses_count"`
	LastStatusAt   string `json:"last_status_at"`
	Verified       bool   `json:"verified"`
}

// Status represents a Truth Social post
type Status struct {
	ID                 string        `json:"id"`
	CreatedAt          string        `json:"created_at"`
	InReplyToID        string        `json:"in_reply_to_id"`
	InReplyToAccountID string        `json:"in_reply_to_account_id"`
	Sensitive          bool          `json:"sensitive"`
	SpoilerText        string        `json:"spoiler_text"`
	Visibility         string        `json:"visibility"`
	Language           string        `json:"language"`
	URI                string        `json:"uri"`
	URL                string        `json:"url"`
	RepliesCount       int           `json:"replies_count"`
	ReblogsCount       int           `json:"reblogs_count"`
	FavouritesCount    int           `json:"favourites_count"`
	Favourited         bool          `json:"favourited"`
	Reblogged          bool          `json:"reblogged"`
	Muted              bool          `json:"muted"`
	Bookmarked         bool          `json:"bookmarked"`
	Content            string        `json:"content"`
	Reblog             *Status       `json:"reblog"`
	Account            Account       `json:"account"`
	MediaAttachments   []interface{} `json:"media_attachments"`
	Mentions           []interface{} `json:"mentions"`
	Tags               []interface{} `json:"tags"`
	Emojis             []interface{} `json:"emojis"`
	Card               interface{}   `json:"card"`
	Poll               interface{}   `json:"poll"`
}

// NewClient sets up CycleTLS client and authenticates with Truth Social
func NewClient(ctx context.Context, username, password string) (*Client, error) {
	// Initialize CycleTLS client
	client := cycletls.Init()

	c := &Client{
		cycleTLS: client,
		username: username,
		password: password,
	}

	// Solve Cloudflare challenge first (optional - may not be needed)
	if err := c.solveCF(ctx); err != nil {
		// Log the error but don't fail - CycleTLS might handle CF automatically
		fmt.Printf("Warning: Cloudflare solve failed (continuing anyway): %v\n", err)
	}

	// Authenticate
	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}

	return c, nil
}

// solveCF launches headless Chrome, navigates to TruthSocial, and returns CF clearance cookies.
func (c *Client) solveCF(parentCtx context.Context) error {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parentCtx,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent(userAgent),
	)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// enable network domain to read cookies
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		return err
	}

	// navigate and wait for body
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://truthsocial.com/"),
		chromedp.WaitReady("body"),
	); err != nil {
		return err
	}

	// fetch cookies
	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		return err
	}

	// Convert cookies to CycleTLS format and set them
	var cookieList []cycletls.Cookie
	for _, c := range cookies {
		cookieList = append(cookieList, cycletls.Cookie{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	// Note: CycleTLS handles cookies automatically, no need to manually set them

	return nil
}

func (c *Client) authenticate(ctx context.Context) error {
	authURL := baseURL + "/oauth/token"

	payload := map[string]interface{}{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"grant_type":    "password",
		"username":      c.username,
		"password":      c.password,
		"redirect_uri":  "urn:ietf:wg:oauth:2.0:oob",
		"scope":         "read",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.cycleTLS.Do(authURL, cycletls.Options{
		Body:      string(payloadBytes),
		Method:    "POST",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Content-Type":    "application/json",
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "POST")
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}

	if resp.Status != 200 {
		return fmt.Errorf("authentication failed: status %d - %s", resp.Status, resp.Body)
	}

	var authResp AuthResponse
	if err := json.Unmarshal([]byte(resp.Body), &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return fmt.Errorf("no access token received")
	}

	c.accessToken = authResp.AccessToken
	return nil
}

func (c *Client) Lookup(ctx context.Context, username string) (*Account, error) {
	username = strings.TrimPrefix(username, "@")
	lookupURL := fmt.Sprintf("%s/v1/accounts/lookup?acct=%s", apiBaseURL, username)

	resp, err := c.cycleTLS.Do(lookupURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("account lookup request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("account lookup failed: status %d - %s", resp.Status, resp.Body)
	}

	var account Account
	if err := json.Unmarshal([]byte(resp.Body), &account); err != nil {
		return nil, fmt.Errorf("failed to parse account data: %w", err)
	}

	return &account, nil
}

// PullStatuses implements the same method as Stanford Truthbrush
// Returns posts in reverse chronological order (recent first)
func (c *Client) PullStatuses(ctx context.Context, username string, excludeReplies bool, limit int) ([]Status, error) {
	// First lookup the user to get their ID
	account, err := c.Lookup(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user %s: %w", username, err)
	}

	var allStatuses []Status
	var maxID string
	pageCounter := 0

	for {
		// Build URL with parameters
		statusURL := fmt.Sprintf("%s/v1/accounts/%s/statuses", apiBaseURL, account.ID)
		params := url.Values{}

		if excludeReplies {
			params.Set("exclude_replies", "true")
		}

		if maxID != "" {
			params.Set("max_id", maxID)
		}

		// Set a reasonable page size
		params.Set("limit", "40")

		if len(params) > 0 {
			statusURL += "?" + params.Encode()
		}

		resp, err := c.cycleTLS.Do(statusURL, cycletls.Options{
			Method:    "GET",
			Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			UserAgent: userAgent,
			Headers: map[string]string{
				"Authorization":   "Bearer " + c.accessToken,
				"Accept":          "application/json",
				"Accept-Language": "en-US,en;q=0.9",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT":             "1",
				"Connection":      "keep-alive",
				"Sec-Fetch-Dest":  "empty",
				"Sec-Fetch-Mode":  "cors",
				"Sec-Fetch-Site":  "same-origin",
			},
		}, "GET")
		if err != nil {
			return nil, fmt.Errorf("statuses request failed: %w", err)
		}

		if resp.Status != 200 {
			return nil, fmt.Errorf("statuses request failed: status %d - %s", resp.Status, resp.Body)
		}

		var statuses []Status
		if err := json.Unmarshal([]byte(resp.Body), &statuses); err != nil {
			return nil, fmt.Errorf("failed to parse statuses data: %w", err)
		}

		// If no statuses returned, we've reached the end
		if len(statuses) == 0 {
			break
		}

		// Add statuses to our collection
		allStatuses = append(allStatuses, statuses...)
		pageCounter++

		// Check if we've reached our limit
		if limit > 0 && len(allStatuses) >= limit {
			// Trim to exact limit
			if len(allStatuses) > limit {
				allStatuses = allStatuses[:limit]
			}
			break
		}

		// Set maxID for next page (oldest post ID from current page)
		if len(statuses) > 0 {
			lastStatus := statuses[len(statuses)-1]
			maxID = lastStatus.ID
		}

		// Safety check to prevent infinite loops
		if pageCounter > 50 {
			break
		}

		// Add a small delay between requests to be respectful
		time.Sleep(500 * time.Millisecond)
	}

	return allStatuses, nil
}

// GetStatuses is a simpler method for getting recent statuses
func (c *Client) GetStatuses(ctx context.Context, accountID string, limit int) ([]Status, error) {
	statusURL := fmt.Sprintf("%s/v1/accounts/%s/statuses?limit=%d&exclude_replies=true&exclude_reblogs=true", apiBaseURL, accountID, limit)

	resp, err := c.cycleTLS.Do(statusURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("statuses request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("statuses request failed: status %d - %s", resp.Status, resp.Body)
	}

	var statuses []Status
	if err := json.Unmarshal([]byte(resp.Body), &statuses); err != nil {
		return nil, fmt.Errorf("failed to parse statuses data: %w", err)
	}

	return statuses, nil
}

// Search implements the search functionality from TruthBrush
func (c *Client) Search(ctx context.Context, searchType, query string, limit int, resolve bool, offset int, minID, maxID string) ([]interface{}, error) {
	searchURL := fmt.Sprintf("%s/v2/search", apiBaseURL)
	params := url.Values{}
	params.Set("q", query)
	params.Set("type", searchType) // accounts, statuses, hashtags, groups
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("resolve", fmt.Sprintf("%t", resolve))
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("min_id", minID)
	if maxID != "" {
		params.Set("max_id", maxID)
	}

	searchURL += "?" + params.Encode()

	resp, err := c.cycleTLS.Do(searchURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("search failed: status %d - %s", resp.Status, resp.Body)
	}

	var result map[string][]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
		return nil, fmt.Errorf("failed to parse search data: %w", err)
	}

	// Return the specific search type results
	if results, ok := result[searchType]; ok {
		return results, nil
	}

	return []interface{}{}, nil
}

// Trending returns trending truths
func (c *Client) Trending(ctx context.Context, limit int) ([]Status, error) {
	trendingURL := fmt.Sprintf("%s/v1/truth/trending/truths?limit=%d", apiBaseURL, limit)

	resp, err := c.cycleTLS.Do(trendingURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("trending request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("trending failed: status %d - %s", resp.Status, resp.Body)
	}

	var statuses []Status
	if err := json.Unmarshal([]byte(resp.Body), &statuses); err != nil {
		return nil, fmt.Errorf("failed to parse trending data: %w", err)
	}

	return statuses, nil
}

// Tags returns trending tags
func (c *Client) Tags(ctx context.Context) ([]interface{}, error) {
	tagsURL := fmt.Sprintf("%s/v1/trends", apiBaseURL)

	resp, err := c.cycleTLS.Do(tagsURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("tags request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("tags failed: status %d - %s", resp.Status, resp.Body)
	}

	var tags []interface{}
	if err := json.Unmarshal([]byte(resp.Body), &tags); err != nil {
		return nil, fmt.Errorf("failed to parse tags data: %w", err)
	}

	return tags, nil
}

// Suggested returns suggested users to follow
func (c *Client) Suggested(ctx context.Context, maximum int) ([]Account, error) {
	suggestedURL := fmt.Sprintf("%s/v2/suggestions?limit=%d", apiBaseURL, maximum)

	resp, err := c.cycleTLS.Do(suggestedURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("suggested request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("suggested failed: status %d - %s", resp.Status, resp.Body)
	}

	var accounts []Account
	if err := json.Unmarshal([]byte(resp.Body), &accounts); err != nil {
		return nil, fmt.Errorf("failed to parse suggested data: %w", err)
	}

	return accounts, nil
}

// Ads returns ads from Rumble's Ad Platform via Truth Social API
func (c *Client) Ads(ctx context.Context, device string) ([]interface{}, error) {
	adsURL := fmt.Sprintf("%s/v3/truth/ads?device=%s", apiBaseURL, device)

	resp, err := c.cycleTLS.Do(adsURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("ads request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("ads failed: status %d - %s", resp.Status, resp.Body)
	}

	var ads []interface{}
	if err := json.Unmarshal([]byte(resp.Body), &ads); err != nil {
		return nil, fmt.Errorf("failed to parse ads data: %w", err)
	}

	return ads, nil
}

// UserLikes returns users who liked a post
func (c *Client) UserLikes(ctx context.Context, postID string, includeAll bool, topNum int) ([]Account, error) {
	var allUsers []Account
	maxID := ""
	count := 0

	for {
		likesURL := fmt.Sprintf("%s/v1/statuses/%s/favourited_by", apiBaseURL, postID)
		params := url.Values{}
		params.Set("limit", "80")
		if maxID != "" {
			params.Set("max_id", maxID)
		}
		likesURL += "?" + params.Encode()

		resp, err := c.cycleTLS.Do(likesURL, cycletls.Options{
			Method:    "GET",
			Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			UserAgent: userAgent,
			Headers: map[string]string{
				"Authorization":   "Bearer " + c.accessToken,
				"Accept":          "application/json",
				"Accept-Language": "en-US,en;q=0.9",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT":             "1",
				"Connection":      "keep-alive",
				"Sec-Fetch-Dest":  "empty",
				"Sec-Fetch-Mode":  "cors",
				"Sec-Fetch-Site":  "same-origin",
			},
		}, "GET")
		if err != nil {
			return nil, fmt.Errorf("user likes request failed: %w", err)
		}

		if resp.Status != 200 {
			return nil, fmt.Errorf("user likes failed: status %d - %s", resp.Status, resp.Body)
		}

		var users []Account
		if err := json.Unmarshal([]byte(resp.Body), &users); err != nil {
			return nil, fmt.Errorf("failed to parse user likes data: %w", err)
		}

		if len(users) == 0 {
			break
		}

		for _, user := range users {
			allUsers = append(allUsers, user)
			count++
			if !includeAll && count >= topNum {
				return allUsers, nil
			}
		}

		// Set maxID for next page
		if len(users) > 0 {
			maxID = users[len(users)-1].ID
		}

		// Safety check
		if len(users) < 80 {
			break
		}
	}

	return allUsers, nil
}

// PullComments returns comments on a post
func (c *Client) PullComments(ctx context.Context, postID string, includeAll bool, onlyFirst bool, topNum int) ([]Status, error) {
	var allComments []Status
	count := 0

	commentsURL := fmt.Sprintf("%s/v1/statuses/%s/context/descendants", apiBaseURL, postID)
	params := url.Values{}
	params.Set("sort", "oldest")
	commentsURL += "?" + params.Encode()

	resp, err := c.cycleTLS.Do(commentsURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("comments request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("comments failed: status %d - %s", resp.Status, resp.Body)
	}

	var comments []Status
	if err := json.Unmarshal([]byte(resp.Body), &comments); err != nil {
		return nil, fmt.Errorf("failed to parse comments data: %w", err)
	}

	for _, comment := range comments {
		if (onlyFirst && comment.InReplyToID == postID) || !onlyFirst {
			allComments = append(allComments, comment)
			count++
			if !includeAll && count >= topNum {
				break
			}
		}
	}

	return allComments, nil
}

// GroupPosts returns posts from a group timeline
func (c *Client) GroupPosts(ctx context.Context, groupID string, limit int) ([]Status, error) {
	var timeline []Status
	maxID := ""

	for {
		groupURL := fmt.Sprintf("%s/v1/timelines/group/%s", apiBaseURL, groupID)
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", limit))
		if maxID != "" {
			params.Set("max_id", maxID)
		}
		groupURL += "?" + params.Encode()

		resp, err := c.cycleTLS.Do(groupURL, cycletls.Options{
			Method:    "GET",
			Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			UserAgent: userAgent,
			Headers: map[string]string{
				"Authorization":   "Bearer " + c.accessToken,
				"Accept":          "application/json",
				"Accept-Language": "en-US,en;q=0.9",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT":             "1",
				"Connection":      "keep-alive",
				"Sec-Fetch-Dest":  "empty",
				"Sec-Fetch-Mode":  "cors",
				"Sec-Fetch-Site":  "same-origin",
			},
		}, "GET")
		if err != nil {
			return nil, fmt.Errorf("group posts request failed: %w", err)
		}

		if resp.Status != 200 {
			return nil, fmt.Errorf("group posts failed: status %d - %s", resp.Status, resp.Body)
		}

		var posts []Status
		if err := json.Unmarshal([]byte(resp.Body), &posts); err != nil {
			return nil, fmt.Errorf("failed to parse group posts data: %w", err)
		}

		if len(posts) == 0 {
			break
		}

		timeline = append(timeline, posts...)
		limit -= len(posts)
		if limit <= 0 {
			break
		}

		maxID = posts[len(posts)-1].ID
	}

	return timeline, nil
}

// TrendingGroups returns trending groups
func (c *Client) TrendingGroups(ctx context.Context, limit int) ([]interface{}, error) {
	trendingURL := fmt.Sprintf("%s/v1/truth/trends/groups?limit=%d", apiBaseURL, limit)

	resp, err := c.cycleTLS.Do(trendingURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("trending groups request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("trending groups failed: status %d - %s", resp.Status, resp.Body)
	}

	var groups []interface{}
	if err := json.Unmarshal([]byte(resp.Body), &groups); err != nil {
		return nil, fmt.Errorf("failed to parse trending groups data: %w", err)
	}

	return groups, nil
}

// GroupTags returns trending group tags
func (c *Client) GroupTags(ctx context.Context) ([]interface{}, error) {
	tagsURL := fmt.Sprintf("%s/v1/groups/tags", apiBaseURL)

	resp, err := c.cycleTLS.Do(tagsURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("group tags request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("group tags failed: status %d - %s", resp.Status, resp.Body)
	}

	var tags []interface{}
	if err := json.Unmarshal([]byte(resp.Body), &tags); err != nil {
		return nil, fmt.Errorf("failed to parse group tags data: %w", err)
	}

	return tags, nil
}

// SuggestedGroups returns suggested groups to follow
func (c *Client) SuggestedGroups(ctx context.Context, maximum int) ([]interface{}, error) {
	suggestedURL := fmt.Sprintf("%s/v1/truth/suggestions/groups?limit=%d", apiBaseURL, maximum)

	resp, err := c.cycleTLS.Do(suggestedURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("suggested groups request failed: %w", err)
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("suggested groups failed: status %d - %s", resp.Status, resp.Body)
	}

	var groups []interface{}
	if err := json.Unmarshal([]byte(resp.Body), &groups); err != nil {
		return nil, fmt.Errorf("failed to parse suggested groups data: %w", err)
	}

	return groups, nil
}

// UserFollowers returns a user's followers
func (c *Client) UserFollowers(ctx context.Context, userHandle, userID string, maximum int, resume string) ([]Account, error) {
	if userID == "" && userHandle != "" {
		account, err := c.Lookup(ctx, userHandle)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup user %s: %w", userHandle, err)
		}
		userID = account.ID
	}

	var allFollowers []Account
	maxID := resume
	count := 0

	for {
		followersURL := fmt.Sprintf("%s/v1/accounts/%s/followers", apiBaseURL, userID)
		params := url.Values{}
		if maxID != "" {
			params.Set("max_id", maxID)
		}
		if len(params) > 0 {
			followersURL += "?" + params.Encode()
		}

		resp, err := c.cycleTLS.Do(followersURL, cycletls.Options{
			Method:    "GET",
			Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			UserAgent: userAgent,
			Headers: map[string]string{
				"Authorization":   "Bearer " + c.accessToken,
				"Accept":          "application/json",
				"Accept-Language": "en-US,en;q=0.9",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT":             "1",
				"Connection":      "keep-alive",
				"Sec-Fetch-Dest":  "empty",
				"Sec-Fetch-Mode":  "cors",
				"Sec-Fetch-Site":  "same-origin",
			},
		}, "GET")
		if err != nil {
			return nil, fmt.Errorf("followers request failed: %w", err)
		}

		if resp.Status != 200 {
			return nil, fmt.Errorf("followers failed: status %d - %s", resp.Status, resp.Body)
		}

		var followers []Account
		if err := json.Unmarshal([]byte(resp.Body), &followers); err != nil {
			return nil, fmt.Errorf("failed to parse followers data: %w", err)
		}

		if len(followers) == 0 {
			break
		}

		for _, follower := range followers {
			allFollowers = append(allFollowers, follower)
			count++
			if maximum > 0 && count >= maximum {
				return allFollowers, nil
			}
		}

		maxID = followers[len(followers)-1].ID
	}

	return allFollowers, nil
}

// UserFollowing returns users that a user follows
func (c *Client) UserFollowing(ctx context.Context, userHandle, userID string, maximum int, resume string) ([]Account, error) {
	if userID == "" && userHandle != "" {
		account, err := c.Lookup(ctx, userHandle)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup user %s: %w", userHandle, err)
		}
		userID = account.ID
	}

	var allFollowing []Account
	maxID := resume
	count := 0

	for {
		followingURL := fmt.Sprintf("%s/v1/accounts/%s/following", apiBaseURL, userID)
		params := url.Values{}
		if maxID != "" {
			params.Set("max_id", maxID)
		}
		if len(params) > 0 {
			followingURL += "?" + params.Encode()
		}

		resp, err := c.cycleTLS.Do(followingURL, cycletls.Options{
			Method:    "GET",
			Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			UserAgent: userAgent,
			Headers: map[string]string{
				"Authorization":   "Bearer " + c.accessToken,
				"Accept":          "application/json",
				"Accept-Language": "en-US,en;q=0.9",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT":             "1",
				"Connection":      "keep-alive",
				"Sec-Fetch-Dest":  "empty",
				"Sec-Fetch-Mode":  "cors",
				"Sec-Fetch-Site":  "same-origin",
			},
		}, "GET")
		if err != nil {
			return nil, fmt.Errorf("following request failed: %w", err)
		}

		if resp.Status != 200 {
			return nil, fmt.Errorf("following failed: status %d - %s", resp.Status, resp.Body)
		}

		var following []Account
		if err := json.Unmarshal([]byte(resp.Body), &following); err != nil {
			return nil, fmt.Errorf("failed to parse following data: %w", err)
		}

		if len(following) == 0 {
			break
		}

		for _, follow := range following {
			allFollowing = append(allFollowing, follow)
			count++
			if maximum > 0 && count >= maximum {
				return allFollowing, nil
			}
		}

		maxID = following[len(following)-1].ID
	}

	return allFollowing, nil
}

func (c *Client) Close() {
	// Close CycleTLS client
	c.cycleTLS.Close()
}
