# SPEC-101 — Lint-Profil (`make lint-profil`, `make lint`)

**Werkzeug:** `make lint-profil` und `make lint`.

**Festlegung:** Lint-Profil `.golangci.yml`; Unterdrückungen nur im Profil mit `# Why:`.

**Festlegung `make lint-profil`:** Prüft das Profil mit `golangci-lint config verify` gegen das Schema von golangci-lint v2; kein Code.

**Festlegung `make lint`:** Baut die Stufe `lint` des `Dockerfile` (golangci-lint-Image per Digest aus `harness/mk/lint-profil.mk`, ohne Netz) und führt `tools/harness/lint.sh` aus.
`lint.sh` läuft an der Wurzel des Moduls und hat zwei Teile, die beide laufen, auch wenn der erste einen Befund meldet:
1. Eigene Prüfungen (siehe unten).
2. `golangci-lint run -c .golangci.yml ./...`.
Das Ziel schlägt fehl, wenn einer der beiden Teile einen Befund meldet.

**Eigene Prüfungen** (Befund je Verstoß, Form `lint.sh: <Datei>:<Zeile>: <Text>`):
1. Keine Direktive `//nolint`, `//lint:ignore` oder eine dieser Direktiven mit Leerzeichen nach `//` in einer Datei `*.go`.
   Gesucht wird der Text, an jeder Stelle einer Zeile, auch in Testdateien und in Zeichenketten.
2. Jede Regel unter `linters.exclusions.rules` im Profil steht unmittelbar unter einem Kommentarblock, dessen erste Zeile mit `# Why:` beginnt und danach Text trägt.
   Eine Leerzeile zwischen Block und Regel löst den Block von der Regel.
3. Das Profil setzt `linters.exclusions.warn-unused: true`.
   Meldet golangci-lint daraufhin eine Regel ohne Wirkung, steht die Warnung in der Ausgabe und führt nicht zum Fehlschlag: eine Regel darf dem Code vorauseilen.
4. Es gibt mindestens eine Datei `*.go`.

**Randformen:**
1. `.golangci.yml` fehlt oder ist nicht lesbar: Meldung `lint.sh: Profil .golangci.yml fehlt oder ist nicht lesbar`, Exit-Code 2; golangci-lint läuft nicht.
2. Das Profil weicht von der Form ab, die `lint.sh` liest: `linters.exclusions` fehlt, trägt kein `warn-unused: true`, oder eine Regel unter `rules` beginnt nicht mit `      - ` (sechs Leerzeichen): Befund der Prüfung 3 bzw. 2, Exit-Code 1.
   Die gelesene Form steht im Kopfkommentar des Profils.
3. `exclusions.rules` fehlt oder ist `[]`: kein Befund der Prüfung 2.
4. Ein Verzeichnis ohne Go-Dateien: kein Befund, golangci-lint übergeht es. Fehlt jede Go-Datei im Modul, meldet Prüfung 4 einen Befund.
5. Eine Direktive in `*_test.go`: derselbe Befund wie im Produktcode.
6. Abbruch und Herunterfahren: `lint.sh` hält keinen Zustand; ein Abbruch des Builds lässt nichts zurück.
7. Exit-Codes: 0 kein Befund, 1 mindestens ein Befund der eigenen Prüfungen oder von golangci-lint, 2 Profil fehlt oder ist nicht lesbar.

**Stand:** vorhanden für die Schema-Prüfung des Profils und die Prüfung des Codes.

**Grenze:** Die Schwellen des Profils (Komplexität 15, Funktionslänge 100) beißen am heutigen Skelett nicht.
Die Gegenprobe zu `make lint` fehlt; sie liefert `SPEC-102`.
