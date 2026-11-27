# Ein Test lässt Zustand auf einem begrenzten Dateisystem liegen und ist im selben Container nicht wiederholbar

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Test legt sein Datenverzeichnis auf einem kleinen, festen Dateisystem an (`/voll`) und räumt es nicht. Ein Lauf je frischem Container ist grün; `-test.count=N` in einem Prozess scheitert ab dem zweiten oder späteren Lauf am Aufbau. Eine Rate je Lauf (`AGENTS.md` §3.10) lässt sich dann nur mit von Hand geleertem Dateisystem messen, und ein Rot aus dem Aufbau sieht wie ein Rot des Fängers aus.
