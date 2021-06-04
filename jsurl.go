package jsurl

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var reserved map[string]interface{}
var reDefault *regexp.Regexp
var reDefault2 *regexp.Regexp

func init() {
	reserved = map[string]interface{}{
		"true":  true,
		"false": false,
		"null":  nil,
	}
	reDefault = regexp.MustCompile(`[^)~]`)
	reDefault2 = regexp.MustCompile(`[\d\-]`)
}

func Stringify(itf interface{}) string {
	tmp := []string{}

	switch v := reflect.ValueOf(itf); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "~" + fmt.Sprint(v.Int())
	case reflect.Float32, reflect.Float64:
		return "~" + fmt.Sprint(v.Float())
	case reflect.Bool:
		return "~" + fmt.Sprint(v.Bool())
	case reflect.String:
		return "~'" + encode(v.String())
	case reflect.Func:
		return ""
	case reflect.Array, reflect.Slice:
		if v.Len() < 1 {
			return "~(~)"
		}

		for i := 0; i < v.Len(); i++ {
			strElem := Stringify(v.Index(i).Interface())
			if strElem == "" {
				strElem = "~null"
			}
			tmp = append(tmp, strElem)
		}

		return fmt.Sprintf("~(%s)", strings.Join(tmp, ""))
	case reflect.Struct:
		if v.IsZero() {
			return "~()"
		}
		tmp := []string{}

		fields := reflect.TypeOf(itf)
		values := reflect.ValueOf(itf)

		num := v.NumField()

		for i := 0; i < num; i++ {
			field := fields.Field(i)
			value := values.Field(i)

			if strValue := Stringify(value.Interface()); strValue != "" {
				tmp = append(tmp, encode(field.Name)+strValue)
			}

		}

		return "~(" + strings.Join(tmp, "~") + ")"
	default:
		return "~null"
	}

}

func encode(s string) string {
	if ok, _ := regexp.MatchString(`[^\w-.]`, s); !ok {
		return s
	}

	re := regexp.MustCompile(`[\w-.]`)
	out := ""
	for _, v := range s {
		if re.MatchString(string(v)) {
			out += string(v)
			continue
		}

		//replace $ w !
		if v == 0x24 {
			out += "!"
			continue
		}

		if v <= 0xff {
			out += fmt.Sprintf("*%x", v)
			continue
		}
		out += fmt.Sprintf("**%x", v)
	}

	return out
}

func Parse(s string, itf interface{}) error {
	if s == "" {
		return nil
	}
	re := regexp.MustCompile(`%(25)*27`)
	s = re.ReplaceAllString(s, `'`)
	i := 0
	len := len(s)

	runeSliceOfS := []rune(s)

	rs, _, err := parseOne(runeSliceOfS, i, len)
	if err != nil {
		return err
	}

	buffer, err := json.Marshal(rs)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buffer, itf)
	if err != nil {
		return err
	}

	return nil
}

func eat(s []rune, i int, expected rune) (int, error) {
	if s[i] != expected {
		return i, fmt.Errorf("bad JSURL syntax: expected %s, got %s", string(s[i]), string(expected))
	}
	i += 1
	return i, nil
}

func decode(s []rune, i int, len int) (int, string) {
	beg := i
	var ch rune
	r := ""

	for i < len {
		ch = s[i]
		//~ = 0x7E || ) 0x29
		if ch == 0x7E || ch == 0x29 {
			break
		}
		switch ch {
		case '*':
			if beg < i {
				r += string(s[beg:i])
			}
			if s[i+1] == '*' {
				cCode, _ := strconv.ParseInt(string(s[i+2:i+6]), 16, 32)
				r += string(rune(cCode))
				i += 6
				beg = i
				break
			}
			cCode, _ := strconv.ParseInt(string(s[i+1:i+3]), 16, 32)
			r += string(rune(cCode))
			i += 3
			beg = i
		case '!':
			if beg < i {
				r += string(s[beg:i])
			}
			r += "$"
			i += 1
			beg = i
		default:
			i += 1
		}
	}
	return i, r + string(s[beg:i])
}

func parseOne(s []rune, i int, len int) (interface{}, int, error) {
	var result interface{}
	var beg int

	i, err := eat(s, i, '~')
	if err != nil {
		return result, i, err
	}

	ch := s[i]
	switch ch {
	case '(':
		i += 1
		if s[i] == '~' {
			out := []interface{}{}
			if s[i+1] == ')' {
				i += 1
			} else {
				for {
					var rs interface{}
					rs, i, err = parseOne(s, i, len)
					if err != nil {
						return result, i, err
					}
					out = append(out, rs)
					if s[i] != '~' {
						break
					}
				}
			}
			result = out
		} else {
			out := map[string]interface{}{}
			if s[i] != ')' {
				for {
					key := ""
					i, key = decode(s, i, len)
					out[key], i, err = parseOne(s, i, len)
					if err != nil {
						return result, i, err
					}

					if s[i] != '~' {
						break
					}
					i += 1
				}
			}
			result = out
		}
		i, err = eat(s, i, ')')
		if err != nil {
			return result, i, err
		}
	case 0x27: //thats '
		i += 1
		i, result = decode(s, i, len)
	default:
		beg = i
		i += 1
		for i < len {
			if !reDefault.MatchString(string(s[i])) {
				break
			}
			i += 1
		}
		sub := string(s[beg:i])
		if reDefault2.MatchString(string(ch)) {
			rs, err := strconv.ParseFloat(sub, 64)
			if err != nil {
				return result, i, err
			}
			result = rs
		} else {
			rs, ok := reserved[sub]
			if !ok {
				return result, i, fmt.Errorf("bad value keyword: %s", sub)
			}
			result = rs
		}
	}
	return result, i, nil
}
