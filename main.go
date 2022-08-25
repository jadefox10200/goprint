package goprint

import(          
     "reflect"
     "syscall"
     "unsafe"
     "io/ioutil"
     "fmt"
     "strings"
)

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
     documentProperties = dll.MustFindProc("DocumentPropertiesW")
     getPrinter = dll.MustFindProc("GetPrinterW")
     setPrinter = dll.MustFindProc("SetPrinterW")
)

// DOCINFO struct.
type DocInfo struct {
     cbSize       int32
     lpszDocName  *uint16
     lpszOutput   *uint16
     lpszDatatype *uint16
     fwType       uint32
}
 
type DOC_INFO_1 struct{
     pDocName       *uint16
     pOutputFile    *uint16
     pDatatype      *uint16
}

type PRINTER_INFO_5 struct{
     PrinterName              *uint16
     PortName                 *uint16
     Attributes               uint32
     DeviceNotSelectedTimeout uint32
     TransmissionRetryTimeout uint32
}

// PRINTER_INFO_2 attribute values
const (
     PRINTER_ATTRIBUTE_QUEUED            uint32 = 0x00000001
     PRINTER_ATTRIBUTE_DIRECT            uint32 = 0x00000002
     PRINTER_ATTRIBUTE_DEFAULT           uint32 = 0x00000004
     PRINTER_ATTRIBUTE_SHARED            uint32 = 0x00000008
     PRINTER_ATTRIBUTE_NETWORK           uint32 = 0x00000010
     PRINTER_ATTRIBUTE_HIDDEN            uint32 = 0x00000020
     PRINTER_ATTRIBUTE_LOCAL             uint32 = 0x00000040
     PRINTER_ATTRIBUTE_ENABLE_DEVQ       uint32 = 0x00000080
     PRINTER_ATTRIBUTE_KEEPPRINTEDJOBS   uint32 = 0x00000100
     PRINTER_ATTRIBUTE_DO_COMPLETE_FIRST uint32 = 0x00000200
     PRINTER_ATTRIBUTE_WORK_OFFLINE      uint32 = 0x00000400
     PRINTER_ATTRIBUTE_ENABLE_BIDI       uint32 = 0x00000800
     PRINTER_ATTRIBUTE_RAW_ONLY          uint32 = 0x00001000
     PRINTER_ATTRIBUTE_PUBLISHED         uint32 = 0x00002000
)

// PRINTER_INFO_2 status values.
const (
     PRINTER_STATUS_PAUSED               uint32 = 0x00000001
     PRINTER_STATUS_ERROR                uint32 = 0x00000002
     PRINTER_STATUS_PENDING_DELETION     uint32 = 0x00000004
     PRINTER_STATUS_PAPER_JAM            uint32 = 0x00000008
     PRINTER_STATUS_PAPER_OUT            uint32 = 0x00000010
     PRINTER_STATUS_MANUAL_FEED          uint32 = 0x00000020
     PRINTER_STATUS_PAPER_PROBLEM        uint32 = 0x00000040
     PRINTER_STATUS_OFFLINE              uint32 = 0x00000080
     PRINTER_STATUS_IO_ACTIVE            uint32 = 0x00000100
     PRINTER_STATUS_BUSY                 uint32 = 0x00000200
     PRINTER_STATUS_PRINTING             uint32 = 0x00000400
     PRINTER_STATUS_OUTPUT_BIN_FULL      uint32 = 0x00000800
     PRINTER_STATUS_NOT_AVAILABLE        uint32 = 0x00001000
     PRINTER_STATUS_WAITING              uint32 = 0x00002000
     PRINTER_STATUS_PROCESSING           uint32 = 0x00004000
     PRINTER_STATUS_INITIALIZING         uint32 = 0x00008000
     PRINTER_STATUS_WARMING_UP           uint32 = 0x00010000
     PRINTER_STATUS_TONER_LOW            uint32 = 0x00020000
     PRINTER_STATUS_NO_TONER             uint32 = 0x00040000
     PRINTER_STATUS_PAGE_PUNT            uint32 = 0x00080000
     PRINTER_STATUS_USER_INTERVENTION    uint32 = 0x00100000
     PRINTER_STATUS_OUT_OF_MEMORY        uint32 = 0x00200000
     PRINTER_STATUS_DOOR_OPEN            uint32 = 0x00400000
     PRINTER_STATUS_SERVER_UNKNOWN       uint32 = 0x00800000
     PRINTER_STATUS_POWER_SAVE           uint32 = 0x01000000
     PRINTER_STATUS_SERVER_OFFLINE       uint32 = 0x02000000
     PRINTER_STATUS_DRIVER_UPDATE_NEEDED uint32 = 0x04000000
)

