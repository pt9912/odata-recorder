# Slice slice-blob-ablage: Blob-Ablage: content-adressiert, mit Capture-Abschluss

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-persistenz](../welle-persistenz.md).

**Bezug:** `LH-FA-02`, `LH-FA-14`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `SPEC-106` (Vertrag der Blob-Ablage) · `SPEC-046` (Sicheres Streaming-Capture) · `SPEC-013` (`MAX_CAPTURE_BODY`) · `ARC-010`.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-19.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Der Driven Adapter `src/adapters/driven/blobs` legt Bodies content-adressiert (SHA-256) im Dateisystem ab, liefert den `CaptureWriter` mit genau einem `Commit` oder `Abort`, und prüft die Integrität beim Lesen.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Redaction der Bodies — Fachlogik der Welle welle-record-replay; die Ablage speichert Bytes, die sie bekommt.
- Abgleich der Blobs gegen die Referenzmenge — `StorageMaintenance` und `MaintenanceUseCase` kommen mit [slice-workspace-state-persistenz](../open/slice-workspace-state-persistenz.md) und den Services.
- Verschlüsselung ruhender Daten — nicht zugesagt.

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

- [ ] Outbound Port `BlobRepository` mit `CaptureWriter` und der Adapter: schreiben über temporäre Datei und atomares Umbenennen, Rechte `0600`, Hash und Länge beim Abschluss.
- [ ] Lesen prüft den Hash (`ODR-E3001` bei Abweichung); Auflisten und Entfernen; gleiche Inhalte teilen eine Datei.
- [ ] Tests der Randformen aus §6 gegen ein temporäres Verzeichnis.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Closure-Notiz mit Steering-Loop-Lerneintrag.
- [ ] Beobachtungs-Register (`../observations/`) fortgeschrieben — neues Verzeichnis `BEO-<KUERZEL>/<slug>/` oder eine weitere Datei in dessen `evidence/`; **kein Zaehler wird gesetzt**, er folgt aus den Dateien. Keine Beobachtung angefallen ist ebenfalls eine Antwort und wird in §7 notiert.
- [ ] Jedes Risiko aus §6 trägt einen Ausgang (eingetreten / entfallen / weiter offen).
- [ ] Die drei Paarungen (Anker · Folge-Slice · Register) sind getragen — von der nächsten Welle-Closure geprüft, auch für Slices ohne Wellen-Zugehörigkeit.

## 3. Plan (vor Code)

Regeln dieser Sektion: Baseline-Regelwerk `grundlagen-bootstrap.md`
§Was ist eine Sub-Area? — diese Liste liefert die **Pfad-Kandidaten** für §8,
nicht die Antwort: Pfad-Berührung ist nicht hinreichend, und eine
Aussagen-Berührung steht hier gar nicht.

| Datei / Komponente | Änderungs-Art | Begründung |
|---|---|---|
| `src/hexagon/ports/outbound/` | update | `BlobRepository`, `CaptureWriter` |
| `src/adapters/driven/blobs/` | neu | Dateisystem-Ablage, content-adressiert |
| `spec/spezifikation.md` | update | `SPEC-106` mit den Randformen: `Write` nach `Commit`, doppelter `Commit`, `Abort` nach `Commit`, Überschreiten der Grenze, Verzeichnis nicht beschreibbar, Datei fehlt oder beschädigt |
| `src/hexagon/ports/outbound/outboundtest/` | update | Vertragstest `BlobVertrag` am Port, gefahren gegen den Adapter |

- Der Adapter nutzt nur die Standardbibliothek; Dateisystemzugriff gehört in diesen Adapter und nirgends sonst hin.

**Zusage · Fänger** (`AGENTS.md` §3.15):

