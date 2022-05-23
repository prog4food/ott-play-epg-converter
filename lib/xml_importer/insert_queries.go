package xml_importer

import (
	"database/sql"

	"github.com/rs/zerolog/log"
)

type PreQueries struct {
  Ch_data    *sql.Stmt
  Ch_names   *sql.Stmt
  Ch_ids     *sql.Stmt
  Ch_icons   *sql.Stmt
  Epg_data   *sql.Stmt
  Epg_title  *sql.Stmt
  Epg_desc   *sql.Stmt
  Epg_icon   *sql.Stmt
}

var DbPre PreQueries

func PrecompileQuery(ch_tx *sql.Tx, epg_tx *sql.Tx)  {
  DbPre.Ch_data  = prepare_query(ch_tx, "insert into ch_data values(?, ?, ?, ?)")
  DbPre.Ch_names = prepare_query(ch_tx, "insert into h_ch_names values(?, ?)")
  DbPre.Ch_ids   = prepare_query(ch_tx, "insert into h_ch_ids values(?, ?)")
  DbPre.Ch_icons = prepare_query(ch_tx, "insert into h_ch_icons values(?, ?)")

  DbPre.Epg_data  = prepare_query(epg_tx, "insert into epg.temp_data values(?, ?, ?, ?, ?, ?, ?)")
  DbPre.Epg_title = prepare_query(epg_tx, "insert into epg.h_title values(?, ?)")
  DbPre.Epg_desc  = prepare_query(epg_tx, "insert into epg.h_desc values(?, ?)")
  DbPre.Epg_icon  = prepare_query(epg_tx, "insert into epg.h_icon values(?, ?)")
}

// Компиляция SQL запроса
func prepare_query(tx *sql.Tx, q string) *sql.Stmt {
  stmt, err := tx.Prepare(q)
  if err != nil {
    log.Panic().Err(err).Send()
  }
  return stmt
}
