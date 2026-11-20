# Das Verweisverbot der Spec-Straten wird auf einen Werkzeugvertrag unter `harness/vertraege/` übertragen

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Klasse `spec-straten` verbietet einen Verweis nach unten nur für die drei Dateien unter `spec/`. Ein Plan nimmt an, derselbe Satz gelte für einen Vertrag unter `harness/vertraege/`, und führt daraus Folgen ab (kein Anker im Vertrag, ein Negativ mit der Begründung des Verbots); [MR-001](../../../../../../harness/conventions/MR-001-harness-vertraege-ausserhalb-der-spezifikation.md) Nr. 5 sagt das Gegenteil, und `make docs-check` lässt den Verweis zu.
