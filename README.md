# mysql-client

## Config

```
const (
	dataSourceName = `sillyhat:sillyhat@tcp(127.0.0.1:3308)/sillyhat`
	maxIdleConns   = 5
	maxOpenConns   = 10
)
```

## SQL script

```
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
		select id,name, age, TIMESTAMP(birthday) birthday, description, (is_delete = b'1') is_delete, created_date, last_modified_date from userinfo
	`

	findOne_sql = `
		select id,name, age, TIMESTAMP(birthday) birthday, description, (is_delete = b'1') is_delete, created_date, last_modified_date from userinfo where id = 
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

```

## Insert One

```
func TestClientInsert(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	id, err := Client.Insert(insert_sql, "test name", 21, "1989-06-09", "This is description", false)
	log.Info("id : ", id)
	assert.Nil(t, err)
	assert.EqualValues(t, id, 2)
}
```

## Insert Batch

```
func TestClientBatchInsert(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	result, err := Client.BatchInsert(func(tx *sql.Tx) (int, error) {
		totalCount := 0
		for i := 1001; i <= 2000; i++ {
			_, err := tx.Exec(insert_sql, "test name"+strconv.Itoa(i), 21, "1989-06-09", "This is description", false)
			assert.Nil(t, err)
			totalCount++
		}
		return totalCount, nil
	})
	assert.Nil(t, err)
	assert.EqualValues(t, result, 1000)
}
```

## Update One

```
func TestClientUpdate(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	count, err := Client.Update(update_sql, "test name update", 22, "2000-01-01", "This is update result", true, 2)
	log.Info("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 1)
}
```

## Update Batch

```
func TestClientBatchUpdate(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	result, err := Client.BatchUpdate(func(tx *sql.Tx) (int, error) {
		totalCount := 0
		for i := 3; i <= 1002; i++ {
			_, err := tx.Exec(update_sql, "test update name -"+strconv.Itoa(i), 21, "2005-01-30", "This is update", true, i)
			assert.Nil(t, err)
			totalCount++
		}
		return totalCount, nil
	})
	assert.Nil(t, err)
	assert.EqualValues(t, result, 1000)
}

```

## Delete One

```
func TestClientDeleteOne(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	count, err := Client.DeleteByPrimaryKey(deleteOne_sql, 3)
	log.Info("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 1)
}
```

## Delete

```
func TestClientDelete(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	count, err := Client.Delete(delete_sql, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23)
	log.Info("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 10)
}
```

## Find One

```
func TestClientFindOne(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	result, err := Client.FindOne(findOne_sql,2)
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
```

## Find

```
func TestClientFind(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	results, err := Client.Find(findAll_sql,21, true, "%update name%")
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
	assert.EqualValues(t, len(userArray), 2)
	assert.EqualValues(t, userArray[0].Id, 1)
	assert.EqualValues(t, userArray[1].Id, 2)
}

```

## Count

```
func TestClientCount(t *testing.T) {
	InitialDBClient(dataSourceName, 5, 10)
	count, err := Client.Count(count_sql)
	log.Info("count : ", count)
	assert.Nil(t, err)
	assert.EqualValues(t, count, 1002)
}
```
