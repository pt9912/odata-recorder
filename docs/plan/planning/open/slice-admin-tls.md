# Slice slice-admin-tls: Transportschutz des Admin-Listeners

**Lifecycle:** Der Zustand dieses Slice ist das Verzeichnis, in dem diese
Datei liegt — eines von `open/`, `next/`, `in-progress/`, `done/`. Er
wechselt nur durch `git mv`, siehe
Baseline-Regelwerk `modul-05-planning-harness.md` §Lifecycle als State Machine.
Übernimmt ein anderer Slice den Gegenstand oder entfällt er, geht diese Datei
aus `open/` oder `next/` nach `done/` — §7 nennt in der Zeile `Gegenstand:`
Kennung oder Grund, die Liefer-Punkte der DoD bleiben leer
(§Ein Slice, dessen Gegenstand ein anderer übernimmt).

**Welle:** [welle-admin-api](../welle-admin-api.md).

**Bezug:** `LH-FA-13`, `LH-RB-02`, `LH-QA-04`, [ADR-0001](../../adr/0001-implementierungssprache-go.md)

**Berührte Spec-Stellen:** `LH-FA-13.a` (Transportschutz) · `LH-RB-02.a` (Listener-TLS) · `LH-QA-04.a` · `SPEC-008`, `SPEC-089`, `SPEC-090` · `SPEC-110` Nr. 10 (Muster der Zusammenhangs-Prüfung) · `ARC-007`.
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

**Ziel:** Der Admin-Listener bedient mit `ADMIN_TLS_CERT` und `ADMIN_TLS_KEY` HTTPS; bei nicht-loopback `ADMIN_LISTEN` verlangt die Start-Prüfung TLS und verhindert sonst die Ausführung (`ODR-E2001`), nur `ADMIN_ALLOW_PLAINTEXT=true` erlaubt Klartext, mit einer Warnung im Log; nur einer der beiden Schlüssel des Zertifikatspaars ist `ODR-E2001`.

**Übernimmt:** den wiederverwendbaren TLS-Baustein des Listeners aus `slice-listener-tls` (`welle-betrieb` §4 nennt ihn als Grundlage des Admin-Listeners); der Slice nutzt ihn, statt das Laden und Prüfen des Zertifikatspaars ein zweites Mal zu schreiben.

**Ausdrücklich NICHT in diesem Slice** — je Punkt mit Begründung:

- Der Listener und die Authentisierung selbst — `slice-admin-listener-token`; TLS umhüllt ihn.
- TLS des OData-Listeners (`LISTEN_TLS_*`) — `slice-listener-tls` (`welle-betrieb`).
- Erzeugen oder Erneuern von Zertifikaten, Client-Zertifikate (mTLS) und Neuladen zur Laufzeit — `SPEC-089` verlangt sie nicht; der Betreiber stellt das Paar bereit.
- Das Lesen der `ADMIN_*`-Schlüssel aus der Umgebung und das Aufheben von `SPEC-110` Nr. 2 — `slice-bootstrap-m6` (Grund wie in `slice-admin-listener-token`).
- Der Netzwerknachweis von außen — `slice-test-security`; hier prüfen Tests die Start-Prüfung und den Handshake.

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

- [ ] HTTPS-Betrieb: das Zertifikatspaar wird geladen und bedient; ein Client mit der CA gelingt, ein Client mit unbekannter CA lehnt ab; nur einer der beiden Schlüssel, eine unlesbare Datei, ein ungültiges oder nicht passendes Paar sind `ODR-E2001`.
- [ ] Start-Prüfung: ein nicht-loopback `ADMIN_LISTEN` (auch die Wildcard-Adressen) ohne TLS und ohne `ADMIN_ALLOW_PLAINTEXT=true` verhindert den Start (`ODR-E2001`); mit `ADMIN_ALLOW_PLAINTEXT=true` läuft Klartext mit einer Warnung im Log; Loopback ohne TLS bleibt zulässig.
- [ ] Je Zelle der Matrix Loopback × TLS × Plaintext-Flag ein Fänger (nicht nur die genannten Fälle); Schreibweise des Hosts (`localhost`, IPv6, Wildcard) nach dem Vertrag des Architect.
- [ ] `make gates` grün.
- [ ] Review durchgeführt, Report unter `docs/reviews/` liegt vor
      (`.harness/skills/reviewer.md`) — Rollenwechsel nach Schritt 8 des
      Minimal Agent Workflow (`AGENTS.md` §6), kein Self-Review (Modul 8).
