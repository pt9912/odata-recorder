# Slice slice-v2-query-filter: Leser von `$filter` des V2-Umfangs (Grammatik, Kompatibilitätsfehler im Ausdruck)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-QA-05`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md), [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-173` Nr. 3 und 5 (Literale, Grammatik von `$filter`) · `SPEC-174` (Kompatibilitätsfehler, Ursachen `filter` und `literal_typ`) · `SPEC-160` Nr. 8 (Einhängestelle `WithFilter`, in `slice-v2-query` geliefert) · `SPEC-164` (V4-Grammatik, die V2 mit Abweichung kopiert) · `SPEC-165` (Query-AST) · `SPEC-170` Nr. 3 · `LH-FA-17.a` (versionierter Adapter, keine Vermischung von V2 und V4) · `LH-FA-17` Negative (nicht unterstützte V2-Eigenschaft: expliziter Kompatibilitätsfehler) · `ARC-001`, `ARC-004`, `ARC-011` · `architecture.md` §7 (Erweiterungspunkt OData V2).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** — bis zur Priorisierung.

**Autor:** pt9912. **Datum:** 2026-11-01.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Leser von `$filter` für V2 (`urlparserv2.NewFilterLeser()`, `SPEC-173` Nr. 5) liest den dekodierten Wert der Option in den Query-AST des Domain-Modells (`SPEC-165`), mit den V2-Literalen aus `slice-v2-query` (`SPEC-173` Nr. 3), und hängt sich über `WithFilter` in den Parser ein; ein Operator, eine Funktion oder ein Pfad außerhalb des Umfangs ist der Kompatibilitätsfehler `filter`, ein Literal außerhalb `literal_typ` (`ODR-E6006`, `SPEC-174`), ein ungültiger Ausdruck `ODR-E6005`. Der Slice liest Text und kennt das Modell nicht; er schreibt keine Antwort.

**Übernimmt (Sendung von `slice-v2-query`, Architect-Ausgang zum Schnitt, `AGENTS.md` §3.13):** die Grammatik von `$filter` (`SPEC-173` Nr. 5) samt Lexer, Tiefe und Rangfolge der Operatoren; die Zuordnung der Fehler des Lesers (`SPEC-173` Nr. 1 nach der Berichtigung L7: `*URLFehler` der Klasse `ungueltig` und `*KompatFehler` mit `filter` oder `literal_typ` gelten mit ihrer Ursache, Ort `query/<n>`); die Kompatibilitätsfehler `filter` und `literal_typ` im Ausdruck; die Berichtigung L6 in `SPEC-173` Nr. 5 (`has`, `in`, `divby`, V4-Literale ohne Präfix, `INF` und `NaN`, Pfad im Ausdruck); die Literal-Funktion und den Lauf aus `slice-v2-query` (Code-Abhängigkeit, §4). Der Slice schließt `$filter` nicht aus; er ist die Adresse dafür.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Pfad, Schlüsselliterale, Systemoptionen, Kompatibilitätsfehlerwert `KompatFehler`, Code `ODR-E6006`, Abbildung in `outbound`, Feld `URL` in `CodecSatz` und die Einhängestelle `WithFilter` samt `FilterLeser` — `slice-v2-query`; dieser Slice nutzt sie und ändert keine von ihnen (Teile ohne geteilten Mechanismus mit der Grammatik).
- Ausführung des Ausdrucks und Typprüfung gegen das Modell — der Query-Interpreter aus `slice-query-interpreter` (`welle-mock`); der Parser kennt das Modell nicht (`SPEC-171` Nr. 5).
- Verdrahtung `CodecSatz{URL: urlparserv2.New(urlparserv2.WithFilter(urlparserv2.NewFilterLeser()))}`, Rundlauf und Handbuch — `slice-v2-bootstrap`, das diesen Slice als Voraussetzung nennt.
- Methode (`methode` erzeugt der Mock, `SPEC-174` Nr. 3) — `slice-v2-bootstrap`; der Parser hat keine Methode.
- Änderungen am V4-Parser `urlparser` — [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md): kein V2-Paket importiert ein V4-Paket; die Grammatik ist eine Kopie mit Abweichung (Produktcode rund 400 Zeilen).
- Vollständige Kompatibilität zu herstellerspezifischen (SAP-)Erweiterungen — Out-of-Scope von `LH-FA-17`.

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

