TODO
====

### Milestone 1

Currently in progress. Private code base is being refactored and published to github.

- Rework error handling, ensure all errors are clear, emit what go package raised them, use multi-error handling in config and Zpkgfile parsing
- Freeze support
- Uri handler repository abstraction first type should be file system path
- Static Repositories
- Proper Transaction Manager
- Package signing infrastructure
- Some shortcuts were taken that will not ensure perfect FS integrity
- Dir permissions (we aren't checking for conflicts or solving for multiple dirs in the same package)
- Action policy configuration and enforcement
- Create online docs, man pages
- Complete when ZPS can be safely used as an ancillary (non-root) package manager

### Testing

- Add tests once design is stable

### Milestone 2

- TBD