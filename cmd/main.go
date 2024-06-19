package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/sirupsen/logrus"
)

const (
	timeFormatStr string = "20060102150405"
	osWindowsStr  string = "windows"
)

type bomRecord struct {
	Reference string
	Value     string
	Datasheet string
	Footprint string
	Qty       int32
	DNP       string
	// Modul string
	FootprintAlias string
}

var indexByFootprintComponents = []string{"Willing_Library:paimu", "Willing_Library:niujiao", "Connector_PinHeader"}

var footprintTrasnlate = map[string]string{
	"Capacitor_SMD:C_0603_1608Metric":                          "0603",
	"Capacitor_SMD:C_0805_2012Metric":                          "0805",
	"Capacitor_SMD:C_1206_3216Metric":                          "1206",
	"Resistor_SMD:R_0603_1608Metric":                           "0603",
	"Resistor_SMD:R_0805_2012Metric":                           "0805",
	"Resistor_SMD:R_1206_3216Metric":                           "1206",
	"Resistor_SMD:R_1206_3216Metric_Pad1.30x1.75mm_HandSolder": "1206",
	"LED_SMD:LED_0402_1005Metric":                              "0402",
	"LED_SMD:LED_0603_1608Metric":                              "0603",
	"LED_SMD:LED_0603_1608Metric_Pad1.05x0.95mm_HandSolder":    "0603",
	"LED_SMD:LED_0805_2012Metric":                              "0805",
	"LED_SMD:LED_1206_3216Metric":                              "1206",
	// "":                                                         "",
	// "":                                                         "",
	// "":                                                         "",
	// "":                                                         "",
}

var bomMap map[string]*bomRecord

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("bom cantractor start...")

	// 初始化map
	bomMap = make(map[string]*bomRecord, 0)
	// ToDo：从命令行读取路径

	// 读取目录下的所有bom文件名
	// folder := "C:\\work\\willingCarPCB\\bom"  // C:\work\willingCarPCB\bom
	folder := os.Args[1]

	// folder := "/Users/luwenjin/work/willing/willingCarPCB/bom"
	if len(folder) <= 0 {
		logrus.Error("invaild path\n")
		return
	}
	files := readFolder(folder)

	// 读取所有bom数据，并进行整理
	for _, file := range files {
		readbom(file)
	}

	// 打印bom
	printMap()

	bomRecords := mapToSlice()

	saveBomToCSV(folder, bomRecords)
}

func readFolder(path string) []string {
	filePaths := make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, file := range files {
		// logrus.Infof("files: %+v\n", file)
		filePath := fmt.Sprintf("%s/%s", path, file.Name())
		if runtime.GOOS == osWindowsStr {
			filePath = fmt.Sprintf("%s\\%s", path, file.Name())
		}

		if strings.HasSuffix(filePath, ".csv") {
			logrus.Infof("%s\n", filePath)
			filePaths = append(filePaths, filePath)
		}
	}

	return filePaths
}

func readbom(fileName string) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		logrus.Errorf("打开文件失败：%s, %+v\n", fileName, err)
		return
	}

	defer file.Close()

	var bomRecords []*bomRecord
	err = gocsv.UnmarshalFile(file, &bomRecords)
	if err != nil {
		logrus.Errorf("读取文件失败：%s, %+v\n", fileName, err)
		return
	}

	// 读取成功，便利records，插入到map
	for _, record := range bomRecords {
		key := record.Footprint
		if !indexByFootprint(record.Footprint) {
			key = fmt.Sprintf("%s-%s", record.Footprint, record.Value)
		}

		record.FootprintAlias = footprintTrasnlate[record.Footprint]

		mapRecord := bomMap[key]
		if mapRecord != nil {
			mapRecord.Reference = mapRecord.Reference + "," + record.Reference // todo: 增加一个map，用于将同一个文件中的相同型号的器件放到一个组中，统一加前缀
			mapRecord.Qty += record.Qty
		} else {
			bomMap[key] = record
		}
	}
}

func indexByFootprint(footprint string) bool {
	for _, component := range indexByFootprintComponents {
		if strings.Contains(footprint, component) {
			return true
		}
	}

	return false
}

func printMap() {
	for _, v := range bomMap {
		logrus.Infof("%+v\n", v)
	}
}

func mapToSlice() []*bomRecord {
	// 排序
	var sortKeys []string
	for k, _ := range bomMap {
		sortKeys = append(sortKeys, k)
	}

	sort.Strings(sortKeys)

	var bomRecords []*bomRecord

	for _, k := range sortKeys {
		record := bomMap[k]
		bomRecords = append(bomRecords, record)
	}

	// for _, v := range bomMap {
	// 	bomRecords = append(bomRecords, v)
	// }

	return bomRecords
}

func saveBomToCSV(path string, bombomRecords []*bomRecord) {
	now := time.Now()
	csvName := fmt.Sprintf("%s/bom-%s.csv", path, now.Format(timeFormatStr))
	if runtime.GOOS == osWindowsStr {
		csvName = fmt.Sprintf("%s\\bom-%s.csv", path, now.Format(timeFormatStr))
	}

	file, err := os.OpenFile(csvName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		logrus.Errorf("create file fail: file-%s  err-%+v\n", csvName, err)
		return
	}

	defer file.Close()

	err = gocsv.MarshalFile(&bombomRecords, file)
	if err != nil {
		logrus.Errorf("save csv file fail: file-%s  err-%+v\n", csvName, err)
		return
	}
}
