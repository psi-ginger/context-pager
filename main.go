package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Symbol represents an indexed code block pointer.
type Symbol struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	File      string `json:"file"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	Lines     int    `json:"lines"`
	Signature string `json:"signature"`
}

var (
	// Regex patterns for multi-language symbol extraction (Heuristic)
	patterns = []*regexp.Regexp{
		// Python: def name(...), async def name(...), class Name(...)
		regexp.MustCompile(`^\s*(?:async\s+)?(def)\s+([a-zA-Z0-9_]+)`),
		regexp.MustCompile(`^\s*(class)\s+([a-zA-Z0-9_]+)`),

		// Rust: fn name(...), struct Name, enum Name, trait Name
		regexp.MustCompile(`^\s*(?:pub(?:\([^)]+\))?\s+)?(?:async\s+)?(fn)\s+([a-zA-Z0-9_]+)`),
		regexp.MustCompile(`^\s*(?:pub(?:\([^)]+\))?\s+)?(struct|enum|trait)\s+([a-zA-Z0-9_]+)`),

		// JS/TS: function name, class name, const name = () =>
		regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?(function)\s+([a-zA-Z0-9_]+)`),
		regexp.MustCompile(`^\s*(?:export\s+)?(class|interface)\s+([a-zA-Z0-9_]+)`),
		regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let)\s+([a-zA-Z0-9_]+)\s*=\s*(?:async\s*)?\([^)]*\)\s*=>`),

		// C/C++: class Name, struct Name, type funcName(...)
		regexp.MustCompile(`^\s*(class|struct)\s+([a-zA-Z0-9_]+)\s*(?:;|:[^{]*)?\{?`),
		regexp.MustCompile(`^\s*(?:[a-zA-Z_][a-zA-Z0-9_]*[\s*&]+)+([a-zA-Z_][a-zA-Z0-9_]*)\s*\([^;]*\)\s*(?:const)?\s*\{`),
	}

	excludeDirs = map[string]bool{
		".git":         true,
		"node_modules": true,
		"target":       true,
		"build":        true,
		"dist":         true,
		".venv":        true,
		"__pycache__":  true,
		".idea":        true,
		".vscode":      true,
	}

	supportedExts = map[string]bool{
		".go":   true,
		".py":   true,
		".rs":   true,
		".js":   true,
		".ts":   true,
		".jsx":  true,
		".tsx":  true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".hpp":  true,
	}
)

func parseGoFile(path string) ([]Symbol, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(fileBytes), "\n")

	var symbols []Symbol
	normalizedPath := filepath.ToSlash(path)

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			start := fset.Position(d.Pos()).Line
			end := fset.Position(d.End()).Line
			name := d.Name.Name
			sig := ""
			if start <= len(lines) {
				sig = strings.TrimSpace(lines[start-1])
				sig = strings.TrimSpace(strings.TrimSuffix(sig, "{"))
			}
			symbols = append(symbols, Symbol{
				Kind:      "fn",
				Name:      name,
				File:      normalizedPath,
				Start:     start,
				End:       end,
				Lines:     end - start + 1,
				Signature: sig,
			})
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						start := fset.Position(ts.Pos()).Line
						end := fset.Position(ts.End()).Line
						kind := "type"
						switch ts.Type.(type) {
						case *ast.StructType:
							kind = "struct"
						case *ast.InterfaceType:
							kind = "interface"
						}
						sig := ""
						if start <= len(lines) {
							sig = strings.TrimSpace(lines[start-1])
							sig = strings.TrimSpace(strings.TrimSuffix(sig, "{"))
						}
						symbols = append(symbols, Symbol{
							Kind:      kind,
							Name:      ts.Name.Name,
							File:      normalizedPath,
							Start:     start,
							End:       end,
							Lines:     end - start + 1,
							Signature: sig,
						})
					}
				}
			}
		}
	}
	return symbols, nil
}

func stripStringAndComments(s string, ext string) string {
	// Strip string literals and single line comments for reliable brace counting
	inQuotes := false
	var quoteChar rune
	var buf strings.Builder

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if !inQuotes {
			// Special handling for Rust lifetimes: 'a, 'static, etc. are NOT quote strings
			if ext == ".rs" && r == '\'' {
				if i+1 < len(runes) && (unicode.IsLetter(runes[i+1]) || runes[i+1] == '_') {
					// Check if this is a single character literal like 'a'
					if !(i+2 < len(runes) && runes[i+2] == '\'') {
						buf.WriteRune(r)
						continue
					}
				}
			}

			if r == '"' || r == '`' || r == '\'' {
				inQuotes = true
				quoteChar = r
				continue
			}
			if r == '/' && i+1 < len(runes) && runes[i+1] == '/' {
				break // comment till end of line
			}
			buf.WriteRune(r)
		} else {
			if r == quoteChar && (i == 0 || runes[i-1] != '\\') {
				inQuotes = false
			}
		}
	}
	return buf.String()
}

