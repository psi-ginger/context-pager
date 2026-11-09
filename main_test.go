package main

import (
	"encoding/json"
	"fmt"
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

func TestReadSliceRealImplementation(t *testing.T) {
	// 1. Test standard slicing directly through the real readSlice function
	var sb strings.Builder
	for i := 1; i <= 100; i++ {
		sb.WriteString(fmt.Sprintf("Line %d\n", i))
	}

	content := sb.String()
	lines, total, err := readSlice(strings.NewReader(content), 10, 20)
	if err != nil {
		t.Fatalf("readSlice failed: %v", err)
	}

	if total != 100 {
		t.Errorf("Expected total 100 lines, got %d", total)
	}
	if len(lines) != 11 {
		t.Fatalf("Expected 11 sliced lines (10..20), got %d", len(lines))
	}
	if lines[0] != "Line 10" {
		t.Errorf("Expected first line to be 'Line 10', got: %s", lines[0])
	}
	if lines[10] != "Line 20" {
		t.Errorf("Expected last line to be 'Line 20', got: %s", lines[10])
	}

	// 2. Test end boundary exceeding total lines (e.g. lines 90..200 on a 100-line file)
	endBeyondLines, endTotal, err := readSlice(strings.NewReader(content), 90, 200)
	if err != nil {
		t.Fatalf("readSlice with end > total failed: %v", err)
	}
	if endTotal != 100 {
		t.Errorf("Expected total 100, got %d", endTotal)
	}
	if len(endBeyondLines) != 11 { // 90 to 100 inclusive is 11 lines
		t.Fatalf("Expected 11 lines when end > total, got %d", len(endBeyondLines))
	}
	if endBeyondLines[0] != "Line 90" || endBeyondLines[10] != "Line 100" {
		t.Errorf("Mismatch in clamped slice content: first=%s, last=%s", endBeyondLines[0], endBeyondLines[10])
	}

	// 3. Test exact CRLF, leading indentation, and trailing space preservation
	crlfContent := "    Line with leading and trailing spaces    \r\n\tSecond line with tabs\t\r\nThird clean line\n"
	crlfLines, _, err := readSlice(strings.NewReader(crlfContent), 1, 3)
	if err != nil {
		t.Fatalf("readSlice on CRLF failed: %v", err)
	}
	if crlfLines[0] != "    Line with leading and trailing spaces    " {
		t.Errorf("Leading/trailing spaces altered: got %q", crlfLines[0])
	}
	if crlfLines[1] != "\tSecond line with tabs\t" {
		t.Errorf("Leading/trailing tabs altered: got %q", crlfLines[1])
	}
	if crlfLines[2] != "Third clean line" {
		t.Errorf("Clean line altered: got %q", crlfLines[2])
	}

	// 4. Test boundary validation errors directly on readSlice
	_, _, err = readSlice(strings.NewReader(content), 0, 10)
	if err == nil || !strings.Contains(err.Error(), "start line must be >= 1") {
		t.Errorf("Expected start < 1 error, got: %v", err)
	}

	_, _, err = readSlice(strings.NewReader(content), 25, 20)
	if err == nil || !strings.Contains(err.Error(), "cannot be greater than end line") {
		t.Errorf("Expected start > end error, got: %v", err)
	}

	_, _, err = readSlice(strings.NewReader(content), 150, 200)
	if err == nil || !strings.Contains(err.Error(), "exceeds total lines in file") {
		t.Errorf("Expected start > total error, got: %v", err)
	}

	_, _, err = readSlice(strings.NewReader(""), 1, 5)
	if err == nil || !strings.Contains(err.Error(), "file is empty") {
		t.Errorf("Expected empty file error, got: %v", err)
	}
}

// customErrorReader simulates an I/O read failure after emitting initial data.
type customErrorReader struct {
	data   string
	read   bool
	errVal error
}

func (r *customErrorReader) Read(p []byte) (n int, err error) {
	if !r.read {
		r.read = true
		copy(p, r.data)
		return len(r.data), nil
	}
	return 0, r.errVal
}

func TestReadSliceMidStreamReadError(t *testing.T) {
	expectedErr := os.ErrPermission
	mockReader := &customErrorReader{
		data:   "Line 1\nLine 2\n",
		errVal: expectedErr,
	}

	_, _, err := readSlice(mockReader, 1, 5)
	if err == nil {
		t.Fatalf("Expected read error, got nil")
	}
	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("Expected error to contain %v, got: %v", expectedErr, err)
	}
}

func BenchmarkReadSliceScaling(b *testing.B) {
	sizes := []struct {
		name  string
		count int
	}{
		{"1K_Lines", 1000},
		{"10K_Lines", 10000},
		{"100K_Lines", 100000},
	}

	for _, s := range sizes {
		var sb strings.Builder
		for i := 1; i <= s.count; i++ {
			sb.WriteString("func Worker() { int count = 1; }\n")
		}
		payload := sb.String()

		b.Run(s.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				lines, total, err := readSlice(strings.NewReader(payload), 500, 530)
				if err != nil || total != s.count || len(lines) != 31 {
					b.Fatalf("benchmark failed: %v", err)
				}
			}
		})
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
