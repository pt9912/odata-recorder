# Slice slice-test-contract: `make test-contract` — Referenzanfragen und Negativfälle am HTTP-Eingang

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-RB-01`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)

**Berührte Spec-Stellen:** `SPEC-042` (Vertrag `make test-contract`, `harness/vertraege/SPEC-042-test-contract.md`: *OData-V4-JSON, Metadata, Fehlerfälle und HTTP-Semantik*; Referenzanfragen und Negativfälle) · `SPEC-049` (Abnahmematrix, Fixtures `ODV4-*`, Nr. 3 und 11 a) · `LH-RB-01.a` (Nachweis: Referenzanfragen und Negativfälle) · `LH-QA-05.a` (Fehlerklassen) · `SPEC-109` (Tabelle der Codes, Fehlerobjekt) · `SPEC-136` Nr. 5 und 8 (Rahmung, Fehlerobjekt, Korrelations-ID) · `SPEC-195` Nr. 9 und `SPEC-196` Nr. 2 (Metadaten des Mock) · [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) (Schicht `tests`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-test-contract`, Worktree `/Development/odata-recorder-wt-test-contract`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-31.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make test-contract` (`SPEC-042`) fährt Referenzanfragen und Negativfälle gegen den HTTP-Eingang: die Fixtures der Muss-Zeilen aus `SPEC-049` (`ODV4-*`, geliefert von `slice-mock-abnahme`) und je Fehlerklasse von `LH-QA-05` (fehlender Replay-Treffer, ungültige Metadaten, nicht unterstützte Query-Option, unvollständige Datenabdeckung) einen Negativfall, der Fehlerklasse, Status und diagnostischen Kontext prüft, ohne ein Geheimnis offenzulegen (Messmethode von `LH-QA-05`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Neue Golden-Fixtures — `slice-mock-abnahme` schreibt sie; dieser Slice ruft sie auf und verschiebt sie nicht (ob die Fixtures dort bereits unter `make test` laufen und hier ein zweites Mal, ist §6).
- Zeilen der Spalte *Soll* und *Nicht in MVP* von `SPEC-049` (ETag, `Prefer`, `$expand`, V2) — die Folgewellen [welle-etag](../welle-etag.md), [welle-expand-batch](../welle-expand-batch.md) und [welle-odata-v2](../welle-odata-v2.md); sie hängen ihre Fälle an dieses Ziel, wenn sie es nutzen wollen (Entscheidung dort).
- Konformität gegen jede OData-Implementierung — Grenze von `SPEC-042`: geprüft wird gegen Referenzanfragen und Negativfälle.

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

- [ ] Das Ziel `make test-contract` (Treiber, Fragment, Stufe `contract` des `Dockerfile`) und der Vertrag `SPEC-042` (der Stand „entsteht mit dem Slice, der den Code liefert“ wird zum gelieferten Umfang); die Muss-Fixtures `ODV4-*` aus `slice-mock-abnahme` laufen über echtes HTTP, aufgerufen aus `abFixtures()` (Aufruf, keine Kopie, `SPEC-042` Nr. 3 a).
- [ ] Je Fehlerklasse von `LH-QA-05` ein Negativfall am HTTP-Eingang, der Klasse, Status und diagnostischen Kontext prüft und keine Geheimnisse offenlegt.
- [ ] Je Zusage eine Mutation, vom Implementer rot gesehen (`AGENTS.md` §3.10): eine Fehlerklasse, die auf eine andere abgebildet wird, färbt den Negativfall rot.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-042`, Sensors-Tabelle in `harness/README.md`).
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
| `harness/mk/test-contract.mk`, `tools/harness/test-contract.sh` | neu | Ziel `make test-contract` und Treiber (`SPEC-042` Nr. 2), Form von `test-determinism.mk` und `test-determinism.sh`; kein Eintrag in `GATE_CHECKS` ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) |
| `Dockerfile` | update | neue Stufe `contract` (zwei `go test -c`, `SPEC-042` Nr. 1); keine bestehende Stufe ändert sich |
| `src/bootstrap/vertrag_http_test.go` | neu | `TestVertragMatrix` und die Vergleichs- und Zähltests über `abFixtures()` und die Hilfen der Abnahme (`SPEC-042` Nr. 3, 4); Paket `bootstrap_test` (Entscheidung Nr. 7) |
| `tests/integration/vertrag_fehlerklassen_test.go` | neu | `TestVertragFehlerklassen`, `TestVertragFehlerobjekt` (`SPEC-042` Nr. 4 bis 6); auf `m3Aufbau` und den Hilfen von `m2_test.go`, Schicht `tests` (nennt Modell, Inbound-Ports, `sqlite`, Blob-Adapter; `bootstrap` ohne Kante, [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)) |
| `harness/vertraege/SPEC-042-test-contract.md` | update | Umfang, Randformen, Negative (Architect, vor dem Code); der Implementer zieht den Stand nach |
| `harness/README.md` | update | Sensors-Tabelle: Ziel, Vertrag, Bindung; erst mit dem Ziel (`AGENTS.md` §4: kein Target nennen, das nicht existiert) |

