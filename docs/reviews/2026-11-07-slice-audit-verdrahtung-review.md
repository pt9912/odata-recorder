# Review-Report: slice-audit-verdrahtung — 2026-11-07

**Review-Art:** Code — Code-Review gegen Plan + Konventionen (nur Tests, Produktcode 0).

**Gegenstand:** `slice-audit-verdrahtung`, Zweig `slice-audit-verdrahtung`, Diff `8667f9d..1693714`.

**Skill:** `.harness/skills/reviewer.md` @ 1693714 ·
**Modell:** claude-sonnet-5-5 · **Datum:** 2026-11-07

**Eingangs-Kontext:**

- Plan `slice-audit-verdrahtung` (§3 Tabelle V01a bis V07, Fälle F1 bis F7, Zeitpunkte Z1 bis Z4, Mengen M1 bis M4, Zweig-Menge W/N, Stand der Fänger; §6 Ausgänge und Raster)
- `LH-RB-04`, `LH-FA-13`, `LH-QA-04`; `SPEC-139` Nr. 2 bis 4, 6, 7, 10 und *Ungeprüft*; `SPEC-234` Nr. 3 und 8; `SPEC-235` Nr. 1, 2, 6 bis 8
- [ADR-0001](../plan/adr/0001-implementierungssprache-go.md), [ADR-0006](../plan/adr/0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md)
- `AGENTS.md` (Hard Rules, besonders §3.9, §3.10, §3.15 bis §3.18)

**Eigene Läufe:** `tools/harness/mutationslauf.sh`, Image `odata-recorder:test`, 4g, unter `flock /tmp/odata-recorder-heavy.lock`, 5 Mutationen, Klasse `andere`, 1 bis 2 Prozesse je Mutation, `last` 4,3 bis 6,2 von 4 (Host belegt, Messender: Reviewer). Die 20 Mutationen des Implementers sind nicht wiederholt.

| Mutation (lokal, nicht eingecheckt) | Test | Ergebnis |
|---|---|---|
| `clock.ZufallsIDs.KorrelationsID` liefert `c` + Hex(`cnry_passwort-G2`) + 9 Zeichen (Kennungsform, je Zeile verschieden; Stellvertreter für einen Wrapper oder eine ID-Quelle des Roots) | `TestAuditCanary` | gruen (Paket ebenfalls grün) |
| dieselbe Mutation | `TestAuditEintraegeJeFall` | gruen (Paket grün) |
| tote fünfte Aufrufstelle `a := v.k.Audit; a.Append(...)` in `session/loeschen.go` | `TestAuditAufrufstellen` | gruen |
| tote fünfte Aufrufstelle `v.k.Audit.Append(...)` (Kontrolle) | `TestAuditAufrufstellen` | rot |
| `Append` vor `Sessions.Delete` (Ablehnung wegen Referenz schriebe einen Eintrag) | `TestAuditAblehnungenKeinEintrag` | gruen; das Paket wurde rot-paket allein durch `TestAuditAufrufstellen`, weil die Mutation eine zweite Aufrufstelle anlegt |

---

