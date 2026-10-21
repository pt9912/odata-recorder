# Nachverifikation: slice-upstream-adapter

* Datum: 2026-10-21
* Rolle: verifier (frischer Kontext; kein Code, keine Spezifikation, kein Plan, kein Lastenheft, keine `AGENTS.md` geändert)
* Gegenstand: Spitze `015da8c`, `git diff 7a8fde7..015da8c`: der Architect (`daad9e8`: Entscheidungen zu V5, V9, I-4, V3 und V4 in `SPEC-120` Nr. 1, 2, 4, 8 und `SPEC-110` Nr. 7 und *Ungeprüft*, Plan nachgezogen) und der Implementer (`015da8c`: Fänger zu V1 bis V7 und V9, vier Produktkommentare, Plan, Stempel der Gegenproben) gegen die Befunde V1 bis V12 der Verifikation `docs/reviews/2026-10-21-slice-upstream-adapter-verifikation.md` (`7a8fde7`). Produktlogik unverändert: die 17 Zeilen von `git diff 7a8fde7..015da8c -- src/ ':!*_test.go'` sind Kommentare, und `client.go`, `fehler.go`, `antwort.go`, `leser.go`, `rahmen.go`, `frist.go` sind ohne ihre Kommentarzeilen byte-gleich; `src/bootstrap/` ist unberührt. Der TLS-Teil (`0716718`) ist nicht Gegenstand, nur sein Cherry-Pick (Abschnitt 8).
* Plan: `docs/plan/planning/done/slice-upstream-adapter.md`
* Maßstab: DoD (§2), Tabelle *Zusage · Fänger* (§3) und Raster (§6) gegen `SPEC-120` Kopf, Nr. 1 ohne `CAPEM`, Nr. 2 bis 6, 8, 9 und *Ungeprüft*, `SPEC-110` Nr. 7, 12 und *Ungeprüft*; Lastenheft [LH-FA-01](../../spec/lastenheft.md), [LH-RB-02](../../spec/lastenheft.md), [LH-QA-07](../../spec/lastenheft.md) (nur gelesen); `AGENTS.md` §3.7, §3.9 bis §3.16; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0002](../plan/adr/0002-architektur-gate-a-check.md), [ADR-0003](../plan/adr/0003-lint-profil-golangci.md), [ADR-0005](../plan/adr/0005-gegenproben-ausserhalb-der-gate-kette.md)
* Belege selbst gefahren: 435 Mutationen auf Kopien von `src/` (Abschnitte 1 bis 6; 414 rot, 21 grün; darunter 6 Wiederholungen mit anderem `-run` zum Zählen der Fälle von A14 und 5 Mutationen mit geänderten Testfällen als Probe), davon 151 zusätzlich auf dem Stand `7a8fde7` (Differenzlauf, Abschnitt 3), dazu zwölf Sondenläufe zur Grenze der Frist (Abschnitt 4); `go test -count=50` über `httpupstream`, 100 Läufe mit `GOMAXPROCS=1`, 30 mit `--cpus=0.05`, die Raten von A4 (c) und A15, `make test` und `make lint` ungecacht, `make gegenproben`, `git merge-tree` samt Bau und Test des Ergebnisses, `git diff --numstat`, `make docs-check`, `make gates` (Abschnitt 10). Jedes `go test` der Mutanten lief im Image `odata-recorder:deps-mut` (Stufe `deps`, `golang:1.27-alpine@sha256:8a5910…`, Go 1.27.1), `--network none`, Mutanten in Kopien aus `git archive` im Scratch-Verzeichnis, nie im Arbeitsbaum; der Container ist entfernt, `git status` zeigt vor und nach den Läufen nur die fremde Änderung an `.claude/agents/architect.md`.

## Urteil

**Freigegeben: kein HIGH, kein MEDIUM.** Die drei MEDIUM der ersten Verifikation (V1 bis V3) sind behoben: jede Mutation aus dem Bericht, von mir nachgebaut, ist aus dem benannten Grund rot (Abschnitt 1), und die Zahlen des Implementers stimmen (256 von 256 Token-Mutanten, bei V2 und V5 mehr Mutanten rot als gemeldet, die Klemmen des Timers auf 30 min, 24 h und 29 Tage rot). Die Umbauten an den Hilfen der Fälle (`mitGeheim` setzt `Methode` unbedingt, `setzeZiel` fügt den Marker ein, `fakeKopfMit` trägt Status-Text und Header, der Kopf 101 des Listeners trägt Marker in Status-Text, Header-Werten und den Bytes danach) machen keinen Fall blind: 151 Mutanten, auf `7a8fde7` und auf `015da8c` gefahren, kein einziger wechselt von rot nach grün (Abschnitt 3). `make test` und `make lint` ungecacht, `make gates` und der Cherry-Pick sind grün.

* LOW (Implementer): V13 die Zusage „ohne Obergrenze“ der `Frist` trennt A8 (f) nur bis 30 Tage; ein Wert bis rund 260 Jahre trennt ohne Kosten mehr (Abschnitt 4). V14 zwei Zellen der Allaussage A14 ohne Marker: die Namen der Antwort-Header und der Benutzername in den Zugangsdaten von `Ziel`.
* INFO: V15 (Implementer) die Begründung im Test, drei Fälle der Rahmung trügen den Marker in `Methode` nicht, trägt nicht. V16 (Implementer) „der Fehler nennt das Feld“ hat keinen Fänger.
* Offen beim Planner, unverändert: V8 (I-4 hat keinen Träger in §6), V10 (`slice-upstream-tls` §1 nennt A19 nicht).

