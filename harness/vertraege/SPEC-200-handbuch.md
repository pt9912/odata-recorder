# SPEC-200 — Benutzerhandbuch (`make handbuch-check`)

**Werkzeug:** `make handbuch-check`; die Prüfung ist `pruefung handbuch` aus `tools/pruefung/` (`SPEC-036`, eigenes Go-Modul ohne Abhängigkeiten), gefahren in der Stufe `handbuch` des `Dockerfile` (abgeleitet von der Stufe `build`, jeder `RUN` ohne Netz).
Das Ziel hängt an `GATE_CHECKS`; `make gates` fährt es.

**Festlegung:** Das Ziel prüft das Handbuch (docs/user/benutzerhandbuch.md) gegen das gebaute Binary und gegen die Regel des Handbuchs (Nr. 1): Was das Handbuch über das Binary sagt, ist belegt; es sagt nichts Geplantes, keine Chronik und nennt keine Kennung der Planung.
Es hängt nicht am Abgleich der Kataloge: Die Regel gilt vom ersten Satz an, und die Prüfung braucht kein Binary, das Schlüssel liest oder Fehlercodes erzeugt.
Grün ist das Ziel nur mit einem Handbuch; ohne Datei ist es rot (Nr. 2 c).

1. **Regel des Handbuchs.** Der Wortlaut ist die bindende Entscheidung des Users:
   *Das Handbuch beschreibt ausschließlich den Ist-Zustand des gebauten Binary der angegebenen Software-Version: nichts Geplantes, keine Ausblicke, keine Bezüge auf Roadmap, Wellen, Slices, ADRs oder Kennungen der Spezifikation, keine Chronik.*
   Eine Aussage des Handbuchs gehört genau einer Klasse an; die Klasse nennt den Beleg:

   | Klasse | Aussage | Beleg |
   |---|---|---|
   | K1 Aufruf | Kommando, Argumente, Ausgabe je Strom und Ausgangswert des Binary | ein gebundenes Beispiel (Nr. 5, 6), gefahren gegen das gebaute Binary (Nr. 7) |
   | K2 Version | die Software-Version im Kopf | Vergleich mit der Ausgabe von `version` desselben Binary (Nr. 4) |
   | K3 Host-Befehl | `make <Ziel>` in Inline-Code | das Ziel steht in der Zielliste (Nr. 9); die Wirkung belegt das Gate des Ziels selbst |
   | K4 Aufrufpräfix | der Aufruf des Binary im Image: `docker run --rm odata-recorder:build /out/odata-recorder` | Image-Name und Pfad stehen im Bau (Nr. 8); der Aufruf von `docker` selbst läuft nicht (Nr. 14 e) |
   | K5 Zustandsfreiheit | nichts Geplantes, keine Chronik, keine Kennung der Planung | die Mengen A, B, C (Nr. 10 bis 12) und die Grenze in Nr. 14 a |
   | K6 Aufbau | Zielgruppe, Gliederung, Hinweise auf andere Kapitel des Handbuchs | kein Lauf; der Reviewer und die Abnahme lesen (Nr. 14 b) |

   Ein Verhalten des Binary steht im Handbuch nur als gebundenes Beispiel; die Prosa daneben sagt es nicht zusätzlich zu (Nr. 14 b).
   Host-Befehle stehen als Inline-Code, nie als umzäunter Block: jeder umzäunte Block ist ein gebundenes Beispiel (Nr. 5).
