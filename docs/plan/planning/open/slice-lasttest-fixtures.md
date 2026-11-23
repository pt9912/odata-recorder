# Slice slice-lasttest-fixtures: Profile, Referenzdienst, Aufbau über `bootstrap.Build`, Listen und Umgebung des Lasttests (Teil 1 des Schnitts von `slice-lasttest`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-QA-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)

**Berührte Spec-Stellen** (Teil 1 des Schnitts von [slice-lasttest](slice-lasttest.md); die Stellen der Teile 2 und 3 stehen in deren Plänen): `SPEC-216` ([der Vertrag des Werkzeugs](../../../../harness/vertraege/SPEC-216-lasttest.md), Architect, 2026-11-23: Nr. 4 c, 5, 6, 7, 9 a, 14 a) · `LH-QA-01.a` (Messaufbau: 10.000 Exchanges, Bodies bis 64 KiB; Mockserver mit 10.000 Entities) · `SPEC-004` (Match-Key-Index `(session_id, match_key)` ohne Tabellenscan; belegt von `make schema-check` und `make migration-check`, `SPEC-216` Nr. 6 g) · `SPEC-110` (die Namen der Defaults von `REDACT_*`, die der Referenzdienst meidet) · `SPEC-037` Nr. 6 (der Übergang, dem der Aufbau folgt) · spezifikation.md §6 (Abnahmeparameter: „der Messaufbau muss reproduzierbar im Test-Harness hinterlegt sein“) · `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-lasttest](slice-lasttest.md), Vorschlag des Architect vom 2026-11-23). **Datum:** 2026-11-23 (der Plan des ungeteilten Slice vom 2026-10-30 trug Ziel, Tabelle und Raster in einer Datei).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Paket `tests/lasttest` im Hauptmodul bekommt die Hälfte des Werkzeugs, die ohne Binary und ohne Last auskommt: die Profile `abnahme` und `rauch` mit den Werten von `LH-QA-01.a`, den deterministischen Referenzdienst (ein HTTP-Server im Prozess mit 5 EntitySets zu je 2000 Entities, Bodies bis 60 KiB), den Aufbau (10.000 Aufnahmen über `bootstrap.Build` im Modus `record`, danach Import und Workspace aus der Baseline, mit Frist), die Listen der Anfragen für `replay` und `mock`, die Umgebung des Binary und die Schnittstelle der Uhr (`SPEC-216` Nr. 4 c, 5, 6, 7, 9 a, 14 a). Der Teil schreibt nur Dateien unter `tests/lasttest/` und liest `src/` (`bootstrap.Build`, `sqlite`), ohne es zu ändern; er braucht keine Zeile des Binary und keinen Einbau-Slice. **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-23): die Schätzung des ungeschnittenen Plans (*S* = 6055, Marke 8200, 34 Zeilen) liegt weit über der Schnittgrenze von `SPEC-128` Nr. 1; drei Teile, die nacheinander laufen und je einzeln lieferbar sind: **1** (diese Datei, F1 bis F10), **2** [slice-lasttest-auswertung](slice-lasttest-auswertung.md) (Leser und Auswertung, A1 bis A13) und **3** [slice-lasttest](slice-lasttest.md) (Lauf: Takt, Prozess, Treiber, Gegenproben, L1 bis L11; der Teil trägt die Kennung des ungeteilten Slice, Abweichung vom Namensvorschlag `slice-lasttest-lauf`, damit jede Zeile „`slice-lasttest` in `done/`“ in den Plänen davor und danach „alle drei Teile in `done/`“ heißt, weil Teil 3 als letzter startet und die zwei anderen voraussetzt).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Leser der Span-Zeilen, Phasen, p95, Anteil, Last-Proben, Zielvergleich, Streuung, Urteil, Exit und Ausgabe — [slice-lasttest-auswertung](slice-lasttest-auswertung.md) (Teil 2): sie brauchen die Uhr und das Profil dieses Teils, nicht seinen Aufbau.
- Der Takt und der Lastgenerator, der Prozess des Binary, Ablauf und Abbruch, der Treiber `tools/harness/lasttest.sh`, die Stufe `lasttest` des `Dockerfile`, das Fragment `harness/mk/lasttest.mk`, die Gegenproben, die Zeile in `harness/README.md` und das Hauptprogramm — [slice-lasttest](slice-lasttest.md) (Teil 3): nur Teil 3 ändert die Gate-Kette (Übergangsregeln Nr. 7). <!-- d-check:ignore (geplante Datei, entsteht mit dem Code) -->
- Eine Admin-Schnittstelle zum Aufbau der Last (Import, Workspace) — `welle-admin-api`; die Fixtures entstehen im Programm über `bootstrap.Build` im Modus `record` (Aufnahme der Anfragen eines Referenzdienstes), `Import().Import` und `Workspaces().Create` (`SPEC-216` Nr. 6 a), wie der Übergang von `SPEC-037` Nr. 6.
- Der Einbau der Spans — `slice-telemetry-einbau-gateway`, `slice-telemetry-einbau-replay`, `slice-telemetry-einbau-mock` und `slice-telemetry-einbau-mock-mutation` (alle in `done/`); dieser Teil liest keinen Span, die Zeile des Adapters liest Teil 2.
- Latenz im Record- und im Pass-through-Betrieb — `LH-QA-01` sagt dort keine Latenz zu; der Aufbau läuft im Modus `record` und misst nichts.
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

- [ ] Profile, Referenzdienst und Listen: die Profile `abnahme` und `rauch` mit den Werten von `LH-QA-01.a` (10.000 Exchanges einer Session: `$metadata`, 200 Seiten, 9799 Lookups, Bodies bis 64 KiB; ein Workspace aus 10.000 Entities in 5 EntitySets), erzeugt deterministisch aus einem Startwert; die Listen `replay` und `mock` (Mischung 70 % Einzel-Lookups, 20 % Collection mit `$filter` und `$top`, 10 % Collection mit `$top`) und die Umgebung des Binary (Zeilen F1 bis F3, F7 bis F9 der Tabelle in §3, vom Architect; `LH-QA-01.a`, `SPEC-216` Nr. 5, 6 b, 6 c, 7, 9 a).
- [ ] Aufbau und Uhr: der Aufbau über `bootstrap.Build` im Modus `record`, Import und Workspace, mit Frist (30 min, `rauch` 5 min) und Prüfung des Verzeichnisses `/daten`; der Match-Key-Index ist durch `make schema-check` und `make migration-check` belegt, das Werkzeug erzeugt seine Datenbank durch die Migration des Adapters; jede Zeitgrenze liest die Uhr über die Schnittstelle (Zeilen F4 bis F6 und F10; `SPEC-004`, `SPEC-216` Nr. 4 c, 6 a, 6 d bis 6 g, 14 a).
- [ ] `make gates` grün; der Teil ändert weder Binary noch Gate-Kette, die Kette prüft das Paket mit `make test`, `make test-race`, `make a-check` und `make lint` (`SPEC-216` Nr. 1 und 14).
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-216` steht vor dem Start auf `main`; dieser Teil schreibt keine Spezifikation und keinen Vertrag).
- [ ] Handbuch: der Teil führt kein sichtbares Verhalten des Binary ein (ein Werkzeug außerhalb des Binary, nichts am Binary); das Kataloge-Gate ist grün, und §7 sagt, dass das Binary unverändert blieb.
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
| `tests/lasttest/` | neu | das Paket im Hauptmodul (Schicht `tests`, [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)): Schnittstelle der Uhr, Profile, Generator aus dem Startwert, Referenzdienst, Aufbau über `bootstrap.Build`, Listen, Umgebung; im Hauptmodul statt unter `tools/`, weil es den Composition Root und den Adapter `sqlite` braucht (die Module unter `tools/` haben keine Abhängigkeit) und `make test`, `make test-race` und `make a-check` es dann prüfen (`SPEC-216` Nr. 14 e); die Dateien nennt der Implementer im Block, sobald er sie anlegt (§4) |
| Tests | neu | die Fälle der Zeilen F1 bis F10, Mengen als Schleifen, Fake-Uhr (`SPEC-216` Nr. 14) |
| `harness/vertraege/` | geliefert (Architect, 2026-11-23) | `SPEC-216` und die Zeile in der Liste der Verträge ([ADR-0004](../../adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)); dieser Plan schreibt sie nicht |

