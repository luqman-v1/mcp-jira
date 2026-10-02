# mcp-jira

A high-performance Model Context Protocol (MCP) server written in Go for Atlassian Jira, running over `stdio`.

## Features

- **`get_jira_issue`**: Fetch issue details by issue key (e.g. `PROJ-123`) or full browser URL (`https://your-domain.atlassian.net/browse/PROJ-123`). Formats summary, status, assignee, priority, timestamps, description, and comments. Supports both plain text and Atlassian Document Format (ADF).
- **`search_jira_issues`**: Run queries using Jira Query Language (JQL) and retrieve matching issues with metadata.
- **Resilient Config**: Automatically reads from environment variables (`JIRA_URL`, `JIRA_EMAIL`, `JIRA_TOKEN`) or falls back to `.env` files.

## Installation

```bash
make install
```
This builds and places the binary at `~/go/bin/mcp-jira`.

## Configuration

### Environment Variables

| Variable | Description | Example |
|---|---|---|
| `JIRA_URL` | Base Jira Cloud/Server URL | `https://your-domain.atlassian.net` |
| `JIRA_EMAIL` | Account email address | `your-email@example.com` |
| `JIRA_TOKEN` | API token / PAT | `<YOUR_JIRA_API_TOKEN>` |

### MCP Client Setup (`mcp.json`)

```json
{
  "mcpServers": {
    "mcp-jira": {
      "type": "stdio",
      "command": "/path/to/go/bin/mcp-jira",
      "env": {
        "JIRA_URL": "https://your-domain.atlassian.net",
        "JIRA_EMAIL": "your-email@example.com",
        "JIRA_TOKEN": "<YOUR_JIRA_API_TOKEN>"
      }
    }
  }
}
```
