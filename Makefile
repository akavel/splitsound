# Makefile wrapper for gb-generated 'build' script

all:
	./build

%:
	./build $@
