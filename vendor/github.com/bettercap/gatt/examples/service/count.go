package service

import (
	"fmt"
	"log"
	"time"

	"github.com/bettercap/gatt"
)

func NewCountService() *gatt.Service {
	n := 0
	s := gatt.NewService(gatt.MustParseUUID("09fc95c0-c111-11e3-9904-0002a5d5c51b"))
	s.AddCharacteristic(gatt.MustParseUUID("11fac9e0-c111-11e3-9246-0002a5d5c51b")).HandleReadFunc(
		func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
			fmt.Fprintf(rsp, "count: %d", n)
			n++
		})

	s.AddCharacteristic(gatt.MustParseUUID("16fe0d80-c111-11e3-b8c8-0002a5d5c51b")).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			log.Println("Wrote:", string(data))
			return gatt.StatusSuccess
		})

	s.AddCharacteristic(gatt.MustParseUUID("1c927b50-c116-11e3-8a33-0800200c9a66")).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			cnt := 0
			for !n.Done() {
				fmt.Fprintf(n, "Count: %d", cnt)
				cnt++
				time.Sleep(time.Second)
			}
		})

	return s
}
