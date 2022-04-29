package epg_jsoner

import (
  "fmt"
  "os"
  "bytes"
  "database/sql"
  "github.com/rs/zerolog/log"
  "ott-play-epg-converter/lib/prov-meta"
  "ott-play-epg-converter/lib/arg-reader"
)

var (
  file_newline = []byte{0x2c,0x0a}
  empty_string = ""
)

type ChList map[uint32]uint32

func ProcessDB(db *sql.DB, prov *arg_reader.ProvRecord) bool {
  ch_in_epg := EpgGenerate(db, prov)
  if ch_in_epg != nil {
    err := ChListGenerate(db, prov, ch_in_epg)
    if err == nil { return true }
    log.Err(err).Send()
  }
  return true
}

func epgJson2File(prname string, ch_hash uint32, rec_count uint32, f *bytes.Buffer) bool {
  retn := false
  if rec_count > 0  {
    f.WriteString("\n]}");
    err := os.WriteFile(fmt.Sprintf("%s%d.json", prname, ch_hash), f.Bytes(), 0644)
    retn = (err == nil)
    if !retn { log.Err(err).Send() }    
  }
  f.Reset()
  return retn
}

func EpgGenerate(db *sql.DB, prov *arg_reader.ProvRecord) ChList {
  provEpgPath := fmt.Sprintf("%s%cepg%c", prov.Id, os.PathSeparator, os.PathSeparator)
  if _, err := os.Stat(provEpgPath); os.IsNotExist(err) {
    if err := os.MkdirAll(provEpgPath, 0755); err != nil {
      log.Err(err).Send()
      return nil
    }
  }
  rows, err := db.Query(`
    SELECT epg.data.h_ch_id, epg.h_title.data, epg.h_desc.data, epg.h_icon.data,
    epg.data.t_start, epg.data.t_stop
    FROM epg.data
    LEFT JOIN h_ch_ids ON epg.data.h_ch_id = h_ch_ids.h
    LEFT JOIN epg.h_title ON epg.data.h_title = epg.h_title.h
    LEFT JOIN epg.h_desc ON epg.data.h_desc = epg.h_desc.h
    LEFT JOIN epg.h_icon ON epg.data.h_icon = epg.h_icon.h
    WHERE h_ch_ids.h IS NOT NULL
    ORDER BY epg.data.h_ch_id, epg.data.t_start
    `)
  if err != nil {
    log.Err(err).Send()
    return nil
  }
  defer rows.Close()
  var buf []byte
  _rec := EpgRecord{}
  curr_channel := uint32(0) 
  prev_channel := uint32(0) 
  var f bytes.Buffer
  f.Grow(2097152) // Buffer 2MB
  _max_time    := uint64(0) // Максимальное время в epg
  _count_epg   := uint32(0) // Счетчик записей в epg
  _count_files := uint32(0) // Счетчик файлов с epg
  ch_map := make(ChList)
  for rows.Next() {
    // Чтение данных
    err = rows.Scan(&curr_channel, &_rec.Name, &_rec.Descr, &_rec.Icon, &_rec.Time, &_rec.TimeTo)
    if err != nil { log.Err(err).Send(); continue }
    // Канал поменялся?
    if curr_channel != prev_channel {
      if epgJson2File(provEpgPath, prev_channel, _count_epg, &f) {
        _count_files++  
        ch_map[prev_channel] = _count_epg
      }
      _count_epg = 0
      f.WriteString("{\"epg_data\":[\n");
    } else {
      f.Write(file_newline)  
    }
    _count_epg++
    // Fix: Descr всегда ""
    if _rec.Descr == nil { _rec.Descr = &empty_string }
    buf, err = _rec.MarshalJSON();
    if err != nil { log.Err(err).Send(); continue }
    f.Write(buf)
    prev_channel = curr_channel
    if _max_time < _rec.Time { _max_time = _rec.Time }
  }
  err = rows.Err(); if err != nil {
    log.Err(err).Send()
  }
  // Добавление последней записи
  if epgJson2File(provEpgPath, prev_channel, _count_epg, &f) {
    _count_files++  
    ch_map[prev_channel] = _count_epg
  }
  
  log.Info().Msgf("[%s] files count: %d", prov.Id, _count_files)
  prov_meta.InitProv(prov, _max_time)
  
  return ch_map
}


