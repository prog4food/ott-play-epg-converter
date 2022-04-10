package main

import (
  "regexp"
  "os"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
  "github.com/rs/zerolog/log"
  "ott-play-epg-converter/import/robbiet480/xmltv"
  "ott-play-epg-converter/lib/arg-reader"
  "ott-play-epg-converter/lib/string-hashes"
)

var (
  sting_flater = regexp.MustCompile(`\r?\n`)
)

func SeedDB(dbname string) *sql.DB {
  var err error
  var db *sql.DB
  db, err = sql.Open("sqlite3", dbname); if err != nil {
    log.Panic().Err(err).Send()
  }
  // Check schema
  if _, err = db.Exec("SELECT COUNT(*) FROM _dbver_1;"); err != nil {
    log.Info().Msg("Creating channel cache database...")
    if _, err = db.Exec(`
    PRAGMA foreign_keys = off;
    BEGIN TRANSACTION;
    -- Таблица: _dbver_1
    DROP TABLE IF EXISTS _dbver_1;
    CREATE TABLE _dbver_1 (nop INT8);
    -- Таблица: ch_data
    DROP TABLE IF EXISTS ch_data;
    CREATE TABLE ch_data (h_prov_id INTEGER NOT NULL, h_id INTEGER NOT NULL, h_name BIGINT NOT NULL, h_icon BIGINT, prov_order INT8, PRIMARY KEY (h_prov_id, h_id, h_name) ON CONFLICT REPLACE);
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

func InitEPG(maindb *sql.DB) {
  var err error
  if _, err = os.Stat("epgcache.tmp"); err == nil {
    if err = os.Remove("epgcache.tmp"); err != nil {
      log.Panic().Err(err).Send()  
    }
  }
  if _, err = maindb.Exec("ATTACH 'epgcache.tmp' AS epg;"); err != nil {
    log.Panic().Err(err).Send()
  }
  // Seed EPG database
  if _, err = maindb.Exec(`
  PRAGMA foreign_keys = off;
  BEGIN TRANSACTION;
  -- Таблица: epg.data
  CREATE TABLE epg.data (h_prov_id INTEGER NOT NULL, h_ch_id INTEGER NOT NULL, h_title BIGINT NOT NULL, h_desc BIGINT, t_start BIGINT NOT NULL, t_stop BIGINT NOT NULL, PRIMARY KEY (h_prov_id, h_ch_id, t_start, t_stop) ON CONFLICT REPLACE);
  -- Таблица: epg.h_desc
  CREATE TABLE epg.h_desc (h BIGINT PRIMARY KEY ON CONFLICT IGNORE NOT NULL, data STRING);
  -- Таблица: epg.h_title
  CREATE TABLE epg.h_title (h BIGINT PRIMARY KEY ON CONFLICT IGNORE NOT NULL, data STRING);
  COMMIT TRANSACTION;
  PRAGMA foreign_keys = on;
  `); err != nil {
    log.Panic().Err(err).Send()
  }
}

func FinallyEPG(maindb *sql.DB) {
  var err error
  if _, err = maindb.Exec("DETACH epg;"); err != nil {
    log.Panic().Err(err).Send()
  }
}

func CleanEPG(db *sql.DB) {
  if _, err := db.Exec("DELETE FROM epg_data; DELETE FROM h_epg_title; DELETE FROM h_epg_desc; VACUUM;"); err != nil {
    log.Panic().Err(err).Send()
  }
}

func PreQuery(tx *sql.Tx, q string) *sql.Stmt {
  stmt, err := tx.Prepare(q)
  if err != nil {
    log.Panic().Err(err).Send()
  }
  return stmt
}

// XML2SQL: Кешируем запись <channel>
// Берем Id, DisplayName[*] и Icon[0]
func NewChannelCache(ch_data *sql.Stmt, ch_ids *sql.Stmt, ch_names *sql.Stmt, ch_icons *sql.Stmt, ch *xmltv.Channel, prov *arg_reader.ProvRecord) {
  var err error
  
  // 2SQL: dedup Id канала
  h_id   := string_hashes.HashSting32(ch.ID)
  if _, err = ch_ids.Exec(h_id, ch.ID); err != nil {
    log.Err(err).Send()
	}
  // 2SQL: dedup Icon
  h_icon := uint64(0)
  if len(ch.Icons) > 0  {
    h_icon = string_hashes.HashSting64(ch.Icons[0].Source)
    if _, err = ch_icons.Exec(h_icon, ch.Icons[0].Source); err != nil {
      log.Err(err).Send()
  	  }
  }
  // Обход <display-name>
  h_name := uint64(0)
  names_len := len(ch.DisplayNames)
  if names_len == 0 {
    log.Error().Msgf("Channel %s has no display names", ch.ID)
    // 2SQL: Связи
    if _, err = ch_data.Exec(prov.IdHash, h_id, h_name, h_icon, prov.Order); err != nil {
      log.Err(err).Send()
  	  }
  }
  for i := 0; i < names_len; i++ {
    // 2SQL: dedup Название
    h_name = string_hashes.HashSting64(ch.DisplayNames[i].Value)
    if _, err = ch_names.Exec(h_name, ch.DisplayNames[i].Value); err != nil {
      log.Err(err).Send()
  		}
    // 2SQL: Связи
    if _, err = ch_data.Exec(prov.IdHash, h_id, h_name, h_icon, prov.Order); err != nil {
      log.Err(err).Send()
  	  }
  }
}


// XML2SQL: Кешируем запись <programme>
// Берем только Title[0] и Desc[0]
func NewProgCache(epg_data *sql.Stmt, epg_title *sql.Stmt, epg_desc *sql.Stmt, pr *xmltv.Programme, prov *arg_reader.ProvRecord) {
  var err error
  // Проверки
  if len(pr.Titles) == 0 || pr.Start == nil || pr.Stop == nil {
    return
  }
  // Хеширование
  h_ch_id := string_hashes.HashSting32(pr.Channel)
  h_desc  := uint64(0)
  // 2SQL: dedup Название
  h_title := string_hashes.HashSting64(pr.Titles[0].Value)
  if _, err = epg_title.Exec(h_title, pr.Titles[0].Value); err != nil {
    log.Err(err).Send()
	}
  // 2SQL: dedup Описание
  if len(pr.Descriptions) > 0  {
    h_desc = string_hashes.HashSting64(pr.Descriptions[0].Value)
    if h_title != h_desc  {
      flat_string := sting_flater.ReplaceAllString(pr.Descriptions[0].Value, "|")
      if _, err = epg_desc.Exec(h_desc, flat_string); err != nil {
        log.Err(err).Send()
    		}
    } else {
      h_desc = 0
    }
  }
  // 2SQL: Связи
  epg_data.Exec(prov.IdHash, h_ch_id, h_title, h_desc, pr.Start.Unix(), pr.Stop.Unix())
}

/*
func MakeChannelStruct(db *sql.DB) error {
  // Database: Open
  db, err := sql.Open("sqlite3", "./chcache.db"); if err != nil {
    return err
  }
  defer db.Close()
    rows, err := db.Query("select id, name from foo")
	if err != nil {
    return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
      return err
		}
		log.Info().Msg(fmt.Sprintln(id,name))
	}
	err = rows.Err()
	if err != nil {
		log.Fatal().Err(err)
	}
  return nil
}*/