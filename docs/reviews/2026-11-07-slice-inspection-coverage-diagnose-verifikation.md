# Verifikation: slice-inspection-coverage-diagnose — 2026-11-07

**Art:** Verifikation (einmal), Code und Plan gegen die DoD des Slice, `SPEC-236` Nr. 1, 2, 6, 7, 8, 10, `SPEC-109` Nr. 5, `SPEC-110` Nr. 2, `SPEC-230` Nr. 9, `LH-FA-13`, `LH-FA-11`, `LH-QA-04`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0007](../plan/adr/0007-race-detektor-im-gate.md), [ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md) und die Hard Rules.

**Gegenstand:** Zweig `slice-inspection-coverage-diagnose`, HEAD `0ae9fd8` (Implementer `331d3d8`, Review-Report), Basis `9a9180d`. Plan: `docs/plan/planning/done/slice-inspection-coverage-diagnose.md`. Review: `docs/reviews/2026-11-07-slice-inspection-coverage-diagnose-review.md`.

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden, mit Auflagen an den Planner und den Architect; keine Fix-Runde des Implementers nötig.** Kein Befund ist MEDIUM oder höher. Die DoD-Punkte, die vor der Closure erfüllbar sind, sind erfüllt (Coverage, Diagnose, `make gates` grün, Review liegt vor); die übrigen sind Aufgabe der Closure des Planners (Closure-Notiz, Register, Ausgänge in §6, Doku-Update, Merge). Die vier LOW und die INFO des Reviews sind bestätigt; die Verifikation fand zwei weitere LOW in denselben Klassen (Wertformen der Einstellungen, Formen der Baseline-Kennung) und einen veralteten Satz in §6. Die vier Abweichungen des Implementers sind nachgerechnet und sachlich richtig.

## 1. Belege der Sensoren (selbst gefahren)

| Sensor | Beleg |
|---|---|
| `make gates` | einmal selbst, unter `flock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/verif/gates.log`, Ende `EXIT=0`; darin `a-check` (grün, [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md)), `docs-check`, `lint` ([ADR-0003](../plan/adr/0003-lint-profil-golangci.md)), `test`, `race` ([ADR-0007](../plan/adr/0007-race-detektor-im-gate.md)), Stufe `groessenmarke`: `groessenmarke slice-inspection-coverage-diagnose.md: ok, Messung 1725 von Marke 3000`, `zeichen: ok`. |
| Größenmarke | `git diff --numstat 9a9180d..HEAD` über `src/hexagon/ports/inbound/inspection.go` und `src/hexagon/services/inspection/`: 1678 + 47 = 1725, gleich Block und Gate ([ADR-0009](../plan/adr/0009-groessenmarke-als-gate-mit-messung-durch-den-host.md)); Auslastung 0,58. |
| Diff-Umfang | `git diff --name-only 9a9180d..HEAD`: Plan, Review-Report, `ports/inbound/inspection.go` und neun Dateien unter `services/inspection/`. `ports/outbound/`, SQLite-Adapter, Telemetry-Adapter, `spec/`, `harness/`, `Makefile`, `.a-check.yml`, `.golangci.yml`, `model/` sind nicht berührt (Plan-vs-Code, siehe 2). |
| `//nolint` | keine im Paket und im Port (`grep`). |
| Rohbytes (`AGENTS.md` §3.18) | `LC_ALL=C grep -P` auf Backslash-`u`-Hex-Folge, BOM, NBSP, Zeilentrenner über `services/inspection/*.go`, `inbound/inspection.go`, Plan und Review-Report: kein Treffer. Dasselbe wurde an diesem Report nach dem Schreiben wiederholt (siehe 8). |
| Abdeckung | `-race -covermode=atomic -coverpkg` im Container `odata-recorder:race` aus Baumkopie (`--memory=4g`, `flock`): `TestGleichzeitig` allein 94,1 % (`New` 60,0 %, alle übrigen Funktionen 100 %, also `Coverage`, `Diagnose`, `Erlaubt` ohne Block ohne Treffer), alle Tests 100,0 %. Das stimmt mit dem Plan überein; der einzige Rest sind die vier Verwurf-Blöcke von `New`, im Plan als Negativ mit Grund und Folge benannt. Register `zweig-ohne-lauf-unter-dem-detektor`: `TestGleichzeitig*` zehnmal als eigenes Binary mit Detektor ohne Mutation: 0 von 10 rot (grün, keine Störung durch Last). |
| Zählprüfung M-Codes | Baumkopie mit `model/fehler.go` und `ressource_mengen.go` aus dem Zweig `slice-mock-lesen` (38. Code `ODR-E6008`): `go test -race` über `services/inspection` und `ports/...` grün. Im Test steht keine feste Zahl (`grep` auf 27, 28, 37, 38, 185: kein Treffer); die Mengen laufen über `model.AlleCodes()` und `outbound.EintraegeAbbildung()`. |
| K5 Bindung | Baumkopie, Container `odata-recorder:test`, `go vet` im Paket: Zusatzparameter bei `Coverage` meldet `./port_test.go:19:35 … does not implement inbound.InspectionUseCase (wrong type for method Coverage)`, `Diagnose` umbenannt `… (missing method Diagnose)`, Nullprobe ohne Meldung. Die Zeile von `port_test.go` und der Wortlaut stimmen mit dem Plan überein; es ist der Nachweis der Übersetzungsmeldung, kein Rot im Mutationslauf (so im Plan benannt). |

