all: hello_world/hello_world.so

hello_world/hello_world.so:
	cd hello_world && go build -buildmode=plugin

clean:
	rm -f hello_world/hello_world.so
