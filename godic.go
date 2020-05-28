package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/jimsmart/schema"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
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
	dbSchema   = flag.String("db_schema", "public", "database schema for search_path")
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
	_, err = db.Exec(fmt.Sprintf("SET search_path=%s", *dbSchema))
	if err != nil {
		panic(err)
	}

	DB = db
	fmt.Println("You connected to your database: ", *dbName)
	err = setupDBMetaData(storage)
	if err != nil {
		_logger.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", index())
	srv := http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		_logger.Println(err)
	}
}

func index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sb, err := Asset("index.html")
		if err != nil {
			_logger.Println(err)
			http.Error(w, fmt.Sprintf("Template error: %s", err), http.StatusInternalServerError)
			return
		}
		tpl, err := template.New("").Parse(string(sb))
		if err != nil {
			_logger.Println(err)
			http.Error(w, fmt.Sprintf("Template error: %s", err), http.StatusInternalServerError)
			return
		}
		err = tpl.Execute(w, nil)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		}
	}
}

// setupDBMetadata stores in repository all the DB metadata.
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

		primaryKeys, err := getPrimaryKeys()
		if err != nil {
			return err
		}

		foreignKeys, err := getForeignKeys()
		if err != nil {
			return err
		}

		enums, err := getColsAndEnums()
		if err != nil {
			return err
		}

		uniques, err := getUniqueCols()
		if err != nil {
			return err
		}

		for i := range tableNames {
			tableColumns, err := schema.Table(DB, tableNames[i])
			if err != nil {
				return err
			}
			for _, col := range tableColumns {
				colMeta := colMetaData{}
				colMeta.Name = col.Name()
				colMeta.DBType = col.DatabaseTypeName()
				colMeta.Nullable = parseNullableFromCol(col)
				colMeta.GoType = col.ScanType().String()
				colMeta.Length = parseLengthFromCol(col)
				colMeta.TBName = tableNames[i]

				if isPK := primaryKeys.exists(colMeta.Name); isPK {
					colMeta.IsPrimaryKey = true
				}

				if isFK := foreignKeys.exists(colMeta.Name); isFK {
					fk, err := foreignKeys.get(colMeta.Name)
					if err != nil {
						return err
					}
					colMeta.IsForeignKey = true
					colMeta.TargetTableFK = fk.TargetTable
					colMeta.DeleteRule = fk.DeleteRule
					colMeta.UpdateRule = fk.UpdateRule
				}

				if hasEnum := enums.exists(colMeta.Name); hasEnum {
					enum, err := enums.get(colMeta.Name)
					if err != nil {
						return err
					}
					colMeta.HasENUM = true
					colMeta.ENUMName = enum.EnumName
					colMeta.ENUMValues = strings.Split(enum.EnumValue, ",")
				}

				if exists := uniques.exists(colMeta.Name, tableNames[i]); exists {
					uCol, err := uniques.get(colMeta.Name, tableNames[i])
					if err != nil {
						return err
					}
					colMeta.IsUnique = true
					colMeta.UniqueIndexDefinition = uCol.Definition
				}

				t := table{Name: tableNames[i]}
				err = storage.AddTable(t)
				if err != nil {
					return err
				}

				err = storage.AddColMetaData(tableNames[i], colMeta)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
