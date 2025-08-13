# Feature 003: Handle Multiple Workspace Files with Selection Prompt

## Overview
Implement functionality to handle scenarios where multiple `.code-workspace` files are found in the current working directory, providing an interactive selection prompt to allow users to choose which workspace file to use.

## Background
The Cursor Git WorkTree Manager (wtm) needs to handle cases where users have multiple workspace configurations in the same directory. This can happen when:
- Users maintain different workspace configurations for different purposes
- Legacy workspace files are not cleaned up
- Users are experimenting with different workspace setups

When multiple workspace files are detected, the system should provide a clear, user-friendly way to select the appropriate workspace file rather than failing or making an arbitrary choice.

## Requirements

### Functional Requirements
1. **Multiple Workspace Detection**: Detect when multiple `.code-workspace` files exist in the current directory
2. **Interactive Selection Prompt**: Present a numbered list of available workspace files to the user
3. **User Input Validation**: Validate user input and handle invalid selections gracefully
4. **Selection Confirmation**: Allow users to confirm their selection before proceeding
5. **Cancel/Exit Option**: Provide a way for users to cancel the selection and exit gracefully
6. **Integration with Workspace Detection**: Seamlessly integrate with the workspace detection flow from Feature 002

### Non-Functional Requirements
1. **Performance**: Selection prompt should be responsive (< 50ms to display)
2. **User Experience**: Clear, intuitive interface that doesn't require documentation
3. **Cross-Platform**: Work consistently on Windows, macOS, and Linux
4. **Testability**: Use Uber gomock for mocking user input operations
5. **Minimal Dependencies**: Use only Go standard library + gomock for input operations
6. **Accessibility**: Support for users who may have difficulty with interactive prompts

## Technical Specification

### Interface Design

#### FS Package Extension (File System Adapter)
**No new interface methods required** - uses existing `Glob()` method from Feature 002.

#### WTM Package Extension (Business Logic)
**New Interface Methods:**
- `handleMultipleWorkspaces(workspaceFiles []string) (string, error)`: Handle multiple workspace file selection
- `displayWorkspaceSelection(workspaceFiles []string)`: Display the selection prompt
- `getUserSelection(maxChoice int) (int, error)`: Get and validate user input
- `confirmSelection(workspaceFile string) (bool, error)`: Confirm user's selection

**Implementation Structure:**
- Extends existing WTM package with multiple workspace handling
- Private helper methods for selection UI and input validation
- Error handling with wrapped errors
- Clean separation of concerns

**Key Characteristics:**
- **NO direct file system access** - all operations go through FS adapter
- **ONLY unit tests** using mocked input/output operations
- Business logic focused on user interaction and selection
- Testable through dependency injection
- **Pure business logic** with no file system dependencies

### Implementation Details

#### 1. Multiple Workspace Detection
The system will detect multiple workspace files using the existing `Glob()` method:

**Detection Flow:**
1. Use `fs.Glob("*.code-workspace")` to find all workspace files in current directory only
2. If exactly one file found: proceed with normal workspace detection (no selection needed)
3. If multiple files found: trigger selection prompt
4. If no files found: fall back to single repository detection

**File Filtering:**
- Only consider files with `.code-workspace` extension in current directory
- Exclude directories that might have similar names
- Sort files alphabetically for consistent ordering
- No recursive search - only current directory

#### 2. Selection Prompt Interface
The selection prompt will provide a clear, numbered interface:

**Display Format:**
```
Multiple workspace files found. Please select one:

1. project.code-workspace
2. project-dev.code-workspace
3. project-staging.code-workspace

Enter your choice (1-3) or 'q' to quit: 
```

**Key Features:**
- Numbered list starting from 1
- Clear file names for easy identification
- Option to quit/cancel the operation
- Input validation with retry on invalid input

#### 3. User Input Handling
The system will handle various user input scenarios:

**Valid Inputs:**
- Single digit numbers (1, 2, 3, etc.)
- Quit commands ('q', 'quit', 'exit', 'cancel')

**Input Validation:**
- Check for numeric input within valid range
- Handle whitespace and case-insensitive quit commands
- Provide clear error messages for invalid input
- Allow retry on invalid input (up to 3 attempts)

