# Review-Report: slice-admin-tls-betrieb (TLS des Admin-Listeners im Betrieb, Zeilen AT11 bis AT17) — 2026-11-28

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf.

**Gegenstand:** `git diff 7ced9461..e51017e1` (eine neue Testdatei `src/bootstrap/admintls_betrieb_test.go` und der Slice-Plan; kein Produktcode).

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`e51017e1`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-28

**Eingangs-Kontext:** Plan `slice-admin-tls-betrieb` (§1, §3 mit den Abweichungen von der Quelle und der Messung, §6); `SPEC-299` Nr. 7, 10, 11; `SPEC-300`; `LH-FA-13`, `LH-RB-02`, `LH-QA-04`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `AGENTS.md` §3; die Quelle `git show 3726eda0:src/bootstrap/admintls_betrieb_test.go`.

**Urteil: nicht freigegeben bis F-1.** Kein HIGH. Die vier Ergänzungen (Zählprüfungen, Query und Body, "meldet nichts", Handbuchsatz) tragen, was sie nennen, mit den Grenzen aus F-2 und F-3.

**Vorgehen.** Wegwerfkopie (`git archive e51017e1` über `src`, `tests`, `go.mod`, `go.sum`) per Tar-Strom in genau einen Container aus dem Image `odata-recorder:race` (`--network none --memory=4g --memory-swap=4g --cpus=4`, kein Mount), unter `flock /tmp/odata-recorder-heavy.lock`, zweimal nacheinander gestartet. Beim Start 14 GiB verfügbar, Last 3,7 und 6,2. Messend: Reviewer. Filter F7: die sieben Tests der Datei. Jede Ersetzung traf genau eine Stelle.

- Grundlauf F7 mit Detektor grün; je Test höchstens 0,04 s, das Paket mit F7 1,2 s.
- Flakiness, je 0 rote Läufe: mit Detektor `-count=30` (4 CPU) und `-count=15` bei `GOMAXPROCS=1`; ohne Detektor `-count=60 -cpu 1,4`; sechs parallele Prozesse mit Detektor und `-test.count=10`.
- Lecks: nach F7 (einmal und zehnmal im selben Prozess) 2 Goroutinen und 7 offene Deskriptoren, ohne Wachstum; keine hängende Verbindung.
- Zeitgrenzen: jedes Dial, Lesen und jeder Client trägt eine Frist; zum Warten auf die Close-Meldungen siehe F-4.
- Nicht gefahren: `make gates`, `make test`, `make test-race`, `make lint`, `make gegenproben`. `TestGleichzeitigeSpans` lag außerhalb der gefahrenen Pakete und wurde nicht gesehen.

**Eigene Mutationen** (nur in der Wegwerfkopie):

