package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"log"
	"strings"
)

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

type ConfigSection struct {
	Name      string
	KeyValue  map[string]string
	IsDefault bool
}

type Config struct {
	Sections []ConfigSection
}

func (conf Config) GetConfigIndicated(section string, key string) (string, error) {
	var value string
	var ok bool
	var findSection = false
	if section == "" {
		section = "DEFAULT"
	}
	for i := range conf.Sections {
		if conf.Sections[i].Name == section {
			findSection = true
			value, ok = conf.Sections[i].KeyValue[key]
			if ok {
				break
			} else {
				return "", errors.New("key: " + key + " not exist")
			}
		}
	}
	if findSection {
		return value, nil
	} else {
		return "", errors.New("section: " + section + " not exist")
	}
}

func (conf Config) PrintSection(section string) {
	for i := range conf.Sections {
		if conf.Sections[i].Name == section {
			for k, v := range conf.Sections[i].KeyValue {
				fmt.Printf("%s = %s\n", k, v)
			}
		}
	}
}

func main() {
	fileName := "conf_parse.log"
	logFile, err := os.Create(fileName)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("open file conf_parse.log error")
	}
	infoLog := log.New(logFile, "[Info]", log.Llongfile)

	var configFile, configKey string
	flag.StringVar(&configFile, "configFile", "", "Absolute config file path.")
	flag.StringVar(&configKey, "configKey", "", "Section.Key format about key.")
	flag.Parse()

	if configFile == "" {
		fmt.Println("Wrong configFile path, try 'config_parse -h' for more infomation.")
		return
	}
	if !checkFileIsExist(configFile) {
		fmt.Println("configFile not exist.")
		return
	}

	file, err := os.OpenFile(configFile, os.O_RDONLY|os.O_RDWR, 0666)
	if err != nil {
		infoLog.Printf("Open file error: %s\n", err)
		return
	}
	defer file.Close()

	var fileConfig Config
	var firstLine, finishOneSection = false, false
	var lastLine = ""
	var lineNum = 1
	reader := bufio.NewReader(file)
	for {
		var str string
		if finishOneSection {
			str = lastLine
			finishOneSection = false
		} else {
			str, err = reader.ReadString('\n')
			lineNum += 1
			if err != nil {
				infoLog.Printf("Read file reach EOF\n")
				break
			}
		}

		var theLine = strings.TrimSpace(str)
		if strings.HasPrefix(theLine, "#") {
			infoLog.Printf("find a config comment line: %s\n", theLine)
			continue
		}

		var tempSection ConfigSection
		tempSection.KeyValue = make(map[string]string)
		if strings.HasPrefix(theLine, "[") && strings.HasSuffix(theLine, "]") {
			firstLine = true
			if theLine == "[DEFAULT]" {
				tempSection.IsDefault = true
				tempSection.Name = "DEFAULT"
			} else {
				tempSection.IsDefault = false
				tempSection.Name = theLine[1:len(theLine)-1]
			}
			for {
				secStr, err := reader.ReadString('\n')
				lineNum += 1
				if err != nil {
					infoLog.Printf("Read file reach EOF\n")
					break
				}
				var secTheLine = strings.TrimSpace(secStr)
				if strings.HasPrefix(secTheLine, "[") && strings.HasSuffix(secTheLine, "]") {
					lastLine = secTheLine
					finishOneSection = true
					break
				} else if strings.HasPrefix(secTheLine, "#") {
					infoLog.Printf("find a config comment line: %s\n", secTheLine)
					continue
				} else if secTheLine == "" {
					continue
				} else {
					var sliceLine = strings.SplitN(secTheLine, "=", 2)
					if len(sliceLine) != 2 {
						infoLog.Printf("find a line is not correct: %s\n", secTheLine)
						fmt.Printf("Line: %d is invalid: %s\n", lineNum - 1, secTheLine)
						return
					}
					tempSection.KeyValue[strings.TrimSpace(sliceLine[0])] = strings.TrimSpace(sliceLine[1])
					infoLog.Printf("key: %s, value: %s\n", sliceLine[0], sliceLine[1])
				}
			}
			fileConfig.Sections = append(fileConfig.Sections, tempSection)
		} else if !firstLine {
			infoLog.Printf(theLine)
			infoLog.Printf("invalid config format, first available line must be a section")
			return
		}
	}

	if configKey == "" {
		for t := range fileConfig.Sections {
			fmt.Println("========================================")
			fmt.Println(fileConfig.Sections[t].Name)
			fileConfig.PrintSection(fileConfig.Sections[t].Name)
		}
	} else {
		keyValSlice := strings.SplitN(configKey, ".", 2)
		if len(keyValSlice) != 2 {
			var temp []string
			keyValSlice = append(temp, "DEFAULT", keyValSlice[0])
		}
		val, err := fileConfig.GetConfigIndicated(keyValSlice[0], keyValSlice[1])
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(val)
		}
	}
}
