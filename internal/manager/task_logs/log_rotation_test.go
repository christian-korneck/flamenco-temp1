package task_logs

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func setUpTest(t *testing.T) string {
	temppath, err := ioutil.TempDir("", "testlogs")
	assert.NoError(t, err)
	return temppath
}

func tearDownTest(temppath string) {
	os.RemoveAll(temppath)
}

func TestCreateNumberedPath(t *testing.T) {
	temppath := setUpTest(t)
	defer tearDownTest(temppath)

	numtest := func(path string, number int, basepath string) {
		result := createNumberedPath(path)
		assert.Equal(t, numberedPath{path, number, basepath}, result)
	}

	numtest("", -1, "")
	numtest(" ", -1, " ")
	numtest("jemoeder.1", 1, "jemoeder")
	numtest("jemoeder.", -1, "jemoeder.")
	numtest("jemoeder", -1, "jemoeder")
	numtest("jemoeder.abc", -1, "jemoeder.abc")
	numtest("jemoeder.-4", -4, "jemoeder")
	numtest("jemoeder.1.2.3", 3, "jemoeder.1.2")
	numtest("jemoeder.001", 1, "jemoeder")
	numtest("jemoeder.01", 1, "jemoeder")
	numtest("jemoeder.010", 10, "jemoeder")
	numtest("jemoeder 47 42.327", 327, "jemoeder 47 42")
	numtest("/path/üničøde.327/.47", 47, "/path/üničøde.327/")
	numtest("üničøde.327.what?", -1, "üničøde.327.what?")
}

func TestNoFiles(t *testing.T) {
	temppath := setUpTest(t)
	defer tearDownTest(temppath)

	filepath := filepath.Join(temppath, "nonexisting.txt")
	err := rotateLogFile(zerolog.Nop(), filepath)
	assert.NoError(t, err)
	assert.False(t, fileExists(filepath))
}

func TestOneFile(t *testing.T) {
	temppath := setUpTest(t)
	defer tearDownTest(temppath)

	filepath := filepath.Join(temppath, "existing.txt")
	fileTouch(filepath)

	err := rotateLogFile(zerolog.Nop(), filepath)
	assert.NoError(t, err)
	assert.False(t, fileExists(filepath))
	assert.True(t, fileExists(filepath+".1"))
}

func TestMultipleFilesWithHoles(t *testing.T) {
	temppath := setUpTest(t)
	defer tearDownTest(temppath)

	filepath := filepath.Join(temppath, "existing.txt")
	assert.NoError(t, ioutil.WriteFile(filepath, []byte("thefile"), 0666))
	assert.NoError(t, ioutil.WriteFile(filepath+".1", []byte("file .1"), 0666))
	assert.NoError(t, ioutil.WriteFile(filepath+".2", []byte("file .2"), 0666))
	assert.NoError(t, ioutil.WriteFile(filepath+".3", []byte("file .3"), 0666))
	assert.NoError(t, ioutil.WriteFile(filepath+".5", []byte("file .5"), 0666))
	assert.NoError(t, ioutil.WriteFile(filepath+".7", []byte("file .7"), 0666))

	err := rotateLogFile(zerolog.Nop(), filepath)

	assert.NoError(t, err)
	assert.False(t, fileExists(filepath))
	assert.True(t, fileExists(filepath+".1"))
	assert.True(t, fileExists(filepath+".2"))
	assert.True(t, fileExists(filepath+".3"))
	assert.True(t, fileExists(filepath+".4"))
	assert.False(t, fileExists(filepath+".5"))
	assert.True(t, fileExists(filepath+".6"))
	assert.False(t, fileExists(filepath+".7"))
	assert.True(t, fileExists(filepath+".8"))
	assert.False(t, fileExists(filepath+".9"))

	read := func(filename string) string {
		content, err := ioutil.ReadFile(filename)
		assert.NoError(t, err)
		return string(content)
	}

	assert.Equal(t, "thefile", read(filepath+".1"))
	assert.Equal(t, "file .1", read(filepath+".2"))
	assert.Equal(t, "file .2", read(filepath+".3"))
	assert.Equal(t, "file .3", read(filepath+".4"))
	assert.Equal(t, "file .5", read(filepath+".6"))
	assert.Equal(t, "file .7", read(filepath+".8"))
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func fileTouch(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDONLY, 0666)
	if err != nil {
		panic(err.Error())
	}
	file.Close()
}
