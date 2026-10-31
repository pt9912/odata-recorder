# Felder derselben Referenz tragen in der Fixture denselben Wert, und eine Mutation, die das falsche Feld nimmt, bleibt grün

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Eine Fixture setzt mehrere Zeichenketten- oder Zeitfelder derselben Referenz auf denselben Wert; eine Mutation, die ein Feld gegen ein anderes tauscht, ändert dann nichts. `AGENTS.md` §3.15 sagt "Eine Fixture gleicht nie dem Wert der Mutation"; weder `.claude/commands/implement-slice.md` noch `.harness/skills/reviewer.md` noch ein Rollen-Typ nennen das Wort *Fixture*. Nahe an `fix-deckt-die-genannte-mutation-nicht-die-klasse`.
