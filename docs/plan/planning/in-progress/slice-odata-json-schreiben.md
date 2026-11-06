# Slice slice-odata-json-schreiben: OData-Codec zum Schreiben von OData-JSON

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-mock](../welle-mock.md).

**Bezug:** `LH-FA-07`, `LH-FA-08`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-07.b` (`@odata.context`, `@odata.count`, `value`, `Location`) · `SPEC-168` (Vertrag von `ODataSchreiber`) · `SPEC-169` (Format der Version 4) · `SPEC-006` · `SPEC-126` (Gegenstück zum Lesen) · `SPEC-055` · `SPEC-172` Nr. 6 g und `SPEC-175` Nr. 3 (Verweis auf den Vertrag) · `ARC-004`, `ARC-011` · `architecture.md` §1 (Codec-Zugriff).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig slice-odata-json-schreiben, Worktree /Development/odata-recorder-wt-odata-json-schreiben)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Die Schnittstelle `ODataSchreiber` der Codec-Gruppe (`SPEC-168`) und ihre Umsetzung für Version 4 im Paket `jsoncodec` (`ARC-011`, `SPEC-169`) serialisieren eine Entity, eine Collection, die Anzahl und das Fehlerobjekt in die OData-V4-JSON-Darstellung mit `@odata.context`, `@odata.count` und `value`, stellen die acht Typen verlustfrei dar (`Edm.Int64` und `Edm.Decimal` ziffernweise, nie über `float64`), handeln Metadatenstufe und Zahlform aus `Accept` aus, erzeugen `Location` und `@odata.id` aus Wurzel und typisiertem Key und liefern mit den Bytes den Darstellungskopf (`Content-Type`, `OData-Version: 4.0`).

**Übernimmt (Sendung von `slice-v2-bootstrap`, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) `Accepted`):** die Schreib-Schnittstelle liefert mit den Bytes den Darstellungskopf (Content-Type und Versionskennzeichen der Darstellung); der Schreiber der Version 4 liefert `OData-Version: 4.0` (für Version 2 steht das Gegenstück in `SPEC-172` Nr. 6 g). Der Vertrag steht als `SPEC-168` und `SPEC-169` in der Spezifikation, `Codecs` im Vertrag von `MockUseCase` als `SPEC-168` Nr. 10 (der Vertrag von `MockUseCase` selbst ist noch nicht geschrieben und übernimmt diese Nummer). Kein Ausschluss dieses Slice trifft sie.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Das Lesen von Request-Bodies — `slice-odata-json-eingabe`; sonst teilten beide Richtungen einen Mechanismus, den nur eine braucht.
- `@odata.nextLink` und serverseitiges Paging — nicht in der Muss-Untermenge; der Mock antwortet mit der angeforderten Seite, Paging ist `$top` und `$skip` des Clients.
- Navigationen und eingebettete Entities — `welle-expand-batch`.
- Die Metadatenstufe `full` (`@odata.editLink`, Typ je Property) — nicht im Umfang (`SPEC-049`); die Aushandlung beantwortet sie als `minimal` und nennt das im `Content-Type` (`SPEC-169` Nr. 3).
- Die Verdrahtung im Composition Root und der Aufruf durch den Mock — die Dienste von Teil 2 der Welle (Adresse: Planner, §6); dieser Slice ändert weder `src/bootstrap` noch einen Driving Adapter noch den Katalog des Binary.

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

- [ ] Entity und Collection mit `@odata.context`, `@odata.count` und `value`, Metadatenstufe und Zahlform nach der Aushandlung der Spezifikation (`SPEC-169` Nr. 3).
- [ ] Die acht Typen verlustfrei; Round-trip: was der Codec schreibt, liest `ODataCodec` (`SPEC-126`) in dieselbe Entity.
- [ ] `Location` und `@odata.id` aus Service-Root und typisiertem Key ohne unvalidierte Host-Injektion; Fehlerobjekt im OData-V4-Format.
- [ ] `make gates` grün.
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
| `src/hexagon/model/schreiben.go` | neu | `Darstellung`, `Kopfzeile`, `EntityAntwort`, `CollectionAntwort` (`SPEC-168`); ohne `schreiben_test.go`: vier Typen ohne Verhalten, jeder Test des Pakets `jsoncodec` führt sie (Abweichung vom Entwurf, Block unverändert) |
| `src/hexagon/ports/outbound/odata_schreiber.go` | neu | Schnittstelle `ODataSchreiber` der Codec-Gruppe |
| `src/hexagon/ports/outbound/codec_satz.go` | update | Feld `Schreiber` im `CodecSatz` (`SPEC-175` Nr. 3) |
| `src/adapters/driven/odata/jsoncodec/` | update | `Schreiber` in neuen Dateien neben dem Leser: `schreiber.go` (Vertrag, vier Operationen, Ablauf), `schreiber_bau.go` (Sicht, Entity, Collection), `schreiber_kanon.go` (Art gegen Typ, kanonischer Text), `schreiber_adresse.go` (Wurzel, Kodierung), `schreiber_accept.go` (Aushandlung), `schreiber_ausgabe.go` (Escape); der Leser bleibt unberührt, seine Hilfen `datumTeile`, `ziffernWert`, `ziffer`, `hexWert` und `maxBruchteil` werden gerufen |
| `spec/spezifikation.md`, `spec/architecture.md` | update | Vertrag mit Randformen (Architect, erledigt: `SPEC-168`, `SPEC-169`, Verweise) |
| Tests (`schreiber_*_test.go` im Paket `jsoncodec`, extern) | neu | die Zeilen der Tabelle unten, Round-trip gegen `ODataCodec`; `schreiber_helfer_test.go` trägt die Fakes von `dst` und `ctx` und das Modell |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen unten **vor** dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen hat der Architect entschieden, nicht der Implementer. Hält der Implementer an einer Zelle ohne Ort an, geht sie an den Architect, und §6 trägt sie im selben Commit mit Ausgang *offen* (`AGENTS.md` §3.13).

- Ein Wert mit Escape-Folge oder Rohbytes (Strings mit Zeichen jenseits von ASCII, ungültiges UTF-8, Steuerzeichen) entsteht in den Fixtures aus Bytewerten und wird nach dem Schreiben gegengelesen (`AGENTS.md` §3.18).
- **Aufträge des Architect an den Implementer:** (1) Das Orakel jeder Zeile ist eine Tabelle oder ein Literal im Test, nie der Schreiber und nie ein zweiter Kodierer im Produktcode; der Leser (`ODataCodec`) ist das Gegenstück nur in den Zeilen 31 bis 33. (2) Der Fake von `dst` hält jeden Aufruf fest und zählt jede Operation, auch die, die der Schreiber nie rufen darf (`WriteString`, `ReadFrom`, `Close`); der Fake von `ctx` zählt `Err`, `Done` und `Deadline` und teilt sein Ereignisprotokoll mit dem Fake von `dst` (Zeile 4). (3) Die Mengen der Tabelle sind Schleifen über alle Elemente (`AGENTS.md` §3.17); die Elemente der Zeilen 11, 19, 21 und 22 stehen in der zweiten Tabelle. (4) Die Tabelle *Zusage · Mutation · roter Test* mit Raten (Zahl der Prozesse, Messender) trägt der Implementer in einen Absatz von §3 ein (`AGENTS.md` §3.10); zwei Zeilen sind `nebenläufig` (36, 37), Zeile 26 hängt an der Streuung der Iteration über eine `map` und nennt ihre Rate je Mutation.

**Entscheidungen des Architect** (2026-11-06). Der Vertrag steht in `spec/spezifikation.md` als `SPEC-168` und `SPEC-169`. Eine ADR entsteht nicht: es gibt keine Bibliothek-Frage (Standardbibliothek, [ADR-0001](../../adr/0001-implementierungssprache-go.md)), und die Zahlform und die Stufe sind Vertrag des Formats mit Grund und Folge in `SPEC-169` Nr. 3 (`AGENTS.md` §3.8). Je Auftrag: Entscheidung, Alternative, Grund, Folge im Betrieb.

1. **Eine Schnittstelle für beide Versionen.** `ODataSchreiber` hat fünf Methoden (`SchreibeEntity`, `SchreibeCollection`, `SchreibeAnzahl`, `SchreibeFehler`, `Adresse`); `CodecSatz.Schreiber` trägt sie. *Alternative:* je Operation ein Port oder ein Wert `Antwort` mit Art. *Grund:* der Mock ruft genau diese vier Antworten und die Adresse; `SPEC-172` Nr. 6 (V2) füllt dieselben fünf. *Folge:* eine sechste Antwort (etwa `$metadata` zu schreiben) ändert die Schnittstelle.
2. **`Accept` ist ein Argument, die Aushandlung steht im Schreiber.** Ein Satz von Regeln (`SPEC-169` Nr. 3), keine Ablehnung, kein `406`. *Alternative:* der HTTP-Adapter oder der Mock handelt aus und übergibt eine Stufe. *Grund:* der Mock verzweigt nie nach der Version (`SPEC-175` Nr. 3), V2 wertet `Accept` nicht aus (`SPEC-174` Nr. 5), und ein Fehlercode „nicht annehmbar“ verlängerte den Katalog des Binary. *Folge:* `$format=…;odata.metadata=none` wirkt nicht (`SPEC-162` Nr. 8, berichtigt); `full` wird als `minimal` beantwortet.
3. **Zahlform folgt dem Client.** Zeichenkette für `Edm.Int64`, `Edm.Decimal` und `@odata.count` genau bei `IEEE754Compatible=true`, sonst Zahl. *Alternative:* immer Zeichenkette (wie V2, `SPEC-172` Nr. 6 f). *Grund:* der Standard von OData V4 ist die Zahl; ein Client, der Rundung vermeiden will, nennt den Parameter, und der Schreiber rundet nie. *Folge:* zwei Wege je Typ und eine Achse mehr in den Fängern (Zeilen 22, 23).
4. **`@odata.id` in der Stufe `minimal`.** `LH-FA-07.b` nennt die Entity-ID; der Mock schreibt sie in jede Entity außer bei `none`. *Alternative:* nur bei `full`. *Grund:* `full` ist nicht im Umfang, und `Location` und `@odata.id` müssen dieselbe Zeichenfolge sein (Zeile 20). *Folge:* ein wenig größere Bodies.
5. **Genau ein `Write`, Aufbau vor dem Schreiben.** *Alternative:* Strom mit `json.Encoder`. *Grund:* der Schreiber entscheidet über Erfolg und Fehler vor dem ersten Byte; ein Verstoß im letzten Element lässt `dst` leer. *Folge:* Speicher linear mit der Ausgabe (die Seite begrenzt `$top`), akzeptiert in `SPEC-168` Nr. 6.
6. **`Wurzel` wird geprüft, nicht vertraut.** Eine enge Form (`SPEC-169` Nr. 2) statt `net/url.Parse`; der Schreiber liest keinen Header. *Grund:* `Wurzel` steht im Body und in `Location`; die Quelle (Konfiguration, nie `Host`) sichert der Aufrufer, die Form sichert der Schreiber gegen Kopfzeilen- und JSON-Injektion. *Folge:* ein Hostname jenseits von ASCII muss als Punycode kommen.
7. **Das Fehlerobjekt wird gegen die Tabelle der Codes geprüft** (`SPEC-168` Nr. 5 f). *Alternative:* beliebiger Text. *Grund:* `message` ist ein fester Text je Code und nie ein Wert aus Request oder Daten (`SPEC-109` Nr. 3). *Folge:* ein Admin-Fehler mit Korrelationsfeldern (`SPEC-006`) geht nicht durch diesen Schreiber; ihn schreibt der HTTP-Adapter.
8. **Keine neuen Fehlercodes, kein Schlüssel, kein Kommando.** Jeder Verstoß ist `model.ErrUngueltig` (`ODR-E9001`, Programmfehler des Aufrufers). Das Binary ändert sich nicht; die Handbuch-Pflicht aus `welle-mock.md` §4 trifft diesen Slice nicht.
9. **Gate-Kette.** Keine Änderung nötig (Bedingung aus §4 geprüft): `jsoncodec`, `outbound` und `model` existieren, `.a-check.yml` und `versionstrennung_test.go` decken das Paket; Zeile 38 nennt sie als Fänger.
10. **Nachzüge im selben Commit** (`AGENTS.md` §3.9): `SPEC-125` und `SPEC-126` (der Name statt „eigene Schnittstelle“), `SPEC-162` Nr. 8 (die Stufe gehört dem Schreiber, nicht dem HTTP-Adapter), `SPEC-006` (Verweis), `SPEC-172` Nr. 6 g und Einleitung (Nummer des Vertrags, Auftrag 2), `SPEC-175` Nr. 3 (Name im Satz), `architecture.md` (`ARC-004`, `ARC-011`, §1, Tabelle der Ports), die Historie der Spezifikation.
11. **Zahlen für die Marke** (`SPEC-128`): 42 Zeilen (*a* = 2 nebenläufig, *b* = 40 andere), *f* = 1 (Fake von `dst`; `ctx` ist kein Port), *p* = 1000 Zeilen Go ohne Tests (Spanne 800 bis 1250: Typen und Port 110, Prüfung und Ablauf 250, Werte 220, Wurzel und Kodierung 130, Aushandlung 90, Ausgabe 200; Vergleich `urlparser` 1230 am Ende bei 1700 erwartet), *S* = 85·2 + 80·40 + 200·1 + 1000 = **4570**, Marke = 100 · ⌈135·4570 / 10000⌉ = **6200** (`make groessenmarke` nennt beide, die Marke setzt der Planner). **Schnittfrage** (*S* ≥ 2223): gestellt, **kein Schnitt, `schnitt-grund` im Block.** Geprüft: (a) Aushandlung (Zeilen 22, 23), (b) Adresse und Wurzel (19 bis 21), (c) Werte (10, 11), (d) Antworten und Kopf (14, 16 bis 18, 24 bis 30). Jede Hälfte braucht `Darstellung`, die Prüfung der Argumente und die Ausgabe von Zeile 24 bis 27; (a) ändert Kopf und Zahlform jeder Ausgabe, (b) steht in jeder Entity. Der Leser fremder Bytes entfällt: der Slice schreibt sein eigenes Format.

Tabelle *Zusage · Fänger* (Spalten nach `SPEC-128` Nr. 3 a; „Menge“ ist die Klasse der Zusage aus dem Wortlaut, `AGENTS.md` §3.17; eine Nummer ohne Nennung der Spezifikationsstelle gehört in den Zeilen 1 bis 17 und 34 bis 37 zu `SPEC-168`, in den Zeilen 19 bis 33 zu `SPEC-169`; die Zeilen 10, 11, 18 und 20 nennen beide). Test-Pakete: `jsoncodec` (Namen der Fänger), `model`, `outbound`.

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | `ctx` nil und `dst` nil sind `model.ErrUngueltig` ohne `Code` und ohne `*DatenFehler` (Nr. 1 a); nichts in `dst`, `Darstellung` der Nullwert. Menge 8: 2 Argumente × 4 Operationen mit `ctx` und `dst` | `TestSchreibeNilArgumente`: Schleife Operation × Argument, Zähl-Fake von `dst` | die nil-Prüfung von `dst` fehlt in `SchreibeFehler`; `ctx` nil in `SchreibeAnzahl` erst hinter der Prüfung von `n` | andere |
| 2 | Ein ungültiges Argument ist `ErrUngueltig`, auch bei beendetem Kontext (Nr. 1). Menge 8: 4 Operationen × 2 Zustände (abgebrochen, Frist abgelaufen), das ungültige Argument je Operation: Set unbekannt, Set unbekannt, `n` gleich -1, leeres Fehlerobjekt | `TestArgumentVorKontext` | `ctx.Err()` vor der Prüfung der Argumente gefragt | andere |
| 3 | Beendeter Kontext bei gültigen Argumenten: `ctx.Err()` unverändert (`==`), ohne `Code`, `dst` leer (Nr. 1 c). Menge 12: 4 Operationen × 3 Zustände (`Canceled`, `DeadlineExceeded`, abgebrochen mit Ursache `WithCancelCause`: das Ergebnis ist `context.Canceled`, nicht die Ursache); Gegenfall: aktiver Kontext schreibt | `TestKontextBeendet`: `err == ctx.Err()`, nicht `errors.Is` | der Fehler mit `fmt.Errorf("%w")` verpackt; `context.Cause(ctx)` statt `ctx.Err()` | andere |
| 4 | Der Kontext wird genau einmal über `ctx.Err()` gefragt, nach den Argumenten und dem Aufbau, unmittelbar vor dem `Write` (Nr. 1). Menge 8: 4 Operationen × {gültige, ungültige Argumente}; gültig: Ereignisfolge `[Err, Write]`, ungültig: `Err` 0-mal; `Done` und `Deadline` nie | `TestKontextEinmalVorWrite`: Ereignisprotokoll, das Fake von `ctx` und Fake von `dst` teilen | `Err` zweimal gefragt; `Err` nach dem `Write`; `ctx.Done()` statt `Err` | andere |
| 5 | `m.Version()` ungleich `4.0` ist `ErrUngueltig` (Nr. 2). Menge 12: 4 Versionen (`2.0`, leer als Nullwert, `3.0`, `4.0` mit Leerzeichen) × 3 Operationen mit Modell (Entity, Collection, Adresse); Gegenfall `4.0` | `TestModellVersion` | die Prüfung entfällt; nur gegen `2.0` verglichen | andere |
| 6 | `Set` leer, unbekannt, in anderer Schreibweise (`products`), mit Namensraum (`Ns.Products`), ein Set, dessen Typ das Modell nicht auflöst, und das Nullwert-Modell sind `ErrUngueltig` (Nr. 2). Menge 18: 6 Formen × 3 Operationen (Entity, Collection, Adresse) | `TestSetUngueltig` | Vergleich ohne Beachtung der Schreibweise; ein Set ohne Typ wird durchgelassen | andere |
| 7 | Der Typ jeder Entity ist ein nicht abstrakter Typ des Sets oder von ihm abgeleitet; die Kette endet bei jedem Modell (Nr. 3 a). Verstöße 8: leer, unbekannt, abstrakt, außerhalb der Ableitung, Ring in `BaseType` ohne den Typ des Sets der Länge 1, 2, 3, 4; Gegenfälle 3: Typ des Sets, abgeleitet in Tiefe 1 und 2 (mit `@odata.type`). Stellen 4 (`SchreibeEntity`; erstes, mittleres, letztes von 3 Elementen von `SchreibeCollection`): 32 Zellen | `TestEntityTyp`: Schleife Verstoß × Stelle, Zeitgrenze des Tests für die Ringe | die Kette ohne Besuchszähler (Endlosschleife im Ring); `abstract` nicht geprüft; nur das erste Element der Collection geprüft | andere |
| 8 | `Schluessel` nicht leer, gültiges UTF-8, ohne U+0000, ohne Längengrenze (Nr. 3 b). Verstöße 5: leer, ungültiges UTF-8, U+0000 am Anfang, in der Mitte, am Ende; Gegenfälle 3: `'O''Hara'`, Zeichen jenseits von ASCII, 100000 Zeichen; Stellen 4 wie Zeile 7: 20 Zellen | `TestSchluessel` | U+0000 nur am Anfang geprüft; eine Längengrenze von 1024 | andere |
| 9 | Die Namen von `Properties` sind wirksame Properties unterstützten Typs, Zeichen für Zeichen, und kommen höchstens einmal vor (Nr. 3 c). Verstöße 9: unbekannt, Navigation-Property, Property nicht unterstützten Typs, leer, andere Schreibweise (`name`), zweimal mit gleichem Wert, zweimal mit verschiedenem Wert, zweimal nicht benachbart, zweimal an erster und letzter Stelle; Gegenfälle: `Properties` `nil` und leer; Stellen 4: 36 Zellen | `TestPropertyNamen` | das Duplikat nur bei gleichem Wert erkannt; Vergleich ohne Beachtung der Schreibweise; nur benachbarte Duplikate | andere |
| 10 | `Wert.Art` gehört zum `EdmType`; `null` nur bei nullbarer Property und mit Text `""`; `redigiert` und jede andere Art sind `ErrUngueltig` (Nr. 4, `SPEC-169` Nr. 4). Menge 112: 8 Typen × 8 Arten (ohne `null`, `redigiert`) = 64 Paare (8 gültig, 56 ungültig); `null` nicht nullbar 8 ungültig, nullbar 8 gültig; `null` mit Text 8; `redigiert` 8; leere und unbekannte Art 16: 96 ungültig, 16 gültig | `TestArtGegenTyp`: Modell mit acht Properties je nullbar und nicht nullbar, Schleife über die Paare | die Paarung nur auf der Diagonale der ersten fünf Typen geprüft; `redigiert` als `null` geschrieben | andere |
| 11 | `Wert.Text` ist die kanonische Schreibweise der Art, jede andere ist `ErrUngueltig`; der Schreiber normiert nie (Nr. 4, `SPEC-169` Nr. 4). Menge 140 Elemente: 33 gültig, 107 ungültig (je Art in der zweiten Tabelle); Stellen 4 (Entity: erste und letzte Property; Collection: erstes und letztes Element): 560 Zellen | `TestTextKanonisch`: Schleife Art × Element × Stelle; ungültige Elemente aus Bytewerten (`AGENTS.md` §3.18) | `-0` für `int32` angenommen; die Grenze 1024 mit `>=` statt `>`; Guid nur am Anfang auf Kleinbuchstaben geprüft; `$` statt `\z` am Ende (ein Zeilenende bleibt unentdeckt) | andere |
| 12 | `Auswahl`: `nil` und leer heißen keine Auswahl; jedes Element ist `*` oder eine wirksame Property oder Navigation-Property des Typs des Sets, sonst `ErrUngueltig`; die Prüfung gilt in beiden Stufen (Nr. 5 a). Ungültig 5: unbekannt, leer, andere Schreibweise, Property nur eines abgeleiteten Typs, Element mit Leerzeichen; gültig 6: `*`, Navigation-Property, dasselbe Element zweimal, `nil`, leeres Slice, ein Element. Menge 40 = 5 × Stelle in der Liste (erste, letzte) × Operation (Entity, Collection) × Stufe (minimal, none) | `TestAuswahl` | bei `none` nicht geprüft; nur das erste Element der Liste geprüft | andere |
| 13 | `Anzahl{Vorhanden}` wahr: `Wert` ab 0, sonst `ErrUngueltig`; `Vorhanden` falsch: `Wert` ohne Wirkung; `SchreibeAnzahl`: `n` ab 0 (Nr. 5 b, e). Menge 12: `Wert` -1, 0, 1, `MaxInt64` bei wahr (4); 0, 5, -1 bei falsch (3); `n` -1, `MinInt64`, 0, 1, `MaxInt64` (5) | `TestAnzahlGrenzen` | `Wert` auch bei `Vorhanden` falsch geprüft; `> 0` statt `>= 0` | andere |
| 14 | `SchreibeFehler` nimmt genau `NeuesFehlerobjekt(c)` zu einem Code der Tabelle (Nr. 5 f). Menge: gültig `len(model.AlleCodes())` (36 am Start); ungültig 8: leeres Objekt, unbekannter Code `ODR-E0000`, Code in Kleinbuchstaben, Code mit Leerzeichen, `message` abweichend, `message` mit Text einer Anfrage, `target` abweichend, `target` leer | `TestFehlerobjektTabelle`: Schleife über `AlleCodes()` und die Abweichungen je Feld | nur `code` geprüft; `message` nicht verglichen; die Schleife über die ersten Codes | andere |
| 15 | Aufbau vor dem Schreiben: ein Verstoß an irgendeiner Stelle der Eingabe lässt `dst` leer, auch hinter gültigen Teilen; zwei Verstöße zugleich sind `ErrUngueltig` (Nr. 1 b, 6). Menge 8: Collection mit 3 Elementen, Verstoß im 1., 2., 3. Element × Art (Typ, Wert) = 6; Verstoß erst in der letzten Property des letzten Elements; Paar Set unbekannt und Typ abstrakt | `TestNichtsBeiVerstoss`: Zähl-Fake von `dst`, `Write` 0-mal | die Elemente laufend nach `dst` geschrieben | andere |
| 16 | Genau ein `Write`, sonst keine Methode von `dst`; die Ausgabe ist nie leer (Nr. 6). Menge 8: 4 Operationen × 2 Größen (klein; 1000 Elemente, über 100 KiB), `WriteString`, `ReadFrom`, `Close` je 0 | `TestEinWrite` | in Blöcke zu 4096 Bytes geteilt; `io.WriteString` genutzt | andere |
| 17 | Fehler von `dst`: der Fehler unverändert (`==`), ein kurzer `Write` ohne Fehler `io.ErrShortWrite`, `Darstellung` der Nullwert (Nr. 6). Menge 14: 4 Operationen × {Sentinel-Fehler, kurzer `Write` mit n gleich 0, mit n gleich Länge-1} = 12; Paare: ungültiges Argument und Fehler von `dst` (`ErrUngueltig`, `Write` 0-mal), beendeter Kontext und Fehler von `dst` (Fehler des Kontexts, `Write` 0-mal) | `TestDstFehler` | n kleiner als die Länge ignoriert; der Fehler verpackt | andere |
| 18 | Darstellungskopf jeder Ausgabe: `Content-Type` und `OData-Version: 4.0` (Nr. 7, `SPEC-169` Nr. 3, 7). Menge 10: Entity und Collection × Stufe (2) × Zahlform (2) = 8; Anzahl; Fehlerobjekt; der String steht als Literal im Test | `TestDarstellungskopf` | `charset` fehlt; `IEEE754Compatible=true` bei Zahlform `Zahl`; Version `4.01`; Fehlerobjekt mit `text/plain` | andere |
| 19 | Kodierung von `Set` und `Schluessel` in der Adresse (`SPEC-169` Nr. 2): `A-Za-z0-9-._~',=` unverändert, jedes andere Byte als `%` und zwei Hex-Ziffern in Großbuchstaben; `net/url.PathUnescape` ergibt den Text zurück. Menge 130 Zeichen: U+0001 bis U+007F (127) und je ein Zeichen mit 2, 3, 4 Bytes; Stellen 2 (Schlüssel, Set eines Modells von Hand): 260 Zellen | `TestAdresseKodierung`: Schleife; die erlaubte Menge steht als Tabelle im Test | `~` kodiert; Hex-Ziffern klein; `+` roh; `%` roh; `,` kodiert; ein Mehrbyte-Zeichen nur im ersten Byte kodiert | andere |
| 20 | `Adresse(m, w, s, k)`, `@odata.id` der Entity und `@odata.id` des Elements einer Collection sind dieselbe Zeichenfolge (Nr. 8, `SPEC-169` Nr. 2); der Kontext trägt `Set` roh. Menge 33: 3 Stellen × 10 Schlüssel (`42`, `'O''Hara'`, `Id=1,Region='EU'`, Zeichen jenseits von ASCII, mit Leerzeichen, `/`, `?`, `#`, `%`, `+`) = 30; Set mit Zeichen jenseits von ASCII (`Prodükte`) in der Adresse kodiert, im Kontext roh: 3 Stellen | `TestAdresseGleich` | `@odata.id` mit eigenem Kodierer; das Set nicht kodiert; der Kontext mit kodiertem Set | andere |
| 21 | `Wurzel` hat die Form von `SPEC-169` Nr. 2, sonst `ErrUngueltig`. Menge: gültig 11, ungültig 44 (zweite Tabelle) × 3 Operationen mit `Wurzel` (Entity, Collection, Adresse) = 165 Zellen | `TestWurzel` | `ftp` angenommen; `//` hinter dem Host; Userinfo; `%2E` als Segment; Länge 2049; Host mit `ä`; Port ohne Ziffern | andere |
| 22 | Aushandlung: Stufe und Zahlform aus `Accept` nach `SPEC-169` Nr. 3, nie ein Fehler. Achsen (zweite Tabelle): `odata.metadata` 8 Werte × `IEEE754Compatible` 7 Werte = 56 im ersten Bereich; Stelle 4; Rand 4; Schreibweise 3; Sonderformen 6; doppelter Parameter 2 | `TestAushandlung`: Schleife über die Achsen, das Orakel ist die Tabelle des Tests | Parameter aus dem letzten Bereich; `full` als Stufe geschrieben; Wert `1` als wahr; `q` ausgewertet; Fehler bei unbekanntem Medientyp | andere |
| 23 | Zahlform: `int32` ist immer eine Zahl; `int64`, `decimal` und `@odata.count` sind bei `Zahl` Zahlen, bei `Zeichenkette` Strings, ziffernweise wie übergeben (`SPEC-169` Nr. 4, 6). Menge 14: 3 Arten × 2 Formen × 2 Operationen = 12; Anzahl × 2 Formen in `SchreibeCollection` | `TestZahlform` | `int32` als String bei `Zeichenkette`; ein Wert über `float64` geführt (`9223372036854775807` rundet); `@odata.count` immer Zahl | andere |
| 24 | Bytes der Entity (`SPEC-169` Nr. 5). Menge 12: Stufe (minimal, none) × {ohne Auswahl, `Id,Name`, `*`, abgeleiteter Typ, `Properties` `nil`, ein `null`-Wert} | `TestEntityBytes`: Literale aus Bytewerten | `@odata.type` für den Typ des Sets; `@odata.id` vor `@odata.type`; ein Leerzeichen hinter `:`; `/$entity` fehlt; bei `none` ein Kontext | andere |
| 25 | Bytes der Collection (`SPEC-169` Nr. 6): `value` immer vorhanden, bei keinem Element `[]`, nie `null`; `@odata.count` vor `value`. Menge 24: Stufe (2) × Entities {`nil`, leer, 1, 3} × Anzahl {fehlt, 0, 42} | `TestCollectionBytes` | `"value":null` bei `nil`; `@odata.count` hinter `value`; die Anzahl bei `Vorhanden` falsch mit `Wert` 5 geschrieben; Elemente ohne `,` | andere |
| 26 | Die Properties stehen in der Reihenfolge des Modells, unabhängig von der Übergabe (Nr. 3 c). Menge 48: 24 Permutationen von 4 Properties einer Entity, 24 desselben Typs mit geerbten Properties (die Reihenfolge des Modells), je 5 Aufrufe im Prozess | `TestPropertyReihenfolge` | die Reihenfolge der Übergabe; Iteration über eine `map`; Sortierung nach Name | andere |
| 27 | Escapes (`SPEC-169` Nr. 1): `"`, `\`, U+0000 bis U+001F als Escape aus Backslash, `u` und vier Hex-Ziffern in Kleinbuchstaben (keine Kurzform `\n`); `/`, U+007F, U+2028, U+2029 und Zeichen jenseits von ASCII unverändert. Menge 57: 41 Zeichen im String-Wert (32 + 2 + 4 + 3 Mehrbyte) + 4 Zeichen (`"`, `\`, U+0001, U+001F) an 4 weiteren Stellen (Name einer Property aus einem Modell von Hand, `Set` im Kontext, Element der Auswahl, `Typ` in `@odata.type`) | `TestEscapes` | die Kurzform `\n`; `/` escaped; U+007F escaped; U+2028 escaped; Hex-Ziffern groß; der Name nicht escaped | andere |
| 28 | Eine Property, die `Properties` nicht nennt, fehlt im Body (kein `null`); `Uebersprungen` wird nie geschrieben (Nr. 3 c, d). Menge 10: fehlend nullbar und nicht nullbar (2) + Gründe `typ`, `navigation`, `offen` (3) = 5 × 2 Operationen | `TestNichtGenannt` | die fehlende nullbare Property als `null` geschrieben; der Name aus `Uebersprungen` als `null` geschrieben | andere |
| 29 | Bytes des Fehlerobjekts: `code`, `message`, `target` in dieser Reihenfolge, alle Strings, ohne Leerraum (`SPEC-169` Nr. 7). Menge: ein Fall je Code der Tabelle (36 am Start) | `TestFehlerBytes`: Erwartung aus `NeuesFehlerobjekt` und einem Literalformat im Test | `target` fehlt; `message` vor `code`; der Status im Body | andere |
| 30 | Bytes der Anzahl: die Ziffern, ohne Zeilenende (`SPEC-169` Nr. 7). Menge 5: 0, 1, 9, 10, `MaxInt64` | `TestAnzahlBytes` | ein Zeilenende angehängt; Tausendertrennung | andere |
| 31 | Rundlauf: was der Schreiber der Stufe `minimal` schreibt, liest `ODataCodec` mit demselben Modell in dieselbe Entity (`SPEC-169` Nr. 8). Menge 132 + 28: jedes gültige Element von Zeile 11 (33) × Zahlform (2) × Operation (Entity, Collection mit 3 Elementen) = 132; `null` bei nullbarer Property je Typ (8) × 2 Operationen = 16; abgeleiteter Typ Tiefe 1 und 2 (2) × 2 = 4; leere Collection (2 Formen); Anzahl 0, 1, `MaxInt64` × 2 Formen = 6 → 28 | `TestRundlauf`: Schleife; das Gegenstück ist `jsoncodec.New()` | `Edm.Int64` über `float64`; `Edm.Decimal` mit Exponent; `Edm.DateTimeOffset` ohne Bruchteil; `Edm.Guid` in Großbuchstaben (Prüfung entfernt) | andere |
| 32 | Der Kontext der Auswahl ist für den Leser gültig (`SPEC-169` Nr. 5, `SPEC-126` Nr. 7). Menge 10: Auswahl {`Id,Name`, `*`, ein Element zweimal, Navigation-Property} × 2 Operationen = 8; ohne Auswahl × 2 | `TestRundlaufAuswahl` | Elemente durch `, ` getrennt; Auswahl ohne Klammer; `/$entity` bei der Collection | andere |
| 33 | Stufe `none`: Body ohne Kontext, `@odata.type` und `@odata.id`, mit `@odata.count`; der Leser lehnt ihn mit `kontext`, Ort `/@odata.context` ab (`SPEC-169` Nr. 5, 6, 8). Menge 4: Entity, abgeleiteter Typ, Collection ohne Anzahl, Collection mit Anzahl | `TestStufeNone` | bei `none` ein Kontext geschrieben; `@odata.count` bei `none` weggelassen | andere |
| 34 | Gleiche Argumente ergeben gleiche Bytes und `Darstellung`, geänderte ergeben geänderte; kein Cache (Nr. 9). Menge 4 Operationen × {100 gleiche Aufrufe; je ein geändertes Feld: Wert, `Schluessel`, `Set`, `Accept`} | `TestWiederholung` | ein Cache nach `Schluessel`; die Ausgabe hängt an der Zahl der Aufrufe | andere |
| 35 | Der Schreiber ändert kein Argument (Nr. 9). Menge 8: 4 Felder (`Properties` in einer Reihenfolge, die nicht die des Modells ist, `Auswahl`, `Accept`, `Uebersprungen`) tief gleich vor und nach dem Aufruf, an `SchreibeEntity` und `SchreibeCollection` | `TestArgumenteUnveraendert` | `sort.Slice` auf `Properties` des Aufrufers; `Auswahl` in place sortiert | andere |
| 36 | Gleichzeitige Aufrufe beeinflussen sich nicht (Nr. 9): jedes Ergebnis gleich dem des einzelnen Aufrufs. Menge 16 Goroutinen × 4 Operationen, 50 Läufe je Goroutine, eigene Argumente je Goroutine | `TestGleichzeitig` (Wirkungs-Fänger: zählt die Abweichungen) | ein Puffer als Variable des Pakets; ein gemeinsamer Cache | nebenläufig |
| 37 | Kein Datenrennen im Schreiber (Nr. 9) | `TestGleichzeitig` unter `make test-race` | ein ungeschützter Zähler im Paket (Rate je Prozess vom Implementer) | nebenläufig |
| 38 | `*Schreiber` erfüllt `outbound.ODataSchreiber`; das Paket importiert Domain, Ports und Standardbibliothek, keinen Adapter und keine Fremdbibliothek (`SPEC-168` Kopf) | `var _ outbound.ODataSchreiber = (*Schreiber)(nil)` (Übersetzung), `make a-check`, `versionstrennung_test.go` | eine Methode entfernt (Übersetzung rot); ein Import von `jsoncodecv2` (a-check und `versionstrennung_test.go` rot) | andere |
| 39 | Keine Verdrahtung, keine neuen Codes: `src/bootstrap` und `src/adapters/driving/` bleiben unverändert, `model.AlleCodes()` hat dieselben Einträge (Entscheidung 8) | benannte Ausgabe: `git diff --stat <Basis>..HEAD -- src/bootstrap src/adapters/driving` ist leer (Reviewer, Verifier); die bestehenden Tests der Tabelle der Codes im Modell | ein Code `ODR-E6007` für `406` in `codeZeilen()` (die Tests der Tabelle werden rot) | andere |
| 40 | Akzeptiertes Negativ: ein Typed-nil als `dst` besteht die nil-Prüfung und bricht bei der ersten Nutzung ab | – (Grund und Folge: `SPEC-168` *Ungeprüft*; Programmfehler des Aufrufers) | – | andere |
| 41 | Akzeptiertes Negativ: `Schluessel` wird nicht gegen die Schlüssel-Properties, `Auswahl` nicht gegen die Properties verglichen | – (Grund und Folge: `SPEC-168` Nr. 3 b und *Ungeprüft*; Zeile 31 prüft beides für die gültigen Eingaben des Tests) | – | andere |
| 42 | Akzeptiertes Negativ: `Accept` mit Gewicht `q`, Parameterwert in Anführungszeichen und Komma im Parameter; `$format` | – (Grund und Folge: `SPEC-169` *Ungeprüft*; Folge: die Vorgabe, nie ein ungültiger Body; Zeile 22 hält die Vorgabe) | – | andere |