Dieser Teil ändert weder `Dockerfile` noch `harness/mk/` noch `harness/README.md` noch `tools/gegenproben/` noch den Stempel: er liegt außerhalb der Gate-Kette (Übergangsregeln Nr. 7) und darf neben jedem Slice laufen, der sie ändert.

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstanden mit dem Vertrag **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidungen des Architect** (2026-11-23; die Aufträge des ungeschnittenen Plans mit ihrem Ort in [SPEC-216](../../../../harness/vertraege/SPEC-216-lasttest.md), soweit dieser Teil sie trägt; der Vertrag trägt die Gründe und die verworfenen Alternativen, dieser Plan wiederholt sie nicht, `AGENTS.md` §3.8):

1. **Reproduzierbarkeit** (Nr. 6 c, 11): Startwert 1 (Vorgabe), feste Permutationen aus `math/rand/v2` mit PCG; die Grenzen des Containers und die zwei Läufe je Betriebsart trägt Teil 3, die Toleranz und die Last-Proben Teil 2.
2. **Listen** (Nr. 7): Replay: die 10.000 Anfragen des Aufbaus in einer Permutation, neue Replay-Session alle 100 Anfragen, Strategie `sequential` (Vorgabe); Mock: 70 % Einzel-Lookup, 20 % Collection mit `$filter` (`Gruppe eq <g>`) und `$top=10`, 10 % Collection mit `$top` (10 oder 50), fünf EntitySets; der offene Takt dazu ist Teil 3.
3. **Fixture-Aufbau** (Nr. 6): ein Aufbau je Container für beide Betriebsarten; Aufnahme der 10.000 Anfragen eines Referenzdienstes (`$metadata`, 200 Seiten, 9799 Lookups) über `bootstrap.Build` im Modus `record`, danach Import und Workspace aus der Baseline; Bodies bis 60 KiB (200 B, 8 KiB, 60 KiB im Verhältnis 7 : 2 : 1); Daten in der Schreibschicht des Containers, mit `--rm` gelöscht, `daten_mib` in der Ausgabe.
4. **Tabelle, Raster, Produktcode-Größe:** unten und in §6; die Größe in §4.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-23, Zeilen F1 bis F10 der Tabelle des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; jede Zeile steht in genau einem der drei Teile; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). `Nr.` ist die Kennung des ungeschnittenen Plans: **F** gehört zu Teil 1 (Fixtures, Aufbau, Uhr), **A** zu Teil 2 (Leser und Auswertung), **L** zu Teil 3 (Lauf: Prozess, Last, Treiber). *Fake-Uhr* ersetzt die Schnittstelle der Zeit (`SPEC-216` Nr. 14). Die Tabelle trägt die Spalten, die `SPEC-128` Nr. 3 a liest; `Nr.` bezieht sich auf `SPEC-216`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| F1 | (Nr. 5, 6 d) Die Profile tragen die Werte von `LH-QA-01.a` (`abnahme`: 10000 Exchanges, 5 × 2000 Entities, 20/s, 60 s und 600 s, 2 Läufe) und die Formel geht auf: Exchanges = 1 + Seiten + Lookups, Lookups mindestens 1 und höchstens Sets × Entities | `TestProfile`: Literal je Wert beider Profile (11 Werte × 2 gleich 22) und die Rechnung für beide (Seiten 200 und 10, Lookups 9799 und 289); Zählprüfung 22 und 2 | Rate 20 auf 10; Messphase 600 auf 60; 5 Sets auf 4; Lookups 9800 (Summe 10001) | andere |
| F2 | (Nr. 6 c) Gleicher Startwert ergibt byte-gleiche Antworten des Referenzdienstes und gleiche Listen, ein anderer eine andere Permutation; kein Wert hängt von Zeit, Prozess oder Reihenfolge einer Map ab | `TestDeterminismus`: ein Hash über alle Antworten (`$metadata`, 200 Seiten, 10000 Entities) und beide Listen, zweimal gebildet gleich für die Startwerte 0, 1 und 100000000000000000; Startwerte 1 und 2 verschieden; die Permutation trägt jedes der 10000 Paare genau einmal (Schleife); Literal: die ersten fünf Paare für Startwert 1; `go/parser`: kein Import von `math/rand` und kein Aufruf des globalen Generators | globaler Generator; `time.Now` im Startwert; eine Map in der Liste ohne Sortierung | andere |
| F3 | (Nr. 6 b) Menge und Größe des Referenzdienstes: 5 Sets × 2000 Entities, `Gruppe` gleich `Id` modulo 20, `Nutzlast` zu 7000 Entities 200 B, 2000 Entities 8 KiB, 1000 Entities 60 KiB, kein JSON-Body über 64 KiB, Seiten zu 50 mit `@odata.nextLink` bis zur letzten, kein Name aus den Defaults von `REDACT_*` | `TestReferenzdienst`: Schleife über alle 10000 Entities: Zahl, Größe ≤ 65536, Klassen 7000 · 2000 · 1000 (Literal), je Gruppe 100 je Set; 200 Seiten (40 je Set), die letzte ohne `@odata.nextLink`; Schleife über die Namen der Defaults (`SPEC-110`): keiner im Dienst; Zählprüfung 10000, 200, 20 Gruppen × 5 Sets | Gruppe `Id` modulo 10; Nutzlast 70 KiB; `@odata.nextLink` auch an der letzten Seite; Seite zu 51 | andere |
| F4 | (Nr. 6 a, d, e, f) Der Aufbau liefert genau so viele Exchanges, Entities und eine Baseline `complete`, wie das Profil sagt; ein Status ≠ 200, ein Fehler von Import oder Workspace oder eine Abweichung ist `Aufbau gescheitert` ohne Läufe | `TestAufbau` gegen ein `Handle` im Prozess (`bootstrap.Build`, ein Profil des Tests: 2 Sets × 20 Entities, Seite 10): Gutfall; Seite 2 liefert 500; eine Seite weniger (Zahl ≠); falscher `MetadataHash`; Baseline `partial` (die Seite ohne Fortsetzung); Orakel: Zahl der Exchanges, Entities je Set, Status der Session, Fehler `aufbau` und kein Lauf; Zählprüfung 5 Fälle | Zählprüfung entfällt; Status ≠ 200 gilt als Erfolg; Hash aus anderen Bytes; `partial` gilt als Erfolg | andere |
| F5 | (Nr. 6 e, 4 a) Der Aufbau legt nicht über vorhandene Daten: `/daten` besteht nicht oder ist leer, sonst `DATA_DIR nicht leer`; Daten bleiben unberührt | `TestAufbauDatenverzeichnis`: 4 Formen (fehlt, leer, mit einer Datei, ist eine Datei); Orakel Fehler oder Aufbau; die Datei byte-gleich nach dem Fehler | Prüfung entfällt; die Datei wird überschrieben | andere |
| F6 | (Nr. 4 c, 6 f, 5) Die Frist des Aufbaus gilt (30 min, `rauch` 5 min); über ihr `frist`, das `Handle` ist heruntergefahren | `TestAufbauFrist`: Fake-Uhr springt über die Frist, während der Referenzdienst eine Antwort hält; Fehler `frist` genau einmal, `Herunterfahren` genau einmal; 2 Profile gleich 2 Fälle, je mit der Frist des anderen als Gegenprobe (keine Frist bei 29 min) | Frist nicht geprüft; Frist des Profils `rauch` für beide | andere |
| F7 | (Nr. 7 a, c) Liste `replay`: jede Anfrage des Aufbaus einmal je Durchgang in der Permutation, drei Header, `Accept` wie in der Aufnahme, neue Session alle 100 Anfragen, in einer Session keine Anfrage zweimal, Sessions je Lauf verschieden | `TestListeReplay`: Schleife über 13200 Anfragen (`abnahme`): Zählung je Element (10000 genau einmal im ersten Durchgang, 3200 zweimal); Session ⌊*i*÷100⌋ an den Grenzen *i* = 99, 100, 101, 9999, 10000; Schleife über 132 Sessions: kein Element doppelt; Session-IDs von Lauf 1 und Lauf 2 ohne Schnitt; Zählprüfung 13200 und 132 | Wechsel alle 101; gleiche IDs in beiden Läufen; `Accept` fehlt; Header der Session fehlt | andere |
| F8 | (Nr. 7 b, c) Liste `mock`: Art *i* modulo 10 (70 · 20 · 10 %), 5 Sets, `Id` 1 bis 2000, *g* 0 bis 19, *n* 10 oder 50, aus dem Startwert, Query prozent-kodiert | `TestListeMock`: Schleife über 13200: Arten 9240 · 2640 · 1320 (Literal), 15 Zellen Art × Set je mindestens 1; Wertebereiche (Startwert 1, Literal): `Id` 1 und 2000, *g* 0 und 19, *n* beide Werte erreicht; `$filter=Gruppe%20eq%20<g>`; `rauch` (260): 182 · 52 · 26; Zählprüfung 13200 und 260 | `Id` bis 1999; Art *i* modulo 9; *g* 1 bis 20; Query ohne Kodierung | andere |
| F9 | (Nr. 9 a) Umgebung des Binary: genau fünf Schlüssel (`MODE`, `LISTEN`, `DATA_DIR`, `SERVICE_ID`, `SERVICE_ROOT_PATH`), `MODE` je Betriebsart, kein `REPLAY_STRATEGY`, keine `UPSTREAM_URL` | `TestUmgebung`: Literal je Betriebsart (2): Schlüssel und Werte, Zahl 5; kein weiterer Schlüssel; Zählprüfung 2 × 5 | `REPLAY_STRATEGY` gesetzt; `UPSTREAM_URL` gesetzt; `MODE` für beide gleich | andere |
| F10 | (Nr. 14 a) Jede Zeitgrenze liest die Uhr über die Schnittstelle: kein `time.Now`, `time.Sleep`, `time.After`, `time.NewTimer`, `time.NewTicker`, `time.AfterFunc` und kein `context.WithTimeout` außerhalb der Datei der Uhr; kein Test schläft länger als eine Sekunde | `TestPaketOhneZeit` (`go/parser` über alle Dateien des Pakets, auch die späteren Teile): 7 Namen; Orakel: kein Treffer außerhalb der Uhr; die Datei der Uhr trägt sie (Kontrolle) | `time.Now()` im Takt; `time.After` in der Frist; `context.WithTimeout` im Aufbau | andere |

