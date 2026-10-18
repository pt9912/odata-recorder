"""Negativtests des generierten SQLite-DDL (SPEC-004). Aufruf: check.py <schema.sql>.

Jeder Fall nennt die Zusage, die das Schema halten muss; ein Fall besteht, wenn die
Anweisung abgewiesen wird (reject) bzw. durchgeht (accept). Löschfälle sind isoliert:
das Objekt hat genau eine abhängige Zeile, so dass nur die geprüfte Regel abweisen kann.
Exit 0 nur ohne Abweichung. Nur-transaktionale und JSON-inhaltliche Invarianten
(SPEC-004) prüft der Anwendungstest, nicht dieses Skript.
"""
import sqlite3
import sys

db = sqlite3.connect(":memory:")
db.execute("PRAGMA foreign_keys = ON")
db.executescript(open(sys.argv[1], encoding="utf-8").read())

for sql in [
    "INSERT INTO services VALUES ('svc','http://x/','4.0',NULL,NULL)",
    "INSERT INTO services VALUES ('svc2','http://y/','4.0',NULL,NULL)",
    "INSERT INTO sessions VALUES ('s1','svc','t','open')",
    "INSERT INTO sessions VALUES ('s2','svc2','t','open')",
    "INSERT INTO snapshots VALUES ('b1','svc','t','{}','h','m','baseline',NULL,NULL)",
    "INSERT INTO snapshots VALUES ('w1','svc','t','{}','h','m','workspace','n1','ws1')",
    "INSERT INTO workspaces VALUES ('ws1','svc','b1','baseline',0,'t')",
]:
    db.execute(sql)

COLS = ["id", "session_id", "sequence_no", "method", "path", "raw_query", "match_key", "started_at",
        "finished_at", "key_redacted", "request_headers_json", "request_body_state", "request_blob",
        "response_status", "response_headers_json", "response_body_state", "response_blob",
        "capture_status", "request_redaction_status", "response_redaction_status",
        "body_integrity", "replay_equivalence"]
DEFAULT = dict(id="e", session_id="s1", sequence_no=1, method="GET", path="/p", raw_query="",
               match_key="k", started_at="t", finished_at=None, key_redacted=0,
               request_headers_json="{}", request_body_state="absent", request_blob=None,
               response_status=200, response_headers_json="{}", response_body_state="absent",
               response_blob=None, capture_status="complete", request_redaction_status="none",
               response_redaction_status="none", body_integrity="verified", replay_equivalence="byte")

fails = []


def report(name, ok):
    print(("ok   " if ok else "FAIL ") + name)
    if not ok:
        fails.append(name)


def run(sql, args=()):
    """True, wenn die Anweisung durchgeht; False, wenn eine Constraint abweist."""
    db.execute("SAVEPOINT c")
    try:
        db.execute(sql, args)
        return True
    except sqlite3.IntegrityError:
        return False
    finally:
        db.execute("ROLLBACK TO c")
        db.execute("RELEASE c")


def exchange(**kw):
    row = dict(DEFAULT, **kw)
    return ("INSERT INTO exchanges (%s) VALUES (%s)" % (",".join(COLS), ",".join("?" * len(COLS))),
            [row[c] for c in COLS])


def ex(name, accept=False, **kw):
    report(name, run(*exchange(**kw)) == accept)


def case(name, sql, accept=False):
    report(name, run(sql) == accept)


ex("Exchange mit konsistentem Body-Zustand", True, request_body_state="present", request_blob="b")
ex("Request-Body present ohne Blob", request_body_state="present")
ex("Request-Body absent mit Blob", request_blob="b")
ex("Response-Body present ohne Blob", response_body_state="present")
ex("Response-Body absent mit Blob", response_blob="b")
ex("Request-Body-Zustand ausserhalb der Menge", request_body_state="unknown")
ex("Response-Body-Zustand ausserhalb der Menge", response_body_state="unknown")
ex("capture_status ausserhalb der Menge", capture_status="x")
ex("request_redaction_status ausserhalb der Menge", request_redaction_status="x")
ex("response_redaction_status ausserhalb der Menge", response_redaction_status="x")
ex("body_integrity ausserhalb der Menge", body_integrity="x")
ex("replay_equivalence ausserhalb der Menge", replay_equivalence="x")
ex("key_redacted ausserhalb 0/1", key_redacted=2)
ex("sequence_no 0", sequence_no=0)
ex("Verworfener Response-Body bei complete", response_body_state="discarded")
ex("Verworfener Request-Body bei complete", request_body_state="discarded")
ex("Verworfener Body bei incomplete", True, response_body_state="discarded", capture_status="incomplete")
ex("Response unsafe ohne verworfenen Body", response_redaction_status="unsafe", replay_equivalence="none")
ex("Response unsafe mit Replay-Aequivalenz", response_redaction_status="unsafe",
   response_body_state="discarded", capture_status="incomplete", replay_equivalence="byte")
