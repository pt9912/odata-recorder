# Slice slice-hybrid-service: HybridUseCase mit Entscheidungsmatrix nach ODR-Code

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, `LH-FA-04`, `LH-FA-11`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` (Entscheidungsmatrix) · `LH-FA-04.c` (Adressierung, `X-Recorder-*`-Header) · `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080` · `SPEC-020`, `SPEC-021`, `SPEC-024` · `SPEC-120` (`UpstreamClient`, Fehler) · `SPEC-092` (`LINK_REWRITE`) · `ARC-002`, `ARC-003` · `architecture.md` §2.2, §2.5, §4.6.
Der Verweis zeigt **aufwärts**: Die Spec nennt diesen Slice nie.

**Verantwortlich:** —

**Autor:** pt9912. **Datum:** 2026-10-30.

---

## 1. Ziel und Abgrenzung

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Ziel-Form: Slice — Schnitt nach Lieferwert, nicht nach Schichten; jeder Slice
ist einzeln lieferbar. **§1 nennt Ziel und Abgrenzung** (Out-of-Scope-Disziplin
des Lastenhefts, auf den Slice-Plan angewandt); die vier Klassen des
Ausschlusses stehen in **eben diesem Abschnitt** des Baseline-Regelwerks,
zusammen mit der Begründungs-Pflicht je Punkt.

**Ziel:** `HybridUseCase` (Inbound Port in `src/hexagon/ports/inbound`, Dienst in `src/hexagon/services/hybrid`) ruft `MockUseCase` über dessen Inbound Port und entscheidet bei einem Fehler des Mock nach dem `ODR`-Code (nie nach dem HTTP-Status) mit der Policy aus `slice-fallback-policy`: Antwort oder Fehler bleiben (Erfolg, fachliche `404`, `409`, `412`, ungültige Anfrage), `ReplayUseCase` wird nur für die freigegebenen Kategorien gerufen, der `UpstreamClient` nur nach der Policy; ein Replay-Fehler außer `ODR-E5001`, insbesondere ein erschöpfter Key (`ODR-E5002`), wird unverändert ausgeliefert und löst nie einen Upstream-Zugriff aus (`LH-FA-12.b`).

**Übernimmt:** das Risiko der Upstream-Links einer Fallback-Antwort aus `welle-hybrid` §6 (`LINK_REWRITE` ist nach `SPEC-092` nur `off`); es steht als Risiko in §6 dieses Plans.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Die Policy (Hauptschalter, Allowlist, Routenfreigaben, Kategorien) und ihre Syntax — `slice-fallback-policy`; der Dienst ruft sie und entscheidet nichts selbst.
- Der Modus `hybrid` im Composition Root, die Konfiguration und der Listener — `slice-bootstrap-m4`; der Dienst läuft hier nur gegen Fakes der Ports.
- Speichern einer Fallback-Antwort und Datenabgleich zwischen Mock und Upstream — Out-of-Scope von `LH-FA-12`.
- Umschreiben absoluter Upstream-Links (`LINK_REWRITE`) — `SPEC-092` kennt nur `off`; das Risiko steht in §6.
- Spans, Audit und Metriken des Fallbacks — `welle-betrieb`; es gibt noch keinen Port `Telemetry`.

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

- [ ] Mock zuerst: Erfolg und jeder Fehler ohne Fallback-Auslöser (fachliche `404`, `409`, `412`, ungültige Anfrage, Fehler ohne `ODR`-Code) gehen unverändert an den Aufrufer; die Zuordnung (`X-Recorder-*`) geht unverändert an Mock und Replay.
- [ ] Replay-Fallback nur für die Kategorien der Policy (`ODR-E6001` und `ODR-E6002` als `unsupported`, `ODR-E6003` als `coverage_insufficient`): ein Treffer wird ausgeliefert, `ODR-E5001` geht weiter zur Entscheidung über den Upstream, jeder andere Replay-Fehler (Liste in `LH-FA-12.b`) unverändert; ein erschöpfter Key geht nie an den Upstream.
- [ ] Upstream-Fallback über `UpstreamClient` nur nach Freigabe der Policy: alle `X-Recorder-*`-Header entfernt, Body des Requests nach der Entscheidung des Architect, Antwort-Strom genau einmal geschlossen, ohne Freigabe kein Aufruf des Upstream (Zähler 0).
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Modus `hybrid` erreicht es erst mit `slice-bootstrap-m4`).
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
| `src/hexagon/ports/inbound/` | update | `HybridUseCase` |
| `src/hexagon/services/hybrid/` | update | Dienst neben der Policy aus `slice-fallback-policy` |
| `spec/spezifikation.md` | update (Architect) | Vertrag des Dienstes samt Randformen; Nachzug `LH-FA-12.b`, `architecture.md` §2.2 und §4.6 |
| Tests | neu | Matrix `LH-FA-12.b` × Zustände der Policy; die Fakes von `MockUseCase`, `ReplayUseCase` und `UpstreamClient` halten jedes Argument fest und zählen jede Operation, auch die, die der Dienst nie rufen darf (`AGENTS.md` §3.15); der Dienst importiert weder Adapter noch das Paket eines anderen Service |

Die Tabelle *Zusage · Fänger* (`AGENTS.md` §3.15) und das Raster *Operation × Argumentart* in §6 (`AGENTS.md` §3.16) entstehen mit dem Vertrag in der Spezifikation **vor** dem Übergang `next` → `in-progress` und vor dem ersten Code-Commit (`AGENTS.md` §3.12); die Randformen entscheidet der Architect, nicht der Implementer. Die Tabelle ist beim Anlegen des Plans leer:

| Zusage | Fänger |
|---|---|
| *(Architect: eine Zeile je Zusage, Randform und Fall des Vertrags)* | *(ein Test, ein Fall der Gegenprobe oder eine benannte Ausgabe des Gates)* |

**Betriebsregeln für alle Rollen dieses Slice** (Lehren der Läufe vom 2026-10-24; sie stehen im Plan, damit ein Lauf sie am Arbeitsort liest):

- Repo-Dateien nur mit `Edit` und `Write`: kein `sed -i`, kein Heredoc, kein `>>`, kein `awk`, kein Python; keine Hilfsdateien im Wurzelverzeichnis (`BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt`).
- Ein Wert mit Backslash-u-Folge, BOM, NBSP oder anderem Zeichen jenseits von ASCII entsteht aus Bytewerten, die geschriebene Datei wird auf Rohbytes gegengelesen (`LC_ALL=C grep -P`), der Bericht nennt sie (`AGENTS.md` §3.18).
- Jeder eigene `docker run` läuft mit `--memory=4g --memory-swap=4g --cpus=4` unter `flock /tmp/odata-recorder-heavy.lock`, höchstens zwei Container gleichzeitig, Mutationen nacheinander; `make gates` nie parallel zu Mutationsläufen (`BEO-REPO/gegenprobe-haengt-an-der-laufzeit-des-produkts`, `BEO-REPO/rate-ohne-detektor-schwankt-mit-der-last`). Docker sonst nur über `make` (`AGENTS.md` §3.1).
- Generator-Pflicht (`.claude/commands/implement-slice.md` Schritt 14): jede Klasse der Tabelle *Zusage · Fänger*, die ein Generator aus der Grammatik erzeugen kann, führt schon der Handoff als Generator oder Schleife mit unabhängigem Orakel und Zählprüfung. Stellen-Regel (Schritt 19, `AGENTS.md` §3.17): ein Fix bildet die Menge aus dem Wortlaut der Zusage und nennt Stellen und Elemente je Stelle.
- Der Handoff ist ein Commit für Code und Plan (`AGENTS.md` §3.9); Gegenprobe und Stempel dürfen folgen. Der Stempel `harness/gegenproben.stempel` entsteht erst im Integrationsschritt nach dem letzten Commit unter den Eingabepfaden (`BEO-REPO/stempel-gehoert-zu-keinem-committeten-stand`).
- Eine Abweichung vom Plan oder eine Rückgabe an eine andere Rolle steht im selben Commit in §6 mit Ausgang *offen* und der Adresse (`AGENTS.md` §3.13).
- `AGENTS.md`, `.claude/agents/`, `.claude/commands/` und `.harness/skills/` schreibt keine Rolle dieses Slice; Änderungen dort gehen als Diff-Vorschlag an den User.

**Aufträge an den Architect** (Entscheidungen und Texte, die kein Plan trifft; vor dem Übergang `next` → `in-progress`, `AGENTS.md` §3.12):

1. **Vertrag von `HybridUseCase`:** Signatur in der Form der Use Cases der anderen Modi, Fehlerabbildung nach `SPEC-109`, Ort in der Spezifikation.
2. **Ablauf und Rangfolge:** Mock, `ODR`-Code, Policy, Replay, Policy, Upstream; welche Stufe welcher Code auslöst (`LH-FA-12.b`), Fehler ohne `ODR`-Code, Abbruch des Kontextes zwischen den Stufen.
3. **Body des Requests:** der Mock liest den Strom, ein Upstream-Fallback eines Requests mit Body braucht ihn erneut. Entscheiden: Pufferung (nur wenn die Policy einen Fallback erlauben kann; Grenze und Fehlercode bei Überschreitung) oder Fallback nur für Requests ohne Body.
4. **Zuordnung:** `X-Recorder-Workspace` geht unverändert an Mock und Replay; `X-Recorder-Recording-Session` und `X-Recorder-Replay-Session` sind Pflicht, sobald ein Replay-Fallback konfiguriert ist (`LH-FA-04.c`): wann `ODR-E5003` entsteht (vor dem Aufruf des Mock oder erst im Fallback); alle `X-Recorder-*`-Header fallen vor dem Upstream weg.
5. **Antwort des Fallbacks:** ob sie sich als Fallback kennzeichnet (Header), welche Header der Upstream-Antwort weitergehen, und der Umgang mit absoluten Upstream-Links in der Antwort (`SPEC-092`, Risiko in §6).
6. **Fehler des Upstream im Fallback** (`ODR-E4001`, `SPEC-120` Nr. 8): unverändert ausliefern, kein Wiederholen, kein Rückfall auf Replay.
7. **Ströme:** die verworfene Antwort des Mock und jede Antwort, die nicht ausgeliefert wird, schließt der Dienst genau einmal, auch nach Abbruch.
8. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *von diesem Slice an* `slice-bootstrap-m4` (Verdrahtung des Dienstes im Composition Root): angenommen, §1 und DoD Punkt 1 von `slice-bootstrap-m4` tragen sie. *An diesen Slice:* das Risiko der Upstream-Links aus `welle-hybrid` §6, angenommen (§1 *Übernimmt*, §6).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-fallback-policy`, `slice-mock-lesen`, `slice-mock-schreiben` und `slice-replay-service` liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum diese vier:**

