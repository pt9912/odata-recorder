# Slice slice-bootstrap-m4: Composition Root für den Modus hybrid und der M4-Test

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, `LH-QA-02`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` · `LH-FA-04.c` (Pflicht der `X-Recorder-*`-Header im Modus `hybrid`) · `SPEC-110` Nr. 2, 8, 9 (Schlüssel dieser Version, Modus, Services) · `SPEC-010`, `SPEC-062`, `SPEC-075`, `SPEC-076`, `SPEC-079`, `SPEC-080` · `SPEC-055` (`PUBLIC_BASE_URL` in `hybrid`) · `ARC-005`, `ARC-012` · `architecture.md` §2.5 (Modus-Routing).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-bootstrap-m4)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Composition Root bindet im Modus `hybrid` genau `HybridUseCase` an den OData-Listener, liest die Schlüssel `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` (Syntax und Zusammenhänge nach `SPEC-110`), verlangt bei freigegebenem Upstream-Fallback die `UPSTREAM_URL` des Service, und der M4-Test aus §3 der Welle läuft grün in `make test`: ein Referenzdienst zählt Verbindungen und Methoden und belegt die Entscheidungsmatrix.

**Übernimmt:** den Nachzug von `SPEC-110` Nr. 2 und 8 für den Modus `hybrid` und die Schlüssel `HYBRID_*` und `FALLBACK_*`, den `slice-bootstrap-m3` für den Modus `mock` hinterlässt (dessen §1 *Ausdrücklich NICHT* weist beides diesem Slice zu), und das Lesen der Policy-Werte aus der Umgebung (Sendung von `slice-fallback-policy`, dessen §1 *Ausdrücklich NICHT* diesen Slice nennt), und den Zähler des Referenzdienstes nach Methode (Sendung von `slice-m2-referenzdienst`, dessen §1 *Ausdrücklich NICHT* diesen Slice nennt; der Referenzdienst dort zählt Verbindungen und Requests, die Erweiterung ist hier Umfang, Ausgang in §6), und aus dem Vertrag der Policy (`SPEC-261` Nr. 4 e, Nr. 8 c, Nr. 10 b; Sendung von `slice-fallback-policy-entscheidung`, dessen §1 *Ausdrücklich NICHT* diesen Slice nennt): ein `*KonfigFehler` der Policy wird `ODR-E2001` (die Reihenfolge der Schlüssel in der Meldung folgt dem Fehler), `HYBRID_FALLBACK=enabled` ohne `UPSTREAM_URL` wird `ODR-E2001`, und das Handbuch-Kapitel nennt die Wechselwirkung aus Nr. 4 e (wer für eine Kategorie den Replay-Fallback setzt, schaltet den schreibenden Upstream-Fallback für deren Codes ab, weil die Zeile Replay-Miss der Matrix nur `GET` und `HEAD` kennt) und die Einschränkung des Hosts (ein `UPSTREAM_URL` mit Bytes ab `0x80` oder der Zone einer IPv6-Adresse gibt nie einen Fallback frei; der Betreiber setzt die ASCII-Form, Punycode; Hinweis I6 des Reviews von `slice-fallback-policy-entscheidung`, `SPEC-261` Nr. 5 c), und den Host-Fänger am echten Adapter (Sendung von `slice-hybrid-service-upstream`, Z35, dessen §1 *Ausdrücklich NICHT* diesen Slice nicht nennt: der Referenzdienst zählt die Verbindung unter dem Host, den die Policy verglich, mit dem Adapter `httpupstream` aus dem Composition Root, `SPEC-284` *Ungeprüft* 3).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Neue Fachlogik — der Slice verdrahtet, was `slice-fallback-policy`, `slice-fallback-policy-entscheidung`, `slice-hybrid-service` und `slice-hybrid-service-upstream` liefern.
- Die Schlüssel `ADMIN_*` und `LISTEN_TLS_*` — `welle-admin-api` und `welle-betrieb`; sie bleiben `ODR-E2001`, damit keine Sicherheitseinstellung still wirkungslos ist (`SPEC-110` Nr. 2).
- Das Umschalten des Modus zur Laufzeit — `LH-FA-13` schließt es aus dem ersten Abnahmeumfang aus.
- Die Anzeige des Modus in der Diagnose — `welle-admin-api`; die Admin-API braucht den Hybrid-Betrieb nicht.
- `LINK_REWRITE` — bleibt `off` (`SPEC-092`).

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

- [ ] Verdrahtung und Konfiguration: der Modus `hybrid` bindet `HybridUseCase`; die vier Schlüssel kommen aus dem Wert des Starts bis in die Policy (Fänger, dass der Root sie weiterreicht und den Nullwert nicht stehen lässt; Muster aus `slice-upstream-tls`); `UPSTREAM_URL` ist bei `HYBRID_FALLBACK=enabled` Pflicht und fehlt sie, ist es `ODR-E2001`; ein `*KonfigFehler` der Policy wird `ODR-E2001` (Sendung von `slice-fallback-policy-entscheidung`); Zusammenhänge und Randformen nach `SPEC-110`; der Root übergibt `hybrid.Konfig` nach `SPEC-284` Nr. 1: `MaxBytes` gleich `MAX_CAPTURE_BODY`, `Upstream` immer gesetzt, `Policy` aus `policy.New`, `UpstreamURL` je Service (leer ohne `UPSTREAM_URL`) (Sendung von `slice-hybrid-service`; die Felder `Upstream` und `UpstreamURL` liefert `slice-hybrid-service-upstream`, das den Pfad mit trägt).
- [ ] M4-Test aus §3 der Welle (dazu als Stufe von `make test-integration` am gebauten Binary das Szenario des Verbindungszählers, `SPEC-037`, für die Kriterien, die die [Rückverfolgungs-Matrix](../rueckverfolgung.md) diesem Slice zuweist): ein Treffer des Mock ergibt 0 Verbindungen; eine lesende Anfrage bei unvollständiger Coverage mit Freigabe genau eine lesende Verbindung, ohne Freigabe `ODR-E6003` und 0; ein POST auf eine nicht unterstützte Operation ohne Eintrag in `FALLBACK_WRITE_ALLOW` 0, mit passendem Eintrag eine; ein Replay-Miss `ODR-E5001` mit Opt-in nur für GET und HEAD zum freigegebenen Host, ein erschöpfter Key `ODR-E5002` und fachliche `404` und `409` 0; die Verbindung des lesenden Fallbacks geht mit dem echten Adapter `httpupstream` an den Host, den die Policy verglich (Sendung `slice-hybrid-service-upstream`, Z35).
- [ ] Nicht gelieferte Schlüssel (`ADMIN_*`, `LISTEN_TLS_*`) bleiben `ODR-E2001`, und die Spezifikation führt den Nachzug von `SPEC-110` Nr. 2 und 8.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand: Modus `hybrid`, Schlüssel `HYBRID_*` und `FALLBACK_*`, die Fehlercodes des Fallbacks, die Wechselwirkung von Replay-Fallback und schreibendem Upstream-Fallback aus `SPEC-261` Nr. 4 e und die Einschränkung des Hosts (Bytes ab `0x80` und Zone im `UPSTREAM_URL` geben nie einen Fallback frei, Punycode setzen; `SPEC-261` Nr. 5 c) (beides Sendung von `slice-fallback-policy-entscheidung`), belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Das Kapitel nennt auch das Verhalten des Hybrid-Dienstes nach `SPEC-284`: den Header `X-Recorder-Fallback` einer Fallback-Antwort, dass absolute Upstream-Links einer Fallback-Antwort unverändert bleiben (`LINK_REWRITE` ist `off`), dass `ODR-E5003` fehlender Session-Header erst im Replay-Fallback entsteht und dass ein Request mit Body über `MAX_CAPTURE_BODY` nie einen Fallback erhält (Sendung von `slice-hybrid-service`). Voraussetzung ist `slice-handbuch-meldungskatalog` in `done/` (Trigger von `welle-mock`).
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
| `src/bootstrap/` | update | Wahl des Use Case je Modus um `hybrid` erweitert; die vier Schlüssel lesen, prüfen und an die Policy geben; `UPSTREAM_URL` als Pflicht bei freigegebenem Upstream-Fallback |
| `src/adapters/driving/httpodata/` | — | kein Änderungsbedarf erwartet, der Adapter kennt den Modus nicht; ein Bedarf geht an den Architect (§6) |
| `src/bootstrap/` (Tests) | neu | `m4_konfig_test.go` (Lesen, Syntax, Zusammenhänge, `*KonfigFehler`), `m4_wurzel_test.go` (Bindung des Modus, `hybrid.Konfig`, Wirkung je Schlüssel am Referenzdienst); Muster `m3_konfig_test.go`, `m3_wurzel_test.go` |
| `tests/integration/` | neu / update | `m4_test.go`: der M4-Test im Prozess von `go test` mit `bootstrap.Build`, dem echten Adapter `httpupstream` und dem Referenzdienst; der Referenzdienst (`referenzdienst_test.go`) zählt Verbindungen und Requests, die Erweiterung nach Methode ist Umfang (§6) |
| `tests/e2e/` | neu / update | Stufe `TestE2EHybrid` am gebauten Binary (`stufe_hybrid_test.go`) mit dem eigenen Referenzdienst `referenz_test.go` (`SPEC-037` Nr. 4; Zähler nach Methode ist Umfang) |
| `tools/harness/test-integration.sh` | update | Stufe 10 mit den drei Zeilen der Deklaration (`SPEC-037` Nr. 3); `docs/plan/planning/e2e-abdeckung.md` wird aus `make e2e-abdeckung` neu erzeugt; `harness/README.md` (Zeile `make test-integration`, „neun Stufen“) ist Gate-Kette und geht an den Integrationsschritt (§6) |
| `docs/user/` | update | Handbuch-Kapitel und Beispiele als Datei (DoD Punkt 7); kein Pfad der Größenmarke (`SPEC-128` Nr. 3 m) |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-037-test-integration.md` | update (Architect) | Nachzug von `SPEC-110` Nr. 2, 8, 9, 14 und der Schlüsselzeilen, Stufe 10 in `SPEC-037` Nr. 2, 3, 7; Verträge und Pläne zählen nicht in die Marke |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. **Stand der Tabelle** (Planner, 2026-11-18, nach dem Start): der Entwurf steht, jede Zeile nennt Zusage, Fänger, Mutation und Klasse; der Architect prüft jede Zeile gegen den Nachzug von `SPEC-110`, den er noch liefert (`SPEC-110` Nr. 2 und 8 lehnen die vier Schlüsselfamilien und `hybrid` heute noch ab), und trägt in §6 die Orte nach. Zeilen, deren Vertragsstelle noch fehlt, tragen „(Nachzug)“ und sind bis dahin Entwurf. Ein Fänger nennt seinen Test mit dem Kurznamen ohne `Test`; Dateien siehe oben. Jede Mutation, die ein Fänger fängt, fährt der Implementer selbst (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| M01 · `SPEC-075`, `SPEC-284` | Der Modus `hybrid` bindet genau den Hybrid-Dienst (`HybridUseCase`) an den OData-Listener; `mock` und `replay` binden ihn nie | Tabellentest Modus × Anfrage mit Fallback-Auslöser: die Antwort trägt im Modus `hybrid` den Header `X-Recorder-Fallback`, in `mock` und `replay` nie | `hybrid` bindet `MockUseCase`; `mock` bindet den Hybrid-Dienst | andere |
| M02 · `SPEC-261` Nr. 10 b | Jeder der vier Schlüssel erreicht `policy.Werte` und wirkt; der Nullwert bleibt nie stehen | Schleife über die vier Schlüssel × {gesetzt, ungesetzt} (8 Fälle, Zählprüfung gegen die Menge der Schlüssel), Orakel: erwartete Verbindungszahl nach `SPEC-261` Nr. 4, Wirkung am Referenzdienst | Schlüssel nicht gelesen; `Hosts` und `Schreibfreigaben` vertauscht; `Upstream` fest wahr oder falsch; `ReplayKategorien` nicht übergeben | andere |
| M03 · `SPEC-284` Nr. 1, 5 c | Der Root übergibt `hybrid.Konfig`: `MaxBytes` gleich `MAX_CAPTURE_BODY` (Default und gesetzter Wert), `Upstream` immer gesetzt, auch bei `disabled`, `Policy` aus `policy.New`, `UpstreamURL` je Service (leer ohne `UPSTREAM_URL`) | Body mit `Laenge` an beiden Seiten der Grenze (−1, gleich, +1) bei Default und gesetztem Wert: bis zur Grenze Fallback, darüber nie; zwei Services, einer ohne `UPSTREAM_URL`: `Ziel` je Service am Referenzdienst, der ohne URL erreicht ihn nie | `MaxBytes` fest; `UpstreamURL` des ersten Services für alle; `Upstream` nil bei `disabled` | andere |
| M04 · `SPEC-261` Nr. 8 b, 10 b (Nachzug: Reihenfolge in der Meldung) | Jeder `*KonfigFehler` der Policy wird `ODR-E2001`, der Start scheitert, kein Listener öffnet; die Meldung nennt Schlüssel und Grund, nie den Wert | Schleife über die sieben Ursachen × die drei Listenschlüssel und `HYBRID_FALLBACK` (Zählprüfung gegen die Ursachen von `SPEC-261`), dazu je Paar von Verstößen an zwei Schlüsseln die Reihenfolge | Abbildung auf `ODR-E9001`; Wert in der Meldung; nur der erste Fehler gemeldet; Listener vor der Prüfung geöffnet | andere |
| M05 · `SPEC-261` Nr. 10 b, `SPEC-076` | `HYBRID_FALLBACK=enabled` ohne `UPSTREAM_URL` in irgendeinem Service ist `ODR-E2001`; `disabled` ohne `UPSTREAM_URL` startet | Tabellentest: ein Service; zwei Services, nur einer ohne URL, beide Reihenfolgen; `SERVICES_FILE` mit Eintrag ohne `upstream_url`; `disabled` und `enabled` je mit und ohne | Prüfung nur am ersten Service; Prüfung auch bei `disabled` | andere |
| M06 · `SPEC-110` Nr. 1, 3, 6 (Nachzug) | Syntax der vier Schlüssel: leer gilt als ungesetzt; `HYBRID_FALLBACK` nur `enabled` oder `disabled` in dieser Schreibweise; Listen durch Komma, Leerraum am Rand entfernt, leerer Eintrag und `none` ungültig, Doppelte ohne Wirkung | Schleife Randform × vier Schlüssel mit Zählprüfung gegen die Tabelle des Architect (Raster §6, Zeilen 1 und 2) | je Randform eine (Leerraum nicht entfernt; `Enabled` angenommen; leerer Eintrag angenommen) | andere |
| M07 · `SPEC-110` Nr. 2, 14 (Nachzug) | Gesetzte `HYBRID_*` und `FALLBACK_*` in `record`, `replay` und `mock`, `PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid`: wie der Architect entscheidet | Schleife Modus (4) × Schlüsselfamilie (4) und die zwei Schlüssel im Modus `hybrid` | je Zelle eine | andere |
| M08 · `SPEC-110` Nr. 2 | `ADMIN_*` und `LISTEN_TLS_*` bleiben `ODR-E2001` („nicht unterstützt“) in jedem Modus, auch `hybrid`; die vier geprüften Familien sind es nicht mehr | Schleife Modus (4) × Familie (6 Familien: `ADMIN_*`, `LISTEN_TLS_*`, die vier Schlüssel) mit Zählprüfung; die Meldung nennt keinen Wert | Ablehnung nur außerhalb von `hybrid`; `HYBRID_*` weiter abgelehnt | andere |
| M09 · `SPEC-110` Nr. 8 (Nachzug) | `hybrid` ist lieferbar: der Start mit gültiger Umgebung gelingt; ein ungültiger Wert von `MODE` nennt die lieferbaren Werte (alle vier); die Syntax aller Schlüssel wird mitgemeldet | Test Start und Meldung (Text im Katalog); ungültiger Wert mit zugleich ungültigem Schlüssel | Ablehnung von `hybrid` bleibt; Meldung nennt drei Werte | andere |
| M10 · `SPEC-138` Nr. 7, 8 | Start und Herunterfahren im Modus `hybrid` (Prozess): SIGTERM endet mit Exit 0, danach nimmt die Schleife nichts an; ungültige Konfiguration Exit 1 mit `ODR-E2001`, kein Listener | `TestE2EHybrid` (Teil Prozess), Container ohne Netz | Signal nicht an den Hybrid-Pfad gereicht; Exit 0 bei Fehler | andere |
| M11 · `SPEC-037` Nr. 4 | Der Referenzdienst zählt Verbindungen (`StateNew`) und Requests nach Methode, je atomar; die Zähler stimmen mit den gesendeten Requests überein, die bisherigen Zähler (`slice-m2-referenzdienst`) und seine Tests bleiben unverändert | Selbsttest wie `TestReferenzdienstSelbst`: 10 Methoden × Verbindungen und Requests, neue Verbindung je Request und Keep-Alive | zählt jede Methode als `GET`; Verbindungen gleich Requests; Zähler nicht atomar (Detektor, Raten im Absatz *Nebenläufigkeit* nach dem Handoff, `AGENTS.md` §3.10) | nebenläufig |
| M12 · `SPEC-261` Nr. 3 c, 4 a | Kein Auslöser, kein Upstream: Treffer des Mock (lesend und schreibend), fachliche `404`, `409`, `412` und ungültige Anfrage (`ODR-E6005`, `ODR-E6007`) erreichen den Upstream nie, auch bei der offensten Policy (`enabled`, Host, Route, beide Kategorien); 0 Verbindungen | Schleife Fall (Treffer lesend, Treffer schreibend, 404, 409, 412, ungültig) × offenste Policy, Zählprüfung gegen die Liste von `SPEC-261` Nr. 3 c | Auslöser um den Code der fachlichen `404` (`ODR-E6008`) erweitert; Entscheidung am Status statt am Code | andere |
| M13 · `SPEC-261` Nr. 4 c, `SPEC-284` Nr. 8, 9 | Lesend und `ODR-E6003` oder `ODR-E6001` mit `enabled` und freigegebenem Host: genau eine Verbindung, Methode `GET` bzw. `HEAD`, Antwort trägt `X-Recorder-Fallback: upstream`; der verbundene Host ist der der Policy (echter Adapter `httpupstream`, Z35 von `slice-hybrid-service-upstream`) | Schleife {GET, HEAD} × {E6003, E6001}; zweiter Referenzdienst unter anderem Host und Port: 0 Verbindungen dort; Orakel: Verbindungen je Referenzdienst | `Ziel` aus dem `Host`-Header des Clients; falscher Port; Methode fest `GET`; Kennzeichnung fehlt | andere |
| M14 · `LH-FA-12.b`, `SPEC-261` Nr. 4 f, 5 | Lesend ohne Freigabe: `ODR-E6003` unverändert, 0 Verbindungen: Hauptschalter aus bei gesetzten Listen, `enabled` mit fremdem Host, mit Host anderen Ports, `enabled` ohne `UPSTREAM_URL` des Service (Service ohne URL neben einem mit) | Schleife Policy-Zustand (6) × {GET, HEAD}; die Antwort des Mock kehrt byte-gleich zurück | Hauptschalter ignoriert; Host-Vergleich ohne Port; Fehler des Mock ersetzt | andere |
| M15 · `SPEC-261` Nr. 4 c, 6 | Mutierend auf eine nicht unterstützte Operation: ohne Eintrag 0 und `ODR-E6002`; mit passendem Eintrag eine Verbindung, Methode, Pfad und Body byte-gleich; Eintrag für andere Methode oder anderen Pfad: 0; Host nicht freigegeben: 0 | Schleife {POST, PUT, PATCH, DELETE, MERGE} × Eintrag {fehlt, passt, andere Methode, anderer Pfad, Host fremd}; Orakel: unabhängige Funktion „darf senden“ | Eintrag ignoriert (jede mutierende Methode frei); Pfadmuster ohne Segmentgrenze; Body nicht weitergereicht | andere |
| M16 · `SPEC-261` Nr. 4 c, e, `SPEC-284` Nr. 3 d | Replay-Miss `ODR-E5001` mit Opt-in: `GET` und `HEAD` zum freigegebenen Host eine Verbindung; jede mutierende Methode, auch mit Route, 0 und `ODR-E5001` unverändert; ohne Freigabe 0 | Schleife 10 Methoden × {Route, keine} × {Host frei, fremd} | mutierend nach Miss zugelassen; Miss wie `ODR-E6003` behandelt | andere |
| M17 · `SPEC-284` Nr. 3 d, `LH-FA-12.b` | Ein erschöpfter Key (`ODR-E5002`) erreicht den Upstream nie und kehrt unverändert zurück, auch bei der offensten Policy | Fall: Aufzeichnung mit einem Austausch, derselbe Request zweimal, Zähler 0, Code `ODR-E5002` | `ODR-E5002` wie `ODR-E5001` behandelt | andere |
| M18 · `SPEC-261` Nr. 4 b, `SPEC-284` Nr. 6, 7 | Replay-Fallback (Kategorie gesetzt): Antwort aus der Aufzeichnung mit `X-Recorder-Fallback: replay`, 0 Verbindungen, lesend und mutierend; ohne Session-Header `ODR-E5003` aus R, nie aus M | Schleife Methoden (6) × Kategorie (2) × Header {vollständig, fehlt}; Fixture aus Import und Aufzeichnung mit Session | Replay nur lesend; Header-Prüfung vor M | andere |
| M19 · `LH-QA-02`, `AGENTS.md` §3.10 | Der M4-Test läuft mit Referenzdienst, Recorder und Client im selben Prozess unter dem Detektor: Ordnung der Fänger über Kanäle statt Schlaf, die Zähler stimmen bei gleichzeitigen Anfragen | `make test-race` über `tests/integration`; Wirkungs-Fänger: 64 gleichzeitige Anfragen, Summe der Zähler gleich der Zahl der Anfragen | Zähler ohne atomaren Zugriff; Schlaf statt Kanal (Test wird flackernd, Rate im Absatz *Nebenläufigkeit*) | nebenläufig |
| M20 · `SPEC-037` Nr. 2, 3, 7 (Nachzug) | Stufe am gebauten Binary im Container ohne Netz: Mock-Treffer 0, lesender Fallback 1, `POST` ohne Eintrag 0 und mit Eintrag 1, Replay-Miss nur `GET` und `HEAD`; die drei Zeilen der Deklaration stehen im Treiber, `make e2e-abdeckung` führt die Stufe | `TestE2EHybrid` als Stufe 10 in `make test-integration`; Verbindungszähler am Referenzdienst von `tests/e2e` | Name der Stufe im Treiber falsch (`Test … nicht gefunden`); `main` bindet `hybrid` nicht | andere |
| M21 · `SPEC-200`, `SPEC-201` | Handbuch-Kapitel nennt Modus `hybrid`, vier Schlüssel, Fehlercodes des Fallbacks, Wechselwirkung (`SPEC-261` Nr. 4 e), Host-Einschränkung (`SPEC-261` Nr. 5 c), `X-Recorder-Fallback`, Upstream-Links, `ODR-E5003`, Body über `MAX_CAPTURE_BODY`; Beispiele als Datei gegen das Binary | `make handbuch-check` und `make katalog-check` (Katalog der Schlüssel und Codes); Inhalt der Sätze liest der Reviewer, kein Gate fängt ihn (akzeptiertes Negativ, Folge: ein Satz kann dem Ist-Zustand widersprechen, ohne rot zu werden) | Schlüssel `FALLBACK_WRITE_ALLOW` aus dem Handbuch gestrichen (Katalog rot) | andere |

**Entwurf und Rest.** Die Zeilen M06 bis M09 hängen am Nachzug von `SPEC-110` (Aufträge 1 bis 3); M04 an der Reihenfolge der Meldung (§6, Zeile *Reihenfolge der Meldung*); M10 an der Zeile *Start und Herunterfahren* des Rasters. Die Mutationen sind Vorschläge des Planners aus dem Wortlaut der Zusagen; der Implementer ergänzt sie je Fänger und nennt die roten Tests im Bericht. Mengen für Schleifen (Generator-Pflicht): Auslöser und Nicht-Auslöser aus `SPEC-261` Nr. 3, Methoden (10) aus `SPEC-261` Nr. 2, Ursachen aus `SPEC-261` Nr. 8, Zustände der Policy aus `SPEC-261` Nr. 4 und 5; Orakel ist eine unabhängige Fassung der Regel, nicht der Code der Policy.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Nachzug von `SPEC-110`:** Nr. 2 (die vier Schlüssel wandern aus der Ablehnung in die geprüften), Nr. 8 (`hybrid` lieferbar) und Syntax nach Nr. 3 und 6: `HYBRID_FALLBACK` (`disabled` oder `enabled`), `HYBRID_REPLAY_FALLBACK` (Liste der Kategorien), `FALLBACK_ALLOW_HOSTS` (Liste), `FALLBACK_WRITE_ALLOW` (Liste von Einträgen in der Syntax aus `slice-fallback-policy`); Randformen je Schlüssel (leer, unbekannte Kategorie, Doppelte, ungültiger Eintrag).
2. **Zusammenhänge:** die Zusammenhänge aus Auftrag 6 von `slice-fallback-policy`; `UPSTREAM_URL` je Service Pflicht bei `HYBRID_FALLBACK=enabled` (`SPEC-076`); gesetzte `HYBRID_*` und `FALLBACK_*` in den Modi `record`, `replay` und `mock`: ignoriert oder `ODR-E2001`.
3. **`PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid`** (`SPEC-055`, `SPEC-011`): Pflicht und Default wie im Modus `mock`, oder eigene Regel.
4. **Aufbau des M4-Tests:** Referenzdienst, der Verbindungen und Methoden zählt; die Fälle aus §3 der Welle als Tabelle; Fixture aus Import (Baseline) und Aufzeichnung (Recording-Session und Header für den Replay-Fallback); Ordnung unter dem Detektor (Kanäle statt Schlaf, Muster aus `slice-m2-referenzdienst`).
5. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* aus `slice-bootstrap-m3` (Nachzug für `hybrid`, `HYBRID_*`, `FALLBACK_*`), `slice-fallback-policy` (Lesen der Schlüssel) und `slice-fallback-policy-entscheidung` (`*KonfigFehler` wird `ODR-E2001`, `enabled` ohne `UPSTREAM_URL` wird `ODR-E2001`, Wechselwirkung im Handbuch; Planner, 2026-11-08, nach Schnitt von `slice-fallback-policy`): angenommen, §1 *Übernimmt* und DoD (Punkte 1 und 7) tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie. Dazu am 2026-11-18 (Planner): `slice-hybrid-service-upstream` (Host-Fänger am echten Adapter, Z35; Konfig mit `Upstream` und `UpstreamURL`, Handbuch-Sätze zu Links und Kennzeichnung): angenommen, §1 *Übernimmt* und DoD Punkte 1, 2 und 7 tragen sie mit der Kennung des Gebers; der Geber liegt in `done/`, ist also keine Adresse für Rückgaben. *Von diesem Slice an:* die Rückgaben an den Architect (§6) und an den Planner (`harness/README.md`, §6); kein Folge-Slice nimmt etwas an.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-hybrid-service`, `slice-hybrid-service-upstream` und `slice-bootstrap-m3` liegen in `done/`. Die Spezifikation trägt den Nachzug von `SPEC-110`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). **Geprüft am 2026-11-18 (Planner, Zweig `slice-bootstrap-m4`):** die drei Vorgänger liegen in `done/` (`slice-bootstrap-m3` im Archiv `done/welle-mock/`); der Nachzug von `SPEC-110` fehlt noch (Nr. 2 lehnt die vier Schlüsselfamilien und Nr. 8 den Modus `hybrid` ab): **kein Code-Commit vor ihm**, Adresse Architect (§6).

**Warum:** `slice-hybrid-service` und `slice-hybrid-service-upstream` (die zwei Teile des geschnittenen Plans) liefern `HybridUseCase`, den der Root bindet, Teil 2 den Upstream-Fallback, den der M4-Test belegt (*Code-Abhängigkeit*). `slice-bootstrap-m3` liefert den Mock-Pfad im Composition Root (Bau von `MockUseCase`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`, den Nachzug von `SPEC-110` für `mock`) und die Fixture aus Import, Workspace und Referenzdienst, die der M4-Test erweitert (*Code-Abhängigkeit*: dieselben Dateien in `src/bootstrap/`). `ReplayUseCase` und der Upstream-Adapter liegen über `slice-bootstrap-m2` im Root vor.

