package infra

import (
	"fmt"
	"git.garena.com/xinlong.wu/zoo/config"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"xorm.io/xorm"
)

var DB *xorm.Engine

func InitDB() {
	c := config.Config
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s", c.Mysql.Username, c.Mysql.Password, c.Mysql.Address, c.Mysql.Database)
	db, err := xorm.NewEngine("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("InitDB error: %s\n", err.Error())
	}
	log.Printf("InitDB success")
	DB = db
}
