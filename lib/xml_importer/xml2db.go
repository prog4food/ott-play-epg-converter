package xml_importer

import (
	"regexp"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/import/robbiet480/xmltv"
	"ott-play-epg-converter/lib/app_config"
	"ott-play-epg-converter/lib/helpers"
)

var (
  sting_flater = regexp.MustCompile(`\r?\n`)
)

// XML2SQL: Кешируем запись <channel>
// Берем Id, DisplayName[*] и Icon[0]
func NewChannelCache(ch *xmltv.Channel, prov *app_config.ProvRecord) {
  var err error
  
  // 2SQL: dedup Id канала
  h_id   := helpers.HashSting32i(ch.ID)
  if _, err = DbPre.Ch_ids.Exec(h_id, ch.ID); err != nil {
    log.Err(err).Send()
  }
  // 2SQL: dedup Icon[0] (только первая)
  h_icon := uint64(0)
  if len(ch.Icons) > 0  {
    h_icon = helpers.HashSting64(ch.Icons[0].Source)
    if _, err = DbPre.Ch_icons.Exec(h_icon, ch.Icons[0].Source); err != nil {
      log.Err(err).Send()
      }
  }
  // Обход <display-name>
  h_name := uint64(0)
  names_len := len(ch.DisplayNames)
  if names_len == 0 {
    log.Error().Msgf("Channel %s has no display names", ch.ID)
    // 2SQL: Связи
    if _, err = DbPre.Ch_data.Exec(prov.IdHash, h_id, h_name, h_icon); err != nil {
      log.Err(err).Send()
      }
  }
  for i := 0; i < names_len; i++ {
    // 2SQL: dedup Название
    h_name = helpers.HashSting64i(ch.DisplayNames[i].Value)
    if _, err = DbPre.Ch_names.Exec(h_name, ch.DisplayNames[i].Value); err != nil {
      log.Err(err).Send()
      }
    // 2SQL: Связи
    if _, err = DbPre.Ch_data.Exec(prov.IdHash, h_id, h_name, h_icon); err != nil {
      log.Err(err).Send()
      }
  }
}

// XML2SQL: Кешируем запись <programme>
// Берем только Title[0] и Desc[0]
func NewProgCache(pr *xmltv.Programme, prov *app_config.ProvRecord) {
  var err error
  // Проверки
  if len(pr.Titles) == 0 || pr.Start == nil || pr.Stop == nil {
    log.Warn().Msgf("[%s] bad programme record", pr.Channel)
    return
  }
  // Хеширование
  h_ch_id := helpers.HashSting32i(pr.Channel)
  // 2SQL: dedup Название[0] (только первое)
  h_title := helpers.HashSting64(pr.Titles[0].Value)
  if _, err = DbPre.Epg_title.Exec(h_title, pr.Titles[0].Value); err != nil {
    log.Err(err).Send()
	}
  // 2SQL: dedup Описание
  h_desc  := uint64(0)
  if len(pr.Descriptions) > 0  {
    h_desc = helpers.HashSting64(pr.Descriptions[0].Value)
    if h_title != h_desc  {
      flat_string := sting_flater.ReplaceAllString(pr.Descriptions[0].Value, "<br/>")
      if _, err = DbPre.Epg_desc.Exec(h_desc, flat_string); err != nil {
        log.Err(err).Send()
    		}
    } else {
      // Описание дублирует название, пропускаем
      h_desc = 0
    }
  }
  // 2SQL: dedup Постер[0] (только первый)
  h_icon := uint64(0)
  if len(pr.Icons) > 0  {
    h_icon = helpers.HashSting64(pr.Icons[0].Source)
    if _, err = DbPre.Epg_icon.Exec(h_icon, pr.Icons[0].Source); err != nil {
      log.Err(err).Send()
  	  }
  }
  // 2SQL: Связи
  DbPre.Epg_data.Exec(prov.IdHash, h_ch_id, pr.Start.Unix(), pr.Stop.Unix(), h_title, h_desc, h_icon)
}
