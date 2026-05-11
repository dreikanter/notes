package note

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractHashtagsBasic(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"simple", "hello #world", []string{"world"}},
		{"multiple", "#alpha and #beta here", []string{"alpha", "beta"}},
		{"digits and dashes", "#a-b_c #123 #x1", []string{"a-b_c", "123", "x1"}},
		{"slash terminates", "see #foo/bar", []string{"foo"}},
		{"punctuation after", "ok #tag, next.", []string{"tag"}},
		{"parens", "(#tag)", []string{"tag"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ExtractHashtags([]byte(c.in))
			assert.Equal(t, c.want, got)
		})
	}
}

func TestExtractHashtagsNegative(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"heading h1", "# heading not a tag"},
		{"heading h2", "## another heading"},
		{"indented heading", "   # still a heading"},
		{"word-prefixed", "foo#bar baz"},
		{"bare hash", "look here: # not-tag"},
		{"lone hash", "just # alone"},
		{"url anchor", "https://www.teamviewer.com/en/#screenshotsAnchor"},
		{"url anchor bare", "see example.com/path/#section for more"},
		{"backticked tag", "prose `#hashtag` continues"},
		{"chained hashes", "#one#two"},
		{"chained three", "prefix #one#two#three suffix"},
		{"domain anchor", "visit foo.bar/#frag here"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ExtractHashtags([]byte(c.in))
			assert.Empty(t, got)
		})
	}
}

func TestExtractHashtagsInlineCode(t *testing.T) {
	in := "real #out and `inline #in` and #back"
	want := []string{"out", "back"}
	got := ExtractHashtags([]byte(in))
	assert.Equal(t, want, got)
}

func TestExtractHashtagsFencedBlock(t *testing.T) {
	in := "before #a\n```\n#hidden\n#also-hidden\n```\nafter #b\n"
	want := []string{"a", "b"}
	got := ExtractHashtags([]byte(in))
	assert.Equal(t, want, got)
}

func TestExtractHashtagsFencedBlockWithInfoString(t *testing.T) {
	in := "top #ok\n```go\n// #comment\n```\nend #done\n"
	want := []string{"ok", "done"}
	got := ExtractHashtags([]byte(in))
	assert.Equal(t, want, got)
}

func TestExtractHashtagsCRLF(t *testing.T) {
	in := "before #a\r\n```\r\n#hidden\r\n```\r\nafter #b\r\n"
	want := []string{"a", "b"}
	got := ExtractHashtags([]byte(in))
	assert.Equal(t, want, got)
}

func TestExtractHashtagsBareHash(t *testing.T) {
	cases := []string{"#", "text # and #", "line #\nnext #"}
	for _, in := range cases {
		got := ExtractHashtags([]byte(in))
		assert.Empty(t, got)
	}
}

func TestIsTagRune(t *testing.T) {
	yes := []rune{'a', 'Z', '0', '9', '_', '-', 'é', 'ü', 'я', '中', '工', '٢'}
	for _, r := range yes {
		assert.True(t, IsTagRune(r), "expected %q to be a tag rune", r)
	}
	no := []rune{' ', '.', '/', '#', '\t', '\n', '!', ',', '(', ')'}
	for _, r := range no {
		assert.False(t, IsTagRune(r), "expected %q not to be a tag rune", r)
	}
}

func TestValidateTag(t *testing.T) {
	valid := []string{"a", "Work", "work-stuff", "snake_case", "1tag", "café", "工作", "mixed_тест-1", "ü"}
	for _, s := range valid {
		assert.NoError(t, ValidateTag(s), "expected %q to be valid", s)
	}
	invalid := []string{"", "has space", "tag.", "/foo", "#tag", "a/b", "a b c", "tag!"}
	for _, s := range invalid {
		assert.Error(t, ValidateTag(s), "expected %q to be invalid", s)
	}
}

