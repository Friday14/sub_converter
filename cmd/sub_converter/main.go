package main

import (
    "flag"
    "fmt"
    "golang.org/x/text/language"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"
    "sub_converter/subtitles"
)

var (
    scanDir = flag.String("dir", ".", "Directory for search sub")
)

func main() {
    flag.Parse()
    registerAppCredentials()
    root, _ := filepath.Abs("./")
    _ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", root+"\\"+"google.env.json")
    files, err := ioutil.ReadDir(*scanDir)
    checkErr(err)
    fullPathDir, _ := filepath.Abs(*scanDir)

    // Filter by format .srt
    for _, file := range files {

        if filepath.Ext(file.Name()) != ".srt" {
            continue
        }

        filename := file.Name()
        ext := filepath.Ext(filename)

        // Check for translation
        _, err := os.Stat(filename[0:len(filename)-len(ext)] + "_ru" + ext)

        if strings.Contains(filename, "_ru") || !os.IsNotExist(err) {
            fmt.Println("Skipped " + filename)
            continue
        }

        fmt.Println("Start for " + filename)
        sub, err := subtitles.CreateSub(fullPathDir + "\\" + filename)

        if err == nil {
            newSub, err := sub.Translate(language.Russian)
            checkErr(err)
            err = newSub.SaveToFile()

            if err != nil {
                fmt.Println("File " + newSub.Filename() + " completed")
            }
        }
    }
    fmt.Println("Finished")
}

func registerAppCredentials() {
    root, _ := filepath.Abs("./")
    if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
        _ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", root+"\\"+"google.env.json")
    }
}

func checkErr(err error) {
    if err != nil {
        log.Fatalln(err.Error())
    }
}
