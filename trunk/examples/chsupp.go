package main

import (
	"go-conc.googlecode.com/svn/trunk/goconc"
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"http"
	"io"
	"io/ioutil"
	"json"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
	"url"
)

var waterfallList = []string{
	"http://build.chromium.org/p/chromium.memory/json",
	"http://build.chromium.org/p/chromium.memory.fyi/json",
}

var stepBlackList = map[string]bool{
	"update": true,
	"update_scripts": true,
	"extract_build": true,
	"taskkill": true,
	"svnkill": true,
}

const cacheRoot = "./suppress_cache"

type suppressData struct {
	suppression string
	count int
}

type SuppressionMap map[string]int


func execWithTimeout(t int64, f func() interface{}) (interface{}, bool) {
	c := make(chan interface{})
	go func() {
		c <- f()
	}()
	select {
	case r := <- c:
		return r, true
	case <-time.After(t):
	}
	return nil, false
}




func main() {
	os.Mkdir(cacheRoot, 0755)
	log.SetFlags(log.Ltime|log.Lmicroseconds)
	nRuns := flag.Int("num-runs", 1, "Number of runs")
	flag.Parse()

	var p goconc.Pipeline
	builderStage := p.Stage(5)
	discoverStage := p.Stage(5)
	processStage := p.Stage(20)
	mergeStage := p.Stage(1)
	
	suppressions := make(map[string]int)
	for _, w := range waterfallList {
		w := w
		builderStage.Go(func() {
			log.Printf("builderStage: %s", w)
			for _, b := range getBuilderList(w) {
				path := fmt.Sprintf("%s/builders/%s/builds", w, b)
				discoverStage.Go(func() {
					log.Printf("discoverStage: %s", path)
					for _, l := range getRunLogs(path, *nRuns) {
						logUrl, err := url.Parse(l)
						if err != nil { continue }
						processStage.Go(func() {
							log.Printf("processStage: %s", logUrl)
							supp := processLog(logUrl)
							if supp == nil { return }
							mergeStage.Go(func() {
									for s, n := range supp {
										suppressions[s] += n
									}
							})
						})
					}
				})
			}
		})
	}
	p.Wait()
	output(suppressions)
}

type suppressList []suppressData
func (s suppressList) Len() int { return len(s) }
func (s suppressList) Less(i, j int) (bool) {
	if s[i].count != s[j].count {
		return s[j].count < s[i].count
	}
	return s[j].suppression < s[i].suppression
}
func (s suppressList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func output(supp map[string]int) {
	list := make(suppressList, 0, len(supp))
	for s, n := range supp {
		list = append(list, suppressData{s, n})
	}
	sort.Sort(list)
	for _, s := range list {
		fmt.Printf("%v\t%v\n", s.count, s.suppression) 
	}
}

func determineFilename(u *url.URL) string {
	h := sha1.New()
	h.Write([]byte(u.String()))
	return hex.EncodeToString(h.Sum())
}

func parseBuilderList(data []byte) (builders[] string) {
	var f interface{};
	err := json.Unmarshal(data, &f)
	if err != nil {
		log.Fatal(err)
	}
	builders = make([]string, 0, 100)
	l := f.(map[string]interface{})
	for key, _ := range l {
		builders = append(builders, key)
	}
	return builders
}

func getBuilderList(jsonRoot string) (builders[] string) {
	// Create the URL first
	url, err := url.Parse(jsonRoot + "/builders")
	if err != nil {
		log.Fatal(err)
	}
	r, err := http.Get(url.String())
	if err != nil {
		log.Fatal(err)
	}
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	return parseBuilderList(bodyData)
}

func generateBuilderUrl(builder string, numRuns int) (logUrl string) {
	urlBase := builder
	separator := "?"
	for i := 0; i < numRuns; i++ {
		urlBase = fmt.Sprintf("%s%sselect=-%d", urlBase, separator, i + 1)
		separator = "&"
	}
	u, err := url.Parse(urlBase)
	if err != nil {
		log.Fatal(err)
	}
	return u.String()
}

func parseBuilderOutput(data []byte) (logs []string) {
	var f interface{};
	err := json.Unmarshal(data, &f)
	if err != nil {
		log.Fatal(err)
	}
	// Yarggh, nested JSON parsing a bit gross.
	l := f.(map[string]interface{})
	for _, value := range l {
		x := value.(map[string]interface{})
		for k, b := range x {
			if (k != "steps") {
				continue;
			}
			steps := b.([]interface{})
			for _, stepIter := range steps {
				step := stepIter.(map[string]interface{})
				if stepBlackList[step["name"].(string)] {
					continue;
				}
				lo := step["logs"].([]interface{})
				for _, logIter := range lo {
					// TODO(cbentzel): stdio logs only
					logList := logIter.([]interface {})
					logName := logList[1].(string)
					if !strings.HasSuffix(logName, "/stdio") {
						continue
					}
					logs = append(logs, logName)
				}
			}
		}
	}
	return
}

func getRunLogs(builder string, numRuns int) (runLogs []string) {
	builderUrl := generateBuilderUrl(builder, numRuns)
	//log.Printf("Trying to get %v", builderUrl)
	r, err := http.Get(builderUrl)
	if err != nil {
		log.Printf("oh noes")
		log.Fatal(err)
	}
	defer r.Body.Close()
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("whoopsie do")
		log.Fatal(err)
	}
	//log.Printf("Got %v", builderUrl)
	return parseBuilderOutput(bodyData)
}