- `slice-fallback-policy`: Typen und Entscheidungsfunktion, die der Dienst ruft (*Code-Abhängigkeit*).
- `slice-mock-lesen`: legt den Inbound Port `MockUseCase` an (er liegt heute nicht in `src/hexagon/ports/inbound/`) und liefert `ODR-E6001` und `ODR-E6003`, nach denen der Dienst entscheidet (*Code-Abhängigkeit, Port*).
- `slice-mock-schreiben`: legt fest, welche Mutationen der Mock mit `ODR-E6002` als nicht unterstützt abweist; ohne diesen Vertrag hat die Zeile „POST auf eine nicht unterstützte Operation“ der Matrix keinen Boden (*Vertrags-Abhängigkeit*, keine Typ-Abhängigkeit).
- `slice-replay-service`: legt den Inbound Port `ReplayUseCase` an (heute nicht vorhanden) und liefert `ODR-E5001`, `ODR-E5002` und `ODR-E5003` (*Code-Abhängigkeit, Port*).

**Nicht gebraucht:** `slice-bootstrap-m3`, `slice-snapshot-use-case`, `slice-mock-abnahme` und M3, denn der Dienst läuft gegen Fakes der Ports und der Composition Root gehört `slice-bootstrap-m4`; auch M2 (`slice-bootstrap-m2`) braucht der Dienst nicht selbst, nur das, was `slice-replay-service` und `slice-mock-lesen` von ihr voraussetzen. **Ohne die Ports des Mock startet der Slice nicht.**

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): der Body eines Requests lässt sich für den Upstream-Fallback nicht erhalten, ohne den Port `MockUseCase` zu ändern (Auftrag 3), oder die Zuordnung der Replay-Header im Hybrid-Modus ist nicht entschieden (siehe §6).

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

