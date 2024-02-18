package storage

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"

	"server/config"
)

type Database struct {
	DB           *pgx.ConnPool
	User         string
	Password     string
	DataBaseName string
}

func (dbInfo *Database) Init(config config.Config) error {
	dbInfo.User = config.Database.User
	dbInfo.Password = config.Database.Password
	dbInfo.DataBaseName = config.Database.Name

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s port=%s",
		config.Database.Host,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SslMode,
		config.Database.Port)
	pgxConn, err := pgx.ParseConnectionString(connStr)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	pgxConn.PreferSimpleProtocol = true
	confPGX := pgx.ConnPoolConfig{
		ConnConfig:     pgxConn,
		MaxConnections: 100,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	}

	connPool, err := pgx.NewConnPool(confPGX)
	for {
		if err != nil {
			fmt.Println(err.Error() + " : " + connStr)
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			fmt.Println("Connected to db!")
			break
		}
		connPool, err = pgx.NewConnPool(confPGX)
	}
	dbInfo.DB = connPool

	return err
}
