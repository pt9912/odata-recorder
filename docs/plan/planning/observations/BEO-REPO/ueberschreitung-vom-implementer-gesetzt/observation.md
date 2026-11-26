# Der Implementer schreibt den Eintrag `ueberschreitung` in den Block des Plans, den kein Vertrag einem Schreiber zuweist

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

`SPEC-128` weist `basis`, `schaetzung` und `marke` dem Planner zu (Nr. 3 k); den Eintrag `ueberschreitung` weist weder Nr. 3 k noch Nr. 5 d einem Schreiber zu. Reißt die Messung die Marke, meldet `make groessenmarke` den Befund mit der Zeile, die ein Mensch einträgt, und der Implementer schreibt den Eintrag nach der Fix-Runde selbst; sein *N* ist die Messung dieses Moments, und die nächste Runde (Auflage) reißt es erneut (hier 3200 gegen 3277). Der Plan nennt den Eintrag „Eintrag des Implementers, der Planner prüft ihn“, das Gate bleibt bis zur Closure rot, und der Planner setzt *N* und Grund am Ende neu. Nahe an `marke-nach-dem-ersten-code-commit-ersetzt` (dort ersetzt der Planner die Marke nach dem Code) und `groessenmarke-ueberschritten-ohne-rueckfuehrung` (dort die Überschreitung selbst).
