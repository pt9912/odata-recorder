# Slice slice-hybrid-service: HybridUseCase, Teil 1 (Mock zuerst und Replay-Fallback nach ODR-Code)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, `LH-FA-04`, `LH-FA-11`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` (Entscheidungsmatrix) · `LH-FA-04.c` (Adressierung, `X-Recorder-*`-Header) · `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080` · `SPEC-020`, `SPEC-021`, `SPEC-024` · `SPEC-120` (`UpstreamClient`, Fehler) · `SPEC-092` (`LINK_REWRITE`) · `ARC-002`, `ARC-003` · `architecture.md` §2.2, §2.5, §4.6 · `SPEC-261`, `SPEC-284` (Fallback-Policy und Vertrag des Dienstes) · `SPEC-135`, `SPEC-195` (Replay und Mock) · `LH-FA-04.b`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-hybrid-service`, Worktree `/Development/odata-recorder-wt-hybrid-service`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Schnitt (Planner, 2026-11-08, vor dem Code):** der Plan ist in zwei Teile geschnitten (§4). **Dieser Slice ist Teil 1**: Mock zuerst und Replay-Fallback. Der Upstream-Fallback (Stufe U, Header-Filter, Kennzeichnung `upstream`, Fehlerabbildung, Referenzdienst) ist Teil 2, [slice-hybrid-service-upstream](../open/slice-hybrid-service-upstream.md); es führt die Zeilen Z07, Z19 bis Z25, Z27, Z35 und je eine Upstream-Seitenzeile zu Z01, Z08, Z09, Z10, Z11, Z26 und Z28.

**Ziel:** `HybridUseCase` (Inbound Port in `src/hexagon/ports/inbound`, Dienst in `src/hexagon/services/hybrid`) ruft `MockUseCase` über dessen Inbound Port und entscheidet bei einem Fehler des Mock nach dem `ODR`-Code (nie nach dem HTTP-Status) mit der Policy aus `slice-fallback-policy`: Antwort oder Fehler bleiben (Erfolg, fachliche `404`, `409`, `412`, ungültige Anfrage), `ReplayUseCase` wird nur für die freigegebenen Kategorien gerufen; ein Replay-Fehler, auch `ODR-E5001` und ein erschöpfter Key (`ODR-E5002`), wird in diesem Teil unverändert ausgeliefert (`LH-FA-12.b`); die Stufe `Upstream` fügt Teil 2 hinzu, bis dahin gibt es keinen Upstream-Zugriff, der Zwischenstand ist der sichere. Der Body einer fallback-fähigen Anfrage wird gepuffert. Den Vertrag des Dienstes trägt `SPEC-284`.

**Übernimmt:** der Replay-Fallback einer mutierenden Methode ändert den Zustand des Mock nicht (Sendung von `slice-fallback-policy-entscheidung`, dessen §1 *Ausdrücklich NICHT* diesen Slice nennt; der Client sieht eine aufgezeichnete Erfolgsantwort, der Mock bleibt unverändert; Fänger Z29). Die Pflichten des Aufrufers aus `SPEC-261` Nr. 10 a bis c, die das Senden an den Upstream betreffen (dieselbe Zeichenkette als `Ziel`; `Entscheide` vor jedem Senden; die drei Header zum Überschreiben der Methode nie; Fänger: der Referenzdienst empfängt unter dem Host, den die Policy freigab, Sendung von `slice-fallback-policy`), und das Risiko der Upstream-Links einer Fallback-Antwort aus `welle-hybrid` §6 (`LINK_REWRITE` ist nach `SPEC-092` nur `off`) gehören zu Teil 2; er trägt sie in §1 und DoD mit den Kennungen der Geber (`AGENTS.md` §3.13).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Policy (Hauptschalter, Allowlist, Routenfreigaben, Kategorien) und ihre Syntax — `slice-fallback-policy` (Prüfung der Einträge und `New`) und `slice-fallback-policy-entscheidung` (`Entscheide`); der Dienst ruft sie und entscheidet nichts selbst.
- Der Modus `hybrid` im Composition Root, die Konfiguration und der Listener — `slice-bootstrap-m4`; der Dienst läuft hier nur gegen Fakes der Ports.
- Der Upstream-Fallback — [slice-hybrid-service-upstream](../open/slice-hybrid-service-upstream.md) (Teil 2): `Konfig.Upstream` und `Dienst.UpstreamURL`, der Zweig `Upstream`, `UpstreamAnfrage` mit Filter der Header, Kennzeichnung `upstream`, Fehlerabbildung von `Senden`, Referenzdienst und Fake von `UpstreamClient`; Teil 2 schreibt die Pufferung und den Ablauf dieses Teils nicht um (Ausnahme: §6, Rückgabe an den Architect).
- Speichern einer Fallback-Antwort und Datenabgleich zwischen Mock und Upstream — Out-of-Scope von `LH-FA-12`.
- Umschreiben absoluter Upstream-Links (`LINK_REWRITE`) — `SPEC-092` kennt nur `off`; das Risiko trägt Teil 2.
- Spans, Audit und Metriken des Fallbacks — `welle-betrieb`; es gibt noch keinen Port `Telemetry`.

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

