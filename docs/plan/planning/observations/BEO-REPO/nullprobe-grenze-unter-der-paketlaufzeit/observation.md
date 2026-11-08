# Die Nullprobe des Mutationslaufs hat eine feste Zeitgrenze, und die Laufzeit des Pakets wächst mit dem Produkt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Nullprobe von `tools/mutationslauf` fährt das Paket einer Liste mit `-test.timeout=2m`. Ein Paket, das mit den Slices wächst (`src/bootstrap`: rund 150 s ohne Detektor), reißt die Grenze, ohne dass sich am Werkzeug etwas geändert hat; die Liste endet mit `Ausgangsstand rot`, und die Rolle fährt ihre Mutationen mit einem eigenen Runner. Nahe an `gegenprobe-haengt-an-der-laufzeit-des-produkts`; hier ist die Grenze Teil des Werkzeugs, nicht eines Falls.
