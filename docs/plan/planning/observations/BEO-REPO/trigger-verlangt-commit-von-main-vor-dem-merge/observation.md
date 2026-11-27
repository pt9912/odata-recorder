# Ein Auflösungs-Trigger verlangt einen Commit von `main`, gemessen wird aber vor dem Merge

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Carveout nennt als Trigger eine Messung „auf einem Commit von `main`“. Im Parallelbetrieb liefert ein Slice-Zweig die Behebung, die DoD verlangt die Messung auf einem Commit, der sie trägt, und die Closure im Zweig soll den Carveout schließen; der Merge folgt erst danach. Zum Zeitpunkt der Messung ist der Trigger nicht erfüllbar, und die Auflösung hängt an einer Zusage über den späteren Merge (gleicher Stand der gemessenen Pfade), die kein Gate prüft.
