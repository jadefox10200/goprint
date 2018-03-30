package goprint

import(

     "syscall"
     "unsafe"
     "io/ioutil"
     "fmt"
     "strings"
)
 
type DOC_INFO_1 struct{
     pDocName       *uint16
     pOutputFile    *uint16
     pDatatype      *uint16
}

type PRINTER_INFO_5 struct {
     PrinterName              *uint16
     PortName                 *uint16
     Attributes               uint32
     DeviceNotSelectedTimeout uint32
     TransmissionRetryTimeout uint32
}

var(
     dll = syscall.MustLoadDLL("winspool.drv")
     getDefaultPrinter = dll.MustFindProc("GetDefaultPrinterW")
     openPrinter = dll.MustFindProc("OpenPrinterW")
     startDocPrinter = dll.MustFindProc("StartDocPrinterW")
     startPagePrinter = dll.MustFindProc("StartPagePrinter")
     writePrinter = dll.MustFindProc("WritePrinter")
     endPagePrinter = dll.MustFindProc("EndPagePrinter")
     endDocPrinter = dll.MustFindProc("EndDocPrinter")
     closePrinter= dll.MustFindProc("ClosePrinter")    
     procEnumPrintersW = dll.MustFindProc("EnumPrintersW") 
)

//Opens a printer which can then be used to send documents to. Must be closed by user once. 
func GoOpenPrinter(printerName string) (uintptr, error) {

     // printerName, printerName16 := getDefaultPrinterName();     
     
     printerName16 := syscall.StringToUTF16(printerName)     
     printerHandle, err := openPrinterFunc(printerName, printerName16)      
     if err != nil {return 0, err}
     
     return printerHandle, nil
}
 
func GoPrint(printerHandle uintptr, path string) error {
     
     var err error

     startPrinter(printerHandle, path)
     startPagePrinter.Call(printerHandle)
     err = writePrinterFunc(printerHandle, path)
     endPagePrinter.Call(printerHandle)
     endDocPrinter.Call(printerHandle)
     
     return err
}

func GoClosePrinter(printerHandle uintptr) {

     closePrinter.Call(printerHandle)  
     
     return
}
 
func writePrinterFunc(printerHandle uintptr, path string) error {
     fmt.Println("About to write file to path: ", path)
     fileContents, err := ioutil.ReadFile(path)     
     if err != nil { return err }
     var contentLen uintptr = uintptr(len(fileContents))
     var writtenLen int
     _, _, err = writePrinter.Call(printerHandle, uintptr(unsafe.Pointer(&fileContents[0])),  contentLen, uintptr(unsafe.Pointer(&writtenLen)))
     fmt.Println("Writing to printer:", err)

     return nil
}
 
func startPrinter(printerHandle uintptr, path string) {     
          
     arr := strings.Split(path, "/")
     l := len(arr)
     name := arr[l-1]

     d := DOC_INFO_1{
          pDocName:    &(syscall.StringToUTF16(name))[0],
          pOutputFile: nil,
          pDatatype:   &(syscall.StringToUTF16("RAW"))[0],
     }     
     r1, r2, err := startDocPrinter.Call(printerHandle, 1, uintptr(unsafe.Pointer(&d)))
     fmt.Println("startDocPrinter: ", r1, r2, err)

     return
}
 
func openPrinterFunc(printerName string, printerName16 []uint16) (uintptr, error) {

     var printerHandle uintptr
     _, _, msg := openPrinter.Call(uintptr(unsafe.Pointer(&printerName16[0])), uintptr(unsafe.Pointer(&printerHandle)), 0)
     fmt.Println("open printer: ", msg)

     if printerHandle == 0 {return 0, fmt.Errorf("Couldn't find printer: printerName")}

     return printerHandle, nil

}
 
func GetDefaultPrinterName() (string, []uint16){

     var pn[256] uint16
     plen := len(pn)
     getDefaultPrinter.Call(uintptr(unsafe.Pointer(&pn)), uintptr(unsafe.Pointer(&plen)))
     printerName := syscall.UTF16ToString(pn[:])
     fmt.Println("Printer name:", printerName)     
     printer16 := syscall.StringToUTF16(printerName)     
     return printerName, printer16
}

func GetPrinterNames() ([]string, error) {     
     var needed, returned uint32
     buf := make([]byte, 1)
     err := enumPrinters(2, nil, 5, &buf[0], uint32(len(buf)), &needed, &returned)
     if err != nil {
          if err != syscall.ERROR_INSUFFICIENT_BUFFER {
               return nil, err
          }
          buf = make([]byte, needed)
          err = enumPrinters(2, nil, 5, &buf[0], uint32(len(buf)), &needed, &returned)
          if err != nil {
               return nil, err
          }
     }
     ps := (*[1024]PRINTER_INFO_5)(unsafe.Pointer(&buf[0]))[:returned]
     defaultPrinter, _ := GetDefaultPrinterName()
     names := make([]string, 0, returned)
     names = append(names, defaultPrinter)
     for _, p := range ps {
          v := (*[1024]uint16)(unsafe.Pointer(p.PrinterName))[:]
          printerName := syscall.UTF16ToString(v)
          if printerName == defaultPrinter {continue}
          names = append(names, printerName)
     }
     return names, nil
}

func enumPrinters(flags uint32, name *uint16, level uint32, buf *byte, bufN uint32, needed *uint32, returned *uint32) (err error) {
     r1, _, e1 := syscall.Syscall9(procEnumPrintersW.Addr(), 7, uintptr(flags), uintptr(unsafe.Pointer(name)), uintptr(level), uintptr(unsafe.Pointer(buf)), uintptr(bufN), uintptr(unsafe.Pointer(needed)), uintptr(unsafe.Pointer(returned)), 0, 0)
     if r1 == 0 {
          if e1 != 0 {
               err = error(e1)
          } else {
               err = syscall.EINVAL
          }
     }
     return
}