**Nicht ohne den Mock:** der Slice braucht `welle-mock` ganz (über `slice-bootstrap-m3`) und M3. **Reihenfolge zu `slice-bootstrap-m6`:** beide ändern `src/bootstrap/` und `SPEC-110` Nr. 2; sie laufen nacheinander und nicht parallel, welcher zuerst, ist Priorisierung (`welle-admin-api` ist von `welle-hybrid` unabhängig), und der zweite zieht den Stand des ersten nach (`AGENTS.md` §3.9).

**Größenmarke** (Planner, 2026-11-18, vor dem ersten Code-Commit; Tabelle *Zusage · Fänger* in §3 als Entwurf mit 21 Zeilen): *a* = 2 (M11, M19), *b* = 19, *f* = 2 (Referenzdienst von `tests/integration` und der von `tests/e2e`, je um die Zählung nach Methode erweitert), Produktcode rund 280 (Vorschlag des Planners, der Architect bestätigt oder ändert ihn: Lesen und Prüfen der vier Schlüssel mit Zusammenhängen rund 150, Wahl des Use Case und `hybrid.Konfig` im Root rund 90, Meldungen und Stufe 10 im Treiber rund 40). *S* = 85·2 + 80·19 + 200·2 + 280 = **2370**, *M* = 100 · ⌈135·2370 / 10000⌉ = **3200**. **Schnittfrage nach der Formel: ja** (135·2370 = 319950 > 300000, `SPEC-128` Nr. 1). Geschnitten wird vor dem Code; ein Teil, der keinen Mechanismus teilt, ist vorhanden:

