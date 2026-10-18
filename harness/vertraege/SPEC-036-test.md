# SPEC-036 — Build und Unit-Tests (`make build`, `make test`)

**Werkzeug:** `make build` und `make test`, beide über das `Dockerfile`; Eingangs-Image per Digest gepinnt, alle Stufen außer `deps` ohne Netz.

**Festlegung `make build`:** Baut das Binary aus `src/bootstrap/main` mit `-trimpath` und ohne VCS-Stempel; es liegt im Image `odata-recorder:build` unter `/out/odata-recorder`.
Die Version kommt aus der Make-Variablen `VERSION` (Vorgabe `dev`).
Das Ziel schlägt fehl, wenn `odata-recorder version` nicht `odata-recorder <VERSION>` ausgibt oder die Build-Information des Binary nicht `-trimpath=true` trägt.

**Festlegung `make test`:** Schlägt fehl, wenn `gofmt -l` eine Datei nennt, `go vet` einen Befund meldet oder ein Unit-Test rot ist.
Der Fachumfang (Domain, Matching, Query-Interpreter, CRUD) wächst mit dem Code.

**Randformen von `bootstrap.Run`:**
1. Keine Argumente oder genau `version`: Ausgabe `odata-recorder <Version>` nach stdout, Exit-Code 0.
2. Jedes andere Aufrufmuster — unbekanntes Kommando, leeres Argument, Flag wie `--help`, `version` mit Zusatzargument: `Unbekanntes Kommando: <erstes Argument>` nach stderr, Exit-Code 2.
3. Ein nil-Writer verwirft die Ausgabe; der Exit-Code bleibt unverändert.

**Stand:** vorhanden für das Skelett aus Composition Root und Version.

**Grenze:** Prüft kein Zusammenspiel mit Upstream oder Datenbank; das tun `make test-integration` und die Verträge `SPEC-037` und `SPEC-042`.
