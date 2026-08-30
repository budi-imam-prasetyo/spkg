package output

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/ryoukaii/spkg/internal/detector"
	"gopkg.in/yaml.v3"
)

var (
	checkMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("✓")
	crossMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("✗")
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	runtimeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
)

type Format string

const (
	FormatPretty Format = "pretty"
	FormatJSON   Format = "json"
	FormatYAML   Format = "yaml"
	FormatTable  Format = "table"
	FormatPlain  Format = "plain"
)

func ParseFormat(value string) (Format, bool) {
	switch strings.ToLower(value) {
	case "table":
		return FormatTable, true
	case "json":
		return FormatJSON, true
	case "yaml":
		return FormatYAML, true
	case "plain":
		return FormatPlain, true
	case "pretty":
		return FormatPretty, true
	default:
		return FormatTable, false
	}
}

// printConfidence renders the confidence score, its band label, and the
// contributing reasons. Low/unknown scores are surfaced with a hint so users
// know the result is a best guess.
func printConfidence(r *detector.Result) {
	if r.Confidence <= 0 && len(r.Reasons) == 0 {
		return
	}
	band := detector.BandFor(r.Confidence)
	label := detector.LabelFor(r.Confidence)
	fmt.Println(labelStyle.Render("Confidence:"))
	// Color the score by band for quick visual triage.
	var scoreStyle lipgloss.Style
	switch band {
	case detector.LevelCertain, detector.LevelHigh:
		scoreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	case detector.LevelMedium:
		scoreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	default:
		scoreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	}
	fmt.Printf("  %s (%s)\n", scoreStyle.Render(fmt.Sprintf("%d%%", r.Confidence)), label)
	if len(r.Reasons) > 0 {
		fmt.Println(labelStyle.Render("Reasons:"))
		for _, reason := range r.Reasons {
			fmt.Printf("  • %s\n", reason)
		}
	}
	fmt.Println()
}

func PrintResult(r *detector.Result, format Format, verbose bool) {
	switch format {
	case FormatPretty:
		printPretty(r)
	case FormatJSON:
		printJSON(r)
	case FormatYAML:
		printYAML(r)
	case FormatTable:
		printTableResult(r)
	case FormatPlain:
		printPlainResult(r)
	default:
		printPretty(r)
	}
}

func PrintResults(results []*detector.Result, format Format) {
	switch format {
	case FormatPretty, FormatTable:
		printTable(results)
	case FormatJSON:
		printJSONList(results)
	case FormatYAML:
		printYAMLList(results)
	case FormatPlain:
		printPlainTable(results)
	default:
		printTable(results)
	}
}

func PrintNotFound(name string, format Format) {
	switch format {
	case FormatJSON:
		r := map[string]interface{}{
			"found": false,
			"query": name,
			"error": "executable not found",
		}
		out, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(out))
	case FormatYAML:
		r := map[string]interface{}{
			"found": false,
			"query": name,
			"error": "executable not found",
		}
		out, _ := yaml.Marshal(r)
		fmt.Print(string(out))
	default:
		fmt.Printf("%s %s\n", crossMark, name+" command not found")
	}
}

