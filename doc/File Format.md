ZPKG File Format
================

### WHY?

No one has created an indexed archive format that isn't encumbered with legacy garbage.

### HEADER

magic number:         string  6 bytes           zpkg66
version:              uint8                     1
manifest length:      uint32                    length of manifest

### MANIFEST

manifest              bytes[] manifest.length   HCL(json interop) encoded manifest

### PAYLOAD

xz indexed streams currently, after reading xz critiscm paper will probably switch to bzip2
