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

**Bezug:** `LH-FA-02`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0007](../../adr/0007-race-detektor-im-gate.md)

**Berührte Spec-Stellen:** `LH-FA-02.c` · `SPEC-059` · `SPEC-109` Nr. 2 · `SPEC-136` Nr. 11 bis 15 · `SPEC-137` · `SPEC-230` Nr. 8 · `SPEC-124` Nr. 14.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der `Server` des Kerns (`slice-http-adapter`) schließt in zwei Schritten (`Anhalten`, `Schliessen(frist)`), weist neue Requests auf bestehenden Verbindungen mit 503 `ODR-E5010` ab, trägt feste Grenzen (`SPEC-136` Nr. 11), und `ODR-E5010` steht in der Code-Tabelle (35 Codes). Der Schnitt gegen den Kern ist vor dem Code gemacht (`AGENTS.md` §3.14): der Lebenszyklus ist der einzige Teil ohne geteilten Mechanismus mit ihm; er braucht vom Kern nur `Server` und den Eintritt des Handlers (eine Zeile `if angehalten`) und seine eigene Sperre.

**Übernimmt:** die Sendung aus `slice-session-beenden` („Fänger der beiden Schritte“, Zeilen L3 bis L12 und L16 der Tabelle in §3) und das Risiko „Request-Größe und Zeitgrenzen“ aus `slice-http-adapter` (Fänger L1, L2); `slice-http-adapter` führt beide als übergeben (§1 und §6 dort, im selben Commit nachgezogen).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Reihenfolge des Herunterfahrens im Composition Root (`Anhalten` → `CloseAll` → `Schliessen(rest)` → Repositories) — `slice-bootstrap-m2` (dort in §1 und DoD *Herunterfahren* eingetragen); der Adapter liefert die zwei Schritte, nicht die Reihenfolge.
- TLS des Listeners — `slice-listener-tls`, Bestand der Welle [welle-betrieb](../welle-betrieb.md).
- Spans und Zähler je Fehlercode — `slice-telemetry-einbau-gateway`; der Kern liefert `ODR-E5010` nicht aus, die Zählstellen des neuen Codes (§3) liegen bei diesem Slice.

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