- [ ] Mock zuerst: Erfolg und jeder Fehler ohne Fallback-Auslöser (fachliche `404`, `409`, `412`, ungültige Anfrage, Fehler ohne `ODR`-Code) gehen unverändert an den Aufrufer; die Zuordnung (`X-Recorder-*`) geht unverändert an Mock und Replay.
- [ ] Replay-Fallback nur für die Kategorien der Policy (`ODR-E6001` und `ODR-E6002` als `unsupported`, `ODR-E6003` als `coverage_insufficient`): ein Treffer wird ausgeliefert mit `X-Recorder-Fallback: replay`, jeder Replay-Fehler (Liste in `LH-FA-12.b`, auch `ODR-E5001`) unverändert; ein erschöpfter Key geht an keinen Upstream; ein Replay-Fallback einer mutierenden Methode lässt den Zustand des Mock unverändert (`SPEC-261` Nr. 10, Sendung von `slice-fallback-policy-entscheidung`).
- [ ] Body des Requests gepuffert nach `SPEC-284` Nr. 5 (nur fallback-fähige Anfragen, kein Fallback über `MAX_CAPTURE_BODY`, Body der Anfrage nie geschlossen), eine verworfene Antwort von Mock oder Replay genau einmal geschlossen und eine ausgelieferte nie (`SPEC-284` Nr. 10).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Modus `hybrid` erreicht es erst mit `slice-bootstrap-m4`).
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
| `src/hexagon/ports/inbound/` | update | `HybridUseCase` |
| `src/hexagon/services/hybrid/` | update | Dienst neben der Policy aus `slice-fallback-policy` und `slice-fallback-policy-entscheidung` |
| `spec/spezifikation.md` | update (Architect) | Vertrag des Dienstes samt Randformen; Nachzug `LH-FA-12.b`, `architecture.md` §2.2 und §4.6 |
| Tests | neu | Matrix `LH-FA-12.b` × Zustände der Policy (Stufen M und R); die Fakes von `MockUseCase` und `ReplayUseCase` halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service; Fake von `UpstreamClient`, Referenzdienst (Verbindungszähler) und Prüf-Client gehören zu Teil 2 |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15, Spalte *Klasse* nach `SPEC-128` Nr. 3 a) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstanden mit dem Vertrag `SPEC-284` vor dem ersten Code-Commit (Architect, 2026-11-08); die Randformen sind in der Spezifikation entschieden, nicht im Code (`AGENTS.md` §3.12). **Die Tabelle unten ist die von Teil 1**: 26 Zeilen, die Texte der Zeilen nennen nur die Stufen M und R. Die Zeilen Z07, Z19 bis Z25, Z27 und Z35 und die Upstream-Seite von Z01, Z08, Z09, Z10, Z11, Z26 und Z28 (Wortlaut der Architect-Fassung, in Teil 2 ausgeschrieben) führt [slice-hybrid-service-upstream](../open/slice-hybrid-service-upstream.md); im Raster und in der Zusatztabelle in §6 gehören die Zeile *Upstream-Fallback* und die Spalte U zu Teil 2. Der Ort steht in der Zelle *Nr.* (`SPEC-284 Nr. <n>`). Die Fakes von `MockUseCase` und `ReplayUseCase` halten jedes Argument fest (`ctx`, `Anfrage`, `Header`-Map, Body-Leser), zählen jede Operation und tragen einen Zählstrom (Zähler für `Read` und `Close`). Alle Fänger, die über eine Menge laufen, laufen als Schleife über die Menge mit einer Zählprüfung gegen die Quelle der Menge (`model.AlleCodes()`, die Namen aus `SPEC-121` Nr. 4); Orakel ist eine unabhängige Fassung der Regel aus `SPEC-261` Nr. 4, nicht der Code.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z01 · SPEC-284 Nr. 1 | `New` weist jeden Fehler der Konfig dieses Teils ab (kein Dienst, leerer `Root`, nil in `Mock`, `Replay`, `Policy`, `MaxBytes` 0 und -1) und nimmt `MaxBytes` 1 an; es speichert `Dienste` nicht, denn Teil 1 liest die Liste nach der Prüfung nie (§6, Abweichung Dienstwahl; `Upstream` und `UpstreamURL`: Z01 (U) in Teil 2) | Tabellentest über die Menge der Felder, Gutfälle auf beiden Seiten der Grenze (`TestNew`) | `MaxBytes < 1` zu `< 0`; Prüfung von `Replay`, `Mock`, `Policy`, `Root` oder der Liste gestrichen | andere |
| Z02 · SPEC-284 Nr. 2 | `ctx` nil, `Methode` leer, `Laenge` -2 und `Body` nil bei `Laenge` 1 ergeben `model.ErrUngueltig` ohne Code, ohne Aufruf einer Stufe oder von `Entscheide`, ohne Lesen des Bodys; `Laenge` -1 und 0 sind gültig | Tabellentest; die Fakes zählen jede Operation (0), der Zählreader jedes `Read` (0) | `Laenge < -1` zu `<= -1`; Prüfung nach dem Aufruf von M | andere |
| Z03 · SPEC-284 Nr. 3 a, 3 b | Erfolg (Fehler nil, auch mit Status 404 oder 500 in der Antwort) geht unverändert und ohne Kennzeichnung aus; von jedem Code in `model.AlleCodes()` als Fehler von M lösen unter der offensten Policy genau `ODR-E6001`, `ODR-E6002`, `ODR-E6003` eine Stufe aus, jeder andere kehrt als derselbe Wert zurück (R 0); Codes gleichen Status, verschiedener Auslösung (`ODR-E6001` gegen `ODR-E6006`, beide 501) | Schleife über alle Codes mit Zählprüfung gegen `AlleCodes()`; Orakel: Auslöser nach `SPEC-261` Nr. 3 a; Identität mit `==` | `ODR-E6005`, `ODR-E6006` oder `ODR-E6008` als Auslöser; Entscheidung am Status; Fehler verpackt; Kennzeichnung auch bei Erfolg | andere |
| Z04 · SPEC-284 Nr. 3 b, 2 | Ein Fehler von M ohne Code (`errors.New`, `model.ErrUngueltig`, `*Fehler` mit Nullwert) oder mit Kontextfehler (`Canceled`, `DeadlineExceeded`, verpackt, im `errors.Join` neben `ODR-E6003`) kehrt unverändert zurück, R läuft nicht | Schleife über sieben Formen unter der offensten Policy | Kontextprüfung entfernt (der Join mit `ODR-E6003` löst R aus); `errors.As` ohne Prüfung des Nullwerts | andere |
| Z05 · SPEC-284 Nr. 3 c | `Entscheide` bekommt `Methode`, vollen `Pfad` und `Code`; die Entscheidung hängt an jedem der drei; `Ziel` ist in diesem Teil immer leer (es gibt kein Feld `UpstreamURL`; `SPEC-261` Nr. 5 c: ein leeres `Ziel` gibt die Policy nie frei; die Übergabe des `Ziel` fängt Z19 in Teil 2) | Schleife über drei Felder, je eine Policy, die nur an diesem Feld hängt (Methodenklasse, Route auf den vollen Pfad, Code); Orakel: Replay ja oder nein | `Pfad` gleich dem OData-Pfad (in Teil 1 kein Fänger: akzeptiertes Negativ, §6); `Methode` fest `GET`; `Code` fest `ODR-E6003` | andere |
| Z06 · SPEC-284 Nr. 3 c, 7 | Replay-Weg: `ODR-E6001` und `ODR-E6002` (Kategorie `unsupported`), `ODR-E6003` (`coverage_insufficient`) × Kategorie gesetzt oder nicht: gesetzt ruft R genau einmal, die Antwort von R geht mit Kennzeichnung `replay` und unverändertem Status und Body aus; nicht gesetzt kehrt der Fehler von M zurück | Tabellentest 3 × 2 und die Fehlzuordnung (`ODR-E6003` mit nur `unsupported`); Fake R hält `Anfrage` | Kategorie ignoriert; `ODR-E6002` ohne Replay; R ohne Freigabe gerufen | andere |
| Z08 · SPEC-284 Nr. 3 d | Replay-Miss: liefert R `ODR-E5001`, kehrt es in diesem Teil als derselbe Wert zurück, auch unter einer Policy, die jeden Host freigibt (das Ergebnis `Upstream` entsteht mit leerem `Ziel` nie; die Upstream-Seite: Z08 (U) in Teil 2), für jede Methode, M und R genau einmal | Schleife über `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `MERGE`, `OPTIONS`, `TRACE`, `get`, je mit einer Policy ohne und mit Host-Freigabe | Fehler verpackt; R ein zweites Mal gerufen; `Ziel` aus einer festen Zeichenkette statt leer | andere |
| Z09 · SPEC-284 Nr. 3 d | Jeder Replay-Fehler (jeder Code von `AlleCodes()` als Fehler von R, offenste Policy, `GET`) kehrt als derselbe Wert zurück, auch der erschöpfte Key `ODR-E5002` (die Upstream-Seite „U 0“: Z09 (U) in Teil 2) | Schleife über alle Codes mit Zählprüfung; Identität mit `==` | Fehler umgebaut; Replay-Fehler als `ODR-E5001` gedeutet | andere |
| Z10 · SPEC-284 Nr. 3 e | Abbruch zwischen den Stufen: ist `ctx` nach M beendet (Fake M beendet es und liefert `ODR-E6003`), läuft R nicht; das Ergebnis ist `ctx.Err()` ohne Code (nach R und vor U: Z10 (U) in Teil 2) | zwei Fälle: nach M, vor dem Aufruf | Prüfung vor R entfernt | andere |
| Z11 · SPEC-284 Nr. 2, 12 | Jede Aufrufstelle H2 und H4 erhält `ctx` unverändert (Wert und `Done` identisch), und ein Kontextfehler dort (auch verpackt) kehrt unverändert ohne Code zurück (H6: Z11 (U) in Teil 2) | Schleife über zwei Stellen × {Identität, Kontextfehler}; die Fakes halten `ctx` | `WithoutCancel` an R; `context.Background()` an M | andere |
| Z12 · SPEC-284 Nr. 5 a, 5 b | Ohne Pufferung bleibt der Body ungelesen und geht unverändert (Identität von `Body` und `Laenge`) an M: Anfrage nicht fallback-fähig (alle Schlüssel leer, nie freigegebene Methode, nur eine Host-Freigabe ohne Replay-Kategorie, denn die Probe ruft `Entscheide` mit leerem `Ziel`), ohne Body (nil, `Laenge` 0 mit Leser), `Laenge` größer `MaxBytes` | Schleife über die Formen; Zählreader (0 `Read`) | immer puffern; Probe nur mit `ODR-E6003` (Fähigkeit für einen Code); `Ziel` der Probe nicht leer | andere |
| Z13 · SPEC-284 Nr. 5 c | Pufferung: M und R erhalten je einen eigenen vollständigen Leser mit denselben Bytes und `Laenge` gleich der Länge, auch wenn eine Vorstufe nichts, ein Byte oder alles liest, bei deklarierter `Laenge` und bei -1 (der Leser für U: Z24 in Teil 2) | Schleife Stufen × Lesemenge × {deklariert, -1}; die Fake-Stufen vergleichen die Bytes mit dem Orakel | gemeinsamer Leser; Puffer nach M gekürzt; `Laenge` -1 weitergereicht | andere |
| Z14 · SPEC-284 Nr. 5 c | Grenze: genau `MaxBytes` Bytes (deklariert und -1) werden gepuffert und fallen zurück; `MaxBytes` + 1 (deklariert: `Laenge` größer; -1: gelesenes Byte `MaxBytes` + 1) gibt keinen Fallback, M erhält alle Bytes in Reihenfolge, sein Ergebnis kehrt zurück, R und `Entscheide` 0 | zwei mal zwei Fälle, dazu `MaxBytes` 1 | `>` zu `>=`; `LimitReader(MaxBytes)` statt `+1` (M sähe einen gekürzten Body); Rest nicht angehängt | andere |
| Z15 · SPEC-284 Nr. 5 c, 5 e | Lesefehler beim Puffern (nicht `io.EOF`, auch ein Kontextfehler, nach 0, 3 oder `MaxBytes` Bytes): kein Fallback, M erhält die gelesenen Bytes und dann den Fehler, sein Ergebnis kehrt unverändert zurück | Fehler-Leser mit k Bytes; Fake M liest und hält | Fehler verschluckt (gekürzter Body gilt als vollständig, Fallback mit gekürztem Body) | andere |
| Z16 · SPEC-284 Nr. 5 d | Der Dienst schließt den Body der Anfrage nie, in keinem Pfad (Erfolg, Replay, Fehler, Abbruch; der Pfad Upstream: Teil 2) | Zählreader in den Fällen von Z06 bis Z15 (`Close` 0) | `defer body.Close()` | andere |
| Z17 · SPEC-284 Nr. 6 | M und R erhalten dieselbe `Header`-Map mit allen Namen (auch `X-Recorder-*`), und die Map der Anfrage ist nach dem ganzen Aufruf unverändert (nach U: Z23 in Teil 2) | Fakes halten die Map; Tiefenkopie vor dem Aufruf gegen Map nach dem Aufruf | Filter löscht in der Original-Map; Kopie ohne `X-Recorder-*` an R | andere |
| Z18 · SPEC-284 Nr. 6, 3 d | `ODR-E5003` fehlender Session-Header entsteht im Replay-Fallback und kehrt unverändert zurück (U 0); ein Mock-Erfolg ohne Session-Header wird ausgeliefert, der Dienst prüft die Header nicht vor M | zwei Fälle mit Fakes: M Erfolg ohne die Header (M 1, R 0); M `ODR-E6003`, R `ODR-E5003` | Vorabprüfung der Header vor M | andere |
| Z26 · SPEC-284 Nr. 9 | Kennzeichnung: eine Antwort aus R trägt `X-Recorder-Fallback: replay`, aus M nie (`upstream`: Z26 (U) in Teil 2); ein Header gleichen Namens der Stufe (jede Schreibweise) wird ersetzt (genau ein Wert); die Map der Stufe bleibt unverändert | Schleife Stufen × Vorbelegung des Headers | angehängt statt ersetzt; in der Map der Stufe gesetzt; bei M gesetzt | andere |
| Z28 · SPEC-284 Nr. 10 | Liefert M oder R Fehler und Antwort mit Body zugleich, schließt der Dienst den Body genau einmal und liefert den Fehler (auch einen Kontextfehler) unverändert; ausgelieferte Antworten von M und R schließt er nie (U: Z28 (U) in Teil 2) | Schleife Stufen × Fehlerart am Zählstrom | `Close` fehlt; `Close` doppelt; `Close` bei Erfolg | andere |
| Z29 · SPEC-284 Nr. 7 | Mutierender Replay-Fallback: `POST`, `PUT`, `PATCH`, `DELETE`, `MERGE` mit `ODR-E6002` und Kategorie `unsupported`: M genau einmal und danach nie, R einmal, die aufgezeichnete Erfolgsantwort geht mit Kennzeichnung aus (Sendung von `slice-fallback-policy-entscheidung`, Mock-Zustand unverändert) | Schleife über fünf Methoden; Zustands-Fake von M (Schreibzugriffe 0, Aufrufe 1) | M nach R erneut gerufen; Wiederholung des Mock mit anderem Argument | andere |
| Z30 · SPEC-284 Nr. 13 | Wiederholung: derselbe Aufruf zweimal ruft alle Stufen erneut, kein Ergebnis wird gemerkt | Fakes zählen; Aufruf 2 mal in jedem Pfad von Z06 bis Z08 | Ergebnis je Pfad und Methode gemerkt | andere |
| Z31 · SPEC-284 Nr. 13 | Gleichzeitige Aufrufe (64, gemischt über Erfolg, Replay, Fehler, Puffer): jeder erhält seine eigene Antwort und seine eigenen Body-Bytes, die Summe der Aufrufe aller Fakes gleich der Erwartung, kein Datenrennen | Wirkungs-Fänger (Bytes je Aufruf eindeutig, Summen) und `make test-race` je Prozess (`-test.count=1`) | gemeinsamer Puffer oder Variable auf Paketebene (Rate je Lauf und je Prozess im Bericht) | nebenläufig |
| Z32 · SPEC-284 Nr. 13, 10 | Gleichzeitiger Abbruch (64 Aufrufe, die Hälfte mit beendetem `ctx` mitten in M oder R): jedes Ergebnis gehört zum eigenen Aufruf, `Close` der verworfenen Antworten gleich der Zahl der Abbrüche mit Antwort, nie doppelt | Wirkungs-Fänger mit Zählstrom, unter `make test-race` | Kontextfehler durch einen Code ersetzt; Fehler einer fremden Anfrage ausgeliefert (gemeinsame Variable); `Close` doppelt unter Abbruch | nebenläufig |
| Z33 · SPEC-284 *Ungeprüft* 1 | Der Speicher über alle gleichzeitigen Puffer: akzeptiertes Negativ. Grund: Prüfgrenze `MaxBytes` je Request, eine Gesamtgrenze hätte einen Zähler mit Sperre verlangt. Folge: bis *N* · `MaxBytes` bei *N* gleichzeitigen fallback-fähigen Requests mit Body | keiner | – | andere |
| Z34 · SPEC-284 *Ungeprüft* 2 | Der Zustand des echten Mock beim mutierenden Replay-Fallback: akzeptiertes Negativ. Grund: der Fake zählt Aufrufe, dass ein Fehler des Mock nichts schreibt, tragen die Fänger von `SPEC-192` Nr. 5 h. Folge: bräche der Mock sie, sähe kein Test des Dienstes es | keiner (Z29 zählt die Aufrufe des Fakes) | – | andere |
| Z36 · SPEC-284 *Ungeprüft* 4 | Eine Entscheidung der Policy außerhalb der drei Werte: akzeptiertes Negativ. Grund: der Typ der Entscheidung ist im Paket `policy` geschlossen. Folge: keine (`UpstreamAntwort.Body` nil und Änderung durch den Adapter, *Ungeprüft* 5 und 6: Z35 in Teil 2) | keiner | – | andere |

**Stand der Umsetzung Teil 1** (Implementer, 2026-11-08; `AGENTS.md` §3.9, §3.10): Port `HybridUseCase` in `src/hexagon/ports/inbound/`, Dienst `hybrid.Verteiler` in `src/hexagon/services/hybrid/` (Produktcode 234 Zeilen, Messung der Pfade aus dem Block 1989 Zeilen gegen die Marke 3900). Mutationslauf `tools/harness/mutationslauf.sh` mit 83 Mutationen in einem Container (`--memory=2g`, `--cpus=4`, ohne Netz), 75 s, Last des Hosts 4,3 bis 5,6 bei 4 CPUs, Messender der Lauf selbst: 79 rot, 4 grün, 0 flatternd. Die vier grünen sind benannt: Prüfung des Nullwerts von `Code` (äquivalent, `Entscheide` mit Code 0 ergibt `Bleiben`), Probe ohne `ODR-E6002` (äquivalent, `ODR-E6001` deckt dieselbe Kategorie), Probe ohne `ODR-E5001` (äquivalent in Teil 1, mit leerem `Ziel` ergibt `Entscheide` für diesen Code nie `Replay`), `Pfad` an `Entscheide` (akzeptiertes Negativ, §6). Zusage · Mutation · roter Test (Name des Tests der Zeile):

| Zeilen | Mutationen (je rot gesehen) | Test |
|---|---|---|
| Z01 | `MaxBytes < 0`; `Replay`, `Mock`, `Policy`, `Root`, Liste ungeprüft | `TestNew` |
| Z02 | `Laenge <= -1`; `Laenge > 1` bei `Body` nil; `ctx` nil, `Methode`, ganze Prüfung entfernt | `TestUngueltigeAnfrage` |
| Z03, Z04 | Kennzeichnung bei Erfolg; Status 501, `ODR-E6005`, `ODR-E6006`, `ODR-E6008` als Auslöser; Fehler verpackt (Pfad ohne Auslöser); Kontextprüfung entfernt, nur `Canceled`, nur `DeadlineExceeded` | `TestErfolgUnveraendert`, `TestCodesSchleife`, `TestFehlerOhneCode` |
| Z05, Z06 | `Methode` fest `GET`; `Code` fest 0; Replay ohne Freigabe; Replay nie; `ODR-E6002` ohne Replay | `TestEntscheideArgumente`, `TestKategorieOrakel` |
| Z08 bis Z11 | R zweimal; Fehler von R verpackt; Kontextprüfung vor R entfernt; Fehler der Vorstufe statt `ctx.Err()`; `WithoutCancel` an R; `Background` an M | `TestReplayMiss`, `TestReplayFehlerSchleife`, `TestAbbruchZwischenStufen`, `TestKontextAnStellen` |
| Z12 | immer puffern; `Laenge` 0 puffern; `Ziel` der Probe nicht leer; Probe nur `ODR-E6003`; Probe nur `ODR-E6001` | `TestOhnePufferung`, `TestPufferung` |
| Z13 | gemeinsamer Leser; `Laenge` von M oder R weitergereicht; M erhält den Leser der Anfrage | `TestPufferung`, `TestPufferungVollstaendig` |
| Z14 | `>=` an beiden Grenzen; `LimitReader` ohne `+1`; `math.MaxInt64` ohne Schutz; Rest nicht angehängt; Fallback über der Grenze | `TestGrenze` |
| Z15 | Lesefehler verschluckt; Fehler nicht an M; Fallback nach Lesefehler | `TestLesefehler` |
| Z16 | `Close` des Bodys der Anfrage über `io.Closer` | `TestBodyNieGeschlossen` |
| Z17, Z18 | Kopie ohne `X-Recorder-*` an R; Original-Map verändert; Vorabprüfung der Header vor M | `TestHeaderUnveraendert`, `TestSessionHeaderErstImReplay` |
| Z26 | angehängt statt ersetzt; Schreibweise beachtet; Map der Stufe gesetzt; Wertliste geteilt; Kennzeichnung bei M; falscher Wert; zwei Werte | `TestKennzeichnung` |
| Z28 | `Close` von M oder R fehlt; doppelt; bei Erfolg; auf die ausgelieferte Antwort von R | `TestStromSchliessen` |
| Z29, Z30 | M nach R erneut gerufen; Ergebnis je Methode und Pfad gemerkt | `TestMutierenderReplayFallback`, `TestWiederholung` |
| Quelltext (Kopf, Z31) | Goroutine; Variable auf Paketebene; Import von `os`; Import eines anderen Dienstes; Methode auf `Dienst` | `TestQuelltext` |

Z31 und Z32 (nebenläufig, 64 Aufrufe, `n_detektor` und `n_ohne` je 30 bei den Mutationen mit Fenster, je 5 bei den deterministischen): `z31-gemeinsamer-puffer` und `z31-fehler-ueber-paketvariable` (`TestGleichzeitig`) sowie `z32-fehler-ueber-paketvariable` (`TestGleichzeitigerAbbruch`) sind mit Detektor 30 von 30 Prozessen rot (je Prozess `-test.count=1`) und ohne Detektor 30 von 30, **weil die Mutation selbst ein Fenster von 300000 Schleifendurchläufen öffnet**; mit einem Fenster von 2000 leeren Durchläufen (vom Übersetzer entfernt) waren es mit Detektor 30 von 30 und ohne Detektor 0 von 30. Die Zusage über Nebenläufigkeit ist damit durch den Detektor gefangen (30 von 30), der Wirkungs-Fänger zählt die Wirkung nur, wo ein Fenster offen ist; `z32-kontextfehler-durch-code`, `z32-close-doppelt` und `z32-ctxpruefung-entfernt` sind deterministisch rot, mit und ohne Detektor 5 von 5. Generatoren und Orakel: alle Codes aus `model.AlleCodes()` mit jedem Status des Codes und drei fremden (`TestCodesSchleife`, Zählprüfung gegen die Zahl der Codes und die Namen `ODR-E6001`, `ODR-E6002`, `ODR-E6003` aus `SPEC-261` Nr. 3 a), Replay-Fehler über alle Codes (`TestReplayFehlerSchleife`), Kategorien 3 Codes × 4 Mengen (12), Pufferung 4 Policies × 3 Lesemengen × 2 Längenformen × 4 Längen (96), Grenze 4 Werte von `MaxBytes` × bis zu 6 Längen × 3 Formen (deklariert, `-1`, gelogen), Lesefehler 4 Fehler × 3 Bytezahlen × 2 Längenformen (24), Kennzeichnung 16 Vorbelegungen × 2 Stufen. Gegengelesen auf Rohbytes (`LC_ALL=C grep -P`, §3.18) sind alle Dateien des Pakets und `hybrid_test.go` des Ports; das Zeichen KELVIN SIGN entsteht in `antwort_test.go` aus der Bytefolge `0xE2 0x84 0xAA`. Nicht gesehen: eine Mutation auf `strings.EqualFold` statt des ASCII-Vergleichs (der Fänger mit KELVIN SIGN ist geschrieben, aber ohne zweite Import-Stelle nicht als Mutation fahrbar).

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Vertrag von `HybridUseCase`:** Signatur in der Form der Use Cases der anderen Modi, Fehlerabbildung nach `SPEC-109`, Ort in der Spezifikation. **Entschieden** (Architect, 2026-11-08): `SPEC-284` Kopf und Nr. 1, 2, 11 (`Hybrid(ctx, Anfrage) (Antwort, error)`, Dienst `hybrid.Verteiler`, Fehler von Mock und Replay unverändert, `ODR-E4001` und `ODR-E9001` aus `Senden`).
2. **Ablauf und Rangfolge:** Mock, `ODR`-Code, Policy, Replay, Policy, Upstream; welche Stufe welcher Code auslöst (`LH-FA-12.b`), Fehler ohne `ODR`-Code, Abbruch des Kontextes zwischen den Stufen. **Entschieden:** `SPEC-284` Nr. 3 (Code des Ergebnisses, Entscheidung durch `Entscheide`, Replay-Miss, Prüfung von `ctx` vor R und U, jede Stufe höchstens einmal).
3. **Body des Requests:** der Mock liest den Strom, ein Upstream-Fallback eines Requests mit Body braucht ihn erneut. Entscheiden: Pufferung (nur wenn die Policy einen Fallback erlauben kann; Grenze und Fehlercode bei Überschreitung) oder Fallback nur für Requests ohne Body. **Entschieden:** `SPEC-284` Nr. 5: Pufferung nur für fallback-fähige Anfragen (vier Aufrufe von `Entscheide`), Grenze `MAX_CAPTURE_BODY`; darüber und bei einem Lesefehler kein Fallback und kein neuer Fehlercode, der Mock entscheidet wie im Modus `mock`; die Gesamtgrenze über gleichzeitige Requests ist ein akzeptiertes Negativ (Z33).
4. **Zuordnung:** `X-Recorder-Workspace` geht unverändert an Mock und Replay; `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session` sind Pflicht, sobald ein Replay-Fallback konfiguriert ist (`LH-FA-04.c`): wann `ODR-E5003` entsteht (vor dem Aufruf des Mock oder erst im Fallback); alle `X-Recorder-*`-Header fallen vor dem Upstream weg. **Entschieden:** `SPEC-284` Nr. 6 (erst im Replay-Fallback, durch den Replay-Dienst; Nachzug in `LH-FA-04.c`) und Nr. 8 c (Filter).
5. **Antwort des Fallbacks:** ob sie sich als Fallback kennzeichnet (Header), welche Header der Upstream-Antwort weitergehen, und der Umgang mit absoluten Upstream-Links in der Antwort (`SPEC-092`, Risiko in §6). **Entschieden:** `SPEC-284` Nr. 9: Header `X-Recorder-Fallback: replay` oder `upstream`; Header der Upstream-Antwort außer Hop-by-Hop unverändert; Links bleiben, der Hinweis steht im Handbuch-Kapitel von `slice-bootstrap-m4` (Sendung, §6).
6. **Fehler des Upstream im Fallback** (`ODR-E4001`, `SPEC-120` Nr. 8): unverändert ausliefern, kein Wiederholen, kein Rückfall auf Replay. **Entschieden:** `SPEC-284` Nr. 8 und 11 (der Fehler von `Senden` ersetzt den früheren Fehler).
7. **Ströme:** die verworfene Antwort des Mock und jede Antwort, die nicht ausgeliefert wird, schließt der Dienst genau einmal, auch nach Abbruch. **Entschieden:** `SPEC-284` Nr. 10: eine ausgelieferte Antwort schließt der Adapter, der Dienst nie; eine verworfene (Fehler zugleich mit Antwort) genau einmal; den Body des Requests schließt der Dienst nie. Die Zeile im Auftrag nannte „jede Antwort, die nicht ausgeliefert wird“: eine solche gibt es außer dieser Doppelform nicht.
8. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben. **Erledigt** (Architect, 2026-11-08): Tabelle in diesem Abschnitt, Raster in §6, Nachzug in `LH-FA-04.c`, `LH-FA-12.b`, `architecture.md` §2.2 und im Änderungsprotokoll der Spezifikation.
9. **Header des Upstream-Fallbacks** (`SPEC-261` Nr. 10, *Ungeprüft*): welche Request-Header der Dienst sendet, darunter die Header zum Überschreiben der Methode (`X-HTTP-Method-Override` und Verwandte): **Entschieden** (Architect, 2026-11-08): entfernen; `SPEC-261` Nr. 10 c nennt die drei Header, *Ungeprüft* nennt andere Namen und die Query. **Vertrag:** `SPEC-284` Nr. 8 c (Kopie der Header ohne `X-Recorder-*`, Hop-by-Hop, `Expect`, `Host`, `Content-Length` und die drei Namen; `Accept-Encoding` unverändert).

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen; ergänzt am 2026-11-08): *an diesen Slice* aus `slice-fallback-policy-entscheidung` (dieselbe Zeichenkette als `Ziel`, Aufruf vor jedem Senden, Header-Auswahl, Zustand des Mock beim mutierenden Replay-Fallback): angenommen, §1 *Übernimmt* und DoD (Punkte 2 und 3) tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie (der Ausschluss der Policy betrifft deren Syntax und Entscheidung, nicht den Aufruf). *Von diesem Slice an* `slice-bootstrap-m4` (Verdrahtung des Dienstes im Composition Root): angenommen, §1 und DoD Punkt 1 von `slice-bootstrap-m4` tragen sie. *An diesen Slice:* das Risiko der Upstream-Links aus `welle-hybrid` §6, angenommen (§1 *Übernimmt*, §6). **Verteilung nach dem Schnitt** (Planner, 2026-11-08; §1 *Ausdrücklich NICHT* und DoD von [slice-hybrid-service-upstream](../open/slice-hybrid-service-upstream.md) gelesen, es trägt die Sendungen vor dem Geber: §1 *Übernimmt* und DoD Punkt 1 mit den Kennungen der Geber, kein Ausschluss trifft sie): an Teil 2 gehen `Ziel` als dieselbe Zeichenkette, `Entscheide` vor jedem Senden, die drei Header (`SPEC-261` Nr. 10 a bis c), der Host-Fänger am Referenzdienst und das Risiko der Upstream-Links; bei Teil 1 bleibt der Zustand des Mock beim mutierenden Replay-Fallback (Z29). `slice-bootstrap-m4` trägt die Pflichten des Roots (Sendung von diesem Slice) mit den Kennungen beider Teile.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-fallback-policy`, `slice-fallback-policy-entscheidung`, `slice-mock-zuordnung`, `slice-mock-lesen`, `slice-mock-schreiben` und `slice-replay-service` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum diese sechs:**

