# Review-Report: slice-abnahme-belege (Belege zur Erstabnahme: unbekannte Erweiterungen auf dem Record-Pfad, `ODV4-Q-11` b, Handbuch-Kapitel *Schutz der Daten*) — 2026-11-27

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf, gesammelt.

**Gegenstand:** `git diff 20728b70..bb18ecd9`, davon die eigenen Commits `428e9b85`, `6e47adc9`, `0947bdc1`, `bb18ecd9` (Zweig `slice-abnahme-belege`). Die Änderungen aus `main` im selben Bereich sind nicht Gegenstand; der Merge `aeeaf78c` trägt nach `git show --remerge-diff` keine eigene Änderung.

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`bb18ecd9`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-27

**Eingangs-Kontext:** Plan `slice-abnahme-belege` (§1, §3 Tabelle E1 bis E9, Q1, Q2, H1, H2, Raster, Bericht der Mutationen, §4, §6); `LH-QA-07.a`; `SPEC-049` Nr. 3 und 4; `SPEC-042` Nr. 3 e; `SPEC-046` Nr. 3; `SPEC-053`; `SPEC-121` Nr. 4; `SPEC-193` Nr. 3; `SPEC-200` Nr. 1; `LH-QA-07`, `LH-FA-08`, `LH-RB-04`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `AGENTS.md` (Hard Rules). Tabelle und Raster stammen vom Planner ohne Architect-Lauf; sie sind gegen den Wortlaut der Verträge gelesen. Die Mutationen des Implementers sind nicht übernommen.

**Urteil: nicht freigegeben bis F-1.** Kein HIGH, ein MEDIUM, zwei LOW, drei INFO. Am Produkt ist in diesem Lauf kein Fehler gefunden.

**Vorgehen.** Wegwerfkopie (`git archive HEAD`), per Tar-Strom in genau einen Container aus dem Image `odata-recorder:test` (`--network none --memory=4g --memory-swap=4g --cpus=4`, kein Mount), unter `flock /tmp/odata-recorder-heavy.lock`; beim Start 13 GiB verfügbar, Last 5,9, am Ende 6,1; Messender Reviewer. Je Mutation ein Lauf mit `-count=1`; kein Fänger hängt an Nebenläufigkeit.

- Grundlauf grün: `TestRecordErweiterungen` (auch `-count=10`), `TestAbnahme*` und `TestVertragMatrix*` in `src/bootstrap`.
- Probe zum Handbuch (eigener Test, nicht im Repo, über `bootstrap.Build` und den Listener): in den Modi `mock` und `hybrid` mit `REDACT_JSON_PATHS=$..Name` liefert `GET` den per `POST` geschriebenen Wert unverändert, und er steht im Klartext in `odata-recorder.db`. Das gebaute Binary selbst ist nicht gestartet.
- Nicht gefahren: `make gates`, `make gegenproben`, `make handbuch-check`, `make katalog-check`, `make test-race`.

**Eigene Mutationen:**

| Nr. | Mutation | Ergebnis |
|---|---|---|
| X1 | `$top` ohne Wirkung, wenn `$filter` gesetzt ist (`FuehreAus`) | **grün:** `TestAbnahmeMatrix`, `TestVertragMatrix`, `TestAbnahmeMatrixMitEtag`, ganz `tests/integration`; rot in `model` und `services/mock` |
| X2 | `$top` um eins erhöht, wenn `$filter` gesetzt ist | **grün:** wie X1; rot in `model` und `services/mock` |
| X3 | `$top` immer ohne Wirkung | rot: `ODV4-Q-11` und weitere IDs der drei Matrix-Läufe |
| X4 | `$skip` vor `$filter` | rot: `ODV4-Q-11` in allen drei Matrix-Läufen |
| X5 | die Ablage verliert alle Antwort-Header außer `Content-Type`, `capture_status` bleibt | **grün:** `TestRecordErweiterungen`; rot in `services/record`, `TestDurchstichM2`, `TestTextRoundTrip` |
| X6 | zwei Feldzeilen `X-Vendor-Folge` zum Client zu einer zusammengefasst | rot: `…/Header` |
| X7 | ein `%22` der Query zum Upstream dekodiert | rot: `…/Query` |
| X8 | `sap-client` zum Upstream verworfen | rot: `…/Query` |
| X9 | im Test steht statt des Falls `Pfad` ein zweites Mal `Query` | **grün:** `TestRecordErweiterungen` |

## Findings

