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

**Berührte Spec-Stellen:** `ARC-012` (Composition Root) · `LH-FA-12.a` · `LH-FA-02.c` · `SPEC-048` · `architecture.md` §2.5 (Modus-Routing), §4.9 · `SPEC-049` (Fixture-Rahmen des Referenzdienstes) · `SPEC-120` Nr. 1, `SPEC-110` Nr. 10, `LH-RB-02.a` (Verdrahtung des Upstream-Adapters, Warnung beim Start).
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

- [ ] Verdrahtung und Lebenszyklus: der Start-Befehl ruft `LadeOderMelde` (Exit-Code 1, Meldung nach `SPEC-110`, `mock`/`hybrid` scheitern mit `ODR-E2001`; Zusage von `slice-fehlermodell-konfiguration`); Reihenfolge `MaintenanceUseCase` → erste Session → Listener; ein Use Case je Modus am Listener; Herunterfahren mit `CloseAll`; ein Handle `Build`, über den Tests die Use Cases erreichen; Warnung beim Start bei abgeschalteter Zertifikatsprüfung zum Upstream (`LH-RB-02.a`, Sendung von `slice-upstream-adapter`); der Root übergibt `Pruefen` aus dem Wert des Starts, Fänger und Mutationen in §1 (Sendung von `slice-upstream-tls`).
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

**Start** (`next` → `in-progress`): `slice-http-adapter` und `slice-baseline-import` und `slice-workspace-use-case` und `slice-session-wartung` und `slice-upstream-adapter` und `slice-upstream-tls` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

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
