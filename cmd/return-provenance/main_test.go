package main

import (
    "bufio"
    "encoding/binary"
    "strings"
    "testing"
)

func parseText(t *testing.T, text string) map[uint32]function {
    t.Helper()
    f, err := parseFunctions(bufio.NewScanner(strings.NewReader(text)))
    if err != nil {
        t.Fatal(err)
    }
    return f
}

func TestWrapperPassthrough(t *testing.T) {
    text := `
0x089E2B54 addiu sp,sp,-0x90
0x089E2B58 sw s0,0x80(sp)
0x089E2B5C lui s0,0x8A7
0x089E2B60 lw a1,-0x53B4(s0)
0x089E2B64 addiu a2,sp,0x10
0x089E2B68 sw a1,0x10(sp)
0x089E2B6C li a1,0x2
0x089E2B70 sw a2,-0x53B4(s0)
0x089E2B74 sb a1,0x14(sp)
0x089E2B78 li a1,0x08A6A008
0x089E2B80 sw ra,0x84(sp)
0x089E2B84 jal z_un_089E2B00
0x089E2B88 sw a1,0x18(sp)
0x089E2B8C lw a0,0x10(sp)
0x089E2B90 sw a0,-0x53B4(s0)
0x089E2B94 lw ra,0x84(sp)
0x089E2B98 lw s0,0x80(sp)
0x089E2B9C jr ra
0x089E2BA0 addiu sp,sp,0x90
`
    funcs := parseText(t, text)
    v := classify(funcs[0x089E2B54])
    if v.Kind != "wrapper-passthrough" {
        t.Fatalf("kind=%s detail=%s", v.Kind, v.Detail)
    }
    if !v.HasDescend || v.DescendTo != 0x089E2B00 {
        t.Fatalf("descend=%v target=0x%08X", v.HasDescend, v.DescendTo)
    }
}

func TestValueTransformStops(t *testing.T) {
    text := `
0x08A00000 jal z_un_08A01000
0x08A00004 nop
0x08A00008 addiu v0,v0,0x10
0x08A0000C jr ra
0x08A00010 nop
`
    funcs := parseText(t, text)
    v := classify(funcs[0x08A00000])
    if v.Kind != "value-transform" || v.HasDescend {
        t.Fatalf("unexpected verdict: %+v", v)
    }
}

func TestLeafValueOrigin(t *testing.T) {
    text := `
0x08A02000 lw v0,0x20(a0)
0x08A02004 jr ra
0x08A02008 nop
`
    funcs := parseText(t, text)
    v := classify(funcs[0x08A02000])
    if v.Kind != "value-origin" {
        t.Fatalf("unexpected verdict: %+v", v)
    }
}

func TestBinaryRuntimeToFileMapping(t *testing.T) {
    img := &binaryImage{Data: make([]byte, 0x21F200), RuntimeBase: 0x08804000, FileBias: 0x80, MaxInsns: 16}
    off, ok := img.fileOffset(0x08A23064)
    if !ok || off != 0x21F0E4 {
        t.Fatalf("mapping ok=%v off=0x%08X", ok, off)
    }
}

func TestBinaryWrapperPassthroughDecode(t *testing.T) {
    const start = uint32(0x08A00000)
    const target = uint32(0x08A01000)
    data := make([]byte, 0x200)
    words := []uint32{
        0x0C000000 | ((target >> 2) & 0x03FFFFFF), // jal target
        0x00000000,                               // delay slot
        0x03E00008,                               // jr ra
        0x00000000,                               // return delay slot
    }
    for i, w := range words {
        binary.LittleEndian.PutUint32(data[i*4:], w)
    }
    img := &binaryImage{Data: data, RuntimeBase: start, FileBias: 0, MaxInsns: 16}
    fn, ok := img.decodeFunction(start)
    if !ok {
        t.Fatal("decodeFunction failed")
    }
    v := classify(fn)
    if v.Kind != "wrapper-passthrough" || !v.HasDescend || v.DescendTo != target {
        t.Fatalf("unexpected verdict: %+v", v)
    }
}
