# Video Description

## WIP

### Screen Resolutions

- 800x480, 640x384, 480x288, 320x192
- Constant ratio of 5/3
- RGB565 / 16-bit / 65,536 colors


Default Screen resolution "Mode 0" is 640x384: 640x384 x2 = 1280x768

Modes:
- Mode 0 : 640x384
- Mode 1 : 320x192
- Mode 2 : 480x288
- Mode 3 : 800x480

Framebuffers A and B for smooth animation
- 800x480x2 = 768,000 bytes
- x2 FB = 1,536,000 bytes ~ 1.5 MiB

### BG

Tile-based renderer with 4 Background TileSet (BG)
- BG0 : foreground tilemap
- BG1 : midground tilemap
- BG2 : far background
- BG3 : static sky or parallax layer

Drawing order (with priority disabled)
BG3 -> BG2 -> BG1 -> BG0 -> Sprites

Each BG has its own tilemap, scrolling parameters and affine transformation matrix
Tiles are 16x16 4bpp:
- 1 tile 16x16 4bpp = 128B / tile
- 1 sheet => 256 tiles => 128x256 = 32,768B (32KiB)
- 1 array => 16 sheets => 32KiB x 16 = 512KiB
- 16 sheets x 256 Tiles = 4096 tiles total

Tilemap entries (2B per entry)
- 800x480 /16 = 50x30 => 64x32 = 2048 tiles
- Scrolling: 1 line above and below, 7 column left and right
- Total entries: 2048 x 2B = 4096B (4KiB) per map
- 4xBG x 4KiB = 16KiB
- 4 TileMaps per BG => 64KiB

Tilemap Entry:
- bit 0-7: tile index number
- bit 8-11: sheet number
- bit 12-13: mirror/flip
- bit 14-15: reserved

Tile #0 is the empty tile, used when no tile is needed

BG Palette
- 4bpp => 16 colors x 2B = 32B / palette
- 256 palette => 8,192 Bytes (8KiB)

BG CSR Registers
- BGx_A: A parameter for affine transformation
- BGx_B: B parameter for affine transformation
- BGx_C: C parameter for affine transformation
- BGx_D: D parameter for affine transformation
- BGx_X: Scroll X
- BGx_Y: Scroll Y
- BGx_CTRL: BG control register
  - Enable / disable (1b)
  - Wrap or Clip for affine transformation
  - priority / depth (2b)
- BGx_TILEMAP: tilemap address/selector
- BGx_BLEND: alpha blending / mode (?)

Per scanline modification of the registers to allow for interesting effects
like Mode 7 from the SNES


### Sprites

8bpp sprites, color 0 = transparent color
Shape & Size determine the sprite size
- Shape: 2-bit
  - 00: square
  - 01: wide (horizontal)
  - 10: tall (vertical)
  - 11: reserved
- Size: 2-bit
  - 00: small
  - 01: medium
  - 10: large
  - 11: extra-large

Correspondance matrix:
| Shape | Size | Dimension |
| --- | --- | :---: |
| Square | Small | 8x8 |
| Square | Medium | 16x16 |
| Square | Large | 32x32 |
| Square | Extra-Large | 64x64 |
| Wide | Small | 16x8 |
| Wide | Medium | 32x8 |
| Wide | Large | 32x16 |
| Wide | Extra-Large | 64x32 |
| Tall | Small | 8x16 |
| Tall | Medium | 8x32 |
| Tall | Large | 16x32 |
| Tall | Extra-Large | 32x64 |

- Sprites have priority
  - 0 = draw last (on top of everything)
  - 3 = draw first
  - Priority checked against the BG priority

- Sprite memory allocation: 512 KiB
  - 8,192 sprites 8x8
  - 2,048 sprites 16x16
  - 512 sprites 32x32
  - 128 sprites 64x64

- Sprite palette
  - 256 colors x 2B = 512B / palette
  - 32 palettes x 512B = 16KiB

### Object Attribute Map (OAM)

8 bytes per sprite, 16KiB total = 2048 sprites

Attributes:
- X position (10b)
- Y position (10b)
- Enable/Disable (1b)
- Flip (2b)
- Priority (2b)
- Size (2b)
- Shape (2b)
- Palette (5b)
- Blending (4b)
- Affine (6b)
- Index (13b)
- Center (2b)
- Reserved (5b)

**Flip**:
- 00: None (normal)
- 01: Horizontal
- 10: Vertical
- 11: Both

**Blending**: always done against the Framebuffer 
Formula: final = A * FG + B * BG
- Enable (1b)
- Mode (3b)
  - normal (000): no blending
  - alpha 25/75 (001): sprite 25% / framebuffer 75% 
  - alpha 50/50 (010)
  - alpha 75/25 (011)
  - additive (100): max(sprite+framebuffer, 255)
  - subtractive (101): min(sprite-framebuffer, 0)
  - unused (110)
  - unused (111)

**Center**: for affine transformation / rotation
- 00: Center is top-left corner (X, Y)
- 01: Center is top-middle point (X + W/2, Y)
- 10: Center is left-middle point (X, Y + H/2)
- 11: Center is middle-middle point (X + W/2, Y + H/2)

### Affine Matrices

```
x_screen = [A * (x-XC) + B * (y-YC)] + (X+XC)
y_screen = [C * (x-XC) + D * (y-YC)] + (Y+YC)

(x,y) = pixel inside the sprite / BG
(X,Y) = current position of the sprite / BG
```

- Only A, B, C, D are stored in Fixed-Point 16-bit format.
- CX, CY optional center for Rotation
- 64 matrices: 64 x 8B = 512B

### Fonts
- 256KiB for font
- 256 characters
- 8x8, 16x16 or 32x32
- 4bpp or 8bpp
- Configuration per CSR Registers
  - Size
  - Depth
  - Palette

Fonts Palette:
- 16 palettes with 16 colors (16 x 2B x 16 = 512B)
- 16 palettes with 256 colors (256 x 2B x 15 = 8192B)

### Misc

**Viewport**:
- Reduce the display screen
- CSR Registers to specify the viewport dimensions:
  - start_x, start_y, width, height

**Color Math Engine**:
- Support multiple effects: blending, fading, sepia
