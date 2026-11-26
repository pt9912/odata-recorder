# Slice slice-bootstrap-m6-transport: Transportschutz des Admin-Listeners im Composition Root: `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT`, die vier Aufrufe von `SPEC-299`, Warnung, Umhüllung und Handbuch zum TLS (Teil C des Schnitts von `slice-bootstrap-m6`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-RB-02`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen** (Teil C des Schnitts von [slice-bootstrap-m6](slice-bootstrap-m6.md)): `LH-FA-13.a` · `SPEC-300` Nr. 2 (die drei Schlüssel), 3, 10 und 15 zum TLS (Admin-Listener im Composition Root, Architect 2026-11-23) · `SPEC-299` Nr. 3 bis 9 und 14 (TLS des Admin-Listeners; Nr. 8, 9 und 14 sind die Aufrufe, die dieser Teil setzt) · `SPEC-110` Nr. 2, 10 und 11 (Schlüssel dieser Version, Zusammenhang, Dateien) · `SPEC-089`, `SPEC-090` · `SPEC-291` Nr. 8 (die Vereinigung der Verstöße beider Paare) · `SPEC-139` Nr. 6 und 8 (Hochfahren, Start) · `SPEC-200`, `SPEC-201` Nr. 4 (Handbuch und Katalog) · `ARC-007`, `ARC-012` · `architecture.md` §2.5, §4.5.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912 (Planner; Schnitt vor dem Code aus [slice-bootstrap-m6](slice-bootstrap-m6.md), Vorschlag des Architect vom 2026-11-23). **Datum:** 2026-11-23.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Composition Root liest `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` und ruft die Funktionen von `slice-admin-tls` (`pruefeAdminTransport`, `adminLoopback`, `umhuelle`, `warne` in `src/bootstrap/admintls.go`) an den vier Stellen von `SPEC-299` Nr. 8 in der Reihenfolge von Nr. 8 a: beim Laden der Konfiguration, in `Build`, in `Hochfahren` und in `Start`. Ein Verstoß verhindert den Start vor dem Anlegen von `DATA_DIR`, vor der Token-Datei und vor jedem Listener, ein Verstoß erzeugt nie ein Token; `Build` vereinigt die Verstöße von `ADMIN_TLS_*` mit denen von `LISTEN_TLS_*`, nach Schlüssel sortiert; der Weg Klartext schreibt am Start einmal die Warnung `start: Admin-Listener ohne TLS erreichbar` in jedem Modus (`SPEC-299` Nr. 9); bei `wegTLS` umhüllt `umhuelle` den Admin-Listener. Der Zwischenstand von Teil B (Zeile B19: die Prüfung mit leerem Paar und falschem Flag) endet; Handbuch und Katalog tragen die drei Schlüssel und das Kapitel `TLS` zum Admin-Listener (`SPEC-300` Nr. 2, 3, 10, 15). Danach startet `slice-admin-tls-betrieb`, sobald auch Teil D in `done/` liegt. **Schnitt vor dem Code** (Vorschlag des Architect, Planner, 2026-11-23): die Schätzung des ungeschnittenen Plans (*S* = 4366) liegt über der Schnittgrenze von `SPEC-128` Nr. 1; Teil A [slice-bootstrap-m6-endpunkte](../in-progress/slice-bootstrap-m6-endpunkte.md) und Teil B [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md) liegen vor diesem Teil, Teil C (diese Datei, C01 bis C08), Teil D [slice-bootstrap-m6](slice-bootstrap-m6.md) (M6-Test, D01 bis D06) folgt.

**Übernimmt:** das Lesen und Prüfen von `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` samt Aufheben ihrer Ablehnung in `SPEC-110` Nr. 2 (Sendung von `slice-admin-listener-token` und `slice-admin-tls`, deren §1 *Ausdrücklich NICHT* den Root nennen; `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE` liegen bei Teil B) (`AGENTS.md` §3.13).

