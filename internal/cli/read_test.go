package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runRead(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runReadInRoot(t, testdataPath(t), args...)
}

func runReadInRoot(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()

	readCmd.ResetFlags()
	registerReadFlags()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"read", "--path", root}, args...))

	err := rootCmd.Execute()
	return strings.TrimSpace(buf.String()), err
}

func TestReadByID(t *testing.T) {
	out, err := runRead(t, "8823")
	require.NoError(t, err)
	assert.Contains(t, out, "Plain note")
}

func TestReadMultipleIDs(t *testing.T) {
	out, err := runRead(t, "8823", "8818")
	require.NoError(t, err)
	assert.Contains(t, out, "Plain note")
	assert.Contains(t, out, "Standup notes")
	assert.Less(t, strings.Index(out, "Plain note"), strings.Index(out, "Standup notes"))
}

func TestReadMissingID(t *testing.T) {
	_, err := runRead(t, "999999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestReadMissingIDAmongMany(t *testing.T) {
	_, err := runRead(t, "8823", "999999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "note 999999 not found")
}

func TestReadNonIntegerArg(t *testing.T) {
	_, err := runRead(t, "not-an-id")
	require.Error(t, err)
}

func TestReadNonIntegerArgAmongMany(t *testing.T) {
	_, err := runRead(t, "8823", "not-an-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id must be an integer: not-an-id")
}

func TestReadNoArgsErrors(t *testing.T) {
	_, err := runRead(t)
	require.Error(t, err)
}

func TestReadNoFrontmatter(t *testing.T) {
	out, err := runRead(t, "8818", "--no-frontmatter")
	require.NoError(t, err)
	// Frontmatter should be stripped; "tags:" should not appear
	assert.NotContains(t, out, "tags:")
	assert.Contains(t, out, "Standup notes")
}

func TestReadNoFrontmatterMultipleIDs(t *testing.T) {
	out, err := runRead(t, "8818", "8814", "--no-frontmatter")
	require.NoError(t, err)
	assert.NotContains(t, out, "tags:")
	assert.Contains(t, out, "Standup notes")
	assert.Contains(t, out, "Todo")
}

func TestReadJSON(t *testing.T) {
	out, err := runRead(t, "8818", "--json")
	require.NoError(t, err)

	var records []readJSONRecord
	require.NoError(t, json.Unmarshal([]byte(out), &records))
	require.Len(t, records, 1)

	record := records[0]
	assert.Equal(t, 8818, record.ID)
	assert.Equal(t, "20260104_8818", record.UID)
	assert.Equal(t, "meeting", record.Slug)
	assert.Equal(t, []string{"meeting", "work"}, record.Tags)
	assert.Equal(t, "2026-01-04T00:00:00Z", record.Date)
	assert.False(t, record.Public)
	assert.Contains(t, record.Body, "Standup notes")
	assert.NotContains(t, record.Body, "tags:")
}

func TestReadJSONMultipleIDs(t *testing.T) {
	out, err := runRead(t, "8823", "8814", "--json")
	require.NoError(t, err)

	var records []readJSONRecord
	require.NoError(t, json.Unmarshal([]byte(out), &records))
	require.Len(t, records, 2)
	assert.Equal(t, []int{8823, 8814}, []int{records[0].ID, records[1].ID})
	assert.Equal(t, "20260106_8823", records[0].UID)
	assert.Equal(t, "20260102_8814", records[1].UID)
	assert.Contains(t, records[0].Body, "Plain note")
	assert.Contains(t, records[1].Body, "Todo")
}

func TestReadUsesFilenamePathWhenFrontmatterDateDiffers(t *testing.T) {
	root := copyTestdata(t)
	writeFixtureDate(t, root, "2026-02-01")

	out, err := runReadInRoot(t, root, "8818")
	require.NoError(t, err)
	assert.Contains(t, out, "Standup notes")
}

func TestReadJSONUIDUsesFilenameDateWhenFrontmatterDateDiffers(t *testing.T) {
	root := copyTestdata(t)
	writeFixtureDate(t, root, "2026-02-01")

	out, err := runReadInRoot(t, root, "8818", "--json")
	require.NoError(t, err)

	var records []readJSONRecord
	require.NoError(t, json.Unmarshal([]byte(out), &records))
	require.Len(t, records, 1)
	assert.Equal(t, "20260104_8818", records[0].UID)
	assert.Equal(t, "2026-02-01T00:00:00Z", records[0].Date)
}

func writeFixtureDate(t *testing.T, root, date string) {
	t.Helper()
	path := filepath.Join(root, "2026/01/20260104_8818_meeting.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	updated := strings.Replace(string(data), "---\n", "---\ndate: "+date+"\n", 1)
	require.NoError(t, os.WriteFile(path, []byte(updated), 0o644))
}

func TestReadJSONNoTagsUsesEmptyArray(t *testing.T) {
	out, err := runRead(t, "6973", "--json")
	require.NoError(t, err)

	var raw []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &raw))
	tags, ok := raw[0]["tags"].([]any)
	require.True(t, ok)
	assert.Empty(t, tags)
}
