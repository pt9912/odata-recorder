# Der Report einer Rolle liegt im Repo, und kein Gate läuft nach dem Schreiben

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Report des Reviewers oder Verifiers unter `docs/reviews/` fällt unter dieselben Gates wie der Plan (`make docs-check`). Die Rolle, die ihn schreibt, fährt nach dem Schreiben kein Gate; eine nackte `ADR-*`-Kennung ohne Link in einer Kopfzeile macht `make gates` rot, und erst die nächste Rolle findet es.
