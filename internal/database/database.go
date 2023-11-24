package database

import (
	"context"

	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePgxConf(cfg conf.Database) *pgxpool.Config {

	conf := &pgxpool.Config{}
	if cfg.Use {
		conf.ConnConfig.User = cfg.DbUser
		conf.ConnConfig.Password = cfg.DbPassword
		conf.ConnConfig.Host = cfg.DbAddress
		conf.ConnConfig.Port = uint16(cfg.DbPort)
		conf.ConnConfig.Database = cfg.DbName
		conf.MaxConns = int32(cfg.MaxConnections)
		conf.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	return conf

}

func CreateDbPool(ctx context.Context, conf *pgxpool.Config) (*pgxpool.Pool, error) {

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func ClosePool(pool *pgxpool.Pool) {
	pool.Close()
}
