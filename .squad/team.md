# Squad Team

> hyoka — Go evaluation tool for AI-generated Azure SDK code, powered by Copilot SDK and multi-model review panels.

## Coordinator

| Name | Role | Notes |
|------|------|-------|
| Squad | Coordinator | Routes work, enforces handoffs and reviewer gates. |

## Members

| Name | Role | Charter | Status |
|------|------|---------|--------|
| Morpheus 🕶️ | Lead | `.squad/agents/morpheus/charter.md` | 🏗️ Lead |
| Neo 💊 | Core Eval Framework | `.squad/agents/neo/charter.md` | 🔧 Core Eval |
| Tank 📡 | CLI Dev | `.squad/agents/tank/charter.md` | ⌨️ CLI Dev |
| Trinity 🖤 | Frontend Dev | `.squad/agents/trinity/charter.md` | ⚛️ Frontend |
| Switch 🤍 | Testing | `.squad/agents/switch/charter.md` | 🧪 Testing |
| Oracle 🔮 | Docs | `.squad/agents/oracle/charter.md` | 📝 Docs |
| Scribe | Scribe | `.squad/agents/scribe/charter.md` | 📋 Scribe |
| Ralph | Work Monitor | `.squad/agents/ralph/charter.md` | 🔄 Monitor |

## Project Context

- **Project:** hyoka
- **Stack:** Go 1.26.1+, GitHub Copilot CLI/SDK, MCP servers (Azure MCP via npx)
- **Description:** A curated library of prompts for evaluating AI-generated Azure SDK code, with a Go tool that runs prompts through the Copilot SDK, reviews code via a multi-model panel, and produces criteria-based pass/fail reports.
- **User:** Ronnie Geraghty
- **Created:** 2026-04-03
- **Universe:** The Matrix
