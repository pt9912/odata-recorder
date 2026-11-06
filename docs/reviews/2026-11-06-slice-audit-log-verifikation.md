# Verifikations-Report: slice-audit-log (Teil A, Port und Adapter) — 2026-11-06

**Frage:** „Bauen wir es richtig?" — DoD- und Entscheidungs-Konformität, Plan gegen Code-Diff, Tabelle gegen Vertrag. Nicht das Review (Diff gegen Plan, Hard Rules) und nicht die Validierung.

**Gegenstand:** Zweig `slice-audit-log`, HEAD `d58715c` (Diff `1067fcd..d58715c`; Implementer `a9cf341`/`25158e3`/`49fa701`, Review `969f232`, Fix-Runde `6ebe88e`, Architect `4a1c5e2`/`d58715c`, Merge von `main` `39628c2`). Zweig-eigene Änderung gegen den Abzweig `e2b9c7e`: 10 Dateien, alle Audit (Plan, Review, `spec/spezifikation.md`, Port, Adapter, Tests).

**Eingang:** Plan `docs/plan/planning/in-progress/slice-audit-log.md` (DoD, §3 Tabelle A01–A21, Pflicht-Mengen, Raten, §4 Messung, §6), Vertrag `SPEC-234`, `SPEC-235`, `SPEC-123` Nr. 6/8 (Teil E, nur Naht), `LH-RB-04`, `LH-FA-13`, `LH-FA-14` (kein Geheimnis im Eintrag), Review `docs/reviews/2026-11-06-slice-audit-log-review.md`. Rollen-Skripte (`.harness/skills/reviewer.md`, `.claude/commands/implement-slice.md`): nach dem Merge von `main` unverändert gegenüber der Fassung des Reviews (`git diff 1067fcd..d58715c` über beide Pfade: leer).

**Rolle · Modell:** Verifier · claude-sonnet-5-5.

## Ergebnis in einem Satz

0 HIGH, 3 MEDIUM, 5 LOW, 5 INFO. Die DoD des Teils A ist bis auf die Closure-Pflichten erfüllt, der Code folgt `SPEC-234`/`SPEC-235` Nr. 1–7 Zeile für Zeile, die Raten der nebenläufigen Zeilen und der neuen Fänger halten bei eigener Messung. Die drei MEDIUM sind Fänger, die schmaler sind als die Zusage (Klasse, nicht die genannte Mutation, `AGENTS.md` §3.17) beziehungsweise eine Zusage ohne Fänger (`AGENTS.md` §3.15). **Empfehlung: Closure ja**, nachdem V1 bis V3 als Fänger (Schleifen, keine Produktänderung) oder als akzeptiertes Negativ mit Grund und Folge im Plan §3 stehen; kein neues Review nötig.

## Methode und Messbedingungen

- Wegwerfkopien des Zweigs unter dem Scratchpad `ver-al` (nicht `impl-al`/`rev-al`), Perl-Ersetzung im Text der Kopie, `go test -c` und Lauf des Testbinary im gepinnten Image (`golang:1.27-alpine`, Digest `8a5910f31396…`), netzlos, `--memory=2g`, `--user` des Hosts, unter `flock /tmp/odata-recorder-heavy.lock`. Der Baum des Zweigs blieb unverändert; Kopien und Binaries danach gelöscht.
- Grundlauf unverändert grün (BASE 0 von 3).
- **Rate** = Zahl roter Testbinary-Prozesse von N, je Prozess `-test.count=1`, nacheinander, `--cpus=4`. Messender: dieser Verifier. Last (Load-Average, 1 Minute) 8 bis 20 bei 20 Kernen, zwei weitere Agenten bauten und mutierten zugleich (einer hielt die Sperre der schweren Läufe über 25 Minuten).
- Statische Mutationen N = 1 (deterministisch).
- Docker-Speicher: `df` 29 GB frei, `free -g` 8 GB verfügbar vor den Läufen.

## Konformität gegen die DoD

