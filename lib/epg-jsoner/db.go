package epg_jsoner

import (
  "fmt"
  "os"
  "bytes"
  "database/sql"
  "github.com/rs/zerolog/log"
  "ott-play-epg-converter/lib/string-hashes"
)

var (
  file_newline = []byte{0x2c,0x0a}
  empty_string = ""
)


func SyncFile(prname string, chname string, rec_count uint32, f *bytes.Buffer) {
  if rec_count > 0  {
    f.WriteString("\n]}");
    err := os.WriteFile(fmt.Sprintf("%s%d.json", prname, string_hashes.HashSting32(chname)), f.Bytes(), 0644)
    if err != nil { log.Err(err).Send() }
  }
  f.Reset()
}

func ProcessDB(db *sql.DB, provName string, provIdHash uint32) error {
  provEpgPath := fmt.Sprintf("%s%cepg%c",provName, os.PathSeparator, os.PathSeparator)
  if _, err := os.Stat(provEpgPath); os.IsNotExist(err) {
    if err := os.MkdirAll(provEpgPath, 0755); err != nil { return err }
  }
  rows, err := db.Query(
    `
    SELECT h_ch_ids.data, epg.h_title.data, epg.h_desc.data,
    epg.data.t_start, epg.data.t_stop
    FROM epg.data
    LEFT JOIN h_ch_ids ON epg.data.h_ch_id = h_ch_ids.h
    LEFT JOIN epg.h_title ON epg.data.h_title = epg.h_title.h
    LEFT JOIN epg.h_desc ON epg.data.h_desc = epg.h_desc.h
    WHERE h_ch_ids.data IS NOT NULL
    ORDER BY epg.data.h_ch_id, epg.data.t_start
    `)
  if err != nil {
    return err
  }
  defer rows.Close()
  //var f *os.File = nil
  var buf []byte
  tmp_Record := EpgRecord{}
  current_channel := ""
  prev_channel    := ""
  var f bytes.Buffer
  f.Grow(2097152) // Buffer 2MB
  epg_rec_count := uint32(0)
  epg_provider_files := 0
  for rows.Next() {
    // Reset fields
    tmp_Record.Name   = ""
    tmp_Record.Descr  = nil
    tmp_Record.Time   = 0
    tmp_Record.TimeTo = 0
    // Read Row
    err = rows.Scan(&current_channel, &tmp_Record.Name, &tmp_Record.Descr, &tmp_Record.Time, &tmp_Record.TimeTo)
    if err != nil { log.Err(err).Send(); continue }
    // Channel changed?
    if current_channel != prev_channel {
      SyncFile(provEpgPath, prev_channel, epg_rec_count, &f)
      epg_provider_files++
      epg_rec_count = 0
      f.WriteString("{\"epg_data\":[\n");
    } else {
      f.Write(file_newline)  
    }
    epg_rec_count++
    // Fix: Empty description as ""
    if tmp_Record.Descr == nil { tmp_Record.Descr = &empty_string }
    buf, err = tmp_Record.MarshalJSON();
    if err != nil { log.Err(err).Send(); continue }
    f.Write(buf)
    prev_channel = current_channel
  }
  SyncFile(provEpgPath, prev_channel, epg_rec_count, &f)
  epg_provider_files++
  
  log.Info().Msgf("[%s] files count: %d", provName, epg_provider_files)
  
  err = rows.Err(); if err != nil {
    return err
  }
  return nil
}


func MakeChainList(db *sql.DB, provName string, provIdHash uint32) error {
  return nil
}

