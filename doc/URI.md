ZPKG URI
========

Example: zpkg://solvent/tests/testpkg@0.0.1:20151106T023134Z

Fields:

- publisher: solvent
- category: test (may have multiple /)
- name: testpkg
- pkg: semver
- 8601 timestamp stripped punctuation

Uniqueness:

For now package name will be the unique value. This is to support layering of multiple provider repositories, as users may wish to install more recent packages from other providers. The publisher namespace will allow a single remote repository to service multiple publishers.
