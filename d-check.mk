# d-check.mk — Doku-Referenz-Gate via d-check. Emittiert von ai-harness-init,
# adaptiert aus `d-check --print-mk`: doc-check -> docs-check (das Befund-Gate,
# einziges als Gate behauptetes Target) und DCHECK_DIGEST auf den erzeugenden
# Image-Digest gepinnt (Reproduzierbarkeit). advisory doc-*-Targets verbatim, ohne
# die Kennungen aus dem Register von d-check in Kommentaren und Hilfetexten.
# Einbinden: `include d-check.mk`; eigene .d-check.yml danebenlegen.
DCHECK_IMAGE ?= ghcr.io/pt9912/d-check:v0.86.1
DCHECK_DIGEST ?= sha256:3e0b9779a71e2fa942960e8b961428fba036455c535513d34a75bfb799ffce0e
# TRACE_FLAGS: optionale Flags für die RTM-Targets (z. B. --json).
TRACE_FLAGS ?=

# Ein gesetzter DCHECK_DIGEST sticht den Tag von DCHECK_IMAGE.
ifeq ($(strip $(DCHECK_DIGEST)),)
DCHECK_REF := $(DCHECK_IMAGE)
else
DCHECK_REF := ghcr.io/pt9912/d-check@$(DCHECK_DIGEST)
endif

.PHONY: docs-check
docs-check: ## Doku-Referenzen prüfen (Befund-Gate)
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF)

.PHONY: doc-trace
doc-trace: ## Requirements Traceability Matrix auf stdout (advisory)
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --trace $(TRACE_FLAGS)

.PHONY: doc-complete
doc-complete: ## Vollständigkeits-Gate: Requirements-Waise ⇒ Exit 1
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --trace --require-complete $(TRACE_FLAGS)

.PHONY: doc-doctor
doc-doctor: ## erklärende Diagnose mit Fix-Kandidaten
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --doctor

.PHONY: doc-repair
doc-repair: ## Reparatur-Patch (unified diff) auf stdout, git-apply-rein
	@docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --repair

.PHONY: doc-immutable
doc-immutable: ## Doc-/ADR-Immutabilität via git-Diff (Modul vcs); RANGE=base..head oder STAGED=1
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --enable vcs --disable links --disable anchors --disable ids --disable matrix --disable external --disable codepaths --disable spans --disable hostpaths --disable diagrams --disable versions --disable pins --disable immutable --disable commits --disable planning --disable tracked --disable targets --disable citations --disable sources --disable structure --disable workflows --disable reviews --disable mentions --disable file $(if $(STAGED),--staged,--range $(RANGE))

.PHONY: doc-commits
doc-commits: ## Commit-Message-Traceability via Modul commits; RANGE=base..head
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --enable commits --disable links --disable anchors --disable ids --disable matrix --disable external --disable codepaths --disable spans --disable hostpaths --disable diagrams --disable versions --disable pins --disable immutable --disable vcs --disable planning --disable tracked --disable targets --disable citations --disable sources --disable structure --disable workflows --disable reviews --disable mentions --disable file --range $(RANGE)

.PHONY: doc-planning
doc-planning: ## Planning-Lifecycle-Konsistenz (Roadmap <-> in-progress) via Modul planning; hermetisch, ohne Range
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --enable planning --disable links --disable anchors --disable ids --disable matrix --disable external --disable codepaths --disable spans --disable hostpaths --disable diagrams --disable versions --disable pins --disable immutable --disable vcs --disable commits --disable tracked --disable targets --disable citations --disable sources --disable structure --disable workflows --disable reviews --disable mentions --disable file

.PHONY: doc-tracked
doc-tracked: ## Getrackt-Status aufloesbarer Referenz-Ziele via Modul tracked; braucht .git im Mount, ohne Range
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --enable tracked --disable links --disable anchors --disable ids --disable matrix --disable external --disable codepaths --disable spans --disable hostpaths --disable diagrams --disable versions --disable pins --disable immutable --disable vcs --disable commits --disable planning --disable targets --disable citations --disable sources --disable structure --disable workflows --disable reviews --disable mentions --disable file

.PHONY: doc-targets
doc-targets: ## Deklarations-Konsistenz Doku<->Build-Targets via Modul targets; hermetisch, ohne Range
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --enable targets --disable links --disable anchors --disable ids --disable matrix --disable external --disable codepaths --disable spans --disable hostpaths --disable diagrams --disable versions --disable pins --disable immutable --disable vcs --disable commits --disable planning --disable tracked --disable citations --disable sources --disable structure --disable workflows --disable reviews --disable mentions --disable file

.PHONY: doc-structure
doc-structure: ## Struktur-Invarianten innerhalb der Dokumente via Modul structure; hermetisch, ohne Range
	docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --enable structure --disable links --disable anchors --disable ids --disable matrix --disable external --disable codepaths --disable spans --disable hostpaths --disable diagrams --disable versions --disable pins --disable immutable --disable vcs --disable commits --disable planning --disable tracked --disable targets --disable citations --disable sources --disable workflows --disable reviews --disable mentions --disable file

.PHONY: doc-usage
doc-usage: ## Aufruf und Optionen von d-check selbst (--help)
	@docker run --rm --network none -v "$(CURDIR):/repo:ro" $(DCHECK_REF) --help

.PHONY: doc-help
doc-help: ## diese Liste der doc-*-Targets
	@grep -hE '^docs?-[a-z-]+:.*## ' $(MAKEFILE_LIST) | sort | sed -E 's/:.*## /  /'
