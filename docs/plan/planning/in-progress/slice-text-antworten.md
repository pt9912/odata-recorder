# Slice slice-text-antworten: Text-Antworten (`$count`, `$value`) aufzeichnen und abspielen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md), Linie A.

**Bezug:** `LH-FA-04` (Happy Path: Status und Nutzdaten auch für Text-Antworten), `LH-FA-14` (Boundary: fachliche Nutzlast bleibt erhalten; Negative: fail-closed, das Geheimnis unter `/<Property>/$value` erreicht den Blob nicht), `LH-QA-07` (unbekannte Formen verhindern die Aufzeichnung nicht grundsätzlich), `LH-FA-05` (Sequenz identischer Requests, hier `$count`), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, Tests in der Stufe `test`).

**Berührte Spec-Stellen:** `SPEC-118` Nr. 1, 3, 7 und 11 (Redaktor: Ziele, Typ *Text*, Grund `charset`, Zeichensatz, Zählung, Wert, `TextZiel`) · `SPEC-046` Nr. 3 (bekannte Typen) · `SPEC-119` Nr. 2 (`Tee` weist ein Ziel beim Request und ungültige Kombinationen ab) und Nr. 8 (Eingang des `Tee`, Gründe des Verstoßes) · `SPEC-121` Nr. 5 (Aufzeichnung der Antwort, Ableitung der Ziele) · `SPEC-135` Nr. 7 c (Request-Weg des Replay, Liste der Gründe) · `SPEC-045` (Replayfähigkeit, unverändert).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-01.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Antwort auf `/<...>/$count` (Dezimalzahl) und auf `/<Property>/$value` mit einem `text/*`-Typ wird aufgezeichnet (`capture_status=complete`) und im Replay unverändert ausgeliefert; vorher verwarf der Redaktor sie als `typ` (`failed`, im Replay `ODR-E5008`), obwohl die Spezifikation `/$count` selbst als `text/plain` beschreibt. Der Referenzdienst liefert `/Products/$count` als `text/plain` und führt eine Route `/Products(<Schlüssel>)/<Property>/$value`; die Fixture des M2-Tests wechselt darauf.

**Herkunft:** Befund S3 des Reviews zu [slice-m2-referenzdienst](../done/slice-m2-referenzdienst.md): `text/*` war nicht abspielbar. Der Architect hat die Lücke als Fehler der Spezifikation eingeordnet, nicht als Absicht; keine ADR schloss `text/*` aus, die Regel stand nur in `SPEC-118` Nr. 3 und `SPEC-046` Nr. 3. Der Owner hat die Option B bestätigt (Text an den zwei Stellen `/$count` und `/<Property>/$value`, sonst unbekannt); der Wortlaut steht vor dem Code in der Spezifikation (Historie: *Text-Antworten*), das Lastenheft bleibt unberührt. Die Entscheidung steht inline in `SPEC-118` Nr. 3 und Nr. 11, es gibt keine ADR.

**Übernimmt (aus `slice-m2-referenzdienst`, `AGENTS.md` §3.13):** Die Umstellung des Referenzdienstes und der M2-Fixture von `/Products/$count`: dort bleibt die Antwort JSON, bis dieser Slice in `done/` liegt (der Aufruf als `text/plain` meldete vorher `ODR-E5008`). Dieser Slice ändert den Referenzdienst aus `tests/integration/` und die Fixture des M2-Tests, nicht dessen Zusagen (M4 bleibt als Test bestehen).