| DoD-Punkt | Befund |
|---|---|
| Port `AuditLog`, `Eintrag`, `AuditArt`, `Arten()` (A01, A06–A09) | erfüllt. Typen und Tabelle stimmen mit `SPEC-234` Nr. 1–3 überein (Typzeile mit `AuditArt`, vier Konstanten, `Arten()` als Kopie, fünf Felder). Schärfe der Fänger: siehe V1, V3, V6, V8. |
| Adapter `audit.Datei` (A02–A05, A10–A21) | erfüllt gegen Nr. 1–7: Prüfung vor der Sperre, `Uhr.Now()` genau einmal unter der Sperre, ID genau einmal, Öffnen `O_RDWR\|O_APPEND\|O_CREATE` `0600`, Zeilenende-Prüfung, ein `Write`, `Sync` vor `nil`, `Close` in `defer`, Sperre auf jedem Pfad (`defer`), Log nach der Freigabe, Schlüsselmenge des Log-Eintrags nach Nr. 5. Kein Eintrag enthält Pfad, Fehlertext oder Wert eines ungültigen Eintrags. |
| `make gates` grün | siehe Abschnitt *Gates* unten. |
| Review durchgeführt, Report liegt vor | erfüllt: `docs/reviews/2026-11-06-slice-audit-log-review.md`, Skill von `main`. Seine sechs MEDIUM sind in `6ebe88e` behandelt (Fänger M1–M6, L1, L2). |
| Doku-Update öffentlicher Vertrag | erfüllt: `SPEC-234`/`SPEC-235` und der Nachzug stehen (Architect); `harness/README.md` nennt kein neues Gate, ein Gate änderte sich nicht. |
| Handbuch-Punkt | entfällt zu Recht: `src/bootstrap/` unberührt (`git diff e2b9c7e HEAD -- src/bootstrap` leer), kein Schlüssel, kein Code, kein Kommando; §6 trägt den Ausgang *weiter offen* mit Adresse `slice-admin-sessions`. In §7 muss bei der Closure stehen, dass das Binary unverändert blieb. |
| Closure-Notiz, Register, Ausgänge der Risiken, Paarungen | **offen, Closure-Pflicht**: §7 trägt „Wird bei der Closure gefüllt“. §6: Zeile *Pfad ohne Lauf unter dem Detektor* steht auf *offen (bei der Closure)*; mit A14/A15/A18–A20 und dem Negativ A21 kann sie *entfallen* oder *weiter offen* (Adresse Teil E für E13/E19) werden. Vier Zeilen stehen auf *weiter offen* mit Adresse (Fenster Löschung/Eintrag → `slice-audit-dienste`; Handbuch → `slice-admin-sessions`; Telemetrie-Abbildung → `slice-telemetry-einbau-gateway`), das ist ein zulässiger Ausgang. |

## Plan gegen Code-Diff

- `src/hexagon/ports/support/` (update) und `src/adapters/driven/audit/` (neu): wie geplant. `supporttest/` unberührt (FakeAudit gehört Teil E). Weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `.a-check.yml`, `tools/` noch `src/bootstrap/`, `tools/schema/schema.yaml` oder `src/adapters/driven/sqlite/` im Zweig-Diff: **Gate-Kette unberührt, SQLite-Entscheidung ([ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)) unberührt** (keine SQLite-Änderung, keine Migration; der Adapter importiert nur Standardbibliothek, `model` und `ports/support`).
- Messung §4: `git diff --numstat 1c99ede..d58715c` über die zwei Pfade = **1698** (Produktcode 285 = 222 + 63), Marke 3000: bestätigt, keine Überschreitung.
- Der Typname `AuditArt` steht in Spezifikation, Port, Adapter, Tests und Plan einheitlich; der Kommentar an `AuditArt` nennt den Grund („`Art` trägt im Paket den Typ der Attribute von `Telemetry`“), `Art` im Paket ist tatsächlich `type Art int` in `telemetry.go`: der Kommentar beschreibt, was da ist (§3.7).
- A08 im Plan nennt „Länge 0“ für alle fünf Felder; für `KorrelationsID` ist die Länge 0 kein Verstoß (`SPEC-234` Nr. 2). Der Implementer hat es unter *Befunde beim Schreiben der Fänger* (2) vermerkt, die Zeile der Tabelle blieb: LOW (V9).

## Entscheidungs-Konformität

