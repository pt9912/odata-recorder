# Verifikations-Report: slice-v2-metadaten — 2026-11-01

**Verifikations-Art:** DoD- und Entscheidungs-Konformität, Plan-vs-Code-Diff — geprüft gegen `docs/plan/planning/done/slice-v2-metadaten.md` (DoD §2, Tabelle *Zusage · Fänger* A1 bis J4 mit 59 Zeilen, Raster §6, Block der Größenmarke §4), `SPEC-171` Nr. 1 bis 10 in der geltenden Fassung (mit der Grenze der Bindings in Nr. 6 und 7), `SPEC-170`, `SPEC-125` Nr. 2 (mit `io.ErrNoProgress`) und *Ungeprüft*, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md), `AGENTS.md` §3.9, §3.10, §3.12, §3.15 bis §3.18.

**Gegenstand:** `slice-v2-metadaten` (Welle `welle-odata-v2`), `git diff b5ac4a5..HEAD`, HEAD `d02d7e1488f929b7194d7c1e1a04408f6ffe01a5` (Zweig `slice-v2-metadaten`, Worktree `/Development/odata-recorder-wt-v2-metadaten`). Bezug: `LH-FA-17`, `LH-RB-01`.

**Rolle / Modell / Datum:** Verifier · claude-sonnet-5-5 · 2026-11-01. Frischer Kontext, einmalige Verifikation, kein Push.

**Methode.** Der Bericht des Implementers lag nicht vor; geprüft ist der Beleg, nicht die Behauptung. `make gates` ist selbst gefahren; die Größenmarke ist aus `git diff --numstat` selbst gerechnet; die Mutationen laufen auf Kopien (`git archive HEAD` nach `.../scratchpad/ver/base`, je Mutante `sed` auf der Kopie im Container, Image `odata-recorder:race`, `--network none --memory=4g --memory-swap=4g --cpus=4`, unter `flock /tmp/odata-recorder-heavy.lock`); der Worktree-Code ist unverändert. Die Mutanten sind **neu gebildet an der Klasse**, nicht aus der Tabelle des Plans oder des Reviews übernommen: soweit sie ein Element nennen, das Plan oder Review schon führen, steht es unten als Wiederholung (Prüfung der Fix-Runde); die übrigen sind Elemente, die weder Plan noch Review nannten.

---

## Ergebnis in einem Satz

Kein HIGH, kein MEDIUM. Ein LOW (Kleinschreibung des `edm.`-Anfangs ohne Fänger) und vier INFO. Die Fixes der Fix-Runde (F-1 bis F-7, I-1) halten gegen die Klasse: 108 neue Mutanten, 99 rot, 9 grün (2 äquivalent, 2 im akzeptierten Negativ von N3, 2 als LOW-1, 1 als INFO-1, 2 als INFO-2).

---

## Belege (selbst gemessen)