ex("Response unsafe konsistent", True, response_redaction_status="unsafe", response_body_state="discarded",
   capture_status="incomplete", replay_equivalence="none")
ex("Request unsafe ohne capture failed", request_redaction_status="unsafe", request_body_state="discarded")
ex("Request unsafe konsistent", True, request_redaction_status="unsafe", request_body_state="discarded",
   capture_status="failed")

db.execute(*exchange(id="e1", sequence_no=1))
ex("sequence_no je Session eindeutig", id="e2", sequence_no=1)
ex("Exchange ohne Session", session_id="nope")

case("Session-Status ausserhalb der Menge", "INSERT INTO sessions VALUES ('s3','svc','t','x')")
case("Snapshot baseline mit workspace_id", "INSERT INTO snapshots VALUES ('b2','svc','t','{}','h','m','baseline','n','ws1')")
case("Snapshot workspace ohne workspace_id", "INSERT INTO snapshots VALUES ('w2','svc','t','{}','h','m','workspace','n',NULL)")
case("Snapshot workspace ohne Namen", "INSERT INTO snapshots VALUES ('w2','svc','t','{}','h','m','workspace',NULL,'ws1')")
case("Snapshot-Art ausserhalb der Menge", "INSERT INTO snapshots VALUES ('x1','svc','t','{}','h','m','other',NULL,NULL)")
case("Snapshot-Name je Workspace eindeutig", "INSERT INTO snapshots VALUES ('w3','svc','t','{}','h','m','workspace','n1','ws1')")
case("Workspace ohne Baseline-Snapshot", "INSERT INTO workspaces VALUES ('ws2','svc','nope','baseline',0,'t')")
case("Workspace nimmt Workspace-Snapshot als Baseline", "INSERT INTO workspaces VALUES ('ws2','svc','w1','baseline',0,'t')")
case("Workspace mit fremdem Service der Baseline", "INSERT INTO workspaces VALUES ('ws2','svc2','b1','baseline',0,'t')")
case("Workspace baseline_kind ausserhalb 'baseline'", "INSERT INTO workspaces VALUES ('ws2','svc','b1','workspace',0,'t')")
case("Workspace-Revision negativ", "INSERT INTO workspaces VALUES ('ws2','svc','b1','baseline',-1,'t')")
case("Cursor: Session eines anderen Service", "INSERT INTO replay_cursors VALUES ('ws1','r','s2','sequential',0,'[]','svc')")
case("Cursor: Workspace eines anderen Service", "INSERT INTO replay_cursors VALUES ('ws1','r','s2','sequential',0,'[]','svc2')")
case("Cursor mit unbekannter Strategie", "INSERT INTO replay_cursors VALUES ('ws1','r','s1','random',0,'[]','svc')")
case("Cursor mit negativer Position", "INSERT INTO replay_cursors VALUES ('ws1','r','s1','strict',-1,'[]','svc')")
case("Entity mit unbekannter Vollstaendigkeit", "INSERT INTO entities VALUES ('ws1','E','1','{}','[]','x',0,0)")
case("Entity mit negativer Version", "INSERT INTO entities VALUES ('ws1','E','1','{}','[]','complete',-1,0)")
case("Entity deleted ausserhalb 0/1", "INSERT INTO entities VALUES ('ws1','E','1','{}','[]','complete',0,2)")
case("Entity ohne Workspace", "INSERT INTO entities VALUES ('nope','E','1','{}','[]','complete',0,0)")
case("Coverage-Status ausserhalb der Menge", "INSERT INTO coverage VALUES ('b1','baseline','E','x','{}')")
case("Coverage fuer Workspace-Snapshot", "INSERT INTO coverage VALUES ('w1','workspace','E','complete','{}')")
case("Coverage mit falscher Art fuer Baseline", "INSERT INTO coverage VALUES ('b1','workspace','E','complete','{}')")
case("Baseline-Quelle fuer Workspace-Snapshot", "INSERT INTO baseline_sources VALUES ('w1','workspace','s1')")
case("Import ohne Baseline", "INSERT INTO imports (id,request_hash,service_id,created_at,report_json) VALUES ('i0','h','svc','t','{}')")
case("Import mit fremdem Service", "INSERT INTO imports VALUES ('i3','h3','b1','svc2','t','{}')")
case("Erster Import je Baseline", "INSERT INTO imports VALUES ('i1','h','b1','svc','t','{}')", True)

