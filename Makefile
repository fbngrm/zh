source_lib_dir=./lib/$(source)
audio_lib_dir=./lib/$(source)/audio

.PHONY: clean-log
clean-log:
	rm $(source_lib_dir)/log
	touch $(source_lib_dir)/log

.PHONY: clean-blacklist
clean-blacklist:
	rm $(source_lib_dir)/blacklist
	touch $(source_lib_dir)/blacklist

.PHONY: clean
clean: clean-blacklist clean-log

.PHONY: generate
generate:
	go run cmd/anki-gen/main.go \
		-i $(source_lib_dir)/$(file) \
		-t ./templates/$(source).tmpl \
		-e $(source_lib_dir)/log \
		-b $(source_lib_dir)/blacklist \
		-d $(source)

.PHONY: record
record:
	go run cmd/rec/main.go \
		-d $(source) \
		-i ./gen/$(source)/$(file).yaml \
		-c

.PHONY: force-record
force-record:
	go run cmd/rec/main.go \
		-d $(source) \
		-i ./gen/$(source)/$(file).yaml \
		-c \
		-f

.PHONY: generate-noise-profile
generate-noise-profile:
	cd $(audio_lib_dir); ffmpeg -i ./$(file) -ss 00:00:02 -t 00:00:03 ./noisesample.wav
	cd $(audio_lib_dir); sox ./noisesample.wav -n noiseprof ./noise_profile_file

.PHONY: remove-noise
audio_gen_dir=./gen/$(source)/audio
clean_dir=$(audio_gen_dir)/clean
backup_dir=$(audio_gen_dir)//original
remove-noise:
	mkdir -p $(clean_dir)
	mkdir -p $(backup_dir)
	cp $(audio_gen_dir)/*.mp3 ./gen/$(source)/audio/original
	cd $(audio_gen_dir); ls -r -1 *.mp3 | xargs -L1 -I{} sox {} {}_cleaned.mp3  noisered ../../noise_profile_file 0.30
	for i in $(audio_gen_dir)/*.mp3; do ffmpeg -ss 0.75 -i "$$i" "$${i%.*}_shortened.mp3"; done
	mv $(audio_gen_dir)/*.mp3_cleaned_shortened.mp3 $(clean_dir)
	find $(clean_dir) -name "*shortened.mp3" -exec rename -v ".mp3_cleaned_shortened" "" {} ";"
	rm $(audio_gen_dir)/*.mp3
