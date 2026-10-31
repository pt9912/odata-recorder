# SPEC-040 — Doku-Gate (`make docs-check`)

**Werkzeug:** `make docs-check`: d-check im Image des Pins (`d-check.mk`, Digest), Konfiguration `.d-check.yml`.
Das Ziel hängt an `GATE_CHECKS`; `make gates` fährt es.

**Festlegung:** Kennungen, Links und Dokumentkonventionen prüfen, dazu die Größe der Guide-Dateien (Nr. 1 bis 9).
Rot ist das Ziel für die Größe, wenn eine Datei der Prüfmenge (Nr. 2) mehr Zeilen oder mehr Bytes trägt, als ihre Regel erlaubt (Nr. 5), wenn eine Regel keine Datei trifft (Nr. 6 a) oder wenn ein Abschnitt `### 3.<n>` von `AGENTS.md` mehr Zeilen trägt, als erlaubt sind (Nr. 9).

1. **Ort und Aufbau.**
   - **a)** Die Größenprüfung sind die Module `file` und `structure` von d-check, aktiv in `modules:` von `.d-check.yml`, mit dem Block `file:` (Nr. 2 bis 8) und einem Eintrag in `structure:` (Nr. 9). Es gibt kein eigenes Ziel, kein Skript und kein eigenes Programm.
   - **b)** Grund: Ziel, Image und Mount (netzlos, `:ro`) bestehen, das Modul zählt roh und erreicht die Dateien (Nr. 6 d); ein Programm nach dem Muster von `zeichen-check` brächte Stufe, Treiber und Vertrag für dieselbe Zählung. Die Bindung ist [ADR-0010](../../docs/plan/adr/0010-groesse-der-guide-dateien-als-gate.md).
   - **c)** Das Modul ist hermetisch: kein `git`, kein Netz.
2. **Prüfmenge.** Fünf Regeln in `.d-check.yml`; jede gilt für jede Datei ihres Globs zur Zeit des Laufs, auch für eine Datei, die es beim Schreiben der Regel nicht gab (Nr. 6 e).
   - `AGENTS.md`: das Briefing, das jede Sitzung lädt (Anlass des Gates).
   - `.claude/commands/*.md`: die Anweisung je Workflow-Command; jeder Lauf des Commands lädt seine Datei.
   - `.claude/agents/*.md`: die Rollen-Typen; jeder Lauf der Rolle lädt ihren Typ.
   - `.harness/skills/*.md`: die Urteilsgrundlage von Reviewer und Verifier.
   - `harness/conventions.md`: jede Sitzung lädt sie; sie wächst nicht mit dem Gate-Index (Nr. 10 d).
   Was nicht zur Menge gehört, steht in Nr. 10 d.
3. **Metrik.** Je Regel zwei Maße, beide gesetzt. **Bytes** sind die rohe Länge der Datei (Fenced- und Inline-Code zählen mit), nicht Zeichen: ein Umlaut zählt 2. **Zeilen** sind die Zahl der Zeilenvorschübe, plus 1, wenn die Datei nicht auf einen Zeilenvorschub endet; die leere Datei hat 0. Eine Regel mit nur einem Maß ist nicht zulässig (Nr. 4 a).
4. **Grenzen.**
   - **a)** Die Grenze einer Regel ist die Messung am Stand ihrer Aufnahme oder ihrer letzten Senkung: der Ist-Stand, kein Spielraum, beide Maße. Eine Regel mit Glob über mehrere Dateien steht auf dem Maximum ihrer Dateien, je Maß getrennt.
   - **b)** Die Werte und ihre Messung (Datei, Zeilen, Bytes, Stand) stehen einmal, im Kommentar von `.d-check.yml` neben dem Block; dieser Vertrag nennt keine Grenze.
   - **c)** Gleich der Grenze ist grün, eins darüber rot (Nr. 5).
   - **d)** Eine Regel trägt kein `hint`: er ersetzt die Meldung und mit ihr die Zahlen (Nr. 5).
5. **Überschreitung.** Je Datei, Regel und Maß ein Befund `file-lines-exceeded` bzw. `file-bytes-exceeded` mit der Datei als Fundstelle (Zeile 1) und der Meldung `Datei hat <N> Zeilen, erlaubt sind <M>` bzw. `Datei hat <N> Bytes, erlaubt sind <M>`. Das Ziel endet mit Exit 1.
   - Überschreitet eine Datei beide Maße, sind es zwei Befunde, ohne Vorrang.
   - Überschreiten mehrere Dateien einer Regel, ist es ein Befund je Datei und Maß.
   - Trifft eine Datei zwei Regeln, gilt jede: ein Befund je Regel (Nr. 10 b).
