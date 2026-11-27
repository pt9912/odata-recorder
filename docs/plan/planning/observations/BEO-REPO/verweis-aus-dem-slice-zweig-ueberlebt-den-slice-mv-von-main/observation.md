# Ein im Slice-Zweig geschriebener Verweis auf einen Nachbar-Plan bricht beim Merge eines `slice-mv` von `main`

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

`make slice-mv` zieht die Verweise nach, die es auf seinem Zweig sieht. Schreibt ein anderer Zweig vorher einen Verweis auf den alten Ort desselben Plans, bringt der Merge von `main` den Move, aber nicht den Nachzug dieser Zeile; Git meldet keinen Konflikt. Läuft nach dem Merge kein Doku-Gate, trägt der Stand den toten Link bis zur nächsten Rolle, die `make gates` fährt. Nahe an `roadmap-links-brechen-im-zweig-nach-dem-start-move`: dort ist die Ursache der Start-Move im eigenen Zweig, hier der Move eines fremden Slice.
