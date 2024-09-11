# app-builder-bin

## 5.0.0-alpha.8

### Minor Changes

- [#130](https://github.com/develar/app-builder/pull/130) [`df4f272`](https://github.com/develar/app-builder/commit/df4f27286a92b6fa17dd333abbdca9d53c8fc1cb) Thanks [@tisoft](https://github.com/tisoft)! - Added support for OpenSUSE to rpm

### Patch Changes

- [#132](https://github.com/develar/app-builder/pull/132) [`1092684`](https://github.com/develar/app-builder/commit/1092684f6771af6abe3ef5614f6136000858003d) Thanks [@beyondkmp](https://github.com/beyondkmp)! - fix: find the real parent node module

## 5.0.0-alpha.7

### Patch Changes

- [#126](https://github.com/develar/app-builder/pull/126) [`f910175`](https://github.com/develar/app-builder/commit/f9101753dd2b93b857864d4051baeb6d8856dd64) Thanks [@mmaietta](https://github.com/mmaietta)! - fix: to resolve appimage issues in electron builder, and since we can't update electron-builder-binaries repo, we should just downgrade to the last working version of appimage

## 5.0.0-alpha.6

### Patch Changes

- [#124](https://github.com/develar/app-builder/pull/124) [`52ad062`](https://github.com/develar/app-builder/commit/52ad0626206c3ff7b7170afabe2136ef97107042) Thanks [@mmaietta](https://github.com/mmaietta)! - fix: set correct compression enums and remove default

## 5.0.0-alpha.5

### Patch Changes

- [#123](https://github.com/develar/app-builder/pull/123) [`20feb29`](https://github.com/develar/app-builder/commit/20feb293f5fa2dc46c4e52212ec9e17e6db669a0) Thanks [@mmaietta](https://github.com/mmaietta)! - fix current mksquashfs version only allows xz and gzip compressions

- [#118](https://github.com/develar/app-builder/pull/118) [`94485c6`](https://github.com/develar/app-builder/commit/94485c6d500fda34b92a6b4e0ef8314d2cc1a88d) Thanks [@fabienr](https://github.com/fabienr)! - fix: hoist dependencies to the real parent in nodeModuleCollector

## 5.0.0-alpha.4

### Patch Changes

- [#119](https://github.com/develar/app-builder/pull/119) [`6a940e4`](https://github.com/develar/app-builder/commit/6a940e46da11d733f8b7c6f31b183c0e402882aa) Thanks [@beyondkmp](https://github.com/beyondkmp)! - fix: alias name issue in node modules resolution dependency tree

- [#120](https://github.com/develar/app-builder/pull/120) [`189519a`](https://github.com/develar/app-builder/commit/189519a8292f939d9e5d3b47c6407444fee70334) Thanks [@beyondkmp](https://github.com/beyondkmp)! - change node module symlink to real path

## 5.0.0-alpha.3

### Minor Changes

- [#116](https://github.com/develar/app-builder/pull/116) [`be4e7ec`](https://github.com/develar/app-builder/commit/be4e7ec9c438e7f803c120a66148950ba294dae5) Thanks [@beyondkmp](https://github.com/beyondkmp)! - feat: add flatten option to `node-dep-tree` for rendering dependency conflicts in a different manner

## 5.0.0-alpha.2

### Patch Changes

- [#113](https://github.com/develar/app-builder/pull/113) [`43f7a34`](https://github.com/develar/app-builder/commit/43f7a3473cfbbefc5eba03f7fb04f88f54a1adf2) Thanks [@mmaietta](https://github.com/mmaietta)! - fix: revert appimage 13.0.1 to 13.0.0 due to mksquash arch compilation issues

## 5.0.0-alpha.1

### Minor Changes

- [#109](https://github.com/develar/app-builder/pull/109) [`e53b84c`](https://github.com/develar/app-builder/commit/e53b84c9a36105f281825a6e6d168481ddf543a9) Thanks [@mmaietta](https://github.com/mmaietta)! - feat: allow providing env var for custom app-builder binary as opposed to accessing directly from the PATH env var

### Patch Changes

- [`64bb497`](https://github.com/develar/app-builder/commit/64bb4971150edc37dbfb3819f115e4d767cf89c6) Thanks [@mmaietta](https://github.com/mmaietta)! - fix(snap): Parse user command line options as last values

## 5.0.0-alpha.0

### Major Changes

- [#107](https://github.com/develar/app-builder/pull/107) [`f4642dd`](https://github.com/develar/app-builder/commit/f4642ddcd85b482d1a7ed49f14d27c509eb5aa6b) Thanks [@mmaietta](https://github.com/mmaietta)! - chore: changing repo structure for release automation

### Minor Changes

- [#98](https://github.com/develar/app-builder/pull/98) [`3ed22df`](https://github.com/develar/app-builder/commit/3ed22df75fcff132a5b794ce1a421bec263bc118) Thanks [@yzewei](https://github.com/yzewei)! - feat: Add loongarch64 support

### Patch Changes

- [#106](https://github.com/develar/app-builder/pull/106) [`9704964`](https://github.com/develar/app-builder/commit/970496449b0b02780d654d61af1e3277515a2545) Thanks [@theogravity](https://github.com/theogravity)! - fix: Use npm config.mirror first before env variables for download URL