- [ ] Lebenszyklus: `Anhalten` und `Schliessen(frist)` in zwei getrennten Schritten, neue Requests auf bestehenden Verbindungen mit 503 `ODR-E5010` und `Connection: close`, `Serve` endet mit `nil`, die Zulassung ist ein Abschnitt (Zeilen L3 bis L16 der Tabelle in §3; `LH-FA-02.c`, `SPEC-059`, `SPEC-124` Nr. 14).
- [ ] Feste Grenzen des Servers (`ReadHeaderTimeout`, `IdleTimeout`, `MaxHeaderBytes`, kein `ReadTimeout` und kein `WriteTimeout`; Zeilen L1, L2).
- [ ] Der Code `ODR-E5010` in der Tabelle (503, Ziel `service`, 35 Codes) samt allen Zählstellen aus §3 (Zeile L17; `LH-QA-05.a`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-109` Nr. 2, `SPEC-137`, `SPEC-230` Nr. 8, Historie).
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
| `src/adapters/driving/httpodata/` | update | Zustand der Zulassung, `Anhalten`, `Schliessen`, 503-Pfad, Grenzen des Servers |
| `src/hexagon/model/fehler.go` | update | `CodeE5010` nach `CodeE5009`, Zeile in `codeZeilen()` |
| `spec/spezifikation.md` | update | `SPEC-136` Nr. 11 bis 15 und `SPEC-137` (Vorabentwurf des Architect), `SPEC-109` Nr. 2, `SPEC-230` Nr. 8, Historie |
| Tests des Pakets und der Zählstellen | neu / update | die Tabelle unten; Zählstellen siehe die Liste darunter |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vorabentwurf des Architect (Vertrag `SPEC-136`) vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| L1 | Voreinstellungen des Servers: `ReadHeaderTimeout` 10 s, `IdleTimeout` 120 s, `MaxHeaderBytes` 65536, `ReadTimeout` und `WriteTimeout` 0, Nr. 11 | `TestServerVoreinstellungen` (im Paket, liest das Feld des `http.Server`): 5 Felder, je Wert | `ReadTimeout` gesetzt (30 s): rot; `MaxHeaderBytes` auf den Standard der Bibliothek (1 MiB): rot | andere |
| L2 | Ein Kopf über der Grenze ergibt 431 von `net/http`, ein Kopf darunter geht durch, Nr. 4i, 11 | `TestKopfGrenze`: 2 Fälle (Kopf 60 KiB geht durch und der Fake sieht ihn; Kopf 100 KiB ergibt 431 Klartext, `Verarbeiten`- und `IDs`-Zähler 0) | `MaxHeaderBytes` nicht gesetzt (Standard 1 MiB): Fall 2 liefert 200, rot | andere |
| L3 | `Anhalten` schließt den Listener und kehrt ohne Warten zurück, auch wenn ein Request läuft, Nr. 13a, 13d | `TestAnhaltenSchliesstListener`: ein Request hängt im Fake (Kanal); `Anhalten` kehrt in 2 s zurück (Kanal, kein Schlafen); danach scheitert `net.Dial` auf die Adresse; der hängende Request ist noch offen und der Fake wird danach freigegeben | `Anhalten` ruft `http.Server.Shutdown` (wartet): kehrt nie zurück, rot per Frist; Listener nicht geschlossen: `Dial` gelingt | nebenläufig |
| L4 | Nach `Anhalten` ergibt ein Request auf einer bestehenden Verbindung 503 `ODR-E5010` mit `Connection: close`, ohne `Verarbeiten`, ohne Lesen des Bodys, mit `X-Correlation-Id`, Nr. 3b, 13b | `TestAngehaltenDreiNullDrei`: Menge 3 Methoden (GET, POST mit 1 KiB Body, HEAD) auf einer Keep-Alive-Verbindung, die vor `Anhalten` einen Request erledigte; erwartet 503, Fehlerobjekt `ODR-E5010` (HEAD ohne Body), `Connection: close`, der Verbindung folgt EOF; `Verarbeiten`-Zähler bleibt bei 1, `IDs`-Zähler wächst um 1 je Request | Status 500: rot; `Connection: close` fehlt: der Client sieht die Verbindung weiter offen, rot; `Verarbeiten` doch gerufen: Zähler 2 | andere |
| L5 | Ein vor `Anhalten` zugelassener Request läuft zu Ende, vollständig, `Close` einmal, Nr. 13c | `TestLaufenderRequestEndet`: GET hängt im Fake; `Anhalten`; Fake freigeben; der Client erhält 200 mit allem Body (1 MiB), `Close` 1 | Handler bricht beim Anhalten ab: der Client erhält einen Fehler, rot | nebenläufig |
| L6 | Ein weiteres und ein gleichzeitiges `Anhalten` sind wirkungslos, Nr. 13d | `TestAnhaltenWiederholt`: 8 Goroutinen rufen `Anhalten` gleichzeitig, danach nochmals; keine Panik, kein Fehler, ein laufender Request bleibt unberührt | zweites `Anhalten` schließt den Listener doppelt und gerät in Panik (`close` eines geschlossenen Kanals): rot | nebenläufig |
| L7 | `Serve` endet nach `Anhalten` und nach `Schliessen` mit `nil`, Nr. 12 | `TestServeEndetMitNil`: 2 Fälle (nach `Anhalten`, nach `Schliessen(0)`); `Serve` liefert `nil` | `Serve` gibt `http.ErrServerClosed` zurück: rot | andere |
| L8 | `Schliessen(frist < 0)` ergibt `model.ErrUngueltig` und ändert nichts, Nr. 14 | `TestSchliessenNegativ`: Frist -1 ns und `math.MinInt64`: `errors.Is(err, model.ErrUngueltig)`; danach nimmt der Server weiter Requests an (200) | negative Frist wie 0: der Server ist danach zu, rot | andere |
| L9 | `Schliessen` wartet höchstens `frist` und kehrt früher zurück, wenn alle Requests fertig sind, Nr. 14 | `TestSchliessenFrueh`: Request hängt im Fake; `Schliessen(30 s)` in der Hauptgoroutine; eine zweite Goroutine gibt den Fake nach 100 ms frei; Rückgabe `nil` nach weniger als 5 s und der Client erhielt 200 mit vollem Body | sofortiges Schließen der Verbindungen: der Client erhält einen Fehler, rot; Warten auf die volle Frist: Dauer über 5 s, rot | nebenläufig |
| L10 | Läuft die Frist ab, schließt `Schliessen` die Verbindungen, `ctx` des Requests endet, es wartet auf die Rückkehr des Handlers (`Close` gerufen) und liefert `ErrFristAbgelaufen`, Nr. 14 | `TestSchliessenFristAbgelaufen`: Request hängt im Fake bis `ctx.Done()`, dann liefert der Fake `Close` mit 150 ms Verzögerung; `Schliessen(200 ms)`: Rückgabe `ErrFristAbgelaufen`, nach dem Aufruf ist `Close` gerufen (Zähler 1), `ctx.Err() == context.Canceled` | Rückkehr vor dem Handler: `Close`-Zähler 0 bei der Rückkehr, rot; `nil` statt `ErrFristAbgelaufen`: rot | nebenläufig |
| L11 | `Schliessen(0)` wartet nicht, Nr. 14 | `TestSchliessenNull`: Request hängt im Fake bis `ctx.Done()`; `Schliessen(0)` kehrt in 5 s zurück (nach dem Ende des Handlers) mit `ErrFristAbgelaufen`; ohne laufenden Request `nil` | Frist 0 als „ohne Ende“: kehrt nie zurück, rot per Frist | nebenläufig |
| L12 | Die Dauer ist eine Frist, nicht die Summe: `Anhalten` plus `Schliessen(F)` dauern höchstens `F` plus die Zeit zum Beenden, Nr. 13, 14, 15 | `TestSummeDerFristen`: Request hängt bis `ctx.Done()`; Zeit t0 vor `Anhalten`, `Anhalten`, `Schliessen(300 ms)`; Dauer zwischen 300 ms und 1,5 s | `Anhalten` wartet 300 ms zusätzlich: Dauer über 600 ms, rot (Grenze 600 ms im Test); `Schliessen` wartet zweimal: rot | nebenläufig |
| L13 | `Schliessen` schließt Keep-Alive-Verbindungen ohne laufenden Request sofort, `Anhalten` lässt sie offen, Nr. 13b, 14 | `TestIdleVerbindung`: eine Verbindung nach einem Request im Leerlauf; nach `Anhalten` sendet der Client einen Request und erhält 503 (die Verbindung lebte); danach `Schliessen(1 s)`: der Client erhält EOF | Idle-Verbindung bleibt: der Client erhält keine EOF, rot | andere |
| L14 | `Schliessen` ohne vorheriges `Anhalten` hält an, Nr. 14 | `TestSchliessenOhneAnhalten`: `Schliessen(1 s)` ohne `Anhalten`: `Dial` scheitert danach, `Serve` kehrt mit `nil` zurück | `Schliessen` nur wartend: `Dial` gelingt, rot | andere |
| L15 | Ein weiteres `Schliessen` liefert `nil` und wartet nicht, Nr. 14 | `TestSchliessenZweimal`: zweimal nacheinander; das zweite in unter 1 s mit `nil` | zweites `Schliessen` panikt oder wartet: rot | andere |
| L16 | Nach der Rückkehr von `Schliessen` läuft kein Handler mehr und keiner beginnt; Prüfung und Zählen der Zulassung sind ein Abschnitt, Nr. 13c, 14 | `TestZulassungAtomar` (Wirkungs-Fänger): 200 Requests gleichzeitig (Goroutinen) gegen `Anhalten` und `Schliessen(2 s)` aus der Hauptgoroutine nach 20 ms; nach der Rückkehr: der Zähler der laufenden Handler im Fake ist 0, der Zähler der Aufrufe von `Verarbeiten` wächst in den nächsten 200 ms nicht, und jeder Request endete mit 200 (zugelassen und beendet) oder mit 503/Verbindungsfehler; mit `-count=50` | Prüfung auf angehalten (`atomic.Bool`) und Zählen getrennt (zwei Abschnitte): ein Handler beginnt nach der Rückkehr; der Fänger wird nur mit Wahrscheinlichkeit rot, die Rate je Mutation (Zahl der Läufe, Messender Implementer) steht im Plan; der Detektor meldet es nicht (alles unter Sperre bzw. atomar) | nebenläufig |
| L17 | `ODR-E5010` steht in der Tabelle (503, `service`, message „Service is shutting down“), 35 Codes, Nr. 13b | Tests, die der Slice ändert: `model/fehler_test.go` (Zeile für `ODR-E5010`, die `fremd`-Liste `{200, 418, 503}` wird `{200, 418, 507}`), `TestMessageFest` (die message ist eindeutig und ohne `%{}<>`), `adapters/driven/telemetry/telemetry_test.go` und `regeln_test.go` (34 auf 35, 38 auf 39, `alleCodes` um `ODR-E5010`) | Code fehlt in der Tabelle: `AlleCodes` hat 34, rot; Status 502 statt 503: rot | andere |

Zeilen L1 bis L17: **17**, davon **8 nebenläufig** (L3, L5, L6, L9, L10, L11, L12, L16).

**Raten (`AGENTS.md` §3.10).** Jede nebenläufige Zeile ordnet mit Kanälen. Der Implementer fährt die Mutation je nebenläufiger Zeile mit `-count=50` in einem Prozess und nennt „x von 50 rot“, für ein Datenrennen die Zahl roter Prozesse mit `-test.count=1` (mindestens 20 Prozesse), den Messenden und trägt beides in einen Absatz von §3 ein; „hoch“ steht nie ohne Zahl. L16 ist das einzige Fenster der Sperre (zwischen Prüfen auf angehalten und Zählen): es hat keinen Haltepunkt an einem Argument des Tests (`ctx`, Strom, Fake), der Detektor sieht es nicht, der Fänger L16 ist ein Wirkungs-Fänger mit Rate; fällt die Rate unter 50 von 50, steht die Zeile als Negativ mit der Zahl im Plan (Folge: ein Handler beginnt nach der Rückkehr von `Schliessen`, `SPEC-136` *Ungeprüft*).

**Zählstellen des neuen Codes `ODR-E5010`** (Tabelle 34 auf 35 Codes; im Commit dieses Slice, nicht des Kerns, denn der Kern liefert den Code nicht aus):

- `spec/spezifikation.md`: `SPEC-109` Nr. 2 (Zeile und der Satz zur Zahl), §4 (`SPEC-137`), `SPEC-230` Nr. 8 (35), Historie, Zeiger. `SPEC-109` Nr. 5 unverändert.
- `src/hexagon/model/fehler.go`: `CodeE5010` nach `CodeE5009` (die folgenden Konstanten verschieben sich um 1; das Gate fängt es, keine Persistenz speichert `Code`-Zahlen), Zeile in `codeZeilen()`: `{CodeE5010, "ODR-E5010", []int{503}, ZielService, "Service is shutting down"}`.
- `src/hexagon/model/fehler_test.go`: Zeile `"ODR-E5010": {[]int{503}, "service"}`; die Liste `fremd` `{200, 418, 503}` wird `{200, 418, 507}` (sonst schlägt `TestCodeTabelle` fehl, weil `ODR-E5010` 503 trägt).
- `src/adapters/driven/telemetry/telemetry_test.go` (die 34 in den Zeilen 44 bis 49, 162 bis 164, 206 bis 207, 382, 413 mit `len(erwartet) != 35` und `len(codes) != 39`, 480 bis 496) und `regeln_test.go` (Zeilen 52 bis 58, 391 bis 400: `alleCodes` um `ODR-E5010`, 34 auf 35); `telemetry.go` folgt `model.AlleCodes()` und bleibt.
- `src/hexagon/ports/inbound/fehler_test.go`: keine Änderung (prüft `ODR-E5009`).
- Der Implementer sucht den Rest mit `grep -rn 'AlleCodes\|\b34\b\|\b33\b' src` und nennt jeden Fund im Bericht.
- Pläne in `open/`, die 34 Codes zählen (vom Planner im selben Commit wie dieser Plan nachgezogen): `slice-handbuch-meldungskatalog`, `slice-handbuch-record-replay`, `slice-telemetry-spans`. `harness/README.md` nennt die Zahl der Codes nicht.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-http-adapter` liegt in `done/`. Die Spezifikation trägt `SPEC-136` Nr. 11 bis 15 und `SPEC-137` (Vorabentwurf des Architect), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Reihenfolge der Welle: `slice-http-adapter` → `slice-http-lebenszyklus` → `slice-bootstrap-m2`.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 2100 Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/adapters/driving/httpodata/`, `src/hexagon/model/fehler.go` und die Tests der Zählstellen, Tests eingeschlossen; Basis ist der letzte Commit des Architect). Schätzung S = 1540 (a=8, b=9, f=0, p=140 nach der Tabelle des Architect; `model/fehler.go` 3 Zeilen), Marke 100 · aufrunden(135·1540/10000) = 2100. Den Block nach `SPEC-128` Nr. 3 setzt der Planner beim Start.
- `in-progress` → `open` (blockiert): der Kern liefert `Server` und den Eintritt des Handlers nicht in der Form, die der Vorabentwurf voraussetzt (Rückgabe an den Architect, `AGENTS.md` §3.12).

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

- Das Fenster der Sperre zwischen Prüfen auf angehalten und Zählen (L16) hat keinen Haltepunkt an einem Argument des Tests, und der Detektor von `make test-race` meldet es nicht (alles unter Sperre bzw. atomar); der Fänger ist ein Wirkungs-Fänger mit Rate je Mutation — **Ausgang:** offen; Adresse Implementer (Zahl der Läufe und Messender im Plan); fällt die Rate unter 50 von 50, steht die Zeile als Negativ mit der Zahl und der Folge „ein Handler beginnt nach der Rückkehr von `Schliessen`“.
- Der neue Code `ODR-E5010` verändert alle Zählstellen (Tabelle 34 auf 35, Zähler 35, `fremd`-Liste in `fehler_test.go`); die Pläne `slice-handbuch-meldungskatalog`, `slice-handbuch-record-replay` und `slice-telemetry-spans` tragen die Zahl und sind im selben Commit wie dieser Plan nachgezogen — **Ausgang:** offen; Adresse Implementer, sucht den Rest mit `grep` und nennt jeden Fund im Bericht.
- Die Reihenfolge `Anhalten` → `CloseAll` → `Schliessen(rest)` hängt am Composition Root: ruft er `CloseAll` vor `Anhalten`, erhalten neue Requests `ODR-E5009` statt 503, wartet er mit der vollen Frist, dauert es die Summe — **Ausgang:** weiter offen; Adresse `slice-bootstrap-m2` (Sendung dort in §1 und DoD angenommen, Fänger für den Root dort).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster *Operation × Argumentart* des Vorabentwurfs für diesen Slice ist die Tabelle unten; jede Zelle nennt ihren Ort in `SPEC-136` oder „gilt nicht“ mit Grund — **Ausgang:** entschieden (Architect), Zeilen unten.

| Operation · Argument | fehlend | leer | fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `Serve` · Zustand des `Listener` | gilt nicht: `New` weist nil ab (Nr. 1) | gilt nicht: kein Wert | Nr. 12: von außen geschlossen ergibt einen Fehler | Nr. 12, 13, 14: `Anhalten` und `Schliessen` beenden mit `nil` | Nr. 12 *Ungeprüft*: ein zweites `Serve` |
| `Anhalten` · Zustand | gilt nicht: kein Argument | gilt nicht: kein Argument | Nr. 13b: Request danach 503 | gilt nicht: `Anhalten` hat keinen Kontext und keine Frist | Nr. 13d: wirkungslos, auch gleichzeitig |
| `Schliessen` · `frist` | gilt nicht: eine Zahl hat keinen fehlenden Zustand; `0` ist „nicht warten“ (Nr. 14) | Nr. 14: `0` | Nr. 14: negativ ergibt `model.ErrUngueltig` ohne Änderung | Nr. 14: Frist abgelaufen schließt die Verbindungen | Nr. 14: das zweite ergibt `nil` |

Zusatztabelle, Paare von Verstößen am selben Schlüssel:

| Paar | Ergebnis | Ort | Fänger |
|---|---|---|---|
| angehalten und Fehler des Dienstes | angehalten gewinnt, `Verarbeiten` nicht gerufen | Nr. 3b | L4 |
| angehalten und ungültiger Kopf (zu groß) | `net/http` weist zuerst ab (431), keine 503 | Nr. 4i | L2 |

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (8×, verkörpert in `AGENTS.md` §3.10; L16 ist das Fenster der Sperre, das der Detektor nicht meldet), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (3×, offen; die Rate je Mutation von L16 wird unter Last gemessen), `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9; der neue Code verändert mehrere Zählstellen); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
