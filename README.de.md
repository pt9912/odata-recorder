# odata-recorder

[English](README.md) · **Deutsch**

> **Zeichnet die Kommunikation zwischen Ihrem OData-Client und einem OData-Dienst auf, spielt sie ohne Dienst reproduzierbar wieder ab und stellt daraus einen zustandsbehafteten OData-Mockdienst bereit.**

## Was ist odata-recorder?

`odata-recorder` ist ein Werkzeug für Softwareentwickler und Testautomatisierung.
Es vermittelt zwischen einem OData-Client und einem bestehenden OData-Dienst,
zeichnet die HTTP/OData-Austausche in Recording-Sessions auf und gibt sie später
deterministisch wieder. Auf Grundlage der Metadaten und der aufgezeichneten Daten
stellt es außerdem einen lokal veränderbaren Mockdienst bereit, mit
CRUD-Operationen, Abfrageoptionen, isolierten Workspaces und Snapshots.
Integrationstests laufen so ohne dauerhaft erreichbaren Produktiv- oder
Testserver. Es richtet sich an Entwickler, Tester und CI-Systeme. Der erste
Abnahmeumfang ist OData V4 mit JSON und XML-Metadaten. Die Anforderungen stehen im
[Lastenheft](spec/lastenheft.md).

## Was kann ich heute tun?

Es gibt noch keinen lauffähigen Code. Vorhanden sind die Spezifikation und der
Harness, der sie absichert:

- Das [Lastenheft](spec/lastenheft.md) (Entwurf, Version 0.5.0) legt die
  Anforderungen mit Akzeptanzkriterien fest, die [Spezifikation](spec/spezifikation.md)
  die technischen Details und die [Architektur](spec/architecture.md) die Komponenten.
- `make gates` läuft grün: `baseline-verify` prüft die vendored Baseline, `docs-check`
  die Doku-Referenzen. `make help` zeigt alle Targets.

## Warum odata-recorder?

Integrationstests gegen OData-Dienste brauchen häufig einen erreichbaren Dienst
mit passendem Datenbestand. Das macht lokale Umgebungen und CI-Läufe aufwendig,
langsam und schwer reproduzierbar, und Tests können den Dienst unbeabsichtigt
verändern. Für viele dieser Tests ist nicht der Dienst der Gegenstand, sondern das
aus Sicht des Clients beobachtbare Verhalten. Genau dieses Verhalten zeichnet
`odata-recorder` auf und macht es offline wiederholbar.

## Kerngedanke

Eine Aufzeichnung enthält nicht zwingend einen vollständigen Datenbestand. Der
Recorder täuscht keine Vollständigkeit vor: Unvollständige Abdeckung bleibt
erkennbar, und ein Speicherfehler wird gemeldet statt als erfolgreiche Aufzeichnung
ausgegeben. Eine vollständige Simulation beliebiger Geschäftslogik wird nicht
verlangt.

## Was macht es vertrauenswürdig?

- **Prozess:** [`AGENTS.md`](AGENTS.md) (Hard Rules: die Regeln, die jede Änderung einhalten muss), [`harness/README.md`](harness/README.md) (Source Precedence: welche Quelle bei Konflikt gewinnt, und die Gates).
- **Verträge:** [`spec/lastenheft.md`](spec/lastenheft.md) (`LH-*`-IDs mit Akzeptanzkriterien), danach [`spec/spezifikation.md`](spec/spezifikation.md) und [`spec/architecture.md`](spec/architecture.md).
- **Gates:** `make gates` führt die Gates aus; welche es sind und was jedes prüft, steht in [`harness/README.md` §Sensors](harness/README.md#sensors-feedback-gates).
- **Auditierbarkeit:** Entscheidungen in [`docs/plan/adr/`](docs/plan/adr/), Planung in [`docs/plan/planning/`](docs/plan/planning/).
- **Rückverfolgbarkeit:** `make doc-trace` gibt die Requirements Traceability Matrix aus (Anforderung, Entscheidungen, Slices); sie ist ein Bericht und kein Gate.

## Lizenz

[MIT](LICENSE)
