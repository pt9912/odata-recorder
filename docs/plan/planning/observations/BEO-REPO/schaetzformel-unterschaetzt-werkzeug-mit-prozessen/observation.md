# Die Schätzformel der Größenmarke unterschätzt ein Werkzeug mit echten Prozessen und Fängern als Schleifen über Mengen

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Schätzung aus `SPEC-128` (Zeilen der Tabelle mal 85 oder 80, 200 je Fake, Produktcode) liegt unter der Messung, wenn der Slice echte Prozesse startet (Fixture-Modul, Ersatz-Skript für `docker`, Signal-Tests) und die Fix-Runde nach dem Review die Fänger als Schleifen über alle Elemente der Mengen schreibt. Die Marke (Schätzung mal 1,35) hält dann nicht, und der Eintrag `ueberschreitung` trägt die Differenz. Das Gegenbild zu `schaetzformel-ueberschaetzt-den-slice`: dort gemeinsame Fakes und kleine Mengen, hier Prozesse und große Mengen.
