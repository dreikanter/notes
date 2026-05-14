# Feature Specification: Generate Agent Skill

**Feature Branch**: `claude/implement-notes-skill-dOgl8`

**Created**: 2026-05-14

**Status**: Draft

**Input**: User request — "Implement `notes skill` feature" following the
pattern of `dotfiles-cli/specs/001-generate-skill`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Print the skill to stdout (Priority: P1) — MVP

As a user (human or agent) of an AI coding assistant, I want a single
command that emits a ready-to-use *skill* describing how to drive `notes`,
so that my assistant can learn the CLI's surface, conventions, and store
layout from one authoritative document instead of guessing from `--help`
output piecemeal.

**Why this priority**: Standalone value. With only stdout output, a user
can pipe the result anywhere — into a file, a clipboard tool, an agent
configuration, or a chat message. Every later story builds on this
content.

**Independent Test**: Run `notes skill` on a fresh install. Confirm the
output is a single self-contained markdown document, beginning with a YAML
frontmatter block (`name`, `description`), and that the body covers every
top-level subcommand of the CLI and the global flags. Nothing else is
written to disk.

**Acceptance Scenarios**:

1. **Given** a working `notes` binary, **When** the user runs
   `notes skill`, **Then** a single markdown document is printed to stdout
   and the process exits with status `0`.
2. **Given** the same binary, **When** the user runs `notes skill --help`,
   **Then** the help text documents the command and its flags.
3. **Given** the binary is rebuilt, **When** the user runs `notes skill`
   on two different machines, **Then** both invocations emit byte-identical
   output, because the skill content is embedded into the binary at build
   time from a versioned markdown source.

---

### User Story 2 — Install the skill into a known agent location (Priority: P2)

As a user who already runs an AI coding assistant locally, I want a single
command that drops the skill into the assistant's well-known skills
directory, so that I don't have to know where the directory lives, what
filename to use, or how to format the frontmatter.

**Independent Test**: On a machine where the target agent's skills
directory exists, run `notes skill --install --agent=<name>`. Confirm a
single file is created at the agent's documented skill location, its
content matches `notes skill` exactly, and the command reports the
destination path.

**Acceptance Scenarios**:

1. **Given** the target agent's skills directory exists and no skill file
   is present, **When** the user runs `notes skill --install --agent=<name>`,
   **Then** the skill file is created at the agent-specific path and the
   command prints the destination path.
2. **Given** a skill file already exists at the target path with identical
   content, **When** the user re-runs the same command, **Then** no file
   is modified and the action is reported as `skip`.
3. **Given** a skill file already exists with different content, **When**
   the user runs without `--force`, **Then** no file is modified, the
   command exits non-zero, and the message identifies the conflict.
4. **Given** the same precondition, **When** the user adds `--force`,
   **Then** the existing file is overwritten and the action is reported.
5. **Given** any of the above, **When** the user adds `-n, --dry-run`,
   **Then** the command prints the action it *would* take and writes
   nothing to disk.
6. **Given** the user passes `--agent` with an unknown value, **Then** the
   command exits non-zero with an error naming the unknown agent and
   listing the supported values.

---

### User Story 3 — Auto-detect installed agents (Priority: P3)

As a user who has more than one AI assistant installed, I want
`notes skill --install` (without `--agent`) to discover which assistants
are present on this machine and install the skill into each one.

**Independent Test**: On a machine where a supported agent is installed,
run `notes skill --install`. Confirm the skill is created in the agent's
skill directory. On a machine with no supported agents detected, confirm
the command exits non-zero with an actionable error.

**Acceptance Scenarios**:

1. **Given** at least one supported agent is detected on the machine,
   **When** the user runs `notes skill --install`, **Then** the skill is
   installed into every detected agent's location and each destination
   path is reported.
2. **Given** no supported agent is detected, **When** the same command is
   run, **Then** the command exits non-zero with a message listing
   supported agents and how to pass `--agent` explicitly.
