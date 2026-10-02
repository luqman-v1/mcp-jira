package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/luqman-v1/mcp-jira/pkg/jira"
)

var version = "0.0.1"

func main() {
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("mcp-jira version %s\n", version)
		os.Exit(0)
	}

	mcpServer := server.NewMCPServer(
		"mcp-jira",
		version,
		server.WithToolCapabilities(true),
	)

	// Tool 1: get_jira_issue
	getIssueTool := mcp.NewTool("get_jira_issue",
		mcp.WithDescription("Fetch details of a Jira issue by its key (e.g. 'PROJ-123') or full URL (e.g. 'https://your-domain.atlassian.net/browse/PROJ-123')."),
		mcp.WithString("issue",
			mcp.Required(),
			mcp.Description("The Jira issue key (e.g. 'PROJ-123') or full browser URL."),
		),
	)
	mcpServer.AddTool(getIssueTool, handleGetJiraIssue)

	// Tool 2: search_jira_issues
	searchTool := mcp.NewTool("search_jira_issues",
		mcp.WithDescription("Search Jira issues using JQL (Jira Query Language). E.g. 'project = PROJ AND statusCategory != Done ORDER BY updated DESC'."),
		mcp.WithString("jql",
			mcp.Required(),
			mcp.Description("The JQL query string to execute."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of issues to return (default: 10, max: 50)."),
		),
	)
	mcpServer.AddTool(searchTool, handleSearchJiraIssues)

	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server stdio error: %v\n", err)
		os.Exit(1)
	}
}

func getJiraClient() (*jira.Client, error) {
	jiraURL := os.Getenv("JIRA_URL")
	jiraEmail := os.Getenv("JIRA_EMAIL")
	jiraToken := os.Getenv("JIRA_TOKEN")

	// If missing, check fallback .env files
	if jiraURL == "" || jiraEmail == "" || jiraToken == "" {
		home, _ := os.UserHomeDir()
		candidates := []string{
			".env",
			filepath.Join(home, ".config", "jira", ".env"),
		}
		for _, path := range candidates {
			if envMap, err := parseDotEnv(path); err == nil {
				if jiraURL == "" {
					jiraURL = envMap["JIRA_URL"]
				}
				if jiraEmail == "" {
					jiraEmail = envMap["JIRA_EMAIL"]
				}
				if jiraToken == "" {
					jiraToken = envMap["JIRA_TOKEN"]
				}
				if jiraURL != "" && jiraEmail != "" && jiraToken != "" {
					break
				}
			}
		}
	}

	if jiraURL == "" {
		return nil, fmt.Errorf("JIRA_URL is not configured. Please set JIRA_URL in your environment or MCP config.")
	}
	if jiraToken == "" {
		return nil, fmt.Errorf("JIRA_TOKEN is not configured. Please set JIRA_TOKEN in your environment or MCP config.")
	}

	return jira.NewClient(jiraURL, jiraEmail, jiraToken), nil
}

func parseDotEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			res[k] = v
		}
	}
	return res, scanner.Err()
}

func handleGetJiraIssue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issueParam := mcp.ParseString(req, "issue", "")
	if issueParam == "" {
		return mcp.NewToolResultError("Argument 'issue' is required"), nil
	}

	client, err := getJiraClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	issue, err := client.GetIssue(ctx, issueParam)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch Jira issue %s: %v", issueParam, err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# [%s] %s\n\n", issue.Key, issue.Summary))
	sb.WriteString(fmt.Sprintf("**URL:** %s\n\n", issue.URL))

	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|---|---|\n")
	sb.WriteString(fmt.Sprintf("| **Type** | %s |\n", issue.IssueType))
	sb.WriteString(fmt.Sprintf("| **Status** | %s |\n", issue.Status))
	sb.WriteString(fmt.Sprintf("| **Priority** | %s |\n", issue.Priority))
	sb.WriteString(fmt.Sprintf("| **Assignee** | %s |\n", defaultVal(issue.Assignee, "Unassigned")))
	sb.WriteString(fmt.Sprintf("| **Reporter** | %s |\n", defaultVal(issue.Reporter, "Unknown")))
	if !issue.Created.IsZero() {
		sb.WriteString(fmt.Sprintf("| **Created** | %s |\n", issue.Created.Format("2006-01-02 15:04:05 MST")))
	}
	if !issue.Updated.IsZero() {
		sb.WriteString(fmt.Sprintf("| **Updated** | %s |\n", issue.Updated.Format("2006-01-02 15:04:05 MST")))
	}
	sb.WriteString("\n## Description\n\n")
	if strings.TrimSpace(issue.Description) != "" {
		sb.WriteString(issue.Description)
	} else {
		sb.WriteString("_No description provided._")
	}
	sb.WriteString("\n\n")

	if len(issue.Comments) > 0 {
		sb.WriteString("## Comments\n\n")
		for _, c := range issue.Comments {
			timeStr := ""
			if !c.CreatedAt.IsZero() {
				timeStr = fmt.Sprintf(" (%s)", c.CreatedAt.Format("2006-01-02 15:04"))
			}
			sb.WriteString(fmt.Sprintf("### %s%s\n\n", c.Author, timeStr))
			sb.WriteString(c.Body)
			sb.WriteString("\n\n---\n\n")
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleSearchJiraIssues(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jql := mcp.ParseString(req, "jql", "")
	if jql == "" {
		return mcp.NewToolResultError("Argument 'jql' is required"), nil
	}
	limit := mcp.ParseInt(req, "limit", 10)
	if limit <= 0 {
		limit = 10
	}

	client, err := getJiraClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	results, err := client.SearchIssues(ctx, jql, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Jira search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No issues found matching JQL: `%s`", jql)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### Jira Search Results (%d issues)\n\n", len(results)))
	sb.WriteString(fmt.Sprintf("**Query:** `%s`\n\n", jql))
	sb.WriteString("| Key | Type | Status | Assignee | Summary |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	for _, item := range results {
		sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s | %s | %s |\n",
			item.Key,
			item.URL,
			item.IssueType,
			item.Status,
			defaultVal(item.Assignee, "-"),
			sanitizeSummary(item.Summary),
		))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func defaultVal(val, def string) string {
	if strings.TrimSpace(val) == "" {
		return def
	}
	return val
}

func sanitizeSummary(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
