# Slice slice-fallback-policy: Fallback-Policy als reine Entscheidung

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-hybrid](../welle-hybrid.md).

**Bezug:** `LH-FA-12`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-12.a` · `LH-FA-12.b` (Entscheidungsmatrix) · `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080` · `SPEC-020`, `SPEC-021`, `SPEC-022`, `SPEC-023`, `SPEC-024` (die Codes, nach denen die Matrix entscheidet) · `SPEC-110` Nr. 2 (die Schlüssel, die diese Version noch ablehnt) · `ARC-002` · `architecture.md` §2.5 (Modus-Routing).
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

**Ziel:** Eine reine Entscheidung (Paket unter `src/hexagon/services/hybrid`, der Ort ist Sache des Architect) legt aus Methode, Zielhost, Pfad, dem `ODR`-Code des Ergebnisses und den typisierten Werten der Policy (Hauptschalter, Allowlist der Hosts, Routenfreigaben für mutierende Methoden, Kategorien des Replay-Fallbacks) fest, ob die Antwort bleibt, ein Replay-Fallback oder ein Upstream-Fallback erlaubt ist; ohne ausdrückliche Freigabe lautet jede Entscheidung „kein Fallback“, und sie hängt am `ODR`-Code, nie am HTTP-Status (`LH-FA-12.b`).

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Aufruf von `MockUseCase`, `ReplayUseCase` und `UpstreamClient`, Strom und Zustand — `slice-hybrid-service`; die Policy ist eine Funktion ohne Port, damit die Matrix ohne Fakes vollständig prüfbar ist und kein Zweig von einem Fake abhängt.
- Das Lesen der Schlüssel `HYBRID_FALLBACK`, `HYBRID_REPLAY_FALLBACK`, `FALLBACK_ALLOW_HOSTS` und `FALLBACK_WRITE_ALLOW` aus der Umgebung und das Aufheben ihrer Ablehnung in `SPEC-110` Nr. 2 — `slice-bootstrap-m4`; solange kein Composition Root die Policy verdrahtet, bliebe ein gelesener Schlüssel eine still wirkungslose Sicherheitseinstellung. Die Policy nimmt typisierte Werte, nie die Umgebung; Syntax und Prüfung eines Eintrags der Routenfreigabe (Methode, Pfadmuster) gehören dagegen zu ihr, `slice-bootstrap-m4` ruft sie beim Lesen der Liste.
- Das Senden an den Upstream und das Speichern einer Fallback-Antwort — `slice-hybrid-service` (Senden); Speichern und Datenabgleich sind Out-of-Scope von `LH-FA-12`.
- Spans, Audit und Metriken der Entscheidung — `welle-betrieb`; es gibt noch keinen Port `Telemetry`.

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

- [ ] Entscheidungsmatrix `LH-FA-12.b`: je Zeile (Erfolg, fachliche `404`/`409`/`412`, Coverage unvollständig, nicht unterstützte Operation, ungültige Anfrage, Replay-Miss) und je Zustand der Policy das Ergebnis; Auslöser ist der `ODR`-Code, nie der HTTP-Status; unter den Replay-Fehlern löst nur `ODR-E5001` einen Fallback aus.
- [ ] Upstream-Freigabe: nur bei Hauptschalter `enabled`, nur zu einem Host aus `FALLBACK_ALLOW_HOSTS`, der der Host der konfigurierten `UPSTREAM_URL` ist, lesend nur mit GET und HEAD, mutierend nur mit passendem Eintrag aus `FALLBACK_WRITE_ALLOW` (Methode und Pfadmuster); sonst nichts.
- [ ] Replay-Fallback: nur für die Kategorien `unsupported` und `coverage_insufficient` aus `HYBRID_REPLAY_FALLBACK`, unabhängig vom Hauptschalter; Syntax des Eintrags der Routenfreigabe und Vergleich von Host und Pfad nach dem Vertrag des Architect.
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
| `src/hexagon/services/hybrid/` | neu | Policy als reine Funktion über `model`-Typen und typisierte Werte (Ort und Namen: Architect) |
| `spec/spezifikation.md` | update (Architect) | Vertrag der Policy samt Randformen; Nachzug an jeder Stelle, die dieselbe Aussage trägt (`AGENTS.md` §3.9): `LH-FA-12.a`, `LH-FA-12.b`, `SPEC-010`, `SPEC-062`, `SPEC-079`, `SPEC-080`, `architecture.md` §2.5 |
| Tests | neu | Matrix `LH-FA-12.b` als Tabelle × Zustände der Policy; Generator über Methoden, Hosts und Muster mit unabhängigem Orakel; die Policy importiert weder Port noch Adapter (`make a-check`) |

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

1. **Vertrag der Policy** in der Spezifikation: Eingaben (Methode, Zielhost, Pfad, `ODR`-Code oder Kategorie des Ergebnisses, Werte der Policy), Ergebnis (bleiben, Replay-Fallback, Upstream-Fallback) und die Fehler der Prüfung eines Eintrags; Ort und Namen der Typen (Paket unter `src/hexagon/services/hybrid`, `model` oder ein eigenes Paket), die `HybridUseCase` und der Composition Root rufen.
2. **Pfadmuster von `FALLBACK_WRITE_ALLOW`** (`SPEC-080` nennt nur „Methode und Pfadmuster“): Form eines Eintrags, Platzhalter, Bezugspfad (OData-Pfad nach dem Service-Root oder voller Pfad, mit oder ohne Query), Behandlung von Escapes, Schreibweise der Methode, doppelte und leere Einträge.
3. **Host-Vergleich** (`SPEC-062`: nur zum Host der konfigurierten `UPSTREAM_URL`, und nur wenn er in `FALLBACK_ALLOW_HOSTS` steht): Form eines Eintrags (Host oder Host mit Port), Schreibweise, IPv6, und was ein Eintrag bedeutet, der nicht der Host der `UPSTREAM_URL` ist.
4. **Methodenklassen:** lesend sind GET und HEAD (`LH-FA-12.b`); OPTIONS, TRACE, CONNECT und unbekannte Methoden gelten als lesend, mutierend oder nie freigegeben, mit Grund.
5. **Auslöser:** welche `ODR`-Codes die Kategorien `unsupported` (`ODR-E6001`, `ODR-E6002`) und `coverage_insufficient` (`ODR-E6003`) bilden; ein Ergebnis ohne `ODR`-Code (ungültige Anfrage, Kontextfehler, `ODR-E9001`) löst keinen Fallback aus; die Zeile „Replay-Miss“ gilt nur für `ODR-E5001`.
6. **Zusammenhänge der Werte** für `slice-bootstrap-m4`: Hauptschalter `disabled` bei gesetzten `FALLBACK_*`-Schlüsseln, `enabled` mit leerer Allowlist, ein Eintrag von `FALLBACK_WRITE_ALLOW` ohne Host in der Allowlist; Verhalten der Policy und ob der Start sie als `ODR-E2001` abweist (`SPEC-110` Nr. 10 als Muster).
7. **Replay-Fallback und Methode:** `LH-FA-12.b` nennt den Replay-Fallback „nur explizite Policy“ ohne Bezug auf die Methode; gilt er für alle Methoden oder nur für lesende? Die Zelle hat heute keinen Ort in `SPEC-079`.
8. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstelle): *von diesem Slice an* `slice-bootstrap-m4` (Lesen der vier Schlüssel und Aufruf der Syntax-Prüfung): angenommen, §1 *Übernimmt* und DoD Punkt 1 von `slice-bootstrap-m4` tragen sie mit der Kennung des Gebers, kein *Ausdrücklich NICHT* dort trifft sie. *An diesen Slice:* keine.

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): kein Slice muss in `done/` liegen. Die Spezifikation trägt den Vertrag samt Randformen und der Syntax von Pfadmuster und Host-Vergleich, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Code-Abhängigkeit: keine.** Die Policy liest nur `model` (die Codes `CodeE5001` und `CodeE6001` bis `CodeE6003` stehen seit `slice-fehlermodell-konfiguration` in `src/hexagon/model/fehler.go`) und typisierte Werte; sie ruft keinen Port.

**Unabhängig von `welle-mock`, M2 und `welle-record-replay`:** die Policy braucht weder `MockUseCase` noch `ReplayUseCase` noch `UpstreamClient` und kann auf dem heutigen Stand von `src/` entstehen. Der Trigger der Welle (`welle-hybrid` §2: `welle-mock` in `done/`, M3) hält sie zurück, ohne dass der Code es verlangt; ein früherer Start ist eine Umplanung mit Eintrag im Drift-Log der Roadmap und nicht Inhalt dieses Plans.

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): eine Zelle des Rasters (Pfadmuster, Host-Vergleich, Methodenklasse) hat keinen Ort in der Spezifikation, und der Architect entscheidet sie nicht vor dem Code (siehe §6).

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
| Entscheidung über Replay-Fallback (Code, Kategorien, Methode) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: reine Funktion ohne `ctx` | (Vertrag, offen) |
| Entscheidung über Upstream-Fallback (Methode, Host, Pfad, Code, Werte) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: reine Funktion ohne `ctx` | (Vertrag, offen) |
| Prüfung eines Eintrags von `FALLBACK_WRITE_ALLOW` (Methode und Pfadmuster) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: reine Funktion ohne `ctx` | (Vertrag, offen) |
| Vergleich des Hosts mit `FALLBACK_ALLOW_HOSTS` und `UPSTREAM_URL` | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: reine Funktion ohne `ctx` | (Vertrag, offen) |

- Pfadmuster und Host-Vergleich haben in `SPEC-062` und `SPEC-080` keine Syntax — **Ausgang:** offen; Adresse: Architect (Aufträge 2 und 3), vor dem ersten Code-Commit.
- Methoden außerhalb von GET, HEAD, POST, PUT, PATCH und DELETE (Auftrag 4) und der Replay-Fallback für mutierende Methoden (Auftrag 7) haben keinen Ort — **Ausgang:** offen; Adresse: Architect.
- Der Replay-Fallback hat im Lastenheft weder Anforderung noch Akzeptanzkriterium ([Abgleichsbericht](../../../reviews/2026-10-30-lastenheft-wellen-abgleich.md) §8 Nr. 2); die Policy leitet ihn aus `LH-FA-12.a` und `LH-FA-12.b` der Spezifikation ab — **Ausgang:** weiter offen als benannte Lücke des Lastenhefts; Adresse: User (Change Request vor `Accepted` oder Bestätigung).
- Eine Policy, die in einem Zustand fehlt, den die Matrix nicht nennt (unbekannter Code, unbekannte Methode, leere Werte), erlaubt nie einen Fallback; der Fänger ist eine Schleife über Methoden × Zustände × Codes mit Zählprüfung, die Mutation „Standardzweig erlaubt“ färbt ihn rot — **Ausgang:** offen bis zur Closure.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): die Policy hat keine Nebenläufigkeit, der Detektor hat nichts zu melden; jeder Zweig der Matrix hat einen Fall der Tabelle — **Ausgang:** entfallen: reine Funktion ohne Zustand, Zählprüfung der Fälle gegen die Zeilen von `LH-FA-12.b` statt eines Laufs unter dem Detektor.
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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Reine Funktion mit großer Matrix: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×, offen) gilt nur eingeschränkt (§6). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16), `BEO-REPO/universalaussage-im-vertrag-ungeprueft` (9×, `AGENTS.md` §3.16, für „kein Fall …“ der Matrix) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Offen außerdem `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×); keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
