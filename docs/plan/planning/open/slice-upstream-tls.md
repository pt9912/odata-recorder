# Slice slice-upstream-tls: TLS-Konfiguration des Upstream-Adapters `httpupstream`

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** [`LH-RB-02`](../../../../spec/lastenheft.md#lh-rb-02--http-schnittstelle) (HTTPS und Zertifikatsfehler), [`LH-FA-01`](../../../../spec/lastenheft.md#lh-fa-01--transparente-weiterleitung-muss), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, nur Standardbibliothek: `crypto/tls`, `crypto/x509`), [ADR-0003](../../adr/0003-lint-profil-golangci.md) (Lint-Profil)

**Berührte Spec-Stellen:** `SPEC-120` Nr. 1 (`CAPEM` und seine Leseregel), Nr. 7 (TLS, HTTP/1.1 auch über TLS), Nr. 8 (ein Fehler der Prüfung ist `ErrUpstreamNichtErreichbar`, die Allaussage über die TLS-Fälle) und *Ungeprüft* (Inhalt der Vertrauensanker des Systems, TLS-Versionen, HTTP/2) · `SPEC-110` Nr. 10, Nr. 11 (dieselbe Leseregel am Start für `UPSTREAM_CA_FILE`) · `SPEC-086` · `LH-RB-02.a` · `ARC-008`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-21.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Adapter `httpupstream` wendet `CAPEM` und `Pruefen` nach `SPEC-120` Nr. 1 und 7 an: `New` liest `CAPEM` nach der Leseregel aus Nr. 1 (auch bei `Pruefen=false`), bei `https` prüft der Client Kette und Hostnamen gegen die Anker des Systems und `CAPEM` (ergänzt, ersetzt nicht), `Pruefen=false` prüft nichts, jeder Fehler der Prüfung ist `ErrUpstreamNichtErreichbar`, und der Client spricht HTTP/1.1 auch über TLS. Sein Lieferwert ist der Nachweis zu `LH-RB-02` für HTTPS und Zertifikatsfehler auf der Ebene des Adapters und die Wirkung der Schalter, die `slice-bootstrap-m2` verdrahtet.

**Übernimmt:** `slice-upstream-adapter`, Schnitt beim Handoff (dort §4, vorab benannt): der TLS-Teil, wie ihn der Commit `0716718` trägt — `tls.go` (`tlsTransport`, `leseAnker`), der Aufruf in `New` (`client.go`), `tls_test.go` (`TestMain` mit S als einzigem Anker des Systems, A10 `TestTLS`, die `CAPEM`-Fälle von A1 `TestNewCAPEM`, der TLS-Fall von A11 `TestTransportTLS`, die TLS-Fälle von A14 `tlsFehlerFaelle`) und die Zeile in `alleFehlerFaelle` (`fehler_test.go`); dazu der Fund des Implementers aus dem Handoff, der im TLS-Teil keinen Ort hat (§6, *Block `CERTIFICATE` mit PEM-Headern*). Code und Fänger sind geschrieben, je Zusage nach der Commit-Nachricht von `0716718` mit einer Mutation rot gesehen; `slice-upstream-adapter` nimmt sie mit einem reinen Revert heraus, dieser Slice bringt sie mit `git cherry-pick -x 0716718` zurück (§4). Den Stand dazwischen führt `slice-upstream-adapter` als akzeptiertes Negativ (dort §3; Entscheidung des Architect dort in §4): `SPEC-120` bleibt mit Nr. 1 und 7 der Vertrag, den dieser Slice liefert, und nennt den Stand nicht; der Kommentar von `Konfig` ist dort neutral geworden (er verweist auf `New`) und stimmt nach dem Cherry-Pick weiter, dieser Slice ändert ihn nicht.

**Übergaben an Folge-Slices:** keine neue Sendung. Die Verdrahtung von `httpupstream.New` und die Warnung beim Start bei abgeschalteter Prüfung (`LH-RB-02.a`) sendet `slice-upstream-adapter` an `slice-bootstrap-m2` (dort §1 *Übernimmt (Upstream)*); dieser Slice liefert die Wirkung von `Pruefen` und `CAPEM`, nach der die Warnung fragt, und der Start von `slice-bootstrap-m2` verlangt ihn in `done/` (geprüft: `slice-bootstrap-m2` liegt in `open/`, kein *Ausdrücklich NICHT* dort trifft die Verdrahtung, der erste Punkt seiner DoD nennt Verdrahtung und Warnung).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der übrige Adapter (`SPEC-120` Kopf, Nr. 1 ohne `CAPEM`, Nr. 2 bis 6, 8, 9, kein Proxy aus Nr. 7; Fänger A1 ohne `CAPEM`, A2 bis A9, A11 `TestTransport`, A12 bis A19; A19 samt der Prüfung des Hosts in `normalisiereURL`, `src/bootstrap/konfig_werte.go`) — `slice-upstream-adapter`; dieser Slice ändert ihn nicht.
- Verdrahtung, Lesen von `UPSTREAM_CA_FILE` und die Warnung beim Start — `slice-bootstrap-m2`; der Adapter erhält Bytes und Schalter, keine Datei und keinen Logger. Ausnahme, nur wenn die Entscheidung zum Fund in §6 die Leseregel aus `SPEC-120` Nr. 1 ändert: dann zieht dieser Slice die Prüfung am Start (`SPEC-110` Nr. 11, `src/bootstrap/konfig_dienste.go`) mit, denn Nr. 1 nennt sie dieselbe Regel, und zwei Slices ließen ihre zwei Leser dazwischen auseinanderlaufen.
- Die TLS-Versionen und Verfahren der Bibliothek, HTTP/2, der Inhalt der Vertrauensanker des Systems im Image — akzeptierte Negative in `SPEC-120` *Ungeprüft*, je mit Grund und Folge; kein Fänger und keine Abnahme verlangt sie.
- TLS des Listeners (`LISTEN_TLS_CERT`, `LISTEN_TLS_KEY`) — Folgewelle (Welle §6).

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

- [ ] `CAPEM` nach `SPEC-120` Nr. 1: `New` liefert `ErrUngueltig` und keinen Client bei einem nichtleeren `CAPEM` ohne lesbares Zertifikat, auch bei `Pruefen=false`; Text, Blöcke anderer Art und ungültige Blöcke neben einem lesbaren Zertifikat bleiben ohne Wirkung; der Block `CERTIFICATE` mit PEM-Headern nach der Entscheidung des Architect (§6); ändert sie die Leseregel, folgt die Prüfung am Start (`SPEC-110` Nr. 11) im selben Slice (Fänger A1 `CAPEM`).
- [ ] TLS und Transport nach `SPEC-120` Nr. 7 und 8: Kette und Hostname gegen die Anker des Systems und `CAPEM` (ergänzt, ersetzt nicht), `Pruefen=false` prüft nichts, jeder Fehler der Prüfung (unbekannte Stelle, Hostname, abgelaufen) ist `ErrUpstreamNichtErreichbar` ohne Werte im Text, HTTP/1.1 auch über TLS — der Nachweis zu `LH-RB-02` für HTTPS und Zertifikatsfehler (Fänger A10, A11 TLS, A14 TLS).
- [ ] Der erste Code-Commit ist `git cherry-pick -x 0716718`, sonst nichts; jede Änderung danach steht in eigenem Commit, und der Bericht des Handoffs nennt beide Messungen (§4).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`SPEC-120` Nr. 1, 7: Nachzug nach Review und Verifikation).
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
| `spec/spezifikation.md` | update | `SPEC-120` Nr. 1 trägt die Entscheidung des Architect zum Block `CERTIFICATE` mit PEM-Headern **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); im Slice sonst nur Nachzug nach Review und Verifikation |
| `src/adapters/driven/httpupstream/` | update | `tls.go` neu (`tlsTransport`: `Protocols` nur HTTP/1, `TLSClientConfig` mit `RootCAs` und `InsecureSkipVerify`; `leseAnker`: Anker des Systems und `CAPEM`, mindestens ein lesbares Zertifikat), `client.go` (`New` ruft `tlsTransport`): Nr. 1, 7 |
| Tests von `httpupstream` | update | `tls_test.go` neu (`TestMain` macht ein erzeugtes Zertifikat S für `127.0.0.1` zum einzigen Anker des Systems und legt den Schlüssel neben S; Server A, S und E; `TestTLS`, `TestNewCAPEM`, `TestTransportTLS`, `tlsFehlerFaelle`), `fehler_test.go` (`alleFehlerFaelle` nimmt `tlsFehlerFaelle` auf): Fänger nach `LH-RB-02`, `LH-FA-01` |
| `src/bootstrap/konfig_dienste.go` und seine Tests | update, nur bedingt | nur wenn die Entscheidung zum Block mit PEM-Headern die Leseregel ändert: die Prüfung am Start an `UPSTREAM_CA_FILE` (`SPEC-110` Nr. 11) folgt derselben Regel (§1) |

