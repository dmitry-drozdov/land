//go:build exclude

package main

import (
	"encoding/binary"

	"golang.org/x/sys/cpu"
)

type Cipher struct {
	key     [8]uint32
	counter uint32
}

const bufSize = 256

//go:noescape
func chaCha20_ctr32_vsx(out, inp *byte, len int, key *[8]uint32, counter *uint32)

func (c *Cipher) xorKeyStreamBlocks(dst, src []byte) {
	chaCha20_ctr32_vsx(&dst[0], &src[0], len(src), &c.key, &c.counter)
}

//go:noescape
func updateVX(state *macState, msg []byte)

type mac struct {
	macState

	buffer [16 * TagSize]byte // size must be a multiple of block size (16)
	offset int
}

//go:noescape
func update(state *macState, msg []byte)

type mac struct{ macGeneric }

func chacha20Poly1305Open(dst []byte, key []uint32, src, ad []byte) bool

func chacha20Poly1305Seal(dst []byte, key []uint32, src, ad []byte)

var (
	useAVX2 = cpu.X86.HasAVX2 && cpu.X86.HasBMI2
)

func setupState(state *[16]uint32, key *[32]byte, nonce []byte) {
	state[0] = 0x61707865
	state[1] = 0x3320646e
	state[2] = 0x79622d32
	state[3] = 0x6b206574

	state[4] = binary.LittleEndian.Uint32(key[0:4])
	state[5] = binary.LittleEndian.Uint32(key[4:8])
	state[6] = binary.LittleEndian.Uint32(key[8:12])
	state[7] = binary.LittleEndian.Uint32(key[12:16])
	state[8] = binary.LittleEndian.Uint32(key[16:20])
	state[9] = binary.LittleEndian.Uint32(key[20:24])
	state[10] = binary.LittleEndian.Uint32(key[24:28])
	state[11] = binary.LittleEndian.Uint32(key[28:32])

	state[12] = 0
	state[13] = binary.LittleEndian.Uint32(nonce[0:4])
	state[14] = binary.LittleEndian.Uint32(nonce[4:8])
	state[15] = binary.LittleEndian.Uint32(nonce[8:12])
}
