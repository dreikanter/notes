package note

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func putForRename(t *testing.T, s *OSStore, day time.Time, tags []string, body string) Entry {
	t.Helper()
	entry, err := s.Put(Entry{
		Meta: Meta{Tags: tags, CreatedAt: day},
		Body: body,
	})
	require.NoError(t, err)
	return entry
}

func readBody(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	_, body, err := ParseNote(data)
	require.NoError(t, err)
	return string(body)
}

func readTags(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	fm, _, err := ParseNote(data)
	require.NoError(t, err)
	return fm.Tags
}

func TestRenameTag_FrontmatterOnly(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"work", "other"}, "no body tag\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	tags := readTags(t, res.ModifiedPaths[0])
	assert.ElementsMatch(t, []string{"other", "personal"}, tags)
}

func TestRenameTag_BodyOnly(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Body has the renamed tag plus an unrelated hashtag. The unrelated
	// hashtag is promoted to frontmatter on write (documented behavior).
	putForRename(t, s, day, nil, "see #work and #foo please\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	body := readBody(t, res.ModifiedPaths[0])
	assert.Equal(t, "see #personal and #foo please\n", body)
	tags := readTags(t, res.ModifiedPaths[0])
	assert.ElementsMatch(t, []string{"foo", "personal"}, tags)
}

func TestRenameTag_FrontmatterAndBody(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"work"}, "ref #work here\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	body := readBody(t, res.ModifiedPaths[0])
	assert.Equal(t, "ref #personal here\n", body)
	tags := readTags(t, res.ModifiedPaths[0])
	assert.Equal(t, []string{"personal"}, tags)
}

func TestRenameTag_NewTagWinsOverExistingCaseVariant(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"Personal", "work"}, "body\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	tags := readTags(t, res.ModifiedPaths[0])
	assert.Equal(t, []string{"personal"}, tags)
}

func TestRenameTag_BystanderUntouched(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	bystander := putForRename(t, s, day, []string{"Personal"}, "no work here\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	assert.Empty(t, res.ModifiedPaths)

	tags := readTags(t, s.AbsPath(bystander))
	assert.Equal(t, []string{"Personal"}, tags)
}

func TestRenameTag_BothFormsPresent(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"work", "personal"}, "body\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	tags := readTags(t, res.ModifiedPaths[0])
	assert.Equal(t, []string{"personal"}, tags)
}

func TestRenameTag_CaseOnlyRename(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"work"}, "ref #work\n")

	res, err := RenameTag(s, "work", "WORK", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	tags := readTags(t, res.ModifiedPaths[0])
	assert.Equal(t, []string{"WORK"}, tags)
	body := readBody(t, res.ModifiedPaths[0])
	assert.Equal(t, "ref #WORK\n", body)
}

func TestRenameTag_MixedCaseAcrossNotes(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"Work"}, "a\n")
	putForRename(t, s, day, []string{"WORK"}, "b\n")
	putForRename(t, s, day, []string{"work"}, "c\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 3)

	for _, p := range res.ModifiedPaths {
		assert.Equal(t, []string{"personal"}, readTags(t, p))
	}
}

func TestRenameTag_ZeroMatches(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	putForRename(t, s, day, []string{"other"}, "no tag here\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	assert.Empty(t, res.ModifiedPaths)
}

func TestRenameTag_DryRun(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	e := putForRename(t, s, day, []string{"work"}, "ref #work\n")
	originalInfo, err := os.Stat(s.AbsPath(e))
	require.NoError(t, err)
	originalMtime := originalInfo.ModTime()

	// Ensure a real mtime difference would be observable if we wrote.
	time.Sleep(10 * time.Millisecond)

	res, err := RenameTag(s, "work", "personal", RenameOpts{DryRun: true})
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

func TestRenameTag_PartialFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("filesystem-based failure injection is Unix-only")
	}
	s := newOSTestStore(t)
	d1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)

	e1 := putForRename(t, s, d1, []string{"work"}, "a\n")
	e2 := putForRename(t, s, d2, []string{"work"}, "b\n")

	// store.All returns newest-first, so e2 comes first. Block writes to
	// e1's path by replacing its .tmp target with a directory: WriteAtomic
	// writes "<path>.tmp" first, then renames, so a directory at that path
	// causes the write to fail with EISDIR. This works regardless of
	// process uid (chmod is bypassed for root, mkdir collisions are not).
	first := s.AbsPath(e2)
	second := s.AbsPath(e1)
	require.NotEqual(t, first, second)

	blocker := second + ".tmp"
	require.NoError(t, os.MkdirAll(blocker, 0o755))
	defer os.RemoveAll(blocker)

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.Error(t, err)

	require.NoError(t, os.RemoveAll(blocker))
	assert.Equal(t, []string{"personal"}, readTags(t, first))
	assert.Equal(t, []string{"work"}, readTags(t, second))

	assert.Equal(t, []string{first}, res.ModifiedPaths)
}

func TestRenameTag_BodyTagInsideCodeFenceIsNoop(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	body := "before\n```\n#work hidden\n```\nafter\n"
	e := putForRename(t, s, day, nil, body)

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	assert.Empty(t, res.ModifiedPaths)

	// File unchanged: still has #work in the fence and no frontmatter tags.
	got := readBody(t, s.AbsPath(e))
	assert.Equal(t, body, got)
}

func TestRenameTag_URLAndHeadingGuards(t *testing.T) {
	s := newOSTestStore(t)
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	body := "See [docs](https://example.com/#work) and #work and\n# work\nend\n"
	putForRename(t, s, day, nil, body)

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 1)

	got := readBody(t, res.ModifiedPaths[0])
	want := "See [docs](https://example.com/#work) and #personal and\n# work\nend\n"
	assert.Equal(t, want, got)
}

func TestRenameTag_ValidateTagRejectsInvalid(t *testing.T) {
	// Sanity check that ValidateTag catches bad new-tag inputs the CLI
	// will use for early validation.
	cases := []string{"", "has space", "a/b", "tag."}
	for _, s := range cases {
		err := ValidateTag(s)
		assert.Error(t, err, s)
	}
}

func TestRenameTag_OrderIsNewestFirst(t *testing.T) {
	s := newOSTestStore(t)

	// Three days, each gets one note with tag work.
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	putForRename(t, s, d1, []string{"work"}, "a\n")
	putForRename(t, s, d2, []string{"work"}, "b\n")
	putForRename(t, s, d3, []string{"work"}, "c\n")

	res, err := RenameTag(s, "work", "personal", RenameOpts{})
	require.NoError(t, err)
	require.Len(t, res.ModifiedPaths, 3)

	// Paths should be ordered newest-first (d3 then d2 then d1).
	got := make([]string, len(res.ModifiedPaths))
	for i, p := range res.ModifiedPaths {
		got[i] = filepath.Base(p)
	}
	// Date prefix order: 20260103 > 20260102 > 20260101
	sortedDesc := append([]string(nil), got...)
	sort.Sort(sort.Reverse(sort.StringSlice(sortedDesc)))
	assert.Equal(t, sortedDesc, got, "expected newest-first order")
	for _, p := range res.ModifiedPaths {
		assert.True(t, strings.HasSuffix(p, ".md"))
	}
}