2. **Aufruf, Ausgabe, Ausgangswert.**
   - **a) Aufruf.** `pruefung handbuch <Binary> <Wurzel>`: genau zwei nicht leere Argumente, sonst der Text `Aufruf: pruefung handbuch <Binary> <Wurzel>` auf stderr und Exit 2. *Binary* ist der Pfad des gebauten Binary, *Wurzel* ein Verzeichnis, das die Dateien nach Nr. 15 a trägt.
   - **b) Ausgabe.** Ohne Befund: `pruefung handbuch: ok` auf stdout, Exit 0. Je Befund eine Zeile `pruefung handbuch: <Text>` auf stderr, Exit 1; alle Befunde stehen in der Ausgabe, kein Befund bricht ab. Je Eingabefehler eine Zeile `pruefung handbuch: Eingabe: <Text>` auf stderr, Exit 2. *Name* in einem Befund ist der Dateiname einer Beispiel-Datei mit der Endung `.txt` und ohne Verzeichnis (`aufruf-version.txt`), für den Lauf von `version` das Wort `version`.
   - **c) Eingabefehler** sind: `Binary fehlt: <Pfad>` (kein Eintrag oder keine reguläre Datei), `Handbuch fehlt: <Pfad>` (die Datei docs/user/benutzerhandbuch.md unter *Wurzel*), `Handbuch nicht lesbar: <Grund>` (Nr. 3 a), `Binary nicht startbar: <Fehler>` (Nr. 7 a) und `Arbeitsverzeichnis nicht anlegbar: <Fehler>` (das eigene Arbeitsverzeichnis, Nr. 2 e). Die Prüfung der Argumente, des Binary und des Handbuchs geht in dieser Reihenfolge dem ersten Start des Binary voraus; ein Eingabefehler vor dem Start startet das Binary nicht. Alle Eingabefehler dieser Prüfungen stehen in der Ausgabe, je einer in einer Zeile, in der Reihenfolge der Prüfung (`Binary fehlt`, dann `Handbuch fehlt` oder `Handbuch nicht lesbar`); fehlen Binary und Handbuch zugleich, sind es zwei Zeilen. Der erste Eingabefehler bricht die Prüfung der folgenden nicht ab, das Binary startet danach nicht. `Binary nicht startbar` und `Arbeitsverzeichnis nicht anlegbar` stehen darum nie neben einem anderen Eingabefehler. Ein Eingabefehler schließt jeden Befund aus (Exit 2, nur Eingabefehler in der Ausgabe).
   - **d) Reihenfolge.** Befunde stehen in der Reihenfolge der Nummern dieses Vertrags (3, 4, 5, 6, 7, 8, 9, 10, 11, 12), innerhalb einer Nummer nach Zeile des Handbuchs aufsteigend, dann nach Name der Datei in Bytefolge. Gleiche Eingabe gibt gleiche Ausgabe.
   - **e) Wirkung.** Die Prüfung liest *Wurzel* und startet das Binary; sie schreibt nur nach stdout und stderr und in ein eigenes Arbeitsverzeichnis, das sie am normalen Ende entfernt (Nr. 14 g).
3. **Lesen des Handbuchs.**
   - **a) Eingabe.** Gültiges UTF-8 ohne NUL und ohne BOM; sonst der Eingabefehler `Handbuch nicht lesbar: <Grund>` (`kein UTF-8`, `NUL`, `BOM`).
   - **b) Zeilen.** Zeilen trennt `\n`; ein `\r` unmittelbar davor gehört nicht zur Zeile (CRLF gilt wie LF), ein einzelnes `\r` ist ein Zeichen der Zeile. „Getrimmt“ heißt in diesem Vertrag: Leerraum an den Enden entfernt; Leerraum ist jedes Zeichen, das `unicode.IsSpace` trifft (so auch NBSP). „Leer“ in Nr. 4 a, 4 b und 5 a heißt: nach dem Trimmen ohne Zeichen.
   - **c) Zeichen.** Außer Tabulator ist jedes Zeichen der Kategorien Cc, Cf, Co, Cs, Mn und Me (Steuerzeichen, Formatzeichen wie Zero-Width und Soft-Hyphen, Privatnutzung, kombinierende Zeichen) Befund `Zeile <n>: Zeichen U+<XXXX> nicht zulässig`, je Zeile und Zeichen einmal. Grund: Solche Zeichen verbergen ein Wort vor den Mengen (Nr. 10 bis 12). Geschützte Leerzeichen (NBSP) und andere Zeichen der Kategorie Zs sind zulässig und werden beim Abgleich zu einem Leerzeichen (Nr. 13 a).
