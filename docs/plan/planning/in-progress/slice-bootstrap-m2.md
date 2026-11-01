# Slice slice-bootstrap-m2: Composition Root (Verdrahtung, Start und Herunterfahren)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-12`, `LH-FA-04`, `LH-FA-02`, `LH-QA-02`, `LH-RB-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-139` (Vertrag des Composition Root, Nr. 1 bis 10) · `SPEC-138` Nr. 8 (Pflichten des Starts, e und f) · `SPEC-136` Nr. 12 (Rückgabe von `Serve`), Nr. 15 (Reihenfolge des Herunterfahrens) · `ARC-012` (Composition Root) · `LH-FA-12.a` · `LH-FA-02.c` · `SPEC-048` · `architecture.md` §2.5 (Modus-Routing), §4.9 · `SPEC-120` Nr. 1, `SPEC-110` Nr. 10, `LH-RB-02.a` (Verdrahtung des Upstream-Adapters, Warnung beim Start) · `SPEC-123` Nr. 1, 3, 9 (Wächter vor dem Blob-Port, Startwartung) · `SPEC-124` Nr. 5, 14 (erste Session, `CloseAll`) · `SPEC-107` Nr. 14 (Services anlegen).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-bootstrap-m2)

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Composition Root (`SPEC-139`) verdrahtet Konfiguration, Adapter, Services und Listener je Betriebsmodus (`record`, `replay`), fährt Startwartung, erste Session und geordnetes Herunterfahren, und der Start-Befehl lädt die Konfiguration über `LadeOderMelde` (Exit-Code 1 und Meldung nach `SPEC-110` bei Verstoß; `mock` und `hybrid` scheitern mit `ODR-E2001`; übernommen von `slice-fehlermodell-konfiguration`, das Laden selbst liegt dort). Der sichtbare Schritt: Aufnehmen ist mit diesem Slice benutzbar (Start, ein Request durch den HTTP-Adapter an einen Upstream, Herunterfahren mit Session `complete`); Abspielen ist verdrahtet (ein Request im Modus `replay` erreicht den Replay-Dienst und erhält `ODR-E5003`), aber ohne Import und Workspace nicht zu erreichen.

**Herkunft (Schnitt vor dem ersten Code-Commit, `AGENTS.md` §3.14):** Der Plan trug Root-Kern und M2-Beleg in einem Slice; die Formel von `SPEC-128` gab für beide zusammen S rund 3200 (Schnittfrage), für den Kern allein S = 2120 (Marke 2900, §4). Geschnitten ist dort, wo kein Mechanismus geteilt wird: Referenzdienst, Import und Workspace sind additive Zugänge am `Handle` (`SPEC-139` Nr. 1), und der Kern kommt für seinen sichtbaren Schritt mit einem kleinen Upstream im Testcode von `src/bootstrap` aus. Der M2-Beleg gehört dem Slice `slice-m2-referenzdienst` (nach diesem Slice, vor `slice-protokolltreue`). Der Rest teilt `Handle` und Naht (§3) und wird nicht weiter geteilt.

**Übernimmt:** Der Composition Root erzeugt genau eine `clock.NewSystemuhr()` und eine `clock.NewZufallsIDs()` je Prozess und reicht sie an alle Verbraucher; zwei Uhren wären zwei nicht aufeinander abgestimmte Höchstwerte (`SPEC-114` Nr. 2, `SPEC-139` Nr. 3; Sendung von `slice-clock-idgenerator`).

**Übernimmt (Session und Wartung):** Der Composition Root verdrahtet `MaintenanceUseCase` und `SessionUseCase` (`SPEC-139` Nr. 3, 6, 7): beim Start zuerst die Startwartung (Sendung von `slice-session-wartung`), danach im Modus `record` `OpenInitialSession` je konfiguriertem Service (`ARC-012`; Sendung von `slice-session-use-case`, `SPEC-124` Nr. 5), beim Herunterfahren `CloseAll` mit `SHUTDOWN_TIMEOUT` (Sendung von `slice-session-beenden`, `SPEC-124` Nr. 14 und *Ungeprüft*, `LH-FA-02.c`); die drei Slices liegen in `done/` (§4). `SessionUseCase` erhält dieselbe Instanz von `record.New` wie der HTTP-Adapter: `Drain` wirkt im Zustand des Dienstes, der die Aufnahme annimmt (`SPEC-121` Nr. 13, `SPEC-124` *Ungeprüft*); mit einer zweiten Instanz endete `CloseAll` ohne Wirkung auf die Annahme. Der Root ruft `CloseAll` nach `Anhalten` und vor `Schliessen`; ein Fehler von `CloseAll` oder `Fehlgeschlagen` größer `0` hält das Schließen der Ablage nicht auf (die Sessions setzt der nächste Start auf `incomplete`), und der Kontext des Herunterfahrens ist `context.WithoutCancel(ctx)` (`SPEC-138` Nr. 8 b). Die Fänger stehen in der Tabelle *Zusage · Fänger* (§3: S1, B1, B5, B6).

**Übernimmt (Services):** Der Composition Root legt jeden konfigurierten Service an, den die Ablage nicht kennt: nach der Startwartung und vor `OpenInitialSession`, einmal je Service, mit `MetadataRepository.Create` (`odata_version` `4.0`, Metadaten leer); ein gespeicherter Service mit anderem `service_root` oder anderer `odata_version` verhindert den Start (`ODR-E2001`), einer, den die Konfiguration nicht nennt, bleibt unberührt (`SPEC-107` Nr. 14, `SPEC-139` Nr. 6; Sendung von `slice-odata-codec-csdl` und Entscheidung der benannten Spec-Lücke aus `slice-session-use-case`, `AGENTS.md` §3.13). Die Metadaten setzt allein der Import (`SPEC-107` Nr. 15); der Root ruft `SetMetadata` nie. Fänger: B2.

