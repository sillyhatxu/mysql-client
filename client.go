package dbclient

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
	"github.com/sillyhatxu/retry-utils"
	"github.com/sirupsen/logrus"
	"net/url"
	"sync"
	"time"
)

type MysqlClient struct {
	userName string
	password string
	host     string
	port     int
	schema   string
	config   *Config
	db       *sql.DB
	mu       sync.Mutex
}

func NewMysqlClient(userName string, password string, host string, port int, schema string, opts ...Option) (*MysqlClient, error) {
	//default
	config := &Config{
		local:           "Asia/Singapore",
		parseTime:       true,
		maxIdleConns:    20,
		maxOpenConns:    40,
		connMaxLifetime: 24 * time.Hour,
		attempts:        3,
		delay:           200 * time.Millisecond,
		ddlPath:         "",
		flyway:          false,
	}
	for _, opt := range opts {
		opt(config)
	}
	mysqlClient := &MysqlClient{
		userName: userName,
		password: password,
		host:     host,
		port:     port,
		schema:   schema,
		config:   config,
	}
	return mysqlClient, mysqlClient.initial()
}

func (mc *MysqlClient) getMysqlDataSourceName() string {
	params := url.Values{}
	params.Add("loc", mc.config.local)
	params.Add("parseTime", fmt.Sprintf("%t", mc.config.parseTime))
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", mc.userName, mc.password, mc.host, mc.port, mc.schema, params.Encode())
}

func (mc *MysqlClient) initial() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	db, err := mc.OpenDataSource()
	if err != nil {
		logrus.Errorf("mc.OpenDataSource error; MysqlClient : %#v,%v", mc, err)
		return err
	}
	mc.db = db
	err = mc.Ping()
	if err != nil {
		logrus.Errorf("ping database error; %v", err)
		return err
	}
	err = mc.initialFlayway()
	if err != nil {
		logrus.Errorf("initial flayway error; %v", err)
		return err
	}
	return nil
}

func (mc *MysqlClient) OpenDataSource() (*sql.DB, error) {
	var resultDB *sql.DB
	err := retry.Do(func() error {
		dataSourceName := mc.getMysqlDataSourceName()
		logrus.Infof("connect : %s", dataSourceName)
		db, err := sql.Open("mysql", dataSourceName)
		if err != nil {
			logrus.Errorf("open data source name error. %v", err)
			return err
		}
		resultDB = db
		return nil
	}, retry.ErrorCallback(func(n uint, err error) {
		logrus.Errorf("retry [%d] open data source error. %v", n, err)
	}))
	if err != nil {
		return nil, err
	}
	if resultDB == nil {
		return nil, fmt.Errorf("open datasource error. db is nil. %#v", mc)
	}
	return resultDB, nil
}

func (mc *MysqlClient) Ping() error {
	return mc.db.Ping()
}

func (mc *MysqlClient) GetDB() (*sql.DB, error) {
	if err := mc.Ping(); err != nil {
		logrus.Errorf("ping database error. %v", err)
		db, err := mc.OpenDataSource()
		if err != nil {
			logrus.Errorf("mc.OpenDataSource error; MysqlClient : %#v,%v", mc, err)
			return nil, err
		}
		mc.db = db
	}
	return mc.db, nil
}

func (mc *MysqlClient) GetTransaction() (*sql.Tx, error) {
	var transaction *sql.Tx
	err := retry.Do(func() error {
		db, err := mc.GetDB()
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			logrus.Errorf("db.begin get transaction error; %v", err)
			return err
		}
		transaction = tx
		return nil
	}, retry.ErrorCallback(func(n uint, err error) {
		logrus.Errorf("retry [%d] get transaction error. %v", n, err)
	}))
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (mc *MysqlClient) ExecDDL(ddl string) error {
	db, err := mc.GetDB()
	if err != nil {
		return err
	}
	logrus.Infof("exec ddl : ")
	logrus.Infof(ddl)
	logrus.Infof("--------------------")
	_, err = db.Exec(ddl)
	return err
}

type FieldFunc func(rows *sql.Rows) error

func (mc *MysqlClient) FindCustom(query string, fieldFunc FieldFunc, args ...interface{}) error {
	db, err := mc.GetDB()
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

func (mc *MysqlClient) FindMapArray(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	tx, err := mc.GetTransaction()
	if err != nil {
		logrus.Errorf("get transaction error; %v", err)
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

func (mc *MysqlClient) FindMapFirst(sql string, args ...interface{}) (map[string]interface{}, error) {
	array, err := mc.FindMapArray(sql, args...)
	if err != nil {
		return nil, err
	}
	if array == nil || len(array) == 0 {
		return nil, nil
	}
	return array[0], nil
}

func (mc *MysqlClient) FindList(sql string, input interface{}, args ...interface{}) error {
	if isSlice(input) {
		return fmt.Errorf("%v must be a slice", input)
	}
	results, err := mc.FindMapArray(sql, args...)
	if err != nil {
		return err
	}
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc("2006-01-02T15:04:05Z07:00"),
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

func (mc *MysqlClient) FindListByConfig(sql string, config *mapstructure.DecoderConfig, args ...interface{}) error {
	results, err := mc.FindMapArray(sql, args...)
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

func (mc *MysqlClient) FindFirst(sql string, input interface{}, args ...interface{}) error {
	if isStruct(input) {
		return fmt.Errorf("%v must be a struct or a struct pointer", input)
	}
	result, err := mc.FindMapFirst(sql, args...)
	if err != nil {
		return err
	}
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc("2006-01-02T15:04:05Z07:00"),
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

func (mc *MysqlClient) FindFirstByConfig(sql string, config *mapstructure.DecoderConfig, args ...interface{}) error {
	result, err := mc.FindMapFirst(sql, args...)
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

func (mc *MysqlClient) Count(sql string, args ...interface{}) (int64, error) {
	tx, err := mc.GetTransaction()
	if err != nil {
		logrus.Errorf("db.begin get connection error; %v", err)
		return 0, err
	}
	defer tx.Commit()
	var count int64
	err = tx.QueryRow(sql, args...).Scan(&count)
	if err != nil {
		logrus.Errorf("query count error; %v", err)
		return 0, err
	}
	return count, nil
}

func (mc *MysqlClient) Insert(sql string, args ...interface{}) (int64, error) {
	db, err := mc.GetDB()
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

func (mc *MysqlClient) Transaction(callback TransactionCallback) error {
	tx, err := mc.GetTransaction()
	if err != nil {
		logrus.Errorf("db.begin get transaction error; %v", err)
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

func (mc *MysqlClient) Update(sql string, args ...interface{}) (int64, error) {
	db, err := mc.GetDB()
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

func (mc *MysqlClient) Delete(sql string, args ...interface{}) (int64, error) {
	db, err := mc.GetDB()
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
