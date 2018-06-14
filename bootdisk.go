package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/rekby/mbr"
	"github.com/rekby/gpt"
)

const PART_LINUX = 131

func getSize(dev string) uint64 {
	path := path.Join("/sys/block", path.Base(dev), "queue/logical_block_size")
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return 512
	}
	v, err := strconv.Atoi(string(d))
	if err == nil {
		return uint64(v)
	}
	return 512
}

func bootPartition() (bootDisk string, bootPart string, err error) {
	disks := []string{"/dev/sda", "/dev/sdb", "/dev/vda", "/dev/vdb"}

	firstDisk := ""
	firstPart := ""
	for _, d := range disks {
		f, err := os.Open(d)
		if err != nil {
			continue
		}
		defer f.Close()

		mbrTable, err := mbr.Read(f)
		if err != nil {
			fmt.Printf("error reading mbr on %s: %v\n", d, err)
			continue
		}
		if mbrTable.IsGPT() {
			fmt.Printf("%s is gpt\n", d)
			sz := getSize(d)
			fmt.Printf("sz for %s is %d\n", d, sz)
			gpt_table, err := gpt.ReadTable(f, sz)
			fmt.Printf("read table, %v err %v\n", gpt_table, err)
			for n, p := range gpt_table.Partitions {
				if p.Name() == "bios_grub" {
					return d, fmt.Sprintf("%s%d", d, n), nil
				}
			}
			continue
		}
		parts := mbrTable.GetAllPartitions()
		for n, p := range parts {
			if firstDisk == "" && n == 0 && p.GetType() == PART_LINUX {
				firstDisk = d
				firstPart = fmt.Sprintf("%s1", d)
			}
		}
	}

	if firstDisk == "" {
		return "", "", fmt.Errorf("Failed to find boot partition")
	}
	return firstDisk, firstPart, nil
}

func main() {
	d, p, e := bootPartition()
	fmt.Printf("disk %s part %s err %v\n", d, p, e)
}
