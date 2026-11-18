# Review-Report: slice-bootstrap-m4-konfig — 2026-11-18

**Review-Art:** Code (gegen Plan, `SPEC-110` Nr. 16, `SPEC-261`, `SPEC-139`, `SPEC-037`, [ADR-0001](../plan/adr/0001-implementierungssprache-go.md) und `AGENTS.md` §3)

**Gegenstand:** Diff `ca7925ea..82875b8f` (Zweig `slice-bootstrap-m4`; Code, Tests, Handbuch, Plan; Teil 1 des Plans `slice-bootstrap-m4`, Zeilen M04 bis M09)

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-18

**Eingangs-Kontext:**

- Slice-Plan `slice-bootstrap-m4-konfig` (§1, §3 mit Tabelle *Zusage · Fänger*, §4 Größenmarke, §6 Raster und Entscheidungen (a) bis (g))
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `LH-FA-12`
- `SPEC-110` Nr. 1, 2, 8, 9, 12, 14, 15, 16; `SPEC-261` Nr. 8, 10 b; `SPEC-139` Nr. 12; `SPEC-037`; `SPEC-128`
- `AGENTS.md` (Hard Rules); vorherige Reports am Modul `src/bootstrap` wurden nicht neu gelesen

**Messung dieses Laufs.** Eigener Container-Treiber im Scratchpad (Image `odata-recorder:test`, ohne Netz, `--memory=4g --cpus=4`, unter `flock` auf `.wt/heavy.lock`, Last 5 bis 8 zu Beginn der Läufe, Messender der Treiber); `tools/harness/mutationslauf.sh` wurde nicht benutzt (Nullprobe von `src/bootstrap` über 2 Minuten, Register `BEO-REPO/nullprobe-grenze-unter-der-paketlaufzeit`). Je Mutation ein deterministischer Prozess mit den Tests `TestHybrid|TestKonfig|TestWerteErreichen`; der Ausgangsstand ist grün. Die Mutationen stehen am Quelltext einer Wegwerfkopie im Container, nicht im Repo. `make groessenmarke`: `Messung 671 von Marke 900`, gleich der Summe aus `git diff --numstat 7a5170d9..HEAD -- src/bootstrap/` (671).

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Dieselbe Ursache (fehlende `UPSTREAM_URL`) meldet im Modus `record` den Grund „im Modus record erforderlich“, im Modus `hybrid` „Pflichtfeld fehlt“. `SPEC-110` Nr. 16 f sagt für `hybrid` „wie im Modus `record` (Nr. 9)“, und Nr. 9 nennt für `upstream_url` im Modus `record` „Pflichtfeld fehlt“; der Text „im Modus record erforderlich“ steht nirgends in der Spezifikation, nur in Code und in den Tests A02 und M3. Die Stelle ist in §6 mit Ausgang *offen* und Adresse Architect geführt (der Implementer folgte für `hybrid` dem Wortlaut von Nr. 16 f und entschied nicht im Code). | `SPEC-110` Nr. 9, 16 f; `AGENTS.md` §3.16 | `src/bootstrap/konfig_dienste.go` · `return "im Modus record erforderlich"` neben `return "Pflichtfeld fehlt"` | ja — `make test` (`TestHybridUpstreamPflichtModus` hält beide Texte nebeneinander fest) | Spec-interner Widerspruch Grund der Pflicht, Code gegen Spezifikation |
| F-2 | MEDIUM | Die Zusage „ungültiges `HYBRID_FALLBACK` gilt weder als `enabled` noch als `disabled`: `ohne-hosts` und die Pflicht aus (f) entfallen“ (`SPEC-110` Nr. 16 e; Zeile 4 der Zusatztabelle des Plans nennt M04 und M05 als Fänger) hat für `ohne-hosts` keinen Fänger. Mutation: `policy.New` bekommt `Upstream=true`, wenn der Rohwert `Enabled` lautet (`HYBRID_FALLBACK=Enabled` mit leerer Allowlist meldete dann zusätzlich `ohne-hosts`): alle Tests `TestHybrid*`, `TestKonfig*` und `TestWerteErreichenVerbraucher` bleiben grün. Jeder Fall mit `Enabled` führt `FALLBACK_ALLOW_HOSTS=a.example`. Die Pflicht von `UPSTREAM_URL` bei ungültigem Schalter ist gefangen (`TestHybridUpstreamPflichtModus`). | `AGENTS.md` §3.15, §3.16; `SPEC-110` Nr. 16 e | `src/bootstrap/konfig_hybrid.go` · `_, err := policy.New(k.Fallback)`; Test `m4_konfig_test.go` · `"Enabled"` | ja — `make test` mit `HYBRID_FALLBACK=Enabled` und leerer `FALLBACK_ALLOW_HOSTS` | Zusage ohne Fänger, Zeile der Zusatztabelle ohne Fall |
| F-3 | MEDIUM | Die Zusage „`disabled` mit gesetzten Listen, eine Freigabe ohne Host ist kein Verstoß“ (`SPEC-110` Nr. 16 e, Raster Zeile 3, `SPEC-261` Nr. 8 c) hat auf der Konfiguration keinen Fänger. Mutation: der Root meldet bei `disabled`, leerer `FALLBACK_ALLOW_HOSTS` und gesetzter `FALLBACK_WRITE_ALLOW` einen Verstoß, ebenso bei gesetzter `HYBRID_REPLAY_FALLBACK`: beide bleiben grün (dieselbe Abwandlung ohne Einschränkung auf `disabled` wird rot, aber nur zufällig im Fall `enabled` mit leerer Allowlist und ungültiger Freigabe der Reihenfolge-Prüfung, nicht aus dem Grund der Zusage). Jeder Fall mit gesetzter Freigabe oder Kategorie trägt `FALLBACK_ALLOW_HOSTS=a.example`. | `AGENTS.md` §3.15, §3.16; `SPEC-110` Nr. 16 e | `src/bootstrap/m4_konfig_test.go` · `hybridBasis` (`"FALLBACK_ALLOW_HOSTS=a.example"`) | ja — `make test` mit `HYBRID_FALLBACK=disabled`, `FALLBACK_ALLOW_HOSTS` leer, `FALLBACK_WRITE_ALLOW=POST /a` | Fixture trägt in jedem Fall denselben Host, Zusage ohne Fänger |
| F-4 | LOW | Die Klasse „Leerraum um einen Eintrag“ (`SPEC-110` Nr. 6, 16 b) führt der Fänger nur mit dem Leerzeichen innerhalb der Listen. Mutation: `strings.Trim(e[i], " ")` statt `strings.TrimSpace(e[i])` bleibt grün (ein Tabulator vor oder nach einem Eintrag machte den Host ungültig). Die Spezifikation legt die Zeichen des Leerraums nicht fest. | `AGENTS.md` §3.15 (Fänger führt jede Form der Eingabe) | `src/bootstrap/konfig_hybrid.go` · `e[i] = strings.TrimSpace(e[i])` | ja — Fall mit Tabulator im Eintrag | Klasse Leerraum nur eine Form, Zeichenmenge in der Spezifikation offen |
| F-5 | INFO | `meldeKonfigFehler` verwirft einen Fehler, der weder `*policy.KonfigFehler` noch ein Join ist, ohne Meldung; `NeuePolicy` gibt denselben Fehler unverändert zurück (kein `*KonfigFehler`, kein `ODR-E2001`). `policy.New` liefert nach seinem Vertrag nur einen Join aus `*KonfigFehler`, kein Fall belegt den anderen Pfad; das akzeptierte Negativ in §6 nennt nur die Ursache außerhalb der sieben. | Maintainability | `src/bootstrap/konfig_hybrid.go` · `case *policy.KonfigFehler:` (kein `default`) | nein — kein Lauf erreicht den Pfad | Pfad ohne Fänger, Asymmetrie von Laden und `NeuePolicy` |
| F-6 | INFO | Die Zusage M04 „der Start scheitert, kein Listener öffnet“ ist für die Modi `record` und `replay` gefangen (`TestHybridKonfigFehlerStart`, der Kopf nennt genau diese zwei); `mock` und `hybrid` führt kein Start-Fall. Die Prüfung liegt in `LadeKonfiguration` vor jedem Modus (Mutation „Prüfung nur im Modus `hybrid`“ ist rot); der Kopf des Tests sagt die Lücke ehrlich. | `AGENTS.md` §3.11 | `src/bootstrap/m4_konfig_test.go` · „in jedem der beiden Modi mit Haken“ | nein | Start-Fänger nur zwei von vier Modi |
| F-7 | INFO | Handbuch (Ist-Zustand): die Zeilen `PUBLIC_BASE_URL`, `STRICT_COVERAGE` und `ETAG_ENABLED` nennen nur den Modus `mock`; die Prüfung verlangt `PUBLIC_BASE_URL` jetzt auch bei `MODE=hybrid` (Verstoß „im Modus mock oder hybrid bei diesem Host erforderlich“). Der Dienst startet in `hybrid` nicht; `slice-bootstrap-m4` (Teil 2) schreibt die Zeilen um (dessen §3, Zeile `docs/user/`). | `SPEC-200` Nr. 1 | `docs/user/benutzerhandbuch.md` · Zeile `PUBLIC_BASE_URL`, Satz „bei jedem anderen Host verlangt der Modus mock sie“ | nein — `make handbuch-check` prüft es nicht | Handbuch-Zeile nennt eine Pflicht der Prüfung nicht |
| F-8 | INFO | Die Rückgaben aus §6 (Feld `Fallback` ohne Verbraucher in `TestWerteErreichenVerbraucher`; Name `NeuePolicy`; Nullprobe; `slice-bootstrap-m6`) tragen Ausgang *offen* und Adresse im selben Commit wie der Code (`AGENTS.md` §3.13 Abs. 3 erfüllt). §1 und DoD von `slice-bootstrap-m4` (Teil 2) nehmen `Fallback` in der Tabelle der Verbraucher und den Namen `NeuePolicy` noch nicht an; solange `Fallback` in der Liste `ohneVerbraucher` steht, fängt kein Test, ob Teil 2 es bindet. | `AGENTS.md` §3.13 | Plan `slice-bootstrap-m4-konfig` §6 · Zeile „Feld Fallback der Konfiguration“ | nein | Folge-Slice nimmt Sendung noch nicht an |
| F-9 | INFO | Die 41 Mutationen des Implementers stehen als Namen im Plan; die Liste (`alt` und `neu`) liegt im Scratchpad des Implementers und nicht im Repo. Ein Folgelauf kann sie nicht wiederholen und misst, wie dieser, mit eigenen Mutationen. | `AGENTS.md` §3.10 | Plan `slice-bootstrap-m4-konfig` §3 · Satz „Alle 41 Mutationen sind Klasse andere“ | nein | Mutationsliste nicht im Repo |

