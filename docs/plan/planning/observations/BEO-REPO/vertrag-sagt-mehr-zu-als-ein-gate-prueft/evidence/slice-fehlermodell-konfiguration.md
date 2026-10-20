**Vorgang:** slice-fehlermodell-konfiguration
**Fund:** Die Zusage „ein Feld falschen Typs meldet nur den Typ, bei `upstream_url` in jedem Modus“ hatte nur im Modus `record` einen Fänger (Re-Review N-1): die Mutation „Typfehler nur in `record` melden“ blieb grün. Behoben durch Fälle für `replay` und `mock` in `TestKonfigDienstefeldTyp`.
