# yx CLI Rewrite: Bash to Go

## TL;DR

> **Quick Summary**: Rewrite the 845-line Bash `yx` CLI tool to idiomatic Go using Cobra, maintaining 100% feature parity with all 103 ShellSpec tests passing. Shell out to git for all git operations.
> 
> **Deliverables**:
> - Go CLI binary at `bin/yx-go` (swap to `bin/yx` after verification)
> - All 9 commands with identical behavior and output
> - Cobra-generated shell completions
> - devenv.nix integration for Go toolchain
> - Go unit tests for internal packages
> 
> **Estimated Effort**: Large (multi-week project)
> **Parallel Execution**: YES - 3 waves after foundation
> **Critical Path**: Setup → Core Types → Simple Commands → Complex Commands → Git Integration → Sync → Completions → Final Verification

---

## Context

### Original Request
Rewrite the `yx` CLI tool from Bash to Go, maintaining 100% feature parity with all 103 ShellSpec tests passing.

### Interview Summary
**Key Discussions**:
- Interactive add mode: Must read stdin line-by-line, empty line terminates
- Editor integration: Shell out to $EDITOR for context editing
- Completions: Generate via Cobra, update install command
- Binary location: `bin/yx-go` initially, swap after tests pass
- Testing: ShellSpec integration + Go unit tests via TDD
- Build setup: Integrate with existing devenv.nix

**Research Findings**:
- Current bash is 845 lines with 103 test cases
- Git operations use temp index, merge-tree, commit-tree, ref manipulation
- Output includes ANSI escape codes (`\e[90m...\e[0m`) that tests verify byte-exact
- Sorting: done first, then by mtime within state groups
- Fuzzy matching with ambiguity detection
- Migration from old `done` file to `state` file format

