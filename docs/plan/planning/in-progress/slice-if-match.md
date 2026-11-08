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

**Berührte Spec-Stellen:** `LH-FA-15.a` (`If-Match`/`If-None-Match` auswerten, Konflikt ergibt `412`, Versionserhöhung atomar) · `LH-FA-07.a` (Zustand bleibt bei Fehler unverändert) · `SPEC-049` (Zeilen Header `If-Match` und Fehler `412`, Spalte Soll) · `SPEC-108` Nr. 1 (`WithWorkspaceTransaction`) · `SPEC-109` (Fehlermodell; `ODR-E6010` für `412`) · `SPEC-283` (Vertrag der Vorbedingung, Architect 2026-11-08) · `SPEC-192` Nr. 2, 6 und 8, `SPEC-195` Nr. 10, `SPEC-230` Nr. 8, `SPEC-282`, `SPEC-280` Nr. 1 und 3, `SPEC-201` Nr. 3 e (Nachzüge) · `ARC-002`, `ARC-003`.
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

**Übernimmt (Fehlercode):** Der Architect hat den Code vor dem ersten Code-Commit ergänzt (2026-11-08): `ODR-E6010`, HTTP `412`, Zielbereich `mock`, in `SPEC-109` Nr. 2 und 5, im Katalog (§4 der Spezifikation) und im Vertrag `SPEC-283`. Der Slice liefert den Code in `src/hexagon/model/fehler.go`, seine Verwendung, die Zeile im Handbuch (Kataloge-Gate) und den Fänger; die Zahl der Codes steigt von 40 auf 41 (Zählstellen: §6).

**Übernimmt (Entscheidungen aus `slice-etag-versionen`, `SPEC-280` Nr. 1 und 3, Sendung des Architect vom 2026-11-08):** (1) Der ETag hat die schwache Form `W/"<n>"`; welcher Vergleich für `If-Match` gilt, legt der Architect dieses Slice fest (RFC 9110 §13.1.1 verlangt für `If-Match` den starken Vergleich, ein schwacher Wert träfe dann nie). (2) Bei `ETAG_ENABLED=false` bleiben `If-Match` und `If-None-Match` ohne Wirkung (ignoriert, kein `412`), es sei denn, der Vertrag dieses Slice entscheidet anders und zieht `SPEC-280` Nr. 3 mit einer neuen Nummer nach. (3) Der Satz zu `If-Match` in `SPEC-192` *Ungeprüft* und `SPEC-280` *Ungeprüft* fällt mit der Auswertung; bis dahin ändert ein `PATCH` oder `PUT` mit `If-Match` den Eintrag bedingungslos.

**Angenommen (Planner, 2026-11-08, Start; `AGENTS.md` §3.13):** Die drei Entscheidungen oben sind angenommen. Gelesen: §1 *Ausdrücklich NICHT* und die DoD dieses Slice schließen keine davon aus (Punkt 1 liegt in der Auswertung der DoD, Punkt 2 in der Zusatztabelle von §6 „mit und ohne aktivierte Unterstützung“, Punkt 3 im Doku-Update der DoD). Die Nachzüge in der Spezifikation sind Pflicht des Architect vor dem ersten Code-Commit: (a) der Vergleich für `If-Match` in der neuen Nummer (Auftrag 2), (b) bei „aus ohne Wirkung“ die Bestätigung von `SPEC-280` Nr. 3 zweiter Satz, sonst eine neue Nummer, die Nr. 3 nachzieht, (c) der Satz zu `If-Match` in `SPEC-192` *Ungeprüft* und `SPEC-280` *Ungeprüft* fällt, im selben Commit wie die Auswertung, die ihn ersetzt (`AGENTS.md` §3.9).

**Entschieden (Architect, 2026-11-08, Vertrag `SPEC-283`, zu den drei Entscheidungen oben):** (1) `If-Match` vergleicht schwach (RFC 9110 §8.8.3.2), ein `W/` im Request ist zulässig und ohne Wirkung; die Abweichung vom starken Vergleich der RFC (§13.1.1) steht mit Grund in `SPEC-283` Nr. 4. (2) Bei `ETAG_ENABLED=false` bleiben beide Header ohne Wirkung: `SPEC-280` Nr. 3 ist bestätigt (`SPEC-283` Nr. 1), eine neue Nummer dafür gibt es nicht. (3) Die Sätze zu `If-Match` in `SPEC-192`, `SPEC-280` und `SPEC-281` *Ungeprüft* sind in der Spezifikation schon gefallen, nicht erst mit dem Code der Auswertung: die Spezifikation führt (Greenfield), und `SPEC-283` stünde sonst neben einer Aussage, die ihm widerspricht; offen ist nur der Code. Die Tests, die das Ignorieren bei an festhalten (`TestEtagHeaderKeinEinfluss` in `src/hexagon/services/mock/etag_test.go`, der Fall `If-Match` in `TestPreferIgnoriert` in `schreiben_antwort_test.go`, soweit `Dienst.Etag` dort wahr ist), zieht der Slice mit der Auswertung nach (`AGENTS.md` §3.9). Ein unlesbarer Wert ist `412`, kein `400` (`SPEC-283` Nr. 3); `POST` wertet die Header nie aus (Nr. 1).

