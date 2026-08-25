package note

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// IsTagRune reports whether r is allowed inside a tag token: any unicode
// letter, any unicode digit, '_', or '-'.
func IsTagRune(r rune) bool {
	if r == '_' || r == '-' {
		return true
	}
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// ValidateTag reports whether s is a non-empty string of tag runes.
// Returns nil on valid input; an error describing the offending rune
// otherwise.
func ValidateTag(s string) error {
	if s == "" {
		return fmt.Errorf("tag is empty")
	}
	for _, r := range s {
		if !IsTagRune(r) {
			return fmt.Errorf("invalid tag character %q in %q", r, s)
		}
	}
	return nil
}

// isWordRune reports whether r is a "word" rune: unicode letter, unicode
// digit, or '_'. Used to detect that a '#' is attached to prose.
func isWordRune(r rune) bool {
	if r == '_' {
		return true
	}
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isHashtagLeadingRune reports whether r may legally precede a '#' that
// starts a hashtag. Excludes word runes (any unicode letter/digit/'_')
// and URL-path bytes ('/', ':', '.', '?', '=', '&', '~', '#').
func isHashtagLeadingRune(r rune) bool {
	if isWordRune(r) {
		return false
	}
	switch r {
	case '/', ':', '.', '?', '=', '&', '~', '#':
		return false
	}
	return true
}

// ExtractHashtags scans body text and returns hashtag tokens (without the
// leading '#'), preserving source order and including duplicates. Rules:
//   - Lines whose first non-whitespace content is a run of '#' followed by
//     whitespace or end-of-line are Markdown headings and are skipped entirely.
//   - Fenced code blocks (``` on a line, optionally indented, with optional
//     info string) are skipped until the next fence line. Tilde fences (~~~)
//     are not recognised.
//   - Inline backtick spans on a single line are skipped. An unclosed
//     backtick suppresses hashtags for the remainder of its line.
//   - A '#' preceded by a word rune (any unicode letter/digit/'_') or a
//     URL-path byte (`/`, `:`, `.`, `?`, `=`, `&`, `~`, `#`) is not a tag.
//     This prevents matches inside URLs (`example.com/#anchor`), inline
//     chains (`#one#two`), and adjacency to prose in any script
//     (`café#bar`, `работа#tag`).
//   - Tag characters are any unicode letter, any unicode digit, '_', or
//     '-'; other runes terminate a tag. A bare '#' with no following tag
//     rune produces no output. A tag immediately followed by another '#'
//     (e.g. `#one#two`) is rejected.
func ExtractHashtags(body []byte) []string {
	var out []string
	scanBodyAbs(body, func(_ int, line []byte, j, k int) {
		out = append(out, string(line[j+1:k]))
	})
	return out
}

// ReplaceBodyHashtags walks body using the same rules as ExtractHashtags.
// For every hashtag token whose lowercased value equals lowerMatch, the
// "#token" span (the leading '#' plus the token bytes) is replaced with
// the bytes returned by replace(token); token is the original byte slice
// of the hashtag without '#'. Returns the rewritten body and the number
// of replacements. body is returned unchanged (and n=0) if no
// replacements occur.
func ReplaceBodyHashtags(body []byte, lowerMatch string, replace func(token []byte) []byte) (out []byte, n int) {
	type span struct {
		start int
		end   int
		token []byte
	}
	var spans []span

	scanBodyAbs(body, func(absLineStart int, line []byte, j, k int) {
		token := line[j+1 : k]
		if strings.ToLower(string(token)) != lowerMatch {
			return
		}
		spans = append(spans, span{
			start: absLineStart + j,
			end:   absLineStart + k,
			token: token,
		})
	})

	if len(spans) == 0 {
		return body, 0
	}

	var buf bytes.Buffer
	buf.Grow(len(body))
	last := 0
	for _, sp := range spans {
		buf.Write(body[last:sp.start])
		buf.Write(replace(sp.token))
		last = sp.end
	}
	buf.Write(body[last:])
	return buf.Bytes(), len(spans)
}

// scanBodyAbs invokes fn for every hashtag token found in body. line is
// the current line's bytes (without trailing CR/LF); j is the index of
// the '#' within line and k is the exclusive end of the token. The
// absLineStart parameter is the absolute byte offset of the line's start
// within body, so callers can translate (j, k) into absolute body
// offsets when rewriting.
func scanBodyAbs(body []byte, fn func(absLineStart int, line []byte, j, k int)) {
	inFence := false
	fence := []byte("```")

	pos := 0
	for pos < len(body) {
		lineStart := pos
		nl := bytes.IndexByte(body[pos:], '\n')
		var line []byte
		if nl < 0 {
			line = body[pos:]
			pos = len(body)
		} else {
			line = body[pos : pos+nl]
			pos = pos + nl + 1
		}
		if n := len(line); n > 0 && line[n-1] == '\r' {
			line = line[:n-1]
		}

		trim := 0
		for trim < len(line) && (line[trim] == ' ' || line[trim] == '\t') {
			trim++
		}

		if bytes.HasPrefix(line[trim:], fence) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		// Heading detection: leading '#' run followed by space/tab/EOL.
		if trim < len(line) && line[trim] == '#' {
			k := trim
			for k < len(line) && line[k] == '#' {
				k++
			}
			if k == len(line) || line[k] == ' ' || line[k] == '\t' {
				continue
			}
		}

		scanLineHashtags(line, lineStart, fn)
	}
}

// scanLineHashtags scans a single line (already stripped of CR/LF) for
// hashtag tokens, honoring inline-backtick spans and the leading-rune
// rule. It invokes fn for every accepted token.
func scanLineHashtags(line []byte, absLineStart int, fn func(absLineStart int, line []byte, j, k int)) {
	inInline := false
	j := 0
	for j < len(line) {
		c := line[j]
		if c == '`' {
			inInline = !inInline
			j++
			continue
		}
		if c != '#' || inInline {
			j++
			continue
		}
		// Check the rune immediately preceding '#'.
		if j > 0 {
			r, _ := utf8.DecodeLastRune(line[:j])
			if !isHashtagLeadingRune(r) {
				j++
				continue
			}
		}
		// Consume tag runes starting at j+1.
		k := j + 1
		for k < len(line) {
			r, size := utf8.DecodeRune(line[k:])
			if r == utf8.RuneError && size <= 1 {
				break
			}
			if !IsTagRune(r) {
				break
			}
			k += size
		}
		if k > j+1 && (k == len(line) || line[k] != '#') {
			fn(absLineStart, line, j, k)
		}
		if k > j+1 {
			j = k
		} else {
			j++
		}
	}
}

// hasAllTags reports whether every entry in required appears in noteTags,
// case-insensitively. Used by both MemStore and OSStore for WithTag filtering.
func hasAllTags(noteTags []string, required []string) bool {
	set := make(map[string]struct{}, len(noteTags))
	for _, t := range noteTags {
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, r := range required {
		if _, ok := set[strings.ToLower(r)]; !ok {
			return false
		}
	}
	return true
}

// computeMergedTags builds the sorted, lowercased, deduplicated union of
// frontmatter tags and body hashtags. bodyHashtags is assumed already
// lowercased (as produced by normalizeHashtags). Returns nil when the
// union is empty.
func computeMergedTags(fmTags, bodyHashtags []string) []string {
	set := make(map[string]struct{}, len(fmTags)+len(bodyHashtags))
	for _, t := range fmTags {
		if t == "" {
			continue
		}
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, t := range bodyHashtags {
		set[t] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	return slices.Sorted(maps.Keys(set))
}

// normalizeHashtags lowercases and deduplicates a hashtag list from
// ExtractHashtags into the canonical form merged into Meta.Tags by
// OSStore.
func normalizeHashtags(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(raw))
	for _, t := range raw {
		if t == "" {
			continue
		}
		set[strings.ToLower(t)] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	return slices.Sorted(maps.Keys(set))
}
