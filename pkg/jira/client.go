package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var issueKeyRegex = regexp.MustCompile(`[A-Z0-9]+-\d+`)

// Client handles communication with Jira REST API.
type Client struct {
	baseURL string
	email   string
	token   string
	client  *http.Client
}

// NewClient initializes a new Jira API client.
func NewClient(rawURL, email, token string) *Client {
	baseURL := strings.TrimRight(rawURL, "/")
	return &Client{
		baseURL: baseURL,
		email:   email,
		token:   token,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// ExtractIssueKey extracts a Jira issue key (e.g. SBPD-38292) from either a key or a URL.
func ExtractIssueKey(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("empty issue input")
	}

	upper := strings.ToUpper(trimmed)
	match := issueKeyRegex.FindString(upper)
	if match == "" {
		return "", fmt.Errorf("no valid Jira issue key found in %q", input)
	}

	return match, nil
}

// GetIssue fetches issue details by key or URL.
func (c *Client) GetIssue(ctx context.Context, issueKeyOrURL string) (*Issue, error) {
	key, err := ExtractIssueKey(issueKeyOrURL)
	if err != nil {
		return nil, fmt.Errorf("extract issue key: %w", err)
	}

	reqURL := fmt.Sprintf("%s/rest/api/2/issue/%s?expand=renderedFields", c.baseURL, url.PathEscape(key))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("issue %s not found (HTTP 404)", key)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira returned HTTP %d for issue %s", resp.StatusCode, key)
	}

	var raw rawIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	issue := &Issue{
		Key:         raw.Key,
		URL:         fmt.Sprintf("%s/browse/%s", c.baseURL, raw.Key),
		Summary:     raw.Fields.Summary,
		Description: parseContent(raw.Fields.Description),
		Status:      raw.Fields.Status.Name,
		IssueType:   raw.Fields.IssueType.Name,
		Priority:    raw.Fields.Priority.Name,
		Created:     parseTime(raw.Fields.Created),
		Updated:     parseTime(raw.Fields.Updated),
	}

	if raw.Fields.Assignee != nil {
		issue.Assignee = raw.Fields.Assignee.DisplayName
	}
	if raw.Fields.Reporter != nil {
		issue.Reporter = raw.Fields.Reporter.DisplayName
	}

	for _, cmt := range raw.Fields.Comment.Comments {
		issue.Comments = append(issue.Comments, Comment{
			Author:    cmt.Author.DisplayName,
			Body:      parseContent(cmt.Body),
			CreatedAt: parseTime(cmt.Created),
		})
	}

	return issue, nil
}

// SearchIssues searches Jira issues using JQL syntax.
func (c *Client) SearchIssues(ctx context.Context, jql string, limit int) ([]SearchResultItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Try POST /rest/api/2/search/jql first (Jira Cloud standard)
	reqURL := fmt.Sprintf("%s/rest/api/2/search/jql", c.baseURL)
	payload := map[string]interface{}{
		"jql":        jql,
		"maxResults": limit,
		"fields":     []string{"summary", "status", "issuetype", "assignee", "priority", "updated"},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal search payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create search request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira search returned HTTP %d", resp.StatusCode)
	}

	var searchResp rawSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	results := make([]SearchResultItem, 0, len(searchResp.Issues))
	for _, raw := range searchResp.Issues {
		item := SearchResultItem{
			Key:       raw.Key,
			URL:       fmt.Sprintf("%s/browse/%s", c.baseURL, raw.Key),
			Summary:   raw.Fields.Summary,
			Status:    raw.Fields.Status.Name,
			IssueType: raw.Fields.IssueType.Name,
			Priority:  raw.Fields.Priority.Name,
			Updated:   raw.Fields.Updated,
		}
		if raw.Fields.Assignee != nil {
			item.Assignee = raw.Fields.Assignee.DisplayName
		}
		results = append(results, item)
	}

	return results, nil
}

func (c *Client) setAuthHeaders(req *http.Request) {
	if c.email != "" {
		req.SetBasicAuth(c.email, c.token)
	} else if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func parseTime(tStr string) time.Time {
	if tStr == "" {
		return time.Time{}
	}
	// Jira formats e.g. 2026-09-25T13:34:35.277+0700 or RFC3339
	layouts := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z0700",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, tStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseContent(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case map[string]interface{}:
		var sb strings.Builder
		renderADFNode(v, &sb)
		res := strings.TrimSpace(sb.String())
		if res != "" {
			return res
		}
		b, _ := json.MarshalIndent(v, "", "  ")
		return string(b)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func renderADFNode(node map[string]interface{}, sb *strings.Builder) {
	nodeType, _ := node["type"].(string)

	if nodeType == "text" {
		if text, ok := node["text"].(string); ok {
			sb.WriteString(text)
		}
		return
	}

	content, ok := node["content"].([]interface{})
	if !ok {
		return
	}

	for _, item := range content {
		childMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		childType, _ := childMap["type"].(string)
		switch childType {
		case "paragraph":
			renderADFNode(childMap, sb)
			sb.WriteString("\n\n")
		case "heading":
			level := 1
			if attrs, ok := childMap["attrs"].(map[string]interface{}); ok {
				if lvl, ok := attrs["level"].(float64); ok {
					level = int(lvl)
				}
			}
			sb.WriteString(strings.Repeat("#", level) + " ")
			renderADFNode(childMap, sb)
			sb.WriteString("\n\n")
		case "bulletList", "orderedList":
			renderADFNode(childMap, sb)
		case "listItem":
			sb.WriteString("- ")
			renderADFNode(childMap, sb)
		case "codeBlock":
			sb.WriteString("```\n")
			renderADFNode(childMap, sb)
			sb.WriteString("\n```\n\n")
		default:
			renderADFNode(childMap, sb)
		}
	}
}
