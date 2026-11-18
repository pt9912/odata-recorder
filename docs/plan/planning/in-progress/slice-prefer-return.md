# Slice slice-prefer-return: Prefer return=minimal und return=representation für mutierende Requests

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-etag](../welle-etag.md).

**Bezug:** `LH-FA-07`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-07.b` (CRUD-Vertragsmatrix: `Prefer` für POST, PATCH, PUT; ohne `Prefer` gilt das Muss-Verhalten; ein nicht unterstützter Wert ergibt keinen Fehler) · `SPEC-285` bis `SPEC-289` (der Vertrag dieses Slice) · `SPEC-194` Nr. 4 (gilt nur noch ohne Wahl), `SPEC-192`, `SPEC-049` (Zeile Header, Spalte Soll: `Prefer: return=…`) · `SPEC-006` · `ARC-002`, `ARC-004`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-prefer-return)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Mock wertet den Header `Prefer` bei POST, PATCH und PUT aus: `return=minimal` ergibt eine Antwort ohne Entity im Body (Status und Header nach `SPEC-286`), `return=representation` eine Antwort mit der Entity im Body; die Antwort nennt den angewandten Wert im Header `Preference-Applied`. Ohne `Prefer` und bei jedem anderen Wert gilt das Muss-Verhalten ohne Fehler (POST `201` mit Entity, PATCH und PUT `204`). `Prefer` ändert den Zustand des Workspace nie: dieselbe Mutation ergibt mit und ohne Header denselben Zustand, und ein fehlgeschlagener Vorgang antwortet mit demselben Fehlerobjekt (`SPEC-287`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- ETag, `If-Match`, `If-None-Match` und `412` — `slice-etag-versionen` und `slice-if-match`; die beiden Äste sind unabhängig (`welle-etag` §5). Berührung ohne Kante: trägt eine Antwort mit Entity den ETag (aus `slice-etag-versionen`), prüft der Test des später liegenden Slice den Fall „`return=representation` mit ETag“.
- `Prefer`-Werte außer `return=minimal` und `return=representation` (etwa `odata.maxpagesize`, `respond-async`, `odata.include-annotations`) — Out-of-Scope der Welle (`welle-etag` §6); sie werden ohne Fehler ignoriert (`LH-FA-07.b`, `SPEC-285` Nr. 3).
- `Prefer` bei GET und DELETE — GET liefert immer die Entity, DELETE kennt keinen Body; der Header wird ohne Wirkung und ohne Fehler ignoriert (`SPEC-287` Nr. 3); der Slice liefert dafür nur den Fänger der Wirkung.
- `Prefer` im Record und Replay — der Header bleibt Teil des Exchange und geht als Vary-Header in den Match-Key (`SPEC-061`); der Mock-Pfad ist der einzige, der ihn auswertet (`SPEC-287` Nr. 6).

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

- [x] `return=minimal` und `return=representation` für POST, PATCH und PUT: Status, Body, `Location` und `OData-EntityId`, `Preference-Applied` nach `SPEC-286`; ohne `Prefer` unverändert das Muss-Verhalten.
- [x] Randformen des Headers (Groß- und Kleinschreibung, Leerraum, mehrere Präferenzen in einem Header, mehrere Header, widersprüchliche Werte, leerer oder unbekannter Wert): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; ein nicht unterstützter Wert ergibt keinen Fehler.
- [x] Unveränderter Zustand: mit und ohne `Prefer` derselbe Zustand des Workspace und dieselbe Revision; bei fehlgeschlagener Mutation dasselbe Fehlerobjekt.
- [x] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
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
| `src/hexagon/services/mock/` | update | Wahl der Antwort nach der Präferenz (`SPEC-286`, `SPEC-288`), Aufrufe des Schreibers je Antwort: `src/hexagon/services/mock/schreiben_antwort.go` (neu: `antwort`, `antwortMitEntity`, `antwortOhneBody`, `adresse`, `kopf204`; aus `schreiben.go` herausgelöst), `mutation.kopf` trägt die Header-Map bis zur Antwort |
| `src/hexagon/model/` | update | `src/hexagon/model/rueckgabe.go` (neu): `Rueckgabe` und `LiesRueckgabe` (`SPEC-285`); die angewandte Präferenz steht als Header in der Antwort, kein neuer Typ am Port |
| `src/adapters/driving/httpodata/` | prüfen | Der Adapter reicht alle Kopfzeilen `Prefer` durch und schreibt `Preference-Applied`, `OData-EntityId` und Status `200` unverändert (`SPEC-286`); der Code ändert sich nicht, `src/adapters/driving/httpodata/prefer_test.go` (neu) fängt beide Richtungen |
| `src/bootstrap/prefer_ende_zu_ende_test.go` | neu | Zeile 18 mit dem echten Schreiber, der echten Eingabe und der SQLite-Transaktion: Body gleich `GET` danach; liegt außerhalb der Pfade der Größenmarke (§4) |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect, `SPEC-285` bis `SPEC-289`) |
| `src/hexagon/services/mock/schreiben_antwort_test.go` | update | `TestPreferIgnoriert` sagt zu, dass `Prefer` nie gelesen wird (`SPEC-194` Nr. 4 alt); mit `SPEC-286` gilt das nur ohne Wahl, der Test wird auf die Fälle ohne Wahl begrenzt, die Wahlen fängt die Tabelle (Zeilen 10 bis 22) |
| `src/hexagon/services/hybrid/prefer_test.go` | neu | Fall des Hybrid-Dienstes (Zeile 24); der vorhandene Fake `stufe` des Replay-Fallbacks hält die Header fest, es entsteht kein neuer |
| `src/hexagon/model/rueckgabe_test.go`, `src/hexagon/services/mock/prefer_test.go` | neu | Fälle von `LH-FA-07.b` je Methode und je Wert; Generatoren über die Klassen der Header-Formen (Zeilen 1 bis 9) und der Antworten (Zeilen 10 bis 23); die vorhandenen Fakes (`schreiber`, `prot`) halten jedes Argument fest und zählen jede Operation |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Auftrag an den Architect** — erledigt: der Vertrag steht in `SPEC-285` (Grammatik und Wahl), `SPEC-286` (Antwort je Methode und Wahl, `Vary`), `SPEC-287` (GET, DELETE, Fehler, Zustand, Modi), `SPEC-288` (Herkunft der Entity, Aufrufe des Schreibers; der Schreiber aus `SPEC-168` genügt unverändert), `SPEC-289` (Nebenläufigkeit, Ungeprüftes). Die vier Punkte des Auftrags: (1) Antwort je Methode und Wahl `SPEC-286` Nr. 1 und 2; (2) Grammatik `SPEC-285` Nr. 2 bis 4; (3) GET, DELETE und Fehler `SPEC-287` Nr. 2 und 3; (4) Entity aus dem geschriebenen Eintrag `SPEC-288` Nr. 1 und 2.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Mengen als Schleife oder Generator aus dem Wortlaut der Zusage, `AGENTS.md` §3.17; die Mutation je Zeile fährt der Implementer selbst rot und trägt die Raten und den Bericht in diesen Absatz ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | `LiesRueckgabe` liefert für jede Zeile der Beispiele von `SPEC-285` Nr. 4 das dort genannte Ergebnis | Tabellentest über alle Zeilen der Tabelle, Werte aus Bytewerten gebaut (`AGENTS.md` §3.18), Datei gegengelesen | Vergleich der Werte byte-genau statt ohne Beachtung der Schreibweise | andere |
| 2 | Name und Wert ohne Beachtung der ASCII-Schreibweise (Nr. 3) | Schleife über alle 64 Schreibweisen von `return`, alle 128 von `minimal` und alle 16384 von `representation` | nur Kleinschreibung des Namens, nur Kleinschreibung eines Wertes | andere |
| 3 | Leerraum ist genau Leerzeichen und Tabulator, an jeder Stelle (Nr. 2) | Schleife über die Stellen (vor dem Namen, vor und nach `=`, nach dem Wort, um `,`, vor `;`) × die Zeichen `0x00` bis `0xFF` (Orakel: `0x20` und `0x09` sind Leerraum, jedes andere Byte nicht) | `strings.TrimSpace` statt `OWS` (fängt `0x0B`, `0x0C`, `0xA0`, NBSP als `0xC2 0xA0`); nur Leerzeichen als `OWS` | andere |
| 4 | Die erste lesbare Instanz `return` entscheidet; spätere, auch widersprüchliche, nie (Nr. 3) | Schleife über alle Folgen der Länge 0 bis 3 aus den neun Elementen {`return=minimal`, `return=representation`, `return=foo`, `return`, `return=`, `respond-async`, `odata.x=1`, `return@x`, leer} × jede Verteilung auf Kopfzeilen (eine Zeile, je Element eine Zeile, jede Teilung); Orakel: erstes lesbares `return` | letzte Instanz statt erster; erste unterstützte statt erster; nur die erste Kopfzeile gelesen | andere |
| 5 | Mehrere Kopfzeilen und mehrere Schreibweisen des Namens werden in der Bytefolge der Namen, dann der Reihenfolge nach gelesen (Nr. 1) | Schleife über Schreibweisen des Namens (`Prefer`, `prefer`, `PREFER`, Paare davon) mit verschiedenen Werten; Orakel nach der Bytefolge | nur `Prefer` gelesen; Reihenfolge nach Eingang statt nach Bytefolge | andere |
| 6 | Das Komma in einem Paar trennt nicht; ein offenes Paar reicht bis zum Ende der Zeile, nie in die nächste (Nr. 2) | Fälle `x="a, return=representation"`, `return="minimal`, `x="` gefolgt von einer zweiten Zeile `return=minimal`, Backslash vor dem Schlusszeichen; dazu Schleife über die Stellen eines Paares im Element | Trennung an jedem `,`; Zeilen vor dem Lesen zusammengefügt | andere |
| 6a | Ein Backslash im Wort in Anführungszeichen ersetzt sich durch das folgende Zeichen, der Wert gilt nach dem Auflösen; ein Backslash am Ende hält das Paar offen (Nr. 2, 3) | `TestLiesRueckgabeBackslashWort`: die Klasse „Backslash im Wort“ als Generator aus dem Wortlaut (jede Stelle des Wortes, alle Zeichen, sieben eingefügte und drei ersetzende Zeichen, zwei offene Formen) × 3 Wörter × 3 Zusammenhänge; Orakel: Auflösen im Test (`loeseAuf`, ohne Code des Dienstes) und der Orakel-Leser aus Zeile 9 | Backslash im Wort ohne Wirkung (R15); Zeichen nach dem Backslash gestrichen (B1) oder der Backslash bleibt (B2); nur an der ersten Stelle (B4); nicht vor `n` (B6); Backslash am Ende der Zeile schließt das Paar (B8) | andere |
| 6b | Ein Backslash in einem Paar eines fremden Elements schützt das folgende Zeichen: eine Folge von *k* Backslashes vor dem Schlusszeichen schützt es bei ungeradem, lässt es bei geradem *k* schließen; ein Backslash außerhalb eines Paares hat keine Wirkung, auch nicht vor `,` oder `"` (Nr. 2) | `TestLiesRueckgabeBackslashFremd`: die Klasse „Backslash in einem fremden Element“ an beiden Stellen der Grammatik, die ihn nennen; im Paar Stelle des Paares 4 × Inhalt davor 3 × *k* 0 bis 7 × Zeichen hinter der Folge (`"`, `,`, `a`, Ende der Zeile) 4 = 384, außerhalb Stelle der Folge 4 × Zeichen dahinter (`,`, Paar) 2 × *k* 0 bis 7 = 64; Orakel: Parität von *k* im Test und der Orakel-Leser aus Zeile 9 | Backslash im Paar ohne Wirkung (m06c); schützt nur vor `"` (EA1); schützt nicht vor einem Backslash (EA3); schützt nicht nach einem Backslash (EA4); schützt nur am Anfang des Paares (EA5); schützt zwei Zeichen (EA6); schützt nicht nach drei Backslashes (EA7); ein Backslash außerhalb eines Paares schützt das folgende Zeichen (EA2) oder nur vor `,` (EA8) | andere |
| 7 | `Rest` ab dem ersten `;` hat keine Wirkung (Nr. 2) | Schleife über Formen des Rests (`;`, `; a`, `; a=b`, `;;`, `; =`, `; "`, `; ,x`) × beide Werte; das Ergebnis gleicht dem ohne Rest (ein Komma im Rest außerhalb eines Paares beginnt ein neues Element, das der Orakel-Fall von Zeile 4 deckt) | Rest als Fehler gelesen (`Keine`) | andere |
| 8 | Ein unlesbares Element zählt nicht, ein lesbares mit unbrauchbarem Wert zählt (Nr. 2, 3) | Schleife über jedes Byte `0x00` bis `0xFF` als erstes Zeichen des Namens, als Zeichen direkt hinter `return` und als Zeichen hinter dem Wort; Orakel nach `tchar`, `OWS`, `=`, `;`, `,`, `"` | Name mit einem Zeichen außerhalb von `tchar` als lesbar behandelt; Wert mit Rest hinter dem Wort als `Minimal` | andere |
| 9 | `LiesRueckgabe` ändert die Eingabe nie, liefert nie einen Fehler und löst keine Panik aus (Nr. 1) | Schleife über alle Folgen der Länge 0 bis 6 über dem Alphabet {`return`, `=`, `;`, `"`, Backslash, `,`, Leerzeichen, Tabulator, `minimal`, `x`, `0xA0`}, die Eingabe vorher kopiert und danach verglichen | Sortieren oder Kürzen der Eingabe an Ort und Stelle; Indexzugriff hinter dem Ende | andere |
| 9a | Die Zeit von `LiesRueckgabe` wächst linear mit der Länge der Zeilen (Nr. 1) | `TestLiesRueckgabeZeit`: die Stellen des langen Laufs in der Grammatik als Schleife über 20 Formen (Zeilen, Elemente, Kommas, Leerraum vor dem Namen, hinter dem Namen, hinter `=`, hinter dem Wort, Name, Wert als Token, Wort in Anführungszeichen ohne Backslash, mit Backslash, offen, Rest hinter dem Wert, Wert, Rest und Paar eines fremden Elements ohne Backslash, mit Backslash, mit Folge von Backslashes, offen, Backslash außerhalb eines Paares) mit 2000 und mit 64000 Wiederholungen; bei linearem Wachstum das 32-fache, Grenze das 128-fache der besten Zeit der kleinen Länge, bestanden mit einem von fünf Versuchen; ein Zufallsfänger (Rate in der Tabelle der Mutationen) | je Element eine Kopie der Zeile (Q1); je Zeichen im Wort mit Backslash eine Kopie (Q2); je Element erneutes Lesen der Zeile ohne Zuweisung (Q3); quadratisch im Namen (Q5), im Wert als Token (QW), im Wort ohne Backslash (QG), im Lauf eines fremden Elements (QR), im fremden Paar (QP), bei Backslash im fremden Paar (QB), im Leerraum hinter `=` (QL2) und hinter dem Wort (QL3) | andere |
| 10 | Die Antwort je Methode und Wahl trägt Status, Body und genau die Header der Tabelle von `SPEC-286` Nr. 1, nie `Vary` | Schleife über 3 Methoden × 3 Wahlen × `Dienst.Etag` an und aus × `Accept` mit den Stufen `none` und `minimal`: Menge der Header-Namen gleich der erwarteten (Gleichheit, nicht Enthaltensein), Status, Body leer oder nicht, der Body ist nie `nil` (`preferSende` meldet jede Antwort ohne Fehler mit `Body` nil, auch bei `DELETE`, `GET`, `HEAD`; im Test mit echtem Schreiber `TestPreferMinimalEntityId`) | je Zelle der Tabelle eine Mutation (`POST` · Minimal als `201`; `PATCH` · Entity als `204`; `Content-Type` an der `204`; `Vary: Prefer`; `ETag` bei aus); `Body` nil an jeder Zelle der `204` (sechs Zellen und `DELETE`) | andere |
| 11 | `Location` und `OData-EntityId` tragen dieselbe Zeichenfolge, aus `Adresse`, nie aus `Host` (`SPEC-286` Nr. 1) | Schleife über Methoden und Wahlen mit Header `Host` und `X-Forwarded-Host` fremd; der Fake des Schreibers hält `Wurzel` und `Schluessel` jedes Aufrufs fest | `Wurzel` aus dem Header `Host`; `OData-EntityId` aus `Location` abgeleitet und gekürzt | andere |
| 12 | `Preference-Applied` ist genau `return=minimal` oder `return=representation`, in einer Kopfzeile, ohne Echo (`SPEC-286` Nr. 1) | Schleife über die Formen der Zeilen von Fänger 1 bis 5 mit Erfolg: Anzahl der Kopfzeilen, Wert byte-gleich | Echo des Client-Werts; die Schreibweise des Clients; zweite Kopfzeile je weitere Präferenz | andere |
| 13 | `Preference-Applied` steht nur an der Antwort einer gelungenen Mutation mit Wahl (`SPEC-286` Nr. 3) | Schleife über 5 Methoden × 3 Wahlen × {Erfolg, jeder Fehlercode, den die Mutation erreicht} (`ODR-E6002`, `E6003`, `E6005`, `E6007`, `E6008`, `E6009`, `E6010` bei `Etag` an, `E9001` aus dem Schreiber) | Header vor der Entscheidung gesetzt; an `GET`, `DELETE` gesetzt | andere |
| 14 | Eine gescheiterte Mutation antwortet mit und ohne `Prefer` gleich, ohne `Preference-Applied` und ohne `OData-EntityId`; Zustand und Revision bleiben (`SPEC-287` Nr. 2) | Schleife über jeden Fehlerfall von Fänger 13 × jede Form der Zeilen aus 4, 5 und 7: Fehlerobjekt, Status und Header byte-gleich dem Lauf ohne Header, Zustand und Revision unverändert | `return=minimal` ergibt bei einem Fehler `204`; `Preference-Applied` an der Fehlerantwort; ein Wert aus `Prefer` im Text des Fehlers | andere |
| 15 | `Prefer` ändert nie den Zustand (`SPEC-287` Nr. 4) | Schleife über 3 Methoden × alle Formen aus 4, 5 und 7: Eintrag (`Payload`, `KnownProperties`, `Version`, `Completeness`), Revision und die Zahl jeder Operation der Fakes (`Transaktion` 1, `Put` 1, `Advance` 1; der Fake hält `Put` fest) gleich dem Lauf ohne Header | `Put` mit anderer `Version` bei `Entity`; ein zweites `Entities.Get` nach `Put`; `Advance` ausgelassen bei `Minimal` | andere |
| 16 | `GET`, `HEAD` und `DELETE` ignorieren `Prefer` ohne Wirkung und ohne Fehler (`SPEC-287` Nr. 3) | Schleife über die drei Methoden × alle Formen aus 4, 5 und 7: Status, Header und Body byte-gleich dem Lauf ohne Header; `DELETE` bei `return=representation` `204` | `DELETE` bei `return=representation` als `200` mit Entity; `GET` bei `return=minimal` ohne Body | andere |
| 17 | Der Header wird vor Schritt (l) nicht gelesen, und kein Fehler hängt an ihm (`SPEC-287` Nr. 1) | Wirkungs-Fänger über Fänger 14; der Zugriff selbst ist ein akzeptiertes Negativ (`SPEC-289` Nr. 3) | Auswerten vor Schritt (b) mit Fehler bei unbekanntem Wert | andere |
| 18 | Die Entity der Antwort ist der geschriebene Eintrag, nicht der Body der Anfrage (`SPEC-288` Nr. 1) | Schleife über `PATCH` mit Teilbody (die ungenannten Properties stehen in der Antwort), `PATCH` mit `{}`, `PUT` mit fehlender nullbarer Property (`null` steht in der Antwort), `POST` × `Accept` `none` und `minimal`: der Body gleicht dem `GET` danach (bei `Etag` an mit `@odata.etag` der neuen Version) | Entity aus `EntityEingabe` gebildet; `Etag` der Version vor der Mutation; Entity nach einem zweiten `Entities.Get` | andere |
| 19 | Die Aufrufe des Schreibers je Antwort sind die Zahlen von `SPEC-288` Nr. 3 | Der Fake des Schreibers zählt `SchreibeEntity`, `Adresse`, `SchreibeAnzahl` je Test; Schleife über die fünf Zeilen der Tabelle × beide Werte von `Dienst.Etag` | `Adresse` bei `PATCH` ohne Wahl; `SchreibeEntity` bei `POST` · Minimal; `SchreibeAnzahl` bei `PATCH` · Entity | andere |
| 20 | Jeder Aufruf des Schreibers erhält `ctx` unverändert, `Accept` und `Wurzel`; `Auswahl` ist `nil`, auch mit `$select` (`SPEC-288` Nr. 2, 3) | Der Fake hält Argumente jedes Aufrufs fest; Schleife über die Stellen S24, S25, S26 × die Antworten, die sie rufen × Anfragen mit und ohne `$select` und mit `Accept` | `Auswahl` aus `$select`; `context.Background()` statt `ctx`; `Accept` nicht weitergereicht | andere |
| 21 | Ein Fehler eines Aufrufs des Schreibers ist `ODR-E9001` mit Rollback, ein Kontextfehler bleibt ohne Code, der Zustand bleibt (`SPEC-288` Nr. 3, `SPEC-287` Nr. 5) | Schleife über die Stellen, die die Antwort ruft (Fänger 19) × {`ErrUngueltig`, anderer Fehler, Kontextfehler, `ErrNichtGefunden`, `ErrBereitsVorhanden`}: Code, Zustand, Revision; der Gegenstand der Abbildung ist je Stelle `Sonstiges` (`SPEC-195` Nr. 11): die beiden letzten Fehler sind nur dort `ODR-E9001`, und der Test zählt, dass jeder der sieben anderen Gegenstände von einem Fehler der Liste getrennt wird | Fehler verschluckt (Antwort ohne Entity); Rollback ausgelassen; ein anderer Gegenstand an S24, S25, S26 (sieben Gegenstände × drei Stellen) | andere |
| 22 | Ein leeres Kennzeichen oder ein leerer `ContentType` der `Darstellung` ist `ODR-E9001` ohne Antwort (`SPEC-288` Nr. 2) | Schleife über die Antworten mit Aufruf von `SchreibeEntity` und von `SchreibeAnzahl` × {leeres Kennzeichen, leerer Name, leerer Wert, leerer `ContentType`} | Header gesetzt trotz leerem Wert | andere |
| 23 | Von *n* gleichzeitigen `PATCH` mit `Entity` trägt jede Antwort den Eintrag ihrer Mutation (`SPEC-289` Nr. 2) | Wirkungs-Fänger: zählt die `@odata.etag` der Bodys (genau `EtagFuer(v+1)` bis `EtagFuer(v+n)`, jede einmal), `-count=20`; die Zusage ist eine Wirkung, nicht ein Datenrennen | zweites Lesen nach `Put` (zwei Antworten mit derselben Version); gemeinsame Variable für die Entity | nebenläufig |
| 24 | Im Modus `hybrid` geht `Prefer` unverändert in den Fallback, und der Dienst setzt oder streicht `Preference-Applied` nie (`SPEC-287` Nr. 6) | Fall im Test des Hybrid-Dienstes: der Mock endet mit einem freigegebenen Code, der Fake der Stufe hält die Header fest; Schleife über die Formen aus 4 und 5 × die acht Formen von `vorbereiten` (ohne Body, Laenge 0 mit Leser, gepuffert bei deklarierter Laenge und bei -1, nicht fallback-fähig, Laenge über der Grenze, Body über der Grenze bei Laenge -1, Lesefehler) in `TestPreferGehtUnveraendertInDenFallback`; wo kein Fallback bleibt, erhält der Mock den Header und der Fehler des Mock kommt zum Client; der Upstream (Weg nach M und nach R) in `TestPreferGehtUnveraendertInDenUpstream`; die Antwort byte-gleich der der Stufe | `Prefer` aus den Headern des Fallbacks genommen; `Preference-Applied` angefügt; der Header in einem Zweig von `vorbereiten` an einer Stufe (R, M) gestrichen; `Prefer` in `upstreamHeader` gefiltert (U1, U2) | andere |
| 25 | Der HTTP-Adapter reicht jede Kopfzeile `Prefer` in der Reihenfolge und mit jeder Schreibweise des Namens durch und schreibt Status `200` und `204` mit den Headern der Antwort unverändert | Test am HTTP-Eingang: Fake des Dienstes hält `Anfrage.Header` fest; Schleife über die Zahl der Zeilen (0 bis 3) und die Schreibweisen des Namens; Gegenrichtung: Fake liefert `Preference-Applied`, `OData-EntityId`, `Location`, `200` und `204`, der Adapter schreibt sie byte-gleich | nur die erste Kopfzeile durchgereicht; `OData-EntityId` gefiltert; `200` auf `201` gesetzt | andere |
| 26 | gilt nicht: das Handbuch beschreibt `Prefer` nicht (`SPEC-200` Nr. 1) | gilt nicht, Grund: `Prefer` ist kein Schlüssel der Einstellungen und kein Kommando, also keine Aussage der Klassen K1 bis K4; das HTTP-Verhalten tragen `SPEC-285` bis `SPEC-289` und die Tests dieser Tabelle (Entscheidung des Architect, §6) | gilt nicht | andere |
| 27 | Dieselbe Mutation zweimal sind zwei Mutationen, und die zweite Antwort mit Entity zeigt den Eintrag der zweiten (`SPEC-287` Nr. 7) | `TestPreferWiederholungGleichGet` (bootstrap, echter Schreiber und SQLite): Methode (`PATCH`, `PUT`) × Etag 2 × Mutation 2 = 8; Body gleich dem `GET` danach und mit dem Namen der eigenen Mutation; die Aufrufe des Dienstes liegen bei Zeile 15, `TestWiederholungMutation` (Bestand) trägt die Wirkung ohne Wahl | Antwort mit dem Body der vorigen Mutation (`staleBodies`, Paketvariable, E4) | andere |
| 28 | Dieselbe Anfrage gegen denselben Zustand ergibt dieselben Bytes und Header; der Dienst hält keinen Zustand für die Wahl und liest keine Uhr (`SPEC-289` Nr. 1); `LiesRueckgabe` ist rein | `TestPaketOhneZustand` (mock, `SPEC-195` Nr. 12: keine Paketvariable) für den Zustand, die ETag-Tests (`TestEtagWiederholung`) für Bytes und Header bei `Dienst.Etag` wahr, `TestLiesRueckgabeFolgen` für die Reinheit (dieselbe Zeile allein und zwischen fremden Zeilen, Eingabe unverändert) | Paketvariable für die letzte Wahl; Antwort aus einem Zwischenspeicher | andere |

