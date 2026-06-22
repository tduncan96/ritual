package db

import (
	"errors"
	"fmt"
)

type Host struct {
	HostId  int64  `db:"HostId"`
	Name    string `db:"Name"`
	Address string `db:"Address"`
	User    string `db:"User"`
	Port    int64  `db:"Port"`
	KeyPath string `db:"KeyPath"`
}

func GetHost(hostName string) (host Host, err error) {
	err = DB.Get(&host, "SELECT * FROM Hosts WHERE Name = ?", hostName)
	return host, err
}

func (h *Host) AddHost() (int64, error) {
	result, err := DB.NamedExec(
		`INSERT INTO Jobs (Name, Address, User, Port, KeyPath)
		VALUES (:Name, :Address, :User, :Port, :KeyPath)
		ON CONFLICT (Name) DO NOTHING`,
		h,
	)
	if err != nil {
		return 0, err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if num == 0 {
		var collision Host
		var errs []error
		if getErr := DB.Get(&collision, `SELECT HostId, Address FROM Hosts WHERE Name = ?`, h.Name); getErr != nil {
			errs = append(errs, getErr)
		}
		qErr := fmt.Errorf("collision with Host #%v at Address %v", collision.HostId, collision.Address)
		errs = append(errs, qErr)
		return 0, errors.Join(errs...)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (h *Host) UpdateHost() error {
	if _, err := DB.NamedExec(
		`UPDATE Hosts Set
			Name = :Name,
			Address = :Address,
			Port = :Port,
			KeyPath = :KeyPath
			Updated = datetime('now)`,
		h,
	); err != nil {
		return err
	}
	return nil
}
