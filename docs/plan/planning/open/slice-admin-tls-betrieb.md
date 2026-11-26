# Slice slice-admin-tls-betrieb: Handshake des Admin-Adapters über TLS

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

**Berührte Spec-Stellen:** `LH-FA-13.a` (Transportschutz) · `LH-RB-02.a` (Listener-TLS) · `LH-QA-04.a` · `SPEC-299` Nr. 7, 10 und 11 (der Adapter über TLS: Parameter, Handshake, Neuladen) · `SPEC-291` Nr. 5 und 6 (TLS 1.2 und 1.3, nur HTTP/1.1, kein mTLS) · `SPEC-238` Nr. 1 und 11 (der Adapter kennt TLS nicht, `Anhalten`) · `SPEC-136` Nr. 13 (Listener des Adapters) · `ARC-007` · `architecture.md` §Sicherheit.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-23.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Datei `src/bootstrap/admintls_betrieb_test.go` (Paket `bootstrap_test`) belegt, dass der Admin-Adapter über den von `umhuelle` umhüllten Listener TLS spricht (`SPEC-299` Nr. 7, 10, 11): ein Test-Server führt `admin.New` über `umhuelle` und `net.Listen` mit zählenden `IDs`, `Audit` und `Telemetrie` aus `supporttest`; die Tests prüfen die Verbindung (TLS 1.2 und 1.3, nur HTTP/1.1, kein mTLS, Kette der Datei), den Rahmen über TLS (dieselbe Antwort wie im Klartext), den gescheiterten Handshake (erreicht den Adapter nie), den hängenden Handshake (hält keinen anderen Client auf), `Anhalten` und `Schliessen` über TLS und das fehlende Neuladen der Dateien (Zeilen AT11 bis AT17). Der Slice liefert keinen Produktcode und ändert das Binary nicht.

**Übernimmt:** den Teil *Betrieb* (Zeilen AT11 bis AT17) aus [slice-admin-tls](../done/slice-admin-tls.md) (Geber; Entscheidung des Architect, 2026-11-23, §6 dort): die Datei `admintls_betrieb_test.go` liegt fertig, gemessen und grün auf dem Zweig `slice-admin-tls-betrieb` (Kopf `3726eda0`, 627 Zeilen) und wird von dort unverändert übernommen, bis auf die Methode `ladeAdmin`, die seither in `admintls_test.go` des Gebers steht (§3).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Funktionen `adminLoopback`, `pruefeAdminTransport`, `umhuelle` und `warne`, ihre Tests (Zeilen AT01 bis AT10 und AT18 bis AT23) und die Naht in `export_test.go` — [slice-admin-tls](../done/slice-admin-tls.md); dieser Slice ruft sie und ändert sie nicht.
- Eine Ausnahme der Datei in den vier Wächtern `TestIM16BinaryUnveraendert`, `TestI18BinaryUnveraendert`, `TestS17BinaryUnveraendert` und `TestWS15BinaryUnveraendert` (Paket `admin`) — eine Ausnahme lockerte vier Fänger fertiger Slices (`AGENTS.md` §3.6) und braucht den User; der Slice braucht sie nicht, weil Teil A des Schnitts von `slice-bootstrap-m6` ([slice-bootstrap-m6-endpunkte](slice-bootstrap-m6-endpunkte.md)) die Wächter ohnehin durch Fänger der Verdrahtung ersetzt (dort §1, DoD und §6).
- Die Verdrahtung der Funktionen im Binary, das Lesen der `ADMIN_*`-Schlüssel, die Warnung am Start und das Handbuch — die Teile des Schnitts von `slice-bootstrap-m6` (Planner, 2026-11-23): [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md) (`ADMIN_LISTEN`, `ADMIN_TOKEN_FILE`, Listener) und [slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md) (die drei Schlüssel des Transportschutzes, die vier Aufrufe, die Warnung, das Handbuch zum TLS).
- Der Netzwerkzugriff auf jeden Admin-Endpunkt ohne und mit falschem Token und der Canary-Test — [slice-test-security](../done/slice-test-security.md); hier prüft der Test-Server den Rahmen mit einer Route und einem Token.
- Änderungen am Adapter `admin` — er kennt TLS nicht (`SPEC-299` Nr. 1, `SPEC-238` Nr. 1); der Diff nennt keinen Pfad unter `src/adapters/driving/admin/`.

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

