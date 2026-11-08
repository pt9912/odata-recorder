# Verifikation: slice-bootstrap-m3 (Teil A) — 2026-11-08

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-110` Nr. 2, 8, 9, 14 a bis e, `SPEC-139` Nr. 11, `SPEC-136` Nr. 8, `SPEC-201` Nr. 3 e, die Tabelle *Zusage · Fänger* (A01 bis A16) und die Hard Rules; nicht Diff-Review (siehe `2026-11-08-slice-bootstrap-m3-review.md`) und nicht Validierung.

**Gegenstand:** Zweig `slice-bootstrap-m3`, HEAD `5a6f0b5e` (Fix-Runde `a4804086` nach dem Review, Architect `affb15a2`), Basis des Blocks `ada8bd88`, Worktree `/Development/odata-recorder-wt-bootstrap-m3`. Plan: `docs/plan/planning/done/slice-bootstrap-m3.md`. Bezug: `LH-FA-12`, `LH-FA-06`, `LH-FA-11`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden am Zweig, mit Auflagen A1 bis A5 an den Integrations-Planner (A1 und A2 vor dem Merge, A3 ist der einzige Textkonflikt).** `make gates` ist im eigenen Lauf auf HEAD grün, die Marke hält (1337 von 2400), die Spezifikation stimmt mit dem Code überein (kein Abweichungsfund gegen 2,2 Mio. Zeichenfolgen), und die Fix-Runde deckt die Klassen von M1 bis M3 und A05 bis auf eine Nachbarklasse (A4). **Aber:** der Zweig ist hinter `main` (bekannte Merge-Kopplung mit `slice-handbuch-katalog-schluessel-kommandos`), und im Wegwerf-Merge ist `make katalog-check` **rot**: `Schlüssel: PUBLIC_BASE_URL fehlt im Handbuch`, `Schlüssel: STRICT_COVERAGE fehlt im Handbuch`. Die Kürzung der DoD (Zeilen der Tabelle *Einstellungen* entfallen) war gegen ein Handbuch ohne diese Tabelle gerechnet und gilt auf `main` nicht mehr (A1).

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, `DOCKER_BUILD` mit `--memory=4g`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/gates.out`: `rc=0`; `d-check: 880 Datei(en) geprüft, 0 Befund(e)`; Stufen `pruefung test`, `pruefung race`, `pruefung build`, `pruefung handbuch`, Katalog-Stufe, `zeichen: ok, 1782 Datei(en)`. |
| `make groessenmarke` | im selben Lauf: `slice-bootstrap-m3.md: ok, Messung 1337 von Marke 2400` (`Messpunkt: zeilen=16 nebenlaeufig=1 fakes=1 produktcode=260 schaetzung=1745`). Eigene Nachrechnung: `git diff --numstat ada8bd88..HEAD` über die vier Pfade des Blocks = 1273 Einfügungen + 64 Löschungen = 1337. |
| Plan-vs-Code-Diff | 20 Dateien seit `ada8bd88`. `src/bootstrap/` (Produktcode: `konfig_basis.go` neu, `konfig.go`, `konfig_dienste.go`, `konfig_werte.go`, `wurzel.go`; Tests), `tests/integration/m3_test.go` neu, `tests/e2e/stufe_prozess_test.go` (allein der Fall `leer`), `harness/katalog-ausnahmen.txt` (-8), `docs/user/benutzerhandbuch.md` (+8 Zeilen *Fehlercodes*), `SPEC-201` Nr. 3 e, `spec/spezifikation.md` (Nr. 14 e und Historie), Plan, Review, Plan des Nehmers. `Makefile`, `Dockerfile`, `harness/mk/`, `.golangci.yml`, `.a-check.yml`, `tools/` unberührt: stimmt mit §3 und §4 (Gate-Kette Linie B). Keine Abweichung vom Plan. |
| §3.2 / §3.18 | `nolint` im Diff: 0; `LC_ALL=C grep -nP` auf BOM, NBSP, Nullbreite, Steuerzeichen über alle geänderten Dateien: 0 Treffer; Folge Backslash-`u`-Hex in `m3_konfig_test.go`, `m3_test.go`: 0. |
| Merge-Nähe | `git merge-tree --write-tree --name-only main HEAD`: **ein** Konflikt, `docs/plan/planning/open/slice-bootstrap-m3-e2e-mock.md` (beide Seiten haben die Pfade ohne `./` verschieden berichtigt: `main` ohne Schrägstrich am Ende, der Zweig mit). `docs/user/benutzerhandbuch.md`, `SPEC-201`, `spec/spezifikation.md` mergen ohne Konflikt. Branch-Basis `2cefe4f2`, `main` `43d67170`. |
| Wegwerf-Merge (Worktree im Scratchpad, entfernt) | `make katalog-check`: **rot** (siehe A1); `make handbuch-check`: `pruefung handbuch: ok`. |

