package db

import (
	"fmt"
	"modelgenerator/conf"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var Instance *gorm.DB

func Init() {
	dsn := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s",
		conf.DbConfig.User,
		conf.DbConfig.Pwd,
		conf.DbConfig.Host,
		conf.DbConfig.Port,
		conf.DbConfig.DbName,
	)

	var err error
	Instance, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}