- `slice-fallback-policy` und `slice-fallback-policy-entscheidung` (der Policy-Slice ist vor dem Code geschnitten, der zweite baut auf dem ersten): Typen, `New` und Entscheidungsfunktion, die der Dienst ruft (*Code-Abhängigkeit*).
- `slice-mock-zuordnung` und `slice-mock-lesen` (die beiden Teile des geschnittenen Plans): `slice-mock-zuordnung` legt den Inbound Port `MockUseCase` an (er liegt heute nicht in `src/hexagon/ports/inbound/`) und liefert `ODR-E6001`; `slice-mock-lesen` liefert `ODR-E6003`, nach dem der Dienst mit `ODR-E6001` entscheidet (*Code-Abhängigkeit, Port*).
- `slice-mock-schreiben`: legt fest, welche Mutationen der Mock mit `ODR-E6002` als nicht unterstützt abweist; ohne diesen Vertrag hat die Zeile „POST auf eine nicht unterstützte Operation“ der Matrix keinen Boden (*Vertrags-Abhängigkeit*, keine Typ-Abhängigkeit).
- `slice-replay-service`: legt den Inbound Port `ReplayUseCase` an (heute nicht vorhanden) und liefert `ODR-E5001`, `ODR-E5002` und `ODR-E5003` (*Code-Abhängigkeit, Port*).

