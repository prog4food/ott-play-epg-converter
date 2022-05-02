package main

import (
	"bufio"
	"compress/gzip"
	"database/sql"
	"encoding/xml"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/import/robbiet480/xmltv"
	"ott-play-epg-converter/lib/arg_reader"
	"ott-play-epg-converter/lib/epg_jsoner"
)

func isGZip(in_reader *bufio.Reader) (*gzip.Reader, error) {
  is_gzip, err := in_reader.Peek(2); if err != nil {
    return nil, err
  }
  if is_gzip[0] == 0x1f && is_gzip[1] == 0x8b {
    return gzip.NewReader(in_reader)
  }
  return nil, nil
}


func processXml(db *sql.DB, provData *arg_reader.ProvRecord) error {
  metric_start := time.Now()
  var d *xml.Decoder
  var in_reader *bufio.Reader
  if provData.File != nil && *provData.File == "-" {
    // Reader: StdIn
    log.Info().Msgf("[%s] Read EPG from: StdIn", provData.Id)
    in_reader = bufio.NewReader(os.Stdin)
  } else if provData.File != nil {

    // Reader: File
    log.Info().Msgf("[%s] Read EPG from: %s", provData.Id, *provData.File)
    xmlFile, err := os.Open(*provData.File); if err != nil {
      log.Err(err).Send()
      return err
    }
    defer xmlFile.Close()
    in_reader = bufio.NewReaderSize(xmlFile, 1048576)

  } else if len(provData.Urls) > 0 {
    // Reader: HTTP
    log.Info().Msgf("[%s] Download EPG from: %s", provData.Id, provData.Urls[0])
    resp, err := http.Get(provData.Urls[0]); if err != nil {
      log.Err(err).Send()
      return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
      log.Error().Msgf("[%s] Download failed. %d:%s", provData.Id, resp.StatusCode, resp.Status)
      return errors.New("download failed")      
    }
    in_reader = bufio.NewReader(resp.Body)
  } else {
    log.Error().Msgf("[%s] provider has no eligible sources", provData.Id)
    return errors.New("has no eligible sources")
  }

  // Check for GZip header
  in_is_gzip, err := isGZip(in_reader); if err != nil {
    log.Err(err).Send()
    return err
  }
  if in_is_gzip != nil {
    log.Info().Msgf("[%s] input is gzipped", provData.Id)
    defer in_is_gzip.Close()
    d = xml.NewDecoder(in_is_gzip)
  } else {
    d = xml.NewDecoder(in_reader)
  }
  log.Info().Msgf("[%s] Input ready %f", provData.Id, time.Since(metric_start).Seconds())

  // Database: CleanUp
  InitEPG(db)

  log.Info().Msgf("[%s] EpgDb wiped %f", provData.Id, time.Since(metric_start).Seconds())

  // DB: Start transaction
  epgtx, err := db.Begin(); if err != nil {
    log.Err(err).Send()
    return err
  }

  // DB: Start transaction
  chtx, err := db.Begin(); if err != nil {
    log.Err(err).Send()
    return err
  }
 
  //// CH QUERY - BEGIN
  // PrecompiledQuery: ch_data insert
  sql_ch_data := PreQuery(chtx, "insert into ch_data values(?, ?, ?, ?, ?)")
  defer sql_ch_data.Close()
  // PrecompiledQuery: h_ch_names insert
  sql_ch_names := PreQuery(chtx, "insert into h_ch_names values(?, ?)")
  defer sql_ch_names.Close()
  // PrecompiledQuery: h_ch_ids insert
  sql_ch_ids := PreQuery(chtx, "insert into h_ch_ids values(?, ?)")
  defer sql_ch_ids.Close()
  // PrecompiledQuery: h_ch_icons insert
  sql_ch_icons := PreQuery(chtx, "insert into h_ch_icons values(?, ?)")
  defer sql_ch_ids.Close()
  //// CH QUERY - END
  
  //// EPG QUERY - BEGIN
  // PrecompiledQuery: epg.data insert
  sql_epg_data := PreQuery(epgtx, "insert into epg.temp_data values(?, ?, ?, ?, ?, ?, ?)")
  defer sql_epg_data.Close()

  // PrecompiledQuery: epg.h_title insert
  sql_epg_title := PreQuery(epgtx, "insert into epg.h_title values(?, ?)")
  defer sql_epg_title.Close()
  
  // PrecompiledQuery: epg.h_desc insert
  sql_epg_desc := PreQuery(epgtx, "insert into epg.h_desc values(?, ?)")
  defer sql_epg_desc.Close()
  
  // PrecompiledQuery: epg.h_icon insert
  sql_epg_icon := PreQuery(epgtx, "insert into epg.h_icon values(?, ?)")
  defer sql_epg_icon.Close()
  //// EPG QUERY - END
  
  // XML: Process elements
  for {
    t, err := d.Token(); if err != nil { break }
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
        NewProgCache(sql_epg_data, sql_epg_title, sql_epg_desc, sql_epg_icon, &tvP, provData)
      }
    }
  }

  // DB: Final
  log.Info().Msgf("[%s] Epg parsing is ready %f", provData.Id, time.Since(metric_start).Seconds())
  chtx.Commit()
  
  // Create json
  log.Info().Msgf("[%s] Database commit is ready %f", provData.Id, time.Since(metric_start).Seconds())
  epg_jsoner.ProcessDB(epgtx, provData)
  //epgtx.Commit()
  epgtx.Rollback()
  FinallyEPG(db)
  log.Info().Msgf("[%s] provider is ready %f", provData.Id, time.Since(metric_start).Seconds())
  return nil
}