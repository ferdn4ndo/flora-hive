package repositories

import (
	"go.uber.org/fx"

	"flora-hive/internal/domain/ports"
	"flora-hive/lib"
)

// Module wires repository constructors.
var Module = fx.Options(
	fx.Provide(NewDatabaseHandler),
)

// NewRepositories builds transactional repositories for a QueryAble (DB or Tx).
func NewRepositories(q lib.QueryAble) *ports.Repositories {
	return &ports.Repositories{
		HiveUser:    NewHiveUserRepo(q),
		Environment: NewEnvironmentRepo(q),
		Device:      NewDeviceRepo(q),
	}
}

// DatabaseHandler implements ports.DatabaseHandler.
type DatabaseHandler struct {
	db lib.Database
}

// NewDatabaseHandler creates the handler.
func NewDatabaseHandler(db lib.Database) ports.DatabaseHandler {
	return &DatabaseHandler{db: db}
}

// WithTrx runs cb inside a transaction.
func (d *DatabaseHandler) WithTrx(cb func(r *ports.Repositories) error) error {
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}
	q := NewRepositories(tx)
	err = cb(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}
	return tx.Commit()
}