**Stand der Bedingungen beim Start** (Planner, 2026-11-08, `main` = `33d3ba55`): alle sechs Start-Slices liegen in `done/` ([slice-fallback-policy](../done/slice-fallback-policy.md), [slice-fallback-policy-entscheidung](../done/slice-fallback-policy-entscheidung.md), [slice-mock-zuordnung](../done/welle-mock/slice-mock-zuordnung.md), [slice-mock-lesen](../done/welle-mock/slice-mock-lesen.md), [slice-mock-schreiben](../done/welle-mock/slice-mock-schreiben.md), [slice-replay-service](../done/welle-record-replay/slice-replay-service.md)). **Noch offen vor dem ersten Code-Commit:** der Vertrag des Architect in der Spezifikation (Aufträge 1 bis 8 in §3, die Randformen des Rasters in §6), die Tabelle *Zusage · Fänger* in §3 und, daraus folgend, der Block der Größenmarke (`SPEC-128` Nr. 3); der Architect arbeitet im Zweig nach dem Start-Commit (Nr. 3 der Übergangsregeln). **Gate-Kette (Nr. 7 der [Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln)):** kein Kettenslice in `in-progress/` (dort liegen nur `slice-if-match` und dieser Slice); dieser Slice ändert `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` und den Stempel nicht (Pfade in §3: `src/hexagon/ports/inbound/`, `src/hexagon/services/hybrid/`, `spec/spezifikation.md`, Tests), Linie B, kein Kettenslice. Parallel läuft `slice-if-match` (`src/hexagon/services/mock/`, `src/hexagon/model/`, `src/adapters/driving/httpodata/`); kein gemeinsamer Pfad aus §3 außer `spec/spezifikation.md` und gegebenenfalls `src/hexagon/ports/inbound/` (Port `MockUseCase` unverändert zu lassen ist Voraussetzung dieses Slice; ändert `slice-if-match` ihn, nimmt der spätere Zweig `main` vor dem Handoff auf). Kennungen: Linie B vergibt aus `SPEC-260` bis `SPEC-299` (zuletzt `SPEC-282` belegt, `SPEC-261` ist die Policy); weitere Nummern für den Vertrag von `HybridUseCase` vergibt der Architect aus diesem Bereich. **Nachtrag** (Architect, 2026-11-08): der Vertrag `SPEC-284` mit Randformen, die Tabelle *Zusage · Fänger* in §3, das Raster in §6 und der Block der Größenmarke (Teil des Architect, §4) liegen vor; offen bleibt der Teil des Planner (Schnitt beschließen, `basis`, `schaetzung`, `marke`).

