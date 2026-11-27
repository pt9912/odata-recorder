# Der Abfrageplan der Datenbank nennt für eine Suche und für einen Durchlauf des Bereichs dieselbe Zeile

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Fänger zum Aufwand einer Abfrage vergleicht den Text von `EXPLAIN QUERY PLAN` mit einem Literal. SQLite nennt für `max(spalte)`, `max(spalte + 0)` und `count(*)` unter derselben Bedingung dieselbe Zeile; eine Abfrage, die den ganzen Bereich durchläuft, bleibt grün. Den Unterschied zeigt erst ein Fänger, der die Wirkung misst (hier: eine Datei, in der die ersten Blätter der B-Bäume genullt sind, so dass ein Durchlauf scheitert).
