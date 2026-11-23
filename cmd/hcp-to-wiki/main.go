package main

// Takes a CSV file representing home computer prices taken from adverts and
// processes that data to produce output in a format suitable for inclusion in a wiki.
//
// The data is grouped by quarter in half decades in each table.
// Systems are listd alphabetically; only systems with at least one valid data point in that table are included.

// TODO:
//   handle "*" in the expected way
//   add option to suppress entries with fewer than N quarters of data
//   --low-frequency-limit N  TODO: better name needed
//
//   add option to display entries with few data points in a separate set of tables
//   --generate-low-frequency-tables   TODO: better name needed

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
	adv_magazine      = 0  //
	adv_yyyy_mm       = 1  //
	adv_page_num      = 2  //
	adv_system        = 3  //
	adv_price_ex_vat  = 4  //
	adv_price_inc_vat = 5  //
	adv_blank_1       = 6  //
	adv_variant       = 7  //
	adv_skip          = 8  //
	adv_kit           = 9  //
	adv_board         = 10 //
	adv_system_copy   = 13 // temp
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
	variant  string // System variant
	kit      string // TODO: True if the system had to be assembled
	board    string // TODO: True if the system was a system board
}

func main() {

	flag.Parse()

	inputs := flag.Args()
	if len(inputs) != 1 {
		log.Fatalf("Exactly 1 argument required but %d supplied\n", len(inputs))
	}

	entryFilename := flag.Arg(0)

	data := readCSV(entryFilename)

	// Massage the original CSV data into an array of advertInfo data
	adverts, minDate, maxDate := parseData(data)

	// Build a map of adverts that have no variant
	blankVariantSystems := buildBlankVariantMap(adverts)

	// Build a collection of prices for each system
	systems := buildBySystem(adverts, blankVariantSystems, minDate, maxDate)

	systems = preprocessSystemData(systems)

	// Build array of keys (system names) in alphabetical order
	keys := make([]string, 0, len(systems))
	for key, _ := range systems {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(os.Stderr, "%-40.40s: %v\n", key, systems[key])
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
// Return an array of advertInfo structs and also the minimum and maximum date-indices seen when processing the data.
func parseData(data [][]string) (adverts []advertInfo, minDate int, maxDate int) {
	minDate = (max_year + 1) * 4
	maxDate = -1
	adverts = make([]advertInfo, 0)

	searching_for_header := true
	for i, row := range data {
		csvRowIndex := i + 1

                // Ignore empty rows.
		if len(row) == 0 {
			continue
		}
		valid := true

		// Skip all data until a row with a suitable header line is seen
		if searching_for_header {
			if row[adv_magazine] == "Source" {
				searching_for_header = false
			}
			continue
		}

		// Check that the row contains all required fields.
		if len(row) <= adv_board {
			fmt.Fprintf(os.Stderr, "Line %d: Incomplete CSV row: expected at least %d fields, got %d: [%v]\n",
				csvRowIndex, adv_board+1, len(row), row)
			continue
		}

		// Make sure the system name has no leading or trailing spaces
		system := strings.TrimSpace(row[adv_system])

		// Decide if this entry should be dropped
		skip := strings.ToUpper(strings.TrimSpace(row[adv_skip]))
		if strings.Contains(skip, "Y") {
			fmt.Fprintf(os.Stderr, "Line %d: Skipping: %s\n", csvRowIndex, system)
			continue
		}

		// Entirely empty lines must be ignored. As an approximation, ignore any line without a system title, as that cannot contain meaningful data.
		if len(system) == 0 {
			continue
		}

		// The YYYY-DD field must be of the correct format DD must be 01..12 and YYYY must be greater than 1945 but less than 2099
		year, month, err := handle_yyyy_mm(row[adv_yyyy_mm])
		if err != nil {
			valid = false
			fmt.Fprintf(os.Stderr, "Line %d: Bad YYYY-DD [%s] (%s) in [%v]\n", csvRowIndex, row[adv_yyyy_mm], err, row)
		}

		// The page format must be pN{1,5}}, so at least one N but no more than 5.
		// Note that the page number does not influence the final output, so "valid" is not adjusted and the data may be used
		page, err := handle_page_number(row[adv_page_num])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Line %d: Bad page number [%s] (%s) in [%v]\n", csvRowIndex, row[adv_page_num], err, row)
		}

		// The price must be in pounds, must be an integer and must be less than £100,000
		// The CSV will be encoded as UTF-8 and the "£" symbol will have to be checked as UTF-8
		price, err := handle_price(row[adv_price_inc_vat])
		if err != nil {
			valid = false
			fmt.Fprintf(os.Stderr, "Line %d: Bad price [%s] (%s) in [%v]\n", csvRowIndex, row[adv_price_inc_vat], err, row)
		}

		// TODO
		//  The kit field must be Y, N, ? or blank

		variant := strings.ToUpper(strings.TrimSpace(row[adv_variant]))

		if !valid {
			continue
		}

		advert := advertInfo{csvRowIndex, row[adv_magazine], year, month, page, system, price, variant, row[adv_kit], row[adv_board]}
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

	// If there is no valid price data, produce an empty wiki table.
	if minDate > maxDate {
		fmt.Printf("__TOC__\n\n")
		fmt.Println("{| class=\"wikitable\" style=\"width: 95%;\"")
		fmt.Println("|-")
		fmt.Println("! System")
		fmt.Println("|}")
		fmt.Println()
		fmt.Println("[[Category:Classic Computing]]")
		return
	}

	// Loop through quarters in groups of five years.
	// Take the lowest year and make the starting point either YYY0 or YYY5
	// Process data for that group
	// Move on five years and repeat until the start point exceeds the maxDate
	minYear, _ := decodeIndexByQuarter(minDate)
	maxYear, _ := decodeIndexByQuarter(maxDate)
	startYear := (minYear / 5) * 5
	const groupYearsBy = 5
        fmt.Printf("__TOC__\n\n") // Add the TOC
	for groupYear := startYear; groupYear <= maxYear; groupYear = groupYear + groupYearsBy {
		fmt.Printf("== %d - %d ==\n\n", groupYear, groupYear+groupYearsBy-1)
		fmt.Println("{| class=\"wikitable\"  style=\"width: 95%;\"")
		fmt.Printf("|-\n")
		fmt.Printf("!  || colspan=\"4\" | %d || colspan=\"4\" | %d || colspan=\"4\" | %d || colspan=\"4\" | %d || colspan=\"4\" | %d\n", groupYear, groupYear+1, groupYear+2, groupYear+3, groupYear+4)
		fmt.Printf("|-\n")
		fmt.Println(" ! style=\"width: 12%;\" | System ")
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
        fmt.Printf("[[Category:Classic Computing]]\n\n") // Add the category

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

	// Check the length before attempting to access individual characters.
	if len(yyyy_mm) != 7 {
		return year, month, fmt.Errorf("bad YYYY-MM: length invalid: [%s]", yyyy_mm)
	}

	// Check the separator.
	if yyyy_mm[4] != '-' {
		return year, month, fmt.Errorf("bad YYYY-MM separator [%c] from [%s]", yyyy_mm[4], yyyy_mm)
	}

	// Extract and validate the year.
	year_text := yyyy_mm[0:4]
	year, err = strconv.Atoi(year_text)
	if err != nil {
		return -1, -1, fmt.Errorf("bad Year digits [%s] (%w)", year_text, err)
	}

	if (year < min_year) || (year > max_year) {
		return -1, -1, fmt.Errorf("bad Year [%d] outside range %d-%d", year, min_year, max_year)
	}

	// Extract and validate the month.
	month_text := yyyy_mm[5:7]
	month, err = strconv.Atoi(month_text)
	if err != nil {
		return -1, -1, fmt.Errorf("bad Month digits [%s]", month_text)
	}

	if (month < 1) || (month > 12) {
		return -1, -1, fmt.Errorf("bad Month [%d]", month)
	}

	return year, month, nil
}

// Process a page number of the form "pNNNN".
// return an error if the page number appears to be invalid
// Otherwise return the page number as an integer.
//
// TODO: allow for roman numberals: e.g. pii

func handle_page_number(page_num_text string) (page int, err error) {
	page_num_text = strings.TrimSpace(page_num_text) // Trim leading and trailing spaces: strconv.Atoi() doesn't like it!
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
	price_text = strings.TrimSpace(price_text)
	if len(price_text) == 0 {
		return 0, fmt.Errorf("Missing Price")
	}
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
		fmt.Fprintf(os.Stderr, "Built index %d for %v\n", index, advert)
		if systemMap, ok := byDate[index]; ok {
			if storedAdvert, ok := systemMap[advert.system]; ok {
				// fmt.Printf("systemMap entry exists: %v\n", systemMap[advert.system])
				stored_price := storedAdvert.price
				if (advert.price > 0) && (advert.price < stored_price) {
					fmt.Fprintf(os.Stderr, "%d/%d %s found as cheaper (%d against %d); row %d replaces row %d\n", advert.year, advert.month, advert.system, advert.price, stored_price, advert.row, storedAdvert.row)
					systemMap[advert.system] = advert
				} else {
					fmt.Fprintf(os.Stderr, "%d/%d %s found as pricier (%d against %d); row %d LEAVES   row %d\n", advert.year, advert.month, advert.system, advert.price, stored_price, advert.row, storedAdvert.row)
				}
			} else {
				// fmt.Printf("systemMap entry missing\n")
				systemMap[advert.system] = advert
			}
		} else {
			byDate[index] = make(map[string]advertInfo, 0)
			systemMap = byDate[index]
			systemMap[advert.system] = advert
			fmt.Fprintf(os.Stderr, "%d/%d %s found for first time at %d; row %d\n", advert.year, advert.month, advert.system, advert.price, advert.row)
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
	systemsToSuppress := []string{"xApple II", "xCommodore PET", "xExidy Sorcerer", "xTandy TRS-80 Model 1"}
	result := make(map[string][]int, 0)
	for name, _ := range systems {
		if sliceContainsString(systemsToSuppress, name) {
			// Drop this data
			fmt.Fprintf(os.Stderr, "Dropping %s\n", name)
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
//
// Rather than just using the system name when building prices, build a "full system name" from the system name and the varint string.
// This better fits the new spreadhseet format where something like the Atari 400, which cmae in 16K, 32K and 48K versions, can now be listed as "Atari 400" with three variations.

func buildBySystem(adverts []advertInfo, blankVariantSystems map[string][]advertInfo, minDate int, maxDate int) map[string][]int {
	result := make(map[string][]int, 0)
	fullSystemNameMap := buildFullSystemNameMap(adverts)

	// Check for common data entry errors
	validateSystemNameMap(fullSystemNameMap, blankVariantSystems)

	for _, advert := range adverts {
		fullSystemName := buildAdvertSystemName(advert, fullSystemNameMap)
		if _, ok := result[fullSystemName]; !ok {
			// This system has been seen for the first time.
			// Create its price array
			result[fullSystemName] = make([]int, maxDate-minDate+1)
		}
		// By this point the price array must exist, so update if appropriate.
		index := buildIndexFromAdvertInfo(advert)
		storedPrice := result[fullSystemName][index-minDate]
		if (advert.price < storedPrice) || (storedPrice <= 0) {
			result[fullSystemName][index-minDate] = advert.price
		}
	}

	// TODO
	// work through systems and look for cases where there are exactly two variants, one of which is *
	// In those cases, collapse * data into other variants data, choosing the lowest price in each case.
	// Issue a message?

	return result
}

// A helper function that determines whether there is price data available for the specified period
func systemHasPriceData(startYear int, endYear int, minDate int, maxDate int, prices []int) bool {
	systemHasPriceData := false

	lowestIndex := buildIndexFromYearAndQuarter(startYear, 1)
	lowestValidIndex := max(lowestIndex, minDate)
	highestIndex := buildIndexFromYearAndQuarter(endYear, 4)
	highestValidIndex := min(highestIndex, maxDate)

	for idx := lowestValidIndex; idx <= highestValidIndex; idx++ {
		if prices[idx-minDate] > 0 {
			systemHasPriceData = true
			break
		}
	}
	return systemHasPriceData
}

// Given an array of avertInfo this function builds a map of system name to an array of variant strings for that system.
// Note that if an advertInfo has an empty variant, that is included.
func buildFullSystemNameMap(adverts []advertInfo) map[string][]string {
	result := make(map[string][]string, 0)
	for _, advert := range adverts {
		if _, ok := result[advert.system]; !ok {
			// This system has been seen for the first time.
			// Create its variant array
			result[advert.system] = make([]string, 0)
		}
		fullSystemName := buildFullSystemName(advert)
		if !sliceContainsString(result[advert.system], fullSystemName) {
			result[advert.system] = append(result[advert.system], fullSystemName)
		}
	}
	return result
}

// Given an array of advertInfo, build a map of system => array of advertInfo that have no variant.
// This will be used when listing "suspicious" entries to list the lines in the original CSV
// where those entries originated.
func buildBlankVariantMap(adverts []advertInfo) map[string][]advertInfo {
	result := make(map[string][]advertInfo, 0)
	for _, advert := range adverts {
		if _, ok := result[advert.system]; !ok {
			// This system has been seen for the first time.
			// Create its variant array
			result[advert.system] = make([]advertInfo, 0)
		}
		if advert.variant == "" {
			result[advert.system] = append(result[advert.system], advert)
		}
	}
	return result
}

// Given an advert object, this function returns the "full system name".
//
// The "full system name" is defined as follows:
// If the variant field is blank (leading and trailing spaces have already been suppressed), then just use the system name.
// Otherwise use the system name followed by a space and the variant in brackets.
func buildFullSystemName(advert advertInfo) string {
	if advert.variant == "" {
		return advert.system
	} else {
		return advert.system + " (" + advert.variant + ")"
	}
}

// Return the name by which this system variant should be known.
//
// If the system is not registered or is registered with no entries, return an error string.
// If either of these happens, it must be a program logic error rather than a data error.
// Otherwise if there is only one entry, then there is no need to use the variant as a disambiguator, so use the system name alone.
// Finally, if more than one variant exists, then build a full system name that depends on the variant.
func buildAdvertSystemName(advert advertInfo, fullSystemNameMap map[string][]string) string {
	if entry, ok := fullSystemNameMap[advert.system]; ok {
		if len(entry) == 0 {
			return "BAD**BAD"
		} else if len(entry) == 1 {
			return advert.system
		} else {
			return buildFullSystemName(advert)
		}
	} else {
		return "FAILED**FAILED"
	}
}

// A common error is for a variant to be ommitted by accident.
// Normally you would expect either all systems to have an identical variant (e.g. "Gemini Galaxy 2" which always has a variant of "64K")
// or for there to be multiple variants (e.g. "ZX Spectrum" which has "16K" and "48K").
// In the former case, because only one variant exists, the system is listed as "Gemini Galaxy 2".
// In the latter case, two systems are listed "ZX Spectrum (16K)" "ZX Spectrum (48K)".
// If by accident a Galaxy entry omits the variant, there will now be a new entry for "Gemini Galaxy 2" which will cause a "Galaxy Gemini 2 (64K)" entry to erroneously appear.
//
// To guard against this, scan the systems map and look for variant arrays of len > 1 where one entry is empty (i.e. "").
func validateSystemNameMap(fullSystemNameMap map[string][]string, blankVariantSystems map[string][]advertInfo) {
	for name, _ := range fullSystemNameMap {
		variants := fullSystemNameMap[name]
		// Complain if there is more than one variant and any of the variants exactly matches the name (i.e. the key)
		if len(variants) > 1 {
			if sliceContainsString(variants, name) {
				refs := "REFS: "
				for _, advert := range blankVariantSystems[name] {
					refs = refs + strconv.Itoa(advert.row) + " "
				}
				fmt.Fprintf(os.Stderr, "Suspect entry: %s   %s : %s\n", name, variants, refs)
			}
		}
	}

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

func sliceContainsString(slice []string, candidate string) bool {
	for _, member := range slice {
		if member == candidate {
			return true
		}
	}
	return false
}
