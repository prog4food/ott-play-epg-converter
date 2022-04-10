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
  log.Error().Msg(`ott-play-FOSS EPG parser
  Command line: <app> [-l] [-e|-c OPTS]
  Main options:
    -e <opts>  parse epg file from arguments
    -с <opts>  parse epg files from json config
    -l         generate channel list
  NOTE: The character "," is a separator in the options
    
  -e xml_file|-,prov_id[,prov_order]
    xml_file   read data from specified <xml_file>
    -          read data from StdIn pipe input
    prov_id    provider id
    prov_order provider selection order (for auto search, default 50)
  
  -c prov_config[,prov_name]
    prov_config  profider config file in json format
    prov_name    select only one provider from config
    
  Sample:
    Encode epg from "epg.xml" file:
      ott-play-epg-preparator -e epg.xml,blabla
    Encode "bestprov" epg from "conf1.json" file:
      ott-play-epg-preparator -c conf1.json,bestprov
    Encode ALL epg from "provs.json" + make channel list:
      ott-play-epg-preparator -l -c provs.json
    Encode 4(or more) gzipped EPG and generate channel list at end:
      zcat epgone.xml.gz | ott-play-epg-preparator -e -,myfirstprov
      curl --silent http://prov.host/epg.xml.gz | gzip -d -c - | ott-play-epg-preparator -e -,otherprov1,49
      curl --silent --compressed http://prov.host/epg.xml | ott-play-epg-preparator -e -,otherprov2,51
      ...
      zcat epglast.xml.gz | ott-play-epg-preparator -l -e -,islastprov,99`)
}

func main() {
  tstart := time.Now()
  //doXml := false; doList := false
  log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02T15:04:05", NoColor: true})
  log.Info().Msg("EPG compiler for OTT-play FOSS v0.3.2")
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