# Eine Testhilfe, die ein Argument des Falls ersetzt, verdeckt, was das Original trägt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Eine Hilfe der Suite umhüllt oder ersetzt ein Argument, das der Fall dem Code übergibt (`src`, `ctx`, ein Fake), und trägt nur einen Teil dessen, was das Original trägt. Ein Code, der die übrigen Eigenschaften nutzt (eine optionale Schnittstelle, eine Methode des Fakes), sieht die Hilfe statt des Originals; die Zähler am Original bleiben 0, und eine Mutation, die die Zeile vorher fing, bleibt grün.
