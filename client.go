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
	defer db.Close()
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
	defer db.Close()
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

func (client *ClientConfig) Find(sql string) ([]map[string]interface{}, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return nil, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Error("mysql client get transaction error.", err)
		return nil, err
	}
	defer tx.Commit()
	rows, err := tx.Query(sql)
	if err != nil {
		log.Error("Query error.", err)
		return nil, err
	}
	defer rows.Close()
	//Read database columns
	columns, err := rows.Columns()
	if err != nil {
		log.Error("rows.Columns() error.", err)
		return nil, err
	}
	//values是每个列的值，这里获取到byte里
	values := make([][]byte, len(columns))
	//query.Scan的参数，因为每次查询出来的列是不定长的，用len(cols)定住当次查询的长度
	scans := make([]interface{}, len(columns))
	//让每一行数据都填充到[][]byte里面
	for i := range values {
		scans[i] = &values[i]
	}
	//最后得到的map
	var results []map[string]interface{}
	for rows.Next() { //循环，让游标往下推
		if err := rows.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			return nil, err
		}
		row := make(map[string]interface{}) //每行数据
		for k, v := range values {          //每行数据是放在values里面，现在把它挪到row里
			key := columns[k]
			//valueType := reflect.TypeOf(v)
			//log.Info(valueType)
			row[key] = string(v)
		}
		results = append(results, row)
	}
	return results, nil
}

func (client *ClientConfig) FindOne(sql string) (map[string]interface{}, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return nil, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Error("mysql client get transaction error.", err)
		return nil, err
	}
	defer tx.Commit()
	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//读出查询出的列字段名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	//values是每个列的值，这里获取到byte里
	values := make([][]byte, len(columns))
	//query.Scan的参数，因为每次查询出来的列是不定长的，用len(cols)定住当次查询的长度
	scans := make([]interface{}, len(columns))
	//让每一行数据都填充到[][]byte里面
	for i := range values {
		scans[i] = &values[i]
	}
	//最后得到的map
	for rows.Next() { //循环，让游标往下推
		if err := rows.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			return nil, err
		}
		row := make(map[string]interface{}) //每行数据
		for k, v := range values {          //每行数据是放在values里面，现在把它挪到row里
			key := columns[k]
			row[key] = string(v)
		}
		return row, nil
	}
	return nil, nil
}

type BatchCallback func(*sql.Tx) (int, error)

func (client *ClientConfig) BatchInsert(callback BatchCallback) (int, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return 0, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Error("mysql client get transaction error.", err)
		return 0, err
	}
	result, err := callback(tx)
	if err != nil {
		log.Error("batch insert data error.", err)
		return 0, err
	}
	defer tx.Commit()
	return result, nil
}

func (client *ClientConfig) BatchUpdate(callback BatchCallback) (int, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return 0, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Error("mysql client get transaction error.", err)
		return 0, err
	}
	result, err := callback(tx)
	if err != nil {
		log.Error("batch update data error.", err)
		return 0, err
	}
	defer tx.Commit()
	return result, nil
}

func (client *ClientConfig) Count(sql string) (int, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return 0, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Error("mysql client get connection error.", err)
		return 0, err
	}
	defer tx.Commit()
	var count int
	countErr := tx.QueryRow(sql).Scan(&count)
	if countErr != nil {
		log.Error("Query count error.", err)
		return 0, err
	}
	return count, nil
}

//return affected count
func (client *ClientConfig) Delete(sql string, args ...interface{}) (int64, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return 0, err
	}
	defer db.Close()
	stm, err := db.Prepare(sql)
	if err != nil {
		log.Error("mysql client get connection error.", err)
		return 0, err
	}
	defer stm.Close()
	result, err := stm.Exec(args...)
	if err != nil {
		log.Error("delete data error.", err)
		return 0, err
	}
	return result.RowsAffected()
}

func (client *ClientConfig) DeleteByPrimaryKey(sql string, id int64) (int64, error) {
	db, err := client.getConnection()
	if err != nil {
		log.Error("mysql get connection error.", err)
		return 0, err
	}
	defer db.Close()
	stm, err := db.Prepare(sql)
	if err != nil {
		log.Error("mysql client get connection error.", err)
		return 0, err
	}
	defer stm.Close()
	result, err := stm.Exec(id)
	if err != nil {
		log.Error("DeleteByPrimaryKey error.", err)
		return 0, err
	}
	return result.RowsAffected()
}
