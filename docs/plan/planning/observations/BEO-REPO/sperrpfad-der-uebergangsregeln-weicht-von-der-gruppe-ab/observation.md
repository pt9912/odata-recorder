# Die Übergangsregeln nennen einen Sperrpfad, den die Maschinen der Gruppe nicht nehmen

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Nr. 1 der Übergangsregeln nennt `/tmp/odata-recorder-heavy.lock` für schwere Läufe; die Gruppe der Maschinen sperrt unter `/home/builder/work/odata-recorder/.wt/heavy.lock`. Wer der Regel folgt, sperrt gegen niemanden, der die andere Datei nimmt, und zwei schwere Läufe überschreiben die festen Image-Tags des Hosts; die Rollen weichen im Plan auf den zweiten Pfad aus, die Regel bleibt, wie sie steht (nur der Integrationsschritt schreibt sie, auf Anweisung des Users).
