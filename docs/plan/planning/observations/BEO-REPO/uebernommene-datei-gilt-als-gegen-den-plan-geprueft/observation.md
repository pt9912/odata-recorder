# Eine Datei, die fertig und grün auf einem Zweig liegt, gilt als gegen ihren Plan geprüft

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Slice übernimmt eine Datei, die ein anderer Vorgang geschrieben, gemessen und grün hinterlassen hat (Schnitt vor oder nach dem Code). Der Plan führt sie als Übernahme mit wenigen benannten Änderungen; Review und Verifikation hat sie nicht durchlaufen. Beim Aufnehmen fehlen Zählprüfungen, Fixtures und Fänger, die der Plan verlangt, und die Messung des Gebers trägt die Lücken mit. Nahe an `fixture-gleicht-dem-wert-der-mutation` und `vertrag-sagt-mehr-zu-als-ein-gate-prueft` (dort die einzelne Lücke, hier die Annahme, die sie durchreicht).
