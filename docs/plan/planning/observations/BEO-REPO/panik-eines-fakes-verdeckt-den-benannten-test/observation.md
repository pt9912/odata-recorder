# Ein Fänger wird rot, aber nicht in dem Test, den die Tabelle nennt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Eine Mutation macht einen Fake paniken oder eine Sperre hängen; das Testprogramm endet mit `panic:` oder am Zeitlimit, ohne `--- FAIL` und ohne den Namen des Tests, den die Tabelle *Zusage · Fänger* nennt. Rot kommt aus dem richtigen Grund, aber nicht aus dem benannten Test, und die übrigen Tests laufen nicht mehr.
