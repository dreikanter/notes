# Contract: `notes skill` Command Surface

**Branch**: `claude/implement-notes-skill-dOgl8` | **Date**: 2026-05-14

## Synopsis

```text
notes skill                              # write skill to stdout
notes skill --install                    # auto-detect agents, install into each
notes skill --install --agent=<name>     # install into a single named agent
notes skill --install --force            # overwrite an existing destination
notes skill --install -n / --dry-run     # print planned actions, write nothing
```

## Flags

| Flag                | Type   | Modes        | Effect                                                                          |
|---------------------|--------|--------------|---------------------------------------------------------------------------------|
| `--install`         | bool   | toggles mode | Switch from stdout mode to install mode.                                        |
| `--agent <name>`    | string | install only | Install into a single named agent. Default: auto-detect.                        |
| `--force`           | bool   | install only | Overwrite an existing destination file with diverging content.                  |
| `-n`, `--dry-run`   | bool   | install only | Print the planned actions but do not write any files.                           |

The persistent `--path` flag inherited from `rootCmd` is accepted (since
Cobra never rejects unknown persistent flags), but the `skill` command
does not read it — generating the skill does not depend on a notes store.

## Flag-interaction errors (exit non-zero with `SilenceUsage = true`)

- `--agent` outside of install mode.
- `--force` outside of install mode.
- `--dry-run` outside of install mode.
- `--agent=<unknown>`: error message names the unknown value and lists
  supported agents.
- `--install` with `--agent` empty and no agent detected: error message
  names supported agents and recommends `--agent`.

## Action set (install mode)

See `data-model.md` § `installAction` for the four possible actions
(`create`, `skip`, `conflict`, `overwrite`) and the decision table.

## Exit codes

- `0` — stdout mode succeeded, or every action is `create`/`skip`/`overwrite`.
- non-zero — any flag-interaction error; any `conflict` action; any I/O
  error while reading or writing a destination.

## Human output (install mode)

One line per action:

```text
<action> <agent> <path>
```

Examples:

```text
create   claude  /home/alice/.claude/skills/notes/SKILL.md
skip     claude  /home/alice/.claude/skills/notes/SKILL.md
conflict claude  /home/alice/.claude/skills/notes/SKILL.md
overwrite claude /home/alice/.claude/skills/notes/SKILL.md
```

For dry-run, the same lines are printed prefixed with `would `:

```text
would create  claude  /home/alice/.claude/skills/notes/SKILL.md
```

## Filesystem behaviour

- The command writes exactly one file per selected target.
- It creates intermediate directories under an existing agent skills
  directory (e.g. it creates `~/.claude/skills/notes/` if
  `~/.claude/skills/` exists), but it does NOT create the agent's
  containing skills directory itself (`~/.claude/skills/`) — that
  directory's absence is the detection signal that the user does not run
  this agent.