**Übernimmt (bedingt, Beispiel im Handbuch; Geber `slice-etag-versionen` über `slice-handbuch-mock`, dessen §1 die dritte Aussage der zweiten Sendung führt):** Trägt `slice-handbuch-mock` beim Start dieses Slice in `done/` ein gebundenes Beispiel, das `If-Match` oder `If-None-Match` als ohne Wirkung zeigt, ist es Ist-Zustand und fällt mit der Auswertung: dieser Slice ersetzt es im selben Commit durch ein Beispiel zum neuen Verhalten oder streicht es, das Kataloge-Gate und `make handbuch-check` bleiben grün. Liegt `slice-handbuch-mock` nach diesem Slice (die Reihenfolge der Kettenslices steht in der Roadmap), gibt es nichts zu ersetzen, und das Beispiel zu `ETAG_ENABLED` dort beschreibt das Verhalten nach der Auswertung. Das Beispiel zu `If-Match` selbst (`412` am Listener) ist Handbuch-Pflicht dieses Slice (DoD, Handbuch-Punkt), unter der Grenze der Anfragen am Listener, die der Architect in `slice-handbuch-mock` entscheidet (`SPEC-200` Nr. 6 d kennt weder Eingabe noch Listener); entscheidet er dort gegen Anfragen im Beispiel-Format, trägt dieser Slice die Aussage zu `412` nur als Zeile des Katalogs und als Test (siehe §6).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Bildung und Auslieferung des ETag und der Zähler — `slice-etag-versionen` (Voraussetzung); dieser Slice vergleicht, er bildet nicht.
- `Prefer: return=…` — `slice-prefer-return`; die beiden Slices teilen keinen Mechanismus.
- `If-Match` in einem Changeset von `$batch` — Berührung ohne Kante (`welle-etag` §5): wer von beiden später liegt, ergänzt den Fall in seinem Test. Liegt `welle-etag` zuerst in `done/`, trägt `slice-batch-changesets` den Fall (Sendung an `slice-batch-changesets`, dort in §1 angenommen); liegt `welle-expand-batch` zuerst, ergänzt dieser Slice den Fall „PATCH mit `If-Match` in einem Changeset“ als Zeile der Tabelle *Zusage · Fänger*.
- `412` im Modus `hybrid` ohne Fallback (`LH-FA-12.b`) — der Fall gehört dem Test der Welle, die später schließt (`welle-etag` §3); dieser Slice kennt weder `HybridUseCase` noch einen Upstream.
- Verteilte Versionskoordination über mehrere Instanzen — Out-of-Scope von `LH-FA-15`.
- `If-None-Match` und `If-Match` bei `GET` und `HEAD` (`304`) — gehört nicht zu `LH-FA-15` (Entscheidung des Architect, `SPEC-283` Nr. 10): das Lastenheft handelt von bedingten Änderungen und vom Konflikt `412`, `304` ist Cache-Validierung und änderte die Antwort jedes Lesens. Die Lücke steht benannt in der Spezifikation; einen Folge-Slice gibt es nicht, weil kein Anforderungstext ihn trägt.

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
| `src/hexagon/model/` | update | `VorbedingungErfuellt` (Grammatik und Vergleich, reine Funktion, neue Datei) und `CodeE6010` in `fehler.go` |
| `src/hexagon/services/mock/` | update | Zeilen der beiden Header lesen, Auswertung in `f` nach der Entscheidung und vor `Put` (`SPEC-283` Nr. 6) |
| `src/adapters/driving/httpodata/` | nur Tests | der Adapter reicht `r.Header` unverändert durch (`SPEC-136` Nr. 4) und bildet `412` über den `*inbound.Fehler` (`SPEC-136` Nr. 7 a); kein Produktcode erwartet, ein Fänger am Eingang (F27) |
| `src/bootstrap/` | neu (Test) | Ende-zu-Ende-Test `if_match_ende_zu_ende_test.go`; keine Verdrahtung, `Dienst.Etag` liegt seit `slice-etag-versionen` in der Wurzel, `konfig.go` und `wurzel.go` bleiben unberührt |
| `src/adapters/driven/audit/audit_test.go`, `src/adapters/driven/telemetry/regeln_test.go`, `src/hexagon/services/snapshot/fehler_test.go`, `src/hexagon/model/fehler_test.go`, `src/hexagon/services/mock/lesen_fehler_test.go` | update | Zählstellen der Codes (40 auf 41; 15 auf 16 erreichbare Codes des Mock), `nie()` in `snapshot/fehler_test.go`; ein weiterer Test, der rot wird, nennt sich selbst |
| `src/hexagon/services/hybrid/policy/entscheide_test.go`, `src/hexagon/services/hybrid/policy/total_test.go` | update (bedingt) | Siebente Zählstelle der Codes, Sendung von `slice-fallback-policy-entscheidung` (Verifikation vom 2026-11-08, Auflage 3): feste Zahlen 40, 43 (Nullwert, −1, N + 1), 62 608, 774 000 und die Zählung von Hand 1176 / 110 von 40 auf 41 Codes (44, neu berechnet), Stellen laut dem Plan des Gebers; nur wenn `slice-fallback-policy-entscheidung` vor diesem Slice nach `main` gelangt ist (Merge von `main` in diesen Zweig zeigt die Tests rot) |
| `src/hexagon/services/mock/rang_test.go`, `src/hexagon/services/mock/lesen_gleichzeitig_test.go`, `src/hexagon/services/mock/schreiben_weitere_test.go`, `src/hexagon/services/mock/etag_test.go` | update | weitere Zählstellen und der Nachzug des Ignorierens (Lauf von `make test`): `spec195Nr10` (16 Codes, `ODR-E6010` gehört nicht zu den elf Codes des Teils der Stufen), `zuordnung` in `TestZweigeGleichzeitig` (der Code kommt nur bei Mutationen vor), `mutationsCodes` (ein Lauf mit `ODR-E6010`), `TestEtagHeaderKeinEinfluss` (nur noch bei aus) |
| `src/hexagon/model/vorbedingung.go`, `src/hexagon/services/mock/schreiben.go` | neu, update | `VorbedingungErfuellt` mit `liesEtagListe` und `liesTag` (Grammatik, `SPEC-283` Nr. 3 bis 5); im Dienst ein Aufruf in `imWorkspace` nach `entscheide` und vor `Put`, `vorbedingungVerletzt`, Lesen der Zeilen vor der Transaktion |
| `docs/user/benutzerhandbuch.md` | update | Zeile `ODR-E6010` in der Tabelle *Fehlercodes* (Kataloge-Gate); Beispiel: §6 |
| `spec/spezifikation.md`, `spec/architecture.md`, `harness/vertraege/SPEC-201-handbuch-katalog.md` | update (Architect, erledigt) | `SPEC-283` und die Nachzüge; Nr. 3 e von `SPEC-201` nennt `ODR-E6010` unter den Codes in H |
| Tests | neu | Tabelle *Zusage · Fänger* unten, je Zeile ein Fänger mit Mutation |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) stehen seit dem Vertrag des Architect (2026-11-08) im Plan, vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer.

- Nebenläufig: die Zeilen der Tabelle zählen im Block (`SPEC-128`) als nebenläufig; „unter einer Sperre, ohne sie dazwischen freizugeben“ bekommt je Fenster der Sperre eine Zeile mit Fänger oder benanntem Negativ (`AGENTS.md` §3.10).

**Ergebnis des Architect** (2026-11-08; Ort: `SPEC-283`, die Nachzüge stehen im Kopf dieses Plans):

1. Fehlercode: `ODR-E6010`, `412`, Zielbereich `mock`, `message` "Precondition failed"; die Antwort trägt keinen ETag (Nr. 7).
2. Formen des Headers: Zeilen und Liste (Nr. 2), Grammatik mit `*`, `W/`, Leerraum und leeren Elementen (Nr. 3), schwacher Vergleich auf den Bytes des Tags (Nr. 4); ein unlesbarer Wert, ein leerer Wert und ein gesetzter Header ohne Element sind `ODR-E6010` (Nr. 3).
3. `If-None-Match` bei `PATCH`, `PUT`, `DELETE`: `*` bei vorhandenem Eintrag und ein Treffer sind `412` (Nr. 5); bei `GET` und `HEAD` gehört `304` nicht zu `LH-FA-15`, die Lücke steht in Nr. 10 und in §1.
4. Rangfolge: die Vorbedingung ist der letzte Verstoß (Nr. 6); `POST` wertet sie nie aus (Nr. 1).
5. Fenster: Lesen der Version, Vergleich und Schreiben in derselben Transaktion, ohne Port dazwischen (Nr. 6, 8).
6. Aus: beide Header ohne Wirkung (Nr. 1).