**Größenmarke** (Planner, 2026-11-08, `SPEC-128` Nr. 1): *S* = 85·*a* + 80·*b* + 200·*f* + *p*; *M* = 100 · ⌈135·*S* / 10000⌉; Schnittfrage genau dann, wenn *S* ≥ 2223. Die Marke dieses Teils steht im Block unten; der Architect lieferte Tabelle, Klassen und Produktcode, der Planner hat vor dem Code geschnitten (nächster Absatz).

**Größe nach dem Vertrag** (Architect, 2026-11-08, Teil des Architect nach `SPEC-128` Nr. 3 k; `basis`, `schaetzung` und `marke` setzt der Planner vor dem ersten Code-Commit): die Tabelle in §3 trägt 36 Zeilen, davon *a* = 2 nebenläufig (Z31, Z32) und *b* = 34 andere (darunter vier akzeptierte Negative, Z33 bis Z36); *f* = 4 (Fakes von `MockUseCase`, `ReplayUseCase`, `UpstreamClient` und der Referenzdienst mit Prüf-Client); der Produktcode wird mit 580 Zeilen erwartet (Port 25, Konfig und `New` 70, Ablauf und Entscheidung 140, Pufferung 90, `UpstreamAnfrage` mit Header-Filter 120, Antwort und Kennzeichnung 70, Fehler und Dienstwahl 65). Das ergibt *S* = 85·2 + 80·34 + 200·4 + 580 = 4270, *M* = 5800 und die Schnittfrage mit **ja** (*S* ≥ 2223).

**Schnitt, vorab geprüft (Antwort des Architect):** der Vorschlag des Planner trägt. Ein Schnitt entlang der Stufen ist additiv: Teil 2 nutzt, was Teil 1 schreibt, und schreibt nichts davon um.

