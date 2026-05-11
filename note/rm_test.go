package note

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveTag_FrontmatterOnly(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"work", "personal"}, "no body tag\n")

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	tags := readTags(t, res.ModifiedPaths[0])
	assert.Equal(t, []string{"personal"}, tags)
	body := readBody(t, res.ModifiedPaths[0])
	assert.Equal(t, "no body tag\n", body)
}

func TestRemoveTag_BodyOnly(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Body has the removed tag plus an unrelated hashtag. The unrelated
	// hashtag is promoted to frontmatter on write (documented behavior).
	putForRename(t, s, day, nil, "text #work more and #other\n")

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	body := readBody(t, res.ModifiedPaths[0])
	assert.Equal(t, "text work more and #other\n", body)
	tags := readTags(t, res.ModifiedPaths[0])
	assert.Equal(t, []string{"other"}, tags)
}

func TestRemoveTag_FrontmatterAndBody(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"work"}, "ref #work here\n")

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	body := readBody(t, res.ModifiedPaths[0])
	assert.Equal(t, "ref work here\n", body)
	tags := readTags(t, res.ModifiedPaths[0])
	assert.Empty(t, tags)
}

func TestRemoveTag_MixedCaseAcrossNotes(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, nil, "see #Work here\n")
	putForRename(t, s, day, nil, "see #WORK here\n")
	putForRename(t, s, day, nil, "see #work here\n")

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 3)

	bodies := make([]string, 0, 3)
	for _, p := range res.ModifiedPaths {
		bodies = append(bodies, readBody(t, p))
		assert.Empty(t, readTags(t, p))
	}
	assert.ElementsMatch(t, []string{
		"see Work here\n",
		"see WORK here\n",
		"see work here\n",
	}, bodies)
}

func TestRemoveTag_ZeroMatches(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"other"}, "no tag here\n")

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.NoError(t, err)
	assert.Empty(t, res.ModifiedPaths)
}

func TestRemoveTag_DryRun(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	e := putForRename(t, s, day, []string{"work"}, "ref #work\n")
	originalInfo, err := os.Stat(s.AbsPath(e))
	require.NoError(t, err)
	originalMtime := originalInfo.ModTime()

	time.Sleep(10 * time.Millisecond)

	res, err := RemoveTag(s, "work", RemoveOpts{DryRun: true})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	info, err := os.Stat(s.AbsPath(e))
	require.NoError(t, err)
	assert.True(t, info.ModTime().Equal(originalMtime), "file mtime should be unchanged in dry-run")

	tags := readTags(t, s.AbsPath(e))
	assert.Equal(t, []string{"work"}, tags)
	body := readBody(t, s.AbsPath(e))
	assert.Equal(t, "ref #work\n", body)
}

func TestRemoveTag_PartialFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("filesystem-based failure injection is Unix-only")
	}
	s := newOSTestStore(t)
	d1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)

	e1 := putForRename(t, s, d1, []string{"work"}, "a\n")
	e2 := putForRename(t, s, d2, []string{"work"}, "b\n")

	first := s.AbsPath(e2)
	second := s.AbsPath(e1)
	require.NotEqual(t, first, second)

	blocker := second + ".tmp"
	require.NoError(t, os.MkdirAll(blocker, 0o755))
	defer os.RemoveAll(blocker)

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.Error(t, err)

	require.NoError(t, os.RemoveAll(blocker))
	assert.Empty(t, readTags(t, first))
	assert.Equal(t, []string{"work"}, readTags(t, second))

	assert.Equal(t, []string{first}, res.ModifiedPaths)
}

func TestRemoveTag_BodyTagInsideCodeFenceIsNoop(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	body := "before\n```\n#work hidden\n```\nafter\n"
	e := putForRename(t, s, day, nil, body)

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.NoError(t, err)
	assert.Empty(t, res.ModifiedPaths)

	got := readBody(t, s.AbsPath(e))
	assert.Equal(t, body, got)
}

func TestRemoveTag_URLAndHeadingGuards(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	body := "See [docs](https://example.com/#work) and #work and\n# work\nend\n"
	putForRename(t, s, day, nil, body)

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	got := readBody(t, res.ModifiedPaths[0])
	want := "See [docs](https://example.com/#work) and work and\n# work\nend\n"
	assert.Equal(t, want, got)
}

func TestRemoveTag_InlineCodeIsNoop(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	body := "prose `#work` continues\n"
	e := putForRename(t, s, day, nil, body)

	res, err := RemoveTag(s, "work", RemoveOpts{})
	require.NoError(t, err)
	assert.Empty(t, res.ModifiedPaths)

	got := readBody(t, s.AbsPath(e))
	assert.Equal(t, body, got)
}

func TestRemoveTag_UnicodeCasePreserved(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"café"}, "drink #Café please\n")

	res, err := RemoveTag(s, "café", RemoveOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	body := readBody(t, res.ModifiedPaths[0])
	assert.Equal(t, "drink Café please\n", body)
	tags := readTags(t, res.ModifiedPaths[0])
	assert.Empty(t, tags)
}