Von den ersten Befunden sind V1, V2, V3, V5, V6, V9, V11 und V12 gelöst; V4 gelöst bis auf den Rest V13, V7 bis auf die Zellen von V14 (Abschnitt 7 und Tabelle *Stand der Befunde*); die DoD steht in Abschnitt 12. Von den 21 grünen Mutanten sind 7 erlaubt, gleichwertig oder künstlich und ohne Finding (Abschnitt 5), einer trägt V16 und 13 tragen V13 und V14; zu 12 dieser 13 trennt eine Probe die Lücke und ist auf dem Code grün (die Klemme auf 290 Jahre liegt über der Grenze der Blase, Abschnitt 4).

## 1. Findings der ersten Verifikation, je mit eigener Mutation

Ein Go-Programm im Image kopiert die Quelle je Mutante, ersetzt genau eine Stelle (trifft die Ersetzung nicht genau einmal, bricht es ab) und fährt `go test -count=1` über `httpupstream` oder `bootstrap`. Gezählt wird Rot nur aus dem benannten Grund: je Mutante prüft die Auswertung, dass die Fehlschläge den benannten Test nennen und kein Build-Fehler vorliegt.

| Finding | Mutationen (eigene Kennungen) | Ergebnis |
|---|---|---|
| V1 | `ungueltig("Ziel")` nennt das ganze `Ziel` (v1a), `url.URL.Redacted()` (v1b), den Pfad aus `url.URL` (v1c), den Pfad aus dem Text (v1d), den Text des Fehlers von `url.Parse` (v1e) | alle fünf rot in `TestFehlerOhneWerte/Ziel …` (28, 19, 19, 26 und 9 Fälle); Host und Port aus `Ziel` im Fehler (v1f) und an jedem Fehler (v1h): grün, erlaubt |
| V2 | je Zweig von `ergebnis` (101, `ctx`, Frist) Status, ganzer Header, Wert `X-K`, Body im Text (12), bei 101 Wert von `Upgrade` und Status samt `Upgrade` (2), Text des Fehlers der Bibliothek verpackt und mit `http.Client` (2) | 16 von 16 rot: `ctx` an c, d, h von A16; Frist an A8 (d) und i von A16; 101 an der Antwort 101 des Listeners und an A18 (g); der Body nur an den Fällen, deren Body den Marker trägt (wie der Plan sagt) |
| V3 | `ErrUngueltig` für `Laenge` über 1 GiB, über 1 KiB, über 2^62, ab `math.MaxInt64`, ab `math.MaxInt64 - 1`; `ContentLength` als `int32` abgeschnitten | 6 von 6 rot in `TestRahmung/Laenge_math.MaxInt64_mit_ab` (der Fehler zuerst, kein Warten an der Zeitgrenze); die `int32`-Mutation auch in `TestFehlerOhneWerte` |
| V4 | `New` weist `Frist` über 24 h, 1 h, 30 d, 1000 d ab; Timer auf 30 min, 24 h, 29 d, 30 d geklemmt; Frist ±1 ns; `Frist + 1 ms` (Überlauf bei `math.MaxInt64`) | `New`: 4 von 4 rot (`TestNewArgumente` und A8 (f)); Klemmen bis 30 d und die drei Verschiebungen rot; Klemmen über 30 d grün: V13 |
| V5 | Adapter: erneut lesen mit `%25` (a), nur `%7F` (b), ohne Klammern (c), `%` im Host abgewiesen (d), `%8` im Text abgewiesen (e), IPvFuture angenommen (f), `[127.0.0.1]` angenommen (g), nur `%41` angenommen (h); Start: a, b, c | 8 von 8 rot in `TestSendenArgumente/Ziel …`, 3 von 3 rot in `TestKonfigURL`; dazu A19 in beiden Lesern (Prüfung entfernt, je ein Byte weniger, auf den Text, auf `u.Host`, ohne Auflösen): 8 und 6 rot |
| V6 | jedes der 256 Bytes als Token angenommen (179) oder abgewiesen (77), je in `istToken`; Aufrufer-spezifisch: Header-Namen ungeprüft, `Methode` nur nichtleer, leerer Text als Token | 259 von 259 rot (Abschnitt 2) |
| V9 | `Methode` im Fehler der ungültigen `Methode` (a), an jeden Fehler von `pruefe` (b), an jeden Fehler von `ergebnis` (c), an den Fehler des vorab beendeten `ctx` (d), an jeden Fehler (e), nur bei `CONNECT` (f) | 6 von 6 rot (Abschnitt 2) |
| V11 | die acht grünen Mutanten des Plans §3 | Abschnitt 5 |

## 2. Token-Mutanten und Zahlen des Plans zu A14

**256 von 256.** Je Byte genau eine Ersetzung in `istToken` (angenommen: `continue`, abgewiesen: `return false`), Lauf `-run TestTokenZeichen`. In jedem der 256 Läufe sind die zwei Untertests dieses Bytes (`Methode 0xNN`, `Header-Name 0xNN`) rot und der Aufbau hält; bei `G` und `T` (die Ränder jedes Falls) und bei `E` sind es mehr, wie der Plan sagt. Die 179 und 77 aus A2 stimmen: 256 minus die 77 Token-Zeichen von RFC 9110 §5.6.2.

**A14, `TestFehlerOhneWerte` über 119 Fälle**, je Mutation nachgezählt (Zeilen des Plans in A14):

