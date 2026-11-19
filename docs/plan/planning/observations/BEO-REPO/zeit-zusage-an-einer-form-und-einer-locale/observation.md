# Eine Zusage über Zeit oder Größe ist an einer Form und einer Locale gemessen und gilt für alle

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Vertrag sagt eine Obergrenze zu („ein Befehl von 32768 Zeichen braucht höchstens 3 s“); der Fänger misst eine Form der Eingabe (eine Umleitung hinter einem Wort, gefüllt auf die Länge) in der Locale des Prüflaufs (`LC_ALL=C`). Andere Formen derselben Länge (Zitat und Operator im Wechsel, viele Heredoc-Operatoren, JSON in Anführungszeichen mit Escapes) brauchen das Vierfache bis Zwanzigfache, und in einer UTF-8-Locale kostet jede Zeichenoperation ein Vielfaches; kein Lauf sieht es, weil die Messung die Form wählt, die der Code schnell liest. Nahe an `laufzeitmessung-widerspricht-sich-unter-last` (dort widersprechen sich Messungen derselben Form, hier fehlt die Form) und an `unbeschraenkte-menge-ohne-benannte-pruefgrenze` (dort fehlt die Prüfgrenze einer Menge, hier die Menge der Formen einer Zeit-Zusage).
