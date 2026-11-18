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

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` · `LH-FA-04.c` (Pflicht der `X-Recorder-*`-Header im Modus `hybrid`) · `SPEC-110` Nr. 8, 16 (Start im Modus `hybrid`, Schlüssel) · `SPEC-139` Nr. 9, 12 (Bau, Start, Herunterfahren) · `SPEC-075`, `SPEC-076` · `SPEC-055` (`PUBLIC_BASE_URL` in `hybrid`) · `SPEC-037` Nr. 2, 3, 4, 5, 7, 10 n (Stufe 10) · `ARC-005`, `ARC-012` · `architecture.md` §2.5 (Modus-Routing).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Teil 2 eines vor dem Code geschnittenen Plans:** Teil 1 ist [slice-bootstrap-m4-konfig](../in-progress/slice-bootstrap-m4-konfig.md) (Lesen und Prüfen der Schlüssel `HYBRID_*` und `FALLBACK_*`, Nachzug von `SPEC-110` Nr. 2 und 8). Dieser Slice braucht ihn in `done/` (*Code-Abhängigkeit*, §4).

**Verantwortlich:** pt9912 (Zweig slice-bootstrap-m4)

**Autor:** pt9912. **Datum:** 2026-10-30 (geschnitten am 2026-11-18).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Composition Root bindet im Modus `hybrid` genau `HybridUseCase` an den OData-Listener, reicht die Werte der vier Schlüssel (gelesen und geprüft von `slice-bootstrap-m4-konfig`) über die Policy bis in den Dienst, und der M4-Test aus §3 der Welle läuft grün in `make test`: ein Referenzdienst zählt Verbindungen und Methoden und belegt die Entscheidungsmatrix.

**Übernimmt:** den Zähler des Referenzdienstes nach Methode (Sendung von `slice-m2-referenzdienst`, dessen §1 *Ausdrücklich NICHT* diesen Slice nennt; der Referenzdienst dort zählt Verbindungen und Requests, die Erweiterung ist hier Umfang, Ausgang in §6); das Handbuch-Kapitel mit der Wechselwirkung aus `SPEC-261` Nr. 4 e (wer für eine Kategorie den Replay-Fallback setzt, schaltet den schreibenden Upstream-Fallback für deren Codes ab, weil die Zeile Replay-Miss der Matrix nur `GET` und `HEAD` kennt) und der Einschränkung des Hosts (ein `UPSTREAM_URL` mit Bytes ab `0x80` oder der Zone einer IPv6-Adresse gibt nie einen Fallback frei; der Betreiber setzt die ASCII-Form, Punycode; Hinweis I6 des Reviews von `slice-fallback-policy-entscheidung`, `SPEC-261` Nr. 5 c; Sendung von `slice-fallback-policy-entscheidung`, dessen §1 *Ausdrücklich NICHT* diesen Slice nennt); den Host-Fänger am echten Adapter (Sendung von `slice-hybrid-service-upstream`, Z35, dessen §1 *Ausdrücklich NICHT* diesen Slice nicht nennt: der Referenzdienst zählt die Verbindung unter dem Host, den die Policy verglich, mit dem Adapter `httpupstream` aus dem Composition Root, `SPEC-284` *Ungeprüft* 3); `hybrid.Konfig` mit `Upstream` immer gesetzt und `UpstreamURL` je Service und die Handbuch-Sätze zu Links und Kennzeichnung (Sendungen von `slice-hybrid-service` und `slice-hybrid-service-upstream`); den Aufruf der Funktion, die `slice-bootstrap-m4-konfig` aus den gelesenen Werten die Policy bauen lässt, `Konfiguration.NeuePolicy` (Rückgabe `*policy.Policy`, Fehler `*KonfigFehler` mit `ODR-E2001`; Sendung von `slice-bootstrap-m4-konfig`, dessen §1 *Ausdrücklich NICHT* diesen Slice als Nehmer der Bindung nennt); das Feld `Fallback` der Konfiguration in der Tabelle der Verbraucher von `TestWerteErreichenVerbraucher` in `wurzel_test.go` (es steht dort bis zur Bindung bei den Feldern ohne Verbraucher; Sendung von `slice-bootstrap-m4-konfig`); die Zeilen `PUBLIC_BASE_URL`, `STRICT_COVERAGE` und `ETAG_ENABLED` der Tabelle der Einstellungen im Handbuch nennen den Modus `hybrid` wie den Modus `mock`, denn die Prüfung verlangt `PUBLIC_BASE_URL` bei fremdem Host schon in `hybrid` (Sendung von `slice-bootstrap-m4-konfig`, Review F-7, F-8).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Lesen und Prüfen der Schlüssel `HYBRID_*` und `FALLBACK_*`, `*KonfigFehler` als `ODR-E2001`, `enabled` ohne `UPSTREAM_URL`, der Nachzug von `SPEC-110` Nr. 2 und 8 und die Zeilen der Schlüssel in der Tabelle der Einstellungen — `slice-bootstrap-m4-konfig` (Teil 1): dieser Slice ruft nur die Funktion, die die Policy baut, und schreibt `konfig*.go` nicht um.
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

- [ ] Verdrahtung: der Modus `hybrid` bindet `HybridUseCase`; der Start mit gültiger Umgebung gelingt; die vier Schlüssel kommen über `Konfiguration.NeuePolicy` aus `slice-bootstrap-m4-konfig` bis in die Policy (Fänger, dass der Root sie weiterreicht und den Nullwert nicht stehen lässt; Muster aus `slice-upstream-tls`); das Feld `Fallback` steht in der Tabelle der Verbraucher von `TestWerteErreichenVerbraucher` (`wurzel_test.go`) und nicht mehr bei den Feldern ohne Verbraucher (Sendung von `slice-bootstrap-m4-konfig`); der Root übergibt `hybrid.Konfig` nach `SPEC-284` Nr. 1: `MaxBytes` gleich `MAX_CAPTURE_BODY`, `Upstream` immer gesetzt, `Policy` aus `policy.New`, `UpstreamURL` je Service (leer ohne `UPSTREAM_URL`) (Sendung von `slice-hybrid-service`; die Felder `Upstream` und `UpstreamURL` liefert `slice-hybrid-service-upstream`, das den Pfad mit trägt).
- [ ] M4-Test aus §3 der Welle (dazu als Stufe von `make test-integration` am gebauten Binary das Szenario des Verbindungszählers, `SPEC-037`, für die Kriterien, die die [Rückverfolgungs-Matrix](../rueckverfolgung.md) diesem Slice zuweist): ein Treffer des Mock ergibt 0 Verbindungen; eine lesende Anfrage bei unvollständiger Coverage mit Freigabe genau eine lesende Verbindung, ohne Freigabe `ODR-E6003` und 0; ein POST auf eine nicht unterstützte Operation ohne Eintrag in `FALLBACK_WRITE_ALLOW` 0, mit passendem Eintrag eine; ein Replay-Miss `ODR-E5001` mit Opt-in nur für GET und HEAD zum freigegebenen Host, ein erschöpfter Key `ODR-E5002` und fachliche `404` und `409` 0; die Verbindung des lesenden Fallbacks geht mit dem echten Adapter `httpupstream` an den Host, den die Policy verglich (Sendung `slice-hybrid-service-upstream`, Z35).
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand: die Zeilen `PUBLIC_BASE_URL`, `STRICT_COVERAGE` und `ETAG_ENABLED` der Einstellungen gelten in den Modi `mock` und `hybrid`, Sendung von `slice-bootstrap-m4-konfig`; Modus `hybrid`, Schlüssel `HYBRID_*` und `FALLBACK_*` (die Zeilen der Einstellungen liefert `slice-bootstrap-m4-konfig`), die Fehlercodes des Fallbacks, die Wechselwirkung von Replay-Fallback und schreibendem Upstream-Fallback aus `SPEC-261` Nr. 4 e und die Einschränkung des Hosts (Bytes ab `0x80` und Zone im `UPSTREAM_URL` geben nie einen Fallback frei, Punycode setzen; `SPEC-261` Nr. 5 c) (beides Sendung von `slice-fallback-policy-entscheidung`), belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Das Kapitel nennt auch das Verhalten des Hybrid-Dienstes nach `SPEC-284`: den Header `X-Recorder-Fallback` einer Fallback-Antwort, dass absolute Upstream-Links einer Fallback-Antwort unverändert bleiben (`LINK_REWRITE` ist `off`), dass `ODR-E5003` fehlender Session-Header erst im Replay-Fallback entsteht und dass ein Request mit Body über `MAX_CAPTURE_BODY` nie einen Fallback erhält (Sendung von `slice-hybrid-service`). Voraussetzung ist `slice-handbuch-meldungskatalog` in `done/` (Trigger von `welle-mock`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `src/bootstrap/` | update | Wahl des Use Case je Modus um `hybrid` erweitert (`vorbereiten` und `wurzel.go`); `hybrid.Konfig` bauen; die Funktion aus `slice-bootstrap-m4-konfig` rufen; `konfig*.go` bleibt unberührt |
| `src/adapters/driving/httpodata/` | — | kein Änderungsbedarf erwartet, der Adapter kennt den Modus nicht; ein Bedarf geht an den Architect (§6) |
| `src/bootstrap/` (Tests) | neu / update | `m4_wurzel_test.go` (Bindung des Modus, `hybrid.Konfig`, Wirkung je Schlüssel am Referenzdienst, Warnung und Werte des Upstream-Clients); Muster `m3_wurzel_test.go`; der Fall `hybrid` in `wurzel_test.go`, der heute `ErrUngueltig` erwartet, kippt auf Erfolg |
| `tests/integration/` | neu / update | `m4_test.go`: der M4-Test im Prozess von `go test` mit `bootstrap.Build`, dem echten Adapter `httpupstream` und dem Referenzdienst; der Referenzdienst (`referenzdienst_test.go`) zählt Verbindungen und Requests, die Erweiterung nach Methode ist Umfang (§6) |
| `tests/e2e/` | neu / update | Stufe `TestE2EHybrid` am gebauten Binary (`stufe_hybrid_test.go`) mit dem eigenen Referenzdienst `referenz_test.go` (`SPEC-037` Nr. 4; Zähler nach Methode ist Umfang) |
| `tools/harness/test-integration.sh` | update | Stufe 10 mit den drei Zeilen der Deklaration (`SPEC-037` Nr. 3); `docs/plan/planning/e2e-abdeckung.md` wird aus `make e2e-abdeckung` neu erzeugt; `harness/README.md` (Zeile `make test-integration`, „neun Stufen“) ist Gate-Kette und geht an den Integrationsschritt (§6) |
| `docs/user/` | update | Handbuch-Kapitel und Beispiele als Datei (DoD Punkt 3); die Zeilen `MODE`, `UPSTREAM_URL`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`, `ETAG_ENABLED` und die vier Zeilen aus `slice-bootstrap-m4-konfig` nennen den Modus `hybrid` als lieferbar und seine Wirkung (`SPEC-110` Nr. 14, 15, 16); kein Pfad der Größenmarke (`SPEC-128` Nr. 3 m) |
| `harness/vertraege/SPEC-037-test-integration.md` | update (Architect) | Stufe 10 in `SPEC-037` Nr. 4, 5, 7, 10 n (Architect, 2026-11-18; Nr. 2 und 3 tragen sie ohne Änderung); Verträge und Pläne zählen nicht in die Marke |
| `spec/spezifikation.md` | update (Architect) | `SPEC-139` Nr. 2, 5, 9, 12 (Bau, Start ohne Session, wartende Anfrage beim Herunterfahren) und `SPEC-110` Nr. 16 (Architect, 2026-11-18); zählt nicht in die Marke |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. **Stand der Tabelle** (Planner, 2026-11-18, nach dem Schnitt; die Zeilen M04 bis M09 liegen bei `slice-bootstrap-m4-konfig`, die Nummern bleiben): der Architect hat jede Zeile geprüft und die Orte nachgetragen (M22 ist neu), jede Zeile nennt Zusage, Fänger, Mutation und Klasse. Ein Fänger nennt seinen Test mit dem Kurznamen ohne `Test`; Dateien siehe oben. Jede Mutation, die ein Fänger fängt, fährt der Implementer selbst (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| M01 · `SPEC-075`, `SPEC-284`, `SPEC-110` Nr. 8, `SPEC-139` Nr. 12 a, b | Der Modus `hybrid` bindet genau den Hybrid-Dienst (`HybridUseCase`) an den OData-Listener, der Start mit gültiger Umgebung gelingt (Rest von M09 aus Teil 1); `mock` und `replay` binden ihn nie | Tabellentest Modus × Anfrage mit Fallback-Auslöser: die Antwort trägt im Modus `hybrid` den Header `X-Recorder-Fallback`, in `mock` und `replay` nie; Start im Modus `hybrid` ohne Fehler; nach dem Start gibt es keine Session, und `Sessions()`, `Record()` und `Import()` sind der Nullwert der Schnittstelle | `hybrid` bindet `MockUseCase`; `mock` bindet den Hybrid-Dienst; `vorbereiten` lehnt `hybrid` weiter ab; `Hochfahren` legt im Modus `hybrid` eine Session an | andere |
| M02 · `SPEC-261` Nr. 10 b | Jeder der vier Schlüssel erreicht `policy.Werte` und wirkt; der Nullwert bleibt nie stehen | Schleife über die vier Schlüssel × {gesetzt, ungesetzt} (8 Fälle, Zählprüfung gegen die Menge der Schlüssel), Orakel: erwartete Verbindungszahl nach `SPEC-261` Nr. 4, Wirkung am Referenzdienst | Schlüssel nicht gelesen; `Hosts` und `Schreibfreigaben` vertauscht; `Upstream` fest wahr oder falsch; `ReplayKategorien` nicht übergeben | andere |
| M03 · `SPEC-284` Nr. 1, 5 c, `SPEC-139` Nr. 12 a | Der Root übergibt `hybrid.Konfig`: `MaxBytes` gleich `MAX_CAPTURE_BODY` (Default und gesetzter Wert), `Upstream` immer gesetzt, auch bei `disabled`, `Policy` aus `policy.New`, `UpstreamURL` je Service (leer ohne `UPSTREAM_URL`) | Body mit `Laenge` an beiden Seiten der Grenze (−1, gleich, +1) bei Default und gesetztem Wert: bis zur Grenze Fallback, darüber nie; zwei Services, einer ohne `UPSTREAM_URL`: `Ziel` je Service am Referenzdienst, der ohne URL erreicht ihn nie | `MaxBytes` fest; `UpstreamURL` des ersten Services für alle; `Upstream` nil bei `disabled` | andere |
| M10 · `SPEC-138` Nr. 7, 8, `SPEC-139` Nr. 12 c (das Zusammenspiel der Frist mit einer wartenden Upstream-Anfrage ist akzeptiertes Negativ, `SPEC-037` Nr. 10 n) | Start und Herunterfahren im Modus `hybrid` (Prozess): SIGTERM endet mit Exit 0, danach nimmt die Schleife nichts an; ungültige Konfiguration Exit 1 mit `ODR-E2001`, kein Listener | `TestE2EHybrid` (Teil Prozess), Container ohne Netz | Signal nicht an den Hybrid-Pfad gereicht; Exit 0 bei Fehler | andere |
| M11 · `SPEC-037` Nr. 4 | Der Referenzdienst zählt Verbindungen (`StateNew`) und Requests nach Methode, je atomar; die Zähler stimmen mit den gesendeten Requests überein, die bisherigen Zähler (`slice-m2-referenzdienst`) und seine Tests bleiben unverändert | Selbsttest wie `TestReferenzdienstSelbst`: 10 Methoden × Verbindungen und Requests, neue Verbindung je Request und Keep-Alive | zählt jede Methode als `GET`; Verbindungen gleich Requests; Zähler nicht atomar (Detektor, Raten im Absatz *Nebenläufigkeit* nach dem Handoff, `AGENTS.md` §3.10) | nebenläufig |
| M12 · `SPEC-261` Nr. 3 c, 4 a | Kein Auslöser, kein Upstream: Treffer des Mock (lesend und schreibend), fachliche `404`, `409`, `412` und ungültige Anfrage (`ODR-E6005`, `ODR-E6007`) erreichen den Upstream nie, auch bei der offensten Policy (`enabled`, Host, Route, beide Kategorien); 0 Verbindungen | Schleife Fall (Treffer lesend, Treffer schreibend, 404, 409, 412, ungültig) × offenste Policy, Zählprüfung gegen die Liste von `SPEC-261` Nr. 3 c | Auslöser um den Code der fachlichen `404` (`ODR-E6008`) erweitert; Entscheidung am Status statt am Code | andere |
| M13 · `SPEC-261` Nr. 4 c, `SPEC-284` Nr. 8, 9 | Lesend und `ODR-E6003` oder `ODR-E6001` mit `enabled` und freigegebenem Host: genau eine Verbindung, Methode `GET` bzw. `HEAD`, Antwort trägt `X-Recorder-Fallback: upstream`; der verbundene Host ist der der Policy (echter Adapter `httpupstream`, Z35 von `slice-hybrid-service-upstream`) | Schleife {GET, HEAD} × {E6003, E6001}; zweiter Referenzdienst unter anderem Host und Port: 0 Verbindungen dort; Orakel: Verbindungen je Referenzdienst | `Ziel` aus dem `Host`-Header des Clients; falscher Port; Methode fest `GET`; Kennzeichnung fehlt | andere |
| M14 · `LH-FA-12.b`, `SPEC-261` Nr. 4 f, 5 | Lesend ohne Freigabe: `ODR-E6003` unverändert, 0 Verbindungen: Hauptschalter aus bei gesetzten Listen, `enabled` mit fremdem Host, mit Host anderen Ports, `enabled` ohne `UPSTREAM_URL` des Service (Service ohne URL neben einem mit) | Schleife Policy-Zustand (6) × {GET, HEAD}; die Antwort des Mock kehrt byte-gleich zurück | Hauptschalter ignoriert; Host-Vergleich ohne Port; Fehler des Mock ersetzt | andere |
| M15 · `SPEC-261` Nr. 4 c, 6 | Mutierend auf eine nicht unterstützte Operation: ohne Eintrag 0 und `ODR-E6002`; mit passendem Eintrag eine Verbindung, Methode, Pfad und Body byte-gleich; Eintrag für andere Methode oder anderen Pfad: 0; Host nicht freigegeben: 0 | Schleife {POST, PUT, PATCH, DELETE, MERGE} × Eintrag {fehlt, passt, andere Methode, anderer Pfad, Host fremd}; Orakel: unabhängige Funktion „darf senden“ | Eintrag ignoriert (jede mutierende Methode frei); Pfadmuster ohne Segmentgrenze; Body nicht weitergereicht | andere |
| M16 · `SPEC-261` Nr. 4 c, e, `SPEC-284` Nr. 3 d | Replay-Miss `ODR-E5001` mit Opt-in: `GET` und `HEAD` zum freigegebenen Host eine Verbindung; jede mutierende Methode, auch mit Route, 0 und `ODR-E5001` unverändert; ohne Freigabe 0 | Schleife 10 Methoden × {Route, keine} × {Host frei, fremd} | mutierend nach Miss zugelassen; Miss wie `ODR-E6003` behandelt | andere |
| M17 · `SPEC-284` Nr. 3 d, `LH-FA-12.b` | Ein erschöpfter Key (`ODR-E5002`) erreicht den Upstream nie und kehrt unverändert zurück, auch bei der offensten Policy | Fall: Aufzeichnung mit einem Austausch, derselbe Request zweimal, Zähler 0, Code `ODR-E5002` | `ODR-E5002` wie `ODR-E5001` behandelt | andere |
| M18 · `SPEC-261` Nr. 4 b, `SPEC-284` Nr. 6, 7 | Replay-Fallback (Kategorie gesetzt): Antwort aus der Aufzeichnung mit `X-Recorder-Fallback: replay`, 0 Verbindungen, lesend und mutierend; ohne Session-Header `ODR-E5003` aus R, nie aus M | Schleife Methoden (6) × Kategorie (2) × Header {vollständig, fehlt}; Fixture aus Import und Aufzeichnung mit Session | Replay nur lesend; Header-Prüfung vor M | andere |
| M19 · `LH-QA-02`, `AGENTS.md` §3.10 | Der M4-Test läuft mit Referenzdienst, Recorder und Client im selben Prozess unter dem Detektor: Ordnung der Fänger über Kanäle statt Schlaf, die Zähler stimmen bei gleichzeitigen Anfragen | `make test-race` über `tests/integration`; Wirkungs-Fänger: 64 gleichzeitige Anfragen, Summe der Zähler gleich der Zahl der Anfragen | Zähler ohne atomaren Zugriff; Schlaf statt Kanal (Test wird flackernd, Rate im Absatz *Nebenläufigkeit*) | nebenläufig |
| M20 · `SPEC-037` Nr. 2, 3, 4, 5, 7 | Stufe am gebauten Binary im Container ohne Netz: Mock-Treffer 0, lesender Fallback 1, `POST` ohne Eintrag 0 und mit Eintrag 1, Replay-Miss nur `GET` und `HEAD`; `enabled` ohne Host und `enabled` ohne `UPSTREAM_URL` enden mit Exit 1, `ODR-E2001` und ohne Listener; die drei Zeilen der Deklaration stehen im Treiber, `make e2e-abdeckung` führt die Stufe | `TestE2EHybrid` als Stufe 10 in `make test-integration`; Verbindungszähler am Referenzdienst von `tests/e2e` | Name der Stufe im Treiber falsch (`Test … nicht gefunden`); `main` bindet `hybrid` nicht | andere |
| M21 · `SPEC-200`, `SPEC-201` | Handbuch-Kapitel nennt Modus `hybrid`, vier Schlüssel, Fehlercodes des Fallbacks, Wechselwirkung (`SPEC-261` Nr. 4 e), Host-Einschränkung (`SPEC-261` Nr. 5 c), `X-Recorder-Fallback`, Upstream-Links, `ODR-E5003`, Body über `MAX_CAPTURE_BODY`; Beispiele als Datei gegen das Binary | `make handbuch-check` und `make katalog-check` (Katalog der Schlüssel und Codes); Inhalt der Sätze liest der Reviewer, kein Gate fängt ihn (akzeptiertes Negativ, Folge: ein Satz kann dem Ist-Zustand widersprechen, ohne rot zu werden) | Schlüssel `FALLBACK_WRITE_ALLOW` aus dem Handbuch gestrichen (Katalog rot) | andere |
| M22 · `SPEC-139` Nr. 9, 12 a, `SPEC-110` Nr. 16 a | Im Modus `hybrid` erreichen `UPSTREAM_TIMEOUT`, `UPSTREAM_TLS_VERIFY` und `UPSTREAM_CA_FILE` den Upstream-Client wie im Modus `record`; die Warnung `tls_pruefung_aus` steht genau einmal bei `hybrid`, `enabled` und `UPSTREAM_TLS_VERIFY=false`, sonst nie, in `record` wie bisher | Schleife Modus (4) × `HYBRID_FALLBACK` (2) × `UPSTREAM_TLS_VERIFY` (2) zählt die Einträge `tls_pruefung_aus` (Orakel: `record`, oder `hybrid` mit `enabled`, jeweils bei `false`: 1, sonst 0); die drei Werte am Client nach dem Muster von `slice-upstream-tls` | Warnung auch ohne `enabled`; Warnung fehlt in `hybrid`; ein Wert des Clients fest | andere |

**Entwurf und Rest.** Die Orte hat der Architect am 2026-11-18 nachgetragen (§6); M22 stammt von ihm (Folge von `SPEC-139` Nr. 9, 12 a) und steht nicht in der Schätzung des Planners: mit ihr ist *b* = 14, *S* = 1820 und die Marke nach `SPEC-128` Nr. 1 gleich 2500, die der Planner beim Start setzt (§6). Die Mutationen sind Vorschläge des Planners aus dem Wortlaut der Zusagen; der Implementer ergänzt sie je Fänger und nennt die roten Tests im Bericht. Mengen für Schleifen (Generator-Pflicht): Auslöser und Nicht-Auslöser aus `SPEC-261` Nr. 3, Methoden (10) aus `SPEC-261` Nr. 2, Ursachen aus `SPEC-261` Nr. 8, Zustände der Policy aus `SPEC-261` Nr. 4 und 5; Orakel ist eine unabhängige Fassung der Regel, nicht der Code der Policy.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12; der Nachzug von `SPEC-110` Nr. 2 und 8, die Zusammenhänge und die Syntax liegen bei `slice-bootstrap-m4-konfig`):