// PRINTER_INFO_2 struct.
type PRINTER_INFO_2 struct {
     pServerName         *uint16
     pPrinterName        *uint16
     pShareName          *uint16
     pPortName           *uint16
     pDriverName         *uint16
     pComment            *uint16
     pLocation           *uint16
     pDevMode            *DevMode
     pSepFile            *uint16
     pPrintProcessor     *uint16
     pDatatype           *uint16
     pParameters         *uint16
     pSecurityDescriptor uintptr
     attributes          uint32
     priority            uint32
     defaultPriority     uint32
     startTime           uint32
     untilTime           uint32
     status              uint32
     cJobs               uint32
     averagePPM          uint32
}

type PRINTER_INFO_9 struct {
     dev *DevMode
}

type HANDLE uintptr

func (pi *PRINTER_INFO_2) GetDataType() string{

     return utf16PtrToString(pi.pDatatype)
}

func (hPrinter *HANDLE) Print(path string) error {
     pathArray := strings.Split(path, "/")
     l := len(pathArray)
     name := pathArray[l-1]
     ptr2, err := hPrinter.GetPrinter2()
     dataType := ptr2.GetDataType()
     d := DOC_INFO_1{
          pDocName:      &(syscall.StringToUTF16(name))[0],
          pOutputFile:   nil,
          pDatatype:     &(syscall.StringToUTF16(dataType))[0],
     }

     //Start the documnet - If function fails, return is 0
     r1, _, err := startDocPrinter.Call(uintptr(*hPrinter), 1, uintptr(unsafe.Pointer(&d)))
     if r1 == 0 {return err}

     //Start the page for printing. - If function fails, return is 0
     r1, _, err = startPagePrinter.Call(uintptr(*hPrinter))
     if r1 == 0 {return err}

     fc, err := ioutil.ReadFile(path)
     if err != nil {return err}

     var clen uintptr = uintptr(len(fc))
     var writtenLen int

     //Write to printer - If function fails, the return value is 0
     r1, _, err = writePrinter.Call(uintptr(*hPrinter), uintptr(unsafe.Pointer(&fc[0])), clen, uintptr(unsafe.Pointer(&writtenLen)))
     if r1 == 0 {return err}  

     //End the page - if function fails, the return value is 0
     r1, _, err = endPagePrinter.Call(uintptr(*hPrinter))
     if r1 == 0 {return err}

     //End the document - if function fails, the return value is 0
     r1, _, err = endDocPrinter.Call(uintptr(*hPrinter))
     if r1 == 0 {return err}

     return nil

}

func OpenPrinter(printerName string) (HANDLE, error) {
     var pPrinterName *uint16
     pPrinterName, err := syscall.UTF16PtrFromString(printerName)
     if err != nil {
          return 0, err
     }

     var hPrinter HANDLE
     r1, _, err := openPrinter.Call(uintptr(unsafe.Pointer(pPrinterName)), uintptr(unsafe.Pointer(&hPrinter)), 0)
     if r1 == 0 {
          return 0, err
     }
     return hPrinter, nil
}