Elemente der Mengen (zweite Tabelle; *Menge · Achsen mit Elementzahl · eine Mutation je Achse, die die Tabelle oben nicht nennt*, `AGENTS.md` §3.17).

| Menge | Achsen · Elemente je Achse | Mutation je Achse (nicht in der Tabelle oben) |
|---|---|---|
| Zeile 11, Text je Art | **Art** 8 mit (gültig/ungültig): `string` 6/4, `boolean` 2/8, `int32` 4/17, `int64` 4/14, `decimal` 7/22, `guid` 2/10, `date` 4/14, `datetimeoffset` 4/18; **Stelle** 4; **Randform** je Art (Leerraum links und rechts, Zeilenende, Groß- und Kleinschreibung, Länge an der Grenze ±1, Vorzeichen, Ziffernform, Ziffer jenseits von ASCII, Exponent, Bruchteil). `string` gültig: leer, `a`, U+00E9, `e` mit U+0301, U+0000, U+10000; ungültig: `80`, `C0 80`, `ED A0 80`, `C3` am Ende (Hex-Bytes). `boolean` gültig `true`, `false`; ungültig `True`, `TRUE`, `1`, `0`, leer, ` true`, `true` und LF, `yes`. `int32` gültig `0`, `-1`, `2147483647`, `-2147483648`; ungültig `2147483648`, `-2147483649`, `12345678901`, `-0`, `+1`, `01`, `00`, `-01`, `1.0`, `1e2`, leer, ` 1`, `1 `, `1` und LF, U+0661, `0x1`, `--1`. `int64` gültig `0`, `-1`, Höchst- und Tiefstwert; ungültig Höchstwert+1, Tiefstwert-1, `99999999999999999999`, `-0`, `+1`, `01`, `1.0`, `1e2`, leer, ` 1`, `1 `, `1` und LF, U+0661, `-`. `decimal` gültig `0`, `-1`, `0.5`, `-0.5`, `10`, `123456789012345678901234567890.123456789`, 1024 Zeichen; ungültig `-0`, `0.0`, `-0.0`, `1.50`, `1.`, `.5`, `00.5`, `01`, `+1`, `1e2`, `1E2`, leer, ` 1`, `1 `, `1` und LF, `1,5`, `NaN`, `INF`, `-`, `1.5.5`, 1025 Zeichen, U+0661. `guid` gültig `0f8fad5b-d9cb-469f-a165-70867728950e`, `00000000-0000-0000-0000-000000000000`; ungültig ganz in Großbuchstaben, ein Großbuchstabe, ohne Bindestriche, in Klammern, 35 und 37 Zeichen, Bindestrich verschoben, `g` statt Hex, Leerzeichen am Rand, LF am Ende. `date` gültig `0001-01-01`, `9999-12-31`, `2024-02-29`, `2000-02-29`; ungültig `0000-01-01`, `10000-01-01`, `2023-02-29`, `1900-02-29`, `2024-13-01`, `2024-00-01`, `2024-01-00`, `2024-04-31`, `2024-1-5`, `20240105`, `+2024-01-01`, `2024-01-01T00:00:00Z`, ` 2024-01-01`, `2024-01-01` und LF. `datetimeoffset` gültig `0001-01-01T00:00:00Z`, `9999-12-31T23:59:59Z`, `2020-01-01T00:00:00.5Z`, `2020-01-01T00:00:00.123456789012Z`; ungültig mit Versatz `+01:00`, `z`, `t`, Leerzeichen statt `T`, ohne Sekunden, `.0Z`, `.50Z`, 13 Ziffern, Sekunde 60, Stunde 24, Minute 60, Monat 13, 29. Februar 2023, Jahr 0000, ohne `Z`, `.Z`, Leerzeichen am Rand, LF am Ende | **Art:** `date` ohne die Regel der Jahrhunderte (1900-02-29 gültig); **Stelle:** nur die erste Property der Entity geprüft; **Randform:** ein Wert mit LF am Ende wird als `datetimeoffset` angenommen; die Grenze von `decimal` bei 1023 statt 1024 |
| Zeile 21, `Wurzel` | **gültig 11:** `http://h`, `https://h`, `http://h:8080`, `http://[::1]:8080`, `http://h/a`, `http://h/a/b`, `http://h/%C3%A4`, `http://127.0.0.1:1`, `http://h/a.b`, `http://h/...`, Länge 2048. **ungültig 44:** leer, `ftp://h`, `HTTP://h`, `Https://h`, `//h`, `h`, `http://`, `http:///a`, `http://h/`, `http://h//a`, `http://h/a/`, `http://u@h`, `http://u:p@h`, `http://h?x=1`, `http://h#f`, `http://h/a?b`, `http://h/a#b`, ein Rückstrich im Pfad, ein Leerzeichen im Pfad, Leerzeichen vor `http`, `ä` im Pfad, `ä` im Host, `http://h/%`, `%1`, `%GG`, `http://h/.`, `http://h/..`, `http://h/a/./b`, `http://h:`, Port mit 6 Ziffern, Port `80a`, zwei Ports, `http://[::1`, `http://[]`, `http://[g::1]`, `http://h%41`, LF, CR LF mit Kopfzeile, NUL und DEL hinter dem Pfad, Länge 2049, `http://h/%2E`, `http://h/%2e%2e`, `http://h/.%2E`. **Operation** 3 | **Operation:** `Adresse` prüft die Wurzel nicht; **Zeichen:** NUL nicht erkannt; Länge: Grenze bei 2047; ein Host in eckigen Klammern mit `g` angenommen |
| Zeile 19, Zeichen der Adresse | **Zeichen** 130 (U+0001 bis U+007F, je ein Zeichen mit 2, 3, 4 Bytes); **Stelle** 2 (Schlüssel, Set); **Schreibweise** der Hex-Ziffern 1 | **Zeichen:** U+007F roh; **Stelle:** das Set wird kodiert, der Schlüssel nicht; **Schreibweise:** ein Byte unter 0x10 mit einer Ziffer (`%9`) |
| Zeile 22, Aushandlung | **`odata.metadata`** 8 (`none`, `NONE`, `None`, `minimal`, `full`, `foo`, leer, fehlt); **`IEEE754Compatible`** 7 (`true`, `TRUE`, `false`, `1`, `"true"` in Anführungszeichen, leer, fehlt); **Stelle** 4 (erster Bereich; zweiter Bereich `application/json` ignoriert; Bereich anderen Typs davor; zweite Kopfzeile); **Rand** 4 (Leerzeichen oder Tab um `;` und um `=`); **Schreibweise** 3 (`Application/JSON`, `ODATA.METADATA`, `ieee754compatible`); **Sonderformen** 6 (`nil`, leeres Slice, leerer String, `*/*`, `application/*`, `text/html`); **doppelter Parameter** 2 (`none` vor `minimal`, `minimal` vor `none`) | **`odata.metadata`:** der Wert `none` mit Leerzeichen dahinter als `none`; **`IEEE754Compatible`:** `True` als falsch; **Stelle:** der erste Bereich mit `application/*` gilt; **Rand:** ein Tab vor dem Namen nicht entfernt; **Schreibweise:** `application/JSON` nicht erkannt; **doppelter Parameter:** das letzte gilt |
| Zeilen 7 bis 9 und 11, Stellen | `SchreibeEntity`; `SchreibeCollection` mit 3 Elementen: erstes, mittleres, letztes = 4 Stellen | **Stelle:** das mittlere Element der Collection nicht geprüft |
| Zeile 10, Art gegen Typ | **Typ** 8, **Art** 8 (+ `null`, `redigiert`, leer, unbekannt), **Nullbarkeit** 2 | **Typ:** `Edm.Date` und `Edm.DateTimeOffset` vertauscht; **Nullbarkeit:** `null` bei nullbarer Property als Text `null` statt als Literal |
| Zeilen 1 bis 4, Zustände | **Operation** 4; **ctx** 5 (nil, aktiv, abgebrochen, Frist abgelaufen, abgebrochen mit Ursache); **dst** 4 (nil, gut, Fehler, kurz) | **ctx:** `DeadlineExceeded` als `Canceled` gemeldet; **dst:** der erste Aufruf von `Write` mit Fehler wiederholt |
| Zeile 27, Escapes | **Zeichen** 41; **Stelle** 5 | **Zeichen:** U+001F nicht escaped (Grenze 0x1F gegen 0x20); **Stelle:** der Name der Property nicht escaped |

