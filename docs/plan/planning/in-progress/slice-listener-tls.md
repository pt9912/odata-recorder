# Slice slice-listener-tls: wiederverwendbarer TLS-Baustein `ladeTLS` und Test-Hilfspaket `tlstest` (Teil 1 des Schnitts; Konfiguration, Listener und Handbuch liegen in `slice-listener-tls-verdrahtung`)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-betrieb](../welle-betrieb.md).

**Bezug:** `LH-RB-02`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-RB-02.a` (HTTP/TLS-Konfiguration; der Nachweis, Integrationstests für HTTP, HTTPS und Zertifikatsfehler, liegt in Teil 2) · `SPEC-291` (neu, Architect, 2026-11-21: Kopf und Nr. 1 bis 6, 10 bis 12: Baustein `ladeTLS`, Gründe, Gültigkeit, Schlüsselformen, ALPN, Version, Meldung, Zustandslosigkeit) · `SPEC-089` (Admin, Gegenstück) · `SPEC-114` (Uhr) · `ARC-005`, `ARC-012`. Die Stellen von Start, Handshake und Konfiguration (`SPEC-088`, `SPEC-110`, `SPEC-055`, `SPEC-139`, `SPEC-136`, `SPEC-059`) berührt Teil 2.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-listener-tls`, Worktree `/Development/odata-recorder-wt-listener-tls`; der Zweig besteht und nimmt `main` nach dem Start-Commit auf)

