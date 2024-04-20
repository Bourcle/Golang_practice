package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Read2VecProcess struct {
	SampleID   string
	RefPanel   string
	SampleGMSV string
	InputPath  string
	GMSVPath   string
	OutputPath string
}

func NewRead2VecProcess(inputPath, sampleID, refPanel string) *Read2VecProcess {
	return &Read2VecProcess{
		SampleID:   sampleID,
		RefPanel:   refPanel,
		SampleGMSV: fmt.Sprintf("%s.gMSV", sampleID),
		InputPath:  filepath.Join(inputPath, sampleID),
		GMSVPath:   filepath.Join(inputPath, "07.cpg_report"),
		OutputPath: filepath.Join(inputPath, "08.read2vec"),
	}
}

func (r *Read2VecProcess) RunCmd(cmd string) error {
	cmdParts := strings.Fields(cmd)
	command := exec.Command(cmdParts[0], cmdParts[1:]...)
	err := command.Run()
	if err != nil {
		return fmt.Errorf("something went wrong while executing %s: %v", cmd, err)
	}
	return nil
}

func (r *Read2VecProcess) MakeDir() error {
	fmt.Printf("<Make Encoding Directory Under %s>\n", r.SampleID)
	if _, err := os.Stat(r.OutputPath); os.IsNotExist(err) {
		err := os.MkdirAll(r.OutputPath, 0755)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("08.Encoding directory already exists")
	}
	return nil
}

func (r *Read2VecProcess) CpGCount(line string) int {
	if line == "NA" {
		return 0
	}
	return len(strings.Split(line, ";"))
}

func (r *Read2VecProcess) CheckProcess(process string) int {
	numLines := 0
	checkPath := ""
	if process == "gMSV" {
		checkPath = filepath.Join(r.InputPath, "07.gMSV", fmt.Sprintf("%s.%s", r.SampleID, process))
	} else {
		checkPath = filepath.Join(r.OutputPath, fmt.Sprintf("%s.%s", r.SampleID, process))
	}

	file, err := os.Open(checkPath)
	if err != nil {
		fmt.Printf("error opening file %s: %v\n", checkPath, err)
		return numLines
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		numLines++
		if numLines >= 5 {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("error scanning file %s: %v\n", checkPath, err)
	}
	return numLines
}

func (r *Read2VecProcess) GMSVToBed() error {
	// implementation
	return nil
}

func (r *Read2VecProcess) IntersectBed() error {
	// implementation
	return nil
}

func (r *Read2VecProcess) BedToEncoding() error {
	// implementation
	return nil
}

func (r *Read2VecProcess) MakeBed() error {
	fmt.Printf("<Make Bed of %s>\n", r.SampleID)
	if _, err := os.Stat(filepath.Join(r.OutputPath, fmt.Sprintf("%s.gMSV2bed.bed", r.SampleID))); err == nil {
		fmt.Println("Bed file already exists")
		if r.CheckProcess("gMSV2bed.bed") >= 5 {
			fmt.Println("Make bed process is already done")
		} else {
			if r.CheckProcess("gMSV") < 5 {
				return fmt.Errorf("%s went wrong, please check %s.gMSV file", r.SampleID, r.SampleID)
			}
			fmt.Printf("Convert %s.gMSV to Bed because it does not exist\n", r.SampleID)
			if err := r.GMSVToBed(); err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("Convert %s.gMSV to Bed because it does not exist\n", r.SampleID)
		if err := r.GMSVToBed(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Read2VecProcess) RunBedtools() error {
	fmt.Printf("<Run BedTools to intersect Bed by region of %s>\n", r.RefPanel)
	if _, err := os.Stat(filepath.Join(r.OutputPath, fmt.Sprintf("%s.intersected.bed", r.SampleID))); err == nil {
		fmt.Println("Intersect Bed file already exists")
		if r.CheckProcess("intersected.bed") >= 5 {
			fmt.Println("Intersect Bed is already done")
		} else {
			if r.CheckProcess("gMSV2bed.bed") < 5 {
				return fmt.Errorf("%s went wrong, please check %s.gMSV2bed.bed", r.SampleID, r.SampleID)
			}
			fmt.Printf("Intersect %s.gMSV2bed.bed because it has not been done yet\n", r.SampleID)
			if err := r.IntersectBed(); err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("Intersect %s.gMSV2bed.bed because it has not been done yet\n", r.SampleID)
		if err := r.IntersectBed(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Read2VecProcess) MakeEncoding() error {
	fmt.Printf("<Convert Intersected %s Bed to Encoding input>\n", r.SampleID)
	if _, err := os.Stat(filepath.Join(r.OutputPath, fmt.Sprintf("%s.Encoding.txt", r.SampleID))); err == nil {
		fmt.Println("Encoding input already made")
		if r.CheckProcess("Encoding.txt") >= 5 {
			fmt.Println("Converting Encoding input is already done")
		} else {
			if r.CheckProcess("intersected.bed") < 5 {
				return fmt.Errorf("%s went wrong, please check %s.intersected.bed", r.SampleID, r.SampleID)
			}
			fmt.Printf("Convert %s.intersected.bed to Encoding input because it does not exist\n", r.SampleID)
			if err := r.BedToEncoding(); err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("Convert %s.intersected.bed to Encoding input because it does not exist\n", r.SampleID)
		if err := r.BedToEncoding(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run script.go <input_path> <sample_id> <reference>")
		os.Exit(1)
	}
	inputPath := os.Args[1]
	sampleID := os.Args[2]
	refPanel := os.Args[3]

	process := NewRead2VecProcess(inputPath, sampleID, refPanel)

	if err := process.MakeDir(); err != nil {
		fmt.Printf("Error creating Encoding directory: %v\n", err)
		os.Exit(1)
	}

	if err := process.MakeBed(); err != nil {
		fmt.Printf("Error making bed: %v\n", err)
		os.Exit(1)
	}

	if err := process.RunBedtools(); err != nil {
		fmt.Printf("Error running BedTools: %v\n", err)
		os.Exit(1)
	}

	if err := process.MakeEncoding(); err != nil {
		fmt.Printf("Error making Encoding: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All processes completed successfully!")
}
