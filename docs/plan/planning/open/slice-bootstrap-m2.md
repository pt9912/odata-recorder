# Slice slice-bootstrap-m2: Composition Root, Referenzdienst und M2-Test

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-12`, `LH-FA-04`, `LH-FA-02`, `LH-FA-18`, `LH-QA-02`, `LH-RB-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `ARC-012` (Composition Root) · `LH-FA-12.a` · `LH-FA-02.c` · `SPEC-048` · `architecture.md` §2.5 (Modus-Routing), §4.9 · `SPEC-049` (Fixture-Rahmen des Referenzdienstes) · `SPEC-120` Nr. 1, `SPEC-110` Nr. 10, `LH-RB-02.a` (Verdrahtung des Upstream-Adapters, Warnung beim Start) · `SPEC-123` Nr. 1, 9 (Wächter vor dem Blob-Port) · `SPEC-127` Nr. 1 (Verdrahtung des Imports).
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

**Ziel:** Der Composition Root verdrahtet Konfiguration, Adapter, Services und Listener je Betriebsmodus (`record`, `replay`), fährt Startwartung, erste Session und geordnetes Herunterfahren, der Start-Befehl lädt die Konfiguration über `LadeOderMelde` (Exit-Code 1 und Meldung nach `SPEC-110` bei Verstoß; `mock` und `hybrid` scheitern mit `ODR-E2001`; übernommen von `slice-fehlermodell-konfiguration`, das Laden selbst liegt dort), und der M2-Test zeichnet über den Recorder gegen einen Referenzdienst auf und spielt ohne Upstream byte-gleich ab.

**Übernimmt:** Der Composition Root erzeugt genau eine `clock.NewSystemuhr()` und eine `clock.NewZufallsIDs()` je Prozess und reicht sie an alle Verbraucher; zwei Uhren wären zwei nicht aufeinander abgestimmte Höchstwerte (`SPEC-114` Nr. 2; Sendung von `slice-clock-idgenerator`).

