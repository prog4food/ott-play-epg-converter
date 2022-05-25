package json_exporter

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	json "encoding/json"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"crawshaw.io/sqlite"
	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/lib/app_config"
	"ott-play-epg-converter/lib/helpers"
)

var (
  tar_file     *os.File
  tar_writer   *tar.Writer
  gzip_writer  *gzip.Writer
  tar_std_file *tar.Header
  file_newline = []byte{0x2c,0x0a}  // byte вариант ",\n"
  empty_string = ""
  path_sep = string(os.PathSeparator)
)

type ChList map[uint32]uint64


func SetTarOutput() {
  var err error
  if app_config.Args.Tar == "" {          //  TAR: No
    return
  }
  _conf := app_config.Args
  // Куда выводить
  if _conf.Tar == "-" {  // TAR: StdOut
    tar_file = os.Stdout
  } else {                                // TAR: File
    tar_file, err = os.Create(_conf.Tar)
    if err != nil { log.Panic().Err(err).Msg("tar file: create error") }
  }

  // Сжимать ли поток
  if _conf.Gzip != 255 {
    // Gzip - ON
    gzip_writer, err = gzip.NewWriterLevel(tar_file, _conf.Gzip)
    if err != nil { log.Panic().Err(err).Msg("gzip writer: create error") }
    tar_writer = tar.NewWriter(gzip_writer)
  } else {
    // Gzip - OFF
    tar_writer = tar.NewWriter(tar_file)
  }
  tar_std_file = &tar.Header{ Typeflag: tar.TypeReg, Mode: 0644 }
  path_sep = "/"
}

func CloseTarOutput() {
  if tar_file != nil {
    tar_writer.Close()
    if gzip_writer != nil { gzip_writer.Close() }
    tar_file.Close()
    tar_file = nil
  }
}

func ProcessDB(db *sqlite.Conn, prov *app_config.ProvRecord) bool {
  if tar_file != nil {
    var err error
    var tar_dir *tar.Header
    tar_dir = &tar.Header{
      Typeflag: tar.TypeDir,
      Name: prov.Id,
      Mode: 0755,
    }
    err = tar_writer.WriteHeader(tar_dir)
    if err == nil { 
      tar_dir.Name = prov.Id + "/epg"
      err = tar_writer.WriteHeader(tar_dir);
    }
    if err != nil { 
      log.Err(err).Msg("cannot write tar header")
      return false
    }
  }
  ch_in_epg := EpgGenerate(db, prov)
  if ch_in_epg != nil {
    err := ChListGenerate(db, prov, ch_in_epg)
    if err == nil { return true }
    log.Err(err).Send()
  }
  return true
}

func epgJson2File(prname string, ch_hash uint32, top_time_ch uint64, f *bytes.Buffer) bool {
  var err error
  retn := false
  if top_time_ch > 0  {
    f.WriteString("\n]}");
    if tar_file == nil {
      err = os.WriteFile(prname + strconv.FormatUint(uint64(ch_hash), 10) + ".json", f.Bytes(), 0644)
    } else {
      tar_std_file.Name = prname + strconv.FormatUint(uint64(ch_hash), 10) + ".json"
      tar_std_file.Size = int64(f.Len())
      err = tar_writer.WriteHeader(tar_std_file)
      f.WriteTo(tar_writer)
    }
    retn = (err == nil)
    if !retn { log.Err(err).Send() }
  }
  f.Reset()
  return retn
}