- [ ] Leser: `urlparserv2.NewFilterLeser()` liest die Grammatik von `$filter` des V2-Umfangs (Vergleiche, Verknüpfungen, `not`, Klammern, die Literale von `SPEC-173` Nr. 3) in den Query-AST; ungültig `ODR-E6005`, ein Operator, eine Funktion oder ein Pfad außerhalb `filter`, ein Literal außerhalb `literal_typ` (`ODR-E6006`), nie eine scheinbar gültige Antwort; `New(WithFilter(NewFilterLeser()))` ordnet die Fehler des Lesers nach `SPEC-173` Nr. 1 (L7) zu.
- [ ] Randformen des Lesers fremder Bytes (Zeichen außerhalb der Grammatik, Minus-Läufe, Leerraum, Tiefe, Länge, Wiederholung, V4-Formen ohne Präfix): je Zelle des Rasters ein Ort in der Spezifikation und ein Fänger; Generator über die Klasse der Ausdrücke, Allaussage über Operatoren und Funktionen außerhalb des Umfangs.
- [ ] Nebenläufigkeit: gleichzeitige Aufrufe des Lesers und Wächter der Knotenzahl und von `TotalAlloc` (Zeilen der Tabelle in §3).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (die Berichtigung L6 in `SPEC-173` Nr. 5 steht vor dem ersten Code-Commit).
- [ ] Der Handbuch-Eintrag entfällt: der Slice bringt keinen Code und kein Kommando in den Katalog des Binary (`ODR-E6006` kommt mit `slice-v2-query`; das Binary ändert sich erst mit `slice-v2-bootstrap`).
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
| `src/adapters/driven/odata/urlparserv2/filter.go`, `filter_token.go` | neu | Leser `NewFilterLeser()` und Lexer (`SPEC-173` Nr. 5); Kopie von `urlparser` mit Abweichung (Operatorwörter, Literale mit Präfix, keine Zeit mit Zone), Produktcode rund 400 |
| `src/adapters/driven/odata/urlparserv2/` (Tests) | neu | Tabelle *Zusage · Fänger* unten, Zähl-Fake und Generator; eigene Test-Hilfen, keine geteilten (Black-Box-Tests anderer Pakete sind nicht importierbar) |
| `src/adapters/driven/odata/urlparserv2/` (Produktcode aus `slice-v2-query`) | update | nur, was die Grammatik an Literal-Funktion und Lauf braucht; die Signaturen aus `slice-v2-query` bleiben |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) füllt der Architect beim Start dieses Slice (Übersicht in 37 Zeilen unten); das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) steht vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Der Leser liest fremde Bytes (den Text des Ausdrucks): die Produktcode-Erwartung des Architect (400) ist ein fester Anteil der Schätzung (`SPEC-128`); Eingaben mit Escape-Folgen oder Rohbytes entstehen aus Bytewerten (`AGENTS.md` §3.18).
- Das Paket `urlparserv2` importiert kein V4-Paket ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)); `TestKeinV4Import` aus `slice-v2-query` deckt die neuen Dateien mit.

**Übersicht der Gruppen** (Architect, Maße von `slice-odata-url-filter`, auf V2 angepasst; keine Tabelle *Zusage · Fänger*, sie trägt deshalb andere Köpfe):

| Gruppe | Zeilen | Menge und Inhalt |
|---|---|---|
| Vergleiche, Verknüpfungen, `not`, Rangfolge, Assoziativität, Klammern | 5 | Menge 54, 6, 58, 18, 67 |
| Schlüsselwörter, Operator unbekannt, Operatorwort als Name, Namen | 4 | rund 196, 12, 19, 66 (die V2-Operatorwörter `add sub mul div mod` und `has in divby` sind Kompatibilität `filter`, L6) |
| Zeichen außerhalb der Grammatik, Kommentarformen, Minus-Läufe, Leerzeichen, anderer Leerraum | 5 | 195, 35, 24, 72 |
| Länge des Literals, Länge vor Form, Rest zählt nicht, Ende des Literals | 3 | 19, 27 |
| Literale gleich dem Schlüssel (Differenz zu `slice-v2-query`, Zeilen 28 bis 31) | 1 | rund 110 |
| V4-Formen ohne Präfix im Filter sind `wert` (Guid, Datum, Zeitpunkt, Uhrzeit), `INF` und `NaN` sind `literal_typ` | 1 | 24 |
| `null`, Wurzel ohne Typ | 2 | 37, 9 |
| Pfad ist `filter` (L1, L6), Aufruf (22 Funktionen aus `SPEC-170` Nr. 3, 5 V4-Namen, `geo.distance`, 5 erfundene × 5 Formen + 3), Operatoren und weitere Konstrukte | 3 | 10, 168, 49 |
| Tiefe, Rang im Filter, Fehlerwert, Einhängung mit echtem Leser, Dekodierung, Struktur, Wiederholung | 7 | 77, 49, 53, 13, 22, 24, 64 |
| nebenläufig (gleichzeitige Aufrufe, Wächter der Knotenzahl und `TotalAlloc`) | 2 + 1 | wie `slice-odata-url-filter` Zeilen 29 und 30; Wächter 18 |
| Allaussage (Operatoren 5, Funktionen 28 × 3 Einbettungen), Generator (2000 Läufe), echte V2-Filter (die 5 `filter`-Anfragen der Erfassung aus `TestEchteV2URLs` in `slice-v2-query`, `substringof('a',Name) eq true`, `Date gt datetime'2020-01-01T00:00:00'`, `Id eq guid'…'`, `Price gt 1.5M`, `Big eq 5L`, `not (Age gt 3)`, `Name eq null`, 5 weitere) | 3 | rund 100, 2000, 17 |

