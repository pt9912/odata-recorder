# Slice slice-admin-lebenszyklus: Admin-Rahmen, Teil 4 — Serve, Anhalten, Schliessen und Grenzen des Servers

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-QA-04`, `LH-RB-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-238` Nr. 3 b (angehaltener Server), Nr. 9 (`ctx` der Route), Nr. 10 (Grenzen des Servers), Nr. 11 (Serve, Anhalten, Schliessen) · `SPEC-136` Nr. 12 bis 14 (zweistufiges Herunterfahren, auf das Nr. 11 verweist) · `SPEC-124` *Ungeprüft* (Beenden einer Session kann länger dauern als jedes Schreib-Timeout) · `LH-FA-13.a` (Listener) · `ARC-007`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie. `SPEC-238` steht auf `main` (seit dem Merge von `slice-admin-listener-token`).

**Verantwortlich:** —

**Autor:** pt9912 (Planner). **Datum:** 2026-11-07.

**Herkunft:** Teil 4 des Schnitts von [slice-admin-listener-token](../done/slice-admin-listener-token.md) (siehe dort und [slice-admin-rahmen](slice-admin-rahmen.md) für die Reihenfolge der sechs Teile). Der Plan trägt die Zeilen A2, K5, S1 bis S10, X2 und G1 bis G3 der Tabelle des Architect (16 von 102); sie stehen im Wortlaut in §3. Gegenüber dem Vorschlag des Architect trägt dieser Teil zusätzlich A2 (Teil 3 des Vorschlags) und K5 (Teil 2 des Vorschlags): beide brauchen `Anhalten`, das erst hier entsteht.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Admin-Adapter bedient den übergebenen Listener und fährt zweistufig herunter wie der OData-Adapter (`SPEC-238` Nr. 11, `SPEC-136` Nr. 12 bis 14): `Serve` liefert nil, wenn das Schließen von `Anhalten` oder `Schliessen` stammt, sonst den letzten Fehler des `Accept`; `Anhalten` schließt den Listener ohne Warten, wiederholt und gleichzeitig wirkungslos; ein angehaltener Server beantwortet jeden Request, auch `GET /healthz`, mit 503 `ODR-E5010` und `Connection: close` vor Authentisierung, Zuordnung, Route und Audit; `Schliessen(frist)` wartet höchstens `frist`, beendet danach den `ctx` jeder laufenden Route und wartet ohne Frist (`ErrFristAbgelaufen`); der `ctx` der Route endet auch, wenn der Client trennt. Dazu die Werte des `http.Server` (`ReadHeaderTimeout` 10 s, `ReadTimeout` 30 s, `IdleTimeout` 120 s, `MaxHeaderBytes` 65536, kein `WriteTimeout`, `ErrorLog` verwirft, allgemeiner OPTIONS-Handler aus) und der Body der Route auf 1 MiB begrenzt (`SPEC-238` Nr. 10).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Reihenfolge des Herunterfahrens zusammen mit dem OData-Listener — `slice-bootstrap-m6` (Pflicht des Composition Root, `SPEC-136` Nr. 15); dieser Teil liefert `Anhalten` und `Schliessen`, ruft sie aber nirgends.
- TLS um den Listener — `slice-admin-tls`; der Adapter bedient jeden `net.Listener`, den er erhält, und kennt keine Standardadresse (K5).
- Kontextfehler der Route als Antwort (keine Antwort, kein Zählen), Fehlerabbildung, Panik, Zählen — `slice-admin-fehlerabbildung`; die Zeile X2 prüft nur, dass der `ctx` endet, nicht was die Route mit dem Fehler tut.
- Authentisierung, Zuordnung, Header — `slice-admin-authentisierung`, `slice-admin-rahmen` (Voraussetzungen dieses Teils; A2 und G2 setzen sie voraus).
- Audit des Abweisens und gleichzeitige Requests über alle Wege (N1) — `slice-admin-abweisen-audit`.
- Ein `WriteTimeout` — bewusst nicht gesetzt: das Beenden einer Session kann länger dauern als jedes Schreib-Timeout (`SPEC-124` *Ungeprüft*, entschieden in `SPEC-238` Nr. 10); die Folge (ein langsamer Client hält eine Verbindung bis `ReadTimeout`/`IdleTimeout`) trägt der Betreiber.
- Die Endpunkte, das Lesen von `ADMIN_*` und das Binary — `slice-admin-sessions`, `slice-admin-imports`, `slice-admin-workspaces`, `slice-bootstrap-m6`; das Binary bleibt unverändert.

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

- [ ] Serve, Anhalten, Schliessen: `Serve` und seine Rückgaben, `Anhalten` mit der Zulassung unter einer Sperre ohne Freigabe dazwischen, 503 `ODR-E5010` mit `Connection: close` für jeden Request eines angehaltenen Servers (auf bestehender Verbindung, ohne Route, ohne den Body zu lesen), `Schliessen(frist)` mit `ErrFristAbgelaufen` und zweimal wirkungslos, zwei Server teilen nichts (A2, K5, S1 bis S10); die Schleifen über die Requestarten *R* (A1, E4) und die Antwortarten (E2) tragen die Form 503.
- [ ] Grenzen und `ctx`: die Werte des `http.Server` und der Body auf 1 MiB, `OPTIONS *` durch die Authentisierung, der `ctx` der Route endet beim Trennen des Clients und beim Ablauf der Frist (G1 bis G3, X2).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Für diesen Slice: kein neuer Code (`ODR-E5010` ist bekannt); das Binary bleibt unverändert (kein Aufruf von `admin` aus `bootstrap`, `src/bootstrap` ohne Änderung), in §7 vermerkt.
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
| `src/adapters/driving/admin/` | update | `Serve`, `Anhalten`, `Schliessen`, Zulassung unter einer Sperre, 503 vor allem anderen, `http.Server` mit den festen Werten, Begrenzung des `Body`, `ctx` der Route (`SPEC-238` Nr. 3 b, 9, 10, 11; `SPEC-136` Nr. 12 bis 14) |
| Tests | neu / update | Adapter gegen einen echten Listener auf `127.0.0.1:0` mit rohem Client (Keep-Alive-Verbindung, `OPTIONS *`) und gegen `ServeHTTP`; der Test-Server mit Client ist der neue Fake; `listenerFake`, `routeFake` aus `slice-admin-rahmen`; Kanäle statt Schlaf; der Adapter importiert weder Adapter noch Service (`make a-check`) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stammen vom Architect (2026-11-07, Plan von [slice-admin-listener-token](../done/slice-admin-listener-token.md)); der Schnitt gibt jedem Teil ganze Zeilen im Wortlaut (A Ablauf, K Konfiguration, S Server, X Abbruch, G Grenzen). **Wachsende Mengen:** dieser Teil ergänzt die Form 503 in den Schleifen über die Requestarten *R* (A1, E4 in `slice-admin-authentisierung`) und die Antwortarten (E2); die Formen „Route mit Fehler“ und „Route mit Panik“ kommen mit `slice-admin-fehlerabbildung`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A2 | (Nr. 3 b) Ein angehaltener Server antwortet jedem Request mit 503 `ODR-E5010` vor Authentisierung, Zuordnung, Route und Audit | Schleife über Requestarten R (ohne Token, mit Token, healthz, unbekannter Pfad, Route): 503; `Vergleich`, Route und Audit 0 | Prüfung auf angehalten hinter der Authentisierung | andere |
| K5 | (Nr. 1) Der Adapter hat keine Standardadresse, bedient den übergebenen `Listener`, und zwei Server teilen nichts | Server A und B auf zwei Listenern: `Anhalten` von A lässt B 200 liefern; AST-Test: das Paket ruft nirgends `net.Listen` | `net.Listen("tcp", ":8081")` im Paket | andere |
| S1 | (Nr. 11; `SPEC-136` Nr. 12) `Serve` liefert nil, wenn das Schließen von `Anhalten` oder `Schliessen` stammt, und sofort nil, wenn es erst danach beginnt | Fälle: `Anhalten` während `Serve`, `Schliessen` während `Serve`, `Serve` nach `Anhalten`; echter TCP-Listener | `http.ErrServerClosed` durchgereicht | andere |
| S2 | (Nr. 11; `SPEC-136` Nr. 12) Sonst liefert `Serve` den letzten Fehler des `Accept`, nie nil | `listenerFake` mit Accept-Fehler E; von außen geschlossener Listener; Fehler des Accept und `Anhalten` unmittelbar danach | Fehler verschluckt | andere |
| S3 | (Nr. 11; `SPEC-136` Nr. 13 a, d) `Anhalten` schließt den Listener, kehrt ohne Warten zurück, ist wiederholt und gleichzeitig wirkungslos | `routeFake` blockiert; 16 gleichzeitige `Anhalten` in 20 Läufen: jede Rückkehr, `Dial` scheitert, `listenerFake` zählt `Close` | `Anhalten` wartet auf Handler | nebenläufig |
| S4 | (Nr. 11; `SPEC-136` Nr. 13 b) Ein Request nach `Anhalten` auf einer bestehenden Verbindung ergibt 503 `ODR-E5010` mit `Connection: close`, ohne Route, ohne den Body zu lesen | Keep-Alive-Verbindung vor `Anhalten` (roher Client), dann Formen: `GET /healthz`, `GET` einer Route, `POST` mit angekündigtem Body, `HEAD`, `OPTIONS *`, unbekannter Pfad | 503 ohne `Connection: close` | nebenläufig |
| S5 | (Nr. 11; `SPEC-136` Nr. 13 c) Ein vor `Anhalten` zugelassener Request läuft zu Ende | `routeFake` blockiert, `Anhalten`, Freigabe: die Antwort 200 ist vollständig | Zulassung nach dem Handler geprüft | nebenläufig |
| S6 | (Nr. 11; `SPEC-136` Nr. 13 c) Die Zulassung (Prüfen auf angehalten, Zählen) liegt unter einer Sperre ohne Freigabe dazwischen | Wirkungs-Fänger: 64 Requests gegen gleichzeitiges `Anhalten` und `Schliessen` in 200 Läufen: jeder Request ist 503 oder vollständig 200; nach `Schliessen` läuft kein Handler und keiner beginnt; Rate je Mutation im Plan | Sperre zwischen Prüfen und Zählen freigegeben | nebenläufig |
| S7 | (Nr. 11; `SPEC-136` Nr. 14) `Schliessen(frist)` mit `frist` unter 0 ist `ErrUngueltig` und ändert nichts | Aufruf mit -1 ns: Fehler; danach antwortet der Server 200 | Prüfung entfernt | andere |
| S8 | (Nr. 11; `SPEC-136` Nr. 14) `Schliessen` wartet höchstens `frist` (0: nicht); sind alle Requests zurück, ist die Rückgabe nil | `routeFake` blockiert: `frist` 0 gibt sofort `ErrFristAbgelaufen`; bei Freigabe vor Ablauf nil | Rückgabe nil bei laufendem Handler | nebenläufig |
| S9 | (Nr. 11; `SPEC-136` Nr. 14) Nach Ablauf endet der `ctx` jeder laufenden Route, `Schliessen` wartet danach ohne Frist auf jeden Handler und liefert `ErrFristAbgelaufen` | `routeFake` wartet auf `ctx.Done()` und läuft dann zu Ende; der Zähler laufender Handler ist bei der Rückkehr 0 | Rückkehr vor dem Ende des Handlers | nebenläufig |
| S10 | (Nr. 11; `SPEC-136` Nr. 14) Nach der Rückkehr läuft kein Handler und keiner beginnt; ein weiteres `Schliessen` liefert nil | Zähler 0 und ein neuer `Dial` scheitert; zweites `Schliessen` nil | zweites `Schliessen` liefert `ErrFristAbgelaufen` | andere |
| X2 | (Nr. 9) Der `ctx` der Route endet, wenn der Client trennt und wenn `Schliessen` die Frist erreicht | `routeFake` wartet auf `ctx.Done()` und meldet `ctx.Err()`: Fall 1 der Client schließt die Verbindung, Fall 2 `Schliessen` mit abgelaufener Frist; beide `context.Canceled` | `ctx` gleich `context.Background()` | nebenläufig |
| G1 | (Nr. 10) Werte des `http.Server`: `ReadHeaderTimeout` 10 s, `ReadTimeout` 30 s, `IdleTimeout` 120 s, `MaxHeaderBytes` 65536, `WriteTimeout` 0, `ErrorLog` verwirft, der allgemeine OPTIONS-Handler ist abgeschaltet | der Test liest die Felder des `http.Server` (Zugriff im Paket) mit Literalen, Zählprüfung 7; `ErrorLog`: Panik der Route schreibt kein Byte nach stderr oder `log.Default()` | `ReadTimeout` 30 s zu 3 s; `DisableGeneralOptionsHandler` entfernt | andere |
| G2 | (Nr. 10) `OPTIONS *` geht durch die Authentisierung | rohe TCP-Verbindung mit `OPTIONS * HTTP/1.1` ohne Token: 401 mit Fehlerobjekt | `DisableGeneralOptionsHandler` entfernt (200 von `net/http`) | andere |
| G3 | (Nr. 10) Der `Body` der Route ist auf 1 MiB begrenzt: 1048576 Bytes sind lesbar, das 1048577. ergibt `*http.MaxBytesError` | Fälle 1048575, 1048576, 1048577 Bytes | Grenze 1048577 | andere |

Zeilen gesamt 16: *a* = 7 (S3, S4, S5, S6, S8, S9, X2), *b* = 9. G1 gehört hierher, weil der Teil den `http.Server` baut; sein `ErrorLog`-Fall (Panik der Route) trifft in diesem Teil die Panik ohne Abfangen durch den Adapter, mit `slice-admin-fehlerabbildung` fängt der Adapter die Panik selbst (B8): dessen Implementer hält G1 grün (Risiko in §6).

**Pflicht-Mengen mit Zählprüfung gegen die Quelle** (Architect, Auftrag C): die Requestarten *R* in A2 (5 Elemente), die Formen von S4 (6), die Fälle von G1 (7 Felder) und G3 (3 Grenzen) sind Listen im Test mit Zählprüfung gegen die Zahl in der Zeile; das Orakel trägt Statuscodes, Codes und Header als Literale aus dem Vertrag.

**Unbeschränkte Menge und ihre Prüfgrenze (akzeptiertes Negativ):** die Zahl gleichzeitiger Requests (Prüfgrenze 64 in S6; Grund: der Zustand ist ein Zähler unter einer Sperre; Folge: bei 1000 gleichzeitigen Requests ist das Verhalten durch die Eigenschaft der Sperre, nicht durch einen eigenen Fänger getragen).

**Zweige als Menge, Runden unter dem Detektor** (Architect, Auftrag D; Zählprüfung gegen die Quelle: das Profil von `go test -race -coverprofile` über das Paket `admin` hat null Blöcke mit Zähler 0, der Implementer nennt jeden Ausreißer mit Datei und Funktion im Bericht):

| Zweig | Fänger | Runde (ein Lauf je Zweig unter `-race -coverprofile`) |
|---|---|---|
| `Serve`: Rückgabe nil, Fehler des Accept; `Anhalten`; 503; `Schliessen`: Frist unter 0, alle zurück, Ablauf, zweites | S1 bis S10, A2 | Fehler des Accept (S2) und Ablauf der Frist (S9): je ein Lauf |
| Werte des Servers, `OPTIONS *`, Begrenzung des `Body` | G1 bis G3 | – |

**Aufrufstellen und zählende Fakes** (Architect, Auftrag E; „jeder Fall je Aufrufstelle“ gilt für diese Stellen):

| Aufrufstelle | Port | Fall je Stelle | Fake und Zählung |
|---|---|---|---|
| `Accept` und `Close` des Listeners | `net.Listener` | S2, S3, K5 | `listenerFake`: zählt `Accept` und `Close`, Fehler des Accept einstellbar |
| genau eine je Route | `Handler` | S5, S8, S9, X2 | `routeFake`: zählt, hält `Anfrage` und `ctx` (dieselbe Identität), blockiert auf Kanälen |
| genau eine, vor allem anderen | `IDs.KorrelationsID` | A2 (auch das 503 trägt genau eine ID) | `zaehlendeIDs` |
| keine beim 503 | `Audit.Append`, `Telemetrie.CountError`, `Vergleich` | A2 | `FakeAudit`, `FakeTelemetry`, zählende Funktion des Tests: 0 |

*f* = 1 (der Test-Server mit rohem Client auf einem echten Listener); `listenerFake`, `routeFake` und `zaehlendeIDs` liegen seit `slice-admin-rahmen`.

**Raten der Fänger** (`AGENTS.md` §3.10): Eine Zusage über Nebenläufigkeit (S3 bis S6, S8, S9, X2) gilt nur als gefangen, wenn ihre Mutation in jedem Lauf rot wird; der Implementer trägt die Tabelle *Zusage · Mutation · roter Test* und die Raten je Mutation (Zahl der Prozesse, Messender, Last) in einen Absatz von §3 ein, bevor er an den Reviewer übergibt. S6 ist der Wirkungs-Fänger zu „Prüfen und Zählen unter einer Sperre ohne Freigabe dazwischen“; das Fehlen der Sperre überhaupt fängt der Detektor von `make test-race` je Prozess (`-test.count=1`).

**Größe der Sicht des Architect** (`SPEC-128` Nr. 1, Aufteilung durch den Planner): 16 Zeilen (*a* = 7, *b* = 9), *f* = 1, *p* = 100.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.
- Reports von Reviewer und Verifier tragen Kennungen von ADRs nie nackt, sondern als auflösenden Link, sonst meldet `make docs-check` `id-unlinked` (`BEO-REPO/report-der-rolle-faellt-docs-check`); der Planner reicht den Satz an beide weiter.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-admin-authentisierung](slice-admin-authentisierung.md) liegt in `done/` (Code-Abhängigkeit: die Authentisierung vor der Zuordnung, ohne die A2 und G2 keinen Gegenstand haben; über sie transitiv `slice-admin-rahmen`). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-238` Nr. 3 b, 9, 10, 11) seit dem Zweig von `slice-admin-listener-token` (`AGENTS.md` §3.12). Stand der Bedingungen beim Schnitt (Planner, 2026-11-07, `main` = `aadbef0`): `slice-audit-log` und `slice-telemetry-port` liegen in `done/`; der Teil ruft beide Ports nicht.

