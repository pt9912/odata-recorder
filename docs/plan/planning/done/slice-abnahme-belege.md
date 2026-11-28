# Slice slice-abnahme-belege: Belege zur Erstabnahme: unbekannte Erweiterungen auf dem Record-Pfad, kombinierte Abfrage, Handbuch-Absatz zum Schutz der Daten

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung über die DoD hinaus. Die Wellen, deren Gegenstand er belegt, liegen in `done/` (`welle-record-replay`, `welle-mock`), und kein Closure-Kriterium von [welle-betrieb](../welle-betrieb.md) nennt diese Belege (§3 dort). Gebraucht wird er von M7 (Erstabnahme): die [Roadmap](../in-progress/roadmap.md) führt ihn dort als Bedingung.

**Bezug:** `LH-QA-07`, `LH-FA-08`, `LH-RB-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-QA-07.a` (Nachweis für unbekannte Erweiterungen auf dem Record-Pfad, Vergleich Nr. 1 bis 6) · `SPEC-049` Nr. 3 und 4 (`ODV4-Q-11`) · [SPEC-042](../../../../harness/vertraege/SPEC-042-test-contract.md) Nr. 3 e (Zählung) · `SPEC-046` Nr. 3 (unbekannter Medientyp) · `SPEC-053` (Betriebsbedingungen, Betreiberpflicht nach dem Nachzug des Architect) · [SPEC-200](../../../../harness/vertraege/SPEC-200-handbuch.md) Nr. 1 und 14 b (Prosa im Handbuch). In der Spezifikation ändert der Slice allein die Vermerke *Stand* beim Liefern.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-abnahme-belege`, Worktree `/Development/odata-recorder-wt-belege`; der Zweig entsteht von `main` nach dem Start-Commit, der Start liegt auf `main`)

**Autor:** pt9912 (Planner, aus der Bewertung der Beleg-Lücken durch den Architect, `2e289d4d`). **Datum:** 2026-11-27.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Was zur Erstabnahme fehlt, entsteht ohne Produktcode: (1) als Test: unbekannte OData-Erweiterungen verhindern die Aufzeichnung nicht (`LH-QA-07`, zweiter Halbsatz; fünf Fälle nach `LH-QA-07.a`); (2) als Test: Filter, Sortierung und Paging kombiniert auf fünf Datensätzen (`LH-FA-08` Happy Path) als Schritt `ODV4-Q-11` b der Abnahmematrix, die damit 93 statt 92 Schritte zählt; (3) ein kurzer Absatz *Schutz der Daten* im Benutzerhandbuch mit den zwei Betreiberpflichten (`LH-RB-04`; Antwort 7 des Users vom 2026-11-27, Wortlaut in [welle-betrieb](../welle-betrieb.md) §4).

**Schnitt vor dem Code** (Planner, 2026-11-27): Die Bewertung des Architect nennt einen dritten Beleg, den Absturztest (`SPEC-048`, `LH-QA-03`). Er ist [slice-absturztest](../done/slice-absturztest.md): ungeschnitten *S* = 3280 über der Schnittgrenze (§4); die Teile teilen keinen Mechanismus.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Absturztest (`SPEC-048`) — [slice-absturztest](../done/slice-absturztest.md), Schnitt oben.
- Produktcode — die Belege prüfen Verhalten, das die Spezifikation als vorhanden beschreibt; zeigt ein Test eine Abweichung, greift die Rückführung in §4.
- Das byte-gleiche Binary (`LH-RB-03.a`) — [slice-container-image](../in-progress/slice-container-image.md), Zeile 20 dort.
- Neue Fälle in `make test-integration` (`tests/e2e/`) — der Nachweis läuft im Paket `integration` in `make test` (`LH-QA-07.a`); keine Datei der Gate-Kette.
- Weitere Handbuch-Kapitel — die Handbuch-Slices sind für M7 nicht nötig (Antwort 6 des Users, 2026-11-27); kein Verhalten am Binary ändert sich.
- Die Sätze in `spec/` (`SPEC-053`, der Satz zu den Werten im Modus `mock`, der Satz zum vollen Speicher im Lastenheft) und der Satz im README — Architect auf `main` (Antworten 1, 7 und 8).

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

- [x] `LH-QA-07`: ein Test im Paket `integration` führt die fünf Fälle von `LH-QA-07.a` (Query, Pfad, Header, JSON, Metadaten) im Modus `record` über den Listener gegen einen eigenen Referenzdienst, mit der Zahl der Fälle als Literal (E1 bis E7); die zwei Mutationen des Nachweises sind rot gesehen; der Vermerk *Stand* des Nachweises in `LH-QA-07.a` ist im liefernden Commit nachgezogen.
- [x] `LH-FA-08`: der Schritt `ODV4-Q-11` b (`SPEC-049` Nr. 4) läuft in der Abnahme, über HTTP und mit ETag; die Zähler stehen auf 93 (Bereich Q 17) in `abnahme_matrix_test.go`, `abnahme_etag_test.go` und `vertrag_http_test.go` (Q1, Q2); die Vermerke *Stand* in `SPEC-049` Nr. 4 und `SPEC-042` Nr. 3 e sind im liefernden Commit nachgezogen.
- [x] `LH-RB-04`: das Benutzerhandbuch trägt den kurzen Absatz *Schutz der Daten* (H1, H2), geschrieben nach dem Nachzug des Architect auf `main` (§4); `make handbuch-check` und `make katalog-check` sind grün.
- [x] `make gates` grün (Stand `fa5637fa`; die Grenze dieses Belegs steht in §7 *Gate*).
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `tests/integration/erweiterungen_test.go` | neu | die fünf Fälle von `LH-QA-07.a` mit eigenem Referenzdienst; Happy Path von `LH-QA-07`. Er nutzt die Hilfen des Vergleichs (`vergleich_test.go`), soweit sie passen, ohne sie zu ändern |
| `src/bootstrap/abnahme_matrix_test.go` | update | Schritt `ODV4-Q-11` a und b; Zähler 93, Bereich Q 17, Kommentare |
| `src/bootstrap/abnahme_etag_test.go` | update | Zähler 93 |
| `src/bootstrap/vertrag_http_test.go` | update | Zähler 93 (`2*36+93`) |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-042-test-contract.md` | update | Vermerke *Stand* im liefernden Commit |
| `docs/user/benutzerhandbuch.md` | update | Absatz *Schutz der Daten*; **zuletzt**, in eigenem Commit, nach dem Nachzug des Architect (§4). Nicht im Block: `SPEC-128` Nr. 3 m schließt Pfade unter `docs/` aus |

