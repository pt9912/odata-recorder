# Slice slice-replay-service: ReplayUseCase: Zuordnung, Strategien, Cursor

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-04`, `LH-FA-05`, `LH-FA-09`, `LH-FA-14`, `LH-QA-02`, `LH-QA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-04.a`, `LH-FA-04.b`, `LH-FA-04.c` · `LH-FA-09.a` · `SPEC-129` Nr. 11 a · `SPEC-052` · `SPEC-045` · `LH-FA-14.a` (Redaction und Match-Key) · `SPEC-115` Nr. 3 (`Location` mit Marker, Entscheidung des Owners) · `SPEC-118` · `SPEC-009`, `SPEC-078`, `SPEC-083` · `SPEC-020`, `SPEC-021`, `SPEC-050`, `SPEC-065`, `SPEC-070`, `SPEC-071`, `SPEC-074`, `SPEC-093`, `SPEC-097` · `ARC-002`, `ARC-003` · `architecture.md` §4.3 · `SPEC-135` (der Vertrag) · `SPEC-114` Nr. 9 · `SPEC-109` Nr. 5 · `SPEC-116` · `SPEC-118` Nr. 10 · `SPEC-129` Nr. 11 a · `SPEC-117` · `SPEC-020`, `SPEC-050` (Katalog). `LH-FA-14.a` und `SPEC-045` stehen schon.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-replay-service)

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `ReplayUseCase` prüft die Zuordnung aus Workspace, Recording-Session und Replay-Session, bildet den Match-Key aus dem redigierten Request, wählt je Strategie (`sequential`, `strict`, `first`, `last`) einen replayfähigen Exchange, schreibt den Cursor erst nach erfolgreicher Auswahl atomar fort und liefert Status, Header und Body ohne Zugriff auf einen Upstream.

**Übernimmt:** Die Übergabe aus `welle-persistenz`: die Abbildung von `ErrBeschaedigt` und `ErrBlob*` auf `ODR-E3001` beim Ausliefern eines beschädigten Blobs (die Tabelle liefert `slice-fehlermodell-konfiguration`).

**Übernimmt (Redaction, Match-Key, Replayfähigkeit):** Der Dienst ruft `BildeSchluessel` (`SPEC-116`) mit dem eingehenden Request, dem durch den Redaktor geleiteten Request-Body (die Funktion redigiert ihn nicht, `SPEC-116` Einleitung) und der Konfiguration des Records, lässt bei `ErrKodierung` (`reject`) `ODR-E5006` ohne Cursor-Fortschritt folgen, wendet `Replayfaehig` (`SPEC-117`) als Prädikat über die Kandidaten an und liefert Header über `OhneRedigierte` (`SPEC-115` Nr. 5); `Location` mit Marker im Query-Wert wird mit `[REDACTED]` ausgeliefert (`SPEC-115` Nr. 3; die Lesart von `SPEC-045` bleibt, Entscheidung des Owners vom 2026-10-21: Weglassen bräche Weiterleitungen und `201 Created`; die Sätze in `SPEC-045` und `LH-FA-14.a` stehen, `SPEC-135` Nr. 13; Vorschlag, Owner bestätigt); der Key eines Bodys, der als JSON ungültig ist, folgt im Modus `json-canonical` dem Byte-Hash (Art 2), kein Fehler (`SPEC-116` Nr. 6) (Sendung von `slice-redaction-match-key`).
**Übernimmt (Redaktor):** Der eingehende Request-Body geht durch `RedigiereBody` (`SPEC-118`) mit `MaxBytes = MAX_CAPTURE_BODY`, `Metadata=false` und den wirksamen `Pfade`, der Key wird aus der Ausgabe gebildet, nie aus den Eingabebytes (`SPEC-118` Nr. 10, `SPEC-116` Einleitung); jeder Fehler des Redaktors (auch ein Body über der Grenze oder mit unbekanntem Typ) lässt `ODR-E5001` ohne Cursor-Fortschritt folgen; ausgenommen ist ein Kontextfehler (Abbruch des Clients), der ohne Code zurückkehrt, und ein `Content-Encoding` außer `identity` ist `ODR-E5001` ohne Lesen (`SPEC-135` Nr. 7), denn der Key lässt sich nicht bilden und ein `null`-Key passt nie (`SPEC-001`); ein Request ohne Body (0 Bytes) ist kein Fehler (Sendung von `slice-streaming-redaction`).
**Übernimmt (Record-Seite):** Der Record bildet den Key aus den Headern des Requests, wie empfangen (auch `X-Recorder-*` und Hop-by-Hop; `MATCH_VARY_HEADERS` wählt, was eingeht); der Replay bildet ihn aus den Headern seines Requests ebenso, sonst ergibt ein als Vary-Header konfigurierter Name verschiedene Keys (`SPEC-121` Nr. 6). Gespeicherte Response-Header sind die empfangenen ohne Hop-by-Hop, redigiert; `Content-Encoding` und `Content-Length` beschreiben die Leitung, der Blob den entpackten, redigierten Body (`SPEC-119` Nr. 6, `SPEC-121` Nr. 6). Ein Exchange ohne Antwort (Status `0`, `SPEC-121` Nr. 11) ist nie `complete` und nie Kandidat (Sendung von `slice-record-service`).
**Übernimmt (Workspace und Cursor):** Der Replay liest den Workspace über `WorkspaceRepository` (in der Transaktion `stores.Workspaces`), nicht über `WorkspaceUseCase`, und liest, wählt und schreibt den Cursor in **einer** `WithWorkspaceTransaction` (`Cursors.Get`, Auswahl, `Cursors.Put`); ein Reset liegt dann vollständig vor oder nach dem Fortschreiben, und `position` trägt nie den Wert von vor dem Reset plus 1 (`SPEC-129` Nr. 11 a, `LH-FA-09.a`: serialisiert, nicht abgewiesen). Die Revision dient dem Replay nicht als Sperre (`Advance` ruft er nicht) (Sendung von `slice-workspace-use-case`; vom Planner beim Start gelesen: §1 *Ausdrücklich NICHT* und DoD Punkt 2 treffen sie nicht).
**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- HTTP-Eingang, Lesen der Steuerheader aus dem Request und Entfernen der `X-Recorder-*`-Header — `slice-http-adapter` reicht die Header unverändert, der Use Case entnimmt die Zuordnung (`SPEC-052`).
- Hybrid-Fallback aus dem Mock auf Replay — Hybrid, nicht in der Welle; der Port `ReplayUseCase` ist der, den Hybrid später ruft.
- Workspace-Anlage und Reset — `slice-workspace-use-case`; Replay liest den Workspace und die Cursors.

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

- [ ] Zuordnung und Validierung: Pflichtheader und Fehler `ODR-E5003`, `ODR-E5004`, `ODR-E7001`, `ODR-E7003`, `ODR-E7009` ohne Cursoränderung; Cursor-Identität `(workspace, replay_session, recording_session)`; neue Replay-Session startet bei Position 0. Form der Replay-Session-ID (`SPEC-135` Nr. 5).
- [ ] Auswahl: Match-Key aus dem redigierten Request (Header, Query, Body über den Redaktor), vier Strategien, `ODR-E5001`, `ODR-E5002`, `ODR-E5007`, `ODR-E5008`, Replayfähigkeit nach `SPEC-045`, atomares Fortschreiben des Cursors (Lesen, Auswahl und Schreiben in einer `WithWorkspaceTransaction`, Workspace über `WorkspaceRepository`, `SPEC-129` Nr. 11 a; Sendung von `slice-workspace-use-case`), unabhängige Cursors paralleler Replay-Sessions.
- [ ] Auslieferung: Body aus dem Blob mit Integritätsprüfung (`ODR-E3001`), Header nach `SPEC-135` Nr. 13 (redigierte, `Content-Length`, `Content-Encoding` und `VOLATILE_HEADERS` entfallen; `Location` mit Marker ausgeliefert), `X-Recorder-Match-Redacted: true` bei `key_redacted`; der Service hat keinen Upstream-Port und keinen Import des Upstream-Adapters.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: Der Slice liefert kein Verhalten, das ein Anwender des gebauten Binary sieht (der Dienst ist erst über den Composition Root von `slice-bootstrap-m2` und den Adapter von `slice-http-adapter` erreichbar), und bringt keinen Fehlercode, Schlüssel oder Kommando in den Katalog; deshalb kein Kapitel. Was ein Anwender später sieht, geht als Sendung an `slice-handbuch-record-replay` (§6).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/inbound/` | update | `ReplayUseCase` |
| `src/hexagon/services/replay/` | neu | Zuordnung, Auswahl, Cursor, Auslieferung |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, Header-Policy der Auslieferung, Verhalten bei Abbruch |
| Tests | neu | Fälle von `LH-FA-04` (Beispiel A, B, A; zwei parallele Sessions; Reset), je Strategie und je Fehlercode; Test, dass `ReplayUseCase` keinen Upstream-Typ importiert |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstanden mit dem Vertrag `SPEC-135` in der Spezifikation vor dem ersten Code-Commit (`AGENTS.md` §3.12): die Tabelle und die Zusatztabelle stehen weiter unten in diesem Abschnitt, das Raster in §6.