Zeilen: F1 bis F10 gleich 10 (*a* = 0, *b* = 10): keine Zeile ist nebenläufig (die Klasse setzt der Architect, `SPEC-128` Nr. 3 e); Größe und Schnitt: §4. F10 prüft alle Dateien des Pakets, auch die der Teile 2 und 3: die Prüfung läuft ab diesem Teil und färbt jede spätere Datei, die eine der sieben Funktionen der Zeit außerhalb der Uhr nennt.

**Mengen mit Zählprüfung** (Schleife oder Generator, Orakel als Literal im Test, nie aus dem Code; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15): die Zeilen nennen in der Spalte *Fänger* ihre Menge und ihre Zählprüfung; hier die Achsen der größeren.

| Menge | Achsen und Elemente | Orakel | Zählprüfung gegen die Quelle |
|---|---|---|---|
| Fixture (F2, F3) | 5 Sets × 2000 `Id` gleich 10000 Entities; Gruppen 20 × 5 Sets; Nutzlast-Klassen 3; Seiten 40 je Set | Größe, Klasse, Gruppe, `@odata.nextLink` | 10000 · 200 · 100 Entities je Gruppe und Set · 7000 : 2000 : 1000 gleich `SPEC-216` Nr. 5 und 6 b |
| Liste `replay` (F7) | Anfrage *i* (13200) × Session (132) × Durchgang (2) × Lauf (2) | Element, Session, keine Doppelten | 13200 · 132; 10000 + 3200 |
| Liste `mock` (F8) | Art (3) × Set (5) × Parameter (`Id` 2000, *g* 20, *n* 2) | Art je *i* modulo 10; Bereiche | 13200 gleich 9240 + 2640 + 1320; 15 Zellen |

