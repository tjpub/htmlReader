package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type meret struct {
	name   string
	width  float32
	height float32
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
	fmt.Println(" -- listing html files -- ")

	files, err := ioutil.ReadDir("./")
	check(err)

	var fileNames []string
	for _, f := range files {
		fileName := f.Name()
		if strings.HasSuffix(fileName, ".htm") {
			fileNames = append(fileNames, fileName)
		}
	}

	fmt.Println("-->", fileNames)

	var datas meretek

	for _, file := range fileNames {
		fmt.Println("reading html --> ", file)
		dat, err := ioutil.ReadFile(file)
		check(err)
		m := processFile(file, string(dat))
		datas = append(datas, m...)
	}

	fmt.Println("structs ", datas)

	sort.Sort(datas)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)
	writeToFile(datas, dir)
}

func writeToFile(datas meretek, dir string) {
	f, err := os.Create("__file__")
	check(err)
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("Könyvtár: ")
	check(err)
	w.WriteString(dir)
	w.WriteString("\n\n")
	for _, data := range datas {
		w.WriteString(data.name)
		w.WriteString("\t")
		w.WriteString(fmt.Sprintf("%.3f", data.width))
		w.WriteString("\t")
		w.WriteString(fmt.Sprintf("%.3f", data.height))
		_, err = w.WriteString("\n")
		check(err)
	}

	w.Flush()
}

func processFile(fileName string, content string) meretek {
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

func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}
