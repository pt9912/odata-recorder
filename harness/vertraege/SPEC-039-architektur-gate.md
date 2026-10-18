# SPEC-039 — Architektur-Gate (`make a-check`)

**Werkzeug:** `make a-check`.

**Festlegung:** Hexagonale Importregeln gemäß `.a-check.yml`; keine Adapter-Imports im Core, keine lateralen Adapter-Imports.

**Stand:** vorhanden.

**Grenze:** Prüft Importrichtungen, nicht Verhalten; heute einen leeren Baum, bis Code unter `src/` liegt.
