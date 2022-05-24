package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/lib/app_config"
	"ott-play-epg-converter/lib/helpers"
	"ott-play-epg-converter/lib/json_exporter"
	"ott-play-epg-converter/lib/xml_importer"
)

// Устанавливаются при сборке
var depl_ver string

func main() {
  tstart := time.Now()
  app_config.ReadArgs()

  // Если вывод в tar -, то логи пишем в StdErr
  var log_output *os.File
  if app_config.Args.Tar == "-" {
    log_output = os.Stderr
  } else {
    log_output = os.Stdout }
  log.Logger = log.Output(zerolog.ConsoleWriter{Out: log_output, TimeFormat: "2006-01-02T15:04:05", NoColor: true})

  log.Info().Msg("EPG converter for OTT-play FOSS " + depl_ver)
  log.Info().Msg("  git@prog4food (c) 2o22")
  if len(os.Args) == 1 {
    // Запустили без аргументов
    log.Info().Msg("Run with -h for help")
    return
  }
  if len(app_config.Args.EpgSources) == 0 {
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
  provConf := app_config.Args.EpgSources
  for i := 0; i < len(provConf); i++ {
    // Prepare provider hash
    provConf[i].IdHash = helpers.HashSting32(provConf[i].Id)
    // Parse XML
    xml_importer.ProcessXml(db, provConf[i])
  }

  // Сохранение общего мета-списка провайдеров
  json_exporter.ProvList_Save()
  // Закрытие TAR файла и сопутствующих объектов
  json_exporter.CloseTarOutput()

  log.Info().Msgf("Total Execution time: %f", time.Since(tstart).Seconds())
}