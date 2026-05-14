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

func runSkill(t *testing.T, args ...string) (string, error) {
	t.Helper()

	skillCmd.ResetFlags()
	registerSkillFlags()
	skillInstall = false
	skillTarget = ""
	skillForce = false
	skillDryRun = false

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"skill"}, args...))

	execErr := rootCmd.Execute()
	return buf.String(), execErr
}

// sandboxHome redirects the package-level homeDir resolver to a fresh
// tempdir for the lifetime of the test. Any target names passed in
// installed have their RootDir() materialised inside the sandbox so
// Detect() succeeds for them.
func sandboxHome(t *testing.T, installed ...string) string {
	t.Helper()
	dir := t.TempDir()
	prev := homeDir
	homeDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { homeDir = prev })

	for _, name := range installed {
		tg := findTarget(name)
		require.NotNil(t, tg, "unknown target %q", name)
		root, err := tg.RootDir()
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(root, 0o755))
	}
	return dir
}

func TestSkillStdoutHasFrontmatter(t *testing.T) {
	out, err := runSkill(t)
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(out, "---\n"), "missing opening frontmatter delimiter")
	assert.Contains(t, out, "name: notes\n")
	assert.Contains(t, out, "description: ")
	assert.Contains(t, out, "\n---\n\n")
}

func TestSkillStdoutDeterministic(t *testing.T) {
	out1, err := runSkill(t)
	require.NoError(t, err)
	out2, err := runSkill(t)
	require.NoError(t, err)
	assert.Equal(t, out1, out2, "two invocations must produce byte-identical output")
}

func TestSkillStdoutListsKnownCommands(t *testing.T) {
	out, err := runSkill(t)
	require.NoError(t, err)

	for _, name := range []string{"new", "new-todo", "ls", "read", "append", "annotate", "resolve", "rm", "tags", "update", "config", "skill"} {
		assert.Contains(t, out, "notes "+name, "command %s missing from skill body", name)
	}
}

func TestSkillStdoutOmitsHelpAndCompletion(t *testing.T) {
	out, err := runSkill(t)
	require.NoError(t, err)
	assert.NotContains(t, out, "`notes help`")
	assert.NotContains(t, out, "`notes completion`")
}

func TestSkillStdoutEqualsEmbeddedFile(t *testing.T) {
	out, err := runSkill(t)
	require.NoError(t, err)
	assert.Equal(t, skillContent, out, "stdout output must equal the embedded skill.md byte-for-byte")
}

func TestSkillStdoutListsPersistentFlags(t *testing.T) {
	out, err := runSkill(t)
	require.NoError(t, err)
	assert.Contains(t, out, "`--path`")
}

func TestSkillStdoutMentionsStoreLayout(t *testing.T) {
	out, err := runSkill(t)
	require.NoError(t, err)
	assert.Contains(t, out, "NOTES_PATH")
	assert.Contains(t, out, "YYYY/MM/YYYYMMDD_ID")
}

func TestSkillFlagWithoutInstallIsError(t *testing.T) {
	_, err := runSkill(t, "--target=claude")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--target requires --install")

	_, err = runSkill(t, "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--force requires --install")

	_, err = runSkill(t, "--dry-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--dry-run requires --install")
}

func TestSkillUnknownTarget(t *testing.T) {
	sandboxHome(t, "claude")
	_, err := runSkill(t, "--install", "--target=bogus")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown target")
	assert.Contains(t, err.Error(), "claude")
}

func TestSkillInstallCreate(t *testing.T) {
	home := sandboxHome(t, "claude")
	out, err := runSkill(t, "--install", "--target=claude")
	require.NoError(t, err)

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	assert.FileExists(t, target)
	assert.Contains(t, out, "create")
	assert.Contains(t, out, target)

	// File content matches stdout mode byte-for-byte.
	stdoutBytes, err := runSkill(t)
	require.NoError(t, err)
	written, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, stdoutBytes, string(written))
}

func TestSkillInstallSkipOnRerun(t *testing.T) {
	home := sandboxHome(t, "claude")
	_, err := runSkill(t, "--install", "--target=claude")
	require.NoError(t, err)

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	beforeStat, err := os.Stat(target)
	require.NoError(t, err)

	out, err := runSkill(t, "--install", "--target=claude")
	require.NoError(t, err)
	assert.Contains(t, out, "skip")

	afterStat, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, beforeStat.ModTime(), afterStat.ModTime(), "skip must not touch the file")
}

