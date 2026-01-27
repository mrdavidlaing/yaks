# Learnings - yx-go-rewrite

## Conventions & Patterns
<!-- Append discoveries about code conventions, patterns, best practices -->


## Task 1: Project Setup and devenv.nix Integration

### devenv.nix Go Setup
- **Pattern**: Add Go via `packages` list, not `languages.go` module
  - `languages.go.enable = true` alone works fine
  - Specifying `languages.go.package = pkgs.go_1_X` can fail if version is EOL
  - Solution: Use default Go version from devenv (currently 1.25.4)
  - Go 1.23 is EOL and removed from nixpkgs

### Go Module Initialization
- Module name: `github.com/mattwynne/yaks` (matches repo path)
- Cobra dependency: `github.com/spf13/cobra@latest` (v1.10.2)
- Cobra automatically includes transitive deps: mousetrap, pflag

### Directory Structure
- `cmd/yx/` - CLI entry point
- `internal/yak/` - Yak data structures
- `internal/git/` - Git operations
- `internal/display/` - Output formatting
- `internal/cmd/` - Command implementations

### Build & Verification
- Build command: `go build -o bin/yx-go ./cmd/yx`
- Binary size: ~3.5MB (unoptimized)
- Help output works correctly with Cobra's default formatting
- `go mod verify` confirms all dependencies are valid

### Cobra Root Command Pattern
- Use `cobra.Command` struct with `Use`, `Short`, `Long`, `Run` fields
- Default `Run` calls `cmd.Help()` when no subcommand provided
- Cobra auto-generates `-h/--help` flags
- Long description supports multi-line strings


## Task 2: Core Yak Types and File Storage

### Go File I/O Patterns
- `os.MkdirAll(path, 0755)` - creates nested directories (like `mkdir -p`)
- `os.WriteFile(path, []byte("content"), 0644)` - write file atomically
- `os.ReadFile(path)` - read entire file into memory
- `os.Create(path)` - create empty file (returns file handle)
- `os.RemoveAll(path)` - recursive delete (like `rm -rf`)
- `os.Stat(path)` + `os.IsNotExist(err)` - check file existence
- `filepath.WalkDir()` - recursive directory traversal (Go 1.16+)
- `filepath.Rel(base, path)` - get relative path from base
- `filepath.IsAbs(path)` - check if path is absolute
- `filepath.Join()` - cross-platform path joining

### Testing Strategies
- `t.TempDir()` - creates temp directory, auto-cleaned after test
- Each test gets isolated temp directory - no cleanup needed
- Test naming: `TestFunction_Scenario` (e.g., `TestCreate_ValidatesName`)
- Use `t.Fatalf()` for setup failures, `t.Errorf()` for assertions

### State File Format
- Bash `echo "todo"` includes trailing newline (`todo\n`)
- Use `strings.TrimSpace()` when reading state to handle newlines
- Write with explicit newline: `[]byte("todo\n")`

### Validation Pattern
- `strings.ContainsAny(name, "\\:*?|<>\"")` - check forbidden chars
- Forbidden chars: `\ : * ? | < > "`
- Forward slash `/` is allowed (for nested yaks like `parent/child`)
- Return sentinel error `ErrInvalidName` for consistent error handling

### Store API Design
- `NewStore(path)` - constructor handles absolute path conversion
- Methods: `Create`, `Get`, `List`, `Delete`, `Exists`, `Migrate`
- All methods take yak name (relative to BasePath)
- Nested yaks: `parent/child` → `BasePath/parent/child/`

### Migration Pattern
- Old format: `done` file (empty marker file)
- New format: `state` file containing "todo" or "done"
- Migration: find `done` files, create `state` with "done", delete `done` file
- `filepath.WalkDir` to find all `done` files recursively

## Task 3: Fuzzy Matching Implementation

### Fuzzy Matching Algorithm
- **Exact match first**: Use `store.Exists(searchTerm)` to check exact match before fuzzy
- **Substring matching**: Use `strings.Contains(yakName, searchTerm)` for fuzzy matching
- **Match collection**: Loop through all yaks from `store.List()`, collect matches
- **Resolution logic**:
  - 0 matches: Return `ErrYakNotFound(searchTerm)`
  - 1 match: Return the matched yak name
  - 2+ matches: Return `ErrAmbiguousMatch(searchTerm)`

### Error Message Format (Bash Compatibility)
- **Not found**: `"Error: yak '<name>' not found"` (single quotes around name)
- **Ambiguous**: `"Error: yak name '<name>' is ambiguous"` (single quotes around name)
- Use `fmt.Errorf()` with format strings to generate exact messages
- Tests verify error messages byte-for-byte - format must match exactly

