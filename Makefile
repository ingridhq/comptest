SUBDIRS := $(wildcard _example/*/)

all: $(SUBDIRS)
$(SUBDIRS):
	@$(MAKE) -C $@ $(MAKECMDGOALS)

.PHONY: all $(SUBDIRS)
