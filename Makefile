.PHONY: clean-log
clean-log:
	rm gen/$(source)/log
	touch gen/$(source)/log

.PHONY: clean-blacklist
clean-blacklist:
	rm gen/$(source)/blacklist
	touch gen/$(source)/blacklist

.PHONY: clean
clean: clean-blacklist clean-log

.PHONY: generate
generate:
	go run cmd/anki-gen/main.go \
		-i ./lib/$(source)/$(file) \
		-t ./templates/$(source).tmpl \
		-e ./gen/$(source)/log \
		-e ./gen/$(source)/blacklist \
		-d $(source)
