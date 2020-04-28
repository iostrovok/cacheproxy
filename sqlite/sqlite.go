package sqlite

import (
	"database/sql"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/iostrovok/cacheproxy/store"
)

var globalConnMutex sync.RWMutex

var testCounter = 0

func init() {
	globalConnMutex = sync.RWMutex{}
}

type SQL struct {
	mx          sync.RWMutex
	fileName    string
	db          *sql.DB
	testCounter int
	timeForCut  int
}

// Exists reports whether the named file or directory exists.
func existsFile(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func Conn(fileName string) (*SQL, error) {
	globalConnMutex.Lock()
	defer globalConnMutex.Unlock()

	return conn(fileName)
}

func conn(fileName string) (*SQL, error) {

	testCounter++
	c := &SQL{fileName: fileName, testCounter: testCounter}

	if !existsFile(fileName) {
		file, err := os.Create(fileName) // Create SQLite file
		if err != nil {
			return nil, err
		}
		file.Close()

		if err = c.Open(); err == nil {
			err = c.CreateTable()
		}

		return c, err
	}

	err := c.Open()
	return c, err
}

func (s *SQL) Close() error {
	var err error
	if s.db != nil {
		err = s.db.Close()
		s.db = nil
	}
	return err
}

func (s *SQL) Open() error {
	db, err := sql.Open("sqlite3", s.fileName) // Open the SQLite File
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

// execTx executes one command without transaction
func (s *SQL) exec(command string, args ...interface{}) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	statement, err := s.db.Prepare(command)
	if err == nil {
		_, err = statement.Exec(args...)
	}

	return err
}

// execTx executes one command with transaction
func (s *SQL) execTx(command string, args ...interface{}) (sql.Result, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Commit()

	return tx.Exec(command, args...)
}

func (s *SQL) CreateTable() error {
	createTableSQL := "CREATE TABLE main (args TEXT, last_date INTEGER, body BLOB, PRIMARY KEY(args))"
	dropTableSQL := "DROP TABLE main"
	indexSQL := "CREATE INDEX idx_last_date ON main (last_date)"

	if _, err := s.execTx(createTableSQL); err != nil {
		return err
	}

	if _, err := s.execTx(indexSQL); err != nil {
		s.execTx(dropTableSQL) // ignore error
		return err
	}

	return nil
}

// We are passing db reference connection from main to our method with other parameters
func (s *SQL) Upsert(args string, unit *store.StoreUnit) error {
	insertSQL := `INSERT INTO main(args, body, last_date)
  VALUES(?, ?, strftime('%s','now'))
  ON CONFLICT(args) DO 
  UPDATE SET body=excluded.body, last_date = strftime('%s','now') WHERE excluded.args = main.args`

	body, err := unit.ToZip()
	if err != nil {
		return err
	}
	_, err = s.execTx(insertSQL, args, body)
	return err
}

func (s *SQL) Select(args string) (*store.StoreUnit, error) {
	s.mx.RLock()
	row := s.db.QueryRow(`SELECT body FROM main WHERE args = ?`, args)
	s.mx.RUnlock()

	body := make([]byte, 0)

	if err := row.Scan(&body); err != nil {
		return nil, err
	}

	if len(body) > 0 {
		if err := s.setTimeRead(args); err != nil {
			return nil, err
		}
	}

	return store.FromZip(body)
}

func (s *SQL) setTimeRead(args string) error {
	setTimeSQL := `UPDATE main SET last_date = strftime('%s','now') WHERE args = ?`
	_, err := s.execTx(setTimeSQL, args)
	return err
}

func (s *SQL) DeleteOld() (int64, error) {
	count := int64(0)
	res, err := s.execTx("DELETE from main WHERE last_date < ?", s.timeForCut)
	if err == nil {
		count, err = res.RowsAffected()
	}
	return count, err
}

func (s *SQL) fixTimeForCut() error {
	s.mx.RLock()
	row := s.db.QueryRow(`SELECT strftime('%s','now') as t`)
	s.mx.RUnlock()

	return row.Scan(&s.timeForCut)
}