func parseGenericSymbols(path string) ([]Symbol, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\r\n")
			lines = append(lines, line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	var symbols []Symbol
	normalizedPath := filepath.ToSlash(path)
	ext := strings.ToLower(filepath.Ext(path))

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		lineNum := i + 1

		for _, p := range patterns {
			m := p.FindStringSubmatch(line)
			if len(m) > 1 {
				kind := "fn"
				name := m[1]
				if len(m) > 2 {
					kind = m[1]
					name = m[2]
				}

				endLine := lineNum
				if ext == ".py" {
					baseIndent := len(line) - len(strings.TrimLeft(line, " \t"))
					for j := i + 1; j < len(lines); j++ {
						trimmed := strings.TrimSpace(lines[j])
						if trimmed == "" || strings.HasPrefix(trimmed, "#") {
							continue
						}
						currIndent := len(lines[j]) - len(strings.TrimLeft(lines[j], " \t"))
						if currIndent <= baseIndent {
							break
						}
						endLine = j + 1
					}
				} else {
					openBraces := 0
					started := false
					for j := i; j < len(lines); j++ {
						clean := stripStringAndComments(lines[j], ext)
						openBraces += strings.Count(clean, "{") - strings.Count(clean, "}")
						if strings.Contains(clean, "{") {
							started = true
						}
						if started && openBraces <= 0 {
							endLine = j + 1
							break
						}
					}
				}

				sig := strings.TrimSpace(line)
				if strings.HasSuffix(sig, "{") {
					sig = strings.TrimSpace(strings.TrimSuffix(sig, "{"))
				} else if strings.HasSuffix(sig, ":") {
					sig = strings.TrimSpace(strings.TrimSuffix(sig, ":"))
				}

				symbols = append(symbols, Symbol{
					Kind:      kind,
					Name:      name,
					File:      normalizedPath,
					Start:     lineNum,
					End:       endLine,
					Lines:     endLine - lineNum + 1,
					Signature: sig,
				})
				break
			}
		}
	}
	return symbols, nil
}

type IndexReport struct {
	TotalFiles   int
	IndexedFiles int
	Symbols      []Symbol
	Warnings     []string
}