## Mutationen dieses Laufs (Klasse, nicht die genannten Beispiele; §3.17)

22 gültige Mutationen, Klasse `andere` (deterministisch, 1 von 1 je Mutation, kein Fänger wahrscheinlichkeitsbehaftet, keine Nebenläufigkeitszusage im Diff).

| Zusage | Mutation | Ergebnis (aus dem benannten Grund?) |
|---|---|---|
| M04 Eintragszählung | Doppelte in `eintraege` entfernt | rot: `TestHybridKonfigFehlerJeUrsache`, `TestHybridKonfigSyntax` (Zählung `Eintrag 3`) |
| M04 `ohne-hosts` ohne `Eintrag` | `e.Eintrag >= 0` | rot: `...JeUrsache`, `...Reihenfolge`, `TestHybridUpstreamPflichtModus` |
| M04 Text je Ursache | Text von `muster` geändert (Element, das der Plan nicht nannte) | rot: `...JeUrsache`, `TestHybridKonfigSyntax` |
| M04 vier Schlüssel gelesen | `FALLBACK_ALLOW_HOSTS` nicht gelesen; Freigaben mit der Kategorien-Liste vertauscht | rot (beide) |
| M04 `NeuePolicy` | Fehler roh, ohne Abbildung; Nullwert statt Policy | rot: `TestHybridNeuePolicyFehler`, `TestHybridNeuePolicy` |
| M04 Prüfung in jedem Modus | `liesHybrid` nur im Modus `hybrid` | rot: `...JeUrsache`, `...Start`, `...JederModus` |
| M04 Zusammenhang Host ohne Service | Verstoß bei `disabled` mit Hosts | rot: `...JeUrsache`, `...ErsterJeSchluessel`, `...Reihenfolge` |
| M05 Pflicht nur bei gültigem Modus | Pflicht auch bei ungültigem `MODE` | rot: `TestHybridUpstreamPflichtModus` |
| M05 Pflicht bei `enabled` | Pflicht zusätzlich an Hosts gebunden | rot: `TestHybridUpstreamPflichtModus` |
| M05 Typ des Feldes vor Pflicht | falscher Typ gilt als gültig | rot: `TestKonfigDienstefeldTyp` |
| M05 / F-1 | beide Gründe vereinheitlicht | rot: `TestKonfigServicesFile`, `TestKonfigUpstreamModus` (der Widerspruch ist gesehen, F-1) |
| M07 `PUBLIC_BASE_URL` | Länge nur im Modus `mock`; Pflichttext ohne `hybrid` | rot: `TestHybridKonfigMockSchluessel`, `TestKonfigPublicBaseURLAbleitung` |
| M08 Familien | `FALLBACK_` wieder in `nichtGelieferte` | rot: `...JeUrsache`, `...ErsterJeSchluessel`, `...Reihenfolge` |
| M06 Leerraum | `Trim(e[i], " ")` | **grün** (F-4) |
| `SPEC-110` Nr. 16 e, ungültiger Schalter | `Upstream=true` an `policy.New` bei `Enabled` | **grün** (F-2) |
| `SPEC-110` Nr. 16 e, `disabled` mit Freigabe ohne Host | Verstoß bei leerer Allowlist und gesetzter Freigabe, nur bei `disabled` | **grün** (F-3) |
| `SPEC-110` Nr. 16 e, `disabled` mit Kategorie ohne Host | dasselbe für `HYBRID_REPLAY_FALLBACK` | **grün** (F-3) |

