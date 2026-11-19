# Der Entscheidungssatz einer `Accepted` ADR und die Einzelregel, an die sie delegiert, lesen sich verschieden

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die ADR sagt allgemein („ein Lauf außerhalb der Last ist ungültig und wird wiederholt“) und delegiert die Einzelheiten an die Spezifikation; die Spezifikation setzt eine Ausnahme (ein grüner ungültiger Lauf wird nicht wiederholt). Die ADR ist nach `Accepted` unveränderlich (`AGENTS.md` §3.5), die Spezifikation fortschreibbar; ob die Ausnahme eine Schärfung oder ein Widerspruch ist, entscheidet kein Artefakt, bis jemand es tut, und der Plan führt die Frage ohne Adresse, bis die Verifikation sie findet. Nahe an `spezifikation-widerspricht-sich-innerhalb-einer-nummer` (dort zwei Sätze einer Nummer, hier ADR und Einzelregel).
