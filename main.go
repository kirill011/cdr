package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

var CdrLineNodeRoot *CdrLineNode = new(CdrLineNode)

type CdrLine struct {
	CallType       string
	PhoneNumber    string
	DateNachString string
	DateNach       time.Time
	DateConString  string
	DateCon        time.Time
	Tariff         string
}

type CdrLineNode struct {
	line *CdrLine
	next *CdrLineNode
}

// Функция для заполнения ноды
func (node *CdrLineNode) addNode(line *CdrLine) {
	node.line = line
	node.next = new(CdrLineNode)
}

// Функция переводит время начала и конца звонка из строки в time.Time
func (c *CdrLine) strToTime() {
	year := 0
	month := 0
	day := 0
	hour := 0
	minute := 0
	second := 0
	buf := ""
	for i, v := range c.DateNachString {
		switch i {
		case 3:
			buf += string(v)
			year, _ = strconv.Atoi(buf)
			buf = ""
		case 5:
			buf += string(v)
			month, _ = strconv.Atoi(buf)
			buf = ""
		case 7:
			buf += string(v)
			day, _ = strconv.Atoi(buf)
			buf = ""
		case 9:
			buf += string(v)
			hour, _ = strconv.Atoi(buf)
			buf = ""
		case 11:
			buf += string(v)
			minute, _ = strconv.Atoi(buf)
			buf = ""
		case 13:
			buf += string(v)
			second, _ = strconv.Atoi(buf)
			buf = ""
		default:
			buf += string(v)
		}
	}
	c.DateNach = time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
	for i, v := range c.DateConString {
		switch i {
		case 3:
			buf += string(v)
			year, _ = strconv.Atoi(buf)
			buf = ""
		case 5:
			buf += string(v)
			month, _ = strconv.Atoi(buf)
			buf = ""
		case 7:
			buf += string(v)
			day, _ = strconv.Atoi(buf)
			buf = ""
		case 9:
			buf += string(v)
			hour, _ = strconv.Atoi(buf)
			buf = ""
		case 11:
			buf += string(v)
			minute, _ = strconv.Atoi(buf)
			buf = ""
		case 13:
			buf += string(v)
			second, _ = strconv.Atoi(buf)
			buf = ""
		default:
			buf += string(v)
		}
	}
	c.DateCon = time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
}

func main() {

	var filename = flag.String("filename", "cdr.txt", "cdr file")
	var phoneNumber = flag.String("phoneNumber", "all", "Phone number for report")
	flag.Parse()
	// open the file
	file, err := os.Open(*filename)

	//handle errors while opening
	if err != nil {
		log.Fatalf("Error when opening file: %s", err)
	}
	fileScanner := bufio.NewScanner(file)

	nowNode := CdrLineNodeRoot
	// read line by line
	for fileScanner.Scan() {
		cdrReader(fileScanner.Text(), nowNode)
		nowNode = nowNode.next
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file: %s", err)
	}

	file.Close()
	if *phoneNumber == "all" {
		phoneRoot := CdrLineNodeRoot
		for phoneRoot.line != nil {
			cdrReport(*&phoneRoot.line.PhoneNumber)
			phoneRoot = phoneRoot.next
		}
	} else {
		cdrReport(*phoneNumber)
	}
}

func cdrReader(line string, Node *CdrLineNode) {
	buf := ""
	if Node == nil {
		Node = new(CdrLineNode)
	}
	cdrLine := new(CdrLine)
	for i, v := range line {
		switch i {
		case 2:
			cdrLine.CallType = buf
			buf = ""
		case 15:
			cdrLine.PhoneNumber = buf
			buf = ""
		case 31:
			cdrLine.DateNachString = buf
			buf = ""
		case 47:
			cdrLine.DateConString = buf
			buf = ""
		case 50:
			buf += string(v)
			cdrLine.Tariff = buf
			buf = ""
		default:
			if v != ' ' && v != ',' {
				buf += string(v)
			}
		}
	}
	cdrLine.strToTime()
	Node.addNode(cdrLine)
}

func cdrReport(phoneNumber string) {
	Phone := make([]CdrLine, 1)
	rootNode := CdrLineNodeRoot
	tariff := ""
	var dur time.Duration
	for rootNode.line != nil {
		if rootNode.line.PhoneNumber == phoneNumber {
			if Phone[0].PhoneNumber == "" {
				Phone[0] = *rootNode.line
			} else {
				Phone = append(Phone, *rootNode.line)
			}
		}
		rootNode = rootNode.next
	}
	sort.SliceStable(Phone, func(i, j int) bool {
		return Phone[i].DateNach.Before(Phone[j].DateNach)
	})
	if len(Phone) != 0 {
		tariff = Phone[0].Tariff
	}
	file, err := os.Create("reports/" + phoneNumber + ".txt")

	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()

	fmt.Fprintf(file, "Tariff index: %s\n", tariff)
	fmt.Fprintf(file, "----------------------------------------------------------------------------\n")
	fmt.Fprintf(file, "Report for phone number %s:\n", phoneNumber)
	fmt.Fprintf(file, "----------------------------------------------------------------------------\n")
	fmt.Fprintf(file, "| Call Type|    Start Time       |      End Time       | Duration |   Cost |\n")
	fmt.Fprintf(file, "----------------------------------------------------------------------------\n")
	curCost := make([]float64, 1)
	for i, val := range Phone {
		curCost = append(curCost, 0)
		currDuration := val.DateCon.Sub(val.DateNach)
		dur += currDuration
		if tariff == "06" {
			if dur <= time.Minute*300 {
				curCost[i] = 0
			} else {
				curCost[i] = float64(time.Duration(currDuration.Minutes()))
			}
		} else if tariff == "03" {
			curCost[i] = float64(time.Duration(currDuration.Minutes())) * 1.5
		} else if tariff == "11" {
			if val.CallType == "02" {
				curCost[i] = 0
			} else {
				if dur <= time.Minute*100 {
					curCost[i] = float64(time.Duration(currDuration.Minutes())) * 0.5
				}
				if dur > time.Minute*100 {
					curCost[i] = float64(time.Duration(currDuration.Minutes())) * 1.5
				}
			}
		}
		fmt.Fprintf(file, "|    %v    | %v | %v | %8v | %6.2f |\n", val.CallType, val.DateNach.Format("2006-01-02 15:04:05"), val.DateCon.Format("2006-01-02 15:04:05"), currDuration, curCost[i])

	}

	fmt.Fprintf(file, "----------------------------------------------------------------------------\n")
	fmt.Fprintf(file, "|                                           Total Cost:|     %6.2f roubles|\n", sum(curCost, tariff))
	fmt.Fprintf(file, "----------------------------------------------------------------------------\n")
}

func sum(mas []float64, tariff string) float64 {
	sum := 0.0
	for _, val := range mas {
		sum += val
	}
	if tariff == "06" {
		sum += 100
	}
	return sum
}
