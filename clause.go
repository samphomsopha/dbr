package dbr

type clause struct {
	Sql    string
	Values []interface{}
	isOr   bool
}

type clauses struct {
	cl []*clause
}

func Clause(sql string, values ...interface{}) *clauses {
	return &clauses{cl: []*clause{{Sql: sql, Values: values, isOr: true}}}
}

func (c *clauses) Or(sql string, values ...interface{}) *clauses {
	c.cl = append(c.cl, &clause{Sql: sql, Values: values, isOr: true})
	return c
}

func (c *clauses) And(sql string, values ...interface{}) *clauses {
	c.cl = append(c.cl, &clause{Sql: sql, Values: values, isOr: false})
	return c
}
