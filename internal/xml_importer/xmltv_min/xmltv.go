package xmltv_min

/*
Minimal xmltv parsser, based on https://github.com/robbiet480/stirr-for-channels

optimized by prog4food, 2o22
*/

import (
	"encoding/xml"
	"errors"
	"time"
)

var ErrDateWrongFmt = errors.New("XMLTV: Wrong time wrong format")

type UnixTime int64

// UnmarshalXMLAttr is used to unmarshal a time in the XMLTV format to a time.Time.
func (t *UnixTime) UnmarshalXMLAttr(attr xml.Attr) error {
	var dlen = len(attr.Value)
	if dlen != 14 && dlen != 20 {
		return ErrDateWrongFmt
	}
	var (
		vY                 uint16
		vM, vD, vh, vm, vs uint8
		vTZ                int16
	)
	bs := []byte(attr.Value)
	for i := range bs {
		if bs[i] >= '0' && bs[i] <= '9' {
			bs[i] = bs[i] - '0'
		}
	}
	vY = uint16(bs[0])*1000 + // Год
				uint16(bs[1])*100 +
				uint16(bs[2])*10 +
				uint16(bs[3])
	vM = uint8(bs[4])*10 + uint8(bs[5])   // Месяц
	vD = uint8(bs[6])*10 + uint8(bs[7])   // День
	vh = uint8(bs[8])*10 + uint8(bs[9])   // Часы
	vm = uint8(bs[10])*10 + uint8(bs[11]) // Минуты
	vs = uint8(bs[12])*10 + uint8(bs[13]) // Секунды

	if dlen == 20 {
		vTZ = int16(bs[16])*10 + int16(bs[17])  // Часовой пояс (часы)
		vTZ *= 60                               // Переводим в минуты
		vTZ += int16(bs[18])*10 + int16(bs[19]) // Часовой пояс (добавляем минуты)

		switch bs[15] {
		case '+':
			vTZ *= -1
		case '-':
		default:
			return ErrDateWrongFmt
		}
	}
	*t = UnixTime(time.Date(int(vY), time.Month(vM), int(vD), int(vh), int(vm), int(vs), 0, time.UTC).Add(time.Duration(vTZ) * time.Minute).Unix())
	return nil
}

// CommonElement element structure that is common, i.e. <country lang="en">Italy</country>
type CommonElement struct {
	Lang  string `xml:"lang,attr,omitempty"`
	Value string `xml:",chardata"`
}

// Icon associated with the element that contains it
type Icon struct {
	Source string `xml:"src,attr"`
	Width  uint16 `xml:"width,attr,omitempty"`
	Height uint16 `xml:"height,attr,omitempty"`
}

// Length of the programme
type Length struct {
	Units string `xml:"units,attr"`
	Value string `xml:",chardata"`
}

// Channel details of a channel
type Channel struct {
	// XMLName      xml.Name        `xml:"channel"`
	ID           string          `xml:"id,attr"`
	DisplayNames []CommonElement `xml:"display-name"`
	Icons        []Icon          `xml:"icon,omitempty"`
}

// Programme details of a single programme transmission
type Prog struct {
	// XMLName         xml.Name         `xml:"programme"
	Channel      string          `xml:"channel,attr"`
	Length       *Length         `xml:"length,omitempty"`
	Start        *UnixTime       `xml:"start,attr"`
	Stop         *UnixTime       `xml:"stop,attr,omitempty"`
	Titles       []CommonElement `xml:"title"`
	Descriptions []CommonElement `xml:"desc,omitempty"`
	Icons        []Icon          `xml:"icon,omitempty"`
}
