all: gorduino treeprinter

gorduino: gorduino.go
	go build gorduino.go

treeprinter: treeprinter.go
	go build treeprinter.go

test: gorduino
	perl test.pl

clean:
	rm -f gorduino treeprinter