**Übernimmt (Session und Wartung):** Der Composition Root verdrahtet `MaintenanceUseCase` und `SessionUseCase` (Konstruktoren und Ports nach den Verträgen aus `slice-session-wartung`, `slice-session-use-case` und `slice-session-beenden`): beim Start zuerst die Startwartung (`MaintenanceUseCase`, Sendung von `slice-session-wartung`), danach im Modus `record` `OpenInitialSession` je konfiguriertem Service (`ARC-012`; Sendung von `slice-session-use-case`), beim Herunterfahren `CloseAll` mit `SHUTDOWN_TIMEOUT` (`SessionUseCase`, `ARC-012`; Sendung von `slice-session-beenden`); die drei Slices müssen in `done/` liegen (§4 unten). `SessionUseCase` erhält denselben `RecordUseCase` (dieselbe Instanz von `record.New`) wie der HTTP-Adapter: `Drain` wirkt im Zustand des Dienstes, der die Aufnahme annimmt (`SPEC-121` Nr. 13: die Session nimmt „in diesem Dienst“ nichts mehr an), mit einer zweiten Instanz endet `CloseAll` ohne Wirkung auf die Annahme (Sendung von `slice-session-beenden`). Dazu der Fänger: am Handle `Build` läuft ein Exchange (der Referenzdienst hält die Antwort), `CloseAll` wartet auf ihn, der Exchange ist angehängt und die Session `complete`; Mutation: der Root gibt `SessionUseCase` eine zweite Instanz (`CloseAll` kehrt sofort zurück, der Exchange ist nicht angehängt, weil seine Session schon geschlossen ist). Die Zeile gehört in die Tabelle *Zusage · Fänger* (§3). Die Felder der Konfig von `SessionUseCase`, die der Root füllt, nennt dessen Vertrag: `Sessions`, `Dienste`, `Wartung` (derselbe `MaintenanceUseCase`, den der Root vor den Listenern anstößt), `Uhr`, `IDs`, `Log` und die Felder des Beendens `Record` (dieselbe Instanz von `record.New`, die der HTTP-Adapter ruft), `Exchanges` (das `RecordingRepository`) und `Frist` (der Wert von `SHUTDOWN_TIMEOUT`, nicht negativ; `SPEC-124` Nr. 1). **Herunterfahren (Sendung von `slice-session-beenden`, `AGENTS.md` §3.13; `SPEC-124` Nr. 14 und *Ungeprüft*, `LH-FA-02.c`):** der Root ruft `CloseAll` im Modus `record`, nachdem die Listener keinen neuen Request mehr annehmen (neue Requests 503) und bevor er ihre Verbindungen und die Repositories schließt; `CloseAll` wartet die laufenden Exchanges höchstens `SHUTDOWN_TIMEOUT` ab und bricht die überfälligen ab, der Listener wartet danach nicht noch einmal `SHUTDOWN_TIMEOUT` (sonst dauert das Herunterfahren die Summe beider Fristen), und ein Fehler von `CloseAll` oder `Fehlgeschlagen` größer `0` hält das Schließen der Repositories nicht auf (die Einträge schreibt der Dienst, die Sessions setzt der nächste Start auf `incomplete`). Der Root übergibt `CloseAll` den Kontext des Herunterfahrens, den der Prozess nicht vor der Rückkehr beendet (ein vorab beendeter Kontext lässt `FindOpen` scheitern und beendet keine Session, `SPEC-124` Nr. 14 b). Der Fänger gehört in die Tabelle *Zusage · Fänger* (§3): am Handle `Build` wartet `CloseAll` auf einen laufenden Exchange (oben) und das Herunterfahren hält die Reihenfolge Listener, `CloseAll`, Repositories. Aus `SPEC-124` Nr. 5 (Sendung von `slice-session-use-case`, `AGENTS.md` §3.13): der Root ruft `OpenInitialSession` nur im Modus `record` und nach der Startwartung, einmal je Service, und stellt vorher sicher, dass der Service im Speicher steht (`MetadataRepository`, `SPEC-107` Nr. 1; sonst `ODR-E5005`); wer die konfigurierten Services dort anlegt, bestimmt `SPEC-107` Nr. 14: der Root (siehe *Übernimmt (Services)* unten; `SPEC-123` Stufe 3 liest dieselbe Tabelle mit `Dienste.List`; Sendung von `slice-session-use-case`, `AGENTS.md` §3.13).

**Übernimmt (Services):** Der Composition Root legt jeden konfigurierten Service an, den der Speicher nicht kennt: nach der Startwartung und vor `OpenInitialSession`, einmal je Service, mit `MetadataRepository.Create` (`odata_version` `4.0`, Metadaten leer); ein gespeicherter Service mit anderem `service_root` oder anderer `odata_version` verhindert den Start (`ODR-E2001`), einer, den die Konfiguration nicht nennt, bleibt unberührt (`SPEC-107` Nr. 14; Sendung von `slice-odata-codec-csdl`, `AGENTS.md` §3.13, und Entscheidung der benannten Spec-Lücke aus `slice-session-use-case`). Die Metadaten setzt allein der Import (`slice-baseline-import`, `SPEC-107` Nr. 15); der Root ruft `SetMetadata` nie. Dazu die Zeilen der Tabelle *Zusage · Fänger* (§3): am Handle `Build` steht der Service der Konfiguration im Speicher, bevor die erste Session angelegt wird, und ein Start ohne ihn endet nicht in `ODR-E5005`; ein zweiter Start gegen dieselbe Datenbank legt ihn nicht neu an und läuft; ein Start gegen einen gespeicherten Service mit anderem Root endet in `ODR-E2001`, ohne dass eine Session angelegt ist. Mutationen: das Anlegen entfällt (`ODR-E5005`), das Anlegen vor der Startwartung statt danach (Ereignisprotokoll der Reihenfolge Startwartung, `Create`, `OpenInitialSession`), ein Fehler von `Create` bei bekanntem Service wird als Fehler gemeldet (der zweite Start scheitert), der abweichende Root wird still übernommen (der Start läuft).

