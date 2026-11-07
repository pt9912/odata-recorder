# Das Image-Tag `odata-recorder:race` gehört allen Worktrees, ein `docker run` mit dem gebackenen Baum fährt den Code des zuletzt bauenden Zweigs

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Die Stufen des `Dockerfile` backen den Baum nach `/src` und tragen im Image-Tag keinen Zweig. Wer in einem Worktree ein `docker run` gegen das Tag fährt, ohne den eigenen Baum per `docker cp` hineinzukopieren, prüft den Baum des Zweigs, der das Tag zuletzt gebaut hat; ein Lauf gegen fremden Code meldet sich als leerer Lauf („no tests to run“) oder als fremdes Ergebnis. `tools/harness/mutationslauf.sh` kopiert den Baum und ist nicht betroffen.
