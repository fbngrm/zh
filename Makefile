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

.PHONY: download-audio
download-audio:
	go run cmd/rec/main.go \
		-d $(source) \
		-i ./data/gen/$(source)/$(file).yaml \
		-c -a

.PHONY: force-download-audio
force-download-audio:
	go run cmd/rec/main.go \
		-d $(source) \
		-i ./data/gen/$(source)/$(file).yaml \
		-c \
        -a \
		-f

# generate audio
.PHONY: generate-noise-profile
generate-noise-profile:
	cd $(audio_lib_dir); ffmpeg -i ./$(file) -ss 00:00:02 -t 00:00:03 ./noisesample.wav
	cd $(audio_lib_dir); sox ./noisesample.wav -n noiseprof ./noise_profile_file

audio_gen_dir=./data/gen/$(source)/audio
noise_profile_file=../../../data/lib/$(source)/audio/noise_profile_file
.PHONY: copy-audio
copy-audio:
	mkdir -p $(audio_gen_dir)
	cp $(audio_lib_dir)/*.mp3 $(audio_gen_dir)

.PHONY: concat-audio
audio_sub_dir=./data/gen/$(source)/audio/$(subdir)
out_dir=/home/f/data/music/zh/$(source)/$(subdir)
silence=../../../silence.mp3
concat-audio:
	echo $(audio_sub_dir)
	mkdir -p $(out_dir)
	cd $(audio_sub_dir); for i in *.mp3; do ffmpeg -i "concat:$$i|$(silence)|$$i|$(silence)|$$i|$(silence)|$$i|$(silence)|$$i|$(silence)|$(silence)" -acodec copy $(out_dir)/"$${i%.*}_concat.mp3"; done

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

anki_audio_dir="/home/f/.local/share/Anki2/User 1/collection.media/"
.PHONY: copy-anki-audio
copy-anki-audio:
	cd $(audio_gen_dir)
	$(shell find . -type f -name '*.mp3' -exec cp {} $(anki_audio_dir) \;)

.PHONY: generate-anki-deck
generate-anki-deck:
	@printf "Checking for changes in blacklist\n"
	@CHANGES=$$(git status -s --porcelain -- ./data/lib/$(source)/blacklist); \
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
