package usbtool

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/gousb"
	"github.com/google/gousb/usbid"
)

var ErrDeviceNotFound error = errors.New("device not found")
var ErrSearchTermTooBroad error = errors.New("search term too broad, multiple devices found")

type Device struct {
	Path        string
	VendorID    string
	VendorName  string
	ProductId   string
	ProductName string
}

func (d Device) String() string {
	if d.Path != "" {
		return fmt.Sprintf("%s:%s %s (%s) at %s", d.VendorID, d.ProductId, d.ProductName, d.VendorName, d.Path)

	}
	return fmt.Sprintf("%s:%s %s (%s)", d.VendorID, d.ProductId, d.ProductName, d.VendorName)
}

func (d *Device) GetPath() {
	target := "PRODUCT=" + strings.TrimLeft(d.VendorID, "0") + "/" + strings.TrimLeft(d.ProductId, "0")

	usbEntries, err := ioutil.ReadDir("/sys/bus/usb/devices")
	if err != nil {
		return
	}

	var ttyDirs []string
	for _, entry := range usbEntries {
		temp := "/sys/bus/usb/devices/" + entry.Name() + "/uevent"
		_, err := os.Stat(temp)
		if err == nil && fileContains(temp, target) {
			temp = "/sys/bus/usb/devices/" + entry.Name() + "/tty"
			found, err := ioutil.ReadDir(temp)
			if err == nil {
				for _, f := range found {
					ttyDirs = append(ttyDirs, f.Name())
				}
			}

		}
	}

	if len(ttyDirs) == 1 {
		d.Path = ttyDirs[0]
	}
}

func GetAllDevices() (ret []Device) {
	ctx := gousb.NewContext()
	defer ctx.Close()
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		vendorName := "Unknown"
		productName := "Unknown"

		vendor, found := usbid.Vendors[desc.Vendor]
		if found {
			vendorName = vendor.Name
			product, found := vendor.Product[desc.Product]
			if found {
				productName = product.Name
			}
		}

		device := Device{
			Path:        "",
			VendorID:    desc.Vendor.String(),
			VendorName:  vendorName,
			ProductId:   desc.Product.String(),
			ProductName: productName,
		}
		device.GetPath()
		ret = append(ret, device)
		return false
	})

	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()
	if err != nil {
		log.Fatalf("list: %s", err)
	}

	return ret
}

func SearchDevice(term string) (Device, error) {
	var matches []Device
	devices := GetAllDevices()
	for _, d := range devices {
		name := strings.ToLower(d.ProductName)
		vendor := strings.ToLower(d.VendorName)

		if strings.Contains(name, strings.ToLower(term)) || strings.
			Contains(d.ProductId, strings.ToLower(term)) || strings.
			Contains(vendor, strings.ToLower(term)) || strings.
			Contains(d.VendorID, strings.ToLower(term)) {
			matches = append(matches, d)

		}
	}
	if len(matches) == 0 {
		return Device{}, ErrDeviceNotFound
	}
	if len(matches) > 1 {
		return Device{}, ErrSearchTermTooBroad
	}
	return matches[0], nil
}

func fileContains(filepath string, sub string) bool {
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return strings.Contains(string(contents), sub)
}
