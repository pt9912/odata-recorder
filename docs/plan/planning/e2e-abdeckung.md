# E2E-Abdeckung der Stufen des Voll-E2E

Erzeugt von `make e2e-abdeckung` aus den Deklarationen, die jede Stufe von
`tools/harness/test-integration.sh` an sich selbst trägt. Diese Datei ist eine **stabile
Abdeckungs-Deklaration, kein Lauf-Beleg**: der Erzeuger liest den Quelltext der Stufen,
er führt den E2E nicht aus, und er schreibt sie nur bei inhaltlicher Abweichung. Sie
ändert sich mit den Deklarationen und mit dem Ort ihrer Quellen — eine Einfügung
oberhalb einer Stufe verschiebt deren Zeile in der Spalte `Ort`.

Die Spalte `Spec-Kennung` nennt die Anforderung, die diese Stufe trägt, **als
Code-Span statt als Verweis**: ein Verweis brauchte einen Anker, und den müsste dieser
Erzeuger aus der Rohzeile einer Überschrift nachbilden — ein Link, der aussieht wie
einer und im Doku-Gate rot wird, wäre teurer als gar keiner. Gehalten wird die Sicht
deshalb gegen `spec/lastenheft.md`, und zwar vom Leser. Die Spalte
`Kurzbeschreibung` trägt keine Kennung. Die Spalte `Stufe` zählt die
Stufen-Kopfzeilen in der Reihenfolge des Skripts — eine Stufe eröffnet mit ihrer
Ausgabe-Kopfzeile und reicht bis zur nächsten. Die Spalte `Ort` adressiert sie.
Ob eine hier fehlende Anforderung eine Lücke ist, urteilt ebenfalls der Leser; ein
Waisen-Urteil fällt nicht hier.

| Spec-Kennung | Kurzbeschreibung | Stufe | Ort |
| --- | --- | --- | --- |
| `LH-RB-03` | Kein Außennetz, ungültige Konfiguration, SIGTERM (Happy Path, Negative) | Stufe 1 | `tools/harness/test-integration.sh:103` |
| `LH-FA-01` | Antwortarten unverändert, leere Antwort, Upstream gestoppt (Happy Path, Boundary, Negative) | Stufe 2 | `tools/harness/test-integration.sh:107` |
| `LH-FA-02` | Rundlauf mit GET und zwei POST, leere Aufnahme, Speicher voll (Happy Path, Boundary, Negative) | Stufe 3 | `tools/harness/test-integration.sh:111` |
| `LH-FA-03` | Neustart auf demselben DATA_DIR, Session ohne Request, beschädigte Datei (Happy Path, Boundary, Negative) | Stufe 4 | `tools/harness/test-integration.sh:115` |
| `LH-FA-04` | Replay gleich Aufnahme, Folge 7, 11, 13, Fehlerfälle und Übergang (Happy Path, Boundary, Negative) | Stufe 5 | `tools/harness/test-integration.sh:119` |
| `LH-FA-05` | Reihenfolge und Hexziffern, Filterwert, ungültige Kodierung (Happy Path, Boundary, Negative) | Stufe 6 | `tools/harness/test-integration.sh:123` |
| `LH-FA-14` | Geheimnisse, Nutzlast ohne sensible Felder, nicht redigierbarer Body, Speicher voll (Happy Path, Boundary, Negative) | Stufe 7 | `tools/harness/test-integration.sh:127` |
