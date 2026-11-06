# Review-Bericht: slice-audit-dienste — 2026-11-06

**Rolle:** Reviewer (Modul 10) · **Frage:** Ist der Diff gegen Plan, Entscheidungen und Hard Rules wartbar und richtig? — nicht gegen die DoD (Verifier).

**Review-Art:** Code (gegen Plan und Hard Rules).

**Gegenstand:** `slice-audit-dienste`, Zweig `slice-audit-dienste`, HEAD `a888bd2`, Diff `1543b02..a888bd2` (25 Dateien, 1960 Einfügungen, 68 Löschungen). Bezug: `LH-RB-04`, `LH-FA-13`, `LH-QA-04`, [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md).

**Skill:** `.harness/skills/reviewer.md` @ `a888bd2` · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-06

**Eingangs-Kontext:** Plan `slice-audit-dienste` (§1, §3 mit Tabelle E01 bis E19, Mengen M1 bis M7 und Mutationsabsatz, §4, §6 mit Raster), `SPEC-234` Nr. 1 bis 11 und *Ungeprüft*, `SPEC-139` Nr. 2 und 3 (Brücke), `SPEC-210` (Werkzeug), `AGENTS.md` §3.7, §3.9, §3.10, §3.11, §3.13, §3.15, §3.16, §3.17, §3.18.