### Error Handling Pattern
- Create error constructor functions: `ErrYakNotFound(name)` and `ErrAmbiguousMatch(name)`
- Return error values, not error strings
- Allows callers to check error type with `errors.Is()` if needed

### TDD Workflow for Fuzzy Matching
- **Cycle 1**: Exact match test → minimal implementation
- **Cycle 2**: Fuzzy unique match test → algorithm already handles it
- **Cycle 3**: Not found error test → error handling works
- **Cycle 4**: Ambiguous match test → multi-match logic works
- All cycles passed with single implementation - algorithm is complete

### Go String Matching
- `strings.Contains(haystack, needle)` - case-sensitive substring check
- No need for case-insensitive matching (bash is case-sensitive)
- Simple and efficient for small yak lists

### Testing Patterns
- Create multiple yaks with overlapping names to test ambiguity
- Verify exact error message format in assertions
- Use `t.TempDir()` for isolated test environments
- Test setup comments clarify test intent (e.g., "Create multiple yaks")

## Task 4: Add Command Implementation

### Cobra Command Pattern for CLI Commands
- **Command factory function**: `NewAddCmd(store *yak.Store) *cobra.Command`
- **Use field**: Defines command syntax (e.g., `"add [name...]"`)
- **RunE field**: Use `RunE` instead of `Run` to return errors
- **Error handling**: Return error from RunE, Cobra handles exit code (1 on error)
- **Subcommand registration**: `rootCmd.AddCommand(cmd.NewAddCmd(store))` in main

### Interactive Mode Implementation
- **TTY detection**: Not needed - just read stdin with `bufio.Scanner`
- **Scanner pattern**: `scanner := bufio.NewScanner(os.Stdin)` then `scanner.Scan()`
- **Empty line detection**: `if line == ""` to break loop
- **Prompt output**: Use `fmt.Println()` to stdout (not stderr)
- **Error handling**: Check `scanner.Err()` after loop for read errors

### Single Mode Implementation
- **Multi-word argument joining**: `strings.Join(args, " ")` joins all args with spaces
- **Example**: `yx add this is a test` → `"this is a test"` (single yak name)
- **Validation**: Delegated to `store.Create()` which validates the name

### Error Output Pattern
- **Error messages**: Use `fmt.Fprintf(os.Stderr, "Error: %v\n", err)`
- **Return error**: Return the error from RunE for proper exit code
- **Cobra behavior**: Prints usage info after error (expected behavior)

### Store Integration in main.go
- **YAKS_PATH env var**: Check `os.Getenv("YAKS_PATH")` with default `".yaks"`
- **Store creation**: `store := yak.NewStore(yaksPath)`
- **Migration**: Call `store.Migrate()` on startup to handle old `done` files
- **Command wiring**: Add all commands before `rootCmd.Execute()`

### ShellSpec Test Results
- **11 of 13 tests pass** for add command
- **2 tests fail** because they depend on `list` command (Task 5)
- **Independent tests pass**:
  - Single yak creation
  - Interactive mode prompt
  - Nested yak names with `/`
  - All validation tests (forbidden characters)
- **Integration tests fail** (expected):
  - "adds multiple yaks in interactive mode" - needs list
  - "captures multi-word yak names without quotes" - needs list

### Code Quality Notes
- **No unnecessary comments**: Code is self-explanatory
- **Error handling**: Consistent pattern across functions
- **Function naming**: `addInteractive()` and `addSingle()` are clear
- **Separation of concerns**: Store handles validation, command handles I/O

### Build & Verification
- **Build command**: `go build -o bin/yx-go ./cmd/yx`
- **Binary size**: ~3.5MB (same as before)
- **Symlink for tests**: `ln -sf yx-go bin/yx` allows ShellSpec to find binary
- **Test execution**: `PATH="$(pwd)/bin:$PATH" shellspec spec/add.sh`

## Task 6: Remove Command Implementation

### Rm Command Pattern
- **Command factory**: `NewRmCmd(store *yak.Store) *cobra.Command`
- **Use field**: `"rm [name...]"` - accepts one or more arguments
- **Multi-word names**: `strings.Join(args, " ")` joins all args with spaces
- **Example**: `yx rm this is a test` → resolves to "this is a test" yak

### Fuzzy Matching Integration
- **Resolve name**: Use `yak.FindYak(store, name)` to handle fuzzy matching
- **Error handling**: FindYak returns error if not found or ambiguous
- **Error output**: Print error to stderr with `fmt.Fprintf(os.Stderr, "%v\n", err)`
- **Return error**: Return error from RunE for proper exit code

### Deletion Implementation
- **Store method**: `store.Delete(resolvedName)` removes entire yak directory
- **Recursive delete**: Uses `os.RemoveAll()` internally (like `rm -rf`)
- **Nested yaks**: Deletes entire subtree if parent is deleted
- **Error handling**: Check for delete errors and report to stderr

