# Slice slice-admin-tls: Transportschutz des Admin-Listeners

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

**Berührte Spec-Stellen:** `LH-FA-13.a` (Transportschutz) · `LH-RB-02.a` (Listener-TLS) · `LH-QA-04.a` · `SPEC-299` (der Vertrag dieses Slice) · `SPEC-008`, `SPEC-089`, `SPEC-090` (Defaults, verweisen auf `SPEC-299`) · `SPEC-291` (der Baustein, Kopf) · `SPEC-238` (*Ungeprüft*: TLS und HTTP/2) · `SPEC-110` Nr. 10 und 11 (Muster der Zusammenhangs-Prüfung; der Nachzug dort gehört `slice-bootstrap-m6`) · `ARC-007` · `architecture.md` §Sicherheit.
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

**Ziel:** Der Composition Root trägt die Funktionen, die den Admin-Listener schützen (`SPEC-299`, Datei `src/bootstrap/admintls.go`): sie laden das Zertifikatspaar `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` mit dem Baustein von `SPEC-291` und umhüllen den Listener mit TLS; bei nicht-loopback `ADMIN_LISTEN` verlangen sie das Paar und verhindern sonst die Ausführung (`ODR-E2001`), nur `ADMIN_ALLOW_PLAINTEXT=true` erlaubt Klartext, mit einer Warnung im Log; nur einer der beiden Schlüssel des Paars ist `ODR-E2001`. Die Tests belegen Start-Prüfung, Weg, Umhüllung und Handshake an einem Admin-Server im Prozess; das Binary bleibt unverändert, weil erst `slice-bootstrap-m6` die Funktionen ruft.

**Übernimmt:** den wiederverwendbaren TLS-Baustein des Listeners aus `slice-listener-tls` (Teil 1 des Schnitts vom 2026-11-21: `ladeTLS` in `src/bootstrap/tlsdatei.go` und das Test-Hilfspaket `tlstest`, `SPEC-291`; `welle-betrieb` §4 nennt ihn als Grundlage des Admin-Listeners); der Slice nutzt ihn unverändert, statt das Laden und Prüfen des Zertifikatspaars ein zweites Mal zu schreiben.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Listener und die Authentisierung selbst — `slice-admin-listener-token`; TLS umhüllt ihn, und der Adapter `admin` bleibt unverändert (`SPEC-299` Nr. 1: er kennt TLS nicht).
- TLS des OData-Listeners (`LISTEN_TLS_*`) — `slice-listener-tls-verdrahtung` (`welle-betrieb`, Teil 2 des Schnitts von `slice-listener-tls`); diesen Teil braucht der Slice nicht.
- Erzeugen oder Erneuern von Zertifikaten, Client-Zertifikate (mTLS), Rotation und Neuladen zur Laufzeit — `SPEC-299` Nr. 11 schließt sie aus; der Betreiber stellt das Paar bereit.
- Das Lesen der `ADMIN_*`-Schlüssel aus der Umgebung, das Aufheben von `SPEC-110` Nr. 2, der Nachzug von `SPEC-110` Nr. 10 und 11, der Aufruf der Funktionen aus Laden, `Build`, `Hochfahren` und `Start`, das Schreiben der Warnung am Start, das Handbuch und die Gründe des Starts am Binary — `slice-bootstrap-m6` (Grund wie in `slice-admin-listener-token`: ohne Admin-Listener im Binary gibt es nichts, was ein Anwender sähe; die Sendungen stehen in §6).
- Der Netzwerknachweis von außen — `slice-test-security`; hier prüfen Tests die Start-Prüfung und den Handshake.

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

