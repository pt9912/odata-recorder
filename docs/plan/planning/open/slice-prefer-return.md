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

**Verantwortlich:** —

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

- [ ] `return=minimal` und `return=representation` für POST, PATCH und PUT: Status, Body, `Location` und `OData-EntityId`, `Preference-Applied` nach `SPEC-286`; ohne `Prefer` unverändert das Muss-Verhalten.
- [ ] Randformen des Headers (Groß- und Kleinschreibung, Leerraum, mehrere Präferenzen in einem Header, mehrere Header, widersprüchliche Werte, leerer oder unbekannter Wert): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; ein nicht unterstützter Wert ergibt keinen Fehler.
- [ ] Unveränderter Zustand: mit und ohne `Prefer` derselbe Zustand des Workspace und dieselbe Revision; bei fehlgeschlagener Mutation dasselbe Fehlerobjekt.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün.
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
| `src/hexagon/services/mock/` | update | Wahl der Antwort nach der Präferenz (`SPEC-286`, `SPEC-288`), Aufrufe des Schreibers je Antwort |
| `src/hexagon/model/` | update | `Rueckgabe` und `LiesRueckgabe` (`SPEC-285`); die angewandte Präferenz steht als Header in der Antwort, kein neuer Typ am Port |
| `src/adapters/driving/httpodata/` | update | Der Adapter reicht alle Kopfzeilen `Prefer` durch und schreibt `Preference-Applied`, `OData-EntityId` und Status `200` unverändert (`SPEC-286`) |
| `spec/spezifikation.md` | update | Vertrag mit Randformen (Architect, `SPEC-285` bis `SPEC-289`) |
| Tests | neu | Fälle von `LH-FA-07.b` je Methode und je Wert; Generator über die Klasse der Header-Formen |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