Kein Gate wird gelockert und keine Schicht der `.a-check.yml` ändert sich: die Kanten von `tests` bleiben `model`, `ports-inbound`, `adapters-driven` ([ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)); der HTTP-Lauf der Fixtures liegt im Composition Root (Paket `bootstrap_test`), der die Dienste nennen darf. Eine neue Entscheidung (ADR) ist nicht nötig: die Ablage steht in `SPEC-049` Nr. 1 und 11 a, das neue Ziel ist kein Gate nach `AGENTS.md` §3.6.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-08, vor dem ersten Code-Commit). Nr. in der Spalte *Ort* sind Nummern von `SPEC-042`. Die Klasse (`SPEC-128` Nr. 3 e) ist für jede Zeile `andere`: die Matrix und die Klassen laufen nacheinander, kein Fall prüft Nebenläufigkeit.

| Nr. | Zusage | Fänger | Mutation | Ort |
|---|---|---|---|---|
| T01 | Die Matrix läuft über HTTP für genau die 36 IDs und 92 Schritte von `abFixtures()`, jede einmal | `TestVertragMatrixZaehlung` (Tabelle) und `abPruefeZaehlung` am Lauf von `TestVertragMatrix` | die Schleife überspringt eine ID und, getrennt, den Schritt `b` einer ID ⇒ rot | Nr. 3 a, e |
| T02 | Header einer Erfolgsantwort: Menge und Werte gleich der Tabelle, ohne weiteren; allein `Date` und `Content-Length` zählen nicht | `TestVertragVergleich`, Schleife über Fälle: weiterer Header, fehlender, anderer Wert, doppelter Wert (je rot); nur `Date` anders, nur `Content-Length` anders (je grün) | die Ausnahmen um `Odata-Version` erweitert ⇒ der Fall „Header fehlt“ bleibt grün, rot | Nr. 3 d |
| T03 | Body byte-gleich, bis zum Ende gelesen | `TestVertragVergleich`: Zeilenende, Leerzeichen am Ende, gekürzt, anders (je rot) | der Vergleich schneidet Leerraum ab ⇒ Fall „Leerzeichen“ rot | Nr. 3 d |
| T04 | `drahtQuery` kodiert genau die Menge von Nr. 3 c und lässt jedes andere Byte unverändert, Hex in Großbuchstaben | `TestVertragDrahtQuery`, Schleife über alle 256 Byte-Werte (Menge kodiert, Rest gleich; ein Fall mit `%41` unverändert, einer mit leerer Query ohne `?`) | Leerzeichen nicht kodiert ⇒ rot; `%` kodiert ⇒ rot; Hex klein ⇒ rot | Nr. 3 c |
| T05 | Schritte mit `Strikt` gehen an den strikten Server, alle anderen an den lax konfigurierten | `ODV4-E-04` a (`422`, `ODR-E6003`) in `TestVertragMatrix`; `ODV4-GET-01` am lax Server (`200`) | der strikte Schritt geht an den lax Server ⇒ `200` statt `422`, rot | Nr. 3 b |
| T06 | Mutierende Schritte (`POST`, `PATCH`, `PUT`, `DELETE`) tragen Body und `Content-Type` wie in `SPEC-049` Nr. 2 | `ODV4-CRUD-01..06` in `TestVertragMatrix` (`201` mit `Location`, `204`, Fehler) | der Client sendet kein `Content-Type` ⇒ `ODV4-CRUD-01` rot (`ODR-E6007` statt `201`) | Nr. 3 b, c |
| T07 | Nach den Schritten einer ID gilt die Spalte *Revision*, gelesen an der Datei | Stand-Prüfung am Ende jeder ID in `TestVertragMatrix` | der Client wiederholt jede mutierende Anfrage ⇒ Revision falsch, rot | Nr. 3 e |
| T08 | Eine Fehlerantwort trägt `Content-Type: application/json`, `Content-Length` gleich der Länge, nichtleere `X-Correlation-Id`, genau die drei Felder in Reihenfolge mit `code`, `target`, `message` des Codes | `TestVertragFehlerobjekt`: Schleife über die Eigenschaften der Prüffunktion (Content-Type, Länge, ID leer, viertes Feld, Reihenfolge, `target`, `message`, Code außerhalb von `SPEC-109`) mit je einem verletzten Beispiel; die Prüfung läuft an jedem Fehlerschritt der Matrix und jedem Fall von T09 bis T12 | der Adapter lässt `X-Correlation-Id` weg ⇒ rot; der Text von `message` eines Codes geändert ⇒ rot | Nr. 4 |
| T09 | Fehlender Replay-Treffer: `404`, `ODR-E5001`, `replay` | `TestVertragFehlerklassen`, Zeile *fehlender Replay-Treffer* (mit `?x=GEHEIM-Q`) | der Replay-Dienst liefert `ODR-E5003` statt `ODR-E5001` ⇒ rot | Nr. 5 |
| T10 | Ungültige Metadaten: `422`, `ODR-E6004`, `import`, für `$metadata` und für eine Collection | `TestVertragFehlerklassen`, Zeile *ungültige Metadaten*, Schleife über beide Pfade | `ErrMetadatenUngueltig` wird zu `ODR-E9001` abgebildet ⇒ rot | Nr. 5 |
| T11 | Nicht unterstützte Query-Option: `501`, `ODR-E6001`, `mock`, für jede Option der Spalte *Nicht in MVP* | `TestVertragFehlerklassen`, Schleife über `$search`, `$apply`, `$expand` | `$apply` wird als `400` `ODR-E6005` abgewiesen ⇒ nur der Fall `$apply` rot (Element, das kein Fixture-Schritt allein trägt) | Nr. 5 |
| T12 | Unvollständige Coverage: `422`, `ODR-E6003`, `mock`, ohne `X-Recorder-Coverage` | `TestVertragFehlerklassen`, Zeile *unvollständige Coverage* (strikt, `wsTeil`) | `STRICT_COVERAGE` wird nicht gelesen ⇒ `200`, rot | Nr. 5 |
| T13 | Die Codes der vier Klassen sind paarweise verschieden; `ODR-E6004` und `ODR-E6003` (beide `422`) unterscheidet nur der Code | `TestVertragFehlerklassen`: Prüfung der Codes über alle Paare der Tabelle (Schleife über Paare) | der Mock bildet `ErrMetadatenUngueltig` auf `ODR-E6003` ab ⇒ Zeile *ungültige Metadaten* rot; zwei Zeilen mit demselben Code ⇒ die Paarprüfung rot | Nr. 5 |
| T14 | Es gibt genau vier Klassen | Zählprüfung der Tabelle in `TestVertragFehlerklassen` | eine Zeile entfällt ⇒ rot; eine fünfte Zeile kommt hinzu ⇒ rot | Nr. 5 |
| T15 | Keine Antwort enthält `GEHEIM-A`, `GEHEIM-B` oder `GEHEIM-Q` (Header oder Body) | Schleife über alle Antworten der Matrix und der Klassen × drei Marker, in beiden Tests; die Marker gleichen keinem Wert einer Fixture | der Adapter hängt die Query an `message` ⇒ T11 und T09 rot; er spiegelt `Authorization` in einen Header ⇒ jede Antwort rot | Nr. 6 |
| T16 | **Negativ:** Treiber (Argumente, Wertung, Namen gegen `-test.list`, Flags von `docker run`, Frist, Stufe `contract`) | kein automatischer Fall; Handlauf des Implementers vor der Übergabe | — | Nr. 9 f |
| T17 | **Negativ:** Codes der Klassen außer `ODR-E5001`, `ODR-E6004`, `ODR-E6001`, `ODR-E6003`; Import-Abweisung; Binary; Protokoll-ID; keine Kopie; Soll-Spalten | `ODR-E5002`: `TestE2EDeterminismus` Nr. 4; die übrigen: Dienst-Verträge (`SPEC-135`, `SPEC-238`) | — | Nr. 9 a bis e, g, h |

