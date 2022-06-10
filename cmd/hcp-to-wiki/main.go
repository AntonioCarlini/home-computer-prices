package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// These constants represent the offset of the items in each advert read from the CSV file
const ( // iota is reset to 0
	adv_magazine = 0 // transaction ID
	adv_yyyy_mm  = 1 //
	adv_page_num = 2 //
	adv_system   = 3 //
	adv_price    = 4 //
	adv_blank_1  = 5 //
	adv_kit      = 6 //
	adv_board    = 7 //
)

const max_price = 100_000
const max_page_num = 500
const min_year = 1945
const max_year = 2099

type advertInfo struct {
	row      int
	magazine string // Magazine Title
	year     int    // Year (1945..current)
	month    int    // Month (1..12)
	page     int    // page number
	system   string // Computer system name
	price    int    // Price in pounds, including VAT
	kit      string // TODO: True if the system had to be assembled
	board    string // TODO: True if the system was a system board
}

// TODO
func main() {

	flag.Parse()

	inputs := flag.Args()
	if len(inputs) != 1 {
		log.Fatalf("Exactly 1 arguments required but %d supplied\n", len(inputs))
	}

	entryFilename := flag.Arg(0)

	data := readCSV(entryFilename)

	adverts := parseData(data)

	// data_by_date := make(map[int], 0)
	fmt.Printf("entries: %v\n", adverts[0:10])

}

// TODO
func readCSV(filename string) [][]string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Cannot open '%s': %s\n", filename, err.Error())
	}
	defer f.Close()

	r := csv.NewReader(f)

	transactions, err := r.ReadAll()
	if err != nil {
		log.Fatalln("Cannot read CSV data:", err.Error())
	}

	return transactions
}

//TODO
func parseData(data [][]string) []advertInfo {
	searching_for_header := true
	adverts := make([]advertInfo, 0)

	for i, row := range data {
		csvRowIndex := i + 1
		valid := true

		// Skip all data until a row with a suitable header line is seen
		if searching_for_header {
			if row[adv_magazine] == "Source" {
				searching_for_header = false
			}
			continue
		}

		// Entirely empty lines must be ignored. As an approximation, ignore any line without a system title, as that cannot contain meaningful data.
		if len(row[adv_system]) == 0 {
			continue
		}

		// The YYYY-DD field must be of the correct format DD must be 01..12 and YYYY must be greater than 1945 but less than 2099
		year, month, err := handle_yyyy_mm(row[adv_yyyy_mm])
		if err != nil {
			valid = false
			fmt.Printf("Line %d: Bad YYYY-DD [%s] (%s) in [%v]\n", csvRowIndex, row[adv_yyyy_mm], err, row)
		}

		// The page format must be pN{1,5}}, so at least one N but no more than 5.
		// Note that the page number does not influence the final output, so "valid" is not adjusted and the data may be used
		page, err := handle_page_number(row[adv_page_num])
		if err != nil {
			fmt.Printf("Line %d: Bad page number [%s] (%s) in [%v]\n", csvRowIndex, row[adv_page_num], err, row)
		}

		// The price must be in pounds, must be an integer and must be less than £100,000
		// The CSV will be encoded as UTF-8 and the "£" symbol will have to be checked as UTF-8
		price, err := handle_price(row[adv_price])
		if err != nil {
			valid = false
			fmt.Printf("Line %d: Bad price [%s] (%s) in [%v]\n", csvRowIndex, row[adv_price], err, row)
		}

		// TODO
		//  The kit field must be Y, N, ? or blank

		if !valid {
			continue
		}

		entry := advertInfo{csvRowIndex, row[adv_magazine], year, month, page, row[adv_system], price, row[adv_kit], row[adv_board]}
		adverts = append(adverts, entry)
	}

	return adverts
}

// Process a date of the form "YYYY-MM".
// return an error if:
//  o the string does not conform to the pattern NNNN-NN, where N is a numeral
//  o the year is not (inclusively) between min_year and max_year constants
//  o the month is not from 1 to 12
// Otherwose return the year and month as integers.
//
// TODO: make the upper limit for YYYY the current year
func handle_yyyy_mm(yyyy_mm string) (year int, month int, err error) {
	year = -1
	month = -1
	var local_err error

	date_sep := yyyy_mm[4:5]
	if date_sep != "-" {
		local_err = fmt.Errorf("bad YYYY-MM separator [%s] from [%s]", date_sep, yyyy_mm)
	} else if len(yyyy_mm) != 7 {
		local_err = fmt.Errorf("bad YYYY-MM: length invalid: [%s]", yyyy_mm)
	}
	year_text := yyyy_mm[0:4]
	year, err = strconv.Atoi(year_text)
	if err != nil {
		local_err = fmt.Errorf("bad Year digits [%s] (%w)", year_text, err)
	} else if (year < min_year) || (year > max_year) {
		local_err = fmt.Errorf("bad Year  [%d] outside range %d-%d", year, min_year, max_year)
	}
	month_text := yyyy_mm[5:]
	month, err = strconv.Atoi(month_text)
	if err != nil {
		local_err = fmt.Errorf("bad Month digits [%s]", month_text)
	} else if (month < 1) || (month > 12) {
		local_err = fmt.Errorf("bad Month [%d]", month)
	}
	return year, month, local_err
}

// Process a page number of the form "pNNNN".
// return an error if:
// Otherwise return the page number as an integer.
//
// TODO: allow for roman numberals: e.g. pii

func handle_page_number(page_num_text string) (page int, err error) {
	page = -1000
	var local_err error

	if len(page_num_text) < 2 {
		local_err = fmt.Errorf("bad page number text [%s]", page_num_text)
	} else {
		if page_num_text[0] != 'p' {
			local_err = fmt.Errorf("bad page number format [%s]", page_num_text)
		} else {
			page, err = strconv.Atoi(page_num_text[1:])
			if err != nil {
				local_err = fmt.Errorf("bad page number data [%s] (%w)", page_num_text, err)
			}
			if (page < 0) || (page > max_page_num) {
				local_err = fmt.Errorf("bad page number value [%d]", page)
			}
		}
	}
	return page, local_err
}

// Process a price of the form "£NNNN".
// return an error if:
// Otherwise return the price as an integer.
//
// TODO: allow for roman numberals: e.g. pii

func handle_price(price_text string) (price int, err error) {
	price = -1
	var local_err error

	// The price must be in pounds, must be an integer and must be less than £100,000
	// The CSV will be encoded as UTF-8 and the "£" symbol will have to be checked as UTF-8
	price_currency := []rune(price_text)[0]
	price_value := string([]rune(price_text)[1:])          // Remove first character, allowing for possibility that it is UTF-8
	price_value = strings.SplitN(price_value, ".", 2)[0]   // Remove everything after a decimal point
	price_value = strings.ReplaceAll(price_value, ",", "") // Remove all commas
	if price_currency != []rune("£")[0] {
		local_err = fmt.Errorf("bad Price Currency [%c] from [%s]", price_currency, price_text)
	} else {
		possible_price, err := strconv.Atoi(price_value)
		if err != nil {
			local_err = fmt.Errorf("bad Price Data [%s]", price_value)
		} else if possible_price > max_price {
			local_err = fmt.Errorf("unlikely Price Data [%s] (greater than %d)", price_text, max_price)
		} else {
			price = possible_price
		}
	}
	return price, local_err
}
