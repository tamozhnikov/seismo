package db

import (
	"context"
	"fmt"
	"seismo"
	"seismo/collector/db/stubdb"
	"time"
)

type DbType string

const (
	StubDb DbType = "StubDb"
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
	SaveMsg(ctx context.Context, msgs []seismo.Message) error
	GetLastTime(ctx context.Context, sorceId string) (time.Time, error)
}

func NewAdapter(conf DbConfig) (Adapter, error) {
	switch conf.T {
	case StubDb:
		return &stubdb.Adapter{}, nil
	default:
		return nil, fmt.Errorf("unknown data base type: %q", conf.T)
	}
}
