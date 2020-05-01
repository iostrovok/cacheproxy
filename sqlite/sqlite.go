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

//
type Record struct {
	// Args is unique key: MD5 hash from url + request
	ID string `json:"id"`

	// See github.com/iostrovok/cacheproxy/store
	Body *store.Item `json:"body"`
}

type SQL struct {
	mx          sync.RWMutex
	fileName    string
	db          *sql.DB
	testCounter int
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

// execTx executes one command with transaction
func (s *SQL) execTx(command string, args ...interface{}) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	_, err = tx.Exec(command, args...)
	return err
}

// CreateTable just makes new table
func (s *SQL) CreateTable() error {
	return s.execTx("CREATE TABLE main (id TEXT, body BLOB, PRIMARY KEY(id))")
}

// Upsert
func (s *SQL) Upsert(id string, unit *store.Item) error {

	// don't update time if it's not necessary
	insertSQL := `INSERT INTO main(id, body) VALUES(?, ?)
		ON CONFLICT(id) DO UPDATE SET body=excluded.body WHERE excluded.id = main.id`

	body, err := unit.ToZip()
	if err == nil {
		err = s.execTx(insertSQL, id, body)
	}
	return err

}

func (s *SQL) Select(id string) (*store.Item, error) {
	s.mx.RLock()
	row := s.db.QueryRow(`SELECT body FROM main WHERE id = ?`, id)
	s.mx.RUnlock()

	body := make([]byte, 0)

	if err := row.Scan(&body); err != nil {
		return nil, err
	}

	return store.FromZip(body)
}

// SelectAll returns all rows sorted by id
func (s *SQL) SelectAll() ([]*Record, error) {
	s.mx.RLock()
	row, err := s.db.Query("SELECT id, body FROM main ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer row.Close()
	s.mx.RUnlock()

	out := make([]*Record, 0)
	for row.Next() {
		rec := Record{}
		body := make([]byte, 0)
		if err := row.Scan(&rec.ID, &body); err != nil {
			return nil, err
		}
		if rec.Body, err = store.FromZip(body, true); err != nil {
			return nil, err
		}
		out = append(out, &rec)
	}
	return out, nil
}

// SelectAllID returns all row id sorted by id
func (s *SQL) SelectAllID() ([]string, error) {
	s.mx.RLock()
	row, err := s.db.Query("SELECT id FROM main ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer row.Close()
	s.mx.RUnlock()

	out := make([]string, 0)
	for row.Next() {
		rec := ""
		if err := row.Scan(&rec); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

func (s *SQL) DeleteOld(requested map[string]bool) (int64, error) {

	total := int64(0)
	ids, err := s.SelectAllID()
	if err != nil {
		return total, err
	}

	delStmt, err := s.db.Prepare("DELETE from main WHERE id = ?")
	if err != nil {
		return total, err
	}

	for _, id := range ids {
		if requested[id] {
			continue
		}

		tx, err := s.db.Begin()
		if err != nil {
			return total, err
		}
		res, err := tx.Stmt(delStmt).Exec(id)
		if err != nil {
			return total, err
		}

		if err := tx.Commit(); err != nil {
			return total, err
		}

		count, err := res.RowsAffected()
		if err != nil {
			return total, err
		}
		total += count
	}

	return total, err
}
