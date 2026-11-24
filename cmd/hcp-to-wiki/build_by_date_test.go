package main

import "testing"

func TestBuildByDate(t *testing.T) {
	adverts := []advertInfo{
		{
			row: 1, year: 1980, month: 1,
			system: "Computer A", price: 100,
		},
		{
			row: 2, year: 1980, month: 2,
			system: "Computer A", price: 150,
		},
		{
			row: 3, year: 1980, month: 3,
			system: "Computer A", price: 80,
		},
		{
			row: 4, year: 1980, month: 1,
			system: "Computer B", price: 200,
		},
		{
			row: 5, year: 1980, month: 4,
			system: "Computer A", price: 90,
		},
	}

	result := buildByDate(adverts)

	q1 := buildIndexFromYearAndQuarter(1980, 1)
	q2 := buildIndexFromYearAndQuarter(1980, 2)

	if len(result) != 2 {
		t.Fatalf("expected 2 quarter entries, got %d", len(result))
	}

	if len(result[q1]) != 2 {
		t.Fatalf("expected 2 systems in Q1, got %d", len(result[q1]))
	}

	if len(result[q2]) != 1 {
		t.Fatalf("expected 1 system in Q2, got %d", len(result[q2]))
	}

	// Computer A should retain the cheapest Q1 advert.
	q1ComputerA, ok := result[q1]["Computer A"]
	if !ok {
		t.Fatal("Computer A missing from Q1")
	}
	if q1ComputerA.price != 80 {
		t.Errorf("expected Q1 Computer A price 80, got %d",
			q1ComputerA.price)
	}
	if q1ComputerA.row != 3 {
		t.Errorf("expected Q1 Computer A row 3, got %d",
			q1ComputerA.row)
	}

	// Computer B should have its own Q1 entry.
	q1ComputerB, ok := result[q1]["Computer B"]
	if !ok {
		t.Fatal("Computer B missing from Q1")
	}
	if q1ComputerB.price != 200 {
		t.Errorf("expected Q1 Computer B price 200, got %d",
			q1ComputerB.price)
	}

	// Computer A in Q2 should be independent of its Q1 entry.
	q2ComputerA, ok := result[q2]["Computer A"]
	if !ok {
		t.Fatal("Computer A missing from Q2")
	}
	if q2ComputerA.price != 90 {
		t.Errorf("expected Q2 Computer A price 90, got %d",
			q2ComputerA.price)
	}
	if q2ComputerA.row != 5 {
		t.Errorf("expected Q2 Computer A row 5, got %d",
			q2ComputerA.row)
	}
}
