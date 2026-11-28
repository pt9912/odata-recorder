# Verifikation slice-admin-tls-betrieb (TLS des Admin-Listeners im Betrieb, Zeilen AT11 bis AT17) — 2026-11-28

Verifier, frischer Kontext, genau eine Verifikation. **Modell:** Sonnet 5.5.
Stand HEAD `9dc2b429`, Zweig `slice-admin-tls-betrieb-fang` (`main` `fa7b95f5` enthalten), vor der Prüfung `git status` leer.
Gegen: DoD des Plans (§2), Tabelle *Zusage · Fänger* (AT11 bis AT17), §3, §6; `SPEC-299` Nr. 7, 10, 11; `SPEC-300`; `LH-FA-13`, `LH-RB-02`, `LH-QA-04`; `2026-11-28-slice-admin-tls-betrieb-review`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

## Urteil

**DoD-Liefer-Punkte 1 bis 3, `make gates` und Review: bestätigt.** Kein HIGH, kein MEDIUM, ein LOW, zwei INFO. Die Fix-Runde deckt die Klassen F-1 bis F-5, bei F-1 mit dem Rest aus V-1. Keine offene Frage an den Architect blockiert die DoD. Nichts blockiert den Merge.

## Auflagen für die Closure

1. **V-1:** den Vergleich um die fehlenden Teile ergänzen oder sie in §3 als Negativ mit Grund, Folge und Maß nennen (`AGENTS.md` §3.15).
2. **V-2:** die vier Einträge von §6 mit Ausgang *offen* brauchen vor `done/` einen der drei Ausgänge; die zwei mit *weiter offen* und dem Nachtrag zum Stand (Pflichtfeld `Uhr`, gemeinsames Testpaket) sind auf ihren Ausgang zu prüfen.
3. Die üblichen Closure-Punkte der DoD (Closure-Notiz, Register, Paarungen) sind leer.

## DoD

| Punkt | Ergebnis | Beleg |
|---|---|---|
| 1 Verbindung und Rahmen über TLS (AT11 bis AT13, `m10d`, `m10e`) | bestätigt, Rest V-1 | `m10d` und `m10e` selbst rot in `TestAdminVerbindung` und `TestAdminHandshakeScheitert`; Probe am Binary unten, Abschnitte 1 bis 3 |
| 2 Hängen, `Anhalten`, `Schliessen` (AT14 bis AT16, "in jedem Lauf") | bestätigt | je `-count=20` ohne Detektor: `m14a`, `m15a`, `m15b`, `m16a`, `m16c` in 20 von 20 rot, je Untertest; ohne Mutation 20 Läufe grün |
| 3 Kein Neuladen (AT17) | bestätigt | `GetCertificate`, das je Handshake lädt, rot in `TestAdminKeinNeuladen`; Probe am Binary, Abschnitte 5 und 6 |
| `make gates` | bestätigt, nicht neu gefahren | `gates-passed.head` gleich `9dc2b429`, `gates-passed.diffsha` gleich der Ausgabe von `tools/harness/working-tree-hash.sh` am sauberen HEAD (`31057d3b…`); Log des Laufs `rc=0`: `pruefung test: ok` (446,6 s), `pruefung race: ok` (512,8 s), `d-check` 0 Befunde, Marke 855 von 1100. Der Stempel gilt für den Stand vor diesem Report |
| Review | bestätigt | Report liegt vor; ein zweiter Review-Lauf ist nicht erfolgt, die Fix-Runde ist hier gegen die Klassen aufgelöst |
| Doku-Update, Handbuch | bestätigt | der Diff gegen `main` trägt allein die Testdatei, den Plan und den Review-Report; `docs/user`, `spec/`, Gates unberührt |

## Plan-vs-Code-Diff

- §1 "kein Produktcode, Binary unverändert" und "kein Pfad unter `src/adapters/driving/admin/`": stimmt. `admintls.go`, `admintls_test.go`, `export_test.go`, `tlsdatei.go` unberührt.
- §3 Mengen: die Literale im Test (9, 8, 56, 4, 2, 2, 2, 3) stimmen mit dem Plan; die Aufteilung der 14 Zeilen von M-Zuordnung (2 Treffer, 7 mal 404, 5 mal 405) stimmt.
- Marke: `git diff --numstat` über den Pfad des Blocks 855 von 1100.
- Zeile AT16 nennt `ErrFristAbgelaufen` als roten Befund; gemessen (`m16a`) ist `nil` nach 4,0 s, rot an der Dauer. Das bestätigt F-6; die Rückgabe steht in §6.

