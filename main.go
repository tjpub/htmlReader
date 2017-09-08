package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type meret struct {
	name    string
	width   float32
	height  float32
	xrepeat int64
	yrepeat int64
}

type meretek []meret

func (s meretek) Len() int {
	return len(s)
}
func (s meretek) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s meretek) Less(i, j int) bool {
	var si = s[i].name
	var sj = s[j].name
	var siLower = strings.ToLower(si)
	var sjLower = strings.ToLower(sj)
	if siLower == sjLower {
		return si < sj
	}
	return siLower < sjLower
}

func main() {
	wd, _ := os.Getwd()
	argsWithoutProg := os.Args[1:]
	path := wd
	if len(argsWithoutProg) > 0 {
		path = wd + "\\" + argsWithoutProg[0]
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Printf("Könyvtár nem létezik! %s", path)
		os.Exit(0)
	}

	//ha van skip file, feldolgozzuk
	skipFile := wd + "\\skip.txt"
	var skip []string
	if _, err := os.Stat(skipFile); err == nil {
		fmt.Printf(" skip fájl feldolgozása %s -- \n", skipFile)
		skip = processSkips(skipFile)

		fmt.Println("skip: -->", skip)
	}

	files, err := ioutil.ReadDir(path)
	check(err)

	fmt.Printf(" -- listing htm files from %s -- \n", path)

	var fileNames []string
	for _, f := range files {
		fileName := f.Name()
		if strings.HasSuffix(fileName, ".htm") {
			fileNames = append(fileNames, path+"\\"+fileName)
		}
	}

	fmt.Println("htm files: -->", fileNames)

	var datas meretek

	for _, file := range fileNames {
		fmt.Println("reading htm --> ", file)
		dat, err := ioutil.ReadFile(file)
		check(err)
		m := processHTMFile(file, string(dat))
		datas = append(datas, m...)
	}

	fmt.Printf(" -- listing html files from %s -- \n", path)

	fileNames = nil
	for _, f := range files {
		fileName := f.Name()
		if strings.HasSuffix(fileName, ".html") {
			fileNames = append(fileNames, path+"\\"+fileName)
		}
	}

	fmt.Println("html files: -->", fileNames)
	for _, file := range fileNames {
		fmt.Println("reading html --> ", file)
		dat, err := ioutil.ReadFile(file)
		check(err)
		m := processHTMLFile(file, string(dat))
		datas = append(datas, m...)
	}

	fmt.Println("structs ", datas)

	if len(datas) == 0 {
		fmt.Println("no data! ")
		return
	}

	sort.Sort(datas)

	check(err)
	writeToFile(datas, path, skip)
}

func processSkips(path string) []string {
	var skips []string
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		skips = append(skips, scanner.Text())
	}

	return skips
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func writeToFile(datas meretek, dir string, skip []string) {
	f, err := os.Create(dir + "\\__file__")
	check(err)
	defer f.Close()

	w := bufio.NewWriter(f)
	//w.WriteString("Könyvtár: ")
	check(err)
	//w.WriteString(dir)
	//w.WriteString("\n\n")
	for _, data := range datas {
		if !stringInSlice(data.name, skip) {
			w.WriteString(data.name)
			w.WriteString("\t")
			w.WriteString(strings.Replace(fmt.Sprintf("%.3f", data.width), ".", ",", -1))
			w.WriteString("\t")
			w.WriteString(strings.Replace(fmt.Sprintf("%.3f", data.height), ".", ",", -1))

			w.WriteString("\t")
			w.WriteString(fmt.Sprintf("%d", data.xrepeat))
			w.WriteString("\t")
			w.WriteString(fmt.Sprintf("%d", data.yrepeat))

			w.WriteString("\n")
		}
	}

	w.Flush()
}

