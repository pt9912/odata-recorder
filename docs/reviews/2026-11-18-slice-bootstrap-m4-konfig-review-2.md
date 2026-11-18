# Review-Report: slice-bootstrap-m4-konfig — 2026-11-18 (Runde 2)

**Review-Art:** Code (gegen Plan, `SPEC-110` Nr. 1, 6, 9, 16, `SPEC-261`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und `AGENTS.md` §3); Prüfung der Fixes von Runde 1 gegen die Klasse der Zusage (`AGENTS.md` §3.17)

**Gegenstand:** Diff `6da21441..692def3f` (Zweig `slice-bootstrap-m4`; Commits `23646db3` zu F-1 bis F-3 und `692def3f` zu F-4) und Gesamtstand von `src/bootstrap/konfig*.go`, `m4_konfig_test.go`, `SPEC-110` Nr. 1, 6, 9, 16 und Plan `slice-bootstrap-m4-konfig`

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext:**

- Report Runde 1 (`docs/reviews/2026-11-18-slice-bootstrap-m4-konfig-review.md`, F-1 bis F-9)
- Slice-Plan `slice-bootstrap-m4-konfig` (§1, §3, §4 Größenmarke, §6 Raster und Entscheidungen)
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-12`, `LH-FA-14`
- `SPEC-110` Nr. 1, 6, 9, 16; `SPEC-261` Nr. 8, 10 b; `SPEC-128` Nr. 1, 5; `SPEC-115` (Redaktion der Query-Parameter)
- `AGENTS.md` (Hard Rules)

**Messung dieses Laufs.** Eigener Container-Treiber im Scratchpad (Image `odata-recorder:test`, ohne Netz, `--memory=4g --cpus=4`, unter `flock` auf `.wt/heavy.lock`, Last 10 bis 16 zu Beginn der Läufe, erster Wert von `/proc/loadavg`, Messender der Treiber); `tools/harness/mutationslauf.sh` wurde nicht benutzt (Nullprobe von `src/bootstrap` über 2 Minuten, Register `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`). Je Mutation ein deterministischer Prozess an einer Kopie im Container (der Arbeitsbaum blieb unverändert), alle Tests der Familien `TestKonfig*`, `TestHybrid*` und `TestUpstream*` (Laufzeit 2 bis 3,5 s je Lauf); ein Rot zählt nur mit der Meldung des Tests an der Stelle der Zusage. Kein Fänger dieses Diffs ist wahrscheinlichkeitsbehaftet, keine Zusage betrifft Nebenläufigkeit.

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | HIGH | Ein Eintrag von `REDACT_QUERY_PARAMS` mit U+00A0, U+0085 oder VT am Rand (`REDACT_QUERY_PARAMS=<U+00A0>tok`) ist seit `692def3f` ohne Verstoß gültig: die Konfiguration liefert den Namen mit dem Zeichen (Probe: Zeilen leer, Namen enthalten `<U+00A0>tok`), und `model.RedigiereQuery` vergleicht den Namen exakt, so dass `?tok=geheim` unredigiert ins Recording geht; vor dem Fix entfernte `TrimSpace` das Zeichen. `SPEC-110` Nr. 6 sagt dagegen, ein solcher Eintrag „scheitert an der Form seines Namens“, und das akzeptierte Negativ im Plan nennt als Folge „der Start meldet `ODR-E2001`, nie eine stille Freigabe“; beides gilt für Header, JSON-Pfade und die drei Listen von `hybrid`, nicht für Query-Parameter-Namen (`istQueryName` schließt nur Leerraum, `=`, `&`, `#`, `,` aus). Der Fänger `TestKonfigListenFremderLeerraum` lässt beide Ausgänge zu (Verstoß oder Name mit dem Zeichen), nimmt also das Verhalten gegen die Spezifikation hin. | `LH-FA-14` (Geheimnis in persistenter Senke), `SPEC-110` Nr. 6 | `src/bootstrap/konfig_werte.go` · `return s != "" && !strings.ContainsAny(s, " \t\r\n=&#,")`; `src/bootstrap/m4_konfig_test.go` · `keine bereinigte Fassung (Query-Parameter-Namen schließen nur den Leerraum der Nr. 1 aus)` | ja — `make test` mit einem Fall `REDACT_QUERY_PARAMS=<U+00A0>tok` erwartet Verstoß (heute rot-fähig, Probe-Lauf gesehen) | Fail-open durch Fix: Zeichen bleibt im Eintrag, Form des Namens fängt es nicht |
| F-2 | MEDIUM | Die Zusage „Leerraum (Leerzeichen, Tabulator, LF, CR) wird an jeder Stelle entfernt, die kappt“ hat nur an den drei Listen von `hybrid` einen Fänger für alle vier Zeichen und beide Seiten; an den übrigen Stellen führt der Fänger die Zeichen des Leerraums nur teilweise. Grün bleiben: `kappe` in `upstreamURL` ganz entfernt, nur mit Leerzeichen, ohne LF und CR, nur links oder nur rechts (`UPSTREAM_URL` mit Leerraum am Rand hat keinen Fänger); `kappe` in `liste` (die fünf Listen außerhalb von `hybrid`, Einträge zwischen den Kommas) nur mit Leerzeichen oder ohne LF und CR; `kappe` an den Feldern der Services-Datei ohne CR; `wert()` links ohne LF und CR. Die Tabelle des Plans nennt je Stelle einen Fänger („jede der sechs Stellen“) und belegt ihn mit der Mutation `TrimSpace`, die nur die Ausschlussseite (U+00A0 und weitere bleiben im Wert) fängt. | `AGENTS.md` §3.15, §3.17 | `docs/plan/planning/done/slice-bootstrap-m4-konfig.md` · `jede Stelle, die kappt (6 Stellen im Code)` | ja — Mutationen a1, a2, a4, a5, a6, a6b, a6c, a8, d1 dieses Laufs (Tabelle unten) | Fänger führt nur eine Seite der Klasse Leerraum je Stelle |
| F-3 | MEDIUM | `istQueryName` ist eine weitere Stelle, die den Leerraum von `SPEC-110` Nr. 1 und 6 trägt („ein Query-Parameter-Name enthält weder Leerraum noch …“), und fehlt in der Menge der „sechs Stellen“ des Plans. Der Fänger führt allein das Leerzeichen: die Zeichenmenge ohne Tabulator, LF und CR (`" =&#,"`) und die ohne LF bleiben grün; ein Name `a<TAB>b`, `a<LF>b` oder `a<CR>b` ist ein Fall, den kein Test führt. | `AGENTS.md` §3.15 (ein Fall je Element einer Menge), `SPEC-110` Nr. 6 | `src/bootstrap/konfig_werte.go` · `func istQueryName(s string) bool` | ja — Mutationen c2 und d4 dieses Laufs | Menge der vier Zeichen an einer Stelle mit einem Element geführt |
| F-4 | MEDIUM | Die Messung 1050 liegt über der Marke 900; der Eintrag `ueberschreitung: 1050` trägt als Grund die Ursache (226 Zeilen Fänger, Produktcode 162 gegen 150) und überlässt „Marke und Schnitt“ dem Planner, aber §6 des Plans hat keine Zeile mit Ausgang *offen* und der Adresse Planner; §6 führt zur Größe nur den Schnitt vor dem Code. Der Grund sagt nicht, warum kein Schnitt folgt, und die Rückführung `in-progress` → `next` (über 900) steht als offene Bedingung ohne Adresse. | `AGENTS.md` §3.13 (eine zurückgegebene Entscheidung steht im selben Commit in §6), §3.14 | `docs/plan/planning/done/slice-bootstrap-m4-konfig.md` · `Marke und Schnitt entscheidet der Planner` | nein — kein Gate liest §6 | Rückgabe an den Planner nur im Block, nicht in §6 |
| F-5 | LOW | Die Historie der Spezifikation trägt eine Zeile für den Grund der fehlenden `UPSTREAM_URL`, aber keine für die geänderten Sätze `SPEC-110` Nr. 1 und 6 (Zeichen des Leerraums, Verhalten von U+00A0, U+0085, VT und FF); Nr. 1 und 6 sind dieselbe Art Änderung wie die Zeile darüber. | `AGENTS.md` §3.9 (alle Stellen, die eine Aussage tragen) | `spec/spezifikation.md` · `## 7. Historie` | nein | Historie der Spezifikation hinter einer Änderung zurück |
| F-6 | LOW | Das akzeptierte Negativ „die übrigen Unicode-Leerraum-Zeichen sind unbeschränkt“ trifft nicht zu: die Menge der Zeichen mit der Eigenschaft White_Space ist endlich (25, `SPEC-201` Nr. 2 e nennt sie), 21 liegen außerhalb des Leerraums von Nr. 1, der Fänger führt fünf davon (VT, FF, U+0085, U+00A0, U+2007) und das BOM. `kappe` um U+3000, U+2028 oder U+1680 und U+202F erweitert, bei unveränderter Konstante `leerraum`, bleibt grün; die Zählprüfung prüft die Konstante, nicht das Verhalten von `kappe`. | `AGENTS.md` §3.17 (Menge aus dem Wortlaut der Zusage, Prüfgrenze nur bei unbeschränkter Menge) | `docs/plan/planning/done/slice-bootstrap-m4-konfig.md` · `sind unbeschränkt und stehen als akzeptiertes Negativ` | ja — Mutationen z4, z5, z6 dieses Laufs | Endliche Menge als unbeschränkt benannt |
| F-7 | INFO | Die Fänger zu F-4 wiederholen dieselbe Schleife dreimal (`TestKonfigListenFremderLeerraum`, `TestHybridListeFremderLeerraum`, `TestHybridListeLeerraumUmEintrag`: Einträge bauen, Seiten setzen); `TestUpstreamPflichtGrundGleich` prüft zusätzlich zu `je[i] == soll` die Gleichheit `je[0] == je[1]`, die daraus folgt, und der Grund `Pflichtfeld fehlt` steht in vier weiteren Tests (`TestHybridUpstreamPflichtModus`, `TestKonfigUpstreamModus`, `TestKonfigUpstreamNurImRecordPflicht`, `TestKonfigServicesFile`). Die Zählprüfung `len == 4` der Konstante macht die äquivalente Mutation `leerraum` mit doppeltem Leerzeichen rot (z3). Beitrag zur Überschreitung von F-4. | Maintainability | `src/bootstrap/m4_konfig_test.go` · `func TestUpstreamPflichtGrundGleich` | nein | Wiederholte Schleife, abgeleitete Prüfung |
| F-8 | INFO | Die Fixes zu F-1 bis F-3 der Runde 1 decken ihre Klassen: Grund der Pflicht (2 Modi × 5 Aufbauten, Zeilen beider Modi gleich), ungültiger Schalter (4 Modi × 8 Werte × 3 Belegungen) und `disabled` ohne Host (4 Modi × 2 Schalter × 4 Belegungen × 2 Allowlisten) haben Zählprüfung und Schleife; eigene Mutationen an nicht genannten Elementen sind rot (c3 bis c6, siehe Tabelle). Kein Befund; Hinweis für die Verifikation. | `AGENTS.md` §3.17 | `src/bootstrap/m4_konfig_test.go` · `func TestHybridUngueltigerSchalterOhneFolgen` | ja | — |

