# Korean raster catalog

`glyphs.toml` is a generated, reviewable 10x10 4bpp raster catalog. The source font binary is intentionally not stored in this repository.

The maintained generator is `zill korean-font-generate`. The GitHub Actions workflow `.github/workflows/korean-raster.yml` fetches `UnDotum.ttf` from the pinned `deepin-community/fonts-unfonts-core` commit `ef6c68f4f0beb2a03fcf460aa643d73389aac59a`, uses the repository-owned render rule, and uploads only the generated TOML catalog as an artifact.

When Korean message overlays change, regenerate the catalog and commit the resulting `glyphs.toml` together with the translation batch so `zill korean-font-check` remains green.