**Autor:** pt9912. **Datum:** 2026-10-30; geschnitten vor dem Code am 2026-11-21 (Planner, Teil 2: [slice-listener-tls-verdrahtung](../open/slice-listener-tls-verdrahtung.md)).

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Baustein `ladeTLS` (`src/bootstrap/tlsdatei.go`, `SPEC-291` Nr. 1) lädt das Paar aus Zertifikat- und Schlüsseldatei, prüft es (sechs Gründe in vier Rängen je Schlüssel, Gültigkeit des Blatts zur übergebenen Zeit, fünf Schlüsselformen) und liefert ein `*tls.Config` (Mindestversion TLS 1.2, ALPN allein `http/1.1`, die Kette der Datei, keine Client-Zertifikate) oder die Verstöße; er kennt keinen Schlüsselnamen, den der Aufrufer nicht übergibt, und der Admin-Listener (`welle-admin-api`) verwendet ihn mit `ADMIN_TLS_*` wieder, ohne ihn zu ändern. Dazu das Hilfspaket `tlstest` (Aussteller und Blätter zur Laufzeit) für die Tests dieses Teils und für Teil 2. Dieser Teil liefert nichts am Binary: `ladeTLS` ruft nur der Test; die Annahme von `LISTEN_TLS_*` in der Startkonfiguration, der HTTPS-Listener und der HTTPS-Test am Listener (`LH-RB-02.a`) liegen in [slice-listener-tls-verdrahtung](../open/slice-listener-tls-verdrahtung.md) (Teil 2). Der Schnitt ist am 2026-11-21 vor dem Code gelegt (§4).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Startkonfiguration (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`, Nachzug von `SPEC-110` Nr. 2, `PUBLIC_BASE_URL`), die Verdrahtung in `Build` und `Hochfahren`, die Tests am Listener des HTTP-Adapters, das Handbuch und das Kataloge-Gate — [slice-listener-tls-verdrahtung](../open/slice-listener-tls-verdrahtung.md) (Teil 2, hängt an diesem Teil); gegen einen Schnitt „Konfiguration ohne Verdrahtung“ spricht, dass der Start `LISTEN_TLS_*` annähme und der Listener bei HTTP bliebe, eine Sicherheitseinstellung ohne Wirkung (`SPEC-110` Nr. 2).
- TLS des Admin-Listeners, `ADMIN_TLS_*`, `ADMIN_ALLOW_PLAINTEXT` und die Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` — `welle-admin-api` (`slice-admin-tls`, Datei in `open/`, seine Bedingung nennt diesen Slice); dieser Slice liefert nur den Baustein und belegt die zweite Verwendung mit den Namen `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` (Zeile A17).
- TLS zum Upstream (`UPSTREAM_TLS_VERIFY`, `UPSTREAM_CA_FILE`) — Bestand, `slice-upstream-tls` in `done/`; Gegenstand ist die Clientseite, hier die Serverseite.
- Neuladen von Zertifikat oder Schlüssel im Betrieb, Client-Zertifikate (mTLS), Zertifikatsverwaltung — kein Anker im Lastenheft (`LH-RB-02` verlangt HTTPS und Zertifikatsprüfung, nicht Betriebsverfahren); der Start liest die Dateien einmal (`SPEC-291` Nr. 10).
- HTTP/2 über TLS und eine Prüfung der Rechte der Schlüsseldatei — vom Architect entschieden, nicht geliefert (`SPEC-291` Nr. 6 und 10, mit Alternativen und Folge): der Listener spricht auch mit TLS HTTP/1.1; die Zusagen des Adapters sind daran gemessen, und `LH-RB-02` verlangt kein HTTP/2. Ein Anker für beides fehlt im Lastenheft; ein späteres Angebot von HTTP/2 ist additiv und eine neue Entscheidung, kein Folge-Slice mit Kennung.
- Die Zusage „Transportverschlüsselung und Zertifikatsprüfung dürfen nicht ohne ausdrückliche Konfiguration abgeschaltet werden“ für den Upstream — Bestand (`UPSTREAM_TLS_INSECURE_SKIP_VERIFY`, `slice-upstream-tls`).

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

- [ ] Baustein lädt und prüft: `ladeTLS` meldet beide Pfade leer als kein TLS, genau einen leeren Pfad, eine unlesbare Datei, eine Datei ohne PEM-Zertifikat oder ohne PEM-Schlüssel, einen Schlüssel ohne passendes Zertifikat und ein Blatt außerhalb seiner Gültigkeit je mit dem einen Grund seines kleinsten Ranges und dem Namen des Aufrufers (`SPEC-291` Nr. 2 bis 4), nie mit Pfad, Inhalt oder Text der Bibliothek (Nr. 11); die zugesagten Formen der Datei und die fünf Schlüsselformen sind gültig (A01 bis A09, A14).
- [ ] Baustein liefert die Konfiguration: Mindestversion TLS 1.2, ALPN allein `http/1.1`, keine Client-Zertifikate, die Kette der Datei wird gesendet; Handshakes im Prozess mit dem Aussteller aus `tlstest` belegen es (`SPEC-291` Nr. 5, 6; A10 bis A13).
- [ ] Der Baustein `ladeTLS` ist ohne Änderung für den Admin-Listener nutzbar (eigene Schlüsselnamen als Parameter, kein Bezug auf `LISTEN_`), zustandslos und ohne Datenrennen; ein Test belegt die zweite Verwendung mit anderen Schlüsselnamen (`ADMIN_TLS_CERT`, `ADMIN_TLS_KEY`) (A15 bis A17).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch: das Binary bleibt unverändert (`ladeTLS` ruft nur der Test), kein Fehlercode, Schlüssel oder Kommando kommt dazu; das Handbuch-Kapitel TLS und das Kataloge-Gate für `LISTEN_TLS_*` gehören zu Teil 2 (Entscheidung des Planners vom 2026-11-21); §7 sagt, dass das Binary unverändert blieb.
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
| `src/bootstrap/tlsdatei.go` | neu | Baustein `ladeTLS` (Laden, Prüfen, `*tls.Config`), ohne Bezug zu `LISTEN_` (Ort: `SPEC-291` Nr. 1) |
| `src/bootstrap/tlsdatei_test.go` | neu | die Tests A01 bis A17 an `ladeTLS`; eine Testdatei, damit der Pfad im Block der Größenmarke genügt (§4) |
| `src/hexagon/ports/support/tlstest/` | neu | Hilfe für Tests: Aussteller und Blätter zur Laufzeit (Hierarchie, Ablauf, Namen, Schlüsselformen); nur Standardbibliothek, kein Produktcode; Teil 2 nutzt es für die Tests am Listener |
| `src/bootstrap/konfig*.go`, `wurzel.go`, `lebenszyklus.go`, `src/adapters/driving/httpodata/` | keine Änderung | Konfiguration, Verdrahtung und Tests am Listener liegen in Teil 2 ([slice-listener-tls-verdrahtung](../open/slice-listener-tls-verdrahtung.md)); der Adapter bleibt unverändert (`SPEC-291` Nr. 1, 9) |
| `spec/spezifikation.md` | update | geschehen (Architect): `SPEC-291`, `SPEC-110`, `SPEC-139`, `SPEC-136`, §3 |
| `.a-check.yml` | unverändert | der Baustein liegt im Composition Root, `tlstest` unter `ports/support` (nur Standardbibliothek); keine neue Kante |
| `docs/user/benutzerhandbuch.md` | keine Änderung | das Binary bleibt unverändert; Kapitel TLS und Zeilen der Tabelle der Einstellungen gehören zu Teil 2 |
| Tests | neu | Zertifikate und Schlüssel als Fixtures zur Laufzeit erzeugt (kein Geheimnis im Repo); Handshakes im Prozess über `net.Pipe` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Beide liegen vor (Architect, 2026-11-21).

**Aufträge an den Architect — entschieden (2026-11-21, Ort in der Spezifikation):**

1. **Ort und Schnitt des Bausteins** — *entschieden*, `SPEC-291` Kopf und Nr. 1: Datei `src/bootstrap/tlsdatei.go` im Composition Root, nicht in einem Adapter; beide Listener bleiben unverändert und erhalten einen umhüllten `net.Listener` (`SPEC-136` Nr. 1, `SPEC-238` Nr. 1); kein Import eines Driving Adapters im anderen, `.a-check.yml` unverändert. Parameter: `tlsDateien{certSchluessel, keySchluessel, certPfad, keyPfad}` und die Zeit; Rückgabe `(*tls.Config, []verstoss)`.
2. **Prüfung beim Start** — *entschieden*, `SPEC-291` Nr. 2 bis 5 und 10: sechs Gründe in vier Rängen je Schlüssel (leer, nicht lesbar, kein PEM, Gültigkeit des Blatts bzw. Schlüssel passt nicht); der Leser ist die Standardbibliothek (`encoding/pem`, `x509.ParseCertificate`, `tls.X509KeyPair`), der Baustein schreibt keinen eigenen Parser; Mindestversion TLS 1.2, Suiten der Bibliothek, kein mTLS, kein Neuladen. **Rechte der Schlüsseldatei: nicht geprüft** (`SPEC-291` Nr. 10, Alternative und Folge dort). Größe der Dateien nicht begrenzt (wie `UPSTREAM_CA_FILE`, *Ungeprüft*). Meldung nie mit Pfad, Inhalt oder Text der Bibliothek (Nr. 11).
3. **Nachzug `SPEC-110`** — *entschieden*, `SPEC-110` Nr. 2, 7, 11, 13, 14 c und e: `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` sind geliefert, jeder andere `LISTEN_TLS_*` bleibt `ODR-E2001`; Reihenfolge: reine Prüfungen, dann die Dateien (vor `DATA_DIR`); `PUBLIC_BASE_URL` abgeleitet trägt `https`, wenn beide Schlüssel gesetzt sind, eine gesetzte wird nicht abgeglichen (`SPEC-291` Nr. 7 d); Grenze der Länge 2025 Bytes bei TLS. Gegenstand von Teil 2 (Text dort).
4. **Protokolle** — *entschieden*, `SPEC-291` Nr. 5 und 6: ALPN nennt allein `http/1.1`; HTTP/2 wird nicht angeboten, weil die Zusagen von `SPEC-136` Nr. 4 h, 9, 13 und 14 an HTTP/1.1 gemessen sind; ein Client, der allein `h2` anbietet, scheitert im Handshake (Folge im Betrieb dort). Der Baustein trägt es (A11, A12), der Listener belegt es in Teil 2 (B11).
5. **Herunterfahren** — *entschieden*, `SPEC-291` Nr. 9 c und `SPEC-136` Nr. 13 bis 15 unverändert; Gegenstand von Teil 2 (Text dort).
6. **Fixtures** — *entschieden*: Hilfspaket `src/hexagon/ports/support/tlstest` (nur Standardbibliothek, kein Produktcode, erreichbar für Adapter-Tests und Composition Root): Aussteller und Blatt zur Laufzeit mit wählbarem `NotBefore`, `NotAfter`, Namen, Schlüsselform und Kette; die Dateien entstehen unter `t.TempDir()`; PEM mit Zeilenende, Fremdtext und fehlendem Zeilenende aus Bytewerten (`AGENTS.md` §3.18). Keine Datei mit Schlüssel im Repo (`LH-RB-04`).
7. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** — *entschieden*: Tabelle und Raster unten (die Zeilen `A`; die Zeilen `B` stehen in Teil 2), Block `groessenmarke` in §4 (`pfad`, `produktcode`, `fakes`); Gegenproben entfallen (kein neues Werkzeug des Harness).

**Hinweise an den Planner** (Architect, 2026-11-21):

- **Schnitt vor dem Code** (`SPEC-128` Nr. 1, Schnittfrage ab *S* ≥ 2223): die Tabelle unten trägt 39 Zeilen (*a* = 2, *b* = 37), *f* = 1, *p* = 180, also *S* = 85·2 + 80·37 + 200·1 + 180 = 3510 und Marke 4800, über der Schnittgrenze. Die Naht steht in §4 (Teil 1: Baustein, Zeilen `A01` bis `A17`; Teil 2: Konfiguration und Verdrahtung, Zeilen `B01` bis `B22`); die Nummern der Zeilen tragen den Teil. Der Planner hat sie am 2026-11-21 nachgerechnet und geschnitten (§4); Teil 2 ist der Plan `slice-listener-tls-verdrahtung`.
- **Kataloge-Gate** (`SPEC-201` Nr. 4): Gegenstand von Teil 2; der Text steht dort (§3 *Hinweise an den Planner*). Teil 1 ändert weder `konfig*.go` noch `nichtGelieferte` noch das Handbuch.
- **Nehmer `slice-admin-tls`:** ruft `ladeTLS` mit `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` (die Zeile `A17` belegt das mit diesen Namen) und umhüllt den Listener des Admin-Adapters wie `Hochfahren` den OData-Listener; die Pflicht bei nicht-loopback `ADMIN_LISTEN` und `ADMIN_ALLOW_PLAINTEXT` liegen nicht im Baustein (`SPEC-291` Kopf). Bei dem Schnitt trägt der Teil 1 die Kennung, die `slice-admin-tls` als Bedingung nennt: `slice-listener-tls` (der Planner hat sie dort am 2026-11-21 nachgetragen, `AGENTS.md` §3.13; `slice-listener-tls-verdrahtung` ist dort keine Bedingung).
- **Leser fremder Bytes** (§4 *Größenmarke*): der Architect überlässt das Lesen der PEM der Standardbibliothek (`SPEC-291` Nr. 4); der Zuschlag für Leser, den `AGENTS.md` §3.14 für eigene Prüfung nennt, entfällt, und die Formel von `SPEC-128` Nr. 1 kennt ihn nicht.
- **Spec-Nummer:** `SPEC-291` ist vergeben (die Nummer `SPEC-290` gehört dem Mock-Vertrag der Telemetrie, der zeitgleich auf `main` kam); die nächste freie ist `SPEC-292`.
- **Zweite Änderung an `SPEC-139`:** die Telemetrie des Mock (`SPEC-290`) hat `SPEC-139` Nr. 13 angelegt und Nr. 3 und 11 angepasst; Nr. 2, 6 und 8 c ändert Teil 2 (er berührt Nr. 13 nicht). Teil 1 ändert in `src/bootstrap/` nur neue Dateien; Teil 2 und `slice-telemetry-einbau-mock` ändern beide `wurzel.go`: der Planner hat die Reihenfolge festgelegt (§4 *Reihenfolge und Konflikte*).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect, der Implementer trägt die Mutationen und Raten mit Zahl der Prozesse und Messendem in einen Absatz darunter ein, `AGENTS.md` §3.10):