**Übernimmt (Wächter):** Der Composition Root erzeugt genau einen `maintenance.NewBlobWaechter` je Prozess (mit `maintenance.StandardKarenz` und der einen Uhr) über dem Blob-Adapter und gibt ihn als `BlobRepository` jedem Schreiber von Blobs (die Capture-Pipeline des Record-Dienstes, den Import) und der Wartung als `Blobs`; wer das darunterliegende Repository erhält, ist ungeschützt (`SPEC-123` Nr. 1, 9, *Ungeprüft*; Sendung von `slice-session-wartung`). Dazu der Fänger, dass der Root den Wächter an jeden Schreiber gibt: am Handle `Build` committet die Pipeline (und der Import) einen Blob, der keine Referenz trägt; `CleanupBlobs` im Abstand von weniger als der Karenz entfernt ihn nicht (ohne offene Session), und ein `Commit` gleicher Bytes gegen den Abgleich endet in einer bestehenden Datei. Mutationen: der Root gibt einem der Schreiber das darunterliegende Repository (der Blob wird entfernt); der Root erzeugt zwei Wächter (der Blob wird entfernt). Die Zeile gehört in die Tabelle *Zusage · Fänger* (§3).

**Übernimmt (Import):** Der Composition Root verdrahtet `baselineimport.New` (`SPEC-127` Nr. 1; Sendung von `slice-baseline-import`, `AGENTS.md` §3.13) und gibt `ImportUseCase` über den Handle `Build` frei: `Dienste` (`ID` und `UpstreamURL` je konfiguriertem Service, dieselben Werte wie bei `record.New`), `Regeln` (**dieselben** `SchluesselRegeln` wie beim Record-Dienst: sonst passt der Key der Folgeseite zu keinem aufgezeichneten Exchange, und `complete` bleibt aus), `MaxCaptureBody` (`MAX_CAPTURE_BODY`), `Sessions`, `Exchanges`, `Metadaten`, `Snapshots`, `Imports` aus dem SQLite-Adapter, `Blobs` mit dem Wächter (*Übernimmt (Wächter)*: der Import schreibt das Manifest), `MetadatenCodec` aus `csdl.New()`, `DatenCodec` aus `jsoncodec.New()`, die eine Uhr und die eine `IDGenerator`-Instanz. Dazu der Fänger: am Handle `Build` importiert der M2-Test die beendete Session, danach trägt der Service `MetadataHash` gleich der Blob-Referenz des `$metadata` (`SPEC-107` Nr. 15), und die Coverage von `Products` ist `complete`, wenn der Referenzdienst die Collection ohne Link liefert; Mutation: der Root übergibt dem Import andere `Regeln` als dem Record-Dienst (der Fänger liest die Evidenz: Kette `seite_fehlt` statt `proven`, sobald der Referenzdienst eine zweite Seite liefert). Hinweis für den Referenzdienst: eine Antwort mit wechselnden Werten für dieselbe Entity innerhalb einer Session ist ein Konflikt und lässt den Import mit `ODR-E6101` enden (`SPEC-127` Nr. 10); der „wiederholte Request mit verschiedenen Antworten“ des Fixtures liegt deshalb auf einem Pfad, den der Import überspringt (etwa `/Products/$count`), nicht auf einer Entity oder Collection. Ebenso ist die Einzel-Entity des Fixtures eine der zwei Entities der Collection (gleicher Schlüssel, gleiche Werte): trägt die Session eine Entity, deren Schlüssel die belegte Kette nicht nannte, ist die Coverage von `Products` `partial` mit dem Grund `bestand_ausserhalb_kette` (`SPEC-127` Nr. 12 g) und der Fänger rot; das gilt für die Einzel-Entity wie für jede gefilterte Abfrage des Referenzdienstes (Sendung von `slice-baseline-import`, `AGENTS.md` §3.13). Die Zusammenarbeit des Imports mit den echten Lesern fängt allein dieser Test: die Suite von `slice-baseline-import` führt Fakes der Codecs, der Import des M2-Tests läuft mit `csdl.New()` und `jsoncodec.New()`, nicht mit Fakes, und sein Ergebnis (`complete` für `Products`, `MetadataHash`) ist der Beleg, dass Schlüsselform und Antwortform der Leser zum Import passen (Sendung von `slice-baseline-import`, `AGENTS.md` §3.13).