func (hPrinter *HANDLE) ClosePrinter() error {
     r1, _, err := closePrinter.Call(uintptr(*hPrinter))
     if r1 == 0 {
          return err
     }
     *hPrinter = 0
     return nil
}

func (hPrinter HANDLE) GetPrinter2() (*PRINTER_INFO_2, error) {

     var needed uint32  
     var buf []byte = make([]byte, 1)
     var blen uintptr = uintptr(len(buf))

     var printerInfo *PRINTER_INFO_2

     r1, _, err := getPrinter.Call(uintptr(hPrinter), 2, uintptr(unsafe.Pointer(&buf[0])), blen, uintptr(unsafe.Pointer(&needed)))
     if r1 == 0 {          
          var newBuf []byte = make([]byte, int(needed))
          var newLen uintptr = uintptr(len(newBuf))
          r1, _, err = getPrinter.Call(uintptr(hPrinter), 2, uintptr(unsafe.Pointer(&newBuf[0])), newLen, uintptr(unsafe.Pointer(&needed)))
          if r1 == 0{
               fmt.Println("Failed")
               return nil,err
          }

          printerInfo = (*PRINTER_INFO_2)(unsafe.Pointer(&newBuf[0]))

          fmt.Println("Get Printer Info 2 Duplex Setting: ", printerInfo.pDevMode.dmDuplex)

     }
     
     return printerInfo , nil
}

func (hPrinter HANDLE) SetDuplexPrinter2(printerInfo *PRINTER_INFO_2) {

     printerInfo.pDevMode.SetDuplex(2)

     fmt.Println("Set info to duplex...")

     return
}

func (hPrinter HANDLE) SetPrinter(printerInfo *PRINTER_INFO_2) (error){

     // var bin_buf bytes.Buffer
     // binary.Write(&bin_buf, binary.BigEndian, printerInfo)
     // bs := bin_buf.Bytes()

     bs := (*[unsafe.Sizeof(printerInfo)]byte)(unsafe.Pointer(&printerInfo))     

     r1, _, err := setPrinter.Call(uintptr(hPrinter), 2, uintptr(unsafe.Pointer(&bs[0])), 0)
     if r1 != 0 {return err}

     fmt.Println("Set printer to duplex with the info...")

     return nil
}

func (hPrinter HANDLE) GetPrinter9() (*DevMode, error) {

     var needed uint32  
     var buf []byte = make([]byte, 1)
     var blen uintptr = uintptr(len(buf))

     var printerInfo *PRINTER_INFO_9

     r1, _, err := getPrinter.Call(uintptr(hPrinter), 9, uintptr(unsafe.Pointer(&buf[0])), blen, uintptr(unsafe.Pointer(&needed)))
     if r1 == 0 {          
          var newBuf []byte = make([]byte, int(needed))
          var newLen uintptr = uintptr(len(newBuf))
          r1, _, err = getPrinter.Call(uintptr(hPrinter), 9, uintptr(unsafe.Pointer(&newBuf[0])), newLen, uintptr(unsafe.Pointer(&needed)))
          if r1 == 0{
               fmt.Println("Failed")
               return nil,err
          }
          
          printerInfo = (*PRINTER_INFO_9)(unsafe.Pointer(&newBuf[0]))

     }
     
     return printerInfo.dev, nil
}

func (hPrinter HANDLE) DocumentPropertiesGet(deviceName string) (*DevMode, error) {
     pDeviceName, err := syscall.UTF16PtrFromString(deviceName)
     if err != nil {
          return nil, err
     }

     r1, _, err := documentProperties.Call(0, uintptr(hPrinter), uintptr(unsafe.Pointer(pDeviceName)), 0, 0, 0)
     cbBuf := int32(r1)
     if cbBuf < 0 {
          return nil, err
     }

     var pDevMode []byte = make([]byte, cbBuf)
     devMode := (*DevMode)(unsafe.Pointer(&pDevMode[0]))
     devMode.dmSize = uint16(cbBuf)
     devMode.dmSpecVersion = DM_SPECVERSION

     r1, _, err = documentProperties.Call(0, uintptr(hPrinter), uintptr(unsafe.Pointer(pDeviceName)), uintptr(unsafe.Pointer(devMode)), uintptr(unsafe.Pointer(devMode)), uintptr(DM_COPY))
     if int32(r1) < 0 {
          return nil, err
     }

     fmt.Println("From get:", devMode.dmDuplex)

     return devMode, nil
}

