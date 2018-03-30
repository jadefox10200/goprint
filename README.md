# goprint

See <a href="http://www.github.com/alexbrainman/printer">printer</a> repo for a much more user friendly printer. After many hours of trying to get my printer working, I ended up using some of the code found there so I could print PDFs on windows.


Example usage:

```go
package main

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
