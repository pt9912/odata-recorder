# Die Spezifikation legt einen Bezeichner an einen Ort, den das Gate desselben Bezeichners nicht liest

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Ein Vertrag der Spezifikation nennt die Datei, in der ein Name als Konstante steht; der Vertrag eines Gates liest dieselbe Art von Namen allein aus anderen Dateien. Beide Zusagen sind einzeln erfüllbar und zusammen nicht: das Handbuch muss den Namen führen, und das Gate meldet ihn als fremd. Kein Gate hält die zwei Verträge gegeneinander. Nahe an `plan-nennt-datei-die-fremde-waechter-verbieten`: dort sperrt ein fremder Quelltest eine Datei des Plans.