**Raster *Operation × Argumentart*** (`AGENTS.md` §3.16): „(Vertrag, offen)“ heißt, der Architect trägt den Ort in der Spezifikation vor dem Übergang `next` → `in-progress` ein; eine Zelle bleibt nie leer, und der Implementer hält an einer Zelle ohne Ort an (`AGENTS.md` §3.12).

| Operation | fehlender Wert | leerer Wert | unbekannter oder fremder Wert | Abbruch | Wiederholung |
|---|---|---|---|---|---|
| Aufruf des Mock und Entscheidung nach dem Fehlercode | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Replay-Fallback (Zuordnungs-Header, Kategorie, Replay-Fehler) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Upstream-Fallback (Body, `X-Recorder-*`, Antwort des Upstream) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Schließen der Ströme (verworfene Mock-Antwort, ausgelieferte Antwort) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |

- Der Body eines Requests lässt sich nur einmal lesen; der Mock liest ihn, ein Upstream-Fallback braucht ihn erneut — **Ausgang:** offen; Adresse: Architect (Auftrag 3).
- Eine Fallback-Antwort kann absolute Upstream-Links tragen, `LINK_REWRITE` kennt nur `off` (`SPEC-092`; Risiko aus `welle-hybrid` §6) — **Ausgang:** offen; Adresse: Architect (Auftrag 5): Kennzeichnung der Antwort und Hinweis im Handbuch-Kapitel von `slice-bootstrap-m4`.
- Die Zuordnung der Replay-Header im Hybrid-Modus (`ODR-E5003` vor dem Mock-Aufruf oder erst im Fallback) hat keinen Ort in `LH-FA-04.c` — **Ausgang:** offen; Adresse: Architect (Auftrag 4).
- `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×) und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×): die Fakes von `MockUseCase`, `ReplayUseCase` und `UpstreamClient` tragen jede Schnittstelle des Originals (der Body als `ReadCloser`, `Close` gezählt), zählen jeden Aufruf und halten jedes Argument fest; ein Fake, der panikt, verdeckt den benannten Test nicht — **Ausgang:** weiter offen → `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` und `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` im Register.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): die Matrix hat viele Zweige, der Dienst hält keinen Zustand; gleichzeitige Aufrufe laufen als Fall unter `make test-race` — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
- Größe: die Marke steht nicht, bevor die Tabelle *Zusage · Fänger* vorliegt — **Ausgang:** offen; Adresse: Planner beim Start (§4).

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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Der Slice hat drei Fakes und viele Zweige: `BEO-REPO/hilfe-ersetzt-argument-verdeckt-eigenschaften` (2×), `BEO-REPO/panik-eines-fakes-verdeckt-den-benannten-test` (2×) und `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×) sind offen (§6). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15) und `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16). Offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