**Größenmarke** (Planner, 2026-11-07, `SPEC-128` Nr. 1; Prosa-Form, der Block folgt beim Start): 16 Zeilen der Tabelle, *a* = 7, *b* = 9, *f* = 1, *p* = 100: *S* = 85·7 + 80·9 + 200·1 + 100 = 1615, *M* = 2200, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Pfad des Blocks beim Start: `src/adapters/driving/admin/`; `produktcode` 100, `fakes` 1. `basis`, `schaetzung` und `marke` setzt der Planner beim Start in einem Commit vor dem ersten Code-Commit.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 2200 (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein Teil, der sich ohne geteilten Mechanismus abtrennen lässt, steht dann vor dem Code in §4, sonst steht die Messung mit Grund im Plan (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert — Carveout?): der Architect ändert `SPEC-238` Nr. 10 oder 11 oder `SPEC-136` Nr. 12 bis 14 (Rückgabe an den Planner), oder `net/http` erlaubt die Zusage von S4 (503 auf bestehender Verbindung ohne den Body zu lesen) nicht in der festgelegten Form.

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-07, Zeile dieses Teils): jede Zelle nennt eine Nummer von `SPEC-238` (Rahmen), `SPEC-136`, auf die `SPEC-238` Nr. 11 verweist, oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Herunterfahren des Listeners (zweistufig) | Nr. 11 (`SPEC-136` Nr. 12): `Serve` nach `Anhalten` kehrt sofort mit nil zurück | `SPEC-136` Nr. 14: `Schliessen(0)` wartet nicht | `SPEC-136` Nr. 14: Frist unter 0 ist `ErrUngueltig`; Nr. 13 b: Request nach `Anhalten` ist 503 `ODR-E5010` | Nr. 9: der `ctx` der Route endet beim Ablauf der Frist; `SPEC-136` Nr. 14: danach wartet `Schliessen` ohne Frist | `SPEC-136` Nr. 13 d und Nr. 14: `Anhalten` und `Schliessen` wiederholt und gleichzeitig wirkungslos |

