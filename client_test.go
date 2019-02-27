package dbclient

import (
	log "github.com/sillyhatxu/microlog"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type Userinfo struct {
	Id               int64     `mapstructure:"id"`
	Name             string    `mapstructure:"name"`
	Age              int       `mapstructure:"age"`
	Description      string    `mapstructure:"description"`
	IsDelete         bool      `mapstructure:"is_delete"`
	CreatedTime      time.Time `mapstructure:"created_date"`
	LastModifiedDate time.Time `mapstructure:"last_modified_date"`
}

const dataSourceName = `sillyhat:sillyhat@tcp(127.0.0.1:3308)/sillyhat`

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
	InitialDBClient(dataSourceName, 5, 10)
	id, err := Client.Insert(insert_sql, "test name", 21, "1989-06-09", "This is description", false)
	log.Info("id : ", id)
	assert.Nil(t, err)
	assert.EqualValues(t, id, 2)
}

func TestClientUpdate(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	count, err := Client.Update(update_sql, "test name update", 22, "2000-01-01", "This is update result", true, 2)
	log.Info("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 1)
}
