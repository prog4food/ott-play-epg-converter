package prov_meta

import (
	"encoding/json"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/lib/arg_reader"
)

type ProvMeta struct {
  Id        *string  `json:"id"`
  Urls     []uint32  `json:"url-hashes"`
  LastUpd    uint64  `json:"last-upd"`
  LastEpg    uint64  `json:"last-epg"`
}

type ProvMetaShort struct {
  LastUpd  uint64  `json:"last-upd"`
  LastEpg  uint64  `json:"last-epg"`
}

const (
  providers_file = "providers.json"
)

var (
  prov_list = make(map[string]*ProvMetaShort)
)


// Загрузка index файла по провайдерам
func Load() {
  jsonData, err := os.ReadFile(providers_file); if err != nil {
    log.Warn().Msg("Cannot load " + providers_file)
    return
  }
  
  if err := json.Unmarshal(jsonData, &prov_list); err != nil {
    log.Err(err).Send()
    return
  }
}

// Обновление записи о провайдере
func PushProv(p *arg_reader.ProvRecord, t uint64) {
  val, ok := prov_list[p.Id]
  if !ok {
    val = &ProvMetaShort{}
    prov_list[p.Id] = val
  }
  // Meta update
  val.LastUpd = uint64(time.Now().Unix())
  val.LastEpg = t

  if t < val.LastUpd {
    log.Warn().Msgf("[%s] has epg from the past!", p.Id)
  }
  // Prov meta update
  p.LastEpg, p.LastUpd = val.LastEpg, val.LastUpd
}

// Сохранение index файла по провайдерам
func Save() {
  jsonData, err := json.Marshal(&prov_list); if err != nil {
    log.Err(err).Send()
    return
  }
  err = os.WriteFile(providers_file, jsonData, 0644); if err != nil  {
    log.Err(err).Send()
    return
  }
}