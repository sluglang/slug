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

### Constants

#### `BgBrightColour`

```slug
map slug.term.colour#BgBrightColour
```

map of colour name → ANSI escape code string for bright background colours.

#### `BgColour`

```slug
map slug.term.colour#BgColour
```

map of colour name → ANSI escape code string for background colours.

#### `BrightColour`

```slug
map slug.term.colour#BrightColour
```

map of colour name → ANSI escape code string for bright colours.

#### `Colour`

```slug
map slug.term.colour#Colour
```

map of colour name (PascalCase) → ANSI escape code string for standard colours.

#### `ColourSupport`

```slug
num slug.term.colour#ColourSupport
```

number of colours supported by the current terminal, as reported by `tput colors`.

Common values: 0 (no colour), 8, 16, 256.

#### `Effects`

```slug
map slug.term.colour#Effects
```

map of effect name → ANSI escape code string.

Keys: `Bold`, `BoldOff`, `Dim`, `DimOff`, `Underline`, `UnderlineOff`,
`Blink`, `BlinkOff`, `Reverse`, `ReverseOff`, `Hide`, `HideOff`

#### `Styles`

```slug
map slug.term.colour#Styles
```

raw ANSI code number lists per style name.

Keys include colour names, bright/bg variants, `reset`, and effects.

#### `reset`

```slug
str slug.term.colour#reset
```

ANSI reset escape code — resets all styles to terminal default.

### Functions

#### `bgBlack(str)`
```slug
fn slug.term.colour#bgBlack(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBlue(str)`
```slug
fn slug.term.colour#bgBlue(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBrightBlack(str)`
```slug
fn slug.term.colour#bgBrightBlack(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBrightBlue(str)`
```slug
fn slug.term.colour#bgBrightBlue(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBrightCyan(str)`
```slug
fn slug.term.colour#bgBrightCyan(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBrightGreen(str)`
```slug
fn slug.term.colour#bgBrightGreen(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBrightMagenta(str)`
```slug
fn slug.term.colour#bgBrightMagenta(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBrightRed(str)`
```slug
fn slug.term.colour#bgBrightRed(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBrightWhite(str)`
```slug
fn slug.term.colour#bgBrightWhite(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgBrightYellow(str)`
```slug
fn slug.term.colour#bgBrightYellow(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgCyan(str)`
```slug
fn slug.term.colour#bgCyan(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgGreen(str)`
```slug
fn slug.term.colour#bgGreen(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgHex(str, code)`
```slug
fn slug.term.colour#bgHex(@str str, @str code) -> ?
```


wraps `str` with hex background colour.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `code` | @str  | — |

---

#### `bgHexCode(code)`
```slug
fn slug.term.colour#bgHexCode(@str code) -> ?
```


returns the ANSI escape code string for hex background colour.

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | @str  | — |

---

#### `bgMagenta(str)`
```slug
fn slug.term.colour#bgMagenta(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgRed(str)`
```slug
fn slug.term.colour#bgRed(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgRgb(str, r, g, b)`
```slug
fn slug.term.colour#bgRgb(@str str, @num r, @num g, @num b) -> ?
```


wraps `str` with RGB background colour.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `r` | @num  | — |
| `g` | @num  | — |
| `b` | @num  | — |

---

#### `bgRgbCode(r, g, b)`
```slug
fn slug.term.colour#bgRgbCode(@num r, @num g, @num b) -> ?
```


returns the ANSI escape code string for RGB background colour.

| Parameter | Type | Default |
| --- | --- | --- |
| `r` | @num  | — |
| `g` | @num  | — |
| `b` | @num  | — |

---

#### `bgVga(str, code)`
```slug
fn slug.term.colour#bgVga(@str str, @num code) -> ?
```


wraps `str` with VGA background colour `code` (0–255).

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `code` | @num  | — |

---

#### `bgVgaCode(code)`
```slug
fn slug.term.colour#bgVgaCode(@num code) -> ?
```


returns the ANSI escape code string for VGA background colour index `code` (0–255).

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | @num  | — |

---

