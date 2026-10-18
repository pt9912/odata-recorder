# harness/mk/enforce.mk — Enforce-Fragment, emittiert von ai-harness-init. Traegt NUR das
# record-gates-Rezept (Gate-Nachweis-Stempel). Die Ordnungskante und gates: record-gates
# leben im Root-Aggregator (gen), weil sie die akkumulierten Checks erst NACH dem
# Glob-Include vollstaendig sehen (Fragment-Assembly).
.PHONY: record-gates

record-gates: ## Gate-Nachweis stempeln (nur nach gruenen Checks)
	@bash tools/harness/record-gates.sh
