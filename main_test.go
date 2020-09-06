package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLoadsRoute(t *testing.T) {
	// setupServer
	router := SetupServer()
	// read testInput.txt and send as payload
	fileDir, _ := os.Getwd()
	fileName := "testInput.txt"
	filePath := path.Join(fileDir, fileName)

	file, _ := os.Open(filePath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	// Make a request to our server with the validateLoads
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/validateLoads", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	router.ServeHTTP(w, req)

	expected := ""
	readFile, err := os.Open("./testOutput.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer readFile.Close()
	// read from text file line by line
	scanner := bufio.NewScanner(readFile)
	for scanner.Scan() {
		expected += scanner.Text()
	}

	assert.Equal(t, http.StatusOK, w.Code)
	// replce extra special characters
	re := strings.NewReplacer("\\n", "", "\"", "", "\\", "")
	// test api output against expected output from testOutput.txt file
	assert.Equal(t, re.Replace(w.Body.String()), re.Replace(expected))
}
