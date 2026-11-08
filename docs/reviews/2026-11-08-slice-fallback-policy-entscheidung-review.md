# Review slice-fallback-policy-entscheidung (Teil 2: Entscheide)

**Review-Art:** Code (gegen Plan, Spezifikation und Hard Rules)

**Gegenstand:** Zweig `slice-fallback-policy-entscheidung`, Code-Commit `fa4e6549` (Basis `56dc7c58`); der Architect-Commit `531d2ff2` danach ändert nur Spezifikation und Pläne und wurde mitgelesen. Dateien: `entscheidung.go`, `host.go`, `freigabe.go`, `policy.go` und die Tests des Pakets `src/hexagon/services/hybrid/policy`.

**Skill:** `.harness/skills/reviewer.md` · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-08 · Reviewer, ein Lauf.

**Eingangs-Kontext:**

- Plan `slice-fallback-policy-entscheidung` (Tabelle P01 bis P10, P12, P15, P16, P22 bis P24, Raster und Zusatztabelle in §6)
- `SPEC-261` Nr. 1 bis 10 (einschließlich Nr. 2 d, Nr. 5 c, Nr. 10 c, *Ungeprüft*), `LH-FA-12.a`, `LH-FA-12.b` der Spezifikation, [LH-FA-12](../../spec/lastenheft.md)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `AGENTS.md` §3.2, §3.6, §3.7, §3.9 bis §3.18
- vorheriges Review am Paket: `slice-fallback-policy` (Teil 1)

**Nicht geprüft:** die DoD (Verifier); die Gate-Kette (`make gates`) wurde nicht gefahren; die Aufrufer der Policy (kein Code bindet sie: `grep` über `src/` und `tests/` findet außerhalb des Pakets keinen Import).

---

## Zusammenfassung

HIGH 0 · MEDIUM 5 · LOW 0 · INFO 6.

Kein Befund erlaubt einen Upstream-Zugriff, den der Wortlaut nicht erlaubt. Der Code folgt `LH-FA-12.b` und `SPEC-261`; die fünf MEDIUM sind Fänger, die schmaler sind als die Zusage (`AGENTS.md` §3.15, §3.17, §3.10), je mit einer Mutation, die in allen Tests des Pakets grün bleibt.

**Sicherheitsprobe (LH-FA-12.b zuerst).** Eigenes Probe-Paket (nicht im Repo, Scratchpad) mit einem Orakel, das allein aus dem Wortlaut von `LH-FA-12.b` und `SPEC-261` Nr. 2 bis 6 gebaut ist (eigener Parser des Ziels über einen regulären Ausdruck und `net.ParseIP`, eigener rekursiver Muster-Vergleich, keine Zeile des Pakets): 616 Policies (11 Host-Mengen × 7 Freigabe-Mengen × 4 Kategorien-Mengen × Hauptschalter an und aus) × 27 Methoden (die 13 des Plans, dazu Leerraum, NUL, Unicode-Look-alikes, Klein- und Titelschreibung, `GET,POST`) × 12 Codes (Nullwert, −1, N + 5, die Auslöser, `ODR-E5002`, `ODR-E6005`, `ODR-E6006`, `ODR-E9001`, `ODR-E5007`) × 72 Ziele (Userinfo-Tricks, `@`, `#`, Backslash, Prozent-Kodierung, Zone, IPv4 in Klammern, Port-Formen mit führender Null, mit Vorzeichen, mit Überlauf auf 2⁶⁴ + 443, Steuerzeichen, Unicode) × 41 Pfade (für mutierende Methoden mit Mock-Auslöser alle, sonst zwei): **57 657 600 Fälle, 0 Abweichungen in der gefährlichen Richtung** (Policy sagt `Upstream` oder `Replay`, das Orakel `Bleiben`); die einzige Abweichung überhaupt sind 4 312 Fälle, in denen das Ziel einen Port mit führender Null trägt (`https://h:0443`) und die Policy ihn als Dezimalzahl liest (siehe I2). Hauptschalter aus gibt in keinem Fall `Upstream`; unbekannte Methoden geben in keinem Fall `Replay` oder `Upstream`; der Replay geht dem Upstream in jedem Fall vor, in dem beide offen sind.