**Tabelle *Zusage · Fänger*** (`AGENTS.md` §3.15; Architect, 2026-11-08; die Zelle *Nr.* nennt den Ort in `SPEC-283`; Mengen sind Schleifen oder Generatoren aus dem Wortlaut der Zusage, `AGENTS.md` §3.17):

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| F1 · Nr. 5, 6 | Mit passendem `If-Match` gelingen `PATCH`, `PUT` und `DELETE` (`204`); `PATCH` und `PUT` tragen den ETag der Version *v* + 1, die Revision steigt um 1 | Test, Schleife Methode (3) × Version *v* aus {0, 1, 9, 10, 99, 100} (Baseline-Eintrag mit Version 0, mehrfach geänderte) × Form {`W/"v"`, `"v"`}; der Fake hält `Put` 1 und `Advance` 1 fest | Vergleich gegen die Version nach der Mutation statt vor ihr | andere |
| F2 · Nr. 5, 7 | Ein veralteter oder fremder Tag ergibt `ODR-E6010` mit `412`; Entity-State, Revision und Zähler bleiben; der Fake zählt `Transaktion` 1, `Workspaces.Get` 2 (einer vor der Transaktion, `SPEC-192` Nr. 8 c, einer in `f`), `Entities.Get` 1 und `Put`, `Advance`, `Coverage.Get`, Schreiber je 0 | Test, Schleife Methode × Tag {*v* − 1, *v* + 1, 0 bei *v* > 0, `"x"`, `W/"x"`} (beide Seiten der Grenze; der gleiche Wert ist F1) | `==` zu `<=` (jeder ältere Tag träfe) | andere |
| F3 · Nr. 5 a | `If-Match: *` lässt `PATCH`, `PUT` und `DELETE` an einem vorhandenen Eintrag zu, ohne Versionsvergleich | Test, Schleife Methode × Eintrag {`complete`, `partial`} (`PATCH` an `partial`: F14) | `*` als unlesbar behandelt | andere |
| F4 · Nr. 6 | Fehlender Eintrag und Tombstone bleiben `ODR-E6008` oder `ODR-E6003`, nie `ODR-E6010`, mit jedem Wert beider Header | Test, Schleife Methode × Zustand {Tombstone, fehlend mit Coverage `complete`, `partial`, `unknown`} × Header (2) × Wert {erfüllt, nicht erfüllt, `*`, unlesbar}; `Put` 0 | Auswertung vor der Entscheidung nach `SPEC-193` Nr. 2 | andere |
| F5 · Nr. 3, 5 | Eine Liste ist erfüllt, wenn ein Tag trifft, an jeder Position; sonst `ODR-E6010` | Test, Schleife Länge 1 bis 4 × Position des treffenden Tags (und keiner) × Methode × Header (2) | nur das erste Element verglichen | andere |
| F6 · Nr. 4 | Der Vergleich ist schwach und exakt auf den Bytes des Tags: `W/"v"` und `"v"` treffen; `"07"`, `"+7"`, `"7 "`, `"70"` für *v* = 7 und `"1"` für *v* = 10 treffen nicht | Test, Generator über die Schreibweisen von *v* (Präfix `W/`, führende Null, Vorzeichen, Ziffer vorn oder hinten, Leerzeichen) × *v* aus {7, 10, 100} | starker Vergleich (`W/` trifft nie) oder Präfixvergleich | andere |
| F7 · Nr. 3 | Wohlgeformte Werte werden gelesen: Leerzeichen und Tabulator um Elemente, leere Elemente, ein Komma im Tag (`"a,b"` bleibt ein Tag, ein treffendes Tag dahinter trifft) | Test, Generator: Elemente {treffendes Tag, `"a,b"`, leer} in jeder Reihenfolge der Länge 3 × Leerraum {keiner, Leerzeichen, Tabulator, gemischt} an beiden Rändern | Trennen an jedem Komma | andere |
| F8 · Nr. 3 | Ein unlesbarer Wert ist `ODR-E6010`, auch wenn ein anderes Element trifft | Test, Schleife über alle 256 Bytes an einer Stelle im Tag (gültig sind `0x21`, `0x23` bis `0x7E`, `0x80` bis `0xFF`) und über die Formen von Nr. 3 (leer, nur Leerraum und Kommas, ohne Anführungszeichen, nicht geschlossen, Rest hinter dem Tag, `W/` allein, `w/"7"`, Leerraum nach `W/`, `*` neben einem Element, `**`) × Header (2) | ein unlesbares Element wird übersprungen, oder `0x7F` im Tag erlaubt | andere |
| F9 · Nr. 2 | Mehrere Zeilen sind eine Liste, der Name zählt ohne Schreibweise; eine leere Zeile neben einer gültigen zählt nicht; nur leere Zeilen, oder `*` in einer und ein Tag in einer anderen Zeile, sind `ODR-E6010` | Test, Schleife Schreibweise {`If-Match`, `if-match`, `IF-MATCH`, `If-match`} × Zeile des treffenden Tags {1, 2, 3} × Header (2), dazu zwei Schreibweisen im selben Request in der Bytefolge der Namen | nur die erste Zeile oder nur der kanonische Schlüssel gelesen | andere |
| F10 · Nr. 2 | Nicht gesetzt (kein Schlüssel, `nil`, leere Wertliste) wirkt wie ohne Header; eine leere Zeile ist gesetzt und `ODR-E6010` | Test, Schleife Zustand der Map (4) × Header (2) × Methode | leere Wertliste als gesetzt | andere |
| F11 · Nr. 5 b | `If-None-Match`: `*` bei vorhandenem Eintrag und ein Treffer in der Liste sind `ODR-E6010`, ohne Treffer gelingt die Mutation | Test, Schleife Methode × {`*`, Liste der Länge 1 bis 3 mit Treffer an jeder Position, ohne Treffer} | `If-None-Match` wie `If-Match` gewertet | andere |
| F12 · Nr. 5 d | Mit beiden Headern gelingt die Mutation nur, wenn beide erfüllt sind | Test, Schleife Methode × {erfüllt, nicht erfüllt} für `If-Match` × dasselbe für `If-None-Match` | Verknüpfung mit oder statt und | andere |
| F13 · Nr. 6 | Die Vorbedingung geht jedem Verstoß vor der Transaktion nach: Methode und Ressource (`ODR-E6002`, V2-Dienst `ODR-E6006`, `$filter`), Workspace (`ODR-E7001`, `ODR-E5004`), Service und Modell (`ODR-E5005`, `ODR-E6004`, `ODR-E6008` des Sets), Schlüssel der URL (`ODR-E6005`), Body (`ODR-E6007`, `ODR-E6002`, `ODR-E9001`); der Fake zählt `Transaktion` 0 | Test, Schleife über diese Verstöße (jeder Code der Schritte (e) bis (j) von `SPEC-192` Nr. 2) × Methode × nicht erfüllter Header (2) | Auswertung vor dem Lesen des Bodys oder vor dem Workspace | andere |
| F14 · Nr. 6 | In der Transaktion gehen `ODR-E7001`, `ODR-E6003` (`PATCH` an `partial`) und `ODR-E6007` (Typ) der Vorbedingung vor; `PUT` und `DELETE` an `partial` werden verglichen | Test, Schleife über die drei Verstöße × nicht erfüllter Header; `PUT` und `DELETE` an `partial` mit erfülltem (`204`) und nicht erfülltem Wert (`412`) | Auswertung vor der Typprüfung oder vor der Prüfung von `PATCH` an `partial` | andere |
| F15 · Nr. 8 a | Von *n* gleichzeitigen `PATCH` mit demselben `If-Match` gelingt genau einer (`204`), die übrigen sind `ODR-E6010`; Version + 1, Revision + 1, `Put` 1 (Fenster Lesen der Version und Vergleich) | Test mit *n* = 20, echte SQLite-Datei und Fake mit Haltepunkt und Zähler an `Entities.Get`; Rate je Mutation und je Lauf von `go test` im Plan (`AGENTS.md` §3.10) | Version aus einem Lesen außerhalb der Transaktion | nebenläufig |
| F16 · Nr. 8 b | Dasselbe für `PUT` und für eine Mischung aus `PATCH` und `PUT` (Fenster Vergleich und Schreiben) | Test mit *n* = 20 wie F15, Mischung beider Methoden | Vergleich in einer, `Put` in einer zweiten Transaktion | nebenläufig |
| F17 · Nr. 8 c | Gleichzeitige `ODR-E6010` ändern Version, Revision und Zähler nie, und `Put` läuft so oft, wie Mutationen gelingen (Fenster Entscheidung `412` und Ende der Transaktion) | Test: *n* = 20 mit falschem Tag neben einer Mutation mit richtigem: ein `204`, 20 mal `ODR-E6010`, `Put` 1 und `Advance` 1 am Fake | `Put` vor der Auswertung (die Rücknahme der Transaktion verdeckte es in SQLite, der Zähler des Fake nicht) | nebenläufig |
| F18 · Nr. 8 d | Von *n* gleichzeitigen `DELETE` mit demselben `If-Match` gelingt genau einer, die übrigen sind `ODR-E6008`, nicht `ODR-E6010` | Test mit *n* = 20 | Auswertung vor der Prüfung auf den Tombstone | nebenläufig |
| F19 · Nr. 9 a | Nach `DELETE` und `POST` an derselben Adresse trifft der ETag vor dem `DELETE` nie; der ETag der Neuanlage trifft | Test, Folge `DELETE`, `POST`, dann `PATCH` und `PUT` mit dem alten und dem neuen ETag | die Version einer Neuanlage über einen Tombstone ist nicht + 1 | andere |
| F20 · Nr. 9 b | Dieselbe Anfrage ein zweites Mal ist `ODR-E6010`, auch nach `PATCH` mit `{}`; eine abgewiesene bleibt abgewiesen und ändert nichts | Test, Schleife Methode (`PATCH`, `PUT`) × Body {Änderung, `{}`, gleicher Inhalt} | Version bei `PATCH` mit `{}` nicht erhöht | andere |
| F21 · Nr. 7 | `ODR-E6010` trägt Status 412 (nur diesen), Zielbereich `mock`, feste `message`, keine Ursache und keinen `ETag`; `Error()` nennt keinen Wert der Header | Test, Schleife über Headerwerte mit unterscheidbarer Marke × Header (2); `StatusErlaubt` für jeden Status von 0 bis 599 | Wert des Headers im Text von `Error()` | andere |
| F22 · Nr. 7 | `CodeE6010` steht in `model.AlleCodes()` hinter `ODR-E6009`, und die Zahl der Codes ist 41 an jeder Zählstelle | `model/fehler_test.go` (Tabelle, Reihenfolge, Zahl) und die Tests von Audit, Telemetrie, Snapshot (`nie()`) und Mock (16 erreichbare Codes), je rot ohne den Code | Code gestrichen oder vor `ODR-E6009` | andere |
| F23 · Nr. 7 | Das Handbuch führt `ODR-E6010` (Kataloge-Gate) | Ausgabe von `make katalog-check`: rot ohne die Zeile, grün mit ihr | Zeile gestrichen | andere |
| F24 · Nr. 1 | Bei aus bleiben beide Header ohne Wirkung | Test, Schleife Methode (`PATCH`, `PUT`, `DELETE`) × Header (2) × Wert {treffend, veraltet, `*`, unlesbar, leer}: Antwort und Zustand wie ohne Header; `TestEtagHeaderKeinEinfluss` (heute auch für an) angepasst | Auswertung auch bei aus | andere |
| F25 · Nr. 1 | `POST` wertet die Header nie aus (`201`, bei vorhandenem Schlüssel `ODR-E6009`) | Test, Schleife Header (2) × Wert (5 wie F24) × Eintrag {fehlend, vorhanden} bei an | `POST` wertet `If-Match` | andere |
| F26 · Nr. 10 | `GET` und `HEAD` ignorieren die Header: kein `304`, kein `412`, die Antwort ist byte-gleich der ohne Header, bei an und bei aus | Test, Schleife {`GET`, `HEAD`} × {Entity, Collection, Anzahl} × Header (2) × Wert (5) × {aus, an} | `GET` antwortet `304` bei Treffer | andere |
| F27 · Nr. 7 | Der Header erreicht den Dienst unverändert (`r.Header`, `SPEC-136` Nr. 4), und `ODR-E6010` geht mit 412 und dem Fehlerobjekt hinaus (`SPEC-136` Nr. 7 a) | Test am Eingang (`httpodata` mit Fake-Dienst) und Ende-zu-Ende (`bootstrap`, `ETAG_ENABLED` an): `PATCH` mit veraltetem Tag ergibt `412` mit dem Code `ODR-E6010` im Fehlerobjekt, mit aktuellem `204` und neuem ETag | der Adapter filtert den Header oder bildet 412 auf 500 ab | andere |
| F28 · Nr. 7 | `ODR-E6010` löst im Modus `hybrid` keinen Fallback aus (`SPEC-261` Nr. 3 c) | akzeptiertes Negativ: dieser Slice kennt weder `HybridUseCase` noch einen Upstream; den Fall trägt der Test der Welle, die später schließt (§1). Folge: ein Hybrid-Dienst, der den Code als Auslöser führte, bliebe bis dahin ungefangen | – | andere |

