package helper

import (
	"encoding/json"
	"runtime/debug"
	"strconv"
	"tianzong/tzlog"
	"time"
)

// GetString convert interface to string.
func GetString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch v.(type) { //多选语句switch
	case int32:
		return strconv.Itoa(int(v.(int32)))
	case int:
		return strconv.Itoa(int(v.(int)))
	case uint16:
		return strconv.Itoa(int(v.(uint16)))
	case float64:
		return strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case []uint8:
		return string(v.([]uint8))
	case int64:
		return strconv.FormatInt(v.(int64), 10)
	case uint64:
		return strconv.FormatUint(v.(uint64), 10)
	case time.Time:
		return v.(time.Time).Format("2006-01-02 15:04:05")
	case string:
		return v.(string)

	default:
		debug.PrintStack()
		tzlog.W("ToString 类型没有定义:%v>>%T", v, v)
		ret, err := json.Marshal(v)
		if err != nil {
			tzlog.W("ToString Marshal:%v>>%v", err, v)
		}
		return string(ret)
	}
}

// GetInt convert interface to int.
func GetInt(v interface{}) int {
	return int(GetInt32(v))
}

func GetInt32(v interface{}) int32 {
	if v == nil {
		return 0
	}
	switch v.(type) { //多选语句switch
	case int32:
		return v.(int32)
	case int64:
		return int32(v.(int64))
	case int:
		return int32(v.(int))
	case float64:
		return int32(v.(float64))
	case []uint8:
		v1, err := strconv.ParseInt(string(v.([]uint8)), 10, 64)
		if err != nil {
			return 0
		}
		return int32(v1)
	case string:
		if v.(string) == "" {
			return 0
		}
		ret, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			tzlog.W("value %v toint32: %v", v, err)
			debug.PrintStack()
		}
		return int32(ret)
	}
	tzlog.W("ToInt32 类型没有定义:%v  %T", v, v)
	return 0
}

// GetInt64 convert interface to int64.
func GetInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch v.(type) { //多选语句switch
	case int32:
		return int64(v.(int32))
	case int64:
		return v.(int64)
	case int:
		return int64(v.(int))
	case float64:
		return int64(v.(float64))
	case []uint8:
		v1, err := strconv.ParseInt(string(v.([]uint8)), 10, 64)
		if err != nil {
			return 0
		}
		return v1
	case string:
		if v.(string) == "" {
			return 0
		}
		ret, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			tzlog.W("value %v toint32: %v", v, err)
			debug.PrintStack()
		}
		return ret
	}
	debug.PrintStack()
	tzlog.W("ToInt64 类型没有定义:%v :%T", v, v)
	return 0
}

// GetFloat64 convert interface to float64.
func GetFloat64(v interface{}) float64 {
	switch v.(type) { //多选语句switch
	case int32:
		return float64(v.(int32))
	case int:
		return float64(v.(int))
	case float64:
		return v.(float64)
	case float32:
		return float64(v.(float32))
	case string:
		ret, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			tzlog.W("value %v ToFloat64: %v", v, err)
			debug.PrintStack()
		}
		return ret
	}
	debug.PrintStack()
	tzlog.W("ToFloat64 类型没有定义:%v>>%T", v, v)
	return 0
}

// GetBool convert interface to bool.
func GetBool(v interface{}) bool {
	switch result := v.(type) {
	case bool:
		return result
	default:
		if d := GetString(v); d != "" {
			value, _ := strconv.ParseBool(d)
			return value
		}
	}
	return false
}

var defaultTime = time.Date(2006, 1, 2, 15, 4, 5, 0, time.Local)

func GetTime(value interface{}) time.Time {
	if value == nil {
		return defaultTime
	}
	switch value.(type) { //多选语句switch
	case time.Time:
		value := value.(time.Time)
		return value
	case string:
		str := value.(string)
		ret, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
		if err != nil {
			tzlog.W("toTime :%v>>%v", value, err)
			return defaultTime
		}
		return ret
	}
	tzlog.W("ToTime 类型没有定义:%v>>%T", value, value)
	return defaultTime
}
func GetBytes(value interface{}) []byte {
	if value == nil {
		return []byte{}
	}
	switch value.(type) { //多选语句switch
	case []uint8:
		value := []byte(value.([]uint8))
		return value
	}
	tzlog.W("ToBytes 类型没有定义:%v>>%T", value, value)
	return []byte{}
}
