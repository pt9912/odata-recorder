# Ein Stempel gehört zu keinem committeten Stand

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Stempel (`harness/gegenproben.stempel`) hält die Prüfsumme der Eingaben eines Laufs fest, damit der Lauf bei unveränderter Eingabe entfällt. Er wird an einem Arbeitsbaum erzeugt, der nicht der Stand ist, den der nächste Commit festhält (ein Commit unter den Eingabepfaden folgt, oder der Lauf sah ungetrackte Dateien); am HEAD stimmt die Prüfsumme nicht, der Lauf läuft erneut und ändert die Datei im Arbeitsbaum.