## 2. Plan gegen Code

| Plan-Aussage | Befund |
|---|---|
| Port: `CoverageEintrag{EntitySet, Status, Evidenz}`, `FehlerZaehler{Code, Anzahl uint64}`, `Diagnose{Modus, Einstellungen, Fehler}`, Methoden `Coverage(ctx, baselineID)`, `Diagnose(ctx)` | stimmt mit dem Block in `SPEC-236` wörtlich überein. |
| `Konfig` um `Coverage`, `Zaehler`, `Modus`, `Einstellungen`; `Exchanges`/`New`-Signatur bleiben | stimmt; `New` prüft nil in drei Feldern, Modus mit `switch` über die vier Literale, `maps.Clone` für die Einstellungen (Nr. 1). |
| `Erlaubt()` 18 Schlüssel in der Folge des Wortlauts | Liste im Code gleich Nr. 7 (Byte für Byte abgeglichen); alle 18 stehen in `SPEC-110` Nr. 2, keiner trägt Adresse, Pfad oder Geheimnis; die 9 Einzelnamen plus 18 sind die 27 Schlüssel (`UPSTREAM_CA_FILE` ist unter den 19 Schlüsseln ab `REPLAY_STRATEGY`, nicht erlaubt). `ADMIN_TOKEN_FILE` gehört zu `ADMIN_*`. |
| `Coverage`: `ctx` nil, leere Kennung, genau ein `List`, Abbildung mit Gegenstand `Snapshot`, Nullwert `nil` im Fehlerfall, leere Liste nie `nil`, Reihenfolge und Werte unverändert | stimmt (Code Zeilen 98 ff.); kein Kontext-, kein Statuszugriff, kein Cache. |
| `Diagnose`: `ctx` nil, ein `ctx.Err()`, ein `Snapshot`, Einstellungen aus der Allowlist, ein Eintrag je Code von `AlleCodes()`, Eintrag außerhalb verworfen, Ergebnis eine Kopie | stimmt; `Einstellungen` und `Fehler` sind im Erfolgsfall nie `nil`. |
| Unberührt: Port outbound, SQLite, Telemetry, `SPEC-109`, `SPEC-230`, Katalog, Handbuch | bestätigt durch `--name-only`; kein neuer Code, Schlüssel, kein Kommando, das Binary bleibt (§7 der Closure muss es sagen). |
| Sendung `SPEC-109` Nr. 5, Absatz *Anwendung*, Satz zu `Coverage`/`Snapshot` im Merge-Commit | **Auflage 6**: der Satz steht schon auf `main` (und im Zweig, `spec/spezifikation.md` Zeile 856): „`InspectionUseCase` (`SPEC-236` Nr. 8) wendet die Tabelle an, mit dem Gegenstand `Session` bei `Exchanges` … und `Snapshot` bei `Coverage` …; `Diagnose` liefert keinen Fehler mit Code“. Er ist vollständig und stimmt mit dem Code überein; der Merge-Commit hat nichts einzutragen, und §6 (Zeile „`SPEC-109` Nr. 5, Absatz *Anwendung*, nennt `InspectionUseCase` nicht“) ist veraltet. |