### Self-Review Gaps Identified
**Addressed in plan**:
- Edge case: Empty YAKS_PATH directory vs non-existent
- Edge case: Multi-word yak names without quotes
- Edge case: Nested yak depth limits
- Cross-platform stat command differences (handled by Go's os.Stat)
- ANSI escape code output must be byte-exact
- Git working directory context for sync operations

---

## Work Objectives

### Core Objective
Create an idiomatic Go CLI that produces identical behavior and output to the existing Bash implementation, verified by passing all 103 ShellSpec tests.

### Concrete Deliverables
- `bin/yx-go` - Go binary
- `cmd/yx/main.go` - Entry point
- `internal/` packages - Modular Go implementation
- `go.mod` and `go.sum` - Module definition
- Updated `devenv.nix` - Go toolchain integration
- Cobra-generated completions replacing manual files

### Definition of Done
- [ ] `shellspec` passes with 103/103 tests when `PATH` includes Go binary
- [ ] `go test ./...` passes all Go unit tests
- [ ] `go build` produces working binary
- [ ] Binary works on macOS and Linux

### Must Have
- Identical CLI interface (command names, flags, arguments)
- Identical output format (including ANSI escape codes)
- Identical error messages (tests check exact strings)
- Identical exit codes
- Git refs/notes/yaks integration
- YAKS_PATH environment variable support
- Cross-platform compatibility (macOS, Linux)

### Must NOT Have (Guardrails)
- **No new features** - strict parity only
- **No output format changes** - byte-exact match required
- **No go-git library** - shell out to git (go-git lacks merge-tree support)
- **No additional dependencies** beyond Cobra and stdlib
- **No Windows support** - explicitly out of scope
- **No changes to test files** - tests are the contract

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES (ShellSpec)
- **User wants tests**: TDD for Go units + ShellSpec for integration
- **Framework**: ShellSpec (existing) + Go testing (new)

### TDD Workflow for Go Units

Each internal package follows RED-GREEN-REFACTOR:

1. **RED**: Write failing Go test first
   - Test file: `internal/<pkg>/<pkg>_test.go`
   - Test command: `go test ./internal/<pkg>/...`
   - Expected: FAIL

2. **GREEN**: Implement minimum code to pass
   - Command: `go test ./internal/<pkg>/...`
   - Expected: PASS

3. **REFACTOR**: Clean up while green
   - Command: `go test ./...`
   - Expected: PASS

### ShellSpec Integration Testing

**Strategy**: Run ShellSpec against Go binary by manipulating PATH

```bash
# Build Go binary
go build -o bin/yx-go ./cmd/yx

# Create symlink or copy to 'yx' for tests
ln -sf yx-go bin/yx

# Run tests with Go binary (yx will resolve to yx-go via symlink)
PATH="$(pwd)/bin:$PATH" shellspec
```

**Incremental Verification**: After each command implementation:
```bash
# Run specific test file
shellspec spec/add.sh        # After implementing add
shellspec spec/list.sh       # After implementing list
shellspec spec/done.sh       # After implementing done
# ... and so on
```

### Manual Verification Procedures

**For each command, verify:**
- [ ] Command exists: `./bin/yx-go --help`
- [ ] Output matches bash: `diff <(./bin/yx <args>) <(./bin/yx-go <args>)`
- [ ] Error output matches: `diff <(./bin/yx bad 2>&1) <(./bin/yx-go bad 2>&1)`

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 0 (Foundation - Sequential):
└── Task 1: Project setup (devenv, go.mod, structure)

Wave 1 (Core - Sequential):
├── Task 2: Core types and storage
└── Task 3: Fuzzy matching

Wave 2 (Simple Commands - Parallel):
├── Task 4: add command
├── Task 5: list command  
├── Task 6: rm command
└── Task 7: context command

Wave 3 (Stateful Commands - Parallel after Wave 2):
├── Task 8: done command
├── Task 9: move command
└── Task 10: prune command

Wave 4 (Git Integration - Sequential):
├── Task 11: Git logging (log_command)
└── Task 12: sync command

Wave 5 (Completions & Polish - Sequential):
├── Task 13: Cobra completions
├── Task 14: completions install command
└── Task 15: Final integration & swap

Critical Path: 1 → 2 → 3 → 4 → 8 → 11 → 12 → 15
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 | None | 2-15 | None |
| 2 | 1 | 3-15 | None |
| 3 | 2 | 4-15 | None |
| 4 | 3 | 8, 11 | 5, 6, 7 |
| 5 | 3 | 8, 10 | 4, 6, 7 |
| 6 | 3 | 10 | 4, 5, 7 |
| 7 | 3 | None | 4, 5, 6 |
| 8 | 4, 5 | 10, 11 | 9 |
| 9 | 4 | None | 8, 10 |
| 10 | 5, 6, 8 | None | 9 |
| 11 | 4, 8 | 12 | None |
| 12 | 11 | 15 | None |
| 13 | 1 | 14 | 4-12 |
| 14 | 13 | 15 | None |
| 15 | 12, 14 | None | None |

---

## TODOs

### Task 1: Project Setup and devenv.nix Integration

- [x] 1. Project Setup and devenv.nix Integration

  **What to do**:
  - Add Go to devenv.nix with latest stable version (1.22+)
  - Create Go module: `go mod init github.com/mattwynne/yaks`
  - Add Cobra dependency: `go get github.com/spf13/cobra@latest`
  - Create directory structure:
    ```
    cmd/yx/main.go
    internal/yak/
    internal/git/
    internal/display/
    internal/cmd/
    ```
  - Create minimal main.go with root command
  - Verify: `go build -o bin/yx-go ./cmd/yx && ./bin/yx-go --help`

  **Must NOT do**:
  - Add unnecessary dependencies
  - Modify existing bash implementation
  - Change any test files

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Setup task with well-defined steps, no complex logic
  - **Skills**: [`shellspec`]
    - `shellspec`: Project uses ShellSpec for testing, need awareness

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 0 (foundation)
  - **Blocks**: All other tasks
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `devenv.nix:1-50` - Existing devenv configuration to extend
  - `bin/yx:1-10` - Entry point pattern to replicate

  **Documentation References**:
  - Cobra docs: https://github.com/spf13/cobra - CLI framework

  **Acceptance Criteria**:

  - [ ] `devenv.nix` updated with Go language
  - [ ] `go mod init` creates go.mod
  - [ ] `go build -o bin/yx-go ./cmd/yx` succeeds
  - [ ] `./bin/yx-go --help` shows help (can be placeholder)
  - [ ] Directory structure exists as specified

  **Manual Verification**:
  ```bash
  # Verify Go in devenv
  which go  # Should show devenv path
  go version  # Should show 1.22+
  
  # Verify build
  go build -o bin/yx-go ./cmd/yx
  ./bin/yx-go --help
  ```

  **Commit**: YES
  - Message: `feat(go): initialize Go project structure with Cobra CLI`
  - Files: `devenv.nix`, `go.mod`, `go.sum`, `cmd/yx/main.go`, `internal/`
  - Pre-commit: `go build ./...`

---

### Task 2: Core Yak Types and File Storage

- [x] 2. Core Yak Types and File Storage

  **What to do**:
  - Create `internal/yak/yak.go`:
    - `Yak` struct: Name, State (todo/done), Children, ContextPath
    - `State` type: `todo` or `done`
  - Create `internal/yak/store.go`:
    - `Store` struct with `BasePath` (YAKS_PATH)
    - `NewStore(path string)` - create store, handle absolute path conversion
    - `Create(name string)` - create yak dir, state file, context.md
    - `Get(name string)` - load yak from disk
    - `List()` - find all yaks recursively
    - `Delete(name string)` - remove yak directory
    - `Exists(name string)` - check if yak exists
    - `ValidateName(name string)` - check forbidden chars
  - Handle nested yaks: `parent/child` → `YAKS_PATH/parent/child/`
  - Migration: convert old `done` files to `state` files

  **Must NOT do**:
  - Add git operations (separate task)
  - Add display logic (separate task)
  - Change file format from bash version

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Core data structures, needs careful design for extensibility
  - **Skills**: [`incremental-tdd`]
    - `incremental-tdd`: TDD approach for reliable core types

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (core - sequential)
  - **Blocks**: Tasks 3-15
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `bin/yx:64-71` - `validate_yak_name()` - forbidden character validation
  - `bin/yx:73-78` - `find_all_yaks()` - recursive directory finding
  - `bin/yx:80-88` - `is_yak_done()` - state file reading
  - `bin/yx:355-362` - `add_yak_single()` - file creation pattern
  - `bin/yx:15-30` - `migrate_done_to_state()` - migration logic
  - `bin/yx:5-13` - `convert_to_absolute_path()` - path handling

  **Test References**:
  - `spec/add.sh:48-94` - Validation test cases (forbidden chars)
  - `spec/done.sh:75-87` - Migration test case

  **Acceptance Criteria**:

  - [ ] Go test: `go test ./internal/yak/...` passes
  - [ ] Store creates `.yaks/<name>/state` with "todo"
  - [ ] Store creates `.yaks/<name>/context.md` (empty)
  - [ ] Nested yaks: `parent/child` creates correct structure
  - [ ] Validation rejects: `\ : * ? | < > "`
  - [ ] Validation accepts: `/` (for nesting)
  - [ ] Migration converts `done` file to `state` file

  **Manual Verification**:
  ```bash
  # Go REPL equivalent
  go test -v ./internal/yak/... -run TestValidation
  go test -v ./internal/yak/... -run TestCreate
  go test -v ./internal/yak/... -run TestMigration
  ```

  **Commit**: YES
  - Message: `feat(yak): implement core yak types and file storage`
  - Files: `internal/yak/yak.go`, `internal/yak/store.go`, `internal/yak/*_test.go`
  - Pre-commit: `go test ./internal/yak/...`

---

### Task 3: Fuzzy Matching Implementation

- [x] 3. Fuzzy Matching Implementation

  **What to do**:
  - Create `internal/yak/fuzzy.go`:
    - `FindYak(store *Store, searchTerm string) (string, error)`
    - Try exact match first: `store.Exists(searchTerm)`
    - Try fuzzy match: substring matching across all yak names
    - Return error if no match found
    - Return error if multiple matches (ambiguous)
  - Create `internal/yak/errors.go`:
    - `ErrYakNotFound` - "Error: yak '<name>' not found"
    - `ErrAmbiguousMatch` - "Error: yak name '<name>' is ambiguous"
  - Error messages must match bash exactly (tests verify strings)

  **Must NOT do**:
  - Implement advanced fuzzy matching (simple substring only)
  - Change error message format

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Focused algorithm implementation with clear spec
  - **Skills**: [`incremental-tdd`]
    - `incremental-tdd`: Error handling needs precise testing

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1 (core - after Task 2)
  - **Blocks**: Tasks 4-12
  - **Blocked By**: Task 2

  **References**:

  **Pattern References**:
  - `bin/yx:90-97` - `try_exact_match()` - exact match logic
  - `bin/yx:99-117` - `try_fuzzy_match()` - substring matching
  - `bin/yx:119-124` - `find_yak()` - combined logic
  - `bin/yx:136-153` - `require_yak()` - error handling wrapper

  **Test References**:
  - `spec/fuzzy_match.sh:1-27` - Fuzzy matching test cases
  - `spec/done.sh:14-18` - "not found" error message test

  **Acceptance Criteria**:

  - [ ] Go test: `go test ./internal/yak/... -run TestFuzzy` passes
  - [ ] Exact match: `"parent/child"` finds `"parent/child"`
  - [ ] Fuzzy match: `"build"` finds `"ideas/fix the build"` (unique)
  - [ ] Ambiguous: `"fix"` with `"fix the build"` and `"fix the fridge"` returns error
  - [ ] Not found: `"nonexistent"` returns "Error: yak 'nonexistent' not found"
  - [ ] Ambiguous error: returns "Error: yak name 'fix' is ambiguous"

  **Manual Verification**:
  ```bash
  go test -v ./internal/yak/... -run TestFuzzy
  ```

  **Commit**: YES
  - Message: `feat(yak): implement fuzzy matching with ambiguity detection`
  - Files: `internal/yak/fuzzy.go`, `internal/yak/errors.go`, `internal/yak/*_test.go`
  - Pre-commit: `go test ./internal/yak/...`

---

### Task 4: Add Command Implementation

- [x] 4. Add Command Implementation

  **What to do**:
  - Create `internal/cmd/add.go`:
    - Command: `yx add <name>` or `yx add` (interactive)
    - Multi-word names without quotes: `yx add this is a test`
    - Interactive mode: read stdin line-by-line, empty line terminates
    - Print "Enter yaks (empty line to finish):" in interactive mode
    - Validate name before creation
  - Wire up in `cmd/yx/main.go`

  **Must NOT do**:
  - Add git logging (Task 11)
  - Change output messages

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: First command, establishes patterns for others
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `incremental-tdd`: TDD for command logic
    - `shellspec`: Run spec/add.sh to verify

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 5, 6, 7)
  - **Blocks**: Tasks 8, 9, 11
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `bin/yx:340-353` - `add_yak_interactive()` - interactive mode
  - `bin/yx:355-362` - `add_yak_single()` - single yak creation
  - `bin/yx:364-370` - `add_yak()` - dispatch logic
  - `bin/yx:808-810` - Command routing for add

  **Test References**:
  - `spec/add.sh:1-95` - All add command tests

  **Acceptance Criteria**:

  - [ ] `shellspec spec/add.sh` passes (11 tests)
  - [ ] `yx add "Fix the bug"` creates yak
  - [ ] `yx add this is a test` creates "this is a test" yak
  - [ ] `yx add "foo/bar"` creates nested yak
  - [ ] `yx add "foo:bar"` fails with "Invalid yak name"
  - [ ] Interactive mode reads stdin until empty line

  **Manual Verification**:
  ```bash
  ./bin/yx-go add "test yak"
  ls -la .yaks/
  # Should show: test yak/
  
  echo -e "yak1\nyak2\n" | ./bin/yx-go add
  # Should prompt and create both
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement add command with interactive mode`
  - Files: `internal/cmd/add.go`, `internal/cmd/add_test.go`
  - Pre-commit: `shellspec spec/add.sh`

