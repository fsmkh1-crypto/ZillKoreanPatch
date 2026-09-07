# Batch 020 direct contextual review

Baseline: `c9c1f8e0660adbe3a603391ee5ab18ea1a53b6d5`, `milestone/Beta1`.
Explicit user execution signal received: YES (continue Beta1 and execute020).
AGENTS, CONTRIBUTING, Korean style/QA protocol, review decision, v3 policy and
current queue/scope workflows read. Protected branches excluded. No blind edits.
Pinned English: `a98d9ce29f361d666ec23da0dcfd351f24537ffd`.

Direct review order was Japanese meaning and adjacent state/record context,
pinned English treatment (reason inferred from its wording), then Korean.
All84 rows were displayed in the deterministic packet and read, including every
state. The renderer is the unmodified `build-beta1-review-packet.py:render`;
inputs were fetched at exact SHA through GitHub, not an asserted local checkout.
Scope CI independently reconstructs from Git history before granting KEEP.

## Scope and decisions

- `msgsec010-part99.toml`:100000–100018,19 IDs.
- `msgsec011-part99.toml`:110000–110033,34 IDs.
- `msgsec012-part99.toml`:120000–120005,6 IDs.
- `msgsec013-part99.toml`:130000–130024,25 IDs.
- Total84; EDIT27; KEEP53; internal markers4; contextual80; propagated0.
- Approximate visible-state workload281 (split at select/if/end and count
  Hangul-bearing pieces, excluding the four terminal markers). This is a
  workload estimate, not an authenticated runtime parser count or coverage.
- All scope JP matches pinned English JP. None has persisted layout at baseline.

Exact edits and individual JP/EN reasoning are preserved in manifest020.
Main semantic corrections: child as killing target110011; fulfillment110004;
puzzlement120004; island city130004; idiomatic no-harm130006.
Grammar corrections include100006 (player struggles),110001,110016,110023,130009.
Same-entity Dyneskal normalization is restricted to seven directly read IDs;
Underground Road label130003 is matched to guide/choices130010–130013.
No corpus-wide terminology completion is inferred from these limited changes.

## KEEP review notes

These notes summarize actual direct reads; they are not automated decisions.
The exact IDs and complete strings remain reconstructable in scope020's packet.

| IDs | Context and KEEP reason |
| --- | --- |
|100000|Three spirit/Dragon King explanations preserve cause and fear; polite register retained.|
|100001|Elf praises human/spirit harmony and the Thousand-Year Tree; forest departure motive and casual voice intact.|
|100007|Four equipment/element tips preserve the JP mechanics; no speculative mechanic correction from real-world intuitions.|
|100008|Mercenary cohesion, dwindling Mizuchi, then famous adventurer; clauses and fixed operands preserved; EN reorders name/epithet but KO retains source order.|
|100010|Lake creation, island location and free passage all correspond to JP/EN; this also supplies context for130004.|
|100011|Xenetes recommendation, mercenary experience and recognition at fame threshold retain speaker voice.|
|100012,100013|Departure/boarding instructions and free fare agree with JP/EN.|
|100014–100017|Leave/go/not-ready choices retain their distinct destinations; no taste-only punctuation homogenization.|
|110000|Elder's reflections on death/war retain lamentation and intentional repeated states.|
|110003|Two-person watch and peaceful town; shy humor and hypothetical recruitment preserved.|
|110005|Child's captain aspiration and training commands are consistent with following sibling row.|
|110007|Washer Pirates' old code, Hugo dispute and sea escape limitation retain all substantive points.|
|110008|Crystal-ball knowledge, children, flowers and temple road warning retain informal speaker voice.|
|110009|Dwarf cannot swim; horizon/post detail, return reluctance and interest in Er remain accurate.|
|110010|Wind/flowers/daily anticipation and priestess location preserve the cheerful characterization.|
|110012|Lilubee sings and circles bed; repetitive wording is source characterization, not accidental duplication.|
|110015|Wind/Earth relation, bow and forbidden spell clues retained; marked elder register preserved.|
|110017|Level-dependent mountain warning/admission and Er-permission route remain distinct.|
|110018|Old woman's upstairs refusal retained, including source-aware `%4` adjoining visible `2층`; no greedy numeric rewrite.|
|110021,110022|Mother/child city-versus-Elz arc read together; post-Liberdam change of mind and repeated final states retained.|
|110025|Boldan physical-work pride and rival's effort retain rough casual voice.|
|110027|Sea King sighting, statue, three dragons and player recognition preserve source distinctions.|
|110028,110029|Wife's ambition/workload and husband's trade-route/business arc align; no inference that they are unrelated speakers.|
|110030–110032|Paid voyage departure and stay/leave choices align.|
|120000|Ordered chambers left-to-right match both sources.|
|120001|First outside visitor and Esther-inspired adventure wish retain speaker perspective.|
|120002|Pillar destinations and ordinal directions individually checked against JP/EN.|
|120003|Water's value second to life and conservation instruction retained.|
|130002|Outsider non-aggression rules, Dergado spear and protective use of weapons retain argument.|
|130007|Blacksmith's repeated opening and traveling apprenticeship/shop story preserve abrasive voice.|
|130008|Weapons/armor deaths and drunken complaint are intentional emotional repetitions.|
|130010–130023|Guide's route names, choices, departure, waiting and return prompts read as one sequence; destinations and politeness preserved.|

## Exclusions

100018,110033,120005,130024 are source-identical section-ending message markers;
pinned EN is blank. They were read but are not credited as contextual prose.
Exclusion affects only this scope's language coverage, not corpus membership.

## Safety and validation

English patch parity checked: YES. Ordinary semantic text remains separate from
build-owned wrapping; fixed controls, operands and substitutions remain unchanged.
Korean divergence: none introduced. No renderer/font/parser/workflow changes.
Local exact-before/apply/after, JP preservation, control/numeric-literal sequence,
and absence of semantic line-breaks checked. Heavy queue must pass effective-layout,
glyph/font/data and pinned-English storage gates; light scope must independently
reproduce packet and rebuild generated coverage. Final results belong in throughput
and the checkpoint, not guessed here. No retail/emulator/PSP runtime evidence added.