### Error Handling Pattern
- **Not found**: FindYak returns "Error: yak '<name>' not found"
- **Ambiguous**: FindYak returns "Error: yak name '<name>' is ambiguous"
- **Delete error**: Wrap with "Error: " prefix for consistency
- **Stderr output**: All errors go to stderr, not stdout

### ShellSpec Test Results
- **5 of 5 tests pass** for rm command
- **Test coverage**:
  - Removes a yak by name
  - Shows error when yak not found
  - Handles removing the only yak
  - Removes multi-word yak names without quotes
  - Removes a nested yak

### Code Quality
- **No unnecessary comments**: Code is self-documenting
- **Consistent error handling**: Matches add command pattern
- **Minimal implementation**: Only 37 lines of code
- **Reuses existing patterns**: Follows add command structure

### Integration with main.go
- **Command registration**: `rootCmd.AddCommand(cmd.NewRmCmd(store))`
- **Store dependency**: Passed to command factory function
- **No additional setup**: Works with existing store initialization

### Key Insight
- **Fuzzy matching is powerful**: Allows users to delete yaks without exact names
- **Error messages matter**: Tests verify exact error text for user experience
- **Composition over duplication**: Reusing FindYak and store.Delete avoids code duplication

## Task 7: Context Command Implementation

### Context Command Pattern
- **Command factory**: `NewContextCmd(store *yak.Store) *cobra.Command`
- **Use field**: `"context [flags] [name...]"` - accepts flags and yak name
- **Flags**: `--show` (display context) and `--edit` (edit context, default)
- **Flag parsing**: Use `cmd.Flags().GetBool("show")` to check flag values

### Yak Name Resolution
- **Fuzzy matching**: Use `yak.FindYak(store, name)` to resolve yak names
- **Multi-word names**: `strings.Join(args, " ")` joins all args with spaces
- **Error handling**: FindYak returns error if not found or ambiguous
- **Error output**: Print error to stderr with `fmt.Fprintf(os.Stderr, "Error: %v\n", err)`

### Show Mode Implementation
- **Output format**: `<resolved_name>\n\n<context.md contents>`
- **First line**: Always print the resolved yak name
- **Blank line**: Print empty line only if context file has content
- **File reading**: Use `os.ReadFile(contextPath)` to read context
- **Empty check**: Only print content if file exists and has non-zero length
- **No error on missing file**: Show mode succeeds even if context.md doesn't exist

### Edit Mode Implementation
- **TTY detection**: Use `os.Stdin.Stat()` to check if stdin is a TTY
  - `fi, _ := os.Stdin.Stat()`
  - `isTTY := (fi.Mode() & os.ModeCharDevice) != 0`
- **Editor launch**: Use `exec.Command(editor, contextPath)` to launch $EDITOR
  - Default editor: `vi` if $EDITOR not set
  - Connect stdin/stdout/stderr to terminal: `cmd.Stdin = os.Stdin`, etc.
  - Run with `cmd.Run()` to wait for completion
- **Stdin mode**: If not TTY, read from stdin with `io.ReadAll(os.Stdin)`
  - Write directly to context.md with `os.WriteFile(contextPath, content, 0644)`
  - Allows piping: `echo "context" | yx context myak`

### Path Construction
- **YAKS_PATH env var**: Check `os.Getenv("YAKS_PATH")` with default `".yaks"`
- **Context file path**: `$YAKS_PATH/<resolved_name>/context.md`
- **Path expansion**: Use `os.ExpandEnv()` to expand environment variables
- **Fallback**: If env var not set, use literal string as fallback

### Error Handling Pattern
- **Not found**: FindYak returns "Error: yak '<name>' not found"
- **Ambiguous**: FindYak returns "Error: yak name '<name>' is ambiguous"
- **Stdin read error**: Wrap with "Error reading from stdin: %v"
- **File write error**: Wrap with "Error writing context: %v"
- **Stderr output**: All errors go to stderr, not stdout

### ShellSpec Test Results
- **7 of 7 tests pass** for context command
- **Test coverage**:
  - Sets context from stdin (default edit mode)
  - Shows a yak without context (no blank line if empty)
  - Shows a yak with context (includes blank line and content)
  - Replaces existing context from stdin (overwrites old content)
  - Shows error when yak not found (edit mode)
  - Shows error when yak not found (show mode)
  - Sets and shows context for nested yak (parent/child)

### Code Quality
- **Necessary comments**: Explain TTY detection, editor launch, stdin reading
- **Error handling**: Consistent pattern across show and edit modes
- **Function separation**: `showContext()` and `editContext()` are clear
- **Reuses patterns**: Follows add/rm command structure