func (hPrinter HANDLE) DocumentPropertiesSet(deviceName string, devMode *DevMode) error {

     pDeviceName, err := syscall.UTF16PtrFromString(deviceName)
     if err != nil {
          return err
     }

     r1, _, err := documentProperties.Call(0, uintptr(hPrinter), uintptr(unsafe.Pointer(pDeviceName)), uintptr(unsafe.Pointer(devMode)), uintptr(unsafe.Pointer(devMode)), uintptr(DM_MODIFY))
     if int32(r1) < 0 {
          return err
     }

     return nil
}

const (
     CCHDEVICENAME = 32
     CCHFORMNAME   = 32

     DM_SPECVERSION uint16 = 0x0401
     DM_COPY        uint32 = 2
     DM_MODIFY      uint32 = 8

     DM_ORIENTATION        = 0x00000001
     DM_PAPERSIZE          = 0x00000002
     DM_PAPERLENGTH        = 0x00000004
     DM_PAPERWIDTH         = 0x00000008
     DM_SCALE              = 0x00000010
     DM_POSITION           = 0x00000020
     DM_NUP                = 0x00000040
     DM_DISPLAYORIENTATION = 0x00000080
     DM_COPIES             = 0x00000100
     DM_DEFAULTSOURCE      = 0x00000200
     DM_PRINTQUALITY       = 0x00000400
     DM_COLOR              = 0x00000800
     DM_DUPLEX             = 0x00001000
     DM_YRESOLUTION        = 0x00002000
     DM_TTOPTION           = 0x00004000
     DM_COLLATE            = 0x00008000
     DM_FORMNAME           = 0x00010000
     DM_LOGPIXELS          = 0x00020000
     DM_BITSPERPEL         = 0x00040000
     DM_PELSWIDTH          = 0x00080000
     DM_PELSHEIGHT         = 0x00100000
     DM_DISPLAYFLAGS       = 0x00200000
     DM_DISPLAYFREQUENCY   = 0x00400000
     DM_ICMMETHOD          = 0x00800000
     DM_ICMINTENT          = 0x01000000
     DM_MEDIATYPE          = 0x02000000
     DM_DITHERTYPE         = 0x04000000
     DM_PANNINGWIDTH       = 0x08000000
     DM_PANNINGHEIGHT      = 0x10000000
     DM_DISPLAYFIXEDOUTPUT = 0x20000000

     DMORIENT_PORTRAIT  int16 = 1
     DMORIENT_LANDSCAPE int16 = 2

     DMCOLOR_MONOCHROME int16 = 1
     DMCOLOR_COLOR      int16 = 2

     DMDUP_SIMPLEX    int16 = 1
     DMDUP_VERTICAL   int16 = 2
     DMDUP_HORIZONTAL int16 = 3

     DMCOLLATE_FALSE int16 = 0
     DMCOLLATE_TRUE  int16 = 1

     DMNUP_SYSTEM uint32 = 1
     DMNUP_ONEUP  uint32 = 2
)

