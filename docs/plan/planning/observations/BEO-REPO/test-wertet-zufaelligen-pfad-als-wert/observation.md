# Ein Test wertet ein Stück eines zufälligen Temp-Pfads als verbotenen Wert und wird in einem von wenigen Läufen rot

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Test prüft, dass ein Fehlertext keinen Wert nennt, und führt als Marker eine kurze Ziffernfolge (`4242`, `12345`). Der Fehlertext nennt den Pfad der Datei, der Pfad stammt aus `t.TempDir()` und trägt eine Zufallszahl; enthält sie die Folge, ist der Test rot, ohne dass der Code sich geändert hat, und im nächsten Lauf grün. Unabhängig von der Last des Hosts, deshalb nicht `test-fremden-pakets-flackert-unter-last`: dort schwankt eine Messung, hier ein Eingabewert des Aufbaus.
