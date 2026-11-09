---
name: context-pager
description: >
  Surgical on-demand code paging, OOP boundary refactoring, and single Task Card
  orchestrator. Eliminates LLM attention saturation by indexing codebases into
  lightweight symbol pointer tables, maintaining a library check-out ledger
  (TASK_CARD.md), enforcing pure OOP separation, and halting on 3 consecutive test
  failures for human confirmation.
argument-hint: "[target-path]"
license: MIT
---

# Context Pager

A systems-native approach to AI-assisted coding and large-scale refactoring (>10k lines).
Instead of dumping entire codebases or full files into the context window, this skill
enforces **Demand-Paged Code Navigation** inspired by OS Virtual Memory and Library Card Catalogs.

## The Core Invariants

1. **The best context is the context never read.** (Keep working context $O(1)$)
2. **Pure OOP Principle:** Clear boundary isolation (แยกส่วนให้ชัดเจน) + Explicit communication (คุยกันผ่านสัญญา/Interface).
3. **Single Task Card Ledger:** Exactly ONE living card (`TASK_CARD.md`) tracks all active work, call numbers, and shelf states.
4. **3-Hit Circuit Breaker:** If tests or compilation fail > 3 consecutive runs, STOP immediately. Never guess or churn. Request human guidance.

---

## The 4-Stage Refactoring Pipeline

```text
[Task / Issue]
     │
     ▼
[Stage 1: Catalog & Task Card]   ➔ Map symbols, assign Call Numbers, create TASK_CARD.md
     │
     ▼
[Stage 2: Check-Out & Pure OOP]  ➔ Check out 1 card (Lending mode), isolate boundaries
     │
     ▼
[Stage 3: Minimal Diff]          ➔ Shortest working change (/ponytail) on innermost slice
     │
     ▼
[Stage 4: Verify & Check-In]     ➔ 3x hermetic tests. On pass: Shelve 🟢. On fail > 3: HALT 🔴.
```

---

## Stage 1: Pointer Mapping & The Task Card

1. **NEVER sweep full files.** Strictly forbid reading entire files exceeding 80 lines.
2. **Index the Target:** Run the pager tool to extract symbols, boundaries, and signatures:
   ```bash
   go run main.go index <target-file-or-dir>
   ```
3. **Initialize the Single Living Task Card (`TASK_CARD.md`):**
   Group symbols into domain Call Numbers (e.g., `NET.CONN.01`, `GAME.STATE.02`).
   Place `TASK_CARD.md` at the project root using this canonical template:

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

## Stage 2: Check-Out & Pure OOP Isolation

1. **Borrow Exactly One Piece (Check-Out):**
   Change status in `TASK_CARD.md` from `⚪ [IN SHELF]` to `🟡 [CHECKED OUT]`.
   The agent is strictly prohibited from touching any other piece while a card is checked out.
2. **Pure OOP Boundary Enforcement:**
   * **Separation (แยกส่วน):** Ensure the checked-out symbol has a distinct responsibility. Do not let external state bleed in.
   * **Communication (การสื่อสาร):** Inter-module interactions must go through explicit parameter contracts or interfaces (Caller-defined interfaces from `/engineering-rules`).
3. **Dereference Only the Slice:**
   Fetch ONLY the exact lines belonging to the checked-out symbol:
   ```bash
   go run main.go slice <path> <start_line> <end_line>
   ```

---

## Stage 3: Minimal Diff (`/ponytail`)

Apply senior developer laziness:

1. The shortest working diff wins.
2. No unnecessary abstractions, wrappers, or speculative boilerplate.
3. If extracting a helper from a monolithic function, create a clean method on the struct and replace the site with a one-line call.
4. Shortcuts get explicit debt comments: `shortcut: <the limit>, <when to upgrade>`.

---

## Stage 4: Verification, Check-In & The 3-Hit Circuit Breaker

1. **Hermetic Verification (`/engineering-rules`):**
   * Run the test suite at least 3 times.
   * Tests must NOT sleep or depend on wall-clock timing.
   * Compile with warnings as errors.
2. **Success (Check-In):**
   * Update `TASK_CARD.md`: mark item as `🟢 [SHELVED]`.
   * Clear the Active Borrowing Desk.
   * Run Git commit checkpoint immediately.
   * If in **Auto Mode**: proceed to check out the next item.
   * If in **Interactive Mode**: report progress and await human confirmation.
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
| **Reading Code** | `go run main.go slice <path> <start> <end>` | `read_file` on entire 300+ line files |
| **Work Tracking** | Single `TASK_CARD.md` living ledger | Mental tracking or scattered todo notes |
| **Multi-File Edits** | Check-out one symbol at a time | Editing multiple symbols simultaneously |
| **On Failure > 3** | Halt immediately & present breadcrumbs | Silent 4th/5th retry loops and random guessing |
| **Design Mindset** | Pure OOP (Separation + Communication) | Monolithic spaghetti functions |
