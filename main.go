package main

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/internal/config"
	"ott-play-epg-converter/internal/config/config_v"
	"ott-play-epg-converter/internal/helpers"
	"ott-play-epg-converter/internal/json_exporter"
	"ott-play-epg-converter/internal/xml_importer"
)

// Устанавливаются при сборке
var depl_ver = "[devel]"

func main() {
  tstart := time.Now()
  config.ReadArgs()

  helpers.InitLogger("EPG converter for OTT-play FOSS ", depl_ver, (config_v.Args.Tar == "-") )
  if len(os.Args) == 1 {
    // Запустили без аргументов
    log.Info().Msg("Run with -h for help")
    return
  }

  config.ReadConfigs()
  if len(config_v.Args.EpgSources) == 0 {
    // Нечего обрабатывать
    log.Error().Msg("No sources, check config file and arguments")
    return
  }

  // Database: Open
  db := xml_importer.SeedDB()
  defer db.Close()

  // Database tune
  helpers.SimpleExec(db, "PRAGMA journal_mode = MEMORY; PRAGMA synchronous  = NORMAL;", "main database tune error")
  
  // Загрузка общего мета-списка провайдеров
  json_exporter.ProvList_Load()
  // Если необходимо, перенаправление вывода в TAR
  json_exporter.SetTarOutput()
 
  // Processing EPG XMLs
  provConf := config_v.Args.EpgSources
  for i := 0; i < len(provConf); i++ {
    // Prepare provider hash
    provConf[i].IdHash = helpers.HashSting32(provConf[i].Id)
    // Parse XML
    err := xml_importer.ProcessXml(db, provConf[i])
    if err != nil {
      log.Err(err).Msgf("[%s]", provConf[i].Id)
    }
  }

  // Сохранение общего мета-списка провайдеров
  json_exporter.ProvList_Save()
  // Закрытие TAR файла и сопутствующих объектов
  json_exporter.CloseTarOutput()

  log.Info().Msgf("Total Execution time: %f", time.Since(tstart).Seconds())
}