Die Nummern tragen den Teil des Schnitts in §4: `A` ist der Baustein (Teil 1), `B` Konfiguration und Verdrahtung (Teil 2). *Mengen* sind benannt (Zählprüfung unter der Tabelle); jeder Fänger einer Menge ist eine Schleife über alle ihre Elemente mit einem Orakel, das den Code des Bausteins nicht benutzt (`AGENTS.md` §3.15, §3.17). Der Fänger ist ein Test, ein Fall der Gegenprobe oder eine Ausgabe des Gates. Klassen: `nebenläufig` heißt, der Fänger zählt eine Wirkung über Prozesse oder Goroutinen; alles andere ist `andere`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| A01 | Beide Pfade leer: kein TLS, `ladeTLS` liefert `(nil, nil)` und liest nichts (`SPEC-291` Nr. 2) | `TestLadeTLSKeinTLS`: beide Pfade leer, im Arbeitsverzeichnis liegen Dateien der Namen `cert.pem` und `key.pem`; kein Verstoß, Ergebnis `nil` | der Zweig „beide leer“ entfällt: Rang 1 meldet für beide Schlüssel | andere |
| A02 | Genau einer der beiden Pfade leer: der leere Schlüssel meldet `erforderlich, wenn <Name des anderen> gesetzt ist`, der gesetzte wird weiter geprüft (Rang 1, Nr. 2) | `TestLadeTLSEinerLeer`, Menge E (4 Elemente: leerer Schlüssel × gesetzte Datei gültig oder fehlend): der Grund trägt den übergebenen Namen des anderen; die gesetzte Datei meldet genau dann `Datei nicht lesbar`, wenn sie fehlt | Namen der Richtungen vertauscht; die Prüfung des gesetzten Pfads entfällt, sobald der andere leer ist | andere |
| A03 | Rang 2 `Datei nicht lesbar`, nie mit dem Pfad (Nr. 2) | `TestLadeTLSNichtLesbar`, Menge L (6: Datei fehlt, Pfad ist ein Verzeichnis, Pfad mit NUL, je Zertifikat und Schlüssel): der Grund ist das Literal | der Fehler von `os.ReadFile` wird ignoriert (leerer Inhalt): die Fälle melden `kein PEM` | andere |
| A04 | Rang 3 Zertifikat `Datei enthält kein PEM-Zertifikat` (Nr. 2, 4) | `TestLadeTLSKeinPEMZertifikat`, Menge Z (7: leere Datei, nur Text, nur ein Schlüsselblock, Block `CERTIFICATE` mit Bytes, die kein Zertifikat sind, Block `CERTIFICATE` ohne `END`, nur ein Block `X509 CRL`, erster Block `CERTIFICATE` ungültig und zweiter gültig) | der erste Block wird übersprungen, wenn er nicht parst (Suche nach dem ersten lesbaren): das Element sieben wird gültig | andere |
| A05 | Rang 3 Schlüssel `Datei enthält keinen PEM-Schlüssel` (Nr. 2, 4) | `TestLadeTLSKeinPEMSchluessel`, Menge S (4: leere Datei, nur Text, nur ein Zertifikatblock, Block `PRIVATE KEY` ohne `END`) | die Suche nimmt jeden Block statt der Typen `PRIVATE KEY` und `… PRIVATE KEY`: der Zertifikatblock geht an `tls.X509KeyPair` und meldet Rang 4 | andere |
| A06 | Rang 4 Schlüssel `Schlüssel ungültig oder passt nicht zum Zertifikat`; ein Zwischenzertifikat vor dem Blatt, ein verschlüsselter und ein nicht lesbarer Schlüssel zählen dazu (Nr. 2, 4) | `TestLadeTLSSchluesselPasstNicht`, Menge P (6: Schlüssel eines anderen Blatts gleicher Art, Schlüssel anderer Art (RSA zu ECDSA-Blatt), Block `PRIVATE KEY` mit Bytes, die kein Schlüssel sind, Block `ENCRYPTED PRIVATE KEY`, Block `EC PRIVATE KEY` mit Bytes, die kein Schlüssel sind, Zwischenzertifikat vor dem Blatt) | der Aufruf von `tls.X509KeyPair` entfällt: jedes Element wird gültig | andere |
| A07 | Rang 4 Zertifikat `Zertifikat außerhalb seiner Gültigkeit` genau dann, wenn `jetzt` vor `NotBefore` oder nach `NotAfter` liegt; beide Grenzen gelten; geprüft wird allein das Blatt (Nr. 3) | `TestLadeTLSGueltigkeit`, Menge G (6: `NotBefore` minus 1 ns, `NotBefore`, Mitte, `NotAfter`, `NotAfter` plus 1 ns, gültiges Blatt mit abgelaufenem Zwischenzertifikat in der Kette): erwartet ungültig, gültig, gültig, gültig, ungültig, gültig | `After` statt `!Before` und umgekehrt (eine Grenze wird ungültig); die Prüfung läuft über alle Zertifikate der Kette | andere |
| A08 | Ein Grund je Schlüssel, die Schlüssel melden unabhängig, Rang n nur nach Bestehen der Ränge davor, der Rang 4 des Schlüssels nur, wenn beide Rang 3 bestanden (Nr. 2) | `TestLadeTLSPaare`, Menge Q (25 Paare: Zustand des Zertifikats {gültig, leer, nicht lesbar, kein PEM, außerhalb der Gültigkeit} × Zustand des Schlüssels {gültig, leer, nicht lesbar, kein PEM, passt nicht}); das Orakel bildet die erwartete Meldung je Paar aus der Tabelle von Nr. 2, ohne `ladeTLS`; Zählprüfung 25 | `return` nach dem ersten Verstoß (eine Zeile für das Paar); der Rang 4 des Schlüssels läuft ohne bestandenen Rang 3 des Zertifikats | andere |
| A09 | Gültige Formen der Dateien (Nr. 4): kein Verstoß, `*tls.Config` nicht `nil`, auch eine Schlüsseldatei mit Rechten `0666` (Nr. 10, Rechte nicht geprüft) | `TestLadeTLSFormen`, Menge F (8: Zeilenende LF, CRLF, Text vor den Blöcken, Text nach den Blöcken, Kette Blatt und Zwischenzertifikat, Zertifikat und Schlüssel in einer Datei mit beiden Pfaden gleich, kein Zeilenende am Ende, Schlüsseldatei `0666`); aus Bytewerten gebaut (`AGENTS.md` §3.18) | die Prüfung `Mode()&0o077 != 0` wie bei der Token-Datei (Element acht wird ungültig); ein verlangtes Zeilenende am Dateiende (Element sieben wird ungültig) | andere |
| A10 | Die fünf Schlüsselformen sind gültig und tragen einen Handshake (Nr. 4) | `TestLadeTLSSchluesselformen`, Menge K (5: RSA als `RSA PRIVATE KEY`, RSA als `PRIVATE KEY`, ECDSA als `EC PRIVATE KEY`, ECDSA als `PRIVATE KEY`, Ed25519 als `PRIVATE KEY`): kein Verstoß, Handshake im Prozess (`tls.Server` über `net.Pipe`) mit dem Aussteller als Pool gelingt | die Typprüfung der Blöcke auf `PRIVATE KEY` allein (ohne Endung `… PRIVATE KEY`): drei Elemente werden ungültig | andere |
| A11 | Mindestversion TLS 1.2, Höchstversion der Standard, keine Client-Zertifikate (Nr. 5) | `TestLadeTLSVersion`: das Feld `MinVersion` und Handshakes im Prozess: Client mit höchstens TLS 1.1 scheitert, mit höchstens TLS 1.2 verhandelt 1.2, ohne Grenze 1.3; ein Client ohne Zertifikat gelingt (fällt die Bibliothek ohne TLS 1.1 aus, prüft der Test das Feld und die beiden gelingenden Fälle) | `MinVersion: tls.VersionTLS10` (rot); das Weglassen gleicht dem Standard der Bibliothek und bleibt grün (benannt, äquivalent); `ClientAuth: RequireAnyClientCert` (rot, jeder Handshake) | andere |
| A12 | ALPN nennt allein `http/1.1`: ein Client mit `h2` und `http/1.1` und einer ohne ALPN verhandeln `http/1.1` bzw. nichts, ein Client allein mit `h2` scheitert (Nr. 6) | `TestLadeTLSALPN`: Handshakes im Prozess mit drei Clients; das Feld `NextProtos` ist genau `http/1.1` | `h2` in `NextProtos` (der erste Client verhandelt `h2`) | andere |
| A13 | Die Kette der Datei wird gesendet: ein Client mit nur der Wurzel im Pool vertraut Blatt und Zwischenzertifikat (Nr. 5) | `TestLadeTLSKette`: Datei mit Blatt und Zwischenzertifikat, Pool nur die Wurzel, Handshake im Prozess gelingt | das Paar aus dem ersten Zertifikat allein gebildet: das Zwischenzertifikat fehlt, der Client lehnt ab | andere |
| A14 | Kein Geheimnis in der Meldung: jeder Grund ist einer der sechs Texte, ohne Pfad, ohne Bytes des Inhalts, ohne Text der Bibliothek (Nr. 2, 11) | `TestLadeTLSKeinGeheimnis`: Schleife über jede Meldung der Mengen E, L, Z, S, P, G (33 Elemente) mit Pfad `…/geheim-pfad-4711/…`, einem Subject `geheim-cn-4711` und dem Base64 des Schlüssels: keiner dieser Texte, kein `tls:` und kein `x509:` kommt vor; Zählprüfung: die Tabelle von Nr. 2 hat sechs Gründe, der Quelltext von `tlsdatei.go` fünf feste Literale (`go/parser`) | `Grund: err.Error()` oder der Pfad im Grund | andere |
| A15 | `ladeTLS` ändert keine Datei und hält keinen Zustand: dieselben Dateien und Zeit, dasselbe Ergebnis (Nr. 12) | `TestLadeTLSZustandslos`: Bytes und Änderungszeit der beiden Dateien vor und nach dem Aufruf, bei Erfolg und bei Verstoß (Menge: ein Erfolg, ein Verstoß je Rang 2 bis 4 des Zertifikats, 4); zweiter Aufruf gleiches Ergebnis | der Baustein schreibt die Datei zurück (`os.WriteFile`); ein Paketzustand merkt den letzten Fehler | andere |
| A16 | Gleichzeitige Aufrufe teilen nichts (Nr. 12) | `TestLadeTLSGleichzeitig`: 16 Goroutinen rufen mit denselben Dateien, alle Ergebnisse gleich; `make test-race` ist der Fänger der Mutation, je Prozess (`-test.count=1`) | eine Paketvariable als Zwischenspeicher des letzten `*tls.Config`: Datenrennen, der Detektor meldet es in jedem Prozess (Rate je Prozess trägt der Implementer ein) | nebenläufig |
| A17 | Der Baustein ist für einen zweiten Aufrufer ohne Änderung nutzbar: die Namen in der Meldung sind die des Aufrufers (Nr. 2, Kopf) | `TestLadeTLSZweiteVerwendung`, Menge N (2 Namenspaare: `LISTEN_TLS_CERT`/`LISTEN_TLS_KEY` und `ADMIN_TLS_CERT`/`ADMIN_TLS_KEY`) × die Mengen E und L und das Paar „Gültigkeit und passt nicht“ aus A08 (Schleife); Zählprüfung 2 × 11 | ein fester Name `LISTEN_TLS_CERT` im Rang 1 oder im Schlüssel eines Verstoßes | andere |
Zeilen dieses Teils: 17 (A01 bis A17): *a* = 1 (A16), *b* = 16. Die Zeilen B01 bis B22 (*a* = 1, *b* = 21) stehen im Plan von Teil 2 ([slice-listener-tls-verdrahtung](../open/slice-listener-tls-verdrahtung.md)); ungeschnitten waren es 39 Zeilen mit *a* = 2 und *b* = 37. Die Zeile B22 mutiert den Text des Grundes in `tlsdatei.go` aus diesem Teil; ihr Fänger ist die Ausgabe von `make handbuch-check` in Teil 2.