Die Spalte *Klasse* trägt die Klasse nach `SPEC-128` Nr. 3 e (`andere`, nur Zeile 23 `nebenläufig`); die Art des Fängers (Tabelle, Generator, Schleife, Wirkung, Negativ, Gate) steht in der Spalte *Fänger*.

Akzeptierte Negative (Grund und Folge stehen in `SPEC-285` *Ungeprüft* und `SPEC-289` Nr. 3, nicht hier wiederholt): der Zugriff auf `Prefer` bei `GET`, `HEAD`, `DELETE` und bei Fehlern (Zeile 17, geprüft wird die Wirkung); die Länge der Zeile über die Grenze des Adapters; andere Präferenzen als `return`; ein Cache zwischen Client und Mock (kein `Vary`, `SPEC-286` Nr. 2). Die Prüfgrenze der Generatoren (Zeilen 2 bis 9) nennt der Bericht mit Grund.

**Mengen der Generatoren** (`AGENTS.md` §3.17; Orakel stehen im Test und teilen keinen Code mit dem Dienst; die Zählprüfung der Menge steht im Test):

| Zeile | Menge | Generator | Orakel | Elemente |
|---|---|---|---|---|
| 1 | Beispiele von `SPEC-285` Nr. 4 und fremde Zeichen | Tabelle, Bytes aus Bytewerten gebaut | die Tabelle des Vertrags | 36 |
| 2 | Schreibweise von `return`, `minimal`, `representation` | alle 2^n Schreibweisen | Ergebnis je Wort | 64·2 + 128 + 16384 (der Wert auch in Anführungszeichen) |
| 3 | Byte an sieben Stellen; Leerraum außerhalb von ASCII | Byte 0x00 bis 0xFF × Stelle; jede Rune von `unicode.IsSpace` ab U+0080 × vier Stellen | Menge der Bytes je Stelle, die `Minimal` ergeben | 7·256 und 19·4 |
| 4 | Folgen von Elementen auf Kopfzeilen | Länge 0 bis 3 über 9 Elemente × Verteilung auf Zeilen | erstes entscheidendes Element | 3088 |
| 5 | Namen der Kopfzeilen | geordnetes Paar der 64 Schreibweisen von `Prefer`; Teilmengen von `PREFER`, `Prefer`, `prefer` × 5 Zeilenformen | kleinere Bytefolge des Namens zuerst | 4032 und 124 |
| 6 | Paare in Anführungszeichen | 6 Inhalte × 5 Stellen × {ohne, mit folgendem Element}; 5 offene Paare × 2 Formen; 3 Backslash-Fälle | ein Paar hält jedes Komma, ein offenes reicht bis zum Zeilenende | 73 |
| 6a | Backslash im Wort in Anführungszeichen | 3 Wörter (`minimal`, `representation`, `MiNiMaL`) × [Stelle geschützt *L*, alle geschützt 1, eingefügt 7·(*L*+1), ersetzt 3·*L*, offen 2 = 11·*L*+10] × Zusammenhang 3 (allein, mit Rest, mit folgendem Element) | Auflösen im Test und Orakel-Leser aus Zeile 9 | 3 · 338 = 1014 |
| 7 | Rest hinter `;` | 9 Formen × 2 Werte × {ohne, mit späterem `return`} | Ergebnis gleich dem ohne Rest | 36 |
| 8 | Byte an drei Stellen des Elements | Byte 0x00 bis 0xFF × Stelle, mit folgendem Element | Wahl je Byte und Stelle | 768 |
| 9 | Folgen über dem Alphabet der Zeile 9; zusätzlich mit `representation` | Länge 0 bis 6 über 11 Symbole (Eingabe und Gleichheit allein oder zwischen zwei Zeilen); Länge 0 bis 5 über 12 Symbole gegen ein Orakel aus regulären Ausdrücken | Gleichheit der Eingabe; Ergebnis des Orakels | 1.948.717 und 271.453 |
| 6b | Backslash in einem fremden Element | im Paar: Stelle des Paares 4 × Inhalt davor 3 × *k* 0 bis 7 × Zeichen hinter der Folge 4; außerhalb: Stelle 4 × Zeichen dahinter 2 × *k* 0 bis 7 | Parität von *k*; Orakel-Leser aus Zeile 9 | 384 und 64 |
| 9a | Stellen des langen Laufs in der Grammatik | 20 Formen (Zeilen, Elemente, Kommas, Leerraum an vier Stellen, Name, Wert als Token, Wort ohne Backslash, mit Backslash, offen, Rest hinter dem Wert, fremder Wert, fremder Rest, fremdes Paar ohne Backslash, mit Backslash, mit Folge von Backslashes, offen, Backslash außerhalb) × 2 Längen (2000, 64000 Wiederholungen) | Wachstum höchstens das 128-fache bei dem 32-fachen der Länge | 20 Formen |
| 10 | Antwort je Methode, Wahl, `Dienst.Etag`, `Accept` | Schleife über alle vier Faktoren | die Tabelle von `SPEC-286` Nr. 1 | 36 |
| 11, 19, 20 | Aufrufe des Schreibers | Methode × Wahl (× `$select` × `Accept`) | die Tabelle von `SPEC-288` Nr. 3 | 9, 24, 36 |
| 12 | Formen von `Prefer` am Dienst | Länge 0 bis 2 über 11 Elemente × Verteilung auf Zeilen × 3 Methoden | erstes entscheidendes Element | 254 · 3 |
| 13 | Antwort ohne `Preference-Applied` | Methode 5 × Wahl 3 × {Erfolg, jeder Fehlerfall, den die Methode erreicht (9 Fälle, 28 Paare)} | kein Header bei Fehler, `GET`, `DELETE` | 102 |
| 14, 15, 16, 17 | Formen von `Prefer` gegen den Lauf ohne Header | 254 Formen und 4 mit Wert, der in keinem Text stehen darf × (Fehlerfall × Methode, Methode × `Dienst.Etag`, `GET`, `HEAD`, `DELETE`) | der Lauf ohne Header | 5418, 1548, 774 |
| 18 | Entity der Antwort gegen `GET` danach | 6 Fälle × `Accept` 3 × `Dienst.Etag` 2, echter Schreiber und SQLite | Body und ETag des `GET` | 36 (plus 6 mit `return=minimal`) |
| 21, 22 | Aufrufstellen und Fehler | 14 Stellen × 7 Fehler (davon zwei, die den Gegenstand trennen: jeder der sieben Gegenstände außer `Sonstiges` wird von einem Fehler der Liste getrennt, gezählt); 9 Stellen mit Darstellung × 4 Verstöße | Code, Zustand, Revision, `ctx` | 98 und 36 |
| 23 | gleichzeitige `PATCH` | *n* ∈ {2, 8, 32} × 20 frische Aufbauten | `EtagFuer(v+1)` bis `EtagFuer(v+n)`, jeder einmal, in Body und Kopfzeile | 60 Aufbauten |
| 24 | Formen am Hybrid-Dienst | Form von `vorbereiten` 8 × 172 Formen × Name 2 × Methode 5 × Header der Stufe 2; Upstream: Weg × Methode (5) × Body 3 × 172 Formen × Name 2 | die Stufe erhält jede Zeile, der Dienst setzt und streicht nichts | 27520 und 5160 |
| 27 | Antwort der zweiten gleichen Mutation | Methode (`PATCH`, `PUT`) × Etag 2 × Mutation 2, echter Schreiber und SQLite | Body gleich dem `GET` danach, mit dem Namen der eigenen Mutation | 8 |
| 25 | Zeilen am HTTP-Eingang | Zahl der Zeilen 0 bis 3 × Schreibweise je Zeile × Methode 3 × Bindung 2; Status 3 × Wert 2 × Methode 3 × Bindung 2 | Zeilen in der Reihenfolge; Antwort byte-gleich | 240 und 36 |

