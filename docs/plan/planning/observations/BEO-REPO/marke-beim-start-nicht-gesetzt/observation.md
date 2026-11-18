# Die Größenmarke ist beim Start und beim ersten Code-Commit nicht gesetzt, und erst die Closure bemerkt es

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Planner setzt `basis`, `schaetzung` und `marke` eines Plans vor dem ersten Code-Commit (`AGENTS.md` §3.14, `SPEC-128` Nr. 1); beim Start stehen nur der Teil des Architect und eine Nachrechnung in Prosa. `make groessenmarke` meldet einen Plan mit ausstehender Marke als Hinweis mit Exit 0 (`SPEC-128` Nr. 3 p), und der Hinweis steht in jedem Gate-Lauf des Slice, ohne dass er rot wird. Die Bedingung der Nachrechnung (bei Schätzung am Rand der Schnittgrenze vor dem Code schneiden) ist dann nie geprüft, und die Marke entsteht erst bei der Closure. Nahe an `groessenmarke-ueberschritten-ohne-rueckfuehrung` (dort ist die Marke gesetzt und wird überschritten) und `schaetzformel-ueberschaetzt-den-slice` (dort ist die Schätzung zu hoch); hier fehlt der Schritt des Planners vor dem Code.
