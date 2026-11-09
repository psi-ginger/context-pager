package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestStripStringAndCommentsRustLifetime(t *testing.T) {
	line := "fn process<'a, 'b>(data: &'a str) -> &'b str {"
	stripped := stripStringAndComments(line, ".rs")

	if !strings.Contains(stripped, "{") {
		t.Fatalf("Expected stripped output to contain '{', got: %s", stripped)
	}

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

func TestParsePythonSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	samplePy := filepath.Join(tmpDir, "server.py")
	content := `class Worker:
    def __init__(self):
        pass

async def handle_request(req):
    return 200
`
	os.WriteFile(samplePy, []byte(content), 0644)
	symbols, err := parseGenericSymbols(samplePy)
	if err != nil {
		t.Fatalf("parseGenericSymbols failed: %v", err)
	}

	if len(symbols) < 3 {
		t.Fatalf("Expected at least 3 symbols (class Worker, def __init__, async def handle_request), got %d: %+v", len(symbols), symbols)
	}

	if symbols[0].Name != "Worker" || symbols[0].Kind != "class" {
		t.Errorf("Expected class Worker, got: %+v", symbols[0])
	}
	if symbols[1].Name != "__init__" || symbols[1].Kind != "fn" {
		t.Errorf("Expected fn __init__, got: %+v", symbols[1])
	}
	if symbols[2].Name != "handle_request" || symbols[2].Kind != "fn" {
		t.Errorf("Expected fn handle_request, got: %+v", symbols[2])
	}
}

func TestParseRustSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	sampleRs := filepath.Join(tmpDir, "lib.rs")
	content := `pub struct Engine {
    state: i32,
}

pub trait Runner {
    fn run(&self);
}

pub async fn start_loop() {
    let x = 1;
}
`
	os.WriteFile(sampleRs, []byte(content), 0644)
	symbols, err := parseGenericSymbols(sampleRs)
	if err != nil {
		t.Fatalf("parseGenericSymbols failed: %v", err)
	}

	if len(symbols) < 4 {
		t.Fatalf("Expected at least 4 symbols in Rust file (struct, trait, trait method, async fn), got %d: %+v", len(symbols), symbols)
	}

	if symbols[0].Name != "Engine" || symbols[0].Kind != "struct" {
		t.Errorf("Expected struct Engine, got: %+v", symbols[0])
	}
	if symbols[1].Name != "Runner" || symbols[1].Kind != "trait" {
		t.Errorf("Expected trait Runner, got: %+v", symbols[1])
	}
	if symbols[2].Name != "run" || symbols[2].Kind != "fn" {
		t.Errorf("Expected fn run, got: %+v", symbols[2])
	}
	if symbols[3].Name != "start_loop" || symbols[3].Kind != "fn" {
		t.Errorf("Expected fn start_loop, got: %+v", symbols[3])
	}
}

func TestParseJSTSSymbolsWithArrowFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	sampleTs := filepath.Join(tmpDir, "index.ts")
	content := `export interface User {
    id: string;
}

export class Service {
    run() {}
}

export const processOrder = async (order: any) => {
    return true;
};

function fallbackHandler() {
    return false;
}
`
	os.WriteFile(sampleTs, []byte(content), 0644)
	symbols, err := parseGenericSymbols(sampleTs)
	if err != nil {
		t.Fatalf("parseGenericSymbols failed: %v", err)
	}

	if len(symbols) < 4 {
		t.Fatalf("Expected at least 4 symbols in TS file, got %d: %+v", len(symbols), symbols)
	}

	if symbols[0].Name != "User" || symbols[0].Kind != "interface" {
		t.Errorf("Expected interface User, got: %+v", symbols[0])
	}
	if symbols[1].Name != "Service" || symbols[1].Kind != "class" {
		t.Errorf("Expected class Service, got: %+v", symbols[1])
	}
	if symbols[2].Name != "processOrder" || symbols[2].Kind != "fn" {
		t.Errorf("Expected fn processOrder for arrow function, got: %+v", symbols[2])
	}
	if symbols[3].Name != "fallbackHandler" || symbols[3].Kind != "fn" {
		t.Errorf("Expected fn fallbackHandler, got: %+v", symbols[3])
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
	if symbols[1].Name != "Move" || symbols[1].Kind != "fn" {
		t.Errorf("Expected fn Move, got: %+v", symbols[1])
	}
}

func TestIndexTargetDiagnosticsAndJSONReport(t *testing.T) {
	tmpDir := t.TempDir()
	goodGo := filepath.Join(tmpDir, "good.go")
	os.WriteFile(goodGo, []byte("package main\nfunc Hello() {}\n"), 0644)

	badGo := filepath.Join(tmpDir, "bad.go")
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

	file, err := os.Open(sample)
	if err != nil {
		t.Fatalf("Failed to open sample: %v", err)
	}
	defer file.Close()

	start := 10
	end := 20
	var targetLines []string
	lineNum := 0

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

func TestTraceFile(t *testing.T) {
	tmpDir := t.TempDir()
	sampleFile := filepath.Join(tmpDir, "trace_sample.go")
	content := `package main

import "fmt"

func Worker() {
	ExecuteJob()
}

func Other() {
	// ExecuteJob commented out
	fmt.Println("done")
}
`
	os.WriteFile(sampleFile, []byte(content), 0644)

	re := regexp.MustCompile(`\bExecuteJob\b`)
	matches, err := traceFile(sampleFile, re, tmpDir)
	if err != nil {
		t.Fatalf("traceFile failed: %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches for ExecuteJob, got %d: %+v", len(matches), matches)
	}

	if matches[0].Line != 6 {
		t.Errorf("Expected match at line 6, got: %d", matches[0].Line)
	}
	if matches[1].Line != 10 {
		t.Errorf("Expected match at line 10, got: %d", matches[1].Line)
	}
}