**Methode.** Einmal gefahren (Durchstich vor Härtung). Der Bericht des Implementers ist nicht übernommen; seine Zählung (104 Mutationen) ist nicht nachgefahren. Ich habe 35 eigene Mutationen an Elementen gefahren, die Plan und Implementer nicht nannten, mit `tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, `--memory 2g` (Treiber-Vorgabe), `--cpus=4`, je Liste unter `flock /tmp/odata-recorder-heavy.lock`, ein schwerer Lauf zugleich. Messender: der Reviewer-Agent (Sonnet 5.5). Last (erster Wert von `/proc/loadavg` des Hosts, 4 CPUs des Containers): Liste A (28 Mutationen) 4,5 bis 2,9; Liste B (7 Mutationen) 14,5 bis 24,6; Wiederholung von R-B3 10,6. Die Last steht in jeder Ergebniszeile des Werkzeugs; ein Rot unter Last beweist kein Rot auf ruhigem Host (`SPEC-210` Nr. 12 c). Kontrollen: die Nullprobe des Werkzeugs (unveränderter Stand grün, `-test.list` trifft genau einen Test) und die Mutationen R-S1, R-S2, R-M2, R-W1, R-W2 (Art falsch, doppelter Eintrag, dritter Eintrag fehlt), die alle rot wurden; das Werkzeug sieht also ein Rot. Aufräumen: die Wegwerfkopien liegen im Container und gehen mit ihm; nach den Läufen blieb kein Container mit dem Label `odata-recorder.mutationslauf=1`; der Baum des Zweigs blieb bis auf diesen Bericht unberührt (`git status`). `make docs-check` (0 Befunde) und `make zeichen-check` (ok) liefen nach dem Schreiben dieses Berichts grün.

## Ergebnis der eigenen Mutationen

Zeile = Zeile der Tabelle *Zusage · Fänger*, Test = der Test, den die Liste nannte; `gruen` heißt: der benannte Test **und** das ganze Paket blieben grün (Stufe 2 des Werkzeugs).

| Nr. | Zeile · Achse | Mutation (Wegwerfkopie) | Ergebnis |
|---|---|---|---|
| R-S1 | E05 · Art | Session: `ArtSessionAufbewahrung` statt `ArtSessionGeloescht` | rot |
| R-S2 | E05 · Zahl | Session: zweiter `Append` mit gleichem Eintrag vor `CleanupBlobs` | rot |
| R-S3 | E05 · **Zustand** | Session: Eintrag nur bei `s.Status == SessionComplete` | **gruen** (Paket grün) |
| R-S4 | M5 · Länge | Session: `SessionID` auf 32 Bytes gekürzt | rot |
| R-S5 | M5 · Länge | Session: `ServiceID` auf 32 Bytes gekürzt | rot |
| R-S6 | E16 · Nebenläufigkeit | Session: `v.k.Frist++` im gemeinsamen Pfad der beiden `Delete` | Detektor 20/20 rot, ohne 0/20 (40 Prozesse, `flatternd`) |
| R-M1 | E05/M1 · **Dienst** | Aufbewahrung: `dienste[0].ID` statt `d.ID` im Aufruf von `loesche` | **gruen** (Paket grün) |
| R-M2 | E11/M7 · Position | Aufbewahrung: Eintrag nur bis zur zweiten Löschung | rot |
| R-M3 | M7 · Menge | Aufbewahrung: Eintrag nur bis zur dritten Löschung (`<= 3`) | gruen |
| R-M4 | E19 · Position | Aufbewahrung: `Append` der **dritten** Löschung in einer Goroutine | Detektor **2/20**, ohne **12/20** (`flatternd`) |
| R-M5 | E19 · Position | Aufbewahrung: `Append` der **ersten** Löschung in einer Goroutine | Detektor 20/20, ohne 20/20 |
| R-M6 | M5 · Länge | Aufbewahrung: `SessionID` auf 32 Bytes gekürzt | rot |
| R-M7 | M5 · Länge | Aufbewahrung: `ServiceID` auf 32 Bytes gekürzt | rot |
| R-W1 | E05 · Art | Workspace: `ArtSessionGeloescht` | rot |
| R-W2 | E05 · Zahl | Workspace: zweiter `Append` | rot |
| R-W3 | M5 · Länge | Workspace: `WorkspaceID` auf 32 Bytes gekürzt | rot |
| R-W4 | E16 · Nebenläufigkeit | Workspace: `v.k.Log = v.k.Log.With("k", 1)` im gemeinsamen Pfad | Detektor 20/20 rot, ohne 0/20 (`flatternd`) |
| R-I1 | E09/E13 | Import: `ctx.Err() == nil` statt `!kontextFehler(err)` | rot |
| R-I2 | M5 · Länge | Import: `ImportID` auf 32 Bytes gekürzt | rot |
| R-I3 | M5 · Länge | Import: `ServiceID` auf 32 Bytes gekürzt | rot |
| R-I4 | E18/M6 | Import: Teil der `UpstreamURL` hinter dem `@` in `WorkspaceID` | `rot-paket`: `TestAuditKeinGeheimnis` grün, ein anderer Test rot |
| R-I5 | E18/M6 | Import: ganze `UpstreamURL` in `WorkspaceID` | rot |
| R-I6 | E05/M1 · **Dienst** | Import: `i.k.Dienste[0].ID` statt `a.ServiceID` im Eintrag | **gruen** (Paket grün) |
| R-I7 | E14 | Import: Eintrag nur, wenn der Fehler ein `*inbound.Fehler` ist | gruen (äquivalent, siehe I-2) |
| R-F1 | E04 · Nebenläufigkeit | Fake: `Eintraege` ohne Sperre | **Detektor 0/20, ohne 0/20** (`gruen`) |
| R-F2 | E04 · Nebenläufigkeit | Fake: `Aufrufe` ohne Sperre | **0/20, 0/20** (`gruen`) |
| R-F3 | E04 · Nebenläufigkeit | Fake: Schreibzugriff ohne Sperre im Zweig `Vor` gesetzt | **0/20, 0/20** (`gruen`) |
| R-F4 | E04 · Nebenläufigkeit | Fake: Schreibzugriff ohne Sperre im Zweig `Vor` nil | Detektor 20/20, ohne 0/20 (`flatternd`) |
| R-F5 | E03 · Paar | Fake: `appends++` nur bei `Fehler == nil` | **gruen** |
| R-F6 | E03 · Paar | Fake: `Vor` nur bei `Fehler == nil` | **gruen** |
| R-B1 | Brücke · Instanz | Root: eigene `audit.New` für `session.Konfig` | **gruen** (Paket `src/bootstrap`, 2 Prozesse) |
| R-B2 | Brücke · Instanz | Root: eigene `audit.New` für `maintenance.Konfig` | rot (`TestImportErhaeltWerte`) |
| R-B3 | Brücke · Pfad | Root: `<DATA_DIR>/blobs/audit.jsonl` statt `<DATA_DIR>/audit.jsonl` | **gruen** (zweiter Lauf, Last 10,6); im ersten Lauf (Last 16,9) `rot-paket` nur durch das Zeitlimit von 2 Minuten (`panic: test timed out`), nicht aus dem Pfad |
| R-B4 | Brücke · Uhr | Root: `clock.NewSystemuhr()` statt `h.uhr` für den Audit | **gruen** |
| R-B5 | Brücke · Instanz | Root: eigene `audit.New` für `workspace.Konfig` | rot (`TestWorkspaceErhaeltWerte`) |

Nebenläufige Mutationen: je 20 Prozesse mit Detektor (`-race`, `-test.count=1`) und 20 ohne. Alle Raten sind Größenordnungen unter Last des Hosts (siehe Methode), keine festen Werte.

---

## Befunde

**HIGH: keine.** Kein ADR-Verstoß, kein Geheimnis in einer Senke (die sieben Felder des Eintrags tragen nur Kennungen und einen Code, R-I5 wird rot), keine `//nolint`-Direktive im Diff, keine Datei der Gate-Kette berührt (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, `.a-check.yml`, `.golangci.yml` stehen nicht im Diff), keine Kommentar-Klasse verletzt (Stichprobe auf Konjunktiv und „früher“ in den neuen Kommentaren: keine Treffer).

