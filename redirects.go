package redirecter

import (
	"database/sql"
	"fmt"
	"sync/atomic"

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
	urlMap atomic.Value
	logger *zap.Logger
}

var loader = Load

func initRedirecter(pg Pgds, logger *zap.Logger) *Redirecter {
	redirecter := Redirecter{
		Pgds:   pg,
		logger: logger,
	}
	return &redirecter
}

func (r *Redirecter) Reload() error {
	urlMap, err := loader(r)
	if err != nil {
		return err
	}
	r.urlMap.Store(urlMap)
	return nil
}

func (r *Redirecter) FindRedirect(path string) (string, bool) {
	urlMap := (r.urlMap.Load()).(map[string]string)
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
	rows, err := db.Query("SELECT src_url, dst_path FROM public.redirects")
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
	r.logger.Info(fmt.Sprintf("Loaded %d urls", len(newUrlMap)))
	return newUrlMap, nil
}