- **Teil 1 (Konfiguration):** Lesen und Prüfen der vier Schlüssel, Zusammenhänge, `*KonfigFehler` und `enabled` ohne `UPSTREAM_URL` als `ODR-E2001`, die Funktion, die aus den Werten die Policy baut, mit Tests allein auf `Konfig` und dieser Funktion (Zeilen M04 bis M09, *b* = 6, Produktcode rund 150, kein Fake): *S* = 480 + 150 = 630, *M* = 900. Ohne Bindung des Modus ändert sich nichts am Binary: `hybrid` bleibt „nicht unterstützt“ (`SPEC-110` Nr. 8).
- **Teil 2 (Bindung und M4-Test):** Modus `hybrid` im Root, `hybrid.Konfig`, M4-Test, Stufe 10, Handbuch (Zeilen M01 bis M03 und M10 bis M21, *a* = 2, *b* = 13, *f* = 2, Produktcode rund 130): *S* = 170 + 1040 + 400 + 130 = 1740, *M* = 2400. Teil 2 braucht Teil 1 (*Code-Abhängigkeit*), schreibt aber dessen Dateien (`konfig*.go`) nicht um; die beiden Teile teilen keinen Mechanismus, denn die Funktion, die die Policy baut, ist die Naht, und Teil 2 ruft sie nur.