**Mengen und Zählprüfung gegen die Quelle** (je Menge zählt der Test die Elemente und vergleicht mit der Zahl hier; die Quelle steht dahinter; die Mengen V, Modi, R, W, H, B, A, H4, H' und T führt Teil 2): E 4 (`SPEC-291` Nr. 2, zwei Schlüssel × gesetzte Datei gültig oder fehlend) · L 6 (Nr. 2 Rang 2: drei Formen × zwei Schlüssel) · Z 7 und S 4 (Nr. 2 und 4, Rang 3) · P 6 (Rang 4 des Schlüssels) · G 6 (Nr. 3) · Q 25 (5 × 5 Zustände von Nr. 2) · F 8 und K 5 (Nr. 4: die zugesagten Formen der Datei und die fünf Schlüsselformen) · N 2 (Nr. 2 Kopf). Die Gründe: sechs Zeilen in der Tabelle von Nr. 2, fünf feste Literale im Quelltext (A14). Achsen der Mengen: die Stelle des Elements im Wert (Block vor oder nach Fremdtext, erstes Zertifikat der Kette), der Schlüssel (Zertifikat oder Schlüssel), die Grenze (beide Seiten und gleich: G) und der Weg zum Aufruf (Laden im Baustein; `Build` und `Hochfahren` in Teil 2).

*Mutationen, Raten und Messende trägt der Implementer in einen Absatz unter dieser Tabelle ein (`AGENTS.md` §3.10); die Rate von A16 mit Zahl der Prozesse und Last des Hosts.*

**Betriebsregeln für alle Rollen:**

- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1); Mutationen nacheinander, `make gates` nie parallel zu Mutationsläufen.
- Auf Repo-Dateien nur `Edit`/`Write`, kein `sed -i`, Heredoc, `awk` oder Python; keine Hilfsdateien im Wurzelverzeichnis.
- Fixtures mit einer Folge aus Backslash, `u` und vier Hex-Ziffern, BOM oder NBSP entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei. PEM-Fixtures mit Zeilenenden, Leerraum und Fremdtext vor und nach dem Block entstehen aus Bytewerten.
- Nennt die Tabelle eine Klasse generierbar (hier: jede Fehlform von Zertifikat und Schlüssel, jede Kombination der beiden Schlüssel gesetzt/leer/fehlend), führt schon der Handoff den Generator oder die Schleife mit Orakel und Zählprüfung (`implement-slice` Schritt 14).
- Zu jedem neuen Wächter die rot färbende Mutation benennen, bei einem Befund die Menge aus dem Wortlaut der Zusage bilden (`implement-slice` Schritt 19, `AGENTS.md` §3.17).
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); der Stempel der Gegenproben (`harness/gegenproben.stempel`) wird nur im Integrationsschritt gesetzt.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-http-adapter`, `slice-bootstrap-m2` und `slice-handbuch-meldungskatalog` liegen in `done/`; keiner der drei ist eine Code-Abhängigkeit dieses Teils (der Baustein braucht nur die Standardbibliothek; sie sind Bedingung von Teil 2). Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): `SPEC-291` und der Nachzug von `SPEC-110` liegen vor (Architect, 2026-11-21); der Schnitt steht unten (Planner, 2026-11-21).

**Stand der Bedingungen beim Start** (Planner, 2026-11-21, `main` = `ce593554` vor den Moves): `slice-http-adapter`, `slice-bootstrap-m2` und `slice-handbuch-meldungskatalog` liegen in `done/` (geprüft an den Dateien; für Teil 1 keine Code-Abhängigkeit); `SPEC-291` samt Tabelle, Raster und Mengen und der Nachzug von `SPEC-110`, `SPEC-139` und `SPEC-136` stehen auf `main`; die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht (kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`; `tlstest` liegt unter `src/` und ist kein Gate); in `in-progress/` laufen sonst `slice-admin-abweisen-audit` (`src/adapters/driving/admin/`: kein Konflikt) und `slice-telemetry-einbau-mock` (`wurzel.go`, `start.go`, Tests in `src/bootstrap/`; dieser Teil legt dort allein neue Dateien an: kein Konflikt); drei Worktrees. Die Sendung an `slice-admin-tls` (der Baustein, `ladeTLS` mit `ADMIN_TLS_*`) ist geprüft (`AGENTS.md` §3.13): `slice-admin-tls` §1 *Übernimmt* trägt sie, kein Ausschluss trifft sie, die Bedingung dort nennt diesen Slice (im Commit des Schnitts nachgezogen). Teil 2 (`slice-listener-tls-verdrahtung`, Plan in `open/`) startet nach diesem Teil und nach `slice-telemetry-einbau-mock`.

