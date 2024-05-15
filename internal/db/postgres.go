package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	pg "github.com/habx/pg-commands"
)

// Postgres is a wrapper around pg.postgres
// ENV variables:
// DB_HOST
// DB_DATABASE
// DB_USER
// DB_PASSWORD
// DB_PORT

type path string

type Postgres struct {
	pg.Postgres
}

func NewPostgresFromEnv() *Postgres {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	db := os.Getenv("DB_DATABASE")
	if db == "" {
		db = "postgres"
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	port_s := os.Getenv("DB_PORT")
	if port_s == "" {
		port_s = "5432"
	}
	port, err := strconv.Atoi(port_s)
	if err != nil {
		panic(err)
	}
	return NewPostgres(host, port, db, user, password)
}

// postgres://username:password@localhost:32781/database?sslmode=disable
func NewPostgresFromConnString(connString string) (*Postgres, error) {
	var err error = nil
	sections := strings.Split(strings.Split(connString, "://")[1], "@")
	username := strings.Split(sections[0], ":")[0]
	password := strings.Split(sections[0], ":")[1]
	host := strings.Split(sections[1], ":")[0]
	portStr := strings.Split(strings.Split(sections[1], ":")[1], "/")[0]
	db := strings.Split(strings.Split(sections[1], "/")[1], "?")[0]

	port, err := strconv.Atoi(portStr)
	return NewPostgres(host, port, db, username, password), err
}

func NewPostgres(host string, port int, db, user, password string) *Postgres {
	postgres := pg.Postgres{
		Host:     host,
		Port:     port,
		DB:       db,
		Username: user,
		Password: password,
	}
	return &Postgres{postgres}
}

func (db *Postgres) GetUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", db.Username, db.Password, db.Host, db.Port, db.DB)
}

func (db *Postgres) TestConnection() error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, db.GetUrl())
	if err != nil {
		return err
	}
	conn.Close(ctx)
	return nil
}

func (db *Postgres) Dump(format string) pg.Result {
	if format == "" {
		format = "t"
	}
	if format != "t" && format != "p" {
		log.Fatal("Format must be t or p")
	}
	now := time.Now().Format("2006-01-02T15:04:05")
	filename := "dump_" + now + "." + ext(format)
	dump, err := pg.NewDump(&db.Postgres)
	if err != nil {
		panic(err)
	}
	dump.SetFileName(filename)
	dump.SetupFormat(format)
	dumpExec := dump.Exec(pg.ExecOptions{StreamPrint: false})
	if dumpExec.Error != nil {
		fmt.Println(dumpExec.Error.Err)
		fmt.Println(dumpExec.Output)
	}
	fmt.Println(dumpExec.Output)
	fmt.Println("Dump success")
	return dumpExec
}

func (db *Postgres) Restore(path string) error {
	restore, err := pg.NewRestore(&db.Postgres)
	if err != nil {
		panic(err)
	}

	restore.Options = append(restore.Options, "-Ft")

	restoreExec := restore.Exec(path, pg.ExecOptions{StreamPrint: false})
	if restoreExec.Error != nil {
		fmt.Println(restoreExec.Error.Err)
		fmt.Println(restoreExec.Output)
		return restoreExec.Error.Err
	}
	fmt.Println("Restore success")
	fmt.Println(restoreExec.Output)
	return nil
}

func ext(format string) string {
	if format == "p" {
		return "sql"
	}
	return "tar"
}