func cacheLogOutput(logUrl *url.URL) (cacheFilePath string, ok bool) {
	cacheFilePath = path.Join(cacheRoot, determineFilename(logUrl))
	// See if it already exists.
	_, err := os.Stat(cacheFilePath)
	if err == nil { return cacheFilePath, true }
	// Create a cached file.
	tempFile, err := os.Create(cacheFilePath + "-tmp")
	if err != nil {
		log.Printf("Failed to generate temp filename: %s", err)
		return;
	} 
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()
	// Do a URL request, and pipe the data into the temporary file.
	r, err := http.Get(logUrl.String())
	if err != nil {
		log.Printf("Failed to http.Get: %s", err)
		return;
	} 
	defer r.Body.Close()
	_, err = io.Copy(tempFile, r.Body)
	if err != nil {
		log.Printf("Failed to io.Copy HTTP: %s", err)
		return;
	} 
	// Move the file to it's final location.
	tempFile.Close()
	err = os.Rename(tempFile.Name(), cacheFilePath)
	if err != nil {
		log.Printf("Failed to rename temp file: %s", err)
		return;
	} 
	// Pipe the data through
	return cacheFilePath, true
}

func parseLogOutput(logOutput io.Reader) SuppressionMap {
	bufferedReader := bufio.NewReader(logOutput)
	suppressions := make(SuppressionMap)
	withinSuppressions := false
	for {
		line, isPrefix, err := bufferedReader.ReadLine()
		if err == os.EOF { break }
		if err != nil { return nil }
		if isPrefix { continue }
		trimmed := strings.Replace(string(line), "</span><span class=\"stdout\">", "", -1)
		trimmed = strings.TrimSpace(trimmed)
		if withinSuppressions {
			if trimmed[:5] == "-----" {
				withinSuppressions = false
				continue
			}
			// Split the string
			splitStrings := strings.SplitN(trimmed, " ", 2)
			if (len(splitStrings) != 2) {
				log.Print("Its not two split strings")
				continue
			}
			count, err := strconv.Atoi(splitStrings[0])
			if (err != nil) { 
				log.Printf("Invalid input %v %v", splitStrings[0], string(line))
				continue
			}
			suppressName := splitStrings[1]
			existingCount := suppressions[suppressName]
			suppressions[suppressName] = existingCount + count
		} else {
			withinSuppressions = trimmed == "count name"
		}
	}
	return suppressions
}

func parseLogOutputFromFile(cacheFilePath string) SuppressionMap {
	// Open file, parse it, close the file
	logOutputFile, err := os.Open(cacheFilePath)
	if err != nil { return nil }
	defer logOutputFile.Close()
	return parseLogOutput(logOutputFile)
}

func processLog(logUrl *url.URL) SuppressionMap {
	r, _ := execWithTimeout(30*1e9, func() interface{} {
		cacheFilePath, ok := cacheLogOutput(logUrl)
		if !ok { return nil }
		return parseLogOutputFromFile(cacheFilePath)
	})
	if r == nil { return nil }
	return r.(SuppressionMap)
}
 
