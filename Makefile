proto: graph
graph:
	@echo 'Building graph protocols'
	@capnp compile -I$(GOPATH)/src/zombiezen.com/go/capnproto2/std -ogo api/graph/message.capnp
	# @mv api/graph/*.go pkg/graph
deps:
	@go get -u zombiezen.com/go/capnproto2/...