**Abhängigkeiten — Grund je Eintrag:**

- Keine Code-Abhängigkeit: `ladeTLS` und `tlstest` brauchen nur die Standardbibliothek und keinen Zustand des Roots; die Tests führen Handshakes im Prozess über `net.Pipe`, ohne Listener des HTTP-Adapters.
- `slice-http-adapter` (`welle-record-replay`), `slice-bootstrap-m2` (`welle-record-replay`) und `slice-handbuch-meldungskatalog` (`welle-benutzerhandbuch`) sind Bedingung von [slice-listener-tls-verdrahtung](../open/slice-listener-tls-verdrahtung.md) (Teil 2: der Listener, der Composition Root mit `LadeOderMelde`, das Handbuch mit Kataloge-Gate); sie liegen beim Start dieses Teils in `done/`.

**Unabhängig von `welle-mock` / M3, weil** der Baustein weder `MockUseCase` noch den Modus `mock` noch einen Listener braucht. `slice-container-image`, die Telemetrie und `slice-lasttest` sind nicht Bedingung.

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-21, Teil des Planner 2026-11-21, vor dem ersten Code-Commit): der Block unten trägt für diesen Teil 17 Zeilen der Tabelle, *a* = 1 (A16), *b* = 16, *f* = 0, *p* = 100: *S* = 85·1 + 80·16 + 200·0 + 100 = 1465 gleich der Tabelle, *M* = 100 · ⌈135·1465 / 10000⌉ = 100 · ⌈19,78⌉ = 2000 gleich der Formel; keine Schnittfrage (135·*S* = 197775 ≤ 300000, *S* < 2223), kein `schnitt-grund`. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks (`src/bootstrap/tlsdatei.go`, `src/bootstrap/tlsdatei_test.go`, `src/hexagon/ports/support/tlstest/`), Tests eingeschlossen. Der Architect überlässt das Lesen der PEM der Standardbibliothek (`SPEC-291` Nr. 4), der Zuschlag für einen eigenen Leser fremder Bytes entfällt.

