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

func bootPartition(disk string) (bootDisk string, bootPart string, err error) {
	disks := []string{"/dev/sda", "/dev/sdb", "/dev/vda", "/dev/vdb"}
	if disk != "" {
		disks = []string{disk}
	}

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
			if err != nil {
				fmt.Printf("failed to read gpt_table: %v\n", err)
				continue
			}
			for n, p := range gpt_table.Partitions {
				if p.Name() == "boot" {
					return d, fmt.Sprintf("%s%d", d, n), nil
				} else if ! p.IsEmpty() {
					var i int
					for i = 0; i < 36; i++ {
						if p.PartNameUTF16[2*i] == 0 &&
							p.PartNameUTF16[2*i + 1] == 0 {
								fmt.Printf("len %d\n", i)
								break
							}
					}
					fmt.Printf("Found %d: %s\n", n, p.Name()[:i])
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
	disk := ""
	fmt.Printf("args is %v len %d\n", os.Args, len(os.Args))
	if len(os.Args) == 2 {
		disk = os.Args[1]
	}
	d, p, e := bootPartition(disk)
	fmt.Printf("disk %s part %s err %v\n", d, p, e)
}