// DEVMODE struct.
type DevMode struct {
     // WCHAR dmDeviceName[CCHDEVICENAME]
     dmDeviceName, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _ uint16

     dmSpecVersion   uint16
     dmDriverVersion uint16
     dmSize          uint16
     dmDriverExtra   uint16
     dmFields        uint32

     dmOrientation   int16
     dmPaperSize     int16
     dmPaperLength   int16
     dmPaperWidth    int16
     dmScale         int16
     dmCopies        int16
     dmDefaultSource int16
     dmPrintQuality  int16
     dmColor         int16
     dmDuplex        int16
     dmYResolution   int16
     dmTTOption      int16
     dmCollate       int16
     // WCHAR dmFormName[CCHFORMNAME]
     dmFormName, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _ uint16

     dmLogPixels        int16
     dmBitsPerPel       uint16
     dmPelsWidth        uint16
     dmPelsHeight       uint16
     dmNup              uint32
     dmDisplayFrequency uint32
     dmICMMethod        uint32
     dmICMIntent        uint32
     dmMediaType        uint32
     dmDitherType       uint32
     dmReserved1        uint32
     dmReserved2        uint32
     dmPanningWidth     uint32
     dmPanningHeight    uint32
}

func (dm *DevMode) String() string {
     s := []string{
          fmt.Sprintf("device name: %s", dm.GetDeviceName()),
          fmt.Sprintf("spec version: %d", dm.dmSpecVersion),
     }
     if dm.dmFields&DM_ORIENTATION != 0 {
          s = append(s, fmt.Sprintf("orientation: %d", dm.dmOrientation))
     }
     if dm.dmFields&DM_PAPERSIZE != 0 {
          s = append(s, fmt.Sprintf("paper size: %d", dm.dmPaperSize))
     }
     if dm.dmFields&DM_PAPERLENGTH != 0 {
          s = append(s, fmt.Sprintf("paper length: %d", dm.dmPaperLength))
     }
     if dm.dmFields&DM_PAPERWIDTH != 0 {
          s = append(s, fmt.Sprintf("paper width: %d", dm.dmPaperWidth))
     }
     if dm.dmFields&DM_SCALE != 0 {
          s = append(s, fmt.Sprintf("scale: %d", dm.dmScale))
     }
     if dm.dmFields&DM_COPIES != 0 {
          s = append(s, fmt.Sprintf("copies: %d", dm.dmCopies))
     }
     if dm.dmFields&DM_DEFAULTSOURCE != 0 {
          s = append(s, fmt.Sprintf("default source: %d", dm.dmDefaultSource))
     }
     if dm.dmFields&DM_PRINTQUALITY != 0 {
          s = append(s, fmt.Sprintf("print quality: %d", dm.dmPrintQuality))
     }
     if dm.dmFields&DM_COLOR != 0 {
          s = append(s, fmt.Sprintf("color: %d", dm.dmColor))
     }
     if dm.dmFields&DM_DUPLEX != 0 {
          s = append(s, fmt.Sprintf("duplex: %d", dm.dmDuplex))
     }
     if dm.dmFields&DM_YRESOLUTION != 0 {
          s = append(s, fmt.Sprintf("y-resolution: %d", dm.dmYResolution))
     }
     if dm.dmFields&DM_TTOPTION != 0 {
          s = append(s, fmt.Sprintf("TT option: %d", dm.dmTTOption))
     }
     if dm.dmFields&DM_COLLATE != 0 {
          s = append(s, fmt.Sprintf("collate: %d", dm.dmCollate))
     }
     if dm.dmFields&DM_FORMNAME != 0 {
          s = append(s, fmt.Sprintf("formname: %s", utf16PtrToString(&dm.dmFormName)))
     }
     if dm.dmFields&DM_LOGPIXELS != 0 {
          s = append(s, fmt.Sprintf("log pixels: %d", dm.dmLogPixels))
     }
     if dm.dmFields&DM_BITSPERPEL != 0 {
          s = append(s, fmt.Sprintf("bits per pel: %d", dm.dmBitsPerPel))
     }
     if dm.dmFields&DM_PELSWIDTH != 0 {
          s = append(s, fmt.Sprintf("pels width: %d", dm.dmPelsWidth))
     }
     if dm.dmFields&DM_PELSHEIGHT != 0 {
          s = append(s, fmt.Sprintf("pels height: %d", dm.dmPelsHeight))
     }
     if dm.dmFields&DM_NUP != 0 {
          s = append(s, fmt.Sprintf("display flags: %d", dm.dmNup))
     }
     if dm.dmFields&DM_DISPLAYFREQUENCY != 0 {
          s = append(s, fmt.Sprintf("display frequency: %d", dm.dmDisplayFrequency))
     }
     if dm.dmFields&DM_ICMMETHOD != 0 {
          s = append(s, fmt.Sprintf("ICM method: %d", dm.dmICMMethod))
     }
     if dm.dmFields&DM_ICMINTENT != 0 {
          s = append(s, fmt.Sprintf("ICM intent: %d", dm.dmICMIntent))
     }
     if dm.dmFields&DM_DITHERTYPE != 0 {
          s = append(s, fmt.Sprintf("dither type: %d", dm.dmDitherType))
     }
     if dm.dmFields&DM_PANNINGWIDTH != 0 {
          s = append(s, fmt.Sprintf("panning width: %d", dm.dmPanningWidth))
     }
     if dm.dmFields&DM_PANNINGHEIGHT != 0 {
          s = append(s, fmt.Sprintf("panning height: %d", dm.dmPanningHeight))
     }
     return strings.Join(s, ", ")
}

