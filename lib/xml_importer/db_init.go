package xml_importer

import (
	"database/sql"
	"os"
	"ott-play-epg-converter/lib/app_config"

	"github.com/rs/zerolog/log"
)


var EpgTempDb = "epgcache.tmp"


type DbTxInterface interface {
  Query(query string, args ...interface{}) (*sql.Rows, error)
  Exec(query string, args ...interface{}) (sql.Result, error)
}

func PrintAttachedDB(db DbTxInterface, s string)  {
  rows, err := db.Query("PRAGMA database_list;")
  if err != nil { log.Err(err).Send() }
  for rows.Next() {
    var id int
    var name string
    var file string
    err = rows.Scan(&id, &name, &file)
    if err != nil { log.Err(err).Send() }
    log.Printf("Attached[%s]: %s - %s", s ,name, file)
  }
}


func SeedDB(dbname string) *sql.DB {
  var err error
  var db *sql.DB
  db, err = sql.Open("sqlite3", dbname); if err != nil {
    log.Panic().Err(err).Send()
  }
  // Check schema
  if _, err = db.Exec("SELECT COUNT(*) FROM _dbver_2;"); err != nil {
    log.Info().Msg("Creating channel cache database...")
    if _, err = db.Exec(`
    PRAGMA foreign_keys = off;
    BEGIN TRANSACTION;
    -- Таблица: _dbver_2
    DROP TABLE IF EXISTS _dbver_2;
    CREATE TABLE _dbver_2 (nop INT8);
    -- Таблица: ch_data
    DROP TABLE IF EXISTS ch_data;
    CREATE TABLE ch_data (h_prov_id INTEGER NOT NULL, h_id INTEGER NOT NULL, h_name BIGINT NOT NULL, h_icon BIGINT, PRIMARY KEY (h_prov_id, h_id, h_name) ON CONFLICT REPLACE);
    -- Таблица: h_ch_icons
    DROP TABLE IF EXISTS h_ch_icons;
    CREATE TABLE h_ch_icons (h BIGINT PRIMARY KEY ON CONFLICT IGNORE NOT NULL, data STRING);
    -- Таблица: h_ch_ids
    DROP TABLE IF EXISTS h_ch_ids;
    CREATE TABLE h_ch_ids (h BIGINT PRIMARY KEY ON CONFLICT IGNORE NOT NULL, data STRING);
    -- Таблица: h_ch_names
    DROP TABLE IF EXISTS h_ch_names;
    CREATE TABLE h_ch_names (h BIGINT PRIMARY KEY ON CONFLICT IGNORE NOT NULL, data STRING);
    COMMIT TRANSACTION;
    PRAGMA foreign_keys = on;
    `); err != nil {
      log.Panic().Err(err).Send()
    }
  }
  return db
}

// Создание чистой внешней epg db и ее подключение
func AttachEPG(db *sql.DB) {
  var err error
  if app_config.AppConfig.MemDb {
    EpgTempDb = ":memory:"
  } else {
    if _, err = os.Stat(EpgTempDb); err == nil {
      if err = os.Remove(EpgTempDb); err != nil {
        log.Panic().Err(err).Send()  
      }
    }
  }
  if _, err = db.Exec("ATTACH '" + EpgTempDb + "' AS epg;"); err != nil {
    log.Panic().Err(err).Msg("Attach error")
  }
  // Seed EPG database
  if _, err = db.Exec(`
    PRAGMA foreign_keys = off;
    BEGIN TRANSACTION;
    -- Таблица: epg.temp_data
    DROP TABLE IF EXISTS epg.temp_data;
    CREATE TABLE epg.temp_data (h_prov_id INTEGER NOT NULL, h_ch_id INTEGER NOT NULL, t_start BIGINT NOT NULL, t_stop BIGINT NOT NULL, h_title BIGINT NOT NULL, h_desc BIGINT, h_icon BIGINT, PRIMARY KEY (h_prov_id, h_ch_id, t_start, t_stop) ON CONFLICT REPLACE);
    -- Таблица: epg.h_desc
    DROP TABLE IF EXISTS epg.h_desc;
    CREATE TABLE epg.h_desc (h BIGINT PRIMARY KEY ON CONFLICT IGNORE NOT NULL, data STRING);
    -- Таблица: epg.h_icon
    DROP TABLE IF EXISTS epg.h_icon;
    CREATE TABLE epg.h_icon (h BIGINT PRIMARY KEY ON CONFLICT IGNORE NOT NULL, data STRING);
    -- Таблица: epg.h_title
    DROP TABLE IF EXISTS epg.h_title;
    CREATE TABLE epg.h_title (h BIGINT PRIMARY KEY ON CONFLICT IGNORE NOT NULL, data STRING);
    COMMIT TRANSACTION;
    PRAGMA foreign_keys = on;
    PRAGMA epg.synchronous = OFF;
    PRAGMA epg.journal_mode = TRUNCATE;
    --PRAGMA epg.journal_mode = WAL;
  `); err != nil {
    log.Panic().Err(err).Msg("Seed error")
  }
}

// Отключение внешней epg db
func DetachEPG(maindb *sql.DB) {
  var err error
  if _, err = maindb.Exec("DETACH epg;"); err != nil {
    log.Err(err).Msg("DETACH error")
  }
}
