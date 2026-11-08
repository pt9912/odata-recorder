# Slice slice-bootstrap-m3: Composition Root für den Modus mock und der M3-Test

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-12`, `LH-FA-06`, `LH-FA-07`, `LH-FA-08`, `LH-FA-09`, `LH-FA-10`, `LH-FA-11`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `SPEC-110` Nr. 2, 8, 9 und 14 (Schlüssel dieser Version, Modus, kein Upstream, `STRICT_COVERAGE` und `PUBLIC_BASE_URL`) · `SPEC-139` Nr. 1 bis 5, 7, 9 und 11 (Verdrahtung des Mock, des Satzes der Version 4 und des Snapshot-Dienstes) · `SPEC-136` Nr. 8 (Fehlerantwort ohne Kennzeichen der Version) · `SPEC-037` Nr. 5 bis 7 und 10 (Vertrag der Szenarien, `harness/vertraege/`) · `SPEC-011`, `SPEC-055`, `SPEC-075` · `SPEC-052` (Adressierung) · `SPEC-160` Nr. 8 (Einhängestelle `$filter`) · `ARC-005`, `ARC-012` · `architecture.md` §2.5 (Modus-Routing).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-bootstrap-m3`, Worktree `/Development/odata-recorder-wt-bootstrap-m3`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Composition Root bindet im Modus `mock` genau `MockUseCase` an den OData-Listener, der Start verlangt keinen Upstream, die Startkonfiguration liefert `MODE=mock` (auch als Default), `STRICT_COVERAGE` und `PUBLIC_BASE_URL`, und der M3-Test aus §1 der Welle läuft grün in `make test`.

**Übernimmt:** den Nachzug von `SPEC-110` Nr. 2 und 8, den `slice-bootstrap-m2` für die Modi `record` und `replay` hinterlassen hat (der Modus `mock` ist dort „nicht unterstützt“).

**Übernimmt (Sendung von `slice-odata-url-filter`):** das Einhängen des Filter-Lesers im Composition Root, `urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser()))` (`SPEC-160` Nr. 8); der Geber liefert den Leser, die Verdrahtung liegt hier. Kein Ausschluss dieses Slice trifft sie.