Dateiname und Testnamen sind Vorschlag des Planners; wer sie ändert, zieht Tabelle und Block im selben Commit nach.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Planner aus `LH-QA-07.a`, `SPEC-049` Nr. 4 und der Antwort 7 des Users, weil der Slice keinen neuen Vertrag liefert; der Implementer trägt Mutationen und Raten mit Zahl der Prozesse, Last und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| E1 | Fall *Query*: `$apply`, `$search`, `$unbekannt` und `sap-client` erreichen den Referenzdienst unverändert, die Antwort gleicht der direkten | `TestRecordErweiterungen/Query` | der Record-Pfad weist eine Anfrage mit unbekannter Systemoption ab | andere |
| E2 | Fall *Pfad*: `/Products/Ns.Beste()` erreicht den Referenzdienst unverändert, die Antwort gleicht der direkten | `…/Pfad` | der Record-Pfad ändert oder verwirft das Segment der gebundenen Funktion | andere |
| E3 | Fall *Header*: herstellerspezifische Header erreichen in der Anfrage den Referenzdienst und in der Antwort den Client unverändert | `…/Header` | ein eigener Header der Anfrage entfällt; ein eigener Header der Antwort entfällt | andere |
| E4 | Fall *JSON*: eine Antwort mit unbekannten Instanz-Annotationen und eingebettetem `$expand` gleicht der direkten | `…/JSON` | die Antwort wird als JSON gelesen und ohne die Annotation neu geschrieben | andere |
| E5 | Fall *Metadaten*: `$metadata` mit Annotationen eines fremden Namensraums gleicht der direkten | `…/Metadaten` | der Body der Metadaten wird verändert | andere |
| E6 | Je Fall trägt die Session den Exchange mit `capture_status=complete` | Schleife über alle Fälle | der Exchange eines Falls wird nicht angehängt; er wird nicht als `complete` angehängt | andere |
| E7 | Die gelaufenen Fälle sind die fünf des Nachweises; Zahl 5 und Namen stehen als Literal | `TestRecordErweiterungen`. Akzeptiertes Negativ: den Inhalt eines Falls unter seinem Namen prüft die Zählung nicht (ein Fall `Pfad` mit anderer Anfrage bliebe grün), und ein einzeln mit `-run` gefahrener Teilfall macht die Zählung rot. Grund: die Zählung hält Namen. Folge: allein das Lesen der Tabelle belegt den Inhalt; 0 von 5 Inhalten sind gegen den Namen gehalten | ein Fall entfällt aus der Tabelle; ein Fall steht doppelt statt eines anderen | andere |
| E9 | Je Fall (alle fünf) erreichen Pfad, Query und die eigenen Header (je Fall `X-Vendor-Fall`) den Referenzdienst unverändert, und die Antwort gleicht der direkten (`LH-QA-07.a`, Satz „Je Fall“) | Schleife über alle Fälle (`erwPruefeGesehen`, `vergleiche`) | der Header `X-Vendor-Fall` entfällt; eine Query-Option wird verworfen | andere |
| E8 | Akzeptiertes Negativ (`LH-QA-07.a`): ein Body unbekannten Medientyps, auch `multipart/mixed` (`$batch`), wird weitergeleitet, der Exchange mit `capture_status=failed` ohne diesen Body angehängt (`SPEC-046` Nr. 3). Grund: ein unlesbarer Body darf nicht abgelegt werden (`LH-FA-14`). Folge: sichtbar, nicht wiedergebbar | Negativ: kein Fall dieses Nachweises (0 von 5; die Klasse der unbekannten Medientypen ist unbeschränkt). An anderer Stelle führt `src/hexagon/services/record/aufnahme_test.go` die Anfrage mit `text/plain` und `multipart/mixed` (`TestRequestAufnahme`: weitergeleitet, `capture_status=failed`, kein Match-Key, Grund `typ`) und die Antwort mit `application/xml` außerhalb von `/$metadata` (`TestResponseAufnahme`). Offen dort: diese Zeilen prüfen den abgelegten Body nicht („ohne diesen Body“), und den Weg über den Listener führen sie nicht | — | andere |
| E10 | Akzeptiertes Negativ (Reichweite des Nachweises): vom abgelegten Exchange liest der Test `capture_status`, Methode, Pfad und Status, nicht Header, Query oder Bodies; alle fünf Fälle sind `GET` ohne Request-Body. Grund: `LH-QA-07.a` verlangt vom Exchange allein `capture_status=complete`. Folge: ein Record-Pfad, der abgelegte Header, Query oder Bodies verliert, bleibt in diesem Test grün. Maß: 4 gelesene Felder; 0 von 5 Fällen mit Request-Body oder anderer Methode | Negativ. An anderer Stelle rot laut Review (Mutation X5, die Ablage verliert Antwort-Header): `services/record`, `TestDurchstichM2`, `TestTextRoundTrip` | — | andere |
| Q1 | `ODV4-Q-11` b (`$filter=Active ne false&$orderby=Stock&$skip=1&$top=2&$count=true&$select=Id`): 200, `KK`, Auswahl `Id`, `@odata.count` 4, Entities 5, 3 in dieser Folge; jede Stufe wirkt (`SPEC-049` Nr. 4) | `TestAbnahmeMatrix`, `TestVertragMatrix`, `TestAbnahmeMatrixMitEtag` | Filter, Sortierung, `$skip` oder `$top` entfällt; Gegenrichtung der Ordnung; Paging vor dem Filter, vor der Ordnung; `$skip` vor dem Filter; `$top` ohne Wirkung oder um eins erhöht bei gesetztem Filter; der Count zählt nach dem Paging oder vor dem Filter | andere |
| Q2 | Die Matrix führt 36 IDs mit 93 Schritten (Q: 17); ein nicht gelaufener Schritt ist rot (`SPEC-049` Nr. 3 e) | `TestAbnahmeMatrixZaehlung`, `TestVertragMatrixZaehlung`, `TestAbnahmeMatrixMitEtag` | Schritt b entfällt; ein Zähler bleibt auf 92 | andere |
| H1 | Handbuch, Absatz *Schutz der Daten*: Core-Dumps und Swap sind Pflicht des Betreibers; der Recorder prüft sie nicht (Wortlaut nach dem nachgezogenen `SPEC-053`) | Negativ mit Grund: Prosa wird nicht geprüft (`SPEC-200` Nr. 14 b); Reviewer und Verifier lesen gegen die Spezifikation. Folge: ein falscher Satz bleibt grün. `make handbuch-check` fängt allein Ausblick, Chronik und Kennungen, `make katalog-check` unbekannte Namen | eine Kennung der Spezifikation im Absatz ist rot (`make handbuch-check`) | andere |
| H2 | Handbuch, derselbe Absatz: Werte, die ein Client im Modus `mock` oder `hybrid` schreibt, speichert der Dienst unverändert; welche Testdaten zulässig sind, verantwortet der Betreiber (`SPEC-053`, *Handbuch*: Kapitel `Schutz der Daten` mit drei Aussagen) | wie H1. Akzeptiertes Negativ (Verifikation V-4): für eine Änderung, die im Modus `hybrid` in den Replay- oder den Upstream-Fallback geht, ist der Satz nicht gefahren; nach dem Kapitel `Modus hybrid` führt sie nicht der Mock aus. Grund: der Wortlaut folgt `SPEC-053`, und gesehen ist er für Änderungen, die der Mock ausführt. Folge: der Satz ist dort weiter als das Gesehene und warnt eher zu viel. Maß: 0 von 2 Fallback-Wegen gefahren | wie H1 | andere |

