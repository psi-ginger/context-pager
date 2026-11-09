---
name: context-pager
description: >
  Surgical on-demand code paging, OOP boundary refactoring, and single Task Card
  orchestrator. Keeps Go tooling primitive (bounded LLM context) while delegating
  markdown presentation, language syntax validation, approval gating, and
  verifiable benchmark reporting to the AI agent layer. Enforces pure OOP
  separation, library check-out ledger (TASK_CARD.md), "No Magic" variable
  provenance tracking, and halts on 3 consecutive test failures for human guidance.
argument-hint: "[target-path]"
license: MIT
---

# Context Pager

A systems-native protocol for AI-assisted coding and large-scale refactoring (>10k lines).
Enforces **Demand-Paged Code Navigation** inspired by OS Virtual Memory and Library Card Catalogs.

## The Architectural Division of Labor

* **The Go Engine (`pager`):** Kept strictly primitive and zero-bloat. Bounded context delivery to the LLM ($O(1)$ agent context, $O(K)$ streaming slice memory), fast linear repository indexing ($O(N)$), line slicing, and symbol reference searching.
* **The AI Agent Layer:** Serves as the polymorphic presenter, analyst, and guardian. It renders human-friendly Markdown tables, validates language-specific syntax, enforces the "No Magic" provenance policy, manages the single living `TASK_CARD.md`, requests human approval on risky surgery, and outputs reproducible benchmark reports upon task completion.

---

## The Core Invariants

1. **The best context is the context never read.** (Maintain constant bounded $O(1)$ working context)
2. **Pure OOP Principle:** Strict boundary isolation (independent modules) + Explicit communication (caller-driven interfaces).
3. **Single Task Card Ledger:** Exactly ONE living card (`TASK_CARD.md`) tracks all active work, call numbers, and shelf states.
4. **Careful Chunking & Syntax Awareness:** The AI inspects syntax per language (`.go`, `.py`, `.rs`, `.ts`, `.c`), groups chunks with unhurried care, and never rushes.
5. **No Magic (Variable Provenance Policy):** *"If evidence is insufficient, stop tracing. Unknown is better than assumed."*
6. **Interactive Approval Gate:** Propose plans clearly on `TASK_CARD.md` and await human approval before executing non-trivial surgery.
7. **3-Hit Circuit Breaker:** Stop after 3 consecutive failed verification attempts. Never perform a fourth automated fix without human approval.
8. **Verifiable Benchmark Reporting:** Upon completing any task or refactor, output a structured Markdown benchmark comparing total file volume against actual paged consumption.

---

## The 5-Stage Refactoring Pipeline

```text
[Task / Issue]
     │
     ▼
[Stage 1: Catalog & Markdown Presentation] ➔ Run pager index, format clean MD table for human
     │
     ▼
[Stage 2: Check-Out & No-Magic Slicing]    ➔ Borrow 1 card, trace unverified types (Depth <= 2)
     │
     ▼
[Stage 3: Minimal Diff & Approval Gate]    ➔ Propose minimal diff (/ponytail), ask approval if risky
     │
     ▼
[Stage 4: Verify, Shelve & Circuit Breaker] ➔ Hermetic tests. Pass: Shelve 🟢. 3 Failures: HALT 🔴.
     │
     ▼
[Stage 5: Benchmark & Ledger Closure]      ➔ Output final Markdown benchmark report
```

---

## Stage 1: Pointer Mapping & Markdown Presentation

1. **NEVER sweep full files.** Strictly forbid reading entire files exceeding 80 lines into context.
2. **Index the Target:** Run the pager tool to extract symbols, boundaries, and signatures:
   ```bash
   go run main.go index <target-file-or-dir>
   ```
3. **Render Human-Friendly Markdown Table:**
   Do NOT simply dump raw terminal text. Transform the index output into a polished Markdown table for the user:
   | Kind | Call Number | Symbol | Location | Signature |
   | :--- | :--- | :--- | :--- | :--- |
   | `struct` | `NET.CONN.01` | `BotClient` | `client.go:L19-L83` | `type BotClient struct` |