**Übernimmt außerdem** (Sendung von `slice-admin-tls`, Eintrag des Planners vom 2026-11-23 beim Start des Gebers, `AGENTS.md` §3.13; beim Schnitt vom 2026-11-23 an diesen Teil gelegt; §1 *Ausdrücklich NICHT* und DoD dieses Slice vorher gelesen; der Geber liefert die Funktionen `pruefeAdminTransport`, `adminLoopback`, `umhuelle` und `warne` in `src/bootstrap/admintls.go` und ruft sie nicht, `SPEC-299` Nr. 14): (a) die vier Aufrufe von `SPEC-299` Nr. 8 aus dem Laden der Konfiguration, aus `Build`, aus `Hochfahren` und aus `Start`, mit der Reihenfolge von Nr. 8 a (ein Verstoß verhindert den Start vor dem Anlegen von `DATA_DIR`, vor der Token-Datei und vor jedem Listener, ein Verstoß erzeugt nie ein Token; Zeilen C03 bis C06); (b) in `Build` die Vereinigung der Verstöße von `ADMIN_TLS_*` mit denen von `LISTEN_TLS_*`, nach Schlüssel sortiert (`SPEC-299` Nr. 8 b, `SPEC-291` Nr. 8; nur die Vereinigung der beiden Listen, nicht das Lesen oder Binden von `LISTEN_TLS_*`, das `slice-listener-tls-verdrahtung` in `done/` geliefert hat; Zeile C04); (c) die Warnung am Start, einmal je Start in jedem Modus (`SPEC-299` Nr. 9; Zeile C06); (d) die Zelle Modus × Weg (AT22 von `slice-admin-tls`, akzeptiertes Negativ dort): die Funktionen kennen keinen Modus, der Fänger je Modus (`record`, `replay`, `mock`, `hybrid`) liegt im Test des Starts hier (Zeile C07); (e) das Lesen von `ADMIN_ALLOW_PLAINTEXT` mit der Menge der elf Werte aus AT23 (`true`, `false`, `TRUE`, `True`, `1`, `yes`, `on`, leer, nur Leerraum, `true` mit Leerraum am Rand und *ungesetzt*, der Schlüssel fehlt in der Umgebung; Orakel in `slice-admin-tls` §3; Zeile C02); (f) der Nachzug von `SPEC-110` Nr. 10 und 11 mit Zeigern auf `SPEC-299` (Text des Architect, geliefert) und das Handbuch samt Katalog, die der Geber ausdrücklich nicht ändert (`SPEC-299` Nr. 14, `slice-admin-tls` §3 *Handbuch und Katalog*): die Tabelle `Einstellungen` um `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT`, im Kapitel `TLS` ein Abschnitt zum Admin-Listener (Voraussetzung, Vorgehen, Ergebnis, die Tabelle der Gründe unter den Namen `ADMIN_TLS_*` samt dem Grund `außerhalb von Loopback verlangt TLS: ADMIN_TLS_CERT und ADMIN_TLS_KEY` am Schlüssel `ADMIN_LISTEN`, die Warnung des Weges Klartext mit ihrem festen Text, die Verantwortung des Betreibers für den Klartext und die Aussage, dass ein abgelehnter Handshake beim Betreiber keine Zeile hinterlässt) und ein Beispiel als Datei gegen das gebaute Binary (`ADMIN_LISTEN=0.0.0.0:8081` ohne Paar und ohne Flag: Exit 1 und die Zeile `ODR-E2001 ADMIN_LISTEN: …`; Zeile C08); `ODR-E2001` steht schon im Handbuch, ein neuer Code und eine neue Ausnahme entstehen nicht. Der Geber ist [slice-admin-tls](../done/slice-admin-tls.md).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Dienste, Routen und Einstellungen der Diagnose im Root und der Ersatz der vier Wächter — [slice-bootstrap-m6-endpunkte](../in-progress/slice-bootstrap-m6-endpunkte.md) (Teil A, liegt bei diesem Start in `done/`).
- `ADMIN_LISTEN` und `ADMIN_TOKEN_FILE`, die Token-Datei, die Bindung und das Herunterfahren beider Listener, `GET /healthz`, die Zeilen `admin_abgewiesen`, Handbuch und Katalog zu den zwei Schlüsseln und den zwölf Codes — [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md) (Teil B, liegt bei diesem Start in `done/`); dieser Teil ändert sie nicht und ersetzt allein den Zwischenstand von B19.
- Der M6-Test (Rundlauf, Abweisung ohne Zustandsänderung, die Zellen Modus × Art, Stufe 11 `TestE2EAdmin`) — [slice-bootstrap-m6](slice-bootstrap-m6.md) (Teil D, Zeilen D01 bis D06).
- Der Handshake des Admin-Adapters über TLS (Zeilen AT11 bis AT17 von `slice-admin-tls`) — [slice-admin-tls-betrieb](slice-admin-tls-betrieb.md), startet nach allen vier Teilen. Dieser Teil belegt allein, dass der Root umhüllt (C05); die Zusagen des Handshakes trägt der Folge-Slice.
- `LISTEN_TLS_*` (Lesen, Binden und Umhüllen des OData-Listeners) — geliefert von `slice-listener-tls-verdrahtung` (`done/`); dieser Teil vereinigt allein die Verstöße beider TLS-Paare in `Build`. `HYBRID_*` und `FALLBACK_*` — geliefert von `slice-bootstrap-m4-konfig` und `slice-bootstrap-m4` (`done/`).
- Die Begrenzung der Fehlversuche (`LH-FA-13.a` nannte keine) — entschieden (User, 2026-11-24, Variante A, `SPEC-301`); Träger ist der Folge-Slice [slice-admin-abweisen-begrenzen](slice-admin-abweisen-begrenzen.md), Vorlage in §6 von [slice-bootstrap-m6](slice-bootstrap-m6.md). Neue Endpunkte oder Fachlogik und das Umschalten des Modus zur Laufzeit — nicht Gegenstand des Schnitts.

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

