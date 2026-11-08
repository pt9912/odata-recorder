# Review-Report: slice-admin-authentisierung — 2026-11-08

**Review-Art:** Code (gegen Plan, `SPEC-238`, `SPEC-239`, `SPEC-240`, `LH-FA-13.a`, Hard Rules `AGENTS.md` §3.2, §3.6, §3.10 bis §3.18)

**Gegenstand:** Zweig `slice-admin-authentisierung`, Commits `c4c06ed1` und `ddc1e5a9` gegen die Basis `c469ef1e`, Pfad `src/adapters/driving/admin/` (1502 Zeilen nach `git diff --numstat`, Marke 2600).

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs · **Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-08

**Eingangs-Kontext:**

- Plan `slice-admin-authentisierung` (Tabelle Zusage · Fänger mit 22 Zeilen, Raster in §6, Negative N-1 bis N-4, Stand der Umsetzung mit Mutationstabelle in §3)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md)
- `LH-FA-13.a`, `LH-QA-04.a`
- `AGENTS.md` (Hard Rules)

Vorgehen: Code und Tests gelesen, die Tabellen des Plans mit dem Wortlaut von `SPEC-238` Nr. 1 bis 5 und 13 abgeglichen, die Mutationslisten des Implementers als Anhalt genutzt und eigene Mutationen an Elementen gefahren, die keine Zeile der Tabelle des Implementers nennt (ein Container, `--memory=3g`, ohne Netz, Kopie des Baums, kein `make gates`). Nullprobe des Pakets (ohne `TestWettlaufNebenlaeufig`): grün.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Der Rest hinter dem Schema wird mit `strings.TrimLeft(…, " ")` gebildet, also vom Code als Menge der führenden Bytes behandelt. Kein Fänger führt für das erste Byte des Kandidaten mehr als den Tab (U1 Form 4, U3 Form 4); drei Mutationen der Menge bleiben grün: `" \xa0"` (Byte 0xA0, im Header zulässig), `" \x01"` und `" \r\n\v\f"`. Ein Token, das mit 0xA0 beginnt, würde mit dem richtigen Kandidaten abgewiesen (401), ohne dass ein Test rot wird. `TestKandidatAlleBytes` ändert je Stelle um eins; am ersten Byte erreicht es nur 0x00 und 0x01, und 0x01 bleibt unter der Mutation `" \x01"` ungefangen, weil der gekürzte Rest ebenfalls ungleich dem Token ist. Dieselbe Klasse am Ende des Rests: `TrimRight`/`TrimSuffix` mit Leerzeichen oder Tab bleibt grün (auf der Leitung nicht erreichbar, im direkten Aufruf von `ServeHTTP` schon; `SPEC-238` Nr. 4 b: „Byte für Byte unverändert“). | `AGENTS.md` §3.15, §3.17; `SPEC-238` Nr. 4 b; Skill-MEDIUM „Fänger führt nur eine Form der Eingabe“ | `src/adapters/driving/admin/admin.go` · `strings.TrimLeft(wert[len(schemaKlein):], " "), true`; Tests: `src/adapters/driving/admin/authentisierung_test.go` · `TestKandidatAlleBytes` | ja: Mutation des Zeichensatzes von `TrimLeft` auf `" \xa0"` im Paket, `go test` bleibt grün | Fänger deckt die Klasse „erstes und letztes Byte des Kandidaten“ nicht |
| F-2 | INFO | Drei Mutationen von `vergleiche` bleiben grün: `subtle.ConstantTimeCompare(…); return a == b` (Vergleich der Summen als Arrays, früh endend), Vergleich nur der ersten 31 und nur der letzten 31 Bytes der Summe. Die erste ist N-3 des Plans (kein verbotener Aufruf, Bezeichner `a` und `b`); die beiden anderen sind bis auf eine Kollision von SHA-256 äquivalent. | N-3 des Plans; `SPEC-238` *Ungeprüft* (Zeitverhalten) | `src/adapters/driving/admin/admin.go` · `return subtle.ConstantTimeCompare(a[:], b[:]) == 1` | nein (die Mutation ist durch N-3 angenommen) | akzeptiertes Negativ bestätigt |
| F-3 | INFO | Der Kommentar von `vergleiche` und `SPEC-238` Nr. 4 e sagen, die Laufzeit hänge „weder vom Inhalt noch von der Länge“ ab. Die Laufzeit von SHA-256 wächst mit der Länge des Kandidaten (in Blöcken von 64 Bytes); gleichlaufend ist der Vergleich der beiden Summen. Kein Test prüft die Zeit (`SPEC-238` *Ungeprüft*, Grund und Folge dort). Zuständig: Architect. | `AGENTS.md` §3.11 | `src/adapters/driving/admin/admin.go` · `damit weder Inhalt noch Länge die Laufzeit bestimmen` | nein | Zusage im Kommentar enger zu fassen als der Fänger |
| F-4 | INFO | `schreibeFehler` führt zwei Zeichenketten gleichen Typs hintereinander (`allow`, `www`); eine Vertauschung an einer Aufrufstelle übersetzt. Die Zellen von E2 und das Literal von U9 fangen sie heute. | Maintainability | `src/adapters/driving/admin/admin.go` · `func schreibeFehler(w http.ResponseWriter, code model.Code, status int, allow, www string)` | ja: `make test` (E2, U9) | Parameter gleichen Typs in Folge |

## Auflösung der Tabelle Zusage · Fänger und des Rasters