| Mutation | Plan | gemessen |
|---|---|---|
| `Methode` an jeden Fehler | 115 von 119, grün: leere `Methode` und drei Fälle der Rahmung | 115 rot; grün genau `Methode ""`, `Laenge 3 mit ab`, `Laenge 3 mit abcd`, `Laenge math.MaxInt64 mit ab` |
| `Methode` an den Fehler der ungültigen `Methode` | jeder Fall mit nichtleerem Wert, auch `CONNECT` | 12 rot, die 12 Fälle von `Methode` mit nichtleerem Wert |
| `Ziel` an jeden Fehler | 118 von 119, grün: leeres `Ziel` | 118 rot, grün nur `Ziel ""` |
| `url.URL.Redacted()`, Pfad aus `url.URL` | je 19 | je 19 |
| Pfad aus dem Text hinter `://` | 26 | 26 |
| ganzes `Ziel` im Fehler des Ziels | 28 | 28 |

## 3. Differenzlauf: keine Blindheit durch die Umbauten

Gleiche Mutanten, gleiche Ersetzungen, auf `7a8fde7` und auf `015da8c`. 151 Ansätze: die Prüfungen der Argumente (Byte-Prüfung je Grenze, Port, Schema, Host, Zugangsdaten, Query über die Felder von `url.URL`, `CONNECT`, `Body` nil, `Laenge`, Header-Werte je Byte), A19 in beiden Lesern, die Rangfolge (sechs Umordnungen und Auslassungen von `rang`, acht von `ergebnis`; danach die zwei übrigen Umordnungen von `rang`, ebenfalls rot), die Rahmung und die Frist (Ergebnis des Beendens übergangen, ohne `chunked`, ohne `Content-Length`), A14 (Pfad, Query, Header, Text von `errBody`, Pfad und Query am vorab beendeten `ctx`) und die Mutanten aus Abschnitt 1. Auf `7a8fde7`: 57 grün; auf `015da8c`: 136 rot und 15 grün. **Von rot nach grün wechselt keiner.** Die 15 grünen: v1f und v1h (Host und Port, erlaubt), V13 (neun Klemmen), V14 (drei Namen, ein Benutzername). Zwei Mutanten sind über einen Panic rot, nicht über eine Zusicherung (`ctx` nil und `Body` nil ohne Prüfung): bekannt als I-2 des Reviews, im Plan zu A2 benannt. Ein Ansatz brach mit Build-Fehler ab (`if false` ließ `i` ungenutzt), wurde korrigiert wiederholt und ist rot aus dem benannten Grund.

Weitere Mutationen zu den Fällen der Rahmung und von `Methode` (Auftrag Punkt 2): Rahmung bei -1 der Bibliothek überlassen, `ContentLength` nicht gesetzt, `io.LimitReader` um den Body, `Content-Length` nur bei `POST`: je rot in `TestRahmung`; `CONNECT` erlaubt oder ohne Beachtung der Schreibweise, die Methode `S3CR3T-m` als ungültig: rot; Kontext vor den Argumenten, Schema ungeprüft: rot.

## 4. Die Grenze der Frist (Auftrag Punkt 3)

**Messung.** Auf dem Code: A8 (f) mit dem Kopf nach W ist grün für W von 30 Tagen bis 260 Jahre (gemessen 30 d, 31 d, 1, 10, 50, 100, 146, 200, 250, 255, 260 Jahre; nicht committet, `antwortVorAblauf(t, math.MaxInt64, W)`) und rot ab 265 Jahren, schon auf dem Code: die Uhr der Blase beginnt am 2000-01-01, das sind 30 Jahre nach der Epoche, und die Messung passt zu 292,47 minus 30 gleich 262,47 Jahren (`math.MaxInt64` ns). Das ist die Grenze der Testumgebung; die Zusatzzeile *Annahme* zu `synctest` (30 Tage) nennt sie nicht.

| Klemme des Timers | A8 (f) mit 30 d | Probe mit W bis 250 Jahre |
|---|---|---|
| 30 min, 24 h, 29 d | rot | nicht gefahren: A8 (f) trennt schon |
| 30 d | rot, aber nur über den Gleichstand (Kopf und Ablauf im selben Augenblick, die Zusatzzeile *Annahme* zu A8 (e)) | rot |
| 30 d + 1 ns, 31 d, 365 d, 10, 50, 100, 146, 200 Jahre | grün | rot (Probe mit 250 Jahren) |
| 290 Jahre | grün | grün: die Blase trägt W über 262 Jahre nicht |

**Ist das ein Finding?** Ja, nach §3.15 (beide Seiten jeder Grenze, die Fixture gleicht nie dem Wert der Mutation): die Zusage „ohne Obergrenze bis `math.MaxInt64` ns“ (`SPEC-120` Nr. 1) gilt für jeden Wert bis dahin, `UPSTREAM_TIMEOUT` lässt nach `SPEC-110` Nr. 4 jeden Wert zu, der in 63 Bit passt (`106751d`), und der Wert 30 Tage ist ein Beispiel dieser Stelle, nicht ihre Grenze. Eine Klemme auf ein Jahr (365 d) ist eine naheliegende Mutation und bleibt grün; die Klemme auf 30 d fällt nur über den Gleichstand. Klasse: *Zusage über das Fehlen einer Grenze, Fänger am Wert, den die Umgebung ohne Kosten überbietet* (V3 und V4 der ersten Verifikation in schmalerer Form). Kategorie **LOW**, wie V4 vorher: der Code erfüllt die Zusage (die Sonden sind grün), die Wirkung ist eine Frist, die ein Betreiber mit mehr als 30 Tagen nicht bekäme.

