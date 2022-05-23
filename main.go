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

const app_ver = "EPG converter for OTT-play FOSS v0.7.7"

func printHelp() {
  log.Error().Msg(`EPG converter for OTT-play FOSS
  Command line: <app> [--epg-ram] [-c OPTS]
  Main options:
    --epg-ram  process epg in RAM
    -с <opts>  parse epg files from json config
  NOTE: The character "," is a separator in the <opts>
 
  -c config_file[,prov_name]
    config_file  provider config file in json format
    prov_name    select only one provider from config
    
  Sample:
    Encode "it999" epg from "sample_config.json" file:
      ott-play-epg-converter -c sample_config.json,it999
    Encode ALL epg from "provs.json":
      ott-play-epg-converter -c provs.json
    More examples:
      cat epgone.xml | ott-play-epg-converter -c sample_config.json,intest
      zcat epgone.xml.gz | ott-play-epg-converter -c sample_config.json,intest
      curl --silent http://prov.host/epg.xml.gz | gzip -d -c - | ott-play-epg-converter -c sample_config.json,intest
      curl --silent --compressed http://prov.host/epg.xml | ott-play-epg-converter -c sample_config.json,intest
      ...
      ott-play-epg-converter -c sample_config.json,it999`)
}

func main() {
  tstart := time.Now()
  //doXml := false; doList := false
  log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02T15:04:05", NoColor: true})
  log.Info().Msg(app_ver)
  log.Info().Msg("  git@prog4food (c) 2o22")
  if len(os.Args) == 1 {
  // No args
    printHelp()
    return
  }
  // Лень обрабатывать args по-взрослому...
  app_config.ParseArgs(os.Args)
  
  // Database: Open
  db := xml_importer.SeedDB("chcache.db")
  defer db.Close()
  db.SetMaxIdleConns(1)
  //db.SetMaxOpenConns(1)
    
  // Database tune
  if _, err := db.Exec(`
  PRAGMA synchronous  = NORMAL;
  PRAGMA journal_mode = MEMORY;
  `); err != nil {
    log.Error().Err(err).Send()
  }
  
  // Загрузка общего мета-списка провайдеров
  json_exporter.ProvList_Load()
 
  // Processing EPG XMLs
  provConf := app_config.AppConfig.EpgSources
  for i := 0; i < len(provConf); i++ {
    // Prepare provider hash
    provConf[i].IdHash = helpers.HashSting32(provConf[i].Id)
    // Parse XML
    xml_importer.ProcessXml(db, provConf[i])
  }
  // Сохранение общего мета-списка провайдеров
  json_exporter.ProvList_Save()

  log.Info().Msgf("Total Execution time: %f", time.Since(tstart).Seconds())
}