1. **Start und Herunterfahren im Modus `hybrid`:** verlangt der Start eine Session oder einen Workspace (der Modus `mock` verlangt keine, `SPEC-110` Nr. 8); was `SHUTDOWN_TIMEOUT` mit einer Anfrage tut, die auf den Upstream wartet (`SPEC-138` Nr. 8 b).
2. **Aufbau des M4-Tests:** Referenzdienst, der Verbindungen und Methoden zählt; die Fälle aus §3 der Welle als Tabelle; Fixture aus Import (Baseline) und Aufzeichnung (Recording-Session und Header für den Replay-Fallback); Ordnung unter dem Detektor (Kanäle statt Schlaf, Muster aus `slice-m2-referenzdienst`).
3. Tabelle *Zusage · Fänger*, Raster in §6, Stufe 10 in `SPEC-037` Nr. 2, 3, 7.

Erledigt am 2026-11-18 (Architect): die Aufträge 1 bis 3; die Orte stehen in `SPEC-139` Nr. 12 und `SPEC-037` Nr. 4, 5, 7, 10 n und in §6, die Entscheidungen ebenda.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen und am 2026-11-18 nach dem Schnitt erneut: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* aus `slice-hybrid-service` (Konfig mit `MaxBytes`, `Policy`), `slice-hybrid-service-upstream` (Host-Fänger am echten Adapter, Z35; Konfig mit `Upstream` und `UpstreamURL`, Handbuch-Sätze zu Links und Kennzeichnung), `slice-m2-referenzdienst` (Zähler nach Methode) und `slice-fallback-policy-entscheidung` (Wechselwirkung und Host-Einschränkung im Handbuch): angenommen, §1 *Übernimmt* und DoD Punkte 1 bis 3 tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* trifft sie; die Geber liegen in `done/`, sind also keine Adresse für Rückgaben. Die Sendungen von `slice-bootstrap-m3`, `slice-fallback-policy` und `slice-fallback-policy-entscheidung` zum Lesen der Schlüssel, zum Nachzug von `SPEC-110` und zu `*KonfigFehler` trägt `slice-bootstrap-m4-konfig` (dessen §1 *Übernimmt* und DoD Punkte 1 und 2). *Von diesem Slice an:* die Rückgaben an den Architect (§6) und an den Planner (`harness/README.md`, §6); kein Folge-Slice nimmt etwas an. *An `slice-bootstrap-m4-konfig` hat dieser Slice nichts zu geben außer dem Aufruf der Funktion in §1.*

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `next` → `in-progress`): `slice-hybrid-service`, `slice-hybrid-service-upstream`, `slice-bootstrap-m3` und `slice-bootstrap-m4-konfig` liegen in `done/`. Die Spezifikation trägt die Aufträge 1 und 3 (Start im Modus `hybrid`, Stufe 10 in `SPEC-037`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): **Stand 2026-11-18 (Architect): liegt vor** (`SPEC-139` Nr. 12, `SPEC-037` Nr. 4, 5, 7, 10 n). **Geprüft am 2026-11-18 (Planner, Zweig `slice-bootstrap-m4`):** `slice-hybrid-service`, `slice-hybrid-service-upstream` und `slice-bootstrap-m3` liegen in `done/` (`slice-bootstrap-m3` im Archiv `done/welle-mock/`); `slice-bootstrap-m4-konfig` liegt in `in-progress/`; die Aufträge an den Architect sind erledigt (§6, 2026-11-18).

