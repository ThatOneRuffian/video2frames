package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ============= User Input Variables ===============
var inputFile string
var outputSize string
var outputSuffix string
var outputPrefix string
var logOutputDest string
var fileToExifDump string
var exifDataSource string
var programWorkingDir string
var destinationDirectory string
var conversionFactor int
var qualityFactor int
var exifGenerateTemplate bool
var compressOutput bool
var grayScale bool

type ExifData struct {
	Make                    string
	Model                   string
	Focallength             string
	Focallengthin35mmformat string
}

func main() {
	flag.StringVar(&destinationDirectory, "o", ".", "Specify output directory.")
	flag.StringVar(&inputFile, "i", "", "Specify input video file.")
	flag.StringVar(&outputSize, "s", "", "Specify output image resolution. (e.g. 600x800)")
	flag.StringVar(&logOutputDest, "l", "", "Log file output destination.")
	flag.StringVar(&fileToExifDump, "d", "", "Dump the exif data of the provided file.")
	flag.StringVar(&exifDataSource, "exif-data", "", "Provide a JSON file for writing key exif data.")
	flag.StringVar(&outputSuffix, "suffix", "", "Add suffix to the output file.")
	flag.StringVar(&outputPrefix, "prefix", "", "Add prefix to the output file.")
	flag.IntVar(&conversionFactor, "x", 100, "Out of every 100 frames convert X frames.")
	flag.IntVar(&qualityFactor, "q", 1, "Set the quality of the export 1-31. Lower is better quality.")
	flag.BoolVar(&grayScale, "g", false, "Convert output to grayscale.")
	flag.BoolVar(&compressOutput, "c", false, "Compress output into JPEG format. Default uncompressed BMP.")
	flag.BoolVar(&exifGenerateTemplate, "export-exif-template", false, "Generate JSON template file. For use with supported exif data writing (e.g. --exif-data).")
	flag.Parse()

	checkParameters()

	if len(fileToExifDump) == 0 && !exifGenerateTemplate && len(inputFile) > 0 {
		// start video conversion to frames
		startConversion()
	} else if len(fileToExifDump) > 0 {
		// dump target exif data
		dumpExifData(fileToExifDump)
	} else if exifGenerateTemplate {
		//generate exif data template
		exportJSONtemplate()
	}
	if len(exifDataSource) > 0 {
		//write exif data
		writeExifData()
	}
	os.Exit(0)
}

func exportJSONtemplate() {
	//export supported exif data tags
	exifDataTemplate := ExifData{"desired_camera_make", "desired_camera_model", "desired_focallength", "desired_focallengthin35mmformat"}
	dataToWrite, encodeErr := json.Marshal(exifDataTemplate)
	if encodeErr != nil {
		fmt.Println(appendToLog("Unable to encode template data"))
		os.Exit(1)
	}
	writeData(destinationDirectory+"exif_data.JSON", string(dataToWrite), true)
	os.Exit(0)
}

func loadJSONexif() ExifData {
	jsonFile, err := os.Open(exifDataSource)
	var jsonData ExifData
	if err != nil {
		fmt.Println(appendToLog("Could not open JSON file."))
		os.Exit(1)
	}

	fileScan := bufio.NewScanner(jsonFile)

	for fileScan.Scan() {
		inputRow := fileScan.Text()
		decodeErr := json.Unmarshal([]byte(inputRow), &jsonData)

		if decodeErr != nil {
			fmt.Println(appendToLog("Error decoding JSON file."))
			os.Exit(1)
		}

	}
	return jsonData
}

func writeExifData() {
	stdOutLoc := os.Stdout
	customTags := loadJSONexif()
	//Check PATH for exiftool
	exifToolPath, err := exec.LookPath("exiftool")
	if err != nil {
		// if not found in path then search current dir for exiftool

		localDmsgget := fmt.Sprintf(programWorkingDir + "/exiftool")

		exifToolPath, err = exec.LookPath(localDmsgget)
		if err != nil {
			exifError := fmt.Sprintf("Unable to find exiftool in PATH or current dir: %s", localDmsgget)
			fmt.Println(appendToLog(exifError))
			os.Exit(1)
		}
	}

	exifToolTags := []string{exifToolPath, "-overwrite_original"}
	if len(customTags.Make) > 0 {
		exifToolTags = append(exifToolTags, "-make="+customTags.Make)
	}

	if len(customTags.Model) > 0 {
		exifToolTags = append(exifToolTags, "-model="+customTags.Model)
	}

	if len(customTags.Focallength) > 0 {
		exifToolTags = append(exifToolTags, "-FocalLength="+customTags.Focallength)
	}

	if len(customTags.Focallengthin35mmformat) > 0 {
		exifToolTags = append(exifToolTags, "-focallengthin35mmformat="+customTags.Focallengthin35mmformat)
	}

	exifToolTags = append(exifToolTags, destinationDirectory)
	exifToolCmd := &exec.Cmd{
		Path:   exifToolPath,
		Args:   exifToolTags,
		Stdout: stdOutLoc,
		Stderr: os.Stderr,
	}
	fmt.Println("Writing exif data...")
	if err := exifToolCmd.Run(); err != nil {
		fmt.Printf("Error writing exif data:\n", err)
		os.Exit(1)
	}
}

