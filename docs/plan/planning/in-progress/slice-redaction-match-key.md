# Slice slice-redaction-match-key: Redaction von Headern und Query, Match-Key, Replayfähigkeit

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-05`, `LH-FA-14`, `LH-QA-02`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-05.a` (Match-Key) · `LH-FA-14.a` (Redaction) · `SPEC-115`, `SPEC-116`, `SPEC-117` (der Vertrag dieses Slice, vor dem Code geschrieben) · `SPEC-045` (Fähigkeitsmatrix) · `SPEC-060`, `SPEC-061`, `SPEC-063`, `SPEC-081` bis `SPEC-085` · `SPEC-110` Nr. 6 (Grammatik der Pfadmuster) · `SPEC-015`, `SPEC-067` · `SPEC-001` (`key_redacted`) · `ARC-001`.
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

**Ziel:** Das Domain-Modell redigiert Header, `raw_query` und Query-Werte in `Location` und `Content-Location` mit dem festen Marker (`SPEC-115`), bildet den Match-Key als SHA-256 über die eindeutig kodierte Tupelfolge aus der redigierten Darstellung in einem einzigen Aufruf, der die persistierbare Darstellung mitliefert (`SPEC-116`), und entscheidet nach `SPEC-045`, ob ein Exchange replayfähig ist (`SPEC-117`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Redaction der JSON-Bodies im Strom — `slice-streaming-redaction`; hier entsteht die Syntax der Pfadmuster (`$..name`) und der Marker, nicht der Strom.
- Anwendung beim Record und beim Replay — `slice-record-service` und `slice-replay-service` rufen dieselbe Funktion; hier sind es reine Funktionen ohne I/O.
- Semantische Gleichsetzung von `$filter`-Ausdrücken — `LH-FA-05.a` nimmt sie aus dem Muss-Umfang.
- Das Weglassen redigierter Header und die Kennzeichnung `X-Recorder-Match-Redacted` bei der Auslieferung — `slice-replay-service`; hier entsteht `OhneRedigierte` und `KeyRedacted`, nicht die Antwort.
- Die Reaktion auf `ErrKodierung` (HTTP 400, `ODR-E5006`) und auf einen nicht bildbaren Key — `slice-record-service` und `slice-replay-service`; `SPEC-116` Nr. 9 legt sie fest, die Dienste setzen sie um.

**Übergaben an Folge-Slices** (jeder Nehmer führt die Sendung in seinem §1 mit der Kennung dieses Gebers): `slice-streaming-redaction` (Grammatik und Bedeutung der Pfadmuster, Marker), `slice-record-service` und `slice-replay-service` (ein Aufruf `BildeSchluessel`, `ErrKodierung`, `OhneRedigierte`, `Replayfaehig`), `slice-baseline-import` (derselbe Aufruf für die Auflösung von `@odata.nextLink`).

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

- [ ] Redaction von Headern (Default-Liste plus `REDACT_HEADERS`), `raw_query` und Query-Werten in `Location`/`Content-Location` (`REDACT_QUERY_PARAMS`); fester Marker `[REDACTED]`; Parser der Pfadmuster aus `REDACT_JSON_PATHS` (auch vom Start gerufen); `key_redacted`.
- [ ] Match-Key: Tupelkodierung ohne Mehrdeutigkeit, `normalized_path`, sortierte Multimenge der dekodierten Query-Paare (`+` bleibt `+`, ungültige Kodierung `literal` oder `reject`), Vary-Header, Body-Hash in den Modi `bytes` und `json-canonical`, volatile und redigierte Header ausgenommen.
- [ ] Replayfähigkeit nach `SPEC-045` als Funktion über Capture-Status, Body-Redaction und Integrität (`REPLAY_REDACTED_BODIES`).
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
| `src/hexagon/model/` | update | Reine Funktionen ohne Uhr und Kennung (kein `Clock`, kein `IDGenerator`, kein `supporttest` nötig), in `redaction.go`, `schluessel.go`, `schluessel_json.go` und `replay.go`: `RedigiereHeader(Header, Namen, Parameter)`, `RedigiereQuery(RawQuery, Parameter)`, `OhneRedigierte`, `ParseJSONPfad` mit `JSONPfad.Passt` und die Konstante `Marker` (`SPEC-115`), `BildeSchluessel(SchluesselEingabe, SchluesselRegeln)` mit dem Ergebnis `Schluessel` (`SPEC-116`), `Replayfaehig(Exchange, RedigierteBodies)` (`SPEC-117`), `ErrKodierung` |
| `src/bootstrap/` | update | `istJSONPfad` ruft `ParseJSONPfad` des Modells statt einer zweiten Grammatik (`SPEC-110` Nr. 6 und `SPEC-115` Nr. 6 sind dieselbe Aussage); sonst bleibt die Konfiguration unberührt |
| `spec/spezifikation.md` | update | **vor dem Code geschrieben:** `SPEC-115`, `SPEC-116`, `SPEC-117` samt Zeigern in `LH-FA-05.a` und `LH-FA-14.a` |
| Tests des Pakets `model` | neu | Tabellentests mit Mutationen je Zusage nach `AGENTS.md` §3.10, darunter die Beispiele aus `LH-FA-05.a` und `LH-FA-14.a` |

Der Vertrag steht in `SPEC-115` bis `SPEC-117` **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) steht unten, das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16). Die Fänger sind Namen geplanter Tests; der Implementer legt sie unter diesem Namen an oder nennt den Namen, den er stattdessen vergibt, im Bericht.