**Prüfgrenzen** (akzeptierte Negative, `SPEC-289` Nr. 3): die Menge der Zeichenfolgen von `Prefer` ist unbeschränkt; geprüft sind Folgen bis zu drei Elementen (Zeile 4), sechs Symbolen (Zeile 9, Orakel bis fünf) und je Byte 0x00 bis 0xFF an den Stellen von `SPEC-285` Nr. 2. Folge im Betrieb: eine Form mit mehr Elementen oder Symbolen bleibt ungeprüft; die Funktion hinterlässt keinen Zustand. **Zeit (Zeile 9a):** die Achse, die die Formen trennt, ist die Stelle des langen Laufs in der Grammatik von `SPEC-285` Nr. 2, und sie ist endlich: 20 Stellen, je bei zwei Längen mit einem Baustein; jede Stelle, an der ein Lauf von Zeichen stehen kann, ist geführt, und QW (Wert als Token), QG (Wort ohne Backslash), QR (Lauf eines fremden Elements), QP, QB, QL2 und QL3 sind je im Zeit-Test selbst rot gesehen. Unbeschränkt ist der Baustein je Stelle (jede Folge von Zeichen): ein Lauf, der erst bei einem anderen Zeichen oder einer anderen Folge quadratisch liest, bleibt ungeprüft. Grund: die Menge der Bausteine ist unbeschränkt, und ein Zähler im Code gäbe es nur um den Preis eines Hakens im Produktcode. Folge im Betrieb: eine Zeile von der Länge, die der HTTP-Adapter zulässt (`SPEC-136`), kostet bei einem solchen Baustein mehr als lineare Zeit je Mutation; `SPEC-285` *Ungeprüft* nennt diesen Rest. **Menge `tchar`:** sie wirkt auf `LiesRueckgabe` allein an den sechs Trennern (Leerzeichen, Tabulator, `=`, `;`, `,`, `"`) und an den 24 Buchstaben von `return`, `minimal` und `representation` (die zwölf Buchstaben r e t u n m i a l p s o, je in zwei Schreibweisen). Gleichwertig und ungefangen ist jede Änderung der Menge an den übrigen 226 Bytes (0x00 bis 0xFF ohne diese 30): ein Name oder ein Wort mit einem solchen Byte ist, ob das Byte Token ist oder nicht, weder `return` noch `minimal` noch `representation`, das Element zählt in beiden Fällen nicht beziehungsweise ergibt `Keine`. Gesehen (Tabelle der Fix-Runde): T1 (`!` entfernt), T2 (`Z` entfernt), T5 (`5` entfernt), T6 (`_` entfernt) und „`tchar` ab 0x80“ lassen `TestLiesRueckgabeUnlesbar` grün (T2 färbt allein `TestRedigiereHeaderNamen` aus `redaction.go`); je ein Byte aus den 30 (T3 `l`, T7 `=`, T8 `;` rot im benannten Test, T4 `R` rot-paket) färbt Tests der Suite. Folge im Betrieb: keine, die ein Client sieht. Zeile 24: die Stufe U erhält ihre Header aus der ursprünglichen Anfrage (`upstreamHeader(a.Header)`), nicht aus `stufen.upstream`; ein Streichen des Headers in `vorbereiten` an U (H06, H12, H13) ist gleichwertig, und `TestPreferGehtUnveraendertInDenUpstream` fängt `upstreamHeader` (U1, U2). Zeile 23: ein Zustand im Dienst, der innerhalb der Transaktion gesetzt und gelesen wird, bleibt ungefangen. Grund: Transaktionen laufen nacheinander (`SPEC-129` Nr. 11), so dass innerhalb einer Transaktion kein anderer Aufruf dazwischenliegt; der Fänger zählt die Wirkung (ETag je Antwort) und fängt ein Lesen oder Schreiben außerhalb der Transaktion. Folge im Betrieb: eine solche Mutation ändert keine Antwort, die ein Client sieht.

