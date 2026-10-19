**Vorgang:** slice-sqlite-adapter-migration
**Fund:** Der Fund: `TestGleichzeitigesOeffnenMigrierteinmal` scheiterte nichtdeterministisch mit `database is locked`, in zwei von 120 Läufen zu je 40 Wiederholungen, und erst die Wiederholung deckte es auf (Review-Befund H2 und Verifikations-Befund V-H1 des Slice); behoben durch Wiederholung bei `SQLITE_BUSY`, belegt mit 2000 grünen Wiederholungen.