**Zusage · Fänger** (jede Zeile löst der Reviewer mit einer Mutation auf; Nummern sind die der genannten SPEC):

| Zusage | Fänger |
|---|---|
| Marker `[REDACTED]` ist der einzige Platzhalter (`SPEC-115` Nr. 1) | `TestMarker`: Konstante gegen den Text; jede Redaction der Tests vergleicht gegen diese Zeichenfolge |
| Redigierter Header trägt genau einen Wert, den Marker, bei keinem, leerem, einem, mehreren Werten (Nr. 2) | `TestRedigiereHeaderWerte` |
| Name-Abgleich ohne Beachtung der ASCII-Schreibweise gegen die ganze Menge (Default `Authorization`, `Proxy-Authorization`, `Cookie`, `Set-Cookie`, `X-API-Key`, `X-Auth-Token`, `X-CSRF-Token` samt Ergänzung); ähnliche Namen (`Authorisation`) nicht (Nr. 2) | `TestRedigiereHeaderNamen`. Mutation: Abgleich nur auf kanonischer Schreibweise: rot |
| Nicht redigierte Header bleiben mit Werten, Reihenfolge, leerer Liste, leerem Wert (Nr. 2) | `TestRedigiereHeaderUnveraendert` |
| Schreibweisen desselben Namens fallen zusammen, Namen kanonisch, Nicht-Token-Namen unverändert (Nr. 2) | `TestRedigiereHeaderSchreibweise`: `cookie` und `Cookie` ergeben einen Eintrag mit einem Marker; zwei nicht redigierte Schreibweisen ergeben die Werte in der Bytefolge der Namen |
| Nicht-Token-Namen werden nie zusammengeführt, jeder für sich verglichen (Nr. 2) | `TestRedigiereHeaderNichtToken`: `bad name` und `BAD NAME` bleiben zwei Einträge mit je ihren Werten; steht `bad name` in `Namen`, sind beide redigiert (Vergleich ohne Beachtung der Schreibweise), aber weiter zwei Einträge mit je einem Marker; der leere Name bleibt ein eigener Eintrag. Mutation: Nicht-Token-Namen auf die kleine Schreibweise zusammenführen: rot |
| Ergebnis nie `nil`, Eingabe bleibt unverändert (Einleitung, Nr. 2) | `TestRedigiereHeaderKopie`: `nil`, leere Map; Eingabe vorher und nachher tief verglichen |
| `Location` und `Content-Location`: Query ersetzt, Rest und Fragment bleiben, ohne `?` unverändert, mehrere Werte, ganz redigierter Header gewinnt, in Request und Response gleich; ein `#` vor dem ersten `?` ändert nichts (`/p#f?token=1` ergibt `/p#f?token=[REDACTED]`) (Nr. 3) | `TestRedigiereLocation`. Mutation: Query erst nach dem ersten `#` suchen: rot |
| Query-Zerlegung und -Ersetzung: nur `&`, leere Segmente bleiben, Segment ohne `=` unverändert, erster `=` trennt, Treffer auch bei leerem Wert, jedes Vorkommen, Schlüsseltext bleibt, ASCII-Schreibweise, `%`-dekodierter Name trifft, ungültig kodierter Name trifft nicht, `;` trennt nichts, führendes `?` ist Teil des Schlüssels (Nr. 4) | `TestRedigiereQueryRandformen`: Tabelle mit `""`, `?`, `a`, `a=`, `=v`, `&&`, `token=x&token=y`, `TOKEN=x`, `tok%65n=x`, `to%ken=x`, `a=1;token=x`, `token`, `?token=x`, `token=%zz`, `token=` |
| Meldung „Wert änderte sich“, auch beim zweiten Lauf: nein; Marker als Wert ist keine Änderung (Nr. 4) | `TestRedigiereQueryGeaendert`. Mutation: immer `true` melden: rot |
| Wiederholung auf das eigene Ergebnis ändert nichts (Nr. 4) | `TestRedigiereIdempotent` über alle Fälle der beiden Tabellen oben |
| Die Redaction scheitert nie und hängt nicht von `MATCH_INVALID_ENCODING` ab (Nr. 4) | `TestRedigiereQueryUngueltigeKodierung`: `%zz`, `%`, Roh-Byte `0xFF`; die Signatur trägt keinen Fehler (kein eigener Testfall für „kein Fehler“) |
| Allaussage: kein Geheimnis steht im Ergebnis, auch nicht in einem Fehlertext (Einleitung `SPEC-115`, `SPEC-116` Nr. 9) | `TestKeinGeheimnisImErgebnis`: läuft über **alle** Fälle der Tabellen in `TestRedigiere*` und `TestKey*` mit dem Geheimnis `S3CR3T-7f` als Wert; Header, `raw_query`, `Location` und jeder Fehlertext dürfen es nicht enthalten. Mutation: Redaction lässt einen Fall aus (etwa den zweiten von zwei Parametern): rot |
| Ausgelieferte Header: Header, dessen Werte alle der Marker sind, entfällt; leere Liste, leerer Wert, Marker als Teil eines Wertes, Marker neben anderem Wert bleiben (Nr. 5) | `TestOhneRedigierte` |
| Pfadmuster: genau `$..name`, ASCII, ohne Leerraum (Nr. 6) | `TestParseJSONPfad`: gültig `$..password`, `$..api_key`, `$..A-b_1`; ungültig `""`, `$`, `$..`, `$.a`, `$...a`, `..a`, `$..a.b`, `$..a b`, ` $..a`, `$..a `, `$..*`, `$..[0]`, `$..a[0]`, `$..ä`, `$..a,b` |
| Bedeutung des Musters: Name ohne Beachtung der ASCII-Schreibweise (Nr. 6) | `TestJSONPfadPasst` |
| Der Start prüft `REDACT_JSON_PATHS` mit demselben Parser (Nr. 6) | `TestKonfigJSONPfadWieModell` (Paket `bootstrap`): die Tabelle von `TestParseJSONPfad` über `LadeKonfiguration`, gleiche Gültigkeit. Mutation: `istJSONPfad` zurück auf eine eigene Grammatik mit Abweichung: rot |
| Beispiele aus `LH-FA-05.a`: andere Parameter-Reihenfolge ergibt dasselbe, anderer `$filter` ergibt einen anderen Key (`SPEC-116` Nr. 10) | `TestKeyBeispieleLH` |
| Kodierung ist eindeutig: verschobene Feldgrenzen ergeben verschiedene Keys (Nr. 1) | `TestKeyEindeutig`: `Method`/`ServiceID` (`A`,`BC` gegen `AB`,`C`), Pfad gegen Paar, Wertgrenzen der Vary-Header. Mutation: Längenpräfix weglassen: rot |
| Kodierung gemäß Nr. 1: Marke, Reihenfolge der Felder, Art-Bytes, `sha256:` und 64 Hex-Ziffern klein | `TestKeyKodierungHandgebaut`: Bytefolge im Test von Hand aus Nr. 1 zusammengesetzt, SHA-256 und Präfix verglichen (für einen Fall je Art von Paar und Body, dazu `\u001F` als Eingabe und `\u001f` in der Kanonform, `\n` als `\u000a`, U+007F unverändert; Marke und Text der Art 2 mit Längenpräfix); das ist der Fänger der Format-Zusage (`SPEC-015`) |
| `Method` und `ServiceID`: Schreibweise zählt, leer ist `ErrUngueltig` (Nr. 2) | `TestKeyKopf` |
| Pfad: Escape ohne UTF-8 (`%FF`, `%ff`) ist gültig, bleibt als `%FF`, auch bei `reject` (Nr. 3) | `TestKeyPfad` (Fälle `%ff` gleich `%FF`, kein `ErrKodierung`) |
| Pfad: erlaubte Gleichheiten (`%7e`/`~`, `%2f`/`%2F`, `%41`/`A`), erhaltene Unterschiede (`%2F`/`/`, `%27`/`'`, `%20`/`+`, `/a`/`/a/`, `//a`, `/A`, `/a/./b`, `é`/`%C3%A9`), ungültige Formen (`""`, `a`, `/a?b`, `/a#b`, Steuerzeichen), ungültiges `%` bei `literal` und `reject` (Nr. 3) | `TestKeyPfad` |
| Query-Paare: Reihenfolge ohne Wirkung, Wiederholung bleibt (Multimenge), `a=1&a=2` gleich `a=2&a=1`, `a=1&a=1` verschieden von `a=1` (Nr. 4) | `TestKeyQueryReihenfolge`. Mutation: Paare nicht sortieren oder Doppelte entfernen: rot |
| Zerlegung: `a` verschieden von `a=`, `a=&` gleich `a=`, `&&`/`?`/`""` ohne Paar, `=v`, `=`, `a=b=c`, `;` trennt nicht (Nr. 4) | `TestKeyQueryZerlegung` |
| Paar der Art 2: Text byteweise, auch in der Schreibweise der Hex-Ziffern (`a=%2f%zz` gegen `a=%2F%zz`) (Nr. 4, 10) | `TestKeyUngueltigeKodierung`. Mutation: Hex-Ziffern im Text der Art 2 normalisieren: rot |
| Dekodierung: `+` bleibt `+` (`a+b` ≠ `a%20b`), Hex-Schreibweise gleich, `%2B` gleich `+` (beide ergeben das Zeichen `+`), `%41` gleich `A` (Nr. 4) | `TestKeyQueryDekodierung` |
| Ungültige Kodierung: `literal` hält byteidentische Segmente gleich und `a=%zz` verschieden von `a=%25zz`; `reject` liefert `ErrKodierung`, nicht `ErrUngueltig`; kein Wert im Fehlertext (Nr. 4, 9) | `TestKeyUngueltigeKodierung`: `%zz`, `%`, `%2`, `%FF`, Roh-Byte `0xFF` in Schlüssel und in Wert |
| Die Prüfung läuft über die redigierte Darstellung: `token=%zz` ist bei `reject` gültig (Nr. 4) | `TestKeyRedigiertVorKodierung` |
| Vary-Header: fehlend (kein Eintrag, leere Liste) verschieden von leerem Wert; Wert, Zeilen gegen Kommaliste, Schreibweise des Namens, Rand-Leerraum, Reihenfolge der Map; nicht gelistete Header ohne Wirkung; `none` ergibt Anzahl 0 (Nr. 5) | `TestKeyVary` |
| Volatile und redigierte Namen gehen nie ein, auch wenn `MATCH_VARY_HEADERS` sie nennt (Nr. 5) | `TestKeyVaryAusnahme`: `Date` und `Authorization` in der Menge, abweichende Werte ergeben denselben Key. Mutation: Ausnahme nur für `VOLATILE_HEADERS`: rot |
| Body: leer und `nil` sind gleich; `bytes` unterscheidet jedes Byte (Nr. 6) | `TestKeyBodyBytes` |
| `json-canonical`: Schlüsselreihenfolge und Leerraum gleich, `"a"` gleich `"a"`; verschieden: Arrays, Zahlenliteral (`1`, `1.0`, `1e0`, `-0`), doppelte Schlüssel in anderer Reihenfolge (Nr. 6) | `TestKeyBodyKanonisch` |
| Rückfall: ungültiges JSON, BOM, Nachlauf, einzelnes Surrogat, ungültiges UTF-8, Tiefe 513 (512 noch kanonisch) ergeben Art 2 (Byte-Hash), kein Fehler; die Modi ergeben für denselben Body verschiedene Keys (Nr. 6) | `TestKeyBodyRueckfall` |
| `KeyRedacted`: `applied` oder geänderte Query ergibt `true`; Header-Redaction allein, Marker als Wert und `none` ergeben `false`; `unsafe`, leer, unbekannt ergeben `ErrUngueltig` ohne Key (Nr. 7) | `TestKeyRedacted` |
| Ergebnis trägt die redigierte Query und die redigierten Header; Pfad und Body bleiben; Key des redigierten Requests ist der des unredigierten, ein anderes Geheimnis ergibt denselben Key (Nr. 8) | `TestKeyIdempotent`, `TestKeyGeheimnisUnsichtbar` |
| Aufzählungen der Regeln: `MATCH_INVALID_ENCODING` und `MATCH_BODY_MODE` leer oder unbekannt ergeben `ErrUngueltig` (Einleitung `SPEC-116`) | `TestKeyRegelnUngueltig` |
| Reihenfolge der Prüfungen bei mehreren Verstößen (Nr. 9) | `TestKeyReihenfolgeVerstoesse`: leere `Method` mit ungültigem Pfad-`%` ergibt `ErrUngueltig`; ungültiger Pfad und ungültige Query bei `reject` ergeben `ErrKodierung`; `unsafe` mit ungültiger Kodierung ergibt `ErrUngueltig` |
| Allaussage: Gleichwertigkeiten und Unterschiede der Nr. 10 gelten für alle Fälle | `TestKeyGleichwertigkeit`: eine Tabelle aus gleich-Paaren und verschieden-Paaren über Methode, Service, Pfad, Query, Vary, Body; ein Fall je Aussage der Nr. 10. Mutation: eine Normalisierung zu viel (Pfad kleinschreiben) oder zu wenig (Hex-Schreibweise): rot |
| Eingaben bleiben unverändert, gleiche Eingabe ergibt gleichen Key, unabhängig von der Iterationsfolge der Map (Nr. 10) | `TestKeyEingabeUnveraendert`, `TestKeyDeterministisch` (100 Läufe) |
| Replayfähigkeit: jede Zeile der Matrix (`SPEC-117` Nr. 1) | `TestReplayfaehigMatrix`: alle 54 Kombinationen aus drei Capture-, drei Redaction- und drei Integritätswerten und beiden Werten von `RedigierteBodies`, die Erwartung ausgeschrieben und nicht aus derselben Bedingung berechnet |
| Unbekannte oder leere Werte schließen aus (Nr. 1) | `TestReplayfaehigUnbekannt` |
| Keine weiteren Felder wirken (Nr. 2) | `TestReplayfaehigUnabhaengig`: Request-Redaction, `key_redacted`, Header, Body-Zustände, `replay_equivalence`, `match_key`, Zeiten je variiert, Ergebnis gleich |
| Nullwert ist definiert, ohne Fehler und Panik (Nr. 3) | `TestReplayfaehigNullwert` |