**Mutationen und Raten** (`AGENTS.md` §3.10; Implementer, Messender: der Implementer selbst; Last des Hosts bei der Messung: Load average 9 bis 27 auf der Maschine, die weitere Worktrees und Agenten teilen; die Tabelle gilt für zwei vollständige Durchläufe der Mutationsliste auf dem Stand des Commits; jede Mutation läuft per Perl auf einer Kopie des Arbeitsbaums im Container des Images `odata-recorder:race`, netzlos, `--memory=4g --cpus=4`, und die benannten Tests laufen je Prozess mit `go test -count=1`; die Spalte *Rate* nennt rote Prozesse von allen Prozessen). Wahrscheinlich rot sind allein die Zeile 23 (Zahl der Prozesse in der Spalte *Rate*), die Mutation s22c (Reihenfolge der Map, mit der Rate in der Tabelle) und die Zeile 9a (Zeit, Raten im Absatz nach der Tabelle):

| Zeile | Mutation (Kennung) | roter Test | Rate |
|---|---|---|---|
| 1 | Wert byte-genau (m01); Unicode-Faltung des Wortes, `ſ` als `s` (m02b) | `TestLiesRueckgabeBeispiele`, `TestLiesRueckgabeSchreibweise` | 1/1, 1/1 |
| 2 | Wert byte-genau (m01); Name byte-genau (m02) | `TestLiesRueckgabeSchreibweise`, `TestLiesRueckgabeBeispiele` | 1/1, 1/1 |
| 3 | `TrimSpace` statt `OWS` (m03); nur Leerzeichen (m03b); zusätzlich 0x0B (m03c) | `TestLiesRueckgabeLeerraum`, `…UnicodeLeerraum`, `…Unlesbar`, `…Beispiele` | 1/1 je Mutation |
| 4 | letzte Instanz über Zeilen (m04); letzte in der Zeile (m04d); erste unterstützte (m04a); nur erste Kopfzeile (m04c) | `TestLiesRueckgabeErsteInstanz`, `…Beispiele`, `…Folgen` | 1/1 je Mutation |
| 5 | nur `Prefer` gelesen (s22); Reihenfolge absteigend (s22b); Reihenfolge der Map (s22c) | `TestPreferNamenUndReihenfolge` | 1/1 je Mutation; s22c: 10/10 Prozesse |
| 6 | Paare nicht verfolgt (m06); Zeilen vor dem Lesen verbunden (m06b); Backslash schützt nicht (m06c) | `TestLiesRueckgabeAnfuehrung` | 1/1 je Mutation |
| 7 | Rest als Fehler gelesen (m07) | `TestLiesRueckgabeRest` | 1/1 |
| 8 | Name als Präfix lesbar (m08a); Rest hinter dem Wort als Wahl (m08b); leeres Wort als `Minimal` (m10) | `TestLiesRueckgabeUnlesbar`, `…Beispiele` | 1/1 je Mutation; `tchar` ab 0x80 (m08c) bleibt grün, gleichwertig (siehe Prüfgrenzen) |
| 9 | Eingabe sortiert (m09a); letzte Zeile gekürzt (m09c); Indexzugriff hinter dem Ende (m09b) | `TestLiesRueckgabeFolgen`, `…ErsteInstanz` (m09b: Panik in `…Folgen`) | 1/1 je Mutation |
| 10 | POST · Minimal als `201` (s01); PATCH, PUT · Entity als `204` (s02); `Content-Type` an der `204` (s03); `Vary` (s04a, s04b); `ETag` bei aus (s05a, s05b); `Body` nil je Zelle der `204` (D27, sieben Mutationen) | `TestPreferAntwortTabelle`, `TestPreferLesendOhneWirkung`, `TestPreferMinimalEntityId` | 1/1 je Mutation |
| 11 | `Wurzel` aus `Host` (s06a); `OData-EntityId` gekürzt (s06b); `Location` bei PATCH (s06c); `OData-EntityId` ohne Wurzel, mit dem echten Schreiber (e04) | `TestPreferAdresseNieAusHost`, `TestPreferMinimalEntityId` | 1/1 je Mutation |
| 12 | Echo des Client-Werts (s07a); Schreibweise des Clients (s07b); zweite Kopfzeile (s07c); weitere Präferenz im Wert (s07d) | `TestPreferApplied` | 1/1 je Mutation |
| 13 | `DELETE` liest `Prefer` (s08) | `TestPreferAppliedNurBeiErfolg`, `TestPreferLesendOhneWirkung` | 1/1 |
| 14 | unbekannter Wert ist ein Fehler (s09); Wert von `Prefer` im Fehler (s09b) | `TestPreferFehlerGleich` | 1/1 je Mutation |
| 15 | `Put` mit anderer Version (s10); zweites `Entities.Get` (s11); `Advance` ausgelassen bei Minimal (s12) | `TestPreferZustand` | 1/1 je Mutation |
| 16 | `DELETE` mit Wahl (s08); `GET` mit `return=minimal` ohne Body (s23) | `TestPreferLesendOhneWirkung` | 1/1 je Mutation |
| 17 | Auswerten vor Schritt (b) mit Fehler (s09) | `TestPreferFehlerGleich` | 1/1 |
| 18 | ETag der Version vor der Mutation (s20); zweites Lesen (s11); ETag nicht gesetzt (e01); ETag der Version davor (e02); Property fehlt im Body (e03) | `TestPreferAntwortTabelle`, `TestPreferEntityGleichGet` | 1/1 je Mutation |
| 19 | `Adresse` bei PATCH ohne Wahl (s13); `SchreibeEntity` bei POST · Minimal (s14); `SchreibeAnzahl` bei PATCH · Entity (s15) | `TestPreferSchreiberAufrufe`, `TestPreferSchreiberArgumente` | 1/1 je Mutation |
| 20 | `Auswahl` aus `$select` (s16); `ctx` ersetzt an S24 (s17a), an S26 (s17b); `Accept` nicht weitergereicht (s17c) | `TestPreferSchreiberArgumente` | 1/1 je Mutation |
| 21 | Fehler von `Adresse` (s18a), `SchreibeAnzahl` (s18b), `SchreibeEntity` (s18c) verschluckt; Gegenstand an S24, S25, S26 je einer der sieben anderen (G-S, 21 Mutationen) | `TestPreferSchreiberFehler` | 1/1 je Mutation |
| 22 | leere Darstellung zugelassen bei `SchreibeEntity` (s19a), bei `SchreibeAnzahl` (s19b) | `TestPreferDarstellungLeer` | 1/1 je Mutation |
| 23 | ETag der Antwort aus einem Lesen nach der Transaktion (s21) | `TestPreferGleichzeitig` (Body und Kopfzeile) | 20/20 Prozesse, je Prozess 60 frische Aufbauten |
| 24 | `Prefer` aus den Headern des Fallbacks (h01); `Preference-Applied` angefügt (h02), gestrichen (h03); Header in einem Zweig von `vorbereiten` gestrichen (H04 bis H11); `Prefer` in `upstreamHeader` gefiltert (U1, U2) | `TestPreferGehtUnveraendertInDenFallback`, `TestPreferGehtUnveraendertInDenUpstream` | 1/1 je Mutation |
| 25 | nur die erste Kopfzeile (a01); Reihenfolge umgekehrt (a02); `OData-EntityId` gefiltert (a03); `200` als `201` (a04); `Preference-Applied` gefiltert (a05) | `TestPreferHeaderUnveraendert`, `TestPreferAntwortUnveraendert` | 1/1 je Mutation |
| 6a | Backslash im Wort: R15, B1, B2, B4, B6, B8 (Fix-Runde) | `TestLiesRueckgabeBackslashWort` | 1/1 je Mutation |
| 6b | Backslash im Paar eines fremden Elements: EA1 bis EA7; außerhalb eines Paares: EA2, EA8 (Fix-Runde 2) | `TestLiesRueckgabeBackslashFremd` | 1/1 je Mutation |
| 9a | Kopie je Element (Q1), je Zeichen im Wort mit Backslash (Q2), Zeile je Element erneut gelesen (Q3), Name quadratisch (Q5), Wert als Token (QW), Wort ohne Backslash (QG), Lauf eines fremden Elements (QR), fremdes Paar (QP), Backslash im fremden Paar (QB), Leerraum hinter `=` (QL2) und hinter dem Wort (QL3), offenes Wort (QO) | `TestLiesRueckgabeZeit` | je 8/8 für Q1, Q2, Q5, QW, QG, QR, QP, QB, QL2, QL3, QO (Raten im Absatz nach der Tabelle der Fix-Runde); Q3 nicht neu gefahren (1/1 auf dem Stand mit sieben Formen); ohne Fehler im Code siehe Absatz |
| 26 | gilt nicht (kein Inhalt des Handbuchs, §6) | gilt nicht | gilt nicht |
| 27 | Antwort mit dem Body der vorigen Mutation (E4, Paketvariable) | `TestPreferWiederholungGleichGet`, `TestPreferEntityGleichGet` | 1/1 |
| 28 | Paketvariable (M28) | `TestPaketOhneZustand` | 1/1 |

