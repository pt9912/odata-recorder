# Review-Report: slice-bootstrap-m6-transport (Teil C: `ADMIN_TLS_CERT`, `ADMIN_TLS_KEY`, `ADMIN_ALLOW_PLAINTEXT`, die vier Aufrufe von `SPEC-299`, Warnung, Umhüllung, Handbuch) — 2026-11-27

**Review-Art:** Code gegen Plan, Entscheidungen und Hard Rules. Kein Review gegen die DoD (Aufgabe des Verifiers). Genau ein Lauf, gesammelt.

**Gegenstand:** `git diff b0bfdcdc..84eca357` über die Pfade des Blocks (`src/bootstrap/`) und das Handbuch, Zweig `slice-bootstrap-m6-transport`; Code-Commit `cea33fb2`. Die Commits aus `main` im selben Bereich sind nicht Gegenstand, außer dem Nachzug `a147d726` als Eingang (Ort der Namen).

**Skill:** `.harness/skills/reviewer.md` @ Stand des Zweigs (`84eca357`) · **Modell:** Sonnet 5.5 · **Datum:** 2026-11-27

**Eingangs-Kontext:** Plan `slice-bootstrap-m6-transport` (§1, §3 mit Tabelle C01 bis C08, Mutationen und akzeptierten Negativen, §4 Block, §6 Raster, Zusatztabelle, Rückgaben); `SPEC-299`, `SPEC-300`, `SPEC-291`, `SPEC-110` (Nr. 1 bis 3, 7, 12), `SPEC-139` Nr. 2; `SPEC-200`, `SPEC-201`; `LH-FA-13`, `LH-RB-02`, `LH-QA-04`; [ADR-0001](../plan/adr/0001-implementierungssprache-go.md); `AGENTS.md` (Hard Rules). Die Angaben des Implementers zu Mutationen sind nicht übernommen und nicht wiederholt; die Zahlen unten sind eigene Läufe.

**Urteil: nicht freigegeben.** Kein HIGH, vier MEDIUM (F-1 bis F-4), ein LOW, zwei INFO. Am Produkt ist in diesem Lauf kein Fehler gefunden. F-1 bis F-3 sind Zusagen, die sich im Code brechen lassen, ohne dass das Paket `bootstrap` rot wird: 5 eigene Mutationen, 5 grün.

**Vorgehen.** Wegwerfkopie (`git archive HEAD` von `go.mod`, `go.sum`, `src`, `tests`), per `docker cp` in einen Container aus dem Image `odata-recorder:test` (`--network none --memory=4g --memory-swap=4g --cpus=4`, kein Mount), je ein Container unter `flock /tmp/odata-recorder-heavy.lock`. Basis: `go test -count=1 ./src/bootstrap/` grün (364 s). Je Mutation das ganze Paket, ohne Detektor, ein Lauf: M1, M2, M4, M5 grün (320 bis 359 s); M3 grün (337 s; ein erster Lauf endete unter Last 52 an der Zeitgrenze von 400 s und zählt nicht). Dazu eigene Probe-Tests am unveränderten Stand. Nicht gefahren: `make gates`, `make katalog-check`, `make handbuch-check`, `make test-race`, `make gegenproben`, `tests/e2e/`. Der Worktree blieb bis auf diesen Report unverändert.

## Findings

Befunde ohne Lösungsvorschlag nach dem Schema des Skills.

