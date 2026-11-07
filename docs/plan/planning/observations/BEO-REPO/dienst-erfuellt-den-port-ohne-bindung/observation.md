# Ein Kommentar sagt zu, der Dienst setze den Port um, und keine Zusicherung bindet ihn daran

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Kopf eines Dienstes sagt, er setze die Operationen eines Inbound Ports um (`AGENTS.md` §3.11), aber weder ein Test noch eine Zusicherung zur Übersetzungszeit bindet den Typ an die Schnittstelle. Ändert eine Mutation die Signatur (etwa ein zusätzlicher variadischer Parameter), erfüllt der Dienst den Port nicht mehr, und alle Tests des Pakets bleiben grün, weil sie den Dienst über seinen Typ rufen. Der Bruch fiele erst im Einbau des Aufrufers auf.