**Zusatztabelle: Paare von Verstößen und Modi** (`AGENTS.md` §3.16; die Zeilen der Zusatztabelle des ungeschnittenen Plans, soweit Fänger dieses Teils sie tragen; die Fänger der Teile 2 und 3 stehen in deren Plänen):

| Paar oder Modus | Erwartung | Fänger dieses Teils |
|---|---|---|
| Betriebsart `replay`, `mock`, `beide` | die Listen (F7, F8) und die Umgebung (F9) unterscheiden sich; `mock` hat kein `$metadata` in der Liste; der Ablauf (L7, Teil 3) und die Zeile `Abnahme` (A10, Teil 2) unterscheiden sich ebenfalls | F7 bis F9 |
| Profil `abnahme`, `rauch` | alle Werte aus dem Profil; ein Lauf ohne Streuung (A11, Teil 2), nie `abnahme=ja` (A10, Teil 2) | F1 |

**Bestehende Fänger, die der Vertrag nennt** (keine Zeile dieses Plans): der Index `(session_id, match_key)` ohne Tabellenscan (`make schema-check`, `tools/schema/check.py`, `EXPLAIN QUERY PLAN`) und die Gleichheit von Migration und DDL (`make migration-check`).

**Akzeptierte Negative dieses Teils** (Grund, Folge und Maß; ausführlich in `SPEC-216` Nr. 16): (1) Setzungen ohne Messung (Nr. 16 g): die Seitengröße 50 und die Größenklassen der `Nutzlast`, 2 von 7 Setzungen des Vertrags, 0 von 2 gemessen; die Folge im Betrieb und der Trigger der Prüfung (der Architect liest die drei ersten gültigen Läufe der Abnahme) stehen in [slice-lasttest](slice-lasttest.md) §3, dort mit den übrigen fünf. (2) Dauer und Größe des Aufbaus (Nr. 16 e): die Frist von 30 min ist die einzige Grenze; die Dauer und `daten_mib` sind ungemessen, die Zahlen liefert der Handlauf von Teil 3 (§6 dort); Folge: ein Aufbau über der Frist ist `frist`, Exit 3, und die Frage geht an den Architect.

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: die Mischung der Anfragen, die Fixtures der Menge), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Eine Zahl aus einem Lauf nennt Zahl der Prozesse, Last und Messenden; „stabil“ oder „schwankt wenig“ steht nur mit Zahl (`AGENTS.md` §3.10).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` und `slice-bootstrap-m3` (der Aufbau ruft `bootstrap.Build` im Modus `record` und `Workspaces().Create`) liegen in `done/` (geprüft am 2026-11-23: beide). Kein Einbau-Slice ist Bedingung: der Teil liest keinen Span und keine Zeile des Binary (Architect, 2026-11-23). Der Vertrag des Werkzeugs samt Randformen liegt in `harness/vertraege/` ([SPEC-216](../../../../harness/vertraege/SPEC-216-lasttest.md)), Tabelle und Raster in §3 und §6, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Teil 2 ([slice-lasttest-auswertung](slice-lasttest-auswertung.md)) startet, wenn dieser Teil in `done/` liegt, Teil 3 ([slice-lasttest](slice-lasttest.md)) nach Teil 2.

**Warum:** Teil 2 ruft die Uhr und das Profil dieses Teils (*Code-Abhängigkeit*), Teil 3 den Aufbau, die Listen und die Umgebung (*Code-Abhängigkeit*); die Reihenfolge 1, 2, 3 folgt den Schnittstellen.

**Größenmarke** (`SPEC-128` Nr. 1 und 10 g, `AGENTS.md` §3.14; Teil des Architect 2026-11-23 im Block unten; der Planner hat die Zahlen am 2026-11-23 aus der Tabelle nachgerechnet, **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 10 Zeilen der Tabelle (F1 bis F10), *a* = 0, *b* = 10, *f* = 1 (Fake-Uhr), *p* = 1000 (Erwartung des Architect für den Code außerhalb der Tests: Uhr 50, Profil 80, Generator 80, Referenzdienst 300, Listen 160, Umgebung 40 und Aufbau 290; die Erwartung ist geschätzt; am Ende maßen csdl 1506, json 1640 und der Import 1802 Zeilen Produktcode, `SPEC-128` Nr. 2): *S* = 85·0 + 80·10 + 200·1 + 1000 = 2000, 135·2000 = 270000, **Marke 2700** = 100 · ⌈27⌉ (*S* = 2000 bleibt auf diesem Wert, `SPEC-128` Nr. 1); keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`; der Teil liegt 223 unter der Schnittgrenze, jede weitere Zeile kostet 80, die dritte weitere Zeile *b* löst die Schnittfrage aus (*S* = 2240). Summe der drei Teile: 2000 + 1990 (Teil 2) + 2065 (Teil 3) = 6055, gleich dem ungeschnittenen Plan; keine Abweichung vom Vorschlag des Architect bei den Zahlen. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

