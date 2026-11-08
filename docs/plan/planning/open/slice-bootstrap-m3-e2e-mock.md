# Slice slice-bootstrap-m3-e2e-mock: Szenarien des Modus mock am gebauten Binary

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-12`, `LH-FA-09`, `LH-RB-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-037` Nr. 2, 3, 5 bis 7 und 10 (Vertrag von `make test-integration`, `harness/vertraege/`) · `SPEC-139` Nr. 11 d und f (Neustart im Modus `mock`) · `SPEC-049` Nr. 7 a und 10 b (Neustart über die Prozessgrenze, Negativ der Snapshots) · `SPEC-110` Nr. 8 und 14 (Default `mock`, `PUBLIC_BASE_URL` aus `LISTEN`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-08.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung**; jeder Ausschluss trägt
seinen Grund.

**Ziel:** Das gebaute Binary im Modus `mock` ist als Stufen von `make test-integration` belegt: *Mock CRUD → Reset* (`TestE2EMock`) und *Mutation → `SIGTERM` → Start auf demselben `DATA_DIR` → Mutation lesbar* (`TestE2EMockNeustart`), dazu der Befehl `reset` des Übergangs und je Stufe die drei Zeilen im Treiber (`SPEC-037` Nr. 3).

**Übernimmt (Sendung von `slice-bootstrap-m3`, Teil E des vor dem Code geschnittenen Plans, Planner, 2026-11-08; ursprünglicher Geber `slice-mock-abnahme`, `SPEC-049` Nr. 7 a und 10 b, Rückgabe R4):** den Neustart über die Prozessgrenze im Modus `mock` für Mutationen und das Szenario *Mock CRUD → Reset*; die Zeilen E01 bis E09 der Tabelle in §3 (Entscheidung des Architect, 2026-11-08: Stufen `TestE2EMock` und `TestE2EMockNeustart`, Randformen in `SPEC-139` Nr. 11 d und f und `SPEC-037` Nr. 7 und 10 i und j). Kein Ausschluss dieses Slice trifft die Sendung. Der Geber trägt den Eintrag in seinem §1 erst nach diesem (`AGENTS.md` §3.13).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Verdrahtung des Modus `mock` im Composition Root, Konfiguration, M3-Test in `tests/integration/`, Handbuch und der Fall `leer` von `TestE2EProzess` — `slice-bootstrap-m3`; ohne dessen Binary gibt es nichts, was die Stufen fahren (Reihenfolge, kein geteilter Mechanismus: E fährt das Binary von außen).
- Snapshots über die Prozessgrenze — es gibt bis `welle-admin-api` keinen Lesepfad dafür; akzeptiertes Negativ in `SPEC-049` Nr. 10 b und `SPEC-037` Nr. 10 j (E09).
- `SIGTERM` während einer Mutation am Binary — akzeptiertes Negativ in `SPEC-037` Nr. 10 i (E08); getrennt belegt durch `SPEC-136` Nr. 15 und `SPEC-192` Nr. 7.
- Änderungen an `Makefile`, `Dockerfile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und am Stempel — Nr. 7 der Übergangsregeln; die Stufe `integration` baut `./tests/e2e` und `./tests/e2e/uebergang` als Ganzes. Braucht ein Szenario doch eine Zeile dort, ist das eine Rückgabe an den Planner, kein Code dieses Slice.
- Modus `hybrid` und Szenarien von `welle-hybrid` — `slice-bootstrap-m4`.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**.

