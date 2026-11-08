# Slice slice-handbuch-katalog-schluessel-kommandos: Kataloge der Konfigurationsschlüssel und der Kommandos im Abgleich mit dem Handbuch

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-benutzerhandbuch](../welle-benutzerhandbuch.md).

**Bezug:** `LH-QA-05`, `LH-RB-01`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-201` (Vertrag unter `harness/vertraege/`, Abgleich der Kataloge) Nr. 4, 5, 6, 7 (Art 4, 8, 10), 8, 9 e und h, 10 a und die Fälle G16 bis G31 · `SPEC-200` (Handbuch gegen das gebaute Binary) · `SPEC-102` (Gegenproben) · `SPEC-110` Nr. 2 (Schlüssel dieser Version) · `SPEC-138` Nr. 2 (Kommandos des Adapters).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-handbuch-katalog-schluessel-kommandos`, Worktree `/Development/odata-recorder-wt-handbuch-katalog`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912 (Planner). **Datum:** 2026-11-06.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `make katalog-check` hält auch die Kataloge der Schlüssel und der Kommandos gegen das Handbuch (`SPEC-201` Nr. 4, 5): jeder Schlüssel, den die Startkonfiguration liest, und jedes Kommando des Binary steht im Handbuch (Abschnitte *Einstellungen* und *Kommandos*), und das Handbuch nennt keinen, den das Binary nicht kennt oder ablehnt.

**Übernimmt:** `slice-handbuch-meldungskatalog` (Rückgabe des Architect an den Planner, 2026-11-06: der Slice ist an der Grenze der Kataloge geschnitten; die Kataloge der Schlüssel und der Kommandos nach `SPEC-201` Nr. 4, 5, 6, 7 (Art 4, 8, 10), 8 und 10 a, die beiden Tabellen des Handbuchs, die Gegenproben zu ihren Befunden und der Nachzug von `SPEC-102` gehen hierher).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Programm `katalog`, Stufen `katalog-bin` und `katalog`, Ziel `make katalog-check`, Gate-Kette und der Katalog der Fehlercodes samt Ausnahmen-Datei — `slice-handbuch-meldungskatalog`; dieser Slice erweitert das Programm, er legt es nicht an (Start-Trigger, §4).
- Die Entscheidung der ADR-Frage zur Laufzeit ([ADR-0005](../../adr/0005-gegenproben-ausserhalb-der-gate-kette.md)) — der User, vorgelegt in `slice-handbuch-meldungskatalog` §6; dieser Slice misst die Zusatzzeit seiner Erweiterung und trägt sie in §7 nach.
- Bedeutung und Wahrheit der Zeilen der Tabellen (`SPEC-201` Nr. 9 a) — akzeptiertes Negativ des Vertrags; die belegten Beispiele (`slice-handbuch-record-replay`) und der Probelauf (`slice-handbuch-abnahme`) tragen es.
- Optionen des Binary, Namen hinter `Gesetzt (nichtleer) und nicht geliefert sind` in `SPEC-110` Nr. 2 und der Leser für `SERVICE_ODATA_VERSION` — der Vertrag führt keine Optionen (`SPEC-201` Nr. 5 b) und vergleicht die Menge N nicht (Nr. 9 c); den Leser liefert die Welle `welle-odata-v2` (`slice-v2-bootstrap`).
- Eine Änderung von `AGENTS.md` — nur als Diff-Vorschlag an den User (Pflicht der Folgewellen, `SPEC-201` Nr. 11), den schon `slice-handbuch-meldungskatalog` vorlegt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Kataloge im Programm `katalog`: Schlüssel (C aus `konfig*.go`, K aus `SPEC-110` Nr. 2, N der abgelehnten Namen, W aus der Tabelle *Einstellungen*) und Kommandos (C aus `cli.go`, W aus der Tabelle *Kommandos*) nach `SPEC-201` Nr. 4 bis 8 und 10 a; jede Zusage hat einen Fänger, vom Implementer rot gesehen (`LH-QA-05`, `LH-RB-01`).
- [ ] Handbuch: die Abschnitte *Einstellungen* (26 Schlüssel: die 27 von `SPEC-110` Nr. 2 ohne `SERVICE_ODATA_VERSION`, den kein Leser in `konfig*.go` liest; seine Zeile trägt der Slice mit dem Leser, `SPEC-201` Nr. 4 c, 11) und *Kommandos* (`start`, `version`) mit den Tabellen nach `SPEC-201`; die Regel „Jeder Schlüssel und jedes Kommando, das das Binary zeigt, kommt im Handbuch vor, und das Handbuch nennt nichts anderes“ ist durch Fänger über beide Richtungen und alle Elemente der Mengen gefangen (Schleife über die Mengen, `AGENTS.md` §3.15, §3.17): 26 Schlüssel, 6 abgelehnte Namen und Präfixe, 2 Kommandos.
- [ ] Gegenproben zu den Befunden von Schlüsseln und Kommandos laufen in `make gegenproben`; `SPEC-102` trägt ihre Fallnummern nach.
- [ ] Zusatzzeit gemessen (Gate und Gegenproben, Zahlen in §7).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update: `harness/README.md` (Sensors-Zeile von `make katalog-check` nennt alle drei Kataloge), Handbuch.
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
| `tools/katalog/` | update | Leser von `konfig*.go`, `SPEC-110` Nr. 2 und `cli.go`; Befunde Art 4, 8, 10 für Schlüssel und Kommandos |
| `tools/gegenproben/`, `tools/harness/gegenproben-eingabe.sh` | update | Fälle zu Befunden von Schlüsseln und Kommandos |
| `harness/README.md` | update | Zeile *Sensors* von `make katalog-check` |
| `Dockerfile`, `harness/mk/katalog.mk` | update | Stufe `katalog` kopiert `konfig*.go` und cli.go (`SPEC-201` Nr. 10 a; `.dockerignore` bleibt, `!src/` deckt beide); Kommentare der Stufen und Hilfetext des Ziels nennen alle drei Kataloge |
| `harness/vertraege/SPEC-102-gegenproben.md` | update | Fallnummern der neuen Fälle |
| `docs/user/benutzerhandbuch.md` | update | Abschnitte *Einstellungen* und *Kommandos* mit Tabellen |
| Tests | neu | Mengen aus dem Wortlaut: 26 Schlüssel, 6 Namen und Präfixe der Menge N, 2 Kommandos |

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis im Vertrag `SPEC-201` und in der Tabelle unten):