```groessenmarke
pfad: tests/lasttest/
produktcode: 1000
fakes: 1
```

Der Block nennt das Verzeichnis, nicht die Dateien: das Paket ist neu, und die Namen der Dateien stehen vor dem Code nicht fest (der Vorschlag des Architect nennt Bausteine, keine Dateien). Das Verzeichnis gehört den drei Teilen allein (geprüft am 2026-11-23: kein Plan unter `open/`, `next/` und `in-progress/` und kein Zweig schreibt unter `tests/lasttest/`), und sie laufen nacheinander: die Basis des Teils 2 und 3 liegt hinter dem Merge des vorigen, die Zeilen des vorigen zählen dort nicht. Fremde Zeilen (`SPEC-128` Nr. 10 g) kann die Messung nur zählen, wenn ein vierter Schreiber unter diesem Verzeichnis entsteht; der Implementer nennt jede Datei, die er anlegt, im selben Commit im Block (`BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`), sobald sie besteht.

**Kollisionen** (Architect, 2026-11-23, gelesen an den Plänen unter `open/` und `in-progress/` und an `git worktree list`; vom Planner beim Schnitt übernommen; `SPEC-128` Nr. 10 g: die Messung zählt jede Änderung an den Pfaden zwischen `basis` und `HEAD`): dieser Teil legt nur Dateien unter `tests/lasttest/` an und liest `src/` (`bootstrap.Build`, `sqlite`), ohne es zu ändern; er berührt weder `Dockerfile` noch `harness/mk/` noch `harness/README.md` noch `tools/gegenproben/` und kollidiert mit keinem Slice, auch nicht mit `slice-telemetry-einbau-record` (ändert `src/bootstrap/wurzel.go`, `src/adapters/driven/telemetry/` und Tests), den Teilen von `slice-telemetry-einbau-hybrid` und denen von `slice-bootstrap-m6` (`src/bootstrap/`); er darf vor dem Merge dieser Slices starten, ruft aber `bootstrap.Build` und muss nach jedem Merge von `main` übersetzen (ein Fehler der Naht wäre der Rückführung `blockiert`).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein weiterer Schnitt ist nicht vorgesehen, ein Teil, der die Schnittgrenze nach dem Code überschreitet, wird mit Messung und Grund berichtet (`AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): `bootstrap.Build` im Modus `record` oder `Workspaces().Create` ändern ihre Schnittstelle so, dass der Aufbau nicht ohne Eingriff in `src/` zu schreiben ist (§6).

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

- Der Schnitt vor dem Code (Schätzung des ungeschnittenen Plans 6055, Marke 8200, über der Schnittgrenze von 2223): der Architect schlägt drei Teile vor (Fixtures, Auswertung, Lauf) — **Ausgang:** entschieden (Planner, 2026-11-23, nach dem Vorschlag des Architect): Schnitt in drei Teile mit den Marken 2700 (dieser Teil), 2700 und 2800, nachgerechnet aus der Tabelle (keine Abweichung bei den Zahlen); Teil 3 trägt die Kennung `slice-lasttest` statt `slice-lasttest-lauf`; `basis`, `schaetzung` und `marke` setzt der Planner je Teil beim Start (§4).
- Ein Paket `main` ohne Funktion `main` in den Teilen 1 und 2 (das Hauptprogramm gehört zu Teil 3, Erwartung des Architect: 180 Zeilen dort): ob `make test` (`gofmt`, `go vet`, Tests), `make lint` und `make a-check` ein solches Paket tragen, ist ungeprüft — **Ausgang:** offen; Adresse: Architect (vor dem Start von Teil 1: ein leeres `main` kostet Zeilen und gehört dann in den Block, oder der Vertrag nennt das Paket in Teil 1 und 2 nicht `main`).
- Der Aufbau über `bootstrap.Build` im Modus `record` ruft einen Port, dessen Fake dieser Teil nicht liefert (der Referenzdienst ist ein echter HTTP-Server im Prozess); ändert ein Slice der Welle die Schnittstelle von `Handle` oder `Import()`, bricht der Aufbau — **Ausgang:** entschieden (Planner, 2026-11-23): die Slices der Welle ändern `src/bootstrap/` weiter (`slice-telemetry-einbau-record`, die Teile von `slice-telemetry-einbau-hybrid` und `slice-bootstrap-m6`); der Test des Aufbaus (F4) fährt `bootstrap.Build` mit einem Profil des Tests und bricht bei einer Änderung der Schnittstelle zuerst; ein Zweig, der sie ändert, trägt den Nachzug in `tests/lasttest/` (Eintrag in §6 seines Plans mit Adresse Planner, `AGENTS.md` §3.13).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster steht unten, jede Zelle mit Ort in `SPEC-216` oder `gilt nicht` mit Grund — **Ausgang:** entschieden (Architect, 2026-11-23).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23; die Zeile dieses Teils; die Zeilen der Teile 2 und 3 stehen in deren Plänen; jede Zelle nennt ihren Ort in `harness/vertraege/` oder `gilt nicht` mit Grund):

| Operation | fehlender Wert (Zielwert, Startwert, Binary) | leerer Wert (leere Fixture, keine Antwort) | unbekannter oder fremder Wert (Fehlerantwort, Zeitüberschreitung, Zielwert 0) | Abbruch (Lauf vorzeitig beendet) | Wiederholung (zweiter Lauf, andere Last) | Betriebsart (Replay, Mock) |
|---|---|---|---|---|---|---|
| Fixtures erzeugen (Aufbau, Referenzdienst, Listen) | `SPEC-216` Nr. 2 (Startwert ungesetzt: 1), Nr. 6 e (`/daten` fehlt: wird angelegt); ein Binary braucht der Aufbau nicht (Nr. 6 a: `bootstrap.Build` im Prozess) | `SPEC-216` Nr. 6 e: `/daten` nicht leer ist `Aufbau gescheitert`; eine leere Fixture gibt es nicht: Nr. 5 verlangt mindestens einen Lookup (`gilt nicht`: das Profil ist fest) | `SPEC-216` Nr. 6 f: Status ≠ 200, Fehler von Import oder Workspace, Zahl ≠ Profil, `partial`; Startwert außerhalb der Form: Nr. 2 | `SPEC-216` Nr. 6 f (Frist des Aufbaus: Exit 3, `frist`) und Nr. 4 d (Signal, Exit 2) | `SPEC-216` Nr. 6 c: gleicher Startwert, gleiche Bytes; ein zweiter Aufbau auf denselben Daten ist Nr. 6 e | `SPEC-216` Nr. 6 a: ein Aufbau dient beiden Betriebsarten (die Betriebsart ändert den Aufbau nicht) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-23; der Planner zieht Treffer und Zähler beim Start nach). Treffer: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für Testhilfen, die `ctx` und Ströme des Referenzdienstes umhüllen; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Teil zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Keiner der offenen erreicht mit diesem Teil 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
