package arg_reader

import (
  "fmt"
  "os"
  "io"
  "strings"
  "encoding/json"
  "net/http"
  "github.com/rs/zerolog/log"
  "ott-play-epg-converter/lib/string-hashes"
)
// Структура, описывающая один элемент EPG
type ProvRecord struct {
  File    *string  `json:"file"`
  Id       string  `json:"id"`
  IdHash   uint32
  Order    uint8   `json:"order"`
  Urls   []string  `json:"urls"`
}

type ArgData struct {
  MakeList    bool
  EpgSources  []*ProvRecord
}

var AppConfig ArgData
var configCache map[uint32][]*ProvRecord


func ArgPanic(err error, args []string){
  if err != nil {
    log.Err(err).Send();
  }
  log.Panic().Msg(fmt.Sprint("Unrecognized command line: ", args[1:]))
}

func ProcessC(arg_pos int, args []string) int {
  var (
    jsonData []byte
    jsonErr    error
    config_data []*ProvRecord
    filter *string
  )
  
  // Хватает ли аргуметов
  arg_pos++
  if arg_pos > len(args)  { ArgPanic(nil, args) }
  
  // Обработка суб аргуметов
  sub_args := strings.Split(args[arg_pos], ",")
  sub_args_len := len(sub_args)
  if sub_args_len != 1 && sub_args_len != 2 { ArgPanic(nil, args) }
  
  fname   := sub_args[0]
  fname_h := string_hashes.HashSting32(fname)
  if conf, is_cached := configCache[fname_h]; is_cached {
    config_data = conf  // Конфиг в кеше
  } else { // Читаем конфиг
    if strings.HasPrefix(fname, "http://") || strings.HasPrefix(fname, "https://") {
      log.Info().Msgf("Download config from: %s", fname)
      resp, err := http.Get(fname); if err != nil {
        log.Err(err).Send()
        return 1
      }
      if resp.StatusCode != 200 {
        log.Error().Msgf("Download failed. %d:%s", resp.StatusCode, resp.Status)
        return 1
      }
      defer resp.Body.Close()
      jsonData, jsonErr = io.ReadAll(resp.Body)
    } else {
      jsonData, jsonErr = os.ReadFile(fname)
    }
    if jsonErr != nil  {
      log.Err(jsonErr).Send()
      return 1
    }
    
    config_data = []*ProvRecord{}
    if err := json.Unmarshal(jsonData, &config_data); err != nil  {
      log.Err(err).Send()
      return 1
    }
    configCache[fname_h] = config_data
  }
  
  if sub_args_len == 2  { filter = &sub_args[1] }
  // Обработка кофига (с фильтром)
  for i := 0; i < len(config_data); i++ {
    if filter != nil && *filter != config_data[i].Id { continue }
    AppConfig.EpgSources = append(AppConfig.EpgSources, config_data[i])
  }
  return 1
}

func ProcessL(arg_pos int, args []string) int {
  const block_args = 0
  AppConfig.MakeList = true
  return 0
}

func ProcessS(arg_pos int, args []string) int {
  // TODO
  return 0
}

func DetectArg(num int, args []string) func(int, []string) int {
  	switch args[num] {
  	  case "-l": return ProcessL
    case "-c": return ProcessC
    //case "-s": return ProcessS // TODO
    default:   return nil
	}
}

func ParseArgs(args []string) {
  configCache = make(map[uint32][]*ProvRecord)
  arg_len := len(args)
  for i := 1; i < arg_len; i++ {
    arg_action := DetectArg(i, args)
    if arg_action == nil {
      ArgPanic(nil, args)      
    }
    i += arg_action(i, args)
  }
  
}
