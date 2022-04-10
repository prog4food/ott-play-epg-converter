package epg_jsoner

//easyjson:json
type EpgRecord struct {
  Name   string `json:"name"` 
  Time   uint64 `json:"time"` 
  TimeTo uint64 `json:"time_to"` 
  Descr  *string `json:"descr"` 
}

//easyjson:json
type EpgData struct {
  EpgData []EpgRecord `json:"epg_data"` 
}
