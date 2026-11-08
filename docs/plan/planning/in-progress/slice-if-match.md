# Slice slice-if-match: If-Match und If-None-Match bei PATCH, PUT und DELETE, 412 ohne Zustandsänderung

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-etag](../welle-etag.md).

**Bezug:** `LH-FA-15`, `LH-FA-07`, `LH-QA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-15.a` (`If-Match`/`If-None-Match` auswerten, Konflikt ergibt `412`, Versionserhöhung atomar) · `LH-FA-07.a` (Zustand bleibt bei Fehler unverändert) · `SPEC-049` (Zeilen Header `If-Match` und Fehler `412`, Spalte Soll) · `SPEC-108` Nr. 1 (`WithWorkspaceTransaction`) · `SPEC-109` (Fehlermodell; der Code für `412` kommt hinzu) · `ARC-002`, `ARC-003`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** pt9912 (Zweig `slice-if-match`, Worktree `/Development/odata-recorder-wt-if-match`; der Zweig entsteht nach dem Start-Commit von `main`)

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** PATCH, PUT und DELETE auf einer Entity werten `If-Match` und `If-None-Match` gegen den aktuellen ETag der Entity aus: ein passender Wert lässt die Änderung zu und macht einen neuen ETag sichtbar, ein veralteter oder fremder Wert ergibt `412` mit dem Fehlercode für Precondition-Fehler und lässt Zustand, Revision und Zähler der Entity unverändert. Vergleich und Schreiben laufen in derselben Workspace-Transaktion, so dass von zwei gleichzeitigen Änderungen mit demselben ETag genau eine gelingt (Versionserhöhung atomar, `LH-FA-15.a`). Ohne aktivierte ETag-Unterstützung behauptet der Mock keinen Schutz (Wahl des Architect in `slice-etag-versionen`, Auftrag 2 dort).

**Übernimmt (Fehlercode):** Die Fehlertabelle führt keinen Code für `412`; der Architect ergänzt ihn vor dem ersten Code-Commit dieses Slice in der Spezifikation (Sendung von `welle-etag`, §6 dort: Aufgabe des Architect, nicht Gegenstand der Slices). Der Slice liefert die Verwendung des Codes, den Katalogeintrag im Handbuch und den Fänger.

**Übernimmt (Entscheidungen aus `slice-etag-versionen`, `SPEC-280` Nr. 1 und 3, Sendung des Architect vom 2026-11-08):** (1) Der ETag hat die schwache Form `W/"<n>"`; welcher Vergleich für `If-Match` gilt, legt der Architect dieses Slice fest (RFC 9110 §13.1.1 verlangt für `If-Match` den starken Vergleich, ein schwacher Wert träfe dann nie). (2) Bei `ETAG_ENABLED=false` bleiben `If-Match` und `If-None-Match` ohne Wirkung (ignoriert, kein `412`), es sei denn, der Vertrag dieses Slice entscheidet anders und zieht `SPEC-280` Nr. 3 mit einer neuen Nummer nach. (3) Der Satz zu `If-Match` in `SPEC-192` *Ungeprüft* und `SPEC-280` *Ungeprüft* fällt mit der Auswertung; bis dahin ändert ein `PATCH` oder `PUT` mit `If-Match` den Eintrag bedingungslos.

**Angenommen (Planner, 2026-11-08, Start; `AGENTS.md` §3.13):** Die drei Entscheidungen oben sind angenommen. Gelesen: §1 *Ausdrücklich NICHT* und die DoD dieses Slice schließen keine davon aus (Punkt 1 liegt in der Auswertung der DoD, Punkt 2 in der Zusatztabelle von §6 „mit und ohne aktivierte Unterstützung“, Punkt 3 im Doku-Update der DoD). Die Nachzüge in der Spezifikation sind Pflicht des Architect vor dem ersten Code-Commit: (a) der Vergleich für `If-Match` in der neuen Nummer (Auftrag 2), (b) bei „aus ohne Wirkung“ die Bestätigung von `SPEC-280` Nr. 3 zweiter Satz, sonst eine neue Nummer, die Nr. 3 nachzieht, (c) der Satz zu `If-Match` in `SPEC-192` *Ungeprüft* und `SPEC-280` *Ungeprüft* fällt, im selben Commit wie die Auswertung, die ihn ersetzt (`AGENTS.md` §3.9).

