# Slice slice-m2-referenzdienst: Referenzdienst, Import- und Workspace-Verdrahtung und M2-Test

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-01`, `LH-FA-02`, `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-FA-18`, `LH-FA-11`, `LH-FA-09`, [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, Test in der Stufe `test`), [ADR-0007](../../adr/0007-race-detektor-im-gate.md) (Race-Detektor im Gate).

**Berührte Spec-Stellen:** `SPEC-045` (byte-gleiche Wiedergabe) · `SPEC-049` (Fixture-Rahmen des Referenzdienstes; bleibt unverändert, der Referenzdienst ist Testcode und kein Vertrag) · `SPEC-127` Nr. 1 und 10 bis 12 (Verdrahtung und Konflikte des Imports) · `SPEC-107` Nr. 15 (`MetadataHash`) · `SPEC-123` Nr. 1, 9 (Wächter vor dem Blob-Port) · `SPEC-122` Nr. 2 (Pakete der Stufe `source`) · `SPEC-139` Nr. 1 (Handle des Composition Root, von `slice-bootstrap-m2` geliefert; Zugänge `Import()` und `Workspaces()` je Modus) · `SPEC-129` (Dienst-Vertrag des Workspace) · `SPEC-135` Nr. 9 (Abspielen: leere Menge, Cursor) · `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-01.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Test zeichnet über den Recorder gegen einen Referenzdienst auf, beendet die Session, importiert sie, legt einen Workspace an, beendet den Upstream und spielt dieselben Antworten byte-gleich nach `SPEC-045` ab, mit Zähler 0 am Referenzdienst: der Beleg für M2 (die Zelle setzt die Welle-Closure, nicht dieser Slice).

**Herkunft (Schnitt vor dem ersten Code-Commit, `AGENTS.md` §3.14):** `slice-bootstrap-m2` lieferte Composition Root und M2-Beleg in einem Plan; die Formel von `SPEC-128` gab für beide zusammen S rund 3200 (Schnittfrage), für den Root-Kern allein S = 2120. Geschnitten ist an der Stelle, an der kein Mechanismus geteilt wird: Referenzdienst, Import und Workspace sind additive Zugänge am `Handle` ([slice-bootstrap-m2](../in-progress/slice-bootstrap-m2.md), `SPEC-139` Nr. 1), der Kern kommt für seinen sichtbaren Schritt mit einem kleinen Upstream im Testcode von `src/bootstrap` aus. Reihenfolge der Welle: `slice-bootstrap-m2` → dieser Slice → [slice-protokolltreue](slice-protokolltreue.md).

**Übernimmt (aus `slice-bootstrap-m2`, `AGENTS.md` §3.13):**

