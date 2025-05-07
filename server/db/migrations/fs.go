package migrations

import (
	"embed"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
)

//go:embed *.sql
var fs embed.FS

func NewMigrations(rawURL string) (*dbmate.DB, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	db := dbmate.New(u)
	db.FS = fs
	db.AutoDumpSchema = false
	db.MigrationsDir = []string{"."}
	return db, nil
}