- [ ] Verbindung und Rahmen über TLS (`SPEC-299` Nr. 7, 10 b und d): TLS 1.2 und 1.3 gelingen, 1.1 und allein `h2` scheitern, kein Client-Zertifikat verlangt, die Kette der Datei wird gezeigt; die Antwort, die `Anfrage` an die Route und die Einträge über TLS gleichen denen im Klartext; ein gescheiterter Handshake erreicht den Adapter nie und schreibt in keine Senke (Zeilen AT11, AT12, AT13 der Tabelle in §3; dazu die Mutationen `m10d` und `m10e` am Handshake von `umhuelle`).
- [ ] Hängen, `Anhalten` und `Schliessen` über TLS (`SPEC-299` Nr. 10 a und c): ein stummer Client hält weder `Accept` noch einen anderen Client auf; nach `Anhalten` antwortet ein Request über TLS mit 503 `ODR-E5010`, auch wenn sein Handshake erst danach endet; `Schliessen` schließt eine Verbindung mit offenem Handshake sofort (Zeilen AT14, AT15, AT16; jede in jedem Lauf gefangen).
- [ ] Kein Neuladen (`SPEC-299` Nr. 11): ein Wechsel der Dateien nach dem Laden ändert das gezeigte Zertifikat nicht, ein zweiter Aufruf von `pruefeAdminTransport` zeigt das neue, ein im Betrieb ablaufendes Zertifikat wird weiter bedient (Zeile AT17).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: nicht berührt; das Binary bleibt unverändert (§7 hält es fest; die Verdrahtung und das Handbuch-Kapitel liegen bei den Teilen [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md) und [slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md) des Schnitts von `slice-bootstrap-m6`).
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
| `src/bootstrap/admintls_betrieb_test.go` | neu (Paket `bootstrap_test`, Lint `testpackage`) | AT11 bis AT17: der Test-Server (`admin.New` über `umhuelle` und `net.Listen`, zählende `IDs`, `Audit` und `Telemetrie` aus `supporttest`) und die Handshakes mit bekannter und unbekannter CA; die Datei liegt fertig auf dem Zweig `slice-admin-tls-betrieb` (Kopf `3726eda0`, 627 Zeilen, grün in `make test` zu diesem Stand) und wird dort mit `git show` gelesen, der Zweig bleibt bis zum Start unberührt. Eine Änderung beim Aufnehmen: die Methode `ladeAdmin` (dort Zeile 155) steht seit dem Geber in `admintls_test.go` und fällt in der Datei weg (sonst doppelte Definition im Paket `bootstrap_test`); die übrigen 14 Namen der Datei (`adminIDs`, `adminListener`, `adminConn`, `adminRoute`, `adminServer`, `adminTokenWert`, `neuerAdminServer`, `adminTLS`, `dial`, `adminAntwort`, `holeRoh`, `erwarteFehler`, `adminHalt` und `pruefe503`) stehen nicht in `admintls_test.go` (Planner, 2026-11-23 gelesen); der Implementer prüft es beim Start gegen den dann gültigen Stand von `main` |
| `src/bootstrap/admintls.go`, `admintls_test.go`, `export_test.go` | **unberührt** (Geber `slice-admin-tls`) | die Funktionen und die Naht; braucht die Datei einen weiteren Namen in `export_test.go`, hält der Implementer an und gibt ihn an den Architect zurück (`AGENTS.md` §3.12); die Datei steht dann einzeln im Block (§4) |
| `src/adapters/driving/admin/` | **unberührt** | der Adapter kennt TLS nicht (`SPEC-299` Nr. 1); der Diff nennt keinen Pfad dieses Verzeichnisses (Review und Verifier lesen ihn) |
| `src/adapters/driving/admin/*_test.go` (die vier Wächter `TestIM16BinaryUnveraendert`, `TestI18BinaryUnveraendert`, `TestS17BinaryUnveraendert`, `TestWS15BinaryUnveraendert`) | **ersetzt von `slice-bootstrap-m6-endpunkte`** (Teil A des Schnitts von `slice-bootstrap-m6`, nicht von diesem Slice) | sie lesen alle `src/bootstrap/*.go`, auch `_test.go`, und melden jeden Import von `…/driving/admin`; mit dieser Datei ist `make test` rot, solange sie gelten (Wand, §4); der Slice startet erst nach der Ablösung |
| `src/bootstrap/tlsdatei.go`, `src/hexagon/ports/support/tlstest/`, `supporttest` | nutzen, unverändert | Laden und Prüfen des Paars (`ladeTLS`), Zertifikate zur Laufzeit, zählende Fakes; die Hilfen des Pakets `bootstrap_test` (`werkstatt`, `httpsClient`) darf der Test nutzen, nicht ändern |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag `SPEC-299` in der Spezifikation vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); sie kommen aus dem Plan des Gebers (Architect, 2026-11-23) und werden hier unverändert geführt.