Der Planner schneidet nicht allein: ein neuer Slice (`slice-bootstrap-m4-konfig`) ändert Roadmap, `welle-hybrid` §4 und die Rückverfolgungs-Matrix, die der Slice-Zweig nicht schreibt (Übergangsregeln Nr. 4). Die Entscheidung liegt beim User (§6, Zeile *Größe*). Bis dahin steht der Teil des Planner im Block unten nicht (`SPEC-128` Nr. 3 k: *Marke ausstehend*, das Ziel meldet `Marke steht aus: Schätzung 2370, Marke 3200` und fährt grün); sonst, ohne Schnitt, trägt der Block `basis`, `schaetzung: 2370`, `marke: 3200` und einen `schnitt-grund`, der sagt, warum Teil 1 nicht abtrennbar sei. **Basis, falls gesetzt:** der Commit vor dem ersten Code-Commit; die Pfade unten sind bis dahin unberührt (letzter Stand: Start-Move `41422c52`). Ein Merge von `main` in den Zweig, der diese Pfade ändert, zählt in die Messung (`SPEC-128` Nr. 4 g); der Merge steht dann als `ueberschreitung` mit Grund oder als neue Basis mit `ersetzt`.

```groessenmarke
pfad: src/bootstrap/
pfad: tests/integration/
pfad: tests/e2e/
pfad: tools/harness/test-integration.sh
produktcode: 280
fakes: 2
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). **Marke: ausstehend** (Schätzung 2370, Marke nach Formel 3200; die Schnittfrage oben ist vorher zu entscheiden).
- `in-progress` → `open` (blockiert — Carveout?): der Nachzug von `SPEC-110` für `hybrid` liegt nicht vor, oder der Referenzdienst lässt sich nicht nach Methode zählen, ohne die Fixture aus `slice-m2-referenzdienst` zu ändern (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Planner, 2026-11-18, gegen den Stand von `spec/spezifikation.md` am Start des Zweigs): „(offen, Auftrag n)“ heißt, der Architect trägt den Ort in der Spezifikation vor dem ersten Code-Commit ein, die Adresse steht in der Zeile *Offene Zellen* unten; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12). Orte mit „(Nachzug)“ tragen den Wortlaut erst, wenn `SPEC-110` Nr. 2 den Schlüssel nennt; bis dahin lehnt Nr. 2 ihn ab.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Lesen von `HYBRID_FALLBACK` und `HYBRID_REPLAY_FALLBACK` | `SPEC-110` Nr. 1 mit den Defaults `SPEC-010` (`disabled`) und `SPEC-079` (leer) | `SPEC-110` Nr. 1: leer gilt als ungesetzt | `HYBRID_FALLBACK`: `SPEC-010` und `SPEC-110` Nr. 3 (Schreibweise zählt; Nachzug); `HYBRID_REPLAY_FALLBACK`: Eintrag `SPEC-261` Nr. 7 (Ursache `kategorie`, `leer`), Liste `SPEC-110` Nr. 6 (Nachzug); Doppelte `SPEC-261` Nr. 1 a | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-110` Nr. 13 |
| Lesen von `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` | `SPEC-110` Nr. 1 mit den Defaults `SPEC-062` und `SPEC-080` (leer) | `SPEC-110` Nr. 1: leer gilt als ungesetzt; ein leerer Eintrag in der Liste (`a,,b`): (offen, Auftrag 1) | Eintrag: `SPEC-261` Nr. 5 a (Host, Ursache `host`, `leer`) und Nr. 6 a, b (Methode, Muster, Ursachen `trenner`, `methode`, `muster`); Doppelte `SPEC-261` Nr. 1 a; Zerlegung der Liste und Leerraum `SPEC-261` Kopf (Werte), `SPEC-110` Nr. 6 (Nachzug) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-110` Nr. 13 |
| Zusammenhänge der Schlüssel und `UPSTREAM_URL` im Modus `hybrid` | `enabled` ohne `UPSTREAM_URL` in irgendeinem Service: `SPEC-261` Nr. 10 b, `SPEC-076`; der Schlüssel der Meldung bei `SERVICES_FILE` und die Bedingung von `upstream_url` im Modus `hybrid`: (offen, Auftrag 2) | `enabled` mit leerer Allowlist: `SPEC-261` Nr. 8 c (Ursache `ohne-hosts`, Schlüssel `HYBRID_FALLBACK`) | `disabled` mit gesetzten Listen, Freigabe ohne Host, Host ohne Service: kein Fehler, `SPEC-261` Nr. 8 c, 5 e; `ADMIN_*` und `LISTEN_TLS_*`: `SPEC-110` Nr. 2; gesetzte `HYBRID_*` und `FALLBACK_*` in `record`, `replay`, `mock`: (offen, Auftrag 2); `PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid`: (offen, Auftrag 3); Reihenfolge der Meldung: (offen, Auftrag 1) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-110` Nr. 13 |
| Start und Herunterfahren im Modus `hybrid` | Session oder Workspace beim Start: der Modus `mock` verlangt keine (`SPEC-110` Nr. 8), für `hybrid` (offen, Auftrag 1); Session-Header je Request: `SPEC-284` Nr. 6 | gilt nicht: der Start liest keine Werte außer der Umgebung (Zeilen 1 bis 3) | ungültiger `MODE`: `SPEC-110` Nr. 3, 8 | Kontext des Starts: `SPEC-138` Nr. 7, 8; `ctx` des Aufrufs an jede Stufe: `SPEC-284` Nr. 2, 3 e, 12; was `SHUTDOWN_TIMEOUT` mit einer Anfrage tut, die auf den Upstream wartet: (offen, Auftrag 1) | `SPEC-138` Nr. 9 (kein Zustand außerhalb des Aufrufs) |

