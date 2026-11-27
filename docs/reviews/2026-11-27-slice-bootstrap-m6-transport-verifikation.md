# Verifikationsbericht: slice-bootstrap-m6-transport (Teil C: TLS und Absicherung des Admin-Listeners im Composition Root) — 2026-11-27

**Rolle:** Verifier (Modul 11), „Bauen wir es richtig?“ gegen Plan, DoD und Vertrag. Genau ein Lauf, gesammelt.

**Gegenstand:** Zweig `slice-bootstrap-m6-transport`, HEAD `98eb01a1` (Merge von `main` `ce6736cb`), Worktree sauber. Bezug `LH-FA-13`, `LH-RB-02`, `LH-QA-04`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).

**Eingang:** Plan `slice-bootstrap-m6-transport` (§1 bis §6), `SPEC-299`, `SPEC-300` (Nr. 1, 3, 11, 13 bis 15), `SPEC-110`, `SPEC-139` Nr. 2, `SPEC-291`, `SPEC-128` Nr. 5, Review `2026-11-27-slice-bootstrap-m6-transport-review`, Fix-Runde `ba8e7f78`, Planner-Commit `ec52c030`.

## Verdikt

**Nicht bestätigt, solange V-1 offen ist.** Das Produkt erfüllt in allen geprüften Punkten den Vertrag; ein Produktfehler ist nicht gefunden. `make gates` ist auf `98eb01a1` **rot** (DoD-Punkt 3), und ein Gate-Nachweis ist nicht entstanden.

| ID | Kategorie | Befund | Auflage | an |
|---|---|---|---|---|
| V-1 | HIGH | `make gates` rot: `docs-check` meldet `slice-bootstrap-m6-transport.md:213`, `slice-replay-speicher-und-suche.md`, `target-missing` (1 Befund, 1374 Dateien). Zeile 203 desselben Plans zeigt seit dem Merge richtig auf `../done/`, Zeile 213 (Absatz *Kollisionen*, auf dem Zweig seit `2fe5e703`) weiter auf den alten Ort. Der Merge `98eb01a1` hat den Link nicht nachgezogen; `ec52c030` hat die Zeile nicht berührt. | Link auf `../done/` berichtigen, `make gates` neu bis zum Nachweis. Ohne das kein Merge. | Planner bzw. Integrationsschritt |
| V-2 | MEDIUM | Die Klasse von Review-F-1 (kein Wert des einen Paars wirkt auf den anderen Listener, `AGENTS.md` §3.17) deckt die Fix-Runde nicht ganz. Drei eigene Mutationen an nicht genannten Elementen bleiben im **ganzen** Paket `bootstrap` grün (V1 bis V3 unten). Das Produkt verhält sich in V1 und V3 am Binary richtig, in V2 nach dem Code (`ladeTransport` reicht den Fehler von `LISTEN_TLS_*` durch). | Je ein Fall: (a) Warnung am `Start` mit gesetztem `LISTEN_TLS_*`; (b) `Build` mit gültigem Admin-Paar und fehlerhaftem `LISTEN_TLS_*`; (c) beide Paare auf denselben Dateien, beide Listener sprechen TLS (`SPEC-299` Einleitung, so auch im Handbuch; keine Zeile in §3). Oder je ein akzeptiertes Negativ mit Grund, Folge und Maß in §3. | Implementer bzw. Planner |

## DoD

| Punkt | Stand | Beleg |
|---|---|---|
| 1: drei Schlüssel, vier Aufrufe, Reihenfolge, Modus, B19 aufgehoben, Tests nachgezogen, Ort der Namen | erfüllt (Einschränkung V-2 zu den Fängern) | Code gelesen; `test` und `race` frisch grün; Probe am Binary; kein Fänger von Teil B entfallen (`TestKonfigNichtGeliefert` 24 auf 12 abgelehnte Zellen, die drei Schlüssel dafür mit Positivprobe; `TestAdminPraefix` führt jetzt ihre Zeilen) |
| 2: Handbuch, Katalog, `handbuch-check`, kein neuer Code, keine Ausnahme | erfüllt | Aussagen gegen das Binary (unten); `handbuch` frisch `ok`, `katalog-check` grün (Stufe aus dem Schichten-Cache); `harness/katalog-ausnahmen.txt` und `fehler.go` unverändert |
| 3: `make gates` grün | **nicht erfüllt** | V-1 |
| Review liegt vor | erfüllt | Review-Report; F-1 bis F-5 und F-7 in `ba8e7f78`, F-6 in `ec52c030` |
| Doku-Update, Spezifikation nicht durch den Slice | erfüllt | kein Commit des Zweigs außerhalb der Merges ändert `spec/` oder `harness/` |
| Closure-Notiz, Register, Risiko-Ausgänge, Paarungen | offen (Closure) | §7 leer. In §6 stehen mehrere Ausgänge als *entschieden*; die DoD nennt *eingetreten / entfallen / weiter offen* |