### Integration with main.go
- **Command registration**: `rootCmd.AddCommand(cmd.NewContextCmd(store))`
- **Store dependency**: Passed to command factory function
- **No additional setup**: Works with existing store initialization

### Key Insights
- **TTY detection is critical**: Enables both interactive and piped usage
- **Editor integration**: Respects $EDITOR env var for user preference
- **Graceful degradation**: Show mode works even if context.md is empty
- **Composition**: Reuses FindYak for consistent yak resolution

## Task 8: Done Command Implementation

### Done Command Pattern
- **Command factory**: `NewDoneCmd(store *yak.Store) *cobra.Command`
- **Use field**: `"done [name...]"` - accepts yak name
- **Flags**: `--undo` (mark as todo) and `--recursive` (mark all children)
- **Flag parsing**: Use `cmd.Flags().BoolVar(&undo, "undo", false, "...")` for boolean flags

### State Management
- **SetState method**: `store.SetState(name, state)` writes state to file
- **State format**: Write `state.String() + "\n"` to match bash behavior
- **State file path**: `$YAKS_PATH/<name>/state`
- **State values**: `yak.StateTodo` ("todo") and `yak.StateDone` ("done")

### Incomplete Children Check
- **HasIncompleteChildren method**: Checks if any direct children are not done
- **Implementation**: Use `os.ReadDir()` to list child directories
- **Child resolution**: Build child name with `filepath.Join(name, entry.Name())`
- **State check**: Get each child's state with `store.Get(childName)`
- **Return true**: If any child has state != StateDone

### Recursive Marking
- **MarkDoneRecursively method**: Marks yak and all descendants as done
- **Implementation**: Set state to done, then recurse into child directories
- **Depth-first**: Mark current yak first, then process children
- **Error propagation**: Return first error encountered

### Validation Logic
- **Parent protection**: Cannot mark parent done if it has incomplete children
- **Error message**: "Error: cannot mark '<name>' as done - it has incomplete children"
- **Bypass with --recursive**: Marks all children done first, then parent

### ShellSpec Test Results
- **10 of 10 tests pass** for done command
- **Test coverage**:
  - Marks a yak as done
  - Shows error when marking non-existent yak as done
  - Displays mix of done and not-done yaks
  - Handles yak names starting with x
  - Marks yak starting with x as done correctly
  - Unmarks a done yak with --undo flag
  - Marks a nested yak as done
  - Migrates old done files to state files
  - Errors when marking parent with incomplete children
  - Marks parent and all children with --recursive flag

### Code Quality
- **No unnecessary comments**: Code is self-documenting
- **Consistent error handling**: Matches other command patterns
- **Store method placement**: Added SetState, HasIncompleteChildren, MarkDoneRecursively to store.go
- **Reuses patterns**: Follows add/rm/context command structure

### Key Insights
- **Validation before action**: Check incomplete children before marking done
- **Recursive operations**: Use depth-first traversal for consistent behavior
- **Flag handling**: Cobra's BoolVar makes flag parsing simple
- **Error messages**: Match bash format exactly for test compatibility

## Task 9: Move Command Implementation

### Move Command Pattern
- **Command factory**: `NewMoveCmd(store *yak.Store) *cobra.Command`
- **Alias command**: `NewMvCmd(store *yak.Store) *cobra.Command` - identical implementation
- **Use field**: `"move <old> <new...>"` - accepts old name and new name (multi-word)
- **Multi-word names**: `strings.Join(args[1:], " ")` joins all args after first with spaces
- **Example**: `yx move "old name" new parent child` → moves to "new parent child"

### Yak Name Resolution
- **Fuzzy matching**: Use `yak.FindYak(store, oldName)` to resolve old yak name
- **Error handling**: FindYak returns error if not found or ambiguous
- **Error output**: Print error to stderr with `fmt.Fprintf(os.Stderr, "Error: %v\n", err)`
- **Return error**: Return error from RunE for proper exit code

### Name Validation
- **New name validation**: Use `yak.ValidateName(newName)` to check for forbidden characters
- **Forbidden chars**: `\ : * ? | < > "` (same as create command)
- **Forward slash allowed**: `/` is allowed for nested paths like `parent/child`
- **Error handling**: ValidateName returns ErrInvalidName if validation fails

### Parent Creation
- **EnsureParents method**: New Store method that creates parent yaks if needed
- **Implementation**: Split parent path by `/`, create each level iteratively
- **Idempotent**: Only creates parents that don't exist (checks with `s.Exists()`)
- **State files**: Each created parent gets `state` file with "todo" and empty `context.md`
- **Example**: Moving to `a/b/c` creates `a` and `a/b` if they don't exist

