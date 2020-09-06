package main

import (
	"bufio"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLoadsRoute(t *testing.T) {
	// setupServer
	router := SetupServer()

	// Make a request to our server with the validateLoads
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/validateLoads", nil)
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
