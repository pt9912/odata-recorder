# Eine Aussage über einen Exit-Wert nennt nicht, an welcher Stelle er gemessen ist

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Vertrag über ein Ziel, das einen Treiber ruft, sagt „Exit 2“ oder „der Exit des Treibers ist der Exit von `make`“. Am Treiber und am `make`-Aufruf sind das zwei Werte: `make` meldet jedes fehlschlagende Rezept als Exit 2 und nach einem Signal 128 plus Signalnummer, der Treiber kennt 0, 1 und 2. Ein Fänger, der am `make` misst, bestätigt die Aussage für den Treiber trivial, und die Mutation, die nur am Treiber einen anderen Exit liefert, bleibt ungefangen.
**Stand:** offen
