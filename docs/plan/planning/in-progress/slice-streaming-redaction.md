# Slice slice-streaming-redaction: Streaming-Redaction von JSON-Bodies, fail-closed

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-14`, `LH-FA-02`, `LH-FA-05`, `LH-QA-02`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-118` (der Vertrag dieses Slice, vor dem Code geschrieben) · `SPEC-046` (Sicheres Streaming-Capture) · `SPEC-053` (Bedrohungsmodell, Sicherheits-Fixtures) · `LH-FA-14.a` · `LH-FA-02.b` · `LH-FA-05.a` (Key aus der redigierten Darstellung) · `SPEC-115` Nr. 6 · `SPEC-116` (Einleitung, Grenze des Bodys) · `SPEC-013`, `SPEC-017`, `SPEC-084`, `SPEC-085` · `SPEC-001` (`redaction.*`) · `SPEC-106` Nr. 1 (Grenze des Schreibers) · `ARC-001`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Das Domain-Modell liest einen Body als Strom (`RedigiereBody(dst, src, BodyEingang, BodyRegeln)`, `SPEC-118`), ersetzt String-Werte unter den konfigurierten Pfaden durch den Marker, erkennt Namen und Werte über Chunk-Grenzen, meldet `none`, `applied` oder `unsafe` und endet bei unbekanntem Typ, ungültigem oder unvollständigem JSON, Nicht-String-Wert unter einem treffenden Namen, Überschreiten der Grenze oder Abbruch fail-closed mit `ODR-E2401`; ins `dst` gelangt nie ein Byte eines geheimen Wertes, auch nicht bei `unsafe`.

**Übernimmt (Pfadmuster und Marker):** Grammatik und Bedeutung der Pfadmuster liefert `ParseJSONPfad` des Modells (`SPEC-115` Nr. 6: `$..name`, Name ohne Beachtung der ASCII-Schreibweise, Objektmitglieder in beliebiger Tiefe) und den Marker `[REDACTED]` (`SPEC-115` Nr. 1); der Redaktor ruft `JSONPfad.Passt` und führt keine zweite Grammatik und keinen zweiten Vergleich (Sendung von `slice-redaction-match-key`).

**Übernimmt (Pflicht der Aufrufer, Key nur aus redigiertem Body):** `slice-redaction-match-key` hat die Pflicht, den Body vor `BildeSchluessel` zu redigieren, bewusst an die Aufrufer gegeben (`SPEC-116` Einleitung, Grenze). Dieser Slice liefert die Funktion, die diese Pflicht erfüllbar macht, und legt sie in `SPEC-118` Nr. 10 fest (Key aus der Ausgabe, im Replay `MaxBytes = MAX_CAPTURE_BODY`, Fehler des Request-Bodys ist `ODR-E5001`); die Zusage selbst fängt kein Test dieses Slice (sie liegt bei den Diensten, siehe unten und §3, akzeptiertes Negativ).

**Übergaben an Folge-Slices** (jeder Nehmer führt die Sendung in seinem §1 mit der Kennung dieses Gebers): `slice-capture-pipeline` (Aufruf der Funktion, `Metadata` setzen, bei jedem Fehler `dst` verwerfen, Zustände nach Art des Fehlers, Strom entpackt übergeben, redigierte Request-Bytes für den Key), `slice-record-service` (Request-Body durch den Redaktor, Key aus der Ausgabe, `redaction.request=unsafe` mit `null`-Key) und `slice-replay-service` (Request-Body durch den Redaktor mit `MAX_CAPTURE_BODY`, Key aus der Ausgabe, Fehler ist `ODR-E5001` ohne Cursor-Fortschritt).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Tee, Puffergrenzen, Backpressure, Dekompression und das Schreiben des Blobs — `slice-capture-pipeline`; der Redaktor kennt nur einen Eingabestrom und einen Ausgabestrom und sieht `Content-Encoding` nie.
- Wahl von `capture_status` und der Zustände des Exchange nach einem Fehler, Setzen von `Metadata` aus dem Pfad — `slice-capture-pipeline`; der Redaktor meldet Grund und Status, nicht den Exchange.
- Anwendung beim Record und beim Replay (Key, Antwortcodes) — `slice-record-service` und `slice-replay-service`; hier ist es eine Funktion über Ströme.
- Redaction der Header und der Query — `slice-redaction-match-key`.
- XML außer der Antwort auf `/$metadata` — nach `SPEC-046` Nr. 3 unbekannt und damit fail-closed, nicht redigiert; auch das EDMX wird nicht umgeschrieben, nur auf `<` als erstes Zeichen geprüft.
- Weitere Pfadsprachen (`$.a.b`, Indizes, Platzhalter) — `SPEC-115` Nr. 6 lässt nur `$..name` zu; eine andere Form scheitert beim Start (`ODR-E2001`), der Redaktor sieht sie nie.

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

