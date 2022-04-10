package arg_reader

import (
  "fmt"
  "os"
  "strconv"
  "strings"
  "encoding/json"
  "github.com/rs/zerolog/log"
)
// Структура, описывающая один элемент EPG
type ProvRecord struct {
  File   string `json:"file"`
  Id     string `json:"id"`
  IdHash uint32
  Order  uint8   `json:"order"`
}

type ArgData struct {
  MakeList    bool
  EpgSources  []*ProvRecord
}

var AppConfig ArgData

func ArgPanic(err error, args []string){
  if err != nil {
    log.Err(err).Send();
  }
  log.Panic().Msg(fmt.Sprint("Unrecognized command line: ", args[1:]))
}

func ProcessC(arg_pos int, args []string) int {
  // Хватает ли аргуметов
  arg_pos++
  if arg_pos > len(args)  { ArgPanic(nil, args) }
  
  // Обработка суб аргуметов
  sub_args := strings.Split(args[arg_pos], ",") 
  if len(sub_args) != 1 && len(sub_args) != 2 { ArgPanic(nil, args) }
  
  jsonData, err := os.ReadFile(sub_args[0])
  if err != nil  { ArgPanic(err, args) }
  
  config_data := []*ProvRecord{}
  err = json.Unmarshal(jsonData, &config_data)
  if err != nil  { ArgPanic(err, args) }
  
  var filter *string
  if len(sub_args) == 2  { filter = &sub_args[1] }
  
  // Обработка кофига (с фильтром)
  for i := 0; i < len(config_data); i++ {
    if filter != nil && *filter != config_data[i].Id { continue }
    AppConfig.EpgSources = append(AppConfig.EpgSources, config_data[i])
  }
  
  return 1
}

func ProcessE(arg_pos int, args []string) int {
  // Хватает ли аргуметов
  arg_pos++
  if arg_pos > len(args)  { ArgPanic(nil, args) }
  
  // Обработка суб аргуметов
  sub_args := strings.Split(args[arg_pos], ",") 
  if len(sub_args) != 2 && len(sub_args) != 3  { ArgPanic(nil, args) }
 
  prov := ProvRecord{ File: sub_args[0], Id: sub_args[1] }
  if len(sub_args) == 3 {
    _order, err := strconv.Atoi(sub_args[2])
    if err != nil   { ArgPanic(err, args) }
    prov.Order = uint8(_order)
  }
  AppConfig.EpgSources = append(AppConfig.EpgSources, &prov)
  return 1
}

func ProcessL(arg_pos int, args []string) int {
  const block_args = 0
  AppConfig.MakeList = true
  return 0
}

func DetectArg(num int, args []string) func(int, []string) int {
  	switch args[num] {
  	  case "-l": return ProcessL
    case "-e": return ProcessE
    case "-c": return ProcessC
    default:   return nil
	}
}

func ParseArgs(args []string) {
  arg_len := len(args)
  for i := 1; i < arg_len; i++ {
    arg_action := DetectArg(i, args)
    if arg_action == nil {
      ArgPanic(nil, args)      
    }
    i += arg_action(i, args)
  }
  
}