- [ ] Die drei Schlüssel und die vier Aufrufe (Geber `slice-admin-tls`, §1 *Übernimmt außerdem*): `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` gelesen (die Werte kommen aus dem Wert des Starts bis in die Prüfung, der Nullwert steht nie anstelle eines gesetzten Werts; das Flag mit den elf Werten von AT23); die vier Aufrufe von `SPEC-299` Nr. 8 (Laden, `Build` mit der Vereinigung der Verstöße beider TLS-Paare, `Hochfahren` mit der Umhüllung, `Start` mit der Warnung) in der Reihenfolge von Nr. 8 a; der Weg gilt in allen vier Modi unverändert (Zelle Modus × Weg); der Zwischenstand von Zeile B19 aus Teil B ist aufgehoben (Zeilen C01 bis C07 der Tabelle in §3, vom Architect).
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary, `SPEC-200`): die Tabelle `Einstellungen` um `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT`, im Kapitel `TLS` ein Abschnitt zum Admin-Listener mit Gründen, Warnung und Verantwortung für den Klartext, das Beispiel `ADMIN_LISTEN=0.0.0.0:8081` ohne Paar mit `ODR-E2001 ADMIN_LISTEN: …` (Zeile C08; `SPEC-300` Nr. 15); das Kataloge-Gate ist grün, ein neuer Fehlercode und eine neue Ausnahme entstehen nicht.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-299` Nr. 14, `SPEC-300`, `SPEC-110` Nr. 10 und 11 stehen vor dem Start auf `main`; dieser Teil schreibt keine Spezifikation).
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
| `src/bootstrap/` (`konfig_admin.go`, `lebenszyklus.go`, `start.go`, `wurzel.go`) | update | die drei Schlüssel lesen; die vier Aufrufe von `SPEC-299` Nr. 8 (Laden, `Build`, `Hochfahren`, `Start`) mit den Funktionen aus `admintls.go`; die Vereinigung der Verstöße; die Warnung; die Umhüllung des Admin-Listeners |
| `src/bootstrap/` (Tests: `export_test.go`, neu `admin_tls_lesen_test.go`, `admin_tls_stellen_test.go`, `admin_tls_weg_test.go`) | update und neu | Naht (Eingabe des Transports, Stellen des Aufrufs, Weg je Modus); die Fänger der Zeilen unten |
| Handbuch (`docs/user/benutzerhandbuch.md`) | update | die drei Schlüssel in der Tabelle `Einstellungen`, der Abschnitt zum Admin-Listener im Kapitel `TLS` und das Beispiel gegen das gebaute Binary (`SPEC-300` Nr. 15); der Katalog bekommt keinen neuen Code und keine Ausnahme |
| `spec/spezifikation.md` | geliefert (Architect, 2026-11-23) | `SPEC-300` neu; Nachzug von `SPEC-110` Nr. 2, 10 und 11, `SPEC-299` Nr. 14; dieser Teil schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Nachzug in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Sie stehen unten (Architect, 2026-11-23, Zeilen C01 bis C08 der Tabelle des ungeschnittenen Plans, vom Planner beim Schnitt im Wortlaut übernommen; jede Zeile steht in genau einem der vier Teile).

**Zusage · Fänger** (`AGENTS.md` §3.15; Architect, 2026-11-23; `Nr.` ist die Kennung des ungeschnittenen Plans, `A` Endpunkte im Root ([slice-bootstrap-m6-endpunkte](../in-progress/slice-bootstrap-m6-endpunkte.md)), `B` Admin-Listener ([slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md)), `C` Transportschutz (diese Datei), `D` M6-Test ([slice-bootstrap-m6](slice-bootstrap-m6.md)); die Spalte *Zusage* nennt den Ort in `SPEC-300`, `SPEC-299` und den Verträgen; die Tests laufen im Paket `bootstrap_test` ohne Netz außer dem Loopback des Prozesses, Pfad und Namen der Naht in `export_test.go` bestimmt der Implementer, die Mengen stehen mit Achsen unten; der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| C01 | (`SPEC-299` Nr. 4, 5; `SPEC-300` Nr. 2, 3) Der Root übergibt `pruefeAdminTransport` die Werte der Konfiguration unverändert: `Listen`, `CertPfad`, `KeyPfad` (Leerraum am Rand entfernt, leer gleich ungesetzt), `Klartext` und `FormGueltig`, den Nullwert nie anstelle eines gesetzten Werts | `TestAdminEingabe`: Achsen Feld {`Listen`, `CertPfad`, `KeyPfad`, `Klartext`} × Zustand {gesetzt, nur Leerraum, leer, ungesetzt} × Modus 4 = 64 Zellen; Orakel das Feld der Eingabe, abgelesen an der Naht der Stelle `Laden` und der Stelle `Build` (je 64 Zellen mit demselben Literal) | je Feld der Nullwert übergeben (4 Mutationen, je rot in 4 × 2 Zellen mit gesetztem Wert); `Klartext` aus `TRUE` gelesen (rot); das Flag nur im Modus `record` gelesen (rot: 3 Modi) | andere |
| C02 | (`SPEC-299` Nr. 5, AT23) `ADMIN_ALLOW_PLAINTEXT` nimmt genau `true` und `false`; ungesetzt, leer und nur Leerraum sind `false`; `true` mit Leerraum am Rand ist wahr; jeder andere Wert ist ein Verstoß am Schlüssel und gilt für den Zusammenhang weder als wahr noch als falsch | `TestAdminFlag`: Menge M-Flag der 11 Werte (`true`, `true` mit Leerraum, `false`, ungesetzt, leer, nur Leerraum, `TRUE`, `True`, `1`, `yes`, `on`) × 4 Modi = 44 Zellen, mit `ADMIN_LISTEN=0.0.0.0:8081` ohne Paar; Orakel je Zelle als Literal (wahr: Weg Klartext, kein Verstoß; falsch: die Zeile des Zusammenhangs; ungültig: allein die Zeile am Flag, keine Zeile an `ADMIN_LISTEN`) | `strconv.ParseBool` (rot: `1`, `TRUE`, `True`, 12 Zellen); Schreibweise gleichgültig (rot: 8); `yes` wahr (rot: 4); ein ungültiger Wert gilt als falsch (rot: 20, die Zeile am Zusammenhang erscheint zusätzlich); ungesetzt gilt als ungültig (rot: 4, die Zeile am Flag statt der am Zusammenhang); ungesetzt gilt als wahr (rot: 4, Weg Klartext ohne Verstoß statt der Zeile des Zusammenhangs); leer gilt als ungültig (rot: 4); Leerraum am Rand nicht entfernt (rot: 8, `true` mit Leerraum und nur Leerraum werden ungültig) | andere |
| C03 | (`SPEC-299` Nr. 8 a, Stelle *Laden*) Das Laden ruft die Prüfung nach den reinen Prüfungen und zusammen mit den Dateien von `LISTEN_TLS_*`, mit der Zeit des Starts: ein Verstoß beendet den Start, bevor `DATA_DIR` angelegt, die Token-Datei geladen oder ein Listener geöffnet wird | `TestAdminLaden`: 3 Verstöße (Zusammenhang außerhalb von Loopback, Paar mit fehlender Datei, Flag ungültig) × 4 Modi = 12 Zellen; Orakel Exit 1, Zeilen als Literal nach Schlüssel sortiert, `DATA_DIR` nicht angelegt, keine Token-Datei, kein Bind; mit einem Verstoß von `LISTEN_TLS_*` zugleich: eine Meldung | die Prüfung nach der Anlage von `DATA_DIR` (rot: 12); nach der Token-Datei (rot: die Datei entsteht, 12); der Aufruf entfällt (rot: 12); nur im Modus `record` (rot: 9) | andere |
| C04 | (`SPEC-299` Nr. 8 b, Stelle *Build*) `Build` lädt das Paar erneut mit der einen Uhr vor dem Öffnen der Ablage; die Verstöße beider Paare stehen in einer Meldung, nach Schlüssel sortiert; der `adminTransport` bleibt im `Handle` | `TestAdminBuild`: Literal der Konfiguration (umgeht das Laden), Menge {nur Admin-Paar, nur `LISTEN_TLS_*`, beide} × 4 Modi = 12 Zellen; Orakel die Folge der Zeilen als Literal, weder Datenbank noch Blob-Verzeichnis entstanden; ohne Verstoß liefert die Naht den `adminTransport` des `Handle` | der Aufruf entfällt (rot: 8 Zellen mit Admin-Paar); ohne Sortierung (rot: 4); die Ablage vor der Prüfung geöffnet (rot: 12) | andere |
| C05 | (`SPEC-299` Nr. 8 c, 3, 7, Stelle *Hochfahren*) `umhuelle` umhüllt den Admin-Listener bei `wegTLS`, sonst bleibt der Listener der gebundene; über TLS antwortet der Rahmen wie im Klartext | `TestAdminUmhuellung`: 3 Wege (`wegTLS` mit gültigem Paar, `wegLoopback`, `wegKlartext`) × 4 Modi = 12 Zellen; Orakel: `wegTLS`: ein Client mit der CA spricht TLS 1.3 und `GET /healthz` ist 200, ein Klartext-Client scheitert; sonst gelingt der Klartext-Request, kein TLS-Client | `umhuelle` nie gerufen (rot: 4 Zellen `wegTLS`); immer umhüllt (rot: 8); die Umhüllung nur im Modus `record` (rot: 3) | andere |
| C06 | (`SPEC-299` Nr. 8 d, 9, Stelle *Start*) `warne` schreibt nach einem gelungenen `Build` und vor `Hochfahren` genau eine WARN-Zeile `start: Admin-Listener ohne TLS erreichbar` mit `ereignis=admin_klartext`, nur im Weg Klartext, in jedem Modus, einmal je Start | `TestAdminWarnung`: 3 Wege × 4 Modi = 12 Zellen und ein Fehler von `Hochfahren` (Token-Datei ein Verzeichnis) im Weg Klartext (4 Zellen: die Warnung steht, Exit 1); Orakel Zahl und Felder der Zeilen auf `stderr`; zwei Starts nacheinander: zwei Zeilen | der Aufruf entfällt (rot: 4); Warnung auch bei `wegLoopback` (rot: 4); vor `Build` (rot: eine Warnung bei einem Verstoß von `Build`, 4); zwei Zeilen (rot) | andere |
| C07 | (`SPEC-299` Nr. 3, 6) Der Weg gilt in jedem Modus unverändert: dieselbe Konfiguration ergibt in den vier Modi denselben Weg, dieselben Verstöße und dieselbe Warnung (AT22) | `TestAdminWegJeModus`: Menge M-Weg der 32 Zellen von AT02 (8 Hosts × Paar × Flag) × 4 Modi = 128 Zellen an der Naht der Stelle `Build`; Orakel: die Zelle des Modus `record` ist die der anderen drei | der Weg hängt am Modus (rot: der Vergleich der Modi, je Mutation 32 Zellen) | andere |
| C08 | (`SPEC-300` Nr. 15, `SPEC-299` Nr. 14) Das Handbuch trägt `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT`, den Abschnitt zum Admin-Listener im Kapitel `TLS` und das Beispiel `ADMIN_LISTEN=0.0.0.0:8081` ohne Paar mit Exit 1 und der Zeile `ODR-E2001 ADMIN_LISTEN: außerhalb von Loopback verlangt TLS: ADMIN_TLS_CERT und ADMIN_TLS_KEY` | `make handbuch-check` (das Beispiel als Datei gegen das gebaute Binary) und `make katalog-check` grün | das Beispiel ohne Zeile `ODR-E2001` (rot: Beispiel stimmt nicht mit dem Binary); ein Schlüssel ohne Zeile in `Einstellungen` (rot: Art 4) | andere |

Die Zeilen C01 bis C08 sind 8: *a* = 0, *b* = 8. **Mengen mit Achsen** (Orakel als Literal, nie aus dem Code und nie als Aufruf des Codes; die Zahl der Zellen steht im Test als Literal und wird gegen die gezählten verglichen, `AGENTS.md` §3.15): M-Modus (`record`, `replay`, `mock`, `hybrid`, 4); M-Flag (11 Werte, Achsen gültig wahr, gültig falsch, ungültig); M-Weg (8 Hosts × Paar × Flag, wie AT02). Die Mengen M-Route, M-Einstellung (Teil A), M-Bind, M-Abweisung (Teil B) und M-Art (Teil D) stehen in den Plänen der Teile, die sie führen.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18). Das gilt auch für die Beispiele des Handbuch-Kapitels.
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md) (Teil B des Schnitts) liegt in `done/`, und mit ihm Teil A ([slice-bootstrap-m6-endpunkte](../in-progress/slice-bootstrap-m6-endpunkte.md)) und die Bedingungen von Teil A; [slice-admin-tls](../done/slice-admin-tls.md) liegt in `done/` (er liefert `src/bootstrap/admintls.go` mit den vier Funktionen). Die Spezifikation trägt den Nachzug von `SPEC-110` Nr. 10 und 11, den Vertrag `SPEC-299` und `SPEC-300` samt Tabelle und Raster, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12; geliefert am 2026-11-23). Teil D ([slice-bootstrap-m6](slice-bootstrap-m6.md)) startet erst, wenn dieser Teil in `done/` liegt.

**Warum:** der Teil ruft die Funktionen von `slice-admin-tls` (*Code-Abhängigkeit*) an den Stellen, die Teil B liefert (Laden, `Build`, `Hochfahren`, `Start`; *Code-Abhängigkeit*), und hebt den Zwischenstand B19 auf. Teil B, Teil C und Teil A ändern dieselben Dateien in `src/bootstrap/` (`konfig_admin.go`, `lebenszyklus.go`, `start.go`, `wurzel.go`, `export_test.go`) und laufen nacheinander.

**Größenmarke** (`SPEC-128` Nr. 1 und 10 g, `AGENTS.md` §3.14; Teil des Architect 2026-11-23 im Block unten; der Planner hat die Zahlen am 2026-11-23 aus der Tabelle nachgerechnet, **`basis`, `schaetzung` und `marke` setzt der Planner beim Start vor dem ersten Code-Commit**, mit `make groessenmarke`): 8 Zeilen der Tabelle (C01 bis C08), *a* = 0, *b* = 8, *f* = 0 (`tlstest` und der TLS-Client des Tests von `LISTEN_TLS_*` bestehen), *p* = 90 (drei Schlüssel und die vier Aufrufe): *S* = 85·0 + 80·8 + 200·0 + 90 = 730, 135·730 = 98550, **Marke 1000** = 100 · ⌈9,855⌉; keine Schnittfrage (135·*S* ≤ 300000, *S* < 2223), kein `schnitt-grund`; der Teil liegt 1493 unter der Schnittgrenze, jede weitere Zeile kostet 80. Keine Abweichung vom Vorschlag des Architect. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen.

```groessenmarke
pfad: src/bootstrap/konfig_admin.go
pfad: src/bootstrap/lebenszyklus.go
pfad: src/bootstrap/start.go
pfad: src/bootstrap/wurzel.go
pfad: src/bootstrap/export_test.go
pfad: src/bootstrap/admin_tls_lesen_test.go
pfad: src/bootstrap/admin_tls_stellen_test.go
pfad: src/bootstrap/admin_tls_weg_test.go
produktcode: 90
fakes: 0
```

Der Block nennt die Dateien einzeln (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`); die Dateien des Handbuchs und der Spezifikation stehen nicht im Block (`SPEC-128` Nr. 3 m). Die Dateien `admin_tls_lesen_test.go`, `admin_tls_stellen_test.go` und `admin_tls_weg_test.go` gibt es noch nicht (die Namen sind ein Vorschlag des Architect; die Messung zählt dort 0, bis sie entstehen); ein anderer Name ersetzt den Vorschlag im selben Commit, und der Implementer nennt jede weitere Datei, die er ändert oder anlegt, im selben Commit im Block.

