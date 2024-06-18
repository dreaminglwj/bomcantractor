package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/sirupsen/logrus"
)

type bomRecord struct {
	Reference string
	Value     string
	Datasheet string
	Footprint string
	Qty       int32
	DNP       string
	// Modul string
}

var indexByFootprintComponents = []string{"Willing_Library:paimu", "Willing_Library:niujiao", "Connector_PinHeader"}

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
	folder := "C:\\work\\willingCarPCB\\bom"
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
		filePath := fmt.Sprintf("%s\\%s", path, file.Name())
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
	csvName := fmt.Sprintf("%s\\bom-%s.csv", path, now.Format("20060102150405"))

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
