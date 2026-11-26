# Verifikations-Report: slice-bootstrap-m6-listener (Admin-Listener im Composition Root) — 2026-11-26

**Verifikations-Art:** Code und Verhalten gegen DoD (`§2`), §3, §4, §6 des Plans und gegen die Spezifikation: DoD-Konformität, Plan-vs-Code-Diff, Fix-Runde gegen die Klasse (`AGENTS.md` §3.17). Kein Review und keine Validierung.

**Bezug:** `LH-FA-13`, `LH-RB-02`, `LH-QA-04`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)

**Gegenstand:** Zweig slice-bootstrap-m6-listener, `HEAD` `d4af5c30` (Code `9ff36049`, Review `1c7ba7e7`, Fix-Runde `d4af5c30`), Basis der Marke `14b75b45`. `SPEC-300` Nr. 4 und *Ungeprüft* zum abschließenden Trenner im Wortlaut von `main` (`e41dc6cf`); der Zweig trägt diesen Nachzug noch nicht.

**Rolle:** Verifier (Modul 11) · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-26

**Eingangs-Kontext:** Plan slice-bootstrap-m6-listener; `SPEC-300` (ohne `SPEC-301`), `SPEC-110`, `SPEC-139`, `SPEC-238`, `SPEC-200`, `SPEC-201`, `SPEC-037` Nr. 7; `LH-FA-13`; Review-Report 2026-11-26-slice-bootstrap-m6-listener-review; Nachweis-Stempel der Gates im Worktree.

## Urteil

**Freigegeben unter Auflagen.** Kein HIGH. Das Verhalten des gebauten Binary entspricht in allen geprüften Fällen dem Vertrag; ein Fehler im Produkt ist nicht gefunden. Zwei MEDIUM betreffen Fänger und Plan (V-1, V-2), ein LOW (V-3), drei INFO.

Auflagen vor dem Merge bzw. der Closure:

1. **V-1:** Zelle *`DATA_DIR` ist selbst ein Verweis* für die Prüfung der Dateien des Dienstes.
2. **V-2:** Das Negativ in B16 berichtigen (wirkliche Klasse, Adresse `TestE2EAdmin` mit Mutation in Teil D, `AGENTS.md` §3.13) oder den Fänger am Binary jetzt führen.
3. **Offene Punkte aus §6** vor Closure bzw. Merge entscheiden, vor allem *Start vor dem Welle-Trigger* (Ausgang offen → User: verneint er, geht der Slice nach `open/` zurück). Dazu: Merge von `main` und §6-Eintrag zum abschließenden Trenner durch den Planner, *Vereinigung in `Build`*, Welle-Datei, `tests/lasttest/` nach dem Merge beider Zweige.

V-3 geht an den Planner.

## Findings