func indexTarget(targetPath string) (*IndexReport, error) {
	fi, err := os.Stat(targetPath)
	if err != nil {
		return nil, err
	}

	var files []string
	if !fi.IsDir() {
		files = append(files, targetPath)
	} else {
		err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if excludeDirs[info.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExts[ext] {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	report := &IndexReport{
		TotalFiles: len(files),
	}

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		var syms []Symbol
		var parseErr error
		if ext == ".go" {
			syms, parseErr = parseGoFile(f)
		} else {
			syms, parseErr = parseGenericSymbols(f)
		}

		if parseErr != nil {
			report.Warnings = append(report.Warnings, fmt.Sprintf("%s: %v", filepath.ToSlash(f), parseErr))
		} else {
			report.IndexedFiles++
			report.Symbols = append(report.Symbols, syms...)
		}
	}

	sort.Slice(report.Symbols, func(i, j int) bool {
		if report.Symbols[i].File == report.Symbols[j].File {
			return report.Symbols[i].Start < report.Symbols[j].Start
		}
		return report.Symbols[i].File < report.Symbols[j].File
	})

	return report, nil
}

func handleIndex(args []string) {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: pager index [--json] <path>\n")
		os.Exit(1)
	}

	target := fs.Arg(0)
	report, err := indexTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(report.Symbols)
		return
	}

	cwd, _ := os.Getwd()
	fmt.Printf("%-8s %-28s %-12s %s\n", "KIND", "SYMBOL", "LINES", "SIGNATURE / FILE")
	fmt.Println(strings.Repeat("-", 80))
	for _, s := range report.Symbols {
		loc := fmt.Sprintf("L%d-L%d (%dL)", s.Start, s.End, s.Lines)
		rel, err := filepath.Rel(cwd, s.File)
		if err != nil {
			rel = s.File
		}
		rel = filepath.ToSlash(rel)
		fmt.Printf("%-8s %-28s %-12s %s\n", s.Kind, s.Name, loc, s.Signature)
		fmt.Printf("         -> @ %s:%d\n", rel, s.Start)
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Index Summary: %d symbols mapped across %d/%d files\n", len(report.Symbols), report.IndexedFiles, report.TotalFiles)

	if len(report.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarnings (%d file(s) skipped or failed to parse):\n", len(report.Warnings))
		for _, w := range report.Warnings {
			fmt.Fprintf(os.Stderr, "  [!] %s\n", w)
		}
	}
}

func handleSlice(args []string) {
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: pager slice <path> <start_line> <end_line>\n")
		os.Exit(1)
	}

	target := args[0]
	start, err1 := strconv.Atoi(args[1])
	end, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		fmt.Fprintf(os.Stderr, "Error: start and end must be integers\n")
		os.Exit(1)
	}

	if start < 1 {
		fmt.Fprintf(os.Stderr, "Error: start line must be >= 1 (got %d)\n", start)
		os.Exit(1)
	}
	if start > end {
		fmt.Fprintf(os.Stderr, "Error: start line (%d) cannot be greater than end line (%d)\n", start, end)
		os.Exit(1)
	}

	file, err := os.Open(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot open file %s: %v\n", target, err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\r\n")
			lines = append(lines, line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", target, err)
			os.Exit(1)
		}
	}

	total := len(lines)
	if total == 0 {
		fmt.Fprintf(os.Stderr, "Error: file %s is empty\n", target)
		os.Exit(1)
	}
	if start > total {
		fmt.Fprintf(os.Stderr, "Error: start line (%d) exceeds total lines in file (%d)\n", start, total)
		os.Exit(1)
	}
	if end > total {
		end = total
	}

	cwd, _ := os.Getwd()
	rel, _ := filepath.Rel(cwd, target)
	rel = filepath.ToSlash(rel)

	fmt.Printf("--- [SLICE] %s (Lines %d..%d of %d) ---\n", rel, start, end, total)
	for i := start - 1; i < end; i++ {
		fmt.Printf("%4d | %s\n", i+1, lines[i])
	}
	fmt.Printf("--- [END SLICE] (%d lines) ---\n", end-start+1)
}

func handleTrace(args []string) {
	fs := flag.NewFlagSet("trace", flag.ExitOnError)
	limit := fs.Int("limit", 30, "Max trace matches to display")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: pager trace [--limit N] <symbol> [search_root]\n")
		os.Exit(1)
	}

	symbol := fs.Arg(0)
	root := "."
	if fs.NArg() >= 2 {
		root = fs.Arg(1)
	}

	re, err := regexp.Compile(`\b` + regexp.QuoteMeta(symbol) + `\b`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid symbol pattern: %v\n", err)
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	type match struct {
		file string
		line int
		text string
	}
	var matches []match

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() && excludeDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		lineNum := 0
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				lineNum++
				lineClean := strings.TrimRight(line, "\r\n")
				if re.MatchString(lineClean) {
					rel, _ := filepath.Rel(cwd, path)
					matches = append(matches, match{
						file: filepath.ToSlash(rel),
						line: lineNum,
						text: strings.TrimSpace(lineClean),
					})
				}
			}
			if err != nil {
				break
			}
		}
		return nil
	})

	fmt.Printf("--- [TRACE] Referencing '%s' (%d occurrences) ---\n", symbol, len(matches))
	count := len(matches)
	if count > *limit {
		count = *limit
	}
	for i := 0; i < count; i++ {
		fmt.Printf("%s:%d -> %s\n", matches[i].file, matches[i].line, matches[i].text)
	}
	if len(matches) > *limit {
		fmt.Printf("... and %d more matches truncated.\n", len(matches)-*limit)
	}
	fmt.Printf("--- [END TRACE] ---\n")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Context Pager: High-performance on-demand code paging tool for LLMs.")
		fmt.Println("Usage: pager <command> [arguments]")
		fmt.Println("\nCommands:")
		fmt.Println("  index <path> [--json]       Generate compact symbol pointer table")
		fmt.Println("  slice <path> <start> <end>  Read exact line range")
		fmt.Println("  trace <symbol> [root]       Find symbol call sites and references")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "index":
		handleIndex(os.Args[2:])
	case "slice":
		handleSlice(os.Args[2:])
	case "trace":
		handleTrace(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
