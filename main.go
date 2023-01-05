package main

import (
	"os"
  "io"
	"time"
  "runtime"

  "github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/internal/app_config"
	"ott-play-epg-converter/internal/helpers"
	"ott-play-epg-converter/internal/json_exporter"
	"ott-play-epg-converter/internal/xml_importer"
)

// Устанавливаются при сборке
var depl_ver = "[devel]"

func main() {
  var sOut io.Writer

  // Фикс цветной консоли для Windows
  if runtime.GOOS == "windows" { sOut = colorable.NewColorableStdout()
  } else { sOut = os.Stdout }
  //zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
  log.Logger = log.Output(zerolog.ConsoleWriter{Out: sOut, TimeFormat: "2006-01-02T15:04:05"})

  tstart := time.Now()
  app_config.ReadArgs()

  // Если вывод в tar и stdout, то логи пишем в StdErr
  if app_config.Args.Tar == "-" {
    if runtime.GOOS == "windows" { sOut = colorable.NewColorableStderr()
    } else { sOut = os.Stderr }
    log.Logger = log.Output(zerolog.ConsoleWriter{Out: sOut, TimeFormat: "2006-01-02T15:04:05"})
  }

  log.Info().Msg("EPG converter for OTT-play FOSS " + depl_ver)
  log.Info().Msg("  git@prog4food (c) 2o22\n")
  if len(os.Args) == 1 {
    // Запустили без аргументов
    log.Info().Msg("Run with -h for help")
    return
  }
  app_config.ReadConfigs()
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