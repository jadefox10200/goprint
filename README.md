# goprint

Example usage:

```package main

import (
	"log"
	"github.com/jadefox10200/goprint"
)

func main() {

	printerName, _ := goprint.GetDefaultPrinterName()

	//open the printer
	printerHandle, err := goprint.GoOpenPrinter(printerName)	
	if err != nil {log.Fatalln("Failed to open printer")}
	defer goprint.GoClosePrinter(printerHandle)
	
	filePath := "C:/test/myPDF.pdf"
		
	//Send to printer:		
	err = goprint.GoPrint(printerHandle, filePath)
	if err != nil {	log.Fatalln("during the func sendToPrinter, there was an error") }


}
```
