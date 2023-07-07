package db

import (
	"context"
	"fmt"
	"seismo/collector/db/mongodb"
	"seismo/collector/db/stubdb"
	"seismo/provider"
	"time"
)

// // NoDataErr that no data has been fetched,
// // e.g., query returns empty value.
// type NoDataErr struct {
// }

// func (e NoDataErr) Error() string {
// 	return "No Data"
// }

type DbType string

const (
	StubDb  DbType = "StubDb"
	MongoDb DbType = "MongoDb"
	//PostreSQL = "PostgreSQL"

	defDbType  DbType = StubDb
	defConnStr string = ""
)

type DbConfig struct {
	T       DbType
	ConnStr string
}

func DefaultDbConfig() DbConfig {
	return DbConfig{T: defDbType, ConnStr: defConnStr}
}

type Adapter interface {
	Connect(ctx context.Context, connStr string) error
	Close(ctx context.Context) error
	SaveMsg(ctx context.Context, msgs []provider.Message) error
	GetLastTime(ctx context.Context, sorceId string) (time.Time, error)
}

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
