package main

import (
	"encoding/binary"
	"math"
	"math/rand"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"time"
)

var seed int64 = 0
var random *rand.Rand

func unpack8(data []byte) uint8 {
	return uint8(binary.LittleEndian.Uint16(data))
}

func pack8(data int) string {
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, uint16(data))
	return string(bs[0])
}

func unpack16(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data)
}

func pack16(data uint16) string {
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, data)
	return string(bs)
}

func unpack32(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data)
}

func pack32(data uint32) string {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, data)
	return string(bs)
}

func unpack(mode string, data string) int {
	var out int
	switch mode {
	case "B":
		out = int(unpack8([]byte(data + string(0x00))))
		break
	case "H":
		out = int(unpack16([]byte(data)))
		break
	case "I":
		out = int(unpack32([]byte(data)))
		break
	default:
		out = int(unpack32([]byte(data)))
		break
	}
	return out
}

func pack(mode string, data int) string {
	var out string
	switch mode {
	case "B":
		out = pack8(data)
		break
	case "H":
		out = pack16(uint16(data))
		break
	case "I":
		out = pack32(uint32(data))
		break
	default:
		out = pack32(uint32(data))
		break
	}
	return out
}

func randomint(min, max int) int {
	if seed == 0 {
		seed = unixmillisec()
	} else {
		seed = int64(random.Intn(int(unixmillisec())-1) + 1)
	}
	s := rand.NewSource(seed)
	random = rand.New(s)
	return random.Intn(max-min) + min
}

