# Verifier-Bericht slice-test-contract (Konformität gegen DoD, SPEC-042 und Entscheidungen)

Rolle: Verifier (einmal). Stand: Zweig `slice-test-contract`, HEAD `e2218b72`, 2026-11-08.
Bezug: `LH-QA-05`, `LH-RB-01`, [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md), [ADR-0012](../plan/adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md).
Vorgänger: Review `docs/reviews/2026-11-08-slice-test-contract-review.md`.

## Urteil

Konform, mit einer Auflage (Gegenproben-Stempel veraltet, Lauf wegen Host-Last rot) und zwei offenen Adressen, die der Plan schon trägt. Keine DoD-Verletzung, keine Abweichung von `SPEC-042`.

## Belege (selbst gefahren, nicht übernommen)

| Sensor | Ergebnis |
|---|---|
| `make gates` auf HEAD | Exit 0 (docs-check, schema, build, test, test-race, groessenmarke, lint, katalog, zeichen); `groessenmarke`: Messung 1311 von Marke 2000 |
| `make test-contract` | Exit 0, beide Binaries grün (`bootstrap.test`, `integration.test`) |
| Treiber `x` und `""` als Argument | `Aufruf:` auf stderr, Exit 2, ohne Container |
| Messung `git diff --numstat 54a1f1b4..HEAD` über die Pfade des Blocks | 7 + 11 + 679 + 488 + 126 = 1311 (gleich der Messung des Gates) |
| `git merge-tree --write-tree` HEAD gegen `main` und gegen `slice-bootstrap-m3-e2e-mock` | beide sauber (kein Konflikt, keine Überschneidung in `tools/` oder `harness/README.md`) |
| Fehlercodes | 40 verschiedene `ODR-E…` in `spec/spezifikation.md` (`katalog-check` grün) |
| `harness/README.md`, Zeile `make test-contract` | alle Zellen unter 200 Zeichen; `docs-check` grün |

## Plan-vs-Code-Diff

Geändert gegenüber `main`: `Dockerfile` (+7, nur die neue Stufe `contract`, vor der Stufe `handbuch` eingefügt, keine bestehende Stufe berührt), `harness/mk/test-contract.mk` (neu, kein `GATE_CHECKS`), `tools/harness/test-contract.sh` (neu), die zwei Testdateien, `SPEC-042`, `harness/README.md` (Zeile und Satz „Nicht behauptet“), Plan, Review-Report, `harness/gegenproben.stempel`. `Makefile` unverändert. Alles in §3 des Plans genannt; Stempel und Review sind erwartete Zugaben. Keine Datei außerhalb.

## DoD

- Liefer-Punkt 1 (Ziel, Treiber, Stufe, `SPEC-042`; 36 Fixtures über HTTP aus `abFixtures()`, keine Kopie): erfüllt.
- Liefer-Punkt 2 (vier Fehlerklassen, Klasse, Status, Kontext, keine Geheimnisse): erfüllt.
- Liefer-Punkt 3 (Mutation je Zusage, vom Implementer rot gesehen): erfüllt; Tabelle samt Fix-Runde in Plan §3, von mir stichprobenhaft und gegen die Klasse nachgefahren (unten).
- `make gates` grün: erfüllt (mit Auflage zum Stempel, kein Teil von `GATE_CHECKS`).
- Review liegt vor. Closure-Notiz, Register, Paarungen und §6-Ausgänge sind Sache des Planners: §6 trägt noch zwei Zeilen *offen* mit Adresse (Aufnahme in `GATE_CHECKS`; Spalten Soll und Nicht in MVP) und die Zeile zu Stufe und Treiber „offen bis zur Closure“ (Ausgang jetzt: entfallen/entschieden, der Handlauf ist gefahren und `make test-contract` grün).

## Tabelle Zusage · Fänger T01 bis T17 gegen `SPEC-042`

Jede der neun Nummern des Vertrags hat eine Zeile: Nr. 1 und 2 (Aufbau, Treiber) T16 als benanntes Negativ, Treiber-Argumente von mir gefahren; Nr. 3 a bis e T01, T02, T03, T04, T05, T06, T07; Nr. 4 T08 (9 Codes mal 20 Eigenschaften, in beiden Paketen, nachgezählt: 20 Einträge); Nr. 5 T09 bis T14; Nr. 6 T15 (Empfangs- und Sendeseite); Nr. 7 Entscheidung (Doppelung mit `make test`, wie gebaut); Nr. 8 a bis f im Code (`Fatal` bei Aufbau, `Error` bei Abweichung, Herunterfahren rot in beiden Paketen, `-test.count=1`, `t.TempDir`); Nr. 9 a bis j T16, T17. Keine Zusage ohne Fänger gefunden. Die Zählprüfungen (92 Schritte, 36 IDs, 4 Klassen) sind harte Schleifen.

