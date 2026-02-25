# Agent Personas

tfwatch uses structured AI agent personas to enforce consistent, scoped assistance across any AI coding tool. Each agent has a fixed role, output format, file ownership scope, and anti-drift rules.

## Quick Reference

| Agent | Role | When to Use |
|-------|------|-------------|
| **PM** | Project Owner | Evaluate features, scope decisions, roadmap |
| **Dev** | Software Engineer | Design and build new features (Go + Terraform expertise) |
| **Fix** | Maintainer | Bug fixes, code review, refactoring |
| **DevOps** | DevOps Engineer | CI/CD, releases, deploy stack, Makefile |
| **Docs** | Technical Writer | Documentation, website copy |
| **UI** | UI Designer | Website layout, styling, assets |

## File Structure

```
prompts/
├── CONTEXT.md              # Shared project context (all agents read this)
├── GUARDRAILS.md           # AI safety rules (all agents follow this)
└── agents/
    ├── pm.md               # Project Owner persona
    ├── dev.md              # Software Engineer persona
    ├── fix.md              # Maintainer persona
    ├── devops.md           # DevOps Engineer persona
    ├── docs.md             # Technical Writer persona
    └── ui.md               # UI Designer persona
```

## Usage by Tool

### Claude Code

Slash commands are pre-configured in `.claude/commands/`. Just type:

```
/pm should we add GCS backend support?
/dev add GCS backend parsing
/fix the parser panics on empty lock files
/devops add a workflow step for integration tests
/docs update README with GCS backend section
/ui improve the getting-started page hero section
```

`CLAUDE.md` auto-loads project context on every session.

### Cursor

Add a `.cursorrules` file or use Cursor's rules settings to include:

```
Read and follow prompts/CONTEXT.md for project context.
Read and follow prompts/GUARDRAILS.md for AI safety rules.
```

To use a specific agent, add its persona to a Cursor rule or paste it in chat:

```
@prompts/agents/dev.md design a GCS backend parser
```

Or create per-agent `.cursor/rules/` files that reference the prompts.

### GitHub Copilot

Create `.github/copilot-instructions.md`:

```markdown
Read prompts/CONTEXT.md for project context and prompts/GUARDRAILS.md for AI safety rules.
When working on Go code, follow the conventions in prompts/agents/dev.md.
When fixing bugs, follow prompts/agents/fix.md.
```

### Windsurf

Create a `.windsurfrules` file:

```
Read prompts/CONTEXT.md for project context.
Read prompts/GUARDRAILS.md for AI safety rules.
For feature work, adopt the persona in prompts/agents/dev.md.
```

### Aider / Other Tools

Point your tool at the relevant prompt file:

```bash
aider --read prompts/CONTEXT.md --read prompts/agents/dev.md
```

### Manual (Any Tool)

Copy-paste the contents of `prompts/CONTEXT.md` + `prompts/GUARDRAILS.md` + the relevant `prompts/agents/<agent>.md` into your AI tool's system prompt or chat context.

## How It Works

1. **CONTEXT.md** provides shared project knowledge — architecture, conventions, file ownership boundaries
2. **GUARDRAILS.md** sets hard safety rules — what AI must never do, change, or build
3. **Agent files** define per-role behavior — identity, scope, output templates, anti-drift rules

Every agent response follows a fixed template with mandatory sections. This prevents output drift across sessions and ensures consistent, reviewable AI assistance.

## Adding a New Agent

1. Create `prompts/agents/<name>.md` with: Identity, Scope, Output Formats, Anti-Drift Rules
2. Add a row to the ownership table in `prompts/CONTEXT.md`
3. For Claude Code: create `.claude/commands/<name>.md` that references the prompt file
4. For other tools: update their config to include the new agent file
