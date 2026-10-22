# Slice slice-replay-service: ReplayUseCase: Zuordnung, Strategien, Cursor

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-FA-04`, `LH-FA-05`, `LH-FA-14`, `LH-QA-02`, `LH-QA-03`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-04.a`, `LH-FA-04.b`, `LH-FA-04.c` · `SPEC-052` · `SPEC-045` · `LH-FA-14.a` (Redaction und Match-Key) · `SPEC-115` Nr. 3 (`Location` mit Marker, Entscheidung des Owners) · `SPEC-118` · `SPEC-009`, `SPEC-078`, `SPEC-083` · `SPEC-020`, `SPEC-021`, `SPEC-050`, `SPEC-065`, `SPEC-070`, `SPEC-071`, `SPEC-074`, `SPEC-093`, `SPEC-097` · `ARC-002`, `ARC-003` · `architecture.md` §4.3.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-20.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `ReplayUseCase` prüft die Zuordnung aus Workspace, Recording-Session und Replay-Session, bildet den Match-Key aus dem redigierten Request, wählt je Strategie (`sequential`, `strict`, `first`, `last`) einen replayfähigen Exchange, schreibt den Cursor erst nach erfolgreicher Auswahl atomar fort und liefert Status, Header und Body ohne Zugriff auf einen Upstream.

**Übernimmt:** Die Übergabe aus `welle-persistenz`: die Abbildung von `ErrBeschaedigt` und `ErrBlob*` auf `ODR-E3001` beim Ausliefern eines beschädigten Blobs (die Tabelle liefert `slice-fehlermodell-konfiguration`).

**Übernimmt (Redaction, Match-Key, Replayfähigkeit):** Der Dienst ruft `BildeSchluessel` (`SPEC-116`) mit dem eingehenden Request, dem durch den Redaktor geleiteten Request-Body (die Funktion redigiert ihn nicht, `SPEC-116` Einleitung) und der Konfiguration des Records, lässt bei `ErrKodierung` (`reject`) `ODR-E5006` ohne Cursor-Fortschritt folgen, wendet `Replayfaehig` (`SPEC-117`) als Prädikat über die Kandidaten an und liefert Header über `OhneRedigierte` (`SPEC-115` Nr. 5); `Location` mit Marker im Query-Wert wird mit `[REDACTED]` ausgeliefert (`SPEC-115` Nr. 3; die Lesart von `SPEC-045` bleibt, Entscheidung des Owners vom 2026-10-21: Weglassen bräche Weiterleitungen und `201 Created`; die Sätze dazu in `SPEC-045` und `LH-FA-14.a` zieht der Architect vor dem Code nach, §6); der Key eines Bodys, der als JSON ungültig ist, folgt im Modus `json-canonical` dem Byte-Hash (Art 2), kein Fehler (`SPEC-116` Nr. 6) (Sendung von `slice-redaction-match-key`).
**Übernimmt (Redaktor):** Der eingehende Request-Body geht durch `RedigiereBody` (`SPEC-118`) mit `MaxBytes = MAX_CAPTURE_BODY`, `Metadata=false` und den wirksamen `Pfade`, der Key wird aus der Ausgabe gebildet, nie aus den Eingabebytes (`SPEC-118` Nr. 10, `SPEC-116` Einleitung); jeder Fehler des Redaktors (auch ein Body über der Grenze oder mit unbekanntem Typ) lässt `ODR-E5001` ohne Cursor-Fortschritt folgen, denn der Key lässt sich nicht bilden und ein `null`-Key passt nie (`SPEC-001`); ein Request ohne Body (0 Bytes) ist kein Fehler (Sendung von `slice-streaming-redaction`).
**Übernimmt (Record-Seite):** Der Record bildet den Key aus den Headern des Requests, wie empfangen (auch `X-Recorder-*` und Hop-by-Hop; `MATCH_VARY_HEADERS` wählt, was eingeht); der Replay bildet ihn aus den Headern seines Requests ebenso, sonst ergibt ein als Vary-Header konfigurierter Name verschiedene Keys (`SPEC-121` Nr. 6). Gespeicherte Response-Header sind die empfangenen ohne Hop-by-Hop, redigiert; `Content-Encoding` und `Content-Length` beschreiben die Leitung, der Blob den entpackten, redigierten Body (`SPEC-119` Nr. 6, `SPEC-121` Nr. 6). Ein Exchange ohne Antwort (Status `0`, `SPEC-121` Nr. 11) ist nie `complete` und nie Kandidat (Sendung von `slice-record-service`).
**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- HTTP-Eingang, Lesen der Steuerheader aus dem Request und Entfernen der `X-Recorder-*`-Header — `slice-http-adapter` reicht die Header unverändert, der Use Case entnimmt die Zuordnung (`SPEC-052`).
- Hybrid-Fallback aus dem Mock auf Replay — Hybrid, nicht in der Welle; der Port `ReplayUseCase` ist der, den Hybrid später ruft.
- Workspace-Anlage und Reset — `slice-workspace-use-case`; Replay liest den Workspace und die Cursors.

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