Nicht gefangen und als akzeptiertes Negativ benannt: dass `SPEC-116` Nr. 1 und die Hand-Kodierung des Tests dieselbe Lesart haben, weil beide aus einer Quelle stammen; der Reviewer liest Nr. 1 gegen den Test. Die Laufzeit der Kanonform (`SPEC-116` *Ungeprüft*) hat keinen Fänger, weil kein Lasttest in der Welle liegt.

- Der Match-Key ist persistiert (`exchanges.match_key`, Index `(session_id, match_key)`); seine Kodierung ist Teil des Aufzeichnungsformats (`RECORDING_FORMAT_VERSION`) und steht in `SPEC-116`.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` ist eröffnet; kein Vorgänger-Slice. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 3500 Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests und `spec/spezifikation.md` eingeschlossen); die Schätzung liegt bei rund 2400 in derselben Metrik (Produktcode rund 700, Tests rund 1400, Spezifikation rund 300; der Vorgänger lag mit 600 geschätzt bei 2369).
- `in-progress` → `open` (blockiert): der Implementer findet eine Randform ohne Ort in `SPEC-115` bis `SPEC-117` oder im Raster in §6 und muss sie an den Architect geben (`AGENTS.md` §3.12).

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

Der Vertrag ist entschieden in `SPEC-115` (Redaction), `SPEC-116` (Match-Key) und `SPEC-117` (Replayfähigkeit); die Nummern unten sind die dieser Stellen. Eine Zelle, die der Implementer ohne Ort findet, geht an den Architect (`AGENTS.md` §3.12), sie wird nicht im Code entschieden.

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16). *Fehlend* heißt `nil`, nicht vorhanden oder `""` als Nichtangabe, *leer* ein vorhandener leerer Wert, *unbekannt* ein Name oder Wert außerhalb der Menge, *fremd* eine Form, die der Vertrag nicht vorsieht. **(a)** Abbruch: gilt nicht, weil alle Funktionen rein sind, ohne Kontext und ohne Ein-/Ausgabe (`SPEC-115` Einleitung); es gibt nichts abzubrechen.

| Operation | fehlend | leer | unbekannt | fremd | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|
| `RedigiereHeader` | `nil`-Map oder leere Menge der Namen: leere Map, nichts redigiert (`SPEC-115` Nr. 2) | leere Wertliste, leerer Wert: redigiert ein Marker, nicht redigiert unverändert (Nr. 2) | Name außerhalb der Menge bleibt (Nr. 2) | Name in anderer Schreibweise, zwei Schreibweisen; Nicht-Token-Name und zwei Schreibweisen davon bleiben getrennt, leerer Name ist kein Token (Nr. 2) | (a) | idempotent (Nr. 4, letzter Punkt) |
| `Location`/`Content-Location` (in `RedigiereHeader`) | Header fehlt oder Wert ohne `?`: unverändert (Nr. 3) | Wert `""`, `?`, `http://x?` (Nr. 3, 4) | Parameter außerhalb der Menge bleibt (Nr. 3) | Fragment, `#` vor dem ersten `?`, absolute URL, ganz redigierter Header (Nr. 3) | (a) | idempotent (Nr. 4) |
| `RedigiereQuery` | `""`: `""`, keine Änderung (Nr. 4) | `&&`, `token=`, Segment ohne `=` (Nr. 4) | Name außerhalb der Menge, ungültig kodierter Name (Nr. 4) | Roh-Byte, `;`, führendes `?`, mehrere Vorkommen (Nr. 4) | (a) | idempotent, zweite Anwendung meldet nichts (Nr. 4) |
| `ParseJSONPfad` | `""`: `ErrUngueltig` (Nr. 6) | `$..`: `ErrUngueltig` (Nr. 6) | andere Pfadformen: `ErrUngueltig` (Nr. 6) | Nicht-ASCII-Name, Leerraum, Index, Platzhalter: `ErrUngueltig` (Nr. 6) | (a) | gilt nicht: der Parser hat keinen Zustand, und gleiche Eingabe ergibt gleiches Ergebnis (Einleitung) |
| `OhneRedigierte` | `nil`-Map: leere Map (Nr. 5) | leere Wertliste, leerer Wert bleiben (Nr. 5) | Header mit anderem Wert bleibt (Nr. 5) | Marker als Teil eines Wertes oder neben einem anderen Wert bleibt (Nr. 5) | (a) | gilt nicht: keine Zustandsänderung; Eingabe bleibt unverändert (Einleitung) |
| `BildeSchluessel`: `Method`, `ServiceID` | leer: `ErrUngueltig` (`SPEC-116` Nr. 2) | wie fehlend: beide Formen sind dieselbe Angabe in Go | jeder nichtleere Wert wird kodiert (Nr. 2) | Kleinschreibung, Sonderzeichen: unverändert, Schreibweise zählt (Nr. 2) | (a) | gleiche Eingabe, gleicher Key (Nr. 10) |
| `BildeSchluessel`: Pfad | `""` und ohne `/`: `ErrUngueltig` (Nr. 3) | `/`: gültig, der Root (Nr. 3) | ungültiges `%`: `literal` bleibt, `reject` `ErrKodierung` (Nr. 3, 4) | `?`, `#`, Steuerzeichen: `ErrUngueltig`; Nicht-ASCII, `//`, `.`/`..`, abschließender `/`: unverändert; Escape ohne UTF-8 (`%FF`) gültig (Nr. 3) | (a) | Hex-Schreibweise und unreserved-Escape gleich (Nr. 3, 10) |
| `BildeSchluessel`: Query-Paare | `""`: kein Paar (Nr. 4) | leeres Segment, `a=`, `=v`, `=` (Nr. 4) | Parameter außerhalb der Menge: gehen ein (Nr. 4) | ungültige Kodierung, Roh-Byte: `literal` Art 2 (Text als Zeichenfolge, byteweise), `reject` `ErrKodierung`; `;` (Nr. 4) | (a) | gleiche Paare mehrfach bleiben (Multimenge, Nr. 4) |
| `BildeSchluessel`: Vary-Header | Header fehlt, leere Wertliste: nicht im Tupel (Nr. 5) | leerer Wert: im Tupel (Nr. 5) | Header außerhalb der Menge ohne Wirkung; Name in `VOLATILE_HEADERS` oder `REDACT_HEADERS` ausgenommen (Nr. 5) | zwei Zeilen gegen Kommaliste, Rand-Leerraum, zwei Schreibweisen; ein Nicht-Token-Name geht nie ein (Nr. 5; `SPEC-115` Nr. 2) | (a) | Reihenfolge der Map ohne Wirkung (Nr. 10) |
| `BildeSchluessel`: Body | `nil`, `[]byte{}`: Art 0 (Nr. 6) | wie fehlend (Nr. 6) | Modus leer oder unbekannt: `ErrUngueltig` (Einleitung) | ungültiges JSON, BOM, Nachlauf, Surrogat, ungültiges UTF-8, Tiefe über 512: Art 2 (Nr. 6); Steuerzeichen im String: `\u00xx` klein (Nr. 6) | (a) | gleiche Bytes, gleicher Key (Nr. 10) |
| `BildeSchluessel`: `RequestRedaction`, Regeln | leer: `ErrUngueltig` (Nr. 7) | leere Mengen: nichts redigiert, Anzahl 0 (Nr. 5; `SPEC-110` Nr. 1) | `unsafe`, anderer Wert: `ErrUngueltig` (Nr. 7) | `MATCH_INVALID_ENCODING`/`MATCH_BODY_MODE` leer oder fremd: `ErrUngueltig` (Einleitung) | (a) | gleiche Eingabe, gleicher Key und gleiches `KeyRedacted` (Nr. 10) |
| `Replayfaehig` | Nullwert: falsch (`SPEC-117` Nr. 3) | leerer Wert einer Spalte: falsch (Nr. 1) | unbekannter Wert einer Spalte: falsch (Nr. 1) | nicht befragte Felder ohne Wirkung (Nr. 2) | (a) | gleiche Eingabe, gleiches Ergebnis (Nr. 3) |