- **Implementierungssprache** ([ADR-0001](../plan/adr/0001-implementierungssprache-go.md)): Go, nur Standardbibliothek. Erfüllt.
- **Architektur-Gate** ([ADR-0002](../plan/adr/0002-architektur-gate-a-check.md)): `make a-check` im Gate-Lauf; `.a-check.yml` deckt `src/adapters/driven/audit` mit `adapters-driven` ab. Die Import-Richtung (Adapter → `model`, `ports-support`) hält. Der Hinweis auf 55 Dateien unter `tools/` ohne Schicht ist vorbestehend (kein Teil des Zweig-Diffs).
- **SQLite-Bibliothek** ([ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)): keine SQLite-Änderung, `make schema-check` und `make migration-check` unberührt (Ablage ist eine Datei, `SPEC-235` *Entscheidung*).
- Entscheidung ohne ADR (Datei statt Tabelle, Eintrag nach der Löschung): im Vertrag mit Alternativen, Gründen und Folgen; kein Widerspruch zu `AGENTS.md` §3.8.

## Raten (selbst gemessen, 20 Prozesse `-test.count=1`, `--cpus=4`, Last 8 bis 20)

Ausgewählte Läufe: nur die Tests `TestGleichzeitige*`, `TestSperre*`, `TestAnhaengen`, `TestZeilenende`, `TestFenster*`, `TestLogNach*`.

| Zeile | Mutation (Verifier) | rot | Grundlauf |
|---|---|---|---|
| A18 | `Uhr.Now()` vor die Sperre gezogen | 20 von 20 (`TestSperreUmfasstUhr`, `TestGleichzeitigeAppend`) | 0 von 3 |
| A19 | Sperre nur um `Uhr.Now()` | 20 von 20 (`TestGleichzeitigeAppend`) | 0 von 3 |
| A20 | Freigabe im Fehlerpfad der Zeit fehlt (alle drei Zeit-Verstöße) | 20 von 20 (`TestSperreFreiNachFehler`) | 0 von 3 |
| M2 | Log unter der Sperre an allen drei Stellen (Zeit, Öffnen, Schreiben) | 20 von 20 (`TestLogNachFreigabeDerSperre`) | 0 von 3 |
| M3 | Fenster: Sperre lösen und sofort wieder nehmen **nach `Uhr.Now()`** | 20 von 20 | 0 von 3 |
| M3 | … nach der ID | 20 von 20 | 0 von 3 |
| M3 | … nach dem Öffnen | 20 von 20 | 0 von 3 |
| M3 | … nach der Prüfung des Zeilenendes (vor dem Schreiben) | 20 von 20 | 0 von 3 |
| M3 | … nach dem Schreiben, vor `Sync` | **0 von 20** | benanntes Negativ A21, bestätigt |
| M3 | … nach `Sync`, vor `Close` (Stelle, die der Plan nicht nennt) | **0 von 20** | folgenlos wie die vorige; vom Negativ A21 gedeckt, aber nicht genannt (V7) |
| M3 | … vor `Uhr.Now()` (Stelle, die der Plan nicht nennt) | 0 von 20 | äquivalent: das Verhalten ändert sich nicht (Prüfung und Uhr bleiben unter einer Sperre ohne beobachtbare Lücke) |

Zeitpunkte A04/A10, vom Review nicht mutiert, vom Verifier (N = 1, alle rot, aus dem benannten Grund): Jahr-Prüfung vor der Umrechnung nach UTC (`TestZeitAusserhalb`), `Uhr.Now()` zweimal gerufen (`TestSperreUmfasstUhr`, `TestSperreFreiNachFehler`, `TestLogNachFreigabeDerSperre`), untere Grenze `< 2` und obere `>= 9999` (`TestZeitpunkt`), Kürzung auf Mikrosekunden und Rundung auf Millisekunden (`TestZeitpunkt`).

## Befunde nach Schwere

### MEDIUM