**Mutationslauf.** `tools/harness/mutationslauf.sh`, Image `odata-recorder:test` (Go 1.27.1), `--memory=4g`, Quellen als Kopie im Container ohne Netz, unter `flock /tmp/odata-recorder-heavy.lock`; Last des Hosts beim Lauf 7,5 bis 14 bei 4 CPUs; Messender: der Reviewer. Eigene Listen, nicht die des Implementers: 70 Mutationen im Paket `policy` (64 rot, 1 rot-paket, 5 grün), dazu 12 Mutationen über ein Probe-Paket (gleicher Baum, Mutation an `policy`, Test im Probe-Paket: 5 rot, 7 grün). Die Klasse ist `andere` (deterministisch, keine Rate), außer der einen Mutation `p-schreibt-receiver-race` (Klasse `nebenläufig`, Image `odata-recorder:race`, 20 Prozesse mit Detektor: 20 von 20 rot). Eine einzelne Messung des Verhaltens von `url.Parse` (`docker run --memory=1g`, ein Container) lief zu Beginn ohne `flock`; alle späteren Läufe unter dem `flock`.

## Findings

### MEDIUM

**M1**
- kategorie: MEDIUM
- quelle: `AGENTS.md` §3.17 (Fix deckt die Klasse), §3.15; `SPEC-261` Nr. 5 c (Port 1 bis 65535)
- pfad: `src/hexagon/services/hybrid/policy/host.go` · `if n > 65535 {` (Funktion `zielPort`)
- befund: Der Überlaufwächter von `zielPort` hat keinen Fänger: ohne ihn (Mutation `r-ziel-port-ueberlauf`) bleiben alle Tests des Pakets grün, und das Ziel `https://h:18446744073709552059` (2⁶⁴ + 443, `url.Parse` nimmt jede Ziffernfolge an) wird als Port 443 gelesen und mit dem Eintrag `h` freigegeben (das Probe-Paket wird rot). Die Mutation des Implementers (`m-ziel-port-ueberlauf`, Schwelle 6 553 500) ist äquivalent, weil sie den Wächter nur verschiebt; die Behauptung „äquivalent, keine Folge im Betrieb“ gilt für diese Mutation, nicht für die Klasse „Wächter entfernt oder geändert“.
- verifizierbar: ja (Mutation `r-ziel-port-ueberlauf` grün in `TestHostNichtLesbar`, im Paket `rot=0`; ein Fall `https://h:18446744073709552059` in `nichtLesbar()` macht sie rot)
- klasse: Äquivalenz-Behauptung an abgeschwächter Mutation
- Failure-Szenario: heute keine Folge im Betrieb (`Senden` weist einen Port über 65535 ab, `SPEC-110` Nr. 7); die Policy führt die Prüfung aber als eigene Zusage ohne Fänger, und eine Änderung des Wächters bliebe grün, bis sie in `slice-hybrid-service` durchschlägt.