Je Zeile mindestens eine eigene Mutation (Name aus meinem Lauf, Ergebnis rot an dem Test, den der Plan nennt, sofern nicht anders vermerkt). Mutationen der Tabelle des Implementers habe ich nicht erneut gefahren und nicht als Beleg übernommen; die Zeilen stützen sich auf Lesen von Fänger und Code, soweit hier kein eigener Lauf steht.

| Zeile | Eigene Mutation | Ergebnis |
|---|---|---|
| A1 | zweiter Aufruf `KorrelationsID` im Zweig `/healthz` | rot: `TestKorrelationsIDJeRequestart`, `TestKorrelationsHeader` |
| A3, A4 | `/healthz` nur mit leerer Query; `RequestURI` mit Präfix statt Pfad | rot: `TestHealthzMethodeUndPfad` |
| A5, U7, U8 | Lesen gelesen am Fänger (Schleifen über Methoden × Pfade und die Fakes); keine eigene Mutation | Fänger geprüft, ohne Befund |
| U1 | Anzahl der Zeilen auf 1 bis 3 und 4 erlaubt; nur die erste Zeile gelesen (`Values(…)[:1]`) | rot: `TestFormenUngueltigerAuthorization`, `TestLeitungAnnahmenZuNetHTTP` |
| U1/U2 (Klasse Schema) | Schema-Schleife eine Stelle kürzer | rot: `TestSchemaJedesByte` |
| U3 | siehe F-1 | grün bei drei Mutationen des Zeichensatzes, Befund |
| U4, U5 | `a[:31]`, `a[1:]` | grün, äquivalent (F-2) |
| U6 | `return a == b` neben dem Aufruf | grün (N-3, F-2) |
| E2 | 405 mit `WWW-Authenticate: Bearer` | rot: `TestHeaderJeAntwortart` |
| E4 | `X-Correlation-Id` entfernt | rot: `TestKorrelationsHeader` |

Klassen nach `AGENTS.md` §3.17: A7 (8192 Schreibweisen × 8 Wertformen, 65536 und 384 Fälle) ist eine Schleife mit Zählprüfung und einem Orakel, das den Schlüssel klein schreibt; die Menge entspricht dem Wortlaut („keine Zeile von `Authorization`, in keiner Schreibweise, bei keiner Zahl und Form der Werte“), und `TestAnfrageFilterServeHTTP` führt dieselben 8192 Schlüssel über `ServeHTTP`. Schema-Bytes (1524 + 256) und Kandidat-Bytes (256 Stellen) sind Schleifen; die Klasse „Kandidat“ ist an den Rändern des Rests nicht geschlossen (F-1). U1 mit 20 Formen entspricht der Zählprüfung des Plans.

Raster in §6: jede Zelle trägt Nummer des Vertrags oder „gilt nicht“ mit Grund; die Zusatztabelle hat zu jedem Paar einen Fänger. Keine im Code entschiedene Zelle gefunden (die Lesart `URL.Path` für `/healthz` ist N-1 des Plans).

Aufrufstellen: `zaehlendeIDs` zählt alle fünf Operationen (A1 hält die Reihenfolge am Haltepunkt: Vergleich, Routen, Audit und Telemetrie stehen beim Aufruf auf 0); `vergleichFake` hält beide Argumente jedes Aufrufs als Kopie; `routeFake` hält `Anfrage` und `ctx`; `useCaseFake` hat zwei Operationen ohne Argumente und zählt beide (U8), es gibt kein Argument, das es festhalten könnte. Audit und Telemetrie zählen jede Operation (`keinePortAufrufe`).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driving/admin/admin.go` (Ablauf, Authentisierung, `schreibeErfolg`, `schreibeFehler`) | geprüft; Befunde F-1 (Fänger), F-3, F-4 |
| `src/adapters/driving/admin/*_test.go` (Authentisierung, Filter, Quelle, Leitung, Helfer) | geprüft; Befund F-1 |
| Hard Rules §3.2, §3.6 | geprüft, ohne Befund (keine Direktive `//nolint`, kein Gate gelockert, `GATE_CHECKS` unverändert) |
| Kommentare in `admin.go` und den Tests (§3.7, §3.11) | geprüft; Befund F-3, sonst ohne Befund |
| Plan §1, §3, §6 und Kopf gegen den Code (§3.9, §3.13) | geprüft, ohne Befund; die Abweichung zu U1 steht in §6 mit Ausgang und Adresse |
| Rohbytes (§3.18) | Fixtures mit BOM und allen 256 Bytes entstehen aus Bytewerten (`bom()`, `tokenVielfalt()`, Schleife über `byte(i)`); kein Treffer von Backslash-u in den Testdateien gesehen |
| Größenmarke (§3.14) | geprüft, ohne Befund (1502 unter 2600) |
| Geheimnisse (`SPEC-238` Nr. 13 a und b) | geprüft, ohne Befund; 401 trägt feste Literale, `TestKeinGeheimnisInDerAntwort` prüft Teilstücke ab 8 Bytes |

Die Mutation `E2-www-authenticate-bearer-an-jedem-fehler` des Implementers halte ich ebenfalls für äquivalent: Der Wert `www` ist an beiden Aufrufstellen leer oder `Bearer`. Nicht gefahren: `make gates`, Mutationsläufe der Zeilen K4, A6, U9 bis U11 und E3 (aus dem Lesen von Fänger und Code ohne Befund).

## Summary

HIGH 0 · MEDIUM 1 · LOW 0 · INFO 3.

MEDIUM F-1: der Zeichensatz des Zuschnitts am Anfang (und das Fehlen eines Falls am Ende) des Kandidaten ist nicht für das erste Byte als Klasse gedeckt. Keine Rückgabe an eine andere Rolle nötig außer dem Implementer; F-3 geht zur Kenntnis an den Architect.