4. **Kopf.**
   - **a) Umfang.** Der Kopf sind die Zeilen vor der ersten Zeile, die mit `## ` beginnt (außerhalb von Zäunen, Nr. 5); gibt es keine, ist es das ganze Handbuch. Eine Zeile des Kopfes ist leer, die erste nicht leere Zeile ist die Überschrift der Ebene 1; beginnt sie nicht mit `# `, ist es Befund `Zeile <n>: Kopf: Überschrift fehlt` (*n* die Zeile der ersten nicht leeren Zeile, bei einem Kopf ohne nicht leere Zeile 1), und die Zeile wird nicht als Schlüssel gelesen (ein Schlüssel, der dort stand, fehlt dann zusätzlich nach Nr. 4 b). Der Text der Überschrift ist ungeprüft. Jede andere ist `<Schlüssel>: <Wert>`, getrennt am ersten `:`, beide Seiten getrimmt. Eine andere Zeile ist Befund `Zeile <n>: Kopf: Zeile ohne Schlüssel`, ein anderer Schlüssel als `Software-Version`, `Gültigkeitsbereich`, `Stand` Befund `Zeile <n>: Kopf: Schlüssel <k> unbekannt` (so auch `Version`, `Handbuch-Version`, `Autor`, `Änderungsdatum`).
   - **b) Schlüssel.** Jeder der drei Schlüssel steht genau einmal und mit nicht leerem Wert; je Schlüssel höchstens ein Befund, in der Rangfolge `Kopf: <Schlüssel> fehlt`, `Kopf: <Schlüssel> doppelt`, `Kopf: <Schlüssel> leer`.
   - **c) Software-Version.** Das Ziel startet `<Binary> version` (Nr. 7 a bis d) und erwartet Exit 0, leere Fehlerausgabe und die Standardausgabe `odata-recorder <V>` mit einem Zeilenumbruch, *V* nicht leer und ohne Leerraum. Der Wert im Kopf muss gleich *V* sein, sonst Befund `Software-Version <Kopf> weicht von der Ausgabe von version ab (<V>)`. Weicht die Ausgabe von der Form ab, ist es Befund `version: Ausgabe nicht lesbar: <Grund>`, und der Vergleich entfällt. *Grund* ist, in dieser Reihenfolge geprüft, `Signal`, `Exit <n>`, `Fehlerausgabe nicht leer` oder `Standardausgabe <q>` (*q* wie in Nr. 7 d, auf 200 Zeichen gekürzt). Ist der Wert im Kopf leer, fehlt der Schlüssel oder steht er doppelt, entfällt der Vergleich (es bleibt der Befund nach Nr. 4 b), der Lauf von `version` nicht.
   - **d) Stand.** `JJJJ-MM-TT`, vier Ziffern, Strich, zwei Ziffern, Strich, zwei Ziffern, ein Tag des Gregorianischen Kalenders (`2028-02-29` gilt, `2026-02-29` nicht); sonst Befund `Kopf: Stand ungültig: <Wert>`. Das Ziel vergleicht das Datum nicht mit der Uhr (Nr. 14 e).
   - **e) Gültigkeitsbereich.** Nicht leer; der Text ist ungeprüft.
5. **Zäune, Bindung, Dateien.**
   - **a) Zaun.** Ein Zaun öffnet eine Zeile, die ohne Einrückung aus genau drei Backticks besteht, gefolgt von einer Info-Zeichenkette ohne Backtick, und schließt die nächste Zeile, die (rechts getrimmt) aus genau drei Backticks besteht. Jede andere Zaunform (Tilde, vier oder mehr Backticks, eingerückt) ist kein Zaun: ihr Inhalt ist Prosa und fällt unter die Mengen. Ein Zaun ohne Schluss ist Befund `Zeile <n>: Zaun nicht geschlossen` und reicht bis zum Dateiende; seine Zeilen, die Öffnungszeile eingeschlossen, zählen als Prosa. Eine weitere Zeile aus drei Backticks mit Info-Zeichenkette darin öffnet keinen Zaun. Für ihn gilt nur dieser Befund: Info-Zeichenkette (Nr. 5 b) und Bindungszeile werden nicht geprüft. Folge: Ein vergessener Schluss macht den Rest zu Prosa, die Befunde danach gehen auf diesen einen zurück.
   - **b) Jeder Zaun ist ein Beispiel.** Seine Info-Zeichenkette ist leer (Nr. 5 a: drei Backticks, danach nichts oder nur Leerraum, sind ein Zaun mit leerer Info-Zeichenkette), sonst Befund `Zeile <n>: Info-Zeichenkette <x> nicht zulässig`, mit der getrimmten Info-Zeichenkette als *x*. Die Zeile unmittelbar davor (keine Leerzeile dazwischen) ist die Bindungszeile `Datei: [<P>](<P>)` mit demselben *P* im Text und im Ziel; *P* ist `beispiele/` und ein Name aus Kleinbuchstaben und Ziffern, durch einzelne Striche gegliedert, und `.txt`. Steht davor keine Zeile, die mit `Datei: [` beginnt, ist es Befund `Zeile <n>: Beispiel ohne Bindungszeile`; beginnt sie so und hat nicht diese Form, `Zeile <n>: Bindung ungültig: <Zeile>`; *n* ist die Zeile des Zaunes. Der Block gilt dann als ungebunden: sein Inhalt ist Prosa.
   - **c) Bindungszeile ohne Zaun.** Eine Zeile, die mit `Datei: [` beginnt und auf die kein öffnender Zaun folgt, ist Befund `Zeile <n>: Bindungszeile ohne Zaun`.
   - **d) Block gleich Datei.** Die Zeilen im Zaun, je mit einem `\n` abgeschlossen, sind bytegleich dem Inhalt der Datei *P* unter docs/user (*Wurzel*); sonst Befund `Zeile <n>: Beispiel weicht von <P> ab`. Fehlt die Datei, ist es Befund `Zeile <n>: <P> fehlt`, und der Block gilt als ungebunden. Eine Datei darf von mehreren Blöcken gebunden sein; jeder Block gleicht ihr.
   - **e) Gebunden** heißt: Bindungszeile gültig, Datei vorhanden und Block gleich Datei. Nur ein gebundener Block ist ein Zitat (Nr. 13 c).
   - **f) Verwaiste Einträge.** Jeder Eintrag des Verzeichnisses docs/user/beispiele (nicht rekursiv) ist eine Datei, deren Name die Form *P* hat, und wird von mindestens einer gültigen Bindungszeile genannt; sonst Befund `beispiele: <Name> keine Beispiel-Datei` (Unterverzeichnis, andere Endung oder Form) bzw. `beispiele: <Name> ohne Bindung`. Fehlt das Verzeichnis, ist das kein Befund.
