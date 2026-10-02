# Changelog

All notable changes to this project will be documented in this file.

## [0.0.1] - 2026-10-02

### Added
- Initial release of `mcp-jira` stdio MCP server.
- `get_jira_issue`: Fetch issue by key or full URL with support for Atlassian Document Format (ADF) and plain text.
- `search_jira_issues`: JQL issue search with configurable result limits.
- Resilient credentials loading from environment variables and `.env` fallback.
