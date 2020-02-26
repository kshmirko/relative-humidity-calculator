package licelformat

import (
	"bufio"
	//"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/archsh/timefmt"
	//"gonum.org/v1/gonum/mat"
)

const (
	LICEL_MAX_HEADER_LEN = 80
	LICEL_MAX_RESERVED   = 3
)

// LicelProfile - представляет отдельный измерительный канал
type LicelProfile struct {
	Active, Photon bool
	LaserType      int64
	NDataPoints    int64
	reserved       [LICEL_MAX_RESERVED]int64
	HighVoltage    int64
	BinWidth       float64
	Wavelength     float64
	Polarization   string
	BinShift       int64
	DecBinShift    int64
	AdcBits        int64
	NShots         int64
	DiscrLevel     float64
	DeviceID       string
	NCrate         int64
	Data           []float64
}

type LicelProfilesList []LicelProfile

// LicelFile - Структура - единичное измерение на дидаре
type LicelFile struct {
	MeasurementSite       string
	MeasurementStartTime  time.Time
	MeasurementStopTime   time.Time
	AltitudeAboveSeaLevel float64
	Longitude             float64
	Latitude              float64
	Zenith                float64
	Laser1NShots          int64
	Laser1Freq            int64
	Laser2NShots          int64
	Laser2Freq            int64
	NDatasets             int64
	Laser3NShots          int64
	Laser3Freq            int64
	FileLoaded            bool
	Profiles              LicelProfilesList
}

type LicelPack map[string]LicelFile

func NewLicelProfile(line string) (profile LicelProfile) {
	items := strings.Split(line, " ")
	wvlpol := strings.Split(items[7], ".")
	NDataPoints := str2Int(items[3])
	profile = LicelProfile{
		Active:       str2Bool(items[0]),
		Photon:       str2Bool(items[1]),
		LaserType:    str2Int(items[2]),
		NDataPoints:  NDataPoints,
		reserved:     [3]int64{str2Int(items[4]), str2Int(items[8]), str2Int(items[9])},
		HighVoltage:  str2Int(items[5]),
		BinWidth:     str2Float(items[6]),
		Wavelength:   str2Float(wvlpol[0]),
		Polarization: wvlpol[1],
		BinShift:     str2Int(items[10]),
		DecBinShift:  str2Int(items[11]),
		AdcBits:      str2Int(items[12]),
		NShots:       str2Int(items[13]),
		DiscrLevel:   str2Float(items[14]),
		DeviceID:     items[15][:2],
		NCrate:       str2Int(items[15][2:]),
		Data:         nil, //make([]int32, NDataPoints),
	}
	return
}

// LicelFile - загружает содержимое файла измерений в память
func LoadLicelFile(fname string) (licf LicelFile) {

	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	r := bufio.NewReader(f)
	licf = LicelFile{}

	if header_line, err := r.ReadString(10); err == nil {
		header_line = header_line[:len(header_line)-2]
		//licf.MeasurementSite = strings.Trim(header_line, " ")
	}

	if header_line, err := r.ReadString(10); err == nil {
		header_line = header_line[:len(header_line)-2]
		tmp := strings.Split(header_line, " ")
		licf.MeasurementSite = tmp[0]
		licf.MeasurementStartTime, err = timefmt.Strptime(tmp[2]+" "+tmp[3], "%d/%m/%Y %H:%M:%S")
		if err != nil {
			log.Fatal(err)
		}

		licf.MeasurementStopTime, err = timefmt.Strptime(tmp[4]+" "+tmp[5], "%d/%m/%Y %H:%M:%S")
		if err != nil {
			log.Fatal(err)
		}

		licf.AltitudeAboveSeaLevel = str2Float(tmp[6])
		licf.Longitude = str2Float(tmp[7])
		licf.Latitude = str2Float(tmp[8])
		licf.Zenith = str2Float(tmp[9])
	}

	if header_line, err := r.ReadString(10); err == nil {
		header_line = header_line[:len(header_line)-2]
		tmp := strings.Split(header_line, " ")
		//licf.MeasurementSite = strings.Trim(header_line, " ")
		licf.Laser1NShots = str2Int(tmp[1])
		licf.Laser1Freq = str2Int(tmp[2])
		licf.Laser2NShots = str2Int(tmp[3])
		licf.Laser2Freq = str2Int(tmp[4])
		licf.NDatasets = str2Int(tmp[5])
		licf.Laser3NShots = str2Int(tmp[6])
		licf.Laser3Freq = str2Int(tmp[7])
	}

	licf.Profiles = make(LicelProfilesList, licf.NDatasets)
	for i := int64(0); i < licf.NDatasets; i++ {
		if header_line, err := r.ReadString(10); err == nil {
			header_line = strings.Trim(header_line[:len(header_line)-1], " ")
			licf.Profiles[i] = NewLicelProfile(header_line)
		}
	}

	crlf := make([]byte, 2)
	r.Read(crlf)
	//fmt.Printf("CRLF=[%d,%d]\n", crlf[0], crlf[1])
	for i := int64(0); i < licf.NDatasets; i++ {
		pr_tmp := make([]byte, licf.Profiles[i].NDataPoints*4)

		if n, err := io.ReadFull(r, pr_tmp); err != nil {
			log.Fatal(n, err)
		}

		licf.Profiles[i].Data = bytes2Float64Arr(pr_tmp)
		r.Read(crlf)
		//fmt.Printf("CRLF=[%d,%d]\n", crlf[0], crlf[1])
	}

	return
}

func str2Bool(str string) bool {
	if v, err := strconv.ParseBool(str); err == nil {
		return v
	}
	return false
}

func str2Int(str string) int64 {
	if v, err := strconv.ParseInt(str, 10, 32); err == nil {
		return v
	}
	return -9999
}

func str2Float(str string) float64 {
	if v, err := strconv.ParseFloat(str, 32); err == nil {
		return v
	}
	return -9999.999
}

func bytes2Int32(b []byte) (r int32) {
	r = 0
	if len(b) >= 4 {
		r |= int32(b[0])
		r |= int32(b[1]) << 8
		r |= int32(b[2]) << 16
		r |= int32(b[3]) << 24
	}
	return
}

func bytes2Int32Arr(b []byte) (r []int32) {
	i32len := int32(len(b) / 4)
	r = make([]int32, i32len)
	for i := int32(0); i < i32len; i++ {
		r[i] = bytes2Int32(b[i*4 : (i+1)*4])
	}
	return
}

func bytes2Float64Arr(b []byte) (r []float64) {
	i32len := int32(len(b) / 4)
	r = make([]float64, i32len)
	for i := int32(0); i < i32len; i++ {
		r[i] = float64(bytes2Int32(b[i*4:(i+1)*4]) * 1.0)
	}
	return
}

func NewLicelPack(mask string) (pack LicelPack) {
	pack = make(LicelPack)
	if matches, err := filepath.Glob(mask); err == nil {
		for _, fname := range matches {

			pack[fname] = LoadLicelFile(fname)
		}
	}
	return
}

func SelectCertainWavelength1(lf *LicelFile, isPhoton bool, wavelength float64) (lp LicelProfile) {
	for _, v := range lf.Profiles {
		if v.Photon == isPhoton && v.Wavelength == wavelength {
			return v
		}
	}
	return
}

func SelectCertainWavelength2(lp *LicelPack, isPhoton bool, wavelength float64) (lpl LicelProfilesList) {
	for _, v := range *lp {
		tmpProfile := SelectCertainWavelength1(&v, isPhoton, wavelength)
		lpl = append(lpl, tmpProfile)
	}
	return
}
