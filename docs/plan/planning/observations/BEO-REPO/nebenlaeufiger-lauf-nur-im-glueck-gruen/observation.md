# Ein nebenläufiger Lauf war im Review nur im Glücksfall grün

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Werkzeug fährt Fälle nebenläufig und teilt dabei eine Ressource, die Parallelität nicht verträgt (Sperre von golangci-lint in `TMPDIR`); Implementer und Review sahen nur grüne Läufe, erst die Verifikation mit erzwungenen Wiederholungen fand den Abbruch.
