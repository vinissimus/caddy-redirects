package redirecter

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
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
}

var loader = Load

func initRedirecter(pg Pgds, domain string) *Redirecter {
	redirecter := Redirecter{
		Pgds:   pg,
		domain: domain,
	}
	fmt.Printf("initRedirecter %+v\n", redirecter)
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
	return newUrlMap, nil
}