**V1 — Kennungsform: der Fix zu M5 deckt die Klasse nicht (Byte × Länge, Länge allein)**
- `quelle`: `AGENTS.md` §3.17 (Fix gegen die Klasse, nicht die genannte Mutation), §3.15 · `klasse`: Menge ohne vollständige Achsen
- `pfad`: `src/adapters/driven/audit/audit_test.go` `TestKennungsform`, Produktcode `istKennung`
- Die Zusage (`SPEC-234` Nr. 3, A08): Länge 1 bis 64 (64 Längen), 65 gültige Bytes, 191 ungültige; jedes Byte an jeder Stelle jeder Länge. Der Test führt das Byte nur bei Länge 3 (256 Bytes × 3 Stellen) und bei den 11 Längen {1, 2, 3, 4, 5, 8, 31, 32, 33, 63, 64} nur zwei ungültige Bytes (`"` und `0xff`). Zwei Achsen bleiben ungekreuzt: Byte × Länge und die 53 Längen außerhalb der Liste.
- Probe-Mutationen, die weder Review noch Fix nannten, alle **grün** (kein Test rot): ein `/` ab Stelle 3 akzeptiert; ein LF ab Stelle 3 akzeptiert (zerbräche die JSON-Zeile des Audit-Logs); ein Leerzeichen ab Stelle 40 akzeptiert; Länge 40 ungeprüft (`return true`); Länge 17 ungeprüft ab dem ersten Byte. Länge 6 ungeprüft wurde rot, aber nur zufällig durch den sechs Bytes langen Mehrbyte-Wert des Tests (`日本`), nicht durch die Schleife der Fix-Runde.
- Szenario: eine Änderung der Prüfung (etwa ein Tabellenzugriff oder eine Schleife mit Sonderfall) lässt ein Steuerzeichen in einer langen Kennung durch; der Client wählt Session-, Workspace- und Import-Kennungen selbst (`SPEC-234` Nr. 3), die Zeile des Audit-Logs wäre kein gültiges JSON.
- Fix gegen die Klasse (Vorschlag): Schleife über Länge 1 bis 64 × 256 Bytewerte an mindestens zwei Stellen (erste, letzte) gegen das Orakel `inKlasse`, plus die Schleife über Stellen für ein ungültiges Byte der Länge 64. Aufwand: etwa 64 × 2 × 256 × 5 Aufrufe im Arbeitsspeicher-Verzeichnis; kein Produktcode.
- `verifizierbar`: ja.

**V2 — Zeilenende: die Achse „letztes Byte“ führt nur `x`, `b` und CR (Menge 255 Werte ohne LF)**
- `quelle`: `AGENTS.md` §3.17, §3.15 · `klasse`: Zustand ohne Achsenliste
- `pfad`: `src/adapters/driven/audit/audit_test.go` `TestZeilenende`, Produktcode `endetOhneZeilenende`
- Die Zusage (`SPEC-235` Nr. 4 f): „letztes Byte nicht `\n`“ — jedes andere Byte ist eine halbe Zeile. Der Test führt als letztes Byte `x` (Länge 1 bis 64), `b` (`a\nb`), `t` (`halb`: `b`), CR. Probe-Mutationen `b[0] != '\n' && b[0] != X` für `X` in `}`, `"`, `0`, Leerzeichen, NUL: **alle fünf grün**.
- Szenario: ein Absturz zwischen dem Schreiben von `}` und `\n` (die wahrscheinlichste halbe Zeile, denn der Eintrag endet auf `}\n`) lässt ein letztes Byte `}`; eine Prüfung, die solche Zeilen als vollständig ansieht, klebt den nächsten Eintrag an `}`.
- Fix gegen die Klasse: Schleife über alle 256 Bytewerte als letztes Byte (255 mit vorangestelltem LF, eines ohne), Orakel aus dem Wortlaut. Aufwand: eine Schleife.
- `verifizierbar`: ja.

**V3 — „Der Fehlertext nennt keinen Wert“ (`SPEC-234` Nr. 1) von `Append` ohne Fänger**
- `quelle`: `AGENTS.md` §3.15 (Zusage ohne Fänger), `LH-QA-04.a` · `klasse`: Zusage ohne Fänger
- `pfad`: `src/adapters/driven/audit/audit.go` `pruefe` (fünf `fmt.Errorf`, eine für die Zeit)
- Der Text von `ErrUngueltig` aus `Append` wird in keinem Test auf einen Wert geprüft; die Canary-Prüfungen gelten dem Fehlertext von `New` (`TestNewWeistAb`, A02) und dem Log (A16, A17). Probe-Mutation: die ungültige Kennung im Text des Fehlers (`… Form einer Kennung %q`, `f.kennung`): **grün**.
- Szenario: ein Aufrufer, der den Fehler protokolliert (der Vertrag erlaubt es, `SPEC-234` Nr. 5 b sagt nur, dass der Dienst ihn ignoriert), schreibt die vom Client gewählte Kennung ins Diagnose-Log; der Wert ist eine Kennung, kein Token, aber der Vertrag sagt „keinen Wert“.
- Fix: je Art des Verstoßes (Art, Feld, Kennung, `KorrelationsID`, Code, Zeit) ein Eintrag mit Canary und `strings.Contains(err.Error(), canary)`, als Schleife über die Verstoßarten; oder die Zusage im Vertrag auf „Feldnamen, nie den Wert“ präzisieren und den Test auf die Namen einschränken.
- `verifizierbar`: ja.