func unixmillisec() int64 {
	t := time.Now()
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func bitlen(x int) int {
	count := 0
	for {
		if x>>uint32(count) == 0 {
			break
		}
		count++
	}
	return count
}

func maxbit(x int) int {
	max := 1
	for i := 0; i < x-1; i++ {
		max = (max << 1) + 1
	}
	return max
}

func bitflip(data string, flip_size int, op_size int) string {
	if len(data) > op_size/8 {
		pos := randomint(0, len(data)-(op_size/8))
		var mode string
		switch op_size {
		case 8:
			mode = "B"
			break
		case 16:
			mode = "H"
			break
		case 32:
			mode = "I"
			break
		default:
			mode = "I"
			break
		}
		fuzz := unpack(mode, data[pos:pos+(op_size/8)])
		flip := uint32(randomint(0, (op_size)-bitlen(flip_size)+1))
		fuzz = fuzz ^ (flip_size << flip)
		tmp := data[:pos]
		tmp += pack(mode, fuzz)
		tmp += data[pos+(op_size/8):]
		return tmp
	} else {
		return ""
	}
}

func byteflip(data string, size int, op_size int) string {
	if op_size < size {
		op_size = size
	}
	if len(data) > op_size/8 {
		pos := randomint(0, len(data)-(op_size/8))
		flip_size := maxbit(size)
		var mode string
		switch op_size {
		case 8:
			mode = "B"
			break
		case 16:
			mode = "H"
			break
		case 32:
			mode = "I"
			break
		default:
			mode = "I"
			break
		}
		fuzz := unpack(mode, data[pos:pos+(op_size/8)])
		flip := uint32(randomint(0, (op_size)-bitlen(flip_size)+1))
		fuzz ^= flip_size << flip
		tmp := data[:pos]
		tmp += pack(mode, fuzz)
		tmp += data[pos+(op_size/8):]
		return tmp
	} else {
		return ""
	}
}

func arithmetics(data string, size int, op_size int) string {
	if len(data) > op_size/8 {
		pos := randomint(0, len(data)-(op_size/8))
		if size > op_size {
			size = op_size
		}
		op := int(randomint(-maxbit(size), maxbit(size)))
		var mode string
		switch op_size {
		case 8:
			mode = "B"
			break
		case 16:
			mode = "H"
			break
		case 32:
			mode = "I"
			break
		default:
			mode = "I"
			break
		}
		fuzz := unpack(mode, data[pos:pos+(op_size/8)])
		fuzz += op
		tmp := data[:pos]
		tmp += pack(mode, fuzz)
		tmp += data[pos+(op_size/8):]
		return tmp
	} else {
		return ""
	}
}

func randombytes(data string, size int, op_size int) string {
	if len(data) > op_size/8 {
		if size > op_size {
			size = op_size
		}
		pos := randomint(0, len(data)-(op_size/8))
		var mode string
		switch op_size {
		case 8:
			mode = "B"
			break
		case 16:
			mode = "H"
			break
		case 32:
			mode = "I"
			break
		default:
			mode = "I"
			break
		}
		fuzz := randomint(0, maxbit(size))
		tmp := data[:pos]
		tmp += pack(mode, fuzz)
		tmp += data[pos+(op_size/8):]
		return tmp
	} else {
		return ""
	}
}

func knownintegers(data string, size int, op_size int) string {
	if len(data) > op_size/8 {
		pos := randomint(0, len(data)-(op_size/8))
		var mode string
		switch op_size {
		case 8:
			mode = "B"
			break
		case 16:
			mode = "H"
			break
		case 32:
			mode = "I"
			break
		default:
			mode = "I"
			break
		}
		max_t := maxbit(size)
		values := []int{max_t, -max_t / 2, (max_t / 2) + 1}
		fuzz := unpack(mode, data[pos:pos+(op_size/8)])
		fuzz = values[randomint(0, len(values))]
		tmp := data[:pos]
		tmp += pack(mode, fuzz)
		tmp += data[pos+(op_size/8):]
		return tmp
	} else {
		return ""
	}
}

func duplicatetoken(data string, size int, op_size int) string {
	if len(data) > op_size/8 {
		pos := randomint(0, len(data)-(op_size/8))
		fuzz := strings.Repeat(data[pos:pos+(op_size/8)], randomint(1, 30))
		tmp := data[:pos]
		tmp += fuzz
		tmp += data[pos+(op_size/8):]
		return tmp
	} else {
		return ""
	}
}

func removetoken(data string, size int, op_size int) string {
	if len(data) > 4 {
		pos := randomint(0, len(data)-(op_size/8))
		tmp := data[:pos]
		tmp += data[pos+(op_size/8):]
		return tmp
	} else {
		return ""
	}
}

func insertrandombytes(data string, size int, op_size int) string {
	if len(data) > op_size/8 {
		pos := randomint(0, len(data)-(op_size/8))
		bytes := randomint(0, maxbit(op_size))
		var mode string
		switch op_size {
		case 8:
			mode = "B"
			break
		case 16:
			mode = "H"
			break
		case 32:
			mode = "I"
			break
		default:
			mode = "I"
			break
		}
		tmp := data[:pos]
		tmp += pack(mode, bytes)
		tmp += data[pos:]
		return tmp
	} else {
		return ""
	}
}

func permutations(arr []int) [][]int {
	var helper func([]int, int)
	res := [][]int{}

	helper = func(arr []int, n int) {
		if n == 1 {
			tmp := make([]int, len(arr))
			copy(tmp, arr)
			res = append(res, tmp)
		} else {
			for i := 0; i < n; i++ {
				helper(arr, n-1)
				if n%2 == 1 {
					tmp := arr[i]
					arr[i] = arr[n-1]
					arr[n-1] = tmp
				} else {
					tmp := arr[0]
					arr[0] = arr[n-1]
					arr[n-1] = tmp
				}
			}
		}
	}
	helper(arr, len(arr))
	return res
}

func getpermuted(bytes []int) uint {
	permuted := uint(0)
	shift := uint(8)
	for p, k := range bytes {
		if p == 0 {
			permuted |= uint(k)
		} else {

			permuted |= uint(k << shift)
			shift += 8
		}
	}
	return permuted
}

func permutebytes(data string, size int, op_size int) string {
	if op_size < size {
		op_size = size
	}
	if len(data) > op_size/8 {
		pos := randomint(0, len(data)-(op_size/8))
		var mode string
		switch op_size {
		case 8:
			mode = "B"
			break
		case 16:
			mode = "H"
			break
		case 32:
			mode = "I"
			break
		default:
			mode = "I"
			break
		}
		fuzz := uint(unpack(mode, data[pos:pos+(op_size/8)]))
		if size == 8 {
			lsb := fuzz & 0xff
			lsb_left := lsb & 0xf0
			lsb_right := lsb & 0xf
			fuzz = (fuzz >> uint(size)) << uint(size)
			fuzz |= lsb_right<<4 | lsb_left>>4
		} else if size == 16 {
			lsb := fuzz & 0xff
			next_byte := (fuzz >> 8) & 0xff
			res := permutations([]int{int(lsb), int(next_byte)})
			out := res[randomint(0, len(res))]
			fuzz = (fuzz >> uint(size)) << uint(size)
			fuzz |= getpermuted(out)
		} else if size == 32 {
			lsb := fuzz & 0xff
			next_byte_a := (fuzz >> 8) & 0xff
			next_byte_b := (fuzz >> 16) & 0xff
			next_byte_c := (fuzz >> 24) & 0xff
			res := permutations([]int{int(lsb), int(next_byte_a), int(next_byte_b), int(next_byte_c)})
			out := res[randomint(0, len(res))]
			fuzz = getpermuted(out)
		}
		tmp := data[:pos]
		tmp += pack(mode, int(fuzz))
		tmp += data[pos+(op_size/8):]
		return tmp
	} else {
		return ""
	}
}

func countfuzzfactor(data string, factor float64) int {
	f := int(math.Ceil(float64(len(data)) * factor))
	return randomint(f/2, f) + 1
}

func getmethodname(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func fuzz(data string) string {
	var random_size = true
	var strategies = []interface{}{bitflip, byteflip, arithmetics, randombytes, knownintegers, duplicatetoken, removetoken, insertrandombytes, permutebytes}
	var fuzzing_flip = 1
	var fuzzing_size = 8
	var data_size = 8
	var factor = 0.02
	x := 0
	tmp := ""
	valid_sizes := []int{8, 16, 32}
	valid_flips := []int{1, 2, 3, 4}
	max_iterate := countfuzzfactor(data, factor)
	for x < max_iterate {
		method := strategies[randomint(0, len(strategies))]
		if random_size {
			fuzzing_size = valid_sizes[randomint(0, len(valid_sizes)-1)]
			data_size = valid_sizes[randomint(0, len(valid_sizes)-1)]
			fuzzing_flip = valid_flips[randomint(0, len(valid_flips)-1)]
		}
		if getmethodname(method) == "main.bitflip" {
			tmp = method.(func(string, int, int) string)(data, fuzzing_flip, data_size)
		} else {
			tmp = method.(func(string, int, int) string)(data, fuzzing_size, data_size)
		}
		x++
		if tmp != "" {
			data = tmp
		}
	}
	return url.QueryEscape(data)
}