**Übernimmt (Record):** Der Composition Root verdrahtet `record.New` (`SPEC-121` Nr. 1) aus der Startkonfiguration (Dienste mit Root und `UpstreamURL`, `SchluesselRegeln`, `capture.Standard` mit `MAX_CAPTURE_BODY` und `REDACT_JSON_PATHS`) mit dem Client aus `httpupstream.New` als `UpstreamClient` und gibt dem Dienst genau einen `*slog.Logger` (`SPEC-121` Nr. 12; einen Port `Telemetry` gibt es in dieser Welle nicht). `CloseAll` beendet Sessions über `SessionUseCase`, das `Drain` ruft (Sendung von `slice-record-service`).

**Übernimmt (Upstream):** Der Composition Root verdrahtet `httpupstream.New` (`SPEC-120` Nr. 1): `Frist` aus `UPSTREAM_TIMEOUT`, `Pruefen` die wirksame Zertifikatsprüfung nach `SPEC-110` Nr. 10 (falsch genau bei `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true`), `CAPEM` der Inhalt von `UPSTREAM_CA_FILE` (leer ohne Datei); ist die Prüfung abgeschaltet, schreibt der Start eine Warnung ins Log (`LH-RB-02.a`). Der M2-Test ist der erste Lauf von Record-Dienst und Adapter zusammen (Sendung von `slice-upstream-adapter`). Die Wirkung von `Pruefen` und `CAPEM` im Adapter, nach der die Warnung fragt, liefert `slice-upstream-tls` (beim Handoff von `slice-upstream-adapter` geschnitten); die Sendung bleibt dieselbe, der Start verlangt beide Slices in `done/` (§4).
Dazu der Fänger, dass der Root `Pruefen` aus dem Wert des Starts übergibt und den Nullwert (falsch, keine Prüfung) nicht stehen lässt (Sendung von `slice-upstream-tls`, Befund I-1 seines Reviews, Entscheidung des Architect in `SPEC-120` Nr. 1): an dem Client, den der Root baut, prüft ohne beide Schlüssel (`UPSTREAM_TLS_VERIFY`, `UPSTREAM_TLS_INSECURE_SKIP_VERIFY`) der Austausch mit einem Upstream unbekannter Stelle (`ODR-E4001`, 502), und mit `UPSTREAM_TLS_INSECURE_SKIP_VERIFY=true` nicht (200). Mutationen: der Root übergibt immer falsch, der Root übergibt immer wahr; je eine färbt einen der beiden Fälle rot. Die Zeile gehört in die Tabelle *Zusage · Fänger* (§3), die vor dem Übergang `next` → `in-progress` entsteht; der Wert des Starts selbst ist in `src/bootstrap/konfig_test.go` gehalten (`TestKonfigDefaults`, `TestKonfigTLSWiderspruch`).