## 3. DoD

| DoD-Punkt | Stand |
|---|---|
| Coverage je Baseline, unbekannt `ODR-E7002` | erfüllt (Code, C1 bis C6, Mengen; Mutationen unten). |
| Diagnose: Allowlist, Modus, Zähler je Code über `TelemetryReader`, auch mit 0 | erfüllt (D1 bis D8; Zählprüfung gegen `AlleCodes()`, auch mit 38 Codes). |
| `make gates` grün | erfüllt, eigener Lauf (1). |
| Review mit Report | erfüllt (`…-review.md`). |
| Doku-Update `SPEC-109` Nr. 5 | schon auf `main` erfüllt (Auflage 6). |
| Katalog/Handbuch | nicht berührt, das Gate ist grün; §7 muss „Binary unverändert“ sagen. |
| Closure-Notiz, Register, Ausgänge, Paarungen | offen, gehören zur Closure des Planners (Auflage 5). |

## 4. Tabelle Zusage · Fänger gegen den Vertrag

Alle Sätze von `SPEC-236` Nr. 1, 2, 6, 7, 8, 10 und des Absatzes vor Nr. 1 haben nach der Tabelle *Sätze der Nummern* eine Zeile oder ein Negativ; ich habe die Tabelle gegen den Wortlaut gelesen und keinen Satz ohne Zeile gefunden. Zeilen K2 bis K5, A2, C1 bis C6, D1 bis D8, W3, W4 haben Tests mit Schleife über die Menge der Quelle (keine feste Zahl). Der Orakel-Status `Snapshot` (`tabelleSnapshot`, Abweichung 4) stimmt mit `SPEC-109` Nr. 5 überein (`ErrNichtGefunden` gleich `ODR-E7002`, `ErrBereitsVorhanden` und `ErrUngueltig` nennen keine Zeile für `Snapshot` und fallen auf `ODR-E9001`). Zwei Lücken der Klasse „Zusage mit unbeschränkter Werte-Menge, ein Fänger mit endlicher Auswahl“ stehen unter Findings 5 und 6; die Aussagen des Reviews stehen unter 6.

## 5. Stichproben-Mutationen gegen die Klasse (`AGENTS.md` §3.17)

`tools/harness/mutationslauf.sh`, `flock /tmp/odata-recorder-heavy.lock`, `MUTATIONSLAUF_SPEICHER=4g`, Image `odata-recorder:race`, Messender: der Verifier, Last des Hosts 20 bis 27 bei 4 CPUs des Containers. Listen: `…/scratchpad/verif/liste.json` (28 Mutationen) und `…/scratchpad/verif/liste2.json` (7). Alle an Elementen, die weder Implementer noch Review nannten. Ergebnis Liste 1: `rot=22 rot-paket=1 gruen=4 uebersetzungsfehler=1` (die Mutation mit dem Übersetzungsfehler war ein Fehler meiner Liste und ist in Liste 2 als `w-list-doppelt-bei-nichtgefunden` wiederholt). Liste 2: `rot-paket=2 gruen=5`.