### M-1 (MEDIUM): Die Service-Kennung der Aufbewahrung und des Imports ist nur mit einem Dienst gefangen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe; Fixture gleicht dem Wert der Mutation); `SPEC-234` Nr. 4 b und 4 d, Plan Zeile E05 und Menge M1/M5
- `pfad`: `src/hexagon/services/maintenance/startup.go` · `w.loesche(ctx, b, d.ID, s.ID)`; `src/hexagon/services/baselineimport/importer.go` · `ServiceID: a.ServiceID, ImportID: a.ImportID`
- `befund`: R-M1 (`dienste[0].ID` statt `d.ID`) und R-I6 (`i.k.Dienste[0].ID` statt `a.ServiceID`) bleiben grün, auch im ganzen Paket. Jeder Fall der 130 Services setzt genau einen Dienst, dessen Kennung gleich dem Wert der Mutation ist; die Tests der Aufbewahrung mit zwei Diensten (`startup_test.go`) lesen den Audit nicht. Bei zwei konfigurierten Diensten trüge der Eintrag der zweiten Session-Löschung den falschen Service.
- `verifizierbar`: ja (ein Fall mit zwei Diensten, Eintrag je Dienst)
- `klasse`: Fixture gleich Wert der Mutation (Achse Dienst nur mit einem Element)

### M-2 (MEDIUM): `SessionUseCase.Delete` ist nur für den Status `complete` gefangen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe), `LH-RB-04.a`; `SPEC-234` Nr. 4 a (jede nicht offene Session)
- `pfad`: `src/hexagon/services/session/loeschen.go` · `_ = v.k.Audit.Append(support.Eintrag{Art: support.ArtSessionGeloescht`
- `befund`: R-S3 (Eintrag nur bei `SessionComplete`) bleibt grün, im ganzen Paket. Jeder Fall der Datei `audit_test.go` legt die Session mit `model.SessionComplete` an; `incomplete` (die Folge jedes Abbruchs, Stufe 2 der Wartung) ist eine zweite löschbare Form und hat keinen Fall. Eine Regression ließe die Löschung einer `incomplete`-Session ohne Eintrag.
- `verifizierbar`: ja
- `klasse`: Zustand-Achse der Eingabe mit einem Element