- [ ] Doku-Update, falls ein öffentlicher Vertrag berührt ist.
- [ ] Bringt der Slice einen Fehlercode, Schlüssel oder ein Kommando in den Katalog des Binary: Handbuch-Kapitel im selben Slice nachgezogen (Ist-Zustand, belegt durch ein Beispiel als Datei gegen das gebaute Binary); das Kataloge-Gate ist grün. Sonst steht in §7, dass das Binary unverändert blieb (der Admin-Listener erreicht es erst mit `slice-bootstrap-m6`).
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
| `src/adapters/driving/admin/` | update | TLS-Konfiguration des Listeners, Start-Prüfung, Warnung bei Klartext |
| Baustein aus `slice-listener-tls` | nutzen | Laden und Prüfen des Zertifikatspaars; Ort und Namen nennt dessen Plan |
| `spec/spezifikation.md` | update (Architect) | Vertrag der Start-Prüfung: Loopback-Definition, Zusammenhänge, Reihenfolge der Verstöße; Nachzug `LH-FA-13.a`, `SPEC-089`, `SPEC-090` |
| Tests | neu | Zertifikatspaare im Test erzeugt (`crypto/x509`; kein Schlüssel im Repo); Matrix Loopback × TLS × Plaintext-Flag; Handshake mit bekannter und unbekannter CA |

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

1. **Vertrag der Start-Prüfung:** Eingaben (`ADMIN_LISTEN`, Zertifikatspaar, Flag), Ergebnis (TLS, Klartext mit Warnung, Verweigerung), Definition von Loopback (`127.0.0.1`, `::1`, `localhost`; `LH-FA-04.c` nennt diese drei für `PUBLIC_BASE_URL`; die ganze Schleife `127.0.0.0/8`?), Wildcard-Adressen (`0.0.0.0`, `::`) als nicht-loopback, und der Name, der nicht aufgelöst wird (`SPEC-110` Nr. 7).
2. **Zusammenhänge:** `ADMIN_ALLOW_PLAINTEXT=true` zusammen mit einem Paar (TLS gilt, oder Widerspruch), mit Loopback, ein ungültiger Wert des Flags (`SPEC-110` Nr. 3), Reihenfolge der Meldung mit anderen Verstößen.
3. **TLS-Parameter:** Mindestversion, Cipher und HTTP/2; `SPEC-089` nennt keine: Entscheidung mit Grund (Voreinstellung der Standardbibliothek oder ausdrücklich).
4. **Warnung im Log:** Wortlaut, Felder (ohne Geheimnis), einmal beim Start.
5. **Ort der Prüfung:** im Adapter oder im Composition Root; `slice-bootstrap-m6` ruft sie.
6. Tabelle *Zusage · Fänger*, Raster in §6, Nachzug der Stellen aus der Tabelle oben.

**Prüfung der Sendungen durch den Planner** (`AGENTS.md` §3.13, am 2026-10-30 gelesen: §1 *Ausdrücklich NICHT*, DoD und §4 der Gegenstellen): *an diesen Slice* der Baustein aus `slice-listener-tls`: `welle-betrieb` §4 nennt ihn als wiederverwendbar für den Admin-Listener, die Datei von `slice-listener-tls` liegt noch nicht vor, ihr Plan trägt die Lieferung (§6, Ausgang *offen*, Adresse Planner von `welle-betrieb`). *Von diesem Slice an:* das Lesen der `ADMIN_*`-Schlüssel an `slice-bootstrap-m6` (dessen §1 *Übernimmt* trägt es).

## 4. Trigger

Regeln dieser Sektion: Baseline-Regelwerk `modul-05-planning-harness.md`
§Trigger je Lifecycle-Übergang und WIP-Limit.

**Start** (`next` → `in-progress`): `slice-admin-listener-token` und `slice-listener-tls` (`welle-betrieb`) liegen in `done/`. Die Spezifikation trägt den Vertrag samt Randformen, bevor der erste Code-Commit entsteht (`AGENTS.md` §3.12).

**Warum:** der Listener, den TLS umhüllt, kommt aus `slice-admin-listener-token` (*Code-Abhängigkeit*); den wiederverwendbaren Baustein zum Laden und Prüfen des Zertifikatspaars liefert `slice-listener-tls` (*Code-Abhängigkeit*, soweit er so geliefert wird; liefert er ihn nicht, geht der Punkt an den Planner, §6).