| ID | Kategorie | Befund | Quelle | Pfad | Verifizierbar | Klasse |
|---|---|---|---|---|---|---|
| F-1 | MEDIUM | Die Unabhängigkeit der zwei Listener hat keinen Fänger. M1 (der OData-Listener erhält das Admin-Paar, wenn `LISTEN_TLS_*` fehlt) und M2 (der Admin-Listener erhält das Paar von `LISTEN_TLS_*`, auch an Stelle seines eigenen und auf Loopback ohne Paar) bleiben im ganzen Paket grün. Versagen: wer allein `ADMIN_TLS_*` setzt, verliert HTTP am OData-Listener (M1), oder die Admin-Schnittstelle zeigt das Zertifikat des Dienstes (M2); die Probe zeigt beides, am unveränderten Stand sind die drei Zellen {nur Admin, nur `LISTEN`, beide} richtig. `SPEC-299` sagt „kein Wert des einen wirkt auf den anderen“, das Handbuch „wirken unabhängig voneinander“; die Tabelle in §3 führt keine Zeile dazu. | `SPEC-299` (Einleitung), `AGENTS.md` §3.15 | `src/bootstrap/lebenszyklus.go` · `al = h.transport.umhuelle(al)` | ja: Mutationen M1, M2 (`make test` bleibt grün) | Unabhängigkeit zweier Einstellungen ohne Fall, der beide Seiten beobachtet |
| F-2 | MEDIUM | „Es gilt, was `Build` las“ hat keinen Fänger: kein Test ändert die Dateien zwischen Laden und `Build`. M3 (`Build` ruft die Prüfung, nimmt bei gleichen Werten aber den Transport des Ladens) bleibt im ganzen Paket grün. Versagen: nach einem Wechsel des Paars zwischen beiden Lesungen bediente der Listener das alte Zertifikat. C05 löscht die Dateien erst nach `Build` und führt kein Laden; C01 prüft allein die Eingabe. Der unveränderte Stand bedient in der Probe das neue Paar. | `SPEC-299` Nr. 8 b, §6 Raster Zeile *Prüfung in `Build`*, Spalte *Wiederholung*; `AGENTS.md` §3.15, §3.16 | `src/bootstrap/konfig_admin.go` · `transport, vs := pruefe(adminEingabeAus(k, true), jetzt)` | ja: Mutation M3 | Zelle des Rasters ohne Fänger (zweites Lesen mit geänderter Datei) |
| F-3 | MEDIUM | Verstöße von `ADMIN_*` zusammen mit einem Verstoß, den das Laden früher meldet, haben keinen Fall: C03 mischt allein mit `LISTEN_TLS_*`, das danach gelesen wird. M4 (`formGueltig` wird falsch, sobald irgendein früherer Schlüssel einen Verstoß trägt, der Zusammenhang entfällt) und M5 (die Prüfung am Laden läuft nur bei gültigem `MODE`) bleiben im ganzen Paket grün. Versagen: bei Tippfehler in `MODE` und `ADMIN_LISTEN=0.0.0.0:8081` meldet der Start nur `MODE`, der zweite Verstoß erscheint erst im nächsten Lauf. Der unveränderte Stand meldet in der Probe beide Zeilen (auch `LISTEN=x` und `ADMIN_TLS_CERT` neben `MODE`). | `SPEC-299` Nr. 6 und 8 a, `SPEC-110` Nr. 8 und 12; `AGENTS.md` §3.16; Skill MEDIUM „Paar von Ursachen ohne Fänger“ | `src/bootstrap/konfig_admin.go` · `formGueltig = len(p.verstoesse) == vorher` | ja: Mutationen M4, M5 | Paar von Ursachen ohne Fänger (früher gemeldeter Verstoß, ungültiger `MODE`) |
| F-4 | MEDIUM | Das akzeptierte Negativ (2) nennt Folge, aber weder Grund noch Maß. Die genannte Absicherung (Quelltest von `slice-admin-tls`) hält allein `admintls.go`; die in diesem Slice geänderten `konfig_admin.go`, `wurzel.go`, `lebenszyklus.go` und `start.go`, die die Pfade und das Flag tragen, deckt sie nicht. Versagen: ein `println` eines Pfades des Paars dort bliebe grün. | `AGENTS.md` §3.15; Skill MEDIUM „Akzeptiertes Negativ ohne Grund und Folge“ | Plan §3 · `sieht ein Schreiben auf die Deskriptoren 1 und 2` | nein | Akzeptiertes Negativ ohne Grund und Maß |
| F-5 | LOW | Der Satz „ein Fehler im Paar lässt den Start scheitern, nie auf HTTP zurückfallen“ ist grammatisch gebrochen: der zweite Teil hat weder Subjekt noch Verb, das zu ihm passt. Er trägt die Zusage, dass es keinen Rückfall gibt. | `SPEC-200` (Handbuch), `docs/user/benutzerhandbuch-standard.md` §4 | `docs/user/benutzerhandbuch.md` · `nie auf HTTP zurückfallen` | nein | Satz des Handbuchs grammatisch gebrochen |
| F-6 | INFO | Die offenen Punkte zu C03, C04 und C06 in §6 (Ausgang *offen*, Adresse Planner) treffen zu: `TestAdminBuild` führt allein Fehlerfälle ohne `Handle`, den Transport hält `TestAdminWegJeModus`; `TestAdminWarnung` steht in `admintls_test.go`; die reine Zelle des Flags hängt nicht am Aufruf. Über §6 hinaus nennen die Zeilen C03 und C04 je 12 Zellen, die Tests führen 52 und 24 plus 8 (Obermengen). | `AGENTS.md` §3.9, §3.13 | Plan §6 · `Zeilen C03, C04 und C06 sagen etwas anderes` | nein | Plan-Tabelle hinter den Fängern (Adresse eingetragen) |
| F-7 | INFO | `TestWerteErreichenVerbraucher` prüft die drei neuen Felder allein am zweiten Lauf (`replay`): `transport` wird je Lauf überschrieben, während andere Felder beide Läufe vergleichen. Je Modus fängt `TestAdminEingabe`, deshalb kein Befund höherer Stufe. | Maintainability | `src/bootstrap/wurzel_test.go` · `"AdminTLSCert":         transport.CertPfad == k.AdminTLSCert,` | nein | Fixture prüft nur einen von zwei Läufen |

