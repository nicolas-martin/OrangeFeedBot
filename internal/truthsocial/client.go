package truthsocial

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"
)

const (
	baseURL    = "https://truthsocial.com"
	apiBaseURL = "https://truthsocial.com/api"

	// OAuth client credentials extracted from Truth Social's JavaScript
	// These are the same credentials used by Stanford Truthbrush
	clientID     = "9X1Fdd-pxNsAgEDNi_SfhJWi8T-vLuV2WVzKIbkTCw4"
	clientSecret = "ozF8jzI4968oTKFkEnsBC-UbLPCdrSv0MkXGQu2o_-M"

	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
)

type Client struct {
	http        *http.Client
	user        string
	pass        string
	accessToken string
}

type Status struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Account   struct {
		Username string `json:"username"`
	} `json:"account"`
	URL             string `json:"url"`
	InReplyToID     string `json:"in_reply_to_id"`
	ReblogsCount    int    `json:"reblogs_count"`
	FavouritesCount int    `json:"favourites_count"`
	RepliesCount    int    `json:"replies_count"`
}

type Account struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	DisplayName    string `json:"display_name"`
	FollowersCount int    `json:"followers_count"`
	StatusesCount  int    `json:"statuses_count"`
	Verified       bool   `json:"verified"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

// Custom dialer that uses uTLS to mimic Chrome's TLS fingerprint
func createUTLSDialer() func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		// Create a regular TCP connection
		conn, err := net.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		// Parse the address to get the hostname
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			conn.Close()
			return nil, err
		}

		// Create uTLS connection with Chrome fingerprint
		uConn := utls.UClient(conn, &utls.Config{
			ServerName:         host,
			InsecureSkipVerify: false,
		}, utls.HelloChrome_Auto)

		// Perform the TLS handshake
		err = uConn.Handshake()
		if err != nil {
			conn.Close()
			return nil, err
		}

		return uConn, nil
	}
}

func NewClient(ctx context.Context, user, pass string) (*Client, error) {
	jar, _ := cookiejar.New(nil)

	// Create custom transport with uTLS
	transport := &http.Transport{
		Dial:                createUTLSDialer(),
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	c := &Client{
		http: &http.Client{
			Jar:       jar,
			Timeout:   45 * time.Second,
			Transport: transport,
		},
		user: user,
		pass: pass,
	}

	// Authenticate using the same method as Truthbrush
	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return c, nil
}

func (c *Client) authenticate(ctx context.Context) error {
	// Use the same authentication approach as Stanford Truthbrush
	url := baseURL + "/oauth/token"

	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"grant_type":    "password",
		"username":      c.user,
		"password":      c.pass,
		"redirect_uri":  "urn:ietf:wg:oauth:2.0:oob",
		"scope":         "read",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := c.readResponseBody(resp)
		return fmt.Errorf("authentication failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var authResp AuthResponse
	bodyBytes, err := c.readResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return fmt.Errorf("no access token received")
	}

	c.accessToken = authResp.AccessToken
	return nil
}

func (c *Client) setStandardHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
}

func (c *Client) Lookup(ctx context.Context, username string) (*Account, error) {
	username = strings.TrimPrefix(username, "@")
	url := fmt.Sprintf("%s/v1/accounts/lookup?acct=%s", apiBaseURL, username)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	c.setStandardHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("account lookup request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := c.readResponseBody(resp)
		return nil, fmt.Errorf("account lookup failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var account Account
	bodyBytes, err := c.readResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read account response: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, &account); err != nil {
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

		req, _ := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
		c.setStandardHeaders(req)

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("statuses request failed: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := c.readResponseBody(resp)
			resp.Body.Close()
			return nil, fmt.Errorf("statuses request failed: %s - %s", resp.Status, string(bodyBytes))
		}

		var statuses []Status
		bodyBytes, err := c.readResponseBody(resp)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to read statuses response: %w", err)
		}

		if err := json.Unmarshal(bodyBytes, &statuses); err != nil {
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
	}

	return allStatuses, nil
}

// GetStatuses is a simpler method for getting recent statuses
func (c *Client) GetStatuses(ctx context.Context, accountID string, limit int) ([]Status, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s/statuses?limit=%d&exclude_replies=true&exclude_reblogs=true", apiBaseURL, accountID, limit)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	c.setStandardHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("statuses request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := c.readResponseBody(resp)
		return nil, fmt.Errorf("statuses request failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var statuses []Status
	bodyBytes, err := c.readResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read statuses response: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, &statuses); err != nil {
		return nil, fmt.Errorf("failed to parse statuses data: %w", err)
	}

	return statuses, nil
}

func (c *Client) readResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body

	// Handle gzip compression
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	return io.ReadAll(reader)
}
