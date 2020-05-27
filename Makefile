BUILD_DIR := build
BINARY := ftacademy

LINUX_BIN := $(BUILD_DIR)/linux/$(BINARY)

BUILD_AT := `date +%FT%T%z`
LDFLAGS := -ldflags "-w -s -X main.build=${BUILD_AT}"

.PHONY: build run publish linux restart config lastcommit clean test
build :
	go build -o $(BUILD_DIR)/$(BINARY) $(LDFLAGS) -v .

run :
	./$(BUILD_DIR)/$(BINARY)

production :
	./$(BUILD_DIR)/$(BINARY) -production

deploy : linux
	rsync -v $(LINUX_BIN) tk11:/home/node/go/bin/
	ssh tk11 supervisorctl restart $(BINARY)

linux :
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(LINUX_BIN) -v .

# Copy env varaible to server
config :
	rsync -v ../.env nodeserver:/home/node/go

lastcommit :
	git log --max-count=1 --pretty=format:%ad_%h --date=format:%Y_%m%d_%H%M

clean :
	go clean -x
	rm -r build/*

static :
	mkdir -p build/static/b2b
	cp ../b2b-client/dist/b2b-client/*js build/static/b2b/
	cp ../b2b-client/dist/index.html.go internal/app/b2b/controller/

test :
	echo $(BUILD)