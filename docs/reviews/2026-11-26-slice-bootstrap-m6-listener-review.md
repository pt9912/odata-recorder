# Review-Report: slice-bootstrap-m6-listener (Admin-Listener im Composition Root) — 2026-11-26

**Review-Art:** Code, gegen Plan, Entscheidungen und Hard Rules (nicht gegen die DoD).

**Gegenstand:** slice-bootstrap-m6-listener, Zweig-Stand `fa015fde`; eigener Anteil `git diff 14b75b45..HEAD` über die Pfade des Blocks, dazu Handbuch, Katalog-Ausnahmen und `SPEC-201`. Die Änderungen an der Spezifikation aus `main` sind nicht Gegenstand.

**Skill:** `.harness/skills/reviewer.md` @ `fa015fde` ·
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-26

**Eingangs-Kontext:**

- Plan slice-bootstrap-m6-listener (§1, §3 Zusage/Fänger/Messung/Negative, §4, §6)
- `SPEC-300`, `SPEC-299`, `SPEC-139`, `SPEC-110`, `SPEC-238`, `SPEC-239`, `SPEC-201`
- `LH-FA-13`, `LH-RB-02`, `LH-QA-04`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `AGENTS.md` (Hard Rules)

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Prüfung der Dateien des Dienstes hat allein Fälle mit absoluten Pfaden: ohne den Rückfall auf `.` bei einem Namen ohne Verzeichnis (M1) und mit lexikalisch gekürztem Verzeichnis vor `os.SameFile` (M11) bleiben die Tests grün; das Binary mit M1 startet mit `ADMIN_TOKEN_FILE=audit.jsonl` im `DATA_DIR` und schreibt das Token in die erste Zeile von `audit.jsonl`. Der Kommentar sagt die Wirkung von `..` hinter einem Verweis zu, die kein Test prüft; der Code im Stand weist den relativen Namen ab. | `AGENTS.md` §3.15, §3.11; `SPEC-300` Nr. 4, 12 a; `LH-QA-04` | `src/bootstrap/konfig_admin.go` · `if verzeichnis == "" {` und „hinter einem Verweis wie beim Öffnen wirkt“ | ja — Mutation M1 bzw. M11 bleibt in den gefahrenen Tests grün | Fänger führt nur eine Form der Eingabe |
| F-2 | MEDIUM | Das Paar „Bind von `LISTEN` scheitert“ und „`ctx` beendet beim zweiten Bind“ hat keinen Fall: gewinnt in der Mutation der gesammelte Bind-Fehler vor dem Kontextfehler (M12), bleibt alles grün; ein Signal in diesem Augenblick endete dann mit Exit 1 und einer Zeile zum Bind statt als Abbruch. Der Code im Stand lässt den Kontextfehler gewinnen. | `AGENTS.md` §3.16; `SPEC-300` Nr. 7; Skill „Paar von Ursachen ohne Fänger“ | `src/bootstrap/lebenszyklus.go` · `if kontextFehler(err) {` in `binde` | ja — Mutation M12 bleibt in den gefahrenen Tests grün | Paar von Ursachen ohne Fänger |
| F-3 | MEDIUM | Der Kommentar vor `adapterArten` sagt weiter, das Binary rufe `admin` nicht auf; mit diesem Slice ist das falsch. Berichtigt ist allein der Satz im Kopf von `TestAuditAufrufstellen`, und §6 des Plans führt die Berichtigung als erledigt. | `AGENTS.md` §3.9, §3.7; Skill „Wiederholung eines Musters, das schon zweimal LOW war“ | `src/bootstrap/audit_betrieb_test.go` · „kein Use Case schreibt sie, und das“ (Zeilen 843 und 844) | nein — Review | Berichtigung erreicht nicht jede Stelle |
| F-4 | LOW | Das Handbuch sagt, bei gleichem Port von `LISTEN` scheitere der Start; gemessen starten `LISTEN=127.0.0.2:18081` und `ADMIN_LISTEN=127.0.0.1:18081` beide (`start: bereit`, `/healthz` antwortet, Exit 0 nach `SIGTERM`). Der Satz zu jeder anderen Variable mit `ADMIN_` nennt nicht, dass ein leerer Wert kein Verstoß ist. Beides ist kein Gate-Gegenstand (Negativ in B17). | `AGENTS.md` §3.11; `SPEC-300` Nr. 7; `SPEC-110` Nr. 2 | `docs/user/benutzerhandbuch.md` · „Nutzt `LISTEN` denselben Port“ | nein — Lauf am Binary, kein Gate | Handbuch-Satz widerspricht dem Binary |
| F-5 | LOW | §1 weist „die vier Aufrufe von `SPEC-299` Nr. 8“ Teil C zu; dieser Slice ruft `pruefeAdminTransport` am Laden und in `Build` (B19) und vereinigt in `Build` mit `LISTEN_TLS_*`, §1 ist nicht nachgezogen, und §6 führt die Vereinigung als „entschieden (Implementer)“ statt offen mit Adresse. In der Sache entspricht der Code `SPEC-139` Nr. 2 und `SPEC-299` Nr. 8 b, die im Rang über dem Plan stehen. | `AGENTS.md` §3.9, §3.13 | Plan slice-bootstrap-m6-listener §1 · „die vier Aufrufe von `SPEC-299` Nr. 8“; §6 · „Vereinigung in `Build`“ | nein — Review | Plan folgt der Änderung nicht |
| F-6 | LOW | Verträge der Harness beschreiben den Root ohne Admin-Listener, was mit diesem Slice nicht mehr gilt („bis dahin gibt es im Binary keinen Admin-Listener“, „das Binary hat keinen Eingang dafür“, „bis zum Admin-Listener gibt es keinen Lesepfad für Snapshots“); weder berichtigt noch in §6 mit Adresse geführt. | `AGENTS.md` §3.9, §3.13 | `harness/vertraege/SPEC-041-test-security.md` Nr. 10 a, b; `harness/vertraege/SPEC-037-test-integration.md` Nr. 10 a, j | nein — Review | Berichtigung erreicht nicht jede Stelle |
| I-1 | INFO | An den Verifier: der Absatz „Messung der Mutationen“ führt nicht jede Mutation der Fänger-Tabelle (zum Beispiel B01 „Prüfer ohne Bereich“, B16 Token als `art`, im Body, in der Telemetriezeile, B15 fünftes Feld und Zähler), und B16 nennt je 28 statt 24 Zellen. | `AGENTS.md` §3.10, §3.15 | Plan §3 · „Messung der Mutationen“ | nein — Verifikation | Messung deckt die Tabelle nicht |
| I-2 | INFO | Der Name `ADMIN_LISTEN` steht als zwei Konstanten (`adminListen`, `schluesselAdminListen`); die Gruppe `nichtGeliefert` eines Tests trägt den Namen einer entfernten Funktion. | Maintainability | `src/bootstrap/konfig_admin.go` · `schluesselAdminListen = "ADMIN_LISTEN"`; `src/bootstrap/m4_konfig_klassen_test.go` · `{"nichtGeliefert", "ADMIN_X"` | nein | Rest einer Umstellung |