### Move Implementation
- **Move method**: New Store method that renames/relocates yak directory
- **Implementation**: `os.Rename(oldPath, newPath)` - atomic filesystem operation
- **Preserves content**: All files in yak directory (state, context.md) are preserved
- **Nested moves**: Can move flat yak into nested position (e.g., `standalone` → `parent/child`)

### State and Context Preservation
- **Automatic preservation**: `os.Rename()` preserves all directory contents
- **State file**: Existing state (todo/done) is preserved during move
- **Context file**: Existing context.md is preserved during move
- **No special handling needed**: Filesystem operation handles everything

### ShellSpec Test Results
- **8 of 8 tests pass** for move command
- **Test coverage**:
  - Renames a yak (flat to flat)
  - Shows error when source yak not found
  - Preserves done state when renaming
  - Preserves context when renaming
  - Supports mv as an alias for move
  - Rejects new name with forbidden characters
  - Moves a flat yak into a nested position
  - Implicitly creates parent yaks when moving

### Code Quality
- **No unnecessary comments**: Code is self-documenting
- **Consistent error handling**: Matches other command patterns
- **Function separation**: `moveYak()` helper function encapsulates logic
- **Reuses patterns**: Follows add/rm/context/done command structure

### Integration with main.go
- **Command registration**: 
  - `rootCmd.AddCommand(cmd.NewMoveCmd(store))`
  - `rootCmd.AddCommand(cmd.NewMvCmd(store))`
- **Store dependency**: Passed to command factory functions
- **No additional setup**: Works with existing store initialization

### Key Insights
- **Atomic filesystem operations**: `os.Rename()` is atomic and preserves all content
- **Parent creation is idempotent**: Safe to call multiple times
- **Fuzzy matching + validation**: Combines FindYak for resolution with ValidateName for new name
- **Alias commands**: Simple to implement - just duplicate command with different Use field
- **Nested yaks**: Forward slash in names enables hierarchical organization

## Task 10: Prune Command Implementation

### Prune Command Pattern
- **Command factory**: `NewPruneCmd(store *yak.Store) *cobra.Command`
- **Use field**: `"prune"` - no arguments needed
- **Short description**: "Remove all done yaks"
- **No flags**: Simple command with no options