| Beleg | Messung |
|---|---|
| `make gates` | Exit 0 (Log `.../scratchpad/gates.log`, Schluss `zeichen: ok, 1107 Datei(en) geprüft`); `groessenmarke slice-v2-metadaten.md: ok, Messung 4855 von Marke 8200`; Hinweis `Tabelle ergibt 6230, Schätzung 6070` |
| Größenmarke selbst | `git diff --numstat b5ac4a5..HEAD` über die elf Pfade des Blocks: **4855** (Summe aus Einfügungen und Löschungen, Tests eingeschlossen); Marke 8200 hält mit Abstand (59,2 %); 6230 × 1,25 = 7787,5, aufgerundet 7800, also hielte die Marke auch die nachgerechnete Schätzung |
| Gesamt-Diff | 5319 Zeilen in 22 Dateien; davon Slice-Pfade 4855, Plan 262, Review 181, Spezifikation 21 |
| J4 (V4 byte-gleich) | `git diff --numstat --diff-filter=MDRT b5ac4a5..HEAD` listet **nur** den Plan (243/19) und `spec/spezifikation.md` (12/9); `csdl`, `model`, `ports/outbound` (bestehende Dateien), `jsoncodec`, `urlparser`: leer; alle anderen 20 Dateien sind Hinzufügungen (`--name-status`: nur `A`, sonst die zwei `M`); die Tests von `csdl` laufen grün (`go test ./src/adapters/driven/odata/...`: `csdl` ok, `csdlv2` ok, `odata` ok) |
| Tabelle | 59 Zeilen (`awk` über Zeilenanfänge A1 bis J4 in §3 und in der Zeile *Mutationen*); jede Zeile hat einen Fänger und eine Mutation |
| Zählprüfung | `go test -v -run 'TestVertrag/^Z_'` gibt die Zahlen des Plans: A1 30, B4 305, C1 43, D1 26, D2 50, D3 36, D4 35, D5 62 und 9, D6 142, D9 60, D11 156, G1 140, N1 36, N3 54/1555/1566/6264, N6 29, N8 9, A6 48, alle Dokumente der Suite 1557 |
| Rohbytes (`AGENTS.md` §3.18) | gegengelesen: die sechs Suite-Dateien `metadata_codec_v2_*.go`, die drei Fixtures `testdata/v2_*.xml`, die neun Dateien von `csdlv2/` und `versionstrennung_test.go`: `LC_ALL=C grep -P` nach `\u`+4 Hex-Ziffern: 0 Treffer; nach den Rohbytes `C2 A0`, `EF BB BF`, `E2 80 80..8F`: 0 Treffer; NBSP, `U+0085`, `U+2003` entstehen als Escape-Text (`"\xc2\xa0"`, `"\xC2\x85"`, `"\xE2\x80\x83"` in `metadata_codec_v2_assoziation.go:149` und `metadata_codec_v2_lesen.go:246`); jenseits von ASCII stehen nur Umlaute in Kommentaren und Fallnamen |
| Paketvariablen (I2/I3) | `grep` über `csdlv2/`: einzige Paketvariable ist `errKodierung` (Sentinel, nur gelesen); `navs` und `bindungen` sind Felder des Aufrufs (`dokument`, `dokument.go:149/150`), keine Paketvariable, kein Fehlerzeiger; `navs` wird nur indiziert, nie iteriert (Reihenfolge der Bindings aus `c.reihe` und der Slice je Association) |
| Raten I1 bis I3 (30 Prozesse je Zeile, `go test -c -race`, `-test.count=1`, `--cpus=4`) | Basis: 0/0/0 von 30 (I1/I2/I3); Mutante I3 *Fehlerzeiger im Paket*: I1 0, I2 30, I3 30; Mutante I2 *Typenliste als Paketvariable*: 0/30/30; Mutante I3 *Zuweisung und sofortige Rückgabe*: 0/30/30; **mit Detektor 30 von 30 für I2 und I3 in allen drei Mutanten**, wie im Plan. Ohne Detektor (zuerst gemessen, `go test -c`): Fehlerzeiger 0/30/30, Typenliste 0/30/30 — gleich dem Plan |

Anmerkung zur Messung: ein erster Lauf der `-race`-Zeilen scheiterte im Image an `CGO_ENABLED` (`-race requires cgo`) und maß dabei nur die Binärdatei ohne Detektor; die genannten Detektor-Raten stammen aus dem zweiten Lauf mit `CGO_ENABLED=1` (Ausgabe ohne Bauefehler).

---

## DoD, Zeile für Zeile

| DoD-Punkt | Befund |
|---|---|
| Leser und Modell, V4 unberührt | erfüllt: `csdlv2` liefert `model.ServiceModel` mit `Version()` `2.0` (F1, `modell()` ruft `NeuesServiceModel("2.0", …)`); J4 leer; V4-Tests grün |
| Validierungsbericht | erfüllt: Funde `konstrukt`, `erweiterung`, `property_typ`; G1 bis G6; Mutanten an Konstrukten, Attributen und Reihenfolge rot (unten) |
| Verträge und Randformen des Lesers fremder Bytes | erfüllt: Raster §6 hat in jeder Zelle einen Ort (Nummer von `SPEC-171` oder `SPEC-125`) oder „gilt nicht“ mit Grund; keine Zelle im Code entschieden (Stichprobe: `maxBindings`, Reihenfolge Auflösung vor Grenze, `ErrNoProgress`, `Multiplicity`-Liste, Mehrfach-Container — alle in `SPEC-171` Nr. 6, 7, 10 und `SPEC-125` Nr. 2 vor dem Code) |
| `make gates` grün | erfüllt, selbst gefahren (Exit 0) |
| Review, Doku-Update, Closure-Notiz, Register, Risiken-Ausgänge, drei Paarungen | **noch offen und Sache der Closure**, kein Befund an Code: §7 lautet „Wird bei der Closure gefüllt“; in §6 stehen fünf Risiken mit Ausgang *offen* (unbeschränkte Menge der V2-Dokumente, Fänger je Art des Fundes, V4 unberührt, Fakes, Rest-Randformen) und der Planner-Eintrag zum Test der Versionstrennung; der Review-Report liegt (`e8e17fc`); das Handbuch bleibt unberührt (kein Code im Katalog des Binary), das muss §7 nennen |

