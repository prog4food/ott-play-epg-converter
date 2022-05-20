package epg_jsoner

import (
	"bytes"
	"database/sql"
	json "encoding/json"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/lib/arg_reader"
	"ott-play-epg-converter/lib/prov_meta"
	"ott-play-epg-converter/lib/string_hashes"
)

var (
  file_newline = []byte{0x2c,0x0a}
  empty_string = ""
  path_sep = string(os.PathSeparator)
)

type ChList map[uint32]uint64


func ProcessDB(db *sql.Tx, prov *arg_reader.ProvRecord) bool {
  ch_in_epg := EpgGenerate(db, prov)
  if ch_in_epg != nil {
    err := ChListGenerate(db, prov, ch_in_epg)
    if err == nil { return true }
    log.Err(err).Send()
  }
  return true
}

func epgJson2File(prname string, ch_hash uint32, top_time_ch uint64, f *bytes.Buffer) bool {
  retn := false
  if top_time_ch > 0  {
    f.WriteString("\n]}");
    err := os.WriteFile(prname + strconv.FormatUint(uint64(ch_hash), 10) + ".json", f.Bytes(), 0644)
    retn = (err == nil)
    if !retn { log.Err(err).Send() }    
  }
  f.Reset()
  return retn
}

func EpgGenerate(db *sql.Tx, prov *arg_reader.ProvRecord) ChList {
  provEpgPath := prov.Id + path_sep + "epg" + path_sep
  if _, err := os.Stat(provEpgPath); os.IsNotExist(err) {
    if err := os.MkdirAll(provEpgPath, 0755); err != nil {
      log.Err(err).Send()
      return nil
    }
  }
  log.Info().Msgf("[%s] sorting epg_data...", prov.Id)
  _, err := db.Exec(`
    CREATE TABLE epg.data AS SELECT * FROM epg.temp_data ORDER BY h_ch_id, t_start;
    --DROP TABLE epg.temp_data;`)
  if err != nil {
    log.Err(err).Send()
    return nil
  }
  rows, err := db.Query(`
    SELECT epg.data.h_ch_id, epg.h_title.data, epg.h_desc.data, epg.h_icon.data,
    epg.data.t_start, epg.data.t_stop
    FROM epg.data
    LEFT JOIN h_ch_ids ON epg.data.h_ch_id = h_ch_ids.h
    LEFT JOIN epg.h_title ON epg.data.h_title = epg.h_title.h
    LEFT JOIN epg.h_desc ON epg.data.h_desc = epg.h_desc.h
    LEFT JOIN epg.h_icon ON epg.data.h_icon = epg.h_icon.h
    WHERE h_ch_ids.h IS NOT NULL;
    `)
  if err != nil {
    log.Err(err).Send()
    return nil
  }
  defer rows.Close()
  log.Info().Msgf("[%s] generating json...", prov.Id)
  
  _rec := EpgRecord{}  // Временный JSON объект
  var buf []byte       // Временный буфер для JSON.Marshal
  var f bytes.Buffer   // Временный буфер для JSON файла
  f.Grow(2097152)      // Буфер для файла 2MB

  ch_map := make(ChList)  // Список обработанных каналов
  curr_channel  := uint32(0) 
  prev_channel  := uint32(0) 
  _top_time_ch  := uint64(0)  // Крайнее время передачи для канала
  _top_time_epg := uint64(0)  // Крайнее время передачи для провайдера
  for rows.Next() {
    // Чтение данных
    err = rows.Scan(&curr_channel, &_rec.Name, &_rec.Descr, &_rec.Icon, &_rec.Time, &_rec.TimeTo)
    if err != nil { log.Err(err).Send(); continue }
    // Канал поменялся?
    if curr_channel != prev_channel {
      if epgJson2File(provEpgPath, prev_channel, _top_time_ch, &f) {
        ch_map[prev_channel] = _top_time_ch
      }
      _top_time_ch = 0
      f.WriteString("{\"epg_data\":[\n");
    } else {
      f.Write(file_newline)  
    }
    // Fix: Descr всегда "" (TODO: проверить в плеере)
    if _rec.Descr == nil { _rec.Descr = &empty_string }
    buf, err = _rec.MarshalJSON();
    if err != nil { log.Err(err).Send(); continue }
    f.Write(buf)
    prev_channel = curr_channel
    if _top_time_ch < _rec.Time {
      // Обновление _max_time для ch и epg
      _top_time_ch = _rec.Time
      if _top_time_epg < _top_time_ch { _top_time_epg = _top_time_ch }
    }
  }
  err = rows.Err(); if err != nil {
    log.Err(err).Send()
  }
  // Добавление последней записи
  if epgJson2File(provEpgPath, prev_channel, _top_time_ch, &f) {
    ch_map[prev_channel] = _top_time_ch
  }
  
  log.Info().Msgf("[%s] files count: %d", prov.Id, len(ch_map))
  prov_meta.PushProv(prov, _top_time_epg)
  
  return ch_map
}

