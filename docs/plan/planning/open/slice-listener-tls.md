# Slice slice-listener-tls: TLS des OData-Listeners (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`) als wiederverwendbarer Baustein

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

**Berührte Spec-Stellen:** `LH-RB-02.a` (HTTP/TLS-Konfiguration, Nachweis: Integrationstests für HTTP, HTTPS und Zertifikatsfehler) · `SPEC-088` (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`, Zeile geschärft) · `SPEC-291` (neu, Architect, 2026-11-21: Baustein `ladeTLS`, Gründe, Gültigkeit, ALPN, Start, Handshake) · `SPEC-089` (Admin, Gegenstück) · `SPEC-110` Nr. 2, 7, 11, 13, 14 c und e (Nachzug, Architect) · `SPEC-055` (Zeile: Schema `https` der abgeleiteten Basis) · `SPEC-139` Nr. 2, 6, 8 c (`Build` prüft, `Hochfahren` umhüllt) · `SPEC-136` *Ungeprüft* (der Listener ist nicht mehr Klartext) · `SPEC-059` (Herunterfahren) · `SPEC-114` (Uhr) · `ARC-005`, `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der OData-Listener spricht HTTPS, wenn `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` beide gesetzt sind, und HTTP, wenn keiner gesetzt ist; ein Start mit genau einem der beiden Schlüssel oder mit unlesbarem, unpassendem oder abgelaufenem Zertifikat scheitert mit `ODR-E2001` vor dem Öffnen des Listeners; die Startkonfiguration nimmt `LISTEN_TLS_*` an (der Nachzug von `SPEC-110` Nr. 2 durch den Architect). Die TLS-Konfiguration steckt in einem Baustein, den der Admin-Listener (`welle-admin-api`) mit `ADMIN_TLS_*` wiederverwendet, ohne ihn zu ändern. Der HTTPS-Test am Listener zeigt: gültiges Zertifikat wird bedient, ein Client mit unbekannter CA lehnt ab (`LH-RB-02.a`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- TLS des Admin-Listeners, `ADMIN_TLS_*`, `ADMIN_ALLOW_PLAINTEXT` und die Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` — `welle-admin-api` (`slice-admin-tls` der Tabelle dort); dieser Slice liefert nur den Baustein. Der Nehmer hat noch keine Datei (siehe §6).
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

- [ ] Baustein und Listener: der Listener bedient HTTPS mit dem Zertifikat aus `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY`; ein Client mit unbekannter CA lehnt den Handshake ab, ein Client mit der CA erhält die Antwort; ohne beide Schlüssel bleibt es HTTP (`LH-RB-02`, `LH-RB-02.a`).
- [ ] Startkonfiguration: genau einer der beiden Schlüssel (ein leerer Wert gilt als ungesetzt), eine unlesbare Datei, eine Datei ohne PEM-Zertifikat oder ohne PEM-Schlüssel, ein Schlüssel ohne passendes Zertifikat, ein abgelaufenes oder noch nicht gültiges Zertifikat endet mit `ODR-E2001` und Meldung nach `SPEC-110` Nr. 12 und `SPEC-291` Nr. 2, bevor `DATA_DIR` angelegt wird und ein Listener öffnet; `SPEC-110` Nr. 2 lehnt `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` nicht mehr ab, jeder andere Schlüssel `LISTEN_TLS_*` und `ADMIN_*` bleibt abgelehnt (Nachzug des Architect liegt in der Spezifikation).
- [ ] Der Baustein `ladeTLS` ist ohne Änderung für den Admin-Listener nutzbar (eigene Schlüsselnamen als Parameter, kein Bezug auf `LISTEN_`); ein Test belegt die zweite Verwendung mit anderen Schlüsselnamen (`ADMIN_TLS_CERT`, `ADMIN_TLS_KEY`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Die Schlüssel `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` kommen in den Katalog der Konfigurationsschlüssel, ein Start mit nur einem trägt `ODR-E2001` (kein neuer Code: `SPEC-109` Nr. 2 und §4 bleiben). Die Stellen, die das Kataloge-Gate liest, nennt §3 *Hinweise an den Planner*.
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
| `src/bootstrap/konfig_tls.go` | neu | liest `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY`, ruft `ladeTLS`, lehnt jeden anderen `LISTEN_TLS_*` ab; eine Datei `konfig*.go`, weil das Kataloge-Gate nur diese liest |
| `src/bootstrap/` (`konfig.go`, `konfig_basis.go`, `konfig_werte.go`, `wurzel.go`, `lebenszyklus.go`) | update | zwei Felder von `Konfiguration`, Schema der abgeleiteten Basis, `LISTEN_TLS_` aus `nichtGelieferte`, Prüfung in `vorbereiten`, `tls.NewListener` in `starteListener` |
| `src/adapters/driving/httpodata/` | nur Tests | der Adapter bleibt unverändert (`SPEC-291` Nr. 1, 9); Tests mit einem TLS-Listener belegen seine Zusagen mit TLS |
| `src/hexagon/ports/support/tlstest/` | neu | Hilfe für Tests: Aussteller und Blätter zur Laufzeit (Hierarchie, Ablauf, Namen, Schlüsselformen); nur Standardbibliothek, kein Produktcode |
| `spec/spezifikation.md` | update | geschehen (Architect): `SPEC-291`, `SPEC-110`, `SPEC-139`, `SPEC-136`, §3 |
| `.a-check.yml` | unverändert | der Baustein liegt im Composition Root, `tlstest` unter `ports/support` (nur Standardbibliothek); keine neue Kante |
| `docs/user/benutzerhandbuch.md` | update | Kapitel TLS und zwei Zeilen in der Tabelle der Einstellungen (Pflicht Handbuch) |
| Tests | neu | Zertifikate und Schlüssel als Fixtures zur Laufzeit erzeugt (kein Geheimnis im Repo), HTTPS-Test |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Beide liegen vor (Architect, 2026-11-21).

**Aufträge an den Architect — entschieden (2026-11-21, Ort in der Spezifikation):**

1. **Ort und Schnitt des Bausteins** — *entschieden*, `SPEC-291` Kopf und Nr. 1: Datei `src/bootstrap/tlsdatei.go` im Composition Root, nicht in einem Adapter; beide Listener bleiben unverändert und erhalten einen umhüllten `net.Listener` (`SPEC-136` Nr. 1, `SPEC-238` Nr. 1); kein Import eines Driving Adapters im anderen, `.a-check.yml` unverändert. Parameter: `tlsDateien{certSchluessel, keySchluessel, certPfad, keyPfad}` und die Zeit; Rückgabe `(*tls.Config, []verstoss)`.
2. **Prüfung beim Start** — *entschieden*, `SPEC-291` Nr. 2 bis 5 und 10: sechs Gründe in vier Rängen je Schlüssel (leer, nicht lesbar, kein PEM, Gültigkeit des Blatts bzw. Schlüssel passt nicht); der Leser ist die Standardbibliothek (`encoding/pem`, `x509.ParseCertificate`, `tls.X509KeyPair`), der Baustein schreibt keinen eigenen Parser; Mindestversion TLS 1.2, Suiten der Bibliothek, kein mTLS, kein Neuladen. **Rechte der Schlüsseldatei: nicht geprüft** (`SPEC-291` Nr. 10, Alternative und Folge dort). Größe der Dateien nicht begrenzt (wie `UPSTREAM_CA_FILE`, *Ungeprüft*). Meldung nie mit Pfad, Inhalt oder Text der Bibliothek (Nr. 11).
3. **Nachzug `SPEC-110`** — *entschieden*, `SPEC-110` Nr. 2, 7, 11, 13, 14 c und e: `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` sind geliefert, jeder andere `LISTEN_TLS_*` bleibt `ODR-E2001`; Reihenfolge: reine Prüfungen, dann die Dateien (vor `DATA_DIR`); `PUBLIC_BASE_URL` abgeleitet trägt `https`, wenn beide Schlüssel gesetzt sind, eine gesetzte wird nicht abgeglichen (`SPEC-291` Nr. 7 d); Grenze der Länge 2025 Bytes bei TLS.
4. **Protokolle** — *entschieden*, `SPEC-291` Nr. 5 und 6: ALPN nennt allein `http/1.1`; HTTP/2 wird nicht angeboten, weil die Zusagen von `SPEC-136` Nr. 4 h, 9, 13 und 14 an HTTP/1.1 gemessen sind; ein Client, der allein `h2` anbietet, scheitert im Handshake (Folge im Betrieb dort).
5. **Herunterfahren** — *entschieden*, `SPEC-291` Nr. 9 c und `SPEC-136` Nr. 13 bis 15 unverändert: eine Verbindung mit nicht beendetem Handshake hat keinen Handler und wird von `Schliessen` sofort geschlossen; ein Request, dessen Zulassung nach `Anhalten` liegt, erhält über TLS 503 `ODR-E5010`; ein Handshake, der scheitert, erreicht den Adapter nie (Nr. 9 b).
6. **Fixtures** — *entschieden*: Hilfspaket `src/hexagon/ports/support/tlstest` (nur Standardbibliothek, kein Produktcode, erreichbar für Adapter-Tests und Composition Root): Aussteller und Blatt zur Laufzeit mit wählbarem `NotBefore`, `NotAfter`, Namen, Schlüsselform und Kette; die Dateien entstehen unter `t.TempDir()`; PEM mit Zeilenende, Fremdtext und fehlendem Zeilenende aus Bytewerten (`AGENTS.md` §3.18). Keine Datei mit Schlüssel im Repo (`LH-RB-04`).
7. **Tabelle, Raster, Gegenproben, erwartete Produktcode-Größe** — *entschieden*: Tabelle und Raster unten, Block `groessenmarke` in §4 (`pfad`, `produktcode`, `fakes`); Gegenproben entfallen (kein neues Werkzeug des Harness).

**Hinweise an den Planner** (Architect, 2026-11-21):

- **Schnitt vor dem Code** (`SPEC-128` Nr. 1, Schnittfrage ab *S* ≥ 2223): die Tabelle unten trägt 39 Zeilen (*a* = 2, *b* = 37), *f* = 1, *p* = 180, also *S* = 85·2 + 80·37 + 200·1 + 180 = 3510 und Marke 4800, über der Schnittgrenze. Die Naht steht in §4 (Teil 1: Baustein, Zeilen `A01` bis `A17`; Teil 2: Konfiguration und Verdrahtung, Zeilen `B01` bis `B22`); die Nummern der Zeilen tragen den Teil.
- **Kataloge-Gate** (`SPEC-201` Nr. 4; der Plan muss die Stellen nennen, die der Slice berührt): (1) der Quelltext der Konfiguration trägt beide Schlüssel als Literal in einer Datei `konfig*.go` (`konfig_tls.go`); (2) `nichtGelieferte` in `konfig_werte.go` verliert `LISTEN_TLS_` (bleibt `ADMIN_`), sonst liegen beide Namen in der Menge der abgelehnten und fehlen im Handbuch; (3) die Zeile `2. **Schlüssel dieser Version.**` von `SPEC-110` nennt beide (geschehen); (4) die Tabelle `Einstellungen` des Handbuchs bekommt je eine Zeile; (5) die Ablehnung jedes anderen `LISTEN_TLS_*` darf das Präfix nicht als Literal in `konfig*.go` tragen, sonst zählt es als Schlüssel (Art 4); der Baustein kann es aus den beiden Schlüsselnamen bilden (alles bis einschließlich dem letzten `_`). **Fehlercodes: keine Änderung** (`ODR-E2001` besteht; Tabelle von §4, `SPEC-109` Nr. 2 und die Tabelle der Fehlercodes im Handbuch bleiben). **Kommandos: keine.** Im Handbuch zusätzlich: die Zeile `PUBLIC_BASE_URL` (Schema `https` bei TLS), ein Kapitel TLS (Dateien, beide oder keiner, kein Neuladen, kein HTTP/2, Rechte der Schlüsseldatei als Empfehlung `0400` oder `0600`, die Meldungen) mit einem Beispiel, das `make handbuch-check` gegen das gebaute Binary fährt (Zeile `B22`: nur `LISTEN_TLS_CERT` gesetzt, Exit 1, `ODR-E2001 LISTEN_TLS_KEY: erforderlich, wenn LISTEN_TLS_CERT gesetzt ist`; ein Beispiel mit gültigem Paar braucht eine Datei mit Schlüssel im Repo und entfällt, `LH-RB-04`).
- **Nehmer `slice-admin-tls`:** ruft `ladeTLS` mit `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` (die Zeile `A17` belegt das mit diesen Namen) und umhüllt den Listener des Admin-Adapters wie `Hochfahren` den OData-Listener; die Pflicht bei nicht-loopback `ADMIN_LISTEN` und `ADMIN_ALLOW_PLAINTEXT` liegen nicht im Baustein (`SPEC-291` Kopf). Bei einem Schnitt trägt der Teil 1 die Kennung, die `slice-admin-tls` als Bedingung nennt.
- **Leser fremder Bytes** (§4 *Größenmarke*): der Architect überlässt das Lesen der PEM der Standardbibliothek (`SPEC-291` Nr. 4); der Zuschlag für Leser, den `AGENTS.md` §3.14 für eigene Prüfung nennt, entfällt, und die Formel von `SPEC-128` Nr. 1 kennt ihn nicht.
- **Spec-Nummer:** `SPEC-291` ist vergeben (die Nummer `SPEC-290` gehört dem Mock-Vertrag der Telemetrie, der zeitgleich auf `main` kam); die nächste freie ist `SPEC-292`.
- **Zweite Änderung an `SPEC-139`:** die Telemetrie des Mock (`SPEC-290`) hat `SPEC-139` Nr. 13 angelegt und Nr. 3 und 11 angepasst; dieser Slice ändert Nr. 2, 6 und 8 c und berührt Nr. 13 nicht. Beide Slices ändern `src/bootstrap/` (`wurzel.go`, `start.go`); wer später startet, zieht `main` vor dem ersten Code-Commit nach.

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
| B01 | Die Konfiguration trägt die Pfade (Leerraum am Rand entfernt, leer oder nur Leerraum ist ungesetzt, ein NBSP am Rand bleibt), nie den Inhalt der Dateien (Nr. 7 a, `SPEC-110` Nr. 1) | `TestKonfigListenTLSWerte`, Menge W (6: Pfad, Pfad mit Leerzeichen, Tab, LF und CR am Rand, NBSP am Rand (`Datei nicht lesbar`), leer, nur Leerraum, relativer Pfad); die Ausgabe `%+v` der Konfiguration trägt den Pfad, nicht das Base64 des Schlüssels | `kappe` entfällt für beide Schlüssel; die Konfiguration trägt das Paar | andere |
| B02 | Ein Verstoß des Ladens ist eine Zeile `ODR-E2001 <Schlüssel>: <Grund>` mit dem Namen aus `SPEC-110` Nr. 2, sortiert, Exit 1, ohne Pfad und Inhalt; `DATA_DIR` wird nicht angelegt (Nr. 7 b, c, `SPEC-110` Nr. 11, 12) | `TestKonfigListenTLSVerstoesse`, Menge V (8: Rang 1 bis 4 je Schlüssel) über `LadeOderMelde`; ein Paar (Zertifikat Rang 2, Schlüssel Rang 4 nach A08) trägt zwei Zeilen in der Folge `LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`; `DATA_DIR` im Temp-Verzeichnis besteht danach nicht | die Prüfung steht hinter `legeDatenverzeichnisAn`; der Name des Aufrufers statt des Schlüssels von `SPEC-110` | andere |
| B03 | Dieselben Verstöße in jedem Modus, auch bei ungültigem `MODE` (Nr. 7 a, e) | `TestKonfigListenTLSModi`, Menge Modi (5: `record`, `replay`, `mock`, `hybrid`, ungültiger Wert) × zwei Verstöße (Zertifikat Rang 3, Schlüssel Rang 1): die Zeilen von TLS sind in jedem gleich | die Prüfung läuft nur bei gültigem Modus | andere |
| B04 | Jeder andere Schlüssel `LISTEN_TLS_*` bleibt `ODR-E2001` („nicht unterstützt“), die Schreibweise zählt, ein leerer Wert zählt nicht (`SPEC-110` Nr. 2) | `TestKonfigListenTLSRest`, Menge R (8: abgelehnt `LISTEN_TLS_X`, `LISTEN_TLS_CLIENT_CA`, `LISTEN_TLS_`, `LISTEN_TLS_CERTS`, `LISTEN_TLS_KEY_FILE`; ignoriert mit leerem Wert, nur Leerraum, `listen_tls_x`); Zählprüfung 5 und 3 | die Ablehnung entfällt; sie vergleicht den Namen statt des Präfixes | andere |
| B05 | `ADMIN_*` bleibt `ODR-E2001`, in jedem Modus, auch `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` (`SPEC-110` Nr. 2) | die Zelle von `TestHybridKonfigFamilien` und `konfig_test.go` (Modus × Familie, 4 × 6 Namen: `ADMIN_LISTEN`, `ADMIN_TOKEN_FILE`, `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY`, `ADMIN_ALLOW_PLAINTEXT`, `ADMIN_X`); die Zeilen mit `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` wandern nach B01 und B02 | `ADMIN_` aus der Menge der abgelehnten Präfixe genommen | andere |
| B06 | Die abgeleitete `PUBLIC_BASE_URL` trägt `https`, wenn beide Schlüssel gesetzt sind (gleich ob die Dateien gültig sind), sonst `http`; in `record` und `replay` bleibt sie leer (Nr. 7 d, `SPEC-110` Nr. 14 c) | `TestKonfigBasisSchema`, Menge H (6 in `mock`: Hosts `127.0.0.1`, `[::1]`, `localhost` × TLS und kein TLS) und zwei Fälle mit TLS in `record` und `replay` (leer) | immer `http`; immer `https`; das Schema hängt an der Gültigkeit der Dateien | andere |
| B07 | Die Länge der Wurzel zählt das Schema mit: bei TLS ist ein Root, der die Basis auf 2048 Bytes bringt, gültig, einer mit einem Byte mehr nicht; ohne TLS liegt die Grenze ein Byte höher (Nr. 7 d, `SPEC-110` Nr. 14 e) | `TestKonfigBasisLaenge`, Menge B (4: TLS mit Root von 2025 und 2026 Bytes, kein TLS mit 2026 und 2027 Bytes, Host `localhost:65535`) | die Länge der Basis ohne Schema gezählt; die Grenze bei TLS gleich der ohne | andere |
| B08 | Eine gesetzte `PUBLIC_BASE_URL` wird nicht gegen TLS abgeglichen (Nr. 7 d) | `TestKonfigBasisGesetzt`, Menge A (2: TLS mit `http://h.example`, kein TLS mit `https://h.example`): kein Verstoß, Wert unverändert | ein Verstoß bei Widerspruch von Schema und TLS | andere |
| B09 | Das Laden prüft die Gültigkeit zur Zeit des Starts: dieselbe Umgebung ergibt mit früherer und späterer Zeit verschiedene Ergebnisse (Ausnahme von `SPEC-110` Nr. 13, Nr. 3) | `TestKonfigListenTLSZeit`: der interne Einstieg mit fester Zeit (vor, in, nach der Gültigkeit); `LadeKonfiguration` liest die Systemuhr (Fixture mit `NotAfter` eine Stunde zurück: Verstoß; eine Stunde voraus: gültig) | die Zeit des Aufrufs wird ignoriert (Nullzeit) | andere |
| B10 | `Build` lädt das Paar vor dem Öffnen der Ablage, in jedem Modus; ein Verstoß ist ein `*KonfigFehler` mit einer Zeile je Schlüssel, und es entsteht keine Datenbank und kein Blob-Verzeichnis (Nr. 8, `SPEC-139` Nr. 2) | `TestBuildTLSVerstoss`, Modi (4) × zwei Konfigurationen (Zertifikat Rang 2; Zertifikat Rang 4 und Schlüssel Rang 4 zusammen: zwei Zeilen); danach fehlen `odata-recorder.db` und `blobs` | der Aufruf steht hinter `Oeffne` | andere |
| B11 | `Hochfahren` umhüllt den Listener genau dann, wenn beide Pfade gesetzt sind; die Antwort kommt über HTTPS wie sonst, HTTP/1.1 auch bei einem Client, der HTTP/2 anbietet; ein zweiter `Build` auf denselben Dateien (Neustart) ist unabhängig (Nr. 8, 6, 12) | `TestHTTPSEndeZuEnde`, Modi (4): Client mit der Wurzel im Pool, `ForceAttemptHTTP2`, Anfrage an einen unbekannten Pfad: `404`, `ODR-E5005`, `ProtoMajor` 1; danach `Herunterfahren`, ein zweiter `Build` und dieselbe Anfrage; ohne Schlüssel: Klartext-Anfrage gelingt, ein TLS-Client scheitert | `tls.NewListener` entfällt (HTTPS-Client scheitert); die Umhüllung immer (der Klartextlauf scheitert); `h2` in `NextProtos` (`ProtoMajor` 2) | andere |
| B12 | Zertifikatsfehler sind am Client sichtbar und erreichen den Dienst nicht (`LH-RB-02.a`; Nr. 9 b) | `TestHTTPSClientFehler`, Menge H4 (4: Client ohne die CA im Pool → `x509.UnknownAuthorityError`; Blatt ohne den Namen des Hosts → `x509.HostnameError`; Klartext-Client am TLS-Port: Antwort ohne `ODR-`-Fehlerobjekt und ohne `X-Correlation-Id`; TLS-Client am Klartext-Port: Fehler des Handshakes) | die Umhüllung entfällt (die Fälle drei und vier werden rot); die Fixture des Falls zwei trägt einen Namen, der vom Host der Anfrage verschieden ist (sie gleicht dem Wert der Mutation nie, `AGENTS.md` §3.15) | andere |
| B13 | Der Adapter reicht kein Merkmal von TLS an `Verarbeiten`: dieselbe `inbound.Anfrage` und Antwort ohne und mit TLS (Nr. 9 d) | `TestAnfrageMitTLS`, Menge T (6: Transport {Klartext, TLS} × Fälle {GET mit Query und Header, POST mit Länge und mit chunked, Antwort chunked mit Flush je Stück}) im Test des Adapters mit einem Listener aus `tls.NewListener` und dem Aussteller aus `tlstest`; Vergleich aller Felder der `Anfrage` und der Antwort | der Handler setzt bei `r.TLS != nil` einen Header (`X-Forwarded-Proto`) oder ändert den Pfad | andere |
| B14 | Ein gescheiterter Handshake erreicht den Adapter nicht: kein `Verarbeiten`, keine `KorrelationsID`, keine Zeile in einer Senke; danach kehrt `Schliessen(0)` sofort mit `nil` zurück (Nr. 9 b) | `TestHandshakeFehler`, Menge H' (4: unbekannte CA, Client allein mit `h2`, Client höchstens TLS 1.1, Klartext-Bytes): der Fake zählt `Verarbeiten` und `IDs` (je 0), `log.Default()` und stderr bleiben leer | `ErrorLog` des Servers auf `log.Default()` (eine Zeile) | andere |
| B15 | Ein Handshake, der hängt, hält `Accept` und andere Clients nicht auf (Nr. 9 a) | `TestHaengenderHandshake`: eine TCP-Verbindung ohne Byte offen, ein zweiter Client mit gültigem Handshake und Anfrage erhält die Antwort binnen 5 s; Wirkungs-Fänger über 20 Wiederholungen (jeder muss grün sein) | eine Umhüllung, deren `Accept` `Handshake()` ruft: der zweite Client wartet, rot in 20 von 20 Läufen | nebenläufig |
| B16 | `Anhalten` gilt mit TLS: ein Request, dessen Zulassung danach liegt, erhält 503 `ODR-E5010` mit `Connection: close`, auch wenn sein Handshake erst nach `Anhalten` endet; `Verarbeiten` wird nicht gerufen (Nr. 9 c) | `TestAnhaltenMitTLS`, Menge (2: Verbindung mit beendetem Handshake und Keep-Alive; Verbindung, deren Accept durch `meldeListener` feststeht und deren Handshake nach `Anhalten` endet) | die Zulassung ohne Prüfung auf `angehalten` (der Fänger der Zusage ohne TLS, hier an beiden Elementen rot) | andere |
| B17 | `Schliessen` gilt mit TLS: ein Handshake in der Schwebe wird sofort geschlossen, ein laufender Request wird bis zur Frist abgewartet, danach `ErrFristAbgelaufen` (Nr. 9 c) | `TestSchliessenMitTLS`, Menge (4: Verbindung ohne Byte; Verbindung mit halbem ClientHello; laufender Request, der innerhalb der Frist endet → `nil`; laufender Request über die Frist → `ErrFristAbgelaufen`); bei den ersten zwei kehrt der Aufruf mit Frist 10 s binnen 1 s zurück und die Verbindung ist zu | `schliesseOhneHandler` entfällt (die ersten zwei: Verbindung offen) | andere |
| B18 | `Herunterfahren` des Roots gilt mit TLS: der Listener ist danach zu, jede Verbindung geschlossen, binnen `SHUTDOWN_TIMEOUT` (Nr. 9 c, `SPEC-139` Nr. 7) | `TestHerunterfahrenMitTLS` (Modus `mock`): eine offene HTTPS-Verbindung im Leerlauf und ein Handshake in der Schwebe; nach `Herunterfahren` scheitert `Dial`, beide Verbindungen sind zu, die Dauer liegt unter der Frist | eine Umhüllung, deren `Close` den inneren Listener nicht schließt (`Dial` gelingt) | andere |
| B19 | Der Start nennt in keiner Zeile von stderr und keinem Logeintrag einen Pfad, Bytes der Dateien oder das Subject des Zertifikats (Nr. 11) | `TestStartTLSKeinGeheimnis`: `start` mit gültigem Paar (Dateien unter `geheim-pfad-4711`, Subject `geheim-cn-4711`) und mit jedem der 8 Verstöße von B02: stderr enthält keinen der drei Texte | `slog.Info("tls", "cert", pfad)` in `starteListener` | andere |
| B20 | Kein Neuladen: nach `Build` überschriebene Dateien ändern das gelieferte Zertifikat nicht; ein abgelaufenes wird weiter bedient (Nr. 10) | `TestKeinNeuladen`: nach `Hochfahren` werden beide Dateien durch das Paar eines anderen Ausstellers ersetzt, ein neuer Handshake zeigt die Seriennummer des ersten Blatts; der Client mit der Wurzel des zweiten scheitert | `GetCertificate` liest die Dateien bei jedem Handshake | andere |
| B21 | Das Kataloge-Gate hält die zwei Schlüssel: Quelltext, `SPEC-110` Nr. 2 und Handbuch nennen sie, kein anderer `LISTEN_TLS_*` steht im Handbuch (`SPEC-201` Nr. 4) | Ausgabe `katalog: ok` von `make katalog-check` | die Zeile `LISTEN_TLS_KEY` aus der Tabelle des Handbuchs gelöscht (Befund); `LISTEN_TLS_` wieder in `nichtGelieferte` (Befund `Handbuch nennt abgelehnt`); das Präfix als Literal in `konfig_tls.go` (Befund: nicht in der Spezifikation) | andere |
| B22 | Das Handbuch zeigt den Fall „nur ein Schlüssel“ als gebundenes Beispiel gegen das gebaute Binary: Exit 1 und die Zeile `ODR-E2001 LISTEN_TLS_KEY: erforderlich, wenn LISTEN_TLS_CERT gesetzt ist` (`SPEC-200` Nr. 1) | Ausgabe `ok` von `make handbuch-check` | der Text des Grundes in `tlsdatei.go` geändert (das Gate wird rot) | andere |

**Mengen und Zählprüfung gegen die Quelle** (je Menge zählt der Test die Elemente und vergleicht mit der Zahl hier; die Quelle steht dahinter): E 4 (`SPEC-291` Nr. 2, zwei Schlüssel × gesetzte Datei gültig oder fehlend) · L 6 (Nr. 2 Rang 2: drei Formen × zwei Schlüssel) · Z 7 und S 4 (Nr. 2 und 4, Rang 3) · P 6 (Rang 4 des Schlüssels) · G 6 (Nr. 3) · Q 25 (5 × 5 Zustände von Nr. 2) · F 8 und K 5 (Nr. 4: die zugesagten Formen der Datei und die fünf Schlüsselformen) · N 2 (Nr. 2 Kopf) · V 8 (die vier Ränge je Schlüssel, Nr. 2) · Modi 5 und 4 (`SPEC-110` Nr. 8: vier lieferbare Modi, dazu der ungültige Wert) · R 5 abgelehnt und 3 ignoriert (`SPEC-110` Nr. 2) · H 6 (3 Hosts von `SPEC-110` Nr. 14 c × 2) · H4 4 und H' 4 (Nr. 9 b) · T 6 (2 Transporte × 3 Fälle). Die Gründe: sechs Zeilen in der Tabelle von Nr. 2, fünf feste Literale im Quelltext (A14). Achsen der Mengen: die Stelle des Elements im Wert (Block vor oder nach Fremdtext, erstes Zertifikat der Kette), der Schlüssel (Zertifikat oder Schlüssel), die Grenze (beide Seiten und gleich: G), der Modus (B03, B10, B11) und der Weg zum Aufruf (Laden, `Build`, `Hochfahren`: B02, B10, B11).

*Mutationen, Raten und Messende trägt der Implementer in einen Absatz unter dieser Tabelle ein (`AGENTS.md` §3.10); die Raten von A16 und B15 mit Zahl der Prozesse und Last des Hosts.*

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

**Start** (`next` → `in-progress`): `slice-http-adapter`, `slice-bootstrap-m2` und `slice-handbuch-meldungskatalog` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen und den Nachzug von `SPEC-110`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12): `SPEC-291` und der Nachzug liegen vor (Architect, 2026-11-21); offen ist der Schnitt (Planner, siehe *Größenmarke*).

**Abhängigkeiten — Grund je Eintrag:**

- `slice-http-adapter` (`welle-record-replay`): der Listener, dessen Transport der Slice ändert; `src/adapters/driving/httpodata` besteht heute nicht.
- `slice-bootstrap-m2` (`welle-record-replay`): der Composition Root, der Konfiguration lädt und den Listener öffnet. Der Slice ändert die Startkonfiguration und verdrahtet den Baustein; ohne `LadeOderMelde` im Startbefehl (heute ruft der Adapter `cli` (`cli.Run`) keinen Start; der Container startet mit `start`) gibt es keinen Ort für die Meldung `ODR-E2001` am Binary und keinen Prozess für den HTTPS-Test.
- `slice-handbuch-meldungskatalog` (`welle-benutzerhandbuch`, setzt `slice-handbuch-geruest` voraus): die DoD verlangt das Handbuch-Kapitel und ein grünes Kataloge-Gate; zwei neue Schlüssel (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`) und ein Code-Fall brauchen das Gate.

**Unabhängig von `welle-mock` / M3, weil** der Listener modusunabhängig ist und der HTTPS-Test den Modus `record` oder `replay` nutzt; weder `MockUseCase` noch der Modus `mock` werden gebraucht. **Nicht unabhängig von M2:** der Baustein hängt am Listener des HTTP-Adapters und am Start im Composition Root von `slice-bootstrap-m2`. Ein Teilstart mit dem Baustein allein (nach `slice-http-adapter`, ohne Verdrahtung) ist möglich, ergibt aber nicht die DoD und wird hier nicht geplant; schneidet der Architect den Baustein ab, geht die Verdrahtung an einen Folge-Slice und der Planner passt diese Datei an. `slice-container-image`, die Telemetrie und `slice-lasttest` sind nicht Bedingung.

**Größenmarke:** noch nicht gesetzt; der Teil des Architect steht im Block unten (2026-11-21, `SPEC-128` Nr. 3 k), `basis`, `schaetzung` und `marke` setzt der Planner vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit `make groessenmarke` (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Der Architect überlässt das Lesen der PEM der Standardbibliothek (`SPEC-291` Nr. 4), der Zuschlag für einen eigenen Leser fremder Bytes entfällt.

**Schnitt vor dem Code (Vorschlag des Architect, 2026-11-21, `AGENTS.md` §3.14, `SPEC-128` Nr. 1):** die Tabelle in §3 hat 39 Zeilen mit *a* = 2 (`A16`, `B15`), *b* = 37, *f* = 1 (der TLS-Test-Server von `B13` bis `B18`, mit Haltepunkt im Handshake), *p* = 180 (`tlsdatei.go` rund 100, `konfig_tls.go` rund 40, die Änderungen in `konfig.go`, `konfig_basis.go`, `konfig_werte.go`, `wurzel.go` und `lebenszyklus.go` rund 40): *S* = 85·2 + 80·37 + 200·1 + 180 = 3510, *M* = 4800, Schnittfrage (*S* ≥ 2223). Der Teil lässt sich ohne geteilten Mechanismus abtrennen, weil der Baustein nur die Standardbibliothek und keinen Zustand des Roots braucht und der Rest ihn nur aufruft:

| Teil | Zeilen der Tabelle | *a* | *b* | *f* | *p* | *S* | *M* | Inhalt |
|---|---|---|---|---|---|---|---|---|
| 1 (Baustein) | `A01` bis `A17` | 1 | 16 | 0 | 100 | 1465 | 2000 | `src/bootstrap/tlsdatei.go`, `src/hexagon/ports/support/tlstest/`, ihre Tests; kein Eintrag in Konfiguration und Start, nichts am Binary |
| 2 (Konfiguration und Verdrahtung) | `B01` bis `B22` | 1 | 21 | 1 | 80 | 2045 | 2800 | `konfig_tls.go`, die Änderungen in `konfig*.go`, `wurzel.go`, `lebenszyklus.go`, die Tests des Adapters mit TLS, Handbuch, Katalog |

Der Teil 2 liegt mit *S* = 2045 unter der Schnittgrenze (178 Zeilen Abstand); jede weitere Zeile der Tabelle von Teil 2 kostet 80, drei weitere lösen die Schnittfrage neu aus. Teil 1 liefert nichts am Binary (`ladeTLS` ruft nur der Test; `unused` zählt Tests mit, `make lint` bleibt grün) und ist der Teil, den `slice-admin-tls` als Bedingung braucht (der Planner trägt die Kennung dort ein, `AGENTS.md` §3.13); Teil 2 hängt an Teil 1 (*Code-Abhängigkeit*). Gegen einen Schnitt „Konfiguration ohne Verdrahtung“ sprach, dass der Start `LISTEN_TLS_*` dann annähme und der Listener bei HTTP bliebe, eine Sicherheitseinstellung ohne Wirkung (`SPEC-110` Nr. 2); der Schnitt liegt deshalb vor der Annahme der Schlüssel. Geschnitten wird der Plan, nicht die Spezifikation: `SPEC-291` bleibt eine Nummer. Den Block unten ersetzt der Planner beim Schnitt je Teil durch den Block des Teils (Pfade `src/bootstrap/` und `src/hexagon/ports/support/tlstest/` für Teil 1; `src/bootstrap/` und `src/adapters/driving/httpodata/` für Teil 2).

```groessenmarke
pfad: src/bootstrap/
pfad: src/adapters/driving/httpodata/
pfad: src/hexagon/ports/support/tlstest/
produktcode: 180
fakes: 1
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start setzt; der Schnitt vor dem Code liegt vor (siehe *Schnitt vor dem Code* oben, die Naht ist Baustein gegen Konfiguration und Verdrahtung, `AGENTS.md` §3.14).
- `in-progress` → `open` (blockiert): der Nachzug von `SPEC-110` fehlt oder der Baustein lässt sich nur mit einem Import des anderen Driving Adapters teilen (siehe §6).

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
- Sendung ohne Nehmer mit Datei: `ADMIN_TLS_*` und die Start-Prüfung bei nicht-loopback `ADMIN_LISTEN` gehören `welle-admin-api` (`slice-admin-tls`), dessen Datei noch nicht besteht — **Ausgang:** offen; Adresse: Planner beim Anlegen der Slices von `welle-admin-api` (`AGENTS.md` §3.13).
- Start mit einer Schlüsseldatei, die Gruppen- oder Weltrechte trägt: die Spezifikation verlangt es für die Token-Datei, nicht für den TLS-Schlüssel — **Ausgang:** entschieden (Architect, 2026-11-21): der Start prüft die Rechte der Schlüsseldatei nicht (`SPEC-291` Nr. 10, Alternative, Grund und Folge dort); Fänger `A09` (Element acht, Rechte `0666`). Der User hat Option A (keine Prüfung) am 2026-11-21 bestätigt (AskUserQuestion); die Alternative, eine Prüfung wie bei der Token-Datei, ist damit verworfen.
- Geheimnis in Fixtures: ein privater Schlüssel im Repo ist ein Befund gegen `LH-RB-04` — **Ausgang:** entschieden (Auftrag 6): die Fixtures entstehen zur Laufzeit (`tlstest`, `t.TempDir()`), keine Datei mit Schlüssel im Repo; das Beispiel des Handbuchs (`B22`) braucht keine.
- HTTP/2 über TLS ändert das Verhalten von Strömen und Abbruch gegenüber dem Listener ohne TLS — **Ausgang:** entschieden (Architect, 2026-11-21): der Listener bietet HTTP/2 nicht an, ALPN allein `http/1.1` (`SPEC-291` Nr. 6); Fänger `A12` (die drei Clients) und `B11` (`ProtoMajor` 1 bei einem Client, der HTTP/2 anbietet).
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Architect, 2026-11-21): das Raster unten trägt in jeder Zelle einen Ort oder `gilt nicht` mit Grund; eine Zelle, die der Implementer ohne Ort findet, geht an den Architect zurück (`AGENTS.md` §3.12).
- Schnitt und Größenmarke: die Schätzung liegt über der Schnittgrenze (*S* = 3510) — **Ausgang:** offen; Adresse: Planner (Schnitt in Teil 1 und Teil 2, Block je Teil, Kennung des Teils 1 in `slice-admin-tls`, `slice-container-image`, `welle-betrieb`; §4 *Schnitt vor dem Code*).
**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; jede Zelle nennt ihren Ort in der Spezifikation oder `gilt nicht` mit Grund):

