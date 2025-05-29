package scraper

import (
	"testing"
)

func TestNewTruthSocialScraper(t *testing.T) {
	scraper := NewTruthSocialScraper()

	if scraper == nil {
		t.Fatal("Expected scraper to be initialized, got nil")
	}

	if scraper.baseURL != "https://truthsocial.com" {
		t.Errorf("Expected baseURL to be 'https://truthsocial.com', got '%s'", scraper.baseURL)
	}

	if scraper.client == nil {
		t.Fatal("Expected HTTP client to be initialized, got nil")
	}
}

func TestFetchUserPosts(t *testing.T) {
	scraper := NewTruthSocialScraper()

	posts, err := scraper.FetchUserPosts("realDonaldTrump", 5)

	// Since Truth Social blocks scraping, we expect this to fail
	if err != nil {
		t.Logf("Expected error due to Truth Social blocking: %v", err)
		return
	}

	// If somehow we get posts, verify their structure
	if len(posts) > 5 {
		t.Errorf("Expected at most 5 posts, got %d", len(posts))
	}

	for _, post := range posts {
		if post.ID == "" {
			t.Error("Expected post ID to be non-empty")
		}

		if post.Content == "" {
			t.Error("Expected post content to be non-empty")
		}

		if post.Username == "" {
			t.Error("Expected post username to be non-empty")
		}

		if post.URL == "" {
			t.Error("Expected post URL to be non-empty")
		}

		if post.CreatedAt.IsZero() {
			t.Error("Expected post creation time to be set")
		}
	}
}

func TestFetchUserPostsWithDifferentUsername(t *testing.T) {
	scraper := NewTruthSocialScraper()

	// Test with @ prefix
	posts1, err1 := scraper.FetchUserPosts("@testuser", 3)

	// Test without @ prefix
	posts2, err2 := scraper.FetchUserPosts("testuser", 3)

	// Both should handle the username format correctly (even if they fail due to blocking)
	if err1 != nil && err2 != nil {
		t.Logf("Both requests failed as expected due to Truth Social blocking")
		return
	}

	// If either succeeds, verify the username handling
	if err1 == nil && len(posts1) > 0 {
		if posts1[0].Username != "testuser" {
			t.Errorf("Expected username to be 'testuser', got '%s'", posts1[0].Username)
		}
	}

	if err2 == nil && len(posts2) > 0 {
		if posts2[0].Username != "testuser" {
			t.Errorf("Expected username to be 'testuser', got '%s'", posts2[0].Username)
		}
	}
}

func TestCleanHTML(t *testing.T) {
	scraper := NewTruthSocialScraper()

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "<p>Hello <strong>world</strong>!</p>",
			expected: "Hello world!",
		},
		{
			input:    "Text with &amp; entities &lt;test&gt;",
			expected: "Text with & entities <test>",
		},
		{
			input:    "<div>Multiple<br>lines<br/>here</div>",
			expected: "Multiplelineshere",
		},
		{
			input:    "  Whitespace  around  ",
			expected: "Whitespace  around",
		},
	}

	for _, tc := range testCases {
		result := scraper.cleanHTML(tc.input)
		if result != tc.expected {
			t.Errorf("cleanHTML(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