**Übernimmt (Wächter):** Der Composition Root erzeugt genau einen `maintenance.BlobWaechter` je Prozess (mit `maintenance.StandardKarenz` und der einen Uhr) über dem Blob-Adapter und gibt ihn als `BlobRepository` der Capture-Pipeline, dem Replay-Dienst und der Wartung als `Blobs`; wer das darunterliegende Repository erhält, ist ungeschützt (`SPEC-123` Nr. 1, 9, *Ungeprüft*; `SPEC-139` Nr. 3; Sendung von `slice-session-wartung`). Fänger: B3 (Zeigergleichheit des Wächters über die Naht). Den Wächter am Import trägt `slice-m2-referenzdienst`.

**Übernimmt (Record):** Der Composition Root verdrahtet `record.New` (`SPEC-121` Nr. 1) aus der Startkonfiguration (Dienste mit Root und `UpstreamURL`, `SchluesselRegeln`, `capture.Standard` mit `MAX_CAPTURE_BODY` und `REDACT_JSON_PATHS`) mit dem Client aus `httpupstream.New` als `UpstreamClient` und gibt dem Dienst genau einen `*slog.Logger` (`SPEC-121` Nr. 12; einen Port `Telemetry` gibt es in dieser Welle nicht). Im Modus `replay` baut er stattdessen `replay.New` mit den `SchluesselRegeln` desselben Satzes, aber weder Recorder noch Pipeline noch Upstream-Client (`SPEC-139` Nr. 3, 4, 5; Sendung von `slice-record-service`). Fänger: B7, B8.

**Übernimmt (Upstream):** Der Composition Root verdrahtet `httpupstream.New` (`SPEC-120` Nr. 1): `Frist` aus `UPSTREAM_TIMEOUT`, `Pruefen` die wirksame Zertifikatsprüfung nach `SPEC-110` Nr. 10 (falsch genau bei `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true`; der Nullwert von `Pruefen` bleibt nie stehen), `CAPEM` der Inhalt von `UPSTREAM_CA_FILE` (leer ohne Datei); ist die Prüfung abgeschaltet, schreibt der Start genau eine Warnung ins Log (`LH-RB-02.a`, `SPEC-139` Nr. 9). Der Start verlangt `slice-upstream-adapter` und `slice-upstream-tls` in `done/` (§4; Sendungen beider, Befund I-1 des Reviews von `slice-upstream-tls`). Fänger: B4. Der Wert des Starts selbst ist in `src/bootstrap/konfig_test.go` gehalten (`TestKonfigDefaults`, `TestKonfigTLSWiderspruch`).

**Übernimmt (Start-Befehl):** Kommandoauswahl, Signalbehandlung und Exit-Codes liegen im Adapter `src/adapters/driving/cli` ([slice-cli-adapter](../done/slice-cli-adapter.md), `SPEC-138`); `src/bootstrap/main/` setzt `Konfig.Start` (bis dahin `nil`, das Kommando `start` ist unbekannt) auf `bootstrap.Start` (`SPEC-139` Nr. 1, 8). Die Pflichten des Starts gegenüber dem Adapter (`SPEC-138` Nr. 8 a bis f): Rückgabe statt `os.Exit`; das Herunterfahren läuft unter `context.WithoutCancel(ctx)` (sonst bricht `Drain` die Exchanges beim ersten Signal ab, `SPEC-121` Nr. 13); ist der Kontext vor oder in einem Startschritt abgebrochen, öffnet der Root nichts mehr und fährt das Geöffnete herunter (Rückgabe 0); ein Konfigurationsfehler bleibt 1; ein Fehler im Start, im Lauf von `Serve` oder im Herunterfahren ergibt 1 und gewinnt über den Abbruch (Nr. 8 f). Fänger: S1 bis S4.

**Übernimmt (Listener):** Der Root bindet den Kern des HTTP-Adapters (`slice-http-adapter`) mit `Konfig{Verarbeiten: rec.Record (Modus record) bzw. rep.Replay (Modus replay), IDs, Listener}` und ruft `Serve` in einer Goroutine; `rec` ist dieselbe Instanz wie in `session.Konfig.Record` (`SPEC-124`). Der Herunterfahr-Ablauf (`Anhalten`, `CloseAll`, `Schliessen(rest)`) kommt von `slice-http-lebenszyklus` (Sendung, `AGENTS.md` §3.13; `SPEC-136` Nr. 15); der Root ruft `Schliessen` genau einmal je Prozess, denn ein `Schliessen` gleichzeitig mit einem laufenden ist in `SPEC-136` *Ungeprüft* nicht zugesagt (`SPEC-139` Nr. 7; Sendung von `slice-http-lebenszyklus`). Fänger: B5, B6.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Start im Modus `mock` oder `hybrid` — scheitert mit `ODR-E2001`, bis die Folgewellen die Use Cases liefern (Entscheidung steht in `slice-fehlermodell-konfiguration`).
- CLI-Kommandos über Start und Version hinaus — entfallen (Entscheidung des Users vom 2026-10-30, Welle §6): das Lastenheft fordert keine, keine Folgewelle trägt sie. Admin-Listener und Admin-API — Folgewelle. Dieser Slice liefert deshalb keinen Weg für einen Anwender, eine Aufzeichnung zu beenden, zu importieren oder einen Workspace anzulegen; Abspielen im Betrieb ist verdrahtet, aber ohne Import und Workspace nicht erreichbar.
- Referenzdienst als Fixture, `tests/` im Quellstand der Stufe `source` und im Build-Kontext (`Dockerfile`, `.dockerignore`; Sendung von `slice-race-gate`), Verdrahtung von `baselineimport.New` und `workspace.New` samt Zugängen `Import()` und `Workspaces()` am `Handle`, der M2-Test mit seinen Mutationen und die Zeilen zum Import (Regeln gleich, Wächter am Import, `MetadataHash`, Coverage) — `slice-m2-referenzdienst`: der Schnitt trennt dort, wo kein Mechanismus geteilt wird (Herkunft oben); der Nehmer führt die Punkte in seinem §1 und seiner DoD. Der M2-Beleg liegt damit nicht in diesem Slice; die Zelle von M2 setzt die Welle-Closure.
- Lasttest zu `LH-QA-01` — braucht Telemetrie und ein Messprofil (Welle §6); bleibt ausdrücklich ungeprüft.
- Das gebaute Binary im Container, `make test-integration` (`SPEC-037`) — [slice-test-integration](../open/slice-test-integration.md), nach diesem Slice; hier fährt der Test den Pfad im Prozess von `go test`, nicht das Binary. Dass `main` `Start` setzt, fängt erst dieser Lauf am gebauten Binary (`SPEC-138` *Ungeprüft*, §6).
- Vergleichstest zu `LH-QA-07` (Direktzugriff gegen Proxy) — `slice-protokolltreue`, nach `slice-m2-referenzdienst`; er nutzt Referenzdienst und `Handle`.

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

