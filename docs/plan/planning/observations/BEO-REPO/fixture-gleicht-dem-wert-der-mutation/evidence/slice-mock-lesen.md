**Vorgang:** slice-mock-lesen
**Fund:** Der Fall "ohne `#`" in `TestPayloadVerstoesse` nutzte `N.Special`: ohne die Prüfung schneidet `name[1:]` das `N` ab und der Typ ist unbekannt, die Fixture gleicht dem Wert der Mutation (V-4); ebenso Mutation G8 in V-3. Beleg: Report `2026-11-07-slice-mock-lesen-verifikation` (V-3, V-4).
