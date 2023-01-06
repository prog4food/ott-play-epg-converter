package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hjson/hjson-go/v4"
	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/internal/config/config_v"
	"ott-play-epg-converter/internal/helpers"
)

// Обработчик аргументов
func ReadConfigs() {
	var err error
	var i 	int

	for i = range config_files {
		err = processConfigArg(config_files[i])
		if err != nil {
			log.Err(err).Msgf("config: file error - %s", config_files[i])
		}
	}

	for i = range config_raw {
		err = processProviderConfigArg(config_raw[i])
		if err != nil {
			log.Err(err).Msgf("config: raw error - %s", config_raw[i])
		}
	}	
}

// Разбирает конфиг провайдера из аргумента
func processProviderConfigArg(conf_data []byte) error {
	var provCfg = config_v.ProvRecord{}	
	var err = hjson.Unmarshal(conf_data, &provCfg); if err != nil {
		return err
	}
	config_v.Args.EpgSources = append(config_v.Args.EpgSources, &provCfg)
	return nil
}


// Читает один конфиг файл
func processConfigArg(conf_val string) error {
	var (
		conf_data []byte
		err       error
		provCfgs   []*config_v.ProvRecord
	)
	// Обработка доп. параметров
	conFilter := strings.Split(conf_val, ",")

	fname := conFilter[0]
	fname_h := helpers.HashSting32(fname)
	if conf, is_cached := config_cache[fname_h]; is_cached {
    // Конфиг в кеше
		log.Debug().Msgf("Cached config: %s", fname)
		provCfgs = conf
	} else { 
    // Читаем конфиг
		log.Debug().Msgf("Load config: %s", fname)
		if helpers.HasHTTP(fname) {
			// Конфиг указан по ссылке
			var resp *http.Response

			if resp, err = http.Get(fname); err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf("Cannot download config. %d:%s", resp.StatusCode, resp.Status)
			}
			conf_data, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {  // Если была ошибка чтения, вспоминаем о ней
				return err
			}
		} else {    // Конфиг указан файлом
			conf_data, err = os.ReadFile(fname)
		}
		if err != nil {
			return err
		}

		provCfgs = []*config_v.ProvRecord{}
		if err = hjson.Unmarshal(conf_data, &provCfgs); err != nil {
			return err
		}
		config_cache[fname_h] = provCfgs
	}

	sub_args_len := len(conFilter)
	var provCfg *config_v.ProvRecord
	if sub_args_len > 1 {
		// Кофиг с фильтром
		log.Debug().Msgf("  filter by: %v", conFilter[1:])
		
		var i int
		conf_loop:
		for _, provCfg = range provCfgs {
			for i = 1; i < sub_args_len; i++ {
				if provCfg.Id == conFilter[i] {
					config_v.Args.EpgSources = append(config_v.Args.EpgSources, provCfg)
					continue conf_loop
				}
			}
		}
	} else {	
		// Конфиг без фильтра
		for _, provCfg = range provCfgs {
			config_v.Args.EpgSources = append(config_v.Args.EpgSources, provCfg)
		}
	}
	return nil
}
