package helpers

import (
	"strings"

	xxhash32 "github.com/OneOfOne/xxhash"
	"github.com/cespare/xxhash/v2"
	//"github.com/twmb/murmur3"
)

// С удалением unsafe бита, для корректной работы sqlite3
func HashSting64(v string) uint64 {
	return xxhash.Sum64String(v) & 0x7FFFFFFFFFFFFFFF
}

func HashSting64i(v string) uint64 {
  return xxhash.Sum64String(strings.ToLower(v)) & 0x7FFFFFFFFFFFFFFF
}

func HashSting32(v string) uint32 {
  //return murmur3.StringSum32(v)
  return xxhash32.ChecksumString32(v)
}

func HashSting32i(v string) uint32 {
  return xxhash32.ChecksumString32(strings.ToLower(v))
}

// JavaScript Number.MAX_SAFE_INTEGER
func HashSting52(v string) uint64 {
	return xxhash.Sum64String(v) & 0xFFFFFFFFFFFFF
}