| Klasse (Wortlaut der Zusage) | Menge und Elemente | Mutationen (nicht genannt) | Ergebnis |
|---|---|---|---|
| Coverage-Eintrag byte-gleich, nichts abgeleitet | Status {4} × Evidenz {3} × m {5}: 60 Fälle | fremder Status zu `unknown`; fremder Status verworfen; leere Evidenz verworfen; leere Evidenz zu `{}`; letzte zwei Einträge getauscht (m über 2); erster/letzter getauscht ab 100 | alle `rot` (`TestCoverageEintraege`) |
| keine Seitenbildung, kein Kürzen | m {0, 1, 2, 101, 1001} | Kürzen ab 1500 | `gruen` — Prüfgrenze, im Plan bei *Pflicht-Mengen* (Spalte Prüfgrenze von M-Coverage) mit Grund und Folge benannt: INFO 10 des Reviews ist damit erledigt, N2 braucht nur den Zeiger |
| Evidenz unverändert | drei Formen | Kürzen über 4096 Bytes | `gruen` — im Plan als unbeschränkte Evidenz-Form (drei geprüft) mit Folge benannt |
| genau ein `List` | Erfolg leer, Erfolg mit Daten, jeder Fehler | zweiter `List` bei leerer Liste; zweiter `List` bei `ErrNichtGefunden` | `rot-paket` (der benannte Test `TestCoverageSchreibtNichts` ist grün, `TestCoverageEintraege` bzw. die Fehlertests zählen `List` und sind rot) |
| Fehler des Ports mit Code | alle Einträge von `EintraegeAbbildung()` + anderer | `ErrUngueltig` des Ports ohne Code durchgereicht | `rot` |
| Kontextfehler in `Diagnose` | `Canceled`, `DeadlineExceeded` | nur `Canceled` abgewiesen | `rot` |
| `Modus` unverändert | 4 Modi | `hybrid` zu `mock` | `rot` |
| `Einstellungen` nur Allowlist, Wert unverändert | 18 Schlüssel × 4 Modi × 6 Abwandlungen; Werte | `LINK_REWRITE` nur im Modus `mock` verworfen; Wert ohne führenden Leerraum; Wert ohne nachfolgenden Leerraum; Komma-Leerzeichen am Ende gekürzt; Wert mit Zeichen jenseits von ASCII zu `?` | die Schlüssel-Mutation `rot`; die drei Wert-Mutationen mit Leerraum `gruen`; Nicht-ASCII `rot-paket` (Finding 5) |
| Zähler je Code | 5 Stände × `AlleCodes()` | Kappe bei 2 hoch 63; Stand 1 wird 0 | `rot`, `rot` |
| `Erlaubt()` genau die 18 | Liste | 7. und 13. Schlüssel entfernt; `UPSTREAM_CA_FILE` ergänzt; Dublette | alle `rot` |
| `Einstellungen` und `Fehler` nie `nil` | nil-Konfiguration, nil `Errors` | Einstellungen `nil` bei leerer Konfiguration; `Fehler` `nil` bei leeren `Errors` | `rot`, `rot` |
| baselineID unverändert, nur leer ungültig | 9 Formen × 2 Zustände | abgewiesen ab 512 Zeichen; abgewiesen mit NUL am Anfang; abgewiesen `\t`/`\n` allein | alle `gruen` (Finding 6, zusammen mit LOW 2 des Reviews) |
| kein Zustand, kein Zwischenspeicher (nebenläufig) | Zweige von `TestGleichzeitig` | Schreibzugriff auf ein Feld des Dienstes im Fehlerzweig von `Coverage`; im Zweig „erlaubter Schlüssel“ von `Diagnose`; bei leerer Liste; in der Codeschleife nur für den letzten Code | je `rot`, mit Detektor 10/10 Prozesse, ohne Detektor Arm entfällt (Größenordnung unter Last 20 bis 27; die Rate ohne Detektor steht im Plan mit 0/20) |

Zusammenfassung: 35 Mutationen, 22 + 2 `rot`/`rot-paket` an der Klasse wie zugesagt; die Lücken konzentrieren sich auf zwei unbeschränkte Wertemengen (Wert der Einstellung, Form der Kennung), nicht auf einen Zweig oder eine Zusage des Vertrags.

## 6. Findings

| # | Kategorie | Quelle | Pfad · Anker | Befund | Entscheidung |
|---|---|---|---|---|---|
| 1 | LOW | `AGENTS.md` §3.11 | `inspektion.go` · Kommentar von `Erlaubt` „in der Folge von“ | Bestätigt (Review LOW 1): `SPEC-236` Nr. 7 sagt eine Folge der Schlüssel nicht zu (die Aufzählung ist der Wortlaut, `Einstellungen` ist eine Abbildung), `TestErlaubtGleichWortlaut` vergleicht als Menge. Die Folge ist von außen nicht beobachtbar außer über `Erlaubt()` selbst. | Auflage 1: den Kommentar auf „die Schlüssel von `SPEC-236` Nr. 7“ kürzen (eine Zeile, ohne neuen Lauf, im Integrationsschritt oder als Ein-Zeilen-Commit) **oder** das Negativ „Folge von `Erlaubt()` ungeprüft“ in N2 mit Grund (nur Menge zugesagt, Ergebnis ist Abbildung) und Folge (keine im Betrieb) eintragen. Die erste Form ist die kleinere. |
| 2 | LOW | `AGENTS.md` §3.15 | Plan N2 · Zeile „Kennung“ | Review LOW 2 gilt für die ganze Klasse: Mutationen „über 512 Zeichen“, „NUL am Anfang“, „nur Tab oder Zeilenwechsel“ bleiben grün (Finding 6). | akzeptiertes Negativ, siehe Auflage 3 |
| 3 | LOW | `AGENTS.md` §3.15 | Plan N2 | Bestätigt (Review LOW 3): „kein Zähler“ im Dienst ohne Fänger und ohne N2-Zeile. | Auflage 2: N2-Zeile: Grund — ein Zähler im Dienst (Kanal, Paketvariable) ist von außen nicht zu sehen und braucht `Telemetry` nicht, das `W3` nur über Importe und Felder von `Konfig` fängt; Folge — keine für den Betrieb, nur ein toter Zähler; die Zusage „kein Zähler“ meint `Telemetry`/`CountError`, und die fängt D2 (Aufzeichnung). |
| 4 | LOW | `AGENTS.md` §3.9, §3.13 | Plan §6 · Zeile „Die Tests des Vorgängers bauen `Konfig` mit `Exchanges` allein“ | Bestätigt (Review LOW 4): der Nachzug ist getan (`konfigMit`, `neuerDienst`, 17 Wege in 5 Arten; Gates grün), die Zeile steht weiter mit Ausgang *offen*; ebenso die Zeilen „Lesen der Zähler“-Folge „`SPEC-236` steht … liefert der Code nicht“ (Lücke geschlossen) und „Zählprüfungen … Bericht je Befund“. | Auflage 5: Ausgänge bei der Closure setzen (eingetreten/entfallen); der Nachzug-Punkt ist *entfallen* (erledigt im Code-Commit `331d3d8`). |
| 5 | LOW | `AGENTS.md` §3.15, §3.17 | `diagnose_test.go` · Werte `"wert-" + k` | Neu. „mit unverändertem Wert“ (Nr. 7) wird mit ASCII-Werten ohne Leerraum geprüft (`wert-<Schlüssel>`, leer, lang); die drei Mutationen „führenden Leerraum entfernt“, „nachfolgenden Leerraum entfernt“, „Komma-Leerzeichen am Ende gekürzt“ bleiben grün. Echte Werte (Listen wie `a, b`) tragen Leerraum und Komma. Der Dienst hat keinen Weg, der Werte formt; die Mutation setzt eine künftige Umformung voraus. | Auflage 4: akzeptiertes Negativ in N2 mit Grund (Werte sind unbeschränkt, der Dienst reicht sie durch und liest sie nie) und Folge (eine Umformung von Leerraum oder Nicht-ASCII-Zeichen im Wert fiele erst dem Fänger des Roots auf, `slice-bootstrap-m6`; im Betrieb zeigte die Diagnose einen abweichenden Wert, ohne Geheimnis). Alternativ ein Wert mit Leerraum, Komma und Zeichen jenseits von ASCII in der Fixture — das ist ein Implementer-Lauf und deshalb hier nicht verlangt. |
| 6 | LOW | `AGENTS.md` §3.15, §3.17 | `coverage_test.go` · Kennungen von `TestCoverageKennung` | Neu, erweitert LOW 2: die Kennung ist unbeschränkt („jede andere … unverändert“, Nr. 6), geprüft sind 9 Formen; Länge über 64, NUL und `\t`/`\n` sind ohne Fänger. Eine Prüfgrenze für die Achse Kennung nennt der Plan nicht (die Spalte Prüfgrenze von M-Coverage nennt nur m und Evidenz). Der Dienst hat genau eine Prüfung (`== ""`). | Auflage 3: akzeptiertes Negativ in N2 (oder in der Prüfgrenze von M-Coverage) mit Grund (die Menge der Kennungen ist unbeschränkt, der Dienst kennt eine Regel) und Folge (ein späteres Abweisen weiterer Formen verschöbe `ODR-E7002` nach `ErrUngueltig` ohne Test; der Port weist unbekannte Kennungen ohnehin als `ErrNichtGefunden` ab). |
| 7 | INFO | `AGENTS.md` §3.13 | Plan §6 · Abweichung 1 | Nachgerechnet: 5 Zweige ohne Portfehler (`ctx` nil, Kennung leer, `nil`, leere Liste, 101 Einträge) + 5 Kontextformen + (n + 1) Fehler mit Code = 11 + n. Die Zahl der Hand-Liste ist richtig; der Plan nennt sie schon berichtigt. | Auflage 7: Architect bestätigt die Zeile *Zweig-Menge* (formal, kein Wortlaut der Spezifikation). |
| 8 | INFO | `SPEC-236` Nr. 7 | Abweichung 2 | `[]` für `Fehler` ist unerreichbar, weil `AlleCodes()` nie leer ist; der Satz „sieht `{}` und `[]`“ ist als Aussage „nie `nil`“ wahr, im Beispiel ungenau. Der Code folgt der Zusage (nie `nil`). Rückgabe an den Architect (Wortlaut der Spezifikation). | Auflage 8: Architect kürzt den Nebensatz zu „`Einstellungen` und `Fehler` sind nie `nil` (`Einstellungen` im JSON `{}`, wenn leer)“; kein Code- und kein Testwechsel. |
| 9 | INFO | `AGENTS.md` §3.11 | Abweichung 3, `diagnose_test.go` Kommentar von `ausgeschlossenV` | Nachgerechnet: die Aufzählung der Zeile M-Ausgeschlossen ergibt 10 + 4 + 1 + 1 + 1 + 1 = 18, nicht 19; der Test führt 19 (zwei Vertreter von `ADMIN_*`), und die Zählprüfung hält 19 × 3. Der Kommentar „die zehn Namen … je ein Vertreter“ stimmt für die Aufzählung, nicht für die 19. Ohne Wirkung auf die Zusage (der zweite Vertreter ist ein weiteres Element der Menge). | Auflage 9: Planner (oder Architect) berichtigt die Zahl der Zeile (18 oder „19 mit zwei Vertretern von `ADMIN_*`“); der Kommentar im Test darf mitgezogen werden. |
| 10 | INFO | `AGENTS.md` §3.15 | Abweichung 4, `tabelleSnapshot` | Die Tabelle ist aus `tabelleSession` abgeleitet (Zeilen wie `SPEC-109` Nr. 5 für `Snapshot`; unabhängig vom Dienst). Sie hängt an den 24 unveränderten Zeilen von Teil 1; die Mutation „Gegenstand `Session`“ ist rot. | angenommen. |
| 11 | INFO | `AGENTS.md` §3.15 | Plan N2 | Weitere Grünfälle, die Auswahl der Klasse „Eintrag der Liste“ betreffen: zwei Einträge mit gleichem `EntitySet` zusammengefasst (der Port liefert je `entity_set` einen, `SPEC-107`); Kürzen ab 1500 (Prüfgrenze benannt); Evidenz über 4096 Bytes (benannt). | N2 ergänzt um den Zeiger auf *Pflicht-Mengen*, wenn Auflage 3 geschrieben wird; die Dublette ist kein Zustand des Ports. |
| 12 | INFO | `AGENTS.md` §3.9 | Plan §6 · Zeile *SPEC-109 Nr. 5, Absatz Anwendung* | Siehe Plan-vs-Code und Auflage 6: der Satz steht schon auf `main`. | Auflage 6 |

## 7. Auflagen

An den **Planner** (kein neuer Implementer-Lauf, außer er entscheidet anders):

1. Finding 1: Kommentar von `Erlaubt` auf die Menge kürzen (Ein-Zeilen-Edit im Integrationsschritt) oder als Negativ in N2 mit Grund und Folge eintragen.
2. Finding 3: N2 um „kein Zähler im Dienst selbst“ mit Grund und Folge ergänzen.
3. Finding 2, 6, 11: N2 (oder die Spalte Prüfgrenze von M-Coverage) um die Achse Kennung ergänzen — Formen über 64 Zeichen, NUL, Steuerzeichen, Tab und Zeilenwechsel — mit Grund (eine Regel im Dienst: leer) und Folge (siehe Finding 6); dazu ein Zeiger auf die Prüfgrenzen von m und Evidenz (Finding 11).
4. Finding 5: N2 um „Formen des Werts einer Einstellung (Leerraum, Komma, Zeichen jenseits von ASCII)“ mit Grund und Folge ergänzen.
5. Finding 4 und die übrigen Zeilen von §6: Ausgänge bei der Closure setzen (Nachzug *entfallen*, Lücke Vertrag/Code *entfallen*, Abweichungen 1 bis 3 mit der Antwort des Architect); §7 nennt „Binary unverändert, kein Katalog, kein Handbuch“.
6. Finding 12: die §6-Zeile zu `SPEC-109` Nr. 5 berichtigen (der Satz steht seit dem Merge von Teil 1 auf `main`, vollständig für `Exchanges`/`Session`, `Coverage`/`Snapshot` und `Diagnose`); die Sendung im Merge-Commit entfällt, der Integrationsschritt prüft allein, dass der zweite Merge `spezifikation.md` Zeile 856 und die Zeilen von `slice-mock-lesen` (Linie B, dieselbe Stelle, `MockUseCase`) nicht doppelt setzt. Die Zählprüfung der M-Codes hält mit dem 38. Code (belegt).
7. Finding 9: Zahl der Zeile M-Ausgeschlossen berichtigen.

An den **Architect**:

8. Finding 7: *Zweig-Menge*, Zahl `11 + n`, bestätigen.
9. Finding 8: `SPEC-236` Nr. 7, letzter Spiegelstrich, Nebensatz „sieht `{}` und `[]`“ berichtigen (nur Wortlaut; `Fehler` ist durch `AlleCodes()` nie leer).

Keine Rückgabe an den Implementer, kein MEDIUM, keine Rückführung des Slice. Die Fixture-Lücken der Findings 5 und 6 wären mit einem Wert mit Leerraum, Komma und Zeichen jenseits von ASCII sowie Kennungen über 64 Zeichen, mit NUL und mit Steuerzeichen in der Schleife von `TestDiagnoseEinstellungen` und `TestCoverageKennung` zu schließen; das ist bei Bedarf eine Zeile je Schleife.

## 8. Gegenlesen (`AGENTS.md` §3.18)

Dieser Report enthält keine Folge aus Backslash, `u` und vier Hex-Ziffern und kein Zeichen jenseits von ASCII außer Umlauten; er wurde nach dem Schreiben mit `LC_ALL=C grep -P` auf Rohbytes (Backslash-`u`-Hex-Folge, BOM, NBSP, Zeilentrenner) gegengelesen. Kennungen von ADRs stehen als Links.