1. Tabelle *Zusage · Fänger* und Raster beim Start (Zusagen aus `SPEC-201` Nr. 4 bis 8 und 10 a). Hinweise des Planners: Mengen aus dem Wortlaut: 27 Schlüssel von `SPEC-110` Nr. 2 (heute 26 in `konfig*.go`, `SERVICE_ODATA_VERSION` fehlt dort, K ohne C ist kein Befund), 6 abgelehnte Namen und Präfixe (`ADMIN_`, `HYBRID_`, `FALLBACK_`, `LISTEN_TLS_`, `STRICT_COVERAGE`, `PUBLIC_BASE_URL`), 2 Kommandos (`start`, `version`); die Formen sind disjunkt (Schlüssel in Großbuchstaben, Kommandos in Kleinbuchstaben): ein Eintrag in zwei Katalogen mit gleichem Namen gilt nicht.
2. Der Fänger der Leseregel über die Literale in `konfig*.go` über alle Leser (`SPEC-201` Nr. 4 a, 9 e): Sendung von `slice-handbuch-meldungskatalog` (Risiko zur Tabelle der Schlüssel).
3. Größenmarke: Block `groessenmarke` (Teil des Architect) in §4; Hinweis des Planners: *S* grob 80·14 + 400 = 1520 (14 Zeilen), keine Schnittfrage.

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; 22 Zeilen, gefüllt vom Architect am 2026-11-08, alle `andere`: das Programm ist ein einziger Prozess ohne Nebenläufigkeit). „Nr.“ meint die Nummern von `SPEC-201`. Mengen aus dem Wortlaut der Zusage mit ihrer Elementzahl; ein Fänger ist eine Schleife über die Menge, nie eine Auswahl (`AGENTS.md` §3.17). Die Fixtures sind Wegwerfbäume im Test (`t.TempDir`), gebaut aus der Fabrik von `slice-handbuch-meldungskatalog`, erweitert um `konfig*.go`, SPEC-110 und cli.go; sie lesen weder das echte `konfig*.go` noch die echte Ausnahmen-Datei noch `model.AlleCodes()`, und keine Zahl der Codes oder Schlüssel des Repos ist verdrahtet (die 26, 27 und 6 sind Mengen aus dem Wortlaut von `SPEC-110` Nr. 2 als Fixture). Eingaben mit Nicht-ASCII-Zeichen entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18). Jede Mutation fährt der Implementer vor der Übergabe selbst und trägt Zusage, Mutation und roten Test in den Bericht und in einen Absatz unter dieser Tabelle ein (`AGENTS.md` §3.10).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| 1 | C: jedes Zeichenketten-Literal, dessen Wert die Form trifft, ist an jeder Stelle von `konfig*.go` ein Name (Nr. 4 a). Menge 30 Positionen (Absatz *Positionen*) | `TestKatSchlPositionen`: Schleife über 30; Fixture eine Datei mit einem Namen an der Position, K ohne ihn → genau Art 4 an der Zeile des Literals; mit dem Namen in K und im Handbuch → `ok` | Leser besucht nur Argumente von Aufrufen; `case`-Klauseln ausgelassen; Funktionsliterale ausgelassen; Paketebene (`var`, `const`) ausgelassen; Raw-String nicht gelesen; Escape-Folge nicht aufgelöst | andere |
| 2 | C: Form des Namens `^[A-Z][A-Z0-9_]+$`, ASCII, mindestens zwei Zeichen (Nr. 4 a). Menge 16 = 6 gültige (`MO`, `M1`, `MODE_`, `A_B`, `X9_Y`, `ZZ`) + 10 ungültige (`M`, `mode`, `Mode`, `1MODE`, `_MODE`, `MODE-X`, `MODE X`, `MODE` mit Zeilenumbruch im Raw-String, `MÖDE`, `MODE` in Vollbreite; die letzten beiden aus Bytewerten) | `TestKatSchlForm`: Schleife über 16; gültig → Art 4, ungültig → kein Befund und `ok` | `+` als `*` (ein Zeichen zählt); `[A-Z]` als `\p{Lu}`; Anker `^` entfernt; Anker `$` entfernt | andere |
| 3 | C und N: ein Name im Körper jeder Deklaration `nichtGelieferte` ist N, nicht C; außerhalb davon ist er C (Nr. 4 a). Menge 14 Fälle (Absatz *Grenze*) | `TestKatSchlGrenze`: Schleife über 14; erwartet je Fall Art 4 für die Namen außerhalb und keinen Befund für die Namen innen | Grenze am Namen `nichtGeliefert` (Präfix); nur die erste Deklaration; Methode mit Empfänger nicht erkannt; Funktionsliteral im Körper gilt als außerhalb; Literal hinter der schließenden Klammer gilt als innen | andere |
| 4 | W gegen N: ein Name aus N (Name: Gleichheit, Präfix auf `_`: Anfang) ist Art 10 und nie Art 8; ein Name außerhalb von C und N ist Art 8, auch einer aus K ohne C (Nr. 4 a, 4 c, 7). Menge 24 = 6 Elemente von N (4 Präfixe, 2 Namen) × 3 Formen (selbst, mit Anhang `_X`, um ein Zeichen gekürzt) + 6 Art-8-Namen (K ohne C, in keiner Quelle; × erste, mittlere, letzte Zeile der Tabelle) | `TestKatSchlAbgelehnt`: Schleife über 24; erwartet Art 10 bei Gleichheit und beim Anhang an ein Präfix, sonst Art 8, nie beide | alle Elemente als Präfix verglichen (`STRICT_COVERAGE_X` gibt Art 10); alle als Name (`ADMIN_TOKEN` gibt Art 8); Präfix ohne `_` (`ADMIN` gibt Art 10); K ohne C übergangen (kein Art 8) | andere |
| 5 | C: Dateien. Gelesen wird ein Eintrag, der eine Datei ist, mit dem Namen `konfig`, beliebigem Rest und `.go`, nicht auf `_test.go`, nicht rekursiv, in Bytefolge der Namen; ein Name an mehreren Stellen oder in mehreren Dateien ist kein Befund, die Zeile ist die des ersten Vorkommens (Nr. 4 a). Menge 13 = 5 gelesene (`konfig.go`, `konfig_x.go`, `konfigx.go`, `konfig_test_x.go`, Datei mit `//go:build ignore`) + 6 nicht gelesene (`Konfig.go`, `konfig_test.go`, `konfig_x_test.go`, `konfig.go.bak`, `xkonfig.go`, Datei im Unterverzeichnis) + Verzeichnis `konfig_d.go` + Verweis auf eine Datei; dazu derselbe Name in zwei Dateien, in 2 Reihenfolgen der Dateinamen | `TestKatSchlDateien`: Schleife; Namen aus gelesenen Dateien → Art 4, aus den übrigen nichts; Zeile des Befunds = erstes Vorkommen in der Reihenfolge der Dateinamen | `_test.go` mitgelesen; Unterverzeichnis gelesen; Zeile des letzten Vorkommens; Verzeichnisreihenfolge statt Bytefolge; `//go:build` ausgewertet | andere |
| 6 | Eingabefehler von `konfig*.go` und `cli.go` (Nr. 4 a, 5 a, 8). Menge 16 = konfig: Verzeichnis fehlt, `src` ist eine Datei, kein Go (Klammer fehlt, Literal ohne Ende, ohne Paketzeile) 3, zwei Dateien kein Go (je ein Eingabefehler mit Dateiname, Bytefolge), kein Schlüssel (leere Datei, nur Kommentare, nur Namen aus N) 3, eine Datei kein Go und sonst kein Name (`kein Schlüssel gefunden` entfällt) = 10; cli: fehlt, Verzeichnis statt Datei, kein Go 2 (Klammer, Literal ohne Ende), kein Kommando 2 (leere Datei, nur `odata-recorder`) = 6 | `TestKatQuelleEingabe`: Schleife über 16; Exit 2, Text bytegleich, kein Befund | `kein Schlüssel gefunden` auch neben `kein Go`; Scanner statt Parser (Datei ohne Paketzeile gilt); nur der erste Eingabefehler; Dateiname fehlt im Text | andere |
| 7 | K: SPEC-110 lesen, Abschnitt, Zeile, Spans (Nr. 4 b). Menge 38 = 27 Namen (je einer aus K entfernt: Art 4 für die 26 aus C, kein Befund für `SERVICE_ODATA_VERSION`) + 11 Köder (Spans vor `Geprüft werden`; Spans der Folgezeile `Gesetzt …`; Zeile eingerückt; Überschrift `### SPEC-1100 — `; Überschrift mit Bindestrich statt Gedankenstrich; Zeile im Zaun; Zeile hinter `### `; Zeile hinter `## `; Span mit Kleinbuchstaben; Name zweimal in der Zeile; Zaun ohne Schluss hinter der Zeile) | `TestKatSchlK`: Schleife über 38; der Name eines Köders steht in C und nicht in der Zeile → erwartet Art 4; zweimal in der Zeile und Zaun ohne Schluss → `ok` | liest bis zum Ende der Datei; liest die Folgezeile; liest den Inhalt eines Zauns; liest Spans vor `Geprüft werden`; eingerückte Zeile gilt; doppelter Name als Befund | andere |
| 8 | K: Eingabefehler `SPEC-110: Schlüssel fehlen` und `mehrdeutig`, Reihenfolge (Nr. 4 b, 8). Menge 8 = Abschnitt fehlt, Zeile fehlt, `Geprüft werden` fehlt, Zeile ohne Span, Zeile nur mit Spans anderer Form, Abschnitt doppelt, Zeile doppelt, Datei fehlt (ein `spezifikation.md fehlt`, kein zweiter Text); dazu ein Lauf mit Fehler in §4 und in SPEC-110 zugleich | `TestKatSchlKEingabe`: Schleife über 8 und der Doppellauf (§4 vor SPEC-110) | `fehlen` für `mehrdeutig`; Eingabefehler der Datei doppelt gemeldet; Reihenfolge vertauscht | andere |
| 9 | Vergleich C gegen K: C ohne K ist Art 4, K ohne C ist kein Befund (Nr. 1, 4 b, 7). Menge 54 = 27 Namen aus SPEC-110 × 2 Richtungen | `TestKatSchlCgegenK`: Schleife über 54 | Richtung vertauscht; K ohne C als Befund; Art 4 nur für den ersten Namen | andere |
| 10 | H = C: jeder Name aus C hat eine Zeile in der Tabelle `Einstellungen`, sonst Art 7; K ohne C verlangt keine Zeile (Nr. 4 c, 7). Menge 30 = 26 Namen aus C (je Zeile entfernt: genau ein Art 7) + `SERVICE_ODATA_VERSION` ohne Zeile (kein Befund) + 3 Orte, die nicht decken (Zeile nur in Prosa, nur im Zaun, nur in der Tabelle `Kommandos`) | `TestKatSchlPflicht`: Schleife über 30 | Pflicht nur für den ersten Namen; H gleich K statt C (der 27. fordert eine Zeile); Zeile in Prosa deckt | andere |
| 11 | W: Tabelle `Einstellungen` mit Kopf `Schlüssel`, `Bedeutung`; Name aus genau einem Span; Form, Zelle, doppelt (Nr. 4 c, 6 c, 6 d). Menge 11 = Zelle ohne Span, mit zwei Spans, mit Text hinter dem Span (3, Art 2) + Namen `mode`, `ADMIN_*`, `start`, `MODE X` (4, Art 1) + Name in zwei Zeilen (Art 3) + Köder-Tabelle im Zaun + Köder-Tabelle im Abschnitt `Kommandos` + Tabelle ohne Datenzeile (alle 26 fehlen) | `TestKatSchlHandbuch`: Schleife über 11; Befund und Zeile bytegleich | Zelle gilt, wenn sie mit einem Span beginnt; Form für Kleinbuchstaben nicht geprüft; Tabelle im Zaun gelesen; doppelter Name zweimal gemeldet | andere |
| 12 | Kommandos C: Konstanten nach Nr. 5 a. Menge 16 = 6 gelesene (Paketebene einzeln, in einer Gruppe, mit Typ, mehrere Namen in einer Deklaration, in einem Funktionskörper, Raw-String) + 9 nicht gelesene (`odata-recorder` mit Strich, `Start`, `v2`, leerer Text, Verkettung `"sta" + "rt"`, Name einer anderen Konstante als Wert, `var`, `case "stop"`, Vergleich `a == "stop"`) + derselbe Text in zwei Konstanten (ein Kommando, kein Befund) | `TestKatKommLeser`: Schleife über 16; gelesen → Art 7 ohne Zeile im Handbuch, nicht gelesen → kein Befund | nur Paketebene; nur Gruppen; Raw-String nicht gelesen; `var` gelesen; Strich im Muster (`[a-z-]`); doppelter Text als Befund | andere |
| 13 | Kommandos Vergleich: jedes Kommando aus C hat eine Zeile (Art 7), das Handbuch nennt nur Kommandos aus C (Art 8); Kopf `Kommando`, `Bedeutung`; ein Schlüssel in der Tabelle ist Art 1 (Nr. 5 b, 6 d, 7). Menge 9 = 2 Kommandos je Zeile entfernt + 2 unbekannte Zeilen (erste, letzte; Art 8) + `MODE` in `Kommandos` (Art 1) + `start` in `Einstellungen` (Art 1) + Kopf `Schlüssel` statt `Kommando` (Eingabefehler) + Zelle ohne Span + Name doppelt | `TestKatKommVergleich`: Schleife über 9 | Pflicht nur für `start`; Art 8 gegen cli.go statt gegen die Menge C; Form der Kommandos wie die der Schlüssel | andere |
| 14 | Handbuch: Eingabefehler je Abschnitt `Fehlercodes`, `Einstellungen`, `Kommandos` und der Datei (Nr. 6, 8). Menge 27 = 8 Formen × 3 Abschnitte (Abschnitt fehlt, doppelt; Tabelle fehlt, doppelt; Kopf weicht ab; Trennzeile fehlt; Zeile mit 1 Zelle; Zeile mit 3 Zellen) + 3 der Datei (kein UTF-8, NUL, Zaun nicht geschlossen); dazu 2 Läufe: alle drei Abschnitte fehlen (drei Eingabefehler in der Reihenfolge der Kataloge) und `Kommandos` als Unterabschnitt von `Einstellungen` (Tabelle mehrdeutig) | `TestKatHandbuchAbschnitte`: Schleife über 27 und die 2 Läufe | Abschnitt `Kommandos` nicht verlangt; Abbruch nach dem ersten fehlenden Abschnitt; Kopf der dritten Tabelle nicht verglichen; Reihenfolge der Abschnitte vertauscht; Eingabefehler eines Abschnitts neben dem der Datei | andere |
| 15 | Ein Eingabefehler schließt jeden Befund aus; jede Quelle für sich, in der Reihenfolge der Quellen (Nr. 2 b, 8). Menge 7 = 6 Quellen (fehler.go, spezifikation.md, Ausnahmen, konfig*.go, cli.go, Handbuch) × fehlt, bei je einem Befund in jedem der drei Kataloge im übrigen Baum, + ein Lauf mit allen sechs zugleich | `TestKatEingabeAlle`: Schleife über 6 und der Lauf; nur `katalog: Eingabe:`, Exit 2, Reihenfolge bytegleich | Befund neben dem Eingabefehler von `cli.go`; `konfig*.go` hinter `cli.go` gemeldet; Abbruch nach der ersten Quelle | andere |
| 16 | Reihenfolge der Befunde: Katalog (`Fehlercodes`, `Schlüssel`, `Kommandos`), Art, Name in Bytefolge, Zeile (Nr. 2 d). Menge 20 Permutationen (Seed fest) der Zeilen von Handbuch, SPEC-110 und Ausnahmen bei je 6 Befunden in jedem Katalog, 3 Wiederholungen; in der Bytefolge steht `Kommandos` vor `Schlüssel`, die Reihenfolge der Kataloge steht dagegen | `TestKatReihenfolgeKataloge`: Schleife, Ausgabe bytegleich | Katalog nach Bytefolge des Namens; Katalog nach Art; Map-Iteration | andere |
| 17 | Wortlaut der Befunde (Nr. 7): Art 1, 2, 3, 4, 7, 8, 10 für Schlüssel und Art 1, 2, 3, 7, 8 für Kommandos. Menge 12 Texte | `TestKatTexteKataloge`: Schleife über 12, Zeile bytegleich (Quelle `konfig*.go`, `SPEC-110`, `cli.go`) | `konfig*.go` als `konfig.go`; `SPEC-110` als `§4`; `cli.go` als `fehler.go`; Katalog `Schlüssel` ohne Umlaut | andere |
| 18 | Der Stand des Repos ist grün (Nr. 9 f): `make katalog-check` gibt `katalog: ok` mit 26 Zeilen in *Einstellungen* und 2 in *Kommandos*; die Leseregel gilt am echten `konfig*.go` | das Gate selbst; Beleg des Implementers: eine Zeile aus *Einstellungen* und eine aus *Kommandos* entfernt → Rot; Programm mutiert (Leser übergeht `case`, Funktionsliterale oder die Paketebene) → Rot, wo das echte `konfig*.go` die Stelle nutzt; C des Repos gleich 26 nach einem unabhängigen `grep -o` | Zeile entfernt; Zeile für `SERVICE_ODATA_VERSION` hinzugefügt (Art 8) | andere |
| 19 | Gegenproben G16 bis G31 in `make gegenproben` (`SPEC-201`), ein Fall je Befundart und Eingabefehler-Familie. Menge 16 | die Fälle selbst (`SPEC-102`); G1 trägt alle sechs Quellen | Befundart im Programm entfernt → der zugehörige Fall rot (16 Läufe) | andere |
| 20 | Die Tabellen des Handbuchs erfüllen die Mengen A, B, C von `SPEC-200` (Bedeutung in eigenen Worten, keine Kennung) | `make handbuch-check`; Mutation vom Implementer | `noch nicht` in einer Zelle von *Einstellungen*; `SPEC-110` in einer Zelle von *Kommandos* | andere |
| 21 | Entschiedene Randformen halten (Nr. 4 a, 4 b, 5 a, 9 e, 9 h). Menge 5 = `//go:build ignore` wird gelesen; Literal `GET` ist ein Name (rot); `case "stop"` ist kein Kommando (grün); Konstante `ja` ist ein Kommando (rot); Name zweimal in SPEC-110 (grün) | `TestKatEntschieden`: Schleife über 5 | Build-Bedingung ausgewertet; Ausnahmeliste für `GET`; `case` gelesen; Kleinbuchstaben-Konstante übergangen; doppelter Name in K als Befund | andere |
| 22 | Die Prüfungen am selben Namen sind unabhängig (Nr. 7). Menge 7 Paare (Zusatztabelle in §6) | `TestKatSchlPaare`: Schleife über 7 | Art 7 unterdrückt bei Art 4; Art 8 für einen Namen aus N zusätzlich; Art 3 für einen Namen falscher Form; Art 3 unterdrückt Art 8; Art 7 unterdrückt bei Art 2 | andere |

