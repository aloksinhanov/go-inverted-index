all: unzip build

clean: 
	rm -rf skilling-j
	rm -rf __MACOSX

unzip: clean
	unzip enron.zip

build:
	go build -o looker ./src
	$(info To run ./looker)