**Kollisionen** (Architect, 2026-11-23, gelesen an den Zweigen und den Plänen, vom Planner beim Schnitt übernommen; `SPEC-128` Nr. 10 g): (1) Teil A und Teil B ändern `wurzel.go`, `lebenszyklus.go`, `start.go`, `konfig_admin.go` und `export_test.go`: beide liegen vor diesem Teil in `done/` (Start-Bedingung), ihre Zeilen gehören zur Basis. (2) Teil D ändert keine Datei dieses Blocks (`audit_betrieb_test.go`, `tests/`) und läuft nach diesem Teil. (3) `slice-admin-tls` liegt in `done/`; `admintls.go` und `admintls_test.go` stehen nicht im Block. (4) Die Gate-Kette (Nr. 7 der Übergangsregeln): kein Pfad aus `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`; der Planner prüft es beim Start.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md)); ein weiterer Schnitt ist nicht vorgesehen.
- `in-progress` → `open` (blockiert — Carveout?): die Funktionen von `slice-admin-tls` lassen sich an einer der vier Stellen nicht in der Reihenfolge von `SPEC-299` Nr. 8 a rufen, ohne dass ein Verstoß ein Token erzeugt oder `DATA_DIR` anlegt (Befund an den Architect).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23; die Zeile dieses Teils; die Zeilen der Teile B und D stehen in deren Plänen; jede Zelle nennt ihren Ort in `SPEC-300`, `SPEC-299`, den Verträgen davor oder „gilt nicht“ mit Grund; der Implementer hält an einer Zelle ohne Ort an, `AGENTS.md` §3.12):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Lesen von `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` | kein TLS, Flag `false` (`SPEC-299` Nr. 4, 5) | wie fehlend (`SPEC-110` Nr. 1, `SPEC-299` Nr. 4, 5) | genau ein Pfad: Rang 1 am leeren Schlüssel; ungültiges Flag (`TRUE`, `1`, `yes`, `on`): Verstoß am Schlüssel, weder wahr noch falsch (`SPEC-299` Nr. 4, 5, 6) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | dasselbe Ergebnis, die Gültigkeit des Zertifikats hängt an der Zeit des Starts (`SPEC-110` Nr. 13, `SPEC-299` Nr. 13) |

