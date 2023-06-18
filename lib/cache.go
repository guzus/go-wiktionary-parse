package lib

import (
	"encoding/gob"
	"github.com/macdub/go-colorlog"
	"os"
	"time"
)

var (
	logger *colorlog.ColorLog = &colorlog.ColorLog{}
)

// encode the data into a binary cache file
func EncodeCache(data *WikiData, file string) error {
	logger.Info("Creating binary cache: '%s'\n", file)
	cacheFile, err := os.Create(file)
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(cacheFile)

	start := time.Now()
	logger.Debug("Encoding data ... ")
	enc.Encode(data)
	end := time.Now()
	logger.Printc(colorlog.Ldebug, colorlog.Green, "elapsed %s\n", end.Sub(start))

	logger.Info("Binary cache built.\n")
	cacheFile.Close()

	return nil
}

// decode binary cache file into a usable struct
func DecodeCache(file string) (*WikiData, error) {
	logger.Info("Initializing cached object\n")
	cacheFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	data := &WikiData{}
	dec := gob.NewDecoder(cacheFile)

	start := time.Now()
	logger.Debug("Decoding data ... ")
	dec.Decode(data)
	end := time.Now()
	logger.Printc(colorlog.Ldebug, colorlog.Green, "elapsed %s\n", end.Sub(start))

	logger.Info("Cache initialized.\n")
	cacheFile.Close()

	return data, nil
}
