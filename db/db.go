package db

import (
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var (
	defaultDbCatch = &dbCatch{dbs: map[string]*sqlx.DB{}}
)

type MySQLConfig struct {
	ConnStr      string
	MaxOpenConns int
	MaxIdleConns int
}

type dbCatch struct {
	mux sync.RWMutex
	dbs map[string]*sqlx.DB
}

func (catch *dbCatch) add(databaseName string, connPool *sqlx.DB) {
	catch.mux.Lock()
	defer catch.mux.Unlock()

	if catch.dbs == nil {
		catch.dbs = make(map[string]*sqlx.DB)
	}

	catch.dbs[databaseName] = connPool
}

func (catch *dbCatch) get(databaseName string) (*sqlx.DB, bool) {
	catch.mux.RLock()
	defer catch.mux.RUnlock()
	db, exist := catch.dbs[databaseName]
	return db, exist
}

func GetSqlExec(databaseName string) (*sqlx.DB, error) {
	if db, exist := defaultDbCatch.get(databaseName); exist {
		return db, nil
	}
	return nil, errors.New(fmt.Sprintf("database %s not found.", databaseName))
}

func RigisterDB(databaseName string, config *MySQLConfig) error {
	if config.ConnStr == "" {
		return errors.New("init mysql database instance failed. data source name is empty!")
	}

	db, err := sqlx.Connect("mysql", config.ConnStr)
	if err != nil {
		return errors.WithStack(err)
	}

	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}

	defaultDbCatch.add(databaseName, db)
	return nil
}