**Positionen** (Zeile 1; 30): Argument 1, 2 und letztes (von vier) eines Aufrufs · Index (`p.env["N"]`) · erstes und letztes Element einer Liste · Schlüssel und Wert einer Abbildung · Feld eines Literals mit Namen und ohne · verschachtelte Liste · `return` (ein und zwei Werte) · Zuweisung `=` und `:=` · `var` auf Paketebene und in einer Funktion · `const` auf Paketebene und in einer Funktion · `case` · Vergleich · Verkettung · Funktionsliteral · `defer` · Senden auf einen Kanal · Umwandlung `[]byte("N")` · Raw-String · Escape-Folge im Wert (`"\x4eAME"`) · Methode mit Empfänger · Aufruf über mehrere Zeilen. Der Fänger legt den Namen je Position in eine eigene Datei.

**Grenze** (Zeile 3; 14): im Körper (a) in der Rückgabe einer Liste, (b) als Argument, (c) in einem Funktionsliteral, (d) Methode mit Empfänger, (e) Funktion mit Typparameter, (f) zweite Deklaration in zweiter Datei (Vereinigung beider); (g) derselbe Name im Körper und außerhalb (C und N); außerhalb, also C: (h) Literal in der Funktion unmittelbar vor der Deklaration, (i) Literal unmittelbar hinter der schließenden Klammer, (j) `nichtGeliefert` ohne `e`, (k) `NichtGelieferte`, (l) `var nichtGelieferte = func() {…}`; N leer, kein Befund: (m) Deklaration ohne Körper, (n) keine Deklaration (die Namen bleiben dann in C).

