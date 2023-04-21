package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"seismo"
	"seismo/pr/seishub"
	"time"
)

const (
	defBaseAddr = "http://seishub.ru/pipermail/seismic-report/"
	defOutDir   = "out"

	//Modes
	listPageMode = "lp"
	msgPageMode  = "mp"
)

// Main checks all the flag values end runs functions of a specified mode logic
func main() {
	// There are 5 (five) flags: "from", "to", "baseAddr", "mode", "out"
	t := time.Now()
	curMY := seismo.MonthYear{Month: t.Month(), Year: t.Year()}
	fromFlag := seismo.MonthYearFlag("from", curMY, "start point in month/year format")
	toFlag := seismo.MonthYearFlag("to", curMY, "end point in month/year format")

	baseAddrFlag := flag.String("baseAddr", "", "base address (url)")

	modeFlagUsage := fmt.Sprintf("%s - get month pages containting list message names, %s - get message pages", listPageMode, msgPageMode)
	modeFlag := flag.String("mode", listPageMode, modeFlagUsage)

	outFlag := flag.String("out", "", "output folder")

	flag.Parse()

	if fromFlag.Date().After(toFlag.Date()) {
		fmt.Println(`The value of the "from" flag cannot be more than the value of the "to" flag.`)
		return
	}

	if *baseAddrFlag == "" {
		fmt.Printf("The value of the \"baseAddr\" flag is not specified. The default value \"%s\" will be used.\n", defBaseAddr)
		*baseAddrFlag = defBaseAddr
	}

	if *outFlag == "" {
		ep, err := os.Executable()
		if err != nil {
			fmt.Printf("Cannot define the path to the executable file: %v.\n", err)
			return
		}

		fmt.Printf("The value of the out flag is not specified.The default value \"%s\" will be used.\n", defOutDir)
		*outFlag = path.Join(path.Dir(ep), defOutDir)
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
	default:
		fmt.Println("A mode specified incorrectly.")
		return
	}
}

func getMsgPages(from, to seismo.MonthYear, baseAddr, saveDir string) error {
	for my := from.Date(); !my.After(to.Date()); my = my.AddDate(0, 1, 0) {
		sg := monthYearPathSeg(my.Month(), my.Year())
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

		err = saveMsgPages(msgs, saveSubDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveMsgPages(msgs map[string]string, saveDir string) error {
	for name, msg := range msgs {
		err := saveFile(path.Join(saveDir, name), msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func getListPages(from, to seismo.MonthYear, baseAddr, saveDir string) error {
	for my := from.Date(); !my.After(to.Date()); my = my.AddDate(0, 1, 0) {
		sg := monthYearPathSeg(my.Month(), my.Year())
		url, err := url.JoinPath(baseAddr, sg)
		if err != nil {
			return err
		}

		pg, err := seishub.GetMsgNamesPage(context.Background(), url)
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

func monthYearPathSeg(m time.Month, y int) string {
	return fmt.Sprintf("%d-%s", y, m.String())
}
