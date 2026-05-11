---
title: colour (slug.term)
---

## slug.term.colour

slug.term.colour — terminal colour and text effect library

Wraps strings with ANSI escape codes for coloured terminal output.
Inspired by the Python [colorist](https://jakob-bagterp.github.io/colorist-for-python/)
library. See also the [ANSI escape code reference](https://en.wikipedia.org/wiki/ANSI_escape_code#Colors).

## Quick start

```slug
val { red, green, effectBold, bgBlue } = import("slug.term.colour")

"Hello" /> red /> println
"Success" /> green /> effectBold /> println
"Warning" /> bgBlue /> println
```

## Colour functions

Each named colour is available as a function taking a string and returning
the ANSI-wrapped string. Text colours: `black`, `red`, `green`, `yellow`,
`blue`, `magenta`, `cyan`, `white` and bright variants (`brightRed` etc.).
Background variants are prefixed with `bg` (`bgRed`, `bgBrightBlue` etc.).

## Extended colour support

For terminals supporting 256 colours or true colour:
- **VGA (0–255):** `vga(str, code)`, `bgVga(str, code)`
- **RGB:** `rgb(str, r, g, b)`, `bgRgb(str, r, g, b)`
- **Hex:** `hex(str, "#ff5733")`, `bgHex(str, "#ff5733")`
- **xkcd:** `xkcd(str, :electricLime)`, `bgXkcd(str, :neonBlue)`

## Effects

`effectBold`, `effectDim`, `effectUnderline`, `effectBlink`,
`effectReverse`, `effectHide`

## Colour support detection

`ColourSupport` is set at module load time by running `tput colors`.
Use it to gracefully degrade in environments without colour support.

## Pre-built maps

`Colour`, `BrightColour`, `BgColour`, `BgBrightColour` are maps of
colour name → ANSI code string for programmatic use.
`Effects` is a map of effect name → ANSI code string.
`Styles` is the raw map of style name → ANSI code number list.

### TOC

- [BgBrightColour](#bgbrightcolour)
- [BgColour](#bgcolour)
- [BrightColour](#brightcolour)
- [Colour](#colour)
- [ColourSupport](#coloursupport)
- [Effects](#effects)
- [Styles](#styles)
- [reset](#reset)
- [`bgBlack(str)`](#bgblackstr)
- [`bgBlue(str)`](#bgbluestr)
- [`bgBrightBlack(str)`](#bgbrightblackstr)
- [`bgBrightBlue(str)`](#bgbrightbluestr)
- [`bgBrightCyan(str)`](#bgbrightcyanstr)
- [`bgBrightGreen(str)`](#bgbrightgreenstr)
- [`bgBrightMagenta(str)`](#bgbrightmagentastr)
- [`bgBrightRed(str)`](#bgbrightredstr)
- [`bgBrightWhite(str)`](#bgbrightwhitestr)
- [`bgBrightYellow(str)`](#bgbrightyellowstr)
- [`bgCyan(str)`](#bgcyanstr)
- [`bgGreen(str)`](#bggreenstr)
- [`bgHex(str, code)`](#bghexstr-code)
- [`bgHexCode(code)`](#bghexcodecode)
- [`bgMagenta(str)`](#bgmagentastr)
- [`bgRed(str)`](#bgredstr)
- [`bgRgb(str, r, g, b)`](#bgrgbstr-r-g-b)
- [`bgRgbCode(r, g, b)`](#bgrgbcoder-g-b)
- [`bgVga(str, code)`](#bgvgastr-code)
- [`bgVgaCode(code)`](#bgvgacodecode)
- [`bgWhite(str)`](#bgwhitestr)
- [`bgXkcd(str, name)`](#bgxkcdstr-name)
- [`bgXkcdCode(name)`](#bgxkcdcodename)
- [`bgYellow(str)`](#bgyellowstr)
- [`black(str)`](#blackstr)
- [`blue(str)`](#bluestr)
- [`brightBlack(str)`](#brightblackstr)
- [`brightBlue(str)`](#brightbluestr)
- [`brightCyan(str)`](#brightcyanstr)
- [`brightGreen(str)`](#brightgreenstr)
- [`brightMagenta(str)`](#brightmagentastr)
- [`brightRed(str)`](#brightredstr)
- [`brightWhite(str)`](#brightwhitestr)
- [`brightYellow(str)`](#brightyellowstr)
- [`cyan(str)`](#cyanstr)
- [`effectBlink(str)`](#effectblinkstr)
- [`effectBold(str)`](#effectboldstr)
- [`effectDim(str)`](#effectdimstr)
- [`effectHide(str)`](#effecthidestr)
- [`effectReverse(str)`](#effectreversestr)
- [`effectUnderline(str)`](#effectunderlinestr)
- [`green(str)`](#greenstr)
- [`hex(str, code)`](#hexstr-code)
- [`hexCode(code)`](#hexcodecode)
- [`magenta(str)`](#magentastr)
- [`red(str)`](#redstr)
- [`rgb(str, r, g, b)`](#rgbstr-r-g-b)
- [`rgbCode(r, g, b)`](#rgbcoder-g-b)
- [`vga(str, code)`](#vgastr-code)
- [`vgaCode(code)`](#vgacodecode)
- [`white(str)`](#whitestr)
- [`xkcd(str, name)`](#xkcdstr-name)
- [`xkcdCode(name)`](#xkcdcodename)
- [`yellow(str)`](#yellowstr)

### Constants

#### `BgBrightColour`

```slug
map slug.term.colour#BgBrightColour
```

#### `BgColour`

```slug
map slug.term.colour#BgColour
```

#### `BrightColour`

```slug
map slug.term.colour#BrightColour
```

#### `Colour`

```slug
map slug.term.colour#Colour
```

#### `ColourSupport`

```slug
num slug.term.colour#ColourSupport
```

#### `Effects`

```slug
map slug.term.colour#Effects
```

#### `Styles`

```slug
map slug.term.colour#Styles
```

#### `reset`

```slug
str slug.term.colour#reset
```

### Functions

#### `bgBlack(str)`
```slug
fn slug.term.colour#bgBlack(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBlue(str)`
```slug
fn slug.term.colour#bgBlue(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBrightBlack(str)`
```slug
fn slug.term.colour#bgBrightBlack(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBrightBlue(str)`
```slug
fn slug.term.colour#bgBrightBlue(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBrightCyan(str)`
```slug
fn slug.term.colour#bgBrightCyan(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBrightGreen(str)`
```slug
fn slug.term.colour#bgBrightGreen(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBrightMagenta(str)`
```slug
fn slug.term.colour#bgBrightMagenta(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBrightRed(str)`
```slug
fn slug.term.colour#bgBrightRed(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBrightWhite(str)`
```slug
fn slug.term.colour#bgBrightWhite(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgBrightYellow(str)`
```slug
fn slug.term.colour#bgBrightYellow(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgCyan(str)`
```slug
fn slug.term.colour#bgCyan(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgGreen(str)`
```slug
fn slug.term.colour#bgGreen(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgHex(str, code)`
```slug
fn slug.term.colour#bgHex(str:str, code:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `code` | str | — |

---

#### `bgHexCode(code)`
```slug
fn slug.term.colour#bgHexCode(code:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | str | — |

---

#### `bgMagenta(str)`
```slug
fn slug.term.colour#bgMagenta(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgRed(str)`
```slug
fn slug.term.colour#bgRed(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgRgb(str, r, g, b)`
```slug
fn slug.term.colour#bgRgb(str:str, r:num, g:num, b:num):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `r` | num | — |
| `g` | num | — |
| `b` | num | — |

---

#### `bgRgbCode(r, g, b)`
```slug
fn slug.term.colour#bgRgbCode(r:num, g:num, b:num):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` | num | — |
| `g` | num | — |
| `b` | num | — |

---

#### `bgVga(str, code)`
```slug
fn slug.term.colour#bgVga(str:str, code:num):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `code` | num | — |

---

#### `bgVgaCode(code)`
```slug
fn slug.term.colour#bgVgaCode(code:num):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | num | — |

---

#### `bgWhite(str)`
```slug
fn slug.term.colour#bgWhite(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `bgXkcd(str, name)`
```slug
fn slug.term.colour#bgXkcd(str:str, name:sym):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `name` | sym | — |

---

#### `bgXkcdCode(name)`
```slug
fn slug.term.colour#bgXkcdCode(name:sym):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | sym | — |

---

#### `bgYellow(str)`
```slug
fn slug.term.colour#bgYellow(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `black(str)`
```slug
fn slug.term.colour#black(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `blue(str)`
```slug
fn slug.term.colour#blue(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `brightBlack(str)`
```slug
fn slug.term.colour#brightBlack(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `brightBlue(str)`
```slug
fn slug.term.colour#brightBlue(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `brightCyan(str)`
```slug
fn slug.term.colour#brightCyan(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `brightGreen(str)`
```slug
fn slug.term.colour#brightGreen(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `brightMagenta(str)`
```slug
fn slug.term.colour#brightMagenta(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `brightRed(str)`
```slug
fn slug.term.colour#brightRed(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `brightWhite(str)`
```slug
fn slug.term.colour#brightWhite(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `brightYellow(str)`
```slug
fn slug.term.colour#brightYellow(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `cyan(str)`
```slug
fn slug.term.colour#cyan(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `effectBlink(str)`
```slug
fn slug.term.colour#effectBlink(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `effectBold(str)`
```slug
fn slug.term.colour#effectBold(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `effectDim(str)`
```slug
fn slug.term.colour#effectDim(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `effectHide(str)`
```slug
fn slug.term.colour#effectHide(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `effectReverse(str)`
```slug
fn slug.term.colour#effectReverse(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `effectUnderline(str)`
```slug
fn slug.term.colour#effectUnderline(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `green(str)`
```slug
fn slug.term.colour#green(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `hex(str, code)`
```slug
fn slug.term.colour#hex(str:str, code:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `code` | str | — |

---

#### `hexCode(code)`
```slug
fn slug.term.colour#hexCode(code:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | str | — |

---

#### `magenta(str)`
```slug
fn slug.term.colour#magenta(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `red(str)`
```slug
fn slug.term.colour#red(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `rgb(str, r, g, b)`
```slug
fn slug.term.colour#rgb(str:str, r:num, g:num, b:num):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `r` | num | — |
| `g` | num | — |
| `b` | num | — |

---

#### `rgbCode(r, g, b)`
```slug
fn slug.term.colour#rgbCode(r:num, g:num, b:num):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `r` | num | — |
| `g` | num | — |
| `b` | num | — |

---

#### `vga(str, code)`
```slug
fn slug.term.colour#vga(str:str, code:num):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `code` | num | — |

---

#### `vgaCode(code)`
```slug
fn slug.term.colour#vgaCode(code:num):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | num | — |

---

#### `white(str)`
```slug
fn slug.term.colour#white(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |

---

#### `xkcd(str, name)`
```slug
fn slug.term.colour#xkcd(str:str, name:sym):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |
| `name` | sym | — |

---

#### `xkcdCode(name)`
```slug
fn slug.term.colour#xkcdCode(name:sym):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | sym | — |

---

#### `yellow(str)`
```slug
fn slug.term.colour#yellow(str:str):any
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | str | — |