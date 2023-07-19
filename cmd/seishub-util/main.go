package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"seismo/provider"
	"seismo/provider/seishub"
	"time"
)

const (
	defBaseAddr = "http://seishub.ru/pipermail/seismic-report/"
	defOutDir   = "out"

	//Modes
	listPageMode   = "lp"
	msgPageMode    = "mp"
	parseFilesMode = "pf"
	//Max input file size in bytes
	maxInputSize = 1024 * 10 //10 KB
)

// Main checks all the flag values end runs functions of a specified mode logic
func main() {
	// There are 5 (five) flags: "from", "to", "baseAddr", "mode", "out"
	t := time.Now()
	curMY := provider.MonthYear{Month: t.Month(), Year: t.Year()}
	fromFlag := provider.MonthYearFlag("from", curMY, "start point in month/year format")
	toFlag := provider.MonthYearFlag("to", curMY, "end point in month/year format")

	baseAddrFlag := flag.String("baseAddr", "", "base address (url)")

	modeFlagUsage := fmt.Sprintf("%s - get month pages containting list message names, %s - get message pages", listPageMode, msgPageMode)
	modeFlag := flag.String("mode", listPageMode, modeFlagUsage)

	outFlag := flag.String("out", "", "output folder")
	inFlag := flag.String("in", "", "input folder")

	flag.Parse()

	if fromFlag.Date().After(toFlag.Date()) {
		fmt.Println(`The value of the "from" flag cannot be more than the value of the "to" flag.`)
		return
	}

	if *baseAddrFlag == "" {
		fmt.Printf("The value of the \"baseAddr\" flag is not specified. The default value %q will be used.\n", defBaseAddr)
		*baseAddrFlag = defBaseAddr
	}

	ep, err := os.Executable()
	if err != nil {
		fmt.Printf("Cannot define the path to the executable file: %v.\n", err)
		return
	}

	if *outFlag == "" {
		fmt.Printf("The value of the \"out\" flag is not specified.The default value %q will be used.\n", defOutDir)
		*outFlag = path.Join(path.Dir(ep), defOutDir)
	}

	if *inFlag == "" {
		fmt.Printf("The value of the \"in\" flag is not specified. The folder of the executable will be used.\n")
		*inFlag = ep
	}

	if err := os.MkdirAll(*outFlag, os.ModePerm); err != nil {
		fmt.Printf("Cannot create a specified output directory: %v.\n", err)
		return
	}

	//Main logic
	switch *modeFlag {
	case listPageMode:
		err := getListPages(*fromFlag, *toFlag, *baseAddrFlag, *outFlag)
		if err != nil {
			fmt.Printf("Getting month list pages error: %v.\n", err)
		}
	case msgPageMode:
		err := getMsgPages(*fromFlag, *toFlag, *baseAddrFlag, *outFlag)
		if err != nil {
			fmt.Printf("Getting message pages error: %v.\n", err)
		}
	case parseFilesMode:
		err := parseMsgFiles(*inFlag, *outFlag)
		if err != nil {
			fmt.Printf("Parse files error: %v.\n", err)
		}
	default:
		fmt.Println("A mode specified incorrectly.")
		return
	}
}

func parseMsgFiles(inputDir, saveDir string) error {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(saveDir, os.ModePerm)
	if err != nil {
		return err
	}

	for _, f := range files {
		inf, err := f.Info()
		if err != nil {
			log.Printf("Skiping. Cannot read info for %q\n", f.Name())
			continue
		}

		if f.IsDir() || inf.Size() > maxInputSize {
			log.Printf("Skiping. %q is a folder or too big.\n", f.Name())
			continue
		}

		bf, err := os.ReadFile(path.Join(inputDir, f.Name()))
		if err != nil {
			log.Printf("Skiping. Cannot read %q: %v\n", f.Name(), err)
			continue
		}

		msg, err := seishub.ParseMsg(string(bf))
		if err != nil {
			log.Printf("Skiping. Cannot parse %q: %v\n", f.Name(), err)
			continue
		}
		//msg.Link = f.Name()

		js, err := json.MarshalIndent(msg, "", " ")
		if err != nil {
			return err
		}

		err = saveFile(path.Join(saveDir, f.Name()+".json"), string(js))
		if err != nil {
			return err
		}
	}
	return nil
}

func getMsgPages(from, to provider.MonthYear, baseAddr, saveDir string) error {
	for my := from.Date(); !my.After(to.Date()); my = my.AddDate(0, 1, 0) {
		sg := seishub.MonthYearPathSeg(my.Month(), my.Year())
		url, err := url.JoinPath(baseAddr, sg)
		if err != nil {
			return err
		}

		msgs, err := seishub.GetMsgPages(context.Background(), url)
		if err != nil {
			return err
		}

		saveSubDir := path.Join(saveDir, sg)
		err = os.MkdirAll(saveSubDir, os.ModePerm)
		if err != nil {
			return err
		}

		err = saveMappedTexts(msgs, saveSubDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func getListPages(from, to provider.MonthYear, baseAddr, saveDir string) error {
	for my := from.Date(); !my.After(to.Date()); my = my.AddDate(0, 1, 0) {
		sg := seishub.MonthYearPathSeg(my.Month(), my.Year())
		url, err := url.JoinPath(baseAddr, sg)
		if err != nil {
			return err
		}

		pg, err := seishub.GetMsgNamesPage(context.Background(), url, nil)
		if err != nil {
			return err
		}

		err = saveFile(path.Join(saveDir, sg+".html"), pg)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveFile(path, cont string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(cont); err != nil {
		return err
	}
	return nil
}

func saveMappedTexts(msgs map[string]string, saveDir string) error {
	for name, msg := range msgs {
		err := saveFile(path.Join(saveDir, name), msg)
		if err != nil {
			return err
		}
	}
	return nil
}