Der Vertrag steht in `SPEC-120` vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten, das Raster in §6 (`AGENTS.md` §3.16). Die Zeilen tragen die Nummern aus `slice-upstream-adapter` §3, damit Bericht und Mutationen zu `0716718` auflösbar bleiben.

- Der Code ist geschrieben: der erste Code-Commit ist `git cherry-pick -x 0716718`, unverändert (§4). Ein Konflikt mit Änderungen, die Review und Verifikation von `slice-upstream-adapter` an `client.go` oder `fehler_test.go` brachten, wird im Cherry-Pick gelöst, und der Bericht nennt die Zeilen der Lösung. Was die Entscheidung zum Fund in §6 verlangt (Fall in `TestNewCAPEM`, Leseregel, Prüfung am Start), steht in eigenen Commits danach. Zwei Änderungen, die `slice-upstream-adapter` nach dem Revert trägt, liegen so, dass `0716718` ohne Konflikt passt: der neutrale Kommentar von `Konfig` (der Cherry-Pick ändert nur Kommentar und Rumpf von `New`) und die Fälle zum Namen des Hosts (A19 dort) in `argumentFaelle` und `argumentGrenzFaelle` statt in `alleFehlerFaelle`.
- TLS-Fälle laufen außerhalb einer Blase von `testing/synctest` (ihre Uhr beginnt am 2000-01-01 und läse die Gültigkeit erzeugter Zertifikate falsch); `TestMain` setzt `SSL_CERT_FILE` und `SSL_CERT_DIR` vor jedem Test, denn die Bibliothek liest die Anker je Prozess einmal.
- Lint-Profil wie in `slice-upstream-adapter` §3: `gochecknoglobals` (kein Pool und kein Schlüssel als Paketvariable), `testpackage`; keine `//nolint` (`AGENTS.md` §3.2).