**Bericht der Mutationen** (Implementer, 2026-11-27; `AGENTS.md` §3.10). Je Mutation eine Wegwerfkopie von `src` und `tests` in einem Container ohne Netz (Image der Stufe `source`, `--memory=4g --cpus=4`), ein Prozess `go test -count=1`; kein Fänger hängt an Nebenläufigkeit, darum steht keine Rate. Last des Hosts: Eins-Minuten-Wert 5,8 vor und 3,8 nach der Reihe (29 Mutationen). Alle 29 sind rot; der Grundlauf ohne Mutation ist grün, `TestRecordErweiterungen` mit `-count=20` ebenso.

| Zusage | Mutation | roter Test |
|---|---|---|
| E1 | `$apply=` weist der Record-Pfad mit `ODR-E5006` ab; er verwirft `$unbekannt` auf dem Weg zum Upstream | `TestRecordErweiterungen/Query` |
| E2 | der Pfad zum Upstream verliert `()` | `…/Pfad` |
| E3 | ein eigener Header der Anfrage entfällt; ein leerer Wert entfällt; die Feldzeilen werden gedreht; ein Header der Antwort entfällt | `…/Header` |
| E4 | die Antwort auf `/Orders` wird als JSON gelesen und ohne `@com.vendor.Stand` neu geschrieben | `…/JSON` |
| E5 | der Body der Metadaten wird verändert (gleiche Länge) | `…/Metadaten` |
| E6 | der Exchange wird nicht angehängt (je einzeln bei `Pfad`, `JSON`, `Metadaten`); er wird `incomplete` angehängt (dieselben drei) | der jeweilige Teilfall |
| E7 | die Tabelle führt vier Fälle; `Query` steht doppelt statt `Pfad` | `TestRecordErweiterungen` |
| E9 | `X-Vendor-Fall` entfällt zum Upstream (rot in allen fünf Fällen); die Query-Option aus E1 | alle Teilfälle bzw. `…/Query` |
| Q1 | am Produktcode `FuehreAus` und `ordne`: Filter, Sortierung, `$skip`, `$top` entfällt; Gegenrichtung der Ordnung; Paging vor dem Filter; Paging vor der Ordnung; `$skip` vor dem Filter; `$top` bei gesetztem Filter ohne Wirkung; um eins erhöht; Count nach dem Paging; Count vor dem Filter | `ODV4-Q-11` b, je Mutation in `TestAbnahmeMatrix`, `TestVertragMatrix` und `TestAbnahmeMatrixMitEtag` |
| Q2 | Schritt b entfällt; Schritt b läuft nicht; `abPruefeZaehlung` auf 92; Bereich Q auf 16; Zähler in `abnahme_etag_test.go` auf 92; Zähler in `vertrag_http_test.go` auf 92 | je Mutation `TestAbnahmeMatrixZaehlung`, `TestVertragMatrixZaehlung`, `TestAbnahmeMatrix` oder `TestAbnahmeMatrixMitEtag` |

