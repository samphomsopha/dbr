package dbr

import (
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterpolateNil(t *testing.T) {
	args := []interface{}{nil}

	str, err := Interpolate("SELECT * FROM x WHERE a = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = NULL")
}

func TestInterpolateInts(t *testing.T) {
	args := []interface{}{
		int(1),
		int8(-2),
		int16(3),
		int32(4),
		int64(5),
		uint(6),
		uint8(7),
		uint16(8),
		uint32(9),
		uint64(10),
	}

	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ? AND c = ? AND d = ? AND e = ? AND f = ? AND g = ? AND h = ? AND i = ? AND j = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 1 AND b = -2 AND c = 3 AND d = 4 AND e = 5 AND f = 6 AND g = 7 AND h = 8 AND i = 9 AND j = 10")
}

func TestInterpolateBools(t *testing.T) {
	args := []interface{}{true, false}

	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 1 AND b = 0")
}

func TestInterpolateFloats(t *testing.T) {
	args := []interface{}{float32(0.15625), float64(3.14159)}

	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 0.15625 AND b = 3.14159")
}

func TestInterpolateStrings(t *testing.T) {
	args := []interface{}{"hello", "\"hello's \\ world\" \n\r\x00\x1a"}

	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 'hello' AND b = '\\\"hello\\'s \\\\ world\\\" \\n\\r\\x00\\x1a'")
}

func TestInterpolateSlices(t *testing.T) {
	args := []interface{}{[]int{1}, []int{1, 2, 3}, []uint32{5, 6, 7}, []string{"wat", "ok"}}

	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ? AND c = ? AND d = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = (1) AND b = (1,2,3) AND c = (5,6,7) AND d = ('wat','ok')")
}

type myString struct {
	Present bool
	Val     string
}

func (m myString) Value() (driver.Value, error) {
	if m.Present {
		return m.Val, nil
	} else {
		return nil, nil
	}
}

func TestIntepolatingValuers(t *testing.T) {
	args := []interface{}{myString{true, "wat"}, myString{false, "fry"}}

	str, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", args)
	assert.NoError(t, err)
	assert.Equal(t, str, "SELECT * FROM x WHERE a = 'wat' AND b = NULL")
}

func TestInterpolateErrors(t *testing.T) {
	_, err := Interpolate("SELECT * FROM x WHERE a = ? AND b = ?", []interface{}{1})
	assert.Equal(t, err, ErrArgumentMismatch)

	_, err = Interpolate("SELECT * FROM x WHERE", []interface{}{1})
	assert.Equal(t, err, ErrArgumentMismatch)

	_, err = Interpolate("SELECT * FROM x WHERE a = ?", []interface{}{string([]byte{0x34, 0xFF, 0xFE})})
	assert.Equal(t, err, ErrNotUTF8)

	_, err = Interpolate("SELECT * FROM x WHERE a = ?", []interface{}{struct{}{}})
	assert.Equal(t, err, ErrInvalidValue)

	_, err = Interpolate("SELECT * FROM x WHERE a = ?", []interface{}{[]struct{}{struct{}{}, struct{}{}}})
	assert.Equal(t, err, ErrInvalidSliceValue)
}

func TestCommonSQLInjections(t *testing.T) {
	s := createRealSessionWithFixtures()

	// Grab the last given id
	maxID, err := s.Select("MAX(id)").From("dbr_people").ReturnInt64()
	assert.NoError(t, err)

	for _, injectionAttempt := range strings.Split(InjectionAttempts, "\n") {
		// Try to create a user with the attempted injection as the email address
		res, err := s.
			InsertInto("dbr_people").
			Columns("name", "email").
			Values("A. User", injectionAttempt).
			Exec()
		assert.NoError(t, err)

		// Ensure we created a row a new id
		id, err := res.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, id, maxID+1)
		maxID++

		// SELECT the email back and ensure it's equal to the injection attempt
		var email string
		err = s.Select("email").From("dbr_people").Where("id = ?", id).LoadValue(&email)
		assert.Equal(t, injectionAttempt, email)
	}
}

// InjectionAttempts is a newline separated list of common SQL injection exploits
// taken from https://wfuzz.googlecode.com/svn/trunk/wordlist/Injections/SQL.txt
var InjectionAttempts = `
'
"
#
-
--
'%20--
--';
'%20;
=%20'
=%20;
=%20--
\x23
\x27
\x3D%20\x3B'
\x3D%20\x27
\x27\x4F\x52 SELECT *
\x27\x6F\x72 SELECT *
'or%20select *
admin'--
<>"'%;)(&+
'%20or%20''='
'%20or%20'x'='x
"%20or%20"x"="x
')%20or%20('x'='x
0 or 1=1
' or 0=0 --
" or 0=0 --
or 0=0 --
' or 0=0 #
" or 0=0 #
or 0=0 #
' or 1=1--
" or 1=1--
' or '1'='1'--
"' or 1 --'"
or 1=1--
or%201=1
or%201=1 --
' or 1=1 or ''='
" or 1=1 or ""="
' or a=a--
" or "a"="a
') or ('a'='a
") or ("a"="a
hi" or "a"="a
hi" or 1=1 --
hi' or 1=1 --
hi' or 'a'='a
hi') or ('a'='a
hi") or ("a"="a
'hi' or 'x'='x';
@variable
,@variable
PRINT
PRINT @@variable
select
insert
as
or
procedure
limit
order by
asc
desc
delete
update
distinct
having
truncate
replace
like
handler
bfilename
' or username like '%
' or uname like '%
' or userid like '%
' or uid like '%
' or user like '%
exec xp
exec sp
'; exec master..xp_cmdshell
'; exec xp_regread
t'exec master..xp_cmdshell 'nslookup www.google.com'--
--sp_password
\x27UNION SELECT
' UNION SELECT
' UNION ALL SELECT
' or (EXISTS)
' (select top 1
'||UTL_HTTP.REQUEST
1;SELECT%20*
to_timestamp_tz
tz_offset
&lt;&gt;&quot;'%;)(&amp;+
'%20or%201=1
%27%20or%201=1
%20$(sleep%2050)
%20'sleep%2050'
char%4039%41%2b%40SELECT
&apos;%20OR
'sqlattempt1
(sqlattempt2)
|
%7C
*|
%2A%7C
*(|(mail=*))
%2A%28%7C%28mail%3D%2A%29%29
*(|(objectclass=*))
%2A%28%7C%28objectclass%3D%2A%29%29
(
%28
)
%29
&
%26
!
%21
' or 1=1 or ''='
' or ''='
x' or 1=1 or 'x'='y
/
//
//*
*/*
`
