# Roundup brand files

The Roundup mark is three rounded shapes gathered into an R: the tall pill is the stem, the large disc the bowl, the small disc the leg. People rounded up into one plan. The source of truth is the "Roundup" design system; these files reproduce its `Logo` component for places that can't render it, such as this repo's README.

| File | What it is | Use it for |
| --- | --- | --- |
| `banner-dark.svg` / `banner-light.svg` | Lockup, tagline and the brand shapes on a 1280×360 panel | The README header (`<picture>` picks one by theme) |
| `roundup-logo-dark.svg` / `roundup-logo-light.svg` | Lockup (mark + wordmark) on a transparent background, with clear space | Docs, slides, anywhere that needs the full logo |
| `roundup-mark.svg` / `roundup-mark-latte.svg` | The mark alone, Mocha and Latte colours | Favicons, avatars, small spaces |
| `roundup-mark-mono.svg` | The mark in `currentColor` | One-colour use, on an accent fill or a photo |
| `divider.svg` | A mauve → peach → lavender rule | Section breaks in Markdown |

## Colours

| Theme | Stem (`logo-stem`) | Bowl (`logo-bowl`) | Leg (`logo-leg`) | Wordmark (`fg`) |
| --- | --- | --- | --- | --- |
| Mocha (dark) | `#cba6f7` | `#fab387` | `#b4befe` | `#cdd6f4` |
| Latte (light) | `#8839ef` | `#fe640b` | `#7287fd` | `#4c4f69` |

## Rules

- **Lockup geometry:** the mark at size *s*, a gap of round(0.28 *s*), and "Roundup" in Onest Bold at round(0.75 *s*) with −0.03em letter-spacing, centred on the mark.
- **Clear space:** a quarter of the mark's size on every side.
- **Smallest sizes:** 16 px for the mark, 24 px for the lockup.
- **Don't** rotate or outline the mark, recolour single shapes, or put the brand colours on an accent fill (use the mono mark there).

## How these were made

The wordmark and tagline are [Onest](https://fonts.google.com/specimen/Onest) (SIL Open Font License, see `OFL-Onest.txt`), outlined to SVG paths so they look the same everywhere without loading a font. Text was shaped with HarfBuzz and outlined with fontTools from Onest's variable font at weight 700 (wordmark) and 500 (tagline). The mark files are copied unchanged from the design system.
