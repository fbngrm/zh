.PHONY: clean-log
clean-log:
	rm gen/$(source)/log
	touch gen/$(source)/log

.PHONY: clean-blacklist
clean-blacklist:
	rm gen/$(source)/blacklist
	touch gen/$(source)/blacklist

.PHONY: clean
clean: clean-blacklist clean-log

.PHONY: generate
generate:
	go run cmd/anki-gen/main.go \
		-i ./lib/$(source)/$(file) \
		-t ./templates/$(source).tmpl \
		-e ./gen/$(source)/log \
		-e ./gen/$(source)/blacklist \
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
	cd ./gen/$(source); ffmpeg -i ./$(file) -ss 00:00:02 -t 00:00:03 noisesample.wav
	cd ./gen/$(source); sox ./noisesample.wav -n noiseprof ./noise_profile_file

.PHONY: remove-noise
audio_dir=./gen/$(source)/audio
clean_dir=$(audio_dir)/clean
backup_dir=$(audio_dir)//original
remove-noise:
	mkdir -p $(clean_dir)
	mkdir -p $(backup_dir)
	# cp $(audio_dir)/*.mp3 ./gen/$(source)/audio/original
	# cd $(audio_dir); ls -r -1 *.mp3 | xargs -L1 -I{} sox {} {}_cleaned.mp3  noisered ../../noise_profile_file 0.30
	# mv $(audio_dir)/*_cleaned.mp3 $(clean_dir)
	# rm $(audio_dir)/*.mp3
	for i in $(clean_dir)/*; do ffmpeg -ss 0.75 -i "$$i" "$${i%.*}_shortened.mp3"; done
	find $(clean_dir) -name "*clean.mp3" -exec rename -v ".mp3_cleaned_shortened" "" {} ";"
