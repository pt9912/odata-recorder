# Der Schnitt eines Slice lässt zwischen Geber und Nehmer keinen grünen Stand

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Slice-Schnitt legt die Naht zwischen einem Dienst, der ein Pflichtfeld einführt (`New` weist nil ab), und der Verdrahtung im Composition Root, die das Feld füllt. Der Plan weist die Verdrahtung dem Nehmer zu, ohne zu prüfen, ob der Stand nach dem Geber übersetzt und grün ist; ist das Feld Pflicht, scheitert `Build`, und jeder Test über den Root (hier 26 und 19) und damit `make gates` sind rot. Der Implementer muss eine Brücke einziehen, die der Plan nicht kannte, und der Nehmer wird ein Slice mit halbem Ziel. Nahe an `berichtigung-erreicht-nicht-jede-stelle` und `sendung-nennt-nicht-alle-pflichten-des-vertrags`.
