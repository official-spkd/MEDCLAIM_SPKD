# Delegasi ke backend/Makefile supaya `make up` bisa dari root repo.
# Contoh: make up / make down / make logs / make up I=2
I ?= 1

%:
	$(MAKE) -C backend $@ I=$(I)

.PHONY: up
up:
	$(MAKE) -C backend up I=$(I)
