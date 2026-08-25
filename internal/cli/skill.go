package cli

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillContent string

// homeDir is overridable in tests so install-mode paths under "~" can be
// redirected to a temporary directory.
var homeDir = os.UserHomeDir

// installTarget is a supported skills-root location. Each target maps to
// one filesystem destination; multiple AI agents may read from the same
// destination (notably ~/.agents/skills/, which is read by Codex,
// Cursor, OpenCode, Pi, VS Code Copilot, Warp, and others).
//
// RootDir is the directory whose existence indicates the user actually
// uses a harness that reads this location. It is the only directory
// the install path may not create — if it is missing, the install fails
// with an actionable error rather than materialising an unfamiliar
// dotdir. Subdirectories between RootDir and PathFor are created on
// demand.
type installTarget struct {
	Name    string
	PathFor func() (string, error)
	RootDir func() (string, error)
}

func underHome(parts ...string) func() (string, error) {
	return func() (string, error) {
		home, err := homeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(append([]string{home}, parts...)...), nil
	}
}

var targets = []installTarget{
	{
		Name:    "claude",
		PathFor: underHome(".claude", "skills", "notes", "SKILL.md"),
		RootDir: underHome(".claude", "skills"),
	},
	{
		Name:    "pi",
		PathFor: underHome(".pi", "agent", "skills", "notes", "SKILL.md"),
		RootDir: underHome(".pi", "agent", "skills"),
	},
	{
		Name:    "agents",
		PathFor: underHome(".agents", "skills", "notes", "SKILL.md"),
		RootDir: underHome(".agents", "skills"),
	},
}

// Detect reports whether the target appears to be in use on this
// machine, using only os.Stat on the declared RootDir.
func (t installTarget) Detect() (bool, error) {
	dir, err := t.RootDir()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(dir)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// installAction is a single planned filesystem operation against one
// target. See specs/001-generate-skill/data-model.md.
type installAction struct {
	Target string
	Path   string
	Action string
	Error  error
}

const (
	actionCreate    = "create"
	actionSkip      = "skip"
	actionConflict  = "conflict"
	actionOverwrite = "overwrite"
)

var (
	skillInstall bool
	skillTarget  string
	skillForce   bool
	skillDryRun  bool
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Print or install the notes agent skill",
	Long: `Print a self-contained markdown document describing how an AI
agent should drive the notes CLI. The skill is authored as a markdown
file in the repository and embedded into the binary at build time, so
the same bytes are emitted across machines.

With --install, the skill is written to a known skills-root location
instead of stdout. Each target names a filesystem destination, not an
agent — multiple harnesses read from the same locations.

Supported --target values:
  claude    ~/.claude/skills/notes/SKILL.md
            Read by: Claude Code
  pi        ~/.pi/agent/skills/notes/SKILL.md
            Read by: Pi
  agents    ~/.agents/skills/notes/SKILL.md
            Read by: Codex, Cursor, OpenCode, Pi, VS Code Copilot, Warp,
            and other harnesses adopting the cross-harness convention

All targets receive the same SKILL.md body (the format is the Agent
Skills standard, https://agentskills.io/specification). A target is
"detected" when its skills root directory exists. Without --target,
--install writes to every detected target.

Actions (install mode):
  create     destination did not exist; written
  skip       destination existed with identical content; not written
  conflict   destination existed with different content; not written, exit non-zero
  overwrite  destination existed with different content and --force was set; written`,
	Args: cobra.NoArgs,
	RunE: skillRunE,
}

func skillRunE(cmd *cobra.Command, _ []string) error {
	if err := validateSkillFlags(); err != nil {
		return err
	}

	if !skillInstall {
		_, err := io.WriteString(cmd.OutOrStdout(), skillContent)
		return err
	}

	resolved, err := resolveTargets()
	if err != nil {
		return err
	}

	actions := planInstall(skillContent, resolved)
	if err := applyInstall(actions, skillDryRun); err != nil {
		return err
	}
	printActions(cmd.OutOrStdout(), actions, skillDryRun)
	return exitErrorFor(actions)
}

func validateSkillFlags() error {
	if !skillInstall {
		switch {
		case skillTarget != "":
			return errors.New("--target requires --install")
		case skillForce:
			return errors.New("--force requires --install")
		case skillDryRun:
			return errors.New("--dry-run requires --install")
		}
	}
	if skillTarget != "" && findTarget(skillTarget) == nil {
		return fmt.Errorf("unknown target %q (supported: %s)", skillTarget, targetNamesList())
	}
	return nil
}

func findTarget(name string) *installTarget {
	i := slices.IndexFunc(targets, func(t installTarget) bool { return t.Name == name })
	if i < 0 {
		return nil
	}
	return &targets[i]
}

func targetNamesList() string {
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.Name
	}
	slices.Sort(names)
	return strings.Join(names, ", ")
}

