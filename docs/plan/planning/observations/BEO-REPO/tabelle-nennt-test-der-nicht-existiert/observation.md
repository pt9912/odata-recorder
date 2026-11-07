# Die Spalte *Fänger* der Tabelle nennt einen Test, den es nicht gibt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Tabelle *Zusage · Fänger* eines Plans führt in der Spalte *Fänger* Namen von Tests, die der Plan vorsah und der Code nie oder unter anderem Namen bekam (24 Zeilen in `slice-mock-schreiben`, Review F-2). Ein Reviewer, der der Tabelle glaubt, sucht den Test nicht; ein Gate, das die Namen gegen den Quelltext hält, gibt es nicht. Nahe an `vertrag-sagt-mehr-zu-als-ein-gate-prueft`; hier ist die Zusage da, der Fänger nur ein Name.
