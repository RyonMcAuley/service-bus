BUILD_DIR = ./bin
APP_NAME = servicebus
CMD_PATH = ./cmd/servicebus


run:
	go run $(CMD_PATH)

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...