// resolveTargets returns the targets to act on for the current
// invocation: either the single target named by --target, or every
// detected target when --target is empty.
func resolveTargets() ([]installTarget, error) {
	if skillTarget != "" {
		return []installTarget{*findTarget(skillTarget)}, nil
	}
	var detected []installTarget
	for _, t := range targets {
		ok, err := t.Detect()
		if err != nil {
			return nil, err
		}
		if ok {
			detected = append(detected, t)
		}
	}
	if len(detected) == 0 {
		return nil, fmt.Errorf("no supported target detected; pass --target explicitly (supported: %s)", targetNamesList())
	}
	return detected, nil
}

// planInstall computes the action for each target without writing
// anything. An OS error while reading an existing destination is
// captured in the action's Error field rather than aborting the whole
// plan.
func planInstall(content string, targets []installTarget) []installAction {
	bytesContent := []byte(content)
	actions := make([]installAction, len(targets))
	for i, t := range targets {
		actions[i] = planOne(t, bytesContent)
	}
	return actions
}

func planOne(t installTarget, content []byte) installAction {
	detected, err := t.Detect()
	if err != nil {
		return installAction{Target: t.Name, Error: err}
	}
	if !detected {
		root, _ := t.RootDir()
		return installAction{Target: t.Name, Error: fmt.Errorf("skills root directory not found: %s", root)}
	}
	path, err := t.PathFor()
	if err != nil {
		return installAction{Target: t.Name, Error: err}
	}
	existing, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return installAction{Target: t.Name, Path: path, Action: actionCreate}
	case err != nil:
		return installAction{Target: t.Name, Path: path, Error: err}
	}
	if string(existing) == string(content) {
		return installAction{Target: t.Name, Path: path, Action: actionSkip}
	}
	if skillForce {
		return installAction{Target: t.Name, Path: path, Action: actionOverwrite}
	}
	return installAction{Target: t.Name, Path: path, Action: actionConflict}
}

// applyInstall performs the writes implied by the action plan. In
// dryRun mode, it does nothing and returns nil.
func applyInstall(actions []installAction, dryRun bool) error {
	if dryRun {
		return nil
	}
	content := []byte(skillContent)
	for i := range actions {
		a := &actions[i]
		if a.Error != nil {
			continue
		}
		if a.Action != actionCreate && a.Action != actionOverwrite {
			continue
		}
		if err := writeSkillFile(a.Path, content); err != nil {
			a.Error = err
		}
	}
	return nil
}

// writeSkillFile materialises any missing intermediate directories under
// the target's already-existing root directory and writes the file.
// planOne guarantees the target's RootDir exists before this is called.
func writeSkillFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func printActions(out io.Writer, actions []installAction, dryRun bool) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, a := range actions {
		if a.Error != nil {
			fmt.Fprintf(w, "error\t%s\t%s\n", a.Target, a.Error)
			continue
		}
		verb := a.Action
		if dryRun {
			verb = "would " + verb
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", verb, a.Target, a.Path)
	}
	_ = w.Flush()
}

func exitErrorFor(actions []installAction) error {
	var problems []string
	for _, a := range actions {
		switch {
		case a.Error != nil:
			problems = append(problems, fmt.Sprintf("%s: %s", a.Target, a.Error))
		case a.Action == actionConflict:
			problems = append(problems, fmt.Sprintf("%s: destination exists with different content; pass --force to overwrite", a.Target))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return errors.New(strings.Join(problems, "; "))
}

func registerSkillFlags() {
	skillCmd.Flags().BoolVar(&skillInstall, "install", false, "install the skill into one or more skills-root locations")
	skillCmd.Flags().StringVar(&skillTarget, "target", "", "install only into the named target (default: auto-detect)")
	skillCmd.Flags().BoolVar(&skillForce, "force", false, "overwrite an existing destination with diverging content")
	skillCmd.Flags().BoolVarP(&skillDryRun, "dry-run", "n", false, "print planned actions but do not write any files")
}

func init() {
	registerSkillFlags()
	rootCmd.AddCommand(skillCmd)
}