- **Referenzdienst als Fixture** (`httptest`, OData V4, eigenes CSDL; Rahmen in §3). Aus dem Plan von `slice-bootstrap-m2`: Collection `Products` mit zwei Entities, eine Einzel-Entity, die eine der beiden Entities der Collection ist (gleicher Schlüssel, gleiche Werte; sonst ist die Coverage `partial` mit `bestand_ausserhalb_kette`, `SPEC-127` Nr. 12 g), ein wiederholter Request mit wechselnden Antworten auf einem Pfad, den der Import überspringt (`/Products/$count`; auf einer Entity oder Collection wäre es ein Konflikt, `ODR-E6101`, `SPEC-127` Nr. 10), ein Header mit Geheimnis (Antwort und Request), ein Haltepunkt `/Hold` und Zähler für Verbindungen und Requests. Das Port-Fake-Inventar ist dieser eine Test-Server.
- **Quellstand:** `tests/` liegt im Quellstand der Stufe `source` des `Dockerfile` und im Build-Kontext (`.dockerignore`), nicht erst in der Stufe `test`: `make test-race` fährt alle Pakete der Stufe `source` (`SPEC-122` Nr. 2), und der M2-Test ist der erste Lauf von Record-Dienst, Adapter und Client zusammen, in dem ein Datenrennen auftreten kann (Sendung von `slice-race-gate`, dort in `done/`; Nehmer ist ab jetzt dieser Slice).
- **Verdrahtung von Import und Workspace:** `baselineimport.New` (`SPEC-127` Nr. 1) und `workspace.New` mit Zugang am `Handle` (`Import()`, `Workspaces()`; additiv, kein Feld und keine Operation von `SPEC-139` Nr. 1 wird umbenannt oder entfernt). Import: `Dienste` (`ID`, `UpstreamURL` je Service, dieselben Werte wie bei `record.New`), **dieselben** `SchluesselRegeln` wie der Record-Dienst (sonst passt der Key der Folgeseite zu keinem aufgezeichneten Exchange, und `complete` bleibt aus), `MaxCaptureBody`, `Sessions`, `Exchanges`, `Metadaten`, `Snapshots`, `Imports` aus dem SQLite-Adapter, `Blobs` mit dem Wächter (`SPEC-123` Nr. 1, 9: der Import schreibt das Manifest), `MetadatenCodec` aus `csdl.New()`, `DatenCodec` aus `jsoncodec.New()`, die eine Uhr und die eine `IDGenerator`-Instanz (`SPEC-139` Nr. 3).
- **Der M2-Test** und seine Mutationen (Replay kontaktiert den Upstream; ein Geheimnis erreicht die SQLite-Datei oder den Blob-Ordner; die Reihenfolge wiederholter Requests kippt).
- **Die Zeilen des Plans zum Import** (Tabelle *Zusage · Fänger*, §3): Fänger „Regeln gleich“, Wächter am Import, `MetadataHash` gleich der Blob-Referenz des `$metadata` (`SPEC-107` Nr. 15), Coverage von `Products` `complete`. Sie sind zugleich der einzige Beleg, dass Schlüsselform und Antwortform der echten Leser zum Import passen: die Suite von `slice-baseline-import` führt Fakes der Codecs, der Import des M2-Tests läuft mit `csdl.New()` und `jsoncodec.New()` (Sendung von `slice-baseline-import`).