**Offene Zellen** (Rückgabe an den Architect, Planner, `AGENTS.md` §3.13; Ausgang *offen*, Adresse Architect, vor dem ersten Code-Commit):
(a) leerer Eintrag in einer Liste der vier Schlüssel: `SPEC-110` Nr. 6 nennt ihn ungültig, `SPEC-261` Nr. 5 a, 6 a, 7 nennen ihn als Ursache `leer` der Policy; wer meldet ihn, und gilt Nr. 12 (ein Grund je Schlüssel) in beiden Fällen;
(b) Reihenfolge der Meldung: `SPEC-110` Nr. 12 sortiert nach Schlüssel in Bytefolge (`FALLBACK_ALLOW_HOSTS`, `FALLBACK_WRITE_ALLOW`, `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`), `SPEC-261` Nr. 8 b fügt die Fehler in der Reihenfolge `FALLBACK_ALLOW_HOSTS`, `FALLBACK_WRITE_ALLOW`, `HYBRID_REPLAY_FALLBACK`, dann `HYBRID_FALLBACK`; §1 dieses Plans sagt „folgt dem Fehler“; welche Reihenfolge gilt für die Meldung;
(c) Schlüssel der Meldung für `enabled` ohne `UPSTREAM_URL` (`UPSTREAM_URL` oder `SERVICES_FILE[<i>].upstream_url`) und ob `upstream_url` im Modus `hybrid` nur bei `enabled` Pflicht ist (`SPEC-110` Nr. 9 kennt die Pflicht nur im Modus `record`);
(d) gesetzte `HYBRID_*` und `FALLBACK_*` in den Modi `record`, `replay`, `mock`: ignoriert oder `ODR-E2001` (Auftrag 2), und `PUBLIC_BASE_URL` und `STRICT_COVERAGE` im Modus `hybrid` (Auftrag 3);
(e) Start im Modus `hybrid`: verlangt er eine Session oder einen Workspace, und was geschieht mit einer Anfrage im Upstream-Fallback beim Herunterfahren (`SPEC-138` Nr. 8 b, `SHUTDOWN_TIMEOUT`).

