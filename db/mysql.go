package db

import (
        "context"
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

type SqlExec struct {
        sqlx.DB
        ctx context.Context
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

func GetSqlExec(ctx context.Context, databaseName string) (*SqlExec, error) {
        if db, exist := defaultDbCatch.get(databaseName); exist {
                sqlExec := &SqlExec{
                        *db,
                        ctx,
                }
                return sqlExec, nil
        }
        return nil, errors.Errorf("database %s not found.", databaseName)
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

        err = db.Ping()
        if err != nil {
                return errors.WithStack(err)
        }
        defaultDbCatch.add(databaseName, db)
        return nil
}