# Verifikation slice-fallback-policy-entscheidung (Teil 2: Entscheide)

**Rolle:** Verifier (einmal). **Datum:** 2026-11-08. **Stand:** Zweig `slice-fallback-policy-entscheidung`, HEAD `1914ee9f` (Fix-Runde nach dem [Review](2026-11-08-slice-fallback-policy-entscheidung-review.md)). **Bezug:** [LH-FA-12](../../spec/lastenheft.md), [ADR-0001](../plan/adr/0001-implementierungssprache-go.md).
**Plan:** [slice-fallback-policy-entscheidung](../plan/planning/done/slice-fallback-policy-entscheidung.md).

## Urteil

Die Zusagen von `SPEC-261` Nr. 1 bis 10 sind gehalten, soweit der Slice sie trägt; keine Abweichung zwischen Plan, Vertrag und Code, und keine Sicherheitslücke in der Upstream-Freigabe: über 60 Mio. Eingabepaare des eigenen Generators (Orakel allein aus `LH-FA-12.b` und `SPEC-261`) erlaubt kein Paar einen Upstream-Zugriff, den der Wortlaut nicht erlaubt. Die Fix-Runde deckt M2, M3, M4 und M5 als Klasse. **M1 deckt die Klasse nicht ganz** (Auflage 1): eine Mutation am Wächter von `zielPort`, die der Befund nicht nannte, bleibt in `make test` grün. Das ist eine Lücke des Fängers, kein Fehler des Produktcodes. Drei weitere Auflagen (Plan-Text, Zählstellen bei der Integration, ein Hinweis): unten.

## Belege der Sensoren (selbst gefahren)

| Sensor | Ergebnis |
|---|---|
| `make gates` (am HEAD, unter `flock /tmp/odata-recorder-heavy.lock`; der erste Lauf endete ohne Ausgabe und ohne Exit-Meldung nach `zeichen`, nicht gewertet, der zweite lief mit `EXIT=0`) | Exit 0; darunter `baseline-verify` OK, `d-check` 0 Befunde, `test-race`, `a-check`, `lint`, `katalog-check`, `zeichen`, Stufe `groessenmarke` |
| `groessenmarke` | `slice-fallback-policy-entscheidung.md: ok, Messung 1671 von Marke 2000`; `Messpunkt: zeilen=16 nebenlaeufig=0 fakes=0 produktcode=200 schaetzung=1480 marke=2000 messung=1671` |
| eigene Messung `git diff --numstat 56dc7c58..HEAD -- src/hexagon/services/hybrid` | 1668 + 3 = 1671 (Produktcode 266 Zeilen gegen Block 200, eine Erwartung, keine Grenze) |
| Last | Host 5 bis 24 bei 20 CPUs; jeder Lauf unter dem Schloss, eigene Container mit `--memory=4g`, ohne Netz, Quellen schreibgeschützt eingehängt (Probe und Mutationen); `make gates` baut die geteilten Tags wie jeder Lauf |

Die Behauptung des Implementers (Gates grün, 1671) ist damit belegt, nicht nur zugesagt.

## Plan-vs-Code-Diff

- Gegenüber dem Start-Commit `c5dd4b2a` berührt der Zweig genau: `src/hexagon/services/hybrid/policy/` (drei neue Dateien Produktcode, `policy.go` nur Paketkommentar, sieben Testdateien), `spec/spezifikation.md`, den Plan und die zwei Review-Reports. Keine Änderung an `Makefile`, `Dockerfile`, `harness/`, `tools/`, `.golangci.yml`, `.a-check.yml`, `.claude/`, `.harness/`, `AGENTS.md`. Die Änderung von `harness/gegenproben.stempel` gegenüber der Basis `56dc7c58` stammt aus `main` (`98bcbb5d`, Merge), nicht aus diesem Slice.
- Fix-Runde (`531d2ff2..HEAD`): Produktcode unverändert (nur `export_test.go`, `gleichzeitig_test.go`, `vergleich_test.go`, `ziel_test.go`), die Spezifikation in einer Zeile.
- **Architect-Text berührt (`spec/spezifikation.md`, *Ungeprüft*, Zeile „Lesart des Ziels durch den HTTP-Client“):** Wort-Diff `531d2ff2..HEAD` zeigt allein die Streichung des Klammerzusatzes mit der Slice-Kennung vor dem Punkt; Grund, Folge und der Fänger „beim Dienst“ sind byte-gleich geblieben. Inhalt unverändert.
- Das Paket importiert `net/netip`, `net/url`, `strings` und `model`; kein Port, kein Adapter, kein Dienst (`make a-check`, `TestQuelltext` grün). Kein `nolint`. Bytes über `0x7F` stehen nur als Umlaute in Kommentaren; keine Folge aus Backslash, `u` und vier Hexziffern, kein BOM, kein NBSP (`LC_ALL=C grep -P` über alle Dateien des Pakets; dieser Report ist ebenso gegengelesen).
- Die Tabelle *Zusage · Fänger* hat 16 Zeilen (P01 bis P10, P12, P15, P16, P22 bis P24). Gegen den Vertrag: Nr. 1 → P01; Nr. 2 a bis d → P02, P04; Nr. 3 → P03; Nr. 4 a bis f → P05 bis P10; Nr. 5 b bis e → P12; Nr. 6 c bis e → P15, P16; Nr. 6 f und *Ungeprüft* → P24 (Negativ mit Grund und Folge); Nr. 9 a, b → P22; Nr. 9 c → `TestGleichzeitig` und P21 des Teils 1; Nr. 10 → P23 (Negativ). Nr. 5 a, 6 a und b, 7, 8 liegen im Teil 1. Keine Zusage ohne Fänger; die Zahlen der Mengen stimmen mit den Zählprüfungen der Tests überein (62 608 / 1176 / 110, 336 + 28, 135, 168 + 8, 456 × 2961 mit 18 809 Treffern, 960 Port-Fälle, 41 × 12 Paare, 11 + 7 Paare von `zielName`).
- Das Binary bleibt unverändert (keine Datei außerhalb des Pakets; `make handbuch-check` und `make katalog-check` in `make gates` grün).