---

### Task 5: List Command Implementation

- [ ] 5. List Command Implementation

  **What to do**:
  - Create `internal/display/markdown.go`:
    - `FormatMarkdown(yak *Yak, depth int) string`
    - 2-space indentation per depth level
    - `- [ ] name` for todo, `- [x] name` for done
    - ANSI grey for done: `\033[90m...\033[0m`
  - Create `internal/display/plain.go`:
    - `FormatPlain(yak *Yak) string` - just the full path name
  - Create `internal/cmd/list.go`:
    - Command: `yx list` or `yx ls`
    - Flags: `--format markdown|md|plain|raw`, `--only not-done|done`
    - Sort: done first, then by mtime
    - Empty message: "You have no yaks. Are you done?"
    - Recursive tree display
  - Get mtime using `os.Stat()` (cross-platform)

  **Must NOT do**:
  - Use different ANSI codes
  - Change sort order
  - Change empty message

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Output formatting with ANSI codes, visual hierarchy
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `incremental-tdd`: Display logic needs precise testing
    - `shellspec`: Verify exact output format

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 6, 7)
  - **Blocks**: Tasks 8, 10
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `bin/yx:257-277` - `display_yak_markdown()` - ANSI formatting
  - `bin/yx:279-282` - `display_yak_plain()` - plain output
  - `bin/yx:168-178` - `get_sort_priority()` - sort logic
  - `bin/yx:180-189` - `sort_yaks()` - sorting implementation
  - `bin/yx:211-309` - `list_yaks()` - complete list logic
  - `bin/yx:191-209` - `parse_format_option()`, `parse_only_option()`

  **Test References**:
  - `spec/list.sh:1-157` - All list command tests (15 tests)

  **Acceptance Criteria**:

  - [ ] `shellspec spec/list.sh` passes (15 tests)
  - [ ] `yx ls` shows markdown checkboxes
  - [ ] `yx ls --format plain` shows just names
  - [ ] `yx ls --only not-done` filters correctly
  - [ ] Done items shown in grey ANSI
  - [ ] Nested items indented 2 spaces per level
  - [ ] Empty yaks shows "You have no yaks. Are you done?"
  - [ ] Sort order: done first, then by mtime

  **Manual Verification**:
  ```bash
  ./bin/yx-go add "task1" && sleep 0.1 && ./bin/yx-go add "task2"
  ./bin/yx-go ls
  # Check indentation, format
  
  diff <(./bin/yx ls) <(./bin/yx-go ls)
  # Should be empty (identical)
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement list command with format and filter options`
  - Files: `internal/display/*.go`, `internal/cmd/list.go`, tests
  - Pre-commit: `shellspec spec/list.sh`

