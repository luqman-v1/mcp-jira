package jira

import "time"

// Issue represents a parsed Jira issue.
type Issue struct {
	Key         string    `json:"key"`
	URL         string    `json:"url"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	IssueType   string    `json:"issue_type"`
	Priority    string    `json:"priority"`
	Assignee    string    `json:"assignee"`
	Reporter    string    `json:"reporter"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Comments    []Comment `json:"comments"`
}

// Comment represents a single Jira comment.
type Comment struct {
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// SearchResultItem represents a summary item in a search result.
type SearchResultItem struct {
	Key       string `json:"key"`
	URL       string `json:"url"`
	Summary   string `json:"summary"`
	Status    string `json:"status"`
	IssueType string `json:"issue_type"`
	Assignee  string `json:"assignee"`
	Priority  string `json:"priority"`
	Updated   string `json:"updated"`
}

// rawIssueResponse is the Jira REST API response structure.
type rawIssueResponse struct {
	Key    string `json:"key"`
	Fields struct {
		Summary   string      `json:"summary"`
		Description interface{} `json:"description"`
		Status    struct {
			Name string `json:"name"`
		} `json:"status"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Assignee *struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Reporter *struct {
			DisplayName string `json:"displayName"`
		} `json:"reporter"`
		Created string `json:"created"`
		Updated string `json:"updated"`
		Comment struct {
			Comments []struct {
				Author struct {
					DisplayName string `json:"displayName"`
				} `json:"author"`
				Body    interface{} `json:"body"`
				Created string      `json:"created"`
			} `json:"comments"`
		} `json:"comment"`
	} `json:"fields"`
}

// rawSearchResponse is the Jira search API response structure.
type rawSearchResponse struct {
	Issues []rawIssueResponse `json:"issues"`
	Total  int                `json:"total"`
}
