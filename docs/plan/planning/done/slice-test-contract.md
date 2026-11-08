# Slice slice-test-contract: `make test-contract` — Referenzanfragen und Negativfälle am HTTP-Eingang

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](welle-mock.md).

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

- [x] Das Ziel `make test-contract` (Treiber, Fragment, Stufe `contract` des `Dockerfile`) und der Vertrag `SPEC-042` (der Stand „entsteht mit dem Slice, der den Code liefert“ wird zum gelieferten Umfang); die Muss-Fixtures `ODV4-*` aus `slice-mock-abnahme` laufen über echtes HTTP, aufgerufen aus `abFixtures()` (Aufruf, keine Kopie, `SPEC-042` Nr. 3 a).
- [x] Je Fehlerklasse von `LH-QA-05` ein Negativfall am HTTP-Eingang, der Klasse, Status und diagnostischen Kontext prüft und keine Geheimnisse offenlegt.
- [x] Je Zusage eine Mutation, vom Implementer rot gesehen (`AGENTS.md` §3.10): eine Fehlerklasse, die auf eine andere abgebildet wird, färbt den Negativfall rot.
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-042`, Sensors-Tabelle in `harness/README.md`).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

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

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-08, vor dem ersten Code-Commit). Die Nummern hinter *Ort* in der Spalte *Nr.* sind Nummern von `SPEC-042`; die Kopfzeile ist die fünf Zellen von `SPEC-128` Nr. 3 a. Die Klasse (`SPEC-128` Nr. 3 e) ist für jede Zeile `andere`: die Matrix und die Klassen laufen nacheinander, kein Fall prüft Nebenläufigkeit.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| T01 (Ort Nr. 3 a, e) | Die Matrix läuft über HTTP für genau die 36 IDs und 92 Schritte von `abFixtures()`, jede einmal | `TestVertragMatrixZaehlung` (Tabelle) und `abPruefeZaehlung` am Lauf von `TestVertragMatrix` | die Schleife überspringt eine ID und, getrennt, den Schritt `b` einer ID ⇒ rot | andere |
| T02 (Ort Nr. 3 d) | Header einer Erfolgsantwort: Menge und Werte gleich der Tabelle, ohne weiteren; allein `Date` und `Content-Length` zählen nicht | `TestVertragVergleich`, Schleife über Fälle: weiterer Header, fehlender, anderer Wert, doppelter Wert (je rot); nur `Date` anders, nur `Content-Length` anders (je grün) | die Ausnahmen um `Odata-Version` erweitert ⇒ der Fall „Header fehlt“ bleibt grün, rot | andere |
| T03 (Ort Nr. 3 d) | Body byte-gleich, bis zum Ende gelesen | `TestVertragVergleich`: Zeilenende, Leerzeichen am Ende, gekürzt, anders (je rot) | der Vergleich schneidet Leerraum ab ⇒ Fall „Leerzeichen“ rot | andere |
| T04 (Ort Nr. 3 c) | `drahtQuery` kodiert genau die Menge von Nr. 3 c und lässt jedes andere Byte unverändert, Hex in Großbuchstaben | `TestVertragDrahtQuery`, Schleife über alle 256 Byte-Werte (Menge kodiert, Rest gleich; ein Fall mit `%41` unverändert, einer mit leerer Query ohne `?`) | Leerzeichen nicht kodiert ⇒ rot; `%` kodiert ⇒ rot; Hex klein ⇒ rot | andere |
| T05 (Ort Nr. 3 b) | Schritte mit `Strikt` gehen an den strikten Server, alle anderen an den lax konfigurierten | `ODV4-E-04` a (`422`, `ODR-E6003`) in `TestVertragMatrix`; `ODV4-GET-01` am lax Server (`200`) | der strikte Schritt geht an den lax Server ⇒ `200` statt `422`, rot | andere |
| T06 (Ort Nr. 3 b, c) | Mutierende Schritte (`POST`, `PATCH`, `PUT`, `DELETE`) tragen Body und `Content-Type` wie in `SPEC-049` Nr. 2 | `ODV4-CRUD-01..06` in `TestVertragMatrix` (`201` mit `Location`, `204`, Fehler) | der Client sendet kein `Content-Type` ⇒ `ODV4-CRUD-01` rot (`ODR-E6007` statt `201`) | andere |
| T07 (Ort Nr. 3 e) | Nach den Schritten einer ID gilt die Spalte *Revision*, gelesen an der Datei | Stand-Prüfung am Ende jeder ID in `TestVertragMatrix` | der Client wiederholt jede mutierende Anfrage ⇒ Revision falsch, rot | andere |
| T08 (Ort Nr. 4) | Eine Fehlerantwort trägt `Content-Type: application/json`, `Content-Length` gleich der Länge, nichtleere `X-Correlation-Id`, genau die drei Felder in Reihenfolge mit `code`, `target`, `message` des Codes | `TestVertragFehlerobjekt` und die Teilfälle `Fehlerobjekt` in `TestVertragMatrix`, je Paket mit eigener Fassung der Prüffunktion (Nr. 4): Schleife über die Eigenschaften der Prüffunktion (Content-Type, Länge, ID leer, viertes Feld, Reihenfolge, `target`, `message`, Code außerhalb von `SPEC-109`) mit je einem verletzten Beispiel; die Prüfung läuft an jedem Fehlerschritt der Matrix und jedem Fall von T09 bis T12 | der Adapter lässt `X-Correlation-Id` weg ⇒ rot; der Text von `message` eines Codes geändert ⇒ rot | andere |
| T09 (Ort Nr. 5) | Fehlender Replay-Treffer: `404`, `ODR-E5001`, `replay` | `TestVertragFehlerklassen`, Zeile *fehlender Replay-Treffer* (mit `?x=GEHEIM-Q`) | der Replay-Dienst liefert `ODR-E5003` statt `ODR-E5001` ⇒ rot | andere |
| T10 (Ort Nr. 5) | Ungültige Metadaten: `422`, `ODR-E6004`, `import`, für eine Collection (`$metadata` ist Nicht-Fall, Nr. 9 i) | `TestVertragFehlerklassen`, Zeile *ungültige Metadaten* (`GET /odata/Products`) | `ErrMetadatenUngueltig` wird zu `ODR-E9001` abgebildet ⇒ rot | andere |
| T11 (Ort Nr. 5) | Nicht unterstützte Query-Option: `501`, `ODR-E6001`, `mock`, für jede Option der Spalte *Nicht in MVP* | `TestVertragFehlerklassen`, Schleife über `$search`, `$apply`, `$expand` | `$apply` wird als `400` `ODR-E6005` abgewiesen ⇒ nur der Fall `$apply` rot (Element, das kein Fixture-Schritt allein trägt) | andere |
| T12 (Ort Nr. 5) | Unvollständige Coverage: `422`, `ODR-E6003`, `mock`, ohne `X-Recorder-Coverage` | `TestVertragFehlerklassen`, Zeile *unvollständige Coverage* (strikt, `wsTeil`) | `STRICT_COVERAGE` wird nicht gelesen ⇒ `200`, rot | andere |
| T13 (Ort Nr. 5) | Die Codes der vier Klassen sind paarweise verschieden; `ODR-E6004` und `ODR-E6003` (beide `422`) unterscheidet nur der Code | `TestVertragFehlerklassen`: Prüfung der Codes über alle Paare der Tabelle (Schleife über Paare) | der Mock bildet `ErrMetadatenUngueltig` auf `ODR-E6003` ab ⇒ Zeile *ungültige Metadaten* rot; zwei Zeilen mit demselben Code ⇒ die Paarprüfung rot | andere |
| T14 (Ort Nr. 5) | Es gibt genau vier Klassen | Zählprüfung der Tabelle in `TestVertragFehlerklassen` | eine Zeile entfällt ⇒ rot; eine fünfte Zeile kommt hinzu ⇒ rot | andere |
| T15 (Ort Nr. 6) | Keine Antwort enthält `GEHEIM-A`, `GEHEIM-B` oder `GEHEIM-Q` (Header oder Body) | Schleife über alle Antworten der Matrix und der Klassen × drei Marker, in beiden Tests; die Marker gleichen keinem Wert einer Fixture | der Adapter hängt die Query an `message` ⇒ T11 und T09 rot; er spiegelt `Authorization` in einen Header ⇒ jede Antwort rot | andere |
| T16 (Ort Nr. 9 f) | **Negativ:** Treiber (Argumente, Wertung, Namen gegen `-test.list`, Flags von `docker run`, Frist, Stufe `contract`) | kein automatischer Fall; Handlauf des Implementers vor der Übergabe | — | andere |
| T17 (Ort Nr. 9 a bis e, g, h bis j) | **Negativ:** `$metadata` mit ungültigem Dokument (Nr. 9 i); zwei Fassungen von `vertragFehlerobjekt` ohne Gleichheitsfänger (Nr. 9 j); Codes der Klassen außer `ODR-E5001`, `ODR-E6004`, `ODR-E6001`, `ODR-E6003`; Import-Abweisung; Binary; Protokoll-ID; keine Kopie; Soll-Spalten | `ODR-E5002`: `TestE2EDeterminismus` Nr. 4; die übrigen: Dienst-Verträge (`SPEC-135`, `SPEC-238`) | — | andere |

**Mutationen** (Implementer, 2026-11-08; `AGENTS.md` §3.10). Jede Mutation lief in einer schreibgeschützt eingehängten Kopie der Quellen (Container ohne Netz, `--memory 4g`, `go test -run '^TestVertrag' -count=1`); `tools/mutationslauf` scheitert für `src/bootstrap` an der Nullprobe-Grenze (Register `nullprobe-grenze-unter-der-paketlaufzeit`), die Läufe sind eigene Aufrufe. Kein Fänger dieses Slice ist ein Wahrscheinlichkeits-Fänger: der Test fährt die Fälle nacheinander, jede Mutation ist ein Prozess mit `-test.count=1` und wurde in jedem ihrer Läufe rot (1 von 1; gemessen vom Implementer bei Last des Hosts 8 bis 26, Messender: der Implementer; eine Einzelzahl gilt als Größenordnung).

| Nr. | Zusage | Mutation | roter Test |
|---|---|---|---|
| T01 | jede ID, jeder Schritt | Schleife überspringt `ODV4-CRUD-06` / den Schritt `b` von `ODV4-Q-07` | `TestVertragMatrix` (`abPruefeZaehlung`), je rot |
| T02 | Menge der Header | Ausnahme des Vergleichs um `Odata-Version` erweitert | `TestVertragMatrix` (alle Fixtures mit Erfolgsschritt) und die Fälle „Header … fehlt“ in `TestVertragVergleich` (Schleife über jeden Namen) |
| T03 | Body byte-gleich | Vergleich schneidet Leerraum ab | `TestVertragVergleich` (Fälle *Zeilenende*, *Leerzeichen*) |
| T04 | `drahtQuery` | Leerzeichen nicht kodiert · `%` kodiert · Hex klein | `TestVertragMatrix` (nur Leerzeichen); `TestVertragDrahtQuery` (alle drei; `%` und Hex klein trägt keine Fixture) |
| T05 | Server nach `Strikt` | strikte Schritte an den lax Server · alle Schritte an den strikten | `TestVertragMatrix`, Teilfall `ODV4-E-04` (je 1 rot) |
| T06 | Body und `Content-Type` | Client sendet kein `Content-Type` | `TestVertragMatrix` (8 Teilfälle) |
| T07 | Revision | Client sendet jede mutierende Anfrage doppelt | `TestVertragMatrix` (6 Teilfälle) |
| T08 | Fehlerobjekt | Adapter lässt `X-Correlation-Id` weg · hängt an `message` ein Zeichen | `TestVertragMatrix` und `TestVertragFehlerklassen` (12 Teilfälle) |
| T09 | `ODR-E5001` | Replay liefert `ODR-E5003` | `TestVertragFehlerklassen/fehlender_Replay-Treffer` |
| T10 | `ODR-E6004` | `ErrMetadatenUngueltig` auf `ODR-E9001` | `TestVertragFehlerklassen/ungültige_Metadaten` |
| T11 | `ODR-E6001` je Option | `$apply` · `$search` und `$expand` aus der Liste der nicht unterstützten Optionen (`400` `ODR-E6005`) | `TestVertragFehlerklassen/nicht_unterstützte_Query-Option` (Fälle der entfernten Optionen rot, übrige grün); `$apply` auch `ODV4-E-03` |
| T12 | `ODR-E6003` strikt | `STRICT_COVERAGE` nicht gelesen | `TestVertragFehlerklassen/unvollständige_Coverage` |
| T13 | Codes paarweise verschieden | `ErrMetadatenUngueltig` auf `ODR-E6003` · Tabelle mit zwei Zeilen `ODR-E6003` | Zeile *ungültige Metadaten* · Paarprüfung („tragen denselben Code“) |
| T14 | genau vier Klassen | Zeile entfällt · fünfte Zeile | `TestVertragFehlerklassen` („3 Klassen, erwartet genau 4“, „5 Klassen …“) |
| T15 | keine Geheimnisse | Query an `message` · Query in Header (nur Fälle mit `GEHEIM-Q`) · `Authorization` in Header der Fehlerantwort · der Erfolgsantwort | `TestVertragFehlerklassen` · dessen Prüfung über alle Antworten (Teilfälle 0, Elternfall rot) · beide Tests (8 Teilfälle) · `TestVertragMatrix` (33 Teilfälle) |

Fix-Runde (Review, MEDIUM 1 und 2), je Mutation einmal in der Kopie gefahren, rot:

| Nr. | Zusage | Mutation | roter Test |
|---|---|---|---|
| T15 Sendeseite | jede Anfrage der Matrix trägt `Authorization` (`Bearer GEHEIM-A`) und `Cookie` (`GEHEIM-B`); die Klassen *fehlender Replay-Treffer* und *nicht unterstützte Query-Option* tragen `GEHEIM-Q` in der Query, die beiden anderen nicht | Client der Matrix setzt beide Header nicht · `Cookie` fehlt nur bei `DELETE` (nicht im Befund) · Zuordnung der Klassen ohne beide Header · `GEHEIM-Q` fehlt allein im Pfad `$expand` (nicht im Befund) | `TestVertragMatrix` (Server-Spiegel `vertragSpiegel`, `vertragSendeseite`; bei `DELETE` `ODV4-H-02` und `ODV4-CRUD-04`) · `TestVertragFehlerklassen` (Spiegel am Transport des Clients, `q` je Zeile der Tabelle; 4 Zeilen, bei `$expand` allein die Zeile *nicht unterstützte Query-Option*) |
| Nr. 8 d in `tests/integration` | ein Fehler beim Herunterfahren ist rot | `Handle.Herunterfahren` meldet für `replay` und `mock` einen Fehler (der Aufbau der Aufnahme ist ausgenommen, er bricht selbst ab) | `TestVertragFehlerklassen`: Meldung `Herunterfahren replay` bzw. `mock` aus `vertragHoch` (4 Zeilen); `hochgefahren` aus `m2_test.go` bleibt für die anderen Tests unverändert |

Mengen der Sendeseite (`AGENTS.md` §3.17): 3 Marker (A in `Authorization`, B in `Cookie`, Q in der Query) × Abweichung (fehlt, anders, doppelt; Q unerwartet; Zahl der Anfragen zu klein) als verletzte Beispiele in `TestVertragVergleich/Sendeseite` (14 Fälle) und `TestVertragFehlerobjekt` (8 Fälle); an der Matrix 92 Anfragen in 36 IDs (Zahl der gesehenen gleich Zahl der Schritte), an den Klassen 6 Anfragen in 4 Zeilen. Der Spiegel der Matrix liegt am Server (vor `Mock`), der der Klassen am Transport des Clients, weil der Handle seinen Server selbst baut; ein Header, den der Client sendet und der Server nicht erreicht, bleibt in den Klassen ungesehen (akzeptiertes Negativ: der Adapter reicht die Header unverändert, `SPEC-136` Nr. 4, und die Matrix prüft es am Server).

Mengen (`AGENTS.md` §3.17): *Header einer Erfolgsantwort* (3 Namen von `abKK`, je fehlt, anders, doppelt: 9 Fälle in `TestVertragVergleich`); *Byte-Werte von `drahtQuery`* (256, davon 171 kodiert); *Fehlerobjekt* (9 Codes × 20 Eigenschaften = 180 verletzte Beispiele, je mit Stichwort des Grundes, und das unverletzte je Code); *Marker* (3 × Body, Header-Wert, Header-Name in beiden Tests); *Optionen* (3 der Klasse in `TestVertragFehlerklassen`, die Liste `nichtUnterstuetzteOption` des Parsers hat 10 Namen: 7 davon (`$compute`, `$levels`, `$schemaversion`, `$deltatoken`, `$skiptoken`, `$index`, `$id`) führt kein Fall, akzeptiertes Negativ: `SPEC-042` Nr. 5 nennt die Optionen der Spalte *Nicht in MVP* von `SPEC-049`, und eine Abbildung dieser 7 auf einen anderen Code bleibt am Eingang unbemerkt, die Tests des Parsers prüfen sie); *Klassen* (4, Paare 6).

**Handlauf des Treibers** (T16, `SPEC-042` Nr. 9 f; Kopie der Quellen im Scratchpad mit je einer Änderung): Argument `x` und leeres Argument ⇒ `Aufruf:` und Exit 2 ohne Container; `t.Skip` in `TestVertragDrahtQuery` ⇒ `rot (/out/bootstrap.test: SKIP)`, Exit 1, das zweite Binary lief; `t.Error` ⇒ `rot (… Exit 1)`; ein Name im Treiber, den `-test.list` nicht nennt ⇒ `nicht gefunden`, Exit 2; ein Test, der das Außennetz anwählt, ist mit `--network none` grün und ohne das Flag rot. Die Test-Binaries laufen aus `/src/src/bootstrap` und `/src/tests/integration` ohne weitere Dateien (`make test-contract` grün, `SPEC-042` Nr. 1). Der Fall *kein PASS* (Name in `-test.list`, passt nicht auf `^TestVertrag`) ist im Review gefahren: `rot (/out/integration.test: kein PASS (TestM3))`, Exit 1, das erste Binary lief und blieb grün.

**Fänger und Vertrag.** Die Prüfung des Fehlerobjekts liegt als Funktion in beiden Test-Paketen (`vertragFehlerobjekt`), weil `bootstrap_test` und `integration_test` sich nicht importieren; beide Fassungen tragen dieselbe Schleife über 9 Codes × 20 Eigenschaften (`TestVertragVergleich/Fehlerobjekt`, `TestVertragFehlerobjekt`). Die Fixture-Tabelle ist nicht kopiert (`abFixtures()` wird gerufen).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-mock-abnahme` und `slice-bootstrap-m3` liegen in `done/` (die Fixtures `ODV4-*` und der Mock-Betrieb stehen; vier Fehlerklassen von `LH-QA-05` erreichen den Eingang erst dort). Der Vertrag `SPEC-042` trägt Umfang und Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Der Slice ändert `Makefile` und `harness/mk/` und startet nur, wenn kein anderer Slice der Gate-Kette in `in-progress/` liegt ([Übergangsregeln](../README.md#parallelbetrieb-in-worktrees-übergangsregeln) Nr. 7).

**Stand der Bedingungen beim Start** (Planner, 2026-11-08): `slice-mock-abnahme` und `slice-bootstrap-m3` liegen in `done/`; `slice-gate-integration-determinism` ist nach `open/` zurückgestellt (User-Vorrang `welle-mock`), damit hält kein anderer Slice die Gate-Kette (Nr. 7). In `in-progress/` läuft zugleich allein `slice-bootstrap-m3-e2e-mock` (Linie B; Pfade `tests/e2e/` und `tools/harness/test-integration.sh`, keine Kettenpfade). Der Vertrag `SPEC-042`, die Tabelle *Zusage · Fänger* und das Raster fehlen noch: der Architect schreibt sie im Slice-Zweig vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Entscheidungen aus §6 (Doppelung der Fixtures, Weg der Klasse „ungültige Metadaten“) gehören zu seinem Auftrag.

**Größenmarke** (Architect, 2026-11-08, `SPEC-128` Nr. 1 und `AGENTS.md` §3.14): Die Tabelle *Zusage · Fänger* in §3 hat 17 Zeilen (T01 bis T17, alle Klasse `andere`, davon T16 und T17 Negative): *a* = 0, *b* = 17, *f* = 0 (kein Fake: der Prüfling ist der Dienst der Abnahme, der Eingang ein echter Server, das Handle kein Fake), *p* = 90 (Treiber rund 70, Fragment rund 10, Stufe `contract` rund 10): *S* = 80·17 + 90 = 1450, *M* = 100·⌈135·1450/10000⌉ = 2000. *S* liegt unter der Grenze 2223 (135·*S* = 195750 ≤ 300000): keine Schnittfrage; ein Teil ohne geteilten Mechanismus ließe sich auch nicht abtrennen, denn Matrix und Klassen teilen Prüffunktion des Fehlerobjekts und Treiber. Der Block trägt den Teil des Architect (Pfade, `produktcode`, `fakes`) und den des Planner: `basis` ist `54a1f1b4` (der Commit des Architect, auf dem der Zweig aufbaut), `schaetzung` 1450, `marke` 2000 (Planner, 2026-11-08, nach `SPEC-128` Nr. 1 nachgerechnet, vor dem ersten Code-Commit). Die Zeilen T16 und T17 zählen als `andere` (`SPEC-128` Nr. 3 e: ein akzeptiertes Negativ ist eine Zeile); *a* = 0, *b* = 17. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `harness/vertraege/` und `docs/` zählen nicht.

```groessenmarke
pfad: src/bootstrap/vertrag_http_test.go
pfad: tests/integration/vertrag_fehlerklassen_test.go
pfad: harness/mk/test-contract.mk
pfad: tools/harness/test-contract.sh
pfad: Dockerfile
produktcode: 90
fakes: 0
basis: 54a1f1b4
schaetzung: 1450
marke: 2000
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 2000 (*S* 1450, im Block gesetzt, §4 *Größenmarke*) nach `AGENTS.md` §3.14. Der Planner schätzt den Slice als klein.
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
- Weg der Klasse „ungültige Metadaten“ — **Ausgang:** entschieden, `SPEC-042` Nr. 5 und 9 c: der Eingang ist der Listener der Daten (`mock`), das Dokument liegt als `MetadataBlob` kaputt in der Ablage (`sqlite.Open` und Blob-Adapter vor dem Start, erlaubt nach [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)); die Abweisung beim Import ist Sache der Admin-API und kein Eingang dieses Vertrags. `slice-bootstrap-m6` wird dafür nicht gebraucht.

**Offene Punkte:**

- Aufnahme von `make test-contract` in `GATE_CHECKS`: der Vertrag lässt das Ziel daneben (`SPEC-042` Nr. 2, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)); ob es in die Kette gehört, entscheidet wie bei `make test-integration` und `make test-determinism` der Betreiber — **Ausgang:** weiter offen; Adresse Planner, in Roadmap und `welle-mock` als Entscheidung des Users sichtbar (Entscheidung des Users nötig; `slice-gate-integration-determinism` führt in §1 allein die beiden anderen Ziele, eine Sendung dorthin ist damit nicht erfolgt, `AGENTS.md` §3.13).
- Die Spalten *Soll* und *Nicht in MVP* von `SPEC-049` haben kein Ziel, an das ihre Welle ihre Fälle hängen müsste; `SPEC-042` Nr. 9 e hält fest, dass dieser Vertrag sie nicht fährt — **Ausgang:** weiter offen; Entscheidung bei Start der jeweiligen Welle.
- Randformen (`AGENTS.md` §3.12, §3.16) — **Ausgang:** entschieden; das Raster steht oben, jede Zelle nennt Nummer oder „gilt nicht“ mit Grund.
- Stufe `contract` und Treiber (Handlauf, `SPEC-042` Nr. 9 f) und ob die Test-Binaries ohne weitere Dateien aus dem Arbeitsverzeichnis des Pakets laufen (Nr. 1) — **Ausgang:** entfallen; der Handlauf ist gefahren (§3, *Handlauf des Treibers*), `make test-contract` ist grün, die Binaries laufen ohne weitere Dateien, keine Abweichung; der Verifier hat den Fall Argument × Fall bestätigt (`x` und `""` ergeben Exit 2 ohne Container).
- Zeile *ungültige Metadaten* (`SPEC-042` Nr. 5, T10): `GET /odata/$metadata` auf ein Dokument `<kaputt` antwortet `200` mit den gespeicherten Bytes, nicht `422` `ODR-E6004`, denn `SPEC-198` Nr. 4 liest das Modell dort nicht; die Vorgabe „für `$metadata` und für eine Collection“ ist damit unerfüllbar, ohne den Dienst zu ändern. Der Test fährt allein `GET /odata/Products` (`tests/integration/vertrag_fehlerklassen_test.go`); der Implementer entscheidet den zweiten Pfad nicht (`AGENTS.md` §3.12). Der Architect wählt die Beschränkung auf die Collection — **Ausgang:** entschieden, `SPEC-042` Nr. 5 und 9 i: `$metadata` ist Nicht-Fall mit Grund (kleinste vertragsgerechte Lösung; ein leerer Blob ist die Klasse *Service ohne Metadaten*, nach `SPEC-195` Nr. 9 `ODR-E6004`, braucht aber einen Aufbau ohne Workspace-Baseline, den dieser Vertrag nicht führt); T10 und Nr. 5 sind nachgezogen. Der Implementer ändert keinen Test.
- Die Prüfung des Fehlerobjekts (`SPEC-042` Nr. 4: „die Prüfung selbst als Funktion, die auch `TestVertragMatrix` ruft“) liegt zweimal vor, je Test-Paket (`vertragFehlerobjekt` in `src/bootstrap/vertrag_http_test.go` und in `tests/integration/vertrag_fehlerklassen_test.go`), weil `tests/integration` den Composition Root nicht nennt und die Pakete `_test` sich nicht importieren; beide Schleifen (9 Codes × 20 Eigenschaften) laufen gleich. Eine gemeinsame Funktion bräuchte ein weiteres Paket unter `tests/` (Pfad außerhalb des Blocks der Größenmarke, [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)) — **Ausgang:** entschieden, `SPEC-042` Nr. 4 und 9 j: zwei Fassungen mit je eigener Schleife (9 Codes × 20 Eigenschaften) als Fänger; ein gemeinsames Paket unter `tests/` entfällt (neue Schicht oder Kante für eine Funktion, Marke und `.a-check.yml` unberührt halten); die Gleichheit beider Fassungen ist akzeptiertes Negativ mit Grund und Folge (T17). Der Wortlaut von Nr. 4 ist angepasst.

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

**Größen** (`SPEC-128` Nr. 9): `make groessenmarke`: Messung 1311 von Marke 2000, **Auslastung 0,66 (1311 / 2000), nicht unter 0,60: keine Meldung** an den Architect. Keine Überschreitung, keine Rückführung.

**Mutationen:** je Zusage rot gesehen (Tabelle T01 bis T17 in §3). Review-Runde: die Sendeseite der Marker (`Authorization`, `Cookie`, `GEHEIM-Q`) und das Herunterfahren in `tests/integration` hatten keinen Fänger; die Fix-Runde (`69eac720`) deckt beide. Der Verifier löste den Fix gegen die Klasse auf (`AGENTS.md` §3.17) mit 13 eigenen Mutationen an Elementen, die der Befund nicht nannte (B1 bis B4, K1 bis K5, P1 bis P3, H1): alle rot, kein Element der Mengen blieb grün.

**Gate:** `make gates` und `make test-contract` im Zweig grün (Verifier); `make gates` auf dem Merge-Stand und auf `main` nach dem Merge sowie der Lauf der Gegenproben (Stempel) stehen im Commit der Integration. Akzeptierte Negative (Grund und Folge in §3, §6 und `SPEC-042` Nr. 9): ein Header, den der Client sendet und den der Server nicht erreicht, bleibt in den Klassen ungesehen; 7 der 10 nicht unterstützten Optionen führt kein Fall; `$metadata` mit ungültigem Dokument (Nr. 9 i); zwei Fassungen von `vertragFehlerobjekt` ohne Gleichheitsfänger (Nr. 9 j).

**Risiko-Ausgänge (§6):** Stufe und Treiber: entfallen. Entschieden: Doppelung der Fixtures, Weg der Klasse *ungültige Metadaten*, Randformen, Zeile *ungültige Metadaten* (`$metadata`), zwei Fassungen der Prüfung. **Weiter offen mit Adresse (kein Ausgang im Slice):** (1) Aufnahme von `make test-contract` in `GATE_CHECKS`: Entscheidung des Users, in Roadmap und `welle-mock` als Entscheidung des Users sichtbar; (2) Spalten *Soll* und *Nicht in MVP* von `SPEC-049` ohne Ziel: Entscheidung bei Start der jeweiligen Welle ([welle-etag](../welle-etag.md), [welle-expand-batch](../welle-expand-batch.md), [welle-odata-v2](../welle-odata-v2.md)); beide Adressen nehmen die Sendung an (kein Ausschluss trifft sie, §1).

**Steering-Loop-Einträge:**

1. **Die Zusage der Sendeseite hatte keinen Fänger, nur die Empfangsseite:** Marker `Authorization`, `Cookie` und `GEHEIM-Q` wurden in den Antworten gesucht; sendete der Test sie nicht, fand die Suche nichts und blieb grün (Review M-1, Mutation: die Marker aus Anfrage und Pfaden entfernt, beide Pakete grün). Neue Beobachtung `BEO-REPO/sendeseite-der-marker-ohne-faenger` (1×, gezählt, nicht verkörpert). Dazu die Randform *Herunterfahren* nur in einem von zwei Test-Paketen gefangen (Review M-2): Evidenz zu `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15, gezählt, nicht geschärft).
2. **Benannte Spec-Lücke:** `GET /odata/$metadata` auf ein ungültiges Metadaten-Dokument antwortet `200` mit den gespeicherten Bytes, nicht `422` `ODR-E6004`, denn `SPEC-198` Nr. 4 liest das Modell dort nicht; ob `$metadata` ein ungültiges Dokument melden soll, ist eine Frage an `SPEC-198`, nicht an diesen Vertrag. `SPEC-042` Nr. 5 und 9 i halten den Nicht-Fall mit Grund fest; die Rückfrage an `SPEC-198` Nr. 4 bleibt beim Architect.
3. **Beobachtungen ohne neue Regel:** keine weitere. Zur Betriebsregel für Repo-Dateien liegt kein Beleg vor (kein Report nennt `sed` oder Heredoc); eine Evidenz zu `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` entsteht nicht. Die Fix-Runde deckte die Klasse (Verifier); eine Evidenz zu `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` entsteht nicht.

**Spec-Lücken:** eine, benannt in Eintrag 2 (Frage an `SPEC-198` Nr. 4, im Vertrag als Nicht-Fall mit Grund geschlossen).

**Kein Feld `liegt in` in diesem Slice:** nichts wurde in einer Regel, einem Sensor oder einer Spec-Stelle als Lerneintrag verkörpert; Eintrag 1 zählt, Eintrag 2 ist eine benannte Spec-Lücke.

**Drei Paarungen:** Anker (kein Eintrag trägt `liegt in`, nichts zu paaren) · Folge-Slice (die beiden Adressen aus §6 sind in Roadmap und `welle-mock` bzw. bei Start der Folgewellen geführt; die Sendung von `slice-mock-abnahme` ist erfüllt) · Register (Evidenz `slice-test-contract` zu `BEO-REPO/sendeseite-der-marker-ohne-faenger` und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`).

**Nachbarn:** Die Aufnahme in `GATE_CHECKS` bleibt Entscheidung des Users; bis dahin ist `make test-contract` wie `make test-integration` ein Werkzeug neben der Kette (`SPEC-042` Nr. 2, [ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)).

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