### LOW

**V4 — Handle-Fänger (M1) kennt den Dateizustand „halbe Zeile“ nicht**
- `fenster_test.go` `TestKeinHandleBleibt` zählt Handles nur für Dateien mit LF am Ende und für `/dev/full`. Probe: `Close` entfällt, wenn die Datei auf eine halbe Zeile endet (`leck = halb` im `defer`): **grün**. Folge: ein Leck, das nur nach einem Absturz auftritt, bleibt unbemerkt. Fix: dieselbe Zählung mit den drei Zuständen leer, mit LF, halb.

**V5 — Rückkehrpfade der Sperre (A20): Achse „Verstoß der Zeit“ hat nur Jahr 0, Achse „Öffnen“ nur ein fehlendes Elternverzeichnis**
- `TestSperreFreiNachFehler` ruft ein zweites `Append` nach dem Jahr 0 und nach dem fehlenden Elternverzeichnis. Probe-Mutationen, die die Freigabe nur bei Jahr 10000 oder nur bei `EISDIR` auslassen: **grün**; nur bei der Nullzeit: rot, aber durch `TestLogNachFreigabeDerSperre` (dessen Meldung „der Log-Handler lief unter der Sperre“ die Ursache falsch benennt), nicht durch A20; nur bei `ENOSPC`, nur bei `ENOENT`: rot. Die Mutationen sind künstlich (der Code hat `defer`, zwei `return` unter der Sperre); sie zeigen, dass die Zeile „4 Pfade“ der Pflicht-Menge die Verstoßarten je Pfad (3 der Zeit, ≥ 2 des Öffnens) nicht führt. Fix: zweites `Append` nach jeder der fünf Zeit-Formen und beiden Öffnen-Fehlern (Verzeichnis, Elternverzeichnis).

**V6 — `error_code` in einer Art, die ihn verbietet: negative Werte nicht geführt**
- A07 setzt für das verbotene Feld einen positiven Wert (`wertVon`), A09 schleift −5 bis 60 nur in der Art `import_fehlgeschlagen`. Probe: „gesetzt“ als `ErrorCode > 0` statt `!= 0`: **grün** (ein negativer Code in `session_geloescht` würde angenommen und als Text geschrieben). Gegenprobe `> 0` bei der Prüfung der Gültigkeit: rot (`TestErrorCode`). Fix: die Schleife über die 66 Werte auch für die drei anderen Arten (verbotenes Feld).

**V7 — A21 nennt nicht alle akzeptierten Negative des Schreibens; der Grund des Negativs *ohne Leserecht* ist zu stark**
- Probe: `Write` in zwei Aufrufen (Hälften der Zeile) und das vorangestellte LF als eigener `Write`: beide **grün**. `SPEC-235` Nr. 4 g sagt „in einem einzigen Aufruf“ und nennt den Fall unter *Ungeprüft* (zwei Adapter/Prozesse), der Plan führt ihn in der Tabelle A21 nicht; das Negativ braucht Grund und Folge dort (Grund: innerhalb eines Adapters ist die Teilung nicht beobachtbar, die Sperre ordnet; Folge: zwei Adapter oder Prozesse auf derselben Datei können Zeilen mischen, `SPEC-235` *Ungeprüft*). Ebenso fehlt die Stelle „zwischen `Sync` und `Close`“ (Rate 0 von 20) neben „zwischen Schreiben und `Sync`“.
- Der Grund des vom Architect entschiedenen Negativs *Datei ohne Leserecht* („im Image als Superuser nicht auslösbar“, `SPEC-235` *Ungeprüft*, Plan §6) ist überstark: der Superuser des Images darf in den Prüflauf einen Prozess mit niedrigerer Kennung abspalten (`su nobody` in dem Image, `--user 0`, netzlos, beantwortet `Permission denied` auf eine Datei `0200`); ein Test mit `os/exec` und `SysProcAttr.Credential` und einem Verzeichnis, das jenem Benutzer gehört, könnte den Fall auslösen. Die Folge (Einträge fehlen, jeder Fehlschlag steht im Log) ist richtig; der Grund sollte „nicht ohne einen Prozess mit niedrigerer Kennung auslösbar, der Aufwand lohnt nicht“ heißen oder der Fänger entstehen. Adresse: Architect.