## Mutationen dieses Laufs (Klasse, nicht die genannten Beispiele; §3.17)

34 Mutationen, Klasse `andere` (deterministisch, 1 von 1 je Mutation); 20 rot, 14 grün, keine äquivalente unter den grünen (jede ändert das Verhalten für einen Wert der Klasse, den der Test nicht führt; für `d1` ist `url.Parse` mit führendem Leerzeichen ein Fehler). Die Liste liegt im Scratchpad dieses Laufs und nicht im Repo (wie Runde 1, F-9).

| Zusage | Mutation | Ergebnis (aus dem benannten Grund?) |
|---|---|---|
| Leerraum bei den fünf Listen außerhalb `hybrid` | `liste`: `Trim(e, " ")` (a1), `Trim(e, " \t")` (a2) | **grün** (F-2) |
| dasselbe, Seiten | `liste`: `TrimLeft` (a2b), `TrimRight` (a2c) | rot: `TestKonfigListe`, `TestKonfigListenErgaenzen`, `TestKonfigNone` |
| Leerraum an den Feldern der Services-Datei | `Trim(*s, " ")` (a3) | rot: `TestKonfigServicesFile` (Pflichtfeld fehlt) |
| dasselbe | `Trim(*s, " \t\n")` (a4) | **grün** (F-2) |
| dasselbe, Seite | `TrimLeft` (a4b); ganz ohne `kappe` (d2) | rot: `TestKonfigServicesFile` |
| Leerraum bei `UPSTREAM_URL` | nur Leerzeichen (a5), ohne LF und CR (a6), `TrimLeft` (a6b), `TrimRight` (a6c), ganz ohne `kappe` (d1) | **grün** (F-2) |
| Leerraum eines Wertes | `wert()`: nur Leerzeichen (a7) | rot: `TestKonfigLeer`, `TestHybridKonfigSyntax` |
| dasselbe | links ohne LF und CR (a8) | **grün** (F-2) |
| dasselbe | rechts ohne LF und CR (a8b); ganz ohne `kappe` (d3) | rot: `TestKonfigFremderLeerraumAnderenStellen` (`MODE` mit Leerraum am Rand); `TestKonfigEtagEnabled` |
| `ADMIN_*` gesetzt nur bei Nichtleerraum | `nichtGeliefert`: nur Leerzeichen (a9), ohne CR (a10) | rot: `TestKonfigFremderLeerraumAnderenStellen` (`ADMIN_X` mit Leerraum allein) |
| Leerraum um Einträge von `hybrid` | `eintraege`: `TrimLeft` (a11), ohne CR (a12) | rot: `TestHybridKonfigSyntax`, `TestHybridListeLeerraumUmEintrag` |
| Zeichen des Leerraums = vier | Konstante mit U+3000 (z1), mit NUL (z2), mit doppeltem Leerzeichen (z3, äquivalent im Verhalten) | rot: `TestHybridListeLeerraumUmEintrag` (Zählprüfung; z3 ist ein Rot ohne Wirkung, F-7) |
| dasselbe, Verhalten statt Konstante | `kappe` um U+3000 (z4), U+2028 (z5), U+1680 und U+202F (z6) erweitert, Konstante unverändert | **grün** (F-6) |
| Query-Name ohne Leerraum | `istQueryName` ohne alle Zeichen des Leerraums (c1) | rot: `TestKonfigListe` |
| dasselbe | ohne LF (c2); nur Leerzeichen (d4) | **grün** (F-3) |
| Pflicht aus `SPEC-110` Nr. 9, 16 f | `(record \|\| hybrid) && Upstream` (c3); `record \|\| hybrid` (c4); `record \|\| Upstream` (c5) | rot: `TestKonfigDefaults`, `TestKonfigServicesFile`; `TestHybridUpstreamPflicht`, `TestKonfigModus`; `TestHybridUpstreamPflichtModus` |
| `SPEC-110` Nr. 16 e (ungültiger Schalter prüft die Einträge weiter) | `policy.New` nur bei gültigem oder ungesetztem Schalter (c6) | rot: `TestHybridKonfigFehlerReihenfolge` (Paare mit `HYBRID_FALLBACK`) |
| Probe zu F-1 | `REDACT_QUERY_PARAMS` mit U+00A0 vor, U+00A0 nach, U+0085 vor, VT vor dem Namen | Probe (kein Fänger): Zeilen leer, Namen tragen das Zeichen |

