# Ein Fänger deckt die Größen nur bis zur Speichergrenze des Containers, und die Lücke steht nur als Negativ im Plan

**Sub-Area:** `*` (gesamtes Repo, Kürzel `REPO`)

Eine Zusage über den Speicher je gelesenes Byte oder über das Verhalten bei Größen bis zur Grenze (`MAX_CAPTURE_BODY`, Standard 8 MiB) wird nur bis zur Größe geprüft, die der Container mit `--memory 2g` trägt. Der Detektor-Lauf (`make test-race`) führt die Stufen über 1 MiB nicht, und die Messung der Zuteilung über den Stufen der Wachstumsfolge deckt nur eine Form des Bodys. Das Negativ hat einen Grund (Speicher), aber die Folge für den Detektor-Lauf und die ungemessenen Formen stehen nur implizit da.
