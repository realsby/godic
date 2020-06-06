package main

import "errors"

var ErrNoDatabaseMetaDataStored = errors.New("there is no database metadata stored in repository")

type Repository interface {
	Setup
	UpdateAddTableDescription(tableID string, description string) error
	UpdateAddColumnDescription(columnID string, description string) error
	GetTables() (Tables, error)
	GetDatabaseInfo() (databaseInfo, error)
}

type Setup interface {
	AddDatabaseInfo(databaseInfo) error
	AddTable(table) error
	AddColMetaData(tableName string, col colMetaData) error
	RemoveEverything() error
	IsDatabaseMetaDataAdded(databaseName string) (bool, error)
}