## Fix-Runde gegen die Klassen (eigene Mutationen, Wegwerfkopie im Container)

Jede Ersetzung traf genau eine Stelle. Filter: die acht Tests der Testdatei; `Y1`, `Z1`, `Z2` und die Wiederholungen allein die betroffenen Tests. Ohne Mutation grün.

| Klasse | Eigene Mutation | Ergebnis |
|---|---|---|
| F-1 Antwort über TLS wie im Klartext | `V1` `Connection: close` über TLS auf jeder Antwort | rot, aber allein in `TestAdminAnhaltenUeberTLS` (zweiter Request derselben Verbindung), nicht in den Vergleichen |
| F-1 | `V1b` dasselbe allein bei `DELETE` (405) | **grün**, auch in den ganzen Paketen `bootstrap` und `admin` (V-1) |
| F-1 | `V3` zweiter Wert `X-Correlation-Id` über TLS | **grün**, ebenso in den ganzen Paketen (V-1) |
| F-2 "meldet nichts" | `W1` `VerifyConnection` schreibt bei abgelaufenem Blatt nach `stdout` · `W2` `ConnState` meldet das Ende einer TLS-Verbindung über `log` · `W3` `slog.Warn` je Annahme | alle rot (`W1` in AT17, `W2` und `W3` in AT13 und AT17) |
| F-3 Zählprüfung statt Panik | je eine Zelle entfällt: erste Zeile von M-Zuordnung, `allein h2` (M-Ausgang), `ALPN keines`, `Route ohne Token` (M-Anfrage), (i) von M-Anhalten | alle rot mit "Menge … hat n Zellen", keine Panik |
| F-4 Warten und Zählen prüft | `Y1` das Close der angenommenen Verbindung wirkt erst nach 5 s | rot in AT13 und AT17 mit "n von m erwarteten Close-Aufrufen" |
| F-5 Zelle wird aus ihrer Zusage rot | `Z1` Ablehnung nach dem Handshake ohne ALPN · `Z2` TLS 1.2 verlangt ein Client-Zertifikat | rot an der Zelle (`ALPN keines` und `Client ohne Zertifikat`; `TLS 1.2`) |

## Probe am gebauten Binary

Binary aus `git archive 9dc2b429` mit den Flags des `Dockerfile` gebaut; Modus `record`; ein Prozess mit `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` (Blatt und Zwischenzertifikat), ein zweiter im Klartext auf Loopback. 0 Fehler.

1. **Antworten:** 24 Anfragen (200, 201, 400, 401, 404, 405 mit `Allow`, 431, kaputte Zeile, `HEAD`, `OPTIONS *`, HTTP/1.0, Formen der Authentisierung, Diagnose) sind in den Rohbytes samt Statuszeile und `Connection` gleich, bis auf Datum, Zeitstempel und Kennungen; `audit.jsonl` beider Prozesse ebenso (6 Zeilen).
2. **Verbindung:** TLS 1.2 und 1.3 mit Kette aus 2 Zertifikaten; TLS 1.1 und allein `h2` scheitern; `h2` mit `http/1.1` ergibt `http/1.1`; kein `CertificateRequest`.
3. **Handbuchsatz** (*TLS der Admin-Schnittstelle*, "Der Dienst schreibt dazu keine Zeile") ist wahr: unbekannte CA, falscher Name, TLS 1.1, allein `h2`, vom Client als abgelaufen abgelehnt, Klartext-Request (400 der Standardbibliothek ohne `X-Correlation-Id`), Müll, halbes ClientHello, sofortiges Ende und eine stumme Verbindung bis zum Ende durch den Server (10 s): `stdout`, `stderr`, `audit.jsonl` und die Zähler der Diagnose unverändert. Positivkontrolle: dieselbe Erfassung sah die Zeile des Starts.
4. **Hängen und Ende:** 200 hinter 32 stummen Verbindungen; `SIGTERM` mit offenen Handshakes: Exit 0 nach 41 ms, keine Zeile.
5. **Kein Neuladen:** nach Überschreiben und Löschen der Dateien das alte Zertifikat, nach Neustart das neue.
6. **Ablauf im Betrieb** ("bedient die Schnittstelle es weiter und meldet nichts"): wahr; mit der Uhr des Clients vor dem Ablauf 200, mit der Uhr des Hosts lehnt der Client ab, keine Zeile, kein Eintrag.