**Ergebnis des Implementers: Zusage · Mutation · roter Test** (`AGENTS.md` §3.10, §3.15, §3.17; 2026-11-06). Messender ist `tools/harness/mutationslauf.sh` (`SPEC-210`; Image `odata-recorder:race`, `--memory 2g`, `--cpus 4`, ein Container je Liste, unter `flock`); die Listen (je Mutation Datei, Text vor und nach, Test, Klasse) liegen im Scratchpad des Implementers und nicht im Repo. **182 Läufe der Klasse `andere`** (181 verschiedene Mutationen, eine in zwei Listen): alle `rot`, keiner `rot-paket`, `gruen` oder Übersetzungsfehler im letzten Lauf; der zuständige Test ist der rote (kein Rot aus dem Aufbau: die ersten Läufe nannten 4 Übersetzungsfehler (Fehler der Mutation, nicht des Codes), ein `rot-paket` und ein `gruen`, siehe *Lücken, die der Lauf fand*). Die Namen sind die der Liste (`zNN-…` mit der Zeile der Tabelle).

| Zeile | Mutationen (Name ohne `zNN-`) | roter Test |
|---|---|---|
| 1 | `dst-nil`, `ctx-nil` | `TestSchreibeNilArgumente` |
| 2 | `ctx-vor-args-entity`, `-collection`, `-anzahl` | `TestArgumentVorKontext` |
| 3 | `cause` (`context.Cause` statt `ctx.Err()`), `verpackt` | `TestKontextBeendet` |
| 4 | `err-zweimal`, `err-nach-write`, `done`, `deadline` | `TestKontextEinmalVorWrite` |
| 5 | `version-nur-2`, `version-entfaellt`, `adresse-ohne-modell` | `TestModellVersion`, `TestSetUngueltig` |
| 6 | `set-schreibweise`, `set-ohne-typ`, `set-leer-im-modell` | `TestSetUngueltig` |
| 7 | `ring-ohne-zaehler`, `abstract`, `nur-erstes-element`, `mittleres-element-uebersprungen`, `odata-type-immer` | `TestEntityTyp` |
| 8 | `nul-nur-anfang`, `laengengrenze`, `utf8-nicht-geprueft` | `TestSchluessel` |
| 9 | `duplikat-nur-gleicher-wert`, `duplikat-nur-benachbart`, `name-schreibweise` | `TestPropertyNamen` |
| 10 | `date-dto-vertauscht`, `null-bei-nicht-nullbar`, `null-mit-text`, `redigiert-als-null`, `typ-nicht-verglichen` | `TestArtGegenTyp` |
| 11 | 24 Mutationen `z11-…`: `minus-null-int32`, `int-bereich`, `int-rechts-nicht-geprueft`, `fuehrende-null`, `dezimal-grenze-gleich`, `dezimal-grenze-plus-eins`, `dezimal-null-am-ende`, `dezimal-punkt-ohne-bruch`, `dezimal-exponent`, `dezimal-minus-null`, `boolean-gross`, `guid-grossbuchstaben`, `guid-nur-anfang`, `guid-laenge-37`, `guid-bindestrich-verschoben`, `guid-ausgabe-gross`, `datum-jahrhundert`, `dto-stunde-24`, `dto-minute-60`, `dto-sekunde-60`, `dto-bruchteil-13`, `dto-bruchteil-null`, `dto-klein-t`, `dto-nur-erste-property` | `TestTextKanonisch` |
| 12 | `auswahl-nur-erstes-element`, `auswahl-bei-none-entity`, `auswahl-bei-none-collection` | `TestAuswahl` |
| 13 | `anzahl-groesser-null`, `wert-bei-nicht-vorhanden`, `n-groesser-null` | `TestAnzahlGrenzen` |
| 14 | `nur-code`, `message-nicht-verglichen`, `target-nicht-verglichen` | `TestFehlerobjektTabelle` |
| 15 | `teilausgabe` | `TestNichtsBeiVerstoss` |
| 16 | `bloecke`, `writestring`, `close`, `readfrom` | `TestEinWrite` |
| 17 | `kurz-ignoriert`, `kurz-nur-null`, `fehler-verpackt` | `TestDstFehler` |
| 18 | `charset`, `ieee-immer`, `version-401`, `anzahl-contenttype`, `fehler-contenttype` | `TestDarstellungskopf` |
| 19 | 13 Mutationen `z19-…` (`tilde-kodiert`, `plus-roh`, `prozent-roh`, `komma-kodiert`, `apostroph-kodiert`, `gleich-kodiert`, `hex-klein-niedrig`, `hex-klein-hoch`, `nibble-gleich`, `007f-roh`, `folgebytes-roh`, `set-roh`, `schluessel-roh`) | `TestAdresseKodierung` |
| 20 | `kontext-kodiert`, `id-eigener-kodierer` | `TestAdresseGleich` |
| 21 | 18 Mutationen `z21-…` (`ftp`, `schema-gross`, `laenge-2049`, `laenge-2047`, `leeres-segment`, `dekodiert-nicht-geprueft`, `nur-ein-punkt`, `userinfo`, `fragezeichen`, `host-umlaut`, `port-ohne-ziffern`, `port-sechs`, `ipv6-g`, `ipv6-leer`, `nul`, `nur-erstes-segment`, `adresse-prueft-nicht`, `collection-ohne-entities`) | `TestWurzel` |
| 22 | 18 Mutationen `z22-…` (`letzter-bereich`, `application-stern`, `medientyp-schreibweise`, `tab-nicht-entfernt`, `rechts-nicht-entfernt`, `links-nicht-entfernt`, `name-meta-schreibweise`, `name-ieee-schreibweise`, `wert-none-schreibweise`, `wert-true-gemischt`, `wert-eins`, `doppelt-meta-letztes`, `doppelt-ieee-letztes`, `name-prefix`, `wert-none-prefix`, `wert-true-suffix`, `nur-erste-kopfzeile`, `letzte-kopfzeile-zuerst`) | `TestAushandlung` |
| 23 | `int32-string`, `float64`, `count-immer-zahl`, `decimal-immer-zahl`, `negative-nicht-quoted` | `TestZahlform` |
| 24 | `leerzeichen`, `entity-suffix`, `none-kontext`, `none-id`, `odata-type-immer` | `TestEntityBytes` |
| 25 | `value-null`, `count-bei-nicht-vorhanden`, `ohne-komma`, `none-ohne-count`, `none-kontext` | `TestCollectionBytes` |
| 26 | `uebergabe-reihenfolge` (Klasse `andere`); `map-iteration` (Klasse nebenläufig, Rate unten) | `TestPropertyReihenfolge` |
| 27 | 13 Mutationen `z27-…` (`kurzform-n`, `slash`, `007f`, `hex-gross`, `grenze-1f`, `grenze-20`, `2028`, `name-nicht-escaped`, `typ-nicht-escaped`, `kontext-nicht-escaped`, `anfuehrungszeichen`, `backslash`, `utf8-ungeprueft`), dazu 4 `z27u-…` (`name-roh`, `typ-roh`, `kontext-roh-entity`, `kontext-roh-collection`) | `TestEscapes`, `TestUngueltigesUTF8AnVierStellen` |
| 28 | `uebersprungen-als-null`, `fehlende-als-null` | `TestNichtGenannt` |
| 29 | `message-name`, `target-fehlt` | `TestFehlerBytes` |
| 30 | `zeilenende` | `TestAnzahlBytes` |
| 31 | `int64-float64` (und `z11-guid-ausgabe-gross`) | `TestRundlauf` (`TestTextKanonisch`) |
| 32 | `auswahl-komma-leerzeichen`, `auswahl-ohne-klammer`, `collection-entity-suffix` | `TestRundlaufAuswahl` |
| 33 | `none-kontext`, `none-count` | `TestStufeNone` |
| 34 | `cache` (Darstellung nach `Content-Type` gemerkt) | `TestWiederholung` |
| 35 | `auswahl-sortiert`, `properties-sortiert`, `accept-geschrieben`, `uebersprungen-geschrieben`, `entities-geschrieben` | `TestArgumenteUnveraendert` |
| 36, 37 | `z36-puffer-geteilt` und 90 Zähler (Rate unten) | `TestGleichzeitig` |
| 38 | `import-v2` (ein Import von `jsoncodecv2`): `rot` | `TestKeinPaketImportiertEineAndereVersion` |
| 39 | benannte Ausgabe: `git diff --stat 17528be..HEAD -- src/bootstrap src/adapters/driving` ist leer (geprüft am letzten Commit des Implementers) | – |
| 40 bis 42 | akzeptierte Negative, keine Mutation | – |

