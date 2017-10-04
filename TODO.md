TODO
====

### Cleanup

- Normalize use of action.Key(), action.Unique()

### Testing

- Add tests once current design is stable

### Milestone 1

- Switch to kingpin for cli
- Freeze support
- Uri handler repository abstraction first type should be file system path
- Static Repositories
- Proper Transaction Manager
- Package signing infra structure OCSP
- Some shortcuts were taken that will not ensure perfect FS integrity
- Dir permissions (we aren't checking for conflicts or solving for multiple dirs in the same package)
- Action policy configuration and enforcement
- Create online docs, man pages
- Rework error handling, ensure all errors are clear, emit what go package raised them, use multi-error handling in config and Zpkgfile parsing
- Figure out image path install handling
- Complete when ZPS can be safely used as an ancillary (non-root) package manager

### Milestone 2

- TBD