- [ ] Zuordnung und Validierung: Pflichtheader und Fehler `ODR-E5003`, `ODR-E5004`, `ODR-E7001`, `ODR-E7003`, `ODR-E7009` ohne Cursoränderung; Cursor-Identität `(workspace, replay_session, recording_session)`; neue Replay-Session startet bei Position 0.
- [ ] Auswahl: Match-Key aus dem redigierten Request (Header, Query, Body über den Redaktor), vier Strategien, `ODR-E5001`, `ODR-E5002`, `ODR-E5007`, `ODR-E5008`, Replayfähigkeit nach `SPEC-045`, atomares Fortschreiben des Cursors, unabhängige Cursors paralleler Replay-Sessions.
- [ ] Auslieferung: Body aus dem Blob mit Integritätsprüfung (`ODR-E3001`), Header nach Regel (redigierte nie mit Marker, `Location` mit Marker im Query-Wert ausgeliefert nach `SPEC-115` Nr. 3, volatile nach Policy), `X-Recorder-Match-Redacted: true` bei `key_redacted`; der Service hat keinen Upstream-Port und keinen Import des Upstream-Adapters.
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
| `src/hexagon/ports/inbound/` | update | `ReplayUseCase` |
| `src/hexagon/services/replay/` | neu | Zuordnung, Auswahl, Cursor, Auslieferung |
| `spec/spezifikation.md` | update | Vertrag mit Randformen, Header-Policy der Auslieferung, Verhalten bei Abbruch |
| Tests | neu | Fälle von `LH-FA-04` (Beispiel A, B, A; zwei parallele Sessions; Reset), je Strategie und je Fehlercode; Test, dass `ReplayUseCase` keinen Upstream-Typ importiert |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation oder in `harness/vertraege/` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen, die sie tragen müssen, nennt §6.

- Der Request-Body wird für den Match-Key im Modus `bytes` oder `json-canonical` durch denselben Redaktor geleitet wie beim Record; die Grenze für die Größe des Request-Bodys im Replay ist eine Randform (§6).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-workspace-use-case` und `slice-streaming-redaction` und `slice-redaction-match-key` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über 1700 Zeilen (Summe aus `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen); die Schätzung liegt bei rund 1200 Zeilen in derselben Metrik.
- `in-progress` → `open` (blockiert): die Header-Policy der Auslieferung ist nicht entschieden (siehe §6).

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

- `Location` mit Marker im Query-Wert: der Owner hat am 2026-10-21 entschieden, bei der Lesart von `SPEC-115` Nr. 3 zu bleiben (ausliefern mit `[REDACTED]`, weil Weglassen Weiterleitungen und `201 Created` bricht). `SPEC-115` Nr. 3 trägt Regel und Grund; `SPEC-045` und `LH-FA-14.a` in der Spezifikation sagen „redigierte Header … nie mit dem Marker“ ohne Verweis auf `SPEC-115` Nr. 3, und ein Leser dieser Stellen allein (DoD, Review) liest einen Widerspruch — **Ausgang:** offen; Auftrag an den Architect vor dem Code: ein Satz an beiden Stellen, der auf `SPEC-115` Nr. 3 zeigt, samt Zeile der Historie zur Entscheidung des Owners; der Planner schreibt ihn nicht.
- Welche aufgezeichneten Header ausgeliefert werden: `LH-FA-04.a` sagt „ausgewählte End-to-End-Header“, „zeit- oder hostabhängige Header nur nach expliziter Policy“; die Policy ist nicht festgelegt (`Content-Length`, `Transfer-Encoding`, `Content-Encoding`, `Date`, `Server`, `Location`) — **Ausgang:** offen; Architect entscheidet vor dem Code.
- Cursor fortschreiben ist Lesen, Ändern, Schreiben auf `CursorRepository`; „pro Cursor serialisieren“ verlangt eine Transaktion oder Sperre — `SPEC-108` Nr. 11 nennt nur `Put` — **Ausgang:** offen; Architect legt die Regel fest; das Gate fährt `-race` (`make test-race`, `SPEC-122` aus `slice-race-gate`): ein Datenrennen macht es rot; die nebenläufigen Fänger ordnen weiter mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff (Lesen, Ändern, Schreiben in getrennten Abschnitten) nicht meldet (`SPEC-122` Nr. 8b), und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50`.
- Request-Body im Replay: Größe, Kodierung, nicht redigierbar (`unsafe`) — der Key lässt sich nicht bilden (`SPEC-001`); welche Antwort — **Ausgang:** offen; Raster-Zelle.
- Antwort ohne Body (`204`, `HEAD`), Antwort mit `Content-Encoding`, Antwort mit Redaction und `REPLAY_REDACTED_BODIES=true` — **Ausgang:** offen; Raster *Antwortart × Auslieferung*.
- Abbruch des Clients mitten in der Auslieferung: Cursor ist schon fortgeschrieben (`LH-FA-04.a` Nr. 4: „erst nach erfolgreicher Auswahl“) — **Ausgang:** offen; Architect entscheidet, ob „Auswahl“ die Auslieferung einschließt.
- Randformen (`AGENTS.md` §3.12, §3.16): Das Raster *Operation × Argumentart* entsteht vor dem ersten Code-Commit; jede Zelle nennt ihren Ort in der Spezifikation oder „gilt nicht“ mit Grund, und eine Zelle ohne Ort geht an den Architect — **Ausgang:** offen.

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

**Vorgelagert — offene Beobachtungen sichten:** Das Beobachtungs-Register trägt `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16), `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) und `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (5×, Ausgang `geplant` in `slice-race-gate`); der Slice ist nebenläufig (parallele Replay-Sessions, Cursor): das Gate fährt `-race` (`make test-race`, `SPEC-122`), die nebenläufigen Fänger ordnen weiter mit Kanälen, weil der Detektor ein Rennen ohne ungeordneten Zugriff nicht meldet (`SPEC-122` Nr. 8b), und jeder nebenläufige Test läuft vor der Übergabe mit `-count=50` (wie in `slice-capture-pipeline` §3).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
