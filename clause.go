package dbr

type clause struct {
	Sql    string
	Values []interface{}
}

type clauses struct {
	cl []*clause
}

func Clause(sql string, values ...interface{}) *clauses {
	return &clauses{cl: []*clause{{Sql: sql, Values: values}}}
}

func (c *clauses) Or(sql string, values ...interface{}) *clauses {
	c.cl = append(c.cl, &clause{Sql: sql, Values: values})
	return c
}
