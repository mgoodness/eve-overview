---
status: proposed
---

# A ten-tab roster over Z-S's library, with four authored presets

The profile binds **10 tabs** against Z-S Overview v10.06.09's 58-preset library:
the hunter's pressure set (Travel, Engage, Exit!, Holes, Fleet), the two alts' jobs
(PvE, PI, Exploring), and two shared tabs (Loot, Occupancy). Four presets are
authored or modified rather than bound — `Engage`, `Holes`, `Occupancy`, and a copy
of `PvX: Exploring` carrying the decloak hazards. Everything else is a binding of an
untouched Z-S preset.

| # | Tab | List preset | Audience |
| --- | --- | --- | --- |
| 0 | Travel | `✈ Travel (-Ship +Sta)` | all |
| 1 | Engage | *authored* — 55 ship groups + 6 celestials + 6 decloak hazards | hunter |
| 2 | Exit! | `➲ Extra: Podsaver (!)` | all |
| 3 | Holes | *authored* — `Wormhole` (988) | hunter |
| 4 | Fleet | `✪ = Friendly: Fleet` | corp ops + NPSI |
| 5 | PvE | `✜ PvX: Mission Filter (+LCO +LCS)` | alts |
| 6 | PI | `✜ PvX: PI` | alts |
| 7 | Exploring | *modified* `✜ PvX: Exploring` + 4 hazards + Planet/Moon/Belt | alts |
| 8 | Loot | `✜ PvX: Salvage/Loot` | all |
| 9 | Occupancy | *authored* — Moon + Mobile Warp Disruptor | hunter |

## Considered Options

- **This roster (chosen).** 10 of 20 tab slots; Z-S's library intact.
- **Adopt Z-S's shipped 7-tab layout and append.** Rejected: three of Z-S's seven
  tabs are literal placeholders (`3`, `4`, `5`, `Add more ->`), and its `Main` tab
  is the same 524 groups as `PvX: Ratting` under a different name.
- **Prune to a lean hunter-only set.** Rejected: the alts depend on the library, and
  pruning destroys the diff against upstream, making a Z-S update a merge.
- **One tab for everything.** Rejected on §5.2: a list carrying ships, celestials and
  scenery at once cannot be read under fire.

## Binding rules

- A tab's **list** and **bracket** presets are chosen independently. Every tab binds
  `⌘ Brackets: Combat/Targets/Friendly (+Dro)` except `Exit!`, which binds
  `⌘ Brackets: Warpout (!)`.
- **`✜ = CovOps: Collidables` is never bound.** At 645 distinct groups it is a
  superset of the shooting list plus every rat faction, asteroid and billboard. Its
  useful members — `Large Collidable Object`, `Large Collidable Structure`,
  `Control Tower`, `Force Field` — are folded into `Engage` and `Exploring` instead,
  and the standalone preset stays in the library for deliberate sweeps.
- **The tab you are in must carry warp-to targets and hostiles simultaneously**,
  because d-scan is aimed by clicking overview rows and a wormhole's subsequent
  bookmark must be made from the overview while on grid (§5.2).

## Consequences

- 10 of 20 tab slots used, and all three characters see all ten — the cost the union
  profile accepted (ADR-0001).
- The library survives intact, so an unbound preset is still selectable from any
  tab's settings and from the d-scan dropdown.
- **Four authored presets mean a Z-S release requires regeneration, not re-editing.**
  This is the concrete case for the generator chosen in the design's first decision.
