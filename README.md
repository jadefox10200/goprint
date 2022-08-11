# goprint

UPDATE 11 August 2022: Hello! I just discovered today that there was a conflict of what was posted here in this repo and the code I had originally developed for the commit done on 2021-03-05. Therefore, the example in the examples folder obviously didn't work. I currently do not have a windows computer nor do I have access to a printer network so I can't verify that this code is correct and works. Therefore, a new branch was created named 'dev'. I encourage anyone who can to please download that branch and test it. It adds features which allow you to control setting printer options such as duplex, changing your default printer, etc. This repo may be useful to you but I cannot promise all features will work for you  there are many different kinds of printers and as I do not have the ability to test it. If you are testing the dev branch and run into issues, please post them and I will do my best to assist with any corretions.  

See <a href="http://www.github.com/alexbrainman/printer">printer</a> repo for a much more user friendly printer. After many hours of trying to get my printer working, I ended up using some of the code found there so I could print PDFs on windows.

NOTE: This implementation heavily depends on the printer itself to draw complex document types such as PDFs as the code isn't doing any of the drawing. 
Therefore, try sending a simple text file to your printer first. This should work in all cases that I have tested. Then if more complex document types don't work, this will be due to the printer not being able to read and convert that document type itself. If posting issues, please try this first and specify which printer you're using. 

In the end, I did not end up using this package for production and switched to a Linux server and use CUPS with the proper drivers. This allows you to fully control the printer(duplex, stapling, etc) from a CLI.

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
