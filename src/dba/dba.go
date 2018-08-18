package dba

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

type Rep struct {
	AuthRepId string `json:"authRepId"`
	RepId     uint64
}

func (r *Rep) String() string {
	return fmt.Sprintf("authRepId: %s      repId: %d", r.AuthRepId, r.RepId)
}

func GetRep(authRepId string) (*Rep, error) {
	db, err := sql.Open("mysql", "root:@/asapp_dev_companies1")
	defer db.Close()
	if err != nil {
		return nil, errors.Wrap(err, "creating database connection")
	}

	var rep Rep
	// QueryRow selects 1 row
	row := db.QueryRow("SELECT AuthRepId, RepId FROM Rep WHERE AuthRepId = ?", authRepId)

	err = row.Scan(&rep.AuthRepId, &rep.RepId)
	if err != nil {
		return nil, errors.Wrapf(err, "selecting rep %s from database", authRepId)
	}

	return &rep, nil
}

func CreateRep(authRepId string) (int64, error) {
	db, err := sql.Open("mysql", "root:@/asapp_dev_companies1")
	defer db.Close()
	if err != nil {
		return 0, errors.Wrap(err, "creating database connection")
	}

	query := `
		INSERT INTO Rep(
			CreatedTime,
			CRMRepId,
			CompanyId,
			AuthRepId,
			Name
		)
		VALUES(
			1534566746000000,
			?,
			1,
			?,
			'Tom'
		)
	`
	results, err := db.Exec(query, authRepId, authRepId)
	if err != nil {
		return 0, errors.Wrap(err, "inserting rep")
	}

	repId, err := results.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "getting inserted repid")
	}

	return repId, nil
}