**Fix-Runde nach dem Review** (Implementer, 2026-11-27; F-1, F-2, F-3, F-5): Schritt b folgt der Query aus `3fb624d7`, Zähler weiter 93; E7 hält Zahl und Namen der Fälle. Die Zeilen Q1 und E7 der Tabelle sind an der neuen Fassung gefahren: 12 Mutationen zu Q1 und 2 zu E7, alle rot, Grundlauf grün (`TestAbnahme*`, `TestVertragMatrix*`; `TestRecordErweiterungen` mit `-count=20`), gleicher Aufbau, Messender Implementer, Last 12,7 zu Beginn und 4,9 am Ende. Die Mutationen der ersten Runde zu E1 bis E6, E9 und Q2 sind nicht neu gefahren: der geprüfte Code und ihre Fänger sind unverändert.

**Handbuch** (H1, H2; nach dem Merge von `main` mit dem Nachzug `62cfd7f8`): Kapitel `Schutz der Daten` mit den drei Aussagen von `SPEC-053`, *Stand* im Kopf des Handbuchs nachgezogen. Mutation H1 in einer Wegwerfkopie: eine Kennung der Spezifikation im Absatz ist rot (`make handbuch-check`: `Zeile 257: Kennung C6`); im Worktree sind `make handbuch-check` und `make katalog-check` grün. Die Sätze selbst bleiben ungeprüft (Negativ von H1).

**Messung** (`SPEC-128` Nr. 1): 235 Zeilen (Summe aus `git diff --numstat` von der Basis `20728b70` bis zum Handoff über die Pfade des Blocks), Marke 1600; der vorab benannte Schnitt (§4) greift nicht.

**Prüfung von Tabelle und Raster gegen den Vertrag** (Implementer, erster Schritt): keine Randform ohne Entscheidung und kein Fall, der eine Produktänderung verlangt; deshalb kein Anhalten. Berichtigt: zwei Zellen des Rasters nannten einen Ort, der ihre Entscheidung nicht trägt (*Anfrage aufzeichnen × leer*: `SPEC-121` Nr. 4; *Antworten vergleichen × Abbruch*: `LH-QA-07.a` Nr. 7 a); die Tabelle führte den Satz „Je Fall“ nicht als eigene Zusage (E9) und bei Q1 keine Mutation am Count; H2 folgt dem Nachzug von `SPEC-053` (`mock` oder `hybrid`).

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; „Nr.“ ohne Kennung meint `LH-QA-07.a`):

| Operation | fehlend | leer | unbekannt oder fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Anfrage aufzeichnen | `Accept-Encoding` fehlt: Nr. 5 | Header mit leerem Wert: `SPEC-121` Nr. 4 („auch leere Werte“) | die fünf Fälle des Nachweises; Body unbekannten Medientyps: akzeptiertes Negativ dort (E8) | gilt nicht: der Nachweis führt nur vollständige Austausche | gilt nicht: je Fall eine Anfrage |
| Antworten vergleichen | Header nur auf einer Seite: Nr. 1 | leerer Wert ist ein Wert: Nr. 1 | nicht angefragtes Coding: Nr. 4; nicht entpackbare Kette: Nr. 3 | Abbruch im Body: nicht verglichen, Nr. 7 a | gleiche Antwort auf beiden Seiten: Nr. 6 |
| Exchange der Session lesen | fehlender Exchange: rot (Satz „Je Fall“ des Nachweises) | wie *fehlend* | `capture_status` ungleich `complete`: rot (ebendort) | gilt nicht: gelesen wird nach der Antwort | gilt nicht: Lesen ändert nichts |
| Schritt `ODV4-Q-11` b | fehlender Schritt: rot, `SPEC-049` Nr. 3 e | gilt nicht: feste Anfrage | gilt nicht: feste Query; unbekannte Optionen tragen andere Fixtures | gilt nicht: im Prozess ohne Abbruch | je ID frische Umgebung: `SPEC-049` Nr. 3 b; Revision: Nr. 3 d dort |

