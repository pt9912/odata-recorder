# Die Spezifikation beschreibt einen Antworttyp, den ihr eigener Recorder verwirft

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Zwei Stellen der Spezifikation widersprechen sich über die Ebenen hinweg: die Beschreibung von `/$count` nennt `text/plain`, die Typregel des Redaktors (`SPEC-118` Nr. 3, `SPEC-046` Nr. 3) erklärte jeden `text/*` für unbekannt, und die Aufzeichnung endete in `typ`, `failed` und im Replay in `ODR-E5008`. Das Review zu `slice-m2-referenzdienst` fand es (Befund S3); kein Plan und kein Test hatte einen Text-Fall, und keine ADR schloss `text/*` aus.

## Benannt, nicht gezählt

Der Vorgang `slice-m2-referenzdienst` ist noch nicht abgeschlossen und trägt deshalb keine Evidence-Datei; seine Closure legt `evidence/slice-m2-referenzdienst.md` an. Behoben durch `slice-text-antworten` (Wortlaut in `SPEC-118` Nr. 11, Entscheidung des Owners).
