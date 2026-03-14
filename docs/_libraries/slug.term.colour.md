---
title: colour (slug.term)
---

## slug.term.colour

A terminal colours library inspired by the python `colorist` library.

See also:
 - Colorist https://jakob-bagterp.github.io/colorist-for-python/
 - the wikipedia page https://en.wikipedia.org/wiki/ANSI_escape_code#Colors

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

The number of supported colours reported by `tput colors`

#### `Effects`

```slug
map slug.term.colour#Effects
```

A map of effects control codes.

The following keys are defined: bold, boldOff, dim, dimOff, underline, underlineOff,
blink, blinkOff, reverse, reverseOff, hide, hideOff

#### `Styles`

```slug
map slug.term.colour#Styles
```

Styles map, contains a list containing the ANSI escape code numbers for each style,
map keys by colour name plus modifiers, e.g. black, brightBlack, bgBlack, bgBrightBlack
other keys include `reset` and the effects, e.g. blink and blinkOff

#### `reset`

```slug
str slug.term.colour#reset
```

A terminal control code the reset styles back to the default

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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `code` | @str  | — |

---

#### `bgHexCode(code)`
```slug
fn slug.term.colour#bgHexCode(@str code) -> ?
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `code` | @num  | — |

---

#### `bgVgaCode(code)`
```slug
fn slug.term.colour#bgVgaCode(@num code) -> ?
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `name` | @sym  | — |

---

#### `bgXkcdCode(name)`
```slug
fn slug.term.colour#bgXkcdCode(@sym name) -> ?
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `code` | @str  | — |

---

#### `hexCode(code)`
```slug
fn slug.term.colour#hexCode(@str code) -> ?
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `code` | @num  | — |

---

#### `vgaCode(code)`
```slug
fn slug.term.colour#vgaCode(@num code) -> ?
```

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

| Parameter | Type | Default |
| --- | --- | --- |
| `str` | @str  | — |
| `name` | @sym  | — |

---

#### `xkcdCode(name)`
```slug
fn slug.term.colour#xkcdCode(@sym name) -> ?
```

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