| Operation | fehlender Wert (nur einer der beiden Schlüssel) | leerer Wert | unbekannter oder fremder Wert (Datei fehlt, kein PEM, Schlüssel passt nicht, abgelaufen) | Abbruch (Herunterfahren während des Handshakes) | Wiederholung (zweiter Start, zweiter Handshake) | Modus (`record`, `replay`, `mock`, `hybrid`) |
|---|---|---|---|---|---|---|
| Startkonfiguration prüfen (`ODR-E2001`) | `SPEC-291` Nr. 2 Rang 1 und Nr. 7 b (der leere Schlüssel meldet, der gesetzte wird weiter geprüft); `SPEC-110` Nr. 2 und 7 | `SPEC-110` Nr. 1 (leer und nur Leerraum gelten als ungesetzt), `SPEC-291` Nr. 7 a | `SPEC-291` Nr. 2 Ränge 2 bis 4, Nr. 3, 4 und 7 c; jeder andere `LISTEN_TLS_*`: `SPEC-110` Nr. 2 | `SPEC-110` Nr. 13 (das Laden ist synchron ohne Kontext), `SPEC-138` Nr. 8 d | `SPEC-291` Nr. 12; `SPEC-110` Nr. 13 (Ausnahme: die Uhr, `SPEC-291` Nr. 3) | `SPEC-291` Nr. 7 a und e, `SPEC-110` Nr. 8 (in jedem Modus, auch bei ungültigem `MODE`) |
| Zertifikat und Schlüssel laden (`ladeTLS`, `Build`) | `SPEC-291` Nr. 2 (beide leer: kein TLS; einer leer: Rang 1) | `SPEC-291` Nr. 2 Rang 3 (leere Datei) und Rang 1 (leerer Pfad) | `SPEC-291` Nr. 2 Ränge 2 bis 4, Nr. 3, 4 | gilt nicht: `ladeTLS` läuft synchron ohne Kontext und hält keinen Zustand (`SPEC-291` Nr. 12); `Build` liest die Dateien ohne Kontext, wie `UPSTREAM_CA_FILE` (`SPEC-139` Nr. 2) | `SPEC-291` Nr. 8 (zweimal gelesen, `Build` gilt) und Nr. 12 | `SPEC-291` Nr. 8 (`Build` lädt in jedem Modus) |
| Listener mit TLS öffnen und bedienen | `SPEC-291` Nr. 1 und 8 (ohne Pfade bleibt der Listener Klartext) | gilt nicht: `Hochfahren` erhält nur ein geprüftes `*tls.Config` oder keines (`SPEC-291` Nr. 8); leere Pfade sind die Zeile darüber | gilt nicht: ein Verstoß endet in `Build` vor dem Öffnen (`SPEC-291` Nr. 8); eine nicht bindbare Adresse: `SPEC-139` Nr. 8 c | `SPEC-291` Nr. 9 c, `SPEC-136` Nr. 13 bis 15 | `SPEC-291` Nr. 8 und 12 (ein zweiter `Build` ist unabhängig), `SPEC-139` Nr. 6 (ein zweites `Hochfahren` ist `model.ErrUngueltig`) | `SPEC-291` Nr. 7 e, `SPEC-136` Nr. 2 (der Listener kennt den Modus nicht) |
| Handshake des Clients (gültig, unbekannte CA, falscher Hostname) | `SPEC-291` Nr. 5 (ein Client ohne Zertifikat gelingt) und Nr. 6 (ein Client ohne ALPN spricht HTTP/1.1) | `SPEC-291` Nr. 9 a und c (ein Client, der verbindet und nichts sendet) | `SPEC-291` Nr. 9 b (unbekannte CA, falscher Name, Version unter 1.2, Klartext, nur `h2`) | `SPEC-291` Nr. 9 c (Herunterfahren während des Handshakes), Nr. 9 b (der Client bricht ab: der Adapter sieht nichts) | `SPEC-291` Nr. 9 a (ein hängender Handshake hält keinen anderen auf) und Nr. 12 (die Konfiguration ändert sich nicht) | `SPEC-291` Nr. 7 e (der Handshake ist vom Modus unabhängig) |
| Zweite Verwendung des Bausteins mit anderen Schlüsselnamen | `SPEC-291` Nr. 2 Rang 1 (der Name des anderen Schlüssels ist der des Aufrufers) | `SPEC-291` Nr. 2 Rang 1 (leerer Pfad) | `SPEC-291` Nr. 2 Ränge 2 bis 4 (der Schlüssel der Meldung ist der Name des Aufrufers, Nr. 7 c für `LISTEN_`) | gilt nicht: derselbe Grund wie beim Laden (kein Kontext, `SPEC-291` Nr. 12) | `SPEC-291` Nr. 12 (zustandslos: zwei Aufrufer mit zwei Namenspaaren ändern einander nicht) | gilt nicht: der Baustein kennt keinen Modus; die Modus-Pflichten des Admin-Listeners (`SPEC-089`, `SPEC-090`) liegen beim Aufrufer (`SPEC-291` Kopf) |

