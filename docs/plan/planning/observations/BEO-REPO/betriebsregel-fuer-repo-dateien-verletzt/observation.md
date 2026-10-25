# Eine Rolle ändert eine Repo-Datei mit `sed -i` oder Heredoc, obwohl die Betriebsregel des Plans es verbietet

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Plan eines Slice schreibt allen Rollen vor, Repo-Dateien nur mit `Edit` und `Write` zu ändern (kein `sed -i`, kein `awk`, kein Python, kein Heredoc in die Datei). Ein Lauf hält die Regel nicht und meldet es selbst oder die nächste Prüfrunde findet es; kein Gate und kein Wächter erzwingt sie (der Command-Guard des Repos sperrt `python`, nicht `sed`), und die Folge ist erst am Diff zu prüfen.