4. **Initialize Active Working Desk (`TASK_CARD.md`):**
   Place `TASK_CARD.md` at the project root for the active refactoring batch (5–10 symbols maximum, avoiding whole-repo bloat):

```markdown
# 📋 REFACTOR MASTER CARD
- Project: <name>
- Goal: <one-line goal>
- Mode: [Interactive | Auto]

## 📚 Book Catalog (Shelf Ledger)
- ⚪ [IN SHELF] <CALL.NO> (<SymbolName>) - <brief role> @ <file:Lstart..Lend>
- 🟡 [CHECKED OUT] <CALL.NO> (<SymbolName>) - Currently under surgery
- 🟢 [SHELVED] <CALL.NO> (<SymbolName>) - Verified & Tests Passed

## 🎯 Active Borrowing Desk
- Currently Holding: <CALL.NO>
- Scope / Boundary: <lines being touched>
- Variable Evidence (No Magic):
  - `<var_name>`: [VERIFIED (SCHEMA) | VERIFIED (TEST FIXTURE) | INFERRED | UNKNOWN] @ <source_file:line>
- Next Step: <exact planned action>
```

---

## Stage 2: Check-Out & No-Magic Slicing

1. **Borrow Exactly One Piece (Check-Out):**
   Change status in `TASK_CARD.md` from `⚪ [IN SHELF]` to `🟡 [CHECKED OUT]`.
   Hold only one symbol at a time.
2. **Language Syntax Awareness:**
   * Acknowledge language rules (`.go` AST receiver methods, `.py` indentation/decorators, `.rs` borrow/lifetime rules, `.ts` types, `.c`/`.cpp` pointers/classes).
   * Ensure the chunk has clean, isolated boundaries before making changes.
3. **The "No Magic" Variable Provenance Policy:**
   When encountering dynamic/loose types (Python, PHP, Lua, JS, C macros/globals) or unverified parameters:
   * **Depth 1 Default (Max Depth = 2):** Use `go run main.go trace <symbol>` to locate reference call sites. *Note: Trace is a fast regex reference search; always dereference matching lines with `slice` to confirm whether it is an actual caller or comment/log.* If wrapped in an adapter/helper, allow at most ONE additional hop (Depth 2).
   * **Inspect Tests / Fixtures:** Check unit test mocks or test fixtures for concrete input schemas.
   * **Evidence Classification (Recorded in Active Desk):**
     * `VERIFIED (SCHEMA / CONTRACT)` — Explicitly established by source code, DB schema/migration, or formal API contract.
     * `VERIFIED (TEST FIXTURE)` — Confirmed within test mocks or fixtures only (note: production may pass unexercised fields).
     * `INFERRED` — Deduceable from local context, but unverified at root (treat with caution).
     * `UNKNOWN` — Source origin unknown (e.g. untyped dynamic dicts, raw external JSON).
   * **External I/O (Database / API Requests):**
     * *Case A (Schema Exists):* If a DB migration, SQL file, or API contract exists, trace to it as evidence.
     * *Case B (No Schema):* **HALT IMMEDIATELY.** Classify as `UNKNOWN`. Request human confirmation via the Approval Gate. Strictly forbid guessing database columns or API payload shapes!
   * **Golden Rule:** *"If evidence is insufficient, stop tracing. Unknown is better than assumed."*
4. **Dereference the Slice:**
   Fetch ONLY the exact lines belonging to the checked-out symbol:
   ```bash
   go run main.go slice <path> <start_line> <end_line>
   ```

---

## Stage 3: Minimal Diff & Approval Gate

1. **Apply Senior Dev Laziness (`/ponytail`):**
   The shortest working diff wins. Never add speculative abstraction or boilerplate nobody asked for.