Summe 37 Zeilen, davon 2 nebenläufig.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect mit dem Vertrag in 37 Zeilen nach den Gruppen oben, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`open` → `next` → `in-progress`): `slice-v2-query` liegt in `done/` (Code-Abhängigkeit, unten); [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` (erfüllt). Die Berichtigung L6 steht in `SPEC-173` Nr. 5, und die Tabelle *Zusage · Fänger* füllt der Architect vor dem ersten Code-Commit (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-v2-query` (Literal-Funktion, Lauf, `KompatFehler`, `ODR-E6006`, `FilterLeser` und `WithFilter` liegen dort; B erzeugt ohne sie keinen Fehlerwert und hat keine Einhängestelle), `slice-odata-url-filter` und das Query-AST in `done/` (`SPEC-165`).
- **Was der Slice nicht braucht — und warum:** `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-v2-payload-schreiben` (der Parser kennt das Modell nicht), `welle-mock` ganz (der Test über beide Seiten liegt im Rundlauf von `slice-v2-bootstrap`).
- **Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`): vorläufig aus der Grobschätzung des Architect, mit der Tabelle neu zu rechnen: 37 Zeilen, davon 2 nebenläufig, mal 85 bzw. 80, keine Fakes, Produktcode 400 (`filter.go` 311 und `filter_token.go` 113, ohne die Guid-Form des Wortlaufs, mit der Zuordnung der Kompatibilitätsfehler): *S* = 85·2 + 80·35 + 400 = 3370, Marke *M* = 100 · ⌈135·3370 / 10000⌉ = **4600**. Der Planner setzt `basis`, `schaetzung` und `marke` beim Start (nach `slice-v2-query` in `done/`, vor dem ersten Code-Commit) und rechnet mit der gefüllten Tabelle neu. Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/adapters/driven/odata/urlparserv2/`, Tests eingeschlossen; die Messung beginnt bei der Basis nach `slice-v2-query` (derselbe Pfad). Erwartung der Messung (Messreihe Ende ÷ *S* = 0,69 bis 0,78): rund 2300 (2100 bis 2600). Der Block entsteht beim Start:

```text
Entwurf des Architect (kein Block des Gates, solange der Plan in open/ liegt):
pfad: src/adapters/driven/odata/urlparserv2/
produktcode: 400
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start nach `SPEC-128` setzt (vorläufig 4600; Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`, oder `slice-v2-query` verlässt `done/` nicht.

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

- Die Menge der Ausdrücke ist unbeschränkt — **Ausgang:** offen; Generator über die Klasse (2000 Läufe, Saat 1 bis 2000) mit Orakel in eigener Datei und Zählprüfung, Prüfgrenze als akzeptiertes Negativ mit Grund und Folge im Betrieb (`AGENTS.md` §3.17).
- Operatoren und Funktionen außerhalb des Umfangs sind eine Allaussage — **Ausgang:** offen; Schleife über die Listen der Spezifikation (5 Operatoren, 28 Funktionen, je 3 Einbettungen), nicht über die Fälle, die einfielen (`AGENTS.md` §3.15).
- Die Kopie der Grammatik von `urlparser` und die Literale aus `slice-v2-query` laufen auseinander — **Ausgang:** offen; akzeptiertes Negativ mit Fänger: Zeilen zu den Literalen gleich dem Schlüssel und der Rundlauf in `slice-v2-bootstrap`.
- Nebenläufigkeit (gleichzeitige Aufrufe des Lesers, Wächter): eine Zusage über Nebenläufigkeit gilt nur als gefangen, wenn die Mutation in jedem Lauf rot wird; Rate mit Zahl der Prozesse im Plan (`AGENTS.md` §3.10) — **Ausgang:** offen.
- Fakes: der Zähl-Fake des `FilterLeser` aus `slice-v2-query` bleibt dort; dieser Slice braucht keinen Port-Fake (Block `fakes: 0`) — **Ausgang:** offen bis zur Implementierung (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`).
- Die Berichtigung L6 in `SPEC-173` Nr. 5 muss vor dem ersten Code-Commit stehen (Architect-Ausgang zu `slice-v2-query`, Abschnitt 5) — **Ausgang:** offen; Adresse: Architect.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen aus dem Architect-Ausgang zu `slice-v2-query`; der Architect verfeinert sie beim Start mit der Tabelle):

| Operation | Ausdruck fehlt oder leer | unbekannter Operator, Funktion oder Pfad (außerhalb des Umfangs) | fremde Syntax (V4 am V2-Leser) | Literale in V2-Schreibweise, Typfehler | Wiederholung (doppeltes `$filter`), über Grenze (Länge, Tiefe) | Abbruch (Kontext) |
|---|---|---|---|---|---|---|
| Lesen des Ausdrucks (`NewFilterLeser`) | `$filter=` leer: `wert` (`SPEC-162` Nr. 8 mit `SPEC-173` Nr. 4) | `add sub mul div mod`, `has in divby`, Funktionen, Pfad mit `/`: `filter` (`SPEC-173` Nr. 5, L6, L1) | Literal ohne Präfix (Guid, Datum, Zeitpunkt, Uhrzeit): `wert`; `INF`, `NaN`: `literal_typ` (L6) | `SPEC-173` Nr. 3 und 5: `literal`, `literal_typ`, `zu_lang` | Tiefe und Länge: `SPEC-173` Nr. 5, `SPEC-160` Nr. 6; doppelt: `option_doppelt` (`SPEC-162` Nr. 3 a) | gilt nicht: der Leser hat keinen Kontext (`SPEC-160` Nr. 8) |
| Einhängung mit echtem Leser (`WithFilter`) | kein Leser: `filter` (`SPEC-160` Nr. 8, in `slice-v2-query`) | Leser liefert `*KompatFehler`: Ursache und Ort `query/<n>` gelten (L7) | Fehler anderer Art oder Klasse: Fehler des Aufrufers `model.ErrUngueltig` (L7) | `literal_typ` des Lesers gilt mit Ort (L7) | höchstens ein Aufruf je `ParseURL` (`SPEC-160` Nr. 8 a) | gilt nicht (wie oben) |
| Kompatibilitätsfehler im Ausdruck | gilt nicht: er hat immer eine Ursache (`SPEC-174` Nr. 2) | die Menge: `SPEC-170` Nr. 3 | `SPEC-173` Nr. 5 | `literal_typ` (`SPEC-174` Nr. 2) | Rangfolge `SPEC-174` Nr. 4 und `SPEC-173` Nr. 6 | gilt nicht (wie oben) |

**Zusatztabelle** (je Paar von Verstößen am selben Ausdruck und je Modus; `AGENTS.md` §3.16):

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| Kompatibilität und `400` im selben Ausdruck | der erste in der Lesefolge gilt (`SPEC-173` Nr. 1 und 6) | die Gruppe *Rang im Filter* der Tabelle (beim Start vom Architect) |
| Fehler des Lesers gegen eine spätere Option der Query | `SPEC-160` Nr. 5 mit `SPEC-173` Nr. 6 | Zeile *Einhängung mit echtem Leser* |
| mit Leser gegen ohne Leser (Modus der Einhängestelle) | `SPEC-160` Nr. 8 | Fänger in `slice-v2-query` (ohne Leser), hier mit Leser |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-01); der Planner zieht Treffer und Zähler beim Start nach. `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (Schwelle erreicht, Verkörperung offen) gilt für Fakes; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (Schwelle erreicht, Zustand offen) für die Betriebsregel „nur `Edit` und `Write`“. Der Slice ist nebenläufig in zwei Zeilen der Tabelle (§6).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