Eine neue Kennung entsteht nicht: Nr. 4 und 5 stehen in `SPEC-201`, `SPEC-202` bleibt frei. Raten: kein Fänger ist wahrscheinlichkeitsbehaftet, das Programm ist ein einziger Prozess ohne Nebenläufigkeit (`AGENTS.md` §3.10); der Implementer trägt unter diese Tabelle einen Absatz *Mutationen* mit Zusage, Mutation und rotem Test ein.

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write` (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-handbuch-meldungskatalog` liegt in `done/` (das Programm `katalog`, die Stufen, das Ziel `make katalog-check` und die Gate-Kette sind sein Gegenstand). Der Architect hat Tabelle und Raster gefüllt, bevor der erste Code-Commit entsteht.

**Bedingung Gate-Kette** (Übergangsregeln Nr. 7, Planner, 2026-11-06): Dieser Slice ändert dieselben Dateien der Gate-Kette wie `slice-handbuch-meldungskatalog` (`Dockerfile`, dem Fragment `katalog.mk` unter `harness/mk/`, `harness/README.md`, `tools/gegenproben/`) und **startet nicht daneben, sondern erst nach dessen `done/`**. Kein weiterer Slice der Gate-Kette steht in `in-progress/`, solange dieser läuft. Kennungen aus dem Bereich der Linie C (`SPEC` 200 bis 229, `ADR` 0030 bis 0039, `CO` 040 bis 059): die nächste freie `SPEC`-Nummer der Linie C ist `SPEC-202`. Gate-Zeit: `make gates` und `make gegenproben` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Linien zugleich.