**Zusatztabelle** (`AGENTS.md` §3.16: die Operationen melden mehrere Verstöße, und Zusagen hängen vom Modus ab; je Paar von Verstößen am selben Schlüssel und je Modus eine Zeile; die Paare des Listeners stehen in Teil B):

| Paar oder Modus | Ergebnis | Fänger |
|---|---|---|
| Paar von `ADMIN_TLS_*` und Paar von `LISTEN_TLS_*` zugleich mit Verstoß (auch dieselbe Datei) | alle Zeilen in einer Meldung, nach Schlüssel sortiert (`ADMIN_TLS_CERT` vor `LISTEN_TLS_CERT`) | C04, C03 |
| `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` beide mit Verstoß | zwei Zeilen nach Schlüssel (`SPEC-299` Nr. 4) | `slice-admin-tls` (AT04), C03 |
| Zusammenhang von `ADMIN_LISTEN` und ungültiges Flag zugleich | allein die Zeile am Flag, keine an `ADMIN_LISTEN` (`SPEC-299` Nr. 6) | C02 |
| jeder Modus (`record`, `replay`, `mock`, `hybrid`) | derselbe Weg, dieselben Verstöße, dieselbe Warnung (`SPEC-299` Nr. 3, `SPEC-300` Nr. 3) | C07, alle Mengen mit der Achse Modus |

