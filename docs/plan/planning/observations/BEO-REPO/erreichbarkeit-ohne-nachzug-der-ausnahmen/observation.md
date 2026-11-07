# Ein Slice macht einen Fehlercode, Schlüssel oder ein Kommando erreichbar und lässt Ausnahmen und Handbuch stehen

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Abgleich der Kataloge (`SPEC-201`) führt eine Ausnahme in `harness/katalog-ausnahmen.txt` für jeden Code, den das Binary dem Anwender nicht zeigt; der Standard ist die Pflicht im Handbuch. Macht ein späterer Slice einen solchen Code erreichbar und lässt die Zeile stehen, bleibt das Gate grün und das Handbuch schweigt (akzeptiertes Negativ `SPEC-201` Nr. 9 c). Die Pflicht „wer einen Code, Schlüssel oder ein Kommando erreichbar macht, ändert Ausnahmen und Handbuch im selben Slice“ (`SPEC-201` Nr. 11) steht nur im Vertrag und in den Plänen, in keiner Anweisung der Rollen.
