package goprint

import(          
     "reflect"
     "syscall"
     "unsafe"
     "io/ioutil"
     "fmt"
     "strings"
     "errors"
     "os/exec"
)

const (
     uint16Size = 2
     int32Size  = 4
)

var(
     dll =                    syscall.MustLoadDLL("winspool.drv")
     getDefaultPrinter =      dll.MustFindProc("GetDefaultPrinterW")
     openPrinter =            dll.MustFindProc("OpenPrinterW")
     startDocPrinter =        dll.MustFindProc("StartDocPrinterW")
     startPagePrinter =       dll.MustFindProc("StartPagePrinter")
     writePrinter =           dll.MustFindProc("WritePrinter")
     endPagePrinter =         dll.MustFindProc("EndPagePrinter")
     endDocPrinter =          dll.MustFindProc("EndDocPrinter")
     closePrinter=            dll.MustFindProc("ClosePrinter")    
     procEnumPrintersW =      dll.MustFindProc("EnumPrintersW") 
     documentProperties =     dll.MustFindProc("DocumentPropertiesW")
     getPrinter =             dll.MustFindProc("GetPrinterW")
     setPrinter =             dll.MustFindProc("SetPrinterW")

// procIsValidDevmodeW =    dll.MustFindProc("IsValidDevmodeW")

//sys IsvalidDevMode(dev *DevMode, buf uint16) (b bool) =         winspool.IsValidDevmodeW
//sys SetDefaultPrinter(printerName string) (b bool) = winspool.SetDefaultPrinterW 
//sys OpenPrinter2(name *uint16, h *syscall.Handle, defaults uintptr) (err error) = winspool.OpenPrinter2W

     deviceCapabilitiesProc         = dll.MustFindProc("DeviceCapabilitiesW")

     gdi32    = syscall.MustLoadDLL("gdi32.dll")
     abortDocProc                   = gdi32.MustFindProc("AbortDoc")  
     createDCProc                   = gdi32.MustFindProc("CreateDCW")
     deleteDCProc                   = gdi32.MustFindProc("DeleteDC")  
     endDocProc                     = gdi32.MustFindProc("EndDoc")
     endPageProc                    = gdi32.MustFindProc("EndPage")   
     getDeviceCapsProc              = gdi32.MustFindProc("GetDeviceCaps")  
     resetDCProc                    = gdi32.MustFindProc("ResetDCW")
     setGraphicsModeProc            = gdi32.MustFindProc("SetGraphicsMode")
     setWorldTransformProc          = gdi32.MustFindProc("SetWorldTransform")
     startDocProc                   = gdi32.MustFindProc("StartDocW")
     startPageProc                  = gdi32.MustFindProc("StartPage")
)

type PRINTPROCESSOROPENDATA struct {
     pDevMode       *DevMode
     pDatatype      *uint16
     pParameters    *uint16
     pDocumentName  *uint16
     JobID           uint32
     pOutputFile    *uint16
     pPrinterName   *uint16
}

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

func (pi *PRINTER_INFO_5) GetPrinterName() string {
     return utf16PtrToString(pi.PrinterName)
}