func chListMeta(f *bytes.Buffer, prov *arg_reader.ProvRecord) {
  ch_meta := &prov_meta.ProvMeta{}
  ch_meta.Id = &prov.Id
  ch_meta.LastEpg = prov.LastEpg
  ch_meta.LastUpd = prov.LastUpd
  ch_meta.Urls = make([]uint32, len(prov.Urls))
  for i := 0; i < len(prov.Urls); i++ {
    ch_meta.Urls[i] = string_hashes.HashSting32(prov.Urls[i])
  }
  buf, err := json.Marshal(ch_meta);
  if err != nil { log.Err(err).Send(); return }
  f.WriteString(`"meta": `)
  f.Write(buf)
  f.WriteString(",\n")
}

func chListPush(_ch_id uint32, _ch_names uint32, f *bytes.Buffer, rec *ChListData, last_line bool) bool {
  if _ch_names > 0 {
    buf, err := rec.MarshalJSON();
    if err != nil { log.Err(err).Send(); return false }
    f.WriteString(`"` + strconv.FormatUint(uint64(_ch_id), 10) + `":`)
    f.Write(buf)
    if last_line {
      f.WriteString("\n}}\n")
    } else {
      f.Write(file_newline)
    }
    return true
  }
  return false
}

func ChListGenerate(db *sql.Tx, prov *arg_reader.ProvRecord, ch_map ChList) error {
  log.Info().Msgf("[%s] creating channels list...", prov.Id)
  channelsFile := prov.Id + path_sep + "channels.json"

  rows, err := db.Query(`
    SELECT ch_data.h_id, h_ch_ids.data, h_ch_names.data, h_ch_icons.data
    FROM ch_data
    LEFT JOIN h_ch_ids ON ch_data.h_id = h_ch_ids.h
    LEFT JOIN h_ch_names ON ch_data.h_name = h_ch_names.h
    LEFT JOIN h_ch_icons ON ch_data.h_icon = h_ch_icons.h
    WHERE ch_data.h_prov_id = ?
    ORDER BY ch_data.h_id;
    `, prov.IdHash)
  if err != nil {
    return err
  }
  defer rows.Close()
  
  _rec := make(ChListData, 0, 16)
  _meta_buf := make([]byte, 0, 512)
  var f bytes.Buffer
  f.Grow(2097152) // Buffer 2MB
  curr_channel := uint32(0)
  prev_channel := uint32(0)
  _count_names := uint32(0)  // Счетчик имен каналов
  _count_epgch := uint32(0)  // Счетчик каналов с epg
  _count_allch := uint32(0)  // Счетчик всех каналов
  _top_time_ch := uint64(0)  // Крайнее время передачи для канала
  _time_ch_ok  := false
  var _ch_id   *string
  var _ch_name *string
  var _ch_icon *string
  f.WriteString("{")
  chListMeta(&f, prov)
  f.WriteString(`"data": {` + "\n")
  for rows.Next() {
    // Чтение данных
    err = rows.Scan(&curr_channel, &_ch_id, &_ch_name, &_ch_icon)
    if err != nil { log.Err(err).Send(); continue }
    _top_time_ch, _time_ch_ok = ch_map[curr_channel]; if !_time_ch_ok {
      if _ch_icon == nil || (!strings.HasPrefix(*_ch_icon, "http://") && !strings.HasPrefix(*_ch_icon, "https://")) {
        // Если канал не содержит epg или логотипа, то он бесполезен
        log.Warn().Msgf("[%s] channel has no epg and icon: %d/%s", prov.Id, curr_channel, *_ch_id)
        continue
      }
    }
    if curr_channel != prev_channel {
      if chListPush(prev_channel, _count_names, &f, &_rec, false) {
        _count_allch++  
      }
      _count_names = 0
      _rec = _rec[:0]
      if _time_ch_ok { _count_epgch++ }
      if _ch_icon == nil { _ch_icon = &empty_string }
      _meta_buf = _meta_buf[:0]
      _meta_buf = append(_meta_buf, *_ch_id...)
      _meta_buf = append(_meta_buf, 0xc2, 0xa6)
      _meta_buf = strconv.AppendUint(_meta_buf, _top_time_ch, 10)
      _meta_buf = append(_meta_buf, 0xc2, 0xa6)
      _meta_buf = append(_meta_buf, *_ch_icon...)
      _rec = append(_rec, (*string)(unsafe.Pointer(&_meta_buf)))
    }
    _rec = append(_rec, _ch_name)
    _count_names++
    prev_channel = curr_channel
  }
  err = rows.Err(); if err != nil {
    log.Err(err).Send()
  }
  // Добавление последней записи
  if chListPush(prev_channel, _count_names, &f, &_rec, true) {
    _count_allch++  
  }
  // Сохранение файла на диск
  if err := os.WriteFile(channelsFile, f.Bytes(), 0644); err != nil {
    return err
  }
  log.Info().Msgf("[%s] list ready, channels: %d, with epg: %d", prov.Id, _count_allch, _count_epgch)
  return nil
}

