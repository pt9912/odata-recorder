# Worktrees im Arbeitsverzeichnis des Hauptbaums werden von dessen Gates mitgelesen

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Worktrees der Linien liegen unter `.wt/` im Arbeitsverzeichnis des Hauptbaums, jeder ist eine volle Kopie des Repos. Ein Gate, das `.` scannt, liest sie als Teil des Hauptbaums: `make a-check` meldete 1414 Befunde `wrong-direction`, die kein Stand des Hauptbaums trägt. Ob `make docs-check` (`scan.roots` ist `.`, `scan.ignore` nennt `.wt/` nicht), `make lint` und `make zeichen-check` dasselbe tun, ist ungeprüft.
**Stand:** offen
