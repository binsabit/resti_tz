package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/binsabit/resti_tz/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"log"
	"net/url"
	"time"
)

var QueryBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Database struct {
	db *pgxpool.Pool
}

func (d *Database) GetDb() DBTX {
	return d.db
}

func (d *Database) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := d.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return nil, err
	}
	txObj := tx
	return txObj, nil
}

func NewDatabase(ctx context.Context, cfg config.Database) *Database {

	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}

	q := dsn.Query()

	q.Add("sslmode", "disable")
	//q.Add("timezone", "Asia/Almaty")

	dsn.RawQuery = q.Encode()
	poolConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		log.Fatal(err)
	}

	poolConfig.MaxConns = 15
	poolConfig.MaxConnIdleTime = time.Minute * 10

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := pgxPool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	err = NewMigrator(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &Database{
		db: pgxPool,
	}
}

func NewMigrator(cfg config.Database) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	log.Println(dsn)
	// Open connection to the database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		log.Printf("cannot connect to %s \n error: $v", dsn, err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	migrator, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", "migrations"), cfg.Name, driver)
	if err != nil {
		return err
	}
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