**Auftrag an den Architect** — erledigt: der Vertrag steht in `SPEC-285` (Grammatik und Wahl), `SPEC-286` (Antwort je Methode und Wahl, `Vary`), `SPEC-287` (GET, DELETE, Fehler, Zustand, Modi), `SPEC-288` (Herkunft der Entity, Aufrufe des Schreibers; der Schreiber aus `SPEC-168` genügt unverändert), `SPEC-289` (Nebenläufigkeit, Ungeprüftes). Die vier Punkte des Auftrags: (1) Antwort je Methode und Wahl `SPEC-286` Nr. 1 und 2; (2) Grammatik `SPEC-285` Nr. 2 bis 4; (3) GET, DELETE und Fehler `SPEC-287` Nr. 2 und 3; (4) Entity aus dem geschriebenen Eintrag `SPEC-288` Nr. 1 und 2.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Mengen als Schleife oder Generator aus dem Wortlaut der Zusage, `AGENTS.md` §3.17; die Mutation je Zeile fährt der Implementer selbst rot und trägt die Raten und den Bericht in diesen Absatz ein, `AGENTS.md` §3.10):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | `LiesRueckgabe` liefert für jede Zeile der Beispiele von `SPEC-285` Nr. 4 das dort genannte Ergebnis | Tabellentest über alle Zeilen der Tabelle, Werte aus Bytewerten gebaut (`AGENTS.md` §3.18), Datei gegengelesen | Vergleich der Werte byte-genau statt ohne Beachtung der Schreibweise | Tabelle |
| 2 | Name und Wert ohne Beachtung der ASCII-Schreibweise (Nr. 3) | Schleife über alle 64 Schreibweisen von `return`, alle 128 von `minimal` und alle 16384 von `representation` | nur Kleinschreibung des Namens, nur Kleinschreibung eines Wertes | Generator |
| 3 | Leerraum ist genau Leerzeichen und Tabulator, an jeder Stelle (Nr. 2) | Schleife über die Stellen (vor dem Namen, vor und nach `=`, nach dem Wort, um `,`, vor `;`) × die Zeichen `0x00` bis `0xFF` (Orakel: `0x20` und `0x09` sind Leerraum, jedes andere Byte nicht) | `strings.TrimSpace` statt `OWS` (fängt `0x0B`, `0x0C`, `0xA0`, NBSP als `0xC2 0xA0`); nur Leerzeichen als `OWS` | Generator |
| 4 | Die erste lesbare Instanz `return` entscheidet; spätere, auch widersprüchliche, nie (Nr. 3) | Schleife über alle Folgen der Länge 0 bis 3 aus den neun Elementen {`return=minimal`, `return=representation`, `return=foo`, `return`, `return=`, `respond-async`, `odata.x=1`, `return@x`, leer} × jede Verteilung auf Kopfzeilen (eine Zeile, je Element eine Zeile, jede Teilung); Orakel: erstes lesbares `return` | letzte Instanz statt erster; erste unterstützte statt erster; nur die erste Kopfzeile gelesen | Generator |
| 5 | Mehrere Kopfzeilen und mehrere Schreibweisen des Namens werden in der Bytefolge der Namen, dann der Reihenfolge nach gelesen (Nr. 1) | Schleife über Schreibweisen des Namens (`Prefer`, `prefer`, `PREFER`, Paare davon) mit verschiedenen Werten; Orakel nach der Bytefolge | nur `Prefer` gelesen; Reihenfolge nach Eingang statt nach Bytefolge | Generator |
| 6 | Das Komma in einem Paar trennt nicht; ein offenes Paar reicht bis zum Ende der Zeile, nie in die nächste (Nr. 2) | Fälle `x="a, return=representation"`, `return="minimal`, `x="` gefolgt von einer zweiten Zeile `return=minimal`, Backslash vor dem Schlusszeichen; dazu Schleife über die Stellen eines Paares im Element | Trennung an jedem `,`; Zeilen vor dem Lesen zusammengefügt | Tabelle |
| 7 | `Rest` ab dem ersten `;` hat keine Wirkung (Nr. 2) | Schleife über Formen des Rests (`;`, `; a`, `; a=b`, `;;`, `; =`, `; "`, `; ,x`) × beide Werte; das Ergebnis gleicht dem ohne Rest (ein Komma im Rest außerhalb eines Paares beginnt ein neues Element, das der Orakel-Fall von Zeile 4 deckt) | Rest als Fehler gelesen (`Keine`) | Generator |
| 8 | Ein unlesbares Element zählt nicht, ein lesbares mit unbrauchbarem Wert zählt (Nr. 2, 3) | Schleife über jedes Byte `0x00` bis `0xFF` als erstes Zeichen des Namens, als Zeichen direkt hinter `return` und als Zeichen hinter dem Wort; Orakel nach `tchar`, `OWS`, `=`, `;`, `,`, `"` | Name mit einem Zeichen außerhalb von `tchar` als lesbar behandelt; Wert mit Rest hinter dem Wort als `Minimal` | Generator |
| 9 | `LiesRueckgabe` ändert die Eingabe nie, liefert nie einen Fehler und löst keine Panik aus (Nr. 1) | Schleife über alle Folgen der Länge 0 bis 6 über dem Alphabet {`return`, `=`, `;`, `"`, Backslash, `,`, Leerzeichen, Tabulator, `minimal`, `x`, `0xA0`}, die Eingabe vorher kopiert und danach verglichen | Sortieren oder Kürzen der Eingabe an Ort und Stelle; Indexzugriff hinter dem Ende | Generator |
| 10 | Die Antwort je Methode und Wahl trägt Status, Body und genau die Header der Tabelle von `SPEC-286` Nr. 1, nie `Vary` | Schleife über 3 Methoden × 3 Wahlen × `Dienst.Etag` an und aus × `Accept` mit den Stufen `none` und `minimal`: Menge der Header-Namen gleich der erwarteten (Gleichheit, nicht Enthaltensein), Status, Body leer oder nicht | je Zelle der Tabelle eine Mutation (`POST` · Minimal als `201`; `PATCH` · Entity als `204`; `Content-Type` an der `204`; `Vary: Prefer`; `ETag` bei aus) | Tabelle |
| 11 | `Location` und `OData-EntityId` tragen dieselbe Zeichenfolge, aus `Adresse`, nie aus `Host` (`SPEC-286` Nr. 1) | Schleife über Methoden und Wahlen mit Header `Host` und `X-Forwarded-Host` fremd; der Fake des Schreibers hält `Wurzel` und `Schluessel` jedes Aufrufs fest | `Wurzel` aus dem Header `Host`; `OData-EntityId` aus `Location` abgeleitet und gekürzt | Schleife |
| 12 | `Preference-Applied` ist genau `return=minimal` oder `return=representation`, in einer Kopfzeile, ohne Echo (`SPEC-286` Nr. 1) | Schleife über die Formen der Zeilen von Fänger 1 bis 5 mit Erfolg: Anzahl der Kopfzeilen, Wert byte-gleich | Echo des Client-Werts; die Schreibweise des Clients; zweite Kopfzeile je weitere Präferenz | Generator |
| 13 | `Preference-Applied` steht nur an der Antwort einer gelungenen Mutation mit Wahl (`SPEC-286` Nr. 3) | Schleife über 5 Methoden × 3 Wahlen × {Erfolg, jeder Fehlercode, den die Mutation erreicht} (`ODR-E6002`, `E6003`, `E6005`, `E6007`, `E6008`, `E6009`, `E6010` bei `Etag` an, `E9001` aus dem Schreiber) | Header vor der Entscheidung gesetzt; an `GET`, `DELETE` gesetzt | Schleife |
| 14 | Eine gescheiterte Mutation antwortet mit und ohne `Prefer` gleich, ohne `Preference-Applied` und ohne `OData-EntityId`; Zustand und Revision bleiben (`SPEC-287` Nr. 2) | Schleife über jeden Fehlerfall von Fänger 13 × jede Form der Zeilen aus 4, 5 und 7: Fehlerobjekt, Status und Header byte-gleich dem Lauf ohne Header, Zustand und Revision unverändert | `return=minimal` ergibt bei einem Fehler `204`; `Preference-Applied` an der Fehlerantwort; ein Wert aus `Prefer` im Text des Fehlers | Schleife |
| 15 | `Prefer` ändert nie den Zustand (`SPEC-287` Nr. 4) | Schleife über 3 Methoden × alle Formen aus 4, 5 und 7: Eintrag (`Payload`, `KnownProperties`, `Version`, `Completeness`), Revision und die Zahl jeder Operation der Fakes (`Transaktion` 1, `Put` 1, `Advance` 1; der Fake hält `Put` fest) gleich dem Lauf ohne Header | `Put` mit anderer `Version` bei `Entity`; ein zweites `Entities.Get` nach `Put`; `Advance` ausgelassen bei `Minimal` | Schleife |
| 16 | `GET`, `HEAD` und `DELETE` ignorieren `Prefer` ohne Wirkung und ohne Fehler (`SPEC-287` Nr. 3) | Schleife über die drei Methoden × alle Formen aus 4, 5 und 7: Status, Header und Body byte-gleich dem Lauf ohne Header; `DELETE` bei `return=representation` `204` | `DELETE` bei `return=representation` als `200` mit Entity; `GET` bei `return=minimal` ohne Body | Wirkung |
| 17 | Der Header wird vor Schritt (l) nicht gelesen, und kein Fehler hängt an ihm (`SPEC-287` Nr. 1) | Wirkungs-Fänger über Fänger 14; der Zugriff selbst ist ein akzeptiertes Negativ (`SPEC-289` Nr. 3) | Auswerten vor Schritt (b) mit Fehler bei unbekanntem Wert | Negativ |
| 18 | Die Entity der Antwort ist der geschriebene Eintrag, nicht der Body der Anfrage (`SPEC-288` Nr. 1) | Schleife über `PATCH` mit Teilbody (die ungenannten Properties stehen in der Antwort), `PATCH` mit `{}`, `PUT` mit fehlender nullbarer Property (`null` steht in der Antwort), `POST` × `Accept` `none` und `minimal`: der Body gleicht dem `GET` danach (bei `Etag` an mit `@odata.etag` der neuen Version) | Entity aus `EntityEingabe` gebildet; `Etag` der Version vor der Mutation; Entity nach einem zweiten `Entities.Get` | Schleife |
| 19 | Die Aufrufe des Schreibers je Antwort sind die Zahlen von `SPEC-288` Nr. 3 | Der Fake des Schreibers zählt `SchreibeEntity`, `Adresse`, `SchreibeAnzahl` je Test; Schleife über die fünf Zeilen der Tabelle × beide Werte von `Dienst.Etag` | `Adresse` bei `PATCH` ohne Wahl; `SchreibeEntity` bei `POST` · Minimal; `SchreibeAnzahl` bei `PATCH` · Entity | Schleife |
| 20 | Jeder Aufruf des Schreibers erhält `ctx` unverändert, `Accept` und `Wurzel`; `Auswahl` ist `nil`, auch mit `$select` (`SPEC-288` Nr. 2, 3) | Der Fake hält Argumente jedes Aufrufs fest; Schleife über die Stellen S24, S25, S26 × die Antworten, die sie rufen × Anfragen mit und ohne `$select` und mit `Accept` | `Auswahl` aus `$select`; `context.Background()` statt `ctx`; `Accept` nicht weitergereicht | Schleife |
| 21 | Ein Fehler eines Aufrufs des Schreibers ist `ODR-E9001` mit Rollback, ein Kontextfehler bleibt ohne Code, der Zustand bleibt (`SPEC-288` Nr. 3, `SPEC-287` Nr. 5) | Schleife über die Stellen, die die Antwort ruft (Fänger 19) × {`ErrUngueltig`, anderer Fehler, Kontextfehler}: Code, Zustand, Revision | Fehler verschluckt (Antwort ohne Entity); Rollback ausgelassen | Schleife |
| 22 | Ein leeres Kennzeichen oder ein leerer `ContentType` der `Darstellung` ist `ODR-E9001` ohne Antwort (`SPEC-288` Nr. 2) | Schleife über die Antworten mit Aufruf von `SchreibeEntity` und von `SchreibeAnzahl` × {leeres Kennzeichen, leerer Name, leerer Wert, leerer `ContentType`} | Header gesetzt trotz leerem Wert | Schleife |
| 23 | Von *n* gleichzeitigen `PATCH` mit `Entity` trägt jede Antwort den Eintrag ihrer Mutation (`SPEC-289` Nr. 2) | Wirkungs-Fänger: zählt die `@odata.etag` der Bodys (genau `EtagFuer(v+1)` bis `EtagFuer(v+n)`, jede einmal), `-count=20`; die Zusage ist eine Wirkung, nicht ein Datenrennen | zweites Lesen nach `Put` (zwei Antworten mit derselben Version); gemeinsame Variable für die Entity | Wirkung |
| 24 | Im Modus `hybrid` geht `Prefer` unverändert in den Fallback, und der Dienst setzt oder streicht `Preference-Applied` nie (`SPEC-287` Nr. 6) | Fall im Test des Hybrid-Dienstes: der Mock endet mit einem freigegebenen Code, der Fake des Upstream hält die Header fest; Schleife über die Formen aus 4 und 5; die Antwort byte-gleich der des Upstream | `Prefer` aus den Headern des Fallbacks genommen; `Preference-Applied` angefügt | Schleife |
| 25 | Der HTTP-Adapter reicht jede Kopfzeile `Prefer` in der Reihenfolge und mit jeder Schreibweise des Namens durch und schreibt Status `200` und `204` mit den Headern der Antwort unverändert | Test am HTTP-Eingang: Fake des Dienstes hält `Anfrage.Header` fest; Schleife über die Zahl der Zeilen (0 bis 3) und die Schreibweisen des Namens; Gegenrichtung: Fake liefert `Preference-Applied`, `OData-EntityId`, `Location`, `200` und `204`, der Adapter schreibt sie byte-gleich | nur die erste Kopfzeile durchgereicht; `OData-EntityId` gefiltert; `200` auf `201` gesetzt | Schleife |
| 26 | Das Handbuch nennt das Verhalten nur als gebundenes Beispiel gegen das gebaute Binary (`SPEC-200` Nr. 1) | `make handbuch-check` mit einem Beispiel je Wahl (`return=minimal` an `POST`, `return=representation` an `PATCH`) | Beispiel mit anderem Status als das Binary | Gate |

