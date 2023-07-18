// package seismo/collector/db contains basic types for
// the Collector to interact with a database.
package db

import (
	"context"
	"fmt"
	"seismo/collector/db/mongodb"
	"seismo/collector/db/stubdb"
	"seismo/provider"
	"time"
)

// DbType represents various DBMS (e.g., MongoDb, PostreSQL etc),
// and, therefore, their providers.
type DbType string

const (
	StubDb    DbType = "StubDb"
	MongoDb   DbType = "MongoDb"
	PostreSQL DbType = "PostgreSQL"

	defDbType  DbType = StubDb
	defConnStr string = ""
)

// DbConfig represents collector database configuration.
type DbConfig struct {
	//T specifies DBMS.
	T DbType

	//ConnStr specifies a connection string.
	ConnStr string
}

// DefaultDbConfig creates and returns an instance of DbConfig with default values.
func DefaultDbConfig() DbConfig {
	return DbConfig{T: defDbType, ConnStr: defConnStr}
}

// Adapter is implemented to provide the interaction of the Collector with a database.
type Adapter interface {
	//Connect opens a new connection to a database using "connStr" as the connection string.
	Connect(ctx context.Context, connStr string) error
	//Close closes the opened connection.
	Close(ctx context.Context) error
	//SaveMsg saves messages in the connected database.
	SaveMsg(ctx context.Context, msgs []provider.Message) error
	//GetLastTime returns the focus time of the last saved message for specified "sourceId".
	GetLastTime(ctx context.Context, sorceId string) (time.Time, error)
}

// NewAdapter creats a new Adapter implementation depending on a specified in "config" database type.
func NewAdapter(conf DbConfig) (Adapter, error) {
	switch conf.T {
	case StubDb:
		return &stubdb.Adapter{}, nil
	case MongoDb:
		return &mongodb.Adapter{}, nil
	default:
		return nil, fmt.Errorf("NewAdapter: unknown data base type: %q", conf.T)
	}
}
