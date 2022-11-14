package json_exporter

//easyjson:json
type EpgRecord struct {
  Name    string `json:"name"` 
  Descr   string `json:"descr"`
  Icon    string `json:"icon,omitempty"` 
  Time    uint64 `json:"time"` 
  TimeTo  uint64 `json:"time_to"` 
}

//easyjson:json
type EpgData struct {
  EpgData []EpgRecord `json:"epg_data"` 
}

//easyjson:json
type ChListData []*string
