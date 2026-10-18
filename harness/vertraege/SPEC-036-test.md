# SPEC-036 — Build und Unit-Tests (`make build`, `make test`)

**Werkzeug:** `make build` und `make test`, beide über das `Dockerfile`; Eingangs-Image per Digest gepinnt, alle Stufen außer `deps` ohne Netz.

**Festlegung:** `make build` baut das Binary aus `src/bootstrap/main` für die Zielplattform, reproduzierbar (`-trimpath`, ohne VCS-Stempel), mit der Version aus dem Build-Argument `VERSION`.
`make test` schlägt fehl, wenn `gofmt -l` eine Datei nennt, `go vet` einen Befund meldet oder ein Unit-Test rot ist; der Fachumfang (Domain, Matching, Query-Interpreter, CRUD) wächst mit dem Code.

**Stand:** vorhanden für das Skelett aus Composition Root und Version.

**Grenze:** Prüft kein Zusammenspiel mit Upstream oder Datenbank; das tun `make test-integration` und die Verträge `SPEC-037` und `SPEC-042`.
