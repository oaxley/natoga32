# Video Description


## Screen Resolutions

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


