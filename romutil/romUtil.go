package romutil

import (
	"image"
	"image/color"
	"image/draw"
	"os"
)

func LoadROM(fileName string) (int, []byte) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(loloMemoryChecks); i++ {
		if data[loloMemoryChecks[i].offset] == loloMemoryChecks[i].value {
			return loloMemoryChecks[i].gameType, data
		}
	}
	return N_TYPE_ERRORDETECTING, nil
}

func LoadAllLevels(gameType int, romData []byte) (int, int, [][]byte) {
	var loloLevelsUncompressed [][]byte = make([][]byte, 60)
	var levelAmount int
	switch gameType {
	case N_TYPE_LOLO1:
		for i := 0; i <= lolo1Attr.levelAmount; i++ {
			offset := lolo1Attr.levelData + int(romData[lolo1Attr.levelPointerTable+(i*2)+1])<<8 + int(romData[lolo1Attr.levelPointerTable+(i*2)])
			if romData[lolo1Attr.levelBankTable+i] == 6 {
				offset += 0x1000
			}
			if offset+int(romData[lolo1Attr.levelSizeTable+i]) > len(romData) { //???????
				return 1, 0, nil
			}
			loloLevelsUncompressed[i] = decompressLevel(romData[offset:])
			if loloLevelsUncompressed[i] == nil {
				return 1, 0, nil
			}
		}
		levelAmount = lolo1Attr.levelAmount
	case N_TYPE_LOLO2:
		for i := 0; i <= lolo2Attr.levelAmount; i++ {
			offset := lolo2Attr.levelData + int(romData[lolo2Attr.levelPointerTable+(i*2)+1])<<8 + int(romData[lolo2Attr.levelPointerTable+(i*2)])
			if offset+int(romData[lolo2Attr.levelSizeTable+i]) > len(romData) { //???????
				return 1, 0, nil
			}
			loloLevelsUncompressed[i] = decompressLevel(romData[offset:])
			if loloLevelsUncompressed[i] == nil {
				return 1, 0, nil
			}
		}
		levelAmount = lolo2Attr.levelAmount
	}
	return 0, levelAmount, loloLevelsUncompressed
}

func decompressLevel(compressedData []byte) []byte {
	i := 0
	decompressedData := make([]byte, 240)
	for y := 14; y >= 0; y-- {
		for x := 0; x <= 15; x++ {
			if compressedData[i] < 0xF0 {
				decompressedData[(y*16)+x] = compressedData[i]
			}
			if (compressedData[i] >= 0xF0) && (compressedData[i] < 0xFF) {
				if (int(compressedData[i]-0xEF) + x) > 15 {
					return nil
				}
				for temp := int(compressedData[i]-0xEF) + x; temp >= x; temp-- {
					decompressedData[(y*16)+temp] = 0xFF
				}
				x += int(compressedData[i] - 0xEF)
			}
			if compressedData[i] == 0xFF {
				i++
				if (int(compressedData[i]+1) + x) > 15 {
					return nil
				}
				for temp := int(compressedData[i]+1) + x; temp >= x; temp-- {
					decompressedData[(y*16)+temp] = compressedData[i+1]
				}
				x += int(compressedData[i] + 1)
				i++
			}
			i++
		}
	}

	for y := 14; y >= 0; y-- {
		for x := 0; x <= 15; x++ {
			if decompressedData[(y*16)+x] == 0xFF {
				if y < 14 {
					decompressedData[(y*16)+x] = decompressedData[((y+1)*16)+x]
				} else {
					decompressedData[(y*16)+x] = 0
				}
			}
		}
	}
	//printLevelData(decompressedData)
	return decompressedData
}