**V8 — `Arten()` als Kopie: nur Element 0 geprüft**
- `TestArtenLiefertKopie` verändert `a[0]`. Probe: die Pflichtfelder der letzten Art (`import_fehlgeschlagen`) als gemeinsame Package-Variable: **grün**. Fix: Schleife über alle vier Elemente und alle Pflichtfelder (Änderung, zweiter Aufruf, Vergleich).

**V9 — Tabelle A08 des Plans nennt „Länge 0“ für fünf Felder; der Vertrag erlaubt sie bei `KorrelationsID`** (Befund steht unter *Befunde beim Schreiben der Fänger*, die Zeile blieb). Fix: Zeile berichtigen (vier Felder).

### INFO

- **I1 — Zeit und ID der Use Cases**: `IDs.KorrelationsID()` wird unverändert in die Zeile geschrieben, ohne Prüfung der Kennungsform; der Vertrag (`SPEC-235` Nr. 3) setzt voraus, dass jeder Wert eine Form hat, die keine Maskierung braucht. Der Generator ist ein Port des Composition Root (`SPEC-114`); eine Prüfung wäre ein Schutz gegen sich selbst. Kein Befund, nur benannt.
- **I2 — `New` erkennt nur den untypisierten nil-Wert** (`c.Uhr == nil`); ein typisierter nil-Zeiger in `Uhr` besteht `New` und bricht erst im ersten `Append` (Panik). Der Vertrag sagt „nil-Wert“; im Rahmen des Negativs *Nullwert* (Folge dort: Panik, kein Test) liegt es in derselben Klasse. Kein Befund.
- **I3 — Akzeptiertes Negativ *Nullwert `Datei{}`*** (Architect `d58715c`): Grund (der Composition Root ist der einzige Erzeuger) und Folge (Panik, die Sperre wird durch `defer` freigegeben, kein Test sieht sie) stehen in `SPEC-235` Nr. 1/*Ungeprüft* und im Plan §6; tragfähig. Aus dem Code gelesen, nicht ausgeführt: `Datei{}.Append` panikt in `anhaengen` bei `d.cfg.Uhr.Now()`; die Sperre wird durch `defer` freigegeben.
- **I4 — A21 (Prozessende zwischen Commit und Append)**: gehört `slice-audit-dienste` (E19, `SPEC-234` *Ungeprüft*); in diesem Teil gibt es keinen Dienst, der dazwischen läuft. Verträglich: `Sync` vor `nil` ist im Code (`return f.Sync()`) und in `SPEC-235` Nr. 4 h; der Wegfall des Aufrufs ist als Negativ benannt (S3 des Reviews), Fänger fehlt (kein Seam).
- **I5 — Sendung/Nachträge an Teil E und V**: `slice-audit-dienste` auf `main` (`open/`, Zeilen 44 und 145) nennt den Typ noch `Art`, nicht `AuditArt`; der Plan weist den Nachzug dem Planner „bei der Integration“ zu (§6 *Namenskollision*, Ausgang *entschieden*, `AGENTS.md` §3.13: die Adresse ist genannt, die Zeile des Nehmers ist noch nicht geändert). `slice-audit-verdrahtung` nennt den Typ nicht. Zeile E05 und weitere sagen „Art und Feldern“ im Sinne des Feldes `Eintrag.Art`, das seinen Namen behielt. Aktion Planner vor der Integration: Zeilen 44 und 145 auf `AuditArt` stellen.

### Kein Befund (geprüft)

- **Geheimnisse (`LH-FA-14`, `LH-QA-04.a`)**: Kein Tokenwert, kein Pfad, kein Fehlertext in der Zeile oder im Log (A16/A17, vom Review und von meinen Mutationen `Pfad als Attribut`/`Text des Fehlers` gefangen; kein Feld des Eintrags kann ein Token tragen).
- **Rohbytes (§3.18)**: `LC_ALL=C grep -P` auf Plan, Review, `audit.go`, `support/audit.go` und alle Tests: kein doppelt kodiertes UTF-8 (`C3 83 C2 xx`, `C3 82 C2 xx`), kein BOM, kein NBSP, keine Nullbreiten-Zeichen, keine Folge Backslash-`u`-Hex; die Mehrbyte-Zeichen der Tests entstehen aus Codepunkten (`utf8.AppendRune`) oder Bytewerten (`0xc3, 0xa9`).
- **Kommentare (§3.7/§3.11)**: die Kommentare der Produktdateien beschreiben den Ist-Zustand (Zusage/Kopplung/Rang-Zeiger); der Test `TestKennungsform` sagt „der Liste“, nicht „jeder Länge“; `TestKeinHandleBleibt` und `TestFensterZwischenUhrUndSchreiben` sagen zu, was sie prüfen (ihre Grenzen sind V4 und das Negativ A21). `AuditArt` trägt den Grund des Namens.
- **Offene Handles**: `Close` im `defer` jedes Pfades nach dem Öffnen; Mutation „Close nur im Erfolgspfad“ und „entfällt“ vom Fänger M1 gefangen (Implementer, nicht neu gemessen); meine eigene Mutation (halbe Zeile) ist V4.
- **Codes (A09)**: Mutation `> 0` bei der Gültigkeit rot (`TestErrorCode`); die Schleife über 66 Werte mit 36 gültigen und der Zählprüfung ist vollständig für die Art `import_fehlgeschlagen`.
- **Schlüsselmenge des Log-Eintrags (A16)**: Mutationen „zusätzlicher Schlüssel `art`“, „Code E7001 statt E9001“, „leere Schlüssel mitschreiben“, „Eintrag bei Verstoß doppelt“ rot (`TestLogBeiFehler`, `TestLogBeiUngueltig`).
- **Rechte (A11)**: `0640` und `0604` statt `0600` rot (`TestRechte`).

## Gates

`make gates` im Worktree (HEAD `d58715c` samt diesem Bericht, unter `flock`): Exit 0, alle Gates grün. Darunter `make groessenmarke` („slice-audit-log.md: ok, Messung 1698 von Marke 3000“), `make test-race` (grün, das SQLite-Flake trat nicht auf), `make a-check` (0 Befunde; der Hinweis auf Dateien unter `tools/` ohne Schicht ist vorbestehend), `make migration-check` und `make schema-check` (unverändert), `make zeichen-check` (1363 Dateien), `make docs-check` (0 Befunde nach der Verlinkung der ADR-Kennungen dieses Berichts). Die Gegenproben (`make gegenproben`) sind kein Gate und nicht gefahren; kein Gate der Kette änderte sich.

## Was nicht aufgelöst wurde

- **Unbeschränkte Mengen mit Prüfgrenze:** Byte × Länge × Stelle der Kennungsform ist beschränkt (64 × 256 × 64 × 5 Felder); der Fänger kann alle Längen und Bytes führen (V1), die Stellen nur als Auswahl (erste, letzte, mittlere). Eine Probe-Mutation, die genau eine Stelle bei genau einer Länge überspringt (zum Beispiel Stelle 20 bei Länge 41), bleibt auch bei einem Fänger „Länge × zwei Stellen“ offen; sie ist unrealistisch und wurde nicht gemessen.
- **Zwei Adapter/Prozesse auf derselben Datei** und **ein Prozessende zwischen Schreiben und Sync**: nicht auslösbar (Negative im Vertrag).
- **Das SQLite-Flake in `make test-race`** wurde nicht gesucht; falls es im Lauf von `make gates` auftritt, steht es unten vermerkt.

## Empfehlung

**Closure: ja**, nach dem Nachzug von V1 bis V3 (Fänger oder benanntes Negativ mit Grund und Folge im Plan §3) und V9; V4 bis V8 und I5 gehören in dieselbe Runde oder als LOW ins Register. Kein Produktcode ändert sich dafür (alle Probe-Mutationen sind Test-Lücken, die Produktcode-Fassung stimmt mit dem Vertrag überein).