## 2. DoD (§2 des Plans)

Alle Haken stehen offen (`[ ]`); das setzt der Planner mit der Closure, die Zustände unten sind meine Bewertung.

| Punkt | Stand | Beleg |
|---|---|---|
| Verdrahtung und Konfiguration (Modus `mock`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`, Filter-Leser, `Schreiber`, `Eingabe`, Fehlerantwort-Negativ, `SnapshotUseCase` mit Wächter) | **erfüllt** | `wurzel.go` `verdrahteMock`, `satzVersion4`, `ablageRepos`; Fänger A07 bis A11, eigene Mutationen §4. |
| M3-Test in `tests/integration/` | **erfüllt** | `TestM3` (Teile `Mock`, `Coverage`, `Adressen`, `Neustart`); unmutiert ohne Detektor 3 von 3 grün, mit Detektor (`go test -race -c`, ein Prozess je Lauf) 3 von 3 grün; läuft in `make test` und `make test-race`. Negativ A13 (Zähler des Referenzdienstes) trägt Grund und Folge im Plan und im Testkommentar. |
| Nicht gelieferte Schlüssel bleiben `ODR-E2001`; Nachzug von `SPEC-110` | **erfüllt** | `nichtGelieferte()` nennt die vier Präfixe; Code gegen Spec §3. Tippfehler der Zeile: „`LISTEN_TLS_*`), ;“ in §2. |
| `make gates` grün | **erfüllt am Zweig** | §1. Auf `main` nach dem Merge **nicht** (A1). |
| Review durchgeführt, Report liegt vor | erfüllt | `2026-11-08-slice-bootstrap-m3-review.md`, HIGH 0, MEDIUM 3 (alle in `a4804086` bearbeitet). |
| Doku-Update (öffentlicher Vertrag) | erfüllt | `SPEC-110` Nr. 14 e, `SPEC-201` Nr. 3 e, Historie. |
| Handbuch nachgezogen | **teilweise, Kürzung korrekt, Nehmer fehlt, Folge für `main` unvollständig** | acht Zeilen *Fehlercodes* stehen (`ODR-E6001` bis `E6005`, `E6007` bis `E6009`), acht Zeilen in `harness/katalog-ausnahmen.txt` entfernt; die DoD-Zeile, §1 und A16 sagen „Kapitel, Beispiel und Zeilen *Einstellungen* entfallen“. Der Nehmer `slice-handbuch-mock` existiert in keinem Lifecycle-Verzeichnis und in keiner Welle (A2); die Zeilen der Tabelle *Einstellungen* sind auf `main` Gate-Pflicht (A1). |
| Closure-Notiz, Register, Paarungen, Ausgang je Risiko | **offen, Sache des Planners** | §7 *Wird bei der Closure gefüllt*. Ausgänge *offen* oder *weiter offen* in §6 (A5). |

Liefer-Punkte: drei (Verdrahtung, M3-Test, Nicht-Geliefertes), am Limit.

## 3. Spec-Konformität (Code gegen Wortlaut)

- **`SPEC-110` Nr. 2:** 29 Schlüssel in der Zeile *Geprüft werden*, `STRICT_COVERAGE` und `PUBLIC_BASE_URL` darunter; `nichtGelieferte()` ohne die beiden Namen; `ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*` bleiben `ODR-E2001` (`TestKonfigNichtGeliefert`, 12 Namen × 3 Modi).
- **Nr. 8:** Default `mock`, `hybrid` mit dem Text `Modus nicht unterstützt; lieferbar sind record, replay und mock`; ein ungültiger `MODE` lässt den Modus leer (Pflichten am gültigen Modus entfallen).
- **Nr. 9:** `UPSTREAM_URL` im Modus `mock` nicht erforderlich, Form in jedem Modus geprüft (A02).
- **Nr. 14 a:** `STRICT_COVERAGE` ungesetzt oder leer ist `true`, Konfiguration trägt den Wert (A03).
- **Nr. 14 b:** Form von `PUBLIC_BASE_URL`. **Probe über die Klasse (zusätzlich zu den Tests):** `istBasisURL` gegen `jsoncodec.NewSchreiber().Adresse` als Orakel über vier erschöpfend erzeugte Mengen: Segment (402234 Zeichenfolgen über `a . % 2 e E / ~ Leerzeichen 0 F g -`, Länge bis 5), Host (69905), Port (597871), Schema (1111111): **0 Abweichungen**. Der Code ist also nicht nur gegen die Fixtures, sondern gegen das Orakel auf 2,18 Mio. Elementen gleich.
- **Nr. 14 c, d, e:** Pflicht je Modus und Ableitung aus `LISTEN` (Port ohne führende Nullen), ein Grund je Schlüssel (Form schließt Länge aus), Länge der Wurzel mit abgeleiteter Basis (`laengsteWurzel(w, k.Dienste)` nach der Ableitung, Grenze `>` 2048).
- **`SPEC-139` Nr. 1, 5, 11 a bis g:** Zugänge `Sessions()`, `Record()`, `Import()` sind im Modus `mock` der Nullwert der Schnittstelle; sieben Konstruktoren null Aufrufe; Satz der Version 4 mit allen fünf Typfeldern, `NurLesend` falsch; ein Wächter, eine Ablage; `Snapshots()` in jedem Modus; kein `CloseAll`, keine Session; `Coverage` Typ statt Identität als akzeptiertes Negativ (zustandsloses Repository, Grund trägt).
- **`SPEC-136` Nr. 8:** Entscheidung (Fehlerantwort ohne Kennzeichen der Version, akzeptiertes Negativ mit Grund und Folge) steht im Wortlaut der Nr. 8, Spiegel in `SPEC-139` Nr. 11 e; der Slice ruft `SchreibeFehler` nicht.
- **`SPEC-201` Nr. 3 e:** zählt die acht Codes des Mock in H, die Ausnahmen sind entsprechend gekürzt; stimmt mit `harness/katalog-ausnahmen.txt` überein.

## 4. Tabelle *Zusage · Fänger* A01 bis A16 gegen den Vertrag

Jede Zeile nennt einen Fänger, dessen Symbol im Quelltext steht (`grep -rn "^func Test…"` für alle 17 genannten Tests: alle vorhanden). A01 bis A12 sind Schleifen über die Mengen der Zusage (Modus × Wert, 256 Bytes × 5 Stellen, 65536 Paare, 9 Hosts × 3 Ports × 5 Modi, 12 Namen × 3 Modi); die Fakes halten Argumente fest und zählen Aufrufe. A13 ist `nebenläufig` und trägt ein Negativ mit Grund und Folge, A15 und A16 je ein Negativ mit Grund und Folge. Raster §6: jede Zelle nennt eine Nummer oder „gilt nicht“ mit Grund. **Plan-Abweichung:** A09 sagt „fünf Typfelder“, Test zählt fünf (Befund LOW des Reviews, behoben). Kein Eintrag ohne Fänger gefunden **außer** der Nachbarklasse in A4.

## 5. Fix-Runde gegen die Klasse aufgelöst (§3.17)

Messender: Verifier-Agent. Container aus dem Image `odata-recorder:race` (vorhandenes Image, nicht neu gebaut), ohne Netz, `--memory=3g --cpus=4`, `flock /tmp/odata-recorder-heavy.lock`, Quellen des Zweigs schreibgeschützt eingehängt (`:ro`), je Lauf Wegwerfkopie im `tmpfs`, `sed` genau einmal je Mutation, Datei danach zurückgesetzt, `go test -count=1 -run <Test> ./src/bootstrap` (ein Prozess). Host-Last zu Beginn 9,5 bis 27 (Mitläufer), alle Fänger deterministisch; Treiber: `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/mut.sh`, `mut4.sh` bis `mut6.sh`. Mutationen, die nur mit Übersetzungsfehler rot wurden (P9, L2, L6, S2, S4 im ersten Lauf), sind verworfen und mit kompilierbarer Form wiederholt (`mut4.sh`).

| Befund | Menge (Wortlaut der Zusage) | Elemente | Mutation an einem Element, das der Befund nicht nannte | Ergebnis |
|---|---|---|---|---|
| M1 (Prozent-Kodierung) | beide Stellen von `%XY`, alle Bytewerte | 65536 (482 angenommen) | P1 Hex bis `g`; P2 Hex ab `/`; P3 `a` nicht Hex; P7 `E`/`F` nicht Hex; P4 erste Ziffer ungeprüft; P5 Schritt 1 statt 2; P6 `ParseUint` versetzt; P10 nur `%2e` klein abgelehnt | rot `TestKonfigPublicBaseURLProzent` (und Form/Zeichen); alle 8 aus dem benannten Grund |
| M1, Nachbarklasse | Segment, dessen dekodierter Text `.` oder `..` ist (Wortlaut `SPEC-168` Nr. 2: „auch nicht als `%2E`“) | 12 Schreibweisen (`.`, `%2e`, `%2E`; `..` als 3 × 3), Fixtures führen 4 | P9 Vergleich auf dem rohen Segment mit fester Liste statt auf dem dekodierten Text | **grün, Mutation überlebt** (A4) |
| M2 (Meldung nennt nie den Wert) | jede Meldung der neuen Schlüssel: Form, Pflicht, Länge, `STRICT_COVERAGE`, Nicht-geliefert | 5 | N1 Länge plus Wert; N2 Pflicht plus `LISTEN`; N3 Nicht-geliefert plus Wert (alle nicht genannt) | rot: N1 `…AbleitungLaenge`, `…Laenge`; N2 `…Ableitung`; N3 `TestKonfigMeldung` (N1, N2 über Gleichheit der ganzen Zeile, nicht über den Geheimwert) |
| M3 (A12/A07, Sessions) | Zähler Sessions je Service der Ablage = 0; Panik als benannter Fehler; sieben verbotene Konstruktoren | 2 Services; 7 | Z1 Sessions je Service direkt über die Ablage angelegt (ohne Session-Dienst); S5 typisierter Nullzeiger in `sessions`; S6 `NeuerImport` im Modus `mock`; S4 `verarbeiten = h.record.Record`; S3 Startwartung aus; S2 `Schliessen(0)`; X9 Version `2.0` | alle rot aus dem benannten Grund; Z1 rot als `Service a: 1 Sessions in der Ablage, <nil>; erwartet 0` (Zähler). Überlebt, aber gleichwertig: S1/S8/S7 (`CloseAll` nur mit `begonnen`, im Modus `mock` nie gesetzt; Panik-Pfad nicht erreichbar) |
| A05 (Länge bei abgeleiteter Basis) | 3 Hosts × Port × Root auf genau 2048 und ein Byte mehr | 12 | L1 längster Root nur `len(Root)`; L2 Port `TrimPrefix` einer Null; L3 Länge nur beim gesetzten Wert; L4 Root `/` zählt; L5 `LOCALHOST` angenommen; L6 Basis ohne Port | rot (`…AbleitungLaenge`, `…Ableitung`, `…Laenge`) |
| Verdrahtung A08 bis A11 | Felder der Konfig des Mock und des Snapshot-Dienstes | je Dienst 5 + 4 + 6 | X1 leere `PublicBaseURL` ohne Prüfung; X2 `Transaktion` nil; X3 `Coverage` nil; X4 `Entities` des Snapshot-Dienstes eigenes `db.Entities()`; X5 eigene Uhr; X7 eigene `Workspaces`; X8 `Snapshots()` im Modus `mock` nil | alle rot aus dem benannten Grund. X6 (eigener ID-Generator): überlebt, **gleichwertig** (`ZufallsIDs` ist `struct{}` ohne Zustand, Zeiger vergleichen gleich) |
| M3-Test A13 bis A15 | Felder des Mock am Listener | 6 | I1 `StrictCoverage` negiert; I2 `PublicBaseURL` fest; I3 Filter-Leser ohne; I4 `MaxBytes` 1; I5 Root fest; I6 `Eingabe` nil | rot, **je 3 von 3** Prozessläufen (`go test -c` ohne Detektor, ein Prozess, `-test.count=1`), Teile `Coverage`, `Adressen`, `Mock`; Wirkung deterministisch |
| P8 | Segment, das zu `.%` dekodiert (`.%25`), abgelehnt | 1 | P8 | grün; liegt außerhalb jeder Zusage des Vertrags (kein Fixture verlangt die Annahme), als INFO geführt |

Zählung eigene Mutationen: 38 verschiedene (davon 5 verworfen und wiederholt, 4 überlebend gleichwertig oder außerhalb, 1 überlebend in einer Klasse des Vertrags).

## 6. Befunde und Auflagen

### V1 · MEDIUM — Katalog-Gate wird nach dem Merge von `main` rot; die Kürzung der DoD ist gegen ein Handbuch ohne Tabelle gerechnet (A1)

- `quelle`: `SPEC-201` Nr. 4 b, c (K aus `SPEC-110` Nr. 2 minus N; jeder Name aus K, der in C steht, muss in W stehen), `AGENTS.md` §3.9 (der Plan folgt jeder Korrektur), `welle-mock` §4 *Pflicht Handbuch*.
- `pfad`: `docs/user/benutzerhandbuch.md` (*Einstellungen* auf `main`: keine Zeile `STRICT_COVERAGE`, keine `PUBLIC_BASE_URL`; die Zeile `MODE` sagt „Der Start nimmt nur diese beiden Werte an; ohne Angabe gilt `mock`, und der Start lehnt ihn ab“); Plan §2 *Handbuch nachgezogen*, §1 *Ausdrücklich NICHT*, §3 A16 Negativ 2.
- `befund`: Wegwerf-Merge `main` in `5a6f0b5e`: `katalog: Schlüssel: PUBLIC_BASE_URL fehlt im Handbuch`, `katalog: Schlüssel: STRICT_COVERAGE fehlt im Handbuch`, `make katalog-check` Exit 2. Der Zweig hat die Tabelle *Einstellungen* nie gesehen (sie kam mit dem Merge-Partner), die Aussage „die Zeilen entfallen“ (Architect, 2026-11-08) trägt auf `main` nicht. Die Zeile `MODE` ist nach dem Merge falsch (`mock` startet jetzt); kein Gate fängt Prosa. `SERVICE_ODATA_VERSION` bleibt wie dort entschieden ohne Zeile; die Festzahlen „27, 26, 6“ stehen nur in synthetischen Fixtures von `tools/katalog` und `tools/gegenproben` und in Zeitdokumenten (`done/`), nicht in einem lebenden Vertrag: nichts nachzuziehen.
- **Auflage A1 an den Integrations-Planner (im Merge-Commit, ein Commit, vor `done/`):** in der Tabelle *Einstellungen* zwei Zeilen ergänzen (`STRICT_COVERAGE`: Wahrheitswert, ohne Angabe `true`, gilt im Modus `mock` für Abfragen auf nicht vollständigem Bestand; `PUBLIC_BASE_URL`: Adresse, die der Client sieht, Pflicht im Modus `mock` bei anderem Host als Loopback, sonst aus `LISTEN` abgeleitet, ohne abschließenden `/`) und die Zeile `MODE` auf den Ist-Zustand ziehen (`mock` lieferbar und Default, `hybrid` wird abgelehnt); danach `make katalog-check` und `make handbuch-check` auf dem Merge-Stand. Im Plan §1, §2, §3 (A16 Negativ 2) und §6 die Kürzung auf Kapitel und Beispiel einschränken (die Zeilen sind geliefert), in `slice-handbuch-mock` nur noch Kapitel und Beispiel.

### V2 · MEDIUM — der Nehmer `slice-handbuch-mock` existiert nicht; der Geber nennt eine Adresse ohne Eintrag (A2)

- `quelle`: `AGENTS.md` §3.13 (Adresse nimmt an, bevor sie genannt wird), `welle-mock` §3 Schluss und §4.
- `pfad`: `find docs/plan/planning -name '*handbuch-mock*'`: kein Treffer in `open/`, `next/`, `in-progress/`, `done/`; `grep -rn slice-handbuch-mock` nennt nur den Plan dieses Slice (§1, §2, §3 A16, §6); `welle-mock.md` §4 *Pflicht Handbuch* (Zeile 84) und das Kriterium „Das Handbuch deckt alles ab, was das Binary dieser Welle kann (Modus `mock`, Schlüssel, Fehlercodes)“ (Zeile 56) nennen den Slice nicht.
- `befund`: Die DoD-Zeile ist korrekt gekürzt (acht Zeilen *Fehlercodes* geliefert, Kapitel und Beispiel mit Adresse Planner abgegeben), §6 trägt die Zeile mit Adresse Planner. Der Nehmer-Eintrag liegt **nicht** vor.
- **Auflage A2 an den Planner (vor `done/` von `slice-bootstrap-m3`):** `slice-handbuch-mock` in `open/` anlegen (Welle `welle-mock`, Trigger „`slice-bootstrap-m3` in `done/`“), mit der Sendung in §1 und DoD (Kapitel zum Modus `mock`, Beispiel als Datei, Erweiterung von `SPEC-200` Nr. 6 d und 7 b um Umgebung und einen Lauf mit Signal als Eingang für den Architect vor dem Code), `welle-mock` §3/§4 und Roadmap nachziehen (die Welle schließt nicht ohne das Kapitel), Plan §6 der Zeile *Ausgang: entschieden, Adresse Planner* nach Anlage auf *entschieden* mit Verweis auf die angelegte Datei setzen.

### V3 · MEDIUM — Nachbarklasse der Prozent-Kodierung bleibt ungefangen (A4)

- `quelle`: `AGENTS.md` §3.15, §3.17; Wortlaut `SPEC-168` Nr. 2 („sein dekodierter Text ist nicht `.` und nicht `..` (auch nicht als `%2E`)“) und `SPEC-110` Nr. 14 b (Form des Schreibers).
- `pfad`: `src/bootstrap/m3_konfig_test.go`, `TestKonfigPublicBaseURLForm` (ungültige Formen: `http://h/.`, `http://h/..`, `http://h/%2E`, `http://h/%2e%2e`) und `TestKonfigPublicBaseURLProzent` (nur einzelnes `%XY`).
- `befund`: Die Klasse „Segment mit dekodiertem Text `.` oder `..`“ hat 12 Schreibweisen (`.`, `%2e`, `%2E` einzeln; `..` als 3 × 3); die Fixtures führen vier. Mutation P9 (Vergleich auf dem rohen Segment gegen eine feste Liste) bleibt in allen `TestKonfigPublicBaseURL*` grün; `.%2e`, `%2e.`, `%2E%2E`, `%2e%2E`, `%2E%2e`, `.%2E`, `%2E.` sind ungefangen. Der Code ist richtig (die Probe in §3 findet 0 Abweichungen vom Orakel); es fehlt der Fänger.
- **Auflage A4 an den Implementer** (kleiner Test, ein Commit, vor `done/`; im Integrationsschritt zulässig, weil er nichts am Produktcode ändert): `TestKonfigPublicBaseURLForm` oder ein eigener Test als Schleife über die 12 Schreibweisen gegen den Schreiber als Orakel; Mutation P9 muss rot werden; der Bericht nennt Menge, Zahl und eine Mutation an einem nicht genannten Element.

### V4 · LOW — Plan trägt die Ausgänge noch nicht (A5)

- `befund`: Zählt man die Zeilen in §6: *offen* oder *weiter offen* sind (a) Zahl der Codes und Merge-Nähe zu `slice-admin-authentisierung` (kein Code neu, die Berührungspunkte dieses Zweigs mit `main` sind nur die drei oben, Konflikt nur im Plan des Nehmers; `slice-admin-authentisierung` läuft weiter, dessen Merge hat die Zahl in den Zählstellen anzugleichen), (b) Kopplung an `slice-handbuch-katalog-schluessel-kommandos` (jetzt gemerged: Ausgang eingetreten, Folge V1), (c) M3-Test unter dem Detektor (eingetreten ohne Befund: 3 von 3 und `make test-race` im Gate grün), (d) `gegenprobe-haengt-an-der-laufzeit-des-produkts`, (e) Mutationslauf an `src/bootstrap`, (f) `urlparser`-Flackern. Dazu: DoD-Haken, §7 Closure-Notiz, Register, der Tippfehler „`LISTEN_TLS_*`), ;“ in §2.
- **Auflage A5 an den Planner (Closure):** Ausgänge vergeben, §7 füllen, Haken setzen, Register fortschreiben; der Hinweis zu (a): nach dem Merge `model.AlleCodes()` zählen (der Zweig fügt keinen Code hinzu).