**Übernimmt (Quellstand):** Die Pakete unter `tests/` (Referenzdienst, M2-Test) liegen im Quellstand der Stufe `source` des `Dockerfile` und im Build-Kontext (`.dockerignore`), nicht erst in der Stufe `test`: `make test-race` fährt alle Pakete der Stufe `source` (`SPEC-122` Nr. 2), und der M2-Test ist der erste Lauf von Record-Dienst und Adapter zusammen, in dem ein Datenrennen zwischen Proxy, Referenzdienst und Client auftreten kann (Sendung von `slice-race-gate`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Start im Modus `mock` oder `hybrid` — scheitert mit `ODR-E2001`, bis die Folgewellen die Use Cases liefern (Entscheidung steht in `slice-fehlermodell-konfiguration`).
- CLI-Kommandos für Import, Workspace und Session über den Start hinaus, Admin-Listener — Folgewelle; der Test erreicht Import, Workspace und Session-Ende über den Handle des Composition Root.
- Lasttest zu `LH-QA-01` — braucht Telemetrie und ein Messprofil (Welle §6); bleibt ausdrücklich ungeprüft.
- Vergleichstest zu `LH-QA-07` (Direktzugriff gegen Proxy) — `slice-protokolltreue`, nach diesem Slice; er nutzt Referenzdienst und Composition Root von hier und bringt mit, was der Vergleich zusätzlich braucht. Dieser Slice bleibt beim Satz von M2 (Entscheidung des Users vom 2026-10-21).

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

- [ ] Verdrahtung und Lebenszyklus: der Start-Befehl ruft `LadeOderMelde` (Exit-Code 1, Meldung nach `SPEC-110`, `mock`/`hybrid` scheitern mit `ODR-E2001`; Zusage von `slice-fehlermodell-konfiguration`); Reihenfolge `MaintenanceUseCase` → erste Session → Listener, ein Wächter vor dem Blob-Port für alle Schreiber und die Wartung (Fänger in §1, Sendung von `slice-session-wartung`); ein Use Case je Modus am Listener; `OpenInitialSession` je Service im Modus `record`, nachdem die Services der Konfiguration im `MetadataRepository` stehen (der Root legt sie an, §1 *Übernimmt (Services)*, `SPEC-107` Nr. 14; Sendungen von `slice-session-use-case` und `slice-odata-codec-csdl`); Herunterfahren mit `CloseAll` über denselben `RecordUseCase`, den der Adapter ruft, in der Reihenfolge Listener nehmen nichts mehr an, `CloseAll` mit `Frist` gleich `SHUTDOWN_TIMEOUT`, Verbindungen und Repositories schließen (Sendung von `slice-session-beenden`, `SPEC-124` Nr. 14); ein Handle `Build`, über den Tests die Use Cases erreichen, darunter `ImportUseCase` mit denselben `Regeln` wie der Record-Dienst (§1 *Übernimmt (Import)*, Sendung von `slice-baseline-import`); Warnung beim Start bei abgeschalteter Zertifikatsprüfung zum Upstream (`LH-RB-02.a`, Sendung von `slice-upstream-adapter`); der Root übergibt `Pruefen` aus dem Wert des Starts, Fänger und Mutationen in §1 (Sendung von `slice-upstream-tls`).
- [ ] Referenzdienst als Fixture (`httptest`, OData V4, eigenes CSDL, Collection mit zwei Entities, Einzel-Entity, wiederholter Request mit verschiedenen Antworten, ein Header mit Geheimnis) und der M2-Test: aufzeichnen, Session beenden, importieren, Workspace anlegen, Upstream beenden, abspielen; Status, Header und Body byte-gleich nach `SPEC-045`; Zähler am Referenzdienst im Replay gleich 0. Der Test läuft in `make test` und in `make test-race`: `tests/` liegt im Quellstand der Stufe `source` (Sendung von `slice-race-gate`, §1).
- [ ] Mutationen des M2-Tests: Replay kontaktiert den Upstream, ein Geheimnis erreicht die SQLite-Datei oder den Blob-Ordner, die Reihenfolge wiederholter Requests kippt — jede färbt den Test rot.
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
| `src/bootstrap/` | update | Verdrahtung, Startsequenz, Herunterfahren, `Build` |
| `src/bootstrap/main/` | update | Kommando zum Start; ruft `LadeOderMelde`, Exit-Codes (Zusage von `slice-fehlermodell-konfiguration`) |
| `tests/integration/` | neu | Referenzdienst und M2-Test |
| `Dockerfile`, `.dockerignore` | update | `tests/` im Quellstand der Stufe `source` und im Build-Kontext, damit `make test` und `make test-race` dieselben Pakete fahren (Sendung von `slice-race-gate`, `SPEC-122` Nr. 2) |
| `spec/spezifikation.md` | update | Startsequenz und Exit-Codes, falls dort offen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

- Der Test läuft in der Stufe `test` des `Dockerfile` (netzlos, nur Loopback) und ist damit Teil von `make test`; `tests/` kommt dafür in die Stufe `source`, nicht erst in `test`, so dass `make test-race` ihn ebenfalls fährt (§1, Sendung von `slice-race-gate`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-http-adapter` und `slice-baseline-import` und `slice-workspace-use-case` und `slice-session-wartung` und `slice-session-use-case` und `slice-session-beenden` und `slice-upstream-adapter` und `slice-upstream-tls` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 1100 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 800 Zeilen in derselben Metrik.
- `in-progress` → `open` (blockiert): der Test kann Import, Workspace und Session-Ende nicht ohne Admin-API erreichen (siehe §6).

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

- Aufrufweg zu Session-Ende, Import und Workspace ohne Admin-API: der Handle des Composition Root ist eine Entscheidung zwischen Architektur und Test — **Ausgang:** offen; Architect entscheidet vor dem Code, ob `Build` exportiert wird oder ein internes Test-Paket den Root baut.
- Der Test deckt alle Schichten und bricht, wo ein Vorgänger-Slice eine Randform offen ließ — **Ausgang:** offen; jeder Fund geht als Befund an den Slice, der ihn verantwortet, nicht in den Test.
- Loopback und Ports im Testcontainer (`--network none`): `httptest` bindet auf Loopback — **Ausgang:** offen; wird vor dem Code in `make test` erprobt.
- Zeitabhängigkeit: Herunterfahren, Fristen, parallele Clients — **Ausgang:** offen; das Gate fährt `-race` (`make test-race`, `SPEC-122` aus `slice-race-gate`) und damit auch den M2-Test, sobald `tests/` im Quellstand der Stufe `source` liegt (Sendung von `slice-race-gate`, §1): ein Datenrennen macht es rot; die nebenläufigen Fänger ordnen weiter mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8b), und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50`.
- M2 ist ein Meilenstein, kein Slice-Ziel: Die Roadmap setzt ihn erst nach der Closure der Welle auf *erreicht* — **Ausgang:** offen; der Slice liefert den Beleg, die Welle-Closure setzt die Zelle.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- Wer die konfigurierten Services im `MetadataRepository` anlegt, bevor der Composition Root `OpenInitialSession` ruft (Sendung von [slice-session-use-case](../done/slice-session-use-case.md), `AGENTS.md` §3.13): `SPEC-124` Nr. 5 verlangt, dass der Service dort steht (sonst `ODR-E5005`); der Dienst `SessionUseCase` legt keinen Service an (`SPEC-124` Nr. 1, `Dienste` dient allein `List`), und `SPEC-123` Stufe 3 liest dieselbe Tabelle mit `Dienste.List` — **Ausgang:** offen, bis der Slice schließt; der Root legt die Services an, mit allen Randformen (bekannter Service mit gleichem Root: nichts; abweichender Root oder abweichende Version: `ODR-E2001`; nicht genannter Service: unberührt) in `SPEC-107` Nr. 14 (Entscheidung des Architect von `slice-odata-codec-csdl`); die Zeilen *Zusage · Fänger* dazu gehören in die Tabelle in §3 (Fänger und Mutationen in §1 *Übernimmt (Services)*).

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, Ausgang `verkörpert` in `AGENTS.md` §3.10); der Test ist nebenläufig: das Gate fährt `-race` (`make test-race`, `SPEC-122`), auch über `tests/`, sobald die Stufe `source` es aufnimmt (§1), die nebenläufigen Fänger ordnen weiter mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8b), und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
