# Slice slice-v2-import-codecs: Der Import nimmt den Satz der Codecs je Dienst entgegen

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-odata-v2](../welle-odata-v2.md).

**Bezug:** `LH-FA-17`, `LH-FA-18`, [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)

**Berührte Spec-Stellen:** `SPEC-127` Nr. 1 (`baselineimport.Dienst.Codecs`, Prüfung und Wahl) · `SPEC-175` Nr. 3 und 4 (Satz der Implementierungen, Import je Version) · `ARC-002`, `ARC-004`, `ARC-012`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-11-01.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `outbound.CodecSatz` (`Version`, `Metadaten`, `Daten`, `NurLesend`) liegt im Port-Paket, `baselineimport.Dienst` trägt das Feld `Codecs *outbound.CodecSatz`, und der Importer nimmt je Aufrufstelle (`ParseMetadata`, `ParseCollection`, `ParseEntity`) den Codec des Dienstes, sonst den der Konfig (`SPEC-127` Nr. 1, `SPEC-175` Nr. 3 und 4). Die Tests des Imports für Version 4 laufen textgleich grün. Der Import ist damit für beide Versionen derselbe Dienst; sein Versionsunterschied liegt allein im Satz seines Dienstes ([ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md)).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Implementierungen für V2 (Metadaten-Leser, Leser und Schreiber der Antworten, URL-Parser) — `slice-v2-metadaten`, `slice-v2-payload-lesen`, `slice-v2-payload-schreiben`, `slice-v2-query`; dieser Slice ist mit Fakes testbar und braucht keine.
- Die Felder `URL` und `Schreiber` des Satzes — die Slices, die ihre Schnittstellen liefern (`slice-v2-query`, `slice-odata-json-schreiben`), fügen sie hinzu; hier gibt es keine Schnittstelle, die sie tragen könnte.
- Die Tabelle Version → Satz im Composition Root und die Konfiguration der Version — `slice-v2-bootstrap`.
- Der Mock — `slice-mock-lesen` (nimmt `Codecs` je Dienst von Anfang an entgegen, Sendung von `slice-v2-bootstrap`).

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