Eine weitere Mutation (`!= "disabled"` statt `== "enabled"`) ist äquivalent, weil `aufzaehlung` bei ungültigem Wert den Default liefert; sie zählt nicht als Überlebende. Mehrere Mutationen scheiterten zunächst am Aufbau und wurden umgeschrieben (erst der Lauf nach dem Umschreiben zählt).

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/konfig.go`, `konfig_basis.go`, `konfig_werte.go` | geprüft, ohne Befund (Kommentare tragen Zusage oder Kopplung; keine verworfene Alternative) |
| `src/bootstrap/konfig_dienste.go` | geprüft, Befund nur F-1 (Wortlaut) |
| `src/bootstrap/konfig_hybrid.go` | geprüft, Befunde F-2 bis F-5; Reihenfolge der Meldung, Eintragszählung und Abbildung auf `ODR-E2001` gefangen |
| `src/bootstrap/m4_konfig_test.go` und die angepassten Tests (`konfig_test.go`, `m3_konfig_test.go`, `start_test.go`, `wurzel_test.go`) | geprüft; Köpfe nennen, was sie fangen (§3.11), Mengen mit Zählprüfung (§3.15, §3.17); Befunde F-2, F-3, F-4, F-6 |
| `tests/e2e/stufe_prozess_test.go` | geprüft, ohne Befund (`ADMIN_LISTEN` ersetzt `MODE=hybrid` als nicht unterstützter Schlüssel, `SPEC-037` Nr. 5, 7) |
| `docs/user/benutzerhandbuch.md` | geprüft, Befund nur F-7; keine Planung oder Chronik im Text |
| `spec/spezifikation.md`, `harness/vertraege/SPEC-037-test-integration.md` | geprüft gegen Plan und Code: Nr. 2, 8, 9, 12, 14, 15, 16 widersprechen einander nicht außer in F-1 (Nr. 9 gegen den Code); keine Architektur-Sicht berührt |
| Plan `slice-bootstrap-m4-konfig` §1, §3, §4, §6 | geprüft: Kopf, Tabelle und Raster stimmen mit dem Code überein, Orte tragen Wortlaut; Marke 900 gleich Formel, Messung 671 reproduziert; Befunde F-2, F-3 (Zusatztabelle), F-8, F-9 |
| `harness/`, `.harness/skills/`, Hooks | nicht im Diff; keine Unterdrückung (`//nolint`) im Diff gefunden |
| Gate `make gates` | nicht Gegenstand dieses Reviews (Verifier); `make groessenmarke` gelaufen, grün |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 3 |
| LOW | 1 |
| INFO | 5 |

**Finding-Klassen dieses Laufs:** Spec-interner Widerspruch Grund der Pflicht · Zusage ohne Fänger, Zeile der Zusatztabelle ohne Fall · Fixture trägt in jedem Fall denselben Host · Klasse Leerraum nur eine Form · Pfad ohne Fänger · Start-Fänger nur zwei von vier Modi · Handbuch-Zeile nennt Pflicht der Prüfung nicht · Folge-Slice nimmt Sendung noch nicht an · Mutationsliste nicht im Repo

## Verdikt

**Merge-blockierend:** ja — drei MEDIUM (F-1, F-2, F-3). F-1 liegt beim Architect (§6, Ausgang *offen*); F-2 und F-3 sind Zusagen der Zusatztabelle ohne Fänger (`AGENTS.md` §3.15). Kein HIGH: kein ADR-Verstoß, keine Suppression, kein Wert in der Meldung (gefangen).

**Übergabe:** F-1 an den Architect (ein Text für beide Modi oder der Satz „wie im Modus `record`“ gestrichen), F-2 bis F-4 an den Implementer, F-5 bis F-9 zur Kenntnis; die Finding-Klassen gehen in die Slice-Closure §7. Der Report ersetzt keine Verifikation.