6. **Beispiel-Datei.** UTF-8, ohne `\r`, endet mit einem `\n`. Aufbau, jede Zeile genau so:
   ```
   $ docker run --rm odata-recorder:build /out/odata-recorder [<Argument> …]
   --- Standardausgabe
   <Zeilen, null oder mehr>
   --- Fehlerausgabe
   <Zeilen, null oder mehr>
   --- Ausgangswert
   <Zahl>
   ```
   - **a) Kopfzeile.** Das Präfix `$ docker run --rm odata-recorder:build /out/odata-recorder` ist wörtlich; Argumente folgen, jedes mit genau einem Leerzeichen davor, nicht leer, aus `A-Z a-z 0-9 . _ = -` (keine Anführungszeichen, kein `$`, kein Leerraum); die Zeile endet ohne Leerraum.
   - **b) Abschnitte.** Die drei Trennzeilen stehen je genau einmal in dieser Reihenfolge; die erste folgt unmittelbar der Kopfzeile, eine Zeile dazwischen, auch eine leere, ist ein Verstoß. Die Zeilen zwischen den Trennzeilen sind der Inhalt des Abschnitts (leer erlaubt). Eine zweite Zeile, die einer Trennzeile gleicht, ist ein Verstoß. Ein Abschnitt, der eine Zeile gleich einer Trennzeile enthalten müsste, ist nicht ausdrückbar (Nr. 14 g).
   - **c) Ausgangswert.** Eine Zeile mit einer Dezimalzahl von 0 bis 255 ohne führende Null, nach ihr nichts als der Zeilenumbruch.
   - **d) Verstoß.** Jede Abweichung ist Befund `<Name>: Beispiel-Datei ungültig: <Grund>`, und die Datei wird nicht gefahren (Nr. 7). Je Datei steht höchstens ein Befund, der erste Verstoß in der Reihenfolge der Prüfung (Zeichensatz und Zeilenende, Kopfzeile, Trennzeilen, Ausgangswert); der Wortlaut von *Grund* gehört nicht zum Vertrag, wohl der Präfix `<Name>: Beispiel-Datei ungültig: `. Das Format kennt keine Umgebung, keine Eingabe, keine Datei und keinen Listener; wer ein Kapitel mit einer solchen Eingabe braucht, erweitert diese Nummer (Änderung dieses Vertrags).
7. **Lauf gegen das Binary.**
   - **a) Aufruf.** Je Datei, auf die eine gültige Bindungszeile zeigt und die vorhanden und formgültig ist, einmal (auch bei mehreren Blöcken), dazu einmal `version` (Nr. 4 c): `<Binary> <Argument> …` ohne Shell, die Argumente unverändert und in der Reihenfolge der Datei. Ein Binary, das sich nicht starten lässt (kein Exec, falsche Plattform, nicht ausführbar), ist der Eingabefehler `Binary nicht startbar: <Fehler>`.
   - **b) Umgebung.** Leere Umgebung, ein leeres Arbeitsverzeichnis des Laufs, Standardeingabe leer; stdout und stderr getrennt erfasst.
   - **c) Zeitgrenze.** 10 s je Lauf; danach wird der Prozess beendet, und es ist Befund `<Name>: Zeitgrenze überschritten` (für `version`: `version: Zeitgrenze überschritten`).
   - **d) Vergleich.** Standardausgabe, Fehlerausgabe und Ausgangswert werden je für sich gegen die Datei verglichen, bytegleich; je Abweichung ein Befund `<Name>: Standardausgabe weicht ab (erwartet <q>, erhalten <q>)`, ebenso `Fehlerausgabe`, `Ausgangswert` (erhalten `Signal`, wenn der Prozess durch ein Signal endete). *q* ist die Zeichenkette in Go-Schreibweise (`strconv.Quote`, mit den umschließenden Anführungszeichen), gekürzt auf ihre ersten 200 Zeichen (Unicode-Zeichen, nicht Bytes), ohne Auslassungszeichen; eine Zeichenkette von höchstens 200 Zeichen bleibt unverändert. Verglichen wird vor der Kürzung; die Kürzung kann in einer Escape-Folge enden. Bei `Ausgangswert` steht statt *q* die Dezimalzahl oder `Signal`, ohne Anführungszeichen.
