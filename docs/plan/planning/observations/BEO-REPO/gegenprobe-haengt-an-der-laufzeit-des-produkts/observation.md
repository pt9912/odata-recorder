# Ein Fall der Gegenproben fährt das Produkt mit, und seine Laufzeit wächst mit dem Produkt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Gegenproben prüfen ein Werkzeug (`pruefung`), nicht das Produkt. Ein Fall, dessen Eingabe die Pakete des Produkts enthält, fährt dessen Suite mit; die Zeit des Falls wächst mit jedem Slice, und die Frist je Fall reißt, ohne dass sich am Werkzeug etwas geändert hat. Die naheliegende Erklärung ist Last von außen; die Ursache ist der Zuschnitt des Falls.