func TestSkillInstallConflictWithoutForce(t *testing.T) {
	home := sandboxHome(t, "claude")
	_, err := runSkill(t, "--install", "--target=claude")
	require.NoError(t, err)

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	require.NoError(t, os.WriteFile(target, []byte("local changes"), 0o644))

	out, err := runSkill(t, "--install", "--target=claude")
	require.Error(t, err, "conflict must exit non-zero")
	assert.Contains(t, out, "conflict")

	current, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "local changes", string(current), "conflict must not write")
}

func TestSkillInstallForceOverwrites(t *testing.T) {
	home := sandboxHome(t, "claude")
	_, err := runSkill(t, "--install", "--target=claude")
	require.NoError(t, err)

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	require.NoError(t, os.WriteFile(target, []byte("local changes"), 0o644))

	out, err := runSkill(t, "--install", "--target=claude", "--force")
	require.NoError(t, err)
	assert.Contains(t, out, "overwrite")

	written, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.NotEqual(t, "local changes", string(written))
}

func TestSkillInstallDryRun(t *testing.T) {
	home := sandboxHome(t, "claude")
	out, err := runSkill(t, "--install", "--target=claude", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "would create")

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	_, statErr := os.Stat(target)
	assert.True(t, os.IsNotExist(statErr), "dry-run must not write any files")
}

func TestSkillInstallAutoDetectFindsClaude(t *testing.T) {
	home := sandboxHome(t, "claude")
	out, err := runSkill(t, "--install")
	require.NoError(t, err)
	assert.Contains(t, out, "create")
	assert.Contains(t, out, "claude")

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	assert.FileExists(t, target)
}

func TestSkillInstallAutoDetectNoneFound(t *testing.T) {
	sandboxHome(t)
	_, err := runSkill(t, "--install")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no supported target detected")
	assert.Contains(t, err.Error(), "claude")
}

func TestSkillInstallMissingRootDirectoryErrors(t *testing.T) {
	sandboxHome(t) // no ~/.claude/skills
	_, err := runSkill(t, "--install", "--target=claude")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skills root directory not found")
}

func TestSkillInstallAllTargetsAutoDetect(t *testing.T) {
	home := sandboxHome(t, "claude", "pi", "agents")

	out, err := runSkill(t, "--install")
	require.NoError(t, err)
	for _, name := range []string{"claude", "pi", "agents"} {
		assert.Contains(t, out, name)
	}

	for _, p := range []string{
		filepath.Join(home, ".claude", "skills", "notes", "SKILL.md"),
		filepath.Join(home, ".pi", "agent", "skills", "notes", "SKILL.md"),
		filepath.Join(home, ".agents", "skills", "notes", "SKILL.md"),
	} {
		assert.FileExists(t, p)
	}
}

func TestSkillInstallEachTargetExplicit(t *testing.T) {
	cases := []struct {
		name string
		path []string
	}{
		{"claude", []string{".claude", "skills", "notes", "SKILL.md"}},
		{"pi", []string{".pi", "agent", "skills", "notes", "SKILL.md"}},
		{"agents", []string{".agents", "skills", "notes", "SKILL.md"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := sandboxHome(t, tc.name)
			out, err := runSkill(t, "--install", "--target="+tc.name)
			require.NoError(t, err)
			assert.Contains(t, out, "create")

			target := filepath.Join(append([]string{home}, tc.path...)...)
			assert.FileExists(t, target)
		})
	}
}

func TestSkillHelpDocumentsTargets(t *testing.T) {
	out, err := runSkill(t, "--help")
	require.NoError(t, err)
	assert.Contains(t, out, "Supported --target values:")
	for _, fragment := range []string{
		"claude",
		"~/.claude/skills/notes/SKILL.md",
		"Claude Code",
		"pi",
		"~/.pi/agent/skills/notes/SKILL.md",
		"Pi",
		"agents",
		"~/.agents/skills/notes/SKILL.md",
		"Codex",
		"agentskills.io/specification",
	} {
		assert.Contains(t, out, fragment)
	}
}