- Der Dienst arbeitet mit `WorkspaceRepository` und `WithWorkspaceTransaction` (Ports aus `SPEC-108`), nicht mit `WorkspaceUseCase`; der Fänger gegen den Reset ist ein Wirkungs-Fänger, der Fortschreiben und Reset gegeneinander ordnet (der Fake der Transaktion aus `services/workspace/fakes_test.go` darf mitbenutzt werden, wo er die Zusage trägt; Wirkungs-Fänger Z50 bis Z54 des Gebers sind das Muster) — die Zeilen Z46 bis Z54 der Tabelle *Zusage · Fänger* sind angelegt (Sendung von `slice-workspace-use-case`, `SPEC-129` Nr. 11 a).
- Der Request-Body wird für den Match-Key im Modus `bytes` oder `json-canonical` durch denselben Redaktor geleitet wie beim Record; die Grenze für die Größe des Request-Bodys im Replay ist `MAX_CAPTURE_BODY` (`SPEC-135` Nr. 7).

#### Aufträge und Entscheidungen des Architect

**Reihenfolge der Rollen nach dem Start** (jede erst, wenn die vorige übergeben hat, jede in frischem Kontext): Architect (Vertrag `SPEC-135`, Tabelle, Raster, Block mit `pfad`, `produktcode`, `fakes`; erledigt) → Planner (Basis, Schätzung und Marke im Block, eigener Commit, vor dem ersten Code-Commit, mit `make groessenmarke` auf dem dann geltenden Stand von `SPEC-128`) → Implementer → Reviewer → Fixes (Klasse, `AGENTS.md` §3.17) → Verifier → Planner (Closure). Der letzte Commit des Architect ist die `<Basis>` der Marke. Hält der Implementer an einer Zelle ohne Ort an (`AGENTS.md` §3.12), geht sie an den Architect, und §6 trägt sie im selben Commit mit Ausgang *offen*.

**Fixtures und Fakes** (`AGENTS.md` §3.15): Kennungen, Dienst-Kennungen, Header-Werte und `SequenceNo` unterscheiden sich paarweise (`wsp-a`, `rep-a`, `ses-a`, `svc-a` sind vier verschiedene Texte), damit eine Vertauschung den Wert ändert. Der Fake jedes Ports hält jedes Argument fest, zählt jede Operation (auch die der Dienst nie rufen darf), speist an jeder Operation einen Fehler ein und umhüllt den Strom des Blobs und das `ctx` mit einer Hülle, die jede Schnittstelle des Originals trägt und je Aufruf zählt. Fakes: (1) `Sessions`, (2) `Exchanges`, (3) `Blobs` mit Hülle, (4) der Store mit Transaktion (Workspaces, Entities, Cursors, Snapshots, FIFO-Schlange, Haltepunkt nach dem k-ten Schritt, Eindringling, Rollback, Fehler beim Bestätigen); der Store folgt `services/workspace/fakes_test.go` als Muster (kopiert; Testdateien sind nicht importierbar). `FakeTelemetry` (`supporttest`) besteht und zählt nicht als Fake. Rohbytes (BOM, `0xFF`, CRLF) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18).

**Nebenläufige Fänger** (Z46 bis Z54): Sollwert ist jede Mutation in n von n Läufen rot, weil die Ordnung durch Kanäle und die FIFO-Schlange des Stores bestimmt ist. Der Implementer misst `-count=50` je Fänger und `make test-race` je Prozess (`-test.count=1`, 50 Prozesse) und trägt Zahl der Prozesse, Raten und Messenden in einen Absatz von §3 ein (`AGENTS.md` §3.10). Die Fenster-Mutationen (die Transaktion aufteilen) sind keine Datenrennen: jeder Zugriff liegt unter der Sperre, der Detektor meldet sie nicht (`SPEC-122` Nr. 8 b); sie fängt allein der Wirkungs-Fänger.

**Entscheidungen des Architect** (Alternative, Grund und Folge stehen im Vertrag):
1. Vertrag `SPEC-135`, 18 Nummern; kein neuer Code (33 bleiben, die Zählstellen sind unberührt).
2. Rangfolge der Prüfungen und Ort jeder Prüfung (Nr. 3); der Workspace wird erst in der Transaktion gelesen (Nr. 6 b, Folge benannt).
3. Form der Replay-Session-ID (Nr. 5) und `ODR-E5003` für jede ungültige Form.
4. Request-Body: immer durch den Redaktor mit `MaxBytes`, Ausgabe auf `MaxBytes` begrenzt, `Content-Encoding` außer `identity` ist `ODR-E5001`, Kontextfehler ohne Code (Nr. 7).
5. Auswahl je Strategie mit `ODR-E5001` vor allen anderen (Nr. 9).
6. Cursor: entsteht beim ersten Fortschreiben, `first` und `last` schreiben nie, der Cursor entscheidet über die Strategie (Nr. 10).
7. Transaktion: Öffnen des Blobs in `f` vor dem Schreiben; Folge für den Schreiber benannt (Nr. 11).
8. Header der Auslieferung als Policy (Nr. 13), `X-Recorder-Match-Redacted` aus dem Feld des Exchange.
9. Die Auswahl schließt die Auslieferung nicht ein (Nr. 14).
10. Ein Protokolleintrag über den Port `Telemetry` (Nr. 16), keine Fehlerzählung im Dienst.
11. Nummer `SPEC-135`; keine ADR, keine `ARC-*`.

#### Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15; 64 Zeilen, davon 9 nebenläufig und 55 andere, darunter 6 Negative)

