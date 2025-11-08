# Firesyte

Dead code heatmap and repository activity analyzer. Fire‑Sight clones a Git repository, analyzes commit history, computes file and function level “heat” scores, and serves a tree-structured API suitable for a frontend to visualize hotspots and potential dead code.

---

## Contents
- **Overview**
- **Features**
- **Tech Stack**
- **Getting Started**
- **Configuration**
- **Running**
- **API Reference**
- **Usage Examples**
- **Screenshots (placeholders)**
- **Development**
- **Troubleshooting**
- **Roadmap**
- **License**

---

## Overview
Fire‑Sight analyzes Git history to quantify how “hot” each file/function is based on recency and frequency of changes, plus author diversity. It exposes a small HTTP API consumed by a frontend to render a navigable tree of the repository with heat metrics and file/function summaries.

Use cases:
- **Identify hotspots** that change often and may need refactoring or additional tests.
- **Spot potential dead code** based on inactivity and low change frequency.
- **Guide prioritization** for code cleanup and maintenance.

## Features
- **Repository analysis via Git** using a shallow in-memory data model built from a temporary clone.
- **Heat scoring** for files and functions with exponential time decay and author bonus.
- **Hierarchical tree** of folders/files with aggregated folder metrics.
- **Simple HTTP API** with CORS support for a separate frontend app.
- **Ephemeral storage** in a configurable temp directory.

## Tech Stack
- **Language:** Go (module: `github.com/richd0tcom/fire-sight`)
- **HTTP Router:** `github.com/gorilla/mux`
- **Git:** `github.com/go-git/go-git/v5`
- **Go Version:** declared `go 1.25` in `go.mod`



## Getting Started
### Prerequisites
- Go 1.25+ installed
- Network access to the Git repository you want to analyze

### Install dependencies
Dependencies are managed via Go modules; `go run`/`go build` will resolve them automatically.

## Configuration
Configure via environment variables:
- `PORT` (default `8090`): HTTP port.
- `TEMP_DIR` (default `./tmp/dead-code-heatmap`): Workspace for temporary clones.
- `FRONTEND_URL` (default `http://localhost:8080`): Allowed CORS origin.

Example:
```bash
export PORT=8090
export TEMP_DIR=/tmp/fire-sight
export FRONTEND_URL=http://localhost:5173
```

## Running
### Start the server (development)
```bash
go run ./cmd
```

### Build a binary
```bash
go build -o bin/fire-sight ./cmd
./bin/fire-sight
```

On start, the server logs the port and temp directory and begins serving requests.

## API Reference
Base URL: `http://localhost:<PORT>`

- **Health**
  - `GET /health`
  - Response: `{ "status": "healthy", "time": "<RFC3339>" }`

- **Analyze Repository**
  - `POST /analyze`
  - Content-Type: `application/json`
  - Body:
    ```json
    {
      "repoUrl": "https://github.com/owner/repo.git",
      "branch": "main",
      "timeRangeDays": 180,
      "authToken": "<optional-token>"
    }
    ```
    - `branch` default: `main`
    - `timeRangeDays` default: `180`
    - `authToken` is optional and used for private repositories (sent as basic auth password; do not include in URLs)

  - Success Response (example):
    ```json
    {
      "repoId": "d41d8cd98f00b204",
      "status": "complete",
      "fileTree": {
        "id": "root",
        "name": "root",
        "path": "",
        "type": "folder",
        "children": [
          {
            "id": "internal_analyzer_heat.go",
            "name": "heat.go",
            "path": "internal/analyzer/heat.go",
            "type": "file",
            "extension": "go",
            "linesOfCode": 0,
            "lastModified": "2025-01-01T00:00:00Z",
            "heatScore": {
              "path": "internal/analyzer/heat.go",
              "score": 72.3,
              "changeFreq": 1.8,
              "daysSinceEdit": 12,
              "totalFileChanges": 15
            },
            "functions": [
              {
                "name": "CalculateHeatScores",
                "lineStart": 22,
                "lineEnd": 66,
                "lastModified": "2025-01-01T00:00:00Z",
                "heatScore": {
                  "score": 88.1,
                  "changeFreq": 0.7,
                  "daysSinceEdit": 12
                },
                "isDeadCode": false
              }
            ]
          }
        ]
      },
      "duration": "3.142s"
    }
    ```

  - Error Response (example):
    ```json
    {
      "status": "error",
      "error": "Analysis failed: <message>"
    }
    ```

Notes:
- The server may analyze a large number of files; the default handler timeout is 5 minutes.
- Temporary clones are deleted after analysis completes.

## Usage Examples
Analyze a public repository:
```bash
curl -X POST http://localhost:8090/analyze \
  -H "Content-Type: application/json" \
  -d '{
        "repoUrl": "https://github.com/owner/repo.git",
        "branch": "main",
        "timeRangeDays": 180
      }'
```

Analyze a private repository with a token:
```bash
curl -X POST http://localhost:8090/analyze \
  -H "Content-Type: application/json" \
  -d '{
        "repoUrl": "https://github.com/owner/private-repo.git",
        "branch": "main",
        "timeRangeDays": 90,
        "authToken": "<ghp_xxx>"
      }'
```

Health check:
```bash
curl http://localhost:8090/health
```

## Screenshots (placeholders)
- Frontend heatmap overview

  [Add screenshot here]

- File detail view with function heat

  [Add screenshot here]

## Development
- **CORS:** Origin is controlled by `FRONTEND_URL` in middleware.
- **Tree building:** See `internal/analyzer/tree.go` for how file/folder aggregation and sorting work.
- **Heat algorithm:** See `internal/analyzer/heat.go` for decay, author bonus, and normalization.
- **Git analysis:** See `internal/analyzer/git.go` for cloning, walking, and per-file analysis windowing.

### Local notes
- Default port is `8090`.
- Default time window is `180` days; branch defaults to `main`.
- Function-level metrics are attached when available from `FileAnalyzer` results.

## Troubleshooting
- If requests hang or return 500, verify the repo URL, branch, and network access.
- Increase `timeRangeDays` or adjust repo size to control analysis duration.
- Ensure `TEMP_DIR` is writable; the server creates it if missing.
- For private repos, ensure the token has `repo` access and is valid.

## Roadmap
- Add language-aware LOC and file size metrics.
- Improve dead code detection heuristics and surface at the API level.
- Include commit author summaries per path.
- Add caching for repeated analyses of the same repo/branch.
- Provide Dockerfile and CI workflow.

## License
MIT License. See `LICENSE` for details.