**Zusatztabelle** (Operation meldet mehrere Verstöße, Zusage hängt vom Modus ab; `AGENTS.md` §3.16):

| Fall | Ort | Fänger |
|---|---|---|
| Zwei Verstöße an zwei Schlüsseln der Policy (Paare der vier Schlüssel) | `SPEC-110` Nr. 12, `SPEC-261` Nr. 8 b; Reihenfolge (offen, (b)) | M04, Schleife über alle Paare (6) |
| Zwei Verstöße am selben Schlüssel (leerer Eintrag und ungültiger Eintrag in `FALLBACK_ALLOW_HOSTS`; `enabled` mit leerer Allowlist und fehlender `UPSTREAM_URL`) | `SPEC-110` Nr. 12 (höchstens ein Grund je Schlüssel), `SPEC-261` Nr. 8 b (der erste je Schlüssel), `SPEC-261` Nr. 8 c (`ohne-hosts` entfällt bei ungültigem Eintrag); (offen, (a), (c)) | M04, je Paar ein Fall |
| Modus `record`, `replay`, `mock`, `hybrid` × Schlüsselfamilie | `SPEC-110` Nr. 2 (`ADMIN_*`, `LISTEN_TLS_*`); die vier Familien: (offen, (d)) | M07, M08, Schleife 4 × 6 |
| Modus `hybrid` mit und ohne `STRICT_COVERAGE`, `PUBLIC_BASE_URL` | (offen, (d)) | M07 |

