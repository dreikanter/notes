package cli

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillContent string

// homeDir is overridable in tests so install-mode paths under "~" can be
// redirected to a temporary directory.
var homeDir = os.UserHomeDir

// agentTarget is a supported AI assistant. The registry is the package
// variable agents below; new entries are added by appending.
type agentTarget struct {
	Name    string
	PathFor func() (string, error)
	Detect  func() (bool, error)
}

var agents = []agentTarget{
	{
		Name: "claude",
		PathFor: func() (string, error) {
			home, err := homeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, ".claude", "skills", "notes", "SKILL.md"), nil
		},
		Detect: func() (bool, error) {
			home, err := homeDir()
			if err != nil {
				return false, err
			}
			_, err = os.Stat(filepath.Join(home, ".claude", "skills"))
			if err == nil {
				return true, nil
			}
			if errors.Is(err, os.ErrNotExist) {
				return false, nil
			}
			return false, err
		},
	},
}

// installAction is a single planned filesystem operation against one
// agent target. See specs/001-generate-skill/data-model.md.
type installAction struct {
	Agent  string
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
	skillAgent   string
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

With --install, the skill is written to a known per-agent location
instead of stdout. Supported agents are listed below.

Supported --agent values:
  claude    ~/.claude/skills/notes/SKILL.md

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

	targets, err := resolveTargets()
	if err != nil {
		return err
	}

	actions := planInstall(skillContent, targets)
	if err := applyInstall(actions, skillDryRun); err != nil {
		return err
	}
	printActions(cmd.OutOrStdout(), actions, skillDryRun)
	return exitErrorFor(actions)
}

func validateSkillFlags() error {
	if !skillInstall {
		switch {
		case skillAgent != "":
			return errors.New("--agent requires --install")
		case skillForce:
			return errors.New("--force requires --install")
		case skillDryRun:
			return errors.New("--dry-run requires --install")
		}
	}
	if skillAgent != "" && findAgent(skillAgent) == nil {
		return fmt.Errorf("unknown agent %q (supported: %s)", skillAgent, agentNamesList())
	}
	return nil
}

func findAgent(name string) *agentTarget {
	for i := range agents {
		if agents[i].Name == name {
			return &agents[i]
		}
	}
	return nil
}

func agentNamesList() string {
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// resolveTargets returns the agent targets to act on for the current
// invocation: either the single agent named by --agent, or every detected
// agent when --agent is empty.
func resolveTargets() ([]agentTarget, error) {
	if skillAgent != "" {
		return []agentTarget{*findAgent(skillAgent)}, nil
	}
	var detected []agentTarget
	for _, a := range agents {
		ok, err := a.Detect()
		if err != nil {
			return nil, err
		}
		if ok {
			detected = append(detected, a)
		}
	}
	if len(detected) == 0 {
		return nil, fmt.Errorf("no supported agent detected; pass --agent explicitly (supported: %s)", agentNamesList())
	}
	return detected, nil
}

// planInstall computes the action for each target without writing
// anything. An OS error while reading an existing destination is
// captured in the action's Error field rather than aborting the whole
// plan.
func planInstall(content string, targets []agentTarget) []installAction {
	bytesContent := []byte(content)
	actions := make([]installAction, len(targets))
	for i, t := range targets {
		actions[i] = planOne(t, bytesContent)
	}
	return actions
}

func planOne(t agentTarget, content []byte) installAction {
	path, err := t.PathFor()
	if err != nil {
		return installAction{Agent: t.Name, Error: err}
	}
	existing, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return installAction{Agent: t.Name, Path: path, Action: actionCreate}
	case err != nil:
		return installAction{Agent: t.Name, Path: path, Error: err}
	}
	if string(existing) == string(content) {
		return installAction{Agent: t.Name, Path: path, Action: actionSkip}
	}
	if skillForce {
		return installAction{Agent: t.Name, Path: path, Action: actionOverwrite}
	}
	return installAction{Agent: t.Name, Path: path, Action: actionConflict}
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

// writeSkillFile creates the parent directory if its parent already
// exists (e.g. creates ~/.claude/skills/notes/ when ~/.claude/skills/
// exists), but refuses to materialise an absent agent skills directory.
// That absence is the detection signal that the user does not run this
// agent.
func writeSkillFile(path string, content []byte) error {
	parent := filepath.Dir(path)
	if _, err := os.Stat(parent); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		grandparent := filepath.Dir(parent)
		if _, gerr := os.Stat(grandparent); gerr != nil {
			return fmt.Errorf("agent skills directory not found: %s", grandparent)
		}
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, content, 0o644)
}

func printActions(out io.Writer, actions []installAction, dryRun bool) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, a := range actions {
		if a.Error != nil {
			fmt.Fprintf(w, "error\t%s\t%s\n", a.Agent, a.Error)
			continue
		}
		verb := a.Action
		if dryRun {
			verb = "would " + verb
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", verb, a.Agent, a.Path)
	}
	_ = w.Flush()
}

func exitErrorFor(actions []installAction) error {
	var problems []string
	for _, a := range actions {
		switch {
		case a.Error != nil:
			problems = append(problems, fmt.Sprintf("%s: %s", a.Agent, a.Error))
		case a.Action == actionConflict:
			problems = append(problems, fmt.Sprintf("%s: destination exists with different content; pass --force to overwrite", a.Agent))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return errors.New(strings.Join(problems, "; "))
}

func registerSkillFlags() {
	skillCmd.Flags().BoolVar(&skillInstall, "install", false, "install the skill into one or more agent locations")
	skillCmd.Flags().StringVar(&skillAgent, "agent", "", "install only into the named agent (default: auto-detect)")
	skillCmd.Flags().BoolVar(&skillForce, "force", false, "overwrite an existing destination with diverging content")
	skillCmd.Flags().BoolVarP(&skillDryRun, "dry-run", "n", false, "print planned actions but do not write any files")
}

func init() {
	registerSkillFlags()
	rootCmd.AddCommand(skillCmd)
}
