# Slice slice-protokolltreue: Vergleichstest der Protokolltreue (Direktzugriff gegen Proxy)

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-record-replay](../welle-record-replay.md).

**Bezug:** `LH-QA-07` (Messmethode: Direktzugriff gegen Proxy-Zugriff am Referenzdienst), `LH-FA-01` (transparente Weiterleitung), [ADR-0001](../../adr/0001-implementierungssprache-go.md) (Go, Test in der Stufe `test`).

**Berührte Spec-Stellen:** `LH-QA-07.a` (Relevanz der Header und Messmethode, Entscheidung des Owners zu `Content-Encoding`) · `LH-FA-01.a` · `SPEC-119` Nr. 6 (Kette und Codings des Entpackens) · `SPEC-121` Nr. 4 und 7 (Hop-by-Hop, `Accept-Encoding`) · `SPEC-063` (`VOLATILE_HEADERS`).
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie
(Baseline-Regelwerk `grundlagen-referenz-richtung.md`
§Referenz-Richtung (SDP), `grundlagen-source-precedence.md` §ID-Schema als Klammer).

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-21.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** Ein Test ruft den Referenzdienst aus `slice-bootstrap-m2` einmal direkt und einmal über den Recorder im Modus `record` auf und vergleicht die entpackte Darstellung der Antworten nach `LH-QA-07.a` — Status, Body nach dem Entpacken seines `Content-Encoding`, relevante Header ohne `VOLATILE_HEADERS` —; er läuft in der Stufe `test` und ist damit Teil von `make test`. Er ist der Nachweis zur Messmethode von `LH-QA-07`.

**Übernimmt:** Den Vergleichstest der Messmethode von `LH-QA-07`, der in der Welle keinen Träger hatte (Closure-Notiz von `slice-record-service`, *Offene Punkte für den User*); Entscheidung des Users vom 2026-10-21: ein eigener Slice am Ende der Welle. Die Messmethode selbst steht in `LH-QA-07.a` der Spezifikation (Architect, nach der Entscheidung des Owners zu `Content-Encoding`); dieser Slice setzt sie um und ändert sie nicht.

**Nutzt:** Referenzdienst und Composition Root von `slice-bootstrap-m2`; jener bleibt beim Satz von M2 (Entscheidung des Users). Was der Vergleich am Referenzdienst zusätzlich braucht (Antworten mit `Content-Encoding`, mehrere Werte je Header, die Randform aus §6), bringt dieser Slice selbst mit.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Ergänzungen am Referenzdienst oder am M2-Test als Umfang von `slice-bootstrap-m2` — dort bleibt der Satz von M2 (Entscheidung des Users); dieser Slice bringt, was der Vergleich braucht, selbst mit und ändert den M2-Test nicht.
- Replay — `LH-QA-07` misst den Record-Modus; den byte-gleichen Rundlauf über Replay prüft der M2-Test von `slice-bootstrap-m2`.
- Markierung unbekannter OData-Erweiterungen als `unsupported_features` (zweiter Satz von `LH-QA-07.a`, `LH-FA-06.a`) — `slice-odata-codec-csdl` (dort DoD, Validierungsbericht); dieser Slice prüft nur, dass der Proxy eine Antwort unverändert weiterleitet, nicht, was der Import daraus macht.
- Änderungen an Record-Dienst, Capture-Pipeline oder Upstream-Adapter — Schicht-Abgrenzung: ein Unterschied, den der Vergleich findet, ist ein Befund an den Planner (die Nehmer liegen dann in `done/`, `AGENTS.md` §3.13), kein Code in diesem Slice.
- Lasttest zu `LH-QA-01` — Welle §6.

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

- [ ] Vergleichstest nach `LH-QA-07.a`: derselbe Request direkt und über den Recorder im Modus `record` an den Referenzdienst; verglichen werden Status, Body nach dem Entpacken (Kette und Codings wie `SPEC-119` Nr. 6) und die relevanten Header (ohne Hop-by-Hop, `Content-Encoding`, `Content-Length` und `VOLATILE_HEADERS`); der Client nennt in `Accept-Encoding` nur `gzip`, `deflate` und `identity`; je Antwortart des Referenzdienstes ein Fall, darunter `$metadata`, Collection, Einzel-Entity, Antwort mit `Content-Encoding`, Fehlerstatus (`LH-QA-07`, `LH-FA-01`).
- [ ] Mutationen des Vergleichs: der Proxy ändert ein Byte des Bodys, einen relevanten Header oder den Status, der Vergleich übergeht einen relevanten Header — jede färbt den Test rot; ein anderes Coding über den Proxy färbt ihn nicht (`LH-QA-07.a`).
- [ ] Der Test läuft in der Stufe `test` (netzlos, nur Loopback) und ist damit Teil von `make test`; die Randformen aus §6 sind vor dem Code in der Spezifikation entschieden.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist (`LH-QA-07.a`: Nachzug nach Review und Verifikation, durch den Architect).
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
| `tests/integration/` | update | Vergleichstest und die Antworten des Referenzdienstes, die er zusätzlich braucht; Referenzdienst und Composition Root aus `slice-bootstrap-m2` |
| `spec/spezifikation.md` | update | `LH-QA-07.a`: die Randformen aus §6, vor dem Code entschieden vom Architect; im Slice nur Nachzug nach Review und Verifikation |
| Testdatei des Vergleichs | neu | ein Fall je Antwortart und je Randform des Rasters; Mutationen des Proxys und des Vergleichs nach `LH-QA-07` |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit den Randformen in `LH-QA-07.a` **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12).

