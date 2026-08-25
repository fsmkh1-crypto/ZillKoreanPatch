# Upstream provenance

The builder baseline for this repository is:

- upstream: `https://github.com/HK47196/zill`
- pinned commit: `a98d9ce29f361d666ec23da0dcfd351f24537ffd`
- upstream commit message: `Add bidirectional Trinity text search`
- imported into `main` by merge commit: `f35a0927dd4d6d9dccf140985a5836c7f9841e67`

The import commit has the pre-existing `ZillKoreanPatch` history as its first parent and the pinned upstream commit as its second parent. Upstream source, license, notice, and translation-license files were taken from the pinned upstream tree; Korean-project-owned files were overlaid afterward.

For a normal local clone, add the upstream remote explicitly because Git remotes are local configuration and cannot be stored in commit history:

```sh
git remote add upstream https://github.com/HK47196/zill.git
git fetch upstream
```

Do not replace or repurpose `fsmkh1-crypto/ZillOcrOverlay`; it is a separate real-time OCR translation project.

## Licensing

Preserve the upstream `LICENSE`, `NOTICE.md`, and `LICENSES/` files. Upstream builder code is GPL-3.0-or-later. Translation content has its own Creative Commons attribution/share-alike terms as recorded by the upstream project; derived translation distributions must preserve the applicable attribution and license notices.
