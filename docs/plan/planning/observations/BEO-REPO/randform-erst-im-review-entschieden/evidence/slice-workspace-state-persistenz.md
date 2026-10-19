**Vorgang:** slice-workspace-state-persistenz
**Fund:** `SPEC-108` ließ zwei Randformen offen (leere Kennung je Operation, `ReplaceAll` auf unbekanntem Workspace); der Implementer hielt an (`AGENTS.md` §3.12), sie sind vor dem Code entschieden. Das Review fand danach die Teilwirkung einer Operation in `f` bei verschlucktem Fehler und `f == nil` (Befunde F-3, F-8); `SPEC-108` nennt beide.
**Stand:** offen