- [ ] `RedigiereBody` nach `SPEC-118`: Typ (`application/json`, `+json`, mit OData-Parametern; `application/xml` nur mit `Metadata`), strenge JSON-Prüfung, Marker nur für String-Werte unter treffenden Namen (Name nach Dekodierung der Escapes, in beliebiger Tiefe), Status `none`, `applied`, `unsafe`; Ausgabe byte-gleich bis auf die ersetzten Werte, unabhängig von der Zerlegung des Stroms.
- [ ] Fail-closed mit `*UnsicherFehler` (Grund, `ODR-E2401`): unbekannter MIME-Typ, `multipart/mixed`, `text/*`, ungültiges oder unvollständiges JSON, ungültiges UTF-8, Tiefe über 512, Zahl, Objekt oder Array unter treffendem Namen, Überschreiten von `MAX_CAPTURE_BODY`, Fehler von `src` oder `dst`; ins `dst` gelangt nie ein Byte eines geheimen Wertes, und der Speicher hängt nicht von der Größe des Bodys ab.
- [ ] Sicherheits-Fixtures aus `SPEC-053`, soweit der Redaktor sie fängt: Token über Chunk-Grenzen (jede Zweiteilung, einzelne Bytes), MIME-Mismatch, Abbruch mitten im Dokument, Redaction-Fehler; nicht gefangen und benannt: gzip-Bombe, Content-Encoding-Kette, langsamer Consumer, Queue-Überlauf, Crash zwischen Blob-Rename und DB-Commit, Middleware-Logging (`slice-capture-pipeline`, `slice-record-service`).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
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
| `src/hexagon/model/` | update | `redaktor.go` (`RedigiereBody`, `BodyEingang`, `BodyRegeln`, Typ-Klassifikation, `ErrUnsicher`, `UnsicherFehler` samt `Grund` und `Code()`) und `redaktor_json.go` (Zustandsautomat über Bytes); gleicher Ort wie die Funktionen aus `slice-redaction-match-key`, weil Record und Replay denselben Code brauchen und Services einander nicht importieren |
| `spec/spezifikation.md` | update | **vor dem Code geschrieben:** `SPEC-118` samt Zeigern in `SPEC-046`, `SPEC-115` Nr. 6 und `LH-FA-14.a`; im Slice nur noch Nachzug nach Review |
| Tests des Pakets `model` | neu | `redaktor_test.go` mit der einen Tabelle `bodyFaelle` (Eingabe, Typ, Erwartung von Hand ausgeschrieben, nicht berechnet), Zerlegungs-Eigenschaftstest, Speichertest mit erzeugtem Strom, Mutationen je Zusage nach `AGENTS.md` §3.10 |

Der Vertrag steht in `SPEC-118` **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16). Die Fänger sind Namen geplanter Tests; der Implementer legt sie unter diesem Namen an oder nennt den Namen, den er stattdessen vergibt, im Bericht. Alle Tests teilen die Tabelle `bodyFaelle`; ein Fall trägt Name, `ContentType`, `Metadata`, Eingabe, erwartete Ausgabe (Bytes), Status, Grund und ob er das Geheimnis `S3CR3T-7f` unter einem treffenden Namen führt. Die Tabelle ist gegen `SPEC-118` gebaut, nicht gegen den Code: jede Ausgabe steht ausgeschrieben da, und die Erwartung zitiert den Marker als Text `[REDACTED]`, nicht als Konstante. Der Speichertest läuft ohne `t.Parallel`.

**Zusage · Fänger** (jede Zeile löst der Reviewer mit einer Mutation auf; Nummern sind die von `SPEC-118`; „Mutation“ heißt: der Fänger wird aus dem genannten Grund rot, weil die Eingaben sich in genau dem Merkmal unterscheiden, das die Zusage nennt):