**Warum:** `slice-hybrid-service` und `slice-hybrid-service-upstream` (die zwei Teile des geschnittenen Plans) liefern `HybridUseCase`, den der Root bindet, Teil 2 den Upstream-Fallback, den der M4-Test belegt (*Code-Abhängigkeit*). `slice-bootstrap-m4-konfig` liefert Lesen und Prüfen der vier Schlüssel und die Funktion, die die Policy baut (*Code-Abhängigkeit*; dieser Slice schreibt dessen Dateien `konfig*.go` nicht um, die Naht ist die Funktion). `slice-bootstrap-m3` liefert den Mock-Pfad im Composition Root (Bau von `MockUseCase`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`) und die Fixture aus Import, Workspace und Referenzdienst, die der M4-Test erweitert (*Code-Abhängigkeit*: dieselben Dateien in `src/bootstrap/`). `ReplayUseCase` und der Upstream-Adapter liegen über `slice-bootstrap-m2` im Root vor.

**Nicht ohne den Mock:** der Slice braucht `welle-mock` ganz (über `slice-bootstrap-m3`) und M3. **Reihenfolge zu `slice-bootstrap-m6`:** beide ändern `src/bootstrap/` und `SPEC-110` Nr. 2; sie laufen nacheinander und nicht parallel, welcher zuerst, ist Priorisierung (`welle-admin-api` ist von `welle-hybrid` unabhängig), und der zweite zieht den Stand des ersten nach (`AGENTS.md` §3.9).

**Größenmarke** (Planner, 2026-11-18, vor dem ersten Code-Commit; Teil 2 nach dem Schnitt, Tabelle *Zusage · Fänger* in §3 als Entwurf mit 15 Zeilen): *a* = 2 (M11, M19), *b* = 13, *f* = 2 (Referenzdienst von `tests/integration` und der von `tests/e2e`, je um die Zählung nach Methode erweitert), Produktcode rund 130 (Vorschlag des Planners, der Architect bestätigt oder ändert ihn: Wahl des Use Case und `hybrid.Konfig` im Root rund 90, Meldungen und Stufe 10 im Treiber rund 40). *S* = 85·2 + 80·13 + 200·2 + 130 = **1740**, *M* = 100 · ⌈135·1740 / 10000⌉ = **2400**. Keine Schnittfrage (135·1740 = 234900 ≤ 300000, `SPEC-128` Nr. 1). Die ungeschnittene Schätzung war 2370 (Marke 3200, Schnittfrage ja); der Schnitt vor dem Code trennt Teil 1 (*S* 630, Marke 900, `slice-bootstrap-m4-konfig`) ab, der keinen Mechanismus mit diesem Teil teilt: die Funktion, die die Policy baut, ist die Naht, und dieser Teil ruft sie nur. Der Block trägt `basis`, `schaetzung` und `marke`, sobald der Planner beim Start die Basis setzt (`SPEC-128` Nr. 3 k: bis dahin *Marke ausstehend*; das Ziel liegt für Pläne in `open/` nicht im Lauf). **Basis:** der Commit vor dem ersten Code-Commit; ein Merge von `main` in den Zweig, der diese Pfade ändert, zählt in die Messung (`SPEC-128` Nr. 4 g); der Merge steht dann als `ueberschreitung` mit Grund oder als neue Basis mit `ersetzt`.

```groessenmarke
pfad: src/bootstrap/
pfad: tests/integration/
pfad: tests/e2e/
pfad: tools/harness/test-integration.sh
produktcode: 130
fakes: 2
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke 2400 (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); die Marke setzt der Planner vor dem ersten Code-Commit nach `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): `slice-bootstrap-m4-konfig` liegt nicht in `done/`, oder der Referenzdienst lässt sich nicht nach Methode zählen, ohne die Fixture aus `slice-m2-referenzdienst` zu ändern (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Planner, 2026-11-18; die Orte trägt der Architect am 2026-11-18 nach, keine Zelle ist offen; die Zeilen zum Lesen der Schlüssel und zu den Zusammenhängen liegen bei `slice-bootstrap-m4-konfig`): eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Start und Herunterfahren im Modus `hybrid` | Session oder Workspace beim Start: nicht verlangt, `SPEC-139` Nr. 12 b (Session-Header je Request: `SPEC-284` Nr. 6, Workspace-Header: `SPEC-195` Nr. 5) | gilt nicht: der Start liest keine Werte außer der Umgebung (Lesen der Schlüssel: `slice-bootstrap-m4-konfig`) | ungültiger `MODE`: `SPEC-110` Nr. 3, 8; Anfrage mit unbekanntem Workspace bei einem Start ohne Baseline: `SPEC-139` Nr. 12 b (`ODR-E7001`) | Kontext des Starts: `SPEC-138` Nr. 7, 8; `ctx` des Aufrufs an jede Stufe: `SPEC-284` Nr. 2, 3 e, 12; Anfrage, die beim Herunterfahren auf den Upstream wartet, und `SHUTDOWN_TIMEOUT`: `SPEC-139` Nr. 12 c | `SPEC-138` Nr. 9 (kein Zustand außerhalb des Aufrufs) |

**Entscheidungen des Architect** (2026-11-18; Ausgang *entschieden*; die Buchstaben sind die der Rückgabe des Planners; (a) bis (d), (f) und (g) stehen in `slice-bootstrap-m4-konfig` §6):
(e) Start im Modus `hybrid`: weder Session noch Workspace beim Start (in `Hochfahren` entfällt Schritt (3), es gibt keinen Recorder, keinen `SessionUseCase` und keinen Import); Workspace und Session-Header wählt der Client je Request. Herunterfahren: die Reihenfolge von `SPEC-139` Nr. 7 ohne Schritt (2), `rest` ist `SHUTDOWN_TIMEOUT`; eine Anfrage, die im Upstream-Fallback wartet, ist eine laufende Anfrage wie jede andere: `Schliessen` wartet höchstens `rest`, danach endet ihr Kontext und `Senden` kehrt mit dem Kontextfehler zurück, ohne Antwort; keine eigene, längere Frist. Das Zusammenspiel der Frist mit dem Fallback prüft kein Test (akzeptiertes Negativ mit Grund und Folge, `SPEC-037` Nr. 10 n); beide Hälften haben je ihren Fänger. Ort: `SPEC-139` Nr. 12 b, c.
Weitere Entscheidungen für diesen Teil: der Root baut im Modus `hybrid` den Upstream-Client immer, auch bei `disabled` (`hybrid.Konfig.Upstream` ist immer gesetzt, `SPEC-284` Nr. 1); `UPSTREAM_TIMEOUT`, `UPSTREAM_TLS_VERIFY` und `UPSTREAM_CA_FILE` wirken wie im Modus `record`; die Warnung `tls_pruefung_aus` gibt es im Modus `hybrid` nur bei `HYBRID_FALLBACK=enabled` (Zeile M22). Ort: `SPEC-139` Nr. 9, 12 a.

- Der Referenzdienst aus `slice-m2-referenzdienst` zählt Verbindungen und Requests, ob er nach Methode zählt, steht nicht in dessen Plan — **Ausgang:** entschieden: er zählt Requests je Methode (Zeile M11); Ort für `tests/e2e`: `SPEC-037` Nr. 4; der Referenzdienst von `tests/integration` ist Testhilfe ohne Vertragstext, dort gilt die Zeile M11. Der Planner hat gelesen: `tests/integration/referenzdienst_test.go` zählt nur `zaehlerRequests` auf `$count`, `tests/e2e/referenz_test.go` Verbindungen und Requests ohne Methode; die Erweiterung nach Methode ist Umfang dieses Slice (Zeile M11, Klasse `nebenläufig`, `f` = 2 in der Marke), die bestehenden Tests bleiben unverändert.
- Der M4-Test läuft mit Referenzdienst, Recorder und Client im selben Prozess unter dem Detektor (`make test-race`) — **Ausgang:** entschieden: Muster und Ordnung der Fänger wie in `slice-m2-referenzdienst` (Kanäle statt Schlaf), Fänger Zeile M19; Fixture aus Import (Baseline mit unvollständiger Coverage) und Aufzeichnung (Recording-Session mit Austauschen, Header für den Replay-Fallback).
- `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×): der Hybrid-Pfad im Composition Root erweitert die Pakete, deren Laufzeit Fälle der Gegenproben mitfahren — **Ausgang:** weiter offen → `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` im Register.
- Gleichzeitige Änderung von `src/bootstrap/` und `SPEC-110` Nr. 2 durch `slice-bootstrap-m6` — **Ausgang:** offen; Adresse: Planner (Reihenfolge beim Start, keine parallelen Arbeitsbäume auf diesen Dateien).
- Stufennummern im Treiber `tools/harness/test-integration.sh`: `slice-bootstrap-m3-e2e-mock` (Geber, `done/`) nimmt die Stufen 8 und 9; die Stufe dieses Slice nummeriert hinter die vorhandenen und `docs/plan/planning/e2e-abdeckung.md` wird aus `make e2e-abdeckung` neu erzeugt, ohne Konflikt von Hand (`SPEC-037` Nr. 2, 3) — **Ausgang:** offen; Adresse: dieser Slice beim Start, Reihenfolge der Merges mit `slice-bootstrap-m6` ordnet der Planner.
- Größe: die ungeschnittene Schätzung 2370 (Marke 3200) ergab eine Schnittfrage — **Ausgang:** eingetreten (Schnitt vor dem Code, Entscheidung des Users 2026-11-18): Teil 1 `slice-bootstrap-m4-konfig` (*S* 630, Marke 900), dieser Teil *S* 1740, Marke 2400; Roadmap, `welle-hybrid` §4 und `rueckverfolgung.md` im Zweig nachgezogen.
- **Start-Bedingung** (Planner, 2026-11-18): die Aufträge 1 und 3 an den Architect sind erledigt (`SPEC-139` Nr. 12, `SPEC-037`) — **Ausgang:** entschieden; offen bleibt der Trigger von §4: `slice-bootstrap-m4-konfig` muss in `done/` liegen, bevor dieser Slice nach `in-progress` geht.
- **Aufträge 2 und 3 an den Architect, Teil `SPEC-037`** (Planner, 2026-11-18): Stufe 10 in `SPEC-037` Nr. 2, 3, 7 (`harness/vertraege/` trägt den Vertrag der Stufen, der Treiber `tools/harness/test-integration.sh` liegt nicht in der Gate-Kette der Übergangsregeln Nr. 7); `SPEC-037` führt `Stand` und `Ungeprüft` für Hybrid noch als fehlend — **Ausgang:** entschieden: Stufe 10 (`TestE2EHybrid`) in `SPEC-037` Nr. 7, Zähler nach Methode in Nr. 4, `MODE=hybrid` startet in Nr. 5, `Stand` und das akzeptierte Negativ in Nr. 10 n; Nr. 2 und 3 tragen die Stufe ohne Änderung. Die Zeile von `make test-integration` in `harness/README.md` („neun Stufen“) bleibt beim Planner (nächste Zeile).
- **`harness/README.md` nennt „neun Stufen“** in der Zeile von `make test-integration`; mit Stufe 10 stimmt sie nicht mehr. Die Datei ist Gate-Kette (Übergangsregeln Nr. 7); der Slice-Zweig schreibt sie nicht — **Ausgang:** offen; Adresse: Planner (Integrationsschritt, mit `docs/plan/planning/e2e-abdeckung.md` aus `make e2e-abdeckung`).
- **Sendung `slice-hybrid-service-upstream` Z35** (Host-Fänger am echten Adapter; Planner, 2026-11-18, `AGENTS.md` §3.13): §1 *Ausdrücklich NICHT* dieses Slice trägt sie nicht als Ausschluss, §1 *Übernimmt* und DoD Punkt 2 nennen sie mit der Kennung des Gebers; Fänger M13 — **Ausgang:** eingetreten (angenommen).
- **Zeile M22 und Marke** (Architect, 2026-11-18, `AGENTS.md` §3.13): die Warnung und die Werte des Upstream-Clients im Modus `hybrid` (`SPEC-139` Nr. 9, 12 a) brauchen einen Fänger, der in der Schätzung des Planners (15 Zeilen) fehlt — **Ausgang:** offen; Adresse: Planner (Marke beim Start nach `SPEC-128` Nr. 1: *b* = 14, *S* = 1820, *M* = 2500).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts` (1×) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (2×, Prozess, Referenzdienst und Client unter Last) sowie `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, `AGENTS.md` §3.9) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
