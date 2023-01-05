package helpers

func HasHTTP(u string) bool {
  if (len(u) > 9) {
    if ( u[0:8] == "https://" ) || ( u[0:7] == "http://" ) {
      return true
    }
  }
  return false
}

func CutHTTP(u string) string {
  if (len(u) > 9) {
    if u[0:8] == "https://" { u = u[8:]
    } else if u[0:7] == "http://"  { u = u[7:] }
  }
  return u
}