### M-3 (MEDIUM): `FakeAudit` ist als „für gleichzeitigen Aufruf sicher“ zugesagt, gefangen ist nur `Append`

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.10, §3.15; `SPEC-234` Nr. 9, Plan Zeile E04
- `pfad`: `src/hexagon/ports/support/supporttest/audit.go` · `func (f *FakeAudit) Eintraege() []Eintrag` und `func (f *FakeAudit) Aufrufe() int` und der Zweig `if f.Vor != nil`
- `befund`: Ohne Sperre in `Eintraege` (R-F1), ohne Sperre in `Aufrufe` (R-F2) und mit einem Schreibzugriff im Zweig `Vor` gesetzt (R-F3) bleibt `TestFakeAuditGleichzeitig` grün: Detektor 0 von 20, ohne Detektor 0 von 20 (je 41 Prozesse). Der Test ruft nur `Append` gleichzeitig; `Eintraege` läuft erst nach `wg.Wait()`, und `Vor` ist nicht gesetzt. Der Zweig `Vor` nil dagegen fängt der Detektor (R-F4, 20 von 20), das Rennen ist also nur dort sichtbar, wo ein Lauf es ausführt.
- `verifizierbar`: ja (Goroutinen, die `Eintraege` und `Aufrufe` neben `Append` rufen, und ein Lauf mit gesetztem `Vor`)
- `klasse`: Fänger schmaler als Zusage (Zugriff, den kein Lauf ausführt)

### M-4 (MEDIUM): Die Brücke liefert Zusagen von `SPEC-139` Nr. 3 ohne Fänger (bewertet, nicht entschieden)

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.13 (offene Entscheidung steht in §6 mit Ausgang *offen* und Adresse — ist erfüllt), §3.15 (Zusage ohne Fänger); `SPEC-139` Nr. 3, `SPEC-234` Nr. 8
- `pfad`: `src/bootstrap/wurzel.go` · `h.audit, err = audit.New(audit.Config{Pfad: filepath.Join(k.DataDir, audit.DateiName), Uhr: h.uhr, IDs: h.ids, Log: h.log})`
- `befund`: Die Brücke setzt um, was die Spezifikation schon verlangt (eine Instanz für die vier Konfigs, `<DATA_DIR>/audit.jsonl`, die eine Uhr, der eine Generator, der eine Logger), und sie trägt keinen Fänger für drei dieser Zusagen: eigene Instanz für `session.Konfig` (R-B1), anderer Pfad (R-B3) und zweite Uhr (R-B4) bleiben grün; nur Wartung, Import und Workspace sind gefangen (R-B2, R-B5, über den Vergleich mit `h.wartung.Audit`). Die Fänger V01 bis V05 liegen bei `slice-audit-verdrahtung` (`open/`), das den Kern seines Ziels an die Brücke verliert, wenn sie bleibt. Die Entscheidung über Brücke, Schnitt oder Reihenfolge liegt beim Planner (§6, Ausgang *offen*); dieser Befund bewertet sie nur.
- `verifizierbar`: ja (R-B1, R-B3, R-B4 als Fall)
- `klasse`: Zusage ohne Fänger an der Naht zweier Slices

**Sachlage zur Brücke (ohne Wertung der Entscheidung):**
- Für die Brücke spricht: ohne sie scheitert `Build` an `New` (nil-`Audit`), 26 Tests unter `src/bootstrap/` und 19 unter `tests/integration/` sind rot (Plan §6, hier nicht nachgezählt), `make gates` hat keinen grünen Stand; `audit.New` berührt die Datei nicht und legt kein Verzeichnis an (`src/adapters/driven/audit/audit.go`), `Build` legt `audit.jsonl` also nicht an, wie `SPEC-139` Nr. 2 es zusagt; beide Commits sind getrennt (`wurzel.go` allein, `import_workspace_test.go` allein).
- Gegen sie spricht: sie ist 16 Zeilen in `src/bootstrap/` (`wurzel.go` und die Tabellen in `import_workspace_test.go`), die nicht unter den `pfad`-Zeilen des Blocks liegen (die Messung 1864 zählt sie nicht), sie ändert das Verhalten des Binary (ab jetzt schreibt eine Löschung `audit.jsonl` in `DATA_DIR`), und `src/bootstrap/` ist in §1 ausgeschlossen und mit `slice-bootstrap-m2` geteilt.
- Der Revert von `5b7636e` allein lässt `Build` wieder rot (der zweite Commit ist nur die Tabelle); „mit `git revert` rückgängig, ohne den Rest zu berühren“ stimmt für den Zweig, nicht für einen grünen Zwischenstand.

### M-5 (MEDIUM): Der Plan widerspricht sich zur Brücke an drei Stellen, die ihr nicht folgen

