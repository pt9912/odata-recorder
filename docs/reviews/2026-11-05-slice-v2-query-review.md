# Review-Report: slice-v2-query (Teil A) — 2026-11-05

**Review-Art:** Code (gegen Plan, ADR und Hard Rules)

**Gegenstand:** `slice-v2-query`, Zweig `slice-v2-query`, `ab7b308..f4651c9` (Basis des Blocks `ab7b3083`; Merge-Basis zu `main` ohne die neueren Commits von `main`). Gemessen 3973 Zeilen (Einfügungen plus Löschungen, alle Pfade), 3721 über die Pfade des Blocks; Marke 7000.

**Skill:** `.harness/skills/reviewer.md` @ `f4651c9`
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-05

**Eingangs-Kontext:**

- `slice-v2-query` (Plan, §1, §3 mit 49 Zeilen, §6 Raster)
- [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) (versionierter Adapter für OData V2), `Accepted`
- `LH-FA-17`, `LH-QA-05`, `LH-RB-01`
- `SPEC-160` bis `SPEC-162`, `SPEC-173`, `SPEC-174`
- `AGENTS.md` (Hard Rules §3.7, §3.9 bis §3.18)

Lauf: `make gates` im Worktree grün (Exit 0, kein Datei-Diff danach). Zusätzlich 21 eigene Mutationen auf einer Wegwerfkopie (Image `odata-recorder:test`, `--memory=3g`), davon 2 Sed-Fehlgriffe und 1 äquivalente Mutation verworfen; Ergebnis unten. Nicht geprüft: Verifier-Fragen (DoD), Merge-Verträglichkeit mit dem neueren `main`.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Ein Literal außerhalb des Umfangs (`literal_typ`) gefolgt von einem ungültigen Literal im selben Schlüssel hat keinen Fänger: Die Mutation „Rückkehr nach `l.kompat(model.KompatLiteralTyp, ort)` in `leseKey`“ bleibt grün. Szenario: `/S(A=1d,B=null)` müsste `ODR-E6005` (400) liefern, die Mutation lieferte `ODR-E6006` (501). Die Zeilen 7 bis 9 paaren Formen über Pfad und Query, nicht zwei Teile eines Schlüssels; die Zeile 48 führt je Pfadform ein Literal. | `AGENTS.md` §3.16 (Paar von Ursachen ohne Fänger), `SPEC-173` Nr. 1 und 6 | `src/adapters/driven/odata/urlparserv2/schluessel.go` · `l.kompat(model.KompatLiteralTyp, ort)` | ja, Mutation auf Wegwerfkopie: `go test` bleibt grün | Paar `literal_typ` und `literal`/`zu_lang` im selben Schlüssel ohne Fänger |
| F-2 | MEDIUM | Die Rangfolge „Element-Zahl vor Pfad- oder Funktionsform“ ist an beiden Optionen ohne Fänger: `leseSelect` prüft `zu_viele` (257 Elemente) vor dem `/`-Test, `leseOrderBy` prüft `(` und `/` (Kompatibilität `option`) vor `zu_viele` (33 Elemente). Beide Vertauschungen bleiben grün (`$select` mit 257 Elementen und einem `/` müsste `zu_viele` sein, `$orderby` mit 33 Elementen und `tolower(` müsste `option` sein). Zeilen 42 und 43 führen 257 und 33 Elemente nur ohne Pfad und Klammer. | `AGENTS.md` §3.15 (Fänger führt nur eine Form der Eingabe), `SPEC-162` Nr. 6 und 7, `SPEC-173` Nr. 4 | `src/adapters/driven/odata/urlparserv2/query.go` · `if len(elemente) > maxSelect {` und `if strings.ContainsAny(wert, "(/") {` | ja, zwei Mutationen auf Wegwerfkopie: je `go test` grün | Rangfolge Grenze gegen Kompatibilitätsform ohne Fänger |
| F-3 | LOW | Die Zusatztabelle in §6 nennt „Zeile 7 (41 Paare)“ und „Zeile 9 (19 Paare)“, die Zeilen 7 und 9 in §3 und der Bullet in §6 tragen die gezählten Mengen 38 und 18. | `AGENTS.md` §3.9 (Plan folgt jeder Korrektur) | `docs/plan/planning/done/slice-v2-query.md` · `Zeile 7 (41 Paare)` | nein, kein Gate liest die Zahlen | Plan-Zahl an einer Stelle nachgezogen, an einer anderen nicht |
| F-4 | INFO | Die Plan-Zeile 22 (Versions-Kopfzeilen) und der Anhang A (Zeile 49) haben Ausgang *offen* mit Adresse (`slice-v2-bootstrap`, Planner). Die Adressen stehen in §6; die Annahme durch die Nehmer ist am Zweig nicht prüfbar (deren Pläne liegen auf `main`). Zur Rolle des Integrations-Planners. | `AGENTS.md` §3.13 | `docs/plan/planning/done/slice-v2-query.md` · `Der Anhang A, auf den Zeile 49 verwies` | nein | offene Sendung, Annahme beim Nehmer ungeprüft |
| F-5 | INFO | Dieser Zweig fasst `outbound/codec_satz.go` an, wie §6 selbst als Konfliktpunkt nennt; `main` hat sich seit der Merge-Basis geändert (Schicht `tests` im a-check, neue ADR). Das Ergebnis von `make gates` gilt für den Zweig-Stand, nicht für den Merge-Stand. | Maintainability | `src/hexagon/ports/outbound/codec_satz.go` · `URL URLCodec` | ja, `make gates` nach dem Merge | Gate-Ergebnis gilt nur für den ungemergten Stand |