Mengen sind Produkte ihrer Achsen; „leer“ und „fehlend“ sind verschiedene Elemente; Gleichstand zweier verglichener Werte und beide Seiten jeder Grenze sind Elemente. Eine Zeile mit Menge ist eine Schleife über alle ihre Elemente, mit unabhängigem Orakel (Handtabelle des Vertrags, nicht der Code). Layout-Alphabet der Auswahl: `A` = Key A und replayfähig, `a` = Key A, `capture_status=failed`, `B` und `b` ebenso für Key B; angefragt werden die Keys A, B und C (nie aufgezeichnet).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z1 | `New` liefert `model.ErrUngueltig` und keinen Dienst bei: keinem Dienst, leerer Kennung, leerem Root, nil in einem der fünf Felder, `MaxBytes` kleiner 1, unbekannter `Strategie`, `InvalidEncoding`, `BodyMode`, einem Nullwert in `Pfade` (Nr. 1) | Schleife über 18 Verstöße (1+1+1+5+2+3+2+2+1) und 5 Gutfälle (vier Strategien, `MaxBytes`=1); je Element Dienst nil und `errors.Is(err, ErrUngueltig)` | Eine Prüfung entfällt (je Element rot); `MaxBytes` ≤ 1 statt < 1 (Gutfall rot); Strategie ohne Beachtung der Schreibweise (`Sequential` rot) | andere |
| Z2 | `New` ruft keinen Port (Nr. 1) | Zähler aller Fakes nach `New`: 0 über alle 39 Operationen (Sessions 6, Exchanges 3, Blobs 5, Workspaces 5, Entities 5, Cursors 4, Snapshots 5, Transaktion 1, Telemetrie 3, Span 2) | `New` ruft `Telemetrie.Log` oder `Sessions.List` | andere |
| Z3 | Argumente (Nr. 2): `ctx` nil, `Methode` leer, `Laenge` -2 (mit und ohne Body), `Laenge` 1 bei `Body` nil sind `model.ErrUngueltig` ohne `Fehler`, ohne Code, kein Port, `Body` ungelesen; `Laenge` -1 und 0 mit und ohne Body sind keine Verstöße | Menge 10: ctx nil (1), Methode leer (1), `Laenge` {-2, -1, 0, 1} × Body {nil, nicht nil} (8); 5 Verstöße, 5 Gutfälle; Zähler aller Fakes und gelesene Bytes des Bodys | `Laenge` < 0 statt < -1; `Laenge` ≥ 0 statt > 0 bei nil (Gutfall rot); Body gelesen vor der Prüfung | andere |
| Z4 | Jeder Aufruf eines Ports erhält das `ctx` des Aufrufs unverändert, auch in `f` (Nr. 2, 18) | Menge 8: Stellen S1, S3, S4, S5, S6, S7, S8, S10; die Hülle um das `ctx` trägt einen Marker und prüft Identität an jedem Aufruf | `context.Background()`, `WithoutCancel` oder ein abgeleitetes `ctx` an einer Stelle (je Element rot) | andere |
| Z5 | Ein Kontextfehler eines Ports kehrt unverändert und ohne Code zurück, auch neben einem Fehler der Tabelle; danach kein weiterer Port; ein geöffneter Blob wird geschlossen (Nr. 2, 12) | Menge 28: Stelle {S1, S3, S4, S5, S6, S7, S8} × Fehler {`Canceled`, `DeadlineExceeded`, `Join(Canceled, ErrNichtGefunden)`, `Join(ErrNichtGefunden, DeadlineExceeded)}`; erwartet `errors.Is`, kein `*Fehler`, Zähler der späteren Stellen 0, `Close` des Stroms bei S8 genau 1 | Abbildung vor der Prüfung des Kontextes (`ODR-E7003` statt Kontextfehler); Kontextfehler verschluckt; Strom nicht geschlossen | andere |
| Z6 | Die Fehler der Ports werden nach `SPEC-109` Nr. 5 mit dem Gegenstand je Stelle abgebildet (Nr. 15, 18), `ErrNichtGefunden` an S6 ist der fehlende Cursor | Menge 63: Stelle {S1, S3, S4, S5, S6, S7, S8} × Fehler {`ErrNichtGefunden`, `ErrBeschaedigt`, `ErrBlobFehlt`, `ErrBlobBeschaedigt`, `ErrFremderService`, `ErrUngueltig`, ein anderer, `fmt.Errorf("%w", ErrNichtGefunden)`, `Join(ErrNichtGefunden, ErrBeschaedigt)`}; erwartet je Element aus der Handtabelle (S1/S3 `E7003`, S5 `E7001`, S8 `E7003`, S4/S6/S7 `E9001`, Beschädigung immer `E3001`, `ErrFremderService` immer `E5004`; S6 mit `ErrNichtGefunden`: Erfolg) | Gegenstand an einer Stelle falsch (S8 `Workspace` ergibt `E7001`); `==` statt `errors.Is`; erste Zeile der Tabelle statt Rangfolge | andere |
| Z7 | `*Fehler`: `Code`, `Status` des Codes (`SPEC-109` Nr. 2), `Ursache` über `Unwrap`, `Error()` ist Code und `message` ohne Wert aus Request, Antwort oder Ursache (Nr. 15) | Menge 13 (die erreichbaren Codes), erwarteter Status aus der Tabelle: `E3001` 500, `E5001` 404, `E5002` 410, `E5003` 400, `E5004` 409, `E5005` 404, `E5006` 400, `E5007` 409, `E5008` 422, `E7001` 404, `E7003` 404, `E7009` 409, `E9001` 500; Canary in Pfad, Query, Header, Body und Ursache nie in `Error()` | Status fest 404; `Ursache` nicht verpackt; Canary im Text | andere |
| Z8 | Dienst-Wahl: Segmentgrenze, längster Root, byteweise, kein Dienst ist `E5005` (Nr. 4) | Menge 14: Dienste {`/`, `/a`, `/a/b`, `/ab`} × Pfade {`/`, `/a`, `/a/`, `/a/b`, `/a/b/c`, `/ab`, `/abc`, `/A`, `/a%2Fb`, `/x`, `/a/b/`, `/a/bc`, leer, `x`}; Orakel (Dienst, OData-Pfad): `/`→(`/`,`/`), `/a`→(`/a`,`/`), `/a/`→(`/a`,`/`), `/a/b`→(`/a/b`,`/`), `/a/b/c`→(`/a/b`,`/c`), `/ab`→(`/ab`,`/`), `/abc`→(`/`,`/abc`), `/A`→(`/`,`/A`), `/a%2Fb`→(`/`,`/a%2Fb`), `/x`→(`/`,`/x`), `/a/b/`→(`/a/b`,`/`), `/a/bc`→(`/a`,`/bc`), leer und `x`→`E5005` | Präfix ohne Segmentgrenze (`/abc`→`/a`); kürzester Root; Vergleich ohne Schreibweise (`/A`); Escapes dekodiert | andere |
| Z9 | Der Key trägt `Dienst.ID` und den OData-Pfad (nicht den vollen Pfad) (Nr. 4) | Menge 8: Dienst {4} × Pfad {Root, Root+`/x`}; die Aufzeichnung steht unter dem unabhängig gebildeten Key; zwei Dienste mit demselben OData-Pfad finden nur die eigenen Exchanges | Pfad mit Root im Key; feste Dienst-Kennung | andere |
| Z10 | Steuerheader: je Header genau ein nichtleerer Wert, sonst `E5003` vor jedem Port (Nr. 5) | Menge 21: Header {3} × Form {fehlt, leere Wertliste, ein leerer Wert, zwei verschiedene, zwei gleiche, zwei Werte in zwei Schreibweisen des Namens, ein Wert unter kleingeschriebenem Namen}; 18 Verstöße (Zähler aller Fakes 0), 3 Gutfälle | Eine Form nicht geprüft (je Element rot); zwei gleiche Werte als einer; Namensvergleich mit Schreibweise (Gutfall rot) | andere |
| Z11 | Form der Replay-Session-ID: 1 bis 128 Bytes aus `0-9A-Za-z._-`, sonst `E5003` (Nr. 5) | Menge 771: Länge 1, 128 (gut), 129 (Verstoß) und jedes Byte 0x00 bis 0xFF an erster, mittlerer und letzter Stelle einer Kennung der Länge 3 (768; 195 gut, 573 Verstoß); Fixtures aus Bytewerten | Grenze 127 oder 129; ein Zeichen des Alphabets fehlt (`.`, `_`, `-`); Prüfung nur am ersten Byte; `unicode.IsLetter` (Byte ≥ 0x80 gut) | andere |
| Z12 | Workspace- und Session-Kennung gehen byte-gleich an die Ports (Nr. 5) | Menge 6: Header {Workspace, Recording-Session} × Wert {` x`, `x `, `X`}; der Fake hält das Argument von `Workspaces.Get` bzw. `Sessions.Get` fest | `TrimSpace`; `ToLower` | andere |
| Z13 | Keine implizite Zuordnung: ohne Header gibt es keine Wahl (Nr. 5) | Menge 3: Zustand {genau ein Workspace und eine `complete`-Session, mehrere, keine} ohne die drei Header: `E5003`, `Sessions.List` und `FindOpen` zählen 0 | Bei genau einer Session wird sie gewählt | andere |
| Z14 | Zustände der Recording-Session: unbekannt `E7003`, anderer Service `E5004`, `open` `E7009`, `complete` und `incomplete` laufen weiter (Nr. 6 a) | Menge 5; je Element Cursor und Store unverändert, `Put` 0, `Open` 0, kein Aufruf von `Exchanges` bei einem Fehler | `incomplete` abgewiesen; `open` durchgelassen; Service nicht verglichen | andere |
| Z15 | Zustände des Workspace in `f`: unbekannt `E7001`, anderer Service `E5004`, sonst weiter; kein Cursor gelesen, kein `Put` (Nr. 6 b) | Menge 3; Zähler `Cursors.Get` 0 bei den Fehlern | Workspace nicht gelesen; Service nicht verglichen | andere |
| Z16 | Rangfolge der Verstöße am selben Request (Zusatztabelle, 17 Paare P1 bis P17) | Menge 17: je Paar ein Request mit beiden Verstößen; erwartet der Sieger der Zusatztabelle | Zwei Prüfungen vertauscht (je Paar rot) | andere |
| Z17 | Request-Body: jede Form der Eingabe wird entschieden, Key aus der Ausgabe (Nr. 7 a, c) | Menge 16: leer gelesen (`Laenge` -1, 0 Bytes), JSON mit Geheimnis, JSON ohne Treffer, `Content-Type` fehlt, `text/plain`, zwei `Content-Type`, `application/xml` bei Pfad `/$metadata`, Syntaxfehler, ungültiges UTF-8, Tiefe 513 (Verstoß), Tiefe 512 (gut), Eingabe `MaxBytes`+1 (Verstoß), genau `MaxBytes` (gut), Ausgabe `MaxBytes`+1 durch Ersetzung (Verstoß), Ausgabe genau `MaxBytes` (gut), Zahl unter treffendem Namen; `MaxBytes` klein konfiguriert; erwartet `E5001` ohne Cursor oder der unabhängig gebildete Key | Redaktor nicht gerufen; `Metadata` wahr bei `/$metadata`; Grenze der Ausgabe fehlt; `MaxBytes`+1 als gut; Key aus den Eingabebytes | andere |
| Z18 | `Content-Encoding` des Requests: nur `identity`-Teile, sonst `E5001` ohne Lesen; ohne Body ohne Wirkung (Nr. 7 a, b) | Menge 22: Wert {fehlt, leere Liste, `identity`, `IDENTITY`, `identity, identity`, ` identity `, `gzip`, `deflate`, `br`, leerer Wert, `identity,`} × Body {vorhanden, nil}; gelesene Bytes des Bodys bei Verstoß 0 | `gzip` entpackt oder durchgelassen; leerer Teil als `identity`; Prüfung auch ohne Body (Gutfall mit nil rot) | andere |
| Z19 | Der Key entsteht aus der redigierten Ausgabe und dem Modus: ein anderes Geheimnis trifft, andere Reihenfolge nur bei `json-canonical` (Nr. 7 c, d) | Menge 10: Modus {`bytes`, `json-canonical`} × Paar (Aufnahme, Replay) {gleich, Schlüsselreihenfolge anders, Leerraum anders, Geheimnis anders, anderer Inhalt}; Treffer `bytes`: ja, nein, nein, ja, nein; `json-canonical`: ja, ja, ja, ja, nein | Key aus den Rohbytes (Geheimnis anders rot); Modus fest | andere |
| Z20 | Die Header des Keys sind die des Requests wie empfangen, auch `X-Recorder-*` und Hop-by-Hop; Redigierte gehen nie ein (Nr. 7 d) | Menge 4: als Vary konfiguriert {`X-Recorder-Replay-Session`, `Connection`, `Accept`} (Key hängt vom Wert ab) und `Authorization` (Key hängt nicht ab) | `X-Recorder-*` oder Hop-by-Hop vor dem Key entfernt | andere |
| Z21 | `ErrKodierung` ist `E5006` ohne Cursor, bei `literal` kein Fehler (Nr. 7 d) | Menge 4: Ort {Pfad `%zz`, Query `a=%zz`} × Modus {`literal`, `reject`} | `reject` ignoriert; `E5006` als `E9001` | andere |
| Z22 | `ErrUngueltig` des Keys ist `E9001`, nicht `E5006` (Nr. 7 d) | Menge 5: Pfad mit `?`, `#`, 0x00, 0x1F, 0x7F (Dienst Root `/`) | Fehler als `E5006` | andere |
| Z23 | Der Abbruch des Clients beim Lesen des Bodys ist ein Kontextfehler ohne Code, jeder andere Lesefehler `E5001` (Nr. 7 c) | Menge 15: Fehler des Readers {`Canceled`, `DeadlineExceeded`, `Join(Canceled, ErrUnexpectedEOF)`, `ErrUnexpectedEOF`, ein eigener} × Zeitpunkt {vor dem ersten Byte, nach Daten, mit Daten zusammen}; 9 Kontextfehler, 6 `E5001` | Alles `E5001`; alles Kontextfehler | andere |
| Z24 | Ein leerer `MatchKey` passt nie (Nr. 8) | Menge 8: Strategie {4} × {Session nur mit Exchange ohne Key, Exchange ohne Key vor `A`}; erwartet `E5001` bzw. `A` | Leerer Key gleich leerem Key | andere |
| Z25 | `ListBySession` genau einmal, mit der Kennung unverändert, vor der Transaktion; der Dienst nimmt keine Ordnung der Liste an (Nr. 8) | Menge 12: Ordnung der Liste {aufsteigend, absteigend, gemischt} × Strategie {4}; Orakel wie in Z26 bis Z28; Haltepunkt: kein Store-Aufruf läuft beim Aufruf | Erstes Element statt kleinstes oder größtes; Aufruf in `f`; zwei Aufrufe | andere |
| Z26 | `sequential`: kleinste unverbrauchte replayfähige `SequenceNo`; `E5001` bei leerem *M*, `E5008` bei nur nicht replayfähigen unverbrauchten, `E5002` bei allen verbrauchten (Nr. 9) | Menge 1755: Layouts der Länge 0 bis 3 über {A, a, B, b} (85) × verbrauchte Teilmenge (je Layout 2^n; zusammen 585) × angefragter Key {A, B, C}; Orakel eine unabhängige Handtabelle; Put-Zähler 0 und Open-Zähler 0 bei jedem Fehler | `consumed` ignoriert; größte statt kleinste; `E5002` und `E5008` vertauscht; nicht replayfähiger gewählt; Präfix im Keyvergleich | andere |
| Z27 | `strict`: allein `SequenceNo` = `position`+1; `E5001` vor `E5002`, `E5007` vor `E5008` (Nr. 9) | Menge 939: Layouts (85) × `position` 0 bis n (zusammen 313) × Key {A, B, C}; Orakel Handtabelle | `position` statt `position`+1; `E5007` und `E5008` vertauscht; `E5002` vor `E5001` | andere |
| Z28 | `first` und `last`: kleinste bzw. größte replayfähige `SequenceNo`, `consumed` ohne Wirkung, `E5008` ohne replayfähige (Nr. 9) | Menge 3510: Strategie {2} × Layouts mit verbrauchter Teilmenge (585) × Key {A, B, C}; Orakel Handtabelle; `Put` 0 | `consumed` befragt; `first` und `last` vertauscht; nicht replayfähiger gewählt | andere |
| Z29 | Beispiele von `LH-FA-04`: A(1), B(2), A(3), angefragt B, A, A liefert 2, 1, 3, danach `E5002`; zwei gleiche Requests liefern in aufgezeichneter Reihenfolge; `strict`: A, B, A liefert 1, 2, 3; B zuerst `E5007` mit stehendem Cursor; `first` und `last` liefern dreimal denselben; nur `a(1)`: `E5008` ohne Fortschritt (Nr. 9, 17) | Menge 6 benannte Läufe mit der erwarteten Folge von `SequenceNo` und `position` | Verbrauch nach `SequenceNo` des Aufrufs statt des Exchange | andere |
| Z30 | Die Strategie des bestehenden Cursors gilt; ein neuer Cursor bekommt `Strategie` der Konfig; `first` und `last` legen keinen an (Nr. 10) | Menge 20: Konfig {4} × Cursor {fehlt, `sequential`, `strict`, `first`, `last`}; erwartet wirksame Strategie und `Strategy` im `Put` | Konfig vor Cursor; `Put` mit der Strategie des Requests | andere |
| Z31 | `Put` trägt genau `WorkspaceID`, `ReplaySessionID`, `RecordingSessionID`, `ServiceID`, `Strategy`, `Position` (Anzahl), `Consumed` (alt plus `SequenceNo`); `first` und `last` rufen `Put` nie (Nr. 10) | Menge 8: Strategie {4} × Cursor {neu, bestehend mit `consumed` {7, 9}}; Argument des Fakes feldweise, `Validate()` des Cursors | `ServiceID` leer; Replay- und Recording-Session vertauscht; `Position` gleich `SequenceNo`; `Consumed` ersetzt statt ergänzt | andere |
| Z32 | Identität des Cursors: Workspace, Replay-Session und Recording-Session; eine neue Replay-Session beginnt bei 0 (Nr. 10) | Menge 4: je eine Komponente verschieden (3) und alle gleich (1); `[A(1), A(2)]` | Identität ohne eine Komponente (je eine Mutation rot) | andere |
| Z33 | Bei jedem Fehler bleibt der Cursor unverändert, `Put` 0, kein `Advance` (Nr. 3, 9, 11) | Menge 26: die 13 erreichbaren Codes × Strategie {`sequential`, `strict`}; Store-Zustand vorher und nachher gleich | `Put` vor `Open` (`E3001` verbraucht); Cursor vor der Auswahl angelegt | andere |
| Z34 | Status und Body byte-gleich aus dem Blob (Nr. 12) | Menge 24: Status {200, 201, 204, 304, 404, 500} × Body {ein Byte 0x00, BOM und ASCII, 0xFF 0xFE CRLF, 65 536 Bytes mit allen 256 Werten}; Bytes aus Bytewerten, gegengelesen | Status fest 200; Body gekürzt oder umkodiert | andere |
| Z35 | Body `absent`: leerer Strom, nie nil, `Open` nicht gerufen (Nr. 12) | Menge 9: Methode {GET, HEAD, POST} × Status {200, 204, 304} | `Open` mit leerer Referenz; Body nil | andere |
| Z36 | Header der Auslieferung: redigierte, `Content-Length`, `Content-Encoding`, `VOLATILE_HEADERS` und ein gespeichertes `X-Recorder-Match-Redacted` entfallen, alles andere bleibt (Nr. 13) | Menge 135: Name {7 flüchtige, `Content-Length`, `Content-Encoding`, `X-Recorder-Match-Redacted`, `Location`, `ETag`, `Content-Type`, `Set-Cookie` mit Marker, `X-Custom`} (15) × Schreibweise {kanonisch, klein, groß} × Konfig {Default, `none`, `X-Custom`}; erwartet aus der Handtabelle (`none` liefert `Date` aus) | Eine Klasse nicht entfernt; Vergleich mit Schreibweise; `none` ignoriert | andere |
| Z37 | `Location` mit Marker im Query-Wert geht aus; wertbasiert entfällt nur ein Header, dessen Werte alle der Marker sind (Nr. 13 a) | Menge 12: Name {`Set-Cookie`, `Location`} × Wertliste {leer, `[""]`, `[Marker]`, zwei Marker, `[Marker, "x"]`, `["x[REDACTED]"]`} plus `Location` `/p?token=[REDACTED]` unverändert | Alle Header mit Marker entfernt; leere Liste entfernt | andere |
| Z38 | Die Header der Antwort sind eine Kopie: Änderung der Map oder einer Wertliste ändert den Exchange nicht (Nr. 13) | Menge 2: Änderung an Map, Änderung an Wertliste; zwei Aufrufe liefern unabhängige Maps | Die Map des Exchange zurückgegeben | andere |
| Z39 | `X-Recorder-Match-Redacted: true` genau dann, wenn der gewählte Exchange `KeyRedacted` trägt (Nr. 13 e) | Menge 16: `KeyRedacted` {falsch, wahr} × Strategie {4} × Body {present, absent}; genau ein Wert `true`; Feld des Requests wahr bei falschem Feld des Exchange: kein Header | Feld des Requests; Wert `True`; Header immer | andere |
| Z40 | Genau ein Eintrag `WARN` mit festem Text und 7 Attributen, nach der Bestätigung, nie bei Fehler, ohne Wert des Requests (Nr. 16) | Menge 9: Fälle {wahr und erfolgreich (1 Eintrag), wahr mit Auswahlfehler, wahr mit `Put`-Fehler, falsch} (4) und Canary an {Pfad, Query, Request-Header, Response-Header, Body} (5); `FakeTelemetry` hält Level, Text, jedes Attribut | Eintrag vor der Bestätigung; Query im Eintrag; ein Attribut mehr | andere |
| Z41 | Replayfähigkeit und `REPLAY_REDACTED_BODIES` im Dienst: nur `complete`, `verified` oder `absent`, `none` (oder `applied` bei wahr); sonst `E5008` als einziger Kandidat (Nr. 8, 9) | Menge 72: `capture_status` {3} × Integrität {`verified`, `absent`, `invalid`} × Response-Redaction {`none`, `applied`, `unsafe`, leer} × `RedigierteBodies` {2}; 6 geliefert | Capture nicht befragt; `applied` immer oder nie erlaubt | andere |
| Z42 | Ein beschädigter oder fehlender Blob ist `E3001` und verbraucht nichts (Nr. 11, 12) | Menge 4: Fehler {`ErrBlobFehlt`, `ErrBlobBeschaedigt`} × Strategie {`sequential`, `strict`}; Cursor des Stores unverändert, `Put` 0 | `Put` vor `Open` | andere |
| Z43 | Der Strom wird genau einmal geschlossen: bei `Put`-Fehler und bei Fehler der Bestätigung vom Dienst, bei Erfolg vom Aufrufer über `Close` des `Body` (Nr. 12) | Menge 3: Pfad {`Put`-Fehler, Fehler beim Bestätigen, Erfolg}; Zähler der Hülle des Stroms | Kein `Close` bei Fehler der Bestätigung; doppeltes `Close`; `Close` bei Erfolg durch den Dienst | andere |
| Z44 | Die Auswahl schließt die Auslieferung nicht ein: der Cursor steht nach der Rückkehr fest, auch bei Abbruch (Nr. 14) | Menge 3: Body nie gelesen und geschlossen, halb gelesen und geschlossen, `ctx` nach der Rückkehr beendet; `position` je 1 | Fortschreiben erst bei `Close` oder Ende des Bodys; Rollback bei Abbruch | andere |
| Z45 | Kein Upstream: kein Konfig-Feld, kein Import, kein Symbol eines Upstream-Ports (Kopf, Nr. 12) | Menge 15: die 11 Felder von `Konfig` nach Reflexion (keines vom Typ `UpstreamClient`) und 4 Regeln über den Quelltext des Pakets (`go/parser`, auch `_test.go`): kein Import unter `adapters/`, `net/http`, `services/`, kein Symbol `UpstreamClient` | Feld `Upstream outbound.UpstreamClient`; Import von `adapters/driven/httpupstream` | andere |
| Z46 | Zwei bis vier gleichzeitige `Replay` desselben Cursors (`sequential`, ebenso viele replayfähige Exchanges gleichen Keys): jeder erhält einen anderen, `consumed` enthält alle, `position` gleich n, kein Fortschreiben geht verloren (Nr. 17 a) | Goroutinen, Start über Kanal, FIFO des Stores, Haltepunkt nach `Cursors.Get` des ersten Aufrufs reiht den zweiten ein; Menge 3 (n = 2, 3, 4); `-count=50` | `Cursors.Get` und `Cursors.Put` in zwei Transaktionen (zwei Aufrufe liefern denselben Exchange) | nebenläufig |
| Z47 | Zwei Replay-Sessions desselben Workspace und derselben Recording-Session `[A(1), A(3)]` gleichzeitig: jede erhält 1, dann 3, die Cursors sind unabhängig (`LH-FA-04`; Nr. 10, 17 b) | Alle 6 Verschränkungen von je zwei Requests (C(4,2)) durch Kanal und FIFO; Folge und Cursor je Session | Identität des Cursors ohne die Replay-Session (die zweite erhält 3) | nebenläufig |
| Z48 | `Replay` gegen den Reset: das Fortschreiben liegt vor dem Reset (danach `position` 0, `consumed` leer) oder nach ihm (danach 1), nie `position` 2 aus altem Wert plus 1 (Nr. 17 c, `SPEC-129` Nr. 11 a) | Eindringling `Cursors.ResetAll` in einer Transaktion des Stores; Vorbedingung `position` 1; zwei Reihenfolgen durch Kanal und FIFO (Muster Z50 des Gebers); Menge 2 | Cursor vor der Transaktion gelesen und in ihr geschrieben (Reset dazwischen ergibt 2) | nebenläufig |
| Z49 | Fenster vor der Transaktion: ein Reset zwischen `Sessions.Get` bzw. `ListBySession` und dem Beginn der Transaktion wirkt vollständig vor ihr; der Dienst liest den Cursor erst in der Transaktion (Nr. 11) | Haltepunkte am `Sessions`-Fake und am `Exchanges`-Fake (Argumente des Dienstes) reihen den Reset in die FIFO ein; erwartet `position` 1 nach Reset; Menge 2 | Der Cursor in einer ersten Transaktion gelesen, in einer zweiten geschrieben (Reset dazwischen, `position` 2) | nebenläufig |
| Z50 | Fenster 1 der Transaktion, zwischen `Workspaces.Get` und `Cursors.Get`: kein Schreiber dazwischen (Nr. 11) | Haltepunkt des Stores nach dem ersten Schritt reiht einen Eindringling (Reset; Löschen des Workspace) in die FIFO ein, er läuft nach der Transaktion; Menge 2; `-count=50` | `Workspaces.Get` in eigener Transaktion (Reset dazwischen ergibt `position` 1 statt 0; Löschen dazwischen `E7003`) | nebenläufig |
| Z51 | Fenster 2, zwischen `Cursors.Get` und `Open` bzw. `Put` (Nr. 11) | Wie Z50 nach dem zweiten Schritt, Reset; Body {present, absent}; Menge 2 | Auswahl und `Put` in einer zweiten Transaktion (`position` 1 statt 0) | nebenläufig |
| Z52 | Fenster 3, zwischen `Blobs.Open` und `Cursors.Put` (Nr. 11) | Wie Z50 nach dem dritten Schritt (der Haltepunkt am Fake von `Open`), Reset; Menge 1 | `Put` in einer zweiten Transaktion (`position` 1 statt 0) | nebenläufig |
| Z53 | `Replay` gegen Löschen: Workspace gelöscht ist Erfolg (dann Workspace und Cursor weg) oder `E7001`, nie `E7003` und nie ein Cursor ohne Workspace; Recording-Session gelöscht ist `E7003` ohne Cursor (Nr. 17 d, e) | Menge 4: Eindringling {Löschen des Workspace, Löschen der Session} × Reihenfolge {vor, nach}; dazu der Haltepunkt nach `Workspaces.Get` | `Workspaces.Get` außerhalb von `f` oder `Put`-Fehler als `E7003` bei gelöschtem Workspace | nebenläufig |
| Z54 | Der Dienst hält keinen Zustand und keine Sperre; die Fänger Z46 bis Z53 laufen ohne Datenrennen (Nr. 17) | `TestPaketOhneZustand` liest den Quelltext (`go/parser`, ohne `_test.go`) und meldet Paketvariable, `init`, go-Anweisung, `AfterFunc`, Import von `sync` und `sync/atomic`, Typ oder Methode außer `Replayer` und `Konfig`, Schreibzugriff auf die Konfig, Kopie des Empfängers; Menge 10 (5 Regeln × Verstoß und Gutfall); Z46 bis Z53 zusätzlich unter `make test-race`, je Prozess `-test.count=1` | Paketvariable als Zähler; `sync.Mutex` im Dienst; Schreibzugriff `v.k.Strategie = …` | nebenläufig |
| Z55 | Die Fakes zählen jede Operation, halten jedes Argument, speisen an jeder Operation einen Fehler ein, rollen zurück, reichen eine Panik weiter | Reflexion über die 39 Operationen und je Eigenschaft ein Test (FIFO, Rollback, Panik) | Eine Operation zählt nicht; Rollback fehlt | andere |
| Z56 | Die Hülle um Strom und `ctx` trägt jede Schnittstelle des Originals und zählt je Aufruf | Methodenmenge der Hülle enthält die des Originals (Reflexion), Zähler | Hülle ohne `Close` oder `WriterTo` | andere |
| Z57 | Der Dienst ruft nur `Sessions.Get`, `Exchanges.ListBySession`, `WithWorkspaceTransaction`, `Workspaces.Get`, `Cursors.Get`, `Cursors.Put`, `Blobs.Open`, `Telemetrie.Log` (Nr. 11, 18) | Menge 155: die 31 verbotenen Operationen × 5 Läufe (Erfolg je Strategie, ein Fehlerlauf); Zähler 0 | `Cursors.ResetAll`, `Workspaces.Advance` oder `Blobs.Verify` gerufen | andere |
| Z58 | Der Dienst importiert keinen Adapter und keinen anderen Dienst | `make a-check`, benannte Ausgabe | Import von `adapters/driven/sqlite` im Paket | andere |
| Z59 | Akzeptiertes Negativ: ein komprimierter Request-Body ist im Replay nicht abspielbar (Nr. 7 b) | – Grund: Entpacken liegt im Dienst `record`, ein Dienst importiert keinen anderen, eine zweite Entpackung wäre ein zweiter Ort. Folge: Aufnahmen mit `gzip`- oder `deflate`-Request-Body (selten) ergeben `ODR-E5001`; Z18 belegt die Ablehnung | – | andere |
| Z60 | Akzeptiertes Negativ: Aufwand je Request wächst mit der Zahl der Exchanges der Session (Nr. 8) | – Grund: der Port bietet nur `ListBySession`, ein Port `ListByMatchKey` änderte `SPEC-105` und den Adapter, eine Grenze ohne Messung wäre geraten. Folge: bei 10 000 Exchanges wird jeder Request mit dem Satz gelesen; `LH-QA-01` misst `slice-lasttest`, bei Fehlschlag ist der Port die Antwort | – | andere |
| Z61 | Akzeptiertes Negativ: `X-Recorder-Match-Redacted` und Eintrag folgen dem Feld des Exchange, nicht dem des Requests (Nr. 13 e) | – Grund: die Spezifikation bindet beides an den Exchange (`LH-FA-14.a`). Folge: ein Exchange, der den Marker als Literal sandte, und ein Replay mit dem echten Geheimnis fallen ohne Hinweis zusammen | – | andere |
| Z62 | Akzeptiertes Negativ: `REPLAY_ASSIGNMENT=ordered` hat im Dienst keine Wirkung (Nr. 5) | – Grund: der Dienst kennt nur die ausdrückliche Zuordnung, und die Ablehnung des Wertes gehört dem Composition Root (Sendung an `slice-bootstrap-m2`, §6). Folge: ohne diese Entscheidung wirkt `ordered` wie `explicit` | – | andere |
| Z63 | Akzeptiertes Negativ: die Serialisierung der Transaktionen allein durch den Adapter und ein Datenrennen, das kein Test auslöst (Nr. 17, *Ungeprüft*) | – Grund: der Fake ordnet nach seiner Annahme (`SPEC-108` Nr. 1), der M2-Test belegt nur den Gutfall auf der Datei. Folge: bricht der Adapter die Serialisierung, versagen Z46 bis Z53 in Wirklichkeit unbemerkt; der Vertrag der Transaktion fängt es im Adapter | – | andere |
| Z64 | Akzeptiertes Negativ: eine Panik eines Ports und ein Lesefehler des Blob-Stroms nach der Rückkehr (Nr. 12, 14, *Ungeprüft*) | – Grund: die Ports panicen nie, und der Dienst liest den Strom nicht. Folge: ein geöffneter Strom bleibt bei einer Panik offen; ein Lesefehler mitten in der Übertragung lässt den Cursor fortgeschritten | – | andere |

**Fänger je Zusage** (Namen für die Prüfung; der Implementer legt sie an): Z1 `TestNewKonfig`, Z8 `TestDienstWahl`, Z11 `TestReplaySessionForm`, Z17 `TestRequestBodyFormen`, Z26 bis Z28 `TestAuswahlSequential`, `TestAuswahlStrict`, `TestAuswahlFirstLast`, Z46 `TestGleichzeitigGleicherCursor`, Z47 `TestZweiReplaySessions`, Z48 `TestReplayGegenReset`, Z49 bis Z52 `TestFensterDerTransaktion` (Eindringling in der FIFO nach Schritt k), Z53 `TestReplayGegenLoeschen`, Z54 `TestPaketOhneZustand`.

#### Zusatztabelle des Rasters (`AGENTS.md` §3.16): Paare von Verstößen am selben Request und Modus

Fänger für alle Paare ist Z16 (Menge 17, Element Pk); die Modi sind Zeilen für sich.

| Paar | Verstöße am selben Request | Sieger | Grund (Nr.) | Fänger |
|---|---|---|---|---|
| P1 | kein Dienst · Steuerheader fehlt | `E5005` | 3 (b) vor (c) | Z16 P1 |
| P2 | Steuerheader fehlt · Session unbekannt | `E5003` | 3 (c) vor (d) | Z16 P2 |
| P3 | Session unbekannt · Workspace unbekannt | `E7003` | 3 (d) vor (g) | Z16 P3 |
| P4 | Session anderer Service · Session `open` (dieselbe Session) | `E5004` | 6 a Folge | Z16 P4 |
| P5 | Session `open` · Body unzulässig | `E7009` | 3 (d) vor (e) | Z16 P5 |
| P6 | Session unbekannt · Body unzulässig | `E7003` | 3 (d) vor (e) | Z16 P6 |
| P7 | Body unzulässig · Pfad ungültig kodiert bei `reject` | `E5001` | 7 d | Z16 P7 |
| P8 | Pfad ungültig kodiert · Workspace unbekannt | `E5006` | 3 (e) vor (g) | Z16 P8 |
| P9 | Body unzulässig · Workspace unbekannt | `E5001` | 6 b Entscheidung | Z16 P9 |
| P10 | Workspace unbekannt · Key nirgends aufgezeichnet | `E7001` | 3 (g) | Z16 P10 |
| P11 | Workspace anderer Service · Session erschöpft | `E5004` | 6 b | Z16 P11 |
| P12 | `strict`: Key nirgends · `position`+1 hat anderen Key | `E5001` | 9 | Z16 P12 |
| P13 | `strict`: Key nirgends · Session erschöpft | `E5001` | 9 | Z16 P13 |
| P14 | `strict`: anderer Key an `position`+1 · derselbe Exchange nicht replayfähig | `E5007` | 9 | Z16 P14 |
| P15 | `sequential`: alle replayfähigen verbraucht · unverbrauchter nicht replayfähiger | `E5008` | 9 | Z16 P15 |
| P16 | Blob beschädigt · `Put` würde `ErrFremderService` liefern | `E3001` | 11, 12 | Z16 P16 |
| P17 | Replay-Session-ID ungültig · Workspace und Session unbekannt | `E5003` | 5 | Z16 P17 |

| Modus | Aussage, die vom Modus abhängt | Fänger |
|---|---|---|
| `sequential` | wählt den kleinsten unverbrauchten replayfähigen; schreibt; Fehler `E5001`, `E5002`, `E5008` | Z26, Z29, Z31 |
| `strict` | wählt `position`+1; schreibt; Fehler `E5001`, `E5002`, `E5007`, `E5008` | Z27, Z29, Z31 |
| `first` | kleinster replayfähiger; schreibt nie; Fehler `E5001`, `E5008` | Z28, Z29, Z31 |
| `last` | größter replayfähiger; schreibt nie; Fehler `E5001`, `E5008` | Z28, Z29, Z31 |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-workspace-use-case` und `slice-streaming-redaction` und `slice-redaction-match-key` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Lauf springt von `open/` direkt nach `in-progress/` (wie `slice-workspace-use-case`); die drei Vorgänger liegen am 2026-10-31 in `done/`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke dieses Abschnitts, die der Planner nach dem Architect setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen); wird sie überschritten, trägt der Slice einen Eintrag mit Grund im Block (`SPEC-128` Nr. 5), eine Rückführung ist daran nicht geknüpft (`SPEC-128` Nr. 10).

**Größenmarke.** Schätzung *S* = 85·9 + 80·55 + 200·4 + 800 = 6765 (Zeilen der Tabelle: 9 nebenläufig, 55 andere; vier Port-Fakes; Produktcode 800 nach dem Architect, Spanne 650 bis 950), Marke 100·⌈135·6765/10000⌉ = 9200 (`SPEC-128` Nr. 1). Gezählt wie im Plan: Z46 bis Z54 nebenläufig, Z1 bis Z45 und Z55 bis Z64 andere. Der Planner setzt die Marke vor dem ersten Code-Commit; danach ist sie fest (`SPEC-128` Nr. 5 c). `basis` ist der Commit mit `SPEC-135` und dem Plan (`b919969`), der letzte Commit vor dem Code.

```groessenmarke
pfad: src/hexagon/ports/inbound/replay.go
pfad: src/hexagon/services/replay
produktcode: 800
fakes: 4
basis: b919969b03229c6311556c7fc9a18281c8da4305
schaetzung: 6765
marke: 9200
schnitt-grund: Nicht vor dem Code geschnitten, vorab benannter Handoff-Schnitt. Der Slice liefert eine Operation; Zuordnung, Key, Auswahl, Cursor und Auslieferung bilden eine Kette, die Konfig, Fehlerabbildung (SPEC-135 Nr. 15) und dieselben vier Fakes teilt; kein Teil allein ist ein Lieferwert, und ein Schnitt vor dem Code trennte nur rund 450 Zeilen ab. Abtrennbar ist allein der Teil Request-Body (SPEC-135 Nr. 7 a bis c; Zeilen Z17, Z18, Z19, Z23): er ist eine Funktion ohne Port und ohne Fake und teilt mit dem Rest nur Konfig und Fehlerabbildung. Er wird zuletzt in eigenem Commit geschrieben und davor gemessen. Liegt die Messung beim Handoff über der Marke, geht er als eigener Slice an den Planner; bis dahin lehnt der Rest jeden Request mit Body mit ODR-E5001 ab.
```

Messpunkte und Abbruchkriterien (Vorschlag für den Planner): (1) Zwischenmessung nach Zuordnung, Key, Auswahl und Fakes, vor der Auslieferung und den nebenläufigen Fängern; liegt sie über 5000, meldet der Implementer es dem Planner mit Zeilen je Datei (Erwartung 3000 bis 5000). (2) Vor dem Commit des Teils Request-Body: liegt der Rest über 8500, wird der Teil nicht mehr geschrieben. (3) Gesamtmessung über der Marke: Eintrag `ueberschreitung` mit Grund. (4) Fixes und Review über 11 000 geben an den Planner zurück. Erwartung aus den Faktoren von `SPEC-128`: Handoff 4900 bis 7000, Ende 5800 bis 9000.

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

- **`Location` mit Marker im Query-Wert.** Die Sätze in `SPEC-045` und `LH-FA-14.a` stehen und zeigen auf `SPEC-115` Nr. 3, die Historie trägt die Zeile (Vertrag Nr. 13). Entscheidung: *Vorschlag, Owner bestätigt* (Plan §1: Entscheidung vom 2026-10-21); sie blockiert den Code nicht. **Ausgang:** offen bis zur Closure; Adresse Owner für die Bestätigung. Entfällt die Bestätigung, ändert nur Nr. 13 und Z36/Z37.
- **Header der Auslieferung.** Entschieden in `SPEC-135` Nr. 13 (redigierte, `Content-Length`, `Content-Encoding` und `VOLATILE_HEADERS` entfallen; `X-Recorder-Match-Redacted` aus dem Feld des Exchange). Fänger Z36 bis Z39. **Ausgang:** offen bis zur Closure (der Verifier prüft die Policy gegen `LH-FA-04.a` Nr. 3).
- **Cursor fortschreiben, Vertrag und Fenster.** Entschieden in `SPEC-135` Nr. 9 bis 11 und 17: ein `WithWorkspaceTransaction` für Lesen, Auswahl, Öffnen des Blobs und Schreiben; Fenster 0 bis 3 mit Wirkungs-Fängern Z48 bis Z52, Gegenstand Z53. Folge im Betrieb benannt: der Schreiber ist auch für die Hash-Prüfung gehalten. **Ausgang:** offen bis zur Closure; das Gate fährt `-race` (`make test-race`), die nebenläufigen Fänger ordnen mit Kanälen (`SPEC-122` Nr. 8 b), jeder nebenläufige Test läuft vor der Übergabe mit `-count=50`; Sollwert n von n.
- **Request-Body im Replay.** Entschieden in `SPEC-135` Nr. 7: Größe `MaxBytes` für Eingabe und Ausgabe, `Content-Encoding` außer `identity` ist `ODR-E5001`, jeder Fehler des Redaktors `ODR-E5001`, ein Kontextfehler ohne Code. Fänger Z17 bis Z19, Z23; Negativ Z59. **Ausgang:** offen bis zur Closure.
- **Antwort ohne Body, `Content-Encoding`, Redaction mit `REPLAY_REDACTED_BODIES`.** Entschieden in Nr. 12, 13 (Raster *Antwortart × Auslieferung* unten); Fänger Z34 bis Z41. **Ausgang:** offen bis zur Closure.
- **Abbruch des Clients mitten in der Auslieferung.** Entschieden in Nr. 14: die Auswahl schließt die Auslieferung nicht ein; der Exchange ist ab der Rückkehr verbraucht. Fänger Z44; Negativ Z64 (Lesefehler nach der Rückkehr). **Ausgang:** offen bis zur Closure.
- **Randformen.** Raster *Operation × Argumentart* (unten): 16 Zeilen, 112 Zellen (7 Spalten), jede mit Ort in `SPEC-135` oder „gilt nicht“ mit Grund. **Ausgang:** offen bis zur Closure; eine Zelle ohne Ort geht an den Architect.
- **Aufwand je Request** (Z60): `ListBySession` liest die ganze Session. **Ausgang:** weiter offen; Adresse `slice-lasttest` (Sendung, §3.13; dessen §1 *Ausdrücklich NICHT* nennt Tuning nur als eigenen Slice, die Messung trifft nicht).
- **`REPLAY_ASSIGNMENT=ordered`** (Z62): Entscheidung des Users nötig (*Offene Entscheidungen* unten), Adresse `slice-bootstrap-m2`. **Ausgang:** offen.
- **Sendungen an andere Slices** (Abschnitt *Sendungen* unten): jede steht mit Ausgang *offen* und Adresse; ein Bericht allein ist keine Adresse (`AGENTS.md` §3.13).

#### Raster *Operation × Argumentart* (`AGENTS.md` §3.16)

Legende: „Nr. k“ ist `SPEC-135` Nr. k; „Z“ nennt den Fänger.

### F1. *Operation × Argumentart* (Operation ist `Replay`)

| Argument | fehlend | leer | mehrdeutig | unbekannt | fremd oder ungültig | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|---|
| `ctx` | Nr. 2 (Z3) | gilt nicht: ein Kontext hat keinen leeren Wert | gilt nicht: ein Wert | gilt nicht: keine Aufzählung | gilt nicht: kein Fremdwert | Nr. 2, 14 (Z5) | gilt nicht: kein Zustand |
| `Methode` | Nr. 2 (Z3) | Nr. 2 (Z3) | gilt nicht: ein Wert | Nr. 7 d, 9: geht in den Key, trifft nichts, `E5001` (Z26) | gilt nicht: die Schreibweise zählt, `get` ist unbekannt (Nr. 7 d) | gilt nicht: kein Strom | Nr. 9, 17 |
| `Pfad` | Nr. 4 (Z8) | Nr. 4 (Z8) | gilt nicht: ein Wert | Nr. 4: `E5005` (Z8) | Nr. 4 (anderer Root wählt den anderen Dienst); Nr. 7 d (`?`, `#`, Steuerzeichen: `E9001`, Z22) | gilt nicht: kein Strom | Nr. 9, 17 |
| `RawQuery` | Nr. 7 d (kein Paar) | Nr. 7 d | gilt nicht: eine Zeichenkette | Nr. 7 d: unbekannter Parameter geht in den Key (`SPEC-116` Nr. 4) | Nr. 7 d: ungültig kodiert, `E5006` oder `literal` (Z21) | gilt nicht: kein Strom | Nr. 9, 17 |
| Header `X-Recorder-Workspace` | Nr. 5 (Z10) | Nr. 5 (Z10) | Nr. 5 (Z10) | Nr. 6 b: `E7001` (Z15) | Nr. 6 b: `E5004` (Z15) | gilt nicht: ein Wert, kein Strom | Nr. 9, 10, 17 |
| Header `X-Recorder-Recording-Session` | Nr. 5 (Z10) | Nr. 5 (Z10) | Nr. 5 (Z10) | Nr. 6 a: `E7003` (Z14) | Nr. 6 a: `E5004`, `E7009` bei `open` (Z14) | gilt nicht: ein Wert, kein Strom | Nr. 9, 10, 17 |
| Header `X-Recorder-Replay-Session` | Nr. 5 (Z10) | Nr. 5 (Z10) | Nr. 5 (Z10) | gilt nicht: jede wohlgeformte Kennung ist eine neue Replay-Session bei Position 0 (Nr. 10, Z32) | Nr. 5: ungültige Form `E5003` (Z11) | gilt nicht: ein Wert, kein Strom | Nr. 10 (der Cursor besteht), Nr. 9 (Z29) |
| übrige Request-Header | Nr. 7 d: steht nicht im Tupel (Z20) | Nr. 7 d: leerer Wert ist vom fehlenden verschieden | Nr. 7 d: mehrere Werte sind Teil des Keys | Nr. 7 d: Header außerhalb der Vary-Menge gehen nicht ein | gilt nicht: kein Fremdwert | gilt nicht: kein Strom | gilt nicht: Eingabe, kein Zustand |
| `Body` und `Laenge` | Nr. 7 a (`nil`, 0; Z3, Z17) | Nr. 7 c: 0 Bytes gelesen (Z17) | *Ungeprüft*: `Laenge` gegen gelesene Bytes (nur `nil` und 0 steuern) | Nr. 7 c: Typ unbekannt (Z17) | Nr. 7 c: zu groß, ungültig, nicht redigierbar (Z17) | Nr. 7 c: Kontextfehler ohne Code (Z23) | gilt nicht: der Dienst liest den Body einmal je Aufruf |
| `Content-Type` des Requests | Nr. 7 c: `typ` (Z17) | Nr. 7 c: `typ` (Z17) | Nr. 7 c: mehrere Werte, `typ` (Z17) | Nr. 7 c: unbekannter Typ (Z17) | gilt nicht: Parameter werden nicht gewertet (`SPEC-118` Nr. 3) | gilt nicht: kein Strom | gilt nicht: Eingabe |
| `Content-Encoding` des Requests | Nr. 7 b: leere Kette (Z18) | Nr. 7 b: leerer Teil, `E5001` (Z18) | Nr. 7 b: Kette aus mehreren Werten (Z18) | Nr. 7 b: `gzip`, `br`, …, `E5001` (Z18) | gilt nicht: kein Fremdwert | gilt nicht: kein Strom | gilt nicht: Eingabe |
| Workspace (Zustand) | siehe unbekannt | gilt nicht: der Dienst liest keinen Entity-State | gilt nicht: Schlüssel eindeutig (`SPEC-108`) | Nr. 6 b, 17 d (Z15, Z53) | Nr. 6 b: `E5004` (Z15) | Nr. 11, 14 (Z5, Z44) | Nr. 17 c: Reset (Z48) |
| Recording-Session (Zustand) | siehe unbekannt | Nr. 9: keine Exchanges, `E5001` (Z26) | gilt nicht: Schlüssel eindeutig | Nr. 6 a, 8 (Z14, Z6) | Nr. 6 a: anderer Service, `open` (Z14) | Nr. 2 (Z5) | Nr. 9 (Z29) |
| Cursor (Zustand) | Nr. 10: Position 0 (Z30) | Nr. 9: `consumed` leer (Z26) | gilt nicht: Schlüssel eindeutig (`SPEC-108` Nr. 11) | gilt nicht: die Strategie ist gültig (`SPEC-108` Nr. 10) | gilt nicht: die Identität liegt im Schlüssel (Nr. 10, Z32) | Nr. 11, 14 (Z44) | Nr. 9, 17 a (Z29, Z46) |
| Exchanges der Session | Nr. 9: keine, `E5001` (Z26) | Nr. 8: leerer `MatchKey` (Z24) | Nr. 9: gleicher Key mehrfach (Z26, Z29) | Nr. 9: Key nirgends, `E5001` (Z26) | gilt nicht: nur die Session aus dem Header wird gelesen (Nr. 8, Z25) | Nr. 2 (Z5) | Nr. 9: verbraucht, `E5002` (Z29) |
| Blob | Nr. 12: Datei fehlt, `E3001` (Z42) | Nr. 12: Body `absent` (Z35) | gilt nicht: content-adressiert | gilt nicht: die Referenz stammt aus der Aufzeichnung | Nr. 12: Hash passt nicht, `E3001` (Z42) | Nr. 2, 12 (Z5, Z43) | gilt nicht: unveränderlich |

Herunterfahren: gilt nicht für alle Zeilen, mit Grund: der Dienst hält keine Aufnahme und kennt kein `Drain`; ein laufender Aufruf endet mit seinem `ctx` (Nr. 14).

#### Raster *Antwortart × Auslieferung*

| Antwortart | Status | Header | Body | Cursor |
|---|---|---|---|---|
| 200 mit Body | Nr. 12 (Z34) | Nr. 13 (Z36) | Blob byte-gleich (Z34) | Nr. 10 (Z31) |
| 200 mit leerem Body (`absent`) | Nr. 12 | Nr. 13 | leerer Strom, `Open` 0 (Z35) | Nr. 10 |
| 204, 304 | unverändert (Z34, Z35) | Nr. 13 | leer (Z35) | Nr. 10 |
| Antwort auf `HEAD` | unverändert (Z35) | Nr. 13, `Content-Length` entfällt (Z36) | leer (Z35) | Nr. 10 |
| aufgezeichnetes `Content-Encoding` | unverändert | entfällt mit `Content-Length` (Nr. 13 b, Z36) | entpackter Blob (Nr. 12) | Nr. 10 |
| Response-Redaction `applied` | unverändert | Nr. 13 | mit Markern (Nr. 12), nur bei `RedigierteBodies` (Z41) | Nr. 10; sonst `E5008`, Cursor steht (Z41) |
| Response-Redaction `unsafe` oder unbekannt | `E5008` (Z41) | – | – | unverändert (Z33) |
| Header redigiert (`Set-Cookie`) | unverändert | entfällt (Nr. 13 a, Z37) | unverändert | Nr. 10 |
| `Location` mit Marker im Query-Wert | unverändert | geht aus (Nr. 13, Z37) | unverändert | Nr. 10 |
| flüchtige Header (`Date`, `Server`, …) | unverändert | entfallen, bei `none` ausgeliefert (Z36) | unverändert | Nr. 10 |
| Exchange mit `key_redacted` | unverändert | zusätzlich `X-Recorder-Match-Redacted: true` (Z39) | unverändert | Nr. 10; Eintrag (Z40) |
| aufgezeichnete Fehlerantwort (500) | unverändert (Z34) | Nr. 13 | Nr. 12 | Nr. 10 (ein Replay ist kein Fehler des Dienstes) |

#### Sendungen an andere Slices (`AGENTS.md` §3.13)

Gelesen: §1 *Ausdrücklich NICHT* der Nehmer; kein Ausschluss trifft eine Sendung. Der Planner trägt jede Sendung im Plan des Nehmers ein (die Pläne in `open/` bleiben nach der Entscheidung des Users bis zum Start unverändert, `welle-record-replay.md` §4; die Sendungen stehen deshalb in §6 mit Ausgang *offen*) und prüft sie bei der Closure nach.

1. **`slice-http-adapter`:** füllt die `Anfrage` unverändert (auch `X-Recorder-*`); ruft `Close` des `Body` genau einmal; setzt Länge und Coding der Antwort selbst (der Dienst liefert weder `Content-Length` noch `Content-Encoding`; bei `HEAD` keine Länge); behandelt einen Fehler ohne `*Fehler` und ohne Code (Kontextfehler, Abbruch des Clients, `model.ErrUngueltig`) wie in `SPEC-121` Nr. 14; kennt `X-Recorder-Match-Redacted` nicht.
2. **`slice-bootstrap-m2`:** verdrahtet `replay.Konfig` mit allen elf Feldern (`Sessions`, `Exchanges` und `Transaktion` über dieselbe Datenbank, `Blobs` über denselben Blob-Speicher, `Regeln`, `Pfade`, `MaxBytes`, `Strategie`, `RedigierteBodies`, `Telemetrie`); **Entscheidung des Users nötig** zu `REPLAY_ASSIGNMENT=ordered` (*Offene Entscheidungen* unten).
3. **`slice-telemetry-einbau-gateway`:** der Match-Span `odata.replay.match` (`SPEC-028`) kommt im Dienst dazu; die Konfig führt `Telemetrie` schon (`SPEC-135` Nr. 1), der Protokolleintrag läuft schon über den Port, die Fehlerzählung (`CountError`) gehört dem Ort der Auslieferung (Nr. 15).
4. **`slice-lasttest`:** misst den Aufwand je Request bei 10 000 Exchanges (Z60); bei Fehlschlag ist der Port `ListByMatchKey` der Folge-Slice.
5. **`slice-handbuch-record-replay`:** was ein Anwender sieht (Handbuch-Pflicht dieses Slice: **nein**, weil kein Binary, kein neuer Code, kein Schlüssel, kein Kommando; `slice-handbuch-meldungskatalog` bleibt grün): Form der Replay-Session-ID, welche Header im Replay fehlen (redigierte, `Content-Length`, `Content-Encoding`, flüchtige), `X-Recorder-Match-Redacted`, `Location` mit Marker, `ODR-E5001` auch für komprimierten Request-Body und für nicht redigierbaren Body, ein Client-Abbruch verbraucht den Exchange, Beschreibung der vier Strategien und des Cursors.
6. **`slice-hybrid-service`:** keine neue Pflicht; der Port `ReplayUseCase.Replay` ist der gerufene, und nur `ODR-E5001` darf den Fallback auslösen (Nr. 9).

#### Offene Entscheidungen (nicht blockierend, kein Code hängt daran)

1. **`Location` mit Marker** (Owner): Vorschlag ausliefern mit Marker, wie `SPEC-115` Nr. 3 und Plan §1 (Entscheidung vom 2026-10-21); bei anderem Ausgang ändern sich allein Nr. 13 und Z36/Z37.
2. **`REPLAY_ASSIGNMENT=ordered`** (User): `ordered` ist ein gültiger Wert der Konfiguration (`SPEC-078`, `LH-FA-04.b` „zulässig“, nicht verlangt) und hat im Dienst keine Wirkung. Vorschlag: der Start lehnt den Wert mit `ODR-E2001` („nicht unterstützt“) ab, damit er nicht still wirkungslos bleibt (`LH-RB-01`); das ändert `SPEC-110` Nr. 2, die Prüfung in `src/bootstrap/konfig_werte.go` und deren Test (`konfig_test.go` Zeile 231 nennt `ordered` als gültig), also Dateien außerhalb dieses Slice. Adresse: `slice-bootstrap-m2`. Bis zur Entscheidung gilt Z62 als akzeptiertes Negativ.

Alles andere folgt aus Spezifikation oder Lastenheft oder ist im Vertrag mit Alternative, Grund und Folge entschieden (Form der Replay-Session-ID, Rangfolge im Modus `strict`, komprimierter Request-Body, Öffnen des Blobs in der Transaktion, Auswahl ohne Auslieferung).

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (8×, Ausgang `verkörpert` in `AGENTS.md` §3.10); die Zähler am 2026-10-31 aus den Dateien unter `evidence/`: `randform-erst-im-review-entschieden` 22, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 29, `universalaussage-im-vertrag-ungeprueft` 12 (verkörpert in `AGENTS.md` §3.16), `groessenmarke-ueberschritten-ohne-rueckfuehrung` 18 (verkörpert in `harness/mk/groessenmarke.mk`, `make groessenmarke`); der Slice ist nebenläufig (parallele Replay-Sessions, Cursor): das Gate fährt `-race` (`make test-race`, `SPEC-122`), die nebenläufigen Fänger ordnen weiter mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8b), und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
