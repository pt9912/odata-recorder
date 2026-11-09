# Ein bedingter Pfad im Block der Größenmarke zählt die Zeilen, die ein anderer Slice über main bringt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Block eines Plans nennt auch Pfade, die erst nach einem Merge von `main` existieren (Zählstelle in einem Paket, das ein paralleler Slice liefert, Plan §3 *bedingt*). Die Messung zählt jede Zeile an den Pfaden zwischen `basis` und `HEAD` (`SPEC-128` Nr. 10 g): mit dem Merge zählen die Dateien des anderen Slice ganz, nicht nur die Nachzüge. Die Auslastung am Ende liest sich dann um den fremden Anteil zu hoch, und der Plan, der „der Merge ist an diesen Pfaden leer“ sagte, stimmt nicht mehr.
