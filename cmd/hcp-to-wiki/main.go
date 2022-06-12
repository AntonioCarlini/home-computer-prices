package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// These constants represent the offset of the items in each advert read from the CSV file
const ( // iota is reset to 0
	adv_magazine = 0 //
	adv_yyyy_mm  = 1 //
	adv_page_num = 2 //
	adv_system   = 3 //
	adv_price    = 4 //
	adv_blank_1  = 5 //
	adv_kit      = 6 //
	adv_board    = 7 //
)

const max_price = 100_000 // Maximum price allowed: anything higher than this is likely to be an error in the data
const max_page_num = 500  // Maximum magazine page number: anything higher than this is likely to be an error in the data
const min_year = 1945     // Earliest acceptable year
const max_year = 2099     // Latest acceptable year

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

// Takes a CSV file representing home computer prices taken from adverts and
// processes that data to produce output in a format suitable for inclusion in a wiki.
//
// The data is grouped by quarter in half decades in each table.
// Systems are listd alphabetically; only systems with at least one valid data point in that table are included.

func main() {

	flag.Parse()

	inputs := flag.Args()
	if len(inputs) != 1 {
		log.Fatalf("Exactly 1 arguments required but %d supplied\n", len(inputs))
	}

	entryFilename := flag.Arg(0)

	data := readCSV(entryFilename)

	// Massage the original CSV data into an array of advertInfo data
	adverts, minDate, maxDate := parseData(data)

	// Build a collection of prices for each system
	systems := buildBySystem(adverts, minDate, maxDate)

	systems = preprocessSystemData(systems)

	// Build array of keys (system names) in alphabetical order
	keys := make([]string, 0, len(systems))
	for key, _ := range systems {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("%-40.40s: %v\n", key, systems[key])
	}

	// Output the final wiki format data
	outputWikidata(systems, keys, minDate, maxDate)
}

// Read data from a CSV file
// Each row of data is represented as an array
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

// Parse the CSV data.
// Skip everything until the header line (with "Source" in the first column) is seen.
// Ignore empty lines.
// Perform some integrity checks on the data.
// Build up an array of advertInfo containing the data that passes validation.
//
// Return the data and also the minimum and maximum date-indices seen when processing the data.
func parseData(data [][]string) (adverts []advertInfo, minDate int, maxDate int) {
	minDate = (max_year + 1) * 4
	maxDate = -1
	adverts = make([]advertInfo, 0)

	searching_for_header := true
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

		advert := advertInfo{csvRowIndex, row[adv_magazine], year, month, page, row[adv_system], price, row[adv_kit], row[adv_board]}
		adverts = append(adverts, advert)
		dateIndex := buildIndexFromAdvertInfo(advert)
		if dateIndex < minDate {
			minDate = dateIndex
		}
		if dateIndex > maxDate {
			maxDate = dateIndex
		}
	}

	return adverts, minDate, maxDate
}

// Given advert data for a range of systems, outputs that data in a form suitable for including in a wiki page
func outputWikidata(systems map[string][]int, keys []string, minDate int, maxDate int) {
	// Loop through quarters in groups of five years.
	// Take the lowest year and make the starting point either YYY0 or YYY5
	// Process data for that group
	// Move on five years and repeat until the start point exceeds the maxDate
	minYear, _ := decodeIndexByQuarter(minDate)
	maxYear, _ := decodeIndexByQuarter(maxDate)
	startYear := (minYear / 5) * 5
	const groupYearsBy = 5
	fmt.Printf("Start Year: %d\n", startYear)
	for groupYear := startYear; groupYear <= maxYear; groupYear = groupYear + groupYearsBy {
		fmt.Printf("== %d - %d ==\n\n", groupYear, groupYear+groupYearsBy-1)
		fmt.Printf("{| class=\"wikitable\"\n")
		fmt.Printf("|-\n")
		fmt.Printf("!  || colspan=\"4\" | %d || colspan=\"4\" | %d || colspan=\"4\" | %d || colspan=\"4\" | %d || colspan=\"4\" | %d\n", groupYear, groupYear+1, groupYear+2, groupYear+3, groupYear+4)
		fmt.Printf("|-\n")
		fmt.Println(" ! style=\"width: 10%;\" | System ")
		fmt.Printf(" ! JAN-MAR || APR-JUN || JUL-SEP || OCT-DEC || JAN-MAR || APR-JUN || JUL-SEP || OCT-DEC || JAN-MAR || APR-JUN || JUL-SEP || OCT-DEC || JAN-MAR || APR-JUN || JUL-SEP || OCT-DEC || JAN-MAR || APR-JUN || JUL-SEP || OCT-DEC\n")
		for _, key := range keys {
			// Pick up the prices for this system:
			prices := systems[key]
			// Ignore this system if it has no price data in the relevant time period
			if !systemHasPriceData(groupYear, groupYear+groupYearsBy-1, minDate, maxDate, prices) {
				continue
			}

			fmt.Printf("|-\n| %s", key)
			for currentYear := groupYear; currentYear < groupYear+groupYearsBy; currentYear++ {
				for currentQuarter := 1; currentQuarter <= 4; currentQuarter++ {
					currentIndex := buildIndexFromYearAndQuarter(currentYear, currentQuarter)
					// fmt.Printf("Processing date %dQ%d  index=%d\n", currentYear, currentQuarter, currentIndex)
					// for this index, find data and display
					if currentQuarter == 1 {
						fmt.Printf("\n     | ")
					} else {
						fmt.Printf("|| ")
					}
					if (currentIndex < minDate) || (currentIndex > maxDate) || (prices[currentIndex-minDate] <= 0) {
						fmt.Printf("style=\"text-align: center;\" | &mdash; ")
					} else {
						fmt.Printf("style=\"text-align: right;\"  | £%-4d   ", prices[currentIndex-minDate])
					}
				}
			}
			fmt.Println("")
		}
		fmt.Printf("|}\n\n") // Close the "wikitable"
	}
}