---

## Plan-vs-Code-Diff

| Zusage des Plans | Code / Spezifikation | Urteil |
|---|---|---|
| Paket `src/adapters/driven/odata/csdlv2/`, `New() *Leser`, kein Feld | neun Dateien (acht Produktcode, `leser_test.go`); `type Leser struct{}`; J1 | gleich |
| Kein eigenes Modell, kein neuer Port | `loese()` baut `model.ServiceModel` über `model.NeuesServiceModel("2.0", …)`; `csdlv2` importiert kein V4-Paket (J3) | gleich |
| Sechs Suite-Dateien, drei Fixtures, `versionstrennung_test.go` | im Diff vorhanden, Pfade gleich dem Block | gleich |
| Grenze der Bindings 65 536, Auflösung der Set vor der Grenze, Ort die überschreitende Set, Ursache `zu_viele_properties` | `maxBindings = 65536` (`namen.go:25`); `assocSetAufloesen` prüft Association, `EntitySet` beider `End`, Rollen der Navigationen, dann `d.bindungen += len(d.navs[a])` und `> maxBindings` mit `fehler(UrsacheZuVieleMembers, s.ort)`; kein Binding wird vor der Prüfung gebaut (`bindings()` läuft erst in `modell()`) | gleich |
| `io.ErrNoProgress` ab 100 leeren Reads wie ein Fehler von `src` | `tokenFehler`: `errors.Is(err, io.ErrNoProgress)` gibt den Fehler unverändert zurück; `lies()` liefert Nullwerte; die Grenze 100 ist die von `bufio` im Decoder, der Leser zählt nicht selbst | gleich |
| Multiplicity-Liste mit drei Werten, bei jedem `End` der Association geprüft | `assocEnd`: `slices.Contains(multiplizitaeten(), e.mult)`, Ort `<Association>@Multiplicity`, vor dem Rollenvergleich | gleich |
| Reihenfolge der Funde: `property_typ` an der Stelle seiner Property | `member()` legt den Platz (`slot`) vor `attributFunde` an; `funde()` belegt ihn beim Auflösen | gleich |
| Spezifikation vor dem Code, Plan folgt (`AGENTS.md` §3.9, §3.13) | Commits `dfff71f`, `6db282c` vor `a89be28`; Fix-Runde: Spezifikation `7ae4ac6`, Plan `24ac1f7`, Code `f84e595`, Messung `d02d7e1`; §6 trägt die Rückgaben mit Ausgang *entschieden* | gleich |
| V4-Zweig zu `io.ErrNoProgress` ohne Fänger (akzeptiertes Negativ) | `csdl/lesen.go:69` trägt den Zweig seit der Basis; unverändert (J4) | gleich, Negativ bekannt |

Eine Abweichung zwischen Plan und Code gibt es nicht. Eine Entscheidung im Code, die der Plan nicht entschieden hat, habe ich nicht gefunden.

---

## Prüfung der Fixes gegen die Klasse (`AGENTS.md` §3.17)

Gebildet aus dem Wortlaut der Zusage, nicht aus dem Befund. „Rot“ heißt: `TestVertrag` ist aus dem benannten Grund rot, die Zeile steht dabei.