| Zusage | Fänger |
|---|---|
| Argumente: `MaxBytes < 1`, `dst` oder `src` `nil`, Nullwert von `JSONPfad` ergeben `ErrUngueltig`, Status leer, nichts gelesen, nichts geschrieben; leere `Pfade` gültig; die Prüfung kommt vor jeder anderen (Nr. 1) | `TestBodyArgumente`: `MaxBytes` 0 und -1; `Pfade` nur `JSONPfad{}` mit Body `{"password":"x"}`; `src` mit Zähler der Reads (0); Fall `MaxBytes=0` mit `ContentType` `text/plain`: `ErrUngueltig`, nicht `typ`. Mutationen: Nullwert zulassen (Ergebnis `none` ohne Fehler): rot; Argumentprüfung nach der Typprüfung: rot (Fall mit `text/plain`); leere `Pfade` als Fehler: rot |
| Leere Eingabe ist `none`, ohne Ausgabe, für jeden Typ, vor der Typprüfung; nur Leerraum ist es nicht (Nr. 2) | `TestBodyLeer`: 0 Bytes mit `nil`, `text/plain`, `application/json`, `application/xml` mit `Metadata`, dazu ein `src` aus nur leeren Reads und dann EOF; `" "` und `"\n"` mit `application/json` ergeben `unvollstaendig`. Mutationen: Typprüfung vor dem ersten Read: rot (Fall `text/plain` leer); Leerraum als leer: rot (Fall `" "`) |
| Typ: JSON-Typen, ein Wert, Media-Type bis `;`, Schreibweise, Rand-Leerraum, Parameter ohne Wirkung; alles andere `typ` ohne Ausgabe (Nr. 3) | `TestBodyTyp`: Body `{}` für alle Fälle. Gültig: `application/json`, `APPLICATION/JSON`, ` application/json `, `application/json; charset=utf-8`, `application/json;odata.metadata=minimal;odata.streaming=true`, `application/vnd.api+json`, `application/odata+json`. `typ`: `nil`, `[]`, `[""]`, `;charset=utf-8`, zwei Werte `application/json` und `application/json`, `application/json,application/json`, `text/json`, `text/plain`, `text/xml`, `multipart/mixed; boundary=b`, `application/xjson`, `application/jsonx`, `application/json-seq`, `application/+json`, `application/x-www-form-urlencoded`, `application/octet-stream`, `*/*`. Mutationen: Suffix `json` statt `+json` (`application/xjson`): rot; Media-Type nicht am `;` trennen: rot (Parameterfälle); nur den ersten Wert lesen: rot (zwei Werte); `strings.HasPrefix("application/json")` (`application/jsonx`, `-seq`): rot |
| EDMX: `application/xml` nur mit `Metadata`, byte-gleich, `none`, ohne Umschreiben; erstes Zeichen `<` nach BOM und Leerraum, sonst `mismatch`, das Byte wird nie geschrieben (Nr. 3) | `TestBodyEdmx`: Gutfälle `<?xml…><edmx:Edmx>`, mit BOM, mit Leerraum und BOM, mit `<password>S3CR3T-7f</password>` als Element (Ausgabe gleich Eingabe, kein Marker); `application/xml` mit `Metadata=false` und `text/xml` mit `Metadata=true`: `typ`; `mismatch`: Body `{"password":"S3CR3T-7f"}` (Geheimnis nicht in `dst`), `x<`, BOM allein, `"  "` allein, BOM-Rest `EF BB` allein, `EF BB 3C`; `application/json` mit `Metadata=true` und Body `{"password":"x"}`: `applied`. Mutationen: Sniff entfernt: rot (Body mit `{`); `Metadata` ignoriert: rot (`Metadata=false`); `Metadata=true` überschreibt JSON: rot (letzter Fall); BOM nicht erlaubt: rot (Gutfall mit BOM); nicht geprüftes erstes Byte vorab geschrieben: rot (`dst` des `{`-Falls muss leer sein) |
| Marker ersetzt den String-Wert, alle anderen Bytes gleich: Leerraum, Reihenfolge, Zahlenliterale, Escapes, Namen (Nr. 5) | `TestBodyMarker`: `{"password":"S3CR3T-7f"}`; `{ "password" :\t"x" ,\r\n"b":1.50E+2,"a":-0,"c":[1, 2.0],"d":"\u00e9\/"}` (genau diese Bytes in der Ausgabe außer dem einen Wert); leerer String `""`; Wert mit `\"` und `\\` vor dem schließenden Anführungszeichen (`"a\"b"`, `"a\\"`, `"\u0022"`), ein Wert von 1 Byte und von 10 000 Bytes. Mutationen: Ausgabe neu kodieren (Leerraum/Zahlen normalisieren): rot; Escape im Wert ignorieren (Wert endet am `\"`): rot (Rest bliebe stehen); Marker ohne Anführungszeichen: rot |
| Namen: ASCII-Schreibweise, Dekodierung aller Escapes vor dem Vergleich, Hex-Schreibweise gleich, Vergleich allein über `JSONPfad.Passt` (Nr. 5) | `TestBodyNamen`. Treffer: `PassWord`, `PASSWORD`, `pass\u0077ord`, `\u0050ASSWORD`, `api_\u006b\u0065y` und `api_\u004B\u0045Y` (Hex-Ziffern klein und groß, Muster `$..api_key`), `pass\u0077\u006fr\u0064`. Kein Treffer: `passwords`, `passwor`, `xpassword`, `password ` (Leerzeichen), `passwörd`, `api_` + U+212A + `ey` (Kelvin, roh in UTF-8), `api_\u212Aey`, `pass\\u0077ord` (Backslash-Paar, Name `pass\u0077ord` als Text), `""`. Mutationen: Rohvergleich ohne Dekodierung: rot (`pass\u0077ord`); `strings.EqualFold` oder `ToLower` statt `Passt`: rot (Kelvin); nur Hex-Kleinbuchstaben: rot (`\u004B`); Namen als Präfix vergleichen: rot (`passwords`) |
| Alle sieben Default-Muster und mehrere Muster in jeder Reihenfolge wirken (Nr. 5) | `TestBodyDefaults`: je Muster aus `SPEC-085` ein Body `{"<name>":"S3CR3T-7f"}` mit allen sieben `Pfade` in zwei Reihenfolgen; ein Body mit allen sieben Namen. Mutation: nur `Pfade[0]` (oder nur das letzte) prüfen: rot |
| Nur Mitgliedsnamen treffen: Array-Element, Wert, Wurzel nie; Tiefe beliebig, auch im Array; gleiche Namen je für sich (Nr. 5) | `TestBodyMitglieder`: `["password","x"]`, `{"a":"password"}`, `"password"` (Wurzel), `{"a":{"b":[{"c":{"password":"S3CR3T-7f"}}]}}`, `{"password":"S3CR3T-1","x":0,"password":"S3CR3T-2"}`, `{"password":"a","PASSWORD":null,"Password":"b"}` (null bleibt, beide Strings ersetzt). Mutationen: Strings in Arrays als Namen: rot (`["password","x"]` bliebe nicht `none`); nur das erste Vorkommen: rot; nur oberste Ebene: rot |
| Wertarten unter treffendem Namen: String ersetzt; `null`, `true`, `false` unverändert; Zahl, Objekt, Array `nichtstring`, ohne ein Byte des Wertes in `dst` (Nr. 5) | `TestBodyWertarten`: `{"password":null}`, `true`, `false` (Ausgabe gleich Eingabe, `none`); `{"password":12345}`, `{"password":-1.5e3}`, `{"password":{"v":"S3CR3T-7f"}}`, `{"password":["S3CR3T-7f"]}`, `{"secret":{"password":"S3CR3T-7f"}}` (Grund `nichtstring` am äußeren), `{"a":1,"password":0}`; `{"password":}` und `{"password":x}` ergeben `syntax`, nicht `nichtstring`. Mutationen: Zahl durchreichen: rot (`dst` enthält `12345`); Objekt durchreichen: rot (`S3CR3T-7f`); `null` als Fehler: rot; alle Nicht-Strings zu `syntax`: rot (Grund) |
| Marker ist keine Änderung; jeder andere Inhalt ist eine (Nr. 5; `SPEC-115` Nr. 4) | `TestBodyMarkerIdempotent`: `{"password":"[REDACTED]"}` ergibt `none` und gleiche Bytes; `applied` für `"\u005bREDACTED]"`, `"[REDACTED] "`, `"[REDACTED]x"`, `"[REDACTE"`, `"[REDACTED"`, `""`, `"x[REDACTED]"`; Ausgabe jeweils `"[REDACTED]"`. Mutationen: immer `applied`: rot (erster Fall); Präfixvergleich: rot (`[REDACTED]x` bliebe `none`); Vergleich nach Dekodierung: rot (`\u005b…` bliebe `none`) |
| Status `applied` genau bei geänderten Strings, sonst `none` (Nr. 5) | `TestBodyStatus`: jede Zeile von `bodyFaelle` mit dem von Hand gesetzten Status; ein Fall mit Treffer nur als `null`, einer ohne Muster, einer mit leeren `Pfade` und Geheimnis-Namen (`none`, Ausgabe gleich Eingabe). Mutation: `applied` bei Treffer unabhängig von der Änderung: rot (`null`-Fall) |
| JSON streng: jeder Syntaxfehler ist `syntax`, auch im Wert, der ersetzt wird (Nr. 4) | `TestBodySyntax`: je Fall einmal unter nicht treffendem und einmal unter treffendem Namen (`{"p":V}` und `{"password":V}`), Wert `V` aus: `x\y`, `\u12G4`, `\u00`, rohes LF im String, rohes Tab im String, `01`, `1.`, `.5`, `+1`, `1e`, `NaN`, `Infinity`, `True`, `nul`; dazu ganze Dokumente `{"a":1,}`, `[1,]`, `{,}`, `{"a" 1}`, `{a:1}`, `{'a':1}`, `{"a":1}{"b":2}`, `{"a":1} x`, `{}\n{}`, `"a" "b"`, `1 2`, BOM vor `{}`, NUL, `\f` oder NBSP als Leerraum, UTF-16-Bytes. Gut: `{} \n\t\r`, `0x7F` im String, `\u0000`, einzelnes Surrogat-Escape `\ud800` im Namen und im Wert. Mutationen: Prüfung im Überspringen weglassen: rot (Fälle unter `password`); Nachlauf ignorieren: rot; `\f` als Leerraum: rot |
| UTF-8: gültig an jeder Stelle, ungültig `utf8` in Namen, ersetzten und anderen Werten (Nr. 4) | `TestBodyUTF8`: gültig `é`, `€`, `😀` (2, 3, 4 Byte), `0x7F`, U+FFFF (`EF BF BF`), U+10FFFF (`F4 8F BF BF`) in Name, ersetztem und anderem Wert; ungültig `C0 80`, `C1 BF`, `E0 80 80`, `ED A0 80`, `F4 90 80 80`, `80`, `FF`, `E2 82` vor dem schließenden Anführungszeichen, je unter den drei Stellen, Grund `utf8` (nicht `syntax`). Mutationen: Prüfung weglassen: rot; Prüfung nur im nicht ersetzten Wert: rot (Fall unter `password`); Prüfung nur ohne Teilung (`utf8.Valid` je Chunk): rot (`TestBodyZerlegung`) |
| Tiefe 512 ist zulässig, 513 `tiefe`, vor `unvollstaendig` und `syntax` (Nr. 4) | `TestBodyTiefe`: 512 und 513 Arrays `[`…, 512 und 513 Objekte `{"a":`…, gemischt, 513 offene ohne Schluss (Grund `tiefe`, nicht `unvollstaendig`), 513 offene und danach `x` (`tiefe`). Mutationen: Grenze 511 und Grenze 513: beide rot (zwei Seiten) |
| Ende des Stroms vor dem Ende des Wertes ist `unvollstaendig`, jedes echte Präfix (Allaussage); ein am Ende gültiger Skalar ist vollständig (Nr. 4) | `TestBodyPraefixe`: jedes Präfix eines gültigen Dokuments aus `bodyFaelle` (Status `none` oder `applied`, Wurzel Objekt oder Array), das kürzer ist als bis zur schließenden Klammer der Wurzel (nachfolgender Leerraum gehört nicht dazu), ist `unsafe` mit Grund `unvollstaendig`; dazu Skalare `"abc`, `tru`, `-`, `1.`, `1e`, `{"a"`, `{"a":`, `[1,`: `unvollstaendig`; `12`, `"abc"`, `true`, `-0`, `1e5` als ganze Bodys: `none`. Mutationen: EOF bei Tiefe 0 akzeptieren: rot (Skalare); EOF nach vollständigem Mitglied akzeptieren: rot (Präfixe); Grund `syntax`: rot; `tru` als Wert akzeptieren: rot |
| Zerlegung ohne Wirkung: Name, Escape, `\u`-Ziffern, Mehrbyte-Folge in Name, ersetztem und anderem Wert, Zahl, Literal, Marker-Vergleich, Leerraum über Chunk-Grenzen; Daten zusammen mit EOF; leere Reads (Nr. 8) | `TestBodyZerlegung`: für jede Zeile von `bodyFaelle` (Gut- wie Fehlerfälle, höchstens 200 Byte, darunter je ein Fall der genannten Zustände): (a) jede Zweiteilung, (b) einzelne Bytes, (c) ein `src` mit leeren Reads zwischen allen Bytes, (d) ein `src`, das das letzte Byte zusammen mit `io.EOF` liefert; Ausgabe, Status, Grund gleich dem Ergebnis eines einzigen Reads. Mutationen: Namenpuffer am Chunk-Ende leeren: rot (Zweiteilung im Namen); Escape-Zustand verwerfen: rot (Teilung nach `\`); `\u`-Ziffernzähler verwerfen: rot; UTF-8-Rest verwerfen: rot; Marker-Vergleichsstand verwerfen: rot; `n > 0` mit `io.EOF` verwerfen: rot (d) |
| Grenze: genau `MaxBytes` zulässig, `MaxBytes + 1` `zu_gross`, höchstens `MaxBytes + 1` Bytes gelesen, gezählt wird die Eingabe, die Grenze geht vor einem Verstoß dahinter (Nr. 6) | `TestBodyGrenze` mit `MaxBytes = 16`: Dokument von 16 Bytes ohne Ersatz (`none`); Dokument von 16 Bytes mit leerem Geheimnis-String (Ausgabe 26 Bytes, `applied`); 17 Bytes gültig: `zu_gross`; 40 Bytes eines gültigen, unvollständigen Anfangs: `zu_gross` (nicht `unvollstaendig`); 16 Bytes unvollständig: `unvollstaendig`; Syntaxfehler an Byte 17: `zu_gross`, an Byte 16: `syntax`; ein zählender `src` meldet höchstens 17 gelesene Bytes bei 1 MiB Angebot; dazu `application/xml` mit `Metadata` an der Grenze. Mutationen: `>=` statt `>`: rot (16 Bytes); Ausgabe zählen: rot (Fall mit Verlängerung); ganzen Strom lesen: rot (Zähler) |
| Speicher unabhängig von Body, String, Name und Zahl; Namen bis zur Länge des längsten Musters treffen, ein Zeichen mehr nicht (Nr. 6) | `TestBodySpeicher` (`TotalAlloc` vor und nach, kein `t.Parallel`, `MaxBytes` 16 MiB, `dst` zählt nur, `src` erzeugt den Strom ohne ihn zu halten): 9 MiB String unter nicht treffendem Namen, 9 MiB String unter treffendem Namen (Marker), 9 MiB Ziffern einer Zahl, 9 MiB Name (kein Treffer), 9 MiB Leerraum, 9 MiB Array `[0,0,…]`; je Zuwachs unter 1 MiB. `TestBodyNamenLaenge`: Muster `$..client_secret` (13): `client_secret` trifft, `client_secre` und `client_secrett` nicht, ein Name von 100 000 Bytes nicht. Mutationen: Wert, Name oder Zahl puffern: rot (je eigener Fall, Zuwachs ab 9 MiB); Namenpuffer eine Stelle zu kurz: rot (`client_secret`); Namen auf Puffergröße kürzen und als Präfix vergleichen: rot (`client_secrett`) |
| Kein Byte eines geheimen Wertes erreicht `dst`, auch bei jedem `unsafe` und bei jeder Zerlegung; kein Fehlertext enthält ein Byte der Eingabe (Allaussage, Nr. 7, 9) | `TestKeinGeheimnisInAusgabe`: über alle Zeilen von `bodyFaelle` mit dem Merkmal Geheimnis, je Zweiteilung und einzelne Bytes: weder `dst` noch `err.Error()` enthält `S3CR3T-7f`; dasselbe mit einem `src`, das nach jedem Byte `k` mit einem Fehler endet (`eingabe`), und einem `dst`, das beim `k`-ten Write fehlschlägt (`ausgabe`); `TestBodyFaelleDecken`: `bodyFaelle` führt jeden der acht Dokumentgründe (`eingabe` und `ausgabe` entstehen nur durch `src` und `dst`), alle drei Status und jede Wertart (String, `null`, `true`, `false`, Zahl, Objekt, Array) mindestens einmal mit Geheimnis (sonst ist die Allaussage nur über einen Teil geprüft). Mutationen: Wert puffern und bei Fehler flushen: rot; Fehlertext mit Auszug der Eingabe: rot; Fall aus `bodyFaelle` entfernen: `TestBodyFaelleDecken` rot |
| Fehlerform: `*UnsicherFehler`, `Grund` je Fall, `errors.Is(err, ErrUnsicher)`, `Code()` ist `ODR-E2401`, Status `unsafe`, der erste Verstoß im Strom entscheidet (Nr. 7) | `TestBodyGrund`: jede Zeile von `bodyFaelle` prüft den Grund; je Dokumentgrund (`typ`, `mismatch`, `syntax`, `utf8`, `tiefe`, `nichtstring`, `zu_gross`, `unvollstaendig`) mindestens ein Fall, `eingabe` und `ausgabe` fangen `TestBodyEingabeFehler` und `TestBodyAusgabeFehler`. `TestBodyReihenfolge`: `{"password":1,}` ergibt `nichtstring`; `{"a":01,"password":1}` `syntax`; `{"b":"\xFF","password":1}` `utf8`; 513 `[` und `x` `tiefe`; `text/plain` mit `MaxBytes=0` `ErrUngueltig`; leer mit `text/plain` `none`. Mutationen: Gründe vertauschen (`utf8` als `syntax`): rot; weiterlesen nach dem Verstoß (zweiter Verstoß überschreibt): rot (`{"a":01,"password":1}`) |
| Fehler von `src`: `eingabe`, Ursache erreichbar, auch nach vollständigem Body; Daten mit Fehler zuerst verarbeitet (Nr. 7) | `TestBodyEingabeFehler`: ein `src`, das nach jedem Byte `k` von 0 bis `len` mit `errTest` oder `context.Canceled` endet (also auch nach dem letzten Byte eines vollständigen Dokuments): `unsafe`, Grund `eingabe`, `errors.Is`; ein Read, der Daten und Fehler zugleich liefert; Syntaxfehler in den Daten, die mit dem Fehler kommen: `syntax`. Mutationen: Fehler nach vollständigem Dokument ignorieren: rot (`k = len`); Daten bei Fehler verwerfen: rot (Syntaxfall); Ursache nicht einpacken: rot |
| Fehler von `dst`: `ausgabe`, Ursache erreichbar, Kurzschreiben ist Fehler (Nr. 7) | `TestBodyAusgabeFehler`: `dst` schlägt beim `k`-ten Write fehl (`outbound.ErrZuGross` und `errTest`), schreibt kürzer ohne Fehler; nach dem Fehler keine Reads und keine Writes mehr (Zähler). Mutationen: Kurzschreiben ignorieren: rot; weiterlesen nach dem Fehler: rot |
| Zustandslos zwischen Aufrufen; gleiche Eingabe gleiches Ergebnis; Anwendung auf die eigene Ausgabe ist die Identität mit `none` (Nr. 8) | `TestBodyOhneZustand`: Aufruf A bricht mitten in einem geheimen Wert und mitten in einem Namen ab, danach Aufruf B auf denselben `BodyRegeln`: Ausgabe gleich der auf frischem Wert und gleich B vor A. `TestBodyIdempotent`: Ausgabe jeder gültigen Zeile erneut durch die Funktion: gleiche Bytes, `none`. Mutation: Puffer oder Zähler auf Paketebene: rot (`TestBodyOhneZustand`) |
| Abbruch mitten im Dokument und Redaction-Fehler verwerfen den Body (SPEC-053-Fixtures) (Nr. 7, 9) | `TestBodyFixturesSPEC053`: Token über Chunk-Grenzen (`TestBodyZerlegung`), MIME-Mismatch (`TestBodyEdmx`, `TestBodySyntax` mit XML unter `application/json`), Abbruch (`TestBodyEingabeFehler`), Redaction-Fehler (`TestBodyWertarten`) als benannte Tabelle: je Fixture der Name des Falls in `bodyFaelle`; der Test prüft, dass jeder genannte Fall existiert (sonst ist das Fixture nur behauptet). Mutation: Fall aus `bodyFaelle` entfernen: rot |

Nicht gefangen und als akzeptiertes Negativ benannt: dass die Aufrufer ihre Pflicht aus `SPEC-118` Nr. 10 tragen (jeder Body durch die Funktion, Key aus der Ausgabe, `dst` bei Fehler verwerfen, `Metadata` nur für die Antwort auf `$metadata`, Strom entpackt): ein Test der Funktion sieht ihre Aufrufer nicht, die Dienste und die Pipeline tragen die Zusage in ihrem §1 und fangen sie mit ihren Tests (einmaliger, harmloser Blindfleck dieses Slice, kein eigenes Register); dass `SPEC-118` und die Tabelle `bodyFaelle` dieselbe Lesart haben, weil beide aus einer Quelle stammen (der Reviewer liest Nr. 4 und 5 gegen die Tabelle); Datenfreiheit paralleler Aufrufe (die Funktion hält keinen gemeinsamen Zustand, `TestBodyOhneZustand` fängt jeden Zustand, der zwischen Aufrufen bliebe; ein Rennen innerhalb gleichzeitiger Aufrufe sieht das Gate nicht, es fährt nicht mit dem Race-Detektor, und der Lauf ist nicht nebenläufig in `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` gezählt); die Laufzeit und Unicode-Gleichheit von Namen über die ASCII-Schreibweise hinaus (`SPEC-118` *Ungeprüft*).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-redaction-match-key` liegt in `done/`. Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-118`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12); sie ist mit diesem Plan geschrieben.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 2200 Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen; `<Basis>` ist der Commit, der diesen Plan und `SPEC-118` einführt, ihr Text zählt also nicht mit); die Schätzung liegt bei rund 1500 in derselben Metrik (Produktcode rund 450, Tests rund 1000, Nachzug der Spezifikation rund 50; der Vorgänger lag mit 2400 geschätzt bei 1957, davor mit 600 geschätzt bei 2369, deshalb ist die Marke weit über der Schätzung gesetzt).
- `in-progress` → `open` (blockiert): der Implementer findet eine Randform ohne Ort in `SPEC-118` oder im Raster in §6 und muss sie an den Architect geben (`AGENTS.md` §3.12).

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

Der Vertrag ist entschieden in `SPEC-118`; die Nummern unten sind die dieser Stelle (und `SPEC-115` Nr. 6 für die Pfadsprache). Eine Zelle, die der Implementer ohne Ort findet, geht an den Architect (`AGENTS.md` §3.12), sie wird nicht im Code entschieden.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16). Es gibt eine Operation, `RedigiereBody`; die Zeilen sind ihre Argumente und Eingabeformen. *Fehlend* heißt `nil` oder nicht angegeben, *leer* ein vorhandener leerer Wert, *unbekannt* ein Wert außerhalb der Menge, *fremd* eine Form, die der Vertrag nicht vorsieht. **(a)** Abbruch eines Arguments: gilt nicht für Werte, die vor dem Lesen entschieden werden, weil die Funktion dafür nichts liest. **(b)** Wiederholung: gleiche Eingabe, gleiches Ergebnis, die Funktion ist zustandslos (`SPEC-118` Nr. 8); wo die Spalte etwas anderes sagt, steht es dort.

| Argument | fehlend | leer | unbekannt | fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| `ContentType` | `nil`, `[]`: `typ` (Nr. 3) | `[""]`, `;charset=utf-8`: `typ` (Nr. 3); bei leerem Strom `none` (Nr. 2) | `text/plain`, `text/xml`, `multipart/mixed`, `application/json-seq`: `typ` (Nr. 3) | zwei Werte, Kommaliste, `application/+json`, Großschreibung, Rand-Leerraum, Parameter (Nr. 3) | Fehler des ersten Reads: `eingabe` (Nr. 7) | (b) |
| `Metadata` | `false` ist der Nullwert, wie jeder Aufruf ohne `$metadata` (Nr. 3) | gilt nicht: Wahrheitswert | gilt nicht: Wahrheitswert | `true` mit JSON-Typ, `true` mit `text/xml`, `false` mit `application/xml` (Nr. 3) | (a) | (b) |
| `MaxBytes` | `0`: `ErrUngueltig` (Nr. 1) | gilt nicht: Zahl; `0` gleich fehlend | negativ: `ErrUngueltig` (Nr. 1) | genau `MaxBytes`, `MaxBytes + 1`, Verstoß dahinter (Nr. 6) | Fehler von `src` vor der Grenze: `eingabe` (Nr. 7) | (b) |
| `Pfade` | `nil`, leer: gültig, nichts ersetzt, Prüfung bleibt (Nr. 1) | Nullwert `JSONPfad{}`: `ErrUngueltig` (Nr. 1) | Name, der in keinem Mitglied vorkommt: `none` (Nr. 5) | zwei gleiche Muster, mehrere Muster, Muster mit Großbuchstaben (Nr. 5; `SPEC-115` Nr. 6) | (a) | (b) |
| `src`, `dst` | `nil`: `ErrUngueltig` (Nr. 1) | `src` ohne Byte: `none` (Nr. 2); `src` nur leere Reads: dasselbe (Nr. 7) | gilt nicht: Schnittstellen | `src` mit Daten und `io.EOF` zugleich, mit Daten und Fehler zugleich, `dst` schreibt kurz (Nr. 7) | Fehler von `src` oder `dst` an jeder Stelle: `eingabe`, `ausgabe` (Nr. 7) | nach einem Fehler ein weiterer Aufruf, unbeeinflusst (Nr. 8) |
| Gesamter Strom | `src` leer: Nr. 2 | nur Leerraum: `unvollstaendig` (Nr. 4) | kein JSON (XML, Text, UTF-16, BOM): `syntax` (Nr. 3, 4) | zwei Werte, Nachlauf, ein Skalar als Wurzel (gültig), Wurzel Array (Nr. 4) | Ende des Stroms in jedem Zustand: `unvollstaendig` (Nr. 4) | Ausgabe als Eingabe: gleiche Bytes, `none` (Nr. 8) |
| String (Name, Wert) | gilt nicht: ein String ist da oder die Grammatik greift (Nr. 4) | `""` als Name (kein Treffer, Nr. 5), als Wert (ersetzt, Nr. 5) | ungültiger Escape, ungültiges UTF-8: `syntax`, `utf8` (Nr. 4) | roh `0x7F` (gültig), roh unter `0x20` (`syntax`), einzelnes Surrogat-Escape (gültig), `\u0000` (gültig) (Nr. 4) | offener String am Ende: `unvollstaendig` (Nr. 4) | (b) |
| Mitgliedsname | `""`: kein Treffer (Nr. 5) | gleicher Fall | Name außerhalb von `Pfade`, Name mit `@odata.type`-Zusatz: kein Treffer (Nr. 5) | Escape-Schreibweise, Schreibweise, Kelvin-Zeichen, Länge über dem längsten Muster, genau so lang, ein Zeichen länger (Nr. 5, 6) | Ende im Namen: `unvollstaendig` (Nr. 4) | doppelter Name: beide (Nr. 5) |
| Wert unter Treffer | fehlt (`{"password":}`): `syntax` (Nr. 4) | `""`: ersetzt (Nr. 5) | `null`, `true`, `false`: unverändert (Nr. 5) | Zahl, Objekt, Array: `nichtstring`; Marker schon da: keine Änderung (Nr. 5) | Ende nach dem Doppelpunkt: `unvollstaendig` (Nr. 4) | Ausgabe erneut: `none` (Nr. 5, 8) |
| EDMX-Strom | leer: `none` (Nr. 2) | nur BOM oder Leerraum: `mismatch` (Nr. 3) | JSON oder Text als `application/xml`: `mismatch` (Nr. 3) | BOM, Leerraum, dann `<`; UTF-16: `mismatch` (Nr. 3) | Fehler von `src`: `eingabe` (Nr. 7) | Ausgabe gleich Eingabe (Nr. 3) |

**Zusatzzeilen für Kombinationen zweier Merkmale, Darstellung und Kodierung, Nebenläufigkeit und Modi** (`AGENTS.md` §3.16):

| Fall | Ort |
|---|---|
| Treffer × Duplikat (`password` zweimal, einmal String, einmal Zahl) | Nr. 5: je Mitglied für sich, der erste Verstoß im Strom entscheidet (Nr. 7) |
| Treffer × Treffer (`secret` mit `password` darin) | Nr. 5: `nichtstring` am äußeren |
| Treffer × Escape im Namen × Chunk-Grenze mitten im Escape | Nr. 5 und 8 |
| Treffer × ungültiges UTF-8 im ersetzten Wert | Nr. 4: `utf8`, der Wert wird geprüft, bevor ein Byte von ihm zählt |
| Treffer × Marker × Escape (`\u005bREDACTED]`) | Nr. 5: Rohtext zählt, also `applied` |
| Treffer × `MaxBytes` (Ausgabe länger als Eingabe) | Nr. 6: gezählt wird die Eingabe |
| Treffer × Tiefe 512 und 513 | Nr. 4: `tiefe` zählt Klammern, nicht Treffer |
| Typ `application/json` × `Metadata=true` | Nr. 3: Typ JSON bleibt JSON |
| `application/xml` × JSON-Inhalt | Nr. 3: `mismatch` |
| `application/json` × XML-Inhalt | Nr. 3 und 4: `syntax` |
| Leerer Strom × unbekannter Typ × `MaxBytes=0` | Nr. 1 vor Nr. 2 vor Nr. 3: `ErrUngueltig` vor `none` vor `typ` |
| Verstoß vor und Verstoß nach der Grenze | Nr. 6: hinter `MaxBytes + 1` wird nichts gesehen; davor entscheidet der erste Verstoß |
| Fehler von `src` × Syntaxfehler in denselben Daten | Nr. 7: Daten zuerst verarbeitet, `syntax` |
| Fehler von `dst` × Fehler von `src` | Nr. 7: der erste, den der Ablauf trifft; danach kein Zugriff mehr |
| Mehrbyte-Folge, `\u`-Escape, Surrogat-Paar, Zahl, Literal über eine Chunk-Grenze | Nr. 8: `TestBodyZerlegung` |
| Surrogat-Paar als Escape im Namen (`\ud83d\ude00`), einzelnes Surrogat-Escape, Hex-Schreibweise | Nr. 4 und 5: gültig, trifft nie, Hex gleich |
| Zahl mit 400 Ziffern, `1E5`, `-0`, `1.50E+2` | Nr. 4: durchgereicht unverändert; unter treffendem Namen `nichtstring` (Nr. 5) |
| `Content-Encoding`, Transfer-Codierung | gilt nicht: der Strom ist entpackt (Nr. 10); Ort: `SPEC-046` Nr. 2 und `slice-capture-pipeline` |
| Abbruch des Clients, Zeitüberschreitung, Herunterfahren | Nr. 7: Fehler von `src` ist `eingabe` (`context.Canceled` und `DeadlineExceeded` bleiben erreichbar); die Funktion hat weder Kontext noch Uhr |
| Zwei Aufrufe gleichzeitig, Aufruf nach einem abgebrochenen | Kopf von `SPEC-118`: kein Zustand zwischen Aufrufen; Aufruf nach Abbruch: `TestBodyOhneZustand`; Rennen: *Ungeprüft* |
| Request-Body gegen Response-Body, Record gegen Replay | Nr. 10: dieselbe Funktion; `Metadata` nie für einen Request |
| Request-Body mit Fehler im Replay | Nr. 10: `ODR-E5001` ohne Cursor-Fortschritt (Träger `slice-replay-service`) |
| Request-Body mit Fehler im Record | `SPEC-001`: `redaction.request=unsafe`, `null`-Key, `failed` (Träger `slice-record-service`, `slice-capture-pipeline`) |
| Bytes des Bodys erreichen den Key unredigiert | Nr. 10 und `SPEC-116` Einleitung: Pflicht der Aufrufer; kein Fänger in diesem Slice (akzeptiertes Negativ in §3) |

**Risiken:**

- Die Pfadsprache war nicht festgelegt (`$..name` oder auch `$.a.b`, Indizes, Platzhalter) — **Ausgang:** entfallen; entschieden in `SPEC-115` Nr. 6 (nur `$..name`, der Start scheitert bei jeder anderen Form mit `ODR-E2001`), der Redaktor sieht nie eine andere Form.
- Nicht-String-Werte unter einem treffenden Namen (Zahl, `null`, Wahrheitswert, Objekt, Array) — **Ausgang:** entfallen; entschieden in `SPEC-118` Nr. 5: `null`, `true`, `false` bleiben, Zahl, Objekt, Array sind `nichtstring` (fail-closed nach `LH-FA-14`, Negativfall; der Marker ersetzt nur Strings). Folge: ein Austausch, dessen Body etwa `"password": 12345` führt, wird als `failed` mit verworfenem Body aufgezeichnet. Die Umkehr (Zahl zu Marker-String) ändert `SPEC-118` Nr. 5 und die Zeile *Wertarten*, nicht den Rest.
- Gleiche Schlüssel mehrfach, Unicode-Escapes im Schlüssel (`\u0070assword`), große Zahlen — **Ausgang:** entfallen; entschieden in `SPEC-118` Nr. 4 und 5 (je Mitglied für sich, Dekodierung vor dem Vergleich, Zahlen unverändert).
- Speicher: Ein Wert über der Grenze (ein String von 8 MiB) wäre im Entwurf ein Fail-closed-Fall gewesen, weil der Redaktor ein Token puffern sollte — **Ausgang:** entfallen; entschieden in `SPEC-118` Nr. 6: Strings, Zahlen und Namen über dem längsten Muster werden nicht gepuffert, der Speicher hängt nur an Tiefe, längstem Muster und festen Puffern, und die Grenze `MaxBytes` zählt die Eingabe. Fänger: `TestBodySpeicher`.
- Der Speichertest über `TotalAlloc` ist nur ohne parallele Tests belastbar und hängt an der Zusage „höchstens 64 KiB je Richtung“ — **Ausgang:** offen; der Test läuft ohne `t.Parallel` mit der Schwelle 1 MiB bei 9 MiB Strom; ein Wechsel auf eine andere Messung ist eine Änderung der Zeile in §3.
- Das Raster *Operation × Argumentart* (`AGENTS.md` §3.12, §3.16): es steht oben vor dem Code; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.
- Übergaben an `slice-capture-pipeline`, `slice-record-service`, `slice-replay-service` (§1) — **Ausgang:** weiter offen bis zur Closure der Nehmer, die die Sendung in ihrem §1 führen.
- Größe: die Marke in §4 ist weit über der Schätzung gesetzt, weil zwei Vorgänger die Schätzung überschritten — **Ausgang:** offen; gemessen bei der Verifikation.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (3× ohne Ausgang, der Redaktor ist zustandslos und ohne Nebenläufigkeit, der Speichertest läuft deshalb ohne `t.Parallel`) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (die Allaussagen „kein Byte eines Geheimnisses in `dst`“ und „jedes Präfix ist `unvollstaendig`“ tragen Fänger über alle Fälle der Tabelle `bodyFaelle`, `TestBodyFaelleDecken` sichert, dass die Tabelle jede Klasse führt); Sicherheits-Fixtures und Randformen des Redaktors stehen in der Tabelle *Zusage · Fänger* vor dem Code, ebenso die zwei Muster, die das Review des Vorgängers bemängelte: Kombinationen zweier Merkmale und Darstellung oder Kodierung (Zusatzzeilen in §6).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
