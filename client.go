package mysqlclient

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sillyhatxu/mysql-client/customerrors"
	"sync"
)

type MysqlClient struct {
	config *Config
	mu     sync.Mutex
}

func NewMysqlClient(opts ...Option) (*MysqlClient, error) {
	//default
	config := &Config{
		ddlPath: "",
		flyway:  false,
	}
	for _, opt := range opts {
		opt(config)
	}
	mc := &MysqlClient{
		config: config,
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	err := mc.validate()
	if err != nil {
		return nil, err
	}
	err = mc.initialFlayway()
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func (mc *MysqlClient) validate() error {
	if mc.config == nil {
		return customerrors.CheckConfigNilError
	}
	if mc.config.pool == nil {
		return customerrors.CheckDBPoolError
	}
	return mc.Ping()
}

func (mc *MysqlClient) Ping() error {
	return mc.GetDB().Ping()
}

func (mc *MysqlClient) GetDB() *sql.DB {
	return mc.config.pool
}

func (mc *MysqlClient) GetTransaction() (*sql.Tx, error) {
	return mc.GetDB().Begin()
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
