# Slice slice-fehlermodell-konfiguration: Fehlermodell mit ODR-Codes und Startkonfiguration

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-QA-05`, `LH-FA-12`, `LH-FA-04`, `LH-RB-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-109` (Fehlermodell, Abbildung) · `SPEC-110` (Startkonfiguration) · `SPEC-006` (Fehlerobjekt) · `SPEC-016` bis `SPEC-113` (Fehler-Codes, §4 der Spezifikation, darunter die neuen `SPEC-111` bis `SPEC-113`) · `SPEC-007` bis `SPEC-098` (Defaults, §3) · `SPEC-103` (Fehler des SQLite-Adapters) · `LH-FA-04.c` (Adressierung) · `LH-QA-05.a` · `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Domain-Modell trägt ein Fehlermodell (Code, HTTP-Status, Zielbereich, Fehlerobjekt nach `SPEC-006`, Vertrag `SPEC-109`), die Fehler der Outbound Ports und des SQLite-Adapters bilden auf diese Codes ab (darunter `ErrBeschaedigt`, `ErrBlobFehlt` und `ErrBlobBeschaedigt` auf `ODR-E3001`; Fehler ohne Zeile auf `ODR-E9001`), und der Composition Root lädt die Startkonfiguration der Welle aus Umgebung und `SERVICES_FILE` nach `SPEC-110` und lehnt Ungültiges mit `ODR-E2001` ab, darunter `MODE=mock` und `MODE=hybrid`, die diese Version nicht liefert.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Verwendung der Codes in den Services und im HTTP-Adapter — jeder Slice, der einen Code auslöst (`slice-record-service`, `slice-replay-service`, `slice-baseline-import`, `slice-http-adapter`), bildet ihn über dieses Fehlermodell ab; hier entsteht nur die Tabelle und die Abbildung.
- Schlüssel von Admin-API, Hybrid-Fallback und Listener-TLS (`ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`, dazu `STRICT_COVERAGE` und `PUBLIC_BASE_URL`) — diese Bereiche liefert die Welle nicht (Welle §6); die Konfiguration kennt nur, was `record`, `replay` und der Import brauchen, und lehnt einen gesetzten, nicht gelieferten Schlüssel mit `ODR-E2001` ab (`SPEC-110` Nr. 2), statt ihn still zu ignorieren.
- Die Zeile für einen Port, den es noch nicht gibt (Upstream-Transport): sie entsteht mit dem Port in der Abbildung (`SPEC-109`, Ungeprüft); hier gibt es nur die Fehler des Pakets `outbound` und des SQLite-Adapters.
- Der Start-Befehl und sein Verdrahten der Listener: die Konfiguration liefert typisierte Werte und die Meldung (`SPEC-110` Nr. 12); das Kommando, das sie aufruft, gehört nicht hierher.
- Das OData-V4-Fehlerformat des Mock-Dienstes — gehört zum Mock, der nicht in der Welle liegt; hier gilt das Fehlerobjekt `SPEC-006`.

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