**Mutationen.** Der Implementer trägt hier die Tabelle (Zusage · Mutation · roter Test) und, falls ein Fänger nur mit Wahrscheinlichkeit rot wird, die Rate je Mutation und je Lauf von `go test` mit Zahl der Prozesse und Messendem ein (`AGENTS.md` §3.10); *Wird beim Code gefüllt.*

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-mock-abnahme` und `slice-bootstrap-m3` liegen in `done/` (die Fixtures `ODV4-*` und der Mock-Betrieb stehen; vier Fehlerklassen von `LH-QA-05` erreichen den Eingang erst dort). Der Vertrag `SPEC-042` trägt Umfang und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice ändert `Makefile` und `harness/mk/` und startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt ([Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 7).

**Stand der Bedingungen beim Start** (Planner, 2026-11-08): `slice-mock-abnahme` und `slice-bootstrap-m3` liegen in `done/`; `slice-gate-integration-determinism` ist nach `open/` zurückgestellt (User-Vorrang `welle-mock`), damit hält kein anderer Slice die Gate-Kette (Nr. 7). In `in-progress/` läuft zugleich allein `slice-bootstrap-m3-e2e-mock` (Linie B; Pfade `tests/e2e/` und `tools/harness/test-integration.sh`, keine Kettenpfade). Der Vertrag `SPEC-042`, die Tabelle *Zusage · Fänger* und das Raster fehlen noch: der Architect schreibt sie im Slice-Zweig vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Entscheidungen aus §6 (Doppelung der Fixtures, Weg der Klasse „ungültige Metadaten“) gehören zu seinem Auftrag.

**Größenmarke** (Architect, 2026-11-08, `SPEC-128` Nr. 1 und `AGENTS.md` §3.14): Die Tabelle *Zusage · Fänger* in §3 hat 17 Zeilen (T01 bis T17, alle Klasse `andere`, davon T16 und T17 Negative): *a* = 0, *b* = 17, *f* = 0 (kein Fake: der Prüfling ist der Dienst der Abnahme, der Eingang ein echter Server, das Handle kein Fake), *p* = 90 (Treiber rund 70, Fragment rund 10, Stufe `contract` rund 10): *S* = 80·17 + 90 = 1450, *M* = 100·⌈135·1450/10000⌉ = 2000. *S* liegt unter der Grenze 2223 (135·*S* = 195750 ≤ 300000): keine Schnittfrage; ein Teil ohne geteilten Mechanismus ließe sich auch nicht abtrennen, denn Matrix und Klassen teilen Prüffunktion des Fehlerobjekts und Treiber. Der Block trägt den Teil des Architect (Pfade, `produktcode`, `fakes`); `basis`, `schaetzung` und `marke` setzt der Planner (die Prosa-Marke 1430 / 2000 vom Start ist damit ersetzt) in eigenem Commit vor dem ersten Code-Commit und rechnet nach `SPEC-128` Nr. 1 neu. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` und `docs/` zählen nicht.

