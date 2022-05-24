package xml_importer

import (
	"crawshaw.io/sqlite"
	"github.com/rs/zerolog/log"
)

type PreQueries struct {
  ChOk        bool
  Ch_data    *sqlite.Stmt
  Ch_names   *sqlite.Stmt
  Ch_ids     *sqlite.Stmt
  Ch_icons   *sqlite.Stmt
  Epg_data   *sqlite.Stmt
  Epg_title  *sqlite.Stmt
  Epg_desc   *sqlite.Stmt
  Epg_icon   *sqlite.Stmt
}

var DbPre PreQueries

func InsertCh(stmt *sqlite.Stmt, prov,ch uint32, name,icon uint64)  {
  stmt.BindInt64(1, int64(prov))
  stmt.BindInt64(2, int64(ch))
  stmt.BindInt64(3, int64(name))
  stmt.BindInt64(4, int64(icon))
  insert_exec(stmt)
}

func InsertKV[T uint32|uint64](stmt *sqlite.Stmt, h T, s string)  {
  stmt.BindInt64(1, int64(h))
  stmt.BindText (2, s)
  insert_exec(stmt)
}

func insert_exec(stmt *sqlite.Stmt)  {
  var err error
  _, err = stmt.Step()
    if err != nil { log.Err(err).Msg("epg insert: Insert error") }

  err = stmt.Reset();
    if err != nil { log.Error().Msg("epg insert: Reset error") }
  err = stmt.ClearBindings()
    if err != nil { log.Error().Msg("epg insert: ClearBindings error") }
}

func PrecompileQuery(ch_tx *sqlite.Conn, epg_tx *sqlite.Conn)  {
  if !DbPre.ChOk {
    DbPre.Ch_data  = prepare_query(ch_tx, "insert into ch_data values(?, ?, ?, ?)")
    DbPre.Ch_names = prepare_query(ch_tx, "insert into h_ch_names values(?, ?)")
    DbPre.Ch_ids   = prepare_query(ch_tx, "insert into h_ch_ids values(?, ?)")
    DbPre.Ch_icons = prepare_query(ch_tx, "insert into h_ch_icons values(?, ?)")
    DbPre.ChOk     = true
  }

  DbPre.Epg_data  = prepare_query(epg_tx, "insert into epg.temp_data values(?, ?, ?, ?, ?, ?, ?)")
  DbPre.Epg_title = prepare_query(epg_tx, "insert into epg.h_title values(?, ?)")
  DbPre.Epg_desc  = prepare_query(epg_tx, "insert into epg.h_desc values(?, ?)")
  DbPre.Epg_icon  = prepare_query(epg_tx, "insert into epg.h_icon values(?, ?)")
}

func FinalizeEpgQuery()  {
  var err error
  errfunc := func() {
    log.Panic().Err(err).Msg("query finalize error")
  }
  err = DbPre.Epg_data.Finalize()
    if err != nil { errfunc() }
  err = DbPre.Epg_title.Finalize()
    if err != nil { errfunc() }
  err = DbPre.Epg_desc.Finalize()
    if err != nil { errfunc() }
  err = DbPre.Epg_icon.Finalize()
    if err != nil { errfunc() }
}

// Компиляция SQL запроса
func prepare_query(tx *sqlite.Conn, q string) *sqlite.Stmt {
  stmt, err := tx.Prepare(q)
  if err != nil {
    log.Panic().Err(err).Msgf ("prepare error q: %s", q)
  }
  return stmt
}