**Nutzt:** Den Composition Root von `slice-bootstrap-m2`: `Build`, `Handle`, `Hochfahren`, `Herunterfahren`, `Sessions()`, `Record()` (`SPEC-139` Nr. 1). Der Slice ändert dessen Verhalten nicht, er fügt dem `Handle` zwei Zugänge hinzu. Ort der Zusage für beide Zugänge ist `SPEC-139` Nr. 1: `Import()` ist im Modus `record` der Dienst und im Modus `replay` `nil`, `Workspaces()` ist in beiden Modi der Dienst (der Wortlaut kommt mit `slice-bootstrap-m2` auf `main`; §3 Zeile M9b, §6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Composition Root, Start-Befehl, Exit-Codes, Herunterfahren und Wartung — [slice-bootstrap-m2](../in-progress/slice-bootstrap-m2.md); dieser Slice setzt dessen Abschluss voraus und ändert ihn nicht.
- Admin-API oder CLI-Kommandos für Import, Workspace und Session-Ende — Folgewelle bzw. entfallen (Entscheidung des Users, Welle §6): der Test erreicht Session-Ende, Import und Workspace über den `Handle`, nicht über HTTP.
- Das gebaute Binary im Container und `make test-integration` (`SPEC-037`) — [slice-test-integration](slice-test-integration.md); der M2-Test fährt den ganzen Pfad im Prozess von `go test`.
- Ergänzungen am Referenzdienst für den Vergleich von Direktzugriff und Proxy (Antworten mit `Content-Encoding`, mehrere Werte je Header) — [slice-protokolltreue](slice-protokolltreue.md); dessen §1 und §3 führen sie, und der M2-Test bleibt beim Satz von M2 (Entscheidung des Users vom 2026-10-21).
- Zähler nach Methode am Referenzdienst — gehört den Slices, die ihn brauchen (`slice-bootstrap-m4`); dieser Slice zählt Verbindungen und Requests.
- Lasttest zu `LH-QA-01` — braucht Telemetrie und ein Messprofil (Welle §6); bleibt ausdrücklich ungeprüft.

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

- [ ] Verdrahtung und Quellstand: `baselineimport.New` und `workspace.New` sind im Composition Root verdrahtet (§1 *Übernimmt*), `Import()` und `Workspaces()` liegen additiv am `Handle`; `tests/` liegt im Quellstand der Stufe `source` und im Build-Kontext (`Dockerfile`, `.dockerignore`), so dass `make test` und `make test-race` dieselben Pakete fahren (`SPEC-122` Nr. 2; Sendung von `slice-race-gate`).
- [ ] Referenzdienst als Fixture und der M2-Test: aufzeichnen, Session beenden, importieren, Workspace anlegen, Upstream beenden, abspielen; Status, Header und Body byte-gleich nach `SPEC-045`; Zähler am Referenzdienst im Replay gleich 0; der Test läuft in `make test` und in `make test-race`.
- [ ] Zeilen und Mutationen: je Zeile der Tabelle *Zusage · Fänger* (§3) eine Mutation, vom Implementer rot gesehen (`AGENTS.md` §3.10); darunter die Zeilen zum Import (Regeln gleich, Wächter am Import, `MetadataHash`, Coverage `complete`) und die drei Mutationen des M2-Tests.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (erwartet: keiner; `SPEC-049` bleibt, der Referenzdienst ist Testcode).
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
| `src/bootstrap/` | update | `baselineimport.New` und `workspace.New` verdrahten; `Import()` und `Workspaces()` am `Handle` (additiv, `SPEC-139` Nr. 1) |
| `tests/integration/` | neu | Referenzdienst (Fixture) und M2-Test (`SPEC-045`, `LH-FA-04`) |
| `Dockerfile`, `.dockerignore` | update | `tests/` im Quellstand der Stufe `source` und im Build-Kontext, damit `make test` und `make test-race` dieselben Pakete fahren (Sendung von `slice-race-gate`, `SPEC-122` Nr. 2) |

**Fixture-Rahmen des Referenzdienstes** (Vorgabe des Architect; `SPEC-049` bleibt unverändert, der Referenzdienst ist Testcode, kein Vertrag):

- Pfade: `$metadata` (eigenes CSDL mit Entity `Product`), `/Products` (zwei Entities, ohne Folgeseiten-Link), `/Products(<Schlüssel>)` (eine der zwei Entities der Collection, gleiche Werte), `/Products/$count` (wechselnde Antworten bei wiederholtem Request; der Import überspringt den Pfad), `/Hold` (Haltepunkt: die Antwort wartet auf ein Signal des Tests).
- Zähler für Verbindungen und Requests, lesbar für den Test; im Replay ist jeder gleich 0.
- Ein Antwortheader und ein Request-Header mit Geheimnis; das Geheimnis erreicht weder die SQLite-Datei noch den Blob-Ordner.
- Ein Test-Server (`httptest`) ist das ganze Port-Fake-Inventar; Kanäle ordnen, kein Schlaf.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Köpfe wie `SPEC-128` Nr. 3 a). Die Zeilen zum Import stammen aus dem Plan von `slice-bootstrap-m2`; die Fänger-Namen legt der Implementer fest, der Planner erwartet, dass jede Zeile vor dem Handoff einen Test und eine rot gesehene Mutation trägt.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| M1 | Aufzeichnen, Session beenden, importieren, Workspace anlegen, Upstream beenden, abspielen: Status, Header und Body byte-gleich nach `SPEC-045` (`LH-FA-04`, `LH-FA-05`) | M2-Test am `Handle` mit echten Adaptern und Referenzdienst: Vergleich Aufnahme gegen Wiedergabe je Request | Replay liefert eine Antwort mit verändertem Body (ein Byte), verändertem Header oder anderem Status: rot; die Reihenfolge der Schritte vertauscht (abspielen vor Import): `ODR-E5003` oder Fehler statt Gleichheit | nebenläufig |
| M2 | Im Replay erreicht kein Request den Referenzdienst: Zähler für Verbindungen und Requests 0 (`LH-FA-04`) | M2-Test: Upstream beenden, abspielen, Zähler lesen; der Test fordert zusätzlich einen Request, den keine Aufnahme trägt (Fehlerantwort, Zähler weiter 0) | Replay kontaktiert den Upstream (Fallback): Zähler größer 0: rot | andere |
| M3 | Ein Geheimnis (Header, Query, Body-Pfad aus der Konfiguration) erreicht weder die SQLite-Datei noch den Blob-Ordner (`LH-FA-14`) | M2-Test: Dateien in `DATA_DIR` nach der Aufnahme auf Rohbytes nach dem Geheimnis durchsucht, Schleife über jeden Ort (Datei, Blob-Datei) | Redaction der Pipeline fehlt für einen der Orte: Geheimnis in der Datei: rot | andere |
| M4 | Wiederholte Requests auf demselben Pfad werden in der aufgezeichneten Reihenfolge abgespielt (Cursor, `LH-FA-05`) | M2-Test: `/Products/$count` dreimal mit verschiedenen Antworten, Replay gibt sie in derselben Folge | Reihenfolge der Antworten gekippt (oder immer die erste): rot | andere |
| M5 | Der Import erhält dieselben `SchluesselRegeln` wie der Record-Dienst (Verdrahtung, `SPEC-127` Nr. 1) | Der Fänger liest die Evidenz: Referenzdienst mit zweiter Seite (Link), Kette `proven`; Naht der Konfig des Imports gegen die des Record-Dienstes | der Root übergibt dem Import andere `Regeln`: Kette `seite_fehlt` statt `proven`, Coverage nicht `complete`: rot | andere |
| M6 | Der Wächter ist `BlobRepository` des Imports; `CleanupBlobs` im Abstand von weniger als der Karenz entfernt den Manifest-Blob nicht (`SPEC-123` Nr. 1, 9) | M2-Test: nach dem Import `CleanupBlobs` ohne offene Session; Manifest-Blob besteht; Zeigergleichheit des Wächters mit dem der Pipeline | der Root gibt dem Import das darunterliegende Repository: Blob entfernt; zwei Wächter: Blob entfernt | andere |
| M7 | Nach dem Import trägt der Service `MetadataHash` gleich der Blob-Referenz des `$metadata` (`SPEC-107` Nr. 15) | M2-Test: `Metadaten.Get` gegen die Referenz des aufgezeichneten `$metadata` | `MetadatenCodec` ein Fake oder `SetMetadata` entfällt: Hash leer: rot | andere |
| M8 | Die Coverage von `Products` ist `complete`, wenn der Referenzdienst die Collection ohne Link liefert (`SPEC-127` Nr. 12) | M2-Test liest die Coverage nach dem Import; zweite Variante: eine Entity außerhalb der belegten Kette (gefilterte Abfrage) gibt `partial` mit `bestand_ausserhalb_kette` | `DatenCodec` ein Fake oder eine falsche Schlüsselform: Coverage `partial` statt `complete`: rot | andere |
| M9 | `Workspaces()` und `Import()` liefern im Modus `record` die verdrahteten Dienste, und ein Workspace entsteht aus der importierten Baseline (`LH-FA-09`, `SPEC-003`) | M2-Test: `Workspaces().Create` ohne Baseline scheitert, nach dem Import gelingt es | `Workspaces()` ohne Verdrahtung (nil): rot; Workspace aus anderer Baseline-ID: rot | andere |
| M9b | `Import()` ist im Modus `record` der verdrahtete Dienst und im Modus `replay` `nil`; `Workspaces()` ist in beiden Modi der verdrahtete Dienst (`SPEC-139` Nr. 1, `SPEC-129` Kopf, `SPEC-127` Nr. 1) | Test am `Handle` je Modus: `Import() == nil` als Schnittstellenvergleich (nicht `reflect`) in `replay`, `!= nil` in `record`; `Workspaces() != nil` in beiden Modi | der Root verdrahtet `Import` auch in `replay` oder gibt einen typisierten Nullzeiger: rot; `Workspaces` nur in `record` verdrahtet: rot | andere |
| M10 | `tests/` liegt im Quellstand der Stufe `source`: `make test-race` fährt das Paket des M2-Tests (`SPEC-122` Nr. 2) | Gate-Ausgabe von `make test-race` nennt das Paket unter `tests/`; ein absichtlich eingebautes Datenrennen in einem Test unter `tests/` färbt das Ziel rot | `tests/` aus der Stufe `source` entfernt: das Paket fehlt in der Ausgabe, das Rennen bleibt unentdeckt: rot | andere |