- [ ] Verdrahtung und Lebenszyklus nach `SPEC-139`: `bootstrap.Start`, `Build`, `Handle` mit `Hochfahren`, `Herunterfahren`, `Adresse`, `Fehler`, `Wartung`, `Record`, `Sessions`; `src/bootstrap/main/` setzt `Konfig.Start` (der Start-Befehl des Adapters `cli` aus `slice-cli-adapter` ruft `LadeOderMelde`, Exit-Code 1 und Meldung nach `SPEC-110`, `mock`/`hybrid` scheitern mit `ODR-E2001`; Zusage von `slice-fehlermodell-konfiguration`); Ablage `odata-recorder.db` und `blobs` in `DATA_DIR`; Reihenfolge Ablage → `MaintenanceUseCase` → Services anlegen → erste Session (nur `record`) → Listener → `Serve`, ein Wächter vor dem Blob-Port für Pipeline und Wartung, ein Use Case je Modus am Listener (Sendungen von `slice-session-wartung`, `slice-session-use-case`, `slice-odata-codec-csdl`); Herunterfahren in der Reihenfolge `Anhalten` → `CloseAll` (`SHUTDOWN_TIMEOUT`, nur `record`) → `Schliessen(rest)` genau einmal → Warten auf `Serve` → Ablage, unter `context.WithoutCancel(ctx)`, ein Fehler hält die späteren Schritte nicht auf (Sendungen von `slice-session-beenden` und `slice-http-lebenszyklus`, `SPEC-124` Nr. 14, `SPEC-136` Nr. 15); `Pruefen` aus dem Wert des Starts und die Warnung `start: Zertifikatsprüfung zum Upstream abgeschaltet` bei `record` und abgeschalteter Prüfung (`LH-RB-02.a`, Sendungen von `slice-upstream-adapter` und `slice-upstream-tls`); jeder Wert der Konfiguration erreicht seinen Verbraucher, eine Uhr, eine ID-Quelle, ein Logger (`slice-clock-idgenerator`).
- [ ] Tests und Mutationen: die Tabelle *Zusage · Fänger* (§3, S1 bis S4 und B1 bis B8) mit je einer Mutation, vom Implementer rot gesehen, die Raten der nebenläufigen Fänger mit Zahl im Absatz von §3 (`AGENTS.md` §3.10); der Durchstich (Start, Request durch den HTTP-Adapter an einen `httptest`-Upstream mit Haltepunkt, Herunterfahren mit Session `complete`) läuft in `make test` und in `make test-race`; jeder nebenläufige Fänger vor der Übergabe mit `-count=50`.
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
| `src/bootstrap/` | update | Verdrahtung, `Build`, `Handle`, `Hochfahren`, `Herunterfahren`, `Start`; Tests der Fänger (Paket `bootstrap`, Teil von `make test` und `make test-race`) |
| `src/bootstrap/export_test.go` | neu | die Naht `bauteile` für die Verdrahtungs-Fänger (unexportiert, nur im Test gesetzt) |
| `src/bootstrap/main/` | update | setzt `Konfig.Start` des Adapters `cli` (`slice-cli-adapter`); `LadeOderMelde`, Exit-Codes (Zusage von `slice-fehlermodell-konfiguration`) |
| `spec/spezifikation.md` | geliefert | `SPEC-139`, `SPEC-138` Nr. 8 e und f, `SPEC-136` Nr. 12 liegen vor dem ersten Code-Commit (Architect, durch den Planner übernommen) |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht hier vor dem ersten Code-Commit, das Raster *Operation × Argumentart* (`AGENTS.md` §3.16) in §6; der Vertrag liegt in `SPEC-139` (`AGENTS.md` §3.12). Der Fänger-Aufbau: Lebenszyklus-Fänger bauen den `Handle` als Literal mit Fakes (Wartung, Dienste, Sessions, Server/Listener, Ablage-Schließer mit gemeinsamem Ereignisprotokoll, eine Fake-Uhr aus `supporttest`); Verdrahtungs-Fänger lesen die Konfigs über die Naht `bauteile`; Durchstich-Fänger laufen mit echten Adaptern gegen `t.TempDir()` und einen `httptest`-Upstream mit Haltepunkt. Zählung der Größe (§4): S1 und S2 nebenläufig, S3, S4 und B1 bis B8 andere.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| S1 | Das Herunterfahren läuft unter `context.WithoutCancel(ctx)`; ein laufender Exchange wird nach dem Abbruch von `ctx` nicht abgebrochen, sondern bis `SHUTDOWN_TIMEOUT` abgewartet (`SPEC-138` Nr. 8 b, `SPEC-139` Nr. 7); `CloseAll` wartet auf diesen Exchange, weil `Record` dieselbe Instanz ist (`SPEC-139` Nr. 3); ein Request zwischen `Anhalten` und `CloseAll` erhält 503 `ODR-E5010`, nicht `ODR-E5009` | `TestHerunterfahrenOhneAbbruchDesAusloesers`: echte Adapter, Modus `record`; `Start(ctx, …)`, der Upstream hält die Antwort; der Test bricht `ctx` ab, sendet einen zweiten Request auf einer zweiten Verbindung (erwartet 503 `ODR-E5010`, die Verbindung zum ersten besteht), gibt die Antwort frei: Exchange angehängt, Session `complete`, Rückgabe 0 | (1) der Root reicht `ctx` an `CloseAll`: Exchange `incomplete`; (2) `CloseAll` vor `Anhalten`: der zweite Request erhält `ODR-E5009`; (3) `SessionUseCase` erhält eine zweite Instanz von `record.New`: `CloseAll` kehrt sofort zurück, der Exchange ist nicht angehängt | nebenläufig |
| S2 | Ist `ctx` vor oder in einem Startschritt abgebrochen, öffnet der Root nichts Weiteres und liefert 0 (`SPEC-138` Nr. 8 c, `SPEC-139` Nr. 6, 8). Menge: Schritt {Ablage öffnen, Wartung, Service anlegen, erste Session, Listener binden} × Modus; `replay` ohne erste Session: 5 + 4 = 9 | `TestAbbruchInJederPhase`: der Fake des Schritts blockiert auf `ctx.Done()` und liefert `ctx.Err()`; danach: `Dial` auf `LISTEN` scheitert, die Zahl der Sessions und Services ist die des Schritts davor, die Ablage ist geschlossen, Rückgabe 0 | ein Schritt prüft den Kontext nicht (öffnet den Listener trotzdem); je Schritt eine Mutation | nebenläufig |
| S3 | Ein Konfigurationsfehler liefert 1 mit der Meldung, auch bei abgebrochenem Kontext (`SPEC-138` Nr. 8 d). Menge: Kontext {offen, abgebrochen} × Verstoß {`MODE=mock`, ungültiger Schlüssel} = 4 | `TestKonfigFehlerBeiAbbruch`: Rückgabe 1, `stderr` trägt `ODR-E2001 …` | `Start` prüft den Kontext vor dem Laden und liefert 0 | andere |
| S4 | Scheitert ein Startschritt, fällt `Serve` im Lauf aus oder scheitert das Herunterfahren, fährt der Root das Geöffnete herunter, liefert 1 und schreibt genau eine Zeile ohne Wert (`SPEC-138` Nr. 8 e, f; `SPEC-139` Nr. 7, 8). Menge: Phase {Ablage öffnen, Wartung, Service anlegen, erste Session, Listener binden, `Serve`, `CloseAll`, Ablage schließen} × Kontext {offen, nach Beginn der Phase abgebrochen}; Ablage öffnen nur mit offenem Kontext (sonst S2); `replay` ohne erste Session und `CloseAll`: 15 + 11 = 26 | `TestStartFehler`: der Fake der Phase liefert einen Fehler (bei abgebrochenem Kontext einen anderen als `ctx.Err()`); das Geöffnete ist geschlossen (Zähler der Fakes: Listener nur wenn gebunden, `CloseAll` nur wenn eine erste Session begonnen wurde, Ablage immer), Rückgabe 1, `stderr` hat eine Zeile in der Form von Nr. 8 und enthält keinen Wert der Konfiguration | Rückgabe 0 trotz Fehler; bei abgebrochenem Kontext 0 statt 1 (Abbruch gewinnt); Geöffnetes bleibt offen; `CloseAll` ohne begonnene Session; ein Wert im `stderr` | andere |
| B1 | Reihenfolge des Starts: Ablage, Wartung, Service anlegen (je Dienst), erste Session (je Dienst, nur `record`), Listener binden, `Serve` (`SPEC-139` Nr. 6; `SPEC-123` Nr. 3; `SPEC-124` Nr. 5; `SPEC-107` Nr. 14) | `TestStartReihenfolge`: Ereignisprotokoll der Fakes in beiden Modi, Vergleich mit der Folge | je Paar benachbarter Schritte vertauscht (Schleife über die Paare): rot | andere |
| B2 | Services anlegen und Neustart (`SPEC-107` Nr. 14, `SPEC-139` Nr. 2, 6, 10). Menge: Dienst {unbekannt → `Create` mit `4.0`, leeren Metadaten; bekannt mit gleichem Root und gleicher Version → nichts, kein Fehler; anderer Root → `ODR-E2001`, keine Session angelegt; andere Version → `ODR-E2001`, keine Session; nicht genannter Eintrag bleibt unberührt; `Create` mit anderem Fehler als `ErrBereitsVorhanden` → Fehler} und der Neustart gegen dieselbe Ablage | `TestDiensteAnlegen` mit echter Ablage: Start, Herunterfahren, zweiter Start im selben `DATA_DIR`: der Service steht einmal, die Session des ersten Laufs ist `complete`, eine neue ist `open`; im Datenverzeichnis liegen `odata-recorder.db` und `blobs`; je Fall der Menge der Eintrag in `MetadataRepository` | das Anlegen entfällt (erste Session `ODR-E5005`); abweichender Root wird still übernommen; `ErrBereitsVorhanden` wird als Fehler gemeldet (zweiter Start scheitert); Anlegen vor der Wartung; anderer Dateiname der Ablage | andere |
| B3 | Genau ein Wächter je Prozess ist `BlobRepository` der Capture-Pipeline, des Replay-Dienstes und `Blobs` der Wartung (`SPEC-123` Nr. 1, 9; `SPEC-139` Nr. 3) | `TestEinWaechter`: über die Naht `bauteile` (§3) gelesen: die drei Konfigs tragen denselben Zeiger vom Typ `*maintenance.BlobWaechter`, und er umhüllt den Blob-Adapter | der Root gibt der Pipeline das darunterliegende Repository; der Root erzeugt zwei Wächter | andere |
| B4 | `Pruefen` ist `UpstreamTLSVerify`; der Nullwert bleibt nie stehen; die Warnung `start: Zertifikatsprüfung zum Upstream abgeschaltet` steht genau bei `record` und abgeschalteter Prüfung (`SPEC-139` Nr. 4, 9; `SPEC-120` Nr. 1; `LH-RB-02.a`). Menge: `UPSTREAM_TLS_INSECURE_SKIP_VERIFY` {nicht gesetzt, `true`} × Modus {`record`, `replay`} = 4 | `TestPruefenUndWarnung`: Naht `httpupstream.Konfig` (nur in `record` gebaut) trägt `Pruefen` gleich dem Wert; das Log trägt den WARN-Eintrag mit `ereignis=tls_pruefung_aus` in genau dem einen Fall | `Pruefen` immer falsch; immer wahr; die Warnung auch bei Prüfung an; die Warnung auch in `replay`; die Warnung fehlt | andere |
| B5 | Reihenfolge des Herunterfahrens und genau ein `Schliessen` (`SPEC-136` Nr. 15, `SPEC-139` Nr. 7): `Anhalten`, `CloseAll` (nur `record`), `Schliessen`, Warten auf `Serve`, Ablage; wiederholter und gleichzeitiger Aufruf ergibt dasselbe Ergebnis und ruft nichts erneut. Menge: Modus {`record`, `replay`} × Aufruf {einmal, zweimal nacheinander, zweimal gleichzeitig} | `TestHerunterfahrenReihenfolge`: Ereignisprotokoll und Zähler am Server-Fake (`Schliessen` genau 1) | `Schliessen` zweimal; Ablage vor `Schliessen`; `CloseAll` in `replay`; `Schliessen` vor `CloseAll`; ein Fehler von `CloseAll` hält die Ablage nicht zu | andere |
| B6 | `CloseAll` erhält `SHUTDOWN_TIMEOUT` als Frist, `Schliessen` den Rest `SHUTDOWN_TIMEOUT` minus Dauer seit `Anhalten`, nicht unter 0; in `replay` den vollen Wert (`SPEC-136` Nr. 15, `SPEC-139` Nr. 7). Menge: Dauer von `CloseAll` {0, kleiner als die Frist, gleich der Frist, größer} × Modus: 4 + 1 | `TestFristRest`: Fake-Uhr springt im Fake von `CloseAll`; der Server-Fake hält die Frist von `Schliessen` | `Schliessen` mit voller Frist; `rest` negativ durchgereicht; Frist von `CloseAll` ungleich `SHUTDOWN_TIMEOUT`; Messung mit `time.Now` statt der einen Uhr | andere |
| B7 | Jeder Wert der Konfiguration erreicht seinen Verbraucher (`SPEC-139` Nr. 4); genau eine Uhr, eine ID-Quelle, ein Logger in allen Verbrauchern (`SPEC-114` Nr. 2); `Regeln` an Record und Replay gleich | `TestWerteErreichenVerbraucher`: Schleife über die Felder der Konfiguration (jedes bekommt einen eigenen Wert) und die Verbraucher (Naht `bauteile` (§3): `record.Konfig`, `replay.Konfig`, `capture.Konfig`, `session.Konfig`, `maintenance.Konfig`, `httpupstream.Konfig`, `sqlite.Options`); Zeigergleichheit von Uhr, IDs, Logger über alle | ein Wert vertauscht oder nicht durchgereicht (je Feld); zwei Uhren; ein zweiter Logger; `Regeln` von Replay aus anderen Feldern | andere |
| B8 | Der Modus wählt den Dienst am Listener; `mock`, `hybrid` und leer erreichen `Build` nie, und `Build` weist sie ab (`SPEC-139` Nr. 2, 5) | `TestModus`: echte Adapter; `record`: ein Request erreicht den Upstream-Haltepunkt-Server und erhält seine Antwort, Exchange angehängt; `replay`: ein Request ohne Header erhält `ODR-E5003`, der Upstream sieht keine Verbindung (Zähler 0); `Build` mit `mock`, `hybrid`, leer und mit leeren Diensten: `ErrUngueltig`, Ablage geschlossen | die Modi am Listener vertauscht; `replay` ruft den Recorder; `Build` nimmt `mock` an | andere |