func (dm *DevMode) GetDeviceName() string {
     return utf16PtrToStringSize(&dm.dmDeviceName, CCHDEVICENAME*2)
}

func (dm *DevMode) GetOrientation() (int16, bool) {
     return dm.dmOrientation, dm.dmFields&DM_ORIENTATION != 0
}

func (dm *DevMode) SetOrientation(orientation int16) {
     dm.dmOrientation = orientation
     dm.dmFields |= DM_ORIENTATION
}

func (dm *DevMode) GetPaperSize() (int16, bool) {
     return dm.dmPaperSize, dm.dmFields&DM_PAPERSIZE != 0
}

func (dm *DevMode) SetPaperSize(paperSize int16) {
     dm.dmPaperSize = paperSize
     dm.dmFields |= DM_PAPERSIZE
}

func (dm *DevMode) ClearPaperSize() {
     dm.dmFields &^= DM_PAPERSIZE
}

func (dm *DevMode) GetPaperLength() (int16, bool) {
     return dm.dmPaperLength, dm.dmFields&DM_PAPERLENGTH != 0
}

func (dm *DevMode) SetPaperLength(length int16) {
     dm.dmPaperLength = length
     dm.dmFields |= DM_PAPERLENGTH
}

func (dm *DevMode) ClearPaperLength() {
     dm.dmFields &^= DM_PAPERLENGTH
}

func (dm *DevMode) GetPaperWidth() (int16, bool) {
     return dm.dmPaperWidth, dm.dmFields&DM_PAPERWIDTH != 0
}

func (dm *DevMode) SetPaperWidth(width int16) {
     dm.dmPaperWidth = width
     dm.dmFields |= DM_PAPERWIDTH
}

func (dm *DevMode) ClearPaperWidth() {
     dm.dmFields &^= DM_PAPERWIDTH
}

func (dm *DevMode) GetCopies() (int16, bool) {
     return dm.dmCopies, dm.dmFields&DM_COPIES != 0
}

func (dm *DevMode) SetCopies(copies int16) {
     dm.dmCopies = copies
     dm.dmFields |= DM_COPIES
}

func (dm *DevMode) GetColor() (int16, bool) {
     return dm.dmColor, dm.dmFields&DM_COLOR != 0
}

func (dm *DevMode) SetColor(color int16) {
     dm.dmColor = color
     dm.dmFields |= DM_COLOR
}

func (dm *DevMode) GetDuplex() (int16, bool) {
     return dm.dmDuplex, dm.dmFields&DM_DUPLEX != 0
}

func (dm *DevMode) SetDuplex(duplex int16) {
     dm.dmDuplex = duplex
     dm.dmFields |= DM_DUPLEX
}