- [ ] Fehlermodell im Domain-Modell: jeder Code aus §4 der Spezifikation mit HTTP-Status und Zielbereich, Fehlerobjekt nach `SPEC-006`, Abbildung der Fehler der Outbound Ports auf Codes (Tabelle in der Spezifikation, Test über alle Fehler des Pakets `outbound` und des SQLite-Adapters).
- [ ] Konfiguration der Schlüssel `MODE`, `LISTEN`, `DATA_DIR`, `SERVICE_ID`, `SERVICE_ROOT_PATH`, `UPSTREAM_URL`, `SERVICES_FILE`, `REPLAY_*`, `MATCH_*`, `REDACT_*`, `MAX_CAPTURE_BODY`, `UPSTREAM_*`, `SHUTDOWN_TIMEOUT`, `RETENTION_MAX_AGE`, `SQLITE_BUSY_TIMEOUT`, `VOLATILE_HEADERS`, `LINK_REWRITE` mit den Defaults aus §3 der Spezifikation; `ODR-E2001` bei ungültigem Wert, Doppelkonfiguration (Datei und Einzelschlüssel), überlappenden Roots und fehlendem Pflichtwert.
- [ ] Tests der Randformen aus §6 gegen die Spezifikation.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `src/hexagon/model/` | update | Fehlermodell: `Code` (Aufzählung, Nullwert ist kein Code), Status je Code (`Stati`, `StatusErlaubt`), `Zielbereich`, feste `Message`, `Fehlerobjekt` mit `NeuesFehlerobjekt` |
| `src/hexagon/ports/outbound/` | update | Abbildung der Port-Fehler auf Codes (`SPEC-109` Nr. 5), neue Fehler `ErrBeschaedigt` und `ErrPfad` des Pakets (Nr. 7); `Abbilden(err, Gegenstand)` als reine Funktion über eine Tabelle in der Reihenfolge der Spezifikation, `EintraegeAbbildung` nennt die Namen der Tabelle für den Vollständigkeitstest |
| `src/adapters/driven/sqlite/` | update | `ErrBeschaedigt`, `ErrSchemaNeuer`, `ErrFremdeDatenbank` und `ErrPfad` verpacken die Fehler des Pakets `outbound` (`SPEC-109` Nr. 7); Namen und Meldungen bleiben, nur `errors.Is` kommt hinzu; der unexportierte Typ `verpackt` (`Error`, `Unwrap`) trägt Text und Ziel, jede Variable `Err…` des Adapters ist `verpacke(Text, outbound.Err…)` |
| `src/bootstrap/` | update | Konfiguration aus Umgebung und `SERVICES_FILE`, Defaults, Validierung: `LadeKonfiguration(environ)` liefert `Konfiguration` oder `*KonfigFehler`, `LadeOderMelde` schreibt die Meldung nach stderr und liefert `ExitKonfiguration`; Dateien `konfig.go`, `konfig_werte.go`, `konfig_dienste.go`; `none` prüft `pruefeNone` vor der Formprüfung der Einträge, der Port der URL `portGueltig`, die Meldung sortiert `fehler()` nach Schlüssel (Bytefolge), dann Grund; der Start-Befehl ruft sie in diesem Slice nicht auf (§1) |
| `spec/spezifikation.md` | update | **vor dem Code geschrieben:** `SPEC-109` (Fehlermodell, Abbildungstabelle), `SPEC-110` (Syntax der Werte, Verhalten bei `mock` und `hybrid`), Codes `SPEC-111` bis `SPEC-113` |
| Tests der Pakete | neu | Tabellentests je Schlüssel und je Fehler nach `LH-QA-05.a` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16); beide gehören zum Vertrag `SPEC-109` und `SPEC-110`, der vor dem ersten Code-Commit in der Spezifikation steht (`AGENTS.md` §3.12). Die Fänger sind Namen geplanter Tests; der Implementer legt sie unter diesem Namen an oder nennt den Namen, den er stattdessen vergibt, im Bericht.

**Zusage · Fänger** (jede Zeile löst der Reviewer mit einer Mutation auf):