- `kategorie`: MEDIUM
- `quelle`: `AGENTS.md` §3.9 (der Plan folgt jeder Korrektur; wer eine Aussage berichtigt, zieht alle Stellen nach)
- `pfad`: `docs/plan/planning/in-progress/slice-audit-dienste.md` · „dieser Slice verdrahtet nichts und ruft den Adapter nie“ (§1 Ziel); „Dieser Slice ändert `src/bootstrap/` nicht.“ (§4 *Unabhängig von*); „dieser Teil führt weder Konfigurationsschlüssel noch Fehlercode noch Datei ein und verdrahtet nichts; in §7 steht, dass das Binary unverändert blieb“ (§2 Handbuch-Punkt)
- `befund`: Die Brücke ist in §1 *Ausdrücklich NICHT*, in §3 (Tabelle) und in §6 nachgezogen, diese drei Sätze stehen daneben unverändert und sagen das Gegenteil. Der dritte verlangt eine Zeile für §7, die mit der Brücke falsch wäre (das Binary schreibt ab jetzt `audit.jsonl`, wenn gelöscht wird); der Verifier läse die DoD gegen den Zweig.
- `verifizierbar`: nein (Plan-Text; kein Gate prüft Widerspruch im Plan)
- `klasse`: Plan folgt der Korrektur nicht an allen Stellen

### L-1 (LOW): `FakeAudit` führt die Paare Fehler × Zähler und Fehler × `Vor` ohne Fänger

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.16 (Zusatztabelle: Paar von Ursachen), `SPEC-234` Nr. 9, Plan Zeile E03
- `pfad`: `src/hexagon/ports/support/supporttest/audit.go` · `f.appends++` und `if f.Vor != nil {`
- `befund`: R-F5 (`appends` nur ohne `Fehler`) und R-F6 (`Vor` nur ohne `Fehler`) bleiben grün. Die Zeile E03 nennt „Fall mit gesetztem und mit ungesetztem Fehler“, der Test prüft mit gesetztem Fehler nur die Zahl der Einträge, nicht `Aufrufe()` und nicht `Vor`; kein Dienst-Test setzt `Fehler` und `Vor` zugleich. Folge: ein Fake, der bei gesetztem `Fehler` nicht zählt, ließe einen Test „`Append` wurde nicht gerufen“ falsch grün werden.
- `verifizierbar`: ja
- `klasse`: Paar von Ursachen am selben Schlüssel ohne Fänger (Fake)

### L-2 (LOW): E19 hält die Aufbewahrung nur an der zweiten Session; die dritte fängt es mit Wahrscheinlichkeit

- `kategorie`: LOW
- `quelle`: `AGENTS.md` §3.10 (Nebenläufigkeit gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird), §3.17 (Position als Achse); Plan Zeile E19 und Menge M7
- `pfad`: `src/hexagon/services/maintenance/audit_test.go` · `if zaehler != 2 {`
- `befund`: R-M4 (`Append` der dritten Löschung in einer Goroutine) wird mit Detektor in 2 von 20 und ohne in 12 von 20 Prozessen rot (`flatternd`), R-M5 (erste) in 20 von 20 und 20 von 20. Der Plan führt E19 für die Aufbewahrung ausdrücklich nur an der zweiten von drei; die Rate für die übrigen Positionen steht nicht im Plan. Die Mutation setzt einen zweiten Aufrufort voraus (der Code hat einen), das Szenario ist ein späterer Umbau der Schleife.
- `verifizierbar`: ja
- `klasse`: Fänger nur mit Wahrscheinlichkeit, Rate nicht benannt

### I-1 (INFO): Die Menge „Sessions der Aufbewahrung“ ist unbeschränkt, die Prüfgrenze steht nicht im Plan

