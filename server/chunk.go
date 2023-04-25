package server

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type Chunk struct {
	Title    string
	SubTitle string
	Text     string
}

type QAChunk struct {
	Q          string
	A          string
	QEmbedding []float32
	Title      string
	SubTitle   string
}

const BaseFilePath = "./data"

func CreateChunk(dirname string) ([]Chunk, error) {
	newData := make([]Chunk, 0)
	path := BaseFilePath + "/" + dirname

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("there is no document in the path")
	}
	for _, f := range files {
		//fmt.Println(f.Name())
		name := f.Name()
		fullName := path + "/" + name
		data, err := ioutil.ReadFile(fullName)
		if err != nil {
			fmt.Printf("文件打开失败=%v\n", err)
			return nil, err
		}
		// fmt.Println(string(data))
		str := strings.Replace(string(data), " ", "", -1)
		// fmt.Println(str)
		newData = append(newData, Chunk{
			Title:    dirname,
			SubTitle: name,
			Text:     str,
		})
	}
	return newData, nil
}