- [x] `Dienst.Codecs` mit der Prüfung nach `SPEC-127` Nr. 1 (`nil` ist die Codecs der Konfig; `Metadaten` oder `Daten` `nil` ist `ErrUngueltig`) und der Wahl je Aufrufstelle; die Tests des Imports für Version 4 sind textgleich grün.
- [x] Fänger: ein Fake-Satz mit zwei Diensten verschiedener Sätze; je Aufrufstelle gehen `ctx`, `grenze`, Modell und Strom unverändert an den Codec des Dienstes und kein Aufruf an den anderen (Menge: die drei Aufrufstellen mal zwei Dienste, als Schleife); `Codecs` `nil` nimmt die Codecs der Konfig; ein Satz mit `Metadaten` oder `Daten` `nil` (auch der leere Satz) ist `ErrUngueltig` (berichtigt nach dem Architect-Ausgang: `Codecs` `nil` ist gültig, `SPEC-127` Nr. 1).
- [x] Handoff nach `AGENTS.md` §3.9 (ein Commit für Code und Plan), Tabelle *Zusage · Fänger* und Absatz der Mutationen in §3.
- [x] `make gates` grün.
- [x] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [x] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [x] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [x] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [x] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [x] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — im Repo **ohne** Wellen-Betrieb hier geprüft, im Repo **mit** Wellen von der nächsten Welle-Closure (auch für Slices ohne Wellen-Zugehörigkeit).

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/outbound/` | update | `CodecSatz{Version, Metadaten, Daten, NurLesend}` (`SPEC-175` Nr. 3); `URL` und `Schreiber` fügen die Slices hinzu, die ihre Schnittstellen liefern |
| `src/hexagon/services/baselineimport/` | update | `Dienst.Codecs`, Prüfung in `New` (`SPEC-127` Nr. 1), Wahl des Codecs je Aufrufstelle |
| Tests | neu / update | Fake-Satz mit zwei Diensten; Schleife über Aufrufstellen mal Dienste; die Tests für Version 4 bleiben textgleich |

Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15), vom Architect beim Start vor dem ersten Code-Commit. Eine Zeile nennt Zusage und Menge; die Fänger sind Schleifen über die Menge (§3.17), nie eine Liste der Fälle, die einfielen. „Die fünf anderen“ sind die Fakes der zwei übrigen Dienste (je Meta und Daten) und die der Konfig. Das Raster steht in §6, die Größenmarke in §4 (`SPEC-128`).

| Nr. | Zusage | Fänger | Mutation | Klasse |
|---|---|---|---|---|
| Z1 | Wahl je Aufrufstelle (`SPEC-127` Nr. 1, 7 f, 9; `SPEC-175` Nr. 3): `Codecs` `nil` nimmt die Codecs der Konfig, ein Satz die seines Dienstes. `ctx` (dasselbe Objekt), `grenze`, Modell (das `ParseMetadata` desselben Dienstes lieferte), `entitySet` und Strom (dasselbe Objekt, ungehüllt) gehen an den Codec des Dienstes, kein Aufruf an die fünf anderen | `TestSatzWahlJeAufrufstelle`: Schleife über die drei Aufrufstellen (`ParseMetadata`, `ParseCollection`, `ParseEntity`) mal die drei Dienste D0 (`Codecs` `nil`), D1 (Satz S1), D2 (Satz S2) = 9 Fälle. Die Modelle, Vorgaben und Kennungen der drei sind verschieden. Je Fall: Argumente des einen Fakes festgehalten (`ctx` mit Wertschlüssel, `grenze` gleich `MaxCaptureBody`, Modell gleich dem von `ParseMetadata` desselben Dienstes, Strom `*strom`), Zähler jeder Operation der fünf anderen gleich 0, auch der ungerufenen | (a) `ParseEntity` nimmt `i.k.DatenCodec` statt des Satzes; (b) `ParseMetadata` nimmt `i.k.MetadatenCodec`; (c) der Satz kommt von `Dienste[0]` statt von `a.ServiceID`; (d) bei einem Dienst mit Satz wird `ctx` durch `context.Background()` ersetzt; (e) bei einem Dienst mit Satz geht ein Nullmodell an die Antwort-Codecs; (f) der Strom wird nur bei einem Dienst mit Satz gehüllt (`io.NopCloser`, `bufio`) | andere |
| Z2 | Fehler des Satz-Codecs an jeder Aufrufstelle (`SPEC-127` Nr. 2, 7 f, 9, 17): der Verstoß wird nach Nr. 17 abgebildet, ein Kontextfehler kehrt unverändert und ohne Code zurück (auch verpackt), danach kein weiterer Port außer `Close` genau einmal | `TestSatzFehlerJeAufrufstelle`: Schleife über die drei Aufrufstellen mal D0, D1 und D2 (D0 trägt die Codecs der Konfig) mal Art {Verstoß des Codecs, verpackter Kontextfehler} = 18 Fälle; je Fall der Code bzw. der Kontextfehler, `Close` genau einmal, Zähler von `Blobs.Begin` und `Imports.Publish` gleich 0 und die Zähler aller neun Aufrufstellen (drei Aufrufstellen mal drei Dienste, die Codecs der Konfig als D0): 1 für die bis zur Fehlerstelle erreichten Aufrufstellen des Ziels, sonst 0 | (a) `ParseEntity` verwirft bei einem Dienst mit Satz den Fehler (`_ =`); (b) bei einem Dienst mit Satz wird aus dem Kontextfehler `ErrUngueltig`; (c) `ParseMetadata` schließt bei einem Dienst mit Satz den Strom nicht | andere |
| Z3 | `New` weist einen Satz mit `Metadaten` oder `Daten` `nil` ab, an jeder Stelle der Liste (`SPEC-127` Nr. 1): `model.ErrUngueltig` ohne Code, kein Dienst; ein gültiger Nachbar ändert nichts | `TestNewSatzTeilNil`: Schleife über Position {erster, zweiter, letzter von drei Diensten} mal Form {`Metadaten` `nil`, `Daten` `nil`, beide `nil` als `&outbound.CodecSatz{}`} = 9 Fälle, dazu der Fall „Satz teil-`nil` und leere `ID` am selben Dienst“ (welcher Verstoß, ist nicht zugesagt; beide sind `ErrUngueltig`): `errors.Is(err, model.ErrUngueltig)`, `ohneCode`, Rückgabe `nil` | (a) die Prüfung läuft nur für `Dienste[0]`; (b) `\|\|` statt `&&` (nur der leere Satz fällt auf); (c) nur `Metadaten` wird geprüft; (d) die Prüfung läuft nur für Dienste vor dem ersten gültigen Satz | andere |
| Z4 | Die Codecs der Konfig bleiben Pflicht, auch wenn jeder Dienst einen Satz trägt (`SPEC-127` Nr. 1: nil in `MetadatenCodec` oder `DatenCodec`) | `TestNewKonfigCodecsPflicht`: Schleife über Feld {`MetadatenCodec`, `DatenCodec`} mal Dienste {alle ohne Satz, alle mit Satz, gemischt} = 6 Fälle: `ErrUngueltig` ohne Code. Die Fälle „alle ohne Satz“ liegen im Bestand (`konfig_test.go`); der Fänger führt sie in der Schleife mit | die Prüfung der Konfig-Codecs entfällt, sobald jeder Dienst einen Satz trägt | andere |
| Z5 | Der Dienst liest allein `Metadaten` und `Daten` des Satzes; `Version`, `NurLesend` und jedes weitere Feld ändern weder `New` noch `Import`; ein Satz darf von mehreren Diensten geteilt werden (`SPEC-127` Nr. 1 nach dem Nachzug L1; `SPEC-175` Nr. 3, 6). Die Version des Dokuments prüft der Metadatenleser des Satzes (`SPEC-171` Nr. 2), nicht der Dienst | `TestSatzUebrigeFelder`: Schleife über `Version` {leer, `4.0`, `2.0`, `x`} mal `NurLesend` {falsch, wahr} = 8 Fälle (das Modell des Fakes trägt Version `4.0`, also gleicht keine Fixture dem Wert der Mutation): `New` und `Import` gelingen, `ReportJSON` byte-gleich dem Ergebnis desselben Imports über D0; dazu ein Fall mit demselben Satz-Zeiger an zwei Diensten (beide Imports gelingen) | (a) der Import vergleicht `Satz.Version` mit `Modell.Version()`; (b) `New` weist `Version` leer ab; (c) `NurLesend` wahr macht `Import` zu `ErrUngueltig`; (d) `New` weist einen von zwei Diensten mit demselben Satz-Zeiger ab | andere |
| Z6 | Die Normierung der Dienste behält `Codecs`, und `New` hält keinen Zustand mit der Konfig des Aufrufers: ein Satz, den `normiereDienste` verlöre, oder eine Änderung der Liste nach `New` änderte die Wahl (`SPEC-127` Nr. 1, 18); eine doppelte `ID` bleibt ungültig, auch mit verschiedenen Sätzen | `TestNormierungBehaeltSatz`: Schleife über Dienst {D1, D2} {Import nach `New` geht an den Satz}; dann `k.Dienste[i].Codecs = nil` und `k.Dienste[i].ID` ändern nach `New` und derselbe Import geht weiter an den Satz; Fall zwei Dienste mit gleicher `ID` und verschiedenen Sätzen: `ErrUngueltig` | (a) `normiereDienste` baut `Dienst{ID, UpstreamURL}` ohne `Codecs`; (b) `normiereDienste` gibt `in` zurück statt einer Kopie; (c) die Prüfung auf doppelte `ID` vergleicht den ganzen `Dienst` (`==`) | andere |
| Z7 | Gleichzeitige Imports verschiedener Dienste bleiben bei ihrem Satz, in jeder Verschränkung (`SPEC-127` Nr. 18: alle Operationen sicher, kein veränderlicher Zustand; `SPEC-175` Nr. 6: kein Satz teilt Zustand) | `TestSatzGleichzeitig` (Wirkungs-Fänger, nicht Datenrennen-Fänger): die drei Imports D0, D1, D2 laufen als Goroutinen; die `schranke` aus `nebenlaeufig_test.go` hält jeden in `ParseMetadata` seines Codecs (Haltepunkt `setzeHaken`), bis alle drei dort sind, und eine zweite `schranke` jeden in `ParseCollection`, bis alle drei dort sind; danach läuft `ParseEntity`. Die zweite ist nötig, weil (c) nach `ParseMetadata` schreibt und vor der Antwort liest: kein Haltepunkt von `ParseMetadata` liegt zwischen beiden, also verschränken sich Schreiben und Lesen ohne sie dem Zufall nach. Erwartung: jeder Fake sieht nur Aufrufe seines Dienstes, kein Aufruf über Kreuz, jedes Ergebnis das eigene. Dazu hält `gleichschritt` jeden Import nach jedem Aufruf der Ports der Welt mit Haltepunkt (`Imports.Get`, `Sessions.Get`, `Exchanges.ListBySession`, `Blobs.Open`, `Blobs.Begin`, `Imports.Publish`, `Metadaten.SetMetadata`: 9 Schranken in 3 Imports) an, bis alle drei ihn beendet haben; ein Zustand, der zwischen zwei solchen Aufrufen geschrieben und gelesen wird, sieht so die Schreiber der anderen. Die Frist der Schranke macht eine Sperre um den ganzen Import zu einem Fehler statt zu einem Hänger. `make test-race` fährt dieselbe Zeile zusätzlich, ohne dass die Zusage daran hängt | (a) die Wahl steht vor `ParseMetadata` in einer Variablen auf Paketebene (kein Feld: das Feld fängt der Zähler des Typs `Importer`) und wird danach gelesen; (b) ein `sync.Mutex` um den ganzen Import (Frist der Schranke); (c) der Import merkt sich das Modell nach `ParseMetadata` in einer Variablen auf Paketebene | nebenläufig |
| Z8 | Ein Import, der vor dem Lesen endet, und `Get` rufen keinen Codec, und die Wahl des Satzes liegt hinter der Prüfung des Dienstes (`SPEC-127` Nr. 4 a bis d, 5, 6, 16) | `TestSatzKeinAufrufVorDemLesen`: Schleife über Ausgang {Anfrage ungültig (Nr. 4 a), `ServiceID` unbekannt (4 b: `ODR-E5005`), Wiederholung mit gleichem Hash, Wiederholung mit anderem Hash, Session unbekannt, Session eines anderen Service, Session offen, `$metadata` fehlt, `$metadata` mehrdeutig, Hash der Anfrage ungleich der Referenz (Nr. 7 a bis e), Fehler von `Imports.Get`, von `Sessions.Get` und von `Exchanges.ListBySession`} = 13 Ausgänge und `Get`, mal Dienst {D0, D1, D2}: Zähler aller sechs Codec-Fakes gleich 0, kein Panic | (a) die Wahl des Satzes (`dienstVon`) steht vor der Prüfung der `ServiceID` (Panic bei unbekannter `ID`); (b) `ParseMetadata` läuft vor der Wiederholungsprüfung | andere |
| Z9 | `Codecs` `nil` verhält sich wie vor dem Slice: jeder Test des Pakets für Version 4 bleibt textgleich und grün (`SPEC-127` Nr. 1) | die vorhandene Suite des Pakets unverändert (kein Eintrag in `git diff --numstat <Basis>..HEAD` für eine vorhandene `*_test.go`; neue Tests in neuen Dateien, `helfer_test.go` bleibt unberührt); der Fall D0 aus Z1 ist der Fall in der Schleife | `nil` wird zu `ErrUngueltig` oder zu einem Nullwert-Satz: viele Tests der vorhandenen Suite rot | andere |

Weitere Mutationen, die der Implementer an Elementen **außerhalb** der genannten fährt (`AGENTS.md` §3.17): `ParseCollection` statt `ParseEntity` mit dem Satz von D2 bei der Spur Entity vertauscht; Zähler-Fake für den Strom (`Close` doppelt bei einem Dienst mit Satz); Satz von D1 an die Antworten von D2 (Dienst nach `a.ServiceID`, aber Index des vorigen Imports). Der Bericht nennt je Befund Menge, Zahl der Elemente und mindestens eine Mutation an einem nicht genannten Element.

**Akzeptiertes Negativ (Grund und Folge):** `New` prüft `Metadaten == nil` und `Daten == nil` auf Schnittstellenebene; eine Schnittstelle mit nil-Zeiger darin besteht die Prüfung. Grund: ohne Reflexion nicht prüfbar, und die Felder der Konfig (`MetadatenCodec`, `DatenCodec`) haben dieselbe Grenze; Folge: ein Root, der einen nil-Zeiger als Codec setzt, scheitert mit einem Panic im ersten Import statt in `New` (Fehler des Roots, im Rundlauf von `slice-v2-bootstrap` sichtbar). Das Negativ gilt einmal für die Konfig und den Satz und wird nicht weiter verfolgt.

**Akzeptiertes Negativ:** `NurLesend` hat in diesem Slice keinen Leser (der Mock liest es, `slice-v2-bootstrap`, `SPEC-174` Nr. 3); Fänger der Wirkung liegt dort. Folge bis dahin: das Feld ist Vertrag ohne Wirkung, Z5 sichert nur, dass der Importer es nicht liest.

**Nebenläufigkeit (§3.10):** der Dienst hält keine Sperre (`SPEC-127` Nr. 18); „unter einer Sperre ohne Freigabe dazwischen“ gilt nicht, daher keine Zeile je Fenster. Z7 ist die einzige nebenläufige Zeile; erwartet ist Rot in jedem Lauf (Schranke), der Implementer misst und trägt Zahl der Prozesse, Messenden und Rate je Mutation in den Absatz von §3 ein. Ob `make test-race` die Mutation (a) je Prozess meldet, ist nicht zugesagt und wird nicht behauptet. Ein Abschnitt ohne Aufruf eines Ports mit Haltepunkt zwischen Schreiben und Lesen eines Zustands ist ein akzeptiertes Negativ mit Grund und Folge im Absatz *Raten von Z7*; dort steht die Rate des Detektors mit Zahl der Prozesse.

**Mutationen und Raten (Implementer, `AGENTS.md` §3.10).** Die Fänger liegen in `satz_test.go` (neue Datei; keine vorhandene `*_test.go` ist geändert). Der Test-Helfer heißt `satzFake` (Konstruktoren `konfigSatz` für D0, `neuerSatz` für D1 und D2; `welt.baue` legt Dienst, Session und die drei Exchanges an). Mengen, die die Tests zählen: Z1 9 Fälle, Z2 18, Z3 9 und der Fall mit leerer `ID`, Z4 6, Z5 8 und der geteilte Satz, Z6 2 und die doppelte `ID`, Z7 3 Imports und 9 Schranken im Gleichschritt, Z8 39 und 3 Fälle für `Get`. Jede Mutation ist an einer Kopie des Arbeitsbaums im Container ohne Netz gefahren und rot; die Spalte nennt den roten Test, bei weiteren roten Tests auch sie:

| Zeile | Mutation | roter Test |
|---|---|---|
| Z1 | (a) `ParseEntity` nimmt den Codec der Konfig · (b) `ParseMetadata` nimmt den der Konfig · (c) der Satz kommt von `Dienste[0]` · (d) `ctx` von `ParseMetadata` ist bei einem Dienst mit Satz `context.Background()` · (e) Nullmodell an die Antwort-Codecs bei einem Dienst mit Satz · (f) der Strom wird bei einem Dienst mit Satz gehüllt (`io.MultiReader`) | `TestSatzWahlJeAufrufstelle` (bei (a), (b), (c), (f) mit weiteren Tests des Satzes) |
| Z1, nicht genannte Elemente | `grenze` von `ParseCollection` minus 1 bei einem Dienst mit Satz · `ParseCollection` nimmt den Codec der Konfig · `entitySet` von `ParseEntity` bei einem Dienst mit Satz verlängert · `ctx` von `ParseCollection` bei einem Dienst mit Satz `context.Background()` · Hülle des Stroms von `ParseMetadata` bei einem Dienst mit Satz | `TestSatzWahlJeAufrufstelle` (und weitere des Satzes) |
| Z2 | (a) `ParseEntity` verwirft bei einem Dienst mit Satz den Fehler · (b) der Kontextfehler wird bei einem Dienst mit Satz an allen drei Aufrufstellen `ErrUngueltig` · (c) `ParseMetadata` schließt bei einem Dienst mit Satz den Strom nicht | `TestSatzFehlerJeAufrufstelle` |
| Z2, Klasse der Zähler (Fix zu F-1) | (d) nach dem Fehler von `ParseMetadata` ruft der Dienst den `MetadatenCodec` der Konfig, bei einem Dienst mit Satz · (e) nach dem Fehler von `ParseMetadata` ruft er den Metadatenleser des ersten anderen Dienstes mit Satz · (f) nach dem Fehler von `ParseCollection` ruft er `ParseEntity` der Konfig · (g) nach dem Fehler von `ParseEntity` ruft er `ParseCollection` des ersten anderen Dienstes mit Satz · (h) nach dem Fehler von `ParseEntity` ruft er `ParseMetadata` der Konfig · (i) nach einem Kontextfehler von `ParseCollection` ruft er `ParseMetadata` des ersten anderen Dienstes mit Satz | `TestSatzFehlerJeAufrufstelle`, je Mutation rot mit der Zählermeldung des fremden Codecs (`ParseMetadata von D0 1-mal, erwartet 0` bei (d); (e) bis (i) analog); der Befund nannte (d) und (e), (f) bis (i) sind Elemente, die er nicht nannte |
| Z2, nicht genannte Elemente | `Close` doppelt bei `ParseCollection` eines Dienstes mit Satz · `ParseCollection` verwirft bei einem Dienst mit Satz den Fehler · Kontextfehler allein an `ParseCollection` zu `ErrUngueltig` | `TestSatzFehlerJeAufrufstelle` |
| Z3 | (a) die Prüfung nur für `Dienste[0]` · (b) `&&` statt `\|\|` · (c) nur `Metadaten` geprüft · (d) nur vor dem ersten gültigen Satz geprüft · nicht genannt: nur `Daten` geprüft | `TestNewSatzTeilNil` |
| Z4 | die Prüfung der Konfig-Codecs entfällt, sobald jeder Dienst einen Satz trägt | `TestNewKonfigCodecsPflicht` |
| Z5 | (a) der Import vergleicht `Satz.Version` mit `Modell.Version()` · (b) `New` weist `Version` leer ab · (c) `NurLesend` wahr macht `Import` ungültig · (d) `New` weist zwei Dienste mit demselben Satz-Zeiger ab | `TestSatzUebrigeFelder` (bei (b) auch `TestNewKonfigCodecsPflicht`, dessen Gutfall einen Satz ohne `Version` führt) |
| Z6 | (a) `normiereDienste` lässt `Codecs` weg · (b) `normiereDienste` gibt `in` zurück · (c) die doppelte `ID` wird am ganzen `Dienst` (`==`) erkannt | `TestNormierungBehaeltSatz` (bei (a) und (b) mit weiteren) |
| Z7 | (a) die Wahl des Antwort-Codecs steht vor `ParseMetadata` in einer Variablen auf Paketebene · (b) ein `sync.Mutex` auf Paketebene um den ganzen Import · (c) der Import merkt sich das Modell nach `ParseMetadata` in einer Variablen auf Paketebene | `TestSatzGleichzeitig` (bei (b) auch `TestZeileB5GleicheAnfrage`, `TestZeileB6GleicheIDAndereAnfrage`, `TestZeileB7VerschiedeneIDs`, die ebenfalls an der Schranke hängen) |
| Z8 | (a) die Wahl des Satzes steht vor der Prüfung der `ServiceID` · (b) `ParseMetadata` läuft vor der Wiederholungsprüfung | (a) `TestSatzKeinAufrufVorDemLesen` (Panic `index out of range [-1]`; das Paket bricht dort ab, `TestZeileA3DienstUnbekannt` stirbt als Erster) · (b) `TestSatzKeinAufrufVorDemLesen` und viele des Bestands |
| Z8, Klasse der Ausgänge (Fix zu F-2) | (c) `ParseMetadata` des Dienstes läuft im Fehlerzweig von `waehleDokument` (alle drei Ausgänge Nr. 7 a bis e) · (d) dasselbe nur bei `ErrKeineMetadaten` · (e) nur bei `ErrMetadatenAbweichend` · (f) nur bei `ErrMetadatenHash` · (g) `ParseEntity` des Dienstes läuft im Fehlerzweig von `waehleDokument` · (h) `ParseMetadata` läuft bei einem Fehler von `ladeSession` (Sessions.Get, fremder Service, offen, `ListBySession`) · `ParseMetadata` läuft in `wiederhole` bei einer Wiederholung oder einem Fehler von `Imports.Get` · `ParseMetadata` läuft nach einer ungültigen Anfrage | `TestSatzKeinAufrufVorDemLesen`, je Mutation rot mit `… ParseMetadata von D<k> 1-mal gerufen` im Fall des eigenen Ausgangs (der Befund nannte (c); (d) bis (h) und die letzten drei sind Elemente, die er nicht nannte) |
| Z9 | (a) `Codecs` `nil` wird `ErrUngueltig` · (b) `nil` wird ein Nullwert-Satz | die Suite des Bestands: (a) 100 Tests des Pakets rot, (b) Panic in `TestZeileC1Schreibweisen`, das Paket bricht dort ab |

Raten von Z7, Messender der Implementer: 30 Prozesse je Fall und Bauart, jeder Prozess `-test.count=1` (kompiliertes Testprogramm, `-test.run '^TestSatzGleichzeitig$'`, zehn Prozesse gleichzeitig in einem Container mit 4 CPUs ohne Netz), einmal mit `-race` (Image `odata-recorder:race`, cgo) und einmal ohne (cgo aus). Ungemutiert 0 von 30 rot in beiden Bauarten. Die Mutationen sind (a) die Wahl des Antwort-Codecs steht vor `ParseMetadata` in einer Paketvariablen, (b) ein `sync.Mutex` um den ganzen Import, (c) das Modell nach `ParseMetadata` in einer Paketvariablen; neu in der Fix-Runde: (d) eine Paketvariable wird zu Beginn von `liesAntwort` geschrieben und nach `Blobs.Open` gelesen, (e) wie (d) mit `atomic.Value` (kein Datenrennen), (f) `atomic.Value` zu Beginn von `liesAntwort` geschrieben und allein bei `ParseEntity` gelesen, (g) `atomic.Value` in `liesDokument` vor `Blobs.Open` geschrieben und nach ihm gelesen, (h) `atomic.Value` unmittelbar vor dem Lesen geschrieben, ohne Aufruf eines Ports dazwischen, (i) dasselbe mit einer Paketvariablen.

| Mutation | mit `-race`: rot (davon `DATA RACE`) | ohne `-race`: rot |
|---|---|---|
| (a) | 30 (30) | 30 |
| (b) | 30 (0) | 30 |
| (c) | 30 (30) | 30 |
| (d) | 30 (14) | 30 |
| (e) | 30 (0) | 30 |
| (f) | 30 (0) | 30 |
| (g) | 30 (0) | 30 |
| (h) | 2 (0) | **0** |
| (i) | 10 (10) | **0** |

**Zahlen des Verifiers** (Messender: der Verifier, eigene Bauten der Mutationen nach der Beschreibung oben, 30 Prozesse je Zelle, 20 bei (a) und (c), `-test.count=1`, ein Container mit 4 CPUs ohne Netz; ungemutiert 0 von 30 in beiden Bauarten). Die Spalten nennen rot (davon `DATA RACE`); die Zahlen des Implementers stehen in der Tabelle darüber:

| Mutation | Verifier mit `-race` | Verifier ohne `-race` |
|---|---|---|
| (a) | 20 von 20 (20) | 20 von 20 |
| (c) | 20 von 20 (20) | 20 von 20 |
| (d) | 30 von 30 (28) | 30 von 30 |
| (e) | 30 von 30 (0) | 30 von 30 |
| (h) | 3 von 30 (0) | **0 von 30** |
| (i) | 29 von 30 (29) | **0 von 30** |

Die Zahlen der beiden Messenden weichen allein in den `DATA RACE`-Meldungen und im rot der Fenster ohne Haltepunkt mit `-race` ab ((d) 14 gegen 28 von 30, (h) 2 gegen 3, (i) 10 gegen 29); die Rate des Detektors hängt von der Bauart der Mutation und der Last ab, nicht vom Code, und beide Zahlenreihen sind Messungen unter Last, keine feste Größe. Dieselben Aussagen tragen beide: die Wirkungs-Fänger (a) bis (g) sind in jedem Prozess rot, mit und ohne Detektor, und die Fenster ohne Haltepunkt (h), (i) sind ohne `-race` in keinem Prozess rot.

(a) bis (g) sind Wirkungs-Fänger: sie zählen die Wirkung (Aufrufe über Kreuz, Frist), kein Datenrennen, und sind in jedem Lauf rot. Die Rate ist auf 30 Prozesse je Zelle begrenzt. **Akzeptiertes Negativ (Grund und Folge):** ein Abschnitt, der zwischen Schreiben und Lesen keinen Aufruf eines Ports mit Haltepunkt (die sieben des Gleichschritts, die drei Codecs) enthält, hat keinen Wirkungs-Fänger. Grund: der Test setzt Haltepunkte nur in den Fakes der Ports, kein Test kann zwischen zwei Anweisungen des Dienstes anhalten, und die Wirkung einer atomaren Paketvariablen in diesem Fenster ist nicht beobachtbar (Fall (h): 0 von 30 ohne, 2 von 30 mit `-race`, davon keiner durch einen Detektor). Das Datenrennen einer nichtatomaren Variable meldet allein `make test-race`, und nur mit Wahrscheinlichkeit (Fall (i): 10 von 30 Prozessen mit `-race`, 0 von 30 ohne). Folge im Betrieb: ein Zwischenspeicher der Wahl mit atomarem Zugriff, der in einem Abschnitt ohne Portaufruf schreibt und liest, könnte einem Import gelegentlich den Satz eines anderen Dienstes liefern, ohne dass dieser Test oder `make test-race` zuverlässig rot wird; ein Zwischenspeicher als Feld des Importers bricht den vorhandenen Zähler (Risiko 2), einer auf Paketebene fängt allein der Review der Wahl in `metadatenCodec` und `datenCodec`.

**Fix-Runde zum Review (`AGENTS.md` §3.17).** Die Menge ist aus dem Wortlaut der Zusage gebildet, nicht aus dem Befund; der Fänger ist eine Schleife über sie; die Mutation am Element, das der Befund nicht nannte, ist rot gesehen (Mutanten auf Kopien im Container ohne Netz).

| Befund | Menge (Wortlaut) | Elemente | Mutation an einem nicht genannten Element | rot |
|---|---|---|---|---|
| F-1 (Z2) | „danach kein weiterer Port“: jede Aufrufstelle des Fehlers (3) mal jeder Dienst als Ziel (3, D0 trägt die Codecs der Konfig) mal beide Fehlerarten (2); je Fall die Zähler aller Aufrufstellen (3) aller Dienste (3) | 18 Fälle mit je 9 Zählern | nach dem Fehler von `ParseCollection` ruft der Dienst `ParseEntity` der Konfig; nach dem Fehler von `ParseEntity` `ParseCollection` des ersten anderen Satzes oder `ParseMetadata` der Konfig; nach einem Kontextfehler von `ParseCollection` `ParseMetadata` eines anderen Satzes | `TestSatzFehlerJeAufrufstelle`; die beiden Mutanten des Befunds ebenfalls |
| F-2 (Z8) | „vor dem Lesen endet“: jeder Ausgang der Rangfolge `SPEC-127` Nr. 4 bis 7 vor dem Lesen (13: ungültige Anfrage, unbekannte `ServiceID`, zwei Wiederholungen, drei Fehler der Session, `$metadata` fehlt, mehrdeutig, Hash ungleich, Fehler von `Imports.Get`, `Sessions.Get`, `ListBySession`) mal jeder Dienst (3), dazu `Get` (3) | 39 und 3 Fälle | `ParseMetadata` im Fehlerzweig von `ladeSession`, von `wiederhole`, nach ungültiger Anfrage; `ParseEntity` im Fehlerzweig von `waehleDokument`; je einer der drei Ausgänge von Nr. 7 allein | `TestSatzKeinAufrufVorDemLesen` |
| F-3 (Z7) | „in jeder Verschränkung“: jedes Fenster zwischen zwei Aufrufen von Ports mit Haltepunkt (9 Schranken im Gleichschritt); das Fenster ohne Aufruf ist ein akzeptiertes Negativ | 9 Schranken; 1 Negativ | (d) bis (g) im Absatz *Raten von Z7* | `TestSatzGleichzeitig`, 30 von 30 mit und ohne `-race` |
| F-4, F-5 | Kommentare von `TestSatzFehlerJeAufrufstelle` und `satzDienste` | 2 | — | der Kopf nennt die Zähler, die der Test prüft; der Parameter heißt `ersatz` |
| F-6 | die Beschreibung eines eigenen Zweigs für den Satz in den Mutationen von Z1, Z2, Z7 (der Code hat keinen; Satz und Codecs der Konfig teilen `liesAntwort`, `liesDokument` und `oeffne`, allein die Wahl in `metadatenCodec` und `datenCodec` unterscheidet) | alle Stellen des Plans, die so beschrieben haben | — | — |
| F-7 | Messpunkt der Größenmarke | 1 | — | siehe *Messung* in §4 |

**Fakes (Architect).** Neue Typen von Port-Fakes: 0. `CodecSatz` ist eine Struktur, kein Port; die Ports `MetadataCodec` und `ODataCodec` haben ihre Fakes (`fakeMetaCodec`, `fakeDatenCodec` in `helfer_test.go`), sie halten jedes Argument fest und zählen jede Operation. Als 1 Fake gezählt wird der Test-Helfer `satzFake` in einer **neuen** Datei (z. B. `satz_test.go`): baut je Satz ein eigenes `protokoll`, je eine Instanz von `fakeMetaCodec` und `fakeDatenCodec` mit eigener Vorgabe, hängt dieselbe Prüfung auf Hülle, verbrauchten Strom und fehlende Vorgabe wie `neu(t)` in `t.Cleanup` ein (sonst bleibt „Strom unverändert“ für die Dienste mit Satz ungeprüft), und erweitert die Welt um die Dienste D1 und D2 mit je Session, `$metadata`- und Antwort-Exchanges (vier neue Instanzen der vorhandenen Fakes, zwei Sätze). Die vorhandenen Reflexions-Tests über „die acht Fakes“ (`TestFakesZaehlenJedeOperation`, `TestFakesHaltenJedesArgument`) bleiben unberührt; die Satz-Fakes sind dieselben Typen und treten nicht in deren Listen ein (sonst nicht textgleich). Keine Schnittstelle des Satzes hat einen Test-Server.

**Konstruktionsvorgaben aus dem Bestand (Architect):** `Importer` bleibt ein Typ mit genau einem Feld (`k Konfig`; ein vorhandener Test in `konfig_test.go` und `nebenlaeufig_test.go` zählt es); die Wahl des Satzes geschieht je Aufruf aus dem `Dienst`, es gibt keine Tabelle und keinen Zwischenspeicher im `Importer`. `normiereDienste` baut heute `Dienst{ID, UpstreamURL}` neu und muss `Codecs` mitnehmen (Z6).

**Betriebsregeln** (gelten für jede Rolle dieses Slice):

- Auf Repo-Dateien kein `sed -i`, kein Heredoc, kein `>>`, kein `awk` und kein Python, nur `Edit`/`Write`; keine Hilfsdateien im Repo-Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Werte mit Escape-Folge, BOM oder Zeichen jenseits von ASCII entstehen aus Bytewerten und werden auf Rohbytes gegengelesen (`AGENTS.md` §3.18).
- Jeder eigene `docker run` mit `--memory=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`; Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Mengen sind Schleifen aus dem Wortlaut der Zusage, nie eine Liste der Fälle, die einfielen (`AGENTS.md` §3.17).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13). Der Handoff ist **ein Commit für Code und Plan** (`AGENTS.md` §3.9).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) ist `Accepted` und `SPEC-127` Nr. 1 ist nachgezogen (beides erfüllt). Code-Abhängigkeit allein `slice-baseline-import` in `done/`; nicht `welle-mock`, nicht M2.

**Größenmarke** (`AGENTS.md` §3.14, `SPEC-128`), aus der Tabelle in §3 (9 Zeilen *Zusage · Fänger*, davon *a* = 1 nebenläufig (Z7) und *b* = 8 andere; *f* = 1 Fake (`satzFake`); *p* = 120 Produktcode): *S* = 85·1 + 80·8 + 200·1 + 120 = **1045**, Marke *M* = 100 · ⌈135·1045 / 10000⌉ = **1500**. Die Metrik ist die Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über `src/hexagon/ports/outbound/` und `src/hexagon/services/baselineimport/`, Tests eingeschlossen. Keine Schnittfrage (*S* < 2223). Die vorläufige Prosa-Marke 1400 (Grobschätzung mit 8 Zeilen, *S* = 965) ist durch den Block ersetzt, weil die Tabelle des Architect 9 Zeilen hat. `basis` ist `b5ac4a5`, gesetzt vom Planner vor dem ersten Code-Commit.

```groessenmarke
pfad: src/hexagon/ports/outbound
pfad: src/hexagon/services/baselineimport
produktcode: 120
fakes: 1
basis: b5ac4a5
schaetzung: 1045
marke: 1500
```

**Messung beim Handoff:** `make groessenmarke` meldet 695 von Marke 1500 (Schätzung 1045): `codec_satz.go` 16, `importer.go` 43, `metadaten.go` 7, `satz_test.go` 629 Zeilen. **Messung nach der Fix-Runde zum Review:** 758 von Marke 1500 (Stand `3bef041`, Ausgabe `Messpunkt slice-v2-import-codecs.md: zeilen=9 nebenlaeufig=1 fakes=1 produktcode=120 schaetzung=1045 marke=1500 messung=758`). Messpunkt: die Schätzung (1045) liegt um 38 % über der Messung (695 beim Handoff, 758 nach dem Review), weil Fakes und Tabellenzeilen dieses Slice weniger kosten als die Faktoren 85 und 80 je Zeile; die Marke 1500 bleibt unverändert, die Setzung ist Sache des Planners.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke (Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert — Carveout?): [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md) verliert den Status `Accepted`.

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

- Der Slice ändert den Vertrag eines Dienstes in `done/` (`SPEC-127` Nr. 1); die Tests für Version 4 müssen textgleich grün bleiben — **Ausgang:** offen; Fänger: die V4-Tests unverändert, `Codecs` `nil` als Fall. **Ausgang (bei Closure):** nicht eingetreten (Suite des Bestands unverändert grün in `make gates` und `-count=20`; Z9 (a) 100 Tests rot, (b) Panik).
- Fakes der Ports halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften`) — **Ausgang:** nicht eingetreten (`satzFake` benutzt die vorhandenen Fakes, die jedes Argument festhalten und jede Operation zählen; Z1 und Z2 prüfen sie je Aufrufstelle, die Befunde F-1 und F-2 des Reviews betrafen die Breite der Zähler, nicht den Fake).
- Randformen (`AGENTS.md` §3.12, §3.16): das Raster des Architect steht unten; jede Zelle hat einen Ort oder „gilt nicht“ mit Grund. Eine Zelle ohne Ort geht an den Architect — **Ausgang:** entschieden (Raster vollständig; Spec-Lücke L1 in `SPEC-127` Nr. 1 vom Planner vor dem ersten Code-Commit eingetragen, Historienzeile 2026-11-01).
- Spec-Lücke L1 (`SPEC-127` Nr. 1 nannte weder, was der Dienst vom Satz liest, noch den geteilten Satz) — **Ausgang:** entschieden (Klarstellung der Spezifikation, Rang 2, keine Änderung von [ADR-0011](../../adr/0011-versionierter-adapter-fuer-odata-v2.md); die Zellen „nach L1“ des Rasters sind damit entschieden).
- Z7 trägt eine zweite `schranke` in `ParseCollection` (Plan sagte eine in `ParseMetadata`): die Mutation (c) schreibt nach `ParseMetadata` und liest vor der Antwort; ohne zweite Schranke ist sie dem Zufall überlassen. Keine Abweichung im Gegenstand, kein Typ und kein Fake mehr — **Ausgang:** entschieden (Z7 in §3 nachgezogen, 30 von 30 rot).
- Berichtigung des Plans durch den Architect: DoD-Punkt 2 sagte „`Codecs` `nil` und Teil-`nil` als `ErrUngueltig`“; `Codecs` `nil` ist gültig (`SPEC-127` Nr. 1), `ErrUngueltig` ist allein ein Satz mit `Metadaten` oder `Daten` `nil`, auch der leere Satz — **Ausgang:** entschieden (DoD berichtigt).

