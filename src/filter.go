package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type opMethod func(int, int) bool

var usefilter = false

var ops = map[string]opMethod{}

var vars = map[string]*int{}

var filter = map[int][]string{}

func gt(x int, y int) bool {
	return x > y
}

func lt(x int, y int) bool {
	return x < y
}

func eq(x int, y int) bool {
	return x == y
}

func neq(x int, y int) bool {
	return x != y
}

func initFilters(filt string, stats *Stats) {
	if filt != "" {
		usefilter = true
		ops[">"] = opMethod(gt)
		ops["<"] = opMethod(lt)
		ops["=="] = opMethod(eq)
		ops["!="] = opMethod(neq)
		vars["chars"] = &stats.chars
		vars["code"] = &stats.code
		vars["length"] = &stats.length
		vars["lines"] = &stats.lines
		vars["words"] = &stats.words
		vars["tags"] = &stats.tags
		var tmp_filter_index = 0
		groups := RegSplit(filt, "(&&)")
		for _, group := range groups {
			r, _ := regexp.Compile("(?P<line>[a-z]+)(\\s{0,100})(?P<op>(<|>|\\|\\||==|!=))(\\s{0,100})(?P<data>[-\\d]+)")
			check := r.FindStringSubmatch(group)
			if _, ok := vars[check[variable]]; ok {
				filter[tmp_filter_index] = append(filter[tmp_filter_index], check[variable])
				filter[tmp_filter_index] = append(filter[tmp_filter_index], check[op])
				filter[tmp_filter_index] = append(filter[tmp_filter_index], check[data])
				tmp_filter_index++
			} else {
				fmt.Printf("Error: Filter error, parameter '%s' does not exists.\n", check[variable])
				os.Exit(-1)
			}
		}
	}
}

func RegSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:len(text)]
	return result
}

func checkFilter() bool {
	if !usefilter {
		return true
	} else {
		flag := true
		for _, val := range filter {
			i, _ := strconv.Atoi(val[2])
			flag = ops[val[1]](*vars[val[0]], i) && flag
		}
		return flag
	}
}
