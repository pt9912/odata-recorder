# Slice slice-abnahme-belege: Belege zur Erstabnahme: unbekannte Erweiterungen auf dem Record-Pfad und kombinierte Abfrage

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung über die DoD hinaus. Die Wellen, deren Gegenstand er belegt, liegen in `done/` (`welle-record-replay`, `welle-mock`), und kein Closure-Kriterium von [welle-betrieb](../welle-betrieb.md) nennt diese Belege (§3 dort). Gebraucht wird er von M7 (Erstabnahme): die [Roadmap](../in-progress/roadmap.md) führt ihn dort als Bedingung.

**Bezug:** `LH-QA-07`, `LH-FA-08`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-QA-07.a` (Nachweis für unbekannte Erweiterungen auf dem Record-Pfad, Vergleich Nr. 1 bis 6) · `SPEC-049` Nr. 3 und 4 (`ODV4-Q-11`) · [SPEC-042](../../../../harness/vertraege/SPEC-042-test-contract.md) Nr. 3 e (Zählung) · `SPEC-046` Nr. 3 (unbekannter Medientyp). Geändert werden allein die Vermerke *Stand* beim Liefern.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** — (bis zur Priorisierung).

**Autor:** pt9912 (Planner, aus der Bewertung der Beleg-Lücken durch den Architect, `2e289d4d`). **Datum:** 2026-11-27.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Zwei Belege, die zur Erstabnahme fehlen, entstehen als Tests ohne Produktcode: (1) unbekannte OData-Erweiterungen verhindern die Aufzeichnung nicht (`LH-QA-07`, zweiter Halbsatz; fünf Fälle nach `LH-QA-07.a`); (2) Filter, Sortierung und Paging kombiniert auf fünf Datensätzen (`LH-FA-08` Happy Path) als Schritt `ODV4-Q-11` b der Abnahmematrix, die damit 93 statt 92 Schritte zählt.

**Schnitt vor dem Code** (Planner, 2026-11-27): Die Bewertung des Architect nennt einen dritten Beleg, den Absturztest (`SPEC-048`, `LH-QA-03`). Er ist [slice-absturztest](slice-absturztest.md): ungeschnitten *S* = 3280 über der Schnittgrenze (§4); die Teile teilen keinen Mechanismus.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Absturztest (`SPEC-048`) — [slice-absturztest](slice-absturztest.md), Schnitt oben.
- Produktcode — die Belege prüfen Verhalten, das die Spezifikation als vorhanden beschreibt; zeigt ein Test eine Abweichung, greift die Rückführung in §4.
- Das byte-gleiche Binary (`LH-RB-03.a`) — [slice-container-image](slice-container-image.md), Zeile 20 dort.
- Neue Fälle in `make test-integration` (`tests/e2e/`) — der Nachweis läuft im Paket `integration` in `make test` (`LH-QA-07.a`); keine Datei der Gate-Kette.
- Ein Handbuch-Kapitel — kein Verhalten am Binary ändert sich.

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

- [ ] `LH-QA-07`: ein Test im Paket `integration` führt die fünf Fälle von `LH-QA-07.a` (Query, Pfad, Header, JSON, Metadaten) im Modus `record` über den Listener gegen einen eigenen Referenzdienst, mit der Zahl der Fälle als Literal (E1 bis E7); die zwei Mutationen des Nachweises sind rot gesehen.
- [ ] `LH-FA-08`: der Schritt `ODV4-Q-11` b (`SPEC-049` Nr. 4) läuft in der Abnahme, über HTTP und mit ETag; die Zähler stehen auf 93 (Bereich Q 17) in `abnahme_matrix_test.go`, `abnahme_etag_test.go` und `vertrag_http_test.go` (Q1, Q2).
- [ ] Stand-Vermerke im liefernden Commit nachgezogen: `LH-QA-07.a` (Nachweis), `SPEC-049` Nr. 4 (`ODV4-Q-11` b) und `SPEC-042` Nr. 3 e.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
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
| `tests/integration/erweiterungen_test.go` | neu | die fünf Fälle von `LH-QA-07.a` mit eigenem Referenzdienst; Happy Path von `LH-QA-07`. Er nutzt die Hilfen des Vergleichs (`vergleich_test.go`), soweit sie passen, ohne sie zu ändern |
| `src/bootstrap/abnahme_matrix_test.go` | update | Schritt `ODV4-Q-11` a und b; Zähler 93, Bereich Q 17, Kommentare |
| `src/bootstrap/abnahme_etag_test.go` | update | Zähler 93 |
| `src/bootstrap/vertrag_http_test.go` | update | Zähler 93 (`2*36+93`) |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-042-test-contract.md` | update | Vermerke *Stand* im liefernden Commit |

