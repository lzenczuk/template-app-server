package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func readUint32(input io.ReadCloser) uint32 {
	var value uint32

	err := binary.Read(input, binary.LittleEndian, &value)
	if err != nil {
		panic(err)
	}

	return value
}

func readText(input io.ReadCloser) string {
	textLength := readUint32(input)

	textBuffer := make([]byte, textLength)

	_, err := input.Read(textBuffer)
	if err != nil && err != io.EOF {
		panic(err)
	}

	return string(textBuffer)
}

type textFile struct {
	path    string
	content string
}

func readTextFile(input io.ReadCloser) *textFile {
	tf := textFile{}

	// read first file length
	// because in golang ReadCloser is stateful
	// we don't need this number
	readUint32(input)

	tf.path = readText(input)
	tf.content = readText(input)

	return &tf
}

type project struct {
	name  string
	files []textFile
}

func readProject(input io.ReadCloser) project {

	p := project{}

	// skip length of project name
	readUint32(input)

	p.name = readText(input)

	numberOfFiles := readUint32(input)
	p.files = make([]textFile, numberOfFiles)

	for i := 0; uint32(i) < numberOfFiles; i++ {
		// skip text file length section
		readUint32(input)

		p.files[i] = *readTextFile(input)
	}

	fmt.Println("Project:", p)
	return p
}

func saveTextFile(rootFolder string, tf *textFile) {

	sPath := strings.Split(tf.path, "/")

	var path []string
	for _, s := range sPath {
		if s != "" {
			path = append(path, strings.TrimSpace(s))
		}
	}

	fmt.Println(path)

	targetFolder := rootFolder + "/" + strings.Join(path[0:len(path)-1], "/")
	fmt.Println("Target: " + targetFolder)

	err := os.MkdirAll(targetFolder, 0700)
	if err != nil {
		fmt.Println("Error creating path: ", targetFolder)
		panic(err)
	}

	fullPath := targetFolder + "/" + path[len(path)-1]
	f, err := os.Create(fullPath)
	if err != nil {
		fmt.Println("Error creating file path: ", fullPath)
		panic(err)
	}

	_, err = f.WriteString(tf.content)
	if err != nil {
		fmt.Println("Error writing to file path: ", fullPath)
		panic(err)
	}

	f.Close()
}

func clearRootFolder(destFolder string) {
	err := os.RemoveAll(destFolder)
	if err != nil {
		panic(err)
	}

	err = os.Mkdir(destFolder, 0700)
	if err != nil {
		panic(err)
	}
}

func main() {

	rootFolder := "/home/lzenczuk/go/src/eos-p2p/tmp"

	http.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {

		p := readProject(request.Body)

		clearRootFolder(rootFolder)

		for _, tf := range p.files {
			saveTextFile(rootFolder, &tf)
		}
	})

	err := http.ListenAndServe(":7886", nil)
	if err != nil {
		panic(err)
	}
}
