ZPS
===

ZPS is a modern binary package management solution designed to solve the delivery problems encountered
by teams that execute rapidly as well as the compliance problems of those that rely on the software delivered.

#### Design

- No invention of custom serialization formats
- System state can be modeled as a set of packages
- A package system should support multiple roots (install roots)
- A package is composed of a set of actions
- The integrity of a system comprised of packages should be easily cryptographically validated and quickly repaired

- Versioning should be easy to automate

- Repositories should support channels for configurable upgrade work flows
- Repositories should be easily discoverable
- Repositories are multi-vendor from the start
- Repositories should support import work flows

- Packages may be automatically added to channels based on defined metadata queries

#### Platform and Architectures

- Architecture support will be limited to x86_64 (initially) and arm64 (eventually)
- OS support will be limited to OSX, Linux, FreeBSD (initially)
- Must support cross platform builds with variable interpolation

#### Principles

- Software build systems and software delivery are separate concerns, one must not adopt a build system to create a package
- Creating binary software packages should be easy
- Publishing binary software packages should be easy
- Design should support commercial software subscriptions thereby encouraging vendors to provide binary packages
- Vulgarities of technical implementation should not be exposed to the user
- Configuration management is a delivery vehicle for unfinished work (ZPS will handle configuration at install time)
- Design and functionality should not be crippled in order to support a zero value business model
- The Docker hype machine is the result of extended stagnation in the *NIX package management space

#### Prior Art, and Influences

- IPS
- APT/DPKG
- YUM/RPM
- npm
- pub
- libsolv

#### Current State

M1 has been recently completed. M1 can be considered to be of functional prototype quality. Generally it should be fine
for use, however breakage will occur and there are likely bugs.

See GH issues for milestone information.

#### Thank You

- HashiCorp and Martin Atkins for HCL2
- James Nugent and Paul Stack for design validation and moral support

Copyright 2020 Zachary Schneider
