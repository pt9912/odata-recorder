# Verifikation: slice-handbuch-katalog-schluessel-kommandos — 2026-11-08

**Art:** Verifikation (einmal) — gegen DoD, `SPEC-201` Nr. 1, 4, 5, 7, 8 und `SPEC-102`, die Tabelle *Zusage · Fänger* und die Hard Rules; nicht Diff-Review (siehe `2026-11-08-slice-handbuch-katalog-schluessel-kommandos-review.md`).

**Gegenstand:** Zweig `slice-handbuch-katalog-schluessel-kommandos`, HEAD `c76a0a05` (Fix-Runde `136c350c` nach Review, Entscheid des Architect `90b5434c`), Basis `fc5de8c`. Plan: `docs/plan/planning/done/slice-handbuch-katalog-schluessel-kommandos.md`.

**Rolle/Modell:** verifier · claude-sonnet-5-5.

## Urteil

**Bestanden mit Auflagen: ein Befund MEDIUM (V-1, Code weicht von `SPEC-201` Nr. 8 ab), Fänger-Lücken LOW bis MEDIUM an Elementen der Klassen, die die Fix-Runde nicht nannte (V-2), keine Abweichung der Gates.** `make gates` ist grün, die Größenmarke hält (2118 von 3000), die Gegenproben stehen auf dem Stempel der aktuellen Eingabe. Die Fix-Runde deckt die Klassen T1, T5, T6 und T7 in meinen Stichproben, T2 bis T4 mit den unten genannten Rändern. Eine weitere Fix-Runde ist für V-1 nötig (Code oder Vertrag, Entscheid beim Architect); V-2 kann der Implementer in derselben Runde schließen oder der Planner als akzeptiertes Negativ eintragen.

## 1. Belege geprüft, nicht Behauptungen

| Prüfung | Beleg (selbst gefahren) |
|---|---|
| `make gates` | einmal, unter `flock /tmp/odata-recorder-heavy.lock`, Log `/tmp/claude-1000/-Development-odata-recorder/eb0bdee1-0646-479c-9f60-077c8f95bfc6/scratchpad/gates.log`: ohne `make: ***`, 32,6 s (Cache warm); `katalog` (Stufe 8/8 `RUN katalog /katalog`), `handbuch`, `a-check`, `docs-check`, `lint`, `test`, `test-race`, `zeichen` im Lauf. |
| `make groessenmarke` | im selben Lauf: `Messung 2118 von Marke 3000`; eigene Rechnung `git diff --numstat fc5de8c..HEAD` über die sieben Pfade des Blocks: 2118. |
| `make gegenproben` | `Eingaben unverändert seit dem letzten grünen Lauf (b6e3d4ad…)`, 1,4 s. Ich habe die 31 Fälle nicht neu gefahren; der Stempel ist die Bestätigung des Implementers, die Eingabe hat sich seit seinem Lauf nicht geändert (Hash gleich). Die Zeit 169 bis 171 s habe ich deshalb **nicht** nachgemessen. |
| Handbuch | Tabelle *Einstellungen* zählt 26 Zeilen, `SERVICE_ODATA_VERSION` fehlt (richtig nach Nr. 4 c); `make handbuch-check` im Lauf grün. |
| Sensors-Zeile, Dockerfile, `katalog.mk` | `harness/README.md` nennt alle drei Kataloge, Stufe `katalog` kopiert `konfig*.go` und `cli.go`, Hilfetext des Ziels nennt Schlüssel und Kommandos; `.dockerignore` unverändert (`!src/` deckt beide). |
| Rohbytes (`AGENTS.md` §3.18) | dieser Bericht mit `LC_ALL=C grep -P` auf Backslash-u-Folge, BOM, NBSP, U+2028, U+200B gegengelesen: 0 Treffer; `make zeichen-check` im Lauf grün. |
| Merge-Nähe | `git merge-tree --write-tree HEAD main`: ein Konflikt, allein `harness/gegenproben.stempel` (abgeleitet, nach dem Merge neu zu erzeugen). Auf `main` (43 Commits) sind `konfig*.go`, `cli.go` und das Handbuch unverändert; neu ist `ODR-E2102` in `fehler.go`, `harness/katalog-ausnahmen.txt` und `spec/spezifikation.md` (15 Treffer), also grün zu erwarten. Gates auf dem Merge-Stand habe ich nicht gefahren. |

## 2. DoD (§2 des Plans)

