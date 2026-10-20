---
"app-builder-bin": patch
---

Rebuild all distributed binaries with Go 1.26.10 to include the crypto/x509 security fix for compromised certificate authority keys. Pin the release toolchain and log each binary's Go build metadata.