func chListPush(_ch_id uint32, _ch_names uint32, f *bytes.Buffer, rec *ChListData, last_line bool) bool {
  if _ch_names > 0 {
    buf, err := rec.MarshalJSON();
    if err != nil { log.Err(err).Send(); return false }
    f.WriteString(fmt.Sprintf(`"%d":`, _ch_id))
    f.Write(buf)
    if last_line {
      f.WriteString("\n}\n")
    } else {
      f.Write(file_newline)
    }
    return true
  }
  return false
}

func ChListGenerate(db *sql.DB, prov *arg_reader.ProvRecord, ch_map map[uint32]uint32) error {
  log.Info().Msgf("[%s] creating channels list...", prov.Id)
  channelsFile := fmt.Sprintf("%s%cchannels.json", prov.Id, os.PathSeparator)

  rows, err := db.Query(`
    SELECT ch_data.h_id, h_ch_ids.data, h_ch_names.data, h_ch_icons.data
    FROM ch_data
    LEFT JOIN h_ch_ids ON ch_data.h_id = h_ch_ids.h
    LEFT JOIN h_ch_names ON ch_data.h_name = h_ch_names.h
    LEFT JOIN h_ch_icons ON ch_data.h_icon = h_ch_icons.h
    WHERE ch_data.h_prov_id = ?
    ORDER BY ch_data.h_id
    `, prov.IdHash)
  if err != nil {
    return err
  }
  defer rows.Close()
  
  _rec := make(ChListData, 0, 16)
  var f bytes.Buffer
  f.Grow(2097152) // Buffer 2MB
  curr_channel := uint32(0)
  prev_channel := uint32(0)
  _ch_names    := uint32(0)  // Счетчик имен каналов
  _count_epg   := uint32(0)  // Счетчик каналов с epg
  _count_chnls := uint32(0)  // Счетчик каналов
  var _ch_id   *string
  var _ch_name *string
  var _ch_icon *string
//  var _epg_ok  *string
  var in_ch_map bool
  f.WriteString("{\n")
  for rows.Next() {
    // Чтение данных
    err = rows.Scan(&curr_channel, &_ch_id, &_ch_name, &_ch_icon)
    if err != nil { log.Err(err).Send(); continue }
    _, in_ch_map = ch_map[curr_channel]; if !in_ch_map {
      if _ch_icon == nil {
        // Если канал не содержит ни epg ни значка, то он бесполезен
        log.Warn().Msgf("[%s] channel has no epg and icon: %d/%s", prov.Id, curr_channel, *_ch_id)
        continue
      }
    }
    if curr_channel != prev_channel {
      if chListPush(prev_channel, _ch_names, &f, &_rec, false) {
        _count_chnls++  
      }
      _ch_names = 0
      _rec = _rec[:0]
      if in_ch_map { _count_epg++ }
      if _ch_icon == nil { _ch_icon = &empty_string }
      //ch_meta := []
      //ch_meta
      //"¦" + strconv.
      ch_meta := fmt.Sprintf("%s¦ %t¦ %s", *_ch_id, in_ch_map, *_ch_icon)
      //_rec = append(_rec, _epg_ok, _ch_id, _ch_icon)
      _rec = append(_rec, &ch_meta)
    }
    _rec = append(_rec, _ch_name)
    _ch_names++
    prev_channel = curr_channel
  }
  err = rows.Err(); if err != nil {
    log.Err(err).Send()
  }
  // Добавление последней записи
  if chListPush(prev_channel, _ch_names, &f, &_rec, true) {
    _count_chnls++  
  }
  // Сохранение файла на диск
  if err := os.WriteFile(channelsFile, f.Bytes(), 0644); err != nil {
    return err
  }
  log.Info().Msgf("[%s] list ready, channels: %d, with epg: %d", prov.Id, _count_chnls, _count_epg)
  return nil
}

