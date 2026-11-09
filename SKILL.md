---
name: context-pager
description: >
  Surgical on-demand code paging, OOP boundary refactoring, and single Task Card
  orchestrator. Keeps Go tooling primitive (O(1) complexity) while delegating
  markdown presentation, language syntax validation, and approval gating to the
  AI agent layer. Enforces pure OOP separation, library check-out ledger
  (TASK_CARD.md), and halts on 3 consecutive test failures for human guidance.
argument-hint: "[target-path]"
license: MIT
---

# Context Pager

A systems-native approach to AI-assisted coding and large-scale refactoring (>10k lines).
Enforces **Demand-Paged Code Navigation** inspired by OS Virtual Memory and Library Card Catalogs.

## The Architectural Division of Labor

* **The Go Engine (`pager`):** Remains a strictly primitive, zero-bloat Unix tool (O(1) tool complexity). It only indexes symbols, slices exact lines, and traces references.
* **The AI Agent Layer:** Serves as the polymorphic presenter and guardian. It renders human-friendly Markdown tables, validates language-specific syntax, manages the single living `TASK_CARD.md`, and requests approval on risky surgery.

---

## The Core Invariants

1. **The best context is the context never read.** (Keep working context $O(1)$)
2. **Pure OOP Principle:** Clear boundary isolation (แยกส่วนให้ชัดเจน) + Explicit communication (คุยกันผ่านสัญญา/Interface).
3. **Single Task Card Ledger:** Exactly ONE living card (`TASK_CARD.md`) tracks all active work, call numbers, and shelf states.
4. **Careful Chunking & Syntax Awareness:** The AI inspects syntax per language (`.go`, `.py`, `.rs`, `.ts`), groups chunks with unhurried care, and never rushes.
5. **Interactive Approval Gate:** Propose plans clearly on `TASK_CARD.md` and await human approval before performing major surgery.
6. **3-Hit Circuit Breaker:** If tests or compilation fail > 3 consecutive runs, STOP immediately. Never guess or churn. Request human guidance.

---

## The 4-Stage Refactoring Pipeline

```text
[Task / Issue]
     │
     ▼
[Stage 1: Catalog & Markdown Presentation] ➔ Run pager index, format beautiful MD table for human
     │
     ▼
[Stage 2: Check-Out & Syntax-Aware Slice]  ➔ Borrow 1 card (Lending mode), inspect language syntax
     │
     ▼
[Stage 3: Minimal Diff & Approval Gate]    ➔ Propose minimal diff (/ponytail), ask approval if risky
     │
     ▼
[Stage 4: Verify, Shelve & Circuit Breaker] ➔ 3x hermetic tests. Pass: Shelve 🟢. Fail > 3: HALT 🔴.
```

---

## Stage 1: Pointer Mapping & Markdown Presentation

1. **NEVER sweep full files.** Strictly forbid reading entire files exceeding 80 lines.
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
- 🟢 [SHELVED] <CALL.NO> (<SymbolName>) - Verified & 3x Tests Passed

## 🎯 Active Borrowing Desk
- Currently Holding: <CALL.NO>
- Scope / Boundary: <lines being touched>
- Next Step: <exact planned action>
```

---

## Stage 2: Check-Out & Syntax-Aware Slicing

1. **Borrow Exactly One Piece (Check-Out):**
   Change status in `TASK_CARD.md` from `⚪ [IN SHELF]` to `🟡 [CHECKED OUT]`.
   Hold only one symbol at a time.
2. **Language Syntax Awareness:**
   * Before modifying, acknowledge the language rules (`.go` receiver methods, `.py` indentation/decorators, `.rs` borrow/trait rules, `.ts` types).
   * Ensure the chunk has clean, isolated boundaries before making changes.
3. **Dereference the Slice:**
   Fetch ONLY the exact lines belonging to the checked-out symbol:
   ```bash
   go run main.go slice <path> <start_line> <end_line>
   ```

---

## Stage 3: Minimal Diff & Approval Gate

1. **Apply Senior Dev Laziness (`/ponytail`):**
   The shortest working diff wins. Never add speculative abstraction or boilerplate nobody asked for.
2. **Interactive Approval Gate:**
   * If in **Interactive Mode** or performing non-trivial extraction (e.g. slicing a 600-line monolith):
     Present the exact plan and proposed diff in chat / on `TASK_CARD.md`.
     Ask: *"พี่แมวตรวจแผนนี้แล้ว Approve ให้ลุยเลยไหมคะ?"*
     Do NOT touch the disk until approval is granted.
   * If in **Auto Mode**: Proceed with minimal diff directly.

---

## Stage 4: Verification, Shelve & Circuit Breaker

1. **Hermetic Verification (`/engineering-rules`):**
   * Run the test suite at least 3 times.
   * Tests must NOT sleep or depend on wall-clock timing.
   * Compile with warnings as errors.
2. **Success & Re-Indexing (No Line Drift):**
   * Update `TASK_CARD.md`: mark item as `🟢 [SHELVED]`.
   * **MANDATORY RE-INDEX:** Re-run `go run main.go index <file>` on the modified file to immediately refresh all line offsets, preventing line drift errors on subsequent steps.
   * Run Git commit checkpoint immediately.
   * If in **Auto Mode**: proceed to the next item.
   * If in **Interactive Mode**: report progress to user.
3. **🔴 3-Hit Circuit Breaker (Fail > 3 Runs):**
   * If a failure repeats for **3 consecutive attempts**:
     1. **HALT EXECUTION IMMEDIATELY.** Do not attempt a 4th automated fix.
     2. Run `go run main.go trace <symbol>` to collect caller breadcrumbs.
     3. Freeze `TASK_CARD.md` with status `🔴 [CIRCUIT BREAKER TRIGGERED]`.
     4. Report the exact failure symptoms, breadcrumbs, and 2-3 ranked hypotheses to the user.
     5. Await explicit human review and approval before proceeding.

---

## Operational Summary

| Action | Allowed | Forbidden |
| :--- | :--- | :--- |
| **Tool Design** | Keep Go engine simple ($O(1)$) | Bloating Go with Markdown/Linter dependencies |
| **Output Styling** | AI renders Markdown table for human | Raw unreadable terminal dumps |
| **Language Handling** | AI validates syntax rules per extension | Blind text replacement without language context |
| **Post-Edit State** | Mandatory re-index to prevent line drift | Assuming old line numbers remain static |
| **Risky Surgery** | Propose plan & ask for approval | Silent destructive edits in interactive mode |
| **On Failure > 3** | Halt immediately & present breadcrumbs | Silent 4th/5th retry loops and random guessing |