| Befund | Menge (Wortlaut) | Elemente | Meine Mutanten | Ergebnis |
|---|---|---|---|---|
| F-1 / N8 | Zählung der Bindings über alle Associations bis 65 536 | Zählweisen | Zählung nur bei verschiedenen `EntitySet` beider `End` (`n8-selfref`); Addition erst bei Rückkehr, Prüfung mit der Summe davor (`n8-lag`); Ort als Association-Name statt Set-Name (`n8-ort`); Abzug von `n/65536` (`n8-nur-gerade`) | `selfref` rot (Abbruch ohne benannte FAIL-Zeile: bekannte Wächter-Panik-Klasse, nicht untersucht), `lag` rot (`N8/2000_AssociationSet_mal_2000_Navigationen`), `ort` rot (`N8/65537:_eine_zweite_Association_zuletzt`), `nur-gerade` **grün** (INFO-1) |
| F-2 / G1, G2 | Attribute mit Namensraum an gemeldeten Konstrukten (7 benannte plus `edmx`), `xml:`-Attribute, Attribute ohne Namensraum | 7 + 1 Konstrukte; 2 `xml:`-Attribute; 8 Attribute ohne Namensraum | Attributfunde je Konstrukt abgeschaltet: `ComplexType`, `FunctionImport`, `ReferentialConstraint`, `Documentation`, `ValueTerm`, `Using`, `OnDelete`, alle Konstrukte im Namensraum `edmx` (8); `xml:lang` und `xml:space` je als Fund (2); je eines der 8 Attribute ohne Namensraum (`MaxLength`, `Precision`, `Scale`, `Unicode`, `FixedLength`, `ConcurrencyMode`, `Collation`, `StoreGeneratedPattern`) als Fund (8) | **18 von 18 rot** (`G1_Funde`, `G2_ohne_Wirkung`) |
| F-3 / G4 | `property_typ` an der Stelle seiner Property, vor den Attributfunden | Stellung relativ zu Attributen und Folge-Funden | `property_typ` bis zur nächsten Property aufgeschoben (`f3-defer`, neue `funde()`); Platzvergabe hinter `attributFunde` verschoben (`f3-member`) | 2 von 2 rot (`G4_Reihenfolge_der_Funde`) |
| F-4 / D5, D6 | Namensraum oder Alias mit `Edm`-Anfang, aber nicht `Edm` oder `Edm.…`; Typ in der Form `Edm.<unbekannt>` | 9 Fälle; 21 V4-Namen | `Alias`-Prüfung auf Präfix `Edm`; Namensraum-Prüfung auf Präfix `Edm`; `EqualFold` für Namensraum und Alias; Typprüfung `Edm` mit Kleinschreibung (`k-*`, 7); und alle 21 V4-Namen (`Date`, `Duration`, `TimeOfDay`, `Stream`, `Untyped`, je 8 `Geography…`, `Geometry…`) einzeln als bekannt (21) | 4 von 7 rot, 3 grün: **LOW-1** (zwei Kleinschreibungen) und ein äquivalenter (`s == "Edm"`: ohne Punkt löst `qualify` nie auf); **21 von 21 rot** |
| F-5 / N3 | `Multiplicity` jeder Zeichenkette außer den drei Werten | alle Zeichenketten bis Länge 4 über 6 Zeichen, 12 Randformen, 6264 Dokumente | Wert mit LF, CR, `U+2003`, `U+0085`, `U+3000` am Rand getrimmt (`n3-ws`); nur LF getrimmt (`n3-nl`) | beide grün: im akzeptierten Negativ des Plans (Zeichen außerhalb des Alphabets, Tab und NBSP geführt); ein `strings.TrimSpace` fängt das Leerzeichen der Schleife (nicht gemessen, folgt aus dem Alphabet) |
| I-1 / A6 | leere Reads 0, 1, 99, 100 davor, zwischen, danach | 48 Fälle | nicht neu gemutet: die Grenze 100 hält `bufio`, der Code zählt nicht; der Fänger prüft 99 gleich Ergebnis ohne leere Reads und 100 gleich `ErrNoProgress` an jeder der drei Stellen | Zählprüfung 48 bestätigt; keine eigene Mutante an der Bibliothek möglich |
| Klasse darüber hinaus | D4 (Schlüsseltypen), D6 (Pflichtangaben), D9 (Rangfolge), B2 (Zeichenbereich) | D4 8 + 2 + Schrägstrich; D6 11 `pflicht`-Stellen, 6 Namensstellen, `PropertyRef`; D9 7 Nachbartausche; B2 Bereichsgrenzen | D4: je einer der 8 nicht unterstützten Typen als Schlüssel erlaubt, `ComplexType` und Entity-Typ als Schlüssel erlaubt, Schrägstrich-Prüfung entfernt (11); D6: jede der 11 `pflicht`-Stellen und 6 `nameAttr`-Stellen und `PropertyRef@Name` lenient (18); D9: 7 Nachbartausche der Schritte (7); B2: 0x00, 0x08, 0x0B, 0x0C, 0x0E, 0x1F, DEL, C1-Bereich, `U+FFFE`, `U+FFFF`, Ablehnung von `U+FFFD`, `U+10FFFF`, `U+E000`, `U+D7FF`, `U+10000`, `U+FEFF`, `U+FFFC` (17); `strings.Index` statt `LastIndex` in `qualify` (1) | D4 11 von 11 rot; D6 17 von 18 rot (die grüne, `Multiplicity` fehlend, ist äquivalent: ohne `pflicht` fällt der leere Wert in die Liste und ergibt `struktur` mit gleichem Ort `<Association>@Multiplicity`); D9 7 von 7 rot (bei zwei Tauschen, `d9-1` und `d9-2`, ohne benannte FAIL-Zeile: rot durch Abbruch, wie bei `n8-selfref`); B2 15 von 17 rot, 2 grün (INFO-2); `d2-index` rot (`D5_Formen_des_Property-Typs`) |

