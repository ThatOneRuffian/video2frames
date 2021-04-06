package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
)

// ============= User Input Variables ===============
var inputFile string
var destinationDirectory string
var logOutputDest string
var outputSize string
var outputSuffix string
var outputPrefix string
var conversionFactor int
var compressOutput bool
var grayScale bool

func main() {
	flag.StringVar(&destinationDirectory, "o", ".", "Specify output directory.")
	flag.StringVar(&inputFile, "i", "", "Specify input video file.")
	flag.StringVar(&outputSize, "s", "", "Specify output image width. (e.g. 600x800)")
	flag.StringVar(&logOutputDest, "l", "", "Log file output destination.")
	flag.StringVar(&outputSuffix, "suffix", "", "Add suffix to the output file.")
	flag.StringVar(&outputPrefix, "prefix", "", "Add prefix to the output file.")
	flag.IntVar(&conversionFactor, "x", 100, "Out of every 100 frames convert X frames.")
	flag.BoolVar(&grayScale, "g", false, "Convert output to grayscale.")
	flag.BoolVar(&compressOutput, "c", false, "Compress output into PNG format. Default uncompressed BMP.")
	flag.Parse()

	checkParameters()
	startConversion()
	writeExifData()
}

func writeExifData() {
	exifTool, err := exiftool.NewExiftool()
	// snap shot before and after files and loop through the new files
	// write meta data from json file template
	if err != nil {
		appendToLog("Unable to create exiftool. Skiping writing of metadata.")
	}
	fmt.Println(exifTool)
}

func checkParameters() {
	dirHandler(&destinationDirectory)
	if len(logOutputDest) > 0 {
		dirHandler(&logOutputDest)
	}
	checkInputFile(inputFile)
	if conversionFactor > 100 || conversionFactor < 1 {
		panic(appendToLog("Conversion factor must be within range 1-100%."))
	}
	if len(outputSize) > 0 {
		if strings.Contains(outputSize, "x") || strings.Contains(outputSize, "X") {

		} else {
			panic(appendToLog("Size argument must be provided in the following format: WxH"))
		}
	}
}

//check sourceFile
func checkInputFile(sourceFile string) {

	if len(sourceFile) > 0 {
		sourceFileError := fmt.Sprintf("Could not read the source file: %s ", sourceFile)
		_, readError := os.Stat(sourceFile)

		if os.IsNotExist(readError) {
			panic(appendToLog(sourceFileError))
		} else if readError != nil {
			// catch all other file errors and log
			panic(appendToLog(sourceFileError))
		}
	} else {
		panic(appendToLog("Source file not provided (Use: -i source.mp4)"))
	}
}

func startConversion() {
	fileFormat := ".bmp" //uncompressed bitmap default
	if compressOutput {
		fileFormat = ".png"
	}

	conversionFactorFloat := float64(conversionFactor)
	conversionFactorFloat = conversionFactorFloat / 100.0
	conversionFactorString := fmt.Sprint(conversionFactorFloat)
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		panic(appendToLog("Could not find ffmpeg"))
	}
	programArgs := []string{ffmpegPath, "-r", "1", "-i", inputFile, "-r", conversionFactorString}
	fileAbsPath := destinationDirectory + outputPrefix + "frame%03d" + outputSuffix + fileFormat

	if grayScale {
		programArgs = append(programArgs, "-vf")
		programArgs = append(programArgs, "format=gray")
	}
	if len(outputSize) > 0 {
		programArgs = append(programArgs, "-s")
		programArgs = append(programArgs, outputSize)
	}

	programArgs = append(programArgs, fileAbsPath)

	cmd := &exec.Cmd{
		Path:   ffmpegPath,
		Args:   programArgs,
		Stdout: os.Stdout,
	}

	var buffer bytes.Buffer
	cmd.Stderr = &buffer //ffmpeg outputs on standard error
	fmt.Println("Generating frames...")
	if cmd.Run() != nil {
		panic(appendToLog("could not generate frames"))
	}
	appendToLog(buffer.String()) //write ffmpeg output to log
	fmt.Println("Finished generating frames.")
}

//check targetDir and create if non-exist
func dirHandler(targetDir *string) {
	//append "/" if missing from end of provided dir
	const forwardSlash byte = 92
	const backSlash byte = 47
	pathByteArray := []byte(*targetDir)
	lastByteChar := pathByteArray[len(pathByteArray)-1]

	if lastByteChar != forwardSlash && lastByteChar != backSlash {
		*targetDir = *targetDir + "/"
	}

	// create dir if non-existant
	dirError := fmt.Sprintf("Could not create the following dir: %s ", targetDir)
	_, readError := os.ReadDir(*targetDir)

	if os.IsNotExist(readError) {
		makeError := os.MkdirAll(*targetDir, 0777)
		if makeError != nil {
			panic(appendToLog(dirError))
		}
	} else if readError != nil { // catch all other dir errors and log
		panic(appendToLog(dirError))
	}
}

//append string to log
func appendToLog(logEntry string) string {
	if len(logOutputDest) > 0 {
		rawData := time.Now().String() + ": " + logEntry + string('\n')

		dataToWrite := []byte(rawData)

		logFile, err := os.OpenFile(logOutputDest+"log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		//check file open errors
		if err != nil {
			log.Fatal(err)
		}
		//attempt write to log
		if _, err := logFile.Write(dataToWrite); err != nil {
			log.Fatal(err)
		}
		//attempt to close log
		if err := logFile.Close(); err != nil {
			log.Fatal(err)
		}
	}
	return logEntry
}