**Betriebsregeln für alle Rollen:**

- Jeder schwere Lauf unter `flock /tmp/odata-recorder-heavy.lock`, jeder eigene `docker run` mit `--memory=4g --memory-swap=4g --cpus=4` (Übergangsregeln Nr. 1); Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, nie parallel zu `make gates`.
- Auf Repo-Dateien nur `Edit`/`Write`; im Scratchpad nur unter `abnahme-belege-<rolle>/` (Übergangsregeln Nr. 8).
- Fixtures mit Zeichen jenseits von ASCII oder Escape-Folgen entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); `harness/gegenproben.stempel` wird im Zweig nicht committet (Übergangsregeln Nr. 2).
- Eine Abweichung vom Plan oder eine Rückgabe steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): die Verträge stehen auf `main` (`2e289d4d`, Merge `6c1526bc`), und kein Slice in `in-progress/` ändert eine Datei des Blocks. Kein Kettenslice (Übergangsregeln Nr. 7): kein Pfad unter `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md` oder `tools/gegenproben/`.

**Stand beim Start** (Planner, 2026-11-27): Die Bedingung ist erfüllt. In `in-progress/` liegen daneben [slice-bootstrap-m6](../done/slice-bootstrap-m6.md) und [slice-lasttest-leser-zeitgrenze](../done/slice-lasttest-leser-zeitgrenze.md); `git diff --name-only main...slice-bootstrap-m6` nennt beim Start elf Pfade, darunter keinen des Blocks, keinen unter `spec/` und nicht das Handbuch; der Zweig des Lasttest-Lesers ändert allein `tests/lasttest/leser_test.go`. Für die Tests ist keine Frage an den Architect offen; der Nachzug für den Handbuch-Absatz stand beim Start aus und ist mit `62cfd7f8` geliefert (*Reihenfolge im Slice*, unten). Der Slice ist der dritte Worktree.

**Reihenfolge im Slice:** zuerst die Tests (E1 bis E8, Q1, Q2), zuletzt der Handbuch-Absatz (H1, H2). Der Absatz beginnt erst, wenn der Nachzug des Architect zu `SPEC-053` und zum Satz über die Werte im Modus `mock` auf `main` liegt und der Zweig `main` aufgenommen hat (Antwort 7; geliefert mit `62cfd7f8`, mit `main` aufgenommen). Er sagt mit H2 ein Verhalten des Binary in Prosa; `SPEC-200` Nr. 1 lässt das allein zu, wenn ein Vertrag des Produkts das Kapitel verlangt. Der Nachzug verlangt das Kapitel (`SPEC-053`). Er blockiert den Start der Tests nicht. **Vorab benannter Schnitt** (`AGENTS.md` §3.14): liegt der Nachzug bei sonst fertigem Slice nicht auf `main`, geht der Absatz (H1, H2) als eigene Sendung an den Planner; er ist zuletzt, in eigenem Commit, und teilt keinen Mechanismus mit den Tests.

**Größenmarke** (`SPEC-128` Nr. 1, `AGENTS.md` §3.14; Planner): zwölf Zeilen (E1 bis E8, Q1, Q2, H1, H2), *a* = 0, *b* = 12, *f* = 1 (der Referenzdienst), *p* = 0 (nur Test-Code): *S* = 80·12 + 200 = 1160, Marke 1600, keine Schnittfrage. Die Zeilen des Handbuchs zählt die Messung nicht (kein Pfad unter `docs/` im Block, `SPEC-128` Nr. 3 m). Der Teil des Blocks, den sonst der Architect setzt (`pfad`, `produktcode`, `fakes`), ist vom Planner gesetzt; `basis`, `schaetzung` und `marke` sind beim Start gesetzt, vor dem ersten Code-Commit. Ungeschnitten mit dem Absturztest: 31 Zeilen, *f* = 4, *S* = 3280, Marke 4500, über der Schnittgrenze. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `spec/` und `harness/vertraege/` zählen nicht. Jede weitere Datei nennt der Implementer im selben Commit im Block.

```groessenmarke
pfad: tests/integration/erweiterungen_test.go
pfad: src/bootstrap/abnahme_matrix_test.go
pfad: src/bootstrap/abnahme_etag_test.go
pfad: src/bootstrap/vertrag_http_test.go
produktcode: 0
fakes: 1
basis: 20728b70
schaetzung: 1160
marke: 1600
```

**Rückführungen — vorab benannt:**

- `in-progress` → `open` (blockiert): ein Fall oder der Schritt b zeigt eine Abweichung des Produkts von der Spezifikation, oder der Vergleich verlangt eine Entscheidung, die dort nicht steht (`AGENTS.md` §3.12). Der Slice ändert dann keinen Produktcode; er gibt an den Planner und den Architect zurück.
- `in-progress` → `next` (zu groß): die Messung zeigt, dass die Marke nur mit einem weiteren Mechanismus zu halten wäre.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

