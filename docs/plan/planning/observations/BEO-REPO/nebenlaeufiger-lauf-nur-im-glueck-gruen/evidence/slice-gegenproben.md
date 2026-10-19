**Vorgang:** slice-gegenproben
**Fund:** Der Fund: Die Lint-Fälle von `make gegenproben` brachen nebenläufig mit `parallel golangci-lint is running` ab, in 5 von 8 erzwungenen Läufen (Verifikations-Befund V-1, `docs/reviews/2026-10-19-slice-gegenproben-verifikation.md`); behoben durch ein eigenes `TMPDIR` je Fall.
