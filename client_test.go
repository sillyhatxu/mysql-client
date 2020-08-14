package dbclient

import (
	"time"
)

// Model Struct
type User struct {
	Id               int64
	LoginName        string
	Password         string
	UserName         string
	Status           bool
	Platform         string
	Age              *int
	Amount           *float64
	Description      *string
	Birthday         *time.Time
	CreatedTime      time.Time
	LastModifiedTime time.Time
}

const (
	userName = "sillyhat"
	password = "sillyhat"
	host     = "192.168.1.87"
	port     = 3306
	schema   = "sillyhat_remind"
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

//
//func TestGetMysqlDataSourceName(t *testing.T) {
//	dbclient, err := NewMysqlClient(userName, password, host, port, schema)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(dbclient.getMysqlDataSourceName())
//}
//
//func TestClientGetConnection(t *testing.T) {
//	dbclient, err := NewMysqlClient(userName, password, host, port, schema, Attempts(20), ConnMaxLifetime(500*time.Millisecond))
//	if err != nil {
//		panic(err)
//	}
//	for {
//		count, err := dbclient.Count(count_sql)
//		if err != nil {
//			log.Println(err)
//			continue
//		}
//		log.Println(count)
//		time.Sleep(5 * time.Second)
//	}
//
//}
//
//func TestHasTable(t *testing.T) {
//	dbclient, err := NewMysqlClient(userName, password, host, port, schema)
//	if err != nil {
//		panic(err)
//	}
//	b, err := dbclient.HasTable("test")
//	assert.Nil(t, err)
//	assert.EqualValues(t, b, false)
//	b, err = dbclient.HasTable("userinfo")
//	assert.Nil(t, err)
//	assert.EqualValues(t, b, true)
//}
//
//func TestMysqlClient_Initial(t *testing.T) {
//	//sillyhat:sillyhat@tcp(127.0.0.1:3308)/sillyhat_user?loc=Asia%2FSingapore&parseTime=true
//	var Client, err = NewMysqlClient(userName, password, host, port, schema, DDLPath("/Users/shikuanxu/go/src/github.com/sillyhatxu/user-backend/db/migration"))
//	//var Client, err = NewMysqlClient(dataSourceName, DDLPath("/Users/cookie/go/gopath/src/github.com/sillyhatxu/mini-mq/db/migration"))
//	if err != nil {
//		panic(err)
//	}
//	err = Client.Ping()
//	assert.Nil(t, err)
//}
//
//func TestMysqlClient_SchemaVersionArray(t *testing.T) {
//	var Client, err = NewMysqlClient(userName, password, host, port, schema, DDLPath("/Users/shikuanxu/go/src/github.com/sillyhatxu/user-backend/db/migration"))
//	//var Client, err = NewMysqlClient(dataSourceName, DDLPath("/Users/cookie/go/gopath/src/github.com/sillyhatxu/mini-mq/db/migration"))
//	if err != nil {
//		panic(err)
//	}
//	array, err := Client.SchemaVersionArray()
//	assert.Nil(t, err)
//	for _, sv := range array {
//		logrus.Infof("%#v; time : %v", sv, sv.CreatedTime.UnixNano()/int64(time.Millisecond))
//	}
//}
