GOPATH=$(CURDIR)/../../../../
GOPATHCMD=GOPATH=$(GOPATH)

COVERDIR=$(CURDIR)/.cover
COVERAGEFILE=$(COVERDIR)/cover.out

.PHONY: deps deps-ci coverage coverage-ci test test-watch coverage coverage-html

test:
	@${GOPATHCMD} ginkgo -race --failFast ./...

test-watch:
	@${GOPATHCMD} ginkgo watch -cover -r ./...

coverage-ci:
	@mkdir -p $(COVERDIR)
	@${GOPATHCMD} ginkgo -r -covermode=count --cover --trace ./
	@echo "mode: count" > "${COVERAGEFILE}"
	@find ./* -type f -name *.coverprofile -exec grep -h -v "^mode:" {} >> "${COVERAGEFILE}" \; -exec rm -f {} \;

coverage: coverage-ci
	@sed -i -e "s|_$(CURDIR)/|./|g" "${COVERAGEFILE}"
	@cp "${COVERAGEFILE}" coverage.txt

coverage-html:
	@$(GOPATHCMD) go tool cover -html="${COVERAGEFILE}" -o .cover/report.html

dep-ensure:
	@$(GOPATHCMD) dep ensure -v

dep-update:
	@$(GOPATHCMD) dep ensure -update -v

vet:
	@$(GOPATHCMD) go vet ./...

fmt:
	@$(GOPATHCMD) go vet ./...

dcup:
	@docker-compose up -d

dcdn:
	@docker-compose down