Dateiname und Testnamen sind Vorschlag des Planners; wer sie ändert, zieht Tabelle und Block im selben Commit nach.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Planner aus `LH-QA-07.a` und `SPEC-049` Nr. 4, weil der Slice keinen neuen Vertrag liefert, sondern den Nachweis eines bestehenden; der Implementer trägt Mutationen und Raten mit Zahl der Prozesse, Last und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| E1 | Fall *Query*: `$apply`, `$search`, `$unbekannt` und `sap-client` erreichen den Referenzdienst unverändert, die Antwort gleicht der direkten | `TestRecordErweiterungen/Query` | der Record-Pfad weist eine Anfrage mit unbekannter Systemoption ab | andere |
| E2 | Fall *Pfad*: `/Products/Ns.Beste()` erreicht den Referenzdienst unverändert, die Antwort gleicht der direkten | `…/Pfad` | der Record-Pfad ändert oder verwirft das Segment der gebundenen Funktion | andere |
| E3 | Fall *Header*: herstellerspezifische Header erreichen in der Anfrage den Referenzdienst und in der Antwort den Client unverändert | `…/Header` | ein eigener Header der Anfrage entfällt; ein eigener Header der Antwort entfällt | andere |
| E4 | Fall *JSON*: eine Antwort mit unbekannten Instanz-Annotationen und eingebettetem `$expand` gleicht der direkten | `…/JSON` | die Antwort wird als JSON gelesen und ohne die Annotation neu geschrieben | andere |
| E5 | Fall *Metadaten*: `$metadata` mit Annotationen eines fremden Namensraums gleicht der direkten | `…/Metadaten` | der Body der Metadaten wird verändert | andere |
| E6 | Je Fall trägt die Session den Exchange mit `capture_status=complete` | Schleife über alle Fälle | der Exchange eines Falls wird nicht angehängt; er wird nicht als `complete` angehängt | andere |
| E7 | Die Zahl der gelaufenen Fälle ist das Literal 5 | `TestRecordErweiterungen` | ein Fall entfällt aus der Tabelle | andere |
| E8 | Akzeptiertes Negativ (`LH-QA-07.a`): ein Body unbekannten Medientyps, auch `multipart/mixed` (`$batch`), wird weitergeleitet, der Exchange mit `capture_status=failed` ohne diesen Body angehängt (`SPEC-046` Nr. 3). Grund: ein unlesbarer Body darf nicht abgelegt werden (`LH-FA-14`). Folge: sichtbar, nicht wiedergebbar | Negativ: kein Fall dieses Nachweises | — | andere |
| Q1 | `ODV4-Q-11` b: 200, `KK`, Auswahl `Id`, `@odata.count` 3, Entities 3, 1 | `TestAbnahmeMatrix`, `TestVertragMatrix`, `TestAbnahmeMatrixMitEtag` | Paging vor dem Filter; Sortierung entfällt; Filter entfällt | andere |
| Q2 | Die Matrix führt 36 IDs mit 93 Schritten (Q: 17); ein nicht gelaufener Schritt ist rot (`SPEC-049` Nr. 3 e) | `TestAbnahmeMatrixZaehlung`, `TestVertragMatrixZaehlung`, `TestAbnahmeMatrixMitEtag` | Schritt b entfällt; ein Zähler bleibt auf 92 | andere |

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; „Nr.“ ohne Kennung meint `LH-QA-07.a`):