Akzeptierte Negative (Grund und Folge stehen in `SPEC-285` *Ungeprüft* und `SPEC-289` Nr. 3, nicht hier wiederholt): der Zugriff auf `Prefer` bei `GET`, `HEAD`, `DELETE` und bei Fehlern (Zeile 17, geprüft wird die Wirkung); die Länge der Zeile über die Grenze des Adapters; andere Präferenzen als `return`; ein Cache zwischen Client und Mock (kein `Vary`, `SPEC-286` Nr. 2). Die Prüfgrenze der Generatoren (Zeilen 2 bis 9) nennt der Bericht mit Grund.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (etwa Header-Werte mit NBSP) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-mock` liegt in `done/` (Trigger der Welle). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-285` bis `SPEC-289`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-mock-schreiben` (Statuscodes und Antworten der Mutationen, die der Slice umschaltet), `slice-odata-json-schreiben` (die Entity im Body der Antwort), `slice-http-adapter` (der Header erreicht den Use Case als Teil des Requests). Früheste Lage für Code: `slice-mock-schreiben` und `slice-odata-json-schreiben` in `done/`; der Slice kann **nicht** vor ihnen beginnen, weil `MockUseCase` mit Mutation nicht existiert.
- **Binary und Handbuch:** das Beispiel der DoD läuft gegen das Binary im Modus `mock` (`slice-bootstrap-m3`); der Slice schließt damit frühestens nach `slice-bootstrap-m3` und `slice-handbuch-meldungskatalog`. Ein Start vor dem Trigger der Welle ist eine Entscheidung des Users, nicht dieses Plans.
- **Was der Slice nicht braucht:** `slice-etag-versionen` und `slice-if-match` (unabhängiger Ast, `welle-etag` §5), `welle-hybrid`, `welle-betrieb`, `welle-admin-api`, `welle-expand-batch`, `welle-odata-v2`.
- **Größenmarke:** noch nicht gesetzt. Der Planner setzt sie vor dem ersten Code-Commit mit `make groessenmarke` nach `SPEC-128`; die Faktoren stehen allein dort, die Produktcode-Erwartung des Architect ist ein fester Anteil. Der Architect legt fest: `model.LiesRueckgabe` liest fremde Bytes (den Wert eines Headers, `SPEC-285` Nr. 2) und ist ein Leser im Sinn der Formel; die Entscheidung über die Klasse und die Zahl bleibt beim Planner.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Tabelle *Zusage · Fänger* liegt vor (26 Zeilen); der Planner setzt die Zahl.
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

**Raster *Operation × Argumentart*** (jede Zelle nennt ihren Ort; „Keine“ ist die Wahl ohne `Prefer`, `SPEC-285` Nr. 3):

| Operation | Header fehlt | `return=minimal` | `return=representation` | leerer oder unbekannter Wert | mehrere Präferenzen, mehrere Header, widersprüchlich | fehlgeschlagene Mutation |
|---|---|---|---|---|---|---|
| POST | `SPEC-285` Nr. 1 und 3 (Keine), `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1 | `SPEC-285` Nr. 3 und 4 (Keine), `SPEC-286` Nr. 1 | `SPEC-285` Nr. 1 bis 4 | `SPEC-287` Nr. 2 |
| PATCH | `SPEC-285` Nr. 1 und 3 (Keine), `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1, `SPEC-288` Nr. 1 | `SPEC-285` Nr. 3 und 4 (Keine), `SPEC-286` Nr. 1 | `SPEC-285` Nr. 1 bis 4 | `SPEC-287` Nr. 2 |
| PUT | `SPEC-285` Nr. 1 und 3 (Keine), `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1 | `SPEC-286` Nr. 1, `SPEC-288` Nr. 1 | `SPEC-285` Nr. 3 und 4 (Keine), `SPEC-286` Nr. 1 | `SPEC-285` Nr. 1 bis 4 | `SPEC-287` Nr. 2 |
| DELETE | `SPEC-194` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 2 |
| GET | `SPEC-198` Nr. 1 und 8 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | `SPEC-287` Nr. 3 | gilt nicht: GET ändert nichts; der Fehler des Lesens ist unabhängig vom Header (`SPEC-287` Nr. 3) |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für die Fakes von Repository und Schreiber; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nicht nebenläufig. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