- Der Vergleich prüft sich selbst: eine Mutation des Vergleichs (ein relevanter Header fehlt in der Menge, das Entpacken entfällt) muss ihn rot machen, sonst bewiese ein grüner Vergleich nichts (`AGENTS.md` §3.15).
- Testdaten von Hand: erwartete Header-Listen und Bodies stehen ausgeschrieben; Bytes ab `0x80` als Go-Escapes (`BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um`).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-bootstrap-m2` liegt in `done/` (Referenzdienst, Composition Root und M2-Test im Code). Die Spezifikation trägt die Randformen aus §6 in `LH-QA-07.a`, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12). Dieser Slice ist der letzte der Welle.

**Größenmarke** (`AGENTS.md` §3.14): über **1400** Zeilen (Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3 unter `tests/`, Tests eingeschlossen; `<Basis>` ist der Commit, der die Randformen vor dem Code abschließt; Text der Spezifikation zählt nicht mit). Grund der Ersetzung (§3.14): Marke neu gesetzt nach §3.14 auf die vorläufige Schätzung × 1,25, aufgerundet auf volle Hundert; die ältere Marke 1100 entstand vor dem Faktor 1,25.

**Schätzung** (`AGENTS.md` §3.14), vorläufig, ohne Tabelle: rund 10 Zeilen × 85 (nebenläufig: Proxy, Referenzdienst und Client in einem Prozess) = 850, dazu 200 für die Antworten, um die der Test den Referenzdienst ergänzt (ein Test-Server), zusammen 1050. Die Marke ist die ganze Schätzung, den festen Anteil von 200 eingeschlossen, × 1,25 = 1312,5, aufgerundet auf volle Hundert: **1400**. Steht die Tabelle, rechnet der Planner die Schätzung vor dem ersten Code-Commit neu und setzt die Marke erneut; eine ältere Marke ersetzt er dabei mit Grund in §4. Vor dem Code wird nur geschnitten, wenn die Schätzung × 1,25 über 3000 liegt; hier liegt sie bei rund 1310. Nach dem Code wird gemessen und mit Grund berichtet.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): nicht über die Marke (sie ist die Schätzung × 1,25 und schneidet nicht; vor dem Code wird nur über 3000 geschnitten, oben, `AGENTS.md` §3.14); offen bleiben die Kriterien der Baseline (mehr als drei Liefer-Punkte, mehr als zwei Schichten).
- `in-progress` → `open` (blockiert): der Vergleich findet einen Unterschied, den `LH-QA-07.a` nicht als zulässig nennt; das ist ein Befund an den Planner (der verantwortliche Slice liegt in `done/`), kein Test, der den Unterschied hinnimmt; ebenso eine Randform ohne Ort in `LH-QA-07.a` (an den Architect, `AGENTS.md` §3.12).

## 5. Closure-Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Closure- und Lerneintrag-Regeln — zwei beobachtbare Kriterien **und** ein
Lerneintrag; ohne ihn ist der Slice nur abgelegt.

Die DoD ist vollständig, `make gates` ist grün und der Vergleichstest ist Teil von `make test`, und je Zusage ist eine Mutation rot gesehen (`AGENTS.md` §3.10); die Closure-Notiz mit Lerneintrag liegt vor.

## 6. Risiken und offene Punkte

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Offene Risiken werden bei Closure aufgelöst — **jedes** Risiko bekommt genau
**einen** Ausgang, und kein Slice geht nach `done/`, während eines ohne Ausgang
dasteht.

- Nicht angefragtes Coding: der Referenzdienst sendet ein `Content-Encoding`, das der Client nicht genannt hat (etwa `br` auf `Accept-Encoding: gzip`); direkt kann der Test es nicht entpacken (`SPEC-119` Nr. 6: kein Dekoder in der Standardbibliothek), über den Proxy leitet der Record-Pfad die Antwort unverändert weiter — **Ausgang:** offen; nach dem Architect für das Raster dieses Slice nicht entschieden; er entscheidet es vor dem Code in `LH-QA-07.a` (Vergleich, Ausschluss oder Fall mit eigener Erwartung).
- Unterschied, den der Vergleich findet: ein Unterschied in Status, Body oder relevantem Header zwischen Direktzugriff und Proxy ist ein Befund gegen `LH-FA-01` in einem Slice, der dann in `done/` liegt — **Ausgang:** offen; Befund an den Planner (`AGENTS.md` §3.13), der über Behebung entscheidet; der Test nimmt ihn nicht hin.
- Gemeinsamer Referenzdienst: die Ergänzungen dieses Slice dürfen den M2-Test nicht ändern — **Ausgang:** offen; der M2-Test läuft vor und nach dem Slice unverändert grün, die Closure nennt beide Läufe.
- Nebenläufigkeit: Proxy, Referenzdienst und Client laufen in einem Prozess — **Ausgang:** offen; die nebenläufigen Fälle laufen im Ziel mit `-race` aus `slice-race-gate`, wenn dessen Vertrag die Pakete unter `tests/` einschließt; sonst ordnen sie mit Kanälen, und jeder Fänger mit Wahrscheinlichkeit nennt seine Rate je Mutation und je Lauf von `go test` (`AGENTS.md` §3.10; Reviewer-Skill, MEDIUM).
- Randformen (`AGENTS.md` §3.12, §3.16): Raster *Operation × Argumentart* als Gerüst; jede Zelle nennt die Frage, der Ort ist `LH-QA-07.a`, den der Architect vor dem ersten Code-Commit nachzieht; eine Zelle ohne Ort hält den Implementer an — **Ausgang:** offen.

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Antwort vergleichen (Header) | relevanter Header nur auf einer Seite: Abweichung (`LH-QA-07.a`, zu bestätigen) | Header mit leerem Wert: gleich nur bei gleichem leeren Wert? (`LH-QA-07.a`) | Hop-by-Hop-Name aus `Connection` der Antwort: nicht relevant (`LH-QA-07.a`, `SPEC-121` Nr. 7); mehrere Werte je Name in anderer Reihenfolge: Abweichung? (`LH-QA-07.a`) | gilt nicht: der Vergleich läuft auf vollständig gelesenen Antworten — zu bestätigen | `VOLATILE_HEADERS` (`SPEC-063`) zwischen zwei Läufen verschieden: ausgenommen (`LH-QA-07.a`) |
| Antwort vergleichen (Body) | Antwort ohne Body (`HEAD`, 204, 304) auf beiden Seiten: gleich? (`LH-QA-07.a`) | leerer Body mit `Content-Encoding`: entpackt gleich? (`SPEC-119` Nr. 6, `LH-QA-07.a`) | nicht angefragtes Coding (Risiko oben); Kette mehrerer Codings (`SPEC-119` Nr. 6) | Abbruch der Antwort mitten im Body: auf beiden Seiten gleich behandelt? (`LH-QA-07.a`) | gilt nicht: jeder Fall sendet einen Request je Seite — zu bestätigen |
| Request senden | gilt nicht: der Test baut jeden Request selbst | `Accept-Encoding` fehlt: zulässiger Fall? (`LH-QA-07.a`) | Coding außerhalb von `gzip`, `deflate`, `identity` im Client: ausgeschlossen (`LH-QA-07.a`) | Upstream nicht erreichbar: direkt ein Verbindungsfehler, über den Proxy `ODR-E4001` — kein Vergleichsfall? (`LH-QA-07.a`) | gilt nicht: ein Request je Seite |

Zusatztabelle (`AGENTS.md` §3.16), Paare von Ursachen:

| Paar | Entscheidung | Fänger |
|---|---|---|
| anderes Coding über den Proxy und ein geänderter Body | offen: das Entpacken darf den Unterschied im Body nicht verdecken (`LH-QA-07.a`) | offen (Mutation „Proxy ändert ein Byte“ bei einer Antwort mit Coding) |
| Volatiler Header und relevanter Header gleichen Namens in anderer Schreibweise | offen (`LH-QA-07.a`, `SPEC-063`) | offen |

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

**Vorgelagert — offene Beobachtungen sichten:** Register durchgegangen (Sub-Area `*`, alle Einträge). Für diesen Slice zählen: `BEO-REPO/nebenlaeufiger-lauf-nur-im-glueck-gruen` (5×, Ausgang `geplant` in `slice-race-gate`; die Rate je Mutation und je Lauf steht schon in `AGENTS.md` §3.10) — der Test ist nebenläufig (§6, Risiko *Nebenläufigkeit*); `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (verkörpert in `AGENTS.md` §3.16) — „Status, Body und relevante Header unverändert“ ist eine Allaussage über alle Antwortarten und braucht ihren Fänger über alle Fälle des Referenzdienstes; `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (verkörpert in `AGENTS.md` §3.15) — der Vergleich selbst trägt Mutationen (§2); `BEO-REPO/randform-erst-im-review-entschieden` (verkörpert in `AGENTS.md` §3.16) — Raster in §6; `BEO-REPO/werkzeug-eingabe-setzt-escape-in-zeichen-um` (2×) — Testdaten mit Bytes ab `0x80` als Go-Escapes (§3); `BEO-REPO/groessenmarke-ueberschritten-ohne-rueckfuehrung` (verkörpert in `AGENTS.md` §3.14) — Marke und Schätzung in §4. Ohne Bezug: `BEO-REPO/akzeptiertes-negativ-ohne-folge` (2×; der Slice führt kein neues Negativ ein, außer der Architect entscheidet die Randform des nicht angefragten Codings als Ausschluss — dann mit Grund und Folge), `BEO-REPO/berichtigung-erreicht-nicht-jede-stelle` (2×) und `BEO-REPO/befund-an-rolle-ohne-artefakt` (1×).

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
