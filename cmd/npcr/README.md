Dialog
Ch
En
Audio

Sentences
Ch
En
Audio

Words
Ch
En
Audio

Single Hanzi
Ch
En

==> single hanzi, no audio card

---

## note types

### ch-en-audio

#### fields
- Chinese
- Pinyin
- English
- Audio
- Components
- Examples

#### card types

1. Chinese on front

Front:
```
<div class="front" lang="zh-Hans">
<span class="medium japanese">{{Chinese}}</span>
</div>
```

Back:
```
<div class="back" lang="zh-Hans">
<span class="medium japanese">{{Chinese}}</span>
<hr>
<span class="small">{{English}}</span>
<hr>
<span class="tiny japanese">{{Audio}}</span>
<span class="tiny">{{Components}}</span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny">{{Examples}}</span>
</div>
```

2. English on front

Front:
```
<div class="front" lang="en">
  <span class="medium">{{English}}</span>
</div>
```

Back:
```
<div class="back" lang="zh-Hans">
<span class="medium japanese">{{Chinese}}</span>
<hr>
<span class="small">{{English}}</span>
<hr>
<span class="tiny japanese">{{Audio}}</span>
<span class="tiny">{{Components}}</span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny">{{Examples}}</span>
</div>
```

3. Audio on front

Front:
```
<div class="front">
  <span class="small">{{Audio}}</span>
</div>
```

Back:
```
<div class="back" lang="zh-Hans">
<span class="medium japanese">{{Chinese}}</span>
<hr>
<span class="small">{{English}}</span>
<hr>
<span class="tiny japanese">{{Audio}}</span>
<span class="tiny">{{Components}}</span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny">{{Examples}}</span>
</div>
```

### ch-en

#### fields

- Chinese
- Pinyin
- English
- Audio
- Components
- Examples

#### card types

1. Chinese on front

Front:
```
<div class="front" lang="zh-Hans">
  <span class="medium japanese">{{Chinese}}</span>
</div>
```

Back:
```
<div class="back" lang="zh-Hans">
<span class="medium japanese">{{Chinese}}</span>
<hr>
<span class="small">{{English}}</span>
<hr>
<span class="tiny japanese">{{Audio}}</span>
<span class="tiny">{{Components}}</span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny">{{Examples}}</span>
</div>
```

1. English on front

Front:
```
<div class="front" lang="en">
  <span class="medium">{{English}}</span>
</div>
```

Back:
```
<div class="back" lang="zh-Hans">
<span class="medium japanese">{{Chinese}}</span>
<hr>
<span class="small">{{English}}</span>
<hr>
<span class="tiny japanese">{{Audio}}</span>
<span class="tiny">{{Components}}</span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny"></span>
<span class="tiny">{{Examples}}</span>
</div>
```


#### CSS
```
div.front, div.back {
	text-align: center;
	font-family: sans-serif;
	font-size: 10px; /* line height is based on this size in Anki for some reason, so start with the smallest size used */
}
span.tiny {font-size: 10px;}
span.small {font-size: 16px;}
span.medium {font-size: 24px;}
span.large {font-size: 36px;}
span.italic {font-style: italic;}
```