| Nr. | Mutation | Ergebnis |
|---|---|---|
| K1 | `umhuelle` bindet beim Umhüllen einen Melder an `os.Stderr` und meldet bei abgelaufenem Zertifikat je Handshake | F7 **grün**; rot allein `TestAdminGleichzeitig` (Importe `log` und `os` in `admintls.go`) |
| K1b | wie K1, `os.Stderr` erst beim Schreiben gelesen (Kontrolle) | rot: `TestAdminKeinNeuladen` |
| K1c | wie K1, in `ladeTLS`; ganzes Paket `bootstrap` | rot allein `TestLadeTLSVersion`, `TestLadeTLSKeinGeheimnis`, `TestLadeTLSFelder` |
| K2 | `ErrorLog` des Adapters schreibt nach `os.Stderr` | `TestAdminKeinNeuladen` allein **grün** (drei Meldungen `bad certificate` außerhalb der Erfassung); rot: `TestAdminHandshakeScheitert` |
| K3 | der Adapter setzt bei `r.TLS != nil` in den Zweigen 404 und 405 einen Header | **grün:** F7 und im Paket `bootstrap` alle Tests mit `Admin`, `M6` oder `Sicherheit` im Namen |
| K4 | `KorrelationsID` der `Anfrage` bei TLS leer | rot: `TestAdminRahmenUeberTLS` |
| K5 | in M-Anfrage entfällt die letzte oder die erste Anfrage (zwei Läufe) | rot durch `panic: index out of range`, die Zählprüfung läuft nicht |
| K6 | AT13 erwartet für "unbekannte CA" 3 statt 1 Close-Aufrufe | **grün** nach 12 s |
| K7 | Client-Zertifikat verlangt, im Test ohne die Anfrage in der Schleife | rot allein die Zelle TLS 1.2; die Zelle "Client ohne Zertifikat" meldet nichts |

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | `SPEC-299` Nr. 10 d sagt die *Zuordnung* über TLS wie im Klartext zu; M-Anfrage führt allein die getroffene Route und `/healthz`, keinen unbekannten Pfad (404) und keine falsche Methode (405), und weder Tabelle noch Negativ nennen das (K3). Versagen: eine Abweichung des Adapters in diesen Zweigen, die nur über TLS wirkt, besteht die Suite. Die Menge stammt aus dem Plan (Architect). | `SPEC-299` Nr. 10 d; `AGENTS.md` §3.15 | `src/bootstrap/admintls_betrieb_test.go` · `{"Route mit falschem Token", "/probe"` | ja, `make test` bliebe mit K3 grün | Zusage ohne Fänger |
| F-2 | LOW | Die Erfassung zu "meldet nichts" umfasst allein den angenommenen Handshake: Server und Transport entstehen davor, und der Handshake, den der Client am abgelaufenen Zertifikat ablehnt (der Betriebsfall von `SPEC-299` Nr. 11), läuft danach. Eine Meldung mit beim Umhüllen gebundenem Ziel (K1) oder bei der Ablehnung (K2) sieht AT17 nicht; gefangen wird sie nur mittelbar (Quelltext-Wächter anderer Slices, AT13 ohne Ablauf), und der Plan nennt kein Negativ. | `SPEC-299` Nr. 11; `AGENTS.md` §3.15, §3.16 | `src/bootstrap/admintls_betrieb_test.go` · `senke := ausgabeErfasst(t, func() {` in `TestAdminKeinNeuladen` | ja, F7 bliebe mit K1 grün | Fänger führt nur eine Form der Eingabe |
| F-3 | LOW | Die Zählprüfung von M-Anfrage kann das Entfallen eines Elements nicht melden: die Soll-Einträge greifen vorher fest auf `ids[i][2]` und `ids[i][3]` zu, der Test endet in einer Panik (K5). §3 sagt, jede der sieben Mengen zähle gegen das Literal, und nennt M-Anfrage als nicht mutiert. | `AGENTS.md` §3.15; Reviewer-Skill, Hinweis zu `panic:` | `src/bootstrap/admintls_betrieb_test.go` · `KorrelationsID: ids[i][3]` | ja, Lauf mit K5 | Zählprüfung nicht erreichbar |
| F-4 | LOW | `warte` kehrt nach Ablauf der 3 s still zurück: die in `schluss` erwarteten Close-Aufrufe prüft niemand (K6), und die Zähler können dann gelesen werden, bevor der Server den Fall beendet hat. Der Zähler `angenomm` wird nirgends gelesen, obwohl Kommentar und Plan (§3, *der Listener (Annahmen, Schließungen)*) das Zählen der Annahmen nennen. | `AGENTS.md` §3.11, §3.15 | `src/bootstrap/admintls_betrieb_test.go` · `func (l *adminListener) warte(n int)` · `angenomm atomic.Int64` | ja, `make test` bliebe mit K6 grün | Hilfe wartet oder zählt, ohne zu prüfen |
| F-5 | LOW | Die Zelle "Client ohne Zertifikat" prüft allein den Rückgabewert des Dial; unter TLS 1.3 gelingt er auch, wenn der Server ein Zertifikat verlangt (K7). Die Zusage (kein mTLS) fangen die Zellen mit Anfrage, die Zelle selbst nicht. | `SPEC-299` Nr. 7; `AGENTS.md` §3.11 | `src/bootstrap/admintls_betrieb_test.go` · `t.Errorf("Client ohne Zertifikat:` | ja, Lauf mit K7 | Zelle kann aus ihrer Zusage nicht rot werden |
| F-6 | INFO | Die Zeile AT16 nennt als roten Befund `ErrFristAbgelaufen`; gemessen ist die Dauer (§3 *Messung dieses Slice* sagt es). Die Rückgabe in §6 nennt allein die Zeilen AT12 und AT17. Zuständig: Planner und Architect. | `AGENTS.md` §3.13 | Plan `slice-admin-tls-betrieb` §3 · `(rot: ErrFristAbgelaufen)` | nein | Plan-Zeile weicht von der Messung ab |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Übernahme aus der Quelle (Diff gegen `3726eda0`) | geprüft, ohne Befund: `ladeAdmin` entfällt, `Uhr` gesetzt, sonst allein die vier Ergänzungen aus §3 |
| Zählprüfungen von M-Verbindung, M-Haengend, M-Anhalten, M-Offen, M-Neuladen | geprüft (gelesen), ohne Befund: jede Zelle zählt einmal gegen das Literal |
| AT12, Query und Body | geprüft, ohne Befund außer F-1: Literal und Vergleich beider Transporte, `KorrelationsID` gefangen (K4) |
| Handbuchsatz (`docs/user/benutzerhandbuch.md`, *TLS der Admin-Schnittstelle*) | geprüft, ohne Befund: Satz unverändert und mit dem Code verträglich; der Root umhüllt den Listener nur mit `umhuelle`; AT13 wird mit K2 rot (alle vier Arten); das Negativ zum Root steht mit Grund, Folge und Maß |
| AT14, AT15, AT16 (Ordnung, Fristen, Haltepunkt) | geprüft, ohne Befund: Reihenfolge durch Annahme und Kanäle, keine Schlaf-Aufrufe, kein roter Lauf |
| `src/adapters/driving/admin/`, `src/bootstrap/admintls.go`, `export_test.go` | geprüft, ohne Befund: im Diff unberührt |
| Gates (`.golangci.yml`, `Makefile`, `harness/mk/`, `Dockerfile`, `.a-check.yml`, `.d-check.yml`, `//nolint`) | geprüft, ohne Befund: keine Änderung, keine Unterdrückung |
| Kommentare der Testdatei (`AGENTS.md` §3.7) | geprüft, ohne Befund außer F-4 |
| Slice-Plan §1, §4, §6 (Abgrenzung, Rückgabe, Beobachtung) | geprüft, ohne Befund außer F-6 |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 4 |
| INFO | 1 |

**Finding-Klassen dieses Laufs:** Zusage ohne Fänger · Fänger führt nur eine Form der Eingabe · Zählprüfung nicht erreichbar · Hilfe wartet oder zählt, ohne zu prüfen · Zelle kann aus ihrer Zusage nicht rot werden · Plan-Zeile weicht von der Messung ab

## Verdikt

**Merge-blockierend:** ja, wegen F-1. F-2 bis F-5: Annahme oder Begründung. F-6 ohne erwartete Aktion des Implementers.

**Übergabe:** an den Implementer; F-1 berührt die vom Architect gesetzte Menge M-Anfrage (Rückkante zum Plan). Die Finding-Klassen gehen in die Closure-Notiz. Dieser Report ersetzt keine Verifikation.
