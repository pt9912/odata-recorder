# ADR-0008: Blob-Bereinigung im laufenden Betrieb mit Wächter und Zurückstellen

**Status:** Proposed

**Datum:** 2026-10-23

**Autor:** pt9912

**Bezug:** [`LH-RB-04`](../../../spec/lastenheft.md#lh-rb-04--schutz-von-daten-und-zugangsdaten), [`LH-FA-03`](../../../spec/lastenheft.md#lh-fa-03--persistenz-und-wiederverwendung-muss), [`LH-QA-03`](../../../spec/lastenheft.md#lh-qa-03--isolation-und-konsistenz), [ADR-0006](0006-sqlite-bibliothek-ohne-native-abhaengigkeit.md), [ADR-0007](0007-race-detektor-im-gate.md)

**Schärft:** [`SPEC-123`](../../../spec/spezifikation.md#spec-123--wartung-vertrag-von-maintenanceusecase)

**Regeln:** Baseline-Regelwerk `modul-04-adrs.md` §Ziel-Form: ADR (MADR).

---

## Kontext

Ein Blob wird content-adressiert abgelegt und erst danach von einer Zeile der Datenbank referenziert (Exchange, Snapshot, Service-Metadaten).
Dazwischen ist er unreferenziert und sieht aus wie ein verwaister, den die Bereinigung entfernen darf; entfernt sie ihn, zeigt die Referenz danach auf eine fehlende Datei.
Gleiche Bytes haben dieselbe Datei, ein `Commit` auf einen vorhandenen Blob legt nichts an, und ein `Remove` gegen einen `Commit` gleicher Bytes ist im Port offen.
Beim Start läuft kein Schreiber, im laufenden Betrieb nach dem Löschen einer Session oder eines Workspace laufen alle.
Das Lastenheft verlangt die Entfernung „unmittelbar“ nur als `best effort` und „in jedem Fall beim nächsten Start“.
Die Fenster sind zwei: der Exchange des Record-Dienstes zwischen `Commit` und `Append` (ohne Obergrenze, bis der HTTP-Adapter den Strom zu Ende liest) und die übrigen Schreiber, deren Aufruf zwischen `Commit` und Referenz kurz ist (Snapshot, Import, Metadaten).
Annahmen: Alle Schreiber von Blobs erhalten dasselbe `BlobRepository`-Objekt vom Composition Root; ein Exchange läuft nur in einer Session im Zustand `open`, und diese verlässt ihn erst nach der Rückkehr von `Drain`.
Kippt eine der beiden, kippt die Entscheidung.

## Entscheidung

Wir stellen einen **Wächter** vor das `BlobRepository`, den jeder Schreiber von Blobs und die Wartung teilen, und wir **stellen die Bereinigung im Betrieb zurück, solange eine Session offen ist**.
Der Wächter ordnet `Commit` (Lesersperre) und den Abgleich der Wartung (Schreibersperre) und merkt die Zeit jedes gelungenen `Commit`; die Bereinigung im Betrieb entfernt keinen Blob, dessen `Commit` noch nicht eine feste Karenz von zehn Minuten zurückliegt; die Karenz ist ein fester Wert und kein Konfigurationsschlüssel (Option F).
Bei einer offenen Session entfernt sie nichts und überlässt es dem nächsten Start, der in jedem Fall bereinigt.
Beim Start gilt weder das Zurückstellen noch die Karenz.
Was die Wartung im Einzelnen prüft, in welcher Reihenfolge und mit welchen Folgen, steht in `SPEC-123`.

## Verglichene Alternativen

| Option | Pro | Contra |
|---|---|---|
| A — im Betrieb nie entfernen, nur beim Start | keine Fenster, kein Wächter | „unmittelbar“ entfällt auch dort, wo es möglich wäre (Modus `replay`, alle Sessions beendet); die Entfernung wartet immer auf einen Neustart |
| B — nur nach dem Lesen der Referenzmenge erneut prüfen (Referenz vor dem Entfernen) | kein neuer Baustein, wie `SPEC-048` es nennt | schließt das Fenster zwischen `Commit` und `Append` nicht: es dauert, bis der Client den Strom ganz liest; ein `Commit` gleicher Bytes gegen `Remove` bleibt offen |
| C — Zeit seit der letzten Änderung der Datei als Karenz (Zeit des Dateisystems) | braucht keinen Wächter vor dem Port | der Port liefert keine Zeit (Änderung von `BlobRepository`); Dateizeit und `Clock` sind zwei Uhren; die Karenz trägt keinen Exchange ohne Obergrenze |
| D — Schutz je Blob durch Zählen (der Schreiber meldet `Commit` und Referenz, der Wächter zählt) | genau, kein Zeitwert, deckt auch den Exchange | jeder Schreiber muss das Freigeben rufen, auch der Record-Dienst (andere Slices, einer abgeschlossen); ein vergessenes Freigeben hält den Blob bis zum Start; ein Port mehr |
| F — Karenz als Konfigurationsschlüssel (abgelehnt) | der Betrieb könnte die Karenz an einen langsamen Datenträger anpassen | die Karenz ist nur sicher, solange sie über der Dauer eines Schreiber-Aufrufs liegt (`SQLITE_BUSY_TIMEOUT`, Default 5 s); diesen Zusammenhang zweier Schlüssel prüft die Konfiguration nicht, und ein zu kleiner Wert öffnet das Fenster still; der feste Wert liegt beim Hundertfachen des Defaults und ein Schlüssel mehr trägt keinen Fall, den das Lastenheft verlangt |
| **E — Wächter mit Karenz und Zurückstellen bei offener Session (gewählt)** | deckt den Exchange ohne Mitwirkung des Record-Dienstes (offene Session) und die kurzen Fenster der übrigen Schreiber (Karenz); ordnet `Commit` und `Remove` gleicher Bytes; beim Start gilt alles | im Modus `record` bleiben die Blobs gelöschter Sessions bis zum nächsten Start liegen, solange die erste Session offen ist; die Karenz ist eine Annahme über die Dauer eines Aufrufs; jeder Schreiber muss durch den Wächter laufen |

## Konsequenzen

- Positiv: Kein Blob, den ein Exchange, ein Snapshot, ein Import oder die Metadaten eines Service eben committet haben, wird von der Bereinigung entfernt, solange alle Schreiber durch den Wächter laufen; ein `Commit` gleicher Bytes gegen ein `Remove` endet in einer Datei, die besteht.
- Negativ: Im Modus `record` liegen die Blobs gelöschter Sessions und Workspaces bis zum nächsten Start auf dem Datenträger (`LH-RB-04`: der Löschzeitpunkt ist der nächste Start). Ein Schreiber, der den Wächter umgeht, ist ungeschützt; eine Referenz, die nach der zweiten Lesung entsteht, ebenso (Grund und Folge in `SPEC-123` *Ungeprüft*). Die Sperre des Abgleichs hält Commits auf, solange er läuft.
- Folgepflicht: Vertrag `SPEC-123` samt Nachzügen; der Composition Root erzeugt den Wächter einmal je Prozess und gibt ihn jedem Schreiber von Blobs (Verdrahtung und ihr Fänger im Slice des Composition Root); die Fänger der Fenster stehen in der Tabelle *Zusage · Fänger* des Slice der Wartung. Eine spätere Entscheidung, den Record-Dienst die Blobs zwischen `Commit` und `Append` melden zu lassen (Option D), löste das Zurückstellen ab und wäre ein Folge-ADR mit `Supersedes`.

## Fitness Function (falls maschinell prüfbar)

| Tooling | Regel | Make-Target |
|---|---|---|
| `go test` und `go test -race` über die Pakete der Wartung | `Commit` und `Remove` gleicher Bytes enden in einer bestehenden Datei; eine Bereinigung bei offener Session entfernt nichts; kein Datenrennen auf dem Zustand des Wächters | `make test`, `make test-race` |

## Re-Evaluierungs-Trigger

Wenn der Record-Dienst die Blobs zwischen `Commit` und `Append` melden kann (Option D wird billig), wenn ein Betrieb im Modus `record` die Entfernung im Betrieb braucht, oder wenn ein Schreiber von Blobs mit einem Aufruf zwischen `Commit` und Referenz länger als die Karenz hinzukommt.

## Geschichte

| Datum | Ereignis | Verweis |
|---|---|---|
| 2026-10-23 | Proposed | Entwurf des Architect; die Annahme steht beim Nutzer aus |