Die Tabelle hat 28 Zeilen: *a* = 4 (F15 bis F18), *b* = 24 (F28 ist ein akzeptiertes Negativ und zählt als `andere`, `SPEC-128` Nr. 3 e). Die Raten der Fänger F15 bis F18 stehen im Absatz *Raten* unten (`AGENTS.md` §3.10).

**Bericht des Implementers** (Handoff, 2026-11-08; `AGENTS.md` §3.10, §3.15, §3.17; Messender der Mutationen: `tools/harness/mutationslauf.sh`, `SPEC-210`, Image `odata-recorder:race`, `--memory=4g --cpus=4`). 58 Mutationen, jede einmal rot gesehen, keine grün, keine mit Übersetzungsfehler; dazu zwei Läufe außerhalb des Werkzeugs (F23, Ende-zu-Ende über SQLite, unten). Die Namen sind die der Liste des Laufs.

| Zusage | Mutation | roter Test |
|---|---|---|
| F1 Vergleich gegen die Version vor der Mutation | `f1-version-nach-der-mutation` (`neu.Version` statt `es.Version`) | `TestIfMatchPasst` |
| F2 Grenze der Version, nichts geschrieben | `f2-kleiner-gleich-modell`, `f2-kleiner-gleich-dienst` (`<=` statt `==`); `f17-put-vor-auswertung-nicht-nebenlaeufig` (`Put` vor der Auswertung, am Zähler des Fake) | `TestVorbedingungVersionen`, `TestIfMatchVeraltet`, `TestIfMatchVeraltet` |
| F3 `If-Match: *` | `f3-stern-unlesbar-modell`, `f3-stern-unlesbar-dienst` | `TestVorbedingungGoldens`, `TestIfMatchStern` |
| F4 Rangfolge Tombstone, fehlend | `f4-auswertung-vor-entscheidung`, `f4-auswertung-vor-tombstone` | `TestIfMatchRangZustand` |
| F5 Liste, jede Position | `f5-nur-erstes-modell`, `f5-nur-erstes-dienst` | `TestVorbedingungListen`, `TestIfMatchListePosition` |
| F6 schwach, Bytes des Tags | `f6-starker-vergleich` | `TestVorbedingungVersionen` |
| F7 Leerraum, Komma im Tag, leere Elemente | `f7-tabulator`, `f7-komma-im-tag`, `f7-leeres-element` | `TestVorbedingungListen` |
| F8 unlesbar ist `ODR-E6010` | `f8-stern-neben-element`, `f8-w-klein`, `f8-rest-hinter-dem-tag` (`TestVorbedingungGoldens`); `f8-byte-0x21`, `f8-byte-0x7f`, `f8-byte-0x80`, `f8-w-ohne-schraegstrich` (`TestVorbedingungGrammatik`) | wie genannt |
| F9 Zeilen, Schreibweisen des Namens | `f9-nur-kanonischer-schluessel`, `f9-nur-erste-zeile` | `TestIfMatchZeilen` |
| F10 nicht gesetzt gegen leere Zeile | `f10-kein-element`, `f10-leere-liste-gesetzt-im`, `f10-leere-liste-gesetzt-inm` | `TestVorbedingungNichtGesetzt` |
| F11 `If-None-Match` | `f11-inm-wie-im-modell`, `f11-inm-wie-im-dienst` | `TestVorbedingungGoldens`, `TestIfMatchListePosition` |
| F12 beide Header | `f12-nur-if-match`, `f12-nur-if-none-match`, `f12-oder-statt-und` | `TestIfMatchBeide` |
| F13 vor der Transaktion | `f13-auswertung-vor-body`, `f13-auswertung-vor-workspace` | `TestIfMatchVorTransaktion` |
| F14 in der Transaktion | `f14-auswertung-vor-partial`, `f14-auswertung-vor-typ` | `TestIfMatchInTransaktion` |
| F15 Fenster Lesen und Vergleich | `f15-lesen-ausserhalb-patch` (Version aus einem Lesen außerhalb der Transaktion) | `TestIfMatchWettlaufPatch` |
| F16 Fenster Vergleich und Schreiben | `f15-lesen-ausserhalb-put`, `f16-zwei-transaktionen` (Vergleich in einer, `Put` in einer zweiten Transaktion) | `TestIfMatchWettlaufPut` |
| F17 Fenster Entscheidung und Ende | `f17-put-vor-auswertung` | `TestIfMatchWettlaufFalsche` |
| F18 `DELETE` gleichzeitig | `f18-auswertung-vor-tombstone` | `TestIfMatchWettlaufDelete` |
| F19 Neuanlage über Tombstone | `f19-neuanlage-version-eins` | `TestIfMatchNeuanlage` |
| F20 Wiederholung | `f20-patch-leer-ohne-version` (`PATCH {}` erhöht die Version nicht) | `TestIfMatchWiederholung` |
| F21 Fehlerwert | `f21-ursache-gesetzt` (`TestIfMatchVeraltet`); `f21-status-428-erlaubt`, `f21-message`, `f21-zielbereich` (`TestCodeTabelle`) | wie genannt |
| F22 Code in der Tabelle, 41 an jeder Zählstelle | `f22-code-gestrichen-model`, `f22-reihenfolge` (`TestCodeTabelle`); `f22-code-gestrichen-audit` (`TestErrorCode`), `-telemetrie` (`TestFehlercodeForm`), `-snapshot` (`TestRueckgabenSnapshot`), `-mock` (`TestRueckgabenMenge`) | wie genannt |
| F23 Handbuchzeile | Zeile `ODR-E6010` aus der Tabelle des Handbuchs gestrichen, in einer Kopie des Baums: `docker build --target katalog` endet mit `katalog: Fehlercodes: ODR-E6010 fehlt im Handbuch` (Exit 1); mit der Zeile grün (`make katalog-check`) | Ausgabe des Gates |
| F24 bei aus ohne Wirkung | `f24-auswertung-auch-bei-aus`, `f24-auswertung-auch-bei-aus-alt` | `TestIfMatchAus`, `TestEtagHeaderKeinEinfluss` |
| F25 `POST` liest die Header nie | `f25-post-wertet-aus` | `TestIfMatchPostIgnoriert` |
| F26 `GET`, `HEAD` ignorieren | `f26-get-wertet-if-match` | `TestIfMatchLesenIgnoriert` |
| F27 Header unverändert, 412 am Eingang | `f27-header-gefiltert`, `f27-412-wird-500` (httpodata); Ende-zu-Ende: siehe *Raten* | `TestIfMatchHeaderUnveraendert`, `TestIfMatchFehler412` |
| F28 `hybrid` | akzeptiertes Negativ (Plan, Zeile F28) | – |