Keine HIGH-Findings. Geprüft gegen die HIGH-Liste: Geheimnis in Fehlertext (`URLFehler` und `KompatFehler` tragen Ursache und Ort, nie Eingabe; Tests `TestUrlFehlerWert`, `TestKompatFehlerWert`), [ADR-0011](../plan/adr/0011-versionierter-adapter-fuer-odata-v2.md) (kein V4-Import, `TestKeinV4Import`, `make a-check` grün), Gate-Suppression (kein `nolint`), Kommentarklassen (kein Konjunktiv, keine Chronik gefunden), Spec-Straten (Code gegen `SPEC-173` Nr. 1 bis 7 gelesen, keine Abweichung gefunden).

## Mutationen des Reviewers (Wegwerfkopie, `go test ./src/adapters/driven/odata/urlparserv2/`)

Rot (gefangen): `$format` mit `application/json;o` als Präfix · erstes `$`-Vorkommen nach Escape verglichen (`gesehen`) · Unwrap-Liste von `errors.Join` nur erstes Element · leerer Alias `@` zulässig · `.` im Namen einer Sonstigen Option entfernt · `$index` aus der Liste außerhalb entfernt · Stunde 24 zulässig · Bruchteil 8 Ziffern · `INF` und `NaN` nicht als Wörter ausgenommen · `rawQuery` ohne U+007F · `*KompatFehler` und `*URLFehler` in einer Kette nicht als Aufruferfehler gewertet · `None` als `$inlinecount` · `$ref` nicht als Rest-Segment · `$count` hinter Schlüssel an jeder Stelle · Literalgrenze · Präfix jenseits von ASCII · Exponent `+` · Suffix `F` · Guid-Trenner an Stelle 23 · Ziffernzahl von `$top`.

Grün (nicht gefangen): F-1 (eine Mutation), F-2 (zwei Mutationen). Nicht gewertet: eine Mutation ohne Wirkung (`metadatafirst`, äquivalent) und eine mit Sed-Fehler.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/adapters/driven/odata/urlparserv2/` (Produktcode: `parser.go`, `pfad.go`, `schluessel.go`, `literal.go`, `zeit.go`, `query.go`, `dekodieren.go`) | geprüft gegen `SPEC-173`; Befunde F-1 und F-2 (Fänger), sonst ohne Befund; Nebenläufigkeit: kein Zustand der Instanz nach `New` |
| `src/adapters/driven/odata/urlparserv2/*_test.go` | geprüft, Fänger-Lücken F-1 und F-2; Mutationen siehe oben |
| `src/hexagon/model/` (`kompat_fehler.go`, `fehler.go`, Tests) | geprüft, ohne Befund |
| `src/hexagon/ports/outbound/` (`abbildung.go`, `codec_satz.go`, Test) | geprüft, ohne Befund (Reihenfolge der Tabelle: `ErrKompatibilitaet` vor der Gruppe `ODR-E9001`) |
| `src/adapters/driven/telemetry/`, `src/adapters/driving/httpodata/`, `src/hexagon/services/{session,workspace}/` (Zahlen 35 und 36 in Tests) | geprüft, ohne Befund |
| `docs/plan/planning/done/slice-v2-query.md` | geprüft; F-3 und F-4 |
| `spec/spezifikation.md` (Berichtigung `SPEC-173`) | nur gegen den Code gelesen, ohne Befund |
