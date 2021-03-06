package dbr

import (
	"database/sql"

	"github.com/corestoreio/errors"
	"github.com/corestoreio/log"
)

// Tx at a transaction for the given Session
type Tx struct {
	log.Logger
	*sql.Tx
}

// Begin creates a transaction for the given session
func (c *Connection) Begin() (*Tx, error) {
	dbTx, err := c.DB.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "[dbr] transaction.begin.error")
	}
	tx := &Tx{
		Tx: dbTx,
	}
	if c.Log != nil {
		tx.Logger = c.Log.With(log.Bool("transaction", true))
	}
	return tx, nil
}

// Commit finishes the transaction
func (tx *Tx) Commit() error {
	return errors.Wrap(tx.Tx.Commit(), "[dbr] transaction.commit.error")
}

// Rollback cancels the transaction
func (tx *Tx) Rollback() error {
	return errors.Wrap(tx.Tx.Rollback(), "[dbr] transaction.rollback.error")
}

// RollbackUnlessCommitted rolls back the transaction unless it has already been
// committed or rolled back. Useful to defer tx.RollbackUnlessCommitted() -- so
// you don't have to handle N failure cases Keep in mind the only way to detect
// an error on the rollback at via the event log.
func (tx *Tx) RollbackUnlessCommitted() {
	err := tx.Tx.Rollback()
	if err == sql.ErrTxDone {
		// ok
	} else if err != nil {
		//tx.EventErr("dbr.rollback_unless_committed", err)
		panic(err) // todo remove panic
	}
}