## Sensoren (selbst gefahren, auf `98eb01a1`)

`make gates` (seriell, unter `flock /tmp/odata-recorder-heavy.lock`): Ende nach `a-check` und `baseline-verify` an `docs-check`. Danach `make -k gates`, damit die übrigen laufen: **allein `docs-check` rot**, alle anderen grün, darunter `pruefung test: ok` (454 s) und `pruefung race: ok` (552 s) frisch, `lint` 0 Befunde, `handbuch` ok, `groessenmarke` ok (Messung 1175 von Marke 1000, gedeckt durch den Eintrag), `schema-check`, `migration-check`, `test-integration` (10 Stufen), `test-determinism` (100 Läufe, 0 abweichend), `test-contract`, `test-security`, `zeichen-check`. Aus dem Cache kamen unter anderem die Stufe `katalog` und die Modul-Stufen der Werkzeuge. Die Container der Gates tragen ihre eigenen Grenzen; die Vorgabe 4 GiB galt für meine Läufe.

Nicht gefahren: `make gegenproben` (Auftrag); `harness/gegenproben.stempel` gehört in den Integrationsschritt.

## Probe am gebauten Binary

Ein Container ohne Netz (`--memory=4g --memory-swap=4g --cpus=4`), Binary der Stufe `build`, 1456 Prüfungen. 3 Abweichungen, alle aus einer falschen Erwartung der Probe (`:8081` ist am Laden ein Formfehler, §6 Zusatztabelle).

| Gegenstand | Ergebnis |
|---|---|
| 4 Modi × Paare {keins, Admin, `LISTEN`, beide} × Ort {Loopback, Loopback mit Flag, außerhalb mit Flag, außerhalb ohne} | jeder Listener spricht allein sein Paar oder HTTP; ohne Admin-Paar außerhalb ohne Flag Exit 1, genau die Zeile des Zusammenhangs, kein `DATA_DIR`, kein Listener |
| Host-Formen | `127.0.0.1`, `[::1]`, `localhost`: Klartext ohne Warnung; `0.0.0.0`, `[::]`, `127.0.0.2`, `[::ffff:127.0.0.1]`, `LOCALHOST`, `localhost.`, fremder Name und fremde IP: ohne Paar und Flag kein Start |
| Warnung | genau einmal, fester Text und `ereignis=admin_klartext`, vor `start: bereit`, allein im Weg Klartext, in jedem Modus |
| Flag | ungültige Werte: allein die Zeile am Flag, mit fehlendem Paar dazu dessen Zeilen; leer und Leerraum `false`; `true` mit Rand wahr |
| Ungültige Paare (12 Formen × 4 Modi; Loopback mit Flag, außerhalb mit und ohne Flag) | feste Zeilen, Exit 1, nie Rückfall auf HTTP, kein `DATA_DIR`; mit `LISTEN_TLS_*` eine sortierte Meldung. Nach einem Verstoß bleiben ein bestehendes `DATA_DIR` und das Token unverändert |
| Dateien zwischen Laden und `Build` geändert (je Datei des Admin-Paars eine benannte Pipe, 4 Modi × 3) | neues Paar: der Listener zeigt das zweite Blatt, keine dritte Lesung; passt nicht oder abgelaufen: Exit 1 mit Zeile, keine Datenbank, kein Token, kein Listener |
| TLS | 1.2 und 1.3, `http/1.1`; 1.0, 1.1 und allein `h2` scheitern; Klartext am TLS-Port nie 200; abgelehnter Handshake ohne Zeile; `healthz`, 401 und 200 mit Token über TLS wie im Klartext |
| Betrieb | Wechsel und Löschen der Dateien wirken nicht ohne Neustart; Ablauf im Betrieb: weiter bedient, keine Zeile, Klient lehnt ab; Kette und eine gemeinsame Datei für beide Listener; `SIGTERM` mit stummem Klienten Exit 0 |
| Geheimnis | Pfad, Schlüssel- und Zertifikatszeile, Seriennummer, CN, Hostwert, Token (Hälften): 0 Treffer in `stdout`, `stderr` und Antworten; `stdout` leer; Diagnose ohne `ADMIN_*` |