## Findings

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Prüfgrenze »Weitere Kodierungen« nennt als Folge, ein Geheimnis in Kennungsform fiele an M1 (der Wert weicht ab). Für `correlation_id` stimmt das nicht: M1 prüft das Feld nur auf Kennungsform und Eindeutigkeit, nicht auf den Wert; ein kodiertes Geheimnis in `correlation_id` (Hex von `cnry_passwort-G2`) lässt `TestAuditCanary` und `TestAuditEintraegeJeFall` grün. Das Feld ist das eine, dessen Wert der Root (über Uhr, IDs, Wrapper) beeinflusst. | `AGENTS.md` §3.15, §3.17; Reviewer-Skill MEDIUM »Fänger schmaler als Zusage, Menge unbeschränkt oder ohne Achsenliste« | `docs/plan/planning/in-progress/slice-audit-verdrahtung.md` · »fiele an M1 (der Wert weicht ab), nicht an M2«; `src/bootstrap/audit_betrieb_test.go` · `pruefeInhalt` (`kennung.MatchString(m["correlation_id"])`) | ja — Mutationslauf oben, 2 Zeilen gruen | Folge einer Prüfgrenze nennt ein Orakel, das das Feld nicht prüft |
| F-2 | LOW | Die Zählprüfung V07 erkennt nur Aufrufe der Form `X.Audit.Append(...)`. Eine fünfte Aufrufstelle über eine lokale Variable (`a := v.k.Audit; a.Append(...)`) in einem Zweig, den kein Fall auslöst, bleibt ungezählt (Mutation gruen, die Form `v.k.Audit.Append` rot). Die Fundstellen im Bestand haben alle die erkannte Form (4). | `AGENTS.md` §3.15 (Fänger schmaler als Zusage); `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (Verwandtschaft, nicht Nebenläufigkeit) | `src/bootstrap/audit_betrieb_test.go` · `feld.Sel.Name == "Audit"` in `TestAuditAufrufstellen` | ja — Mutationslauf oben | Zählprüfung per Syntaxbaum erkennt eine Form des Aufrufs |
| F-3 | LOW | M3 (Ablehnungen, 6 Elemente) führt »referenziert« nicht: eine Löschung, die der Port wegen einer Baseline-Quelle ablehnt, ist mit dem Helfer `saeeWorkspaces` säbar (Baseline mit Quelle `ses-1`), steht aber weder in M3 noch in den benannten Negativen (diese nennen nur die Aufbewahrung). `Append` vor `Sessions.Delete` lässt den zuständigen Test grün; ob der Fänger der Dienste (`slice-audit-dienste`) die Form fängt, ist in diesem Lauf nicht geprüft. | `AGENTS.md` §3.15 (ein Fall je Element der Menge), §3.16 | `src/bootstrap/audit_betrieb_test.go` · `{"Session offen", model.CodeE7009, ...}` (Liste der Ablehnungen) | ja — Mutationslauf oben | Menge der Ablehnungsgründe ohne den Fall »referenziert« |
| F-4 | LOW | `SessionsMitFehlertext` zählt jede Operation, F7 prüft aber nur die Aufrufe von `Get` (`["ses-f7"]`). Dass der Import nach dem Fehler von `Get` kein `Create`, `SetStatus` oder `Delete` ruft, bleibt unbelegt; das Zählen ist verdrahtet, die Behauptung fehlt. | `AGENTS.md` §3.15 (Fake zählt jede Operation, sonst kein Aufruf) | `src/bootstrap/audit_betrieb_hilfen_test.go` · `func (s *SessionsMitFehlertext) Aufrufe`; `audit_betrieb_test.go` · `a.get` | ja | Fake zählt Operationen, der Test liest eine |
| F-5 | INFO | Die ungedeckten Fehlerzweige der Konstruktoren in `verdrahte` und `verdrahteRecord` (`blobs.New` bis `NeuerReplay`): Entscheidung des Reviewers — **kein Haken je Konstruktor in diesem Slice**. Keine Zusage dieses Slice hängt an einem Zweig (Produktcode 0, `audit.New` von der Wurzel aus nicht auslösbar); ein Haken verlangte Änderungen an `export_test.go` oder `Bauteile` jenseits der Zusagen. Sie sind kein Fall von `zweig-ohne-lauf-unter-dem-detektor` (dort: nebenläufiger Test, der Detektor sieht nur, was läuft); die Zählung dieser Beobachtung (7×) wird nicht um sie erhöht. Adresse: eigener Eintrag im Register (Klasse »Fehlerzweig eines Konstruktors im Root ohne Lauf«), bei der Closure durch den Planner. | `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` | `docs/plan/planning/in-progress/slice-audit-verdrahtung.md` · »Ungedeckte Konstruktor-Fehlerzweige« (§6) | nein | Entscheidung, kein Befund |
| F-6 | INFO | Die vom Implementer gemeldete Grenze trägt: Ein Wrapper des Roots, der einen formfremden Wert in ein Feld setzt, schreibt nichts (der Adapter weist die Form ab), M2 bleibt grün, M1 und die Logzeile werden rot. Die drei formfremden Werte von M2 halten damit allein die Formprüfung des Adapters (deren Fänger im Adapter-Slice liegt). Die zweite Grenze (Aufzeichnung sitzt außen, ein Wrapper zwischen Dienst und Instanz fällt nur an der Datei auf) steht mit Grund und Folge im Plan. Beide sind akzeptierte Negative; die zugehörige Folge für `correlation_id` siehe F-1. | `AGENTS.md` §3.17 | `docs/plan/planning/in-progress/slice-audit-verdrahtung.md` · »Grenzen, die der Lauf zeigte« | nein | — |

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| Produktcode (`git diff 8667f9d..1693714 --stat`: zwei neue Testdateien unter `src/bootstrap/`, der Slice-Plan, `slice-bootstrap-m6.md`; `wurzel.go`, `export_test.go` und alle übrigen Pfade unverändert) | geprüft, ohne Befund: Produktcode 0 bestätigt |
| `src/bootstrap/audit_betrieb_hilfen_test.go` | geprüft; Fake `AuditAufzeichnung` hält jeden Eintrag ungeprüft fest und zählt jeden Aufruf, `SessionsMitFehlertext` hält jedes Argument jeder Operation fest (F-4 zur Auswertung); Canary aus Bytewerten (`[]byte{...}`), drei Formen plus Base64 des Paares; kein Befund außer F-4 |
| `src/bootstrap/audit_betrieb_test.go` | geprüft; Orakel unabhängig (Datei gelesen und geparst, Kennungen aus dem Säen, Feldtabelle als Literal); M2 zählt (6·7+1)·4 = 172 und prüft es; M1 bis M4 mit `len`-Prüfung; V01a bis V06 wie in der Tabelle; Befunde F-1 bis F-3 |
| Nebenläufigkeit | geprüft, ohne Befund: keine Goroutine im Slice; die Mutexe der Fakes schützen nur das Festhalten; keine Zusage über Nebenläufigkeit |
| Vorabprüfung `SPEC-139` Nr. 4 (`TestAuditBuildNimmtZugangsdaten`) | geprüft, ohne Befund: beide Modi, beide Klassen, URL gegen die Bytewerte zurückgeparst |
| Rohbytes (`AGENTS.md` §3.18) | geprüft, ohne Befund: `LC_ALL=C grep -cP` auf `\\u[0-9A-Fa-f]{4}`, BOM, NBSP in beiden Testdateien 0; Bytes ab 0x80 nur in Kommentaren, Meldungstexten und den `[]byte{...}` des Canary |
| Plan folgt Code (`AGENTS.md` §3.9) | geprüft, ohne Befund: Tabelle V01a bis V07, Stand der Fänger, Mutationen, Raten, Mengen und Abdeckung nennen die gelieferten Tests und Zahlen (172, 7, 6, 3, 4); die Marke in §4 nach `SPEC-128` Nr. 1 vom Planner gesetzt |
| Größenmarke | geprüft, ohne Befund: `git diff --numstat 8667f9d..1693714 -- src/bootstrap` = 392 + 829 = 1221, Marke 1900 |
| Sendung an `slice-bootstrap-m6` (Fänger gegen einen Tokenwert in `audit.jsonl`) | geprüft, ohne Befund: auf `main` geführt (`slice-bootstrap-m6.md` §1 *Übernimmt außerdem*, DoD Punkt 2, Prüfung der Sendungen, Geber `slice-audit-verdrahtung`) |
| Hard Rules (`AGENTS.md` §3.2, §3.7, §3.11) | geprüft, ohne Befund: keine `//nolint`-Direktive, Kommentare tragen Zusage oder Kopplung, die Kopfzeilen der Tests nennen, was sie fangen |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 3 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Folge einer Prüfgrenze nennt ein Orakel, das das Feld nicht prüft · Zählprüfung per Syntaxbaum erkennt eine Form des Aufrufs · Menge der Ablehnungsgründe ohne einen Fall · Fake zählt Operationen, der Test liest eine

## Verdikt

**Merge-blockierend:** ja — wegen F-1 (MEDIUM): Prüfgrenze und Orakel widersprechen sich im Feld `correlation_id`. F-2 bis F-4 blockieren nicht. Der Implementer entscheidet, ob Orakel oder Plan-Aussage angepasst wird; ein Fix deckt die Klasse »Felder der Zeile, deren Wert der Root beeinflusst« (`AGENTS.md` §3.17), nicht allein die genannte Mutation.

**Übergabe:** Findings gehen an den Implementer; die Finding-Klassen gehen zusätzlich in die Slice-Closure §7 und von dort in den Zähler. Der Report ersetzt keine Verifikation.