| ID | Kategorie | Befund | Ort | Mutation / Beleg | Klasse |
|---|---|---|---|---|---|
| V-1 | MEDIUM | `SPEC-300` Nr. 4 vergleicht das Verzeichnis vor dem Namen mit `DATA_DIR` per `os.SameFile`, auch über einen Verweis. Geführt ist der Verweis allein auf der Seite von `ADMIN_TOKEN_FILE`; dass `DATA_DIR` selbst ein Verweis ist, führt kein Fall, und die Prüfgrenze in §3 nennt diese Achse nicht. Der Code im Stand weist ab. | `src/bootstrap/konfig_admin.go` · `os.Stat(dataDir)`; `TestAdminTokenVerstoss` (b), `TestAdminTokenPfadformen` | K1 `os.Lstat(dataDir)`: `TestAdmin\|TestKonfig\|TestStart\|TestWarte` grün. Am Binary, `DATA_DIR=<Verweis>`, `ADMIN_TOKEN_FILE=<Ziel>/audit.jsonl`: Stand Exit 1 mit *Datei, die der Dienst selbst schreibt*, `audit.jsonl` fehlt; K1 startet und schreibt das Token als Zeile in `audit.jsonl` | Fänger führt nur eine Seite eines Vergleichs (§3.15, §3.17; Klasse von Review F-1) |
| V-2 | MEDIUM | Das akzeptierte Negativ in B16 nennt die falsche Klasse und keine Adresse. (a) `println` fängt `make lint` (`forbidigo`), er bleibt allein in B16 grün. (b) Durch alle geprüften Fänger gehen `fmt.Fprint*(os.Stderr, …)` in `src/bootstrap` (Ausnahme im Lint-Profil) sowie `log.Printf` und `slog.Info` über den Standard-Logger, im Root und im Adapter; `log`/`slog` nennt das Negativ nicht. (c) Der Grund trägt nicht: `tests/e2e` hat `stdout`, `stderr` und `DATA_DIR` des echten Prozesses. `SPEC-037` Nr. 7 verlangt den Tokenwert in keinem Byte der Ausgaben für `TestE2EAdmin` (Teil D), B16 nennt das nicht, und die Zeile D04 führt keine Mutation dazu. | Plan §3 Zeile B16 und *Messung der Mutationen*; `SPEC-300` Nr. 12 a; `LH-QA-04` | N1 `println` im Adapter: B16 grün, Lint rot. N2 `os.Stderr` im Root: B16 grün, Lint grün, die Pakete `src/bootstrap` und `admin` vollständig grün. N3 `log.Printf`, N4 `slog.Info` im Root, N5 `log.Printf` im Adapter: B16 grün, Lint grün. N1 bis N3 am Binary mit der Probe unten: Token auf `stderr` in 4 von 4 Modi rot. Ein Test unter `tests/` sucht das Admin-Token nicht (gelesen) | akzeptiertes Negativ ohne zutreffende Klasse und Maß (§3.15, §3.10) |
| V-3 | LOW | `SPEC-300` Nr. 12 a nennt als Anlass das gültige Token **auf jeder** schreibenden Route; B16 führt eine (`POST /__admin/sessions`), 1 von 9 im Modus `record`, 1 von 8 sonst, ohne Negativ. Am Binary sind alle 9 bzw. 8 geprüft, ohne Fund. | `TestAdminListenerKeinGeheimnis`; Plan §3 B16 *einer schreibenden Route* | K24 Token als Header bei `DELETE`: B16 grün, Paket `admin` rot (Header-Mengen) | Plan verengt die Menge des Vertrags (§3.15) |
| V-4 | INFO | Das Handbuch sagt, eine Anfrage auf einer offenen Verbindung sei beim Beenden `503` mit `ODR-E5010`. Beobachtet: so im Modus `record` mit hängendem Austausch; hängt allein eine Anfrage am Admin-Listener (`replay`, `mock`, `hybrid`), war eine ruhende Keep-Alive-Verbindung 200 ms nach `SIGTERM` schon geschlossen (EOF). Eine geschlossene Verbindung ist nicht offen; kein Widerspruch zum Vertrag. | `docs/user/benutzerhandbuch.md` · Abschnitt *Admin-Schnittstelle*, Ergebnis | Probe am Binary | Satz ohne Fänger (`SPEC-200` Nr. 14 b) |
| V-5 | INFO | Hängt eine Anfrage bis zum Ende von `SHUTDOWN_TIMEOUT`, auch allein ein langsamer Client mit gültigem Token am Admin-Listener, endet der Prozess mit Exit 1 und `ODR-E9001`. Das entspricht `SPEC-300` Nr. 8 und `SPEC-139` Nr. 8. | `src/bootstrap/lebenszyklus.go` | Probe: 2,01 bis 2,03 s bei `SHUTDOWN_TIMEOUT=2s`, in 4 Modi | Verhalten im Betrieb |
| V-6 | INFO | `harness/gegenproben.stempel` gehört zu einem Stand vor diesem Slice; `make gegenproben` ist auftragsgemäß nicht gefahren. §3 führt `lebenszyklus_test.go` als *update*, die Datei ist unverändert. | Integrationsschritt; Plan §3 | `git log`, `git diff --numstat` | Nachlauf |