func (dm *DevMode) GetCollate() (int16, bool) {
     return dm.dmCollate, dm.dmFields&DM_COLLATE != 0
}

func (dm *DevMode) SetCollate(collate int16) {
     dm.dmCollate = collate
     dm.dmFields |= DM_COLLATE
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//HDC CODE:
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// type HDC uintptr

// func CreateDC(deviceName string, devMode *DevMode) (HDC, error) {
//      lpszDevice, err := syscall.UTF16PtrFromString(deviceName)
//      if err != nil {
//           return 0, err
//      }
//      r1, _, err := createDCProc.Call(0, uintptr(unsafe.Pointer(lpszDevice)), 0, uintptr(unsafe.Pointer(devMode)))
//      if r1 == 0 {
//           return 0, err
//      }
//      return HDC(r1), nil
// }

// func (hDC HDC) ResetDC(devMode *DevMode) error {
//      r1, _, err := resetDCProc.Call(uintptr(hDC), uintptr(unsafe.Pointer(devMode)))
//      if r1 == 0 {
//           return err
//      }
//      return nil
// }

// func (hDC *HDC) DeleteDC() error {
//      r1, _, err := deleteDCProc.Call(uintptr(*hDC))
//      if r1 == 0 {
//           return err
//      }
//      *hDC = 0
//      return nil
// }

// func (hDC HDC) GetDeviceCaps(nIndex int32) int32 {
//      // No error returned. r1 == 0 when nIndex == -1.
//      r1, _, _ := getDeviceCapsProc.Call(uintptr(hDC), uintptr(nIndex))
//      return int32(r1)
// }

// func (hDC HDC) StartDoc(docName string) (int32, error) {
//      var docInfo DocInfo
//      var err error
//      docInfo.cbSize = int32(unsafe.Sizeof(docInfo))
//      docInfo.lpszDocName, err = syscall.UTF16PtrFromString(docName)
//      if err != nil {
//           return 0, err
//      }

//      r1, _, err := startDocProc.Call(uintptr(hDC), uintptr(unsafe.Pointer(&docInfo)))
//      if r1 <= 0 {
//           return 0, err
//      }
//      return int32(r1), nil
// }





/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//PREVIOUS VERSION CODE:
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

func GoPrintFromOpenFile(printerHandle uintptr, docName string, fileContents []byte) error {
     var err error

     startPrinter(printerHandle, docName)
     startPagePrinter.Call(printerHandle)
     err = writePrinterFuncFromFileContents(printerHandle, docName, fileContents)
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
     var contentLen = uintptr(uint32(len(fileContents)))
     var writtenLen int
     _, _, err = writePrinter.Call(printerHandle, uintptr(unsafe.Pointer(&fileContents[0])),  contentLen, uintptr(unsafe.Pointer(&writtenLen)))
     fmt.Println("Writing to printer:", err)

     return nil
}


func writePrinterFuncFromFileContents(printerHandle uintptr, docName string, fileContents []byte ) error {
     fmt.Println("About to write file to path: ", docName)
     var contentLen = uintptr(uint32(len(fileContents)))
     var writtenLen int
     _, _, err := writePrinter.Call(printerHandle, uintptr(unsafe.Pointer(&fileContents[0])),  contentLen, uintptr(unsafe.Pointer(&writtenLen)))
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


///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//UTF funcs:
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const utf16StringMaxBytes = 1024

func utf16PtrToStringSize(s *uint16, bytes uint32) string {
     if s == nil {
          return ""
     }

     hdr := reflect.SliceHeader{
          Data: uintptr(unsafe.Pointer(s)),
          Len:  int(bytes / 2),
          Cap:  int(bytes / 2),
     }
     c := *(*[]uint16)(unsafe.Pointer(&hdr))

     return syscall.UTF16ToString(c)
}

func utf16PtrToString(s *uint16) string {
     return utf16PtrToStringSize(s, utf16StringMaxBytes)
}