3. **Given** auto-detect mode, **When** the skill already exists in some
   locations but not others, **Then** without `--force` the command
   installs into empty locations and reports the conflicts in the
   already-populated locations.

---

### Edge Cases

- The skill content references the binary's version; a stale binary
  prints a stale skill. Users opt in to refresh by re-running.
- The target agent directory's parent (`~/.claude/skills/`) does not
  exist: the command exits non-zero with a message explaining the missing
  directory; it does not create unfamiliar parent directories.
- The user's home directory is not writable: the install mode surfaces
  the OS error and exits non-zero.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide a `notes skill` command exposed as a
  top-level subcommand alongside the existing commands.
- **FR-002**: Invoked with no extra arguments, the command MUST write the
  skill content to stdout as a single markdown document and MUST NOT
  modify the filesystem.
- **FR-003**: The skill content MUST begin with a YAML frontmatter block
  containing at minimum a `name` field and a `description` field that
  follows the convention "Use when …" so AI agents can match it to a
  task.
- **FR-004**: The skill body MUST cover, at minimum: every top-level
  subcommand and its purpose; the global `--path` flag; the role of
  `NOTES_PATH`; the note store layout (date-stamped folders, markdown
  files with YAML frontmatter); and a pointer to `notes <command> --help`
  as the authoritative source for each command's full flag list.
- **FR-005**: The skill content MUST be authored as a markdown file
  checked into the repository and embedded into the binary at build time
  (via Go's `//go:embed` directive), so the same bytes ship with every
  copy of a given build and the content lives under normal source
  control.
- **FR-006**: The command MUST accept an `--install` flag that, when
  present, switches the command from stdout mode to filesystem mode.
- **FR-007**: In install mode, the command MUST accept an optional
  `--agent=<name>` flag selecting the destination. Supported `<name>`
  values MUST be enumerated in `--help`.
- **FR-008**: In install mode without `--agent`, the command MUST attempt
  to auto-detect installed agents from a known list and install into
  every detected agent's location. If none is detected, the command MUST
  exit non-zero with an actionable error.
- **FR-009**: In install mode, the command MUST refuse to overwrite an
  existing destination file unless `--force` is passed.
- **FR-010**: In install mode, the command MUST support `-n, --dry-run`,
  which prints the actions it would take and writes nothing.
- **FR-011**: In install mode, the command MUST report each destination
  path on success.

### Key Entities

- **Skill**: a self-contained markdown document with YAML frontmatter
  describing how an AI agent should drive the `notes` CLI. There is
  exactly one skill per `notes` binary.
- **Agent target**: a named AI assistant (initially Claude Code) that has
  a documented filesystem location for user-installed skills.

## Success Criteria *(mandatory)*

- **SC-001**: A user who has never read the README can install the skill
  into their AI assistant in a single command, and the assistant can then
  perform note creation, listing, and reading workflows correctly on the
  first attempt.
- **SC-002**: The skill output contains an entry for 100% of the commands
  listed by `notes --help`; no command is silently omitted.
- **SC-003**: Re-running `notes skill --install` without `--force`
  against an unchanged target is a no-op. Re-running with `--force`
  produces a destination file byte-for-byte equal to the stdout output of
  `notes skill` on the same binary.
- **SC-004**: The skill source is a single markdown file in the
  repository, reviewable via normal PR review. Updating the skill is a
  one-file change with no code edits required.

## Assumptions

- Initial agent support is Claude Code only. The `--agent=claude` target
  ships in the first version, and auto-detect initially looks for that
  one agent. Additional agents are added in follow-up work.
- The Claude Code skill location is `~/.claude/skills/notes/SKILL.md`.
- The skill content is generated at runtime from the same command
  registry that powers `--help`, so version drift between the binary and
  the skill content is impossible by construction.

## Dependencies

- The CLI's existing Cobra command registry must expose each command's
  name, short description, and flag list (it already does — that's what
  `--help` reads).
- The destination directory for the Claude Code agent
  (`~/.claude/skills/notes/`) must be created on demand only when the
  user passes `--install`.
