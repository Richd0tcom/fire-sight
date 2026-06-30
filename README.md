# Firesyte

Dead code heatmap and repository activity analyzer. Fire‑Sight clones a Git repository, analyzes commit history, computes file and function level “heat” scores, and serves a tree-structured API suitable for a frontend to visualize hotspots and potential dead code.

---

## Contents
- **Overview**
- **Features**
- **Tech Stack**
- **Getting Started**
- **Configuration**
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
