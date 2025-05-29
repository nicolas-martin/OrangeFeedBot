package truthsocial

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const baseURL = "https://truthsocial.com/api/v1"

type Client struct {
	http      *http.Client
	user      string
	pass      string
	csrfToken string
}

type Status struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Account   struct {
		Username string `json:"username"`
	} `json:"account"`
	URL string `json:"url"`
}

type Account struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func NewClient(ctx context.Context, user, pass string) (*Client, error) {
	jar, _ := cookiejar.New(nil)
	c := &Client{
		http: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
		user: user,
		pass: pass,
	}

	// First get the main page to establish session and get CSRF token
	if err := c.initSession(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize session: %w", err)
	}

	if err := c.Login(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) initSession(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://truthsocial.com", nil)
	c.setCommonHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get main page: %w", err)
	}
	defer resp.Body.Close()

	// Try to extract CSRF token from response
	body, _ := c.readResponseBody(resp)
	bodyStr := string(body)

	// Look for CSRF token in meta tag
	if start := strings.Index(bodyStr, `name="csrf-token" content="`); start != -1 {
		start += len(`name="csrf-token" content="`)
		if end := strings.Index(bodyStr[start:], `"`); end != -1 {
			c.csrfToken = bodyStr[start : start+end]
		}
	}

	return nil
}

func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	if c.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
	}
}

func (c *Client) Login(ctx context.Context) error {
	// Try form-based login first
	formData := url.Values{}
	formData.Set("user[email]", c.user)
	formData.Set("user[password]", c.pass)

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://truthsocial.com/auth/sign_in", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setCommonHeaders(req)
	req.Header.Set("Referer", "https://truthsocial.com/auth/sign_in")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("form auth request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if form login worked (should redirect or return 200)
	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return nil
	}

	// If form login failed, try API login
	payload := map[string]string{"username": c.user, "password": c.pass}
	body, _ := json.Marshal(payload)

	req, _ = http.NewRequestWithContext(ctx, "POST", baseURL+"/auth/sign_in", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.setCommonHeaders(req)

	resp, err = c.http.Do(req)
	if err != nil {
		return fmt.Errorf("API auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := c.readResponseBody(resp)
		return fmt.Errorf("auth error: %s - %s", resp.Status, string(bodyBytes))
	}
	return nil
}

func (c *Client) GetAccount(ctx context.Context, username string) (*Account, error) {
	url := fmt.Sprintf("%s/accounts/lookup?acct=%s", baseURL, username)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	c.setCommonHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := c.readResponseBody(resp)
		return nil, fmt.Errorf("account lookup error: %s - %s", resp.Status, string(bodyBytes))
	}

	var account Account
	bodyBytes, err := c.readResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

func (c *Client) GetStatuses(ctx context.Context, accountID string, limit int) ([]Status, error) {
	url := fmt.Sprintf("%s/accounts/%s/statuses?limit=%d", baseURL, accountID, limit)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	c.setCommonHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := c.readResponseBody(resp)
		return nil, fmt.Errorf("statuses error: %s - %s", resp.Status, string(bodyBytes))
	}

	var statuses []Status
	bodyBytes, err := c.readResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, &statuses); err != nil {
		return nil, err
	}
	return statuses, nil
}

func (c *Client) readResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body

	// Check if response is gzipped
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
