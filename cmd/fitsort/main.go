// fitsort
//
// Walks a directory, parses all .fit files in it, and moves all the ones which are
// activities to a separate directory, with the file name containing timestamp,
// activity type and duration.
// Example usage:
//
//	go run ./cmd/fitsort/ -in=$HOME/Downloads/garmin-export/
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/typedef"
)

func parseFitActivity(path string) (*filedef.Activity, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open '%s': %w", path, err)
	}
	defer f.Close()

	lis := filedef.NewListener()
	defer lis.Close()

	dec := decoder.New(f,
		decoder.WithMesgListener(lis),
		decoder.WithBroadcastOnly(),
	)

	if _, err := dec.CheckIntegrity(); err != nil {
		return nil, fmt.Errorf("integrity check failed on '%s': %w", path, err)
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("seek failed: %w", err)
	}

	var file *filedef.Activity = nil
	i := 0
	for dec.Next() {
		if i > 0 {
			panic("not handled")
		}

		_, err := dec.Decode()
		if err != nil {
			return nil, fmt.Errorf("failed to decode '%s': %w", path, err)
		}

		var ok bool
		file, ok = lis.File().(*filedef.Activity)
		if !ok {
			return nil, fmt.Errorf("'%s' is not an activity file", path)
		}
		i++
	}

	if file == nil {
		return nil, fmt.Errorf("failed to find FIT data in '%s'", path)
	}

	return file, nil
}

func nameActivityFile(activity *filedef.Activity) (string, error) {
	if activity.Activity.Type != typedef.ActivityManual {
		return "", errors.New("sport type not set manually")
	}

	if len(activity.Sports) != 1 {
		return "", fmt.Errorf("expected exactly 1 sport, got %d", len(activity.Sports))
	}

	sport := activity.Sports[0].Sport.String()
	ts := activity.Activity.Timestamp
	dur := time.Duration(activity.Activity.TotalTimerTime) * time.Millisecond

	// // TODO: What to do for files where this results in 0?
	// // activity.Splits[].TotalDistance also does not work.
	// distKM := 0.
	// for _, summary := range activity.SplitSummaries {
	// 	distKM += float64(summary.TotalDistance) / 100 / 1000
	// }

	name := fmt.Sprintf("%s %s %s.fit", ts.Local().Format(time.RFC3339), sport, dur.Round(time.Minute))
	return filepath.Join(ts.Format("2006"), name), nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func main() {
	inDir := flag.String("in", ".", "input directory")
	outDir := flag.String("out", "out/", "output directory")
	flag.Parse()

	log.Println("input directory", *inDir)
	log.Println("output directory", *outDir)

	*outDir = filepath.Clean(*outDir)
	err := os.MkdirAll(*outDir, 0755)
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(*inDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() || !strings.HasSuffix(path, ".fit") {
			log.Println("skipping", path)
			return nil
		}

		activity, err := parseFitActivity(path)
		if err != nil {
			log.Println("not an activity file", err)
			return nil
		}

		newPath, err := nameActivityFile(activity)
		if err != nil {
			log.Printf("failed to name '%s': %s", path, err)
			return nil
		}

		newPath = filepath.Join(*outDir, newPath)
		os.MkdirAll(filepath.Dir(newPath), 0755)
		return copyFile(path, newPath)
	})
	if err != nil {
		panic(err)
	}
}
