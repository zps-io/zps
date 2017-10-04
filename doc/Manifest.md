Manifest
========

The package manifest will be serialized (json) version of the package data created from the Zpkgfile config file and the actions added to the package at build time by the builder.

Signatures
==========

Manifests may be signed by multiple entities. A signature section will be included in the manifest, but will be omitted when computing the hash for the manifest.