Summe der neuen Mutanten: 108 (b1 25, d5 21, g 10, b2 17, d6 18, k 17: 2 Reihenfolge der Funde, 8 Attributfunde an Konstrukten, 7 Edm-Anfang); 99 rot, 9 grün. Die 9 grünen: `n8-nur-gerade` (INFO-1), `n3-ws`, `n3-nl` (akzeptiertes Negativ), `k-ns-lower`, `k-prop-lower` (LOW-1), `k-prop-edm-exact` (äquivalent), `d6-end-Mult` (äquivalent), `b2-rej-0xFEFF`, `b2-rej-0xFFFC` (INFO-2). Die Wiederholung der Plan-Mutanten (N8, F-2, F-3, A6, F-4) bestätigt die Zeilen der Tabelle *Mutationen*: jede ist rot; die Mutante `f3-member` ist die Mutation F-3 des Plans, `n8-lag` und `n8-ort` sind ihre Verwandten.

Probe des Reviews (17 496 Zerlegungen) habe ich nicht wiederholt; die Mutanten an `strom.go` (B2) und die Zählprüfung von A4 (1536 Dokumente, 2 329 540 Zweiteilungen laut Plan) bestätigen dieselbe Klasse an anderen Elementen.

---

## Raster §6 (jede Zelle mit Ort, nichts im Code entschieden)

Alle Zellen des Rasters (fünf Operationen × acht Argumentarten, 40 Zellen) sind gelesen. Jede Zelle nennt eine Nummer von `SPEC-171` (oder `SPEC-125`, dort „Nr.“ mit Spec-Zahl) oder „gilt nicht“ mit Grund; die Grenze der Bindings steht in der Zelle *Auflösen der Assoziationen × über Grenze* und im Wortlaut von `SPEC-171` Nr. 6 und 7. Die Zusatztabelle (sieben Paare) nennt je Zeile einen Fänger; die Paare *Verweis ohne Ziel und nicht unterstützter Typ* (D9) und *Navigation ohne Ziel und `AssociationSet` ohne Ziel* (D9, N6) sind in den Mutanten `d9-*` rot. Eine im Code entschiedene Zelle habe ich nicht gefunden. Der Plan nennt in den Zellen „Abschnitt 5, Änderung 1 bis 4“ und „Abschnitt 5 Nr. 10 x“; der Kopf des Rasters erklärt beides (`SPEC-171` Nr. 2 c, 5, 7, 10), so dass keine Zelle ohne Auflösung bleibt.

---

## Findings