**Zusatzzeilen für mehrere Verstöße am selben Schlüssel und für Modi** (`AGENTS.md` §3.16):

| Fall | Ort |
|---|---|
| Ungültige Kodierung **und** redigierter Parameter am selben Paar (`token=%zz`) | `SPEC-116` Nr. 4: die redigierte Darstellung zählt, bei `reject` kein Fehler |
| Zwei Verstöße in einem Aufruf (leere `Method` und ungültiges `%` im Pfad; ungültiger Pfad und ungültige Query; `unsafe` und ungültige Kodierung) | `SPEC-116` Nr. 9: feste Reihenfolge, der erste entscheidet |
| Name in `MATCH_VARY_HEADERS` **und** `VOLATILE_HEADERS` oder `REDACT_HEADERS` | `SPEC-116` Nr. 5: die Ausnahme gewinnt, kein Fehler |
| `Location` in `REDACT_HEADERS` **und** ein redigierter Query-Parameter darin | `SPEC-115` Nr. 3: ganz redigierter Header gewinnt |
| Zwei Schreibweisen desselben Namens, einer redigiert | `SPEC-115` Nr. 2: ein Eintrag, ein Marker |
| Ungültiges Paar (Art 2) neben gültigen Paaren | `SPEC-116` Nr. 4: sortiert nach (Art, …), Art 2 zuletzt |
| JSON-Body mit doppeltem Schlüssel **und** einzelnem Surrogat | `SPEC-116` Nr. 6: Rückfall auf Art 2 |
| Modus `literal` und `reject` × Pfad, Query | `SPEC-116` Nr. 3, 4, 9 |
| Modus `bytes` und `json-canonical` × leer, gültiges JSON, ungültiges JSON | `SPEC-116` Nr. 6 |
| `REPLAY_REDACTED_BODIES` `true`/`false` × `none`, `applied`, `unsafe` | `SPEC-117` Nr. 1 |