| Zusage | Fänger |
|---|---|
| Fehlerobjekt hat genau `code`, `message`, `target`, kein Status-Feld (`SPEC-109` Nr. 1) | `TestFehlerobjektJSON`: Vergleich der Schlüsselmenge des serialisierten Objekts |
| Status und Zielbereich je Code, `ODR-E4001` nur 502 oder 504, jeder Code aus §4 ist abgedeckt (Nr. 2) | `TestCodeTabelle`: Tabelle im Test gegen das Modell; dass die Tabelle im Test der Spezifikation entspricht, prüft die Test-Stufe nicht (sie sieht `spec/` nicht) — akzeptiertes Negativ, der Reviewer gleicht ab |
| `message` ist je Code fest, nichtleer und enthält keinen Wert aus dem Fehler (Nr. 3) | `TestMessageFest`: gleiche `message` bei jeder Bildung, je Code ein eigener nichtleerer Text ohne Platzhalter; dass kein Wert aus dem Fehler eingeht, folgt aus der Signatur (`NeuesFehlerobjekt` nimmt nur den Code) und ist kein Testfall |
| Ein Code außerhalb von §4 ist nicht darstellbar (Nr. 2) | `TestFehlerobjektUngueltigerCode`: `0`, `-1`, `999` liefern `ErrUngueltig` und tragen keine Eigenschaft |
| Abbildung je Zeile und Gegenstand der Tabelle in Nr. 5, Fall `Sonstiges` → `ODR-E9001` | `TestAbbildungTabelle`: eine Zeile je Paar Fehler × Gegenstand |
| Abbildung durch Verpacken und `errors.Join`, erster Treffer gewinnt (Nr. 5) | `TestAbbildungVerpackt`: `fmt.Errorf("%w")`, `errors.Join(ErrUngueltig, ErrBlobFehlt)` |
| `nil` und Kontextfehler: „kein Code“, auch verpackt und zusammen mit einem Fehler der Tabelle in einem `errors.Join` (Nr. 6) | `TestAbbildungKeinCode`: `Join(Canceled, ErrBlobFehlt)`, `Join(ErrBlobFehlt, DeadlineExceeded)`, `Errorf("%w", Canceled)` |
| Fehler des SQLite-Adapters verpacken die Fehler des Pakets (Nr. 7) | `TestSqliteFehlerAbbildung` im Adapter: `Open` auf Nicht-Datenbank, neuere Version, fremde Datenbank, leeren Pfad, dann `Abbilden` → `ODR-E3001` bzw. `ODR-E2001` |
| Jede Variable `Err…` hat einen Eintrag (Nr. 8) | `TestAbbildungVollstaendig` (Paket `outbound`) und `TestSqliteFehlerVollstaendig` (Adapter): durchsuchen die Quelltexte mit `go/parser`; eine neue Variable ohne Zeile macht sie rot, im Adapter verlangt der Test `verpacke(…, outbound.Err…)` mit einem Namen aus der Tabelle |
| Leer gilt als ungesetzt, Leerraum am Rand wird entfernt (`SPEC-110` Nr. 1) | `TestKonfigLeer`: Tabelle je Schlüssel mit `""`, `"  "`, `" 8 MiB "` |
| `REDACT_*` ergänzen den Default und ersetzen ihn nie; `none` nur als einziger Eintrag, klein geschrieben, bei `MATCH_VARY_HEADERS` und `VOLATILE_HEADERS`; sonst ungültig (Nr. 1) | `TestKonfigListenErgaenzen`, `TestKonfigNone`: `none` allein bei beiden Mengen gültig und leer; `None`, `NONE`, `none,Accept`, `Accept,none` bei beiden ungültig; `none` bei `REDACT_HEADERS` und `REDACT_QUERY_PARAMS` ungültig |
| Header-Mengen in kanonischer Schreibweise (Nr. 6) | `TestKonfigHeaderSchreibweise`: `odata-version`, `X-API-KEY`, `etag` ergeben `Odata-Version`, `X-Api-Key`, `Etag`; Default-Mengen in derselben Form |
| Defaults aus §3 für jeden Schlüssel der Nr. 2, darunter `DATA_DIR=data` | `TestKonfigDefaults`: leere Umgebung mit `MODE=record` und `UPSTREAM_URL`, alle Werte gegen §3; `DATA_DIR` ungesetzt, nach `t.Chdir(t.TempDir())`: `Konfiguration.DataDir` ist `data` und `data` liegt im Arbeitsverzeichnis |
| Gesetzter, nicht gelieferter Schlüssel ist `ODR-E2001`; fremde Variablen werden ignoriert (Nr. 2) | `TestKonfigNichtGeliefert`, `TestKonfigFremdeVariable` |
| Aufzählung und Wahrheitswert: nur die Schreibweise aus §3 (Nr. 3) | `TestKonfigEnum`, `TestKonfigBool` mit `Record`, `1`, `yes`, `TRUE` |
| Dauer: Einheiten, Grenzen, Überlauf, `RETENTION_MAX_AGE=0` (Nr. 4) | `TestKonfigDauer`: `30s`, `5000ms`, `30d`, `720h` gültig; `30`, `-1s`, `1.5h`, `1h30m`, `1w`, `0s` bei `UPSTREAM_TIMEOUT`, `106752d`, `2147483648ms` bei `SQLITE_BUSY_TIMEOUT` ungültig |
| Größe: Einheiten, Leerzeichen, Grenzen (Nr. 5) | `TestKonfigGroesse`: `8 MiB`, `8MiB`, `8388608`, `1GiB` gültig; `0`, `8 MB`, `8  MiB`, `1025MiB`, `-1`, `8.5MiB` ungültig |
| Liste: leerer Eintrag, Token, Pfadform (Nr. 6) | `TestKonfigListe`: `a,,b`, `a,`, `Bad Header`, `$.a`, `$..a.b` ungültig |
| Weitere Schlüssel: `LISTEN`, `SERVICE_ID`, Root, URL (Nr. 7) | `TestKonfigLISTEN`, `TestKonfigServiceID`, `TestKonfigRoot`, `TestKonfigURL` mit Randformen (Port 0, 65536, leerer Host, `user:pass@`, Query, `ftp:`); `LISTEN`: Host `a_b.c-d` gültig, `a b`, `a/b`, `[abc]` ungültig; URL: Port 65535 gültig, 0, 65536, 99999 und `host:` ungültig |
| `mock` und `hybrid` gültig, Start scheitert; Default `mock` scheitert ebenso (Nr. 8) | `TestKonfigModus`: vier Werte, leerer `MODE`; die Meldung nennt `MODE` und beide lieferbaren Werte; `TestKonfigModusUngueltig`: ein ungültiger Wert wird einmal gemeldet |
| Doppelkonfiguration, Dateiformat, Pflichtfelder, doppelte `id` (Nr. 9) | `TestKonfigServicesFile`: Tabelle über die Fälle der Nr. 9; zusätzlich Feld doppelt im Objekt (letzter Wert gilt, kein Verstoß), `null` und `""` und `" "` für `id`, `service_root`, `upstream_url` (fehlend; bei `upstream_url` nur im Modus `record` ein Verstoß), `" x "` für `id` (getrimmt gültig), Element `null` (kein Objekt) |
| Überlappende Roots an der Segmentgrenze (Nr. 9) | `TestKonfigRootUeberlappung`: `/` und `/a` ja, `/a` und `/ab` nein, `/a` und `/a/b` ja, gleiche ja |
| Ein Feld falschen Typs meldet allein „muss eine Zeichenkette sein“, bei `upstream_url` in jedem Modus (Nr. 9) | `TestKonfigDienstefeldTyp`: `id`, `service_root`, `upstream_url` je als Zahl in `record`; zusätzlich `upstream_url: 1` in `replay` und in `mock` (mit gültigem `id` und `service_root`): genau die Zeile `SERVICES_FILE[0].upstream_url: muss eine Zeichenkette sein` und im Modus `mock` die Zeile `MODE`, sonst nichts. Mutation: Typfehler von `upstream_url` nur im Modus `record` melden: rot |
| `UPSTREAM_URL` Pflicht in `record` je Service; in `replay` nicht erforderlich, aber Form geprüft; bei ungültigem `MODE` entfällt die Pflicht (Nr. 9) | `TestKonfigRecordOhneUpstream`; `TestKonfigUpstreamModus`: `replay` ohne URL ohne Verstoß, `replay` mit `ftp://x` ein Verstoß, `MODE=Record` ohne URL genau ein Verstoß (`MODE`) |
| TLS-Widersprüche (Nr. 10) | `TestKonfigTLSWiderspruch`: vier Kombinationen mit gültigen Werten |
| Ein ungültiger Wahrheitswert meldet nur die Form; der Widerspruch wird nur bei zwei gültigen Werten geprüft; je Schlüssel höchstens eine Zeile (Nr. 10, 12) | `TestKonfigTLSUngueltig`: genau die Zeilenmenge je Fall, nicht nur „Fehler ja/nein": `VERIFY=yes` und `VERIFY=1` je ohne und mit `SKIP=true` ergeben allein `UPSTREAM_TLS_VERIFY: Wahrheitswert muss true oder false sein`; `VERIFY=false` mit `SKIP=yes` ergibt allein die Formzeile von `SKIP`; `VERIFY=yes` mit `SKIP=yes` zwei Zeilen zu verschiedenen Schlüsseln; keine Zeile enthält „braucht" oder „widerspricht". Mutation: Widerspruch auch bei ungültigem Wert prüfen: rot |
| Reihenfolge: reine Prüfung vor Dateien, `DATA_DIR` erst danach angelegt, bei Verstoß nichts angelegt; `UPSTREAM_CA_FILE` lesbar mit Zertifikat (Nr. 11) | `TestKonfigDatenverzeichnis`, `TestKonfigCADatei` mit `t.TempDir()` |
| Meldung: alle Verstöße, nach Schlüssel (Bytefolge), dann Grund sortiert, ohne Wert, Exit-Code 1 (Nr. 12) | `TestKonfigMeldung`: Wert `geheim-123` kommt in keiner Zeile vor; `TestKonfigMeldungSortierung`: gleiche Zeilen erscheinen einmal (dritte `id` eines Wertes meldet `doppelte id` für `[2]` einmal, nicht zweimal); die Folge nach Grund bei gleichem Schlüssel ist als Totalordnung zugesagt, aber nicht beobachtbar, weil jeder Schlüssel höchstens einen Grund meldet (Nr. 9 falscher Typ, Nr. 10 ungültiger Wahrheitswert; Fänger dieser Aussage: `TestKonfigDienstefeldTyp`, `TestKonfigTLSUngueltig`); die Sortierung nach Grund selbst hat keinen Fänger — akzeptiertes Negativ, Grund: ein Fall, der sie beobachtbar macht, existiert nicht, und ein künstlicher Test über innere Funktionen prüfte nur die Implementierung; ein Feld falschen Typs (`id: 1`, `service_root: 1`, `upstream_url: 1` im Modus `record`) meldet genau eine Zeile „muss eine Zeichenkette sein“ (`TestKonfigDienstefeldTyp`, Nr. 9); Schlüssel `SERVICES_FILE[0]` (unbekanntes Feld) und `SERVICES_FILE[0].id` (Pflichtfeld) stehen in der Reihenfolge der Schlüssel, nicht der Zeilen, dazu `[10]` vor `[2]`; `TestKonfigMeldung` prüft `slices.IsSorted` der Schlüssel, nicht der Zeilen |
| Wiederholung liefert dasselbe (Nr. 13) | `TestKonfigWiederholung`: zweimal laden, gleiche Werte, gleiche Meldung |

