# Ein Datenrennen in einem Kindprozess des Tests macht `make test-race` nicht rot

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Test startet das eigene Test-Binary als Kindprozess. Der Detektor meldet dort ein Rennen allein in dessen Ausgabe und über dessen Exit-Code; wertet der Elternprozess die Ausgabe nur bei Fehlern aus oder erwartet er ohnehin ein anderes Ende (ein Signal, das dem Exit-Code zuvorkommt), bleibt der Lauf unter dem Detektor grün. Die Zusage von `make test-race` reicht dann nicht in den Kindprozess, ohne dass ein Gate es zeigt. Nahe an `zweig-ohne-lauf-unter-dem-detektor` (dort läuft der Zweig nicht, hier läuft er und die Meldung geht verloren).