6. **Randformen der Menge.**
   - **a)** Eine Regel, die keine Datei trifft, ist der Befund `file-no-match` auf ihrem Glob, Exit 1: Datei gelöscht, umbenannt, Glob falsch, Datei nicht lesbar. Eine Regel zu setzen heißt zu behaupten, dass sie Dateien trifft.
   - **b)** Die leere Datei hat 0 Zeilen und 0 Bytes und ist grün.
   - **c)** Eine Datei ohne Zeilenvorschub am Ende zählt ihre letzte Zeile mit.
   - **d)** Dateien in versteckten Verzeichnissen (`.claude/`, `.harness/skills/`) gehören zur Menge; `scan.ignore` (`.harness/**`) gilt für das Modul nicht. Belegt am Pin am 2026-10-31.
   - **e)** Eine neue Datei in einem Glob wird gegen die Grenze der Regel gemessen.
   - **f)** Ein Symlink wird nicht gefolgt (Nr. 10 c).
   - **g)** Name und Endung haben keine Wirkung; gemessen wird jede Datei, die der Glob trifft.
7. **Konfiguration.** d-check weist am Pin ab, Exit 2 ohne Lauf: eine Regel ohne beide Maße, ein negatives Maß, zwei Regeln mit gleichem Glob, einen unbekannten Schlüssel. Ein Exit 2 ist rot, nie grün. Die Grenze 0 ist zulässig und heißt: nur die leere Datei ist grün.
8. **Wandern der Grenze.**
   - **a) Senken** ist frei: ein Commit, der die Grenze auf die neue Messung setzt und sie im Kommentar nachträgt. Es muss nicht im Commit der Verschlankung stehen. Eine Grenze unter dem Ist-Stand färbt das Ziel an der eigenen Datei rot.
   - **b) Anheben** ist eine Lockerung (`AGENTS.md` §3.6). Zulässig ohne neue ADR ist allein die **Verschiebung**: ein Commit, der eine Grenze hebt und im selben Commit mindestens eine andere Grenze desselben Maßes (Zeilen oder Bytes, je getrennt) um mindestens denselben Betrag senkt, sodass die Summe der Grenzen aller Regeln dieses Maßes nicht steigt; Messung beider Dateien und Begründung stehen im Kommentar und in der Commit-Nachricht. Jede andere Anhebung ist eine eigene ADR, die [ADR-0010](../../docs/plan/adr/0010-groesse-der-guide-dateien-als-gate.md) nennt.
   - **c) Aufnahme** einer weiteren Regel (Datei oder Glob) ist eine Verschärfung und frei, mit der Messung als Grenze. Eine neue Datei in einem bestehenden Glob erbt dessen Grenze (Nr. 6 e); passt sie nicht darunter, gilt Nr. 8 b.
   - **d) Folge für Änderungen** an einer Datei der Menge: Wer sie über die Grenze bringt, macht `make docs-check` rot; die Änderung braucht eine Verschiebung (Nr. 8 b) oder eine ADR. Das gilt auch für den Diff-Vorschlag, der `AGENTS.md` um eine Zeile verlängert.
   - **e)** Nr. 8 b und die Aufnahme nach Nr. 8 c prüft kein Gate (Nr. 10 a); sie prüft der Reviewer am Diff von `.d-check.yml`.