**Schnitt vor dem Code (Vorschlag des Architect, 2026-11-21, `AGENTS.md` §3.14, `SPEC-128` Nr. 1; vom Planner nachgerechnet und geschnitten, 2026-11-21):** die Tabelle in §3 hatte 39 Zeilen mit *a* = 2 (`A16`, `B15`), *b* = 37, *f* = 1 (der TLS-Test-Server von `B13` bis `B18`, mit Haltepunkt im Handshake), *p* = 180 (`tlsdatei.go` rund 100, `konfig_tls.go` rund 40, die Änderungen in `konfig.go`, `konfig_basis.go`, `konfig_werte.go`, `wurzel.go` und `lebenszyklus.go` rund 40): *S* = 85·2 + 80·37 + 200·1 + 180 = 3510, *M* = 4800, Schnittfrage (*S* ≥ 2223). Der Teil lässt sich ohne geteilten Mechanismus abtrennen, weil der Baustein nur die Standardbibliothek und keinen Zustand des Roots braucht und der Rest ihn nur aufruft:

| Teil | Zeilen der Tabelle | *a* | *b* | *f* | *p* | *S* | *M* | Inhalt |
|---|---|---|---|---|---|---|---|---|
| 1 (Baustein) | `A01` bis `A17` | 1 | 16 | 0 | 100 | 1465 | 2000 | `src/bootstrap/tlsdatei.go`, `src/hexagon/ports/support/tlstest/`, ihre Tests; kein Eintrag in Konfiguration und Start, nichts am Binary |
| 2 (Konfiguration und Verdrahtung) | `B01` bis `B22` | 1 | 21 | 1 | 80 | 2045 | 2800 | `konfig_tls.go`, die Änderungen in `konfig*.go`, `wurzel.go`, `lebenszyklus.go`, die Tests des Adapters mit TLS, Handbuch, Katalog |

Der Teil 2 liegt mit *S* = 2045 unter der Schnittgrenze (178 Zeilen Abstand); jede weitere Zeile der Tabelle von Teil 2 kostet 80, drei weitere lösen die Schnittfrage neu aus. Teil 1 liefert nichts am Binary (`ladeTLS` ruft nur der Test; `unused` zählt Tests mit, `make lint` bleibt grün) und ist der Teil, den `slice-admin-tls` als Bedingung braucht (der Planner trägt die Kennung dort ein, `AGENTS.md` §3.13); Teil 2 hängt an Teil 1 (*Code-Abhängigkeit*). Gegen einen Schnitt „Konfiguration ohne Verdrahtung“ sprach, dass der Start `LISTEN_TLS_*` dann annähme und der Listener bei HTTP bliebe, eine Sicherheitseinstellung ohne Wirkung (`SPEC-110` Nr. 2); der Schnitt liegt deshalb vor der Annahme der Schlüssel. Geschnitten wird der Plan, nicht die Spezifikation: `SPEC-291` bleibt eine Nummer. Den Block unten ersetzt der Planner beim Schnitt je Teil durch den Block des Teils.

**Geprüft und geschnitten vom Planner (2026-11-21):** Die Zahlen des Architect sind aus der Tabelle nach `SPEC-128` Nr. 1 nachgerechnet: ungeschnitten *S* = 85·2 + 80·37 + 200·1 + 180 = 3510, 135·3510 = 473850, *M* = 4800; Teil 1 (A01 bis A17) *S* = 85·1 + 80·16 + 200·0 + 100 = 1465, 135·1465 = 197775, *M* = 100 · ⌈19,78⌉ = 2000; Teil 2 (B01 bis B22) *S* = 85·1 + 80·21 + 200·1 + 80 = 2045, *M* = 2800. Beide Teile liegen unter der Schnittgrenze (*S* < 2223), die Naht teilt nur den Aufruf von `ladeTLS`. Teil 1 liegt 758 darunter, Teil 2 178 (drei weitere Zeilen lösen die Schnittfrage neu aus, siehe oben). Die Pfade des Blocks von Teil 1 weichen vom Vorschlag des Architect (`src/bootstrap/`) ab: sie nennen die zwei Dateien `src/bootstrap/tlsdatei.go` und `src/bootstrap/tlsdatei_test.go` und das Verzeichnis `src/hexagon/ports/support/tlstest/`, Grund: die Messung zählt jede Änderung an den Pfaden zwischen `basis` und `HEAD` (`SPEC-128` Nr. 10 g), und `slice-telemetry-einbau-mock` ändert zugleich andere Dateien unter `src/bootstrap/`; ein Pfad auf das Verzeichnis zählte sie nach einem `git merge main` in diesen Teil mit. Braucht der Teil eine weitere Testdatei, trägt der Implementer ihren Pfad in den Block nach und nennt es im Bericht.