## Eigene Mutationen und Läufe

Wegwerfkopie des Stands `fa015fde`, je ein Container ohne Netz (`--memory=4g`, 4 CPU) unter der gemeinsamen Sperre, Last des Hosts 2 bis 4; je Mutation ein Lauf `go test -count=1` gezielter Tests von `src/bootstrap`, ohne Detektor. Nullprobe der Muster `TestAdmin|TestKonfig|TestStart|TestWarte|TestServe|TestHerunter|TestKontext|TestFrist` grün.

| Mutation | Ergebnis |
|---|---|
| M1 Rückfall auf `.` entfernt; M11 `filepath.Clean` vor `os.Stat`; M12 Bind-Fehler vor Kontextfehler | grün (F-1, F-2) |
| Leerraum in `ADMIN_X` als Verstoß; Rest der Frist gleich Rest des OData-Servers; jeder Ausfall bei beendetem `ctx` als Abbruch; `nil` nach `Herunterfahren` als Ausfall | rot |
| `Build` ohne Prüfung von `ADMIN_LISTEN`; Zeilen von `LISTEN_TLS_*` verloren; Prüfung mit Flag `true` | rot |
| Admin-Server nur in `record` geschlossen; nur außerhalb von `record` angehalten; Token ins Log; Token-Pfad mit Leerraum | rot |
| Je Zeile B01 bis B08, B12, B14 und B15 eine weitere Mutation (Form nicht geprüft, Pfad unter `DATA_DIR`, Präfix als Teilstring, Token nicht in `mock`, Anfang statt Name, Routen gekürzt, Adresse ohne Listener, Schlüssel der Zeile, Kanal ohne `Once`, Admin-Bind nur in `record`, Telemetrie `nil`) | rot |
| Katalog: Zeile `ODR-E7005` aus dem Handbuch | rot (`katalog: ok` im Stand) |

