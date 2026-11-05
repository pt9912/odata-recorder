# SPEC-039 — Architektur-Gate (`make a-check`)

**Werkzeug:** `make a-check`.

**Festlegung:** Hexagonale Importregeln gemäß `.a-check.yml`; keine Adapter-Imports im Core, keine lateralen Adapter-Imports.

1. **Wurzel.** Der Composition Root ist `src/bootstrap/**` (`composition_root`); nur er ist von den Richtungsregeln und von `tech` ausgenommen.
2. **Prüfpakete.** `tests/**` ist die Schicht `tests` ohne Rolle mit Kanten auf `model`, `ports-inbound` und `adapters-driven`; den Composition Root nennt sie ohne Kante.
   Ein Import von Services, Outbound-Ports oder Driving Adaptern aus `tests/` ist ein `wrong-direction`, ebenso ein Import von `tests/` aus `src/`.
   `tech` gilt für `tests/` wie überall: `modernc.org/sqlite` und `encoding/xml` außerhalb ihres Adapters sind ein `tech-leak`.
   Die Kante auf `adapters-driven` gilt für alle Driven Adapter, nicht für einen namentlich; die Bibliotheksgrenze hält `tech`.
   Eine weitere Kante der Schicht entscheidet eine neue ADR.

**Stand:** vorhanden.

**Gegenprobe:** Die Fälle 171 bis 173 von `SPEC-102`: Bibliothek in `tests/` (`tech-leak`), Import eines Dienstes aus `tests/` (`wrong-direction`), der erlaubte Weg über `bootstrap`, Inbound-Port und Driven Adapter (Exit 0).
Offen bleibt der Import von `tests/` aus `src/`: er ist gemessen ein `wrong-direction`, hat aber keinen eigenen Fall, weil keine Zeile der Datei ihn trägt, die jemand entfernen könnte; er entstünde nur durch eine neue Kante.

**Grenze:** Prüft Importrichtungen, nicht Verhalten.
