# Die Roadmap bricht im Slice-Zweig, sobald der Start-Move den Slice verschiebt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Im Parallelbetrieb liegt der Start-Move nach `in-progress/` im Slice-Zweig, die Roadmap schreibt aber nur der Integrationsschritt (Übergangsregeln Nr. 3 und 4). Die Roadmap führt dann Links auf den alten Ort und den Ruhe-Marker *Nichts in Arbeit.*; `make docs-check` meldet im Zweig Befunde (`target-missing`, `planning-drift`), die der Slice nicht verursacht, und `make gates` ist im Zweig rot, bis jemand die Roadmap im Zweig nachzieht.
**Stand:** offen
