# Data Model: Generate Agent Skill

**Branch**: `claude/implement-notes-skill-dOgl8` | **Date**: 2026-05-14

In-memory types used by the `notes skill` command. None of these is
persisted; they exist only for the lifetime of the invocation.

## `skillContent`

The skill document itself is a single embedded string:

```go
//go:embed skill.md
var skillContent string
```

The bytes are read from `internal/cli/skill.md` at compile time. There
is no in-memory parsing or templating — the file is shipped verbatim,
including its YAML frontmatter (`name`, `description`) and body.

## `agentTarget`

A supported AI assistant.

```go
type agentTarget struct {
    Name    string                  // public flag value, e.g. "claude"
    PathFor func() (string, error)  // absolute install path
    Detect  func() (bool, error)    // best-effort presence check
}
```

The registry is a package-level `var agents = []agentTarget{...}` slice;
appending a new entry adds a new agent.

## `installAction`

A single planned filesystem operation against one agent target.

```go
type installAction struct {
    Agent  string // agent name
    Path   string // absolute destination path
    Action string // "create" | "overwrite" | "skip" | "conflict"
    Error  error  // populated when planning or applying failed
}
```

Action semantics:

- `create` — destination does not exist; will be written.
- `skip` — destination exists with identical bytes; will not be written.
- `conflict` — destination exists with different bytes and `--force` is
  not set; will not be written; contributes a non-zero exit code.
- `overwrite` — destination exists with different bytes and `--force` is
  set; will be written.

`Error != nil` short-circuits Action for that target and contributes a
non-zero exit code regardless of the other actions in the same plan.
