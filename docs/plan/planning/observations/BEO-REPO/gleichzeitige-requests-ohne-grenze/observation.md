# Das Binary begrenzt die Zahl gleichzeitiger Requests nicht, und unter Stau antwortet es mit 500

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Server nimmt jede Zahl gleichzeitiger Requests an. Speicher wächst je offener Anfrage, und die Schreiber einer SQLite-Ablage warten nacheinander; staut sich die Last, enden einzelne Anfragen nach `SQLITE_BUSY_TIMEOUT` mit `ODR-E9001`. Die Spezifikation beschreibt das Warten (`SPEC-135` *Ungeprüft*) und sagt keine Zahl zu. Ob das Binary begrenzen oder abweisen soll, ist ein sichtbares Verhalten und nicht entschieden.