**Raster Randformen** (Architect; Lesart: „Nr. 1“ heißt `SPEC-127` Nr. 1 usw.; „L1“ ist der Nachzug in `SPEC-127` Nr. 1, vor dem ersten Code-Commit eingetragen; „gilt nicht“ trägt seinen Grund):

| Operation | `Codecs` fehlt (`nil`) | Satz teil-leer (`Metadaten`, `Daten`) | übrige Felder (`Version`, `NurLesend`) leer oder fremd | zwei Dienste mit verschiedenen Sätzen, geteilter Satz | unbekannter oder fremder Wert | Abbruch (`ctx`) | Wiederholung | Fehler des Satz-Codecs |
|---|---|---|---|---|---|---|---|---|
| `New` | Nr. 1: die Codecs der Konfig gelten (Z1 D0, Z9) | Nr. 1: `ErrUngueltig` ohne Code (Z3) | Nr. 1 nach L1: ohne Wirkung (Z5) | Nr. 1 nach L1: je Dienst sein Satz, geteilt zulässig (Z5, Z6); `SPEC-175` Nr. 3, 6 | doppelte oder leere `ID`: Nr. 1, der Satz ändert es nicht (Z3, Z6) | gilt nicht: `New(Konfig)` hat keinen `ctx` | Nr. 18: kein Zustand nach `New`, zwei `New` mit derselben Konfig sind zwei Dienste (Z5 geteilter Satz) | gilt nicht: `New` ruft keinen Codec |
| `Import` | Nr. 1 mit Nr. 7 f, 9: Codecs der Konfig (Z1, Z9) | gilt nicht: `New` wies den Satz ab, es gibt keinen Dienst (Z3) | Nr. 1 nach L1; Version des Dokuments: Nr. 7 f und `SPEC-171` Nr. 2, Metadatenleser (Z5) | Nr. 1: nie der Codec eines anderen; `SPEC-175` Nr. 3, 6; Nr. 18 gleichzeitig (Z1, Z7) | Nr. 4 b (`ServiceID` unbekannt, ohne Port), Nr. 4 d und 6 (Session eines anderen Service) (Z8) | Nr. 2: dasselbe `ctx` an beide Codecs, Kontextfehler unverändert (Z1, Z2) | Nr. 5: kein Codec, keine Kennung, keine Zeit (Z8) | Nr. 7 f, 9, 17 (Z2) |
| `Get` | gilt nicht: `Get` liest `Dienste` nicht (Nr. 16) | gilt nicht (wie links) | gilt nicht | gilt nicht | gilt nicht: die Kennung geht unverändert an `Imports.Get` (Nr. 16) | Nr. 2 und 16: `ctx` nil ist `ErrUngueltig` | Nr. 16: kein Zustand | gilt nicht: `Get` ruft keinen Codec (Z8) |

