---
name: notes
description: Use when interacting with a notes store via the notes CLI — creating, reading, listing, updating, annotating, or deleting notes managed under a date-stamped markdown archive.
---

# notes

## Overview

`notes` is a command-line interface for a plain-text note archive. Each note is a markdown file with a YAML frontmatter block, stored under a date-stamped directory tree (`YYYY/YYYYMMDD_NNNN.md`). There is no database, no proprietary format, and no sync service — the files are owned by the user. The CLI provides scriptable access to create, read, list, update, annotate, and delete notes.

## Global flags

- `--path` — path to notes store (default: $NOTES_PATH)

## Commands

- `notes annotate` — Fill empty frontmatter (title, description, tags) using Claude Code
- `notes append` — Append text from stdin to a note
- `notes config` — Print the effective runtime configuration
- `notes ls` — List note IDs, newest first
- `notes new` — Create a new note
- `notes new-todo` — Create today's todo, carrying over incomplete tasks from the previous todo
- `notes read` — Read a note
- `notes resolve` — Print the absolute path of a note
- `notes rm` — Delete a note
- `notes skill` — Print or install the notes agent skill
- `notes tags` — List all tags from frontmatter and body hashtags
- `notes update` — Update note frontmatter (file is renamed automatically when slug, type, or date changes)

Run `notes <command> --help` for the full flag list and behaviour of each command.

## Store layout

The notes store path is resolved in this order: `--path` flag, then `NOTES_PATH` environment variable. If neither is set, the CLI exits with an error. There is no implicit default.

Each note lives at `<store>/<YYYY>/<YYYYMMDD>_<NNNN>.md`, where `NNNN` is a per-day numeric ID. Note IDs are unique within a day; CLI commands that take an ID accept the numeric portion (e.g. `notes read 8823`). Frontmatter is YAML with optional `title`, `slug`, `type`, `tags`, `description`, and `public` fields. Inline `#hashtag` tokens in the body are also indexed.

## Composing with the shell

Most commands take a numeric note ID. To act on "the most recent note of type X" or "the most recent note with slug Y", use `notes ls` or `notes resolve` to turn a filter into an ID, then shell-substitute:

```sh
# Append to the most recent note with a given slug
echo "text" | notes append "$(notes ls --slug claude-sessions --limit 1)"

# Read the most recent todo
notes read "$(notes ls --type todo --limit 1)"

# Open the most recent meeting note in $EDITOR
$EDITOR "$(notes resolve --slug meeting)"
```