- Der Referenzdienst aus `slice-m2-referenzdienst` zählt Verbindungen und Requests, ob er nach Methode zählt, steht nicht in dessen Plan — **Ausgang:** offen; Adresse: Architect (Auftrag 4); der Planner hat gelesen: `tests/integration/referenzdienst_test.go` zählt nur `zaehlerRequests` auf `$count`, `tests/e2e/referenz_test.go` Verbindungen und Requests ohne Methode; die Erweiterung nach Methode ist Umfang dieses Slice (Zeile M11, Klasse `nebenläufig`, `f` = 2 in der Marke), die bestehenden Tests bleiben unverändert.
- Gesetzte `HYBRID_*` und `FALLBACK_*` in den Modi `record`, `replay` und `mock` (Auftrag 2) — **Ausgang:** offen; Adresse: Architect.
- Der M4-Test läuft mit Referenzdienst, Recorder und Client im selben Prozess unter dem Detektor (`make test-race`) — **Ausgang:** offen; Muster und Ordnung der Fänger wie in `slice-m2-referenzdienst` (Kanäle statt Schlaf).
- `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×): der Hybrid-Pfad im Composition Root erweitert die Pakete, deren Laufzeit Fälle der Gegenproben mitfahren — **Ausgang:** weiter offen → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` im Register.
- Gleichzeitige Änderung von `src/bootstrap/` und `SPEC-110` Nr. 2 durch `slice-bootstrap-m6` — **Ausgang:** offen; Adresse: Planner (Reihenfolge beim Start, keine parallelen Arbeitsbäume auf diesen Dateien).
- Stufennummern im Treiber `tools/harness/test-integration.sh`: `slice-bootstrap-m3-e2e-mock` (Geber, `done/`) nimmt die Stufen 8 und 9; die Stufe dieses Slice nummeriert hinter die vorhandenen und `docs/plan/planning/e2e-abdeckung.md` wird aus `make e2e-abdeckung` neu erzeugt, ohne Konflikt von Hand (`SPEC-037` Nr. 2, 3) — **Ausgang:** offen; Adresse: dieser Slice beim Start, Reihenfolge der Merges mit `slice-bootstrap-m6` ordnet der Planner.
- Größe: die Tabelle liegt als Entwurf vor, die Formel ergibt *S* = 2370 und *M* = 3200 und damit eine Schnittfrage (§4); die Marke steht nicht — **Ausgang:** offen; Adresse: User (Entscheidung: Schnitt in `slice-bootstrap-m4-konfig` und diesen Slice, oder ein Grund, warum der Schnitt nicht geht). Folge je Ausgang trägt der Planner im Integrationsschritt nach (Roadmap, `welle-hybrid` §4 und Rückverfolgungs-Matrix schreibt der Slice-Zweig nicht, Übergangsregeln Nr. 4).
- **Start-Bedingung nicht erfüllt** (Planner, 2026-11-18): `SPEC-110` Nr. 2 und 8 tragen den Nachzug nicht, die Tabelle ist Entwurf, die offenen Zellen oben haben keinen Ort — **Ausgang:** offen; Adresse: Architect (Aufträge 1 bis 5, Zeile *Offene Zellen*); kein Code-Commit vor dem Nachzug (`AGENTS.md` §3.12).
- **Aufträge 4 und 5 an den Architect, Teil `SPEC-037`** (Planner, 2026-11-18): Stufe 10 in `SPEC-037` Nr. 2, 3, 7 (`harness/vertraege/` trägt den Vertrag der Stufen, der Treiber `tools/harness/test-integration.sh` liegt nicht in der Gate-Kette der Übergangsregeln Nr. 7); `SPEC-037` führt `Stand` und `Ungeprüft` für Hybrid noch als fehlend — **Ausgang:** offen; Adresse: Architect.
- **`harness/README.md` nennt „neun Stufen“** in der Zeile von `make test-integration`; mit Stufe 10 stimmt sie nicht mehr. Die Datei ist Gate-Kette (Übergangsregeln Nr. 7); der Slice-Zweig schreibt sie nicht — **Ausgang:** offen; Adresse: Planner (Integrationsschritt, mit `docs/plan/planning/e2e-abdeckung.md` aus `make e2e-abdeckung`).
- **Sendung `slice-hybrid-service-upstream` Z35** (Host-Fänger am echten Adapter; Planner, 2026-11-18, `AGENTS.md` §3.13): §1 *Ausdrücklich NICHT* dieses Slice trägt sie nicht als Ausschluss, §1 *Übernimmt* und DoD Punkt 2 nennen sie jetzt mit der Kennung des Gebers; Fänger M13 — **Ausgang:** eingetreten (angenommen).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, Prozess, Referenzdienst und Client unter Last) sowie `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `AGENTS.md` §3.9; der Nachzug von `SPEC-110` und die Welle-Datei sind Stellen, die eine Berichtigung tragen) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
