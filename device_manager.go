package main

import (
	"errors"
)

const RANGE_SIZE int64 = 4 * 1024 * 1024

var DEVICES [3]string = [3]string{"mem", "ssd", "hdd"}

type DeviceManager struct {
	devices [len(DEVICES)]*Device
}

func NewDeviceManager() *DeviceManager {
	dm := new(DeviceManager)

	err := errors.New("no devices")
	for i, name := range DEVICES {
		dir := GConfig["cache."+name+".dir"].(string)
		cap := GConfig["cache."+name+".cap"].(int)
		if dm.devices[i] = NewDevice(dir, cap); dm.devices[i] != nil {
			err = nil
		}
	}

	Success(err)

	return dm
}

func (dm *DeviceManager) Close() {
	for _, d := range dm.devices {
		if d != nil {
			d.Close()
		}
	}
}

func (dm *DeviceManager) Alloc(item *Item) ([]byte, int) {
	for i := len(DEVICES) - 1; i > 0; i-- {
		if dm.devices[i] != nil {
			return dm.devices[i].Alloc(item), i
		}
	}
	return nil, -1
}

func (dm *DeviceManager) Add(dev int, k Key) {
	dm.devices[dev].Add(k)
}

func (dm *DeviceManager) Get(k Key) (*Item, []byte, *string) {
	for i, d := range dm.devices {
		if item, data := d.Get(k); item != nil {
			for j := i - 1; j >= 0; j-- {
				if dm.devices[j] != nil {
					new := *item
					buf := dm.devices[j].Alloc(&new)
					copy(buf, data)
					dm.devices[j].Add(k)
					break
				}
			}

			return item, data, &DEVICES[i]
		}
	}
	return nil, nil, nil
}

func (dm *DeviceManager) Del(match func(*Item) bool) uint {
	n := uint(0)
	for _, d := range dm.devices {
		n += d.DeleteBatch(match)
	}
	return n
}

func (dm *DeviceManager) DelPut(k Key) {
	for _, d := range dm.devices {
		d.DeletePut(k)
	}
}