9. **Hard Rules.** Ein Eintrag in `structure:` begrenzt jeden Abschnitt `### 3.<n>` von `AGENTS.md` (`section-pattern: '^### 3\.\d+ '`, `sections: each`) auf `max-lines` Zeilen. Gezählt werden die Zeilenvorschübe des bereinigten Abschnittstexts (Überschrift, Leerzeilen und Text; Fenced-Code zählt nicht): Überschrift, eine Leerzeile und *k* Textzeilen zählen *k* + 2; eine Leerzeile am Ende des Abschnitts zählt nicht. Die Grenze steht auf dem Maximum der Abschnitte (Nr. 4 a, b). Der Befund ist `section-lines-exceeded` auf der Überschriftszeile, mit der Zahl. Trifft das Muster keinen Abschnitt, ist der Befund `section-missing`. Anheben ist eine ADR; die Verschiebung nach Nr. 8 b gilt nicht, weil es nur einen Wert gibt. Zeilen sind das Maß, nicht Bytes: das Werkzeug zählt Bytes nur je Datei (Nr. 10 e).
10. **Ungeprüft und akzeptierte Negative.** Je Punkt Grund und Folge im Betrieb.
    - **a) Die Höhe der Grenze und die Regel zum Anheben.** Grund: "Grenze gleich Ist" als Prüfung zwänge das Senken in den Commit jeder Kürzung und widerspräche Nr. 8 a. Folge: Eine zu hoch gesetzte oder nach Nr. 8 b nicht gedeckte Anhebung passiert das Ziel; sie fangen der Reviewer am Diff von `.d-check.yml` und, nur grob, Fall 167 der Gegenproben (eine Grenze über 50 000 Zeilen oder 800 000 Bytes macht ihn rot).
    - **b) Eine Datei in zwei Regeln.** Keine Regel dieser Konfiguration überlappt eine andere. Am Pin gelten beide, und die strengere färbt das Ziel (belegt am 2026-10-31). Kein Fänger, weil keine Regel ihn auslöst. Folge: Ein künftiger überlappender Glob ändert nichts an Nr. 5.
    - **c) Symlinks.** Eine Guide-Datei, die nur über einen Symlink erreichbar ist, wird nicht gemessen. Bestand: `.claude/rules/*` sind fünf Symlinks auf `AGENTS.md`, `harness/conventions.md` (beide in der Menge), `harness/README.md` (Nr. 10 d) und zwei Dateien der vendorten Baseline.
    - **d) Nicht in der Menge, mit Grund und Folge.** `harness/README.md`: jede Sitzung lädt sie, aber der Gate-Index wächst um eine Zeile je Gate (`AGENTS.md` §4); die Zellen begrenzt `structure` auf 200 Zeichen, die Datei wächst also um rund eine Zeile je Gate. Eine Grenze auf dem Ist-Stand färbte jeden Slice rot, der ein Gate liefert; Folge: die Datei wächst unbegrenzt, langsam. `.harness/baseline/**`: Kurs-Inhalt, byte-genau mitgeliefert; zwei Dateien daraus lädt jede Sitzung über `.claude/rules/`, das Repo kann sie nicht kürzen. Pläne, ADRs, Berichte, Spezifikationen, Lastenheft: sie wachsen mit dem Gegenstand und werden nicht von jedem Lauf geladen. `.claude/hooks/`, `.claude/settings.json`, `README.md`: keine Anweisung an das Modell oder kein Guide.
    - **e) Form der Hard-Rule-Grenze.** Ein Abschnitt, der nicht `### 3.<n> ` heißt, und Bytes je Abschnitt. Zeilen sind ein grobes Maß: am 2026-10-31 hat §3.10 19 Zeilen und 2512 Zeichen, §3.7 29 Zeilen. Folge: Eine Regel aus wenigen sehr langen Zeilen passiert die Zeilengrenze.
    - **f) Zeichen und Token als Einheit.** Gemessen werden Bytes (Nr. 3); am 2026-10-31 hat `AGENTS.md` 24648 Bytes und 24193 Zeichen. Folge: Eine Kürzung durch Umlaute-freie Wörter spart Zeichen, nicht in dem Verhältnis Bytes.
    - **g) Verhalten von d-check am Pin außerhalb der Fälle:** Exit 2 bei Konfigurationsfehlern (Nr. 7), Determinismus eines zweiten Laufs. Grund: Eigenschaft des Werkzeugs, nicht Zusage dieses Repos. Folge: ein Pin-Wechsel kann sie ändern; er ändert die Prüfsumme der Gegenproben (Pin in den Eingaben), und die Fälle 156 bis 165 fangen die Randformen, die sie führen.
    - **h) Verdrahtung der Gegenprobe.** Dass die Stufe `gegenproben` das Binary aus dem Image des Pins kopiert und `.d-check.yml` in den Eingaben der Prüfsumme steht, fängt kein Fall; fehlt das Binary, sind alle Fälle rot. Folge: Fehlte `.d-check.yml` in `EINGABEN`, liefe die Gegenprobe nach einer Änderung der Grenzen nicht neu. Beleg: der Implementer ändert `.d-check.yml` und zeigt, dass `bash tools/harness/gegenproben-eingabe.sh` eine andere Prüfsumme ausgibt.
11. **Gegenprobe.** Die Fälle 156 bis 170 von `SPEC-102` rufen das Binary `d-check` im Wegwerfverzeichnis: die Fälle 156 bis 165 mit einer Konfiguration des Falls (die Wirkung des Moduls am Pin, mit kleinen Grenzen), die Fälle 166 bis 170 mit den Blöcken `scan`, `structure` und `file` der echten `.d-check.yml` und kleinen Dateien unter den Pfaden der Menge (die Konfiguration bindet jede Datei).

**Stand:** geliefert sind die Module `file` und `structure` in `.d-check.yml` für die Prüfmenge nach Nr. 2, die Zeile von `make docs-check` in `harness/README.md` und die Fälle 156 bis 170 von `SPEC-102`.

**Grenze:** Prüft Verweise, die Richtung der Spec-Straten und die Größe der Guide-Dateien, nicht die inhaltliche Richtigkeit einer Aussage und nicht den Inhalt der Guide-Dateien (Duplikate zwischen Briefing, Commands und Skills). Was die Größenprüfung nicht sieht, steht in Nr. 10.