- *Teil 1, Mock zuerst und Replay-Fallback:* Z01 bis Z06, Z08 bis Z18, Z26 (Wert `replay`), Z28 (M und R), Z29 bis Z34, Z36 = 26 Zeilen (*a* = 2, *b* = 24), *f* = 2 (Mock, Replay), Produktcode rund 330 (Port, Konfig, Ablauf, Entscheidung, Pufferung, Kennzeichnung, Dienstwahl): *S* ≈ 2820, *M* ≈ 3900. Die Konfig führt in Teil 1 weder `Upstream` noch `Dienst.UpstreamURL`; `Entscheide` bekommt an allen Aufrufen (Probe, H1, H3, H5) ein leeres `Ziel`, also entsteht das Ergebnis `Upstream` nie (Entscheidung zu §6), und `ODR-E5001` aus R kehrt unverändert zurück.
- *Teil 2, Upstream-Fallback:* Z07, Z19 bis Z25, Z27, Z35 und je eine Zeile für die Upstream-Seite von Z08 (Miss zu U), Z09 (Replay-Fehler lösen U nie aus), Z10 (Abbruch vor U), Z11 (H6), Z26 (Wert `upstream`) und Z28 (Strom von U) = 16 Zeilen (*b* = 16), *f* = 2 (Fake `UpstreamClient`, Referenzdienst), Produktcode rund 250 (Feld `Upstream` und die Prüfung in `New`, der Zweig `Upstream`, `UpstreamAnfrage` mit Header-Filter, Antwort, Fehlerabbildung): *S* ≈ 1930, *M* ≈ 2700.
- *Geteilter Mechanismus:* Entscheidung am Code, Probe auf Fallback-Fähigkeit, Pufferung des Bodys, Kopie der Header mit Kennzeichnung und die Prüfung von `ctx` zwischen den Stufen liegen in Teil 1; Teil 2 fügt den Zweig `Upstream` im Schalter von Nr. 3 c und das Feld `Konfig.Upstream` hinzu (der Vertrag lässt Port und Konfig additiv wachsen) und ändert keine Zeile der Pufferung. Ein Schnitt ohne geteilten Mechanismus, den der zweite Teil neu schriebe, ist damit möglich. Teil 1 liegt mit *S* ≈ 2820 selbst über 2223; ein weiterer Schnitt wäre nur zwischen Ablauf und Pufferung möglich (Teil 1a: Anfragen mit Body erhalten keinen Fallback; Teil 1b: Pufferung), und die beiden teilen den Aufruf von M und die Probe; der Gewinn (alle Teile unter 2223) kostet einen dritten Plan samt Review, Verifikation und Closure. Empfehlung: zwei Teile, bei Teil 1 die Schnittfrage mit „nein“ beantwortet und diesem Grund in §4.

**Schnitt beschlossen und Marke dieses Teils** (Planner, 2026-11-08, vor dem ersten Code-Commit): zwei Teile, wie der Architect empfiehlt; Teil 2 ist [slice-hybrid-service-upstream](../open/slice-hybrid-service-upstream.md). Zahlen nach `SPEC-128` Nr. 1:

| Teil | Zeilen (*a* / *b*) | *f* | *p* | *S* | *M* | Schnittfrage |
|---|---|---|---|---|---|---|
| ungeschnitten (Architect) | 36 (2 / 34) | 4 | 580 | 4270 | 5800 | ja |
| Teil 1 (dieser Slice) | 26 (2 / 24) | 2 | 330 | 2820 | 3900 | ja, Grund unten |
| Teil 2 | 17 (0 / 17) | 2 | 250 | 2010 | 2800 | nein |
| Alternative: Teil 1a (ohne Pufferung: Z13, Z14, Z15, Z33 fallen weg) | 22 (2 / 20) | 2 | 240 | 2410 | 3300 | ja |
| Alternative: Teil 1b (Pufferung) | 4 (0 / 4) | 0 | 90 | 410 | 600 | nein |

Teil 2 trägt 17 statt 16 Zeilen: Z01 (U) kommt hinzu, weil Teil 1 kein Feld `Upstream` führt (der Teil des Architect nannte Z01 ganz bei Teil 1). Die Rechnung von Teil 1a ist die des Planner: 26 − 4 = 22 Zeilen, Produktcode 330 − 90 = 240; Z12 (Probe) und Z16 bleiben bei 1a, weil die Probe den Aufruf von M bestimmt.

**Grund, dass Teil 1 nicht weiter geschnitten wird** (`AGENTS.md` §3.14): (1) Der einzige weitere Schnitt entlang einer Naht ist der zwischen Ablauf und Pufferung (1a / 1b); er bringt *kein* Teil unter die Grenze 2223: 1a bliebe bei *S* = 2410 mit Schnittfrage, die Behauptung „alle Teile unter 2223“ des Schnittvorschlags trifft nicht zu, und ein weiterer Schnitt von 1a (etwa Z31 und Z32 abzutrennen, *S* 2240) bliebe ebenfalls darüber. (2) 1a und 1b teilen den Mechanismus: die Probe auf Fallback-Fähigkeit (vier Aufrufe von `Entscheide` vor M, Z12) und den Aufruf von M mit dem Leser aus dem Puffer statt dem der Anfrage (Z13 bis Z15); 1b schriebe den Aufruf von M in 1a um. (3) Der Schnitt Mock-nur gegen Replay ergäbe einen Teil ohne Lieferwert (ein Hybrid-Dienst, der nichts als der Mock ist). Der Preis eines dritten Plans (Review, Verifikation, Closure) steht ohne den Gewinn, den der Schnitt verspricht. Teil 1 bleibt ungeteilt mit dem Eintrag `schnitt-grund`; überschreitet die Messung die Marke 3900, gilt `SPEC-128` Nr. 5 d (Eintrag mit Grund), keine Rückführung nach dem Code.

```groessenmarke
pfad: src/hexagon/ports/inbound/
pfad: src/hexagon/services/hybrid/
produktcode: 330
fakes: 2
basis: 6de6da87cff2c132a76def7ac9667ad4ca4f13f4
schaetzung: 2820
marke: 3900
schnitt-grund: Schnittfrage ja (S 2820); ein Schnitt zwischen Ablauf und Pufferung bringt kein Teil unter 2223 (1a S 2410) und teilt die Probe auf Fallback-Fähigkeit und den Aufruf von M mit gepuffertem Body; der Upstream-Fallback ist als Teil 2 abgetrennt (slice-hybrid-service-upstream).
```

**Nicht gebraucht:** `slice-bootstrap-m3`, `slice-snapshot-use-case`, `slice-mock-abnahme` und M3, denn der Dienst läuft gegen Fakes der Ports und der Composition Root gehört `slice-bootstrap-m4`; auch M2 (`slice-bootstrap-m2`) braucht der Dienst nicht selbst, nur das, was `slice-replay-service` und `slice-mock-lesen` von ihr voraussetzen. **Ohne die Ports des Mock startet der Slice nicht.**

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: 3900** (Block oben, Teil 1; *S* = 2820).
- `in-progress` → `open` (blockiert — Carveout?): der Body eines Requests lässt sich für den Upstream-Fallback nicht erhalten, ohne den Port `MockUseCase` zu ändern (Auftrag 3), oder die Zuordnung der Replay-Header im Hybrid-Modus ist nicht entschieden (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-08): jede Zelle nennt ihren Ort in `SPEC-284` oder in der Spezifikation, oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12). Die Fänger der Zellen stehen in der Tabelle in §3 (Zeilen Z01 bis Z36; die Zeile *Upstream-Fallback*, die Spalte U der Zusatztabelle und die Fänger Z07, Z19 bis Z25, Z27, Z35 gehören zu Teil 2, `slice-hybrid-service-upstream`).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Aufruf des Mock und Entscheidung nach dem Fehlercode | `ctx`, `Methode`, `Body` nil bei `Laenge` > 0: `SPEC-284` Nr. 2; Fehler ohne Code: Nr. 3 b; kein Dienst zum Pfad: Nr. 4 | `Laenge` 0 mit Leser: Nr. 5 b; Nullwert des Codes: Nr. 3 b; leere `UpstreamURL`: Nr. 1 und `SPEC-261` Nr. 5 c; leerer Pfad: `ODR-E5005` des Mock, Nr. 4 | unbekannter Code: `SPEC-261` Nr. 3 c über Nr. 3 b; unbekannte Methode: `SPEC-261` Nr. 2 c über Nr. 3 c; Header zum Überschreiben der Methode: `SPEC-261` Nr. 2 d | Nr. 3 e (Prüfung vor R und U), Nr. 2 (Kontextfehler einer Stufe) | Nr. 13 (jeder Aufruf ruft die Stufen erneut) |
| Replay-Fallback (Zuordnungs-Header, Kategorie, Replay-Fehler) | fehlende Session-Header: Nr. 6 (`ODR-E5003` des Replay-Dienstes, `SPEC-135` Nr. 5); fehlender Body: Nr. 5 b | leerer Wert eines Session-Headers: Nr. 6 und `SPEC-135` Nr. 5; leere Kategorie: gilt nicht, die Policy lehnt sie in `New` ab (`SPEC-261` Nr. 7) | unbekannte oder fremde Session: Replay-Fehler unverändert, Nr. 3 d; unbekannte Kategorie: gilt nicht, `SPEC-261` Nr. 7 | Nr. 3 e, Nr. 2 | Nr. 13; Cursor-Verbrauch je Treffer: `SPEC-135` Nr. 17 |
| Upstream-Fallback (Body, `X-Recorder-*`, Antwort des Upstream) | fehlender Body: Nr. 5 b und 8 b; fehlende `UpstreamURL`: Nr. 4; fehlende `X-Recorder-*`: gilt nicht, der Filter entfernt nur (Nr. 8 c) | leerer Header-Wert: geht mit (Nr. 8 c); leere Query: unverändert (Nr. 8 a); leere Antwort: Nr. 9 | fremder Host: `SPEC-261` Nr. 5, Verbindungen 0 (Nr. 3 c); unbekannte Methode: `SPEC-261` Nr. 2 c; unbekannter Header-Name: geht mit (Nr. 8 c, `SPEC-261` *Ungeprüft*); Fehler des Upstream: Nr. 11 | Nr. 3 e, Nr. 11 (Kontextfehler ohne Code) | Nr. 8 (kein Wiederholen von `Senden`), Nr. 13 |
| Schließen der Ströme (verworfene Antwort einer Stufe, ausgelieferte Antwort, Body der Anfrage) | `UpstreamAntwort.Body` nil: *Ungeprüft* (Nr. 10 und Ende des Vertrags); Body der Anfrage nil: Nr. 5 b | leerer Strom der Antwort: unverändert, nie gelesen (Nr. 9, 10) | Antwort zugleich mit Fehler: Nr. 10 | Nr. 10 (genau einmal, auch bei Kontextfehler), Nr. 5 d | zweites `Close`: Pflicht des Adapters (`SPEC-121` Nr. 7), der Dienst schließt eine ausgelieferte Antwort nie (Nr. 10) |

