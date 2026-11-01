**Vorgang:** slice-http-lebenszyklus
**Fund:** V-2 (Senken von `net/http` außer `log.Default()`, `stderr`, `stdout`, L19) und V-5 (Fristen von `Schliessen`: geprüft sind sieben Werte, die Menge ist unbeschränkt) tragen jetzt Prüfgrenze, Grund und Folge im Betrieb als akzeptiertes Negativ. Beleg: Plan §6 (V-5) und §7.