**Stand beim Start** (Planner, 2026-11-08, `main` = `e5c0f8b9`): erfüllt. `slice-handbuch-meldungskatalog` liegt in `done/`, das Programm `katalog`, die Stufen und `make katalog-check` stehen auf `main`. Nr. 7 der Übergangsregeln: kein anderer Slice der Gate-Kette in `in-progress/` (`slice-admin-rahmen` und `slice-mock-abnahme` ändern weder `GATE_CHECKS` noch ein Gate); `slice-admin-rahmen` ändert `harness/katalog-ausnahmen.txt` und `model.AlleCodes()`, Eingaben des Fehlercode-Katalogs, nicht die Kataloge der Schlüssel und Kommandos.

**Größenmarke** (Planner, 2026-11-08, `SPEC-128` Nr. 1; Prosa-Form, der Block folgt vom Architect und wird vom Planner in eigenem Commit vor dem ersten Code-Commit bestätigt): 14 Zeilen der Tabelle, *a* = 0, *b* = 14, *f* = 0, *p* = 400: *S* = 85·0 + 80·14 + 200·0 + 400 = 1520, *M* = 100·⌈135·1520/10000⌉ = 2100, keine Schnittfrage (*S* < 2223). Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade des Blocks, Tests eingeschlossen. Der Architect legt den Block `groessenmarke` (Pfade, `produktcode`, `fakes`) beim Start an; ändert sich die Zeilenzahl der Tabelle, rechnet der Planner neu. `basis`, `schaetzung` und `marke` setzt der Planner.