func printPretty(r *detector.Result) {
	if !r.Found {
		fmt.Printf("%s %s\n", crossMark, "executable not found")
		return
	}

	fmt.Printf("%s Package detected\n\n", checkMark)

	// Executable
	fmt.Println(labelStyle.Render("Executable:"))
	fmt.Printf("  %s\n\n", r.ExecutablePath)

	// Package Manager
	fmt.Println(labelStyle.Render("Package manager:"))
	fmt.Printf("  %s\n\n", string(r.Manager))

	// Runtime (if detected)
	if r.Runtime != "" {
		fmt.Println(runtimeLabel.Render("Runtime:"))
		fmt.Printf("  %s\n", string(r.Runtime))
		if r.RuntimeVersion != "" {
			fmt.Printf("  Runtime version: %s\n", r.RuntimeVersion)
		}
		if r.RuntimeManager != "" {
			fmt.Printf("  Runtime manager: %s\n", r.RuntimeManager)
		}
		fmt.Println()
	}

	// Package details
	if r.Package != "" {
		fmt.Println(labelStyle.Render("Package:"))
		fmt.Printf("  %s\n\n", r.Package)
	}
	if r.Repository != "" {
		fmt.Println(labelStyle.Render("Repository:"))
		fmt.Printf("  %s\n\n", r.Repository)
	}
	if r.IsAUR {
		fmt.Println(labelStyle.Render("AUR:"))
		fmt.Printf("  yes\n\n")
	}
	if r.Version != "" {
		fmt.Println(labelStyle.Render("Installed version:"))
		fmt.Printf("  %s\n\n", r.Version)
	}
	if r.Global != nil && *r.Global {
		fmt.Println(labelStyle.Render("Global:"))
		fmt.Printf("  yes\n\n")
	}
	if r.PythonVersion != "" {
		fmt.Println(labelStyle.Render("Python version:"))
		fmt.Printf("  %s\n\n", r.PythonVersion)
	}
	if r.VirtualEnv != "" {
		fmt.Println(labelStyle.Render("Virtual env:"))
		fmt.Printf("  %s\n\n", r.VirtualEnv)
	}
	if r.Remote != "" {
		fmt.Println(labelStyle.Render("Remote:"))
		fmt.Printf("  %s\n\n", r.Remote)
	}

	// Confidence
	printConfidence(r)

	// Commands
	fmt.Println(labelStyle.Render("Install command:"))
	fmt.Printf("  %s\n\n", r.InstallCmd)
	fmt.Println(labelStyle.Render("Upgrade command:"))
	fmt.Printf("  %s\n\n", r.UpgradeCmd)
	fmt.Println(labelStyle.Render("Remove command:"))
	fmt.Printf("  %s\n", r.RemoveCmd)

	if r.Explain != "" {
		fmt.Println()
		fmt.Println(labelStyle.Render("Explanation:"))
		for _, line := range strings.Split(r.Explain, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}
}

func borderedTable(headers []string, rows [][]string, styleFunc table.StyleFunc) string {
	border := lipgloss.RoundedBorder()
	// Rounded outside corners with clear intersections inside the grid.
	border.MiddleLeft = "├"
	border.MiddleRight = "┤"
	border.Middle = "┼"
	border.MiddleTop = "┬"
	border.MiddleBottom = "┴"

	return table.New().
		Border(border).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		BorderRow(false).
		BorderColumn(true).
		BorderHeader(true).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(styleFunc).
		String()
}

func printTableResult(r *detector.Result) {
	if !r.Found {
		fmt.Printf("%s\n", crossMark+" executable not found")
		return
	}
	var rows [][]string
	row := func(field, value string) {
		if value != "" {
			rows = append(rows, []string{field, value})
		}
	}
	row("Executable", r.ExecutablePath)
	row("Manager", string(r.Manager))
	row("Runtime", string(r.Runtime))
	row("Runtime version", r.RuntimeVersion)
	row("Runtime manager", r.RuntimeManager)
	row("Package", r.Package)
	row("Repository", r.Repository)
	if r.IsAUR {
		row("AUR", "yes")
	}
	row("Version", r.Version)
	if r.Global != nil && *r.Global {
		row("Global", "yes")
	}
	row("Python", r.PythonVersion)
	row("Virtual env", r.VirtualEnv)
	row("Remote", r.Remote)
	if r.Confidence > 0 || len(r.Reasons) > 0 {
		row("Confidence", fmt.Sprintf("%d%% (%s)", r.Confidence, detector.LabelFor(r.Confidence)))
		row("Reasons", strings.Join(r.Reasons, "; "))
	}
	row("Install", r.InstallCmd)
	row("Upgrade", r.UpgradeCmd)
	row("Remove", r.RemoveCmd)

	fmt.Println(borderedTable([]string{"Field", "Value"}, rows, func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if row == table.HeaderRow {
			return style.Foreground(lipgloss.Color("15")).Bold(true).Align(lipgloss.Center)
		}
		if col == 0 {
			return style.Foreground(lipgloss.Color("12")).Bold(true)
		}
		return style
	}))
}

