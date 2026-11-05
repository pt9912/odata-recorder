# Slice slice-protokolltreue: Vergleichstest der Protokolltreue (Direktzugriff gegen Proxy)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-QA-07` (Messmethode: Direktzugriff gegen Proxy-Zugriff am Referenzdienst), `LH-FA-01` (transparente Weiterleitung), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, Test in der Stufe `test`).

**Berührte Spec-Stellen:** `LH-QA-07.a` (Relevanz der Header und Messmethode, Entscheidung des Owners zu `Content-Encoding`) · `LH-FA-01.a` · `SPEC-119` Nr. 6 (Kette und Codings des Entpackens) · `SPEC-121` Nr. 4 und 7 (Hop-by-Hop, `Accept-Encoding`) · `SPEC-063` (`VOLATILE_HEADERS`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** pt9912 (Zweig slice-protokolltreue)

**Autor:** pt9912. **Datum:** 2026-10-21.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Test ruft den Referenzdienst aus [slice-m2-referenzdienst](../done/slice-m2-referenzdienst.md) einmal direkt und einmal über den Recorder im Modus `record` auf und vergleicht die entpackte Darstellung der Antworten nach `LH-QA-07.a` — Status, Body nach dem Entpacken seines `Content-Encoding`, relevante Header ohne `VOLATILE_HEADERS` —; er läuft in der Stufe `test` und ist damit Teil von `make test`. Er ist der Nachweis zur Messmethode von `LH-QA-07`.

**Übernimmt:** Den Vergleichstest der Messmethode von `LH-QA-07`, der in der Welle keinen Träger hatte (Closure-Notiz von `slice-record-service`, *Offene Punkte für den User*); Entscheidung des Users vom 2026-10-21: ein eigener Slice am Ende der Welle. Die Messmethode selbst steht in `LH-QA-07.a` der Spezifikation (Architect, nach der Entscheidung des Owners zu `Content-Encoding`); dieser Slice setzt sie um und ändert sie nicht.

**Nutzt:** den Referenzdienst von [slice-m2-referenzdienst](../done/slice-m2-referenzdienst.md) und den Composition Root (`Build`, `Handle`) von `slice-bootstrap-m2`; `slice-m2-referenzdienst` bleibt beim Satz von M2 (Entscheidung des Users). Was der Vergleich am Referenzdienst zusätzlich braucht (Antworten mit `Content-Encoding`, mehrere Werte je Header, die Randform aus §6), bringt dieser Slice selbst mit.

**Nutzt außerdem (Sendung von [slice-text-antworten](../done/slice-text-antworten.md), `AGENTS.md` §3.13):** Der Referenzdienst trägt dann `text/plain`-Antworten (`/Products/$count` und `/Products(<Schlüssel>)/<Property>/$value`); die Antwortarten des Vergleichs (§2, ein Fall je Antwortart des Referenzdienstes) schließen sie ein, und der Vergleich entpackt nur das Coding, nicht den Typ. Ein Unterschied zwischen Direktzugriff und Proxy bei diesen Antworten ist ein Befund wie jeder andere (§6). Der Slice liegt nach `slice-text-antworten` (Reihenfolge, kein Code-Blocker).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ergänzungen am Referenzdienst oder am M2-Test als Umfang von `slice-m2-referenzdienst` — dort bleibt der Satz von M2 (Entscheidung des Users); dieser Slice bringt, was der Vergleich braucht, selbst mit und ändert den M2-Test nicht.
- Replay — `LH-QA-07` misst den Record-Modus; den byte-gleichen Rundlauf über Replay prüft der M2-Test von `slice-m2-referenzdienst`.
- Markierung unbekannter OData-Erweiterungen als `unsupported_features` (zweiter Satz von `LH-QA-07.a`, `LH-FA-06.a`) — `slice-odata-codec-csdl` (dort DoD, Validierungsbericht); dieser Slice prüft nur, dass der Proxy eine Antwort unverändert weiterleitet, nicht, was der Import daraus macht.
- Änderungen an Record-Dienst, Capture-Pipeline oder Upstream-Adapter — Schicht-Abgrenzung: ein Unterschied, den der Vergleich findet, ist ein Befund an den Planner (die Nehmer liegen dann in `done/`, `AGENTS.md` §3.13), kein Code in diesem Slice.
- Lasttest zu `LH-QA-01` — Welle §6.

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

- [ ] Vergleichstest nach `LH-QA-07.a`: derselbe Request direkt und über den Recorder im Modus `record` an den Referenzdienst; verglichen werden Status, Body nach dem Entpacken (Kette und Codings wie `SPEC-119` Nr. 6) und die relevanten Header (ohne Hop-by-Hop, `Content-Encoding`, `Content-Length` und `VOLATILE_HEADERS`); der Client nennt in `Accept-Encoding` nur `gzip`, `deflate` und `identity`; je Antwortart des Referenzdienstes ein Fall, darunter `$metadata`, Collection, Einzel-Entity, `$count` und `$value` als `text/plain`, Antwort mit `Content-Encoding`, Fehlerstatus (`LH-QA-07`, `LH-FA-01`).
- [ ] Mutationen des Vergleichs: der Proxy ändert ein Byte des Bodys, einen relevanten Header oder den Status, der Vergleich übergeht einen relevanten Header — jede färbt den Test rot; ein anderes Coding über den Proxy färbt ihn nicht (`LH-QA-07.a`).
- [ ] Der Test läuft in der Stufe `test` (netzlos, nur Loopback) und ist damit Teil von `make test`; die Randformen aus §6 sind vor dem Code in der Spezifikation entschieden.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`LH-QA-07.a`: Nachzug nach Review und Verifikation, durch den Architect).
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
| `tests/integration/` | neu | Vergleichstest und die Antworten des Referenzdienstes, die er zusätzlich braucht: ein eigener, zustandsloser Test-Server `vergleichsdienst_test.go` mit 16 Antwortarten; der Referenzdienst von `slice-m2-referenzdienst` zählt `$count` und bleibt unverändert (`tests/integration/m2_test.go`, `tests/integration/referenzdienst_test.go`, `tests/integration/text_test.go` ohne Diff); Composition Root aus `slice-bootstrap-m2` |
| `spec/spezifikation.md` | update | `LH-QA-07.a`: die Randformen aus §6, vor dem Code entschieden vom Architect; im Slice nur Nachzug nach Review und Verifikation |
| `tests/integration/vergleich_test.go`, `protokolltreue_test.go` | neu | die Vergleichsfunktion (Header als Listen der Feldzeilen, ausgenommene Namen, Entpacken je Seite) und die Tests: ein Fall je Antwortart und je Randform des Rasters; der Veränderer zwischen Client und Recorder trägt die Mutationen nach `LH-QA-07` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit den Randformen in `LH-QA-07.a` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12).

