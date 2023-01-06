package config

import (
	"flag"

	"ott-play-epg-converter/internal/config/config_v"
)

var (
	config_cache map[uint32][]*config_v.ProvRecord
	config_files []string
	config_raw   [][]byte
)


func ReadArgs() {
	config_files = make([]string, 0, 5)
	config_raw   = make([][]byte, 0, 5)
	config_cache = make(map[uint32][]*config_v.ProvRecord)

	flag.BoolVar(&config_v.Args.MemDb, "epg-ram", false, "process epg in the ram instead of the temp file")
	flag.StringVar(&config_v.Args.Tar, "tar", "", "output all files to tar archive")
	flag.IntVar(&config_v.Args.Gzip, "z", 0, "gzip tar with a specified level")
	flag.Func("c", "config file, syntax: file_name[,prov_name,...,prov_name] (allow multiple -c)", stackConfigArg)
	flag.Func("cp", "raw provider json config string, syntax: json_string (allow multiple -cp)", stackRawConfigArg)
	flag.Parse()

	// Немного магии с именем файла
	name_len := len(config_v.Args.Tar)
	// Добавление сжатия, если имя файла, .gz
	if name_len >= 3 {
		if config_v.Args.Tar[name_len-3:] == ".gz" {
			if config_v.Args.Gzip == 0 {
				config_v.Args.Gzip = 1
			}
		}
	}
}

func stackConfigArg(conf_val string) error {
	config_files = append(config_files, conf_val)
	return nil
}

func stackRawConfigArg(conf_val string) error {
	config_raw = append(config_raw, []byte(conf_val))
	return nil
}