- Die Konfiguration liegt im Composition Root, weil nur er alle Pakete kennen darf; die Services erhalten typisierte Werte über ihre Konstruktoren, nie die Umgebung.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` ist eröffnet; kein Vorgänger-Slice. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 800 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung lag bei rund 600 Zeilen in derselben Metrik.
  **Überschreitung:** `git diff --numstat 89db02c..<letzter Commit des Slice>` über die Pfade aus §3 ergibt 2315 (Einfügungen plus Löschungen, Tests eingeschlossen; davon Produktcode ohne Tests 1097). Die Marke ist überschritten; die Rückführung nach `next` unterbleibt, weil Fehlermodell, Abbildung und Konfiguration geschrieben, geprüft und im Review sind und ein Schnitt jetzt nur Übergabe kostete. Der Slice wird ungeteilt geschlossen; die Schätzung der Test-Anteile (`AGENTS.md` §3.14: etwa das Doppelte des Produktcodes) war zu niedrig angesetzt.
- `in-progress` → `open` (blockiert): die Abbildung der Port-Fehler lässt sich unter den Kanten von `a-check` nicht an der geplanten Stelle ablegen (Abbildung im Paket `outbound`, Fehlerobjekt im Modell, die Adapter verpacken nur Fehler des Pakets `outbound`, `SPEC-109` Nr. 7); sie wandert dann in die Schicht, die beide Seiten kennen, und der Plan wird nachgezogen.

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

- Die Syntax der Werte (`8 MiB`, `8MiB`, `8388608`, `d` bei Dauern) — **Ausgang:** entfallen; entschieden in `SPEC-110` Nr. 4 bis 6: Dauern mit Pflicht-Einheit `ms`/`s`/`m`/`h`/`d`, Größen mit oder ohne ein Leerzeichen und auch als Bytes, Grenzen 1 ms bzw. 1 Byte bis 1 GiB.
- Was ein Prozess mit `MODE=mock` oder `MODE=hybrid` tut — **Ausgang:** entfallen; entschieden in `SPEC-110` Nr. 8: gültiger Wert, Start scheitert mit `ODR-E2001`; das gilt auch für den Default `mock` ohne gesetzten `MODE`.
- Leere Werte (gesetzt-aber-leer gegen ungesetzt, leere Liste gegen Default) — **Ausgang:** entfallen; entschieden in `SPEC-110` Nr. 1: leer ist ungesetzt, eine leere Liste der Default, `none` nur für die zwei Mengen, deren Default nicht leer ist.
- Randformen (`AGENTS.md` §3.12, §3.16) — **Ausgang:** entfallen; das Raster unten steht vor dem ersten Code-Commit, jede Zelle nennt ihren Ort.
- Die Abbildung der Port-Fehler braucht zwei neue Fehler im Paket `outbound` und eine Änderung des SQLite-Adapters (`errors.Is`, keine Verhaltensänderung) — **Ausgang:** weiter offen; Beleg des Implementers: die vorhandenen Testdateien des Adapters sind unverändert (`git diff` über `sqlite_test.go`, `repositories_test.go`, `baselines_test.go`, `workspaces_test.go` ist leer) und `make test` ist grün; neu ist allein `fehler_test.go`.
- Das Format von `SERVICES_FILE` (JSON statt YAML, um keine Abhängigkeit hinzuzufügen) und die Obergrenze 1 GiB für `MAX_CAPTURE_BODY` sind Setzungen dieses Vertrags, nicht des Lastenhefts — **Ausgang:** weiter offen bis zur Closure; der Nutzer hat beide Setzungen bestätigt, ein Wechsel wäre eine Änderung von `SPEC-110`.
- Das Modell kennt `ErrRevisionVeraltet` und `Sonstiges`-Fehler nur als `ODR-E9001` (`SPEC-109` Nr. 5); der Mock gehört nicht zur Welle — **Ausgang:** weiter offen; akzeptiertes Negativ, ein Vertrag, der den Fehler auswertet, nennt seinen Code.
- Randformen, die das Review beim Code fand und `SPEC-109`/`SPEC-110` nicht nannten (`AGENTS.md` §3.12) — **Ausgang:** eingetreten und entschieden; der Architect hat jede in der Spezifikation festgelegt (Historie 2026-10-20), die Zellen unten und die Zeilen der Tabelle *Zusage · Fänger* tragen sie. Der Code trägt sie: `none` nur allein und klein (Nr. 1), Dienste-Datei mit Leerraum-Trimmen (Nr. 9), `host:` ohne Port und Port 1 bis 65535 in der URL (Nr. 7), Meldung nach Schlüssel, dann Grund sortiert (Nr. 12), ein Feld falschen Typs mit einer Zeile statt zwei (Nr. 9; Verifikation V-1); unverändert bleiben letzter Wert bei doppeltem Feld, `null`/leer als fehlend, Hostzeichen, Kontextfehler vor Tabelle, kanonische Header, Form von `UPSTREAM_URL` in jedem Modus, Pflicht nur bei gültigem Modus. Akzeptiertes Negativ mit Grund: Größe von `UPSTREAM_CA_FILE` ohne Grenze (Betreiber-Datei, einmal beim Start) und Index als Text (`[10]` vor `[2]`, deterministisch).

**Raster Operation × Argumentart** (`AGENTS.md` §3.16). Jede Zelle nennt Nummer der Spezifikation oder „gilt nicht“ mit Grund.

| Operation | fehlender Wert | leerer Wert | unbekannter Wert | fremder / ungültiger Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| Konfiguration laden, Aufzählung und Wahrheitswert | Default (`SPEC-110` Nr. 1) | wie fehlend (Nr. 1) | nicht in der Aufzählung: `ODR-E2001` (Nr. 3) | falsche Schreibweise, `1`, `yes`: `ODR-E2001`, nur die Formzeile, der TLS-Widerspruch entfällt (Nr. 3, 10) | gilt nicht: synchron ohne Kontext, Signal beendet den Prozess (Nr. 13) | gleiches Ergebnis (Nr. 13) |
| Konfiguration laden, Dauer und Größe | Default (Nr. 1) | wie fehlend (Nr. 1) | unbekannte Einheit: `ODR-E2001` (Nr. 4, 5) | Vorzeichen, Bruch, Verkettung, 0, Überlauf, Grenzen: `ODR-E2001` (Nr. 4, 5) | gilt nicht (Nr. 13) | gleiches Ergebnis (Nr. 13) |
| Konfiguration laden, Liste | Default (Nr. 1) | Default bzw. keine Ergänzung (Nr. 1); leerer Eintrag: `ODR-E2001` (Nr. 6) | `none` nur allein und klein bei den zwei Mengen, sonst `ODR-E2001` (Nr. 1) | Token, Pfadform, Doppelte (ohne Wirkung): Nr. 6 | gilt nicht (Nr. 13) | gleiches Ergebnis (Nr. 13) |
| Konfiguration laden, `LISTEN`, Roots, URL, `SERVICE_ID` | Default bzw. Pflichtwert fehlt: `ODR-E2001` (Nr. 1, 9) | wie fehlend (Nr. 1) | nicht aufgelöst, nicht erreichbar, Größe der CA-Datei: ungeprüft, Ende von `SPEC-110` | Form verletzt, Hostzeichen, Port außerhalb 1 bis 65535: `ODR-E2001` (Nr. 7); `UPSTREAM_URL` in jedem Modus geprüft (Nr. 9) | gilt nicht (Nr. 13) | gleiches Ergebnis (Nr. 13) |
| Konfiguration laden, Schlüssel, die die Version nicht liefert | ungesetzt: kein Fehler (Nr. 2) | wie ungesetzt (Nr. 1, 2) | fremde Variable: ignoriert (Nr. 2) | gesetzt: `ODR-E2001` (Nr. 2) | gilt nicht (Nr. 13) | gleiches Ergebnis (Nr. 13) |
| `MODE` prüfen | Default `mock`, Start scheitert (Nr. 8) | wie fehlend (Nr. 1, 8) | `ODR-E2001`; die Pflicht von `UPSTREAM_URL` entfällt bis dahin (Nr. 3, 9) | `mock`, `hybrid`: `ODR-E2001` (Nr. 8) | gilt nicht (Nr. 13) | gleiches Ergebnis (Nr. 13) |
| `SERVICES_FILE` lesen | ohne Datei ein Service (Nr. 9) | wie fehlend (Nr. 1) | unbekanntes Feld, doppelte `id`: `ODR-E2001`; `null` und leer wie fehlend (Nr. 9) | falscher Typ (nur diese eine Zeile, Nr. 9), Element `null`, leeres Array, Nicht-JSON, Anhang nach dem Array, über 1 MiB, Überlappung, Doppelkonfiguration: `ODR-E2001`; Feld doppelt: der letzte Wert (Nr. 9) | gilt nicht (Nr. 13) | gleiches Ergebnis (Nr. 13) |
| Dateien prüfen (`UPSTREAM_CA_FILE`, `DATA_DIR`) | ungesetzt: kein Zugriff (Nr. 11) | wie ungesetzt (Nr. 1) | Datei fehlt, nicht lesbar, ohne Zertifikat, Nicht-Ordner: `ODR-E2001` (Nr. 11) | wie unbekannt (Nr. 11) | Signal während des Anlegens: Verzeichnis kann bestehen (Nr. 13) | `DATA_DIR` anlegen ist idempotent (Nr. 11) |
| Fehler abbilden (`Abbilden(err, Gegenstand)`) | `nil`: kein Code (`SPEC-109` Nr. 6) | gilt nicht: ein Fehlerwert ist nie leer | Fehler ohne Zeile: `ODR-E9001` (Nr. 5) | Gegenstand passt nicht zur Zeile: `ODR-E9001` (Nr. 5); mehrere Treffer: erster (Nr. 5); Kontextfehler neben einem Treffer: kein Code (Nr. 6) | Kontextfehler: kein Code (Nr. 6) | gilt nicht: reine Funktion ohne Zustand |
| Fehlerobjekt bilden und serialisieren | gilt nicht: der Code ist Pflichtargument, ein Code außerhalb von §4 ist nicht darstellbar (Nr. 2) | `message` leer: ausgeschlossen (Nr. 3) | `ODR-E4001` mit anderem Status als 502/504: ungültig (Nr. 2) | Wert aus Request oder Daten in `message`: ausgeschlossen (Nr. 3) | gilt nicht: reine Funktion | gleiche Bytes (`SPEC-109` Nr. 1) |

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (2×); die Randformen der Konfiguration (Wert fehlt, leer, unbekannt, doppelt) stehen als Raster in §6, nicht als Liste.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