## Sicherheit der Upstream-Freigabe (eigener Probe, Orakel unabhängig vom Code)

Der Probe lag als zusätzliche Testdateien in einer Kopie des Pakets (nicht im Repo), Orakel allein aus [LH-FA-12](../../spec/lastenheft.md) `.a`/`.b` und `SPEC-261`; Regex-Grammatik für das Ziel, `math/big` für Ports, `netip`-Werte statt Texte, `regexp` je Segment für Muster.

| Probe | Menge | Ergebnis |
|---|---|---|
| Code × Methode × Zustand × Ziel (Pfad zufällig aus 17) | 44 Codes (40, Nullwert, −1, N + 1, N + 100) × 23 Methoden (auch `GET` mit Leerzeichen davor und dahinter, NUL in der Methode, `HEAD` mit Zeilenumbruch, `get`, `Get`) × 868 Zustände (Hauptschalter aus und an × 16 Host-Listen × 7 Freigabe-Listen × 4 Kategorie-Listen, ohne Schalter an mit leeren Hosts) × 69 Ziele (Benutzerteil in allen Lagen, `@` und Backslash im Pfad, Query und Fragment, Zone, Klammerformen, Ports an den Grenzen, Schema in Großschreibung, `h.`, `xh`, Prozent-Kodierung, Bytes ab `0x80`, NUL) | 60 610 704 Fälle, 12 767 `Upstream`, 628 866 `Replay`, **0 Abweichungen vom Orakel**, und jeder `Upstream` hat ein Ziel, das auch eine von `url.Parse` unabhängige Regex-Grammatik annimmt |
| Zufallsfuzz über Ziele | 7,5 Mio. Ziele aus 34 Bausteinen (`@`, Backslash, Klammern, `%41`, Leerraum, NUL, Umlaute …) × 5 Einträge | 15 300 `Upstream`, jeder mit gleichem Host und gleichem Port nach der Regex-Grammatik |
| Muster × Pfad | 52 059 Muster (Längen 1 bis 3 aus 37 Segmenten, darunter alle sechs Punkt-Formen von Nr. 6 b, `**a`, `a**`, `%`-Reste) × 400 Pfade je gültigem Muster | `New` nimmt genau die 3856 Muster an, die Nr. 6 b zulässt (52 059 von 52 059 gleich); 1 542 400 Paare, 459 Treffer, 0 Abweichungen. Die Trefferquote ist klein, weil der Pfadgenerator zufällig wählt; die Menge mit Rang liefert P15 (18 809 Treffer von 1,35 Mio.) |
| Port (M1) | 16 672 Texte (2^k und 2^k − 1 für k = 0 bis 200 mit sieben Zuschlägen und drei Formen, Vielfache von 2^8 bis 2^128 mit Rest, Nullen, 1 bis 120 Neunen, Einsen, 3000 Zufallsfolgen, Nichtziffern) × 2 Schemata × 8 Einträge | 266 752 Fälle, 817 frei, 0 Abweichungen |
| Benutzerteil (M5) | 1728 Ziele (3 Schemata × 18 Userinfo-Formen × 8 Hosts × 4 Endungen) | 0 frei; `@` hinter dem Host (Pfad, Query) bleibt orakelkonform |

