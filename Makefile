
BINARY_NAME=kadrion

GO_CMD=go

.PHONY: build
build:
	$(GO_CMD) build -o $(BINARY_NAME) ./cmd

.PHONY: run
run: build
	./$(BINARY_NAME)

.PHONY: delete
delete:
	rm -f $(BINARY_NAME)

.PHONY: clean
clean: delete
	$(GO_CMD) clean