Zusatztabelle (Paare, Modus):

| Paar oder Modus | Entscheidung (Ort) | Fänger |
|---|---|---|
| Satz teil-`nil` und ein weiterer Verstoß am selben Dienst (leere `ID`) | beide sind `ErrUngueltig` ohne Code; welcher gemeldet wird, ist nicht zugesagt (Nr. 1) | Z3, Fall „Satz teil-`nil` und leere `ID`“ |
| Dienst mit Satz neben Dienst ohne Satz in derselben Konfig | je Dienst der eigene Codec (Nr. 1) | Z1 (D0, D1, D2 in einer Konfig), Z7 |
| Satz-Codec und Codec der Konfig sind dieselbe Instanz | gleichwertig: die Wahl ist unbeobachtbar, kein Verstoß (Nr. 1 nennt keinen) | gilt nicht: nicht unterscheidbar, kein Fänger möglich (akzeptiert) |
| Satz nach `New` vom Aufrufer geändert | *Ungeprüft* nach L1 (Pflicht des Aufrufers; die Tabelle des Roots ist unveränderlich, `SPEC-175` Nr. 3) | gilt nicht: Zusage des Aufrufers |

**Risiken des Architect:**

1. **Wahl vor Dienstprüfung:** `dienstVon` indiziert mit `-1`, wenn die `ServiceID` unbekannt ist (Panic); die Wahl steht hinter Nr. 4 b — Fänger Z8. **Ausgang:** nicht eingetreten (Mutation „Wahl vor der Dienstprüfung“ rot als Panic in `TestSatzKeinAufrufVorDemLesen`, Review und Verifikation; der Code wählt in `liesDokument` und `liesAntwort`).
2. **`Importer` hält genau ein Feld:** ein Zwischenspeicher für Sätze als Feld bricht den vorhandenen Zähler und die Zusage „kein Zustand“ — Fänger Z7 (a) und der vorhandene Zähler; auf Paketebene fängt ihn Z7 nur, wenn ein Port mit Haltepunkt zwischen Schreiben und Lesen liegt (akzeptiertes Negativ im Absatz *Raten von Z7*). **Ausgang:** nicht eingetreten im Produktcode (`Importer{k Konfig}` unverändert, kein Zwischenspeicher; Z7 (a) bis (g) in jedem Prozess rot); das Fenster ohne Haltepunkt bleibt als akzeptiertes Negativ mit Grund und Folge stehen (F-3 des Reviews, Absatz *Raten von Z7*).
3. **`normiereDienste` verliert `Codecs`:** Z6. **Ausgang:** nicht eingetreten (Mutation (a) und (b) rot in `TestNormierungBehaeltSatz`; `normiereDienste` baut `Dienst{…, Codecs: d.Codecs}` in eine neue Liste).
4. **Textgleichheit der V4-Tests:** eine Änderung an `helfer_test.go` oder an einer vorhandenen Datei ist ein Bruch der Zusage; neue Helfer in neuen Dateien — Z9, Prüfung des Reviewers über `git diff --numstat`. **Ausgang:** nicht eingetreten (`git diff --numstat b5ac4a5..HEAD -- '*_test.go'` nennt allein `satz_test.go`, 692 hinzu, 0 entfernt; Verifikation).
5. **Größe:** *S* = 1045, *M* = 1500; der Aufbau von drei Diensten (Sessions, Exchanges, Blobs, Vorgaben) kann über 120 Zeilen je Fake kosten; bleibt es darüber, nach dem Handoff messen und mit Grund berichten (`AGENTS.md` §3.14), nicht ein Teil abtrennen (alles teilt `satzFake`). **Ausgang:** nicht eingetreten, die Formel überschätzte (Schätzung 1045 gegen Messung 695 beim Handoff und 758 nach dem Review, Marke 1500 nie berührt); Rückgabe der Faktoren siehe *Rückgaben* unten.
6. **`schranke` hängt:** ein Import, der `ParseMetadata` nie erreicht, würde die anderen halten; die Frist der vorhandenen `schranke` (5 s) macht daraus einen Fehler. Der Implementer nutzt sie, baut keine zweite. **Ausgang:** nicht eingetreten (die vorhandene `schranke` wird genutzt, in `TestSatzGleichzeitig` ist eine zweite Instanz des Typs, kein neuer Typ; ungemutiert 0 von 30 rot in beiden Bauarten, Implementer und Verifier).
7. **Kommentar der Konfig:** der Satz „Dienste dient der Prüfung von service_id und der Auflösung der Links“ in `importer.go` ist nach dem Slice unvollständig (die Wahl des Satzes kommt hinzu); im selben Commit nachziehen (`AGENTS.md` §3.7, §3.9). **Ausgang:** nicht eingetreten (Kopf von `Dienst` und `Konfig` im Handoff-Commit nachgezogen; Review (8) und Verifikation ohne Befund).
8. **Akzeptierte Negative:** Schnittstelle mit nil-Zeiger, `NurLesend` ohne Leser, Satz-Codec gleich Konfig-Codec — Grund und Folge stehen in §3 und im Raster. **Ausgang:** akzeptiert.