| Punkt | Stand | Beleg |
|---|---|---|
| Kataloge im Programm `katalog`, jede Zusage mit Fänger, rot gesehen | erfüllt mit Rändern (V-1, V-2) | 22 Zeilen der Tabelle; Fänger laufen als Schleifen; meine Mutationen unter 3. |
| Handbuch: *Einstellungen* (26) und *Kommandos* (2), Allaussage über beide Richtungen | erfüllt | Handbuch gezählt; Fänger `TestKatSchlCgegenK`, `TestKatSchlPflicht`, `TestKatSchlAbgelehnt`, `TestKatKommVergleich`. |
| Gegenproben G16 bis G31 in `make gegenproben`, `SPEC-102` trägt Fallnummern | erfüllt | Fälle 189 bis 204 in `SPEC-102` und `tools/gegenproben/katalog_schluessel.go` gleich nummeriert; Stempel aktuell. |
| Zusatzzeit gemessen (Zahlen in §7) | **offen** | §7 ist leer (*Wird bei der Closure gefüllt*); die Zeit steht nur in §6 und `SPEC-102`. |
| `make gates` grün | erfüllt (eigener Lauf) | s. o. |
| Review durchgeführt | erfüllt | Report liegt vor; Fix-Runde danach. |
| Doku-Update (`harness/README.md`, Handbuch) | erfüllt | s. o. |
| Closure-Notiz, Register, Ausgänge der Risiken, drei Paarungen | offen, Sache des Planners | Haken fehlen noch alle; §6 trägt Zeilen mit *offen* bzw. *weiter offen* (siehe 5). |

Liefer-Punkte: drei, am Limit.

## 3. Plan-vs-Code-Diff und Tabelle gegen den Vertrag

Der Diff `fc5de8c..HEAD` berührt genau die Pfade, die §3 nennt, dazu `SPEC-102`, `SPEC-201`, `harness/gegenproben.stempel` und das Handbuch. `tools/harness/gegenproben-eingabe.sh` ist unverändert, wie §3 es sagt. Ich habe jede der 22 Zeilen gegen den Wortlaut von `SPEC-201` Nr. 1, 4, 5, 7 und 8 gelegt. Jede Aussage hat eine Zeile; der Code folgt dem Entscheid des Architect (H nimmt den Namen am Präfix in `abgelehnt`, die Zeile eines Go-Fehlers ist die der Datei in `parseGo`, Art 7 und 10 nie zugleich). Abweichung zwischen Vertrag und Code: V-1. Zeile 6 der Tabelle nennt als Fänger, dass `kein Schlüssel gefunden` neben `kein Go` entfällt, aber nicht neben einem Verweis ohne Ziel: die Zusage ist in der Tabelle enger als im Vertrag.

**Mutationen an nicht genannten Elementen** (Wegwerfkopie von `tools/katalog` in `.../scratchpad/base`, je Mutation frische Kopie, `go test -count=1 ./...` im Container `odata-recorder:race`, `--network none --memory=3g --cpus=4`, Quelle schreibgeschützt eingehängt, unter `flock`; Nullprobe grün; ein Prozess, kein Fänger wahrscheinlichkeitsbehaftet; Messender der Verifier; Load Average des Hosts zu Beginn 11,3):

| Klasse | Mutation (nicht im Befund oder im Bericht genannt) | Ergebnis |
|---|---|---|
| T1 Knotenarten | `IndexListExpr`, `Ellipsis`, `LabeledStmt` übersprungen | rot (`TestKatSchlKnotenarten`) |
| T1 Tiefe | Verschachtelung über 9 Ebenen nicht gelesen | rot; über 12 und über 20 **grün** (V-2 e) |
| T2 Form | Länge 1 statt 2; letztes Zeichen ab 0x80 gilt; Länge über 24 abgelehnt | rot, rot, rot (die letzte zufällig in `TestKatSchlK`); über 40, über 64 **grün** (V-2 d) |
| T2 Kommando | Zeichen ab 0x80 gelten; Länge über 8 und über 10 abgelehnt | alle drei **grün** (V-2 d) |
| T3 Dateien/Überschrift | Überschrift `### SPEC-110 — ` auch im Zaun; Ende `##` auch im Zaun (zwei Formen); Ende mit eingerückter Zeile | alle vier **grün** (V-2 b) |
| T4 Konstanten | Konstante in Funktionsliteral, in `if`-Block, in `case`-Klausel nicht gelesen; Wert in Klammern `("neu")` gilt als Literal | alle vier **grün** (V-2 a, c) |
| T4 Index | `i >= len(Values)-1` (eine Stelle zu früh) | rot (`TestKatAusgabe`; der Plan nennt `TestKatKommWertarten`, aber rot wird es) |
| T5 erste | Zeile des letzten Go-Fehlers (`max` statt `min`) | rot (`TestKatErsteNennung`) |
| T6 Präfix | Element von N mit `_` irgendwo gilt als Präfix; Namen mit beliebiger Stelle statt Anfang; Präfix gilt nur für längere Namen | rot, rot, rot (`TestKatSchlAbgelehnt`, `TestKatSchlPraefix`) |
| T7 `//line` | Zeile des Go-Fehlers mit Direktive | rot (`TestKatLineDirektive`) |