Befunde ohne Lösungsvorschlag nach dem Schema des Skills.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Im Schritt `ODV4-Q-11` b hat `$top=2` keine Wirkung: der Filter lässt drei Entities, nach `$skip=1` bleiben genau zwei. X1 und X2 bestehen alle drei Läufe der Abnahmematrix. Versagen: der Beleg zu `LH-FA-08` Happy Path (Filter, Sortierung und Paging kombiniert) nimmt einen Mock ab, der in der kombinierten Abfrage zu viele Entities liefert; gefangen wird das allein außerhalb der Abnahme, in Tests von `model` und `services/mock`. Weder nennt Q1 eine Mutation an `$top` noch steht die Grenze als Negativ im Plan. Die Query steht wörtlich in `SPEC-049` Nr. 4; Adresse ist der Architect. | `LH-FA-08` (Happy Path), `SPEC-049` Nr. 4, `AGENTS.md` §3.15 | `src/bootstrap/abnahme_matrix_test.go` · `$skip=1&$top=2&$count=true&$select=Id` | ja: Mutationen X1, X2 (die Matrix-Läufe bleiben grün) | Fänger schmaler als Zusage (ein Parameter der Kombination ohne Wirkung) |
| F-2 | LOW | Das akzeptierte Negativ E8 nennt Grund und Folge, aber kein Maß (n von m) der ungefangenen Klasse. Seine Spalte *Fänger* sagt allein „kein Fall dieses Nachweises“ und nennt weder, dass `services/record` (`aufnahme_test.go`, `text/plain` und `multipart/mixed` im Request) das Weiterleiten mit `capture_status=failed` an anderer Stelle führt, noch was dort offen bleibt. | `AGENTS.md` §3.15 | Plan §3 · `Negativ: kein Fall dieses Nachweises` | nein | Akzeptiertes Negativ ohne Maß |
| F-3 | LOW | §4 (*Reihenfolge im Slice*) sagt weiter, der Nachzug des Architect sei „heute **nicht** geliefert“; §3 und §6 desselben Plans führen ihn als geliefert (`62cfd7f8`) und das Kapitel als geschrieben. | `AGENTS.md` §3.9 | Plan §4 · `heute **nicht** geliefert` | nein | Berichtigung erreicht nicht jede Stelle |
| F-4 | INFO | Reichweite des Nachweises: vom abgelegten Exchange liest der Test `capture_status`, Methode, Pfad und Status, nicht die abgelegten Header, Query oder Bodies (X5 bleibt in `TestRecordErweiterungen` grün, andere Tests fangen sie); alle fünf Fälle sind `GET` ohne Request-Body. Beides entspricht dem Wortlaut des Nachweises in `LH-QA-07.a`; Zuständigkeit Architect. | `LH-QA-07`, `LH-QA-07.a` | `tests/integration/erweiterungen_test.go` · `func erwPruefeSession` | ja: Mutation X5 | Grenze eines Nachweises nicht benannt |
| F-5 | INFO | Die Zählung (E7) hält die Zahl 5, nicht die fünf Fälle: ein Fall doppelt statt eines anderen bleibt grün (X9). `LH-QA-07.a` verlangt allein das Literal. Dieselbe Zählung macht einen mit `-run TestRecordErweiterungen/<Fall>` einzeln gefahrenen Fall rot, auch ohne Fehler (aus dem Code gelesen, nicht gefahren). | `LH-QA-07.a`, Maintainability | `tests/integration/erweiterungen_test.go` · `if gelaufen != 5` | ja: Mutation X9 | Zählung ohne Identität der Elemente |
| F-6 | INFO | Roadmap (Zeile M7, *Offen beim Architect*) und `welle-betrieb` §4 (Wirkung der Antwort 7, „offen, Adresse Architect“) führen die Sätze zu den zwei Betreiberpflichten noch als offen, obwohl `SPEC-053` sie trägt. Nicht Teil der Commits dieses Slice; Zuständigkeit Planner. | `AGENTS.md` §3.9 | Roadmap · `die Sätze zu den zwei Betreiberpflichten` | nein | Berichtigung erreicht nicht jede Stelle |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `tests/integration/erweiterungen_test.go` gegen `LH-QA-07.a` (fünf Fälle, Pfad, Query, eigene Header, Vergleich der Antwort, `capture_status=complete`) | geprüft; X6 bis X8 rot; ohne Befund außer F-4, F-5 |
| `src/bootstrap/` (Zähler 93, Bereich Q 17, Erwartung von Schritt b gegen die fünf Datensätze, Kommentare) | geprüft; X3, X4 rot; kein weiterer Zähler auf 92 gefunden; ohne Befund außer F-1 |
| `docs/user/benutzerhandbuch.md` (drei Aussagen gegen `SPEC-053`, Wahrheit gegen den Code, Ausblick, Chronik, Kennungen) | geprüft, ohne Befund: kein Code zu Core-Dumps oder Swap in `src/`; die Werte aus `mock` und `hybrid` unverändert (Probe). Der Satz zu `hybrid` folgt dem Vertrag; die Frage an den Owner steht in Plan §6 |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-042-test-contract.md` (Vermerke *Stand*) | geprüft, ohne Befund |
| Raster *Operation × Argumentart* in Plan §3 (Orte der Zellen) | geprüft, ohne Befund |
| Lockerung (`.golangci.yml`, `Makefile`, `harness/mk/`, `Dockerfile`, `.d-check.yml`, `.a-check.yml`, `tools/`, `//nolint`) | geprüft, ohne Befund: keine Änderung |
| `docs/plan/planning/` (Verweise nach `in-progress/`, Messung 235 Zeilen) | geprüft, ohne Befund außer F-3 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 2 |
| INFO | 3 |

**Finding-Klassen dieses Laufs:** Fänger schmaler als Zusage (ein Parameter der Kombination ohne Wirkung) · Akzeptiertes Negativ ohne Maß · Berichtigung erreicht nicht jede Stelle · Grenze eines Nachweises nicht benannt · Zählung ohne Identität der Elemente

## Verdikt

**Merge-blockierend:** ja — F-1 (MEDIUM). Der Ort der Query ist `SPEC-049` Nr. 4; die Rückkante geht an den Architect, nicht in den Code des Slice allein.

**Übergabe:** Findings an den Implementer, F-1 und F-4 zugleich an den Architect, F-6 an den Planner; die Finding-Klassen in die Slice-Closure §7. Die DoD-Konformität prüft der Verifier getrennt.
