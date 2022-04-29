package prov_meta

import (
  "os"
  "time"
  "encoding/json"
  "github.com/rs/zerolog/log"
  "ott-play-epg-converter/lib/arg-reader"
  "ott-play-epg-converter/lib/string-hashes"
)

type ProvMeta struct {
  LastUpd  uint64 `json:"last-update"` 
  Urls   []uint32 `json:"url-hashes"`
}

const (
  providers_file = "providers.json"
)

var (
  ProvsMeta = make(map[string]*ProvMeta)
)


// Загрузка index файла по провайдерам
func Load() {
  jsonData, err := os.ReadFile(providers_file); if err != nil {
    log.Warn().Msg("Cannot load " + providers_file)
    return
  }
  
  if err := json.Unmarshal(jsonData, &ProvsMeta); err != nil  {
    log.Err(err).Send()
    return
  }
}

// Обновление записи о провайдере
func InitProv(p *arg_reader.ProvRecord, t uint64) {
  val, ok := ProvsMeta[p.Id]
  if !ok {
    val = &ProvMeta{}
    ProvsMeta[p.Id] = val
  }
  // Meta update
  if t < uint64(time.Now().Unix()) {
    log.Warn().Msgf("[%s] has epg from the past!", p.Id)
  }
  val.LastUpd = t
  val.Urls = make([]uint32, len(p.Urls))
  for i := 0; i < len(p.Urls); i++ {
    val.Urls[i] = string_hashes.HashSting32(p.Urls[i])
  }
}

// Сохранение index файла по провайдерам
func Save() {
  jsonData, err := json.Marshal(&ProvsMeta); if err != nil {
    log.Err(err).Send()
    return
  }
  err = os.WriteFile(providers_file, jsonData, 0644); if err != nil  {
    log.Err(err).Send()
    return
  }
}