**Zusatztabelle: Ergebnis des Mock (Matrix `LH-FA-12.b`) × Zustand der Policy** (die Antwort hängt vom Zustand ab, `AGENTS.md` §3.16); „M“ heißt: der Fehler des Mock kehrt unverändert zurück, „R“ Replay-Fallback, „U“ Upstream-Fallback.

| Ergebnis des Mock | alle Schlüssel leer | Replay-Kategorie gesetzt | `enabled`, Host frei, lesend | `enabled`, Host frei, mutierend mit Route | Fänger |
|---|---|---|---|---|---|
| Erfolg | Antwort | Antwort | Antwort | Antwort | Z03 |
| fachliche `404`, `409`, `412` und jeder andere Code außerhalb der Auslöser | M | M | M | M | Z03 |
| Coverage unvollständig (`ODR-E6003`) | M | R bei Kategorie `coverage_insufficient`, sonst M | U (ohne Kategorie); mit Kategorie R, nach Miss U | U nur mit Route (ohne Kategorie); mit Kategorie R | Z06, Z07, Z21 |
| nicht unterstützte Operation (`ODR-E6001`, `ODR-E6002`) | M | R bei Kategorie `unsupported`, sonst M | U (ohne Kategorie); mit Kategorie R, nach Miss U | U nur mit Route (ohne Kategorie); mit Kategorie R | Z06, Z07, Z21, Z29 |
| ungültige Anfrage (`ODR-E6005`, `ODR-E6007`) | M | M | M | M | Z03 |
| Replay-Miss (`ODR-E5001`) | – | M | U | M (mutierend erreicht nach einem Miss den Upstream nie) | Z08, Z21 |
| anderer Replay-Fehler (auch `ODR-E5002`) | – | M | M | M | Z09 |

- Der Body eines Requests lässt sich nur einmal lesen; der Mock liest ihn, ein Upstream-Fallback braucht ihn erneut — **Ausgang:** weiter offen bis zur Closure; entschieden (Architect, 2026-11-08): `SPEC-284` Nr. 5, Pufferung nur für fallback-fähige Anfragen bis `MAX_CAPTURE_BODY`, darüber kein Fallback; die unbeschränkte Menge ist der Speicher über alle gleichzeitigen Requests, Prüfgrenze `MaxBytes` je Request, akzeptiertes Negativ Z33 mit Folge in `SPEC-284` *Ungeprüft*. Adresse: Implementer (Fänger Z12 bis Z16), Verifier.
- Eine Fallback-Antwort kann absolute Upstream-Links tragen, `LINK_REWRITE` kennt nur `off` (`SPEC-092`; Risiko aus `welle-hybrid` §6) — **Ausgang:** weiter offen → Teil 2 (`slice-hybrid-service-upstream`, §6 dort; Fänger Z26 (U), Z27): die Antwort aus dem Upstream trägt `X-Recorder-Fallback` und bleibt unverändert (`SPEC-284` Nr. 9); der Hinweis im Handbuch-Kapitel von `slice-bootstrap-m4` ist dort in die DoD eingetragen.
- **Rückgabe an den Architect (Planner, 2026-11-08): Probe auf Fallback-Fähigkeit und `Ziel` zwischen den Teilen.** — **Ausgang:** entschieden (Architect, 2026-11-08): Teil 1 führt weder `Dienst.UpstreamURL` noch `Konfig.Upstream` und gibt `Entscheide` an allen vier Stellen (Probe `SPEC-284` Nr. 5 a, H1, H3, H5) ein leeres `Ziel`; die Policy gibt es nie frei (`SPEC-261` Nr. 5 c), das Ergebnis `Upstream` entsteht nicht, ohne dass ein Schalter es abbildet. Die andere Form (Feld in Teil 1, `Upstream` auf `Bleiben` abgebildet) ist verworfen: sie führt tote Abbildung und Feld ohne Leser, die Teil 2 wieder entfernen müsste. *Folge:* Teil 2 ändert an der Pufferung kein Verfahren, sondern das Argument `Ziel` der vier Aufrufe von `Entscheide` (dadurch zählt eine reine Host-Freigabe als fallback-fähig und puffert); die Spezifikation bleibt unverändert, `SPEC-284` beschreibt den Endzustand (Nr. 1, 4, 5 a). Teil 2 trägt dafür die Zeile Z12 (U) (Planner, Hinweis im Bericht). Adresse: erledigt.
- Die Zuordnung der Replay-Header im Hybrid-Modus (`ODR-E5003` vor dem Mock-Aufruf oder erst im Fallback) hat keinen Ort in `LH-FA-04.c` — **Ausgang:** entfallen als offene Frage: entschieden `SPEC-284` Nr. 6 (erst im Replay-Fallback, durch den Replay-Dienst), Nachzug in `LH-FA-04.c`; Fänger Z18.
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×): die Fakes von `MockUseCase`, `ReplayUseCase` und `UpstreamClient` tragen jede Schnittstelle des Originals (der Body als `ReadCloser`, `Close` gezählt), zählen jeden Aufruf und halten jedes Argument fest; ein Fake, der panikt, verdeckt den benannten Test nicht — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (17×): die Matrix hat viele Zweige, der Dienst hält keinen Zustand; gleichzeitige Aufrufe laufen als Fall unter `make test-race` — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- **Abweichungen vom Plan, Teil 1 (Implementer, 2026-11-08, `AGENTS.md` §3.13):** (a) *Dienstwahl*: ohne `Dienst.UpstreamURL` ist `Ziel` immer leer, und der Root bestimmt in Teil 1 kein Ergebnis (`SPEC-284` Nr. 4 hat nur in der Wirkung auf `Ziel` und `ODR-E5005` des Mock Folgen, beide sind in Teil 1 nicht beobachtbar); der Code wählt keinen Dienst, `New` prüft `Dienste` und speichert sie nicht, damit kein Code ohne Leser und ohne Fänger entsteht. Z01 (Mutation „Liste nicht kopiert“, Fänger „Dienstwahl bleibt“) entfällt in Teil 1. (b) *H5* (`Entscheide` nach einem Replay-Miss, `SPEC-284` Nr. 3 d) ist nicht geschrieben: mit leerem `Ziel` ist das Ergebnis immer `Bleiben`, der Zweig `Upstream` gehört zu Teil 2. (c) *akzeptierte Negative mit Grund und Folge* (Zeilen ohne Fänger in Teil 1): das `Pfad`-Argument an `Entscheide` (Z05) und die Zahl der Aufrufe von `Entscheide` (Z02 „ohne Aufruf“, H1 mit vier Aufrufen) — Grund: `*policy.Policy` ist ein konkreter Typ, der Replay-Zweig der Policy liest den Pfad nie und ein leeres `Ziel` wird nie freigegeben; Folge: ein falscher `Pfad` oder eine fehlende Probe bliebe in Teil 1 unentdeckt, bis Teil 2 mit Route und Host (Z19 bis Z21) einen Fänger bekommt. (d) *Randform ohne Ort*: ein `*inbound.Fehler` mit typisiertem nil-Zeiger lässt `errors.Is` in `Unwrap` paniken (`src/hexagon/ports/inbound/fehler.go`); `SPEC-284` nennt die Form nicht, der Dienst entscheidet sie nicht, und der Fänger Z04 führt zehn Formen ohne diese. **Ausgang:** offen; Adresse: (a) Planner: §1 *Ausdrücklich NICHT* von [slice-hybrid-service-upstream](../open/slice-hybrid-service-upstream.md) nennt „Dienstwahl ohne `UpstreamURL`“ als Teil 1, Teil 1 schreibt sie nicht, also ist der Punkt keine Adresse bei Teil 2 (`AGENTS.md` §3.13): entweder nimmt Teil 2 Dienstwahl und Fänger von Z01 und Nr. 4 in §1 auf, oder Teil 1 bekommt eine Dienstwahl ohne Wirkung zurück (dann ohne Fänger, ein akzeptiertes Negativ); (b) und (c) Teil 2 (dessen §1 *Übernimmt* den Zweig `Upstream` in Nr. 3 c und d), Planner prüft die Annahme der Fänger von `Pfad` und Probe; (d) Architect.
- **FRAGEN AN DEN USER (drei, aus `slice-fallback-policy-entscheidung`; der Planner entscheidet sie nicht, sie bremsen den Start nicht, die Antwort ändert `SPEC-261` und, soweit sie den Dienst färbt, Raster und Tabelle dieses Slice):**
  - **FRAGE 1 (User): Replay-Fallback auch für mutierende Methoden?** Bis zur Antwort gilt `SPEC-261` Nr. 4 b = **ja**. Folge für diesen Slice: ein Replay-Fallback einer mutierenden Methode liefert eine aufgezeichnete Erfolgsantwort, der Mock bleibt unverändert (DoD Punkt 2, Fänger zählt die Schreibzugriffe des Mock-Fakes). Antwortet der User „nur lesend“, entfällt die Zeile im Replay-Fallback; Zeilen von Raster und Tabelle ziehen nach.
  - **FRAGE 2 (User): `ODR-E6006` (V2: Eigenschaft außerhalb des Umfangs) als Auslöser?** Bis zur Antwort **nein**, kein Fallback (`SPEC-261` Nr. 3 d). Folge: der Dienst liefert `ODR-E6006` unverändert aus; bei „ja“ kommt eine Stufe der Matrix hinzu.
  - **FRAGE 3 (User): Lücke im Lastenheft?** Der Replay-Fallback hat dort weder Anforderung noch Akzeptanzkriterium (Change Request vor `Accepted` oder Bestätigung); der Dienst baut ihn aus `LH-FA-12.a` und `LH-FA-12.b` der Spezifikation.
  - **Ausgang:** weiter offen; Adresse: User; Beleg: `slice-fallback-policy-entscheidung` §6 und §7.
