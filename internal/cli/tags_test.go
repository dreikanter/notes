package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetTagsFlags clears subcommand flag state that cobra retains between
// rootCmd.Execute() invocations, so independent tests don't leak settings
// (e.g. --dry-run) into each other.
func resetTagsFlags() {
	_ = tagsRenameCmd.Flags().Set("dry-run", "false")
}

func runTags(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()
	resetTagsFlags()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"tags", "--path", root}, args...))

	err := rootCmd.Execute()
	return strings.TrimSpace(buf.String()), err
}

func runTagsSplit(t *testing.T, root string, args ...string) (string, string, error) {
	t.Helper()
	resetTagsFlags()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs(append([]string{"tags", "--path", root}, args...))

	err := rootCmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

func writeIDJSON(t *testing.T, root string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(root, "id.json"), []byte(`{"last_id":0}`), 0o644))
}

func writeTagsTestNote(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

func TestTagsEmptyStore(t *testing.T) {
	root := t.TempDir()
	out, err := runTags(t, root)
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestTagsMergedSourcesSorted(t *testing.T) {
	root := t.TempDir()
	writeTagsTestNote(t, root, "2026/01/20260101_1001.md",
		"---\ntags: [work, planning]\n---\n\nHere is #coffee and #work again.\n")
	writeTagsTestNote(t, root, "2026/01/20260102_1002.md",
		"no fm, just #tea and #work.\n")

	out, err := runTags(t, root)
	require.NoError(t, err)
	assert.Equal(t, []string{"coffee", "planning", "tea", "work"}, strings.Split(out, "\n"))
}

func TestTagsLowercased(t *testing.T) {
	root := t.TempDir()
	writeTagsTestNote(t, root, "2026/01/20260101_1001.md",
		"---\ntags: [Work, PLANNING]\n---\n\nbody with #Coffee and #WORK.\n")

	out, err := runTags(t, root)
	require.NoError(t, err)
	assert.Equal(t, []string{"coffee", "planning", "work"}, strings.Split(out, "\n"))
}

func TestTagsIgnoresCodeBlocks(t *testing.T) {
	root := t.TempDir()
	writeTagsTestNote(t, root, "2026/01/20260101_1001.md",
		"kept #real\n```\n#should-not-appear\n```\nalso #done\n")

	out, err := runTags(t, root)
	require.NoError(t, err)
	assert.NotContains(t, out, "should-not-appear")
	assert.Contains(t, out, "real")
	assert.Contains(t, out, "done")
}

func TestTagsListSubcommandSameAsTags(t *testing.T) {
	root := t.TempDir()
	writeTagsTestNote(t, root, "2026/01/20260101_1001.md",
		"---\ntags: [work, planning]\n---\n\nhere is #coffee\n")

	bare, err := runTags(t, root)
	require.NoError(t, err)
	listed, err := runTags(t, root, "list")
	require.NoError(t, err)
	assert.Equal(t, bare, listed)
	assert.Equal(t, []string{"coffee", "planning", "work"}, strings.Split(bare, "\n"))
}

func TestTagsRenameHappyPath(t *testing.T) {
	root := t.TempDir()
	writeIDJSON(t, root)
	path := filepath.Join(root, "2026", "01", "20260101_1.md")
	writeTagsTestNote(t, root, "2026/01/20260101_1.md",
		"---\ntags: [work]\n---\n\nbody #work here\n")

	stdout, stderr, err := runTagsSplit(t, root, "rename", "work", "personal")
	require.NoError(t, err)
	assert.Equal(t, path+"\n", stdout)
	assert.Equal(t, "renamed \"work\" → \"personal\" in 1 note\n", stderr)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "tags:\n    - personal")
	assert.Contains(t, string(got), "body #personal here")
}

func TestTagsRenameHappyPathPlural(t *testing.T) {
	root := t.TempDir()
	writeIDJSON(t, root)
	writeTagsTestNote(t, root, "2026/01/20260101_1.md",
		"---\ntags: [work]\n---\n\na\n")
	writeTagsTestNote(t, root, "2026/01/20260102_2.md",
		"---\ntags: [work]\n---\n\nb\n")

	_, stderr, err := runTagsSplit(t, root, "rename", "work", "personal")
	require.NoError(t, err)
	assert.Equal(t, "renamed \"work\" → \"personal\" in 2 notes\n", stderr)
}

func TestTagsRenameDryRun(t *testing.T) {
	root := t.TempDir()
	writeIDJSON(t, root)
	path := filepath.Join(root, "2026", "01", "20260101_1.md")
	content := "---\ntags: [work]\n---\n\nbody #work here\n"
	writeTagsTestNote(t, root, "2026/01/20260101_1.md", content)

	stdout, stderr, err := runTagsSplit(t, root, "rename", "--dry-run", "work", "personal")
	require.NoError(t, err)
	assert.Equal(t, path+"\n", stdout)
	assert.Equal(t, "would rename \"work\" → \"personal\" in 1 note\n", stderr)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, string(got), "file should be untouched in dry-run")
}

func TestTagsRenameEmptyArgsFail(t *testing.T) {
	root := t.TempDir()
	writeIDJSON(t, root)

	_, _, err := runTagsSplit(t, root, "rename", "", "new")
	require.Error(t, err)

	_, _, err = runTagsSplit(t, root, "rename", "old", "")
	require.Error(t, err)
}

func TestTagsRenameInvalidNewTag(t *testing.T) {
	root := t.TempDir()
	writeIDJSON(t, root)

	_, _, err := runTagsSplit(t, root, "rename", "old", "has space")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tag character")
}

func TestTagsRenameNoMatches(t *testing.T) {
	root := t.TempDir()
	writeIDJSON(t, root)
	writeTagsTestNote(t, root, "2026/01/20260101_1.md",
		"---\ntags: [other]\n---\n\nbody\n")

	stdout, stderr, err := runTagsSplit(t, root, "rename", "nope", "x")
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(stdout))
	assert.Equal(t, "no notes contained tag \"nope\"\n", stderr)
}

func TestTagsRenameUnicode(t *testing.T) {
	root := t.TempDir()
	writeIDJSON(t, root)
	path := filepath.Join(root, "2026", "01", "20260101_1.md")
	writeTagsTestNote(t, root, "2026/01/20260101_1.md",
		"---\ntags: [café]\n---\n\ndrink #café please\n")

	stdout, stderr, err := runTagsSplit(t, root, "rename", "café", "latte")
	require.NoError(t, err)
	assert.Equal(t, path+"\n", stdout)
	assert.Equal(t, "renamed \"café\" → \"latte\" in 1 note\n", stderr)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "tags:\n    - latte")
	assert.Contains(t, string(got), "drink #latte please")
}

func TestTagsRenameCaseOnly(t *testing.T) {
	root := t.TempDir()
	writeIDJSON(t, root)
	path := filepath.Join(root, "2026", "01", "20260101_1.md")
	writeTagsTestNote(t, root, "2026/01/20260101_1.md",
		"---\ntags: [work]\n---\n\nbody #work here\n")

	stdout, stderr, err := runTagsSplit(t, root, "rename", "work", "WORK")
	require.NoError(t, err)
	assert.Equal(t, path+"\n", stdout)
	assert.Equal(t, "renamed \"work\" → \"WORK\" in 1 note\n", stderr)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "tags:\n    - WORK")
	assert.Contains(t, string(got), "body #WORK here")
}