| # | Schwere | Befund | Regel / Quelle | Stelle | Beleg |
|---|---|---|---|---|---|
| LOW-1 | LOW | Die Kleinschreibung des reservierten Anfangs ist ohne Fänger: `Namespace="edm.Sub"` (oder `EDM.…`) als abgelehnt (`strings.HasPrefix(strings.ToLower(ns), "edm.")` in `schema()`) und ein Typname `edm.Foo` als bekannter `Edm.`-Name (`strings.HasPrefix(strings.ToLower(s), "edm.")` in `property()`) bleiben in der ganzen Suite grün. D6 führt `edm` (ohne Punkt), D5 führt `Edm2`, `EdmX`, `Edmonton`, `Edm_`, `EdmX.Inner` und vier Aliase, aber kein Dokument mit `edm.`-Anfang in Kleinschreibung. `SPEC-171` Nr. 5 sagt „Schreibweise und Leerraum zählen“ und Nr. 10 a „`Edm` und `Edm.`-Anfang reserviert“; die Klasse „Namensraum oder Alias mit Anfang `Edm` in anderer Schreibweise“ hat damit ein Element ohne Fänger. Folge: ein V2-Dienst mit einem Namensraum `edm.…` würde nach einer solchen Änderung unimportierbar (`ODR-E6004`); der heutige Code liest ihn richtig | `AGENTS.md` §3.17 (Klasse), §3.15 (Fänger führt nur eine Form der Eingabe); `SPEC-171` Nr. 5, 10 a | `src/adapters/driven/odata/csdlv2/elemente.go` · `if ns == "Edm" \|\| strings.HasPrefix(ns, "Edm.") {`; `aufloesen.go` · `if strings.HasPrefix(s, "Edm.") {` | Mutanten `k-ns-lower`, `k-prop-lower` im Scratchpad, beide grün |
| INFO-1 | INFO | Die Grenze der Bindings hat keinen Fall, in dem **eine** `AssociationSet` allein 65 536 Navigationen zählt (N8: 256 × 256, 2000 × 2000, drei Associations): die Mutante `bindungen += n - n/65536` (eine Zählung, die bei genau 65 536 Navigationen einer Association um eins zu wenig zählt) bleibt grün. Erreichbar ist die Form nur über einen abstrakten Typ ohne Schlüssel mit 65 536 Navigationen (die Summe der Members lässt keinen Schlüssel dazu zu) und eine zweite Set danach; der Code hat keinen Zweig nach der Zahl. Kein Vertragsbruch, nur eine Seite der Grenze, die kein Fall führt | `AGENTS.md` §3.15 (beide Seiten jeder Grenze) | `Suite metadata_codec_v2_assoziation.go` · Fälle von `zeileV2N8` | Mutante `n8-nur-gerade` grün |
| INFO-2 | INFO | B2 prüft die Zeichen an beiden Seiten der Bereichsgrenzen; innere Zeichen (`U+FEFF` mitten im Dokument, `U+FFFC`) hat kein Fall: die Mutanten `r == 0xFEFF` und `r == 0xFFFC` als ungültig bleiben grün (`U+10000` und die Bereichsränder sind rot). Die Zeile B2 des Plans nennt keine Prüfgrenze mit Grund und Folge für das Innere (rund 1,1 Millionen Zeichen); die Begründung „ein Bereichsprädikat, Grenzen genügen“ steht nicht im Plan. Der Planner kann sie bei der Closure als akzeptiertes Negativ eintragen | `AGENTS.md` §3.17 (unbeschränkte Menge, Prüfgrenze mit Grund und Folge) | `src/adapters/driven/odata/csdlv2/strom.go` · `xmlZeichenGueltig`; Plan §3 Zeile B2 | Mutanten `b2-rej-0xFEFF`, `b2-rej-0xFFFC` grün |
| INFO-3 | INFO | Das akzeptierte Negativ von N3 trägt (Länge 5 und mehr, Zeichen außerhalb des Alphabets) und hat Grund und Folge; zusätzlich hat es keine Mutante gefangen, die LF, CR oder `U+2003` am Rand trimmt (`n3-ws`, `n3-nl` grün). Das ist ausdrücklich das Negativ des Plans („andere Zeichen als das Alphabet“); ein `TrimSpace`-Fehler würde am Leerzeichen der Schleife rot. Hier nur zur Kenntnis, keine Aktion | `AGENTS.md` §3.15 | Plan §3 Zeile N3 | Mutanten `n3-ws`, `n3-nl` |
| INFO-4 | INFO | Closure-Nachträge, die der Verifier nicht erledigt: (a) das Gate nennt `Tabelle ergibt 6230, Schätzung 6070` (die Schätzung des Blocks ist nicht auf die 59 Zeilen nachgerechnet; die Messung 4855 und die Marke 8200 sind davon nicht berührt; die Neuberechnung ist Sache des Planner, wie der Plan selbst sagt); (b) §7 leer, fünf Risiken in §6 mit Ausgang *offen*, Register und Steering-Loop-Eintrag fehlen; (c) §7 muss nennen, dass das Binary unverändert blieb (DoD Handbuch-Punkt) | `harness/README.md` Closure; `modul-05` | Plan §6, §7 | Gate-Log Zeile `groessenmarke slice-v2-metadaten.md: Hinweis: Tabelle ergibt 6230, Schätzung 6070` |