8. **Aufrufpräfix gedeckt.** Gibt es mindestens einen Eintrag in docs/user/beispiele, gilt: Eine Datei `Makefile` oder `harness/mk/*.mk` unter *Wurzel* enthält die Zeichenfolge `-t odata-recorder:build`, und die `Dockerfile` unter *Wurzel* enthält `-o /out/odata-recorder`. Sonst Befund `Aufrufpräfix nicht gedeckt: Image` bzw. `Aufrufpräfix nicht gedeckt: Pfad` (eine fehlende Datei deckt nichts).
9. **Make-Ziele.**
   - **a) Zielliste.** Die Namen `^[a-z0-9-]+` vor dem `:` aller Zeilen der Form `^[a-z0-9-]+:.*##` in `Makefile`, `harness/mk/*.mk` (Dateien mit der Endung `.mk`, nicht rekursiv), `d-check.mk` und `repo.mk` unter *Wurzel*, soweit vorhanden; das ist die Regel von `make help`. Fehlt jede Datei, ist die Liste leer.
   - **b) Inline-Code.** Ein Inline-Code-Span (CommonMark: eine Folge von *n* Backticks öffnet, die nächste Folge von genau *n* Backticks im selben Absatz schließt) in Prosa, dessen Text mit `make ` beginnt, nennt als zweites Wort das Ziel. Das Ziel steht in der Zielliste, sonst Befund `Zeile <n>: make-Ziel <x> unbekannt`; beginnt das zweite Wort mit `-` oder enthält `=`, ist es Befund `Zeile <n>: make-Aufruf ohne Ziel an erster Stelle: <Span>`. Ein Span, der genau `make` ist, ist kein Aufruf. Alles ab dem dritten Wort ist ungeprüft.
10. **Menge A — Planung und Ausblick.** Aus dem Wortlaut („nichts Geplantes, keine Ausblicke“, die Beispiele „kommt in einer späteren Version“ und „folgt“, und das Verbot von „nicht verfügbar“ als Hinweis): Wörter des Plans, des Zeitverweises nach vorn und der Nicht-Verfügbarkeit. 16 Muster, 36 Formen (Schreibweise der Muster: Nr. 13 b). Jede Form trifft, Groß- und Kleinschreibung gleich:
    - A1 `geplant\p{L}*`: geplant, geplante, geplanten, Geplantes
    - A2 `planung\p{L}*`: Planung, in Planung, Planungen
    - A3 `vorgesehen\p{L}*`: vorgesehen, vorgesehene
    - A4 `(künftig|kuenftig)\p{L}*`: künftig, künftige, kuenftig
    - A5 `(zukünftig|zukuenftig)\p{L}*`: zukünftig, zukünftige, zukuenftig
    - A6 `zukunft\p{L}*`: Zukunft, in Zukunft
    - A7 `(spät|spaet)\p{L}*`: später, spätere, spaeter, spät
    - A8 `(demnächst|demnaechst)`: demnächst, demnaechst
    - A9 `bald\p{L}*`: bald
    - A10 `folgt`: folgt, wie folgt
    - A11 `noch nicht`: noch nicht
    - A12 `noch kein\p{L}*`: noch kein, noch keine, noch keinen
    - A13 `nicht (verfügbar|verfuegbar)`: nicht verfügbar, nicht verfuegbar
    - A14 `wird kommen`: wird kommen
    - A15 `ausblick\p{L}*`: Ausblick, Ausblicke
    - A16 `irgendwann`: irgendwann
11. **Menge B — Chronik.** Aus dem Wortlaut („keine Chronik“ mit „kein Änderungsverlauf“, „kein ‚neu in‘“, „kein ‚bisher‘“, „kein ‚seit Version‘“): Wörter der Vergangenheit, des Verlaufs und der Version als Zeitmarke. 14 Muster, 44 Formen:
    - B1 `bisher\p{L}*`: bisher, bisherige, bisherigen
    - B2 `(früher|frueher)\p{L}*`: früher, frühere, frueher
    - B3 `neu (in|ab|seit)`: neu in, neu ab, neu seit
    - B4 `seit (version|release|v\d)`: seit Version, seit Release, seit v2
    - B5 `ab (version|release|v\d)`: ab Version, ab Release, ab v2
    - B6 `nicht mehr`: nicht mehr
    - B7 `(änderungs|aenderungs)\p{L}*`: Änderungshistorie, Änderungsverlauf, Aenderungsdatum
    - B8 `changelog`: Changelog
    - B9 `release[ -]?notes?`: Release Notes, Release-Note, Releasenotes
    - B10 `\p{L}*verlauf\p{L}*`: Verlauf, Versionsverlauf
    - B11 `\p{L}*historie\p{L}*`: Historie, Versionshistorie
    - B12 `handbuch-version`: Handbuch-Version
    - B13 `wurden? (geändert|eingeführt|entfernt|ergänzt|hinzugefügt|ersetzt|umbenannt|behoben|korrigiert|verbessert)`: wurde geändert, wurde eingeführt, wurde entfernt, wurde ergänzt, wurde hinzugefügt, wurde ersetzt, wurde umbenannt, wurde behoben, wurde korrigiert, wurde verbessert, wurden geändert
    - B14 `(damals|ehemals|vormals|inzwischen|mittlerweile)`: damals, ehemals, vormals, inzwischen, mittlerweile
