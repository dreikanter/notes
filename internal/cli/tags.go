package cli

import (
	"fmt"
	"sort"

	"github.com/dreikanter/notes/note"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags from frontmatter and body hashtags",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTagsList(cmd)
	},
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags from frontmatter and body hashtags",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTagsList(cmd)
	},
}

var tagsRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a tag across the store (frontmatter and body hashtags)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldTag := args[0]
		newTag := args[1]

		if oldTag == "" {
			return fmt.Errorf("old tag is empty")
		}
		if newTag == "" {
			return fmt.Errorf("new tag is empty")
		}
		if err := note.ValidateTag(newTag); err != nil {
			return fmt.Errorf("new tag: %w", err)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		store, err := notesStore()
		if err != nil {
			return err
		}

		res, renameErr := note.RenameTag(store, oldTag, newTag, note.RenameOpts{DryRun: dryRun})

		out := cmd.OutOrStdout()
		errOut := cmd.ErrOrStderr()
		for _, p := range res.ModifiedPaths {
			fmt.Fprintln(out, p)
		}

		n := len(res.ModifiedPaths)
		switch {
		case renameErr != nil:
			fmt.Fprintf(errOut, "partial: renamed in %s before error\n", pluralNotes(n))
			return renameErr
		case n == 0:
			fmt.Fprintf(errOut, "no notes contained tag %q\n", oldTag)
		case dryRun:
			fmt.Fprintf(errOut, "would rename %q → %q in %s\n", oldTag, newTag, pluralNotes(n))
		default:
			fmt.Fprintf(errOut, "renamed %q → %q in %s\n", oldTag, newTag, pluralNotes(n))
		}
		return nil
	},
}

func pluralNotes(n int) string {
	if n == 1 {
		return "1 note"
	}
	return fmt.Sprintf("%d notes", n)
}

func runTagsList(cmd *cobra.Command) error {
	store, err := notesStore()
	if err != nil {
		return err
	}
	entries, err := store.All()
	if err != nil {
		return err
	}
	set := make(map[string]struct{})
	for _, e := range entries {
		for _, t := range e.Meta.Tags {
			set[t] = struct{}{}
		}
	}
	tags := make([]string, 0, len(set))
	for t := range set {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	out := cmd.OutOrStdout()
	for _, t := range tags {
		fmt.Fprintln(out, t)
	}
	return nil
}

func init() {
	tagsRenameCmd.Flags().Bool("dry-run", false, "print modifications without writing")
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsRenameCmd)
	rootCmd.AddCommand(tagsCmd)
}
