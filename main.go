package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/jimsmart/schema"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var DB *sql.DB
var _logger *log.Logger

const dbSource string = "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable"

var (
	port       = flag.String("server_port", "8080", "port used for http server")
	dbUser     = flag.String("db_user", "", "database user")
	dbPassword = flag.String("db_password", "", "database password")
	dbHost     = flag.String("db_host", "", "database host")
	dbPort     = flag.Int("db_port", 5432, "database port")
	dbName     = flag.String("db_name", "", "database name")
	dbDriver   = flag.String("db_driver", "postgres", "database driver")
)

func main() {
	logFile, err := os.OpenFile("./error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	_logger = log.New(logFile, "Error Logger:\t", log.Ldate|log.Ltime|log.Lshortfile)
	defer func() {
		err := logFile.Close()
		if err != nil {
			_logger.Println(err)
		}
	}()

	flag.Parse()
	source := fmt.Sprintf(dbSource, *dbUser, *dbPassword, *dbHost, *dbPort, *dbName)
	storage, err := NewJsonStorage()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(*dbDriver, source)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}

	DB = db
	fmt.Println("You connected to your database: ", *dbName)
	err = setupDBMetaData(storage)
	if err != nil {
		_logger.Fatalln(err)
	}
}

func setupDBMetaData(storage Repository) error {
	if exists, err := storage.IsDBAdded(*dbName); err != nil {
		return err
	} else if !exists {
		data := dbInfo{
			Name:     *dbName,
			User:     *dbUser,
			Host:     *dbHost,
			Port:     *dbPort,
			Password: *dbPassword,
			Driver:   *dbDriver,
		}
		err := storage.AddDB(data)
		if err != nil {
			return err
		}

		tableNames, err := schema.TableNames(DB)
		if err != nil {
			return err
		}

		for i := range tableNames {
			tableColumns, err := schema.Table(DB, tableNames[i])
			if err != nil {
				return err
			}
			for _, col := range tableColumns {
				meta := colMetaData{}
				meta.Name = col.Name()
				meta.DBType = col.DatabaseTypeName()
				meta.Nullable = parseNullableFromCol(col)
				meta.GoType = col.ScanType().String()
				meta.Length = parseLengthFromCol(col)
				meta.TBName = tableNames[i]

				t := table{Name: tableNames[i]}
				err = storage.AddTable(t)
				if err != nil {
					return err
				}

				err = storage.AddColMetaData(tableNames[i], meta)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type Repository interface {
	AddDB(dbInfo) error
	AddTable(table) error
	AddColMetaData(tbName string, col colMetaData) error
	IsDBAdded(dbName string) (bool, error)
	GetTables() (Tables, error)
}

type dbInfo struct {
	Name     string `json:"name"`
	User     string `json:"user"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Driver   string `json:"driver"`
}

type colMetaData struct {
	Name         string `json:"name"`
	DBType       string `json:"db_type"`
	Nullable     bool   `json:"nullable"`
	GoType       string `json:"go_type"`
	Length       int64  `json:"length"`
	TBName       string `json:"table_name"`
	Description  string `json:"description"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsForeignKey bool   `json:"is_foreign_key"`
}

type table struct {
	Name        string `json:""`
	Description string `json:"description"`
}

type Tables []table

func (t Tables) Count() int {
	return len(t)
}

func parseNullableFromCol(col *sql.ColumnType) bool {
	if isNullable, ok := col.Nullable(); !ok {
		return false
	} else {
		return isNullable
	}
}

func parseLengthFromCol(col *sql.ColumnType) int64 {
	if length, ok := col.Length(); !ok {
		return 0
	} else {
		return length
	}
}