**Raten** (F15 bis F18, `AGENTS.md` §3.10): `n_ohne` 20 Prozesse je Mutation (`-test.count=1`, ein Prozess je Lauf von `go test`), `n_detektor` 0, weil kein Datenrennen die Zusage bricht: der Dienst hält keinen Zustand (`SPEC-195` Nr. 12), der Fake der Transaktion ordnet die Läufe durch eine Sperre, und die Wirkung (ein Erfolg, Zahl der Transaktionen, `Entities.Get`, `Put`) zählt der Test am Haltepunkt am Eingang der Transaktion (`sperreAmEingang`), nicht der Detektor. Last des Hosts laut `last=` der Ergebniszeile 11,6 bei 4 CPUs des Containers. Ergebnis: `f15-lesen-ausserhalb-patch` 20 von 20 rot, `f15-lesen-ausserhalb-put` 20 von 20, `f16-zwei-transaktionen` 20 von 20, `f17-put-vor-auswertung` 20 von 20, `f18-auswertung-vor-tombstone` 20 von 20; ohne Mutation sind dieselben Tests in 20 von 20 Prozessen grün (Last 12,7 bei 4). Die Raten gelten als Größenordnung bei dieser Last: die Mutationen F16 (zwei Transaktionen) und F17 (`Put` vor der Auswertung) fängt der Zähler der Fake-Operationen auch ohne die Gleichzeitigkeit. **Ende-zu-Ende über die echte SQLite-Transaktion** (`TestIfMatchWettlaufSQLite`, ohne Haltepunkt, 20 gleichzeitige Anfragen je Form, 5 frische Aufbauten): die Mutation `f15` (Version außerhalb der Transaktion gelesen) ist in 20 von 20 Prozessen rot, ohne Mutation 20 von 20 grün (eigener Treiber: das Test-Binary von `src/bootstrap` einmal gebaut, 20 Prozesse; Auftrag des Planner: `src/bootstrap` scheitert an der Nullprobe-Grenze von `mutationslauf.sh`, deshalb der eigene Treiber; der Lauf des Werkzeugs dort wurde nicht versucht; Last 9,4 bis 12,7 bei 4).

**Klassen aus dem Wortlaut** (Schritt 14 von `implement-slice`): Generator · Orakel · Elemente, die der Test zählt.

| Menge | Generator | Orakel | Elemente |
|---|---|---|---|
| Wert eines Headers (Nr. 3) | `TestVorbedingungGrammatik`: 8 wohlgeformte Werte × jede Position × {Byte einfügen, ersetzen} × 256 Bytes, dazu jede Position gelöscht | regulärer Ausdruck der Grammatik über dem Wert (eine Rune je Byte), ohne Code des Modells | 30271 Werte (8366 lesbar, 21905 unlesbar) × 2 Header |
| Liste (Nr. 3, 5) | `TestVorbedingungListen`: 5 Elementarten (Treffer stark, schwach, `"a,b"`, leer, Nichttreffer) in jeder Folge der Länge 1 bis 3 × Leerraum 4 × 4 an den Rändern × {eine Zeile, eine Zeile je Element} | Bauart (mindestens ein Element, Treffer irgendwo) | 4960 Formen × 2 Header |
| Version (Nr. 4) | `TestVorbedingungVersionen`: `etagKlasse` (0 bis 300, Zehner- und Zweierpotenzen mit Nachbarn, 2^63-1, 4000 Stichproben) × Form {`W/`, stark} × {v, v-1, v+1} | Zeichenfolge der Dezimalzahl | 4196 Versionen, 8392 Formen |
| Schreibweise der Version (Nr. 4) | `TestVorbedingungSchreibweisen`: 7 Versionen × 19 bis 21 Schreibweisen (führende Null, Vorzeichen, Ziffer vorn oder hinten, Leerzeichen, Hex, andere Zahl) × Form | Gleichheit der Zeichenfolge | 282 (14 Treffer) |
| Stellen der Ports in `f` | `TestIfMatchVeraltet`, `-Passt`, `-RangZustand`, `-VorTransaktion`, `-InTransaktion`, `-ListePosition`, `-Zeilen`, `-NichtGesetzt`, `-Beide`, `-Aus`, `-PostIgnoriert`, `-LesenIgnoriert`, `-Stern` | Vertrag (Nr. 1 bis 10), Fake mit Zähler je Operation | 90, 72, 96, 72, 20, 84, 128, 24, 12, 30, 20, 80, 24 Läufe |

**Prüfgrenzen als akzeptierte Negative** (`AGENTS.md` §3.15): (1) Die Versionen sind unbeschränkt (2^63 Werte); geprüft sind 4196. Folge im Betrieb: ein Fehler, der nur bei einer nicht gezogenen Zahl auftritt, bliebe ungefangen; der Vergleich arbeitet auf der Zeichenfolge und rechnet nie mit der Zahl, so dass dafür keine Form bekannt ist. (2) Die Länge des Headerwerts ist unbeschränkt (`SPEC-283` *Ungeprüft*); die Grammatik ist an acht Werten mit einem Byte geändert geprüft, zwei gleichzeitige Änderungen und Werte mit mehr als vier Elementen sind es nicht (die Liste des Dienstes: Länge 1 bis 4, die des Modells: 1 bis 3). Folge: ein Fehler, der nur bei zwei Änderungen am Wert auftritt, bliebe ungefangen. (3) Der Zugriff auf die Header bei aus, bei `POST`, `GET` und `HEAD` ist geprüft über die Wirkung, nicht über den Zugriff (`SPEC-280` Nr. 3, dort der Grund).

**Nachtrag der Fix-Runde** (Implementer, 2026-11-08, nach dem Review; die Tabelle F1 bis F28 und die Marke bleiben, die folgende Tabelle trägt keine Spalte *Klasse* und zählt nicht in `SPEC-128`):

| Zusage | Fänger | Mutation | roter Test |
|---|---|---|---|
| F11 erweitert, Klasse Version 0 bis 2^63-1: `If-None-Match: *` ist bei vorhandenem Eintrag nicht erfüllt, auch am unveränderten Baseline-Eintrag (Version 0); `If-Match: *` ist bei jeder Version erfüllt (Nr. 5) | `TestVorbedingungSternBeiJederVersion` (`etagKlasse` mit 4196 Versionen × 5 Schreibweisen des Sterns = 20980 Elemente, dazu die Grenze 0/1/2) und `TestIfMatchSternBeiJederVersion` (Methode 3 × Version 10 {0, 1, 2, 3, 9, 10, 99, 100, 2^32, 2^63-2} × Header 2 = 60 Läufe) | `liste.stern && version > 0`, `> 1`, `>= 1` (die genannten), `== 1` und `!= 2` (nicht genannt: weder Befund noch Review) | beide Tests, je 5 Mutationen rot |
| F2 erweitert: ein nicht erfüllter Tag bei Version 0 ist `ODR-E6010` | `TestIfMatchVeraltet` mit Version 0 (jetzt 120 Läufe; der Tag `"0"` der Zeile F2 ist durch `"v+2"` ersetzt, denn bei Version 0 wäre er der Treffer) | `w-auswertung-nur-ab-version-1` (`es.Version > 0 &&` vor der Auswertung) | `TestIfMatchVeraltet` |
| F29 · Nr. 9 d: ein Fehler oder Kontextabbruch an `Entities.Get` geht der Auswertung vor, auch bei nicht erfülltem Header: Port-Fehler mit seinem Code (`ODR-E3001`), `context.Canceled` und `context.DeadlineExceeded` ohne Code und `errors.Is` gleich; `Put`, `Advance`, `Coverage.Get`, Schreiber je 0 | `TestIfMatchAbbruchEntitiesGet`: Methode 3 × Header 2 × Fehlerart 3 = 18 Läufe; der Fake hält die Argumente und zählt jede Operation | `m2-auswertung-vor-fehler-von-entities-get`, `m2-fehler-gilt-als-fehlend`, `m2-kontextfehler-wird-e9001` | `TestIfMatchAbbruchEntitiesGet` |