db.execute("INSERT INTO imports VALUES ('i1','h','b1','svc','t','{}')")
case("Zweiter Import derselben Baseline", "INSERT INTO imports VALUES ('i2','h2','b1','svc','t','{}')")

# Isolierte Löschfälle: jeweils genau eine abhängige Zeile.
db.execute("INSERT INTO snapshots VALUES ('b3','svc','t','{}','h','m','baseline',NULL,NULL)")
db.execute("INSERT INTO sessions VALUES ('s5','svc','t','open')")
db.execute("INSERT INTO baseline_sources VALUES ('b3','baseline','s5')")
case("Session als Baseline-Quelle nicht loeschbar", "DELETE FROM sessions WHERE id='s5'")
case("Baseline loeschen nimmt Baseline-Quellen mit", "DELETE FROM snapshots WHERE id='b3'", True)
db.execute("INSERT INTO snapshots VALUES ('b4','svc','t','{}','h','m','baseline',NULL,NULL)")
db.execute("INSERT INTO imports VALUES ('i4','h4','b4','svc','t','{}')")
case("Baseline mit Import nicht loeschbar", "DELETE FROM snapshots WHERE id='b4'")
case("Baseline mit Workspace nicht loeschbar", "DELETE FROM snapshots WHERE id='b1'")
db.execute("INSERT INTO coverage VALUES ('b4','baseline','E','complete','{}')")
db.execute("DELETE FROM imports WHERE id='i4'")
db.execute("DELETE FROM snapshots WHERE id='b4'")
n = db.execute("SELECT count(*) FROM coverage").fetchone()[0]
report("Baseline loeschen nimmt Coverage mit", n == 0)

db.execute("INSERT INTO sessions VALUES ('s4','svc','t','open')")
db.execute("INSERT INTO replay_cursors VALUES ('ws1','r','s4','sequential',0,'[]','svc')")
db.execute("DELETE FROM sessions WHERE id='s4'")
n = db.execute("SELECT count(*) FROM replay_cursors").fetchone()[0]
report("Session loeschen nimmt ihre Cursor mit", n == 0)

db.execute("INSERT INTO entities VALUES ('ws1','E','1','{}','[]','complete',0,0)")
db.execute("INSERT INTO replay_cursors VALUES ('ws1','r','s1','sequential',0,'[]','svc')")
db.execute("DELETE FROM workspaces WHERE id='ws1'")
n = db.execute("SELECT (SELECT count(*) FROM entities)+(SELECT count(*) FROM replay_cursors)").fetchone()[0]
report("Workspace loeschen nimmt Entities und Cursor mit", n == 0)

plan = " ".join(r[3] for r in db.execute(
    "EXPLAIN QUERY PLAN SELECT id FROM exchanges WHERE session_id='s1' AND match_key='k'"))
report("Match-Key-Index wird genutzt", "idx_exchange_match" in plan)

bad = db.execute("PRAGMA foreign_key_check").fetchall()
ic = db.execute("PRAGMA integrity_check").fetchone()[0]
report("foreign_key_check", not bad)
report("integrity_check", ic == "ok")

sys.exit(1 if fails else 0)