## Token „nie in Log und Ausgabe“: Urteil zum Negativ in B16

**Am Stand `d4af5c30` hält die Zusage am gebauten Binary.** Sie hält als Eigenschaft des Codes, nicht als gesicherte Eigenschaft: Außer `main` greift kein Produktcode unter `src/` auf `os.Stderr`, `os.Stdout`, `print`, den Paket-Logger `log` oder die Paket-Funktionen von `slog` zu (Suche im Quelltext); ein solcher Zugriff bliebe aber, außer bei `print` und `println`, in den Fängern oben grün (V-2).

Probe am Binary (eigenes Programm in einem Container ohne Netz), je Modus `record`, `replay`, `mock`, `hybrid` auf demselben `DATA_DIR`:

- 227 Abweisungen je Modus (17 Routen × 13 Anlässe und 6 weitere Formen): jede `401` mit `ODR-E2101`, genau 227 Zeilen `admin_abgewiesen` je Modus; gültiges Token nirgends `401`; ohne und mit kaputtem Body bleiben Sessions und Workspaces unverändert.
- Alle schreibenden Routen mit gültigem Token (9 im Modus `record`, 8 sonst, dort `POST /__admin/imports` `404`), Fehlerwege (404, 405, 409, übergroßer Body), Diagnose, OData-Anfragen.
- Das Token (ganz, je Hälfte, Base64, Hex, Großschreibung) steht in keinem Byte von `stdout` (leer), `stderr`, `audit.jsonl`, jeder anderen Datei unter `DATA_DIR` außer der Token-Datei und keiner Antwort (Body und Header), auch nach dem Beenden. Ebenso nach Bind-Fehlern, bei einer eigenen Token-Datei (`0400`) und wenn ein Client das Token selbst in Pfad, Query oder Body trägt (nicht zurückgegeben, nicht abgelegt; kein Anlass des Vertrags).
- Kontrolle der Probe: mit N1, N2 und N3 meldet sie den Fund in jedem Modus.

## Verhalten am Binary

| Gegenstand | Ergebnis |
|---|---|
| Start in allen Modi | bereit; Token-Datei 64 Zeichen, `0600`, über Modi und Neustarts gleich; `GET /healthz` ohne Token `200`; Sockets laut `/proc/net/tcp` je genau einer, Admin auf `127.0.0.1` bzw. `[::1]`, auch bei Default, `localhost` und `LISTEN=0.0.0.0` |
| Außerhalb Loopback | 13 Host-Formen (darunter `0.0.0.0`, `[::]`, leer, fremde Adressen, `127.0.0.2`, `127.1`, `LOCALHOST`, `[::ffff:127.0.0.1]`) × 4 Modi: Exit 1, Zeile am Schlüssel `ADMIN_LISTEN`, kein `DATA_DIR`, kein Listener; mit `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` oder `ADMIN_ALLOW_PLAINTEXT`: *nicht unterstützt*, ebenso kein Listener. Ohne TLS-Paar und mit ihm gibt es im Zwischenstand keinen Listener außerhalb von Loopback |
| Bind-Konflikte | belegt (je und beide), gleiche Adresse, `0.0.0.0` bzw. `[::]` gegen Loopback, `localhost`: Exit 1, die Zeilen sortiert, kein Socket bleibt; `127.0.0.2` und `127.0.0.1` mit gleichem Port starten (wie im Handbuch) |
| Token-Datei | Verstöße (Verzeichnis, FIFO, Verweis ohne Ziel, `0644`, zwei Zeilen, kurz, Elternteil fehlt, Pfad über eine Datei) und vier Dateien des Dienstes, je 4 Modi: genau eine Zeile, weder Inhalt noch Pfad, kein Listener, Dateien unverändert; ein Verstoß der Konfiguration legt nichts an |
| Abschließender Trenner (`SPEC-300` Nr. 4 in `main`) | `/`, `//`, `/.`, `/..` an `audit.jsonl` (fehlend und vorhanden), `odata-recorder.db` und `-wal`, 4 Modi: eine Zeile am Schlüssel (*nicht anlegbar* bzw. *nicht lesbar*), `audit.jsonl` unverändert bzw. weiter fehlend. Der Code entspricht dem Wortlaut |
| Herunterfahren | `SIGTERM` und `SIGINT`: Exit 0, Sockets zu; neue Verbindungen an beiden Listenern abgewiesen; eine zugelassene Anfrage wird noch `201`; mit hängender Anfrage eine Frist, nicht die Summe (V-5) |
| `LH-FA-13` | Sessions und Workspaces mit Kennung und Status; ohne Sessions leere Liste (`replay`, `mock`, `hybrid`; `record` hat die erste Session); Abweisung ohne Zustandsänderung; eine Session ohne Recorder bleibt `open` und ist nach dem nächsten Start nicht mehr offen |

