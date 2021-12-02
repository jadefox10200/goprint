package main

import (
	"fmt"

	"github.com/jadefox10200/goprint"
)

func main() {

	//get printers:
	printers, err := goprint.EnumPrinters2()
	if err != nil {
		fmt.Printf("Failed to enum: %s", err.Error())
		return
	}

	//List the printers in console:
	for _, v := range printers {

		fmt.Printf("%s :: %s :: %s\n", v.GetPrinterName(), v.GetServerName(), v.GetPortName())
	}

	//get the default pritner name:
	printerName, _ := goprint.GetDefaultPrinterName()

	//open the printer:
	hdl, err := goprint.OpenPrinter(printerName)
	if err != nil {
		fmt.Printf("Failed to open printer: %s", err.Error())
		return
	}

	//Changes user print settings to duplex:
	err = hdl.SetDuplexPrinter9(2)
	if err != nil {
		fmt.Printf("Failed to set printer: %s", err.Error())
		return
	}

	//Print using the handler:
	err = hdl.Print("test.txt")
	if err != nil {
		fmt.Printf("Failed to print using hdl: %s", err.Error())
		return
	}

	//To use LPR we need basic info about the printer:
	ptr5, err := hdl.GetPrinter5()
	if err != nil {
		fmt.Printf("Failed to get printer info5: %s", err.Error())
		return
	}

	//Print using simple lpr:
	err = goprint.PrintLPR("test.txt", ptr5.GetPrinterName(), ptr5.GetPortName())
	if err != nil {
		fmt.Printf("Failed to print using lpr: %s", err.Error())
		return
	}

}
