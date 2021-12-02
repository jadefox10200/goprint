# goprint

See <a href="http://www.github.com/alexbrainman/printer">printer</a> repo for a much more user friendly printer. After many hours of trying to get my printer working, I ended up using some of the code found there so I could print PDFs on windows.

NOTE: This implementation heavily depends on the printer itself to draw complex document types such as PDFs as the code isn't doing any of the drawing. 
Therefore, try sending a simple text file to your printer first. This should work in all cases that I have tested. Then if more complex document types don't work, this will be due to the printer not being able to read and convert that document type itself. If posting issues, please try this first and specify which printer you're using. 

UPDATE: I switched to a Linux server and use CUPS with the proper drivers. This allows you to fully control the printer(duplex, stapling, etc) from a CLI.

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

UPDATE: 2021-03-05: New features added as well as new printing, duplex and setting default printer.  Existing functions were left alone and not modified so 
the package is safe to download without losing existing features. 

NOTE: See issues as of 2021-11-28. Branch outFix needs testing in order to confirm this is now handled. Once confirmed, it will be merged into master. Anyone with a windows printer is encouraged to please test this package from the outFix branch. Thank you. 
