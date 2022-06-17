package query

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/olegshs/go-tools/database/drivers/postgres"
	"github.com/olegshs/go-tools/helpers"
)

var (
	helper = new(postgres.Helper)
)

func TestQuery_Select(t *testing.T) {
	q := New(nil, helper).Select(
		"p.id", "p.name", Expr("COUNT(c.id)"),
	).From(
		"posts p",
	).LeftJoin(
		"comments c",
		Eq{"c.post_id": Column("p.id")},
	).Where(
		Or{
			Eq{"name": "hello"},
			Condition{
				"title",
				"LIKE",
				"Hello%",
			},
		},
		Not{
			Ne{"p.content": nil},
			Between{"p.created": {"2019-12-03", "2020-01-01"}},
		},
		In{"p.status": []interface{}{1, 2, 3}},
	).Group(
		"p.id",
	).Having(
		"COUNT(c.id) >= 1",
	).Order(
		Desc("p.created"),
	).Limit(
		20, 10,
	)

	expected := trimSpace(`
		SELECT "p"."id", "p"."name", COUNT(c.id)
		FROM "posts" "p"
		LEFT JOIN "comments" "c" ON "c"."post_id" = "p"."id"
		WHERE (("name" = $1) OR ("title" LIKE $2)) AND (!(("p"."content" IS NOT NULL) AND ("p"."created" BETWEEN $3 AND $4))) AND ("p"."status" IN ($5, $6, $7))
		GROUP BY "p"."id"
		HAVING COUNT(c.id) >= 1
		ORDER BY "p"."created" DESC
		LIMIT $8 OFFSET $9
	`)
	if q.String() != expected {
		t.Error("expected:", "\n"+expected)
		t.Error("got:", "\n"+q.String())
	}

	expectedArgs := []interface{}{
		"hello", "Hello%", "2019-12-03", "2020-01-01", 1, 2, 3, 10, 20,
	}
	err := compareArgs(q.Args(), expectedArgs)
	if err != nil {
		t.Error(err)
	}
}

func TestQuery_SelectSubquery(t *testing.T) {
	q := New(nil, helper).Select(
		"id",
		"name",
		New(nil, helper).Select(
			Expr("COUNT(*)"),
		).From(
			"comments",
		).Where(
			Eq{"post_id": Column("posts.id")},
		).As(
			"comments_count",
		),
	).From(
		"posts",
	)

	expected := trimSpace(`
		SELECT "id", "name", (SELECT COUNT(*)
		FROM "comments"
		WHERE "post_id" = "posts"."id") AS "comments_count"
		FROM "posts"
	`)
	if q.String() != expected {
		t.Error("expected:", "\n"+expected)
		t.Error("got:", "\n"+q.String())
	}
}

func TestQuery_SelectCompositeEq(t *testing.T) {
	q := New(nil, helper).Select(
		"a", "b", "c",
	).From(
		"test",
	).Where(
		CompositeCondition{
			Columns:  []string{"a", "b"},
			Operator: "=",
			Values: []interface{}{
				1, 2,
			},
		},
	)

	expected := trimSpace(`
        SELECT "a", "b", "c"
        FROM "test"
        WHERE ("a", "b") = ($1, $2)
    `)
	if q.String() != expected {
		t.Error("expected:", "\n"+expected)
		t.Error("got:", "\n"+q.String())
	}

	expectedArgs := []interface{}{
		1, 2,
	}
	err := compareArgs(q.Args(), expectedArgs)
	if err != nil {
		t.Error(err)
	}
}

func TestQuery_SelectCompositeIn(t *testing.T) {
	q := New(nil, helper).Select(
		"a", "b", "c",
	).From(
		"test",
	).Where(
		CompositeCondition{
			Columns:  []string{"a", "b"},
			Operator: "IN",
			Values: []interface{}{
				[]interface{}{1, 1},
				[]interface{}{1, 2},
			},
		},
	)

	expected := trimSpace(`
        SELECT "a", "b", "c"
        FROM "test"
        WHERE ("a", "b") IN (($1, $2), ($3, $4))
    `)
	if q.String() != expected {
		t.Error("expected:", "\n"+expected)
		t.Error("got:", "\n"+q.String())
	}

	expectedArgs := []interface{}{
		1, 1, 1, 2,
	}
	err := compareArgs(q.Args(), expectedArgs)
	if err != nil {
		t.Error(err)
	}
}

func TestQuery_Insert(t *testing.T) {
	q := New(nil, helper).Insert("posts", Data{
		"name":  "hello",
		"title": "Hello, world!",
	})

	expected := trimSpace(`
		INSERT INTO "posts"
		("name", "title")
		VALUES ($1, $2)
	`)
	if q.String() != expected {
		t.Error("expected:", "\n"+expected)
		t.Error("got:", "\n"+q.String())
	}

	expectedArgs := []interface{}{
		"hello", "Hello, world!",
	}
	err := compareArgs(q.Args(), expectedArgs)
	if err != nil {
		t.Error(err)
	}
}

func TestQuery_Update(t *testing.T) {
	q := New(nil, helper).Update("posts", Data{
		"name":   "test",
		"status": 1,
	}).Where(Gte{
		"status": 2,
	}).Order(Order{
		"id", "ASC",
	}).Limit(1)

	expected := trimSpace(`
		UPDATE "posts"
		SET "name" = $1, "status" = $2
		WHERE "status" >= $3
		ORDER BY "id" ASC
		LIMIT $4
	`)
	if q.String() != expected {
		t.Error("expected:", "\n"+expected)
		t.Error("got:", "\n"+q.String())
	}

	expectedArgs := []interface{}{
		"test", 1, 2, 1,
	}
	err := compareArgs(q.Args(), expectedArgs)
	if err != nil {
		t.Error(err)
	}
}

func TestQuery_Delete(t *testing.T) {
	q := New(nil, helper).Delete("posts").Where(In{
		"status": []interface{}{
			0, 1,
		},
	}).Order(Order{
		"id", "ASC",
	}).Limit(1)

	expected := trimSpace(`
		DELETE FROM "posts"
		WHERE "status" IN ($1, $2)
		ORDER BY "id" ASC
		LIMIT $3
	`)
	if q.String() != expected {
		t.Error("expected:", "\n"+expected)
		t.Error("got:", "\n"+q.String())
	}

	expectedArgs := []interface{}{
		0, 1, 1,
	}
	err := compareArgs(q.Args(), expectedArgs)
	if err != nil {
		t.Error(err)
	}
}

func trimSpace(s string) string {
	r := regexp.MustCompile(`\n\s+`)
	s = r.ReplaceAllString(s, "\n")
	s = helpers.Trim(s)
	return s
}

func compareArgs(a, b []interface{}) error {
	if len(a) != len(b) {
		return fmt.Errorf("argument lengths do not match")
	}

	for k, v := range a {
		exp := b[k]
		if v != exp {
			return fmt.Errorf("argument[%d]: %v != %v", k, v, exp)
		}
	}

	return nil
}