Bekannt und akzeptiert, nicht doppelt gemeldet: N3 ab Länge 5, V4-Zweig zu `io.ErrNoProgress` ohne Fänger, kein Zeit-Fänger für `navs`, Wächter-Panik statt benanntem FAIL bei zwei N8-Mutationen (die Mutanten `n8-selfref`, `d9-1`, `d9-2` zeigen dasselbe Bild: rot ohne benannte FAIL-Zeile).

---

## Negativbefunde (je Bereich)

| Bereich | Befund |
|---|---|
| Plan-vs-Code (Produktcode `csdlv2/`, neun Dateien, 1526 Zeilen) | geprüft; kein Unterschied zwischen Zusage und Code; keine im Code entschiedene Randform; keine Paketvariable außer dem gelesenen Sentinel `errKodierung`; `navs` und `bindungen` je Aufruf |
| Tabelle *Zusage · Fänger* (59 Zeilen) gegen den Vertrag | geprüft; jede Zusage von `SPEC-171` Nr. 1 bis 10 hat eine Zeile (Nr. 1 A/B/E/I, Nr. 2 C, Nr. 3 C4, Nr. 4 G, Nr. 5 D4/D5/F5, Nr. 6 N1 bis N8, Nr. 7 D1 bis D9, Nr. 8 F, Nr. 9 G3/D10/J2, Nr. 10 D6/D7/N3/N6/N7); 108 eigene Mutanten, 99 rot; die 9 grünen sind LOW-1, INFO-1 bis INFO-3 und 2 äquivalent |
| Raster §6 | geprüft; 40 Zellen mit Ort oder „gilt nicht“ mit Grund; Zusatztabelle mit Fängern; nichts im Code entschieden |
| Fixes der Fix-Runde (F-1 bis F-7, I-1) | geprüft gegen die Klasse; F-1, F-2, F-3, F-4, F-5, I-1 halten; F-6 und F-7 (Plan-Stellen) im Plan nachgezogen (D5 mit 21 Namen, §8 nennt I2 und I3, Rückgabe 3 mit Ausgang *entschieden*) |
| J4 / Versionstrennung J2, J3 | geprüft; `--diff-filter=MDRT` leer außer Plan und Spezifikation; `versionstrennung_test.go` prüft beide Richtungen und die Unterpakete, Gegenprobe auf Wegwerfbaum mit sieben Fällen; Tests grün |
| `make gates` | selbst gefahren, Exit 0 |
| Größenmarke | 4855 von 8200 selbst gerechnet; Hinweis 6230/6070 gehört der Closure |
| Rohbytes §3.18 | geprüft; Dateien oben genannt; kein Treffer |
| Raten I1 bis I3 | geprüft; 30 Prozesse, `-race -c`, `-test.count=1`; Basis 0, Mutanten 30/30 mit Detektor; keine neue Paketvariable seit der letzten Messung der Implementer-Runde |
| Kommentare, Suppressionen (§3.2, §3.7, §3.11) | Stichprobe der Köpfe von `zeileV2G4`, `zeileV2N3`, `zeileV2N8`, `zeileV2A6`: die Zusage steht in beiden Richtungen mit dem Fänger (F-3 behoben); kein `nolint`, kein `d-check:ignore` im Diff |

---

## Summary

Kein HIGH, kein MEDIUM. Ein LOW (Kleinschreibung `edm.` ohne Fänger), drei INFO zu Fängern und Closure-Nachträgen, ein INFO zum Negativ von N3. `make gates` grün, Marke 4855 von 8200, V4-Pfad byte-gleich, Raten 30 von 30. Der Slice kann an die Closure; LOW-1 und INFO-1 bis INFO-3 sind dem Planner zur Entscheidung (Fänger oder akzeptiertes Negativ mit Grund und Folge) gegeben, kein Code-Fix des Verifiers.
