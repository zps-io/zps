TODO
====

### Milestone 1

Currently in progress.

- Rework error handling, ensure all errors are clear, emit what go package raised them, use multi-error handling in config and Zpkgfile parsing
- Proper Transaction Manager
- Package signing infrastructure
- Some shortcuts were taken that will not ensure perfect FS integrity
- Dir permissions (we aren't checking for conflicts or solving for multiple dirs in the same package)
- Action policy configuration and enforcement
- Create online docs, man pages
- Create cryptographic verification implementation
- Complete when ZPS can be safely used as an ancillary (non-root) package manager

### Testing

- Add tests once design is stable

### Milestone 2

- TBD