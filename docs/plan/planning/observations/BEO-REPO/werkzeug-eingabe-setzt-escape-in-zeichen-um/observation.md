# Eine Werkzeug-Eingabe setzt \u-Escape-Folgen in Zeichen um

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Agent schreibt eine Datei oder einen Bericht über Heredoc oder Werkzeug-Eingabe; eine Folge aus Backslash, `u` und vier Hex-Ziffern im Eingabetext kommt als das Zeichen an. Ein Fixture, das den Escape-Text prüfen soll (Rohstring), prüft dann das Zeichen oder umgekehrt, und der Fänger trennt die Mutation nicht, ohne dass Aufbau oder Gate es melden; aufgefallen ist es nur beim Gegenlesen der geschriebenen Datei.