**Reihenfolge und Konflikte** (Planner, 2026-11-21): (1) Dieser Teil legt in `src/bootstrap/` nur neue Dateien an (`tlsdatei.go`, `tlsdatei_test.go`) und das Verzeichnis `tlstest/`; er läuft zugleich mit `slice-telemetry-einbau-mock` (ändert `wurzel.go`, `start.go`): keine Datei gemeinsam. (2) `slice-admin-abweisen-audit` ändert `src/adapters/driving/admin/`, `support/audit.go` und `src/bootstrap/audit_betrieb_test.go`: kein Konflikt. (3) `slice-listener-tls-verdrahtung` startet, wenn dieser Teil und `slice-telemetry-einbau-mock` in `done/` liegen (`wurzel.go`, `welle-betrieb` §5: nacheinander). (4) `slice-admin-tls` (`welle-admin-api`) nennt diesen Slice als Bedingung. (5) Die Gate-Kette (Nr. 7 der Übergangsregeln) berührt der Plan nicht: kein Pfad von `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md` oder `tools/gegenproben/`; `tlstest` liegt unter `src/` und ist kein Gate.

```groessenmarke
pfad: src/bootstrap/tlsdatei.go
pfad: src/bootstrap/tlsdatei_test.go
pfad: src/hexagon/ports/support/tlstest/
produktcode: 100
fakes: 0
basis: de7589e5
schaetzung: 1465
marke: 2000
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; der Schnitt vor dem Code liegt vor (siehe *Schnitt vor dem Code* oben, die Naht ist Baustein gegen Konfiguration und Verdrahtung, `AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): der Baustein lässt sich nur mit einem Import eines Driving Adapters im anderen teilen oder ohne Kenntnis eines Schlüsselnamens des Aufrufers nicht schreiben (siehe §6).

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