**Rückgaben bei Closure** (`AGENTS.md` §3.13):

9. **Faktoren der Formel von `SPEC-128`:** die Schätzung (1045) liegt um 38 % über der Messung (758); die Faktoren 85 und 80 je Zeile und 200 je Fake überschätzen einen Slice, dessen Zeilen Schleifen über kleine Mengen mit einem gemeinsamen Fake sind; gemessen sind 758 Zeilen für 9 Tabellenzeilen, einen Fake und 120 Zeilen erwarteten Produktcode (tatsächlich 66 Zeilen Produktcode, 692 Zeilen Test). Eine Messreihe über die Slices der Welle ist Sache des Architect. **Ausgang:** offen — Adresse: Architect (`SPEC-128`, Faktoren); der Slice nimmt nichts mit, die Marke 1500 hielt.
10. **`AGENTS.md` §3.14 gegen `SPEC-128`:** die Regel setzt die Marke auf die Schätzung × 1,25 und rechnet 40 je andere Zeile, die Spezifikation rechnet × 1,35 und 80; dieser Slice folgt `SPEC-128` (Rang 2), die Bestimmung beim User ist unverändert offen (`slice-workspace-use-case` §7, Offener Rest 6). **Ausgang:** weiter offen — Adresse: Architect (Vorschlag an den User); kein Eintrag in diesem Slice ändert es.