---

### Task 6: Remove Command Implementation

- [x] 6. Remove Command Implementation

  **What to do**:
  - Create `internal/cmd/rm.go`:
    - Command: `yx rm <name>`
    - Support multi-word names: `yx rm this is a test`
    - Use fuzzy matching to resolve name
    - Remove entire yak directory (including nested)
    - Error if yak not found

  **Must NOT do**:
  - Add git logging (Task 11)
  - Change error messages

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple command, uses existing store operations
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `shellspec`: Verify spec/rm.sh passes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 7)
  - **Blocks**: Task 10
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `bin/yx:436-443` - `remove_yak()` - removal logic
  - `bin/yx:821-823` - Command routing for rm

  **Test References**:
  - `spec/rm.sh:1-51` - All rm command tests (5 tests)

  **Acceptance Criteria**:

  - [ ] `shellspec spec/rm.sh` passes (5 tests)
  - [ ] `yx rm "Fix the bug"` removes yak
  - [ ] `yx rm this is a test` removes multi-word yak
  - [ ] `yx rm "Nonexistent"` errors with "not found"
  - [ ] Removes nested yaks correctly

  **Manual Verification**:
  ```bash
  ./bin/yx-go add "to delete"
  ./bin/yx-go rm "to delete"
  ./bin/yx-go ls
  # Should not contain "to delete"
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement rm command`
  - Files: `internal/cmd/rm.go`, tests
  - Pre-commit: `shellspec spec/rm.sh`

---

### Task 7: Context Command Implementation

- [x] 7. Context Command Implementation

  **What to do**:
  - Create `internal/cmd/context.go`:
    - `yx context --show <name>` - display yak name + context
    - `yx context [--edit] <name>` - edit context
    - If stdin is not a TTY, read context from stdin
    - If stdin is TTY, launch $EDITOR
    - Output format: `<resolved_name>\n\n<context.md contents>`
  - Use `os.Stdin.Stat()` to detect TTY vs pipe
  - Shell out to $EDITOR using `os.Exec`

  **Must NOT do**:
  - Add git logging (Task 11)
  - Change output format
  - Use embedded editor

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Straightforward I/O operations
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `shellspec`: Verify spec/context.sh passes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 6)
  - **Blocks**: None (independent leaf)
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `bin/yx:495-505` - `show_yak_context()` - show mode
  - `bin/yx:507-518` - `edit_yak_context()` - edit mode with TTY detection
  - `bin/yx:520-529` - `context_yak()` - flag dispatch

  **Test References**:
  - `spec/context.sh:1-72` - All context command tests (7 tests)

  **Acceptance Criteria**:

  - [ ] `shellspec spec/context.sh` passes (7 tests)
  - [ ] `yx context --show "yak"` shows name + context
  - [ ] `echo "text" | yx context "yak"` sets context from stdin
  - [ ] `yx context "yak"` (TTY) launches $EDITOR
  - [ ] Works with nested yaks
  - [ ] Error on nonexistent yak

  **Manual Verification**:
  ```bash
  ./bin/yx-go add "my yak"
  echo "# Context" | ./bin/yx-go context "my yak"
  ./bin/yx-go context --show "my yak"
  # Should show: my yak\n\n# Context
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement context command with editor support`
  - Files: `internal/cmd/context.go`, tests
  - Pre-commit: `shellspec spec/context.sh`