func processHTMLFile(fileName string, content string) meretek {
	fmt.Println("processing html--> ", fileName)

	var datas meretek

	f := strings.NewReader(content)
	z := html.NewTokenizer(f)
	joTable := false
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			fmt.Println("processing done --> ", fileName)
			fmt.Printf("file has %d elements ", len(datas))
			fmt.Println("")
			return datas

		case tt == html.EndTagToken:
			continue
		case tt == html.StartTagToken:
			t := z.Token()
			if !joTable {
				isTable := t.Data == "table"
				if !isTable {
					continue
				}
				for _, a := range t.Attr {
					if a.Val == "1" {
						//<TABLE border="1">
						joTable = true
						break
					}
				}

				if !joTable {
					continue
				}
				joTable = true
			}

			for {
				z.Next()
				t = z.Token()
				isTd := t.Data == "td"
				if isTd {
					break
				}
			}

			//<TR><TD>Feba_takarofolia_kt02_Black_1.lenl2t.tif<TD>79530.195<TD>Black<TD>355.018<TD>224.017<TD>Advanced01<TD>No<TD>No
			z.Next()
			t = z.Token()

			data := meret{}
			data.name = t.Data
			data.xrepeat = 1
			data.yrepeat = 1
			fmt.Println("Name ", t.Data)

			z.Next()
			z.Next()
			//area kihagyása

			z.Next()
			z.Next()
			//separation kihagyása
			z.Next()
			z.Next()
			t = z.Token()
			value, err := strconv.ParseFloat(t.Data, 32)
			check(err)
			data.width = float32(value)
			fmt.Println("dim x: ", t.Data)
			z.Next()
			z.Next()
			t = z.Token()
			value, err = strconv.ParseFloat(t.Data, 32)
			check(err)
			data.height = float32(value)
			fmt.Println("dim y: ", t.Data)

			datas = append(datas, data)

			//utolsó 3 kihagyása
			z.Next()
			z.Next()
			z.Next()
			z.Next()
			z.Next()
			z.Next()
		default:
		}
	}
}

func processHTMFile(fileName string, content string) meretek {
	fmt.Println("processing --> ", fileName)

	var datas meretek

	f := strings.NewReader(content)
	z := html.NewTokenizer(f)
	imagesFound := false
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			fmt.Println("processing done --> ", fileName)
			fmt.Printf("file has %d elements ", len(datas))
			fmt.Println("")
			return datas

		case tt == html.EndTagToken:
			if imagesFound {
				t := z.Token()

				isTable := t.Data == "table"
				if isTable {
					imagesFound = false
				}
			}

			continue
		case tt == html.StartTagToken:
			if !imagesFound {
				continue
			}

			//előszedjük belőle a tr-eket.
			t := z.Token()
			isTr := t.Data == "tr"
			if !isTr {
				continue
			}

			z.Next()
			t = z.Token()
			//megyünk amíg td-t nem találunk
			if t.Data != "td" {
				continue
			}
			z.Next()
			t = z.Token()

			data := meret{}
			data.name = t.Data
			fmt.Println("Name ", t.Data)

			z.Next()
			z.Next()
			z.Next()
			z.Next()
			t = z.Token()
			value, err := strconv.ParseFloat(t.Data, 32)
			check(err)
			data.width = float32(value)
			fmt.Println("Width ", t.Data)

			z.Next()
			z.Next()
			z.Next()
			z.Next()
			t = z.Token()
			value, err = strconv.ParseFloat(t.Data, 32)
			check(err)
			data.height = float32(value)
			fmt.Println("Height ", t.Data)
			i := 0
			for i < 3 {
				z.Next()
				t = z.Token()
				d := strings.TrimSpace(t.Data)
				//fmt.Println("got", d)
				if d == "td" || d == "" {
					//fmt.Println("rossz t ", d)
				} else {
					i++
					//fmt.Println("jó t ", d)
				}
			}

			val, err := strconv.ParseInt(t.Data, 10, 32)
			check(err)
			data.xrepeat = val
			fmt.Println("xrepeat  ", t.Data)

			i = 0
			for i < 1 {
				z.Next()
				t = z.Token()
				d := strings.TrimSpace(t.Data)
				if d == "td" || d == "" {
				} else {
					i++
				}
			}
			val, err = strconv.ParseInt(t.Data, 10, 32)
			check(err)
			data.yrepeat = val
			fmt.Println("yrepeat  ", t.Data)

			fmt.Println("")

			datas = append(datas, data)
		default:
			t := z.Token()
			if t.Data == "Images:" {
				imagesFound = true
			}
		}
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