**Vereinbar mit dem Wert 30 Tage des Architect?** Die Entscheidung (4) nennt den Grenzwert `math.MaxInt64`, nicht 30 Tage; 30 Tage kommen aus A8 (f) im Plan, der dort „auch bei 365 Tagen“ misst. Ein Wert bis 260 Jahre widerspricht keiner Zeile und kostet nichts (virtuelle Zeit, 0,00 s). Er trennt alle Klemmen unter dem Wert; ungetrennt blieben nur Klemmen zwischen 262 und 292 Jahren, die die Testumgebung nicht erreicht (das gehört als Grenze in die Zusatzzeile *Annahme*). Die Zusatzzeile müsste dann nennen: Uhr der Blase ab 2000-01-01, W höchstens 262 Jahre.

## 5. Die bewusst grünen Mutationen (Auftrag Punkt 4)

Jede im Plan §3 genannte Mutation, mit eigenem Lauf auf einer Kopie:

| Mutation (Kennung) | Lauf | Urteil |
|---|---|---|
| `0xC3` allein im Pfad von `Ziel` abgewiesen (g-0xC3) | grün; in allen drei Eingaben abgewiesen (ra-eqC3, `ohneTrenner`): rot an `0xC3 0xA4` in `Pfad` und `RawQuery`, im Pfad von `Ziel` führt kein Fall das Paar | künstlich: `0x80` und `0xFF` im Pfad von `Ziel` decken beide Enden der Bytes ab `0x80`, jedes Byte geht durch dieselbe Schleife; wie in der ersten Verifikation ohne Finding |
| `0x7F` im Pfad von `Ziel` in `ohneTrenner` angenommen (g-7F) | grün; zusätzlich vor `url.Parse` aus `Ziel` genommen: rot, an `/a\x7fb` | gleichwertig: `url.Parse` weist `0x7F` im Pfad selbst ab |
| Prüfung von `ctx` vor `RoundTrip` entfernt (g-ctx) | grün | gleichwertig: `RoundTrip` endet vor dem Verbinden mit dem Kontextfehler, der Leser liest nicht, `rang` liefert denselben Fehler |
| `Header` der Antwort `nil` statt leerer Map (g-header-nil) | grün | gleichwertig nach dem Port („Header ohne Einträge“) |
| Host und Port aus `Ziel` an den Fehler (v1f, v1h) | grün | von der Zusage erlaubt, A14 führt es als Gegenprobe |
| Stückgröße des Lesers ohne `maxStueck` (g-stueck) | grün | gleichwertig für den einzigen Aufrufer: `net/http` ruft `Read` des Request-Bodys mit höchstens 32768 Bytes (Probe: größtes `len(p)` 32768 bei `Laenge` -1 und bei bekannter Länge), gleich `maxStueck`; für einen anderen Aufrufer wäre sie nicht gleichwertig, der Vertrag sagt nichts zu |
| Klemme der Frist über 30 Tage (v4-klemme-…) | grün | **nicht gleichwertig, die Begründung („darüber trennt kein Fall“) trägt nicht**: V13 |
| `Methode` an jeden Fehler (v9e) | 115 von 119 rot, vier grün | die leere `Methode` kann nichts nennen: gleichwertig; die drei Fälle der Rahmung teilen den Fehlerweg (`rang`, „Verbindung ohne Antwort“) mit den Fällen von A9, die den Marker tragen: im Code gleichwertig; die Begründung im Test („Methoden der Rahmung“) trägt nicht: V15 |

Zusätzlich grün ohne Finding: die Mutation „`ungueltig` ohne Feldname“ (g-feldname-weg): V16.

## 6. Neue Lücken, die die neuen Fänger aufmachen (Auftrag Punkt 2)

Suchraster: Paare gleichzeitiger Ursachen (`rang` und `ergebnis`: alle fünf anderen Reihenfolgen und die drei Auslassungen je Funktion: rot), beide Seiten jeder Grenze (`Frist` und `Laenge` am größten Wert, `%25` und `%80` als gültige Seite, alle 256 Bytes: ohne weiteren Befund), Gleichwerte (Fixture gleich dem Wert der Mutation: nur bei V13), Allaussage über alle Fälle (V14).

**V14, Allaussage A14.** Nr. 8 sagt „ein Byte der Antwort“ und von `Ziel` „höchstens Host und Port, weder Pfad noch Zugangsdaten“; der Kopf von `fehler.go` wiederholt es. Der Marker steht jetzt in Status-Text, Header-Wert, `Upgrade` und Body der Antwort, im Pfad von `Ziel` und im Passwort der Zugangsdaten. Nicht belegt sind zwei Arten: der **Name** eines Antwort-Headers (Mutation: die Namen von `res.Header` an den Fehler beim Kopf, je Zweig 101, `ctx`, Frist: alle drei grün) und der **Benutzername** (Mutation `u.User.Username()` im Fehler des Ziels: grün; das Fixture trägt `http://u:S3CR3T-z@…`). Die Probe schließt beide: Marker im Namen des Fakes (`X-K-S3CR3T`) und im Benutzernamen (`http://S3CR3T-n:S3CR3T-z@…`) lassen den Code grün und machen die vier Mutanten rot, im Zweig 101 an `g Antwort 101`, bei `ctx` an c, d und h, bei der Frist an A8 (d) und i, beim Benutzernamen an `Ziel http://…@…`. Im Kopf des Listeners macht die Bibliothek den Namen kanonisch (`X-K-S3cr3t`): dort bliebe der Marker nicht stehen, der Fake genügt, er läuft durch denselben Zweig von `ergebnis`. Kategorie **LOW**: die Wirkung ist ein Name, kein Wert; ein `Ziel` mit Zugangsdaten kann `UPSTREAM_URL` nach `SPEC-110` Nr. 7 nicht liefern, der Record-Dienst reicht nur dessen Wert durch.

