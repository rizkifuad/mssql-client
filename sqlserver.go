package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

type QueryResult struct {
	Cols    []string   `json:"cols"`
	Rows    [][]string `json:"rows"`
	Elapsed string     `json:"elapsed"`
}

type Model struct {
	ID uint `gorm:"primary_key" json:"id"`
}

type Connection struct {
	Model
	db        *sql.DB          `gorm:"-"`
	Database  string           `json:"database"`
	Server    string           `json:"server"`
	User      string           `json:"user"`
	Password  string           `json:"password"`
	Port      int              `json:"port"`
	Encrypt   bool             `json:"encrypt"`
	Name      string           `json:"name"`
	Databases []ActiveDatabase `json:"databases"  schema:"-"`
}

type ActiveDatabase struct {
	Name         string `json:"name"`
	ConnectionID uint   `json:"connection_id" json:"-"`
}

func (this *Connection) Connect() error {
	enc := "disable"
	if this.Encrypt {
		enc = "enable"
	}
	connectionString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;encrypt=%s;database=%s",
		this.Server, this.User, this.Password, this.Port, enc, this.Database)

	db, err := sql.Open("mssql", connectionString)

	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT name FROM master.dbo.sysdatabases")
	if err != nil {
		return err
	}

	defer rows.Close()

	storage.Exec("DELETE FROM active_databases where connection_id=?;", this.ID)
	var databases []ActiveDatabase
	for rows.Next() {
		var database string
		err = rows.Scan(&database)
		if err != nil {
			log.Println(err.Error())
			break
		}

		activeDb := ActiveDatabase{
			Name:         database,
			ConnectionID: this.ID,
		}

		storage.Create(activeDb)
		databases = append(databases, activeDb)
	}

	this.Databases = databases
	this.db = db
	return nil

}

func (this *Connection) Disconnect() {
	this.db.Close()
}

func (this *Connection) ListDatabases() ([]string, error) {
	var databases []string

	rows, err := this.db.Query("SELECT name FROM master.dbo.sysdatabases")
	if err != nil {
		return databases, err
	}
	defer rows.Close()

	for rows.Next() {
		var database string
		err = rows.Scan(&database)

		databases = append(databases, database)
	}

	return databases, nil
}

func (this *Connection) Query(query string) (QueryResult, error) {
	if query == "" {
		return QueryResult{}, errors.New("Empty query")
	}

	start := time.Now()
	rows, err := this.db.Query(query)

	if err != nil {
		return QueryResult{}, err
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return QueryResult{}, err
	}

	if cols == nil {
		return QueryResult{}, err
	}

	var vals [][]string

	count := 0
	elapsed := time.Since(start).String()
	for rows.Next() {
		var tempVals = make([]interface{}, len(cols))
		for i := 0; i < len(cols); i++ {
			tempVals[i] = new(interface{})
		}

		err = rows.Scan(tempVals...)
		if err != nil {
			log.Printf("Scan failed: %s", err.Error())
			continue
		}

		val := make([]string, len(cols))
		for id, data := range tempVals {
			val[id] = ParseInterface(data.(*interface{}))
		}
		vals = append(vals, val)

		count++
	}

	return QueryResult{
		Cols:    cols,
		Rows:    vals,
		Elapsed: elapsed,
	}, nil
}