- Der Vergleich prüft sich selbst: eine Mutation des Vergleichs (ein relevanter Header fehlt in der Menge, das Entpacken entfällt) muss ihn rot machen, sonst bewiese ein grüner Vergleich nichts (`AGENTS.md` §3.15).
- Testdaten von Hand: erwartete Header-Listen und Bodies stehen ausgeschrieben; Bytes ab `0x80` als Go-Escapes (`BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um`).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Köpfe wie `SPEC-128` Nr. 3 a). Die Zeilen folgen dem Wortlaut von `LH-QA-07.a` Nr. 1 bis 7; die Namen der Fänger legt der Implementer fest, der Planner erwartet, dass jede Zeile vor dem Handoff einen Test und eine selbst rot gesehene Mutation trägt. *Mutator* heißt ein Handler zwischen Client und Proxy-Antwort, der eine Antwort verändert (Byte, Header, Status), ohne den Proxy zu ändern; *Fall* ist ein Aufruf je Antwortart des Referenzdienstes: `$metadata`, Collection, Einzel-Entity, `$count` und `$value` (beide `text/plain`), Fehlerstatus, Antwort mit `Content-Encoding`. Eine Allaussage ist eine Schleife über die Menge, nicht eine Liste der Fälle, die einfielen (`AGENTS.md` §3.15, §3.17); die Fixture gleicht nie dem Wert der Mutation. Die Zusatzfälle des Referenzdienstes (Antwort ohne Body, `br` auf `Accept-Encoding: gzip`, zwei Feldzeilen je Name, ein leerer Wert, Echo des empfangenen `Accept-Encoding` in einem relevanten Header) bringt der Slice selbst mit (§1 *Nutzt*).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z1 | Der Statuscode beider Seiten ist gleich, je Antwortart; die Reason-Phrase zählt nicht (`LH-QA-07.a` Nr. 1) | Schleife über die Fälle; Test der Vergleichsfunktion mit Phrase `200 OK` gegen `200 Fine`: gleich | Mutator setzt 200 auf 201, 404 auf 500 (je Fall); Vergleich liest die Phrase | andere |
| Z2 | Der Body nach dem Entpacken ist gleich, je Antwortart (Nr. 3) | Schleife über die Fälle | Mutator ändert das erste, ein mittleres und das letzte Byte des Bodys (je Fall); der Vergleich vergleicht nur die Länge | andere |
| Z3 | Jeder relevante Header ist gleich, je Antwortart und je Header des Falls (Nr. 1) | Schleife über Fälle und Header | Mutator ändert den Wert eines Headers; der Vergleich übergeht einen relevanten Header (jeder Name der Fälle, einzeln in die Ausnahmemenge) | andere |
| Z4 | Ein relevanter Name nur auf einer Seite ist eine Abweichung, in beiden Richtungen (Nr. 1) | Schleife über Fälle und Richtung | Mutator entfernt den Header, Mutator fügt einen hinzu; der Vergleich prüft nur die Namen der direkten Seite | andere |
| Z5 | Ein leerer Wert ist ein Wert: leer ≠ fehlend, leer gleich leer (Nr. 1) | Fall mit leerem Header; Fixture: Wert `x` statt leer in der Mutation | Mutator macht den leeren Wert zu fehlend und zu `x`; der Vergleich behandelt leer wie fehlend | andere |
| Z6 | Die Reihenfolge der Feldzeilen eines Namens zählt, zwei Zeilen sind nicht eine Zeile mit Komma (Nr. 1) | Fall mit zwei Zeilen desselben Namens | Mutator vertauscht die Zeilen; Mutator führt sie zu einer Zeile zusammen; der Vergleich sortiert die Werte | andere |
| Z7 | Die Schreibweise des Namens ändert nichts: gleich, wenn sie sich nur darin unterscheidet (Nr. 1, 2) | Test der Vergleichsfunktion mit handgebauten Maps in nicht kanonischer Schreibweise (`net/http` kanonisiert auf dem Draht, ein Aufruf sähe es nie) | der Vergleich folgt der Schreibweise des Namens | andere |
| Z8 | Jeder Name von `VOLATILE_HEADERS` (`SPEC-063`, sieben) ist ausgenommen, auch in anderer Schreibweise und bei verschiedenem Wert (Nr. 2) | Schleife über die sieben Namen × zwei Schreibweisen, Wert je Seite verschieden | ein Name fehlt in der Menge (jedes Element einzeln, rot) | andere |
| Z9 | Jeder Hop-by-Hop-Name der festen Liste und jeder Name, den `Connection` einer der beiden Antworten nennt, ist ausgenommen (Nr. 2) | Schleife über die Liste (neun) und über die Seite, die `Connection` trägt | die Liste verliert ein Element (jedes einzeln); die Namen aus `Connection` zählen nur von der direkten Seite | andere |
| Z10 | `Content-Encoding` und `Content-Length` sind ausgenommen: ein anderes Coding über den Proxy mit gleichem entpacktem Inhalt ist grün (Nr. 2, `LH-QA-07.a` Satz 4) | Fall: eine Seite `gzip`, die andere ohne Coding, gleicher Inhalt | der Vergleich zählt `Content-Encoding` oder `Content-Length` als relevant | andere |
| Z11 | Jede Kette der Menge {`identity`, `gzip`, `deflate`, `deflate, gzip`, leerer Rohbody mit `gzip`} entpackt zum selben Inhalt (Nr. 3) | Schleife über die fünf Elemente, je Seite mit eigener Kette | das Entpacken entfällt (Rohbytes verglichen, bei verschiedenen Codings rot); die Reihenfolge der Kette wird vor dem Entpacken nicht umgekehrt; der Dekoder wird auf den leeren Rohbody gerufen | andere |
| Z12 | Paar: anderes Coding und geänderter Body — das Entpacken verdeckt keinen Unterschied (Nr. 3) | Fall mit Coding, Mutator ändert ein Byte des entpackten Inhalts und kodiert neu (anderes Coding als direkt) | Mutator wie links; der Vergleich prüft nur, ob beide Seiten entpackbar sind | andere |
| Z13 | Eine Antwort ohne Body (`HEAD`, 204, 304) ist auf beiden Seiten ohne Body, ihre Header gelten (Nr. 3) | Schleife über die drei Arten; ein Body nach 204 und nach 304 an handgebauten Antworten, auf jeder Seite (`net/http` liefert dort keinen Body aus, ein Mutator kann ihn nicht senden) | Mutator ändert bei 304 den Status auf 200 und bei `HEAD` einen relevanten Header; der Vergleich übergeht den Body bei 204 oder bei 304 | andere |
| Z14 | Eine Kette, die der Test nicht entpackt, macht den Fall rot und wird nicht übersprungen (Nr. 3) | Schleife über {unbekanntes Glied `zstd`, drei Glieder, leeres Glied, abgelehnter Strom} an der Vergleichsfunktion | der Vergleich überspringt den Fall oder vergleicht dann nur den Status | andere |
| Z15 | Nicht angefragtes Coding: `br` auf `Accept-Encoding: gzip` ergibt gleiche Rohbytes und gleichen Wert von `Content-Encoding` (Nr. 4) | Fall mit festen Bytes in `br`; Erwartung byte-gleich | Mutator ändert ein Byte; Mutator ändert den Wert von `Content-Encoding`; der Vergleich entpackt trotzdem oder schließt den Fall aus | andere |
| Z16 | Der `Accept-Encoding` des Clients erreicht den Dienst wie gesetzt: fehlend bleibt fehlend, der Transport entpackt nicht (Nr. 5) | Fall ohne den Header und Fall mit `gzip`; der Referenzdienst gibt den empfangenen Wert in einem relevanten Header zurück; der Client sieht `Content-Encoding` und die Bytes `1f 8b` der direkten Antwort | der Proxy ergänzt `identity`; der Test schaltet das Entpacken des Transports nicht ab | andere |
| Z17 | Der Referenzdienst antwortet auf denselben Request gleich, und der Client folgt keiner Weiterleitung (Nr. 6) | je Fall Selbstvergleich direkt gegen direkt grün; Fall mit 302 und `Location`, Erwartung Status 302 | der Dienst zählt (`$count` der Fixture von M2) und der Selbstvergleich wird rot; der Client folgt der Weiterleitung | andere |
| Z18 | `text/plain`-Antworten (`$count`, `$value`): `Content-Type` und Body gleich über den Proxy (Nr. 1, 3; Sendung von `slice-text-antworten`) | je ein Fall; Erwartung `text/plain` ausgeschrieben | Mutator setzt `Content-Type` auf `application/json`; Mutator ändert den Wert der Zahl | andere |
| Z19 | Der Client nennt in `Accept-Encoding` nur `gzip`, `deflate`, `identity` (Nr. 7 c) | Schleife über die Fälle am Echo des Dienstes: jedes Glied liegt in der Menge | der Client nennt zusätzlich `br` | andere |
| Z20 | Akzeptierte Negative, nicht verglichen: Abbruch im Body, Upstream nicht erreichbar (Nr. 7 a, b) | keiner; Grund und Folge in `LH-QA-07.a` Nr. 7 | — (Negativ: kein Fänger, der Vergleich liest nur vollständige Antworten) | andere |
| Z21 | Proxy, Referenzdienst und Client laufen in einem Prozess ohne Datenrennen (§6) | `make test-race` über den Vergleichstest, rot je Prozess (`-test.count=1`) | der Referenzdienst schreibt einen Zähler ohne Sperre (Rate je Prozess nennt der Bericht) | nebenläufig |
| Z22 | Der M2-Test bleibt unverändert grün (§6, Gemeinsamer Referenzdienst) | `go test` des M2-Tests im selben Lauf, vor und nach dem Slice; `git diff <Basis>..HEAD -- tests/integration/m2_test.go` leer | eine Ergänzung des Referenzdienstes ändert die Antwort auf einen Request des M2-Tests | andere |

