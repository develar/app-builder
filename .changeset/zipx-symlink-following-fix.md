---
"app-builder-bin": patch
---

fix: prevent symlink-following arbitrary file overwrite in zipx.Unzip (VU#293714, CVE-2026-13723). Reject symlink entries whose resolved target escapes the output directory and open regular-file writes with O_NOFOLLOW. Unsafe entries are skipped (non-fatal), consistent with node-tar and bsdtar.