**Übernimmt (bedingt, Beispiel im Handbuch; Geber `slice-etag-versionen` über `slice-handbuch-mock`, dessen §1 die dritte Aussage der zweiten Sendung führt):** Trägt `slice-handbuch-mock` beim Start dieses Slice in `done/` ein gebundenes Beispiel, das `If-Match` oder `If-None-Match` als ohne Wirkung zeigt, ist es Ist-Zustand und fällt mit der Auswertung: dieser Slice ersetzt es im selben Commit durch ein Beispiel zum neuen Verhalten oder streicht es, das Kataloge-Gate und `make handbuch-check` bleiben grün. Liegt `slice-handbuch-mock` nach diesem Slice (die Reihenfolge der Kettenslices steht in der Roadmap), gibt es nichts zu ersetzen, und das Beispiel zu `ETAG_ENABLED` dort beschreibt das Verhalten nach der Auswertung. Das Beispiel zu `If-Match` selbst (`412` am Listener) ist Handbuch-Pflicht dieses Slice (DoD, Handbuch-Punkt), unter der Grenze der Anfragen am Listener, die der Architect in `slice-handbuch-mock` entscheidet (`SPEC-200` Nr. 6 d kennt weder Eingabe noch Listener); entscheidet er dort gegen Anfragen im Beispiel-Format, trägt dieser Slice die Aussage zu `412` nur als Zeile des Katalogs und als Test (siehe §6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Bildung und Auslieferung des ETag und der Zähler — `slice-etag-versionen` (Voraussetzung); dieser Slice vergleicht, er bildet nicht.
- `Prefer: return=…` — `slice-prefer-return`; die beiden Slices teilen keinen Mechanismus.
- `If-Match` in einem Changeset von `$batch` — Berührung ohne Kante (`welle-etag` §5): wer von beiden später liegt, ergänzt den Fall in seinem Test. Liegt `welle-etag` zuerst in `done/`, trägt `slice-batch-changesets` den Fall (Sendung an `slice-batch-changesets`, dort in §1 angenommen); liegt `welle-expand-batch` zuerst, ergänzt dieser Slice den Fall „PATCH mit `If-Match` in einem Changeset“ als Zeile der Tabelle *Zusage · Fänger*.
- `412` im Modus `hybrid` ohne Fallback (`LH-FA-12.b`) — der Fall gehört dem Test der Welle, die später schließt (`welle-etag` §3); dieser Slice kennt weder `HybridUseCase` noch einen Upstream.
- Verteilte Versionskoordination über mehrere Instanzen — Out-of-Scope von `LH-FA-15`.

**Keine Mindestzahl.** Ein Slice mit *einem* echten Ausschluss ist besser als
einer mit vier erfundenen; die vier Klassen sind ein Suchraster, keine
Ausfüll-Liste. Suchreihenfolge: Was übernimmt ein **Folge-Slice** (mit
Kennung — und die Kennung muss den Punkt auch annehmen)? Was bleibt als
**Bestand** bewusst stehen (mit Begründung)? Was wäre ein **anderer Vorgang**?
Welche **Schicht** rührt der Slice nicht an?

Was hier steht, ist die Grenze, an der ein wachsender Slice sich messen lässt:
Wer später etwas mitnimmt, das hier ausgeschlossen war, hat den Plan
**geändert**, nicht nur ergänzt.

## 2. Definition of Done

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — **≤ 3 Liefer-Punkte**; mehr heißt: der Slice ist zu groß und
gehört zurück zur Zerlegung. Gezählt wird nur, was mit dem Umfang wächst — die
Gate-Läufe und die fünf Closure-Pflichten darunter zählen nicht mit.

- [ ] Auswertung: `If-Match` und `If-None-Match` bei PATCH, PUT und DELETE gegen den ETag der Entity (Formen des Headers, Vergleich und Rangfolge der Prüfungen nach dem Vertrag des Architect); passender Wert: Änderung übernommen, neuer ETag in der Antwort; veralteter, fremder oder fehlender Wert: `412`.
- [ ] Atomarität: Vergleich und Schreiben in derselben Workspace-Transaktion; bei `412` bleiben Entity, Zähler und Revision unverändert; von zwei gleichzeitigen Änderungen mit demselben ETag gelingt genau eine, die andere ergibt `412`.
- [ ] Fehlercode für `412` verwendet (der Eintrag in der Fehlertabelle ist Aufgabe des Architect vor dem Code) und das Handbuch führt ihn.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/services/mock/` | update | Vorbedingung in der Workspace-Transaktion der Mutationen |
| `src/hexagon/model/` | update | Typ der Vorbedingung (Liste von ETags, `*`) und Fehler für `412`, falls der Architect sie im Domain-Modell verlangt |
| `src/adapters/driving/httpodata/` | update | Lesen der Header `If-Match` und `If-None-Match` und Abbildung von `412` |
| `spec/spezifikation.md` | update | Vertrag mit Randformen und Eintrag in der Fehlertabelle (Architect) |
| Tests | neu | Fälle von `LH-FA-15` (Happy Path, Negative) je Methode; Wirkungs-Fänger für das Fenster zwischen Vergleich und Schreiben |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Nebenläufig: die Zeilen der Tabelle zählen im Block (`SPEC-128`) als nebenläufig; „unter einer Sperre, ohne sie dazwischen freizugeben“ bekommt je Fenster der Sperre eine Zeile mit Fänger oder benanntem Negativ (`AGENTS.md` §3.10).

**Auftrag an den Architect** (vor dem ersten Code-Commit; Ergebnis in der Spezifikation, Neues unter einer künftigen Nummer, die der Architect vergibt):

1. Fehlercode und HTTP-Status für den Precondition-Fehler samt Eintrag in der Fehlertabelle; ob die Antwort den aktuellen ETag trägt.
2. Formen des Headers: `*`, ein ETag, Liste, schwache und starke Schreibweise, Vergleichsregel, fehlerhaft geformter Wert (`400` oder `412`), leerer Wert, doppelter Header.
3. `If-None-Match` bei PATCH, PUT und DELETE: Bedeutung (etwa `*` auf vorhandener Entity), und ob `If-None-Match` bei GET (`304`) zu `LH-FA-15` gehört; gehört es dazu, wird der Punkt vor dem Code in §1 dieses Plans eingetragen (Adresse: dieser Slice, `AGENTS.md` §3.13), sonst steht er als Lücke in der Spezifikation.
4. Rangfolge: Vorbedingung gegen `404` (Entity fehlt), `400` (Body ungültig), `409`, `ODR-E6003` (partielle Entity) und gegen die Zuordnung des Workspace; `If-Match` bei POST.
5. Fenster zwischen Vergleich und Schreiben: welche Schritte in der Transaktion liegen und wo die Sperre gehalten wird.
6. Verhalten ohne aktivierte Unterstützung (Wahl aus `slice-etag-versionen`, Auftrag 2 dort).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; füllt der Architect mit dem Vertrag, bevor der Slice an den Reviewer übergeht):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII (etwa Header-Werte in den Fixtures) entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18); der Bericht nennt die gegengelesene Datei.
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Generatoren oder Schleifen aus dem Wortlaut der Zusage (Generator-Pflicht, `.claude/commands/implement-slice.md` Schritt 14; Stellen-Regel, Schritt 19; `AGENTS.md` §3.17), nie eine Liste der Fälle, die einfielen.
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9); der Stempel `harness/gegenproben.stempel` entsteht nur im Integrationsschritt, nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-etag-versionen` liegt in `done/` (und damit `welle-mock`, Trigger der Welle). Die Spezifikation trägt den Vertrag samt Randformen und den Code für `412`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

- **Code-Abhängigkeit, mit Grund:** `slice-etag-versionen` (der Vergleich nutzt die Bildung des ETag und den Zähler; ohne sie gibt es nichts zu vergleichen), `slice-mock-schreiben` (die Vorbedingung läuft in der Workspace-Transaktion der Mutationen, `WithWorkspaceTransaction`), `slice-http-adapter` (die Header erreichen den Use Case als Teil des Requests; der Adapter bildet sie unverändert ab). Nicht startbar vor `welle-mock`.
- **Binary und Handbuch:** der Modus `mock` erreicht das Binary mit `slice-bootstrap-m3`; Beispiel und Kataloge-Gate wie in `slice-etag-versionen`.
- **Was der Slice nicht braucht:** `slice-prefer-return`, `welle-hybrid` (das `412` im Modus `hybrid` ist ein Fall der später schließenden Welle), `welle-expand-batch`, `welle-odata-v2`, `welle-betrieb`, `welle-admin-api`.
- **Stand der Bedingungen beim Start** (Planner, 2026-11-08, `main` = `993ad943`): `slice-etag-versionen` liegt in `done/` (Merge `0a3e209f`), damit `welle-mock` (M3) und die Kanten `slice-mock-schreiben`, `slice-http-adapter` und `slice-bootstrap-m3`; der Bezugs-Pfad `ETag` (`SPEC-280`, `SPEC-281`) steht auf `main`. In `in-progress/` laufen parallel `slice-admin-lebenszyklus` (`src/adapters/driving/admin/`) und `slice-fallback-policy-entscheidung` (`src/hexagon/services/hybrid/policy/`); beide teilen keinen Pfad aus §3. **Gate-Kette (Nr. 7 der Übergangsregeln):** kein Slice der Kette in `in-progress/`; dieser Slice ändert `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` und den Stempel nicht (Pfade in §3: `src/hexagon/services/mock/`, `src/hexagon/model/`, `src/adapters/driving/httpodata/`, `spec/spezifikation.md`, Handbuch und Tests), Linie B, kein Kettenslice. **Noch offen vor dem ersten Code-Commit:** der Vertrag des Architect samt Fehlercode für `412` in der Spezifikation, die Tabelle *Zusage · Fänger* in §3 und das Raster in §6 (Nr. 3 der Übergangsregeln: der Architect arbeitet im Zweig nach dem Start-Commit). Kennungen: Linie B vergibt aus `SPEC-260` bis `SPEC-299` (zuletzt `SPEC-282` belegt), der Fehlercode kommt als neuer Code in der Fehlertabelle (`SPEC-109`) und im Katalog (`SPEC-201`).
- **Größenmarke** (Planner, 2026-11-08, `SPEC-128` Nr. 1, **vorläufig, Prosa, kein Block**): Die Tabelle *Zusage · Fänger* fehlt noch; der Block mit `pfad`, `produktcode`, `fakes`, `basis`, `schaetzung` und `marke` entsteht, wenn der Architect die Zeilen nach Klassen und die Produktcode-Erwartung nennt, und der Planner setzt die Marke vor dem ersten Code-Commit mit `make groessenmarke` (`SPEC-128` Nr. 1: *S* = 85·*a* + 80·*b* + 200·*f* + *p*; *M* = 100 · ⌈135·*S* / 10000⌉; Schnittfrage genau dann, wenn *S* ≥ 2223). Orientierung ohne Bindung: Raster 5 Zeilen × 6 Spalten und Zusatztabelle lassen 20 bis 30 Zeilen erwarten; bei 20 Zeilen (*a* = 3, *b* = 17, *p* = 280) wäre *S* = 1895 und *M* = 2600, bei 24 Zeilen (*a* = 4, *b* = 20, *p* = 300) *S* = 2240 und *M* = 3100 **mit Schnittfrage**. Die Schnittfrage ist damit wahrscheinlich: Beantwortet der Architect sie mit „ja“, wird **vor dem Code** geschnitten (`AGENTS.md` §3.14); ein Schnitt entlang der Frage „Grammatik und Vergleich des Headers als reine Funktion im Modell“ gegen „Vorbedingung in der Transaktion samt Adapter und Atomarität“ ist ein Vorschlag, den der Architect prüft (teilen die Teile einen Mechanismus, steht der Grund in §4). Das Format-Merkmal „fremde Bytes“ trifft nicht zu: der Header-Wert ist kurz, seine Grammatik ist Teil des Vertrags.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt (*Größenmarke* oben: vorläufig, Prosa).
- `in-progress` → `open` (blockiert — Carveout?): der Code für `412` oder die Rangfolge der Prüfungen sind nicht entschieden (siehe §6).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Zusage „Vergleich und Schreiben ohne Freigabe der Sperre dazwischen“: der Detektor meldet nur, was läuft, und sieht kein Fenster bei lauter geordneten Zugriffen — **Ausgang:** offen; je Fenster der Sperre (zwischen Lesen des ETag und Schreiben, zwischen Entscheidung `412` und Ende der Transaktion) eine Zeile der Tabelle mit Wirkungs-Fänger, der zählt, was das Fenster zuließe (N gleichzeitige Änderungen mit demselben ETag, genau eine gelingt), mit Rate je Mutation und je Lauf von `go test` samt Zahl der Prozesse und Messendem (`AGENTS.md` §3.10; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, Zähler in §8).
- `If-None-Match` bei GET (`304`) liegt nicht in `welle-etag`; ob es zu `LH-FA-15` gehört, steht nicht im Lastenheft — **Ausgang:** offen; Architect (Auftrag 3), bei Aufnahme Eintrag in §1.
- Ein veralteter `If-Match` trifft eine Entity, die nach DELETE neu angelegt wurde (Folge der Wahl in `slice-etag-versionen`, Risiko 1 dort) — **Ausgang:** offen; Fänger über die Folge DELETE, POST, PATCH mit altem ETag.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, Zähler in §8) — **Ausgang:** offen.
- Sendung aus `slice-etag-versionen` (drei Entscheidungen, §1 *Angenommen*): schwache Form mit Vergleich für `If-Match`, `If-Match` bei aus ohne Wirkung, Fall des Satzes in `SPEC-192` und `SPEC-280` *Ungeprüft* — **Ausgang:** offen; angenommen vom Planner beim Start; Adresse: Architect (Vergleichsregel, Bestätigung oder Nachzug von `SPEC-280` Nr. 3, Streichen der Sätze), vor dem ersten Code-Commit.
- Beispiel im Handbuch (bedingt, §1 *Übernimmt*): ob `slice-handbuch-mock` beim Merge dieses Slice ein Beispiel trägt, das das Ignorieren zeigt, und ob Anfragen im Beispiel-Format zulässig sind — **Ausgang:** offen; Adresse: Planner beim Handoff (prüft den Stand von `slice-handbuch-mock` auf `main`) und Architect von `slice-handbuch-mock` (Grenze der Anfragen am Listener, dort §6).
- Größe: vorläufige Prosa-Marke, Schnittfrage wahrscheinlich (§4) — **Ausgang:** offen; Adresse: Architect (Zeilen nach Klassen, Produktcode-Erwartung, Schnittvorschlag vor dem Code), danach Planner (Block und Marke vor dem ersten Code-Commit).
- Merge-Reihenfolge: `slice-fallback-policy-entscheidung` und `slice-admin-lebenszyklus` laufen parallel; dieser Slice teilt mit ihnen keinen Pfad aus §3, aber `src/bootstrap/` (`konfig.go`, `wurzel.go`), falls der Vertrag eine Verdrahtung verlangt — **Ausgang:** offen; Adresse: Planner (Reihenfolge der Merges; wer später merged, nimmt `main` vor dem Handoff auf).
- Gate-Kette (Nr. 7 der Übergangsregeln): der Slice ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette in §3.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