- Die vier Aufrufe von `SPEC-299` Nr. 8 sind eine Menge von Aufrufstellen (Laden, `Build`, `Hochfahren`, `Start`) mal vier Modi: der Fänger führt je Stelle und je Modus einen Fall, einen am ersten Aufruf der Kette fängt die drei anderen nicht (`AGENTS.md` §3.15; Geber `slice-admin-tls`, AT22 und AT23) — **Ausgang:** entschieden (Architect, 2026-11-23): je Stelle eine Zeile (C03 Laden, C04 `Build`, C05 `Hochfahren`, C06 `Start`), je Zeile die vier Modi, dazu C07 für den Weg je Modus und C01 und C02 für die Werte.
- Zwischenstand von Teil B (B19: die Prüfung mit leerem Paar und falschem Flag, die drei Schlüssel „nicht unterstützt“) — **Ausgang:** teils entschieden (Architect, 2026-11-23): dieser Teil hebt ihn auf, die Zeilen C01 bis C07 führen den Weg mit gelesenem Paar und Flag; **offen, Adresse: Architect** (Planner, 2026-11-23, `AGENTS.md` §3.13): ob der Test von B19 (`TestAdminLoopbackZwischenstand`, 5 Hosts mit leerem Paar und Flag `false`) neben C03 und C04 stehen bleibt oder in sie eingeht, denn die Hosts von B19 sind eine Teilmenge von M-Weg; kein Fänger darf entfallen (`AGENTS.md` §3.6), die Antwort steht vor dem Start dieses Teils im Plan.
- Handbuch zum TLS (Auftrag aus dem ungeschnittenen Plan, Architect, 2026-11-23): die Texte schreibt der Teil, die Stellen sind `SPEC-200` (Beispiele als Datei gegen das Binary), `SPEC-201` Nr. 4 (Schlüssel), in `docs/user/benutzerhandbuch.md` die Tabelle `Einstellungen` und der Abschnitt im Kapitel `TLS` — **Ausgang:** entschieden; Träger: Zeile C08.
- Gemeinsame Dateien in `src/bootstrap/` mit den Teilen A und B (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`) — **Ausgang:** entschieden (Planner, 2026-11-23): Reihenfolge A, B, dieser Teil; keine parallelen Arbeitsbäume auf `wurzel.go`; die Zeilen von A und B gehören zur Basis.
- Größe: die Tabelle *Zusage · Fänger* liegt vor, die Schätzung des ungeschnittenen Plans (4366) liegt über der Schnittgrenze — **Ausgang:** entschieden (Planner, 2026-11-23, nach dem Vorschlag des Architect): Schnitt in vier Teile mit den Marken 1100, 3000, 1000 (dieser Teil) und 1000, nachgerechnet aus der Tabelle (keine Abweichung); `basis`, `schaetzung` und `marke` setzt der Planner beim Start.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag des Schnitts gesichtet (2026-11-23; der Planner zieht Treffer und Zähler beim Start nach). Offen und für den Teil einschlägig: `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×, `src/bootstrap/`, §4) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (22×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (77×, `AGENTS.md` §3.15; vier Aufrufstellen mal vier Modi), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (43×, `AGENTS.md` §3.9; der Nachzug von `SPEC-110` und das Handbuch sind Stellen), `BEO-REPO/randform-erst-im-review-entschieden` (37×, `AGENTS.md` §3.16) und `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (`AGENTS.md` §3.18, für die Beispiele des Handbuchs).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