---

### Task 8: Done Command Implementation

- [x] 8. Done Command Implementation

  **What to do**:
  - Create `internal/cmd/done.go`:
    - `yx done <name>` - mark yak as done (write "done" to state)
    - `yx done --undo <name>` - mark as todo
    - `yx done --recursive <name>` - mark yak and all children done
    - Check for incomplete children before marking parent done
    - Error: "Error: cannot mark '<name>' as done - it has incomplete children"
  - Update state file: "done" or "todo"

  **Must NOT do**:
  - Add git logging (Task 11)
  - Allow marking parent done with incomplete children (without --recursive)

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Complex state logic with children checking, recursion
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `incremental-tdd`: State transitions need careful testing
    - `shellspec`: Verify spec/done.sh passes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Task 9)
  - **Blocks**: Tasks 10, 11
  - **Blocked By**: Tasks 4, 5

  **References**:

  **Pattern References**:
  - `bin/yx:372-389` - `has_incomplete_children()` - children check
  - `bin/yx:391-403` - `mark_yak_done_recursively()` - recursive marking
  - `bin/yx:405-434` - `done_yak()` - main done logic with flags

  **Test References**:
  - `spec/done.sh:1-113` - All done command tests (10 tests)

  **Acceptance Criteria**:

  - [ ] `shellspec spec/done.sh` passes (10 tests)
  - [ ] `yx done "yak"` marks as done (state file = "done")
  - [ ] `yx done --undo "yak"` marks as todo
  - [ ] `yx done --recursive "parent"` marks parent + all children
  - [ ] Error when marking parent with incomplete children
  - [ ] List shows done yaks in grey with [x]

  **Manual Verification**:
  ```bash
  ./bin/yx-go add "parent" && ./bin/yx-go add "parent/child"
  ./bin/yx-go done "parent"  # Should error
  ./bin/yx-go done "parent/child"
  ./bin/yx-go done "parent"  # Should succeed
  ./bin/yx-go ls
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement done command with undo and recursive flags`
  - Files: `internal/cmd/done.go`, tests
  - Pre-commit: `shellspec spec/done.sh`

---

### Task 9: Move Command Implementation

- [x] 9. Move Command Implementation

  **What to do**:
  - Create `internal/cmd/move.go`:
    - `yx move <old> <new>` or `yx mv <old> <new>`
    - Resolve old name with fuzzy matching
    - Validate new name
    - Auto-create parent yaks if moving to nested path
    - Preserve state and context when moving
  - Create `internal/yak/store.go:EnsureParents()`:
    - Create parent yak directories with todo state

  **Must NOT do**:
  - Add git logging (Task 11)
  - Allow moving to invalid names

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: File system operations with clear semantics
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `shellspec`: Verify spec/move.sh passes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 10)
  - **Blocks**: None
  - **Blocked By**: Task 4

  **References**:

  **Pattern References**:
  - `bin/yx:454-477` - `ensure_parent_yaks_exist()` - parent creation
  - `bin/yx:479-493` - `move_yak()` - move logic

  **Test References**:
  - `spec/move.sh:1-81` - All move command tests (8 tests)

  **Acceptance Criteria**:

  - [ ] `shellspec spec/move.sh` passes (8 tests)
  - [ ] `yx move "old" "new"` renames yak
  - [ ] `yx mv` alias works
  - [ ] Preserves done state when moving
  - [ ] Preserves context when moving
  - [ ] Auto-creates parent yaks
  - [ ] Rejects invalid new names

  **Manual Verification**:
  ```bash
  ./bin/yx-go add "standalone"
  ./bin/yx-go move "standalone" "parent/child"
  ./bin/yx-go ls
  # Should show parent with child nested
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement move command with implicit parent creation`
  - Files: `internal/cmd/move.go`, `internal/yak/store.go` updates, tests
  - Pre-commit: `shellspec spec/move.sh`

---

### Task 10: Prune Command Implementation

- [x] 10. Prune Command Implementation

  **What to do**:
  - Create `internal/cmd/prune.go`:
    - `yx prune` - remove all done yaks
    - Iterate through all yaks, remove if state = done
    - Use existing rm logic for each

  **Must NOT do**:
  - Add git logging (Task 11 - but note prune logs each removal)
  - Change behavior

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple iteration over existing operations
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `shellspec`: Verify spec/prune.sh passes

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on 5, 6, 8)
  - **Parallel Group**: Wave 3 (after 8)
  - **Blocks**: None
  - **Blocked By**: Tasks 5, 6, 8

  **References**:

  **Pattern References**:
  - `bin/yx:445-452` - `prune_yaks()` - prune logic

  **Test References**:
  - `spec/prune.sh:1-92` - All prune command tests (7 tests)

  **Acceptance Criteria**:

  - [ ] `shellspec spec/prune.sh` passes (7 tests, excluding git logging tests until Task 11)
  - [ ] `yx prune` removes all done yaks
  - [ ] Keeps all todo yaks
  - [ ] Works with nested done yaks
  - [ ] No-op when no done yaks

  **Manual Verification**:
  ```bash
  ./bin/yx-go add "keep" && ./bin/yx-go add "remove"
  ./bin/yx-go done "remove"
  ./bin/yx-go prune
  ./bin/yx-go ls
  # Should only show "keep"
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement prune command`
  - Files: `internal/cmd/prune.go`, tests
  - Pre-commit: `shellspec spec/prune.sh` (partial - git tests later)

