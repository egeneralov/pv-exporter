package dirsize

import (
	"math"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func Round(val float64, roundOn float64, places int) (newVal float64) {
	log.Debugf("dirsize.Round(\"%f\",\"%f\",\"%d\")", val, roundOn, places)

	var round float64

	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func DirSize(path string) int64 {
	log.Debugf("dirsize.DirSize(\"%v\")", path)

	sizes := make(chan int64, 1000)

	readSize := func(path string, file os.FileInfo, err error) error {
		if err != nil || file == nil {
			return nil
		}
		if !file.IsDir() {
			sizes <- file.Size()
		}
		return nil
	}

	go func() {
		filepath.Walk(path, readSize)
		close(sizes)
	}()

	size := int64(0)
	for s := range sizes {
		size += s
	}

	return size
}

func DirSizeMB(path string) float64 {
	log.Debugf("dirsize.DirSizeMB(\"%v\")", path)
	return float64(DirSize(path)) / 1024.0 / 1024.0
}