## Geprüft, ohne eigenes Finding

- **Kein Listener außerhalb Loopback ohne TLS oder Flag:** das Laden prüft vor `DATA_DIR`; `Build` prüft bei nichtleerem `AdminListen` erneut mit `FormGueltig` wahr. `pruefeAdminTransport` liefert dort entweder einen Weg oder einen Verstoß, der Nullwert `wegKeiner` erreicht `umhuelle` nur ohne Admin-Listener.
- **Ungültiges Paar, Pfadformen:** am `Build` Verweis ohne Ziel, Verzeichnis, Verweis auf ein Verzeichnis, Pfad mit Schrägstrich am Ende, Leerraum im Literal und fehlende relative Datei je `ADMIN_TLS_CERT: Datei nicht lesbar` ohne Pfad in der Meldung; ein Verweis auf die Datei gilt.
- **Mindestversion und Handshake am Admin-Listener:** TLS 1.0 und 1.1 abgewiesen, 1.2 und 1.3 verhandelt, allein `h2` abgewiesen, `h2` mit `http/1.1` ergibt `http/1.1`; deckt die Sätze des Handbuchs.
- **Modus und leerer `AdminListen`:** C07 (128 Zellen) und die Zellen je Modus stehen als Schleifen; ohne `AdminListen` kein Aufruf und kein Transport.
- **Vereinigung mit `LISTEN_TLS_*`:** eine Meldung, nach Schlüssel sortiert; ein Fehler von `ladeListenTLS` ist immer ein `*KonfigFehler` und geht nicht verloren.
- **Geheimnis:** die Zeilen tragen Schlüssel und festen Grund; keine neue Ausgabe außer `warne`; Schlüsselmaterial steht nicht in `Konfiguration` (allein Pfade und Flag).
- **Zusammengelegte und geänderte Bestandstests:** `TestAdminPraefix`, `TestKonfigNichtGeliefert`, `TestWerteErreichenVerbraucher`, `TestAdminLoopbackOhnePaar` (nur umbenannt) sind nicht gelockert; die drei Schlüssel wechseln von *abgelehnt* zu *gelesen* mit eigener Positivprobe. Die verlorene oder nie geschriebene Zusage steht in F-2.
- **Randformen im Code ohne Ort:** keine gefunden.
- **Handbuch gegen Code und Tests:** Tabelle `Einstellungen`, die vier berichtigten Stellen, Gründe, Warnung und Verantwortung decken sich; keine weitere Stelle mit „allein Loopback“ oder „allein HTTP“ gefunden. Außer F-5.

## Negativbefunde

| Bereich | Ergebnis |
|---|---|
| `src/bootstrap/` (Produktcode) | geprüft, ohne Befund am Verhalten |
| `src/bootstrap/` (Tests) | geprüft; F-1 bis F-3, F-7 |
| `docs/user/` | geprüft; F-5 |
| `docs/plan/planning/` (Plan, Roadmap, Welle) | geprüft; F-4, F-6 |
| `spec/`, `harness/`, `tools/`, `Makefile`, `Dockerfile` | durch den Slice nicht geändert (außer dem Nachzug aus `main`); Suche nach den drei Schlüsseln ohne veraltete Aussage |
| Suppression (§3.2), Architektur-Sicht (§3.4), ADR (§3.5, §3.6) | geprüft, ohne Befund |

## Summary

| Kategorie | Anzahl |
|---|---|
| HIGH | 0 |
| MEDIUM | 4 |
| LOW | 1 |
| INFO | 2 |

**Finding-Klassen dieses Laufs:** Unabhängigkeit zweier Einstellungen ohne Fall, der beide Seiten beobachtet · Zelle des Rasters ohne Fänger (zweites Lesen mit geänderter Datei) · Paar von Ursachen ohne Fänger · Akzeptiertes Negativ ohne Grund und Maß · Satz des Handbuchs grammatisch gebrochen · Plan-Tabelle hinter den Fängern · Fixture prüft nur einen von zwei Läufen

## Verdikt

**Merge-blockierend:** ja (F-1 bis F-4). Der Produktcode trägt in den geprüften Punkten die Zusagen; offen sind Fänger und ein Negativ.

**Übergabe:** F-1 bis F-5 an den Implementer, F-6 an den Planner (schon in §6 eingetragen); die Finding-Klassen gehen in die Slice-Closure §7. Die Prüfung der DoD und der Gates bleibt beim Verifier.
