package dbclient

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ClientConf struct {
	DataSourceName  string
	DDLPath         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	Flyway          bool
	db              *sql.DB
	mu              sync.Mutex
}

func NewMysqlClientConf(DataSourceName string) *ClientConf {
	return &ClientConf{
		DataSourceName:  DataSourceName,
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 24 * time.Hour,
	}
}

func (cc *ClientConf) SetMaxIdleConns(MaxIdleConns int) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.MaxIdleConns = MaxIdleConns
}

func (cc *ClientConf) SetMaxOpenConns(MaxOpenConns int) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.MaxOpenConns = MaxOpenConns
}

func (cc *ClientConf) SetConnMaxLifetime(SetConnMaxLifetime time.Duration) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.ConnMaxLifetime = SetConnMaxLifetime
}

func (cc *ClientConf) Initial() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	db, err := sql.Open("mysql", cc.DataSourceName)
	if err != nil {
		logrus.Errorf("open database error; %v", err)
		return err
	}
	err = db.Ping()
	if err != nil {
		logrus.Errorf("ping database error; %v", err)
		return err
	}
	cc.db = db
	return nil
}

func (cc *ClientConf) GetDB() (*sql.DB, error) {
	if err := cc.db.Ping(); err != nil {
		logrus.Errorf("get connect error. %v", err)
		return nil, err
	}
	return cc.db, nil
}

type FieldFunc func(rows *sql.Rows) error

func (cc *ClientConf) FindCustom(query string, fieldFunc FieldFunc, args ...interface{}) error {
	db, err := cc.GetDB()
	if err != nil {
		return err
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := fieldFunc(rows)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}

func (cc *ClientConf) FindMapArray(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	db, err := cc.GetDB()
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		logrus.Errorf("mysql client get transaction error; %v", err)
		return nil, err
	}
	defer tx.Commit()
	rows, err := tx.Query(sql, args...)
	if err != nil {
		logrus.Errorf("query error; %v", err)
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		logrus.Errorf("rows.Columns() error; %v", err)
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

func (cc *ClientConf) FindMapFirst(sql string, args ...interface{}) (map[string]interface{}, error) {
	array, err := cc.FindMapArray(sql, args...)
	if err != nil {
		return nil, err
	}
	if array == nil || len(array) == 0 {
		return nil, nil
	}
	return array[0], nil
}

func (cc *ClientConf) FindList(sql string, input interface{}, args ...interface{}) error {
	results, err := cc.FindMapArray(sql, args...)
	if err != nil {
		return err
	}
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc("2006-01-02 15:04:05"),
		WeaklyTypedInput: true,
		Result:           input,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	err = decoder.Decode(results)
	if err != nil {
		return err
	}
	return nil
}

func (cc *ClientConf) FindListByConfig(sql string, input interface{}, config *mapstructure.DecoderConfig, args ...interface{}) error {
	results, err := cc.FindMapArray(sql, args...)
	if err != nil {
		return err
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	err = decoder.Decode(results)
	if err != nil {
		return err
	}
	return nil
}

func (cc *ClientConf) FindFirst(sql string, input interface{}, args ...interface{}) error {
	result, err := cc.FindMapFirst(sql, args...)
	if err != nil {
		return err
	}
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc("2006-01-02 15:04:05"),
		WeaklyTypedInput: true,
		Result:           input,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	err = decoder.Decode(result)
	if err != nil {
		return err
	}
	return nil
}

func (cc *ClientConf) FindFirstByConfig(sql string, input interface{}, config *mapstructure.DecoderConfig, args ...interface{}) error {
	result, err := cc.FindMapFirst(sql, args...)
	if err != nil {
		return err
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	err = decoder.Decode(result)
	if err != nil {
		return err
	}
	return nil
}

func (cc *ClientConf) Count(sql string, args ...interface{}) (int64, error) {
	db, err := cc.GetDB()
	if err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		logrus.Errorf("mysql client get connection error; %v", err)
		return 0, err
	}
	defer tx.Commit()
	var count int64
	countErr := tx.QueryRow(sql, args...).Scan(&count)
	if countErr != nil {
		logrus.Errorf("query count error; %v", err)
		return 0, err
	}
	return count, nil
}

func (cc *ClientConf) Insert(sql string, args ...interface{}) (int64, error) {
	db, err := cc.GetDB()
	if err != nil {
		return 0, nil
	}
	stm, err := db.Prepare(sql)
	if err != nil {
		logrus.Errorf("prepare mysql error; %v", err)
		return 0, err
	}
	defer stm.Close()
	result, err := stm.Exec(args...)
	if err != nil {
		logrus.Errorf("insert data error; %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

type TransactionCallback func(*sql.Tx) error

func (cc *ClientConf) Transaction(callback TransactionCallback) error {
	db, err := cc.GetDB()
	if err != nil {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		logrus.Errorf("mysql client get transaction error; %v", err)
		return err
	}
	err = callback(tx)
	if err != nil {
		logrus.Errorf("transaction data error; %v", err)
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (cc *ClientConf) Update(sql string, args ...interface{}) (int64, error) {
	db, err := cc.GetDB()
	if err != nil {
		return 0, nil
	}
	stm, err := db.Prepare(sql)
	if err != nil {
		logrus.Errorf("prepare mysql error; %v", err)
		return 0, err
	}
	defer stm.Close()
	result, err := stm.Exec(args...)
	if err != nil {
		logrus.Errorf("update data error; %v", err)
		return 0, err
	}
	return result.RowsAffected()
}

func (cc *ClientConf) Delete(sql string, args ...interface{}) (int64, error) {
	db, err := cc.GetDB()
	if err != nil {
		return 0, nil
	}
	stm, err := db.Prepare(sql)
	if err != nil {
		logrus.Errorf("prepare mysql error; %v", err)
		return 0, err
	}
	defer stm.Close()
	result, err := stm.Exec(args...)
	if err != nil {
		logrus.Errorf("delete data error; %v", err)
		return 0, err
	}
	return result.RowsAffected()
}
