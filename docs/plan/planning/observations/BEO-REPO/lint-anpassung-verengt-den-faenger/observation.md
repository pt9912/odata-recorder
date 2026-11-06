# Eine Anpassung an das Lint-Profil ändert das Verhalten des Tests und verengt seinen Fänger

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Das Lint-Profil verbietet Variablen des Pakets in Tests (`gochecknoglobals`); die Anpassung macht aus einem geteilten Wert eine Funktion, und jeder Aufruf erzeugt einen neuen. Der Produktcode bleibt unverändert, der Test fängt aber nicht mehr, was am geteilten Wert hing (Zustand im Wert eines Dienstes, den alle Anfragen teilen). Kein Gate sieht den Unterschied, weil beide Fassungen grün sind; erst eine Mutation am Wert findet ihn.