**Zusage · Fänger** (jede Zeile löst der Reviewer mit einer Mutation auf; *U* Nr. n ist `SPEC-120` Nr. n; Zeilen aus `slice-upstream-adapter` §3 übernommen, die Fänger von A1 (`CAPEM`) und A10 im Wortlaut, A11 (TLS) und A14 (TLS) als die Teile der dortigen Zeilen, die den TLS-Teil betreffen):

| # | Zusage | Fänger |
|---|---|---|
| A1 (`CAPEM`) | `New` prüft `CAPEM` nach der Leseregel aus *U* Nr. 1, auch bei `Pruefen=false` | `TestNewCAPEM`: `CAPEM` Text ohne PEM, ein Block `PRIVATE KEY` allein, ein Block `CERTIFICATE` mit ungültigem Inhalt, je mit `Pruefen` wahr und falsch: `ErrUngueltig`, kein Client. Gutfälle: `CAPEM` nil und leer (je gegen S: 200); das Zertifikat des TLS-Servers; dasselbe zwischen Text und mit einem Block `PRIVATE KEY` davor; dasselbe hinter einem Block `CERTIFICATE` mit ungültigem Inhalt (je gegen den TLS-Server mit `Pruefen`: 200, die Wirkung ist geprüft, nicht nur `New`). Der Block `CERTIFICATE` mit PEM-Headern: Fall nach der Entscheidung des Architect (§6). Mutationen: `CAPEM` nur bei `Pruefen` geprüft (Fälle mit falsch rot); nur der erste Block gelesen (letzter Gutfall rot); je Prüfung entfernt (ihr Fall rot) |
| A10 | TLS: Kette und Hostname gegen die Anker des Systems und `CAPEM`; `CAPEM` ergänzt; `Pruefen=false` prüft nichts; jeder Fehler der Prüfung ist `ErrUpstreamNichtErreichbar` (*U* Nr. 1, 7) | `TestTLS`; `TestMain` macht vor jedem Test ein erzeugtes Zertifikat S (für `127.0.0.1`) zum einzigen Anker des Systems (`SSL_CERT_FILE` auf S, `SSL_CERT_DIR` auf ein leeres Verzeichnis). Server A: `httptest.NewTLSServer` (Zertifikat für `127.0.0.1`, `::1`, `example.com`, nicht für `localhost`); Server S; Server E mit einem abgelaufenen Zertifikat für `127.0.0.1`. `Pruefen`: ohne `CAPEM` gegen A: `ErrUpstreamNichtErreichbar` (unbekannte Stelle); A als `CAPEM` gegen A: 200; ohne `CAPEM` gegen S: 200 (Anker des Systems); A als `CAPEM` gegen S: 200 (ergänzt, ersetzt nicht); A als `CAPEM` über `https://localhost:p`: `ErrUpstreamNichtErreichbar` (Hostname); E als `CAPEM` gegen E: `ErrUpstreamNichtErreichbar` (abgelaufen). `Pruefen=false` ohne `CAPEM` gegen A über `127.0.0.1` und über `localhost` und gegen E: 200. Mutationen: Prüfung immer aus (unbekannte Stelle, Hostname, abgelaufen rot); `CAPEM` übergangen (A gegen A rot); `CAPEM` ersetzt die Anker (A als `CAPEM` gegen S rot); Anker des Systems nicht geladen (S-Fälle rot); Hostname nicht geprüft (Hostname-Fall rot) |
| A11 (TLS) | HTTP/1.1 auch über TLS (*U* Nr. 7) | `TestTransportTLS`: TLS-Server mit `EnableHTTP2`, `CAPEM` sein Zertifikat: der Handler sieht `HTTP/1.1`. Mutation: HTTP/2 erzwungen (rot). Der Fall ohne Proxy (`TestTransport`) bleibt in `slice-upstream-adapter` |
| A14 (TLS) | Allaussage: kein Fehler von `Senden` nennt Werte der Anfrage, auch nicht bei einem Fehler der Prüfung (*U* Nr. 8) | `TestFehlerOhneWerte` läuft über `alleFehlerFaelle` samt `tlsFehlerFaelle` (unbekannte Stelle, Hostname, abgelaufen), je mit den Markern aus `slice-upstream-adapter` §3, A14. Mutation: der Fehler der Prüfung mit `Pfad` oder `RawQuery` im Text verpackt (die TLS-Fälle rot); welche weiteren Mutationen aus A14 dort die TLS-Fälle trennen, löst der Reviewer auf |

