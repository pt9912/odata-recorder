# Schreiber und Leser einer Darstellung sind nicht als Paar über die Klasse der Texte geprüft

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Schreiber und der Leser derselben Darstellung werden je für sich gegen die Spezifikation geprüft; die Zusage, dass der Leser zurückliest, was der Schreiber ausgibt (`SPEC-169` Nr. 8, `SPEC-172` Nr. 7), gilt für die Klasse der Werte, die beide meinen. Ein Wert, den der Leser anders deutet als der Schreiber ihn meint (der Text `[REDACTED]` einer Art `string`, den `ODataCodec` als Marker liest), fällt durch, solange der Rundlauf nur Einzelfälle und nicht die Klasse der Texte führt.