Zählung (`SPEC-128` Nr. 3 g): *a* = 1 nebenläufig (Z21), *b* = 21 andere Zeilen.

**Stand beim Handoff** (Implementer; Fänger in `tests/integration/`; *Mutation* ist eine Änderung an einer Wegwerfkopie, die den Test aus dem benannten Grund rot färbt, einmal gesehen; *Mutator* wie oben). Vergleichsfunktion `vergleiche` in `vergleich_test.go`, Fälle und Veränderer in `protokolltreue_test.go`, Referenzdienst in `vergleichsdienst_test.go`.

| Nr. | Fänger (Test, Menge) | Mutation, rot gesehen |
|---|---|---|
| Z1 | `TestStatusAbweichung` (16 Fälle), `TestStatusNichtPhrase` | Statusvergleich entfernt; Vergleich liest die Phrase |
| Z2 | `TestBodyAbweichung` (11 Antwortarten × 3 Stellen = 33) | Vergleich nur der Länge; erstes Byte ausgenommen; letztes Byte ausgenommen |
| Z3 | `TestHeaderWertAbweichung` (68 Paare aus Fall und Header, der Orakel-Satz je Fall steht in `faelle()`) | jeder der 9 Namen der Fälle einzeln in der Ausnahmemenge (9 von 9) |
| Z4 | `TestHeaderNurEineSeite` (je relevanter Name ein Fall und je Fall ein hinzugefügter Name) | nur die Namen der direkten Seite; nur die Namen der Proxy-Seite |
| Z5 | `TestLeererWert`, `TestZeilenliste` (17 Formen der Zeilenliste, 17 × 17 = 289 Paare, gleich genau bei derselben Form) | leerer Wert wie fehlend behandelt; leere Werte aus einer Liste mit mehreren Zeilen entfernt |
| Z6 | `TestFeldzeilen`, `TestZeilenliste`, `TestSchreibweisenInEinerMap` (Zeilen zweier Schreibweisen in einer handgebauten Map in der Bytefolge der Schreibweisen) | Werte sortiert; Zeilen mit Komma zu einer Zeile geführt; `slices.Compact`; `TrimSpace` im Vergleich der Werte; Werte ohne Beachtung der Schreibweise; nur die erste Zeile; zweite Schreibweise überschreibt die erste |
| Z7 | `TestSchreibweiseDesNamens` (4 Namen × 9 Paare = 36, je gleich und mit anderem Wert) | Namen nicht kleingeschrieben |
| Z8 | `TestVolatileHeaders` (7 Namen × 9 Paare von Schreibweisen = 63; Orakel `volatileOrakel`, ausgeschrieben) | je ein Element aus dem Default von `VOLATILE_HEADERS` in `src/bootstrap/konfig_werte.go` entfernt (7 von 7) |
| Z9 | `TestHopByHopUndConnection` (9 Namen × 3 Schreibweisen = 27; 5 Formen von `Connection` × 3 Träger = 15) | die Liste verliert je ein Element (9 von 9); `Connection` nur von der direkten Seite; nur von der Proxy-Seite |
| Z10 | `TestCodingUndLaenge` (11 Antwortarten × 4 Ketten = 44, dazu handgebaut: Paare von Codings, `Content-Length` 10 gegen 11) | `Content-Encoding` zählt als relevant; `Content-Length` zählt als relevant |
| Z11 | `TestKetten` (25 + 5 = 30), `TestKettenSchreibweise` (4 Ketten × 4 Schreibweisen mit Randleerraum = 16), `TestMehrereGzipGlieder` (4 Formen mit 2 bis 3 Gliedern, auch leere) | Entpacken entfällt; Kette nicht umgekehrt; Dekoder auf den leeren Rohbody; Name des Codings nicht kleingeschrieben; Randleerraum des Glieds nicht entfernt; gzip ohne mehrere Glieder (`Multistream(false)`) |
| Z12 | `TestUmkodierenMitAenderung` (3 Antworten mit Coding × 3 Stellen = 9) | Vergleich prüft den Body nicht (nur die Entpackbarkeit) |
| Z13 | `TestOhneBody` (HEAD, 204, 304; handgebaut 204 und 304 mit Body auf jeder Seite) | Body bei 204 nicht verglichen; Body bei 304 nicht verglichen; Header ohne Body nicht verglichen |
| Z14 | `TestNichtEntpackbar` (19 Ketten und Ströme × 2 Seiten = 38; dazu 14 Ketten mit leerem Rohbody × 2 Seiten = 28, `SPEC-119` Nr. 6) | Fehler übergangen; nur der Status verglichen; leerer Rohbody vor der Prüfung der Kette als leer entschieden |
| Z15 | `TestNichtAngefragtesCoding` (3 Stellen im Rohbody, Wert von `Content-Encoding`) | Modus roh entpackt trotzdem; `Content-Encoding` im Modus roh ausgenommen; Fall ausgeschlossen |
| Z16 | `TestAcceptEncodingErreichtDenDienst`, `TestFaelleDirektWieErwartet` (Echo je Fall, beide Seiten) | Proxy ergänzt `identity` (`src/hexagon/services/record/header.go`); Client ohne `DisableCompression` |
| Z17 | `TestSelbstvergleich` (16 Fälle direkt gegen direkt, 302 auf beiden Seiten) | Dienst zählt `$count`; Client folgt der Weiterleitung |
| Z18 | `TestTextAntworten` (`$count`, `$value`) | `Content-Type` aus dem Vergleich genommen |
| Z19 | `TestAcceptEncodingNurBekannte` (Echo je Fall) | Client nennt zusätzlich `br` |
| Z20 | akzeptiertes Negativ, kein Fänger (Grund und Folge in `LH-QA-07.a` Nr. 7 a, b) | — |
| Z21 | `TestParallel` (8 Clients zugleich, alle Fälle, beide Seiten) | Zähler ohne Sperre im Referenzdienst: rot in 30 von 30 Prozessen |
| Z22 | `TestDurchstichM2` vor und nach dem Slice; `git diff 195afaa -- tests/integration/m2_test.go` leer | Zeile `Set-Cookie` aus `tests/integration/referenzdienst_test.go` entfernt |

