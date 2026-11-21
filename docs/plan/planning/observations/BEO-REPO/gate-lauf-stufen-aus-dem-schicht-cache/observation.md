# Ein Lauf von `make gates` im Worktree kommt für Stufen aus dem Docker-Schichten-Cache

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Bei gleichem Inhalt von `src/` und gepinnten Eingaben liefert Docker die Stufen `build`, `test`, `race`, `lint`, `handbuch` und `katalog` aus dem Schichten-Cache (`CACHED` je `RUN pruefung …`); der Lauf dauert dann Sekunden statt Minuten und belegt dieselben Eingaben, aber keinen neuen Lauf der Tests. Ein Verifier, der „`make gates` grün“ als Beleg einträgt, belegt damit das Ergebnis eines früheren Laufs, solange er die Stufen nicht frisch fährt.
