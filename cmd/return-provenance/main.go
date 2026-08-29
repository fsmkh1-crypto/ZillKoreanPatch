package main

import (
    "bufio"
    "flag"
    "fmt"
    "os"
    "regexp"
    "sort"
    "strconv"
    "strings"
)

type insn struct {
    Addr uint32
    Op   string
    Args string
    Raw  string
}

type function struct {
    Start uint32
    Insns []insn
}

type verdict struct {
    Start      uint32
    Kind       string
    Detail     string
    DescendTo  uint32
    HasDescend bool
}

const maxSnippetGap uint32 = 0x20

var lineRE = regexp.MustCompile(`^\s*0x([0-9A-Fa-f]{8})\s+(?:encoding=0x[0-9A-Fa-f]{8}\s+)?([A-Za-z0-9_.]+)\s*(.*)$`)

func main() {
    input := flag.String("input", "", "disassembly text file")
    start := flag.String("start", "", "start address, e.g. 0x08A23064")
    maxDepth := flag.Int("max-depth", 32, "maximum wrapper descent depth")
    flag.Parse()
    if *input == "" || *start == "" {
        fmt.Fprintln(os.Stderr, "usage: return-provenance --input FILE --start 0xADDRESS")
        os.Exit(2)
    }
    addr64, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(*start, "0x"), "0X"), 16, 32)
    if err != nil {
        fmt.Fprintln(os.Stderr, "invalid start address:", err)
        os.Exit(2)
    }
    f, err := os.Open(*input)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    defer f.Close()

    funcs, err := parseFunctions(bufio.NewScanner(f))
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    cur := uint32(addr64)
    seen := map[uint32]bool{}
    for depth := 0; depth < *maxDepth; depth++ {
        if seen[cur] {
            fmt.Printf("%02d 0x%08X cycle detected\n", depth, cur)
            return
        }
        seen[cur] = true
        fn, ok := funcs[cur]
        if !ok {
            fmt.Printf("%02d 0x%08X STOP missing function body in input\n", depth, cur)
            return
        }
        v := classify(fn)
        fmt.Printf("%02d 0x%08X %-18s %s\n", depth, cur, v.Kind, v.Detail)
        if !v.HasDescend {
            return
        }
        cur = v.DescendTo
    }
    fmt.Printf("STOP max depth %d reached\n", *maxDepth)
}

func parseFunctions(s *bufio.Scanner) (map[uint32]function, error) {
    all := make([]insn, 0, 256)
    for s.Scan() {
        m := lineRE.FindStringSubmatch(s.Text())
        if m == nil {
            continue
        }
        a, _ := strconv.ParseUint(m[1], 16, 32)
        all = append(all, insn{Addr: uint32(a), Op: strings.ToLower(m[2]), Args: strings.TrimSpace(m[3]), Raw: s.Text()})
    }
    if err := s.Err(); err != nil {
        return nil, err
    }
    sort.Slice(all, func(i, j int) bool { return all[i].Addr < all[j].Addr })
    funcs := map[uint32]function{}
    for i := 0; i < len(all); i++ {
        start := all[i].Addr
        body := []insn{all[i]}
        sawReturn := all[i].Op == "jr" && strings.TrimSpace(all[i].Args) == "ra"
        for j := i + 1; j < len(all); j++ {
            prev := body[len(body)-1].Addr
            next := all[j].Addr
            if next <= prev || next-prev > maxSnippetGap {
                break
            }
            body = append(body, all[j])
            if sawReturn {
                // Include the return delay slot, then stop. Partial evidence may omit
                // unrelated instructions, but never merge through a completed return.
                break
            }
            if all[j].Op == "jr" && strings.TrimSpace(all[j].Args) == "ra" {
                sawReturn = true
            }
        }
        funcs[start] = function{Start: start, Insns: body}
    }
    return funcs, nil
}

func classify(fn function) verdict {
    ret := -1
    for i := len(fn.Insns) - 1; i >= 0; i-- {
        if fn.Insns[i].Op == "jr" && strings.TrimSpace(fn.Insns[i].Args) == "ra" {
            ret = i
            break
        }
    }
    if ret < 0 {
        return verdict{Start: fn.Start, Kind: "uncertain", Detail: "no jr ra found in available body"}
    }
    lastJal := -1
    var jalTarget uint32
    for i := 0; i < ret; i++ {
        if fn.Insns[i].Op == "jal" {
            if t, ok := parseTarget(fn.Insns[i].Args); ok {
                lastJal, jalTarget = i, t
            }
        }
    }
    if lastJal >= 0 {
        writes := []string{}
        for i := lastJal + 2; i < ret; i++ { // skip jal delay slot
            if writesReg(fn.Insns[i], "v0") {
                writes = append(writes, fmt.Sprintf("0x%08X %s %s", fn.Insns[i].Addr, fn.Insns[i].Op, fn.Insns[i].Args))
            }
        }
        if len(writes) == 0 {
            return verdict{Start: fn.Start, Kind: "wrapper-passthrough", Detail: fmt.Sprintf("v0 unchanged after jal 0x%08X", jalTarget), DescendTo: jalTarget, HasDescend: true}
        }
        return verdict{Start: fn.Start, Kind: "value-transform", Detail: "v0 written after last jal: " + strings.Join(writes, "; ")}
    }
    for i := ret - 1; i >= 0; i-- {
        if writesReg(fn.Insns[i], "v0") {
            return verdict{Start: fn.Start, Kind: "value-origin", Detail: fmt.Sprintf("last v0 definition: 0x%08X %s %s", fn.Insns[i].Addr, fn.Insns[i].Op, fn.Insns[i].Args)}
        }
    }
    return verdict{Start: fn.Start, Kind: "uncertain", Detail: "no direct jal and no v0 definition found"}
}

func writesReg(i insn, reg string) bool {
    switch i.Op {
    case "move", "li", "lui", "lw", "lh", "lhu", "lb", "lbu", "addiu", "addu", "subu", "or", "ori", "and", "andi", "xor", "xori", "sll", "srl", "sra", "slt", "sltu", "slti", "sltiu":
        first := strings.TrimSpace(strings.Split(i.Args, ",")[0])
        return first == reg
    default:
        return false
    }
}

func parseTarget(arg string) (uint32, bool) {
    fields := strings.Fields(strings.TrimSpace(arg))
    if len(fields) == 0 {
        return 0, false
    }
    arg = fields[0]
    if strings.HasPrefix(strings.ToLower(arg), "z_un_") {
        arg = arg[len("z_un_"):]
    }
    arg = strings.TrimPrefix(strings.TrimPrefix(arg, "0x"), "0X")
    if len(arg) != 8 {
        return 0, false
    }
    x, err := strconv.ParseUint(arg, 16, 32)
    return uint32(x), err == nil
}