**Raster *Operation × Argumentart*** (Zellen füllt der Architect; Stand: alle `(Vertrag, offen)`):

| Operation | Header fehlt | leerer Wert | `*` | veralteter oder fremder ETag | fehlerhaft geformt, Liste, doppelt | Abbruch, Wiederholung |
|---|---|---|---|---|---|---|
| PATCH mit `If-Match` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| PUT mit `If-Match` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| DELETE mit `If-Match` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| PATCH, PUT, DELETE mit `If-None-Match` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| POST mit `If-Match` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

**Zusatztabelle** (je Paar von Verstößen am selben Schlüssel und je Modus eine Zeile; `AGENTS.md` §3.16): `412` und `404`, `412` und `400`, `412` und `409`, `If-Match` und `If-None-Match` zugleich, mit und ohne aktivierte Unterstützung.

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| (Vertrag, offen) | (Vertrag, offen) | (offen) |

## 7. Closure-Notiz

Regeln dieser Sektion: Baseline-Regelwerk `modul-06-roadmap.md`
§Das Beobachtungs-Register (vorhandene `BEO-<KUERZEL>/<slug>` **zitieren** statt neu
formulieren — sonst zählt das Register zwei Namen getrennt) ·
`grundlagen-traceability.md` §Herkunfts-Anker für Steering-Loop-Regeln (das
Feld `liegt in` steht **nur**, wenn mit diesem Slice wirklich etwas verkörpert
wurde; Feld und Zielort auf **einer** Zeile, Sektionsangabe innerhalb der
Backticks). Ging der Gegenstand an einen anderen Slice oder entfiel er, trägt
diese Sektion die Zeile `Gegenstand:` mit Kennung oder Grund und jedes Risiko
aus §6 seinen Ausgang; die Liefer-Punkte der DoD bleiben leer
(`modul-05-planning-harness.md` §Ein Slice, dessen Gegenstand ein anderer
übernimmt).