**Fix-Runde nach dem Review** (`AGENTS.md` §3.10, §3.17; Messender: der Implementer mit `sh tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, `--memory=4g --cpus=4`, `--network none`, Klasse `andere` ein Prozess je Stufe, Last des Hosts laut Spalte `last` der Ausgabe 4,6 bis 9,9 von 4 CPUs; die Mutationen E3 und E4 am Paket `src/bootstrap` laufen nicht im Werkzeug (`Ausgangsstand rot`, `SPEC-210` Nr. 5 e) und fuhren per Perl auf einer Kopie des Baums im Container, `go test ./src/bootstrap -run`). Je Befund die Menge nach dem Wortlaut der Zusage, ihre Elemente, die Mutationen des Befunds und mindestens eine an einem Element, das der Befund nicht nannte, rot gesehen:

| Befund | Menge · Elemente | Mutationen des Befunds | Mutation an einem nicht genannten Element | Ergebnis |
|---|---|---|---|---|
| F-1 | Backslash im Wort: 3 Wörter × (11·*L*+10 Formen: 87, 164, 87) × 3 Zusammenhänge = 1014 | R15, B1, B2 | B4 (nur an der ersten Stelle), B6 (nicht vor `n`), B8 (Backslash am Ende schließt das Paar) | alle rot (`TestLiesRueckgabeBackslashWort`); B7 (nicht vor `"`) grün, gleichwertig: das Wort enthält dann `"` oder das Paar bleibt offen, beides `Keine` |
| F-2 | Form von `vorbereiten` 8 × Wert 172 × Name 2 × Methode 5 × Header der Stufe 2 = 27520; Upstream 5 × 3 × 172 × 2 = 5160 | H04 (R, gepuffert), H05 (M, gepuffert) | H07 (ohne Body), H08 (Laenge über der Grenze), H09 (nicht fallback-fähig), H10 (Lesefehler), H11 (Body über der Grenze nach dem Lesen), U1, U2 (`upstreamHeader` filtert `Prefer`) | alle rot; H06, H12, H13 (Header von U in `vorbereiten`) grün, gleichwertig (U liest `a.Header`, nicht `stufen.upstream`) |
| F-3 | `Body` der Antwort ohne Entity: sechs Zellen der `204` (POST · Minimal, PATCH · Keine und Minimal, PUT · Keine und Minimal) und `DELETE`; im Test mit echtem Schreiber PATCH, PUT, POST · Minimal | D27 (alle Minimal), E3 (echter Schreiber) | D27 je Zelle (POST · Minimal, PATCH · Keine, PATCH · Minimal, PUT · Keine, PUT · Minimal, `DELETE`), E3 für PUT · Minimal | alle rot (`TestPreferAntwortTabelle`, `TestPreferLesendOhneWirkung`, `TestPreferMinimalEntityId`) |
| F-4 | Zeit je Länge: 7 Formen langer Zeilen × 2 Längen | Q1 (Kopie der Zeile je Element) | Q2 (Kopie je Zeichen im Paar), Q3 (Zeile je Element erneut gelesen, ohne Zuweisung), Q5 (Name quadratisch) | alle rot (`TestLiesRueckgabeZeit`), Raten unten |
| F-8 | Menge `tchar`: 256 Bytes, davon 30 mit Wirkung (sechs Trenner, 24 Buchstaben), 226 gleichwertig | T1 (`!`), T2 (`Z`) | T5 (`5`), T6 (`_`) grün, gleichwertig; T3 (`l`), T7 (`=`), T8 (`;`) rot, T4 (`R`) rot-paket (`TestLiesRueckgabeBeispiele`, `…Schreibweise`) | wie Prüfgrenze; T1 grün, T2 rot-paket durch `TestRedigiereHeaderNamen` |
| F-11 | Gegenstand je Stelle: 7 Gegenstände × S24, S25, S26 = 21 | G-S24, G-S25, G-S26 mit `Workspace` | die übrigen 18 (`Session`, `Snapshot`, `Snapshotname`, `Import`, `Exchange`, `Dienst` je Stelle) | alle 21 rot (`TestPreferSchreiberFehler`) |
| F-5 | Wiederholung: Methode 2 × Etag 2 × Mutation 2 = 8; Paketvariable | E4 (Antwort mit dem Body der vorigen Mutation, Paketvariable `staleBodies`), M28 (Paketvariable `letzteWahl`) | E4 gegen `TestPreferEntityGleichGet` (Bestand) ebenfalls rot | E4 rot (`TestPreferWiederholungGleichGet`, `TestPreferEntityGleichGet`), M28 rot (`TestPaketOhneZustand`) |

Raten von `TestLiesRueckgabeZeit` (Zufallsfänger; je Prozess `-test.count=1`, bis vier Prozesse zugleich, ohne Detektor; Last des Hosts 5,4 bis 9,9 von 4 CPUs; Messender: der Implementer): ohne Fehler im Code (Mutation Z0, nur ein Kommentar) 0 von 60 Prozessen rot; Q2 12 von 12, Q1 8 von 8, Q5 8 von 8; Q3 ein Lauf rot. Zweite Messung auf dem Stand der Fix-Runde (Last 3,3 bis 4,8 von 4 CPUs): Z0 0 von 60, Q2 12 von 12, Q1 4 von 4, Q5 4 von 4. Eine Einzelzahl gilt als Größenordnung: unter stärkerer Last kann der Test ohne Fehler im Code rot werden, die Grenze liegt beim 4-fachen des erwarteten Wachstums, und ein Versuch von fünf genügt.

**Fix-Runde 2 nach dem Review** (`AGENTS.md` §3.10, §3.17; Messender: der Implementer mit `sh tools/harness/mutationslauf.sh`, Image `odata-recorder:race`, `--memory=4g --cpus=4`, `--network none`, ein Prozess je Stufe, unter `flock`; Last des Hosts laut Spalte `last` der Ausgabe 4,6 bis 75 von 4 CPUs, geteilt mit weiteren Läufen). Je Befund die Menge nach dem Wortlaut der Zusage, die Achsen mit Elementzahl, die Mutationen des Befunds und mindestens eine an einem Element, das der Befund nicht nannte, rot gesehen:

| Befund | Menge · Achsen · Elemente | Mutationen des Befunds | Mutation an einem nicht genannten Element | Ergebnis |
|---|---|---|---|---|
| N-1 | Kommentar in neuem Testcode im Indikativ: ein Satz in `prefer_test.go` | Konjunktiv „sonst bliebe …“ | Lektüre der übrigen Kommentare des Diffs `5b09dd6a..HEAD` gegen die Klassen von `AGENTS.md` §3.7 (Suche nach `sonst`, `hätte`, `wäre`, `würde`, `bliebe`): kein weiterer Treffer außer einem Indikativ über zwei Fälle | umgeschrieben; kein Gate fängt Konjunktive |
| N-2 | „Backslash an jeder Stelle der Grammatik“ (`SPEC-285` Nr. 2): Stelle (im Wort, im Paar eines fremden Elements, außerhalb eines Paares) × Länge der Folge *k* 0 bis 7 (gerade und ungerade) × Zeichen dahinter × Stelle des Paares. Neu geführt: im Paar 4 Stellen × 3 Inhalte × 8 Längen × 4 Zeichen = 384; außerhalb 4 Stellen × 2 Zeichen × 8 Längen = 64; im Wort bleiben 1014 (Zeile 6a) | EA4 (nicht nach einem Backslash), EA2 (außerhalb schützt), EA1, EA3 | EA5 (schützt nur am Anfang des Paares), EA6 (schützt zwei Zeichen), EA7 (schützt nicht nach drei Backslashes), EA8 (außerhalb schützt nur vor `,`) | alle acht rot in `TestLiesRueckgabeBackslashFremd` (1 von 1 Prozess je Mutation), EA4 und EA2 im benannten Test |
| N-3 | „Stelle des langen Laufs“ (`SPEC-285` Nr. 1): endliche Achse mit 20 Stellen (Zeilen, Elemente, Kommas, Leerraum an vier Stellen, Name, Wert als Token, Wort ohne Backslash, mit Backslash, offen, Rest hinter dem Wert, fremder Wert, fremder Rest, fremdes Paar ohne Backslash, mit Backslash, mit Folge von Backslashes, offen, Backslash außerhalb) × 2 Längen; je Stelle ein Baustein | QW, QG, QR | QP (fremdes Paar), QB (Backslash im fremden Paar), QL2 (Leerraum hinter `=`), QL3 (Leerraum hinter dem Wort), QO (offenes Wort) | alle rot in `TestLiesRueckgabeZeit`, Raten unten; unbeschränkt bleibt der Baustein je Stelle (§3 Prüfgrenzen) |
| N-4 | Zustandsfeld der Rückfrage (a): ein Ausgang | zwei Ausgänge in einem Feld | alle Einträge von §6 mit zwei Ausgängen in einem Feld: die Rückfrage (d) trägt jetzt allein *offen* | bereinigt |
| N-5 | Sendung „Zeile `ETAG_ENABLED` des Handbuchs kürzen“: Nehmer in `open/` mit Wirkung auf die Zeile | Planner als Adresse ohne Eintrag | `slice-handbuch-mock` (`open/`): §1 *Ausdrücklich NICHT* und DoD gelesen, nimmt die Zeile nicht an; `slice-etag-versionen`, `slice-if-match` in `done/` | §6 trägt die Zeile mit Ausgang *offen* und Adresse Planner |
| N-6 | Zusage „ein Backslash außerhalb eines Paares hat keine Wirkung“ in Zeile 6b der Tabelle | EA2 war nur rot-paket durch `TestLiesRueckgabeFolgen` | EA8 | Zeile 6b nennt die Zusage und den Fänger; EA2 und EA8 rot im benannten Test |

Raten von `TestLiesRueckgabeZeit` auf dem Stand der Fix-Runde 2 (je Prozess `-test.count=1`, ohne Detektor, ein Prozess nach dem anderen; Messender: der Implementer): QW, QG, QR, QP, QB, QL2, QL3, QO, Q1, Q2, Q5 je 8 von 8 Prozessen rot bei Last des Hosts 4,6 bis 75 von 4 CPUs; ohne Fehler im Code (Mutation Z0, nur ein Kommentar) 0 von 48 Prozessen rot (8 bei Last 4,6; 40 bei Last 80 von 4 CPUs). Eine Einzelzahl gilt als Größenordnung; mit 20 Formen hat der Test mehr Gelegenheiten für ein Rot ohne Fehler im Code als mit sieben, und die Grenze (4-faches des erwarteten Wachstums, ein Versuch von fünf) bleibt.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (etwa Header-Werte mit NBSP) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock` auf der Sperrdatei `/home/builder/work/odata-recorder/.wt/heavy.lock` <!-- d-check:ignore (absoluter Pfad der Sperrdatei des Hosts, außerhalb des Repos) -->; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Trigger der Welle). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-285` bis `SPEC-289`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-mock-schreiben` (Statuscodes und Antworten der Mutationen, die der Slice umschaltet), `slice-odata-json-schreiben` (die Entity im Body der Antwort), `slice-http-adapter` (der Header erreicht den Use Case als Teil des Requests). Früheste Lage für Code: `slice-mock-schreiben` und `slice-odata-json-schreiben` in `done/`; der Slice kann **nicht** vor ihnen beginnen, weil `MockUseCase` mit Mutation nicht existiert.
- **Binary und Handbuch:** der Slice schreibt kein Handbuch-Kapitel und liefert kein Beispiel gegen das Binary (Entscheidung des Architect, §6); er schließt nicht an `slice-bootstrap-m3` oder `slice-handbuch-meldungskatalog`. Ein Start vor dem Trigger der Welle ist eine Entscheidung des Users, nicht dieses Plans.
- **Was der Slice nicht braucht:** `slice-etag-versionen` und `slice-if-match` (unabhängiger Ast, `welle-etag` §5), `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-expand-batch`, `welle-odata-v2`.
- **Größenmarke** (Planner, 2026-11-18, `SPEC-128` Nr. 1; die Faktoren stehen allein dort): Die Tabelle in §3 hat 26 Zeilen, *a* = 1 (Zeile 23), *b* = 25. Der Architect nennt keine Zahl für den Produktcode; der Planner setzt *p* = 300 (`model.LiesRueckgabe` mit Grammatik und `Rueckgabe` etwa 130, Wahl und Antworten je Zelle von `SPEC-286` samt `Preference-Applied`, `Location`, `OData-EntityId` im Dienst etwa 130, Adapter etwa 20, Rest etwa 20) und *f* = 1 (ein Fake für den Upstream des Hybrid-Dienstes mit Haltepunkt für die Header, falls der vorhandene `stufe` sie nicht festhält; die Fakes des Schreibers und des Dienstes am Adapter bekommen Zähler und Haltepunkt, es entsteht dort kein neuer). Der Architect legt fest: `model.LiesRueckgabe` liest fremde Bytes; ein Zuschlag für Leser fremder Bytes gibt es nach `SPEC-128` Nr. 2 nicht mehr, die Klasse ändert die Zahl nicht. Rechnung: *S* = 85·1 + 80·25 + 200·1 + 300 = 2585; *M* = 100 · ⌈135·2585 / 10000⌉ = 100 · ⌈34,90⌉ = 3500. `basis` ist `5a98ef2a1da9387ade17d60a3306aeaa7012521c`, der Commit des Architect, der den Vertrag abschließt (Vorfahr von `HEAD`, kein Code danach). Die Pfade sind die Code-Pfade aus §3 samt der neuen Testdatei des Hybrid-Dienstes; das Handbuch zählt nicht (`SPEC-128` Nr. 3 m: `docs/` darf kein Pfad nennen). Die Messung zählt jede Änderung an den Pfaden (Nr. 10 g): ein Slice, der dieselben Pfade ändert, verschiebt sie. Nach der Fix-Runde 2 hat die Tabelle 31 Zeilen (*a* = 1, *b* = 30, Zeile 26 zählt als `andere`) und ergibt *S* = 2985; die Marke bleibt 3500, sie ist nach dem ersten Code-Commit fest (`SPEC-128` Nr. 5 c: im Code ein Hinweis, kein Befund).
- **Schnittfrage** (Planner, vor dem Code, `AGENTS.md` §3.14): *S* ≥ 2223, also ja; die Antwort ist **nein**, Grund im Block unter `schnitt-grund`. Geprüft wurden vier Schnitte: (1) Grammatik und Wahl im Modell (Zeilen 1 bis 9, *S* sänke auf etwa 1850) gegen Antwort im Dienst: das Modell allein ist Code ohne Auslöser und ohne Handbuch-Beispiel, ein Schnitt nach Schichten. (2) `POST` gegen `PATCH` und `PUT`: alle drei teilen Wahl, Leser und Antwort-Zweig des Dienstes (`SPEC-286` Nr. 1 ist eine Tabelle). (3) `return=minimal` gegen `return=representation`: teilen Leser, Header `Preference-Applied` und die Zählung der Aufrufe des Schreibers (`SPEC-288` Nr. 3 ist eine Tabelle). (4) Zeile 23 (nebenläufig): ein Test am selben Code, ein Teil ohne Produktcode und ohne Auslöser. Das Wagnis: *p* ist vom Planner gesetzt, nicht vom Architect; liegt es bei 450, steigt *S* auf 2735 und *M* auf 3700, die Entscheidung bleibt; die Marke ist nach dem ersten Code-Commit fest, die Folge ist ein Eintrag nach `SPEC-128` Nr. 5 d.

