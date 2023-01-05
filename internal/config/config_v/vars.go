package config_v

// Глобальные переменные
var Args struct {
	Tar        string
	EpgSources []*ProvRecord
	Gzip       int
	MemDb      bool
}

// Структура, описывающая один элемент EPG
type ProvRecord struct {
	Id      string   `json:"id"`
	Urls    []string `json:"xmltv"`
	LastEpg uint64
	LastUpd uint64
	IdHash  uint32
}
