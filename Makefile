.PHONY: dev test clean prod

dev:
	make -C create_titleid dev
	make -C transcodelauncher dev

test:
	make -C create_titleid test
	make -C transcodelauncher test

prod:
	mkdir build
	make -C create_titleid prod
	make -C transcodelauncher prod

clean:
	make -C create_titleid clean
	make -C transcodelauncher clean
	rm -rf build