Hinweis zu B2 und `AGENTS.md` §3.15: "Allaussagen" (jedes Feld, jeder Schritt, jedes Paar) sind Schleifen über beide Mengen (B1, B7, S2, S4), nicht Listen. Zu §3.10: die Raten der nebenläufigen Fänger (S1, S2) misst der Implementer je Mutation und trägt sie mit Zahl in §3 ein; S1 (3) und S2 (je Schritt) färben in jedem Lauf rot, weil die Haltepunkte mit Kanälen ordnen. Das Fehlen von `Anhalten` vor `CloseAll` hat zusätzlich einen Wirkungs-Fänger (S1, zweiter Request), kein Rennen.

**Raten und Mutationen (Implementer, `AGENTS.md` §3.10):** *Der Implementer trägt hier vor der Übergabe die Tabelle Zusage · Mutation · roter Test mit den Raten je Mutation der Zeilen S1 und S2 ein (rot in N von N Prozessläufen mit `-test.count=1` unter `make test-race`, N Prozesse, Messender).*

**Schnittstelle von Start und Build, die Naht der Tests** (Entscheidung des Architect):

- **Trägt `Build` den Start? Nein, nur Verdrahtung und die Ablage.** Der Ablauf liegt in `Hochfahren` und `Herunterfahren`; `Start` ist die Hülle (Laden, Build, Hochfahren, Warten, Herunterfahren, Exit-Code). Grund: Tests mit Phasen-Fakes brauchen die Schritte einzeln, und `tests/` (Folge-Slice) braucht einen Handle ohne Signal und ohne Umgebung.
- **Exportiert?** Ja: `Build` und `Handle` (`SPEC-139` Nr. 1). Das schließt die erste Zeile von §6: `tests/` liegt außerhalb von `src/bootstrap`, ein internes Test-Paket erreichte `tests/integration/` nie, und die Admin-Adapter der Folgewellen holen ihre Use Cases am selben `Handle`. Die Oberfläche ist klein (neun Zeilen); `a-check` ändert das nicht (`src/bootstrap/**` ist `composition_root`).
- **Tests mit injiziertem Kontext und Fake-Konfig:** `Build(ctx, Konfiguration{…Literal…}, log)`; `DataDir` ist `t.TempDir()`, `Listen` ist `127.0.0.1:0` (`Adresse()` liefert den Port), `ShutdownTimeout` ein Wert in Millisekunden, `Dienste` zeigen auf einen `httptest`-Upstream. Der Kontext ist der des Tests mit `context.WithCancel`; der Test löst den Abbruch aus, nicht ein Signal.
- **Naht für Fänger, unexportiert** (Datei `src/bootstrap/export_test.go` setzt sie; der Betrieb nutzt die Voreinstellung): `bauteile{ listen func(network, addr string) (net.Listener, error); neuerServer func(httpodata.Konfig) (server, error); neueWartung/neueSession/neuerRecorder/neuerReplay/neuePipeline/neuerUpstream func(…Konfig) (…, error) }`, jede Voreinstellung der Konstruktor selbst, jeder Haken legt die übergebene Konfig ab und ruft den Konstruktor. `server` ist die Schnittstelle `Serve() error; Anhalten(); Schliessen(time.Duration) error`. Lebenszyklus-Fänger bauen den `Handle` als Literal mit Fakes dieser Schnittstellen und brauchen die Naht nicht. Das sind die zwei Port-Fakes (Phasen-Fakes, Server/Listener) plus der Test-Server (Upstream).
- **Wo der Testcode liegt:** unter `src/bootstrap/` (Paket `bootstrap`, Teil von `make test` und `make test-race`, weil `src/` im Quellstand der Stufe `source` liegt). Nichts davon braucht `Dockerfile` oder `.dockerignore`.


## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-http-adapter` und `slice-http-lebenszyklus` und `slice-cli-adapter` und `slice-session-wartung` und `slice-session-use-case` und `slice-session-beenden` und `slice-upstream-adapter` und `slice-upstream-tls` liegen in `done/` (`slice-baseline-import` und `slice-workspace-use-case` verlangt `slice-m2-referenzdienst`, nicht dieser Slice). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): `SPEC-139`, `SPEC-138` Nr. 8 e und f und `SPEC-136` Nr. 12 liegen im Commit `ff8eaeb`.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <basis>..<letzter Commit des Slice>` über die `pfad`-Zeilen des Blocks, Tests eingeschlossen; `basis` ist der Commit, der Spezifikation, Plan und Marke vor dem ersten Code-Commit abschließt. Zählung nach der Tabelle in §3: *a* = 2 nebenläufige (S1, S2) und *b* = 10 andere Zeilen (S3, S4, B1 bis B8), *f* = 3 (ein Phasen-Fake für Wartung, Dienste, Sessions und Ablage mit gemeinsamem Ereignisprotokoll, ein Server/Listener-Fake, ein Upstream-Test-Server mit Haltepunkt), *p* = 550 (Erwartung des Architect: `Build` samt Wertefluss und Hilfen rund 230, `Hochfahren` 90, `Herunterfahren` 70, `Start` mit Meldungen 90, Naht, Fehlerformen und `main` 70; an keinem vierten Slice gemessen). Schätzung S = 85·2 + 80·10 + 200·3 + 550 = 2120, Marke 2900 nach `SPEC-128` Nr. 1. Keine Schnittfrage: S liegt unter 2223 (Spielraum 103 Zeilen: ein Produktcode über 653 Zeilen oder ein vierter Fake kippt sie; die Antwort ist dann schon gegeben, der M2-Teil ist geschnitten, der Rest teilt `Handle` und Naht und wird nicht weiter geteilt).

