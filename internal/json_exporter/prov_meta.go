package json_exporter

import (
	"encoding/json"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/internal/config/config_v"
)

const providers_file = "providers.json"
type provMeta struct {
  Id        *string  `json:"id"`
  Urls     []uint32  `json:"url-hashes"`
  LastUpd    uint64  `json:"last-upd"`
  LastEpg    uint64  `json:"last-epg"`
}
type provMetaShort struct {
  LastUpd  uint64  `json:"last-upd"`
  LastEpg  uint64  `json:"last-epg"`
  IdHash   uint32  `json:"id-hash"`
}


var prov_list = make(map[string]*provMetaShort)


// Загрузка index файла по провайдерам
func ProvList_Load() {
  file_buf, err := os.ReadFile(providers_file); if err != nil {
    log.Warn().Msg("Cannot load " + providers_file)
    return
  }
  
  if err := json.Unmarshal(file_buf, &prov_list); err != nil {
    log.Err(err).Send()
    return
  }
}

// Обновление записи о провайдере
func ProvList_Update(p *config_v.ProvRecord, t uint64) {
  val, ok := prov_list[p.Id]
  if !ok {
    val = &provMetaShort{
      IdHash: p.IdHash,
    }
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
func ProvList_Save() {
  file_buf, err := json.Marshal(&prov_list); if err != nil {
    log.Err(err).Send()
    return
  }
  err = os.WriteFile(providers_file, file_buf, 0644); if err != nil  {
    log.Err(err).Send()
    return
  }
}