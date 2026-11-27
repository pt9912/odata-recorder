# Ein asynchroner Abbruch liefert selten den Fehler des Datenbank-Treibers statt des Kontextfehlers

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Vertrag der Ablage sagt bei abgebrochenem Kontext den Fehler des Kontexts zu (`SPEC-105` Nr. 13). Bricht der Kontext ab, während der Treiber eine Anweisung ausführt, endet die Operation selten mit `interrupted (9)` des Treibers; ein Dienst darüber meldet einen Fehler mit Code statt des Fehlers ohne Code. Ein Haltepunkt am Kontext ohne eigene Goroutine erreicht den Weg nicht, die Wirkung tritt nur mit Wahrscheinlichkeit auf.