DoD vollständig, Review bestanden, `make gates` grün, je Zusage eine Mutation rot gesehen (`AGENTS.md` §3.10), Closure-Notiz mit Lerneintrag.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der Nachzug des Architect zu `SPEC-053` und zum Mock-Satz fehlt, oder er verlangt den Absatz nicht, den `SPEC-200` Nr. 1 für Verhalten in Prosa voraussetzt — **Ausgang:** entfallen: der Nachzug liegt auf `main` (`62cfd7f8`) und verlangt das Kapitel `Schutz der Daten` mit drei Aussagen (`SPEC-053`); der vorab benannte Schnitt (§4) greift nicht.
- Ein Test zeigt, dass das Produkt eine Erweiterung abweist oder verändert, oder dass `ODV4-Q-11` b nicht die erwarteten Entities liefert — **Ausgang:** entfallen: die fünf Fälle und der Schritt b sind ohne Änderung am Produkt grün (§3).
- Der Absatz *Schutz der Daten* nennt den Modus `hybrid` wie `SPEC-053`; der fachliche Owner hat `hybrid` nicht gesondert bestätigt (`SPEC-053`, *Vom Client geschriebene Werte*: „offen“) — **Ausgang:** weiter offen; Adresse: User (Frage (c) der [Roadmap](../in-progress/roadmap.md), Zeile M7; bei anderer Antwort ändern der Architect den Vertrag und danach das Handbuch); Register: `BEO-REPO/antwort-des-owners-auf-weniger-faelle-als-der-vertrag-traegt`.
- Neue Namen im Paket `integration` kollidieren beim Merge mit den Dateien von [slice-bootstrap-m6](../done/slice-bootstrap-m6.md) (`m6_test.go`, `m6_hilfen_test.go`, `referenzdienst_test.go`); die Datei ist verschieden, der Namensraum nicht — **Ausgang:** entfallen: `main` mit diesen Dateien ist aufgenommen, und das Paket baut und läuft grün (`make gates`).
- Schritt `ODV4-Q-11` b in der ersten Fassung ließ `$top` ohne Wirkung (Review F-1) — **Ausgang:** entfallen: der Architect hat die Query in `SPEC-049` Nr. 4 geändert (`3fb624d7`), der Test folgt ihr (§3, Fix-Runde).
- Der Nachweis liest vom abgelegten Exchange weder Header noch Query noch Bodies (Review F-4) — **Ausgang:** weiter offen als akzeptiertes Negativ (E10); Adresse: Architect, wenn `LH-QA-07.a` mehr verlangen soll; Register: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft`.
- Roadmap und [welle-betrieb](../welle-betrieb.md) führten die Sätze zu den zwei Betreiberpflichten als offen (Review F-6) — **Ausgang:** entfallen: auf `main` berichtigt und mit `34c574b5` aufgenommen.
- Der eigene Referenzdienst antwortet nicht auf beiden Seiten gleich (Zähler, Zeit-Header), und der Vergleich wird rot ohne Fehler des Proxys — **Ausgang:** entfallen: der Vertrag regelt die gleiche Antwort auf beiden Seiten (`LH-QA-07.a` Nr. 6), und die fünf Fälle sind mit `-count=20` grün (§3).
- Eine Mutation an E2 oder E5 ist am Produktcode nicht nachstellbar, weil der Record-Pfad Pfad und Metadaten nicht liest (`LH-QA-07.a`) — **Ausgang:** entfallen: beide sind am Weiterleiten nachgestellt und rot (§3).

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

**Closure** (Planner, 2026-11-27; Zweig `slice-abnahme-belege`, Messbasis `20728b70`). Start `428e9b85`, Tests `0947bdc1`, Handbuch und Plan `bb18ecd9`, Review-Report `0bfdaaef`, Änderung der Query durch den Architect `3fb624d7`, Fix-Runde `fa5637fa`, Verifikations-Report `382dc643`, Merge von `main` vor der Closure `ab610285` (V-2). Review (`2026-11-27-slice-abnahme-belege-review`: 0 HIGH, 1 MEDIUM, 2 LOW, 3 INFO; *nicht freigegeben bis F-1*) und Verifikation (`2026-11-27-slice-abnahme-belege-verifikation`: bestätigt mit Auflagen V-1 bis V-4, F-1 geschlossen) liefen je einmal. Die Auflagen: V-1 §6 führt allein die drei Ausgänge; V-2 `main` aufgenommen, die Gates auf dem gemergten Stand sind der Integrationsschritt; V-3 der Satz in §4 ist berichtigt; V-4 steht als akzeptiertes Negativ bei H2.

**Was geliefert ist** (kein Produktcode): `LH-QA-07`, zweiter Halbsatz: `TestRecordErweiterungen` mit den fünf Fällen (`tests/integration/erweiterungen_test.go`). `LH-FA-08`, Happy Path: der Schritt `ODV4-Q-11` b in der Abnahmematrix, die damit 93 Schritte zählt. `LH-RB-04`, die zwei Betreiberpflichten und die vom Client geschriebenen Werte: das Kapitel `Schutz der Daten` im Benutzerhandbuch. Am Binary ändert sich nichts.

**Größe** (`SPEC-128` Nr. 9). Stand *Ende*, Zeile des Programms (`make groessenmarke`, 2026-11-27, Plan in `in-progress/`, nach dem Merge `ab610285`): `Messpunkt slice-abnahme-belege.md: zeilen=14 nebenlaeufig=0 fakes=1 produktcode=0 schaetzung=1320 marke=1600 messung=235`. Produktcode von Hand: 0. Stand *Handoff*: keine Zeile des Programms eingetragen; von Hand `20728b70..bb18ecd9` über die Pfade des Blocks 235. Auslastung 235 ÷ 1600 = **0,15**, Ende ÷ *S* der Tabelle (1320) = 0,18, Ende ÷ Handoff = 1,00. Der Block führt *S* = 1160 aus den zwölf Zeilen beim Start; die Zeilen E9 und E10 kamen danach, das Programm gibt dazu einen Hinweis. Fremde Zeilen 0. **Meldung an den Architect** (Auslastung unter 0,45): er rechnet die Faktoren aus allen Messpunkten neu (`SPEC-128` Nr. 2 und 9); der Punkt ist `zeilen=14`, `nebenlaeufig=0`, `fakes=1`, `produktcode=0`, *S* 1320, *M* 1600, Ende 235. *Ursache:* die fünf Fälle laufen als Tabelle durch einen Test mit den vorhandenen Hilfen des Vergleichs, und Q1 und Q2 sind ein Schritt und Zähler in der vorhandenen Matrix; von den vierzehn Zeilen tragen zehn Test-Code (E1 bis E7, E9, Q1, Q2; 207 Zeilen in der neuen Datei, 28 in den drei geänderten), vier tragen keinen (E8, E10 als Negative, H1 und H2 im Handbuch, das die Messung nicht zählt).

**Mutationen** (§3 und Reports): der Implementer 29 in der ersten Runde und 14 in der Fix-Runde (12 zu Q1, 2 zu E7), alle rot; der Reviewer neun eigene, davon vier grün (X1 und X2: F-1; X5: F-4; X9: F-5); der Verifier achtzehn eigene gegen die Klassen der Fix-Runde, alle rot. Kein Fänger hängt an Nebenläufigkeit, darum steht keine Rate. Last und Messende stehen in §3 und in den Reports.

**Akzeptierte Negative:** (1) E7: die Zählung hält Namen, nicht Inhalte der Fälle (§3); (2) E8: ein Body unbekannten Medientyps ist kein Fall des Nachweises; (3) E10: vom abgelegten Exchange liest der Test vier Felder; (4) H1 und H2: die Sätze des Handbuchs prüft kein Gate; (5) H2: die zwei Fallback-Wege im Modus `hybrid` sind nicht gefahren (V-4).

**Was weder Review noch Verifikation sahen:** (1) der Reviewer sah die Fix-Runde nicht, der Verifier sah sie; (2) die Mutationen X1 und X2 des Reviews fuhr nach der Fix-Runde niemand erneut; (3) das Image von `make build` startete niemand, der Verifier fuhr ein im Container gebautes Binary; (4) `make gegenproben` und `make test-race` fuhr der Verifier nicht; (5) die Texte der Closure (Risiko-Ausgänge, Negativ zu V-4, Register, Roadmap, Welle-Datei) hat niemand geprüft.

**Gate:** `.harness/state/gates-passed.head` im Worktree trägt `fa5637fa` (die Verifikation sah den Hash des Arbeitsbaums gleich; welche Stufen aus dem Schichten-Cache kamen, sagt der Stempel nicht, der Verifier fuhr die neuen Tests selbst). Die Commits danach sind von keinem Lauf der Kette gedeckt; `git diff --stat fa5637fa..ab610285` nennt allein Dateien unter `docs/`. Der Planner fuhr `make docs-check`, `make zeichen-check` und `make groessenmarke`, nicht `make gates` (Vorgabe des Orchestrators). Der Beleg auf dem Stand von `main` ist der Lauf des Integrationsschritts.

**Abweichungen vom Plan:** die Query von `ODV4-Q-11` b ist nach dem Review eine andere (F-1, `3fb624d7`); die Tabelle *Zusage · Fänger* bekam E9 und E10; der Implementer berichtigte zwei Zellen des Rasters, die einen Ort ohne die Entscheidung nannten. Keine Rückführung aus §4 griff, der vorab benannte Schnitt auch nicht.

**Steering-Loop-Einträge:**

1. **Neuer Sensor:** `ODV4-Q-11` b hält in der Abnahme (`make test`, über HTTP und mit ETag), dass Filter, Ordnung, `$skip`, `$top`, Auswahl und Count einer kombinierten Abfrage je wirken; `TestRecordErweiterungen` hält, dass unbekannte Erweiterungen aufgezeichnet werden.
2. **Lernsignal:** die erste Fassung des Schritts war mit wirkungslosem `$top` grün, weil der Filter drei Entities ließ und nach `$skip=1` genau die zwei übrig blieben, die `$top=2` erlaubt. Eine Fixture einer kombinierten Abfrage trägt die Zusage erst, wenn jede Stufe für sich das Ergebnis ändert.
3. **Architektur-Beobachtung:** die Mutationen X1 und X2 waren in den Tests von `model` und `services/mock` rot und in der Abnahme grün. Der Beleg einer Anforderung ist der Lauf, den die Abnahme führt; dass ein Test einer inneren Schicht die Wirkung fängt, belegt die Anforderung nicht.
4. **Schwellen-Übertritt:** `BEO-REPO/teil-zusage-ohne-isolierende-mutation` hat mit diesem Slice drei Belege. Der Slice liegt in keiner Welle, das Repo führt Wellen: der Ausgang (verkörpert, geplant oder gestrichen) ist Gegenstand des Lese-Schritts der nächsten Welle-Closure; bis dahin steht der Eintrag offen.
5. **Kein `liegt in`-Feld:** keine Regel ist mit diesem Slice verkörpert.

**Lerneintrag:** neuer Sensor (Eintrag 1) mit dem Lernsignal und der Architektur-Beobachtung (Einträge 2 und 3).

**Reviewer-Skill:** keine Änderung. F-1 trifft die MEDIUM-Zeile *Fänger schmaler als Zusage, Menge unbeschränkt oder ohne Achsenliste*; F-2 ist an `AGENTS.md` §3.15 gefunden (Maß), F-3 an §3.9; F-4 bis F-6 sind INFO. Zum Eintrag über der Schwelle (Eintrag 4) sind die Fragen der Pflege (Kategorie, Regel, Fitness Function) Teil des Lese-Schritts.

**Register** (Evidenz mit der Closure im Zweig; kein Zähler gesetzt, er folgt aus den Dateien): Evidenz `slice-abnahme-belege` zu sechs vorhandenen Einträgen: `teil-zusage-ohne-isolierende-mutation` (2 → 3 Dateien, F-1), `ein-negativ-benennt-die-ungefangene-klasse-zu-eng` (28 → 29, F-2), `berichtigung-erreicht-nicht-jede-stelle` (56 → 57, F-3, F-6, V-3), `menge-im-plan-weicht-vom-orakel-des-tests-ab` (9 → 10, F-5), `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (89 → 90, F-4) und `schaetzformel-ueberschaetzt-den-slice` (32 → 33). Zwei neue Einträge: `antwort-des-owners-auf-weniger-faelle-als-der-vertrag-traegt` (§6, `hybrid`) und `risiko-ausgang-ausserhalb-der-drei-werte` (V-1). 99 → 101 Einträge, 818 → 826 Evidenz-Dateien. `gate-lauf-stufen-aus-dem-schicht-cache`: keine Evidenz, ein Lauf aus dem Cache ist nicht belegt.