**Nutzt:** den Referenzdienst, den M2-Test und den `Handle` aus [slice-m2-referenzdienst](../done/slice-m2-referenzdienst.md), den Redaktor aus `slice-streaming-redaction`, die Pipeline aus `slice-capture-pipeline`, den Record-Dienst aus `slice-record-service` und den Replay-Dienst aus `slice-replay-service` (alle in `done/`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- `text/*` auf jedem anderen Pfad (HTML-Fehlerseiten, `text/csv`-Exporte) — bleibt `unsafe` mit Grund `typ`: ohne Namen im Strom erkennt kein Muster ein Geheimnis, und `LH-FA-14` verlangt fail-closed (Entscheidung des Owners; Folge im Betrieb: die Antwort wird weitergeleitet, aber nicht aufgezeichnet).
- Der `$value` einer Entity (`/Photos(1)/$value`, Medienressource) und binäres `$value` (`application/octet-stream`, `image/*`) — kein Property-Name, an dem eine Redaktion ansetzen könnte; bleibt `typ`.
- Request-Bodys als Text (`POST` mit `text/plain`) — der Request-Weg ist nie Ziel; ohne Namen kein Match-Key; bleibt `typ` wie bisher. Kein Folge-Slice (Bestand mit Grund, siehe *Akzeptierte Negative*).
- `$metadata` als `text/xml` — `SPEC-127` Nr. 8 g nimmt es im Import an, der Redaktor kennt `Metadata=true` nur für `application/xml`; reale Dienste senden `application/xml`; bleibt Bestand (akzeptiertes Negativ).
- Import von Text-Antworten — `SPEC-127` Nr. 8 b überspringt `$count` und Property-Pfade als `pfad`; `slice-baseline-import` (in `done/`) bleibt unverändert.
- Nachziehen bereits aufgezeichneter `failed`-Exchanges — Aufzeichnungen sind kurzlebig, kein Migrationsbedarf.
- Der Vergleich Direktzugriff gegen Proxy über Text-Antworten — [slice-protokolltreue](../open/slice-protokolltreue.md); dessen §1 führt, dass der Referenzdienst dann `text/plain`-Antworten trägt, die der Vergleich je Antwortart mit abdeckt.
- Änderung an Produktcode des Replay-Dienstes, Replayfähigkeit (`SPEC-117`, `SPEC-045`) und Import — der Replay liefert den Blob und `Content-Type` unverändert aus (`SPEC-135` Nr. 12 und 13); die Sperre sitzt allein im Redaktor. Einzig der Test Z14 liegt im Paket des Replay-Dienstes.

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

- [x] Redaktor und Ableitung nach `SPEC-118` Nr. 1, 3 und 11: `BodyEingang` mit `Zaehlung`, `Wert`, `Name`; Typ *Text* nur mit Ziel; Zeichensatz (Grund `charset`); Ziffernprüfung; Redaktion des `Wert` nach dem Property-Namen; die reine Funktion `model.TextZiel(path)` als einzige Stelle der Ableitung; Record-Dienst (`SPEC-121` Nr. 5) und Eingang des `Tee` (`SPEC-119`) reichen die Ziele für die Antwort durch, nie für den Request; `Tee` weist ein Ziel beim Request und ungültige Kombinationen mit `ErrUngueltig` ab (`SPEC-119` Nr. 2; `LH-FA-04`, `LH-FA-14`, `LH-QA-07`).
- [x] Referenzdienst: `/Products/$count` als `text/plain`, Route `/Products(<Schlüssel>)/<Property>/$value`; die Fixture des M2-Tests wechselt auf `text/plain`; Round-Trip `$count` dreimal in derselben Folge (`LH-FA-05`) und `$value` mit und ohne Treffer, Replay des Treffers nur bei `REPLAY_REDACTED_BODIES=true` (`SPEC-045`); der M2-Test läuft vor und nach dem Wechsel grün.
- [x] Je Zeile der Tabelle *Zusage · Fänger* (§3) ein Test und eine selbst rot gesehene Mutation; der Bericht nennt Zusage · Mutation · roter Test (`AGENTS.md` §3.10, §3.15, §3.17).
- [x] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (die Spezifikation trägt den Wortlaut vor dem Code; im Slice nur Nachzug nach Review und Verifikation, durch den Architect).
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert; `BEO-REPO/spezifikation-verbietet-den-antworttyp-die-sie-selbst-nennt` bekommt den Stand dieses Slice.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/model/` | update, neu (`redaktor_text.go`, Tests in `redaktor_text_test.go`) | `BodyEingang` (`Zaehlung`, `Wert`, `Name`, `Gueltig()`), Typ *Text*, Zeichensatz, Ziffernprüfung, Redaktion des `Wert`, `TextZiel(path)`, Grund `charset` (`SPEC-118` Nr. 1, 3, 11) |
| `src/hexagon/services/record/` | update | Ableitung der Ziele aus dem OData-Pfad und Durchreichen im Eingang des `Tee` (`SPEC-121` Nr. 5, `SPEC-119`), nur für die Antwort; Abweisung eines Ziels beim Request und ungültiger Kombinationen in `capture/` (`SPEC-119` Nr. 2) |
| `tests/integration/` | update | Referenzdienst: `/Products/$count` als `text/plain`, Route `$value`; Fixture des M2-Tests; Round-Trip (`LH-FA-04`, `LH-FA-05`) |
| `src/hexagon/services/replay/` | update (nur Test) | Fall Z14: ein Request als Text auf `/$count` und `/<Property>/$value` bleibt `ODR-E5001`; kein Produktcode (`SPEC-135` Nr. 7 c) |
| Testdateien der Pakete | neu / update | ein Fall je Zeile der Tabelle und je Zelle des Rasters (§6) |

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Köpfe wie `SPEC-128` Nr. 3 a). Die Zeilen folgen dem Wortlaut der Spezifikation; die Fänger-Namen legt der Implementer fest, der Planner erwartet, dass jede Zeile vor dem Handoff einen Test und eine rot gesehene Mutation trägt. Eine Allaussage ist eine Schleife über die Menge, nicht eine Liste der Fälle, die einfielen (`AGENTS.md` §3.15, §3.17).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z1 | Text ist nur mit Ziel bekannt: ohne `Zaehlung` und ohne `Wert` ist jeder `text/*` `unsafe`/`typ` (`SPEC-118` Nr. 3) | Schleife über {kein Ziel, `Zaehlung`, `Wert`} × {`text/plain`, `text/csv`, `text/html`, `TEXT/Plain`, ` text/plain `}: ohne Ziel `typ`, mit Ziel kein `typ` | das Ziel wird ignoriert (`text/*` stets bekannt) oder nie bekannt: rot | andere |
| Z2 | Randformen des Typs: `text/` ohne Untertyp, `text`, `xtext/plain`, `application/text`; fehlender, leerer und mehrfacher `Content-Type` bei gesetztem Ziel sind `typ`; der Untertyp wird nicht auf Bytes geprüft: `text/ x` ist Text (`SPEC-118` Nr. 3) | Schleife über alle Formen × {`Zaehlung`, `Wert`}; `text/ x` mit Inhalt `1` bei `Zaehlung`: kein `typ` | Prüfung `HasPrefix("text")` statt `text/` mit nichtleerem Rest; fehlender Typ gilt als `text/plain`; Untertyp auf Token-Zeichen geprüft: rot | andere |
| Z3 | JSON bleibt JSON bei gesetztem Ziel: ein `$count` als `application/json` mit Inhalt `3` geht durch die JSON-Prüfung, Status `none` (`SPEC-118` Nr. 3) | Fall `Zaehlung` mit `application/json` und mit `application/json` und Inhalt `abc` (`syntax`) | das Ziel verdrängt JSON (Ziffernweg statt JSON-Weg): der Fall mit `abc` bleibt grün oder `3` wird `typ`: rot | andere |
| Z4 | Zeichensatz (`SPEC-118` Nr. 11 c): fehlt der Parameter, gilt UTF-8; `utf-8` und `us-ascii` (Groß-/Kleinschreibung, Paar `"`) sind erlaubt; jeder andere Wert, leer, einzelnes `"`, Resttext, `charset` ohne `=`, ein abweichender unter mehreren: `charset`, ohne Ausgabe | Schleife über {fehlt, `utf-8`, `UTF-8`, `"utf-8"`, `us-ascii`, `iso-8859-1`, `utf-16`, leer, `"utf-8`, `utf-8x`, `charset` ohne `=`, doppelt (`utf-8` dann `latin1` und umgekehrt)} × {`Zaehlung`, `Wert`}; Erwartung je Wert, `dst` bleibt leer | nur der erste Parameter ausgewertet; nur `utf-8` erlaubt; `us-ascii` abgelehnt; leerer Wert erlaubt: je eine Mutation rot | andere |
| Z5 | Zählung (`SPEC-118` Nr. 11 a): der Inhalt besteht aus mindestens einem und nur aus den Bytes `0` bis `9`; jedes andere Byte ist `syntax` und wird nie geschrieben; Status `none`, byte-gleich | Schleife über jedes Byte 0x00 bis 0xFF an erster, mittlerer und letzter Stelle; leerer Rest hinter Leerraum, `\n` am Ende, BOM, `-1`, `1.5`, `+1` | Leerraum oder Zeilenende erlaubt; Prüfung nur des ersten Bytes; Vorzeichen erlaubt: rot | andere |
| Z6 | Redaktion des `Wert` (`SPEC-118` Nr. 11 b): trifft `Name` ein Muster (ASCII-Schreibweise, auch `%XX`-dekodiert), ersetzt der Marker `[REDACTED]` ohne Anführungszeichen den ganzen Inhalt, Status `applied`; kein Treffer: byte-gleich, `none`; ist der Inhalt schon genau `[REDACTED]`: `none` | Fälle `password`, `PASSWORD`, `%70assword`, kein Treffer, Marker schon da, Inhalt `[REDACTED] ` (mit Leerzeichen: `applied`), leerer Inhalt (`none` nach Nr. 2) | Marker mit Anführungszeichen; Inhalt durchgereicht bei Treffer; Marker-Idempotenz fehlt; keine Dekodierung des Namens: rot | andere |
| Z7 | Ins `dst` gelangt nie ein Byte des Geheimnisses: bei Treffer steht vor einem Verstoß nichts, beim Marker erst mit dem letzten gelesenen Byte; ohne Treffer das Präfix vor dem Verstoßbyte (`SPEC-118` Nr. 9, 11 b) | Fake-`dst` hält jeden Write fest; Fälle Treffer mit {ungültigem UTF-8, `zu_gross`, Abbruch von `src`}, jeweils Prüfung auf jedes Byte des Geheimnisses | Inhalt wird vor der Prüfung gepuffert und beim Fehler geschrieben: rot | andere |
| Z8 | UTF-8 beim `Wert` (`SPEC-118` Nr. 11 b, Nr. 4): ungültige Folge ist `utf8` auch bei Treffer; eine Mehrbyte-Folge darf an jeder Stelle zwischen zwei Reads geteilt sein; Ausgabe, Status und Grund hängen nicht von der Zerlegung in Reads ab (Nr. 8) | Schleife über jede Mehrbyte-Länge (2 bis 4 Bytes) an jeder Teilungsstelle; jede Zweiteilung und einzelne Bytes; ungültige Folgen (überlang, Surrogat, über U+10FFFF) | kein Check bei Treffer; Teilung zwischen zwei Reads gilt als ungültig: rot | andere |
| Z9 | Ableitung `TextZiel(Pfad)` (`SPEC-118` Nr. 11 d): `Zaehlung` bei Suffix `/$count` mit Gegenstand davor, `Wert` bei `/<Name>/$value` mit nichtleerem `<Name>` ohne `(`, `)`, `'`, ohne `$` am Anfang; `Name` einmal `%XX`-dekodiert, bei ungültiger Kodierung unverändert; nie beides; rein, für jeden Text definiert | Schleife über die Pfadformen: `/Products/$count`, `//$count` (`Zaehlung`: ein Zeichen davor), `/Products(1)/Orders/$count`, `/$count` (kein Ziel), `/%24count` (kein Ziel), `/Products/Name/$value`, `/Products(1)/Name/$value`, `/Products(1)/$value` (kein Ziel), `/$value` und `//$value` (kein Ziel), je ein Zeichen aus `(`, `)`, `'` im Namen (`/P/a(b/$value` und die beiden anderen: kein Ziel), `/P/$x/$value` (kein Ziel), `/P/%24x/$value` (`Name` `$x`: geprüft wird der Rohtext), `/P/%28/$value` (`Name` `(`), `/P('a/password')/$value` (kein Ziel), `/P/pass%77ord/$value` (`Name` `password`), `/P/%zz/$value` (`Name` `%zz`), leerer Pfad, Query nicht Teil | letztes Segment statt Suffix; `(` im Namen erlaubt; ohne Dekodierung; Prüfung der Zeichen am dekodierten statt am Rohtext; `/$count` allein mit Ziel: rot | andere |
| Z10 | Flag-Kombination ist `ErrUngueltig`, `src` nicht gelesen, in `dst` nichts: mehr als eines von `Metadata`, `Zaehlung`, `Wert`; `Wert` mit leerem `Name`; `Name` ohne `Wert` (`SPEC-118` Nr. 1) | Schleife über die Kombinationen × ein `src`, der Lesen zählt | Prüfung entfällt oder liest zuerst: rot | andere |
| Z11 | Die Rangfolge der bestehenden Nummern gilt für Text: leere Eingabe ist `none` vor dem Typ (Nr. 2); mehr als `MaxBytes` ist `zu_gross`, genau `MaxBytes` zulässig (Nr. 6); Fehler von `src` ist `eingabe`, von `dst` `ausgabe` (Nr. 7) | Schleife über {`Zaehlung`, `Wert`} × {leer, `MaxBytes`, `MaxBytes`+1, Fehler von `src`, Fehler von `dst`}; beide Seiten der Grenze | Typprüfung vor der leeren Eingabe; Grenze `>=` statt `>`: rot | andere |
| Z12 | Aufrufstelle Record (`SPEC-121` Nr. 5): die Antwort trägt `Zaehlung`, `Wert` und `Name` aus `TextZiel` des OData-Pfads, der Request nie; `/%24count` setzt kein Ziel; Antworten ohne Body (`HEAD`, 204, 304) gehen nicht in die Pipeline | Echte `Pipeline` mit Fakes der Ports (es gibt keinen Fake des `Tee`, er ist ein Typ), Wirkung am Exchange; Schleife über Pfadform × Richtung: Antwort `text/plain` auf `$count` und auf `$value` ist `complete`, ein Request als `text/plain` auf denselben Pfaden `failed` (`typ`), `/%24count` und `Entity/$value` als Antwort `failed`; `HEAD`, 204 und 304 ohne Body: kein Aufruf der Pipeline | Ziel auch am Request (Request `complete` statt `failed`); Ableitung am Roh-Pfad mit Query; Ziel für `HEAD`: rot | andere |
| Z13 | Der Eingang des `Tee` reicht die Ziele an den Redaktor durch (`SPEC-119` Nr. 8): eine aufgezeichnete `text/plain`-Antwort auf `$count` endet `complete`, ohne Ziel `failed` | Test an der Pipeline mit echtem Redaktor: Antwort `text/plain` mit und ohne Ziel; zwei gleichzeitige Aufnahmen mit und ohne Treffer teilen keine Bytes | der `Tee` setzt `Eingang` ohne die neuen Felder: `failed` statt `complete`: rot | nebenläufig |
| Z14 | Der Request-Weg ändert sich nicht (`SPEC-135` Nr. 7 c): ein Request-Body als `text/plain` ist `typ` und `ODR-E5001`, auch auf den Pfaden `/$count` und `/<Property>/$value`; Ziele sind dort falsch | Replay-Test mit `POST` und `text/plain`, Inhalt `12` (mit gesetztem Ziel gültig), auf beiden Pfaden, unter `MATCH_BODY_MODE` `bytes` (der Redaktor ist eine Funktion, kein Port: der Fänger sieht die Wirkung, `ODR-E5001`) | der Replay setzt das Ziel aus dem Pfad: der Request wird redigiert statt `ODR-E5001`: rot | andere |
| Z15 | Round-Trip (`LH-FA-04`, `LH-FA-05`, `SPEC-045`, `SPEC-135` Nr. 12 und 13): `/Products/$count` dreimal mit verschiedenen Zahlen wird in derselben Folge abgespielt, `Content-Type` aus den aufgezeichneten Headern; `$value` ohne Treffer ist abspielbar, mit Treffer nur bei `REPLAY_REDACTED_BODIES=true` | M2-Test und Test am `Handle` mit echten Adaptern: aufzeichnen, Session beenden, importieren, Workspace, Upstream beenden, abspielen; `$value` unter `password` und unter einem Namen ohne Muster; beide Werte der Option | Reihenfolge gekippt (oder immer die erste Antwort); Replay liefert den Treffer bei `false` aus; `Content-Type` fehlt: rot | nebenläufig |
| Z16 | Referenzdienst und M2-Test: `/Products/$count` liefert `text/plain` mit wechselnder Zahl, `/Products(<Schlüssel>)/<Property>/$value` liefert den Wert als `text/plain`; der M2-Test bleibt bei seinen Zusagen und läuft vor und nach dem Wechsel grün | Lauf des M2-Tests mit der alten und der neuen Fixture; der Referenzdienst antwortet auf `$count` und `$value` mit dem erwarteten Typ (Test am Fixture selbst) | Fixture liefert `application/json`: Z15 bleibt grün, der Typ-Fänger am Fixture ist rot | andere |
| Z17 | `Tee` weist ab (`SPEC-119` Nr. 2): ein Ziel (`Zaehlung`, `Wert`, nichtleerer `Name`) bei `Request`; bei beiden Richtungen mehr als eines von `Metadata`, `Zaehlung`, `Wert`, `Wert` mit leerem `Name`, nichtleerer `Name` ohne `Wert`: `model.ErrUngueltig`, kein Byte aus `src` gelesen, keine Goroutine | Schleife über {`Request`, `Response`} × alle Kombinationen aus `Metadata`, `Zaehlung`, `Wert` × `Name` {leer, nichtleer}; `src` zählt Reads, ein Zähler der Goroutinen wie im Fall zu `Metadata` bei `Request`; gültige Kombinationen der Antwort werden eröffnet | Prüfung entfällt für ein Ziel beim Request oder für eine Kombination (der Redaktor meldet sie dann später als `ablage`): rot | andere |

Raten (`AGENTS.md` §3.10): Z13 und Z15 sind nebenläufig (Aufnahme-Goroutine des `Tee`; Proxy, Referenzdienst und Client in einem Prozess). Der Implementer misst die Rate je Mutation und je Lauf von `go test` und trägt Zahl der Prozesse und Messenden hier ein; die Fänger ordnen mit Kanälen, nicht mit Schlaf.

Testdaten von Hand: Bytes ab `0x80`, BOM und Folgen mit Escape werden aus Bytewerten gebaut (`esc("…")`, `[]byte{…}`) und die geschriebene Datei wird auf Rohbytes gegengelesen (`AGENTS.md` §3.18, `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um`).

**Belege des Implementers** (`AGENTS.md` §3.10, §3.15, §3.17). Jede Mutation wurde auf einer Wegwerfkopie des Worktrees angewendet (Perl-Ersetzung des Quelltextes, `go test -count=1` im Image der Stufe `deps`, netzlos, `--memory 4g`) und färbt den genannten Test rot; kein Lauf der Kopie ist Teil des Repos. Gezählt sind 111 Mutationen, jede rot gesehen; eine erste Fassung der Zeile Z6 (Leerraum hinter dem Marker) blieb grün, weil die Ersetzung nichts änderte, und wurde durch die Mutation „Byte hinter dem vollständigen Marker unbeachtet“ ersetzt. Die Fänger sind Tests in den Paketen `src/hexagon/model/` (Datei redaktor_text_test.go), `src/hexagon/services/record/capture/` und `src/hexagon/services/record/` (je Datei text_test.go), `src/hexagon/services/replay/` (Datei text_test.go) und `tests/integration/` (text_test.go, m2_test.go, referenzdienst_test.go).

| Nr. | Fänger (Zahl der Fälle, die der Test zählt) | Mutationen, jede rot | Klasse |
|---|---|---|---|
| Z1 | `TestTextNurMitZiel` (24: 3 Ziele × 8 Typen) | Ziel ignoriert; Text nie bekannt; nur `Zaehlung` bekannt; nur `Wert` bekannt | Schleife über die Menge |
| Z2 | `TestTextTypRandformen` (34) | `text` ohne Schrägstrich; leerer Untertyp; mehrere Werte, der erste zählt; Untertyp auf Zeichen geprüft; Rand nicht getrimmt; Schreibweise nicht ASCII-klein (die beiden letzten fängt `TestTextNurMitZiel`) | Schleife |
| Z3 | `TestTextJSONBleibtJSON` | das Ziel verdrängt `application/json`; das Ziel verdrängt `application/<x>+json` | Fälle |
| Z4 | `TestTextZeichensatz` (46 Formen × 2 Ziele), `TestTextZeichensatzRangfolge` | nur der erste Parameter; nur `utf-8`; nur `us-ascii`; leerer Wert erlaubt; Anführungszeichen nicht entfernt; Resttext erlaubt; Name nicht ASCII-klein; Name nicht getrimmt; Wert nicht ASCII-klein; Wert nicht getrimmt; Prüfung nur bei `Wert`; Prüfung nur bei `Zaehlung`; keine Teilung an `;` | Schleife über Formen × Ziele |
| Z5 | `TestTextZaehlungBytes` (768: 256 Bytes × 3 Stellen, dazu 15 Formen und 6 Längen bis 100000) | Leerraum und Zeilenende erlaubt; nur das erste Byte geprüft; Vorzeichen erlaubt; Verstoßbyte geschrieben; Obergrenze `8`; Untergrenze `1` | Generator über alle Bytes, Orakel: Ziffer ist `0` bis `9` |
| Z6 | `TestTextWertRedaktion` (120 Treffer-Fälle: 10 Namen × 12 Inhalte, dazu je Name der Marker, 7 nicht treffende Namen × 13 Inhalte) | Marker mit Anführungszeichen; Inhalt bei Treffer durchgereicht; Marker nicht idempotent; Namensvergleich ohne ASCII-Schreibweise; nur das erste Muster; Marker-Gleichheit nur am Präfix; Status `applied` auch ohne Treffer; Byte hinter dem vollständigen Marker unbeachtet | Schleife über Namen × Inhalte |
| Z7 | `TestTextKeinGeheimnisImDst` (6 Fehlerformen, ein Fake hält jeden Write) | die ersten drei Bytes durchgelassen; Marker schon beim ersten Byte; dazu Z6 `Inhalt bei Treffer durchgereicht` | Fall je Verstoß |
| Z8 | `TestTextWertUTF8` (84 Zerlegungen und 14 ungültige Folgen × 2 Namen × jede Zerlegung), `TestTextWertUTF8Generator` (232 448 Läufe: 116 224 Körper × 2 Namen) | UTF-8 bei Treffer nicht geprüft; Folge zwischen zwei Reads ungültig; Surrogat erlaubt; überlanges `E0` erlaubt; überlanges `F0` erlaubt; über U+10FFFF erlaubt; `C0` und `C1` erlaubt; `F5` bis `F7` erlaubt; halbe Folge am Ende gültig; Folgebyte nicht geprüft; Folge mit zwei Folgebytes zählt eines; Folge mit drei Folgebytes zählt zwei (nur der Generator fängt die letzte) | Generator über jedes Bytepaar und die langen Anfänge, Orakel `utf8.ValidString` |
| Z9 | `TestTextZiel` (52 Pfadformen), `TestTextZielGenerator` (41 370 Pfade aus 14 Segmenten, 1 bis 4 Segmente), `TestTextZielRein` (Name von 100000 Bytes, jedes Byte als Name) | `/$count` allein hat Gegenstand; `(`, `)`, `'` im Namen erlaubt (drei Mutationen); `$` am Anfang erlaubt; leerer Name erlaubt; Zeichen am dekodierten statt am rohen Text geprüft; Name ohne Dekodierung; ungültige Kodierung gilt als kein Ziel; ohne Schrägstrich vor dem Namen; Suffix nur ohne Klammer im Rest | Generator, Orakel arbeitet auf den Segmenten |
| Z10 | `TestTextZielKombinationen` (16), `TestBodyArgumente` unverändert | Prüfung entfällt; mehr als ein Ziel erlaubt; `Name` ohne `Wert` erlaubt; `Wert` ohne `Name` erlaubt; `Metadata`, `Zaehlung` oder `Wert` zählt nicht (drei); Prüfung am leeren Eingang | Schleife über 2³ × 2 |
| Z11 | `TestTextRangfolge` | Typ vor leerer Eingabe; Grenze `>=` statt `>` | Fälle je Seite der Grenze |
| Z12 | `TestTextAntwortZiele` (19 Pfadformen), `TestTextAntwortKeinGeheimnisImBlob`, `TestTextRequestOhneZiel` (12), `TestTextAntwortOhneBody` (8) | Query im Pfad der Ableitung; Pfad mit Service-Root; `Zaehlung`, `Wert` oder `Name` nicht durchgereicht; falscher Name; Request bekommt ein Ziel; `HEAD`, 204 und 304 gehen in die Pipeline | Schleife über Pfadform × Richtung |
| Z13 | `TestTeeReichtZieleDurch` (4 × 2), `TestTeeTextGleichzeitig` (24 Aufnahmen, 12 mit Treffer) | `Eingang` ohne die neuen Felder; je ein Feld nicht durchgereicht (drei); gemeinsamer Zustand des Wert-Automaten (Raten unten) | Fälle; nebenläufig |
| Z14 | `TestRequestTextOhneZiel` (16: 4 Pfade × 2 Modi × 2 Inhalte) | der Replay setzt das Ziel aus dem Pfad; nur `Zaehlung`; nur `Wert` | Schleife |
| Z15 | `TestTextRoundTrip` (beide Werte der Option), `TestDurchstichM2` | Treffer wird bei `false` ausgeliefert; Reihenfolge gekippt; `Content-Type` fehlt in der Wiedergabe; Name falsch durchgereicht (das Geheimnis steht dann im Blob) | Fälle; nebenläufig |
| Z16 | `TestReferenzdienstSelbst`, dazu `TestDurchstichM2` und `TestTextRoundTrip` | `$count` als `application/json`; `$value` als `application/json`; `$count` als JSON-Zahl; das Geheimnis fehlt im Referenzdienst | Fälle |
| Z17 | `TestTeeZielKombinationen` (32 Teiltests: 2 Richtungen × 2³ × 2) | Ziel beim Request nicht abgewiesen (vier: alle drei Flags und jedes einzeln); Kombination nie abgewiesen; nur `Wert` ohne `Name`; nur `Name` ohne `Wert`; jedes der drei Paare von Zielen einzeln erlaubt (drei); `Tee` liest vor der Abweisung; `Tee` startet die Goroutine vor der Abweisung | Schleife |

Der Fänger für „nichtleerer `Name` bei `Request`“ aus `SPEC-119` Nr. 2 ist keine eigene Prüfung im Code: ein `Name` ohne `Wert` ist schon die ungültige Kombination, und die Mutation, die nur diese Klausel im Zweig für `Request` entfernte, ließ sich nicht von der Prüfung der Kombination trennen; der Zweig für `Request` nennt deshalb nur die drei Flags.

Raten (`AGENTS.md` §3.10), gemessen vom Implementer mit dem übersetzten Testbinary, je Lauf `-test.count=1`, 30 Prozesse je Zeile:

| Zusage | Mutation | Fänger | Rate |
|---|---|---|---|
| Z13 | gemeinsamer Zustand: `trifft` des Wert-Automaten als Paketvariable | `TestTeeTextGleichzeitig`, `TestTeeReichtZieleDurch`, ohne `-race` | rot in 30 von 30 Prozessen |
| Z13 | dieselbe Mutation | dieselben Tests mit `-race` (Image von `make test-race`) | rot in 30 von 30 Prozessen |
| Z13 | keine Mutation (Basis) | dieselben Tests ohne `-race` | rot in 0 von 30 Prozessen |
| Z15 | Treffer wird bei `false` ausgeliefert | `TestTextRoundTrip`, `TestDurchstichM2` | rot in 30 von 30 Prozessen |
| Z15 | Reihenfolge gekippt | dieselben | rot in 30 von 30 Prozessen |
| Z15 | `Content-Type` fehlt in der Wiedergabe | dieselben | rot in 30 von 30 Prozessen |
| Z15 | Name falsch durchgereicht | dieselben | rot in 30 von 30 Prozessen |
| Z15 | keine Mutation (Basis) | dieselben | rot in 0 von 30 Prozessen |

Die Mutationen von Z15 sind keine Datenrennen: der Detektor ist dafür nicht der Fänger, sondern der Vergleich von Status, Body und Header der Wiedergabe mit der Aufnahme. Ein Zugriff, den kein Test aus zwei Goroutinen ausführt, bleibt wie in `SPEC-118` *Ungeprüft* ungesehen; `TestTeeTextGleichzeitig` führt 24 Aufnahmen mit je eigenem Wert-Automaten gleichzeitig durch den Redaktor, `make test-race` fährt sie mit dem Detektor.

**Mengen, die der Plan generierbar nannte** (Schritt 14 von `implement-slice`): (1) UTF-8-Folgen beim Wert, Generator über jedes Bytepaar (65 536) und jedes lange Anfangsbyte `E0` bis `EF` und `F0` bis `F5` mit drei oder vier Bytes (50 688), Orakel `utf8.ValidString` der Standardbibliothek, 232 448 Läufe, die der Test zählt; (2) Pfadformen für `TextZiel`, Generator über 14 Segmente und Längen 1 bis 4, Orakel segmentweise (letztes Segment, vorletztes Segment, `url.PathUnescape`), 41 370 Pfade, die der Test zählt; (3) die Bytes `0x00` bis `0xFF` an drei Stellen der Zählung, Orakel „Ziffer“, 768 Fälle; (4) Kombinationen der Ziele, 16 beim Redaktor und 32 beim `Tee`, Orakel als Liste der gültigen Kombinationen. Handgeschriebene Fälle bleiben als Goldens daneben.

**Gelieferte Form** (Festlegungen des Implementers, die die Spezifikation offen lässt): `model.TextZiel(pfad string) (zaehlung, wert bool, name string)`; `BodyEingang` trägt die Methode `Gueltig()` (`SPEC-118` Nr. 1), die `RedigiereBody` und die Prüfung des `Tee` gemeinsam nutzen; der Grund `charset` steht in `model` und in `capture`, und die geschlossene Menge der Gründe im Protokolltest des Record-Dienstes führt ihn.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-m2-referenzdienst](../done/slice-m2-referenzdienst.md) liegt in `done/` (Referenzdienst, M2-Test und `Handle` im Code). Die Spezifikation trägt den Wortlaut vor dem Code (`SPEC-118` Nr. 11, Historie *Text-Antworten*); das Raster in §6 hat für jede Zelle einen Ort (`AGENTS.md` §3.12, §3.16), eine Zelle ohne Ort geht an den Architect. Reihenfolge in der Welle: nach `slice-m2-referenzdienst`, vor [slice-protokolltreue](../open/slice-protokolltreue.md) (Reihenfolge, kein Code-Blocker). Der Planner hat beim Start die Marke nach `SPEC-128` Nr. 1 aus der Tabelle gesetzt (`basis`, `schaetzung`, `marke` im Block unten); der Architect bestätigt `pfad`, `produktcode` und `fakes` im Block (`SPEC-128` Nr. 3 k).

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`), vom Planner gesetzt beim Start am 2026-11-05, vom Architect am selben Tag vor dem ersten Code-Commit (Messung 0) in `pfad` und `produktcode` neu gesetzt (`SPEC-128` Nr. 3 k): die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3 (`src/hexagon/model/`, `src/hexagon/services/record/`, `src/hexagon/services/replay/`, `tests/integration/`), Tests eingeschlossen; Text der Spezifikation zählt nicht mit. Zählung nach der Tabelle in §3: *a* = 2 nebenläufige (Z13, Z15) und *b* = 15 andere Zeilen, *f* = 1 (der Referenzdienst als Test-Server; ein Fake des `Tee` entfällt, denn die `Pipeline` ist ein Typ und die Tests des Record-Dienstes fahren sie echt mit Fakes der Ports), *p* = 260 (Produktcode nach der Erwartung des Architect: Ziffernprüfer 25, Wert-Automat mit UTF-8-Prüfung und Marker 90, Zeichensatz-Parser 45, `TextZiel` mit Dekodierung 45, `BodyEingang`, Typ *Text* und Prüfung der Kombinationen 35, Verdrahtung in `ende.go` und `Tee` 20; die Erwartung 150 ließ dem Wert-Automaten keinen Platz, denn der UTF-8-Leser des JSON-Automaten ist an dessen Zustände gebunden). *pfad* um `src/hexagon/services/replay/` ergänzt: Z14 liegt als Test dort, und die Messung ließe ihn sonst ungezählt, während die Tabelle ihn zählt. Schätzung S = 85·2 + 80·15 + 200·1 + 260 = **1830**, Marke M = 100 · ⌈135·1830 / 10000⌉ = **2500** (135·1830 = 247050); keine Schnittfrage (S liegt 393 unter 2223). Die Marke 2300 (S 1640) ersetzt der Eintrag `ersetzt` im Block; Grund: die Zeile Z17, der Produktcode und der Pfad. Der Block trägt die Fassung von `SPEC-128` Nr. 1, die `make groessenmarke` verlangt. Die Formel von `AGENTS.md` §3.14 (85 und 40 je Zeile, 200 je Test-Server, Produktcode nur bei einem Leser fremder Bytes, × 1,25) ergäbe 85·2 + 40·15 + 200·1 = 970, × 1,25 = 1212,5, Marke 1300 (mit *p* = 260: 1230, Marke 1600); der Konflikt beider Fassungen ist eine offene Entscheidung des Users, nicht des Planner. Die Grobschätzung 1100 des Architect entstand in der alten Metrik und gilt nicht mehr. Der Architect hat `pfad`, `produktcode` und `fakes` bestätigt bzw. neu gesetzt; die Marke folgt der Tabelle (`SPEC-128` Nr. 5 c), und der Planner prüft Zahl und Rechnung beim Handoff. Der Block:

```groessenmarke
pfad: src/hexagon/model/
pfad: src/hexagon/services/record/
pfad: src/hexagon/services/replay/
pfad: tests/integration/
produktcode: 260
fakes: 1
basis: 73e140b99f0ee6ff55adee3f5b39a992fdfebf39
schaetzung: 1830
marke: 2500
ersetzt: 2300 | Zeile Z17 (Abweisung im Tee), Produktcode 260 statt 150 (Wert-Automat), Pfad des Replay-Dienstes für den Test Z14; vor dem ersten Code-Commit
```

**Messung** (Implementer beim Handoff, vor dem Review): die Summe aus Einfügungen und Löschungen von `git diff --numstat` zwischen `basis` des Blocks und dem Stand des Handoffs über die vier Pfade des Blocks ist **1420** (1395 Einfügungen, 25 Löschungen), die Marke ist 2500 und die Schätzung 1830. Der Produktcode ohne Testdateien umfasst 250 Zeilen (erwartet 260), die Tests 1170. Die Marke ist nicht überschritten; Review und Verifikation messen den Stand danach neu.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke (sie ist die Schätzung × 1,35 nach `SPEC-128` und schneidet nicht, `AGENTS.md` §3.14); offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten).
- `in-progress` → `open` (blockiert): eine Randform ohne Ort in der Spezifikation (an den Architect, `AGENTS.md` §3.12); der Referenzdienst oder die Fixture aus `slice-m2-referenzdienst` trägt den Wechsel nicht (Befund an den Planner, der Geber liegt dann in `done/`, `AGENTS.md` §3.13).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, die Aufzeichnung und das Abspielen einer `text/plain`-Antwort auf `/Products/$count` und `/Products(<Schlüssel>)/<Property>/$value` sind Teil von `make test` und `make test-race`, der M2-Test läuft vor und nach dem Wechsel der Fixture grün (die Closure nennt beide Läufe), und je Zeile der Tabelle ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Der Wechsel der Fixture bricht den M2-Test (Referenzdienst und Fixture gehören `slice-m2-referenzdienst`) — **Ausgang:** offen; der M2-Test läuft vor und nach dem Wechsel grün, M4 (Reihenfolge) bleibt als Test bestehen, die Closure nennt beide Läufe. Beleg des Implementers: `TestDurchstichM2` lief mit dem neuen Produktcode und der alten Fixture (JSON `{"count":n}`) grün, dann mit der Fixture `text/plain` grün; der Lauf der neuen Tests auf dem Stand vor dem Slice (ohne den Produktcode, gemessen auf einer Kopie) schlägt mit der neuen Fixture fehl: die Session endet `incomplete`, weil der Redaktor das `$count` als `typ` verwirft; das ist die Herkunft des Slice.
- Auslegung des Wortlauts von `SPEC-118` Nr. 11 b: „beim Marker erst mit dem letzten gelesenen Byte“ ist als *am Ende des Stroms, nur bei Erfolg* umgesetzt (ein Verstoß vor dem Ende lässt das `dst` leer, auch ein `eingabe`-Fehler und `zu_gross`); der Test `TestTextKeinGeheimnisImDst` hält es fest — **Ausgang:** offen; Adresse: Architect (Bestätigung oder Schärfung des Wortlauts; ein Marker schon vor dem Ende wäre eine andere Zusage, keine Mutation dieser).
- Auslegung von `SPEC-118` Nr. 11 b: eine halbe Mehrbyte-Folge am Ende des Stroms ist beim `Wert` der Grund `utf8` (Nr. 4 kennt das nur im String als `unvollstaendig`, für den Text gibt es keinen String); ohne Treffer trägt das `dst` die Bytes der halben Folge (Nr. 9: eine Mehrbyte-Folge vor dem Verstoß kann halb stehen) — **Ausgang:** offen; Adresse: Architect (Bestätigung).
- Auslegung von `SPEC-118` Nr. 11 d: `/<Name>/$value` verlangt einen Schrägstrich vor `<Name>`; `/Name/$value` (Name unmittelbar hinter dem Service-Root) ist ein `Wert` mit `Name`, `x/$value` ohne führenden Schrägstrich hat kein Ziel; `TextZiel` nimmt den Pfad ohne Query, die Query abzuschneiden ist Sache des Aufrufers (der Record-Dienst übergibt den Pfad ohne Query, `TestTextAntwortZiele`) — **Ausgang:** offen; Adresse: Architect (Bestätigung).
- Ein Geheimnis in einem `$value` unter einem Namen, den kein Muster nennt, steht im Blob (wie im JSON, `SPEC-118` *Ungeprüft*) — **Ausgang:** offen; akzeptiertes Negativ, Folge: der Betreiber nennt den Namen in `REDACT_JSON_PATHS`; der Fänger Z6 weist den Fall „kein Treffer: byte-gleich“ aus.
- Nebenläufigkeit: Aufnahme-Goroutine des `Tee`, Proxy, Referenzdienst und Client in einem Prozess (Z13, Z15) — **Ausgang:** offen; `make test-race` fährt `tests/` und die Pakete (`SPEC-122` Nr. 2); ein Fänger mit Wahrscheinlichkeit nennt seine Rate je Mutation und je Lauf von `go test` (`AGENTS.md` §3.10). Raten in §3: jede Mutation von Z13 und Z15 war in 30 von 30 Prozessen rot, die Basis in 0 von 30; der Wirkungs-Fänger von Z13 (`TestTeeTextGleichzeitig` prüft den Inhalt jedes Blobs) fängt den gemeinsamen Zustand auch ohne Detektor.
- Die Ziele werden an der Aufrufstelle des Record-Dienstes abgeleitet; ein zweiter Aufrufer des Redaktors (Replay, Import) bekäme sie nie — **Ausgang:** offen; Z12 und Z14 fangen beide Seiten (Antwort trägt, Request nie); die Menge der Aufrufstellen ist Schleife (`AGENTS.md` §3.15).
- Größenmarke (§4): gesetzt beim Start, vom Architect vor dem Code neu gesetzt, *S* 1830, *M* 2500 — **Ausgang:** entfallen; der Block steht in §4, `pfad`, `produktcode` und `fakes` sind bestätigt. Offen bleibt der Formelkonflikt `AGENTS.md` §3.14 gegen `SPEC-128` (Entscheidung des Users).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* für `RedigiereBody` (neue Ziele); jede Zelle nennt ihren Ort im Wortlaut der Spezifikation, den der Architect vor dem Code entschieden hat; eine Zelle ohne Ort hält den Implementer an — **Ausgang:** offen. „118/3“ ist `SPEC-118` Nr. 3, „118/11 a“ ihre Nr. 11 Unterpunkt a.

| Argumentart | `Zaehlung` | `Wert` |
|---|---|---|
| Content-Type fehlt, leer oder mehrere Werte | 118/3: `typ` | 118/3: `typ` |
| `application/json` (JSON-Typ) | 118/3: JSON-Weg, Ziel ohne Wirkung | 118/3: JSON-Weg |
| `text/plain` | 118/11 a: Ziffernprüfung | 118/11 b: UTF-8 und Namens-Redaktion |
| anderes `text/*` (`text/csv`, `text/html`) | 118/11 a | 118/11 b |
| `text/` ohne Untertyp, `text`, `xtext/plain`, `application/text` | 118/3: unbekannt, `typ` | 118/3: `typ` |
| `text/` mit beliebigen Bytes im Untertyp (`text/ x`) | 118/3: Text, weiter wie `text/plain` | 118/3: Text |
| `application/octet-stream`, `image/*`, `multipart/*` | 118/3: `typ` | 118/3: `typ` |
| Zeichensatz fehlt | 118/11 c: UTF-8 angenommen | 118/11 c |
| Zeichensatz `utf-8` oder `us-ascii` (Schreibweise, Anführungszeichen) | 118/11 c: erlaubt | 118/11 c |
| Zeichensatz anderer, leer, halb zitiert, mehrfach mit Abweichler | 118/11 c, 118/7: `charset` | 118/11 c, 118/7: `charset` |
| Kombination der Ziele (`Metadata` zugleich `Zaehlung` oder `Wert`; `Wert` ohne `Name`; `Name` ohne `Wert`) | 118/1: `ErrUngueltig` | 118/1: `ErrUngueltig` |
| Body leer | 118/2: `none`, vor dem Typ | 118/2 |
| Body mit Nicht-Ziffer, Leerraum, BOM, Zeilenende, Vorzeichen | 118/11 a: `syntax` | gilt nicht: jeder UTF-8-Text ist Wert (118/11 b) |
| Body mit ungültigem UTF-8 (auch zwischen zwei Reads geteilt) | gilt nicht: schon `syntax` (118/11 a) | 118/11 b: `utf8` |
| Body über `MaxBytes` | 118/6: `zu_gross` | 118/6 |
| Abbruch von `src`, Fehler von `dst` | 118/7: `eingabe`, `ausgabe` | 118/7 |
| Zerlegung in Reads, Wiederholung auf eigener Ausgabe | 118/8 | 118/8 und 118/11 b (Marker idempotent) |
| `Name` trifft ein Muster (Schreibweise, `%XX`-dekodiert) | gilt nicht: Ziffern tragen keinen Namen | 118/11 b: Marker, `applied` |
| `Name` trifft kein Muster | gilt nicht: wie oben | 118/11 b: byte-gleich, `none` |
| Body schon `[REDACTED]` bei Treffer | gilt nicht: wie oben | 118/11 b: `none` |
| Text ohne Ziel (Standardpfad, `$value` einer Entity, `Name` mit Klammer) | 118/3: `typ` | 118/3 und 118/11 d: `typ` |

Operation `TextZiel(Pfad)` (Ableitung der Ziele), Zellen je Pfadform, Ort `SPEC-118` Nr. 11 d: `/Products/$count` und `/Products(1)/Orders/$count` → `Zaehlung`; `/$count` allein und `/%24count` → kein Ziel; `/Products/Name/$value` und `/Products(1)/Name/$value` → `Wert`; `/Products(1)/$value` → kein Ziel; `/P('a/password')/$value` → kein Ziel; `/P/pass%77ord/$value` → `Name` `password`; `/P/%zz/$value` → `Name` `%zz`; die Query zählt nicht.

Operation Record-Aufruf (`SPEC-121` Nr. 5): setzt die Ziele nur für die Antwort, nie für den Request und nie für Antworten ohne Body; die Methode ist gleichgültig, `HEAD` hat keinen Body.

Operation `Tee(Eingang)` (`SPEC-119` Nr. 2), Zellen je Argumentart: ein Ziel (`Zaehlung`, `Wert`, nichtleerer `Name`) bei `Request` → `ErrUngueltig`, kein Byte gelesen; bei `Response` mehr als eines von `Metadata`, `Zaehlung`, `Wert`, `Wert` mit leerem `Name`, `Name` ohne `Wert` → `ErrUngueltig`; genau ein Ziel mit passendem `Name` oder keines → eröffnet; Fänger Z17.

Operation Replay: `SPEC-135` Nr. 12 und 13 und `Replayfähig` (`SPEC-117`) ändern sich nicht; der Request-Weg (`SPEC-135` Nr. 7 c) setzt nie ein Ziel. Bereits aufgezeichnete Text-Exchanges bleiben `failed`/`unsafe` und nicht replayfähig.

Zusatztabelle (`AGENTS.md` §3.16), Paare von Verstößen am selben Strom und Modus:

| Paar | Entscheidung | Fänger |
|---|---|---|
| `charset` und Nicht-Ziffer im Inhalt | `charset` zuerst, ohne Ausgabe: der Zeichensatz kommt vor jeder Inhaltsprüfung (`SPEC-118` Nr. 11) | Fall `iso-8859-1` mit Inhalt `1a`: Grund `charset`, `dst` leer |
| `charset` und `zu_gross` | `charset` zuerst (`SPEC-118` Nr. 11) | Fall `latin1` mit Inhalt über `MaxBytes`: `charset` |
| `typ` und `charset` | `typ` zuerst (`SPEC-118` Nr. 3 vor Nr. 11) | Fall `text/plain; charset=latin1` ohne Ziel: `typ` |
| `utf8` und `zu_gross` beim `Wert` | der erste Verstoß im Strom entscheidet (`SPEC-118` Nr. 7) | Fälle mit dem ungültigen Byte vor und hinter dem Byte `MaxBytes`+1 |
| Treffer und Verstoß beim `Wert` | vor dem Verstoß steht nichts im `dst`, auch kein Byte des Inhalts (`SPEC-118` Nr. 9) | Z7 |
| Modus `REPLAY_REDACTED_BODIES` `true` und `false` bei `applied` | Replay liefert `applied` nur bei `true` (`SPEC-045`) | Z15 mit beiden Werten der Option |
| `MATCH_BODY_MODE` | betrifft nur den Request-Body (`SPEC-135` Nr. 7 c), die Antwort kennt keinen Modus | Z14 unter `bytes` |

**Akzeptierte Negative (mit Grund und Folge, `AGENTS.md` §3.15):**

- Strenge Ziffern-Grammatik: ein Dienst, der ein Zeilenende an das `$count` hängt oder eine Fehlerseite als `text/plain` sendet, wird `failed`. Grund: nur so ist der Body nachweislich frei von Geheimnissen. Folge im Betrieb: nicht abspielbar, der Proxy leitet weiter; eine Lockerung ist eine Folge-Entscheidung auf Beobachtung hin.
- Geheimnis unter einem `$value`-Namen, den kein Muster nennt: siehe oben (Risiko 2).
- Ein Dienst ohne `Content-Type` auf `$count` oder `$value`: `typ`. Grund: ohne Typ ist nicht gesagt, dass es Text ist. Folge: Austausch `failed`, der Proxy leitet weiter.
- Request-Bodys als Text und `$metadata` als `text/xml`: unverändert nicht aufzeichenbar (§1). Folge: `POST` mit Text-Body und ein Dienst, der `$metadata` als `text/xml` sendet, werden nicht aufgezeichnet.
- Kein Nachziehen bestehender `failed`-Aufzeichnungen. Folge: sie bleiben nicht replayfähig.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage (2026-11-01) durchgegangen; der Planner zieht Treffer und Zähler beim Start nach. Einschlägig: `BEO-REPO/spezifikation-verbietet-den-antworttyp-die-sie-selbst-nennt` (neu, Quelle dieses Slice, Stand `offen`); `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (verkörpert in `AGENTS.md` §3.16) und `BEO-REPO/ziffern-klasse-ohne-schleife` — die Ziffernprüfung und die Pfadformen sind Allaussagen über Mengen und brauchen Schleifen (Z1, Z5, Z9); `BEO-REPO/fix-deckt-die-genannte-mutation-nicht-die-klasse` (verkörpert in `AGENTS.md` §3.17) — Zeichensatz und Typ-Randformen als Klasse (Z2, Z4); `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16) — Raster in §6; `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (verkörpert in `AGENTS.md` §3.10) — Z13, Z15; `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` — Testdaten mit Bytes ab `0x80` aus Bytewerten (§3); `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (verkörpert in `harness/mk/groessenmarke.mk`) — Marke und Schätzung in §4. Ohne Bezug: die übrigen Einträge.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