### V5 · INFO

- A13-Negativ (Zähler des Referenzdienstes kann nicht rot werden) ist ehrlich geführt; die Zusage „kein Upstream im Modus `mock`“ trägt allein `verboten` in `TestM3WahlJeModus` (sieben Konstruktoren). Mutation S6 (`NeuerImport` im Modus `mock`) und P-/X-Mutationen zeigen, dass der Fänger wirkt.
- Gleichwertige Überlebende (S1, S7, S8, X6) tragen keine Zusage; sie zeigen nur, dass `CloseAll` doppelt abgesichert ist (Modus und `begonnen`) und `ZufallsIDs` zustandslos ist.
- Das Mutationslauf-Werkzeug (`mutationslauf.sh`) scheitert laut Plan an der Nullprobe von `src/bootstrap`; ich habe es nicht gefahren, die Mutationen laufen über eigene Treiber mit `-run`. Raten: alle Fänger deterministisch, kein Detektor-Fänger nötig; A13 ist als `nebenläufig` geführt, die vier Mutationen des Plans (Restore und Reset ohne Wirkung) habe ich nicht neu gefahren (Mutationen im Dienst, nicht im Diff), stattdessen sechs Wurzel-Mutationen I1 bis I6 mit je 3 von 3.

## 7. Auflagen (Zusammenfassung für den Integrations-Planner)

