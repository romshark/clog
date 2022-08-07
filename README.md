# Clog - Cool Log 😎🪵

Pipe structured JSON logging outputs to Clog and let it pretty-print it for you 👨🏻‍🎨

Clog reads JSON logs from Stdin, parses and pretty-prints them according to the configuration. A configuration file can be provided with `-c`, the default config file path is `clog.toml`. If the config file isn't found then the defaults apply.

`clog -i` makes clog ignore OS interrupt signals, this can be useful to allow the process that pipes to clog and handles interrupt signals to exit gracefuly. If you run `clog -i` without piping anything to its input you can just press `CTRL+D` (EOF) and exit clog.

## Installation

```bash
go install -v github.com/romshark/clog@latest
```

Also make sure `go/bin` is in your `PATH`.

## Example

`cat example.txt | clog`

## Styling

Supported style attributes:

```
bold
faint
italic
underline
blinkslow
blinkrapid
reversevideo
concealed
crossedout
fg-black
fg-red
fg-green
fg-yellow
fg-blue
fg-magenta
fg-cyan
fg-white
fg-hiblack
fg-hired
fg-higreen
fg-hiyellow
fg-hiblue
fg-himagenta
fg-hicyan
fg-hiwhite
bg-black
bg-red
bg-green
bg-yellow
bg-blue
bg-magenta
bg-cyan
bg-white
bg-hiblack
bg-hired
bg-higreen
bg-hiyellow
bg-hiblue
bg-himagenta
bg-hicyan
bg-hiwhite
```