func EpgGenerate(db *sqlite.Conn, prov *app_config.ProvRecord) ChList {
  provEpgPath := prov.Id + path_sep + "epg" + path_sep

  _, err := os.Stat(provEpgPath); if os.IsNotExist(err) {
    if err := os.MkdirAll(provEpgPath, 0755); err != nil {
      log.Err(err).Send()
      return nil
    }
  }

  log.Info().Msgf("[%s] sorting epg_data...", prov.Id)
  helpers.SimpleExec(db, "CREATE TABLE epg.data AS SELECT * FROM epg.temp_data ORDER BY h_ch_id, t_start;", "epg sort error"); 
  // ^ ??? DROP TABLE epg.temp_data;

  stmt, tb, err := db.PrepareTransient(`
    SELECT epg.temp_data.h_ch_id, epg.h_title.data, epg.h_desc.data, epg.h_icon.data,
    epg.temp_data.t_start, epg.temp_data.t_stop
    FROM epg.temp_data
    LEFT JOIN h_ch_ids ON epg.temp_data.h_ch_id = h_ch_ids.h
    LEFT JOIN epg.h_title ON epg.temp_data.h_title = epg.h_title.h
    LEFT JOIN epg.h_desc ON epg.temp_data.h_desc = epg.h_desc.h
    LEFT JOIN epg.h_icon ON epg.temp_data.h_icon = epg.h_icon.h
    WHERE h_ch_ids.h IS NOT NULL;`)
  if err != nil {
    log.Err(err).Msg("cannot compile epg final query")
    return nil
  }
  if tb != 0 { log.Error().Int("pos", tb).Msg("epg final query has trailing bytes") }
  defer stmt.Finalize()
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
  var hasRow bool
  for {
    if hasRow, err = stmt.Step(); err != nil {
      log.Err(err).Msg("epg export: cannot read row!")
      continue
    } else if !hasRow {
      // Добавление последней записи
      if epgJson2File(provEpgPath, prev_channel, _top_time_ch, &f) {
        ch_map[prev_channel] = _top_time_ch
      }
      break
    }
    // Чтение данных
    curr_channel = uint32(stmt.ColumnInt32(0))
    _rec.Name    = stmt.ColumnText(1)
    _rec.Descr   = stmt.ColumnText(2)
    _rec.Descr   = stmt.ColumnText(3)
    _rec.Time    = uint64(stmt.ColumnInt64(4))
    _rec.TimeTo  = uint64(stmt.ColumnInt64(5))
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

  log.Info().Msgf("[%s] files count: %d", prov.Id, len(ch_map))
  ProvList_Update(prov, _top_time_epg)  
  return ch_map
}

func chListMeta(f *bytes.Buffer, prov *app_config.ProvRecord) {
  ch_meta := &provMeta{}
  ch_meta.Id = &prov.Id
  ch_meta.LastEpg, ch_meta.LastUpd = prov.LastEpg, prov.LastUpd
  ch_meta.Urls = make([]uint32, len(prov.Urls))
  for i := 0; i < len(prov.Urls); i++ {
    ch_meta.Urls[i] = helpers.HashSting32i(helpers.CutHTTP(prov.Urls[i]))
  }
  buf, err := json.Marshal(ch_meta);
  if err != nil { log.Err(err).Send(); return }
  f.WriteString(`"meta": `)
  f.Write(buf)
  f.Write(file_newline)  
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

func ChListGenerate(db *sqlite.Conn, prov *app_config.ProvRecord, ch_map ChList) error {
  log.Info().Msgf("[%s] creating channels list...", prov.Id)
  channelsFile := prov.Id + path_sep + "channels.json"

  stmt, tb, err := db.PrepareTransient(`
    SELECT ch_data.h_id, h_ch_ids.data, h_ch_names.data, h_ch_icons.data
    FROM ch_data
    LEFT JOIN h_ch_ids ON ch_data.h_id = h_ch_ids.h
    LEFT JOIN h_ch_names ON ch_data.h_name = h_ch_names.h
    LEFT JOIN h_ch_icons ON ch_data.h_icon = h_ch_icons.h
    WHERE ch_data.h_prov_id = ?
    ORDER BY ch_data.h_id;`)
  if err != nil {
    log.Err(err).Msg("cannot compile channels list query")
    return err
  }
  if tb != 0 { log.Error().Int("pos", tb).Msg("channels list query has trailing bytes") }
  defer stmt.Finalize()
  stmt.BindInt64(1, int64(prov.IdHash))

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
  var _ch_id   string
  var _ch_name string
  var _ch_icon string
  f.WriteString("{")
  chListMeta(&f, prov)
  f.WriteString("\"data\": {\n")
  var hasRow bool
  for {
    if hasRow, err = stmt.Step(); err != nil {
      log.Err(err).Msg("epg export: cannot read row!")
      continue
    } else if !hasRow {
      // Добавление последней записи
      if chListPush(prev_channel, _count_names, &f, &_rec, true) {
        _count_allch++  
      }
      break
    }
    // Чтение данных
    curr_channel = uint32(stmt.ColumnInt32(0))
    _ch_id       = stmt.ColumnText(1)
    _ch_name     = stmt.ColumnText(2)
    _ch_icon     = stmt.ColumnText(3)

    _top_time_ch, _time_ch_ok = ch_map[curr_channel]; if !_time_ch_ok {
      if _ch_icon != "" || (!strings.HasPrefix(_ch_icon, "http://") && !strings.HasPrefix(_ch_icon, "https://")) {
        // Если канал не содержит epg или логотипа, то он бесполезен
        log.Warn().Msgf("[%s] channel has no epg and icon: %d/%s", prov.Id, curr_channel, _ch_id)
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
      _meta_buf = _meta_buf[:0]
      _meta_buf = append(_meta_buf, _ch_id...)
      _meta_buf = append(_meta_buf, 0xc2, 0xa6)
      _meta_buf = strconv.AppendUint(_meta_buf, _top_time_ch, 10)
      _meta_buf = append(_meta_buf, 0xc2, 0xa6)
      _meta_buf = append(_meta_buf, _ch_icon...)
      _rec = append(_rec, (*string)(unsafe.Pointer(&_meta_buf)))
    }
    _rec = append(_rec, &_ch_name)
    _count_names++
    prev_channel = curr_channel
  }

  // Сохранение файла на диск
  if tar_file == nil {
    err = os.WriteFile(channelsFile, f.Bytes(), 0644);
  } else {
    tar_std_file.Name = channelsFile
    tar_std_file.Size = int64(f.Len())
    err = tar_writer.WriteHeader(tar_std_file)
    f.WriteTo(tar_writer)
  }
  if err != nil {
    log.Err(err).Msg("epg export: cannot read row!")
    return err
  }
  log.Info().Msgf("[%s] list ready, channels: %d, with epg: %d", prov.Id, _count_allch, _count_epgch)
  return nil
}

