package main

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name      string    `yaml:"name"`
	EventTime time.Time `yamel:"eventTime"`
}

type DataConfig struct {
	Data []Config `yaml:"data"`
}

func writeYaml(src string, data DataConfig) {

	d, err := yaml.Marshal(data) // 第二个表示每行的前缀，这里不用，第三个是缩进符号，这里用tab
	checkError(err)
	err = ioutil.WriteFile(src, d, 0777)
	checkError(err)

}

func readYaml(src string) *DataConfig {

	data := &DataConfig{}

	content, err := ioutil.ReadFile(src)
	checkError(err)

	err = yaml.Unmarshal(content, data)
	checkError(err)
	return data
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
