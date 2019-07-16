COVERDIR=$(CURDIR)/.cover
COVERAGEFILE=$(COVERDIR)/cover.out

EXAMPLES=$(shell ls ./examples)

$(EXAMPLES): %:
	$(eval EXAMPLE=$*)
	@:

run:
	@test ! -z "$(EXAMPLE)" && go run ./examples/$(EXAMPLE) || echo "Usage: make [$(EXAMPLES)] run"

test:
	@ginkgo -race --failFast ./...

test-watch:
	@ginkgo watch -cover -r ./...

coverage-ci:
	@mkdir -p $(COVERDIR)
	@ginkgo -r -covermode=count --cover --trace ./
	@echo "mode: count" > "${COVERAGEFILE}"
	@find ./* -type f -name *.coverprofile -exec grep -h -v "^mode:" {} >> "${COVERAGEFILE}" \; -exec rm -f {} \;

coverage: coverage-ci
	@sed -i -e "s|_$(CURDIR)/|./|g" "${COVERAGEFILE}"
	@cp "${COVERAGEFILE}" coverage.txt

coverage-html:
	@go tool cover -html="${COVERAGEFILE}" -o .cover/report.html

vet:
	@go vet ./...

fmt:
	@go vet ./...

.PHONY: test test-watch coverage coverage-ci coverage-html vet fmt
