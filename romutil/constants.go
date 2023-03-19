package romutil

type MemoryCheck struct {
	gameType int
	offset   int
	value    byte
}

type GameAttributes struct {
	gameType                int
	levelAmount             int
	levelData               int
	levelPointerTable       int
	levelBankTable          int
	levelSizeTable          int
	levelBGGFX              int
	levelBGPallete          int
	TLBlockTSA              int
	TRBlockTSA              int
	BLBlockTSA              int
	BRBlockTSA              int
	blockPalleteIndex       int
	levelSpriteGFX          int
	levelSpritePallete      int
	spriteTSAHIPointerTable int
	spriteTSALOPointerTable int
	CSPRITETSA              int
}

const (
	N_TYPE_INVALIDROM = iota
	N_TYPE_LOLO1
	N_TYPE_LOLO2
	N_TYPE_LOLO3
	N_TYPE_EGGERLAND
	N_TYPE_LOLO1J
	N_TYPE_LOLO2J
	N_TYPE_ERRORDETECTING
	ATTRIB_SPRITE_VFLIP    byte = 0x80
	ATTRIB_SPRITE_HFLIP    byte = 0x40
	ATTRIB_BKG_PRIORITY    byte = 0x20
	lolo2_spritec0tsaindex      = 0x58
)

var (
	lolo1Attr = GameAttributes{
		levelAmount:       49,
		levelData:         0xD010,
		levelPointerTable: 0x3A6A,
		levelBankTable:    0x3B00,
		levelSizeTable:    0x3ACE,
		levelBGGFX:        0x9010,
		levelBGPallete:    0x1A80,
		TLBlockTSA:        0x3b32,
		TRBlockTSA:        0x3c22,
		BLBlockTSA:        0x3d12,
		BRBlockTSA:        0x3e02,
		blockPalleteIndex: 0x3ef2,
	}
	lolo2Attr = GameAttributes{
		levelAmount:             52,
		levelData:               0xE010,
		levelPointerTable:       0x3394,
		levelSizeTable:          0x3406,
		levelBGGFX:              0x9010,
		levelBGPallete:          0x1C4A,
		TLBlockTSA:              0x343B,
		TRBlockTSA:              0x352B,
		BLBlockTSA:              0x361B,
		BRBlockTSA:              0x370B,
		blockPalleteIndex:       0x37FB,
		levelSpriteGFX:          0x8010,
		levelSpritePallete:      0x1C9B,
		spriteTSAHIPointerTable: 0x7422,
		spriteTSALOPointerTable: 0x746E,
		CSPRITETSA:              0x7AE2,
	}
)

var loloMemoryChecks = [...]MemoryCheck{
	{N_TYPE_INVALIDROM, 0, 0},
	{N_TYPE_LOLO1, 0x10, 0x20},
	{N_TYPE_LOLO2, 0x1ac, 0xC5},
	{N_TYPE_LOLO3, 0x11, 0x4c},
	{N_TYPE_EGGERLAND, 0, 0},
	{N_TYPE_LOLO1J, 0x1ac, 0x35},
	{N_TYPE_LOLO2J, 0x11, 0xb6},
}

var nes_palbase_red = [64]byte{
	0x78, 0x00, 0x00, 0x40, 0x94, 0xac, 0xac, 0x8c,
	0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xbc, 0x00, 0x00, 0x68, 0xdc, 0xe4, 0xfc, 0xe4,
	0xac, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00,
	0xfc, 0x38, 0x68, 0x9c, 0xfc, 0xfc, 0xfc, 0xfc,
	0xfc, 0xbc, 0x58, 0x58, 0x00, 0x60, 0x00, 0x00,
	0xfc, 0xa4, 0xbc, 0xdc, 0xfc, 0xf4, 0xf4, 0xfc,
	0xfc, 0xdc, 0xb8, 0xb0, 0x00, 0xc8, 0x00, 0x00}

var nes_palbase_green = [64]byte{
	0x80, 0x00, 0x00, 0x28, 0x00, 0x00, 0x10, 0x18,
	0x30, 0x78, 0x68, 0x58, 0x40, 0x00, 0x00, 0x00,
	0xc0, 0x78, 0x88, 0x48, 0x00, 0x00, 0x38, 0x60,
	0x80, 0xb8, 0xa8, 0xa8, 0x88, 0x2c, 0x00, 0x00,
	0xf8, 0xc0, 0x88, 0x78, 0x78, 0x58, 0x78, 0xa0,
	0xb8, 0xf8, 0xd8, 0xf8, 0xe8, 0x60, 0x00, 0x00,
	0xf8, 0xe8, 0xb8, 0xb8, 0xb8, 0xc0, 0xd0, 0xe0,
	0xd8, 0xf8, 0xf8, 0xf0, 0xf8, 0xc0, 0x00, 0x00}

var nes_palbase_blue = [64]byte{
	0x84, 0xfc, 0xc4, 0xc4, 0x8c, 0x28, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x58, 0x00, 0x00, 0x08,
	0xc4, 0xfc, 0xfc, 0xfc, 0xd4, 0x60, 0x00, 0x18,
	0x00, 0x00, 0x00, 0x48, 0x94, 0x2c, 0x00, 0x00,
	0xfc, 0xfc, 0xfc, 0xfc, 0xfc, 0x9c, 0x58, 0x48,
	0x00, 0x18, 0x58, 0x9c, 0xe4, 0x60, 0x00, 0x00,
	0xfc, 0xfc, 0xfc, 0xfc, 0xfc, 0xe0, 0xb4, 0xb4,
	0x84, 0x78, 0x78, 0xd8, 0xfc, 0xc0, 0x00, 0x00}

var lolo2_spritetsaindex = []byte{
	0x9F, 0xA1, 0xA3, 0xa5, //the green enemy that falls asleep when he touches you
	0x11, 0x14, 0x17, 0x1a, //the walking stone enemy
	0x90, 0x93, 0x96, 0x99, //the grey enemy that rolls
	0x87, 0x88, 0x89, 0x8a, //the pink stationary enemy that shoots fire
	0x85, 0x85, 0x85, 0x85, //the pink monster that moves back and forth and shoots
	0x83, 0x83, 0x83, 0x83, //the grey face that shoots anything in its path
	0x7a, 0x7a, 0x7a, 0x7a, //the skull that comes alive at the end
	0x78, 0x78, 0x78, 0x78, //the friendly green snake
}