Von 28 Mutationen sind 13 rot und 15 grün; die grünen liegen in den Klassen T1 (Tiefe) und T2 bis T4 (siehe V-2). Klasse und Elementzahl: T1 46 Knotenarten und 59 Stellen (Stichprobe 3 von 46 Arten rot, dazu die Tiefe), T2 Kommando 128 ASCII-Zeichen (Stichprobe außerhalb ASCII und der Länge), T3 38 Fälle, T4 31 Fälle, T5 20 Fälle, T6 20 Fälle, T7 12 Fälle.

## 4. Befunde

### V-1 (MEDIUM) — Verweis ohne Ziel: Meldung `kein Schlüssel gefunden` und Reihenfolge der Eingabefehler weichen von `SPEC-201` Nr. 8 ab

Der Vertrag (Nr. 8) sagt: `konfig*.go: kein Schlüssel gefunden` steht nur, wenn keine Datei einen Eingabefehler trägt, und die Eingabefehler je Datei stehen in der Bytefolge der Namen; ein Verweis ohne Ziel ist ein Eintrag, der nicht zu lesen ist (Nr. 4 a, 8). `konfigDateien` meldet den Verweis vor dem Lesen aller Dateien, und `liesKonfig` setzt `vorher` erst danach. Gefahren am echten Baum (Container, Kopie der sechs Quellen):

- Verweis `konfig_l.go` ohne Ziel als einzige Datei: zwei Zeilen, `konfig_l.go: no such file or directory` **und** `kein Schlüssel gefunden`, Exit 2. Der Vertrag verlangt nur die erste.
- Verweis `konfig_l.go` ohne Ziel plus `konfig_a.go` ohne Go: `konfig_l.go` steht **vor** `konfig_a.go`; die Bytefolge verlangt `konfig_a.go` zuerst.

Der Fänger `TestKatSchlDateien` (Verweis ins Leere) hat im Baum andere Dateien mit Namen, darum sieht er keine der beiden Formen. Zusage und Fänger der Tabelle (Zeile 5 und 6) sind enger als der Vertrag. **Adresse:** Architect/Implementer (Code in einer Schleife über die Einträge oder Vertrag nach dem Code); in §6 mit Ausgang *offen*, bevor der Bericht übergeht (`AGENTS.md` §3.13). Fänger nach dem Fix: Schleife über Verweis × {einzige Datei, mit Datei ohne Go davor, danach}.

### V-2 — Fänger-Lücken an Elementen der Klassen (`AGENTS.md` §3.17); der Code ist in allen Fällen mit dem Vertrag vereinbar, es fehlt der Fall

| | Klasse (Wortlaut) | Lücke | Grad |
|---|---|---|---|
| a | T4 Stelle einer Konstante: „auf Paketebene oder in einem Funktionskörper“ (`SPEC-201` Nr. 5 a) | `TestKatKommWertarten` und `TestKatKommLeser` führen nur die oberste Ebene des Körpers; Funktionsliteral, `if`-Block, `case` im Körper sind ungeprüft (C_funclit, C_if, C_switch grün) | MEDIUM: die Zusage „in einem Funktionskörper“ meint alle Verschachtelungen |
| b | T3 Zaun und Einrückung bei der Überschrift und dem Ende von SPEC-110 (Nr. 4 b: „zählen nur außerhalb von Zäunen“) | Köder für die Überschrift im Zaun, `##` und `###` im Zaun als Ende, eingerückte Zeile als Ende fehlen (Z_ueb_zaun, Z_end_zaun, E_trim grün). Der Köder „Zeile im Zaun“ trägt nur die Zeile `2. **…**` | MEDIUM: der Vertrag nennt die Zaun-Regel für alle drei Stellen, der Fänger führt eine |
| c | T4 Wertform: „ein Wert, der kein Literal ist (eine Verkettung, ein Name)“ (Nr. 5 a) | die sieben Wertformen führen keine Klammer; `("neu")` gäbe ein Kommando, bliebe grün | LOW |
| d | T2 Zeichenklassen und Länge: Schlüssel „ASCII, mindestens zwei Zeichen“ (Nr. 4 a), Kommando `^[a-z]+$` (Nr. 5 a) | 128 ASCII-Zeichen an sechs Stellen, Länge bis 3; Nicht-ASCII nur beim Schlüssel (`MÖDE`, Vollbreite), nicht beim Kommando (K_nonascii grün); keine obere Länge (F_len40, F_len64, K_len10 grün) | LOW: die Regex hat keine obere Grenze; Nicht-ASCII beim Kommando ist die Gegenseite von T2 |
| e | T1 Tiefe der Verschachtelung | `ast.Inspect` hat keine Grenze, kein Fall führt mehr als etwa zehn Ebenen (D_tief12 grün) | LOW: die Menge ist unbeschränkt; **Vorschlag: akzeptiertes Negativ** mit Grund (Syntax wächst, der Leser hat keine Tiefengrenze) und Folge (eine künftige Grenze im Leser fiele erst im Betrieb auf), analog zur Prüfgrenze der Knotenarten in §3 |