| Operation | fehlend | leer | unbekannt oder fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Anfrage aufzeichnen | `Accept-Encoding` fehlt: Nr. 5 | Header mit leerem Wert: Nr. 1 | die fünf Fälle des Nachweises; Body unbekannten Medientyps: akzeptiertes Negativ dort (E8) | gilt nicht: der Nachweis führt nur vollständige Austausche | gilt nicht: je Fall eine Anfrage |
| Antworten vergleichen | Header nur auf einer Seite: Nr. 1 | leerer Wert ist ein Wert: Nr. 1 | nicht angefragtes Coding: Nr. 4 | nicht entpackbare Kette: Nr. 3 | gleiche Antwort auf beiden Seiten: Nr. 6 |
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

**Größenmarke** (`SPEC-128` Nr. 1, `AGENTS.md` §3.14; Planner): zehn Zeilen (E1 bis E8, Q1, Q2), *a* = 0, *b* = 10, *f* = 1 (der Referenzdienst), *p* = 0 (nur Test-Code): *S* = 80·10 + 200 = 1000, Marke 1400, keine Schnittfrage. Der Teil des Blocks, den sonst der Architect setzt (`pfad`, `produktcode`, `fakes`), ist vom Planner gesetzt; `basis`, `schaetzung` und `marke` kommen beim Start dazu. Ungeschnitten mit dem Absturztest: 31 Zeilen, *f* = 4, *S* = 3280, Marke 4500, über der Schnittgrenze. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `spec/` und `harness/vertraege/` zählen nicht. Jede weitere Datei nennt der Implementer im selben Commit im Block.

```groessenmarke
pfad: tests/integration/erweiterungen_test.go
pfad: src/bootstrap/abnahme_matrix_test.go
pfad: src/bootstrap/abnahme_etag_test.go
pfad: src/bootstrap/vertrag_http_test.go
produktcode: 0
fakes: 1
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

- Ein Test zeigt, dass das Produkt eine Erweiterung abweist oder verändert, oder dass `ODV4-Q-11` b nicht die erwarteten Entities liefert — **Ausgang:** weiter offen: erste Rückführung in §4, Adresse Planner und Architect.
- Neue Namen im Paket `integration` kollidieren beim Merge mit den Dateien von [slice-bootstrap-m6](../in-progress/slice-bootstrap-m6.md) (`m6_test.go`, `m6_hilfen_test.go`, `referenzdienst_test.go`); die Datei ist verschieden, der Namensraum nicht — **Ausgang:** weiter offen; Adresse: Implementer (nimmt `main` vor dem Handoff auf, Übergangsregeln Nr. 5) und Integrationsschritt.
- Der eigene Referenzdienst antwortet nicht auf beiden Seiten gleich (Zähler, Zeit-Header), und der Vergleich wird rot ohne Fehler des Proxys — **Ausgang:** entschieden (Vertrag, `LH-QA-07.a` Nr. 6).
- Eine Mutation an E2 oder E5 ist am Produktcode nicht nachstellbar, weil der Record-Pfad Pfad und Metadaten nicht liest (`LH-QA-07.a`) — **Ausgang:** weiter offen; Adresse: Implementer (baut die Mutation am Weiterleiten, nennt sie im Bericht, `AGENTS.md` §3.10).

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

**Vorgelagert — Sub-Area-Wahl prüfen:** eine Sub-Area, der Default `*` (Kürzel `REPO`); es gibt keine feinere.

**Vorgelagert — offene Beobachtungen sichten:** Register bei der Anlage gesichtet (2026-11-27; der Zähler ist die Zahl der Dateien unter `evidence/`). Offen und mit dem Slice berührt: `BEO-REPO/gate-lauf-stufen-aus-dem-schicht-cache` (20×; ein grüner Lauf belegt die neuen Tests nur, wenn die Stufe sie wirklich fuhr), `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit` (17×; Mutationsläufe über `bootstrap`), `BEO-REPO/allaussage-an-jeder-stelle-mit-faenger-an-erster-und-zweiter-stelle` (5×; E6 läuft als Schleife über alle Fälle), `BEO-REPO/test-fremden-pakets-flackert-unter-last` (3×), `BEO-REPO/ausgabeformat-wechsel-zieht-tests-ausserhalb-des-blocks-mit` (4×; der Zähler 92 steht an mehreren Stellen), `BEO-REPO/anpassung-eines-bestehenden-tests-lockert-seinen-faenger` (1×; Q2 hält den Zähler als Fänger). Keiner erreicht mit diesem Slice neu 3×.

Alle berührten Sub-Areas GF.
