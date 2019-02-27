package dbclient

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sillyhatxu/microlog"
)

type ClientConfig struct {
	dataSourceName string
	maxIdleConns   int
	maxOpenConns   int
}

var Client ClientConfig

func InitialDBClient(dataSourceName string, maxIdleConns int, maxOpenConns int) {
	Client.dataSourceName = dataSourceName
	Client.maxIdleConns = maxIdleConns
	Client.maxOpenConns = maxOpenConns
}

func (client *ClientConfig) getConnection() (*sql.DB, error) {
	db, err := sql.Open("mysql", client.dataSourceName)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Error("ping mysql error.", err)
		return nil, err
	}
	//mysqlClient.pool.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	db.SetMaxIdleConns(client.maxIdleConns)
	db.SetMaxOpenConns(client.maxOpenConns)
	return db, nil
}

func (client *ClientConfig) Insert(sql string, args ...interface{}) (int64, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return 0, err
	}
	stm, err := db.Prepare(sql)
	if err != nil {
		log.Error("prepare mysql error.", err)
		return 0, err
	}
	defer stm.Close()
	result, err := stm.Exec(args...)
	if err != nil {
		log.Error("insert data error.", err)
		return 0, err
	}
	return result.LastInsertId()
}

func (client *ClientConfig) Update(sql string, args ...interface{}) (int64, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return 0, err
	}
	stm, err := db.Prepare(sql)
	if err != nil {
		log.Error("prepare mysql error.", err)
		return 0, err
	}
	defer stm.Close()
	result, err := stm.Exec(args...)
	if err != nil {
		log.Error("update data error.", err)
		return 0, err
	}
	return result.RowsAffected()
}
