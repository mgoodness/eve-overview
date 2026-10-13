---
status: proposed
---

# The militia States keep their appearance, but lose the militia Blink

The reference pack blinks **State 44** — *at war with your militia* — alongside six
others (`{9, 13, 44, 50, 51, 52, 68}`). The spec's Implementation Decisions reduce
Blinks to the three dangerous States: "Blinks reduce to war target, criminal and low
security status." That reduction drops 44, while the same spec's acceptance criteria
for the appearance layer require that "The militia States keep exactly the treatment
the reference pack gives them, so lowsec behaviour is unchanged for the other
characters" and separately that "Blinks are set only for war target, criminal and low
security status." Read literally the two pull apart: a Blink *is* part of a State's
treatment, so dropping 44 does change something about the militia State.

## Decision

The Blink reduction wins, and its blast radius is confined to State 44's Blink. The
militia States keep the rest of the pack's appearance treatment **unchanged**:

- **Colortag whitelist and priority order**: inherited from the pack verbatim,
  including 44 and 45. This is also required on its own by the appearance ticket.
- **Background whitelist, priority and colour name**: 44 and 45 retain the pack's
  membership, order and `stateColorsNameList` colour. Neither is one of the four
  information-free States whose Background is stripped.
- **`stateColorsNameList`**: the pack's entry for `background_44` (red) is untouched.

So the only lowsec-facing behaviour that changes is that a militia war target no
longer blinks. It still carries its pack Colortag and its pack Background, so a
lowsec pilot's row continues to render exactly as it does today apart from the
animation. The reduction is deliberate: a Blink is meant to be rare enough that it
draws the eye, and the authored profile's Blink set is the war/criminal/lowsec trio
the spec names.

## Considered Options

- **This resolution (chosen).** Reduce Blinks to `{9, 13, 51}`; keep every other part
  of the militia States' pack treatment. Lowsec behaviour changes only in the loss of
  the 44 Blink.
- **Keep 44 blinking to honour "exactly the pack's treatment".** Rejected as a
  misreading: the spec's Blink line is explicit and enumerates the set, so honouring
  it means 44 cannot blink. Keeping 44 would make the Blink set four States, not the
  three named.
- **Drop 44 and 45 entirely.** Rejected: it would discard the pack's Colortag and
  Background treatment and put real lowsec behaviour — not just an animation — out of
  step with the other characters, which is precisely what the acceptance criterion
  forbids.

## Consequences

- The reduced Blink set is a deliberate deviation from the pack, not an omission; the
  appearance assertions compare the emitted Colortag and Background layers against the
  pack, and the Blink set against the spec's three States, so the deviation is pinned
  in both directions.
- A future reader comparing the profile to Z-S will find exactly one intentional
  difference in the militia States: no Blink on 44.
- If the pack's militia treatment is ever revisited, this is the ADR that stops the
  author from "restoring" the 44 Blink on the assumption it was dropped by accident.
