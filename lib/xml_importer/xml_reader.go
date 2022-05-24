package xml_importer

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"net/http"
	"os"
	"time"

	"crawshaw.io/sqlite"
	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/import/robbiet480/xmltv"
	"ott-play-epg-converter/lib/app_config"
	"ott-play-epg-converter/lib/helpers"
	"ott-play-epg-converter/lib/json_exporter"
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


func ProcessXml(db *sqlite.Conn, provData *app_config.ProvRecord) error {
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

  // Epg database: CleanUp & Attach
  db_epg := AttachEPG()
  defer db_epg.Close()
  log.Info().Msgf("[%s] EpgDb ready %f", provData.Id, time.Since(metric_start).Seconds())

  // Готовим запросы
  PrecompileQuery(db, db_epg)
  // Начинаем транзакции
  helpers.SimpleExec(db, "BEGIN TRANSACTION;", "ch - cannot start transaction")
  helpers.SimpleExec(db_epg, "BEGIN TRANSACTION;", "epg - cannot start transaction")

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
        NewChannelCache(&tvC, provData) 
      } else if v.Name.Local == "programme" {
        // XML: Read element <programme>
        tvP := xmltv.Programme{}
        if err = d.DecodeElement(&tvP, &v); err != nil {
          log.Err(err).Send()
        }
        NewProgCache(&tvP, provData)
      }
    }
  }

  log.Info().Msgf("[%s] Epg parsing is ready %f", provData.Id, time.Since(metric_start).Seconds())
  // Commit: Channels
  helpers.SimpleExec(db, "COMMIT TRANSACTION;", "ch - cannot end transaction")

  // Create json
  log.Info().Msgf("[%s] Database commit is ready %f", provData.Id, time.Since(metric_start).Seconds())
  json_exporter.ProcessDB(db_epg, provData)

  // Commit: EPG
  helpers.SimpleExec(db_epg, "COMMIT TRANSACTION; DETACH epg;", "epg - cannot end transaction")
  FinalizeEpgQuery()

  log.Info().Msgf("[%s] provider is ready %f", provData.Id, time.Since(metric_start).Seconds())
  return nil
}