```groessenmarke
pfad: src/bootstrap/
produktcode: 550
fakes: 3
basis: e8c027c
schaetzung: 2120
marke: 2900
ersetzt: 1100 | die ältere Marke aus der Prosa (über 1100 Zeilen, Schätzung rund 800) galt dem Slice samt M2-Beleg ohne die Formel von SPEC-128; der Schnitt nimmt Referenzdienst, Import, Workspace und M2-Test heraus, die Tabelle wuchs um B1 bis B8, S = 2120 ergibt 2900
```

`basis` ist der Commit, der Spezifikation, Plan und Marke abschließt; die Marke ist ab dem ersten Code-Commit fest (`SPEC-128` Nr. 5 a).

**Messpunkt beim Handoff (Implementer, `make groessenmarke`):** *wird vom Implementer eingetragen (Zeile `Messpunkt …`, Produktcode von Hand, Auslastung).*

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 2900 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Es gibt keinen vorab benannten Teil, den der Handoff abtrennen dürfte (der Rest teilt `Handle` und Naht); nach dem Code ist eine Überschreitung ein Eintrag `ueberschreitung` mit Grund im Block, keine Rückführung (`SPEC-128` Nr. 10 c, `AGENTS.md` §3.14). Die Rückführung gilt damit vor dem ersten Code-Commit, wenn Tabelle oder Produktcode die Marke über 2900 heben.
- `in-progress` → `open` (blockiert): eine Randform von `SPEC-139` fehlt oder widerspricht, und der Architect entscheidet sie nicht vor dem Code (Rückgabe an den Architect, `AGENTS.md` §3.12).

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

- Aufrufweg zu Session-Ende, Import und Workspace ohne Admin-API — **Ausgang:** entschieden (Architect): `Build` und `Handle` sind exportiert (`SPEC-139` Nr. 1), weil `tests/` außerhalb von `src/bootstrap` liegt und die Admin-Adapter der Folgewellen ihre Use Cases am selben `Handle` abholen; Import und Workspace kommen additiv mit `slice-m2-referenzdienst`.
- Der Test deckt alle Schichten und bricht, wo ein Vorgänger-Slice eine Randform offen ließ — **Ausgang:** weiter offen; der Fund geht als Befund an den Slice, der ihn verantwortet, nicht in den Test; die Nehmer liegen in `done/`, der Befund geht an den Planner (`AGENTS.md` §3.13). Der Durchstich dieses Slice ist klein (ein Upstream im Testcode); den Lauf mit den echten Lesern von Import und Codec trägt `slice-m2-referenzdienst`, dessen §6 den Punkt führt.
- Loopback und Ports im Testcontainer (`--network none`): `httptest` und `net.Listen` auf `127.0.0.1:0` — **Ausgang:** offen; der Implementer erprobt es vor dem Code in `make test` und trägt das Ergebnis ein; `slice-m2-referenzdienst` übernimmt es.
- Zeitabhängigkeit: Herunterfahren, Fristen, parallele Clients — **Ausgang:** offen; S1 und S2 ordnen mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8 b), und laufen vor der Übergabe mit `-count=50`; `make test-race` fährt `src/bootstrap` (`SPEC-122` Nr. 2). Die Fake-Uhr misst die Frist; `time.Now` im Root ist die Mutation von B6.
- M2 ist ein Meilenstein, kein Slice-Ziel — **Ausgang:** entfallen für diesen Slice: der Beleg wandert in `slice-m2-referenzdienst`; die Roadmap setzt die Zelle erst nach der Closure der Welle auf *erreicht*.
- Raster *Operation × Argumentart* (`AGENTS.md` §3.12, §3.16), Ort je Zelle in `SPEC-139`; „gilt nicht“ steht mit Grund — **Ausgang:** entschieden (Architect), Raster unten; eine Zelle ohne Ort hält den Implementer an (§3.12).