Die Sendung `NurLesend` an `slice-v2-bootstrap` (DoD a dort) steht bereits im Bestand der Welle; dieser Slice weist nichts Neues an einen anderen Slice zu.

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

- **Ergebnis:** `outbound.CodecSatz{Version, Metadaten, Daten, NurLesend}` (`codec_satz.go`), `baselineimport.Dienst.Codecs`, Prüfung in `normiereDienste` (Teil-`nil` ist `ErrUngueltig`, `nil` gilt als die Codecs der Konfig), Wahl je Aufrufstelle in `metadatenCodec` und `datenCodec`; `Importer` bleibt bei einem Feld. Die Suite des Bestands ist unverändert grün (kein Eintrag für eine vorhandene `*_test.go` in `git diff --numstat b5ac4a5..HEAD`). Commits: Spec-Nachzug L1 `d8157e7`, Handoff `d3600a3`, Fix-Runde zum Review `3bef041`, Messpunkt `e1a1de0`, Verifikation `c07c31e`. Review `docs/reviews/2026-11-01-slice-v2-import-codecs-review.md` (0 HIGH, 3 MEDIUM, 2 LOW, 2 INFO; 81 von 84 Mutanten rot, die drei grünen in zwei Fängern); Verifikation `docs/reviews/2026-11-01-slice-v2-import-codecs-verifikation.md` auf `e1a1de0`: **konform**, 0 HIGH, 0 MEDIUM, 1 LOW (F-V1), 3 INFO (F-V2 bis F-V4), 30 von 30 eigene Mutanten rot, `make gates` Exit 0. Die Fix-Runde ist von der Verifikation geprüft (Stand `e1a1de0` liegt nach `3bef041`); kein Rest ohne zweiten Kontext.
- **Was hat funktioniert:** (1) Der Vertrag stand vor dem Code: die Lücke L1 in `SPEC-127` Nr. 1 (was der Dienst vom Satz liest, geteilter Satz, Pflicht der Codecs der Konfig) trug der Planner vor dem ersten Code-Commit ein, und das Raster hatte keine leere Zelle; kein Code entschied eine Randform. (2) Die Tabelle *Zusage · Fänger* mit Menge je Zeile als Schleife: die Verifikation fand mit 30 Mutanten an Elementen, die weder Plan noch Review nannten, 30 von 30 rot. (3) Die Fix-Runde zu F-1 und F-2 deckte die Klasse (18 Fälle mit je 9 Zählern; 13 Ausgänge mal 3 Dienste plus `Get`), und die Mutationen an nicht genannten Elementen waren in der Verifikation rot. (4) Die Marke 1500 hielt mit weitem Abstand.
- **Was ging anders als geplant:** (1) Der Plan sagte zu Z7 eine `schranke` in `ParseMetadata`; die Mutation (c) verlangte eine zweite in `ParseCollection` (Risiko-Zeile in §6, Z7 nachgezogen). (2) DoD-Punkt 2 sagte „`Codecs` `nil` und Teil-`nil` als `ErrUngueltig`“; `nil` ist gültig (Berichtigung des Architect). (3) Drei MEDIUM im Review, alle vom selben Typ: der Fänger war schmaler als die Zusage (F-1 Zähler nur des Ziel-Satzes, F-2 sieben von dreizehn Ausgängen, F-3 das Fenster ohne Haltepunkt weder gefangen noch als Negativ benannt). Der Wortlaut der Zusage nannte jeweils die größere Menge („danach kein weiterer Port“, „vor dem Lesen endet“, „in jeder Verschränkung“); der Fänger der ersten Fassung folgte der Menge, die dem Implementer einfiel, nicht dem Wortlaut. (4) Der Plan nannte den Helfer an drei Stellen `fakeSatz`, der Code heißt `satzFake` (F-V1), und die Zeile F-6 der Fix-Runde enthielt den Ausdruck, dessen Fehlen sie zusagte (F-V3); beide sind in dieser Closure berichtigt. (5) Die Raten des Verifiers weichen in den `DATA RACE`-Zahlen von denen des Implementers ab (F-V2); beide stehen jetzt im Plan, die Aussagen sind gleich. (6) Der Bericht der Verifikation fiel an `make docs-check` (zwei Kennungen `ADR-0011` ohne Link, Zeilen 7 und 128; der Verifier fuhr `make gates` vor dem Schreiben des Berichts); in der Closure verlinkt, Beobachtung `report-der-rolle-faellt-docs-check` (2×).
- **Lerneintrag:** Zwei Einträge. (1) *Benannte Spec-Lücke, verkörpert:* `SPEC-127` Nr. 1 nannte weder das, was der Dienst vom Satz liest (nur `Metadaten` und `Daten`), noch den geteilten Satz, noch dass die Codecs der Konfig auch bei lauter Diensten mit Satz Pflicht bleiben, noch dass die Änderung eines Satzes nach `New` *Ungeprüft* ist (L1). Der Planner schloss sie in der Spezifikation vor dem Code (Historienzeile 2026-11-01); ihre Fänger sind Z3 bis Z6. Eine Lücke dieser Art fiel beim Start des Slice auf, weil die Tabelle des Architect eine Zeile je Feld des Satzes verlangte und die Spezifikation für `Version` und `NurLesend` schwieg. Ein `liegt in` entfällt: Zielort ist die versionierte Spezifikation. (2) *Gezählt, nicht verkörpert:* Eine Zeile der Tabelle *Zusage · Fänger* nennt die Menge der Zusage, und der Fänger prüft erst dann die Klasse, wenn die Achsen der Menge im Wortlaut der Zusage gezählt sind. Hier fehlten Achsen, obwohl `AGENTS.md` §3.15 bis §3.17 verkörpert sind: die Zähler aller *anderen* Dienste (F-1), alle Ausgänge der Rangfolge (F-2) und das Fenster ohne Aufruf (F-3). Diff-Vorlage an den User, nur genannt, nicht geschrieben: `.claude/commands/implement-slice.md` Schritt 14 und die MEDIUM-Zeile des Reviewers nennen als Prüffrage „welche Zähler und welche Ausgänge nennt der Wortlaut der Zusage, und führt der Fänger jede?“. Der Eintrag ist bei den Beobachtungen unten gezählt.
- **Nicht durch einen zweiten Kontext geprüft:** nichts. Review und Verifikation sind je einmal gelaufen; die Verifikation prüfte den Stand nach der Fix-Runde.
- **Steering-Loop-Eintrag:** Mit diesem Slice wurde in `AGENTS.md`, einem Command, einem Rollen-Typ oder einem Skill nichts geschrieben; deshalb kein Feld `liegt in`. Verkörpert ist allein die Spec-Lücke L1 in `spec/spezifikation.md` (`SPEC-127` Nr. 1).
- **Beobachtungs-Register (`../observations/`):** **Geschrieben:** sieben Dateien `evidence/slice-v2-import-codecs.md` in vorhandenen Verzeichnissen von `BEO-REPO/` und ein neues Verzeichnis. `report-der-rolle-faellt-docs-check` (Bericht der Verifikation, zwei Kennungen ohne Link; Zähler 2, offen), `kommentar-nennt-mehr-elemente-als-der-test-prueft` (F-4; Zähler 3, **Schwelle erreicht**, offen, Ausgang noch nicht zugewiesen), `universalaussage-im-vertrag-ungeprueft` (F-1, F-2; verkörpert in `AGENTS.md` §3.16), `nebenlaeufiger-lauf-nur-im-glueck-gruen` (F-3; verkörpert in `AGENTS.md` §3.10), `akzeptiertes-negativ-ohne-folge` (F-3, das Fenster ohne Haltepunkt trug kein benanntes Negativ; verkörpert in `AGENTS.md` §3.15), `rate-ohne-detektor-schwankt-mit-der-last` (F-V2; Zähler 5, **Schwelle überschritten**, offen, Ausgang noch nicht zugewiesen), `berichtigung-erreicht-nicht-jede-stelle` (F-V1, F-V3; verkörpert in `AGENTS.md` §3.9). **Neu:** `BEO-REPO/schaetzformel-ueberschaetzt-den-slice` (F-7 des Reviews, F-V4; 1×, offen). Dreimal oder öfter ohne Ausgang, nur genannt: `kommentar-nennt-mehr-elemente-als-der-test-prueft` und `rate-ohne-detektor-schwankt-mit-der-last`; Adresse User, Verkörperung ist ein Diff-Vorschlag.
- **Größe (Messpunkt *Ende*, `AGENTS.md` §3.14, `SPEC-128` Nr. 9):** 758 Zeilen nach `git diff --numstat b5ac4a5..HEAD` über die zwei Pfade des Blocks (16 + 43 + 7 + 692; Produktcode 66, `satz_test.go` 692) gegen die Marke 1500, die fest blieb: **gehalten**, Auslastung 0,505. Wortlaut des Gates: `slice-v2-import-codecs.md: ok, Messung 758 von Marke 1500`; Messpunkt `zeilen=9 nebenlaeufig=1 fakes=1 produktcode=120 schaetzung=1045 marke=1500 messung=758`. Die Formel überschätzte (Schätzung 1045, Messung 758): Rückgabe 9 in §6.
- **DoD mit Beleg:** (1) `Dienst.Codecs` mit Prüfung und Wahl, V4-Tests textgleich grün: `codec_satz.go`, `importer.go`, `metadaten.go`; `TestNewSatzTeilNil`, `TestSatzWahlJeAufrufstelle`, Suite des Bestands in `make gates` und `go test -count=20` (Verifikation). (2) Fänger: `satz_test.go`, neun Zeilen Z1 bis Z9, Mengen gezählt (Z1 9 Fälle, Z2 18, Z3 9 und 1, Z4 6, Z5 8 und 1, Z6 2 und 1, Z7 3 und 9 Schranken, Z8 39 und 3). (3) Handoff nach `AGENTS.md` §3.9: `d3600a3` trägt Code und Plan. (4) `make gates` Exit 0 am letzten Commit (Wortlaut im Bericht der Closure).
- **Risiken aus §6:** alle mit Ausgang. Nicht eingetreten: Risiken 1 bis 7 des Architect, die beiden ersten Zeilen (V4-Tests, Fakes) und die Größe. Entschieden: Randformen, L1, zweite `schranke`, DoD-Berichtigung. Akzeptiert: Risiko 8 samt dem Fenster ohne Haltepunkt (Absatz *Raten von Z7*). Weiter offen: Rückgaben 9 und 10.
- **Drei Paarungen:** Anker (kein Zielort-Feld in diesem Slice, nichts verkörpert außer L1 in der Spezifikation; die Sensor-Paarung hat keinen Eintrag zu prüfen) · Folge-Slice (Rückgabe 9 und 10 an den Architect, keine Sendung an einen Slice; die bestehende Sendung `NurLesend` an `slice-v2-bootstrap` ist unverändert) · Register (getragen: jeder genannte Fund hat seine Datei `evidence/slice-v2-import-codecs.md`) — die nächste Welle-Closure prüft die Paarungen erneut.
- **Offener Rest:** (1) Rückgabe 9 (Faktoren von `SPEC-128`) und Rückgabe 10 (`AGENTS.md` §3.14 gegen `SPEC-128`), Adresse Architect, die zweite zusätzlich beim User. (2) Die Diff-Vorlage des Lerneintrags (2) und die zwei Beobachtungen bei 3× oder mehr ohne Ausgang, Adresse User. (3) `slice-mv` nach `done/` und der Merge nach `main` sind nicht Teil dieser Closure-Phase.

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-11-01); `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) gilt für die Fakes der Codec-Ports; `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×) für die Betriebsregel „nur `Edit` und `Write`“. Keiner der offenen Einträge erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