**Risiken:**

- `normalized_path` war nicht definiert — **Ausgang:** entfallen; entschieden vor dem Code in `SPEC-116` Nr. 3 (nur `%XX` wird angefasst, Segmentgrenzen, Schreibweise, `//`, abschließender `/` und OData-Literale bleiben, der Service-Root ist nicht Teil des Pfades).
- Query-Randformen ohne Ort (Paar ohne `=`, leerer Schlüssel, leerer Wert, nur `?`, `;`, gleiche Schlüssel in anderer Reihenfolge) — **Ausgang:** entfallen; entschieden in `SPEC-116` Nr. 4 und `SPEC-115` Nr. 4, Zellen im Raster oben.
- Header mit mehreren Werten, mehrfach vorkommende Header, Schreibweise — **Ausgang:** entfallen; entschieden in `SPEC-115` Nr. 2.
- `Vary`-Header fehlend gegen leer — **Ausgang:** entfallen; im Key verschieden (`SPEC-116` Nr. 5).
- Randformen des Vertrags (`AGENTS.md` §3.12, §3.16) — **Ausgang:** eingetreten, vom Architect entschieden. Der Implementer fand drei Zellen, deren Wortlaut mehrere Lesarten zulässt, und setzte sie im Code. Die Spezifikation bestätigt alle drei wie im Code: `\u00XX` mit kleinen Hex-Ziffern (`SPEC-116` Nr. 6; so schreibt auch RFC 8785 und der Key selbst, Nr. 1), Marke und Text der Art 2 als Zeichenfolge mit Längenpräfix (Nr. 1, 4; sonst wäre die Folge nach einem Art-2-Text nicht eindeutig lesbar) und Nicht-Token-Namen, die nie zusammengeführt werden (`SPEC-115` Nr. 2; sie stehen in keiner Konfigurationsliste und gehen nie in den Key ein). Die Prüfung des Codes gegen den Vertrag fand drei weitere Zellen, die der Code entschied und die jetzt stehen: `#` vor dem ersten `?` in `Location` (`SPEC-115` Nr. 3), Escape ohne UTF-8 im Pfad (`SPEC-116` Nr. 3) und die Schreibweise der Hex-Ziffern im Text der Art 2 (Nr. 4, 10). Der Code ändert sich nicht; die Tests oben (Zeilen `TestRedigiereHeaderNichtToken`, `TestRedigiereLocation`, `TestKeyPfad`, `TestKeyUngueltigeKodierung`, `TestKeyKodierungHandgebaut`) ziehen nach. Akzeptiertes Negativ: eine `nil`-Wertliste eines nicht redigierten Headers wird zur leeren Liste; in Go ist der Unterschied nur mit `reflect.DeepEqual` zu sehen, kein Verbraucher unterscheidet ihn.
- Die Auslieferung von `Location` mit dem Marker im Query-Wert (`SPEC-115` Nr. 3) ist eine Auslegung von `SPEC-045` („redigierte Header … nie mit Marker“): der Satz meint Header, die ganz der Marker sind. Trägt der fachliche Owner das nicht, ist die Alternative, `Location` mit geändertem Query im Replay wegzulassen — **Ausgang:** offen bis zur Closure; der Wechsel ist eine Änderung von `SPEC-115` Nr. 3 und `SPEC-045`, keine im Code.
- Größe: die Marke in §4 ist gesetzt, weil der Vorgänger-Slice die Schätzung um das Vierfache überschritt — **Ausgang:** offen.
- Der Rückfall `json-canonical` → Art 2 und das Verhalten bei `reject` im Record (400 vor Weiterleitung, kein Exchange) sind Zusagen an `slice-record-service` und `slice-replay-service`; diese tragen sie in §1 (siehe dort) — **Ausgang:** weiter offen bis zu deren Closure.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (2×); der Match-Key ist ein Vertrag mit vielen Randformen, deshalb steht das Raster vor dem Code (`AGENTS.md` §3.16).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
