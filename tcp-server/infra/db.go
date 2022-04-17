package infra

import (
	"database/sql"
	"fmt"
	"git.garena.com/xinlong.wu/zoo/config"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
	"xorm.io/xorm"
)

var XDB *xorm.Engine
var DB *sql.DB

func getDataSourceName() string {
	c := config.Config
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", c.Mysql.Username, c.Mysql.Password, c.Mysql.Address, c.Mysql.Database)
}

func InitXDB() {
	dataSourceName := getDataSourceName()
	db, err := xorm.NewEngine("mysql", dataSourceName)
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(100)
	//db.ShowSQL(true)
	if err != nil {
		log.Fatalf("InitXDB error: %s\n", err.Error())
	}
	log.Printf("InitXDB success")
	XDB = db
}

func InitDB() {
	dataSourceName := getDataSourceName()
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(100)
	db.SetConnMaxIdleTime(time.Minute * 10)
	DB = db
}
