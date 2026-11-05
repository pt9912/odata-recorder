# Verifikationsbericht: slice-v2-query-filter (Teil B) — 2026-11-05

**Rolle:** Verifier (Modul 11) · **Frage:** Bauen wir es richtig? — gegen Plan und DoD.

**Gegenstand:** Zweig `slice-v2-query-filter`, HEAD `35c7aa8`, Basis des Blocks `80d7910c`. Bezug: `LH-FA-17`, `LH-QA-05`, [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md).

**Stand dieses Berichts: ZWISCHENSTAND** (Sitzung unterbrochen). Die Mutationen gegen die Klasse und die Messung der Zeilen 32/33 stehen noch aus und werden in einem Folge-Commit ergänzt.

## Bisher selbst gemessen

| Größe | Wert |
|---|---|
| `make gates` auf `35c7aa8` | Exit 0; a-check 0 Befunde; d-check `607 Datei(en), 0 Befund(e)`; baseline-verify OK; migration-check ok; zeichen `1198 Datei(en)`; kein Datei-Diff danach |
| `make groessenmarke` | `slice-v2-query-filter.md: ok, Messung 2546 von Marke 4600` |
| `git diff --numstat 80d7910c..HEAD` über `urlparserv2/` | 2506 Einfügungen, 40 Löschungen = 2546 (gleich Plan §4) |
| Grundlinie der Tests (Wegwerfkopie, `go test -run TestFl`) | grün |

## Plan-vs-Code (bisher)

| Plan sagt | Code sagt | Urteil |
|---|---|---|
| Kein V4-Import ([ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md)) | `filter.go` importiert nur `model`; `filter_token.go` nur `strings`; a-check 0 Befunde | konform |
| Tabelle `Zusage · Fänger` 37 Zeilen | Zeilen 1 bis 37 vorhanden | konform |
| `SPEC-173` Nr. 5, F-5 berichtigt | Satz „lesbar ist“ ersetzt durch „der Form nach beginnt“, Historie-Zeile vom 2026-11-05; Code folgt der Form | konform |
| F-8: Kopf nennt `SPEC-164` Nr. 1 bis 4 und 7 bis 9, Nr. 6 mit Zuordnung | Kopf und `SPEC-173` Nr. 5 *Der Leser* nennen dieselbe Menge | konform |
| `SPEC-164` Nr. 1 bis 4, 7 bis 9 | Code gegen den Wortlaut gelesen (Kodierung vor Grammatik, Lexik, Rangfolge, Schlüsselwörter, Operanden, Tiefe 32, Literal 512, Zustandslosigkeit); keine Abweichung gefunden | konform |
| Produktcode kein `nolint` | 0 Treffer | konform |

## Rohbytes (AGENTS.md §3.18)

Gelesen mit `LC_ALL=C grep -P`: alle 18 Dateien von `80d7910c..HEAD` (Plan, Roadmap, Welle, Review, `spec/spezifikation.md`, `filter.go`, `filter_token.go`, 9 `fl_*_test.go`, `literale_test.go`). Folge Backslash-`u`-Hex: 0 im Diff (die 5 Treffer in `spec/spezifikation.md` sind Bestand, 0 davon im Diff), BOM 0, NBSP 0, Nullbreite/Zeilentrenner 0, Steuerzeichen 0, ungültiges UTF-8 0, jede Datei endet mit Zeilenvorschub, `git diff --check` leer. Nicht-ASCII nur `§ ² µ · Ä Ö × Ü ß ä é ö ÷ ü`. Der Bericht selbst wird nach dem Schreiben gegengelesen.

## Offen (Folge-Commit)

Mutationen F-1 (Zeichenklassen vor `(`, Rang Aufruf gegen Namensprüfung), F-2 (V4-Form je Position, Länge, Trenner, Zeichenart), F-3 (Schreibweisen `inf`/`nan`/`-INF`), F-4 (Nicht-ASCII-Faltung), Zeile 18 (Gleichheitsprobe), Zeilen 32/33 mit 30 Prozessen im Image `race`, Probe-Pakete über die Klassen.
