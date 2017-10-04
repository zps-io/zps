Configuration
=============

Configuration will consist of static HCL, and saved values in boltdb files

## Root
```
$PREFIX/etc/zps/conf.hcl
$PREFIX/etc/zps/policy.d/*.hcl
$PREFIX/etc/zps/repo.d/*.hcl
$PREFIX/usr/bin/*
$PREFIX/var/lib/zps/zpm.db
```

## ZPM DB

Buckets

- TBD

## Image DB

Buckets

- packages
- frozen
- root request install index for automatic removal of no longer needed deps
- fs
- txlog
