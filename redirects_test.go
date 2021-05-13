package redirecter

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/ory/dockertest"
	"go.uber.org/zap"
)

func startPg(pgds *Pgds) (*sql.DB, func()) {
	var database = "my_db"
	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("postgres", "9.6", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=" + database})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	var db *sql.DB

	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", fmt.Sprintf("postgres://postgres:secret@localhost:%s/%s?sslmode=disable", resource.GetPort("5432/tcp"), database))
		if err != nil {
			return err
		}
		port, _ := strconv.Atoi(resource.GetPort("5432/tcp"))
		_, err = db.Exec("CREATE TABLE public.redirects (domain varchar(100), src_path varchar(350), dest_path varchar(350));")
		if err != nil {
			return err
		}
		pgds.Host = "localhost"
		pgds.Port = port
		pgds.User = "postgres"
		pgds.Password = "secret"
		pgds.DbName = database
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return db, func() {
		db.Close()

		// When you're done, kill and remove the container
		if err = pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
}

func TestRedirects(t *testing.T) {
	pgds := &Pgds{}

	db, stopPg := startPg(pgds)
	defer stopPg()

	tests := []struct {
		domain     string
		sourcePath string
		destPath   string
	}{
		{"vinissimus.com", "/blog/garnachas-de-culto/", "/es/garnacha"},
	}

	loader = Load

	for i, test := range tests {
		res, err := db.Exec(
			"INSERT INTO public.redirects (domain, src_path, dest_path) VALUES ($1, $2, $3)",
			test.domain, test.sourcePath, test.destPath,
		)
		if err != nil {
			panic(err)
		}
		res.LastInsertId()

		logger, _ := zap.NewDevelopment()
		redirecter := initRedirecter(*pgds, test.domain, logger)
		redirecter.Reload()

		if got, _ := redirecter.FindRedirect(test.sourcePath); got != test.destPath {
			t.Errorf("Test %v: Expected %s got %s", i, test.destPath, got)
		}
	}
}