**V15 (INFO).** Siehe Abschnitt 5: die drei Fälle der Rahmung laufen mit dem Marker in `Methode` grün (Probe), und die Mutation „`Methode` an jeden Fehler“ ist dann in ihnen rot (118 von 119 statt 115). Bei `Laenge` über 0 hängt die Rahmung nicht von der Methode ab; nur bei `Laenge` 0 gilt `Content-Length: 0` je Methode (A5). Im Plan A14 nennt „nicht bei `HEAD`“ einen Fall, den die Tabelle der Fehlerfälle nicht hat. Der Weg ist derselbe wie in A9; ohne Folge für die Zusage.

**V16 (INFO).** `pruefe` („der Fehler nennt das Feld“) und `ungueltig` („mit dem Namen des Felds (`SPEC-120` Nr. 2)“) beschreiben, was da ist (§3.7); ein Fänger fehlt (`ungueltig` ohne Feldname: grün), und `SPEC-120` Nr. 2 und 8 sagen nur „die Art“ zu. Entweder prüft `argumentFall` den Feldnamen im Text, oder der Halbsatz entfällt.

## 7. Kommentare (§3.7, §3.11)

* `fehler.go`, Kopf: jede Aufzählung trägt jetzt einen Fänger (Methode: 6 Mutanten; Pfad, Query, Header: A14; Status-Text, Header-Werte, `Upgrade`, Body der Antwort: 14 Mutanten; Text der Bibliothek und von `errBody`: 3 Mutanten rot; `Ziel` über Host und Port: 5 Mutanten) bis auf die zwei Zellen von V14 („ein Byte … der Antwort“, „von `Ziel` mehr als Host und Port“).
* `client.go`: `pruefe` trägt „der Wert von `Ziel`, `Methode`, `Pfad`, `RawQuery` und der Header steht in keinem Fehlertext“ (belegt, V14 für die Namen ausgenommen) und „der Fehler nennt das Feld“ (ohne Fänger, V16). `leseZiel` („die Prozent-Kodierung … im Host und eine Adresse in eckigen Klammern, die keine IPv6-Adresse ist, weist url.Parse selbst ab“) ist belegt: 7 Fälle in `argumentFaelle` je `ErrUngueltig`, Mutationen rot. `istToken` („HTTP-Token (RFC 9110)“, V7 vorher): belegt durch die 256 Bytes. `Konfig` und `New` wie vorher.
* Tests: `argumentFall`, `setzeZiel`, `mitGeheim`, `fakeKopfMit`, `TestFehlerOhneWerte` beschreiben, was sie tun; der Kommentar von `rahmungFehlerFaelle` („Methoden der Rahmung“) trägt nicht (V15). Kein Konjunktiv über eine verworfene Alternative, kein abwesender Text.

## 8. Cherry-Pick, Größe

`git merge-tree --write-tree --merge-base 0716718^ 015da8c 0716718`: rc 0, kein Konflikt, Baum `766a780377498c5abac59d55eb8b48340d7748d1`. Der Diff von `015da8c` zu diesem Baum gleicht in Pfaden, Zeilenzahlen (vier Dateien: `tls.go`, `tls_test.go`, `client.go`, `fehler_test.go`; 291 Einfügungen, 5 Löschungen) und jeder geänderten Zeile dem von `0716718`. Das Ergebnis, aus dem Baum exportiert und schreibgeschützt eingehängt, besteht `go vet` und `go test` über `httpupstream` und `bootstrap` mit `--network none`; `tlsFehlerFaelle` steht in `alleFehlerFaelle`, die TLS-Fälle laufen mit den neuen Hilfen (`mitGeheim` mit `Methode`, `fakeKopfMit`) grün. `merge-tree` schrieb nur Objekte, keine Referenz.

`git diff --numstat f76ba07..015da8c -- src/`: **2768** (2764 Einfügungen, 4 Löschungen), Produktcode 531, Tests 2237; der Commit `015da8c` allein 205 (159 und 46; 17 Produktcode, 188 Tests), ab der Basis 113 (2768 minus 2655). Die Zahlen des Implementers stimmen; die Marke 2900 hält mit 132 Zeilen (4,6 %) Spielraum. Der Zuwachs liegt in der Schätzung von §4 (135 bis 225).

## 9. Plan gegen Code, Spezifikation, Folge-Slices (§3.9, §3.13)

**Spezifikation (`daad9e8`).** `SPEC-120` Nr. 1 und 4 („ohne Obergrenze bis `math.MaxInt64`“, bei `Frist` in ns), Nr. 2 (die Formen des Hosts als Zusage, `Senden` liest `Ziel` mit `url.Parse`), Nr. 8 (`Methode` in der Aufzählung, auch der Fehler der ungültigen Eingabe, von `Ziel` weder Pfad noch Zugangsdaten), `SPEC-110` Nr. 7 („Den Host von `Ziel` liest `Senden` nach derselben Regel“) und *Ungeprüft* (Grund der Meldung mit Grund und Folge), Historie: untereinander und gegen Plan, Code und Tests widerspruchsfrei; kein Slice im Spec-Stratum. Die Entscheidungen sind durch Fänger belegt (V5, V9, V3 und V4 in Abschnitt 1); V8 ist ein akzeptiertes Negativ mit Grund und Folge.

**Plan (`015da8c`).** Kopf, §1, §3 (A1, A2, A5, A8, A14, A19, Dateitabelle, Liste der grünen Mutationen), §4 (Entscheidung des Architect nach der Verifikation, Größe), §6 (Größe, Risiken) sind in `daad9e8` und `015da8c` nachgezogen. Die Zahlen der Zeile A14 stimmen (Abschnitt 2), die Fänger-Namen stehen im Code (`TestTokenZeichen`, `laengeMaxFall` in `rahmungFehlerFaelle`, `antwortVorAblauf`, `zielGeheim`, `fakeKopfMit`), die Dateitabelle gleicht dem Code. Die Annahme zu `synctest` (30 Tage) gilt, solange V13 offen ist.