Nicht gefangen und als akzeptiertes Negativ benannt, je mit Grund und Folge: der Inhalt der Vertrauensanker des Systems im Image und ein fehlender oder nicht lesbarer Bestand (gilt als leer; `leseAnker` beginnt dann mit einem leeren Pool, und kein Test fängt diesen Zweig), die TLS-Versionen, HTTP/2 — `SPEC-120` *Ungeprüft*, Grund und Folge je Fall dort.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-upstream-adapter` liegt in `done/` (der übrige Adapter ist reviewt und verifiziert, der Revert von `0716718` ist dort im Code). `SPEC-120` Nr. 1 trägt die Entscheidung des Architect zum Block `CERTIFICATE` mit PEM-Headern (§6), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Dieser Slice liegt vor `slice-bootstrap-m2`, dessen Start ihn in `done/` verlangt.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über **440** Zeilen nach dem Review (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3 unter `src/`, Tests eingeschlossen; `<Basis>` ist der Commit vor dem Cherry-Pick). Schätzung aus dem Ist: der TLS-Commit `0716718` misst in dieser Metrik 296 (291 Einfügungen, 5 Löschungen; 48 Produktcode, 248 Tests), dazu bis rund 30 für den Fall zum Fund in §6 (mehr, falls die Leseregel und mit ihr die Prüfung am Start sich ändert); mit dem Zuwachs im Review aus `slice-record-service` (+12,3 %) und `slice-capture-pipeline` (+16,1 %, je ab dem Stand vor dem Review, über die Pfade unter `src/`) rund 330 bis 380 nach dem Review. Der Schnitt, den eine Rückführung nimmt: was der Zuwachs an Gegenstand außerhalb von `SPEC-120` Nr. 1 (`CAPEM`) und Nr. 7 trägt, geht an einen eigenen Slice; Nr. 1 und Nr. 7 selbst teilen `tlsTransport` und sind nicht weiter teilbar.
- **Messpunkte:** nach dem Cherry-Pick, beim Handoff und nach dem Review, je im Bericht mit Produktcode und Tests getrennt; nach dem Review entscheidet der Planner.
- `in-progress` → `open` (blockiert): der Implementer findet eine weitere Randform ohne Ort in `SPEC-120` Nr. 1 oder 7 oder im Raster in §6, oder ein Fänger einer Annahme über die Standardbibliothek wird rot (`AGENTS.md` §3.12); dann ändert der Architect `SPEC-120`, nicht der Implementer den Test.

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16; Zeile aus `slice-upstream-adapter` §6 übernommen, *U* Nr. n ist `SPEC-120` Nr. n). **(a)** Abbruch: gilt nicht für Werte. **(b)** Wiederholung: *U* Nr. 9 („teilen nichts außer den Verbindungen des Clients").

| Operation · Argument | fehlend | leer | unbekannt | fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| `New` · `CAPEM`, `Pruefen` | nil: nur die Anker des Systems (*U* Nr. 1) | leer gleich nil (*U* Nr. 1 „nichtleeren") | kein lesbares Zertifikat, auch ein Block `CERTIFICATE` mit ungültigem Inhalt: `ErrUngueltig`, auch bei `Pruefen=false` (*U* Nr. 1) | Text, Blöcke anderer Art oder ungültige Blöcke neben einem lesbaren Zertifikat: ohne Wirkung (*U* Nr. 1); `Pruefen=false`: keine Prüfung, `CAPEM` ohne Wirkung (*U* Nr. 7); Bestand des Systems fehlt oder ist nicht lesbar: leer (*U* *Ungeprüft*); **Block `CERTIFICATE` mit PEM-Headern und gültigem Inhalt: ohne Ort, beim Architect** (unten) | (a) | (b) |

**Zusatzzeilen:**

| Fall | Ort (Satz) | Fänger |
|---|---|---|
| Modus: `Pruefen` an und aus × `CAPEM` × Anker des Systems × Hostname × abgelaufen | *U* Nr. 1, 7 | A10 |
| Allaussage: kein Fehlertext von `Senden` mit Werten der Anfrage, auch bei einem Fehler der Prüfung | *U* Nr. 8: „Kein Fehler von `Senden` nennt …" | A14 über alle Fehlerfälle samt `tlsFehlerFaelle` |
| Annahme über die Standardbibliothek: die Anker des Systems werden je Prozess einmal gelesen | kein Satz des Vertrags; Testumgebung (`TestMain`) | gilt nicht als Zusage: `TestMain` setzt die Umgebung, bevor ein Test läuft, und die S-Fälle von A10 gelten damit, ob die Bibliothek einmal oder öfter liest |

**Risiken:**

- Randform ohne Ort, gemeldet vom Implementer beim Handoff von `slice-upstream-adapter`: ein Block `CERTIFICATE` mit PEM-Headern und gültigem Inhalt. `AppendCertsFromPEM` überspringt ihn; nach dem Wortlaut von *U* Nr. 1 („mindestens einen Block `CERTIFICATE` mit gültigem Inhalt") zählte er womöglich als gültig. Ein `CAPEM`, das nur ihn trägt, ist im Code von `0716718` `ErrUngueltig`. Dieselbe Funktion liest am Start `UPSTREAM_CA_FILE` (`SPEC-110` Nr. 11, `src/bootstrap/konfig_dienste.go`). Im Code nicht entschieden — **Ausgang:** offen; der Architect entscheidet in `SPEC-120` Nr. 1 (und, wenn die Regel sich ändert, in `SPEC-110` Nr. 11) vor dem ersten Code-Commit dieses Slice; danach trägt A1 (`CAPEM`) den Fall und die Zelle oben ihren Satz.
- Cherry-Pick gegen einen geänderten Rest: Review und Verifikation von `slice-upstream-adapter` können `New` in `client.go` oder `alleFehlerFaelle` in `fehler_test.go` ändern, die `0716718` berührt — **Ausgang:** offen; ein Konflikt wird im Cherry-Pick gelöst, der Bericht nennt seine Zeilen, und der Reviewer sieht die Lösung gegen `0716718`. Stand nach Review, Verifikation und Nachverifikation von `slice-upstream-adapter`: `git merge-tree --write-tree --merge-base 0716718^ <Spitze> 0716718` ohne Konflikt, der Diff des Ergebnisses gegen die Spitze gleicht dem von `0716718` (vier Dateien, 291 Einfügungen, 5 Löschungen); mit Bau und Test des Ergebnisses geprüft auf `015da8c` (Nachverifikation dort), ohne Bau und Test auf `4a43c40`, dem letzten Commit mit Code von `slice-upstream-adapter` (Planner, bei dessen Closure).
- Zwischenstand: zwischen dem Revert in `slice-upstream-adapter` und dem Cherry-Pick hier prüft der Adapter bei `https` gegen die Anker des Systems, `CAPEM` und `Pruefen` bleiben ohne Wirkung, und HTTP/2 über TLS ist nicht ausgeschlossen — **Ausgang:** offen bis zum Cherry-Pick; entschieden (Architect, 2026-10-21, `slice-upstream-adapter` §4): `SPEC-120` nennt den Stand nicht und bleibt mit Nr. 1 und 7 der Vertrag dieses Slice; `slice-upstream-adapter` führt ihn als akzeptiertes Negativ mit Grund und Folge (dort §3), und kein Kommentar sagt eine Wirkung zu. Kein Lauf liefert ihn aus: der Composition Root verdrahtet keinen Adapter, und `slice-bootstrap-m2` verdrahtet erst, wenn beide Slices in `done/` liegen. Mit dem Cherry-Pick (§4) entfällt der Stand.
- Loopback und `localhost` in der Stufe `test`: die TLS-Fälle laufen über `127.0.0.1` und `localhost` (Architect-Probe in `slice-upstream-adapter` §6) — **Ausgang:** offen; der Fall über `localhost` ist nach `slice-upstream-adapter` §3 der erste des TLS-Commits; nach dem Cherry-Pick läuft er als erster in `make test`, bevor der Implementer weitere Fälle schreibt.
- Annahme über die Standardbibliothek (Go 1.27.1): `AppendCertsFromPEM` überspringt einen Block mit Headern (Bericht des Implementers beim Handoff) — **Ausgang:** offen; der Architect entscheidet mit dem Fund oben, ob ein Fall sie fängt; wird ein solcher Fänger rot, ändert er `SPEC-120`.
- Größe: 296 im TLS-Commit, Schätzung rund 330 bis 380 nach dem Review, Marke 440 (§4) — **Ausgang:** offen; gemessen nach dem Cherry-Pick, beim Handoff und nach dem Review.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Register unter `../observations/BEO-REPO/` trägt neun Einträge; dieser Plan antwortet auf jeden (Zahlen aus den Dateien unter `evidence/`):
- `randform-erst-im-review-entschieden` (14×, verkörpert in `AGENTS.md` §3.16): Raster mit der Zeile `New` · `CAPEM`, `Pruefen` und Zusatzzeilen mit Fänger-Spalte; die eine Zelle ohne Ort (Block mit PEM-Headern) steht als solche da und geht vor dem ersten Code-Commit an den Architect, nicht in den Code und nicht ins Review.
- `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (16×, verkörpert in `AGENTS.md` §3.15): Tabelle *Zusage · Fänger* mit vier Zeilen aus `slice-upstream-adapter`; je Element der Aufzählung in *U* Nr. 7 ein Fall (unbekannte Stelle, Hostname, abgelaufen), je Form von `CAPEM` ein Fall.
- `universalaussage-im-vertrag-ungeprueft` (5×, über der Schwelle, ohne Ausgang): die Allaussage zu den Fehlertexten läuft über die Tabelle aller Fehlerfälle samt den TLS-Fällen (A14 TLS).
- `nebenlaeufiger-lauf-nur-im-glueck-gruen` (5×, über der Schwelle, ohne Ausgang): keine nebenläufige Zusage in diesem Slice; die TLS-Fälle halten an keinem Zeitfenster.
- `groessenmarke-ueberschritten-ohne-rueckfuehrung` (11×, ohne Ausgang): dieser Slice ist der Schnitt einer Rückführung, beim Handoff genommen (`slice-upstream-adapter` §4); Schätzung aus dem gemessenen Ist des TLS-Commits, Marke nach dem Review in §4.
- `berichtigung-erreicht-nicht-jede-stelle` (2×): die Entscheidung zum Block mit PEM-Headern berührt zwei Leser derselben Regel (`SPEC-120` Nr. 1, `SPEC-110` Nr. 11); beide stehen in §3 und §6. Suchbegriffe einer späteren Berichtigung: `CAPEM`, `UPSTREAM_CA_FILE`, `CERTIFICATE`, `PEM`.
- `akzeptiertes-negativ-ohne-folge` (2×): die Negative in §3 nennen Grund und Folge in `SPEC-120` *Ungeprüft*.
- `befund-an-rolle-ohne-artefakt` (1×): der Fund in §6 geht an den Architect, eine Rolle mit Artefakt (`SPEC-120`); keine Owner-Entscheidung nötig.
- `werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×): keine Fixture mit Steuerbytes; PEM-Blöcke entstehen mit `pem.EncodeToMemory` im Test.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
