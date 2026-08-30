# A-070 — English-first fixed-data parity

Status: **PASS for storage safety; MISSING for Korean equipment localization**

## Scope

Audit engine-owned fixed strings outside message banks against the upstream English patch:

- patched executable fixed-width strings;
- `data/bindata.dat` equipment-name table.

The purpose is to separate engine/storage safety from localization completeness.

## EBOOT result

English owns a complete table of 557 guarded fields. Korean intentionally owns a sparse reviewed overlay.

This difference is acceptable only because every Korean field uses the same engine-facing contract as English:

- authenticate the manifest-patched ELF by SHA-256;
- authenticate the exact retail source bytes at the declared offset;
- require a source NUL terminator;
- encode through the installed Korean renderer-slot mapping;
- reject replacement bytes larger than the original field;
- reject overlapping fields;
- clear only the original fixed-width span and copy the replacement;
- after the sparse Korean overlay, re-run `elfpatch.VerifyApplied` so localization cannot clobber runtime executable patches.

Classification: **DIFFERENT-BY-DESIGN (sparse Korean coverage), engine/storage contract PASS.**

## BINDATA/equipment result

The English implementation proves the table contract in `internal/fixeddata/equipment.go`:

- authenticated retail `bindata.dat` fingerprint;
- 132 records;
- record stride `0x24`;
- name offset `0x11`;
- name field size 17 bytes;
- therefore at most 16 encoded payload bytes plus NUL;
- exact source-string guard for every selector;
- `0xCD` padding guard after the NUL.

English then materializes all 132 translated names through `addEquipment`.

Korean desktop/mobile currently **do not mutate these 132 equipment-name fields**. They remain retail Japanese. This is not a hidden fixed-buffer violation: the production planners load the authenticated BINDATA, execute `fixeddata.ApplyEquipment` as a structural/source guard, scan its CP932 literals, and reserve those live renderer keys before Korean glyph allocation.

Classification:

- engine/storage safety: **PASS**;
- renderer-key ownership: **PASS**;
- Korean localization completeness for the 132 equipment names: **MISSING**.

The missing localization should not be "fixed" by applying the English strings to the Korean build. A Korean equipment authority must be reviewed separately and then materialized through the existing 17-byte fixed-field contract.

## Gate

`tools/korean/audit-english-first-fixed-data.py` now fails closed if:

- Korean EBOOT loses the English fixed-width/source/NUL/overlap contract;
- executable-manifest postverification disappears;
- the authenticated 132×17 BINDATA contract drifts;
- Korean accidentally starts applying the English equipment strings;
- desktop/mobile planners stop authenticating BINDATA or reserving its structured CP932 renderer ownership.

## Freeze relevance

This closes another static storage/corruption surface. The remaining Korean equipment issue is a visible localization-completeness defect, not evidence of a freeze mechanism.