- Der Baustein soll für den Admin-Listener ohne Änderung taugen; ohne Zuschnitt (Parameter statt Konstanten, Ort neben beiden Adaptern) entsteht ein Import eines Driving Adapters im anderen, den `make a-check` abweist — **Ausgang:** entschieden (Architect, 2026-11-21): der Baustein liegt im Composition Root, kein Adapter importiert ihn (`SPEC-291` Nr. 1); Fänger `A17` und `make a-check`.
- Sendung an den Nehmer `slice-admin-tls`: `ADMIN_TLS_*` und die Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` gehören `welle-admin-api`; der Nehmer ruft den Baustein dieses Teils — **Ausgang:** entschieden (Planner, 2026-11-21, `AGENTS.md` §3.13): die Datei `slice-admin-tls` liegt in `open/`; ihr §1 *Übernimmt* trägt den Baustein, ihr §1 *Ausdrücklich NICHT* nennt allein den OData-Listener (`LISTEN_TLS_*`, `slice-listener-tls-verdrahtung`) und trifft die Sendung nicht; ihre Bedingung nennt `slice-listener-tls` (diesen Teil, nicht die Verdrahtung) seit dem Commit dieses Schnitts, ihr §6 trägt die Zeile mit dem Ausgang *entschieden*.
- Start mit einer Schlüsseldatei, die Gruppen- oder Weltrechte trägt: die Spezifikation verlangt es für die Token-Datei, nicht für den TLS-Schlüssel — **Ausgang:** entschieden (Architect, 2026-11-21): der Start prüft die Rechte der Schlüsseldatei nicht (`SPEC-291` Nr. 10, Alternative, Grund und Folge dort); Fänger `A09` (Element acht, Rechte `0666`). Der User hat Option A (keine Prüfung) am 2026-11-21 bestätigt (AskUserQuestion); die Alternative, eine Prüfung wie bei der Token-Datei, ist damit verworfen.
- Geheimnis in Fixtures: ein privater Schlüssel im Repo ist ein Befund gegen `LH-RB-04` — **Ausgang:** entschieden (Auftrag 6): die Fixtures entstehen zur Laufzeit (`tlstest`, `t.TempDir()`), keine Datei mit Schlüssel im Repo; das Beispiel des Handbuchs (`B22`) braucht keine.
- HTTP/2 über TLS ändert das Verhalten von Strömen und Abbruch gegenüber dem Listener ohne TLS — **Ausgang:** entschieden (Architect, 2026-11-21): der Listener bietet HTTP/2 nicht an, ALPN allein `http/1.1` (`SPEC-291` Nr. 6); Fänger `A12` (die drei Clients) und `B11` (`ProtoMajor` 1 bei einem Client, der HTTP/2 anbietet).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-21): das Raster unten trägt in jeder Zelle einen Ort oder `gilt nicht` mit Grund; eine Zelle, die der Implementer ohne Ort findet, geht an den Architect zurück (`AGENTS.md` §3.12).
- Schnitt und Größenmarke: die Schätzung liegt über der Schnittgrenze (*S* = 3510) — **Ausgang:** entschieden (Planner, 2026-11-21): vor dem Code geschnitten in diesen Teil (*S* = 1465, *M* = 2000) und `slice-listener-tls-verdrahtung` (*S* = 2045, *M* = 2800); die Kennung dieses Teils steht in `slice-admin-tls` als Bedingung, `slice-container-image` und `slice-bootstrap-m6` nennen für `LISTEN_TLS_*` den Teil 2 (`AGENTS.md` §3.13), `welle-betrieb` führt beide; den Block mit `basis`, `schaetzung` und `marke` setzt der Planner beim Start je Teil, den dieses Teils in einem Commit vor dem ersten Code-Commit (geschehen am 2026-11-21: Block in §4, `basis` `de7589e5`, `schaetzung` 1465, `marke` 2000, `make groessenmarke` grün); den Block von Teil 2 setzt der Planner bei dessen Start (Ausgang im Plan von Teil 2).
- Handbuch und Kataloge-Gate: das Binary ändert sich in diesem Teil nicht (`ladeTLS` ruft nur der Test) — **Ausgang:** entschieden (Planner, 2026-11-21): der Handbuch-Punkt und die Schlüssel `LISTEN_TLS_*` im Katalog gehören zu Teil 2; §7 dieses Teils sagt, dass das Binary unverändert blieb.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund):

| Operation | fehlender Wert (nur einer der beiden Schlüssel) | leerer Wert | unbekannter oder fremder Wert (Datei fehlt, kein PEM, Schlüssel passt nicht, abgelaufen) | Abbruch (Herunterfahren während des Handshakes) | Wiederholung (zweiter Start, zweiter Handshake) | Modus (`record`, `replay`, `mock`, `hybrid`) |
|---|---|---|---|---|---|---|
| Zertifikat und Schlüssel laden (`ladeTLS`, `Build`) | `SPEC-291` Nr. 2 (beide leer: kein TLS; einer leer: Rang 1) | `SPEC-291` Nr. 2 Rang 3 (leere Datei) und Rang 1 (leerer Pfad) | `SPEC-291` Nr. 2 Ränge 2 bis 4, Nr. 3, 4 | gilt nicht: `ladeTLS` läuft synchron ohne Kontext und hält keinen Zustand (`SPEC-291` Nr. 12); `Build` liest die Dateien ohne Kontext, wie `UPSTREAM_CA_FILE` (`SPEC-139` Nr. 2) | `SPEC-291` Nr. 8 (zweimal gelesen, `Build` gilt) und Nr. 12 | `SPEC-291` Nr. 8 (`Build` lädt in jedem Modus) |
| Handshake im Prozess mit dem Ergebnis von `ladeTLS` (`tls.Server` über `net.Pipe`) | `SPEC-291` Nr. 5 (ein Client ohne Zertifikat gelingt) und Nr. 6 (ein Client ohne ALPN verhandelt nichts) | gilt nicht: ein Handshake ohne Byte hält keine Verbindung des Bausteins (der Baustein öffnet keinen Listener); der Fall ist die Zeile *Handshake des Clients* in Teil 2 (`SPEC-291` Nr. 9 a und c) | `SPEC-291` Nr. 5 und 6 (Version unter 1.2, nur `h2`: der Handshake scheitert); `SPEC-291` Nr. 9 b (unbekannte CA, falscher Name) liegt am Listener in Teil 2 | gilt nicht: der Baustein hält weder Verbindung noch Zustand (`SPEC-291` Nr. 12); der Abbruch beim Herunterfahren ist `SPEC-291` Nr. 9 c in Teil 2 | `SPEC-291` Nr. 12 (dasselbe Ergebnis bei gleichen Dateien und gleicher Zeit) | gilt nicht: der Baustein kennt keinen Modus (`SPEC-291` Kopf); die Zeile *Handshake des Clients* in Teil 2 trägt Nr. 7 e |
| Zweite Verwendung des Bausteins mit anderen Schlüsselnamen | `SPEC-291` Nr. 2 Rang 1 (der Name des anderen Schlüssels ist der des Aufrufers) | `SPEC-291` Nr. 2 Rang 1 (leerer Pfad) | `SPEC-291` Nr. 2 Ränge 2 bis 4 (der Schlüssel der Meldung ist der Name des Aufrufers, Nr. 7 c für `LISTEN_`) | gilt nicht: derselbe Grund wie beim Laden (kein Kontext, `SPEC-291` Nr. 12) | `SPEC-291` Nr. 12 (zustandslos: zwei Aufrufer mit zwei Namenspaaren ändern einander nicht) | gilt nicht: der Baustein kennt keinen Modus; die Modus-Pflichten des Admin-Listeners (`SPEC-089`, `SPEC-090`) liegen beim Aufrufer (`SPEC-291` Kopf) |

**Zusatztabelle** (`AGENTS.md` §3.16: die Operation meldet mehrere Verstöße und hängt vom Modus ab): je Paar von Verstößen und je Modus eine Zeile.

| Paar oder Modus | Entschieden in | Fänger |
|---|---|---|
| zwei Verstöße am selben Schlüssel (Rang 1 und 2, 2 und 3, 3 und 4): schließen einander aus, jeder Rang setzt das Bestehen der davor voraus; die Meldung trägt den kleinsten Rang | `SPEC-291` Nr. 2 | `A08` (Menge Q, Orakel mit einem Grund je Schlüssel); `A03` bis `A07` je Element |
| zwei Verstöße an verschiedenen Schlüsseln (Zertifikat außerhalb der Gültigkeit und Schlüssel passt nicht; kein PEM-Zertifikat und kein PEM-Schlüssel; nicht lesbar und nicht lesbar; …): beide Zeilen, nach Schlüssel sortiert; der Rang 4 des Schlüssels läuft nur, wenn beide Rang 3 bestanden | `SPEC-291` Nr. 2, `SPEC-110` Nr. 12 | `A08` (25 Paare); `B02` (ein Paar über `LadeOderMelde`) und `B10` (ein Paar in `Build`) in Teil 2 |
| ein Verstoß am Zertifikat (Rang 4) und ein Zertifikat, dessen Zwischenzertifikat abgelaufen ist: allein das Blatt zählt | `SPEC-291` Nr. 3 | `A07` (Element sechs) |
| Modus `record`, `replay`, `mock`, `hybrid` und ungültiger Wert von `MODE` | gilt nicht: der Baustein kennt keinen Modus (`SPEC-291` Kopf); die Zeilen je Modus stehen in der Zusatztabelle von Teil 2 | Teil 2: `B03`, `B06`, `B07`, `B10`, `B11`, `B18` |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30), Treffer und Zähler beim Start nachgezogen (Planner, 2026-11-21; der Zähler ist die Zahl der Dateien unter `evidence/`). Der Slice liest fremde Bytes (PEM-Dateien des Betreibers): `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14, `.claude/agents/implementer.md` und `.harness/skills/reviewer.md`) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (56×, verkörpert in `AGENTS.md` §3.17) gelten für „jede Fehlform“ von Zertifikat und Schlüssel; `BEO-REPO/randform-erst-im-review-entschieden` (37×, verkörpert in `AGENTS.md` §3.16) für das Raster; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14 und `.harness/skills/reviewer.md`) für Testhilfen, die Verbindungen und Ströme umhüllen, und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (21×, verkörpert in den Übergangsregeln Nr. 8 und im Command-Guard). Keine Beobachtung steht mit diesem Slice offen unter 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
