# How to run

Release build:

```
go build --race -ldflags "-s -w" -o Corpus_Macro_Replacer.exe .\src\
```

or

```
fyne build --src .\src\ -o Corpus_Macro_Replacer-debug.exe
# optionally with --release
```

or (the most proper way):

```
 fyne release  --icon .\corpus_replacer_logo.png -appVersion 0.4 -appBuild 1 -developer "Mat" --certificate asda -password 1234
```

# Development

Icons:
```
go generate
```

