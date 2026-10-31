# Eine Zusage über den Inhalt einer Fehlermeldung hat keinen Fänger über jede Stelle, die sie erzeugt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Vertrag sagt zum Text einer Fehlermeldung, was er nennt (den Ort) und was nie (den Wert). Die Tests prüfen den Code und das Vorhandensein eines Fehlers, nicht den Text an jeder Stelle, die ihn erzeugt; eine Stelle, die den Wert in die Meldung schreibt, bleibt grün und ist ein Leck in Logs und Antworten. Nahe an `vertrag-sagt-mehr-zu-als-ein-gate-prueft`; hier eigens geführt, weil der Fänger eine Schleife über die erzeugenden Stellen aus dem Quelltext verlangt.
