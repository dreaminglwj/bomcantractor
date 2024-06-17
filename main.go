package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Gocarina/gocsv"
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

var bomMap map[string]*bomRecord

func main() {
	// 初始化map
	bomMap = make(map[string]*bomRecord, 0)
	// ToDo：从命令行读取路径

	// 读取目录下的所有bom文件名
	folder := ""
	files := readFolder(folder)

	// 读取所有bom数据，并进行整理
	for _, file := range files {
		readbom(file)
	}

	// 打印bom
	printMap()
}

func readFolder(path string) []string {
	filePaths := make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		filePath := fmt.Sprintf("%s/%s", file)
		if strings.HasSuffix(".csv") {
			fmt.Printf("%s\n", filePath)
			filePaths = append(filePaths, filePath)
		}
	}

	return filePaths
}

func readbom(fileName string) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Printf("打开文件失败：%s, %+v\n", fileName, err)
		return
	}

	defer file.Close()

	var bomRecords []*bomRecord
	err = gocsv.UnmarshalFile(file, &bomRecords)
	if err != nil {
		fmt.Printf("读取文件失败：%s, %+v\n", fileName, err)
		return
	}

	// 读取成功，便利records，插入到map
	for _, record := range bomRecords {
		key := fmt.Sprintf("%s-%s", record.Value, record.Footprint)
		mapRecord := bomMap[key]
		if mapRecord != nil {
			mapRecord.Value += record.Value // todo: 增加一个map，用于将同一个文件中的相同型号的器件放到一个组中，统一加前缀
			mapRecord.Qty += record.Qty
		} else {
			bomMap[key] = record
		}
	}
}

func printMap() {
	for _, v := range bomMap {
		fmt.Printf("%+v", v)
	}
}