| Operation | fehlender / leerer Wert | unbekannter / fremder Wert | Abbruch (`ctx`) | Wiederholung / Nebenläufigkeit | Modus |
|---|---|---|---|---|---|
| `Build` | `log` nil verwirft, `Dienste` leer: `ErrUngueltig`, `DataDir` leer: `ErrPfad` (`SPEC-139` Nr. 1, 2) | Modus `mock`/`hybrid`/leer: `ErrUngueltig` (Nr. 2, 5); `DATA_DIR` fehlt oder Datei statt Verzeichnis, `odata-recorder.db` ist Verzeichnis, Datei beschädigt: Fehler der Ablage, abgebildet (Nr. 2; `SPEC-103` Nr. 4 bis 7) | `Open` liefert den Kontextfehler, nichts bleibt offen (Nr. 2; `SPEC-138` Nr. 8 c) | zweiter `Build` auf demselben `DATA_DIR` gleichzeitig: nicht unterstützt (Nr. 2, `SPEC-106`); nach `Herunterfahren`: Neustart (Nr. 10) | `record` baut Recorder, Pipeline, Upstream-Client, `SessionUseCase`; `replay` nicht (Nr. 3) |
| `Hochfahren` | Dienst unbekannt in der Ablage: wird angelegt (Nr. 6; `SPEC-107` Nr. 14) | gespeicherter Service mit anderem Root oder anderer Version: `ODR-E2001`, keine Session (Nr. 6, 8 c); `LISTEN` nicht bindbar: `ODR-E2001 LISTEN` (Nr. 8 c); Wartung scheitert (`ODR-E3001`): kein Listener (Nr. 6) | vor und in jedem der vier Schritte: nichts Weiteres, Kontextfehler (Nr. 6; `SPEC-138` Nr. 8 c) | zweites `Hochfahren`: `ErrUngueltig` (Nr. 6) | erste Session nur in `record` (Nr. 6 (3); `SPEC-124` Nr. 5) |
| `Herunterfahren` | ohne `Hochfahren`: nur Ablage schließen (Nr. 7) | `CloseAll`-Fehler, `Fehlgeschlagen` > 0, `ErrFristAbgelaufen`, Fehler beim Schließen der Ablage: gesammelt, später Schritte laufen (Nr. 7) | `ctx` ohne Wirkung, `WithoutCancel` (Nr. 7; `SPEC-138` Nr. 8 b) | einmal, wiederholt und gleichzeitig: dasselbe Ergebnis (Nr. 7); `rest` < 0 wird 0, `SHUTDOWN_TIMEOUT` ist nie 0 (`SPEC-110` Nr. 4), `Schliessen(0)` ist zulässig (Nr. 7; `SPEC-136` Nr. 14) | `record`: Schritt (2) mit `CloseAll`; `replay`: ohne (2), `rest` gleich `SHUTDOWN_TIMEOUT` (Nr. 7; `SPEC-136` Nr. 15) |
| `Start` | gilt nicht: die Konfiguration liefert `LadeOderMelde` (`SPEC-110` Nr. 9: fehlende `SERVICES_FILE` oder fehlende Dienste sind `ODR-E2001`, Exit 1; der Root prüft davon nichts erneut) | ungültige Konfiguration: 1, auch bei abgebrochenem `ctx` (Nr. 8 a; `SPEC-138` Nr. 8 d) | Kontext vor oder in einem Schritt abgebrochen: 0 (Nr. 8 b; `SPEC-138` Nr. 8 c); gleichzeitiger anderer Fehler: 1 (`SPEC-138` Nr. 8 f); Signal im Herunterfahren: ohne Wirkung (Nr. 7) | gilt nicht: der Adapter ruft `Start` einmal (`SPEC-138` Nr. 9) | `record`: 5 Schritte plus `CloseAll`; `replay`: 4 plus Serve (Nr. 6, 7, 8) |
| `Fehler()`, `Adresse()` | vor `Hochfahren`: `Adresse` leer, `Fehler` ohne Wert (Nr. 6) | gilt nicht: keine Argumente | gilt nicht: keine Argumente | beliebig oft, gleichzeitig (Nr. 10) | `Record()`/`Sessions()` im Modus `replay`: nil (Nr. 3) |

