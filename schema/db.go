package schema

import (
	"database/sql"
	_sql "database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type DB struct {
	db     *_sql.DB
	schema *Schema
}

func NewDB(db *sql.DB, schema *Schema) *DB {
	return &DB{
		db:     db,
		schema: schema,
	}
}

func (d *DB) Create() error {
	sql, err := d.schema.DDL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(sql)
	return err
}

func (d *DB) Reset() error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(`TRUNCATE %s`, d.schema.Name)
	l := log.WithField("sql", sql)
	_, err = tx.Exec(sql)
	if err != nil {
		tx.Rollback()
		l.WithError(err).Error()
		return err
	}
	versions := d.schema.versions()
	for _, version := range versions {
		sql := fmt.Sprintf(`TRUNCATE %s_%s`, d.schema.Name, version)
		l := l.WithField("sql", sql)
		_, err = tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			l.WithError(err).Error()
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) Upsert(doc map[string]interface{}) error {
	err := d.schema.validate(doc)
	if err != nil {
		return err
	}
	sql, values, err := d.schema.Get(doc)
	l := log.WithField("sql", sql).WithField("values", values)
	r := d.db.QueryRow(sql, values...)
	l.Info()
	var id int
	err = r.Scan(&id)
	if err != nil && err != _sql.ErrNoRows {
		l.WithError(err).Error()
		return err
	}
	tx, err2 := d.db.Begin()
	if err2 != nil {
		l.WithError(err2).Error()
		return err2
	}
	versions := d.schema.versions()
	if err == _sql.ErrNoRows { // INSERT
		sql, values, err := d.schema.Insert(doc)
		l = l.WithField("sql", sql)
		if err != nil {
			tx.Rollback()
			l.WithError(err).Error()
			return err
		}
		r = tx.QueryRow(sql, values...)
		err = r.Scan(&id)
		if err != nil {
			tx.Rollback()
			l.WithError(err).Error()
			return err
		}
	} else { // UPDATE
		for _, version := range versions {
			sql := fmt.Sprintf(`DELETE FROM %s_%s WHERE %s=$1;`, d.schema.Name, version, d.schema.Name)
			_, err := tx.Exec(sql, id)
			if err != nil {
				tx.Rollback()
				l.WithError(err).Error()
				return err
			}
		}
		sql, values, err := d.schema.Update(doc)
		if err != nil {
			tx.Rollback()
			l.WithError(err).Error()
			return err
		}
		l = l.WithField("sql", sql)
		_, err = tx.Exec(sql, values...)
		if err != nil {
			tx.Rollback()
			l.WithError(err).Error()
			return err
		}
	}
	for _, version := range versions {
		sql := fmt.Sprintf(`INSERT INTO %s_%s (%s, name, version)
			VALUES ($1, $2, $3);`, d.schema.Name, version, d.schema.Name)
		l = l.WithField("sql", sql)
		ver, ok := doc[version]
		if ok {
			for k, v := range ver.(map[string]interface{}) {
				_, err = tx.Exec(sql, id, k, v)
				if err != nil {
					tx.Rollback()
					l.WithError(err).Error()
					return err
				}
			}
		}
	}
	return tx.Commit()
}