- [ ] HTTPS-Betrieb (`SPEC-299` Nr. 4, 7, 8 c, 10, 11): das Zertifikatspaar wird mit den Namen `ADMIN_TLS_*` geladen, der Listener wird umhüllt und bedient; ein Client mit der CA gelingt, ein Client mit unbekannter CA lehnt ab und erreicht den Adapter nicht; nur einer der beiden Schlüssel, eine unlesbare Datei, ein ungültiges oder nicht passendes Paar sind `ODR-E2001`, und der Start fällt nie auf Klartext zurück (Zeilen AT03, AT04, AT10 bis AT17 der Tabelle in §3).
- [ ] Start-Prüfung (`SPEC-299` Nr. 2, 3, 5, 6, 9): ein nicht-loopback `ADMIN_LISTEN` (auch die Wildcard-Adressen) ohne TLS und ohne `ADMIN_ALLOW_PLAINTEXT=true` ist ein Verstoß (`ODR-E2001`); mit `ADMIN_ALLOW_PLAINTEXT=true` ist der Weg Klartext mit einer Warnung im Log; Loopback ohne TLS bleibt zulässig (Zeilen AT01 bis AT03 und AT05 bis AT08).
- [ ] Je Zelle der Matrix Loopback × TLS × Plaintext-Flag ein Fänger (32 Zellen aus 8 Hosts, AT02) und die Schreibweisen des Hosts als Menge mit Zählprüfung (AT01), nach dem Vertrag des Architect.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/bootstrap/admintls.go` | neu | `SPEC-299` Nr. 1: `adminLoopback`, `pruefeAdminTransport`, `adminTransport` mit `umhuelle` und `warne`, die Namen der Schlüssel als Konstanten; ruft `ladeTLS` mit `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY`; heißt nicht `konfig*.go` (`SPEC-299` Nr. 1, 14). Nichts am Binary: kein Aufruf aus Laden, `Build` und `Start` |
| `src/bootstrap/admintls_test.go` | neu (Paket `bootstrap_test`, Lint `testpackage`) | AT01 bis AT10, AT18: die Mengen der Funktionen ohne Netz; Zertifikatspaare zur Laufzeit mit `tlstest`, Dateien in `t.TempDir()` |
| `src/bootstrap/admintls_betrieb_test.go` | neu (Paket `bootstrap_test`) | AT11 bis AT17: der Test-Server (`admin.New` über `umhuelle` und `net.Listen`, zählende `IDs`, `Audit` und `Telemetrie` aus `supporttest`) und die Handshakes mit bekannter und unbekannter CA; `admin` wird hier als Adapter eines Tests importiert (der Root darf es, `.a-check.yml` bleibt unverändert) |
| `src/bootstrap/export_test.go` | update (Naht, additiv) | die unexportierten Namen von `SPEC-299` für das Paket `bootstrap_test`: `AdminLoopback`, `PruefeAdminTransport`, die Typen von Eingabe und Transport als Alias, `AdminUmhuelle`, `AdminWarne`, die Konstanten der Gründe und Wege (wie `LadeTLS` dort); die einzige Datei dieses Slice, die ein anderer Slice des Pakets schreiben kann, darum im Block einzeln benannt |
| `src/bootstrap/tlsdatei.go`, `src/hexagon/ports/support/tlstest/` | nutzen, unverändert | Laden und Prüfen des Paars (Zeile A17 von `slice-listener-tls` belegt die zweite Verwendung); die Hilfen des Pakets `bootstrap_test` aus `tls_betrieb_test.go` (`httpsClient`, `neueWerkstatt`, `tlsBedient`) darf der Test nutzen, nicht ändern |
| `src/adapters/driving/admin/` | **unberührt** | Berichtigung des ersten Plans: der Adapter erhält den umhüllten `Listener` und kennt TLS nicht (`SPEC-238` Nr. 1, `SPEC-299` Nr. 1); kein Pfad dieses Verzeichnisses steht im Block |
| `src/bootstrap/wurzel.go`, `start.go`, `konfig*.go` | **unberührt** | der Aufruf aus Laden, `Build`, `Hochfahren` und `Start` ist `slice-bootstrap-m6` (`SPEC-299` Nr. 8); keine Kollision mit seinen Dateien |
| `spec/spezifikation.md`, `spec/architecture.md` | geliefert (Architect, 2026-11-23) | `SPEC-299` neu; Zeiger in `LH-FA-13.a`, `LH-RB-02.a`, `SPEC-089`, `SPEC-090`, `SPEC-291` (Kopf) und `SPEC-238` (*Ungeprüft*); ein Satz in `architecture.md` §Sicherheit; dieser Slice schreibt sie nicht |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen mit dem Vertrag in der Spezifikation vor dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Entscheidungen des Architect** (2026-11-23; Auftrag aus der Anlage des Plans → Ort in der Spezifikation):

| Auftrag | Entscheidung | Ort |
|---|---|---|
| 1 Vertrag der Start-Prüfung, Loopback, Wildcard, Name | Loopback sind genau `127.0.0.1`, `[::1]` und `localhost` (Schreibweise zählt, wie `SPEC-110` Nr. 14 c); alles andere, auch `127.0.0.2`, `0.0.0.0`, `[::]`, der leere Host und jeder Name, ist nicht-loopback; kein Name wird aufgelöst | `SPEC-299` Nr. 2 |
| 2 Zusammenhänge, Flag, Reihenfolge der Meldung | Flag nur `true` oder `false`, ungesetzt ist `false`; Paar gewinnt über das Flag, das Flag ist bei TLS und bei Loopback ohne Wirkung und kein Verstoß; ein Verstoß am Schlüssel `ADMIN_LISTEN`, wenn kein Pfad, nicht-loopback, Flag falsch; ungültige Form oder genau ein Pfad: der Zusammenhang entfällt; Reihenfolge nach `SPEC-110` Nr. 12 | `SPEC-299` Nr. 3, 5, 6 |
| 3 TLS-Parameter | wie `SPEC-291` Nr. 5 und 6 (TLS 1.2 und 1.3, Suiten der Bibliothek, nur HTTP/1.1, kein mTLS), keine eigene Entscheidung für den Admin-Listener | `SPEC-299` Nr. 7 |
| 4 Warnung im Log | Stufe WARN, Text `start: Admin-Listener ohne TLS erreichbar`, ein Attribut `ereignis` (`admin_klartext`), einmal je Start, nur im Weg Klartext, kein Geheimnis | `SPEC-299` Nr. 9 |
| 5 Ort der Prüfung | im Composition Root, Datei `src/bootstrap/admintls.go`; der Adapter bleibt unverändert; `slice-bootstrap-m6` ruft sie | `SPEC-299` Nr. 1, 8 |
| 6 Tabelle, Raster, Nachzug | Tabelle und Mengen unten, Raster in §6, Nachzug der Stellen aus der Tabelle oben | dieser Abschnitt, §6 |

**Handbuch und Katalog** (Auftrag des Architect; Texte schreibt `slice-bootstrap-m6`, `SPEC-200`, `SPEC-201` Nr. 11): dieser Slice ändert weder Handbuch noch `harness/katalog-ausnahmen.txt` noch `konfig*.go`, denn das Binary liest die Schlüssel nicht und zeigt weder Warnung noch Gründe (`SPEC-299` Nr. 14); `make katalog-check` und `make handbuch-check` bleiben ohne Änderung grün (AT20). `slice-bootstrap-m6` zieht nach: (1) die Tabelle `Einstellungen` des Handbuchs um `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` (und `ADMIN_LISTEN`, `ADMIN_TOKEN_FILE` aus seinem Anteil) und `SPEC-110` Nr. 2 (*Geprüft werden*) um dieselben Schlüssel, das Präfix `ADMIN_` aus `nichtGelieferte` heraus; (2) das Kapitel `TLS` um einen Abschnitt für den Admin-Listener nach dem Muster des Abschnitts zu `LISTEN_TLS_*`: Voraussetzung, Vorgehen, Ergebnis (TLS 1.2 und 1.3, HTTP/1.1, Dateien einmal beim Start, Ablauf im Betrieb), die Tabelle der Gründe (die sechs von `ladeTLS` unter den Namen `ADMIN_TLS_*`, dazu der Grund `außerhalb von Loopback verlangt TLS: ADMIN_TLS_CERT und ADMIN_TLS_KEY` am Schlüssel `ADMIN_LISTEN`), die Warnung des Weges Klartext mit ihrem festen Text und die Verantwortung des Betreibers für den Klartext, und die Aussage, dass ein abgelehnter Handshake beim Betreiber keine Zeile hinterlässt; (3) ein Beispiel als Datei gegen das gebaute Binary (`SPEC-200`): `ADMIN_LISTEN=0.0.0.0:8081` ohne Paar und ohne Flag, Exit 1 und die Zeile `ODR-E2001 ADMIN_LISTEN: …`; (4) Fehlercodes: `ODR-E2001` steht schon im Handbuch (H), kein neuer Code und keine neue Ausnahme.

**Zusage · Fänger** (`AGENTS.md` §3.15; Architect, 2026-11-23; Kennung `AT`; die Spalte *Zusage* nennt den Ort in `SPEC-299`; jede Zeile trägt ihre Menge, die Mengen stehen mit Achsen unten; alle Tests laufen im Paket `bootstrap_test` ohne Netz außer dem Loopback des Prozesses, kein Schlüssel im Repo; die Namen der Funktionen stehen hier, wie `SPEC-299` sie führt, der Test erreicht sie über die Naht in `export_test.go`):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| AT01 | (Nr. 2) `adminLoopback` ist wahr genau für die Hosts `127.0.0.1`, `[::1]` und `localhost`, bytegleich, bei jedem Port; jeder andere Host, auch ohne Host und ohne `:`, ist falsch | `TestAdminLoopback`, Menge M-Host (3 Hosts loopback, 17 nicht-loopback, je 3 Ports, dazu 3 Eingaben ohne `:`: 63 Zellen) und die Nachbarn M-Nachbar (je loopback-Host jede Einzelzeichen-Änderung über das Alphabet `0-9`, `a-z`, `A-Z`, `.`, `:`, `[`, `]`, `-`, `_` und jede andere Groß-/Kleinschreibung von `localhost`: Orakel falsch, Zählprüfung gegen die Formel je Host); Orakel als Literal, nie aus dem Code | Vergleich ohne Beachtung der Schreibweise (rot für `LOCALHOST`, `Localhost`); Präfix `127.` (rot für `127.0.0.2`); `net.ParseIP(…).IsLoopback()` nach Entfernen der Klammern (rot für `[::ffff:127.0.0.1]`, `127.0.0.2`); `strings.HasPrefix(host, "localhost")` (rot für `localhostx`, `localhost.`); `strings.HasSuffix(host, ".1")` (rot für `10.0.0.1`); Port 8081 im Vergleich (rot für Port 1 und 65535) | andere |
| AT02 | (Nr. 3) Der Weg folgt der Tabelle: TLS gewinnt über Loopback und Flag, Loopback ohne Paar ist `wegLoopback`, nicht-loopback ohne Paar mit Flag ist `wegKlartext`, ohne Flag der Verstoß; das Flag wirkt nur im Weg Klartext | `TestAdminWeg`, Menge M-Weg: 8 Hosts (3 loopback, 5 nicht-loopback) × TLS (gültiges Paar, kein Paar) × Flag (falsch, wahr) = 32 Zellen; Orakel je Zelle als Literal (Weg oder die eine Zeile); Zählprüfung 8 × 2 × 2 | `klartext` wirkt bei Loopback (rot: `wegKlartext` statt `wegLoopback`, 3 Zellen); Flag überstimmt TLS (rot: 8 Zellen); Flag ignoriert (rot: 5 Zellen); TLS ignoriert bei Loopback (rot: 6 Zellen); die Bedingung des Weges Klartext mit „oder“ statt „und“ zwischen Loopback und Flag (rot) | andere |
| AT03 | (Nr. 3) Ein Paar mit Verstoß liefert keinen Weg, und der Start fällt nie auf Klartext zurück, auch nicht mit Flag und nicht bei Loopback | `TestAdminKeinRueckfall`, Menge M-Rueckfall: die 9 einzelnen Verstöße von M-Paar × Host (`0.0.0.0:8081`, `127.0.0.1:8081`) × Flag wahr = 18 Zellen; Orakel: dieselben Zeilen wie AT04, `Weg` gleich `wegKeiner`, `TLS` nil | Rückfall `wegKlartext`, wenn Verstöße und Flag (rot: 9 Zellen mit `0.0.0.0`); Rückfall `wegLoopback`, wenn Verstöße und Loopback (rot: 9 Zellen) | andere |
| AT04 | (Nr. 4) Das Paar geht unverändert durch `ladeTLS` mit den Namen `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY`: jeder Verstoß steht unter dem Namen seines Schlüssels, der Grund des Rangs 1 nennt den anderen Namen, beide Pfade leer ist kein TLS, und nichts wird gelesen; zwei Verstöße stehen nach Schlüssel sortiert | `TestAdminPaar`, Menge M-Paar: 9 Verstöße (Rang 1 an beiden Seiten, Rang 2 an beiden Seiten, Rang 3 an beiden Seiten, Rang 4: abgelaufen, noch nicht gültig, Schlüssel passt nicht) + beide Pfade leer (Orakel: `TLS` nil, kein Verstoß) + 2 Paare von Verstößen (Rang 2 und Rang 3 an beiden Schlüsseln, Reihenfolge `ADMIN_TLS_CERT` vor `ADMIN_TLS_KEY`) = 12 Zellen; Orakel: Schlüssel und Grund als Literal, Gleichheit der ganzen Liste der Verstöße (die Zeile `ODR-E2001 <Schlüssel>: <Grund>` bildet der Aufrufer, `SPEC-110` Nr. 12) | die Namen `LISTEN_TLS_CERT` und `LISTEN_TLS_KEY` an `ladeTLS` (rot: 12 Zellen); `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` vertauscht (rot: Rang 1, Rang 2, Rang 3, Rang 4); der Pfad im Grund (rot: 11 Zellen); ein Standardpfad für einen leeren Pfad (rot: 1 Zelle, die Datei fehlt und es entsteht ein Verstoß statt kein TLS); `ADMIN_TLS_KEY` vor `ADMIN_TLS_CERT` in der Meldung (rot: 2 Zellen) | andere |
| AT05 | (Nr. 4, 6) Genau ein gesetzter Pfad ist der Verstoß des Rangs 1 am leeren Schlüssel (der gesetzte wird weiter geprüft), und der Zusammenhang entfällt | `TestAdminEinPfad`, Menge M-EinPfad: Seite (nur Zertifikat, nur Schlüssel) × Host (`127.0.0.1:8081`, `0.0.0.0:8081`) × Flag (falsch, wahr) × gesetzte Datei (gültig, ein Verzeichnis) = 16 Zellen; Orakel: gültig: genau die Zeile des Rangs 1, nie eine Zeile an `ADMIN_LISTEN`; Verzeichnis: die zwei Zeilen Rang 2 und Rang 1, sortiert; `Weg` gleich `wegKeiner` | der Zusammenhang prüft `CertPfad == ""` allein statt „kein Pfad gesetzt“ (rot: Schlüssel gesetzt, Zertifikat leer, nicht-loopback, Flag falsch: Zeile an `ADMIN_LISTEN` zusätzlich); der gesetzte Pfad wird nicht weiter geprüft (rot: 8 Zellen mit Verzeichnis) | andere |
| AT06 | (Nr. 6) Der Zusammenhang ist genau eine Zeile `ODR-E2001 ADMIN_LISTEN: außerhalb von Loopback verlangt TLS: ADMIN_TLS_CERT und ADMIN_TLS_KEY`, ohne den Wert von `ADMIN_LISTEN`, ohne Pfad und ohne den Namen des Flags | `TestAdminZusammenhang`, Menge M-Nichtloopback: 5 Hosts (`0.0.0.0:8081`, `[::]:8081`, `:8081`, `recorder.internal:8081`, `10.0.0.5:8081`), kein Pfad, Flag falsch, `FormGueltig` wahr; Orakel: die Liste der Verstöße ist genau der eine Verstoß (`ADMIN_LISTEN`, der feste Grund) als Literal, und kein Host-Text steht in ihm (`strings.Contains` je Zelle) | Schlüssel `ADMIN_TLS_CERT` statt `ADMIN_LISTEN` (rot); der Host im Grund (rot: 5 Zellen); der Grund nennt `ADMIN_ALLOW_PLAINTEXT=true` (rot: Literal); zwei Zeilen (rot) | andere |
| AT07 | (Nr. 6) Bei `FormGueltig` falsch entfällt der Zusammenhang: keine Zeile an `ADMIN_LISTEN`, `Weg` gleich `wegKeiner`, `TLS` nil; die Dateien eines gesetzten Paares werden dennoch geprüft | `TestAdminFormUngueltig`, Menge M-Form: 5 Hosts (wie AT06) × Paar (kein Pfad, gültiges Paar, Paar mit fehlender Datei) × Flag (falsch, wahr) = 30 Zellen; Orakel: kein Verstoß am Schlüssel `ADMIN_LISTEN`, die Zeilen des Paares wie AT04 (bei „fehlende Datei“ genau die eine), `Weg` gleich `wegKeiner` | `FormGueltig` nicht beachtet (rot: 5 Zellen ohne Pfad und mit Flag falsch); die Prüfung des Paares entfällt bei `FormGueltig` falsch (rot: 10 Zellen mit fehlender Datei); `TLS` gesetzt bei `FormGueltig` falsch und gültigem Paar (rot: 10 Zellen) | andere |
| AT08 | (Nr. 9) `warne` schreibt genau dann einen Eintrag, wenn der Weg `wegKlartext` ist: Stufe WARN, Text `start: Admin-Listener ohne TLS erreichbar`, ein Attribut `ereignis` mit `admin_klartext`; `log` nil verwirft | `TestAdminWarnung`, Menge M-Warnung: Weg (4: `wegKeiner`, `wegTLS`, `wegLoopback`, `wegKlartext`) × Log (nil, Handler, der Einträge festhält) = 8 Zellen; Orakel: ein Eintrag mit genau diesen Feldern nur für `wegKlartext` mit Handler, sonst keiner, nie eine Panik | Warnung auch bei `wegLoopback` (rot: 1 Zelle); Stufe INFO (rot); ein zweites Attribut `listen` (rot); zwei Einträge (rot); keine Warnung bei Klartext (rot); Panik bei `log` nil (rot) | andere |
| AT09 | (Nr. 12, 13) Weder Pfad noch Inhalt noch Host noch Eigenschaft des Zertifikats stehen in einer Zeile oder im Eintrag; die Funktionen schreiben nichts auf `stdout`, `stderr` und `log.Default()` und ändern, löschen und verschieben keine Datei, auch bei einem Verstoß nicht | `TestAdminKeinGeheimnis`, Menge M-Geheimnis: Ausgabeweg (Zeilen der Verstöße, Eintrag der Warnung, übrige Ausgabe: `stdout`, `stderr` und `log.Default()` umgeleitet) × Canary (Pfad mit Canary im Verzeichnisnamen, Zeile mit Canary vor dem Block in der Datei, Host `canary-host.example:8081`) = 9 Zellen, dazu der Schnappschuss des Verzeichnisses (Art, Modus, Größe, Bytes) vor und nach jedem der 12 Aufrufe von M-Paar | der Pfad oder `err.Error()` im Grund (rot); der Host im Eintrag der Warnung (rot); `fmt.Println` im Verstoß (rot: übrige Ausgabe); `os.Remove` oder `os.Chmod` bei einem Verstoß (rot: Schnappschuss) | andere |
| AT10 | (Nr. 8 c) `umhuelle` umhüllt nur im Weg `wegTLS` (`tls.NewListener` mit der Konfiguration), sonst liefert sie den übergebenen Listener selbst | `TestAdminUmhuelle`, Menge M-Umhuellung: Weg (4); `wegTLS`: ein Client mit der CA spricht TLS, ein Klartext-Client scheitert; die anderen drei: Ergebnis gleich dem Argument (`==`), ein Klartext-Client gelingt | Umhüllung bei jedem Weg (rot: 3 Zellen); nie umhüllen (rot: 1 Zelle); Umhüllung mit einer leeren `tls.Config` (rot: Handshake) | andere |
| AT11 | (Nr. 7) Die Verbindung folgt `SPEC-291` Nr. 5 und 6: TLS 1.2 und 1.3 gelingen, 1.1 scheitert; ALPN: allein `h2` scheitert, `h2` und `http/1.1` und kein ALPN und `http/1.1` sprechen HTTP/1.1; ein Client ohne Zertifikat gelingt (kein mTLS); die Kette der Datei wird gezeigt | `TestAdminVerbindung`, Menge M-Verbindung: Version (1.1, 1.2, 1.3) + ALPN (4 Angebote) + Client ohne Zertifikat + Kette mit Zwischenzertifikat (`len(PeerCertificates)` gleich 2) = 9 Zellen am Test-Server; Orakel: Literal je Zelle, `ConnectionState().Version` und `NegotiatedProtocol` | `MinVersion` TLS 1.3 (rot: 1.2); `MinVersion` TLS 1.0 (rot: 1.1); `NextProtos` um `h2` ergänzt (rot: 2 Zellen, allein `h2` und `h2` mit `http/1.1`); `ClientAuth` auf `RequireAnyClientCert` (rot); nur das Blatt statt der Kette (rot) | andere |
| AT12 | (Nr. 10 d) Der Rahmen gilt über TLS wie im Klartext: dieselbe Antwort, dieselbe `Anfrage` an die Route, dieselben Einträge und Zähler | `TestAdminRahmenUeberTLS`, Menge M-Anfrage: Anfrage (`GET /healthz` ohne Token, Route mit Token, Route ohne Token, Route mit falschem Token) × Transport (Klartext, TLS) = 8 Zellen; Orakel: Status, Body und Header (ohne den Wert von `X-Correlation-Id`) je Anfrage gleich, die `Anfrage` der Route gleich (Fake hält sie fest), `Audit` mit einem Eintrag `admin_abgewiesen` je 401, `CountError` je Fehler einmal | der Adapter ergänzt bei `r.TLS != nil` einen Header der Anfrage (rot: 4 Zellen mit TLS); `GET /healthz` verlangt über TLS ein Token (rot) | andere |
| AT13 | (Nr. 10 b) Ein Handshake, der scheitert, erreicht den Adapter nicht: kein `Handler`, kein `KorrelationsID`, kein `Append`, kein `CountError`, kein `X-Correlation-Id`, keine Zeile in einer Senke; ein Klartext-Request an den TLS-Port erhält nie die Antwort des Rahmens | `TestAdminHandshakeScheitert`, Menge M-Ausgang: unbekannte CA, Klartext-Request an den TLS-Port, Version 1.1, allein `h2` = 4 Zellen; Orakel: der Client sieht den Fehler (oder die Antwort der Standardbibliothek ohne `X-Correlation-Id` und ohne Fehlerobjekt), alle Zähler des Test-Servers 0, die Senke leer | Rückfall auf Klartext: der umhüllte Listener reicht eine Verbindung mit gescheitertem Handshake als Klartext an den Adapter (rot: Zelle Klartext-Request, der Handler läuft); der `ErrorLog` des Adapters schreibt statt zu verwerfen (rot: die Senke trägt `http: TLS handshake error` in den Zellen unbekannte CA, Version 1.1 und `h2`) | andere |
| AT14 | (Nr. 10 a) Ein Client, der verbindet und nichts sendet, hält weder `Accept` noch einen anderen Client auf | `TestAdminHaengenderHandshake`, Menge M-Haengend: 1 und 8 stumme Verbindungen, offen vor dem zweiten Client; ein Client mit der CA erhält `GET /healthz` 200 innerhalb einer Frist, die kein Schlaf ist (Kanal, `context` des Clients); Orakel: Status 200; deterministisch durch die Reihenfolge (die stumme Verbindung ist angenommen, bevor der Client beginnt); Rate: rot in n von n Prozessläufen | der Handshake im `Accept` (die Hülle von `umhuelle` ruft `Handshake()` vor der Rückgabe): rot in jedem Lauf | nebenläufig |
| AT15 | (Nr. 10 c) Nach `Anhalten` erhält ein Request über TLS 503 `ODR-E5010` mit `Connection: close` und `X-Correlation-Id`, auch wenn sein Handshake erst nach `Anhalten` endet | `TestAdminAnhaltenUeberTLS`, Menge M-Anhalten: (i) eine fertige TLS-Verbindung sendet nach `Anhalten` einen Request, (ii) eine Verbindung, deren Handshake vor `Anhalten` begann und danach endet (Haltepunkt: eine Hülle um das `net.Conn` des Clients hält seinen letzten Teil des Handshakes bis nach der Rückkehr von `Anhalten`, gesteuert durch einen Kanal), dann der Request; Orakel: 503, Code, Header, kein Aufruf der Route | `Anhalten` schließt die offenen Verbindungen des TLS-Listeners (rot: (i) bricht statt 503); die Zulassung wird beim `Accept` statt je Request geprüft (rot: (ii) liefert 200); jeweils rot in jedem Lauf | nebenläufig |
| AT16 | (Nr. 10 c) `Schliessen` schließt eine Verbindung, deren Handshake nicht beendet ist, sofort und wartet nicht auf sie | `TestAdminSchliessenOffenerHandshake`, Menge M-Offen: eine stumme Verbindung, eine mit einem halben ClientHello; `Schliessen(frist)` kehrt mit `nil` vor Ablauf der Frist zurück, die Verbindung ist geschlossen (Lesen liefert einen Fehler, nie einen Timeout des Tests) | `Schliessen` wartet auf offene Verbindungen (rot: `ErrFristAbgelaufen`); `Schliessen` lässt die Verbindung offen (rot: Lesen läuft in den Timeout); jeweils rot in jedem Lauf | nebenläufig |
| AT17 | (Nr. 11) Der Start liest die Dateien einmal: ein Wechsel nach dem Laden ändert das gezeigte Zertifikat nicht (kein Neuladen), ein Neustart zeigt das neue; läuft das Zertifikat im Betrieb ab, bedient der Listener es weiter | `TestAdminKeinNeuladen`, Menge M-Neuladen: Dateien überschrieben nach dem Laden und vor dem ersten Handshake, nach dem ersten Handshake, ein Zertifikat, das im Betrieb abläuft (Uhr des Clients nach dem Ablauf); Orakel: die Seriennummer des ersten Blatts, der Client mit der Wurzel des zweiten scheitert, der zweite Aufruf von `pruefeAdminTransport` (Neustart) zeigt das dritte | `GetCertificate` oder `GetConfigForClient` lädt die Dateien bei jedem Handshake (rot: 3 Zellen) | andere |
| AT18 | (Nr. 13) Die Funktionen halten keinen Zustand; gleichzeitige Aufrufe teilen nichts, jeder Aufruf liefert ein eigenes `*tls.Config`, zwei Paare auf denselben Dateien sind unabhängig | `TestAdminGleichzeitig`: 16 Goroutinen × 20 Läufe `pruefeAdminTransport` auf demselben Paar und je zwei Paare auf denselben Dateien; Orakel: Ergebnisse gleich, die Zeiger der `*tls.Config` paarweise verschieden, ein geänderter Wert des einen ändert den anderen nicht; Detektor von `make test-race` je Prozess (`-test.count=1`) | Paketvariable als Cache der letzten `*tls.Config` (rot: Datenrennen und gleicher Zeiger, in n von n Prozessläufen); `warne` mit gemeinsamem Zähler im Paket (rot: Datenrennen) | nebenläufig |
| AT19 | (Nr. 1) Der Adapter `admin` kennt TLS nicht und bleibt unverändert; kein Adapter importiert den Root | `make a-check` (0 Befunde); der Diff des Slice nennt keinen Pfad unter `src/adapters/driving/admin/` (Review und Verifier lesen ihn) | ein Import von `src/bootstrap` aus dem Paket `admin` (rot, wenn `make a-check` die Kante von `adapters-driving` auf den Composition Root meldet; der Implementer sieht es rot oder nennt das Ergebnis) | andere |
| AT20 | (Nr. 14) Kein neuer Fehlercode, kein Schlüssel im Katalog, Handbuch und Ausnahmen unverändert; die Namen stehen als Konstanten, nicht als Literal in `konfig*.go` | `make katalog-check` und `make handbuch-check` grün ohne Änderung an `docs/user/benutzerhandbuch.md` und `harness/katalog-ausnahmen.txt` | das Literal `"ADMIN_TLS_CERT"` in einer Datei `konfig_admin.go` (rot: `make katalog-check`, Art 4, Name nicht in der Liste von `SPEC-110` Nr. 2; der Implementer sieht es rot oder nennt das Ergebnis) | andere |
| AT21 | (Nr. 2, akzeptiertes Negativ) Der Name `localhost` wird nicht aufgelöst: zeigt er auf dem Host auf eine Adresse außerhalb von Loopback, bindet der Listener dort ohne TLS | kein Fänger. Haltepunkt gesucht: `adminLoopback` hat kein Argument, das der Test übergibt, außer dem Text, und eine Auflösung gehört nicht zum Vertrag. Grund: der Start löst keinen Namen auf (`SPEC-110` Nr. 7), eine Auflösung hinge am Netz des Starts. Folge: der Betreiber des Hosts trägt `/etc/hosts` und die Auflösung. Maß: 1 von 3 Hosts der Menge hängt an einem Namen | nicht anwendbar | andere |
| AT22 | (Nr. 3, akzeptiertes Negativ) Der Weg gilt in jedem Modus; die Funktionen kennen keinen Modus und haben keinen Fänger je Modus | kein Fänger in diesem Slice. Grund: die Funktionen nehmen weder `Konfiguration` noch Modus; je Modus zeigt erst der Aufruf im Start den Weg. Folge: ein Root, der die Prüfung nur in einem Modus ruft, bleibt hier grün. Maß: 4 Modi von 4 ohne Fänger hier; der Fänger ist die Zelle Modus × Weg im Test des Starts bei `slice-bootstrap-m6` (Sendung, §6) | nicht anwendbar | andere |
| AT23 | (Nr. 5, 8) Lesen und Aufrufe: `ADMIN_ALLOW_PLAINTEXT` nimmt genau `true` und `false`, ungesetzt und leer sind `false`, jeder andere Wert ist ein Verstoß am Schlüssel; Laden, `Build`, `Hochfahren` und `Start` rufen die Funktionen in der Reihenfolge von Nr. 8 (Verstoß vor `DATA_DIR` und vor der Token-Datei, Warnung einmal je Start nach `Build`) | kein Fänger in diesem Slice: das Lesen und der Aufruf gehören `slice-bootstrap-m6` (Sendung, §6); die Menge der Werte steht hier: `true`, `false`, `TRUE`, `True`, `1`, `yes`, `on`, leer, nur Leerraum, `true` mit Leerraum am Rand = 10 Werte mit Orakel (wahr: `true` und `true` mit Leerraum am Rand; falsch: `false`, leer, nur Leerraum; ungültig: `TRUE`, `True`, `1`, `yes`, `on`) | nicht anwendbar | andere |

**Zeilen:** AT01 bis AT23 gleich 23; *a* = 4 (AT14, AT15, AT16, AT18), *b* = 19 (AT21 bis AT23 sind Negative oder Sendungen und zählen als `andere`, `SPEC-128` Nr. 3 e).

**Mutationen und Raten** (Implementer; `AGENTS.md` §3.10, §3.15, §3.17): *wird beim Handoff gefüllt.* Je Zeile die Mutationen und der Test, der aus dem benannten Grund rot wurde; die Raten der Klasse `nebenläufig` (AT14, AT15, AT16, AT18) je Mutation mit Zahl der Prozesse, Messendem, Zahl der Prozessoren und Last. AT14 bis AT16 sind durch die Reihenfolge deterministisch gebaut (die Verbindung steht vor dem zweiten Schritt, ein Kanal ordnet), der Bericht nennt n von n; AT18 zählt der Detektor je Prozess (`-test.count=1`, `make test-race`). Der Implementer trägt die Tabelle und die Raten in einen Absatz dieser Stelle ein (`AGENTS.md` §3.10), denn Reviewer und Verifier haben den Bericht nicht.

**Pflicht-Mengen mit Zählprüfung** (Architect, 2026-11-23; Orakel als Literal, nie aus dem Code und nie als Aufruf der Funktion; jede Menge nennt ihre Achsen, `AGENTS.md` §3.15; die Zahl der Zellen steht im Test als Literal und wird gegen die gezählten Zellen verglichen). Der Modus (`record`, `replay`, `mock`, `hybrid`) ist bei keiner Menge eine Achse: die Funktionen kennen keinen Modus (AT22).

| Menge | Zeile | Achsen | Elemente |
|---|---|---|---|
| M-Host | AT01 | Art des Hosts (Name, IPv4, IPv6 in Klammern, Wildcard, leer) · Stelle der Abweichung (Schreibweise, Präfix, Suffix, Bereich `127.0.0.0/8`, Klammerform) · Port (1, 8081, 65535) · Form (mit und ohne `:`) | loopback: `127.0.0.1`, `[::1]`, `localhost`; nicht-loopback (17): `0.0.0.0`, `[::]`, leer, `127.0.0.2`, `127.255.255.255`, `LOCALHOST`, `Localhost`, `localhost.`, `xlocalhost`, `localhostx`, `[0:0:0:0:0:0:0:1]`, `[::ffff:127.0.0.1]`, `[::2]`, `::1`, `10.0.0.1`, `recorder.internal`, `127.0.0.1.`; je Host 3 Ports = 60 Zellen, dazu 3 Eingaben ohne `:` (`127.0.0.1`, `localhost`, leer) = 63 |
| M-Nachbar | AT01 | Host der drei loopback-Hosts · Art der Änderung (Einfügen, Löschen, Ersetzen eines Zeichens, andere Groß-/Kleinschreibung) · Zeichen (68: `0-9`, `a-z`, `A-Z`, `.`, `:`, `[`, `]`, `-`, `_`) | alle verschiedenen Nachbarn je Host ohne den Host selbst, dazu die 511 anderen Schreibweisen von `localhost`; Orakel falsch; die Zahl je Host steht im Test als Literal, vom Implementer mit einem zweiten Verfahren (Formel statt Generator) bestimmt |
| M-Weg | AT02 | Host (8: 3 loopback, 5 nicht-loopback: `0.0.0.0:8081`, `[::]:8081`, `:8081`, `recorder.internal:8081`, `10.0.0.1:8081`) · TLS (gültiges Paar, kein Paar) · Flag (falsch, wahr) | 8 × 2 × 2 = 32 |
| M-Rueckfall | AT03 | Art des Verstoßes (die 9 einzelnen von M-Paar) · Host (`0.0.0.0:8081`, `127.0.0.1:8081`) · Flag (wahr) | 9 × 2 = 18 |
| M-Paar | AT04 | Schlüssel (Zertifikat, Schlüssel) · Rang (1 bis 4) · Art im Rang 4 (abgelaufen, noch nicht gültig, passt nicht) · Zahl der Verstöße (0, 1, 2) | 9 einzelne (Rang 1: 2, Rang 2: 2, Rang 3: 2, Rang 4: 3) + beide Pfade leer + 2 Paare (Rang 2 an beiden, Rang 3 an beiden) = 12 |
| M-EinPfad | AT05 | Seite des gesetzten Pfades (Zertifikat, Schlüssel) · Host (`127.0.0.1:8081`, `0.0.0.0:8081`) · Flag (falsch, wahr) · gesetzte Datei (gültig, ein Verzeichnis) | 2 × 2 × 2 × 2 = 16 |
| M-Nichtloopback | AT06 | Art des nicht-loopback-Hosts (Wildcard IPv4, Wildcard IPv6, leer, Name, IP-Adresse) | 5 |
| M-Form | AT07 | Host (die 5 von M-Nichtloopback) · Paar (kein Pfad, gültiges Paar, fehlende Datei) · Flag (falsch, wahr) | 5 × 3 × 2 = 30 |
| M-Warnung | AT08 | Weg (`wegKeiner`, `wegTLS`, `wegLoopback`, `wegKlartext`) · Log (nil, Handler) | 4 × 2 = 8 |
| M-Geheimnis | AT09 | Ausgabeweg (Zeilen der Verstöße, Eintrag der Warnung, übrige Ausgabe) · Canary (Pfad, Inhalt, Host) | 3 × 3 = 9; dazu 12 Schnappschüsse des Verzeichnisses (je Zelle von M-Paar) |
| M-Umhuellung | AT10 | Weg (4) | 4 |
| M-Verbindung | AT11 | Version (1.1, 1.2, 1.3) · ALPN (allein `h2`, `h2` und `http/1.1`, keines, `http/1.1`) · Client ohne Zertifikat · Kette mit Zwischenzertifikat | 3 + 4 + 1 + 1 = 9 |
| M-Anfrage | AT12 | Anfrage (`GET /healthz` ohne Token, Route mit Token, Route ohne Token, Route mit falschem Token) · Transport (Klartext, TLS) | 4 × 2 = 8 |
| M-Ausgang | AT13 | Art des Scheiterns (unbekannte CA, Klartext an den TLS-Port, Version 1.1, allein `h2`) | 4 |
| M-Haengend | AT14 | Zahl der stummen Verbindungen (1, 8) | 2 |
| M-Anhalten | AT15 | Zeitpunkt des Handshakes (fertig vor `Anhalten`, endet nach `Anhalten`) | 2 |
| M-Offen | AT16 | Zustand der Verbindung (stumm, halbes ClientHello) | 2 |
| M-Neuladen | AT17 | Zeitpunkt der Änderung (vor dem ersten Handshake, danach) · Ablauf im Betrieb | 2 + 1 = 3 |
| M-Gleichzeitig | AT18 | 16 Goroutinen × 20 Läufe · Paar (dasselbe, zwei auf denselben Dateien) | je Prozess; Zählprüfung 320 Aufrufe |

**Unbeschränkte Mengen und ihre Prüfgrenze** (akzeptierte Negative; Grund, Folge, Maß; `AGENTS.md` §3.15, §3.17): (1) *Host* (AT01): jede Zeichenkette; Prüfgrenze: die 17 benannten Hosts, jeder Nachbar im Abstand 1 und jede Schreibweise von `localhost`; Grund: der Vergleich ist die Gleichheit mit drei Literalen, und ein Zeichen, das er anders behandelte, zeigt sich im Abstand 1; Folge: ein Vergleich mit einer Ausnahme für eine Zeichenkette im Abstand von mindestens 2 zu allen drei Hosts bliebe grün; Maß: Abstand 1 von unbeschränkt; 17 benannte Hosts in 7 Arten (Wildcard, Bereich 127/8, Schreibweise, Präfix, Suffix, Klammerform, Name). (2) *Zahl der stummen Verbindungen* (AT14): Prüfgrenze 1 und 8; Grund: mehr Verbindungen kosten Laufzeit und ändern die Klasse nicht; Folge: ein Hängen, das erst mit der neunten Verbindung einträte, bliebe grün, die Standardbibliothek hat keine solche Grenze im `Accept`; Maß: 8 von unbeschränkt. (3) *Zahl der gleichzeitigen Aufrufe* (AT18): Prüfgrenze 16 Goroutinen in 20 Läufen je Prozess; Grund und Folge wie bei `slice-token-datei-wettlauf` TW5 (ein Zustand, den erst die siebzehnte Goroutine träfe, bliebe grün; die Funktionen haben keinen Zustand, den eine Zahl der Aufrufer öffnet); Maß: 16 von unbeschränkt. (4) *Formen der Dateien und des Zertifikats*: Prüfgrenzen und Folgen von `SPEC-291` (*Ungeprüft*, Zeilen A01 bis A17 und die Mengen von `slice-listener-tls`); dieser Slice ändert sie nicht, AT04 führt je Rang einen Fall mit den Namen `ADMIN_TLS_*`. AT21 bis AT23 stehen in der Tabelle.

**Größe und Schnitt** (Architect, 2026-11-23; `SPEC-128` Nr. 1, `AGENTS.md` §3.14; der Planner setzt die Marke beim Start): *a* = 4, *b* = 19, *f* = 1 (der Test-Server mit zählenden `IDs`, `Audit` und `Telemetrie` aus `supporttest`), *p* = 120 (`admintls.go`: Typen und Konstanten rund 25 Zeilen, `adminLoopback` 12, `pruefeAdminTransport` 55, `umhuelle` 8, `warne` 12, Kommentare), also *S* = 85·4 + 80·19 + 200·1 + 120 = 340 + 1520 + 200 + 120 = 2180; 135·2180 = 294300, *M* = 100 · ⌈29,43⌉ = **3000**; keine Schnittfrage (*S* = 2180 < 2223, 135·2180 ≤ 300000), aber 43 unter der Schnittgrenze: jede weitere Zeile `andere` kostet 80 und löst die Schnittfrage aus (*S* = 2260). **Vorab benannter Schnitt** (`AGENTS.md` §3.14, für den Fall der Rückführung oder wachsender Tabelle): Teil *Betrieb* = `admintls_betrieb_test.go`, AT11 bis AT17, zuletzt in eigenem Commit geschrieben und davor gemessen (*a* = 3, *b* = 4, *f* = 1, *p* = 0, *S* = 775, *M* = 1100); der Rest, AT01 bis AT10 und AT18 bis AT23, mit *a* = 1, *b* = 15, *f* = 0, *p* = 120, *S* = 1405, *M* = 1900. Beide Teile teilen keinen Mechanismus außer dem Aufruf von `pruefeAdminTransport` und `umhuelle` (Code des Rests, vom Teil *Betrieb* nur gerufen); die Summe der zwei Schätzungen gleicht *S* (775 + 1405 = 2180).

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Prüfung der Sendungen** (`AGENTS.md` §3.13; Planner am 2026-10-30 und 2026-11-21, Architect am 2026-11-23): *an diesen Slice* der Baustein aus `slice-listener-tls`: `welle-betrieb` §4 nennt ihn als wiederverwendbar für den Admin-Listener; die Datei von `slice-listener-tls` liegt in `done/` und trägt die Lieferung (Teil 1 des Schnitts, §1, DoD und Zeile A17 dort, kein Ausschluss trifft die Sendung); der Architect hat sie am 2026-11-23 gegen `src/bootstrap/tlsdatei.go` gelesen: `ladeTLS` ist ohne Änderung nutzbar. *Von diesem Slice an `slice-bootstrap-m6`* (gelesen am 2026-11-23: §1 *Ausdrücklich NICHT* nennt `HYBRID_*`, `LISTEN_TLS_*`, neue Endpunkte und das Umschalten des Modus und trifft keine der Sendungen; seine DoD trägt Verdrahtung, `ADMIN_*` bis in den TLS-Baustein und das Handbuch-Kapitel): (1) das Lesen von `ADMIN_LISTEN`, `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY` und `ADMIN_ALLOW_PLAINTEXT` samt dem Aufheben von `SPEC-110` Nr. 2 trägt sein §1 *Übernimmt* schon (mit der Kennung dieses Slice); (2) neu und von ihm noch nicht genannt: die vier Aufrufe von `SPEC-299` Nr. 8 (Laden, `Build`, `Hochfahren`, `Start`), die Warnung am Start (Nr. 9), die Reihenfolge Verstoß vor `DATA_DIR` und vor der Token-Datei (Nr. 8 a), die Zelle Modus × Weg (AT22), die Menge der zehn Werte des Flags (AT23), der Nachzug von `SPEC-110` Nr. 10 und 11 mit Zeigern auf `SPEC-299` und das Handbuch (siehe *Handbuch und Katalog* oben). Nach Nr. 4 der Übergangsregeln schreibt dieser Zweig die Datei von `slice-bootstrap-m6` nicht; die Sendung steht in §6 mit Ausgang *offen* und Adresse Planner, der sie in §1 und DoD von `slice-bootstrap-m6` einträgt.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token` samt seinen Teilen `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` (Schnitt vor dem Code, 2026-11-07; TLS umhüllt den Listener, den `Serve` aus `slice-admin-lebenszyklus` bedient) und `slice-listener-tls` (`welle-betrieb`, Teil 1 des Schnitts vom 2026-11-21: der Baustein `ladeTLS`; `slice-listener-tls-verdrahtung` ist keine Bedingung) liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** der Listener, den TLS umhüllt, kommt aus `slice-admin-listener-token` (*Code-Abhängigkeit*); den wiederverwendbaren Baustein zum Laden und Prüfen des Zertifikatspaars liefert `slice-listener-tls` (*Code-Abhängigkeit*; er ist als wiederverwendbarer Baustein geschnitten, §6).

