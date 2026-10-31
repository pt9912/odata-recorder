# Slice slice-http-lebenszyklus: HTTP-Adapter — Lebenszyklus, Grenzen des Servers und 503

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-FA-02`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md)

**Berührte Spec-Stellen:** `LH-FA-02.c` · `SPEC-059` · `SPEC-109` Nr. 2 · `SPEC-136` Nr. 2, 3b, 4c, 4g, 4h, 7a, 8, 11 bis 16 und *Ungeprüft* (Nr. 3b, 11 bis 15 stehen seit dem Kern in der Spezifikation, die Zusätze hat der Planner vor dem ersten Code-Commit nach dem Entwurf des Architect geschrieben; der Implementer schreibt keinen Text der Spezifikation) · `SPEC-137` · `SPEC-230` Nr. 8 · `SPEC-121` Nr. 2e (Grund der Fenster W3 und W4) · `SPEC-124` Nr. 14.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-http-lebenszyklus)

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der `Server` des Kerns (`slice-http-adapter`) schließt in zwei Schritten (`Anhalten`, `Schliessen(frist)`), weist neue Requests auf bestehenden Verbindungen mit 503 `ODR-E5010` ab, trägt feste Grenzen (`SPEC-136` Nr. 11), und `ODR-E5010` steht in der Code-Tabelle (35 Codes). Der Schnitt gegen den Kern ist vor dem Code gemacht (`AGENTS.md` §3.14): der Lebenszyklus ist der einzige Teil ohne geteilten Mechanismus mit ihm. Er ändert vom Kern drei Stellen: `New` (die Grenzen und der Basis-Kontext des Servers, `BaseContext`), `Serve` (Rückgabe `nil` nach `Anhalten`; heute nie `nil`, der bestehende Test auf „Serve liefert nie nil“ gilt danach nur für den von außen geschlossenen Listener, L7) und den Eintritt des Handlers (Zulassung und 503-Pfad; `EnableFullDuplex` als erste Anweisung). `ErrorLog` auf `io.Discard` und `DisableGeneralOptionsHandler` stehen schon im Kern. Dazu kommt die eigene Sperre.

**Übernimmt:** die Sendung aus `slice-session-beenden` („Fänger der beiden Schritte“, Zeilen L3 bis L12 und L16 der Tabelle in §3) und das Risiko „Request-Größe und Zeitgrenzen“ aus `slice-http-adapter` (Fänger L1, L2); `slice-http-adapter` führt beide als übergeben (§1 und §6 dort). Der Vertrag `SPEC-136` Nr. 3b, 11 bis 15 stand mit dem Kern in der Spezifikation. Die Zusätze des Architect (Nr. 2, 4c, 4g, 4h, 7a, 8, 11 bis 14, 16, *Ungeprüft*), `SPEC-137`, die Zeile in `SPEC-109` Nr. 2, `SPEC-230` Nr. 8 und die Historie hat der Planner vor dem ersten Code-Commit geschrieben; Code und die Zählstellen im Code gehören dem Implementer.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Reihenfolge des Herunterfahrens im Composition Root (`Anhalten` → `CloseAll` → `Schliessen(rest)` → Repositories) — `slice-bootstrap-m2` (dort in §1 und DoD *Herunterfahren* eingetragen); der Adapter liefert die zwei Schritte, nicht die Reihenfolge.
- TLS des Listeners — `slice-listener-tls`, Bestand der Welle [welle-betrieb](../welle-betrieb.md).
- Spans und Zähler je Fehlercode — `slice-telemetry-einbau-gateway`; der Kern liefert `ODR-E5010` nicht aus, die Zählstellen des neuen Codes (§3) liegen bei diesem Slice.
- Die Berichtigung von `.a-check.yml` (`ports-support`, Rückgabe a aus `slice-http-adapter`) — eigener Lauf des Architect (Berichtigung nach `ARC-021`, keine Senkung); kein Gegenstand dieses Vertrags.

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

- [ ] Lebenszyklus: `Anhalten` und `Schliessen(frist)` in zwei getrennten Schritten, neue Requests auf bestehenden Verbindungen mit 503 `ODR-E5010` und `Connection: close`, `Serve` endet mit `nil`, die Zulassung ist ein Abschnitt, der `ctx` endet mit der Frist auch bei ungelesenem Body (Zeilen L3 bis L16, L18 und L20 der Tabelle in §3; `LH-FA-02.c`, `SPEC-059`, `SPEC-124` Nr. 14).
- [ ] Feste Grenzen des Servers (`ReadHeaderTimeout`, `IdleTimeout`, `MaxHeaderBytes`, kein `ReadTimeout` und kein `WriteTimeout`) und kein Log der Bibliothek (`ErrorLog` verwirft; Zeilen L1, L2, L19).
- [ ] Der Code `ODR-E5010` in der Tabelle (503, Ziel `service`, 35 Codes) samt allen Zählstellen aus §3 (Zeile L17; `LH-QA-05.a`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-109` Nr. 2, `SPEC-137`, `SPEC-230` Nr. 8, Historie; vom Planner vor dem Code geschrieben, der Implementer prüft den Stand).
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
| `src/adapters/driving/httpodata/` | update | Zustand der Zulassung (`sync.Mutex`, `angehalten`, `geschlossen`, `sync.WaitGroup`), `Anhalten`, `Schliessen`, `Serve` mit `nil`, 503-Pfad, `EnableFullDuplex`, Grenzen und `BaseContext` in `New` |
| `src/hexagon/model/fehler.go` | update | `CodeE5010` nach `CodeE5009`, Zeile in `codeZeilen()` |
| `spec/spezifikation.md` | update (Planner, vor dem ersten Code-Commit) | `SPEC-137`, `SPEC-109` Nr. 2, `SPEC-230` Nr. 8, Historie und die Zusätze zu `SPEC-136` nach dem Entwurf des Architect; der Implementer ändert die Datei nicht |
| Tests des Pakets und der Zählstellen | neu / update | die Tabelle unten; Zählstellen siehe die Liste darunter |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Entwurf des Architect (Vertrag `SPEC-136`, `SPEC-137`) vor dem ersten Code-Commit (`AGENTS.md` §3.12). Die Nummern L1 bis L17 bleiben, damit die Verweise aus `slice-session-beenden` und `slice-http-adapter` („L3 bis L12 und L16“, „L1, L2“) stehen bleiben; L18 bis L20 sind neu. Jede Menge ist eine Schleife über alle Elemente (`AGENTS.md` §3.17); die Elementzahl nennt der Fänger, der Implementer berichtet je Zeile eine Mutation an einem Element, das der Plan nicht nannte.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| L1 | Voreinstellungen des Servers: `ReadHeaderTimeout` 10 s, `IdleTimeout` 120 s, `MaxHeaderBytes` 65536, `ReadTimeout` und `WriteTimeout` 0, Nr. 11 | `TestServerVoreinstellungen` (im Paket, liest die Felder des `http.Server`): 5 Felder, je Wert | `ReadTimeout` gesetzt (30 s): rot; `MaxHeaderBytes` auf den Standard der Bibliothek (1 MiB): rot | andere |
| L2 | Ein Kopf über der Grenze ergibt 431 von `net/http`, ein Kopf darunter geht durch, Nr. 4i, 11 (die Grenze gilt auf rund 4 KiB genau, Nr. 11) | `TestKopfGrenze`: 2 Fälle weit von der Grenze (Kopf 60 KiB geht durch und der Fake sieht ihn; Kopf 100 KiB ergibt 431 Klartext, `Verarbeiten`- und `IDs`-Zähler 0) | `MaxHeaderBytes` nicht gesetzt (Standard 1 MiB): Fall 2 liefert 200, rot | andere |
| L3 | `Anhalten` schließt den Listener und kehrt ohne Warten zurück, auch wenn ein Request läuft, Nr. 13a, 13d | `TestAnhaltenSchliesstListener`: ein Request hängt im Fake (Kanal); `Anhalten` kehrt in 2 s zurück (Kanal, kein Schlafen); danach scheitert `net.Dial`; der hängende Request ist noch offen und der Fake wird danach freigegeben | `Anhalten` ruft `http.Server.Shutdown` (wartet): kehrt nie zurück, rot per Frist; Listener nicht geschlossen: `Dial` gelingt | nebenläufig |
| L4 | Nach `Anhalten` ergibt jeder Request auf einer bestehenden Verbindung 503 `ODR-E5010` mit `Connection: close`, ohne `Verarbeiten`, ohne Lesen des Bodys, mit `X-Correlation-Id`, die Antwort kommt vollständig an, die Verbindung wird nicht wiederverwendet, Nr. 2, 3b, 8, 13b | `TestAngehaltenDreiNullDrei`: Menge der Request-Formen, **7 Elemente**: GET; POST mit 1 KiB Body; POST mit 1 MiB Body (der Client sendet den Body weiter); POST mit angekündigtem 1 MiB und 10 gesendeten Bytes (die Antwort muss kommen, bevor der Body fertig ist); HEAD; `OPTIONS *`; `CONNECT h:443`. Je Element eine eigene Keep-Alive-Verbindung, die vor `Anhalten` einen Request erledigte (alle sieben stehen vor `Anhalten`). Erwartet je Element: 503, Fehlerobjekt `ODR-E5010` (HEAD ohne Body), `Connection: close`, `X-Correlation-Id`, EOF danach (außer beim Element mit ungesendetem Body, dort schließt der Client); der Zähler von `Verarbeiten` bleibt bei 7, der von `IDs` wächst um 1 je Request | Status 500: rot; `Connection: close` fehlt: der Client sieht die Verbindung weiter offen, rot; `Verarbeiten` doch gerufen: Zähler über 7, rot; Body vor dem 503 gelesen: Element 4 liefert keine Antwort, rot per Frist; `DisableGeneralOptionsHandler` entfernt: Element 6 liefert 200, rot | andere |
| L5 | Ein vor `Anhalten` zugelassener Request läuft zu Ende, vollständig, `Close` einmal, Nr. 13c | `TestLaufenderRequestEndet`: GET hängt im Fake; `Anhalten`; Fake freigeben; der Client erhält 200 mit allem Body (1 MiB), `Close` 1 | Handler bricht beim Anhalten ab: der Client erhält einen Fehler, rot | nebenläufig |
| L6 | Ein weiteres und ein gleichzeitiges `Anhalten` sind wirkungslos; nach der Rückkehr **jedes** Aufrufs ist der Listener zu, Nr. 13a, 13d | `TestAnhaltenWiederholt`: 8 Goroutinen rufen `Anhalten` gleichzeitig; jede prüft nach ihrer Rückkehr, dass `Dial` scheitert; danach nochmals; keine Panik, kein Fehler, ein laufender Request bleibt unberührt | zweites `Anhalten` schließt den Listener doppelt und gerät in Panik: rot; Flag per Vergleich-und-Tausch, `Close` außerhalb der Sperre: ein zweiter Aufruf kehrt vor dem `Close` zurück, `Dial` gelingt, rot nur mit Wahrscheinlichkeit (Rate je Mutation im Plan, Fenster W2) | nebenläufig |
| L7 | `Serve` endet mit `nil` nach `Anhalten` und `Schliessen`, auch wenn es erst danach beginnt; ein von außen geschlossener Listener ergibt einen Fehler, nie `nil` und nie `http.ErrServerClosed`, Nr. 12 | `TestServeEndetMitNil`: Menge der Enden, **6 Elemente**: (1) `Serve` läuft, dann `Anhalten`; (2) `Serve` läuft, dann `Schliessen(0)`; (3) `Anhalten` vor `Serve`; (4) `Schliessen(0)` vor `Serve`: Rückgabe `nil` (1 bis 4); (5) Listener von außen geschlossen, `Serve` läuft; (6) Listener geschlossen übergeben: Rückgabe ein Fehler, nicht `nil`, nicht `http.ErrServerClosed` | `Serve` gibt `http.ErrServerClosed` zurück: rot (1 bis 4); `Serve` gibt bei jedem Fehler des Accept `nil` zurück: (5), (6) rot; Flag nur im Lauf gesetzt, nicht beim Beginn: (3), (4) rot | andere |
| L8 | `Schliessen(frist < 0)` ergibt `model.ErrUngueltig` und ändert nichts; `0` ist gültig, Nr. 14 | `TestSchliessenNegativ`: Menge der Fristen, 3 Elemente: -1 ns und `math.MinInt64` ergeben `errors.Is(err, model.ErrUngueltig)`, danach nimmt der Server weiter Requests an (200); `0` (die Gegenseite der Grenze) ergibt kein `ErrUngueltig` | negative Frist wie 0: der Server ist danach zu, rot; Prüfung `frist <= 0`: `0` ergibt `ErrUngueltig`, rot | andere |
| L9 | `Schliessen` wartet höchstens `frist` und kehrt früher zurück, wenn alle Requests fertig sind, Nr. 14 | `TestSchliessenFrueh`: Request hängt im Fake; `Schliessen(30 s)` in der Hauptgoroutine; eine zweite Goroutine gibt den Fake nach 100 ms frei; Rückgabe `nil` nach weniger als 5 s und der Client erhielt 200 mit vollem Body | sofortiges Schließen der Verbindungen: der Client erhält einen Fehler, rot; Warten auf die volle Frist (auch ein verlorenes Aufwecken): Dauer über 5 s, rot | nebenläufig |
| L10 | Läuft die Frist ab, schließt `Schliessen` die Verbindungen, `ctx` des Requests endet, es wartet auf die Rückkehr des Handlers (`Close` gerufen) und liefert `ErrFristAbgelaufen` (nicht `model.ErrUngueltig`), Nr. 14 | `TestSchliessenFristAbgelaufen`: Request hängt im Fake bis `ctx.Done()`, dann liefert der Fake `Close` mit 150 ms Verzögerung; `Schliessen(200 ms)`: `errors.Is(err, ErrFristAbgelaufen)`, nicht `errors.Is(err, model.ErrUngueltig)`, nach dem Aufruf ist `Close` gerufen (Zähler 1), `ctx.Err() == context.Canceled` | Rückkehr vor dem Handler: `Close`-Zähler 0 bei der Rückkehr, rot; `nil` statt `ErrFristAbgelaufen`: rot | nebenläufig |
| L11 | `Schliessen(0)` wartet nicht; jede Frist ab 0 ohne laufenden Request kehrt sofort mit `nil` zurück, Nr. 14 | `TestSchliessenNull`: Request hängt im Fake bis `ctx.Done()`; `Schliessen(0)` kehrt in 5 s zurück (nach dem Ende des Handlers) mit `ErrFristAbgelaufen`; ohne laufenden Request: Menge der Fristen, **4 Elemente** (0, 1 ns, 30 s, `math.MaxInt64`), je ein frischer Server, Rückgabe `nil` in unter 5 s | Frist 0 als „ohne Ende“: kehrt nie zurück, rot per Frist; Überlauf bei `MaxInt64` (Frist als Zeitpunkt addiert): Rückgabe sofort `ErrFristAbgelaufen`, rot | nebenläufig |
| L12 | Die Dauer ist eine Frist, nicht die Summe: `Anhalten` plus `Schliessen(F)` dauern höchstens `F` plus die Zeit zum Beenden, Nr. 13, 14, 15 | `TestSummeDerFristen`: Request hängt bis `ctx.Done()`; Zeit t0 vor `Anhalten`, `Anhalten`, `Schliessen(300 ms)`; Dauer mindestens 300 ms und unter 550 ms | `Anhalten` wartet 300 ms zusätzlich: Dauer über 600 ms, rot; `Schliessen` wartet zweimal: rot | nebenläufig |
| L13 | `Schliessen` schließt Verbindungen ohne laufenden Request sofort, `Anhalten` lässt sie offen; ein abgewiesener Request zählt nicht als laufend, Nr. 13b, 13c, 14 | `TestIdleVerbindung`: zwei Verbindungen V1, V2, je nach einem Request im Leerlauf. Nach `Anhalten`: ein `Read` auf V1 mit 200 ms Frist endet im Timeout, nicht in EOF (offen); V2 sendet einen Request und erhält 503. Dann `Schliessen(30 s)` in unter 5 s mit `nil`; ein `Read` auf V1 liefert EOF | `Anhalten` schließt Idle-Verbindungen: V1 liefert EOF vor `Schliessen`, rot; `Schliessen` schließt sie nicht: V1 liefert keine EOF, rot per Frist; der 503-Pfad zählt als laufend und gibt nicht frei: `Schliessen(30 s)` dauert 30 s, rot | andere |
| L14 | `Schliessen` ohne vorheriges `Anhalten` hält an, Nr. 14 | `TestSchliessenOhneAnhalten`: `Schliessen(1 s)` ohne `Anhalten`: `Dial` scheitert danach, `Serve` kehrt mit `nil` zurück | `Schliessen` nur wartend: `Dial` gelingt, rot | andere |
| L15 | Ein weiteres `Schliessen` liefert `nil` und wartet nicht, Nr. 14 | `TestSchliessenZweimal`: zweimal nacheinander; das zweite in unter 1 s mit `nil` | zweites `Schliessen` panikt oder wartet: rot | andere |
| L16 | Nach der Rückkehr von `Schliessen` läuft kein Handler mehr und keiner beginnt; Prüfung und Zählen der Zulassung sind ein Abschnitt, Nr. 13c, 14 | `TestZulassungAtomar`: **Fall 0 (deterministisch, Haltepunkt am `IDs`-Fake):** der Fake blockiert den ersten Aufruf von `KorrelationsID`; `Anhalten`; `Schliessen(1 s)` kehrt mit `nil` zurück; dann den Fake freigeben: der Request ergibt 503, der Zähler von `Verarbeiten` bleibt 0. **Fall 1 (Wirkungs-Fänger):** 200 Requests gleichzeitig (Goroutinen) gegen `Anhalten` und `Schliessen(2 s)` aus der Hauptgoroutine nach 20 ms; nach der Rückkehr ist der Zähler der laufenden Handler im Fake 0, der Zähler von `Verarbeiten` wächst in den nächsten 200 ms nicht, und jeder Request endete mit 200 (zugelassen und beendet) oder 503/Verbindungsfehler; `-count=50` | Fall 0: Prüfen auf angehalten vor `IDs`, Zählen danach (ein Aufruf im Fenster): `Verarbeiten` läuft nach der Rückkehr von `Schliessen`, rot in jedem Lauf; Fall 1: Prüfen und Zählen in zwei Abschnitten ohne Aufruf dazwischen: rot nur mit Wahrscheinlichkeit, Rate je Mutation im Plan, der Detektor meldet es nicht (Fenster W1) | nebenläufig |
| L17 | `ODR-E5010` steht in der Tabelle (503, `service`, message „Service is shutting down“), 35 Codes, `SPEC-137`, Nr. 13b | Tests, die der Slice ändert: `model/fehler_test.go` (Zeile für `ODR-E5010`, die `fremd`-Liste wird `{200, 418, 507}`), `TestMessageFest` (message eindeutig und ohne `%{}<>`), `adapters/driven/telemetry/telemetry_test.go` und `regeln_test.go` (34 auf 35, 38 auf 39, `alleCodes` um `ODR-E5010`), `httpodata/fehler_test.go` (34 Paare × 2 Bindungen = 68 über `model.AlleCodes()`) | Code fehlt in der Tabelle: `AlleCodes` hat 34, rot; Status 502 statt 503: rot | andere |
| L18 | Der Zustand (`angehalten`, Zähler der Handler) wird nur unter der Sperre gelesen und geschrieben, Nr. 10, 13c | `TestZulassungAtomar` unter `make test-race` (Detektor), mindestens 20 Prozesse mit `-test.count=1`: dieselben Fälle wie L16 | Prüfung auf angehalten ohne Sperre (`angehalten` ohne Schutz gelesen): der Detektor meldet ein Datenrennen, rot je Prozess (Zahl im Plan) | nebenläufig |
| L19 | Der Adapter schreibt kein Log von `net/http` (Panik des Dienstes erscheint weder in `log.Default()` noch auf stderr), Nr. 16 | `TestKeinLogDerBibliothek`: `Verarbeiten` des Fakes panikt mit dem Wert „geheim-wert-L19“; `log.SetOutput` auf einen gesperrten Puffer; der Client sieht den Abbruch (EOF); der Puffer enthält den Wert nicht | `ErrorLog` nicht gesetzt: `net/http` schreibt die Panik mit dem Wert in `log.Default()`, rot (gemessen) | andere |
| L20 | Der `ctx` endet beim Ablauf der Frist auch bei ungelesenem Body, Nr. 4g, 14 | `TestSchliessenKontextUngelesenerBody`: POST mit angekündigten 100 Bytes und 3 gesendeten; der Fake liest den Body nicht und wartet auf `ctx.Done()`, dann `Close`; `Schliessen(200 ms)`: Rückgabe `ErrFristAbgelaufen` in unter 5 s, `Close` 1, `ctx.Err() == context.Canceled` | `BaseContext` nicht gesetzt (nur `http.Server.Close`): der `ctx` endet nie (gemessen), `Schliessen` hängt, rot per Frist | nebenläufig |

