# Die Gegenproben erzeugen die Last selbst, die ihre eigene Frist reißt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

`make gegenproben` startet die neun `lint/*`-Fälle zugleich, und jeder Fall startet `golangci-lint` auf allen Kernen des Hosts. Die Last steigt innerhalb von 30 Sekunden von der Ruhe auf über 100 und endet bei über 270; die Fälle reißen ihre Frist von 120 s, und der Lauf ist rot, auch wenn auf dem Host sonst nichts läuft. Die Messvorschrift in `SPEC-102` *Zeitbudget* Nr. 3 nennt einen Lauf mit Last am Ende über der Kernzahl ungültig und Nr. 5 e verlangt, ihn zu wiederholen; die Wiederholung ergibt dasselbe, weil die Last aus dem Lauf stammt und nicht aus fremden Prozessen. Verwandt mit `laufzeitmessung-widerspricht-sich-unter-last` (dort widersprechen sich Messungen verschiedener Läufer, hier ist die Messung selbst der Verursacher).