**Error Handling:**
- Invalid numeric input: "Please enter a number between 1 and X"
- Out of range input: "Please enter a number between 1 and X"
- Non-numeric input: "Please enter a number or 'q' to quit"

#### 4. Selection Confirmation
After user selects a workspace file, provide confirmation:

**Confirmation Format:**
```
You selected: project-dev.code-workspace

Proceed with this workspace? (y/n): 
```

**Confirmation Options:**
- 'y', 'yes', 'Y', 'YES' - proceed with selection
- 'n', 'no', 'N', 'NO' - return to selection prompt
- 'q', 'quit' - exit the program
- Invalid input - ask for clarification

#### 5. Integration with Main Flow
The multiple workspace handling integrates seamlessly with the main detection flow:

**Flow Integration:**
1. Detect workspace files using `Glob()` in current directory
2. If multiple files found: call `handleMultipleWorkspaces()` for user selection
3. If exactly one file found: proceed with normal workspace detection (skip selection)
4. If no files found: fall back to single repository detection

**Error Propagation:**
- Selection cancellation: exit with error code
- Invalid workspace file: return validation error
- User exit: exit with error code

### Error Handling

#### Error Types
1. **MultipleWorkspacesError**: When multiple workspace files are found but user cancels (exits with error code)
2. **InvalidSelectionError**: When user provides invalid input (handled internally with retry)
3. **UserCancelledError**: When user explicitly cancels the operation (exits with error code)
4. **WorkspaceValidationError**: When selected workspace file is invalid

#### Error Messages
- **Multiple workspaces found**: "Multiple workspace files found. Please select one:"
- **Invalid input**: "Please enter a number between 1 and X or 'q' to quit"
- **User cancelled**: "Operation cancelled by user" (then exit with error code)
- **Invalid workspace**: "Selected workspace file is invalid: [details]"

### Testing Strategy

#### Unit Tests (WTM Package)
**Test Cases:**
1. **Multiple workspace detection**: Test with 2, 3, 5 workspace files
2. **User input validation**: Test valid numbers, invalid numbers, quit commands
3. **Selection confirmation**: Test yes/no responses and invalid inputs
4. **Error handling**: Test cancellation, invalid workspaces, etc.
5. **Edge cases**: Test with 0 files, 1 file, many files

**Mock Strategy:**
- Mock user input/output operations
- Mock workspace file validation
- Test all error conditions and edge cases

#### Integration Tests (FS Package)
**Test Cases:**
1. **File system operations**: Test `Glob()` with multiple workspace files
2. **Cross-platform compatibility**: Test on different operating systems
3. **File permission handling**: Test with read-only directories

### User Experience Considerations

#### Accessibility
- Clear, simple interface that works with screen readers
- Consistent numbering and formatting
- Support for users who may have difficulty with complex interactions

#### Internationalization
- Use simple English text that can be easily translated
- Avoid complex formatting that might not work in all locales
- Consider future i18n requirements

#### Error Recovery
- Clear error messages that explain what went wrong
- Multiple retry attempts for user input
- Easy way to cancel and exit gracefully

## Dependencies
- **Blocked by:** Feature 001 (Detect Single Repository Mode)
- **Blocked by:** Feature 002 (Detect Workspace Mode)
- **Blocks:** Feature 004 (Validate Project Structure and Git Configuration)

## Success Criteria
1. **Functional**: Successfully handles multiple workspace files with user selection
2. **User Experience**: Clear, intuitive interface that doesn't require documentation
3. **Error Handling**: Graceful handling of all error conditions
4. **Integration**: Seamless integration with existing detection flow
5. **Testing**: Comprehensive test coverage for all scenarios
6. **Performance**: Responsive interface with minimal delay

## Future Considerations
1. **Advanced Selection**: Future support for fuzzy matching or search
2. **Workspace Preview**: Show workspace contents before selection
3. **Default Selection**: Remember user's previous choice for similar scenarios
4. **Batch Operations**: Support for processing multiple workspaces
5. **Configuration**: Allow users to configure selection preferences