| Nr. | An | Was | Wann |
|---|---|---|---|
| A1 | Integrations-Planner | Zwei Zeilen `STRICT_COVERAGE`, `PUBLIC_BASE_URL` in *Einstellungen* ergänzen, Zeile `MODE` auf den Ist-Zustand; `make katalog-check` und `make handbuch-check` auf dem Merge-Stand; Plan (§1, §2, §3 A16, §6) auf Kapitel und Beispiel einschränken | im Merge-Commit |
| A2 | Planner | `slice-handbuch-mock` in `open/` anlegen (Sendung Kapitel, Beispiel, `SPEC-200`-Erweiterung vor dem Code), `welle-mock` und Roadmap nachziehen | vor `done/` des Gebers |
| A3 | Integrations-Planner | Konflikt in `docs/plan/planning/open/slice-bootstrap-m3-e2e-mock.md` lösen: die Fassung von `main` übernehmen (Pfade `tests/e2e` und `tests/e2e/uebergang` ohne Schrägstrich am Ende), die Fassung des Zweigs verwerfen; `make docs-check` danach | im Merge-Commit |
| A4 | Implementer | Fänger über die 12 Schreibweisen des dekodierten Punkt-Segments (Mutation P9) | vor `done/` |
| A5 | Planner | Ausgänge in §6, §7, Haken, Register; DoD-Tippfehler; nach dem Merge `model.AlleCodes()` zählen | Closure |

## 8. Gegenlesen (`AGENTS.md` §3.18)

Dieser Bericht ist nach dem Schreiben mit `LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}|\xEF\xBB\xBF|\xC2\xA0'` gegengelesen (Datei `docs/reviews/2026-11-08-slice-bootstrap-m3-verifikation.md`); Beispiele mit Umlauten sind UTF-8, keine Escape-Folgen.
