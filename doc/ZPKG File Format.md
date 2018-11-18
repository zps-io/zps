ZPKG File Format
================

### WHY?

There was no preexisting indexed archive file format that satisfied ZPS development goals.

### HEADER

magic number:         string  6 bytes           zpkg66

version:              uint8                     0

compression:          uint8                     0 (bzip2)

manifest length:      uint32                    length of manifest

### MANIFEST

manifest              []byte manifest.length   JSON encoded manifest

### PAYLOAD

bzip2 indexed streams