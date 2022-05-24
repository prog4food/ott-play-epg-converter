package app_config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/lib/helpers"
)

// Структура, описывающая один элемент EPG
type ProvRecord struct {
  File    *string  `json:"file"`
  Id       string  `json:"id"`
  IdHash   uint32
  Urls   []string  `json:"urls"`
  LastEpg  uint64
  LastUpd  uint64
}

var Args struct {
  MemDb  bool
  Tar    string
  Gzip   int
  EpgSources []*ProvRecord
}
var config_cache map[uint32][]*ProvRecord

func ArgPanic(err error, args []string){
  if err != nil {
    log.Err(err).Send();
  }
  log.Panic().Msg(fmt.Sprint("Unrecognized command line: ", args[1:]))
}

func ReadArgs() {
  config_cache = make(map[uint32][]*ProvRecord)

  flag.BoolVar(&Args.MemDb, "epg-ram", false, "process epg in the ram instead of the file")
  flag.StringVar(&Args.Tar, "tar", "", "output all files to tar archive")
  flag.IntVar(&Args.Gzip, "z", 255, "gzip tar with a specified level")
  flag.Func("c", "config file, syntax: file_name[,prov_name] (allow multiple -c)", processConfigArg)
  flag.Parse()

  // Немного магии с именем файла
  name_len := len(Args.Tar)
  // Добавление сжатия, если имя файла, .gz
  if name_len >= 3 {
    if Args.Tar[name_len-3:] == ".gz" {
      if Args.Gzip == 255 { Args.Gzip = 1 }
    }
  }
}

func processConfigArg(conf_val string) error {
  var (
    conf_data []byte
    err    error
    _config []*ProvRecord
    filter *string
  )
  // Обработка суб аргуметов
  sub_args := strings.Split(conf_val, ",")
  sub_args_len := len(sub_args)
  if sub_args_len != 1 && sub_args_len != 2 {
    return fmt.Errorf("Cannot parce config argument: %s", conf_val)
  }
  
  fname   := sub_args[0]
  fname_h := helpers.HashSting32(fname)
  if conf, is_cached := config_cache[fname_h]; is_cached {
    _config = conf  // Конфиг в кеше
  } else { // Читаем конфиг
    if helpers.HasHTTP(fname) {
      // Конфиг указан по ссылке
      var resp *http.Response
      
      log.Info().Msgf("Download config from: %s", fname)
      resp, err = http.Get(fname)
        if err != nil { return err }
      if resp.StatusCode != 200 {
        return fmt.Errorf("Cannot download config. %d:%s", resp.StatusCode, resp.Status)
      }
      defer resp.Body.Close()
      conf_data, err = io.ReadAll(resp.Body)
    } else {
      // Конфиг указан файлом
      conf_data, err = os.ReadFile(fname)
    }
    if err != nil  { return err }
    
    _config = []*ProvRecord{}
    err = json.Unmarshal(conf_data, &_config)
      if err != nil  { return err }
    config_cache[fname_h] = _config
  }
  
  if sub_args_len == 2  { filter = &sub_args[1] }
  // Обработка кофига (с фильтром)
  for i := 0; i < len(_config); i++ {
    if filter != nil && *filter != _config[i].Id { continue }
    Args.EpgSources = append(Args.EpgSources, _config[i])
  }
  return nil
}
