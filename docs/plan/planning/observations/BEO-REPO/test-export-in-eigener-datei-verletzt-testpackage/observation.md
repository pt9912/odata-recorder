# Ein Plan setzt einen Test-Export mit `package admin` in eine eigene Datei, und das Lint-Profil lässt ihn nur in `export_test.go` zu

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Plan nennt in §3 für einen Test-Export, der interne Zustände des Pakets erreicht (ein Haltepunkt, eine Naht), eine eigene Datei `*_test.go` mit `package admin` (hier `tokendatei_halt_test.go`). Das Lint-Profil schaltet `testpackage` ein und nimmt allein `export_test.go` aus (`.golangci.yml`, `skip-regexp`); `make lint` meldet an der eigenen Datei `package should be admin_test instead of admin`, und `make gates` ist rot, bis der Plan umgebaut ist. Das Profil wird nicht gelockert, und `export_test.go` gehört mehreren Slices des Pakets; der Implementer hält an, und der Planner ändert den Block der Größenmarke und die Kollisionsprüfung nach dem Handoff.
