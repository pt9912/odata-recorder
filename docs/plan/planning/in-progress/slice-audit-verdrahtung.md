# Slice slice-audit-verdrahtung: Audit-Log im Composition Root, Datei im Datenverzeichnis und Canary

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-RB-04`, `LH-FA-13`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0006](../../adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)

**Berührte Spec-Stellen:** `LH-RB-04.a` · `LH-QA-04.a` (kein Geheimnis in den Einträgen) · `SPEC-139` Nr. 2 bis 4, 6, 7 und 10 (Build, eine Instanz je Prozess, Modi, Hochfahren, Herunterfahren, Neustart) · `SPEC-077` (`DATA_DIR`, Datei `audit.jsonl`) · `SPEC-234` Nr. 3, 4, 5, 8 und 10 (Wertformen, Stellen, Reihenfolge, Konfig der Dienste, Wiederholung) · `SPEC-235` Nr. 1, 2, 5, 6, 7, 8 (Pfad, Datei entsteht mit dem ersten Eintrag, Fehler, Geheimnisse, Neustart, Name der Datei) · `ARC-020`, `ARC-021`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-audit-verdrahtung)

**Autor:** pt9912. **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel (Teil V des Schnitts von `slice-audit-log` in drei Teile, verkleinert; Entscheidung des Planner bei der Closure von `slice-audit-dienste`, 2026-11-06):** Der Composition Root erzeugt je Prozess genau einen `audit.New` über `<DATA_DIR>/audit.jsonl` mit der einen Uhr, dem einen Generator und dem einen Logger und gibt diese eine Instanz als `Audit` in `maintenance.Konfig`, `session.Konfig`, `workspace.Konfig` und `baselineimport.Konfig` (`SPEC-139` Nr. 3, `SPEC-234` Nr. 8) — **das ist geliefert**: `slice-audit-dienste` trägt es als Brücke in `src/bootstrap/wurzel.go` (zwei Commits »Brücke:«, `5b7636e`, `d210747`), weil `New` ein nil-`Audit` abweist und ohne sie kein grüner Stand zwischen den beiden Slices besteht. Dieser Slice liefert die **Fänger der Datei im Betrieb** über das `Handle`: nach einer Löschung steht der Eintrag in der Datei, im Modus `record` mit allen vier Arten, im Modus `replay` für Aufbewahrung und Workspace (V01); ein zweites `Build` auf demselben Datenverzeichnis hängt an (V02); ein Canary (Userinfo und Query-Wert der `UpstreamURL`, Fehlertext) steht in keinem Byte der Datei (V03); ohne Löschung entsteht die Datei auch nicht beim Hochfahren und Herunterfahren (V05, Rest; dass `Build` allein sie nicht anlegt, fängt `TestAuditPfadUndZeile`) (`SPEC-235` Nr. 2); ein Schreibfehler der Ablage ändert das Ergebnis des Dienstes nicht (V06); die Zahl der Aufrufstellen von `Append` ist die Zahl der Zeilen von V01 (V07). Der Code der Brücke ändert sich hier nur, wenn ein Fänger eine Lücke zeigt; dann geht der Fund als Befund an den Reviewer, nicht still in den Code.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Port, Typen und Adapter — `slice-audit-log`; die Aufrufstellen und das Feld `Audit` der vier Konfigs — `slice-audit-dienste`. Grund: dieser Slice verdrahtet fertige Teile und schreibt keinen Dienst um.
- Die Instanz, der Pfad und die Übergabe an die vier Konfigs in `src/bootstrap/wurzel.go` samt Zeile V04 (dieselbe Instanz) und dem Logger der Instanz — **geliefert von `slice-audit-dienste`** (Brücke, Geber: `slice-audit-dienste`), gefangen von `TestAuditEineInstanz`, `TestAuditPfadUndZeile` und `TestAuditLoggerDesProzesses` in `src/bootstrap/`; hier nicht wiederholt. Akzeptiertes Negativ der Brücke (Grund und Folge im Plan von `slice-audit-dienste`, §3): eine zweite Uhr oder ein zweiter Generator für den Adapter bleibt ungefangen, weil beide ohne Zustand sind.
- Audit-Einträge der Admin-Schreibendpunkte und der Eintrag zu `ODR-E2101` — `welle-admin-api` (`SPEC-005`, `LH-FA-13.a`); die Nehmer tragen die Entscheidung `SPEC-234` Nr. 7.
- Das Handbuch-Kapitel zur Datei `audit.jsonl` — mit `slice-admin-sessions` (erster Slice, dessen Binary eine Löschung auslöst); hier steht in §7, was das Binary nach der Verdrahtung tut und warum kein gebundenes Beispiel die Datei erzeugen kann (siehe §6).
- Die Abbildung auf die Telemetrie (`audit_fehlgeschlagen`, Adapter `audit`) — `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock` (`SPEC-231` Nr. 4 und 7).
- Rotation, Lesen und Export der Datei — kein Abnahmeumfang (`SPEC-235`, Entscheidung).
- Ein Fänger gegen einen Tokenwert (`ADMIN_TOKEN_FILE`) in der Datei — der Vertrag, der den Admin-Listener verdrahtet. Grund: `Build` erhält kein Feld dafür und liest die Datei nicht; Grund und Folge stehen in `SPEC-139` *Ungeprüft*.

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

- [ ] Datei im Betrieb über das `Handle`: nach einer Löschung und nach einem fehlgeschlagenen Import steht der Eintrag in `<DATA_DIR>/audit.jsonl`, in beiden Modi, wo der Dienst besteht (`SPEC-139` Nr. 4, `SPEC-235` Nr. 2; Zeilen V01a bis V01d); ein Schreibfehler der Ablage ändert das Ergebnis des Dienstes nicht (V06). Die Verdrahtung selbst (eine Instanz in allen Diensten, Zeile V04) liefert `slice-audit-dienste` als Brücke und ist hier nicht Gegenstand.
- [ ] Neustart und leere Ablage: ein zweites `Build` hängt an und ändert Vorhandenes nicht, ohne Löschung entsteht die Datei nicht, auch nicht beim Hochfahren und Herunterfahren und nicht bei einer abgelehnten Löschung (`SPEC-235` Nr. 2, 7; Zeilen V02a, V02b, V05a bis V05c).
- [ ] Kein Geheimnis: ein Canary-Secret (Userinfo und Query-Wert der `UpstreamURL`, Fehlertext) erscheint in keinem Byte der Datei (`LH-QA-04`; Zeile V03); der Tokenwert der Admin-Authentisierung erreicht `Build` nicht (§1, `SPEC-139` *Ungeprüft*).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: die Datei `audit.jsonl` ist für den Anwender sichtbar, aber kein gebundenes Beispiel gegen das gebaute Binary kann sie erzeugen (das Binary kennt `version` und `start`; ein Admin-Endpunkt fehlt); das Kapitel folgt mit `slice-admin-sessions`. In §7 steht, was das Binary nach der Verdrahtung tut.
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
| `src/bootstrap/audit_betrieb_test.go`, `src/bootstrap/audit_betrieb_hilfen_test.go` | neu (nur Tests) | Fänger der Datei im Betrieb über das `Handle` (`BuildMit` mit `StandardBauteile`, echte Dienste, echte SQLite, echter Adapter); der Code der Verdrahtung steht als Brücke in `wurzel.go` (Lieferung von `slice-audit-dienste`), Produktcode 0 |
| `src/bootstrap/export_test.go` | nur wenn ein Zugang fehlt | Bestand; die Haken von `Bauteile` genügen nach heutigem Stand |
| `src/hexagon/`, `src/adapters/driven/audit/` | entfällt | Port, Typen, Dienste und Adapter liefern `slice-audit-log` und `slice-audit-dienste` |
| `tools/schema/schema.yaml`, `src/adapters/driven/sqlite/migrations/` | entfällt | die Ablage ist eine Datei (`SPEC-235`): kein Schema, keine Migration |
| `spec/spezifikation.md` | update (Architect, getan) | `SPEC-139` Nr. 2, 6, 7, 10 und *Ungeprüft* (Randformen), `SPEC-077` unverändert |
| Tests | neu | Tests des Roots gegen das `Handle`; die Datei gelesen und geparst |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation vor dem ersten Code-Commit (`AGENTS.md` §3.12).

**Entscheidungen des Architect** (2026-11-07, zu den Aufträgen des Planner; jede mit Ausgang in §6):

1. **Produktcode 0 bestätigt.** `wurzel.go` erzeugt `audit.New` in `verdrahte` einmal, nach dem Blob-Verzeichnis und nach `b.Oeffne`, mit `filepath.Join(k.DataDir, audit.DateiName)`, `h.uhr`, `h.ids`, `h.log`; dieselbe Instanz geht in `maintenance` (Zeile 251), `workspace` (258), `session` (292) und `baselineimport` (300); im Modus `replay` bestehen nur die ersten beiden. `Build` berührt die Datei nicht, `Herunterfahren` ruft nichts am Audit. V01, V02, V03, V05 und V06 brauchen daran nichts; ein Fänger, der anderes zeigt, ist ein Befund an den Reviewer (§1).
2. **Der Auslöser der Aufbewahrung ist `Hochfahren`, nicht `Build`.** `Wartung.Startup` läuft in `Hochfahren` Schritt (1) (`SPEC-139` Nr. 6), und die Stufe 3 findet nur Sessions von Services, die die Ablage schon kennt (`legeDiensteAn` folgt ihr): die Ausgangslage kommt aus einem früheren Lauf oder einer gesäten Ablage. Der Satz »Aufbewahrung beim zweiten `Build`« der ersten Fassung der Tabelle war falsch; die Zeilen V01b und V02b tragen die berichtigte Form.
3. **Ausgangslage durch Säen.** Die Fälle säen die Ablage vor `Build` über `sqlite.Open` auf `<DATA_DIR>/odata-recorder.db` und den Blob-Speicher auf `<DATA_DIR>/blobs` (die Fixtures von `workspace_sqlite_test.go` zeigen den Weg: Service, abgeschlossene Session, veröffentlichte Baseline, Workspace aus `Create`); danach schließt der Test die Ablage und baut. Das Orakel kennt die gesäten Kennungen aus dem Säen, nie aus dem Dienst.
4. **Der Fake von `UpstreamURL` und Token hält jedes Argument fest.** Zwei Test-Doubles, beide in `audit_betrieb_hilfen_test.go`: `AuditAufzeichnung` (Decorator über der Instanz des Roots, über die Haken `NeueWartung`, `NeueSession`, `NeueWorkspaces`, `NeuerImport` in alle vier Konfigs gesetzt; hält jeden `Eintrag` unverändert und ungeprüft fest, auch einen, den der Adapter abwiese, und zählt jeden Aufruf; reicht weiter an die Instanz des Roots, sodass die Datei die Wirkung bleibt) und `SessionsMitFehlertext` (Decorator über `outbound.SessionRepository`, über `NeuerImport`; `Get` liefert einen Fehler, dessen Text den Canary trägt; jedes Argument festgehalten). Beide zählen als Fakes (§4, `fakes: 2`).
5. **Der Canary entsteht aus Bytewerten** (`AGENTS.md` §3.18). Die Werte: Benutzer, Passwort und Query-Wert in der `UpstreamURL` jedes Dienstes der Konfiguration, je in zwei Klassen. *Formgültig* (Bytes aus `A-Z a-z 0-9 . _ -`, bis 64 Bytes, `SPEC-234` Nr. 3): sie kämen an der Formprüfung des Adapters vorbei und erreichten die Datei, wenn der Root sie in ein Feld legte. *Formfremd* (Sonderzeichen `@ : / % ?` und ein Byte ab `0x80`, aus `[]byte{…}` gebaut, in der URL prozent-kodiert): sie sind der Fall, den ein Betreiber hat. Der Fehlertext-Canary ist ein Text, den `SessionsMitFehlertext` im Fehler trägt. Der Test liest die geschriebene Testdatei mit `LC_ALL=C grep -P` auf Rohbytes gegen, der Bericht nennt die Datei. Vorab prüft der Implementer, dass `Build` die URLs mit Zugangsdaten und Query annimmt (`SPEC-139` Nr. 4: jeder Wert unverändert; `baselineimport.New` prüft nur Schema und Host); nimmt es sie nicht an, ist das eine Rückgabe an den Architect, keine Änderung der URL im Test.
6. **Schnittfrage entschieden: kein Schnitt.** Die Tabelle hat 12 Zeilen (alle `andere`), `fakes` 2, Produktcode 0: *S* = 80·12 + 200·2 = 1360, *M* = 1900, weit unter der Schnittgrenze (*S* 2223). Der Planner erwartete 4 Zeilen; die Zeilen wachsen, weil der Auftrag 5 je Aufrufstelle eine Zeile verlangt und Auftrag 2 den Fake nennt. Jede weitere Zeile oder jeder weitere Fake ist wieder eine Schnittfrage (`SPEC-128` Nr. 10 b).
7. **Kennungen:** keine neue `SPEC`-Nummer; geschrieben ist nur in `SPEC-139` (Nr. 2, 6, 7, 10, *Ungeprüft*, Historie). `SPEC-077` und `SPEC-235` bleiben unverändert (ihr Wortlaut trägt die Zellen). Nichts an `SPEC-110`, nichts an Linie B.
8. **Katalog:** kein neuer Fehlercode, Schlüssel oder Kommando (`SPEC-139` Nr. 2 bleibt ohne einen; `SPEC-201` Nr. 3 und 11 unberührt); `harness/katalog-ausnahmen.txt` bleibt.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Spalten nach `SPEC-128` Nr. 3 a; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10). Die Zeilen V01 bis V05 sind die des Architect aus dem ungeschnittenen Plan von `slice-audit-log`, hier je Aufrufstelle, je Weg und je Zusage aufgeteilt (Architect, 2026-11-07); V04 ist geliefert (Brücke, `slice-audit-dienste`) und steht nicht in der Tabelle. Zwölf Zeilen zählt die Tabelle. Alle Mutationen sind lokal und werden nicht eingecheckt; »Aufrufstelle« heißt eine der vier in `src/hexagon/services/`. Die Fälle F1 bis F7 und die Zeitpunkte Z1 bis Z4 stehen unter der Tabelle.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| V01a | `Sessions().Delete` einer abgeschlossenen Session im Modus `record` schreibt genau eine Zeile `session_geloescht` mit `service_id` und `session_id` der Ausgangslage in `<DATA_DIR>/audit.jsonl` (`SPEC-234` Nr. 4 a, `SPEC-139` Nr. 3) | Fall F1 der Schleife über die Fälle (M1); die Datei gelesen und geparst, genau die erwarteten Zeilen | `session.Konfig.Audit` in `wurzel.go` auf eine zweite Instanz über `<DATA_DIR>/session-audit.jsonl` | andere |
| V01b | `Hochfahren` löscht je Service die Session, die älter ist als `RetentionMaxAge`, und lässt die jüngere; eine Zeile `session_aufbewahrung` je gelöschte Session, in beiden Modi (`SPEC-234` Nr. 4 b, `SPEC-139` Nr. 6) | Fälle F2 (`record`) und F3 (`replay`): zwei Services, je eine alte und eine junge Session gesät | `maintenance.Konfig.Audit` in `wurzel.go` auf eine zweite Instanz über `<DATA_DIR>/wartung-audit.jsonl` | andere |
| V01c | `Workspaces().Delete` schreibt genau eine Zeile `workspace_geloescht` mit `workspace_id` und ohne `service_id`, in beiden Modi (`SPEC-234` Nr. 4 c) | Fälle F4 (`record`) und F5 (`replay`) der Schleife | `workspace.Konfig.Audit` in `wurzel.go` auf eine zweite Instanz über `<DATA_DIR>/workspace-audit.jsonl` | andere |
| V01d | `Import().Import` im Modus `record` schreibt bei einem Fehler nach dem Fall »neu« genau eine Zeile `import_fehlgeschlagen` mit `service_id`, `import_id` und dem Code des zurückgegebenen Fehlers, und der Fehler des Aufrufs bleibt der des Imports; im Modus `replay` ist `Import()` `nil`, es gibt keinen Fall (`SPEC-234` Nr. 4 d, 5 b, `SPEC-139` Nr. 1 und 4) | Fälle F6 (unbekannte Session, Schritt (d)) und F7 (Fehler von `Sessions.Get` mit Fehlertext-Canary) der Schleife | `baselineimport.Konfig.Audit` in `wurzel.go` auf eine zweite Instanz über `<DATA_DIR>/import-audit.jsonl` | andere |
| V02a | Neustart: ein zweites `Build` auf demselben `DATA_DIR` (im selben Modus, ohne Auslöser, `RetentionMaxAge` 0) ändert die Datei nicht: die Bytes nach `Build`, nach `Hochfahren` und nach `Herunterfahren` des zweiten Laufs sind gleich den Bytes nach `Herunterfahren` des ersten (`SPEC-235` Nr. 2 und 7, `SPEC-139` Nr. 10) | Zeitpunkte Z3 und Z4 der Schleife über die sieben Fälle; Vergleich der Bytes mit Z2, gleiche Werte als eigener Fall | `Build` schneidet die Datei ab (`os.WriteFile(pfad, nil, 0o600)` nach `audit.New`) | andere |
| V02b | Neustart hängt an: löscht der zweite Lauf über `Hochfahren` eine weitere gesäte Session, stehen die Bytes des ersten Laufs unverändert als Anfang der Datei, danach genau eine neue Zeile nach dem Orakel (`SPEC-235` Nr. 7, `SPEC-139` Nr. 10) | Szenario N: Lauf 1 mit F2, Säen einer zweiten alten Session zwischen den Läufen, Lauf 2 mit `RetentionMaxAge`; `bytes.HasPrefix`, Zahl der Zeilen plus 1 | der Pfad trägt eine Laufkennung (`audit-<n>.jsonl`) statt `audit.jsonl` | andere |
| V03 | Der Canary steht in keinem Byte der Datei, roh, prozent-kodiert und als Base64 von `benutzer:passwort`; jeder Eintrag, den ein Dienst übergibt, trägt ihn nicht; jede Zeile hat nur Felder der geschlossenen Menge und je Art genau die Pflichtfelder (`LH-QA-04.a`, `SPEC-234` Nr. 2 und 3, `SPEC-235` Nr. 6) | Schleife M2 (172 Prüfungen) über Canary × Fall × Zeitpunkt; Orakel: die Bytes der Datei und die Aufzeichnung der Einträge, die Feldtabelle als Literal aus `SPEC-234` Nr. 2 | `h.audit` in `wurzel.go` durch einen Wrapper ersetzt, der das formgültige Passwort der ersten `UpstreamURL` als `KorrelationsID` setzt | andere |
| V05a | `Hochfahren` und `Herunterfahren` ohne Löschung legen die Datei nicht an, in beiden Modi: Wartung ohne Treffer, Anlegen der Services, erste Sessions (`record`), Listener auf `127.0.0.1:0`, `Herunterfahren` mit offenen ersten Sessions (`CloseAll`) und ein zweites `Herunterfahren` (`SPEC-139` Nr. 6 und 7, `SPEC-235` Nr. 2) | `DATA_DIR` nach `Hochfahren`, nach `Herunterfahren` und nach dem zweiten `Herunterfahren`: keine Datei `audit.jsonl` | `Herunterfahren` ruft `h.audit.Append` mit einem gültigen Eintrag nach `CloseAll`; `Hochfahren` legt die Datei an (`os.WriteFile` in Schritt (1)) | andere |
| V05b | Eine Ablehnung vor der Tatsache hinterlässt keinen Eintrag, und eine Wiederholung keinen zweiten: unbekannte Session (`ODR-E7003`), leere Kennung, offene Session (`ODR-E7009`), unbekannter Workspace (`ODR-E7001`), unbekannter Dienst beim Import (`ODR-E5005`), zweites `Delete` derselben Session (`SPEC-234` Nr. 4 und 10, `SPEC-124` Nr. 7) | Schleife über sechs Ablehnungen; danach keine Datei oder dieselbe Zahl Zeilen (Wiederholung: eine Zeile) | die Aufrufstelle in `session/loeschen.go` vor die Prüfung `offen()` gezogen | andere |
| V05c | Ein Fehler der Ablage beim `Build` (`DATA_DIR` fehlt, `DATA_DIR` ist eine Datei, `odata-recorder.db` ist ein Verzeichnis) kommt zurück, und `audit.jsonl` entsteht nicht (`SPEC-139` Nr. 2) | Schleife über drei Fehlerbilder, danach `os.ReadDir` des Verzeichnisses | der Root legt `audit.jsonl` vor `b.Oeffne` an (`os.OpenFile` mit `O_CREATE`), sichtbar am Fehlerbild `odata-recorder.db` als Verzeichnis | andere |
| V06 | Ein Schreibfehler der Ablage (`audit.jsonl` ist ein Verzeichnis) lässt `Build` gelingen und ändert in jedem Fall F1 bis F7 weder Bericht noch Fehler noch `Hochfahren`; das Log des Prozesses trägt je Eintrag genau eine Zeile `audit: Eintrag nicht geschrieben` mit `audit_fehlgeschlagen`, ohne Pfad und ohne Canary; die Datei bleibt ein Verzeichnis (`SPEC-139` Nr. 2, `SPEC-234` Nr. 5 b, `SPEC-235` Nr. 5) | Schleife über F1 bis F7 mit der Ablage als Verzeichnis; Ergebnis des Aufrufs gleich dem mit gesunder Ablage, Log in einen Puffer, Zeilen gezählt | `Build` öffnet `audit.jsonl` und gibt den Fehler zurück; ein Dienst gibt den Fehler von `Append` weiter | andere |
| V07 | Die Zahl der Aufrufstellen von `Audit.Append` im Code der Dienste ist die Zahl der Zeilen V01a bis V01d (vier), und jede Fundstelle steht in der Modell-Menge W; die Zahl der Arten ist vier (`SPEC-234` Nr. 2 und 4) | Zählprüfung: `go/parser` über `src/hexagon/services/` ohne `_test.go`, Zahl und Dateien gegen das Literal W; `len(support.Arten())` gleich 4; Zellen (Modus, Art) gleich Modus × Arten ohne (`replay`, `session_geloescht`) und (`replay`, `import_fehlgeschlagen`) | eine fünfte Aufrufstelle in `session/verwaltung.go`; die Aufrufstelle in `workspace/loeschen.go` entfernt | andere |

**Geliefert durch die Brücke, nicht Gegenstand dieses Slice (Geber: `slice-audit-dienste`):** V04 (dieselbe Instanz in allen Konfigs; `TestAuditEineInstanz`, Mutation: eigene Instanz für Session, Import oder Workspace, rot), der Pfad (`TestAuditPfadUndZeile`, dazu: `Build` allein legt die Datei nicht an) und der eine Logger (`TestAuditLoggerDesProzesses`). Die Zeilen V01a bis V01d fangen die zweite Instanz auch in der Datei; V04 fängt sie am Zeiger (die Überschneidung ist Absicht: V04 sieht eine Instanz, die Datei sieht eine Wirkung).

**Fälle und Zeitpunkte.** Fälle F1 bis F7 = (Modus, Art, Auslöser): F1 (`record`, `session_geloescht`, `Sessions().Delete`) · F2 (`record`, `session_aufbewahrung`, `Hochfahren`) · F3 (`replay`, `session_aufbewahrung`, `Hochfahren`) · F4 (`record`, `workspace_geloescht`, `Workspaces().Delete`) · F5 (`replay`, `workspace_geloescht`, `Workspaces().Delete`) · F6 (`record`, `import_fehlgeschlagen`, `Import().Import` mit unbekannter Session) · F7 (`record`, `import_fehlgeschlagen`, `Import().Import` mit Fehler von `Sessions.Get`). Zeitpunkte: Z1 nach dem Auslöser bei laufendem `Handle` · Z2 nach `Herunterfahren` · Z3 nach dem zweiten `Build` auf demselben `DATA_DIR` · Z4 nach `Hochfahren` und `Herunterfahren` des zweiten Laufs (`RetentionMaxAge` 0, kein Auslöser).

**Pflicht-Mengen mit Achsen** (`AGENTS.md` §3.17; `implement-slice` Schritt 14). Jede Menge trägt Achsen, Produkt, Generator, Orakel und Prüfgrenze; der Generator ist eine Tabelle im Test, die Zahl der Elemente prüft der Test selbst (`len` gleich dem Produkt).

| Menge | Achsen | Elemente | Generator | Orakel |
|---|---|---|---|---|
| M1 Einträge | Modus (2) × Art (4) × Auslöser; `record`: 5 (die Arten, `import_fehlgeschlagen` mit zwei Auslösern), `replay`: 2 (Aufbewahrung, Workspace) | 7 (F1 bis F7); die Zellen (`replay`, `session_geloescht`) und (`replay`, `import_fehlgeschlagen`) gelten nicht: `Sessions()` und `Import()` sind `nil` (`SPEC-139` Nr. 1 und 4) | Literal der Fälle; `len` gleich 7; Gegenprobe: die Menge der Zellen (Modus, Art) gleich Modus × `support.Arten()` minus der beiden | die Datei gelesen und je Zeile geparst: Art, Kennungen aus dem Säen, `error_code` als Text des Codes des zurückgegebenen Fehlers, `timestamp` in UTC mit neun Stellen im Fenster des Aufrufs, `correlation_id` nichtleer in der Form einer Kennung, Zahl der Zeilen; nie der Dienst und nie der Adapter |
| M2 Canary | Canary-Wert × Fall × Zeitpunkt: 3 Orte (Benutzer, Passwort, Query-Wert) × 2 Klassen (formgültig, formfremd) = 6 Werte, dazu der Fehlertext (nur F7) | (6 · 7 + 1 · 1) · 4 = 172 | Schleife über die Tabelle der Werte × Fälle × Zeitpunkte; `len` gleich 172 | `bytes.Contains` der Datei je Wert in drei Formen (roh, prozent-kodiert, Base64 von `benutzer:passwort`) und in der Aufzeichnung der Einträge; dazu das Struktur-Orakel (Feldmenge und Pflichtfelder je Art als Literal aus `SPEC-234` Nr. 2) |
| M3 Ablehnungen | Gegenstand (Session, Workspace, Import) × Grund (unbekannt, leer, offen, Dienst unbekannt, Wiederholung) | 6 (V05b) | Literal; `len` gleich 6 | keine Datei oder dieselbe Zahl Zeilen vor und nach dem Aufruf, Code des Fehlers aus `SPEC-124` Nr. 7, `SPEC-127` Nr. 4 |
| M4 Fehlerbilder der Ablage beim Build | Fehlerbild (`DATA_DIR` fehlt, `DATA_DIR` ist Datei, Datenbank ist Verzeichnis) | 3 (V05c) | Literal; `len` gleich 3 | Fehler kommt zurück, `audit.jsonl` nicht im Verzeichnis |

**Prüfgrenzen als akzeptierte Negative** (`AGENTS.md` §3.15 und §3.17; je Grund und Folge):

- *Weitere Kodierungen des Canary* (Hash, ROT13, Teilstücke kürzer als der Wert). Grund: die Menge der Kodierungen ist unbeschränkt; die Formprüfung des Adapters (`SPEC-234` Nr. 3) lässt nur Kennungen bis 64 Bytes und Codes zu, und das Orakel von M1 vergleicht jedes Feld mit dem erwarteten Wert der Ausgangslage. Folge: ein Root, der ein Geheimnis in Kennungsform kodiert in ein Feld legt, fiele an M1 (der Wert weicht ab), nicht an M2; ein Wert, den auch M1 nicht erwartet, ist kein Fall dieses Slice.
- *Ein Neustart, nicht n.* Grund: der Adapter hält keinen Zustand zwischen den Aufrufen (`SPEC-235` Nr. 2: öffnet je `Append`). Folge: ein Zustand, der erst nach dem zweiten Neustart wirkte, bliebe ungesehen.
- *Eine Form des Schreibfehlers* (`audit.jsonl` als Verzeichnis). Grund: andere Formen (voller Datenträger, Rechte, `/dev/full`) sind im Image nur teilweise auslösbar (`SPEC-235` *Ungeprüft*). Folge: dieselbe Wirkung des Dienstes (`SPEC-234` Nr. 5 b) ist für die anderen Formen vom Adapter-Vertrag, nicht vom Root belegt.
- *Die Grenze der Frist der Aufbewahrung* (Beginn genau gleich `jetzt − MaxAge`). Grund: der Root hat die Systemuhr, der Test kann den Gleichstand nicht herstellen. Folge: beide Seiten der Grenze (alt, jung) sind gefangen, der Gleichstand nur in den Fängern der Wartung (`SPEC-123` Nr. 6, `slice-session-wartung`).
- *Eintrag bei `ErrReferenziert` der Aufbewahrung, bei einem Kontextfehler des Imports und bei der Wiederholung mit gleichem Hash.* Grund: Quelle einer Baseline und erfolgreicher Import brauchen aufgezeichnete Exchanges, ein Aufbau jenseits dieses Slice. Folge: gefangen von den Fängern der Dienste (`slice-audit-dienste`), im Root über dieselbe Instanz (V04) verdrahtet.
- *Tokenwert der Admin-Authentisierung* (§1, `SPEC-139` *Ungeprüft*). Folge: die Datei wird gegen einen Tokenwert erst geprüft, wenn der Prozess ihn trägt.

**Modell-Menge der Zweige** (`BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, 6×; der Slice ist nicht nebenläufig, die Regel gilt für die Zweige). W = Aufrufstellen von `Audit.Append` in `src/hexagon/services/` (die vier Arten schreibt allein der Use Case, `SPEC-234` Nr. 7); die Zählprüfung V07 zählt sie per Syntaxbaum gegen dieses Literal. N = Wege, auf denen die Datei nicht entsteht.