Anmerkung (kein Befund): Nr. 2 nennt die Namen der sechs Tests im Treiber; sie stehen dort und werden gegen `-test.list` geprüft.

## Fix-Runde gegen die Klasse aufgelöst (`AGENTS.md` §3.17)

Mengen aus dem Wortlaut von Nr. 6 und Nr. 8 d, Mutationen in einer Kopie der Quellen im Container (Image-Stufe `source` unter eigenem Tag, ohne Netz, `--memory 6g`, `-count=1`), Last des Hosts zwischen 25 und 149, Messender: Verifier, je ein Lauf:

| Nr. | Menge | Mutation an einem vom Befund nicht genannten Element | Ergebnis |
|---|---|---|---|
| M1 Matrix-Sendeseite (Header × Methode, Anfrageart) | `Authorization`/`Cookie` je Anfrage der 92 | `Cookie` fehlt nur bei `PATCH` (B1) · `Authorization` fehlt bei Anfragen mit Query (B2) · `Cookie` doppelt (B3) · `Authorization` fehlt bei Anfragen mit Body (B4) | je rot in `TestVertragMatrix` (Server-Spiegel) |
| M1 Klassen-Sendeseite (Marker × Zeile × Pfad) | 3 Marker, 6 Anfragen in 4 Zeilen | `GEHEIM-Q` fehlt allein in `$search` (K1) · allein im Pfad des Replay-Treffers (K2) · `GEHEIM-Q` unerwartet in der Zeile *ungültige Metadaten* (K3) · `Authorization` anders (K4) · `Cookie` fehlt (K5) | je rot, mit Nennung von Zeile und Anfrage |
| M1 Empfangsseite (Marker × Statusklasse) | Antworten mit Echo | `X-Echo: Cookie` nur bei 501 (P1) · `Authorization` nur bei 404 (P2) · `Cookie` nur bei 422 (P3) im Adapter | je rot in beiden Tests |
| M2 Herunterfahren | jeder Server und jedes Handle | `httpodata.Server.Schliessen` meldet `ErrFristAbgelaufen` (H1) | rot in `TestVertragMatrix` (Cleanup des Servers) und in `TestVertragFehlerklassen` (`vertragHoch`; dort auch über den Aufbau der Aufnahme) |

Basislauf der Kopie: beide Pakete grün. Die Klasse ist gedeckt; kein Element der Mengen blieb grün.

Akzeptierte Negative (Plan trägt Grund und Folge): ein Header, den der Client sendet und den der Server nicht erreicht, bleibt in den Klassen ungesehen; 7 der 10 nicht unterstützten Optionen führt kein Fall; `$metadata` mit ungültigem Dokument (Nr. 9 i); zwei Fassungen von `vertragFehlerobjekt` ohne Gleichheitsfänger (Nr. 9 j, vom Architect entschieden).

## Auflage an den Integrations-Planner: Stempel der Gegenproben

`harness/gegenproben.stempel` ist für HEAD veraltet. Stand `58a03b48…`, Eingabe-Hash auf HEAD `8fa22172…` (die Fix-Runde `69eac720` änderte `src/bootstrap/vertrag_http_test.go` nach dem Stempel `aaeb8f71`). `make gegenproben` fuhr ich zweimal unter `flock`, beide Male 204 Fälle, rot allein die 8 bzw. 9 Fälle `lint/*` mit „Zeit überschritten“; alle übrigen Fälle (a-check, test, race, groessenmarke, zeichen, katalog, docs-check, migration) grün. Der Host hatte bei Beginn Last 24 bis 90 und in beiden Läufen mitten im Bau 118 bis 149 (andere Rollen), derselbe Befund wie bei den drei Läufen des Implementers. Ich habe deshalb keinen Stempel committet. Folge: Der Stempel muss nach der Integration auf `main` bei niedriger Last neu erzeugt werden (`make gegenproben`, bei grünem Lauf den Stempel committen). Keine Fehlermeldung außer der Zeitgrenze der `lint/*`-Fälle; ob die Fälle bei Last unter etwa 20 grün werden, habe ich nicht belegt.

## Offene Adressen (unverändert, nicht neu)

1. Aufnahme von `make test-contract` in `GATE_CHECKS`: Entscheidung des Users, Adresse Planner (Plan §6).
2. Spalten *Soll* und *Nicht in MVP* von `SPEC-049` ohne Ziel: bei Start der jeweiligen Welle.

## Befunde

Keine HIGH, MEDIUM oder LOW zum Code. Dokumentarisch: die §6-Zeile zu Stufe und Treiber trägt noch „offen bis zur Closure“; der Planner setzt den Ausgang bei der Closure.