- `kategorie`: INFO
- `quelle`: `AGENTS.md` §3.15, §3.17; Plan Menge M7
- `pfad`: `src/hexagon/services/maintenance/audit_test.go` · `func alteSessions(u *welt, svc string) []string`
- `befund`: R-M3 (Eintrag nur bis zur dritten Löschung) bleibt grün; alle Läufe führen höchstens drei Sessions und einen Dienst. Die Position (erste, mittlere, letzte) deckt die Schleife, die Grenze „drei“ ist nicht als akzeptiertes Negativ mit Grund und Folge genannt. Die Mutation ist künstlich; die Folge im Betrieb wäre nur eine Schleife mit Zähler, die der Code nicht hat.
- `verifizierbar`: ja
- `klasse`: unbeschränkte Menge ohne genannte Prüfgrenze

### I-2 (INFO): Der Rückfall `ODR-E9001` von `codeVon` ist über `Import` nicht erreichbar

- `kategorie`: INFO
- `quelle`: Plan Zeile E14 und Menge M3
- `pfad`: `src/hexagon/services/baselineimport/importer.go` · `func codeVon(err error) model.Code`
- `befund`: R-I7 (Eintrag nur bei `*inbound.Fehler`) ist äquivalent und bleibt grün: jeder Fehler der zwölf Stellen läuft durch `mitCode` und ist ein `*inbound.Fehler`. Die Menge M3 des Architect (Stelle × Code × Verpackung am Import) ist im Test auf zwei Ebenen verteilt (Stelle × zehn Fehler am Import, Verpackung nur an `CodeVon` über `export_test.go`); diese Abweichung der Form steht nicht in §6. Die Mutation „verpackter Fehler ergibt `ODR-E9001`“ ist auf Ebene `CodeVon` gefangen.
- `verifizierbar`: nein
- `klasse`: Menge anders geschnitten als vom Architect genannt, ohne Zeile in §6

### I-3 (INFO): Das Geheimnis-Fixture von E18 trägt nur das Userinfo der `UpstreamURL`

- `kategorie`: INFO
- `quelle`: `LH-QA-04`, `SPEC-234` Nr. 3; Plan Zeile E18, Menge M6
- `pfad`: `src/hexagon/services/baselineimport/audit_test.go` · `UpstreamURL: "https://nutzer:" + geheimnis + "@example.test/svc"`
- `befund`: R-I4 (ein Teil der URL ohne Userinfo in einem Feld) lässt `TestAuditKeinGeheimnis` grün; ein anderer Test des Pakets wird rot, weil das Feld `WorkspaceID` bei dieser Art verboten ist. Pfad oder Query der URL tragen im Fixture kein Geheimnis. Der Schutz ist doppelt (Feldmenge je Art), E18 allein ist schmaler als „kein Query, kein Token“.
- `verifizierbar`: ja
- `klasse`: Fänger schmaler als Zusage, durch einen zweiten Fänger gedeckt

---

## Zu den ausdrücklich genannten Punkten

1. **Brücke:** M-4 und M-5; Sachlage oben, Entscheidung beim Planner.
2. **Aufbewahrung, `ServiceID` = `dienstID`:** gegengelesen, `SPEC-234` Nr. 4 b. Der Code übergibt `d.ID` der Schleife über `Dienste.List`, die zugleich `Sessions.List(ctx, d.ID)` bestimmt; die Zusage stimmt. Der Fänger ist mit einem Dienst gefangen (Länge und Zeichen über die 65 Zeichen), mit mehreren nicht (M-1).
3. **`UpstreamURL`-Quelle nur im Import:** nur `baselineimport` hat die `UpstreamURL` (die drei anderen Konfigs tragen keine); E18 setzt sie dort, R-I5 wird rot, R-I4 nur im Paket (I-3).
4. **E19 im Import hält nur an `Commit`:** der Aufrufort von `Append` ist einer (nach `neuerImport`), die zwölf Fehlerschritte laufen vor ihm; ein Halt an einer anderen Stelle prüfte denselben Aufruf. Kein Befund, aber der Plan-Wortlaut „Schleife über die vier Stellen“ trägt für den Import eine Stelle von zwölf (Negativ nicht genannt, Folge keine).
5. **Rohbytes (`AGENTS.md` §3.18):** `src/hexagon/ports/support/supporttest/audit_test.go` gegengelesen (`LC_ALL=C grep -P`): die Werte mit Zeichen jenseits von ASCII stehen als Escape von Bytewerten (`dienst` mit zwei Bytes, `geändert`), keine Umsetzung in ein sichtbares Zeichen. Dieser Bericht trägt keine solche Folge.
6. **`//nolint`:** keiner im Diff (`git diff` über die geänderten `.go`-Dateien).
7. **Gate-Kette:** unberührt (Dateiliste des Diffs: `src/hexagon/…`, `src/bootstrap/…`, der Plan, `spec/spezifikation.md`).
8. **Größenmarke:** `git diff --numstat 1543b02..a888bd2` über die fünf `pfad`-Zeilen: 1840 Einfügungen + 24 Löschungen = **1864**, Marke 3000, nachgerechnet. `src/bootstrap/` liegt außerhalb der Pfade (21 Zeilen nach `--numstat`, davon 16 die Brücke und 5 die Anpassung von `workspace_sqlite_test.go`; nicht gezählt).

