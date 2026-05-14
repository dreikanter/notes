# Implementation Plan: Generate Agent Skill

**Branch**: `claude/implement-notes-skill-dOgl8` | **Date**: 2026-05-14 | **Spec**: [spec.md](./spec.md)

## Summary

Add a `notes skill` Cobra subcommand that, by default, prints a single
self-contained markdown document (with YAML frontmatter) describing how
an AI agent should drive the CLI. The document is authored as
`internal/cli/skill.md` in the repository and embedded into the binary
at build time via Go's `//go:embed` directive, so the same bytes ship
with every copy of a given build and the content is reviewed through the
normal PR review process. The same command, invoked with `--install`,
writes the embedded skill to one or more agent-specific filesystem
locations, honouring `--dry-run` and `--force`. The first release
supports Claude Code (`~/.claude/skills/notes/SKILL.md`); additional
agent targets are add-only entries in a small in-process registry.

## Technical Context

**Language/Version**: Go 1.25 (per `go.mod`).

**Primary Dependencies**:
- `github.com/spf13/cobra` — already in use; required for command
  registration.
- `embed` (Go standard library) — used via `//go:embed skill.md` to
  bake the skill into the binary at build time.
- `github.com/stretchr/testify` — already in use; sufficient for new
  tests.
- No new direct dependencies.

**Storage**: None at the data-store level. The command reads only the
embedded markdown bytes (compiled into the binary). In install mode it
writes exactly one file per selected agent target under the user's home
directory; no note-store files are touched.

**Testing**: Tests live in `internal/cli/skill_test.go`. Filesystem tests
use `t.TempDir()` and override the package-level home-dir resolver to
avoid touching the real `~/.claude`.

**Target Platform**: macOS and Linux, anywhere Go 1.25 builds.

**Project Type**: Single-project Go CLI tool.

**Constraints**:
- **Deterministic output**: two invocations on the same binary in the
  same environment MUST produce byte-identical skill content. Implies:
  stable iteration order over commands and flags, no timestamps in
  output.
- **No network access.** The skill is generated locally.
- **No writes outside the explicit install target.** Stdout mode is a
  pure function; install mode writes exactly one file per target.

**Scope**: One new top-level subcommand. Estimated footprint: ~1 new
markdown source `internal/cli/skill.md`, ~1 new Go file
`internal/cli/skill.go`, ~1 new test file `internal/cli/skill_test.go`,
plus a one-line `CHANGELOG.md` entry and a short README section.

## Project Structure

### Documentation (this feature)

```text
specs/001-generate-skill/
├── plan.md              # This file
├── spec.md              # User-facing specification
├── data-model.md        # In-memory types
├── quickstart.md        # Contributor smoke-tests
└── contracts/
    └── cli.md           # Command-surface contract: flags, exit codes, action set
```

### Source Code (repository root)

```text
cmd/
└── notes/
    └── main.go                       # unchanged — entrypoint delegates to internal/cli

internal/
└── cli/
    ├── root.go                       # unchanged
    ├── skill.md                      # NEW — authored skill content, embedded into the binary
    ├── skill.go                      # NEW — command def, install(), agent registry
    ├── skill_test.go                 # NEW — tests for stdout/install/dry-run/force
    └── …                             # other existing files unchanged
```

**Structure Decision**: Single-project layout, matches the existing
`internal/cli/<command>.go` per-command convention. The skill renderer,
agent registry, and Cobra handler all live in `internal/cli/skill.go`
because they share the same data (Cobra's command tree).

## Design Decisions

### How is the skill content authored and shipped?

The skill is a hand-authored markdown file at `internal/cli/skill.md`,
including its YAML frontmatter block. The Go file declares:

```go
//go:embed skill.md
var skillContent string
```

At build time the markdown bytes are baked into the binary, so the
runtime cost is zero and the bytes emitted are byte-identical across
machines. Editing the skill is a one-file PR; no code changes are
required to update wording, structure, or examples.

### What is the structure of the skill body?

Authored as five sections inside `skill.md`:

1. **Overview** — one paragraph: what `notes` does, its store model.
2. **Global flags** — list of persistent flags (`--path`).
3. **Commands** — bullet list of subcommands with their `Short` blurbs,
   pointing the reader at `notes <name> --help` for the full flag set.
4. **Store layout** — short paragraph describing how `NOTES_PATH`
   resolves and the on-disk format (`YYYY/YYYYMMDD_NNNN.md` with YAML
   frontmatter).
5. **Composing with the shell** — short paragraph showing the
   `ls`/`resolve` → ID pattern that lets `read`, `append`, `update`,
   `rm`, `annotate` chain together.

### Keeping the skill in sync with the CLI

Because the skill is authored, not generated, an out-of-date skill is a
review concern. The acceptance scenario for keeping the document in sync
with new commands is "the author of a new command updates `skill.md` in
the same PR." Tests assert that the embedded content is non-empty,
starts with the expected frontmatter, and lists a known core set of
commands.

### How are agent install locations represented?

A small `[]agentTarget` slice declared in `internal/cli/skill.go`. Each
entry has `Name string`, `PathFor func() (string, error)`, and
`Detect func() (bool, error)`. Adding a new agent is one append.

### Detection signal

For Claude Code, the agent is considered "detected" when the directory
`~/.claude/skills/` exists on the machine. Cheap, side-effect-free, and
uses only `os.Stat`. We do not probe `$PATH` for a `claude` binary
(the name collides with too much else).

### How does install mode decide between Create / Overwrite / Skip / Conflict?

| Existing file? | Bytes equal to generated skill? | `--force`? | Action      | Writes? | Exit |
|----------------|---------------------------------|------------|-------------|---------|------|
| No             | —                               | —          | `create`    | yes     | 0    |
| Yes            | Yes                             | —          | `skip`      | no      | 0    |
| Yes            | No                              | No         | `conflict`  | no      | ≠ 0  |
| Yes            | No                              | Yes        | `overwrite` | yes     | 0    |

`--dry-run` reports the action without writing.

### How are filesystem paths under `~` resolved in tests?

Provide a package-level `homeDir func() (string, error)` defaulting to
`os.UserHomeDir`. Tests assign a stub that returns `t.TempDir()` and
reset it via `t.Cleanup`.