**Unabhängig von `welle-mock`, M2 und `slice-bootstrap-m6`** — bis auf das, was `slice-listener-tls` selbst voraussetzt; sein Plan nennt es (der Baustein braucht nur die Standardbibliothek). Der Slice braucht den Audit-Port nicht (die Warnung geht ins Log, nicht ins Audit).

**Stand der Bedingungen** (Architect, 2026-11-23, an den Verzeichnissen gelesen, `main` = `bad9b892`; der Planner prüft sie beim Start neu): `slice-admin-listener-token` und seine fünf Teile `slice-admin-rahmen`, `slice-admin-authentisierung`, `slice-admin-lebenszyklus`, `slice-admin-fehlerabbildung` und `slice-admin-abweisen-audit` sowie `slice-listener-tls` liegen in `done/`; der Vertrag (`SPEC-299`), die Tabelle und das Raster stehen in diesem Zweig (`slice-admin-tls-vertrag`) und gehen mit ihm nach `main`. Die Bedingung ist erfüllt, sobald dieser Zweig auf `main` liegt.

**Reihenfolge zu den Slices der Welle und Kollisionen** (Architect, 2026-11-23): (1) **Vor** `slice-bootstrap-m6` und `slice-test-security`: beide nennen diesen Slice in ihrem Start-Trigger (*Code-Abhängigkeit*: `slice-bootstrap-m6` ruft die Funktionen, `slice-test-security` belegt „nur über Loopback oder mit TLS erreichbar“); dieser Slice braucht keinen von beiden. Die Funktionen sind im Paket `bootstrap` unexportiert (Naht in `export_test.go`, wie `ladeTLS`): ein Test im Paket `tests/` erreicht sie nicht; wie `slice-test-security` den Satz „ein nicht-loopback `ADMIN_LISTEN` ohne TLS und ohne Flag startet nicht“ vor `slice-bootstrap-m6` belegt, ist offen und steht in §6 (Adresse Planner). (2) **Laufende Slices** (`in-progress/`, 2026-11-23): `slice-token-datei-wettlauf` (`src/adapters/driving/admin/tokendatei*`, `spec/spezifikation.md` `SPEC-239`) und `slice-admin-imports-audit` (`src/adapters/driving/admin/imports*`, `src/bootstrap/audit_betrieb_test.go`, `src/hexagon/ports/support/audit*`, `SPEC-297`): keine Datei gemeinsam mit dem Block unten; die Zeilen von `spec/spezifikation.md`, die dieser Zweig ändert (`SPEC-299` neu, Zeiger in `LH-FA-13.a`, `LH-RB-02.a`, `SPEC-238` *Ungeprüft*, `SPEC-291` Kopf, `SPEC-089` und `SPEC-090`), liegen nicht in den Abschnitten `SPEC-239` und `SPEC-297`. (3) **`src/bootstrap/wurzel.go`, `start.go` und `konfig*.go`** ändert dieser Slice nicht (neue Dateien und die Naht `export_test.go`); `slice-bootstrap-m6`, `slice-bootstrap-m4-konfig` und `slice-bootstrap-m4` ändern sie und laufen nacheinander, nie zugleich (`welle-admin-api` §5); mit diesem Slice teilen sie keine Datei außer `export_test.go`, das keiner der laufenden Slices berührt; der Planner prüft vor dem Start, ob ein anderer Zweig `export_test.go` ändert (`SPEC-128` Nr. 10 g, `BEO-REPO/bedingter-blockpfad-zaehlt-fremde-zeilen`). (4) **Gate-Kette** (Nr. 7 der Übergangsregeln): der Plan berührt `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md` und `tools/gegenproben/` nicht; kein anderer Slice der Kette in `in-progress/` ist gefordert. (5) **SPEC-Kennung:** `SPEC-299` ist die letzte Kennung des zweiten Bereichs der Linie B (260 bis 299, Übergangsregeln Nr. 6); der nächste Vertrag braucht einen neuen Bereich (§6, Adresse Planner).