| Element | Weg | Modus | Zeile mit Fänger |
|---|---|---|---|
| W1 | `session/loeschen.go`: `session_geloescht` | `record` | V01a |
| W2 | `maintenance/startup.go`: `session_aufbewahrung`, in `Hochfahren` (1) | beide | V01b |
| W3 | `workspace/loeschen.go`: `workspace_geloescht` | beide | V01c |
| W4 | `baselineimport/importer.go`: `import_fehlgeschlagen` | `record` | V01d |
| N1 | `Build` allein | beide | `TestAuditPfadUndZeile` (Bestand) |
| N2 | `Build` mit Fehler der Ablage | beide | V05c |
| N3 | `Hochfahren` ohne Löschung | beide | V05a |
| N4 | `Herunterfahren`, auch wiederholt, mit offenen Sessions | beide | V05a |
| N5 | Ablehnung vor der Tatsache, Löschung mit leerer Ablage, Wiederholung | beide, `record` für Session und Import | V05b |
| N6 | Aufbewahrung ohne Treffer (die junge Session bleibt) | beide | V01b (die junge Session im selben Fall) |
| N7 | Schreibfehler der Ablage bei W1 bis W4 | beide | V06 |
| N8 | Neustart ohne Auslöser | beide | V02a |
| — | `ErrReferenziert` der Aufbewahrung, Kontextfehler und gleicher Hash beim Import | — | benanntes Negativ (oben): kein Lauf des Roots erreicht sie ohne aufgezeichnete Exchanges |

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Nennt die Tabelle eine Klasse generierbar (hier: jeder Fall, jeder Canary, jeder Zeitpunkt), führt schon der Handoff die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-audit-dienste` und `slice-bootstrap-m2` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-audit-log` (`welle-betrieb`): der Adapter `audit.Datei` besteht erst dort (*Code-Abhängigkeit*, mittelbar über `slice-audit-dienste`).
- `slice-audit-dienste` (`welle-betrieb`): die Felder `Audit` der vier Konfigs und die Brücke in `wurzel.go` (Instanz, Pfad, Übergabe) bestehen erst dort; die Fänger dieses Slice laufen gegen `Build` mit dieser Verdrahtung (*Code-Abhängigkeit*).
- `slice-bootstrap-m2` (`welle-record-replay`, `done/`): hat `src/bootstrap/` für die Dienste und Listener verdrahtet; die Brücke ergänzt dessen Stand um die eine Instanz des Audit. Beide liegen in `done/`, die Reihenfolge der Dateien ist entschieden.
- Handbuch-Punkt der DoD: siehe §6, Adresse Planner.

