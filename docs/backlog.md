# Git WorkTree Manager - Backlog

- **12. Integrate IDE opening functionality (`-o` command)**
  - Support opening repositories in different IDEs (VS Code, GoLand, IntelliJ, etc.)
  - Command format: `wtm create [branch-name] -o <ide-name>`
  - IDE detection and validation
  - Cross-platform support (macOS, Linux, Windows)

- **13. Create worktrees for multi-repo workspaces**
  - **Blocked by:** 4, 8, 9, 10, 11

- **14. Implement collision detection and prevention**
  - **Blocked by:** 11

- **15. When config doesn't exists, then copy the default one in place (should be embedded from configs/default.yaml)**

- **16. Support for persistent worktrees (default)**
  - **Blocked by:** 11

- **17. Safe Git state management**
  - **Blocked by:** 11

- **18. List worktrees for current project**
  - **Blocked by:** 11, 13

- **19. List all worktrees across projects (`--all` flag)**
  - **Blocked by:** 18

- **20. Human-readable output format**
  - **Blocked by:** 18

- **21. JSON output format for extensions (`--json` flag)**
  - **Blocked by:** 18

- **22. Safe deletion with confirmation**
  - **Blocked by:** 11, 13

- **23. Force deletion option (`--force` flag)**
  - **Blocked by:** 22

- **24. Proper Git state cleanup**
  - **Blocked by:** 22

- **25. Path validation and error handling**
  - **Blocked by:** 22
