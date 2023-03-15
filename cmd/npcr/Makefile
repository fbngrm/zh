data_dir=./data/$(deck)/$(lesson)
audio_dir=./data/$(deck)/$(lesson)/audio

.PHONY: gen
gen:
	rm -r -f $(data_dir)/output/
	mkdir $(data_dir)/output/
	go run cmd/main.go -l $(lesson) -d $(deck)

.PHONY: add
add:
	apy add-from-file $(data_dir)/output/cards.md
	@printf "Done. Don't forget to sync: make sync\n"

anki_audio_dir="/home/f/.local/share/Anki2/User 1/collection.media/"
.PHONY: cp-audio
cp-audio:
	cd $(audio_dir)
	$(shell find . -type f -name '*.mp3' -exec cp {} $(anki_audio_dir) \;)

.PHONY: sync
sync: gen add cp-audio
	apy check-media
	apy sync
	@echo "don't forget to commit ignore file!"

.PHONY: new
new:
	mkdir -p $(data_dir)
	touch  $(data_dir)/dialog
	touch  $(data_dir)/sentences
	touch  $(data_dir)/words

# .PHONY: commit-ignore
# ignore_path=$(data_dir)/../ignore
# prev_ignore_path=./data/prev_ignore_commit
# commit-ignore:
# 	$(shell git add $(ignore_path))
# 	$(shell git commit -m "commit ignore for lesson $(lesson)")
# 	$(shell git rev-parse HEAD > $(prev_ignore_path))

# .PHONY: reset-ignore
# reset-ignore:
# 	@echo $(prev_ignore_path)
# 	$(shell git revert $(shell cat $(prev_ignore_path)))
# 	rm $(prev_ignore_path)

# .PHONY: reset-files
# reset-files:
# 	rm $(data_dir)/cards.md $(data_dir)/dialog*

# .PHONY: reset
# reset: reset-ignore reset-files

.PHONY: audio
out_dir=../../../../data//$(deck)/$(lesson)/audio/concat
silence=../../../../data/silence_64kb.mp3
audio:
	mkdir -p $(out_dir)
	cd $(audio_dir); for i in *.mp3; do ffmpeg -i "$$i" -filter:a "atempo=0.85" /tmp/"$${i%.*}_slow.mp3"; done
	cd $(audio_dir); for i in *.mp3; do ffmpeg -i "concat:$$i|$(silence)|/tmp/$${i%.*}_slow.mp3|$(silence)|$$i|$(silence)|/tmp/$${i%.*}_slow.mp3|$(silence)|$$i|$(silence)|$(silence)" -acodec copy $(out_dir)/"$${i%.*}_concat.mp3"; done