**Zusage · Fänger** (`AGENTS.md` §3.15; Architect, 2026-11-23; Kennung `AT` wie beim Geber, die Nummern AT01 bis AT10 und AT18 bis AT23 stehen in dessen Plan; die Spalte *Zusage* nennt den Ort in `SPEC-299`; alle Tests laufen im Paket `bootstrap_test` ohne Netz außer dem Loopback des Prozesses, kein Schlüssel im Repo; die Namen der Funktionen stehen hier, wie `SPEC-299` sie führt, der Test erreicht sie über die Naht in `export_test.go`):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| AT11 | (Nr. 7) Die Verbindung folgt `SPEC-291` Nr. 5 und 6: TLS 1.2 und 1.3 gelingen, 1.1 scheitert; ALPN: allein `h2` scheitert, `h2` und `http/1.1` und kein ALPN und `http/1.1` sprechen HTTP/1.1; ein Client ohne Zertifikat gelingt (kein mTLS); die Kette der Datei wird gezeigt | `TestAdminVerbindung`, Menge M-Verbindung: Version (1.1, 1.2, 1.3) + ALPN (4 Angebote) + Client ohne Zertifikat + Kette mit Zwischenzertifikat (`len(PeerCertificates)` gleich 2) = 9 Zellen am Test-Server; Orakel: Literal je Zelle, `ConnectionState().Version` und `NegotiatedProtocol` | `MinVersion` TLS 1.3 (rot: 1.2); `MinVersion` TLS 1.0 (rot: 1.1); `NextProtos` um `h2` ergänzt (rot: 2 Zellen, allein `h2` und `h2` mit `http/1.1`); `ClientAuth` auf `RequireAnyClientCert` (rot); nur das Blatt statt der Kette (rot) | andere |
| AT12 | (Nr. 10 d) Der Rahmen gilt über TLS wie im Klartext: dieselbe Antwort, dieselbe `Anfrage` an die Route, dieselben Einträge und Zähler | `TestAdminRahmenUeberTLS`, Menge M-Anfrage: Anfrage (`GET /healthz` ohne Token, Route mit Token, Route ohne Token, Route mit falschem Token) × Transport (Klartext, TLS) = 8 Zellen; Orakel: Status, Body und Header (ohne den Wert von `X-Correlation-Id`) je Anfrage gleich, die `Anfrage` der Route gleich (Fake hält sie fest), `Audit` mit einem Eintrag `admin_abgewiesen` je 401, `CountError` je Fehler einmal | der Adapter ergänzt bei `r.TLS != nil` einen Header der Anfrage (rot: 4 Zellen mit TLS); `GET /healthz` verlangt über TLS ein Token (rot) | andere |
| AT13 | (Nr. 10 b) Ein Handshake, der scheitert, erreicht den Adapter nicht: kein `Handler`, kein `KorrelationsID`, kein `Append`, kein `CountError`, kein `X-Correlation-Id`, keine Zeile in einer Senke; ein Klartext-Request an den TLS-Port erhält nie die Antwort des Rahmens | `TestAdminHandshakeScheitert`, Menge M-Ausgang: unbekannte CA, Klartext-Request an den TLS-Port, Version 1.1, allein `h2` = 4 Zellen; Orakel: der Client sieht den Fehler (oder die Antwort der Standardbibliothek ohne `X-Correlation-Id` und ohne Fehlerobjekt), alle Zähler des Test-Servers 0, die Senke leer | Rückfall auf Klartext: der umhüllte Listener reicht eine Verbindung mit gescheitertem Handshake als Klartext an den Adapter (rot: Zelle Klartext-Request, der Handler läuft); der `ErrorLog` des Adapters schreibt statt zu verwerfen (rot: die Senke trägt `http: TLS handshake error` in den Zellen unbekannte CA, Version 1.1 und `h2`) | andere |
| AT14 | (Nr. 10 a) Ein Client, der verbindet und nichts sendet, hält weder `Accept` noch einen anderen Client auf | `TestAdminHaengenderHandshake`, Menge M-Haengend: 1 und 8 stumme Verbindungen, offen vor dem zweiten Client; ein Client mit der CA erhält `GET /healthz` 200 innerhalb einer Frist, die kein Schlaf ist (Kanal, `context` des Clients); Orakel: Status 200; deterministisch durch die Reihenfolge (die stumme Verbindung ist angenommen, bevor der Client beginnt); Rate: rot in n von n Prozessläufen | der Handshake im `Accept` (die Hülle von `umhuelle` ruft `Handshake()` vor der Rückgabe): rot in jedem Lauf | nebenläufig |
| AT15 | (Nr. 10 c) Nach `Anhalten` erhält ein Request über TLS 503 `ODR-E5010` mit `Connection: close` und `X-Correlation-Id`, auch wenn sein Handshake erst nach `Anhalten` endet | `TestAdminAnhaltenUeberTLS`, Menge M-Anhalten: (i) eine fertige TLS-Verbindung sendet nach `Anhalten` einen Request, (ii) eine Verbindung, deren Handshake vor `Anhalten` begann und danach endet (Haltepunkt: eine Hülle um das `net.Conn` des Clients hält seinen letzten Teil des Handshakes bis nach der Rückkehr von `Anhalten`, gesteuert durch einen Kanal), dann der Request; Orakel: 503, Code, Header, kein Aufruf der Route | `Anhalten` schließt die offenen Verbindungen des TLS-Listeners (rot: (i) bricht statt 503); die Zulassung wird beim `Accept` statt je Request geprüft (rot: (ii) liefert 200); jeweils rot in jedem Lauf | nebenläufig |
| AT16 | (Nr. 10 c) `Schliessen` schließt eine Verbindung, deren Handshake nicht beendet ist, sofort und wartet nicht auf sie | `TestAdminSchliessenOffenerHandshake`, Menge M-Offen: eine stumme Verbindung, eine mit einem halben ClientHello; `Schliessen(frist)` kehrt mit `nil` vor Ablauf der Frist zurück, die Verbindung ist geschlossen (Lesen liefert einen Fehler, nie einen Timeout des Tests) | `Schliessen` wartet auf offene Verbindungen (rot: `ErrFristAbgelaufen`); `Schliessen` lässt die Verbindung offen (rot: Lesen läuft in den Timeout); jeweils rot in jedem Lauf | nebenläufig |
| AT17 | (Nr. 11) Der Start liest die Dateien einmal: ein Wechsel nach dem Laden ändert das gezeigte Zertifikat nicht (kein Neuladen), ein Neustart zeigt das neue; läuft das Zertifikat im Betrieb ab, bedient der Listener es weiter | `TestAdminKeinNeuladen`, Menge M-Neuladen: Dateien überschrieben nach dem Laden und vor dem ersten Handshake, nach dem ersten Handshake, ein Zertifikat, das im Betrieb abläuft (Uhr des Clients nach dem Ablauf); Orakel: die Seriennummer des ersten Blatts, der Client mit der Wurzel des zweiten scheitert, der zweite Aufruf von `pruefeAdminTransport` (Neustart) zeigt das dritte | `GetCertificate` oder `GetConfigForClient` lädt die Dateien bei jedem Handshake (rot: 3 Zellen) | andere |

