package string_hashes

import (
  "github.com/cespare/xxhash/v2"
  xxhash32 "github.com/OneOfOne/xxhash"
  //"github.com/twmb/murmur3"
)

// С удалением unsafe бита, для корректной работы sqlite3
func HashSting64(v string) uint64 {
	return xxhash.Sum64String(v) & 0x7FFFFFFFFFFFFFFF
}

func HashSting32(v string) uint32 {
  //return murmur3.StringSum32(v)
  return xxhash32.ChecksumString32(v)
}

// JavaScript Number.MAX_SAFE_INTEGER
func HashSting52(v string) uint64 {
	return xxhash.Sum64String(v) & 0xFFFFFFFFFFFFF
}
