# Harness-Konventionen

---

## Purpose

Diese Datei deklariert die *repo-lokalen* Strukturregeln dieses Repos
gegenüber der adoptierten Harnesskonvention (Baseline). Sie ist der
Default-Ort für:

- **Adaptionen** ggü. der Baseline (mit Begründung und Auflösungs-Trigger).
- **ID-Schema-Deklaration** — welches Präfix-Schema dieses Repo nutzt.
  Der Baseline-Default wird als Teil der `MR-000`-Aussage festgehalten;
  ein abweichendes Präfix oder Schema ist ein eigener `MR`-Eintrag.
- **Zusatzklassen-Deklarationen** für repo-spezifische
  Bindung-Klassen in der Sensors-Tabelle, die über die vier kanonischen
  hinausgehen (ADR, Carveout, Schwelle, Reproduzierbarkeit).
- **Modus-Deklarationen** pro Sub-Area (Greenfield / Brownfield /
  Hybrid) inklusive Konvergenz-Auftrag bei BF.

Bei Konflikt zwischen dieser Datei und einer kanonischen Quelle gilt die
kanonische Quelle (Source Precedence). Diese Datei ist konformitäts-
bringend für *Form*-Fragen, nicht autoritativ über Inhalt.

## Baseline



- **Konvention:** AI-Harness-Kurs (`pt9912/ai-harness-course`)
- **Stand:** `v6.18.1` (Kurs-Welle 161 · 2026-10-10)
- **Datum der Adoption:** 2026-10-18



## Adoptierte Konventions-Quellen



- **Extern (Lehrmaterial):** https://github.com/pt9912/ai-harness-course
- **Vendored Baseline (Regelwerk + Templates):** aus dem self-contained
  Release-Asset
  https://github.com/pt9912/ai-harness-course/releases/download/v6.18.1/lab-regelwerk.zip
  nach `.harness/baseline/<tag>/{regelwerk,templates}/` entpackt (netzlos,
  `SHA256SUMS`) — adoptierten Stand notieren (Stand-Zeile in
  `regelwerk/README.md`, z. B. „Kurs-Welle 24 · 2026-07-16"; Wellen-Register:
  CHANGELOG.md im Kurs-Repo); für harte Reproduzierbarkeit das Asset eines Tags
  ziehen statt `latest`.
