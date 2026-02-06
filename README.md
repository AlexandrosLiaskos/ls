# ls

Minimal, styled directory listing — Go + [Charm Lipgloss](https://github.com/charmbracelet/lipgloss).

Green-on-black output, dirs first, human-readable sizes.

## Usage

```
ls              # list files and dirs
ls -a           # include hidden
ls -f           # files only
ls path/to/dir  # list specific directory
```

## Install

```
go install github.com/AlexandrosLiaskos/ls@latest
```

Or build from source:

```
go build -o ls.exe .
```

## License

MIT
