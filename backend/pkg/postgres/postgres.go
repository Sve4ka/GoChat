package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/pkg/config"
	"backend/pkg/log"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // Register postgres driver.
	"github.com/pressly/goose/v3"
)

type Pg struct {
	maxPoolSize  int32
	connAttempts int
	connTimeout  time.Duration
	Pool         *pgxpool.Pool
}

func MustInitPg(cfg *config.Config) *Pg {
	connString := fmt.Sprintf("user=%v host=%v port=%v dbname=%v password=%v sslmode=disable",
		cfg.PGUser,
		cfg.PGHost,
		cfg.PGPort,
		cfg.PGName,
		cfg.PGPassword,
	)

	pg := &Pg{
		maxPoolSize:  cfg.MaxPool,
		connAttempts: cfg.ConnAttempts,
		connTimeout:  cfg.PGTimeout,
	}

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		panic(fmt.Sprintf("error parsing config: %v", err.Error()))
	}

	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			pingErr := pg.Pool.Ping(context.Background())
			if pingErr == nil {
				break
			}

			err = pingErr
		}

		log.Log.Info(fmt.Sprintf("Pg is trying to connect, attempts left: %d", pg.connAttempts))

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		panic(fmt.Sprintf("error while connecting to db: %v", err.Error()))
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Log.Error(err)
	}

	if err = goose.Up(db, "../deploy/migration"); err != nil {
		log.Log.Error(err)
	}

	defer db.Close()

	return pg
}

func (p *Pg) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
