package app_config

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hjson/hjson-go/v4"
	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/lib/helpers"
)

// Структура, описывающая один элемент EPG
type ProvRecord struct {
  Id         string  `json:"id"`
  Urls     []string  `json:"urls"`
  LastEpg    uint64
  LastUpd    uint64
  IdHash     uint32
}

var Args struct {
  Tar    string
  EpgSources []*ProvRecord
  Gzip   int
  MemDb  bool
}
var (
  config_cache map[uint32][]*ProvRecord
  config_files []string
)


func ReadArgs() {
  config_files = make([]string, 0, 5)
  config_cache = make(map[uint32][]*ProvRecord)

  flag.BoolVar(&Args.MemDb, "epg-ram", false, "process epg in the ram instead of the file")
  flag.StringVar(&Args.Tar, "tar", "", "output all files to tar archive")
  flag.IntVar(&Args.Gzip, "z", 0, "gzip tar with a specified level")
  flag.Func("c", "config file, syntax: file_name[,prov_name,...,prov_name] (allow multiple -c)", stackConfigArg)
  flag.Parse()

  // Немного магии с именем файла
  name_len := len(Args.Tar)
  // Добавление сжатия, если имя файла, .gz
  if name_len >= 3 {
    if Args.Tar[name_len-3:] == ".gz" {
      if Args.Gzip == 0 { Args.Gzip = 1 }
    }
  }
}

func stackConfigArg(conf_val string) error {
  config_files = append(config_files, conf_val)
  return nil
}


func ReadConfigs() {
  var err error
  for i := 0; i < len(config_files); i++ {
    err = processConfigArg(config_files[i])
    if err != nil { log.Err(err).Send() }
  }
}
func processConfigArg(conf_val string) error {
  var (
    conf_data []byte
    err    error
    _config []*ProvRecord
    filter []string
  )
  // Обработка доп. параметров
  sub_args := strings.Split(conf_val, ",")

  fname   := sub_args[0]
  fname_h := helpers.HashSting32(fname)
  if conf, is_cached := config_cache[fname_h]; is_cached {
    log.Debug().Msgf("Cached config: %s", fname)
    _config = conf  // Конфиг в кеше
  } else { // Читаем конфиг
    log.Debug().Msgf("Load config: %s", fname)
    if helpers.HasHTTP(fname) {
      // Конфиг указан по ссылке
      var resp *http.Response
      
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
    err = hjson.Unmarshal(conf_data, &_config)
      if err != nil  { return err }
    config_cache[fname_h] = _config
  }
  
  // Обработка кофига (с фильтром)
  sub_args_len := len(sub_args)
  if sub_args_len > 1 {
    filter = sub_args[1:]
    sub_args_len--
    log.Debug().Msgf("  filter by: %v", filter)
  }
  read_config:
  for i := 0; i < len(_config); i++ {
    if filter != nil {
      for g := 0; g < sub_args_len; g++ {
        if _config[i].Id == filter[g] {
          Args.EpgSources = append(Args.EpgSources, _config[i])
          continue read_config
        }
      }
    }
  }
  return nil
}
