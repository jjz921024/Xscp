package utils

import "strconv"

func ConvertPerm(perm string) string {
	if len(perm) != 10 {
		return "0400"
	}

	result := "0"
	for i := 0; i < 3; i++ {
		s := perm[i*3+1 : i*3+1+3]
		var r int
		if s[0:1] == "r" {
			r += 4
		}
		if s[1:2] == "w" {
			r += 2
		}
		if s[2:3] == "x" {
			r += 1
		}
		result += strconv.Itoa(r)
	}

	return result
}
