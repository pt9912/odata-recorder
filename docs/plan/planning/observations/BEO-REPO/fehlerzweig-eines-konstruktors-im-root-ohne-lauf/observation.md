# Der Fehlerzweig eines Konstruktors im Composition Root läuft in keinem Test

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

`verdrahte`, `verdrahteRecord` und `verdrahteReplay` rufen die Konstruktoren der Bauteile (`blobs.New`, `NeuerWaechter`, `audit.New`, `NeueWartung`, `NeueSession`, `NeuePipeline`, `NeuerUpstream`, `NeuerRecorder`, `NeueTelemetrie`, `NeuerReplay`) und geben bei einem Fehler `err` zurück. Kein Test löst einen dieser Fehler aus, weil die Wurzel die Argumente selbst baut (zum Beispiel ist `audit.New` von der Wurzel aus nicht auslösbar) und ein Auslöser einen Haken in `Bauteile` oder `export_test.go` verlangte. Ein Konstruktor, der den Fehler verschluckt oder die Ablage offen lässt, fiele an keinem Test. Nicht dasselbe wie `zweig-ohne-lauf-unter-dem-detektor`: dort sieht der Detektor nur, was ein nebenläufiger Test ausführt; hier gibt es keinen Lauf, weil der Fehler nicht erzeugbar ist, und es geht nicht um Nebenläufigkeit.