`SPEC-300` Nr. 5 d (`Uhr`) und die Begrenzung von Nr. 12 b sind nach §6 und `slice-admin-abweisen-begrenzen` nicht geliefert (jede Abweisung eine Zeile); Nr. 3 (`umhuelle`, `warne`) und die drei TLS-Schlüssel gehören Teil C. Das sind benannte Zwischenstände, keine Befunde.

## DoD (§2)

Die Häkchen im Plan sind nicht gesetzt.

| Punkt | Stand | Beleg |
|---|---|---|
| Verdrahtung und Konfiguration (B01 bis B13, B19) | erfüllt, Auflage V-1 | Verhalten oben; die 20 Testnamen der Tabelle bestehen; Mutationen unten |
| Fänger des Tokens und der Abweisung (B14 bis B16) | Verhalten erfüllt; Fänger mit Auflage V-2, dazu V-3 | Probe am Binary, N1 bis N5, K24 |
| Andere `ADMIN_*` abgelehnt (B03) | erfüllt | Probe: `ADMIN_X=1` Verstoß, leer und Leerraum ohne Wirkung, `admin_listen` ignoriert; Präfix nicht mehr in `nichtGelieferte` |
| `make gates` grün | erfüllt laut Stempel | `.harness/state/gates-passed.head` = `d4af5c30`, `gates-passed.diffsha` gleich dem selbst berechneten Hash des Arbeitsbaums (vor diesem Report), Stempel 21:18 nach dem Commit 21:00. Ein Protokoll des Laufs liegt nicht vor |
| Review, Report liegt vor | erfüllt | Review-Report, Fix-Runde `d4af5c30` |
| Doku-Update öffentlicher Vertrag | erfüllt | `SPEC-201` Nr. 3 e; Verträge `SPEC-037` und `SPEC-041` über den Merge von `main`; keine Spezifikation durch den Slice geschrieben |
| Handbuch | erfüllt | `make handbuch-check`, `make katalog-check` Exit 0 (aus dem Cache bei gleichen Eingaben); Schlüssel, Endpunkte, Gründe vorhanden, `ADMIN_TLS_*` nicht genannt; Sätze zum Port-Konflikt am Binary bestätigt |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen (Closure) | §7 leer; §6 führt Ausgänge *offen* |

## Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Mutationen in Wegwerfkopien im Container; die Auswahl `TestAdmin|TestKonfig|TestStart|TestWarte` ist ohne Mutation grün.