**Stand der Tabelle beim Start** (Architect, 2026-11-08): Die Tabelle von §3 hat 22 Zeilen, alle `andere` (*a* = 0, *b* = 22), *f* = 0. Die Annahme *b* = 14 trägt die Mengen nicht: jede Zusage von `SPEC-201` Nr. 4 bis 8 und 10 a mit ihrer Menge (30 Positionen, 14 Grenzfälle, 13 Dateiformen, 38 Fälle von SPEC-110, 27 Eingabefehler des Handbuchs) braucht eine Zeile, die Gegenproben und die Verdrahtung je eine. *p* = 400 bestätigt (Leser von `konfig*.go` mit Parser und Grenze etwa 140, SPEC-110 etwa 60, cli.go etwa 50, Vergleich, Befund-Texte und Ausgabe etwa 90, Handbuch-Abschnitte etwa 60). Damit ist *S* = 85·0 + 80·22 + 200·0 + 400 = 2160 und *M* = 100·⌈135·2160/10000⌉ = 3000; die Schnittfrage entfällt (*S* < 2223), eine 23. Zeile gäbe *S* = 2240 und damit die Schnittfrage. Metrik wie oben; das Handbuch und `harness/vertraege/` zählen nach `SPEC-128` Nr. 3 m und 4 a nicht; der Block unten trägt den Teil des Architect, `basis`, `schaetzung` und `marke` setzt der Planner (Rückgabe in §6, Ausgang *offen*).

