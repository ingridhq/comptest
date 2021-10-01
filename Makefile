TOPTARGETS := all clean
SUBDIRS := $(wildcard _example/*/)

$(TOPTARGETS): $(SUBDIRS)
$(SUBDIRS):
	@$(MAKE) -C $@ $(MAKECMDGOALS)

.PHONY: $(TOPTARGETS) $(SUBDIRS)