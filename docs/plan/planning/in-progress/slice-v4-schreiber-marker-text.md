# Slice slice-v4-schreiber-marker-text: Der Text `[REDACTED]` eines `Edm.String` im V4-Schreiber

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** ohne Welle — keine Closure-Bedingung außer der DoD dieses Slice.

**Bezug:** `LH-FA-07`, `LH-QA-05`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-169` Nr. 1 und 4 (Zuordnung der Arten, Rundlauf Nr. 8) · `SPEC-168` Nr. 4 (Wert; `redigiert` als Art ist verboten, der Text `[REDACTED]` eines `Edm.String` nicht) · `SPEC-126` Nr. 10 f (der Leser liest den Rohtext `[REDACTED]` als Marker) · `SPEC-172` Nr. 6 n (die Entscheidung der Version 2 als Vorbild).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-v4-schreiber-marker-text)

**Autor:** pt9912 (Planner). **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Rundlauf von `SPEC-169` Nr. 8 gilt für jeden Text eines `Edm.String`, auch für `[REDACTED]`: der Schreiber `jsoncodec` (Version 4) schreibt den Text so, dass `ODataCodec` ihn nicht als Marker liest. **Entschieden** (Architect, 2026-11-07, `SPEC-169` Nr. 4 und `SPEC-168` Nr. 4): Escape für `[` nach dem Vorbild von `SPEC-172` Nr. 6 n, nicht `ErrUngueltig`; die Klasse der Texte, die den Escape brauchen, hat ein Element (genau `[REDACTED]`), jeder ähnliche Text bleibt in der Form von `SPEC-169` Nr. 1.

**Übernimmt:** `slice-v2-payload-schreiben` (Fund, Rückgabe 1 des Architect an den Planner, 2026-11-06: der V4-Schreiber schreibt `[REDACTED]` roh, `ODataCodec` liest den Rohtext als Marker; V2 löst es mit `SPEC-172` Nr. 6 n).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der V2-Schreiber `jsoncodecv2` — `slice-v2-payload-schreiben` löst den Fall dort selbst (Entscheidung 3, Zeile 31 der Tabelle); dieser Slice ändert keine Datei dort.
- Der Leser `ODataCodec` und die Markerform von `SPEC-126` Nr. 10 f — Vertrag in `done/`; die Lücke liegt im Schreiber, und eine Änderung des Lesers wäre eine Änderung des Formats, das der Import aufgezeichnet hat.
- Weitere Texte, die der Leser besonders liest, soweit der Architect keine findet — Suchauftrag des Architect (§3), kein eigener Punkt; findet er welche, gehören sie in diesen Slice, solange sie den Schreiber `jsoncodec` betreffen. **Ergebnis der Suche:** der Wortlaut von `SPEC-126` Nr. 10 nennt allein den Rohtext `[REDACTED]`; Nr. 11 lehnt einen Schlüssel mit U+0000 ab, das ist ein Negativ in `SPEC-169` *Ungeprüft*, kein Gegenstand des Schreibens (Z12).
- Die Verdrahtung des Schreibers im Composition Root — `slice-bootstrap-m3`; der Slice ändert den Wert, den `NewSchreiber()` liefert, nicht die Stelle, die ihn einhängt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [x] Vertrag: `SPEC-169` (und, falls nötig, `SPEC-168` Nr. 4) nennt die Form für den Text `[REDACTED]` und die Randformen daneben (`[REDACTED`, `[REDACTED]x`, `[redacted]`, Text mit Leerraum davor und danach, Text in `Collection`); der Architect hat sie vor dem ersten Code-Commit entschieden (2026-11-07: `SPEC-169` Nr. 4, Nr. 8, *Ungeprüft*, `SPEC-168` Nr. 4; die Randformen sind die Elemente der Menge M1 in §3, `Collection` ist im Raster in §6 „gilt nicht“ mit Ort).
- [x] Schreiber: `jsoncodec` setzt den Vertrag um; Rundlauf mit `ODataCodec` über die Klasse der Texte (Schleife über alle Elemente, `AGENTS.md` §3.17), Fänger rot ohne die Änderung.
- [x] Der Vertragstext von `SPEC-169` Nr. 8 und die *Ungeprüft*-Absätze nennen keine Ausnahme mehr für diesen Text (Beleg, Planner 2026-11-07: ein Lauf über den Abschnitt `SPEC-169` in `spec/spezifikation.md` findet das Wort Ausnahme nicht; auf `main` nannten sie nie eine; Nr. 8 sagt jetzt ausdrücklich, dass der Rundlauf auch den Text `[REDACTED]` meint; der Implementer nimmt das nicht zurück).
- [x] `make gates` grün (Verifier, `exit=0`; im Zweig nach dem Merge von `main` vom Planner wiederholt, siehe §7).
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `src/adapters/driven/odata/jsoncodec/schreiber_ausgabe.go` | update | Konstante `markerText` und neue Methode `ausgabe.stringWert(s string)`: ist `s` gleich dem Marker, schreibt sie `"`, das Zeichen `[` als Escape (Backslash, `u`, `0`, `0`, `5`, `b`, aus Bytewerten gebaut), `REDACTED]` und `"`; sonst ruft sie `a.text(s)` |
| `src/adapters/driven/odata/jsoncodec/schreiber_bau.go` | update | `ausgabe.wert`: `case model.ArtString` ruft `a.stringWert(w.Text)` statt `a.text(w.Text)`; keine andere Stelle ändert sich |
| `src/adapters/driven/odata/jsoncodec/schreiber_marker_test.go` | neu | Mengen M1 bis M3, M8 (Z01 bis Z04, Z11) |
| `src/adapters/driven/odata/jsoncodec/schreiber_marker_zweige_test.go` | neu | Mengen M4 bis M7 und die Zählprüfung der Zweig-Menge W (Z05 bis Z10) |
| `spec/spezifikation.md` | update (Architect, erledigt) | `SPEC-169` Nr. 4, 8 und *Ungeprüft*, `SPEC-168` Nr. 4, Historie |

**Entscheidungen des Architect** (2026-11-07, vor dem ersten Code-Commit; Ergebnis in der Spezifikation):

1. **Form: Escape, nicht `ErrUngueltig`** (`SPEC-169` Nr. 4, Absatz *Der Text `[REDACTED]`*). Der Leser setzt `marker` nur, wenn der Rohtext genau `[REDACTED]` ist und kein Escape im String stand (`wert.go`, `scan.go`: `!s.escape && text == "[REDACTED]"`); ein einziger Escape macht den String zu einem gewöhnlichen, und `[` ist das Zeichen, das auch V2 escaped. *Alternative verworfen:* `ErrUngueltig` — der Text kommt aus dem Zustand des Workspace, den ein Client setzen kann (`SPEC-180`: dort ist `[REDACTED]` ein gewöhnlicher Text), `ErrUngueltig` ist ein Programmfehler des Aufrufers (`ODR-E9001`), und ein gültiger Wert würde zu einer Fehlerantwort. *Folge für Bytes:* nur dieser eine Text hat andere Bytes in der Antwort des Mock; der Wert ist gleich. Aufgezeichnete Bytes sind nicht betroffen: der Schreiber schreibt allein Antworten des Mock und hat keinen Zustand. Das Risiko „Escape ändert Bytes“ ist damit entschieden (§6).
2. **Die Klasse der Texte** aus dem Wortlaut von `SPEC-126` Nr. 10 f, nicht aus dem Fund: der Leser liest nur einen String als Marker, dessen Rohtext genau `[REDACTED]` ist, und nur für Typen, die als String stehen dürfen. Der Schreiber schreibt für `Edm.String` den Rohtext gleich dem Text, solange kein `"`, `\` und kein Zeichen unter U+0020 vorkommt; ein Text mit einem davon ist länger als zehn Zeichen. Die Klasse hat **ein Element**. Die Nachbarn (Menge M1, N) sind die Texte, die dem Marker ähneln und **nicht** verändert werden dürfen; jede ihrer Achsen hat einen Fänger, der rot wird, wenn der Vergleich zu weit greift. Die Arten `guid`, `date`, `datetimeoffset`, `int64`, `decimal` kann der Schreiber nicht mit dem Text `[REDACTED]` schreiben (nicht kanonisch, `ErrUngueltig`, Menge M4); `int32`, `boolean` stehen dort nie als String.
3. **Tabelle *Zusage · Fänger*** und **Raster** stehen unten; die Größenmarke steht in §4 (Planner, 2026-11-07).
4. **Pflicht-Mengen** M1 bis M8 unten, mit Achsen, Zahl der Elemente als Produkt, Orakel und Prüfgrenze (Z11).
5. **Zweige und Rückgaben** als Modell-Menge W (18 Verzweigungsstellen) mit Zählprüfung (Z10). Die Wahl der Form (Escape) führt **keine neue Fehlerrückgabe** ein; die Rückgaben, die den Text berühren, sind die Zeilen W12 bis W18.
6. **Kennungen und Dateien:** keine neue `SPEC`-, `ARC`- oder `ADR`-Nummer nötig; der Slice hat in `spezifikation.md` nur `SPEC-169` Nr. 4, 8, *Ungeprüft*, `SPEC-168` Nr. 4 und eine Zeile der Historie geschrieben, nichts in den Bereichen `SPEC-192` bis `SPEC-199`. Eine ADR ist nicht nötig: die Entscheidung ist ein Wortlaut des Formats (`AGENTS.md` §3.8).
7. **Schnittfrage:** bestätigt, kein Schnitt. Produktcode ist eine Methode von rund 15 Zeilen und eine geänderte Zeile; der Rest sind Tests. Richtwert nach `SPEC-128` Nr. 1 (§4) liegt weit unter der Schnittgrenze.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`, 8×).
- Fixtures mit Escape-Folgen oder Rohbytes aus Bytewerten bauen und die geschriebene Datei auf Rohbytes gegenlesen (`AGENTS.md` §3.18, `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um`, 5×). Hier gilt es doppelt: der Escape des Schreibers (Backslash, `u`, `005b`) und die Texte der Menge M1 mit eigenem Escape-Text entstehen mit `esc("005b")` aus `schreiber_helfer_test.go` oder aus `[]byte{…}`; Zeichen jenseits von ASCII (U+0415, U+FF3B, U+1F600, U+00A0, U+FEFF) als `string([]rune{…})`.
- Schwere Läufe unter `flock /tmp/odata-recorder-heavy.lock`, Docker mit `--memory=4g --memory-swap=4g --cpus=4` (Übergangsregeln Nr. 1 und 8).

**Mengen** der Fänger (`implement-slice` Schritt 14, `AGENTS.md` §3.17): Achsen, Zahl der Elemente als Produkt, Generator über die Achsen, Orakel (nie der Schreiber; der Rundlauf mit `ODataCodec` nur zusätzlich). Der Bericht des Implementers nennt je Menge die Zahl, die der Test zählt, und je Befund eine Mutation an einem Element, das der Befund nicht nannte.

| Nr. | Menge | Achsen und Elemente | Orakel | Fänger |
|---|---|---|---|---|
| M1 | Texte der Art `string` = E ∪ N, **143** | E = 1 (`[REDACTED]`). N = **142**: Abweichung vom Marker 90 (leer 1 · Kürzung auf ein Präfix der Länge 1 bis 9 und auf ein Suffix der Länge 1 bis 9 = 18 · Ersetzung des Zeichens an Stelle 0 bis 9 durch `x`, U+0415, U+FF3B, U+1F600 = 10 × 4 = 40 · Streichung an Stelle 0 bis 9 = 10 · Einfügung von `x` vor Stelle 0 bis 10 = 11 · ein Buchstabe an Stelle 1 bis 8 in der anderen Schreibweise 8 und alles klein 1 = 9 · der Marker doppelt 1), Rest 36 (Seite davor, dahinter, beide = 3 × Zeichen 12: `x`, Leerzeichen, Tabulator, LF, CR, U+00A0, U+0000, U+FEFF, `[`, `]`, `"`, `\`), Rest-Länge 4 (Seite davor, dahinter × 4096 und 65 537 mal `x`), Escape-Folge im Text 12 (an Stelle 0 bis 9 das Zeichen als Text Backslash, `u`, vier Hex-Ziffern in Kleinbuchstaben = 10 · `[` als Großbuchstaben-Folge 1 · Backslash vor `[` 1). 90 + 36 + 4 + 12 = 142 | Bytes: fester Wert aus Bytewerten für E; Referenzkodierer im Test (Schleife über die Bytes nach `SPEC-169` Nr. 1) für N; Wert: `encoding/json` dekodiert den Wert zum Text | Z01, Z02, Z03 |
| M2 | Stelle × Modus, **28** | Stelle 7: Entity Nichtschlüssel (`NS` von `Ns.Alle`) · Entity Schlüssel (`S`, `Schluessel` die Literalform des Textes, `'` verdoppelt) · Entity beide Properties · Entity abgeleiteten Typs (`Extra` von `Ns.Sonder` in `Produkte`) · `Entities[0]`, `Entities[1]`, `Entities[2]` (drei Elemente `Ns.Alle`, der Text in `NS` des einen, die anderen `k`). Modus 4: Stufe `minimal`, `none` × Zahlform `Zahl`, `Zeichenkette` (`Accept`) | Wie M1; der Text mit U+0000 an Stelle Schlüssel ist `ErrUngueltig` (`SPEC-168` Nr. 3 b), 3 Texte × 2 Stellen × 4 Modi = 24 Elemente | Z01 bis Z03 |
| M3 | Rundlauf, **2002** = M1 143 × Stelle 7 × Zahlform 2 (Stufe `minimal`) | Elemente von M1 × M2 ohne Stufe `none` (der Leser lehnt sie ab, `SPEC-169` Nr. 8); 12 Elemente mit U+0000 an Stelle Schlüssel sind `ErrUngueltig` | erwartete Entity aus dem Text (`Art` `string`, `Text` gleich, `Schluessel` die Literalform, `Uebersprungen` leer); `encoding/json` dekodiert den Wert zum Text | Z04 |
| M4 | Andere Arten und Typen mit dem Text `[REDACTED]`, **16** | Art außer `string` an einer Property des passenden Typs 7 (`boolean`, `int32`, `int64`, `decimal`, `guid`, `date`, `datetimeoffset`) · Art `redigiert` mit dem Text 1 · Art `null` mit dem Text an nullbarer Property 1 · Art `string` an einer Property jedes der 7 anderen EDM-Typen 7 | `errors.Is(err, model.ErrUngueltig)`, `Darstellung` der Nullwert, `dst` ohne Aufruf von `Write` | Z05 |
| M5 | Ungültiges UTF-8 neben dem Marker, **16** | Seite 2 (davor, dahinter) × Bytefolge 4 (`ff`; `c0 80`; `ed a0 80`; `e2 82`, abgeschnitten) × Operation 2 (Entity, Collection) | wie M4 | Z06 |
| M6 | Abbruch und Fehler von `dst` mit dem Text `[REDACTED]`, **10** | Operation 2 × Fall 5: `Write` meldet einen Fehler · meldet `len−1` · meldet `len−5` (die Länge ohne den Escape) · `ctx` beendet · ungültiges Argument bei beendetem `ctx` | Fehler unverändert (`==`) bzw. `io.ErrShortWrite` bzw. `ErrUngueltig`; `Darstellung` der Nullwert; höchstens ein `Write`, bei `ctx` beendet und `ErrUngueltig` keiner | Z07, Z08 |
| M7 | Eingaben und Wiederholung, **286** | M1 143 × Operation 2 (Entity, Collection) | tiefe Kopie der Argumente vor dem Aufruf gleich den Argumenten danach (`reflect.DeepEqual`); zwei Aufrufe ergeben gleiche Bytes | Z09 |
| M8 | Fehlerobjekt und Anzahl, **`len(model.AlleCodes())` × 3** | Code (alle des Katalogs, die Zahl zählt der Test gegen `model.AlleCodes()`) × Feld 3 (`code`, `message`, `target`) | Rohtext aller Strings der Ausgabe nie `[REDACTED]` | Z11 |

**Zweig-Menge W** (Modell-Menge der Verzweigungsstellen des Schreibpfades eines `Edm.String`-Textes, `AGENTS.md` §3.17): Die Zählprüfung (`TestMarkerZweige`) parst die Dateien des Pakets mit `go/parser` und zählt in den Funktionen `ausgabe.wert`, `ausgabe.stringWert`, `ausgabe.text`, `pruefeWert` und `ausgeben` die Knoten `IfStmt`, `CaseClause` (mit `default`), `CommClause`, `ForStmt` und `RangeStmt`; die Zahl ist **18**, gleich den Zeilen unten. Fehlt eine Funktion oder weicht die Zahl ab, ist der Test rot, und der Autor der Änderung ergänzt die Zeile mit Fänger. Die Operanden von `||` und `&&` in einer Bedingung zählt die Prüfung nicht (benanntes Negativ: W09 und W14 führen jeden Operanden als Element von M1 und M4, Z03 und Z05).

| Nr. | Funktion · Stelle | Fänger |
|---|---|---|
| W01 | `wert` · `case model.ArtNull` | `TestRundlauf` (`null` bei nullbaren Properties) |
| W02 | `wert` · `case model.ArtString` (ruft `stringWert`) | Z01 bis Z04 |
| W03 | `wert` · `case model.ArtBoolean, model.ArtInt32` | `TestRundlauf` |
| W04 | `wert` · `case model.ArtInt64, model.ArtDecimal` | `TestRundlauf`, `TestZahlform` |
| W05 | `wert` · `default` (`guid`, `date`, `datetimeoffset`); die Art `redigiert` erreicht ihn nie, weil `pruefeWert` sie abweist | `TestRundlauf`; M4 (Art `redigiert` ist `ErrUngueltig`) |
| W06 | `stringWert` · `if s == Marker` (beide Seiten) | Z01 (wahr), Z03 (falsch) |
| W07 | `text` · `if !utf8.ValidString` | `TestUngueltigesUTF8AnVierStellen` (Namen von Property, Set, Typ und Navigation aus einem von Hand gebauten Modell); die Stelle des Wertes erreicht erst M5 über `pruefeWert`, das ihr vorausgeht |
| W08 | `text` · `range len(s)` | `TestEscapes`; Z03 (Text leer, 1 bis 65 547 Zeichen) |
| W09 | `text` · `case c == '"' \|\| c == '\\'` | `TestEscapes`; Z03 (Rest `"` und `\` einzeln) |
| W10 | `text` · `case c < 0x20` | `TestEscapes`; Z03 (Rest Tabulator, LF, CR, U+0000) |
| W11 | `text` · `default` | Z03 |
| W12 | `pruefeWert` · `if w.Art == model.ArtNull` | Z05 (Art `null` mit dem Text) |
| W13 | `pruefeWert` · `if !p.Nullable \|\| w.Text != ""` | Z05 (Text `[REDACTED]` an nullbarer Property) |
| W14 | `pruefeWert` · `if !bekannt \|\| edm != p.EdmType` | Z05 (Art `redigiert`; Art `string` an 7 fremden Typen); `TestArtGegenTyp` |
| W15 | `pruefeWert` · `if !kanon(w.Text)` | Z05 (7 Arten mit dem Text), Z06 (UTF-8); `TestTextKanonisch` |
| W16 | `ausgeben` · `if err := ctx.Err()` | Z08 |
| W17 | `ausgeben` · `if err != nil` nach `Write` | Z07 |
| W18 | `ausgeben` · `if n < len(body)` | Z07 (`len−1` und `len−5`) |

**Tabelle *Zusage · Fänger*:**

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z01 | Der Text `[REDACTED]` einer Art `string` steht als `"` + Escape von `[` + `REDACTED]` + `"` (`SPEC-169` Nr. 4), an jeder Stelle (7) und in jedem Modus (4); die Adresse der Stelle Schlüssel bleibt in der Form von Nr. 2 (`Alles('%5BREDACTED%5D')`: `'` bleibt unkodiert, `[` und `]` sind `%5B` und `%5D`) | `TestMarkerBytes`: E × M2 = 28, Orakel Bytes aus Bytewerten, bei Stelle Schlüssel auch `@odata.id` | Zweig entfällt (Text roh) · Escape nur in der Stufe `minimal` · Escape nur für Nichtschlüssel-Properties | andere |
| Z02 | Kein String-Wert der Ausgabe hat den Rohtext `[REDACTED]` (Allaussage über jede Stelle, jeden Modus, jeden Text der Klasse und der Nachbarn) | `TestMarkerKeinRohtext`: M1 × M2 = 4004; Orakel: Zerlegung der Bytes in Strings mit eigener Schleife (Anführungszeichen, Backslash überspringt das nächste Byte), weder Schreiber noch Leser; Zählprüfung 4004 | Escape nur für das Element `Entities[0]` (Fall der Schleife über die Elemente) · Escape nur in `SchreibeEntity` | andere |
| Z03 | Jeder andere Text bleibt in der Form von `SPEC-169` Nr. 1 (byte-gleich), auch jeder, der dem Marker gleicht | `TestMarkerNachbarn`: N × M2 = 3976 (Z01 28 + Z03 3976 = 4004), Orakel Referenzkodierer im Test und `encoding/json` | Vergleich mit `strings.EqualFold` · `strings.HasPrefix(s, "[REDACTED")` · `strings.Contains` · `strings.TrimSpace(s) == Marker` · Länge zehn und `[` am Anfang | andere |
| Z04 | Rundlauf (`SPEC-169` Nr. 8): `ODataCodec` liest `Art` `string`, den Text, `Schluessel` und `Uebersprungen` gleich, für jeden Text von M1, an jeder Stelle, in beiden Zahlformen | `TestMarkerRundlauf`: M3 = 2002, Zählprüfung 2002, Orakel erwartete Entity und `encoding/json` | Escape lässt das folgende Zeichen aus (`[` + `EDACTED]`) · Zweig entfällt (Leser liest `redigiert` bzw. `schluessel`) | andere |
| Z05 | Der Text `[REDACTED]` bei jeder anderen Art und an jedem anderen Typ ist `ErrUngueltig`; nichts in `dst` | `TestMarkerAndereArten`: M4 = 16, Zählprüfung 16 | `pruefeWert` lässt die Art `redigiert` durch · die Prüfung `edm != p.EdmType` entfällt · die Prüfung von `guid` akzeptiert zehn Zeichen | andere |
| Z06 | Ungültiges UTF-8 vor oder hinter dem Marker ist `ErrUngueltig`; nichts in `dst` | `TestMarkerUTF8`: M5 = 16 | beide Wächter des UTF-8 entfallen (die Prüfung der Art `string` in `artPruefer` und die in `text`) · Texte mit dem Präfix `[REDACTED` nehmen den Zweig des Escapes, und die Prüfung der Art `string` entfällt. **Akzeptiertes Negativ für die Isolation des Wächters in `artPruefer`:** *Grund:* zwei UTF-8-Wächter prüfen denselben Fehler (`artPruefer` vor dem Schreiben, `text` als Eigenschaft von `text`); für jeden Wert, den `artPruefer` abweist, weist `text` denselben Fehler ab, der Wächter in `artPruefer` ist verhaltensäquivalent, ein isolierter Fänger für ihn existiert nicht. *Folge im Betrieb:* entfällt er allein, ändert sich kein Byte und kein Fehler; entfällt der in `text`, fangen `TestUngueltigesUTF8AnVierStellen` und `TestMarkerZweige` (nicht `TestMarkerUTF8`); entfallen beide, fängt `TestMarkerUTF8`. Fänger ist das Paar | andere |
| Z07 | Fehler von `dst` und kurzer Schreibvorgang mit dem Text `[REDACTED]`: Fehler unverändert, `io.ErrShortWrite`, `Darstellung` Nullwert, ein `Write`; die Länge bemisst sich an den geschriebenen Bytes mit dem Escape | `TestMarkerDst`: M6 (Fälle 1 bis 3 × 2 Operationen = 6) | die Prüfung `n < len(body)` zählt die Länge ohne die 5 Bytes des Escapes | andere |
| Z08 | `ctx` beendet bei dem Text `[REDACTED]`: Fehler des Kontexts unverändert, nichts in `dst`; ein ungültiges Argument geht dem Kontext vor | `TestMarkerKontext`: M6 (Fälle 4 und 5 × 2 Operationen = 4) | der Kontext wird vor der Prüfung der Argumente gefragt | andere |
| Z09 | Der Schreiber verändert kein Argument und liefert gleiche Bytes bei gleichen Argumenten (`SPEC-168` Nr. 9) | `TestMarkerArgumente`: M7 = 286 | der Escape ersetzt `Properties[i].Wert.Text` im übergebenen Slice | andere |
| Z10 | Die Zweig-Menge W ist vollständig: Zahl der Verzweigungsstellen in den fünf Funktionen von W (`wert`, `stringWert`, `text`, `pruefeWert`, `ausgeben`) gleich Zahl der Zeilen (18), jede Zeile mit Fänger; Verzweigungen in `artPruefer`, `zahl` oder neuen Hilfsfunktionen zählt der Test nicht | `TestMarkerZweige`: Zählprüfung mit `go/parser` | ein zweiter `if` in `stringWert` ohne Zeile in W | andere |
| Z11 | Kein anderer String der Ausgabe kann `[REDACTED]` sein: Felder des Fehlerobjekts aller Codes (Namen, Kontext, `@odata.type`, `@odata.id` über Z02) | `TestMarkerFehlerobjekt`: M8, Rohtext-Zerlegung wie Z02, Zählprüfung gegen `model.AlleCodes()` | `message` eines Codes im Katalog gleich `[REDACTED]` | andere |
| Z12 | **Akzeptiertes Negativ:** die Schlüssel-Property mit U+0000 (der Leser lehnt sie als `schluessel` ab) | kein Fänger; `SPEC-169` *Ungeprüft*, zweiter Absatz. Grund: `Schluessel` mit U+0000 ist `ErrUngueltig`, der Mock bildet beide aus demselben Zustand. Folge: ein Aufrufer mit abweichendem `Schluessel` bekommt einen Body, den der Leser ablehnt | — | andere |
| Z13 | **Akzeptiertes Negativ, Prüfgrenze:** die Menge der Texte der Art `string` ist unbeschränkt (jeder gültige UTF-8-Text, jede Länge); geprüft sind die 143 Texte von M1, Rest bis 65 537 Zeichen | kein Fänger über M1 hinaus. Grund: der Schreiber unterscheidet nur die Gleichheit des ganzen Textes mit dem Marker (W06), der Leser liest allein den Rohtext besonders (`SPEC-126` Nr. 10 f), alle anderen Texte laufen durch dieselbe Schleife (W08 bis W11). Folge im Betrieb: ein Text außerhalb von M1, den ein Code anders behandelte, träte nur nach einer Änderung am Schreibpfad auf; Z10 meldet jede neue Verzweigung in den fünf Funktionen von W und jede Änderung ihrer Bedingung; eine Verzweigung in `artPruefer`, `zahl` oder einer neuen Hilfsfunktion meldet der Test nicht (die Funktionsliste der Zählprüfung ist eine Setzung, kein Aufruf-Graph) | — | andere |

**Bericht des Implementers** (2026-11-07; `AGENTS.md` §3.9, §3.10, §3.17). Code: `ausgabe.stringWert` mit der Konstante `markerText` (`schreiber_ausgabe.go`), `ausgabe.wert` ruft sie für `Edm.String` (`schreiber_bau.go`); der Leser ist unverändert. Tests: `TestMarkerBytes` (Z01), `TestMarkerKeinRohtext` (Z02), `TestMarkerNachbarn` (Z03), `TestMarkerRundlauf` (Z04), `TestMarkerFehlerobjekt` (Z11) in `schreiber_marker_test.go`; `TestMarkerAndereArten` (Z05), `TestMarkerUTF8` (Z06), `TestMarkerDst` (Z07), `TestMarkerKontext` (Z08), `TestMarkerArgumente` (Z09), `TestMarkerZweige` (Z10) in `schreiber_marker_zweige_test.go`. Der Escape und jeder Wert mit Backslash entstehen aus Bytewerten (`esc`, `bs`, `string([]rune{…})`); Rohbytes beider Testdateien und beider Quelldateien sind gegengelesen (`LC_ALL=C grep -P` auf Backslash, `u` und vier Hex-Ziffern: kein Treffer; Zeichen jenseits von ASCII nur als deutsche Umlaute in Kommentaren und Namen).

*Mengen* (die Orakel teilen keinen Code mit Schreiber oder Leser: byteweiser Kodierer `mkEnc`, Adress-Kodierer `mkAdresse`, Zerleger `mkStrings`, `encoding/json`, erwartete Entity aus dem Text):

| Menge | Zählung im Test | Elemente | davon `ErrUngueltig` |
|---|---|---|---|
| M1 | `len(mkTexte())` | 143 (E 1, N 142) | — |
| M1 × M2 (Z01) | Elemente der Schleife | 28 | 0 |
| M1 × M2 (Z02) | Elemente der Schleife | 4004 | 24 (U+0000 an Schlüssel und beide, 3 Texte × 2 × 4) |
| N × M2 (Z03) | Elemente der Schleife | 3976 | 24 |
| M3 (Z04) | Elemente der Schleife | 2002 | 12 |
| M4 (Z05) | Fälle | 16 | 16 |
| M5 (Z06) | Fälle | 16 | 16 |
| M6 (Z07, Z08) | Fälle | 6 und 4 | 0 und 2 (Z08: zwei Fälle mit Fehler des Kontexts, zwei mit `ErrUngueltig`) |
| M7 (Z09) | Texte × Operationen | 286 | 0 |
| M8 (Z11) | Codes × 3 Felder | 37 × 3 = 111 (der Test zählt gegen `len(model.AlleCodes())`) | — |
| W (Z10) | Zeilen, Knoten, Funktionen | 18, 18, 5 | — |

M1 zählt 143 Einträge, aber nur 139 verschiedene Texte: die Streichung an Stelle 0 gleicht dem Suffix der Länge 9, die an Stelle 9 dem Präfix der Länge 9, und `x` als Rest davor und dahinter der Einfügung von `x` vor Stelle 0 und 10; die Schleifen führen die Einträge, wie die Tabelle sie zählt (offen, siehe §6). Z10 prüft zusätzlich den Wortlaut jeder Verzweigung gegen die Zeile von W und dass jeder Test, den eine Zeile nennt, im Paket steht.

*Zusage · Mutation · roter Test* (`mutationslauf.sh`, Klasse `andere`, ein Prozess je Mutation, Image `odata-recorder:race`, `--memory=4g`; 48 Mutationen in einem Lauf, 60,9 s, Host-Last beim Lauf 3,2 von 4 CPUs laut `last=` der Ausgabe, Messender das Werkzeug; kein Fänger ist wahrscheinlichkeitsbehaftet, kein Test nutzt Nebenläufigkeit oder Zufall, eine Rate gibt es nicht):

| Zusage | Mutationen (jede rot im genannten Test) | Test |
|---|---|---|
| Z01 | Zweig `s == markerText` entfällt · Escape nur in der Stufe `minimal` · nur für Nichtschlüssel-Properties · nur in der Zahlform `Zahl` · Hex-Ziffer `B` statt `b` | `TestMarkerBytes` |
| Z02 | Escape nur im ersten Element der Collection · nicht in der Collection · Zweig entfällt | `TestMarkerKeinRohtext` |
| Z03 | `EqualFold` · `HasPrefix(…, "[REDACTED")` · `HasSuffix(…, "REDACTED]")` · `Contains(…, "REDACTED")` · `TrimSpace` · Länge zehn und `[` am Anfang · Länge zehn mit beliebigem ersten Zeichen · mit beliebigem letzten · `ToUpper` · `Contains` des ganzen Markers · Text endet mit dem Marker · beginnt mit ihm | `TestMarkerNachbarn` |
| Z04 | Escape lässt das folgende Zeichen aus · Zweig entfällt · Ziffern `005a` statt `005b` | `TestMarkerRundlauf` |
| Z05 | Art `redigiert` durchgelassen · Typprüfung `edm != p.EdmType` entfällt · Typprüfung entfällt nur für `Edm.DateTimeOffset` · nur für `Edm.Boolean` · `guid` mit zehn Zeichen · `null` mit Text · `null` mit dem Text `[REDACTED]` | `TestMarkerAndereArten` |
| Z06 | beide Wächter des UTF-8 entfallen (Art `string` in `artPruefer` und `text`) · Präfix `[REDACTED` nimmt den Zweig des Escapes bei entfallener Prüfung der Art `string`: beide rot, in einer Kopie im Container gesehen (`sed` auf der Kopie, nicht im Arbeitsbaum). Einzeln: Prüfung der Art entfällt = grün (gleichwertig, `text` fängt), Prüfung in `text` entfällt = `rot-paket` (`TestMarkerUTF8` grün, `TestUngueltigesUTF8AnVierStellen` und `TestMarkerZweige` rot) | `TestMarkerUTF8` |
| Z07 | Länge ohne die 5 Bytes des Escapes · Länge minus 1 · Fehler von `dst` verpackt · zweiter `Write` | `TestMarkerDst` |
| Z08 | Kontext vor der Prüfung der Argumente (Entity, Collection je eine) · Kontextfehler verpackt · Kontext nach dem `Write` | `TestMarkerKontext` |
| Z09 | die Property im übergebenen Slice wird verändert | `TestMarkerArgumente` |
| Z10 | zweiter `if` in `stringWert` · zweiter `if` in `pruefeWert` · `for` in `wert` · ein `case` in `wert` geteilt | `TestMarkerZweige` |
| Z11 | `message` eines Codes · `message` aller Codes · `target` aller Codes gleich `[REDACTED]` | `TestMarkerFehlerobjekt` |

Ergebnis des Laufs: 46 `rot` (darunter `m05b` im zweiten Lauf, im ersten ein Übersetzungsfehler der Mutation), 1 `rot-paket` (Z06, Prüfung in `text`), 1 `gruen` (Z06, Prüfung der Art; gleichwertig). Nicht vom Plan genannte Elemente, rot gesehen: Z03 die Ersetzung an Stelle 0 und 9 (`m03g`, `m03h`), `ToUpper`, `Contains` des Markers; Z05 `Edm.DateTimeOffset` und `Edm.Boolean` einzeln; Z01 die Zahlform `Zahl`.

*Abdeckung* (`go test -coverprofile`, Paket `jsoncodec`, Image `odata-recorder:race`): das ganze Paket deckt 94,1 % der Anweisungen, jede Anweisung von `wert`, `stringWert`, `text`, `pruefeWert` und `ausgeben` läuft (der eine Block ohne Lauf, `schreiber.go` Zeilen 87 und 88, liegt in `SchreibeFehler`). Die neuen Tests allein lassen in diesen Funktionen nur die Zweige aus, die W01, W03, W04, W05 und W07 den älteren Tests zuweist (`TestRundlauf`, `TestZahlform`, `TestUngueltigesUTF8AnVierStellen`).

Abweichungen von diesem Plan: (1) Z01 nannte die Adresse `%27…%27`; `SPEC-169` Nr. 2 lässt `'` unkodiert, die Zelle ist berichtigt. (2) Z06: die genannte Mutation allein ist grün, weil zwei Wächter denselben Fehler fangen; die Zelle nennt jetzt beide.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `next` → `in-progress`): `slice-odata-json-eingabe` liegt in `done/` (der Zweig ändert `jsoncodec` und `SPEC-125`/`SPEC-168` zugleich; ein Start davor erzeugt Konflikte im selben Paket). Der Vertrag steht in der Spezifikation, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Stand beim Start** (Planner, 2026-11-07): erfüllt, `slice-odata-json-eingabe` liegt in `done/`. Der Slice liegt in der Linie B neben `slice-v2-payload-schreiben` (Auftrag 6 in §3; `jsoncodec/` gegen `jsoncodecv2/`, getrennte Pakete) und berührt keine Datei der Gate-Kette (Übergangsregeln Nr. 7: weder `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` noch den Stempel; Stand beim Start: `slice-handbuch-meldungskatalog`, Linie C, ist der eine Slice der Gate-Kette in `in-progress/`); ändert Vertrag oder Code eine Datei der Gate-Kette (auch `.a-check.yml`, `.golangci.yml`), geht das als Rückgabe an den Planner (§6).
- **Größenmarke** (`SPEC-128`; Faktoren allein dort; Stand beim Start, Planner). Der Block mit `basis`, `schaetzung` und `marke` steht auf `main` nicht (ein Block ohne Zahlen der Tabelle wäre ein Platzhalter und machte `make groessenmarke` rot); im Zweig steht er seit dem Commit des Planner (unten, *Marke des Planner*). Rollenfolge: Der Architect legt den Block im Zweig an, der Planner setzt `basis`, `schaetzung` und `marke` **zusammen** in eigenem Commit vor dem ersten Code-Commit, dann der Implementer. **Teil des Architect (2026-11-07):** der Block unten mit den Pfaden (neue Dateien statt des Verzeichnisses, Tests eingeschlossen), `produktcode` 30 (eine Methode mit Kommentar, eine geänderte Zeile) und `fakes` 0 (`ziel` und `kontext` aus `schreiber_helfer_test.go` genügen). `basis` fehlt im Block, weil sie der Commit ist, der diesen Block mitbringt, und weil `SPEC-128` Nr. 3 k einen Block mit nur einem der drei Planner-Schlüssel als Eingabefehler liest (`make groessenmarke` wäre rot); `basis` ist der Commit dieser Entscheidung. **Richtwert, nicht bindend:** 13 Zeilen der Tabelle (Z01 bis Z13, keine nebenläufig), *S* = 80·13 + 30 = 1070, *M* = 100·⌈135·1070 / 10000⌉ = 1500, keine Schnittfrage (*S* < 2223). Das Register führt die Überschätzung (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`, 12×); die Marke folgt trotzdem der Formel.

```groessenmarke
pfad: src/adapters/driven/odata/jsoncodec/schreiber_ausgabe.go
pfad: src/adapters/driven/odata/jsoncodec/schreiber_bau.go
pfad: src/adapters/driven/odata/jsoncodec/schreiber_marker_test.go
pfad: src/adapters/driven/odata/jsoncodec/schreiber_marker_zweige_test.go
produktcode: 30
fakes: 0
basis: 94dafe9
schaetzung: 1070
marke: 1500
```

**Marke des Planner** (2026-11-07, vor dem ersten Code-Commit, in eigenem Commit): `basis` ist `94dafe9`, der Commit des Architect mit dem Block; `schaetzung` 1070 und `marke` 1500 sind der Richtwert des Architect nach `SPEC-128` Nr. 1 (13 Zeilen der Klasse `andere`, *S* = 80·13 + 30 + 200·0 = 1070, *M* = 100·⌈135·1070 / 10000⌉ = 1500); `make groessenmarke` im Zweig rechnet dieselben Zahlen. *S* liegt unter 2223, es gibt keine Schnittfrage und keinen Schnitt (§3 Nr. 7). Die Zeilen Z12 und Z13 sind akzeptierte Negative und zählen als `andere` (`SPEC-128` Nr. 3 e). Das Register führt die Überschätzung durch die Formel (`BEO-REPO/schaetzformel-ueberschaetzt-den-slice`); die Marke folgt trotzdem der Formel, und eine Messung darüber trüge einen Eintrag mit Grund (`SPEC-128` Nr. 5).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): eine Randform, die das Raster in §6 nicht nennt und die der Architect nicht vor dem Code entscheidet (die Form des Textes selbst ist entschieden, §3 Nr. 1).

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

- Der Schreiber `jsoncodec` und der Leser `ODataCodec` sind nicht als Paar über die Klasse der Texte geprüft (Beobachtung „Schreiber und Leser nicht als Paar geprüft“; das Register legt `slice-v2-payload-schreiben` bei seiner Closure an, dort `BEO-REPO/` mit dem Beleg dieses Fundes) — **Ausgang:** weiter offen; Fänger: Rundlauf als Schleife über die Klasse (Z04, M3, DoD, zweiter Punkt). **Ausgang (Planner, Closure):** eingetreten und aufgelöst für den V4-Schreiber: `TestMarkerRundlauf` läuft über M3 (2002 Elemente), Evidenz im Register `BEO-REPO/schreiber-und-leser-nicht-als-paar-geprueft`.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster unten ist vor dem ersten Code-Commit gefüllt, jede Zelle mit Ort oder „gilt nicht“ und Grund; eine Randform, die der Implementer darüber hinaus findet, geht an den Architect — **Ausgang:** entfallen (Review und Verifikation fanden keine Zelle ohne Ort und keine neue Randform; der Implementer gab keine zurück).
- Ein Escape ändert die Bytes, die ein bereits aufgezeichneter oder ausgelieferter Wert trägt — **Ausgang:** eingetreten (Escape gewählt, §3 Nr. 1); die Folge im Betrieb steht in `SPEC-169` Nr. 4: nur dieser eine Text hat andere Bytes in der Antwort des Mock, der Wert ist gleich, aufgezeichnete Bytes sind nicht betroffen.
- Pflicht-Mengen und Zweig-Menge mit Zählprüfung (Aufträge 4 und 5) — **Ausgang:** Teil des Architect erledigt (Mengen M1 bis M8, Zweig-Menge W01 bis W18, Tabelle Z01 bis Z13 in §3); Teil des Implementers geliefert (Zählprüfung in jedem Test, `TestMarkerZweige`, Bericht in §3); der Teil, den ein Befund auslöst, ist abgeschlossen: Review (121 eigene Mutationen) und Verifikation (20 Mutationen an Elementen, die weder Plan noch Review nannten) blieben ohne grüne Mutation außer der äquivalenten (Z06). **Ausgang:** entfallen.
- M1 zählt 143 Einträge, aber nur 139 verschiedene Texte (vier Einträge nennen einen Text, den ein anderer auch nennt, Bericht in §3); die Schleifen führen die Einträge wie die Tabelle — **Ausgang:** eingetreten, akzeptiert (Planner, Closure): die Zählprüfung zählt Einträge der Tabelle, nicht verschiedene Texte; Folge: keine, die vier Doppelten prüfen dieselben Bytes zweimal und kein Element der Klasse fehlt (Verifikation I-1).
- Schnittfrage und Marke (Auftrag 7, §4) — **Ausgang:** Schnittfrage entschieden (kein Schnitt, §3 Nr. 7); die Marke ist gesetzt (Planner, 2026-11-07, §4: `basis` `94dafe9`, `schaetzung` 1070, `marke` 1500, ein Commit) — **Ausgang:** entfallen.
- Linie B neben `slice-v2-payload-schreiben` (Auftrag 6): `spezifikation.md` schreiben beide Zweige (verschiedene Nummern); dieser Zweig fügt eine Zeile an den Kopf der Historie in §7 der Spezifikation und Absätze in `SPEC-168` Nr. 4 sowie `SPEC-169` Nr. 4, 8 und *Ungeprüft* ein, der V2-Zweig ändert `SPEC-168` Nr. 10 und Zeilen in `SPEC-172`, `SPEC-178`; ein Konflikt ist in der Historie möglich. Die Reihenfolge der Merges entscheidet der erste Verifier-Pass, der zweite Zweig nimmt `main` per Merge auf. **Merge-Hinweis (Planner, 2026-11-07):** `slice-v2-payload-schreiben` steht im Verifier-Lauf und wird voraussichtlich zuerst integriert; dieser Zweig nimmt danach `main` per Merge auf. Ein Konflikt am Kopf der Historie von `spec/spezifikation.md` (beide Zweige fügen dort eine Zeile ein) wird gelöst, indem **beide Zeilen bleiben**; sonst ist kein Konflikt zu erwarten (verschiedene Nummern) — **Ausgang:** entfallen (Planner, Closure: `slice-v2-payload-schreiben` liegt in `main`, der Merge von `main` in den Zweig verlief ohne Konflikt).
- Gate-Kette (Stand beim Start in §4): ein Bedarf an einer Änderung der Gate-Kette — **Ausgang:** entfallen (keine Datei der Gate-Kette berührt, `make gates` grün); Adresse war: Planner. Dazu die Zählprüfung (Z10) liest Quelltext des Pakets mit `go/parser` aus dem Arbeitsverzeichnis des Tests; meldet `make test`, `make lint` oder `make a-check` dagegen etwas, geht das als Rückgabe an den Planner, nicht als Lockerung im Code.
- Rückgabe an den Planner (`AGENTS.md` §3.13): der DoD-Punkt „Vertragstext nennt keine Ausnahme mehr“ ist auf `main` schon erfüllt (Nr. 8 und *Ungeprüft* nannten nie eine); abgehakt in §2 mit Beleg (Planner, 2026-11-07) — **Ausgang:** entfallen.

**Raster *Operation × Argumentart*** (Zellen gefüllt vom Architect, 2026-11-07; `Nr.` bezieht sich auf `SPEC-168` bzw. `SPEC-169`):

| Operation | Text `[REDACTED]` | Text mit Rest davor oder dahinter, ähnlicher Text | anderer Text, den der Leser besonders liest | Text in `Collection`-Property | Wert anderer Art mit dem Text | ungültiges UTF-8 | Abbruch, Fehler von `dst` | Wiederholung |
|---|---|---|---|---|---|---|---|---|
| Entity schreiben | `SPEC-169` Nr. 4, Absatz *Der Text `[REDACTED]`* (Escape) | `SPEC-169` Nr. 4 (jeder andere Text in der Form von Nr. 1), Nr. 1 | gilt nicht: `SPEC-126` Nr. 10 f nennt allein den Rohtext; U+0000 in einer Schlüssel-Property: `SPEC-169` *Ungeprüft*, zweiter Absatz | gilt nicht: eine Property des Typs `Collection(…)` ist nicht unterstützt und `ErrUngueltig` (`SPEC-168` Nr. 3 c) | `SPEC-169` Nr. 4 (`ErrUngueltig`), `SPEC-168` Nr. 4 | `SPEC-169` Nr. 1 (`ErrUngueltig`) | `SPEC-168` Nr. 1 c und Nr. 6 | `SPEC-168` Nr. 9 |
| Collection von Entities schreiben | wie Entity, für jedes Element (`SPEC-168` Nr. 3, `SPEC-169` Nr. 6) | wie Entity | wie Entity | gilt nicht: wie Entity | wie Entity | wie Entity | wie Entity | wie Entity |
| `SchreibeFehler`, `SchreibeAnzahl`, `Adresse` | gilt nicht: die Felder des Fehlerobjekts sind der Katalog (`SPEC-168` Nr. 5 f, Fänger Z11), die Anzahl sind Ziffern (`SPEC-169` Nr. 7), die Adresse ist prozentkodiert (`SPEC-169` Nr. 2, `[` ist `%5B`) | gilt nicht: wie links | gilt nicht: wie links | gilt nicht: kein Wert | gilt nicht: kein Wert | `SPEC-168` Nr. 5 f (Fehlerobjekt), Nr. 8 (Adresse) | `SPEC-168` Nr. 1 c (`Adresse` hat keinen `ctx`, Nr. 8) und Nr. 6 | `SPEC-168` Nr. 9 |

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

**Stand der Closure** (Planner, 2026-11-07, nach dem Merge von `main` im Zweig; das reine `git mv` nach `done/` folgt als eigener Commit, danach der Merge nach `main`, kein Push). Belege: Review-Report `docs/reviews/2026-11-07-slice-v4-schreiber-marker-text-review.md` (0 HIGH, 0 MEDIUM, 2 LOW, 2 INFO; 121 eigene Mutationen), Verifikations-Report `docs/reviews/2026-11-07-slice-v4-schreiber-marker-text-verifikation.md` (bestanden, Auflagen an den Planner; `make gates` Exit 0, 20 Mutationen an nicht genannten Elementen: 17 rot, 3 `rot-paket`, 0 grün). Review und Verifikation sind je einmal gelaufen (Durchstich vor Härtung); es gab keine Fix-Runde.

**Gegenstand:** geliefert. `jsoncodec.NewSchreiber()` schreibt den Text `[REDACTED]` einer Art `string` als `"` + Escape von `[` (Backslash, `u`, `005b`) + `REDACTED]` + `"` (`ausgabe.stringWert`, `schreiber_ausgabe.go`; `schreiber_bau.go` ruft sie im `case model.ArtString`); jeder andere Text bleibt in der Form von `SPEC-169` Nr. 1. Der Rundlauf von `SPEC-169` Nr. 8 gilt damit für jeden Text (M3, 2002 Elemente). Der Leser, `jsoncodecv2` und das Binary blieben unverändert.

**Größen:** Messung 771 von Marke 1500 (`make groessenmarke`: ok), Schätzung *S* = 1070, Ende ÷ *S* = 0,72. Die Marke ist nicht verändert. Datenpunkt für die Neukalibrierung, Register `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`.

**Mutationen:** Tabelle *Zusage · Fänger* Z01 bis Z13 mit der Mutation je Zusage in §3 (48 Mutationen des Implementers, 121 des Reviews, 20 des Verifiers an nicht genannten Elementen; keine Wahrscheinlichkeit im Spiel). Abdeckung 100 % der berührten Funktionen (`stringWert`, `text`, `wert`, `pruefeWert`, `artPruefer`, `ausgeben`).

**Auflagen des Verifiers, erledigt:** F-1 (Z06: Wortlaut berichtigt, akzeptiertes Negativ mit Grund und Folge, Fänger ist das Paar), F-2 (Z10 und Z13: Zusage auf die fünf Funktionen von W beschränkt, §3.11), I-1 (M1: Ausgang in §6 gesetzt). Weder Code noch Tests wurden geändert.

**Risiko-Ausgänge (§6):** je Punkt im Plan selbst. Entfallen: Randformen, Pflicht-Mengen, Schnittfrage und Marke, Linie B, Gate-Kette, Rückgabe zum DoD-Punkt der Ausnahme. Eingetreten: Escape ändert Bytes (entschieden, Folge in `SPEC-169` Nr. 4), M1 143 Einträge/139 Texte (akzeptiert, Folge keine), Schreiber und Leser als Paar (aufgelöst, Rundlauf als Schleife). Weiter offen: keiner.

**Doku-Update:** der öffentliche Vertrag steht in `spec/spezifikation.md` (Architect, vor dem Code); `make handbuch-check` ist grün, das Handbuch nennt den Text nicht, kein Katalogeintrag kam hinzu.

**Steering-Loop-Einträge:**

- Lerneintrag, gezählt: der Rundlauf als Schleife über die Klasse der Texte hat die Lücke zwischen V4-Schreiber und Leser geschlossen, die `slice-v2-payload-schreiben` fand; zweiter Beleg zu `BEO-REPO/schreiber-und-leser-nicht-als-paar-geprueft` (2×, offen).
- Lerneintrag, gezählt, neue Beobachtung: eine Mutation bleibt grün, weil ein zweiter Wächter denselben Fehler fängt (UTF-8 in `artPruefer` und `text`); die Zeile der Tabelle sagte „Fänger“, der Fänger ist das Paar. Register `BEO-REPO/aequivalente-mutation-hinter-doppeltem-waechter` (ein Beleg, offen). Das Mittel ist das akzeptierte Negativ mit Grund und Folge (`AGENTS.md` §3.15), die Zeile nennt das Paar als Fänger.
- Lerneintrag, gezählt, neue Beobachtung: die Zusage „meldet jede neue Verzweigung“ war weiter als die Funktionsliste der Zählprüfung (fünf Funktionen, kein Aufruf-Graph). Register `BEO-REPO/zusage-ueber-verzweigungen-reicher-als-die-funktionsliste` (ein Beleg, offen). Das Mittel ist der Wortlaut, der die Funktionsliste nennt (`AGENTS.md` §3.11).
- Lerneintrag, gezählt: die Zweig-Menge W mit Zählprüfung und das Abdeckungsprofil (100 % der berührten Funktionen) haben das Muster „Zweig ohne Lauf“ am Schreiber geschlossen; Evidenz zu `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (7×, der Eintrag steht bei der Schwelle; Ausgang beim Lese-Schritt von `welle-odata-v2`).
- Lerneintrag, gezählt: die Formel von `SPEC-128` überschätzt den Slice (0,72). Evidenz zu `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`.
- Keine benannte Spec-Lücke neu: `SPEC-169` Nr. 4 trägt die Entscheidung vor dem Code. Kein `liegt in`: in diesem Slice wurde keine Regel an einem Zielort verkörpert.

**Register:** neu `BEO-REPO/aequivalente-mutation-hinter-doppeltem-waechter` und `BEO-REPO/zusage-ueber-verzweigungen-reicher-als-die-funktionsliste` (je ein Beleg `slice-v4-schreiber-marker-text`); Evidenz `slice-v4-schreiber-marker-text` zu `BEO-REPO/schreiber-und-leser-nicht-als-paar-geprueft`, `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` und `BEO-REPO/schaetzformel-ueberschaetzt-den-slice`.

**Drei Paarungen:** Anker (kein Feld `liegt in`, nichts verkörpert) · Folge-Slice (keine Sendung an einen anderen Slice) · Register (jede genannte Kennung `BEO-REPO/<slug>` existiert, jedes `evidence/` ist nicht leer) — die nächste Welle-Closure prüft sie erneut.

**Offener Rest:** das reine `git mv` nach `done/` und der Merge nach `main`; kein Push.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-06) und beim Start nachgezogen (Planner, 2026-11-07; Zähler = Zahl der Dateien in `evidence/`). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (28×, `AGENTS.md` §3.17), `BEO-REPO/randform-erst-im-review-entschieden` (33×, `AGENTS.md` §3.16), `BEO-REPO/suite-ohne-generator-fuer-klassen-der-spezifikation` (6×) und `BEO-REPO/unbeschraenkte-menge-ohne-benannte-pruefgrenze` (11×, beide `implement-slice` Schritt 14), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (14×) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (14×, beide `AGENTS.md` §3.15 und §3.16), `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (5×, `AGENTS.md` §3.18). Offen: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (5×, Schwelle erreicht, Verkörperung beim Lese-Schritt der Closure von `welle-mock` oder des nächsten Slice mit Nebenläufigkeit) gilt hier für die Zweig-Menge (Auftrag 5), nicht für Nebenläufigkeit; `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, verkörpert in `implement-slice` Schritt 14) für den Fake von `dst`; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (8×, `geplant` in `slice-guard-schreibwege`) für die Betriebsregel „nur `Edit` und `Write`“; `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (12×) für die Marke. Mit diesem Slice erreicht kein offener Eintrag ohne Ausgang 3×, außer den genannten.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
