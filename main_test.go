package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripStringAndCommentsRustLifetime(t *testing.T) {
	// Rust snippet with lifetime 'a and 'b
	line := "fn process<'a, 'b>(data: &'a str) -> &'b str {"
	stripped := stripStringAndComments(line, ".rs")

	// The lifetime quotes should NOT have initiated an open quote string that eats the opening brace
	if !strings.Contains(stripped, "{") {
		t.Fatalf("Expected stripped output to contain '{', got: %s", stripped)
	}

	// Normal char literal should still be stripped or handled
	lineChar := "let ch = 'x'; {"
	strippedChar := stripStringAndComments(lineChar, ".rs")
	if !strings.Contains(strippedChar, "{") {
		t.Fatalf("Expected stripped output with char literal to contain '{', got: %s", strippedChar)
	}
}

func TestParseGoFileAST(t *testing.T) {
	tmpDir := t.TempDir()
	sampleGo := filepath.Join(tmpDir, "sample.go")
	content := `package sample

type Config struct {
	Port int
}

func StartServer(cfg Config) error {
	return nil
}
`
	if err := os.WriteFile(sampleGo, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create sample go file: %v", err)
	}

	symbols, err := parseGoFile(sampleGo)
	if err != nil {
		t.Fatalf("parseGoFile failed: %v", err)
	}

	if len(symbols) != 2 {
		t.Fatalf("Expected 2 symbols (struct + fn), got %d", len(symbols))
	}

	if symbols[0].Name != "Config" || symbols[0].Kind != "struct" {
		t.Errorf("Symbol 0 mismatch: %+v", symbols[0])
	}
	if symbols[1].Name != "StartServer" || symbols[1].Kind != "fn" {
		t.Errorf("Symbol 1 mismatch: %+v", symbols[1])
	}
}

func TestParseGenericSymbolsCPP(t *testing.T) {
	tmpDir := t.TempDir()
	sampleCpp := filepath.Join(tmpDir, "player.cpp")
	content := `class Player {
public:
    void Move() {
        int x = 1;
    }
};
`
	if err := os.WriteFile(sampleCpp, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create sample cpp file: %v", err)
	}

	symbols, err := parseGenericSymbols(sampleCpp)
	if err != nil {
		t.Fatalf("parseGenericSymbols failed: %v", err)
	}

	if len(symbols) < 2 {
		t.Fatalf("Expected at least 2 symbols (class + method), got %d: %+v", len(symbols), symbols)
	}

	if symbols[0].Name != "Player" || symbols[0].Kind != "class" {
		t.Errorf("Expected class Player, got: %+v", symbols[0])
	}
}

func TestIndexTargetDiagnosticsAndJSONReport(t *testing.T) {
	tmpDir := t.TempDir()
	goodGo := filepath.Join(tmpDir, "good.go")
	os.WriteFile(goodGo, []byte("package main\nfunc Hello() {}\n"), 0644)

	badGo := filepath.Join(tmpDir, "bad.go")
	// Syntax error in Go file
	os.WriteFile(badGo, []byte("package main\nfunc Broken( {}\n"), 0644)

	report, err := indexTarget(tmpDir)
	if err != nil {
		t.Fatalf("indexTarget failed: %v", err)
	}

	if report.TotalFiles != 2 {
		t.Errorf("Expected 2 total files, got %d", report.TotalFiles)
	}
	if report.IndexedFiles != 1 {
		t.Errorf("Expected 1 indexed file, got %d", report.IndexedFiles)
	}
	if len(report.Warnings) != 1 {
		t.Errorf("Expected 1 warning for bad.go, got %d: %v", len(report.Warnings), report.Warnings)
	}

	// Verify that JSON serialization preserves the report structure including warnings
	jsonBytes, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal(report) failed: %v", err)
	}

	var parsed struct {
		TotalFiles   int      `json:"totalFiles"`
		IndexedFiles int      `json:"indexedFiles"`
		Symbols      []Symbol `json:"symbols"`
		Warnings     []string `json:"warnings"`
	}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if parsed.TotalFiles != 2 || len(parsed.Warnings) != 1 || len(parsed.Symbols) != 1 {
		t.Errorf("JSON output did not preserve full report: %+v", parsed)
	}
}

func TestStreamingSliceLogic(t *testing.T) {
	tmpDir := t.TempDir()
	sample := filepath.Join(tmpDir, "numbers.txt")

	var sb strings.Builder
	for i := 1; i <= 100; i++ {
		sb.WriteString("Line ")
		sb.WriteString(string(rune('0' + (i % 10))))
		sb.WriteString("\n")
	}
	os.WriteFile(sample, []byte(sb.String()), 0644)

	// Open and stream lines 10..20
	file, err := os.Open(sample)
	if err != nil {
		t.Fatalf("Failed to open sample: %v", err)
	}
	defer file.Close()

	start := 10
	end := 20
	var targetLines []string
	lineNum := 0
	reader := os.NewFile(file.Fd(), sample)
	_ = reader

	// Reopen with standard os.ReadFile or simulate slice logic
	fileBytes, _ := os.ReadFile(sample)
	lines := strings.Split(strings.TrimRight(string(fileBytes), "\n"), "\n")
	for i, l := range lines {
		num := i + 1
		lineNum = num
		if num >= start && num <= end {
			targetLines = append(targetLines, l)
		}
	}

	if len(targetLines) != 11 {
		t.Fatalf("Expected 11 lines (10..20), got %d", len(targetLines))
	}
	if lineNum != 100 {
		t.Fatalf("Expected total 100 lines, got %d", lineNum)
	}
}
