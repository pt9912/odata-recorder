# Eine Antwort mit unbekanntem Typ ist nie abspielbar, auch nicht die echte Antwort auf `/$count` und `$value`

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Spezifikation macht jeden unbekannten Inhaltstyp, `text/*` eingeschlossen, zu `unsafe` (`SPEC-118` Nr. 3), und `SPEC-045` verbietet den Replay einer solchen Antwort (`ODR-E5008`). Die echte OData-Antwort auf `/$count` und auf `$value` ist `text/plain`; sie wäre damit nie abspielbar. Ob das gewollt ist, berührt `LH-FA-04` und entscheidet der User.
