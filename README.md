# mcp-jira

[![CI](https://github.com/luqman-v1/mcp-jira/actions/workflows/ci.yml/badge.svg)](https://github.com/luqman-v1/mcp-jira/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/luqman-v1/mcp-jira)](https://github.com/luqman-v1/mcp-jira/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/luqman-v1/mcp-jira)](https://goreportcard.com/report/github.com/luqman-v1/mcp-jira)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A lightweight, high-performance [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server written in Go for **Atlassian Jira**, communicating over standard I/O (`stdio`).

Seamlessly connect Claude, Cursor, Antigravity, or any MCP-compatible agent to your Jira Cloud or Jira Data Center / Server instance.

---

## Features

- 🔍 **`get_jira_issue`**: Fetch issue details by issue key (`PROJ-123`) or full browser URL (`https://your-domain.atlassian.net/browse/PROJ-123`).
  - Summarizes issue type, status, priority, reporter, assignee, timestamps, and comments.
  - Native parser for both plain-text wiki markup and **Atlassian Document Format (ADF)** JSON trees.
- 🔎 **`search_jira_issues`**: Search issues using standard **Jira Query Language (JQL)** with customizable limits (e.g. `project = PROJ AND statusCategory != Done ORDER BY updated DESC`).
- ⚡ **Zero-dependency Runtime**: Statically compiled binary via `CGO_ENABLED=0` with negligible startup overhead.
- 🛡️ **Defensive & Resilient Configuration**: Reads credentials from MCP `env`, system environment variables, or local `.env` files (`~/.config/jira/.env` or `.env`).

---

## Installation

### Option 1: Download Pre-built Binary
Download the pre-compiled binary for your operating system and architecture from the [Releases page](https://github.com/luqman-v1/mcp-jira/releases).

```bash
# Example for macOS Apple Silicon
curl -sSL https://github.com/luqman-v1/mcp-jira/releases/download/v0.0.1/mcp-jira-v0.0.1-darwin-arm64.tar.gz | tar -xz
sudo mv mcp-jira-v0.0.1-darwin-arm64/mcp-jira /usr/local/bin/
```

### Option 2: Install via Go
```bash
go install github.com/luqman-v1/mcp-jira@latest
```
*(Binary will be placed in `$GOPATH/bin` or `~/go/bin`)*

### Option 3: Build from Source
```bash
git clone https://github.com/luqman-v1/mcp-jira.git
cd mcp-jira
make install
```

---

## Configuration

### Environment Variables

| Variable | Description | Required | Example |
|---|---|:---:|---|
| `JIRA_URL` | Base Jira Cloud or Server URL | Yes | `https://your-domain.atlassian.net` |
| `JIRA_EMAIL` | Account email address (Jira Cloud Basic Auth) | Cloud only | `user@example.com` |
| `JIRA_TOKEN` | Jira API Token (Cloud) or Personal Access Token (Server/DC) | Yes | `<YOUR_API_TOKEN>` |

> **Note on Jira Cloud**: API tokens can be generated at [Atlassian Account API Tokens](https://id.atlassian.com/manage-profile/security/api-tokens).

---

## MCP Client Setup

### Claude Desktop (`claude_desktop_config.json`)
```json
{
  "mcpServers": {
    "jira": {
      "command": "/path/to/mcp-jira",
      "env": {
        "JIRA_URL": "https://your-domain.atlassian.net",
        "JIRA_EMAIL": "your-email@example.com",
        "JIRA_TOKEN": "YOUR_API_TOKEN"
      }
    }
  }
}
```

### Cursor (`~/.cursor/mcp.json`)
```json
{
  "mcpServers": {
    "jira": {
      "command": "/path/to/mcp-jira",
      "env": {
        "JIRA_URL": "https://your-domain.atlassian.net",
        "JIRA_EMAIL": "your-email@example.com",
        "JIRA_TOKEN": "YOUR_API_TOKEN"
      }
    }
  }
}
```

### Omp Agent (`~/.omp/agent/mcp.json`)
```json
{
  "mcpServers": {
    "mcp-jira": {
      "type": "stdio",
      "command": "/path/to/mcp-jira",
      "env": {
        "JIRA_URL": "https://your-domain.atlassian.net",
        "JIRA_EMAIL": "your-email@example.com",
        "JIRA_TOKEN": "YOUR_API_TOKEN"
      }
    }
  }
}
```

---

## Available Tools

### 1. `get_jira_issue`
Retrieves issue details and converts rich ADF content to clean Markdown.

* **Arguments:**
  * `issue` *(string, required)*: Issue key (e.g. `PROJ-123`) or full Jira browser URL (`https://your-domain.atlassian.net/browse/PROJ-123`).

### 2. `search_jira_issues`
Runs a JQL search query and returns matching tickets in a formatted Markdown table.

* **Arguments:**
  * `jql` *(string, required)*: Jira Query Language expression (e.g. `project = PROJ AND statusCategory != Done ORDER BY updated DESC`).
  * `limit` *(number, optional)*: Maximum number of tickets to return (default: `10`, max: `50`).

---

## Development

```bash
# Run unit tests
make test

# Run race detector and vet
make check

# Cross-compile for Darwin, Linux, and Windows
make cross
```

---

## License

This project is licensed under the [MIT License](LICENSE).
