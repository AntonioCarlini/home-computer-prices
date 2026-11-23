package main

import "testing"

func TestParseDataValidRows(t *testing.T) {
	header := []string{"Source"}
	first_row := make([]string, adv_board+1)
	first_row[adv_magazine] = "Magazine A"
	first_row[adv_yyyy_mm] = "1980-01"
	first_row[adv_page_num] = "p10"
	first_row[adv_system] = "  Computer A  "
	first_row[adv_price_inc_vat] = "£100"
	first_row[adv_variant] = "  16k  "
	first_row[adv_kit] = "N"
	first_row[adv_board] = "N"

	second_row := make([]string, adv_board+1)
	second_row[adv_magazine] = "Magazine B"
	second_row[adv_yyyy_mm] = "1980-04"
	second_row[adv_page_num] = "p20"
	second_row[adv_system] = "Computer B"
	second_row[adv_price_inc_vat] = "£200"
	second_row[adv_variant] = "32K"

	data := [][]string{
		header,
		first_row,
		second_row,
	}

	adverts, minDate, maxDate := parseData(data)

	if len(adverts) != 2 {
		t.Fatalf("expected 2 adverts, got %d", len(adverts))
	}

	first := adverts[0]

	if first.row != 2 {
		t.Errorf("expected first advert row 2, got %d", first.row)
	}
	if first.magazine != "Magazine A" {
		t.Errorf("expected magazine Magazine A, got %q", first.magazine)
	}
	if first.year != 1980 || first.month != 1 {
		t.Errorf("expected date 1980-01, got %04d-%02d", first.year, first.month)
	}
	if first.page != 10 {
		t.Errorf("expected page 10, got %d", first.page)
	}
	if first.system != "Computer A" {
		t.Errorf("expected system Computer A, got %q", first.system)
	}
	if first.price != 100 {
		t.Errorf("expected price 100, got %d", first.price)
	}
	if first.variant != "16K" {
		t.Errorf("expected variant 16K, got %q", first.variant)
	}
	if first.kit != "N" {
		t.Errorf("expected kit N, got %q", first.kit)
	}
	if first.board != "N" {
		t.Errorf("expected board N, got %q", first.board)
	}

	second := adverts[1]

	if second.row != 3 {
		t.Errorf("expected second advert row 3, got %d", second.row)
	}
	if second.system != "Computer B" {
		t.Errorf("expected system Computer B, got %q", second.system)
	}
	if second.variant != "32K" {
		t.Errorf("expected variant 32K, got %q", second.variant)
	}

	wantMinDate := buildIndexFromYearAndQuarter(1980, 1)
	wantMaxDate := buildIndexFromYearAndQuarter(1980, 2)

	if minDate != wantMinDate {
		t.Errorf("expected minDate %d, got %d", wantMinDate, minDate)
	}
	if maxDate != wantMaxDate {
		t.Errorf("expected maxDate %d, got %d", wantMaxDate, maxDate)
	}
}
