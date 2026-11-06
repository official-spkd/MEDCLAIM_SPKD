# Delegasi ke backend/Makefile supaya make bisa dipanggil dari folder frontend.
# Contoh: make up / make fe / make fe-demo / make down / make up I=2
I ?= 1

%:
	$(MAKE) -C ../backend $@ I=$(I)

.PHONY: up
up:
	$(MAKE) -C ../backend up I=$(I)