Zeilen L1 bis L20: **20**, davon **10 nebenläufig** (L3, L5, L6, L9, L10, L11, L12, L16, L18, L20) und **10 andere** (L1, L2, L4, L7, L8, L13, L14, L15, L17, L19).

**Raten (`AGENTS.md` §3.10).** Jede nebenläufige Zeile ordnet mit Kanälen. Der Implementer fährt die Mutation je nebenläufiger Zeile mit `-count=50` in einem Prozess und nennt „x von 50 rot“, für ein Datenrennen (L18) die Zahl roter Prozesse mit `-test.count=1` (mindestens 20 Prozesse), den Messenden und trägt beides in einen Absatz von §3 ein; „hoch“ steht nie ohne Zahl. Wahrscheinlich-rote Mutationen: L6 (Fenster W2), L16 Fall 1 (Fenster W1) und L12 (Zeitgrenzen unter Last, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`); deterministisch rot: L16 Fall 0 (Haltepunkt am `IDs`-Fake, nur für die Mutation mit einem Aufruf im Fenster), L3, L10, L11 und L20. Die Rate von L16 gilt nur für die Mutation ohne Aufruf im Fenster: eine Zulassung, die vor `IDs.KorrelationsID()` prüft und danach zählt, hat den Fake im Fenster, und Fall 0 macht sie in jedem Lauf rot. Fällt die Rate einer Zeile unter 50 von 50, steht sie als Negativ mit der Zahl und der Folge im Plan (L16: „ein Handler beginnt nach der Rückkehr von `Schliessen`“; L6: „ein zweiter Aufruf von `Anhalten` kehrt zurück, bevor der Listener zu ist“).

**Gemessene Raten (trägt der Implementer vor der Übergabe ein):** je Zeile Mutation · x von 50 bzw. Zahl roter Prozesse · Zahl der Prozesse · Messender · Mutation an einem Element, das der Plan nicht nannte. *Noch nicht gemessen.*

**Sperren und Fenster (`AGENTS.md` §3.10).** Die Zusage „unter einer Sperre, ohne sie dazwischen freizugeben“ (Nr. 13c) und die Sperre von `Anhalten`: je Fenster eine Zeile.

| Fenster | Seite | Was ein Freigeben dort ändern würde | Fänger oder Negativ |
|---|---|---|---|
| W1 Zulassung: zwischen Prüfen auf angehalten und Zählen | Handler | ein Handler beginnt nach der Rückkehr von `Schliessen` | L16 Fall 0 (Haltepunkt `IDs`-Fake, nur wenn ein Aufruf im Fenster liegt), Fall 1 (Wirkung, mit Rate); fehlende Sperre: L18 (Detektor) |
| W2 `Anhalten`: zwischen Setzen des Flags und Schließen des Listeners | `Anhalten`, zweiter Aufrufer | der zweite Aufruf kehrt zurück, der Listener ist noch offen (Nr. 13d) | L6 (`Dial` nach der Rückkehr jedes Aufrufs, mit Rate) |
| W3 `Anhalten`: nach dem Setzen des Flags, vor der Freigabe | Handler | keines: die Zulassung weist nach dem Setzen jeden Request ab, ein abgewiesener ist 503 | benanntes Negativ: Grund wie `SPEC-121` Nr. 2e: nach dem Setzen hängt keine Zusage an der Sperre, die laufenden können nur abnehmen; Folge: keine |
| W4 `Schliessen`: zwischen `Anhalten` und dem Warten auf die Handler | `Schliessen` | keines: nach dem Setzen nimmt die Zulassung keinen Handler an, der Zähler kann nur sinken | benanntes Negativ, Grund wie W3; Folge: keine |
| W5 `Schliessen`: Lesen des Zählers und Warten (verlorenes Aufwecken) | `Schliessen`, Handler | `Schliessen` kehrt erst mit Ablauf der Frist zurück | kein Fenster, wenn das Warten `WaitGroup.Wait` ist (Vorgabe); sonst L9 (Rückkehr unter 5 s bei Frist 30 s, mit Rate) |

Haltepunkt-Suche (Pflicht vor einem Negativ): W1 hat einen am `IDs`-Fake (Fall 0, für die Mutation mit Aufruf im Fenster); W2 bis W4 haben keinen an einem Argument des Tests, W3 und W4 sind benannte Negative aus Grund, nicht aus Bequemlichkeit.

**Zählstellen des neuen Codes `ODR-E5010`** (Tabelle 34 auf 35 Codes; im Commit dieses Slice, nicht des Kerns, denn der Kern liefert den Code nicht aus):

- `spec/spezifikation.md`: `SPEC-109` Nr. 2 (die 18. Zeile), §4 (`SPEC-137`), `SPEC-230` Nr. 8 (35) und die Historie hat der Planner vor dem Code geschrieben; `SPEC-109` Nr. 5 und `SPEC-129` Nr. 12 bleiben unverändert (kein Dienst liefert `ODR-E5010`), `SPEC-109` Nr. 2 trägt keinen Satz zur Zahl der Codes. Weitere Zählstellen der Zahl in der Spezifikation gibt es nicht.
- `src/hexagon/model/fehler.go`: `CodeE5010` nach `CodeE5009` (die folgenden Konstanten verschieben sich um 1; das Gate fängt es, keine Persistenz speichert `Code`-Zahlen), Zeile in `codeZeilen()`: `{CodeE5010, "ODR-E5010", []int{503}, ZielService, "Service is shutting down"}`.
- `src/hexagon/model/fehler_test.go`: Zeile `"ODR-E5010": {[]int{503}, "service"}`; die Liste `fremd` `{200, 418, 503}` wird `{200, 418, 507}` (sonst schlägt `TestCodeTabelle` fehl, weil `ODR-E5010` 503 trägt).
- `src/adapters/driven/telemetry/telemetry_test.go` (die 34 in den Zeilen 44 bis 49, 162 bis 164, 206 bis 207, 382, 413 mit `len(erwartet) != 35` und `len(codes) != 39`, 480 bis 496, darunter die Kommentare der Zeilen 480 und 483) und `regeln_test.go` (Zeilen 52 bis 58, 391 bis 400: `alleCodes` um `ODR-E5010`, 34 auf 35); `telemetry.go` folgt `model.AlleCodes()` und bleibt.
- `src/hexagon/ports/inbound/fehler_test.go`: keine Änderung (prüft `ODR-E5009`).
- `src/adapters/driving/httpodata/`: die Menge der (Code, Status)-Paare des Fängers A26 aus `slice-http-adapter` (`TestFehlerobjektJeCode`, `TestFehlerHeadAlleCodes`) wächst mit dem neuen Code: vor `ODR-E5010` 33 Paare (alle Stati ungleich 0 der 34 Codes, `ODR-E4001` mit zwei) × 2 Bindungen = 66, danach 34 Paare × 2 = 68 (503 ist ein neuer Status). Feste Zahlen: `fehler_test.go` Zeilen 82 bis 83 (`len(codes) != 33` und die Meldung, 33 auf 34) und der Kommentar in `klassen_test.go` Zeile 43 („alle 33 Paare“); die Schleife über `model.AlleCodes()` folgt von selbst.
- Der Implementer sucht den Rest mit `grep -rn 'AlleCodes\|\b34\b\|\b33\b' src` und nennt jeden Fund im Bericht.
- Pläne in `open/`, die 34 Codes zählen (vom Planner im selben Commit wie ein früherer Stand dieses Plans nachgezogen): `slice-handbuch-meldungskatalog`, `slice-handbuch-record-replay`, `slice-telemetry-spans`. `harness/README.md` nennt die Zahl der Codes nicht.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-http-adapter` liegt in `done/`. Die Spezifikation trägt `SPEC-136` Nr. 3b, 11 bis 15 mit den Zusätzen des Architect und `SPEC-137`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Reihenfolge der Welle: `slice-http-adapter` → `slice-http-lebenszyklus` → `slice-bootstrap-m2`.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <basis>..<letzter Commit des Slice>` über die `pfad`-Zeilen des Blocks, Tests eingeschlossen; `basis` ist der Commit, der Spezifikation und Plan vor dem ersten Code-Commit abschließt. Zählung nach der Tabelle in §3: *a* = 10 nebenläufige und *b* = 10 andere Zeilen (vom Planner an L1 bis L20 nachgezählt, gleich dem Entwurf des Architect), *f* = 1 (ein Test-Fake für Verarbeiter und `IDs` mit Kanälen, Zählern der laufenden Handler und von `Close`, Haltepunkt am `ctx` und verzögertem `Close`; der bestehende `dienst` in `fake_test.go`, die rohe Verbindung `roh`/`rohOffen` und `schranke` werden erweitert, ein Roh-Client mit Keep-Alive-Verbindungen für V1 und V2 ist darin eingeschlossen), *p* = 140 (Erwartung des Architect: `server.go` +100, `fehler.go` +8, `model/fehler.go` +4, Rest Reserve). Schätzung S = 1990, Marke 2700 nach `SPEC-128` Nr. 1.

```groessenmarke
pfad: src/adapters/driving/httpodata/
pfad: src/hexagon/model/fehler.go
pfad: src/hexagon/model/fehler_test.go
pfad: src/adapters/driven/telemetry/telemetry_test.go
pfad: src/adapters/driven/telemetry/regeln_test.go
produktcode: 140
fakes: 1
basis: 023617994ec0ff202a464b5dbad0f7c4cba213cc
schaetzung: 1990
marke: 2700
ersetzt: 2100 | der Entwurf zählte a = 8, b = 9, f = 0; drei neue Zeilen (L18, L19, L20), L16 mit zwei Fällen, die Mengen von Serve und Schliessen als Zeilen der Tabelle und ein Test-Fake
```

`basis` ist der Commit, der Spezifikation und Plan abschließt; die Marke ist ab dem ersten Code-Commit fest (`SPEC-128` Nr. 5 a).

**Die ältere Marke 2100 ist ersetzt** (`ersetzt` im Block, vor dem ersten Code-Commit): der Entwurf zählte a = 8, b = 9, f = 0 (S = 1540). Grund: drei neue Zeilen (L18, L19, L20), L16 mit zwei Fällen, die Mengen von `Serve` (6 Elemente) und `Schliessen` (Fristen) sind Zeilen der Tabelle, und der Test-Fake trägt ein *f*.

**Schnittfrage:** keine; S = 1990 liegt unter der Schwelle (`SPEC-128` Nr. 1). Für den Fall, dass ein Reviewer oder der Planner Zeilen hinzufügt und die Schwelle erreicht wird, schlägt der Architect den vorab benannten Teil vor: die Grenzen des Servers (L1, L2; 2 Zeilen, 160 plus 5 Produktzeilen), denn er teilt keinen Mechanismus mit dem Lebenszyklus (andere Felder von `New`, kein Zustand, keine Sperre). L17 (Code und Zählstellen) lässt sich nicht trennen, weil L4 `CodeE5010` braucht. Geschnitten wird vor dem Code; nach dem Code wird gemessen und mit Grund berichtet.

**Rückführungen — vorab benannt:**

- Eine Überschreitung der Marke ist ein Eintrag `ueberschreitung` mit Grund im Block, keine Rückführung nach dem Code (`SPEC-128` Nr. 10 c).
- `in-progress` → `open` (blockiert): der Kern liefert `Server` und den Eintritt des Handlers nicht in der Form, die der Entwurf voraussetzt (Rückgabe an den Architect, `AGENTS.md` §3.12).

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

- Das Fenster der Sperre zwischen Prüfen auf angehalten und Zählen (W1, L16) hat einen Haltepunkt am `IDs`-Fake nur für die Mutation mit einem Aufruf im Fenster (Fall 0, deterministisch); die Mutation ohne Aufruf im Fenster fängt nur der Wirkungs-Fänger mit Rate (Fall 1), und der Detektor von `make test-race` meldet sie nicht (alles unter Sperre) — **Ausgang:** offen; Adresse Implementer (Zahl der Läufe und Messender im Plan, §3 *Gemessene Raten*); fällt die Rate unter 50 von 50, steht die Zeile als Negativ mit der Zahl und der Folge „ein Handler beginnt nach der Rückkehr von `Schliessen`“.
- Der neue Code `ODR-E5010` verändert alle Zählstellen (Tabelle 34 auf 35, Zähler 35, `fremd`-Liste in `fehler_test.go`, 33 auf 34 Paare in `httpodata`); die Pläne `slice-handbuch-meldungskatalog`, `slice-handbuch-record-replay` und `slice-telemetry-spans` tragen die Zahl und sind nachgezogen — **Ausgang:** offen; Adresse Implementer, sucht den Rest mit `grep` und nennt jeden Fund im Bericht.
- Die Reihenfolge `Anhalten` → `CloseAll` → `Schliessen(rest)` hängt am Composition Root: ruft er `CloseAll` vor `Anhalten`, erhalten neue Requests `ODR-E5009` statt 503, wartet er mit der vollen Frist, dauert es die Summe — **Ausgang:** weiter offen; Adresse `slice-bootstrap-m2` (Sendung dort in §1 und DoD angenommen, Fänger für den Root dort).
- Der Root ruft `Schliessen` genau einmal (ein gleichzeitiges `Schliessen` ist nicht zugesagt, `SPEC-136` *Ungeprüft*, Nr. 15). `slice-bootstrap-m2` §1 und DoD nennen die Reihenfolge, aber „einmal“ nicht für `Schliessen` (`AGENTS.md` §3.13) — **Ausgang:** offen; Adresse Planner auf `main` (ein Halbsatz in §1 oder DoD von `slice-bootstrap-m2`; der Planner dieses Zweigs ändert Pläne in `open/` nicht).
- Rückgaben aus [slice-http-adapter](../done/slice-http-adapter.md) §6 an den Architect, vor dem ersten Code-Commit entschieden (Ort in der Spezifikation, Fänger in §3):

| Rückgabe | Entscheidung | Ort · Fänger | Ausgang |
|---|---|---|---|
| `ErrorLog` | bleibt `io.Discard`: die Meldung einer Panik trägt Wert und Stapel, die Request-Werte nennen können, `SPEC-136` §4 verbietet Credentials und unredigierte Bodies im Log; Folge: keine Zeile zu einer Panik des Dienstes (akzeptiertes Negativ, kein Folge-Slice: eine Senke mit Redaktion ist additiv, `Konfig` wächst nach Nr. 1) | Nr. 16 (zweiter Absatz), *Ungeprüft* · L19 | entschieden (Architect) |
| `EnableFullDuplex` und Wiederverwendung beim 503-Pfad | bleibt erste Anweisung des Handlers auf jedem Pfad; der 503-Pfad liest den Body nicht und setzt `Connection: close`, ist also nicht wiederverwendbar; der Normalpfad verwirft einen Rest bis 256 KiB, darüber schließt die Verbindung | Nr. 4h, 13b, *Ungeprüft* · L4 (Elemente 3, 4), L5 | entschieden (Architect) |
| `DisableGeneralOptionsHandler` und `OPTIONS *` bei angehaltenem Server | das Mittel steht in der Spezifikation; `OPTIONS *` ist ein Request wie jeder und ergibt bei angehaltenem Server 503 `ODR-E5010` mit `Connection: close` | Nr. 2, 4c, 13b · L4 (Element 6); im Normalbetrieb A6 Fall 2 des Kerns | entschieden (Architect) |
| Typed-nil in `Listener` (`Serve`, `Anhalten`) | akzeptiertes Negativ, keine Prüfung in `New`: entsteht nur, wenn der Root das Ergebnis von `net.Listen` ohne Prüfung des Fehlers übergibt; Folge: Panik in `Serve` beim Start (der Prozess endet sichtbar) bzw. in `Anhalten` (die Sperre wird per `defer` frei) | *Ungeprüft* (Zeile ersetzt) · kein Fänger (benanntes Negativ) | entschieden (Architect) |
| nil-Zeiger `*inbound.Fehler` (Rückgabe e) | zählt nicht als dieser Fehler, sondern wie (3) von Nr. 7a; der Code handelt schon so (`f != nil`) | Nr. 7a (2) · A27 des Kerns | entschieden (Architect) |
| `.a-check.yml` / `ports-support` (Rückgabe a) | kein Gegenstand dieses Slice (§1) | `ARC-021` | weiter offen; Adresse Architect, eigener Lauf |

- Akzeptierte Negative aus dem Entwurf des Architect (Grund und Folge stehen in `SPEC-136` *Ungeprüft*; keines ist einen Folge-Slice wert: jedes ist einmalig und folgenlos außerhalb des Prozessendes und kehrt im Betrieb nicht unbemerkt wieder): Verbindungen im Backlog des Listeners beim Schließen; Abschluss der Antwort nach `Schliessen`; ungesendeter Body auf der Verbindung eines 503; Wirkung von `ReadHeaderTimeout` und `IdleTimeout`; gleichzeitiges `Schliessen`; Rand der Grenze von `MaxHeaderBytes` (rund 4 KiB); Messwerte der Standardbibliothek — **Ausgang:** entschieden (Architect), akzeptiertes Negativ.
- Annahmen über `net/http` (`slice-http-adapter` Lerneintrag 1): gemessen vom Architect im Image sind `Serve` nach `Close` des Listeners (kein `ErrServerClosed`), `OPTIONS *` mit und ohne das Mittel, 503 mit `Connection: close` bei 1 KiB, 1 MiB, `HEAD`, `OPTIONS *`, `CONNECT`, chunked und `Expect`, `Shutdown` mit abgebrochenem Kontext, `http.Server.Close` und `ctx` (endet nicht bei ungelesenem Body), der Wiederverwendungsrand bei 262144 und 262145 Bytes, der Puffer der Kopfgrenze und die Panik-Meldung mit und ohne `ErrorLog`. Nicht gemessen: das Verhalten unter `-race` und die Rate der Mutationen — **Ausgang:** offen; Adresse Implementer (Rate je Mutation, §3).
- `AGENTS.md` §3.14 nennt andere Faktoren als `SPEC-128` Nr. 1; dieser Plan rechnet nach `SPEC-128` (Rang 2 vor Rang 8) — **Ausgang:** weiter offen; die Rückgabe an den User aus `slice-http-adapter` steht dort.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster *Operation × Argumentart* für diesen Slice ist die Tabelle unten; jede Zelle nennt ihren Ort in `SPEC-136` (Nummer samt Zusatz) oder „gilt nicht“ mit Grund — **Ausgang:** entschieden (Architect), Zeilen unten.

| Operation · Argument | fehlend | leer | fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `Serve` · Zustand des `Listener` | gilt nicht: `New` weist nil ab (Nr. 1); Typed-nil: *Ungeprüft* | gilt nicht: kein Wert | Nr. 12: von außen geschlossen oder geschlossen übergeben ergibt einen Fehler (L7 Element 5, 6) | Nr. 12 und Zusatz: `Anhalten` und `Schliessen`, auch davor, beenden mit `nil` (L7 Element 1 bis 4) | *Ungeprüft* (Nr. 12): ein zweites `Serve` |
| `Anhalten` · Zustand | gilt nicht: kein Argument | gilt nicht: kein Argument | Nr. 13b und Zusatz: Request danach 503, jede Form (L4) | gilt nicht: `Anhalten` hat keinen Kontext und keine Frist | Nr. 13d und Zusatz: wirkungslos, auch gleichzeitig, Listener nach jeder Rückkehr zu (L6) |
| `Schliessen` · `frist` | gilt nicht: eine Zahl hat keinen fehlenden Zustand; `0` ist „nicht warten“ (Nr. 14) | Nr. 14: `0` (L8 Gegenseite, L11) | Nr. 14: negativ ergibt `model.ErrUngueltig` ohne Änderung (L8); sehr groß (`math.MaxInt64`): Nr. 14 „wartet höchstens `frist`“, keine Obergrenze (L11) | Nr. 14: Frist abgelaufen schließt die Verbindungen, endet `ctx`, wartet auf die Handler (L10, L20) | Nr. 14: das zweite ergibt `nil` (L15); gleichzeitig: *Ungeprüft* |
| Handler · Request-Form bei angehaltenem Server | gilt nicht: der Server erzeugt immer einen Request | Nr. 13b und Zusatz: ohne Body und `HEAD` (L4 Element 5) | Nr. 13b und Zusatz: `OPTIONS *`, `CONNECT`, jede Methode (Nr. 2; L4 Element 6, 7) | Nr. 13b und Zusatz: der Client sendet den Body weiter oder nicht fertig, die Antwort kommt trotzdem (L4 Element 3, 4); der Client trennt: gilt nicht, ein Schreibfehler auf der Fehlerantwort nimmt denselben Pfad wie jede Fehlerantwort (`schreibeFehlerobjekt`, Nr. 6e gilt für die Antwort, kein eigener Schreibpfad) | Nr. 13b und Zusatz: ein weiterer Request auf der Verbindung kommt nicht an (`Connection: close`, EOF; L4) |
| Handler · Body-Zustand bei `Schliessen` | gilt nicht: ein Request ohne Body hat `Body` nil (Nr. 4e); `ctx` endet (L10) | gilt nicht: leerer Body und kein Body sind für den Handler gleich (Nr. 4e) | Nr. 4g und 14 Zusatz: ungelesener Body, `ctx` endet mit der Frist (L20) | Nr. 14: Frist abgelaufen (L10, L20) | gilt nicht: eine Frist läuft einmal ab |

Zusatztabelle, Paare von Verstößen am selben Schlüssel:

| Paar | Ergebnis | Ort | Fänger |
|---|---|---|---|
| angehalten und Fehler des Dienstes | angehalten gewinnt, `Verarbeiten` nicht gerufen | Nr. 3b | L4 (Zähler von `Verarbeiten`) |
| angehalten und ungültiger Kopf (zu groß) | `net/http` weist zuerst ab (431), keine 503 | Nr. 4i | L2 |
| angehalten und Dienst liefert `ODR-E5009` (Root ruft `CloseAll` vor `Anhalten`) | 503 `ODR-E5010` vom Adapter, solange angehalten; sonst `ODR-E5009` | Nr. 15 (Pflicht des Roots) | `slice-bootstrap-m2` (Fänger für den Root, dort angenommen) |
| `Schliessen` mit Frist und zugelassener Handler, der den Body nicht liest | `ctx` endet, `ErrFristAbgelaufen` | Nr. 4g, 14 Zusatz | L20 |
| `Serve` schlägt fehl und `Anhalten` folgt | der Fehler des Accept vor dem Aufruf bleibt der Fehler | Nr. 12 Zusatz | L7 Element 5 |

Fenster der Sperre: siehe §3, Tabelle *Sperren und Fenster*.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (8×, verkörpert in `AGENTS.md` §3.10; W1 bis W5 sind die Fenster der Sperre, die der Detektor nicht meldet), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (3×, offen; die Rate je Mutation von L16 und die Zeitgrenzen von L12 werden unter Last gemessen), `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9; der neue Code verändert mehrere Zählstellen); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
