# Verifikationsbericht: slice-abnahme-belege (Belege zur Erstabnahme) — 2026-11-27

**Gegenstand:** HEAD `fa5637fa`, Zweig `slice-abnahme-belege`, gegen DoD und Plan `slice-abnahme-belege`. Genau eine Verifikation. · **Rolle:** Verifier · **Modell:** Sonnet 5.5

**Gelesen:** Plan (§1 bis §6, §8); `LH-QA-07`, `LH-FA-08`, `LH-RB-04`; `LH-QA-07.a`; `SPEC-049` Nr. 4; `SPEC-053`; `SPEC-046` Nr. 3; `SPEC-042` Nr. 3 e; `SPEC-200` Nr. 1; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); Report `2026-11-27-slice-abnahme-belege-review`.

## Verdikt

**Bestätigt mit Auflagen.** Die drei Liefer-Punkte der DoD sind erfüllt, die Tests belegen die Kriterien, die drei Aussagen des Handbuch-Kapitels sind wahr (Reichweite unten), kein Produktcode und keine Lockerung. F-1 des Reviews ist geschlossen. Die Auflagen betreffen allein die Closure und ändern keinen Code.

| ID | Kategorie | Auflage |
|---|---|---|
| V-1 | LOW | §6 des Plans führt zwei Ausgänge außerhalb der drei Werte der DoD (eingetreten / entfallen / weiter offen): *offen* (Modus `hybrid`, Frage an den User) und *entschieden (Vertrag)* (Referenzdienst). Vor der Closure umstellen. |
| V-2 | LOW | `main` (`6d1a7cbe`) ist 5 Commits voraus; `git merge-tree` meldet keinen Konflikt. Vor dem Merge nach `main` den Zweig nachziehen und die Gates auf dem Stand fahren, der gemergt wird. Auch dieser Bericht ändert den Baum-Hash: der Gate-Stempel gilt für `fa5637fa` ohne ihn. |
| V-3 | INFO | §4 *Stand beim Start* sagt weiter, der Nachzug für den Handbuch-Absatz „steht aus (unten)“; „unten“ (*Reihenfolge im Slice*) führt ihn als geliefert. Der Satz ist als Stand beim Start datiert; die Berichtigung zu Review F-3 hat nur die zweite Stelle erreicht. |
| V-4 | INFO | Der Handbuch-Satz zu `hybrid` folgt dem Wortlaut von `SPEC-053`. Gesehen ist er für Änderungen, die der Mock im Modus `hybrid` ausführt. Eine Änderung, die in den Replay- oder Upstream-Fallback geht, führt nach dem Kapitel `Modus hybrid` nicht der Mock aus; der Satz ist dort weiter als das Gesehene und warnt eher zu viel (nicht gefahren). Die Frage an den Owner zu `hybrid` steht in §6. |

## DoD

| DoD-Punkt | Befund | Beleg |
|---|---|---|
| 1 `LH-QA-07`, fünf Fälle | erfüllt | `TestRecordErweiterungen` selbst gefahren (`-count=3`, fünf Teilfälle PASS). Am Quelltext: Modus `record` über den Listener, eigener Referenzdienst, Pfad mit Query byte-gleich, eigene Header, Vergleich nach `LH-QA-07.a`, `capture_status=complete`, Literal 5 mit Namen. Die zwei Mutationen des Nachweises rot (E-g; E-e und E-f unten). *Stand* im liefernden Commit `0947bdc1` |
| 2 `LH-FA-08`, `ODV4-Q-11` b | erfüllt | `TestAbnahmeMatrix`, `TestVertragMatrix` (HTTP), `TestAbnahmeMatrixMitEtag` und beide Zählungen selbst gefahren, grün, kein SKIP; Erwartung gegen die Query der Spezifikation (Count 4, Entities 5, 3). Zähler 93 und Q 17 in den drei Dateien; *Stand* in `SPEC-049` Nr. 4 und `SPEC-042` Nr. 3 e in `0947bdc1` |
| 3 `LH-RB-04`, Handbuch | erfüllt | Kapitel `Schutz der Daten` trägt die drei Aussagen von `SPEC-053`, *Handbuch*; `bb18ecd9` nach dem Merge des Nachzugs; Prüfung am Binary unten. `make handbuch-check` und `make katalog-check` über den Gate-Stempel |
| `make gates` | erfüllt, nicht neu gefahren | `.harness/state/gates-passed.head` ist `fa5637fa`, `gates-passed.diffsha` gleicht dem selbst gerechneten Hash des sauberen Baums; Stempel 23:17, Commit 22:56. Der Stempel sagt nicht, welche Stufen aus dem Cache kamen; die neuen Tests sind deshalb oben selbst gefahren |
| Review | Report liegt vor | Urteil dort *nicht freigegeben bis F-1*; ein zweites Review gibt es nicht, die Auflösung ist unten geprüft |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen (Closure) | §7 leer, Häkchen der DoD nicht gesetzt; V-1 |

## Plan gegen Code

