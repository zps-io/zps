Rationale
=========

ZPS is a modern binary package management solution designed with a focus on usability for developers and operators, as well as a foundation in good design and architecture.

#### Principles

- It should be easy for developers to create packages for their software
- It should be easy for developers to publish packages to repositories
- A package based workflow should support promotion across channels in repositories and rolling upgrades
- Users should not be exposed to the vulgarities of implementation infrastructure
- Support cross platform builds with variable interpolation
- We shall not invent metadata formats, formats used will be easily consumable by third parties
- OS support will be limited to OSX, Linux, FreeBSD, and Illumos (initially)
- Architecture support will be limited to amd64 (initially) and arm64 (eventually)
- System state can be modeled as a set of packages
- A package is composed of a set of actions
- Repositories should be easily discoverable
- Repositories are multivendor from the start
- Repositories should support import work flows
- A Package system should support multiple roots (install roots)
- The integrity of a system comprised of packages should be easily validated and quickly repaired
- Images should be comprised of package sets
- The Docker hype machine is the result of extended stagnation in the *NIX package management space
- Should integrate cleanly with the HashiCorp product family

#### Prior Art, and Influences

- IPS
- APT/dpkg
- Yum/RPM
- npm
- dart pub
