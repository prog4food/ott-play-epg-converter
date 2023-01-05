package app_config

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hjson/hjson-go/v4"
	"github.com/rs/zerolog/log"

	"ott-play-epg-converter/internal/helpers"
)

// Структура, описывающая один элемент EPG
type ProvRecord struct {
	Id      string   `json:"id"`
	Urls    []string `json:"xmltv"`
	LastEpg uint64
	LastUpd uint64
	IdHash  uint32
}

var Args struct {
	Tar        string
	EpgSources []*ProvRecord
	Gzip       int
	MemDb      bool
}
var (
	config_cache map[uint32][]*ProvRecord
	config_files []string
)

func ReadArgs() {
	config_files = make([]string, 0, 5)
	config_cache = make(map[uint32][]*ProvRecord)

	flag.BoolVar(&Args.MemDb, "epg-ram", false, "process epg in the ram instead of the file")
	flag.StringVar(&Args.Tar, "tar", "", "output all files to tar archive")
	flag.IntVar(&Args.Gzip, "z", 0, "gzip tar with a specified level")
	flag.Func("c", "config file, syntax: file_name[,prov_name,...,prov_name] (allow multiple -c)", stackConfigArg)
	flag.Parse()

	// Немного магии с именем файла
	name_len := len(Args.Tar)
	// Добавление сжатия, если имя файла, .gz
	if name_len >= 3 {
		if Args.Tar[name_len-3:] == ".gz" {
			if Args.Gzip == 0 {
				Args.Gzip = 1
			}
		}
	}
}

func stackConfigArg(conf_val string) error {
	config_files = append(config_files, conf_val)
	return nil
}

func ReadConfigs() {
	var err error
	for i := 0; i < len(config_files); i++ {
		err = processConfigArg(config_files[i])
		if err != nil {
			log.Err(err).Send()
		}
	}
}

func processConfigArg(conf_val string) error {
	var (
		conf_data []byte
		err       error
		_config   []*ProvRecord
	)
	// Обработка доп. параметров
	sub_args := strings.Split(conf_val, ",")

	fname := sub_args[0]
	fname_h := helpers.HashSting32(fname)
	if conf, is_cached := config_cache[fname_h]; is_cached {
    // Конфиг в кеше
		log.Debug().Msgf("Cached config: %s", fname)
		_config = conf
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

		_config = []*ProvRecord{}
		if err = hjson.Unmarshal(conf_data, &_config); err != nil {
			return err
		}
		config_cache[fname_h] = _config
	}

	sub_args_len := len(sub_args)
	var _el *ProvRecord
	if sub_args_len > 1 {
		// Обработка кофига: с фильтром
		log.Debug().Msgf("  filter by: %v", sub_args[1:])
	read_config:
		for i := 1; i < sub_args_len; i++ {
			for _, _el = range _config {
				if sub_args[i] == _el.Id {
					Args.EpgSources = append(Args.EpgSources, _el)
					continue read_config
				}
			}
			log.Warn().Msgf("filter [%s] not found in config", sub_args[i])
		}
	} else {
		// Обработка конфига: без фильтра
		for _, _el = range _config {
			Args.EpgSources = append(Args.EpgSources, _el)
		}
	}

	return nil
}
