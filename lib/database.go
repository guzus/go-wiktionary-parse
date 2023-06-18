package lib

import (
	"database/sql"
	"fmt"
)

func Init(db string) (dbh *sql.DB, err error) {
	logger.Info("Opening database\n")
	dbh, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&mode=rwc&_mutex=full&_busy_timeout=500", db))
	if err != nil {
		return
	}

	dbh.SetMaxOpenConns(1)

	sth, err := dbh.Prepare(`CREATE TABLE IF NOT EXISTS dictionary
                             (
                                 id INTEGER PRIMARY KEY,
                                 word TEXT,
                                 lexical_category TEXT,
                                 etymology_no INTEGER,
                                 definition_no INTEGER,
                                 definition TEXT
                             )`)
	if err != nil {
		return
	}
	sth.Exec()

	sth, err = dbh.Prepare(`CREATE INDEX IF NOT EXISTS dict_word_idx
                            ON dictionary (word, lexical_category, etymology_no, definition_no)`)
	if err != nil {
		return
	}
	sth.Exec()
	return
}

func PerformInserts(dbh *sql.DB, inserts []*Insert) int {
	insCount := 0
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
				insCount++
			}
		}
	}

	err = tx.Commit()
	Check(err)

	return insCount
}
