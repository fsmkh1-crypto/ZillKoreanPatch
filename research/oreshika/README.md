# Oreshika PSP Korean Patch – Hangul Input Research

This directory isolates research on the Korean fan patch for **俺の屍を越えてゆけ (Ore no Shikabane wo Koeteyuke)** from the Zill O'll patch project.

## Target observed by the user

- Platform: PSP / PPSSPP
- Game ID: `UCJS-10117`
- Game-reported version: `1.01`
- Patched ISO CRC reported by PPSSPP: `7179FB43`
- Korean patch currently indexed publicly: `2.0 Final`
- Public index date: `2026-08-01`
- Public index distributor/author label: `레갤러`

## Research goal

Identify how the Korean patch implements Korean surname/given-name entry, especially its custom on-screen keyboard and Hangul composition behavior.

## Current status

The existence and development history of the Korean name-input patch are verified from public posts and the RetroDB index. The currently indexed patch binary itself has **not yet been recovered from an accessible public download endpoint** during this research pass. An August 1 community post explicitly reports that the previous distribution link had been deleted; RetroDB lists a new `2.0 Final` record on the same date.

For that reason, this branch does **not** contain an unverified binary or any game/ISO data. It records only public research evidence and a binary-analysis plan. If the exact patch package used to produce CRC `7179FB43` is obtained later, it should be added only if redistribution is permitted; otherwise store hashes/diffs and analysis, not copyrighted game data.

See:

- `ANALYSIS.md` — preliminary technical analysis and hypotheses
- `SOURCES.md` — public evidence and source URLs

## Branch isolation

Research branch: `research/oreshika-korean-input`

Nothing here should be merged into Zill O'll localization work without an explicit decision.