**M2**
- kategorie: MEDIUM
- quelle: `AGENTS.md` §3.15 (Zusage ohne Fänger); `SPEC-261` Nr. 5 c (Klammerform des Ziels)
- pfad: `src/hexagon/services/hybrid/policy/host.go` · `if strings.HasPrefix(roh, "[") != strings.Contains(host, ":") {`
- befund: Die Klammerprüfung in `zielName` ist laut Architect eine Prüfung der Policy (Entscheid 2026-11-08), hat aber im Diff keinen Fänger: die Mutation entfällt grün (Paket und Probe-Paket, 57,6 Mio. Fälle ohne Abweichung). Der Auftrag `TestZielName` an den Implementer steht in §6 als entschieden, der Test ist nicht im Diff.
- verifizierbar: ja (Mutation `r-ziel-klammer-pruefung-fehlt` grün)
- klasse: Zusage ohne Fänger (Prüfung, die die Standardbibliothek zurzeit doppelt)
- Messung zur Äquivalenz (Image Go 1.27.1): `url.Parse` weist `http://[1.2.3.4]`, `http://[h]` und `http://[]` ab, mit der go-Zeile 1.27 wie mit 1.25 und mit `GODEBUG=urlstrictcolons=0`; `http://::1` und `http://h:80:80` weist es mit der go-Zeile 1.27 ab, mit 1.25 und mit `urlstrictcolons=0` nimmt es sie an (Hostname `:` bzw. `h:80`), und `netip.ParseAddr` lehnt beide ab. Die Mutation ist also in allen drei gemessenen Konfigurationen verhaltensgleich; die Abhängigkeit von der Version der Standardbibliothek (nicht gemessen für ältere als 1.25) bleibt ohne den Fänger unentdeckt. Der Auftrag ist bekannt und nicht herabgestuft: bis der Test im Diff steht, ist es eine Zusage ohne Fänger.

**M3**
- kategorie: MEDIUM
- quelle: `AGENTS.md` §3.10 (Zusage über Nebenläufigkeit), §3.11; `SPEC-261` Nr. 9 c
- pfad: `src/hexagon/services/hybrid/policy/entscheidung.go` · `ändert nichts (SPEC-261 Nr. 4, 9)` (Kommentar von `Entscheide`) und die Zusage „für gleichzeitigen Aufruf sicher, weil es nichts schreibt“
- befund: Kein Test ruft `Entscheide` aus zwei Goroutinen (`grep` nach `go func`, `sync.`, `t.Parallel` im Paket: kein Treffer), und `TestQuelltext` prüft nur Importe, `go`-Statements, Variablen auf Paketebene und Empfänger der Methoden, keine Zuweisung an ein Feld des Empfängers. Die Mutation `r-schreibt-receiver` (`p.upstream = !(!p.upstream)` zu Beginn von `Entscheide`) bleibt in allen Tests grün; derselbe Schreibzugriff wird mit einem Test aus vier Goroutinen unter `-race` in 20 von 20 Prozessen rot (`p-schreibt-receiver-race`). Der Plan nennt für Nr. 9 c weder einen Wirkungs-Fänger noch ein akzeptiertes Negativ mit Grund und Folge (§6: „entfallen: reine Funktion ohne Zustand“).
- verifizierbar: ja (Mutation grün; ein Test mit Goroutinen unter `make test-race` wird rot)
- klasse: Zusage über Nebenläufigkeit ohne Lauf unter dem Detektor
- Failure-Szenario: ein späterer Cache oder Zähler im Empfänger, der Werte nicht ändert (kein Fall von P22 sieht ihn), ist ein Datenrennen im Hybrid-Dienst, der die Policy aus mehreren Anfragen ruft.