Zeile 38, erster Halbsatz (eine Methode entfernt): das ist ein Übersetzungsfehler von `var _ outbound.ODataSchreiber = (*Schreiber)(nil)` im Testpaket und kein Rot des Werkzeugs; das Werkzeug zählt ihn nirgends als rot (`SPEC-210` Nr. 5 d), der Beleg ist die Übersetzung selbst.

**Nebenläufige Zeilen (36, 37) und Zeile 26: Rate mit Zahl, Last und Messendem.** Messender wie oben; die Prozesse eines Arms laufen höchstens vier zugleich, jeder mit `-test.count=1`; der Detektor ist `-race` (cgo, Image `odata-recorder:race`). Last des Hosts (`/proc/loadavg`, erster Wert) zu Beginn der Mutationen 3,75 bis 23,5 bei 4 CPUs des Containers (andere Rollen liefen zugleich).

| Zusage · Mutation | Klasse | Prozesse (Messender) | mit Detektor | ohne Detektor |
|---|---|---|---|---|
| 36: ein Puffer als Variable des Pakets, den alle Aufrufe überschreiben (`z36-puffer-geteilt`) | nebenläufig | 20 mit, 50 ohne | 20 von 20 rot | **50 von 50 rot** (der Wirkungs-Fänger zählt abweichende Ausgaben, `TestGleichzeitig`) |
| 37: ein Zähler als Variable des Pakets, in **jeder Funktion** (47 Funktionen des Schreibers einschließlich der Einzeiler und der, die den Schreiber liefert) und **in jedem Zweig, den ein Lauf ausführt** (43 Zweige: jede Zeile `ungueltig(…)`, `ctx.Err()`, `Write`-Fehler, kurzer `Write`, UTF-8, Port, IPv6, Prozent, Typ im Cache, Stern der Auswahl, jeder `case` von `wert`, jeder `case` der Aushandlung); 90 Mutationen (44 + 3 `zf-…` und 43 `zz-…`) | nebenläufig | 20 je Mutation | **20 von 20 rot, in 89 von 90 Mutationen**; die 90. (`zz-utf8`, der Zweig des Escape für ungültiges UTF-8) war im ersten Lauf `gruen` (0 von 20): kein Lauf erreichte den Zweig, denn die Prüfung des Wertes kommt vor dem Escape. Der Fall steht jetzt in `TestGleichzeitig` (Namen des Modells mit ungültigem UTF-8 an vier Stellen); Wiederholung auf dem letzten Stand: 20 von 20 | – (der Zähler ändert keine Ausgabe, der Arm entfällt) |
| 26: Iteration über eine `map` statt der Reihenfolge des Modells (`z26-map-iteration`) | nebenläufig | 50 ohne | – (kein Arm, kein Rennen) | **50 von 50 rot** je Prozess (`TestPropertyReihenfolge`, 24 Anordnungen × 5 Aufrufe × 2 Typen je Prozess) |

