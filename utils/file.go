package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func GetLastLine(fileName string) (int64, error) {

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err.Error())
	}

	if _, err := f.Seek(0, 0); err != nil {
		return -1, err
	}

	r := bufio.NewReader(f)
	pos := int64(0)
	for {
		data, err := r.ReadBytes('\n')
		pos += int64(len(data))
		if err == nil || err == io.EOF {
			if len(data) > 0 && data[len(data)-1] == '\n' {
				data = data[:len(data)-1]
			}
			if len(data) > 0 && data[len(data)-1] == '\r' {
				data = data[:len(data)-1]
			}
			//fmt.Printf("Pos: %d, Read: %s\n", pos, data)
			//time.Sleep(time.Second)
		}
		if err != nil {
			if err != io.EOF {
				return -1, err
			}
			break
		}
	}
	return pos, nil
}