**Rate zu F15 am SQLite-Test (Messung des Reviewers):** die Mutation derselben Klasse "Version in einer eigenen Transaktion vor der Auswertung gelesen" ist am Test `TestIfMatchWettlaufSQLite` in 14 von 20 Prozessen rot (6 grün; Last 5,5 bis 9,7 bei 4 CPUs), am Mock-Test mit Haltepunkt in 20 von 20. Die 20 von 20 oben gelten für die Form "Lesen ohne Transaktion"; der SQLite-Lauf ist wahrscheinlich, der Haltepunkt deterministisch. Eine Größenordnung, keine feste Zahl.

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
- **Stand der Bedingungen beim Start** (Planner, 2026-11-08, `main` = `993ad943`): `slice-etag-versionen` liegt in `done/` (Merge `0a3e209f`), damit `welle-mock` (M3) und die Kanten `slice-mock-schreiben`, `slice-http-adapter` und `slice-bootstrap-m3`; der Bezugs-Pfad `ETag` (`SPEC-280`, `SPEC-281`) steht auf `main`. In `in-progress/` laufen parallel `slice-admin-lebenszyklus` (`src/adapters/driving/admin/`) und `slice-fallback-policy-entscheidung` (`src/hexagon/services/hybrid/policy/`); beide teilen keinen Pfad aus §3. **Gate-Kette (Nr. 7 der Übergangsregeln):** kein Slice der Kette in `in-progress/`; dieser Slice ändert `Makefile`, `harness/mk/`, `Dockerfile`, `harness/README.md`, `tools/gegenproben/` und den Stempel nicht (Pfade in §3: `src/hexagon/services/mock/`, `src/hexagon/model/`, `src/adapters/driving/httpodata/`, `spec/spezifikation.md`, Handbuch und Tests), Linie B, kein Kettenslice. **Vom Architect erledigt** (2026-11-08, im Zweig nach dem Start-Commit, Nr. 3 der Übergangsregeln): der Vertrag `SPEC-283` samt `ODR-E6010` (Fehlertabelle `SPEC-109`, Katalog §4), die Tabelle *Zusage · Fänger* in §3 und das Raster in §6; Linie B vergab `SPEC-283` aus `SPEC-260` bis `SPEC-299`. **Größenmarke gesetzt** (Planner, 2026-11-08, vor dem ersten Code-Commit): `basis`, `schaetzung` und `marke` stehen im Block (unten).
- **Größenmarke** (Teil des Architect, 2026-11-08, `SPEC-128` Nr. 1): Die Tabelle in §3 hat 28 Zeilen, *a* = 4, *b* = 24; erwartet *f* = 0 (die Fakes von `fakes_schreiben_test.go` bekommen Haltepunkt und Zähler, es entsteht kein neuer) und *p* = 200 (Modell etwa 110: Leser des Werts mit der Grammatik von `SPEC-283` Nr. 3 und der Vergleich; Dienst etwa 70: Zeilen der Header, Aufruf in `f`, Fehler; Code und Tabelle etwa 20). Rechnung nach `SPEC-128` Nr. 1: *S* = 85·4 + 80·24 + 200·0 + 200 = 2460; *M* = 100 · ⌈135·2460 / 10000⌉ = 100 · ⌈33,21⌉ = 3400. **Schnittfrage:** *S* ≥ 2223, also ja; der Architect beantwortet sie mit **nein** (Grund im Block unter `schnitt-grund`): alle 28 Zeilen prüfen einen Entscheidungspunkt; der Schnitt Grammatik und Vergleich im Modell gegen Vorbedingung in der Transaktion wäre ein Schnitt nach Schichten, und der Modellteil allein ist kein einzeln lieferbarer Slice. Der wertmäßig trennbare Teil, `If-None-Match`, teilt Leser und Vergleich und senkt *S* auf etwa 2280. Ein Vorgänger gleicher Größe (`slice-etag-versionen`: 24 Zeilen, *S* = 2190) endete bei 1999 Zeilen. Das Handbuch zählt nicht (`SPEC-128` Nr. 3 m: `docs/` darf kein Pfad nennen). Der Block trägt `pfad`, `produktcode`, `fakes` und `schnitt-grund` (Architect) und die Dateien der Zählstellen der Codes (§3).

**Größenmarke (Planner, 2026-11-08, `SPEC-128` Nr. 1, nachgerechnet):** *S* = 85·4 + 80·24 + 200·0 + 200 = 2460; *M* = 100 · ⌈135·2460 / 10000⌉ = 100 · ⌈33,21⌉ = 3400. `basis` ist `89f4747d`, der Commit des Architect, der den Vertrag abschließt (Vorfahr von `HEAD`; `main` = `784f6166` liegt davor; der Commit ändert keinen Pfad des Blocks, Messung 0). Die Marke liegt über 3000; `SPEC-128` kennt keine Obergrenze, die Schnittfrage (*S* ≥ 2223) verlangt einen `schnitt-grund`, und der steht im Block.
**Schnittentscheidung (Planner): kein Schnitt.** Geprüft, ob sich ein Teil ohne geteilten Mechanismus abtrennen lässt (`AGENTS.md` §3.14): (1) Teil `If-None-Match` (F11, F12): teilt Leser, Grammatik und Vergleich, die Zeilen F5 bis F10 führen beide Header als Schleife und fallen nicht weg; *S* sinkt höchstens auf etwa 2280 (26 Zeilen, *p* etwa 180), bleibt über 2223, der Rest ist wieder eine Schnittfrage. (2) Teil `If-Match` gegen `If-None-Match` als zwei Slices: derselbe Befund, der zweite Teil bestünde aus zwei Zeilen und einem Aufruf. (3) Teil Nebenläufigkeit (F15 bis F18, vier Zeilen zu 85): der Rest hätte *S* = 80·24 + 200 = 2120, unter der Grenze, aber der abgetrennte Teil besteht aus Tests an demselben Code (die Transaktion, in der gelesen, verglichen und geschrieben wird, `SPEC-283` Nr. 8): ein Teil ohne Produktcode und ohne Auslöser, und der Rest lieferte `LH-FA-15.a` (Versionserhöhung atomar) ohne Fänger, eine Zusage ohne Fänger (`AGENTS.md` §3.15, DoD-Punkt 2). (4) Schnitt nach Schichten (Modell gegen Dienst): kein einzeln lieferbarer Teil (Architect, oben). Ein Schnitt hat damit keinen Teil, der keinen Mechanismus mit dem Rest teilt; Grund im Block. Das Wagnis: *p* = 200 ist knapp (Leser mit Grammatik und Vergleich etwa 110, Dienst etwa 70, Code etwa 20); liegt es bei 300, steigt *S* auf 2560 und *M* auf 3500, ändert aber weder Schnittfrage noch Entscheidung, und die Marke ist nach dem ersten Code-Commit fest. Dagegen steht der Vorgänger (24 Zeilen, *S* = 2190, Ende 1999, Auslastung 0,67) und der Zuschlag 1,35 auf *S*; endet der Slice darüber, gilt `SPEC-128` Nr. 5 d (Eintrag mit Grund), keine Rückführung. Die Messung zählt jede Änderung an den Pfaden zwischen `basis` und `HEAD` (Nr. 10 g): ein Merge von `main` in den Zweig ist an diesen Pfaden bis jetzt leer, weil die Parallel-Slices sie nicht berühren (§6).

```groessenmarke
pfad: src/hexagon/model/
pfad: src/hexagon/services/mock/
pfad: src/adapters/driving/httpodata/
pfad: src/bootstrap/if_match_ende_zu_ende_test.go
pfad: src/adapters/driven/audit/audit_test.go
pfad: src/adapters/driven/telemetry/regeln_test.go
pfad: src/hexagon/services/snapshot/fehler_test.go
pfad: src/hexagon/services/hybrid/policy/entscheide_test.go
pfad: src/hexagon/services/hybrid/policy/total_test.go
produktcode: 200
fakes: 0
basis: 89f4747d
schaetzung: 2460
marke: 3400
schnitt-grund: Ein Entscheidungspunkt (eine reine Funktion im Modell, ein Aufruf in f, ein Code); der Schnitt entlang der Schichten liefert keinen einzeln lieferbaren Teil, der Modellteil ohne Dienst ist ein Code ohne Auslöser mit einer Handbuchzeile ohne Beleg; If-None-Match, der wertmäßig trennbare Teil, teilt Leser und Vergleich (die Zeilen F5 bis F10 führen beide Header als Schleife) und senkt S nur auf etwa 2280, weiter über 2223; die Fänger der Nebenläufigkeit (F15 bis F18) abzutrennen senkt S unter 2223, teilt aber die Transaktion, in der Vergleich und Schreiben laufen, und ließe die Zusage der Atomarität (LH-FA-15.a, SPEC-283 Nr. 8) ohne Fänger (AGENTS.md 3.15); ein Vorgänger gleicher Größe (slice-etag-versionen, 24 Zeilen, S 2190) endete bei 1999 Zeilen
```

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner nach `SPEC-128` setzt (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). Die Marke ist die des Planner im Block unter *Größenmarke* oben (Schätzung 2460, Marke 3400).
- `in-progress` → `open` (blockiert — Carveout?): eine Randform des Rasters in §6 hat keinen Ort in der Spezifikation (`AGENTS.md` §3.12: der Implementer hält an und gibt sie an den Architect zurück).

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