- **In-Repo (verkörperte Form):** [`AGENTS.md`](../AGENTS.md),
  [`harness/README.md`](README.md), diese Datei und die Spec-Dateien unter
  [`spec/`](../spec/) — die vendored `.harness/baseline/<tag>/templates/` sind die
  Referenz-Form („Ziel-Form" des Regelwerks); deine eigenen Dateien sind daraus
  kopiert und ausgefüllt.

### Was der mitgelieferte Baum ist — und was er nicht verspricht

Der Baum unter `.harness/baseline/` ist Kurs-Inhalt und in diesem Repo nicht
autoritativ: er wird byte-genau so mitgeliefert, wie der Kurs ihn veröffentlicht, und
vom Doku-Gate ausgenommen — `scan.ignore` in `.d-check.yml` nennt ihn. Was er an
`make`-Namen nennt, sind **Beispiele des Kurses**, keine Ziele dieses Repos —
maßgeblich ist allein `make help`. Dasselbe gilt
für die Vorlagen darunter: wer eine kopiert, prüft ihre Ziel-Namen gegen `make help`,
bevor er sie in ein lebendes Dokument übernimmt.

### Der mitgelieferte Baum altert still

Der Baum ist auf einen Tag gepinnt — sein Verzeichnisname unter `.harness/baseline/`
ist dieser Tag, und `make baseline-verify` nennt ihn in seiner Ausgabe. Erscheint
upstream ein neueres Release, ändert sich hier nichts. `make baseline-verify` hält den
Baum gegen `SHA256SUMS` — Integrität und Vollständigkeit, netzlos; über sein **Alter**
sagt er nichts.

Der Freshness-Audit ist deshalb eine **geschuldete Handlung dieses Repos**, kein Lauf:
die **Release-Liste** des Kurs-Repos gegen den gepinnten Tag halten. Die Liste, nicht
das Asset — ein Asset-Abgleich sagt, ob sich dieser Tag änderte, nicht, ob ein neuerer
existiert. Die Adresse ist die **Release-Übersicht des Repos**, aus dem das Asset im
Abschnitt *Adoptierte Konventions-Quellen* darüber stammt — nicht jene Asset-URL
selbst, die genau einen Tag nennt. Der Audit ist im mitgelieferten Regelwerk
ausgeschrieben.

**Ein Sensor dafür kommt nicht mit, und das ist eine Aussage.** Er bräuchte einen
Netz-Abruf der Release-Liste und damit eine Host-Abhängigkeit über `bash`, `git` und
`docker` hinaus. Ein Repo, das nur diese drei voraussetzt, bekommt ihn nicht; was hier
läuft, ist `make baseline-verify`, und das ist die Integritäts-Hälfte.

### Welche Regelblöcke des Baums hier einen Träger haben

Je Regelblock des mitgelieferten Regelwerks genau einer von drei Werten — *Träger
kommt mit* · *liegt bei, nicht verdrahtet* · *kommt nicht mit* (mit Grund und Dauer).
Die Zelle sagt den Zustand **dieses** Repos, nicht den Stand einer Entscheidung.
Wie viele Regelblöcke der Baum führt, sagt `ls .harness/baseline/*/regelwerk/*.md`.

**Gesagt ist, was ein frisches Repo bekommt.** Ein Teil dieser Adressen gehört dem
Adopter: dort legt der Bootstrap nur ab, wo nichts liegt, und eine vorhandene
Fassung überlebt jeden weiteren Lauf unberührt. **Diese Pfade nennt der Lauf:**
`.githooks/commit-msg`, `harness/sensors/.gitkeep`, `repo.mk`, `harness/mk/.gitattributes`, `.claude/hooks/.gitattributes` und `.githooks/.gitattributes`, sobald dort eine Datei liegt;
und die Skills `.harness/skills/reviewer.md`
und `.harness/skills/closure-note-reviewer.md`, sobald die liegende Fassung von der
mitgelieferten abweicht — samt der Vorlage im vendored Baum zum Abgleich. Für jeden
anderen schweigt er:
für die Dokumente der Doku-Kette, die Spec-Dateien, die Roadmap, `README.md`,
`.d-check.yml`, die Rollen-Typen und die Workflow-Commands steht in der Ausgabe
nichts — gleichgültig, ob er sie geschrieben oder stehen gelassen hat. Ohne Ausgabe
erkennbar ist es am Inhalt: eine Datei, die der Bootstrap geschrieben hat, trägt
seinen Text; eine, die er stehen ließ, den dieses Repos.

**Gemessen gegen den Kurs-Stand `v6.18.1`, und die Tabelle wandert nicht mit.** Sie
nennt Regelblöcke beim Namen; ein Baseline-Sprung kann einen umbenennen, hinzufügen
oder wegnehmen. Dieses Dokument wird von einem erneuten Bootstrap **nicht**
überschrieben, und kein Lauf dieses Repos hält die Tabelle gegen den Baum, der hier
liegt — wer den Baum tauscht, prüft sie von Hand gegen das Kommando oben.

| Regelblock | Wert | Träger bzw. Grund und Dauer |
|---|---|---|
| `README.md` | liegt bei, nicht verdrahtet | Der Index des Regelwerks liegt im vendored Baum und wird bei Bedarf gelesen; ein Injektor, der ihn je Sitzung in den Kontext hebt, kommt nicht mit. |
| `grundlagen-begriffe.md` | liegt bei, nicht verdrahtet | Begriffs-Definitionen ohne eigene Mechanik — der Text ist sein eigener Träger und hängt an keinem Trigger. |
| `grundlagen-bootstrap.md` | Träger kommt mit | `harness/conventions.md` führt den Abschnitt *Modus-Deklaration pro Sub-Area*, in dem Sub-Area, Kürzel, Modus und Graduation deklariert werden. |
| `grundlagen-durchsetzungsschicht.md` | Träger kommt mit | `.claude/hooks/pretooluse-command-guard.sh`, `.claude/hooks/stop-require-gates.sh`, `tools/harness/record-gates.sh`, `tools/harness/working-tree-hash.sh` und der Eintrag in `.claude/settings.json`. |
| `grundlagen-harness-dateien.md` | Träger kommt mit | `AGENTS.md`, `harness/README.md` und `harness/conventions.md` liegen als ausgefüllte Dateien, nicht als Vorlagen. |
| `grundlagen-klassifikation.md` | liegt bei, nicht verdrahtet | Die Einordnung von Sensoren und der Steering Loop sind Lesestoff; kein Artefakt des Ziels hängt daran. |
| `grundlagen-referenz-richtung.md` | Träger kommt mit | Das Doku-Gate führt die Klasse `spec-straten` mit `direction: no-downward` in `.d-check.yml`; `make docs-check` fährt sie. |
| `grundlagen-source-precedence.md` | Träger kommt mit | `harness/README.md` §Source precedence und der Kopf von `AGENTS.md` tragen die Rangfolge. |
| `grundlagen-traceability.md` | Träger kommt mit | Die Prüfung `tools/harness/commit-msg-traceability.sh` liegt in jedem Lauf. Der Träger `.githooks/commit-msg`, der sie ruft, kommt nur an einem freien Pfad — sein Name gehört git, und führt das Repo dort schon einen eigenen Hook, bleibt der stehen und ruft die Prüfung nur, wenn er es selbst tut. Aktiviert wird der Träger durch `make hooks-install`; bis dahin liegt er unwirksam da. |
| `modul-00-einfuehrung.md` | liegt bei, nicht verdrahtet | Einführung ohne eigene Mechanik — der Text ist sein eigener Träger und hängt an keinem Trigger. |
| `modul-01-entwicklungszyklus.md` | Träger kommt mit | Die Ziel-Form des Moduls ist der Source-Precedence-Block, und er liegt ausgefüllt in `harness/README.md`. |
| `modul-02-harness-bootstrap.md` §Gate-Fragment und vendored Baseline | Träger kommt mit | `tools/harness/baseline-verify.sh` mit dem Fragment `harness/mk/baseline.mk` hängt `baseline-verify` an die Gate-Kette; das Doku-Gate-Fragment liegt daneben. |
| `modul-02-harness-bootstrap.md` §Freshness-Audit der vendored Baseline | kommt nicht mit | Kein Sensor prüft, ob upstream ein neuerer Stand erschienen ist — er bräuchte Netzzugriff über `bash`, `git` und `docker` hinaus. Dauerhaft, solange diese drei die einzigen Host-Abhängigkeiten sind; der Audit ist die Handlung, die der Abschnitt darüber beschreibt. |
| `modul-03-spec.md` | Träger kommt mit | `spec/lastenheft.md`, `spec/spezifikation.md` und `spec/architecture.md` liegen als ausgefüllte Dateien; die Klasse `spec-straten` des Doku-Gates hält ihre Richtung. |
| `modul-04-adrs.md` | Träger kommt mit | `docs/plan/adr/` ist angelegt, die ADR-Vorlage liegt im mitgelieferten Vorlagen-Baum, und `.d-check.yml` verlangt für jede ADR-Kennung einen auflösenden Link. |
| `modul-05-planning-harness.md` | Träger kommt mit | Die vier Lifecycle-Verzeichnisse unter `docs/plan/planning/` sind angelegt; `make slice-mv` bewegt einen Slice und zieht seine Verweise nach. |
| `modul-06-roadmap.md` | Träger kommt mit | `docs/plan/planning/in-progress/roadmap.md` und die Register-Ablage `docs/plan/planning/observations/` liegen; `make archive-welle` archiviert die Zeitdokumente einer geschlossenen Welle. |
| `modul-07-carveouts.md` | liegt bei, nicht verdrahtet | `docs/plan/carveouts/` ist angelegt und die Carveout-Vorlage liegt im mitgelieferten Vorlagen-Baum; kein Sensor prüft Frist oder Auflösungs-Trigger. |
| `modul-08-agentenrollen.md` | Träger kommt mit | Unter `.claude/agents/` liegt je ein Rollen-Typ für die sechs kanonischen Rollen; der Typname trägt die Rolle in den Span. |
| `modul-09-implementierung.md` | Träger kommt mit | `.claude/commands/implement-slice.md` führt den 8-Schritt-Workflow der Implementation-Rolle. |
| `modul-10-review-harness.md` | Träger kommt mit | `.harness/skills/reviewer.md` trägt die Urteilsgrundlage der Review-Rolle. |
| `modul-11-verification.md` | liegt bei, nicht verdrahtet | `.harness/skills/closure-note-reviewer.md` liegt im Ziel; ein eigenes Verifikations-Ziel führt der Aggregator nicht — die Prüfung hängt am Lauf, nicht an der Gate-Kette. |
| `modul-12-replay-evaluierung.md` | kommt nicht mit | Weder Golden Set noch Replay-Gate kommen mit — beide sind Gegenstand der Domäne dieses Repos, nicht des Bootstraps. Dauerhaft: ein repräsentatives Golden Set kann kein Werkzeug von außen setzen. |
| `modul-13-quality-gates.md` | Träger kommt mit | `make gates` aggregiert die Fragmente unter `harness/mk/`; jedes hängt seine Prüfungen an `GATE_CHECKS`, der Nachweis läuft zuletzt. |
| `modul-14-docker-harness.md` | Träger kommt mit | Das Doku-Gate fährt in einem per Digest gepinnten Image; Tag und Digest stehen in `d-check.mk`. |
| `modul-15-observability.md` §Erfassung und Token-Attribution | Träger kommt mit | `harness/mk/erfassung.mk` trägt `span-report` und `span-clean` — beide in jedem Lauf. Der schreibende Teil hängt an einer Bedingung: Träger, Hook-Wrapper `.claude/hooks/span-emit.sh`, Hook-Eintrag und die Feldliste `harness/erfassung-feldliste.md` entstehen nur, wenn die Ablage des Trägers gelang; misslingt sie, endet der Bootstrap dennoch erfolgreich und diese vier fehlen. |
| `modul-15-observability.md` §Doku-Konsistenz-Drift | liegt bei, nicht verdrahtet | Das Doku-Gate bringt Module mit, die Ziel-Ansprüche und Code-Pfade prüfen; die Zeile `modules:` in `.d-check.yml` führt sie nicht — ihre Aktivierung ist eine eigene Entscheidung. |
| `modul-16-produktiver-betrieb.md` | liegt bei, nicht verdrahtet | Betriebs-Regeln ohne eigene Mechanik im Bootstrap — der Text ist sein eigener Träger und hängt an keinem Trigger. |

## Adaptions-Block

Regeln dieser Sektion: Diese Datei trägt den **Index**, nicht die Einträge.
Jede Adaption ist eine eigene Datei unter `harness/conventions/`, kopiert aus
`.harness/baseline/<tag>/templates/harness/conventions/MR-NNN-titel.template.md` der vendored Baseline;
ist ihr Auflösungs-Trigger eingetreten, wandert sie per `git mv` nach
`conventions/done/`. Der Zustand ist die Verzeichnis-Position, kein
Status-Feld. Der Grund für den Schnitt: Was hier steht, liest **jeder**
Agentenlauf — aufgelöste Adaptionen gehören nicht in diesen Pfad
(Baseline-Regelwerk `grundlagen-harness-dateien.md`
§harness/conventions.md als Konventionsspeicher).

### MR-000 — Baseline-Aussage

Bleibt hier: Sie ist keine Adaption, sondern die Adoptions-Erklärung, und
sie gilt für jeden Lauf.

- **Datum:** 2026-10-18
- **Geltungsbereich:** gesamtes Repo
- **Ersetzt-Baseline-Regel:** — *(keine; dieser Eintrag ist die
  Adoptions-Erklärung, keine Adaption)*
- **Adaption:** *keine inhaltlichen Adaptionen ggü. Baseline-Default
  für Verzeichniskonvention, Lifecycle-Regeln, Carveout-Disziplin,
  ID-Schema (`LH-FA-*`, `LH-QA-*`, `LH-RB-*`, `SPEC-<NNN>`, `ARC-<NNN>`,
  `ADR-<NNNN>`, `CO-<NNN>`, `slice-<Kennung>`, `MR-<NNN>`, `BEO-<KUERZEL>/<slug>`, `RC-<NNN>` — nur das
  Vertrags-Präfix wird repo-weit festgelegt, z. B. `LH`; `SPEC-*` und
  `ARC-*` kodieren das Stratum und sind fest, siehe Baseline-Regelwerk
  `grundlagen-source-precedence.md` §ID-Schema als Klammer;
  bei mehreren gleichzeitig schreibenden Entwicklern für die Artefakte mit
  je eigener Datei (`ADR-*`, `CO-*`, `slice-*`, `welle-*`) zusätzlich das
  Bereichssegment und damit den Zählraum je Sub-Area festlegen —
  `SPEC-*`/`ARC-*` bleiben davon ausgenommen und zählen fortlaufend je
  Datei, siehe Baseline-Regelwerk `grundlagen-source-precedence.md` §Vergabe).*
- **Begründung:** Initial-Setzung. Spätere Adaptionen werden als
  `MR-<NNN>` nachgetragen.
- **Auflösungs-Trigger:** permanent.

### Aktive Adaptionen



| MR | Titel | Geltungsbereich | Ersetzt-Baseline-Regel |
|---|---|---|---|
| <a id="mr-001"></a>[MR-001](conventions/MR-001-harness-vertraege-ausserhalb-der-spezifikation.md) | Verträge der Harness-Werkzeuge im Ordner `harness`, nicht in der Produktspezifikation ([ADR-0004](../docs/plan/adr/0004-vertraege-der-harness-werkzeuge-im-ordner-harness.md)) | `harness/vertraege/`, `spec/spezifikation.md`, Source Precedence | [`grundlagen-referenz-richtung.md` §Spec-Straten](../.harness/baseline/v6.18.1/regelwerk/grundlagen-referenz-richtung.md#spec-straten-mehr-als-ein-spec-dokument) |

### Aufgelöste Adaptionen



| MR | aufgelöst durch |
|---|---|
| — | — |

## Zusatzklassen-Deklaration für Sensors-Bindung



| Klasse | Form | Bedeutung | Beispiel |
|---|---|---|---|
| — | — | keine Zusatzklassen deklariert | — |



## Modus-Deklaration pro Sub-Area

Die **Kürzel**-Spalte tragen nur Repos, deren Kennungen ein Bereichssegment
führen (`ADR-<KUERZEL>-NNNN`, `slice-<KUERZEL>-NNN`); wer ohne Segment zählt,
streicht sie. Regel: Baseline-Regelwerk `grundlagen-harness-dateien.md`
§Konventionsspeicher.



| Sub-Area (Pfad / Modul) | Kürzel | Modus | Begründung | Graduation-Bedingung / Folge-Slice |
|---|---|---|---|---|
| `*` (Default für gesamtes Repo) | `REPO` | Greenfield | Neues Repo ohne Bestandscode; die Spec führt, der Code folgt. | n/a (GF) |

## Glossar (optional)



| Begriff | Bedeutung |
|---|---|
| Baseline | Unveränderlicher Snapshot der Art `baseline`, aus Recordings importiert; Ausgangspunkt jedes Workspace |
| Match-Key | Aus der redigierten Repräsentation eines Requests gebildeter Schlüssel, über den Replay einen aufgezeichneten Austausch findet |
| Replay-Cursor | Pro Workspace, Replay-Session und Recording-Session geführte Position samt Menge verbrauchter Austausche |
| Spec-Straten | Die drei Spec-Dateien Lastenheft, Spezifikation, Architektur; Verweise gehen nur nach oben |
