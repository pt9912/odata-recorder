# Ein Stand aus mehreren Sammlungen unter einer Sperre hat ein Fenster zwischen den Sammlungen, das weder Detektor noch Suite sehen

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Eine Zusage der Form „ein Stand zu einem Zeitpunkt“ gilt für mehrere Sammlungen unter einer Sperre. Gibt der Code die Sperre zwischen zwei Sammlungen frei, bleibt jeder Zugriff geordnet, der Detektor von `make test-race` meldet nichts, und ein Test, der nur einen Teil der Zähler ändert, sieht das Fenster nicht. Nahe an `nebenlaeufiger-lauf-nur-im-glueck-gruen` und `zweig-ohne-lauf-unter-dem-detektor`.
