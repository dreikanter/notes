package note

import (
	"strings"
)

// RenameOpts configures RenameTag.
type RenameOpts struct {
	// DryRun reports the modified paths without writing.
	DryRun bool
}

// RenameResult is the outcome of a RenameTag call.
type RenameResult struct {
	// ModifiedPaths lists the absolute path of every note that was (or
	// would be, in dry-run mode) modified, in newest-first order.
	ModifiedPaths []string
}

// RenameTag rewrites every occurrence of oldTag (matched case-insensitively)
// across the store, both in frontmatter "tags:" lists and in body "#hashtag"
// tokens, replacing it with newTag written literally. The store root is
// locked for the duration. On mid-run failure, RenameTag returns the error
// together with the partial path list of notes already written.
func RenameTag(store *OSStore, oldTag, newTag string, opts RenameOpts) (RenameResult, error) {
	if !opts.DryRun {
		unlock, err := lockStoreRoot(store.Root())
		if err != nil {
			return RenameResult{}, err
		}
		defer unlock()
	}

	lowerOld := strings.ToLower(oldTag)

	entries, err := store.All(WithTag(oldTag))
	if err != nil {
		return RenameResult{}, err
	}

	var result RenameResult
	for _, entry := range entries {
		newTags, tagsChanged := rewriteMetaTags(entry.Meta.Tags, lowerOld, newTag)
		newBody, bodyN := ReplaceBodyHashtags([]byte(entry.Body), lowerOld, func(_ []byte) []byte {
			return []byte("#" + newTag)
		})

		if !tagsChanged && bodyN == 0 {
			continue
		}

		entry.Meta.Tags = newTags
		entry.Body = string(newBody)

		if opts.DryRun {
			result.ModifiedPaths = append(result.ModifiedPaths, store.AbsPath(entry))
			continue
		}

		saved, err := store.Put(entry)
		if err != nil {
			return result, err
		}
		result.ModifiedPaths = append(result.ModifiedPaths, store.AbsPath(saved))
	}

	return result, nil
}

// rewriteMetaTags returns the rewritten tag list and whether any change
// occurred. Every case-variant of lowerOld is dropped; if at least one was
// present, newTag is appended and any pre-existing case variant of newTag
// is dropped first so the user-typed surface form wins.
func rewriteMetaTags(tags []string, lowerOld, newTag string) ([]string, bool) {
	matched := false
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if strings.ToLower(t) == lowerOld {
			matched = true
			continue
		}
		out = append(out, t)
	}
	if !matched {
		return tags, false
	}
	lowerNew := strings.ToLower(newTag)
	filtered := out[:0]
	for _, t := range out {
		if strings.ToLower(t) != lowerNew {
			filtered = append(filtered, t)
		}
	}
	return append(filtered, newTag), true
}