Auf dem letzten Stand des Tests liefen nach der Ergänzung nur 13 der 92 Mutationen noch einmal (`z36-puffer-geteilt`, `z26-map-iteration`, `zz-utf8`, sieben Zähler an `ausgeben`, `text`, `entity`, `kanonDezimal`, `aushandeln`, einer Zeile `ungueltig(Wurzel)`, einer Zeile `ungueltig(Code)`, dazu drei Zähler in `NewSchreiber`, `kanonUhrzeit` und `kanonBruchteil`, die erst die Teilung von `kanonZeitpunkt` nach dem Lint hinzufügte): alle rot, jede 20 von 20 beziehungsweise 50 von 50. Die übrigen 79 hat der Test davor gemessen (er rief weniger Aufrufe); ein Rot kann durch zusätzliche Aufrufe nicht grün werden, die Zahlen sind für sie der Stand vor der Ergänzung. Die Zähler brauchen eine Variable im Paket, die der Schreiber nicht hat: der Lauf legte dafür eine Testdatei export_test.go mit einer Variablen `Zaehler` an und entfernte sie danach; sie ist nicht im Repo.

**Pflicht-Mengen** (zweite Tabelle, `AGENTS.md` §3.17): jede Allaussage ist eine Schleife über alle Elemente der Klasse mit Zählprüfung im Test (`t.Errorf` bei falscher Zahl) und je Achse mindestens einer Mutation, die der Plan oben nicht nannte.

