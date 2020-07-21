package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/willf/bitset"
)

func get_freq_idx(freq int) uint64 {
	switch freq {
	case 44100:
		return 4
	case 8000:
		return 11
	default:
		return 11
	}
}

func get_chn_idx(channel int) uint64 {
	if channel < 0 {
		channel = 0
	}

	if channel > 8 {
		channel = 8
	}

	return uint64(channel)
}

func pack(writer io.Writer, reader io.Reader) error {
	bits := bitset.New(8 * 7)

	sync := bitset.From([]uint64{0xfff00000000000})
	bits = bits.Union(sync)

	mpeg_ver := bitset.From([]uint64{0x00080000000000})
	bits = bits.Union(mpeg_ver)

	protect := bitset.From([]uint64{0x00010000000000})
	bits = bits.Union(protect)

	profile := bitset.From([]uint64{0x00004000000000})
	bits = bits.Union(profile)

	freq := bitset.From([]uint64{get_freq_idx(sample_rate) << 34})
	bits = bits.Union(freq)

	chn := bitset.From([]uint64{get_chn_idx(channel) << 30})
	bits = bits.Union(chn)

	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	if err != nil {
		fmt.Printf("Fail to copy buffer: %s\n", err)
		return err
	}

	sz := bitset.From([]uint64{uint64(buf.Len()+7) << 13})
	bits = bits.Union(sz)

	full := bitset.From([]uint64{0x7ff << 2})
	bits = bits.Union(full)

	mask := bits.Bytes()[0]
	mask_buf := make([]byte, 8)
	binary.BigEndian.PutUint64(mask_buf, mask)

	_, err = writer.Write(mask_buf[1:])
	if err != nil {
		fmt.Printf("Fail to write: %s\n", err)
		return err
	}

	_, err = writer.Write(buf.Bytes())
	if err != nil {
		fmt.Printf("Fail to write: %s\n", err)
		return err
	}

	return nil
}

var channel int
var sample_rate int

func main() {
	var out_path string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <aac frame> ...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&channel, "c", 1, "Number of channels")
	flag.IntVar(&sample_rate, "s", 8000, "Sampling rate")
	flag.StringVar(&out_path, "o", "outout.aac", "Output file path")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	output, err := os.Create(out_path)
	if err != nil {
		fmt.Printf("Fail to create file: %s\n", err)
		os.Exit(1)
	}
	defer output.Close()

	files := flag.Args()
	for _, f := range files {
		input, err := os.Open(f)
		if err != nil {
			fmt.Printf("Fail to open file: %s\n", err)
			os.Exit(1)
		}

		err = pack(output, input)
		if err != nil {
			fmt.Printf("Fail to pack: %s\n", err)
			os.Exit(1)
		}

		input.Close()
	}
}
