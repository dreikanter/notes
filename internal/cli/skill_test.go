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
	skillAgent = ""
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
// tempdir for the lifetime of the test. Returns the sandbox path.
func sandboxHome(t *testing.T, withClaudeSkills bool) string {
	t.Helper()
	dir := t.TempDir()
	if withClaudeSkills {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude", "skills"), 0o755))
	}
	prev := homeDir
	homeDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { homeDir = prev })
	return dir
}

func TestSkillStdoutHasFrontmatter(t *testing.T) {
	out, err := runSkill(t)
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(out, "---\n"), "missing opening frontmatter delimiter")
	assert.Contains(t, out, "name: notes\n")
	assert.Contains(t, out, "description: Use when ")
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

	for _, name := range []string{"new", "ls", "read", "append", "annotate", "resolve", "rm", "tags", "update", "config", "skill"} {
		assert.Contains(t, out, "`notes "+name+"`", "command %s missing from skill body", name)
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
	assert.Contains(t, out, "YYYY/YYYYMMDD_NNNN.md")
}

func TestSkillFlagWithoutInstallIsError(t *testing.T) {
	_, err := runSkill(t, "--agent=claude")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--agent requires --install")

	_, err = runSkill(t, "--force")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--force requires --install")

	_, err = runSkill(t, "--dry-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--dry-run requires --install")
}

func TestSkillUnknownAgent(t *testing.T) {
	sandboxHome(t, true)
	_, err := runSkill(t, "--install", "--agent=bogus")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown agent")
	assert.Contains(t, err.Error(), "claude")
}

func TestSkillInstallCreate(t *testing.T) {
	home := sandboxHome(t, true)
	out, err := runSkill(t, "--install", "--agent=claude")
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
	home := sandboxHome(t, true)
	_, err := runSkill(t, "--install", "--agent=claude")
	require.NoError(t, err)

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	beforeStat, err := os.Stat(target)
	require.NoError(t, err)

	out, err := runSkill(t, "--install", "--agent=claude")
	require.NoError(t, err)
	assert.Contains(t, out, "skip")

	afterStat, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, beforeStat.ModTime(), afterStat.ModTime(), "skip must not touch the file")
}

func TestSkillInstallConflictWithoutForce(t *testing.T) {
	home := sandboxHome(t, true)
	_, err := runSkill(t, "--install", "--agent=claude")
	require.NoError(t, err)

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	require.NoError(t, os.WriteFile(target, []byte("local changes"), 0o644))

	out, err := runSkill(t, "--install", "--agent=claude")
	require.Error(t, err, "conflict must exit non-zero")
	assert.Contains(t, out, "conflict")

	current, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "local changes", string(current), "conflict must not write")
}

func TestSkillInstallForceOverwrites(t *testing.T) {
	home := sandboxHome(t, true)
	_, err := runSkill(t, "--install", "--agent=claude")
	require.NoError(t, err)

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	require.NoError(t, os.WriteFile(target, []byte("local changes"), 0o644))

	out, err := runSkill(t, "--install", "--agent=claude", "--force")
	require.NoError(t, err)
	assert.Contains(t, out, "overwrite")

	written, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.NotEqual(t, "local changes", string(written))
}

func TestSkillInstallDryRun(t *testing.T) {
	home := sandboxHome(t, true)
	out, err := runSkill(t, "--install", "--agent=claude", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "would create")

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	_, statErr := os.Stat(target)
	assert.True(t, os.IsNotExist(statErr), "dry-run must not write any files")
}

func TestSkillInstallAutoDetectFindsClaude(t *testing.T) {
	home := sandboxHome(t, true)
	out, err := runSkill(t, "--install")
	require.NoError(t, err)
	assert.Contains(t, out, "create")
	assert.Contains(t, out, "claude")

	target := filepath.Join(home, ".claude", "skills", "notes", "SKILL.md")
	assert.FileExists(t, target)
}

func TestSkillInstallAutoDetectNoneFound(t *testing.T) {
	sandboxHome(t, false)
	_, err := runSkill(t, "--install")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no supported agent detected")
	assert.Contains(t, err.Error(), "claude")
}

func TestSkillInstallMissingParentDirectoryErrors(t *testing.T) {
	sandboxHome(t, false) // no ~/.claude/skills
	_, err := runSkill(t, "--install", "--agent=claude")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent skills directory not found")
}

func TestSkillInstallMultipleAgents(t *testing.T) {
	home := sandboxHome(t, true)

	fakeDir := filepath.Join(home, ".fake", "skills")
	require.NoError(t, os.MkdirAll(fakeDir, 0o755))

	prev := agents
	agents = append(append([]agentTarget{}, agents...), agentTarget{
		Name: "fake",
		PathFor: func() (string, error) {
			h, err := homeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(h, ".fake", "skills", "notes", "SKILL.md"), nil
		},
		Detect: func() (bool, error) {
			h, err := homeDir()
			if err != nil {
				return false, err
			}
			_, statErr := os.Stat(filepath.Join(h, ".fake", "skills"))
			return statErr == nil, nil
		},
	})
	t.Cleanup(func() { agents = prev })

	out, err := runSkill(t, "--install")
	require.NoError(t, err)
	assert.Contains(t, out, "claude")
	assert.Contains(t, out, "fake")
	assert.FileExists(t, filepath.Join(home, ".claude", "skills", "notes", "SKILL.md"))
	assert.FileExists(t, filepath.Join(home, ".fake", "skills", "notes", "SKILL.md"))
}

func TestSkillHelpDocumentsAgents(t *testing.T) {
	out, err := runSkill(t, "--help")
	require.NoError(t, err)
	assert.Contains(t, out, "Supported --agent values:")
	assert.Contains(t, out, "claude")
	assert.Contains(t, out, "~/.claude/skills/notes/SKILL.md")
}