- Randformen (`AGENTS.md` §3.12): das Raster oben ersetzt die Aufzählung — **Ausgang:** entschieden durch `SPEC-139` vor dem Code.
- `DATA_DIR` in `Build` (`SPEC-139` Nr. 2): `Build` mit Literal-Konfiguration legt das Verzeichnis nicht an, der Start legt es über `SPEC-110` Nr. 11 an — **Ausgang:** entschieden (Architect), stimmig; für die Tests ein Setup-Hinweis: `t.TempDir()` existiert bereits.
- Wer die konfigurierten Services im `MetadataRepository` anlegt, bevor der Root `OpenInitialSession` ruft (Sendung von [slice-session-use-case](../done/slice-session-use-case.md), `AGENTS.md` §3.13) — **Ausgang:** entschieden: der Root legt sie an, mit allen Randformen in `SPEC-107` Nr. 14 und `SPEC-139` Nr. 6; Fänger B2.
- `.a-check.yml`: `ports-support` ohne `direction` (`ARC-021`) — **Ausgang:** entschieden (Architect): die Klarstellung ist keine Lockerung, keine ADR, kein Folge-Slice. `ARC-021` lässt jede Schicht die Querschnitts-Ports nutzen; die Kanten von `services`, `adapters-driving`, `adapters-driven` und `ports-*` zu `ports-support` stehen einzeln in `edges`, und `ports-support` hat keine Kante zu `ports-inbound` oder `ports-outbound`. Was ohne `direction` entfällt, ist allein die Prüfung `port-direction-mismatch` für Kanten zu einer Schicht, deren Richtung nach `ARC-021` keine ist; ein Import von `ports/outbound` aus `httpodata` bleibt über `ports-outbound` ein Befund (`AGENTS.md` §3.6 trifft Senkungen, hier gilt die Regel strenger oder gleich). Akzeptiertes Negativ: die fehlende Gegenprobe zu diesem Fall; Grund und Folge stehen in `slice-http-adapter` §6, keine neue Pflicht.
- Längen-Differenz zwischen Filter und Schlüssel (`SPEC-161`, `SPEC-164`; 1641 Abweichungen in 167634 Läufen) — **Ausgang:** entschieden, vertraglich getragen und kein Fehler: der Root liest keine OData-URL, `record` und `replay` bilden den Match-Key aus `SPEC-116`; alle 1641 Abweichungen sind `literal` im Filter gegen `zu_lang` im Schlüssel und erklärt durch die Regel „der Rest hinter dem schließenden `'` zählt nicht zur Länge“ (`SPEC-164` Nr. 4 a; Verifikation `docs/reviews/2026-10-31-slice-odata-url-filter-verifikation.md`, Stufe c). Eine Angleichung wäre eine Änderung des Vertrags des Lesers; wer sie will, entscheidet sie in `slice-query-interpreter`, dessen Lauf beide Pfade verwendet (Adresse, im selben Commit in dessen §1 eingetragen); ohne Beschluss dort bleibt die Regel, keine Folgepflicht.
- `SPEC-136` Nr. 12, Rückgabe von `Serve` — **Ausgang:** entschieden durch die Berichtigung in `SPEC-136` Nr. 12 (Commit `ff8eaeb`); der Code stimmt mit dem Wortlaut überein (`merkListener.Accept`, `haltAnLocked`) und mit der Klasse L7 (`TestServeFehlerKlasse`, 35 Elemente) und ändert sich nicht. Die Randform (b), ein Fehler ohne Bezug zum Schließen, ist im Wortlaut entschieden und durch L7 Teilmenge A gefangen; die Randform (a) steht unten als akzeptiertes Negativ.
- *Ungeprüft* von `SPEC-124` (Verdrahtung des Beendens), `SPEC-123` (Wächter, Verdrahtung) und `SPEC-136` (`Schliessen` gleichzeitig mit `Schliessen`) — **Ausgang:** getragen von `SPEC-139` Nr. 3, 6, 7; die Fänger sind S1, B1, B3, B5, B6 (der Root trägt die Zusage genau eines `Schliessen`, der Adapter bleibt unverändert).
- Akzeptierte Negative (Grund und Folge; ohne Fänger, `AGENTS.md` §3.15) — **Ausgang:** akzeptiertes Negativ, je Punkt:
  1. Zwei Prozesse auf demselben `DATA_DIR` (`SPEC-106`, `SPEC-123` *Ungeprüft*, `SPEC-139` *Ungeprüft*). Grund: eine Sperrdatei ist plattformabhängig und eine eigene Zusage für ein Betreiberversehen. Folge: die Startwartung des zweiten Prozesses setzt die offene Session des ersten auf `incomplete`; der Betreiber startet einen Prozess je Datenverzeichnis. Das Handbuch (`slice-handbuch-record-replay`) soll den Satz „ein Prozess je Datenverzeichnis“ nennen (Adresse, im selben Commit in dessen §1 eingetragen); ohne ihn bleibt es bei `SPEC-106`, keine Pflicht.
  2. `REPLAY_ASSIGNMENT=ordered` wirkt wie `explicit`. Grund: die erste Antwort `ODR-E5003` sagt, was fehlt; eine Warnung oder Ablehnung in `SPEC-110` wäre ein Eingriff in einen abgeschlossenen Slice. Folge: ein Betreiber, der `ordered` setzt, sieht es nach dem ersten Request.
  3. `main` setzt `Start` (`SPEC-138` *Ungeprüft*): eine Zeile. Folge: ungefangen, bis `slice-test-integration` (`SPEC-037`) das gebaute Binary startet.
  4. Der Wortlaut der Logeinträge außer den zwei fest benannten (`SPEC-139` Nr. 8 c und 9) und die Dauer des Herunterfahrens über `SHUTDOWN_TIMEOUT` hinaus (`SPEC-124` *Ungeprüft*). Folge: Auswertungen suchen nur nach den festen Texten.
  5. `Herunterfahren` gleichzeitig mit `Hochfahren` (`SPEC-139` *Ungeprüft*). Grund: `Start` ruft sie nacheinander, kein anderer Aufrufer ist vorgesehen. Folge: nicht zugesagt; ein Test, der es tut, trägt die Folge.
  6. Ein `Listener` mit `Close` gleich `nil` nach äußerem Schließen (`SPEC-136` Nr. 12). Grund: die Unterscheidung wäre am Aufruf des Adapters nicht möglich, ohne den Typ des Listeners zu kennen. Folge: bei einem solchen Wrapper ist `Serve` `nil` statt Fehler; der Root übergibt `net.TCPListener` (liefert `net.ErrClosed`) und ist nicht betroffen.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, Ausgang `verkörpert` in `AGENTS.md` §3.10); der Test ist nebenläufig: das Gate fährt `-race` (`make test-race`, `SPEC-122`) über `src/bootstrap`, die nebenläufigen Fänger ordnen weiter mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8b), und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
