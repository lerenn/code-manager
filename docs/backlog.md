# Cursor Git WorkTree Manager - Backlog

- **2. Detect workspace mode (`.code-workspace` files)**
  - **Blocked by:** 1

- **3. Handle multiple workspace files with selection prompt**
  - **Blocked by:** 1, 2

- **4. Validate project structure and Git configuration**
  - **Blocked by:** 1, 2, 3

- **5. Implement `$HOME/.cursor/cgwt/repos/<repo>/<branch>/` structure**
  - **Blocked by:** 4

- **6. Handle repository name extraction and sanitization**
  - **Blocked by:** 5

- **7. Branch name validation and sanitization**
  - **Blocked by:** 5

- **8. Implement `$HOME/.cursor/cgwt/workspaces/<workspace>/<branch>/<repo>/` structure**
  - **Blocked by:** 4, 5

- **9. Workspace name extraction from `.code-workspace` files**
  - **Blocked by:** 8

- **10. Multi-repo workspace support**
  - **Blocked by:** 8, 9

- **11. Create worktrees for single repositories**
  - **Blocked by:** 4, 5, 6, 7

- **12. Create worktrees for multi-repo workspaces**
  - **Blocked by:** 4, 8, 9, 10, 11

- **13. Implement collision detection and prevention**
  - **Blocked by:** 11

- **14. Support for ephemeral worktrees (`-e` flag)**
  - **Blocked by:** 11

- **15. Support for persistent worktrees (default)**
  - **Blocked by:** 11

- **16. Safe Git state management**
  - **Blocked by:** 11

- **17. List worktrees for current project**
  - **Blocked by:** 11, 12

- **18. List all worktrees across projects (`--all` flag)**
  - **Blocked by:** 17

- **19. Human-readable output format**
  - **Blocked by:** 17

- **20. JSON output format for extensions (`--json` flag)**
  - **Blocked by:** 17

- **21. Safe deletion with confirmation**
  - **Blocked by:** 11, 12

- **22. Force deletion option (`--force` flag)**
  - **Blocked by:** 21

- **23. Proper Git state cleanup**
  - **Blocked by:** 21

- **24. Path validation and error handling**
  - **Blocked by:** 21
