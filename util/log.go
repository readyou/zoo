package util

import "log"

func ConfigLog() {
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
}
