package utils

import (
	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"
	"unicode"
)

func RandSample(list []interface{}, n int, canRepeat bool) []interface{} {
	if n <= 0 {
		return []interface{}{}
	}
	if !canRepeat && len(list) <= n {
		return list
	}

	var ret []interface{}
	var set map[int]struct{}
	if !canRepeat {
		set = make(map[int]struct{})
	}
	for len(ret) < n {
		i := rand.Intn(len(list))
		if !canRepeat {
			if _, ok := set[i]; ok {
				continue
			}
			set[i] = struct{}{}
		}
		ret = append(ret, list[i])
	}
	return ret
}

func RandInt32Sample(list []int32, n int, canRepeat bool) []int32 {
	var args []interface{}
	for _, e := range list {
		args = append(args, e)
	}

	sampleList := RandSample(args, n, canRepeat)
	var ret []int32
	for _, e := range sampleList {
		ret = append(ret, e.(int32))
	}
	return ret
}

func RandIntSample(list []int, n int, canRepeat bool) []int {
	var args []interface{}
	for _, e := range list {
		args = append(args, e)
	}

	sampleList := RandSample(args, n, canRepeat)
	var ret []int
	for _, e := range sampleList {
		ret = append(ret, e.(int))
	}
	return ret
}

func RandUInt32Sample(list []uint32, n int, canRepeat bool) []uint32 {
	var args []interface{}
	for _, e := range list {
		args = append(args, e)
	}

	sampleList := RandSample(args, n, canRepeat)
	var ret []uint32
	for _, e := range sampleList {
		ret = append(ret, e.(uint32))
	}
	return ret
}

func IntAbs(n int) int {
	if n < 0 {
		return -n
	} else {
		return n
	}
}

type IShuffleAble interface {
	ShuffleTag()
}

func Shuffle(a []interface{}) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
		//a[i], a[j] = a[j], a[i]
	}
}

func ShuffleInt(a []int) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
		//a[i], a[j] = a[j], a[i]
	}
}

func ShuffleString(a []string) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

func ShuffleInt32(a []int32) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
		//a[i], a[j] = a[j], a[i]
	}
}

func ShuffleUInt32(a []uint32) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
		//a[i], a[j] = a[j], a[i]
	}
}

func RandIndexOfWeights(ws []int) int {
	tw := 0

	for _, w := range ws {
		tw += w
	}

	rw := rand.Intn(tw + 1)
	tw = 0

	for i, w := range ws {
		tw += w

		if rw < tw {
			return i
		}
	}

	return -1
}

func RandFewNumberWithSum(sum, n int) []int {
	var numList []int
	if sum < n {
		numList = append(numList, n)
		return numList
	}
	var l []int
	for i := 1; i<sum; i++ {
		l = append(l, i)
	}
	ls := RandIntSample(l, n-1, false)
	ls = append(ls, 0, sum)
	sort.Ints(ls)
	for i := 1; i < len(ls); i++ {
		a := ls[i] - ls[i-1]
		numList = append(numList, a)
	}
	return numList
}

// 生成n个和为sum且小于max的随机数
func RandFewIntWithLimit(sum, n, max int) []int {
	var vals, numList []int
	valSet := map[int]struct{}{}
	// 可取浮点型的3倍平均数, 也可直接除取整数部分
	//max := sum / n * 3   //int(math.Floor(float64(sum) / float64(n) * 3))  可取浮点型的3倍平均数

	// 最大值不能小于等于平均数
	min := float64(sum)/float64(n)
	if float64(max) <= min {
		max = int(math.Floor(min))+1
	}

	if max < sum {
		max2 := max
		// 将0-sum分割成长度小于max的x段
		for i := 0; i < n-1; i++ {
			val := max2 - 1
			valSet[val] = struct{}{}
			max2 = val + max
			if max2 > sum {
				break
			}
		}
	}

	// 把筛选出还没被选中的数
	var ls []int
	for i := 1; i < sum; i++ {
		if _, ok := valSet[i]; !ok {
			ls = append(ls, i)
		} else {
			vals = append(vals, i)
		}
	}

	// 分割成n段需要n-1个随机数，从筛选出的数中随机出剩余个数的数
	//log.Printf(" ====== ls=%v  vals=%v", ls, vals)
	remainCnt := n - 1 - len(vals)
	vals = append(vals, RandIntSample(ls, remainCnt, false)...)

	// 在n-1个数中加上首端和末端， 即0和sum
	vals = append(vals, 0, sum)

	// 对n+1个数进行排序
	sort.Ints(vals)
	//log.Printf(" ====== raminCnt=%d   vals=%v", remainCnt, vals)

	// 求出每一段的长度， 即n个随机数
	for i := 1; i < len(vals); i++ {
		a := vals[i] - vals[i-1]
		numList = append(numList, a)
	}
	return numList
}

func randIntsDelSum(sum, min int, nums []int) (sum2 int, nums2 []int) {
	if sum <= 0 {
		return sum, nums
	}
	flag := 0
	for j, num := range nums {
		if sum <= 0 {
			break
		}
		if num > min {
			nums[j] = num - 1
			sum--
			flag = 1
		}
	}
	if sum > 0 && flag == 1{
		return randIntsDelSum(sum, min, nums)
	}
	return sum, nums
}

func RandFewIntWithMaxMinLimit(sum, n, max, min int) []int {
	nums := RandFewIntWithLimit(sum, n, max)
	delSum := 0
	for i, num := range nums {
		if num < min {
			nums[i] = min
			delSum += min - num
		}
	}
	_, nums = randIntsDelSum(delSum, min, nums)
	return nums
}

const (
	TimeFormat1 = "060102"
	TimeFormat2 = "2006-01-02 15:04:05"
)

func TimeToString(timestamp int64, format string) string {
	tm := time.Unix(timestamp, 0)
	return tm.Format(format)
}

func StringToTime(value string, format string) (time.Time, error) {
	return time.ParseInLocation(format, value, time.Local)
}

func UnixToTime(t int64) (time.Time, error) {
	timeStr := time.Unix(t, 0).Format(TimeFormat2)
	return StringToTime(timeStr, TimeFormat2)
}

func HourToTodayTime(hour int) (time.Time, error) {
	h := hour
	if h < 0 || h >= 24{
		h = 0
	}
	y, m, d := time.Now().Date()
	var month, day string
	if m < 10 {
		month = "0" + strconv.Itoa(int(m))
	} else {
		month = strconv.Itoa(int(m))
	}

	if d < 10 {
		day = "0" + strconv.Itoa(int(d))
	} else {
		day = strconv.Itoa(int(d))
	}
	timeStr := strconv.Itoa(y) + "-" + month + "-" + day + " " + strconv.Itoa(h) + ":00:00"
	times, err := StringToTime(timeStr, TimeFormat2)
	if hour == 24 {
		times = times.AddDate(0,0,1)
	}
	return times, err
}

func IsSameDay(timestamp1, timestamp2 int64) bool {
	tm1 := time.Unix(timestamp1, 0)
	tm2 := time.Unix(timestamp2, 0)
	return tm1.Year() == tm2.Year() && tm1.Month() == tm2.Month() && tm1.Day() == tm2.Day()
}

func StrIsNumber(str string) bool {
	for _, v := range str {
		if !unicode.IsNumber(v) {
			return false
		}
	}
	return true
}