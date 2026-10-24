# Ein Zweig läuft in keinem Test unter dem Detektor, und der Test sagt zu, der Detektor melde jedes Paar

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Test prüft unter `-race`, dass ein Dienst keinen Zustand hält, und führt nur die Zweige aus, die sein Autor nannte. Ein geteilter Zugriff in einem anderen Zweig bleibt ungemeldet (`make test-race` sieht nur, was läuft), und der Kopf des Tests sagt zu, der Detektor melde jedes ungeordnete Paar. Jede Prüfrunde findet den nächsten Zweig, solange die Menge der Zweige aus der Erinnerung gebildet wird und nicht aus dem Abdeckungsprofil der Detektor-Tests.
