
signaling:
	$(MAKE) -C ./signaling/server push

turnserver:
	$(MAKE) -C ./turnserver push

sample:
	$(MAKE) -C ./signaling/sample run