**Zusatztabelle: Paare von Verstößen am selben Schlüssel und Modi** (`AGENTS.md` §3.16; die Fänger sind Zeilen der Tabelle in §3):

| Schlüssel und Paar oder Modus | Ergebnis (Ort) | Fänger |
|---|---|---|
| Header: ungültige Form von `Authorization` und angehaltener Server | 503 `ODR-E5010`, kein Eintrag (`SPEC-238` Nr. 3 b) | A2 |
| Modus `GET /healthz` bei angehaltenem Server | 503 `ODR-E5010` (Nr. 3 b) | A2, S4 |
| Modus Request auf bestehender Verbindung nach `Anhalten`: `GET`, `POST` mit angekündigtem Body, `HEAD`, `OPTIONS *`, unbekannter Pfad | 503 mit `Connection: close`, ohne Route, ohne den Body zu lesen (Nr. 11, `SPEC-136` Nr. 13 b) | S4 |
| Modus `Schliessen` mit Frist 0, mit Ablauf, mit Freigabe vor Ablauf, zweites `Schliessen` | `ErrFristAbgelaufen`, `ErrFristAbgelaufen`, nil, nil (`SPEC-136` Nr. 14) | S8, S9, S10 |

- Der `ErrorLog`-Fall von G1 (Panik der Route schreibt kein Byte) setzt in diesem Teil voraus, dass die Panik bis `net/http` durchreicht; `slice-admin-fehlerabbildung` fängt sie im Adapter (B8) und lässt den Fall leer — **Ausgang:** offen; Adresse: Planner (beim Start von `slice-admin-fehlerabbildung`: der Fall G1 bleibt dort grün, etwa mit `http.ErrAbortHandler` oder einem Fehler außerhalb der Route; die Zeile G1 liegt in diesem Plan, die Mutation „`ErrorLog` nicht verworfen“ bleibt zu fangen).
- Kein `WriteTimeout` (`SPEC-124` *Ungeprüft*: das Beenden einer Session dauert länger als jedes Schreib-Timeout) — **Ausgang:** entschieden (Architect, 2026-11-07, `SPEC-238` Nr. 10): kein `WriteTimeout`, `ReadTimeout` 30 s; die Folge trägt der Betreiber (langsame Clients halten eine Verbindung bis `IdleTimeout`).
- Die Schleifen über *R* und die Antwortarten (A1, E2, E4 in `slice-admin-authentisierung`) wachsen um die Form 503 — **Ausgang:** entschieden (Planner, Schnitt): dieser Teil trägt sie ein (DoD Punkt 1).
- Die Reihenfolge des Herunterfahrens beider Listener (`SPEC-136` Nr. 15) ist Pflicht des Composition Root — **Ausgang:** entschieden (Planner): `slice-bootstrap-m6` (dessen §1 und DoD, am 2026-11-07 eingetragen).
- `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (13×): S3 bis S6, S8, S9 und X2 sind Zusagen über Nebenläufigkeit; jede Mutation muss in jedem Lauf rot werden (`AGENTS.md` §3.10) — **Ausgang:** weiter offen → `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` im Register.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (12×): jeder Zweig von `Serve` und `Schliessen` hat einen Lauf unter dem Detektor (Z-Liste in §3) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): `listenerFake` und `routeFake` zählen jede Operation — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- Größe: die Marke steht als Prosa in §4 (*S* = 1615, *M* = 2200) — **Ausgang:** offen; Adresse: Planner beim Start (Block mit `basis`, `schaetzung` und `marke` in einem Commit vor dem ersten Code-Commit).
- Gate-Kette (Nr. 7 der Übergangsregeln von `welle-benutzerhandbuch`): der Teil ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette im Plan (§3).

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

**Vorgelagert — offene Beobachtungen sichten:** Register beim Schnitt nachgezogen (Planner, 2026-11-07, `main` = `aadbef0`, Zähler aus den Dateien unter `evidence/`): einschlägig sind `nebenlaeufiger-lauf-nur-im-glueck-gruen` 13×, `zweig-ohne-lauf-unter-dem-detektor` 12×, `hilfe-ersetzt-argument-verdeckt-eigenschaften` 4×, `panik-eines-fakes-verdeckt-den-benannten-test` 4×, `betriebsregel-fuer-repo-dateien-verletzt` 9×, `vertrag-sagt-mehr-zu-als-ein-gate-prueft` 44×, `randform-erst-im-review-entschieden` 34×, `rate-ohne-detektor-schwankt-mit-der-last`, `schaetzformel-ueberschaetzt-den-slice` 21×, `report-der-rolle-faellt-docs-check` 5× (Verkörperung, soweit die Schwelle erreicht ist, beim Lese-Schritt der Welle-Closure); kein neues Verzeichnis, kein Zähler gesetzt.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