```groessenmarke
pfad: tools/katalog
pfad: tools/gegenproben
pfad: tools/harness/gegenproben-eingabe.sh
pfad: Dockerfile
pfad: harness/mk/katalog.mk
pfad: harness/README.md
produktcode: 400
fakes: 0
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt.
- `in-progress` → `open` (blockiert — Carveout?): der Vertrag `SPEC-201` Nr. 4 oder 5 verlangt eine Änderung von `konfig*.go` oder `cli.go`; das ist nach dem Vertrag nicht vorgesehen (der Abgleich liest, er ändert nichts).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` und `make gegenproben` sind grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Die Schlüssel stehen als Literale an vielen Stellen von `konfig*.go`, nicht in einer Tabelle (Sendung von `slice-handbuch-meldungskatalog`, Leseregel `SPEC-201` Nr. 4 a, 9 e) — **Ausgang:** offen; Fänger entworfen (Architect, 2026-11-08): §3 Zeilen 1, 3, 5 und 18 (30 Positionen, 14 Grenzfälle, 13 Dateiformen und das Gate am echten `konfig*.go`); der Implementer sieht sie rot.
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster entsteht beim Start vor dem ersten Code-Commit, jede Zelle mit Ort in `SPEC-201` oder „gilt nicht“ mit Grund; eine Zelle ohne Ort geht an den Architect — **Ausgang:** entfallen: Raster gefüllt (unten), Orte in `SPEC-201` Nr. 4 a, 4 b, 4 c, 5 a, 5 b, 6 b, 8, 9 e und 9 h nachgetragen (Architect, 2026-11-08); eine im Code gefundene Zelle ohne Ort öffnet das Risiko neu.
- Eine Allaussage des Vertrags („jeder Schlüssel in C steht in K“) ohne Fänger über alle Fälle (`BEO-REPO/universalaussage-im-vertrag-ungeprueft`) — **Ausgang:** offen; Schleifen über die Mengen (DoD, zweiter Punkt; §3 Zeilen 1 bis 22).
- Eine Welle macht einen Schlüssel oder ein Kommando erreichbar und ändert Ausnahmen und Handbuch nicht (Pflicht der Folgewellen, `SPEC-201` Nr. 11) — **Ausgang:** weiter offen (akzeptiertes Negativ des Vertrags, Nr. 9 c); Adresse: der User über den Diff-Vorschlag zu `AGENTS.md` aus `slice-handbuch-meldungskatalog`.
- Die DoD sagte 27 Zeilen in *Einstellungen* (§2, zweiter Punkt); der Vertrag lässt nur 26 zu: ein Name aus K ohne C in der Tabelle ist Befund Art 8 (`SPEC-201` Nr. 1, 4 c), und das Handbuch sagt nur den Ist-Zustand (`SPEC-200` Nr. 1) — **Ausgang:** entfallen: die Entscheidung gilt, die DoD ist berichtigt (Architect, 2026-11-08); die Zeile für `SERVICE_ODATA_VERSION` trägt der Slice, der den Leser liefert (`SPEC-201` Nr. 11).
- §3 nannte `Dockerfile` und `harness/mk/katalog.mk` nicht, obwohl `SPEC-201` Nr. 10 a die Stufe `katalog` um `konfig*.go` und cli.go erweitert — **Ausgang:** entfallen: Zeile in §3 ergänzt (Architect, 2026-11-08).
- Größenmarke: die Annahme der Prosa (14 Zeilen, *S* = 1520, *M* = 2100) trägt die Tabelle nicht, sie hat 22 Zeilen (*S* = 2160, *M* = 3000, §4) — **Ausgang:** offen; Adresse: Planner (bestätigt oder ändert den Block, setzt `basis`, `schaetzung` und `marke` in eigenem Commit vor dem ersten Code-Commit).
- Merge-Nähe: ein Slice, der vor dem Merge dieses Slice einen Schlüssel in `konfig*.go` oder eine Konstante in cli.go einführt, macht `make katalog-check` auf dem Merge rot (Art 4 und 7 bzw. 7), bis seine Zeile im Handbuch steht. Heute ändert keiner der Slices in `in-progress/` diese Dateien; `slice-admin-rahmen` ändert `fehler.go` und `harness/katalog-ausnahmen.txt`, Eingaben des Katalogs der Fehlercodes, die dieser Slice nicht anfasst, und die Fänger dieses Slice bauen Wegwerfbäume (weder das echte `model.AlleCodes()` noch die echte Ausnahmen-Datei noch eine Zahl der Codes sind Eingabe) — **Ausgang:** weiter offen; Adresse: Planner (Merge-Reihenfolge; Pflicht der Folgewellen `SPEC-201` Nr. 11).

