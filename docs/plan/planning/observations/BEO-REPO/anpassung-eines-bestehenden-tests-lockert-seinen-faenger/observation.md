# Die Anpassung eines bestehenden Tests eines früheren Teils lockert seinen Fänger, und kein Gate sieht den Unterschied

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Slice ändert die Ausgabe, die ein Test eines früheren Slice liest (hier: im Modus `record` steht hinter der Zeile des Gesamtspans die des Capture-Spans), und passt den Test an. Die Anpassung ersetzt eine strengere Prüfung durch eine schwächere (die Zeile des Gesamtspans steht in jedem Fall zuletzt, ersetzt durch: es gibt genau eine solche Zeile, gleich dem Literal), auch in den Fällen, in denen die strengere weiter gälte. Der Test bleibt grün, die Mutation „eine Zeile hinter dem Gesamtspan“ ebenfalls; die Tests der Kinder (`mock`, `replay`, `record` mit 200) halten die Stelle nur für ihre Fälle. Nahe an `lint-anpassung-verengt-den-faenger` (dort verengt eine Anpassung an das Lint-Profil den Fänger, hier eine Anpassung an eine geänderte Ausgabe) und an `redundante-pruefung-ohne-wirkung`.
