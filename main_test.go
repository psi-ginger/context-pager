package main

import (
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

func TestIndexTargetDiagnostics(t *testing.T) {
	tmpDir := t.TempDir()
	sampleGo := filepath.Join(tmpDir, "good.go")
	os.WriteFile(sampleGo, []byte("package main\nfunc Hello() {}\n"), 0644)

	report, err := indexTarget(tmpDir)
	if err != nil {
		t.Fatalf("indexTarget failed: %v", err)
	}

	if report.TotalFiles != 1 || report.IndexedFiles != 1 {
		t.Errorf("Expected 1/1 files indexed, got %d/%d", report.IndexedFiles, report.TotalFiles)
	}
	if len(report.Symbols) != 1 || report.Symbols[0].Name != "Hello" {
		t.Errorf("Expected symbol Hello, got %+v", report.Symbols)
	}
	if len(report.Warnings) != 0 {
		t.Errorf("Expected 0 warnings, got %v", report.Warnings)
	}
}
