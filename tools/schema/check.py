"""Negativtests des generierten SQLite-DDL (SPEC-004). Aufruf: check.py <schema.sql>.

Jeder Fall nennt die Zusage, die das Schema halten muss; ein Fall besteht, wenn die
Anweisung abgewiesen wird (reject) bzw. durchgeht (accept). Exit 0 nur ohne Abweichung.
"""
import sqlite3
import sys

db = sqlite3.connect(":memory:")
db.execute("PRAGMA foreign_keys = ON")
db.executescript(open(sys.argv[1], encoding="utf-8").read())

BASE = [
    "INSERT INTO services VALUES ('svc','http://x/','4.0',NULL,NULL)",
    "INSERT INTO services VALUES ('svc2','http://y/','4.0',NULL,NULL)",
    "INSERT INTO sessions VALUES ('s1','svc','t','open')",
    "INSERT INTO sessions VALUES ('s2','svc2','t','open')",
    "INSERT INTO snapshots VALUES ('b1','svc','t','{}','h','m','baseline',NULL,NULL)",
    "INSERT INTO snapshots VALUES ('w1','svc','t','{}','h','m','workspace','n1','ws1')",
    "INSERT INTO workspaces VALUES ('ws1','svc','b1','baseline',0,'t')",
]
for sql in BASE:
    db.execute(sql)

EX = ("INSERT INTO exchanges VALUES ('{id}','s1',{seq},'GET','/p','','k','t',NULL,0,'{{}}',"
      "'{rb}',{rblob},200,'{{}}','{pb}',{pblob},'complete','none','none','verified','byte')")


def exchange(id, seq, rb="absent", rblob="NULL", pb="absent", pblob="NULL"):
    return EX.format(id=id, seq=seq, rb=rb, rblob=rblob, pb=pb, pblob=pblob)


fails = []


def case(name, sql, accept=False):
    db.execute("SAVEPOINT c")
    try:
        db.execute(sql)
        ok = accept
    except sqlite3.IntegrityError:
        ok = not accept
    db.execute("ROLLBACK TO c")
    db.execute("RELEASE c")
    if not ok:
        fails.append(name)
    print(("ok   " if ok else "FAIL ") + name)


case("Exchange mit konsistentem Body-Zustand", exchange("e", 1, "present", "'b'", "absent"), True)
case("Body present ohne Blob", exchange("e", 1, "present", "NULL"))
case("Body absent mit Blob", exchange("e", 1, "absent", "'b'"))
case("Body-Zustand ausserhalb der Menge", exchange("e", 1, "unknown"))
db.execute(exchange("e1", 1))
case("sequence_no je Session eindeutig", exchange("e2", 1))
case("Session-Status ausserhalb der Menge", "INSERT INTO sessions VALUES ('s3','svc','t','x')")
case("Snapshot baseline mit workspace_id", "INSERT INTO snapshots VALUES ('b2','svc','t','{}','h','m','baseline','n','ws1')")
case("Snapshot workspace ohne Namen", "INSERT INTO snapshots VALUES ('w2','svc','t','{}','h','m','workspace',NULL,'ws1')")
case("Snapshot-Name je Workspace eindeutig", "INSERT INTO snapshots VALUES ('w3','svc','t','{}','h','m','workspace','n1','ws1')")
case("Workspace ohne Baseline-Snapshot", "INSERT INTO workspaces VALUES ('ws2','svc','nope','baseline',0,'t')")
case("Workspace nimmt Workspace-Snapshot als Baseline", "INSERT INTO workspaces VALUES ('ws2','svc','w1','baseline',0,'t')")
case("Workspace mit fremdem Service der Baseline", "INSERT INTO workspaces VALUES ('ws2','svc2','b1','baseline',0,'t')")
case("Workspace-Revision negativ", "INSERT INTO workspaces VALUES ('ws2','svc','b1','baseline',-1,'t')")
case("Cursor auf Session eines anderen Service", "INSERT INTO replay_cursors VALUES ('ws1','r','s2','sequential',0,'[]','svc')")
case("Cursor mit unbekannter Strategie", "INSERT INTO replay_cursors VALUES ('ws1','r','s1','random',0,'[]','svc')")
case("Cursor mit negativer Position", "INSERT INTO replay_cursors VALUES ('ws1','r','s1','strict',-1,'[]','svc')")
case("Entity mit unbekannter Vollstaendigkeit", "INSERT INTO entities VALUES ('ws1','E','1','{}','[]','x',0,0)")
case("Coverage-Status ausserhalb der Menge", "INSERT INTO coverage VALUES ('b1','E','x','{}')")
case("Erster Import je Baseline", "INSERT INTO imports VALUES ('i1','h','b1','svc','t','{}')", True)

db.execute("INSERT INTO imports VALUES ('i1','h','b1','svc','t','{}')")
case("Zweiter Import derselben Baseline", "INSERT INTO imports VALUES ('i2','h2','b1','svc','t','{}')")
case("Import mit fremdem Service", "INSERT INTO imports VALUES ('i3','h3','w1','svc2','t','{}')")

db.execute("INSERT INTO baseline_sources VALUES ('b1','s1')")
case("Session als Baseline-Quelle nicht loeschbar", "DELETE FROM sessions WHERE id='s1'")
case("Baseline mit Import nicht loeschbar", "DELETE FROM snapshots WHERE id='b1'")

db.execute("INSERT INTO sessions VALUES ('s4','svc','t','open')")
db.execute("INSERT INTO replay_cursors VALUES ('ws1','r','s4','sequential',0,'[]','svc')")
db.execute("DELETE FROM sessions WHERE id='s4'")
n = db.execute("SELECT count(*) FROM replay_cursors").fetchone()[0]
print(("ok   " if n == 0 else "FAIL ") + "Session loeschen nimmt ihre Cursor mit")
if n:
    fails.append("Cursor-Cascade")

db.execute("INSERT INTO entities VALUES ('ws1','E','1','{}','[]','complete',0,0)")
db.execute("INSERT INTO replay_cursors VALUES ('ws1','r','s1','sequential',0,'[]','svc')")
db.execute("DELETE FROM workspaces WHERE id='ws1'")
n = db.execute("SELECT (SELECT count(*) FROM entities)+(SELECT count(*) FROM replay_cursors)").fetchone()[0]
print(("ok   " if n == 0 else "FAIL ") + "Workspace loeschen nimmt Entities und Cursor mit")
if n:
    fails.append("Workspace-Cascade")

bad = db.execute("PRAGMA foreign_key_check").fetchall()
ic = db.execute("PRAGMA integrity_check").fetchone()[0]
print(("ok   " if not bad else "FAIL ") + "foreign_key_check")
print(("ok   " if ic == "ok" else "FAIL ") + "integrity_check")
if bad or ic != "ok":
    fails.append("pragma")

sys.exit(1 if fails else 0)
