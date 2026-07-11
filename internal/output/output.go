package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/ryoukaii/spkg/internal/detector"
	"gopkg.in/yaml.v3"
)

var (
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("✓")
	crossMark = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("✗")
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
)

type Format string

const (
	FormatPretty Format = "pretty"
	FormatJSON   Format = "json"
	FormatYAML   Format = "yaml"
	FormatTable  Format = "table"
	FormatPlain  Format = "plain"
)

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
			"found":  false,
			"query":  name,
			"error":  "executable not found",
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
		fmt.Printf("%s %s\n", crossMark, name+" command tidak ditemukan")
	}
}

func printPretty(r *detector.Result) {
	if !r.Found {
		fmt.Printf("%s %s\n", crossMark, "executable tidak ditemukan")
		return
	}

	fmt.Printf("%s Package ditemukan\n\n", checkMark)
	fmt.Println(labelStyle.Render("Executable:"))
	fmt.Printf("  %s\n\n", r.ExecutablePath)
	fmt.Println(labelStyle.Render("Package manager:"))
	fmt.Printf("  %s\n\n", string(r.Manager))
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

func printTableResult(r *detector.Result) {
	if !r.Found {
		fmt.Printf("%s\n", crossMark+" executable tidak ditemukan")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\n", checkMark+" Package ditemukan\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Executable:"), r.ExecutablePath)
	fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Package manager:"), string(r.Manager))
	if r.Package != "" {
		fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Package:"), r.Package)
	}
	if r.Repository != "" {
		fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Repository:"), r.Repository)
	}
	if r.IsAUR {
		fmt.Fprintf(w, "%s\tyes\n", labelStyle.Render("AUR:"))
	}
	if r.Version != "" {
		fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Version:"), r.Version)
	}
	if r.Global != nil && *r.Global {
		fmt.Fprintf(w, "%s\tyes\n", labelStyle.Render("Global:"))
	}
	if r.PythonVersion != "" {
		fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Python:"), r.PythonVersion)
	}
	if r.VirtualEnv != "" {
		fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Virtual env:"), r.VirtualEnv)
	}
	if r.Remote != "" {
		fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Remote:"), r.Remote)
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Install:"), r.InstallCmd)
	fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Upgrade:"), r.UpgradeCmd)
	fmt.Fprintf(w, "%s\t%s\n", labelStyle.Render("Remove:"), r.RemoveCmd)
	w.Flush()
}

func printPlainResult(r *detector.Result) {
	if !r.Found {
		fmt.Println("not found")
		return
	}
	fmt.Printf("executable: %s\n", r.ExecutablePath)
	fmt.Printf("manager: %s\n", string(r.Manager))
	if r.Package != "" {
		fmt.Printf("package: %s\n", r.Package)
	}
	if r.Version != "" {
		fmt.Printf("version: %s\n", r.Version)
	}
	if r.Repository != "" {
		fmt.Printf("repository: %s\n", r.Repository)
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

func printTable(results []*detector.Result) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", titleStyle.Render("Executable"), titleStyle.Render("Manager"))
	fmt.Fprintf(w, "%s\t%s\n", "────────────", "──────────────")
	for _, r := range results {
		name := r.ExecutablePath
		if idx := strings.LastIndex(name, "/"); idx != -1 {
			name = name[idx+1:]
		}
		fmt.Fprintf(w, "%s\t%s\n", name, string(r.Manager))
	}
	w.Flush()
}

func printPlainTable(results []*detector.Result) {
	fmt.Printf("%-20s %s\n", "Executable", "Manager")
	fmt.Printf("%-20s %s\n", strings.Repeat("─", 20), strings.Repeat("─", 14))
	for _, r := range results {
		name := r.ExecutablePath
		if idx := strings.LastIndex(name, "/"); idx != -1 {
			name = name[idx+1:]
		}
		fmt.Printf("%-20s %s\n", name, string(r.Manager))
	}
}

func printJSONList(results []*detector.Result) {
	type item struct {
		Executable string `json:"executable"`
		Manager    string `json:"manager"`
	}
	var items []item
	for _, r := range results {
		items = append(items, item{
			Executable: r.ExecutablePath,
			Manager:    string(r.Manager),
		})
	}
	out, _ := json.MarshalIndent(items, "", "  ")
	fmt.Println(string(out))
}

func printYAMLList(results []*detector.Result) {
	type item struct {
		Executable string `yaml:"executable"`
		Manager    string `yaml:"manager"`
	}
	var items []item
	for _, r := range results {
		items = append(items, item{
			Executable: r.ExecutablePath,
			Manager:    string(r.Manager),
		})
	}
	out, _ := yaml.Marshal(items)
	fmt.Print(string(out))
}

type data = detector.Result
