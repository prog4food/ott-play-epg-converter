package xml_importer

import (
	"regexp"

	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/pkg/robbiet480/xmltv"
	"ott-play-epg-converter/internal/app_config"
	"ott-play-epg-converter/internal/helpers"
)

var (
  sting_flater = regexp.MustCompile(`\r?\n`)
)

// XML2SQL: Кешируем запись <channel>
// Берем Id, DisplayName[*] и Icon[0]
func NewChannelCache(ch *xmltv.Channel, prov *app_config.ProvRecord) {
  // 2SQL: dedup Id канала
  h_id := helpers.HashSting32i(ch.ID)
  InsertKV(DbPre.Ch_ids, h_id, ch.ID)
  // 2SQL: dedup Icon[0] (только первая)
  h_icon := uint64(0)
  if len(ch.Icons) > 0  {
    h_icon = helpers.HashSting64(ch.Icons[0].Source)
    InsertKV(DbPre.Ch_icons, h_icon, ch.Icons[0].Source)
  }
  // Обход <display-name>
  h_name := uint64(0)
  names_len := len(ch.DisplayNames)
  if names_len == 0 {
    log.Error().Msgf("Channel %s has no display names", ch.ID)
    // 2SQL: Связи
    InsertCh(DbPre.Ch_data, prov.IdHash, h_id, h_name, h_icon)
  }
  for i := 0; i < names_len; i++ {
    // 2SQL: dedup Название
    h_name = helpers.HashSting64i(ch.DisplayNames[i].Value)
    InsertKV(DbPre.Ch_names, h_name, ch.DisplayNames[i].Value)
    // 2SQL: Связи
    InsertCh(DbPre.Ch_data, prov.IdHash, h_id, h_name, h_icon)
  }
}

// XML2SQL: Кешируем запись <programme>
// Берем только Title[0] и Desc[0]
func NewProgCache(pr *xmltv.Programme, prov *app_config.ProvRecord) {
  // Проверки
  if len(pr.Titles) == 0 || pr.Start == nil || pr.Stop == nil {
    log.Warn().Msgf("[%s] bad programme record", pr.Channel)
    return
  }
  // Хеширование
  h_ch_id := helpers.HashSting32i(pr.Channel)
  // 2SQL: dedup Название[0] (только первое)
  h_title := helpers.HashSting64(pr.Titles[0].Value)
  InsertKV(DbPre.Epg_title, h_title, pr.Titles[0].Value)
  // 2SQL: dedup Описание
  h_desc  := uint64(0)
  if len(pr.Descriptions) > 0  {
    h_desc = helpers.HashSting64(pr.Descriptions[0].Value)
    if h_title != h_desc  {
      flat_string := sting_flater.ReplaceAllString(pr.Descriptions[0].Value, "<br/>")
      InsertKV(DbPre.Epg_desc, h_desc, flat_string)
    } else {
      // Описание дублирует название, пропускаем
      h_desc = 0
    }
  }
  // 2SQL: dedup Постер[0] (только первый)
  h_icon := uint64(0)
  if len(pr.Icons) > 0  {
    h_icon = helpers.HashSting64(pr.Icons[0].Source)
    InsertKV(DbPre.Epg_icon, h_icon, pr.Icons[0].Source)
  }
  // 2SQL: Связи
  DbPre.Epg_data.BindInt64(1, int64(prov.IdHash))
  DbPre.Epg_data.BindInt64(2, int64(h_ch_id))
  DbPre.Epg_data.BindInt64(3, pr.Start.Unix())
  DbPre.Epg_data.BindInt64(4, pr.Stop.Unix())
  DbPre.Epg_data.BindInt64(5, int64(h_title))
  DbPre.Epg_data.BindInt64(6, int64(h_desc))
  DbPre.Epg_data.BindInt64(7, int64(h_icon))
  insert_exec(DbPre.Epg_data)
}
