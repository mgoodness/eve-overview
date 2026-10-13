---
status: proposed
---

# One account-wide union profile, not separated launcher profiles

EVE stores one overview profile per **account**, not per character, so this profile
is shared by three characters with different jobs: a cloaky J-space PvP hunter, and
two alts flying Planetary Industry, exploration and PvE. We build a **single union
profile** — Z-S Overview v10.06.09's full 58-preset library, our hunter presets
added alongside it, and one tab roster serving both audiences — rather than giving
the hunter a separate launcher profile. The union costs only extra tabs nobody is
forced to click while keeping the alts' presets intact; separation would cost a
client restart per role switch and risks cross-account settings bleed when two
accounts share a launcher profile.

## Considered Options

- **Union profile (chosen).** One profile, ~11 of the 20 available tabs, every tab
  visible to all three characters.
- **Separated launcher profiles.** The hunter gets its own launcher profile and its
  own settings folder. CCP-sanctioned, and the only way to truly split the tab
  budget — but it requires a client restart to swap roles, and bleeds settings
  across different accounts that reuse one launcher profile.
- **Per-character manual re-import.** Keep several exported YAMLs and re-import on
  character switch. Safe within one account, since only one character per account
  can be online at a time, but manual — and a mid-session mistake overwrites the
  shared record.

## Consequences

- The **tab roster** is a shared budget: ~11 of 20 slots, ordered so the hunter's
  pressure set (Travel, Engage, Exit!, Holes, Collidables) comes first and the
  job-specific tabs follow.
- The **preset library is never pruned.** An unbound preset stays selectable from
  any tab's settings and from the d-scan dropdown, so dropping a tab never removes
  a capability. This is what keeps the alts whole — it is also why the Z-S lineage
  stays additively diffable rather than forked.
- The **appearance layer** — colortag and background priority, blinks, ship labels
  — is shared and cannot be made per-character. It is where the three audiences
  genuinely conflict, and therefore the part of this decision most likely to be
  revisited.