func (pi *PRINTER_INFO_5) GetPortName() string {
     return utf16PtrToString(pi.PortName)
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


type PRINTER_DEFAULTS struct {
     pDatatype      *uint16
     pDevMode       *DevMode
     DesiredAccess  uint32
}

//C++
//pDatatype    LPTSTR
//pDevMode     LPDEVMODE
//DesiredAccess     ACCESS_MASK

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

//Prints Processor data to console.
func (pi *PRINTER_INFO_2) ShowPrintProcessor() {

     fmt.Println(utf16PtrToString(pi.pPrintProcessor))
     fmt.Println(utf16PtrToString(pi.pDatatype))

     return
}

func (pi *PRINTER_INFO_2) DevModeIsValid() bool {
     return IsvalidDevMode(pi.GetDevMode(), pi.GetDevMode().GetSize())
}

//PRINTER_INFO_2: 

func enumPrintersLevel(level uint32) ([]byte, uint32, error) {
     var cbBuf, pcReturned uint32
     _, _, err := procEnumPrintersW.Call(PRINTER_ENUM_LOCAL, 0, uintptr(level), 0, 0, uintptr(unsafe.Pointer(&cbBuf)), uintptr(unsafe.Pointer(&pcReturned)))
     if err != ERROR_INSUFFICIENT_BUFFER {
          return nil, 0, err
     }

     var pPrinterEnum []byte = make([]byte, cbBuf)
     r1, _, err := procEnumPrintersW.Call(PRINTER_ENUM_LOCAL, 0, uintptr(level), uintptr(unsafe.Pointer(&pPrinterEnum[0])), uintptr(cbBuf), uintptr(unsafe.Pointer(&cbBuf)), uintptr(unsafe.Pointer(&pcReturned)))
     if r1 == 0 {
          return nil, 0, err
     }

     return pPrinterEnum, pcReturned, nil
}

//Get PRINTER_INFO_2 for all available printers 
func EnumPrinters2() ([]PRINTER_INFO_2, error) {
     pPrinterEnum, pcReturned, err := enumPrintersLevel(2)
     if err != nil {
          return nil, err
     }

     hdr := reflect.SliceHeader{
          Data: uintptr(unsafe.Pointer(&pPrinterEnum[0])),
          Len:  int(pcReturned),
          Cap:  int(pcReturned),
     }
     printers := *(*[]PRINTER_INFO_2)(unsafe.Pointer(&hdr))
     return printers, nil
}

//SHOULD get the data type RAW or XPS_PASS. Needs to be tested for XPS_PASS.
func (pi *PRINTER_INFO_2) GetDataTypeString() string {
     return utf16PtrToString(pi.pDatatype)
}

func (pi *PRINTER_INFO_2) GetPrinterName() string {
     return utf16PtrToString(pi.pPrinterName)
}

func (pi *PRINTER_INFO_2) GetServerName() string {
     return utf16PtrToString(pi.pServerName)
}

func (pi *PRINTER_INFO_2) GetPortName() string {
     return utf16PtrToString(pi.pPortName)
}

func (pi *PRINTER_INFO_2) GetDriverName() string {
     return utf16PtrToString(pi.pDriverName)
}

func (pi *PRINTER_INFO_2) GetLocation() string {
     return utf16PtrToString(pi.pLocation)
}

func (pi *PRINTER_INFO_2) GetDevMode() *DevMode {
     return pi.pDevMode
}


func (pi *PRINTER_INFO_2) GetAttributes() uint32 {
     return pi.attributes
}

func (pi *PRINTER_INFO_2) GetStatus() uint32 {
     return pi.status
}

func (pi *PRINTER_INFO_2) PrintDevMode() {

     fmt.Println(pi.pDevMode.String())

     return
}



// PRINTER_ENUM_VALUES struct.
type PrinterEnumValues struct {
     pValueName  *uint16
     cbValueName uint32
     dwType      uint32
     pData       uintptr
     cbData      uint32
}

// First parameter to EnumPrinters().
const (
     PRINTER_ENUM_DEFAULT     = 0x00000001
     PRINTER_ENUM_LOCAL       = 0x00000002
     PRINTER_ENUM_CONNECTIONS = 0x00000004
     PRINTER_ENUM_FAVORITE    = 0x00000004
     PRINTER_ENUM_NAME        = 0x00000008
     PRINTER_ENUM_REMOTE      = 0x00000010
     PRINTER_ENUM_SHARED      = 0x00000020
     PRINTER_ENUM_NETWORK     = 0x00000040
     PRINTER_ENUM_EXPAND      = 0x00004000
     PRINTER_ENUM_CONTAINER   = 0x00008000
     PRINTER_ENUM_ICONMASK    = 0x00ff0000
     PRINTER_ENUM_ICON1       = 0x00010000
     PRINTER_ENUM_ICON2       = 0x00020000
     PRINTER_ENUM_ICON3       = 0x00040000
     PRINTER_ENUM_ICON4       = 0x00080000
     PRINTER_ENUM_ICON5       = 0x00100000
     PRINTER_ENUM_ICON6       = 0x00200000
     PRINTER_ENUM_ICON7       = 0x00400000
     PRINTER_ENUM_ICON8       = 0x00800000
     PRINTER_ENUM_HIDE        = 0x01000000
)

// Errors returned by GetLastError().
const (
     NO_ERROR                  = syscall.Errno(0)
     ERROR_INVALID_PARAMETER   = syscall.Errno(87)
     ERROR_INSUFFICIENT_BUFFER = syscall.Errno(122)
)

//Prints a simple file using the lpr windows command. In many cases, the pServerName should actually be the port from the PRINTER_INFO_2 struct.
func PrintLPR(path string, printerName string, pServerName string) error {
     
     _, err := exec.Command("lpr", fmt.Sprintf(`-S %s -P "%s" %s`, pServerName, printerName, path)).Output()
     if err != nil { return fmt.Errorf("Failed to print : %v\n", err) }

     return nil
}

type HANDLE uintptr

//mainprinterfunc

//FROM WINDOWS DOCS:
// "When a high-level document (such as an Adobe PDF or Microsoft Word file) or other printer data (such PCL, PS, or HPGL) is sent directly to a printer, 
// the print settings defined in the document take precedent over Windows print settings. Documents output when the value of the pDatatype member of the DOC_INFO_1 structure 
// that was passed in the pDocInfo parameter of the StartDocPrinter call is "RAW" must fully describe the DEVMODE-style print job settings in the language understood by the hardware."
//
//Also from WINDOWS docs:
// "The typical sequence for a print job is as follows:

// To begin a print job, call StartDocPrinter.
// To begin each page, call StartPagePrinter.
// To write data to a page, call WritePrinter.
// To end each page, call EndPagePrinter.
// Repeat 2, 3, and 4 for as many pages as necessary.
// To end the print job, call EndDocPrinter.
// Note that calling StartPagePrinter and EndPagePrinter may not be necessary, such as if the print data type includes the page information."
//
//startPagePrinter is called called once before the writePrinter call, just in case. endPagePrinter is also called only once after the writePrinter. 
//Therefore, this method doesn't print page by page and assumes the printer can handle the data you are sending it. This is sending raw bytes to the printer. Results may vary...
//
//The pDatatype in the DOC_INFO_1 stuct should be either set to "RAW" or XPS_PASS". This is obtained from the PRINTER_INFO_2 struct and passed in directly. To check which data type
//your printer is using, use the ShowPrintProcessor() method to print this infomation to the console. 
func (hPrinter *HANDLE) Print(path string) error {
     pathArray := strings.Split(path, "/")
     l := len(pathArray)
     name := pathArray[l-1]
     
     ptr2, err := hPrinter.GetPrinter2()
     if err != nil {return err}

     dType := ptr2.GetDataTypeString()
     d := DOC_INFO_1{
          pDocName:      &(syscall.StringToUTF16(name))[0],
          pOutputFile:   nil,
          pDatatype:     &(syscall.StringToUTF16(dType))[0],
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

//Obtain a printer handle to print with.
func OpenPrinter(printerName string) (HANDLE, error) {
     var pPrinterName *uint16
     pPrinterName, err := syscall.UTF16PtrFromString(printerName)
     if err != nil {
          return 0, err
     }

     //NEED TO IMPLEMENT THIS....
     //PRINTER_DEFAULTS struct:
     //pDatatype    LPTSTR
     //pDevMode     LPDEVMODE
     //DesiredAccess     ACCESS_MASK

     //PRINTER_ACCESS_USE
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

//Get the PRITNER_INFO_2 struct for the printer. This will allows you to make global system admin changes to a printers settings. 
//If you want to make local user changes, it is advised you use GetPrinter9(). 
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
     }
     
     return printerInfo , nil
}

//This might work if you have admin access to the printer. However, this will change the setting globally for all users. Not yet tested. It is recommended you use SetDuplexPrinter9() instead.
//The '2' in the name means that it uses the PRINTER_INFO_2 struct to modify the setting.
func (hPrinter HANDLE) SetDuplexPrinter2(i int16) error {

     ptr2, err := hPrinter.GetPrinter2()
     if err != nil {return err}

     ptr2.pDevMode.SetDuplex(i)
     
     err = hPrinter.SetPrinter2(ptr2) 
     if err != nil {return err}     

     return nil
}

//This might work if you have access to the printer. However, this will change the setting globally for all users. Not yet tested.  It is recommended you use SetDuplexPrinter9() instead.
//The '2' in the name means that it uses the PRINTER_INFO_2 struct to modify the setting.
func (hPrinter HANDLE) SetPrinter2(printerInfo *PRINTER_INFO_2) (error){

     // var bin_buf bytes.Buffer
     // binary.Write(&bin_buf, binary.BigEndian, printerInfo)
     // bs := bin_buf.Bytes()

     bs := (*[unsafe.Sizeof(printerInfo)]byte)(unsafe.Pointer(&printerInfo))     

     //hprinter, level, ppPrinter, cmd

     r1, _, err := setPrinter.Call(uintptr(hPrinter), 2, uintptr(unsafe.Pointer(&bs[0])), 0)
     if r1 == 0 {return err}

     return nil
}

//THIS SETS THE USER PRINT SETTINGS.
//With this, we can only set the DevMode data. Therefore, we cannot set stapling as because we can't access the driver specifics.
//To use the Printer_info_2, we need to set the security descriptor. That will probably work on a local network where we have admin rights. 
//Therefore, using PRINTER_INFO_9 makes this a bit easier as it only affects the user settings, not the global settings for the printer. 
func (hPrinter HANDLE) SetPrinter9(printerInfo *PRINTER_INFO_9) (error){
     
     bs := (*[unsafe.Sizeof(printerInfo.pDevMode)]byte)(unsafe.Pointer(&printerInfo.pDevMode))     

     //IN
     //hprinter, level, ppPrinter, cmd

     //OUT
     //0
     //IF THE FUNCTION SUCCEEDS, THE RETURN VALUE IS A NONZERO VALUE. IF THE FUNCTION FAILS, THE RETURN VALUE IS ZERO....
     r1, _, err := setPrinter.Call(uintptr(hPrinter), 9, uintptr(unsafe.Pointer(&bs[0])), 0)
     if r1 == 0 {return err}

     return nil
}

//Set the printer to simplex, duplex based with the value 1,2 or 3 as represented by DMDUP_SIMPLEX, DMDUP_VERTICAL or DMDUP_HORIZONTAL.
//This is set using the PRINTER_INFO_9 which will only affect the printer settings for this user. 
func (hPrinter HANDLE) SetDuplexPrinter9(i int16) error {

     ptr9, err := hPrinter.GetPrinter9()
     if err != nil {return err}

     ptr9.pDevMode.SetDuplex(i)

     err = hPrinter.SetPrinter9(ptr9) 
     if err != nil {return err}     

     return nil
}

//PRINTER_INFO_9 struct. 
type PRINTER_INFO_9 struct {
     pDevMode *DevMode
}

func (pi *PRINTER_INFO_9) GetDevMode() *DevMode {
     return pi.pDevMode
}

func (pi *PRINTER_INFO_9) DevModeIsValid() bool {
     return IsvalidDevMode(pi.GetDevMode(), pi.GetDevMode().GetSize())
}

//Will print DevMode struct data to the console. Mainly useful for checking if a SetPrinter9() call actually took effect
func (pi *PRINTER_INFO_9) PrintDevMode() {

     //Print the private data size:
     fmt.Println("Private data chunk size:", pi.pDevMode.dmDriverExtra)

     //At some point, we need to implement making a pointer to the driver based on pi.dmSize.

     fmt.Println(pi.pDevMode.String())

     return
}

//Gets the PRINTER_INFO_9 struct used for making changes to user printer preferences. 
func (hPrinter HANDLE) GetPrinter9() (*PRINTER_INFO_9, error) {

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
     
     return printerInfo, nil
}

//Gets the PRINTER_INFO_9 struct used for getting basic data about printer: 
func (hPrinter HANDLE) GetPrinter5() (*PRINTER_INFO_5, error) {

     var needed uint32  
     var buf []byte = make([]byte, 1)
     var blen uintptr = uintptr(len(buf))

     var printerInfo *PRINTER_INFO_5

     r1, _, err := getPrinter.Call(uintptr(hPrinter), 5, uintptr(unsafe.Pointer(&buf[0])), blen, uintptr(unsafe.Pointer(&needed)))
     if r1 == 0 { 
          fmt.Println("Needed: ", int(needed))         
          var newBuf []byte = make([]byte, int(needed))
          var newLen uintptr = uintptr(len(newBuf))
          r1, _, err = getPrinter.Call(uintptr(hPrinter), 5, uintptr(unsafe.Pointer(&newBuf[0])), newLen, uintptr(unsafe.Pointer(&needed)))
          if r1 == 0{
               fmt.Println("Failed")
               return nil,err
          }          

          printerInfo = (*PRINTER_INFO_5)(unsafe.Pointer(&newBuf[0]))

     }
     
     return printerInfo, nil
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

     return devMode, nil
}

func (hPrinter HANDLE) DocumentPropertiesSet(deviceName string, devMode *DevMode) error {
     pDeviceName, err := syscall.UTF16PtrFromString(deviceName)
     if err != nil {
          return err
     }

     r1, _, err := documentProperties.Call(0, uintptr(hPrinter), uintptr(unsafe.Pointer(pDeviceName)), uintptr(unsafe.Pointer(devMode)), uintptr(unsafe.Pointer(devMode)), uintptr(DM_COPY|DM_MODIFY))
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

func (dm *DevMode) GetSize() uint16 {
          return dm.dmSize + dm.dmDriverExtra
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

// DevMode.dmPaperSize values.
const (
     DMPAPER_LETTER                        = 1
     DMPAPER_LETTERSMALL                   = 2
     DMPAPER_TABLOID                       = 3
     DMPAPER_LEDGER                        = 4
     DMPAPER_LEGAL                         = 5
     DMPAPER_STATEMENT                     = 6
     DMPAPER_EXECUTIVE                     = 7
     DMPAPER_A3                            = 8
     DMPAPER_A4                            = 9
     DMPAPER_A4SMALL                       = 10
     DMPAPER_A5                            = 11
     DMPAPER_B4                            = 12
     DMPAPER_B5                            = 13
     DMPAPER_FOLIO                         = 14
     DMPAPER_QUARTO                        = 15
     DMPAPER_10X14                         = 16
     DMPAPER_11X17                         = 17
     DMPAPER_NOTE                          = 18
     DMPAPER_ENV_9                         = 19
     DMPAPER_ENV_10                        = 20
     DMPAPER_ENV_11                        = 21
     DMPAPER_ENV_12                        = 22
     DMPAPER_ENV_14                        = 23
     DMPAPER_CSHEET                        = 24
     DMPAPER_DSHEET                        = 25
     DMPAPER_ESHEET                        = 26
     DMPAPER_ENV_DL                        = 27
     DMPAPER_ENV_C5                        = 28
     DMPAPER_ENV_C3                        = 29
     DMPAPER_ENV_C4                        = 30
     DMPAPER_ENV_C6                        = 31
     DMPAPER_ENV_C65                       = 32
     DMPAPER_ENV_B4                        = 33
     DMPAPER_ENV_B5                        = 34
     DMPAPER_ENV_B6                        = 35
     DMPAPER_ENV_ITALY                     = 36
     DMPAPER_ENV_MONARCH                   = 37
     DMPAPER_ENV_PERSONAL                  = 38
     DMPAPER_FANFOLD_US                    = 39
     DMPAPER_FANFOLD_STD_GERMAN            = 40
     DMPAPER_FANFOLD_LGL_GERMAN            = 41
     DMPAPER_ISO_B4                        = 42
     DMPAPER_JAPANESE_POSTCARD             = 43
     DMPAPER_9X11                          = 44
     DMPAPER_10X11                         = 45
     DMPAPER_15X11                         = 46
     DMPAPER_ENV_INVITE                    = 47
     DMPAPER_RESERVED_48                   = 48
     DMPAPER_RESERVED_49                   = 49
     DMPAPER_LETTER_EXTRA                  = 50
     DMPAPER_LEGAL_EXTRA                   = 51
     DMPAPER_TABLOID_EXTRA                 = 52
     DMPAPER_A4_EXTRA                      = 53
     DMPAPER_LETTER_TRANSVERSE             = 54
     DMPAPER_A4_TRANSVERSE                 = 55
     DMPAPER_LETTER_EXTRA_TRANSVERSE       = 56
     DMPAPER_A_PLUS                        = 57
     DMPAPER_B_PLUS                        = 58
     DMPAPER_LETTER_PLUS                   = 59
     DMPAPER_A4_PLUS                       = 60
     DMPAPER_A5_TRANSVERSE                 = 61
     DMPAPER_B5_TRANSVERSE                 = 62
     DMPAPER_A3_EXTRA                      = 63
     DMPAPER_A5_EXTRA                      = 64
     DMPAPER_B5_EXTRA                      = 65
     DMPAPER_A2                            = 66
     DMPAPER_A3_TRANSVERSE                 = 67
     DMPAPER_A3_EXTRA_TRANSVERSE           = 68
     DMPAPER_DBL_JAPANESE_POSTCARD         = 69
     DMPAPER_A6                            = 70
     DMPAPER_JENV_KAKU2                    = 71
     DMPAPER_JENV_KAKU3                    = 72
     DMPAPER_JENV_CHOU3                    = 73
     DMPAPER_JENV_CHOU4                    = 74
     DMPAPER_LETTER_ROTATED                = 75
     DMPAPER_A3_ROTATED                    = 76
     DMPAPER_A4_ROTATED                    = 77
     DMPAPER_A5_ROTATED                    = 78
     DMPAPER_B4_JIS_ROTATED                = 79
     DMPAPER_B5_JIS_ROTATED                = 80
     DMPAPER_JAPANESE_POSTCARD_ROTATED     = 81
     DMPAPER_DBL_JAPANESE_POSTCARD_ROTATED = 82
     DMPAPER_A6_ROTATED                    = 83
     DMPAPER_JENV_KAKU2_ROTATED            = 84
     DMPAPER_JENV_KAKU3_ROTATED            = 85
     DMPAPER_JENV_CHOU3_ROTATED            = 86
     DMPAPER_JENV_CHOU4_ROTATED            = 87
     DMPAPER_B6_JIS                        = 88
     DMPAPER_B6_JIS_ROTATED                = 89
     DMPAPER_12X11                         = 90
     DMPAPER_JENV_YOU4                     = 91
     DMPAPER_JENV_YOU4_ROTATED             = 92
     DMPAPER_P16K                          = 93
     DMPAPER_P32K                          = 94
     DMPAPER_P32KBIG                       = 95
     DMPAPER_PENV_1                        = 96
     DMPAPER_PENV_2                        = 97
     DMPAPER_PENV_3                        = 98
     DMPAPER_PENV_4                        = 99
     DMPAPER_PENV_5                        = 100
     DMPAPER_PENV_6                        = 101
     DMPAPER_PENV_7                        = 102
     DMPAPER_PENV_8                        = 103
     DMPAPER_PENV_9                        = 104
     DMPAPER_PENV_10                       = 105
     DMPAPER_P16K_ROTATED                  = 106
     DMPAPER_P32K_ROTATED                  = 107
     DMPAPER_P32KBIG_ROTATED               = 108
     DMPAPER_PENV_1_ROTATED                = 109
     DMPAPER_PENV_2_ROTATED                = 110
     DMPAPER_PENV_3_ROTATED                = 111
     DMPAPER_PENV_4_ROTATED                = 112
     DMPAPER_PENV_5_ROTATED                = 113
     DMPAPER_PENV_6_ROTATED                = 114
     DMPAPER_PENV_7_ROTATED                = 115
     DMPAPER_PENV_8_ROTATED                = 116
     DMPAPER_PENV_9_ROTATED                = 117
     DMPAPER_PENV_10_ROTATED               = 118
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//HDC CODE:
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//HDC is something to consider for future implementations. It has no real use in this package on it's own as there are no GDI methods implemented. 
type HDC uintptr

func CreateDC(deviceName string, devMode *DevMode) (HDC, error) {
     lpszDevice, err := syscall.UTF16PtrFromString(deviceName)
     if err != nil {
          return 0, err
     }
     r1, _, err := createDCProc.Call(0, uintptr(unsafe.Pointer(lpszDevice)), 0, uintptr(unsafe.Pointer(devMode)))
     if r1 == 0 {
          return 0, err
     }
     return HDC(r1), nil
}

func (hDC HDC) ResetDC(devMode *DevMode) error {
     r1, _, err := resetDCProc.Call(uintptr(hDC), uintptr(unsafe.Pointer(devMode)))
     if r1 == 0 {
          return err
     }
     return nil
}

func (hDC *HDC) DeleteDC() error {
     r1, _, err := deleteDCProc.Call(uintptr(*hDC))
     if r1 == 0 {
          return err
     }
     *hDC = 0
     return nil
}

func (hDC HDC) GetDeviceCaps(nIndex int32) int32 {
     // No error returned. r1 == 0 when nIndex == -1.
     r1, _, _ := getDeviceCapsProc.Call(uintptr(hDC), uintptr(nIndex))
     return int32(r1)
}

func (hDC HDC) StartDoc(docName string) (int32, error) {
     var docInfo DocInfo
     var err error
     docInfo.cbSize = int32(unsafe.Sizeof(docInfo))
     docInfo.lpszDocName, err = syscall.UTF16PtrFromString(docName)
     if err != nil {
          return 0, err
     }

     r1, _, err := startDocProc.Call(uintptr(hDC), uintptr(unsafe.Pointer(&docInfo)))
     if r1 <= 0 {
          return 0, err
     }
     return int32(r1), nil
}

//this won't work... There is no Write Doc function that utilizes the HDC...
// func (hDC HDC) WriteDoc(path string, handler HANDLE) error {
//      fileContents, err := ioutil.ReadFile(path)     
//      if err != nil {          
//           return err
//      }
//      var contentLen uintptr = uintptr(len(fileContents))
//      var writtenLen int
//      _, _, err = writePrinter.Call(uintptr(handler), uintptr(unsafe.Pointer(&fileContents[0])),  contentLen, uintptr(unsafe.Pointer(&writtenLen)))
//      fmt.Println("Write to printer: ", path, " ", err)
//      if err != nil {
//           return err
//      }
//      return nil
// }

func (hDC HDC) EndDoc() error {
     r1, _, err := endDocProc.Call(uintptr(hDC))
     if r1 <= 0 {
          return err
     }
     return nil
}

func (hDC HDC) StartPage() error {
     r1, _, err := startPageProc.Call(uintptr(hDC))
     if r1 <= 0 {
          return err
     }
     return nil
}

func (hDC HDC) EndPage() error {
     r1, _, err := endPageProc.Call(uintptr(hDC))
     if r1 <= 0 {
          return err
     }
     return nil
}

func (hDC HDC) SetGraphicsMode(iMode int32) error {
     r1, _, err := setGraphicsModeProc.Call(uintptr(hDC), uintptr(iMode))
     if r1 == 0 {
          return err
     }
     return nil
}

type XFORM struct {
     eM11 float32 // X scale.
     eM12 float32 // Always zero.
     eM21 float32 // Always zero.
     eM22 float32 // Y scale.
     eDx  float32 // X offset.
     eDy  float32 // Y offset.
}

func NewXFORM(xScale, yScale, xOffset, yOffset float32) *XFORM {
     return &XFORM{xScale, 0, 0, yScale, xOffset, yOffset}
}

func (hDC HDC) SetWorldTransform(xform *XFORM) error {
     r1, _, err := setWorldTransformProc.Call(uintptr(hDC), uintptr(unsafe.Pointer(xform)))
     if r1 == 0 {
          if err == NO_ERROR {
               return fmt.Errorf("SetWorldTransform call failed; return value %d", int32(r1))
          }
          return err
     }
     return nil
}


///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//DEVICE CAPABILITIES: 
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// IN C++:
//int DeviceCapabilitiesA(
//LPCSTR  pDevice,
//LPCSTR  pPort,
//WORD         fwCapability,
//LPSTR   pOutput,
//const DEVMODEA    *pDevMode
//);


// Device capabilities for DeviceCapabilities().
// fwCapability values:
// All notes are taken directly from the Windows page. The following is the C++ structure for reference: 
//
// IN C++:
//int DeviceCapabilitiesA(
//LPCSTR  pDevice,
//LPCSTR  pPort,
//WORD         fwCapability,
//LPSTR   pOutput,
//const DEVMODEA    *pDevMode
//);

const (     
     DC_FIELDS            = 1
     DC_PAPERS            = 2
     DC_PAPERSIZE         = 3
     DC_MINEXTENT         = 4
     DC_MAXEXTENT         = 5
     DC_BINS              = 6
     DC_DUPLEX            = 7
     DC_SIZE              = 8
     DC_EXTRA             = 9
     DC_VERSION           = 10
     DC_DRIVER            = 11     
     DC_BINNAMES          = 12
     DC_ENUMRESOLUTIONS   = 13
     DC_FILEDEPENDENCIES  = 14
     DC_TRUETYPE          = 15
     DC_PAPERNAMES        = 16
     DC_ORIENTATION       = 17
     DC_COPIES            = 18
     DC_BINADJUST         = 19
     DC_EMF_COMPLAINT     = 20
     DC_DATATYPE_PRODUCED = 21
     DC_COLLATE           = 22
     DC_MANUFACTURER      = 23
     DC_MODEL             = 24
     DC_PERSONALITY       = 25
     DC_PRINTRATE         = 26
     DC_PRINTRATEUNIT     = 27
     DC_PRINTERMEM        = 28
     DC_MEDIAREADY        = 29
     DC_STAPLE            = 30
     DC_PRINTRATEPPM      = 31
     DC_COLORDEVICE       = 32
     DC_NUP               = 33
     DC_MEDIATYPENAMES    = 34
     DC_MEDIATYPES        = 35

     PRINTRATEUNIT_PPM = 1
     PRINTRATEUNIT_CPS = 2
     PRINTRATEUNIT_LPM = 3
     PRINTRATEUNIT_IPM = 4
)

// func IsvalidDevMode(dev *DevMode) (b bool) {
//      r0, _, _ := procIsValidDevmodeW.Call(uintptr(unsafe.Pointer(dev)), uintptr(dev.GetSize()))
//      // r0, _, _ := syscall.Syscall(procIsValidDevmodeW.Addr(), 1, uintptr(unsafe.Pointer(dev)), 0, 0)
//      b = r0 != 0
//      return
// }

func deviceCapabilities(device, port string, fwCapability uint16, pOutput []byte) (int32, error) {
     pDevice, err := syscall.UTF16PtrFromString(device)
     if err != nil {
          return 0, err
     }
     pPort, err := syscall.UTF16PtrFromString(port)
     if err != nil {
          return 0, err
     }

     var r1 uintptr
     if pOutput == nil {
          r1, _, _ = deviceCapabilitiesProc.Call(uintptr(unsafe.Pointer(pDevice)), uintptr(unsafe.Pointer(pPort)), uintptr(fwCapability), 0, 0)
     } else {
          r1, _, _ = deviceCapabilitiesProc.Call(uintptr(unsafe.Pointer(pDevice)), uintptr(unsafe.Pointer(pPort)), uintptr(fwCapability), uintptr(unsafe.Pointer(&pOutput[0])), 0)
     }

     if int32(r1) == -1 {
          return 0, errors.New("DeviceCapabilities called with unsupported capability, or there was an error")
     }
     return int32(r1), nil
}

func DeviceCapabilitiesInt32(device, port string, fwCapability uint16) (int32, error) {
     pDevice, err := syscall.UTF16PtrFromString(device)
     if err != nil {
          return 0, err
     }
     pPort, err := syscall.UTF16PtrFromString(port)
     if err != nil {
          return 0, err
     }

     r1, _, _ := deviceCapabilitiesProc.Call(uintptr(unsafe.Pointer(pDevice)), uintptr(unsafe.Pointer(pPort)), uintptr(fwCapability), 0, 0)
     return int32(r1), nil
}

func DeviceCapabilitiesUint16Array(device, port string, fwCapability uint16) ([]uint16, error) {
     nValue, err := deviceCapabilities(device, port, fwCapability, nil)
     if err != nil {
          return nil, err
     }

     if nValue <= 0 {
          return []uint16{}, nil
     }

     pOutput := make([]byte, uint16Size*nValue)
     _, err = deviceCapabilities(device, port, fwCapability, pOutput)
     if err != nil {
          return nil, err
     }

     values := make([]uint16, 0, nValue)
     for i := int32(0); i < nValue; i++ {
          value := *(*uint16)(unsafe.Pointer(&pOutput[i*uint16Size]))
          values = append(values, value)
     }

     return values, nil
}

// DeviceCapabilitiesInt32Pairs returns a slice of an even quantity of int32.
func DeviceCapabilitiesInt32Pairs(device, port string, fwCapability uint16) ([]int32, error) {
     nValue, err := deviceCapabilities(device, port, fwCapability, nil)
     if err != nil {
          return nil, err
     }

     if nValue <= 0 {
          return []int32{}, nil
     }

     pOutput := make([]byte, int32Size*2*nValue)
     _, err = deviceCapabilities(device, port, fwCapability, pOutput)
     if err != nil {
          return nil, err
     }

     values := make([]int32, 0, nValue*2)
     for i := int32(0); i < nValue*2; i++ {
          value := *(*int32)(unsafe.Pointer(&pOutput[i*int32Size]))
          values = append(values, value)
     }

     return values, nil
}

//Gets the device capability as a slice of string. Refer to deviceCapabilities.txt for which DeviceCapability function you should use.
func DeviceCapabilitiesStrings(device, port string, fwCapability uint16, stringLength int32) ([]string, error) {
     nString, err := deviceCapabilities(device, port, fwCapability, nil)
     if err != nil {
          return nil, err
     }

     if nString <= 0 {
          return []string{}, nil
     }

     pOutput := make([]byte, stringLength*uint16Size*nString)
     _, err = deviceCapabilities(device, port, fwCapability, pOutput)
     if err != nil {
          return nil, err
     }

     values := make([]string, 0, nString)
     for i := int32(0); i < nString; i++ {
          value := utf16PtrToString((*uint16)(unsafe.Pointer(&pOutput[i*stringLength])))
          values = append(values, value)
     }

     return values, nil
}



/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//PREVIOUS VERSION CODE:
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//Version 1 code:
//
//Opens a printer which can then be used to send documents to. Must be closed by user once.
func GoOpenPrinter(printerName string) (uintptr, error) {

     // printerName, printerName16 := getDefaultPrinterName();     
     
     printerName16 := syscall.StringToUTF16(printerName)     
     printerHandle, err := openPrinterFunc(printerName, printerName16)      
     if err != nil {return 0, err}
     
     return printerHandle, nil
}
 
//Version 1 code:
//
//User GoOpenPrinter to get the printer handle used for this function.
//This function is VERY simplistic. It may or may not work for your implementation. 
func GoPrint(printerHandle uintptr, path string) error {
     
     var err error

     startPrinter(printerHandle, path)
     startPagePrinter.Call(printerHandle)
     err = writePrinterFunc(printerHandle, path)
     endPagePrinter.Call(printerHandle)
     endDocPrinter.Call(printerHandle)
     
     return err
}

//Version 1 code:
//
//Close the printer once done with it. 
func GoClosePrinter(printerHandle uintptr) {

     closePrinter.Call(printerHandle)  
     
     return
}
 
func writePrinterFunc(printerHandle uintptr, path string) error {     
     
     fileContents, err := ioutil.ReadFile(path)     
     if err != nil { return err }
     var contentLen uintptr = uintptr(len(fileContents))
     var writtenLen int
     _, _, err = writePrinter.Call(printerHandle, uintptr(unsafe.Pointer(&fileContents[0])),  contentLen, uintptr(unsafe.Pointer(&writtenLen)))
     
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

//Version 1 code:
//
//Gets the printer name as a string and as a UTF16 string.
func GetDefaultPrinterName() (string, []uint16){

     var pn[256] uint16
     plen := len(pn)
     getDefaultPrinter.Call(uintptr(unsafe.Pointer(&pn)), uintptr(unsafe.Pointer(&plen)))
     printerName := syscall.UTF16ToString(pn[:])
     fmt.Println("Printer name:", printerName)     
     printer16 := syscall.StringToUTF16(printerName)     
     return printerName, printer16
}

//Version 1 code:
//
//Gets list of the available printers. These names can then be passed to GoOpenPrinter() 
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


///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//NEVER GO FULL RETARD:
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
