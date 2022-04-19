package main

import (
    "time"
    "os"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "ott-play-epg-converter/lib/arg-reader"
    "ott-play-epg-converter/lib/string-hashes"
)

func printHelp() {
  log.Error().Msg(`EPG converter for OTT-play FOSS
  Command line: <app> [-l] [-c OPTS]
  Main options:
    -с <opts>  parse epg files from json config
    -l         generate channel list
  NOTE: The character "," is a separator in the <opts>
 
  -c config_file[,prov_name]
    config_file  provider config file in json format
    prov_name    select only one provider from config
    
  Sample:
    Encode "it999" epg from "sample_config.json" file:
      ott-play-epg-converter -c sample_config.json,it999
    Encode ALL epg from "provs.json" + make channel list:
      ott-play-epg-converter -l -c provs.json
    More examples (and make channel list at end):
      cat epgone.xml | ott-play-epg-converter -c sample_config.json,intest
      zcat epgone.xml.gz | ott-play-epg-converter -c sample_config.json,intest
      curl --silent http://prov.host/epg.xml.gz | gzip -d -c - | ott-play-epg-converter -c sample_config.json,intest
      curl --silent --compressed http://prov.host/epg.xml | ott-play-epg-converter -c sample_config.json,intest
      ...
      ott-play-epg-converter -l -c sample_config.json,it999`)
}

func main() {
  tstart := time.Now()
  //doXml := false; doList := false
  log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02T15:04:05", NoColor: true})
  log.Info().Msg("EPG converter for OTT-play FOSS v0.4.0")
  log.Info().Msg("  git@prog4food (c) 2o22")
  if len(os.Args) == 1 {
  // No args
    printHelp()
    return
  }
  // Лень обрабатывать args по-взрослому...
  arg_reader.ParseArgs(os.Args)
  
  // Database: Open
  db := SeedDB("chcache.db")
  defer db.Close()
    
  // Database tune
  /*if _, err := db.Exec(`
  PRAGMA synchronous  = NORMAL;
  PRAGMA journal_mode = WAL;
  `); err != nil {
    log.Error().Err(err).Send()
  }*/
 
  // Processing EPG XMLs
  provConf := arg_reader.AppConfig.EpgSources
  for i := 0; i < len(provConf); i++ {
    // Prepare provider hash
    provConf[i].IdHash = string_hashes.HashSting32(provConf[i].Id)
    // Prov.Order default = 50
    if provConf[i].Order == 0 { provConf[i].Order = 50 }
    // Parse XML
    processXml(db, provConf[i])
  }

  if arg_reader.AppConfig.MakeList {
    log.Info().Msg("Creating channel map")
    
  }
  log.Info().Msgf("Total Execution time: %f", time.Since(tstart).Seconds())
}