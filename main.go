package main

import (
	//"github.com/gotk3/gotk3/gtk"
	"scanner"
)

func main() {
	//ui()
	console()
}

func console() {
	// TODO: handle multiple devices
	idx := 0
	printInfo(idx)

	uatDev := &UAT{}
	if err := uatDev.sdrConfig(idx); err != nil {
		log.Fatalf("uatDev = &UAT{indexID: id} failed: %s\n", err.Error())
	}
	uatDev.wg = &sync.WaitGroup{}
	uatDev.wg.Add(1)
	log.Printf("\n======= CTRL+C to exit... =======\n\n")

	//go uatDev.read()
	//ch := make(chan, []byte)
	//go uatDev.scan(800*1000*1000, 1200*1000*1000, ch)

	uatDev.sigAbort()
}

/*
func ui() {
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.SetTitle("Band scanner")
	win.SetDefaultSize(800, 600)

	grid, err := gtk.GridNew()
	win.Add(grid)

	l, err := gtk.LabelNew("Hello, gotk3!")
	grid.Attach(l, 0, 0, 1, 1)

	// start
	win.ShowAll()
	gtk.Main()
}
*/


func printInfo(idx int) {
	m, p, s, err := rtl.GetDeviceUsbStrings(idx)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	log.Printf("Usb name: %s %s %s\n", m, p, s)
	log.Printf("Device name: %s\n", rtl.GetDeviceName(idx))
}

func init() {
	if c := rtl.GetDeviceCount(); c == 0 {
		log.Fatal("No device found")
	}
}
