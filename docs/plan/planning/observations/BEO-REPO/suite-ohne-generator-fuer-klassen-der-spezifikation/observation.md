# Eine Klasse der Spezifikation hat in der Suite keinen Generator, und erst ein Probe-Paket des Verifiers findet ihre Lücken

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Vertrag über ein fremdes Format nennt Mengen (jede Referenzart, jede Position, jeder Rest hinter der Klammer); die Suite führt sie als Liste von Fällen, die dem Autor einfielen. Ein Probe-Paket aus Generatoren über dieselben Klassen, einmal gegen den Code und einmal gegen die Mutationen gefahren, findet die Mitglieder, die zehn Mutationen von Hand nicht fanden; die Suite selbst führt keinen Generator.
