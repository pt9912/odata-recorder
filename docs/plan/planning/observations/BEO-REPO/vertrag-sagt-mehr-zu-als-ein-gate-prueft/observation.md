# Ein Vertrag sagt mehr zu, als das Gate prüft

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Vertrag eines Werkzeugs (`harness/vertraege/`) nennt eine Eigenschaft des Ergebnisses (Version aus einem Build-Argument, `-trimpath`), aber keine Mutation macht ein Gate rot, wenn sie entfällt; die Unit-Tests berühren die Stelle nicht.
