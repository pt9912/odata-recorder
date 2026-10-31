# Eine Mutation gilt als äquivalent, weil der Bestand keine Eingabe führt, die sie unterscheidet

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Bericht stuft einen Mutanten als äquivalent ein (`sort.Slice` statt `sort.SliceStable`), weil keine Eingabe des Bestands ihn trennt. Äquivalent heißt, dass keine Eingabe des Vertrags ihn unterscheidet; bei einer Ordnung ist das eine Eingabe, deren Erzeugungsfolge von der Ausgabefolge abweicht und die mehr Elemente hat, als der Sortierer ohne Umordnung stabil behandelt. Dieselbe Beobachtung unter dem Namen `ordnung-prueft-nur-die-schreibweise-der-fixture`: die Fixture trug die Reihenfolge der Ausgabe schon, nur ein Verzeichnis, damit das Register den Vorgang nicht zweimal zählt.