12. **Menge C — Kennungen und Pfade der Planung.** Aus dem Wortlaut („keine Bezüge auf Roadmap, Wellen, Slices, ADRs oder Kennungen der Spezifikation“) und dem ID-Schema des Repos (`harness/conventions.md`, MR-000): die Namen der Planung, die Kennungen, die das Schema vergibt, und die Pfade, unter denen Planung und Spezifikation liegen. 14 Muster, 22 Formen; die Menge gilt auch in Zitaten (Nr. 13 c):
    - C1 `roadmap\p{L}*`: Roadmap, Roadmaps
    - C2 `welle\p{L}*`: Welle, Wellen, welle-benutzerhandbuch
    - C3 `slice\p{L}*`: Slice, Slices, slice-handbuch-geruest
    - C4 `adr-([a-z]+-)?\d{4}`: adr-0001, adr-idx-0004
    - C5 `lh-[a-z]{2}(-[a-z]+)?-\d+`: lh-fa-01, lh-qa-sup-002, lh-rb-03
    - C6 `spec-\d{3}`: spec-110
    - C7 `arc-\d{3}`: arc-004
    - C8 `mr-\d{3}`: mr-001
    - C9 `co-\d{3}`: co-001
    - C10 `rc-\d{3}`: rc-001
    - C11 `beo-[a-z]\p{L}*`: beo-repo/abc
    - C12 `(docs/plan/)\p{L}*`: docs/plan/adr/
    - C13 `(spec/)\p{L}*`: spec/lastenheft.md
    - C14 `(harness/vertraege)\p{L}*`: harness/vertraege/
13. **Abgleich.**
    - **a) Text.** Der Abgleich läuft je Absatz der Prosa: ein Absatz ist eine maximale Folge nicht leerer Zeilen außerhalb gebundener Zäune. Die Zeilen werden mit einem Leerzeichen verbunden; dann werden die Zeichen `*`, `_` und Backtick entfernt, jedes Zeichen der Kategorien Zs, Zl, Zp und der Tabulator wird zu einem Leerzeichen, Folgen von Leerzeichen zu einem, und der Text wird mit `strings.ToLower` kleingeschrieben. Das Entfernen gilt wörtlich: `_geplant_` wird zu `geplant` und trifft A1, `noch_nicht` wird zu `nochnicht` und trifft A11 nicht.
    - **b) Muster.** Die Muster der Nr. 10 bis 12 sind reguläre Ausdrücke (RE2) auf diesem Text, in Kleinschreibung. Ein Treffer beginnt am Textanfang oder nach einem Zeichen, das kein Buchstabe ist, und endet am Textende oder vor einem Zeichen, das kein Buchstabe ist; `\p{L}*` am Ende des Musters nimmt die Beugung auf. Je Absatz und Muster höchstens ein Befund (das erste Vorkommen).
    - **c) Zitat.** Der Inhalt eines gebundenen Zaunes ist die Ausgabe des Binary und fällt nicht unter A und B: er ist bytegleich einer Datei, die das gebaute Binary erzeugt (Nr. 5 d, 7), und kann keinen Text tragen, den das Binary nicht schreibt. Die Menge C gilt auch dort: eine Kennung der Planung in der Ausgabe des Binary ist ein Befund am Binary, nicht am Handbuch. Der Inhalt eines ungebundenen Zaunes ist Prosa.
    - **d) Meldung.** `Zeile <n>: <Menge> <Muster>: <Treffer>`, mit *n* der ersten Zeile des Absatzes (im Zitat: der Zeile des Zaunes), *Menge* `Planung`, `Chronik` oder `Kennung`, *Muster* die Kennung des Musters (`A7`), *Treffer* der Text des Treffers im abgeglichenen Text. Treffer verschiedener Muster im selben Absatz sind je ein Befund, auch aus verschiedenen Mengen.