---

### Task 11: Git Logging (log_command) Implementation

- [x] 11. Git Logging (log_command) Implementation

  **What to do**:
  - Create `internal/git/git.go`:
    - `IsGitRepository()` - check if in git repo
    - `RunGit(args ...string) (string, error)` - exec git command
  - Create `internal/git/log.go`:
    - `LogCommand(yaksPath, message string)` - commit to refs/notes/yaks
    - Implementation:
      1. Create temp index file
      2. `GIT_INDEX_FILE=... git read-tree --empty`
      3. `GIT_INDEX_FILE=... GIT_WORK_TREE=... git add .`
      4. `git write-tree` to get tree SHA
      5. Get parent from `refs/notes/yaks` if exists
      6. `git commit-tree` with parent args
      7. `git update-ref refs/notes/yaks <commit>`
  - Integrate into all commands: add, done, rm, move, context, prune

  **Must NOT do**:
  - Use go-git library
  - Change commit message format
  - Log when not in git repo

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Complex git plumbing, temp file management, env vars
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `incremental-tdd`: Git operations need careful testing
    - `shellspec`: Verify spec/log_command.sh passes

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (sequential)
  - **Blocks**: Task 12
  - **Blocked By**: Tasks 4, 8

  **References**:

  **Pattern References**:
  - `bin/yx:40-62` - `log_command()` - complete git logging implementation
  - `bin/yx:32-34` - `is_git_repository()` - git detection
  - `bin/yx:36-38` - `yaks_path_exists()` - path check

  **Test References**:
  - `spec/log_command.sh:1-85` - All git logging tests (7 tests)
  - `spec/prune.sh:62-91` - Prune logging tests

  **Acceptance Criteria**:

  - [ ] `shellspec spec/log_command.sh` passes (7 tests)
  - [ ] `yx add "test"` creates commit with message "add test"
  - [ ] `yx done "test"` creates commit "done test"
  - [ ] `yx done --undo "test"` creates commit "done --undo test"
  - [ ] `yx rm "test"` creates commit "rm test"
  - [ ] Commits use configured git author
  - [ ] No commit when not in git repo

  **Manual Verification**:
  ```bash
  cd /tmp && git init test-repo && cd test-repo
  git config user.email "test@example.com"
  git config user.name "Test"
  
  YAKS_PATH=".yaks" ./bin/yx-go add "test"
  git log refs/notes/yaks --oneline
  # Should show: add test
  ```

  **Commit**: YES
  - Message: `feat(git): implement log_command for refs/notes/yaks commits`
  - Files: `internal/git/*.go`, command integrations
  - Pre-commit: `shellspec spec/log_command.sh`

---

### Task 12: Sync Command Implementation

- [ ] 12. Sync Command Implementation

  **What to do**:
  - Create `internal/git/sync.go`:
    - `Sync(yaksPath string)` - complete sync operation
    - Implementation:
      1. Check git setup (is repo, has origin)
      2. Fetch: `git fetch origin refs/notes/yaks:refs/remotes/origin/yaks`
      3. Detect local changes (compare YAKS_PATH to local ref)
      4. If local changes + remote: merge remote into local (copy files)
      5. If local changes: log_command("sync")
      6. Merge refs: use `git merge-tree --write-tree --allow-unrelated-histories`
      7. Create merge commit if needed
      8. Push: `git push origin refs/notes/yaks:refs/notes/yaks`
      9. Extract: `git archive refs/notes/yaks | tar -x -C YAKS_PATH`
      10. Cleanup: `git update-ref -d refs/remotes/origin/yaks`
  - Create `internal/cmd/sync.go`:
    - Command: `yx sync`
    - Error messages: "not in a git repository", "no origin remote configured"

  **Must NOT do**:
  - Use go-git library
  - Change merge strategy (last-write-wins via file copy)
  - Change error messages

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Most complex operation, multiple git commands, merge logic
  - **Skills**: [`incremental-tdd`, `shellspec`]
    - `incremental-tdd`: Sync has many edge cases
    - `shellspec`: Verify all sync spec files pass

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (after Task 11)
  - **Blocks**: Task 15
  - **Blocked By**: Task 11

  **References**:

  **Pattern References**:
  - `bin/yx:638-653` - `check_git_setup()` - validation
  - `bin/yx:655-685` - `detect_local_changes()` - change detection
  - `bin/yx:687-736` - Merge helpers
  - `bin/yx:738-744` - `extract_yaks_to_working_dir()`
  - `bin/yx:769-800` - `sync_yaks()` - main sync orchestration

  **Test References**:
  - `spec/sync.sh:1-140` - Main sync tests (7 tests)
  - `spec/sync_push.sh:1-29` - Push test (1 test)
  - `spec/sync_unit.sh:1-56` - Unit tests (3 tests)
  - `spec/sync_no_pollution.sh:1-58` - Pollution tests (2 tests)
  - `spec/sync_worktrees.sh:1-84` - Worktree tests (3 tests)

  **Acceptance Criteria**:

  - [ ] `shellspec spec/sync*.sh` passes (16 tests total)
  - [ ] `yx sync` pushes local yaks to origin
  - [ ] `yx sync` pulls remote yaks
  - [ ] Merges yaks from multiple users
  - [ ] Last-write-wins for concurrent modifications
  - [ ] Works with git worktrees
  - [ ] No pollution of git index or working tree
  - [ ] Error when not in git repo
  - [ ] Error when no origin remote

  **Manual Verification**:
  ```bash
  # Create origin
  git init --bare /tmp/origin
  
  # Create repo 1
  git clone /tmp/origin /tmp/repo1
  cd /tmp/repo1
  git config user.email "user1@test.com"
  YAKS_PATH=".yaks" ./bin/yx-go add "user1 yak"
  YAKS_PATH=".yaks" ./bin/yx-go sync
  
  # Create repo 2 and sync
  git clone /tmp/origin /tmp/repo2
  cd /tmp/repo2
  YAKS_PATH=".yaks" ./bin/yx-go sync
  YAKS_PATH=".yaks" ./bin/yx-go ls
  # Should show "user1 yak"
  ```

  **Commit**: YES
  - Message: `feat(git): implement sync command with merge-tree conflict resolution`
  - Files: `internal/git/sync.go`, `internal/cmd/sync.go`, tests
  - Pre-commit: `shellspec spec/sync*.sh`

