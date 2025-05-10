# How to build and run

Follow Go Fyne instructions for setting up dependencies.

Release build:

```
cd ./src
go build --race -ldflags "-s -w" -o Corpus_Macro_Replacer.exe .\cmd\gui\
.\Corpus_Macro_Replacer.exe

go build --race -ldflags "-s -w" -o RemoveVariablesFromElements.exe .\cmd\removeVariablesFromElements\
.\RemoveVariablesFromElements.exe -h
```

or (not tested)

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

