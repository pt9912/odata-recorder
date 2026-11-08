# Ein Verweis mit Pfad auf einen Review-Report sperrt die Archivierung der Welle

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

`make archive-welle` löscht die Review-Reports der Slices einer Welle und sperrt mit `[haenger]`, solange eine Datei den Pfad eines Reports trägt. Evidence-Dateien des Beobachtungs-Registers, Pläne in `done/` und die Roadmap nennen Reports mit Pfad, weil die Rollen sie so schreiben; die Sperre entsteht neu, sobald nach der Auflösung eine weitere Datei einen Report mit Pfad nennt. Kein Gate und keine Anweisung der Rollen schreibt die pfadlose Kennung vor.
