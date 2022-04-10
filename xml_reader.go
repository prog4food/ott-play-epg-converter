package main

import (
  "time"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "github.com/rs/zerolog/log"
  "os"
  "bufio"
  "encoding/xml"
  "ott-play-epg-converter/import/robbiet480/xmltv"
  "ott-play-epg-converter/lib/epg-jsoner"
  "ott-play-epg-converter/lib/arg-reader"
)

func processXml(db *sql.DB, provData *arg_reader.ProvRecord) error {
  metric_start := time.Now()
  var d *xml.Decoder
  
  if provData.File != "-" {
    // Reader: File
    xmlFile, err := os.Open(provData.File); if err != nil {
      log.Err(err).Send()
      return err
    }
    log.Info().Msgf("[%s] Read EPG from: %s", provData.Id, provData.File)
    defer xmlFile.Close()   
    d = xml.NewDecoder(xmlFile)
  } else {
    // Reader: StdIn
    log.Info().Msgf("[%s] Read EPG from: StdIn", provData.Id)
    in_reader := bufio.NewReader(os.Stdin)
    d = xml.NewDecoder(in_reader)
  }

  // Database: CleanUp
  InitEPG(db)

  log.Info().Msgf("[%s] EpgDb wiped %f", provData.Id, time.Since(metric_start).Seconds())

  // DB: Start transaction
  tx, err := db.Begin(); if err != nil {
    log.Err(err).Send()
    return err
  }
 
  //// CH QUERY - BEGIN
  // PrecompiledQuery: ch_data insert
  sql_ch_data := PreQuery(tx, "insert into ch_data values(?, ?, ?, ?, ?)")
  defer sql_ch_data.Close()
  // PrecompiledQuery: h_ch_names insert
  sql_ch_names := PreQuery(tx, "insert into h_ch_names values(?, ?)")
  defer sql_ch_names.Close()
  // PrecompiledQuery: h_ch_ids insert
  sql_ch_ids := PreQuery(tx, "insert into h_ch_ids values(?, ?)")
  defer sql_ch_ids.Close()
  // PrecompiledQuery: h_ch_icons insert
  sql_ch_icons := PreQuery(tx, "insert into h_ch_icons values(?, ?)")
  defer sql_ch_ids.Close()
  //// CH QUERY - END
  
  //// EPG QUERY - BEGIN
  // PrecompiledQuery: epg.data insert
  sql_epg_data := PreQuery(tx, "insert into epg.data values(?, ?, ?, ?, ?, ?)")
  defer sql_epg_data.Close()
  
  // PrecompiledQuery: epg.h_title insert
  sql_epg_title := PreQuery(tx, "insert into epg.h_title values(?, ?)")
  defer sql_epg_title.Close()
  
  // PrecompiledQuery: epg.h_desc insert
  sql_epg_desc := PreQuery(tx, "insert into epg.h_desc values(?, ?)")
  defer sql_epg_desc.Close()
  //// EPG QUERY - END
  
  // XML: Process elements
  for {
    t, err := d.Token(); if err != nil {
      break
    }
    switch v := t.(type) {
    case xml.StartElement:
      if v.Name.Local == "channel" {
        // XML: Read element <channel>
        tvC := xmltv.Channel{}
          if err := d.DecodeElement(&tvC, &v); err != nil {
          log.Err(err).Send()
          }
        NewChannelCache(sql_ch_data, sql_ch_ids, sql_ch_names, sql_ch_icons, &tvC, provData) 
      } else if v.Name.Local == "programme" {
        // XML: Read element <programme>
        tvP := xmltv.Programme{}
        if err = d.DecodeElement(&tvP, &v); err != nil {
          log.Err(err).Send()
        }
        NewProgCache(sql_epg_data, sql_epg_title, sql_epg_desc, &tvP, provData)
      }
//    case xml.EndElement:
//      fmt.Println(v.Name.Local)
    }
  }

  // DB: Final
  log.Info().Msgf("[%s] Epg parsing is ready %f", provData.Id, time.Since(metric_start).Seconds())
  tx.Commit()
  
  // Create json
  log.Info().Msgf("[%s] Database commit is ready %f", provData.Id, time.Since(metric_start).Seconds())
  if err := epg_jsoner.ProcessDB(db, provData.Id, provData.IdHash); err != nil {
    log.Err(err).Send()
	}
  FinallyEPG(db)
  log.Info().Msgf("[%s] Json files is ready %f", provData.Id, time.Since(metric_start).Seconds())
  //if _, err := db.Exec("DELETE FROM epg_data; DELETE FROM h_epg_title; DELETE FROM h_epg_desc; VACUUM;"); err != nil {
  //  log.Err(err).Send()
  //}
  return nil
}