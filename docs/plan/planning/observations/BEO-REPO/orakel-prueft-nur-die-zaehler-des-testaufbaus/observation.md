# Das Orakel eines Wirkungs-Fängers prüft nur die Zähler, die der Testaufbau verändert

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Wirkungs-Fänger zählt, was ein Fenster der Sperre zuließe, und sein Orakel vergleicht nur die Zähler, die der Aufbau des Tests ändert. Ein Fenster zwischen den übrigen Zählern bleibt offen, obwohl der Code sie im Betrieb zählt; die Begründung „kein Test zählt sie“ ist eine Eigenschaft des Tests, nicht des Codes.