| Menge | Zählprüfung (Test) | Achse · Mutation, im Plan nicht genannt (alle `rot`) |
|---|---|---|
| Zeile 11, Text je Art (33 gültig, 107 ungültig, 560 Zellen) | `TestTextKanonisch` zählt 33, 107, 560 | Zeichen `boolean-gross`; Länge `dezimal-grenze-plus-eins`; Folge `guid-bindestrich-verschoben`; Seite `int-rechts-nicht-geprueft`; Stelle `dto-nur-erste-property`; Zustand `null-mit-text` |
| Zeile 21, `Wurzel` (11 gültig, 44 ungültig, 165 Zellen) | `TestWurzel` zählt 11, 44, 165 | Zeichen `fragezeichen`; Länge `laenge-2049`; Folge `nur-ein-punkt`; Seite `schema-gross`; Stelle `nur-erstes-segment`; Zustand `collection-ohne-entities` |
| Zeile 19, 130 Zeichen der Adresse (260 Zellen) | `TestAdresseKodierung` zählt 130 und 260 | Zeichen `apostroph-kodiert`; Länge `folgebytes-roh`; Folge `nibble-gleich`; Seite `hex-klein-hoch`; Stelle `set-roh` |
| Zeile 22, Aushandlung (88 Zellen) | `TestAushandlung` zählt 88 | Zeichen `wert-true-gemischt`; Länge `name-prefix`; Folge `letzte-kopfzeile-zuerst`; Seite `rechts-nicht-entfernt`; Stelle `nur-erste-kopfzeile`; Zustand `doppelt-meta-letztes` |
| Zeile 27, Escapes (58 Zellen) | `TestEscapes` zählt 58 | Zeichen `grenze-20`; Folge `kurzform-n`; Seite `anfuehrungszeichen`; Stelle `typ-roh` (`TestUngueltigesUTF8AnVierStellen`); Zustand `utf8-ungeprueft` |
| Zeilen 1 bis 4, Zustände | `TestSchreibeNilArgumente` 8, `TestArgumentVorKontext` 8, `TestKontextBeendet` 12, `TestKontextEinmalVorWrite` 8 | ctx: `cause`, `deadline`; Operation: `ctx-vor-args-anzahl`; dst: `kurz-nur-null` |
| Zeilen 7 bis 9 und 11, Stellen | `TestEntityTyp` 44, `TestSchluessel` 32, `TestPropertyNamen` 48 Zellen | Stelle: `mittleres-element-uebersprungen`, `nur-erstes-element` |
| Zeile 10, Art gegen Typ | `TestArtGegenTyp` zählt 16 gültig, 96 ungültig | Typ: `date-dto-vertauscht`; Nullbarkeit: `null-bei-nicht-nullbar`; Zustand: `redigiert-als-null` |

