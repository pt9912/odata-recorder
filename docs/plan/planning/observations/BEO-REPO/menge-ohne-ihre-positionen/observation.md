# Der Fänger führt die Menge, aber nicht die Positionen darin, und eine Mutation an einer Position bleibt grün

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Der Plan nennt eine Menge (die Entities eines Ergebnisses, die Formen von `$select`), der Fänger führt ihre Elemente nur an einer Position (die erste Entity, `*` allein statt `*` neben einem Namen). Eine Mutation an der anderen Position bleibt grün. Nahe an `fix-deckt-die-genannte-mutation-nicht-die-klasse` und `ziffern-klasse-ohne-schleife`; hier fehlt die zweite Achse (Position) einer sonst vollständigen Menge schon im Plan.
