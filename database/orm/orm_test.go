package orm

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/database"
)

var (
	tmpDir = "."
)

func init() {
	dir := os.TempDir()
	if _, err := os.Stat(dir); err == nil {
		tmpDir = dir
	}
}

type Model struct {
	Id int64 `orm:"column=entity_id; primary; auto_increment"`
}

type User struct {
	Model
	Email    string
	Password string
	Posts    []*Post
}

type Post struct {
	Model     `orm:"table=blog_posts"`
	UserId    int64
	Name      string
	Title     string
	Content   string
	Created   *time.Time
	Modified  *time.Time
	Status    int
	User      *User      `orm:"foreign_key=user_id"`
	Comments  []*Comment `orm:"foreign_key=post_id"`
	Something int        `orm:"-"`
}

type Comment struct {
	Model
	PostId  int64
	Content string
	Post    *Post `orm:"foreign_key=post_id"`
}

func TestParseFieldTag(t *testing.T) {
	db := "test_db"
	table := "test_table"
	column := "entity_id"
	primary := true

	s := fmt.Sprintf(
		`database=%s; table = %s ; column =%s;nullable; primary ;;`,
		db, table, column,
	)

	fi := ParseFieldTag(s)

	if fi.Database != db {
		t.Error("database:", fi.Database, "!=", db)
	}

	if fi.Table != table {
		t.Error("table:", fi.Table, "!=", table)
	}

	if fi.Column != column {
		t.Error("column:", fi.Column, "!=", db)
	}

	if fi.Primary != primary {
		t.Error("primary:", fi.Primary, "!=", primary)
	}
}

func TestModelInfo(t *testing.T) {
	mi := GetModelInfo(&Post{})

	fields := []string{"entity_id", "user_id", "name", "title", "content", "created", "modified", "status"}
	if len(mi.Fields) != len(fields) {
		t.Error("fields:", "lengths do not match")
		return
	}
	for i, fi := range mi.Fields {
		if fi.Column != fields[i] {
			t.Error("fields:", fi.Column, "!=", fields[i])
		}
	}

	primary := []string{"entity_id"}
	if len(mi.Primary) != len(primary) {
		t.Error("primary:", "lengths do not match")
		return
	}
	for i, fi := range mi.Primary {
		if fi.Column != primary[i] {
			t.Error("primary:", fi.Column, "!=", primary[i])
		}
	}

	belongsTo := map[string]*Relation{
		"User": {
			Table: "users",
			Key: FieldInfoList{
				&FieldInfo{
					Column: "entity_id",
				},
			},
			Reference: FieldInfoList{
				&FieldInfo{
					Column: "user_id",
				},
			},
		},
	}
	if len(mi.BelongsTo) != len(belongsTo) {
		t.Error("belongsTo:", "lengths do not match")
		return
	}
	for k, relation := range belongsTo {
		mRelation, ok := mi.BelongsTo[k]
		if !ok {
			t.Error("belongsTo:", "not found:", k)
			continue
		}

		err := compareRelations(mRelation, relation)
		if err != nil {
			t.Error("belongsTo:", k+":", err)
		}
	}

	hasMany := map[string]*Relation{
		"Comments": {
			Table: "comments",
			Key: FieldInfoList{
				&FieldInfo{
					Column: "post_id",
				},
			},
			Reference: FieldInfoList{
				&FieldInfo{
					Column: "entity_id",
				},
			},
		},
	}
	if len(mi.HasMany) != len(hasMany) {
		t.Error("hasMany:", "lengths do not match")
		return
	}
	for k, relation := range hasMany {
		mRelation, ok := mi.HasMany[k]
		if !ok {
			t.Error("hasMany:", "not found:", k)
			continue
		}

		err := compareRelations(mRelation, relation)
		if err != nil {
			t.Error("hasMany:", k+":", err)
		}
	}
}

func TestSave(t *testing.T) {
	f, err := initDatabase()
	if err != nil {
		t.Fatal(err)
		return
	}

	db, err := database.Get(database.DefaultDB)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = createTablePosts(db)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Save A
	postA := &Post{
		UserId:  1,
		Name:    "hello",
		Title:   "Hello, world!",
		Content: "Hello, world!  \nThis is my first post :)",
		Status:  1,
	}

	err = Save(postA)
	if err != nil {
		t.Fatal(err)
	}

	// Save B
	postB := &Post{
		UserId:  1,
		Name:    "test",
		Title:   "Test",
		Content: "This is a test.",
		Status:  1,
	}

	err = Save(postB)
	if err != nil {
		t.Fatal(err)
	}

	// Load B
	postX := new(Post)
	postX.Id = postB.Id

	err = First(postX)
	if err != nil {
		t.Fatal(err)
	}

	if postX.Id != postB.Id {
		t.Errorf(
			"Id: %d != %d",
			postX.Id, postB.Id,
		)
	}
	if postX.Name != postB.Name {
		t.Errorf(
			"Name: %s != %s",
			strconv.Quote(postX.Name), strconv.Quote(postB.Name),
		)
	}

	f.Close()
	os.Remove(f.Name())
}

func compareRelations(a *Relation, b *Relation) error {
	if a.Table != b.Table {
		return fmt.Errorf("%s != %s", a.Table, b.Table)
	}

	if len(a.Key) != len(b.Key) {
		return fmt.Errorf("key: lengths do not match")
	}
	for i, fi := range a.Key {
		if fi.Column != b.Key[i].Column {
			return fmt.Errorf("key: %s != %s", fi.Column, b.Key[i].Column)
		}
	}

	if len(a.Reference) != len(b.Reference) {
		return fmt.Errorf("reference: lengths do not match")
	}
	for i, fi := range a.Reference {
		if fi.Column != b.Reference[i].Column {
			return fmt.Errorf("reference: %s != %s", fi.Column, b.Reference[i].Column)
		}
	}

	return nil
}

func initDatabase() (*os.File, error) {
	f, err := ioutil.TempFile(tmpDir, "test.*.db")
	if err != nil {
		return nil, err
	}

	config.Set("database", map[string]interface{}{
		database.DefaultDB: map[string]interface{}{
			"driver": "sqlite3",
			"file":   f.Name(),
			"params": map[string]interface{}{},
		},
	})

	return f, nil
}

func createTablePosts(db *database.DB) error {
	_, err := db.Exec(`
		CREATE TABLE "blog_posts" (
			"entity_id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"user_id"   INTEGER,
			"name"      TEXT,
			"title"     TEXT,
			"content"   TEXT,
			"created"   INTEGER,
			"modified"  INTEGER,
			"status"    INTEGER
		)
	`)
	return err
}
