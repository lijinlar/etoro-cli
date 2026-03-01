# Contributing to etoro-cli

Thanks for taking the time to contribute! Whether it's a bug fix, new command, or a doc improvement — all contributions are welcome.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Coding Guidelines](#coding-guidelines)
- [Reporting Bugs](#reporting-bugs)
- [Requesting Features](#requesting-features)

---

## Getting Started

1. **Fork** the repository on GitHub
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/<your-username>/etoro-cli.git
   cd etoro-cli
   ```
3. **Add upstream** remote:
   ```bash
   git remote add upstream https://github.com/lijinlar/etoro-cli.git
   ```

---

## Development Setup

**Prerequisites:**
- Go 1.21+
- `make`

**Build:**
```bash
make build
# Binary output: ./bin/etoro
```

**Run tests:**
```bash
make test
```

**Lint:**
```bash
make lint
```

**Format:**
```bash
make fmt
```

> **Note:** You do NOT need real eToro API credentials to develop. Most logic can be tested without a live account. For integration testing, use `--dry-run` and `--json` to validate output shapes without executing trades.

---

## Project Structure

```
etoro-cli/
├── cmd/             # Cobra commands (one file per command)
├── internal/
│   ├── api/         # eToro API client
│   ├── config/      # Config loading & validation
│   ├── output/      # Table + JSON formatters
│   └── safety/      # Kill switch, trade limits, allowlist
├── main.go
├── Makefile
├── config.example.yaml
└── README.md
```

Adding a new command? Create `cmd/<command>.go` following the pattern in existing commands.

---

## Making Changes

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, keeping commits focused and descriptive:
   ```bash
   git commit -m "feat: add watchlist remove command"
   ```

3. Keep your branch up to date:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

---

## Submitting a Pull Request

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a PR against `main` on [github.com/lijinlar/etoro-cli](https://github.com/lijinlar/etoro-cli/pulls)

3. Fill in the PR template — describe what changed and why

4. A maintainer will review and provide feedback

**PR checklist:**
- [ ] `make build` passes
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] New commands have `--json` flag support
- [ ] Safety guardrails are not bypassed without explicit user intent
- [ ] README updated if commands/flags changed

---

## Coding Guidelines

- **Follow standard Go conventions** — `gofmt`, idiomatic naming, error wrapping with `%w`
- **Every command must support `--json`** — the CLI is designed to be agent-friendly
- **Safety first** — never bypass kill switch, execution toggle, or trade limits without a deliberate user action
- **No credentials in code** — API keys come from config/env only
- **Keep output clean** — tables for humans, clean JSON for machines (no mixed output)
- **Errors should be actionable** — tell the user what to do, not just what went wrong

---

## Reporting Bugs

Use the [Bug Report template](https://github.com/lijinlar/etoro-cli/issues/new?template=bug_report.yml). Please include:
- etoro-cli version (`etoro --version`)
- Operating system
- The command you ran (redact any keys)
- Expected vs actual output

---

## Requesting Features

Use the [Feature Request template](https://github.com/lijinlar/etoro-cli/issues/new?template=feature_request.yml). Describe the use case — why it matters and how you'd expect it to work.

---

## Questions?

Open a [Discussion](https://github.com/lijinlar/etoro-cli/discussions) — happy to help.