func printPlainResult(r *detector.Result) {
	if !r.Found {
		fmt.Println("not found")
		return
	}
	fmt.Printf("executable: %s\n", r.ExecutablePath)
	fmt.Printf("manager: %s\n", string(r.Manager))
	if r.Runtime != "" {
		fmt.Printf("runtime: %s\n", string(r.Runtime))
		if r.RuntimeVersion != "" {
			fmt.Printf("runtime_version: %s\n", r.RuntimeVersion)
		}
		if r.RuntimeManager != "" {
			fmt.Printf("runtime_manager: %s\n", r.RuntimeManager)
		}
	}
	if r.Package != "" {
		fmt.Printf("package: %s\n", r.Package)
	}
	if r.Version != "" {
		fmt.Printf("version: %s\n", r.Version)
	}
	if r.Repository != "" {
		fmt.Printf("repository: %s\n", r.Repository)
	}
	if r.Confidence > 0 || len(r.Reasons) > 0 {
		fmt.Printf("confidence: %d%% (%s)\n", r.Confidence, detector.LabelFor(r.Confidence))
		for _, reason := range r.Reasons {
			fmt.Printf("reason: %s\n", reason)
		}
	}
}

func printJSON(r *data) {
	out, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(out))
}

func printYAML(r *data) {
	out, _ := yaml.Marshal(r)
	fmt.Print(string(out))
}

// DisplayName returns the name the user typed for a result, falling back to
// the basename of the resolved path when Executable was not recorded. This
// avoids exposing symlink-resolved names like npm-cli.js for `npm`.
func DisplayName(r *detector.Result) string {
	if r.Executable != "" {
		return r.Executable
	}
	return filepath.Base(r.ExecutablePath)
}

func printTable(results []*detector.Result) {
	rows := make([][]string, 0, len(results))
	for _, r := range results {
		rows = append(rows, []string{DisplayName(r), string(r.Manager)})
	}
	fmt.Println(borderedTable([]string{"Executable", "Source"}, rows, func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if row == table.HeaderRow {
			return style.Foreground(lipgloss.Color("15")).Bold(true).Align(lipgloss.Center)
		}
		if col == 1 {
			return style.Foreground(lipgloss.Color("12")).Bold(true)
		}
		return style
	}))
}

func printPlainTable(results []*detector.Result) {
	fmt.Printf("%-20s %-15s %s\n", "Executable", "Manager", "Runtime")
	fmt.Printf("%-20s %-15s %s\n", strings.Repeat("─", 20), strings.Repeat("─", 15), strings.Repeat("─", 15))
	for _, r := range results {
		name := DisplayName(r)
		rt := string(r.Runtime)
		if rt == "" {
			rt = "-"
		}
		fmt.Printf("%-20s %-15s %s\n", name, string(r.Manager), rt)
	}
}

func printJSONList(results []*detector.Result) {
	type item struct {
		Executable     string `json:"executable"`
		Manager        string `json:"manager"`
		Runtime        string `json:"runtime,omitempty"`
		RuntimeVersion string `json:"runtime_version,omitempty"`
		Package        string `json:"package,omitempty"`
	}
	var items []item
	for _, r := range results {
		items = append(items, item{
			Executable:     DisplayName(r),
			Manager:        string(r.Manager),
			Runtime:        string(r.Runtime),
			RuntimeVersion: r.RuntimeVersion,
			Package:        r.Package,
		})
	}
	out, _ := json.MarshalIndent(items, "", "  ")
	fmt.Println(string(out))
}

func printYAMLList(results []*detector.Result) {
	type item struct {
		Executable     string `yaml:"executable"`
		Manager        string `yaml:"manager"`
		Runtime        string `yaml:"runtime,omitempty"`
		RuntimeVersion string `yaml:"runtime_version,omitempty"`
		Package        string `yaml:"package,omitempty"`
	}
	var items []item
	for _, r := range results {
		items = append(items, item{
			Executable:     DisplayName(r),
			Manager:        string(r.Manager),
			Runtime:        string(r.Runtime),
			RuntimeVersion: r.RuntimeVersion,
			Package:        r.Package,
		})
	}
	out, _ := yaml.Marshal(items)
	fmt.Print(string(out))
}

type data = detector.Result