B09 und B11 sind durch die Mutation am Anhalten des Admin-Servers aufgelöst, B10 durch die am Rest der Frist, B13 durch die zwei am Ausfall, B19 durch die drei an `Build`. B16 ist allein mit dem Log-Weg gefahren, B17 allein mit dem Katalog, B18 hat keinen Fänger.

Weitere Läufe: die Muster der Nullprobe samt `TestWerte` und `TestAbbruch` unter `-race` einmal grün. `TestE2EProzess`, Fall `ADMIN_LISTEN`, in einem Container mit `--read-only`: Stand 3 von 3 grün, Binary ohne die Loopback-Prüfung 3 von 3 rot (über `DATA_DIR`). Am gebauten Binary: Token-Datei `0600`; nach Abweisung, Diagnose und Session-Liste steht das Token in keiner geprüften Datei unter `DATA_DIR` außer `admin.token`, nicht in `stdout`, `stderr` und den zwei Antworten; ohne Token `401`; der relative Pfad `audit.jsonl` im `DATA_DIR` wird abgewiesen.

Nicht gefahren: `make gates`, `make gegenproben`, `make handbuch-check`, die ganze Suite von `src/bootstrap` (länger als die Frist von 240 s des Probelaufs), Raten mit Wiederholung.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Token in Log, Fehlertext, Antwort, Audit (`konfig_admin.go`, `lebenszyklus.go`, `start.go`, `wurzel.go`) | geprüft, ohne Befund |
| Offene Route ohne Token; Loopback-Pflicht ohne TLS am Laden und in `Build`, in jedem Modus | geprüft, ohne Befund |
| Ablehnung der Dateien des Dienstes im Verhalten (absolut, relativ, Verweis auf das Verzeichnis) | geprüft, ohne Befund im Code; Fänger siehe F-1 |
| Bindung, Schließen bei Fehlern, Fehlerkanal samt `nil`-Rückkehr, Herunterfahren OData vor Admin, Rest der Frist, Rennen und Hänger | geprüft, ohne Befund im Code; Fänger siehe F-2 |
| Randformen im Code ohne Ort in der Spezifikation (`AGENTS.md` §3.16) | geprüft, ohne Befund |
| Vereinigung der Verstöße von `ADMIN_LISTEN` und `LISTEN_TLS_*` gegen `SPEC-139` Nr. 2 | geprüft, ohne Befund; Plan siehe F-5 |
| Geänderte Bestandstests (`src/bootstrap/`, `tests/`), Lockerung nach `AGENTS.md` §3.6 | geprüft, ohne Befund |
| `harness/katalog-ausnahmen.txt`, `SPEC-201` Nr. 3 e | geprüft, ohne Befund |
| `docs/user/benutzerhandbuch.md`: Schlüssel, Gründe, Routen, Codes | geprüft; Befund F-4 |
| `docs/plan/planning/` (Roadmap, Welle, Teil C) | geprüft, ohne Befund außer F-5 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 3 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fänger führt nur eine Form der Eingabe · Paar von Ursachen ohne Fänger · Berichtigung erreicht nicht jede Stelle · Handbuch-Satz widerspricht dem Binary · Plan folgt der Änderung nicht · Messung deckt die Tabelle nicht · Rest einer Umstellung

## Verdikt

**Merge-blockierend:** ja — F-1, F-2 und F-3 (MEDIUM). Im Verhalten des Binary ist kein Fehler gefunden.

**Übergabe:** Findings an den Implementer, F-6 an den Architect, I-1 an den Verifier; die Finding-Klassen gehen in die Closure des Slice (§7). Der Report ersetzt keine Verifikation.