- Sendungen beim Start (Planner, 2026-11-08, `AGENTS.md` §3.13): *von* `slice-fallback-policy-entscheidung` (`SPEC-261` Nr. 10 a bis c, Mock-Zustand beim mutierenden Replay-Fallback) und *von* `slice-fallback-policy` (Fänger: der Referenzdienst empfängt unter dem Host, den die Policy freigab) — **Ausgang:** eingetreten als Sendung, angenommen: §1 *Übernimmt* und DoD Punkte 2 und 3 tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie; Fänger-Zeilen der Sendung (Host-Fänger am Referenzdienst, Schleife 3 Namen × Schreibweisen, Zähler der Schreibzugriffe des Mock) gehören in die Tabelle des Architect. Adresse: Architect.
- Merge-Reihenfolge: `slice-if-match` läuft parallel; gemeinsam sind nur `spec/spezifikation.md` und, falls `slice-if-match` den Port `MockUseCase` berührt, `src/hexagon/ports/inbound/` — **Ausgang:** offen; Adresse: Planner (wer später merged, nimmt `main` vor dem Handoff auf). Befund des Architect (2026-11-08, Zweig `slice-if-match` bei `e1a41ccc` gelesen): `slice-if-match` ändert `src/hexagon/ports/inbound/` nicht (`MockUseCase` bleibt), führt `SPEC-283` und den Code `ODR-E6010` ein (eine bedingte Mutation, die der Mock mit `412` abweist); `ODR-E6010` ist kein Auslöser von `SPEC-261` Nr. 3 c (geschlossene Liste), Z03 schließt ihn über `model.AlleCodes()` ein, sobald `main` ihn trägt; Z03 und Z09 müssen nach dem Merge grün bleiben (Zählprüfung gegen `AlleCodes()`). Konfliktflächen in `spec/spezifikation.md`: das Änderungsprotokoll (beide Zweige fügen oben eine Zeile ein, der zweite löst den Konflikt) und die Tabelle der Codes in `SPEC-109` (nur `slice-if-match`); `SPEC-284` steht hinter `SPEC-261`, `SPEC-283` davor, die Hunks überlappen nicht.
- Gate-Kette (Nr. 7 der Übergangsregeln): der Slice ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette in §3.
- Größe: ungeschnitten *S* = 4270, *M* = 5800 (36 Zeilen, Schnittfrage **ja**); vor dem Code in zwei Teile geschnitten (§4): Teil 1 (dieser Slice) 26 Zeilen, *S* = 2820, Marke 3900, `schnitt-grund` im Block (kein weiterer Schnitt bringt ein Teil unter 2223); Teil 2 *S* = 2010, Marke 2800 — **Ausgang:** eingetreten als Schnitt; die Messung am Handoff trägt der Planner als Messpunkt ein (`SPEC-128` Nr. 9). Wächst die Tabelle von Teil 1 oder der Produktcode, rechnet der Planner neu (eine Marke nach dem ersten Code-Commit ändert nur mit Eintrag `ersetzt`).
- Sendung *von* diesem Slice *an* `slice-bootstrap-m4` (`AGENTS.md` §3.13, Architect, 2026-11-08; §1 *Ausdrücklich NICHT* und DoD von `slice-bootstrap-m4` gelesen: kein Ausschluss trifft sie, `LINK_REWRITE` bleibt `off` und ist hier nicht berührt): (a) der Root übergibt `hybrid.Konfig` nach `SPEC-284` Nr. 1 (`MaxBytes` gleich `MAX_CAPTURE_BODY`, `Upstream` immer gesetzt, `Policy` aus `policy.New`, `Dienst.UpstreamURL` je Service, leer ohne `UPSTREAM_URL`); (b) das Handbuch-Kapitel nennt `X-Recorder-Fallback`, dass absolute Upstream-Links einer Fallback-Antwort unverändert bleiben, dass `ODR-E5003` fehlender Session-Header erst im Replay-Fallback entsteht und dass ein Request mit Body über `MAX_CAPTURE_BODY` nie einen Fallback erhält — **Ausgang:** eingetreten als Sendung, angenommen: beide Punkte stehen im selben Commit in der DoD von `slice-bootstrap-m4` (Punkt 1 und Handbuch-Punkt) mit der Kennung dieses Slice. Adresse: Planner (prüft die Annahme).
- Sendung *an* diesen Slice, die der Vertrag trägt (`SPEC-261` Nr. 10 a bis c, Mock-Zustand, Host-Fänger): Zeile Z29 bleibt hier, die Zeilen Z19, Z20, Z21, Z22 der Tabelle gehören zu Teil 2; die Lesart von `Ziel` durch `Senden` belegt der Verbindungszähler am gebauten Binary in `slice-bootstrap-m4` (DoD Punkt 2, bereits getragen) — **Ausgang:** eingetreten als Sendung, angenommen; die Pflichten 10 a bis c weitergegeben an Teil 2, der sie in §1 und DoD Punkt 1 trägt (Z35 dort nennt die Grenze).
- Zusage-Raster: die Negative Z33, Z34 und Z36 (Teil 1) und Z35 (Teil 2) tragen Grund und Folge in `SPEC-284` *Ungeprüft*; die Zusage „Entscheide vor jedem Senden“ fängt Z21 (Zähler 0 ohne Freigabe), die Zusage „dieselbe Zeichenkette“ Z19, Z20, beide in Teil 2 — **Ausgang:** entfallen als Frage; der Reviewer löst die Zeilen mit je einer Mutation auf.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-08; Zähler sind die Dateien unter `evidence/`). Der Slice hat drei Fakes und viele Zweige: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×) sind verkörpert (`.claude/commands/implement-slice.md` Schritt 14, `.harness/skills/reviewer.md`) und gelten für die Fakes (§6); `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (17×, offen, Schwelle erreicht, Vorlage für den Reviewer-Skill beim User) für die Zweige der Matrix, die kein Lauf ausführt (§6). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (57×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (15×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (35×, `AGENTS.md` §3.16), `BEO-REPO/zustand-des-arguments-als-dimension` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (14×) für die Pufferung des Bodys (Auftrag 3). Offen und mit dem Slice berührt: `BEO-REPO/sendung-nennt-nicht-alle-pflichten-des-vertrags` (1×; der Plan trägt die vier Pflichten aus Nr. 10 und prüft sie gegen jede Nummer des Vertrags, die eine Pflicht des Aufrufers trägt) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (17×, geplant in `slice-guard-schreibwege`; die Betriebsregeln in §3 gelten bis dahin). Der Slice schreibt keine Zähler, er trägt Evidenz-Dateien bei der Closure bei.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