func TestExtractHashtagsUnicode(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"latin extended", "#café", []string{"café"}},
		{"cjk", "#工作", []string{"工作"}},
		{"mixed script", "#mixed_тест-1", []string{"mixed_тест-1"}},
		{"adjacent prose latin", "café#bar", nil},
		{"adjacent prose cyrillic", "работа#tag", nil},
		{"period terminates", "#tag.", []string{"tag"}},
		{"chained still rejected", "#tag#other", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ExtractHashtags([]byte(c.in))
			assert.Equal(t, c.want, got)
		})
	}
}

func TestReplaceBodyHashtagsBasic(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#personal") }
	out, n := ReplaceBodyHashtags([]byte("here is #work, ok"), "work", replace)
	assert.Equal(t, 1, n)
	assert.Equal(t, "here is #personal, ok", string(out))
}

func TestReplaceBodyHashtagsCaseInsensitive(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#personal") }
	in := "a #Work b #WORK c #work d"
	out, n := ReplaceBodyHashtags([]byte(in), "work", replace)
	assert.Equal(t, 3, n)
	assert.Equal(t, "a #personal b #personal c #personal d", string(out))
}

func TestReplaceBodyHashtagsCallbackReceivesOriginal(t *testing.T) {
	var seen []string
	replace := func(token []byte) []byte {
		seen = append(seen, string(token))
		return token // strip '#' (callback shape future tags rm needs)
	}
	in := "x #Work y #work z"
	out, n := ReplaceBodyHashtags([]byte(in), "work", replace)
	assert.Equal(t, 2, n)
	assert.Equal(t, []string{"Work", "work"}, seen)
	assert.Equal(t, "x Work y work z", string(out))
}

func TestReplaceBodyHashtagsSkipsCodeFences(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#personal") }
	in := "before #work\n```\n#work hidden\n```\nafter #work\n"
	out, n := ReplaceBodyHashtags([]byte(in), "work", replace)
	assert.Equal(t, 2, n)
	assert.Equal(t, "before #personal\n```\n#work hidden\n```\nafter #personal\n", string(out))
}

func TestReplaceBodyHashtagsSkipsInlineBackticks(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#personal") }
	in := "real #work and `inline #work` and trailing #work"
	out, n := ReplaceBodyHashtags([]byte(in), "work", replace)
	assert.Equal(t, 2, n)
	assert.Equal(t, "real #personal and `inline #work` and trailing #personal", string(out))
}

func TestReplaceBodyHashtagsSkipsHeading(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#personal") }
	in := "# work\nbody #work here\n## work\n"
	out, n := ReplaceBodyHashtags([]byte(in), "work", replace)
	assert.Equal(t, 1, n)
	assert.Equal(t, "# work\nbody #personal here\n## work\n", string(out))
}

func TestReplaceBodyHashtagsSkipsURLAnchor(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#personal") }
	in := "see example.com/#work and https://x/y#work but #work here"
	out, n := ReplaceBodyHashtags([]byte(in), "work", replace)
	assert.Equal(t, 1, n)
	assert.Equal(t, "see example.com/#work and https://x/y#work but #personal here", string(out))
}

func TestReplaceBodyHashtagsUnicode(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#latte") }
	in := "drink #café please"
	out, n := ReplaceBodyHashtags([]byte(in), "café", replace)
	assert.Equal(t, 1, n)
	assert.Equal(t, "drink #latte please", string(out))
}

func TestReplaceBodyHashtagsNoMatch(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#personal") }
	in := "no tag here, just #other"
	out, n := ReplaceBodyHashtags([]byte(in), "work", replace)
	assert.Equal(t, 0, n)
	require.Equal(t, "no tag here, just #other", string(out))
}

func TestReplaceBodyHashtagsMultiplePerLine(t *testing.T) {
	replace := func(_ []byte) []byte { return []byte("#personal") }
	in := "#work and #work on same line\nand #work next\n"
	out, n := ReplaceBodyHashtags([]byte(in), "work", replace)
	assert.Equal(t, 3, n)
	assert.Equal(t, "#personal and #personal on same line\nand #personal next\n", string(out))
}
