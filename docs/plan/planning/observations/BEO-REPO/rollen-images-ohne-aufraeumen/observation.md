# Rollen bauen Images und Volumes unter Einmal-Tags und räumen sie nicht auf

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Rollen der Slices (Implementer, Reviewer, Verifier, Mutationsläufe) bauen und starten Docker-Images und -Volumes unter Einmal-Tags und löschen sie am Ende nicht. Kein Gate und keine Anweisung einer Rolle verlangt das Aufräumen; die Ablage des Hosts wächst mit jedem Slice, bis ein Lauf am Speicher scheitert oder der Auftraggeber von Hand aufräumt.