## Nicht aufgelöst

Nicht mit eigener Mutation gefahren (der Bericht des Implementers ist hier nicht übernommen, die Zeilen sind also **ungeprüft**, nicht geprüft):
- Zeilen E01, E02, E06, E08, E10, E12, E15, E17 des Plans; E03 nur in den Paaren (R-F5, R-F6), nicht in den übrigen fünf Mutationen; E09 nur über R-I1, nicht über die 63 Teilmengen und die 20 Wiederholungen; E13 nur indirekt über R-I1; E16 nur mit Zählern im gemeinsamen Pfad (R-S6, R-W4), nicht mit dem Ablauf der Barriere.
- Menge M2 (15 Ablehnungen): keine eigene Mutation; M4 (60 Kontextformen): eine; M5: Achse Länge (32) und Dienst, **nicht die Achse Zeichen** (kein Zeichen ersetzt; die Schleifen vergleichen jedes der 37 und 65 Zeichen mit `strings.Repeat`, der Schluss auf Fänger ist Lesen, kein Rot); M6: Feld `WorkspaceID` und zwei Formen, nicht die übrigen sechs Felder.
- E19 für Session, Workspace und Import nur über die Mutationen des Implementers (Goroutine, `ctx.Done()`), keine eigene Fenster-Mutation außer R-M4 und R-M5.
- Die Raten der 104 Mutationen des Implementers (Absatz in §3) sind nicht nachgefahren.
- `make gates` nicht gefahren (Aufgabe des Verifiers).
- Die Last der Läufe in Liste B war hoch (14 bis 25 bei 4 CPUs); R-B1 und R-B4 sind unter dieser Last `gruen`, ein Grün wird durch Last nicht zum Rot, ein Wiederholungslauf auf ruhigem Host fand nur für R-B3 statt.

## Geprüft, ohne Befund

- `src/hexagon/services/session/` (`loeschen.go`, `verwaltung.go`): Reihenfolge Get, Delete, Append, Bereinigung; `ServiceID` aus `Sessions.Get`; keine Ablehnung schreibt; Länge und Zeichen gefangen — außer M-2.
- `src/hexagon/services/workspace/` (`loeschen.go`, `verwaltung.go`): kein Service im Eintrag, Reihenfolge, Länge gefangen; R-W4 vom Detektor in 20 von 20 gesehen.
- `src/hexagon/services/maintenance/` (`startup.go`, `wartung.go`): Eintrag je gelungener Löschung vor dem nächsten Port, Fehler beendet `Startup` ohne Eintrag; Länge, Position (R-M2) gefangen — außer M-1, L-2, I-1.
- `src/hexagon/services/baselineimport/` (`importer.go`): Eintrag nur im Fall „neu“ bei Fehler ohne Kontextfehler; Länge, Kontextprüfung (R-I1) gefangen — außer M-1, I-2, I-3.
- `src/hexagon/ports/support/supporttest/` (`audit.go`, `audit_test.go`): Kopie, Zähler, Haltepunkt vor dem Festhalten und ohne Sperre gefangen — außer M-3, L-1.
- `spec/spezifikation.md` (`SPEC-234` Nr. 5 d, Nr. 9, *Ungeprüft*, Historie): konsistent mit Code und Plan; keine Aussage der Architektur-Sicht berührt.
- `src/bootstrap/` (Brücke): M-4.
