package lib

import "database/sql"

func PerformInserts(dbh *sql.DB, inserts []*Insert) int {
	ins_count := 0
	query := `INSERT INTO dictionary (word, lexical_category, etymology_no, definition_no, definition)
              VALUES (?, ?, ?, ?, ?)`

	logger.Debug("performInserts> Preparing insert query...\n")
	tx, err := dbh.Begin()
	Check(err)
	defer tx.Rollback()

	sth, err := tx.Prepare(query)
	Check(err)
	defer sth.Close()

	for _, ins := range inserts {
		logger.Debug("performInserts> et_no=>'%d' defs=>'%+v'\n", ins.Etymology, ins.CatDefs)
		for key, val := range ins.CatDefs {
			category := key
			for def_no, def := range val {
				logger.Debug("performInserts> Inserting values: word=>'%s', lexical category=>'%s', et_no=>'%d', def_no=>'%d', def=>'%s'\n",
					ins.Word, category, ins.Etymology, def_no, def)
				_, err := sth.Exec(ins.Word, category, ins.Etymology, def_no, def)
				Check(err)
				ins_count++
			}
		}
	}

	err = tx.Commit()
	Check(err)

	return ins_count
}
