package log

import (
	"time"
	_ "unsafe"
)

const minSec = 60
const hourSec = 3600
const daySec = 3600 * 24 //每天的秒数

const firstYears = 365
const secondYears = 365 + 365
const thirdYears = 365 + 365 + 366
const fourYears = 365 + 365 + 366 + 365 //每个四年的总天数

var norMonth = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}  //平年
var leapMonth = [12]int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31} //闰年
var offset int64

//go:linkname now time.now
func now() (sec int64, nsec int32)

func unix() int64 {
	sec, _ := now()
	return sec
}

func dateClock(unixSecLocal int64) (year, month, day, hour, min, sec, yDay, daySecond int) {
	var nRemain int
	if unixSecLocal < 0 {
		nUnixSec := -unixSecLocal
		nDays := int(nUnixSec / daySec)
		daySecond = (daySec - int(nUnixSec-int64(nDays*daySec))) % daySec
		nYear4 := nDays/fourYears + 1
		nRemain = nYear4*fourYears - nDays
		if daySecond == 0 {
			nRemain += 1
		}
		year = 1970 - nYear4<<2
	} else {
		nDays := int(unixSecLocal / daySec)
		daySecond = int(unixSecLocal - int64(nDays*daySec))
		nYear4 := nDays / fourYears
		nRemain = nDays - nYear4*fourYears + 1
		year = 1970 + nYear4<<2
	}
	pMonth := &norMonth
	if nRemain <= firstYears {

	} else if nRemain <= secondYears {
		year += 1
		nRemain -= firstYears
	} else if nRemain <= thirdYears {
		year += 2
		nRemain -= secondYears
		pMonth = &leapMonth
	} else if nRemain <= fourYears {
		year += 3
		nRemain -= thirdYears
	} else {
		year += 4
		nRemain -= fourYears
	}
	yDay = nRemain
	var nTemp int
	for i := 0; i < 12; i++ {
		nTemp = nRemain - pMonth[i]
		if nTemp < 1 {
			month = i + 1
			if nTemp == 0 {
				day = pMonth[i]
			} else {
				day = nRemain
			}
			break
		}
		nRemain = nTemp
	}
	hour = daySecond / hourSec
	inHourSec := daySecond - hour*hourSec
	min = inHourSec / minSec
	sec = inHourSec - min*minSec
	return
}


type DateTime interface {
	Year()int
	Month()int
	Day()int
	Hour()int
	Min()int
	Sec()int
	YmdHMS() string
}

type datetime struct {
	unix  int64
	year  int
	month int
	day   int
	hour  int
	min   int
	sec   int
}

func (my *datetime) Year()int{
	return my.year
}

func (my *datetime) Month()int{
	return my.month
}

func (my *datetime) Day()int{
	return my.day
}

func (my *datetime) Hour()int{
	return my.hour
}

func (my *datetime) Min()int{
	return my.min
}

func (my *datetime) Sec()int{
	return my.sec
}

func (my *datetime) YmdHMS() string {
	return my.format("%Y/%m/%d %H:%M:%S")
}

func (my *datetime) flushTo(unix int64) {
	if unix == my.unix {
		return
	}
	my.year, my.month, my.day, my.hour, my.min, my.sec, _, _ = dateClock(unix + offset)
}

func (my *datetime) format(formatter string) string {
	//var theTime []byte
	var theTime = getBuffer(len(formatter) * 2)
	defer theTime.free()
	length := len(formatter)
	for i := 0; i < length; {
		c := formatter[i]
		if c == '%' {
			if i+1 == length {
				break
			}
			c2 := formatter[i+1]
			switch c2 {
			case 'Y': //四位数的年份表示（0000-9999）
				theTime.AppendInt(my.year, 4)
			case 'm': //月份（01-12）
				theTime.AppendInt(my.month, 2)
			case 'd': //月内中的一天（0-31）
				theTime.AppendInt(my.day, 2)
			case 'H': //24小时制小时数（0-23）
				theTime.AppendInt(my.hour, 2)
			case 'M': //分钟数（00=59）
				theTime.AppendInt(my.min, 2)
			case 'S': //秒（00-59）
				theTime.AppendInt(my.sec, 2)
			default:
				theTime.AppendBytes(c2)
			}
			i += 2
		} else {
			theTime.AppendBytes(c)
			i += 1
		}
	}
	return string(*theTime)
}

func init()  {
	_, local := time.Now().Zone()
	offset = int64(local)
}