**Folge-Slices.** `slice-bootstrap-m2`: unverändert, die Sendung steht mit der Kennung `slice-upstream-adapter` in §1, kein Ausschluss trifft sie. `slice-upstream-tls` §1 nennt für den übrigen Adapter „A12 bis A18“, nicht A19: V10 bleibt offen, die Datei ist seit `5947116` unberührt. V8: der Architect hat entschieden (§4 (3): der Grund bleibt, akzeptiertes Negativ in §3 und in `SPEC-110` *Ungeprüft*, beide stehen); den Träger zur Frage in §6 führt der Planner, und `grep -n "I-4"` findet den Plan nur in §4: offen beim Planner.

## 10. Zeitläufe, Raten, ungecachte Stufen, Gates

**Rot ohne Codefehler: keines.**

| Lauf | Ergebnis |
|---|---|
| `go test -count=50` über `httpupstream` | grün, 7,5 s |
| Testbinary, `GOMAXPROCS=1`, `-test.count=100` | grün, 11 s |
| Testbinary, `--cpus=0.05`, `-test.count=30` | grün, 50 s |
| A4 (c), letzter Fall, `slices.Sort` entfernt | 53 von 200 Prozessen rot (Plan 50 von 200; erste Verifikation 28 von 100), unmutiert 0 von 100 |
| A15, eine `http.Header`-Map für alle Aufrufe | 60 von 60 rot, unmutiert 0 von 60 |
| `make test`, `--no-cache-filter test` | `pruefung test: ok` (46 s) |
| `make lint`, `--no-cache-filter lint` | „0 issues“ in allen drei Läufen von `lint.sh`, rc 0 |
| `make gegenproben` | „Eingaben unverändert seit dem letzten grünen Lauf (7448bb93…)“; der Stempel ist der von `015da8c`, `git status` unverändert |
| `make gates` auf `015da8c` | rc 0; d-check 177 Dateien, 0 Befunde; a-check 0 Befunde; `baseline-verify` v6.18.1 OK |

Ohne `-race` (das Image baut mit `CGO_ENABLED=0`) bleiben Datenrennen ungeprüft, wie der Plan sagt. Kein `//nolint` unter `src/`; `.golangci.yml`, `.a-check.yml`, `Dockerfile`, `Makefile` und `harness/mk/` sind seit `7a8fde7` unverändert, kein Gate gelockert.

## 11. Lernpunkte für die Closure

**Finding-Klassen dieses Laufs:** Zusage über das Fehlen einer Grenze, Fänger am Beispielwert statt am größten Wert, den die Umgebung trägt (V13) · Allaussage über alle Fälle, eine *Art* von Byte ohne Marker (V14) · Begründung, die nur beschreibt, was der Test nicht tut (V13, V15) · Kommentar beschreibt eine Eigenschaft ohne Fänger (V16).

**Beobachtungs-Register (`docs/plan/planning/observations/BEO-REPO/`):**

| Beobachtung | Berührt durch |
|---|---|
| `vertrag-sagt-mehr-zu-als-ein-gate-prueft` | V13, V14 (derselbe Vorgang zählt einmal; V1 bis V6 der ersten Verifikation lösen die Klasse am Slice) |
| `universalaussage-im-vertrag-ungeprueft` | V14: die Allaussage hatte nach der Korrektur von V1 und V2 noch Arten ohne Marker |
| `akzeptiertes-negativ-ohne-folge` | V13: der Eintrag „Klemme über 30 Tage“ in §3 nennt als Grund die Fixture, keine Folge |
| `befund-an-rolle-ohne-artefakt` | V8: der Architect hat geantwortet, der Träger beim Planner fehlt noch |
| `berichtigung-erreicht-nicht-jede-stelle` | V10 offen |
| `nebenlaeufiger-lauf-nur-im-glueck-gruen` | kein neuer Beleg: Raten A4 (c) 53 von 200, A15 60 von 60, kein Rot ohne Codefehler |
| `groessenmarke-ueberschritten-ohne-rueckfuehrung` | kein Beleg: 2768 unter der Marke 2900 |

**Lehre für §3.15 und §3.16.** (1) „Ohne Obergrenze“ wird am größten Wert gefahren, den die Testumgebung trägt, nicht an einem Beispielwert; die Umgebung nennt die Grenze (die Blase: Uhr ab 2000-01-01, also W höchstens 262 Jahre) und die Zusatzzeile *Annahme* trägt sie. (2) Bei einer Allaussage über Bytes trägt jede *Art* den Marker, die der Fehlertext nennen könnte: Status-Text, Name und Wert des Headers, Body; Name und Passwort der Zugangsdaten; ein Marker, den die Bibliothek kanonisch schreibt (Header-Namen aus `net/http`), steht dort in kanonischer Form oder der Fall läuft über den Fake. (3) Wer eine gemeinsame Hilfe der Fälle umbaut (`mitGeheim`, `setzeZiel`, `fakeKopfMit`), fährt dieselben Mutanten auf dem Stand davor und danach; der Maßstab ist „keiner wechselt von rot nach grün“. (4) Eine Begründung für eine grüne Mutation zeigt Gleichwertigkeit; „darüber trennt kein Fall“ ist keine.

