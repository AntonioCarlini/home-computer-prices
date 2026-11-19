GO := go
SOFFICE := soffice

BIN_DIR := bin

ODS := data/Home-Computer-Price-Data.ods
CSV := $(BIN_DIR)/Home-Computer-Price-Data.csv
PROGRAM := $(BIN_DIR)/hcp-to-wiki
WIKI := $(BIN_DIR)/home-computer-prices.wiki

.PHONY: all csv wiki clean

all: wiki

$(BIN_DIR):
	mkdir -p $@

# Export the ODS spreadsheet to CSV.
$(CSV): $(ODS) | $(BIN_DIR)
	$(SOFFICE) --headless --convert-to 'csv:Text - txt - csv (StarCalc):44,34,76,1,,0,false,true,true' --outdir $(BIN_DIR) $(ODS)

csv: $(CSV)

# Build the Go executable.
$(PROGRAM): go.mod cmd/hcp-to-wiki/main.go | $(BIN_DIR)
	$(GO) build -o $@ ./cmd/hcp-to-wiki

# Generate MediaWiki markup.
$(WIKI): $(CSV) $(PROGRAM)
	$(PROGRAM) $(CSV) > $@

wiki: $(WIKI)

clean:
	rm -rf $(BIN_DIR)