func ExtractTileTextures(showsprites, showadditionalinfo, showblockid bool, romType int, romData []byte) []image.Image {
	blockbitmap := make([]image.Image, 0xF0)
	switch romType {
	case N_TYPE_LOLO1:
		//TODO: implement for lolo1
	case N_TYPE_LOLO2:
		for i := 0; i < 0xF0; i++ {
			tsaindex := i

			if showsprites {
				if i == 0xd4 {
					tsaindex = 0x4e
				} //this makes the treasure chest show correctly
				//this is based on $B01D in lolo 2 where it draws a floor block underneath a sprite
				if ((i >= 0x80) && (i < 0xa0)) || (i == 0xc0) {
					tsaindex = 0x40
				}
			}

			blockbitmap[i] =
				Get16x16BlockBitmap(romData[lolo2Attr.levelBGGFX:], //use the offset of level block GFX
					romData[lolo2Attr.TLBlockTSA+tsaindex], //pass it the correct TSA offsets
					romData[lolo2Attr.TRBlockTSA+tsaindex],
					romData[lolo2Attr.BLBlockTSA+tsaindex],
					romData[lolo2Attr.BRBlockTSA+tsaindex],
					romData[(lolo2Attr.levelBGPallete+int(romData[lolo2Attr.blockPalleteIndex+tsaindex]*4)):],
				)

			if showsprites {
				if ((i >= 0x80) && (i < 0xa0)) || (i == 0xc0) {
					var sprite image.Image
					if i != 0xc0 {
						sprite = get16x16SpriteTSAEntry(int(lolo2_spritetsaindex[i-0x80]), romData)
					} else {
						sprite = get16x16SpriteTSAEntry(lolo2_spritec0tsaindex, romData)
					}
					newBlock := image.NewRGBA(blockbitmap[i].Bounds())
					draw.Draw(newBlock, blockbitmap[i].Bounds(), blockbitmap[i], image.Point{}, draw.Src)
					draw.Draw(newBlock, sprite.Bounds(), sprite, image.Point{}, draw.Over)
					blockbitmap[i] = newBlock
				}
			}
		}
	}

	return blockbitmap
}

func get16x16SpriteTSAEntry(index int, romData []byte) image.Image {
	sprite := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{16, 16}})
	var spritehalf image.Image

	if index >= 0x4c {
		index -= 0x4c
		//draw the left half
		spritehalf = get8x16SpriteBitmap(
			romData[lolo2Attr.levelSpriteGFX:],        //patterntable 0 data
			romData[lolo2Attr.levelBGGFX:],            //patterntable 1 data
			romData[lolo2Attr.CSPRITETSA+(index*4)],   //spriteindex
			romData[lolo2Attr.CSPRITETSA+1+(index*4)], //sprite attributes
			romData[lolo2Attr.levelSpritePallete:],    //nes pallete
			true,
		) //transparency is true
		draw.Draw(sprite, spritehalf.Bounds(), spritehalf, image.Point{}, draw.Src)

		spritehalf = get8x16SpriteBitmap(
			romData[lolo2Attr.levelSpriteGFX:],          //patterntable 0 data
			romData[lolo2Attr.levelBGGFX:],              //patterntable 1 data
			romData[lolo2Attr.CSPRITETSA+(index*4)+2],   //spriteindex
			romData[lolo2Attr.CSPRITETSA+1+(index*4)+2], //sprite attributes
			romData[lolo2Attr.levelSpritePallete:],      //nes pallete
			true,
		)
		draw.Draw(sprite, spritehalf.Bounds().Add(image.Point{8, 0}), spritehalf, image.Point{}, draw.Over)
		return sprite
	}

	//we're here if the sprite index is < 0x4c the sprite drawing code for this is a little more
	//complex
	tsaoffset := int(romData[lolo2Attr.spriteTSAHIPointerTable+index]) +
		int(romData[lolo2Attr.spriteTSALOPointerTable+index])<<8 - 0x8000 + 0x10

	if tsaoffset > 0x8010 {
		return nil
	} //if its too far then something is currupt in this rom

	spriteamount := int(romData[tsaoffset])
	if tsaoffset+(spriteamount*4) > 0x8010 {
		return nil
	} //again someing is currupt if its this far

	tsaoffset++

	xsize := 8
	ysize := 16
	for i := 0; i < spriteamount; i++ {
		if int(romData[tsaoffset+(i*4)])+16 > ysize {
			ysize = int(romData[tsaoffset+(i*4)] + 16)
		}
		if int(romData[tsaoffset+(i*4)+3])+8 > xsize {
			xsize = int(romData[tsaoffset+(i*4)+3] + 8)
		}
	}

	//remember this function is ONLY for 16x16 sprites
	if (xsize != 16) && (ysize != 16) {
		return nil
	}

	for i := 0; i < spriteamount; i++ {
		//now lets draw it
		spritehalf = get8x16SpriteBitmap(
			romData[lolo2Attr.levelSpriteGFX:],     //patterntable 0 data
			romData[lolo2Attr.levelBGGFX:],         //patterntable 1 data
			romData[tsaoffset+(i*4)+1],             //spriteindex
			romData[tsaoffset+(i*4)+2],             //sprite attributes
			romData[lolo2Attr.levelSpritePallete:], //nes pallete
			true,
		) //transparency is true

		//for some reason the game has some sprites (notably sprite 0x17) with a y position of 0xff
		//or -1, this here sets it back to 0 by referencing it as a signed char
		x := romData[tsaoffset+(i*4)+3]
		y := romData[tsaoffset+(i*4)]
		if x == 0xFF {
			x = 0
		}
		if y == 0xFF {
			y = 0
		}
		draw.Draw(sprite, spritehalf.Bounds().Add(image.Point{int(x), int(y)}), spritehalf, image.Point{}, draw.Src)
	}
	return sprite
}