**Risiken in §6 und ihre Ausgänge bei der Closure** (jedes genau einen; die Belege dieses Laufs): Nebenläufigkeit (Raten, `-count=50`, 100 Läufe mit `GOMAXPROCS=1`, `--cpus=0.05`; Datenrennen ohne `-race` ungeprüft: *weiter offen* oder *entfallen* mit dieser Begründung) · Randformen (die zwei Zellen, die der Architect nach der Verifikation entschied, haben Zusage und Fänger: *entfallen*; V13 und V14 sind Zellen mit Ort) · Annahmen über die Standardbibliothek (an Go 1.27.1 gepinnt, neu belegt: Kanonisierung der Header-Namen, Lesepuffer von 32 KiB, Uhr der Blase; *weiter offen*) · Loopback und Namensauflösung (*entfallen* für die Formen dieses Slice) · Record-Dienst gegen Fake, Adapter gegen `SPEC-120` (erster gemeinsamer Lauf im M2-Test: *weiter offen*) · Größe (die Marke hält: *entfallen*). Vier Risiken tragen schon einen Ausgang (Entscheidungen des Architect, akzeptierte Negative, Host, TLS-Block).

## 12. DoD Punkt für Punkt

| DoD | Stand | Beleg |
|---|---|---|
| Adapter `httpupstream` nach `SPEC-120` Kopf, Nr. 1 ohne `CAPEM`, 2 bis 6, 8, 9; Fänger A1, A2 bis A9, A12 bis A19 | Code erfüllt; Fänger tragen bis auf V13 und V14 (beide LOW) | jede Zusage mit Mutation rot aus dem benannten Grund (Abschnitte 1 bis 3, 435 Mutationen); Reste: Obergrenze der `Frist` nur bis 30 Tage getrennt (V13), Namen der Antwort-Header und Benutzername ohne Marker (V14) |
| Transport nach Nr. 7 ohne TLS (A11 `TestTransport`) | erfüllt | `Proxy` nil, `DisableCompression`; `http.ProxyFromEnvironment` rot (`TestTransport`), Kompression an rot (`TestHeaderGenau`); kein anderes Paket importiert `httpupstream` |
| TLS-Teil mit reinem Revert von `0716718` genommen, Messung in §6, `Konfig` ohne Zusage zu `Pruefen` und `CAPEM` | erfüllt | unter `httpupstream/` kein `crypto/tls`, `httptest`, `TestMain`, `tlsFehlerFaelle`; `git diff 7a8fde7..015da8c` berührt keine TLS-Datei; der Cherry-Pick geht ohne Konflikt (Abschnitt 8) |
| `make gates` grün | erfüllt für `015da8c` | rc 0; `make test` und `make lint` zusätzlich ungecacht (Abschnitt 10) |
| Review durchgeführt, Report liegt vor | erfüllt | `docs/reviews/2026-10-21-slice-upstream-adapter-review.md` (`7cecc39`) |
| Doku-Update (`SPEC-120`: Nachzug nach Review und Verifikation) | erfüllt | `daad9e8`: Nr. 1, 2, 4, 8, `SPEC-110` Nr. 7 und *Ungeprüft*, Historie (Abschnitt 9) |
| Closure-Notiz, Register, Ausgänge, Paarungen | offen, beim Planner | §7 „Wird bei der Closure gefüllt“; sechs Risiken in §6 tragen noch *offen* (Abschnitt 11) |

Ein Häkchen setze ich nicht.

## Stand der Befunde V1 bis V12

| ID | Stand | Beleg |
|---|---|---|
| V1 MEDIUM | gelöst | v1a bis v1e rot (28, 19, 19, 26, 9 Fälle); Zahlen des Plans stimmen; Marker im Pfad aller 29 Ziel-Fälle |
| V2 MEDIUM | gelöst | 16 von 16 Mutanten rot; Rest als V14 (Namen der Header) |
| V3 MEDIUM | gelöst | 6 von 6 rot in `TestRahmung/Laenge_math.MaxInt64_mit_ab` |
| V4 LOW | gelöst bis auf V13 | `New` und Klemmen bis 30 d rot; Klemmen über 30 d grün |
| V5 LOW | gelöst | Zusage mit Fänger in beiden Lesern; 11 Mutanten rot |
| V6 LOW | gelöst | 256 von 256 rot, dazu 3 Aufrufer-spezifische |
| V7 LOW | gelöst bis auf V14 und V16 | Kommentare tragen bis auf drei Zellen (Abschnitt 7) |
| V8 LOW | offen, Planner | Entscheidung steht (§4), Träger in §6 fehlt |
| V9 INFO | gelöst | `SPEC-120` Nr. 8, 6 Mutanten rot, A14 mit Marker in `Methode` |
| V10 INFO | offen, Planner | `slice-upstream-tls` §1 ohne A19 |
| V11 INFO | gelöst | die acht grünen Mutanten stehen in Plan §3, sieben tragen ihre Begründung (Abschnitt 5), eine nicht: V13 |
| V12 INFO | gelöst | 2768 gegen 2900 |

## Befunde dieser Nachverifikation

