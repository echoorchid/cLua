package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/milochristiansen/lua/ast"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FileData struct {
	file string
	path string
	line map[int]uint64
}

func main() {

	input := flag.String("i", "", "input cov file")
	root := flag.String("path", "./", "source code path")
	filter := flag.String("f", "", "filter filename")
	showcode := flag.Bool("showcode", true, "show code")
	showtotal := flag.Bool("showtotal", true, "show total")
	showfunc := flag.Bool("showfunc", true, "show func")

	flag.Parse()

	if len(*input) == 0 {
		flag.Usage()
		return
	}

	filedata, ok := parse(*input, *root)
	if !ok {
		return
	}

	if len(*filter) != 0 {
		for _, p := range filedata {
			if p.file == *filter {
				calc(p, *showcode, *showtotal, *showfunc)
			}
		}
	} else {
		for _, p := range filedata {
			calc(p, *showcode, *showtotal, *showfunc)
		}
	}
}

func parse(filename string, root string) ([]FileData, bool) {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("ReadFile fail %v\n", err)
		return nil, false
	}

	var filedata []FileData
	n := 0
	i := 0
	for {
		if i+4 > len(data) {
			break
		}
		strlen := binary.LittleEndian.Uint32(data[i : i+4])
		i += 4

		if i+int(strlen) > len(data) {
			break
		}
		str := string(data[i : i+int(strlen)])
		i += int(strlen)
		if i >= len(data) {
			break
		}

		if i+8 > len(data) {
			break
		}
		count := binary.LittleEndian.Uint64(data[i : i+8])
		i += 8

		str = strings.TrimLeft(str, "@")
		if strings.Count(str, ":") != 1 {
			continue
		}
		params := strings.Split(str, ":")
		if len(params) < 2 {
			fmt.Printf("Split fail %s\n", str)
			return nil, false
		}
		filename := params[0]
		line, err := strconv.Atoi(params[1])
		if err != nil {
			fmt.Printf("Atoi fail  %s %v\n", str, err)
			return nil, false
		}

		path, err := filepath.Abs(root + "/" + filename)
		if err != nil {
			fmt.Printf("Path fail %s %s %v\n", root, str, err)
			return nil, false
		}

		if !fileExists(path) {
			fmt.Printf("File not found %s\n", path)
			return nil, false
		}

		file := filepath.Base(path)
		file = strings.TrimSuffix(file, filepath.Ext(file))

		find := false
		for index, _ := range filedata {
			if filedata[index].path == path {
				filedata[index].line[line] += count
				find = true
				break
			}
		}

		if !find {
			f := FileData{file, path, make(map[int]uint64)}
			f.line[line] = count
			filedata = append(filedata, f)
		}

		n++
	}

	fmt.Printf("total points = %d, files = %d\n", n, len(filedata))

	return filedata, true
}

func readfile(filename string) ([]string, bool) {

	var filecontent []string

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Open File Fail %v\n", err)
		return filecontent, false
	}
	defer file.Close()

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	for {
		str, err := reader.ReadString('\n')
		filecontent = append(filecontent, str)
		if err != nil {
			break
		}
	}

	return filecontent, true
}

type luaVisitor struct {
	f func(n ast.Node)
}

func (lv *luaVisitor) Visit(n ast.Node) ast.Visitor {
	lv.f(n)
	return lv
}

func calc(f FileData, showcode bool, showtotal bool, showfunc bool) {

	fmt.Printf("coverage of %s:\n", f.path)

	filecontent, ok := readfile(f.path)
	if !ok {
		return
	}

	block, ok := parseLua(f.path)
	if !ok {
		return
	}

	validline := make(map[int]int)
	v := luaVisitor{f: func(n ast.Node) {
		if n != nil {
			validline[n.Line()]++
		}
	}}
	for _, stmt := range block {
		ast.Walk(&v, stmt)
	}

	if showcode {
		maxpre := uint64(0)
		for _, c := range f.line {
			if c > maxpre {
				maxpre = c
			}
		}
		pre := 0
		for maxpre > 0 {
			maxpre /= 10
			pre++
		}

		for index, str := range filecontent {
			val, ok := f.line[index+1]
			if ok {
				fmt.Printf(fmt.Sprintf("%%-%d", pre)+"v", val)
			} else {
				fmt.Printf(fmt.Sprintf("%%-%d", pre)+"v", " ")
			}
			fmt.Printf(" %s\n", strings.TrimRight(str, "\n"))
		}
	}

	if showtotal {
		valid := 0
		for index, _ := range filecontent {
			_, ok := f.line[index+1]
			if ok {
				_, ok = validline[index+1]
				if ok {
					valid++
				}
			}
		}
		fmt.Printf("%s total coverage %d%% %d/%d\n", f.path, valid*100/len(validline), valid, len(validline))
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func parseLua(filename string) ([]ast.Stmt, bool) {

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Open File Fail %v\n", err)
		return nil, false
	}
	defer file.Close()

	source, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("ReadAll File Fail %v\n", err)
		return nil, false
	}

	block, err := ast.Parse(string(source), 1)
	if err != nil {
		fmt.Printf("Parse File Fail %v\n", err)
		return nil, false
	}

	return block, true
}
