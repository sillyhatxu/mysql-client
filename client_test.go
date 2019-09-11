package dbclient

import (
	"github.com/stretchr/testify/assert"
	"log"
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

func TestClientGetConnection(t *testing.T) {
	dbclient := NewMysqlClientConf(dataSourceName, Attempts(20), ConnMaxLifetime(500*time.Millisecond))
	err := dbclient.Initial()
	assert.Nil(t, err)
	for {
		count, err := dbclient.Count(count_sql)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(count)
		time.Sleep(5 * time.Second)
	}

}

func TestHasTable(t *testing.T) {
	dbclient := NewMysqlClientConf(dataSourceName)
	err := dbclient.Initial()
	assert.Nil(t, err)
	b, err := dbclient.HasTable("test")
	assert.Nil(t, err)
	assert.EqualValues(t, b, false)
	b, err = dbclient.HasTable("userinfo")
	assert.Nil(t, err)
	assert.EqualValues(t, b, true)

}