14. **Ungeprüft und akzeptierte Negative.** Je Punkt Grund und Folge im Betrieb.
    - **a) Ausblick und Chronik ohne Stichwort.** Futur und Ausblick in Worten außerhalb der Tabellen („Das Binary wird Fehler melden“, „eine Erweiterung ist denkbar“, „in einer der nächsten Fassungen“) und Chronik ohne Marke („die Ausgabe hat sich gewandelt“) bleiben grün. Grund: Das Urteil über Bedeutung ist keine Prüfung von Wörtern. Folge: Die Abnahme liest den Rest; ein Fund mit Muster erweitert die Tabelle dieses Vertrags (Änderung des Vertrags).
    - **b) Verhaltensaussage in Prosa.** Eine Zusage über das Binary in Prosa ohne gebundenes Beispiel („es braucht kein Netz“) wird nicht geprüft. Grund: Nur ein Lauf belegt eine Aussage über Verhalten, und Prosa lässt sich nicht in Aufrufe zerlegen. Folge: Der Reviewer und die Abnahme lesen jede Aussage gegen das Binary; die Regel (Nr. 1) lässt Verhalten nur als Beispiel zu.
    - **c) Wortformen außerhalb der Tabellen.** Andere Beugungen und Schreibungen, Trennstriche mitten im Wort, Tippfehler und andere Sprachen („planned“, „later“, „previously“). Grund: Das Handbuch ist deutsch; die Tabellen sind endlich. Folge: ein solcher Satz bleibt grün; ein Fund geht wie unter a an die Tabelle. Ebenso Bezeichner, in denen ein entferntes Zeichen (`*`, `_`, Backtick) mitten im Wort steht. Grund: Die Zeichen werden entfernt, damit hervorgehobene Wörter (`_geplant_`, `**geplant**`) treffen; ein Ersatz durch ein Leerzeichen würde Bezeichner trennen. Folge: `noch_nicht` bleibt grün, und ein Bezeichner, dessen Teile zusammen ein Wort ergeben, kann treffen (falscher Treffer, den der Autor im Wortlaut ändert).
    - **d) Das Ziel in Fließtext und Blöcken.** `make` im Fließtext (ohne Inline-Code), `make` hinter dem zweiten Wort eines Spans und Befehle in Ketten (`cd x && make y`) sind ungeprüft. Folge: ein Ziel, das es nicht gibt, steht dann ungeprüft im Text; die Regel in Nr. 1 verlangt den Inline-Code.
    - **e) Der Aufruf von `docker`, das Datum und die Zeit.** Das Ziel startet das Binary im Container der Stufe, nicht über `docker run` (Docker in Docker verlangte einen Socket und widerspräche dem Prüflauf ohne Mount); Nr. 8 deckt Image-Name und Pfad, nicht die Optionen, den Namen des Images in einer anderen Datei oder dessen Inhalt. Das Datum in `Stand` wird nicht gegen die Uhr oder das Alter des Handbuchs gehalten, weil die Uhr nicht zur Eingabe gehört. Je Lauf des Binary gilt die Zeitgrenze von 10 s, für den Rest des Ziels gibt es keine; die Länge der Ausgabe des Binary ist unbegrenzt. Folge: Ein veraltetes Datum, ein Fehler im Aufruf von `docker` oder eine endlose Ausgabe fällt dem Probelauf der Abnahme oder dem Reviewer auf.
    - **f) Unsichtbare Zeichen außerhalb der Klassen** von Nr. 3 c und Dateien außer dem Handbuch (die Beispiel-Dateien tragen die Form von Nr. 6, mehr nicht). Folge: Der Sensor für unsichtbare Zeichen, sobald er liegt, deckt sie.
    - **g) Plattform, Abbruch, Beispiel mit Trennzeile.** Das Binary der Stufe `build` ist für die Plattform des Bau-Hosts gebaut; ein Bau für eine fremde Plattform endet in `Binary nicht startbar` (Exit 2). Ein Signal an `pruefung handbuch` lässt das Arbeitsverzeichnis des Laufs zurück (der Container der Stufe wird verworfen). Ein Beispiel, dessen Ausgabe eine Zeile gleich einer Trennzeile enthielte, ist nicht ausdrückbar und endet als Abweichung. Folge: Exit 2 bzw. ein verwaistes Verzeichnis im verworfenen Container bzw. ein Befund an einer Stelle, die ein Mensch umschreibt.
    - **h) Verdrahtung.** Dass das `Dockerfile` die Stufe `handbuch` führt, das Ziel an `GATE_CHECKS` hängt, die Stufe ohne Netz läuft und `.dockerignore` die Dateien nach Nr. 15 durchlässt, fängt kein Fall. Folge: Fehlte die Verdrahtung, bliebe `make gates` grün, ohne dass das Handbuch geprüft würde. Beleg: Der Implementer fährt das Ziel vor der Übergabe auf einem Handbuch mit einem Wort aus A und nennt das Rot.
