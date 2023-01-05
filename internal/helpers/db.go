package helpers

import (
	"strconv"
	"strings"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/rs/zerolog/log"
)

// Простой запрос
func QueryOneS(db *sqlite.Conn, q string) {
  fn := func(stmt *sqlite.Stmt) error {
    s := make([]string, 0, 5)
    for i := 0; i < stmt.ColumnCount(); i++ {
      if stmt.ColumnType(i) == sqlite.SQLITE_TEXT {
        s = append(s, stmt.ColumnText(i))
      } else if stmt.ColumnType(i) == sqlite.SQLITE_INTEGER {
        s = append(s, strconv.Itoa(stmt.ColumnInt(i)))
      }
    }
    log.Print("... ", strings.Join(s, " -- "))
    return nil
  }
  if err := sqlitex.ExecTransient(db, q, fn); err != nil {
    log.Err(err).Msg("query error")
  }
}


// Быстрый запрос в базу, не получающий результат 
// основан на ExecScript
func SimpleExec(db *sqlite.Conn, q,err_msg string) error{
  var trailingBytes int
  var err error

  for {
    //if q == "" { break }
    var stmt *sqlite.Stmt
    stmt, trailingBytes, err = db.PrepareTransient(q)
      if err != nil { log.Err(err).Msg(err_msg); return err }
    _, err = stmt.Step()
    stmt.Finalize()
      if err != nil { log.Err(err).Msg(err_msg) }
    if trailingBytes == 0 { break }
    //q = strings.TrimSpace(q[len(q) - trailingBytes:])
    q = q[len(q) - trailingBytes:] // Быстрее, но последняя строка не должна содержать пробелов
  }
  return nil
}