- Zusage „Vergleich und Schreiben ohne Freigabe der Sperre dazwischen“: der Detektor meldet nur, was läuft, und sieht kein Fenster bei lauter geordneten Zugriffen — **Ausgang:** offen; je Fenster der Sperre (`SPEC-283` Nr. 8: Lesen der Version und Vergleich F15, Vergleich und Schreiben F16, Entscheidung `412` und Ende der Transaktion F17, dazu `DELETE` F18) eine Zeile der Tabelle mit Wirkungs-Fänger, der zählt, was das Fenster zuließe (N gleichzeitige Änderungen mit demselben ETag, genau eine gelingt), mit Rate je Mutation und je Lauf von `go test` samt Zahl der Prozesse und Messendem (`AGENTS.md` §3.10; `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`, `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor`, Zähler in §8).
- `If-None-Match` bei `GET` (`304`) gehört nicht zu `LH-FA-15` (Entscheidung des Architect, `SPEC-283` Nr. 10, §1 *Ausdrücklich NICHT*) — **Ausgang:** offen bis zur Closure; Fänger F26 (kein `304`, Antwort wie ohne Header); die Lücke bleibt in der Spezifikation benannt, ein Folge-Slice entsteht nicht, weil kein Anforderungstext ihn trägt.
- Ein veralteter `If-Match` trifft eine Entity, die nach DELETE neu angelegt wurde (Folge der Wahl in `slice-etag-versionen`, Risiko 1 dort) — **Ausgang:** offen bis zur Closure; entschieden in `SPEC-283` Nr. 9 a (der alte ETag trifft nie), Fänger F19 über die Folge DELETE, POST, PATCH mit altem ETag.
- Fakes der Ports halten jedes Argument fest und zählen jede Operation (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`, Zähler in §8) — **Ausgang:** offen.
- Sendung aus `slice-etag-versionen` (drei Entscheidungen, §1 *Angenommen*) — **Ausgang:** offen bis zur Closure; vom Architect entschieden (§1 *Entschieden*, `SPEC-283` Nr. 1 und 4): schwacher Vergleich, bei aus ohne Wirkung, die Sätze in den Abschnitten *Ungeprüft* sind gefallen. Vom Slice bleibt das Nachziehen der Tests, die das Ignorieren bei an festhalten (F24).
- Beispiel im Handbuch (bedingt, §1 *Übernimmt*): `slice-handbuch-mock` liegt in `open/`; sein Beispiel zu `ETAG_ENABLED` (Lesen, Ändern, neuer ETag) sendet kein `If-Match` und bleibt von der Auswertung unberührt, und ein Beispiel zu `412` braucht Anfragen am Listener, die `SPEC-200` Nr. 6 d nicht kennt. Bewertung des Architect: **keine neue Sendung** an `slice-handbuch-mock`. Dieser Slice trägt `412` als Zeile des Katalogs (`ODR-E6010`, F23) und als Test (F27); zeigt ein Beispiel dort beim Merge `If-Match` oder `If-None-Match` als ohne Wirkung, ersetzt es dieser Slice (§1), und ein Beispiel zu `412` ergänzt der Architect von `slice-handbuch-mock`, wenn er Anfragen im Format zulässt. Akzeptiertes Negativ bis dahin: das Handbuch sagt nichts zum Verhalten von `If-Match` (Folge: ein Leser findet den Code, nicht den Weg dorthin) — **Ausgang:** offen bis zur Closure; Adresse: Planner beim Handoff (Stand von `slice-handbuch-mock` auf `main` prüfen). **Stand beim Handoff** (Implementer): `slice-handbuch-mock` liegt in `open/` (auf `main` und im Zweig), das Handbuch trägt kein Beispiel mit `If-Match` oder `If-None-Match`; geliefert sind die Zeile `ODR-E6010` in der Tabelle der Fehlercodes und der Satz zur Auswertung in der Zeile `ETAG_ENABLED`, ohne Beispiel als Datei.
- Größe: Tabelle mit 28 Zeilen, Schätzung 2460, Marke 3400, Schnittfrage vom Architect und vom Planner mit nein beantwortet (§4, vier Schnitte geprüft, keiner ohne geteilten Mechanismus); `basis` `89f4747d`, `schaetzung` 2460 und `marke` 3400 stehen im Block (Planner, vor dem ersten Code-Commit) — **Ausgang:** offen bis zur Closure; Adresse: Implementer (Messpunkt-Zeile `Handoff`, Planner `Ende`, `SPEC-128` Nr. 9); über der Marke steht ein Eintrag mit Grund (`SPEC-128` Nr. 5), keine Rückführung. **Messung beim Handoff** (Implementer; Summe aus Einfügungen und Löschungen von `git diff --numstat 89f4747d..HEAD` über die Pfade des Blocks, Tests eingeschlossen, 1458 + 20): 1478 von Marke 3400, Auslastung 0,43; von Hand gemessener Produktcode 126 (`src/hexagon/model/vorbedingung.go` 105, `src/hexagon/services/mock/schreiben.go` 19, `src/hexagon/model/fehler.go` 2) gegen die Erwartung 200. Messpunkt-Zeile: `Messpunkt slice-if-match.md: zeilen=28 nebenlaeufig=4 fakes=0 produktcode=126 schaetzung=2460 marke=3400 messung=1478`, Stand `Handoff`. Die Auslastung liegt unter 0,60: die Meldung an den Architect trägt der Planner bei der Closure (`SPEC-128` Nr. 9).
- Merge-Reihenfolge: `slice-fallback-policy-entscheidung` und `slice-admin-lebenszyklus` laufen parallel; dieser Slice teilt mit ihnen keinen Pfad aus §3, und `src/bootstrap/` berührt er nur mit einer neuen Testdatei (Architect, 2026-11-08: die Verdrahtung von `Dienst.Etag` liegt seit `slice-etag-versionen` in der Wurzel, `konfig.go` und `wurzel.go` bleiben unberührt). Gemeinsam ist allein `spec/spezifikation.md`: `SPEC-283` steht unmittelbar vor `SPEC-261`, das `slice-fallback-policy-entscheidung` ändert — **Ausgang:** offen; Adresse: Planner (Reihenfolge der Merges; wer später merged, nimmt `main` vor dem Handoff auf, ein Konflikt höchstens am Rand des Einschubs).
- Gate-Kette (Nr. 7 der Übergangsregeln): der Slice ändert weder `GATE_CHECKS` noch ein Gate — **Ausgang:** entfallen; kein Pfad der Gate-Kette in §3.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster steht unten; jede Zelle nennt ihren Ort in `SPEC-283` oder steht auf "gilt nicht" mit Grund, keine Zelle ist leer — **Ausgang:** offen bis zur Closure (der Reviewer löst das Raster mit je einer Mutation auf).
- **Angenommen (Planner, 2026-11-08, `AGENTS.md` §3.13; Geber `slice-fallback-policy-entscheidung`, Verifikation Auflage 3):** `src/hexagon/services/hybrid/policy/entscheide_test.go` und `total_test.go` halten feste Zahlen aus `model.AlleCodes()` (40, 43, 62 608, 774 000, 1176 / 110) und werden mit `ODR-E6010` rot. Gelesen: §1 *Ausdrücklich NICHT* und die DoD dieses Slice schließen die Sendung nicht aus (die Nachzüge der Zählstellen gehören zur Auswertung und zum Code der DoD, `make gates` grün). Wer zuletzt nach `main` geht, zieht die Zahlen nach; `slice-fallback-policy-entscheidung` ist weiter (Verifier fertig) und geht voraussichtlich zuerst, dann trägt dieser Slice die Tests als siebente Zählstelle (§3, bedingt; die Dateien stehen im Block); geht dieser Slice zuerst, trägt `slice-fallback-policy-entscheidung` sie. **Ausgang:** offen bis zum Merge des zweiten Slice; Adresse: Implementer dieses Slice (beim Merge von `main` oder beim ersten roten `make test`).
- Zählstellen der Codes (40 auf 41, `ODR-E6010`): `src/hexagon/model/fehler_test.go` (Tabelle, Reihenfolge, Zahl), `src/hexagon/services/mock/lesen_fehler_test.go` (die 15 erreichbaren Codes werden 16), `src/adapters/driven/audit/audit_test.go` (Liste der Codes), `src/adapters/driven/telemetry/regeln_test.go` (Liste der Texte), `src/hexagon/services/snapshot/fehler_test.go` (`nie()`); `SPEC-230` Nr. 8 und `SPEC-195` Nr. 10 hat der Architect nachgezogen, weitere Tests, die mit dem neuen Code rot werden, nennt der Lauf von `make test`. Hinweis des Planner: Weitere Tests lesen `model.AlleCodes()` (`src/hexagon/services/inspection/diagnose_test.go`, `src/adapters/driven/odata/jsoncodecv2/schreiber_vertrag_test.go`, `src/hexagon/services/mock/rang_test.go` mit der Untergrenze 40 der Fälle); sie hängen an der Länge der Quelle und sollen grün bleiben, ohne dass der Slice sie ändert — wird einer rot, ändert ihn der Implementer im selben Commit, und seine Datei liegt außerhalb der Pfade des Blocks (die Messung zählt sie dann nicht) — **Stand des Handoff:** rot wurden außer den fünf Dateien `src/hexagon/services/mock/rang_test.go` (`spec195Nr10` mit 16 Codes; `ODR-E6010` gehört nicht zu den elf Codes des Teils der Stufen), `src/hexagon/services/mock/lesen_gleichzeitig_test.go` (`zuordnung` in `TestZweigeGleichzeitig`) und `src/hexagon/services/mock/schreiben_weitere_test.go` (`TestRueckgabenMenge` verlangt `ODR-E6010` in den Läufen von `mutationsCodes`); alle liegen unter einem Pfad des Blocks, die Messung zählt sie; **Ausgang:** offen; Adresse: Implementer (ein Commit für Code und Plan, `AGENTS.md` §3.9).
- Zahl der Aufrufe von `Workspaces.Get` bei `ODR-E6010` (Randform im Wortlaut des Vertrags, vom Implementer im Test festgehalten, `AGENTS.md` §3.12): `SPEC-283` Nr. 6 und F2 nennen `Workspaces.Get` 1; der Dienst ruft ihn zweimal, einmal in Schritt (f) vor der Transaktion und einmal in `f` (`SPEC-192` Nr. 6 (1) und Nr. 8 c; das ist der Bestand der Mutationen, nicht neu). `TestIfMatchVeraltet` erwartet 2 und zählt `Transaktion` 1 und `Entities.Get` 1 wie der Vertrag — **Ausgang:** eingetreten; entschieden (Architect, 2026-11-08): die zwei Aufrufe sind Folge von `SPEC-192` Nr. 8 c, kein Fehler; `SPEC-283` Nr. 6 und Zeile F2 nennen jetzt `Workspaces.Get` 2, der Test stimmt.
- Rang von `ODR-E9001` aus der Bildung des neuen Eintrags (`legeAn`, `aendere` bei einem Payload, den der Dienst nie schreibt) gegen `ODR-E6010` (Zelle ohne Ort in `SPEC-283` Nr. 6, die nur „nach (3) und vor (4)“ sagt): der Dienst wertet in `imWorkspace` nach `entscheide` aus, also nach der Bildung des Eintrags; ein Programmfehler in `legeAn` oder `aendere` geht der Vorbedingung vor. Kein Test hält die Reihenfolge fest, kein Client erreicht den Fall; **Ausgang:** eingetreten; entschieden (Architect, 2026-11-08): der Programmfehler geht der Vorbedingung vor, `SPEC-283` Nr. 6 sagt es und führt es als akzeptiertes Negativ ohne Fänger (Grund: kein Client erreicht den Fall; Folge: ein Programmfehler dort bleibt bei veraltetem Tag als `ODR-E9001` sichtbar). Kein Code, kein Test.
**Raster *Operation × Argumentart*** (Architect, 2026-11-08; jede Zelle nennt ihren Ort in `SPEC-283` (Nr. N) oder steht auf "gilt nicht" mit Grund):

| Operation | Header fehlt | leerer Wert | `*` | veralteter oder fremder ETag | fehlerhaft geformt, Liste, doppelt | Abbruch, Wiederholung |
|---|---|---|---|---|---|---|
| PATCH mit `If-Match` | Nr. 2 (nicht gesetzt: Mutation wie ohne Header) | Nr. 3 (kein Element: unlesbar, `ODR-E6010`) | Nr. 5 a (erfüllt bei vorhandenem Eintrag; fehlender Eintrag oder Tombstone: Nr. 6) | Nr. 4, 5 a, 7 (`ODR-E6010`, Zustand unverändert) | Nr. 3 (Grammatik, Liste), Nr. 2 (mehrere Zeilen, Schreibweisen des Namens) | Nr. 9 b (Wiederholung), Nr. 9 d (Abbruch; Fänger F29) |
| PUT mit `If-Match` | Nr. 2 | Nr. 3 | Nr. 5 a | Nr. 4, 5 a, 7 | Nr. 3, Nr. 2 | Nr. 9 b, 9 d (Abbruch: Fänger F29) |
| DELETE mit `If-Match` | Nr. 2 | Nr. 3 | Nr. 5 a, Nr. 6 (Tombstone: `ODR-E6008`) | Nr. 4, 5 a, 7 | Nr. 3, Nr. 2 | Nr. 8 d (gleichzeitig), Nr. 9 d (Abbruch: Fänger F29) |
| PATCH, PUT, DELETE mit `If-None-Match` | Nr. 2 | Nr. 3 | Nr. 5 b (bei vorhandenem Eintrag nicht erfüllt: `ODR-E6010`) | Nr. 5 b (ohne Treffer erfüllt, mit Treffer `ODR-E6010`) | Nr. 3, Nr. 2 | Nr. 9 b, 9 d (Abbruch: Fänger F29) |
| POST mit `If-Match` oder `If-None-Match` | gilt nicht: `POST` liest die Header nie (Nr. 1) | gilt nicht (Nr. 1) | gilt nicht (Nr. 1) | gilt nicht (Nr. 1) | gilt nicht (Nr. 1) | gilt nicht (Nr. 1); die Wiederholung ist `ODR-E6009` (`SPEC-193` Nr. 2) |
| GET, HEAD mit `If-Match` oder `If-None-Match` | gilt nicht: Nr. 10 (ignoriert) | gilt nicht (Nr. 10) | gilt nicht (Nr. 10) | gilt nicht (Nr. 10) | gilt nicht (Nr. 10) | gilt nicht (Nr. 10) |

**Zusatztabelle** (je Paar von Verstößen am selben Schlüssel und je Modus eine Zeile; `AGENTS.md` §3.16):

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| `412` und `404` oder `422`: Tombstone, fehlend mit Coverage `complete`, `partial`, `unknown` | Nr. 6: die Mutation scheitert zuerst (`ODR-E6008`, `ODR-E6003`), mit jedem Wert beider Header | F4 |
| `412` und `ODR-E6003`: `PATCH` an `partial`; `PUT` und `DELETE` an `partial` | Nr. 6: `PATCH` ist `ODR-E6003`; `PUT` und `DELETE` werden verglichen | F14 |
| `412` und `400`: Body (`ODR-E6007`), Schlüssel der URL (`ODR-E6005`), Typ-Widerspruch (`ODR-E6007`) | Nr. 6: der Verstoß der Anfrage geht vor | F13, F14 |
| `412` und `409` (`ODR-E6009`) | gilt nicht: `ODR-E6009` gibt es nur bei `POST`, und `POST` liest die Header nie (Nr. 1) | F25 |
| `412` und `501` (`ODR-E6002`, `ODR-E6006`: Methode, Ressource, `$filter`, V2-Dienst) sowie `ODR-E7001`, `ODR-E5004`, `ODR-E5005`, `ODR-E6004`, `ODR-E6008` des Sets | Nr. 6: alle vor der Vorbedingung | F13 |
| `If-Match` und `If-None-Match` zugleich | Nr. 5 d: beide müssen erfüllt sein | F12 |
| ohne aktivierte Unterstützung (aus) | Nr. 1: beide Header ohne Wirkung (`SPEC-280` Nr. 3) | F24 |
| mit aktivierter Unterstützung (an) | Nr. 1 bis 9 | F1 bis F23, F25 bis F27 |
| Modus `hybrid` | Nr. 7: kein Fallback bei `ODR-E6010` (`SPEC-261` Nr. 3 c) | F28 (akzeptiertes Negativ) |
| Fehler oder Kontextabbruch an `Entities.Get` bei nicht erfülltem Header | Nr. 9 d: der Port- oder Kontextfehler geht der Auswertung vor (Port-Fehler mit seinem Code, Kontextfehler ohne Code), nichts wird geschrieben | F29 |
| Reset und Restore | Nr. 9 c: ein ETag der verlassenen Geschichte kann treffen (`SPEC-280` Nr. 4, akzeptiertes Negativ, Grund dort) | keiner, keine Folgepflicht |

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