```groessenmarke
pfad: src/hexagon/model/
pfad: src/hexagon/services/mock/
pfad: src/adapters/driving/httpodata/
pfad: src/hexagon/services/hybrid/prefer_test.go
produktcode: 300
fakes: 1
basis: 5a98ef2a1da9387ade17d60a3306aeaa7012521c
schaetzung: 2585
marke: 3500
schnitt-grund: Ein Mechanismus: die Wahl (Leser im Modell, Auswertung im Dienst) und die Antwort je Zelle von SPEC-286; die Teile (Modell gegen Dienst, Methoden, Wahlen, Nebenläufigkeit) teilen Leser, Header und Zählung der Aufrufe des Schreibers, und der Modellteil allein liefert Code ohne Auslöser und ohne Handbuch-Beispiel.
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Tabelle *Zusage · Fänger* liegt vor (26 Zeilen), die Marke steht im Block darüber (3500).
- `in-progress` → `open` (blockiert — Carveout?): eine Randform, die `SPEC-285` bis `SPEC-289` nicht entscheiden (siehe §6).

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

- Die Menge der Header-Formen ist unbeschränkt (jede Zeichenfolge) — **Ausgang:** offen; Generator über die Klasse (Tabelle §3, Zeilen 2 bis 9) mit Prüfgrenze und Folge im Betrieb nach `SPEC-289` Nr. 3 (`AGENTS.md` §3.17).
- `Prefer` darf den Zustand nie ändern; ein Fänger vergleicht den Zustand nach derselben Mutation mit und ohne Header, über alle Methoden und alle Werte (Schleife über beide Mengen, `AGENTS.md` §3.15; Zeile 15) — **Ausgang:** offen.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, 2×; Zeilen 15, 19, 20) — **Ausgang:** offen.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster steht unten, jede Zelle mit ihrem Ort; eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen bis zur Closure.
- Der Body der Antwort gleicht dem `GET` danach (`SPEC-288` Nr. 1, Zeile 18): zeigt der Fänger einen Unterschied, den `SPEC-288` nicht erklärt, hält der Implementer an und gibt ihn an den Architect zurück (`AGENTS.md` §3.13) — **Ausgang:** offen.

- Rückfrage an den Architect (a): Gilt `SPEC-288` Nr. 1 (der Body der Antwort gleicht, bei gleichem `Accept` und ohne `$select`, byte-gleich dem `GET` derselben Entity unmittelbar danach) gegen den Code? Der Stand von `src/hexagon/services/mock/schreiben.go` bildet die Entity bei `POST` aus `neu` (`deute`, `etagVon`) und liest nicht erneut; für `PATCH` und `PUT` gibt es noch keinen Zweig, und ob das `GET` danach dieselben Bytes liefert (Reihenfolge der Mitglieder, `@odata.context`, `@odata.etag`, Coverage-Header), hat niemand gemessen. Fänger: Zeile 18 — **Adresse:** Architect. Messung des Implementers: Zeile 18 ist grün mit dem echten Schreiber, der echten Eingabe und SQLite (`src/bootstrap/prefer_ende_zu_ende_test.go`, 36 Fälle: `PATCH` mit Teilbody und mit `{}`, `PUT` mit fehlender nullbarer Property und mit allen Properties, `POST` mit und ohne Wahl, `Accept` ohne, `none`, `minimal`, `Dienst.Etag` aus und an): Body und ETag der Antwort gleichen dem `GET` danach, es gibt keinen Unterschied, den `SPEC-288` erklären müsste; **Ausgang:** entschieden — `SPEC-288` Nr. 1 gilt gegen den Code; der Fänger ist `TestPreferEntityGleichGet` (36 Fälle) mit dem Beleg der Mutationen E3 und E4 (siehe Tabelle der Mutationen), und `TestPreferWiederholungGleichGet` (8 Fälle) fängt die Antwort der zweiten gleichen Mutation.
- Prüfung des Planner (b): Die Aussage in `SPEC-049` Nr. 10 f, die Fixtures der Muss-Zeilen sendeten `Prefer` nicht, ist gegen die Fixtures geprüft: `grep -i prefer` über `src/bootstrap/abnahme_*.go` findet keinen Treffer; die Anfragen der Abnahme tragen nur `Accept` (`abGH`, GET), `Content-Type` und `X-Recorder-Workspace` (`abAnfrage.anfrage`), die Mutationen der Fixtures setzen keinen Header; `tests/` nennt `Prefer` nur als Vary-Header in `determinismus_test.go` (Antwort, nicht Anfrage). Die Aussage gilt — **Ausgang:** entfallen. Ein Treffer in `src/hexagon/services/mock/schreiben_antwort_test.go` (`TestPreferIgnoriert`) sagt das alte Verhalten zu und ist in §3 als Änderung geführt.
- Rückfrage an den Architect (c): Der Plan trug keine Zahl für die Produktcode-Erwartung; der Planner hat *p* = 300 und *f* = 1 gesetzt (§4). Bestätigt der Architect eine andere Zahl, gilt sie vor dem ersten Code-Commit mit `ersetzt:` im Block (`SPEC-128` Nr. 3 n); danach ist die Marke fest. Messung des Implementers (`git diff --numstat` gegen `basis`): der Produktcode umfasst rund 200 Zeilen (131 in `src/hexagon/model/rueckgabe.go`, 119 in `src/hexagon/services/mock/schreiben_antwort.go`, abzüglich 48 netto in `src/hexagon/services/mock/schreiben.go`) gegen *p* = 300; die Marke bleibt — **Ausgang:** entschieden.
- Rückfrage an den Architect (d): `SPEC-285` Nr. 1 sagt zu, die Zeit von `LiesRueckgabe` wachse linear mit der Gesamtlänge; die Liste *Ungeprüft* nennt nur die Länge über der Grenze des Adapters. `TestLiesRueckgabeZeit` (Zeile 9a) führt die 20 Stellen des langen Laufs der Grammatik, je mit einem Baustein, und ist ein Zufallsfänger unter Last des Hosts (Raten in der Tabelle der Mutationen); ein Lauf mit einem anderen Baustein an einer dieser Stellen, der quadratisch liest, bleibt ungefangen (§3 Prüfgrenzen). `SPEC-285` *Ungeprüft* nennt diesen Rest (Verhältnis 32, Grenze 128-fach, Grund: ein Zähler verlangte einen Haken im Produktcode); die Zusage in Nr. 1 bleibt. Der Wortlaut von *Ungeprüft* ist der Stand des Fängers: die Liste der 20 Stellen steht in der Spezifikation und im Test — **Adresse:** Architect (nimmt den Wortlaut an) — **Ausgang:** offen.
- Die Sperrdatei der schweren Läufe liegt außerhalb des Arbeitsverzeichnisses: `/home/builder/work/odata-recorder/.wt/heavy.lock` <!-- d-check:ignore (absoluter Pfad der Sperrdatei des Hosts, außerhalb des Repos) --> (Vorgabe des Operators; §3 Betriebsregeln nennt dieselbe Datei); die schweren Läufe dieses Slice fuhren unter ihr — **Ausgang:** entschieden.
- Handbuch-Kapitel der DoD: `make handbuch-check` bindet nur Beispiele, die das Binary mit Argumenten aus `A-Z a-z 0-9 . _ = -` aufrufen und deren Ausgabe, Fehlerausgabe und Ausgangswert sie nennen (`SPEC-200` Nr. 5 und 6); ein HTTP-Request an den Mock (Header `Prefer`, Antwort mit Status und Headern) ist darin nicht ausdrückbar, und `SPEC-200` Nr. 1 lässt Verhalten des Binary im Handbuch nur als gebundenes Beispiel zu. Der Implementer hat kein Kapitel geschrieben und keinen Weg im Code oder im Handbuch entschieden; Zeile 26 hat keinen Fänger. **Entscheidung (Architect, 2026-11-18): Weg B, kein Eingriff in `SPEC-200`.** Das Handbuch beschreibt `Prefer` nicht: `Prefer` ist kein Schlüssel der Einstellungen und kein Kommando, also keine Aussage der Klassen K1 bis K4 (`SPEC-200` Nr. 1); das HTTP-Verhalten tragen `SPEC-285` bis `SPEC-289`, der Fänger sind die Tests des Slice. Eine Beispiel-Form für HTTP (Weg A) verlangte eine laufende Instanz, einen Client und eine Erweiterung von Nr. 6 samt Prüfung in `tools/pruefung` für eine Zusage, die Spezifikation und Tests schon tragen. Zeile 26 der Tabelle in §3 gilt damit nicht (Grund: kein Inhalt des Handbuchs), der DoD-Punkt „Handbuch-Kapitel“ entfällt für diesen Slice, `make handbuch-check` bleibt unberührt. **Akzeptiertes Negativ:** die Zeile `ETAG_ENABLED` der Schlüsseltabelle nennt HTTP-Verhalten (Kopfzeile, `@odata.etag`, `412`) ohne gebundenes Beispiel; das ist kein Vorbild, sondern Prosa über Verhalten, die `SPEC-200` Nr. 14 b ungeprüft lässt (nur Reviewer und Abnahme lesen sie) und die Nr. 1 widerspricht. Sie gehört nicht zu diesem Slice; der Planner nimmt sie beim Slice auf, der `ETAG_ENABLED` liefert oder ändert (Zeile auf den Schlüssel und seine Wirkung als Einstellung kürzen). Folge im Betrieb: das Handbuch weicht in dieser einen Zeile von seiner Regel ab, bis das geschieht. — **Nachzug** (Implementer, im selben Commit wie die Messung): der DoD-Punkt „Handbuch-Kapitel“ ist gestrichen, Zeile 26 in §3 gilt nicht, der Punkt „Binary und Handbuch“ in §4 ist umgeschrieben — **Ausgang:** entschieden.
- Die Zeile `ETAG_ENABLED` der Schlüsseltabelle des Handbuchs nennt HTTP-Verhalten ohne gebundenes Beispiel und weicht von `SPEC-200` Nr. 1 ab (Akzeptiertes Negativ der Entscheidung oben). Es gibt keinen Nehmer: `slice-etag-versionen` und `slice-if-match` liegen in `done/`; `slice-handbuch-mock` (`open/`) liefert das Beispiel zu `ETAG_ENABLED`, sagt aber in §1 ausdrücklich „Die Zeile `ETAG_ENABLED` der Tabelle *Einstellungen* trägt der Geber, nicht dieser Slice“ und führt das Kürzen der Zeile weder in §1 noch in seiner DoD; kein anderer Slice in `open/` liefert oder ändert den Schlüssel (gelesen am 2026-11-18: §1 *Ausdrücklich NICHT* und DoD von `slice-handbuch-mock`). Der Punkt geht an den Planner, der den Nehmer bestimmt und die Sendung in dessen §1 oder DoD einträgt (`AGENTS.md` §3.13) — **Adresse:** Planner — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (jede Zelle nennt ihren Ort; „Keine“ ist die Wahl ohne `Prefer`, `SPEC-285` Nr. 3):

| Operation | Header fehlt | `return=minimal` | `return=representation` | leerer oder unbekannter Wert | mehrere Präferenzen, mehrere Header, widersprüchlich | fehlgeschlagene Mutation | Abbruch | Wiederholung |
|---|---|---|---|---|---|---|---|---|
| POST | `SPEC-285` Nr. 1 und 3 (Keine), `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1 | `SPEC-285` Nr. 3 und 4 (Keine), `SPEC-286` Nr. 1 | `SPEC-285` Nr. 1 bis 4 | `SPEC-287` Nr. 2 | `SPEC-287` Nr. 5 (Zeile 21) | `SPEC-287` Nr. 7, `SPEC-289` Nr. 1 (Zeilen 27, 28) |
| PATCH | `SPEC-285` Nr. 1 und 3 (Keine), `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1, `SPEC-288` Nr. 1 | `SPEC-285` Nr. 3 und 4 (Keine), `SPEC-286` Nr. 1 | `SPEC-285` Nr. 1 bis 4 | `SPEC-287` Nr. 2 | `SPEC-287` Nr. 5 (Zeile 21) | `SPEC-287` Nr. 7, `SPEC-289` Nr. 1 (Zeilen 27, 28) |
| PUT | `SPEC-285` Nr. 1 und 3 (Keine), `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1, `SPEC-288` Nr. 1 | `SPEC-285` Nr. 3 und 4 (Keine), `SPEC-286` Nr. 1 | `SPEC-285` Nr. 1 bis 4 | `SPEC-287` Nr. 2 | `SPEC-287` Nr. 5 (Zeile 21) | `SPEC-287` Nr. 7, `SPEC-289` Nr. 1 (Zeilen 27, 28) |
| DELETE | `SPEC-194` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 2 | `SPEC-287` Nr. 5 (Aufruf S26, Zeile 21) | `SPEC-289` Nr. 1 (Zeile 28) |
| GET | `SPEC-198` Nr. 1 und 8 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | gilt nicht: GET ändert nichts; der Fehler des Lesens ist unabhängig vom Header (`SPEC-287` Nr. 3) | gilt nicht: GET liest `Prefer` nie (`SPEC-287` Nr. 3); der Abbruch des Lesens steht in `SPEC-198` | gilt nicht: GET liest `Prefer` nie (`SPEC-287` Nr. 3), die Wiederholung des Lesens trägt `SPEC-198` |