**Größenmarke** (`SPEC-128` Nr. 1; Teil des Architect 2026-11-23): *S* = 2180 gleich der Tabelle (23 Zeilen: *a* = 4, *b* = 19; *f* = 1; *p* = 120, Rechenweg in §3 *Größe und Schnitt*), **Marke 3000** gleich der Formel (135·2180 = 294300, ⌈29,43⌉ = 30); keine Schnittfrage, 43 unter der Grenze; der Schnitt, falls nötig, steht vorab in §3 (Teil *Betrieb*). Der Block trägt den Teil des Architect (`pfad`, `produktcode`, `fakes`); `basis`, `schaetzung` (2180) und `marke` (3000) setzt der Planner beim Start vor dem ersten Code-Commit (`AGENTS.md` §3.14) und rechnet dabei die Tabelle neu. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen; die Pfade nennen die Dateien einzeln, damit fremde Zeilen im Verzeichnis nicht mitzählen. Legt der Implementer eine weitere Datei an oder ändert er eine weitere Datei, nennt er sie im selben Commit im Block.

```groessenmarke
pfad: src/bootstrap/admintls.go
pfad: src/bootstrap/admintls_test.go
pfad: src/bootstrap/admintls_betrieb_test.go
pfad: src/bootstrap/export_test.go
produktcode: 120
fakes: 1
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen). Der vorab benannte Schnitt (Teil *Betrieb*, §3) ist der Weg, den ein Handoff über der Marke gehen darf; ein anderer Teil wird beim Handoff nicht geschnitten.
- `in-progress` → `open` (blockiert — Carveout?): der Baustein aus `slice-listener-tls` lässt sich für den Admin-Listener nicht ohne Änderung nutzen, oder der Implementer bemerkt eine Randform ohne Ort in `SPEC-299` (`AGENTS.md` §3.12); beides geht mit Datum und Adresse Architect nach §6, der Implementer entscheidet es nicht im Code.

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Architect, 2026-11-23): jede Zelle nennt ihren Ort in `SPEC-299` (oder in `SPEC-291`, `SPEC-110`), oder „gilt nicht“ mit Grund; keine Zelle ist leer. „Abbruch“ gilt für die Funktionen nicht, weil sie synchron ohne Kontext laufen (`SPEC-110` Nr. 13, `SPEC-299` Nr. 13); der Handshake kennt ihn.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Host von `ADMIN_LISTEN` klassifizieren (`adminLoopback`) | `SPEC-299` Nr. 2: ungesetzt gilt der Default `127.0.0.1:8081` (Loopback), die Funktion erhält den wirksamen Wert | `SPEC-299` Nr. 2: die leere Eingabe, die Eingabe ohne `:` und der leere Host (`:8081`) sind nicht-loopback | `SPEC-299` Nr. 2: jede andere Schreibweise und jeder andere Host (Wildcard, `127.0.0.2`, Name, `LOCALHOST`) ist nicht-loopback; kein Name wird aufgelöst | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-299` Nr. 13: dieselbe Eingabe, dasselbe Ergebnis |
| `ADMIN_ALLOW_PLAINTEXT` lesen (`Klartext`, `FormGueltig`) | `SPEC-299` Nr. 5: ungesetzt ist `false` | `SPEC-299` Nr. 5: leer und nur Leerraum sind `false` (`SPEC-110` Nr. 1) | `SPEC-299` Nr. 5: jeder Wert außer `true` und `false` ist der Verstoß der Form, `FormGueltig` falsch (Nr. 6) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-299` Nr. 13: dasselbe Ergebnis |
| Weg und Zusammenhang bestimmen (`pruefeAdminTransport`, ohne die Dateien) | `SPEC-299` Nr. 3 und 6: kein Paar und kein Flag ist bei Loopback `wegLoopback`, sonst der Verstoß an `ADMIN_LISTEN` | `SPEC-299` Nr. 4: ein Pfad nur aus Leerraum gilt als ungesetzt (`SPEC-110` Nr. 1) | `SPEC-299` Nr. 6: `FormGueltig` falsch und genau ein gesetzter Pfad lassen den Zusammenhang entfallen; Nr. 3: ein Paar mit Verstoß liefert keinen Weg und nie Klartext | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-299` Nr. 13: dasselbe Ergebnis für dieselbe Eingabe und Zeit |
| Zertifikatspaar laden (`ladeTLS` mit `ADMIN_TLS_*`) | `SPEC-299` Nr. 4 und `SPEC-291` Nr. 2: ein Pfad fehlt (Rang 1), eine Datei fehlt (Rang 2) | `SPEC-299` Nr. 4 und `SPEC-291` Nr. 2: beide Pfade leer ist kein TLS, eine leere Datei ist Rang 3 | `SPEC-299` Nr. 4 und `SPEC-291` Nr. 2 bis 4: kein PEM, ein Schlüssel zu einem anderen Zertifikat, außerhalb der Gültigkeit (Ränge 3 und 4) | gilt nicht: synchron ohne Kontext (`SPEC-110` Nr. 13) | `SPEC-299` Nr. 13 und `SPEC-291` Nr. 12: zustandslos, ein zweiter Aufruf ist unabhängig |
| Listener umhüllen (`umhuelle`) | gilt nicht: nur ein Transport mit Weg `wegTLS` trägt ein `*tls.Config`, ein Paar mit Verstoß liefert keinen Weg (`SPEC-299` Nr. 3); der Root übergibt den Listener aus `net.Listen` (Nr. 8 c) | gilt nicht: wie links, der Root übergibt nie einen leeren Listener (Nr. 8 c) | `SPEC-299` Nr. 8 c: nur `wegTLS` umhüllt, jeder andere Weg liefert den Listener selbst | gilt nicht: kein Kontext, kein Blockieren; der Handshake läuft in der Verbindung (Nr. 10 a) | `SPEC-299` Nr. 8 c und `SPEC-139` Nr. 6: ein zweites `Hochfahren` ist `model.ErrUngueltig`, `umhuelle` läuft einmal je `Handle` |
| HTTPS bedienen (Handshake und Anfrage) | `SPEC-299` Nr. 7: ein Client ohne Zertifikat gelingt (kein mTLS), ohne ALPN gilt HTTP/1.1 | `SPEC-299` Nr. 10 a und b: ein Client, der nichts sendet, hält nichts auf; ein Klartext-Request an den TLS-Port erreicht den Rahmen nie | `SPEC-299` Nr. 7 und 10 b: unbekannte CA, Version unter 1.2 und allein `h2` scheitern im Handshake und erreichen den Adapter nicht | `SPEC-299` Nr. 10 a und c: `Anhalten` und `Schliessen` über TLS wie `SPEC-238` Nr. 11 | `SPEC-299` Nr. 11 und 10 d: kein Neuladen der Dateien, dieselbe `Anfrage` bei jedem Handshake |
| Warnung bei Klartext (`warne`) | `SPEC-299` Nr. 9: `log` nil verwirft den Eintrag | `SPEC-299` Nr. 9: bei `wegKeiner`, `wegTLS` und `wegLoopback` kein Eintrag | `SPEC-299` Nr. 9: bei `wegKlartext` ein Eintrag mit festem Text, fester Stufe und dem einen Attribut | gilt nicht: einmaliger Eintrag beim Start | `SPEC-299` Nr. 9 und 13: ein Aufruf, ein Eintrag; ein zweiter Start schreibt seinen eigenen |

