package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/wangxuesong/tcpshadow/cmd"
	"github.com/wangxuesong/tcpshadow/model"
)

type SavePackage struct {
	number  int
	Forware model.DataForward
	length  int
	Buffer  []byte
}

func readPackages() []SavePackage {
	file, err := os.Open("./aaa")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	packages := make([]SavePackage, 0, 30)

	for {
		var header struct {
			Index   uint16
			Forward uint8
			Length  uint32
		}
		err := binary.Read(file, binary.LittleEndian, &header)
		if err == io.EOF {
			return packages
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(header)
		index := int(header.Index)
		forward := model.DataForward(header.Forward)
		length := int(header.Length)
		buf := make([]byte, length)
		data := SavePackage{
			number:  index,
			Forware: forward,
			length:  length,
			Buffer:  buf,
		}
		count, err := file.Read(data.Buffer)
		fmt.Println(index, forward, length, count)
		fmt.Println(data)
		if err != nil {
			return packages
		}
		packages = append(packages, data)
	}
}

func main() {
	cmd.Execute()
	//readPackages()
	//capture()
}
