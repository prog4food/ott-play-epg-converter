package xml_importer

import (
	"os"

	"github.com/rs/zerolog/log"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"

	"ott-play-epg-converter/internal/app_config"
	"ott-play-epg-converter/internal/helpers"
)


var db_name_epg = "epgcache.tmp"
var db_name_ch  = "chcache.db"
var db_flags = sqlite.OpenFlags(0b1000000001000110)
  // ^ SQLITE_OPEN_READWRITE|SQLITE_OPEN_CREATE|SQLITE_OPEN_URI|SQLITE_OPEN_NOMUTEX

func SeedDB() *sqlite.Conn {
  var err error
  var db *sqlite.Conn

  db, err = sqlite.OpenConn(db_name_ch, db_flags); if err != nil {
    log.Panic().Err(err).Send()
  }

  // Check schema
  if err = sqlitex.ExecTransient(db, "SELECT COUNT(*) FROM _dbver_2;", nil); err != nil {
    log.Info().Msg("chdb: create...")
    err = helpers.SimpleExec(db, `
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
      PRAGMA foreign_keys = on;`, "chdb: create error");
      if err != nil { panic(nil) }
  }
  return db
}

// Создание чистой внешней epg db и ее подключение
func AttachEPG() *sqlite.Conn {
  var err error
  var db *sqlite.Conn

  _db_name_epg := db_name_epg
  if app_config.Args.MemDb {
    _db_name_epg = ":memory:"
  } else {
    f, err := os.Create(_db_name_epg)
    if err != nil { log.Panic().Err(err).Send() }
    f.Close()
  }

  db, err = sqlite.OpenConn(db_name_ch, db_flags);
    if err != nil { log.Panic().Err(err).Msg("epgdb: cannot open connection") }

  err = sqlitex.ExecTransient(db, "ATTACH '" + _db_name_epg + "' AS epg;", nil);
    if err != nil { log.Panic().Err(err).Msg("epgdb: attach error") }
  // Seed EPG database
  err = helpers.SimpleExec(db,`
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
    PRAGMA foreign_keys = on;`, "epgdb: create error");
    if err != nil { panic(nil) }

  // Database tune
  helpers.SimpleExec(db, "PRAGMA epg.journal_mode = OFF; PRAGMA epg.synchronous = OFF;", "epgdb: tune error"); // Раньше был TRUNCATE
  return db
}