**Raster *Operation × Argumentart*** (gefüllt vom Architect am 2026-11-08, Stand beim Start; „Nr.“ meint `SPEC-201`; „gilt nicht“ trägt seinen Grund, den Abbruch eines Prozesses ohne Kontext, der nur liest und nichts schreibt, Nr. 2 c):

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Schlüssel C lesen (`konfig*.go`) | Verzeichnis fehlt: `konfig*.go fehlt` (Nr. 8); keine Datei oder nur Namen aus N: `kein Schlüssel gefunden` (Nr. 4 a); keine Deklaration `nichtGelieferte`: N leer, kein Befund (Nr. 4 a) | leere Datei oder Literal `""`: kein Name, ohne Namen insgesamt `kein Schlüssel gefunden` (Nr. 4 a) | Literal anderer Form: kein Name (Nr. 4 a); Name der Form, der kein Schlüssel ist (`GET`): rot, Art 4 (Nr. 9 h); Build-Bedingung wird gelesen (Nr. 9 h); `_test.go` nicht gelesen (Nr. 4 a); kein Go (Nr. 8) | gilt nicht: kein Kontext, nichts geschrieben oder zu räumen (Nr. 2 c) | Name an mehreren Stellen oder in mehreren Dateien: kein Befund, Zeile des ersten Vorkommens (Nr. 4 a); mehrere Deklarationen `nichtGelieferte`: Vereinigung (Nr. 4 a) |
| Schlüssel K lesen (SPEC-110) | Abschnitt, Zeile, `Geprüft werden` oder Name fehlt: `SPEC-110: Schlüssel fehlen` (Nr. 4 b); Datei fehlt: `spezifikation.md fehlt` (Nr. 8) | Zeile ohne Span: `SPEC-110: Schlüssel fehlen` (Nr. 4 b) | Span anderer Form: weder Name noch Befund; Spans vor `Geprüft werden` und der Satz `Gesetzt …`: nicht Teil von K (Nr. 4 b, 9 c) | gilt nicht: kein Kontext (Nr. 2 c) | Name zweimal in der Zeile: Satz, kein Befund (Nr. 4 b); Abschnitt oder Zeile doppelt: `SPEC-110: Schlüssel mehrdeutig` (Nr. 4 b) |
| Abgelehnte Namen N lesen (Körper `nichtGelieferte`) | keine Deklaration oder ohne Körper: N leer, kein Befund (Nr. 4 a) | Körper ohne Literal der Form: N leer (Nr. 4 a) | Literal im Körper, das die Form verfehlt: kein Element (Nr. 4 a); Element ohne `_` am Ende: Name, nicht Präfix (Nr. 4 a, 4 c) | gilt nicht: kein Kontext (Nr. 2 c) | Name im Körper und außerhalb: in C und in N, Beziehungen unabhängig (Nr. 4 a, 7); mehrere Deklarationen: Vereinigung (Nr. 4 a) |
| Kommandos C lesen (cli.go) | Datei fehlt: `cli.go fehlt` (Nr. 8); keine Konstante der Form: `cli.go: kein Kommando gefunden` (Nr. 5 a) | Konstante mit leerem Text verfehlt `^[a-z]+$`, kein Kommando (Nr. 5 a) | Programmname mit Strich, Großbuchstaben, Ziffer, Verkettung, `var`, `case`: kein Kommando (Nr. 5 a, 9 e); Konstante in Kleinbuchstaben, die kein Kommando ist: rot (Nr. 9 h); kein Go (Nr. 8) | gilt nicht: kein Kontext (Nr. 2 c) | zwei Konstanten mit gleichem Text: ein Kommando, kein Befund (Nr. 5 a) |
| Handbuch lesen (Einstellungen, Kommandos) | Datei, Abschnitt oder Tabelle fehlt: Eingabefehler, jeder Abschnitt für sich (Nr. 6 a bis c, 8) | Tabelle ohne Datenzeile: W leer, jeder Name aus H fehlt (Nr. 6 c); erste Zelle leer oder ohne Span: Art 2 (Nr. 6 d) | Name anderer Form, auch ein Kommando in *Einstellungen* und ein Schlüssel in *Kommandos*: Art 1 (Nr. 4 c, 5 b, 6 d); anderer Kopf, andere Zellenzahl, Zaun ohne Schluss: Eingabefehler (Nr. 6 a, 6 c) | gilt nicht: kein Kontext (Nr. 2 c) | Name in zwei Zeilen: Art 3 (Nr. 6 d); Abschnitt zweimal: mehrdeutig (Nr. 6 b); Abschnitt im Abschnitt: Tabelle mehrdeutig (Nr. 6 b) |
| Mengen vergleichen (Schlüssel, Kommandos) | Name aus H ohne Zeile: Art 7 (Nr. 7); Name aus K ohne C: kein Befund (Nr. 1, 4 b) | leeres W: jeder Name aus H fehlt (Nr. 6 c); leeres C: Eingabefehler, kein Vergleich (Nr. 4 a, 5 a, 8) | Name in W außerhalb von C und N: Art 8; in N: Art 10 (Nr. 4 c, 7); Name aus K ohne C in W: Art 8 (Nr. 1, 4 c) | gilt nicht: kein Kontext (Nr. 2 c) | derselbe Lauf zweimal gibt dieselbe Ausgabe, die Reihenfolge der Quellzeilen wirkt nicht (Nr. 2 d); mehrere Befunde am selben Namen sind unabhängig (Nr. 7, Zusatztabelle) |

**Zusatztabelle: Paare von Verstößen am selben Namen** (`SPEC-201` Nr. 7: „kein Befund unterdrückt einen anderen“; Fänger `TestKatSchlPaare`, Zeile 22 der Tabelle von §3):

| Paar | erwartet | Fänger |
|---|---|---|
| Name in C, nicht in K, ohne Zeile im Handbuch | Art 4 und Art 7 | `TestKatSchlPaare` Fall 1 |
| Name in C und in N, nicht in K, mit Zeile im Handbuch | Art 4 und Art 10, weder Art 7 noch Art 8 | Fall 2 |
| Name nicht in C und nicht in N, in zwei Zeilen des Handbuchs | Art 3 und Art 8 (einmal, an der Zeile der ersten Nennung) | Fall 3 |
| Name in N, in zwei Zeilen des Handbuchs | Art 3 und Art 10 (einmal, an der Zeile der ersten Nennung) | Fall 4 |
| Name falscher Form, in zwei Zeilen des Handbuchs | zwei Art 1, kein Art 3 | Fall 5 |
| Name in C, dessen Zeile eine Zelle ohne Span trägt | Art 2 und Art 7 | Fall 6 |
| Name in K, nicht in C, im Handbuch genannt | Art 8, kein Art 4 | Fall 7 |

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-06): `BEO-REPO/randform-erst-im-review-entschieden` (`AGENTS.md` §3.16) und `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (`AGENTS.md` §3.15, §3.16) gelten; der Planner zieht die Zähler beim Start nach.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