### Implementation Logic
- **List all yaks**: Use `store.List()` to get all yak names
- **Iterate and check**: Loop through each yak name
- **Get state**: Use `store.Get(name)` to retrieve yak state
- **Check if done**: Compare `y.State == yak.StateDone`
- **Delete if done**: Call `store.Delete(name)` for done yaks
- **Error handling**: Continue on individual errors (don't fail entire prune)

### Error Handling Pattern
- **Get errors**: Use `continue` to skip yaks that can't be retrieved
- **Delete errors**: Log to stderr with `fmt.Fprintf(cmd.OutOrStderr(), ...)`
- **No fatal errors**: Prune always succeeds (returns nil) even if some deletes fail
- **Graceful degradation**: Removes as many done yaks as possible

### ShellSpec Test Results
- **6 of 6 tests pass** for prune command
- **Test coverage**:
  - Removes all done yaks (mixed todo/done)
  - Handles prune when no yaks exist (empty directory)
  - Keeps all yaks when none are done (no-op)
  - Removes all yaks when all are done (complete cleanup)
  - Removes done child yaks (nested yaks)
  - Logs each yak removal individually (git notes integration)

### Code Quality
- **Minimal implementation**: Only 36 lines of code
- **No unnecessary comments**: Code is self-documenting
- **Consistent error handling**: Matches other command patterns
- **Reuses patterns**: Follows add/rm/context/done/move command structure

### Integration with main.go
- **Command registration**: `rootCmd.AddCommand(cmd.NewPruneCmd(store))`
- **Store dependency**: Passed to command factory function
- **No additional setup**: Works with existing store initialization

### Key Insights
- **Iteration pattern**: Loop through all items, filter by condition, apply action
- **Graceful error handling**: Continue on errors rather than failing entire operation
- **No-op safety**: Prune succeeds even if there's nothing to prune
- **Nested yak support**: Works automatically with nested yaks (parent/child)
- **Simple and effective**: Straightforward implementation with no special cases

## Task 11: Git Logging (log_command) Implementation

### Git Package Structure
- **Package location**: `internal/git/` - separate package for git operations
- **Files created**:
  - `git.go` - IsGitRepository(), RunGit(), RunGitSilent() helpers
  - `log.go` - LogCommand() implementation
  - `git_test.go` - Unit tests for all functions

### Git Plumbing Implementation
- **Temp index file**: Use `os.CreateTemp("", "yx-git-index-*")` to avoid polluting working tree
- **Environment variables**: Pass via map to RunGit/RunGitSilent functions
- **Command sequence** (matches bash reference):
  1. Check IsGitRepository() - return early if not in git repo
  2. Check yaksPath exists - return early if not
  3. Create temp index file with cleanup via defer
  4. `git read-tree --empty` with GIT_INDEX_FILE env
  5. `git add .` with GIT_INDEX_FILE and GIT_WORK_TREE env
  6. `git write-tree` with GIT_INDEX_FILE env → get tree SHA
  7. `git rev-parse refs/notes/yaks` → get parent SHA (ignore error if not exists)
  8. `git commit-tree <tree> [-p <parent>] -m <message>` → get new commit SHA
  9. `git update-ref refs/notes/yaks <new_commit>`

### Error Handling Pattern
- **Silent failures**: LogCommand() never returns errors to caller
- **Graceful degradation**: Commands succeed even if logging fails
- **Early returns**: Check preconditions first, return silently if not met
- **No stderr output**: Unlike other commands, git logging is completely silent on failure

### Integration Pattern
- **Store.BasePath**: Pass `store.BasePath` to LogCommand() for yaks directory path
- **Message format**: Match bash exactly - `"add <name>"`, `"done <name>"`, `"rm <name>"`, etc.
- **Placement**: Call LogCommand() AFTER successful operation, not before
- **Prune special case**: Log `"rm <name>"` for each deleted yak individually

### Command-Specific Messages
- **add**: `"add <name>"`
- **done**: `"done <name>"`
- **done --undo**: `"done --undo <name>"`
- **done --recursive**: `"done --recursive <name>"`
- **rm**: `"rm <name>"`
- **move**: `"move <old> <new>"`
- **context**: `"context <name>"`
- **prune**: `"rm <name>"` for each deleted yak

### Testing Strategy
- **Go unit tests**: Test IsGitRepository, RunGit, LogCommand in isolation
- **ShellSpec tests**: 7 tests in spec/log_command.sh verify end-to-end behavior
- **Test setup**: Create temp git repo with user.email and user.name configured
- **Verification**: Check refs/notes/yaks exists and has correct commit messages

### Key Insights
- **Git plumbing vs porcelain**: Use low-level git commands (read-tree, write-tree, commit-tree, update-ref) for precise control
- **Temp index isolation**: Critical for not polluting working tree or main index
- **Parent ref handling**: Check if refs/notes/yaks exists before adding -p flag
- **Environment variable passing**: Go's exec.Command needs explicit env setup


## Task 11: Git Logging (log_command) Implementation

### Git Plumbing Pattern
- **IsGitRepository()**: Use `git rev-parse --is-inside-work-tree` to check if in git repo
- **RunGit()**: Helper function to execute git commands with environment variables
- **RunGitSilent()**: Helper for commands where output is not needed (read-tree, add, update-ref)
- **Environment variables**: Use map[string]string to pass GIT_INDEX_FILE and GIT_WORK_TREE

### Temp Index File Management
- **Create temp file**: `os.CreateTemp("", "yx-git-index-*")` creates unique temp file
- **Close immediately**: Close file handle after creation (git will reopen it)
- **Defer cleanup**: `defer os.Remove(tempIndexPath)` ensures cleanup even on error
- **Path extraction**: Use `tempIndex.Name()` to get path before closing

### Git Command Sequence (CRITICAL)
1. **read-tree --empty**: Initialize empty index in temp file
2. **add .**: Add all files from YAKS_PATH to temp index
3. **write-tree**: Create tree object from temp index, get SHA
4. **rev-parse refs/notes/yaks**: Get parent commit SHA (if exists)
5. **commit-tree**: Create commit with tree SHA, parent args, and message
6. **update-ref**: Update refs/notes/yaks to point to new commit

### Environment Variable Handling
- **GIT_INDEX_FILE + GIT_WORK_TREE**: Used together for read-tree and add
- **GIT_INDEX_FILE only**: Used for write-tree (no work tree needed)
- **No env vars**: Used for rev-parse, commit-tree, update-ref (operate on repo)

### Error Handling Pattern
- **Silent failures**: LogCommand() never returns errors to caller
- **Early returns**: Return immediately on any error (graceful degradation)
- **No stderr output**: Errors are silently ignored (logging is best-effort)
- **Commands succeed**: Even if logging fails, the command operation succeeds

### Integration Pattern
- **Import git package**: `import "github.com/mattwynne/yaks/internal/git"`
- **Call after success**: Only call LogCommand() after store operation succeeds
- **Pass BasePath**: Use `store.BasePath` as yaksPath argument
- **Message format**: Match bash exactly (e.g., "add <name>", "done --undo <name>")

### Testing Strategy
- **Go unit tests**: Test IsGitRepository(), RunGit(), RunGitSilent() in isolation
- **ShellSpec integration**: Test full workflow with real git repository
- **Test coverage**: 7 Go tests + 7 ShellSpec tests = 14 total tests

### Key Insights
- **Temp index isolation**: Using temp index file prevents polluting working tree
- **Parent ref handling**: Check if refs/notes/yaks exists before adding -p flag
- **Graceful degradation**: Logging failures don't break commands
- **Exact bash parity**: Git command sequence matches bash implementation exactly
- **Environment variable scope**: Different commands need different env vars

### Code Quality
- **Helper functions**: RunGit() and RunGitSilent() reduce duplication
- **Clear separation**: git package handles all git operations
- **No comments needed**: Code is self-documenting with clear function names
- **Consistent pattern**: All commands integrate LogCommand() the same way


## Task 12: Sync Command Implementation

### Sync Architecture
- **Main orchestration**: `Sync()` function coordinates all sync operations
- **Helper functions**: CheckGitSetup, GetLocalRef, GetRemoteRef, DetectLocalChanges, MergeRemoteIntoLocalYaks, MergeLocalAndRemote, ExtractYaksToWorkingDir
- **Error handling**: CheckGitSetup returns errors, all other operations are best-effort (silent failures)

### Git Operations Sequence
1. **CheckGitSetup**: Validate git repo and origin remote
2. **Fetch**: `git fetch origin refs/notes/yaks:refs/remotes/origin/yaks` (ignore errors)
3. **Get refs**: GetRemoteRef and GetLocalRef to get SHA values
4. **Detect changes**: Compare YAKS_PATH to local ref using `diff -qr`
5. **Merge local changes**: If local changes exist and remote has changes, merge remote into local (last-write-wins)
6. **Log changes**: Call LogCommand("sync") if local changes detected
7. **Merge refs**: Use `git merge-tree --write-tree --allow-unrelated-histories` for ref merging
8. **Push**: `git push origin refs/notes/yaks:refs/notes/yaks` (ignore errors)
9. **Extract**: Extract yaks from refs/notes/yaks to YAKS_PATH
10. **Cleanup**: Delete refs/remotes/origin/yaks

### Directory Comparison Pattern
- **Create temp dir**: `os.MkdirTemp("", "yx-sync-ref-*")`
- **Extract ref**: Use `git archive <ref> | tar -x -C <dir>` via piped commands
- **Compare**: Use `diff -qr dir1 dir2` to compare directories
- **Cleanup**: `defer os.RemoveAll(tempDir)` for automatic cleanup

### Git Archive Extraction
- **Pipe pattern**: Create pipe between `git archive` stdout and `tar` stdin
- **Start both commands**: archiveCmd.Start() then tarCmd.Start()
- **Wait for both**: archiveCmd.Wait() then tarCmd.Wait()
- **Error handling**: Silent failures (return early on errors)

### Last-Write-Wins Merge Strategy
- **Extract remote to temp**: Extract remote ref to temp directory
- **Copy local over remote**: Use copyDir to copy local files over remote files
- **Replace YAKS_PATH**: Remove old YAKS_PATH, create new, copy merged result
- **Local wins**: Local files overwrite remote files in conflicts

### Git Merge-Tree Usage
- **Command**: `git merge-tree --write-tree --allow-unrelated-histories <local> <remote>`
- **Returns**: Tree SHA on success
- **Create merge commit**: `git commit-tree <tree> -p <local> -p <remote> -m "Merge yaks"`
- **Update ref**: `git update-ref refs/notes/yaks <commit>`

### Test Results
- **8/16 tests pass**: All tests that don't use `yx ls` pass
- **8 tests fail**: All failures are due to missing `ls` command (Task 5 blocked)
- **Sync functionality verified**: Manual testing confirms push/pull/merge works correctly

### Key Insights
- **Piped commands in Go**: Use StdoutPipe() and set Stdin to pipe for command chaining
- **Directory comparison**: `diff -qr` is simple and effective for comparing directories
- **Best-effort operations**: Network operations (fetch/push) should not fail the sync
- **Ref cleanup**: Always delete refs/remotes/origin/yaks after sync to avoid pollution
- **Test dependencies**: Many sync tests depend on `ls` command which is blocked (Task 5)

## Task 12: Sync Command Implementation

### Sync Architecture
- **CheckGitSetup()**: Validates git repository and origin remote existence
- **GetLocalRef() / GetRemoteRef()**: Retrieve SHA from refs/notes/yaks and refs/remotes/origin/yaks
- **DetectLocalChanges()**: Compare YAKS_PATH to local ref using temp directory and diff
- **MergeRemoteIntoLocalYaks()**: Last-write-wins merge by copying local over remote
- **MergeLocalAndRemote()**: Merge refs using git merge-tree with unrelated histories
- **ExtractYaksToWorkingDir()**: Extract from ref to filesystem using git archive + tar
- **Sync()**: Main orchestration function that coordinates all operations

### Git Archive + Tar Pattern
- **Pipe commands**: Use `StdoutPipe()` to connect git archive to tar
- **Start both**: Start archive first, then tar
- **Wait for both**: Wait for archive, then tar
- **Error handling**: Silent failures, continue on errors

### Directory Comparison
- **diff -qr**: Use `diff -qr dir1 dir2` to compare directory trees
- **Exit code**: 0 = equal, non-zero = different
- **Suppress output**: Set Stderr and Stdout to nil

### Last-Write-Wins Merge Strategy
- **Extract remote**: Extract remote ref to temp directory
- **Copy local over**: Copy local YAKS_PATH files over remote files (overwrites)
- **Replace YAKS_PATH**: Remove old YAKS_PATH, replace with merged result
- **Ensures local wins**: Local changes always take precedence in conflicts

### Git Merge-Tree Usage
- **Command**: `git merge-tree --write-tree --allow-unrelated-histories <local> <remote>`
- **Returns tree SHA**: Output is the SHA of the merged tree
- **Create merge commit**: `git commit-tree <tree> -p <local> -p <remote> -m "Merge yaks"`
- **Update ref**: `git update-ref refs/notes/yaks <merge_commit>`

### Sync Workflow
1. Check git setup (repo + origin)
2. Fetch remote: `git fetch origin refs/notes/yaks:refs/remotes/origin/yaks`
3. Get local and remote refs
4. Detect if local has changes
5. If local changes + remote exists: merge remote into local (last-write-wins)
6. If local changes: log_command("sync") to create commit
7. Merge local and remote refs (git merge-tree)
8. Push to origin: `git push origin refs/notes/yaks:refs/notes/yaks`
9. Extract to working directory
10. Cleanup: delete refs/remotes/origin/yaks

### Error Handling
- **CheckGitSetup errors**: Return error to caller (fatal)
- **Fetch/push errors**: Silently ignore (network issues)
- **Extract errors**: Silently ignore (empty refs)
- **Merge errors**: Silently ignore (no conflicts expected)

### Key Insights
- **Temp directory cleanup**: Use defer to ensure cleanup
- **Silent operations**: Most git operations are best-effort
- **Ref manipulation**: Sync operates only on refs, not working tree
- **Network resilience**: Continue even if fetch/push fails
- **Last-write-wins**: Simple conflict resolution strategy


## Task 13: Cobra Shell Completions

### Completions Command Pattern
- **Command factory**: `NewCompletionsCmd(store *yak.Store) *cobra.Command`
- **Use field**: `"completions [command] [flags]"`
- **DisableFlagParsing**: Set to `true` to allow `--undo` to be passed as argument, not parsed as flag
- **Argument parsing**: First arg is command name (e.g., "done"), second arg is flag (e.g., "--undo")

### Flag Parsing Issue
- **Problem**: Cobra normally parses `--undo` as a flag for the completions command itself
- **Solution**: Set `DisableFlagParsing: true` in command definition
- **Result**: All args are passed as-is to RunE function, allowing manual parsing

### Filtering Logic
- **For "done" command without flag**: Show only incomplete yaks (state == StateTodo)
- **For "done" command with "--undo" flag**: Show only done yaks (state == StateDone)
- **For all other commands**: Show all yaks
- **Error handling**: Skip yaks that can't be retrieved (return false from shouldInclude)

### Output Format
- **One yak per line**: Use `fmt.Println(name)` for each yak
- **Sorted output**: Use `sort.Strings(yaks)` for consistent ordering
- **Empty output**: If no yaks match filter, output nothing (empty string)

### ShellSpec Test Results
- **5 of 5 tests pass** for completions command
- **Test coverage**:
  - Outputs nothing when no yaks exist
  - Lists all yak names (one per line)
  - Filters to incomplete yaks for "done" command
  - Filters to done yaks for "done --undo" command
  - Includes nested yak paths in completions

### Code Quality
- **No unnecessary comments**: Code is self-documenting
- **Minimal implementation**: Only 50 lines of code
- **Reuses patterns**: Follows other command structure
- **DisableFlagParsing**: Critical for handling `--undo` as argument

### Integration with main.go
- **Command registration**: `rootCmd.AddCommand(cmd.NewCompletionsCmd(store))`
- **Store dependency**: Passed to command factory function
- **No additional setup**: Works with existing store initialization

### Key Insights
- **DisableFlagParsing is powerful**: Allows custom argument parsing for special cases
- **Sorting matters**: Consistent output order is important for shell completion
- **Filter by state**: Reuses yak.StateTodo and yak.StateDone constants
- **Graceful degradation**: Skips yaks that can't be retrieved instead of failing