**Führende Null im Port (Entscheid des Architect `531d2ff2`):** bestätigt. `https://h:0443` und `http://h:00080` gelten als Port 443 und 80 (`SPEC-261` Nr. 5 c im Wortlaut), der Eintrag mit führender Null wird in `New` abgewiesen (Nr. 5 a). Der Probe, dessen Orakel aus dem Wortlaut stammt, findet keine weitere Abweichung; die einzige vom Reviewer gemessene ist damit die entschiedene. Sie öffnet nichts, was der Eintrag nicht nennt: derselbe Host, derselbe numerische Port.

## Die Fix-Runde gegen die Klasse (`AGENTS.md` §3.17)

Eigene Mutationen an Elementen, die der Befund nicht nannte: 26 gegen `go test` des Pakets (`odata-recorder:deps`, Quellen schreibgeschützt) und 4 gegen den Race-Detektor (`odata-recorder:race`).

| Befund | Menge (aus dem Wortlaut) | Eigene Mutation (nicht genannt) | Ergebnis |
|---|---|---|---|
| M1 `zielPort` | jede Ziffernfolge als Port: Wert 1 bis 65535 oder nicht lesbar, auch bei Überlauf eines Zählers und mit beliebig vielen Nullen | Wächter `n > 65535` entfernt; Prüfung erst nach der Schleife; nur in den ersten sechs Ziffern; Schwelle 65534 und `>= 65535`; Überlauf `n %= 65536`, `n &= 0xFFFF`, `n = 65535` | alle rot in `TestZielPort` |
| | | **Wächter setzt `n = 0` zurück statt abzuweisen** | **grün** in `make test` (Paket `ok`); rot im Probe: `http://h:6553680`, `http://h:65536080`, `http://h:655360080`, … geben für den Eintrag `h` frei (Mutant: Port 80). Kein Text der 80 Port-Texte hat ein Präfix über 65535 mit gültigem Rest |
| | | `n >= 0` statt `n >= 1` | grün, äquivalent (Port 0 gleicht keinem Eintrag) |
| M2 `zielName` | Klammerform in beiden Richtungen | Richtung 1 allein, Richtung 2 allein, Prüfung entfernt; `Contains(host, "::")` statt `":"`; `HasSuffix(roh, "]")` statt `HasPrefix(roh, "[")` | alle rot (`TestZielName`; die beiden letzten zusätzlich in `TestHostVergleich`) |
| M3 `TestGleichzeitig` | jeder Schreibzugriff auf den Empfänger in `Entscheide`, `hostFrei`, `routeFrei` | `p.upstream = !(!p.upstream)`; `p.replayCoverage = p.replayCoverage`; `p.freigaben = p.freigaben[:len(p.freigaben)]` in `routeFrei`; `p.hosts[0].port = p.hosts[0].port` in `hostFrei` | je **10 von 10 Prozessen rot** mit `DATA RACE` (`-test.count=1`, ein Binary, zehn Prozesse); Grundlauf ohne Mutation 0 von 10 rot. Die zwei letzten Mutationen sitzen in Funktionen, die der Befund nicht nannte |
| M4 Anfang vor dem `*` | jede Prüfung von Anfang, Ende und Länge in `sternPasst` | `Contains` statt `HasPrefix`; `HasSuffix` statt `HasPrefix`; `HasPrefix` und `Contains` statt `HasSuffix` für das Ende; `+1` der Mindestlänge gestrichen | alle rot (`TestMusterVergleich`, teils mit `TestPfadOhneDekodierung`, `TestWMenge`, `TestPunktSegmente`) |
| M5 Benutzerteil | jede Form von Benutzerteil | nur Benutzer mit Namen abgewiesen; nur mit Passwort abgewiesen; leeres Passwort erlaubt | alle rot (`TestHostNichtLesbar`) |

Die von mir ergänzte Menge zu M1 („Präfix über 65535 plus gültiger Rest“, 11 Präfixe × 9 Reste × 3 Mitten × 6 Einträge × 2 Schemata = 3564 Fälle, im unveränderten Code 0 frei, im Mutanten rot) ist ein Element der Klasse „jede Ziffernfolge“, das keiner der 80 Texte und keine der 16 Randformen trifft; die Fix-Runde nennt die Menge, aber ihre Elemente sind eine Auswahl aus Zweierpotenzen und Nullen.

## Antworten auf die drei User-Fragen (nicht entschieden, nur geprüft)

Der Code folgt dem geltenden Wortlaut: `Replay` für mutierende Methoden (Nr. 4 b = ja), `ODR-E6006` kein Auslöser (Nr. 3 d). Eine andere Antwort ändert im Code je eine Zeile; die Mutationen dazu (Replay nur lesend; `ODR-E6006` als Auslöser) färben genau diese Fänger:

| Antwort | Rote Tests |
|---|---|
| Replay nur lesend | `TestReplayFallback` (P05), `TestMatrix` (P06) und `TestMethodenKlassen` (P02, weil dessen Soll aus dem Orakel von P06 kommt) |
| `ODR-E6006` Auslöser | `TestAusloeser` (P03), `TestMatrix` (P06) |

Der Plan (§6, Frage 1) nennt nur P05 und P06; P02 ändert sich mit, über das gemeinsame Orakel (Hinweis, kein Befund). Sonst ändert sich nichts. Die Lücke im Lastenheft (Frage 3) betrifft keinen Code.

## Integration

- `git merge-tree --write-tree HEAD main` (`main` = `6f68c2a3`): sauber, keine Konflikte; die Zeilen der Änderungstabelle und der Historie für `SPEC-261` bleiben erhalten. Die erwarteten Textkonflikte in der Änderungstabelle traten nicht auf; ein späterer Stand von `main` kann sie bringen.
- `git merge-tree --write-tree HEAD slice-if-match`: sauber; im Ergebnis steht `SPEC-283` (Zeile 3336) vor `SPEC-261` (Zeile 3373), ohne Konfliktmarken.
- Fehlercode-Zahl: 40 (`model.AlleCodes()`); `entscheide_test.go` und `total_test.go` halten 40, 43 (mit Nullwert, −1, N + 1), 62 608 und 774 000 als feste Zahlen. **Mit `ODR-E6010` (41 Codes, `slice-if-match`) werden diese Tests rot**; der Plan von `slice-if-match` nennt als Zählstellen fünf andere Dateien und weitere Leser von `AlleCodes()`, aber nicht dieses Paket (Auflage 3).

## Auflagen

1. **Vor der Closure (Fänger, MEDIUM nach `AGENTS.md` §3.15):** `TestZielPort` um die Familie „Ziffernfolge mit Präfix über 65535 und gültigem Rest“ erweitern (Präfixe wie 65536, 100000, 2^32, 2^64; Reste 1, 80, 443, 8443, 65535, auch mit Nullen dazwischen), die Mutation „Wächter setzt `n = 0` zurück“ rot sehen und die Menge im Plan-Absatz zu M1 nennen. Der Produktcode bleibt unverändert.
2. **Plan (`AGENTS.md` §3.9, Nachzug):** §3 sagt „Die Tabelle hat 16 Zeilen, alle der Klasse `andere`: die Policy hat keine Nebenläufigkeit“, obwohl die Fix-Runde zwei Mutationen der Klasse `nebenläufig` führt; die Zeile P22 nennt Nr. 9 c nicht (der Fänger steht nur in der Tabelle der Fix-Runde). Der Satz „die Spezifikation ist unverändert zu führen, ein Nachzug entfällt“ trifft seit `531d2ff2` und `1914ee9f` nicht mehr. §6 (Frage 1) nennt P02 nicht unter den Zeilen, die eine andere Antwort ändert.
3. **Integration (`AGENTS.md` §3.13):** wer von `slice-if-match` und diesem Slice zuletzt nach `main` geht, zieht `src/hexagon/services/hybrid/policy/entscheide_test.go` (Zeilen 56 bis 61, 158, 267 bis 268, 388 bis 418) und `total_test.go` (Zeilen 177, 213) von 40 auf 41 Codes nach (43 → 44, 62 608, 774 000 und die Zählung von Hand 1176 / 110 neu); der Planner trägt die Adresse in einen der beiden Pläne ein.
4. **Hinweis (kein Befund):** der Block nennt `produktcode: 200`, der Produktcode hat 266 Zeilen; die Marke 2000 und die Messung 1671 stimmen. Die Closure-Notiz, der Eintrag im Beobachtungs-Register und die Ausgänge „weiter offen bis zur Closure“ in §6 stehen noch aus (Aufgabe des Planner); die drei User-Fragen bleiben offen und bremsen die Closure nicht.

## Grenzen dieser Verifikation

- Mein Orakel liest das Ziel ebenfalls mit `url.Parse` (so verlangt es `SPEC-261` Nr. 5 c); gegen einen gemeinsamen Fehler von `url.Parse` schützen nur die Regex-Grammatik und der Fuzz, und die decken die Formen der Bausteine, keine unbeschränkte Menge.
- Die Raten der Race-Mutationen sind 10 von 10 Prozessen je Mutation bei Last 5 bis 24 (ein Binary je Mutation, Messender: ich); für `AGENTS.md` §3.10 (jeder Lauf rot) heißt das: keine Ausnahme in 40 Prozessen über vier Mutationen.
- Probe-Dateien und Mutationen liegen im Scratchpad der Sitzung, nicht im Repo; der Report nennt Mengen und Formen, damit der Lauf sich neu schreiben lässt.