Vorschlag zu a bis d: ein Fall je Element im Fänger (Funktionsliteral, `if`, `case`; Überschrift im Zaun, Ende im Zaun, Ende eingerückt; Klammer; ein Nicht-ASCII-Byte im Kommando); e als Negativ. Der Aufwand ist klein (etwa 40 Zeilen); die Marke 3000 trägt ihn (Reserve 882).

## 5. Offene Risiken aus §6 mit Ausgang *offen* — Vorschlag, keine Entscheidung

1. **Zeit der Gegenproben** (169,3 s und 171,1 s gegen das Budget von 120 s, `SPEC-102` Zeitbudget; kein Gate prüft es; Lockerung ist ein ADR, `AGENTS.md` §3.6). Ich habe sie nicht neu gemessen (Stempel unverändert, 1,4 s). `SPEC-102` führt für frühere Slices 143 s und 127 s bei Last ähnlicher Größe, das Budget ist also **vor** diesem Slice schon gerissen worden; die Zusatzzeit der 16 Fälle ist laut `SPEC-102` klein, der Befund hängt an der Last. **Vorschlag an User/Planner:** Entscheid per ADR in einem eigenen Schritt, nicht in diesem Slice: entweder das Budget auf eine Zahl setzen, die unter definierter Last (Kernzahl, Load Average) wiederholbar hält, oder das Budget als Soll mit Messvorschrift führen. Der Slice kann mit Ausgang *weiter offen* und Adresse (User) nach `done/`, wenn §7 die Zahlen trägt; die DoD verlangt Messung, keine Einhaltung.
2. **Merge-Nähe.** Geprüft: `main` ändert `konfig*.go`, `cli.go` und das Handbuch nicht; `ODR-E2102` ist in `fehler.go`, Ausnahmen und `spec/spezifikation.md` zugleich eingetragen. **Vorschlag an den Planner:** diesen Slice vor jedem Slice mergen, der einen Schlüssel oder ein Kommando einführt; beim Merge den Stempel neu erzeugen (`make gegenproben`) und `make gates` auf dem Merge-Stand einmal fahren; den Ausgang danach auf *entfallen* setzen.
3. **Literale in `konfig*.go` (Sendung des Vorgängers, erste Zeile des §6).** Die Fänger sind da (46 Knotenarten, 30 Positionen); offen bleiben V-2 a, b und e. **Vorschlag:** kein Schlüssel blieb ungesehen; nach V-2 a, b und dem Negativ e *entfallen* mit Verweis auf die Prüfgrenze in §3.
4. **Allaussage ohne Fänger** (dritte Zeile des §6): nach V-1 und V-2 *entfallen*; vorher nicht.

## 6. Auflagen an den Planner (Closure)

- V-1 und V-2 a bis e an den Architect/Implementer zurück, in §6 mit Ausgang *offen* und Adresse (`AGENTS.md` §3.13).
- §7: Closure-Notiz mit der Zusatzzeit (Zahlen), Lerneintrag, Register (`BEO-REPO/universalaussage-im-vertrag-ungeprueft` und `BEO-REPO/randform-erst-im-review-entschieden` zählen weiter: V-1 ist eine Randform, die der Vertrag entschied und der Code nicht hielt), alle DoD-Haken.
- Absatz *Mutationen und Raten* des Plans: die Mutation W_index2 wird in `TestKatAusgabe` rot, nicht in `TestKatKommWertarten` (klein).
