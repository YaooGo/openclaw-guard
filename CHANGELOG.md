# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2026-03-19

### Added
- **Security Scan Enhancement**: Upgraded from 4 to 9 comprehensive modules (MCP, Claude, PII, Shell History, etc.).
- **One-click Fix**: Automated repair for file permissions, SSH config, and Claude security settings.
- **GitHub Templates**: Added issue and pull request templates for better community collaboration.
- **CI/CD**: Added GitHub Actions workflow for automated macOS builds.
- **Root Build System**: Introduced a master `package.json` to manage backend, frontend, and electron builds.
- **Version Tracking**: Added a global `VERSION` file.

### Fixed
- **Backend**: Resolved Goroutine leaks in monitor, fixed duplicated log persistence, and corrected unique ID generation.
- **Frontend**: Fixed IPC race conditions, memory leaks in timers, and timestamp unit errors.
- **Electron**: Fixed file watcher management and path blacklist matching logic.

### Security
- Added a security disclaimer to the README.
- Hardened `.gitignore` to prevent secret leaks.
