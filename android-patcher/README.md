# Zill Font Extractor (Android)

Read-only first-stage extractor for **Zill O'll Infinite Plus ULJM-05410 v1.03**.

The app uses Android's Storage Access Framework. It requests no broad storage permission and no network permission. The selected ISO is opened with mode `r` and is never modified.

Validation before extraction:

- ISO9660 primary volume descriptor
- `PSP_GAME/PARAM.SFO`
- `DISC_ID = ULJM-05410`
- `DISC_VERSION = 1.03`
- `PSP_GAME/USRDIR/pa.bin` / `pa.arc`
- PAA member count = 14,231
- member #13611 = `font/zillfont.par`, offset `0x3D8F510`, size `0x80470`
- member #13612 = `2d/font/jillbtn.par`, offset `0x3E0F980`, size `0x18E60`

Output ZIP:

```text
font/zillfont.par
2d/font/jillbtn.par
manifest.json
```

The 16-byte PAA member wrapper is stripped from each exported PAR. `manifest.json` records the target/version, source ISO filename and size, PAA metadata, and SHA-256 for each exported PAR.

## Build

```sh
gradle -p android-patcher :app:testDebugUnitTest

gradle -p android-patcher :app:assembleDebug
```

GitHub Actions workflow `.github/workflows/android-extractor.yml` performs the tests, builds the debug APK, computes its SHA-256, and uploads both as the `zill-font-extractor-debug` artifact.