Die Rate je Mutation ist 1 von 1 (ein Prozess, deterministisch, kein Detektor nötig).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/konfig.go` (`leerraum`, `kappe`, `wert`) | geprüft, ohne Befund am Code; Kommentare tragen Zusage oder Kopplung; Fänger siehe F-2 |
| `src/bootstrap/konfig_dienste.go` (`upstreamPflicht`, `upstreamURL`, `dateiEintrag`) | geprüft, Grund der Pflicht in beiden Modi ein Text, Pflicht-Matrix (c3 bis c5 rot); Befund nur F-2 (Fänger) |
| `src/bootstrap/konfig_hybrid.go`, `konfig_werte.go` | geprüft; Befunde F-1, F-3 (Stellen `istQueryName`) und F-2 (Fänger); keine verworfene Alternative im Kommentar |
| `src/bootstrap/export_test.go`, `m3_konfig_test.go`, `konfig_test.go` | geprüft, ohne Befund (Textanpassung `Pflichtfeld fehlt`, `Leerraum` als Konstante exportiert) |
| `src/bootstrap/m4_konfig_test.go` | geprüft; Köpfe nennen ihre Menge und die Zählung, Werte aus Bytewerten (`fremdeZeichen`), Datei mit `LC_ALL=C grep -P` auf Rohbytes VT, FF, U+0085, U+00A0, U+2007 und U+FEFF gegengelesen, keine Folge Backslash-u im Quelltext (`AGENTS.md` §3.18); Befunde F-1, F-2, F-3, F-6, F-7 |
| `spec/spezifikation.md` (`SPEC-110` Nr. 1, 6, 9, 16 f) | geprüft gegen Code und Plan: Nr. 9 und 16 f nennen denselben Grund; Nr. 6 gegen den Code für Query-Namen im Widerspruch (F-1); Historie F-5 |
| Plan `slice-bootstrap-m4-konfig` (Kopf, §1, §3, §4, §6) | geprüft: Kopf, §1 und §3 tragen die Nachzüge von F-1 bis F-4; Mengen 36, 54, 60, 90, 96, 64, 10 stimmen mit den Zählprüfungen überein; Befunde F-2, F-4, F-6 |
| `docs/plan/planning/in-progress/slice-bootstrap-m4.md`, Report Runde 1 | geprüft, ohne Befund (Zeilen von Teil 2 und Verweise nachgezogen) |
| Größenmarke | Marke 900 gleich Formel; Messung 1050 reproduziert (`git diff --numstat 7a5170d9..HEAD -- src/bootstrap/`); Eintrag `ueberschreitung: 1050` deckt; ein Schnitt beim Handoff ist nach `AGENTS.md` §3.14 nicht erlaubt (kein vorab benannter Teil), die Konstante und ihre Fänger teilen den Mechanismus; Befund F-4 nur zur Adresse |
| `harness/`, `.harness/skills/`, Hooks, `docs/user/` | nicht im Diff; keine Unterdrückung (`//nolint`) im Diff gefunden |
| Gate `make gates` | nicht Gegenstand dieses Reviews (Verifier) |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 1 |
| MEDIUM | 3 |
| LOW | 2 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Fail-open durch Fix: Zeichen bleibt im Eintrag, Form des Namens fängt es nicht · Fänger führt nur eine Seite der Klasse Leerraum je Stelle · Menge der vier Zeichen an einer Stelle mit einem Element geführt · Rückgabe an den Planner nur im Block, nicht in §6 · Historie der Spezifikation hinter einer Änderung zurück · Endliche Menge als unbeschränkt benannt · Wiederholte Schleife, abgeleitete Prüfung

## Verdikt

**Merge-blockierend:** ja — ein HIGH (F-1) und drei MEDIUM (F-2, F-3, F-4). F-1 ist ein Rollen-Widerspruch zu erwarten (die Entscheidung des Architect nennt das Zeichen als im Wert verbleibend); ein HIGH wird nicht herabgestuft, und der Konflikt läuft als Sequenz über den Architect mit Übergabe-Artefakt (dieser Report, F-1 mit Probe).

**Übergabe:** F-1 an den Architect (`SPEC-110` Nr. 6 und der Plan gegen den Code: entweder die Form der Query-Parameter-Namen oder der Satz „scheitert an der Form seines Namens“ trägt die Entscheidung; das Negativ „nie eine stille Freigabe“ und die Raster-Zelle ziehen mit); F-2, F-3 an den Implementer (Klasse: Zeichen × Stelle, mit einer Mutation an einem Element, das dieser Report nicht nannte, `AGENTS.md` §3.17); F-4 an den Planner (Zeile in §6 mit Ausgang *offen* und Adresse, Grund zum Schnitt); F-5, F-6 an den Implementer, F-7, F-8 zur Kenntnis. Die Finding-Klassen gehen in die Slice-Closure §7. Der Report ersetzt keine Verifikation.