#### `bgWhite(str)`
```slug
fn slug.term.colour#bgWhite(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `bgXkcd(str, name)`
```slug
fn slug.term.colour#bgXkcd(@str str, @sym name) -> ?
```


wraps `str` with an xkcd background colour.

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `name` | @sym  | — |

---

#### `bgXkcdCode(name)`
```slug
fn slug.term.colour#bgXkcdCode(@sym name) -> ?
```


returns the ANSI escape code string for an xkcd background colour.

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | @sym  | — |

---

#### `bgYellow(str)`
```slug
fn slug.term.colour#bgYellow(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `black(str)`
```slug
fn slug.term.colour#black(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `blue(str)`
```slug
fn slug.term.colour#blue(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `brightBlack(str)`
```slug
fn slug.term.colour#brightBlack(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `brightBlue(str)`
```slug
fn slug.term.colour#brightBlue(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `brightCyan(str)`
```slug
fn slug.term.colour#brightCyan(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `brightGreen(str)`
```slug
fn slug.term.colour#brightGreen(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `brightMagenta(str)`
```slug
fn slug.term.colour#brightMagenta(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `brightRed(str)`
```slug
fn slug.term.colour#brightRed(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `brightWhite(str)`
```slug
fn slug.term.colour#brightWhite(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `brightYellow(str)`
```slug
fn slug.term.colour#brightYellow(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `cyan(str)`
```slug
fn slug.term.colour#cyan(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `effectBlink(str)`
```slug
fn slug.term.colour#effectBlink(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `effectBold(str)`
```slug
fn slug.term.colour#effectBold(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `effectDim(str)`
```slug
fn slug.term.colour#effectDim(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `effectHide(str)`
```slug
fn slug.term.colour#effectHide(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `effectReverse(str)`
```slug
fn slug.term.colour#effectReverse(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `effectUnderline(str)`
```slug
fn slug.term.colour#effectUnderline(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `green(str)`
```slug
fn slug.term.colour#green(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `hex(str, code)`
```slug
fn slug.term.colour#hex(@str str, @str code) -> ?
```


wraps `str` with hex colour (e.g. `hex("Hello", "#ff5733")`).

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `code` | @str  | — |

---

#### `hexCode(code)`
```slug
fn slug.term.colour#hexCode(@str code) -> ?
```


returns the ANSI escape code string for hex colour `code` (e.g. `"#ff5733"`).

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | @str  | — |

---

#### `magenta(str)`
```slug
fn slug.term.colour#magenta(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `red(str)`
```slug
fn slug.term.colour#red(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `rgb(str, r, g, b)`
```slug
fn slug.term.colour#rgb(@str str, @num r, @num g, @num b) -> ?
```


wraps `str` with RGB colour `r`, `g`, `b` (0–255 each).

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `r` | @num  | — |
| `g` | @num  | — |
| `b` | @num  | — |

---

#### `rgbCode(r, g, b)`
```slug
fn slug.term.colour#rgbCode(@num r, @num g, @num b) -> ?
```


returns the ANSI escape code string for RGB colour.

| Parameter | Type | Default |
| --- | --- | --- |
| `r` | @num  | — |
| `g` | @num  | — |
| `b` | @num  | — |

---

#### `vga(str, code)`
```slug
fn slug.term.colour#vga(@str str, @num code) -> ?
```


wraps `str` with VGA colour `code` (0–255).

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `code` | @num  | — |

---

#### `vgaCode(code)`
```slug
fn slug.term.colour#vgaCode(@num code) -> ?
```


returns the ANSI escape code string for VGA colour index `code` (0–255).

| Parameter | Type | Default |
| --- | --- | --- |
| `code` | @num  | — |

---

#### `white(str)`
```slug
fn slug.term.colour#white(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |

---

#### `xkcd(str, name)`
```slug
fn slug.term.colour#xkcd(@str str, @sym name) -> ?
```


wraps `str` with an xkcd colour (e.g. `xkcd("Hello", :electricLime)`).

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `name` | @sym  | — |

---

#### `xkcdCode(name)`
```slug
fn slug.term.colour#xkcdCode(@sym name) -> ?
```


returns the ANSI escape code string for an xkcd colour name (e.g. `:electricLime`).

| Parameter | Type | Default |
| --- | --- | --- |
| `name` | @sym  | — |

---

#### `yellow(str)`
```slug
fn slug.term.colour#yellow(@str str) -> ?
```

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |