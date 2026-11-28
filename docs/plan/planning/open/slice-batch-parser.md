# Slice slice-batch-parser: Parser für $batch-Requests (Leser fremder Bytes), leerer Batch

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-expand-batch](../welle-expand-batch.md).

**Bezug:** `LH-FA-16`, `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-16.a` (Erweiterungspunkt `$batch` mit atomaren Changesets, eigener Parser-Contract-Test, unbekannte Operationen `501`) · `SPEC-049` (Lastenheft §5: `$batch` nicht im ersten Abnahmeumfang) · `SPEC-022` (`ODR-E6001`) · `SPEC-109` (Fehlermodell, Nr. 5 mit den Batch-Fehlern) · `SPEC-180` (`ODR-E6007`) · `SPEC-182`, `SPEC-183` (Vertrag dieses Slice) · `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkte).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Port `ODataBatch` der Codec-Gruppe liest den Body eines `$batch`-Requests im Format `multipart/mixed` des OData-V4-Batch (`SPEC-182`, `SPEC-183`) in typisierte Domain-Werte: eine Folge von Teilen, jeder Teil entweder ein Einzelrequest (Methode, Ziel, Header, Body-Bytes, `Content-ID`) oder ein Changeset (Folge von Requests). Ein leerer Batch ergibt eine gültige leere Folge, ein Verstoß einen Fehler mit Ursache (`ODR-E6007`), eine gültige, aber nicht unterstützte Form `501` mit `ODR-E6001`. Der Parser liest Bytes und entscheidet nichts über Zustand, Methode, Ziel oder Body der Einzelrequests (`LH-FA-16.a`: Parser-Contract-Test).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Prüfung der Struktur (`model.PruefeBatch` und `BatchRequest.Verweis`: Methoden je Ebene, `Content-ID` je Changeset eindeutig, Verweise nur rückwärts, `SPEC-184`) und der Validator-Contract-Test von `LH-FA-16.a` — `slice-batch-validator` (Schnitt vor dem Code, §4; dort als Übernahme dieses Slice); er liest die Domain-Werte und den Fehlerwert, die dieser Slice liefert, und ändert den Leser nicht.
- Ausführung der Einzelrequests, die atomare Workspace-Transaktion eines Changesets, die Antwort des Batch und die Route `POST /$batch` — `slice-batch-changesets` (Executor- und Serializer-Contract-Test von `LH-FA-16.a` liegen dort); dieser Slice liefert die gelesene Struktur.
- Parsen der URL und des Bodys eines Einzelrequests, Auflösung des Ziels gegen den Service-Root (auch absolute URL und fremder Host) und Auflösung eines Verweises `$<id>` — die Ausführung (`slice-batch-changesets`, dort als Sendung dieses Slice) nutzt dafür `urlparser` und den Leser der Client-Bodies aus `welle-mock`; der Parser hält Ziel und Body als Text und Bytes.
- JSON-Batch (`application/json`) und OData V2 `$batch` — nicht im Umfang (`SPEC-183` Nr. 1, `SPEC-173` Nr. 2); sie ergeben `501`.
- Aufzeichnen und Abspielen von Batch-Requests — der Record-Pfad leitet sie als HTTP unverändert weiter, der Replay-Pfad bedient sie unabhängig vom Parser (`welle-expand-batch` §6, `LH-FA-16.a`).
- Herstellerspezifische Batch-Erweiterungen, Actions und Functions in Batch-Teilen — Out-of-Scope von `LH-FA-16` bzw. Lastenheft §5; sie ergeben `501`.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Lesen der Hülle (`SPEC-182`, `SPEC-183`): Grenzzeichenfolge aus dem `Content-Type`, Teile, Changesets, Einzelrequest mit Methode, Ziel, Header, Body und `Content-ID`, Grenzen (Größe, Zahl der Teile und Kopfzeilen, Tiefe); ein leerer Batch ergibt die leere Folge; Ergebnis und Fehler hängen nicht davon ab, wie der Strom in Reads zerlegt ist.
- [ ] Ergebnis- und Fehlerwert (`SPEC-182`): Port `ODataBatch`, Domain-Werte, `BatchFehler` mit allen Ursachen von `SPEC-182` Nr. 4 (auch den vieren, die erst `slice-batch-validator` meldet) und beiden Sentinels, zwei Zeilen der Abbildung (`SPEC-109` Nr. 5).
- [ ] Randformen: je Zelle des Rasters (§6) ein Ort in der Spezifikation und ein Fänger der Tabelle (§3); Parser-Contract-Test (`LH-FA-16.a`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (die Route `/$batch` erreicht es erst mit `slice-batch-changesets`).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/model/batch.go`, `batch_fehler.go` (mit Tests) | neu | Domain-Werte, `BatchFehler` mit den Ursachen und Sentinels (`SPEC-182`) |
| `src/hexagon/ports/outbound/odata_batch.go` | neu | Port `ODataBatch` (`SPEC-182`) |
| `src/hexagon/ports/outbound/abbildung.go`, `abbildung_batch_test.go` | update, neu | zwei Zeilen der Abbildung (`SPEC-109` Nr. 5) |
| `src/adapters/driven/odata/batchcodec/` | neu | Leser (`SPEC-182`, `SPEC-183`) mit Tests |
| `spec/spezifikation.md`, `spec/architecture.md` | update | Vertrag mit Randformen, Port `ODataBatch` (Architect, mit diesem Plan geschrieben) |

`outbound.CodecSatz`, HTTP-Adapter und Composition Root bleiben unverändert; die Verdrahtung ist Sache von `slice-batch-changesets`. Die Tabelle *Zusage · Fänger* und das Raster in §6 stehen mit `SPEC-182` und `SPEC-183` vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen sind dort entschieden, der Implementer entscheidet keine.

- Der Leser liest fremde Bytes (ein Format, das dieses Repo nicht schreibt): die Produktcode-Erwartung des Architect ist ein fester Anteil der Schätzung (`SPEC-128`); `CRLF`, `LF`, Steuerzeichen und Bytes jenseits von ASCII in Fixtures entstehen aus Bytewerten (`AGENTS.md` §3.18).

**Entscheidungen des Architect** (Ort; keine neue ADR, kein neuer Code, keine offene Owner-Frage): Format nur `multipart/mixed`, JSON-Batch `501` (`SPEC-183` Kopf, Nr. 1) · eigener Leser ohne `mime/multipart` und `net/http`, keine Fremdbibliothek (`SPEC-182` Kopf) · Grenzen: Body `MAX_CAPTURE_BODY`, 1000 Teile, 100 Zeilen je Block, kein Changeset im Changeset (`SPEC-182` Nr. 7, `SPEC-183` Nr. 7) · `respond-async`, `continue-on-error` sind Präferenzen des Hüllen-Requests und gehören zur Ausführung, `dependsOn` gibt es nur im JSON-Batch (`format`) · Ursachen, Codes, Rangfolge, Ort als Teilpfad ohne Zeile (`SPEC-182` Nr. 4, 5) · Domain-Werte (`SPEC-182`, Signaturen).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Dateien unter `src/adapters/driven/odata/batchcodec/`, wo nichts anderes steht; die Zeilen 37 bis 43 der Fassung des Architect, die Prüfung, führt `slice-batch-validator` unter denselben Nummern; die Testnamen sind die des Architect, der Implementer fährt je Zeile die Mutation selbst):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | `ctx` nil, `src` nil, `grenze` unter 1: `ErrUngueltig` ohne `BatchFehler`, `src` ungelesen, auch bei beendetem Kontext | `TestBatchProgrammfehler`: jede Kombination der drei Verstöße × {Kontext offen, beendet} × `grenze` {0, -1, `MinInt64`}; Gegenprobe `grenze` 1 | `grenze` 0 wird angenommen | andere |
| 2 | Beendeter Kontext: sein Fehler `==`, ohne Code, vor dem ersten und jedem weiteren Read und nach dem letzten, auch bei ungültigem Inhalt | `TestBatchAbbruch`: Strom, der beim Read k abbricht, Schleife über jedes k bis zum Ende | Keine Prüfung nach dem letzten Read | andere |
| 3 | Hüllen-`Content-Type`: keine, zwei, leere Zeile, anderer Medientyp sind `content_type`; `multipart/mixed` in jeder ASCII-Schreibweise mit `SP`, `HTAB` am Rand wird gelesen; `src` bei Verstoß ungelesen | `TestBatchHuelle`: Schleife über Zahl der Zeilen (0 bis 3), Medientypen und Schreibweisen; Strom zählt Reads | Nur die erste Kopfzeile zählt | andere |
| 4 | `application/json` in jeder Schreibweise und mit Parametern ist `format`, `src` ungelesen | `TestBatchJSONFormat` | `application/json` ist `content_type` | andere |
| 5 | Grenzzeichenfolge: genau ein Parameter `boundary`, mit oder ohne `"`, 1 bis 70 Zeichen der Menge, letztes kein `SP`; sonst `boundary` | `TestBatchBoundary`: alle 256 Bytes an Anfang, Mitte, Ende; Länge 0, 1, 70, 71; fehlt, doppelt (gleich, verschieden), ohne `=`, einzelnes `"`; Schreibweise des Namens; fremde Parameter davor und dahinter | Der erste von zwei Parametern gewinnt | andere |
| 6 | Gelesen wird allein mit `Read` bis `io.EOF`, nie `Close`; Ergebnis und Fehler hängen nicht von der Zerlegung ab | `TestBatchZerlegung`: jede Zweiteilung, Einzelbytes, `0, nil` dazwischen, `n > 0` mit `io.EOF` über Beispiel und je eine Fixture je Ursache; Strom mit allen Methoden von `io`, jede gezählt | Nach dem ersten Read wird zerlegt | andere |
| 7 | Ein Fehler von `src` ist kein `BatchFehler`, bleibt über `errors.Is` erreichbar und geht jedem Verstoß des Inhalts vor | `TestBatchStromfehler`: Fehler (auch `ErrUnexpectedEOF`, umhülltes `EOF`) an jeder Stelle, mit und ohne Bytes, bei gültigem und ungültigem Inhalt | Fehler wird wie `EOF` behandelt | andere |
| 8 | `grenze` Bytes sind gültig, das nächste ist `zu_gross`, vor jedem Inhaltsverstoß und vor dem Fehler von `src` im selben Read; kein Read fordert mehr als `grenze + 1` abzüglich der gelieferten Bytes; `MaxInt64` gültig | `TestBatchGroesse`: beide Seiten, `grenze` {1, Länge-1, Länge, Länge+1, `MaxInt64`} × Blockgrößen, angeforderte Bytes gezählt | `>=` statt `>` | andere |
| 9 | Grenzzeile nur am Anfang der Ebene oder hinter `CRLF`, `--` und Zeichenfolge bytegleich, optional `--`, `SP`/`HTAB`, dann `CRLF` oder Ende; alles andere ist Inhalt | `TestBatchGrenzzeile`: Fast-Treffer (Suffix, führendes `SP`, hinter `LF`, hinter `CR`, andere Schreibweise, Füllung 0 bis 3) × öffnend und schließend × beide Ebenen | Präfix-Vergleich ohne Zeilenende | andere |
| 10 | Präambel und Epilog ohne Wirkung; die Ebene endet an der ersten schließenden Grenzzeile | `TestBatchPraeambelEpilog`: Epilog mit weiteren Teilen und Grenzzeilen, beide Ebenen | Grenzzeilen hinter der schließenden werden gelesen | andere |
| 11 | Ohne schließende Grenzzeile `rahmen`: leerer Body, nur Präambel, abgeschnitten an jeder Stelle, Zeilenenden nur `LF` | `TestBatchRahmen`: jeder echte Anfang des Beispiels vor dem Ende der schließenden Grenzzeile, `LF`-Fassung, beide Ebenen mit Ort | Fehlende schließende Zeile wird angenommen | andere |
| 12 | Erste Grenzzeile schließend: leerer Batch, `Teile` leer und nicht `nil` | `TestBatchLeer`: mit und ohne Präambel, `CRLF` am Ende, Füllung, Epilog | Kein Teil ist `rahmen` | andere |
| 13 | Inhalt eines Teils bytegenau: das `CRLF` vor der Grenzzeile gehört nicht dazu, davor stehende bleiben; unmittelbar folgende Grenzzeile gibt leeren Inhalt | `TestBatchInhalt`: Bodies mit 0 bis 3 `CRLF` am Ende, mit `NUL` und Bytes ab `0x80`; leerer Teil ist `content_type` | Ein `CRLF` zu viel oder zu wenig | andere |
| 14 | Kopfzeile `Name ":" Wert` mit `tchar`, Wert getrimmt; Fortsetzungszeile, Leerraum vor `:`, Steuerzeichen, `0x7F`, einzelnes `CR`/`LF`, ohne `:` sind `kopfzeile` | `TestBatchKopfzeile`: alle 256 Bytes im Namen und im Wert × Block des Teils und des Requests | Gefaltete Zeile wird angehängt | andere |
| 15 | Ein Block endet an der leeren Zeile oder am Ende des Inhalts | `TestBatchBlockEnde`: je Block mit und ohne leere Zeile, mit und ohne `CRLF` hinter der letzten Zeile | Ohne leere Zeile ist es `kopfzeile` | andere |
| 16 | 100 Zeilen je Block sind gültig, die 101. ist `zu_viele`, vor ihrer Form; je Block gezählt | `TestBatchZeilenzahl`: 99, 100, 101 × beide Blöcke; zwei Blöcke zu je 100 | Gezählt wird über Blöcke hinweg | andere |
| 17 | 1000 Teile über beide Ebenen gültig, 1001 `zu_viele` mit dem Ort der Ebene; der Rahmen der Ebene geht vor | `TestBatchTeilzahl`: Verteilungen (alle in der Hülle, in einem Changeset, auf mehrere) beidseits der Grenze; 1001 Teile ohne schließende Zeile ist `rahmen` | Grenze je Ebene statt zusammen | andere |
| 18 | `Content-Type`, `Content-Transfer-Encoding`, `Content-ID` höchstens einmal je Teil, jede Schreibweise; andere dürfen sich wiederholen | `TestBatchDoppelteKopfzeile`: Schleife über die drei × Schreibweise × Abstand × beide Ebenen; fremde Namen doppelt | Der letzte Wert gewinnt | andere |
| 19 | `Content-Type` des Teils: fehlt oder leer `content_type`; `application/http` Request; `multipart/mixed` in der Hülle Changeset, im Changeset `verschachtelt`; sonst `teil_typ` | `TestBatchTeilTyp`: Schleife über Werte × Schreibweise × Parameter × beide Ebenen | Geschachteltes Changeset wird gelesen | andere |
| 20 | `Content-Transfer-Encoding` fehlt, `binary`, `7bit`, `8bit` gültig; sonst `transfer_kodierung` | `TestBatchTransferKodierung`: Werte (auch leer, `base64`, `quoted-printable`) × Schreibweise × Teilart | `base64` ohne Wirkung | andere |
| 21 | `Content-ID` eines Requests: 1 bis 128 Zeichen der Menge, sonst `content_id`; im Changeset-Teil und in den Kopfzeilen des Requests ohne Wirkung | `TestBatchContentID`: alle 256 Bytes, Länge 0, 1, 128, 129, drei Orte | Form ungeprüft | andere |
| 22 | Anfragezeile `Methode SP Ziel SP HTTP/1.1`, sonst `anfragezeile` (auch Teil ohne Request); Methode und Ziel wie geschrieben | `TestBatchAnfragezeile`: 0 bis 3 `SP`, `HTAB`, andere Versionen, leere Zeile davor, alle 256 Bytes in Methode und Ziel, Ziel relativ, absolut, fremder Host | Zerlegung an beliebigem Leerraum | andere |
| 23 | `Header` in Reihenfolge, Namen wie geschrieben, Doppelte bleiben; `Body` jedes Byte hinter der leeren Zeile, sonst `nil`; `Content-Length` ohne Wirkung | `TestBatchRequest`: falsche und richtige `Content-Length`, Bodies mit Grenzzeilen-Fast-Treffern und `CRLF CRLF` | Body wird an `Content-Length` gekürzt | andere |
| 24 | `Transfer-Encoding` im Request ist `transfer_kodierung`, nach dem ganzen Block | `TestBatchRequestTransferEncoding`: Schreibweise, Werte, Stelle; dahinter eine `kopfzeile` gewinnt | Die Kopfzeile wird durchgereicht | andere |
| 25 | Changeset: Grenzzeichenfolge nach den Regeln der Hülle (`boundary`, `rahmen` mit Ort des Changeset); die Hülle wird zuerst zerlegt | `TestBatchChangeset`: Menge aus Zeile 5 am Changeset; gleiche Zeichenfolge wie die Hülle; Zeichenfolge als Anfang der anderen | Der Changeset-Rahmen darf offen enden | andere |
| 26 | Ein Changeset ohne Teil ist gültig, `Requests` leer und nicht `nil` | `TestBatchChangesetLeer` | Leeres Changeset ist `rahmen` | andere |
| 27 | Teile und Requests in Reihenfolge des Bodys, ein Einzelteil mit genau einem Request, Slices bei Erfolg nie `nil`; bei Fehler der Nullwert | `TestBatchErgebnis`: gemischte Folgen der Länge 0 bis 4 aus {Request, leeres Changeset, Changeset mit 1 und 2}; Fehler hinter gültigen Teilen | Teilergebnis bei Fehler | andere |
| 28 | Jede Ursache trägt ihre Klasse, ihren Code, genau ihren Sentinel; `Error()` und `Ort` nennen kein Byte der Eingabe | `TestBatchFehlerform` (`model`): Schleife über `AlleBatchUrsachen` und Orte; im Leser-Test unverwechselbare Texte in jeder Verstoß-Fixture | Falsche Klasse; Zeile in der Meldung | andere |
| 29 | `ErrBatchUngueltig` bildet auf `ODR-E6007`, `ErrBatchNichtUnterstuetzt` auf `ODR-E6001` ab, jeder Gegenstand | `TestAbbildungBatch` (`ports/outbound`), `TestAbbildungVollstaendig` | Eine Zeile fehlt | andere |
| 30 | Rangfolge des Lesers (`SPEC-182` Nr. 5): Ebene vor Teil, Teil vor Teil, Folge im Teil, kein Vorrang der Klasse | `TestBatchRang`: alle Paare von Verstößen an verschiedenen Teilen in beiden Reihenfolgen und an denselben Teil nach der Folge der Nummer | Klasse `ungueltig` hat Vorrang | andere |
| 31 | Gleiche Eingabe, gleiches Ergebnis; das Ergebnis gehört dem Aufrufer, kein Wert teilt Speicher mit einem anderen | `TestBatchEigentum`: `append` und Überschreiben an jedem `Body`, `Header`, `Requests`, Vergleich aller übrigen und eines zweiten Aufrufs | `Body` als Teilstück ohne Kapazitätsgrenze | andere |
| 32 | Gleichzeitige Aufrufe auf einer und auf mehreren Instanzen beeinflussen sich nicht | `TestBatchGleichzeitig`: 4 Goroutinen, Beispiel und Verstöße, Vergleich mit dem Einzelaufruf; `make test-race` | Puffer als Feld des Lesers | nebenläufig |
| 33 | Zuteilung höchstens 256 · n + 1 MiB | `TestBatchWaechter`: Familien (viele Grenzzeilen, Kopfzeilen, Fast-Treffer, Präambel, langer Body) bis 8 MiB | Verkettung in einer Schleife | andere |
| 34 | Jede Folge über dem Alphabet des Rahmens hat die Klasse eines Orakels ohne Code des Lesers | `TestBatchAlphabet`: alle Folgen bis Länge 6 über {Grenzzeilen-Anfang `--b`, `--`, `CRLF`, `CR`, `LF`, `SP`, `x`}; Prüfgrenze Länge 6, Folge: eine Regel erst ab Länge 7 bleibt ungefangen | `LF` gilt als Zeilenende | andere |
| 35 | Das Beispiel von `SPEC-183` Nr. 8 ergibt die genannten Werte des Lesers und die Zeilen des Lesers seiner Tabelle das Genannte (Parser-Contract-Test; das genannte `nil` von `PruefeBatch` steht in `slice-batch-validator` Zeile 43) | `TestBatchContract`: Beispiel als Fixture, je Zeile eine Änderung | Eine Antwort wird vertauscht | andere |
| 36 | Akzeptiertes Negativ: die Länge einer Zeile hat keine eigene Grenze | kein Fänger; Grund: `grenze` begrenzt den Body, das Ziel begrenzt sein Leser; Folge: eine Kopfzeile bis `grenze` Bytes ist gültig | — | andere |

**Teilung des Produktcodes** (Planner, 2026-11-28, mit dem Schnitt): von den 600 Zeilen, die der Architect für beide Teile erwartete, stehen 500 hier und 100 in `slice-batch-validator`. Der Architect nannte keine Teilung; die Zahl ist eine Setzung des Planner und er darf sie vor dem ersten Code-Commit ändern.

```groessenmarke
pfad: src/adapters/driven/odata/batchcodec/
pfad: src/hexagon/model/batch.go
pfad: src/hexagon/model/batch_test.go
pfad: src/hexagon/model/batch_fehler.go
pfad: src/hexagon/model/batch_fehler_test.go
pfad: src/hexagon/ports/outbound/odata_batch.go
pfad: src/hexagon/ports/outbound/abbildung.go
pfad: src/hexagon/ports/outbound/abbildung_batch_test.go
produktcode: 500
fakes: 0
schnitt-grund: Die Prüfung ist als slice-batch-validator abgetrennt (liest nur Domain-Werte und Fehlerwert). Der Rest teilt in einem Mechanismus den Ergebniswert des Ports, den Fehlerwert, die Rangfolge und die Zerlegung des Rahmens, der Kopfzeilen und des Einzelrequests; keiner davon liefert allein einen Ergebniswert.
```

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): keine Slice-Abhängigkeit über `done/` hinaus. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** keine auf einen offenen Slice. Der Leser braucht das Fehlermodell (`SPEC-109`, `slice-fehlermodell-konfiguration` in `done/`), die Konvention der Codec-Ports (`SPEC-125`, `SPEC-126` als Vorbild für Zerlegung des Stroms, Grenzen und Fehler, `done/`) und das Domain-Modell; er ruft weder `urlparser` noch Workspace, Repositories, Interpreter oder `MockUseCase`. Der Welle-Trigger (`welle-mock` in `done/`) ist für diesen Slice Reihenfolge-Vorsicht (Priorisierung nach den Muss-Wellen), keine Code-Abhängigkeit: ein Start davor ist eine Entscheidung des Users, nicht dieses Plans.
- **Was der Slice nicht braucht — und warum:** `MockUseCase` und `WithWorkspaceTransaction` (die Ausführung liegt in `slice-batch-changesets`), die Parser der OData-URL und der Client-Bodies (Ziele und Bodies bleiben hier Text und Bytes), HTTP-Adapter, Composition Root, M2 und `slice-bootstrap-m3` (das Binary ändert sich erst mit der Route), `slice-expand-parser` und die anderen Wellen. Das heißt: **unabhängig von `welle-mock` und von M2**, soweit der Code betroffen ist.
- **Größenmarke** (`AGENTS.md` §3.14, `SPEC-128` Nr. 1; die Faktoren stehen allein dort; Planner, 2026-11-28, nach dem Schnitt): gezählt ist die Tabelle in §3 mit 36 Zeilen, *a* = 1 nebenläufige (Zeile 32), *b* = 35 andere (das akzeptierte Negativ zählt als `andere`), *f* = 0, *p* = 500. *S* = 3385, Marke *M* = **4600**. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Block in §3 trägt `basis`, `schaetzung` und `marke` noch nicht (`make groessenmarke` meldet *Marke steht aus*); der Planner setzt sie beim Start, vor dem ersten Code-Commit, in eigenem Commit. Ändert der Architect bis dahin Zeilen oder `produktcode`, rechnet der Planner neu.
- **Schnittfrage** (*S* ≥ 2223): gestellt und vor dem Code beantwortet (Planner, 2026-11-28). Ungeschnitten waren es 43 Zeilen, *p* = 600, *S* = 4045, Marke 5500. (1) Abgetrennt ist `slice-batch-validator` (Zeilen 37 bis 43; *S* = 660, Marke 900): reine Funktionen im Domain-Modell, eigene Tests und eigener Contract-Test, kein gemeinsamer Pfad. Er teilt mit dem Leser die Domain-Werte und den Fehlerwert, die dieser Slice liefert, und die Regel der Form einer `Content-ID` (§6). (2) Der Leser bleibt über der Schnittgrenze und ohne weiteren Schnitt, Grund im Block (`schnitt-grund`): Rahmen, Teile, Kopfzeilen und Einzelrequest liefern erst zusammen den Ergebniswert des Ports. Geprüft ist auch der Vorab-Slice nur für Fehlerwert und Abbildung (Zeilen 28, 29): ohne Erzeuger kein Lieferwert, und der Rest läge weiter über der Schnittgrenze.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 4600 Zeilen (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Eine Überschreitung schneidet nicht (`SPEC-128` Nr. 10 c); sie bekommt einen Eintrag `ueberschreitung` mit Grund im Block.
- `in-progress` → `open` (blockiert — Carveout?): eine Randform fehlt in `SPEC-182` oder `SPEC-183`, oder der Leser braucht entgegen `SPEC-182` eine Fremdbibliothek (dann ADR).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Menge der Hüllen ist unbeschränkt (jede Bytefolge, jede Zerlegung des Stroms in Reads) — **Ausgang:** offen; Generator mit Orakel und Prüfgrenze (Zeilen 6 und 34, `AGENTS.md` §3.17).
- Das Verhalten einer Bibliothek als Grundlage einer Zusage (Zeilenenden, Falten, doppelte Grenze) — **Ausgang:** offen; entschieden: eigener Leser (`SPEC-182` Kopf), Zeilen 5, 9, 11, 14; der Review prüft die Importe.
- Größe eines Batch-Bodys: zwei Quellen der Grenze — **Ausgang:** offen; entschieden: `MAX_CAPTURE_BODY`, der HTTP-Adapter begrenzt nicht (`SPEC-182` Nr. 7), Zeile 8.
- Zusammenspiel mit Prüfung und Ausführung: Form der Domain-Werte und des Fehlerwerts — **Ausgang:** offen; `slice-batch-validator` §1 liest sie, `slice-batch-changesets` §1 führt die Sendung.
- Die Regel der Form einer `Content-ID` (`SPEC-183` Nr. 4 c) braucht der Leser (Zeile 21) und `Verweis` in `slice-batch-validator` (Zeile 39); ob sie an einem Ort im Domain-Modell steht oder zweimal, sagt der Vertrag nicht, und Zeile 39 führt nur die Längengrenzen, nicht das Alphabet (`AGENTS.md` §3.17) — **Ausgang:** offen; Adresse Architect, vor dem ersten Code-Commit dieses Slice.
- Der Strom des Tests hält jedes Argument fest und zählt jede Operation, auch `Close` und jeden Read (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×); Ports ruft der Leser keine — **Ausgang:** offen; Zeilen 3, 6, 8.
- Randformen (`AGENTS.md` §3.12, §3.16) — **Ausgang:** offen; Raster unten, eine Zelle ohne Ort geht an den Architect.
- Schnittfrage (`AGENTS.md` §3.14): die Formel lag über der Schnittgrenze — **Ausgang:** entfallen: entschieden (Planner, 2026-11-28): geschnitten nach der Empfehlung des Architect, die Prüfung ist `slice-batch-validator`; der Leser bleibt mit `schnitt-grund` (§4).

**Raster *Operation × Argumentart*:**

| Operation | Body fehlt oder leer | Grenze fehlt, leer oder ungültig | unbekannter Teiltyp oder Methode | fremde Form (JSON, absolute URL, anderer Host) | über Grenze (Größe, Teile, Tiefe) | Wiederholung (`Content-ID`, Changeset) | Abbruch (Kontext, Strom) |
|---|---|---|---|---|---|---|---|
| Parser: Hülle und Teile | `SPEC-183` Nr. 2 (`rahmen`, leerer Batch) | `SPEC-183` Nr. 1, Nr. 5 | `SPEC-183` Nr. 4 a; Methode gilt nicht: sie gehört zum Einzelrequest | `SPEC-183` Nr. 1 (`format`); URL und Host gelten nicht: die Hülle hat kein Ziel | `SPEC-182` Nr. 2, `SPEC-183` Nr. 7 | `SPEC-183` Nr. 1 (Parameter doppelt), Nr. 4 (Kopfzeile doppelt), Nr. 5 (gleiche Zeichenfolge) | `SPEC-182` Nr. 1, 2 |
| Parser: Einzelrequest (Zeile, Header, Body) | `SPEC-183` Nr. 6 (`anfragezeile`), `SPEC-182` Nr. 3 (`Body` `nil`) | gilt nicht: ein Request hat keine Grenzzeichenfolge; sein `Content-Type` ist ein Header (`SPEC-183` Nr. 6) | `SPEC-183` Nr. 6 (jedes Token; das Urteil über die Methode liegt nicht hier, `SPEC-182` Kopf) | `SPEC-183` Nr. 6 (Ziel ist Text), Auflösung `SPEC-182` Kopf (nicht hier) | `SPEC-183` Nr. 7 (Zeilen je Block, Länge) | `SPEC-183` Nr. 4 c, Nr. 6 (doppelte Header) | gilt nicht: zerlegt wird nach dem ganzen Strom ohne Kontext (`SPEC-182` Nr. 1, 2) |

Die Zeilen des Rasters zur Prüfung der Struktur führt `slice-batch-validator` §6.

**Zusatztabelle** (je Paar von Verstößen; `AGENTS.md` §3.16). Eine Zeile je Modus gilt nicht: der Leser kennt keinen Modus (`SPEC-182`). Die Paare mit der Prüfung führt `slice-batch-validator` §6.

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| ungültige Grenzzeichenfolge der Hülle · ungültiger Teil | `boundary`, `src` ungelesen (`SPEC-182` Nr. 5 c) | Zeilen 5, 30 |
| ungültige Grenzzeichenfolge eines Changeset · Verstoß in seinen Teilen | `boundary` (`SPEC-182` Nr. 5 g) | Zeilen 25, 30 |
| Größe über der Grenze · Verstoß davor oder dahinter | `zu_gross` (`SPEC-182` Nr. 2) | Zeile 8 |
| fehlende schließende Grenzzeile · Verstoß in einem Teil | `rahmen` (`SPEC-182` Nr. 5 f) | Zeile 30 |
| fehlende schließende Grenzzeile · zu viele Teile | `rahmen` (`SPEC-182` Nr. 5 f) | Zeile 17 |
| zwei Verstöße in verschiedenen Teilen | der frühere Teil (`SPEC-182` Nr. 5 f) | Zeile 30 |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für Fakes und Hüllen von Strömen (jede Schnittstelle des Originals, jeder Read gezählt); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Code des Slice ist nicht nebenläufig; allein Zeile 32 der Tabelle hat einen nebenläufigen Fänger. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