Handbuch: Tabelle `Einstellungen`, Abschnitt *TLS der Admin-Schnittstelle* und die vier berichtigten Stellen decken sich mit der Probe. Ein Hinweis ohne Auflage: *bevor der Dienst das Datenverzeichnis anlegt* gilt für Verstöße am Laden; entsteht der Verstoß erst in `Build` (Datei dazwischen geändert), besteht das leere `DATA_DIR` schon (`SPEC-299` Nr. 8 a und b, kein Vertragsbruch).

Nicht geprüft: Zugriff von einem anderen Host (`SPEC-299` *Ungeprüft*), Schreiben auf die Deskriptoren auf anderem Weg (akzeptiertes Negativ (2)).

## Fix-Runde gegen die Klassen (`AGENTS.md` §3.17)

Eigene Mutationen in einer Wegwerfkopie im Container, `go test -count=1 ./src/bootstrap/`, ohne Detektor, je ein Lauf, Last des Hosts 3 bis 11.

| Befund | Mutation (nicht im Review, nicht im Plan) | Lauf | Ergebnis |
|---|---|---|---|
| F-1 | V1: `Start` warnt allein ohne `LISTEN_TLS_*` (`h.tls == nil`) | ganzes Paket | **grün** |
| F-1 | V2: `Build` verwirft den Fehler von `LISTEN_TLS_*`, wenn der Admin-Weg `wegTLS` ist | ganzes Paket | **grün** |
| F-1 | V3: keine Umhüllung, wenn beide Paare dieselbe Zertifikatsdatei nennen | ganzes Paket | **grün** |
| F-2 | V7: auf Loopback fällt ein in `Build` nicht passendes Paar auf Klartext | ganzes Paket | rot (`TestAdminBuildLiestErneut`) |
| F-3 | V4: `FormGueltig` erst nach der Ablehnung anderer `ADMIN_*` bestimmt | `-run` Admin und Nachbarn | rot (`TestAdminLadenNebenFrueheren`) |
| F-4 | V5: `syscall.Write` in `lebenszyklus.go` | `-run` des Tests | rot (`TestAdminKeineAusgabeAmStrom`) |
| F-7 | V6: Schlüsselpfad in `Build` im Modus `replay` ersetzt | `-run` des Tests | rot (`TestWerteErreichenVerbraucher`) |

F-4: Grund, Folge und Maß stehen in §3. F-5: Satz berichtigt. F-6: siehe unten.

## Plan gegen Code, Entscheidungen

- Die 13 Pfade des Blocks sind genau die geänderten Dateien unter `src/`; dazu allein Handbuch, Plan und Review-Report.
- Zellzahlen der Zeilen C01 bis C09 stimmen mit den Tests überein (64; 44; 52 und 144; 24, 8 und 24; 12; 12 und 16; 128; Namenstest und Gates; 48).
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) eingehalten; keine neue `//nolint`-Direktive, `a-check` 0 Befunde, kein Gate gelockert, keine ADR geändert.

## Planner-Commit `ec52c030` (nach Anmeldefehler abgebrochen)

Nur der Plan, in sich vollständig. Nachgerechnet und richtig: Messung 1175 bei `ba8e7f78` und HEAD, 962 bei `cea33fb2`; 808 Testzeilen (701, 27, 80); +213 (132, 69, 8, 4); Produktcode 154 (68, 40, 19, 12, 9, 6); *S* = 810 und Formel-Marke 1100 bei 9 Zeilen; `SPEC-128` Nr. 5 c und d tragen den Eintrag bei fester Marke. Die Zeilen C03, C04, C06 und C09 nennen die geschriebenen Fänger; die zwei offenen Punkte in §6 haben einen Ausgang. Aus dem Abbruch ist kein Rest zu sehen.

## Übergabe

An den Planner: V-1 vor jedem Merge; V-2 einmal entscheiden (Fänger oder Negativ), eine zweite Verifikation ist dafür nicht verlangt. Für §7: ein auf dem Zweig entstandener Link auf einen Nachbar-Plan überlebt den Merge eines `slice-mv` von `main` nicht; die Unabhängigkeit zweier Einstellungen braucht auch die Querwirkungen auf Warnung, `Build` und gemeinsame Dateien.
