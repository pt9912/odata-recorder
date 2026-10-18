# harness/mk/arch-gate.mk — Architektur-Gate-Fragment.
# Bindet das tool-generierte a-check.mk ein (Gate a-check) und haengt a-check an
# GATE_CHECKS an; der Root-Aggregator faehrt es via make gates.
include a-check.mk

GATE_CHECKS += a-check