## Befunde

**V-1 (LOW) — der Vergleich "dieselbe Antwort" lässt Teile der Antwort aus.** `adminAntwort` vergleicht Status, Header und Body. `net/http` nimmt `Connection: close` beim Lesen aus den Headern (`resp.Close`), die Version der Statuszeile steht in `resp.Proto`; beides vergleicht der Test nicht, und von `X-Correlation-Id` liest er allein den ersten Wert. `V1b` und `V3` bleiben in beiden Paketen grün (2 von 2); `tests/` nennt `ADMIN_TLS` nicht. Der Plan nennt dazu kein Negativ, die Kommentare sagen "dieselbe Antwort" (`AGENTS.md` §3.11, §3.15). LOW statt MEDIUM, weil die Zusage für Status, Body und Header einen Fänger hat, das Binary heute byteweise gleich antwortet (Probe 1) und eine Abweichung verlangt, dass der Adapter `r.TLS` liest. Den Satz "er liest `r.TLS` nicht" (`SPEC-299` Nr. 10 d) trägt der Code heute (Suche ohne Treffer); die gefahrenen grünen Mutationen zeigen, dass die Pakete `bootstrap` und `admin` ein solches Lesen nicht als solches fangen.

**V-2 (INFO) — die offenen Punkte in §6 blockieren die DoD nicht.**

- *Abweichung von der Quelle* (Wortlaut AT12, AT17): die Tests führen mehr, als die Zeilen nennen; die DoD ist erfüllt. Planner und Architect.
- *M-Zuordnung* in der Zeile AT12, Formen der Authentisierung und Fehler der Routen über TLS: die DoD verlangt Antwort, `Anfrage` und Einträge wie im Klartext und ist für die geführten Anfragen belegt; der Rest steht als Negativ (2) mit Grund, Folge und Maß. Hinweis an den Architect: eine Quelltext-Prüfung, dass der Adapter `r.TLS` nicht liest, fienge das Negativ (2) und V-1, soweit die Abweichung im Adapter liegt, als eine Klasse.
- *Zeile AT16:* reine Wortlaut-Frage (siehe Plan-vs-Code-Diff).
- *`TestGleichzeitigeSpans`:* nicht Gegenstand des Slice; im Gate-Lauf auf HEAD grün. Planner (Register).

Vor `done/` ist *offen* kein zulässiger Ausgang.

**V-3 (INFO) — Laufzeit von `src/bootstrap` unter `-race` mit 4 CPUs: 415,9 s** (Grenze 600 s je Paket, bisher 404 bis 410 s). Ein Lauf auf HEAD, `go test -race -count=1 -trimpath -buildvcs=false`, `GOMAXPROCS` nicht gesetzt, Container `--cpus=4 --memory=4g`, Host mit 20 Prozessoren; Last vorher 5,35 / 6,71 / 8,17, nachher 3,94 / 4,65 / 6,65; gemessen vom Verifier. Der Gate-Log nennt die Zeit je Paket nicht. Ein Lauf belegt keine Zunahme gegenüber 410 s (`AGENTS.md` §3.10); der Abstand zur Grenze ist 184 s.

## Nicht geprüft

- Die Nullblöcke von `admintls.go` (§3, 16 Blöcke) und die Raten mit Detektor und mit `GOMAXPROCS=1` aus §3.
- Die Probe am Binary in den Modi `replay`, `mock` und `hybrid`.
- `Anhalten` über TLS am Binary (allein über die Tests und Mutationen belegt).
- `make gegenproben` (Auftrag).

## Lauf-Bedingungen

Schwere Läufe nacheinander in je einem Container, ohne Netz und ohne Mount, `--memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`; nach dem Warten 13 GiB verfügbar, Last 5,35 bis 6,38. Image: die Stufe `race` (Toolchain und Module), `src/` und `tests/` aus `git archive 9dc2b429` ersetzt; für den letzten Lauf (ganze Pakete) die inzwischen neu gebaute Stufe desselben Tags. Im Worktree ist außer diesem Report nichts geändert.