| ID | Kategorie | Befund | Quelle | Pfad | Zuständig |
|---|---|---|---|---|---|
| V13 | LOW | Die Zusage „ohne Obergrenze bis `math.MaxInt64` ns“ der `Frist` trennt A8 (f) nur bis 30 Tage. Gemutet bleibt der Timer auf 30 d + 1 ns, 31 d, 365 d, 10, 50, 100, 146, 200 Jahre geklemmt grün (9 Mutanten); die Klemme auf 30 d fällt nur über den Gleichstand von Kopf und Ablauf. Der Plan führt „darüber trennt kein Fall“ als bewusst grüne Mutation; die Begründung zeigt keine Gleichwertigkeit. Die Probe mit dem Kopf nach 100 und 250 Jahren ist auf dem Code grün (grün bis 260 Jahre, rot ab 265 durch den Überlauf der Uhr der Blase, die 2000-01-01 beginnt) und auf jeder Klemme unter dem Wert rot; Kosten 0,00 s. | `SPEC-120` Nr. 1; `SPEC-110` Nr. 4; `AGENTS.md` §3.15 | `src/adapters/driven/httpupstream/frist_test.go` · `antwortVorAblauf(t, time.Duration(math.MaxInt64), 30*24*time.Hour)` | Implementer (Wert in A8 (f)), Planner (Zeile A8, Zusatzzeile *Annahme* zu `synctest`, Eintrag „Klemme über 30 Tage“ in §3) |
| V14 | LOW | Zwei Arten von Bytes der Allaussage A14 ohne Marker: der Name eines Antwort-Headers (die Namen von `res.Header` an den Fehler beim Kopf, je Zweig 101, `ctx`, Frist: 3 Mutanten grün) und der Benutzername der Zugangsdaten von `Ziel` (`u.User.Username()` im Fehler: grün; das Fixture trägt `u:S3CR3T-z`). Status-Text, Header-Wert, Body und Passwort sind belegt. Die Probe (Marker im Namen des Fakes, im Benutzernamen) lässt den Code grün und macht die vier Mutanten rot; im Kopf des Listeners schreibt die Bibliothek den Namen kanonisch, der Fake genügt. Der Kopf von `fehler.go` und `SPEC-120` Nr. 8 sagen „ein Byte … der Antwort“ und „weder Pfad noch Zugangsdaten“. | `SPEC-120` Nr. 8; `AGENTS.md` §3.11, §3.15, §3.16 (Allaussage) | `src/adapters/driven/httpupstream/helfer_test.go` · `Header: http.Header{"X-K": {geheim + "-h"}}, Body: b,` | Implementer |
| V15 | INFO | Die Fälle `Laenge 3 mit ab`, `Laenge 3 mit abcd` und `Laenge math.MaxInt64 mit ab` tragen den Marker in `Methode` nicht, weil der Kommentar sie „Methoden der Rahmung“ nennt; bei `Laenge` über 0 hängt die Rahmung nicht von der Methode ab. Mit dem Marker laufen die Fälle grün, und „`Methode` an jeden Fehler“ ist dann in ihnen rot (118 von 119). Im Code gleichwertig (derselbe Fehlerweg wie A9); die Begründung im Test und „nicht bei `HEAD`“ im Plan A14 (kein solcher Fehlerfall) tragen nicht. | `AGENTS.md` §3.11, §3.15 | `src/adapters/driven/httpupstream/anfrage_test.go` · `// Server liest Kopf und drei Bytes und antwortet nie. Die Methoden PUT und POST sind Methoden der Rahmung` | Implementer (Marker oder Kommentar), Planner (Satz in A14) |
| V16 | INFO | „Der Fehler nennt das Feld“ (`pruefe`) und „mit dem Namen des Felds“ (`ungueltig`) haben keinen Fänger: `ungueltig` ohne Feldname bleibt grün. Beschreibung des Zustands (§3.7), keine Zusage der Spezifikation (Nr. 2 und 8 nennen „die Art“); Fänger oder Halbsatz entfällt. | `AGENTS.md` §3.11 | `src/adapters/driven/httpupstream/client.go` · `// (SPEC-120 Nr. 2); der Fehler nennt das Feld; der Wert von` | Implementer |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Produktcode seit `7a8fde7` | nur Kommentare geändert; ohne Kommentare byte-gleich |
| V1 bis V3, V5, V6, V9 | 16 + 5 + 6 + 8 + 3 + 259 + 6 Mutanten, alle rot aus dem benannten Grund; Gegenproben grün |
| Umbauten an `mitGeheim`, `setzeZiel`, `fakeKopfMit`, Listener 101 | Differenzlauf mit 151 Mutanten auf `7a8fde7` und `015da8c`, kein rot→grün |
| Rangfolge (`rang`, `ergebnis`), Rahmung, Frist, Argument-Prüfungen, A19 in beiden Lesern | rot bis auf die Gegenproben; ohne Befund |
| Spezifikation `daad9e8` gegen Plan, Code, Tests | widerspruchsfrei; Entscheidungen belegt |
| Cherry-Pick von `0716718` | kein Konflikt, Diff gleich, Ergebnis baut und besteht |
| Zeitabhängigkeit | kein Rot ohne Codefehler (`-count=50`, `GOMAXPROCS=1`, `--cpus=0.05`); Raten A4 (c), A15 wie im Plan |
| Gegenproben, Gates, Suppression | Stempel aktuell; `make gates`, `make test`, `make lint` ungecacht rc 0; kein `//nolint`, kein Gate gelockert |
| Übergaben (`AGENTS.md` §3.13) | `slice-bootstrap-m2` ohne Befund; `slice-upstream-tls`: V10 |

## Zuständigkeiten

* **Implementer:** V13 (Wert in A8 (f): ein Kopf nach 200 oder 250 Jahren, W höchstens 262 Jahre), V14 (Marker im Namen des Fake-Kopfs und im Benutzernamen), V15 (Marker in `Methode` der drei Rahmungsfälle oder den Kommentar berichtigen), V16 (Fänger oder Halbsatz streichen). Keiner davon blockiert.
* **Planner:** V8 (Träger in §6 mit Ausgang), V10 (A19 in `slice-upstream-tls` §1); die Zeile A8 und die Zusatzzeile *Annahme* zu `synctest` mit V13; der Satz „nicht bei `HEAD`“ in A14; DoD-Häkchen, Closure-Notiz in §7, die Ausgänge der Risiken in §6, die Register-Einträge (Abschnitt 11) und die drei Paarungen.
* **Architect:** nichts.
