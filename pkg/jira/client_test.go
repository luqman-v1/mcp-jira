package jira_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luqman-v1/mcp-jira/pkg/jira"
)

func TestExtractIssueKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{input: "PROJ-123", expected: "PROJ-123", wantErr: false},
		{input: "proj-123", expected: "PROJ-123", wantErr: false},
		{input: "https://your-domain.atlassian.net/browse/PROJ-123", expected: "PROJ-123", wantErr: false},
		{input: "https://your-domain.atlassian.net/browse/PROJ-123?focusedCommentId=123", expected: "PROJ-123", wantErr: false},
		{input: "ABC-1", expected: "ABC-1", wantErr: false},
		{input: "", expected: "", wantErr: true},
		{input: "invalid-url-with-no-key", expected: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := jira.ExtractIssueKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExtractIssueKey(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.expected {
				t.Errorf("ExtractIssueKey(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestClient_GetIssue(t *testing.T) {
	mockJSON := `{
		"key": "PROJ-123",
		"fields": {
			"summary": "Fix login issue",
			"description": "Some description text",
			"status": {"name": "In Progress"},
			"issuetype": {"name": "Bug"},
			"priority": {"name": "Medium"},
			"assignee": {"displayName": "John Doe"},
			"reporter": {"displayName": "Jane Smith"},
			"created": "2026-09-25T13:34:35.277+0700",
			"updated": "2026-10-02T09:11:47.429+0700",
			"comment": {
				"comments": [
					{
						"author": {"displayName": "Bob Builder"},
						"body": "Working on it",
						"created": "2026-09-26T10:00:00.000+0700"
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/issue/PROJ-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	client := jira.NewClient(server.URL, "test@example.com", "secret")
	issue, err := client.GetIssue(context.Background(), "PROJ-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if issue.Key != "PROJ-123" {
		t.Errorf("expected key PROJ-123, got %s", issue.Key)
	}
	if issue.Summary != "Fix login issue" {
		t.Errorf("unexpected summary: %s", issue.Summary)
	}
	if issue.Assignee != "John Doe" {
		t.Errorf("expected assignee John Doe, got %s", issue.Assignee)
	}
	if len(issue.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(issue.Comments))
	}
	if issue.Comments[0].Author != "Bob Builder" {
		t.Errorf("expected comment author Bob Builder, got %s", issue.Comments[0].Author)
	}
}

func TestClient_SearchIssues(t *testing.T) {
	mockJSON := `{
		"issues": [
			{
				"key": "PROJ-123",
				"fields": {
					"summary": "Fix login issue",
					"status": {"name": "In Progress"},
					"issuetype": {"name": "Bug"},
					"priority": {"name": "High"},
					"assignee": {"displayName": "John Doe"},
					"updated": "2026-10-02"
				}
			}
		],
		"total": 1
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/search/jql" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	client := jira.NewClient(server.URL, "test@example.com", "secret")
	issues, err := client.SearchIssues(context.Background(), "project = PROJ", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Key != "PROJ-123" {
		t.Errorf("expected key PROJ-123, got %s", issues[0].Key)
	}
}
