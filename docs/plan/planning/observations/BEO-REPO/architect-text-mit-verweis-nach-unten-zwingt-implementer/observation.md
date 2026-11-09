# Ein Satz im Text des Architect verweist auf einen Slice, `make docs-check` meldet ihn, und der Implementer ändert den Architect-Text

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Spezifikation darf nach `AGENTS.md` §3.4 und der Klasse `spec-straten` keinen Slice nennen. Steht ein Slice-Name in einer Zeile, die der Architect schrieb (hier die *Ungeprüft*-Liste von `SPEC-261`), meldet `make docs-check` einen Verweis nach unten (`matrix-forbidden`), sobald der Slice-Plan in der Gate-Kette liegt. Der Implementer ist nicht der Besitzer des Textes und kann die Zeile nur ändern, um das Gate zu erfüllen; die Rolle prüft vorher nicht, ob ein Architect-Satz einen Slice nennt.