**Zusatztabelle** (`AGENTS.md` §3.16: die Operation meldet mehrere Verstöße und hängt vom Modus ab): je Paar von Verstößen und je Modus eine Zeile.

| Paar oder Modus | Entschieden in | Fänger |
|---|---|---|
| zwei Verstöße am selben Schlüssel (Rang 1 und 2, 2 und 3, 3 und 4): schließen einander aus, jeder Rang setzt das Bestehen der davor voraus; die Meldung trägt den kleinsten Rang | `SPEC-291` Nr. 2 | `A08` (Menge Q, Orakel mit einem Grund je Schlüssel); `A03` bis `A07` je Element |
| zwei Verstöße an verschiedenen Schlüsseln (Zertifikat außerhalb der Gültigkeit und Schlüssel passt nicht; kein PEM-Zertifikat und kein PEM-Schlüssel; nicht lesbar und nicht lesbar; …): beide Zeilen, nach Schlüssel sortiert; der Rang 4 des Schlüssels läuft nur, wenn beide Rang 3 bestanden | `SPEC-291` Nr. 2, `SPEC-110` Nr. 12 | `A08` (25 Paare); `B02` (ein Paar über `LadeOderMelde`); `B10` (ein Paar in `Build`) |
| ein Verstoß am Zertifikat (Rang 4) und ein Zertifikat, dessen Zwischenzertifikat abgelaufen ist: allein das Blatt zählt | `SPEC-291` Nr. 3 | `A07` (Element sechs) |
| Modus `record` | `SPEC-291` Nr. 7 a und e, Nr. 8 | `B03`, `B10`, `B11` |
| Modus `replay` | wie `record` | `B03`, `B10`, `B11`, `B06` (die Basis bleibt leer) |
| Modus `mock` | wie `record`; `PUBLIC_BASE_URL` trägt `https` bei TLS (Nr. 7 d) | `B03`, `B06`, `B07`, `B10`, `B11`, `B18` |
| Modus `hybrid` | wie `mock` | `B03`, `B10`, `B11` |
| ungültiger Wert von `MODE` | `SPEC-110` Nr. 8 (die Syntax aller Schlüssel wird in jedem Modus geprüft) | `B03` |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. Der Slice liest fremde Bytes (PEM-Dateien des Betreibers): `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (3×, verkörpert in `.claude/commands/implement-slice.md` Schritt 14, `.claude/agents/implementer.md` und `.harness/skills/reviewer.md`) und `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (5×, verkörpert in `AGENTS.md` §3.17) gelten für „jede Fehlform“ von Zertifikat und Schlüssel; `BEO-REPO/randform-erst-im-review-entschieden` (20×, verkörpert in `AGENTS.md` §3.16) für das Raster. Offen: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) für Testhilfen, die Verbindungen und Ströme umhüllen, und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
