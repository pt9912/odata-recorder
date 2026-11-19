# Eine Sendung sagt „Zustand unverändert“ nach einem Fehler, und der Fänger des Nehmers liest den Zustand nicht zurück

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Eine Sendung an einen Nehmer-Slice trägt zwei Zusagen in einem Satz („0 Verbindungen und Zustand unverändert“); der Nehmer übernimmt sie als einen DoD-Punkt, und die Zeile seiner Tabelle *Zusage · Fänger* führt nur die erste Hälfte. Die Hälfte „Zustand unverändert“ trägt dann allein das Verhalten eines anderen Dienstes (hier die Transaktion des Mock), und in diesem Slice hat sie weder Fänger noch ein benanntes Negativ; die Prüfung der Sendungen (`AGENTS.md` §3.13) liest §1 *Ausdrücklich NICHT* und die DoD der Gegenstelle, nicht Hälfte für Hälfte die Fänger-Zeile. Nahe an `sendung-nennt-nicht-alle-pflichten-des-vertrags` (dort fehlt die Pflicht in der Sendung, hier die Hälfte im Fänger des Nehmers) und `akzeptiertes-negativ-ohne-folge`.

**Rest (benannt):** ein Fall in `tests/integration/m4_test.go` (`ohneAusloeser`), der nach dem `412` im Modus `hybrid` die Entity zurückliest (`GET` der Entity nach dem `PATCH` gleich dem Stand davor) und mit dem ETag einer früheren Version statt `"nope"` fährt. Adresse: Planner, Sichtungs-Schritt der nächsten Planung eines Slice, der diese Datei oder die Fälle des Mock bei `412` berührt.
