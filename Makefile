source_lib_dir=./data/lib/$(source)
audio_lib_dir=./data/lib/$(source)/audio

# generate data

.PHONY: clean-ignore
clean-ignore:
	rm $(source_lib_dir)/ignore
	touch $(source_lib_dir)/ignore

.PHONY: clean-blacklist
clean-blacklist:
	rm $(source_lib_dir)/blacklist
	touch $(source_lib_dir)/blacklist

.PHONY: clean
clean: clean-blacklist clean-ignore

.PHONY: generate
generate:
	go run cmd/anki-gen/main.go \
		-i $(source_lib_dir)/$(file) \
		-t ./templates/$(source).tmpl \
		-e $(source_lib_dir)/ignore \
		-b $(source_lib_dir)/blacklist \
		-d $(source)

.PHONY: record
record:
	go run cmd/rec/main.go \
		-d $(source) \
		-i ./data/gen/$(source)/$(file).yaml \
		-c

.PHONY: force-record
force-record:
	go run cmd/rec/main.go \
		-d $(source) \
		-i ./data/gen/$(source)/$(file).yaml \
		-c \
		-f

# generate audio

.PHONY: generate-noise-profile
generate-noise-profile:
	cd $(audio_lib_dir); ffmpeg -i ./$(file) -ss 00:00:02 -t 00:00:03 ./noisesample.wav
	cd $(audio_lib_dir); sox ./noisesample.wav -n noiseprof ./noise_profile_file

.PHONY: copy-audio
audio_gen_dir=./data/gen/$(source)/audio
noise_profile_file=../../../data/lib/$(source)/audio/noise_profile_file
copy-audio:
	mkdir -p $(audio_gen_dir)
	cp $(audio_lib_dir)/*.mp3 $(audio_gen_dir)

.PHONY: remove-noise
remove-noise:
	cd $(audio_gen_dir); ls -r -1 *.mp3 | xargs -L1 -I{} sox {} {}_cleaned.mp3  noisered $(noise_profile_file) 0.30

.PHONY: shorten-audio
shorten-audio:
	cd $(audio_gen_dir); for i in *mp3_cleaned.mp3; do ffmpeg -ss 0.75 -i "$$i" "$${i%.*}_shortened.mp3"; done

.PHONY: rename-audio
rename-audio:
	cd $(audio_gen_dir); rm ./*cleaned.mp3
	cd $(audio_gen_dir); find . -name "*shortened.mp3" -exec rename -v ".mp3_cleaned_shortened" "" {} ";"

.PHONY: clean-audio
clean-audio: copy-audio remove-noise shorten-audio rename-audio

# generate anki

.PHONY: copy-anki-audio
anki_audio_dir="/home/f/.local/share/Anki2/User 1/collection.media/"
copy-anki-audio:
	cp $(audio_gen_dir)/*.mp3 $(anki_audio_dir)

.PHONY: generate-anki-deck
generate-anki-deck:
	@printf "Checking for changes in blacklist\n"
	@CHANGES=$$(git status -s --porcelain -- ./lib/$(source)/blacklist); \
	if [ ! -z "$${CHANGES}" ]; \
	then \
		echo "Please re-generate after blacklist was changed: make generate source=$(source) file=$(file)"; \
		exit 1; \
	fi
	@printf "Blacklist was not changed\n"
	apy add-from-file ./data/gen/$(source)/$(file).md
	@printf "Done. Don't forget to sync: make sync-anki\n"

.PHONY: sync-anki
sync-anki: copy-anki-audio
	apy check-media
	apy sync
