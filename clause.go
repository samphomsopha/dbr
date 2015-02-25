package dbr

type clause struct {
	Sql    string
	Values []interface{}
}

type clauses struct {
	cl []*clause
}

func Clause(sql string, values ...interface{}) *clauses {
	var c []*clause
	c = append(c, &clause{Sql: sql, Values: values})
	return &clauses{cl: c}
}

func (c *clauses) Or(sql string, values ...interface{}) *clauses {
	c.cl = append(c.cl, &clause{Sql: sql, Values: values})
	return c
}