| Befund | Klasse | Eigene Mutation (vom Review nicht genannt) | Ergebnis |
|---|---|---|---|
| F-1 Pfadformen | Formen des Pfads und beide Seiten des Vergleichs | K3 relative Pfade mit `..` ausgenommen; K4 nackter Datenbank-Name ausgenommen; K1 `Lstat` auf `DATA_DIR`; K2 `Lstat` auf das Verzeichnis; K28 Trenner gekürzt | K3, K4 rot (`TestAdminTokenPfadformen`); **K1 grün (V-1)**; K2 grün und gleichwertig (der Trenner am Ende folgt dem Verweis); K28 grün, nach *Ungeprüft* zulässig (Grund nicht zugesagt) |
| F-2 Bind-Paare | Ausgänge beider Binds | K6 Kontextfehler vereinheitlicht; K7 allein bei beendetem `ctx`; K8 zweiter Bind nach Fehler ausgelassen; K9 erster Listener bleibt offen | alle rot: **gedeckt** |
| F-3 Kommentar | Stellen, die dem Binary den Admin-Listener absprechen | Suche in `src`, `tests`, Verträgen, Handbuch | keine weitere Stelle |
| F-4 Handbuch | Sätze zum Port-Konflikt und zu `ADMIN_*` | Probe am Binary | bestätigt |
| F-5, F-6, I-2 | Plan, Verträge, eine Konstante | gelesen | F-5 in §6 an den Planner; F-6 und I-2 erledigt |
| I-1 Messung | Zeilen der Tabelle | gelesen und Stichprobe | Zeilen vorhanden; B16 *println grün* gilt für den Test, nicht für das Gate (V-2) |

Weitere Zusagen, Stichprobe: K10 eine Route weniger, K11 Rest der Frist des OData-Servers für den Admin-Server, K12 Fehler des Admin-Schließens verworfen außerhalb `record`, K13 `ADMIN_TOKEN_FILE` in `mock` ignoriert, K14 Prüfung in `hybrid` ausgelassen: alle rot.

## Plan-vs-Code

Die geänderten Pfade unter `src/`, `tests/` und `harness/katalog-ausnahmen.txt` stehen im Block. `make groessenmarke`: Messung 2192 von Marke 3000. Außerhalb des Blocks allein Handbuch, `SPEC-201`, Plan, der Hinweis in `slice-bootstrap-m6-transport` und der Review-Report. Abweichungen: V-3, V-6 und die in §6 geführten (*Vereinigung in `Build`*).

## Selbst gefahrene Sensoren

Je Lauf ein Container, `--network none` bzw. Build-Stufe, unter der gemeinsamen Sperre, `docker run` mit 4 GB und 4 CPU, Last des Hosts 2 bis 6.

| Sensor | Ergebnis |
|---|---|
| `go test` Auswahl oben, B16 | grün |
| `go test` `src/bootstrap` und `admin` vollständig (mit N2, die kein Test sieht) | grün (310 s, 36 s) |
| `go test` `tests/integration`, `tests/e2e/uebergang` | grün |
| `TestE2EProzess` am gebauten Binary | PASS, Fall `ADMIN_LISTEN` eingeschlossen |
| `-race -count=10`: `TestAdminZulassung`, `TestAdminFehlerKanal`, `TestAdminHerunterfahren`, `TestAdminBindung` | grün |
| `tools/harness/lint.sh` am Wurzelmodul | 0 Befunde |
| `make katalog-check`, `handbuch-check`, `zeichen-check`, `a-check`, `groessenmarke` | Exit 0 |

**Nicht gefahren:** `make gates` als Ganzes, `make gegenproben`, `make test-race` und `make test-integration` vollständig, `test-contract`, `test-determinism`, `test-security`, Lint der Module unter `tools/`, Schema-Gates, Raten der Zeilen B11 und B12 über `-count=10` hinaus. Der Admin-Listener ist von keinem anderen Host aus angesprochen (netzlos).

**Hinweis zum Lauf:** Mein erster Lauf der Probe blieb in der Probe selbst stehen (Lesen eines FIFO, den sie für den Fall *keine reguläre Datei* angelegt hat; Goroutine-Dump gesichert); kein Hänger des Binary. Die übrigen Fälle sind danach mit Zeitgrenze neu gefahren: 660 Prüfungen ohne Fehler.