**Übernimmt (Sendung von `slice-v2-bootstrap`, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted`):** der Root baut den Satz der Version 4 (`outbound.CodecSatz` mit `csdl`, `jsoncodec`, `urlparser`, dem Schreiber der Version 4 und `NurLesend` falsch, `SPEC-175` Nr. 3) und gibt ihn dem Mock; die Tabelle Version → Satz wird erst in `slice-v2-bootstrap` vollständig. Kein Ausschluss dieses Slice trifft sie (Verdrahtung, keine neue Fachlogik).

**Übernimmt (Sendung von `slice-odata-json-schreiben`, `SPEC-168`, `SPEC-175` Nr. 3):** das Feld `Schreiber` des Satzes der Version 4 füllt der Root mit `jsoncodec.NewSchreiber()`; der Geber liefert den Schreiber (ein Wert ohne Zustand, von allen Anfragen geteilt, `SPEC-168` Nr. 9), die Verdrahtung liegt hier; Fänger: der Satz trägt einen Schreiber der Version 4 und keinen `nil`. Dazu reicht der Root `PUBLIC_BASE_URL` an die Konfiguration des `MockUseCase`, aus der der Dienst `Wurzel` bildet (Konfig: `slice-mock-zuordnung`; `Wurzel`: `slice-mock-lesen`, `SPEC-168` Nr. 10). Kein Ausschluss dieses Slice trifft sie (Verdrahtung, keine neue Fachlogik).

**Übernimmt (Rückgabe R3 des Architect aus `slice-mock-lesen`, `SPEC-195` Nr. 10, `SPEC-136` Nr. 7 und 8, `SPEC-168` Nr. 7, `SPEC-172` Nr. 6; Geber: `slice-mock-zuordnung` und `slice-mock-lesen`; Eintrag beim Schnitt, Planner, 2026-11-07):** der Dienst `MockUseCase` liefert einen Fehlerwert, das Fehlerobjekt schreibt der Adapter ohne `OData-Version` und ohne `DataServiceVersion`, auch für einen V2-Service; `SchreibeFehler` des Schreibers im Satz hat damit keinen Aufrufer. Der Architect entscheidet vor dem Code (Auftrag beim Start, `AGENTS.md` §3.12), ob die Version in der Fehlerantwort des Mock ankommt und auf welchem Weg (ein additives Feld am Fehlerwert, ein Aufruf des Schreibers im Composition Root oder im Adapter) oder ob die Lücke als akzeptiertes Negativ mit Grund und Folge in `SPEC-136` und `SPEC-195` steht; der Slice setzt die Entscheidung um (Verdrahtung des vorhandenen `SchreibeFehler`, keine neue Fachlogik des Dienstes) oder trägt das Negativ. Bedarf an einer Änderung am Fehlerwert `*inbound.Fehler`, die `slice-mock-zuordnung` geliefert hat, geht an den Planner. Kein Ausschluss dieses Slice trifft den Punkt.
**Entscheidung des Architect (2026-11-08, vor dem Code):** die Lücke ist ein **akzeptiertes Negativ** mit Grund und Folge in `SPEC-136` Nr. 8 (Spiegel: `SPEC-195` Nr. 10, `SPEC-168` Nr. 7): die Fehlerantwort des Mock trägt kein Kennzeichen der Version, der Slice setzt nichts um und ruft `SchreibeFehler` nicht. Der Grund in einem Satz: der Root legt bis zur Wahl nach der Version nur Services der Version 4 an, der Adapter kennt die Version nicht (`SPEC-175` Nr. 7), und kein Kriterium des Lastenhefts verlangt sie an der Fehlerantwort; das Kennzeichen später zu ergänzen ist additiv. Eine Änderung am Fehlerwert `*inbound.Fehler` entfällt, es geht nichts an den Planner.

**Übernimmt (Sendung von `slice-odata-json-eingabe`, `SPEC-180`, `SPEC-175` Nr. 3, `SPEC-109`):** das Feld `Eingabe` des Satzes der Version 4 füllt der Root mit `jsoncodec.NewEingabe()`; der Geber liefert den Leser (ein Wert ohne Zustand, von allen Anfragen geteilt, `SPEC-180` Nr. 5), die Verdrahtung liegt hier; Fänger: der Satz trägt einen Leser der Version 4 und keinen `nil`. Dazu gehört der Handbuch-Eintrag des Fehlercodes `ODR-E6007` (Request-Body des Mock nicht lesbar oder verstößt gegen das Modell, `400`): er erreicht einen Client erst mit dem Modus `mock`, die Handbuch-Pflicht der Welle (§4) gilt hier verbindlich. Kein Ausschluss dieses Slice trifft sie (Verdrahtung, keine neue Fachlogik).

**Übernimmt (Sendung von `slice-snapshot-use-case`, `SPEC-123` Nr. 9, Rückgabe V-3 der Verifikation, Planner, 2026-11-07):** die Verdrahtung des Snapshot-Dienstes im Composition Root: der Wächter der Ablage wird als `Blobs` an `snapshot.New` gereicht, Workspaces, Entities, Snapshots und Transaktion liegen in **derselben** Datenbank, und der `SnapshotUseCase` wird im Root gebunden; jeder Schreiber des Blobs läuft über denselben Wächter, auch der des Mock. Der Geber liefert Dienst, Vertrag und den Fänger mit Wächter (T59 dort); sein Plan führt die Verdrahtung als akzeptiertes Negativ (T81), das hier getragen wird. Kein Ausschluss dieses Slice trifft sie (Verdrahtung, keine neue Fachlogik; die Fachlogik-Zeile unten nennt `slice-snapshot-use-case` als Gegenstand der Verdrahtung). Die Weitergabe des Dienstes an den Admin-Adapter ist nicht hier: `slice-bootstrap-m6` verdrahtet den Admin-Adapter mit den Use Cases, `slice-admin-workspaces` liefert nur die Endpunkte.

**Übernimmt (Sendung von `slice-mock-abnahme`, `SPEC-049` Nr. 7 a und 10 b, Rückgabe R4, Planner, 2026-11-07):** den Neustart über die Prozessgrenze im Modus `mock` für **Mutationen**: das gebaute Binary im Modus `mock` mit einem `DATA_DIR`, eine Mutation (`POST`) über den Listener, `SIGTERM` (`LH-FA-02.c`), Start auf demselben `DATA_DIR`, die Mutation ist über den Listener lesbar. Der Geber prüft den Neustart im Prozess des Tests (Dienste und Ablage); die Prozessgrenze fährt dieser Slice als Stufe von `make test-integration` (Szenario neben *Mock CRUD → Reset*; Entscheidung des Architect, 2026-11-08: Stufe `TestE2EMockNeustart`, Randformen in `SPEC-139` Nr. 11 d und f und `SPEC-037` Nr. 7 und 10 i; Teil E des Schnitts, §4). **Snapshots über die Prozessgrenze sind nicht Teil der Sendung:** es gibt bis `welle-admin-api` keinen Lesepfad dafür (`slice-container-image` §1); sie bleiben das akzeptierte Negativ aus `SPEC-049` Nr. 10 b. Kein Ausschluss dieses Slice trifft die Sendung (Verdrahtung und Szenario, keine neue Fachlogik).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Verdrahtung des Modus `hybrid` und der Schlüssel `HYBRID_*` und `FALLBACK_*` — `slice-bootstrap-m4`; sie bleiben bis dahin `ODR-E2001`, damit keine Sicherheitseinstellung still wirkungslos ist (`SPEC-110` Nr. 2).
- TLS des Listeners (`LISTEN_TLS_*`) und die Admin-Schlüssel — `welle-betrieb` und `welle-admin-api`.
- Neue Fachlogik — der Slice verdrahtet, was `slice-mock-zuordnung`, `slice-mock-lesen`, `slice-mock-schreiben` und `slice-snapshot-use-case` liefern.

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

- [ ] Verdrahtung und Konfiguration: Modus `mock` bindet `MockUseCase`, der Start verlangt keinen Upstream, `STRICT_COVERAGE` und `PUBLIC_BASE_URL` (Pflicht bei nicht-Loopback-Host) kommen aus dem Wert des Starts; Fänger, dass der Root sie weiterreicht und den Nullwert nicht stehen lässt (Muster aus `slice-upstream-tls`). Dazu hängt der Root den Filter-Leser ein, `urlparser.New(urlparser.WithFilter(urlparser.NewFilterLeser()))` (Sendung von `slice-odata-url-filter`); Fänger: `$filter` im Mock ergibt keine Ursache `filter` (`501`) mehr, und `New()` ohne Leser bleibt im Root unerreichbar. Dazu füllt der Root `CodecSatz.Schreiber` mit `jsoncodec.NewSchreiber()` und reicht `PUBLIC_BASE_URL` an die Konfiguration des `MockUseCase` (Sendung von `slice-odata-json-schreiben`, `SPEC-168`); Fänger: der Satz hat keinen `nil`-Schreiber. Ebenso füllt der Root `CodecSatz.Eingabe` mit `jsoncodec.NewEingabe()` (Sendung von `slice-odata-json-eingabe`, `SPEC-180`); Fänger: der Satz hat keinen `nil`-Leser. Die Fehlerantwort des Mock trägt kein Kennzeichen der Version (Sendung von `slice-mock-zuordnung` und `slice-mock-lesen`, Rückgabe R3: akzeptiertes Negativ nach der Entscheidung des Architect, §1, `SPEC-136` Nr. 8); es ist nichts zu verdrahten, und die Zeile A15 der Tabelle in §3 trägt das Negativ. Dazu bindet der Root den `SnapshotUseCase` mit dem Wächter der Ablage als `Blobs` und einer gemeinsamen Datenbank für Workspaces, Entities, Snapshots und Transaktion (Sendung von `slice-snapshot-use-case`, `SPEC-123` Nr. 9); Fänger: ein Schreiber des Blobs am Root-Aufbau, der den Wächter nicht trägt, wird rot (Mutation: Wächter weggelassen oder zweite Datenbank), und ein Snapshot-Rundlauf über den Root-Aufbau gelingt.
- [ ] M3-Test aus §3 der Welle in `tests/integration/` (Ort: §6), Referenzdienst beendet, keine Verbindung zu ihm (`LH-FA-12`, Happy Path); dazu als Stufe von `make test-integration` am gebauten Binary das Szenario *Mock CRUD → Reset* und das Szenario *Mock: Mutation → `SIGTERM` → Start auf demselben `DATA_DIR` → Mutation lesbar* (Sendung von `slice-mock-abnahme`, `SPEC-049` Nr. 7 a; Fänger: die Mutation vor dem Signal fehlt nach dem Start, wenn der Start das Verzeichnis leert oder das Beenden die Transaktion nicht abschließt) (`SPEC-037`, Ziel aus [slice-test-integration](../done/slice-test-integration.md)) für die Kriterien, die die [Rückverfolgungs-Matrix](../rueckverfolgung.md) diesem Slice zuweist.
- [ ] Nicht gelieferte Schlüssel bleiben `ODR-E2001` (`HYBRID_*`, `FALLBACK_*`, `ADMIN_*`, `LISTEN_TLS_*`), ; der Nachzug von `SPEC-110` Nr. 2, 8, 9 und 14 steht in der Spezifikation (Architect, 2026-11-08), der Implementer prüft ihn gegen den Code.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand: Modus `mock`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`, die Fehlercodes des Mock einschließlich `ODR-E6007` (Sendung von `slice-odata-json-eingabe`), belegt durch ein Beispiel als Datei gegen das gebaute Binary; die Tabelle *Einstellungen* trägt `STRICT_COVERAGE` und `PUBLIC_BASE_URL`, sobald sie besteht, und die acht Zeilen des Mock (`ODR-E6001` bis `ODR-E6005`, `ODR-E6007` bis `ODR-E6009`) sind aus `harness/katalog-ausnahmen.txt` entfernt); das Kataloge-Gate ist grün. Voraussetzung ist `slice-handbuch-meldungskatalog` in `done/` (Trigger der Welle).
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
| `src/bootstrap/` | update | Wahl des Use Case je Modus, Start im Modus `mock` (Konfiguration, Verdrahtung, `Handle.Snapshots()`, neue Haken `NeuerMock` und `NeuerSnapshot`), Schlüssel `STRICT_COVERAGE` und `PUBLIC_BASE_URL` samt Ableitung, Einhängen von `urlparser.WithFilter(urlparser.NewFilterLeser())` (von `slice-odata-url-filter`), Bindung des `SnapshotUseCase` mit Wächter als `Blobs` und gemeinsamer Datenbank (von `slice-snapshot-use-case`) |
| `src/adapters/driving/httpodata/` | keine Änderung erwartet | `Verarbeiter` ist modusfrei, `Methode` und Body gehen unverändert an den Port (`SPEC-136`); der M3-Test belegt die Routen. Ein Mangel, den er findet, ist ein Fund an den Eigner (`AGENTS.md` §3.12), keine Erweiterung im Code |
| `tests/integration/m3_test.go` (Ort: Architect, §6) | neu | M3-Test mit dem Referenzdienst aus `slice-m2-referenzdienst` (gleiches Paket `integration_test`) |
| `tests/e2e/` (Teil E, §4) | neu | Stufen `TestE2EMock` und `TestE2EMockNeustart`, Befehl `reset` des Übergangs in `tests/e2e/uebergang/` mit Test; im Teil A allein der Fall `leer` von `TestE2EProzess` (`stufe_prozess_test.go`) |
| `tools/harness/test-integration.sh`, `docs/plan/planning/e2e-abdeckung.md` (Teil E) | update | drei Zeilen je Stufe (`SPEC-037` Nr. 3); das Abdeckungs-Dokument entsteht aus `make e2e-abdeckung`; `Makefile`, `Dockerfile`, `harness/mk/` unberührt (§4) |
| `harness/katalog-ausnahmen.txt`, `docs/user/` | update | acht Zeilen des Mock entfernen; Handbuch-Kapitel und Beispiel als Datei (DoD) |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-037-test-integration.md` | update | Nachzug von `SPEC-110`, `SPEC-139`, `SPEC-136`, `SPEC-049`, `SPEC-037` (Architect, 2026-11-08, erledigt vor dem ersten Code-Commit) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16); beide hat der Architect am 2026-11-08 vor dem ersten Code-Commit geschrieben, die Randformen sind in der Spezifikation entschieden (`AGENTS.md` §3.12). Die Zeilen **A01 bis A16** gehören zum Teil A, **E01 bis E09** zum Teil E des Schnitts (§4). Die Klasse `nebenläufig` tragen A13 (Referenzdienst, Recorder und Client im Prozess unter dem Detektor) und E04 (Beenden und Start des Servers); jede andere Zeile, auch ein akzeptiertes Negativ, ist `andere` (`SPEC-128` Nr. 3 e). Ein Fänger nennt, was er führt; der Implementer sieht die Mutation vor der Übergabe rot und trägt die Raten, falls ein Fänger nur mit Wahrscheinlichkeit rot wird, in den Absatz *Mutationen* dieser Sektion ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A01 | Modus: `MODE` ungesetzt oder leer ist `mock` und lädt ohne Verstoß; `mock`, `record`, `replay` laden; `hybrid` ist `ODR-E2001 MODE: Modus nicht unterstützt; lieferbar sind record, replay und mock`, auch neben `HYBRID_*`; `Mock` und `MOCK` melden nur die Form (`SPEC-110` Nr. 3, 8) | `TestKonfigModus` und die Zeilentabelle in `konfig_test.go`: Schleife über ungesetzt, leer, `mock`, `record`, `replay`, `hybrid`, `Mock`, `MOCK` mit der genauen Zeile | Default `record` oder `mock` weiter abgelehnt; Text „record und replay“ ⇒ rot | andere |
| A02 | Kein Upstream im Modus `mock`: `UPSTREAM_URL` und `upstream_url` dürfen fehlen, eine gesetzte wird in jedem Modus auf ihre Form geprüft (`SPEC-110` Nr. 9); `Build` im Modus `mock` liest `UPSTREAM_CA_FILE` nicht (`SPEC-139` Nr. 2) | Schleife über die Modi `record`, `replay`, `mock` × fehlt, gültig, ungültig × Einzelschlüssel, Datei (`konfig_test.go`); `Build` im Modus `mock` mit einer Datei, die nicht besteht, ohne Fehler (`wurzel_test.go`) | Pflicht auch im Modus `mock`; `Build` liest die Datei ⇒ rot | andere |
| A03 | `STRICT_COVERAGE` ist ein Wahrheitswert, ungesetzt oder leer `true`, in jedem Modus geprüft; die Konfiguration trägt den wirksamen Wert (`SPEC-110` Nr. 14 a) | Schleife über ungesetzt, leer, `true`, `false`, `TRUE`, `1`, `yes` × `record`, `replay`, `mock` | Default `false` (Nullwert des Feldes) ⇒ rot | andere |
| A04 | `PUBLIC_BASE_URL`, gesetzt: gültig ist genau, was der Schreiber der Version 4 als Wurzel annimmt; die Wurzel (Basis plus längster Root) hat im Modus `mock` die Grenze 2048 Bytes (`SPEC-110` Nr. 14 b, d) | Schleife über die Listen der gültigen und der ungültigen Formen (Userinfo, `?`, `#`, `\`, Leerraum, Steuerzeichen, Zeichen jenseits ASCII, `//` hinter dem Host, abschließender `/`, `ftp://`, `HTTP://`, leerer Host, Port mit sechs Ziffern, Segment `.`, `..`, `%2E`, `%` ohne zwei Hex-Ziffern); Orakel: dieselbe Zeichenfolge als `Wurzel` an `jsoncodec.NewSchreiber` ergibt `ErrUngueltig` genau dann; Länge 2048 gültig, 2049 Verstoß, Root `/` zählt nichts, bei zwei Diensten zählt der längste | abschließenden `/` zulassen; `<` statt `<=` an der Grenze ⇒ rot | andere |
| A05 | `PUBLIC_BASE_URL`, ungesetzt oder leer: im Modus `mock` Ableitung aus `LISTEN` bei Host `127.0.0.1`, `[::1]` oder `localhost` mit dem Port ohne führende Nullen, jeder andere Host ein Verstoß; in `record`, `replay`, `hybrid` und bei ungültigem `MODE` kein Verstoß und ein leerer Wert (`SPEC-110` Nr. 14 c) | Schleife über die drei Hosts und `0.0.0.0`, `[::]`, `LOCALHOST`, `127.0.0.2`, `example.org`, `[0:0:0:0:0:0:0:1]` × Port `8080`, `08080`, `000080` × die vier Modi und einen ungültigen | einen Host in die Liste aufnehmen; Port roh übernehmen ⇒ rot | andere |
| A06 | Nicht gelieferte Schlüssel bleiben `ODR-E2001`, auch im Modus `mock`: `ADMIN_*`, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`; `STRICT_COVERAGE` und `PUBLIC_BASE_URL` gehören nicht mehr dazu (`SPEC-110` Nr. 2) | Liste in `konfig_test.go` (Zeile 219) ohne die zwei Namen: Schleife über Namen und Präfixe × drei Modi, ein leerer Wert ohne Zeile | einen Namen oder ein Präfix aus `nichtGelieferte()` streichen; die zwei Namen zurücksetzen ⇒ rot | andere |
| A07 | Wahl je Modus: im Modus `mock` ist `Verarbeiten` des Listeners `Mock`, und es entstehen kein Upstream-Client, keine Pipeline, kein Recorder, keine Session, kein Import; `Sessions()`, `Record()`, `Import()` sind der Nullwert der Schnittstelle, `Workspaces()` und `Snapshots()` in jedem Modus gesetzt; `hybrid` bleibt `ErrUngueltig` (`SPEC-139` Nr. 1, 5, 11) | Tabelle Modus × Zugang mit `== nil` auf der Schnittstelle (`wurzel_test.go`); die Haken `NeuerUpstream`, `NeuePipeline`, `NeueSession`, `NeuerRecorder`, `NeuerImport` zählen im Modus `mock` null Aufrufe; der Fall `hybrid` bleibt | `h.verarbeiten = h.record.Record` im Modus `mock`; ein typisierter Nullzeiger ⇒ rot | andere |
| A08 | Konfig des Mock, jeder Wert unverändert: Dienste (`ID`, `Root`), `Metadaten` ist das Repository der Anlage, `Workspaces`, `Entities`, `Coverage`, `Transaktion` aus der einen Ablage, `MaxBytes`, `StrictCoverage` (beide Werte), `PublicBaseURL`; ein leerer `PublicBaseURL` ist `ErrUngueltig` und lässt die Ablage geschlossen (`SPEC-139` Nr. 11 a) | der neue Haken `NeuerMock` legt `mock.Konfig` ab (Fake, der jedes Argument festhält und seine Aufrufe zählt); Schleife über die Felder, zwei Dienste, `StrictCoverage` wahr und falsch; Literal-`Konfiguration` ohne `PublicBaseURL` | ein Feld auf den Nullwert; `StrictCoverage: true` fest ⇒ rot | andere |
| A09 | Satz der Version 4: ein Satz, von jedem Dienst getragen: `Version` `4.0`, `Metadaten`, `Daten`, `URL`, `Schreiber`, `Eingabe` nicht nil, `NurLesend` falsch; `MetadatenCodec`, `URLCodec`, `Schreiber` der Konfig sind die des Satzes (`SPEC-139` Nr. 11 b) | Haken-Test: Schleife über die Dienste und die sechs Felder mit Zählprüfung (sechs je Dienst) und den Typen der Werte; M3: `POST` mit Body ist `201` mit `Location` | `Schreiber` oder `Eingabe` nil; `Codecs: nil` ⇒ rot | andere |
| A10 | Filter-Leser: jeder Aufruf von `urlparser.New` im Root hängt `WithFilter(NewFilterLeser())` ein; `$filter` im Mock ist `200`, nicht `501` mit `ODR-E6001` (`SPEC-160` Nr. 8, `SPEC-139` Nr. 11 b) | M3: kombinierte Abfrage mit `$filter`; `go/ast`-Test über alle Aufrufe von `urlparser.New` in `src/bootstrap/*.go` ohne `_test.go` (Schleife, Zählprüfung mindestens 1) | `urlparser.New()` ohne Leser ⇒ rot | andere |
| A11 | Wächter und Snapshot-Dienst: ein Wächter je Prozess; `Blobs` von Mock, Snapshot-Dienst, Workspace-Dienst und Wartung sind derselbe Wert; Workspaces, Entities und Snapshots des Snapshot-Dienstes sind die Datenbank des Mock und des Workspace-Dienstes; der Rundlauf Snapshot, Mutation, Restore gelingt (`SPEC-139` Nr. 11 c, `SPEC-123` Nr. 9) | die Haken `NeuerWaechter` (ein Aufruf), `NeuerSnapshot`, `NeuerMock`, `NeueWorkspaces`, `NeueWartung` legen ihre Konfig ab; Vergleich der Zeiger der vier `Blobs` und der Repositories (Schleife über die vier Verbraucher); M3: Rundlauf | `Blobs` im Snapshot-Dienst auf das Repository unter dem Wächter; ein zweites `sqlite.Open` ⇒ rot | andere |
| A12 | Hochfahren und Herunterfahren im Modus `mock`: Services mit `4.0` angelegt, keine Session, `Wartung.Startup` gerufen; kein `CloseAll`; `Schliessen` mit `SHUTDOWN_TIMEOUT`; Ablage geschlossen; ein Start auf derselben Ablage (Build, Hochfahren, Herunterfahren, Build) hält Services, Workspaces, Entities und Snapshots (`SPEC-139` Nr. 6, 7, 11 d, f) | `lebenszyklus_test.go`: Zahl der Sessions der Ablage vor und nach `Hochfahren` gleich null; der Fake des Servers hält die Frist fest; Neustart im Prozess mit Lesen einer Mutation | `OpenInitialSession` im Modus `mock`; `Schliessen` ohne Frist ⇒ rot | andere |
| A13 | M3, Happy Path: Aufnahme (Modus `record`), Import, Workspace, Start im Modus `mock`, Referenzdienst beendet; `$metadata`; kombinierte Abfrage (`$filter`, `$orderby`, `$top`, `$skip`, `$count`); `POST`, `PATCH`, `DELETE` mit Lesen nach jedem Schritt; Snapshot, weitere Änderung, Restore, Reset; nach dem Stoppen keine Verbindung und kein Request am Referenzdienst (`LH-FA-12`, `LH-FA-06` bis `LH-FA-10`) | `TestM3Mock` in `tests/integration/m3_test.go` unter `make test` und `make test-race`; Zähler des Referenzdienstes vor und nach; kein Schlaf, Kanäle und Bereitschaft | der Mock ruft den Upstream; Restore oder Reset ohne Wirkung ⇒ rot | nebenläufig |
| A14 | Coverage: eine Collection auf nicht vollständigem Bestand ist im strikten Modus `ODR-E6003` (422), ohne ihn `200` mit `X-Recorder-Coverage`; auf vollständigem Bestand `complete` in beiden (`SPEC-011`, `LH-FA-11`) | M3: Schleife über strikt und nicht strikt × Bestand `complete` und nicht `complete` (zwei Baselines aus zwei Aufnahmen) | `StrictCoverage` fest oder vertauscht ⇒ rot | andere |
| A15 | Adressen und Fehlerantwort: `Location` und `@odata.id` sind `PUBLIC_BASE_URL` plus Root plus Set(Schlüssel), mit Basis ohne und mit Pfad, nie aus dem `Host` des Clients; die Fehlerantwort trägt `Content-Type: application/json` und `X-Correlation-Id` und **kein** Kennzeichen der Version. **Negativ** (Rückgabe R3, `SPEC-136` Nr. 8): Grund und Folge stehen dort; ein Fänger für die Abwesenheit entfällt, weil die spätere Ergänzung additiv ist | M3: Anfrage mit fremdem `Host`-Header; Basis `http://127.0.0.1:<Port>` und mit dem Pfad `/api`; Fehlerantwort `ODR-E6008` | die Wurzel aus `r.Host` ⇒ rot | andere |
| A16 | Handbuch und Kataloge: Kapitel mit `mock`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL` und den Codes des Mock einschließlich `ODR-E6007`, belegt durch ein Beispiel als Datei; die acht Zeilen des Mock stehen nicht mehr in `harness/katalog-ausnahmen.txt`. **Negativ:** das Gate misst die Erreichbarkeit eines Codes in den Ausnahmen nicht (`SPEC-201` Nr. 9 c); eine stehen gebliebene Zeile bliebe grün, die Prüfung liegt beim Reviewer, Folge: der Anwender sähe einen Code ohne Eintrag im Handbuch | `make handbuch-check`, `make katalog-check` | das Beispiel löschen; einen Code aus dem Handbuch streichen ⇒ rot | andere |
| E01 | `TestE2EMock`: Mock CRUD am gebauten Binary: Start im Modus `mock` ohne `UPSTREAM_URL` und `PUBLIC_BASE_URL`, `POST` `201` mit `Location` aus `LISTEN`, `GET`, `PATCH` `204`, `DELETE` `204`, `GET` `404` `ODR-E6008`, nach jedem Schritt der Zustand; die Collection ist `complete`; ein nicht unterstützter Ausdruck von `$filter` ist `501` (`SPEC-037` Nr. 7) | Stufe 8 des Treibers | das Binary lehnt `mock` ab; `Location` mit festem Host ⇒ rot | andere |
| E02 | Reset und Zuordnung am Binary: `SIGTERM`, `reset` des Übergangs, Start: die Collection gleicht der Antwort vor den Mutationen; Request ohne `X-Recorder-Workspace` ist `400` `ODR-E5003`, mit unbekanntem Wert `404` `ODR-E7001` | `TestE2EMock` | Reset ohne Wirkung ⇒ rot | andere |
| E03 | Kein Upstream: nach dem Stoppen des Referenzdienstes zählt er in keinem Schritt eine Verbindung oder einen Request mehr (`LH-FA-12`) | Zähler des Referenzdienstes in `TestE2EMock` und `TestE2EMockNeustart` | der Pfad des Mock greift auf einen Upstream zu ⇒ rot | andere |
| E04 | `TestE2EMockNeustart`: Mutationen (`POST`, `PATCH`, `DELETE` an drei Entities), `SIGTERM` (Exit 0), Start auf demselben `DATA_DIR`: die Zustände sind lesbar, ein zweiter Neustart ohne Änderung ebenso (`SPEC-049` Nr. 7, `SPEC-139` Nr. 11 f) | Stufe 9 des Treibers; die Antworten vor dem Signal sind das Orakel | der Start leert `DATA_DIR`; das Beenden schließt die Transaktion nicht ab ⇒ die Mutation fehlt, rot | nebenläufig |
| E05 | `DATA_DIR` besteht nicht: der Start legt es an und ist bereit, jeder Request mit Workspace ist `404` `ODR-E7001` (`SPEC-139` Nr. 11 f) | `TestE2EMockNeustart`, Boundary | das Anlegen entfällt ⇒ der Start scheitert ⇒ rot | andere |
| E06 | Befehl `reset` des Übergangs: fehlendes oder leeres Argument Exit 2, unbekannter Workspace Exit 1 mit `ODR-E7001`, sonst `revision=<Zahl>` (`SPEC-037` Nr. 6) | Test der Funktion `lauf` in `tests/e2e/uebergang/` (Schleife über die Argumentformen); gültiger Aufruf in `TestE2EMock` | Exit-Codes vertauscht ⇒ rot | andere |
| E07 | Stufen im Treiber: je Stufe drei Zeilen, der Name steht in `-test.list`, ohne `PASS` ist die Stufe rot; `make e2e-abdeckung` führt Stufe 8 und 9 (`SPEC-037` Nr. 2, 3) | ein Lauf von `make test-integration`; `docs/plan/planning/e2e-abdeckung.md` aus `make e2e-abdeckung` erzeugt | eine Anker-Zeile mit falschem Namen ⇒ `nicht gefunden`, Exit 2 | andere |
| E08 | **Negativ:** `SIGTERM` während einer Mutation am Binary; Grund und Folge in `SPEC-037` Nr. 10 i und `SPEC-139` *Ungeprüft* | kein Fänger am Binary; getrennt belegt: `SPEC-136` Nr. 15 (Frist) und `SPEC-192` Nr. 7 (Atomarität) | — | andere |
| E09 | **Negativ:** Snapshots über die Prozessgrenze; Grund und Folge in `SPEC-037` Nr. 10 j und `SPEC-049` Nr. 10 b | im Prozess: `SPEC-049` Nr. 7 d | — | andere |

- Der Default `MODE=mock` lässt den Start bis zu diesem Slice mit `ODR-E2001` scheitern; wer auf dieses Verhalten testet, ist mit zu ändern. **Liste der Tests** (Architect, 2026-11-08; der Implementer ergänzt sie mit `grep -n 'MODE=mock\|"mock"\|lieferbar sind\|STRICT_COVERAGE\|PUBLIC_BASE_URL' src/bootstrap/*_test.go tests/e2e/*.go` und nennt im Bericht jeden Treffer, geändert oder nicht):
  - `src/bootstrap/konfig_test.go`: `TestKonfigModus` (Schleife über `mock`, `hybrid`, ungesetzt: `mock` und ungesetzt sind kein Verstoß mehr), der Fall mit `MODE=mock` und `LISTEN=kaputt` (bei Zeile 391), die Liste der ungültigen Werte bei Zeile 701 und der Fall *Modus ungültig* bei Zeile 724, der `mock`-Fall von `upstream_url` bei Zeile 788 (der Text nennt „record und replay“) und die Listen der Schlüssel bei den Zeilen 97 und 219 (`STRICT_COVERAGE` und `PUBLIC_BASE_URL` sind geliefert);
  - `src/bootstrap/start_test.go` bei Zeile 232: `TestKonfigFehlerBeiAbbruch` nimmt `MODE=mock` als Verstoß, es wird `MODE=hybrid`;
  - `src/bootstrap/wurzel_test.go` bei Zeile 522: `Build` mit `Modus` `mock` ist nicht mehr `ErrUngueltig`, `hybrid` bleibt;
  - `tests/e2e/stufe_prozess_test.go`: der Fall `leer` von `ungueltigeKonfiguration` erwartet `MODE`; die leere Umgebung startet jetzt, der Fall wird `MODE=hybrid` (`SPEC-037` Nr. 5, 7).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-mock-abnahme`, `slice-snapshot-use-case`, `slice-test-integration` und `slice-odata-url-filter` liegen in `done/` (der Leser, den der Root einhängt, muss existieren). Die Spezifikation trägt den Nachzug von `SPEC-110`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Stand beim Start** (Planner, 2026-11-08, `main` = `741e7ec9`, Stempel der Gegenproben gleich dem Stand nach `slice-mock-abnahme`): `slice-mock-abnahme`, `slice-snapshot-use-case`, `slice-test-integration` und `slice-odata-url-filter` liegen in `done/` (Trigger Teil 1 erfüllt). Trigger Teil 2 (Nachzug von `SPEC-110` Nr. 2 und 8 in der Spezifikation) **am 2026-11-08 erfüllt** durch den Architect (Commit dieses Zweigs: `SPEC-110` Nr. 2, 8, 9 und 14, `SPEC-139`, `SPEC-136` Nr. 8, `SPEC-037`); er war Bedingung für den ersten Code-Commit, nicht für den Start. §1 *Ausdrücklich NICHT* und DoD aller Nehmer-Sendungen gelesen (`AGENTS.md` §3.13): kein Ausschluss trifft sie. Codes: `model.AlleCodes()` zählt auf `main` 40 (mit `ODR-E2102` aus `slice-admin-rahmen`); dieser Slice fügt höchstens `ODR-E5004`-Fälle hinzu, wenn der User sie entscheidet (Kennungsfrage R3 beim User); wer als zweiter nach `main` merged (`slice-admin-authentisierung` läuft parallel), gleicht die Zahl in den Zählstellen an (`welle-mock` §5, Merge-Nähe). Register am Start nachgezogen: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (5×), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×), `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (4×), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (24×, verkörpert in `AGENTS.md` §3.9).

**Gate-Kette (Nr. 7 der Übergangsregeln):** Linie B. Von den Slices der `welle-mock` ändert allein dieser die Verdrahtung der Szenarien von `make test-integration` (Treiber `tools/harness/test-integration.sh`, Stufe `integration` des `Dockerfile`, `tests/e2e`); `slice-test-contract` folgt danach, nicht zugleich. `Makefile`, `harness/mk/`, `harness/README.md`, `tools/gegenproben/` und Stempel bleiben unberührt. **Bestätigt** (Architect, 2026-11-08): Linie B, aus drei Gründen. Nr. 7 nennt `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` und den Stempel; der Treiber `tools/harness/test-integration.sh` steht nicht darunter, und eine Stufe ist die Deklaration von drei Zeilen dort und ein Test (`SPEC-037` Nr. 3). Die Stufe `integration` des `Dockerfile` baut `./tests/e2e` und `./tests/e2e/uebergang` als Ganzes (`go test -c`, `go build`), so dass eine neue Datei und ein neuer Befehl des Übergangs keine Zeile des `Dockerfile` ändern; `reset` nutzt das Handle im Modus `record` und braucht keinen Modus `mock` im Übergang. Das Abdeckungs-Dokument entsteht aus dem Treiber (`make e2e-abdeckung`). Grenze: `make test-integration` ist kein Gate (`GATE_CHECKS` ohne es); die Rückgabe an den Planner bleibt, falls ein Szenario doch das `Dockerfile` braucht. Braucht ein Szenario eine Änderung des `Dockerfile`, geht das als Rückgabe an den Planner (Nr. 7): `slice-handbuch-katalog-schluessel-kommandos` hält die Gate-Kette in `in-progress/`.

Größenmarke und Schnitt (Architect, 2026-11-08, `SPEC-128` Nr. 1; die Prosa-Form des Planners vom selben Tag ersetzt, weil die Tabelle in §3 jetzt vorliegt): Der Block unten trägt den Teil des Architect für den **ungeschnittenen** Slice; `basis`, `schaetzung` und `marke` setzt der Planner in eigenem Commit vor dem ersten Code-Commit (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; `docs/` zählt nach `SPEC-128` Nr. 3 m nicht, das Handbuch also auch nicht).

**Schnitt, vor dem Code benannt** (`AGENTS.md` §3.14): **Teil E** sind die Zeilen E01 bis E09 der Tabelle: die Stufen `TestE2EMock` und `TestE2EMockNeustart` am gebauten Binary, der Befehl `reset` des Übergangs, die drei Zeilen je Stufe im Treiber, das erzeugte Abdeckungs-Dokument; Pfade: `tests/e2e/` (neue Dateien und `uebergang/`) und `tools/harness/test-integration.sh`. **Teil A** ist der Rest (A01 bis A16): Konfiguration, Verdrahtung, M3-Test, Handbuch, dazu allein der Fall `leer` von `TestE2EProzess` in `tests/e2e/stufe_prozess_test.go`, den der Wechsel des Defaults bricht (`SPEC-037` Nr. 5). **Kein geteilter Mechanismus:** E fährt das gebaute Binary von außen (Prozess, Umgebung, Signal, HTTP) und ruft aus `src/` allein `bootstrap.Build` im Übergang, wie die Stufen 1 bis 7; A hat keinen Code, den E mitbenutzt, und `tests/e2e` und `tests/integration` teilen kein Paket und keine Hilfe. E braucht das Binary aus A (Reihenfolge, kein Mechanismus).
Zur Nachrechnung des Planners (nicht verbindlich, `SPEC-128` Nr. 1): Teil A: *a* = 1, *b* = 15, *f* = 1 (Fake des Haken-Tests), *p* = 260: *S* = 1745, *M* = 2400, keine Schnittfrage. Teil E: *a* = 1, *b* = 8, *f* = 0, *p* = 40: *S* = 765, *M* = 1100. Ungeschnitten: *a* = 2, *b* = 23, *f* = 1, *p* = 300: *S* = 2510, *M* = 3400, Schnittfrage ja (*S* ≥ 2223). **Vorschlag:** der Planner schneidet vor dem Code: `slice-bootstrap-m3` behält A; ein neuer Slice (Kennung und Welle: Planner) nimmt E samt der Sendung von `slice-mock-abnahme` (`SPEC-049` Nr. 7 a und 10 b), mit dem Eintrag beim Nehmer vor dem Eintrag beim Geber (`AGENTS.md` §3.13) und dem Start-Trigger „`slice-bootstrap-m3` in `done/`“. Schneidet er nicht, ist E der vorab benannte Teil nach `AGENTS.md` §3.14: zuletzt in eigenem Commit geschrieben und davor gemessen. Nach dem Schnitt entfallen im Block von A der Pfad `tests/e2e/` zugunsten von `tests/e2e/stufe_prozess_test.go`, der Pfad `tools/harness/test-integration.sh`, und `produktcode` wird 260.

```groessenmarke
pfad: src/bootstrap/
pfad: tests/integration/
pfad: tests/e2e/
pfad: tools/harness/test-integration.sh
pfad: harness/katalog-ausnahmen.txt
produktcode: 300
fakes: 1
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Tabelle *Zusage · Fänger* liegt vor (§3); die Marke setzt der Planner (Block oben), den Schnitt-Vorschlag nennt der Absatz *Größenmarke und Schnitt*.
- `in-progress` → `open` (blockiert — Carveout?): der Wechsel des Defaults `MODE` bricht Tests, die den Start ohne `MODE` erwarten, und der Nachzug von `SPEC-110` liegt nicht vor (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-08, Zeilen dieses Slice): jede Zelle nennt eine Nummer der Spezifikation oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12). `SPEC-139` und `SPEC-110` meinen die Fassung nach dem Nachzug vom 2026-11-08.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Laden von `MODE` | `SPEC-110` Nr. 8: der Default `mock` startet | Nr. 1: gleich ungesetzt | Nr. 3 (Schreibweise), Nr. 8 (`hybrid`: `ODR-E2001`) | gilt nicht: das Laden hat keinen Kontext (Nr. 13); der Abbruch beim Start: `SPEC-139` Nr. 8 (a), (b) | Nr. 13: dieselbe Umgebung, dasselbe Ergebnis |
| Laden von `STRICT_COVERAGE` | Nr. 14 a: ungesetzt ist `true` | Nr. 1 und 14 a | Nr. 3: `1`, `yes`, `TRUE` sind ungültig | gilt nicht (wie `MODE`) | Nr. 13 |
| Laden von `PUBLIC_BASE_URL` | Nr. 14 c: Ableitung aus `LISTEN` oder Verstoß, je Host und Modus | Nr. 1 und 14 c | Nr. 14 b: Form; Nr. 14 d: Länge der Wurzel | gilt nicht (wie `MODE`) | Nr. 13 |
| Laden von `UPSTREAM_URL` im Modus `mock` | Nr. 9: nicht erforderlich | Nr. 1 | Nr. 7 und 9: die Form wird in jedem Modus geprüft | gilt nicht (wie `MODE`) | Nr. 13 |
| `Build` im Modus `mock` | `SPEC-139` Nr. 2: keine Dienste sind `ErrUngueltig`; Nr. 11 a: ein leerer `PublicBaseURL` ebenso | Nr. 2: ein leeres `DATA_DIR` ist `ErrPfad` | Nr. 5: `hybrid` ist `ErrUngueltig`; Nr. 2: ein Modus außer den dreien | Nr. 2: ein Fehler der Ablage kommt zurück, `audit.jsonl` bleibt unberührt | Nr. 10 und 11 f: ein zweites `Build` nach `Herunterfahren` ist der Neustart; ohne `Herunterfahren` nicht unterstützt (Nr. 2) |
| `Hochfahren` im Modus `mock` | Nr. 11 d: keine Session entsteht | gilt nicht: kein Argument außer `ctx` | Nr. 6 (2) und 11 f: ein gespeicherter Service mit anderem `service_root` oder `odata_version` verhindert den Start | Nr. 6: `ctx` wird vor jedem Schritt geprüft | Nr. 6: ein zweites `Hochfahren` ist `ErrUngueltig` |
| `Herunterfahren` im Modus `mock` | gilt nicht: kein Argument | gilt nicht: kein Argument | gilt nicht: kein Argument | Nr. 7 (`WithoutCancel`); eine laufende Mutation: Nr. 11 d | Nr. 7: wiederholbar, ein Lauf (`sync.Once`) |
| Zugänge des `Handle` (`Import`, `Sessions`, `Record`, `Workspaces`, `Snapshots`) | Nr. 1 und 11: der Nullwert der Schnittstelle, nie ein typisierter Nullzeiger | gilt nicht: kein Argument | Nr. 1 und 11: die Tabelle Modus × Zugang (Zeile A07) | gilt nicht: lesen nur Felder | Nr. 10: gleichzeitiger Aufruf ist sicher |
| Request am Listener im Modus `mock` | `SPEC-195` Nr. 5: ohne `X-Recorder-Workspace` ist es `ODR-E5003` | Nr. 5: ein leerer Wert ebenso | Nr. 8: ein unbekannter Workspace ist `ODR-E7001`, ein fremder Service `ODR-E5004` | `SPEC-136` Nr. 9 (a) | `SPEC-192` und `SPEC-193`: ein doppelter Schlüssel ist `ODR-E6009` |
| Start auf vorhandenem `DATA_DIR` | `SPEC-139` Nr. 11 f und `SPEC-110` Nr. 11: ein fehlendes Verzeichnis wird angelegt, der Start gelingt | Nr. 11 f: eine Ablage ohne Baseline und ohne Metadaten ist `ODR-E7001` und `ODR-E6004` | Nr. 11 f: die Ablage eines früheren Laufs in einem anderen Modus wird übernommen | Nr. 11 d: eine Mutation beim Herunterfahren ist fertig oder hinterlässt nichts; am Binary ein Negativ (`SPEC-037` Nr. 10 i) | Nr. 11 f und `SPEC-049` Nr. 7 d: ein zweiter Neustart ohne Änderung ergibt dasselbe |
| Befehl `reset` des Übergangs | `SPEC-037` Nr. 6: ein fehlendes Argument ist Exit 2 | Nr. 6: ein leeres ebenso | Nr. 6: ein unbekannter Workspace ist Exit 1 mit `ODR-E7001` | gilt nicht: ein Hilfsprogramm des Tests, das Signal beendet den Prozess | `SPEC-129` Nr. 10: jeder Reset erhöht die Revision |

Zusatztabelle (mehrere Verstöße und Zusagen je Modus; `AGENTS.md` §3.16):

| Fall | Ort | Fänger |
|---|---|---|
| `record`: `UPSTREAM_URL` Pflicht, `PUBLIC_BASE_URL` ungesetzt ohne Verstoß | `SPEC-110` Nr. 9, 14 c | A02, A05 (Schleife über die Modi) |
| `replay` und `mock`: `UPSTREAM_URL` nicht erforderlich, eine gesetzte wird auf ihre Form geprüft | `SPEC-110` Nr. 9 | A02 |
| `mock`: `PUBLIC_BASE_URL` Pflicht bei anderem Host als Loopback | `SPEC-110` Nr. 14 c | A05 |
| `hybrid` und ungültiger `MODE`: nur die Zeile `MODE`, keine Zeile `UPSTREAM_URL` oder `PUBLIC_BASE_URL` wegen der Pflicht | `SPEC-110` Nr. 9, 14 c | A01, A05 |
| `MODE=hybrid` zugleich mit `HYBRID_*`: zwei Zeilen (zwei Schlüssel), nach Schlüssel sortiert | `SPEC-110` Nr. 2, 12 | A01 |
| `PUBLIC_BASE_URL` mit ungültiger Form und zu langer Wurzel: nur die Form, höchstens ein Grund je Schlüssel | `SPEC-110` Nr. 14 d, 12 | A04 |
| `STRICT_COVERAGE` ungültig zugleich mit `MODE` ungültig: zwei Zeilen | `SPEC-110` Nr. 3, 12 | A03 |

- Neustart über die Prozessgrenze im Modus `mock` (Sendung von `slice-mock-abnahme`, `SPEC-049` Nr. 7 a): das Szenario braucht ein `DATA_DIR` im Modus `mock` und ein geordnetes Beenden durch `SIGTERM`; Randformen (Verzeichnis fehlt, Signal während einer Mutation) entscheidet der Architect vor dem Code (`AGENTS.md` §3.12), Snapshots bleiben das Negativ aus `SPEC-049` Nr. 10 b — **Ausgang:** entschieden (Architect, 2026-11-08): die Stufe `TestE2EMockNeustart` (Teil E, Zeilen E04 und E05); Randformen im Raster (Zeile *Start auf vorhandenem `DATA_DIR`*) und in `SPEC-139` Nr. 11 d und f: ein fehlendes Verzeichnis wird angelegt, der Start gelingt und jede Anfrage mit Workspace ist `ODR-E7001`; ein Signal während einer Mutation ist ein akzeptiertes Negativ mit Grund und Folge (`SPEC-037` Nr. 10 i, `SPEC-139` *Ungeprüft*), Snapshots bleiben das Negativ (`SPEC-037` Nr. 10 j).
- Ablage des M3-Tests (Rückgabe R1 von `slice-mock-abnahme`, Planner, 2026-11-08): das Architektur-Gate gibt dem Paket `tests` Kanten auf Modell, Inbound-Ports und Driven Adapter, nicht auf die Dienste und nicht auf den Composition Root; die Abnahme des Gebers liegt deshalb in `src/bootstrap/abnahme_*_test.go`. Die Zeile `tests/` in §3 gilt nur, wenn der M3-Test über erlaubte Kanten auskommt (Binary und Listener von außen); sonst liegt er in `src/bootstrap/`. Eine Änderung der Architekturregel oder eine ADR beginnt dieser Slice nicht — **Ausgang:** entschieden (Architect, 2026-11-08): der M3-Test liegt in `tests/integration/m3_test.go` (Paket `integration_test`, neben `m2_test.go` und dem Referenzdienst, den er mitbenutzt). Er kommt mit den Kanten von [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md) aus: Modell, Inbound-Ports, Driven Adapter, dazu `bootstrap` als Composition Root (ohne Kante, wie `m2_test.go` und `tests/e2e/uebergang`); er baut mit `bootstrap.Build`, nimmt Import, Workspaces, Snapshots und Listener vom `Handle` und fährt HTTP mit `net/http` und braucht keinen Dienst (`services/`) und keinen Outbound-Port. Der Grund, aus dem die Abnahme des Gebers in `src/bootstrap/` liegt, trifft nicht zu: jener Test baut `mock.New` und `snapshot.New` selbst und braucht dafür die Dienste. Keine Folge-ADR und keine Änderung von `.a-check.yml`. Braucht der Test doch eine Kante, die die Schicht nicht hat (Dienste, Outbound-Ports, Driving Adapter), liegt er in `src/bootstrap/m3_*_test.go` (der Composition Root darf alles importieren, wie die Abnahme); das ist kein Ausweg, der ein Gate lockert, und der Bericht nennt den Grund.
- Der Wechsel des Defaults `MODE=mock` macht aus einem heute abgewiesenen Start einen gültigen; Tests, die den alten Fehler erwarten, sind mit zu ändern — **Ausgang:** entschieden (Architect, 2026-11-08): die Liste steht in §3, der Wechsel in `SPEC-110` Nr. 8; ob eine Zeile fehlt, zeigt das `grep` dort.
- `PUBLIC_BASE_URL` ist bei Wildcard-Host (Container) Pflicht und prägt `Location` und `@odata.id` (`SPEC-055`); Randformen (fehlt, leer, mit Pfad, mit abschließendem `/`) — **Ausgang:** entschieden (Architect, 2026-11-08): `SPEC-110` Nr. 14 und das Raster unten; ein Pfad ist gültig, ein abschließender `/` ungültig, fehlend wird bei Loopback-`LISTEN` abgeleitet und ist sonst ein Verstoß, leer gleich fehlend.
- Größe: die Schätzung (*S* 2810, *M* 3800) liegt über der Schwelle der Schnittfrage (`SPEC-128` Nr. 1), die Tabelle fehlt (§4) — **Ausgang:** Tabelle (§3), Block und Schnitt-Vorschlag (§4) liegen vor (Architect, 2026-11-08); weiter offen bleiben die Marke und die Antwort auf die Schnittfrage; Adresse: Planner (vor dem ersten Code-Commit).
- Gate-Kette: ein Szenario, das das `Dockerfile` ändert, kollidiert mit `slice-handbuch-katalog-schluessel-kommandos` (hält Nr. 7) — **Ausgang:** entschieden (Architect, 2026-11-08): Linie B bestätigt (§4); kein Szenario braucht eine Änderung des `Dockerfile`. Braucht ein Szenario sie doch, ist das weiter eine Rückgabe an den Planner (Adresse), nicht Sache des Codes.
- Zahl der Codes und Merge-Nähe zu `slice-admin-authentisierung` (läuft parallel) — der Architect führt keinen Code ein (Rückgabe R3 ist ein Negativ) und berührt keine der sechs Zählstellen; die Berührungspunkte beim Merge sind die Zeile `SPEC-110` Nr. 2 und die Historie der Spezifikation, `harness/katalog-ausnahmen.txt` und das Handbuch — **Ausgang:** offen; Adresse: Planner (Integrationsschritt, `welle-mock` §5).
- Kopplung an `slice-handbuch-katalog-schluessel-kommandos` (hält die Gate-Kette, läuft parallel): dessen Katalog der Schlüssel liest K aus der Zeile `Geprüft werden` von `SPEC-110` Nr. 2 und zählt 27 Schlüssel, 26 davon im Code, und 6 abgelehnte Namen und Präfixe; dieser Slice macht `STRICT_COVERAGE` und `PUBLIC_BASE_URL` zu gelesenen Schlüsseln (K 29, im Code 28) und lässt 4 Präfixe abgelehnt. Die Zeile bleibt im Format, das `SPEC-201` Nr. 5 b liest; liegt der Katalog vor diesem Slice in `main`, fügt dieser Slice zwei Zeilen in die Tabelle *Einstellungen* ein und zieht die festen Zahlen der Fänger (27, 26, 6) nach, liegt er dahinter, tut das jener Slice — **Ausgang:** offen; Adresse: Planner (Integrationsschritt, Reihenfolge der Merges).
- Der M3-Test läuft mit Referenzdienst, Recorder und Client im selben Prozess unter dem Detektor (`make test-race`) — **Ausgang:** offen; Muster und Ordnung der Fänger wie in `slice-m2-referenzdienst` (Kanäle statt Schlaf).
- `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×): der Mock-Pfad im Composition Root erweitert die Pakete, die Fälle der Gegenproben mitfahren — **Ausgang:** weiter offen → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` im Register.
- Nachzug von `SPEC-110` Nr. 2 und 8 (Bedingung für den ersten Code-Commit, §4) — **Ausgang:** entschieden (Architect, 2026-11-08): im Zweig nachgezogen (`SPEC-110` Nr. 2, 8, 9, neue Nr. 14; `SPEC-139` Nr. 1 bis 5, 7, 9, neue Nr. 11; `SPEC-136` Nr. 8; `SPEC-049` Nr. 10 a; `SPEC-037` Nr. 5 bis 7 und 10; Historie der Spezifikation). Der Implementer prüft den Wortlaut gegen den Code und hält an einer Abweichung an (`AGENTS.md` §3.12).
- Version in der Fehlerantwort des Mock (Rückgabe R3 von `slice-mock-zuordnung` und `slice-mock-lesen`) — **Ausgang:** entschieden (Architect, 2026-11-08): akzeptiertes Negativ mit Grund und Folge in `SPEC-136` Nr. 8 (§1); nichts zu verdrahten, nichts an den Planner.
- `Import()` im Modus `mock` ist `nil` (`SPEC-139` Nr. 1 und 11, *Ungeprüft*): ob der Admin-Eingang einen Import im Modus `mock` mit gesetzter `UPSTREAM_URL` zulässt, entscheidet der Vertrag, der ihn verdrahtet — **Ausgang:** entschieden (Architect, 2026-11-08): akzeptiertes Negativ mit Grund und Folge in `SPEC-139` *Ungeprüft*; der Architect des Admin-Eingangs liest es dort, eine Adresse in einem Nachbarplan ist nicht nötig, weil nichts spurlos verschwindet (`SPEC-139` ist die Quelle, die er für den Root ohnehin liest).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Offen und für diesen Slice einschlägig: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×, der Mock-Pfad erweitert die Pakete, die Fälle der Gegenproben mitfahren; Ausgang in §6), `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) für die Fakes im Root-Test. Verkörpert: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `AGENTS.md` §3.9) für den Nachzug von `SPEC-110`. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
