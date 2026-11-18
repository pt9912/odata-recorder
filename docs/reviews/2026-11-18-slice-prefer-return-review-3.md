# Review slice-prefer-return, Runde 3 (Fix 1c2379b9)

* Rolle: Reviewer, frischer Kontext. Gegenstand: `git show 1c2379b9` (Fix-Runde 2) gegen die Klasse (`AGENTS.md` §3.17), §3.11, §3.9.
* Vorrunden: `2026-11-18-slice-prefer-return-review.md`, `-review-2.md`.
* Bezug: LH-FA-07, ADR-0001.
* Messung: `sh tools/harness/mutationslauf.sh` (Image `odata-recorder:race`, netzlos, unter `flock` auf `.wt/heavy.lock`), 25 Mutationen, je ein Prozess je Stufe, ohne Detektor; Last des Hosts 34,3 bis 5,0 von 4 CPUs (Spalte `last`); Messender: der Reviewer. Die Liste liegt nicht im Repo.

## Findings

### MEDIUM

- kategorie: MEDIUM
- quelle: `AGENTS.md` §3.15, §3.17 (Fänger schmaler als Zusage)
- pfad: `src/hexagon/model/rueckgabe_test.go` · `TestLiesRueckgabeBackslashFremd` (zweite Schleife, `stellen := []string{`x`, `x=1`, `x=1;p=a`, `respond-async; p=`}`)
- befund: Die Zusage „ein Backslash außerhalb eines Paares hat keine Wirkung" (`SPEC-285` Nr. 2; Plan Zeile 6b) ist nur für Backslashes geführt, die hinter einem Lauf ohne Paar stehen. Der Backslash unmittelbar hinter dem Schlusszeichen eines Paares ist eine eigene Stelle im Code (hinter der Schleife über das Paar in `elementEnde`) und bleibt ungefangen. Mutation B1: nach dem schließenden `i++` des Paares `if i < len(z) && z[i] == '\\' { i++ }`. Dann liest `x="a"\, return=minimal` als Keine statt Minimal. Ergebnis: `gruen` auch in Stufe 2 (ganzes Paket model), `TestLiesRueckgabeBackslashFremd` und `TestLiesRueckgabeFolgen` bleiben grün (die Folgen reichen mit Länge 6 nicht an `""\,return=minimal`).
- verifizierbar: ja (Mutation oben)
- klasse: Klasse „Stelle der Grammatik" nach Befund-Beispielen statt aus dem Wortlaut gebildet (§3.17)

### LOW

- kategorie: LOW
- quelle: `AGENTS.md` §3.11 (Kommentar nennt nur, was die Prüfung fängt)
- pfad: `src/hexagon/model/rueckgabe_test.go` · `TestLiesRueckgabeBackslashFremd`: „Stelle der Folge (Name, Wert, Rest, Rest eines anderen Elements)"
- befund: Der Kopf nennt die Stellen außerhalb eines Paares als Menge; der Backslash am Anfang eines Elements (Mutation B2: `\` am Elementanfang überspringt zwei Bytes) ist keine Stelle des benannten Tests. B2 wird nur von `TestLiesRueckgabeFolgen` rot (`"\\=return=minimal": 1, Orakel 0`), nicht vom benannten Test. Die Klasse ist im Paket gedeckt, die Zuordnung Zeile 6b zu Fänger trifft für diese Stelle nicht zu.
- verifizierbar: ja
- klasse: Fänger-Zuordnung im Plan breiter als der benannte Test

### INFO

- INFO-1, N-1: Der Kommentar in `src/hexagon/services/mock/prefer_test.go` („Jeder Gegenstand außer `Sonstiges` wird von mindestens einem Fehler der Liste von `Sonstiges` getrennt; die Schleife prüft es je Gegenstand") steht im Indikativ und stimmt mit der Schleife darunter überein. Behoben.
- INFO-2, N-3 (Klasse „Stelle des langen Laufs"): Quadratische Mutationen an Stellen, die weder Befund noch Implementer nannten, sind rot: QL0 (Leerraum am Elementanfang, 2/2), QL1 (Leerraum hinter dem Namen von `return`, 2/2), QK (Komma, 2/2), QZ (Zeilen, quadratisch in der Zahl der Zeilen, 2/2), QR (jede Stelle in `elementEnde`, 2/2), QW (Wert als Token, 2/2), QG (Zeichen im Wort, 2/2). Je Prozess ein Lauf; ohne Fehler im Code (Mutation Z0, nur Kommentar) 0 von 6 Prozessen rot bei Last 21,5 bis 28,8 von 4 CPUs, also stärker belastet als die Messung des Plans (4,6 bis 75, Z0 0 von 48). Die Raten im Plan §3 (8 von 8, Z0 0 von 48) sind plausibel; eine Einzelzahl bleibt Größenordnung.
- INFO-3, Baustein je Stelle (im Plan als akzeptiertes Negativ mit Grund und Folge geführt, `SPEC-285` *Ungeprüft*): grün bleiben, wie dort zugesagt, ein Lauf von `;` (QU), von Leerzeichen (QS) innerhalb eines fremden Elements und viele geschlossene Paare hintereinander (QX1). Das ist die genannte Grenze; kein Befund, nur Beleg, dass sie genau so verläuft.
- INFO-4, N-2 (Klasse „Backslash im Paar eines fremden Elements"): Die Zählung 384 und 64 im Kopf stimmt mit den Schleifen überein; die Matrix führt Parität, Folgenlänge 0 bis 7, Zeichen hinter der Folge und offenes Paar am Zeilenende. Der Fänger deckt sie, nur die Stelle hinter dem Paar fehlt (MEDIUM oben).
- INFO-5, §3.9: Zeile 6b, Zeile 9a, Prüfgrenzen, Mutationstabelle und Fix-Runde-2-Tabelle in §3 und die Aufzählung der 20 Stellen in `SPEC-285` *Ungeprüft*, im Plan und im Kopf von `LiesRueckgabe` stimmen miteinander überein (20 Stellen gezählt). Wird B1 behoben, ziehen Zeile 6b (Stelle, Zählung 64 → neue Zahl), die Zählungen im Plan §3 und der Kopf des Tests mit.

## Geprüft, ohne Befund

- `src/hexagon/services/mock/prefer_test.go` (Kommentar N-1): geprüft, ohne Befund.
- `spec/spezifikation.md` SPEC-285 *Ungeprüft*: geprüft, ohne Befund (Aufzählung gleich der Liste in `zeitFormen`).
- `src/hexagon/model/rueckgabe.go` (Kopfkommentar): geprüft, ohne Befund.
- `docs/plan/planning/in-progress/slice-prefer-return.md` §3: geprüft, ohne Befund außer Zuordnung Zeile 6b (LOW).

## Verdikt

Nicht freigegeben bis zur Behebung des MEDIUM: Der Fix deckt die Klasse „Backslash außerhalb eines Paares" nicht in jeder Stelle (hinter dem Schlusszeichen eines Paares). N-1 und N-3 sind behoben, N-2 zu einem Teil. Kein HIGH.