- `git diff main...HEAD` (Basis `34c574b5`): neun Dateien, alle in §3 oder Report und Plan selbst. Kein Produktcode; unter `src/` allein drei `_test.go`.
- Größe: `git diff --numstat 20728b70..HEAD` über die vier Pfade des Blocks gibt 235, wie in §3, Marke 1600.
- Lockerung: keine. Keine Änderung an Gate-Konfiguration, `Makefile`, `harness/mk/`, `Dockerfile` oder `tools/` im eigenen Diff; kein neues `nolint`, `t.Skip` oder `d-check:ignore` in Code oder Handbuch; an bestehenden Tests ändern sich Zähler (92 auf 93, Q 16 auf 17), die Fixture von `ODV4-Q-11` und Kommentare.
- Tabelle *Zusage · Fänger*: E8 am Quelltext gelesen (`TestRequestAufnahme` mit `text/plain` und `multipart/mixed`, `TestResponseAufnahme` mit `application/xml`; der abgelegte Body wird dort in diesen Zeilen nicht geprüft, wie behauptet). E10 und H1/H2 sind benannte Negative.
- Entscheidungen: [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) (Tests in Go, `go.mod` unverändert).

## Fix-Runde gegen die Klassen (`AGENTS.md` §3.17)

Eigene Mutationen in einer Wegwerfkopie, je ein Lauf mit `-count=1`, kein Fänger hängt an Nebenläufigkeit. Alle 18 rot, keine ungültig; Grundlauf grün.

| Klasse | Mutation | Ergebnis |
|---|---|---|
| F-1, jede Stufe der kombinierten Abfrage wirkt (`FuehreAus`, je bei gesetztem Filter) | V1 `$top` vor `$skip`; V3 `$skip` ohne Wirkung; V6 `$select` vor `$orderby`; V7 `$orderby` ohne Wirkung | rot allein in `ODV4-Q-11` (b), in allen drei Matrix-Läufen |
| ebenso | V5 Count gleich Treffer minus `$skip` (ohne Bedingung an den Filter) | wie oben |
| ebenso | V2 `$top` um eins verringert; V4 `$select` ohne Wirkung; V8 Filter verwirft `null` | rot, `ODV4-Q-11` neben weiteren IDs (Ausgabe bei V2 und V4 gekürzt) |
| F-5, Zählung hält die Fälle | T1 `Metadaten` heißt `JSON`; T2 sechster Fall; T3 `Header` umbenannt | rot, `TestRecordErweiterungen` |
| `LH-QA-07.a`, an Elementen, die Plan und Review nicht nennen | E-a `$search` und E-b `$expand` fehlen zum Upstream; E-c `Sap-Message` und E-d der leere Antwort-Header entfallen; E-e der Fall `Header` wird `failed`; E-f der Fall `Query` wird nicht angehängt; E-g `$search` wird mit `ODR-E5006` abgewiesen | rot, je im betroffenen Teilfall |

F-1: geschlossen; die Query steht in `SPEC-049` Nr. 4 geändert (`3fb624d7`), `$top` schneidet im Schritt b eine Entity ab. Die Mutationen X1 und X2 des Reviews selbst sind nicht neu gefahren. F-2 (Maß und Fänger an E8), F-4 (E10, Adresse Architect), F-5 (Namen als Literal), F-6 (Roadmap und `welle-betrieb` §4 auf dem Stand) geschlossen; F-3 siehe V-3.

## Handbuch am gebauten Binary

Binary aus HEAD mit den Flags der Stufe `build` im Container gebaut, nicht das Image von `make build`. Weg: `record` gegen einen eigenen Upstream, Session beenden, Import, Workspace, dann `mock` und `hybrid` (ohne Fallback), je mit `REDACT_JSON_PATHS=$..Name`.

| Aussage | Befund |
|---|---|
| Der Recorder prüft Core-Dumps und Swap nicht | Start in `record`, `mock` und `hybrid` mit unbegrenzter Größe der Core-Datei, keine Zeile dazu im Log; unter `src/` kein Code zu Core-Dumps oder Swap (Suche; ein Treffer allein zum Heap-Wächter einer Test-Suite). Ein abweichender Swap-Zustand ist nicht hergestellt worden |
| Werte aus `mock` oder `hybrid` unverändert gespeichert | `POST`, `PATCH`, `PUT` je Modus: `GET` liefert den geschriebenen Wert; der letzte Wert steht im Klartext in `odata-recorder.db` (die früheren sind überschrieben). Gegenprobe: der aufgezeichnete Wert unter demselben Pfad steht nach `record` in keiner Datei von `DATA_DIR`. Snapshots nicht geprüft |
| Zulässige Testdaten verantwortet der Betreiber | normativ, gleich `LH-RB-04` |

## Nicht gefahren

`make gates` (Stempel), `make gegenproben`, `make test-race`, `make handbuch-check`, `make katalog-check`, die Mutationen des Implementers, `-count=20`.

**Lauf:** ein Container aus `odata-recorder:test` (`--network none --memory=4g --memory-swap=4g --cpus=4`, kein Mount) unter `flock /tmp/odata-recorder-heavy.lock`, 23:31 bis 23:35; beim Start nach dem Warten 13 GiB verfügbar, Last 5,9, am Ende 4,4; Messender Verifier. Danach allein `make docs-check` für diesen Bericht.

**Übergabe an den Planner:** V-1 und V-2 vor Closure und Merge; V-3, V-4 zur Kenntnis.