---

### Task 13: Cobra Shell Completions

- [ ] 13. Cobra Shell Completions

  **What to do**:
  - Update `internal/cmd/root.go`:
    - Add Cobra's built-in completion command
    - Configure `ValidArgsFunction` for each command that takes yak names
  - Create `internal/cmd/completions.go`:
    - `yx completions [cmd] [flags]` - output yak names for completion
    - For `done` command: show only incomplete yaks
    - For `done --undo`: show only done yaks
    - For other commands: show all yaks
  - Generate bash/zsh completion scripts via Cobra

  **Must NOT do**:
  - Keep old manual completion files (replace them)
  - Change completion output format

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Cobra provides most functionality
  - **Skills**: [`shellspec`]
    - `shellspec`: Verify spec/completions.sh passes

  **Parallelization**:
  - **Can Run In Parallel**: YES (independent of commands)
  - **Parallel Group**: Can start after Task 1
  - **Blocks**: Task 14
  - **Blocked By**: Task 1 (needs Cobra setup)

  **References**:

  **Pattern References**:
  - `bin/yx:602-636` - `completions()` - completion output logic
  - `completions/yx.bash:1-47` - Current bash completion (to replace)
  - `completions/yx.zsh:1-70` - Current zsh completion (to replace)

  **Test References**:
  - `spec/completions.sh:1-49` - Completion output tests (5 tests)

  **External References**:
  - Cobra completions: https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md

  **Acceptance Criteria**:

  - [ ] `shellspec spec/completions.sh` passes (5 tests)
  - [ ] `yx completions` outputs yak names (one per line)
  - [ ] `yx completions done` filters to incomplete yaks
  - [ ] `yx completions done --undo` filters to done yaks
  - [ ] `yx completion bash` generates bash completion script
  - [ ] `yx completion zsh` generates zsh completion script

  **Manual Verification**:
  ```bash
  ./bin/yx-go add "task1" && ./bin/yx-go add "task2"
  ./bin/yx-go done "task1"
  
  ./bin/yx-go completions
  # Should show: task1\ntask2
  
  ./bin/yx-go completions done
  # Should show: task2
  
  ./bin/yx-go completions done --undo
  # Should show: task1
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement Cobra shell completions`
  - Files: `internal/cmd/completions.go`, `internal/cmd/root.go`
  - Pre-commit: `shellspec spec/completions.sh`

---

### Task 14: Completions Install Command

- [ ] 14. Completions Install Command

  **What to do**:
  - Extend `internal/cmd/completions.go`:
    - `yx completions install [--dry-run]`
    - Detect shell from $SHELL (bash or zsh)
    - Get rc file path (~/.bashrc or ~/.zshrc)
    - Add source line for Cobra-generated completion
    - `--dry-run`: show what would be added
  - Update completion files in `completions/` to use Cobra generation
  - Handle errors gracefully with manual instructions

  **Must NOT do**:
  - Support fish or powershell (not in original)
  - Change error message format

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: File operations with clear logic
  - **Skills**: [`shellspec`]
    - `shellspec`: Verify spec/completions-install.sh passes

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 5 (after Task 13)
  - **Blocks**: Task 15
  - **Blocked By**: Task 13

  **References**:

  **Pattern References**:
  - `bin/yx:532-600` - `install_completions()` - install logic
  - `bin/yx:541-548` - `get_shell_rc_file()` - rc file detection

  **Test References**:
  - `spec/completions-install.sh:1-34` - Install tests (4 tests)

  **Acceptance Criteria**:

  - [ ] `shellspec spec/completions-install.sh` passes (4 tests)
  - [ ] Detects bash shell → ~/.bashrc
  - [ ] Detects zsh shell → ~/.zshrc
  - [ ] `--dry-run` shows what would be added
  - [ ] Error for unknown shell
  - [ ] Graceful error when can't write to rc file

  **Manual Verification**:
  ```bash
  SHELL=/bin/bash ./bin/yx-go completions install --dry-run
  # Should mention .bashrc
  
  SHELL=/bin/zsh ./bin/yx-go completions install --dry-run
  # Should mention .zshrc
  ```

  **Commit**: YES
  - Message: `feat(cmd): implement completions install command`
  - Files: `internal/cmd/completions.go`, updated completion files
  - Pre-commit: `shellspec spec/completions-install.sh`

---

