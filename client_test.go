package dbclient

import (
	"database/sql"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"log"
	"strconv"
	"testing"
	"time"
)

type Userinfo struct {
	Id               int64     `mapstructure:"id"`
	Name             string    `mapstructure:"name"`
	Age              int       `mapstructure:"age"`
	Birthday         time.Time `mapstructure:"birthday"`
	Description      string    `mapstructure:"description"`
	IsDelete         bool      `mapstructure:"is_delete"`
	CreatedTime      time.Time `mapstructure:"created_date"`
	LastModifiedDate time.Time `mapstructure:"last_modified_date"`
}

const (
	dataSourceName = `sillyhat:sillyhat@tcp(127.0.0.1:3308)/sillyhat`
	maxIdleConns   = 5
	maxOpenConns   = 10
)

const (
	insert_sql = `
		insert into userinfo 
		(name, age, birthday, description, is_delete, created_date, last_modified_date)
		values (?, ?, ?, ?, ?, now(), now())
	`
	update_sql = `
		UPDATE userinfo
		SET name               = ?,
		    age                = ?,
		    birthday           = ?,
		    description        = ?,
		    is_delete          = ?,
		    last_modified_date = now()
		WHERE id = ?;
	`

	count_sql = `
		select count(1) from userinfo
	`

	findAll_sql = `
		select id,
		       name,
		       age,
		       TIMESTAMP(birthday) birthday,
		       description,
		       (is_delete = b'1')  is_delete,
		       created_date,
		       last_modified_date
		from userinfo
		where age > ? and is_delete = ? and name like ?
	`

	findOne_sql = `
		select id,name, age, TIMESTAMP(birthday) birthday, description, (is_delete = b'1') is_delete, created_date, last_modified_date from userinfo where id = ? and is_delete = ?
	`

	deleteOne_sql = `
		delete from userinfo where id = ? 
	`

	delete_sql = `
		delete from userinfo where id in (?,?,?,?,?,?,?,?,?,?)
	`

	create_table = `
		create table userinfo
		(
			id int auto_increment,
			name varchar(100) null,
			age int null,
			birthday date null,
			description text null,
			is_delete bit default 0 not null,
			created_date timestamp null,
			last_modified_date timestamp null,
			constraint userinfo_pk primary key (id)
		)		
		`
)

func TestClientInsert(t *testing.T) {
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	id, err := Client.Insert(insert_sql, "test name", 21, "1989-06-09", "This is description", false)
	log.Println("id : ", id)
	assert.Nil(t, err)
	assert.EqualValues(t, id, 2)
}

func TestClientUpdate(t *testing.T) {
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	count, err := Client.Update(update_sql, "test name update", 22, "2000-01-01", "This is update result", true, 2)
	log.Println("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 1)
}

func TestClientFindOne(t *testing.T) {
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	result, err := Client.FindOne(findOne_sql, "2", true)
	assert.Nil(t, err)
	var user *Userinfo
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc("2006-01-02 15:04:05"),
		WeaklyTypedInput: true,
		Result:           &user,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic(err)
	}
	err = decoder.Decode(result)
	if err != nil {
		panic(err)
	}
	layout := "2006-01-02 15:04:05"
	assert.EqualValues(t, user.Id, 2)
	assert.EqualValues(t, user.Name, "test name update")
	assert.EqualValues(t, user.Description, "This is update result")
	assert.EqualValues(t, user.IsDelete, true)
	birthday, err := time.Parse(layout, "2000-01-01 00:00:00")
	assert.EqualValues(t, user.Birthday, birthday)
	createdTime, err := time.Parse(layout, "2019-02-27 05:39:55")
	assert.EqualValues(t, user.CreatedTime, createdTime)
	lastModifiedDate, err := time.Parse(layout, "2019-02-27 07:42:54")
	assert.EqualValues(t, user.LastModifiedDate, lastModifiedDate)
}

func TestClientFind(t *testing.T) {
	log.Printf("initial db client. dataSourceName : %v ; maxIdleConns : %v ; maxOpenConns : %v", dataSourceName, maxIdleConns, maxOpenConns)
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	results, err := Client.Find(findAll_sql, 21, true, "%update name%")
	assert.Nil(t, err)
	var userArray []Userinfo
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeHookFunc("2006-01-02 15:04:05"),
		WeaklyTypedInput: true,
		Result:           &userArray,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic(err)
	}
	err = decoder.Decode(results)
	if err != nil {
		panic(err)
	}
	assert.EqualValues(t, len(userArray), 13)
	assert.EqualValues(t, userArray[0].Id, 357)
	assert.EqualValues(t, userArray[1].Id, 358)

	//assert.EqualValues(t, len(userArray), 14)
	//assert.EqualValues(t, userArray[0].Id, 2)
	//assert.EqualValues(t, userArray[1].Id, 357)
}

func TestClientBatchInsert(t *testing.T) {
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	result, err := Client.BatchInsert(func(tx *sql.Tx) (int64, error) {
		totalCount := 0
		for i := 1001; i <= 2000; i++ {
			_, err := tx.Exec(insert_sql, "test name"+strconv.Itoa(i), 21, "1989-06-09", "This is description", false)
			assert.Nil(t, err)
			totalCount++
		}
		return int64(totalCount), nil
	})
	assert.Nil(t, err)
	assert.EqualValues(t, result, 1000)
}

func TestClientBatchUpdate(t *testing.T) {
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	result, err := Client.BatchUpdate(func(tx *sql.Tx) (int64, error) {
		totalCount := 0
		for i := 3; i <= 1002; i++ {
			_, err := tx.Exec(update_sql, "test update name -"+strconv.Itoa(i), 21, "2005-01-30", "This is update", true, i)
			assert.Nil(t, err)
			totalCount++
		}
		return int64(totalCount), nil
	})
	assert.Nil(t, err)
	assert.EqualValues(t, result, 1000)
}

func TestClientCount(t *testing.T) {
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	count, err := Client.Count(count_sql)
	log.Println("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 981)
}

func TestClientDeleteOne(t *testing.T) {
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	count, err := Client.DeleteByPrimaryKey(deleteOne_sql, 3)
	log.Println("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 1)
}

func TestClientDelete(t *testing.T) {
	InitialDBClient(dataSourceName, maxIdleConns, maxOpenConns)
	count, err := Client.Delete(delete_sql, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23)
	log.Println("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 10)
}