*Raten.* Z21 ist der einzige nebenläufige Fänger: Test-Binary mit `-race` einmal gebaut, 30 Prozesse nacheinander mit `-test.count=1` (Image `odata-recorder:race`, Messender: der Implementer): mit der Mutation rot in 30 von 30 Prozessen, ohne sie in 0 von 30. Jede andere Zeile läuft in einem Prozess ohne Nebenläufigkeit und ist je Lauf deterministisch; `go test -count=20` über das Paket `tests/integration` ist grün (20 von 20).

*Echte Mutationen am Proxy* (zusätzlich zu den Mutatoren, `src/adapters/driving/httpodata/antwort.go`): leere Werte verworfen, Zeilen eines Namens umgekehrt, Status 200 als 201, nur die erste Zeile eines Namens — jede färbt `TestProtokolltreue` rot (4 von 4).

*Klassen als Schleife mit Orakel und Zählprüfung* (`AGENTS.md` §3.15, §3.17; Generator ist die Schleife, das Orakel steht ausgeschrieben im Test, die Zählprüfung nennt die Menge): Z2, Z3, Z7, Z8, Z9, Z10, Z11, Z12 und Z14, Mengen in der Tabelle. Orakel: die sieben Namen (`volatileOrakel`), die neun Hop-by-Hop-Namen, die Namen der relevanten Header je Fall (`relevant` in `faelle()`, gegen die Antwort des Dienstes geprüft), die Menge der 19 Ketten und Ströme (`kaputteKetten`).

