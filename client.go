package dbclient

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/url"
	"sync"
	"time"
)

type MysqlClient struct {
	config *Config
	pool   *sql.DB
	mu     sync.Mutex
}

//checkConnLiveness = true
//flayway open multiStatements = true
const (
	dsnFormat               = "%s:%s@tcp(%s:%d)/%s?%s"
	driverName              = "mysql"
	allowAllFiles           = true
	allowCleartextPasswords = true
	allowNativePasswords    = false
	allowOldPasswords       = true
	charset                 = "utf8mb4"
	checkConnLiveness       = false
	collation               = "utf8mb4_general_ci"
	clientFoundRows         = false
	columnsWithAlias        = false
	interpolateParams       = false
	loc                     = "Asia/Singapore" //Asia%2FSingapore
	maxAllowedPacket        = 4194304          //default : 4MB
	multiStatements         = false
	parseTime               = true
	readTimeout             = time.Duration(30) * time.Second
	rejectReadOnly          = false
	serverPubKey            = "none"
	timeout                 = time.Duration(30) * time.Second
	tls                     = false
	writeTimeout            = time.Duration(30) * time.Second
	maxIdleConns            = 2
	maxOpenConns            = 5
	connMaxLifetime         = time.Duration(6) * time.Hour
)

func NewMysqlClient(opts ...Option) (*MysqlClient, error) {
	//default
	config := &Config{
		allowAllFiles:           allowAllFiles,
		allowCleartextPasswords: allowCleartextPasswords,
		allowNativePasswords:    allowNativePasswords,
		allowOldPasswords:       allowOldPasswords,
		charset:                 charset,
		checkConnLiveness:       checkConnLiveness,
		collation:               collation,
		clientFoundRows:         clientFoundRows,
		columnsWithAlias:        columnsWithAlias,
		interpolateParams:       interpolateParams,
		loc:                     loc,
		maxAllowedPacket:        maxAllowedPacket,
		multiStatements:         multiStatements,
		parseTime:               parseTime,
		readTimeout:             readTimeout,
		rejectReadOnly:          rejectReadOnly,
		serverPubKey:            serverPubKey,
		timeout:                 timeout,
		tls:                     tls,
		writeTimeout:            writeTimeout,
		maxIdleConns:            maxIdleConns,
		maxOpenConns:            maxOpenConns,
		connMaxLifetime:         connMaxLifetime,
		ddlPath:                 "",
		flyway:                  false,
	}
	for _, opt := range opts {
		opt(config)
	}
	mc := &MysqlClient{
		config: config,
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	pool, err := mc.getDatabasePool()
	if err != nil {
		return nil, err
	}
	mc.pool = pool
	err = mc.Ping()
	if err != nil {
		return nil, err
	}
	err = mc.initialFlayway()
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func (mc *MysqlClient) getMysqlDataSourceName() string {
	params := url.Values{}
	params.Add("allowAllFiles", setupBool(mc.config.allowAllFiles))
	params.Add("allowCleartextPasswords", setupBool(mc.config.allowCleartextPasswords))
	params.Add("allowNativePasswords", setupBool(mc.config.allowNativePasswords))
	params.Add("allowOldPasswords", setupBool(mc.config.allowOldPasswords))
	params.Add("charset", mc.config.charset)
	params.Add("checkConnLiveness", setupBool(mc.config.checkConnLiveness))
	params.Add("collation", mc.config.collation)
	params.Add("clientFoundRows", setupBool(mc.config.clientFoundRows))
	params.Add("columnsWithAlias", setupBool(mc.config.columnsWithAlias))
	params.Add("interpolateParams", setupBool(mc.config.interpolateParams))
	params.Add("loc", mc.config.loc)
	params.Add("maxAllowedPacket", setupInt64(mc.config.maxAllowedPacket))
	params.Add("multiStatements", setupBool(mc.config.multiStatements))
	params.Add("parseTime", fmt.Sprintf("%t", mc.config.parseTime))
	params.Add("readTimeout", setupTime(mc.config.readTimeout))
	params.Add("rejectReadOnly", setupBool(mc.config.rejectReadOnly))
	params.Add("serverPubKey", mc.config.serverPubKey)
	params.Add("timeout", setupTime(mc.config.timeout))
	params.Add("tls", setupBool(mc.config.tls))
	params.Add("writeTimeout", setupTime(mc.config.writeTimeout))
	params.Add("maxIdleConns", setupInt(mc.config.maxIdleConns))
	params.Add("maxOpenConns", setupInt(mc.config.maxOpenConns))
	params.Add("connMaxLifetime", setupTime(mc.config.connMaxLifetime))

	params.Add("parseTime", fmt.Sprintf("%t", mc.config.parseTime))
	return fmt.Sprintf(dsnFormat, mc.config.userName, mc.config.password, mc.config.host, mc.config.port, mc.config.schema, params.Encode())
}

func (mc *MysqlClient) getDatabasePool() (*sql.DB, error) {
	dataSourceName := mc.getMysqlDataSourceName()
	pool, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	pool.SetMaxIdleConns(mc.config.maxIdleConns)
	pool.SetMaxOpenConns(mc.config.maxOpenConns)
	pool.SetConnMaxLifetime(mc.config.connMaxLifetime)
	return pool, nil
}

func (mc *MysqlClient) Ping() error {
	return mc.pool.Ping()
}

func (mc *MysqlClient) GetDB() *sql.DB {
	return mc.pool
}

func (mc *MysqlClient) GetTransaction() (*sql.Tx, error) {
	return mc.GetDB().Begin()
}

func (mc *MysqlClient) ExecDDL(ddl string) error {
	startT := time.Now()
	result, err := mc.GetDB().Exec(ddl)
	if err != nil {
		return err
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	rowsAffected, err := result.LastInsertId()
	if err != nil {
		return err
	}
	tc := time.Since(startT)
	log.Println("lastInsertId:", lastInsertId, "; rowsAffected : ", rowsAffected, " (execution: ", tc, ")")
	return nil
}

type FieldFunc func(rows *sql.Rows) error

func (mc *MysqlClient) FindCustom(query string, fieldFunc FieldFunc, args ...interface{}) error {
	rows, err := mc.GetDB().Query(query, args...)
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

//func (mc *MysqlClient) FindMapArray(sql string, args ...interface{}) ([]map[string]interface{}, error) {
//	tx, err := mc.GetTransaction()
//	if err != nil {
//		logrus.Errorf("get transaction error; %v", err)
//		return nil, err
//	}
//	defer tx.Commit()
//	rows, err := tx.Query(sql, args...)
//	if err != nil {
//		logrus.Errorf("query error; %v", err)
//		return nil, err
//	}
//	defer rows.Close()
//	columns, err := rows.Columns()
//	if err != nil {
//		logrus.Errorf("rows.Columns() error; %v", err)
//		return nil, err
//	}
//	//values是每个列的值，这里获取到byte里
//	values := make([][]byte, len(columns))
//	//query.Scan的参数，因为每次查询出来的列是不定长的，用len(cols)定住当次查询的长度
//	scans := make([]interface{}, len(columns))
//	//让每一行数据都填充到[][]byte里面
//	for i := range values {
//		scans[i] = &values[i]
//	}
//	//最后得到的map
//	var results []map[string]interface{}
//	for rows.Next() { //循环，让游标往下推
//		if err := rows.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
//			return nil, err
//		}
//		row := make(map[string]interface{}) //每行数据
//		for k, v := range values {          //每行数据是放在values里面，现在把它挪到row里
//			key := columns[k]
//			//valueType := reflect.TypeOf(v)
//			//log.Info(valueType)
//			row[key] = string(v)
//		}
//		results = append(results, row)
//	}
//	return results, nil
//}
//
//func (mc *MysqlClient) FindMapFirst(sql string, args ...interface{}) (map[string]interface{}, error) {
//	array, err := mc.FindMapArray(sql, args...)
//	if err != nil {
//		return nil, err
//	}
//	if array == nil || len(array) == 0 {
//		return nil, nil
//	}
//	return array[0], nil
//}
//
//func (mc *MysqlClient) FindList(sql string, input interface{}, args ...interface{}) error {
//	if isSlice(input) {
//		return fmt.Errorf("%v must be a slice", input)
//	}
//	results, err := mc.FindMapArray(sql, args...)
//	if err != nil {
//		return err
//	}
//	config := &mapstructure.DecoderConfig{
//		DecodeHook:       mapstructure.StringToTimeHookFunc("2006-01-02T15:04:05Z07:00"),
//		WeaklyTypedInput: true,
//		Result:           input,
//	}
//	decoder, err := mapstructure.NewDecoder(config)
//	if err != nil {
//		return err
//	}
//	err = decoder.Decode(results)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (mc *MysqlClient) FindListByConfig(sql string, config *mapstructure.DecoderConfig, args ...interface{}) error {
//	results, err := mc.FindMapArray(sql, args...)
//	if err != nil {
//		return err
//	}
//	decoder, err := mapstructure.NewDecoder(config)
//	if err != nil {
//		return err
//	}
//	err = decoder.Decode(results)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (mc *MysqlClient) FindFirst(sql string, input interface{}, args ...interface{}) error {
//	if isStruct(input) {
//		return fmt.Errorf("%v must be a struct or a struct pointer", input)
//	}
//	result, err := mc.FindMapFirst(sql, args...)
//	if err != nil {
//		return err
//	}
//	config := &mapstructure.DecoderConfig{
//		DecodeHook:       mapstructure.StringToTimeHookFunc("2006-01-02T15:04:05Z07:00"),
//		WeaklyTypedInput: true,
//		Result:           input,
//	}
//	decoder, err := mapstructure.NewDecoder(config)
//	if err != nil {
//		return err
//	}
//	err = decoder.Decode(result)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (mc *MysqlClient) FindFirstByConfig(sql string, config *mapstructure.DecoderConfig, args ...interface{}) error {
//	result, err := mc.FindMapFirst(sql, args...)
//	if err != nil {
//		return err
//	}
//	decoder, err := mapstructure.NewDecoder(config)
//	if err != nil {
//		return err
//	}
//	err = decoder.Decode(result)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (mc *MysqlClient) Count(sql string, args ...interface{}) (int64, error) {
//	tx, err := mc.GetTransaction()
//	if err != nil {
//		logrus.Errorf("db.begin get connection error; %v", err)
//		return 0, err
//	}
//	defer tx.Commit()
//	var count int64
//	err = tx.QueryRow(sql, args...).Scan(&count)
//	if err != nil {
//		logrus.Errorf("query count error; %v", err)
//		return 0, err
//	}
//	return count, nil
//}
//
//func (mc *MysqlClient) Insert(sql string, args ...interface{}) (int64, error) {
//	db, err := mc.GetDB()
//	if err != nil {
//		return 0, nil
//	}
//	stm, err := db.Prepare(sql)
//	if err != nil {
//		logrus.Errorf("prepare mysql error; %v", err)
//		return 0, err
//	}
//	defer stm.Close()
//	result, err := stm.Exec(args...)
//	if err != nil {
//		logrus.Errorf("insert data error; %v", err)
//		return 0, err
//	}
//	return result.LastInsertId()
//}
//
//type TransactionCallback func(*sql.Tx) error
//
//func (mc *MysqlClient) Transaction(callback TransactionCallback) error {
//	tx, err := mc.GetTransaction()
//	if err != nil {
//		logrus.Errorf("db.begin get transaction error; %v", err)
//		return err
//	}
//	err = callback(tx)
//	if err != nil {
//		logrus.Errorf("transaction data error; %v", err)
//		tx.Rollback()
//		return err
//	}
//	return tx.Commit()
//}
//
//func (mc *MysqlClient) Update(sql string, args ...interface{}) (int64, error) {
//	db, err := mc.GetDB()
//	if err != nil {
//		return 0, nil
//	}
//	stm, err := db.Prepare(sql)
//	if err != nil {
//		logrus.Errorf("prepare mysql error; %v", err)
//		return 0, err
//	}
//	defer stm.Close()
//	result, err := stm.Exec(args...)
//	if err != nil {
//		logrus.Errorf("update data error; %v", err)
//		return 0, err
//	}
//	return result.RowsAffected()
//}
//
//func (mc *MysqlClient) Delete(sql string, args ...interface{}) (int64, error) {
//	db, err := mc.GetDB()
//	if err != nil {
//		return 0, nil
//	}
//	stm, err := db.Prepare(sql)
//	if err != nil {
//		logrus.Errorf("prepare mysql error; %v", err)
//		return 0, err
//	}
//	defer stm.Close()
//	result, err := stm.Exec(args...)
//	if err != nil {
//		logrus.Errorf("delete data error; %v", err)
//		return 0, err
//	}
//	return result.RowsAffected()
//}