| Zusage | Fänger (Teiltest von `TestBlobVertrag`, sonst benannt) |
|---|---|
| Referenzform: `sha256:` und 64 Kleinbuchstaben-Hexziffern, vor dem Kontext geprüft | `ungültige Referenzen` (Großbuchstaben, `../`, kurz, lang, leer, anderer Hash, Präfix `sha257:` gleicher Länge) bei `Open`, `Verify`, `Remove`; Mutationen: Längenprüfung entfernt, Präfixprüfung entfernt, Großbuchstaben zugelassen, Kontext zuerst |
| `Begin`: `grenze` ≥ 1, vor dem Kontext geprüft | `Begin mit Grenze 0 und -1` (auch bei abgebrochenem Kontext); Mutationen: Grenze 0 zugelassen, Kontext zuerst |
| `Write` über der Grenze ist `ErrZuGross`, Datei gelöscht, genau `grenze` zulässig | `Grenze genau erreicht und um eins überschritten`; Mutationen: `>` zu `>=`, temporäre Datei bleibt |
| `Commit` liefert Referenz und Länge, leerer Body zulässig | `Commit liefert Hash und Länge, leerer Body`; Mutation: Länge falsch |
| Gleicher Inhalt, eine Datei, auch gleichzeitig | `gleicher Inhalt teilt eine Datei, auch gleichzeitig`; Mutation: Umbenennen als `Link` (scheitert an der vorhandenen Datei); der Mutex je Writer entfällt, ein Writer gehört einer Goroutine |
| `Commit` heilt eine beschädigte Datei und entfernt eine Nicht-Datei | `Commit ersetzt eine beschädigte Datei`, `Open und Verify prüfen den Hash` (Verzeichnis); Mutationen: vorhandene Datei bleibt, Entfernen der Nicht-Datei entfernt |
| Genau ein Abschluss | `genau ein Abschluss`, `Grenze genau erreicht …` (`Write`/`Commit` nach `ErrZuGross`); Mutationen: Zustandsprüfung in `Write` und `Commit` entfernt |
| `Abort` hinterlässt keine Datei, ist wiederholbar, lässt einen Commit bestehen | `Abort räumt auf und ist wiederholbar`, `genau ein Abschluss`; Mutation: `Abort` meldet Fehler beim zweiten Aufruf |
| Writer ist an den Kontext gebunden | `Abbruch des Kontexts beendet Write und Commit` (Verzeichnis leer); Mutationen: Kontext in `Write`, in `Commit` ignoriert |
| Datei erst nach `Commit` sichtbar (temporäre Datei, Umbenennen) | `Datei erst nach Commit sichtbar`; Mutation: temporäre Datei mit hexförmigem Namen im Blob-Verzeichnis |
| Rechte `0600`, Verzeichnisse `0700` unabhängig von der `umask` | `TestRechteUnabhaengigVonUmask` mit `umask 0` und `0277`; Mutationen: Rechte `0644`/`0755`, `Chmod` der Datei und des Verzeichnisses entfernt |
| `Open` prüft den Hash vor dem Leser; fehlend `ErrBlobFehlt`, beschädigt `ErrBlobBeschaedigt` | `Open und Verify prüfen den Hash` (fehlend, verändert, gekürzt, leer, verlängert, Verzeichnis); Mutationen: Hashprüfung entfernt, Dateiprüfung entfernt, Fehler vertauscht |
| `Verify` wie `Open` | `Open und Verify prüfen den Hash`; Mutation: `Verify` ignoriert `ErrBlobBeschaedigt` |
| `List` sortiert, ignoriert temporäre, fremde Dateien und Verzeichnisse | `List sortiert und ignoriert Fremdes`; Mutationen: Reihenfolge umgekehrt (die Ordnung leistet `os.ReadDir`), Fremdes gelistet, Filter `IsRegular` entfernt |
| Verweis an einer Referenz ist eine Nicht-Datei (`Open`, `Verify`, `List`, `Remove`, `Commit`) | `Verweis an einer Referenz ist eine Nicht-Datei`; Mutationen: `Lstat` zu `Stat` in `Open`, Verweise in `List` gezählt; `Commit` ersetzt den Verweis durch das Umbenennen selbst, ein `Lstat` zu `Stat` dort ist äquivalent |
| `Remove` löscht auch eine Nicht-Datei, unbekannt ist kein Fehler | `Remove, unbekannt ist kein Fehler`, `Open und Verify prüfen den Hash` (gefülltes Verzeichnis); Mutationen: Fehler bei unbekannt, `Remove` ohne `RemoveAll` |
| Erzeugen räumt temporäre Dateien früherer Läufe | `TestErzeugenLoeschtReste`; Mutation: Aufräumen entfernt |
| Verzeichnis fehlt, ist kein Verzeichnis, Unterverzeichnis nicht anlegbar (`sha256`, `tmp`), leerer Pfad: keine Datei angelegt | `TestErzeugenMitUngueltigemVerzeichnis`; Mutation: `New` legt das Verzeichnis an; leerer Pfad und Nicht-Verzeichnis scheitern an `os.Stat` und `MkdirAll` und haben keine eigene Prüfung |
| Abgebrochener Kontext beendet jede Operation ohne Änderung | `abgebrochener Kontext beendet jede Operation ohne Änderung`; Mutationen: Prüfung in `Begin`, `List`, der gemeinsamen Referenzprüfung entfernt |
| `fsync` der Datei und des Verzeichnisses | kein Fänger: ob der Aufruf geschieht und hält, ist von außen nicht zu beobachten; `SPEC-106` führt es unter *Ungeprüft* |

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): [slice-sqlite-adapter-migration](../done/slice-sqlite-adapter-migration.md) liegt in `done/`.

**Marke überschritten:** Die Summe aus `git diff --numstat` über die Pfade aus §3 liegt bei rund 980 Zeilen, getragen von den Vertragstests (`outboundtest`) und den Adapter-Tests. Der Slice bleibt ungeteilt, weil Port, Adapter und Vertragstest einzeln nichts liefern.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 500 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen).
- `in-progress` → `open` (blockiert): die Rechte `0600` lassen sich im Container-Dateisystem nicht festlegen.

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün, und eine Mutation je Zusage wurde rot gesehen; die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Randformen der Ablage (`AGENTS.md` §3.12): leere und überlange Bodies, doppelter und gleichzeitiger `Commit`, `Write` und `Abort` nach dem Abschluss, Abbruch des Kontexts, Absturz vor dem Umbenennen, fehlende, veränderte und gekürzte Datei, ungültige Referenz, Verzeichnis nicht beschreibbar — **Ausgang:** entfallen, bis auf zwei: `Remove` gegen `Commit` gleicher Bytes und die Nutzung eines Writers aus mehreren Goroutinen bleiben unentschieden und gehen an den Architect (Verifikations-Befund V-3); entschieden in `SPEC-106` ist alles übrige, darunter die Nicht-Datei an einer Referenz, die Reihenfolge beim Erzeugen und der Fehler des Verzeichnis-`fsync` fanden Review und Verifikation und stehen jetzt dort.
- Dateisystem-Eigenheiten des Containers (Rechte, Umbenennen über Mountgrenzen) — **Ausgang:** weiter offen: wird im ersten Lauf im Container sichtbar und dann als Beobachtung geführt.

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

Wird bei der Closure geschrieben.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (2×), `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (1×) und `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15); sie fließen als Randformen in §6 und als Tabelle *Zusage · Fänger* in §3 ein.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