func get8x16SpriteBitmap(patterntable0data, patterntable1data []byte, spriteindex, spriteattributes byte, nespallete []byte, transparency bool) image.Image {
	spritebitmap := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{8, 16}})

	// this tells it to use pattern table 0 or 1 data depending on the tileindex, just like the nes does
	// for 8x16 sprites
	var vromdata []byte
	if spriteindex&1 != 0 { //???
		vromdata = patterntable1data
	} else {
		vromdata = patterntable0data
	}

	spriteindex &= 0xFE //get rid of the lowest bit

	for x := 0; x < 8; x++ {
		for y := 0; y < 16; y++ {

			var tempx, tempy int
			var tempindex, pixelvalue byte
			if y < 8 {
				tempindex = spriteindex
			} else {
				tempindex = spriteindex + 1
			}
			if spriteattributes&ATTRIB_SPRITE_VFLIP != 0 { //???
				tempy = 7 - (y & 7)
			} else {
				tempy = (y & 7)
			}
			if spriteattributes&ATTRIB_SPRITE_HFLIP != 0 { //???
				tempx = 7 - (x & 7)
			} else {
				tempx = (x & 7)
			}
			pixelvalue = getVRomPatternTablePixel(vromdata, tempindex, tempx, tempy)

			nespalleteindex := nespallete[pixelvalue+((spriteattributes&3)*4)]
			if transparency {
				if pixelvalue > 0 {
					spritebitmap.Set(x, y, color.RGBA{
						nes_palbase_red[nespalleteindex&0x3F],
						nes_palbase_green[nespalleteindex&0x3F],
						nes_palbase_blue[nespalleteindex&0x3F],
						0xFF})
				} else {
					spritebitmap.Set(x, y, color.RGBA{0, 0, 0, 0})
				}
			} else {
				spritebitmap.Set(x, y, color.RGBA{
					nes_palbase_red[nespalleteindex&0x3F],
					nes_palbase_green[nespalleteindex&0x3F],
					nes_palbase_blue[nespalleteindex&0x3F],
					0xFF})
			}
		}
	}

	return spritebitmap
}

func Get16x16BlockBitmap(vromdata []byte, tlblocktsa, trblocktsa, blblocktsa, brblocktsa byte, nespallete []byte) image.Image {
	image := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{16, 16}})
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			var nespalleteindex byte
			if x < 8 && y < 8 {
				nespalleteindex = nespallete[getVRomPatternTablePixel(vromdata, tlblocktsa, x%8, y%8)]
			}
			if x >= 8 && y < 8 {
				nespalleteindex = nespallete[getVRomPatternTablePixel(vromdata, trblocktsa, x%8, y%8)]
			}
			if x < 8 && y >= 8 {
				nespalleteindex = nespallete[getVRomPatternTablePixel(vromdata, blblocktsa, x%8, y%8)]
			}
			if x >= 8 && y >= 8 {
				nespalleteindex = nespallete[getVRomPatternTablePixel(vromdata, brblocktsa, x%8, y%8)]
			}
			image.Set(x, y, color.RGBA{
				nes_palbase_red[nespalleteindex&0x3F],
				nes_palbase_green[nespalleteindex&0x3F],
				nes_palbase_blue[nespalleteindex&0x3F],
				0xFF})
		}
	}
	return image
}

func getVRomPatternTablePixel(VRomData []byte, TileNumber byte, x, y int) byte {
	offset1 := int(TileNumber)*16 + y
	offset2 := offset1 + 8

	switch x {

	case 0:
		return ((VRomData[offset1] & 0x80) >> 7) + ((VRomData[offset2] & 0x80) >> 6)

	case 1:
		return ((VRomData[offset1] & 0x40) >> 6) + ((VRomData[offset2] & 0x40) >> 5)

	case 2:
		return ((VRomData[offset1] & 0x20) >> 5) + ((VRomData[offset2] & 0x20) >> 4)

	case 3:
		return ((VRomData[offset1] & 0x10) >> 4) + ((VRomData[offset2] & 0x10) >> 3)

	case 4:
		return ((VRomData[offset1] & 0x08) >> 3) + ((VRomData[offset2] & 0x08) >> 2)

	case 5:
		return ((VRomData[offset1] & 0x04) >> 2) + ((VRomData[offset2] & 0x04) >> 1)

	case 6:
		return ((VRomData[offset1] & 0x02) >> 1) + (VRomData[offset2] & 0x02)

	case 7:
		return (VRomData[offset1] & 0x01) + ((VRomData[offset2] & 0x01) << 1)

	}
	return 0
}