**Zusatztabelle** (`AGENTS.md` §3.16: je Paar von Verstößen am selben Schlüssel und je Modus):

| Paar oder Fall | Entscheidung | Ort | Fänger |
|---|---|---|---|
| `ADMIN_LISTEN`: Form × Zusammenhang | schließen einander aus: bei ungültiger Form entfällt der Zusammenhang | `SPEC-299` Nr. 6 | AT07 (30 Zellen), AT06 |
| `ADMIN_ALLOW_PLAINTEXT`: Form × Zusammenhang | der Zusammenhang meldet nie an diesem Schlüssel, die Form entfällt nie; ein ungültiger Wert gilt weder als wahr noch als falsch | `SPEC-299` Nr. 5, 6 | AT07 (die Zellen mit Flag), AT23 (Lesen: Sendung) |
| `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY`: Rang × Rang am selben Schlüssel | höchstens ein Grund je Schlüssel, der erste Rang gilt | `SPEC-299` Nr. 4, `SPEC-291` Nr. 2 | AT04 (12 Zellen, darunter 2 Paare) und die Zeilen A01 bis A17 von `slice-listener-tls` |
| Verstoß des Paares × Zusammenhang | schließen einander aus: der Zusammenhang gilt bei keinem gesetzten Pfad, die Verstöße des Paares bei mindestens einem; bei genau einem Pfad nur Rang 1 | `SPEC-299` Nr. 6 | AT05 (16 Zellen), AT06 (5 Zellen) |
| Verstoß des Paares × Flag wahr | kein Rückfall auf Klartext, auch nicht bei Loopback | `SPEC-299` Nr. 3 | AT03 (18 Zellen) |
| `ADMIN_TLS_*` × `LISTEN_TLS_*` zugleich | eine Meldung, vereinigt und nach Schlüssel sortiert | `SPEC-299` Nr. 8 b | Sendung an `slice-bootstrap-m6` (siehe unten, AT23) |
| je Modus (`record`, `replay`, `mock`, `hybrid`) | der Weg gilt in jedem Modus | `SPEC-299` Nr. 3 | AT22 (akzeptiertes Negativ), die Zelle Modus × Weg bei `slice-bootstrap-m6` |