```groessenmarke
pfad: src/bootstrap/vertrag_http_test.go
pfad: tests/integration/vertrag_fehlerklassen_test.go
pfad: harness/mk/test-contract.mk
pfad: tools/harness/test-contract.sh
pfad: Dockerfile
produktcode: 90
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner vor dem ersten Code-Commit aus dem Block setzt (Architect-Schätzung *S* 1450, *M* 2000, §4 *Größenmarke*) nach `AGENTS.md` §3.14. Der Planner schätzt den Slice als klein.
- `in-progress` → `open` (blockiert): eine Fehlerklasse von `LH-QA-05` erreicht den HTTP-Eingang nicht (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-08; Nr. = Nummer in `SPEC-042`): jede Zelle nennt einen Ort oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12) und gibt sie an den Architect zurück.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Aufruf des Treibers (Argument) | gilt nicht: der Aufruf ohne Argument ist der Normalfall (Nr. 2) | Nr. 2 (Exit 2) | Nr. 2 (jedes Argument, Exit 2) | Nr. 2, 8 c (Exit 2) | Nr. 8 f (frischer Container je Binary) |
| Schritt der Matrix am Draht | Nr. 3 c (leere Query ohne `?`, leerer Body ohne Body und `Content-Type`) | Nr. 3 c (dieselbe Regel) | Nr. 3 b und 6 (zusätzliche Header `Authorization`, `Cookie`); unbekannte Entity oder Option: die Fixtures `ODV4-E-01..03`, Nr. 3 d | Nr. 8 a (Verbindungsfehler, `Fatal`) | Nr. 3 b, 8 f (frische Umgebung und Client je ID) |
| Fall einer Fehlerklasse | gilt nicht: ein fehlender oder unbekannter Workspace ist ein Zuordnungsfehler (`ODR-E5003`, `ODR-E7001`), nicht eine der vier Klassen von Nr. 5; ihn fangen `SPEC-195` und `TestE2EMock` | gilt nicht (derselbe Grund) | Nr. 5 (nie aufgenommener Pfad, kaputtes Metadaten-Dokument, Option aus *Nicht in MVP*, unvollständige Collection) | Nr. 8 a, d | Nr. 8 f |
| Prüfung des Fehlerobjekts | Nr. 4 (leere Korrelations-ID, fehlender Header oder Feld ist ein Verstoß) | Nr. 4 (leerer Body, leere `message` sind Verstöße) | Nr. 4 (Code außerhalb von `SPEC-109` Nr. 2, fremdes Feld, falscher Zielbereich) | gilt nicht: reine Funktion über eine gelesene Antwort | gilt nicht: ohne Zustand |
| Aufbau von Umgebung und Handle | gilt nicht: `t.TempDir` legt `DATA_DIR` an (Nr. 8 f) | gilt nicht (derselbe Grund) | Nr. 5 (Zeile *ungültige Metadaten*: ein Dokument `<kaputt` in der Ablage) | Nr. 8 d (Herunterfahren, Fehler dabei ist rot) | Nr. 8 f (ein Verzeichnis je ID und Klasse) |

Zusatztabelle (Modi und mehrere Verstöße; `AGENTS.md` §3.16):

| Fall | Fänger |
|---|---|
| Modus lax (alle Schritte außer `Strikt`) | T05 (`ODV4-GET-01` am lax Server) |
| Modus strikt (`StrictCoverage`) | T05 (`ODV4-E-04` a) |
| Modus `replay` | T09 |
| Modus `mock` | T10, T11, T12 |
| mehrere Verstöße an einer Anfrage | gilt nicht: jeder Fall von Nr. 5 trägt genau einen Verstoß; die Rangfolge mehrerer Verstöße entscheiden `SPEC-195` und `SPEC-160`, nicht dieser Vertrag |

**Entscheidungen des Architect** (2026-11-08, aus den Fragen der Planung):

- Doppelung der Fixtures mit `make test` — **Ausgang:** entschieden, `SPEC-042` Nr. 7: der HTTP-Lauf liegt im Paket `bootstrap_test` neben der Abnahme, ruft `abFixtures()` ohne Kopie und läuft auch unter `make test`; `make test-contract` wählt per Name aus und fügt das Urteil des Treibers hinzu. Die Fixtures liegen in `_test.go` des Composition Root und sind aus `tests/` nicht importierbar; `SPEC-049` Nr. 11 a weist dem Test am HTTP-Eingang die Fixtures als Quelle zu.
- Weg der Klasse „ungültige Metadaten“ — **Ausgang:** entschieden, `SPEC-042` Nr. 5 und 9 c: der Eingang ist der Listener der Daten (`mock`), das Dokument liegt als `MetadataBlob` kaputt in der Ablage (`sqlite.Open` und Blob-Adapter vor dem Start, erlaubt nach ADR-0012); die Abweisung beim Import ist Sache der Admin-API und kein Eingang dieses Vertrags. `slice-bootstrap-m6` wird dafür nicht gebraucht.

**Offene Punkte:**

- Aufnahme von `make test-contract` in `GATE_CHECKS`: der Vertrag lässt das Ziel daneben (`SPEC-042` Nr. 2, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); ob es in die Kette gehört, entscheidet wie bei `make test-integration` und `make test-determinism` der Betreiber — **Ausgang:** offen; Adresse Planner (Entscheidung des Users nötig; `slice-gate-integration-determinism` führt in §1 allein die beiden anderen Ziele, eine Sendung dorthin ist damit nicht erfolgt, `AGENTS.md` §3.13).
- Die Spalten *Soll* und *Nicht in MVP* von `SPEC-049` haben kein Ziel, an das ihre Welle ihre Fälle hängen müsste; `SPEC-042` Nr. 9 e hält fest, dass dieser Vertrag sie nicht fährt — **Ausgang:** offen; Entscheidung bei Start der jeweiligen Welle.
- Randformen (`AGENTS.md` §3.12, §3.16) — **Ausgang:** entschieden; das Raster steht oben, jede Zelle nennt Nummer oder „gilt nicht“ mit Grund.
- Stufe `contract` und Treiber (Handlauf, `SPEC-042` Nr. 9 f) und ob die Test-Binaries ohne weitere Dateien aus dem Arbeitsverzeichnis des Pakets laufen (Nr. 1) — **Ausgang:** offen; der Implementer prüft es beim ersten Lauf und gibt eine Abweichung an den Architect zurück.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register (`../observations/`, Stand 2026-10-31) trägt Treffer: `BEO-REPO/randform-erst-im-review-entschieden` (22×, verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (30×, verkörpert in `AGENTS.md` §3.15), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (15×, verkörpert in `AGENTS.md` §3.9) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (12×, verkörpert in `AGENTS.md` §3.16; „je Fehlerklasse“ ist eine Allaussage mit Fänger über alle Klassen); der Stand jedes Eintrags steht in seiner `state.md`.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
