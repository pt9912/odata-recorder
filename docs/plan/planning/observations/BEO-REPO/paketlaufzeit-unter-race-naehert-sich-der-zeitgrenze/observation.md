# Ein erschöpfender Test nähert unter dem Detektor die Zeitgrenze je Paket an

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Fänger über eine unbeschränkte Klasse wird erschöpfend gebaut (Länge der Folgen als Prüfgrenze); unter `go test -race` kostet er das Vielfache der Zeit ohne Detektor, und die Zeitgrenze von zehn Minuten je Paket (`make test-race`) wird zur obersten Schranke der Menge. Die Reserve hängt an den CPUs des Containers und an der Last des Hosts, nicht am Code. Verwandt mit `nullprobe-grenze-unter-der-paketlaufzeit` (dort die Grenze des Mutationswerkzeugs, hier die des Gates `test-race`).