**Risiken und Ausgänge:**

- Die Definition von Loopback und der Umgang mit Namen, die nicht aufgelöst werden (Auftrag 1) — **Ausgang:** entschieden (Architect, 2026-11-23): `SPEC-299` Nr. 2; AT01 und AT21.
- Mindestversion und Cipher von TLS (Auftrag 3) — **Ausgang:** entschieden (Architect, 2026-11-23): wie `SPEC-291` Nr. 5 und 6, keine eigene Entscheidung (`SPEC-299` Nr. 7); AT11.
- Der Baustein aus `slice-listener-tls` muss die Wiederverwendung tragen, sonst ist die Übernahme keine Adresse (`AGENTS.md` §3.13) — **Ausgang:** entschieden (Planner, 2026-11-21; Architect, 2026-11-23, gegen `tlsdatei.go` gelesen): der Plan von `slice-listener-tls` (Teil 1 des Schnitts) trägt sie (§1 Ziel, DoD, Zeile A17 mit `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY`, `SPEC-291` Kopf); `ladeTLS` ist ohne Änderung nutzbar, `SPEC-299` Nr. 4; die Bedingung dieses Slice nennt Teil 1 und nicht die Verdrahtung des OData-Listeners.
- Zertifikate und Schlüssel der Tests dürfen nicht im Repo liegen (`harness/README.md` §Safety and scope boundaries) — **Ausgang:** entfallen: Die Tests erzeugen das Paar zur Laufzeit (`tlstest`).
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): die Zweige der Start-Prüfung sind eine Matrix; jede Zelle hat einen Lauf (AT02: 32 Zellen, AT03: 18, AT07: 30) — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- Größe: Marke 3000 (*S* = 2180, 43 unter der Schnittgrenze, vorab benannter Schnitt in §3) — **Ausgang:** offen; Adresse: Planner beim Start (§4, setzt `basis`, `schaetzung`, `marke` und prüft `export_test.go` auf fremde Änderungen).
- **Sendung an `slice-bootstrap-m6`** (`AGENTS.md` §3.13, Übergangsregeln Nr. 4: dieser Zweig schreibt dessen Datei nicht): die vier Aufrufe von `SPEC-299` Nr. 8, die Warnung am Start, die Reihenfolge Verstoß vor `DATA_DIR` und Token-Datei, die Zelle Modus × Weg (AT22), die Menge der zehn Werte des Flags (AT23), die Vereinigung der Verstöße beider Paare in `Build`, der Nachzug von `SPEC-110` Nr. 10 und 11 mit Zeigern auf `SPEC-299`, Handbuch und Katalog (§3 *Handbuch und Katalog*) — **Ausgang:** offen; Adresse: Planner (trägt die Sendung in §1 und DoD von `slice-bootstrap-m6` ein; §1 *Ausdrücklich NICHT* und DoD dort sind am 2026-11-23 gelesen und treffen sie nicht).
- **`slice-test-security` und die Funktionen:** „ein nicht-loopback `ADMIN_LISTEN` ohne TLS und ohne Flag startet nicht“ braucht einen Zugang zu `pruefeAdminTransport`, bevor `slice-bootstrap-m6` den Start verdrahtet; die Funktionen sind unexportiert, das Paket `tests/` erreicht sie nicht, und `slice-test-security` nennt `slice-bootstrap-m6` nicht in seinem Trigger — **Ausgang:** offen; Adresse: Planner (entweder liegt der Test im Paket `bootstrap` über `export_test.go`, oder `slice-test-security` hängt am Start aus `slice-bootstrap-m6`).
- **SPEC-Kennung:** `SPEC-299` ist die letzte Kennung des zweiten Bereichs der Linie B (260 bis 299, Übergangsregeln Nr. 6) — **Ausgang:** offen; Adresse: Planner (ein neuer Vertrag, etwa für `slice-bootstrap-m6`, braucht einen Bereich).
- AT19 und AT20 haben Mutationen, die ein Gate melden soll (`make a-check`, `make katalog-check`); ob beide rot werden, ist nicht gemessen — **Ausgang:** offen; Adresse: Implementer (sieht sie rot oder nennt das Ergebnis im Bericht, `AGENTS.md` §3.10).
- Die Begrenzung der Fehlversuche am Admin-Listener (`LH-FA-13.a` nennt keine) — **Ausgang:** weiter offen, nicht Gegenstand dieses Slice (`SPEC-299` *Ungeprüft*); Adresse: User (Rückgabe R5 des Architect, `welle-admin-api` §5).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×, Matrix der Start-Prüfung) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