*Wird bei der Closure gefüllt.*

## 8. Sub-Area-Prüfungen und Modus-Begründung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Sub-Area-Modus-Begründung — dort die **zwei vorgelagerten
Schritte** (sie stehen in jedem Slice-Plan, unabhängig von Modus und
Slice-Typ) und die **vier Pflichtkriterien** (Konventionen-Dichte ·
Phase-Reife · Evidenz-/Diskrepanz-Risiko · Reconciliation-Aufwand), vier und
nicht mehr.

**Der Abschnitt selbst entfällt nie.** Die zwei vorgelagerten Prüfungen laufen
in **jedem** Slice-Plan — sie hängen weder am Modus noch am Slice-Typ. Bedingt
ist allein der Modus-Begründungsblock am Ende; deshalb nennt der Titel beide
Hälften.

**Vorgelagert — Sub-Area-Wahl prüfen:** Berührt ist allein der Default `*` aus der Modus-Deklaration in `harness/conventions.md`; es gibt keine feinere Sub-Area, die zu prüfen wäre.

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30) und beim Start nachgezogen (Planner, 2026-11-08; Zähler sind die Dateien unter `evidence/`). Der Slice ist nebenläufig (Vergleich und Schreiben): `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last` (8×, verkörpert in `AGENTS.md` §3.10) gilt für die Rate des Wirkungs-Fängers; `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (16×, offen, Schwelle erreicht, Vorlage für den Reviewer-Skill beim User) für Zweige, die kein Lauf ausführt (Detektor-Zeile je Fenster, §6); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (4×, verkörpert in `implement-slice.md` Schritt 14 und im Reviewer-Skill) für die Fakes; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (17×, geplant in `slice-guard-schreibwege`; die Betriebsregeln in §3 gelten bis dahin). Verkörpert: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (`AGENTS.md` §3.10, Sensor `make test-race`). Die alten Zähler (2×, 1×, 2×, 1×) sind überholt; der Slice schreibt keine Zähler, er trägt Evidenz-Dateien bei der Closure bei.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