**M4**
- kategorie: MEDIUM
- quelle: `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe), §3.17; `SPEC-261` Nr. 6 c
- pfad: `src/hexagon/services/hybrid/policy/freigabe.go` · `!strings.HasPrefix(t, s.anfang)`
- befund: Die Pfadmenge von P15 (1 891 Pfade aus zwölf Segmenten) enthält kein Segment, das den Anfang eines Musters (`a`, `Orders(`) enthält, aber nicht damit beginnt. Die Mutation `r-anfang-contains` (`strings.Contains` statt `HasPrefix` für den Anfang vor dem `*`) bleibt in allen Tests des Pakets grün, die Gegenseite `r-ende-contains` wird rot (Pfad `aba` gegen `*b`). Mit ihr gibt das Muster `PUT /svc/Orders(*)` den Pfad `/svc/XOrders(1)` frei, das Muster `/svc/a*b/x` den Pfad `/svc/xab/x` (Probe-Paket rot).
- verifizierbar: ja (Mutation grün; ein Pfad mit Anfang an zweiter Stelle, etwa das Segment `xa`, macht sie rot)
- klasse: Generator deckt nur eine Seite einer Grenze
- Failure-Szenario: eine Routenfreigabe gäbe mehr frei, als der Betreiber las; heute nicht eingetreten, der Code ist richtig.

**M5**
- kategorie: MEDIUM
- quelle: `AGENTS.md` §3.15 (Fänger führt nur eine Form), `SPEC-261` Nr. 5 c (kein Benutzerteil)
- pfad: `src/hexagon/services/hybrid/policy/host.go` · `if err != nil || u.User != nil {`
- befund: `nichtLesbar()` führt zwei Formen mit Benutzerteil (`http://u@h`, `http://@h`), keine mit Passwort. Die Mutation `r-user-passwort-erlaubt` (ein Benutzerteil gilt, wenn es ein Passwort trägt) bleibt grün; `http://u:p@h` und `http://:p@h` werden dann freigegeben (Probe-Paket rot).
- verifizierbar: ja (Mutation grün; die Ziele `http://u:p@h` und `http://:p@h` in `nichtLesbar()` machen sie rot)
- klasse: Fänger führt nur eine Form der Eingabe
- Failure-Szenario: dieselbe wie M1; `Senden` weist Zugangsdaten ab (`SPEC-110` Nr. 7), die Policy trägt die Zusage allein im Test.

### INFO

**I1**
- kategorie: INFO
- quelle: `AGENTS.md` §3.15; Plan §6 (P01, akzeptiertes Negativ a)
- pfad: `src/hexagon/services/hybrid/policy/policy.go` · `type Policy struct {`
- befund: Die Sicherheitsfolge von P01 (a) ist heute null: `Policy` hält nur geparste Typen (`[]hostEintrag` mit `string` und `int`, `[]freigabe` mit Strings und Segmenten, zwei `bool`) und keinen Verweis auf eine `[]string` der Werte; Strings sind unveränderlich, eine Änderung der übergebenen Slices wirkt nicht. Grund und Folge im Plan stimmen. Der Fänger `TestListenSindMengen` („überschrieben und gekürzt“) ist für diese Form wirkungslos, fängt aber eine künftige Bauweise, die eine Liste hielte; ob er es täte, ist ungemessen (die Mutation lässt sich ohne Umbau nicht schreiben).
- verifizierbar: nein
- klasse: akzeptiertes Negativ mit Grund und Folge, bestätigt

**I2**
- kategorie: INFO
- quelle: `SPEC-261` Nr. 5 c, Entscheid des Architect 2026-11-08 (IPv6 nur in Klammern, Schema in Großschreibung, Port mit führender Null)
- pfad: `src/hexagon/services/hybrid/policy/host.go` · `zielPort liest den Port eines Ziels: eine Dezimalzahl von 1 bis 65535, auch mit führender Null`
- befund: Sicherheitsseite bewertet: (1) IPv6 nur in Klammern, ein Name nie in Klammern: weniger frei, nie mehr; `url.Parse` weist die Formen heute selbst ab (M2). (2) Das Schema `HTTP` gilt, weil `url.Parse` es auf `http` faltet; derselbe Leser ist `net/http`, der Host und die Wahl des Ports sind dieselben. (3) `https://h:0443` ist Port 443 für `url.Parse` und für `net.Dial` (Dezimalzahl); der Eintrag dagegen verlangt Ziffern ohne führende Null, das Ziel nicht. Die Folge ist kein Host außerhalb der Allowlist: der Name bleibt gleich, nur ein Leser mit anderer Lesart des Ports (etwa ein Proxy, der `0443` oktal liest) erreichte einen anderen Port desselben freigegebenen Hosts. Das ist die Abweichung der Leser, die *Ungeprüft* nennt (Fänger bei `slice-hybrid-service`). Im Probe-Paket sind es 4 312 von 57,6 Mio. Fällen, alle dieser Art. Die Wahl ist vertretbar; „im Zweifel nichts freigeben“ spräche für die strengere Lesart, der Betreiber setzt `UPSTREAM_URL` aber selbst, ein Angreifer nicht.
- verifizierbar: ja (Probe-Paket)
- klasse: Randform des Ziel-Lesers, Entscheid bestätigt

**I3**
- kategorie: INFO
- quelle: `AGENTS.md` §3.15; Plan §3 (acht äquivalente Mutationen)
- pfad: `docs/plan/planning/done/slice-fallback-policy-entscheidung.md` · `**Äquivalente Mutationen** (8, grün, Grund und Folge)`
- befund: Von den acht sind sechs äquivalent und mit dem Probe-Paket gemessen (je grün über 57,6 Mio. Fälle, Mutationen des Implementers unverändert): `m-w-schraegstrich`, `m-ziel-port-null-gueltig`, `m-ziel-host-leer`, `m-ziel-namenszeichen`, `m-ziel-zone-erlaubt`, `m-ziel-ipv6-nur-adresse`. `m-host-klammer-pruefung-fehlt` ist zurzeit äquivalent und hängt an der Standardbibliothek (M2). `m-ziel-port-ueberlauf` ist es nur in der gewählten Form (M1). Eine Zeile Plan sagt „Folge: keine im Betrieb“ für alle acht; für M1 und M2 gilt das nur, solange `url.Parse` und `Senden` dieselbe Lesart halten.
- verifizierbar: ja
- klasse: Äquivalenz-Behauptung gemessen

**I4**
- kategorie: INFO
- quelle: Maintainability
- pfad: `src/hexagon/services/hybrid/policy/freigabe.go` · `func mutierend(m string) bool` und `entscheidung.go` · `case "POST", "PUT", "PATCH", "DELETE", "MERGE":`
- befund: Die Liste der fünf mutierenden Methoden steht zweimal (Prüfung des Eintrags und Klasse der Anfrage). Weicht eine von der anderen ab, ist das Ergebnis die sichere Seite (eine Freigabe, die nie wirkt, oder ein Eintrag, der abgelehnt wird); `TestMethodenKlassen` und `TestFreigabeEintraege` fangen beide Listen getrennt.
- verifizierbar: nein
- klasse: doppelte Liste einer Methodenmenge

**I5**
- kategorie: INFO
- quelle: `AGENTS.md` §3.10, Hinweis des Skills zu `panic`
- pfad: `src/hexagon/services/hybrid/policy/entscheidung.go` · `if p == nil {`
- befund: Die Mutation `r-nil-pruefung-fehlt` ist rot durch eine Panik, nicht durch ein `--- FAIL` mit Meldung. Das ist hier der benannte Grund: P22 sagt zu, dass `Entscheide` nie in Panik gerät; die Panik ist der Bruch der Zusage.
- verifizierbar: ja
- klasse: Rot durch Panik, Zusage ist die Panikfreiheit

**I6**
- kategorie: INFO
- quelle: `SPEC-261` Nr. 5 c, `SPEC-110` Nr. 7 (Rolle: Architect, Planner)
- pfad: `src/hexagon/services/hybrid/policy/host.go` · `func zielName`
- befund: `UPSTREAM_URL` darf nach `SPEC-110` Nr. 7 Bytes ab `0x80` und die Zone einer IPv6-Adresse tragen; die Policy liest beides als nicht lesbar und gibt dafür nie einen Fallback frei (auch der Eintrag kennt beides nicht). Das ist die sichere Seite und von Nr. 5 a gedeckt; der Betreiber, der einen internationalen Hostnamen nutzt, muss die ASCII-Form (Punycode) setzen, und das Handbuch (`slice-bootstrap-m4`) nennt es nicht in den Sendungen.
- verifizierbar: ja
- klasse: Einschränkung ohne Hinweis im Handbuch

## Auflösung der Tabelle und des Rasters (je Zeile eigene Mutationen)

Die Mutationen stehen in den Listen des Reviewers (Scratchpad `rev-q7k2`: `liste-r.json`, `liste-r2.json`, `liste-p2.json`, `liste-p3.json`); „rot“ heißt rot im benannten Test, sofern nicht anders gesagt.

| Zeile | Eigene Mutationen (rot, sofern nicht anders) | Befund |
|---|---|---|
| P01 | `s-host-nur-erster` (`TestHostListeAlsMenge`), `s-freigabe-nur-erste` (`TestListenSindMengen`); die Form „Slice ohne Kopie“ nicht schreibbar (I1) | I1 |
| P02 | `r-head-nie`, `r-put-nie`, `r-patch-nie`, `r-delete-nie`, `r-post-nie`, `r-unbekannt-lesend`, `s-get-klein-lesend`, `s-leer-lesend`, `s-options-lesend`, `s-merge-klein-mutierend` | ohne Befund |
| P03 | `r-e6002-nicht`, `s-e6006-ausloeser`, `s-nullwert-ausloeser`, `s-e5002-ausloeser` | ohne Befund |
| P04 | `s-feld-header` | ohne Befund |
| P05 | `r-e5001-als-coverage`, `r-e6003-als-unsupported`, `r-kategorie-beliebig`, `r-replay-unsupported-braucht-schalter`, `r-replay-coverage-braucht-schalter`, `s-replay-ziel-leer` | ohne Befund |
| P06 | `r-host-pfad-statt-ziel`, `r-host-nur-mutierend`; Probe-Paket mit eigenem Orakel über 57,6 Mio. Fälle | ohne Befund (Orakel unabhängig: Tabelle der Zeilen von `LH-FA-12.b` mit von Hand gezählten 1 176 und 110) |
| P07 | `r-route-methode-fest`, `r-route-ziel-statt-pfad`, `r-route-nur-e6001` | ohne Befund |
| P08 | `r-miss-mutierend-offen` | ohne Befund |
| P09 | `s-replay-nach-upstream` | ohne Befund |
| P10 | `s-schalter-nur-lesend`, `s-schalter-nur-mutierend` | ohne Befund |
| P12 | `r-host-oder`, `r-standard-fest`, `r-http-443`, `r-https-80`, `r-ziel-schreibweise`, `r-eintrag-ipv6-textuell`, `r-ziel-ipv6-textuell`, `s-host-teilstring`, `s-host-endung`, `s-host-port-ignoriert`, `s-schema-ftp`, `s-leerer-port-ungueltig`; grün: `r-user-passwort-erlaubt` (M5), `r-ziel-port-ueberlauf` (M1), `r-ziel-klammer-pruefung-fehlt` (M2) | M1, M2, M5 |
| P15 | `r-rest-min-zwei`, `r-stern-plus-zwei`, `r-ende-contains`, `r-rest-nur-erstes`, `r-rest-nur-letztes`, `r-gedeckt-leer`, `r-pfad-trailing-slash`, `s-pfad-ohne-slash`; grün: `r-anfang-contains` (M4) | M4 |
| P16 | `r-rest-erstes-ungeprueft`, `r-gedeckt-ohne-punkt`, `r-gedeckt-ohne-punktpunkt`, `r-stern-ganz-ohne-w`, `r-stern-teil-ohne-w`, `s-w-prozent`, `s-w-backslash`, `s-w-frage`, `s-w-raute`, `s-w-untergrenze`, `s-w-obergrenze` | ohne Befund |
| P22 | `r-nil-pruefung-fehlt` (rot durch Panik, I5); grün: `r-schreibt-receiver` (M3) | M3, I5 |
| P23, P24 | akzeptierte Negative mit Grund und Folge; P23: kein Import der Policy außerhalb des Pakets (gemessen) | ohne Befund |

Raster §6: Die Zellen „Abbruch“ gelten nicht (reine Funktion ohne `ctx`). Die Zelle „Wiederholung“ (`SPEC-261` Nr. 9 b) fängt `TestTotalitaet` für Wirkungen auf das Ergebnis; ein Schreibzugriff ohne Wirkung auf das Ergebnis bleibt ungefangen (M3). Die Zellen „fehlender, leerer, fremder Wert“ der vier Operationen sind durch die Zeilen P03, P05, P06, P12, P15, P16, P22 und die Mutationen oben gefangen, ausgenommen die in M1, M2, M4 und M5 genannten Formen. Die Zusatztabelle (Hauptschalter) ist durch `s-schalter-nur-lesend`, `s-schalter-nur-mutierend`, `r-replay-unsupported-braucht-schalter` und `r-replay-coverage-braucht-schalter` gefangen.

Mengen auf Unabhängigkeit des Orakels: P06 (62 608 Fälle, Orakel als Tabelle der Matrix mit von Hand gezählten Ergebnissen) und P12 (Soll je Paar von Hand) sind vom Code unabhängig. P15 (456 × 1 891) nutzt ein Orakel aus regulären Ausdrücken über der Klasse W, unabhängig geschrieben, in der Form dem Algorithmus nahe, mit fester Trefferzahl 13 017 gegen ein Orakel, das mit dem Code zugleich bricht; die Lücke der Pfadmenge steht in M4.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/hexagon/services/hybrid/policy/entscheidung.go` (Klassen, Auslöser, Reihenfolge der Regeln, nil und Nullwert) | geprüft, ohne Befund außer M3, I5 |
| `src/hexagon/services/hybrid/policy/host.go` (Eintrag, Ziel, Vergleich) | geprüft, Befund M1, M2, M5, I2, I6 |
| `src/hexagon/services/hybrid/policy/freigabe.go` (Muster-Vergleich, Menge W, Punkt-Segmente) | geprüft, Befund M4 |
| `src/hexagon/services/hybrid/policy/policy.go` (Paketkommentar) | geprüft, ohne Befund: im Indikativ, in Sätzen, nennt `model` als erlaubten Import, nichts über abwesenden Text (`AGENTS.md` §3.7) |
| Tests des Pakets (`entscheide_test.go`, `vergleich_test.go`, `total_test.go`) | geprüft, Befund siehe M1 bis M5; Rohbytes: keine Folge aus Backslash, `u` und vier Hexziffern, kein BOM, kein NBSP in `entscheidung.go`, `host.go`, `freigabe.go`, `policy.go` und den Tests (`LC_ALL=C grep -P`) |
| `//nolint`, Lockerung eines Gates (`AGENTS.md` §3.2, §3.6) | geprüft, ohne Befund: keine Direktive, keine Änderung an `.golangci.yml` oder Gates im Diff |
| Plan §3, §6 (Folgepflicht `AGENTS.md` §3.9, §3.13) | geprüft, ohne Befund: Abweichung (P15, P12, P16, P02) und Rückgabe an den Architect stehen mit Ausgang in §6 |
| Aufrufer der Policy | geprüft, ohne Befund: kein Import außerhalb des Pakets, die Sendungen P23 und P24 haben Nehmer |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 5 |
| LOW | 0 |
| INFO | 6 |

**Finding-Klassen dieses Laufs:** Äquivalenz-Behauptung an abgeschwächter Mutation · Zusage ohne Fänger (Prüfung, die die Standardbibliothek doppelt) · Zusage über Nebenläufigkeit ohne Lauf unter dem Detektor · Generator deckt nur eine Seite einer Grenze · Fänger führt nur eine Form der Eingabe

## Verdikt

**Merge-blockierend:** ja für M1, M2, M3, M4 und M5 nach `AGENTS.md` §3.15 (eine Zusage ohne Fänger ist ein MEDIUM); kein Befund ändert das Verhalten des Codes, jeder ist ein zusätzlicher Test oder ein Eintrag als akzeptiertes Negativ mit Grund und Folge. Der Code selbst bleibt in der Sicherheitsprobe ohne Abweichung in der gefährlichen Richtung.

**Übergabe:** Die Findings gehen an den Implementer; M2 trägt bereits einen Auftrag des Architect (`TestZielName`). Die Finding-Klassen gehen zusätzlich in die Slice-Closure §7.