**Lücken, die der Lauf fand** (alle vor der Übergabe geschlossen): (1) `z27-grenze-20` (U+0020 als Escape) blieb in `TestEscapes` grün — das Leerzeichen als Grenze fehlte in der Klasse; es steht jetzt dort (58 statt 57 Zellen). (2) `z27-utf8-ungeprueft` blieb grün — ungültiges UTF-8 in einem Namen des Modells erreicht den Escape nur über Property, Set, Typ und Auswahl; `TestUngueltigesUTF8AnVierStellen` führt vier Formen an vier Stellen und beiden Operationen (32 Zellen). (3) `zz-utf8` (oben). (4) `z06-set-leer-im-modell`: die Prüfung `Set` nicht leer war gegenüber der Suche im Modell nicht unterscheidbar, solange das Modell kein Set mit leerem Namen hat; `TestSetUngueltig` führt ein Modell mit einem Set `""` (21 statt 18 Zellen). (5) `TestZahlform` führte nur positive Werte; `negative-nicht-quoted` fand die Lücke (jetzt 26 statt 14 Zellen, drei negative Werte). Dazu zwei Elemente mehr in der Klasse von Zeile 22 (`True`, zwei Kopfzeilen mit `application/json`, Namen und Werte mit Zusatz). **Gleichwertig, nicht fangbar:** die Prüfung `Unterstuetzt()` in `entity` (`Edm.Binary` ist keine Art von `SPEC-169` Nr. 4, die Paarung von Art und Typ lehnt jede Property nicht unterstützten Typs ohnehin ab; die Prüfung bleibt, weil `SPEC-168` Nr. 3 c sie nennt).

Zusatztabelle der Paare und Modi: unter dem Raster in §6.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `welle-record-replay` liegt in `done/` (eingetreten, Beleg `welle-record-replay-results`; der Codec zum Lesen aus `slice-odata-codec-json` ist das Gegenstück des Round-trips); der Start als Teil 1 von `welle-mock` folgt dem Trigger aus deren §2 (Linie B). Die Spezifikation trägt den Vertrag samt Randformen (`SPEC-168`, `SPEC-169`), bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nach dem Werkzeug keine Rückführung durch eine Überschreitung der Marke; eine Überschreitung trägt einen Eintrag `ueberschreitung` mit Grund im Block (`SPEC-128` Nr. 5 d, Nr. 10). Die Schnittfrage ist beantwortet: kein Schnitt, der Grund steht im Block (§3, Entscheidung 11); schneidet der Planner dennoch vor dem ersten Code-Commit, entstehen zwei Pläne, und dieser behält den ersten.

