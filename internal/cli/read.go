package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/dreikanter/notes/note"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <id> [id...]",
	Short: "Read one or more notes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, err := parseReadIDs(args)
		if err != nil {
			return err
		}

		store, err := notesStore()
		if err != nil {
			return err
		}

		entries, err := readEntries(store, ids)
		if err != nil {
			return err
		}

		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return writeReadJSON(cmd, entries)
		}

		noFrontmatter, _ := cmd.Flags().GetBool("no-frontmatter")
		return writeReadMarkdown(cmd, store, entries, noFrontmatter)
	},
}

type readJSONRecord struct {
	ID          int      `json:"id"`
	UID         string   `json:"uid"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Tags        []string `json:"tags"`
	Date        string   `json:"date"`
	Description string   `json:"description"`
	Public      bool     `json:"public"`
	Body        string   `json:"body"`
}

func parseReadIDs(args []string) ([]int, error) {
	ids := make([]int, len(args))
	for i, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return nil, fmt.Errorf("id must be an integer: %s", arg)
		}
		ids[i] = id
	}
	return ids, nil
}

func readEntries(store *note.OSStore, ids []int) ([]note.Entry, error) {
	entries := make([]note.Entry, len(ids))
	for i, id := range ids {
		entry, err := store.Get(id)
		if err != nil {
			if errors.Is(err, note.ErrNotFound) {
				return nil, fmt.Errorf("note %d not found", id)
			}
			return nil, err
		}
		entries[i] = entry
	}
	return entries, nil
}

func writeReadMarkdown(cmd *cobra.Command, store *note.OSStore, entries []note.Entry, noFrontmatter bool) error {
	out := cmd.OutOrStdout()
	for _, entry := range entries {
		data, err := os.ReadFile(store.AbsPath(entry))
		if err != nil {
			return err
		}
		if noFrontmatter {
			data = note.StripFrontmatter(data)
		}
		if _, err := out.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func writeReadJSON(cmd *cobra.Command, entries []note.Entry) error {
	records := make([]readJSONRecord, len(entries))
	for i, entry := range entries {
		records[i] = readJSONRecord{
			ID:          entry.ID,
			UID:         entryUID(entry),
			Title:       entry.Meta.Title,
			Slug:        entry.Meta.Slug,
			Tags:        jsonTags(entry.Meta.Tags),
			Date:        entry.Meta.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Description: entry.Meta.Description,
			Public:      entry.Meta.Public,
			Body:        entry.Body,
		}
	}
	return json.NewEncoder(cmd.OutOrStdout()).Encode(records)
}

func entryUID(entry note.Entry) string {
	if entry.UID != "" {
		return entry.UID
	}
	return note.UID(entry.Meta.CreatedAt.Format(note.DateFormat), entry.ID)
}

func jsonTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}

func registerReadFlags() {
	readCmd.Flags().Bool("no-frontmatter", false, "exclude YAML frontmatter from Markdown output")
	readCmd.Flags().Bool("json", false, "emit notes as a JSON array")
}

func init() {
	registerReadFlags()
	rootCmd.AddCommand(readCmd)
}
