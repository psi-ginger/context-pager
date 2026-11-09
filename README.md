# Context Pager

> **Demand-Paged Code Navigation for Large Language Models**  
> *Stop stuffing entire repositories into context windows. Paging beats brute force.*

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![Zero Dependency](https://img.shields.io/badge/Dependencies-Zero%20(Standard%20Lib)-green.svg)]()

---

## 1. The Problem: The Long-Context Illusion

Modern LLM workflows have fallen into a brute-force trap: **Context Saturation**.

Developers routinely dump 20+ files or tens of thousands of lines of code into a model's context window. While frontier models advertise multi-million token context windows, real-world execution reveals distinct failure modes:
1. **Attention Dilution:** Retrieval accuracy degrades as irrelevant tokens dilute self-attention weights ("Needle in a Haystack" syndrome).
2. **Hallucination & Drift:** Distant code constructs leak into local variable logic, causing subtle bugs and phantom refactoring.
3. **Local AI Impossibility:** Running 7B–8B models on consumer hardware (12GB–16GB VRAM) becomes unusable when KV caches blow past 16k–32k tokens, causing token generation speeds to collapse to single digits.

---

## 2. The Systems Insight: Virtual Memory & Demand Paging

Operating systems solved memory scarcity over fifty years ago. 

An OS does **not** load an entire 50GB binary into RAM to execute a function. Instead, it creates a **Virtual Page Table** with memory pointers and triggers a **Page Fault** to load only the required 4KB page into physical memory when referenced.

```text
Traditional AI Coding:
  [Entire Codebase (100k tokens)] ➔ [Context Window] ➔ [Attention Saturation / Hallucination]

Context Pager:
  [Symbol Pointer Table (~500 tokens)] ➔ [Agent Plans via Signatures]
                                                  │
                                                  ▼
                        [Demand-Paged Slice (~30 lines)] ➔ [Razor-Sharp Minimal Diff]
```

**Context Pager** applies this classic OS principle to LLM context management:
- **Map First:** Build a lightweight, non-intrusive symbol address table (classes, functions, line offsets).
- **Page on Demand:** Slices and loads only the exact line range corresponding to the function being modified.
- **Context Stays Hot:** Working memory remains tiny (< 2,000 tokens), fast, and 100% focused.

---

## 3. Features

- **Demand-Paged Code Navigation:** Context size remains $O(1)$ regardless of codebase scale.
- **Pure OOP Boundary Enforcement:** Separates concerns into isolated parts that communicate strictly through well-defined contracts.
- **Single Task Card (`TASK_CARD.md`):** One living card tracks the whole session using Library Call Numbers (`NET.CONN.01`), Shelf States (`IN SHELF`, `CHECKED OUT`, `SHELVED`), and human-in-the-loop approvals.
- **3-Hit Circuit Breaker:** Halts immediately on 3 consecutive test failures to prevent agent churn and hallucinated guessing.
- **Polyglot Symbol Extraction:** AST-accurate indexing for Go; robust block detection for Python, Rust, JavaScript/TypeScript, and C/C++.
- **Blazing Fast Startup:** Native Go execution via `go run main.go` with minimal latency and automatic build caching.
- **Deterministic 1-Based Slicing:** Extracts exact line ranges with clean boundary metadata for surgical agent diffs.
- **Call-Chain Tracing:** Fast caller/callee reference discovery without loading full files.
- **Agent Skill Ready:** Drop-in compatibility with Antigravity, Claude Code, Cursor, and OpenCode workflows via `SKILL.md`.

---

## 4. Quickstart

Run directly from source (no build step required):

### 1. Map a File or Repository
Generate the compact symbol pointer table:

```bash
go run main.go index src/
```

Output:
```text
KIND     SYMBOL                       LINES        SIGNATURE / FILE
--------------------------------------------------------------------------------
fn       HandlePacket                 L45-L78 (34L) func HandlePacket(conn NetConn, pkt Packet)
         -> @ src/network.go:45
struct   PlayerSession                L10-L40 (31L) type PlayerSession struct
         -> @ src/session.go:10
--------------------------------------------------------------------------------
Total mapped pointers: 2 symbols
```

### 2. Dereference an Exact Slice
Fetch only the lines needed for the current edit:

```bash
go run main.go slice src/network.go 45 78
```

Output:
```text
--- [SLICE] src/network.go (Lines 45..78 of 120) ---
  45 | func HandlePacket(conn NetConn, pkt Packet) {
  46 | 	if !pkt.IsValid() {
  47 | 		return
  ...
--- [END SLICE] (34 lines) ---
```

### 3. Trace Reference Callers
Follow call sites across files without reading full files:

```bash
go run main.go trace HandlePacket src/
```

---

## 5. The Pipeline Workflow

Context Pager is designed to orchestrate alongside clean engineering disciplines:

1. **Stage 1: Pointer Mapping (`context-pager`)**  
   Index symbols and offsets. Never sweep entire files.
2. **Stage 2: Innermost Layer First (`engineering-rules`)**  
   Page in only the deepest core logic or pure struct. Callers define the interface.
3. **Stage 3: Minimal Diff (`ponytail`)**  
   Write the shortest working change. Eliminate unnecessary abstractions or speculative glue.
4. **Stage 4: Hermetic Verification (`debug-mantra`)**  
   Verify with deterministic tests (no sleep timers). Trace failures through pointer call sites.

---

## 6. Companion Ecosystem (Backlinks)

Context Pager does not operate in isolation; it forms the perceptual foundation for a unified systems-first engineering stack:

| Skill | Role in Stack | Core Philosophy |
| :--- | :--- | :--- |
| **[context-pager](.)** | **Perception / Paging** | *The best context is the context never read.* Demand-paged AST slicing over brute-force token stuffing. |
| **[engineering-rules](https://github.com/vtno/skills/tree/main/skills/engineering-rules)** | **Structure & Invariants** | Innermost layers first, caller-defined interfaces, hand-written fakes, zero wall-clock flakiness. |
| **[ponytail](https://github.com/vtno/skills/tree/main/skills/ponytail)** | **Diff Minimization** | *The best code is the code never written.* YAGNI at senior-dev precision with the shortest working diff. |
| **[debug-mantra](https://github.com/vtno/skills/tree/main/skills/debug-mantra)** | **Fail-Path Tracing** | 4-step discipline: reliable reproduction, fail-path tracing via call pointers, hypothesis falsification, and breadcrumb cross-referencing. |

---

## 7. Philosophy

> *"The best code is the code never written.*  
> *The best context is the context never read."*

---

## License

[MIT](LICENSE) © PsiGinger & Aun