**Start-Ausnahme je Slice** (Linie D, `welle-betrieb` §2): wie `slice-audit-log`; `welle-mock` ist nicht Bedingung, M2 ist es über `slice-bootstrap-m2`.

**Stand beim Start** (Planner, 2026-11-07): erfüllt, `slice-audit-dienste` und `slice-bootstrap-m2` liegen in `done/`; die Brücke (`audit.New` in `src/bootstrap/wurzel.go`, `audit_bruecke_test.go`) ist Bestand. Der Slice berührt keine Datei der Gate-Kette (Übergangsregeln Nr. 7: weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` noch den Stempel; in `in-progress/` liegt kein Slice der Gate-Kette, die beiden Zweige sind `slice-mock-lesen` und `slice-v4-schreiber-marker-text`, beide Linie B) und keine Datei der Linie B (`src/adapters/driven/odata/**`, `SPEC-160` bis `SPEC-199`; beide Zweige nennen `src/bootstrap/` nicht in ihrem Plan). Ändert Vertrag oder Code eine Datei der Gate-Kette (auch `.a-check.yml`, `.golangci.yml`) oder der Linie B, geht das als Rückgabe an den Planner (§6).

**Größenmarke** (`SPEC-128`; Faktoren allein dort; Block vom Architect, 2026-11-07): Der Block trägt den Teil des Architect: die Pfade sind die beiden neuen Testdateien und `export_test.go` (Bestand, 0 Zeilen, solange kein Zugang fehlt); die Zeilen der Brücke und alle anderen Dateien von `src/bootstrap/` fallen nicht in die Messung (`SPEC-128` Nr. 10 g). `produktcode` ist 0 (bestätigt, §3 Entscheidung 1), `fakes` ist 2 (`AuditAufzeichnung`, `SessionsMitFehlertext`, §3 Entscheidung 4). `basis`, `schaetzung` und `marke` setzt der Planner in eigenem Commit vor dem ersten Code-Commit; `basis` ist der Commit, der Vertrag, Tabelle und Raster abschließt und hinter dem letzten Commit von `slice-audit-dienste` liegt. Rechnung des Architect zur Kontrolle: 12 Zeilen, alle `andere` (keine nebenläufige), *S* = 85·0 + 80·12 + 200·2 + 0 = 1360, *M* = 1900; die Schnittfrage stellt sich nicht (*S* unter 2223, Abstand 863). Der Slice liefert keinen Leser fremder Bytes.

```groessenmarke
pfad: src/bootstrap/audit_betrieb_test.go
pfad: src/bootstrap/audit_betrieb_hilfen_test.go
pfad: src/bootstrap/export_test.go
produktcode: 0
fakes: 2
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke ohne gedeckten Eintrag.
- `in-progress` → `open` (blockiert): ein Slice, der `src/bootstrap/` ändert (etwa `slice-bootstrap-m3`), ändert den Aufbau von `Build` so, dass die Verdrahtung des Audit eine Entscheidung des Architect verlangt (Reihenfolge der Instanzen, `SPEC-139` Nr. 3), oder eine Änderung der Gate-Kette (nicht zu erwarten).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor. Mit diesem Slice liegt `slice-audit-log` vollständig geliefert vor (Teile A, E, V): die Welle-Bedingung „jede Löschung und jeder fehlgeschlagene Import schreibt einen Audit-Eintrag“ ist belegt.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- **Handbuch** (DoD): Die Datei `audit.jsonl` im Datenverzeichnis ist eine für den Anwender sichtbare Datei, entsteht aber erst bei einer Löschung, und kein gebundenes Beispiel gegen das gebaute Binary kann sie erzeugen (das Binary kennt `version` und `start`; eine Löschung ohne Admin-Endpunkt löst nur die Aufbewahrung bei vorhandenen Daten aus). Eine Aussage im Handbuch ohne Beispiel verstieße gegen `SPEC-200` Nr. 1 (K1) — **Ausgang:** weiter offen; Adresse: `slice-admin-sessions` (trägt das Kapitel in seiner DoD; Geber: dieser Slice und `slice-audit-log`).
- Reihenfolge der Verdrahtung gegenüber `slice-bootstrap-m2` (beide ändern `src/bootstrap/`) — **Ausgang:** entfallen durch den Start-Trigger (§4): dieser Slice startet nach ihm.
- Die Brücke aus `slice-audit-dienste` trägt den Kern der Verdrahtung (Entscheidung des Planner, 2026-11-06, Geber: `slice-audit-dienste`): sie bleibt, weil ohne sie zwischen den Slices kein grüner Stand besteht (`SPEC-139` Nr. 3 und `SPEC-234` Nr. 8 verlangen das Feld, `New` weist nil ab). Dieser Slice ist verkleinert auf die Fänger der Datei; Ziel, DoD, Tabelle und Marke sind nachgezogen. Fällt ein Fänger dieses Slice an der Brücke rot, ist es ein Befund an den Reviewer und keine stille Änderung — **Ausgang:** entfallen (entschieden; die Brücke ist Bestand).
- Ein zweiter `Build` im selben Prozess auf demselben `DATA_DIR` ist nicht unterstützt (`SPEC-106`, `SPEC-139` *Ungeprüft*), V02a und V02b fahren zwei `Build` nacheinander nach dem Herunterfahren — **Ausgang:** weiter offen (bei der Closure); die Fänger rufen `Herunterfahren` zwischen den Läufen.
- Telemetrie-Abbildung: `SPEC-231` Nr. 4 und 7 brauchen später `audit_fehlgeschlagen` und den Adapter `audit` — **Ausgang:** weiter offen; Adresse: `slice-telemetry-einbau-gateway` und `slice-telemetry-einbau-mock`.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster unten steht vor dem ersten Code-Commit, jede Zelle mit Ort oder „gilt nicht“ und Grund; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen (bis zum Review); Stand des Architect (2026-11-07): alle Zellen tragen einen Ort, die Lücken in `SPEC-139` (Nr. 2, 6, 7, 10) sind geschlossen.
- Produktcode 0 (Auftrag 1 in der Fassung des Start-Plans) — **Ausgang:** entfallen: vom Architect am Stand `4de0256` bestätigt (§3 Entscheidung 1); ein Fänger, der anderes zeigt, ist ein Befund an den Reviewer.
- Pflicht-Mengen mit Achsen und Zweig-Menge mit Zählprüfung — **Ausgang:** offen; Adresse: Implementer (Schleifen M1 bis M4, Zählprüfung V07, Bericht je Befund mit Menge, Zahl der Elemente und einer Mutation an einem Element, das der Befund nicht nannte, `AGENTS.md` §3.17); die Tabellen und Achsen stehen in §3.
- Schnittfrage und Marke — **Ausgang:** offen; Adresse: Planner (setzt `basis`, `schaetzung`, `marke` nach der Tabelle: 12 Zeilen, `fakes` 2, *S* 1360, *M* 1900); die Schnittfrage ist vom Architect entschieden (kein Schnitt, §3 Entscheidung 6).
- Kennungen und Katalog: keine neue `SPEC`-Nummer, kein Fehlercode, Schlüssel oder Kommando — **Ausgang:** entfallen (§3 Entscheidungen 7 und 8).
- Das Token der Admin-Authentisierung erreicht `Build` nicht: die DoD-Zeile der ersten Fassung nannte »Token« im Canary; der Canary trägt Userinfo und Query-Wert der `UpstreamURL` und den Fehlertext, der Tokenwert ist ein akzeptiertes Negativ mit Grund und Folge (§1, §3, `SPEC-139` *Ungeprüft*) — **Ausgang:** offen; Adresse: Planner (bestätigt die geänderte DoD-Zeile) und der Vertrag, der den Admin-Listener verdrahtet (trägt den Fänger gegen einen Tokenwert; Geber: dieser Slice).
- Ein Fall des Canary setzt voraus, dass `Build` URLs mit Zugangsdaten und Query annimmt (§3 Entscheidung 5) — **Ausgang:** offen; Adresse: Implementer (prüft vor dem ersten Test; ein Nein geht an den Architect).
- Gate-Kette und Linie B (Stand beim Start in §4): ein Bedarf an einer Änderung der Gate-Kette oder einer Datei der Linie B — **Ausgang:** weiter offen (nicht zu erwarten); Adresse: Planner.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund). Die Operationen sind `Build` über `DATA_DIR`, `Hochfahren`, `Herunterfahren` und die Auslöser über das `Handle`; die Argumente sind der Pfad des Datenverzeichnisses, der Modus und die Kennungen der Anfrage.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch oder Schreibfehler der Ablage | Wiederholung | Geheimnis im Gegenstand |
|---|---|---|---|---|---|---|
| `Build` mit der Instanz des Audit | `DATA_DIR` leer: Fehler der Ablage (`outbound.ErrPfad`), die Datei entsteht nicht: `SPEC-139` Nr. 2 (V05c) | vorhandenes leeres Verzeichnis: `Build` gelingt, legt `audit.jsonl` nicht an: `SPEC-139` Nr. 2 (`TestAuditPfadUndZeile`) | Modus außer `record` und `replay`: `model.ErrUngueltig`, vor `audit.New`: `SPEC-139` Nr. 2 und 5; `DATA_DIR` fehlt, ist eine Datei, Datenbank ist ein Verzeichnis: Fehler der Ablage, keine Audit-Datei: `SPEC-139` Nr. 2 (V05c); `audit.jsonl` ist ein Verzeichnis: `Build` gelingt: `SPEC-139` Nr. 2 (V06) | beendeter `ctx`: `Build` berührt die Datei nie: `SPEC-139` Nr. 2; Schreibfehler erst mit dem ersten Eintrag, ohne Wirkung auf den Dienst: `SPEC-235` Nr. 5, `SPEC-234` Nr. 5 b (V06) | zweites `Build` nach `Herunterfahren` hängt an: `SPEC-235` Nr. 7, `SPEC-139` Nr. 10 (V02a, V02b); ohne `Herunterfahren`: nicht unterstützt, `SPEC-139` *Ungeprüft* | `SPEC-235` Nr. 6, `SPEC-234` Nr. 3 (V03); ein Tokenwert: `SPEC-139` *Ungeprüft* |
| `Hochfahren` | gilt nicht: das einzige Argument ist `ctx` | leere Ablage und `RetentionMaxAge` 0: die Wartung löscht nichts, die Datei entsteht nicht: `SPEC-139` Nr. 6 (1), `SPEC-234` Nr. 4 b (V05a) | Sessions eines früheren Laufs im anderen Modus: die Aufbewahrung gilt in beiden Modi: `SPEC-234` Nr. 4 b, `SPEC-139` Nr. 6 (V01b: F2, F3) | beendeter `ctx` vor einem Schritt: nichts Weiteres, kein Eintrag: `SPEC-139` Nr. 6; Fehler der Ablage beim Eintrag: `Hochfahren` bricht deswegen nicht ab: `SPEC-234` Nr. 5 b (V06) | zweites `Hochfahren`: `model.ErrUngueltig`: `SPEC-139` Nr. 6; ein zweites `Startup` ohne neue Daten schreibt nichts: `SPEC-234` Nr. 10 | `SPEC-234` Nr. 3 (V03) |
| `Herunterfahren` | gilt nicht: das einzige Argument ist `ctx` | ohne vorheriges `Hochfahren`: schließt nur die Ablage, kein Eintrag: `SPEC-139` Nr. 7 (V05a) | gilt nicht: kein Argument außer `ctx` | läuft unter `context.WithoutCancel`: `SPEC-139` Nr. 7; schreibt nie einen Eintrag, auch nicht bei offenen Sessions: `SPEC-139` Nr. 7 (V05a) | wiederholbar, ein Lauf, kein Eintrag: `SPEC-139` Nr. 7 (V05a) | gilt nicht: `Herunterfahren` schreibt nichts: `SPEC-139` Nr. 7 |
| Auslöser über das `Handle` (`Sessions().Delete`, `Workspaces().Delete`, `Import().Import`) | leere Kennung: wie unbekannt, kein Eintrag: `SPEC-124` Nr. 7 (Kennung unverändert an den Port), `SPEC-234` Nr. 4 (V05b) | Löschung mit leerer Ablage: unbekannt (`ODR-E7003`, `ODR-E7001`), kein Eintrag: `SPEC-234` Nr. 4 (V05b) | offene Session `ODR-E7009`, unbekannter Dienst beim Import `ODR-E5005`: kein Eintrag: `SPEC-234` Nr. 4 (V05b); Import mit unbekannter Session: Eintrag: `SPEC-234` Nr. 4 d (V01d: F6) | Fehler eines Ports im Import (Schritte (d) bis (l)): Eintrag, Fehler des Aufrufs unverändert: `SPEC-234` Nr. 4 d und 5 b (V01d: F7); Schreibfehler der Ablage: V06; beendeter `ctx` nach der Tatsache: gilt nicht im Root, `SPEC-234` Nr. 5 c fangen die Dienste | zweites `Delete`: `ODR-E7003` bzw. `ODR-E7001`, kein zweiter Eintrag: `SPEC-234` Nr. 10 (V05b); zweiter gescheiterter Import mit gleicher `import_id`: zweiter Eintrag, gilt nicht im Root, `SPEC-234` Nr. 10 fangen die Dienste | `SPEC-234` Nr. 3, `SPEC-235` Nr. 6 (V03) |

**Zusatztabelle** (`AGENTS.md` §3.16: mehrere Verstöße und Abhängigkeit vom Modus), mit der Spalte *Fänger*:

| Fall | Ort | Fänger |
|---|---|---|
| Modus `record`: alle vier Arten | `SPEC-139` Nr. 4 (Session-Dienst und Import nur im Modus `record`) | V01a bis V01d, Modus `record` (F1, F2, F4, F6, F7) |
| Modus `replay`: Aufbewahrung und Workspace, kein Session-Dienst und kein Import | `SPEC-139` Nr. 1 und 4 | V01b und V01c, Modus `replay` (F3, F5); Zählprüfung V07 (Zellen der Menge M1) |
| Schreibfehler der Ablage je Aufrufstelle | `SPEC-234` Nr. 5 b, `SPEC-235` Nr. 5 | V06, Schleife über F1 bis F7 |
| Mehrere Verstöße am selben Schlüssel | gilt nicht: der Root meldet keinen Verstoß, jeder Dienst entscheidet seinen Rang (`SPEC-124` Nr. 7, `SPEC-127` Nr. 4) | — |

Die Fenster der Sperre und der Abbruch gehören zu `slice-audit-log`, die Aufrufstellen zu `slice-audit-dienste`.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-06, aus dem Plan von `slice-audit-log` übernommen) und beim Start nachgezogen (Planner, 2026-11-07; Zähler = Zahl der Dateien in `evidence/`). Dieser Slice hat zwei Decorators als Fakes (§3 Entscheidung 4). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (28×, `AGENTS.md` §3.17), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (42×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (34×, `AGENTS.md` §3.16), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14). Offen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (6×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure der Welle oder des nächsten Slice mit Nebenläufigkeit) gilt hier für die Zweig-Menge (Modell-Menge W und N in §3) und den Schreibfehler der Ablage (V06); `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (9×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (14×) für die Marke. Mit diesem Slice erreicht kein offener Eintrag ohne Ausgang 3×, außer den genannten.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
