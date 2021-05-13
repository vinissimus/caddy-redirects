package redirecter

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type Pgds struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbname"`
}

type Redirecter struct {
	Pgds
	domain string
	urlMap *map[string]string
	logger *zap.Logger
}

var loader = Load

func initRedirecter(pg Pgds, domain string, logger *zap.Logger) *Redirecter {
	redirecter := Redirecter{
		Pgds:   pg,
		domain: domain,
		logger: logger,
	}
	logger.Info(fmt.Sprintf("initRedirecter() for domain %s\n", domain))
	return &redirecter
}

func (r *Redirecter) Reload() {
	urlMap, err := loader(r)
	if err != nil {
		panic(err)
	}
	r.urlMap = &urlMap
}

func (r *Redirecter) FindRedirect(path string) (string, bool) {
	urlMap := *r.urlMap
	val, ok := urlMap[path]
	return val, ok
}

func Load(r *Redirecter) (map[string]string, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		r.Host, r.Port, r.User, r.Password, r.DbName,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	newUrlMap := make(map[string]string)
	rows, err := db.Query("SELECT src_path, dest_path FROM public.redirects WHERE domain = $1", r.domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sourcePath string
		var destPath string
		err = rows.Scan(&sourcePath, &destPath)
		if err != nil {
			return nil, err
		}
		newUrlMap[sourcePath] = destPath
	}
	r.logger.Info(fmt.Sprintf("Loaded %d urls for domain %s", len(newUrlMap), r.domain))
	return newUrlMap, nil
}