### Task 15: Final Integration, Testing, and Binary Swap

- [ ] 15. Final Integration, Testing, and Binary Swap

  **What to do**:
  - Run full ShellSpec test suite: `shellspec` (all 103 tests)
  - Fix any remaining test failures
  - Run Go tests: `go test ./...`
  - Build final binary: `go build -o bin/yx-go ./cmd/yx`
  - Verify cross-platform build:
    - `GOOS=linux go build -o bin/yx-go-linux ./cmd/yx`
    - `GOOS=darwin go build -o bin/yx-go-darwin ./cmd/yx`
  - Swap binaries:
    - `mv bin/yx bin/yx.bash.bak`
    - `mv bin/yx-go bin/yx`
  - Run full test suite again with swapped binary
  - Remove old completion files (now generated by Cobra)
  - Update README.md with Go build instructions

  **Must NOT do**:
  - Delete bash version until all tests pass
  - Ship without 103/103 tests passing

  **Recommended Agent Profile**:
  - **Category**: `ultrabrain`
    - Reason: Final verification, debugging any remaining issues
  - **Skills**: [`shellspec`, `incremental-tdd`]
    - `shellspec`: Run full test suite
    - `incremental-tdd`: Debug any failures

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 5 (final)
  - **Blocks**: None (end)
  - **Blocked By**: Tasks 12, 14

  **References**:

  **Documentation References**:
  - `README.md:1-80` - Current README to update

  **Acceptance Criteria**:

  - [ ] `shellspec` passes 103/103 tests
  - [ ] `go test ./...` passes all Go tests
  - [ ] `./bin/yx --help` (Go version) works
  - [ ] Cross-platform builds succeed
  - [ ] Binary swapped: `bin/yx` is now Go version
  - [ ] `bin/yx.bash.bak` preserved for rollback
  - [ ] Old completion files removed
  - [ ] README updated with build instructions

  **Manual Verification**:
  ```bash
  # Full test suite
  shellspec
  # Expected: 103 examples, 0 failures
  
  # Go tests
  go test ./...
  # Expected: all pass
  
  # Binary check
  file bin/yx
  # Should show: Mach-O 64-bit executable (or ELF for Linux)
  
  # Smoke test
  ./bin/yx add "final test"
  ./bin/yx ls
  ./bin/yx done "final test"
  ./bin/yx rm "final test"
  ```

  **Commit**: YES
  - Message: `feat(release): complete Go rewrite with 100% test parity`
  - Files: `bin/yx` (new), `bin/yx.bash.bak`, `README.md`, removed old completions
  - Pre-commit: `shellspec && go test ./...`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 1 | `feat(go): initialize Go project structure with Cobra CLI` | devenv.nix, go.mod, cmd/, internal/ | `go build` |
| 2 | `feat(yak): implement core yak types and file storage` | internal/yak/*.go | `go test ./internal/yak/...` |
| 3 | `feat(yak): implement fuzzy matching with ambiguity detection` | internal/yak/fuzzy.go, errors.go | `go test ./internal/yak/...` |
| 4 | `feat(cmd): implement add command with interactive mode` | internal/cmd/add.go | `shellspec spec/add.sh` |
| 5 | `feat(cmd): implement list command with format and filter options` | internal/cmd/list.go, internal/display/ | `shellspec spec/list.sh` |
| 6 | `feat(cmd): implement rm command` | internal/cmd/rm.go | `shellspec spec/rm.sh` |
| 7 | `feat(cmd): implement context command with editor support` | internal/cmd/context.go | `shellspec spec/context.sh` |
| 8 | `feat(cmd): implement done command with undo and recursive flags` | internal/cmd/done.go | `shellspec spec/done.sh` |
| 9 | `feat(cmd): implement move command with implicit parent creation` | internal/cmd/move.go | `shellspec spec/move.sh` |
| 10 | `feat(cmd): implement prune command` | internal/cmd/prune.go | `shellspec spec/prune.sh` |
| 11 | `feat(git): implement log_command for refs/notes/yaks commits` | internal/git/*.go | `shellspec spec/log_command.sh` |
| 12 | `feat(git): implement sync command with merge-tree conflict resolution` | internal/git/sync.go, internal/cmd/sync.go | `shellspec spec/sync*.sh` |
| 13 | `feat(cmd): implement Cobra shell completions` | internal/cmd/completions.go | `shellspec spec/completions.sh` |
| 14 | `feat(cmd): implement completions install command` | internal/cmd/completions.go | `shellspec spec/completions-install.sh` |
| 15 | `feat(release): complete Go rewrite with 100% test parity` | bin/yx, README.md | `shellspec` (103/103) |

---

## Success Criteria

### Verification Commands
```bash
# All ShellSpec tests pass
shellspec
# Expected: 103 examples, 0 failures

# All Go tests pass
go test ./...
# Expected: ok for all packages

# Binary works
./bin/yx --help
# Expected: Usage message

# Cross-platform build
GOOS=linux go build -o /tmp/yx-linux ./cmd/yx && file /tmp/yx-linux
# Expected: ELF 64-bit executable
```

### Final Checklist
- [ ] 103/103 ShellSpec tests passing
- [ ] All Go unit tests passing
- [ ] Binary at `bin/yx` is Go version
- [ ] Bash backup at `bin/yx.bash.bak`
- [ ] devenv.nix includes Go
- [ ] README updated with Go build instructions
- [ ] Old completion files removed (Cobra generates them)
- [ ] Cross-platform builds work (macOS, Linux)
