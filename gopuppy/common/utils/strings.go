package utils

func EditDistance(s1, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)
	len1 := len(r1)
	len2 := len(r2)
	cache := make([]int, len2 + 1)

	for i := 0; i <= len2; i++ {
		cache[i] = i
	}

	var old int
	for i := 1; i <= len1; i++ {
		old = i - 1
		cache[0] = i

		for j := 1; j <= len2; j++ {
			temp := cache[j]

			if r1[i-1] == r2[j-1] {
				cache[j] = old
			} else {
				d1 := cache[j] + 1
				d2 := cache[j-1] + 1
				d3 := old + 1
				if d2 < d1 {
					d1 = d2
				}
				if d3 < d1 {
					d1 = d3
				}
				cache[j] = d1
			}

			old = temp
		}
	}

	return cache[len2]
}