Zu M10: die Gegenprobe der Verdrahtung des Dockerfile ist laut `harness/README.md` §Sensors nicht behauptet (*Nicht behauptet*, Stufen von `Dockerfile`); der Implementer prüft die Mutation von Hand und nennt sie im Bericht, oder die Zeile steht als akzeptiertes Negativ mit Grund und Folge (Folge: ein Paket unter `tests/` entginge dem Detektor, bis jemand es sieht).

Raten (`AGENTS.md` §3.10): M1 ist nebenläufig (Proxy, Referenzdienst, Client in einem Prozess). Der Implementer misst die Rate je Mutation und je Lauf von `go test` und trägt Zahl der Prozesse und Messenden hier ein; die Fänger ordnen mit Kanälen (Haltepunkt `/Hold`), nicht mit Schlaf.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` liegt in `done/` (der `Handle` mit `Build`, `Hochfahren`, `Herunterfahren`, `Sessions()` ist im Code, `SPEC-139` in der Spezifikation) und `slice-baseline-import`, `slice-workspace-use-case`, `slice-race-gate` und `slice-odata-codec-json` liegen in `done/`. Das Raster in §6 hat vor dem ersten Code-Commit für jede Zelle einen Ort (`AGENTS.md` §3.12, §3.16); eine Zelle ohne Ort geht an den Architect. Der Planner setzt die Marke vor dem ersten Code-Commit mit `make groessenmarke` (`SPEC-128`) und trägt den Block ein.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`), vorläufiger Entwurf des Architect (Vorabentwurf zu `slice-bootstrap-m2`, Abschnitt A7): die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3 (`src/bootstrap/`, `tests/`, `Dockerfile`, `.dockerignore`), Tests eingeschlossen; Text der Spezifikation zählt nicht mit. Zählung nach der Tabelle in §3: *a* = 1 nebenläufige und *b* = 10 andere Zeilen (M9b eingeschlossen), *f* = 1 (der Referenzdienst als Test-Server), *p* = 100 (Produktcode: zwei Konstruktor-Aufrufe, zwei Zugänge am `Handle`, Dockerfile und `.dockerignore`), Schätzung S = 85·1 + 80·10 + 200·1 + 100 = **1185**, Marke M = 100 · ⌈135·1185 / 10000⌉ = **1600** (der Architect entwarf 1500 bei 9 Zeilen; die Zeile M9b hebt S um 80). Das ist ein Entwurf: die Marke setzt der Planner vor dem ersten Code-Commit mit dem Werkzeug und nach der Zählung der dann geltenden Tabelle, und ersetzt sie dabei mit Grund; der Block (`pfad`, `produktcode`, `fakes`, `basis`, `schaetzung`, `marke`) entsteht dann hier. Der Slice liegt weit unter der Schnittfrage (S ab 2223).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke (sie ist die Schätzung × 1,35 nach `SPEC-128` und schneidet nicht, `AGENTS.md` §3.14); offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten).
- `in-progress` → `open` (blockiert): der Test erreicht Session-Ende, Import oder Workspace nicht über den `Handle` (der Zugang ist additiv in `SPEC-139` Nr. 1 vorgesehen; fehlt er dort, geht die Frage an den Architect), oder `make test-race` fährt `tests/` nicht, ohne dass `SPEC-122` Nr. 2 geändert wird.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zeile der Tabelle ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor. Die Zelle M2 der Roadmap setzt dieser Slice nicht; sie ist offen bis zur Closure der Welle.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der Test deckt alle Schichten und bricht, wo ein Vorgänger-Slice eine Randform offen ließ — **Ausgang:** offen; jeder Fund geht als Befund an den Slice, der ihn verantwortet (die Nehmer liegen in `done/`, der Befund geht an den Planner, `AGENTS.md` §3.13), nicht in den Test.
- Loopback und Ports im Testcontainer (`--network none`): `httptest` bindet auf Loopback — **Ausgang:** offen; `slice-bootstrap-m2` erprobt dasselbe vor seinem Code, der Implementer übernimmt dessen Ergebnis.
- Nebenläufigkeit: Proxy, Referenzdienst und Client laufen in einem Prozess — **Ausgang:** offen; `make test-race` fährt `tests/` (M10); ein Rennen ohne ungeordneten Zugriff meldet der Detektor nicht (`SPEC-122` Nr. 8 b), die Fänger ordnen mit Kanälen, und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50`.
- M2 ist ein Meilenstein, kein Slice-Ziel: die Roadmap setzt ihn erst nach der Closure der Welle auf *erreicht* — **Ausgang:** offen; der Slice liefert den Beleg, die Welle-Closure setzt die Zelle.
- Der Fixture-Rahmen wächst, bis er die Folge-Slices trägt (`slice-protokolltreue`, `slice-bootstrap-m3`, `slice-bootstrap-m4`, `slice-test-security`): Ergänzungen sind dort Umfang, nicht hier — **Ausgang:** offen; die Folge-Slices tragen ihren Bedarf in §1 (Nachzug dieses Commits); der M2-Test bleibt beim Satz von M2 (Entscheidung des Users vom 2026-10-21).
- Raster *Operation × Argumentart* (`AGENTS.md` §3.12, §3.16), jede Zelle nennt ihren Ort oder „gilt nicht“ mit Grund (vom Architect entschieden), eine Zelle ohne Ort hält den Implementer an; der Wortlaut von `SPEC-139` Nr. 1 für `Import()` und `Workspaces()` kommt mit `slice-bootstrap-m2` auf `main` — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| `Import()` am `Handle` | Modus `replay`: `nil` (der Import löst Links gegen `UPSTREAM_URL` auf, im Modus `replay` nicht erforderlich): `SPEC-139` Nr. 1 (Wortlaut kommt mit `slice-bootstrap-m2` auf `main`) | gilt nicht: der Zugang hat kein Argument | gilt nicht: kein Argument | Kontext des Aufrufs `Import(ctx, ...)` abgebrochen: Fehler unverändert, ohne Code, keine Baseline: `SPEC-127` Nr. 2 | beliebig oft, gleichzeitig: `SPEC-139` Nr. 1 (Satz zur Nebenläufigkeit) und Nr. 10 |
| `Workspaces()` am `Handle` | nie `nil`, der Dienst steht in beiden Modi: `SPEC-139` Nr. 1 (Wortlaut kommt mit `slice-bootstrap-m2` auf `main`) und `SPEC-129` Kopf | gilt nicht: kein Argument | gilt nicht: kein Argument | gilt nicht: Zugang ohne Argument; Abbruch der Operationen: `SPEC-129` Nr. 2 | wie `Import()` |
| Import der Session | Session ohne Exchange: `ODR-E6004` (Ursache `baselineimport.ErrKeineMetadaten`, 422): `SPEC-127` Nr. 7 b, Fehlerliste Nr. 17 | `$metadata` ohne Entity-Set: akzeptiertes Negativ, Grund und Folge unter der Tabelle | Konflikt, wechselnde Antwort auf Entity oder Collection: `ODR-E6101` (`SPEC-127` Nr. 10) | Kontext abgebrochen: Fehler des Kontexts unverändert, keine Baseline, kein Publish: `SPEC-127` Nr. 2 und 4 | gleiche `import_id`, gleiche Anfrage: `Neu` falsch, `ReportJSON` byte-gleich; gleiche `import_id`, andere Anfrage: `ODR-E6102`: `SPEC-127` Nr. 5; andere `import_id` auf derselben Session: gilt nicht, ein Import je Lauf (Entscheidung über Publish, `SPEC-107` Nr. 7) |
| Workspace anlegen | Baseline fehlt: leere Baseline-ID `ErrUngueltig`, unbekannte Baseline `ODR-E7005`: `SPEC-129` Nr. 4 b und 5, `SPEC-109` Nr. 5 | leere Kennung: der Dienst vergibt sie (`IDs.WorkspaceID()`): `SPEC-129` Nr. 3 und 4 e | Baseline einer fremden Service-ID: gilt nicht, `Create` erhält keine Service-ID, der Service kommt aus der Baseline: `SPEC-129` Kopf | Kontext abgebrochen: Fehler des Kontexts unverändert, ohne Code: `SPEC-129` Nr. 2 | doppeltes Anlegen: `ODR-E7007` (409): `SPEC-129` Nr. 3 und `SPEC-109` Nr. 5 (`ErrBereitsVorhanden`, Gegenstand `Workspace`) |
| Abspielen (M2-Test) | Request ohne Zuordnungs-Header: `ODR-E5003` (400): `SPEC-135`, `SPEC-139` Nr. 5 | gilt nicht: der Test baut jeden Request selbst | Request ohne Aufnahme: `ODR-E5001` (404), Zähler 0: `SPEC-135` Nr. 9 (M leer) | Upstream beendet: erwartet (Zähler 0) | wiederholter Request: Cursor `sequential`: `SPEC-135` Nr. 9 (M4) |

**Akzeptiertes Negativ: `$metadata` ohne Entity-Set.** Grund: Der Referenzdienst führt `Products`; die Zusage gehört dem Dienst (`slice-baseline-import`, in `done/`) und folgt aus `SPEC-127` Nr. 12 und 14 (je Entity-Set ein Eintrag, `entity_count` gleich der Zahl der Entities: eine Baseline mit 0 Entities und leerer Coverage), sofern `SPEC-125` das Dokument annimmt. Folge: Kein Fall im M2-Test; die Frage bleibt Sache der Suite des Dienstes, und wird sie dort gestellt, geht sie an den Architect. Der Fall wird nicht in den M2-Test gezogen.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage (2026-11-01, 39 Verzeichnisse unter `BEO-REPO`) durchgegangen; der Planner zieht Treffer und Zähler beim Start nach. Einschlägig: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (12 Belege, verkörpert in `AGENTS.md` §3.10 und `make test-race`; der M2-Test ist nebenläufig, Zeile M1, §6), `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (4 Belege, offen; Prozess, Sockets und Referenzdienst laufen unter Last, Kanäle statt Schlaf, Raten mit Zahl der Läufe), `BEO-REPO/fixture-gleicht-dem-wert-der-mutation` (2 Belege, offen; der Referenzdienst trägt Werte, die eine Mutation nicht erzeugt: das Geheimnis, M3), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (37 Belege, verkörpert in `AGENTS.md` §3.15; Tabelle in §3), `BEO-REPO/randform-erst-im-review-entschieden` (26 Belege, verkörpert in `AGENTS.md` §3.16; Raster in §6) und `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (16 Belege, verkörpert in `AGENTS.md` §3.9; der Schnitt aus `slice-bootstrap-m2` zieht jede Stelle nach, die den Referenzdienst oder den M2-Test dort nennt). Den Stand der beiden offenen Einträge nennt ihre `state.md`; der Planner zieht ihn beim Start nach, ein Beleg dieses Slice kommt dann als Datei unter `evidence/` dazu.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
