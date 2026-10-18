# SPEC-038 — Determinismus-Test (`make test-determinism`)

**Werkzeug:** `make test-determinism`.

**Festlegung:** 100 Wiederholungen identischer Inputs; identische normalisierte Outputs und Cursor-Stände.

**Stand:** entsteht mit dem Slice, der den Code liefert.

**Grenze:** Normalisiert nur die flüchtigen Header aus `VOLATILE_HEADERS`; jede andere Abweichung ist ein Befund.