// Process a date of the form "YYYY-MM".
// return an error if:
//  o the string does not conform to the pattern NNNN-NN, where N is a numeral
//  o the year is not (inclusively) between min_year and max_year constants
//  o the month is not from 1 to 12
// Otherwise return the year and month as integers.
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

// Process the advertInfo array to produce
// Take current entry
// is there a map for that "index"?
// If not, create and populate
// If there is, find this system and replace only iff new price is lower
// byDate map is index=>systemsMap  map[int]
// systemsMap is system=>advtertInfo map[string]advertInfo
func buildByDate(adverts []advertInfo) map[int]map[string]advertInfo {
	byDate := make(map[int]map[string]advertInfo)
	for _, advert := range adverts {
		// fmt.Printf("Processing row %d: %v\n", advert.row, advert)
		index := buildIndexFromAdvertInfo(advert)
		fmt.Printf("Built index %d for %v\n", index, advert)
		if systemMap, ok := byDate[index]; ok {
			if storedAdvert, ok := systemMap[advert.system]; ok {
				// fmt.Printf("systemMap entry exists: %v\n", systemMap[advert.system])
				stored_price := storedAdvert.price
				if (advert.price > 0) && (advert.price < stored_price) {
					fmt.Printf("%d/%d %s found as cheaper (%d against %d); row %d replaces row %d\n", advert.year, advert.month, advert.system, advert.price, stored_price, advert.row, storedAdvert.row)
					systemMap[advert.system] = advert
				} else {
					fmt.Printf("%d/%d %s found as pricier (%d against %d); row %d LEAVES   row %d\n", advert.year, advert.month, advert.system, advert.price, stored_price, advert.row, storedAdvert.row)
				}
			} else {
				// fmt.Printf("systemMap entry missing\n")
				systemMap[advert.system] = advert
			}
		} else {
			byDate[index] = make(map[string]advertInfo, 0)
			systemMap = byDate[index]
			systemMap[advert.system] = advert
			fmt.Printf("%d/%d %s found for first time at %d; row %d\n", advert.year, advert.month, advert.system, advert.price, advert.row)
		}
	}
	return byDate
}

// This function applies some pre-processing to the gathered data.
// For now this is hard-coded, but may later be driven by an external configuration file.
// The following changes are made:
// o "Science of Cambridge MK14" is re-written as "MK14"
// o Data for "Apple II" is suppressed, as the configuration is unclear
// o Data for "Exidy Sorcerer" is suppressed as the configuration is unclear
func preprocessSystemData(systems map[string][]int) map[string][]int {
	result := make(map[string][]int, 0)
	for name, _ := range systems {
		if (name == "Apple II") || (name == "Exidy Sorcerer") {
			// Drop this data
		} else if name == "Science of Cambridge MK14" {
			result["MK14"] = systems[name]
		} else {
			result[name] = systems[name]
		}
	}
	return result
}

// Given an advertInfo, this function produces an int that represents that year and quarter.
// Months 1-3 are 0 (Q1), months 4-6 are 1 (Q2) etc.
// The final index is (year*12 + quarter)
func buildIndexFromAdvertInfo(advert advertInfo) int {
	quarter := ((advert.month - 1) / 3)
	return (advert.year * 4) + quarter
}

// Given a year and a quarter, combine them into a date-index integer
func buildIndexFromYearAndQuarter(year int, quarter int) int {
	return (year * 4) + (quarter - 1)
}

// Given a date-index, return the year and quarter which it represents
func decodeIndexByQuarter(index int) (year int, quarter int) {
	year = index / 4
	quarter = index - (year * 4) + 1
	return year, quarter
}

// Given a number of advertInfo objects, build a map of system => price-array
// The price array index should be 0 for minDate and increase up to (maxDate-minDate) for maxDate
func buildBySystem(adverts []advertInfo, minDate int, maxDate int) map[string][]int {
	result := make(map[string][]int, 0)

	for _, advert := range adverts {
		if _, ok := result[advert.system]; !ok {
			// This system has been seen for the first time.
			// Create its price array
			result[advert.system] = make([]int, maxDate-minDate+1)
		}
		// By this point the price array must exist, so update if appropriate.
		index := buildIndexFromAdvertInfo(advert)
		storedPrice := result[advert.system][index-minDate]
		if (advert.price < storedPrice) || (storedPrice <= 0) {
			result[advert.system][index-minDate] = advert.price
		}
	}
	return result
}

// A helper function that determines whether there is price data available for the specified period
func systemHasPriceData(startYear int, endYear int, minDate int, maxDate int, prices []int) bool {
	systemHasPriceData := false

	lowestIndex := buildIndexFromYearAndQuarter(startYear, 1)
	lowestValidIndex := max(lowestIndex, minDate)
	highestIndex := buildIndexFromYearAndQuarter(endYear, 1)
	highestValidIndex := min(highestIndex, maxDate)

	for idx := lowestValidIndex; idx <= highestValidIndex; idx++ {
		if prices[idx-minDate] > 0 {
			systemHasPriceData = true
			break
		}
	}
	return systemHasPriceData
}

// golang doesn't have min/max so provide them here
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