func dumpExifData(filePath string) {
	exifToolCmd := exec.Command("exiftool", filePath)
	exifToolCmd.Stderr = os.Stderr
	fmt.Println("Dumping meta data now: ", filePath)
	stdOut, err := exifToolCmd.StdoutPipe()
	if nil != err {
		fmt.Println(appendToLog("Error attaching to exiftool stdout:"), err.Error())
		os.Exit(1)
	}
	stdOutReader := bufio.NewReader(stdOut)
	go func(stdOutReader io.Reader) {
		scanner := bufio.NewScanner(stdOutReader)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}(stdOutReader)

	if err := exifToolCmd.Start(); nil != err {
		fmt.Println(fmt.Sprintf("Error starting program: %s, %s", exifToolCmd.Path, err.Error()))
		fmt.Println("Make sure that exiftool is installed on your system")
		os.Exit(1)
	}
	exifToolCmd.Wait()
}

func checkParameters() {
	dirHandler(&destinationDirectory)
	if len(logOutputDest) > 0 {
		dirHandler(&logOutputDest)
	}
	if len(fileToExifDump) == 0 && !exifGenerateTemplate {
		checkInputFile(inputFile)
	}

	if conversionFactor > 100 || conversionFactor < 1 {
		fmt.Println(appendToLog("Conversion factor must be within range 1-100%."))
		os.Exit(1)
	}
	if len(outputSize) > 0 {
		if strings.Contains(outputSize, "x") || strings.Contains(outputSize, "X") {

		} else {
			fmt.Println(appendToLog("Size argument must be provided in the following format: WxH"))
			os.Exit(1)
		}
	}
	if qualityFactor < 1 {
		qualityFactor = 1
	} else if qualityFactor > 31 {
		qualityFactor = 31
	}
	_programWorkingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(appendToLog("Error obtaining program's working dir, using relative ./"))
		programWorkingDir = "./"
	} else {
		programWorkingDir = _programWorkingDir
	}
}

// check sourceFile
func checkInputFile(sourceFile string) {

	if len(sourceFile) > 0 {
		sourceFileError := fmt.Sprintf("Could not read the source file: %s ", sourceFile)
		_, readError := os.Stat(sourceFile)

		if os.IsNotExist(readError) {
			fmt.Println(appendToLog(sourceFileError))
			os.Exit(1)
		} else if readError != nil {
			// catch all other file errors and log
			fmt.Println(appendToLog(sourceFileError))
			os.Exit(1)
		}
	} else if !(len(exifDataSource) > 0) {
		fmt.Println(appendToLog("Source file not provided (Use: -i source.mp4)"))
		os.Exit(1)
	}
}

func startConversion() {
	fileFormat := ".bmp" //uncompressed bitmap default
	if compressOutput {
		fileFormat = ".jpg"
	}

	conversionFactorFloat := float64(conversionFactor)
	conversionFactorFloat = conversionFactorFloat / 100.0
	conversionFactorString := fmt.Sprint(conversionFactorFloat)
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println(appendToLog("Could not find ffmpeg. Make sure it can be found in system's PATH"))
		os.Exit(1)
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

	programArgs = append(programArgs, "-qscale:v")
	programArgs = append(programArgs, fmt.Sprint(qualityFactor))
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
		fmt.Println(appendToLog("Could not generate frames"))
		os.Exit(1)
	}
	appendToLog(buffer.String()) //write ffmpeg output to log
	fmt.Println("Finished generating frames.")
}

// check targetDir and create if non-exist
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
			fmt.Println(appendToLog(dirError))
			os.Exit(1)
		}
	} else if readError != nil { // catch all other dir errors and log
		fmt.Println(appendToLog(dirError))
		os.Exit(1)
	}
}

// append string to log
func appendToLog(logEntry string) string {
	if len(logOutputDest) > 0 {
		dataToWrite := time.Now().String() + ": " + logEntry + string('\n')

		writeData(logOutputDest+"log.txt", dataToWrite, false)
	}
	return logEntry
}

func writeData(filePath string, rawData string, overWrite bool) {

	if overWrite {
		err := os.Remove(filePath)
		if err != nil {
			appendToLog("Unable to remove file.")
		}
	}
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//check file open errors
	if err != nil {
		log.Fatal(err)
	}
	//attempt file write
	if _, err := logFile.Write([]byte(rawData)); err != nil {
		log.Fatal(err)
	}
	//attempt file close
	if err := logFile.Close(); err != nil {
		log.Fatal(err)
	}
}
