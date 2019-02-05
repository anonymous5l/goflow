all: hello_services/hello_services.so hello_world/hello_world.so

hello_services/hello_services.so: 
	cd hello_services && go build -buildmode=plugin

hello_world/hello_world.so: 
	cd hello_world && go build -buildmode=plugin

clean:
	rm -f hello_world/hello_world.so
	rm -f hello_services/hello_services.so