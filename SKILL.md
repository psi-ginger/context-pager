---
name: context-pager
description: >
  Surgical on-demand code paging and pipeline orchestrator. Eliminates LLM
  attention saturation and context bloat by indexing codebases into lightweight
  symbol pointer tables and slicing only exact lines. Pipelines sequentially
  into innermost-layer execution (/engineering-rules), minimal diffs (/ponytail),
  and fail-path tracing (/debug-mantra).
argument-hint: "[target-path]"
license: MIT
---

# Context Pager

A systems-native approach to AI-assisted coding. Instead of dumping entire codebases
or full files into the context window, this skill enforces **Demand-Paged Code Navigation**
inspired by OS Virtual Memory.

The Golden Invariant: **The best context is the context never read.**

---

## The 4-Stage Pipeline

```text
[Task / Issue]
     │
     ▼
[Stage 1: Pointer Mapping]       ➔ Map symbols & offsets (No full-file reads!)
     │
     ▼
[Stage 2: Innermost Layer First] ➔ Slices innermost logic (/engineering-rules)
     │
     ▼
[Stage 3: Minimal Diff]          ➔ Shortest working change (/ponytail)
     │
     ▼
[Stage 4: Verification / Trace]  ➔ Hermetic 3x test (/debug-mantra on failure)
```

---

## Stage 1: Pointer Mapping (No Sweeping Allowed)

When assigned a coding, refactoring, or debugging task:

1. **NEVER sweep full files.** Strictly forbid reading entire files exceeding 80 lines into context.
2. **Generate the Pointer Map:** Run the pager tool to extract symbols, line boundaries, and signatures:
   ```bash
   go run main.go index <target-file-or-dir>
   ```
3. **Plan using Pointers:** Study only the resulting symbol table:
   * Class / Function names
   * Line boundaries (`L<start>-L<end>`)
   * Signatures / Parameter contracts

---

## Stage 2: Innermost Layer First (`/engineering-rules`)

Before touching any glue code or outer interfaces:

1. Identify the **innermost layer** from the pointer map (the pure computation, core struct, or isolated algorithm).
2. **Dereference the Slice:** Fetch ONLY the exact lines belonging to that symbol:
   ```bash
   go run main.go slice <path> <start_line> <end_line>
   ```
3. Define interfaces strictly by what the caller needs. Do not create speculative abstractions.

---

## Stage 3: Minimal Diff (`/ponytail`)

Apply senior developer laziness:

1. Take the smallest change that fully solves the requirement.
2. No abstractions, wrappers, unnecessary type conversions, or "for later" options.
3. Every shortcut with a known limit gets a comment: `shortcut: <the limit>, <when to upgrade>`.
4. Ensure the shortest working diff wins.

---

## Stage 4: Verification & Fail-Path Tracing

1. **Verify Hermetically (`/engineering-rules`):**
   * Run the test suite at least 3 times.
   * Tests must NOT sleep or depend on wall-clock timing.
   * Compile with warnings as errors.
2. **If an issue or bug arises (`/debug-mantra`):**
   * Do NOT randomly dump adjacent files into context.
   * Trace the fail path using call pointers:
     ```bash
     go run main.go trace <symbol_name> <root_path>
     ```
   * Inspect only the connecting call sites (`file:line`) to reconstruct the execution chain.

---

## Operational Summary

| Action | Allowed | Forbidden |
| :--- | :--- | :--- |
| **Reading Code** | `go run main.go slice <path> <start> <end>` | `read_file` on entire 300+ line files |
| **Exploration** | `go run main.go index <dir>` | Blind recursive file dumps |
| **Tracing Bugs** | `go run main.go trace <symbol>` | Guessing caller signatures without checking |
| **Implementation** | Innermost layer first -> Minimal diff | Writing all layers at once |