15. **Stufe, Wurzel, Verdrahtung.**
    - **a) Wurzel.** Die Stufe `handbuch` stellt unter ihrem Arbeitsverzeichnis genau diese Dateien bereit: docs/user (das Handbuch und docs/user/beispiele), `Makefile`, `d-check.mk`, `repo.mk`, `Dockerfile` und harness/mk; sie ruft `pruefung handbuch /out/odata-recorder <Arbeitsverzeichnis>`. `pruefung` und das Binary kommen aus der Stufe `build`.
    - **b) Image.** Das Ziel baut `--target handbuch` und taggt nichts Neues; der Image-Name im Aufrufpräfix (`odata-recorder:build`) gehört der Stufe `build` (`make build`, `SPEC-036`).
    - **c) Kontext.** `.dockerignore` ist eine Allowlist (`*`, dann `!<Pfad>`); die Stufe braucht sechs weitere Einträge: docs/user, `Makefile`, `d-check.mk`, `repo.mk`, `Dockerfile` und harness/mk. Keine andere Stufe kopiert diese Pfade (das `Dockerfile` kennt kein `COPY . .`), also ändert sich der Cache-Schlüssel der Stufen `deps` bis `lint` nicht; die Stufe `handbuch` läuft neu, wenn sich `src/`, `tools/pruefung`, das Handbuch, die Beispiele, die Makefiles oder das `Dockerfile` ändern. Die Übertragung des Kontexts wächst um diese Textdateien.
    - **d) Sensors.** Das Ziel braucht eine Zeile in `harness/README.md` §Sensors (Zellen höchstens 200 Zeichen, `.d-check.yml` Block `structure`); `.d-check.yml` ändert sich sonst nicht: das Handbuch trägt keine Kennung und keinen Pfad in Inline-Code, sodass `ids` und `codepaths` es ohne Ausnahme prüfen, und die Klasse `aussen` begrenzt nur Verweise der Spec-Straten auf das Handbuch.

**Gegenprobe:** Die Fälle H1 bis H10 fahren `pruefung handbuch` in Wegwerfverzeichnissen gegen ein Ersatz-Binary (ein Skript, das seine Argumente festhält und jeden Aufruf zählt) und ein Handbuch, das im Fall verändert wird:
- H1 sauberes Handbuch, Beispiel gebunden, Ersatz-Binary liefert die Datei: Exit 0, `ok`.
- H2, H3, H4 je ein Wort aus A (`geplant`), B (`bisher`), C (`spec-110`) in der Prosa: Exit 1, die Meldung nennt `Planung`, `Chronik` bzw. `Kennung`.
- H5 Software-Version im Kopf weicht von der Ausgabe von `version` ab: Exit 1.
- H6 das Ersatz-Binary weicht vom Beispiel ab: Exit 1, `Standardausgabe weicht ab`.
- H7 der Block weicht von der Datei ab: Exit 1, `Beispiel weicht von`.
- H8 `make` mit einem Ziel, das die Makefiles nicht führen: Exit 1.
- H9 das Binary fehlt: Exit 2 und das Ersatz-Binary wurde nicht gestartet.
- H10 ein Wort aus A im gebundenen Block: Exit 0 (das Zitat gilt nicht als Prosa).

`SPEC-102` nimmt die Fälle bei der Lieferung unter den dann nächsten Fallnummern auf; sie fangen die Verdrahtung in der Stufe und im Ziel nicht (Nr. 14 h).

**Stand:** geliefert: das Programm `pruefung handbuch` (`tools/pruefung/handbuch*.go`), die Stufe `handbuch` im `Dockerfile`, das Ziel `make handbuch-check` (`harness/mk/handbuch.mk`), die Zeile in `harness/README.md`, das Handbuch mit vier Beispiel-Dateien und die Einträge der Allowlist. Die Fälle H1 bis H10 der Gegenprobe liegen in `tools/gegenproben/handbuch.go` (in `SPEC-102` unter den vorläufigen Nummern 121 bis 130).

**Grenze:** Das Ziel belegt, dass das Handbuch einen Kopf mit der Software-Version des gebauten Binary trägt, dass jedes Beispiel gleich seiner Datei ist und gegen das gebaute Binary läuft, dass jedes Ziel von `make` in Inline-Code existiert und dass kein Wort der drei Mengen in der Prosa steht. Es beurteilt weder Verständlichkeit noch Vollständigkeit noch Ausblick ohne Stichwort (Nr. 14).
