# Die Token-Datei meldet beim gleichzeitigen Erzeugen gelegentlich "keine reguläre Datei", weil der Vertrag den Wettlauf nicht nennt

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

`lade` in `src/adapters/driving/admin/tokendatei.go` unterscheidet "fehlt" von "Symlink ohne Ziel" mit `os.Stat` und danach `os.Lstat` (`istFehlend`). Verlinkt ein gleichzeitiger Aufruf die Datei zwischen beiden Aufrufen, ist `Stat` gescheitert und `Lstat` gelingt; `statFehler` bildet es auf den Grund "keine reguläre Datei" ab (`ODR-E2001 ADMIN_TOKEN_FILE`). `SPEC-239` Nr. 1 a entscheidet "fehlt" gegen "Symlink ohne Ziel" und nennt den Wettlauf nicht; Nr. 3 d (gleichzeitiges Erzeugen) sagt nicht, welchen Ausgang der Verlierer in diesem Fenster hat. Folge: `TestWettlaufNebenlaeufig` ist unter Last gelegentlich rot (3 von 5 Läufen im Container bei Load Average 55 bis 69); im Gate des Verifiers von `slice-admin-rahmen` war er grün.