**Zeilen:** AT11 bis AT17 gleich 7; *a* = 3 (AT14, AT15, AT16), *b* = 4 (AT11, AT12, AT13, AT17), `SPEC-128` Nr. 3 e. Die Mutationen `m10d` und `m10e` (am Handshake von `umhuelle`, Zeile AT10 des Gebers) färben allein Tests dieser Datei (`TestAdminVerbindung` und `TestAdminHandshakeScheitert`) und stehen darum hier; die Zeile AT10 bleibt beim Geber.

**Mutationen und Raten** (Implementer des Gebers, 2026-11-23, an der Datei auf dem Zweig `slice-admin-tls-betrieb`, Kopf `3726eda0`; vom Planner aus dem Plan des Gebers übernommen; `AGENTS.md` §3.10, §3.15, §3.17; der Implementer dieses Slice fährt sie beim Start in der Wegwerfkopie neu und nennt Abweichungen). *Messende:* ein Treiber des Implementers im Scratchpad (Wegwerfkopie des Baums je Mutation, Ersetzung genau einer Stelle je Änderung, `go test -count=1 -run '^TestAdmin' ./src/bootstrap/` im Image `odata-recorder:test`, ohne Netz und ohne schreibbaren Mount, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`, nacheinander); `tools/harness/mutationslauf.sh` fährt `src/bootstrap` nicht (Frist der Nullprobe 2 Minuten unter der Paketzeit, `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`). Last des Hosts bei den Läufen 2,4 bis 20,2 (`last=` der Ergebniszeilen, Fremdlast anderer Linien; 20 Prozessoren im Host, 4 im Container). Die Mutationen sind in `admintls.go` (wo nicht anders genannt) gesetzt; `lebenszyklus.go`, `admin.go` im Paket `admin` und `tlsdatei.go` wurden nur in der Wegwerfkopie geändert.

| Zeile | Mutationen (Name: Änderung) | roter Test |
|---|---|---|
| AT10 (Anteil dieser Datei) | `m10d` `umhuelle` hängt `h2` an die geteilte Konfiguration · `m10e` ändert deren `MinVersion` | `TestAdminVerbindung` und `TestAdminHandshakeScheitert` |
| AT11 | `m11a` `MinVersion` 1.3 · `m11b` 1.0 · `m11c` `h2` angeboten · `m11d` Client-Zertifikat verlangt · `m11e` nur das Blatt · `m11f` `MaxVersion` 1.2 · `m11g` ohne `NextProtos` (alle in `tlsdatei.go`) | `TestAdminVerbindung` |
| AT12 | `m12a` Header der Anfrage bei `r.TLS != nil` · `m12b` `GET /healthz` verlangt über TLS ein Token · `m12d` `RawQuery` bei TLS | `TestAdminRahmenUeberTLS` (`m12a`, `m12d`: die Zelle „Route mit Token“; `m12b`: die Zelle `/healthz`) |
| AT13 | `m13a` Rückfall auf Klartext (Verbindung mit gescheitertem Handshake als Klartext an den Adapter) · `m13b`, `m13c` `ErrorLog` des Adapters schreibt nach `os.Stderr` und nach `log.Default()` | `TestAdminHandshakeScheitert` |
| AT14 | `m14a` Handshake im `Accept` | `TestAdminHaengenderHandshake` (rot in beiden Zellen, 1 und 8 stumme Verbindungen, je nach 5 s Frist des Clients) |
| AT15 | `m15a` `Anhalten` schließt die offenen Verbindungen · `m15b` Zulassung immer · `m15d` ohne `Connection: close` · `m15e` ohne `X-Correlation-Id` bei 503 | `TestAdminAnhaltenUeberTLS` (jede der vier in beiden Teilen (i) und (ii): `m15a` Lesen schlägt fehl, `m15b` Status 200, `m15d` kein `Connection: close`, `m15e` keine ID) |
| AT16 | `m16a` `Shutdown` wartet · `m16c` `schliesseOhneHandler` und Abbruch des Kontexts entfallen beide | `TestAdminSchliessenOffenerHandshake` (`m16a`: Dauer über 2 s in beiden Zellen, `m16c`: Lesen läuft in den Timeout in beiden Zellen) |
| AT17 | `m17a` `GetConfigForClient` lädt bei jedem Handshake mit der Uhr des Hosts · `m17b` mit der Uhr des Starts · `m17c` `GetCertificate` · `e01` `ladeTLS` mit der Uhr des Hosts statt `jetzt` | `TestAdminKeinNeuladen` |

**Grüne Mutationen und Maß** (`AGENTS.md` §3.15, akzeptierte Negative; eigene Mutationen an Elementen, die die Tabelle nicht nennt): (1) *Schließen offener Handshakes über zwei Wege*: `Schliessen` schließt eine Verbindung ohne Handler über `schliesseOhneHandler` und über den Abbruch des Basis-Kontexts, den `tls.Conn.HandshakeContext` beobachtet; `m16b` (nur der erste Weg entfällt) und `m16d` (nur der zweite) sind grün, `m16c` (beide) rot. Grund: jeder Weg allein erfüllt die Zusage von `SPEC-299` Nr. 10 c, die Mutation ist für den Aufrufer gleichwertig; Folge: ein Fehler, der einen der beiden Wege entfernt, bleibt grün, solange der andere trägt; Maß: 2 von 3 Mutationen am Schließen grün. (2) *`Anhalten` schließt den Listener nicht* (`m15f`): grün, 1 von 1; der Fänger dafür gehört den Tests des Adapters (`SPEC-136` Nr. 13, nicht `SPEC-299`); AT15 prüft die Wirkung nach `Anhalten`, nicht ihren Mechanismus.

**Mengen mit Zählprüfung im Test** (Literal gegen gezählte Zellen; Generatoren, Orakel und Zahl der Elemente, `.claude/commands/implement-slice.md` Schritt 14): M-Verbindung 9 · M-Anfrage 8 · M-Ausgang 4 · M-Haengend 2 · M-Anhalten 2 · M-Offen 2 · M-Neuladen 3. Die Fakes zählen jede Operation: `adminIDs` (alle fünf Operationen), `FakeAudit` (`Aufrufe`), `FakeTelemetry` (alle Aufrufe), die Route des Test-Servers und der Listener (Annahmen, Schließungen). *Nullblöcke:* der Implementer misst beim Start unter `-race -coverprofile` im Paket `bootstrap` (`-run '^TestAdmin'`) die Blöcke von `admintls.go` mit der Datei dieses Slice; der Geber nennt 16 Blöcke, 0 mit Zähler 0, gemessen am Zweig mit dem Teil *Betrieb*.

**Raten der Klasse `nebenläufig`** (`AGENTS.md` §3.10; Messende ein Skript des Implementers des Gebers im Scratchpad: ein Test-Binary je Arm, ohne und mit `-race`, N Prozesse hintereinander mit `-test.count=1 -test.timeout=120s`, gezählt die Prozesse mit Exit ungleich 0; Container `--cpus=4`, `GOMAXPROCS` per Umgebung auf 1 und auf 4; Host mit 20 Prozessoren, Last 1,8 bis 17,4 von 20 in den Ergebniszeilen, Fremdlast anderer Linien). Ohne Mutation, die vier Tests AT14, AT15, AT16 und AT18 (der letzte gehört dem Geber) zusammen, 50 Prozesse je Arm: `GOMAXPROCS=1` 0 von 50 ohne und 0 von 50 mit Detektor; `GOMAXPROCS=4` 0 von 50 und 0 von 50; die ganze Reihe `^TestAdmin` (30 Prozesse je Arm): 0 von 30 und 0 von 30 bei beiden Werten. Mutiert, rot von N Prozessen (`GOMAXPROCS` 1 und 4 gleich): `m14a` (Handshake im `Accept`, AT14) 20 von 20 ohne Detektor, jeder Lauf 10 s; `m15a` und `m15b` (AT15) je 50 von 50 ohne Detektor; `m16a` (AT16) 20 von 20 ohne Detektor, `m16b` (nur ein Weg entfällt, grün) 0 von 20. AT14 bis AT16 fangen durch die Reihenfolge, nicht durch einen Treffpunkt der Prozessoren (die Verbindung steht vor dem zweiten Schritt, ein Kanal ordnet), darum hängt ihre Rate nicht an `GOMAXPROCS`. Folge: jede nebenläufige Zusage dieser Zeilen ist in jedem Lauf gefangen (n von n). Die Raten tragen den Stand des Gebers; der Implementer dieses Slice misst sie beim Handoff neu, mit Zahl der Prozesse, Last des Hosts und Messendem.

**Pflicht-Mengen mit Zählprüfung** (Architect, 2026-11-23; Orakel als Literal, nie aus dem Code und nie als Aufruf der Funktion; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15; die Zahl der Zellen steht im Test als Literal und wird gegen die gezählten Zellen verglichen). Der Modus (`record`, `replay`, `mock`, `hybrid`) ist bei keiner Menge eine Achse: die Funktionen kennen keinen Modus (AT22 des Gebers).

| Menge | Zeile | Achsen | Elemente |
|---|---|---|---|
| M-Verbindung | AT11 | Version (1.1, 1.2, 1.3) · ALPN (allein `h2`, `h2` und `http/1.1`, keines, `http/1.1`) · Client ohne Zertifikat · Kette mit Zwischenzertifikat | 3 + 4 + 1 + 1 = 9 |
| M-Anfrage | AT12 | Anfrage (`GET /healthz` ohne Token, Route mit Token, Route ohne Token, Route mit falschem Token) · Transport (Klartext, TLS) | 4 × 2 = 8 |
| M-Ausgang | AT13 | Art des Scheiterns (unbekannte CA, Klartext an den TLS-Port, Version 1.1, allein `h2`) | 4 |
| M-Haengend | AT14 | Zahl der stummen Verbindungen (1, 8) | 2 |
| M-Anhalten | AT15 | Zeitpunkt des Handshakes (fertig vor `Anhalten`, endet nach `Anhalten`) | 2 |
| M-Offen | AT16 | Zustand der Verbindung (stumm, halbes ClientHello) | 2 |
| M-Neuladen | AT17 | Zeitpunkt der Änderung (vor dem ersten Handshake, danach) · Ablauf im Betrieb | 2 + 1 = 3 |

**Unbeschränkte Mengen und ihre Prüfgrenze** (akzeptierte Negative; Grund, Folge, Maß; `AGENTS.md` §3.15, §3.17): (1) *Zahl der stummen Verbindungen* (AT14): Prüfgrenze 1 und 8; Grund: mehr Verbindungen kosten Laufzeit und ändern die Klasse nicht; Folge: ein Hängen, das erst mit der neunten Verbindung einträte, bliebe grün, die Standardbibliothek hat keine solche Grenze im `Accept`; Maß: 8 von unbeschränkt. (2) *Formen der Dateien und des Zertifikats*: Prüfgrenzen und Folgen von `SPEC-291` (*Ungeprüft*, Zeilen A01 bis A17 und die Mengen von `slice-listener-tls`); dieser Slice ändert sie nicht.

**Größe** (Planner, 2026-11-23, nach dem Schnitt des Gebers; `SPEC-128` Nr. 1, `AGENTS.md` §3.14; vom Architect genannt: *a* = 3, *b* = 4, *f* = 1 (der Test-Server mit zählenden `IDs`, `Audit` und `Telemetrie` aus `supporttest`), *p* = 0 (kein Produktcode); der Planner hat die Zeilen der Tabelle oben gezählt und nachgerechnet): *S* = 85·3 + 80·4 + 200·1 + 0 = 255 + 320 + 200 = 775; 135·775 = 104625, *M* = 100 · ⌈10,4625⌉ = **1100**; keine Schnittfrage (*S* = 775 < 2223). Gemessen ist die Datei am Zweig des Gebers mit 627 Zeilen (Messung des Implementers: Summe aus Einfügungen und Löschungen von `git diff --numstat`), also 527 unter der Marke; `ladeAdmin` (rund 10 Zeilen) fällt weg.

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Prüfung der Sendungen** (`AGENTS.md` §3.13; Planner, 2026-11-23): *an diesen Slice* der Teil *Betrieb* aus [slice-admin-tls](../done/slice-admin-tls.md) (Zeilen AT11 bis AT17, Entscheidung des Architect): angenommen, §1 *Übernimmt* und die DoD tragen ihn mit der Kennung des Gebers; kein *Ausdrücklich NICHT* dieses Slice trifft ihn. *Von diesem Slice an:* keine. *Gelesen für den Start-Trigger* ([slice-bootstrap-m6-endpunkte](slice-bootstrap-m6-endpunkte.md), Teil A des Schnitts von `slice-bootstrap-m6`, 2026-11-23): die Ablösung der vier Wächter durch Fänger der Verdrahtung steht dort in §1 (*Übernimmt*), in der DoD und in §6 mit Ausgang *entschieden* (Zeilen A02 bis A07 der Tabelle *Zusage · Fänger*, Architect, 2026-11-23); der Ausschluss dort (Listener, Transportschutz, M6-Test, `HYBRID_*`, `FALLBACK_*`, `LISTEN_TLS_*`, neue Endpunkte, Umschalten des Modus) trifft sie nicht.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-admin-tls](../done/slice-admin-tls.md) (Geber der Funktionen und der Naht) und alle vier Teile des Schnitts von `slice-bootstrap-m6` ([slice-bootstrap-m6-endpunkte](slice-bootstrap-m6-endpunkte.md), [slice-bootstrap-m6-listener](slice-bootstrap-m6-listener.md), [slice-bootstrap-m6-transport](slice-bootstrap-m6-transport.md) und [slice-bootstrap-m6](slice-bootstrap-m6.md) als Teil D, der als letzter startet und die drei anderen voraussetzt; „`slice-bootstrap-m6` in `done/`“ heißt damit „alle vier Teile in `done/`“, Planner, 2026-11-23) liegen in `done/`. Die Spezifikation trägt `SPEC-299` Nr. 7, 10 und 11 samt Randformen (geliefert, `AGENTS.md` §3.12).

**Warum:** der Test ruft `pruefeAdminTransport` und `umhuelle` über die Naht des Gebers (*Code-Abhängigkeit*). Der zweite Grund ist eine **Wand**: der Test-Server führt `admin.New` in einer Datei von `src/bootstrap`, und vier Tests im Paket `admin` (`TestIM16BinaryUnveraendert`, `TestI18BinaryUnveraendert`, `TestS17BinaryUnveraendert`, `TestWS15BinaryUnveraendert`) verbieten jeder Datei von `src/bootstrap/*.go`, auch jeder `_test.go`, den Import von `admin`; mit der Datei ist `make test` rot. Teil A des Schnitts von `slice-bootstrap-m6` importiert `admin` im Root ohnehin und löst die Wächter durch Fänger der Verdrahtung ab (die Zusage „Binary unverändert“ endet mit ihr); erst danach darf eine Datei von `src/bootstrap` `admin` nennen. Eine Ausnahme der Datei in den Wächtern lockerte vier Fänger fertiger Slices (`AGENTS.md` §3.6) und bleibt eine Entscheidung des User, sie ist nicht gewählt. Ein Prüfpaket unter `tests/` scheidet aus: der Import eines Driving Adapters aus `tests/` ist ein `wrong-direction` (`SPEC-039` Nr. 2, [ADR-0012](../../adr/0012-pruefpakete-unter-tests-als-eigene-schicht-des-architektur-gates.md)), und `umhuelle` ist von dort nicht erreichbar.

**Reihenfolge und Kollisionen:** der Slice liegt **nach** `slice-bootstrap-m6` (`welle-admin-api` §5); er ändert nur eine neue Datei und kollidiert mit keinem Slice des Pakets `bootstrap` außer über `export_test.go`, falls die Datei einen weiteren Namen braucht (der Planner prüft es beim Start gegen `main`, `SPEC-128` Nr. 10 g). `slice-bootstrap-m4-konfig` und `slice-bootstrap-m4` laufen nicht zugleich mit `slice-bootstrap-m6`; eine Kante zu diesem Slice haben sie nicht.

**Stand der Bedingungen** (Planner, 2026-11-23, an den Verzeichnissen gelesen): `slice-admin-tls` liegt in `in-progress/`, die vier Teile des Schnitts von `slice-bootstrap-m6` in `open/`; **nicht erfüllt**. Der Planner prüft beim Start neu und setzt `basis`, `schaetzung` und `marke` in den Block.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-23, Rechnung des Planner): *S* = 775 gleich der Tabelle (7 Zeilen: *a* = 3, *b* = 4; *f* = 1; *p* = 0, Rechenweg in §3 *Größe*), **Marke 1100** gleich der Formel (135·775 = 104625, ⌈10,4625⌉ = 11); keine Schnittfrage. Der Block trägt den Teil des Architect; **`basis`, `schaetzung` (775) und `marke` (1100) setzt der Planner beim Start vor dem ersten Code-Commit** (`AGENTS.md` §3.14, `BEO-REPO/marke-beim-start-nicht-gesetzt`): `basis` ist der letzte Commit von `main` vor dem Code-Start und muss dann ein Vorfahr von `HEAD` sein, der Planner prüft ihn mit `git merge-base --is-ancestor` und nennt die Messbasis nach `SPEC-128` Nr. 10 g (Fremdzeilen am Pfad: die Datei gibt es auf `main` vorher nicht, also 0). Das Werkzeug (`make groessenmarke`) prüft den Plan erst in `in-progress/`; in `open/` meldet es ihn nicht. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; die Pfade nennen die Dateien einzeln. Legt der Implementer eine weitere Datei an oder ändert er eine weitere Datei (etwa `export_test.go`), nennt er sie im selben Commit im Block.

```groessenmarke
pfad: src/bootstrap/admintls_betrieb_test.go
produktcode: 0
fakes: 1
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Ein Schnitt beim Handoff ist nicht vorgesehen.
- `in-progress` → `open` (blockiert — Carveout?): die Datei lässt sich nicht aufnehmen, weil ein Wächter oder `make test` weiter rot ist (die Ablösung durch `slice-bootstrap-m6` fehlt oder deckt `admintls_betrieb_test.go` nicht), oder der Implementer bemerkt eine Randform ohne Ort in `SPEC-299` (`AGENTS.md` §3.12); beides geht mit Datum und Adresse Architect nach §6.

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23, aus dem Plan des Gebers): jede Zelle nennt ihren Ort in `SPEC-299`, oder „gilt nicht“ mit Grund; keine Zelle ist leer. Die übrigen Operationen des Vertrags (Host klassifizieren, `ADMIN_ALLOW_PLAINTEXT` lesen, Weg bestimmen, Paar laden, Listener umhüllen, Warnung) stehen im Raster des Gebers.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| HTTPS bedienen (Handshake und Anfrage) | `SPEC-299` Nr. 7: ein Client ohne Zertifikat gelingt (kein mTLS), ohne ALPN gilt HTTP/1.1 | `SPEC-299` Nr. 10 a und b: ein Client, der nichts sendet, hält nichts auf; ein Klartext-Request an den TLS-Port erreicht den Rahmen nie | `SPEC-299` Nr. 7 und 10 b: unbekannte CA, Version unter 1.2 und allein `h2` scheitern im Handshake und erreichen den Adapter nicht | `SPEC-299` Nr. 10 a und c: `Anhalten` und `Schliessen` über TLS wie `SPEC-238` Nr. 11 | `SPEC-299` Nr. 11 und 10 d: kein Neuladen der Dateien, dieselbe `Anfrage` bei jedem Handshake |

**Risiken und Ausgänge:**

- Mindestversion und Cipher von TLS — **Ausgang:** entschieden (Architect, 2026-11-23): wie `SPEC-291` Nr. 5 und 6, keine eigene Entscheidung (`SPEC-299` Nr. 7); AT11.
- **Wand: vier Wächter im Paket `admin`** verbieten jeder Datei von `src/bootstrap` den Import von `admin` (siehe §4 *Warum*; Geber `slice-admin-tls`, Plan-Defekt dort) — **Ausgang:** entschieden (Architect, 2026-11-23, Plan des Gebers §6): Folge-Slice nach `slice-bootstrap-m6`, keine Ausnahme der Wächter; die Ablösung der Wächter durch Fänger der Verdrahtung trägt Teil A des Schnitts von `slice-bootstrap-m6` ([slice-bootstrap-m6-endpunkte](slice-bootstrap-m6-endpunkte.md): §1, DoD und §6, Ausgang *entschieden*, Tabelle A02 bis A07 vom Architect, 2026-11-23); dieser Slice startet nach allen vier Teilen (Planner, 2026-11-23). Folge: bis dieser Slice in `done/` liegt, ist der Handshake des Adapters über TLS nicht belegt; die Funktionen ruft bis `slice-bootstrap-m6` nichts im Binary, der Handshake von `umhuelle` selbst belegt AT10 des Gebers. Hängt der User die Ausnahme der Wächter an, ist das eine Entscheidung des User und ändert Trigger und Reihenfolge (dann liefert der Geber die Datei).
- Doppelte Methode `ladeAdmin`: die Datei auf dem Zweig `slice-admin-tls-betrieb` und `admintls_test.go` des Gebers definieren sie beide, das Paket `bootstrap_test` lässt sich dann nicht übersetzen — **Ausgang:** entschieden (Planner, 2026-11-23): die Methode fällt in der Datei dieses Slice weg (§3); alle anderen Namen sind frei, geprüft gegen den Stand des Gebers; der Implementer prüft es beim Start erneut (Namen anderer Slices des Pakets, etwa `audit_betrieb_test.go`, `tls_betrieb_test.go`).
- Zeilen fremder Slices an den Pfaden des Blocks (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`): der Block nennt nur die neue Datei, die es auf `main` nicht gibt; braucht der Implementer `export_test.go`, steht sie einzeln im Block, und jede fremde Zeile dort zählt mit — **Ausgang:** weiter offen; Adresse: Implementer (misst beim Handoff `git diff --numstat <basis>..<main-Elter des letzten Merge von main> -- <Pfade des Blocks>` und nennt die fremden Zeilen).
- Größe: Marke 1100 (*S* = 775, Rechenweg in §3, keine Schnittfrage) — **Ausgang:** weiter offen; Adresse: Planner (setzt `basis`, `schaetzung` und `marke` beim Start in den Block, §4; `BEO-REPO/marke-beim-start-nicht-gesetzt`).
- Die Mutationen, Raten und Mengen stammen aus der Messung des Gebers an der Datei auf dem Zweig `slice-admin-tls-betrieb`; nach dem Aufnehmen (ohne `ladeAdmin`, auf dem Stand von `main` nach `slice-bootstrap-m6`) gelten sie als Größenordnung, nicht als Messung dieses Slice (`AGENTS.md` §3.10: eine Rate steht mit Last und Messendem) — **Ausgang:** weiter offen; Adresse: Implementer (fährt die Mutationen und Raten der Tabelle beim Handoff selbst, `m10d`, `m10e`, `m11*` bis `m17*`, `e01`, und trägt sie in §3 ein).
- Die Begrenzung der Fehlversuche am Admin-Listener (`LH-FA-13.a` nennt keine) — **Ausgang:** entschieden (User, 2026-11-24, Variante A, `SPEC-301`): nicht Gegenstand dieses Slice, Träger ist der Folge-Slice [slice-admin-abweisen-begrenzen](slice-admin-abweisen-begrenzen.md) (`SPEC-299` *Ungeprüft*: dieser Vertrag ändert sie nicht, TLS schützt das Token auf dem Weg, nicht gegen Raten; Rückgabe R5 des Architect, `welle-admin-api` §5).
- `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (20×, verkörpert in `AGENTS.md` §3.10) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×, `AGENTS.md` §3.10) für AT14 bis AT16: Ordnung durch Kanäle, Raten mit Zahl der Prozesse und Messendem — **Ausgang:** weiter offen → `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` im Register.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-23, Zähler wie in §8 von `slice-admin-tls`; der Planner zieht Treffer und Zähler beim Start nach). Offen und für den Slice einschlägig: `BEO-REPO/marke-beim-start-nicht-gesetzt` und `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen` (6×) für die Größenmarke (§4). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (20×, `AGENTS.md` §3.10) und `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (9×, `AGENTS.md` §3.10) für AT14 bis AT16, `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (22×, Übergangsregeln Nr. 8) und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (75×, `AGENTS.md` §3.15). Keiner erreicht mit diesem Slice 3×, den neuen Befund (die Wächter fertiger Slices zählen den Import von `admin` in `src/bootstrap` bis in `_test.go`, der Plan des Gebers nannte die Bedingung nicht) trägt der Implementer des Gebers oder der Reviewer ins Register ein, nicht dieser Plan.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
