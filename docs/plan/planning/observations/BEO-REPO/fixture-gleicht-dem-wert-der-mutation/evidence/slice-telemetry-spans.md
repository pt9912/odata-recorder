**Vorgang:** slice-telemetry-spans
**Fund:** Die erste Fassung der Mutation `s11-offene-spans-gehalten` hielt den Span in einer Goroutine, deren Parameter nach dem Compiler tot ist; sie war grün als gleichwertige Mutation und fing nichts. Ersetzt durch einen Zeitgeber, der den Span hält (rot in `TestSpanOhneEnde`). Beleg: Plan §3 (Kopf der Tabelle *Zusage · Mutation · roter Test*).