**Risiko-Ausgänge (§6):** 9 Punkte, je ein Ausgang: entfallen 7, eingetreten 0, weiter offen 2 (beide im Register).

**Sendungen (`AGENTS.md` §3.13):** *an diesen Slice:* die Sendung des Handbuch-Absatzes aus [welle-betrieb](../welle-betrieb.md) §4 (Antwort 7), geliefert. *Von diesem Slice:* an keinen Slice. *An den Architect:* die Meldung der Auslastung; E10, wenn `LH-QA-07.a` vom abgelegten Exchange mehr verlangen soll. *An den User:* die Frage (c) zu `hybrid` in der [Roadmap](../in-progress/roadmap.md).

**Folge-Slice:** keiner aus der Closure. [slice-absturztest](../done/slice-absturztest.md) stammt aus dem Schnitt vor dem Code und führt seinen Gegenstand selbst.

**Drei Paarungen:** Anker (kein Eintrag trägt ein Zielort-Feld) · Folge-Slice (`slice-absturztest` steht unter `open/`) · Register (die in §6 und §7 genannten Verzeichnisse tragen `evidence/`).

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

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, der Default `*` (Kürzel `REPO`); es gibt keine feinere.

**Vorgelagert — offene Beobachtungen sichten:** Register bei der Anlage und beim Start gesichtet (2026-11-27; der Zähler ist die Zahl der Dateien unter `evidence/`). Offen und mit dem Slice berührt: `BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache` (20×; ein grüner Lauf belegt die neuen Tests nur, wenn die Stufe sie wirklich fuhr), `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (17×; Mutationsläufe über `bootstrap`), `BEO-REPO/allaussage-an-jeder-stelle-mit-faenger-an-erster-und-zweiter-stelle` (5×; E6 läuft als Schleife über alle Fälle), `BEO-REPO/test-fremden-pakets-flackert-unter-last` (3×), `BEO-REPO/ausgabeformat-wechsel-zieht-tests-ausserhalb-des-blocks-mit` (4×; der Zähler 92 steht an mehreren Stellen), `BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger` (1×; Q2 hält den Zähler als Fänger). Keiner erreicht mit diesem Slice neu 3×. Verkörpert und zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (`AGENTS.md` §3.15; H1 und H2 sind benannte Negative). Geplant: `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×; der Block nennt allein Dateien).

Alle berührten Sub-Areas GF.
