# Ein Test eines Pakets, das der Slice nicht berührt, wird unter Last rot und im nächsten Lauf grün

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Test misst eine Größe, die an der Last des Hosts hängt (Zuteilungen, Wettlauf zweier Goroutinen), und fällt in einem Lauf von `make test` oder `make gates` und im nächsten nicht. Die Rolle, die den Lauf fährt, berührt das Paket nicht und kann den Fall weder beheben noch einem Slice zuordnen. Verwandt mit `rate-ohne-detektor-schwankt-mit-der-last` (dort die Rate eines Fängers, hier ein Test des Bestands).