*Zahlen des Architect und Metrik:* der Teil des Architect steht im Block unten; *Basis* (`17528be`, der letzte Commit des Architect, der Vertrag, Tabelle und Raster abschließt), *Schätzung* (4570) und *Marke* (6200) hat der Planner in eigenem Commit vor dem ersten Code-Commit gesetzt; eine ältere Marke gab es nicht, also kein `ersetzt`. Die Schnittfrage hat der Planner selbst geprüft und bestätigt: kein Schnitt. Die Schätzung liegt zwischen `slice-odata-url-filter` (3460, Marke 4700) und den Lesern `csdl` (5036, Ende 4333) und `json` (5430, Ende 7236); `slice-edm-ausdruecke` (3645, Messung 3229) lag unter seiner Schätzung. Der Schreiber ist kein Leser fremder Bytes, die Aushandlung (Zeilen 22 und 23, zwei Zeilen) und die Adresse (19 bis 21) hätten als Teil weit unter 2223 und teilten `Darstellung` und den Aufbau der Ausgabe. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..HEAD` über die Pfade des Blocks, Tests eingeschlossen (`SPEC-128` Nr. 4 a). Die Zahlen des Architect (§3, Entscheidung 11): 42 Zeilen der Tabelle (2 nebenläufig, 40 andere), ein Fake, 1000 Zeilen Produktcode, *S* = 4570, Marke 6200. Die Pfade nennen Dateien statt Verzeichnissen, wo `main` und andere Zweige dasselbe Verzeichnis ändern (`src/hexagon/model/`, `src/hexagon/ports/outbound/`); `src/adapters/driven/odata/jsoncodec/` ist ein Verzeichnis, weil dieser Slice dort nur neue Dateien anlegt und kein anderer Zweig das Paket ändert.

```groessenmarke
pfad: src/hexagon/model/schreiben.go
pfad: src/hexagon/model/schreiben_test.go
pfad: src/hexagon/ports/outbound/odata_schreiber.go
pfad: src/hexagon/ports/outbound/codec_satz.go
pfad: src/adapters/driven/odata/jsoncodec/
produktcode: 1000
fakes: 1
basis: 17528be
schaetzung: 4570
marke: 6200
schnitt-grund: Kein Teil ist ohne geteilten Mechanismus abtrennbar. Die Aushandlung von Stufe und Zahlform (Zeilen 22 und 23), Adresse und Wurzel (19 bis 21), die Werte (10 und 11) und die Antworten mit Kopf (14, 16 bis 18, 24 bis 30) teilen den Typ Darstellung, die Prüfung der Argumente (Zeilen 1 bis 9, 12, 13) und den Aufbau der Ausgabe (Zeilen 24 bis 27): die Aushandlung ändert Kopf und Zahlform jeder Ausgabe, die Adresse steht in jeder Entity. Ein Teil ohne diese Mechanismen hätte unter 2223 nur wenige Zeilen. Die Linie B führt ihre Slices nacheinander, ein Schnitt gewänne keine Parallelität und verdoppelte Review und Verifikation des Kerns; derselbe Grund schließt den Handoff-Schnitt (AGENTS.md 3.14) aus.
```

**Messung des Implementers beim Handoff** (2026-11-06, `make groessenmarke` am Commit vor diesem Eintrag): **3000 von Marke 6200** (Schätzung 4570); kein Eintrag `ueberschreitung`. Aufteilung: Produktcode rund 810 Zeilen (Schätzung 1000: `schreiber.go` 124, `schreiber_bau.go` 243, `schreiber_kanon.go` 155, `schreiber_adresse.go` 132, `schreiber_accept.go` 55, `schreiber_ausgabe.go` 46, `schreiben.go` 31, `odata_schreiber.go` 22, `codec_satz.go` 2), Tests rund 2190 Zeilen in sieben Dateien. Die Messung liegt 34 % unter der Schätzung (der Faktor 1,25 der Marke deckt den Stand nach dem Review, nicht den Handoff); die Formel zählte 85 Zeilen je nebenläufige Zeile und 40 je andere, tatsächlich trugen die Fänger als Schleifen über Mengen weniger Zeilen je Tabellenzeile als geschätzt (rund 52 Zeilen Test je Zeile).

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7): **geprüft, nicht berührt.** Der Slice ändert laut §3 keine Datei der Gate-Kette (`Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/`, Stempel); das Paket `jsoncodec` und die Ports existieren, `.a-check.yml` braucht keine Regel. Muss der Implementer doch eine ändern, hält er an und gibt es als Rückgabe an den Planner zurück (§6, Ausgang *offen*, `AGENTS.md` §3.13); der Slice ändert sie nicht selbst.

**Aufträge an den Architect:** erledigt (2026-11-06): (1) Vertrag → `SPEC-168`, `SPEC-169`; (2) Nummer in `SPEC-172` Nr. 6 g und Einleitung (und `SPEC-175` Nr. 3, der Name) nachgezogen; (3) Tabelle *Zusage · Fänger* in §3; (4) Raster in §6, Aushandlung und Zahlform entschieden; (5) erwartete Größe: *p* = 1000, *S* = 4570, Marke 6200; (6) Schnittfrage gestellt und beantwortet (§3, Entscheidung 11).

- `in-progress` → `open` (blockiert — Carveout?): eine Randform des Vertrags hat keinen Ort in der Spezifikation (Zelle im Raster in §6 ohne Nummer); die Aushandlung der Stufe ist entschieden (`SPEC-169` Nr. 3).

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

- Aushandlung des JSON-Metadatenlevels (`Accept`, `odata.metadata`, `$format`): `LH-FA-07.b` sagt „folgen dem ausgehandelten Metadatenlevel“, ohne die Aushandlung zu regeln — **Ausgang:** entfallen als Lücke des Plans: `SPEC-169` Nr. 3 entscheidet (Stufen `minimal` und `none`, `full` als `minimal`, kein `406`, `$format` ohne Wirkung); Fänger Zeilen 22, 24, 25, 33.
- Darstellung von `Edm.Int64` und `Edm.Decimal` im JSON (Zahl oder Zeichenkette) und die Gegenprobe über den Leser — **Ausgang:** entfallen als Lücke des Plans: `SPEC-169` Nr. 3 und 4 entscheiden (Zeichenkette genau bei `IEEE754Compatible=true`, sonst Zahl); Fänger Zeilen 23 und 31 (Round-trip über alle acht Typen in beiden Formen).
- Zeichen jenseits von ASCII und Escape-Folgen in den Fixtures — **Ausgang:** entfallen: jeder Wert mit Sonderbytes entsteht in den Tests aus Bytewerten (`"\xc3\xa9"`, `"\xff"`, `esc("001f")` aus `bs` und Hex-Ziffern, Raw-Strings); gegengelesen sind alle `schreiber_*_test.go` des Pakets `jsoncodec` mit `LC_ALL=C grep -nP '\\u[0-9a-fA-F]{4}'` (kein Treffer) und `LC_ALL=C grep -nP '[\x80-\xff]'` (Treffer nur in Meldungen und Namen der Fälle, nie in einem Wert, der geprüft wird).
- Die Quelle von `Wurzel` im Mock: der Schreiber prüft nur die Form (`SPEC-169` Nr. 2); dass `Wurzel` aus `PUBLIC_BASE_URL` und dem Service-Root entsteht und nie aus dem `Host`-Header, sichert der Aufrufer (`SPEC-168` Nr. 10) — **Ausgang:** offen; Adresse: Planner (Zuweisung an die Dienste von Teil 2 der Welle; kein Fänger in diesem Slice).
- Sendungen an die Dienste von Teil 2 (`AGENTS.md` §3.13): (a) der Mock trägt `Codecs *outbound.CodecSatz`, ruft `URL` und `Schreiber`, setzt `Content-Type`, `Kennzeichen` und `Location` aus der `Darstellung` bzw. `Adresse` (`SPEC-168` Nr. 7, 8, 10); (b) die Verdrahtung `CodecSatz{Schreiber: jsoncodec.NewSchreiber()}` im Composition Root; (c) der V2-Schreiber erfüllt `SPEC-168` (Nr. 1 bis 9) mit dem Format von `SPEC-172` Nr. 6. Der Planner hat §1 *Ausdrücklich NICHT* und DoD der Nehmer in `open/` auf `main` gelesen (2026-11-06): kein Ausschluss trifft eine der Sendungen, alle Nehmer liegen in `open/`. Zeile (a) teilt sich: `Codecs`, `URL`, `Schreiber`, `Content-Type` und `Kennzeichen` gehören `slice-mock-lesen` (dessen „Übernimmt“ trägt `Codecs *outbound.CodecSatz` schon, nennt aber weder `SPEC-168` Nr. 10 noch die `Darstellung`), `Location` aus `Adresse` gehört `slice-mock-schreiben` (Ziel und DoD nennen `201` mit `Location`; es hat kein „Übernimmt“ zu `Codecs` und `Adresse`). Zeile (b) trägt `slice-bootstrap-m3` („Übernimmt“: der Root baut den Satz der Version 4 mit dem Schreiber der Version 4), ohne den Namen `jsoncodec.NewSchreiber()` und `SPEC-168`. Zeile (c): `slice-v2-payload-schreiben` hängt in seiner Code-Abhängigkeit an diesem Slice und trägt die Nummer des Vertrags erst nach dem Eintrag des Architect (sein Raster). Die Nachträge schreibt der Planner auf `main` im Commit des Starts des jeweiligen Nehmers (`welle-mock.md` §4, die Pläne in `open/` bleiben bis dahin unverändert) — **Ausgang:** offen; Adresse: Planner, Nachträge: `slice-mock-lesen` „Übernimmt“ um `SPEC-168` Nr. 7 und 10 (`Darstellung`, Satz) ergänzen; `slice-mock-schreiben` neues „Übernimmt“ für `Adresse` und `Location` (`SPEC-168` Nr. 8) samt DoD-Zeile; `slice-bootstrap-m3` „Übernimmt“ um `jsoncodec.NewSchreiber()` (`SPEC-168`) ergänzen.
- Pfade des Blocks, die Linie A oder andere Zweige ebenfalls berühren (`src/hexagon/ports/outbound/codec_satz.go`): `git merge main` brächte deren Änderungen in die Messung (`SPEC-128` Nr. 10 g) — **Ausgang:** entfallen: der Merge von `main` vor dem ersten Code-Commit (`4f3043e`) brachte keine Zeile in die Pfade des Blocks; `git diff --numstat 17528be..HEAD` über die Pfade des Blocks zählt allein Zeilen dieses Slice (`codec_satz.go`: 2 Einfügungen, das Feld `Schreiber`).
- Zeile 33 und `SPEC-169` Nr. 8: „der Leser lehnt den Body der Stufe `none` mit `kontext`, Ort `/@odata.context` ab“ gilt für eine Entity, deren Properties alle zum Typ des Sets gehören. Hat die Entity den abgeleiteten Typ und eine Property nur dieses Typs, schreibt die Stufe `none` kein `@odata.type`, und der Leser meldet die Property als `unbekannte_property` (Ort `/Extra`), weil er die Mitglieder in der Reihenfolge des Dokuments prüft und den fehlenden Kontext erst am Ende — die Ablehnung bleibt, nur der Grund ist ein anderer. Der Test `TestStufeNone` führt den abgeleiteten Typ deshalb mit Properties des Typs des Sets; der Code entscheidet nichts — **Ausgang:** offen; Adresse: Architect (Zeile 33 und `SPEC-169` Nr. 8 um „oder `unbekannte_property`, wenn die Entity eine Property nur des abgeleiteten Typs trägt“ präzisieren; die Stufe `none` ist kein Gegenstand des Rundlaufs).
- Gate-Kette (Bedingung aus §4) — **Ausgang:** entfallen: keine Datei der Gate-Kette ist zu ändern (geprüft, §3 Entscheidung 9).
- Handbuch-Pflicht und Katalog des Binary — **Ausgang:** entfallen: kein neuer Fehlercode, Schlüssel oder Kommando, das Binary ändert sich nicht (§3 Entscheidung 8).
- Speicher des Aufrufs wächst mit der Ausgabe (genau ein `Write`) — **Ausgang:** entfallen als Lücke: akzeptiert mit Grund und Folge in `SPEC-168` Nr. 6.

**Raster *Operation × Argumentart*** (Orte des Architect; `Nr.` mit `SPEC-168` oder `SPEC-169` davor; „gilt nicht“ mit Grund):

| Operation | fehlender oder leerer Wert | unbekannter oder fremder Wert | ungültige Form (Zeichen, Schreibweise) | Grenze (Länge, Bereich) | Abbruch und Fehler des Ziels | Wiederholung, Nebenläufigkeit |
|---|---|---|---|---|---|---|
| `SchreibeEntity` | `ctx`, `dst` nil: `SPEC-168` Nr. 1 a; `Typ` und `Schluessel` leer: Nr. 3 a, b; `Properties` `nil` oder leer zulässig: Nr. 3; `Auswahl`, `Accept` `nil` zulässig: Nr. 5 a, c; `Wurzel` leer: Nr. 5 d | `Set`, Version: Nr. 2; `Typ`: Nr. 3 a; Name einer Property: Nr. 3 c; `Art` gegen Typ: Nr. 4; Element der `Auswahl`: Nr. 5 a | `Wurzel`: `SPEC-169` Nr. 2; `Text` nicht kanonisch: Nr. 4; ungültiges UTF-8: Nr. 1; U+0000 im `Schluessel`: `SPEC-168` Nr. 3 b | Bereich und Länge der Zahlen, Jahr: `SPEC-169` Nr. 4; `Wurzel` 2048 Bytes: Nr. 2; `Schluessel` und Ausgabe: keine Grenze (`SPEC-168` Nr. 3 b, 6) | Kontext: `SPEC-168` Nr. 1 c; `dst`: Nr. 6 | `SPEC-168` Nr. 9 |
| `SchreibeCollection` | wie oben; `Entities` `nil` oder leer: `SPEC-169` Nr. 6 | wie oben, je Element | wie oben, je Element | wie oben; `Anzahl`: `SPEC-168` Nr. 5 b | wie oben | `SPEC-168` Nr. 9 |
| `SchreibeAnzahl` | `ctx`, `dst` nil: `SPEC-168` Nr. 1 a; leer: gilt nicht, `n` ist eine Zahl (0 ist gültig: `SPEC-169` Nr. 7) | gilt nicht: kein Name, kein Modell | negativ: `SPEC-168` Nr. 5 e | `MaxInt64` gültig: `SPEC-169` Nr. 7 | wie oben | `SPEC-168` Nr. 9 |
| `SchreibeFehler` | leeres Objekt: `SPEC-168` Nr. 5 f | unbekannter Code: Nr. 5 f | abweichende `message` oder `target`: Nr. 5 f | gilt nicht: feste Texte der Tabelle (`SPEC-109` Nr. 2) | wie oben | `SPEC-168` Nr. 9 |
| `Adresse` | `Wurzel`, `Set`, `schluessel` leer: `SPEC-168` Nr. 5 d, 2, 3 b; `ctx`, `dst`: gilt nicht, eine reine Funktion (Nr. 8) | `Set`, Version: Nr. 2 | `Wurzel`: `SPEC-169` Nr. 2; `schluessel`: `SPEC-168` Nr. 3 b; Kodierung: `SPEC-169` Nr. 2 | `Wurzel` 2048 Bytes: `SPEC-169` Nr. 2; `schluessel`: keine Grenze (`SPEC-168` Nr. 3 b) | gilt nicht: kein `ctx`, kein `dst` (Nr. 8) | `SPEC-168` Nr. 9 |

**Zusatztabelle** (Paare von Verstößen am selben Aufruf und Modi; Spalte *Fänger*, Zeile der Tabelle in §3):

| Paar oder Modus | Erwartung | Ort | Fänger |
|---|---|---|---|
| ungültiges Argument und beendeter Kontext | `ErrUngueltig` | `SPEC-168` Nr. 1 | Zeile 2 |
| ungültiges Argument und Fehler von `dst` | `ErrUngueltig`, `Write` 0-mal | `SPEC-168` Nr. 1, 6 | Zeile 17 |
| beendeter Kontext und Fehler von `dst` | Fehler des Kontexts, `Write` 0-mal | `SPEC-168` Nr. 1 c, 6 | Zeile 17 |
| zwei Verstöße in verschiedenen Argumenten (Set und Typ; Wert und Name) | `ErrUngueltig`, welcher gemeldet wird, ist nicht zugesagt; `dst` leer | `SPEC-168` Nr. 1 b | Zeile 15 |
| Verstoß der `Auswahl` in der Stufe `none` | `ErrUngueltig` | `SPEC-168` Nr. 5 a | Zeile 12 |
| Stufe × Zahlform × {Entity, Collection}: `minimal`/`none` × `Zahl`/`Zeichenkette` | je Zelle Bytes und Kopf nach `SPEC-169` Nr. 3, 5, 6, 7 | `SPEC-169` Nr. 3 | Zeilen 18, 23, 24, 25 |
| `full` verlangt | wie `minimal`, der `Content-Type` nennt `minimal` | `SPEC-169` Nr. 3 | Zeile 22 |
| Allaussage „kein Fall liefert eine Teilausgabe“ | über alle Fälle der Zeilen 1 bis 14 und 15 | `SPEC-168` Nr. 6 | Zeile 15 und der Ergebnis-Helfer, der in jedem Fehlerfall `Darstellung{}` und leeres `dst` prüft |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (4×, verkörpert in `AGENTS.md` §3.18) gilt für die Fixtures mit Zeichen jenseits von ASCII, `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (11×, verkörpert in `AGENTS.md` §3.9) für den Nachzug dieses Plans, `BEO-REPO/randform-erst-im-review-entschieden` (20×, verkörpert in `AGENTS.md` §3.16) für das Raster. Offen: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