**Zusatztabelle** (die Zusage hängt vom Modus und von `Dienst.Etag` ab; `AGENTS.md` §3.16):

| Modus | Zusage | Ort | Fänger |
|---|---|---|---|
| `mock`, `ETAG_ENABLED` aus | Antwort nach der Tabelle ohne `ETag` | `SPEC-286` Nr. 1, `SPEC-280` Nr. 2 | Zeile 10 (`Dienst.Etag` aus) |
| `mock`, `ETAG_ENABLED` an | Antwort nach der Tabelle mit `ETag` der Version nach der Mutation, Body mit `@odata.etag` | `SPEC-286` Nr. 1, `SPEC-288` Nr. 1 | Zeilen 10 und 18 (`Dienst.Etag` an) |
| `hybrid` | `Prefer` geht unverändert in den Fallback | `SPEC-287` Nr. 6 | Zeile 24 |
| `replay` | der Dienst liest `Prefer` nie; der Header bleibt Vary-Header des Match-Keys | `SPEC-287` Nr. 6, `SPEC-116` Nr. 5 | bestehender Fall des Match-Keys (`SPEC-061`, `SPEC-116`); dieser Slice fügt keinen hinzu (Negativ, Grund: Pfad ohne Änderung durch den Slice) |
| `record` | der Dienst zeichnet `Prefer` wie jeden Header auf, ohne ihn zu lesen | `SPEC-287` Nr. 6 | bestehender Fall der Aufzeichnung; dieser Slice fügt keinen hinzu (Negativ, Grund: Pfad ohne Änderung durch den Slice) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für die Fakes von Repository und Schreiber; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice liefert eine Zusage über Nebenläufigkeit (Zeile 23, `SPEC-289` Nr. 2: *n* gleichzeitige `PATCH` mit der Wahl `Entity`); sie hat einen Wirkungs-Fänger mit `-count=20`, weil der Detektor von `make test-race` kein Rennen meldet, das kein Zugriff auslöst (`AGENTS.md` §3.10). Sonst gibt es keinen gemeinsamen Zustand: `LiesRueckgabe` ist rein. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
