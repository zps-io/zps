ZPS
===

ZPS is a modern binary package management solution created and designed to solve the delivery and design problems of existing solutions.

#### Principles

- Creating binary software packages should be easy
- Publishing binary software packages should be easy
- Repositories should support channels for configurable upgrade work flows
- Vulgarities of technical implementation should not impact the user
- Must support cross platform builds with variable interpolation
- No invention of custom serialization formats
- OS support will be limited to OSX, Linux, FreeBSD (initially)
- Architecture support will be limited to amd64 (initially) and arm64 (eventually)
- System state can be modeled as a set of packages
- A package is composed of a set of actions
- Repositories should be easily discoverable
- Repositories are multivendor from the start
- Repositories should support import work flows
- A Package system should support multiple roots (install roots)
- The integrity of a system comprised of packages should be easily cryptographically validated and quickly repaired
- Configuration management is a delivery vehicle for unfinished work (ZPS will handle configuration at install time)
- The Docker hype machine is the result of extended stagnation in the *NIX package management space

#### Prior Art, and Influences

- IPS
- APT/DPKG
- YUM/RPM
- npm
- pub

#### Current State

M1 refactor is in progress, M1 can be considered a functional prototype. See GH issues for milestone information.

Copyright 2018 Zachary Schneider
