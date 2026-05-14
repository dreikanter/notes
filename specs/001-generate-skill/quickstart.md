# Quickstart: Generate Agent Skill

**Branch**: `claude/implement-notes-skill-dOgl8` | **Date**: 2026-05-14

Local instructions for building, testing, and exercising the new
`notes skill` command.

## Build

```sh
make build         # produces ./notes bound to the current git tag
./notes --help     # confirm the new "skill" subcommand is listed
```

## Smoke-test stdout mode

```sh
./notes skill | head -40                       # human/plain output
./notes skill --help                            # documents flags
diff <(./notes skill) <(./notes skill)          # MUST be empty (determinism)
```

## Smoke-test install mode against a sandbox HOME

Avoid touching the real `~/.claude`. Use a tempdir as HOME:

```sh
SANDBOX=$(mktemp -d)
mkdir -p "$SANDBOX/.claude/skills"  # signal Claude Code is "installed"

# Auto-detect mode
HOME="$SANDBOX" ./notes skill --install
# Expect: create claude  $SANDBOX/.claude/skills/notes/SKILL.md

# Re-run: should be a no-op (byte-identical)
HOME="$SANDBOX" ./notes skill --install
# Expect: skip claude ...

# Mutate the file and re-run without --force: conflict
echo CHANGED >> "$SANDBOX/.claude/skills/notes/SKILL.md"
HOME="$SANDBOX" ./notes skill --install
# Expect: conflict, process exit non-zero

# Re-run with --force: overwrites
HOME="$SANDBOX" ./notes skill --install --force
# Expect: overwrite, exit 0

# Dry-run never writes
rm "$SANDBOX/.claude/skills/notes/SKILL.md"
HOME="$SANDBOX" ./notes skill --install --dry-run
# Expect: would create ...
test ! -e "$SANDBOX/.claude/skills/notes/SKILL.md"   # nothing written
```

## Smoke-test agent selection and error paths

```sh
HOME="$SANDBOX" ./notes skill --install --agent=claude    # explicit OK
HOME="$SANDBOX" ./notes skill --install --agent=bogus     # error: unknown agent
./notes skill --force                                       # usage error (no --install)
HOME=$(mktemp -d) ./notes skill --install                   # no agents detected → error
```

## Run the test suite

```sh
make test
make lint
```

Both MUST pass before pushing or opening the PR.

## Inspect what the skill says

```sh
./notes skill | less
```

The body should contain:

- An Overview paragraph naming the binary.
- A Global Flags section with `--path`.
- A Commands table with one row per available subcommand, alphabetised.
- A Store layout section pointing at `NOTES_PATH` and the
  `YYYY/YYYYMMDD_NNNN.md` file layout.
- A shell-composition section showing the `ls`/`resolve` → ID pattern.
