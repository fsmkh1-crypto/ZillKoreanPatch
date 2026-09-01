# U6 stabilization closure

## Purpose

This note consolidates the error fixes that belong to making U6 build and run correctly. They are not promoted into U7/U8-style milestones merely because they were discovered later in the forensic cycle.

The rule for this closure is evidence-based:

- a defect with an established failing mechanism and a justified repair is applied under U6;
- a warning or runtime hypothesis without an established repair contract remains evidence/QA scope and is not mutated merely to make counts disappear.

## Applied U6 fixes

### 1. Historical PR14 diagnostic could block the mobile build

Established failure: the Android payload could omit the historical H0 fixture and `auditPR14HistoricalPolicies` could propagate that diagnostic failure as a fatal build error before the production planner ran.

Current U6 state:

- Android payload packaging includes both `pr14-eboot-h0.toml` and `pr14-eboot-full.toml`.
- payload verification requires both fixtures;
- PR14 H0/B/A/Combined replay is explicitly diagnostic-only;
- an unavailable historical replay emits `FORENSIC PR14_POLICY_AUDIT_UNAVAILABLE ... build_blocked=false` and does not bypass or weaken the current production planner or downstream safety gates.

This is a build-path repair, not evidence that a PSP runtime freeze cause was identified.

### 2. Captured retail scanner hardening was applied beyond its authenticated consumer scope

Established failure mechanism: the broad scanner-derived storage wrapper could mutate C5 records even though the captured scanner evidence did not establish that every message consumer was NUL-scanner owned. On authenticated single-page C5 records, newly introduced line breaks could advance the three-line C5 page cursor into a second page. The correct C5 validator then rejected that derived state.

Current U6 state:

- production uses `DeriveKoreanC22RetailScannerLayouts`;
- scanner derivation iterates only authenticated `C22IDs`;
- within C22 it additionally requires authenticated retail source bytes containing a NUL terminator compatible with the captured scanner model;
- non-NUL C22 records are excluded rather than forced through a false universal scanner contract.

This removes the cross-consumer mutation instead of weakening later validators.

### 3. Historical 9-message / 15-branch single-page C5 failure

The exact affected IDs were:

`1280007, 1280008, 1280012, 1280017, 1280020, 1280021, 1280043, 1280050, 1280051`.

All nine are authenticated upstream-English `single_page_c5_ids`. They are not C22 scanner IDs. Their previous `2 pages > maximum 1` population is therefore handled by the scanner-scope repair above, not by shortening Korean text or relaxing `maxPages = 1`.

U6 now adds a regression gate in `internal/layout/retail_scanner_scope_test.go` that hard-asserts both facts for the exact historical population.

### 4. Message 210065 build-local C22 layout/materialization gap

The previously established build-time payload problem for message `210065` is already repaired on the current line:

- the production projection uses the established eight-line build-local layout;
- the materialized record preserves the terminating control;
- the derived storage remains below the captured `0x100` inline-span boundary;
- canonical Korean semantic text is not rewritten merely to encode the storage layout.

The remaining historical question about what an older H1 execution actually consumed is runtime provenance, not an unimplemented repair.

### 5. Release-path parity defects found during the English-first audit

The current U6 line also contains the already-landed parity repairs that removed divergent Korean-only production paths where upstream English already supplied the authority. These include the shared consumer-storage contract chain, shared ISO rebuild/failure policy, fixed-data discovery, reference font/metrics authority, raw-font and DATA.BIN paths, Android embedded project parity, and shared bank assembly.

These are retained as U6 stabilization. They are not split into new user-visible U-stages.

## Items deliberately not mutated

The following are not deferred known fixes; they remain evidence boundaries because the required runtime contract or failure mechanism is still unproven:

- unbounded runtime substitutions other than independently bounded cases such as player name `$28`;
- `$15` maximum runtime expansion and its historical freeze relationship;
- the 13 item-description single-line warnings where upstream English does not establish an automatic reflow repair;
- generic authoring-ceiling warnings that do not survive into a proven consumer/runtime failure;
- runtime invalid-pointer / allocator / return-provenance hypotheses that narrow investigation but do not yet establish a safe code mutation.

Leaving these unchanged is not postponing an established fix. It avoids converting warning populations or hypotheses into speculative payload changes.

## U6 save rule

The final green stabilization commit is the U6 target. No U7/U8 milestone is required for the fixes above. Successful CI or one successful device run remains non-reproduction evidence only; any later freeze/crash remains strong failure evidence and must be investigated on its own machine-state evidence.
