# Der Transaktionsvertrag des SQLite-Adapters war in einem Gate-Lauf unter Last einmal rot

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Lauf von `make gates` meldete `TestTransaktionVertrag` im Paket `sqlite` rot, die Wiederholung nicht. Ob der Test von der Last oder von einem Zustand des Adapters abhängt, ist offen. Die Beobachtung gehört dem Adapter und nicht dem Slice, in dem sie auftrat.
