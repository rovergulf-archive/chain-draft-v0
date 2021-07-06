package database

import "errors"

var (
	ErrGenesisNotExists = errors.New("genesis does not exists")
	ErrBlockNotExists   = errors.New("block does not exists")
	ErrTxNotExists      = errors.New("transaction does not exists")
)
