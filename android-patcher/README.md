# Zill Korean Beta Patcher (Android)

Android on-device patcher for **Zill O'll Infinite Plus ULJM-05410 v1.03**.

The app uses Android's Storage Access Framework and requests no broad storage or network permission. The selected retail ISO is opened read-only, copied into the app's temporary workspace, authenticated, and used to build a separate Korean Beta ISO. The original ISO is never modified.

Current beta payload embeds the repository's reviewed canonical Korean corpus, the deterministic Korean glyph catalog, and the authenticated BOOT/EBOOT/SFO patch manifests. The same Go build core used by the maintainer tooling is cross-compiled for Android arm64 and packaged as `libzill.so`; the Java UI only handles document selection, validation, progress, and result export.

Validation before build includes:

- ISO9660 primary volume descriptor
- `PSP_GAME/PARAM.SFO`
- `DISC_ID = ULJM-05410`
- `DISC_VERSION = 1.03`
- authenticated retail archive/message-bank structure
- supported EBOOT fingerprint
- current Korean slot/font requirements

The beta build currently targets the reviewed Korean overlay (42,016 accepted records) and the current deterministic 1,308-glyph catalog. Records outside the accepted Korean overlay retain their authenticated retail Japanese bytes.

## Build

```sh
gradle -p android-patcher :app:testDebugUnitTest

gradle -p android-patcher :app:assembleDebug
```

GitHub Actions workflow `.github/workflows/android-extractor.yml` prepares the embedded Korean project data, cross-compiles the arm64 Go builder, runs Go and Android unit tests, builds the debug APK, verifies the embedded native/data payload, computes SHA-256, and uploads both files as the `zill-korean-beta-patcher-debug` artifact.

Runtime PPSSPP testing remains the authority for line wrapping, clipping, glyph appearance, and branch-specific presentation.