**Unabhängig von `welle-mock`, M2 und `slice-bootstrap-m6`** — bis auf das, was `slice-listener-tls` selbst voraussetzt; dessen Plan liegt noch nicht vor und nennt es. Der Slice braucht den Audit-Port nicht (die Warnung geht ins Log, nicht ins Audit).

**Rückführungen — vorab benennen, nicht erst im Nachhinein begründen:**

- `in-progress` → `next` (zu groß, zurück zur Zerlegung): über der Marke, die der Planner beim Start vor dem ersten Code-Commit nach `AGENTS.md` §3.14 mit dem Werkzeug aus `slice-groessenmarke-werkzeug` setzt (Vertrag [SPEC-128](../../../../harness/vertraege/SPEC-128-groessenmarke.md); Metrik: Summe aus Einfügungen und Löschungen von `git diff --numstat <Basis>..<letzter Commit des Slice>` über die Pfade aus §3, Tests eingeschlossen). **Marke: noch nicht gesetzt**; eine Zahl steht hier erst, wenn die Tabelle *Zusage · Fänger* vorliegt, und bis das Werkzeug liegt, gilt die Formel aus `AGENTS.md` §3.14.
- `in-progress` → `open` (blockiert — Carveout?): der Baustein aus `slice-listener-tls` lässt sich für den Admin-Listener nicht nutzen oder ist nicht wiederverwendbar geschnitten (siehe §6).

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
| Start-Prüfung (`ADMIN_LISTEN`, Zertifikatspaar, Plaintext-Flag) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron beim Start ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| Zertifikatspaar laden (Dateien, Paar) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: synchron beim Start ohne Kontext (`SPEC-110` Nr. 13) | (Vertrag, offen) |
| HTTPS bedienen (Handshake) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) |
| Warnung bei Klartext | (Vertrag, offen) | (Vertrag, offen) | (Vertrag, offen) | gilt nicht: einmaliger Eintrag beim Start | (Vertrag, offen) |

- Die Definition von Loopback und der Umgang mit Namen, die nicht aufgelöst werden, haben keinen Ort (Auftrag 1) — **Ausgang:** offen; Adresse: Architect.
- Mindestversion und Cipher von TLS nennt `SPEC-089` nicht (Auftrag 3) — **Ausgang:** offen; Adresse: Architect.
- Der Baustein aus `slice-listener-tls` hat noch keinen Plan; sein Plan muss die Wiederverwendung tragen, sonst ist die Übernahme keine Adresse (`AGENTS.md` §3.13) — **Ausgang:** offen; Adresse: Planner von `welle-betrieb`.
- Zertifikate und Schlüssel der Tests dürfen nicht im Repo liegen (`harness/README.md` §Safety and scope boundaries) — **Ausgang:** entfallen: Die Tests erzeugen das Paar zur Laufzeit.
- `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×): die Zweige der Start-Prüfung sind eine Matrix; jede Zelle hat einen Lauf — **Ausgang:** weiter offen → `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` im Register.
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

**Vorgelagert — offene Beobachtungen sichten:** Register am Tag der Anlage gesichtet (2026-10-30, 19 Einträge); der Planner zieht Treffer und Zähler beim Start nach. Offen und für den Slice einschlägig: `BEO-REPO/zweig-ohne-lauf-unter-dem-detektor` (1×, Matrix der Start-Prüfung) und `BEO-REPO/betriebsregel-fuer-repo-dateien-verletzt` (1×). Verkörpert und mit dem Slice zu beachten: `BEO-REPO/vertrag-sagt-mehr-zu-als-ein-gate-prueft` (25×, `AGENTS.md` §3.15), `BEO-REPO/randform-erst-im-review-entschieden` (20×, `AGENTS.md` §3.16) und `BEO-REPO/akzeptiertes-negativ-ohne-folge` (10×, `AGENTS.md` §3.15). Keiner erreicht mit diesem Slice 3×.

**Modus-Begründungsblock — Umfang.** Pflicht, sobald mindestens eine berührte
Sub-Area BF oder Hybrid ist — einer pro Sub-Area. Bei reinem GF genügt der
Hinweis *"alle berührten Sub-Areas GF"*; bei reinem Refactor ohne neue
Sub-Area-Berührung entfällt **er** — nicht der Abschnitt.

Alle berührten Sub-Areas GF.