*Zuordnung und Abweichungen vom Plan.* (1) Der Referenzdienst dieses Slice ist ein eigener Test-Server: der von M2 zählt `$count`, der Vergleich braucht zustandslose Antworten (Entscheidung Nr. 6 des Architect), das Echo von `Accept-Encoding` (Z16) steht in jeder Antwort als `X-Echo-Accept-Encoding`. (2) Z13: `net/http` sendet bei HEAD, 204 und 304 keinen Body, ein Mutator kann ihn nicht ausliefern; der Body nach 204 und 304 steht an handgebauten Antworten (beide Seiten), bei HEAD prüft dieselbe Funktion, die die Methode nicht kennt, den leeren Body (Negativ mit Grund und Folge in `LH-QA-07.a` Nr. 7 d). Der erste Lauf der Mutation „Body bei 304 nicht verglichen“ blieb grün; die handgebaute Schleife über 204 und 304 schließt die Lücke. (3) Z11 und Z14 prüfen die Ketten an handgebauten Antworten (die fünf Ketten und die Kette über zwei Zeilen); über den Draht laufen `gzip`, `deflate` und `deflate, gzip` (Fälle `gzip`, `deflate`, `kette`). (4) Z17, Z19: die Mutation ändert den Test-Server beziehungsweise die Fälle; kein Produktcode. (5) Z18: der Mutator des Plans (Typ, Wert) steckt im Test, die rot gesehene Mutation ist die Ausnahme von `content-type` aus dem Vergleich. (6) Fix-Runde nach dem Review: die Zeilenliste (`TestZeilenliste`, 17 Formen) und die Kettenklasse (Schreibweise, leerer Rohbody bei unbekannter Kette, mehrere gzip-Glieder) tragen Fänger über ihre Mengen; die Reihenfolge der Zeilen zweier Schreibweisen in einer handgebauten Map (`TestSchreibweisenInEinerMap`) ist eine Festlegung der Vergleichsfunktion, nicht des Produkts; auf dem Draht kommt sie nicht vor (Urteil des Architect, §6). Der Vergleich fand mit dem echten Recorder keine Abweichung in den 16 Antwortarten (§6, Risiko *Unterschied*: nicht eingetreten).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2`, `slice-m2-referenzdienst` und `slice-text-antworten` liegen in `done/` (Composition Root im Code; Referenzdienst und M2-Test im Code; Text-Antworten des Referenzdienstes im Code, Reihenfolge der Welle). Die Spezifikation trägt die Randformen aus §6 in `LH-QA-07.a`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Dieser Slice ist der letzte der Welle.

**Größenmarke** (`AGENTS.md` §3.14): über **2700** Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3 unter `tests/`, Tests eingeschlossen; `<Basis>` ist der Commit, der die Randformen vor dem Code abschließt; Text der Spezifikation zählt nicht mit). Grund der Ersetzung der vorläufigen Marke 1400 (§3.14): Sie entstand ohne Tabelle; die Formel des Gates (`SPEC-128` Nr. 1) ist die durchgesetzte, `make groessenmarke` meldete „Marke steht aus: Schätzung 1965, Marke 2700“, und die Marke des Blocks muss ihr gleichen. Der Konflikt mit `AGENTS.md` §3.14 (dort 1500) bleibt eine offene Entscheidung des Users (§6).

**Schätzung** (`AGENTS.md` §3.14, `SPEC-128`), nach der Tabelle in §3 (*a* = 1, *b* = 21, *f* = 1: der Referenzdienst mit den Zusatzfällen als Test-Server; *p* = 0, der Slice liefert keinen Produktcode; `pfad`: `tests/integration/`). Der Architect hat `pfad`, `produktcode` und `fakes` gesetzt, `basis`, `schaetzung` und `marke` hat der Planner gesetzt (`SPEC-128` Nr. 3 k): `basis` ist 7c4e1df, der Commit, der die Randformen vor dem Code abschließt.
- Nach `SPEC-128` Nr. 1: *S* = 85·1 + 80·21 + 200·1 + 0 = **1965**; Marke 100 · ⌈135·1965 / 10000⌉ = **2700** (135·1965 = 265275); keine Schnittfrage (*S* < 2223).
- Nach `AGENTS.md` §3.14 (85 je nebenläufige, 40 je andere Zeile, 200 je Test-Server, × 1,25): 85 + 840 + 200 = 1125, × 1,25 = 1406,25, Marke **1500**.
- Beide Fassungen widersprechen einander (offene Entscheidung des Users, §6); der Architect setzt keine Marke. Gesetzt ist die Marke des Gates, 2700; die vorläufige Marke 1400 ist ersetzt (oben).

**Messpunkt, Stand Fix-Runde nach dem Review** (`SPEC-128` Nr. 9; `git diff --numstat 7c4e1df..HEAD -- tests/`: 1141 + 251 + 125 = 1517; beim Handoff 1404, Tests eingeschlossen, Produktcode 0):
`Messpunkt slice-protokolltreue.md: zeilen=22 nebenlaeufig=1 fakes=1 produktcode=0 schaetzung=1965 marke=2700 messung=1517`; Auslastung 1517 ÷ 2700 = 0,56, die Marke wird nicht überschritten. Zum Vergleich ergibt `AGENTS.md` §3.14 mit 1125 vor und 1406 nach dem Faktor 1,25 eine Schätzung, die der Messung gleicht (1404).

```groessenmarke
pfad: tests/integration/
produktcode: 0
fakes: 1
basis: 7c4e1df4d8c6afebf83683c30bf7cc0ea5a4808e
schaetzung: 1965
marke: 2700
```


**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke (sie ist die Schätzung × 1,25 und schneidet nicht; vor dem Code wird nur über 3000 geschnitten, oben, `AGENTS.md` §3.14); offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten).
- `in-progress` → `open` (blockiert): der Vergleich findet einen Unterschied, den `LH-QA-07.a` nicht als zulässig nennt; das ist ein Befund an den Planner (der verantwortliche Slice liegt in `done/`), kein Test, der den Unterschied hinnimmt; ebenso eine Randform ohne Ort in `LH-QA-07.a` (an den Architect, `AGENTS.md` §3.12).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün und der Vergleichstest ist Teil von `make test`, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Nicht angefragtes Coding: der Referenzdienst sendet ein `Content-Encoding`, das der Client nicht genannt hat (etwa `br` auf `Accept-Encoding: gzip`); direkt kann der Test es nicht entpacken (`SPEC-119` Nr. 6), über den Proxy leitet der Record-Pfad die Antwort unverändert weiter — entschieden in `LH-QA-07.a` Nr. 4: eigener Fall mit eigener Erwartung (Rohbytes und `Content-Encoding` byte-gleich), Fänger Z15; Folge im Betrieb dort — **Ausgang:** entfallen bei Closure, wenn Z15 mit rot gesehener Mutation steht.
- Unterschied, den der Vergleich findet: ein Unterschied in Status, Body oder relevantem Header zwischen Direktzugriff und Proxy ist ein Befund gegen `LH-FA-01` in einem Slice, der dann in `done/` liegt — **Ausgang:** offen; Befund an den Planner (`AGENTS.md` §3.13), der über Behebung entscheidet; der Test nimmt ihn nicht hin. Stand Handoff: der Vergleich fand über den echten Recorder in keiner der 16 Antwortarten einen Unterschied.
- Gemeinsamer Referenzdienst: die Ergänzungen dieses Slice dürfen den M2-Test nicht ändern — **Ausgang:** offen; der M2-Test läuft vor und nach dem Slice unverändert grün, die Closure nennt beide Läufe. Stand Handoff: `TestDurchstichM2` grün auf 195afaa und auf dem Stand des Handoffs, kein Diff an den drei Dateien des M2-Dienstes und -Tests (§3, Z22).
- Nebenläufigkeit: Proxy, Referenzdienst und Client laufen in einem Prozess — **Ausgang:** offen; die nebenläufigen Fälle laufen im Ziel `make test-race` aus `slice-race-gate`: dessen Vertrag fährt alle Pakete der Stufe `source` (`SPEC-122` Nr. 2), und `tests/` liegt dort seit `slice-bootstrap-m2` (§1 dort, Sendung von `slice-race-gate`), der Vergleichstest läuft in der Stufe `test` (§1); ein Fänger mit Wahrscheinlichkeit, den der Detektor nicht ersetzt (ein Rennen ohne ungeordneten Zugriff, `SPEC-122` Nr. 8b), nennt seine Rate je Mutation und je Lauf von `go test` (`AGENTS.md` §3.10; Reviewer-Skill, MEDIUM). Stand Handoff: Z21 rot in 30 von 30 Prozessen mit der Mutation, 0 von 30 ohne (§3).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart*; jede Zelle trägt ihren Ort in `LH-QA-07.a` (Nr. 1 bis 7, vor dem ersten Code-Commit eingetragen) oder „gilt nicht“ mit Grund; eine Randform ohne Ort hält den Implementer an — **Ausgang:** offen (entfällt bei Closure, wenn keine Randform im Code entschieden wurde).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Antwort vergleichen (Header) | relevanter Name nur auf einer Seite: Abweichung, in beiden Richtungen (`LH-QA-07.a` Nr. 1; Fänger Z4) | leerer Wert ist ein Wert, leer ≠ fehlend (Nr. 1; Z5) | Name aus `Connection` einer der beiden Antworten: ausgenommen, Vereinigung beider (Nr. 2; Z9); mehrere Zeilen je Name in anderer Reihenfolge: Abweichung (Nr. 1; Z6); Name in anderer Schreibweise: derselbe Name, `VOLATILE_HEADERS` ohne Beachtung der Schreibweise (Nr. 2; Z7, Z8) | gilt nicht: der Vergleich liest nur vollständige Antworten (Nr. 7 a; Z20) | gilt nicht: ein Request je Seite; Zustandslosigkeit des Dienstes (Nr. 6; Z17); `VOLATILE_HEADERS` zwischen zwei Läufen verschieden: ausgenommen (Nr. 2; Z8) |
| Antwort vergleichen (Body) | Antwort ohne Body (`HEAD`, 204, 304) auf beiden Seiten ohne Body, Header gelten (Nr. 3; Z13) | leerer Rohbody mit `Content-Encoding`: leerer Body ohne Dekoder (Nr. 3; Z11) | nicht angefragtes Coding: eigener Fall, Rohbytes und `Content-Encoding` byte-gleich (Nr. 4; Z15); Kette mehrerer Codings: je Seite entpackt, bis zwei Glieder (Nr. 3; Z11); eine Kette, die der Test nicht entpackt: rot (Nr. 3; Z14) | Abbruch im Body: nicht verglichen, Grund und Folge (Nr. 7 a; Z20) | gilt nicht: jeder Fall sendet einen Request je Seite (Nr. 6; Z17) |
| Request senden | gilt nicht: der Test baut jeden Request selbst | fehlender `Accept-Encoding` ist zulässig und bleibt auch über den Proxy fehlend (Nr. 5; Z16) | Coding des Clients außerhalb von `gzip`, `deflate`, `identity`: ausgeschlossen (Nr. 7 c; Z19) | Upstream nicht erreichbar: nicht verglichen, Grund und Folge (Nr. 7 b; Z20) | gilt nicht: ein Request je Seite (Nr. 6; Z17) |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen:

| Paar | Entscheidung | Fänger |
|---|---|---|
| anderes Coding über den Proxy und ein geänderter Body | das Entpacken verdeckt den Unterschied nicht: beide Seiten werden je mit ihrer eigenen Kette entpackt und die Bytes verglichen (`LH-QA-07.a` Nr. 3) | Z12: Mutator ändert ein Byte und kodiert neu |
| Volatiler Header und relevanter Header gleichen Namens in anderer Schreibweise | gilt nicht als Paar: ein Name ist ausgenommen oder relevant, nie beides, die Schreibweise zählt nicht (Nr. 2) | Z7 und Z8: Schreibweise an der Vergleichsfunktion mit handgebauten Maps |
| Name nur auf einer Seite und in `Connection` der anderen genannt | ausgenommen: die Namen aus `Connection` beider Antworten bilden eine Menge (Nr. 2) | Z9: die Seite, die `Connection` trägt, wechselt |
| nicht angefragtes Coding und geänderter Body | Rohbytes werden verglichen, ein geändertes Byte ist rot (Nr. 4) | Z15: Mutator ändert ein Byte |

- Negativ HEAD (`AGENTS.md` §3.12, §3.13; Fund des Reviews): `net/http` des Recorders sendet bei HEAD, 204 und 304 keinen Body, ein Mutator kann ihn nicht ausliefern; der Body-Vergleich dort läuft für 204 und 304 an handgebauten Antworten, bei HEAD prüft dieselbe, methodenunabhängige Funktion den leeren Body. Folge im Betrieb: ein Proxy, der bei HEAD einen Body hielte, fiele dem Vergleich nicht auf. Grund und Folge stehen in `LH-QA-07.a` Nr. 7 d — **Ausgang:** entschieden (Architect). Ebenso entschieden: die Reihenfolge der Zeilen verschiedener Schreibweisen desselben Namens in einer handgebauten Map (`TestSchreibweisenInEinerMap`) bleibt Festlegung der Vergleichsfunktion und kommt nicht in Nr. 1, weil `net/http` je Name einen kanonischen Schlüssel liefert, der Fall auf dem Draht nicht vorkommt und kein Vertrag des Produkts daran hängt.
- Formelkonflikt der Größenmarke: `AGENTS.md` §3.14 (85 und 40 je Zeile, × 1,25) und `SPEC-128` Nr. 1 (85 und 80, × 1,35) ergeben für diese Tabelle 1500 und 2700 (§4) — **Ausgang:** offen; Adresse: der User entscheidet die Formel. Der Planner hat `basis`, `schaetzung` und `marke` (2700, Formel des Gates) gesetzt und die vorläufige Marke 1400 mit Grund ersetzt (§4); die Marke des Blocks bleibt die des Gates, solange `SPEC-128` gilt.

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle Einträge). Für diesen Slice zählen: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (6×, Ausgang `verkörpert` in `AGENTS.md` §3.10, samt Rate je Mutation und je Lauf) — der Test ist nebenläufig (§6, Risiko *Nebenläufigkeit*); `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (verkörpert in `AGENTS.md` §3.16) — „Status, Body und relevante Header unverändert“ ist eine Allaussage über alle Antwortarten und braucht ihren Fänger über alle Fälle des Referenzdienstes; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) — der Vergleich selbst trägt Mutationen (§2); `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16) — Raster in §6; `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×) — Testdaten mit Bytes ab `0x80` als Go-Escapes (§3); `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (verkörpert in `AGENTS.md` §3.14) — Marke und Schätzung in §4. Ohne Bezug: `BEO-REPO/akzeptiertes-negativ-ohne-folge` (2×; der Slice führt kein neues Negativ ein, außer der Architect entscheidet die Randform des nicht angefragten Codings als Ausschluss — dann mit Grund und Folge), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (2×) und `BEO-REPO/befund-an-rolle-ohne-artefakt` (1×).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
