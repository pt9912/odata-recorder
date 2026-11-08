# Die Sendeseite eines Markers hat keinen Fänger, nur die Empfangsseite

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Test sucht einen Marker (Header, Query-Wert, Geheimnis) in der Antwort und sagt damit zu, dass kein Geheimnis durchsickert. Sendet der Test den Marker gar nicht, findet die Suche nichts und bleibt grün: die Zusage hat zwei Seiten (der Marker steht in der Anfrage; er steht nicht in der Antwort), und ein Fänger, der nur die Empfangsseite prüft, wird stumm, sobald die Sendeseite entfällt. Nahe an `vertrag-sagt-mehr-zu-als-ein-gate-prueft`.