- [ ] Die Stufen `TestE2EMock` und `TestE2EMockNeustart` in `tests/e2e/` samt den drei Zeilen je Stufe in `tools/harness/test-integration.sh` und der Befehl `reset` des Übergangs in `tests/e2e/uebergang/` mit Test; `make test-integration` fährt beide grün (`SPEC-037`, Zeilen E01 bis E07 der Tabelle).
- [ ] Das Abdeckungs-Dokument `docs/plan/planning/e2e-abdeckung.md` ist aus `make e2e-abdeckung` neu erzeugt und führt die beiden Stufen; die Negative E08 und E09 stehen mit Grund und Folge in der Spezifikation.
- [ ] `make gates` grün; `make test-integration` als Pflichtlauf vor der Closure grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel: entfällt, kein Ergebnis dieses Slice ist am Binary für einen Anwender neu (Stufen und `reset` sind Prüfwerkzeug); das Handbuch-Gate bleibt grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tests/e2e/` | neu | Stufen `TestE2EMock` und `TestE2EMockNeustart`; Befehl `reset` des Übergangs in `tests/e2e/uebergang/` mit Test (`lauf`) |
| `tools/harness/test-integration.sh` | update | drei Zeilen je Stufe (`SPEC-037` Nr. 3), Stufe 8 und 9 |
| `docs/plan/planning/e2e-abdeckung.md` | update (erzeugt) | entsteht aus `make e2e-abdeckung` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster (§6) hat der Architect am 2026-11-08 vor dem ersten Code-Commit geschrieben (Zeilen E01 bis E09 des Plans von `slice-bootstrap-m3`, hier übernommen; die Randformen sind in der Spezifikation entschieden, `AGENTS.md` §3.12). Die Klasse `nebenläufig` trägt E04 (Beenden und Start des Servers); jede andere Zeile, auch ein akzeptiertes Negativ, ist `andere` (`SPEC-128` Nr. 3 e). Der Implementer sieht die Mutation vor der Übergabe rot und trägt die Raten, falls ein Fänger nur mit Wahrscheinlichkeit rot wird, in einen Absatz *Mutationen* dieser Sektion ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| E01 | `TestE2EMock`: Mock CRUD am gebauten Binary: Start im Modus `mock` ohne `UPSTREAM_URL` und `PUBLIC_BASE_URL`, `POST` `201` mit `Location` aus `LISTEN`, `GET`, `PATCH` `204`, `DELETE` `204`, `GET` `404` `ODR-E6008`, nach jedem Schritt der Zustand; die Collection ist `complete`; ein nicht unterstützter Ausdruck von `$filter` ist `501` (`SPEC-037` Nr. 7) | Stufe 8 des Treibers | das Binary lehnt `mock` ab; `Location` mit festem Host ⇒ rot | andere |
| E02 | Reset und Zuordnung am Binary: `SIGTERM`, `reset` des Übergangs, Start: die Collection gleicht der Antwort vor den Mutationen; Request ohne `X-Recorder-Workspace` ist `400` `ODR-E5003`, mit unbekanntem Wert `404` `ODR-E7001` | `TestE2EMock` | Reset ohne Wirkung ⇒ rot | andere |
| E03 | Kein Upstream: nach dem Stoppen des Referenzdienstes zählt er in keinem Schritt eine Verbindung oder einen Request mehr (`LH-FA-12`) | Zähler des Referenzdienstes in `TestE2EMock` und `TestE2EMockNeustart` | der Pfad des Mock greift auf einen Upstream zu ⇒ rot | andere |
| E04 | `TestE2EMockNeustart`: Mutationen (`POST`, `PATCH`, `DELETE` an drei Entities), `SIGTERM` (Exit 0), Start auf demselben `DATA_DIR`: die Zustände sind lesbar, ein zweiter Neustart ohne Änderung ebenso (`SPEC-049` Nr. 7, `SPEC-139` Nr. 11 f) | Stufe 9 des Treibers; die Antworten vor dem Signal sind das Orakel | der Start leert `DATA_DIR`; das Beenden schließt die Transaktion nicht ab ⇒ die Mutation fehlt, rot | nebenläufig |
| E05 | `DATA_DIR` besteht nicht: der Start legt es an und ist bereit, jeder Request mit Workspace ist `404` `ODR-E7001` (`SPEC-139` Nr. 11 f) | `TestE2EMockNeustart`, Boundary | das Anlegen entfällt ⇒ der Start scheitert ⇒ rot | andere |
| E06 | Befehl `reset` des Übergangs: fehlendes oder leeres Argument Exit 2, unbekannter Workspace Exit 1 mit `ODR-E7001`, sonst `revision=<Zahl>` (`SPEC-037` Nr. 6) | Test der Funktion `lauf` in `tests/e2e/uebergang/` (Schleife über die Argumentformen); gültiger Aufruf in `TestE2EMock` | Exit-Codes vertauscht ⇒ rot | andere |
| E07 | Stufen im Treiber: je Stufe drei Zeilen, der Name steht in `-test.list`, ohne `PASS` ist die Stufe rot; `make e2e-abdeckung` führt Stufe 8 und 9 (`SPEC-037` Nr. 2, 3) | ein Lauf von `make test-integration`; `docs/plan/planning/e2e-abdeckung.md` aus `make e2e-abdeckung` erzeugt | eine Anker-Zeile mit falschem Namen ⇒ `nicht gefunden`, Exit 2 | andere |
| E08 | **Negativ:** `SIGTERM` während einer Mutation am Binary; Grund und Folge in `SPEC-037` Nr. 10 i und `SPEC-139` *Ungeprüft* | kein Fänger am Binary; getrennt belegt: `SPEC-136` Nr. 15 (Frist) und `SPEC-192` Nr. 7 (Atomarität) | — | andere |
| E09 | **Negativ:** Snapshots über die Prozessgrenze; Grund und Folge in `SPEC-037` Nr. 10 j und `SPEC-049` Nr. 10 b | im Prozess: `SPEC-049` Nr. 7 d | — | andere |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `in-progress`): `slice-bootstrap-m3` liegt in `done/` (das Binary im Modus `mock` und der Nachzug von `SPEC-037` Nr. 5 bis 7 und 10 sowie `SPEC-139` Nr. 11 liegen damit in `main`). Beim Start liest der Planner §1 *Ausdrücklich NICHT* und die DoD der Nachbarn neu (`AGENTS.md` §3.13).

**Gate-Kette (Nr. 7 der Übergangsregeln):** Linie B. Dieser Slice ändert `tools/harness/test-integration.sh` und `tests/e2e/`, kein Pfad aus Nr. 7 (Architect, 2026-11-08, im Plan von `slice-bootstrap-m3` bestätigt: die Stufe `integration` baut `./tests/e2e` und `./tests/e2e/uebergang` als Ganzes). Braucht ein Szenario eine Änderung des `Dockerfile`, geht das als Rückgabe an den Planner. `slice-bootstrap-m4` und `slice-bootstrap-m6` führen je ein weiteres Szenario im selben Treiber; der Planner ordnet die Merges beim Start (Risiko in §6).

Größenmarke (Architect, 2026-11-08, `SPEC-128` Nr. 1; Planner-Rechnung gleich): *a* = 1 (E04), *b* = 8, *f* = 0, *p* = 40: *S* = 85 + 640 + 0 + 40 = 765, *M* = 1100, keine Schnittfrage. Der Block trägt den Teil des Architect; `basis`, `schaetzung` und `marke` setzt der Planner beim Start in eigenem Commit vor dem ersten Code-Commit (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `docs/` zählt nicht).

```groessenmarke
pfad: tests/e2e/
pfad: tools/harness/test-integration.sh
produktcode: 40
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt.
- `in-progress` → `open` (blockiert — Carveout?): ein Szenario braucht eine Zeile in `Dockerfile`, `Makefile` oder `harness/mk/` (Nr. 7) oder das Binary von `slice-bootstrap-m3` fehlt.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make test-integration` sind grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-08, aus dem Plan von `slice-bootstrap-m3` übernommen): jede Zelle nennt eine Nummer der Spezifikation oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Start auf vorhandenem `DATA_DIR` | `SPEC-139` Nr. 11 f und `SPEC-110` Nr. 11: ein fehlendes Verzeichnis wird angelegt, der Start gelingt | Nr. 11 f: eine Ablage ohne Baseline und ohne Metadaten ist `ODR-E7001` und `ODR-E6004` | Nr. 11 f: die Ablage eines früheren Laufs in einem anderen Modus wird übernommen | Nr. 11 d: eine Mutation beim Herunterfahren ist fertig oder hinterlässt nichts; am Binary ein Negativ (`SPEC-037` Nr. 10 i) | Nr. 11 f und `SPEC-049` Nr. 7 d: ein zweiter Neustart ohne Änderung ergibt dasselbe |
| Befehl `reset` des Übergangs | `SPEC-037` Nr. 6: ein fehlendes Argument ist Exit 2 | Nr. 6: ein leeres ebenso | Nr. 6: ein unbekannter Workspace ist Exit 1 mit `ODR-E7001` | gilt nicht: ein Hilfsprogramm des Tests, das Signal beendet den Prozess | `SPEC-129` Nr. 10: jeder Reset erhöht die Revision |

- Neustart über die Prozessgrenze im Modus `mock` (Sendung, `SPEC-049` Nr. 7 a): Randformen — **Ausgang:** entschieden (Architect, 2026-11-08): Verzeichnis fehlt wird angelegt (E05), Signal während einer Mutation ist ein akzeptiertes Negativ (E08), Snapshots bleiben das Negativ (E09).
- Reihenfolge der Merges im Treiber `tools/harness/test-integration.sh`: `slice-bootstrap-m4` und `slice-bootstrap-m6` fügen ebenfalls Stufen hinzu; Nummern der Stufen (8, 9) und das Abdeckungs-Dokument können beim Zusammenführen kollidieren — **Ausgang:** offen; Adresse: Planner (Start des Slice und Integrationsschritt).
- Die Stufe 9 beendet und startet den Server und liegt in der Klasse `nebenläufig` (E04) — **Ausgang:** offen; Muster der Fänger wie bei den Stufen 1 bis 7 (Bereitschaft statt Schlaf), Raten trägt der Implementer in §3.

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register · `grundlagen-traceability.md` §Herkunfts-Anker für
Steering-Loop-Regeln (das Feld `liegt in` steht **nur**, wenn mit diesem Slice
wirklich etwas verkörpert wurde).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage nicht einzeln gesichtet; der Planner sichtet beim Start (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` ist einschlägig: die Stufen fahren das gebaute Binary).

Alle berührten Sub-Areas GF.
