GO=go 
SRC=router.go

all: $(SRC)
	$(GO) build -o devicekeeper $(SRC)
	GOOS=linux GOARCH=386  $(GO) build -o devicekeeper.linux $(SRC)