# Usage

This repository contains a collection of UK home computer prices, as published in contemporary computer magazines, together with a Go program to process the data and generate tables suitable for inclusion in a MediaWiki page.

The source data is maintained in an OpenDocument spreadsheet (`data/Home-Computer-Price-Data.ods`).

The Go program reads an exported CSV version of this spreadsheet and generates MediaWiki markup showing the prices of individual computer systems over time.

## Building

The program is written in Go and uses only the standard library.

To build the executable:

    cd cmd/hcp-to-wiki
    GO111MODULE=off go build -o hcp-to-wiki main.go

The resulting executable is `cmd/hcp-to-wiki/hcp-to-wiki`.

## Usage

- Open `data/Home-Computer-Price-Data.ods` in LibreOffice Calc.

- Export the spreadsheet as a CSV file. Use UTF-8 encoding and comma-separated fields.

  Note that the spreadsheet layout is significant. Do not remove or reorder columns when exporting the data.

- Run the program, supplying the CSV filename as its only argument:

      ./cmd/hcp-to-wiki/hcp-to-wiki data/home-computer-prices.csv

- The program processes the data and generates MediaWiki table markup on standard output.

The output can then be used as the basis for updating the relevant wiki page.

# Data format

The spreadsheet contains one record for each published price observation.

The principal columns are:

| Column | Description |
| --- | --- |
| Source | Magazine title |
| Date | Publication date, in YYYY-MM format |
| Ref | Page reference within the magazine |
| System | Computer system name |
| Price (ex VAT) | Published price excluding VAT |
| Price (inc VAT) | Published price including VAT |
| Var | System variant, where applicable |
| Skip | Set to Y to exclude the record |
| Kit | Indicates whether the system was supplied as a kit |
| Board | Indicates whether the entry represents a system board |

The program uses fixed column positions rather than locating columns by their header names. The original spreadsheet layout must therefore be preserved.

The price used for generating the tables is **Price (inc VAT)**. The ex-VAT price is currently ignored.

# Processing

## Input validation

The program performs the following checks:

- The publication date must be in YYYY-MM format.
- The year must be between 1945 and 2099.
- The month must be between 1 and 12.
- The price must be expressed in pounds sterling, must be an integer number of pounds after removing any decimal portion, and must not exceed £100,000.
- The page reference is expected to be in the form pNNNN, with a maximum of five digits.

Records with an invalid date or price are discarded.

An invalid page reference produces a warning but does not prevent the record from being used.

Rows with no system name are ignored.

## Excluding records

A record is excluded if its Skip column contains Y (case-insensitive).

This allows individual observations to be retained in the source spreadsheet without including them in the generated tables.

## Variants

Some computer systems were sold in multiple configurations, typically with different amounts of memory.

The Var column is used to distinguish these configurations.

If a system has only one recorded variant, the variant is omitted from its displayed name.

If multiple variants exist, each is displayed separately, with the variant in parentheses.

For example:

- ZX Spectrum (16K)
- ZX Spectrum (48K)

Variant names are converted to upper case during processing.

The program also checks for potentially missing variant information. If a system has multiple variants and one of its records has a blank variant, a warning is generated identifying the relevant source rows.

## Grouping prices

Publication dates are converted into calendar quarters:

- Q1: January to March
- Q2: April to June
- Q3: July to September
- Q4: October to December

If multiple observations exist for the same system in the same quarter, only the lowest recorded price is retained.

This avoids showing multiple prices for a system in a single quarter.

## System name processing

Some system names are modified or suppressed during preprocessing.

Currently, this processing is hard-coded in the program:

- Science of Cambridge MK14 is renamed MK14.
- A list of system names is excluded from the generated data. The current list is defined in `preprocessSystemData()`.

This processing is intended to accommodate cases where the available price information does not identify a sufficiently well-defined system configuration.

# Output

The output consists of MediaWiki tables grouped into five-year periods.

For example, a period covering 1975–1979 contains one table with a row for each system and four columns for each year, representing the four calendar quarters.

Systems are listed alphabetically.

Only systems with at least one valid price observation in the relevant five-year period are included in that table.

A missing price observation is represented by an em dash (`&mdash;`).

Prices are displayed in pounds sterling, including VAT.

The tables therefore provide a chronological comparison of published UK retail prices, allowing the price history of individual systems to be followed over time.