2. **Interactive Approval Gate:**
   * If in **Interactive Mode**, holding an `UNKNOWN` variable status, or performing non-trivial extraction (e.g. slicing a 600-line monolith):
     Present the exact plan, variable evidence, and proposed diff on `TASK_CARD.md`.
     Ask: *"Please review the proposed refactoring plan and approve before proceeding."*
     Do NOT touch the disk until approval is granted.
   * If in **Auto Mode**: Proceed with minimal diff directly (unless an `UNKNOWN` variable triggers the gate).

---

## Stage 4: Verification, Shelve & Circuit Breaker

1. **Hermetic Verification (`/engineering-rules`):**
   * Run the test suite hermetically.
   * Tests must NOT sleep or depend on wall-clock timing.
   * Compile with warnings as errors (`-Wall -Werror` for C, `mypy` for Python, `phpstan` for PHP).
2. **Success & Re-Indexing (No Line Drift):**
   * Update `TASK_CARD.md`: mark item as `🟢 [SHELVED]`.
   * Clear the active variable evidence in `Active Borrowing Desk`.
   * **MANDATORY RE-INDEX:** Re-run `go run main.go index <file>` on the modified file to immediately refresh all line offsets, preventing line drift errors on subsequent steps.
   * Create Git commit checkpoint when authorized by user or when working tree is in a verified clean state.
   * If in **Auto Mode**: proceed to the next item.
   * If in **Interactive Mode**: report progress to user.
3. **🔴 3-Hit Circuit Breaker:**
   * **Stop after 3 consecutive failed verification attempts. Never perform a fourth automated fix without human approval.**
   * Run `go run main.go trace <symbol>` to collect caller breadcrumbs.
   * Freeze `TASK_CARD.md` with status `🔴 [CIRCUIT BREAKER TRIGGERED]`.
   * Report the exact failure symptoms, breadcrumbs, and 2-3 ranked hypotheses to the user.
   * Await explicit human review and guidance before proceeding.

---

## Stage 5: Verifiable Benchmark Report

Upon completing the session or all cards on the task list, output a structured Benchmark Report formatted in clean Markdown.

### Report Template

```markdown
### 📊 Context Pager Benchmark Report

| Metric | Measured Value | Baseline (Full Sweep) | Efficiency Gain |
| :--- | :--- | :--- | :--- |
| **Total Lines in Scope** | `<actual_lines_paged>` lines | `<total_lines_in_files>` lines | **`<reduction_percentage>% reduction`** |
| **Estimated Token Ingestion** | `~<paged_tokens>` tokens | `~<full_tokens>` tokens | **`~<token_saved>% tokens saved`** |
| **Tool Invocations** | `<count>` calls (idx: `<N>`, slc: `<N>`, trc: `<N>`) | N/A | Deterministic execution |
| **Variable Provenance** | `<N>` Verified, `0` Unknown | Assumed/Hallucinated | Zero Assumption |
| **Hermetic Tests** | Passed (100%) | Flaky/Unchecked | Verified Correctness |
| **Net Code Diff** | `+<added> / -<removed>` lines | Full file rewrite | Minimal Diff |

> **Key Takeaway:** Solved the objective with `<actual_lines_paged>` lines of working context instead of ingesting `<total_lines_in_files>` lines.
```

---

## Operational Summary

| Action | Allowed | Forbidden |
| :--- | :--- | :--- |
| **Tool Design** | Primitive Go engine (streaming $O(K)$ RAM) | Bloating Go with heavy AST linters for every language |
| **Output Styling** | AI renders Markdown table for human | Raw unreadable terminal dumps |
| **Dynamic Variables** | Trace caller (Depth $\le$ 2) or ask human | Guessing dictionary keys / DB schemas |
| **Variable State** | Tri-State (`VERIFIED`, `INFERRED`, `UNKNOWN`) | Pretending assumptions are facts |
| **Post-Edit State** | Mandatory re-index to prevent line drift | Assuming old line numbers remain static |
| **Risky Surgery** | Propose plan & ask for approval | Silent destructive edits in interactive mode |
| **On Failure > 3** | Stop after 3 consecutive failed verification attempts | Silent 4th/5th retry loops and random guessing |
| **Task Completion** | Output verifiable Markdown benchmark | Ending session without measured proof |
